package rate_limit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"mtsn2kolut-super-app/backend/internal/api"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimit returns a middleware that limits requests per client IP.
// It keeps a small in-memory visitor map with TTL-based cleanup to avoid
// unbounded growth during long-running uptime.
func RateLimit(burst int, ratePerSecond float64) func(http.Handler) http.Handler {
	var (
		mu          sync.Mutex
		visitors    = make(map[string]*visitor)
		visitorTTL  = 10 * time.Minute
		lastCleanup time.Time
	)

	getLimiter := func(ip string) *rate.Limiter {
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		if now.Sub(lastCleanup) >= time.Minute {
			for key, item := range visitors {
				if now.Sub(item.lastSeen) > visitorTTL {
					delete(visitors, key)
				}
			}
			lastCleanup = now
		}

		if item, ok := visitors[ip]; ok {
			item.lastSeen = now
			return item.limiter
		}

		limiter := rate.NewLimiter(rate.Limit(ratePerSecond), burst)
		visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: now,
		}
		return limiter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if ip == "" {
				ip = r.RemoteAddr
			}
			if !getLimiter(ip).Allow() {
				api.TooManyRequests(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
