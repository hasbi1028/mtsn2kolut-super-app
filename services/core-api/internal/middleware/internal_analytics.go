package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type internalAnalyticsRecorder interface {
	CreateEvent(ctx context.Context, in service.CreateInternalAnalyticsEventInput) (service.InternalAnalyticsEventReceipt, error)
}

func InternalAnalytics(recorder internalAnalyticsRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if recorder == nil || shouldSkipInternalAnalyticsMiddleware(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			eventName, eventGroup, module := internalAnalyticsEventForRequest(r, ww.Status())
			if eventName == "" {
				return
			}
			claims, _ := api.ClaimsFromContext(r.Context())
			_, _ = recorder.CreateEvent(r.Context(), service.CreateInternalAnalyticsEventInput{
				EventName:       eventName,
				EventGroup:      eventGroup,
				SourceSurface:   "core_api",
				ActorUserID:     analyticsActorUserID(claims),
				ActorRole:       analyticsActorRole(claims),
				RouteGroup:      module,
				Module:          module,
				Result:          analyticsResultForStatus(ww.Status()),
				StatusCodeClass: analyticsStatusClass(ww.Status()),
				DurationBucket:  analyticsDurationBucket(time.Since(start)),
				Metadata: map[string]any{
					"http_method": r.Method,
					"operation":   analyticsOperationForMethod(r.Method),
					"resource":    module,
				},
			})
		})
	}
}

func shouldSkipInternalAnalyticsMiddleware(path string) bool {
	return !strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/api/internal-analytics") ||
		strings.HasPrefix(path, "/api/auth/refresh") ||
		strings.HasPrefix(path, "/api/auth/login")
}

func internalAnalyticsEventForRequest(r *http.Request, status int) (eventName, eventGroup, module string) {
	path := r.URL.Path
	if status == 403 {
		if strings.HasPrefix(path, "/api/rbac") {
			return "rbac.permission_denied", "rbac", "rbac"
		}
		return "security.forbidden", "security", routeModule(path)
	}

	module = routeModule(path)
	switch {
	case strings.HasPrefix(path, "/api/bank-soal") || strings.HasPrefix(path, "/api/cbt/questions") || strings.HasPrefix(path, "/api/cbt/assets"):
		return bankSoalAnalyticsEvent(path, r.Method)
	case strings.HasPrefix(path, "/api/pusaka"):
		return pusakaAnalyticsEvent(path, r.Method)
	case strings.HasPrefix(path, "/api/users"):
		return usersAnalyticsEvent(path, r.Method)
	case strings.HasPrefix(path, "/api/rbac"):
		return rbacAnalyticsEvent(r.Method)
	case strings.HasPrefix(path, "/api/school-profile"):
		if r.Method == http.MethodGet {
			return "security.settings_view", "security", "settings"
		}
		return "security.settings_update", "security", "settings"
	default:
		return "", "", ""
	}
}

func bankSoalAnalyticsEvent(path, method string) (string, string, string) {
	if strings.Contains(path, "/export") || strings.Contains(path, "/template") {
		return "bank_soal.export", "bank_soal", "bank_soal"
	}
	if strings.Contains(path, "/import-legacy") {
		return "bank_soal.import_complete", "bank_soal", "bank_soal"
	}
	if strings.Contains(path, "/workflow") || strings.Contains(path, "/bulk-workflow") {
		return "bank_soal.review_decision", "bank_soal", "bank_soal"
	}
	if strings.Contains(path, "/assets") && method == http.MethodPost {
		return "bank_soal.asset_upload", "bank_soal", "bank_soal"
	}
	switch method {
	case http.MethodGet:
		if strings.Contains(path, "/summary") {
			return "bank_soal.readiness_check", "bank_soal", "bank_soal"
		}
		return "bank_soal.list_view", "bank_soal", "bank_soal"
	case http.MethodPost:
		return "bank_soal.question_create", "bank_soal", "bank_soal"
	case http.MethodPut, http.MethodPatch:
		return "bank_soal.question_update", "bank_soal", "bank_soal"
	default:
		return "", "", ""
	}
}

