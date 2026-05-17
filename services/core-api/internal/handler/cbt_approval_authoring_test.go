package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeCbtApprovalService struct {
	listInput   service.ListCbtApprovalInput
	listRows    []db.ListCbtApprovalRecordsRow
	listErr     error
	approveIn   service.SaveCbtApprovalInput
	approveRow  db.CbtApprovalRecord
	approveErr  error
	revokeID    pgtype.UUID
	revokeActor pgtype.UUID
	revokeNotes string
	revokeRow   db.CbtApprovalRecord
	revokeErr   error
}

func (f *fakeCbtApprovalService) List(ctx context.Context, in service.ListCbtApprovalInput) ([]db.ListCbtApprovalRecordsRow, error) {
	f.listInput = in
	return f.listRows, f.listErr
}

func (f *fakeCbtApprovalService) Approve(ctx context.Context, in service.SaveCbtApprovalInput) (db.CbtApprovalRecord, error) {
	f.approveIn = in
	return f.approveRow, f.approveErr
}

func (f *fakeCbtApprovalService) Revoke(ctx context.Context, id pgtype.UUID, actorUserID pgtype.UUID, notes string) (db.CbtApprovalRecord, error) {
	f.revokeID = id
	f.revokeActor = actorUserID
	f.revokeNotes = notes
	return f.revokeRow, f.revokeErr
}

func decodeHandlerResponse(t *testing.T, rec *httptest.ResponseRecorder) api.Response {
	t.Helper()
	var out api.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return out
}

func TestCbtApprovalHandlerListParsesFiltersAndRequiresAdmin(t *testing.T) {
	entityID := handlerTestUUID(111)
	fake := &fakeCbtApprovalService{listRows: []db.ListCbtApprovalRecordsRow{{ID: handlerTestUUID(112), EntityType: "event", EntityID: entityID}}}
	h := &CbtApproval{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/cbt/approvals?entity_type=event&entity_id="+entityID.String(), ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listInput.EntityType != "event" || fake.listInput.EntityID != entityID {
		t.Fatalf("List() input = %+v, want entity filter", fake.listInput)
	}

	rec = httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/cbt/approvals", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("List(no admin) status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/cbt/approvals?entity_id=bad", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("List(bad entity_id) status = %d, want 400", rec.Code)
	}
}

func TestCbtApprovalHandlerApproveParsesActorBodyAndMapsErrors(t *testing.T) {
	entityID := handlerTestUUID(121)
	actorID := handlerTestUUID(1) // adminRequest uid is 01000000-0000-0000-0000-000000000000.
	fake := &fakeCbtApprovalService{approveRow: db.CbtApprovalRecord{EntityID: entityID, ApprovalType: "package_ready"}}
	h := &CbtApproval{svc: fake}

	rec := httptest.NewRecorder()
	h.Approve(rec, adminRequest(http.MethodPost, "/api/cbt/approvals", `{"entity_type":"package","entity_id":"`+entityID.String()+`","approval_type":"package_ready","notes":"siap"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("Approve() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.approveIn.EntityType != "package" || fake.approveIn.EntityID != entityID || fake.approveIn.ApprovalType != "package_ready" || fake.approveIn.Notes != "siap" || fake.approveIn.ActorUserID != actorID {
		t.Fatalf("Approve() input = %+v, want parsed body and actor", fake.approveIn)
	}

	rec = httptest.NewRecorder()
	h.Approve(rec, adminRequest(http.MethodPost, "/api/cbt/approvals", `{"entity_id":"bad"}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Approve(bad entity_id) status = %d, want 400", rec.Code)
	}

	fake.approveErr = domain.ErrConflict
	rec = httptest.NewRecorder()
	h.Approve(rec, adminRequest(http.MethodPost, "/api/cbt/approvals", `{"entity_type":"package","entity_id":"`+entityID.String()+`","approval_type":"package_ready"}`))
	if rec.Code != http.StatusConflict {
		t.Fatalf("Approve(conflict) status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtApprovalHandlerRevokeParsesRouteActorAndNotes(t *testing.T) {
	approvalID := handlerTestUUID(131)
	actorID := handlerTestUUID(1)
	fake := &fakeCbtApprovalService{revokeRow: db.CbtApprovalRecord{ID: approvalID, RevokedBy: actorID}}
	h := &CbtApproval{svc: fake}

	rec := httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPost, "/api/cbt/approvals/"+approvalID.String()+"/revoke", `{"notes":" cabut "}`), "id", approvalID.String())
	h.Revoke(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Revoke() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.revokeID != approvalID || fake.revokeActor != actorID || fake.revokeNotes != " cabut " {
		t.Fatalf("Revoke() args id=%v actor=%v notes=%q, want route id, actor, raw notes", fake.revokeID, fake.revokeActor, fake.revokeNotes)
	}

	rec = httptest.NewRecorder()
	h.Revoke(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/approvals/bad/revoke", `{}`), "id", "bad"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Revoke(bad id) status = %d, want 400", rec.Code)
	}

	fake.revokeErr = domain.ErrNotFound
	rec = httptest.NewRecorder()
	h.Revoke(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/approvals/"+approvalID.String()+"/revoke", `{}`), "id", approvalID.String()))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Revoke(not found) status = %d, want 404", rec.Code)
	}
}

