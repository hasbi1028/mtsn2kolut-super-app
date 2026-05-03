package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeQuestionStore struct {
	current db.GetCbtQuestionRow
	detail  db.GetCbtQuestionDetailRow

	listRows      []db.ListCbtQuestionsRow
	filteredRows  []db.ListCbtQuestionsFilteredRow
	stemRows      []db.ListCbtQuestionStemTextsBySubjectRow
	listFilterArg db.ListCbtQuestionsFilteredParams
	countArg      db.CountCbtQuestionsFilteredParams
	count         int64

	createParams  db.CreateCbtQuestionParams
	createCalls   int
	createRow     db.CbtQuestion
	createHistory []db.CreateCbtQuestionParams

	updateParams db.UpdateCbtQuestionParams
	updateCalls  int

	deleteID    pgtype.UUID
	deleteCalls int
}

func (f *fakeQuestionStore) ListCbtQuestions(ctx context.Context) ([]db.ListCbtQuestionsRow, error) {
	return f.listRows, nil
}

func (f *fakeQuestionStore) ListCbtQuestionsFiltered(ctx context.Context, arg db.ListCbtQuestionsFilteredParams) ([]db.ListCbtQuestionsFilteredRow, error) {
	f.listFilterArg = arg
	return f.filteredRows, nil
}

func (f *fakeQuestionStore) CountCbtQuestionsFiltered(ctx context.Context, arg db.CountCbtQuestionsFilteredParams) (int64, error) {
	f.countArg = arg
	return f.count, nil
}

func (f *fakeQuestionStore) ListCbtQuestionStemTextsBySubject(ctx context.Context, subjectID pgtype.UUID) ([]db.ListCbtQuestionStemTextsBySubjectRow, error) {
	return f.stemRows, nil
}

func (f *fakeQuestionStore) GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if !f.current.ID.Valid {
		return db.GetCbtQuestionRow{}, errors.New("not found")
	}
	return f.current, nil
}

func (f *fakeQuestionStore) GetCbtQuestionDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error) {
	return f.detail, nil
}

func (f *fakeQuestionStore) CreateCbtQuestion(ctx context.Context, arg db.CreateCbtQuestionParams) (db.CbtQuestion, error) {
	f.createParams = arg
	f.createCalls++
	f.createHistory = append(f.createHistory, arg)
	return f.createRow, nil
}

func (f *fakeQuestionStore) UpdateCbtQuestion(ctx context.Context, arg db.UpdateCbtQuestionParams) (db.CbtQuestion, error) {
	f.updateParams = arg
	f.updateCalls++
	return db.CbtQuestion{
		ID:             arg.ID,
		QuestionText:   arg.QuestionText,
		WorkflowStatus: arg.WorkflowStatus,
		Status:         arg.Status,
		StemHtml:       arg.StemHtml,
		StimulusHtml:   arg.StimulusHtml,
	}, nil
}

func (f *fakeQuestionStore) DeleteCbtQuestion(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	f.deleteCalls++
	return nil
}

