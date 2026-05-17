package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtQuestionAuthoringExtraTimelineAndWorkflowLowBranches(t *testing.T) {
	questionID := handlerTestUUID(180)
	tests := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{
			name: "timeline forbidden without cbt access",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).Timeline,
			req:  withRouteParam(httptest.NewRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String()+"/timeline", nil), "id", questionID.String()),
			want: http.StatusForbidden,
		},
		{
			name: "timeline service forbidden",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{timelineErr: domain.ErrForbidden}}).Timeline,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String()+"/timeline", ""), "id", questionID.String()),
			want: http.StatusForbidden,
		},
		{
			name: "timeline service internal",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{timelineErr: errors.New("db down")}}).Timeline,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String()+"/timeline", ""), "id", questionID.String()),
			want: http.StatusInternalServerError,
		},
		{
			name: "workflow events bad id",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{}}).WorkflowEvents,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/bad/workflow-events", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "workflow events not found",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{workflowEventErr: domain.ErrNotFound}}).WorkflowEvents,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String()+"/workflow-events", ""), "id", questionID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "workflow events internal",
			fn:   (&CbtQuestion{svc: &fakeCbtQuestionService{workflowEventErr: errors.New("db down")}}).WorkflowEvents,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String()+"/workflow-events", ""), "id", questionID.String()),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestCbtQuestionAuthoringExtraTimelineForwardsActor(t *testing.T) {
	questionID := handlerTestUUID(181)
	fake := &fakeCbtQuestionService{}
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/questions/"+questionID.String()+"/timeline", nil), jwt.MapClaims{
		"roles":       []string{"guru"},
		"permissions": []string{"bank_soal.read"},
		"usr":         "guru.bio",
		"uid":         handlerTestUUID(182).String(),
		"sub":         handlerTestUUID(182).String(),
	})
	req = withRouteParam(req, "id", questionID.String())
	rec := httptest.NewRecorder()
	(&CbtQuestion{svc: fake}).Timeline(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Timeline() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.timelineID != questionID {
		t.Fatalf("Timeline id = %v, want %v", fake.timelineID, questionID)
	}
	actor := cbtQuestionActorFromRequest(req)
	if actor.Username != "guru.bio" || !actor.UserID.Valid || len(actor.Roles) != 1 || actor.Roles[0] != "guru" || len(actor.Permissions) != 1 || actor.Permissions[0] != "bank_soal.read" {
		t.Fatalf("cbtQuestionActorFromRequest() = %+v, want username/user id/string roles/string permissions", actor)
	}
}

func TestCbtQuestionAuthoringExtraQuestionInputFromBodyBranches(t *testing.T) {
	subjectID := handlerTestUUID(183)
	eventID := handlerTestUUID(184)
	questionID := handlerTestUUID(185)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/questions", strings.NewReader(`{}`))
	req = withClaims(req, jwt.MapClaims{
		"roles": []any{"guru"},
		"sub":   "fallback-subject-user",
	})

	input, err := questionInputFromBody(req, cbtQuestionBody{
		SubjectID:     subjectID.String(),
		EventID:       eventID.String(),
		QuestionText:  "Legacy target",
		QuestionType:  "essay",
		AnswerKey:     "  kunci  ",
		Difficulty:    string(db.CbtQuestionDifficultyEnumMedium),
		Status:        string(db.CbtQuestionStatusEnumDraft),
		GradeLevel:    9,
		TargetLevel:   "",
		MediaAssetIDs: []string{"asset-a", "asset-b"},
	}, questionID)
	if err != nil {
		t.Fatalf("questionInputFromBody() error = %v", err)
	}
	if input.ID != questionID || input.SubjectID != subjectID || input.EventID != eventID {
		t.Fatalf("questionInputFromBody() ids = %v/%v/%v, want route/subject/event ids", input.ID, input.SubjectID, input.EventID)
	}
	if input.TargetLevel != "IX" || input.AnswerKey != "kunci" || input.AuthorUsername != "fallback-subject-user" || input.ReviewerUsername != "fallback-subject-user" || input.ApproverUsername != "fallback-subject-user" {
		t.Fatalf("questionInputFromBody() mapped target/answer/users = %q/%q/%q/%q/%q", input.TargetLevel, input.AnswerKey, input.AuthorUsername, input.ReviewerUsername, input.ApproverUsername)
	}
	if len(input.MediaAssetIDs) != 2 || input.MediaAssetIDs[1] != "asset-b" || input.Difficulty != db.CbtQuestionDifficultyEnumMedium {
		t.Fatalf("questionInputFromBody() metadata = %+v, want media and difficulty forwarded", input)
	}

	if _, err := questionInputFromBody(req, cbtQuestionBody{SubjectID: subjectID.String(), EventID: "bad"}, questionID); err == nil {
		t.Fatal("questionInputFromBody(invalid event_id) error = nil, want parse error")
	}
	if got := legacyQuestionTargetLevelFromInt(10); got != "" {
		t.Fatalf("legacyQuestionTargetLevelFromInt(10) = %q, want empty fallback", got)
	}
}
