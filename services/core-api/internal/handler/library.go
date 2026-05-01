package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Library struct{ svc *service.Library }

func NewLibrary(svc *service.Library) *Library { return &Library{svc: svc} }

func (h *Library) Stats(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// ---- Books ----

func (h *Library) ListBooks(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	search := r.URL.Query().Get("search")
	kategori := r.URL.Query().Get("kategori")
	books, err := h.svc.ListBooks(r.Context(), search, kategori)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, books)
}

func (h *Library) CreateBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Kode           string `json:"kode"`
		Judul          string `json:"judul"`
		Pengarang      string `json:"pengarang"`
		Isbn           string `json:"isbn"`
		Kategori       string `json:"kategori"`
		Penerbit       string `json:"penerbit"`
		TahunTerbit    int    `json:"tahun_terbit"`
		TotalEksemplar int32  `json:"total_eksemplar"`
		LokasiRak      string `json:"lokasi_rak"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	tahun := pgtype.Int4{}
	if body.TahunTerbit > 0 {
		_ = tahun.Scan(int32(body.TahunTerbit))
	}
	book, err := h.svc.CreateBook(r.Context(), db.CreateBookParams{
		Kode:           strings.TrimSpace(body.Kode),
		Judul:          strings.TrimSpace(body.Judul),
		Pengarang:      strings.TrimSpace(body.Pengarang),
		Isbn:           strings.TrimSpace(body.Isbn),
		Kategori:       strings.TrimSpace(body.Kategori),
		Penerbit:       strings.TrimSpace(body.Penerbit),
		TahunTerbit:    tahun,
		TotalEksemplar: body.TotalEksemplar,
		LokasiRak:      strings.TrimSpace(body.LokasiRak),
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, book)
}

func (h *Library) UpdateBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Kode           string `json:"kode"`
		Judul          string `json:"judul"`
		Pengarang      string `json:"pengarang"`
		Isbn           string `json:"isbn"`
		Kategori       string `json:"kategori"`
		Penerbit       string `json:"penerbit"`
		TahunTerbit    int    `json:"tahun_terbit"`
		TotalEksemplar int32  `json:"total_eksemplar"`
		LokasiRak      string `json:"lokasi_rak"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	tahun := pgtype.Int4{}
	if body.TahunTerbit > 0 {
		_ = tahun.Scan(int32(body.TahunTerbit))
	}
	book, err := h.svc.UpdateBook(r.Context(), db.UpdateBookParams{
		ID:             id,
		Kode:           strings.TrimSpace(body.Kode),
		Judul:          strings.TrimSpace(body.Judul),
		Pengarang:      strings.TrimSpace(body.Pengarang),
		Isbn:           strings.TrimSpace(body.Isbn),
		Kategori:       strings.TrimSpace(body.Kategori),
		Penerbit:       strings.TrimSpace(body.Penerbit),
		TahunTerbit:    tahun,
		TotalEksemplar: body.TotalEksemplar,
		LokasiRak:      strings.TrimSpace(body.LokasiRak),
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, book)
}

func (h *Library) DeleteBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.DeleteBook(r.Context(), id); err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.NoContent(w)
}

// ---- Loans ----

func (h *Library) ListLoans(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	status := r.URL.Query().Get("status")
	loans, err := h.svc.ListLoans(r.Context(), status)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, loans)
}

func (h *Library) LoanBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		BookID     string `json:"book_id"`
		MemberType string `json:"member_type"`
		MemberID   string `json:"member_id"`
		DueDays    int    `json:"due_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.DueDays == 0 {
		body.DueDays = 7
	}
	loan, err := h.svc.LoanBook(r.Context(), body.BookID, body.MemberType, body.MemberID, body.DueDays)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, loan)
}

func (h *Library) ReturnBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	loanID := chi.URLParam(r, "id")
	loan, err := h.svc.ReturnBook(r.Context(), loanID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, loan)
}

func (h *Library) MarkDendaLunas(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	loanID := chi.URLParam(r, "id")
	loan, err := h.svc.MarkDendaLunas(r.Context(), loanID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, loan)
}

// libraryAccessAllowed checks admin or staf role.
func libraryAccessAllowed(r *http.Request) bool {
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
	return false
}
