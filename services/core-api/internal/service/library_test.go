package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeLibraryStore struct {
	listBooksArg    db.ListBooksParams
	listBooksRows   []db.LibraryBook
	listBooksErr    error
	book            db.LibraryBook
	bookErr         error
	getBookID       pgtype.UUID
	createBookArg   db.CreateBookParams
	createBookRow   db.LibraryBook
	createBookErr   error
	updateBookArg   db.UpdateBookParams
	updateBookRow   db.LibraryBook
	updateBookErr   error
	deleteBookID    pgtype.UUID
	deleteBookErr   error
	decrementID     pgtype.UUID
	decrementErr    error
	incrementID     pgtype.UUID
	incrementErr    error
	listLoansStatus string
	listLoansRows   []db.ListLoansRow
	listLoansErr    error
	getLoanID       pgtype.UUID
	getLoanRow      db.GetLoanRow
	getLoanErr      error
	createLoanArg   db.CreateLoanParams
	createLoanRow   db.LibraryLoan
	createLoanErr   error
	updateReturnArg db.UpdateLoanReturnParams
	updateReturnErr error
	markDendaID     pgtype.UUID
	markDendaRow    db.LibraryLoan
	markDendaErr    error
	activeMemberID  pgtype.UUID
	activeLoanCount int64
	activeLoanErr   error
	stats           db.GetLibraryStatsRow
	statsErr        error
}

func (f *fakeLibraryStore) ListBooks(ctx context.Context, arg db.ListBooksParams) ([]db.LibraryBook, error) {
	f.listBooksArg = arg
	return f.listBooksRows, f.listBooksErr
}

func (f *fakeLibraryStore) GetBook(ctx context.Context, id pgtype.UUID) (db.LibraryBook, error) {
	f.getBookID = id
	return f.book, f.bookErr
}

func (f *fakeLibraryStore) CreateBook(ctx context.Context, arg db.CreateBookParams) (db.LibraryBook, error) {
	f.createBookArg = arg
	if f.createBookErr != nil {
		return db.LibraryBook{}, f.createBookErr
	}
	if f.createBookRow.ID.Valid {
		return f.createBookRow, nil
	}
	return db.LibraryBook{Kode: arg.Kode, Judul: arg.Judul, TotalEksemplar: arg.TotalEksemplar, Tersedia: arg.TotalEksemplar}, nil
}

func (f *fakeLibraryStore) UpdateBook(ctx context.Context, arg db.UpdateBookParams) (db.LibraryBook, error) {
	f.updateBookArg = arg
	if f.updateBookErr != nil {
		return db.LibraryBook{}, f.updateBookErr
	}
	if f.updateBookRow.ID.Valid {
		return f.updateBookRow, nil
	}
	return db.LibraryBook{ID: arg.ID, Kode: arg.Kode, Judul: arg.Judul, TotalEksemplar: arg.TotalEksemplar}, nil
}

func (f *fakeLibraryStore) DeleteBook(ctx context.Context, id pgtype.UUID) error {
	f.deleteBookID = id
	return f.deleteBookErr
}

func (f *fakeLibraryStore) DecrementTersedia(ctx context.Context, id pgtype.UUID) error {
	f.decrementID = id
	return f.decrementErr
}

func (f *fakeLibraryStore) IncrementTersedia(ctx context.Context, id pgtype.UUID) error {
	f.incrementID = id
	return f.incrementErr
}

func (f *fakeLibraryStore) ListLoans(ctx context.Context, filterStatus string) ([]db.ListLoansRow, error) {
	f.listLoansStatus = filterStatus
	return f.listLoansRows, f.listLoansErr
}

func (f *fakeLibraryStore) GetLoan(ctx context.Context, id pgtype.UUID) (db.GetLoanRow, error) {
	f.getLoanID = id
	return f.getLoanRow, f.getLoanErr
}

func (f *fakeLibraryStore) CreateLoan(ctx context.Context, arg db.CreateLoanParams) (db.LibraryLoan, error) {
	f.createLoanArg = arg
	if f.createLoanErr != nil {
		return db.LibraryLoan{}, f.createLoanErr
	}
	if f.createLoanRow.ID.Valid {
		return f.createLoanRow, nil
	}
	return db.LibraryLoan{ID: libraryTestUUID(30), BookID: arg.BookID, MemberType: arg.MemberType, StudentID: arg.StudentID, EmployeeID: arg.EmployeeID}, nil
}

