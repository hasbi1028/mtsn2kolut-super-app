package service

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// documentCycleTestUUID creates a UUID with the given seed byte for testing.
func documentCycleTestUUID(seed byte) pgtype.UUID {
	var buf [16]byte
	buf[0] = seed
	return pgtype.UUID{Bytes: buf, Valid: true}
}
