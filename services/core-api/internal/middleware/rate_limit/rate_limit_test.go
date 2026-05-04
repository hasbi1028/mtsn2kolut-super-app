package rate_limit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitBlocksAfterBurst(t *testing.T) {
	middleware := RateLimit(1, 0)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("first request status = %d, want %d", rec1.Code, http.StatusNoContent)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitUsesForwardedIPOnlyFromTrustedProxy(t *testing.T) {
	middleware := RateLimitWithTrustedProxies(1, 0, []string{"10.0.0.0/24"})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	req1.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("forwarded request status = %d, want %d", rec1.Code, http.StatusNoContent)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.2:2222"
	req2.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.2")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second forwarded request status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitUsesNearestUntrustedForwardedIPFromTrustedProxy(t *testing.T) {
	middleware := RateLimitWithTrustedProxies(1, 0, []string{"10.0.0.0/24"})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	req1.Header.Set("X-Forwarded-For", "198.51.100.200, 203.0.113.10, 10.0.0.1")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("first spoofed forwarded request status = %d, want %d", rec1.Code, http.StatusNoContent)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.2:2222"
	req2.Header.Set("X-Forwarded-For", "198.51.100.201, 203.0.113.10, 10.0.0.2")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second spoofed forwarded request status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitIgnoresForwardedIPFromUntrustedRemote(t *testing.T) {
	middleware := RateLimitWithTrustedProxies(1, 0, []string{"10.0.0.0/24"})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "198.51.100.1:1111"
	req1.Header.Set("X-Forwarded-For", "203.0.113.10")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusNoContent {
		t.Fatalf("first untrusted request status = %d, want %d", rec1.Code, http.StatusNoContent)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "198.51.100.2:2222"
	req2.Header.Set("X-Forwarded-For", "203.0.113.10")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("second untrusted request status = %d, want forwarded header ignored", rec2.Code)
	}
}

func TestClientIPFallsBackThroughHeadersAndRemoteAddress(t *testing.T) {
	tests := []struct {
		name       string
		forwarded  string
		realIP     string
		remoteAddr string
		trusted    []string
		want       string
	}{
		{
			name:       "forwarded nearest untrusted hop",
			forwarded:  " 203.0.113.20 , 10.0.0.1 ",
			remoteAddr: "10.0.0.1:1234",
			trusted:    []string{"10.0.0.1"},
			want:       "203.0.113.20",
		},
		{
			name:       "forwarded ignores spoofed leftmost hop",
			forwarded:  "198.51.100.200, 203.0.113.20, 10.0.0.1",
			remoteAddr: "10.0.0.1:1234",
			trusted:    []string{"10.0.0.1"},
			want:       "203.0.113.20",
		},
		{
			name:       "real ip fallback",
			realIP:     " 203.0.113.21 ",
			remoteAddr: "10.0.0.1:1234",
			trusted:    []string{"10.0.0.1"},
			want:       "203.0.113.21",
		},
		{
			name:       "host port remote",
			remoteAddr: "192.0.2.44:5678",
			want:       "192.0.2.44",
		},
		{
			name:       "malformed remote fallback",
			remoteAddr: " 192.0.2.45 ",
			want:       "192.0.2.45",
		},
		{
			name: "empty",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			if got := clientIP(req, parseTrustedProxyCIDRs(tt.trusted)); got != tt.want {
				t.Fatalf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