func (f *fakeLibraryStore) UpdateLoanReturn(ctx context.Context, arg db.UpdateLoanReturnParams) (db.LibraryLoan, error) {
	f.updateReturnArg = arg
	if f.updateReturnErr != nil {
		return db.LibraryLoan{}, f.updateReturnErr
	}
	return db.LibraryLoan{ID: arg.ID, DendaTotal: arg.DendaTotal, Status: db.LoanStatusEnumReturned}, nil
}

func (f *fakeLibraryStore) MarkLoanDendaLunas(ctx context.Context, id pgtype.UUID) (db.LibraryLoan, error) {
	f.markDendaID = id
	return f.markDendaRow, f.markDendaErr
}

func (f *fakeLibraryStore) CountActiveLoansForMember(ctx context.Context, memberID pgtype.UUID) (int64, error) {
	f.activeMemberID = memberID
	return f.activeLoanCount, f.activeLoanErr
}

func (f *fakeLibraryStore) GetLibraryStats(ctx context.Context) (db.GetLibraryStatsRow, error) {
	return f.stats, f.statsErr
}

func libraryTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func mustLibraryUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("Scan(%q) error = %v", value, err)
	}
	return id
}

func TestLibraryListBooksTrimsFilters(t *testing.T) {
	store := &fakeLibraryStore{}
	svc := &Library{q: store}

	if _, err := svc.ListBooks(context.Background(), "  matematika  ", "  referensi  "); err != nil {
		t.Fatalf("ListBooks() error = %v", err)
	}
	if store.listBooksArg.Search != "matematika" || store.listBooksArg.Kategori != "referensi" {
		t.Fatalf("ListBooks() arg = %+v, want trimmed filters", store.listBooksArg)
	}
}

func TestLibraryGetBookForwardsID(t *testing.T) {
	bookID := libraryTestUUID(40)
	store := &fakeLibraryStore{book: db.LibraryBook{ID: bookID, Kode: "BK-1", Judul: "Buku"}}
	svc := &Library{q: store}

	book, err := svc.GetBook(context.Background(), bookID)
	if err != nil {
		t.Fatalf("GetBook() error = %v", err)
	}
	if book.ID != bookID || store.getBookID != bookID {
		t.Fatalf("GetBook() = %+v, store id=%v; want %v", book, store.getBookID, bookID)
	}
}

func TestLibraryCreateBookValidatesAndNormalizes(t *testing.T) {
	tests := []struct {
		name    string
		arg     db.CreateBookParams
		wantErr string
	}{
		{name: "empty kode", arg: db.CreateBookParams{Judul: "Buku", TotalEksemplar: 1}, wantErr: "kode buku wajib diisi"},
		{name: "empty judul", arg: db.CreateBookParams{Kode: "BK-1", TotalEksemplar: 1}, wantErr: "judul buku wajib diisi"},
		{name: "invalid total", arg: db.CreateBookParams{Kode: "BK-1", Judul: "Buku"}, wantErr: "jumlah eksemplar minimal 1"},
		{name: "valid", arg: db.CreateBookParams{Kode: " BK-1 ", Judul: " Buku Guru ", TotalEksemplar: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLibraryStore{}
			svc := &Library{q: store}

			_, err := svc.CreateBook(context.Background(), tt.arg)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("CreateBook() error = %v, want %q", err, tt.wantErr)
				}
				if store.createBookArg.Kode != "" {
					t.Fatalf("CreateBook store was called after validation failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateBook() error = %v", err)
			}
			if store.createBookArg.Kode != "BK-1" || store.createBookArg.Judul != "Buku Guru" {
				t.Fatalf("CreateBook() arg = %+v, want trimmed kode/judul", store.createBookArg)
			}
		})
	}
}

