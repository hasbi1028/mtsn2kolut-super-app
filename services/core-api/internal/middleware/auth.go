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
				if t.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
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
			if mustChangePassword(claims) && !mustChangePasswordAllowedPath(r.URL.Path) {
				api.Forbidden(w)
				return
			}
			ctx := context.WithValue(r.Context(), api.ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func mustChangePassword(claims jwt.MapClaims) bool {
	value, _ := claims["must_change_password"].(bool)
	return value
}

func mustChangePasswordAllowedPath(path string) bool {
	switch path {
	case "/api/auth/account", "/api/auth/change-password", "/api/auth/logout-all", "/api/auth/sessions":
		return true
	default:
		return strings.HasPrefix(path, "/api/auth/sessions/")
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

func InternalKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-Internal-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				api.Unauthorized(w)
				return
			}
			next.ServeHTTP(w, r)
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

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return RequireAnyPermission(permission)
}

func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := api.ClaimsFromContext(r.Context())
			if !ok {
				api.Unauthorized(w)
				return
			}
			if !HasAnyPermission(claims, permissions...) {
				api.Forbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermissionOrRole is a migration-safe RBAC gate. It prefers dynamic
// permissions, but still accepts legacy role claims while modules are being
// migrated from role-only authorization.
func RequireAnyPermissionOrRole(permissions []string, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := api.ClaimsFromContext(r.Context())
			if !ok {
				api.Unauthorized(w)
				return
			}
			if HasAnyPermission(claims, permissions...) || HasAnyRole(claims, roles...) {
				next.ServeHTTP(w, r)
				return
			}
			api.Forbidden(w)
		})
	}
}

func HasAnyPermission(claims jwt.MapClaims, permissions ...string) bool {
	for _, expected := range permissions {
		expected = strings.TrimSpace(expected)
		if expected == "" {
			continue
		}
		if rawPermissions, ok := claims["permissions"].([]any); ok {
			for _, permission := range rawPermissions {
				if value, ok := permission.(string); ok && value == expected {
					return true
				}
			}
		}
		if rawPermissions, ok := claims["permissions"].([]string); ok {
			for _, permission := range rawPermissions {
				if permission == expected {
					return true
				}
			}
		}
		if permission, _ := claims["permission"].(string); permission == expected {
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
	// RFC 7235 declares the auth scheme name case-insensitive; some clients
	// (and proxies) send "bearer" or "BEARER".
	if len(h) >= 7 && strings.EqualFold(h[:7], "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}
