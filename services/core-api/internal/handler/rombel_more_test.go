package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestRombelMoreReadAllowedRolesPermissionsAndListGetEdges(t *testing.T) {
	classID := handlerTestUUID(81)

	readReq := func(claims jwt.MapClaims) *http.Request {
		return withClaims(httptest.NewRequest(http.MethodGet, "/api/academic/rombel", nil), claims)
	}

	t.Run("read gate returns unauthorized forbidden and allows scoped roles", func(t *testing.T) {
		h := &Rombel{svc: &fakeRombelService{}}
		tests := []struct {
			name       string
			req        *http.Request
			wantStatus int
			wantCalled bool
		}{
			{name: "unauthenticated", req: httptest.NewRequest(http.MethodGet, "/api/academic/rombel", nil), wantStatus: http.StatusUnauthorized},
			{name: "unscoped role", req: readReq(jwt.MapClaims{"roles": []any{"siswa"}}), wantStatus: http.StatusForbidden},
			{name: "guru role", req: readReq(jwt.MapClaims{"roles": []any{"guru"}}), wantStatus: http.StatusOK, wantCalled: true},
			{name: "staf role", req: readReq(jwt.MapClaims{"roles": []any{"staf"}}), wantStatus: http.StatusOK, wantCalled: true},
			{name: "kesiswaan role", req: readReq(jwt.MapClaims{"roles": []any{"kesiswaan"}}), wantStatus: http.StatusOK, wantCalled: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fake := &fakeRombelService{}
				h.svc = fake
				rec := httptest.NewRecorder()
				h.List(rec, tt.req)
				if rec.Code != tt.wantStatus || fake.listCalled != tt.wantCalled {
					t.Fatalf("status/called = %d/%v, want %d/%v; body=%s", rec.Code, fake.listCalled, tt.wantStatus, tt.wantCalled, rec.Body.String())
				}
			})
		}
	})

	t.Run("List maps service error after read gate", func(t *testing.T) {
		fake := &fakeRombelService{err: errors.New("list failed")}
		h := &Rombel{svc: fake}
		rec := httptest.NewRecorder()
		h.List(rec, adminRequest(http.MethodGet, "/api/academic/rombel", ""))
		if rec.Code != http.StatusInternalServerError || !fake.listCalled {
			t.Fatalf("List service error status/called = %d/%v, want 500/true; body=%s", rec.Code, fake.listCalled, rec.Body.String())
		}
	})

	t.Run("Get validates id, maps not-found, and stops on subject/timetable errors", func(t *testing.T) {
		h := &Rombel{svc: &fakeRombelService{}}
		rec := httptest.NewRecorder()
		h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/bad", ""), "id", "bad"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Get invalid id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}

		fake := &fakeRombelService{err: pgx.ErrNoRows}
		h = &Rombel{svc: fake}
		rec = httptest.NewRecorder()
		h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String(), ""), "id", classID.String()))
		if rec.Code != http.StatusNotFound || fake.listStudentsClassID.Valid {
			t.Fatalf("Get no rows status/students = %d/%v, want 404/no child loads", rec.Code, fake.listStudentsClassID)
		}

		childErr := errors.New("child failed")
		fake = &fakeRombelService{err: childErr}
		h = &Rombel{svc: &rombelStudentsFake{fakeRombelService: fake, rows: nil}}
		rec = httptest.NewRecorder()
		h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String(), ""), "id", classID.String()))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Get child error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestRombelMoreListStudentsSubjectsAndTimetableReadEdges(t *testing.T) {
	classID := handlerTestUUID(82)
	h := &Rombel{svc: &fakeRombelService{}}

	studentRows := []db.ListStudentsByClassWithParentsRow{
		{StudentID: handlerTestUUID(83), StudentName: "Tanpa Orang Tua"},
	}
	studentFake := &rombelStudentsFake{fakeRombelService: &fakeRombelService{}, rows: studentRows}
	h.svc = studentFake
	rec := httptest.NewRecorder()
	h.ListStudents(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/students", ""), "id", classID.String()))
	if rec.Code != http.StatusOK || studentFake.listStudentsClassID != classID || !strings.Contains(rec.Body.String(), `"parents":[]`) {
		t.Fatalf("ListStudents no-parent status/arg/body = %d/%v/%s, want grouped empty parents", rec.Code, studentFake.listStudentsClassID, rec.Body.String())
	}

	tests := []struct {
		name       string
		fn         func(http.ResponseWriter, *http.Request)
		req        *http.Request
		svc        *fakeRombelService
		wantStatus int
		assert     func(*testing.T, *fakeRombelService)
	}{
		{name: "ListStudents invalid id", fn: h.ListStudents, req: withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/bad/students", ""), "id", "bad"), svc: &fakeRombelService{}, wantStatus: http.StatusBadRequest},
		{name: "ListStudents service error", fn: h.ListStudents, req: withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/students", ""), "id", classID.String()), svc: &fakeRombelService{err: errors.New("students failed")}, wantStatus: http.StatusInternalServerError},
		{name: "ListSubjectAssignments invalid id", fn: h.ListSubjectAssignments, req: withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/bad/subject-assignments", ""), "id", "bad"), svc: &fakeRombelService{}, wantStatus: http.StatusBadRequest},
		{name: "ListSubjectAssignments service error", fn: h.ListSubjectAssignments, req: withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/subject-assignments", ""), "id", classID.String()), svc: &fakeRombelService{err: errors.New("subjects failed")}, wantStatus: http.StatusInternalServerError, assert: func(t *testing.T, f *fakeRombelService) {
			t.Helper()
			if f.listSubjectClassID != classID {
				t.Fatalf("subject class = %v, want %v", f.listSubjectClassID, classID)
			}
		}},
		{name: "ListTimetableSlots invalid id", fn: h.ListTimetableSlots, req: withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/bad/timetable-slots", ""), "id", "bad"), svc: &fakeRombelService{}, wantStatus: http.StatusBadRequest},
		{name: "ListTimetableSlots service error", fn: h.ListTimetableSlots, req: withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/timetable-slots", ""), "id", classID.String()), svc: &fakeRombelService{err: errors.New("timetable failed")}, wantStatus: http.StatusInternalServerError, assert: func(t *testing.T, f *fakeRombelService) {
			t.Helper()
			if f.listTimetableClassID != classID {
				t.Fatalf("timetable class = %v, want %v", f.listTimetableClassID, classID)
			}
		}},
		{name: "ListSubjectAssignments read forbidden before parse", fn: h.ListSubjectAssignments, req: withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, "/api/academic/rombel/bad/subject-assignments", nil), jwt.MapClaims{"roles": []any{"siswa"}}), "id", "bad"), svc: &fakeRombelService{}, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.svc = tt.svc
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.assert != nil {
				tt.assert(t, tt.svc)
			}
		})
	}
}
