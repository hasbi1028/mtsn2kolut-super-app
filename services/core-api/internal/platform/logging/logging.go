package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

const RequestIDHeader = "X-Request-ID"

var sensitiveKeyPattern = regexp.MustCompile(`(?i)(authorization|cookie|password|passwd|token|secret|api[_-]?key|database[_-]?url|connection[_-]?string|jwt)`)
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,96}$`)

func NewRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req_unknown"
	}
	return "req_" + hex.EncodeToString(b[:])
}

func SanitizeRequestID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestIDPattern.MatchString(requestID) {
		return requestID
	}
	return ""
}

func RequestIDFromHeader(r *http.Request) string {
	if r == nil {
		return NewRequestID()
	}
	for _, key := range []string{RequestIDHeader, "X-Request-Id", "X-Correlation-ID"} {
		if value := SanitizeRequestID(r.Header.Get(key)); value != "" {
			return value
		}
	}
	return NewRequestID()
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	requestID = SanitizeRequestID(requestID)
	if requestID == "" {
		requestID = NewRequestID()
	}
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return SanitizeRequestID(requestID)
	}
	return ""
}

func Attr(key string, value any) slog.Attr {
	if sensitiveKeyPattern.MatchString(key) {
		return slog.String(key, "[REDACTED]")
	}
	return slog.Any(key, value)
}

func RedactAttrs(attrs ...slog.Attr) []slog.Attr {
	redacted := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		if sensitiveKeyPattern.MatchString(attr.Key) {
			redacted = append(redacted, slog.String(attr.Key, "[REDACTED]"))
			continue
		}
		redacted = append(redacted, attr)
	}
	return redacted
}

func ErrorAttrs(err error) []slog.Attr {
	if err == nil {
		return nil
	}
	return append([]slog.Attr{slog.String("error", err.Error())}, PgErrorAttrs(err)...)
}

func Info(ctx context.Context, event string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelInfo, event, withRequestID(ctx, attrs...)...)
}

func Warn(ctx context.Context, event string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelWarn, event, withRequestID(ctx, attrs...)...)
}

func Error(ctx context.Context, event string, err error, attrs ...slog.Attr) {
	all := append(attrs, ErrorAttrs(err)...)
	slog.LogAttrs(ctx, slog.LevelError, event, withRequestID(ctx, RedactAttrs(all...)...)...)
}

func withRequestID(ctx context.Context, attrs ...slog.Attr) []slog.Attr {
	requestID := RequestID(ctx)
	if requestID == "" {
		return RedactAttrs(attrs...)
	}
	out := make([]slog.Attr, 0, len(attrs)+1)
	out = append(out, slog.String("request_id", requestID))
	out = append(out, RedactAttrs(attrs...)...)
	return out
}
