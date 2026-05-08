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

type examDeviceMatchState int

const (
	examDeviceMatchOK examDeviceMatchState = iota
	examDeviceMatchMissingContext
	examDeviceMatchMismatch
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
			switch examDeviceMatchForRequest(r, p) {
			case examDeviceMatchOK:
			case examDeviceMatchMismatch:
				api.Err(w, http.StatusConflict, "token already bound to another device")
				return
			default:
				api.Unauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), ExamParticipantKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func examDeviceMatchForRequest(r *http.Request, p db.GetParticipantByTokenRow) examDeviceMatchState {
	bound := strings.TrimSpace(p.DeviceFingerprint.String)
	if !p.DeviceFingerprint.Valid || bound == "" {
		return examDeviceMatchMissingContext
	}
	requested := strings.TrimSpace(r.Header.Get(DeviceFingerprintHdr))
	if requested == "" {
		return examDeviceMatchMissingContext
	}
	if requested != bound {
		return examDeviceMatchMismatch
	}
	return examDeviceMatchOK
}

// ParticipantFromContext retrieves the exam participant injected by ExamToken middleware.
func ParticipantFromContext(ctx context.Context) (db.GetParticipantByTokenRow, bool) {
	p, ok := ctx.Value(ExamParticipantKey).(db.GetParticipantByTokenRow)
	return p, ok
}

// ExamTokenOrJWT allows asset/file routes to be accessed either by a logged-in
// admin/guru request (JWT) or by an active exam participant using the same
// header-bound token and device check as operational exam endpoints.
func ExamTokenOrJWT(
	jwtSecret string,
	currentVersion authVersionProvider,
	validateSession accessSessionValidator,
	lookup ParticipantLookup,
) func(http.Handler) http.Handler {
	jwtOnly := JWT(jwtSecret, currentVersion, validateSession)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := strings.TrimSpace(r.Header.Get("X-Exam-Token")); token != "" {
				p, err := lookup(r.Context(), token)
				if err != nil {
					api.Unauthorized(w)
					return
				}
				if p.SessionStatus != db.CbtSessionStatusEnumActive {
					api.Forbidden(w)
					return
				}
				switch examDeviceMatchForRequest(r, p) {
				case examDeviceMatchOK:
				case examDeviceMatchMismatch:
					api.Err(w, http.StatusConflict, "token already bound to another device")
					return
				default:
					api.Unauthorized(w)
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
