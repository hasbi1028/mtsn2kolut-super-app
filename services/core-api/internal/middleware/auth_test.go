package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"mtsn2kolut-super-app/backend/internal/api"
)

func TestRequireAdminDoesNotBypassWithInternalKeyHeader(t *testing.T) {
	middleware := RequireAdmin()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Internal-Key", "shared-secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAdminAllowsAdminRole(t *testing.T) {
	middleware := RequireAdmin()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role": "admin",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAnyRoleAllowsMatchingRoleFromArray(t *testing.T) {
	middleware := RequireAnyRole("admin", "staf")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/library", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"guru", "staf"},
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAnyRoleRejectsNonMatchingRole(t *testing.T) {
	middleware := RequireAnyRole("admin", "staf")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/library", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role": "guru",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequirePermissionAllowsMatchingPermission(t *testing.T) {
	middleware := RequirePermission("users.read")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"permissions": []any{"users.read", "roles.read"},
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("RequirePermission() status = %d, want 204", rec.Code)
	}
}

func TestRequirePermissionRejectsMissingPermissionAndUnauthenticated(t *testing.T) {
	middleware := RequirePermission("users.read")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"permissions": []string{"bank_soal.read"},
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("RequirePermission(missing) status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/users", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("RequirePermission(unauthenticated) status = %d, want 401", rec.Code)
	}
}

func TestRequireAnyPermissionAllowsAnyOnePermission(t *testing.T) {
	middleware := RequireAnyPermission("users.manage_roles", "roles.manage")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPut, "/rbac", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"permissions": []string{"roles.manage"},
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("RequireAnyPermission() status = %d, want 204", rec.Code)
	}
}

func TestHasAnyPermissionSupportsSingularClaim(t *testing.T) {
	claims := jwt.MapClaims{"permission": "audit.read"}
	if !HasAnyPermission(claims, "users.read", "audit.read") {
		t.Fatal("HasAnyPermission() = false, want true for singular permission claim")
	}
}

func TestRequireAnyPermissionOrRoleAllowsPermissionWithoutLegacyRole(t *testing.T) {
	middleware := RequireAnyPermissionOrRole([]string{"bank_soal.read"}, "admin", "guru")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/bank-soal/questions", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role":        "staf",
		"permissions": []string{"bank_soal.read"},
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("RequireAnyPermissionOrRole(permission) status = %d, want 204", rec.Code)
	}
}

func TestRequireAnyPermissionOrRoleFallsBackToLegacyRole(t *testing.T) {
	middleware := RequireAnyPermissionOrRole([]string{"bank_soal.read"}, "admin", "guru")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/bank-soal/questions", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role": "guru",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("RequireAnyPermissionOrRole(role fallback) status = %d, want 204", rec.Code)
	}
}

func TestRequireAnyPermissionOrRoleRejectsMissingPermissionAndRole(t *testing.T) {
	middleware := RequireAnyPermissionOrRole([]string{"roles.manage"}, "admin")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role":        "guru",
		"permissions": []string{"roles.read"},
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("RequireAnyPermissionOrRole(missing) status = %d, want 403", rec.Code)
	}
}
