package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type studentIDCardPortalAuthService interface {
	LoginStudentByID(ctx context.Context, studentID pgtype.UUID, meta service.SessionMeta) (domain.TokenPair, error)
}

type StudentIDCard struct {
	svc  studentIDCardService
	auth studentIDCardPortalAuthService
}

type studentIDCardService interface {
	List(ctx context.Context, limit, offset int32) ([]db.ListStudentIDCardsRow, error)
	Get(ctx context.Context, id pgtype.UUID) (db.GetStudentIDCardRow, error)
	Issue(ctx context.Context, studentID, actorID pgtype.UUID, baseURL string) (service.StudentIDCardIssueResult, error)
	UpdateStatus(ctx context.Context, id, actorID pgtype.UUID, status, reason string) (db.StudentIDCard, error)
	MarkPrinted(ctx context.Context, id, actorID pgtype.UUID) (db.StudentIDCard, error)
	Reissue(ctx context.Context, oldID, actorID pgtype.UUID, baseURL, reason string) (service.StudentIDCardIssueResult, error)
	PublicVerify(ctx context.Context, token, ip, userAgent string) (service.StudentIDCardVerifyResult, error)
	StartPortalLogin(ctx context.Context, token, ip, userAgent string) (service.StudentCardPortalChallenge, error)
	CompletePortalLogin(ctx context.Context, challenge, pin string) (service.StudentCardPortalCompleteResult, error)
	AttendanceScan(ctx context.Context, token, activityCode, scanType string, actorID pgtype.UUID, ip, userAgent string) (db.StudentActivityAttendanceScan, error)
	LibraryScan(ctx context.Context, token string, actorID pgtype.UUID, ip, userAgent string) (db.GetStudentIDCardByHashRow, error)
	CBTValidate(ctx context.Context, token string, actorID pgtype.UUID, ip, userAgent string) ([]db.ValidateCardForCBTRow, error)
	Events(ctx context.Context, cardID pgtype.UUID, limit, offset int32) ([]db.StudentIDCardEvent, error)
	AuditLogs(ctx context.Context, cardID pgtype.UUID, limit, offset int32) ([]db.StudentIDCardAuditLog, error)
}

func NewStudentIDCard(svc *service.StudentIDCard, auth studentIDCardPortalAuthService) *StudentIDCard {
	return &StudentIDCard{svc: svc, auth: auth}
}

func (h *StudentIDCard) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context(), queryInt32(r, "limit", 50), queryInt32(r, "offset", 0))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, scrubIDCardSecrets(map[string]any{"items": items}))
}

func (h *StudentIDCard) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	row, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu siswa tidak ditemukan")
		return
	}
	api.OK(w, scrubIDCardSecrets(row))
}

func (h *StudentIDCard) Generate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StudentID string `json:"student_id"`
		BaseURL   string `json:"base_url"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	studentID, ok := scanUUID(w, body.StudentID, "student_id")
	if !ok {
		return
	}
	actorID := actorUUID(r)
	res, err := h.svc.Issue(r.Context(), studentID, actorID, body.BaseURL)
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal membuat kartu siswa")
		return
	}
	api.Created(w, scrubIDCardSecrets(res))
}

func (h *StudentIDCard) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var body struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	card, err := h.svc.UpdateStatus(r.Context(), id, actorUUID(r), strings.TrimSpace(body.Status), body.Reason)
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal mengubah status kartu")
		return
	}
	api.OK(w, scrubIDCardSecrets(card))
}

func (h *StudentIDCard) MarkPrinted(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	card, err := h.svc.MarkPrinted(r.Context(), id, actorUUID(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal menandai kartu tercetak")
		return
	}
	api.OK(w, scrubIDCardSecrets(card))
}

func (h *StudentIDCard) Reissue(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var body struct {
		BaseURL string `json:"base_url"`
		Reason  string `json:"reason"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	res, err := h.svc.Reissue(r.Context(), id, actorUUID(r), body.BaseURL, body.Reason)
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal cetak ulang kartu")
		return
	}
	api.OK(w, scrubIDCardSecrets(res))
}

func (h *StudentIDCard) Events(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.svc.Events(r.Context(), id, queryInt32(r, "limit", 50), queryInt32(r, "offset", 0))
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal membaca riwayat kartu")
		return
	}
	api.OK(w, scrubIDCardSecrets(map[string]any{"items": items}))
}

func (h *StudentIDCard) AuditLogs(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	items, err := h.svc.AuditLogs(r.Context(), id, queryInt32(r, "limit", 50), queryInt32(r, "offset", 0))
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal membaca audit kartu")
		return
	}
	api.OK(w, scrubIDCardSecrets(map[string]any{"items": items}))
}

