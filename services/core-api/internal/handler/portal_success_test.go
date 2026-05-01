package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakePortalService struct {
	studentOverviewID   pgtype.UUID
	studentTimetableID  pgtype.UUID
	teacherTimetableID  pgtype.UUID
	parentOverviewID    pgtype.UUID
	parentTimetableID   pgtype.UUID
	studentOverviewErr  error
	studentTimetableErr error
	teacherTimetableErr error
	parentOverviewErr   error
	parentTimetableErr  error
}

func (f *fakePortalService) StudentOverview(ctx context.Context, studentID pgtype.UUID) (db.GetStudentByIDRow, []db.ListStudentParentsRow, []db.ListStudentExamSessionsRow, error) {
	f.studentOverviewID = studentID
	return db.GetStudentByIDRow{ID: studentID, Nama: "Siswa A"}, []db.ListStudentParentsRow{}, []db.ListStudentExamSessionsRow{}, f.studentOverviewErr
}

func (f *fakePortalService) StudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	f.studentTimetableID = studentID
	return []db.ListStudentTimetableRow{}, f.studentTimetableErr
}

func (f *fakePortalService) TeacherTimetable(ctx context.Context, employeeID pgtype.UUID) ([]db.ListTeacherTimetableRow, error) {
	f.teacherTimetableID = employeeID
	return []db.ListTeacherTimetableRow{}, f.teacherTimetableErr
}

func (f *fakePortalService) ParentChildrenTimetable(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenTimetableRow, error) {
	f.parentTimetableID = parentID
	return []db.ListParentChildrenTimetableRow{}, f.parentTimetableErr
}

func (f *fakePortalService) ParentOverview(ctx context.Context, parentID pgtype.UUID) (db.Parent, []db.ListParentChildrenRow, error) {
	f.parentOverviewID = parentID
	return db.Parent{ID: parentID, Nama: "Orang Tua"}, []db.ListParentChildrenRow{}, f.parentOverviewErr
}

func portalClaimsRequest(role, claimKey string, id pgtype.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/portal", nil)
	return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles":  []any{role},
		claimKey: id.String(),
	}))
}

func TestPortalSuccessHandlersForwardScopedIDs(t *testing.T) {
	studentID := handlerTestUUID(200)
	employeeID := handlerTestUUID(201)
	parentID := handlerTestUUID(202)
	svc := &fakePortalService{}
	h := &Portal{svc: svc}

	rec := httptest.NewRecorder()
	h.StudentMe(rec, portalClaimsRequest("siswa", "sid", studentID))
	if rec.Code != http.StatusOK || svc.studentOverviewID != studentID || svc.studentTimetableID != studentID {
		t.Fatalf("StudentMe() status/ids = %d/%v/%v, want 200/%v", rec.Code, svc.studentOverviewID, svc.studentTimetableID, studentID)
	}

	rec = httptest.NewRecorder()
	h.TeacherTimetable(rec, portalClaimsRequest("guru", "eid", employeeID))
	if rec.Code != http.StatusOK || svc.teacherTimetableID != employeeID {
		t.Fatalf("TeacherTimetable() status/id = %d/%v, want 200/%v", rec.Code, svc.teacherTimetableID, employeeID)
	}

	rec = httptest.NewRecorder()
	h.ParentMe(rec, portalClaimsRequest("ortu", "pid", parentID))
	if rec.Code != http.StatusOK || svc.parentOverviewID != parentID || svc.parentTimetableID != parentID {
		t.Fatalf("ParentMe() status/ids = %d/%v/%v, want 200/%v", rec.Code, svc.parentOverviewID, svc.parentTimetableID, parentID)
	}
}

func TestPortalAccessAndServiceErrors(t *testing.T) {
	studentID := handlerTestUUID(203)
	employeeID := handlerTestUUID(204)
	parentID := handlerTestUUID(205)
	errDB := errors.New("db down")

	claimsRequest := func(claims jwt.MapClaims) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/api/portal", nil)
		return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
	}
	tests := []struct {
		name       string
		handler    func(*Portal, http.ResponseWriter, *http.Request)
		svc        *fakePortalService
		req        *http.Request
		wantStatus int
	}{
		{name: "student without claims", handler: (*Portal).StudentMe, req: httptest.NewRequest(http.MethodGet, "/api/portal/student", nil), wantStatus: http.StatusForbidden},
		{name: "student wrong role", handler: (*Portal).StudentMe, req: claimsRequest(jwt.MapClaims{"roles": []any{"guru"}, "sid": studentID.String()}), wantStatus: http.StatusForbidden},
		{name: "student missing id", handler: (*Portal).StudentMe, req: claimsRequest(jwt.MapClaims{"roles": []any{"siswa"}}), wantStatus: http.StatusForbidden},
		{name: "student invalid id", handler: (*Portal).StudentMe, req: claimsRequest(jwt.MapClaims{"roles": []any{"siswa"}, "sid": "bad"}), wantStatus: http.StatusForbidden},
		{name: "student overview error", handler: (*Portal).StudentMe, svc: &fakePortalService{studentOverviewErr: errDB}, req: portalClaimsRequest("siswa", "sid", studentID), wantStatus: http.StatusInternalServerError},
		{name: "student timetable error", handler: (*Portal).StudentMe, svc: &fakePortalService{studentTimetableErr: errDB}, req: portalClaimsRequest("siswa", "sid", studentID), wantStatus: http.StatusInternalServerError},
		{name: "teacher without claims", handler: (*Portal).TeacherTimetable, req: httptest.NewRequest(http.MethodGet, "/api/portal/teacher", nil), wantStatus: http.StatusForbidden},
		{name: "teacher missing id", handler: (*Portal).TeacherTimetable, req: claimsRequest(jwt.MapClaims{"roles": []any{"guru"}}), wantStatus: http.StatusForbidden},
		{name: "teacher invalid id", handler: (*Portal).TeacherTimetable, req: claimsRequest(jwt.MapClaims{"roles": []any{"guru"}, "eid": "bad"}), wantStatus: http.StatusForbidden},
		{name: "teacher timetable error", handler: (*Portal).TeacherTimetable, svc: &fakePortalService{teacherTimetableErr: errDB}, req: portalClaimsRequest("guru", "eid", employeeID), wantStatus: http.StatusInternalServerError},
		{name: "parent without claims", handler: (*Portal).ParentMe, req: httptest.NewRequest(http.MethodGet, "/api/portal/parent", nil), wantStatus: http.StatusForbidden},
		{name: "parent missing id", handler: (*Portal).ParentMe, req: claimsRequest(jwt.MapClaims{"roles": []any{"ortu"}}), wantStatus: http.StatusForbidden},
		{name: "parent invalid id", handler: (*Portal).ParentMe, req: claimsRequest(jwt.MapClaims{"roles": []any{"ortu"}, "pid": "bad"}), wantStatus: http.StatusForbidden},
		{name: "parent overview error", handler: (*Portal).ParentMe, svc: &fakePortalService{parentOverviewErr: errDB}, req: portalClaimsRequest("ortu", "pid", parentID), wantStatus: http.StatusInternalServerError},
		{name: "parent timetable error", handler: (*Portal).ParentMe, svc: &fakePortalService{parentTimetableErr: errDB}, req: portalClaimsRequest("ortu", "pid", parentID), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakePortalService{}
			}
			rec := httptest.NewRecorder()
			tt.handler(&Portal{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
