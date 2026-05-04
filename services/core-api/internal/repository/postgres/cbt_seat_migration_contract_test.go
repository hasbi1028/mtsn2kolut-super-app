package db

import (
	"os"
	"strings"
	"testing"
)

func TestCBTSeatInvariantMigrationContract(t *testing.T) {
	content, err := os.ReadFile("../../../db/migrations/061_cbt_participant_seat_invariants.sql")
	if err != nil {
		t.Fatalf("read seat invariant migration: %v", err)
	}
	sql := strings.ToLower(string(content))

	required := []string{
		"seat_no <= 0",
		"raise exception 'invalid cbt participant seat numbers exist",
		"group by room_id, seat_no",
		"having count(*) > 1",
		"raise exception 'duplicate cbt participant seats exist",
		"check (seat_no is null or seat_no > 0)",
		"create unique index if not exists idx_cbt_participants_room_seat_unique",
		"where room_id is not null",
		"and seat_no is not null",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("seat invariant migration missing %q", needle)
		}
	}
}
