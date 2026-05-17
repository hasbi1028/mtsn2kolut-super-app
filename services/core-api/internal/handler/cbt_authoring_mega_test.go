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
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtAuthoringMegaQuestionGetEdgesAndRedaction(t *testing.T) {
	questionID := handlerTestUUID(200)
	subjectID := handlerTestUUID(201)
	detail := db.GetCbtQuestionDetailRow{
		ID:             questionID,
		SubjectID:      subjectID,
		SubjectName:    "IPA",
		Code:           "Q-SECRET",
		QuestionText:   "Stem rahasia",
		QuestionType:   "essay",
		AnswerKey:      "KUNCI-RAHASIA",
		RubricHtml:     "<p>rubrik rahasia</p>",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		AuthorUsername: "guru.pemilik",
	}

	tests := []struct {
		name     string
		h        *CbtQuestion
		req      *http.Request
		wantCode int
		want     string
		forbid   string
	}{
		{
			name:     "forbidden without CBT access",
			h:        &CbtQuestion{svc: &fakeCbtQuestionService{}},
			req:      withRouteParam(httptest.NewRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String(), nil), "id", questionID.String()),
			wantCode: http.StatusForbidden,
		},
		{
			name:     "bad route id",
			h:        &CbtQuestion{svc: &fakeCbtQuestionService{}},
			req:      withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/bad", ""), "id", "bad"),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "service not found maps to 404",
			h:        &CbtQuestion{svc: &fakeCbtQuestionService{getErr: domain.ErrNotFound}},
			req:      withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()),
			wantCode: http.StatusNotFound,
		},
		{
			name: "non author detail response is redacted by service actor contract",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{getDetailRow: detail}},
			req: withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String(), nil), jwt.MapClaims{
				"roles":       []string{"guru"},
				"permissions": []string{"bank_soal.read"},
				"usr":         "guru.lain",
				"uid":         handlerTestUUID(202).String(),
				"sub":         handlerTestUUID(202).String(),
			}), "id", questionID.String()),
			wantCode: http.StatusOK,
			want:     "Q-SECRET",
			forbid:   "KUNCI-RAHASIA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.h.Get(rec, tt.req)
			if rec.Code != tt.wantCode {
				t.Fatalf("Get() status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.want != "" && !strings.Contains(rec.Body.String(), tt.want) {
				t.Fatalf("Get() body = %s, want substring %q", rec.Body.String(), tt.want)
			}
			if tt.forbid != "" && strings.Contains(rec.Body.String(), tt.forbid) {
				t.Fatalf("Get() body leaked redacted substring %q: %s", tt.forbid, rec.Body.String())
			}
		})
	}
}

func TestCbtAuthoringMegaQuestionCreateUpdateDeleteEdges(t *testing.T) {
	questionID := handlerTestUUID(203)
	subjectID := handlerTestUUID(204)
	baseDetail := db.GetCbtQuestionDetailRow{
		ID:             questionID,
		SubjectID:      subjectID,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		AuthorUsername: "guru.pemilik",
	}
	validBody := `{"subject_id":"` + subjectID.String() + `","question_text":"Teks","question_type":"essay","difficulty":"medium","status":"draft"}`

	tests := []struct {
		name     string
		h        *CbtQuestion
		run      func(*CbtQuestion, *httptest.ResponseRecorder)
		wantCode int
	}{
		{
			name: "create invalid json",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Create(rec, adminRequest(http.MethodPost, "/api/cbt/questions", `{`))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "create service bad request",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{createErr: domain.ErrBadRequest}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Create(rec, adminRequest(http.MethodPost, "/api/cbt/questions", validBody))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "update non author forbidden before body decode",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{getDetailRow: baseDetail}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				req := withClaims(httptest.NewRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), strings.NewReader(validBody)), jwt.MapClaims{
					"roles":       []any{"guru"},
					"permissions": []any{"bank_soal.update_own"},
					"usr":         "guru.lain",
					"uid":         handlerTestUUID(205).String(),
					"sub":         handlerTestUUID(205).String(),
				})
				h.Update(rec, withRouteParam(req, "id", questionID.String()))
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "update locked question conflicts",
			h: &CbtQuestion{svc: &fakeCbtQuestionService{getDetailRow: func() db.GetCbtQuestionDetailRow {
				row := baseDetail
				row.PackageCount = 1
				return row
			}()}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Update(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), validBody), "id", questionID.String()))
			},
			wantCode: http.StatusConflict,
		},
		{
			name: "update service conflict mapping",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{getDetailRow: baseDetail, updateErr: domain.ErrConflict}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Update(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/questions/"+questionID.String(), validBody), "id", questionID.String()))
			},
			wantCode: http.StatusConflict,
		},
		{
			name: "delete get detail internal error",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{getErr: errors.New("db down")}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "delete locked question conflicts",
			h: &CbtQuestion{svc: &fakeCbtQuestionService{getDetailRow: func() db.GetCbtQuestionDetailRow {
				row := baseDetail
				row.AnswerCount = 2
				return row
			}()}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/questions/"+questionID.String(), ""), "id", questionID.String()))
			},
			wantCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.run(tt.h, rec)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}