func TestQuestionAuthoringAndSummaryHelpers(t *testing.T) {
	for value, want := range map[int16]string{7: "VII", 8: "VIII", 9: "IX", 6: "", 0: ""} {
		if got := legacyQuestionTargetLevelFromInt(value); got != want {
			t.Fatalf("legacyQuestionTargetLevelFromInt(%d) = %q, want %q", value, got, want)
		}
	}

	authorReq := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru1"})
	if !questionAnswerKeyAllowed(authorReq, " guru1 ") {
		t.Fatalf("questionAnswerKeyAllowed(author) = false, want true")
	}
	if questionAnswerKeyAllowed(authorReq, "other") {
		t.Fatalf("questionAnswerKeyAllowed(other author) = true, want false")
	}
	if !questionAnswerKeyAllowed(adminRequest(http.MethodGet, "/", ""), "other") {
		t.Fatalf("questionAnswerKeyAllowed(admin) = false, want true")
	}

	versionID := handlerTestUUID(141)
	serialized := serializeQuestionVersionRow(db.ListCbtQuestionVersionsRow{ID: versionID, Code: "Q-1", WorkflowStatus: "approved", Status: db.CbtQuestionStatusEnumPublished, VersionNumber: 2, IsLatestVersion: true, AuthorUsername: "guru1"})
	if serialized["id"] != versionID.String() || serialized["version_number"] != int32(2) || serialized["is_latest_version"] != true || serialized["status"] != db.CbtQuestionStatusEnumPublished {
		t.Fatalf("serializeQuestionVersionRow() = %+v, want version fields", serialized)
	}
}

func TestCbtQuestionSummaryHandlerSerializesAndPassesActor(t *testing.T) {
	fake := &fakeCbtQuestionService{summaryResult: service.CbtQuestionSummary{
		Counts:           db.GetCbtQuestionSummaryCountsRow{Total: 5, Draft: 1, Review: 2, RevisionNeeded: 3, Approved: 4, Published: 5, PackageReady: 6},
		BySubject:        []db.ListCbtQuestionSummaryBySubjectRow{{SubjectID: handlerTestUUID(151), SubjectName: "IPA", SubjectCode: "IPA", Total: 2}},
		ByCognitiveLevel: []db.ListCbtQuestionSummaryByCognitiveLevelRow{{CognitiveLevel: "C4", Total: 1}},
		Recent:           []db.ListCbtQuestionSummaryRecentRow{{ID: handlerTestUUID(152), Code: "Q-2", SubjectID: handlerTestUUID(151), SubjectName: "IPA", SubjectCode: "IPA", WorkflowStatus: "draft", Status: db.CbtQuestionStatusEnumDraft, AuthorUsername: "guru1"}},
	}}
	h := &CbtQuestion{svc: fake}
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/questions/summary", nil), jwt.MapClaims{
		"roles":       []any{"guru"},
		"permissions": []any{"bank_soal.read"},
		"usr":         "guru1",
		"uid":         "00000000-0000-0000-0000-000000000153",
	})
	rec := httptest.NewRecorder()
	h.Summary(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Summary() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.summaryActor.Username != "guru1" || !fake.summaryActor.UserID.Valid || !fake.summaryActor.HasRole("guru") || !fake.summaryActor.HasPermission("bank_soal.read") {
		t.Fatalf("Summary() actor = %+v, want claims-derived actor", fake.summaryActor)
	}
	decoded := decodeHandlerResponse(t, rec)
	data, ok := decoded.Data.(map[string]any)
	if !ok {
		t.Fatalf("Summary() data = %#v, want object", decoded.Data)
	}
	counts := data["counts"].(map[string]any)
	if counts["total"].(float64) != 5 || counts["revision_needed"].(float64) != 3 || counts["package_ready"].(float64) != 6 {
		t.Fatalf("Summary() counts = %#v, want serialized counts", counts)
	}
	if len(data["by_subject"].([]any)) != 1 || len(data["by_cognitive_level"].([]any)) != 1 || len(data["recent"].([]any)) != 1 {
		t.Fatalf("Summary() sections = %#v, want one row each", data)
	}

	fake.summaryErr = errors.New("boom")
	rec = httptest.NewRecorder()
	h.Summary(rec, req)
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("Summary(error) status/body = %d/%s, want 500 internal", rec.Code, rec.Body.String())
	}
}
