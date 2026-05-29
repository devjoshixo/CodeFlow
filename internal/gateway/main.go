package gateway

import (
	"codeflow/internal/execution"
	"codeflow/internal/platform/middleware"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	// "github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	// how long we wait for a single write to complete before giving up
	writeWait = 10 * time.Second
	// if we don't hear a pong from the client within this window, the conn is dead
	pongWait = 60 * time.Second
	// send a ping a bit more often than pongWait so the client always has time to reply
	pingPeriod = (pongWait * 9) / 10
)

type Hub struct {
	Connections map[string][]*websocket.Conn
	mu          sync.RWMutex
}

func (h *Hub) AddConnection(executionID string, conn *websocket.Conn) {
	h.mu.Lock()
	h.Connections[executionID] = append(h.Connections[executionID], conn)
	h.mu.Unlock()
}

func (h *Hub) RemoveConnection(executionID string, conn *websocket.Conn) {
	h.mu.Lock()
	conns := h.Connections[executionID]

	for i, c := range conns {
		if c == conn {
			h.Connections[executionID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(executionID string, message []byte) {
	h.mu.RLock()
	conns := h.Connections[executionID]
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			h.RemoveConnection(executionID, conn)
		}
	}
}

func (h *Hub) HandleWebSocket(jwtSecret string, execSvc *execution.ExecutionService, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.handleWebSocket(w, r, jwtSecret, execSvc, redisClient)
	}
}

func (h *Hub) handleWebSocket(w http.ResponseWriter, r *http.Request, jwtSecret string, execSvc *execution.ExecutionService, redisClient *redis.Client) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		http.Error(w, "invalid jwt token", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	executionID := r.PathValue("id")
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	exec, err := execSvc.GetByID(ctx, executionID, userID)
	if err != nil {
		http.Error(w, "execution not found", http.StatusNotFound)
		return
	}
	_ = exec

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	h.AddConnection(executionID, conn)
	defer h.RemoveConnection(executionID, conn)

	// Once Upgrade() hijacks the TCP socket, net/http stops watching it, so
	// r.Context() will never be cancelled when the client disconnects. We make
	// our own cancelable context and let the read pump cancel it on disconnect.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Read pump: gorilla only processes ping/pong/close frames while a read is
	// in flight, and a read is the only way to notice the client went away.
	// Every pong resets the read deadline; if pongs stop, ReadMessage errors
	// out, we cancel the context and the write loop below exits.
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	go func() {
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	pubsub := redisClient.Subscribe(ctx, fmt.Sprintf("execution:%s:output", executionID))
	defer pubsub.Close()
	redisChan := pubsub.Channel()

	// Ping the client periodically so the connection stays alive through idle
	// gaps and so a dead peer is detected (no pong -> read deadline fires).
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case msg, ok := <-redisChan:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				return
			}
		}
	}
}