func TestNewCbtQuestionAndReadDelegation(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	store := &fakeQuestionStore{
		listRows: []db.ListCbtQuestionsRow{{ID: questionID, Code: "Q-1"}},
		current:  db.GetCbtQuestionRow{ID: questionID, Code: "Q-1", QuestionText: "Soal"},
		detail:   db.GetCbtQuestionDetailRow{ID: questionID, Code: "Q-1", QuestionText: "Soal detail"},
	}
	svc := NewCbtQuestion(nil)
	svc.q = store

	rows, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Code != "Q-1" {
		t.Fatalf("List() = %+v, want one Q-1 row", rows)
	}

	got, err := svc.Get(context.Background(), questionID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != questionID || got.QuestionText != "Soal" {
		t.Fatalf("Get() = %+v, want current row", got)
	}

	detail, err := svc.GetDetail(context.Background(), questionID)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if detail.ID != questionID || detail.QuestionText != "Soal detail" {
		t.Fatalf("GetDetail() = %+v, want detail row", detail)
	}
}

func TestCbtQuestionFilterCreateAndDeleteDelegation(t *testing.T) {
	store := &fakeQuestionStore{
		filteredRows: []db.ListCbtQuestionsFilteredRow{{ID: pgtype.UUID{Valid: true}, Code: "Q-1"}},
		count:        7,
		createRow: db.CbtQuestion{
			ID:           pgtype.UUID{Valid: true},
			QuestionText: "Soal mudah",
		},
	}
	svc := &CbtQuestion{q: store}

	rows, total, err := svc.ListFiltered(context.Background(), ListCbtQuestionsInput{
		SubjectID:      pgtype.UUID{Valid: true},
		WorkflowStatus: " draft ",
		QuestionType:   " multiple_choice ",
		HotsFilter:     " true ",
		RevisionSource: " item_analysis ",
		SearchQuery:    " aljabar ",
		Limit:          25,
		Offset:         5,
	})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(rows) != 1 || total != 7 {
		t.Fatalf("ListFiltered() rows/total = %d/%d, want 1/7", len(rows), total)
	}
	if store.listFilterArg.WorkflowStatus != "draft" || store.listFilterArg.QuestionType != "multiple_choice" || store.listFilterArg.HotsFilter != "true" || store.listFilterArg.RevisionSource != "item_analysis" || store.listFilterArg.SearchQuery != "aljabar" {
		t.Fatalf("ListFiltered() arg = %+v, want trimmed filters", store.listFilterArg)
	}
	if store.countArg.WorkflowStatus != store.listFilterArg.WorkflowStatus || store.countArg.RevisionSource != store.listFilterArg.RevisionSource || store.countArg.SearchQuery != store.listFilterArg.SearchQuery {
		t.Fatalf("ListFiltered() count arg = %+v, want same trimmed filters", store.countArg)
	}

	created, err := svc.Create(context.Background(), SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "single_choice",
		QuestionText:   " Soal mudah ",
		OptionA:        " A ",
		OptionB:        " B ",
		OptionC:        " C ",
		OptionD:        " D ",
		AnswerKey:      " b ",
		MediaAssetIDs:  []string{"asset-1"},
		AuthorUsername: " guru ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !created.ID.Valid || store.createCalls != 1 {
		t.Fatalf("Create() row/calls = %+v/%d, want store row and one call", created, store.createCalls)
	}
	if store.createParams.QuestionType != "multiple_choice" || store.createParams.QuestionText != "Soal mudah" || store.createParams.AnswerKey != "B" {
		t.Fatalf("Create() params = %+v, want normalized type/text/answer", store.createParams)
	}
	if store.createParams.OptionA != "A" || store.createParams.OptionD != "D" || store.createParams.Version != 1 || store.createParams.AuthorUsername != "guru" {
		t.Fatalf("Create() params = %+v, want legacy options/version/author", store.createParams)
	}
	if string(store.createParams.MediaAssetIds) != `["asset-1"]` {
		t.Fatalf("Create() media_asset_ids = %s, want asset JSON", string(store.createParams.MediaAssetIds))
	}

	deleteID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	if err := svc.Delete(context.Background(), deleteID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleteID != deleteID || store.deleteCalls != 1 {
		t.Fatalf("Delete() id/calls = %v/%d, want %v/1", store.deleteID, store.deleteCalls, deleteID)
	}
}

func TestCbtQuestionImportLegacyCSVMapsRows(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		"kode;soal;opsi_a;opsi_b;opsi_c;opsi_d;jawaban;gambar_soal;gambar_b;bobot;is_rtl",
		"MTK-1;<p>Berapa 2+2?</p>;3;4;5;6;1;https://cdn.test/soal.png;https://cdn.test/b.png;2;true",
		"MTK-1;Berapa 3+3?;5;6;7;8;B;;;;",
		"; ;A;B;C;D;A;;;;",
		"MTK-2;Pilih huruf;A;B;C;D;C;;;;",
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Username:  " guru.cbt ",
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV() error = %v", err)
	}
	if got.TotalRows != 4 || got.Imported != 2 || got.Skipped != 2 {
		t.Fatalf("ImportLegacyCSV() result = %+v, want 4 total, 2 imported, 2 skipped", got)
	}
	if len(got.Errors) != 2 || !strings.Contains(strings.Join(got.Errors, "\n"), "kode MTK-1 duplikat") || !strings.Contains(strings.Join(got.Errors, "\n"), "soal kosong") {
		t.Fatalf("ImportLegacyCSV() errors = %+v, want duplicate code and empty question errors", got.Errors)
	}
	if len(got.DuplicateCodes) != 1 || got.DuplicateCodes[0] != "MTK-1" {
		t.Fatalf("ImportLegacyCSV() duplicate codes = %+v, want MTK-1", got.DuplicateCodes)
	}
	if store.createCalls != 2 || len(store.createHistory) != 2 {
		t.Fatalf("CreateCbtQuestion() calls/history = %d/%d, want 2/2", store.createCalls, len(store.createHistory))
	}
	first := store.createHistory[0]
	if first.SubjectID != subjectID || first.Code != "MTK-1" || first.AnswerKey != "B" || first.AuthorUsername != "guru.cbt" {
		t.Fatalf("first imported params = %+v, want subject/code/answer/author mapped", first)
	}
	if !strings.Contains(first.StemHtml, "https://cdn.test/soal.png") || !strings.Contains(string(first.Options), "https://cdn.test/b.png") {
		t.Fatalf("first imported media = stem %q options %s, want legacy image URLs", first.StemHtml, string(first.Options))
	}
	if !strings.Contains(first.WriterNotes, "Bobot legacy: 2") || !strings.Contains(first.WriterNotes, "RTL legacy: ya") {
		t.Fatalf("first imported notes = %q, want legacy metadata notes", first.WriterNotes)
	}
	if store.createHistory[1].Code != "MTK-2" || store.createHistory[1].AnswerKey != "C" {
		t.Fatalf("second imported params = %+v, want MTK-2 with answer C", store.createHistory[1])
	}
}

func TestCbtQuestionImportLegacyCSVSkipsExistingStem(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		stemRows: []db.ListCbtQuestionStemTextsBySubjectRow{
			{QuestionText: "Berapa 2+2?", StemHtml: "<p>Berapa 2+2?</p>"},
		},
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		"kode;soal;opsi_a;opsi_b;opsi_c;opsi_d;jawaban",
		"MTK-1;<p>Berapa 2+2?</p>;3;4;5;6;B",
		"MTK-2;Berapa 3+3?;5;6;7;8;B",
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Username:  "guru.cbt",
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV() error = %v", err)
	}
	if got.Imported != 1 || got.Skipped != 1 || store.createCalls != 1 {
		t.Fatalf("ImportLegacyCSV() result/calls = %+v/%d, want 1 imported, 1 skipped, 1 create", got, store.createCalls)
	}
	if !strings.Contains(strings.Join(got.Errors, "\n"), "duplikat dengan bank soal") {
		t.Fatalf("ImportLegacyCSV() errors = %+v, want existing duplicate message", got.Errors)
	}
	if store.createHistory[0].Code != "MTK-2" {
		t.Fatalf("created code = %q, want only non-duplicate row", store.createHistory[0].Code)
	}
}

