package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeGradeService struct {
	*service.Grade

	overviewAssignmentID  pgtype.UUID
	overviewComponentID   pgtype.UUID
	overviewPublishedOnly bool
	overviewTeacherID     pgtype.UUID
	overviewResult        service.GradeOverview
	overviewErr           error
	reportSettingsRow     service.ReportSettings
	reportSettingsErr     error
	updateSettingsArg     db.UpsertReportSettingsParams
	updateSettingsRow     service.ReportSettings
	updateSettingsErr     error
	descriptionArg        db.UpsertGradeStudentSubjectDescriptionParams
	descriptionTeacherID  pgtype.UUID
	descriptionRow        db.GradeStudentSubjectDescription
	descriptionErr        error
	createArg             db.CreateGradeComponentParams
	createTeacherID       pgtype.UUID
	createRow             db.GradeComponent
	createErr             error
	updateArg             db.UpdateGradeComponentParams
	updateTeacherID       pgtype.UUID
	updateRow             db.GradeComponent
	updateErr             error
	publishID             pgtype.UUID
	publishValue          bool
	publishTeacherID      pgtype.UUID
	publishRow            db.GradeComponent
	publishErr            error
	finalizeAssignmentID  pgtype.UUID
	finalizeBy            string
	finalizeNotes         string
	finalizeTeacherID     pgtype.UUID
	finalizeRow           db.GradeAssignmentFinalization
	finalizeErr           error
	reopenAssignmentID    pgtype.UUID
	reopenTeacherID       pgtype.UUID
	reopenErr             error
	deleteID              pgtype.UUID
	deleteTeacherID       pgtype.UUID
	deleteErr             error
	entryComponentID      pgtype.UUID
	entryStudentID        pgtype.UUID
	entryTeacherID        pgtype.UUID
	entryScore            float64
	entryNotes            string
	entryGradedBy         string
	entryRow              db.GradeEntry
	entryErr              error
}

func (f *fakeGradeService) Overview(_ context.Context, assignmentID, componentID pgtype.UUID, publishedOnly bool, teacherEmployeeID pgtype.UUID) (service.GradeOverview, error) {
	f.overviewAssignmentID = assignmentID
	f.overviewComponentID = componentID
	f.overviewPublishedOnly = publishedOnly
	f.overviewTeacherID = teacherEmployeeID
	return f.overviewResult, f.overviewErr
}

func (f *fakeGradeService) GetReportSettings(_ context.Context) (service.ReportSettings, error) {
	return f.reportSettingsRow, f.reportSettingsErr
}

func (f *fakeGradeService) UpdateReportSettings(_ context.Context, arg db.UpsertReportSettingsParams) (service.ReportSettings, error) {
	f.updateSettingsArg = arg
	return f.updateSettingsRow, f.updateSettingsErr
}

func (f *fakeGradeService) UpsertStudentSubjectDescription(_ context.Context, arg db.UpsertGradeStudentSubjectDescriptionParams, teacherEmployeeID pgtype.UUID) (db.GradeStudentSubjectDescription, error) {
	f.descriptionArg = arg
	f.descriptionTeacherID = teacherEmployeeID
	return f.descriptionRow, f.descriptionErr
}

func (f *fakeGradeService) CreateComponent(_ context.Context, arg db.CreateGradeComponentParams, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error) {
	f.createArg = arg
	f.createTeacherID = teacherEmployeeID
	return f.createRow, f.createErr
}

func (f *fakeGradeService) UpdateComponent(_ context.Context, arg db.UpdateGradeComponentParams, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error) {
	f.updateArg = arg
	f.updateTeacherID = teacherEmployeeID
	return f.updateRow, f.updateErr
}

func (f *fakeGradeService) SetComponentPublished(_ context.Context, id pgtype.UUID, isPublished bool, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error) {
	f.publishID = id
	f.publishValue = isPublished
	f.publishTeacherID = teacherEmployeeID
	return f.publishRow, f.publishErr
}

func (f *fakeGradeService) FinalizeAssignment(_ context.Context, assignmentID pgtype.UUID, finalizedBy, notes string, teacherEmployeeID pgtype.UUID) (db.GradeAssignmentFinalization, error) {
	f.finalizeAssignmentID = assignmentID
	f.finalizeBy = finalizedBy
	f.finalizeNotes = notes
	f.finalizeTeacherID = teacherEmployeeID
	return f.finalizeRow, f.finalizeErr
}

