package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func guruRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"guru"},
		"role":  "guru",
		"eid":   "02000000-0000-0000-0000-000000000000",
	})
}

func requireAnySlice(t *testing.T, value any, key string) []any {
	t.Helper()
	got, ok := value.([]any)
	if !ok {
		t.Fatalf("%s = %#v (%T), want []any", key, value, value)
	}
	return got
}
