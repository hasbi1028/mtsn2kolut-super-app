package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestProfileChangeRequestExtraCancelOwnLowBranches(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	requestID := "22222222-2222-2222-2222-222222222222"

	t.Run("invalid route id is bad request before service", func(t *testing.T) {
		svc := &fakeProfileChangeRequestService{}
		h := NewProfileChangeRequest(svc)
		req := withRouteParam(httptest.NewRequest(http.MethodPost, "/api/auth/account/change-requests/bad/cancel", nil), "id", "bad")
		req = req.WithContext(withAuthClaims(req.Context(), userID))
		rec := httptest.NewRecorder()

		h.CancelOwn(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("CancelOwn(invalid id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if svc.cancelUserID.Valid || svc.cancelID.Valid {
			t.Fatalf("CancelOwn service called despite invalid id: user=%v id=%v", svc.cancelUserID, svc.cancelID)
		}
	})

	t.Run("domain error maps through client error", func(t *testing.T) {
		svc := &fakeProfileChangeRequestService{cancelErr: domain.ErrConflict}
		h := NewProfileChangeRequest(svc)
		req := withRouteParam(httptest.NewRequest(http.MethodPost, "/api/auth/account/change-requests/"+requestID+"/cancel", nil), "id", requestID)
		req = req.WithContext(withAuthClaims(req.Context(), userID))
		rec := httptest.NewRecorder()

		h.CancelOwn(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("CancelOwn(conflict) status = %d, want 409; body=%s", rec.Code, rec.Body.String())
		}
		if svc.cancelUserID != mustUUID(t, userID) || svc.cancelID != mustUUID(t, requestID) {
			t.Fatalf("CancelOwn args = %v/%v, want JWT user/request id", svc.cancelUserID, svc.cancelID)
		}
	})
}

func TestProfileChangeRequestExtraListSelfRequestableFieldsErrorsAndShape(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"

	t.Run("unauthorized without current user", func(t *testing.T) {
		h := NewProfileChangeRequest(&fakeProfileChangeRequestService{})
		rec := httptest.NewRecorder()

		h.ListSelfRequestableFields(rec, httptest.NewRequest(http.MethodGet, "/api/auth/account/change-request-fields", nil))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("ListSelfRequestableFields(no claims) status = %d, want 401", rec.Code)
		}
	})

	t.Run("domain error uses client mapping", func(t *testing.T) {
		svc := &fakeProfileChangeRequestService{fieldErr: domain.ErrForbidden}
		h := NewProfileChangeRequest(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/auth/account/change-request-fields", nil)
		req = req.WithContext(withAuthClaims(req.Context(), userID))
		rec := httptest.NewRecorder()

		h.ListSelfRequestableFields(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("ListSelfRequestableFields(forbidden) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("inactive field shape is preserved", func(t *testing.T) {
		svc := &fakeProfileChangeRequestService{fieldResult: []service.ProfileChangeRequestField{{
			ProfileType:     "employee",
			FieldKey:        "no_hp",
			Label:           "Nomor HP",
			ValueType:       "text",
			SelfRequestable: false,
			IsActive:        false,
		}}}
		h := NewProfileChangeRequest(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/auth/account/change-request-fields", nil)
		req = req.WithContext(withAuthClaims(req.Context(), userID))
		rec := httptest.NewRecorder()

		h.ListSelfRequestableFields(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("ListSelfRequestableFields() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		for _, expected := range []string{`"profile_type":"employee"`, `"field_key":"no_hp"`, `"self_requestable":false`, `"is_active":false`} {
			if !strings.Contains(rec.Body.String(), expected) {
				t.Fatalf("body = %s, want %q", rec.Body.String(), expected)
			}
		}
	})
}

func TestCurrentClaimStringsAcceptsArrayAndScalarClaims(t *testing.T) {
	req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{
		"roles":       []any{"admin", "", 7, "staf"},
		"role":        "kesiswaan",
		"permissions": []string{"profile_changes.review", "  ", "users.read"},
		"permission":  "users.create",
	})

	if got, want := currentClaimStrings(req, "roles", "role"), []string{"admin", "staf", "kesiswaan"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("role claim strings = %#v, want %#v", got, want)
	}
	if got, want := currentClaimStrings(req, "permissions", "permission"), []string{"profile_changes.review", "users.read", "users.create"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("permission claim strings = %#v, want %#v", got, want)
	}

	if got := currentClaimStrings(httptest.NewRequest(http.MethodGet, "/", nil), "roles", "role"); got != nil {
		t.Fatalf("currentClaimStrings(no claims) = %#v, want nil", got)
	}
}

func TestProfileChangeRequestExtraReviewScalarClaimsAndBadInput(t *testing.T) {
	reviewerID := "11111111-1111-1111-1111-111111111111"
	requestID := "22222222-2222-2222-2222-222222222222"

	svc := &fakeProfileChangeRequestService{reviewResult: db.ProfileChangeRequest{
		ID:              mustUUID(t, requestID),
		RequesterUserID: mustUUID(t, "33333333-3333-3333-3333-333333333333"),
		ProfileType:     "employee",
		FieldKey:        "nama",
		Status:          db.ProfileChangeRequestStatusRejected,
	}}
	h := NewProfileChangeRequest(svc)
	req := withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/users/change-requests/"+requestID, bytes.NewBufferString(`{"status":"rejected","review_note":"tidak sesuai"}`)), "id", requestID)
	req = withClaims(req, jwt.MapClaims{"sub": reviewerID, "role": "admin", "permission": service.ProfileChangesReviewPermission})
	rec := httptest.NewRecorder()

	h.Review(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Review() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if len(svc.reviewInput.ReviewerRoles) != 1 || svc.reviewInput.ReviewerRoles[0] != "admin" || len(svc.reviewInput.ReviewerPermissions) != 1 || svc.reviewInput.ReviewerPermissions[0] != service.ProfileChangesReviewPermission {
		t.Fatalf("Review scalar claims = roles %v permissions %v", svc.reviewInput.ReviewerRoles, svc.reviewInput.ReviewerPermissions)
	}

	rec = httptest.NewRecorder()
	h.Review(rec, withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/users/change-requests/bad", bytes.NewBufferString(`{"status":"approved"}`)), "id", "bad"))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Review(no claims before bad id) status = %d, want 401", rec.Code)
	}

	req = withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/users/change-requests/"+requestID, bytes.NewBufferString(`{`)), "id", requestID)
	req = req.WithContext(withAuthClaims(req.Context(), reviewerID))
	rec = httptest.NewRecorder()
	h.Review(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Review(bad json) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}