func (f *fakeGradeService) ReopenAssignment(_ context.Context, assignmentID, teacherEmployeeID pgtype.UUID) error {
	f.reopenAssignmentID = assignmentID
	f.reopenTeacherID = teacherEmployeeID
	return f.reopenErr
}

func (f *fakeGradeService) DeleteComponent(_ context.Context, id, teacherEmployeeID pgtype.UUID) error {
	f.deleteID = id
	f.deleteTeacherID = teacherEmployeeID
	return f.deleteErr
}

func (f *fakeGradeService) UpsertEntry(_ context.Context, componentID, studentID, teacherEmployeeID pgtype.UUID, score float64, notes, gradedBy string) (db.GradeEntry, error) {
	f.entryComponentID = componentID
	f.entryStudentID = studentID
	f.entryTeacherID = teacherEmployeeID
	f.entryScore = score
	f.entryNotes = notes
	f.entryGradedBy = gradedBy
	return f.entryRow, f.entryErr
}

func gradeGuruRequest(method, target, body string, teacherID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"guru"},
		"role":  "guru",
		"eid":   teacherID.String(),
		"usr":   "guru.ipa",
		"uid":   "01000000-0000-0000-0000-000000000000",
		"sub":   "01000000-0000-0000-0000-000000000000",
		"ssid":  "sess-grade-1",
	})
}

func gradeComponent(id, assignmentID pgtype.UUID, title string) db.GradeComponent {
	return db.GradeComponent{ID: id, AssignmentID: assignmentID, Title: title, Category: "daily", Weight: 1, MaxScore: 100, IsPublished: true}
}

func TestGradeReportSettingsAndDescriptionHandlers(t *testing.T) {
	academicYearID := handlerTestUUID(144)
	assignmentID := handlerTestUUID(145)
	studentID := handlerTestUUID(146)
	teacherID := handlerTestUUID(147)
	fake := &fakeGradeService{
		Grade:             &service.Grade{},
		reportSettingsRow: service.ReportSettings{AcademicYearID: academicYearID.String(), AcademicYearName: "2026/2027", ShowRankingOnReport: true, RankingMethod: "average", RankingTiePolicy: "same_rank", Notes: "Aktif"},
		updateSettingsRow: service.ReportSettings{AcademicYearID: academicYearID.String(), ShowRankingOnReport: false, RankingMethod: "weighted", RankingTiePolicy: "dense", Notes: "Revisi"},
		descriptionRow:    db.GradeStudentSubjectDescription{ID: handlerTestUUID(148), AssignmentID: assignmentID, StudentID: studentID, Description: "Sangat baik"},
	}
	h := &Grade{svc: fake}

	rec := httptest.NewRecorder()
	h.GetReportSettings(rec, gradeGuruRequest(http.MethodGet, "/api/grades/report-settings", "", teacherID))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetReportSettings status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "2026/2027") {
		t.Fatalf("GetReportSettings body = %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateReportSettings(rec, gradeGuruRequest(http.MethodPut, "/api/grades/report-settings", `{"academic_year_id":"`+academicYearID.String()+`","show_ranking_on_report":false,"ranking_method":"weighted","ranking_tie_policy":"dense","notes":"Revisi"}`, teacherID))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateReportSettings status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateSettingsArg.AcademicYearID != academicYearID || fake.updateSettingsArg.ShowRankingOnReport || fake.updateSettingsArg.RankingMethod != "weighted" || fake.updateSettingsArg.RankingTiePolicy != "dense" || fake.updateSettingsArg.Notes != "Revisi" {
		t.Fatalf("UpdateReportSettings arg = %+v", fake.updateSettingsArg)
	}

	rec = httptest.NewRecorder()
	h.UpsertStudentSubjectDescription(rec, withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/assignments/"+assignmentID.String()+"/descriptions", `{"student_id":"`+studentID.String()+`","description":"Sangat baik"}`, teacherID), "id", assignmentID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpsertStudentSubjectDescription status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.descriptionArg.AssignmentID != assignmentID || fake.descriptionArg.StudentID != studentID || fake.descriptionArg.Description != "Sangat baik" || fake.descriptionTeacherID != teacherID {
		t.Fatalf("description arg=%+v teacher=%v", fake.descriptionArg, fake.descriptionTeacherID)
	}
}

