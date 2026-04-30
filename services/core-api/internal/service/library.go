package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type libraryStore interface {
	ListBooks(ctx context.Context, arg db.ListBooksParams) ([]db.LibraryBook, error)
	GetBook(ctx context.Context, id pgtype.UUID) (db.LibraryBook, error)
	CreateBook(ctx context.Context, arg db.CreateBookParams) (db.LibraryBook, error)
	UpdateBook(ctx context.Context, arg db.UpdateBookParams) (db.LibraryBook, error)
	DeleteBook(ctx context.Context, id pgtype.UUID) error
	DecrementTersedia(ctx context.Context, id pgtype.UUID) error
	IncrementTersedia(ctx context.Context, id pgtype.UUID) error
	ListLoans(ctx context.Context, filterStatus string) ([]db.ListLoansRow, error)
	GetLoan(ctx context.Context, id pgtype.UUID) (db.GetLoanRow, error)
	CreateLoan(ctx context.Context, arg db.CreateLoanParams) (db.LibraryLoan, error)
	UpdateLoanReturn(ctx context.Context, arg db.UpdateLoanReturnParams) (db.LibraryLoan, error)
	MarkLoanDendaLunas(ctx context.Context, id pgtype.UUID) (db.LibraryLoan, error)
	CountActiveLoansForMember(ctx context.Context, memberID pgtype.UUID) (int64, error)
	GetLibraryStats(ctx context.Context) (db.GetLibraryStatsRow, error)
}

type Library struct{ q libraryStore }

func NewLibrary(q *db.Queries) *Library { return &Library{q: q} }

func (s *Library) ListBooks(ctx context.Context, search, kategori string) ([]db.LibraryBook, error) {
	return s.q.ListBooks(ctx, db.ListBooksParams{
		Search:   strings.TrimSpace(search),
		Kategori: strings.TrimSpace(kategori),
	})
}

func (s *Library) GetBook(ctx context.Context, id pgtype.UUID) (db.LibraryBook, error) {
	return s.q.GetBook(ctx, id)
}

func (s *Library) CreateBook(ctx context.Context, arg db.CreateBookParams) (db.LibraryBook, error) {
	arg.Kode = strings.TrimSpace(arg.Kode)
	arg.Judul = strings.TrimSpace(arg.Judul)
	if arg.Kode == "" {
		return db.LibraryBook{}, fmt.Errorf("kode buku wajib diisi")
	}
	if arg.Judul == "" {
		return db.LibraryBook{}, fmt.Errorf("judul buku wajib diisi")
	}
	if arg.TotalEksemplar < 1 {
		return db.LibraryBook{}, fmt.Errorf("jumlah eksemplar minimal 1")
	}
	return s.q.CreateBook(ctx, arg)
}

func (s *Library) UpdateBook(ctx context.Context, arg db.UpdateBookParams) (db.LibraryBook, error) {
	arg.Kode = strings.TrimSpace(arg.Kode)
	arg.Judul = strings.TrimSpace(arg.Judul)
	if arg.Kode == "" {
		return db.LibraryBook{}, fmt.Errorf("kode buku wajib diisi")
	}
	if arg.Judul == "" {
		return db.LibraryBook{}, fmt.Errorf("judul buku wajib diisi")
	}
	if arg.TotalEksemplar < 1 {
		return db.LibraryBook{}, fmt.Errorf("jumlah eksemplar minimal 1")
	}
	return s.q.UpdateBook(ctx, arg)
}

func (s *Library) DeleteBook(ctx context.Context, id pgtype.UUID) error {
	book, err := s.q.GetBook(ctx, id)
	if err != nil {
		return err
	}
	if book.Tersedia != book.TotalEksemplar {
		return fmt.Errorf("buku masih dipinjam, tidak bisa dihapus")
	}
	return s.q.DeleteBook(ctx, id)
}

func (s *Library) ListLoans(ctx context.Context, filterStatus string) ([]db.ListLoansRow, error) {
	return s.q.ListLoans(ctx, filterStatus)
}

func (s *Library) GetLoan(ctx context.Context, id pgtype.UUID) (db.GetLoanRow, error) {
	return s.q.GetLoan(ctx, id)
}

