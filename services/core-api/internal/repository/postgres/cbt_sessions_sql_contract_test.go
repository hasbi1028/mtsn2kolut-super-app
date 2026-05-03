package db

import (
	"strings"
	"testing"
)

func TestUpdateParticipantScoresDoesNotMutateSubmittedAt(t *testing.T) {
	if strings.Contains(strings.ToLower(updateParticipantScores), "submitted_at") {
		t.Fatal("UpdateParticipantScores must not mutate submitted_at; scoring recalculation may run before final submit")
	}
	if !strings.Contains(strings.ToLower(updateParticipantScores), "set score") {
		t.Fatal("UpdateParticipantScores should still update participant score")
	}
}

func TestSubmitParticipantExamKeepsSubmitGuard(t *testing.T) {
	sql := strings.ToLower(submitParticipantExam)
	if !strings.Contains(sql, "submitted_at = coalesce") {
		t.Fatal("SubmitParticipantExam should be the only participant flow that stamps submitted_at")
	}
	if !strings.Contains(sql, "submitted_at is null") {
		t.Fatal("SubmitParticipantExam must keep the duplicate submit guard")
	}
}
