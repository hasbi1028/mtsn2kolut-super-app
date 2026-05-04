package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type AuditWriter interface {
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

// Audit logs every mutating (non-GET) request that returns 2xx. The user is
// extracted from JWT claims; internal-key requests pass an empty user_id.
func Audit(q AuditWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := r.Method
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			rec := &recordingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			if rec.status < 200 || rec.status >= 400 {
				return
			}

			var uid pgtype.UUID
			var actor string
			if claims, ok := api.ClaimsFromContext(r.Context()); ok {
				if raw, ok := claims["uid"].(string); ok {
					_ = uid.Scan(raw)
				}
				actor, _ = claims["usr"].(string)
				if actor == "" {
					actor, _ = claims["sub"].(string)
				}
			}

			entityID, routePattern, routeParams := auditRouteMetadata(r)

			meta, _ := json.Marshal(map[string]any{
				"method":       method,
				"path":         r.URL.Path,
				"route":        routePattern,
				"route_params": routeParams,
				"status":       rec.status,
				"actor":        actor,
				"request_id":   r.Header.Get("X-Request-ID"),
			})

			_, _ = q.CreateAuditLog(r.Context(), db.CreateAuditLogParams{
				UserID:     uid,
				Action:     method,
				EntityType: entityFromPath(r.URL.Path),
				EntityID:   entityID,
				Metadata:   meta,
			})
		})
	}
}

func auditRouteMetadata(r *http.Request) (string, string, map[string]string) {
	path := sanitizedPath(r)
	entityID := path
	routeParams := map[string]string{}

	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return entityID, "", routeParams
	}

	for i, key := range rctx.URLParams.Keys {
		if i >= len(rctx.URLParams.Values) {
			continue
		}
		value := rctx.URLParams.Values[i]
		routeParams[key] = value
		if entityID == path && isIDRouteParam(key) && value != "" {
			entityID = value
		}
	}

	return entityID, rctx.RoutePattern(), routeParams
}

func sanitizedPath(r *http.Request) string {
	if r.URL == nil {
		return ""
	}
	return r.URL.Path
}

func isIDRouteParam(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	return key == "id" || strings.HasSuffix(key, "_id") || strings.HasSuffix(key, "id")
}

type recordingResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *recordingResponseWriter) WriteHeader(s int) {
	if w.wroteHeader {
		return
	}
	w.status = s
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(s)
}

func (w *recordingResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(b)
}

// entityFromPath extracts the resource segment from /api/<entity>/...
// e.g. /api/students/{id} -> "students"
func entityFromPath(path string) string {
	// strip leading /api/
	if len(path) > 5 && path[:5] == "/api/" {
		path = path[5:]
	}
	// take first segment
	for i, c := range path {
		if c == '/' {
			return path[:i]
		}
	}
	return path
}
