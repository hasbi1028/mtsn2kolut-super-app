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

type cbtPortalService interface {
	LoginByNISN(ctx context.Context, nisn, code, clientIP string) (service.CbtPortalLoginResult, error)
	ScheduleByStudentID(ctx context.Context, studentID pgtype.UUID) ([]service.CbtPortalScheduleItem, error)
	AssertCanStart(ctx context.Context, studentID, participantID pgtype.UUID, clientIP string) error
}

type cbtPortalExamService interface {
	LoginCbtPortalDirect(ctx context.Context, participantID, studentID pgtype.UUID, deviceFingerprint, loginIP, browserFingerprint, userAgent string) (service.LoginResult, error)
	SubmitCbtPortalAnswer(ctx context.Context, participantID, studentID, questionID pgtype.UUID, answer string) error
	SubmitCbtPortalExam(ctx context.Context, participantID, studentID pgtype.UUID) error
	Heartbeat(ctx context.Context, participantID pgtype.UUID) error
	RecordClientEvent(ctx context.Context, participantID pgtype.UUID, eventType string, data map[string]any) error
	ListPendingCommands(ctx context.Context, participantID pgtype.UUID) ([]service.ParticipantCommand, error)
	AcknowledgeCommand(ctx context.Context, participantID pgtype.UUID, commandID, status string) error
}

type CbtPortal struct {
	svc       cbtPortalService
	examSvc   cbtPortalExamService
	jwtSecret []byte
}

type cbtPortalClaims struct {
	StudentID string `json:"student_id"`
	NISN      string `json:"nisn"`
	jwt.RegisteredClaims
}

func NewCbtPortal(svc *service.CbtPortal, examSvc *service.Exam, jwtSecret string) *CbtPortal {
	return &CbtPortal{svc: svc, examSvc: examSvc, jwtSecret: []byte(jwtSecret)}
}

func (h *CbtPortal) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NISN string `json:"nisn"`
		Code string `json:"code"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		api.BadRequest(w, "Data login tidak valid")
		return
	}
	result, err := h.svc.LoginByNISN(r.Context(), body.NISN, body.Code, trustedClientIP(r))
	if err != nil {
		if errors.Is(err, service.ErrCbtPortalLoginInvalid) {
			api.Err(w, http.StatusUnauthorized, "NISN atau kode masuk tidak sesuai")
			return
		}
		writeDomainOrInternal(w, err, "Login CBT tidak dapat diproses")
		return
	}
	token, err := h.issueStudentToken(result.Student.ID, result.Student.NISN)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"access_token": token,
		"student":      result.Student,
		"schedule":     result.Schedule,
	})
}

func (h *CbtPortal) Schedule(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.studentIDFromBearer(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ScheduleByStudentID(r.Context(), studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal CBT tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": items})
}

func (h *CbtPortal) Start(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.studentIDFromBearer(w, r)
	if !ok {
		return
	}
	participantID, err := parseUUID(strings.TrimSpace(chi.URLParam(r, "participantID")))
	if err != nil {
		api.BadRequest(w, "Peserta ujian tidak valid")
		return
	}
	clientIP := trustedClientIP(r)
	if err := h.svc.AssertCanStart(r.Context(), studentID, participantID, clientIP); err != nil {
		writeDomainOrInternal(w, err, "Ujian belum dapat dimulai")
		return
	}
	var body struct {
		DeviceFingerprint  string `json:"device_fingerprint"`
		BrowserFingerprint string `json:"browser_fingerprint"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil && err.Error() != "EOF" {
		api.BadRequest(w, "Data perangkat tidak valid")
		return
	}
	result, err := h.examSvc.LoginCbtPortalDirect(r.Context(), participantID, studentID, body.DeviceFingerprint, clientIP, body.BrowserFingerprint, r.UserAgent())
	if err != nil {
		writeCbtPortalExamStartError(w, err)
		return
	}
	api.OK(w, result)
}

