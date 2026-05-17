package logging

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestRequestIDContextRoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req_test")
	if got := RequestID(ctx); got != "req_test" {
		t.Fatalf("RequestID() = %q", got)
	}
	ctx = WithRequestID(context.Background(), "bad request id with spaces and a very long invalid suffix")
	if got := RequestID(ctx); got == "bad request id with spaces and a very long invalid suffix" || got == "" {
		t.Fatalf("invalid RequestID was not regenerated safely: %q", got)
	}
}

func TestRedactAttrs(t *testing.T) {
	attrs := RedactAttrs(
		slog.String("authorization", "Bearer abc"),
		slog.String("password", "secret"),
		slog.String("subject_id", "safe"),
	)
	values := map[string]string{}
	for _, attr := range attrs {
		values[attr.Key] = attr.Value.String()
	}
	if values["authorization"] != "[REDACTED]" || values["password"] != "[REDACTED]" {
		t.Fatalf("sensitive attrs not redacted: %#v", values)
	}
	if values["subject_id"] != "safe" {
		t.Fatalf("safe attr unexpectedly changed: %#v", values)
	}
}

func TestPgErrorAttrs(t *testing.T) {
	err := &pgconn.PgError{Code: "23503", ConstraintName: "cbt_questions_event_id_fkey", TableName: "cbt_questions", ColumnName: "event_id", Message: "violates foreign key constraint"}
	attrs := PgErrorAttrs(err)
	values := map[string]string{}
	for _, attr := range attrs {
		values[attr.Key] = attr.Value.String()
	}
	if values["sqlstate"] != "23503" || values["constraint"] != "cbt_questions_event_id_fkey" || values["table"] != "cbt_questions" {
		t.Fatalf("unexpected pg attrs: %#v", values)
	}
	if got := PgErrorAttrs(errors.New("plain")); len(got) != 0 {
		t.Fatalf("plain error attrs len = %d", len(got))
	}
}
