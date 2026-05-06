package handler

import (
	"bytes"
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

type fakeProfileChangeRequestService struct {
	createResult db.ProfileChangeRequest
	createErr    error
	createUserID pgtype.UUID
	createInput  service.CreateProfileChangeRequestInput

	listOwnResult []service.ProfileChangeRequestListItem
	listOwnErr    error
	listOwnUserID pgtype.UUID

	fieldResult []service.ProfileChangeRequestField
	fieldErr    error
	fieldUserID pgtype.UUID

	cancelResult db.ProfileChangeRequest
	cancelErr    error
	cancelUserID pgtype.UUID
	cancelID     pgtype.UUID

	listAdminResult  []service.ProfileChangeRequestListItem
	listAdminErr     error
	listAdminFilter  service.ProfileChangeRequestListFilter
	listAdminLimit   int32
	listAdminOffset  int32
	countAdminResult int32
	countAdminErr    error
	countAdminFilter service.ProfileChangeRequestListFilter

	reviewResult db.ProfileChangeRequest
	reviewErr    error
	reviewerID   pgtype.UUID
	reviewID     pgtype.UUID
	reviewInput  service.ReviewProfileChangeRequestInput
}

func (f *fakeProfileChangeRequestService) Create(ctx context.Context, requesterUserID pgtype.UUID, input service.CreateProfileChangeRequestInput) (db.ProfileChangeRequest, error) {
	f.createUserID = requesterUserID
	f.createInput = input
	return f.createResult, f.createErr
}

func (f *fakeProfileChangeRequestService) ListOwn(ctx context.Context, requesterUserID pgtype.UUID) ([]service.ProfileChangeRequestListItem, error) {
	f.listOwnUserID = requesterUserID
	return f.listOwnResult, f.listOwnErr
}

func (f *fakeProfileChangeRequestService) ListSelfRequestableFields(ctx context.Context, requesterUserID pgtype.UUID) ([]service.ProfileChangeRequestField, error) {
	f.fieldUserID = requesterUserID
	return f.fieldResult, f.fieldErr
}

func (f *fakeProfileChangeRequestService) CancelOwn(ctx context.Context, requesterUserID, id pgtype.UUID) (db.ProfileChangeRequest, error) {
	f.cancelUserID = requesterUserID
	f.cancelID = id
	return f.cancelResult, f.cancelErr
}

func (f *fakeProfileChangeRequestService) ListAdmin(ctx context.Context, filter service.ProfileChangeRequestListFilter, limit, offset int32) ([]service.ProfileChangeRequestListItem, error) {
	f.listAdminFilter = filter
	f.listAdminLimit = limit
	f.listAdminOffset = offset
	return f.listAdminResult, f.listAdminErr
}

func (f *fakeProfileChangeRequestService) CountAdmin(ctx context.Context, filter service.ProfileChangeRequestListFilter) (int32, error) {
	f.countAdminFilter = filter
	return f.countAdminResult, f.countAdminErr
}

func (f *fakeProfileChangeRequestService) Review(ctx context.Context, reviewerUserID, id pgtype.UUID, input service.ReviewProfileChangeRequestInput) (db.ProfileChangeRequest, error) {
	f.reviewerID = reviewerUserID
	f.reviewID = id
	f.reviewInput = input
	return f.reviewResult, f.reviewErr
}

func TestProfileChangeRequestCreateOwnForwardsJWTUserAndBody(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	requestID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	employeeID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	svc := &fakeProfileChangeRequestService{
		createResult: db.ProfileChangeRequest{
			ID:               requestID,
			RequesterUserID:  mustUUID(t, userID),
			ProfileType:      "employee",
			TargetEmployeeID: employeeID,
			FieldKey:         "nama",
			CurrentValue:     "Lama",
			RequestedValue:   "Baru",
			Reason:           "Dokumen resmi",
			Status:           db.ProfileChangeRequestStatusPending,
		},
	}
	h := NewProfileChangeRequest(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/account/change-requests", bytes.NewBufferString(`{"profile_type":"employee","field_key":"name","requested_value":"Baru","reason":"Dokumen resmi"}`))
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.CreateOwn(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if svc.createUserID != mustUUID(t, userID) || svc.createInput.FieldKey != "name" || svc.createInput.RequestedValue != "Baru" {
		t.Fatalf("create args = %v/%+v, want JWT user/body", svc.createUserID, svc.createInput)
	}
	for _, expected := range []string{requestID.String(), employeeID.String(), `"status":"pending"`, `"field_key":"nama"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("body = %s, want %q", rec.Body.String(), expected)
		}
	}
}

func TestProfileChangeRequestSelfRoutesMapUnauthorizedAndCancel(t *testing.T) {
	h := NewProfileChangeRequest(&fakeProfileChangeRequestService{})
	rec := httptest.NewRecorder()
	h.ListOwn(rec, httptest.NewRequest(http.MethodGet, "/api/auth/account/change-requests", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ListOwn(no claims) status = %d, want 401", rec.Code)
	}

	userID := "11111111-1111-1111-1111-111111111111"
	requestID := "22222222-2222-2222-2222-222222222222"
	svc := &fakeProfileChangeRequestService{
		cancelResult: db.ProfileChangeRequest{
			ID:              mustUUID(t, requestID),
			RequesterUserID: mustUUID(t, userID),
			ProfileType:     "student",
			TargetStudentID: mustUUID(t, "33333333-3333-3333-3333-333333333333"),
			FieldKey:        "tanggal_lahir",
			Status:          db.ProfileChangeRequestStatusCancelled,
		},
	}
	h = NewProfileChangeRequest(svc)
	req := withRouteParam(httptest.NewRequest(http.MethodPost, "/api/auth/account/change-requests/"+requestID+"/cancel", nil), "id", requestID)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec = httptest.NewRecorder()
	h.CancelOwn(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("CancelOwn() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.cancelUserID != mustUUID(t, userID) || svc.cancelID != mustUUID(t, requestID) {
		t.Fatalf("cancel args = %v/%v, want JWT user/request id", svc.cancelUserID, svc.cancelID)
	}
}

func TestProfileChangeRequestListSelfRequestableFields(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	svc := &fakeProfileChangeRequestService{
		fieldResult: []service.ProfileChangeRequestField{{
			ProfileType:        "student",
			FieldKey:           "tanggal_lahir",
			Label:              "Tanggal lahir",
			ValueType:          "date",
			SelfRequestable:    true,
			ReviewerPermission: service.ProfileChangesReviewPermission,
			IsActive:           true,
		}},
	}
	h := NewProfileChangeRequest(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/account/change-request-fields", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.ListSelfRequestableFields(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ListSelfRequestableFields() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.fieldUserID != mustUUID(t, userID) {
		t.Fatalf("field user id = %v, want JWT user", svc.fieldUserID)
	}
	for _, expected := range []string{`"profile_type":"student"`, `"field_key":"tanggal_lahir"`, `"value_type":"date"`, `"reviewer_permission":"profile_changes.review"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("body = %s, want %q", rec.Body.String(), expected)
		}
	}
}

func TestProfileChangeRequestListAdminForwardsFiltersAndShapesRows(t *testing.T) {
	requestID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	svc := &fakeProfileChangeRequestService{
		listAdminResult: []service.ProfileChangeRequestListItem{{
			ID:                   requestID,
			RequesterUserID:      mustUUID(t, "11111111-1111-1111-1111-111111111111"),
			RequesterUsername:    "guru.ipa",
			RequesterDisplayName: "Guru IPA",
			ProfileType:          "employee",
			ProfileNama:          "Guru IPA",
			FieldKey:             "nama",
			FieldLabel:           "Nama resmi",
			CurrentValue:         "Lama",
			RequestedValue:       "Baru",
			Status:               db.ProfileChangeRequestStatusPending,
		}},
	}
	h := NewProfileChangeRequest(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/users/change-requests?status=pending&profile_type=employee&field=nama&search=guru&page=2&per_page=10", nil)
	rec := httptest.NewRecorder()

	h.ListAdmin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ListAdmin() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.listAdminFilter.Status != "pending" || svc.listAdminFilter.ProfileType != "employee" || svc.listAdminFilter.FieldKey != "nama" || svc.listAdminFilter.Search != "guru" || svc.listAdminLimit != 10 || svc.listAdminOffset != 10 {
		t.Fatalf("list args = %+v/%d/%d, want filters/10/10", svc.listAdminFilter, svc.listAdminLimit, svc.listAdminOffset)
	}
	if !strings.Contains(rec.Body.String(), `"requester_username":"guru.ipa"`) || !strings.Contains(rec.Body.String(), `"field_label":"Nama resmi"`) || !strings.Contains(rec.Body.String(), requestID.String()) {
		t.Fatalf("body = %s, want requester/field label/request id", rec.Body.String())
	}
}

func TestProfileChangeRequestPendingCountAndExport(t *testing.T) {
	requestID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	svc := &fakeProfileChangeRequestService{
		countAdminResult: 7,
		listAdminResult: []service.ProfileChangeRequestListItem{{
			ID:                   requestID,
			RequesterUserID:      mustUUID(t, "11111111-1111-1111-1111-111111111111"),
			RequesterUsername:    "guru.ipa",
			RequesterDisplayName: "Guru IPA",
			ProfileType:          "student",
			ProfileNama:          "Ahmad Fauzi",
			FieldKey:             "tanggal_lahir",
			FieldLabel:           "Tanggal lahir",
			CurrentValue:         "2010-01-02",
			RequestedValue:       "2011-03-04",
			Reason:               "=Nomor dokumen 123456789",
			Status:               db.ProfileChangeRequestStatusPending,
		}},
	}
	h := NewProfileChangeRequest(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/users/change-requests/pending-count?profile_type=student", nil)
	rec := httptest.NewRecorder()
	h.CountPendingAdmin(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"pending":7`) {
		t.Fatalf("CountPendingAdmin() status=%d body=%s, want pending count", rec.Code, rec.Body.String())
	}
	if svc.countAdminFilter.Status != "pending" || svc.countAdminFilter.ProfileType != "student" {
		t.Fatalf("count filter = %+v, want pending student", svc.countAdminFilter)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/users/change-requests/export?status=pending", nil)
	rec = httptest.NewRecorder()
	h.ExportAdminCSV(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("ExportAdminCSV() status=%d content-type=%q body=%s", rec.Code, rec.Header().Get("Content-Type"), body)
	}
	for _, expected := range []string{"request_id,status,created_at", "Tanggal lahir", "2010-**-**", "2011-**-**", "'=Nomor dokumen ****"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("export body = %s, want %q", body, expected)
		}
	}
	if strings.Contains(body, "123456789") {
		t.Fatalf("export body = %s, contains unmasked document number", body)
	}
}

func TestProfileChangeRequestReviewForwardsReviewerAndMapsErrors(t *testing.T) {
	reviewerID := "11111111-1111-1111-1111-111111111111"
	requestID := "22222222-2222-2222-2222-222222222222"
	svc := &fakeProfileChangeRequestService{
		reviewResult: db.ProfileChangeRequest{
			ID:              mustUUID(t, requestID),
			RequesterUserID: mustUUID(t, "33333333-3333-3333-3333-333333333333"),
			ProfileType:     "parent",
			TargetParentID:  mustUUID(t, "44444444-4444-4444-4444-444444444444"),
			FieldKey:        "nama",
			Status:          db.ProfileChangeRequestStatusApproved,
		},
	}
	h := NewProfileChangeRequest(svc)
	req := withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/users/change-requests/"+requestID, bytes.NewBufferString(`{"status":"approved","review_note":"sesuai"}`)), "id", requestID)
	req = withClaims(req, jwt.MapClaims{"sub": reviewerID, "roles": []any{"admin"}, "permissions": []any{service.ProfileChangesReviewPermission}})
	rec := httptest.NewRecorder()

	h.Review(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Review() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.reviewerID != mustUUID(t, reviewerID) || svc.reviewID != mustUUID(t, requestID) || svc.reviewInput.Status != "approved" {
		t.Fatalf("review args = %v/%v/%+v", svc.reviewerID, svc.reviewID, svc.reviewInput)
	}
	if len(svc.reviewInput.ReviewerRoles) != 1 || svc.reviewInput.ReviewerRoles[0] != "admin" || len(svc.reviewInput.ReviewerPermissions) != 1 || svc.reviewInput.ReviewerPermissions[0] != service.ProfileChangesReviewPermission {
		t.Fatalf("review access claims = roles %v permissions %v", svc.reviewInput.ReviewerRoles, svc.reviewInput.ReviewerPermissions)
	}

	req = withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/users/change-requests/"+requestID, bytes.NewBufferString(`{"status":"approved"}`)), "id", requestID)
	req = req.WithContext(withAuthClaims(req.Context(), reviewerID))
	rec = httptest.NewRecorder()
	NewProfileChangeRequest(&fakeProfileChangeRequestService{reviewErr: domain.ErrConflict}).Review(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Review(conflict) status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}

	req = withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/users/change-requests/"+requestID, bytes.NewBufferString(`{"status":"approved"}`)), "id", requestID)
	req = req.WithContext(withAuthClaims(req.Context(), reviewerID))
	rec = httptest.NewRecorder()
	NewProfileChangeRequest(&fakeProfileChangeRequestService{reviewErr: errors.New("db down")}).Review(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Review(internal) status = %d, want 500", rec.Code)
	}
}
