package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Letter struct {
	svc   letterService
	audit cbtAuthoringAuditWriter
}

type letterService interface {
	ListClassifications(ctx context.Context) ([]db.LetterClassification, error)
	ListIncoming(ctx context.Context, search, filterStatus string) ([]db.ListIncomingLettersRow, error)
	CreateIncoming(ctx context.Context, p service.CreateIncomingParams) (db.IncomingLetter, error)
	GetIncoming(ctx context.Context, id pgtype.UUID) (db.GetIncomingLetterRow, error)
	UpdateIncoming(ctx context.Context, p service.UpdateIncomingParams) (db.IncomingLetter, error)
	UpdateIncomingStatus(ctx context.Context, id pgtype.UUID, status string) (db.IncomingLetter, error)
	DeleteIncoming(ctx context.Context, id pgtype.UUID) error
	ListOutgoing(ctx context.Context, search string) ([]db.ListOutgoingLettersRow, error)
	CreateOutgoing(ctx context.Context, p service.CreateOutgoingParams) (db.OutgoingLetter, error)
	GetOutgoing(ctx context.Context, id pgtype.UUID) (db.GetOutgoingLetterRow, error)
	UpdateOutgoing(ctx context.Context, p service.UpdateOutgoingParams) (db.OutgoingLetter, error)
	DeleteOutgoing(ctx context.Context, id pgtype.UUID) error
	PreviewOutgoingNumber(classificationCode, tanggalSurat string) (string, error)
	ListDispositions(ctx context.Context, incomingLetterID, filterStatus string) ([]db.ListDispositionsRow, error)
	CreateDisposition(ctx context.Context, p service.CreateDispositionParams) (db.LetterDisposition, error)
	GetDisposition(ctx context.Context, id pgtype.UUID) (db.GetDispositionRow, error)
	UpdateDisposition(ctx context.Context, p service.UpdateDispositionParams) (db.LetterDisposition, error)
	DeleteDisposition(ctx context.Context, id pgtype.UUID) error
}

func NewLetter(svc *service.Letter, audit ...cbtAuthoringAuditWriter) *Letter {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &Letter{svc: svc, audit: writer}
}

func tuAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		return mw.HasAnyRole(claims, "admin", "staf")
	}
	return false
}

func tuEmployeeID(r *http.Request) pgtype.UUID {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if eid, _ := claims["eid"].(string); eid != "" {
			var id pgtype.UUID
			if err := id.Scan(strings.TrimSpace(eid)); err == nil {
				return id
			}
		}
	}
	return pgtype.UUID{}
}

