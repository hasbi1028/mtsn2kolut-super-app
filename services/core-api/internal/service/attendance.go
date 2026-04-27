package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Attendance struct {
	q *db.Queries
}

func NewAttendance(q *db.Queries) *Attendance { return &Attendance{q: q} }

func (s *Attendance) List(ctx context.Context, limit, offset int32) ([]db.ListAttendanceRow, int64, error) {
	rows, err := s.q.ListAttendance(ctx, db.ListAttendanceParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	count, err := s.q.CountAttendance(ctx)
	return rows, count, err
}

func (s *Attendance) ByDate(ctx context.Context, date pgtype.Date) ([]db.ListAttendanceByDateRow, error) {
	return s.q.ListAttendanceByDate(ctx, date)
}

func (s *Attendance) ByEmployee(ctx context.Context, empID pgtype.UUID, limit, offset int32) ([]db.ListAttendanceByEmployeeRow, error) {
	return s.q.ListAttendanceByEmployee(ctx, db.ListAttendanceByEmployeeParams{
		EmployeeID: empID,
		Limit:      limit,
		Offset:     offset,
	})
}

func (s *Attendance) Upsert(ctx context.Context, p db.UpsertAttendanceParams) (db.AttendanceRecord, error) {
	return s.q.UpsertAttendance(ctx, p)
}