func (h *StudentIDCard) PublicVerify(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		token = r.URL.Query().Get("t")
	}
	res, err := h.svc.PublicVerify(r.Context(), token, r.RemoteAddr, r.UserAgent())
	if err != nil {
		api.OK(w, scrubIDCardSecrets(res))
		return
	}
	api.OK(w, scrubIDCardSecrets(res))
}

func (h *StudentIDCard) PortalLoginStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QRToken string `json:"qr_token"`
		Token   string `json:"token"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	tok := firstNonEmptyLocal(body.QRToken, body.Token)
	res, err := h.svc.StartPortalLogin(r.Context(), tok, r.RemoteAddr, r.UserAgent())
	if err != nil {
		writeDomainOrInternal(w, err, "QR kartu tidak valid")
		return
	}
	api.OK(w, scrubIDCardSecrets(res))
}

func (h *StudentIDCard) PortalLoginComplete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChallengeID string `json:"challenge_id"`
		PIN         string `json:"pin"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	res, err := h.svc.CompletePortalLogin(r.Context(), body.ChallengeID, body.PIN)
	if err != nil {
		writeDomainOrInternal(w, err, "PIN tidak valid")
		return
	}
	var studentID pgtype.UUID
	if err := studentID.Scan(res.StudentID); err != nil || !studentID.Valid {
		api.Internal(w, err)
		return
	}
	pair, err := h.auth.LoginStudentByID(r.Context(), studentID, service.SessionMeta{IPAddress: r.RemoteAddr, UserAgent: r.UserAgent(), DeviceLabel: "Portal Siswa QR"})
	if err != nil {
		writeDomainOrInternal(w, err, "Akun siswa belum siap untuk login portal")
		return
	}
	api.OK(w, map[string]any{"ok": true, "message": res.Message, "student_id": res.StudentID, "access_token": pair.AccessToken, "refresh_token": pair.RefreshToken, "must_change_password": pair.MustChangePassword})
}

func (h *StudentIDCard) AttendanceScan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QRToken      string `json:"qr_token"`
		ActivityCode string `json:"activity_code"`
		ScanType     string `json:"scan_type"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	res, err := h.svc.AttendanceScan(r.Context(), body.QRToken, body.ActivityCode, body.ScanType, actorUUID(r), r.RemoteAddr, r.UserAgent())
	if err != nil {
		writeDomainOrInternal(w, err, "Gagal scan presensi")
		return
	}
	api.Created(w, scrubIDCardSecrets(res))
}

func (h *StudentIDCard) LibraryScan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QRToken string `json:"qr_token"`
		Token   string `json:"token"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	res, err := h.svc.LibraryScan(r.Context(), firstNonEmptyLocal(body.QRToken, body.Token), actorUUID(r), r.RemoteAddr, r.UserAgent())
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu siswa tidak valid")
		return
	}
	api.OK(w, scrubIDCardSecrets(map[string]any{"member_type": "student", "member_id": pgUUIDString(res.StudentID), "student": res}))
}

func (h *StudentIDCard) CBTValidate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		QRToken string `json:"qr_token"`
		Token   string `json:"token"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit, disallowUnknownJSONFields) {
		return
	}
	items, err := h.svc.CBTValidate(r.Context(), firstNonEmptyLocal(body.QRToken, body.Token), actorUUID(r), r.RemoteAddr, r.UserAgent())
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu siswa tidak valid untuk CBT")
		return
	}
	api.OK(w, map[string]any{"items": items, "note": "Validasi kartu tidak membuka atau mengembalikan token ujian/ruang."})
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (pgtype.UUID, bool) {
	return scanUUID(w, chi.URLParam(r, key), key)
}
func scanUUID(w http.ResponseWriter, raw, key string) (pgtype.UUID, bool) {
	var id pgtype.UUID
	if err := id.Scan(strings.TrimSpace(raw)); err != nil || !id.Valid {
		api.BadRequest(w, key+" tidak valid")
		return id, false
	}
	return id, true
}
func actorUUID(r *http.Request) pgtype.UUID {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return pgtype.UUID{}
	}
	for _, k := range []string{"uid", "sub"} {
		if raw, ok := claims[k].(string); ok {
			var id pgtype.UUID
			if id.Scan(raw) == nil {
				return id
			}
		}
	}
	return pgtype.UUID{}
}
func queryInt32(r *http.Request, key string, def int32) int32 {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return int32(n)
}
func firstNonEmptyLocal(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func scrubIDCardSecrets(value any) any {
	payload, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var decoded any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return value
	}
	return scrubIDCardSecretValue(decoded)
}

func scrubIDCardSecretValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			normalized := strings.ToLower(key)
			switch normalized {
			case "token_hash", "pinhash", "pin_hash", "password_hash":
				delete(v, key)
			default:
				v[key] = scrubIDCardSecretValue(child)
			}
		}
		return v
	case []any:
		for i, child := range v {
			v[i] = scrubIDCardSecretValue(child)
		}
		return v
	default:
		return value
	}
}