func TestCbtQuestionImportLegacyCSVSupportsStructuredTypes(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		importCSVRow("kode", "tipe", "soal", "opsi_a", "opsi_b", "opsi_c", "opsi_d", "opsi_e", "jawaban", "kiri_a", "kanan_1", "kiri_b", "kanan_2", "distraktor_1", "rubrik"),
		importCSVRow("MTK-G", "pg_kompleks", "Pilih bilangan genap", "1", "2", "3", "4", "6", "B E", "", "", "", "", "", ""),
		importCSVRow("MTK-BS", "benar_salah", "Air membeku pada suhu rendah", "", "", "", "", "", "Benar", "", "", "", "", "", ""),
		importCSVRow("MTK-IS", "isian", "Ibu kota Sulawesi Tenggara", "", "", "", "", "", "Kendari | kendari kota", "", "", "", "", "", ""),
		importCSVRow("MTK-ES", "essay", "Jelaskan fotosintesis", "", "", "", "", "", "", "", "", "", "", "", "<p>Ketepatan konsep</p>"),
		importCSVRow("MTK-M", "menjodohkan", "Pasangkan angka", "", "", "", "", "", "", "Satu", "1", "Dua", "2", "Tiga", ""),
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Username:  "guru.cbt",
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV() error = %v", err)
	}
	if got.TotalRows != 5 || got.Imported != 5 || got.Skipped != 0 {
		t.Fatalf("ImportLegacyCSV() result = %+v, want 5 imported and no skipped rows", got)
	}
	if store.createCalls != 5 || len(store.createHistory) != 5 {
		t.Fatalf("CreateCbtQuestion() calls/history = %d/%d, want 5/5", store.createCalls, len(store.createHistory))
	}
	if store.createHistory[0].QuestionType != "multiple_answer" || store.createHistory[0].AnswerKey != "B,E" || !strings.Contains(string(store.createHistory[0].Options), `"label":"E"`) {
		t.Fatalf("multiple answer import = type %q answer %q options %s, want type/options/key mapped", store.createHistory[0].QuestionType, store.createHistory[0].AnswerKey, string(store.createHistory[0].Options))
	}
	if store.createHistory[1].QuestionType != "true_false" || store.createHistory[1].AnswerKey != "A" || !strings.Contains(string(store.createHistory[1].Options), "Benar") {
		t.Fatalf("true/false import = %+v options %s, want fixed options and answer A", store.createHistory[1], string(store.createHistory[1].Options))
	}
	if store.createHistory[2].QuestionType != "short_answer" || store.createHistory[2].AnswerKey != "Kendari|kendari kota" {
		t.Fatalf("short answer import = %+v, want aliases preserved", store.createHistory[2])
	}
	if store.createHistory[3].QuestionType != "essay" || store.createHistory[3].RubricHtml != "<p>Ketepatan konsep</p>" {
		t.Fatalf("essay import = %+v, want rubric mapped", store.createHistory[3])
	}
	if store.createHistory[4].QuestionType != "matching" || store.createHistory[4].AnswerKey != "A=1;B=2" || !strings.Contains(string(store.createHistory[4].Options), `"is_distractor":true`) {
		t.Fatalf("matching import = %+v options %s, want default answer and distractor", store.createHistory[4], string(store.createHistory[4].Options))
	}
}

func importCSVRow(values ...string) string {
	return strings.Join(values, ";")
}

