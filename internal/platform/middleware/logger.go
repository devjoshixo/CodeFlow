package middleware

import (
	"codeflow/internal/platform/logger"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func RequestLogger(next http.Handler) http.Handler {
	l := logger.Get()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)
		traceID := r.Context().Value(TraceIDKey).(string)

		l.Info("request", "trace_id", traceID, "method", r.Method, "path", r.URL.Path, "status", wrapped.status, "duration", time.Since(start).String())
	})
}
