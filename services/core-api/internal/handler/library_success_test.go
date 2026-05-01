package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeLibraryService struct {
	*service.Library

	statsRow      db.GetLibraryStatsRow
	statsErr      error
	bookSearch    string
	bookKategori  string
	books         []db.LibraryBook
	listBooksErr  error
	createArg     db.CreateBookParams
	createRow     db.LibraryBook
	createErr     error
	updateArg     db.UpdateBookParams
	updateRow     db.LibraryBook
	updateErr     error
	deleteID      pgtype.UUID
	deleteErr     error
	loanStatus    string
	loans         []db.ListLoansRow
	listLoansErr  error
	loanBookID    string
	loanMemberTyp string
	loanMemberID  string
	loanDueDays   int
	loanRow       db.LibraryLoan
	loanErr       error
	returnLoanID  string
	returnRow     db.LibraryLoan
	returnErr     error
	dendaLoanID   string
	dendaRow      db.LibraryLoan
	dendaErr      error
}

func (f *fakeLibraryService) Stats(context.Context) (db.GetLibraryStatsRow, error) {
	return f.statsRow, f.statsErr
}

func (f *fakeLibraryService) ListBooks(_ context.Context, search, kategori string) ([]db.LibraryBook, error) {
	f.bookSearch = search
	f.bookKategori = kategori
	return f.books, f.listBooksErr
}

func (f *fakeLibraryService) CreateBook(_ context.Context, arg db.CreateBookParams) (db.LibraryBook, error) {
	f.createArg = arg
	return f.createRow, f.createErr
}

func (f *fakeLibraryService) UpdateBook(_ context.Context, arg db.UpdateBookParams) (db.LibraryBook, error) {
	f.updateArg = arg
	return f.updateRow, f.updateErr
}

func (f *fakeLibraryService) DeleteBook(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeLibraryService) ListLoans(_ context.Context, filterStatus string) ([]db.ListLoansRow, error) {
	f.loanStatus = filterStatus
	return f.loans, f.listLoansErr
}

func (f *fakeLibraryService) LoanBook(_ context.Context, bookID, memberType, memberID string, dueDays int) (db.LibraryLoan, error) {
	f.loanBookID = bookID
	f.loanMemberTyp = memberType
	f.loanMemberID = memberID
	f.loanDueDays = dueDays
	return f.loanRow, f.loanErr
}

func (f *fakeLibraryService) ReturnBook(_ context.Context, loanID string) (db.LibraryLoan, error) {
	f.returnLoanID = loanID
	return f.returnRow, f.returnErr
}

func (f *fakeLibraryService) MarkDendaLunas(_ context.Context, loanID string) (db.LibraryLoan, error) {
	f.dendaLoanID = loanID
	return f.dendaRow, f.dendaErr
}

func libraryTestBook(id pgtype.UUID, title string) db.LibraryBook {
	return db.LibraryBook{ID: id, Kode: "BK-001", Judul: title, Pengarang: "Penulis", Kategori: "fiksi", TotalEksemplar: 3, Tersedia: 2, LokasiRak: "A1"}
}

func libraryTestLoan(id, bookID, memberID pgtype.UUID) db.LibraryLoan {
	return db.LibraryLoan{ID: id, BookID: bookID, MemberType: db.LibraryMemberTypeStudent, StudentID: memberID, Status: db.LoanStatusEnumActive, DendaPerHari: 500}
}