// GET /api/tu/surat/klasifikasi
func (h *Letter) ListClassifications(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListClassifications(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// =====================
// Incoming Letters
// =====================

// GET /api/tu/surat/incoming
func (h *Letter) ListIncoming(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	search := r.URL.Query().Get("search")
	filterStatus := r.URL.Query().Get("status")
	data, err := h.svc.ListIncoming(r.Context(), search, filterStatus)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// POST /api/tu/surat/incoming
func (h *Letter) CreateIncoming(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		NomorSurat           string `json:"nomor_surat"`
		TanggalSurat         string `json:"tanggal_surat"`
		TanggalTerima        string `json:"tanggal_terima"`
		Asal                 string `json:"asal"`
		Perihal              string `json:"perihal"`
		Sifat                string `json:"sifat"`
		Catatan              string `json:"catatan"`
		ReceivedByEmployeeID string `json:"received_by_employee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	letter, err := h.svc.CreateIncoming(r.Context(), service.CreateIncomingParams{
		NomorSurat:           body.NomorSurat,
		TanggalSurat:         body.TanggalSurat,
		TanggalTerima:        body.TanggalTerima,
		Asal:                 body.Asal,
		Perihal:              body.Perihal,
		Sifat:                body.Sifat,
		Catatan:              body.Catatan,
		ReceivedByEmployeeID: body.ReceivedByEmployeeID,
	})
	if err != nil {
		writeClientError(w, err, "Data surat masuk tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_INCOMING_CREATE", "incoming_letter", pgUUIDString(letter.ID), map[string]any{
		"nomor_surat": letter.NomorSurat,
		"perihal":     letter.Perihal,
		"asal":        letter.Asal,
	})
	api.Created(w, letter)
}

// GET /api/tu/surat/incoming/{id}
func (h *Letter) GetIncoming(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	row, err := h.svc.GetIncoming(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

// PUT /api/tu/surat/incoming/{id}
func (h *Letter) UpdateIncoming(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		NomorSurat    string `json:"nomor_surat"`
		TanggalSurat  string `json:"tanggal_surat"`
		TanggalTerima string `json:"tanggal_terima"`
		Asal          string `json:"asal"`
		Perihal       string `json:"perihal"`
		Sifat         string `json:"sifat"`
		Catatan       string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpdateIncoming(r.Context(), service.UpdateIncomingParams{
		ID:            id,
		NomorSurat:    body.NomorSurat,
		TanggalSurat:  body.TanggalSurat,
		TanggalTerima: body.TanggalTerima,
		Asal:          body.Asal,
		Perihal:       body.Perihal,
		Sifat:         body.Sifat,
		Catatan:       body.Catatan,
	})
	if err != nil {
		writeClientError(w, err, "Perubahan surat masuk tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_INCOMING_UPDATE", "incoming_letter", pgUUIDString(row.ID), map[string]any{
		"nomor_surat": row.NomorSurat,
		"perihal":     row.Perihal,
		"asal":        row.Asal,
	})
	api.OK(w, row)
}

// PATCH /api/tu/surat/incoming/{id}/status
func (h *Letter) UpdateIncomingStatus(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpdateIncomingStatus(r.Context(), id, body.Status)
	if err != nil {
		writeClientError(w, err, "Status surat masuk tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_INCOMING_STATUS_UPDATE", "incoming_letter", pgUUIDString(row.ID), map[string]any{
		"status": row.Status,
	})
	api.OK(w, row)
}

// DELETE /api/tu/surat/incoming/{id}
func (h *Letter) DeleteIncoming(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteIncoming(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_INCOMING_DELETE", "incoming_letter", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}

// =====================
// Outgoing Letters
// =====================

// GET /api/tu/surat/outgoing
func (h *Letter) ListOutgoing(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListOutgoing(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// POST /api/tu/surat/outgoing
func (h *Letter) CreateOutgoing(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		ClassificationCode string `json:"classification_code"`
		TanggalSurat       string `json:"tanggal_surat"`
		Tujuan             string `json:"tujuan"`
		Perihal            string `json:"perihal"`
		Sifat              string `json:"sifat"`
		Catatan            string `json:"catatan"`
		ManualNomor        string `json:"manual_nomor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	eid := tuEmployeeID(r)
	eidStr := ""
	if eid.Valid {
		eidStr = eid.String()
	}
	letter, err := h.svc.CreateOutgoing(r.Context(), service.CreateOutgoingParams{
		ClassificationCode: body.ClassificationCode,
		TanggalSurat:       body.TanggalSurat,
		Tujuan:             body.Tujuan,
		Perihal:            body.Perihal,
		Sifat:              body.Sifat,
		Catatan:            body.Catatan,
		IssuedByEmployeeID: eidStr,
		ManualNomor:        body.ManualNomor,
	})
	if err != nil {
		writeClientError(w, err, "Data surat keluar tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_OUTGOING_CREATE", "outgoing_letter", pgUUIDString(letter.ID), map[string]any{
		"nomor_surat": letter.NomorSurat,
		"perihal":     letter.Perihal,
		"tujuan":      letter.Tujuan,
	})
	api.Created(w, letter)
}

// GET /api/tu/surat/outgoing/{id}
func (h *Letter) GetOutgoing(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	row, err := h.svc.GetOutgoing(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

// PUT /api/tu/surat/outgoing/{id}
func (h *Letter) UpdateOutgoing(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		TanggalSurat string `json:"tanggal_surat"`
		Tujuan       string `json:"tujuan"`
		Perihal      string `json:"perihal"`
		Sifat        string `json:"sifat"`
		Catatan      string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpdateOutgoing(r.Context(), service.UpdateOutgoingParams{
		ID: id, TanggalSurat: body.TanggalSurat, Tujuan: body.Tujuan,
		Perihal: body.Perihal, Sifat: body.Sifat, Catatan: body.Catatan,
	})
	if err != nil {
		writeClientError(w, err, "Perubahan surat keluar tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_OUTGOING_UPDATE", "outgoing_letter", pgUUIDString(row.ID), map[string]any{
		"nomor_surat": row.NomorSurat,
		"perihal":     row.Perihal,
		"tujuan":      row.Tujuan,
	})
	api.OK(w, row)
}

// DELETE /api/tu/surat/outgoing/{id}
func (h *Letter) DeleteOutgoing(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteOutgoing(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_OUTGOING_DELETE", "outgoing_letter", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}

// GET /api/tu/surat/outgoing/preview-number
func (h *Letter) PreviewOutgoingNumber(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	code := r.URL.Query().Get("classification_code")
	tanggal := r.URL.Query().Get("tanggal")
	preview, err := h.svc.PreviewOutgoingNumber(code, tanggal)
	if err != nil {
		writeClientError(w, err, "Preview nomor surat tidak valid")
		return
	}
	api.OK(w, map[string]string{"preview": preview})
}

// =====================
// Dispositions
// =====================

// GET /api/tu/surat/disposisi
func (h *Letter) ListDispositions(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	letterID := r.URL.Query().Get("incoming_letter_id")
	filterStatus := r.URL.Query().Get("status")
	data, err := h.svc.ListDispositions(r.Context(), letterID, filterStatus)
	if err != nil {
		writeClientError(w, err, "Filter disposisi tidak valid")
		return
	}
	api.OK(w, data)
}

// POST /api/tu/surat/disposisi
func (h *Letter) CreateDisposition(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		IncomingLetterID   string `json:"incoming_letter_id"`
		AssigneeEmployeeID string `json:"assignee_employee_id"`
		Instruksi          string `json:"instruksi"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	eid := tuEmployeeID(r)
	eidStr := ""
	if eid.Valid {
		eidStr = eid.String()
	}
	disp, err := h.svc.CreateDisposition(r.Context(), service.CreateDispositionParams{
		IncomingLetterID:     body.IncomingLetterID,
		AssigneeEmployeeID:   body.AssigneeEmployeeID,
		Instruksi:            body.Instruksi,
		DisposedByEmployeeID: eidStr,
	})
	if err != nil {
		writeClientError(w, err, "Data disposisi tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_DISPOSITION_CREATE", "letter_disposition", pgUUIDString(disp.ID), map[string]any{
		"incoming_letter_id":   pgUUIDString(disp.IncomingLetterID),
		"assignee_employee_id": pgUUIDString(disp.AssigneeEmployeeID),
	})
	api.Created(w, disp)
}

// GET /api/tu/surat/disposisi/{id}
func (h *Letter) GetDisposition(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	row, err := h.svc.GetDisposition(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

// PUT /api/tu/surat/disposisi/{id}
func (h *Letter) UpdateDisposition(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id := chi.URLParam(r, "id")
	var body struct {
		Instruksi           string `json:"instruksi"`
		CatatanTindakLanjut string `json:"catatan_tindak_lanjut"`
		Status              string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpdateDisposition(r.Context(), service.UpdateDispositionParams{
		ID:                  id,
		Instruksi:           body.Instruksi,
		CatatanTindakLanjut: body.CatatanTindakLanjut,
		Status:              body.Status,
	})
	if err != nil {
		writeClientError(w, err, "Perubahan disposisi tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_DISPOSITION_UPDATE", "letter_disposition", pgUUIDString(row.ID), map[string]any{
		"status": row.Status,
	})
	api.OK(w, row)
}

// DELETE /api/tu/surat/disposisi/{id}
func (h *Letter) DeleteDisposition(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteDisposition(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LETTER_DISPOSITION_DELETE", "letter_disposition", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}
