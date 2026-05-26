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
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
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
	svc        cbtAccessCardService
	jwtSecret  []byte
	sessionSvc cbtProctorPortalSessionService
}

type cbtProctorPortalSessionService interface {
	UpdateStatus(ctx context.Context, id pgtype.UUID, status db.CbtSessionStatusEnum) (db.CbtExamSession, error)
	GetRoomProctoringDashboard(ctx context.Context, roomID pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error)
	ListRoomProctors(ctx context.Context, roomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error)
	GetProctoringStatusForRoom(ctx context.Context, sessionID, roomID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error)
	ListParticipantEventsForRoom(ctx context.Context, sessionID, participantID, roomID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error)
}

type cbtProctorPortalControlService interface {
	UnlockParticipantAntiCheat(ctx context.Context, participantID pgtype.UUID, actor, notes string) (db.UnlockParticipantAntiCheatRow, error)
	AcknowledgeProctorEvent(ctx context.Context, participantID pgtype.UUID, eventID, actor, notes string) error
	RecordIncidentAction(ctx context.Context, participantID pgtype.UUID, eventID, action, actor, notes string) error
	SendParticipantCommand(ctx context.Context, participantID pgtype.UUID, commandType, message, actor string) error
}

func NewCbtAccessCard(svc *service.CbtAccessCard, jwtSecret string) *CbtAccessCard {
	return &CbtAccessCard{svc: svc, jwtSecret: []byte(jwtSecret)}
}

func (h *CbtAccessCard) WithProctorPortalSessionService(svc cbtProctorPortalSessionService) *CbtAccessCard {
	h.sessionSvc = svc
	return h
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

func (h *CbtAccessCard) ProctorPortalDashboard(w http.ResponseWriter, r *http.Request) {
	result, sessionID, roomID, ok := h.requireVerifiedProctorPortal(w, r)
	if !ok {
		return
	}
	room, err := h.sessionSvc.GetRoomProctoringDashboard(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	proctors, err := h.sessionSvc.ListRoomProctors(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	participants, err := h.sessionSvc.GetProctoringStatusForRoom(r.Context(), sessionID, roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	events, err := h.sessionSvc.ListParticipantEventsForRoom(r.Context(), sessionID, pgtype.UUID{}, roomID, 50)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"card":         result,
		"room":         room,
		"proctors":     proctors,
		"participants": serializeProctoringRows(participants, false),
		"events":       events,
	})
}

func (h *CbtAccessCard) ProctorPortalAcknowledge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token   string `json:"token"`
		PIN     string `json:"pin"`
		EventID string `json:"event_id"`
		Notes   string `json:"notes"`
	}
	if !decodeProctorPortalBody(w, r, &body) {
		return
	}
	participantID, ok := h.requirePortalParticipant(w, r, body.Token, body.PIN)
	if !ok {
		return
	}
	control, ok := h.proctorPortalControl(w)
	if !ok {
		return
	}
	eventID := strings.TrimSpace(body.EventID)
	if eventID == "" {
		api.BadRequest(w, "Event pengawasan wajib dipilih")
		return
	}
	if err := control.AcknowledgeProctorEvent(r.Context(), participantID, eventID, "portal_pengawas", strings.TrimSpace(body.Notes)); err != nil {
		writeDomainOrInternal(w, err, "Kejadian belum dapat ditandai diperiksa")
		return
	}
	api.OK(w, map[string]string{"status": "acknowledged"})
}

func (h *CbtAccessCard) ProctorPortalIncidentAction(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token   string `json:"token"`
		PIN     string `json:"pin"`
		EventID string `json:"event_id"`
		Action  string `json:"action"`
		Notes   string `json:"notes"`
	}
	if !decodeProctorPortalBody(w, r, &body) {
		return
	}
	participantID, ok := h.requirePortalParticipant(w, r, body.Token, body.PIN)
	if !ok {
		return
	}
	control, ok := h.proctorPortalControl(w)
	if !ok {
		return
	}
	eventID := strings.TrimSpace(body.EventID)
	action := strings.TrimSpace(body.Action)
	if eventID == "" || action == "" {
		api.BadRequest(w, "Event dan tindakan pengawasan wajib diisi")
		return
	}
	if err := control.RecordIncidentAction(r.Context(), participantID, eventID, action, "portal_pengawas", strings.TrimSpace(body.Notes)); err != nil {
		writeDomainOrInternal(w, err, "Tindakan insiden belum dapat disimpan")
		return
	}
	api.OK(w, map[string]string{"status": "recorded"})
}

func (h *CbtAccessCard) ProctorPortalCommand(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token       string `json:"token"`
		PIN         string `json:"pin"`
		CommandType string `json:"command_type"`
		Message     string `json:"message"`
	}
	if !decodeProctorPortalBody(w, r, &body) {
		return
	}
	participantID, ok := h.requirePortalParticipant(w, r, body.Token, body.PIN)
	if !ok {
		return
	}
	control, ok := h.proctorPortalControl(w)
	if !ok {
		return
	}
	commandType := strings.TrimSpace(body.CommandType)
	message := strings.TrimSpace(body.Message)
	if commandType == "" || message == "" {
		api.BadRequest(w, "Jenis instruksi dan pesan wajib diisi")
		return
	}
	if err := control.SendParticipantCommand(r.Context(), participantID, commandType, message, "portal_pengawas"); err != nil {
		writeDomainOrInternal(w, err, "Instruksi ke aplikasi siswa belum dapat dikirim")
		return
	}
	api.OK(w, map[string]string{"status": "sent"})
}

