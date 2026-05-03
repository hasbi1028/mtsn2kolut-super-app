package handler

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestSerializeNonTestAssessmentRowsIncludeSyncFreshness(t *testing.T) {
	lastReviewedAt := pgtype.Timestamptz{
		Time:  time.Date(2026, 5, 3, 10, 30, 0, 0, time.UTC),
		Valid: true,
	}

	listItem := serializeNonTestAssessmentListRow(db.ListNonTestAssessmentsRow{
		ID:                          handlerTestUUID(31),
		SubjectID:                   handlerTestUUID(32),
		LastReviewedAt:              lastReviewedAt,
		UnsyncedReviewedSubmissions: 2,
	})
	if got, ok := listItem["last_reviewed_at"].(pgtype.Timestamptz); !ok || !got.Valid || !got.Time.Equal(lastReviewedAt.Time) {
		t.Fatalf("list last_reviewed_at = %#v, want %v", listItem["last_reviewed_at"], lastReviewedAt.Time)
	}
	if got := listItem["unsynced_reviewed_submissions"]; got != int32(2) {
		t.Fatalf("list unsynced_reviewed_submissions = %#v, want 2", got)
	}

	detailItem := serializeNonTestAssessmentDetailRow(db.GetNonTestAssessmentRow{
		ID:                          handlerTestUUID(33),
		SubjectID:                   handlerTestUUID(34),
		LastReviewedAt:              lastReviewedAt,
		UnsyncedReviewedSubmissions: 3,
	})
	if got, ok := detailItem["last_reviewed_at"].(pgtype.Timestamptz); !ok || !got.Valid || !got.Time.Equal(lastReviewedAt.Time) {
		t.Fatalf("detail last_reviewed_at = %#v, want %v", detailItem["last_reviewed_at"], lastReviewedAt.Time)
	}
	if got := detailItem["unsynced_reviewed_submissions"]; got != int32(3) {
		t.Fatalf("detail unsynced_reviewed_submissions = %#v, want 3", got)
	}
}
