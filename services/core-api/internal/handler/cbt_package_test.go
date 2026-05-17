package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeCbtPackageService struct {
	listEventID pgtype.UUID
	listRows    []db.ListCbtPackagesRow
	listQs      []db.ListCbtPackageQuestionsRow
	listErr     error

	detailID  pgtype.UUID
	detailRow service.CbtPackageDetailResult
	detailErr error

	readinessEventID pgtype.UUID
	readinessRow     service.CbtPackageReadinessOverview
	readinessErr     error

	createInput service.CreateCbtPackageInput
	createRow   db.CbtPackage
	createErr   error

	updateInput service.UpdateCbtPackageInput
	updateRow   service.CbtPackageDetailResult
	updateErr   error

	replaceInput service.ReplaceCbtPackageQuestionsInput
	replaceRow   service.CbtPackageDetailResult
	replaceErr   error

	cloneInput service.CloneCbtPackageInput
	cloneRow   service.CbtPackageDetailResult
	cloneErr   error

	lockPackageID pgtype.UUID
	lockUserID    pgtype.UUID
	lockReason    string
	lockRow       service.CbtPackageSnapshotResult
	lockErr       error

	deleteID  pgtype.UUID
	deleteErr error
}

func (f *fakeCbtPackageService) List(_ context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error) {
	f.listEventID = eventID
	return f.listRows, f.listQs, f.listErr
}

func (f *fakeCbtPackageService) Detail(_ context.Context, id pgtype.UUID) (service.CbtPackageDetailResult, error) {
	f.detailID = id
	return f.detailRow, f.detailErr
}

func (f *fakeCbtPackageService) Readiness(_ context.Context, eventID pgtype.UUID) (service.CbtPackageReadinessOverview, error) {
	f.readinessEventID = eventID
	return f.readinessRow, f.readinessErr
}

func (f *fakeCbtPackageService) Create(_ context.Context, input service.CreateCbtPackageInput) (db.CbtPackage, error) {
	f.createInput = input
	return f.createRow, f.createErr
}

func (f *fakeCbtPackageService) UpdateMetadata(_ context.Context, input service.UpdateCbtPackageInput) (service.CbtPackageDetailResult, error) {
	f.updateInput = input
	return f.updateRow, f.updateErr
}

func (f *fakeCbtPackageService) ReplaceQuestions(_ context.Context, input service.ReplaceCbtPackageQuestionsInput) (service.CbtPackageDetailResult, error) {
	f.replaceInput = input
	return f.replaceRow, f.replaceErr
}

func (f *fakeCbtPackageService) Clone(_ context.Context, input service.CloneCbtPackageInput) (service.CbtPackageDetailResult, error) {
	f.cloneInput = input
	return f.cloneRow, f.cloneErr
}

func (f *fakeCbtPackageService) LockAndSnapshot(_ context.Context, packageID, lockedBy pgtype.UUID, reason string) (service.CbtPackageSnapshotResult, error) {
	f.lockPackageID = packageID
	f.lockUserID = lockedBy
	f.lockReason = reason
	return f.lockRow, f.lockErr
}

func (f *fakeCbtPackageService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func cbtPackageAuthedRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	claims := jwt.MapClaims{
		"roles": []any{"admin"},
		"sub":   handlerTestUUID(240).String(),
		"uid":   handlerTestUUID(240).String(),
		"usr":   "admin.test",
		"ssid":  "session-1",
	}
	return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
}

func cbtPackageRouteParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func cbtPackageDetail(id pgtype.UUID, title string) service.CbtPackageDetailResult {
	return service.CbtPackageDetailResult{
		Package: db.GetCbtPackageDetailRow{ID: id, Title: title, SubjectName: "Matematika", DurationMinutes: 90, IsActive: true},
		Questions: []db.ListCbtPackageQuestionsByPackageRow{
			{PackageID: id, QuestionID: handlerTestUUID(111), Position: 1, Points: 2, QuestionCode: "Q-001", QuestionType: "pg"},
		},
		Readiness: service.CbtPackageReadinessStatus{Status: "ready", Ready: true, QuestionCount: 1, TotalPoints: 2},
	}
}

