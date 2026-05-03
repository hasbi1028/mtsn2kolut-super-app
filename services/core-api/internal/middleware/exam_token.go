package middleware

import (
	"context"
	"net/http"
	"strings"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type examContextKey string

const (
	ExamParticipantKey   examContextKey = "exam_participant"
	DeviceFingerprintHdr string         = "X-Device-Fingerprint"
)

type ParticipantLookup func(ctx context.Context, token string) (db.GetParticipantByTokenRow, error)

// ExamToken validates the X-Exam-Token and device fingerprint headers, then
// injects the participant into context.
func ExamToken(lookup ParticipantLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimSpace(r.Header.Get("X-Exam-Token"))
			if token == "" {
				api.Unauthorized(w)
				return
			}
			p, err := lookup(r.Context(), token)
			if err != nil {
				api.Unauthorized(w)
				return
			}
			if p.SessionStatus != db.CbtSessionStatusEnumActive {
				api.Forbidden(w)
				return
			}
			if !examDeviceMatchesRequest(r, p) {
				api.Err(w, http.StatusConflict, "token already bound to another device")
				return
			}
			ctx := context.WithValue(r.Context(), ExamParticipantKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func examDeviceMatchesRequest(r *http.Request, p db.GetParticipantByTokenRow) bool {
	bound := strings.TrimSpace(p.DeviceFingerprint.String)
	if !p.DeviceFingerprint.Valid || bound == "" {
		return false
	}
	return strings.TrimSpace(r.Header.Get(DeviceFingerprintHdr)) == bound
}

// ParticipantFromContext retrieves the exam participant injected by ExamToken middleware.
func ParticipantFromContext(ctx context.Context) (db.GetParticipantByTokenRow, bool) {
	p, ok := ctx.Value(ExamParticipantKey).(db.GetParticipantByTokenRow)
	return p, ok
}

// ExamTokenOrJWT allows asset/file routes to be accessed either by a logged-in
// admin/guru request (JWT) or by an active exam
// participant using the `exam_token` query parameter.
func ExamTokenOrJWT(
	jwtSecret string,
	currentVersion authVersionProvider,
	validateSession accessSessionValidator,
	lookup ParticipantLookup,
) func(http.Handler) http.Handler {
	jwtOnly := JWT(jwtSecret, currentVersion, validateSession)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := strings.TrimSpace(r.URL.Query().Get("exam_token")); token != "" {
				p, err := lookup(r.Context(), token)
				if err != nil {
					api.Unauthorized(w)
					return
				}
				if p.SessionStatus != db.CbtSessionStatusEnumActive {
					api.Forbidden(w)
					return
				}
				ctx := context.WithValue(r.Context(), ExamParticipantKey, p)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			jwtOnly(next).ServeHTTP(w, r)
		})
	}
}
