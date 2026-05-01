package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type parentStore interface {
	ListParents(ctx context.Context) ([]db.Parent, error)
	GetParent(ctx context.Context, id pgtype.UUID) (db.Parent, error)
	CreateParent(ctx context.Context, arg db.CreateParentParams) (db.Parent, error)
	UpdateParent(ctx context.Context, arg db.UpdateParentParams) (db.Parent, error)
	DeleteParent(ctx context.Context, id pgtype.UUID) error
	LinkParentStudent(ctx context.Context, arg db.LinkParentStudentParams) error
	UnlinkParentStudent(ctx context.Context, arg db.UnlinkParentStudentParams) error
	ListParentChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error)
}

type Parent struct {
	q parentStore
}

func NewParent(q *db.Queries) *Parent { return &Parent{q: q} }

func (s *Parent) List(ctx context.Context) ([]db.Parent, error) {
	return s.q.ListParents(ctx)
}

func (s *Parent) Get(ctx context.Context, id pgtype.UUID) (db.Parent, error) {
	return s.q.GetParent(ctx, id)
}

func (s *Parent) Create(ctx context.Context, nama, phone, address string) (db.Parent, error) {
	return s.q.CreateParent(ctx, db.CreateParentParams{
		Nama:    nama,
		Phone:   phone,
		Address: address,
	})
}

func (s *Parent) Update(ctx context.Context, id pgtype.UUID, nama, phone, address string) (db.Parent, error) {
	return s.q.UpdateParent(ctx, db.UpdateParentParams{
		ID:      id,
		Nama:    nama,
		Phone:   phone,
		Address: address,
	})
}

func (s *Parent) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteParent(ctx, id)
}

func (s *Parent) LinkStudent(ctx context.Context, parentID, studentID pgtype.UUID) error {
	return s.q.LinkParentStudent(ctx, db.LinkParentStudentParams{
		ParentID:  parentID,
		StudentID: studentID,
	})
}

func (s *Parent) UnlinkStudent(ctx context.Context, parentID, studentID pgtype.UUID) error {
	return s.q.UnlinkParentStudent(ctx, db.UnlinkParentStudentParams{
		ParentID:  parentID,
		StudentID: studentID,
	})
}

func (s *Parent) ListChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	return s.q.ListParentChildren(ctx, parentID)
}
