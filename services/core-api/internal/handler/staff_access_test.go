package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
)

func TestStaffModuleAccessAllowsAdminAndStaf(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{
			name:   "single admin role",
			claims: jwt.MapClaims{"role": "admin"},
		},
		{
			name:   "single staf role",
			claims: jwt.MapClaims{"role": "staf"},
		},
		{
			name:   "roles array staf",
			claims: jwt.MapClaims{"roles": []any{"guru", "staf"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/library/stats", nil)
			req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, tc.claims))

			if !libraryAccessAllowed(req) {
				t.Fatalf("libraryAccessAllowed() = false, want true")
			}
			if !inventoryAccessAllowed(req) {
				t.Fatalf("inventoryAccessAllowed() = false, want true")
			}
			if !governanceAccessAllowed(req) {
				t.Fatalf("governanceAccessAllowed() = false, want true")
			}
			if !tuAccessAllowed(req) {
				t.Fatalf("tuAccessAllowed() = false, want true")
			}
		})
	}
}

func TestStaffModuleAccessRejectsGuru(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/inventory/items", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"roles": []any{"guru"}}))

	if libraryAccessAllowed(req) {
		t.Fatalf("libraryAccessAllowed() = true, want false")
	}
	if inventoryAccessAllowed(req) {
		t.Fatalf("inventoryAccessAllowed() = true, want false")
	}
	if governanceAccessAllowed(req) {
		t.Fatalf("governanceAccessAllowed() = true, want false")
	}
	if tuAccessAllowed(req) {
		t.Fatalf("tuAccessAllowed() = true, want false")
	}
}

func TestStaffModuleAccessRejectsMissingClaims(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/document-cycles/stats", nil)

	if libraryAccessAllowed(req) {
		t.Fatalf("libraryAccessAllowed() = true, want false")
	}
	if inventoryAccessAllowed(req) {
		t.Fatalf("inventoryAccessAllowed() = true, want false")
	}
	if governanceAccessAllowed(req) {
		t.Fatalf("governanceAccessAllowed() = true, want false")
	}
	if tuAccessAllowed(req) {
		t.Fatalf("tuAccessAllowed() = true, want false")
	}
}