func (h *CbtAccessCard) ProctorPortalUnlock(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
		PIN   string `json:"pin"`
		Notes string `json:"notes"`
	}
	if !decodeProctorPortalBody(w, r, &body) {
		return
	}
	participantID, ok := h.requirePortalParticipant(w, r, body.Token, body.PIN)
	if !ok {
		return
	}
	control, ok := h.proctorPortalControl(w)
	if !ok {
		return
	}
	unlocked, err := control.UnlockParticipantAntiCheat(r.Context(), participantID, "portal_pengawas", strings.TrimSpace(body.Notes))
	if err != nil {
		writeDomainOrInternal(w, err, "Kunci peserta belum dapat dibuka")
		return
	}
	api.OK(w, map[string]any{"status": "unlocked", "participant": unlocked})
}

func (h *CbtAccessCard) requirePortalParticipant(w http.ResponseWriter, r *http.Request, token, pin string) (pgtype.UUID, bool) {
	_, sessionID, roomID, ok := h.verifyProctorPortalCredentials(w, r, token, pin)
	if !ok {
		return pgtype.UUID{}, false
	}
	participantID, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "Peserta pengawasan tidak valid")
		return pgtype.UUID{}, false
	}
	participants, err := h.sessionSvc.GetProctoringStatusForRoom(r.Context(), sessionID, roomID)
	if err != nil {
		api.Internal(w, err)
		return pgtype.UUID{}, false
	}
	for _, row := range participants {
		if row.ParticipantID == participantID {
			return participantID, true
		}
	}
	api.Forbidden(w)
	return pgtype.UUID{}, false
}

func (h *CbtAccessCard) proctorPortalControl(w http.ResponseWriter) (cbtProctorPortalControlService, bool) {
	control, ok := h.sessionSvc.(cbtProctorPortalControlService)
	if !ok {
		api.Internal(w, errors.New("proctor portal control service unavailable"))
		return nil, false
	}
	return control, true
}

func (h *CbtAccessCard) ProctorPortalUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token  string `json:"token"`
		PIN    string `json:"pin"`
		Status string `json:"status"`
	}
	if !decodeProctorPortalBody(w, r, &body) {
		return
	}
	_, sessionID, _, ok := h.verifyProctorPortalCredentials(w, r, body.Token, body.PIN)
	if !ok {
		return
	}
	status := db.CbtSessionStatusEnum(strings.TrimSpace(body.Status))
	if status != db.CbtSessionStatusEnumActive && status != db.CbtSessionStatusEnumFinished {
		api.BadRequest(w, "Status portal pengawasan tidak valid")
		return
	}
	updated, err := h.sessionSvc.UpdateStatus(r.Context(), sessionID, status)
	if err != nil {
		writeDomainOrInternal(w, err, "Status sesi tidak dapat diperbarui")
		return
	}
	api.OK(w, map[string]any{"session": updated, "status": updated.Status})
}

func (h *CbtAccessCard) requireVerifiedProctorPortal(w http.ResponseWriter, r *http.Request) (service.CbtAccessCardVerifyResult, pgtype.UUID, pgtype.UUID, bool) {
	var body struct {
		Token string `json:"token"`
		PIN   string `json:"pin"`
	}
	if !decodeProctorPortalBody(w, r, &body) {
		return service.CbtAccessCardVerifyResult{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	return h.verifyProctorPortalCredentials(w, r, body.Token, body.PIN)
}

func (h *CbtAccessCard) verifyProctorPortalCredentials(w http.ResponseWriter, r *http.Request, token, pin string) (service.CbtAccessCardVerifyResult, pgtype.UUID, pgtype.UUID, bool) {
	if h.sessionSvc == nil {
		api.Internal(w, errors.New("proctor portal session service unavailable"))
		return service.CbtAccessCardVerifyResult{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	result, err := h.svc.Verify(r.Context(), service.CbtAccessCardVerifyInput{Token: token, PIN: pin, CardType: "proctor", ClientIP: trustedClientIP(r)})
	if err != nil {
		writeCbtAccessCardVerifyError(w, err)
		return service.CbtAccessCardVerifyResult{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	if result.RoomID == nil || strings.TrimSpace(*result.RoomID) == "" {
		api.Err(w, http.StatusUnauthorized, "Lembar pengawas ruang tidak cocok")
		return service.CbtAccessCardVerifyResult{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	sessionID, err := parseUUID(result.SessionID)
	if err != nil {
		api.BadRequest(w, "Sesi pengawasan tidak valid")
		return service.CbtAccessCardVerifyResult{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	roomID, err := parseUUID(*result.RoomID)
	if err != nil {
		api.BadRequest(w, "Ruang pengawasan tidak valid")
		return service.CbtAccessCardVerifyResult{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	return result, sessionID, roomID, true
}

func decodeProctorPortalBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		api.BadRequest(w, "Data portal pengawasan tidak valid")
		return false
	}
	return true
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