func TestCbtAuthoringMegaQuestionWorkflowRemainingEdges(t *testing.T) {
	questionID := handlerTestUUID(206)
	subjectID := handlerTestUUID(207)
	row := cbtQuestionHandlerModel(questionID, subjectID)

	tests := []struct {
		name     string
		h        *CbtQuestion
		run      func(*CbtQuestion, *httptest.ResponseRecorder)
		wantCode int
		assert   func(*testing.T, *fakeCbtQuestionService, string)
	}{
		{
			name: "workflow bad json",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.WorkflowAction(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", `{`), "id", questionID.String()))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "publish forbidden without permission",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				req := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"publish"}`)), jwt.MapClaims{
					"roles":       []any{"guru"},
					"permissions": []any{"bank_soal.read"},
					"usr":         "guru.reader",
					"uid":         handlerTestUUID(212).String(),
					"sub":         handlerTestUUID(212).String(),
				})
				h.WorkflowAction(rec, withRouteParam(req, "id", questionID.String()))
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "approve requires actor user id for non admin",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.WorkflowAction(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/workflow", strings.NewReader(`{"action":"approve"}`)), jwt.MapClaims{
					"roles":       []any{"guru"},
					"permissions": []any{"bank_soal.approve"},
					"usr":         "guru.reviewer",
				}), "id", questionID.String()))
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "duplicate bad id",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Duplicate(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/bad/duplicate", ``), "id", "bad"))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "duplicate service not found",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{duplicateErr: domain.ErrNotFound}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.Duplicate(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/duplicate", ``), "id", questionID.String()))
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "mark revision bad json",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.MarkRevision(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/revision", `{`), "id", questionID.String()))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "mark revision forwards notes",
			h:    &CbtQuestion{svc: &fakeCbtQuestionService{revisionRow: row}},
			run: func(h *CbtQuestion, rec *httptest.ResponseRecorder) {
				h.MarkRevision(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/questions/"+questionID.String()+"/revision", `{"notes":"buat revisi"}`), "id", questionID.String()))
			},
			wantCode: http.StatusCreated,
			assert: func(t *testing.T, fake *fakeCbtQuestionService, body string) {
				t.Helper()
				if fake.revisionID != questionID || fake.revisionNotes != "buat revisi" || !strings.Contains(body, "Q-HANDLER") {
					t.Fatalf("MarkRevision forwarded id/notes/body = %v/%q/%s", fake.revisionID, fake.revisionNotes, body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.run(tt.h, rec)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.assert != nil {
				fake, ok := tt.h.svc.(*fakeCbtQuestionService)
				if !ok {
					t.Fatalf("service = %T, want fakeCbtQuestionService", tt.h.svc)
				}
				tt.assert(t, fake, rec.Body.String())
			}
		})
	}
}

func TestCbtAuthoringMegaPackageCreateDeleteEdges(t *testing.T) {
	packageID := handlerTestUUID(208)
	subjectID := handlerTestUUID(209)
	questionID := handlerTestUUID(210)
	validCreate := `{"subject_id":"` + subjectID.String() + `","title":"Paket Mega","duration_minutes":90,"question_ids":["` + questionID.String() + `"],"question_weights":{"` + questionID.String() + `":2},"is_active":true}`

	tests := []struct {
		name     string
		h        *CbtPackage
		run      func(*CbtPackage, *httptest.ResponseRecorder)
		wantCode int
		assert   func(*testing.T, *fakeCbtPackageService)
	}{
		{
			name: "create forbidden without package permission",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Create(rec, withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/packages", strings.NewReader(validCreate)), jwt.MapClaims{"roles": []any{"guru"}, "permissions": []any{"bank_soal.read"}}))
			},
			wantCode: http.StatusForbidden,
		},
		{
			name: "create rejects duration below range",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Create(rec, cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages", `{"subject_id":"`+subjectID.String()+`","title":"P","duration_minutes":0}`))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "create rejects invalid event id",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Create(rec, cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages", `{"event_id":"bad","subject_id":"`+subjectID.String()+`","title":"P","duration_minutes":90}`))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "create rejects invalid subject id",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Create(rec, cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages", `{"subject_id":"bad","title":"P","duration_minutes":90}`))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "create rejects invalid question id",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Create(rec, cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages", `{"subject_id":"`+subjectID.String()+`","title":"P","duration_minutes":90,"question_ids":["bad"]}`))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "create service conflict mapping and forwards weights",
			h:    &CbtPackage{svc: &fakeCbtPackageService{createErr: domain.ErrConflict}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Create(rec, cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages", validCreate))
			},
			wantCode: http.StatusConflict,
			assert: func(t *testing.T, fake *fakeCbtPackageService) {
				t.Helper()
				if fake.createInput.SubjectID != subjectID || len(fake.createInput.QuestionIDs) != 1 || fake.createInput.QuestionIDs[0] != questionID || fake.createInput.QuestionWeights[questionID.String()] != 2 {
					t.Fatalf("Create input = %+v, weights=%v", fake.createInput, fake.createInput.QuestionWeights)
				}
			},
		},
		{
			name: "delete bad id",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Delete(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodDelete, "/api/cbt/packages/bad", ""), "id", "bad"))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "delete service not found mapping",
			h:    &CbtPackage{svc: &fakeCbtPackageService{deleteErr: domain.ErrNotFound}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Delete(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodDelete, "/api/cbt/packages/"+packageID.String(), ""), "id", packageID.String()))
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.run(tt.h, rec)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.assert != nil {
				fake, ok := tt.h.svc.(*fakeCbtPackageService)
				if !ok {
					t.Fatalf("service = %T, want fakeCbtPackageService", tt.h.svc)
				}
				tt.assert(t, fake)
			}
		})
	}
}

