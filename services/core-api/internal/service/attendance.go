package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type pusakaAttendanceStore interface {
	ListAttendance(ctx context.Context, arg db.ListAttendanceParams) ([]db.ListAttendanceRow, error)
	CountAttendance(ctx context.Context) (int64, error)
	ListAttendanceByDate(ctx context.Context, tanggal pgtype.Date) ([]db.ListAttendanceByDateRow, error)
	ListAttendanceInRange(ctx context.Context, arg db.ListAttendanceInRangeParams) ([]db.ListAttendanceInRangeRow, error)
	ListAttendanceByEmployee(ctx context.Context, arg db.ListAttendanceByEmployeeParams) ([]db.ListAttendanceByEmployeeRow, error)
	GetMonthlyAttendanceSummary(ctx context.Context, arg db.GetMonthlyAttendanceSummaryParams) ([]db.GetMonthlyAttendanceSummaryRow, error)
	UpsertAttendance(ctx context.Context, arg db.UpsertAttendanceParams) (db.AttendanceRecord, error)
}

type PusakaAttendance struct {
	q pusakaAttendanceStore
}

func NewPusakaAttendance(q *db.Queries) *PusakaAttendance { return &PusakaAttendance{q: q} }

func (s *PusakaAttendance) List(ctx context.Context, limit, offset int32) ([]db.ListAttendanceRow, int64, error) {
	rows, err := s.q.ListAttendance(ctx, db.ListAttendanceParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	count, err := s.q.CountAttendance(ctx)
	return rows, count, err
}

func (s *PusakaAttendance) ByDate(ctx context.Context, date pgtype.Date) ([]db.ListAttendanceByDateRow, error) {
	return s.q.ListAttendanceByDate(ctx, date)
}

func (s *PusakaAttendance) ListInRange(ctx context.Context, start, end pgtype.Date) ([]db.ListAttendanceInRangeRow, error) {
	return s.q.ListAttendanceInRange(ctx, db.ListAttendanceInRangeParams{
		Tanggal:   start,
		Tanggal_2: end,
	})
}

func (s *PusakaAttendance) ByEmployee(ctx context.Context, empID pgtype.UUID, limit, offset int32) ([]db.ListAttendanceByEmployeeRow, error) {
	return s.q.ListAttendanceByEmployee(ctx, db.ListAttendanceByEmployeeParams{
		EmployeeID: empID,
		Limit:      limit,
		Offset:     offset,
	})
}

func (s *PusakaAttendance) GetSummary(ctx context.Context, start, end pgtype.Date) ([]db.GetMonthlyAttendanceSummaryRow, error) {
	return s.q.GetMonthlyAttendanceSummary(ctx, db.GetMonthlyAttendanceSummaryParams{
		Tanggal:   start,
		Tanggal_2: end,
	})
}

func (s *PusakaAttendance) Upsert(ctx context.Context, p db.UpsertAttendanceParams) (db.AttendanceRecord, error) {
	return s.q.UpsertAttendance(ctx, p)
}
