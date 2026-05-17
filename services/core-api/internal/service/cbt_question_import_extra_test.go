package service

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestNormalizeLegacyAnswerExtra(t *testing.T) {
	cases := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: " a ", want: "A", ok: true},
		{input: "F", want: "F", ok: true},
		{input: "0", want: "A", ok: true},
		{input: "5", want: "F", ok: true},
		{input: "6", ok: false},
		{input: "G", ok: false},
		{input: "", ok: false},
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := normalizeLegacyAnswer(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("normalizeLegacyAnswer(%q) = %q, %v; want %q, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestImportTargetLevelExtra(t *testing.T) {
	cases := []struct {
		name string
		row  map[string]string
		want string
	}{
		{name: "canonical target level is normalized", row: map[string]string{"targetlevel": " viii "}, want: "VIII"},
		{name: "kelas alias is accepted", row: map[string]string{"kelas": "IX"}, want: "IX"},
		{name: "legacy grade without target currently remains empty", row: map[string]string{"gradelevel": "07"}, want: ""},
		{name: "invalid target does not use numeric target as grade", row: map[string]string{"targetlevel": "7"}, want: ""},
		{name: "invalid grade is empty", row: map[string]string{"gradelevel": "VII"}, want: ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := importTargetLevel(tt.row); got != tt.want {
				t.Fatalf("importTargetLevel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestImportLegacyCSVExtraDryRunParsesSuccessAndSkipsInvalidRows(t *testing.T) {
	svc := NewCbtQuestion(nil)
	store := &fakeQuestionStore{
		stemRows: []db.ListCbtQuestionStemTextsBySubjectRow{{QuestionText: "Soal sudah ada"}},
	}
	svc.q = store
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000801")

	csvText := strings.Join([]string{
		"kode;tipe;soal;opsi_a;opsi_b;jawaban;grade_level;kesulitan;hots",
		"Q-1;pg;Soal valid;Salah;Benar;1;07;mudah;ya",
		"Q-2;pg;Soal valid;Salah;Benar;A;07;mudah;tidak",
		"Q-3;pg;Soal sudah ada;Salah;Benar;A;07;mudah;tidak",
		"Q-4;pg;Soal jawaban buruk;Salah;Benar;Z;07;mudah;tidak",
	}, "\n")

	result, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Actor:     CbtQuestionActor{Roles: []string{"admin"}, Username: "admin"},
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV(dry run) error = %v", err)
	}
	if result.TotalRows != 4 || !result.DryRun || result.WouldImport != 1 || result.Imported != 0 || result.Skipped != 3 {
		t.Fatalf("ImportLegacyCSV(dry run) result = %+v, want one valid dry-run row and three skipped rows", result)
	}
	if store.createCalls != 0 {
		t.Fatalf("ImportLegacyCSV(dry run) createCalls = %d, want 0", store.createCalls)
	}
	joinedErrors := strings.Join(result.Errors, "\n")
	for _, want := range []string{"duplikat dalam file import", "duplikat dengan bank soal", "kunci jawaban tidak valid"} {
		if !strings.Contains(joinedErrors, want) {
			t.Fatalf("ImportLegacyCSV(dry run) errors = %q, want substring %q", joinedErrors, want)
		}
	}
}

func TestImportLegacyCSVExtraCreatesParsedQuestion(t *testing.T) {
	svc := NewCbtQuestion(nil)
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8, 2}, Valid: true}},
	}
	svc.q = store
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000802")
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000803")

	csvText := strings.Join([]string{
		"kode,tipe,soal,opsi_a,opsi_b,jawaban,target_level,grade_level,kesulitan,hots,bobot,is_shuffle",
		"Q-10,pg,Soal dibuat,A salah,B benar,1,VIII,08,sulit,true,2,1",
	}, "\n")

	result, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		EventID:   eventID,
		CSVText:   csvText,
		Actor:     CbtQuestionActor{Roles: []string{"admin"}, Username: "admin"},
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV(create) error = %v", err)
	}
	if result.Imported != 1 || result.Skipped != 0 || result.WouldImport != 0 {
		t.Fatalf("ImportLegacyCSV(create) result = %+v, want one imported row", result)
	}
	if store.createCalls != 1 {
		t.Fatalf("ImportLegacyCSV(create) createCalls = %d, want 1", store.createCalls)
	}
	params := store.createParams
	if params.Code != "Q-10" || params.QuestionType != "multiple_choice" || params.AnswerKey != "B" {
		t.Fatalf("created params core = code %q type %q answer %q, want Q-10 multiple_choice B", params.Code, params.QuestionType, params.AnswerKey)
	}
	if params.TargetLevel.String != "VIII" || !params.TargetLevel.Valid || params.Difficulty != db.CbtQuestionDifficultyEnumHard || !params.HotsFlag {
		t.Fatalf("created params metadata = target %+v difficulty %q hots %v, want VIII hard true", params.TargetLevel, params.Difficulty, params.HotsFlag)
	}
	if !strings.Contains(params.WriterNotes, "Bobot legacy: 2") || !strings.Contains(params.WriterNotes, "Shuffle opsi legacy: ya") {
		t.Fatalf("created params WriterNotes = %q, want legacy notes", params.WriterNotes)
	}
}

func TestImportLegacyCSVExtraRejectsMalformedOrTooShortCSV(t *testing.T) {
	svc := NewCbtQuestion(nil)
	svc.q = &fakeQuestionStore{}
	input := ImportLegacyQuestionsInput{
		SubjectID: mustQuestionUUID(t, "00000000-0000-0000-0000-000000000804"),
		Actor:     CbtQuestionActor{Roles: []string{"admin"}, Username: "admin"},
	}

	input.CSVText = "kode,soal\n\"unterminated"
	if _, err := svc.ImportLegacyCSV(context.Background(), input); err == nil || !strings.Contains(err.Error(), "CSV tidak valid") {
		t.Fatalf("ImportLegacyCSV(malformed) error = %v, want CSV tidak valid", err)
	}

	input.CSVText = "kode,soal\n"
	if _, err := svc.ImportLegacyCSV(context.Background(), input); err == nil || !strings.Contains(err.Error(), "minimal berisi header") {
		t.Fatalf("ImportLegacyCSV(short) error = %v, want minimal header error", err)
	}
}
