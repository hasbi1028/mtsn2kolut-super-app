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

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeBankSoalReviewerScopeService struct {
	listErr   error
	upsertErr error
	deleteErr error

	upsertInput service.BankSoalReviewerScopeInput
	deleteID    pgtype.UUID
}

func (f *fakeBankSoalReviewerScopeService) List(context.Context) ([]db.ListBankSoalReviewerScopesRow, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListBankSoalReviewerScopesRow{{
		ID:                    handlerTestUUID(11),
		UserID:                handlerTestUUID(12),
		Username:              "reviewer.ipa",
		UserDisplayName:       "Reviewer IPA",
		UserRoles:             []byte(`["guru","reviewer"]`),
		SubjectID:             handlerTestUUID(13),
		SubjectName:           "IPA",
		SubjectCode:           "IPA",
		GradeLevel:            pgtype.Int2{Int16: 7, Valid: true},
		CanReview:             true,
		CanApprove:            false,
		AssignedBy:            handlerTestUUID(14),
		AssignedByDisplayName: "Admin",
	}}, nil
}

func (f *fakeBankSoalReviewerScopeService) Upsert(_ context.Context, input service.BankSoalReviewerScopeInput) (db.UpsertBankSoalReviewerScopeRow, error) {
	f.upsertInput = input
	if f.upsertErr != nil {
		return db.UpsertBankSoalReviewerScopeRow{}, f.upsertErr
	}
	return db.UpsertBankSoalReviewerScopeRow{
		ID:         handlerTestUUID(15),
		UserID:     input.UserID,
		SubjectID:  input.SubjectID,
		GradeLevel: input.GradeLevel,
		CanReview:  input.CanReview,
		CanApprove: input.CanApprove,
		AssignedBy: input.AssignedBy,
	}, nil
}

