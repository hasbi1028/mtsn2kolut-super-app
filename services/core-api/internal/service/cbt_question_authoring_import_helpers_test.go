package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestQuestionAuthoringPlainTextHelpers(t *testing.T) {
	if !isPlainTextBlockElement("p") || !isPlainTextBlockElement("table") || !isPlainTextBlockElement("li") {
		t.Fatal("isPlainTextBlockElement returned false for known block tags")
	}
	if isPlainTextBlockElement("span") || isPlainTextBlockElement("script") || isPlainTextBlockElement("") {
		t.Fatal("isPlainTextBlockElement returned true for non-block tags")
	}

	html := `<div>Hello<br><span>world</span><script>alert(1)</script><style>.x{}</style><p>Next&nbsp;line</p><svg><text>hidden</text></svg></div>`
	if got, want := derivePlainText(html), "HelloworldNext linehidden"; got != want {
		t.Fatalf("derivePlainText() = %q, want %q", got, want)
	}

	long := ""
	for i := 0; i < 510; i++ {
		long += "a"
	}
	if got := derivePlainText(long); len(got) != 500 {
		t.Fatalf("derivePlainText(long) length = %d, want 500", len(got))
	}
}

func TestQuestionAuthoringNormalizeNoRows(t *testing.T) {
	if got := normalizeNoRows(pgx.ErrNoRows); !errors.Is(got, domain.ErrNotFound) {
		t.Fatalf("normalizeNoRows(pgx.ErrNoRows) = %v, want ErrNotFound", got)
	}
	wrapped := errors.New("wrapped: " + pgx.ErrNoRows.Error())
	if got := normalizeNoRows(wrapped); got != wrapped {
		t.Fatalf("normalizeNoRows(non-wrapped text match) = %v, want original error", got)
	}
	other := errors.New("boom")
	if got := normalizeNoRows(other); got != other {
		t.Fatalf("normalizeNoRows(other) = %v, want original error", got)
	}
}

func TestQuestionAuthoringWorkflowRoleHelpers(t *testing.T) {
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000601")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000602")
	otherSubjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000603")
	otherEventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000604")

	if !memberSubjectMatches(pgtype.UUID{}, subjectID) {
		t.Fatal("memberSubjectMatches should treat empty member subject as wildcard")
	}
	if !memberSubjectMatches(subjectID, subjectID) {
		t.Fatal("memberSubjectMatches should accept the same subject")
	}
	if memberSubjectMatches(otherSubjectID, subjectID) {
		t.Fatal("memberSubjectMatches accepted a different subject")
	}

	svc := NewCbtQuestion(nil)
	store := &fakeQuestionStore{membersByUsername: []db.CbtEventMember{
		{EventID: otherEventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer},
		{EventID: eventID, SubjectID: otherSubjectID, Role: db.CbtEventMemberRoleReviewer},
		{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer},
		{EventID: eventID, SubjectID: pgtype.UUID{}, Role: db.CbtEventMemberRolePanitia},
	}}
	svc.q = store
	current := db.GetCbtQuestionRow{EventID: eventID, SubjectID: subjectID}

	if err := svc.requireWorkflowRole(context.Background(), " reviewer ", current, db.CbtEventMemberRoleReviewer); err != nil {
		t.Fatalf("requireWorkflowRole(reviewer) error = %v", err)
	}
	if err := svc.requireWorkflowRole(context.Background(), "panitia", current, db.CbtEventMemberRoleReviewer); err != nil {
		t.Fatalf("requireWorkflowRole(panitia fallback) error = %v", err)
	}

	svc.q = &fakeQuestionStore{membersByUsername: []db.CbtEventMember{
		{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer},
	}}
	if err := svc.requireWorkflowRole(context.Background(), "reviewer", current, db.CbtEventMemberRolePembuatSoal); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("requireWorkflowRole(wrong role) error = %v, want ErrForbidden", err)
	}
	if err := svc.requireWorkflowRole(context.Background(), "", current, db.CbtEventMemberRoleReviewer); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("requireWorkflowRole(blank username) error = %v, want ErrForbidden", err)
	}
	current.EventID = pgtype.UUID{}
	if err := svc.requireWorkflowRole(context.Background(), "reviewer", current, db.CbtEventMemberRoleReviewer); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("requireWorkflowRole(no event) error = %v, want ErrForbidden", err)
	}
}

func TestQuestionImportPureHelpers(t *testing.T) {
	compactCases := map[string]bool{
		"AB":  true,
		" f ": true,
		"":    false,
		"A G": false,
		"1":   false,
	}
	for input, want := range compactCases {
		if got := importTokenIsCompactLabels(input); got != want {
			t.Fatalf("importTokenIsCompactLabels(%q) = %v, want %v", input, got, want)
		}
	}

	gradeCases := []struct {
		name string
		row  map[string]string
		want string
	}{
		{name: "grade_level", row: map[string]string{"gradelevel": " 7 "}, want: "7"},
		{name: "leading zero", row: map[string]string{"gradelevel": "09"}, want: "9"},
		{name: "invalid", row: map[string]string{"gradelevel": "VII"}, want: ""},
		{name: "missing", row: map[string]string{}, want: ""},
	}
	for _, tt := range gradeCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := importLegacyGradeLevel(tt.row); got != tt.want {
				t.Fatalf("importLegacyGradeLevel() = %q, want %q", got, tt.want)
			}
		})
	}

	labelCases := map[string]string{
		"multiple_choice": "Pilihan Ganda",
		"multiple_answer": "Pilihan Ganda Kompleks",
		"true_false":      "Benar/Salah",
		"agree_disagree":  "Setuju/Tidak Setuju",
		"short_answer":    "Isian Singkat",
		"matching":        "Menjodohkan",
		"essay":           "Essay",
		"custom_type":     "custom_type",
	}
	for input, want := range labelCases {
		if got := importQuestionTypeLabel(input); got != want {
			t.Fatalf("importQuestionTypeLabel(%q) = %q, want %q", input, got, want)
		}
	}
}
