package middleware

import (
	"net/http"

	"mtsn2kolut-super-app/backend/internal/platform/logging"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := logging.RequestIDFromHeader(r)
		ctx := logging.WithRequestID(r.Context(), requestID)
		w.Header().Set(logging.RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
