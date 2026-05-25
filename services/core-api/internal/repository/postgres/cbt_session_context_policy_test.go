package db

import (
	"strings"
	"testing"
)

func TestListCbtExamSessionsExposesEventTitleForUiContext(t *testing.T) {
	if !strings.Contains(listCbtExamSessions, "COALESCE(e.title, '') AS event_title") {
		t.Fatalf("ListCbtExamSessions must expose event_title so UI can distinguish Kegiatan vs Mandiri sessions")
	}
	if !strings.Contains(listCbtExamSessions, "LEFT JOIN cbt_exam_events e ON e.id = s.event_id") {
		t.Fatalf("ListCbtExamSessions must join cbt_exam_events for UI context badges")
	}
}
