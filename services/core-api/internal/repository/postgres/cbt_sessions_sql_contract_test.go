package db

import (
	"strings"
	"testing"
)

func TestUpdateParticipantScoresDoesNotMutateSubmittedAt(t *testing.T) {
	sql := strings.ToLower(updateParticipantScores)
	if strings.Contains(sql, "set submitted_at") {
		t.Fatal("UpdateParticipantScores must not mutate submitted_at; final submit owns that timestamp")
	}
	if !strings.Contains(sql, "ep.submitted_at is not null") {
		t.Fatal("UpdateParticipantScores should only score submitted participants")
	}
	if !strings.Contains(sql, "set score") {
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