func (s *Library) LoanBook(ctx context.Context, bookID, memberType, memberID string, dueDays int) (db.LibraryLoan, error) {
	bid, err := parseLibraryUUID(bookID)
	if err != nil {
		return db.LibraryLoan{}, fmt.Errorf("book_id tidak valid")
	}
	mid, err := parseLibraryUUID(memberID)
	if err != nil {
		return db.LibraryLoan{}, fmt.Errorf("member_id tidak valid")
	}
	if memberType != "student" && memberType != "employee" {
		return db.LibraryLoan{}, fmt.Errorf("member_type harus 'student' atau 'employee'")
	}
	if dueDays < 1 || dueDays > 30 {
		return db.LibraryLoan{}, fmt.Errorf("durasi pinjam harus antara 1-30 hari")
	}

	book, err := s.q.GetBook(ctx, bid)
	if err != nil {
		return db.LibraryLoan{}, fmt.Errorf("buku tidak ditemukan")
	}
	if book.Tersedia < 1 {
		return db.LibraryLoan{}, fmt.Errorf("stok buku habis, tidak ada eksemplar yang tersedia")
	}

	activeCount, err := s.q.CountActiveLoansForMember(ctx, mid)
	if err != nil {
		return db.LibraryLoan{}, err
	}
	if activeCount >= 3 {
		return db.LibraryLoan{}, fmt.Errorf("anggota sudah memiliki 3 pinjaman aktif, kembalikan dulu sebelum meminjam lagi")
	}

	due := pgtype.Timestamptz{}
	if err := due.Scan(time.Now().Add(time.Duration(dueDays) * 24 * time.Hour)); err != nil {
		return db.LibraryLoan{}, err
	}

	params := db.CreateLoanParams{
		BookID:       bid,
		MemberType:   db.LibraryMemberType(memberType),
		JatuhTempo:   due,
		DendaPerHari: 500,
	}
	if memberType == "student" {
		params.StudentID = mid
	} else {
		params.EmployeeID = mid
	}

	loan, err := s.q.CreateLoan(ctx, params)
	if err != nil {
		return db.LibraryLoan{}, err
	}
	if err := s.q.DecrementTersedia(ctx, bid); err != nil {
		return db.LibraryLoan{}, err
	}
	return loan, nil
}

func (s *Library) ReturnBook(ctx context.Context, loanID string) (db.LibraryLoan, error) {
	lid, err := parseLibraryUUID(loanID)
	if err != nil {
		return db.LibraryLoan{}, fmt.Errorf("loan_id tidak valid")
	}

	loan, err := s.q.GetLoan(ctx, lid)
	if err != nil {
		return db.LibraryLoan{}, fmt.Errorf("data pinjaman tidak ditemukan")
	}
	if loan.Status != db.LoanStatusEnumActive {
		return db.LibraryLoan{}, fmt.Errorf("pinjaman sudah dikembalikan sebelumnya")
	}

	var denda int32
	due := loan.JatuhTempo.Time
	now := time.Now()
	if now.After(due) {
		days := int32(math.Ceil(now.Sub(due).Hours() / 24))
		denda = days * loan.DendaPerHari
	}

	returned, err := s.q.UpdateLoanReturn(ctx, db.UpdateLoanReturnParams{
		ID:         lid,
		DendaTotal: denda,
	})
	if err != nil {
		return db.LibraryLoan{}, err
	}
	if err := s.q.IncrementTersedia(ctx, loan.BookID); err != nil {
		return db.LibraryLoan{}, err
	}
	return returned, nil
}

func (s *Library) MarkDendaLunas(ctx context.Context, loanID string) (db.LibraryLoan, error) {
	lid, err := parseLibraryUUID(loanID)
	if err != nil {
		return db.LibraryLoan{}, fmt.Errorf("loan_id tidak valid")
	}
	return s.q.MarkLoanDendaLunas(ctx, lid)
}

func (s *Library) Stats(ctx context.Context) (db.GetLibraryStatsRow, error) {
	return s.q.GetLibraryStats(ctx)
}

func parseLibraryUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}