func TestCbtAuthoringMegaPackageHelpersAndReadEdges(t *testing.T) {
	if parsed, err := parseOptionalUUID("  "); err != nil || parsed.Valid {
		t.Fatalf("parseOptionalUUID(blank) = %v/%v, want zero nil", parsed, err)
	}
	if _, err := parseOptionalUUID("bad"); err == nil {
		t.Fatal("parseOptionalUUID(bad) error = nil, want error")
	}

	packageID := handlerTestUUID(211)
	tests := []struct {
		name     string
		h        *CbtPackage
		run      func(*CbtPackage, *httptest.ResponseRecorder)
		wantCode int
	}{
		{
			name: "list internal service error",
			h:    &CbtPackage{svc: &fakeCbtPackageService{listErr: errors.New("db down")}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.List(rec, cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages", ""))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "readiness bad event id",
			h:    &CbtPackage{svc: &fakeCbtPackageService{}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Readiness(rec, cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/readiness?event_id=bad", ""))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "get internal service error",
			h:    &CbtPackage{svc: &fakeCbtPackageService{detailErr: errors.New("db down")}},
			run: func(h *CbtPackage, rec *httptest.ResponseRecorder) {
				h.Get(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/"+packageID.String(), ""), "id", packageID.String()))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.run(tt.h, rec)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}

var _ = pgtype.UUID{}
