package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type cbtQuestionMutationFakeTxStarter struct {
	tx       *cbtQuestionMutationFakeTx
	beginErr error
	begins   int
}

func (s *cbtQuestionMutationFakeTxStarter) Begin(ctx context.Context) (pgx.Tx, error) {
	s.begins++
	if s.beginErr != nil {
		return nil, s.beginErr
	}
	return s.tx, nil
}

type cbtQuestionMutationFakeTx struct {
	commitErr error
	commits   int
	rollbacks int
}

func (tx *cbtQuestionMutationFakeTx) Begin(ctx context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *cbtQuestionMutationFakeTx) Commit(ctx context.Context) error {
	tx.commits++
	return tx.commitErr
}
func (tx *cbtQuestionMutationFakeTx) Rollback(ctx context.Context) error {
	tx.rollbacks++
	return nil
}
func (tx *cbtQuestionMutationFakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom")
}
func (tx *cbtQuestionMutationFakeTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}
func (tx *cbtQuestionMutationFakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }
func (tx *cbtQuestionMutationFakeTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare")
}
func (tx *cbtQuestionMutationFakeTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}
func (tx *cbtQuestionMutationFakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}
func (tx *cbtQuestionMutationFakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}
func (tx *cbtQuestionMutationFakeTx) Conn() *pgx.Conn { return nil }

func TestCbtQuestionWithMutationStoreWithoutTxUsesBaseStore(t *testing.T) {
	ctx := context.Background()
	svc := &CbtQuestion{}
	want := db.CbtQuestion{Code: "Q-001"}
	called := false

	got, err := svc.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		called = true
		if store != nil {
			t.Fatalf("expected nil base store to be passed through, got %T", store)
		}
		return want, nil
	})
	if err != nil {
		t.Fatalf("withMutationStore returned error: %v", err)
	}
	if !called {
		t.Fatal("callback was not called")
	}
	if got.Code != want.Code {
		t.Fatalf("got code %q, want %q", got.Code, want.Code)
	}
}

func TestCbtQuestionWithMutationStoreWithoutTxReturnsCallbackError(t *testing.T) {
	ctx := context.Background()
	svc := &CbtQuestion{}
	wantErr := errors.New("callback failed")

	got, err := svc.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		return db.CbtQuestion{Code: "returned-before-error"}, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if got.Code != "returned-before-error" {
		t.Fatalf("got code %q, want callback row to pass through without tx", got.Code)
	}
}

func TestCbtQuestionWithMutationStoreBeginError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("begin failed")
	starter := &cbtQuestionMutationFakeTxStarter{beginErr: wantErr}
	svc := &CbtQuestion{tx: starter}
	called := false

	_, err := svc.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		called = true
		return db.CbtQuestion{}, nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if called {
		t.Fatal("callback should not be called when Begin fails")
	}
	if starter.begins != 1 {
		t.Fatalf("begins = %d, want 1", starter.begins)
	}
}

func TestCbtQuestionWithMutationStoreCommitsOnSuccess(t *testing.T) {
	ctx := context.Background()
	tx := &cbtQuestionMutationFakeTx{}
	starter := &cbtQuestionMutationFakeTxStarter{tx: tx}
	svc := &CbtQuestion{tx: starter}
	want := db.CbtQuestion{Code: "Q-OK"}

	got, err := svc.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		if store == nil {
			t.Fatal("expected transactional store, got nil")
		}
		return want, nil
	})
	if err != nil {
		t.Fatalf("withMutationStore returned error: %v", err)
	}
	if got.Code != want.Code {
		t.Fatalf("got code %q, want %q", got.Code, want.Code)
	}
	if starter.begins != 1 || tx.commits != 1 || tx.rollbacks != 0 {
		t.Fatalf("begins/commits/rollbacks = %d/%d/%d, want 1/1/0", starter.begins, tx.commits, tx.rollbacks)
	}
}

