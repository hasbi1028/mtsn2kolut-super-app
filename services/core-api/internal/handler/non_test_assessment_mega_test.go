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

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type megaNonTestAssessmentHandlerService struct {
	listCalled bool
	listErr    error

	createCalled bool
	createErr    error
	createInput  service.SaveNonTestAssessmentInput

	updateCalled bool
	updateErr    error
	updateInput  service.SaveNonTestAssessmentInput

	deleteCalled bool
	deleteErr    error
	deleteID     pgtype.UUID

	listSubmissionsCalled bool
	listSubmissionsErr    error

	generateCalled bool
	generateErr    error

	syncCalled  bool
	syncErr     error
	syncPublish bool

	upsertCalled bool
	upsertErr    error
	upsertInput  service.SaveNonTestSubmissionInput

	ownsClassSubject bool
	ownsClassErr     error
	classCalls       int

	ownsAssessment bool
	ownsAssessErr  error
	assessCalls    int
}

func (f *megaNonTestAssessmentHandlerService) List(context.Context, service.ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error) {
	f.listCalled = true
	return nil, 0, f.listErr
}

func (f *megaNonTestAssessmentHandlerService) Get(context.Context, pgtype.UUID) (db.GetNonTestAssessmentRow, error) {
	return db.GetNonTestAssessmentRow{}, nil
}

func (f *megaNonTestAssessmentHandlerService) Create(_ context.Context, input service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	f.createCalled = true
	f.createInput = input
	return db.NonTestAssessment{ID: handlerTestUUID(211), SubjectID: input.SubjectID, ClassID: input.ClassID, Title: input.Title}, f.createErr
}

func (f *megaNonTestAssessmentHandlerService) Update(_ context.Context, input service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	f.updateCalled = true
	f.updateInput = input
	return db.NonTestAssessment{ID: input.ID, SubjectID: input.SubjectID, ClassID: input.ClassID, Title: input.Title}, f.updateErr
}

func (f *megaNonTestAssessmentHandlerService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteCalled = true
	f.deleteID = id
	return f.deleteErr
}

func (f *megaNonTestAssessmentHandlerService) ListSubmissions(context.Context, pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	f.listSubmissionsCalled = true
	return nil, f.listSubmissionsErr
}

func (f *megaNonTestAssessmentHandlerService) GenerateSubmissions(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) ([]db.NonTestAssessmentSubmission, error) {
	f.generateCalled = true
	return nil, f.generateErr
}

func (f *megaNonTestAssessmentHandlerService) UpsertSubmission(_ context.Context, input service.SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error) {
	f.upsertCalled = true
	f.upsertInput = input
	return db.NonTestAssessmentSubmission{AssessmentID: input.AssessmentID, StudentID: input.StudentID, Status: input.Status}, f.upsertErr
}

func (f *megaNonTestAssessmentHandlerService) SyncToGrade(context.Context, pgtype.UUID, pgtype.UUID, string, bool) (service.SyncNonTestAssessmentToGradeResult, error) {
	f.syncCalled = true
	return service.SyncNonTestAssessmentToGradeResult{}, f.syncErr
}

func (f *megaNonTestAssessmentHandlerService) TeacherOwnsClassSubject(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) (bool, error) {
	f.classCalls++
	if f.ownsClassErr != nil {
		return false, f.ownsClassErr
	}
	return f.ownsClassSubject, nil
}

func (f *megaNonTestAssessmentHandlerService) TeacherOwnsAssessment(context.Context, pgtype.UUID, pgtype.UUID) (bool, error) {
	f.assessCalls++
	if f.ownsAssessErr != nil {
		return false, f.ownsAssessErr
	}
	return f.ownsAssessment, nil
}

