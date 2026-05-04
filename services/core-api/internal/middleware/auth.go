package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"mtsn2kolut-super-app/backend/internal/api"
)

type authVersionProvider func(context.Context, string) (int64, error)
type accessSessionValidator func(context.Context, string, string) (bool, error)

func JWT(secret string, currentVersion authVersionProvider, validateSession accessSessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				api.Unauthorized(w)
				return
			}
			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil {
				api.Unauthorized(w)
				return
			}
			if tokenType, _ := claims["type"].(string); tokenType != "access" {
				api.Unauthorized(w)
				return
			}
			if currentVersion != nil {
				ver, ok := claims["ver"].(float64)
				if !ok {
					api.Unauthorized(w)
					return
				}
				sub, _ := claims["sub"].(string)
				if strings.TrimSpace(sub) == "" {
					api.Unauthorized(w)
					return
				}
				want, err := currentVersion(r.Context(), sub)
				if err != nil || int64(ver) != want {
					api.Unauthorized(w)
					return
				}
				if validateSession != nil {
					sessionID, _ := claims["ssid"].(string)
					if strings.TrimSpace(sessionID) == "" {
						api.Unauthorized(w)
						return
					}
					ok, err := validateSession(r.Context(), sub, sessionID)
					if err != nil || !ok {
						api.Unauthorized(w)
						return
					}
				}
			}
			ctx := context.WithValue(r.Context(), api.ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func InternalKeyOrJWT(internalKey, jwtSecret string, currentVersion authVersionProvider, validateSession accessSessionValidator) func(http.Handler) http.Handler {
	jwtMW := JWT(jwtSecret, currentVersion, validateSession)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if internalKey != "" && r.Header.Get("X-Internal-Key") == internalKey {
				next.ServeHTTP(w, r)
				return
			}
			jwtMW(next).ServeHTTP(w, r)
		})
	}
}

// RequireAdmin enforces that the JWT claim "roles" includes "admin" or "role" == "admin".
// Internal-key requests (BFF) bypass this check — the BFF is responsible for
// gating admin-only routes via SvelteKit hooks before forwarding.
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireAnyRole("admin")
}

func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := api.ClaimsFromContext(r.Context())
			if !ok {
				api.Unauthorized(w)
				return
			}
			if !HasAnyRole(claims, roles...) {
				api.Forbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func HasAnyRole(claims jwt.MapClaims, roles ...string) bool {
	for _, expected := range roles {
		if rawRoles, ok := claims["roles"].([]any); ok {
			for _, role := range rawRoles {
				if value, ok := role.(string); ok && value == expected {
					return true
				}
			}
		}
		if rawRoles, ok := claims["roles"].([]string); ok {
			for _, role := range rawRoles {
				if role == expected {
					return true
				}
			}
		}
		if role, _ := claims["role"].(string); role == expected {
			return true
		}
	}
	return false
}

func WorkerKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-Worker-Key")
			if got == "" {
				got = bearerToken(r)
			}
			if subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				api.Unauthorized(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