func TestLibraryUpdateBookValidatesAndNormalizes(t *testing.T) {
	bookID := libraryTestUUID(1)
	store := &fakeLibraryStore{}
	svc := &Library{q: store}

	_, err := svc.UpdateBook(context.Background(), db.UpdateBookParams{
		ID:             bookID,
		Kode:           " BK-2 ",
		Judul:          " Buku Siswa ",
		TotalEksemplar: 3,
	})
	if err != nil {
		t.Fatalf("UpdateBook() error = %v", err)
	}
	if store.updateBookArg.ID != bookID || store.updateBookArg.Kode != "BK-2" || store.updateBookArg.Judul != "Buku Siswa" {
		t.Fatalf("UpdateBook() arg = %+v, want id and trimmed kode/judul", store.updateBookArg)
	}

	_, err = svc.UpdateBook(context.Background(), db.UpdateBookParams{ID: bookID, Kode: "BK-2", Judul: "Buku", TotalEksemplar: 0})
	if err == nil || err.Error() != "jumlah eksemplar minimal 1" {
		t.Fatalf("UpdateBook(invalid total) error = %v, want validation error", err)
	}

	_, err = svc.UpdateBook(context.Background(), db.UpdateBookParams{ID: bookID, Judul: "Buku", TotalEksemplar: 1})
	if err == nil || err.Error() != "kode buku wajib diisi" {
		t.Fatalf("UpdateBook(empty code) error = %v, want validation error", err)
	}
	_, err = svc.UpdateBook(context.Background(), db.UpdateBookParams{ID: bookID, Kode: "BK-2", TotalEksemplar: 1})
	if err == nil || err.Error() != "judul buku wajib diisi" {
		t.Fatalf("UpdateBook(empty title) error = %v, want validation error", err)
	}

	expectedErr := errors.New("update failed")
	svc = &Library{q: &fakeLibraryStore{updateBookErr: expectedErr}}
	_, err = svc.UpdateBook(context.Background(), db.UpdateBookParams{ID: bookID, Kode: "BK-2", Judul: "Buku", TotalEksemplar: 1})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateBook(store error) error = %v, want %v", err, expectedErr)
	}
}

func TestLibraryDeleteBookRequiresAllCopiesAvailable(t *testing.T) {
	bookID := libraryTestUUID(2)
	store := &fakeLibraryStore{
		book: db.LibraryBook{ID: bookID, TotalEksemplar: 4, Tersedia: 3},
	}
	svc := &Library{q: store}

	err := svc.DeleteBook(context.Background(), bookID)
	if err == nil || err.Error() != "buku masih dipinjam, tidak bisa dihapus" {
		t.Fatalf("DeleteBook() error = %v, want active-loan guard", err)
	}
	if store.deleteBookID.Valid {
		t.Fatalf("DeleteBook store was called while copies were borrowed")
	}

	store.book = db.LibraryBook{ID: bookID, TotalEksemplar: 4, Tersedia: 4}
	if err := svc.DeleteBook(context.Background(), bookID); err != nil {
		t.Fatalf("DeleteBook(available) error = %v", err)
	}
	if store.deleteBookID != bookID {
		t.Fatalf("DeleteBook() id = %v, want %v", store.deleteBookID, bookID)
	}
}

func TestLibraryDeleteBookPropagatesStoreErrors(t *testing.T) {
	bookID := libraryTestUUID(41)
	expectedErr := errors.New("store failed")

	svc := &Library{q: &fakeLibraryStore{bookErr: expectedErr}}
	if err := svc.DeleteBook(context.Background(), bookID); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteBook(get error) error = %v, want %v", err, expectedErr)
	}

	svc = &Library{q: &fakeLibraryStore{
		book:          db.LibraryBook{ID: bookID, TotalEksemplar: 2, Tersedia: 2},
		deleteBookErr: expectedErr,
	}}
	if err := svc.DeleteBook(context.Background(), bookID); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteBook(delete error) error = %v, want %v", err, expectedErr)
	}
}

