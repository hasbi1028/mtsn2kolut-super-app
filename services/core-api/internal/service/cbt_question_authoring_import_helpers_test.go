package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestQuestionAuthoringPlainTextDocumentsCurrentHTMLFlattening(t *testing.T) {
	input := `<article><h1> Judul </h1><p>Paragraf<br><span>lanjutan</span></p><script>alert(1)</script><style>.x{}</style><svg><text>ikon</text></svg><ul><li>Satu</li><li>Dua</li></ul></article>`
	got := derivePlainText(input)
	if strings.Contains(got, "alert") || strings.Contains(got, ".x") {
		t.Fatalf("derivePlainText() = %q, want script/style text skipped", got)
	}
	if got != "Judul ParagraflanjutanikonSatuDua" {
		t.Fatalf("derivePlainText() = %q, want current flattened HTML text", got)
	}

	if got := plainTextFromHTML(""); got != "" {
		t.Fatalf("plainTextFromHTML(blank) = %q, want empty", got)
	}
}

func TestCbtQuestionFromCurrentCopiesAuthoringAndVersionFields(t *testing.T) {
	id := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000701")
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000702")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000703")
	groupID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000704")
	sourceID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000705")
	supersedesID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000706")
	options, err := json.Marshal([]QuestionOption{{Label: "A", Text: "Benar"}})
	if err != nil {
		t.Fatalf("Marshal options error = %v", err)
	}
	createdAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), Valid: true}
	updatedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC), Valid: true}
	reviewedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC), Valid: true}
	approvedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC), Valid: true}

	current := db.GetCbtQuestionRow{
		ID:                   id,
		EventID:              eventID,
		SubjectID:            subjectID,
		Code:                 "Q-701",
		QuestionText:         "Teks soal",
		QuestionType:         "multiple_choice",
		Options:              options,
		OptionA:              "A",
		OptionB:              "B",
		AnswerKey:            "A",
		Explanation:          "Karena A",
		Difficulty:           db.CbtQuestionDifficultyEnumMedium,
		Status:               db.CbtQuestionStatusEnumPublished,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
		StemHtml:             "<p>Stem</p>",
		StemLatex:            "x^2",
		StimulusHtml:         "<p>Stimulus</p>",
		StimulusLatex:        "y^2",
		ExplanationHtml:      "<p>Expl</p>",
		RubricHtml:           "<p>Rubric</p>",
		AcademicPhase:        "D",
		TargetLevel:          pgtype.Text{String: "8", Valid: true},
		CpRef:                "CP",
		TpRef:                "TP",
		KdRef:                "KD",
		IndicatorRef:         "IND",
		MaterialTopic:        "Peluang",
		CognitiveLevel:       "C4",
		HotsFlag:             true,
		MediaAssetIds:        []byte(`["asset-1"]`),
		WorkflowStatus:       "published",
		Version:              3,
		VersionGroupID:       groupID,
		VersionNumber:        4,
		SourceQuestionID:     sourceID,
		SupersedesQuestionID: supersedesID,
		IsLatestVersion:      true,
		VersionNote:          "catatan versi",
		AuthorUsername:       "guru.ipa",
		ReviewerUsername:     "reviewer",
		ReviewedAt:           reviewedAt,
		ApproverUsername:     "approver",
		ApprovedAt:           approvedAt,
		WriterNotes:          "catatan penulis",
		ReviewNotes:          "catatan review",
	}

	got := cbtQuestionFromCurrent(current)
	if got.ID != id || got.EventID != eventID || got.SubjectID != subjectID || got.VersionGroupID != groupID || got.SourceQuestionID != sourceID || got.SupersedesQuestionID != supersedesID {
		t.Fatalf("cbtQuestionFromCurrent() ids = %+v, want copied IDs", got)
	}
	if got.Code != current.Code || got.QuestionText != current.QuestionText || got.QuestionType != current.QuestionType || got.AnswerKey != current.AnswerKey || got.WorkflowStatus != current.WorkflowStatus {
		t.Fatalf("cbtQuestionFromCurrent() core fields = %+v, want copied core fields", got)
	}
	if string(got.Options) != string(options) || got.StemHtml != current.StemHtml || got.RubricHtml != current.RubricHtml || string(got.MediaAssetIds) != string(current.MediaAssetIds) {
		t.Fatalf("cbtQuestionFromCurrent() rich fields = %+v, want copied rich fields", got)
	}
	if got.Version != 3 || got.VersionNumber != 4 || !got.IsLatestVersion || got.VersionNote != current.VersionNote || got.AuthorUsername != current.AuthorUsername || got.ReviewedAt != reviewedAt || got.ApprovedAt != approvedAt {
		t.Fatalf("cbtQuestionFromCurrent() workflow/version fields = %+v, want copied workflow/version fields", got)
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