func TestCbtQuestionWithMutationStoreRollsBackOnCallbackError(t *testing.T) {
	ctx := context.Background()
	tx := &cbtQuestionMutationFakeTx{}
	starter := &cbtQuestionMutationFakeTxStarter{tx: tx}
	svc := &CbtQuestion{tx: starter}
	wantErr := errors.New("callback failed")

	got, err := svc.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		if store == nil {
			t.Fatal("expected transactional store, got nil")
		}
		return db.CbtQuestion{Code: "ignored"}, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if got.Code != "" {
		t.Fatalf("got row %+v, want zero value", got)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Fatalf("commits/rollbacks = %d/%d, want 0/1", tx.commits, tx.rollbacks)
	}
}

func TestCbtQuestionWithMutationStoreRollsBackOnCommitError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("commit failed")
	tx := &cbtQuestionMutationFakeTx{commitErr: wantErr}
	starter := &cbtQuestionMutationFakeTxStarter{tx: tx}
	svc := &CbtQuestion{tx: starter}

	got, err := svc.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		return db.CbtQuestion{Code: "ignored"}, nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if got.Code != "" {
		t.Fatalf("got row %+v, want zero value", got)
	}
	if tx.commits != 1 || tx.rollbacks != 1 {
		t.Fatalf("commits/rollbacks = %d/%d, want 1/1", tx.commits, tx.rollbacks)
	}
}

func TestCbtQuestionWithMutationStoreExecWithoutTxUsesBaseStore(t *testing.T) {
	ctx := context.Background()
	svc := &CbtQuestion{}
	called := false

	err := svc.withMutationStoreExec(ctx, func(store cbtQuestionStore) error {
		called = true
		if store != nil {
			t.Fatalf("expected nil base store to be passed through, got %T", store)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withMutationStoreExec returned error: %v", err)
	}
	if !called {
		t.Fatal("callback was not called")
	}
}

func TestCbtQuestionWithMutationStoreExecBeginError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("begin failed")
	starter := &cbtQuestionMutationFakeTxStarter{beginErr: wantErr}
	svc := &CbtQuestion{tx: starter}
	called := false

	err := svc.withMutationStoreExec(ctx, func(store cbtQuestionStore) error {
		called = true
		return nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if called {
		t.Fatal("callback should not be called when Begin fails")
	}
	if starter.begins != 1 {
		t.Fatalf("begins = %d, want 1", starter.begins)
	}
}

func TestCbtQuestionWithMutationStoreExecCommitsOnSuccess(t *testing.T) {
	ctx := context.Background()
	tx := &cbtQuestionMutationFakeTx{}
	starter := &cbtQuestionMutationFakeTxStarter{tx: tx}
	svc := &CbtQuestion{tx: starter}

	err := svc.withMutationStoreExec(ctx, func(store cbtQuestionStore) error {
		if store == nil {
			t.Fatal("expected transactional store, got nil")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("withMutationStoreExec returned error: %v", err)
	}
	if starter.begins != 1 || tx.commits != 1 || tx.rollbacks != 0 {
		t.Fatalf("begins/commits/rollbacks = %d/%d/%d, want 1/1/0", starter.begins, tx.commits, tx.rollbacks)
	}
}

func TestCbtQuestionWithMutationStoreExecRollsBackOnCallbackError(t *testing.T) {
	ctx := context.Background()
	tx := &cbtQuestionMutationFakeTx{}
	starter := &cbtQuestionMutationFakeTxStarter{tx: tx}
	svc := &CbtQuestion{tx: starter}
	wantErr := errors.New("callback failed")

	err := svc.withMutationStoreExec(ctx, func(store cbtQuestionStore) error {
		if store == nil {
			t.Fatal("expected transactional store, got nil")
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if tx.commits != 0 || tx.rollbacks != 1 {
		t.Fatalf("commits/rollbacks = %d/%d, want 0/1", tx.commits, tx.rollbacks)
	}
}

func TestCbtQuestionWithMutationStoreExecRollsBackOnCommitError(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("commit failed")
	tx := &cbtQuestionMutationFakeTx{commitErr: wantErr}
	starter := &cbtQuestionMutationFakeTxStarter{tx: tx}
	svc := &CbtQuestion{tx: starter}

	err := svc.withMutationStoreExec(ctx, func(store cbtQuestionStore) error {
		return nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if tx.commits != 1 || tx.rollbacks != 1 {
		t.Fatalf("commits/rollbacks = %d/%d, want 1/1", tx.commits, tx.rollbacks)
	}
}
