package middleware

import (
	"codeflow/internal/platform/logger"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	l := logger.Get()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				l.Error("panic recovered", "error", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal Server Error"}`))
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
