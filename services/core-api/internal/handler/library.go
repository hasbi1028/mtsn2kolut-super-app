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

type Library struct {
	svc   libraryService
	audit cbtAuthoringAuditWriter
}

type libraryService interface {
	Stats(ctx context.Context) (db.GetLibraryStatsRow, error)
	ListBooks(ctx context.Context, search, kategori string) ([]db.LibraryBook, error)
	CreateBook(ctx context.Context, arg db.CreateBookParams) (db.LibraryBook, error)
	UpdateBook(ctx context.Context, arg db.UpdateBookParams) (db.LibraryBook, error)
	DeleteBook(ctx context.Context, id pgtype.UUID) error
	ListLoans(ctx context.Context, filterStatus string) ([]db.ListLoansRow, error)
	LoanBook(ctx context.Context, bookID, memberType, memberID string, dueDays int) (db.LibraryLoan, error)
	ReturnBook(ctx context.Context, loanID string) (db.LibraryLoan, error)
	MarkDendaLunas(ctx context.Context, loanID string) (db.LibraryLoan, error)
}

func NewLibrary(svc *service.Library, audit ...cbtAuthoringAuditWriter) *Library {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &Library{svc: svc, audit: writer}
}

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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	tahun := pgtype.Int4{}
	if body.TahunTerbit > 0 {
		tahun = pgtype.Int4{Int32: int32(body.TahunTerbit), Valid: true}
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
		writeClientError(w, err, "Data buku tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LIBRARY_BOOK_CREATE", "library_book", pgUUIDString(book.ID), map[string]any{
		"kode":     book.Kode,
		"judul":    book.Judul,
		"kategori": book.Kategori,
	})
	api.Created(w, book)
}

func (h *Library) UpdateBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	tahun := pgtype.Int4{}
	if body.TahunTerbit > 0 {
		tahun = pgtype.Int4{Int32: int32(body.TahunTerbit), Valid: true}
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
		writeClientError(w, err, "Perubahan buku tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LIBRARY_BOOK_UPDATE", "library_book", pgUUIDString(book.ID), map[string]any{
		"kode":     book.Kode,
		"judul":    book.Judul,
		"kategori": book.Kategori,
	})
	api.OK(w, book)
}

func (h *Library) DeleteBook(w http.ResponseWriter, r *http.Request) {
	if !libraryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.DeleteBook(r.Context(), id); err != nil {
		writeClientError(w, err, "Penghapusan buku tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LIBRARY_BOOK_DELETE", "library_book", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if body.DueDays == 0 {
		body.DueDays = 7
	}
	loan, err := h.svc.LoanBook(r.Context(), body.BookID, body.MemberType, body.MemberID, body.DueDays)
	if err != nil {
		writeClientError(w, err, "Peminjaman buku tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LIBRARY_LOAN_CREATE", "library_loan", pgUUIDString(loan.ID), map[string]any{
		"book_id":     pgUUIDString(loan.BookID),
		"member_type": loan.MemberType,
		"member_id":   strings.TrimSpace(body.MemberID),
		"due_days":    body.DueDays,
	})
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
		writeClientError(w, err, "Pengembalian buku tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LIBRARY_LOAN_RETURN", "library_loan", pgUUIDString(loan.ID), map[string]any{
		"book_id":     pgUUIDString(loan.BookID),
		"returned_by": currentUsername(r),
	})
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
		writeClientError(w, err, "Pelunasan denda tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "LIBRARY_LOAN_DENDA_LUNAS", "library_loan", pgUUIDString(loan.ID), map[string]any{
		"book_id":    pgUUIDString(loan.BookID),
		"updated_by": currentUsername(r),
	})
	api.OK(w, loan)
}

// libraryAccessAllowed checks admin or staf role.
func libraryAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		return mw.HasAnyRole(claims, "admin", "staf")
	}
	return false
}