func TestLibrarySuccessHandlersForwardPayloads(t *testing.T) {
	bookID := handlerTestUUID(126)
	loanID := handlerTestUUID(127)
	memberID := handlerTestUUID(128)
	fake := &fakeLibraryService{
		Library:   &service.Library{},
		statsRow:  db.GetLibraryStatsRow{TotalJudul: 3, SedangDipinjam: 1},
		books:     []db.LibraryBook{libraryTestBook(bookID, "Buku IPA")},
		createRow: libraryTestBook(bookID, "Buku Baru"),
		updateRow: libraryTestBook(bookID, "Buku Revisi"),
		loans:     []db.ListLoansRow{{ID: loanID, BookID: bookID, MemberNama: "Siswa", Status: db.LoanStatusEnumActive}},
		loanRow:   libraryTestLoan(loanID, bookID, memberID),
		returnRow: libraryTestLoan(loanID, bookID, memberID),
		dendaRow:  libraryTestLoan(loanID, bookID, memberID),
	}
	h := &Library{svc: fake}

	rec := httptest.NewRecorder()
	h.Stats(rec, adminRequest(http.MethodGet, "/api/library/stats", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("Stats status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListBooks(rec, adminRequest(http.MethodGet, "/api/library/books?search=ipa&kategori=fiksi", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListBooks status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.bookSearch != "ipa" || fake.bookKategori != "fiksi" {
		t.Fatalf("ListBooks filters = (%q, %q), want query filters", fake.bookSearch, fake.bookKategori)
	}

	rec = httptest.NewRecorder()
	h.CreateBook(rec, adminRequest(http.MethodPost, "/api/library/books", `{"kode":" BK-002 ","judul":" Buku Baru ","pengarang":" Guru ","isbn":" 123 ","kategori":" fiksi ","penerbit":" MTs ","tahun_terbit":2026,"total_eksemplar":4,"lokasi_rak":" B2 "}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateBook status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createArg.Kode != "BK-002" || fake.createArg.Judul != "Buku Baru" || fake.createArg.Pengarang != "Guru" || fake.createArg.Isbn != "123" || fake.createArg.Kategori != "fiksi" || fake.createArg.Penerbit != "MTs" || !fake.createArg.TahunTerbit.Valid || fake.createArg.TotalEksemplar != 4 || fake.createArg.LokasiRak != "B2" {
		t.Fatalf("CreateBook arg = %+v, want trimmed payload", fake.createArg)
	}

	rec = httptest.NewRecorder()
	h.UpdateBook(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/library/books/"+bookID.String(), `{"kode":"BK-002","judul":"Buku Revisi","pengarang":"Guru","kategori":"referensi","tahun_terbit":2025,"total_eksemplar":5,"lokasi_rak":"C3"}`), "id", bookID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateBook status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateArg.ID != bookID || fake.updateArg.Judul != "Buku Revisi" || fake.updateArg.Kategori != "referensi" || !fake.updateArg.TahunTerbit.Valid {
		t.Fatalf("UpdateBook arg = %+v, want route id and decoded payload", fake.updateArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteBook(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/library/books/"+bookID.String(), ""), "id", bookID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteBook status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != bookID {
		t.Fatalf("DeleteBook id = %v, want %v", fake.deleteID, bookID)
	}

	rec = httptest.NewRecorder()
	h.ListLoans(rec, adminRequest(http.MethodGet, "/api/library/loans?status=active", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListLoans status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.loanStatus != "active" {
		t.Fatalf("ListLoans status filter = %q, want active", fake.loanStatus)
	}

	rec = httptest.NewRecorder()
	h.LoanBook(rec, adminRequest(http.MethodPost, "/api/library/loans", `{"book_id":"`+bookID.String()+`","member_type":"student","member_id":"`+memberID.String()+`"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("LoanBook status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.loanBookID != bookID.String() || fake.loanMemberTyp != "student" || fake.loanMemberID != memberID.String() || fake.loanDueDays != 7 {
		t.Fatalf("LoanBook args = (%q, %q, %q, %d), want default 7-day loan", fake.loanBookID, fake.loanMemberTyp, fake.loanMemberID, fake.loanDueDays)
	}

	rec = httptest.NewRecorder()
	h.ReturnBook(rec, withRouteParam(adminRequest(http.MethodPost, "/api/library/loans/"+loanID.String()+"/return", ""), "id", loanID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ReturnBook status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.returnLoanID != loanID.String() {
		t.Fatalf("ReturnBook id = %q, want %q", fake.returnLoanID, loanID.String())
	}

	rec = httptest.NewRecorder()
	h.MarkDendaLunas(rec, withRouteParam(adminRequest(http.MethodPost, "/api/library/loans/"+loanID.String()+"/denda-lunas", ""), "id", loanID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("MarkDendaLunas status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.dendaLoanID != loanID.String() {
		t.Fatalf("MarkDendaLunas id = %q, want %q", fake.dendaLoanID, loanID.String())
	}
}

func TestLibraryMutationHandlersWriteAuditEvents(t *testing.T) {
	bookID := handlerTestUUID(140)
	loanID := handlerTestUUID(141)
	memberID := handlerTestUUID(142)
	fake := &fakeLibraryService{
		Library:   &service.Library{},
		createRow: libraryTestBook(bookID, "Buku Baru"),
		updateRow: libraryTestBook(bookID, "Buku Revisi"),
		loanRow:   libraryTestLoan(loanID, bookID, memberID),
		returnRow: libraryTestLoan(loanID, bookID, memberID),
		dendaRow:  libraryTestLoan(loanID, bookID, memberID),
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &Library{svc: fake, audit: audit}
	auditedStaffRequest := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"staf"},
			"uid":   "01000000-0000-0000-0000-000000000000",
			"sub":   "01000000-0000-0000-0000-000000000000",
			"usr":   "staf.perpus",
			"ssid":  "sess-library-1",
		})
	}

	rec := httptest.NewRecorder()
	h.CreateBook(rec, auditedStaffRequest(http.MethodPost, "/api/library/books", `{"kode":"BK-002","judul":"Buku Baru","pengarang":"Guru","isbn":"123","kategori":"fiksi","penerbit":"MTs","tahun_terbit":2026,"total_eksemplar":4,"lokasi_rak":"B2"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateBook status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateBook(rec, withRouteParam(auditedStaffRequest(http.MethodPatch, "/api/library/books/"+bookID.String(), `{"kode":"BK-002","judul":"Buku Revisi","pengarang":"Guru","kategori":"referensi","tahun_terbit":2025,"total_eksemplar":5,"lokasi_rak":"C3"}`), "id", bookID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateBook status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.DeleteBook(rec, withRouteParam(auditedStaffRequest(http.MethodDelete, "/api/library/books/"+bookID.String(), ""), "id", bookID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteBook status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.LoanBook(rec, auditedStaffRequest(http.MethodPost, "/api/library/loans", `{"book_id":"`+bookID.String()+`","member_type":"student","member_id":"`+memberID.String()+`"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("LoanBook status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ReturnBook(rec, withRouteParam(auditedStaffRequest(http.MethodPost, "/api/library/loans/"+loanID.String()+"/return", ""), "id", loanID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ReturnBook status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.MarkDendaLunas(rec, withRouteParam(auditedStaffRequest(http.MethodPost, "/api/library/loans/"+loanID.String()+"/denda-lunas", ""), "id", loanID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("MarkDendaLunas status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	if len(audit.entries) != 6 {
		t.Fatalf("audit entries = %d, want 6", len(audit.entries))
	}

	checks := []struct {
		index      int
		action     string
		entityType string
		entityID   string
	}{
		{0, "LIBRARY_BOOK_CREATE", "library_book", bookID.String()},
		{1, "LIBRARY_BOOK_UPDATE", "library_book", bookID.String()},
		{2, "LIBRARY_BOOK_DELETE", "library_book", bookID.String()},
		{3, "LIBRARY_LOAN_CREATE", "library_loan", loanID.String()},
		{4, "LIBRARY_LOAN_RETURN", "library_loan", loanID.String()},
		{5, "LIBRARY_LOAN_DENDA_LUNAS", "library_loan", loanID.String()},
	}
	for _, check := range checks {
		got := audit.entries[check.index]
		if got.Action != check.action || got.EntityType != check.entityType || got.EntityID != check.entityID {
			t.Fatalf("audit[%d] = %+v, want action/type/id %q/%q/%q", check.index, got, check.action, check.entityType, check.entityID)
		}
	}

	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["username"] != "staf.perpus" || meta["judul"] != "Buku Baru" {
		t.Fatalf("create book metadata = %+v, want username/judul", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[2].Metadata)
	if meta["deleted_by"] != "staf.perpus" {
		t.Fatalf("delete book metadata = %+v, want deleted_by", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[3].Metadata)
	if meta["member_type"] != string(db.LibraryMemberTypeStudent) || meta["member_id"] != memberID.String() {
		t.Fatalf("loan create metadata = %+v, want member_type/member_id", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[4].Metadata)
	if meta["returned_by"] != "staf.perpus" {
		t.Fatalf("return loan metadata = %+v, want returned_by", meta)
	}
}

func TestLibraryHandlersMapServiceErrors(t *testing.T) {
	bookID := handlerTestUUID(129)
	loanID := handlerTestUUID(130)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*Library, http.ResponseWriter, *http.Request)
		svc        *fakeLibraryService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "stats",
			handler:    (*Library).Stats,
			svc:        &fakeLibraryService{Library: &service.Library{}, statsErr: errDB},
			req:        adminRequest(http.MethodGet, "/api/library/stats", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "list books",
			handler:    (*Library).ListBooks,
			svc:        &fakeLibraryService{Library: &service.Library{}, listBooksErr: errDB},
			req:        adminRequest(http.MethodGet, "/api/library/books", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create book",
			handler:    (*Library).CreateBook,
			svc:        &fakeLibraryService{Library: &service.Library{}, createErr: errors.New("kode buku wajib diisi")},
			req:        adminRequest(http.MethodPost, "/api/library/books", `{}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update book",
			handler:    (*Library).UpdateBook,
			svc:        &fakeLibraryService{Library: &service.Library{}, updateErr: errors.New("judul buku wajib diisi")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/library/books/"+bookID.String(), `{}`), "id", bookID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete book",
			handler:    (*Library).DeleteBook,
			svc:        &fakeLibraryService{Library: &service.Library{}, deleteErr: errors.New("buku masih dipinjam")},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/library/books/"+bookID.String(), ""), "id", bookID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "list loans",
			handler:    (*Library).ListLoans,
			svc:        &fakeLibraryService{Library: &service.Library{}, listLoansErr: errDB},
			req:        adminRequest(http.MethodGet, "/api/library/loans", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "loan book",
			handler:    (*Library).LoanBook,
			svc:        &fakeLibraryService{Library: &service.Library{}, loanErr: errors.New("stok buku habis")},
			req:        adminRequest(http.MethodPost, "/api/library/loans", `{}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "return book",
			handler:    (*Library).ReturnBook,
			svc:        &fakeLibraryService{Library: &service.Library{}, returnErr: errors.New("data pinjaman tidak ditemukan")},
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/library/loans/"+loanID.String()+"/return", ""), "id", loanID.String()),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "denda lunas",
			handler:    (*Library).MarkDendaLunas,
			svc:        &fakeLibraryService{Library: &service.Library{}, dendaErr: errors.New("loan_id tidak valid")},
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/library/loans/"+loanID.String()+"/denda-lunas", ""), "id", loanID.String()),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Library{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == http.StatusBadRequest && strings.Contains(rec.Body.String(), "internal server error") {
				t.Fatalf("body = %s, want client-safe validation message", rec.Body.String())
			}
		})
	}
}
