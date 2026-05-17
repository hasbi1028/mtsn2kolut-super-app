package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestRombelHandlerMoreHomeroomListEdges(t *testing.T) {
	classID := handlerTestUUID(90)
	h := &Rombel{svc: &fakeRombelService{}}

	t.Run("read gate rejects before route parsing", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ListHomeroomAssignments(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/api/academic/rombel/bad/homeroom-assignments", nil), "id", "bad"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unauthenticated status = %d, want 401; body=%s", rec.Code, rec.Body.String())
		}

		fake := &fakeRombelService{}
		h.svc = fake
		rec = httptest.NewRecorder()
		req := withClaims(httptest.NewRequest(http.MethodGet, "/api/academic/rombel/bad/homeroom-assignments", nil), jwt.MapClaims{"roles": []any{"siswa"}})
		h.ListHomeroomAssignments(rec, withRouteParam(req, "id", "bad"))
		if rec.Code != http.StatusForbidden || fake.listHomeroomClassID.Valid {
			t.Fatalf("forbidden status/called = %d/%v, want 403/no service call", rec.Code, fake.listHomeroomClassID.Valid)
		}
	})

	t.Run("validates class id and maps service errors", func(t *testing.T) {
		fake := &fakeRombelService{}
		h.svc = fake
		rec := httptest.NewRecorder()
		h.ListHomeroomAssignments(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/bad/homeroom-assignments", ""), "id", "bad"))
		if rec.Code != http.StatusBadRequest || fake.listHomeroomClassID.Valid {
			t.Fatalf("invalid id status/called = %d/%v, want 400/no service call", rec.Code, fake.listHomeroomClassID.Valid)
		}

		fake = &fakeRombelService{err: errors.New("homeroom list failed")}
		h.svc = fake
		rec = httptest.NewRecorder()
		h.ListHomeroomAssignments(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", ""), "id", classID.String()))
		if rec.Code != http.StatusInternalServerError || fake.listHomeroomClassID != classID {
			t.Fatalf("service error status/class = %d/%s, want 500/%s; body=%s", rec.Code, fake.listHomeroomClassID.String(), classID.String(), rec.Body.String())
		}
	})
}

func TestRombelHandlerMoreHomeroomMutationEdges(t *testing.T) {
	classID := handlerTestUUID(91)
	assignmentID := handlerTestUUID(92)
	employeeID := handlerTestUUID(93)
	academicYearID := handlerTestUUID(94)
	h := &Rombel{svc: &fakeRombelService{}}

	t.Run("create rejects auth route json and optional parser errors before service", func(t *testing.T) {
		tests := []struct {
			name string
			req  *http.Request
		}{
			{name: "unauthenticated", req: withRouteParam(httptest.NewRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", strings.NewReader(`{}`)), "id", classID.String())},
			{name: "invalid class", req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/bad/homeroom-assignments", `{}`), "id", "bad")},
			{name: "malformed json", req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", `{"employee_id"`), "id", classID.String())},
			{name: "invalid employee", req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", `{"employee_id":"bad"}`), "id", classID.String())},
			{name: "invalid optional academic year", req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", `{"employee_id":"`+employeeID.String()+`","academic_year_id":"bad"}`), "id", classID.String())},
			{name: "invalid start date", req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", `{"employee_id":"`+employeeID.String()+`","start_date":"17-05-2026"}`), "id", classID.String())},
			{name: "invalid end date", req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", `{"employee_id":"`+employeeID.String()+`","end_date":"17-05-2026"}`), "id", classID.String())},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fake := &fakeRombelService{}
				h.svc = fake
				rec := httptest.NewRecorder()
				h.CreateHomeroomAssignment(rec, tt.req)
				want := http.StatusBadRequest
				if tt.name == "unauthenticated" {
					want = http.StatusUnauthorized
				}
				if rec.Code != want || fake.createHomeroomArg.ClassID.Valid {
					t.Fatalf("%s status/called = %d/%v, want %d/no service call; body=%s", tt.name, rec.Code, fake.createHomeroomArg.ClassID.Valid, want, rec.Body.String())
				}
			})
		}
	})

	t.Run("create defaults active true and keeps blank optionals nil", func(t *testing.T) {
		fake := &fakeRombelService{}
		h.svc = fake
		rec := httptest.NewRecorder()
		body := `{"employee_id":"` + employeeID.String() + `","academic_year_id":"   ","start_date":" ","end_date":"","notes":"  wali cadangan  "}`
		h.CreateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", body), "id", classID.String()))
		if rec.Code != http.StatusCreated {
			t.Fatalf("create default status = %d, want 201; body=%s", rec.Code, rec.Body.String())
		}
		if fake.createHomeroomArg.ClassID != classID || fake.createHomeroomArg.EmployeeID != employeeID || !fake.createHomeroomArg.HomeroomIsActive || fake.createHomeroomArg.HomeroomAcademicYearID != nil || fake.createHomeroomArg.HomeroomStartDate != nil || fake.createHomeroomArg.HomeroomEndDate.Valid || fake.createHomeroomArg.Notes != "wali cadangan" {
			t.Fatalf("create default arg = %+v, want defaults, nil optionals, trimmed notes", fake.createHomeroomArg)
		}
	})

	t.Run("create forwards optional academic year and service conflict", func(t *testing.T) {
		fake := &fakeRombelService{err: domain.ErrConflict}
		h.svc = fake
		rec := httptest.NewRecorder()
		body := `{"employee_id":"` + employeeID.String() + `","academic_year_id":"` + academicYearID.String() + `","start_date":"2026-07-01","end_date":"2027-06-30"}`
		h.CreateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", body), "id", classID.String()))
		if rec.Code != http.StatusConflict || fake.createHomeroomArg.HomeroomAcademicYearID == nil || fake.createHomeroomArg.HomeroomStartDate == nil || !fake.createHomeroomArg.HomeroomEndDate.Valid {
			t.Fatalf("create conflict status/arg = %d/%+v, want 409 with parsed optionals", rec.Code, fake.createHomeroomArg)
		}
	})

	t.Run("update validates assignment json and maps service not-found", func(t *testing.T) {
		badCases := []struct {
			name string
			req  *http.Request
		}{
			{name: "invalid assignment id", req: withRouteParam(adminRequest(http.MethodPut, "/api/academic/homeroom-assignments/bad", `{}`), "assignmentID", "bad")},
			{name: "malformed json", req: withRouteParam(adminRequest(http.MethodPut, "/api/academic/homeroom-assignments/"+assignmentID.String(), `{"employee_id"`), "assignmentID", assignmentID.String())},
			{name: "invalid employee", req: withRouteParam(adminRequest(http.MethodPut, "/api/academic/homeroom-assignments/"+assignmentID.String(), `{"employee_id":"bad"}`), "assignmentID", assignmentID.String())},
		}
		for _, tt := range badCases {
			t.Run(tt.name, func(t *testing.T) {
				fake := &fakeRombelService{}
				h.svc = fake
				rec := httptest.NewRecorder()
				h.UpdateHomeroomAssignment(rec, tt.req)
				if rec.Code != http.StatusBadRequest || fake.updateHomeroomArg.ID.Valid {
					t.Fatalf("%s status/called = %d/%v, want 400/no service call; body=%s", tt.name, rec.Code, fake.updateHomeroomArg.ID.Valid, rec.Body.String())
				}
			})
		}

		fake := &fakeRombelService{err: domain.ErrNotFound}
		h.svc = fake
		rec := httptest.NewRecorder()
		body := `{"employee_id":"` + employeeID.String() + `","notes":" update "}`
		h.UpdateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPut, "/api/academic/homeroom-assignments/"+assignmentID.String(), body), "assignmentID", assignmentID.String()))
		if rec.Code != http.StatusNotFound || fake.updateHomeroomArg.ID != assignmentID || !fake.updateHomeroomArg.HomeroomIsActive || fake.updateHomeroomArg.Notes != "update" {
			t.Fatalf("update not-found status/arg = %d/%+v, want 404 parsed defaults", rec.Code, fake.updateHomeroomArg)
		}
	})

	t.Run("delete validates assignment id and maps service bad request", func(t *testing.T) {
		fake := &fakeRombelService{}
		h.svc = fake
		rec := httptest.NewRecorder()
		h.DeleteHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/academic/homeroom-assignments/bad", ""), "assignmentID", "bad"))
		if rec.Code != http.StatusBadRequest || fake.deleteHomeroomID.Valid {
			t.Fatalf("delete invalid id status/called = %d/%v, want 400/no service call", rec.Code, fake.deleteHomeroomID.Valid)
		}

		fake = &fakeRombelService{err: domain.ErrBadRequest}
		h.svc = fake
		rec = httptest.NewRecorder()
		h.DeleteHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/academic/homeroom-assignments/"+assignmentID.String(), ""), "assignmentID", assignmentID.String()))
		if rec.Code != http.StatusBadRequest || fake.deleteHomeroomID != assignmentID {
			t.Fatalf("delete bad request status/id = %d/%s, want 400/%s", rec.Code, fake.deleteHomeroomID.String(), assignmentID.String())
		}
	})
}

func TestRombelHandlerMoreSubjectMatrixEdges(t *testing.T) {
	classID := handlerTestUUID(95)
	subjectID := handlerTestUUID(96)
	teacherID := handlerTestUUID(97)
	h := &Rombel{svc: &fakeRombelService{}}

	t.Run("get matrix gates reads and maps service error", func(t *testing.T) {
		fake := &fakeRombelService{}
		h.svc = fake
		rec := httptest.NewRecorder()
		h.GetSubjectAssignmentMatrix(rec, httptest.NewRequest(http.MethodGet, "/api/academic/subject-assignment-matrix", nil))
		if rec.Code != http.StatusUnauthorized || fake.matrixCalled {
			t.Fatalf("unauthorized status/called = %d/%v, want 401/false", rec.Code, fake.matrixCalled)
		}

		fake = &fakeRombelService{err: errors.New("matrix failed")}
		h.svc = fake
		rec = httptest.NewRecorder()
		h.GetSubjectAssignmentMatrix(rec, adminRequest(http.MethodGet, "/api/academic/subject-assignment-matrix", ""))
		if rec.Code != http.StatusInternalServerError || !fake.matrixCalled {
			t.Fatalf("service error status/called = %d/%v, want 500/true; body=%s", rec.Code, fake.matrixCalled, rec.Body.String())
		}
	})

	t.Run("update matrix rejects auth json and invalid ids before service", func(t *testing.T) {
		tests := []struct {
			name string
			req  *http.Request
			want int
		}{
			{name: "unauthenticated", req: httptest.NewRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", strings.NewReader(`{}`)), want: http.StatusUnauthorized},
			{name: "forbidden", req: withClaims(httptest.NewRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", strings.NewReader(`{"class_id":"bad"}`)), jwt.MapClaims{"roles": []any{"guru"}}), want: http.StatusForbidden},
			{name: "malformed json", req: adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", `{"class_id"`), want: http.StatusBadRequest},
			{name: "invalid class", req: adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", `{"class_id":"bad","subject_id":"`+subjectID.String()+`"}`), want: http.StatusBadRequest},
			{name: "invalid subject", req: adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", `{"class_id":"`+classID.String()+`","subject_id":"bad"}`), want: http.StatusBadRequest},
			{name: "invalid optional teacher", req: adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", `{"class_id":"`+classID.String()+`","subject_id":"`+subjectID.String()+`","teacher_employee_id":"bad"}`), want: http.StatusBadRequest},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fake := &fakeRombelService{}
				h.svc = fake
				rec := httptest.NewRecorder()
				h.UpdateSubjectAssignmentMatrixCell(rec, tt.req)
				if rec.Code != tt.want || fake.matrixUpdateArg.ClassID.Valid {
					t.Fatalf("%s status/called = %d/%v, want %d/no service call; body=%s", tt.name, rec.Code, fake.matrixUpdateArg.ClassID.Valid, tt.want, rec.Body.String())
				}
			})
		}
	})

	t.Run("update matrix accepts blank teacher and maps conflict", func(t *testing.T) {
		fake := &fakeRombelService{err: domain.ErrConflict}
		h.svc = fake
		rec := httptest.NewRecorder()
		body := `{"class_id":"` + classID.String() + `","subject_id":"` + subjectID.String() + `","teacher_employee_id":"   ","customization_notes":"butuh guru","additional_weekly_hours":2}`
		h.UpdateSubjectAssignmentMatrixCell(rec, adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", body))
		if rec.Code != http.StatusConflict || fake.matrixUpdateArg.ClassID != classID || fake.matrixUpdateArg.SubjectID != subjectID || fake.matrixUpdateArg.TeacherEmployeeID.Valid || fake.matrixUpdateArg.CustomizationNotes != "butuh guru" || fake.matrixUpdateArg.AdditionalWeeklyHours == nil || *fake.matrixUpdateArg.AdditionalWeeklyHours != 2 {
			t.Fatalf("matrix conflict status/arg = %d/%+v, want 409 with blank teacher parsed invalid", rec.Code, fake.matrixUpdateArg)
		}

		fake = &fakeRombelService{}
		h.svc = fake
		rec = httptest.NewRecorder()
		body = `{"class_id":"` + classID.String() + `","subject_id":"` + subjectID.String() + `","teacher_employee_id":"` + teacherID.String() + `"}`
		h.UpdateSubjectAssignmentMatrixCell(rec, adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", body))
		if rec.Code != http.StatusOK || fake.matrixUpdateArg.TeacherEmployeeID != teacherID {
			t.Fatalf("matrix success status/teacher = %d/%s, want 200/%s; body=%s", rec.Code, fake.matrixUpdateArg.TeacherEmployeeID.String(), teacherID.String(), rec.Body.String())
		}
	})
}

func TestRombelHandlerMoreOptionalUUIDArgAndCreateHomeroomParser(t *testing.T) {
	if got := optionalUUIDArg(pgtype.UUID{}); got != nil {
		t.Fatalf("optionalUUIDArg(invalid) = %#v, want nil", got)
	}
	validID := handlerTestUUID(98)
	if got, ok := optionalUUIDArg(validID).(pgtype.UUID); !ok || got != validID {
		t.Fatalf("optionalUUIDArg(valid) = %#v, want UUID %s", got, validID.String())
	}

	rec := httptest.NewRecorder()
	if _, ok := parseCreateHomeroomAssignment(rec, handlerTestUUID(99), homeroomAssignmentRequest{EmployeeID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseCreateHomeroomAssignment invalid employee ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	active := false
	rec = httptest.NewRecorder()
	params, ok := parseCreateHomeroomAssignment(rec, handlerTestUUID(100), homeroomAssignmentRequest{
		EmployeeID: handlerTestUUID(101).String(),
		IsActive:   &active,
		Notes:      " nonaktif ",
	})
	if !ok || params.HomeroomIsActive || params.Notes != "nonaktif" || params.HomeroomAcademicYearID != nil || params.HomeroomStartDate != nil || params.HomeroomEndDate.Valid {
		t.Fatalf("parseCreateHomeroomAssignment inactive/no optionals = (%+v,%v), want inactive trimmed with nil/invalid optionals", params, ok)
	}
}
