package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
)

func withKesiswaanClaims(req *http.Request, claims jwt.MapClaims) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
}

func TestGetKesiswaanAccessAllowsAdminAndKesiswaanManage(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{name: "admin role", claims: jwt.MapClaims{"role": "admin"}},
		{name: "kesiswaan role array", claims: jwt.MapClaims{"roles": []any{"guru", "kesiswaan"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withKesiswaanClaims(httptest.NewRequest(http.MethodGet, "/api/kesiswaan/stats", nil), tt.claims)
			access := getKesiswaanAccess(req)
			if !access.canRead {
				t.Fatal("canRead = false, want true")
			}
			if !access.canManage {
				t.Fatal("canManage = false, want true")
			}
		})
	}
}

func TestGetKesiswaanAccessScopesGuruByEmployeeID(t *testing.T) {
	req := withKesiswaanClaims(httptest.NewRequest(http.MethodGet, "/api/kesiswaan/stats", nil), jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   "11111111-1111-1111-1111-111111111111",
	})
	access := getKesiswaanAccess(req)
	if !access.canRead {
		t.Fatal("canRead = false, want true")
	}
	if access.canManage {
		t.Fatal("canManage = true, want false")
	}
	if !access.teacherEmployeeID.Valid {
		t.Fatal("teacherEmployeeID.Valid = false, want true")
	}
}

func TestGetKesiswaanAccessRejectsGuruWithoutEmployeeID(t *testing.T) {
	req := withKesiswaanClaims(httptest.NewRequest(http.MethodGet, "/api/kesiswaan/stats", nil), jwt.MapClaims{
		"roles": []any{"guru"},
	})
	access := getKesiswaanAccess(req)
	if access.canRead {
		t.Fatal("canRead = true, want false")
	}
	if access.canManage {
		t.Fatal("canManage = true, want false")
	}
}

func TestKesiswaanStatsForbiddenForGuruWithoutScopedEmployeeID(t *testing.T) {
	h := NewKesiswaan(nil)
	req := withKesiswaanClaims(httptest.NewRequest(http.MethodGet, "/api/kesiswaan/stats", nil), jwt.MapClaims{
		"roles": []any{"guru"},
	})
	rec := httptest.NewRecorder()

	h.Stats(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestKesiswaanUpdateStudentProfileForbiddenForGuruReadOnlyScope(t *testing.T) {
	h := NewKesiswaan(nil)
	req := withKesiswaanClaims(httptest.NewRequest(http.MethodPut, "/api/kesiswaan/students/11111111-1111-1111-1111-111111111111/profile", nil), jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   "22222222-2222-2222-2222-222222222222",
	})
	rec := httptest.NewRecorder()

	h.UpdateStudentProfile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestKesiswaanStatsAdminGuruRemainsAdminWide(t *testing.T) {
	svc := &fakeKesiswaanService{}
	h := &Kesiswaan{svc: svc}
	req := withKesiswaanClaims(httptest.NewRequest(http.MethodGet, "/api/kesiswaan/stats", nil), jwt.MapClaims{
		"roles": []any{"admin", "guru"},
		"eid":   "33333333-3333-3333-3333-333333333333",
	})
	rec := httptest.NewRecorder()

	h.Stats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.statsTeacher.Valid {
		t.Fatalf("stats teacher scope = %v, want invalid admin-wide scope", svc.statsTeacher)
	}
}
