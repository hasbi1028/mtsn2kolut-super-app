package rate_limit

import (
	"net/http"

	"golang.org/x/time/rate"

	"mtsn2kolut-super-app/backend/internal/api"
)

// RateLimit returns a middleware that limits requests per IP.
// burst is the maximum number of tokens available in a token bucket.
// ratePerSecond is the rate at which new tokens are added to the bucket per second.
func RateLimit(burst int, ratePerSecond float64) func(http.Handler) http.Handler {
	// Create a limiter per IP
	var limiterPerIP = make(map[string]*rate.Limiter)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			// Create or get limiter for this IP
			var limiter *rate.Limiter
			if lim, exists := limiterPerIP[ip]; !exists {
				limiter = rate.NewLimiter(rate.Limit(ratePerSecond), burst)
				limiterPerIP[ip] = limiter
			} else {
				// In a production environment, you'd want to clean up old limiters
				// This is a simplified version for demonstration
				limiter = lim
			}

			// Check if request is allowed
			if !limiter.Allow() {
				api.TooManyRequests(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}