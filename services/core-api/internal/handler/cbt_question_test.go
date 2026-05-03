package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func cbtQuestionTestTimestamp(hour int) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  time.Date(2026, time.May, 1, hour, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

func requireAnySlice(t *testing.T, value any, key string) []any {
	t.Helper()
	got, ok := value.([]any)
	if !ok {
		t.Fatalf("%s = %#v (%T), want []any", key, value, value)
	}
	return got
}

type fakeCbtQuestionService struct {
	listInput service.ListCbtQuestionsInput
	listRows  []db.ListCbtQuestionsFilteredRow
	listTotal int64
	listErr   error

	getID        pgtype.UUID
	getDetailRow db.GetCbtQuestionDetailRow
	getErr       error

	createInput service.SaveCbtQuestionInput
	createRow   db.CbtQuestion
	createErr   error

	updateInput service.SaveCbtQuestionInput
	updateRow   db.CbtQuestion
	updateErr   error

	deleteID  pgtype.UUID
	deleteErr error

	submitReviewID    pgtype.UUID
	submitReviewUser  string
	submitReviewNotes string
	submitReviewRow   db.CbtQuestion
	submitReviewErr   error

	approveID    pgtype.UUID
	approveUser  string
	approveNotes string
	approveRow   db.CbtQuestion
	approveErr   error

	publishID   pgtype.UUID
	publishUser string
	publishRow  db.CbtQuestion
	publishErr  error

	archiveID   pgtype.UUID
	archiveUser string
	archiveRow  db.CbtQuestion
	archiveErr  error

	duplicateID   pgtype.UUID
	duplicateUser string
	duplicateRow  db.CbtQuestion
	duplicateErr  error

	revisionID    pgtype.UUID
	revisionUser  string
	revisionNotes string
	revisionRow   db.CbtQuestion
	revisionErr   error

	exportInput  service.ListCbtQuestionsInput
	exportResult service.ExportCbtQuestionsCSVResult
	exportErr    error

	templateResult service.ExportCbtQuestionsCSVResult
	templateErr    error
}

func (f *fakeCbtQuestionService) ListFiltered(_ context.Context, in service.ListCbtQuestionsInput) ([]db.ListCbtQuestionsFilteredRow, int64, error) {
	f.listInput = in
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.listRows, f.listTotal, nil
}

func (f *fakeCbtQuestionService) GetDetail(_ context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error) {
	f.getID = id
	if f.getErr != nil {
		return db.GetCbtQuestionDetailRow{}, f.getErr
	}
	return f.getDetailRow, nil
}

func (f *fakeCbtQuestionService) Create(_ context.Context, input service.SaveCbtQuestionInput) (db.CbtQuestion, error) {
	f.createInput = input
	if f.createErr != nil {
		return db.CbtQuestion{}, f.createErr
	}
	return f.createRow, nil
}

func (f *fakeCbtQuestionService) Update(_ context.Context, input service.SaveCbtQuestionInput) (db.CbtQuestion, error) {
	f.updateInput = input
	if f.updateErr != nil {
		return db.CbtQuestion{}, f.updateErr
	}
	return f.updateRow, nil
}

func (f *fakeCbtQuestionService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeCbtQuestionService) SubmitReview(_ context.Context, id pgtype.UUID, username string, reviewNotes string) (db.CbtQuestion, error) {
	f.submitReviewID = id
	f.submitReviewUser = username
	f.submitReviewNotes = reviewNotes
	if f.submitReviewErr != nil {
		return db.CbtQuestion{}, f.submitReviewErr
	}
	return f.submitReviewRow, nil
}

func (f *fakeCbtQuestionService) Approve(_ context.Context, id pgtype.UUID, username string, reviewNotes string) (db.CbtQuestion, error) {
	f.approveID = id
	f.approveUser = username
	f.approveNotes = reviewNotes
	if f.approveErr != nil {
		return db.CbtQuestion{}, f.approveErr
	}
	return f.approveRow, nil
}

func (f *fakeCbtQuestionService) Publish(_ context.Context, id pgtype.UUID, username string) (db.CbtQuestion, error) {
	f.publishID = id
	f.publishUser = username
	if f.publishErr != nil {
		return db.CbtQuestion{}, f.publishErr
	}
	return f.publishRow, nil
}

func (f *fakeCbtQuestionService) Archive(_ context.Context, id pgtype.UUID, username string) (db.CbtQuestion, error) {
	f.archiveID = id
	f.archiveUser = username
	if f.archiveErr != nil {
		return db.CbtQuestion{}, f.archiveErr
	}
	return f.archiveRow, nil
}

func (f *fakeCbtQuestionService) DuplicateAsDraft(_ context.Context, id pgtype.UUID, username string) (db.CbtQuestion, error) {
	f.duplicateID = id
	f.duplicateUser = username
	if f.duplicateErr != nil {
		return db.CbtQuestion{}, f.duplicateErr
	}
	return f.duplicateRow, nil
}

func (f *fakeCbtQuestionService) DuplicateForRevision(_ context.Context, id pgtype.UUID, username string, reviewNotes string) (db.CbtQuestion, error) {
	f.revisionID = id
	f.revisionUser = username
	f.revisionNotes = reviewNotes
	if f.revisionErr != nil {
		return db.CbtQuestion{}, f.revisionErr
	}
	return f.revisionRow, nil
}

func (f *fakeCbtQuestionService) ExportCSV(_ context.Context, input service.ListCbtQuestionsInput) (service.ExportCbtQuestionsCSVResult, error) {
	f.exportInput = input
	if f.exportErr != nil {
		return service.ExportCbtQuestionsCSVResult{}, f.exportErr
	}
	return f.exportResult, nil
}

func (f *fakeCbtQuestionService) TemplateCSV() (service.ExportCbtQuestionsCSVResult, error) {
	if f.templateErr != nil {
		return service.ExportCbtQuestionsCSVResult{}, f.templateErr
	}
	return f.templateResult, nil
}

func cbtQuestionHandlerModel(id, subjectID pgtype.UUID) db.CbtQuestion {
	return db.CbtQuestion{
		ID:             id,
		SubjectID:      subjectID,
		Code:           "Q-HANDLER",
		QuestionText:   "Apa inti materi?",
		QuestionType:   "essay",
		Difficulty:     db.CbtQuestionDifficultyEnumEasy,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		Version:        1,
		AuthorUsername: "guru.ipa",
	}
}

func TestCbtQuestionDecodeAndInputMapping(t *testing.T) {
	body := `{
		"subject_id":"01000000-0000-0000-0000-000000000000",
		"authoring_mode":"advance",
		"code":"Q-001",
		"question_text":"Pilih jawaban",
		"question_type":"multiple_choice",
		"options":[{"label":"A","text":"Satu"}],
		"option_a":"Satu",
		"answer_key":" A ",
		"explanation":"Pembahasan",
		"difficulty":"easy",
		"status":"draft",
		"stem_html":"<p>Stem</p>",
		"stem_latex":"x^2",
		"stimulus_html":"<p>Stimulus</p>",
		"stimulus_latex":"y",
		"explanation_html":"<p>Bahas</p>",
		"rubric_html":"<p>Rubrik</p>",
		"academic_phase":"D",
		"grade_level":8,
		"cp_ref":"CP-1",
		"tp_ref":"TP-1",
		"kd_ref":"KD-1",
		"indicator_ref":"IND-1",
		"material_topic":"Bilangan",
		"cognitive_level":"C2",
		"hots_flag":true,
		"media_asset_ids":["asset-1"],
		"workflow_status":"draft",
		"writer_notes":"catatan penulis",
		"review_notes":"catatan review"
	}`
	req := httptest.NewRequest(http.MethodPost, "/cbt/questions", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"usr": "guru.ipa",
		"sub": "fallback-user",
	}))

	decoded, err := decodeQuestionBody(req)
	if err != nil {
		t.Fatalf("decodeQuestionBody() error = %v", err)
	}
	inputID := handlerTestUUID(9)
	input, err := questionInputFromBody(req, decoded, inputID)
	if err != nil {
		t.Fatalf("questionInputFromBody() error = %v", err)
	}
	if input.ID != inputID || !input.SubjectID.Valid {
		t.Fatalf("questionInputFromBody() ids = %v/%v, want provided id and valid subject", input.ID, input.SubjectID)
	}
	if input.AnswerKey != "A" || input.AuthorUsername != "guru.ipa" || input.ReviewerUsername != "guru.ipa" || input.ApproverUsername != "guru.ipa" {
		t.Fatalf("questionInputFromBody() answer/usernames = %q/%q/%q/%q, want trimmed answer and current user", input.AnswerKey, input.AuthorUsername, input.ReviewerUsername, input.ApproverUsername)
	}
	if !input.GradeLevel.Valid || input.GradeLevel.Int16 != 8 {
		t.Fatalf("questionInputFromBody() grade_level = %v, want 8", input.GradeLevel)
	}
	if input.Difficulty != db.CbtQuestionDifficultyEnumEasy || input.Status != db.CbtQuestionStatusEnumDraft || len(input.Options) != 1 || len(input.MediaAssetIDs) != 1 {
		t.Fatalf("questionInputFromBody() detail = %+v, want mapped difficulty/status/options/media", input)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/cbt/questions", strings.NewReader(`{"subject_id":"bad"}`))
	if _, err := questionInputFromBody(badReq, cbtQuestionBody{SubjectID: "bad"}, pgtype.UUID{}); err == nil {
		t.Fatal("questionInputFromBody(invalid subject) error = nil, want parse error")
	}
	if _, err := decodeQuestionBody(httptest.NewRequest(http.MethodPost, "/cbt/questions", strings.NewReader(`{`))); err == nil {
		t.Fatal("decodeQuestionBody(invalid JSON) error = nil, want decode error")
	}
}