func TestCbtQuestionExportCSVMapsStructuredTypes(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	multipleAnswerOptions, err := EncodeQuestionOptions([]QuestionOption{
		{Label: "A", Text: "Satu"},
		{Label: "B", Text: "Dua"},
		{Label: "C", Text: "Tiga"},
		{Label: "D", Text: "Empat"},
		{Label: "E", Text: "Enam"},
	})
	if err != nil {
		t.Fatalf("EncodeQuestionOptions(multiple answer) error = %v", err)
	}
	matchingOptions, err := EncodeQuestionOptions([]QuestionOption{
		{Label: "A", Text: "Satu", MatchLabel: "1", MatchText: "1"},
		{Label: "B", Text: "Dua", MatchLabel: "2", MatchText: "2"},
		{MatchLabel: "3", MatchText: "Tiga", IsDistractor: true},
	})
	if err != nil {
		t.Fatalf("EncodeQuestionOptions(matching) error = %v", err)
	}
	store := &fakeQuestionStore{
		filteredRows: []db.ListCbtQuestionsFilteredRow{
			{
				SubjectID:      subjectID,
				SubjectName:    "Matematika",
				Code:           "MTK-G",
				QuestionText:   "Pilih bilangan genap",
				QuestionType:   "multiple_answer",
				Options:        multipleAnswerOptions,
				AnswerKey:      "B,E",
				Difficulty:     db.CbtQuestionDifficultyEnumMedium,
				Status:         db.CbtQuestionStatusEnumDraft,
				WorkflowStatus: "draft",
				GradeLevel:     pgtype.Int2{Int16: 8, Valid: true},
				HotsFlag:       true,
			},
			{
				SubjectID:      subjectID,
				SubjectName:    "Matematika",
				Code:           "MTK-M",
				QuestionText:   "Pasangkan angka",
				QuestionType:   "matching",
				Options:        matchingOptions,
				AnswerKey:      "A=1;B=2",
				Difficulty:     db.CbtQuestionDifficultyEnumMedium,
				Status:         db.CbtQuestionStatusEnumDraft,
				WorkflowStatus: "draft",
			},
		},
		count: 2,
	}
	svc := &CbtQuestion{q: store}

	got, err := svc.ExportCSV(context.Background(), ListCbtQuestionsInput{
		SubjectID:      subjectID,
		WorkflowStatus: " draft ",
		Limit:          0,
	})
	if err != nil {
		t.Fatalf("ExportCSV() error = %v", err)
	}
	if got.Count != 2 || !strings.HasPrefix(got.Filename, "bank-soal-cbt-") {
		t.Fatalf("ExportCSV() result = %+v, want count and generated filename", got)
	}
	if store.listFilterArg.LimitCount != 2000 || store.listFilterArg.WorkflowStatus != "draft" {
		t.Fatalf("ExportCSV() list arg = %+v, want default export limit and trimmed workflow", store.listFilterArg)
	}
	records, err := csv.NewReader(strings.NewReader(string(got.Content))).ReadAll()
	if err != nil {
		t.Fatalf("ExportCSV() CSV parse error = %v content=%s", err, string(got.Content))
	}
	if len(records) != 3 {
		t.Fatalf("ExportCSV() records = %d, want header + 2 rows", len(records))
	}
	header := csvHeaderIndex(records[0])
	if records[1][header["tipe"]] != "pg_kompleks" || records[1][header["opsi_e"]] != "Enam" || records[1][header["jawaban"]] != "B,E" || records[1][header["grade_level"]] != "8" || records[1][header["hots_flag"]] != "true" {
		t.Fatalf("multiple answer CSV row = %+v, want type/options/key/metadata mapped", records[1])
	}
	if records[2][header["tipe"]] != "menjodohkan" || records[2][header["kiri_a"]] != "Satu" || records[2][header["kanan_2"]] != "2" || records[2][header["distraktor_1"]] != "Tiga" {
		t.Fatalf("matching CSV row = %+v, want pairs and distractor mapped", records[2])
	}

	importStore := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{9}, Valid: true}},
	}
	importSvc := &CbtQuestion{q: importStore}
	imported, err := importSvc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   string(got.Content),
		Username:  "guru.import",
	})
	if err != nil {
		t.Fatalf("roundtrip ImportLegacyCSV() error = %v", err)
	}
	if imported.Imported != 2 || imported.Skipped != 0 {
		t.Fatalf("roundtrip import result = %+v, want 2 imported and no skipped rows", imported)
	}
	if importStore.createHistory[0].QuestionType != "multiple_answer" || importStore.createHistory[0].AnswerKey != "B,E" || importStore.createHistory[0].GradeLevel.Int16 != 8 || !importStore.createHistory[0].HotsFlag {
		t.Fatalf("roundtrip multiple answer params = %+v, want type/key/metadata preserved", importStore.createHistory[0])
	}
	if importStore.createHistory[1].QuestionType != "matching" || importStore.createHistory[1].AnswerKey != "A=1;B=2" || !strings.Contains(string(importStore.createHistory[1].Options), `"is_distractor":true`) {
		t.Fatalf("roundtrip matching params = %+v options %s, want matching pair and distractor preserved", importStore.createHistory[1], string(importStore.createHistory[1].Options))
	}
}

func TestCbtQuestionTemplateCSVRoundtripsThroughImport(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	svc := &CbtQuestion{q: &fakeQuestionStore{}}
	template, err := svc.TemplateCSV()
	if err != nil {
		t.Fatalf("TemplateCSV() error = %v", err)
	}
	if template.Count < 6 || template.Filename != "template-bank-soal-cbt.csv" {
		t.Fatalf("TemplateCSV() result = %+v, want sample rows and stable filename", template)
	}
	records, err := csv.NewReader(strings.NewReader(string(template.Content))).ReadAll()
	if err != nil {
		t.Fatalf("TemplateCSV() CSV parse error = %v content=%s", err, string(template.Content))
	}
	header := csvHeaderIndex(records[0])
	if header["tipe"] == 0 || header["rubrik"] == 0 || header["distraktor_1"] == 0 {
		t.Fatalf("TemplateCSV() header = %+v, want multi-type columns", records[0])
	}

	importStore := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{10}, Valid: true}},
	}
	importSvc := &CbtQuestion{q: importStore}
	imported, err := importSvc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   string(template.Content),
		Username:  "guru.template",
	})
	if err != nil {
		t.Fatalf("template ImportLegacyCSV() error = %v", err)
	}
	if imported.Imported != template.Count || imported.Skipped != 0 {
		t.Fatalf("template import result = %+v, want all template rows imported", imported)
	}
	types := make(map[string]bool, len(importStore.createHistory))
	for _, params := range importStore.createHistory {
		types[params.QuestionType] = true
	}
	for _, questionType := range []string{"multiple_choice", "multiple_answer", "true_false", "short_answer", "matching", "essay"} {
		if !types[questionType] {
			t.Fatalf("template import types = %+v, want %s", types, questionType)
		}
	}
}

func csvHeaderIndex(headers []string) map[string]int {
	out := make(map[string]int, len(headers))
	for idx, header := range headers {
		out[header] = idx
	}
	return out
}