func TestLibraryLoanBookValidatesInput(t *testing.T) {
	tests := []struct {
		name       string
		bookID     string
		memberType string
		memberID   string
		dueDays    int
		wantErr    string
	}{
		{name: "invalid book id", bookID: "bad", memberType: "student", memberID: "00000000-0000-0000-0000-000000000002", dueDays: 7, wantErr: "book_id tidak valid"},
		{name: "invalid member id", bookID: "00000000-0000-0000-0000-000000000001", memberType: "student", memberID: "bad", dueDays: 7, wantErr: "member_id tidak valid"},
		{name: "invalid member type", bookID: "00000000-0000-0000-0000-000000000001", memberType: "parent", memberID: "00000000-0000-0000-0000-000000000002", dueDays: 7, wantErr: "member_type harus 'student' atau 'employee'"},
		{name: "invalid due days", bookID: "00000000-0000-0000-0000-000000000001", memberType: "student", memberID: "00000000-0000-0000-0000-000000000002", dueDays: 31, wantErr: "durasi pinjam harus antara 1-30 hari"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLibraryStore{}
			svc := &Library{q: store}

			_, err := svc.LoanBook(context.Background(), tt.bookID, tt.memberType, tt.memberID, tt.dueDays)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("LoanBook() error = %v, want %q", err, tt.wantErr)
			}
			if store.getBookID.Valid {
				t.Fatalf("GetBook() was called after invalid input")
			}
		})
	}
}

func TestLibraryLoanBookGuardsAvailabilityAndLoanLimit(t *testing.T) {
	bookID := "00000000-0000-0000-0000-000000000001"
	memberID := "00000000-0000-0000-0000-000000000002"

	t.Run("no stock", func(t *testing.T) {
		store := &fakeLibraryStore{
			book: db.LibraryBook{ID: mustLibraryUUID(t, bookID), TotalEksemplar: 2, Tersedia: 0},
		}
		svc := &Library{q: store}

		_, err := svc.LoanBook(context.Background(), bookID, "student", memberID, 7)
		if err == nil || err.Error() != "stok buku habis, tidak ada eksemplar yang tersedia" {
			t.Fatalf("LoanBook() error = %v, want no-stock guard", err)
		}
		if store.activeMemberID.Valid {
			t.Fatalf("CountActiveLoansForMember() was called after no-stock guard")
		}
	})

	t.Run("active limit", func(t *testing.T) {
		store := &fakeLibraryStore{
			book:            db.LibraryBook{ID: mustLibraryUUID(t, bookID), TotalEksemplar: 2, Tersedia: 1},
			activeLoanCount: 3,
		}
		svc := &Library{q: store}

		_, err := svc.LoanBook(context.Background(), bookID, "student", memberID, 7)
		if err == nil || !strings.Contains(err.Error(), "3 pinjaman aktif") {
			t.Fatalf("LoanBook() error = %v, want active-loan limit", err)
		}
		if store.createLoanArg.BookID.Valid {
			t.Fatalf("CreateLoan() was called after active-loan limit")
		}
	})
}

func TestLibraryLoanBookPropagatesStoreErrors(t *testing.T) {
	bookID := "00000000-0000-0000-0000-000000000001"
	memberID := "00000000-0000-0000-0000-000000000002"
	parsedBookID := mustLibraryUUID(t, bookID)
	expectedErr := errors.New("store failed")

	svc := &Library{q: &fakeLibraryStore{bookErr: expectedErr}}
	_, err := svc.LoanBook(context.Background(), bookID, "student", memberID, 7)
	if err == nil || err.Error() != "buku tidak ditemukan" {
		t.Fatalf("LoanBook(get book error) error = %v, want public not-found message", err)
	}

	svc = &Library{q: &fakeLibraryStore{
		book:          db.LibraryBook{ID: parsedBookID, TotalEksemplar: 2, Tersedia: 1},
		activeLoanErr: expectedErr,
	}}
	_, err = svc.LoanBook(context.Background(), bookID, "student", memberID, 7)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("LoanBook(active loan error) error = %v, want %v", err, expectedErr)
	}

	svc = &Library{q: &fakeLibraryStore{
		book:          db.LibraryBook{ID: parsedBookID, TotalEksemplar: 2, Tersedia: 1},
		createLoanErr: expectedErr,
	}}
	_, err = svc.LoanBook(context.Background(), bookID, "student", memberID, 7)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("LoanBook(create loan error) error = %v, want %v", err, expectedErr)
	}

	svc = &Library{q: &fakeLibraryStore{
		book:         db.LibraryBook{ID: parsedBookID, TotalEksemplar: 2, Tersedia: 1},
		decrementErr: expectedErr,
	}}
	_, err = svc.LoanBook(context.Background(), bookID, "student", memberID, 7)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("LoanBook(decrement error) error = %v, want %v", err, expectedErr)
	}
}