func TestCbtQuestionSerializerHelpers(t *testing.T) {
	listRow := db.ListCbtQuestionsFilteredRow{
		ID:             handlerTestUUID(1),
		SubjectID:      handlerTestUUID(2),
		SubjectName:    "Matematika",
		SubjectCode:    "MTK",
		Code:           "Q-001",
		QuestionText:   "2 + 2 = ?",
		QuestionType:   "multiple_choice",
		Options:        []byte(`[{"label":"A","text":"4"}]`),
		OptionA:        "4",
		AnswerKey:      "A",
		Explanation:    "Empat",
		Difficulty:     db.CbtQuestionDifficultyEnumEasy,
		Status:         db.CbtQuestionStatusEnumDraft,
		CreatedAt:      cbtQuestionTestTimestamp(8),
		UpdatedAt:      cbtQuestionTestTimestamp(9),
		GradeLevel:     pgtype.Int2{Int16: 8, Valid: true},
		MediaAssetIds:  []byte(`["asset-1"]`),
		WorkflowStatus: "draft",
		Version:        2,
		AuthorUsername: "guru.mtk",
	}
	listMap := serializeQuestionListRow(listRow)
	if listMap["subject_name"] != "Matematika" || listMap["subject_code"] != "MTK" || listMap["authoring_mode"] != "beginner" {
		t.Fatalf("serializeQuestionListRow() = %+v, want subject metadata and beginner mode", listMap)
	}
	if listMap["grade_level"] != int16(8) {
		t.Fatalf("serializeQuestionListRow() grade_level = %#v, want int16(8)", listMap["grade_level"])
	}
	if got := requireAnySlice(t, listMap["options"], "options"); len(got) != 1 {
		t.Fatalf("serializeQuestionListRow() options len = %d, want 1", len(got))
	}
	if got := requireAnySlice(t, listMap["media_asset_ids"], "media_asset_ids"); len(got) != 1 || got[0] != "asset-1" {
		t.Fatalf("serializeQuestionListRow() media_asset_ids = %#v, want asset-1", got)
	}

	detailRow := db.GetCbtQuestionDetailRow{
		ID:             handlerTestUUID(3),
		SubjectID:      handlerTestUUID(4),
		SubjectName:    "Bahasa Arab",
		SubjectCode:    "ARB",
		Code:           "Q-ARB",
		QuestionText:   "Terjemahkan",
		QuestionType:   "essay",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		RubricHtml:     "<p>Rubrik</p>",
		WorkflowStatus: "draft",
		Version:        1,
	}
	detailMap := serializeQuestionDetailRow(detailRow)
	if detailMap["subject_name"] != "Bahasa Arab" || detailMap["suggested_mode"] != "advance" || detailMap["rubric_html"] != "<p>Rubrik</p>" {
		t.Fatalf("serializeQuestionDetailRow() = %+v, want detail metadata and advance mode", detailMap)
	}

	model := db.CbtQuestion{
		ID:             handlerTestUUID(5),
		SubjectID:      handlerTestUUID(6),
		Code:           "Q-HOTS",
		QuestionText:   "Analisis",
		QuestionType:   "multiple_choice",
		Options:        []byte(`[{"label":"A","html":"<b>A</b>"}]`),
		Difficulty:     db.CbtQuestionDifficultyEnumHard,
		Status:         db.CbtQuestionStatusEnumPublished,
		HotsFlag:       true,
		MediaAssetIds:  []byte(`["asset-2","asset-3"]`),
		WorkflowStatus: "published",
		Version:        3,
	}
	modelMap := serializeQuestionModel(model)
	if modelMap["authoring_mode"] != "advance" || modelMap["status"] != db.CbtQuestionStatusEnumPublished || modelMap["grade_level"] != nil {
		t.Fatalf("serializeQuestionModel() = %+v, want advance published model with nil grade", modelMap)
	}
	if got := requireAnySlice(t, modelMap["media_asset_ids"], "model media_asset_ids"); len(got) != 2 {
		t.Fatalf("serializeQuestionModel() media_asset_ids len = %d, want 2", len(got))
	}
}