func TestNormalizeQuestionInputSanitizesDangerousHTML(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		StemHTML:       `<p onclick="alert(1)">Halo</p><script>alert(2)</script>`,
		StimulusHTML:   `<img src="javascript:alert(1)" onerror="alert(1)">`,
		QuestionText:   "",
		Options:        []QuestionOption{{Label: "A", HTML: `<span onclick="x()">Aman</span>`}, {Label: "B", Text: "B"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.StemHTML != `<p>Halo</p>` {
		t.Fatalf("StemHTML = %q, want sanitized paragraph", got.StemHTML)
	}
	if got.StimulusHTML != `<img>` {
		t.Fatalf("StimulusHTML = %q, want sanitized img", got.StimulusHTML)
	}
	if got.Options[0].HTML != `<span>Aman</span>` {
		t.Fatalf("Option HTML = %q, want sanitized span", got.Options[0].HTML)
	}
}

func TestValidateQuestionRequiresApprovedBeforePublish(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal uji",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumPublished,
		WorkflowStatus: "draft",
	}
	_, err := normalizeQuestionInput(input)
	if err == nil {
		t.Fatal("normalizeQuestionInput() error = nil, want publish gating error")
	}
}

func TestSubmitReviewUpdatesWorkflowAndReviewer(t *testing.T) {
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{
			ID:             pgtype.UUID{Valid: true},
			SubjectID:      pgtype.UUID{Valid: true},
			QuestionText:   "Soal",
			QuestionType:   "multiple_choice",
			Options:        []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"}]`),
			AnswerKey:      "A",
			Difficulty:     db.CbtQuestionDifficultyEnumMedium,
			Status:         db.CbtQuestionStatusEnumDraft,
			WorkflowStatus: "draft",
		},
	}
	svc := &CbtQuestion{q: store}

	_, err := svc.SubmitReview(context.Background(), store.current.ID, "reviewer1", "cek redaksi")
	if err != nil {
		t.Fatalf("SubmitReview() error = %v", err)
	}
	if store.updateParams.WorkflowStatus != "review" {
		t.Fatalf("WorkflowStatus = %q, want review", store.updateParams.WorkflowStatus)
	}
	if store.updateParams.ReviewerUsername != "reviewer1" {
		t.Fatalf("ReviewerUsername = %q, want reviewer1", store.updateParams.ReviewerUsername)
	}
}

func TestCbtQuestionWorkflowActions(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	current := db.GetCbtQuestionRow{
		ID:             questionID,
		SubjectID:      subjectID,
		Code:           "Q-1",
		QuestionText:   "Soal",
		QuestionType:   "multiple_choice",
		Options:        []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"},{"label":"C","text":"C"},{"label":"D","text":"D"}]`),
		OptionA:        "A",
		OptionB:        "B",
		OptionC:        "C",
		OptionD:        "D",
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "approved",
		ReviewNotes:    "catatan lama",
	}

	t.Run("approve updates workflow and notes", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.Approve(context.Background(), questionID, "waka", "siap")
		if err != nil {
			t.Fatalf("Approve() error = %v", err)
		}
		if store.updateParams.WorkflowStatus != "approved" || store.updateParams.ReviewerUsername != "waka" {
			t.Fatalf("Approve() params = %+v, want approved reviewer waka", store.updateParams)
		}
		if !store.updateParams.ReviewedAt.Valid || store.updateParams.ReviewNotes != "siap" {
			t.Fatalf("Approve() review timestamp/notes = %v/%q, want valid/siap", store.updateParams.ReviewedAt, store.updateParams.ReviewNotes)
		}
	})

	t.Run("publish sets published status and approver", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.Publish(context.Background(), questionID, "kepala")
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
		if store.updateParams.Status != db.CbtQuestionStatusEnumPublished || store.updateParams.WorkflowStatus != "approved" {
			t.Fatalf("Publish() params = %+v, want published approved", store.updateParams)
		}
		if store.updateParams.ApproverUsername != "kepala" || !store.updateParams.ApprovedAt.Valid {
			t.Fatalf("Publish() approver = %q/%v, want kepala with timestamp", store.updateParams.ApproverUsername, store.updateParams.ApprovedAt)
		}
	})

	t.Run("archive sets archived status", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.Archive(context.Background(), questionID, "admin")
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if store.updateParams.Status != db.CbtQuestionStatusEnumArchived {
			t.Fatalf("Archive() status = %q, want archived", store.updateParams.Status)
		}
	})

	t.Run("duplicate creates clean draft copy", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.DuplicateAsDraft(context.Background(), questionID, "guru")
		if err != nil {
			t.Fatalf("DuplicateAsDraft() error = %v", err)
		}
		if store.createCalls != 1 {
			t.Fatalf("CreateCbtQuestion() calls = %d, want 1", store.createCalls)
		}
		if store.createParams.Code != "Q-1-COPY" || store.createParams.Status != db.CbtQuestionStatusEnumDraft || store.createParams.WorkflowStatus != "draft" {
			t.Fatalf("DuplicateAsDraft() create params = %+v, want clean draft copy", store.createParams)
		}
		if store.createParams.ReviewerUsername != "" || store.createParams.ApproverUsername != "" || store.createParams.ReviewNotes != "" {
			t.Fatalf("DuplicateAsDraft() reviewer/approver/notes = %q/%q/%q, want cleared", store.createParams.ReviewerUsername, store.createParams.ApproverUsername, store.createParams.ReviewNotes)
		}
	})

	t.Run("duplicate for revision creates rejected draft copy with notes", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.DuplicateForRevision(context.Background(), questionID, "reviewer", "Daya pembeda rendah")
		if err != nil {
			t.Fatalf("DuplicateForRevision() error = %v", err)
		}
		if store.createCalls != 1 {
			t.Fatalf("CreateCbtQuestion() calls = %d, want 1", store.createCalls)
		}
		if !strings.HasPrefix(store.createParams.Code, "Q-1-REV-") || store.createParams.Status != db.CbtQuestionStatusEnumDraft || store.createParams.WorkflowStatus != "rejected" {
			t.Fatalf("DuplicateForRevision() create params = %+v, want rejected draft revision copy", store.createParams)
		}
		if store.createParams.ReviewerUsername != "reviewer" || store.createParams.ApproverUsername != "" || store.createParams.ReviewNotes != "Daya pembeda rendah" {
			t.Fatalf("DuplicateForRevision() reviewer/approver/notes = %q/%q/%q, want reviewer/no approver/notes", store.createParams.ReviewerUsername, store.createParams.ApproverUsername, store.createParams.ReviewNotes)
		}
	})
}

