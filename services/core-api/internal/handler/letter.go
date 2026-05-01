package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Letter struct {
	svc *service.Letter
}

func NewLetter(svc *service.Letter) *Letter { return &Letter{svc: svc} }

func tuAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if rawRoles, ok := claims["roles"].([]any); ok {
			for _, role := range rawRoles {
				if role == "admin" || role == "staf" {
					return true
				}
			}
		}
		if role, _ := claims["role"].(string); role == "admin" || role == "staf" {
			return true
		}
		return false
	}
	return true
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
		api.BadRequest(w, "invalid json")
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
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, "invalid json")
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
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateIncomingStatus(r.Context(), id, body.Status)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, "invalid json")
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
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uq_outgoing") {
			api.Conflict(w, "nomor surat sudah digunakan")
			return
		}
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateOutgoing(r.Context(), service.UpdateOutgoingParams{
		ID: id, TanggalSurat: body.TanggalSurat, Tujuan: body.Tujuan,
		Perihal: body.Perihal, Sifat: body.Sifat, Catatan: body.Catatan,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, err.Error())
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
		api.BadRequest(w, err.Error())
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
		api.BadRequest(w, "invalid json")
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
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateDisposition(r.Context(), service.UpdateDispositionParams{
		ID:                  id,
		Instruksi:           body.Instruksi,
		CatatanTindakLanjut: body.CatatanTindakLanjut,
		Status:              body.Status,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
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
	api.NoContent(w)
}
