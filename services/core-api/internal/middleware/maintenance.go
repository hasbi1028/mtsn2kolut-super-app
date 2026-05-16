package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type MaintenanceStatusProvider interface {
	Status(ctx context.Context) (service.SystemMaintenanceStatus, error)
}

type MaintenanceGuard struct {
	provider MaintenanceStatusProvider
	ttl      time.Duration
	now      func() time.Time

	mu      sync.Mutex
	cached  service.SystemMaintenanceStatus
	expires time.Time
}

func Maintenance(provider MaintenanceStatusProvider, ttl time.Duration) func(http.Handler) http.Handler {
	guard := &MaintenanceGuard{
		provider: provider,
		ttl:      ttl,
		now:      time.Now,
	}
	return guard.Middleware
}

func (g *MaintenanceGuard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.provider == nil || maintenanceAlwaysAllowedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		status, err := g.status(r.Context())
		if err != nil {
			slog.Error("maintenance guard status lookup failed", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		if !status.Active || status.Window == nil {
			next.ServeHTTP(w, r)
			return
		}
		if maintenanceBypassAllowed(r, status.Window) {
			next.ServeHTTP(w, r)
			return
		}

		switch status.Mode {
		case service.MaintenanceModeGlobal:
			writeMaintenanceBlocked(w, status.Window)
			return
		case service.MaintenanceModeReadOnly:
			if isMaintenanceMutation(r.Method) {
				writeMaintenanceBlocked(w, status.Window)
				return
			}
		case service.MaintenanceModeModule:
			if isMaintenanceMutation(r.Method) && service.MaintenanceModulesMatchAPIPath(status.Window.AffectedModules, r.URL.Path) {
				writeMaintenanceBlocked(w, status.Window)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (g *MaintenanceGuard) status(ctx context.Context) (service.SystemMaintenanceStatus, error) {
	now := g.now()
	if g.ttl <= 0 {
		return g.provider.Status(ctx)
	}

	g.mu.Lock()
	if now.Before(g.expires) {
		cached := g.cached
		g.mu.Unlock()
		return cached, nil
	}
	g.mu.Unlock()

	status, err := g.provider.Status(ctx)
	if err != nil {
		return service.SystemMaintenanceStatus{}, err
	}

	g.mu.Lock()
	g.cached = status
	g.expires = now.Add(g.ttl)
	g.mu.Unlock()
	return status, nil
}

func maintenanceAlwaysAllowedPath(path string) bool {
	if path == "/health" {
		return true
	}
	if path == "/api/system/maintenance/status" || strings.HasPrefix(path, "/api/system/maintenance/") {
		return true
	}
	if path == "/api/auth/login" || path == "/api/auth/refresh" || path == "/api/auth/logout" {
		return true
	}
	return false
}

func isMaintenanceMutation(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func maintenanceBypassAllowed(r *http.Request, window *service.MaintenanceWindowView) bool {
	if window == nil || !window.AllowAdminBypass {
		return false
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return HasAnyRole(claims, window.BypassRoles...)
}

func writeMaintenanceBlocked(w http.ResponseWriter, window *service.MaintenanceWindowView) {
	message := "Sistem sedang dalam mode pemeliharaan."
	if window != nil && strings.TrimSpace(window.Message) != "" {
		message = window.Message
	}
	w.Header().Set("Retry-After", "60")
	api.Err(w, http.StatusServiceUnavailable, message)
}
