package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type cbtApprovalStore interface {
	ListCbtApprovalRecords(ctx context.Context, arg db.ListCbtApprovalRecordsParams) ([]db.ListCbtApprovalRecordsRow, error)
	CreateCbtApprovalRecord(ctx context.Context, arg db.CreateCbtApprovalRecordParams) (db.CbtApprovalRecord, error)
	RevokeCbtApprovalRecord(ctx context.Context, arg db.RevokeCbtApprovalRecordParams) (db.CbtApprovalRecord, error)
}

type CbtApproval struct {
	q cbtApprovalStore
}

func NewCbtApproval(q *db.Queries) *CbtApproval {
	return &CbtApproval{q: q}
}

type ListCbtApprovalInput struct {
	EntityType string
	EntityID   pgtype.UUID
}

type SaveCbtApprovalInput struct {
	EntityType   string
	EntityID     pgtype.UUID
	ApprovalType string
	Notes        string
	ActorUserID  pgtype.UUID
}

func (s *CbtApproval) List(ctx context.Context, in ListCbtApprovalInput) ([]db.ListCbtApprovalRecordsRow, error) {
	entityType := strings.TrimSpace(in.EntityType)
	if entityType != "" && !validCbtApprovalEntityType(entityType) {
		return nil, errors.Join(domain.ErrBadRequest, errors.New("jenis data pengesahan tidak valid"))
	}
	rows, err := s.q.ListCbtApprovalRecords(ctx, db.ListCbtApprovalRecordsParams{
		EntityType: pgtype.Text{String: entityType, Valid: entityType != ""},
		EntityID:   in.EntityID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtApprovalRecordsRow{}, nil
	}
	return rows, nil
}

func (s *CbtApproval) Approve(ctx context.Context, in SaveCbtApprovalInput) (db.CbtApprovalRecord, error) {
	in = normalizeCbtApprovalInput(in)
	if err := validateCbtApprovalInput(in); err != nil {
		return db.CbtApprovalRecord{}, err
	}
	return s.q.CreateCbtApprovalRecord(ctx, db.CreateCbtApprovalRecordParams{
		EntityType:   in.EntityType,
		EntityID:     in.EntityID,
		ApprovalType: in.ApprovalType,
		ApprovedBy:   in.ActorUserID,
		Notes:        in.Notes,
	})
}

func (s *CbtApproval) Revoke(ctx context.Context, id pgtype.UUID, actorUserID pgtype.UUID, notes string) (db.CbtApprovalRecord, error) {
	if !id.Valid {
		return db.CbtApprovalRecord{}, errors.Join(domain.ErrBadRequest, errors.New("id pengesahan tidak valid"))
	}
	if !actorUserID.Valid {
		return db.CbtApprovalRecord{}, errors.Join(domain.ErrUnauthorized, errors.New("aktor pengesahan tidak valid"))
	}
	return s.q.RevokeCbtApprovalRecord(ctx, db.RevokeCbtApprovalRecordParams{
		ID:        id,
		RevokedBy: actorUserID,
		Notes:     strings.TrimSpace(notes),
	})
}

func normalizeCbtApprovalInput(in SaveCbtApprovalInput) SaveCbtApprovalInput {
	in.EntityType = strings.TrimSpace(in.EntityType)
	in.ApprovalType = strings.TrimSpace(in.ApprovalType)
	in.Notes = strings.TrimSpace(in.Notes)
	return in
}

func validateCbtApprovalInput(in SaveCbtApprovalInput) error {
	if !validCbtApprovalEntityType(in.EntityType) {
		return errors.Join(domain.ErrBadRequest, errors.New("jenis data pengesahan tidak valid"))
	}
	if !in.EntityID.Valid {
		return errors.Join(domain.ErrBadRequest, errors.New("id data pengesahan tidak valid"))
	}
	if !validCbtApprovalType(in.ApprovalType) {
		return errors.Join(domain.ErrBadRequest, errors.New("jenis pengesahan tidak valid"))
	}
	if !in.ActorUserID.Valid {
		return errors.Join(domain.ErrUnauthorized, errors.New("aktor pengesahan tidak valid"))
	}
	return nil
}

func validCbtApprovalEntityType(value string) bool {
	switch value {
	case "event", "session", "package", "result":
		return true
	default:
		return false
	}
}

func validCbtApprovalType(value string) bool {
	switch value {
	case "package_ready", "participants_rooms_ready", "tokens_cards_ready", "results_verified", "final_archive", "session_minutes", "room_handover":
		return true
	default:
		return false
	}
}
