package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeKesiswaanService struct {
	statsErr        error
	classOptionsErr error
	listErr         error
	mutationErr     error
	deleteErr       error
	photoErr        error
	canReadPhotoErr error
	statsTeacher    pgtype.UUID

	listStudentsSearch  string
	listStudentsStatus  string
	listStudentsClassID string
	listStudentsTeacher pgtype.UUID

	updateStudentProfileArg db.UpdateKesiswaanStudentProfileParams
	savePhotoInput          service.UploadStudentPhotoInput
	canReadPhotoFilename    string
	canReadPhotoValue       bool
	photoPathFilename       string
	photoPathValue          string
	photoPathOK             bool

	categorySearch    string
	createCategoryArg db.CreateViolationCategoryParams
	updateCategoryArg db.UpdateViolationCategoryParams
	deleteCategoryID  pgtype.UUID

	listViolationsSearch    string
	listViolationsStatus    string
	listViolationsStudentID string
	createViolationArg      db.CreateStudentViolationParams
	updateViolationArg      db.UpdateStudentViolationParams
	deleteViolationID       pgtype.UUID

	listAchievementsSearch    string
	listAchievementsLevel     string
	listAchievementsStudentID string
	createAchievementArg      db.CreateStudentAchievementParams
	updateAchievementArg      db.UpdateStudentAchievementParams
	deleteAchievementID       pgtype.UUID

	listExtracurricularsSearch     string
	listExtracurricularsActiveOnly bool
	createExtracurricularArg       db.CreateExtracurricularParams
	updateExtracurricularArg       db.UpdateExtracurricularParams
	deleteExtracurricularID        pgtype.UUID

	listMembersSearch            string
	listMembersStatus            string
	listMembersExtracurricularID string
	listMembersStudentID         string
	createMemberArg              db.CreateExtracurricularMemberParams
	updateMemberArg              db.UpdateExtracurricularMemberParams
	deleteMemberID               pgtype.UUID

	listCounselingSearch       string
	listCounselingStatus       string
	listCounselingStudentID    string
	listCounselingConfidential bool
	createCounselingArg        db.CreateCounselingSessionParams
	updateCounselingArg        db.UpdateCounselingSessionParams
	deleteCounselingID         pgtype.UUID

	listTransfersSearch  string
	listTransfersType    string
	listTransfersStudent string
	createTransferArg    db.CreateStudentTransferParams
}

func (f *fakeKesiswaanService) Stats(ctx context.Context, teacherEmployeeID pgtype.UUID) (service.KesiswaanStats, error) {
	f.statsTeacher = teacherEmployeeID
	return service.KesiswaanStats{ActiveStudents: 12}, f.statsErr
}

func (f *fakeKesiswaanService) ClassOptions(ctx context.Context) ([]db.ListKesiswaanClassOptionsRow, error) {
	if f.classOptionsErr != nil {
		return nil, f.classOptionsErr
	}
	return []db.ListKesiswaanClassOptionsRow{{ID: handlerTestUUID(91), Code: "VII-A", Name: "VII A", Level: "VII"}}, nil
}

func (f *fakeKesiswaanService) ListStudents(ctx context.Context, search, status, classID string, teacherEmployeeID pgtype.UUID) ([]db.ListKesiswaanStudentsRow, error) {
	f.listStudentsSearch = search
	f.listStudentsStatus = status
	f.listStudentsClassID = classID
	f.listStudentsTeacher = teacherEmployeeID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListKesiswaanStudentsRow{{ID: handlerTestUUID(92), Nama: "Siswa A"}}, nil
}

func (f *fakeKesiswaanService) UpdateStudentProfile(ctx context.Context, arg db.UpdateKesiswaanStudentProfileParams) (db.Student, error) {
	f.updateStudentProfileArg = arg
	return db.Student{ID: arg.ID, Nama: "Siswa A"}, f.mutationErr
}

func (f *fakeKesiswaanService) SaveStudentPhoto(ctx context.Context, input service.UploadStudentPhotoInput) (db.Student, error) {
	f.savePhotoInput = input
	return db.Student{ID: input.StudentID, PhotoUrl: "/api/kesiswaan/student-photos/foto.jpg"}, f.photoErr
}

func (f *fakeKesiswaanService) StudentPhotoPath(filename string) (string, bool) {
	f.photoPathFilename = filename
	return f.photoPathValue, f.photoPathOK
}

func (f *fakeKesiswaanService) CanReadStudentPhoto(ctx context.Context, filename string, teacherEmployeeID pgtype.UUID) (bool, error) {
	f.canReadPhotoFilename = filename
	return f.canReadPhotoValue, f.canReadPhotoErr
}

func (f *fakeKesiswaanService) ListCategories(ctx context.Context, search string) ([]db.ViolationCategory, error) {
	f.categorySearch = search
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ViolationCategory{{ID: handlerTestUUID(93), Code: "TL", Name: "Terlambat"}}, nil
}

func (f *fakeKesiswaanService) CreateCategory(ctx context.Context, arg db.CreateViolationCategoryParams) (db.ViolationCategory, error) {
	f.createCategoryArg = arg
	return db.ViolationCategory{ID: handlerTestUUID(93), Code: arg.Code, IsActive: arg.IsActive}, f.mutationErr
}

func (f *fakeKesiswaanService) UpdateCategory(ctx context.Context, arg db.UpdateViolationCategoryParams) (db.ViolationCategory, error) {
	f.updateCategoryArg = arg
	return db.ViolationCategory{ID: arg.ID, Code: arg.Code, IsActive: arg.IsActive}, f.mutationErr
}

func (f *fakeKesiswaanService) DeleteCategory(ctx context.Context, id pgtype.UUID) error {
	f.deleteCategoryID = id
	return f.deleteErr
}

func (f *fakeKesiswaanService) ListViolations(ctx context.Context, search, status, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentViolationsRow, error) {
	f.listViolationsSearch = search
	f.listViolationsStatus = status
	f.listViolationsStudentID = studentID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListStudentViolationsRow{{ID: handlerTestUUID(94), Status: "open"}}, nil
}

func (f *fakeKesiswaanService) CreateViolation(ctx context.Context, arg db.CreateStudentViolationParams) (db.StudentViolation, error) {
	f.createViolationArg = arg
	return db.StudentViolation{ID: handlerTestUUID(94), StudentID: arg.StudentID, Status: arg.Status}, f.mutationErr
}

