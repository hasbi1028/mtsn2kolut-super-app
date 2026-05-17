package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
)

func websiteActorRequest(claims jwt.MapClaims) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/website/content", nil)
	return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
}

func TestWebsiteActorUsernamePrefersUserClaimAndFallsBackToSubject(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
		want   string
	}{
		{name: "usr preferred over sub", claims: jwt.MapClaims{"usr": "editor", "sub": "user-id"}, want: "editor"},
		{name: "sub fallback when usr missing", claims: jwt.MapClaims{"sub": "user-id"}, want: "user-id"},
		{name: "sub fallback when usr empty", claims: jwt.MapClaims{"usr": "", "sub": "user-id"}, want: "user-id"},
		{name: "non-string usr ignored", claims: jwt.MapClaims{"usr": 123, "sub": "user-id"}, want: "user-id"},
		{name: "non-string sub ignored", claims: jwt.MapClaims{"sub": 123}, want: ""},
		{name: "no claims", claims: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/website/content", nil)
			if tt.claims != nil {
				req = websiteActorRequest(tt.claims)
			}
			if got := websiteActorUsername(req); got != tt.want {
				t.Fatalf("websiteActorUsername() = %q, want %q", got, tt.want)
			}
		})
	}
}
