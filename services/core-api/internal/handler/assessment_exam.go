package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type assessmentExamService interface {
	List(ctx context.Context, search, status string, limit, offset int32) ([]service.AssessmentExamView, error)
	Get(ctx context.Context, id pgtype.UUID) (service.AssessmentExamView, error)
	Create(ctx context.Context, input service.AssessmentExamInput) (service.AssessmentExamView, error)
	Update(ctx context.Context, id pgtype.UUID, input service.AssessmentExamInput) (service.AssessmentExamView, error)
	PrepareRooms(ctx context.Context, id pgtype.UUID) (service.AssessmentPrepareRoomsResult, error)
	IssueCards(ctx context.Context, id pgtype.UUID) (service.AssessmentIssueCardsResult, error)
	ListParticipantCards(ctx context.Context, id pgtype.UUID) ([]service.AssessmentParticipantCardView, error)
	IssueParticipantCards(ctx context.Context, id pgtype.UUID, input service.AssessmentCardIssueInput) (service.AssessmentParticipantCardIssueResult, error)
	AssignmentPreview(ctx context.Context, id pgtype.UUID, input service.AssessmentAssignmentRequest) (service.AssessmentAssignmentResult, error)
	AssignmentApply(ctx context.Context, id pgtype.UUID, input service.AssessmentAssignmentRequest) (service.AssessmentAssignmentResult, error)
	ListParticipantPlacements(ctx context.Context, id pgtype.UUID) ([]service.AssessmentParticipantPlacementView, error)
	MoveParticipantSeat(ctx context.Context, id pgtype.UUID, input service.AssessmentParticipantSeatInput) (service.AssessmentParticipantPlacementView, error)
}

type AssessmentExam struct {
	svc assessmentExamService
}

type assessmentExamRequest struct {
	Title      string     `json:"title"`
	SubjectID  string     `json:"subject_id"`
	GradeLevel *int16     `json:"grade_level"`
	Status     string     `json:"status"`
	StartsAt   *time.Time `json:"starts_at"`
	EndsAt     *time.Time `json:"ends_at"`
}

func NewAssessmentExam(svc assessmentExamService) *AssessmentExam {
	return &AssessmentExam{svc: svc}
}

func (h *AssessmentExam) List(w http.ResponseWriter, r *http.Request) {
	if !assessmentAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	items, err := h.svc.List(
		r.Context(),
		r.URL.Query().Get("search"),
		r.URL.Query().Get("status"),
		int32Param(r.URL.Query().Get("limit"), 20),
		int32Param(r.URL.Query().Get("offset"), 0),
	)
	if err != nil {
		writeDomainOrInternal(w, err, "Daftar asesmen tidak dapat dibuka")
		return
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *AssessmentExam) Get(w http.ResponseWriter, r *http.Request) {
	if !assessmentAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "Asesmen tidak ditemukan")
		return
	}
	api.OK(w, item)
}

func (h *AssessmentExam) Create(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	req, ok := decodeAssessmentExamRequest(w, r)
	if !ok {
		return
	}
	item, err := h.svc.Create(r.Context(), req.toServiceInput(actorUserID(r)))
	if err != nil {
		writeDomainOrInternal(w, err, "Draft asesmen tidak valid")
		return
	}
	api.Created(w, item)
}

func (h *AssessmentExam) Update(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	req, ok := decodeAssessmentExamRequest(w, r)
	if !ok {
		return
	}
	item, err := h.svc.Update(r.Context(), id, req.toServiceInput(actorUserID(r)))
	if err != nil {
		writeDomainOrInternal(w, err, "Perubahan asesmen tidak valid")
		return
	}
	api.OK(w, item)
}

func (h *AssessmentExam) PrepareRooms(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	result, err := h.svc.PrepareRooms(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "Ruang ujian belum dapat disiapkan")
		return
	}
	api.OK(w, result)
}

func (h *AssessmentExam) ListParticipantCards(w http.ResponseWriter, r *http.Request) {
	if !assessmentCardsAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListParticipantCards(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "Daftar kartu peserta belum dapat dibuka")
		return
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *AssessmentExam) IssueCards(w http.ResponseWriter, r *http.Request) {
	if !assessmentCardsAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	var req service.AssessmentCardIssueInput
	if r.Body != nil && r.ContentLength != 0 {
		if !decodeJSON(w, r, &req, 32<<10, disallowUnknownJSONFields) {
			return
		}
	}
	result, err := h.svc.IssueParticipantCards(r.Context(), id, req)
	if err != nil {
		writeDomainOrInternal(w, err, "Kartu peserta belum dapat diterbitkan")
		return
	}
	api.OK(w, result)
}