func (f *fakeKesiswaanService) UpdateViolation(ctx context.Context, arg db.UpdateStudentViolationParams) (db.StudentViolation, error) {
	f.updateViolationArg = arg
	return db.StudentViolation{ID: arg.ID, StudentID: arg.StudentID, Status: arg.Status}, f.mutationErr
}

func (f *fakeKesiswaanService) DeleteViolation(ctx context.Context, id pgtype.UUID) error {
	f.deleteViolationID = id
	return f.deleteErr
}

func (f *fakeKesiswaanService) ListAchievements(ctx context.Context, search, level, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentAchievementsRow, error) {
	f.listAchievementsSearch = search
	f.listAchievementsLevel = level
	f.listAchievementsStudentID = studentID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListStudentAchievementsRow{{ID: handlerTestUUID(95), Title: "Juara"}}, nil
}

func (f *fakeKesiswaanService) CreateAchievement(ctx context.Context, arg db.CreateStudentAchievementParams) (db.StudentAchievement, error) {
	f.createAchievementArg = arg
	return db.StudentAchievement{ID: handlerTestUUID(95), StudentID: arg.StudentID, Title: arg.Title}, f.mutationErr
}

func (f *fakeKesiswaanService) UpdateAchievement(ctx context.Context, arg db.UpdateStudentAchievementParams) (db.StudentAchievement, error) {
	f.updateAchievementArg = arg
	return db.StudentAchievement{ID: arg.ID, StudentID: arg.StudentID, Title: arg.Title}, f.mutationErr
}

func (f *fakeKesiswaanService) DeleteAchievement(ctx context.Context, id pgtype.UUID) error {
	f.deleteAchievementID = id
	return f.deleteErr
}

func (f *fakeKesiswaanService) ListExtracurriculars(ctx context.Context, search string, activeOnly bool) ([]db.ListExtracurricularsRow, error) {
	f.listExtracurricularsSearch = search
	f.listExtracurricularsActiveOnly = activeOnly
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListExtracurricularsRow{{ID: handlerTestUUID(96), Code: "PRM", Name: "Pramuka"}}, nil
}

func (f *fakeKesiswaanService) CreateExtracurricular(ctx context.Context, arg db.CreateExtracurricularParams) (db.Extracurricular, error) {
	f.createExtracurricularArg = arg
	return db.Extracurricular{ID: handlerTestUUID(96), Code: arg.Code, IsActive: arg.IsActive}, f.mutationErr
}

func (f *fakeKesiswaanService) UpdateExtracurricular(ctx context.Context, arg db.UpdateExtracurricularParams) (db.Extracurricular, error) {
	f.updateExtracurricularArg = arg
	return db.Extracurricular{ID: arg.ID, Code: arg.Code, IsActive: arg.IsActive}, f.mutationErr
}

func (f *fakeKesiswaanService) DeleteExtracurricular(ctx context.Context, id pgtype.UUID) error {
	f.deleteExtracurricularID = id
	return f.deleteErr
}

func (f *fakeKesiswaanService) ListExtracurricularMembers(ctx context.Context, search, status, extracurricularID, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListExtracurricularMembersRow, error) {
	f.listMembersSearch = search
	f.listMembersStatus = status
	f.listMembersExtracurricularID = extracurricularID
	f.listMembersStudentID = studentID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListExtracurricularMembersRow{{ID: handlerTestUUID(97), Role: "member"}}, nil
}

func (f *fakeKesiswaanService) CreateExtracurricularMember(ctx context.Context, arg db.CreateExtracurricularMemberParams) (db.ExtracurricularMember, error) {
	f.createMemberArg = arg
	return db.ExtracurricularMember{ID: handlerTestUUID(97), StudentID: arg.StudentID, Role: arg.Role}, f.mutationErr
}

func (f *fakeKesiswaanService) UpdateExtracurricularMember(ctx context.Context, arg db.UpdateExtracurricularMemberParams) (db.ExtracurricularMember, error) {
	f.updateMemberArg = arg
	return db.ExtracurricularMember{ID: arg.ID, StudentID: arg.StudentID, Role: arg.Role}, f.mutationErr
}

func (f *fakeKesiswaanService) DeleteExtracurricularMember(ctx context.Context, id pgtype.UUID) error {
	f.deleteMemberID = id
	return f.deleteErr
}

func (f *fakeKesiswaanService) ListCounselingSessions(ctx context.Context, search, status, studentID string, canReadConfidential bool, teacherEmployeeID pgtype.UUID) ([]db.ListCounselingSessionsRow, error) {
	f.listCounselingSearch = search
	f.listCounselingStatus = status
	f.listCounselingStudentID = studentID
	f.listCounselingConfidential = canReadConfidential
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListCounselingSessionsRow{{ID: handlerTestUUID(98), Topic: "Belajar"}}, nil
}

func (f *fakeKesiswaanService) CreateCounselingSession(ctx context.Context, arg db.CreateCounselingSessionParams) (db.CounselingSession, error) {
	f.createCounselingArg = arg
	return db.CounselingSession{
		ID:             handlerTestUUID(98),
		StudentID:      arg.StudentID,
		Topic:          arg.Topic,
		Status:         arg.Status,
		IsConfidential: arg.IsConfidential,
	}, f.mutationErr
}

func (f *fakeKesiswaanService) UpdateCounselingSession(ctx context.Context, arg db.UpdateCounselingSessionParams) (db.CounselingSession, error) {
	f.updateCounselingArg = arg
	return db.CounselingSession{
		ID:             arg.ID,
		StudentID:      arg.StudentID,
		Topic:          arg.Topic,
		Status:         arg.Status,
		IsConfidential: arg.IsConfidential,
	}, f.mutationErr
}

func (f *fakeKesiswaanService) DeleteCounselingSession(ctx context.Context, id pgtype.UUID) error {
	f.deleteCounselingID = id
	return f.deleteErr
}

func (f *fakeKesiswaanService) ListStudentTransfers(ctx context.Context, search, transferType, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentTransfersRow, error) {
	f.listTransfersSearch = search
	f.listTransfersType = transferType
	f.listTransfersStudent = studentID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListStudentTransfersRow{{ID: handlerTestUUID(99), TransferType: "out"}}, nil
}

