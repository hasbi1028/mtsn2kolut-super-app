package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeMiddlewareAnalyticsRecorder struct {
	inputs []service.CreateInternalAnalyticsEventInput
}

func (f *fakeMiddlewareAnalyticsRecorder) CreateEvent(_ context.Context, in service.CreateInternalAnalyticsEventInput) (service.InternalAnalyticsEventReceipt, error) {
	f.inputs = append(f.inputs, in)
	return service.InternalAnalyticsEventReceipt{EventName: in.EventName, EventGroup: in.EventGroup}, nil
}

func TestInternalAnalyticsMiddlewareRecordsModuleActivity(t *testing.T) {
	recorder := &fakeMiddlewareAnalyticsRecorder{}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/bank-soal/questions", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"sub":   "01000000-0000-0000-0000-000000000001",
		"roles": []any{"guru"},
	}))
	w := httptest.NewRecorder()

	InternalAnalytics(recorder)(next).ServeHTTP(w, req)

	if len(recorder.inputs) != 1 {
		t.Fatalf("analytics calls = %d, want 1", len(recorder.inputs))
	}
	got := recorder.inputs[0]
	if got.EventName != "bank_soal.question_create" || got.EventGroup != "bank_soal" || got.SourceSurface != "core_api" {
		t.Fatalf("analytics input = %+v, want bank soal create event", got)
	}
	if !got.ActorUserID.Valid || got.ActorRole != "guru" || got.Result != "success" || got.StatusCodeClass != "2xx" {
		t.Fatalf("analytics actor/status = %+v, want safe actor and success status", got)
	}
	if got.Metadata["http_method"] != "POST" || got.Metadata["resource"] != "bank_soal" {
		t.Fatalf("metadata = %+v, want aggregate-safe route metadata", got.Metadata)
	}
}

func TestInternalAnalyticsMiddlewareRecordsForbiddenWithoutRawPath(t *testing.T) {
	recorder := &fakeMiddlewareAnalyticsRecorder{}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	req := httptest.NewRequest(http.MethodPut, "/api/rbac/roles/guru/permissions?token=secret", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"roles": []any{"guru"}}))
	w := httptest.NewRecorder()

	InternalAnalytics(recorder)(next).ServeHTTP(w, req)

	if len(recorder.inputs) != 1 {
		t.Fatalf("analytics calls = %d, want 1", len(recorder.inputs))
	}
	got := recorder.inputs[0]
	if got.EventName != "rbac.permission_denied" || got.EventGroup != "rbac" || got.Result != "blocked" || got.StatusCodeClass != "4xx" {
		t.Fatalf("analytics input = %+v, want permission denied event", got)
	}
	for _, value := range got.Metadata {
		if value == "/api/rbac/roles/guru/permissions?token=secret" || value == "token=secret" {
			t.Fatalf("metadata leaked raw path/query: %+v", got.Metadata)
		}
	}
}

func TestInternalAnalyticsMiddlewareSkipsInternalAnalyticsRoutes(t *testing.T) {
	recorder := &fakeMiddlewareAnalyticsRecorder{}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/internal-analytics/events", nil)
	w := httptest.NewRecorder()

	InternalAnalytics(recorder)(next).ServeHTTP(w, req)

	if len(recorder.inputs) != 0 {
		t.Fatalf("analytics calls = %d, want 0 for self ingestion", len(recorder.inputs))
	}
}