func (h *CbtPortal) Answer(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.studentIDFromBearer(w, r)
	if !ok {
		return
	}
	participantID, err := parseUUID(strings.TrimSpace(chi.URLParam(r, "participantID")))
	if err != nil {
		api.BadRequest(w, "Peserta ujian tidak valid")
		return
	}
	var body struct {
		QuestionID string `json:"question_id"`
		Answer     string `json:"answer"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		api.BadRequest(w, "Jawaban tidak valid")
		return
	}
	questionID, err := parseUUID(strings.TrimSpace(body.QuestionID))
	if err != nil {
		api.BadRequest(w, "Soal tidak valid")
		return
	}
	if err := h.examSvc.SubmitCbtPortalAnswer(r.Context(), participantID, studentID, questionID, body.Answer); err != nil {
		writeCbtPortalExamStartError(w, err)
		return
	}
	api.OK(w, map[string]any{"saved": true})
}

func (h *CbtPortal) Commands(w http.ResponseWriter, r *http.Request) {
	_, participantID, ok := h.portalParticipantFromBearer(w, r)
	if !ok {
		return
	}
	commands, err := h.examSvc.ListPendingCommands(r.Context(), participantID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"commands": commands})
}

func (h *CbtPortal) AcknowledgeCommand(w http.ResponseWriter, r *http.Request) {
	_, participantID, ok := h.portalParticipantFromBearer(w, r)
	if !ok {
		return
	}
	commandID := strings.TrimSpace(chi.URLParam(r, "cid"))
	if commandID == "" {
		api.BadRequest(w, "command id required")
		return
	}
	status := "seen"
	var body struct {
		Status string `json:"status"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil && err.Error() != "EOF" {
		api.BadRequest(w, "Data status instruksi tidak valid")
		return
	}
	if trimmed := strings.TrimSpace(body.Status); trimmed != "" {
		status = trimmed
	}
	if err := h.examSvc.AcknowledgeCommand(r.Context(), participantID, commandID, status); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": status})
}

func (h *CbtPortal) Heartbeat(w http.ResponseWriter, r *http.Request) {
	_, participantID, ok := h.portalParticipantFromBearer(w, r)
	if !ok {
		return
	}
	if err := h.examSvc.Heartbeat(r.Context(), participantID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

func (h *CbtPortal) Event(w http.ResponseWriter, r *http.Request) {
	_, participantID, ok := h.portalParticipantFromBearer(w, r)
	if !ok {
		return
	}
	var body struct {
		EventType string         `json:"event_type"`
		Data      map[string]any `json:"data"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		api.BadRequest(w, "Data event perangkat tidak valid")
		return
	}
	eventType := strings.TrimSpace(body.EventType)
	if eventType == "" {
		api.BadRequest(w, "event_type required")
		return
	}
	if err := h.examSvc.RecordClientEvent(r.Context(), participantID, eventType, body.Data); err != nil {
		if errors.Is(err, service.ErrExamInvalidTelemetry) {
			api.BadRequest(w, "event_type tidak didukung")
			return
		}
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "recorded"})
}

func (h *CbtPortal) Submit(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.studentIDFromBearer(w, r)
	if !ok {
		return
	}
	participantID, err := parseUUID(strings.TrimSpace(chi.URLParam(r, "participantID")))
	if err != nil {
		api.BadRequest(w, "Peserta ujian tidak valid")
		return
	}
	if err := h.examSvc.SubmitCbtPortalExam(r.Context(), participantID, studentID); err != nil {
		writeCbtPortalExamStartError(w, err)
		return
	}
	api.OK(w, map[string]any{"submitted": true})
}

func (h *CbtPortal) portalParticipantFromBearer(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	studentID, ok := h.studentIDFromBearer(w, r)
	if !ok {
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	participantID, err := parseUUID(strings.TrimSpace(chi.URLParam(r, "participantID")))
	if err != nil {
		api.BadRequest(w, "Peserta ujian tidak valid")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	items, err := h.svc.ScheduleByStudentID(r.Context(), studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal CBT tidak tersedia")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	for _, item := range items {
		if strings.TrimSpace(item.ParticipantID) == pgUUIDString(participantID) {
			return studentID, participantID, true
		}
	}
	api.Forbidden(w)
	return pgtype.UUID{}, pgtype.UUID{}, false
}

func (h *CbtPortal) issueStudentToken(studentID, nisn string) (string, error) {
	now := time.Now()
	claims := cbtPortalClaims{
		StudentID: studentID,
		NISN:      nisn,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   studentID,
			Issuer:    "mtsn2kolut-cbt-portal",
			Audience:  []string{"cbt-portal"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(6 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.jwtSecret)
}

func (h *CbtPortal) studentIDFromBearer(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	tokenString := strings.TrimSpace(auth[len("Bearer "):])
	claims := &cbtPortalClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrUnauthorized
		}
		return h.jwtSecret, nil
	}, jwt.WithAudience("cbt-portal"), jwt.WithIssuer("mtsn2kolut-cbt-portal"))
	if err != nil || !token.Valid || strings.TrimSpace(claims.StudentID) == "" {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	studentID, err := parseUUID(claims.StudentID)
	if err != nil {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	return studentID, true
}

func writeCbtPortalExamStartError(w http.ResponseWriter, err error) {
	switch err {
	case service.ErrExamNotFound:
		api.Err(w, http.StatusNotFound, "ujian tidak ditemukan")
	case service.ErrExamNotActive:
		api.Err(w, http.StatusForbidden, "sesi ujian belum aktif")
	case service.ErrExamNotStarted:
		api.Err(w, http.StatusForbidden, "sesi ujian belum dimulai")
	case service.ErrExamAlreadySubmit:
		api.Err(w, http.StatusConflict, "ujian sudah dikumpulkan")
	case service.ErrExamLocked:
		api.Err(w, http.StatusLocked, "ujian terkunci oleh kebijakan pengawasan")
	case service.ErrExamWindowClosed:
		api.Err(w, http.StatusForbidden, "waktu ujian sudah berakhir")
	case service.ErrDeviceMismatch:
		api.Err(w, http.StatusConflict, "akun ujian sudah terikat perangkat lain")
	case service.ErrDeviceRequired:
		api.BadRequest(w, "identitas perangkat tidak tersedia")
	case service.ErrWebFallbackDisabled, service.ErrCbtPortalDisabled:
		api.Err(w, http.StatusForbidden, "portal CBT belum diizinkan untuk sesi ini")
	default:
		api.Internal(w, err)
	}
}