func (f *fakeKesiswaanService) CreateStudentTransfer(ctx context.Context, arg db.CreateStudentTransferParams) (db.StudentTransfer, error) {
	f.createTransferArg = arg
	return db.StudentTransfer{ID: handlerTestUUID(99), StudentID: arg.StudentID, TransferType: arg.TransferType}, f.mutationErr
}

func TestKesiswaanSuccessHandlersForwardPayloads(t *testing.T) {
	studentID := handlerTestUUID(101)
	categoryID := handlerTestUUID(102)
	extracurricularID := handlerTestUUID(103)
	supervisorID := handlerTestUUID(104)
	violationID := handlerTestUUID(105)
	achievementID := handlerTestUUID(106)
	memberID := handlerTestUUID(107)
	counselingID := handlerTestUUID(108)
	h := &Kesiswaan{svc: &fakeKesiswaanService{}}
	svc := h.svc.(*fakeKesiswaanService)

	rec := httptest.NewRecorder()
	h.Stats(rec, adminRequest(http.MethodGet, "/api/kesiswaan/stats", ""))
	if rec.Code != http.StatusOK || svc.statsTeacher.Valid {
		t.Fatalf("Stats() status/teacher = %d/%v, want 200/admin-wide", rec.Code, svc.statsTeacher)
	}

	rec = httptest.NewRecorder()
	h.ClassOptions(rec, adminRequest(http.MethodGet, "/api/kesiswaan/classes", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ClassOptions() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListStudents(rec, adminRequest(http.MethodGet, "/api/kesiswaan/students?search=andi&status=active&class_id="+studentID.String(), ""))
	if rec.Code != http.StatusOK || svc.listStudentsSearch != "andi" || svc.listStudentsStatus != "active" || svc.listStudentsClassID != studentID.String() {
		t.Fatalf("ListStudents() status/args = %d/%q/%q/%q", rec.Code, svc.listStudentsSearch, svc.listStudentsStatus, svc.listStudentsClassID)
	}

	profileBody := `{"nik":"1234567890123456","tempat_lahir":"Kolaka","tanggal_lahir":"2010-05-01","alamat":"Lasoani","agama":"Islam","anak_ke":2,"phone":"0812","parent_name":"Ortu","parent_phone":"0813"}`
	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/students/"+studentID.String()+"/profile", profileBody), "id", studentID.String())
	h.UpdateStudentProfile(rec, req)
	if rec.Code != http.StatusOK || svc.updateStudentProfileArg.ID != studentID || !svc.updateStudentProfileArg.TanggalLahir.Valid || svc.updateStudentProfileArg.AnakKe.Int32 != 2 {
		t.Fatalf("UpdateStudentProfile() status/arg = %d/%+v", rec.Code, svc.updateStudentProfileArg)
	}

	rec = httptest.NewRecorder()
	h.ListCategories(rec, adminRequest(http.MethodGet, "/api/kesiswaan/violation-categories?search=telat", ""))
	if rec.Code != http.StatusOK || svc.categorySearch != "telat" {
		t.Fatalf("ListCategories() status/search = %d/%q", rec.Code, svc.categorySearch)
	}

	categoryBody := `{"code":"TL","name":"Terlambat","point":5,"severity":"ringan","description":"Datang terlambat","is_active":false}`
	rec = httptest.NewRecorder()
	h.CreateCategory(rec, adminRequest(http.MethodPost, "/api/kesiswaan/violation-categories", categoryBody))
	if rec.Code != http.StatusCreated || svc.createCategoryArg.Code != "TL" || svc.createCategoryArg.IsActive {
		t.Fatalf("CreateCategory() status/arg = %d/%+v", rec.Code, svc.createCategoryArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/violation-categories/"+categoryID.String(), categoryBody), "id", categoryID.String())
	h.UpdateCategory(rec, req)
	if rec.Code != http.StatusOK || svc.updateCategoryArg.ID != categoryID || svc.updateCategoryArg.IsActive {
		t.Fatalf("UpdateCategory() status/arg = %d/%+v", rec.Code, svc.updateCategoryArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/kesiswaan/violation-categories/"+categoryID.String(), ""), "id", categoryID.String())
	h.DeleteCategory(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteCategoryID != categoryID {
		t.Fatalf("DeleteCategory() status/id = %d/%v", rec.Code, svc.deleteCategoryID)
	}

	rec = httptest.NewRecorder()
	h.ListViolations(rec, adminRequest(http.MethodGet, "/api/kesiswaan/violations?search=telat&status=open&student_id="+studentID.String(), ""))
	if rec.Code != http.StatusOK || svc.listViolationsStatus != "open" || svc.listViolationsStudentID != studentID.String() {
		t.Fatalf("ListViolations() status/args = %d/%q/%q", rec.Code, svc.listViolationsStatus, svc.listViolationsStudentID)
	}

	violationBody := `{"student_id":"` + studentID.String() + `","category_id":"` + categoryID.String() + `","incident_date":"2026-05-01","points":5,"description":"Terlambat","action_taken":"Pembinaan","status":"open"}`
	rec = httptest.NewRecorder()
	h.CreateViolation(rec, adminRequest(http.MethodPost, "/api/kesiswaan/violations", violationBody))
	if rec.Code != http.StatusCreated || svc.createViolationArg.StudentID != studentID || svc.createViolationArg.CategoryID != categoryID || svc.createViolationArg.RecordedByUserID != handlerTestUUID(1) {
		t.Fatalf("CreateViolation() status/arg = %d/%+v", rec.Code, svc.createViolationArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/violations/"+violationID.String(), violationBody), "id", violationID.String())
	h.UpdateViolation(rec, req)
	if rec.Code != http.StatusOK || svc.updateViolationArg.ID != violationID || svc.updateViolationArg.StudentID != studentID {
		t.Fatalf("UpdateViolation() status/arg = %d/%+v", rec.Code, svc.updateViolationArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/kesiswaan/violations/"+violationID.String(), ""), "id", violationID.String())
	h.DeleteViolation(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteViolationID != violationID {
		t.Fatalf("DeleteViolation() status/id = %d/%v", rec.Code, svc.deleteViolationID)
	}

	rec = httptest.NewRecorder()
	h.ListAchievements(rec, adminRequest(http.MethodGet, "/api/kesiswaan/achievements?search=juara&level=school&student_id="+studentID.String(), ""))
	if rec.Code != http.StatusOK || svc.listAchievementsLevel != "school" || svc.listAchievementsStudentID != studentID.String() {
		t.Fatalf("ListAchievements() status/args = %d/%q/%q", rec.Code, svc.listAchievementsLevel, svc.listAchievementsStudentID)
	}

	achievementBody := `{"student_id":"` + studentID.String() + `","achievement_date":"2026-05-02","title":"Juara","level":"school","category":"Akademik","organizer":"Kemenag","description":"Juara 1","document_url":"/bukti"}`
	rec = httptest.NewRecorder()
	h.CreateAchievement(rec, adminRequest(http.MethodPost, "/api/kesiswaan/achievements", achievementBody))
	if rec.Code != http.StatusCreated || svc.createAchievementArg.StudentID != studentID || svc.createAchievementArg.Title != "Juara" || svc.createAchievementArg.RecordedByUserID != handlerTestUUID(1) {
		t.Fatalf("CreateAchievement() status/arg = %d/%+v", rec.Code, svc.createAchievementArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/achievements/"+achievementID.String(), achievementBody), "id", achievementID.String())
	h.UpdateAchievement(rec, req)
	if rec.Code != http.StatusOK || svc.updateAchievementArg.ID != achievementID || svc.updateAchievementArg.Title != "Juara" {
		t.Fatalf("UpdateAchievement() status/arg = %d/%+v", rec.Code, svc.updateAchievementArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/kesiswaan/achievements/"+achievementID.String(), ""), "id", achievementID.String())
	h.DeleteAchievement(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteAchievementID != achievementID {
		t.Fatalf("DeleteAchievement() status/id = %d/%v", rec.Code, svc.deleteAchievementID)
	}

	rec = httptest.NewRecorder()
	h.ListExtracurriculars(rec, adminRequest(http.MethodGet, "/api/kesiswaan/extracurriculars?search=pramuka&active_only=true", ""))
	if rec.Code != http.StatusOK || svc.listExtracurricularsSearch != "pramuka" || !svc.listExtracurricularsActiveOnly {
		t.Fatalf("ListExtracurriculars() status/args = %d/%q/%v", rec.Code, svc.listExtracurricularsSearch, svc.listExtracurricularsActiveOnly)
	}

	extracurricularBody := `{"code":"PRM","name":"Pramuka","category":"Wajib","description":"Latihan","supervisor_employee_id":"` + supervisorID.String() + `","schedule_text":"Jumat","is_active":false}`
	rec = httptest.NewRecorder()
	h.CreateExtracurricular(rec, adminRequest(http.MethodPost, "/api/kesiswaan/extracurriculars", extracurricularBody))
	if rec.Code != http.StatusCreated || svc.createExtracurricularArg.Code != "PRM" || svc.createExtracurricularArg.SupervisorEmployeeID != supervisorID || svc.createExtracurricularArg.IsActive {
		t.Fatalf("CreateExtracurricular() status/arg = %d/%+v", rec.Code, svc.createExtracurricularArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/extracurriculars/"+extracurricularID.String(), extracurricularBody), "id", extracurricularID.String())
	h.UpdateExtracurricular(rec, req)
	if rec.Code != http.StatusOK || svc.updateExtracurricularArg.ID != extracurricularID || svc.updateExtracurricularArg.SupervisorEmployeeID != supervisorID {
		t.Fatalf("UpdateExtracurricular() status/arg = %d/%+v", rec.Code, svc.updateExtracurricularArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/kesiswaan/extracurriculars/"+extracurricularID.String(), ""), "id", extracurricularID.String())
	h.DeleteExtracurricular(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteExtracurricularID != extracurricularID {
		t.Fatalf("DeleteExtracurricular() status/id = %d/%v", rec.Code, svc.deleteExtracurricularID)
	}

	rec = httptest.NewRecorder()
	h.ListExtracurricularMembers(rec, adminRequest(http.MethodGet, "/api/kesiswaan/extracurricular-members?search=andi&status=active&extracurricular_id="+extracurricularID.String()+"&student_id="+studentID.String(), ""))
	if rec.Code != http.StatusOK || svc.listMembersStatus != "active" || svc.listMembersExtracurricularID != extracurricularID.String() || svc.listMembersStudentID != studentID.String() {
		t.Fatalf("ListExtracurricularMembers() status/args = %d/%q/%q/%q", rec.Code, svc.listMembersStatus, svc.listMembersExtracurricularID, svc.listMembersStudentID)
	}

	memberBody := `{"extracurricular_id":"` + extracurricularID.String() + `","student_id":"` + studentID.String() + `","joined_at":"2026-05-03","role":"member","status":"active","notes":"aktif"}`
	rec = httptest.NewRecorder()
	h.CreateExtracurricularMember(rec, adminRequest(http.MethodPost, "/api/kesiswaan/extracurricular-members", memberBody))
	if rec.Code != http.StatusCreated || svc.createMemberArg.StudentID != studentID || svc.createMemberArg.ExtracurricularID != extracurricularID || svc.createMemberArg.RecordedByUserID != handlerTestUUID(1) {
		t.Fatalf("CreateExtracurricularMember() status/arg = %d/%+v", rec.Code, svc.createMemberArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/extracurricular-members/"+memberID.String(), memberBody), "id", memberID.String())
	h.UpdateExtracurricularMember(rec, req)
	if rec.Code != http.StatusOK || svc.updateMemberArg.ID != memberID || svc.updateMemberArg.StudentID != studentID {
		t.Fatalf("UpdateExtracurricularMember() status/arg = %d/%+v", rec.Code, svc.updateMemberArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/kesiswaan/extracurricular-members/"+memberID.String(), ""), "id", memberID.String())
	h.DeleteExtracurricularMember(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteMemberID != memberID {
		t.Fatalf("DeleteExtracurricularMember() status/id = %d/%v", rec.Code, svc.deleteMemberID)
	}

	rec = httptest.NewRecorder()
	h.ListCounselingSessions(rec, adminRequest(http.MethodGet, "/api/kesiswaan/counseling-sessions?search=belajar&status=open&student_id="+studentID.String(), ""))
	if rec.Code != http.StatusOK || svc.listCounselingStatus != "open" || svc.listCounselingStudentID != studentID.String() || !svc.listCounselingConfidential {
		t.Fatalf("ListCounselingSessions() status/args = %d/%q/%q/%v", rec.Code, svc.listCounselingStatus, svc.listCounselingStudentID, svc.listCounselingConfidential)
	}

	counselingBody := `{"student_id":"` + studentID.String() + `","session_date":"2026-05-04","topic":"Belajar","summary":"Perlu pendampingan","follow_up":"Jadwal ulang","status":"open","is_confidential":true}`
	rec = httptest.NewRecorder()
	h.CreateCounselingSession(rec, adminRequest(http.MethodPost, "/api/kesiswaan/counseling-sessions", counselingBody))
	if rec.Code != http.StatusCreated || svc.createCounselingArg.StudentID != studentID || svc.createCounselingArg.Topic != "Belajar" || !svc.createCounselingArg.IsConfidential || svc.createCounselingArg.RecordedByUserID != handlerTestUUID(1) {
		t.Fatalf("CreateCounselingSession() status/arg = %d/%+v", rec.Code, svc.createCounselingArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/kesiswaan/counseling-sessions/"+counselingID.String(), counselingBody), "id", counselingID.String())
	h.UpdateCounselingSession(rec, req)
	if rec.Code != http.StatusOK || svc.updateCounselingArg.ID != counselingID || svc.updateCounselingArg.StudentID != studentID {
		t.Fatalf("UpdateCounselingSession() status/arg = %d/%+v", rec.Code, svc.updateCounselingArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/kesiswaan/counseling-sessions/"+counselingID.String(), ""), "id", counselingID.String())
	h.DeleteCounselingSession(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteCounselingID != counselingID {
		t.Fatalf("DeleteCounselingSession() status/id = %d/%v", rec.Code, svc.deleteCounselingID)
	}

	rec = httptest.NewRecorder()
	h.ListStudentTransfers(rec, adminRequest(http.MethodGet, "/api/kesiswaan/student-transfers?search=pindah&transfer_type=out&student_id="+studentID.String(), ""))
	if rec.Code != http.StatusOK || svc.listTransfersType != "out" || svc.listTransfersStudent != studentID.String() {
		t.Fatalf("ListStudentTransfers() status/args = %d/%q/%q", rec.Code, svc.listTransfersType, svc.listTransfersStudent)
	}

	transferBody := `{"student_id":"` + studentID.String() + `","transfer_date":"2026-05-05","transfer_type":"out","previous_school":"MTsN 2","destination_school":"MTsN lain","reason":"Pindah domisili","document_ref":"SK-1","notes":"lengkap"}`
	rec = httptest.NewRecorder()
	h.CreateStudentTransfer(rec, adminRequest(http.MethodPost, "/api/kesiswaan/student-transfers", transferBody))
	if rec.Code != http.StatusCreated || svc.createTransferArg.StudentID != studentID || svc.createTransferArg.TransferType != "out" || svc.createTransferArg.RecordedByUserID != handlerTestUUID(1) {
		t.Fatalf("CreateStudentTransfer() status/arg = %d/%+v", rec.Code, svc.createTransferArg)
	}
}

func TestKesiswaanPhotoHandlersForwardUploadAndServeFile(t *testing.T) {
	studentID := handlerTestUUID(120)
	svc := &fakeKesiswaanService{}
	h := &Kesiswaan{svc: svc}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "foto.jpg")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00")); err != nil {
		t.Fatalf("multipart write error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close error = %v", err)
	}
	req := withRouteParam(adminRequest(http.MethodPost, "/api/kesiswaan/students/"+studentID.String()+"/photo", body.String()), "id", studentID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.UploadStudentPhoto(rec, req)

	if rec.Code != http.StatusOK || svc.savePhotoInput.StudentID != studentID || svc.savePhotoInput.OriginalName != "foto.jpg" || svc.savePhotoInput.FileSize <= 0 {
		t.Fatalf("UploadStudentPhoto() status/input = %d/%+v", rec.Code, svc.savePhotoInput)
	}

	path := t.TempDir() + "/foto.jpg"
	if err := os.WriteFile(path, []byte("image-bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	svc.canReadPhotoValue = true
	svc.photoPathValue = path
	svc.photoPathOK = true
	req = withRouteParam(adminRequest(http.MethodGet, "/api/kesiswaan/student-photos/foto.jpg", ""), "filename", "foto.jpg")
	rec = httptest.NewRecorder()

	h.StudentPhotoFile(rec, req)

	if rec.Code != http.StatusOK || svc.canReadPhotoFilename != "foto.jpg" || svc.photoPathFilename != "foto.jpg" || rec.Body.String() != "image-bytes" {
		t.Fatalf("StudentPhotoFile() status/file/body = %d/%q/%q/%q", rec.Code, svc.canReadPhotoFilename, svc.photoPathFilename, rec.Body.String())
	}
}

func TestKesiswaanMutationHandlersWriteAuditEvents(t *testing.T) {
	studentID := handlerTestUUID(130)
	categoryID := handlerTestUUID(131)
	violationID := handlerTestUUID(132)
	counselingID := handlerTestUUID(133)
	audit := &fakeCbtSessionAuditWriter{}
	h := &Kesiswaan{svc: &fakeKesiswaanService{}, audit: audit}
	auditedAdminRequest := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, bytes.NewBufferString(body)), jwt.MapClaims{
			"roles": []any{"admin"},
			"uid":   "01000000-0000-0000-0000-000000000000",
			"sub":   "01000000-0000-0000-0000-000000000000",
			"usr":   "admin",
			"ssid":  "sess-kesiswaan-1",
		})
	}

	profileBody := `{"nik":"1234567890123456","tempat_lahir":"Kolaka","tanggal_lahir":"2010-05-01","alamat":"Lasoani","agama":"Islam","anak_ke":2,"phone":"0812","parent_name":"Ortu","parent_phone":"0813"}`
	rec := httptest.NewRecorder()
	req := withRouteParam(auditedAdminRequest(http.MethodPut, "/api/kesiswaan/students/"+studentID.String()+"/profile", profileBody), "id", studentID.String())
	h.UpdateStudentProfile(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateStudentProfile() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	violationBody := `{"student_id":"` + studentID.String() + `","category_id":"` + categoryID.String() + `","incident_date":"2026-05-01","points":5,"description":"Terlambat","action_taken":"Pembinaan","status":"open"}`
	rec = httptest.NewRecorder()
	h.CreateViolation(rec, auditedAdminRequest(http.MethodPost, "/api/kesiswaan/violations", violationBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateViolation() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(auditedAdminRequest(http.MethodPut, "/api/kesiswaan/violations/"+violationID.String(), violationBody), "id", violationID.String())
	h.UpdateViolation(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateViolation() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(auditedAdminRequest(http.MethodDelete, "/api/kesiswaan/violations/"+violationID.String(), ""), "id", violationID.String())
	h.DeleteViolation(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteViolation() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	counselingBody := `{"student_id":"` + studentID.String() + `","session_date":"2026-05-04","topic":"Belajar","summary":"Perlu pendampingan","follow_up":"Jadwal ulang","status":"open","is_confidential":true}`
	rec = httptest.NewRecorder()
	h.CreateCounselingSession(rec, auditedAdminRequest(http.MethodPost, "/api/kesiswaan/counseling-sessions", counselingBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateCounselingSession() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(auditedAdminRequest(http.MethodPut, "/api/kesiswaan/counseling-sessions/"+counselingID.String(), counselingBody), "id", counselingID.String())
	h.UpdateCounselingSession(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateCounselingSession() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(auditedAdminRequest(http.MethodDelete, "/api/kesiswaan/counseling-sessions/"+counselingID.String(), ""), "id", counselingID.String())
	h.DeleteCounselingSession(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteCounselingSession() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	req = withClaims(newKesiswaanMultipartAdminRequest(t, studentID, true), jwt.MapClaims{
		"roles": []any{"admin"},
		"uid":   "01000000-0000-0000-0000-000000000000",
		"sub":   "01000000-0000-0000-0000-000000000000",
		"usr":   "admin",
		"ssid":  "sess-kesiswaan-1",
	})
	rec = httptest.NewRecorder()
	h.UploadStudentPhoto(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UploadStudentPhoto() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	transferBody := `{"student_id":"` + studentID.String() + `","transfer_date":"2026-05-05","transfer_type":"out","previous_school":"MTsN 2","destination_school":"MTsN lain","reason":"Pindah domisili","document_ref":"SK-1","notes":"lengkap"}`
	rec = httptest.NewRecorder()
	h.CreateStudentTransfer(rec, auditedAdminRequest(http.MethodPost, "/api/kesiswaan/student-transfers", transferBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateStudentTransfer() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	if len(audit.entries) != 9 {
		t.Fatalf("audit entries = %d, want 9", len(audit.entries))
	}

	checks := []struct {
		index      int
		action     string
		entityType string
		entityID   string
	}{
		{0, "KESISWAAN_STUDENT_PROFILE_UPDATE", "student", studentID.String()},
		{1, "KESISWAAN_VIOLATION_CREATE", "student_violation", handlerTestUUID(94).String()},
		{2, "KESISWAAN_VIOLATION_UPDATE", "student_violation", violationID.String()},
		{3, "KESISWAAN_VIOLATION_DELETE", "student_violation", violationID.String()},
		{4, "KESISWAAN_COUNSELING_CREATE", "counseling_session", handlerTestUUID(98).String()},
		{5, "KESISWAAN_COUNSELING_UPDATE", "counseling_session", counselingID.String()},
		{6, "KESISWAAN_COUNSELING_DELETE", "counseling_session", counselingID.String()},
		{7, "KESISWAAN_STUDENT_PHOTO_UPLOAD", "student", studentID.String()},
		{8, "KESISWAAN_STUDENT_TRANSFER_CREATE", "student_transfer", handlerTestUUID(99).String()},
	}
	for _, check := range checks {
		got := audit.entries[check.index]
		if got.Action != check.action || got.EntityType != check.entityType || got.EntityID != check.entityID {
			t.Fatalf("audit[%d] = %+v, want action/type/id %q/%q/%q", check.index, got, check.action, check.entityType, check.entityID)
		}
	}

	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["username"] != "admin" || meta["student_id"] != studentID.String() {
		t.Fatalf("profile audit metadata = %+v, want username/student_id", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[1].Metadata)
	if meta["status"] != "open" || meta["student_id"] != studentID.String() {
		t.Fatalf("violation create metadata = %+v, want open/student_id", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[4].Metadata)
	if meta["is_confidential"] != true || meta["student_id"] != studentID.String() {
		t.Fatalf("counseling create metadata = %+v, want confidential/student_id", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[6].Metadata)
	if meta["deleted_by"] != "admin" {
		t.Fatalf("counseling delete metadata = %+v, want deleted_by admin", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[7].Metadata)
	if meta["photo_url"] != "/api/kesiswaan/student-photos/foto.jpg" {
		t.Fatalf("photo upload metadata = %+v, want photo_url", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[8].Metadata)
	if meta["transfer_type"] != "out" || meta["student_id"] != studentID.String() {
		t.Fatalf("transfer create metadata = %+v, want transfer_type/student_id", meta)
	}
}

func TestKesiswaanValidationAndServiceErrors(t *testing.T) {
	serviceErr := errors.New("service down")
	studentID := handlerTestUUID(121)
	categoryID := handlerTestUUID(122)
	extracurricularID := handlerTestUUID(123)
	memberID := handlerTestUUID(124)
	counselingID := handlerTestUUID(125)
	violationID := handlerTestUUID(126)
	achievementID := handlerTestUUID(127)
	supervisorID := handlerTestUUID(128)

	profileBody := `{"nik":"1234567890123456","tempat_lahir":"Kolaka","tanggal_lahir":"2010-05-01","alamat":"Lasoani","agama":"Islam","anak_ke":2}`
	categoryBody := `{"code":"TL","name":"Terlambat","point":5,"severity":"ringan","description":"Datang terlambat","is_active":true}`
	violationBody := `{"student_id":"` + studentID.String() + `","category_id":"` + categoryID.String() + `","incident_date":"2026-05-01","points":5,"description":"Terlambat","action_taken":"Pembinaan","status":"open"}`
	achievementBody := `{"student_id":"` + studentID.String() + `","achievement_date":"2026-05-02","title":"Juara","level":"school","category":"Akademik","organizer":"Kemenag"}`
	extracurricularBody := `{"code":"PRM","name":"Pramuka","category":"Wajib","supervisor_employee_id":"` + supervisorID.String() + `","schedule_text":"Jumat","is_active":true}`
	memberBody := `{"extracurricular_id":"` + extracurricularID.String() + `","student_id":"` + studentID.String() + `","joined_at":"2026-05-03","role":"member","status":"active"}`
	counselingBody := `{"student_id":"` + studentID.String() + `","session_date":"2026-05-04","topic":"Belajar","summary":"Perlu pendampingan","status":"open"}`
	transferBody := `{"student_id":"` + studentID.String() + `","transfer_date":"2026-05-05","transfer_type":"out","previous_school":"MTsN 2","destination_school":"MTsN lain"}`

	tests := []struct {
		name       string
		svc        *fakeKesiswaanService
		method     string
		target     string
		body       string
		paramKey   string
		paramValue string
		call       func(*Kesiswaan, http.ResponseWriter, *http.Request)
		wantStatus int
	}{
		{name: "stats service error", svc: &fakeKesiswaanService{statsErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/stats", call: (*Kesiswaan).Stats, wantStatus: http.StatusInternalServerError},
		{name: "class options service error", svc: &fakeKesiswaanService{classOptionsErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/classes", call: (*Kesiswaan).ClassOptions, wantStatus: http.StatusInternalServerError},
		{name: "list students service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/students", call: (*Kesiswaan).ListStudents, wantStatus: http.StatusBadRequest},
		{name: "update profile bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/students/bad/profile", body: profileBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateStudentProfile, wantStatus: http.StatusBadRequest},
		{name: "update profile invalid json", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/students/" + studentID.String() + "/profile", body: `{`, paramKey: "id", paramValue: studentID.String(), call: (*Kesiswaan).UpdateStudentProfile, wantStatus: http.StatusBadRequest},
		{name: "update profile bad date", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/students/" + studentID.String() + "/profile", body: `{"tanggal_lahir":"bad"}`, paramKey: "id", paramValue: studentID.String(), call: (*Kesiswaan).UpdateStudentProfile, wantStatus: http.StatusBadRequest},
		{name: "update profile service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/students/" + studentID.String() + "/profile", body: profileBody, paramKey: "id", paramValue: studentID.String(), call: (*Kesiswaan).UpdateStudentProfile, wantStatus: http.StatusBadRequest},
		{name: "list categories service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/violation-categories", call: (*Kesiswaan).ListCategories, wantStatus: http.StatusInternalServerError},
		{name: "create category invalid json", svc: &fakeKesiswaanService{}, method: http.MethodPost, target: "/api/kesiswaan/violation-categories", body: `{`, call: (*Kesiswaan).CreateCategory, wantStatus: http.StatusBadRequest},
		{name: "create category service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/violation-categories", body: categoryBody, call: (*Kesiswaan).CreateCategory, wantStatus: http.StatusBadRequest},
		{name: "update category bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/violation-categories/bad", body: categoryBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateCategory, wantStatus: http.StatusBadRequest},
		{name: "update category service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/violation-categories/" + categoryID.String(), body: categoryBody, paramKey: "id", paramValue: categoryID.String(), call: (*Kesiswaan).UpdateCategory, wantStatus: http.StatusBadRequest},
		{name: "delete category bad id", svc: &fakeKesiswaanService{}, method: http.MethodDelete, target: "/api/kesiswaan/violation-categories/bad", paramKey: "id", paramValue: "bad", call: (*Kesiswaan).DeleteCategory, wantStatus: http.StatusBadRequest},
		{name: "delete category service error", svc: &fakeKesiswaanService{deleteErr: serviceErr}, method: http.MethodDelete, target: "/api/kesiswaan/violation-categories/" + categoryID.String(), paramKey: "id", paramValue: categoryID.String(), call: (*Kesiswaan).DeleteCategory, wantStatus: http.StatusBadRequest},
		{name: "list violations service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/violations", call: (*Kesiswaan).ListViolations, wantStatus: http.StatusBadRequest},
		{name: "create violation invalid json", svc: &fakeKesiswaanService{}, method: http.MethodPost, target: "/api/kesiswaan/violations", body: `{`, call: (*Kesiswaan).CreateViolation, wantStatus: http.StatusBadRequest},
		{name: "create violation service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/violations", body: violationBody, call: (*Kesiswaan).CreateViolation, wantStatus: http.StatusBadRequest},
		{name: "update violation bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/violations/bad", body: violationBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateViolation, wantStatus: http.StatusBadRequest},
		{name: "update violation service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/violations/" + violationID.String(), body: violationBody, paramKey: "id", paramValue: violationID.String(), call: (*Kesiswaan).UpdateViolation, wantStatus: http.StatusBadRequest},
		{name: "list achievements service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/achievements", call: (*Kesiswaan).ListAchievements, wantStatus: http.StatusBadRequest},
		{name: "create achievement service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/achievements", body: achievementBody, call: (*Kesiswaan).CreateAchievement, wantStatus: http.StatusBadRequest},
		{name: "update achievement bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/achievements/bad", body: achievementBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateAchievement, wantStatus: http.StatusBadRequest},
		{name: "update achievement service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/achievements/" + achievementID.String(), body: achievementBody, paramKey: "id", paramValue: achievementID.String(), call: (*Kesiswaan).UpdateAchievement, wantStatus: http.StatusBadRequest},
		{name: "list extracurriculars service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/extracurriculars", call: (*Kesiswaan).ListExtracurriculars, wantStatus: http.StatusBadRequest},
		{name: "create extracurricular bad supervisor", svc: &fakeKesiswaanService{}, method: http.MethodPost, target: "/api/kesiswaan/extracurriculars", body: `{"supervisor_employee_id":"bad"}`, call: (*Kesiswaan).CreateExtracurricular, wantStatus: http.StatusBadRequest},
		{name: "create extracurricular service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/extracurriculars", body: extracurricularBody, call: (*Kesiswaan).CreateExtracurricular, wantStatus: http.StatusBadRequest},
		{name: "update extracurricular bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/extracurriculars/bad", body: extracurricularBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateExtracurricular, wantStatus: http.StatusBadRequest},
		{name: "update extracurricular service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/extracurriculars/" + extracurricularID.String(), body: extracurricularBody, paramKey: "id", paramValue: extracurricularID.String(), call: (*Kesiswaan).UpdateExtracurricular, wantStatus: http.StatusBadRequest},
		{name: "list members service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/extracurricular-members", call: (*Kesiswaan).ListExtracurricularMembers, wantStatus: http.StatusBadRequest},
		{name: "create member service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/extracurricular-members", body: memberBody, call: (*Kesiswaan).CreateExtracurricularMember, wantStatus: http.StatusBadRequest},
		{name: "update member bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/extracurricular-members/bad", body: memberBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateExtracurricularMember, wantStatus: http.StatusBadRequest},
		{name: "update member service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/extracurricular-members/" + memberID.String(), body: memberBody, paramKey: "id", paramValue: memberID.String(), call: (*Kesiswaan).UpdateExtracurricularMember, wantStatus: http.StatusBadRequest},
		{name: "list counseling service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/counseling-sessions", call: (*Kesiswaan).ListCounselingSessions, wantStatus: http.StatusBadRequest},
		{name: "create counseling service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/counseling-sessions", body: counselingBody, call: (*Kesiswaan).CreateCounselingSession, wantStatus: http.StatusBadRequest},
		{name: "update counseling bad id", svc: &fakeKesiswaanService{}, method: http.MethodPut, target: "/api/kesiswaan/counseling-sessions/bad", body: counselingBody, paramKey: "id", paramValue: "bad", call: (*Kesiswaan).UpdateCounselingSession, wantStatus: http.StatusBadRequest},
		{name: "update counseling service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPut, target: "/api/kesiswaan/counseling-sessions/" + counselingID.String(), body: counselingBody, paramKey: "id", paramValue: counselingID.String(), call: (*Kesiswaan).UpdateCounselingSession, wantStatus: http.StatusBadRequest},
		{name: "list transfers service error", svc: &fakeKesiswaanService{listErr: serviceErr}, method: http.MethodGet, target: "/api/kesiswaan/student-transfers", call: (*Kesiswaan).ListStudentTransfers, wantStatus: http.StatusBadRequest},
		{name: "create transfer service error", svc: &fakeKesiswaanService{mutationErr: serviceErr}, method: http.MethodPost, target: "/api/kesiswaan/student-transfers", body: transferBody, call: (*Kesiswaan).CreateStudentTransfer, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Kesiswaan{svc: tt.svc}
			req := adminRequest(tt.method, tt.target, tt.body)
			if tt.paramKey != "" {
				req = withRouteParam(req, tt.paramKey, tt.paramValue)
			}
			rec := httptest.NewRecorder()

			tt.call(h, rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestKesiswaanPhotoValidationAndServiceErrors(t *testing.T) {
	serviceErr := errors.New("service down")
	studentID := handlerTestUUID(129)
	h := &Kesiswaan{svc: &fakeKesiswaanService{}}

	req := withRouteParam(adminRequest(http.MethodPost, "/api/kesiswaan/students/bad/photo", ""), "id", "bad")
	rec := httptest.NewRecorder()
	h.UploadStudentPhoto(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadStudentPhoto(bad id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	req = withRouteParam(adminRequest(http.MethodPost, "/api/kesiswaan/students/"+studentID.String()+"/photo", "not multipart"), "id", studentID.String())
	req.Header.Set("Content-Type", "multipart/form-data; boundary=broken")
	rec = httptest.NewRecorder()
	h.UploadStudentPhoto(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadStudentPhoto(invalid multipart) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	req = newKesiswaanMultipartAdminRequest(t, studentID, false)
	rec = httptest.NewRecorder()
	h.UploadStudentPhoto(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadStudentPhoto(missing file) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	h = &Kesiswaan{svc: &fakeKesiswaanService{photoErr: serviceErr}}
	req = newKesiswaanMultipartAdminRequest(t, studentID, true)
	rec = httptest.NewRecorder()
	h.UploadStudentPhoto(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadStudentPhoto(service error) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	h = &Kesiswaan{svc: &fakeKesiswaanService{}}
	req = newOversizedKesiswaanPhotoRequest(t, studentID)
	rec = httptest.NewRecorder()
	h.UploadStudentPhoto(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadStudentPhoto(oversized) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	photoCases := []struct {
		name       string
		svc        *fakeKesiswaanService
		wantStatus int
	}{
		{name: "access error", svc: &fakeKesiswaanService{canReadPhotoErr: serviceErr}, wantStatus: http.StatusInternalServerError},
		{name: "access denied", svc: &fakeKesiswaanService{canReadPhotoValue: false}, wantStatus: http.StatusNotFound},
		{name: "path missing", svc: &fakeKesiswaanService{canReadPhotoValue: true, photoPathOK: false}, wantStatus: http.StatusNotFound},
	}
	for _, tt := range photoCases {
		t.Run(tt.name, func(t *testing.T) {
			h := &Kesiswaan{svc: tt.svc}
			req := withRouteParam(adminRequest(http.MethodGet, "/api/kesiswaan/student-photos/foto.jpg", ""), "filename", "foto.jpg")
			rec := httptest.NewRecorder()

			h.StudentPhotoFile(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("StudentPhotoFile() status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func newKesiswaanMultipartAdminRequest(t *testing.T, studentID pgtype.UUID, includeFile bool) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if includeFile {
		part, err := writer.CreateFormFile("file", "foto.jpg")
		if err != nil {
			t.Fatalf("CreateFormFile() error = %v", err)
		}
		if _, err := part.Write([]byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00")); err != nil {
			t.Fatalf("multipart write error = %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close error = %v", err)
	}
	req := withRouteParam(adminRequest(http.MethodPost, "/api/kesiswaan/students/"+studentID.String()+"/photo", body.String()), "id", studentID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func newOversizedKesiswaanPhotoRequest(t *testing.T, studentID pgtype.UUID) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "foto.jpg")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(bytes.Repeat([]byte("x"), (3<<20)+1)); err != nil {
		t.Fatalf("multipart write error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close error = %v", err)
	}
	req := withRouteParam(adminRequest(http.MethodPost, "/api/kesiswaan/students/"+studentID.String()+"/photo", body.String()), "id", studentID.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