func TestNormalizeQuestionInputBeginnerDefaultsToDraft(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal mudah",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		Status:         db.CbtQuestionStatusEnumPublished,
		WorkflowStatus: "approved",
		Difficulty:     db.CbtQuestionDifficultyEnumHard,
		WriterNotes:    "should be cleared",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.Status != db.CbtQuestionStatusEnumDraft {
		t.Fatalf("Status = %q, want draft", got.Status)
	}
	if got.WorkflowStatus != "draft" {
		t.Fatalf("WorkflowStatus = %q, want draft", got.WorkflowStatus)
	}
	if got.Difficulty != db.CbtQuestionDifficultyEnumMedium {
		t.Fatalf("Difficulty = %q, want medium", got.Difficulty)
	}
	if got.WriterNotes != "" {
		t.Fatalf("WriterNotes = %q, want cleared", got.WriterNotes)
	}
}

func TestNormalizeQuestionInputBeginnerSupportsJuknisTypes(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "short_answer",
		QuestionText:   "Jawab singkat",
		AnswerKey:      " Fotosintesis | fotosintesis | foto sintesis ",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.QuestionType != "short_answer" || got.AnswerKey != "Fotosintesis|foto sintesis" {
		t.Fatalf("normalizeQuestionInput() type/key = %q/%q, want short_answer normalized aliases", got.QuestionType, got.AnswerKey)
	}
}

func TestNormalizeShortAnswerKey(t *testing.T) {
	got := normalizeShortAnswerKey("  Iman kepada Allah | iman   kepada allah | Iman\u00a0Kepada\u00a0Rasul  | ")
	want := "Iman kepada Allah|Iman Kepada Rasul"
	if got != want {
		t.Fatalf("normalizeShortAnswerKey() = %q, want %q", got, want)
	}
	if got := normalizeShortAnswerComparable("  Iman\u00a0  Kepada   Allah "); got != "iman kepada allah" {
		t.Fatalf("normalizeShortAnswerComparable() = %q, want normalized lowercase whitespace", got)
	}
}

func TestCbtQuestionBuildUpdatePublishedParams(t *testing.T) {
	currentReviewedAt := pgtype.Timestamptz{Valid: true}
	currentApprovedAt := pgtype.Timestamptz{Valid: true}
	current := db.GetCbtQuestionRow{
		ID:               pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		ReviewerUsername: "old-reviewer",
		ReviewedAt:       currentReviewedAt,
		ApproverUsername: "old-approver",
		ApprovedAt:       currentApprovedAt,
	}

	params, err := buildUpdateQuestionParams(current, SaveCbtQuestionInput{
		ID:               current.ID,
		SubjectID:        pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		AuthoringMode:    "advance",
		QuestionType:     "essay",
		QuestionText:     "Uraikan proses fotosintesis",
		Status:           db.CbtQuestionStatusEnumPublished,
		WorkflowStatus:   "approved",
		RubricHTML:       "<p>Rubrik lengkap</p>",
		ReviewerUsername: "reviewer-baru",
		ApproverUsername: "approver-baru",
		ReviewNotes:      " siap ",
	})
	if err != nil {
		t.Fatalf("buildUpdateQuestionParams() error = %v", err)
	}
	if params.ID != current.ID || params.QuestionType != "essay" || params.Status != db.CbtQuestionStatusEnumPublished || params.WorkflowStatus != "approved" {
		t.Fatalf("buildUpdateQuestionParams() identity/status = %+v, want published approved essay", params)
	}
	if params.ReviewerUsername != "reviewer-baru" || !params.ReviewedAt.Valid {
		t.Fatalf("buildUpdateQuestionParams() reviewer = %q/%v, want reviewer-baru with timestamp", params.ReviewerUsername, params.ReviewedAt)
	}
	if params.ApproverUsername != "approver-baru" || !params.ApprovedAt.Valid {
		t.Fatalf("buildUpdateQuestionParams() approver = %q/%v, want approver-baru with timestamp", params.ApproverUsername, params.ApprovedAt)
	}
	if params.ReviewNotes != "siap" || params.RubricHtml != "<p>Rubrik lengkap</p>" {
		t.Fatalf("buildUpdateQuestionParams() notes/rubric = %q/%q, want trimmed sanitized fields", params.ReviewNotes, params.RubricHtml)
	}
}