func TestLibraryLoanBookCreatesStudentAndEmployeeLoans(t *testing.T) {
	bookID := "00000000-0000-0000-0000-000000000001"
	memberID := "00000000-0000-0000-0000-000000000002"
	parsedBookID := mustLibraryUUID(t, bookID)
	parsedMemberID := mustLibraryUUID(t, memberID)

	tests := []struct {
		name       string
		memberType string
	}{
		{name: "student", memberType: "student"},
		{name: "employee", memberType: "employee"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLibraryStore{
				book: db.LibraryBook{ID: parsedBookID, TotalEksemplar: 2, Tersedia: 1},
			}
			svc := &Library{q: store}

			loan, err := svc.LoanBook(context.Background(), bookID, tt.memberType, memberID, 7)
			if err != nil {
				t.Fatalf("LoanBook() error = %v", err)
			}
			if loan.BookID != parsedBookID || loan.MemberType != db.LibraryMemberType(tt.memberType) {
				t.Fatalf("LoanBook() loan = %+v, want parsed book/member type", loan)
			}
			if store.createLoanArg.BookID != parsedBookID || store.createLoanArg.DendaPerHari != 500 || !store.createLoanArg.JatuhTempo.Valid {
				t.Fatalf("CreateLoan() arg = %+v, want parsed book, due date, and fine", store.createLoanArg)
			}
			if tt.memberType == "student" {
				if store.createLoanArg.StudentID != parsedMemberID || store.createLoanArg.EmployeeID.Valid {
					t.Fatalf("CreateLoan() arg = %+v, want student member id only", store.createLoanArg)
				}
			} else {
				if store.createLoanArg.EmployeeID != parsedMemberID || store.createLoanArg.StudentID.Valid {
					t.Fatalf("CreateLoan() arg = %+v, want employee member id only", store.createLoanArg)
				}
			}
			if store.decrementID != parsedBookID {
				t.Fatalf("DecrementTersedia() id = %v, want %v", store.decrementID, parsedBookID)
			}
		})
	}
}

func TestLibraryReturnBookValidatesAndCalculatesFine(t *testing.T) {
	loanID := "00000000-0000-0000-0000-000000000010"
	bookID := libraryTestUUID(4)
	parsedLoanID := mustLibraryUUID(t, loanID)
	store := &fakeLibraryStore{
		getLoanRow: db.GetLoanRow{
			ID:           parsedLoanID,
			BookID:       bookID,
			Status:       db.LoanStatusEnumActive,
			JatuhTempo:   pgtype.Timestamptz{Time: time.Now().Add(-49 * time.Hour), Valid: true},
			DendaPerHari: 500,
		},
	}
	svc := &Library{q: store}

	returned, err := svc.ReturnBook(context.Background(), loanID)
	if err != nil {
		t.Fatalf("ReturnBook() error = %v", err)
	}
	if returned.Status != db.LoanStatusEnumReturned || returned.DendaTotal != 1500 {
		t.Fatalf("ReturnBook() = %+v, want returned loan with 1500 fine", returned)
	}
	if store.updateReturnArg.ID != parsedLoanID || store.updateReturnArg.DendaTotal != 1500 {
		t.Fatalf("UpdateLoanReturn() arg = %+v, want loan id and fine", store.updateReturnArg)
	}
	if store.incrementID != bookID {
		t.Fatalf("IncrementTersedia() id = %v, want %v", store.incrementID, bookID)
	}
}

func TestLibraryReturnBookRejectsInvalidOrNonActiveLoans(t *testing.T) {
	t.Run("ID data tidak valid", func(t *testing.T) {
		store := &fakeLibraryStore{}
		svc := &Library{q: store}

		_, err := svc.ReturnBook(context.Background(), "bad")
		if err == nil || err.Error() != "loan_id tidak valid" {
			t.Fatalf("ReturnBook() error = %v, want invalid loan id", err)
		}
	})

	t.Run("already returned", func(t *testing.T) {
		store := &fakeLibraryStore{
			getLoanRow: db.GetLoanRow{Status: db.LoanStatusEnumReturned},
		}
		svc := &Library{q: store}

		_, err := svc.ReturnBook(context.Background(), "00000000-0000-0000-0000-000000000010")
		if err == nil || err.Error() != "pinjaman sudah dikembalikan sebelumnya" {
			t.Fatalf("ReturnBook() error = %v, want already-returned guard", err)
		}
		if store.updateReturnArg.ID.Valid {
			t.Fatalf("UpdateLoanReturn() was called for non-active loan")
		}
	})
}