func pusakaAnalyticsEvent(path, method string) (string, string, string) {
	switch {
	case strings.Contains(path, "/scheduler/tick"):
		return "pusaka.scheduler_tick", "pusaka", "pusaka"
	case strings.Contains(path, "/run-all") || strings.Contains(path, "/run-now") || strings.Contains(path, "/sync-attendance"):
		return "pusaka.manual_run", "pusaka", "pusaka"
	case strings.Contains(path, "/cancel"):
		return "pusaka.queue_cancel", "pusaka", "pusaka"
	case strings.Contains(path, "/settings") && method != http.MethodGet:
		return "pusaka.settings_update", "pusaka", "pusaka"
	case strings.Contains(path, "/employees") && method != http.MethodGet:
		return "pusaka.employee_scope_update", "pusaka", "pusaka"
	case method == http.MethodGet:
		return "pusaka.dashboard_view", "pusaka", "pusaka"
	default:
		return "", "", ""
	}
}

func usersAnalyticsEvent(path, method string) (string, string, string) {
	if strings.Contains(path, "/reset-password") || strings.Contains(path, "/force-password-change") {
		return "users.reset_password", "users", "users"
	}
	if strings.Contains(path, "/status") {
		return "users.suspend_change", "users", "users"
	}
	switch method {
	case http.MethodGet:
		return "users.list_view", "users", "users"
	case http.MethodPost:
		return "users.create", "users", "users"
	case http.MethodPut, http.MethodPatch:
		return "users.update", "users", "users"
	default:
		return "", "", ""
	}
}

func rbacAnalyticsEvent(method string) (string, string, string) {
	if method == http.MethodGet {
		return "rbac.roles_view", "rbac", "rbac"
	}
	return "rbac.permission_update", "rbac", "rbac"
}

func routeModule(path string) string {
	trimmed := strings.TrimPrefix(path, "/api/")
	first := strings.Split(trimmed, "/")[0]
	first = strings.ReplaceAll(first, "-", "_")
	if first == "" {
		return "api"
	}
	return first
}

func analyticsOperationForMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead:
		return "read"
	case http.MethodPost:
		return "create_or_action"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "other"
	}
}

func analyticsResultForStatus(status int) string {
	if status == 0 {
		status = http.StatusOK
	}
	switch {
	case status == http.StatusForbidden:
		return "blocked"
	case status >= 200 && status < 400:
		return "success"
	default:
		return "failed"
	}
}

func analyticsStatusClass(status int) string {
	if status == 0 {
		status = http.StatusOK
	}
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	default:
		return "unknown"
	}
}

func analyticsDurationBucket(duration time.Duration) string {
	switch {
	case duration < 100*time.Millisecond:
		return "lt_100ms"
	case duration < 500*time.Millisecond:
		return "100_500ms"
	case duration < time.Second:
		return "500_1000ms"
	case duration < 3*time.Second:
		return "1_3s"
	default:
		return "gt_3s"
	}
}

func analyticsActorUserID(claims jwt.MapClaims) pgtype.UUID {
	for _, key := range []string{"sub", "uid"} {
		if raw, ok := claims[key].(string); ok {
			var id pgtype.UUID
			if err := id.Scan(strings.TrimSpace(raw)); err == nil && id.Valid {
				return id
			}
		}
	}
	return pgtype.UUID{}
}

func analyticsActorRole(claims jwt.MapClaims) string {
	if rawRoles, ok := claims["roles"].([]any); ok {
		for _, rawRole := range rawRoles {
			if role, ok := rawRole.(string); ok && strings.TrimSpace(role) != "" {
				return strings.TrimSpace(role)
			}
		}
	}
	if rawRoles, ok := claims["roles"].([]string); ok {
		for _, role := range rawRoles {
			if strings.TrimSpace(role) != "" {
				return strings.TrimSpace(role)
			}
		}
	}
	if role, ok := claims["role"].(string); ok {
		return strings.TrimSpace(role)
	}
	return ""
}