func TestCbtQuestionModeAndJSONHelpers(t *testing.T) {
	if got := decodeJSONBytes(nil); len(requireAnySlice(t, got, "empty json")) != 0 {
		t.Fatalf("decodeJSONBytes(empty) = %#v, want empty slice", got)
	}
	if got := decodeJSONBytes([]byte(`not-json`)); len(requireAnySlice(t, got, "invalid json")) != 0 {
		t.Fatalf("decodeJSONBytes(invalid) = %#v, want empty slice", got)
	}
	if got := decodeJSONBytes([]byte(`{"ok":true}`)); got.(map[string]any)["ok"] != true {
		t.Fatalf("decodeJSONBytes(object) = %#v, want object", got)
	}
	if got := nullableInt(pgtype.Int2{}); got != nil {
		t.Fatalf("nullableInt(invalid) = %#v, want nil", got)
	}
	if got := nullableInt(pgtype.Int2{Int16: 7, Valid: true}); got != int16(7) {
		t.Fatalf("nullableInt(valid) = %#v, want int16(7)", got)
	}

	tests := []struct {
		name string
		args []string
		hots bool
		want string
	}{
		{name: "beginner multiple choice", args: []string{"multiple_choice", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "beginner"},
		{name: "unsupported type", args: []string{"ordering", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "latex content", args: []string{"essay", " x^2 ", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "curriculum metadata", args: []string{"essay", "", "", "D", "", "", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "hots flag", args: []string{"essay", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, hots: true, want: "advance"},
		{name: "workflow review", args: []string{"essay", "", "", "", "", "", "", "", "", "", "review", "", "", ""}, want: "advance"},
		{name: "notes", args: []string{"essay", "", "", "", "", "", "", "", "", "", "draft", "catatan", "", ""}, want: "advance"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serviceAuthoringModeFromRow(
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
				t.Fatalf("serviceAuthoringModeFromRow() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCbtQuestionHandlersForwardSuccessPaths(t *testing.T) {
	subjectID := handlerTestUUID(20)
	questionID := handlerTestUUID(21)
	model := cbtQuestionHandlerModel(questionID, subjectID)

	listFake := &fakeCbtQuestionService{
		listTotal: 1,
		listRows: []db.ListCbtQuestionsFilteredRow{
			{
				ID:             questionID,
				SubjectID:      subjectID,
				SubjectName:    "IPA",
				SubjectCode:    "IPA",
				Code:           "Q-001",
				QuestionText:   "Energi",
				QuestionType:   "essay",
				Difficulty:     db.CbtQuestionDifficultyEnumEasy,
				Status:         db.CbtQuestionStatusEnumDraft,
				WorkflowStatus: "review",
				Version:        1,
			},
		},
	}
	h := &CbtQuestion{svc: listFake}
	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/cbt/questions?subject_id="+subjectID.String()+"&workflow_status=review&question_type=essay&hots=true&revision_source=reviewer&q=energi&limit=50&offset=10", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if listFake.listInput.SubjectID != subjectID || listFake.listInput.WorkflowStatus != "review" || listFake.listInput.QuestionType != "essay" || listFake.listInput.HotsFilter != "true" || listFake.listInput.RevisionSource != "reviewer" || listFake.listInput.SearchQuery != "energi" {
		t.Fatalf("ListFiltered input = %+v, want forwarded filters", listFake.listInput)
	}
	if listFake.listInput.Limit != 50 || listFake.listInput.Offset != 10 || !strings.Contains(rec.Body.String(), "IPA") {
		t.Fatalf("ListFiltered paging/body = %+v/%s, want forwarded paging and item", listFake.listInput, rec.Body.String())
	}

	exportFake := &fakeCbtQuestionService{
		exportResult: service.ExportCbtQuestionsCSVResult{
			Filename: "bank-soal-cbt-test.csv",
			Content:  []byte("kode,tipe\nQ-001,essay\n"),
			Count:    1,
		},
	}
	rec = httptest.NewRecorder()
	(&CbtQuestion{svc: exportFake}).ExportCSV(rec, adminRequest(http.MethodGet, "/api/cbt/questions/export?subject_id="+subjectID.String()+"&workflow_status=review&question_type=essay&q=energi", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ExportCSV() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if exportFake.exportInput.SubjectID != subjectID || exportFake.exportInput.WorkflowStatus != "review" || exportFake.exportInput.QuestionType != "essay" || exportFake.exportInput.SearchQuery != "energi" || exportFake.exportInput.Limit != 2000 {
		t.Fatalf("ExportCSV() input = %+v, want forwarded filters with export limit", exportFake.exportInput)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/csv") {
		t.Fatalf("ExportCSV() content-type = %q, want text/csv", got)
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, "bank-soal-cbt-test.csv") {
		t.Fatalf("ExportCSV() disposition = %q, want filename", got)
	}
	if rec.Body.String() != "kode,tipe\nQ-001,essay\n" {
		t.Fatalf("ExportCSV() body = %q, want CSV content", rec.Body.String())
	}

	templateFake := &fakeCbtQuestionService{
		templateResult: service.ExportCbtQuestionsCSVResult{
			Filename: "template-bank-soal-cbt.csv",
			Content:  []byte("kode,tipe\nTPL-PG-001,pg\n"),
			Count:    1,
		},
	}
	rec = httptest.NewRecorder()
	(&CbtQuestion{svc: templateFake}).TemplateCSV(rec, adminRequest(http.MethodGet, "/api/cbt/questions/template", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("TemplateCSV() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, "template-bank-soal-cbt.csv") {
		t.Fatalf("TemplateCSV() disposition = %q, want template filename", got)
	}
	if rec.Body.String() != "kode,tipe\nTPL-PG-001,pg\n" {
		t.Fatalf("TemplateCSV() body = %q, want CSV content", rec.Body.String())
	}

	getFake := &fakeCbtQuestionService{getDetailRow: db.GetCbtQuestionDetailRow{
		ID:             questionID,
		SubjectID:      subjectID,
		SubjectName:    "IPA",
		SubjectCode:    "IPA",
		Code:           "Q-001",
		QuestionText:   "Energi",
		QuestionType:   "essay",
		Difficulty:     db.CbtQuestionDifficultyEnumEasy,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		Version:        1,
	}}
	rec = httptest.NewRecorder()
	(&CbtQuestion{svc: getFake}).Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Get() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if getFake.getID != questionID {
		t.Fatalf("GetDetail id = %v, want %v", getFake.getID, questionID)
	}

	body := `{"subject_id":"` + subjectID.String() + `","question_text":"Apa inti materi?","question_type":"essay","answer_key":"Energi","difficulty":"easy","status":"draft"}`
	createFake := &fakeCbtQuestionService{createRow: model}
	createAudit := &fakeCbtSessionAuditWriter{}
	rec = httptest.NewRecorder()
	createReq := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions", strings.NewReader(body)), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "uid": "01000000-0000-0000-0000-000000000000", "sub": "01000000-0000-0000-0000-000000000000", "ssid": "sess-1"})
	(&CbtQuestion{svc: createFake, audit: createAudit}).Create(rec, createReq)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if createFake.createInput.SubjectID != subjectID || createFake.createInput.AuthorUsername != "guru.ipa" || createFake.createInput.AnswerKey != "Energi" {
		t.Fatalf("Create input = %+v, want mapped subject/user/answer", createFake.createInput)
	}
	if len(createAudit.entries) != 1 || createAudit.entries[0].Action != "CBT_QUESTION_CREATE" {
		t.Fatalf("Create audit = %#v, want one create audit", createAudit.entries)
	}

	updateFake := &fakeCbtQuestionService{updateRow: model}
	updateAudit := &fakeCbtSessionAuditWriter{}
	rec = httptest.NewRecorder()
	updateReq := withClaims(httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), strings.NewReader(body)), jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin", "uid": "01000000-0000-0000-0000-000000000000", "sub": "01000000-0000-0000-0000-000000000000", "ssid": "sess-1"})
	updateReq = withRouteParam(updateReq, "id", questionID.String())
	(&CbtQuestion{svc: updateFake, audit: updateAudit}).Update(rec, updateReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("Update() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if updateFake.updateInput.ID != questionID || updateFake.updateInput.SubjectID != subjectID {
		t.Fatalf("Update input ids = %v/%v, want route/question subject", updateFake.updateInput.ID, updateFake.updateInput.SubjectID)
	}
	if len(updateAudit.entries) != 1 || updateAudit.entries[0].Action != "CBT_QUESTION_UPDATE" {
		t.Fatalf("Update audit = %#v, want one update audit", updateAudit.entries)
	}

	deleteFake := &fakeCbtQuestionService{}
	deleteAudit := &fakeCbtSessionAuditWriter{}
	rec = httptest.NewRecorder()
	(&CbtQuestion{svc: deleteFake, audit: deleteAudit}).Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if deleteFake.deleteID != questionID {
		t.Fatalf("Delete id = %v, want %v", deleteFake.deleteID, questionID)
	}
	if len(deleteAudit.entries) != 1 || deleteAudit.entries[0].Action != "CBT_QUESTION_DELETE" {
		t.Fatalf("Delete audit = %#v, want one delete audit", deleteAudit.entries)
	}
}

func TestCbtQuestionWorkflowAndDuplicateForwardActor(t *testing.T) {
	questionID := handlerTestUUID(30)
	subjectID := handlerTestUUID(31)
	model := cbtQuestionHandlerModel(questionID, subjectID)

	tests := []struct {
		name   string
		action string
		check  func(t *testing.T, fake *fakeCbtQuestionService)
	}{
		{
			name:   "submit review",
			action: "submit_review",
			check: func(t *testing.T, fake *fakeCbtQuestionService) {
				t.Helper()
				if fake.submitReviewID != questionID || fake.submitReviewUser != "admin.cbt" || fake.submitReviewNotes != "cek" {
					t.Fatalf("SubmitReview args = %v/%q/%q, want id/user/notes", fake.submitReviewID, fake.submitReviewUser, fake.submitReviewNotes)
				}
			},
		},
		{
			name:   "approve",
			action: "approve",
			check: func(t *testing.T, fake *fakeCbtQuestionService) {
				t.Helper()
				if fake.approveID != questionID || fake.approveUser != "admin.cbt" || fake.approveNotes != "cek" {
					t.Fatalf("Approve args = %v/%q/%q, want id/user/notes", fake.approveID, fake.approveUser, fake.approveNotes)
				}
			},
		},
		{
			name:   "publish",
			action: "publish",
			check: func(t *testing.T, fake *fakeCbtQuestionService) {
				t.Helper()
				if fake.publishID != questionID || fake.publishUser != "admin.cbt" {
					t.Fatalf("Publish args = %v/%q, want id/user", fake.publishID, fake.publishUser)
				}
			},
		},
		{
			name:   "archive",
			action: "archive",
			check: func(t *testing.T, fake *fakeCbtQuestionService) {
				t.Helper()
				if fake.archiveID != questionID || fake.archiveUser != "admin.cbt" {
					t.Fatalf("Archive args = %v/%q, want id/user", fake.archiveID, fake.archiveUser)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCbtQuestionService{
				submitReviewRow: model,
				approveRow:      model,
				publishRow:      model,
				archiveRow:      model,
			}
			audit := &fakeCbtSessionAuditWriter{}
			req := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"`+tt.action+`","notes":"cek"}`)), jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin.cbt", "uid": "01000000-0000-0000-0000-000000000000", "sub": "01000000-0000-0000-0000-000000000000", "ssid": "sess-1"})
			req = withRouteParam(req, "id", questionID.String())
			rec := httptest.NewRecorder()
			(&CbtQuestion{svc: fake, audit: audit}).WorkflowAction(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("WorkflowAction(%s) status = %d, want 200; body=%s", tt.action, rec.Code, rec.Body.String())
			}
			tt.check(t, fake)
			if len(audit.entries) != 1 {
				t.Fatalf("WorkflowAction(%s) audit len = %d, want 1", tt.action, len(audit.entries))
			}
		})
	}

	duplicateFake := &fakeCbtQuestionService{duplicateRow: cbtQuestionHandlerModel(handlerTestUUID(32), subjectID)}
	duplicateAudit := &fakeCbtSessionAuditWriter{}
	req := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/duplicate", nil), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "uid": "01000000-0000-0000-0000-000000000000", "sub": "01000000-0000-0000-0000-000000000000", "ssid": "sess-1"})
	req = withRouteParam(req, "id", questionID.String())
	rec := httptest.NewRecorder()
	(&CbtQuestion{svc: duplicateFake, audit: duplicateAudit}).Duplicate(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Duplicate() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if duplicateFake.duplicateID != questionID || duplicateFake.duplicateUser != "guru.ipa" {
		t.Fatalf("Duplicate args = %v/%q, want id/user", duplicateFake.duplicateID, duplicateFake.duplicateUser)
	}
	if len(duplicateAudit.entries) != 1 || duplicateAudit.entries[0].Action != "CBT_QUESTION_DUPLICATE" {
		t.Fatalf("Duplicate audit = %#v, want one duplicate audit", duplicateAudit.entries)
	}

	revisionRow := cbtQuestionHandlerModel(handlerTestUUID(33), subjectID)
	revisionRow.WorkflowStatus = "rejected"
	revisionFake := &fakeCbtQuestionService{revisionRow: revisionRow}
	revisionAudit := &fakeCbtSessionAuditWriter{}
	revisionReq := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/revision", strings.NewReader(`{"notes":"Daya pembeda rendah"}`)), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "uid": "01000000-0000-0000-0000-000000000000", "sub": "01000000-0000-0000-0000-000000000000", "ssid": "sess-1"})
	revisionReq = withRouteParam(revisionReq, "id", questionID.String())
	revisionRec := httptest.NewRecorder()
	(&CbtQuestion{svc: revisionFake, audit: revisionAudit}).MarkRevision(revisionRec, revisionReq)
	if revisionRec.Code != http.StatusCreated {
		t.Fatalf("MarkRevision() status = %d, want 201; body=%s", revisionRec.Code, revisionRec.Body.String())
	}
	if revisionFake.revisionID != questionID || revisionFake.revisionUser != "guru.ipa" || revisionFake.revisionNotes != "Daya pembeda rendah" {
		t.Fatalf("MarkRevision args = %v/%q/%q, want id/user/notes", revisionFake.revisionID, revisionFake.revisionUser, revisionFake.revisionNotes)
	}
	if len(revisionAudit.entries) != 1 || revisionAudit.entries[0].Action != "CBT_QUESTION_MARK_REVISION" {
		t.Fatalf("MarkRevision audit = %#v, want one revision audit", revisionAudit.entries)
	}
}

func TestCbtQuestionTeacherWriteScopeRequiresAuthorOwnership(t *testing.T) {
	subjectID := handlerTestUUID(40)
	questionID := handlerTestUUID(41)
	body := `{"subject_id":"` + subjectID.String() + `","question_text":"Apa inti materi?","question_type":"essay","answer_key":"Energi","difficulty":"easy","status":"draft"}`

	reqClaims := jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "sub": "01000000-0000-0000-0000-000000000010"}

	t.Run("update forbidden for non author", func(t *testing.T) {
		fake := &fakeCbtQuestionService{getDetailRow: db.GetCbtQuestionDetailRow{ID: questionID, AuthorUsername: "guru.lain"}}
		req := withClaims(httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), strings.NewReader(body)), reqClaims)
		req = withRouteParam(req, "id", questionID.String())
		rec := httptest.NewRecorder()
		(&CbtQuestion{svc: fake}).Update(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Update() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.updateInput.ID.Valid {
			t.Fatalf("Update() unexpectedly called service update: %+v", fake.updateInput)
		}
	})

	t.Run("delete forbidden for non author", func(t *testing.T) {
		fake := &fakeCbtQuestionService{getDetailRow: db.GetCbtQuestionDetailRow{ID: questionID, AuthorUsername: "guru.lain"}}
		req := withClaims(httptest.NewRequest(http.MethodDelete, "/api/cbt/questions/"+questionID.String(), nil), reqClaims)
		req = withRouteParam(req, "id", questionID.String())
		rec := httptest.NewRecorder()
		(&CbtQuestion{svc: fake}).Delete(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Delete() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.deleteID.Valid {
			t.Fatalf("Delete() unexpectedly called service delete: %v", fake.deleteID)
		}
	})

	t.Run("submit review forbidden for non author", func(t *testing.T) {
		fake := &fakeCbtQuestionService{getDetailRow: db.GetCbtQuestionDetailRow{ID: questionID, AuthorUsername: "guru.lain"}}
		req := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"submit_review","notes":"cek"}`)), reqClaims)
		req = withRouteParam(req, "id", questionID.String())
		rec := httptest.NewRecorder()
		(&CbtQuestion{svc: fake}).WorkflowAction(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("WorkflowAction() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.submitReviewID.Valid {
			t.Fatalf("WorkflowAction() unexpectedly called submit review: %v", fake.submitReviewID)
		}
	})

	t.Run("update allowed for author", func(t *testing.T) {
		fake := &fakeCbtQuestionService{
			getDetailRow: db.GetCbtQuestionDetailRow{ID: questionID, AuthorUsername: "guru.ipa"},
			updateRow:    cbtQuestionHandlerModel(questionID, subjectID),
		}
		req := withClaims(httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), strings.NewReader(body)), reqClaims)
		req = withRouteParam(req, "id", questionID.String())
		rec := httptest.NewRecorder()
		(&CbtQuestion{svc: fake}).Update(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Update(author) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if fake.updateInput.ID != questionID {
			t.Fatalf("Update(author) id = %v, want %v", fake.updateInput.ID, questionID)
		}
	})

	t.Run("update allowed for author revision draft", func(t *testing.T) {
		fake := &fakeCbtQuestionService{
			getDetailRow: db.GetCbtQuestionDetailRow{
				ID:             questionID,
				AuthorUsername: "guru.ipa",
				WorkflowStatus: "rejected",
				Status:         db.CbtQuestionStatusEnumDraft,
			},
			updateRow: cbtQuestionHandlerModel(questionID, subjectID),
		}
		req := withClaims(httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), strings.NewReader(body)), reqClaims)
		req = withRouteParam(req, "id", questionID.String())
		rec := httptest.NewRecorder()
		(&CbtQuestion{svc: fake}).Update(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Update(revision author) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if fake.updateInput.ID != questionID {
			t.Fatalf("Update(revision author) id = %v, want %v", fake.updateInput.ID, questionID)
		}
	})
}

func TestCbtQuestionDirectUpdateGuardrails(t *testing.T) {
	subjectID := handlerTestUUID(43)
	questionID := handlerTestUUID(44)
	baseBody := `{"subject_id":"` + subjectID.String() + `","question_text":"Apa inti materi?","question_type":"essay","answer_key":"Energi","difficulty":"easy","status":"draft","workflow_status":"draft"}`
	adminClaims := jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin.cbt", "sub": "01000000-0000-0000-0000-000000000011"}
	guruClaims := jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "sub": "01000000-0000-0000-0000-000000000012"}

	tests := []struct {
		name   string
		claims jwt.MapClaims
		body   string
		row    db.GetCbtQuestionDetailRow
	}{
		{
			name:   "locked by package usage",
			claims: adminClaims,
			body:   baseBody,
			row: db.GetCbtQuestionDetailRow{
				ID:             questionID,
				AuthorUsername: "guru.ipa",
				WorkflowStatus: "draft",
				Status:         db.CbtQuestionStatusEnumDraft,
				PackageCount:   1,
			},
		},
		{
			name:   "locked by student answers",
			claims: adminClaims,
			body:   baseBody,
			row: db.GetCbtQuestionDetailRow{
				ID:             questionID,
				AuthorUsername: "guru.ipa",
				WorkflowStatus: "draft",
				Status:         db.CbtQuestionStatusEnumDraft,
				AnswerCount:    1,
			},
		},
		{
			name:   "direct approve rejected",
			claims: adminClaims,
			body:   `{"subject_id":"` + subjectID.String() + `","question_text":"Apa inti materi?","question_type":"essay","answer_key":"Energi","difficulty":"easy","status":"draft","workflow_status":"approved"}`,
			row: db.GetCbtQuestionDetailRow{
				ID:             questionID,
				AuthorUsername: "guru.ipa",
				WorkflowStatus: "review",
				Status:         db.CbtQuestionStatusEnumDraft,
			},
		},
		{
			name:   "direct publish rejected",
			claims: adminClaims,
			body:   `{"subject_id":"` + subjectID.String() + `","question_text":"Apa inti materi?","question_type":"essay","answer_key":"Energi","difficulty":"easy","status":"published","workflow_status":"draft"}`,
			row: db.GetCbtQuestionDetailRow{
				ID:             questionID,
				AuthorUsername: "guru.ipa",
				WorkflowStatus: "draft",
				Status:         db.CbtQuestionStatusEnumDraft,
			},
		},
		{
			name:   "teacher cannot edit review item directly",
			claims: guruClaims,
			body:   baseBody,
			row: db.GetCbtQuestionDetailRow{
				ID:             questionID,
				AuthorUsername: "guru.ipa",
				WorkflowStatus: "review",
				Status:         db.CbtQuestionStatusEnumDraft,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCbtQuestionService{getDetailRow: tt.row}
			req := withClaims(httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), strings.NewReader(tt.body)), tt.claims)
			req = withRouteParam(req, "id", questionID.String())
			rec := httptest.NewRecorder()
			(&CbtQuestion{svc: fake}).Update(rec, req)
			if rec.Code != http.StatusConflict {
				t.Fatalf("Update() status = %d, want 409; body=%s", rec.Code, rec.Body.String())
			}
			if fake.updateInput.ID.Valid {
				t.Fatalf("Update() unexpectedly called service update: %+v", fake.updateInput)
			}
		})
	}
}

func TestCbtQuestionDeleteRejectsLockedQuestion(t *testing.T) {
	questionID := handlerTestUUID(45)
	fake := &fakeCbtQuestionService{
		getDetailRow: db.GetCbtQuestionDetailRow{
			ID:             questionID,
			AuthorUsername: "guru.ipa",
			WorkflowStatus: "draft",
			Status:         db.CbtQuestionStatusEnumDraft,
			PackageCount:   1,
		},
	}
	req := withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String())
	rec := httptest.NewRecorder()
	(&CbtQuestion{svc: fake}).Delete(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Delete() status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID.Valid {
		t.Fatalf("Delete() unexpectedly called service delete: %v", fake.deleteID)
	}
}

func TestCbtQuestionRequireAuthorOrAdminBranches(t *testing.T) {
	questionID := handlerTestUUID(42)
	tests := []struct {
		name       string
		claims     jwt.MapClaims
		svc        *fakeCbtQuestionService
		wantOK     bool
		wantStatus int
	}{
		{
			name:   "admin bypasses author lookup",
			claims: jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin.cbt"},
			svc:    &fakeCbtQuestionService{getErr: errors.New("should not be called")},
			wantOK: true,
		},
		{
			name:       "missing claims forbidden",
			svc:        &fakeCbtQuestionService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non guru forbidden",
			claims:     jwt.MapClaims{"roles": []any{"staf"}, "usr": "tu"},
			svc:        &fakeCbtQuestionService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "guru missing username forbidden",
			claims:     jwt.MapClaims{"roles": []any{"guru"}},
			svc:        &fakeCbtQuestionService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "detail lookup error internal",
			claims:     jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			svc:        &fakeCbtQuestionService{getErr: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "author mismatch forbidden",
			claims:     jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			svc:        &fakeCbtQuestionService{getDetailRow: db.GetCbtQuestionDetailRow{ID: questionID, AuthorUsername: "guru.matematika"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "author allowed",
			claims: jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			svc:    &fakeCbtQuestionService{getDetailRow: db.GetCbtQuestionDetailRow{ID: questionID, AuthorUsername: "guru.ipa"}},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), nil)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			rec := httptest.NewRecorder()
			gotOK := (&CbtQuestion{svc: tt.svc}).requireQuestionAuthorOrAdmin(rec, req, questionID)
			if gotOK != tt.wantOK {
				t.Fatalf("requireQuestionAuthorOrAdmin() = %v, want %v", gotOK, tt.wantOK)
			}
			if !tt.wantOK && rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtQuestionHandlersMapServiceErrors(t *testing.T) {
	subjectID := handlerTestUUID(40)
	questionID := handlerTestUUID(41)
	body := `{"subject_id":"` + subjectID.String() + `","question_text":"Apa inti materi?","question_type":"essay","answer_key":"Energi","difficulty":"easy","status":"draft"}`

	tests := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{
			name: "list internal error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{listErr: errors.New("db down")}}).List,
			req:  adminRequest(http.MethodGet, "/api/cbt/questions", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "list invalid subject",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).List,
			req:  adminRequest(http.MethodGet, "/api/cbt/questions?subject_id=bad", ""),
			want: http.StatusBadRequest,
		},
		{
			name: "get internal error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{getErr: errors.New("db down")}}).Get,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()),
			want: http.StatusInternalServerError,
		},
		{
			name: "get invalid id",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Get,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/bad", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "create client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{createErr: errors.New("question_text wajib diisi")}}).Create,
			req:  withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions", strings.NewReader(body)), jwt.MapClaims{"roles": []any{"admin"}}),
			want: http.StatusBadRequest,
		},
		{
			name: "create invalid json",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Create,
			req:  adminRequest(http.MethodPost, "/api/cbt/questions", `{`),
			want: http.StatusBadRequest,
		},
		{
			name: "create invalid subject",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Create,
			req:  adminRequest(http.MethodPost, "/api/cbt/questions", `{"subject_id":"bad"}`),
			want: http.StatusBadRequest,
		},
		{
			name: "update invalid id",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Update,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/cbt/questions/bad", body), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "update invalid json",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Update,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/cbt/questions/"+questionID.String(), `{`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update invalid subject",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Update,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/cbt/questions/"+questionID.String(), `{"subject_id":"bad"}`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{updateErr: errors.New("answer_key wajib diisi")}}).Update,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/cbt/questions/"+questionID.String(), body), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "delete internal error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{deleteErr: errors.New("db down")}}).Delete,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()),
			want: http.StatusInternalServerError,
		},
		{
			name: "delete invalid id",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Delete,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/questions/bad", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow invalid id",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/bad/workflow", `{"action":"submit_review"}`), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow invalid json",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow unsupported action",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{"action":"skip"}`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow approve forbidden for guru",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowAction,
			req:  withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"approve"}`)), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"}), "id", questionID.String()),
			want: http.StatusForbidden,
		},
		{
			name: "workflow publish forbidden for guru",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowAction,
			req:  withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"publish"}`)), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"}), "id", questionID.String()),
			want: http.StatusForbidden,
		},
		{
			name: "workflow archive forbidden for guru",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowAction,
			req:  withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"archive"}`)), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"}), "id", questionID.String()),
			want: http.StatusForbidden,
		},
		{
			name: "workflow submit review client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{submitReviewErr: errors.New("workflow tidak valid")}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{"action":"submit_review"}`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow approve client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{approveErr: errors.New("workflow tidak valid")}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{"action":"approve"}`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow publish client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{publishErr: errors.New("workflow tidak valid")}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{"action":"publish"}`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow archive client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{archiveErr: errors.New("workflow tidak valid")}}).WorkflowAction,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{"action":"archive"}`), "id", questionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "duplicate client error",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{duplicateErr: errors.New("tidak ditemukan")}}).Duplicate,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/duplicate", ""), "id", questionID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "duplicate invalid id",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Duplicate,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/bad/duplicate", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
