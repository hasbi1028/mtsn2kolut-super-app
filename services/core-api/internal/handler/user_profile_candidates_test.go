package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeUserProfileCandidatesService struct {
	filter service.UserProfileCandidateFilter
	result service.UserProfileCandidatesResult
	err    error
}

func (f *fakeUserProfileCandidatesService) List(ctx context.Context, filter service.UserProfileCandidateFilter) (service.UserProfileCandidatesResult, error) {
	f.filter = filter
	return f.result, f.err
}

func TestUserProfileCandidatesHandlerAuthorizesAndForwardsQuery(t *testing.T) {
	classID := "03000000-0000-0000-0000-000000000000"
	svc := &fakeUserProfileCandidatesService{
		result: service.UserProfileCandidatesResult{
			Role:       "siswa",
			Profile:    "student",
			ClassID:    classID,
			Candidates: []service.UserProfileCandidate{{ID: "student-1", ProfileType: "student", Nama: "Ahmad"}},
		},
	}
	h := &User{profileCandidates: svc}

	rec := httptest.NewRecorder()
	req := accountGenerationRequest(http.MethodGet, "/api/users/profile-candidates?role=siswa&class_id="+classID+"&q=ahmad&include_linked=true&limit=75", "01000000-0000-0000-0000-000000000222", "users.create")
	h.ListProfileCandidates(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ListProfileCandidates() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.filter.Role != "siswa" || svc.filter.ClassID != classID || svc.filter.Query != "ahmad" || !svc.filter.IncludeLinked || svc.filter.Limit != 75 {
		t.Fatalf("forwarded filter = %+v", svc.filter)
	}
	if strings.Contains(rec.Body.String(), "password") || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("ListProfileCandidates() response exposes credential-like field: %s", rec.Body.String())
	}
	var payload struct {
		Data service.UserProfileCandidatesResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response json error = %v", err)
	}
	if payload.Data.Profile != "student" || len(payload.Data.Candidates) != 1 {
		t.Fatalf("response data = %+v", payload.Data)
	}
}

func TestUserProfileCandidatesHandlerRejectsUnauthorizedAndBadFilters(t *testing.T) {
	svc := &fakeUserProfileCandidatesService{err: domain.ErrBadRequest}
	h := &User{profileCandidates: svc}

	rec := httptest.NewRecorder()
	h.ListProfileCandidates(rec, httptest.NewRequest(http.MethodGet, "/api/users/profile-candidates?role=siswa", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ListProfileCandidates(rec, accountGenerationRequest(http.MethodGet, "/api/users/profile-candidates?role=siswa", "01000000-0000-0000-0000-000000000223", "students.read"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("missing users.read status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ListProfileCandidates(rec, accountGenerationRequest(http.MethodGet, "/api/users/profile-candidates?role=siswa", "01000000-0000-0000-0000-000000000224", "users.create"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad filter status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	svc.err = errors.New("db down")
	rec = httptest.NewRecorder()
	h.ListProfileCandidates(rec, accountGenerationRequest(http.MethodGet, "/api/users/profile-candidates?role=guru", "01000000-0000-0000-0000-000000000225", "users.create"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("service error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}