func TestGradeReportSettingsAndDescriptionErrors(t *testing.T) {
	assignmentID := handlerTestUUID(149)
	studentID := handlerTestUUID(150)
	teacherID := handlerTestUUID(151)
	errDB := errors.New("db down")
	tests := []struct {
		name    string
		handler func(*Grade, http.ResponseWriter, *http.Request)
		svc     *fakeGradeService
		req     *http.Request
		want    int
	}{
		{"settings forbidden", (*Grade).GetReportSettings, &fakeGradeService{Grade: &service.Grade{}}, httptest.NewRequest(http.MethodGet, "/api/grades/report-settings", nil), http.StatusForbidden},
		{"settings internal", (*Grade).GetReportSettings, &fakeGradeService{Grade: &service.Grade{}, reportSettingsErr: errDB}, gradeGuruRequest(http.MethodGet, "/api/grades/report-settings", "", teacherID), http.StatusInternalServerError},
		{"update invalid json", (*Grade).UpdateReportSettings, &fakeGradeService{Grade: &service.Grade{}}, gradeGuruRequest(http.MethodPut, "/api/grades/report-settings", `{`, teacherID), http.StatusBadRequest},
		{"update invalid year", (*Grade).UpdateReportSettings, &fakeGradeService{Grade: &service.Grade{}}, gradeGuruRequest(http.MethodPut, "/api/grades/report-settings", `{"academic_year_id":"bad"}`, teacherID), http.StatusBadRequest},
		{"update service validation", (*Grade).UpdateReportSettings, &fakeGradeService{Grade: &service.Grade{}, updateSettingsErr: errors.New("metode ranking tidak valid")}, gradeGuruRequest(http.MethodPut, "/api/grades/report-settings", `{"academic_year_id":"`+assignmentID.String()+`"}`, teacherID), http.StatusBadRequest},
		{"description invalid assignment", (*Grade).UpsertStudentSubjectDescription, &fakeGradeService{Grade: &service.Grade{}}, withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/assignments/bad/descriptions", `{}`, teacherID), "id", "bad"), http.StatusBadRequest},
		{"description invalid json", (*Grade).UpsertStudentSubjectDescription, &fakeGradeService{Grade: &service.Grade{}}, withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/assignments/"+assignmentID.String()+"/descriptions", `{`, teacherID), "id", assignmentID.String()), http.StatusBadRequest},
		{"description invalid student", (*Grade).UpsertStudentSubjectDescription, &fakeGradeService{Grade: &service.Grade{}}, withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/assignments/"+assignmentID.String()+"/descriptions", `{"student_id":"bad"}`, teacherID), "id", assignmentID.String()), http.StatusBadRequest},
		{"description service validation", (*Grade).UpsertStudentSubjectDescription, &fakeGradeService{Grade: &service.Grade{}, descriptionErr: errors.New("deskripsi terlalu panjang")}, withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/assignments/"+assignmentID.String()+"/descriptions", `{"student_id":"`+studentID.String()+`"}`, teacherID), "id", assignmentID.String()), http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Grade{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestGradeSuccessHandlersForwardPayloads(t *testing.T) {
	assignmentID := handlerTestUUID(131)
	componentID := handlerTestUUID(132)
	studentID := handlerTestUUID(133)
	teacherID := handlerTestUUID(134)
	fake := &fakeGradeService{
		Grade:          &service.Grade{},
		overviewResult: service.GradeOverview{Readiness: service.GradeReadiness{Ready: true}},
		createRow:      gradeComponent(componentID, assignmentID, "UH 1"),
		updateRow:      gradeComponent(componentID, assignmentID, "UH Revisi"),
		publishRow:     gradeComponent(componentID, assignmentID, "UH Revisi"),
		finalizeRow:    db.GradeAssignmentFinalization{AssignmentID: assignmentID, FinalizedBy: "guru.ipa", Notes: "Siap"},
		entryRow: db.GradeEntry{
			ID:          handlerTestUUID(135),
			ComponentID: componentID,
			StudentID:   studentID,
			Score:       pgtype.Float8{Float64: 88.5, Valid: true},
			Notes:       "Baik",
			GradedBy:    "guru.ipa",
		},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &Grade{svc: fake, audit: audit}

	rec := httptest.NewRecorder()
	h.Overview(rec, gradeGuruRequest(http.MethodGet, "/api/grades?assignment_id="+assignmentID.String()+"&component_id="+componentID.String()+"&published_only=published", "", teacherID))
	if rec.Code != http.StatusOK {
		t.Fatalf("Overview status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.overviewAssignmentID != assignmentID || fake.overviewComponentID != componentID || !fake.overviewPublishedOnly || fake.overviewTeacherID != teacherID {
		t.Fatalf("Overview args = (%v, %v, %v, %v), want query and teacher scope", fake.overviewAssignmentID, fake.overviewComponentID, fake.overviewPublishedOnly, fake.overviewTeacherID)
	}

	rec = httptest.NewRecorder()
	h.CreateComponent(rec, gradeGuruRequest(http.MethodPost, "/api/grades/components", `{"assignment_id":"`+assignmentID.String()+`","title":"UH 1","category":"daily","weight":1.5,"max_score":100,"is_published":true}`, teacherID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateComponent status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createArg.AssignmentID != assignmentID || fake.createArg.Title != "UH 1" || fake.createArg.Category != "daily" || fake.createArg.Weight != 1.5 || fake.createArg.MaxScore != 100 || !fake.createArg.IsPublished || fake.createTeacherID != teacherID {
		t.Fatalf("CreateComponent args = %+v teacher=%v, want decoded payload and teacher", fake.createArg, fake.createTeacherID)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "GRADE_COMPONENT_CREATE" {
		t.Fatalf("CreateComponent audit = %#v, want create audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.UpdateComponent(rec, withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/"+componentID.String(), `{"title":"UH Revisi","category":"exam","weight":2,"max_score":90}`, teacherID), "id", componentID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateComponent status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateArg.ID != componentID || fake.updateArg.Title != "UH Revisi" || fake.updateArg.Category != "exam" || fake.updateArg.Weight != 2 || fake.updateArg.MaxScore != 90 || fake.updateTeacherID != teacherID {
		t.Fatalf("UpdateComponent args = %+v teacher=%v, want decoded payload and teacher", fake.updateArg, fake.updateTeacherID)
	}
	if len(audit.entries) != 2 || audit.entries[1].Action != "GRADE_COMPONENT_UPDATE" {
		t.Fatalf("UpdateComponent audit = %#v, want update audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.SetComponentPublished(rec, withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/"+componentID.String()+"/publish", `{"is_published":false}`, teacherID), "id", componentID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("SetComponentPublished status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.publishID != componentID || fake.publishValue || fake.publishTeacherID != teacherID {
		t.Fatalf("SetComponentPublished args = (%v, %v, %v), want id false teacher", fake.publishID, fake.publishValue, fake.publishTeacherID)
	}
	if len(audit.entries) != 3 || audit.entries[2].Action != "GRADE_COMPONENT_PUBLISH_TOGGLE" {
		t.Fatalf("SetComponentPublished audit = %#v, want publish toggle audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.FinalizeAssignment(rec, withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/"+assignmentID.String()+"/finalize", `{"notes":"Siap"}`, teacherID), "id", assignmentID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("FinalizeAssignment status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.finalizeAssignmentID != assignmentID || fake.finalizeBy != "guru.ipa" || fake.finalizeNotes != "Siap" || fake.finalizeTeacherID != teacherID {
		t.Fatalf("FinalizeAssignment args = (%v, %q, %q, %v), want assignment user notes teacher", fake.finalizeAssignmentID, fake.finalizeBy, fake.finalizeNotes, fake.finalizeTeacherID)
	}
	if len(audit.entries) != 4 || audit.entries[3].Action != "GRADE_ASSIGNMENT_FINALIZE" {
		t.Fatalf("FinalizeAssignment audit = %#v, want finalize audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.ReopenAssignment(rec, withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/"+assignmentID.String()+"/reopen", "", teacherID), "id", assignmentID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("ReopenAssignment status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.reopenAssignmentID != assignmentID || fake.reopenTeacherID != teacherID {
		t.Fatalf("ReopenAssignment args = (%v, %v), want assignment teacher", fake.reopenAssignmentID, fake.reopenTeacherID)
	}
	if len(audit.entries) != 5 || audit.entries[4].Action != "GRADE_ASSIGNMENT_REOPEN" {
		t.Fatalf("ReopenAssignment audit = %#v, want reopen audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.DeleteComponent(rec, withRouteParam(gradeGuruRequest(http.MethodDelete, "/api/grades/components/"+componentID.String(), "", teacherID), "id", componentID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteComponent status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != componentID || fake.deleteTeacherID != teacherID {
		t.Fatalf("DeleteComponent args = (%v, %v), want component teacher", fake.deleteID, fake.deleteTeacherID)
	}
	if len(audit.entries) != 6 || audit.entries[5].Action != "GRADE_COMPONENT_DELETE" {
		t.Fatalf("DeleteComponent audit = %#v, want delete audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.UpsertEntry(rec, withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/components/"+componentID.String()+"/entries", `{"student_id":"`+studentID.String()+`","score":88.5,"notes":"Baik"}`, teacherID), "id", componentID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpsertEntry status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.entryComponentID != componentID || fake.entryStudentID != studentID || fake.entryTeacherID != teacherID || fake.entryScore != 88.5 || fake.entryNotes != "Baik" || fake.entryGradedBy != "guru.ipa" {
		t.Fatalf("UpsertEntry args = component:%v student:%v teacher:%v score:%v notes:%q by:%q", fake.entryComponentID, fake.entryStudentID, fake.entryTeacherID, fake.entryScore, fake.entryNotes, fake.entryGradedBy)
	}
	if len(audit.entries) != 7 || audit.entries[6].Action != "GRADE_ENTRY_UPSERT" {
		t.Fatalf("UpsertEntry audit = %#v, want entry upsert audit", audit.entries)
	}
}

func TestGradeHandlersMapServiceErrors(t *testing.T) {
	assignmentID := handlerTestUUID(136)
	componentID := handlerTestUUID(137)
	studentID := handlerTestUUID(138)
	teacherID := handlerTestUUID(139)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*Grade, http.ResponseWriter, *http.Request)
		svc        *fakeGradeService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "overview",
			handler:    (*Grade).Overview,
			svc:        &fakeGradeService{Grade: &service.Grade{}, overviewErr: errDB},
			req:        gradeGuruRequest(http.MethodGet, "/api/grades?assignment_id="+assignmentID.String(), "", teacherID),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create validation",
			handler:    (*Grade).CreateComponent,
			svc:        &fakeGradeService{Grade: &service.Grade{}, createErr: errors.New("judul komponen wajib diisi")},
			req:        gradeGuruRequest(http.MethodPost, "/api/grades/components", `{"assignment_id":"`+assignmentID.String()+`"}`, teacherID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update validation",
			handler:    (*Grade).UpdateComponent,
			svc:        &fakeGradeService{Grade: &service.Grade{}, updateErr: errors.New("bobot tidak boleh negatif")},
			req:        withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/"+componentID.String(), `{}`, teacherID), "id", componentID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "publish internal",
			handler:    (*Grade).SetComponentPublished,
			svc:        &fakeGradeService{Grade: &service.Grade{}, publishErr: errDB},
			req:        withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/"+componentID.String()+"/publish", `{"is_published":true}`, teacherID), "id", componentID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "finalize validation",
			handler:    (*Grade).FinalizeAssignment,
			svc:        &fakeGradeService{Grade: &service.Grade{}, finalizeErr: errors.New("assignment belum siap difinalisasi")},
			req:        withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/"+assignmentID.String()+"/finalize", `{}`, teacherID), "id", assignmentID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "reopen internal",
			handler:    (*Grade).ReopenAssignment,
			svc:        &fakeGradeService{Grade: &service.Grade{}, reopenErr: errDB},
			req:        withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/"+assignmentID.String()+"/reopen", "", teacherID), "id", assignmentID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "delete internal",
			handler:    (*Grade).DeleteComponent,
			svc:        &fakeGradeService{Grade: &service.Grade{}, deleteErr: errDB},
			req:        withRouteParam(gradeGuruRequest(http.MethodDelete, "/api/grades/components/"+componentID.String(), "", teacherID), "id", componentID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "upsert validation",
			handler:    (*Grade).UpsertEntry,
			svc:        &fakeGradeService{Grade: &service.Grade{}, entryErr: errors.New("nilai tidak boleh negatif")},
			req:        withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/components/"+componentID.String()+"/entries", `{"student_id":"`+studentID.String()+`","score":-1}`, teacherID), "id", componentID.String()),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Grade{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestGradeValidationBranches(t *testing.T) {
	assignmentID := handlerTestUUID(140)
	componentID := handlerTestUUID(141)
	studentID := handlerTestUUID(142)
	teacherID := handlerTestUUID(143)

	tests := []struct {
		name       string
		handler    func(*Grade, http.ResponseWriter, *http.Request)
		req        *http.Request
		wantStatus int
	}{
		{name: "overview invalid assignment", handler: (*Grade).Overview, req: gradeGuruRequest(http.MethodGet, "/api/grades?assignment_id=bad", "", teacherID), wantStatus: http.StatusBadRequest},
		{name: "overview invalid component", handler: (*Grade).Overview, req: gradeGuruRequest(http.MethodGet, "/api/grades?component_id=bad", "", teacherID), wantStatus: http.StatusBadRequest},
		{name: "create invalid json", handler: (*Grade).CreateComponent, req: gradeGuruRequest(http.MethodPost, "/api/grades/components", `{`, teacherID), wantStatus: http.StatusBadRequest},
		{name: "create invalid assignment", handler: (*Grade).CreateComponent, req: gradeGuruRequest(http.MethodPost, "/api/grades/components", `{"assignment_id":"bad"}`, teacherID), wantStatus: http.StatusBadRequest},
		{name: "update invalid id", handler: (*Grade).UpdateComponent, req: withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/bad", `{}`, teacherID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "update invalid json", handler: (*Grade).UpdateComponent, req: withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/"+componentID.String(), `{`, teacherID), "id", componentID.String()), wantStatus: http.StatusBadRequest},
		{name: "publish invalid id", handler: (*Grade).SetComponentPublished, req: withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/bad/publish", `{}`, teacherID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "publish invalid json", handler: (*Grade).SetComponentPublished, req: withRouteParam(gradeGuruRequest(http.MethodPatch, "/api/grades/components/"+componentID.String()+"/publish", `{`, teacherID), "id", componentID.String()), wantStatus: http.StatusBadRequest},
		{name: "finalize invalid id", handler: (*Grade).FinalizeAssignment, req: withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/bad/finalize", `{}`, teacherID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "finalize invalid json", handler: (*Grade).FinalizeAssignment, req: withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/"+assignmentID.String()+"/finalize", `{`, teacherID), "id", assignmentID.String()), wantStatus: http.StatusBadRequest},
		{name: "reopen invalid id", handler: (*Grade).ReopenAssignment, req: withRouteParam(gradeGuruRequest(http.MethodPost, "/api/grades/assignments/bad/reopen", "", teacherID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "delete invalid id", handler: (*Grade).DeleteComponent, req: withRouteParam(gradeGuruRequest(http.MethodDelete, "/api/grades/components/bad", "", teacherID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "entry invalid component", handler: (*Grade).UpsertEntry, req: withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/components/bad/entries", `{}`, teacherID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "entry invalid json", handler: (*Grade).UpsertEntry, req: withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/components/"+componentID.String()+"/entries", `{`, teacherID), "id", componentID.String()), wantStatus: http.StatusBadRequest},
		{name: "entry invalid student", handler: (*Grade).UpsertEntry, req: withRouteParam(gradeGuruRequest(http.MethodPut, "/api/grades/components/"+componentID.String()+"/entries", `{"student_id":"bad"}`, teacherID), "id", componentID.String()), wantStatus: http.StatusBadRequest},
		{name: "entry valid with sub username", handler: (*Grade).UpsertEntry, req: withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/api/grades/components/"+componentID.String()+"/entries", strings.NewReader(`{"student_id":"`+studentID.String()+`","score":75}`)), jwt.MapClaims{"roles": []any{"guru"}, "role": "guru", "eid": teacherID.String(), "sub": "guru-sub"}), "id", componentID.String()), wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Grade{svc: &fakeGradeService{Grade: &service.Grade{}}}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestGradeIdentityHelpersFallbacks(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/grades", nil)
	if got := currentGradeUsername(req); got != "" {
		t.Fatalf("currentGradeUsername(no claims) = %q, want empty", got)
	}
	req = withClaims(req, jwt.MapClaims{"sub": "guru-sub"})
	if got := currentGradeUsername(req); got != "guru-sub" {
		t.Fatalf("currentGradeUsername(sub) = %q, want guru-sub", got)
	}
	if id := gradeTeacherEmployeeID(withClaims(httptest.NewRequest(http.MethodGet, "/api/grades", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": "bad"})); id.Valid {
		t.Fatalf("gradeTeacherEmployeeID(bad eid).Valid = true, want false")
	}
}
