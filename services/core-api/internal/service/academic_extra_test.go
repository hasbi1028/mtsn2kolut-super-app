package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeAcademicTxStarter struct {
	tx       *fakeAcademicTx
	beginErr error
	begins   int
}

func (f *fakeAcademicTxStarter) Begin(ctx context.Context) (pgx.Tx, error) {
	f.begins++
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	if f.tx == nil {
		f.tx = &fakeAcademicTx{}
	}
	return f.tx, nil
}

type fakeAcademicTx struct {
	commits   int
	rollbacks int
	commitErr error
}

func (f *fakeAcademicTx) Begin(ctx context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeAcademicTx) Commit(ctx context.Context) error {
	f.commits++
	return f.commitErr
}
func (f *fakeAcademicTx) Rollback(ctx context.Context) error {
	f.rollbacks++
	return nil
}
func (f *fakeAcademicTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	panic("unexpected CopyFrom")
}
func (f *fakeAcademicTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	panic("unexpected SendBatch")
}
func (f *fakeAcademicTx) LargeObjects() pgx.LargeObjects { panic("unexpected LargeObjects") }
func (f *fakeAcademicTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	panic("unexpected Prepare")
}
func (f *fakeAcademicTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}
func (f *fakeAcademicTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}
func (f *fakeAcademicTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	panic("unexpected QueryRow")
}
func (f *fakeAcademicTx) Conn() *pgx.Conn { return nil }

func TestAcademicExtraWithAcademicStoreWithoutTxUsesBaseStore(t *testing.T) {
	store := &fakeAcademicStore{}
	svc := &Academic{q: store}
	called := false
	err := svc.withAcademicStore(context.Background(), func(got academicStore) error {
		called = true
		if got != store {
			t.Fatalf("withAcademicStore store = %T, want base fake store", got)
		}
		return nil
	})
	if err != nil || !called {
		t.Fatalf("withAcademicStore(no tx) = %v called=%v, want nil/true", err, called)
	}
}

func TestAcademicExtraWithAcademicStoreTxCommitRollbackAndErrors(t *testing.T) {
	t.Run("begin error", func(t *testing.T) {
		beginErr := errors.New("begin failed")
		txStarter := &fakeAcademicTxStarter{beginErr: beginErr}
		svc := &Academic{q: &fakeAcademicStore{}, tx: txStarter}
		err := svc.withAcademicStore(context.Background(), func(academicStore) error {
			t.Fatal("callback should not be called on begin error")
			return nil
		})
		if !errors.Is(err, beginErr) || txStarter.begins != 1 {
			t.Fatalf("withAcademicStore(begin error) = %v begins=%d, want beginErr/1", err, txStarter.begins)
		}
	})

	t.Run("callback error rolls back without commit", func(t *testing.T) {
		callbackErr := errors.New("callback failed")
		tx := &fakeAcademicTx{}
		txStarter := &fakeAcademicTxStarter{tx: tx}
		svc := &Academic{q: &fakeAcademicStore{}, tx: txStarter}
		err := svc.withAcademicStore(context.Background(), func(got academicStore) error {
			if got == svc.q {
				t.Fatal("callback got base store, want transaction-backed store")
			}
			return callbackErr
		})
		if !errors.Is(err, callbackErr) || tx.commits != 0 || tx.rollbacks != 1 {
			t.Fatalf("withAcademicStore(callback error) err=%v commits=%d rollbacks=%d, want callbackErr/0/1", err, tx.commits, tx.rollbacks)
		}
	})

	t.Run("commit error is returned after deferred rollback", func(t *testing.T) {
		commitErr := errors.New("commit failed")
		tx := &fakeAcademicTx{commitErr: commitErr}
		txStarter := &fakeAcademicTxStarter{tx: tx}
		svc := &Academic{q: &fakeAcademicStore{}, tx: txStarter}
		err := svc.withAcademicStore(context.Background(), func(academicStore) error { return nil })
		if !errors.Is(err, commitErr) || tx.commits != 1 || tx.rollbacks != 1 {
			t.Fatalf("withAcademicStore(commit error) err=%v commits=%d rollbacks=%d, want commitErr/1/1", err, tx.commits, tx.rollbacks)
		}
	})

	t.Run("success commits and deferred rollback is safe", func(t *testing.T) {
		tx := &fakeAcademicTx{}
		txStarter := &fakeAcademicTxStarter{tx: tx}
		svc := &Academic{q: &fakeAcademicStore{}, tx: txStarter}
		err := svc.withAcademicStore(context.Background(), func(academicStore) error { return nil })
		if err != nil || tx.commits != 1 || tx.rollbacks != 1 {
			t.Fatalf("withAcademicStore(success) err=%v commits=%d rollbacks=%d, want nil/1/1", err, tx.commits, tx.rollbacks)
		}
	})
}

var _ pgx.Tx = (*fakeAcademicTx)(nil)
