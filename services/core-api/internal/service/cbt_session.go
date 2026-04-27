package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtSession struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewCbtSession(pool *pgxpool.Pool) *CbtSession {
	return &CbtSession{q: db.New(pool), pool: pool}
}

func (s *CbtSession) List(ctx context.Context) ([]db.ListCbtExamSessionsRow, error) {
	rows, err := s.q.ListCbtExamSessions(ctx)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamSessionsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) Get(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error) {
	return s.q.GetCbtExamSession(ctx, id)
}

type CreateCbtSessionInput struct {
	PackageID      pgtype.UUID
	ClassID        pgtype.UUID
	Title          string
	ScheduledStart pgtype.Timestamptz
	ScheduledEnd   pgtype.Timestamptz
	Status         db.CbtSessionStatusEnum
}

func (s *CbtSession) Create(ctx context.Context, in CreateCbtSessionInput) (db.CbtExamSession, error) {
	return s.q.CreateCbtExamSession(ctx, db.CreateCbtExamSessionParams{
		PackageID:      in.PackageID,
		ClassID:        in.ClassID,
		Title:          in.Title,
		ScheduledStart: in.ScheduledStart,
		ScheduledEnd:   in.ScheduledEnd,
		Status:         in.Status,
	})
}

func (s *CbtSession) UpdateStatus(ctx context.Context, id pgtype.UUID, status db.CbtSessionStatusEnum) (db.CbtExamSession, error) {
	return s.q.UpdateCbtExamSessionStatus(ctx, db.UpdateCbtExamSessionStatusParams{
		ID:     id,
		Status: status,
	})
}

func (s *CbtSession) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCbtExamSession(ctx, id)
}

func (s *CbtSession) ListParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	rows, err := s.q.ListCbtExamParticipants(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamParticipantsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) EnrollClass(ctx context.Context, sessionID, classID pgtype.UUID) error {
	return s.q.EnrollClassToSession(ctx, db.EnrollClassToSessionParams{
		SessionID: sessionID,
		ClassID:   classID,
	})
}

func (s *CbtSession) RecordAnswer(ctx context.Context, participantID, questionID pgtype.UUID, answer string) error {
	return s.q.UpsertStudentAnswer(ctx, db.UpsertStudentAnswerParams{
		ParticipantID: participantID,
		QuestionID:    questionID,
		Answer:        answer,
	})
}

// ScoreSession marks is_correct for all answers then updates participant scores — runs in a transaction.
func (s *CbtSession) ScoreSession(ctx context.Context, sessionID pgtype.UUID) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.q.WithTx(tx)

	if err := qtx.UpdateAnswerCorrectness(ctx, sessionID); err != nil {
		return err
	}
	if err := qtx.UpdateParticipantScores(ctx, sessionID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *CbtSession) GetResults(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error) {
	rows, err := s.q.GetSessionResults(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetSessionResultsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error) {
	rows, err := s.q.GetParticipantAnswers(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetParticipantAnswersRow{}, nil
	}
	return rows, nil
}