func TestCbtPackageReadinessAndGetForwardIDs(t *testing.T) {
	eventID := handlerTestUUID(60)
	packageID := handlerTestUUID(61)
	fake := &fakeCbtPackageService{
		readinessRow: service.CbtPackageReadinessOverview{Items: []service.CbtPackageReadinessItem{
			{Package: db.ListCbtPackageReadinessRow{ID: packageID, EventID: eventID, Title: "Paket A"}, Readiness: service.CbtPackageReadinessStatus{Status: "ready", Ready: true}},
		}},
		detailRow: cbtPackageDetail(packageID, "Paket A"),
	}
	h := &CbtPackage{svc: fake}

	rec := httptest.NewRecorder()
	h.Readiness(rec, cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/readiness?event_id="+eventID.String(), ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("Readiness status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.readinessEventID != eventID {
		t.Fatalf("Readiness event id = %v, want %v", fake.readinessEventID, eventID)
	}
	if !strings.Contains(rec.Body.String(), "Paket A") {
		t.Fatalf("Readiness body = %s, want package title", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Get(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/"+packageID.String(), ""), "id", packageID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Get status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.detailID != packageID {
		t.Fatalf("Get id = %v, want %v", fake.detailID, packageID)
	}
}

func TestCbtPackageUpdateReplaceCloneAndLockForwardInputs(t *testing.T) {
	packageID := handlerTestUUID(62)
	questionA := handlerTestUUID(63)
	questionB := handlerTestUUID(64)
	fake := &fakeCbtPackageService{
		updateRow:  cbtPackageDetail(packageID, "Paket Revisi"),
		replaceRow: cbtPackageDetail(packageID, "Paket Revisi"),
		cloneRow:   cbtPackageDetail(handlerTestUUID(65), "Paket Clone"),
		lockRow: service.CbtPackageSnapshotResult{
			PackageID:         packageID.String(),
			LockedAt:          "2026-05-17T18:00:00Z",
			LockReason:        "final",
			SnapshotVersion:   1,
			SnapshotRowsAdded: 2,
		},
	}
	h := &CbtPackage{svc: fake}

	rec := httptest.NewRecorder()
	updateBody := `{"title":"Paket Revisi","description":"desc","duration_minutes":120,"randomize_questions":true,"randomize_options":true,"source_mode":"event_pool","draw_pg_count":20,"draw_essay_count":5,"random_seed":"seed-1","is_active":true}`
	h.Update(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPatch, "/api/cbt/packages/"+packageID.String(), updateBody), "id", packageID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Update status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateInput.ID != packageID || fake.updateInput.Title != "Paket Revisi" || fake.updateInput.DurationMinutes != 120 || !fake.updateInput.RandomizeOptions {
		t.Fatalf("Update input = %+v, want forwarded metadata", fake.updateInput)
	}

	rec = httptest.NewRecorder()
	replaceBody := `{"questions":[{"question_id":"` + questionA.String() + `","points":3},{"question_id":"` + questionB.String() + `","points":4}]}`
	h.ReplaceQuestions(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPut, "/api/cbt/packages/"+packageID.String()+"/questions", replaceBody), "id", packageID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ReplaceQuestions status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.replaceInput.PackageID != packageID || len(fake.replaceInput.QuestionIDs) != 2 || fake.replaceInput.QuestionWeights[questionA.String()] != 3 || fake.replaceInput.QuestionWeights[questionB.String()] != 4 {
		t.Fatalf("ReplaceQuestions input = %+v, weights=%v", fake.replaceInput, fake.replaceInput.QuestionWeights)
	}

	rec = httptest.NewRecorder()
	h.Clone(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+packageID.String()+"/clone", `{"title":"Paket Clone"}`), "id", packageID.String()))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Clone status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.cloneInput.SourceID != packageID || fake.cloneInput.Title != "Paket Clone" {
		t.Fatalf("Clone input = %+v, want source id and title", fake.cloneInput)
	}

	rec = httptest.NewRecorder()
	h.Lock(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+packageID.String()+"/lock", `{"reason":"final"}`), "id", packageID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Lock status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.lockPackageID != packageID || fake.lockUserID != handlerTestUUID(240) || fake.lockReason != "final" {
		t.Fatalf("Lock args package=%v user=%v reason=%q", fake.lockPackageID, fake.lockUserID, fake.lockReason)
	}
}

func TestCbtPackageHandlersRejectBadInput(t *testing.T) {
	h := &CbtPackage{svc: &fakeCbtPackageService{}}

	tests := []struct {
		name string
		run  func(*httptest.ResponseRecorder)
	}{
		{
			name: "readiness bad event id",
			run: func(rec *httptest.ResponseRecorder) {
				h.Readiness(rec, cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/readiness?event_id=bad", ""))
			},
		},
		{
			name: "get bad package id",
			run: func(rec *httptest.ResponseRecorder) {
				h.Get(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/bad", ""), "id", "bad"))
			},
		},
		{
			name: "update bad json",
			run: func(rec *httptest.ResponseRecorder) {
				h.Update(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPatch, "/api/cbt/packages/"+handlerTestUUID(66).String(), `{`), "id", handlerTestUUID(66).String()))
			},
		},
		{
			name: "replace bad question id",
			run: func(rec *httptest.ResponseRecorder) {
				h.ReplaceQuestions(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPut, "/api/cbt/packages/"+handlerTestUUID(67).String()+"/questions", `{"question_ids":["bad"]}`), "id", handlerTestUUID(67).String()))
			},
		},
		{
			name: "clone bad json",
			run: func(rec *httptest.ResponseRecorder) {
				h.Clone(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+handlerTestUUID(68).String()+"/clone", `{`), "id", handlerTestUUID(68).String()))
			},
		},
		{
			name: "lock bad json",
			run: func(rec *httptest.ResponseRecorder) {
				h.Lock(rec, cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+handlerTestUUID(69).String()+"/lock", `{`), "id", handlerTestUUID(69).String()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.run(rec)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCbtPackageQuestionInputParser(t *testing.T) {
	questionA := handlerTestUUID(70)
	questionB := handlerTestUUID(71)

	ids, weights, ok := packageQuestionInput(
		[]string{questionA.String(), questionB.String()},
		map[string]int32{questionB.String(): 5},
		nil,
	)
	if !ok || len(ids) != 2 || ids[0] != questionA || ids[1] != questionB || weights[questionB.String()] != 5 {
		t.Fatalf("packageQuestionInput(ids) = ids:%v weights:%v ok:%v", ids, weights, ok)
	}

	ids, weights, ok = packageQuestionInput(nil, nil, []struct {
		QuestionID string `json:"question_id"`
		Points     int32  `json:"points"`
	}{
		{QuestionID: questionA.String(), Points: 3},
		{QuestionID: questionB.String(), Points: 4},
	})
	if !ok || len(ids) != 2 || ids[0] != questionA || ids[1] != questionB || weights[questionA.String()] != 3 || weights[questionB.String()] != 4 {
		t.Fatalf("packageQuestionInput(questions) = ids:%v weights:%v ok:%v", ids, weights, ok)
	}

	ids, weights, ok = packageQuestionInput([]string{questionA.String()}, nil, nil)
	if !ok || len(ids) != 1 || weights == nil || len(weights) != 0 {
		t.Fatalf("packageQuestionInput(nil weights) = ids:%v weights:%v ok:%v", ids, weights, ok)
	}

	if ids, weights, ok = packageQuestionInput([]string{"not-a-uuid"}, nil, nil); ok || ids != nil || weights != nil {
		t.Fatalf("packageQuestionInput(invalid) = ids:%v weights:%v ok:%v, want failure", ids, weights, ok)
	}
}
