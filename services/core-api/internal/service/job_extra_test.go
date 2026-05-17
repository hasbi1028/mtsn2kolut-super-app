package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestJobExtraCompleteAffectedZeroBranches(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{21}, Valid: true}

	t.Run("already successful job is idempotent", func(t *testing.T) {
		base := &fakeJobStore{getRow: db.GetJobRow{ID: jobID, Status: db.JobStatusEnumSuccess}}
		store := &completeZeroJobStore{fakeJobStore: base}
		svc := &PusakaJob{q: store}
		if err := svc.Complete(context.Background(), jobID, "worker-1"); err != nil {
			t.Fatalf("Complete(already success) error = %v, want nil", err)
		}
		if base.getID != jobID {
			t.Fatalf("GetJob id = %v, want %v", base.getID, jobID)
		}
	})

	t.Run("non successful job is conflict", func(t *testing.T) {
		store := &completeZeroJobStore{fakeJobStore: &fakeJobStore{getRow: db.GetJobRow{ID: jobID, Status: db.JobStatusEnumRunning}}}
		svc := &PusakaJob{q: store}
		if err := svc.Complete(context.Background(), jobID, "worker-1"); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("Complete(running affected 0) error = %v, want conflict", err)
		}
	})

	t.Run("missing job is conflict", func(t *testing.T) {
		store := &completeZeroJobStore{fakeJobStore: &fakeJobStore{getErr: pgx.ErrNoRows}}
		svc := &PusakaJob{q: store}
		if err := svc.Complete(context.Background(), jobID, "worker-1"); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("Complete(missing affected 0) error = %v, want conflict", err)
		}
	})
}

type completeZeroJobStore struct {
	*fakeJobStore
}

func (s *completeZeroJobStore) CompleteJob(ctx context.Context, arg db.CompleteJobParams) (int64, error) {
	s.completeArg = arg
	return 0, nil
}

func TestJobExtraCompleteWithAttendanceNilUsesPlainComplete(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{22}, Valid: true}
	store := &fakeJobStore{}
	svc := &PusakaJob{q: completionlessJobStore{inner: store}}

	if err := svc.CompleteWithAttendance(context.Background(), jobID, "worker-2", nil); err != nil {
		t.Fatalf("CompleteWithAttendance(nil attendance) error = %v", err)
	}
	if store.completeArg.ID != jobID || store.completeArg.ClaimedBy != "worker-2" {
		t.Fatalf("CompleteJob arg = %+v, want plain Complete path", store.completeArg)
	}
}

type fakePusakaJobTxStarter struct {
	tx  *fakePusakaJobTx
	err error
}

func (s *fakePusakaJobTxStarter) Begin(ctx context.Context) (pgx.Tx, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.tx == nil {
		s.tx = &fakePusakaJobTx{}
	}
	return s.tx, nil
}

type fakePusakaJobTx struct {
	committed  bool
	rolledBack bool
}

func (tx *fakePusakaJobTx) Begin(ctx context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *fakePusakaJobTx) Commit(ctx context.Context) error {
	tx.committed = true
	return nil
}
func (tx *fakePusakaJobTx) Rollback(ctx context.Context) error {
	tx.rolledBack = true
	return nil
}
func (tx *fakePusakaJobTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom")
}
func (tx *fakePusakaJobTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (tx *fakePusakaJobTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (tx *fakePusakaJobTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare")
}
func (tx *fakePusakaJobTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}
func (tx *fakePusakaJobTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}
func (tx *fakePusakaJobTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return fakePusakaJobRow{}
}
func (tx *fakePusakaJobTx) Conn() *pgx.Conn { return nil }

type fakePusakaJobRow struct{}

func (fakePusakaJobRow) Scan(dest ...any) error { return errors.New("unexpected QueryRow") }

func TestJobExtraWithCompletionStoreTransactionLifecycle(t *testing.T) {
	t.Run("begin error", func(t *testing.T) {
		expected := errors.New("begin failed")
		svc := &PusakaJob{q: &fakeJobStore{}, tx: &fakePusakaJobTxStarter{err: expected}}
		if err := svc.withCompletionStore(context.Background(), func(store pusakaJobCompletionStore) error { return nil }); !errors.Is(err, expected) {
			t.Fatalf("withCompletionStore(begin error) = %v, want %v", err, expected)
		}
	})

	t.Run("commit after successful callback and deferred rollback", func(t *testing.T) {
		tx := &fakePusakaJobTx{}
		svc := &PusakaJob{q: &fakeJobStore{}, tx: &fakePusakaJobTxStarter{tx: tx}}
		called := false
		if err := svc.withCompletionStore(context.Background(), func(store pusakaJobCompletionStore) error {
			called = true
			if store == nil {
				t.Fatal("store is nil")
			}
			return nil
		}); err != nil {
			t.Fatalf("withCompletionStore(success) error = %v", err)
		}
		if !called || !tx.committed || !tx.rolledBack {
			t.Fatalf("called/committed/rolledBack = %v/%v/%v, want all true", called, tx.committed, tx.rolledBack)
		}
	})

	t.Run("callback error rolls back without commit", func(t *testing.T) {
		tx := &fakePusakaJobTx{}
		expected := errors.New("callback failed")
		svc := &PusakaJob{q: &fakeJobStore{}, tx: &fakePusakaJobTxStarter{tx: tx}}
		if err := svc.withCompletionStore(context.Background(), func(store pusakaJobCompletionStore) error { return expected }); !errors.Is(err, expected) {
			t.Fatalf("withCompletionStore(callback error) = %v, want %v", err, expected)
		}
		if tx.committed || !tx.rolledBack {
			t.Fatalf("committed/rolledBack = %v/%v, want false/true", tx.committed, tx.rolledBack)
		}
	})
}
