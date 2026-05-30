package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeAssessmentExamHandlerService struct {
	listCalled           bool
	createInput          service.AssessmentExamInput
	prepareCalled        bool
	previewCalled        bool
	applyCalled          bool
	listCardsCalled      bool
	issueCardsCalled     bool
	listPlacementsCalled bool
	moveSeatCalled       bool
	assignmentReq        service.AssessmentAssignmentRequest
	seatInput            service.AssessmentParticipantSeatInput
	cardIssueInput       service.AssessmentCardIssueInput
	listItems            []service.AssessmentExamView
	createItem           service.AssessmentExamView
	prepareResult        service.AssessmentPrepareRoomsResult
	previewResult        service.AssessmentAssignmentResult
	applyResult          service.AssessmentAssignmentResult
	placements           []service.AssessmentParticipantPlacementView
	movedPlacement       service.AssessmentParticipantPlacementView
	participantCards     []service.AssessmentParticipantCardView
	issueCardsResult     service.AssessmentParticipantCardIssueResult
}

func (f *fakeAssessmentExamHandlerService) List(context.Context, string, string, int32, int32) ([]service.AssessmentExamView, error) {
	f.listCalled = true
	return f.listItems, nil
}

func (f *fakeAssessmentExamHandlerService) Get(context.Context, pgtype.UUID) (service.AssessmentExamView, error) {
	return service.AssessmentExamView{}, nil
}

func (f *fakeAssessmentExamHandlerService) Create(_ context.Context, input service.AssessmentExamInput) (service.AssessmentExamView, error) {
	f.createInput = input
	return f.createItem, nil
}

func (f *fakeAssessmentExamHandlerService) Update(context.Context, pgtype.UUID, service.AssessmentExamInput) (service.AssessmentExamView, error) {
	return service.AssessmentExamView{}, nil
}

func (f *fakeAssessmentExamHandlerService) PrepareRooms(context.Context, pgtype.UUID) (service.AssessmentPrepareRoomsResult, error) {
	f.prepareCalled = true
	return f.prepareResult, nil
}

func (f *fakeAssessmentExamHandlerService) IssueCards(context.Context, pgtype.UUID) (service.AssessmentIssueCardsResult, error) {
	return service.AssessmentIssueCardsResult{}, nil
}

func (f *fakeAssessmentExamHandlerService) ListParticipantCards(context.Context, pgtype.UUID) ([]service.AssessmentParticipantCardView, error) {
	f.listCardsCalled = true
	return f.participantCards, nil
}

func (f *fakeAssessmentExamHandlerService) IssueParticipantCards(_ context.Context, _ pgtype.UUID, input service.AssessmentCardIssueInput) (service.AssessmentParticipantCardIssueResult, error) {
	f.issueCardsCalled = true
	f.cardIssueInput = input
	return f.issueCardsResult, nil
}

func (f *fakeAssessmentExamHandlerService) AssignmentPreview(_ context.Context, _ pgtype.UUID, req service.AssessmentAssignmentRequest) (service.AssessmentAssignmentResult, error) {
	f.previewCalled = true
	f.assignmentReq = req
	return f.previewResult, nil
}

func (f *fakeAssessmentExamHandlerService) ListParticipantPlacements(context.Context, pgtype.UUID) ([]service.AssessmentParticipantPlacementView, error) {
	f.listPlacementsCalled = true
	return f.placements, nil
}

func (f *fakeAssessmentExamHandlerService) MoveParticipantSeat(_ context.Context, _ pgtype.UUID, input service.AssessmentParticipantSeatInput) (service.AssessmentParticipantPlacementView, error) {
	f.moveSeatCalled = true
	f.seatInput = input
	return f.movedPlacement, nil
}

func (f *fakeAssessmentExamHandlerService) AssignmentApply(_ context.Context, _ pgtype.UUID, req service.AssessmentAssignmentRequest) (service.AssessmentAssignmentResult, error) {
	f.applyCalled = true
	f.assignmentReq = req
	return f.applyResult, nil
}

func (f *fakeAssessmentExamHandlerService) ListPackageMaps(context.Context, pgtype.UUID) ([]service.AssessmentPackageMapView, error) {
	return nil, nil
}

func (f *fakeAssessmentExamHandlerService) SavePackageMaps(context.Context, pgtype.UUID, service.AssessmentPackageMapRequest) (service.AssessmentPackageMapSaveResult, error) {
	return service.AssessmentPackageMapSaveResult{Count: 1}, nil
}

func (f *fakeAssessmentExamHandlerService) DeletePackageMap(context.Context, pgtype.UUID, pgtype.UUID) error {
	return nil
}

func (f *fakeAssessmentExamHandlerService) ListPackageOptions(context.Context, pgtype.UUID) ([]service.AssessmentPackageOptionView, error) {
	return nil, nil
}

func TestAssessmentExamListRequiresAssessmentRead(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/asesmen/exams", nil), jwt.MapClaims{"roles": []any{"siswa"}})

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("List status = %d, want forbidden", rec.Code)
	}
	if svc.listCalled {
		t.Fatal("List called service despite forbidden role")
	}
}

func TestAssessmentExamCreateDecodesDraftInput(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{createItem: service.AssessmentExamView{ID: "exam-1", Title: "PAT", Status: "draft"}}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := adminRequest(http.MethodPost, "/api/asesmen/exams", `{"title":" PAT ","grade_level":8}`)

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want created; body=%s", rec.Code, rec.Body.String())
	}
	if svc.createInput.Title != " PAT " || !svc.createInput.GradeLevel.Valid || svc.createInput.GradeLevel.Int16 != 8 {
		t.Fatalf("create input = %+v, want decoded title and grade", svc.createInput)
	}
}

