package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeRBACTxStarter struct {
	tx       *fakeRBACTx
	beginErr error
}

func (f *fakeRBACTxStarter) Begin(ctx context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	if f.tx == nil {
		f.tx = &fakeRBACTx{}
	}
	return f.tx, nil
}

type fakeRBACTx struct {
	commitErr error
	commits   int
	rollbacks int
}

func (f *fakeRBACTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, errors.New("nested tx not implemented")
}
func (f *fakeRBACTx) Commit(ctx context.Context) error {
	f.commits++
	return f.commitErr
}
func (f *fakeRBACTx) Rollback(ctx context.Context) error {
	f.rollbacks++
	return nil
}
func (f *fakeRBACTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("copy from not implemented")
}
func (f *fakeRBACTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeRBACTx) LargeObjects() pgx.LargeObjects                               { return pgx.LargeObjects{} }
func (f *fakeRBACTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("prepare not implemented")
}
func (f *fakeRBACTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("exec not implemented")
}
func (f *fakeRBACTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, errors.New("query not implemented")
}
func (f *fakeRBACTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (f *fakeRBACTx) Conn() *pgx.Conn                                               { return nil }

func TestRBACExtraWithRBACStoreNoTxUsesBaseStore(t *testing.T) {
	store := &fakeRBACStore{}
	svc := &RBAC{q: store}
	called := false

	if err := svc.withRBACStore(context.Background(), func(got rbacStore) error {
		called = true
		if got != store {
			t.Fatalf("store = %T, want base fake store", got)
		}
		return nil
	}); err != nil {
		t.Fatalf("withRBACStore() error = %v", err)
	}
	if !called {
		t.Fatal("withRBACStore() did not invoke callback")
	}
}

func TestRBACExtraWithRBACStoreBeginErrorSkipsCallback(t *testing.T) {
	expected := errors.New("begin failed")
	svc := &RBAC{q: &fakeRBACStore{}, tx: &fakeRBACTxStarter{beginErr: expected}}
	called := false

	err := svc.withRBACStore(context.Background(), func(store rbacStore) error {
		called = true
		return nil
	})
	if !errors.Is(err, expected) {
		t.Fatalf("withRBACStore() error = %v, want %v", err, expected)
	}
	if called {
		t.Fatal("callback was invoked after Begin error")
	}
}

func TestRBACExtraWithRBACStoreCallbackErrorRollsBackWithoutCommit(t *testing.T) {
	expected := errors.New("mutation failed")
	tx := &fakeRBACTx{}
	svc := &RBAC{q: &fakeRBACStore{}, tx: &fakeRBACTxStarter{tx: tx}}

	err := svc.withRBACStore(context.Background(), func(store rbacStore) error {
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("withRBACStore() error = %v, want %v", err, expected)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Fatalf("tx commits=%d rollbacks=%d, want 0 commit and 1 deferred rollback", tx.commits, tx.rollbacks)
	}
}

func TestRBACExtraWithRBACStoreCommitErrorIsReturned(t *testing.T) {
	expected := errors.New("commit failed")
	tx := &fakeRBACTx{commitErr: expected}
	svc := &RBAC{q: &fakeRBACStore{}, tx: &fakeRBACTxStarter{tx: tx}}

	err := svc.withRBACStore(context.Background(), func(store rbacStore) error { return nil })
	if !errors.Is(err, expected) {
		t.Fatalf("withRBACStore() error = %v, want %v", err, expected)
	}
	if tx.commits != 1 || tx.rollbacks != 1 {
		t.Fatalf("tx commits=%d rollbacks=%d, want 1 commit and 1 deferred rollback", tx.commits, tx.rollbacks)
	}
}
