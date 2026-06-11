package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func mustQuestionUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("scan uuid %q: %v", value, err)
	}
	return id
}
