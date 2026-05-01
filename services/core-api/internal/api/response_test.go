package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestClaimsFromContext(t *testing.T) {
	claims := jwt.MapClaims{"sub": "user-1", "role": "admin"}
	ctx := context.WithValue(context.Background(), ClaimsKey, claims)

	got, ok := ClaimsFromContext(ctx)
	if !ok {
		t.Fatal("ClaimsFromContext() ok = false, want true")
	}
	if got["sub"] != "user-1" || got["role"] != "admin" {
		t.Fatalf("ClaimsFromContext() = %v, want user claims", got)
	}

	if got, ok := ClaimsFromContext(context.Background()); ok || got != nil {
		t.Fatalf("ClaimsFromContext(empty) = %v, %v; want nil, false", got, ok)
	}
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		name       string
		call       func(http.ResponseWriter)
		wantStatus int
		wantBody   string
	}{
		{
			name: "ok",
			call: func(w http.ResponseWriter) {
				OK(w, struct {
					Status string `json:"status"`
				}{Status: "ok"})
			},
			wantStatus: http.StatusOK,
			wantBody:   "{\"data\":{\"status\":\"ok\"}}\n",
		},
		{
			name: "created",
			call: func(w http.ResponseWriter) {
				Created(w, struct {
					ID int `json:"id"`
				}{ID: 7})
			},
			wantStatus: http.StatusCreated,
			wantBody:   "{\"data\":{\"id\":7}}\n",
		},
		{
			name:       "bad request",
			call:       func(w http.ResponseWriter) { BadRequest(w, "payload salah") },
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"payload salah\"}\n",
		},
		{
			name:       "conflict",
			call:       func(w http.ResponseWriter) { Conflict(w, "duplikat") },
			wantStatus: http.StatusConflict,
			wantBody:   "{\"error\":\"duplikat\"}\n",
		},
		{
			name:       "unauthorized",
			call:       func(w http.ResponseWriter) { Unauthorized(w) },
			wantStatus: http.StatusUnauthorized,
			wantBody:   "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:       "forbidden",
			call:       func(w http.ResponseWriter) { Forbidden(w) },
			wantStatus: http.StatusForbidden,
			wantBody:   "{\"error\":\"forbidden\"}\n",
		},
		{
			name:       "not found",
			call:       func(w http.ResponseWriter) { NotFound(w) },
			wantStatus: http.StatusNotFound,
			wantBody:   "{\"error\":\"not found\"}\n",
		},
		{
			name:       "internal nil error",
			call:       func(w http.ResponseWriter) { Internal(w, nil) },
			wantStatus: http.StatusInternalServerError,
			wantBody:   "{\"error\":\"internal server error\"}\n",
		},
		{
			name:       "internal non nil error",
			call:       func(w http.ResponseWriter) { Internal(w, errors.New("db down")) },
			wantStatus: http.StatusInternalServerError,
			wantBody:   "{\"error\":\"internal server error\"}\n",
		},
		{
			name:       "too many requests",
			call:       func(w http.ResponseWriter) { TooManyRequests(w) },
			wantStatus: http.StatusTooManyRequests,
			wantBody:   "{\"error\":\"too many requests\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.call(rec)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content-type = %q, want application/json", got)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	NoContent(rec)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("NoContent() status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("NoContent() body length = %d, want 0", rec.Body.Len())
	}
}