func megaNonTestAdminReq(method, path, body string) *http.Request {
	return withClaims(httptest.NewRequest(method, path, strings.NewReader(body)), jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin.mega"})
}

func megaNonTestGuruReq(method, path, body string) *http.Request {
	return withClaims(httptest.NewRequest(method, path, strings.NewReader(body)), jwt.MapClaims{"roles": []any{"guru"}, "eid": "33333333-3333-3333-3333-333333333333", "usr": "guru.mega"})
}

func megaNonTestAssessmentBody(subjectID, classID string) string {
	return `{"subject_id":"` + subjectID + `","class_id":"` + classID + `","assessment_type":"observasi","title":"Observasi mega","status":"active"}`
}

func TestNonTestAssessmentMegaAuthAndInvalidJSONEdges(t *testing.T) {
	assessmentID := handlerTestUUID(212)
	subjectID := pgUUIDString(handlerTestUUID(213))
	classID := pgUUIDString(handlerTestUUID(214))
	studentID := pgUUIDString(handlerTestUUID(215))

	t.Run("list rejects missing CBT role before service", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		h.List(rec, httptest.NewRequest(http.MethodGet, "/api/non-test-assessments", nil))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("List() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.listCalled {
			t.Fatalf("List service called despite missing role")
		}
	})

	t.Run("create rejects malformed JSON before ownership and service", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsClassSubject: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		h.Create(rec, megaNonTestGuruReq(http.MethodPost, "/api/non-test-assessments", `{`))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Create() bad JSON status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if fake.classCalls != 0 || fake.createCalled {
			t.Fatalf("Create bad JSON reached ownership/service: classCalls=%d create=%v", fake.classCalls, fake.createCalled)
		}
	})

	t.Run("update rejects invalid subject before assessment ownership", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsAssessment: true, ownsClassSubject: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(megaNonTestGuruReq(http.MethodPut, "/", `{"subject_id":"bad","class_id":"`+classID+`"}`), "id", pgUUIDString(assessmentID))
		h.Update(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Update() bad subject status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if fake.assessCalls != 0 || fake.updateCalled {
			t.Fatalf("Update bad subject reached ownership/service: assessCalls=%d update=%v", fake.assessCalls, fake.updateCalled)
		}
	})

	t.Run("generate and sync reject malformed JSON after assessment ownership", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(megaNonTestGuruReq(http.MethodPost, "/", `{`), "id", pgUUIDString(assessmentID))
		h.GenerateSubmissions(rec, req)
		if rec.Code != http.StatusBadRequest || fake.assessCalls != 1 || fake.generateCalled {
			t.Fatalf("Generate bad JSON status/calls = %d/%d/%v; body=%s", rec.Code, fake.assessCalls, fake.generateCalled, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(megaNonTestGuruReq(http.MethodPost, "/", `{`), "id", pgUUIDString(assessmentID))
		h.SyncGrade(rec, req)
		if rec.Code != http.StatusBadRequest || fake.assessCalls != 2 || fake.syncCalled {
			t.Fatalf("Sync bad JSON status/calls = %d/%d/%v; body=%s", rec.Code, fake.assessCalls, fake.syncCalled, rec.Body.String())
		}
	})

	t.Run("upsert rejects malformed JSON and invalid route id", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(megaNonTestGuruReq(http.MethodPut, "/", `{`), "id", pgUUIDString(assessmentID))
		h.UpsertSubmission(rec, req)
		if rec.Code != http.StatusBadRequest || fake.assessCalls != 1 || fake.upsertCalled {
			t.Fatalf("Upsert bad JSON status/calls = %d/%d/%v; body=%s", rec.Code, fake.assessCalls, fake.upsertCalled, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(megaNonTestGuruReq(http.MethodPut, "/", `{"student_id":"`+studentID+`"}`), "id", "bad")
		h.UpsertSubmission(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Upsert bad id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("create with valid body but no teacher employee id is forbidden before scope service", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsClassSubject: true}
		h := &NonTestAssessment{svc: fake}
		req := withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(megaNonTestAssessmentBody(subjectID, classID))), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.noeid"})
		rec := httptest.NewRecorder()
		h.Create(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Create() no eid status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.classCalls != 0 || fake.createCalled {
			t.Fatalf("Create no eid reached scope/service: classCalls=%d create=%v", fake.classCalls, fake.createCalled)
		}
	})
}

func TestNonTestAssessmentMegaTeacherOwnershipEdges(t *testing.T) {
	assessmentID := handlerTestUUID(216)
	subjectID := pgUUIDString(handlerTestUUID(217))
	classID := pgUUIDString(handlerTestUUID(218))
	body := megaNonTestAssessmentBody(subjectID, classID)

	t.Run("create maps class subject ownership store error to internal and does not create", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsClassErr: errors.New("scope store down")}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		h.Create(rec, megaNonTestGuruReq(http.MethodPost, "/", body))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Create() ownership error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
		if fake.classCalls != 1 || fake.createCalled {
			t.Fatalf("Create ownership error classCalls/create = %d/%v", fake.classCalls, fake.createCalled)
		}
	})

	t.Run("update stops when new class subject is not owned", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsAssessment: true, ownsClassSubject: false}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(megaNonTestGuruReq(http.MethodPut, "/", body), "id", pgUUIDString(assessmentID))
		h.Update(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Update() class denied status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.assessCalls != 1 || fake.classCalls != 1 || fake.updateCalled {
			t.Fatalf("Update class denied calls/service = assess %d class %d update %v", fake.assessCalls, fake.classCalls, fake.updateCalled)
		}
	})

	t.Run("list submissions and sync stop on teacher assessment denial", func(t *testing.T) {
		fake := &megaNonTestAssessmentHandlerService{ownsAssessment: false}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(megaNonTestGuruReq(http.MethodGet, "/", ""), "id", pgUUIDString(assessmentID))
		h.ListSubmissions(rec, req)
		if rec.Code != http.StatusForbidden || fake.listSubmissionsCalled {
			t.Fatalf("ListSubmissions denied status/service = %d/%v; body=%s", rec.Code, fake.listSubmissionsCalled, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(megaNonTestGuruReq(http.MethodPost, "/", `{}`), "id", pgUUIDString(assessmentID))
		h.SyncGrade(rec, req)
		if rec.Code != http.StatusForbidden || fake.syncCalled {
			t.Fatalf("SyncGrade denied status/service = %d/%v; body=%s", rec.Code, fake.syncCalled, rec.Body.String())
		}
		if fake.assessCalls != 2 {
			t.Fatalf("assessment ownership calls = %d, want 2", fake.assessCalls)
		}
	})
}

func TestNonTestAssessmentMegaDomainErrorMapping(t *testing.T) {
	assessmentID := handlerTestUUID(219)
	subjectID := pgUUIDString(handlerTestUUID(220))
	classID := pgUUIDString(handlerTestUUID(221))
	studentID := pgUUIDString(handlerTestUUID(222))
	body := megaNonTestAssessmentBody(subjectID, classID)

	tests := []struct {
		name      string
		setup     func(*megaNonTestAssessmentHandlerService)
		run       func(*NonTestAssessment, *httptest.ResponseRecorder)
		wantCode  int
		wasCalled func(*megaNonTestAssessmentHandlerService) bool
	}{
		{
			name: "create bad request",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.createErr = domain.ErrBadRequest
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				h.Create(rec, megaNonTestAdminReq(http.MethodPost, "/", body))
			},
			wantCode:  http.StatusBadRequest,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.createCalled },
		},
		{
			name: "update conflict",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.updateErr = domain.ErrConflict
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				req := withRouteParam(megaNonTestAdminReq(http.MethodPut, "/", body), "id", pgUUIDString(assessmentID))
				h.Update(rec, req)
			},
			wantCode:  http.StatusConflict,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.updateCalled },
		},
		{
			name: "delete not found",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.deleteErr = domain.ErrNotFound
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				req := withRouteParam(megaNonTestAdminReq(http.MethodDelete, "/", ""), "id", pgUUIDString(assessmentID))
				h.Delete(rec, req)
			},
			wantCode:  http.StatusNotFound,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.deleteCalled },
		},
		{
			name: "list submissions unauthorized",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.listSubmissionsErr = domain.ErrUnauthorized
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				req := withRouteParam(megaNonTestAdminReq(http.MethodGet, "/", ""), "id", pgUUIDString(assessmentID))
				h.ListSubmissions(rec, req)
			},
			wantCode:  http.StatusUnauthorized,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.listSubmissionsCalled },
		},
		{
			name: "generate duplicate conflict",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.generateErr = errors.New("duplicate submission unique")
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				req := withRouteParam(megaNonTestAdminReq(http.MethodPost, "/", `{}`), "id", pgUUIDString(assessmentID))
				h.GenerateSubmissions(rec, req)
			},
			wantCode:  http.StatusConflict,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.generateCalled },
		},
		{
			name: "sync conflict",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.syncErr = domain.ErrConflict
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				req := withRouteParam(megaNonTestAdminReq(http.MethodPost, "/", `{}`), "id", pgUUIDString(assessmentID))
				h.SyncGrade(rec, req)
			},
			wantCode:  http.StatusConflict,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.syncCalled },
		},
		{
			name: "upsert not found",
			setup: func(f *megaNonTestAssessmentHandlerService) {
				f.upsertErr = domain.ErrNotFound
			},
			run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
				req := withRouteParam(megaNonTestAdminReq(http.MethodPut, "/", `{"student_id":"`+studentID+`","status":"draft"}`), "id", pgUUIDString(assessmentID))
				h.UpsertSubmission(rec, req)
			},
			wantCode:  http.StatusNotFound,
			wasCalled: func(f *megaNonTestAssessmentHandlerService) bool { return f.upsertCalled },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &megaNonTestAssessmentHandlerService{ownsAssessment: true, ownsClassSubject: true}
			tt.setup(fake)
			h := &NonTestAssessment{svc: fake}
			rec := httptest.NewRecorder()
			tt.run(h, rec)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
			if !tt.wasCalled(fake) {
				t.Fatalf("expected service method to be called")
			}
		})
	}
}
