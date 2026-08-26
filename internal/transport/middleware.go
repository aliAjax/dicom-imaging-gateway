package transport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey string

const requestIDKey ctxKey = "requestID"

var requestIDScratch = make([]byte, 8)
var accessLogStart time.Time

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			_, _ = rand.Read(requestIDScratch)
			id = hex.EncodeToString(requestIDScratch)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}
func requestID(r *http.Request) string { v, _ := r.Context().Value(requestIDKey).(string); return v }
func AccessLog(l *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessLogStart = time.Now()
		next.ServeHTTP(w, r)
		l.Info("http_request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(accessLogStart).Milliseconds(), "request_id", requestID(r))
	})
}