func TestAssessmentExamPrepareRoomsUsesRouteID(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{prepareResult: service.AssessmentPrepareRoomsResult{ExamID: "11111111-1111-1111-1111-111111111111", RoomCount: 1}}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPost, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/prepare-rooms", ""), "id", "11111111-1111-1111-1111-111111111111")

	h.PrepareRooms(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PrepareRooms status = %d, want ok; body=%s", rec.Code, rec.Body.String())
	}
	if !svc.prepareCalled || !strings.Contains(rec.Body.String(), "room_count") {
		t.Fatalf("PrepareRooms did not call service or serialize result: called=%v body=%s", svc.prepareCalled, rec.Body.String())
	}
}

func TestAssessmentExamAssignmentPreviewDecodesPolicy(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{
		previewResult: service.AssessmentAssignmentResult{
			ExamID: "11111111-1111-1111-1111-111111111111",
			Rooms:  []service.AssessmentAssignmentRoom{{Code: "R01", Capacity: 8}},
		},
	}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPost, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/assignment-preview", `{"room_count":8,"capacity_per_room":30,"mix_policy":"mixed"}`), "id", "11111111-1111-1111-1111-111111111111")

	h.AssignmentPreview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("AssignmentPreview status = %d, want ok; body=%s", rec.Code, rec.Body.String())
	}
	if !svc.previewCalled || svc.assignmentReq.RoomCount != 8 || svc.assignmentReq.CapacityPerRoom != 30 || svc.assignmentReq.MixPolicy != "mixed" {
		t.Fatalf("preview call=%v req=%+v, want decoded assignment request", svc.previewCalled, svc.assignmentReq)
	}
}

func TestAssessmentExamAssignmentApplyRequiresManage(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/assignment-apply", strings.NewReader(`{"room_count":1,"capacity_per_room":30}`)), jwt.MapClaims{"roles": []any{"staf"}}), "id", "11111111-1111-1111-1111-111111111111")

	h.AssignmentApply(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("AssignmentApply status = %d, want forbidden", rec.Code)
	}
	if svc.applyCalled {
		t.Fatal("AssignmentApply called service despite forbidden role")
	}
}

func TestAssessmentExamListParticipantPlacementsRequiresManage(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/participants", nil), jwt.MapClaims{"roles": []any{"staf"}}), "id", "11111111-1111-1111-1111-111111111111")

	h.ListParticipantPlacements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("ListParticipantPlacements status = %d, want forbidden", rec.Code)
	}
	if svc.listPlacementsCalled {
		t.Fatal("ListParticipantPlacements called service despite missing manage permission")
	}
}

func TestAssessmentExamMoveParticipantSeatDecodesManualPlacement(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{movedPlacement: service.AssessmentParticipantPlacementView{ParticipantID: "aaaaaaaa-0000-0000-0000-000000000001", RoomCode: "R02", SeatNo: 7}}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	body := `{"participant_id":"aaaaaaaa-0000-0000-0000-000000000001","room_id":"33333333-3333-3333-3333-333333333333","seat_no":7}`
	req := withRouteParam(adminRequest(http.MethodPatch, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/participants/seat", body), "id", "11111111-1111-1111-1111-111111111111")

	h.MoveParticipantSeat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("MoveParticipantSeat status = %d, want ok; body=%s", rec.Code, rec.Body.String())
	}
	if !svc.moveSeatCalled || svc.seatInput.ParticipantID != "aaaaaaaa-0000-0000-0000-000000000001" || svc.seatInput.RoomID != "33333333-3333-3333-3333-333333333333" || svc.seatInput.SeatNo != 7 {
		t.Fatalf("move call=%v input=%+v, want decoded manual placement", svc.moveSeatCalled, svc.seatInput)
	}
}

func TestAssessmentExamListParticipantCardsRequiresCardsPermission(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/participant-cards", nil), jwt.MapClaims{"roles": []any{"guru"}}), "id", "11111111-1111-1111-1111-111111111111")

	h.ListParticipantCards(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("ListParticipantCards status = %d, want forbidden", rec.Code)
	}
	if svc.listCardsCalled {
		t.Fatal("ListParticipantCards called service without cards permission")
	}
}

func TestAssessmentExamIssueCardsDecodesRegenerate(t *testing.T) {
	svc := &fakeAssessmentExamHandlerService{issueCardsResult: service.AssessmentParticipantCardIssueResult{Count: 1, Cards: []service.AssessmentParticipantCardView{{ParticipantID: "p1", Token: "t", PIN: "1234"}}}}
	h := NewAssessmentExam(svc)
	rec := httptest.NewRecorder()
	req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/asesmen/exams/11111111-1111-1111-1111-111111111111/issue-cards", strings.NewReader(`{"regenerate":true,"expires_hours":72}`)), jwt.MapClaims{"roles": []any{"admin"}}), "id", "11111111-1111-1111-1111-111111111111")

	h.IssueCards(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("IssueCards status = %d, want ok; body=%s", rec.Code, rec.Body.String())
	}
	if !svc.issueCardsCalled || !svc.cardIssueInput.Regenerate || svc.cardIssueInput.ExpiresHours != 72 {
		t.Fatalf("issue cards called=%v input=%+v, want decoded regenerate/expires", svc.issueCardsCalled, svc.cardIssueInput)
	}
	if !strings.Contains(rec.Body.String(), "pin") || !strings.Contains(rec.Body.String(), "token") {
		t.Fatalf("IssueCards body=%s, want issued raw token/PIN in POST response", rec.Body.String())
	}
}
