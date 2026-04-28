package service

import (
	"context"
	"log/slog"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Audit struct {
	q *db.Queries
}

func NewAudit(q *db.Queries) *Audit {
	return &Audit{q: q}
}

func (s *Audit) CleanupOld(ctx context.Context) (int64, error) {
	deleted, err := s.q.DeleteOldAuditLogs(ctx)
	if err != nil {
		slog.Error("audit cleanup failed", "error", err)
		return 0, err
	}
	if deleted > 0 {
		slog.Info("audit cleanup: deleted old entries", "deleted", deleted)
	}
	return deleted, nil
}
