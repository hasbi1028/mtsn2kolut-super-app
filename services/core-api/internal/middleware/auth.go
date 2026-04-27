package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"mtsn2kolut-super-app/backend/internal/api"
)

type authVersionProvider func(context.Context) (int64, error)

func JWT(secret string, currentVersion authVersionProvider) func(http.Handler) http.Handler {
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
				want, err := currentVersion(r.Context())
				if err != nil || int64(ver) != want {
					api.Unauthorized(w)
					return
				}
			}
			ctx := context.WithValue(r.Context(), api.ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func InternalKeyOrJWT(internalKey, jwtSecret string, currentVersion authVersionProvider) func(http.Handler) http.Handler {
	jwtMW := JWT(jwtSecret, currentVersion)
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

func WorkerKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-Worker-Key")
			if got == "" {
				got = bearerToken(r)
			}
			if got != key {
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