func TestNormalizeQuestionInputValidationMatrix(t *testing.T) {
	tests := []struct {
		name    string
		input   SaveCbtQuestionInput
		wantErr string
	}{
		{
			name:    "missing question text",
			input:   SaveCbtQuestionInput{SubjectID: pgtype.UUID{Valid: true}, AuthoringMode: "advance", QuestionType: "essay"},
			wantErr: "question_text atau stem_html/stem_latex wajib diisi",
		},
		{
			name: "objective requires two options",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "multiple_answer",
				QuestionText:   "Pilih",
				Options:        []QuestionOption{{Label: "A", Text: "A"}},
				AnswerKey:      "A",
				WorkflowStatus: "review",
			},
			wantErr: "opsi jawaban minimal 2 untuk tipe soal objektif",
		},
		{
			name: "multiple answer requires two keys",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "multiple_answer",
				QuestionText:   "Pilih semua",
				Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
				AnswerKey:      "A",
				WorkflowStatus: "review",
			},
			wantErr: "multiple_answer membutuhkan minimal 2 kunci jawaban",
		},
		{
			name: "beginner multiple choice requires four options",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "beginner",
				QuestionType:   "multiple_choice",
				QuestionText:   "Pilih",
				Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
				AnswerKey:      "A",
				WorkflowStatus: "review",
			},
			wantErr: "mode beginner membutuhkan minimal 4 opsi untuk pilihan ganda",
		},
		{
			name: "short answer requires key",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "short_answer",
				QuestionText:   "Jawab",
				WorkflowStatus: "review",
			},
			wantErr: "answer_key wajib diisi untuk short_answer",
		},
		{
			name: "objective answer key must match option label",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "multiple_choice",
				QuestionText:   "Pilih",
				Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
				AnswerKey:      "F",
				WorkflowStatus: "review",
			},
			wantErr: "answer_key harus sesuai label opsi yang tersedia",
		},
		{
			name: "approved essay requires rubric",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "essay",
				QuestionText:   "Uraikan",
				WorkflowStatus: "approved",
			},
			wantErr: "rubric_html wajib diisi untuk essay yang diajukan review, di-approve, atau dipublish",
		},
		{
			name: "matching requires complete pair content",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "beginner",
				QuestionType:   "matching",
				QuestionText:   "Cocokkan",
				Options:        []QuestionOption{{Label: "A", Text: "Istilah", MatchLabel: "1"}},
				AnswerKey:      "A=1",
				WorkflowStatus: "review",
			},
			wantErr: "setiap pasangan menjodohkan wajib memiliki kolom kiri dan kanan",
		},
		{
			name: "unsupported type",
			input: SaveCbtQuestionInput{
				SubjectID:     pgtype.UUID{Valid: true},
				AuthoringMode: "advance",
				QuestionType:  "ordering",
				QuestionText:  "Cocokkan",
			},
			wantErr: "question_type tidak didukung",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeQuestionInput(tt.input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("normalizeQuestionInput() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestCbtQuestionNormalizeAndEncodingHelpers(t *testing.T) {
	draft, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "multiple_choice",
		QuestionText:  "Draft awal",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(draft partial) error = %v", err)
	}
	if draft.WorkflowStatus != "draft" || len(draft.Options) != 0 {
		t.Fatalf("normalizeQuestionInput(draft partial) workflow/options = %q/%d, want draft/0", draft.WorkflowStatus, len(draft.Options))
	}

	trueFalse, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "advance",
		QuestionType:  "true_false",
		QuestionText:  "MTsN 2 berada di Kolaka Utara",
		AnswerKey:     "A",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(true_false) error = %v", err)
	}
	if len(trueFalse.Options) != 2 || trueFalse.Options[0].Text != "Benar" || trueFalse.Options[1].Text != "Salah" {
		t.Fatalf("normalizeQuestionInput(true_false) options = %+v, want Benar/Salah", trueFalse.Options)
	}

	agreeDisagree, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "agree_disagree",
		QuestionText:  "Kebersihan kelas adalah tanggung jawab bersama",
		AnswerKey:     "b",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(agree_disagree) error = %v", err)
	}
	if len(agreeDisagree.Options) != 2 || agreeDisagree.Options[0].Text != "Setuju" || agreeDisagree.Options[1].Text != "Tidak Setuju" || agreeDisagree.AnswerKey != "B" {
		t.Fatalf("normalizeQuestionInput(agree_disagree) options/key = %+v/%q, want Setuju/Tidak Setuju with key B", agreeDisagree.Options, agreeDisagree.AnswerKey)
	}

	matching, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "matching",
		QuestionText:  "Jodohkan istilah dengan pengertiannya",
		Options: []QuestionOption{
			{Label: "A", HTML: "<p>Fotosintesis</p>", MatchLabel: "1", MatchHTML: "<p>Proses membuat makanan</p>"},
			{Label: "B", Text: "Evaporasi", MatchLabel: "2", MatchText: "Penguapan"},
			{MatchLabel: "3", MatchText: "Distraktor kanan", IsDistractor: true},
		},
		AnswerKey:      "b=2; a=1",
		WorkflowStatus: "review",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(matching) error = %v", err)
	}
	if matching.AnswerKey != "A=1;B=2" || len(matching.Options) != 3 || matching.Options[0].MatchText != "" || matching.Options[0].MatchHTML != "<p>Proses membuat makanan</p>" || !matching.Options[2].IsDistractor {
		t.Fatalf("normalizeQuestionInput(matching) key/options = %q/%+v, want canonical matching options", matching.AnswerKey, matching.Options)
	}

	fromStem, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "advance",
		QuestionType:  "essay",
		StemHTML:      "<p>Jelaskan&nbsp;EDM</p>",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(stem) error = %v", err)
	}
	if fromStem.QuestionText != "Jelaskan EDM" {
		t.Fatalf("normalizeQuestionInput(stem) QuestionText = %q, want derived plain text", fromStem.QuestionText)
	}

	if got := normalizeQuestionType(" SINGLE_CHOICE "); got != "multiple_choice" {
		t.Fatalf("normalizeQuestionType(single_choice) = %q, want multiple_choice", got)
	}
	if got := normalizeQuestionType(" TRUE_FALSE "); got != "true_false" {
		t.Fatalf("normalizeQuestionType(true_false) = %q, want true_false", got)
	}
	if got := normalizeQuestionType(" AGREE_DISAGREE "); got != "agree_disagree" {
		t.Fatalf("normalizeQuestionType(agree_disagree) = %q, want agree_disagree", got)
	}
	if got := normalizeQuestionType("matching"); got != "matching" {
		t.Fatalf("normalizeQuestionType(matching) = %q, want matching", got)
	}
	if got := normalizeQuestionType("ordering"); got != "ordering" {
		t.Fatalf("normalizeQuestionType(unknown) = %q, want ordering passthrough", got)
	}
	if got := normalizeAuthoringMode("ADVANCE"); got != "advance" {
		t.Fatalf("normalizeAuthoringMode() = %q, want advance", got)
	}
	if got := normalizeWorkflowStatus("approved"); got != "approved" {
		t.Fatalf("normalizeWorkflowStatus(approved) = %q, want approved", got)
	}
	if got := normalizeWorkflowStatus("published"); got != "draft" {
		t.Fatalf("normalizeWorkflowStatus(invalid) = %q, want draft", got)
	}

	sixOptions, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "multiple_choice",
		QuestionText:  "Pilih jawaban",
		Options: []QuestionOption{
			{Label: "A", Text: "A"},
			{Label: "B", Text: "B"},
			{Label: "C", Text: "C"},
			{Label: "D", Text: "D"},
			{Label: "E", Text: "E"},
			{Label: "F", Text: "F"},
		},
		AnswerKey: "f",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(six options) error = %v", err)
	}
	if len(sixOptions.Options) != 6 || sixOptions.AnswerKey != "F" {
		t.Fatalf("normalizeQuestionInput(six options) = %d/%q, want 6/F", len(sixOptions.Options), sixOptions.AnswerKey)
	}

	a, b, c, d, e := legacyOptionColumns([]QuestionOption{
		{Text: "teks"},
		{HTML: "<b>html</b>"},
		{Latex: "x^2"},
	})
	if a != "teks" || b != "html" || c != "x^2" || d != "" || e != "" {
		t.Fatalf("legacyOptionColumns() = %q/%q/%q/%q/%q, want text/html/latex/empty/empty", a, b, c, d, e)
	}

	optionsJSON, err := EncodeQuestionOptions([]QuestionOption{{Label: "A", Text: "Satu"}})
	if err != nil {
		t.Fatalf("EncodeQuestionOptions() error = %v", err)
	}
	var decodedOptions []QuestionOption
	if err := json.Unmarshal(optionsJSON, &decodedOptions); err != nil || len(decodedOptions) != 1 || decodedOptions[0].Label != "A" {
		t.Fatalf("EncodeQuestionOptions() json = %s decoded=%+v err=%v, want one option", string(optionsJSON), decodedOptions, err)
	}
	emptyOptionsJSON, err := EncodeQuestionOptions(nil)
	if string(mustBytes(t, emptyOptionsJSON, err)) != "[]" {
		t.Fatalf("EncodeQuestionOptions(nil) = %s, want []", string(emptyOptionsJSON))
	}
	assetJSON, err := EncodeStringArray([]string{"asset-1"})
	if string(mustBytes(t, assetJSON, err)) != `["asset-1"]` {
		t.Fatalf("EncodeStringArray() = %s, want asset JSON", string(assetJSON))
	}
	if got := decodeQuestionOptions([]byte(`[{"label":"B","text":"Dua"}]`)); len(got) != 1 || got[0].Label != "B" {
		t.Fatalf("decodeQuestionOptions(valid) = %+v, want one B option", got)
	}
	if got := decodeQuestionOptions([]byte(`bad`)); got != nil {
		t.Fatalf("decodeQuestionOptions(invalid) = %+v, want nil", got)
	}
	if got := decodeStringArray([]byte(`["a","b"]`)); strings.Join(got, ",") != "a,b" {
		t.Fatalf("decodeStringArray(valid) = %+v, want a,b", got)
	}
	if got := decodeStringArray([]byte(`bad`)); got != nil {
		t.Fatalf("decodeStringArray(invalid) = %+v, want nil", got)
	}
	if got := mergeNotes("lama", " baru "); got != "baru" {
		t.Fatalf("mergeNotes() = %q, want incoming note only", got)
	}
	if got := mergeNotes("lama", " "); got != "lama" {
		t.Fatalf("mergeNotes(empty incoming) = %q, want existing", got)
	}
}

func TestCbtQuestionSuggestAuthoringMode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		hots bool
		want string
	}{
		{name: "beginner", args: []string{"multiple_choice", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "beginner"},
		{name: "type beginner", args: []string{"short_answer", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "beginner"},
		{name: "latex advance", args: []string{"essay", "", " y ", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "metadata advance", args: []string{"essay", "", "", "", "", "TP", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "hots advance", args: []string{"essay", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, hots: true, want: "advance"},
		{name: "workflow advance", args: []string{"essay", "", "", "", "", "", "", "", "", "", "review", "", "", ""}, want: "advance"},
		{name: "notes advance", args: []string{"essay", "", "", "", "", "", "", "", "", "", "draft", "", "review", ""}, want: "advance"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestQuestionAuthoringMode(
				tt.args[0],
				tt.args[1],
				tt.args[2],
				tt.args[3],
				tt.args[4],
				tt.args[5],
				tt.args[6],
				tt.args[7],
				tt.args[8],
				tt.args[9],
				tt.hots,
				tt.args[10],
				tt.args[11],
				tt.args[12],
				tt.args[13],
			)
			if got != tt.want {
				t.Fatalf("suggestQuestionAuthoringMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func mustBytes(t *testing.T, value []byte, err error) []byte {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	return value
}
