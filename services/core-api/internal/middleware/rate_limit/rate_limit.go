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
	return RateLimitWithTrustedProxies(burst, ratePerSecond, nil)
}

func RateLimitWithTrustedProxies(burst int, ratePerSecond float64, trustedProxyCIDRs []string) func(http.Handler) http.Handler {
	var (
		mu          sync.Mutex
		visitors    = make(map[string]*visitor)
		visitorTTL  = 10 * time.Minute
		lastCleanup time.Time
		trusted     = parseTrustedProxyCIDRs(trustedProxyCIDRs)
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
			ip := clientIP(r, trusted)
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

func clientIP(r *http.Request, trustedProxies []*net.IPNet) string {
	remoteIP := remoteAddrIP(r.RemoteAddr)
	if len(trustedProxies) > 0 && trustedProxy(remoteIP, trustedProxies) {
		if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
			if ip := nearestUntrustedForwardedIP(forwarded, trustedProxies); ip != nil {
				return ip.String()
			}
		}
		if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
			if ip := net.ParseIP(realIP); ip != nil {
				return ip.String()
			}
		}
	}
	if remoteIP != nil {
		return remoteIP.String()
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func nearestUntrustedForwardedIP(forwarded string, trustedProxies []*net.IPNet) net.IP {
	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		ip := net.ParseIP(strings.TrimSpace(parts[i]))
		if ip == nil {
			continue
		}
		if trustedProxy(ip, trustedProxies) {
			continue
		}
		return ip
	}
	return nil
}

func parseTrustedProxyCIDRs(values []string) []*net.IPNet {
	trusted := make([]*net.IPNet, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if ip := net.ParseIP(raw); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			trusted = append(trusted, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}
		_, network, err := net.ParseCIDR(raw)
		if err == nil {
			trusted = append(trusted, network)
		}
	}
	return trusted
}

func remoteAddrIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(strings.TrimSpace(remoteAddr))
}

func trustedProxy(ip net.IP, trustedProxies []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, network := range trustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