func (h *AssessmentExam) AssignmentPreview(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	var req service.AssessmentAssignmentRequest
	if !decodeJSON(w, r, &req, 32<<10, disallowUnknownJSONFields) {
		return
	}
	result, err := h.svc.AssignmentPreview(r.Context(), id, req)
	if err != nil {
		writeDomainOrInternal(w, err, "Preview ruang ujian belum dapat dibuat")
		return
	}
	api.OK(w, result)
}

func (h *AssessmentExam) AssignmentApply(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	var req service.AssessmentAssignmentRequest
	if !decodeJSON(w, r, &req, 32<<10, disallowUnknownJSONFields) {
		return
	}
	result, err := h.svc.AssignmentApply(r.Context(), id, req)
	if err != nil {
		writeDomainOrInternal(w, err, "Pembagian ruang ujian belum dapat disimpan")
		return
	}
	api.OK(w, result)
}

func (h *AssessmentExam) ListParticipantPlacements(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListParticipantPlacements(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "Daftar peserta ruang belum dapat dibuka")
		return
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *AssessmentExam) MoveParticipantSeat(w http.ResponseWriter, r *http.Request) {
	if !assessmentManageAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, ok := assessmentURLID(w, r)
	if !ok {
		return
	}
	var req service.AssessmentParticipantSeatInput
	if !decodeJSON(w, r, &req, 32<<10, disallowUnknownJSONFields) {
		return
	}
	item, err := h.svc.MoveParticipantSeat(r.Context(), id, req)
	if err != nil {
		writeDomainOrInternal(w, err, "Perubahan ruang/kursi peserta belum dapat disimpan")
		return
	}
	api.OK(w, item)
}

func decodeAssessmentExamRequest(w http.ResponseWriter, r *http.Request) (assessmentExamRequest, bool) {
	var req assessmentExamRequest
	return req, decodeJSON(w, r, &req, 32<<10, disallowUnknownJSONFields)
}

func (r assessmentExamRequest) toServiceInput(actor pgtype.UUID) service.AssessmentExamInput {
	return service.AssessmentExamInput{
		Title:      r.Title,
		SubjectID:  optionalAssessmentUUID(r.SubjectID),
		GradeLevel: optionalAssessmentInt2(r.GradeLevel),
		Status:     strings.TrimSpace(r.Status),
		StartsAt:   r.StartsAt,
		EndsAt:     r.EndsAt,
		ActorID:    actor,
	}
}

func optionalAssessmentUUID(raw string) pgtype.UUID {
	var id pgtype.UUID
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return id
	}
	_ = id.Scan(raw)
	return id
}

func optionalAssessmentInt2(value *int16) pgtype.Int2 {
	if value == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: *value, Valid: true}
}

func assessmentURLID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil || !id.Valid {
		api.BadRequest(w, "id asesmen tidak valid")
		return id, false
	}
	return id, true
}

func assessmentAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return hasAnyRoleOrPermission(claims, []string{"admin", "guru", "staf"}, []string{"asesmen.read", "asesmen.manage"})
}

func assessmentManageAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return hasAnyRoleOrPermission(claims, []string{"admin", "guru"}, []string{"asesmen.manage"})
}

func assessmentCardsAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return hasAnyRoleOrPermission(claims, []string{"admin"}, []string{"asesmen.cards_issue", "asesmen.manage"})
}

func hasAnyRoleOrPermission(claims map[string]any, roles []string, permissions []string) bool {
	for _, role := range roles {
		if role != "" && claimsHasValue(claims["roles"], role) {
			return true
		}
	}
	for _, permission := range permissions {
		if permission != "" && claimsHasValue(claims["permissions"], permission) {
			return true
		}
	}
	return false
}

func claimsHasValue(raw any, want string) bool {
	switch values := raw.(type) {
	case []string:
		for _, value := range values {
			if value == want {
				return true
			}
		}
	case []any:
		for _, value := range values {
			if text, ok := value.(string); ok && text == want {
				return true
			}
		}
	case string:
		for _, value := range strings.Split(values, ",") {
			if strings.TrimSpace(value) == want {
				return true
			}
		}
	}
	return false
}

func assessmentInt32Query(raw string, fallback int32) int32 {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return fallback
	}
	return int32(value)
}
