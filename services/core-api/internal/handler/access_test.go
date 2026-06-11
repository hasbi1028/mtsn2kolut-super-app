package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestAcademicReadAccessAllowedAcceptsDynamicPermissions(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{
			name:   "academic read permission without legacy role",
			claims: jwt.MapClaims{"permissions": []any{"academic.read"}},
		},
		{
			name:   "academic manage permission without legacy role",
			claims: jwt.MapClaims{"permissions": []string{"academic.manage"}},
		},
		{
			name:   "legacy guru role remains allowed",
			claims: jwt.MapClaims{"roles": []any{"guru"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withClaims(httptest.NewRequest(http.MethodGet, "/api/academic", nil), tt.claims)
			if !academicReadAccessAllowed(req) {
				t.Fatalf("academicReadAccessAllowed() = false, want true")
			}
		})
	}
}

func TestAcademicReadAccessAllowedRejectsUnrelatedClaims(t *testing.T) {
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/academic", nil), jwt.MapClaims{"permissions": []any{"bank_soal.create"}})
	if academicReadAccessAllowed(req) {
		t.Fatalf("academicReadAccessAllowed() = true, want false for unrelated permission")
	}

	if academicReadAccessAllowed(httptest.NewRequest(http.MethodGet, "/api/academic", nil)) {
		t.Fatalf("academicReadAccessAllowed() = true, want false for missing claims")
	}
}

func TestAcademicSubjectListAccessAllowedAcceptsAcademicPermission(t *testing.T) {
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/academic/subjects", nil), jwt.MapClaims{"permissions": []any{"academic.read"}})
	if !academicSubjectListAccessAllowed(req) {
		t.Fatalf("academicSubjectListAccessAllowed() = false, want true")
	}
}

func TestAcademicSubjectListAccessAllowedRejectsMissingClaims(t *testing.T) {
	if academicSubjectListAccessAllowed(httptest.NewRequest(http.MethodGet, "/api/academic/subjects", nil)) {
		t.Fatalf("academicSubjectListAccessAllowed() = true, want false for missing claims")
	}
}
