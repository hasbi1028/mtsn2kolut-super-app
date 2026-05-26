package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

type cbtAccessCardService interface {
	ListParticipantCards(ctx context.Context, in service.CbtAccessCardIssueInput) ([]service.CbtParticipantAccessCard, error)
	IssueParticipantCards(ctx context.Context, in service.CbtAccessCardIssueInput) ([]service.CbtParticipantAccessCard, error)
	ListProctorCards(ctx context.Context, in service.CbtAccessCardIssueInput) ([]service.CbtProctorAccessCard, error)
	IssueProctorCards(ctx context.Context, in service.CbtAccessCardIssueInput) ([]service.CbtProctorAccessCard, error)
	Verify(ctx context.Context, in service.CbtAccessCardVerifyInput) (service.CbtAccessCardVerifyResult, error)
}

type CbtAccessCard struct {
	svc       cbtAccessCardService
	jwtSecret []byte
}

func NewCbtAccessCard(svc *service.CbtAccessCard, jwtSecret string) *CbtAccessCard {
	return &CbtAccessCard{svc: svc, jwtSecret: []byte(jwtSecret)}
}

func (h *CbtAccessCard) ListParticipantCards(w http.ResponseWriter, r *http.Request) {
	in, ok := h.issueInputFromRequest(w, r)
	if !ok {
		return
	}
	cards, err := h.svc.ListParticipantCards(r.Context(), in)
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu peserta ujian tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"cards": cards})
}

func (h *CbtAccessCard) IssueParticipantCards(w http.ResponseWriter, r *http.Request) {
	in, ok := h.issueInputFromRequest(w, r)
	if !ok {
		return
	}
	if !applyIssueBody(w, r, &in) {
		return
	}
	cards, err := h.svc.IssueParticipantCards(r.Context(), in)
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu peserta ujian tidak dapat diterbitkan")
		return
	}
	api.OK(w, map[string]any{"cards": cards})
}

func (h *CbtAccessCard) ListProctorCards(w http.ResponseWriter, r *http.Request) {
	in, ok := h.issueInputFromRequest(w, r)
	if !ok {
		return
	}
	cards, err := h.svc.ListProctorCards(r.Context(), in)
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu pengawas ujian tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"cards": cards})
}

func (h *CbtAccessCard) IssueProctorCards(w http.ResponseWriter, r *http.Request) {
	in, ok := h.issueInputFromRequest(w, r)
	if !ok {
		return
	}
	if !applyIssueBody(w, r, &in) {
		return
	}
	cards, err := h.svc.IssueProctorCards(r.Context(), in)
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu pengawas ujian tidak dapat diterbitkan")
		return
	}
	api.OK(w, map[string]any{"cards": cards})
}

func (h *CbtAccessCard) VerifyParticipantCard(w http.ResponseWriter, r *http.Request) {
	result, ok := h.verifyCard(w, r, "participant")
	if !ok {
		return
	}
	if result.StudentID == nil || result.ParticipantID == nil {
		api.Err(w, http.StatusUnauthorized, "Kartu peserta tidak cocok")
		return
	}
	token, err := h.issueStudentToken(*result.StudentID, result.NISN)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"access_token": token, "card": result})
}

func (h *CbtAccessCard) VerifyProctorCard(w http.ResponseWriter, r *http.Request) {
	result, ok := h.verifyCard(w, r, "proctor")
	if !ok {
		return
	}
	api.OK(w, map[string]any{"card": result})
}

func (h *CbtAccessCard) verifyCard(w http.ResponseWriter, r *http.Request, cardType string) (service.CbtAccessCardVerifyResult, bool) {
	var body struct {
		Token string `json:"token"`
		PIN   string `json:"pin"`
	}
	tokenParam := strings.TrimSpace(chi.URLParam(r, "token"))
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		api.BadRequest(w, "Data kartu tidak valid")
		return service.CbtAccessCardVerifyResult{}, false
	}
	if tokenParam != "" && strings.TrimSpace(body.Token) == "" {
		body.Token = tokenParam
	}
	result, err := h.svc.Verify(r.Context(), service.CbtAccessCardVerifyInput{Token: body.Token, PIN: body.PIN, CardType: cardType, ClientIP: trustedClientIP(r)})
	if err != nil {
		writeCbtAccessCardVerifyError(w, err)
		return service.CbtAccessCardVerifyResult{}, false
	}
	return result, true
}

func (h *CbtAccessCard) issueInputFromRequest(w http.ResponseWriter, r *http.Request) (service.CbtAccessCardIssueInput, bool) {
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "Kegiatan asesmen tidak valid")
		return service.CbtAccessCardIssueInput{}, false
	}
	sessionID, err := parseOptionalUUID(r.URL.Query().Get("session_id"))
	if err != nil {
		api.BadRequest(w, "Sesi asesmen tidak valid")
		return service.CbtAccessCardIssueInput{}, false
	}
	roomID, err := parseOptionalUUID(r.URL.Query().Get("room_id"))
	if err != nil {
		api.BadRequest(w, "Ruang ujian tidak valid")
		return service.CbtAccessCardIssueInput{}, false
	}
	return service.CbtAccessCardIssueInput{EventID: eventID, SessionID: sessionID, RoomID: roomID, GeneratedBy: requestUserID(r)}, true
}

func applyIssueBody(w http.ResponseWriter, r *http.Request, in *service.CbtAccessCardIssueInput) bool {
	var body struct {
		Regenerate   bool  `json:"regenerate"`
		ExpiresHours int32 `json:"expires_hours"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		if err.Error() == "EOF" {
			return true
		}
		api.BadRequest(w, "Data penerbitan kartu tidak valid")
		return false
	}
	in.Regenerate = body.Regenerate
	in.ExpiresHours = body.ExpiresHours
	return true
}

func (h *CbtAccessCard) issueStudentToken(studentID, nisn string) (string, error) {
	now := time.Now()
	claims := cbtPortalClaims{StudentID: studentID, NISN: nisn, RegisteredClaims: jwt.RegisteredClaims{Subject: studentID, Issuer: "mtsn2kolut-cbt-portal", Audience: []string{"cbt-portal"}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(6 * time.Hour))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
}

func requestUserID(r *http.Request) pgtype.UUID {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return pgtype.UUID{}
	}
	id, err := authUserID(claims)
	if err != nil {
		return pgtype.UUID{}
	}
	return id
}

func writeCbtAccessCardVerifyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCbtAccessCardInvalid):
		api.Err(w, http.StatusUnauthorized, "Kartu ujian tidak cocok atau PIN salah. Silakan panggil pengawas.")
	case errors.Is(err, service.ErrCbtAccessCardLocked):
		api.Err(w, http.StatusLocked, "Kartu ujian terkunci karena PIN salah berulang. Silakan hubungi admin.")
	case errors.Is(err, service.ErrCbtAccessCardExpired):
		api.Err(w, http.StatusForbidden, "Kartu ujian sudah kedaluwarsa. Silakan hubungi admin.")
	case errors.Is(err, domain.ErrBadRequest):
		api.BadRequest(w, "Data kartu tidak valid")
	default:
		api.Internal(w, err)
	}
}
