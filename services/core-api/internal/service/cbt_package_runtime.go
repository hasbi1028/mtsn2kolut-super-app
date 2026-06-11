package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type cbtPackageSnapshotStore interface {
	LockCbtPackageForSnapshot(ctx context.Context, arg db.LockCbtPackageForSnapshotParams) (db.LockCbtPackageForSnapshotRow, error)
	CreateCbtPackageQuestionSnapshots(ctx context.Context, packageID pgtype.UUID) (int64, error)
}

type CbtPackageSnapshotResult struct {
	PackageID         string `json:"package_id"`
	LockedAt          string `json:"locked_at"`
	LockReason        string `json:"lock_reason"`
	SnapshotVersion   int32  `json:"snapshot_version"`
	SnapshotRowsAdded int64  `json:"snapshot_rows_added"`
}

func lockCbtPackageSnapshot(ctx context.Context, q cbtPackageSnapshotStore, packageID, lockedBy pgtype.UUID, reason string) (CbtPackageSnapshotResult, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "session_scheduled"
	}
	row, err := q.LockCbtPackageForSnapshot(ctx, db.LockCbtPackageForSnapshotParams{
		PackageID:  packageID,
		LockedBy:   lockedBy,
		LockReason: reason,
	})
	if err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	inserted, err := q.CreateCbtPackageQuestionSnapshots(ctx, packageID)
	if err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	lockedAt := ""
	if row.LockedAt.Valid {
		lockedAt = row.LockedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return CbtPackageSnapshotResult{
		PackageID:         pgUUIDString(row.ID),
		LockedAt:          lockedAt,
		LockReason:        row.LockReason,
		SnapshotVersion:   row.SnapshotVersion,
		SnapshotRowsAdded: inserted,
	}, nil
}
