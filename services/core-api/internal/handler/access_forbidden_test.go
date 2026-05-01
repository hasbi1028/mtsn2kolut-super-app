package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mtsn2kolut-super-app/backend/internal/service"
)

func assertHandlerForbidden(t *testing.T, name string, fn http.HandlerFunc) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/forbidden?period_year=2026", strings.NewReader(`{}`))
	fn(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("%s status = %d, want 403; body=%s", name, rec.Code, rec.Body.String())
	}
}

func TestGovernanceHandlersRejectMissingRole(t *testing.T) {
	h := &Governance{svc: &service.Governance{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Stats", fn: h.Stats},
		{name: "SNPMatrix", fn: h.SNPMatrix},
		{name: "EmployeeOptions", fn: h.EmployeeOptions},
		{name: "ListUnits", fn: h.ListUnits},
		{name: "CreateUnit", fn: h.CreateUnit},
		{name: "UpdateUnit", fn: h.UpdateUnit},
		{name: "DeleteUnit", fn: h.DeleteUnit},
		{name: "ListPositions", fn: h.ListPositions},
		{name: "CreatePosition", fn: h.CreatePosition},
		{name: "UpdatePosition", fn: h.UpdatePosition},
		{name: "DeletePosition", fn: h.DeletePosition},
		{name: "ListAssignments", fn: h.ListAssignments},
		{name: "CreateAssignment", fn: h.CreateAssignment},
		{name: "UpdateAssignment", fn: h.UpdateAssignment},
		{name: "DeleteAssignment", fn: h.DeleteAssignment},
		{name: "ListDocuments", fn: h.ListDocuments},
		{name: "CreateDocument", fn: h.CreateDocument},
		{name: "UpdateDocument", fn: h.UpdateDocument},
		{name: "DeleteDocument", fn: h.DeleteDocument},
		{name: "ListPrograms", fn: h.ListPrograms},
		{name: "CreateProgram", fn: h.CreateProgram},
		{name: "UpdateProgram", fn: h.UpdateProgram},
		{name: "DeleteProgram", fn: h.DeleteProgram},
		{name: "ListWorkPlanItems", fn: h.ListWorkPlanItems},
		{name: "CreateWorkPlanItem", fn: h.CreateWorkPlanItem},
		{name: "UpdateWorkPlanItem", fn: h.UpdateWorkPlanItem},
		{name: "DeleteWorkPlanItem", fn: h.DeleteWorkPlanItem},
		{name: "ListPerformanceTargets", fn: h.ListPerformanceTargets},
		{name: "CreatePerformanceTarget", fn: h.CreatePerformanceTarget},
		{name: "UpdatePerformanceTarget", fn: h.UpdatePerformanceTarget},
		{name: "DeletePerformanceTarget", fn: h.DeletePerformanceTarget},
		{name: "ListEvidenceItems", fn: h.ListEvidenceItems},
		{name: "CreateEvidenceItem", fn: h.CreateEvidenceItem},
		{name: "UpdateEvidenceItem", fn: h.UpdateEvidenceItem},
		{name: "DeleteEvidenceItem", fn: h.DeleteEvidenceItem},
		{name: "ListComplianceActions", fn: h.ListComplianceActions},
		{name: "CreateComplianceAction", fn: h.CreateComplianceAction},
		{name: "UpdateComplianceAction", fn: h.UpdateComplianceAction},
		{name: "DeleteComplianceAction", fn: h.DeleteComplianceAction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestKesiswaanHandlersRejectMissingRole(t *testing.T) {
	h := &Kesiswaan{svc: &service.Kesiswaan{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Stats", fn: h.Stats},
		{name: "ClassOptions", fn: h.ClassOptions},
		{name: "ListStudents", fn: h.ListStudents},
		{name: "UpdateStudentProfile", fn: h.UpdateStudentProfile},
		{name: "UploadStudentPhoto", fn: h.UploadStudentPhoto},
		{name: "StudentPhotoFile", fn: h.StudentPhotoFile},
		{name: "ListCategories", fn: h.ListCategories},
		{name: "CreateCategory", fn: h.CreateCategory},
		{name: "UpdateCategory", fn: h.UpdateCategory},
		{name: "DeleteCategory", fn: h.DeleteCategory},
		{name: "ListViolations", fn: h.ListViolations},
		{name: "CreateViolation", fn: h.CreateViolation},
		{name: "UpdateViolation", fn: h.UpdateViolation},
		{name: "DeleteViolation", fn: h.DeleteViolation},
		{name: "ListAchievements", fn: h.ListAchievements},
		{name: "CreateAchievement", fn: h.CreateAchievement},
		{name: "UpdateAchievement", fn: h.UpdateAchievement},
		{name: "DeleteAchievement", fn: h.DeleteAchievement},
		{name: "ListExtracurriculars", fn: h.ListExtracurriculars},
		{name: "CreateExtracurricular", fn: h.CreateExtracurricular},
		{name: "UpdateExtracurricular", fn: h.UpdateExtracurricular},
		{name: "DeleteExtracurricular", fn: h.DeleteExtracurricular},
		{name: "ListExtracurricularMembers", fn: h.ListExtracurricularMembers},
		{name: "CreateExtracurricularMember", fn: h.CreateExtracurricularMember},
		{name: "UpdateExtracurricularMember", fn: h.UpdateExtracurricularMember},
		{name: "DeleteExtracurricularMember", fn: h.DeleteExtracurricularMember},
		{name: "ListCounselingSessions", fn: h.ListCounselingSessions},
		{name: "CreateCounselingSession", fn: h.CreateCounselingSession},
		{name: "UpdateCounselingSession", fn: h.UpdateCounselingSession},
		{name: "DeleteCounselingSession", fn: h.DeleteCounselingSession},
		{name: "ListStudentTransfers", fn: h.ListStudentTransfers},
		{name: "CreateStudentTransfer", fn: h.CreateStudentTransfer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestTUHandlersRejectMissingRole(t *testing.T) {
	letter := &Letter{svc: &service.Letter{}}
	archive := &Archive{svc: &service.Archive{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Letter.ListClassifications", fn: letter.ListClassifications},
		{name: "Letter.ListIncoming", fn: letter.ListIncoming},
		{name: "Letter.CreateIncoming", fn: letter.CreateIncoming},
		{name: "Letter.GetIncoming", fn: letter.GetIncoming},
		{name: "Letter.UpdateIncoming", fn: letter.UpdateIncoming},
		{name: "Letter.UpdateIncomingStatus", fn: letter.UpdateIncomingStatus},
		{name: "Letter.DeleteIncoming", fn: letter.DeleteIncoming},
		{name: "Letter.ListOutgoing", fn: letter.ListOutgoing},
		{name: "Letter.CreateOutgoing", fn: letter.CreateOutgoing},
		{name: "Letter.GetOutgoing", fn: letter.GetOutgoing},
		{name: "Letter.UpdateOutgoing", fn: letter.UpdateOutgoing},
		{name: "Letter.DeleteOutgoing", fn: letter.DeleteOutgoing},
		{name: "Letter.PreviewOutgoingNumber", fn: letter.PreviewOutgoingNumber},
		{name: "Letter.ListDispositions", fn: letter.ListDispositions},
		{name: "Letter.CreateDisposition", fn: letter.CreateDisposition},
		{name: "Letter.GetDisposition", fn: letter.GetDisposition},
		{name: "Letter.UpdateDisposition", fn: letter.UpdateDisposition},
		{name: "Letter.DeleteDisposition", fn: letter.DeleteDisposition},
		{name: "Archive.Stats", fn: archive.Stats},
		{name: "Archive.ListCategories", fn: archive.ListCategories},
		{name: "Archive.CreateCategory", fn: archive.CreateCategory},
		{name: "Archive.UpdateCategory", fn: archive.UpdateCategory},
		{name: "Archive.DeleteCategory", fn: archive.DeleteCategory},
		{name: "Archive.ListDocuments", fn: archive.ListDocuments},
		{name: "Archive.UploadDocument", fn: archive.UploadDocument},
		{name: "Archive.GetDocument", fn: archive.GetDocument},
		{name: "Archive.UpdateDocument", fn: archive.UpdateDocument},
		{name: "Archive.DeleteDocument", fn: archive.DeleteDocument},
		{name: "Archive.File", fn: archive.File},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestGradeHandlersRejectMissingRole(t *testing.T) {
	h := &Grade{svc: &service.Grade{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Overview", fn: h.Overview},
		{name: "CreateComponent", fn: h.CreateComponent},
		{name: "UpdateComponent", fn: h.UpdateComponent},
		{name: "SetComponentPublished", fn: h.SetComponentPublished},
		{name: "FinalizeAssignment", fn: h.FinalizeAssignment},
		{name: "ReopenAssignment", fn: h.ReopenAssignment},
		{name: "DeleteComponent", fn: h.DeleteComponent},
		{name: "UpsertEntry", fn: h.UpsertEntry},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestLibraryHandlersRejectMissingRole(t *testing.T) {
	h := &Library{svc: &service.Library{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Stats", fn: h.Stats},
		{name: "ListBooks", fn: h.ListBooks},
		{name: "CreateBook", fn: h.CreateBook},
		{name: "UpdateBook", fn: h.UpdateBook},
		{name: "DeleteBook", fn: h.DeleteBook},
		{name: "ListLoans", fn: h.ListLoans},
		{name: "LoanBook", fn: h.LoanBook},
		{name: "ReturnBook", fn: h.ReturnBook},
		{name: "MarkDendaLunas", fn: h.MarkDendaLunas},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestCbtSessionSetupHandlersRejectMissingRole(t *testing.T) {
	h := &CbtSession{svc: &service.CbtSession{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
		req  *http.Request
	}{
		{name: "Enroll", fn: h.Enroll, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/enroll", strings.NewReader(`{}`)), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "EnrollGrade", fn: h.EnrollGrade, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/enroll-grade", strings.NewReader(`{}`)), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "EnrollSchool", fn: h.EnrollSchool, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/enroll-school", nil), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "GenerateTokens", fn: h.GenerateTokens, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/generate-tokens", nil), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "RegenerateToken", fn: h.RegenerateToken, req: withRouteParams(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/participants/22222222-2222-2222-2222-222222222222/regenerate-token", nil), "id", "11111111-1111-1111-1111-111111111111", "pid", "22222222-2222-2222-2222-222222222222")},
		{name: "AssignSeat", fn: h.AssignSeat, req: withRouteParams(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/participants/22222222-2222-2222-2222-222222222222/seat", strings.NewReader(`{}`)), "id", "11111111-1111-1111-1111-111111111111", "pid", "22222222-2222-2222-2222-222222222222")},
		{name: "AutoAssignSeats", fn: h.AutoAssignSeats, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/seats/auto", nil), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "CreateRoom", fn: h.CreateRoom, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/rooms", strings.NewReader(`{}`)), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "DeleteRoom", fn: h.DeleteRoom, req: withRouteParams(httptest.NewRequest(http.MethodDelete, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/rooms/33333333-3333-3333-3333-333333333333", nil), "id", "11111111-1111-1111-1111-111111111111", "rid", "33333333-3333-3333-3333-333333333333")},
		{name: "ShuffleRooms", fn: h.ShuffleRooms, req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/shuffle-rooms", nil), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "RecordAnswer", fn: h.RecordAnswer, req: withRouteParams(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/participants/22222222-2222-2222-2222-222222222222/answers", strings.NewReader(`{}`)), "id", "11111111-1111-1111-1111-111111111111", "pid", "22222222-2222-2222-2222-222222222222")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("%s status = %d, want 403; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAcademicStudentParentAndUserHandlersRejectMissingRole(t *testing.T) {
	academic := &Academic{svc: &service.Academic{}}
	employee := &Employee{svc: &service.Employee{}}
	employeeSchedule := &EmployeeSchedule{svc: &service.EmployeeSchedule{}}
	student := &Student{svc: &service.Student{}}
	parent := &Parent{svc: &service.Parent{}}
	user := NewUser(nil)
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Academic.Create", fn: academic.Create},
		{name: "Academic.Update", fn: academic.Update},
		{name: "Academic.Delete", fn: academic.Delete},
		{name: "Employee.ListWithStatus", fn: employee.ListWithStatus},
		{name: "Employee.ListPusakaEligibleWithStatus", fn: employee.ListPusakaEligibleWithStatus},
		{name: "Employee.Get", fn: employee.Get},
		{name: "Employee.Create", fn: employee.Create},
		{name: "Employee.Update", fn: employee.Update},
		{name: "Employee.Delete", fn: employee.Delete},
		{name: "Employee.UpdateStatus", fn: employee.UpdateStatus},
		{name: "Employee.GetPusakaStatus", fn: employee.GetPusakaStatus},
		{name: "Employee.UpdatePusakaCredentials", fn: employee.UpdatePusakaCredentials},
		{name: "Employee.UpdatePusakaAccountStatus", fn: employee.UpdatePusakaAccountStatus},
		{name: "Employee.DeletePusakaAccount", fn: employee.DeletePusakaAccount},
		{name: "Employee.ListPusakaAuditLogs", fn: employee.ListPusakaAuditLogs},
		{name: "EmployeeSchedule.List", fn: employeeSchedule.List},
		{name: "EmployeeSchedule.Upsert", fn: employeeSchedule.Upsert},
		{name: "EmployeeSchedule.Delete", fn: employeeSchedule.Delete},
		{name: "Student.List", fn: student.List},
		{name: "Student.GuruAwareList", fn: student.GuruAwareList},
		{name: "Student.Create", fn: student.Create},
		{name: "Student.Update", fn: student.Update},
		{name: "Student.Delete", fn: student.Delete},
		{name: "Student.UpdateLifecycle", fn: student.UpdateLifecycle},
		{name: "Parent.List", fn: parent.List},
		{name: "Parent.Get", fn: parent.Get},
		{name: "Parent.Create", fn: parent.Create},
		{name: "Parent.Update", fn: parent.Update},
		{name: "Parent.Delete", fn: parent.Delete},
		{name: "Parent.LinkStudent", fn: parent.LinkStudent},
		{name: "Parent.UnlinkStudent", fn: parent.UnlinkStudent},
		{name: "Parent.ListChildren", fn: parent.ListChildren},
		{name: "User.Delete", fn: user.Delete},
		{name: "User.UpdateStatus", fn: user.UpdateStatus},
		{name: "User.ListAuditLogs", fn: user.ListAuditLogs},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestCbtHandlersRejectMissingRole(t *testing.T) {
	question := NewCbtQuestion(&service.CbtQuestion{})
	asset := NewCbtQuestionAsset(&service.CbtQuestionAsset{})
	pkg := NewCbtPackage(&service.CbtPackage{})
	event := NewCbtEvent(&service.CbtEvent{})
	session := NewCbtSession(&service.CbtSession{})
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "CbtQuestion.List", fn: question.List},
		{name: "CbtQuestion.Get", fn: question.Get},
		{name: "CbtQuestion.Create", fn: question.Create},
		{name: "CbtQuestion.Update", fn: question.Update},
		{name: "CbtQuestion.Delete", fn: question.Delete},
		{name: "CbtQuestion.WorkflowAction", fn: question.WorkflowAction},
		{name: "CbtQuestion.Duplicate", fn: question.Duplicate},
		{name: "CbtQuestionAsset.List", fn: asset.List},
		{name: "CbtQuestionAsset.Upload", fn: asset.Upload},
		{name: "CbtPackage.List", fn: pkg.List},
		{name: "CbtPackage.Create", fn: pkg.Create},
		{name: "CbtPackage.Delete", fn: pkg.Delete},
		{name: "CbtEvent.List", fn: event.List},
		{name: "CbtEvent.Get", fn: event.Get},
		{name: "CbtEvent.GetResults", fn: event.GetResults},
		{name: "CbtEvent.GetExamCards", fn: event.GetExamCards},
		{name: "CbtEvent.Create", fn: event.Create},
		{name: "CbtEvent.Update", fn: event.Update},
		{name: "CbtEvent.UpdateStatus", fn: event.UpdateStatus},
		{name: "CbtEvent.Delete", fn: event.Delete},
		{name: "CbtSession.List", fn: session.List},
		{name: "CbtSession.Create", fn: session.Create},
		{name: "CbtSession.UpdateStatus", fn: session.UpdateStatus},
		{name: "CbtSession.Delete", fn: session.Delete},
		{name: "CbtSession.GetResults", fn: session.GetResults},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}

func TestClassJournalHandlersRejectMissingRole(t *testing.T) {
	h := &ClassJournal{svc: &service.ClassJournal{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Overview", fn: h.Overview},
		{name: "CreateSession", fn: h.CreateSession},
		{name: "GetSession", fn: h.GetSession},
		{name: "UpdateSession", fn: h.UpdateSession},
		{name: "DeleteSession", fn: h.DeleteSession},
		{name: "BulkUpsertAttendances", fn: h.BulkUpsertAttendances},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHandlerForbidden(t, tt.name, tt.fn)
		})
	}
}
