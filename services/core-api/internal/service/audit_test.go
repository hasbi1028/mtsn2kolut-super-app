package service

import (
	"context"
	"errors"
	"testing"
)

type fakeAuditStore struct {
	deleted int64
	err     error
	calls   int
}

func (f *fakeAuditStore) DeleteOldAuditLogs(ctx context.Context) (int64, error) {
	f.calls++
	if f.err != nil {
		return 0, f.err
	}
	return f.deleted, nil
}

func TestAuditCleanupOld(t *testing.T) {
	t.Run("success with deleted rows", func(t *testing.T) {
		store := &fakeAuditStore{deleted: 7}
		svc := &Audit{q: store}

		deleted, err := svc.CleanupOld(context.Background())
		if err != nil {
			t.Fatalf("CleanupOld() error = %v", err)
		}
		if deleted != 7 || store.calls != 1 {
			t.Fatalf("CleanupOld() = %d with %d calls, want 7 with 1 call", deleted, store.calls)
		}
	})

	t.Run("success with no deleted rows", func(t *testing.T) {
		store := &fakeAuditStore{}
		svc := &Audit{q: store}

		deleted, err := svc.CleanupOld(context.Background())
		if err != nil {
			t.Fatalf("CleanupOld() error = %v", err)
		}
		if deleted != 0 || store.calls != 1 {
			t.Fatalf("CleanupOld() = %d with %d calls, want 0 with 1 call", deleted, store.calls)
		}
	})

	t.Run("store error", func(t *testing.T) {
		storeErr := errors.New("delete failed")
		store := &fakeAuditStore{err: storeErr}
		svc := &Audit{q: store}

		deleted, err := svc.CleanupOld(context.Background())
		if !errors.Is(err, storeErr) {
			t.Fatalf("CleanupOld() error = %v, want %v", err, storeErr)
		}
		if deleted != 0 || store.calls != 1 {
			t.Fatalf("CleanupOld() = %d with %d calls, want 0 with 1 call", deleted, store.calls)
		}
	})
}