func (f *fakeBankSoalReviewerScopeService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func TestBankSoalReviewerScopeListUpsertDeleteSuccess(t *testing.T) {
	fake := &fakeBankSoalReviewerScopeService{}
	h := &BankSoalReviewerScope{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/bank-soal/reviewer-scopes", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "reviewer.ipa") || !strings.Contains(rec.Body.String(), "user_roles") {
		t.Fatalf("List() status/body = %d/%s, want serialized scope", rec.Code, rec.Body.String())
	}

	userID := handlerTestUUID(21)
	subjectID := handlerTestUUID(22)
	actorID := handlerTestUUID(23)
	body := `{"user_id":"` + userID.String() + `","subject_id":"` + subjectID.String() + `","grade_level":8,"can_review":false,"can_approve":true}`
	rec = httptest.NewRecorder()
	h.Upsert(rec, withClaims(httptest.NewRequest(http.MethodPost, "/api/bank-soal/reviewer-scopes", strings.NewReader(body)), jwt.MapClaims{"roles": []any{"admin"}, "uid": actorID.String()}))
	if rec.Code != http.StatusOK {
		t.Fatalf("Upsert() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertInput.UserID != userID || fake.upsertInput.SubjectID != subjectID || fake.upsertInput.GradeLevel.Int16 != 8 || fake.upsertInput.CanReview || !fake.upsertInput.CanApprove || fake.upsertInput.AssignedBy != actorID {
		t.Fatalf("Upsert() input = %+v, want mapped ids/defaults/actor", fake.upsertInput)
	}

	scopeID := handlerTestUUID(24)
	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/bank-soal/reviewer-scopes/"+scopeID.String(), ""), "id", scopeID.String()))
	if rec.Code != http.StatusNoContent || fake.deleteID != scopeID {
		t.Fatalf("Delete() status/id = %d/%v, want 204/%v", rec.Code, fake.deleteID, scopeID)
	}
}

func TestBankSoalReviewerScopeAuthValidationAndErrors(t *testing.T) {
	scopeID := handlerTestUUID(31)
	permissionReq := func(method, target, body string) *http.Request {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		return withClaims(req, jwt.MapClaims{"permissions": []any{"bank_soal.assign_reviewer"}, "uid": handlerTestUUID(32).String()})
	}
	tests := []struct {
		name string
		fn   func(*BankSoalReviewerScope, http.ResponseWriter, *http.Request)
		svc  *fakeBankSoalReviewerScopeService
		req  *http.Request
		want int
	}{
		{name: "list forbidden", fn: (*BankSoalReviewerScope).List, svc: &fakeBankSoalReviewerScopeService{}, req: httptest.NewRequest(http.MethodGet, "/api/bank-soal/reviewer-scopes", nil), want: http.StatusForbidden},
		{name: "list service error", fn: (*BankSoalReviewerScope).List, svc: &fakeBankSoalReviewerScopeService{listErr: errors.New("db down")}, req: adminRequest(http.MethodGet, "/api/bank-soal/reviewer-scopes", ""), want: http.StatusInternalServerError},
		{name: "upsert forbidden", fn: (*BankSoalReviewerScope).Upsert, svc: &fakeBankSoalReviewerScopeService{}, req: httptest.NewRequest(http.MethodPost, "/api/bank-soal/reviewer-scopes", strings.NewReader(`{}`)), want: http.StatusForbidden},
		{name: "upsert invalid json", fn: (*BankSoalReviewerScope).Upsert, svc: &fakeBankSoalReviewerScopeService{}, req: adminRequest(http.MethodPost, "/api/bank-soal/reviewer-scopes", `{`), want: http.StatusBadRequest},
		{name: "upsert invalid user", fn: (*BankSoalReviewerScope).Upsert, svc: &fakeBankSoalReviewerScopeService{}, req: adminRequest(http.MethodPost, "/api/bank-soal/reviewer-scopes", `{"user_id":"bad"}`), want: http.StatusBadRequest},
		{name: "upsert invalid subject", fn: (*BankSoalReviewerScope).Upsert, svc: &fakeBankSoalReviewerScopeService{}, req: adminRequest(http.MethodPost, "/api/bank-soal/reviewer-scopes", `{"user_id":"`+handlerTestUUID(33).String()+`","subject_id":"bad"}`), want: http.StatusBadRequest},
		{name: "upsert missing actor", fn: (*BankSoalReviewerScope).Upsert, svc: &fakeBankSoalReviewerScopeService{}, req: withClaims(httptest.NewRequest(http.MethodPost, "/api/bank-soal/reviewer-scopes", strings.NewReader(`{"user_id":"`+handlerTestUUID(34).String()+`"}`)), jwt.MapClaims{"roles": []any{"admin"}}), want: http.StatusUnauthorized},
		{name: "upsert service client error", fn: (*BankSoalReviewerScope).Upsert, svc: &fakeBankSoalReviewerScopeService{upsertErr: errors.New("scope duplikat")}, req: permissionReq(http.MethodPost, "/api/bank-soal/reviewer-scopes", `{"user_id":"`+handlerTestUUID(35).String()+`"}`), want: http.StatusBadRequest},
		{name: "delete forbidden", fn: (*BankSoalReviewerScope).Delete, svc: &fakeBankSoalReviewerScopeService{}, req: withRouteParam(httptest.NewRequest(http.MethodDelete, "/api/bank-soal/reviewer-scopes/"+scopeID.String(), nil), "id", scopeID.String()), want: http.StatusForbidden},
		{name: "delete invalid id", fn: (*BankSoalReviewerScope).Delete, svc: &fakeBankSoalReviewerScopeService{}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/bank-soal/reviewer-scopes/bad", ""), "id", "bad"), want: http.StatusBadRequest},
		{name: "delete service client error", fn: (*BankSoalReviewerScope).Delete, svc: &fakeBankSoalReviewerScopeService{deleteErr: errors.New("scope tidak ditemukan")}, req: withRouteParam(permissionReq(http.MethodDelete, "/api/bank-soal/reviewer-scopes/"+scopeID.String(), ""), "id", scopeID.String()), want: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(&BankSoalReviewerScope{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestReviewerScopeHelpers(t *testing.T) {
	input, err := reviewerScopeInputFromBody(bankSoalReviewerScopeBody{UserID: handlerTestUUID(41).String()})
	if err != nil {
		t.Fatalf("reviewerScopeInputFromBody(defaults) error = %v", err)
	}
	if !input.CanReview || input.CanApprove || input.SubjectID.Valid || input.GradeLevel.Valid {
		t.Fatalf("reviewerScopeInputFromBody(defaults) = %+v, want review true/approve false/global scope", input)
	}
	if got := pgInt2Value(pgtype.Int2{}); got != nil {
		t.Fatalf("pgInt2Value(invalid) = %v, want nil", got)
	}
	if got := pgInt2Value(pgtype.Int2{Int16: 9, Valid: true}); got != int16(9) {
		t.Fatalf("pgInt2Value(valid) = %v, want 9", got)
	}
	if !bankSoalReviewerScopeAccessAllowed(withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"permissions": []any{"bank_soal.settings"}})) {
		t.Fatal("bankSoalReviewerScopeAccessAllowed(permission) = false, want true")
	}
	if bankSoalReviewerScopeAccessAllowed(httptest.NewRequest(http.MethodGet, "/", nil)) {
		t.Fatal("bankSoalReviewerScopeAccessAllowed(no claims) = true, want false")
	}

	// Keep api import used in this package-specific test file while also verifying claims shape.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"roles": []any{"admin"}}))
	if !bankSoalReviewerScopeAccessAllowed(req) {
		t.Fatal("bankSoalReviewerScopeAccessAllowed(api claims) = false, want true")
	}
}
