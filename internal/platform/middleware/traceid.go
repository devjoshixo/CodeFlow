package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const TraceIDKey contextKey = "trace_id"

func Tracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := uuid.New().String()

		ctx := context.WithValue(r.Context(), TraceIDKey, traceId)

		r = r.WithContext(ctx)

		w.Header().Set("X-Trace-ID", traceId)

		next.ServeHTTP(w, r)
	})
}