func TestLibraryReturnBookPropagatesStoreErrors(t *testing.T) {
	lookupErr := errors.New("lookup failed")
	store := &fakeLibraryStore{getLoanErr: lookupErr}
	svc := &Library{q: store}

	_, err := svc.ReturnBook(context.Background(), "00000000-0000-0000-0000-000000000010")
	if err == nil || err.Error() != "data pinjaman tidak ditemukan" {
		t.Fatalf("ReturnBook() error = %v, want public lookup error", err)
	}

	bookID := libraryTestUUID(42)
	loanID := mustLibraryUUID(t, "00000000-0000-0000-0000-000000000010")
	expectedErr := errors.New("store failed")
	svc = &Library{q: &fakeLibraryStore{
		getLoanRow: db.GetLoanRow{
			ID:         loanID,
			BookID:     bookID,
			Status:     db.LoanStatusEnumActive,
			JatuhTempo: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		},
		updateReturnErr: expectedErr,
	}}
	_, err = svc.ReturnBook(context.Background(), "00000000-0000-0000-0000-000000000010")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("ReturnBook(update error) error = %v, want %v", err, expectedErr)
	}

	svc = &Library{q: &fakeLibraryStore{
		getLoanRow: db.GetLoanRow{
			ID:         loanID,
			BookID:     bookID,
			Status:     db.LoanStatusEnumActive,
			JatuhTempo: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		},
		incrementErr: expectedErr,
	}}
	_, err = svc.ReturnBook(context.Background(), "00000000-0000-0000-0000-000000000010")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("ReturnBook(increment error) error = %v, want %v", err, expectedErr)
	}
}

func TestLibraryMarkDendaLunas(t *testing.T) {
	loanID := "00000000-0000-0000-0000-000000000020"
	parsedLoanID := mustLibraryUUID(t, loanID)
	store := &fakeLibraryStore{
		markDendaRow: db.LibraryLoan{ID: parsedLoanID, DendaLunas: true},
	}
	svc := &Library{q: store}

	loan, err := svc.MarkDendaLunas(context.Background(), loanID)
	if err != nil {
		t.Fatalf("MarkDendaLunas() error = %v", err)
	}
	if store.markDendaID != parsedLoanID || !loan.DendaLunas {
		t.Fatalf("MarkDendaLunas() loan = %+v, store id = %v; want paid loan", loan, store.markDendaID)
	}

	if _, err := svc.MarkDendaLunas(context.Background(), "bad"); err == nil || err.Error() != "loan_id tidak valid" {
		t.Fatalf("MarkDendaLunas(invalid) error = %v, want invalid id", err)
	}
}

func TestLibraryDelegatesLoanReadsAndStats(t *testing.T) {
	loanID := libraryTestUUID(12)
	store := &fakeLibraryStore{
		listLoansRows: []db.ListLoansRow{{ID: loanID, Status: db.LoanStatusEnumActive}},
		getLoanRow:    db.GetLoanRow{ID: loanID, Status: db.LoanStatusEnumActive},
		stats:         db.GetLibraryStatsRow{TotalJudul: 5, SedangDipinjam: 2},
	}
	svc := &Library{q: store}

	loans, err := svc.ListLoans(context.Background(), "active")
	if err != nil {
		t.Fatalf("ListLoans() error = %v", err)
	}
	if store.listLoansStatus != "active" || len(loans) != 1 {
		t.Fatalf("ListLoans() = %+v, filter = %q; want delegated active rows", loans, store.listLoansStatus)
	}

	loan, err := svc.GetLoan(context.Background(), loanID)
	if err != nil {
		t.Fatalf("GetLoan() error = %v", err)
	}
	if store.getLoanID != loanID || loan.ID != loanID {
		t.Fatalf("GetLoan() = %+v, store id = %v; want delegated loan", loan, store.getLoanID)
	}

	stats, err := svc.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalJudul != 5 || stats.SedangDipinjam != 2 {
		t.Fatalf("Stats() = %+v, want store stats", stats)
	}
}
