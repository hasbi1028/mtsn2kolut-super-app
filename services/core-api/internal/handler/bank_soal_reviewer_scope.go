package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type BankSoalReviewerScope struct {
	svc bankSoalReviewerScopeService
}

type bankSoalReviewerScopeService interface {
	List(ctx context.Context) ([]db.ListBankSoalReviewerScopesRow, error)
	Upsert(ctx context.Context, input service.BankSoalReviewerScopeInput) (db.UpsertBankSoalReviewerScopeRow, error)
	Delete(ctx context.Context, id pgtype.UUID) error
}

type bankSoalReviewerScopeBody struct {
	UserID     string `json:"user_id"`
	SubjectID  string `json:"subject_id"`
	GradeLevel *int16 `json:"grade_level"`
	CanReview  *bool  `json:"can_review"`
	CanApprove *bool  `json:"can_approve"`
}

func NewBankSoalReviewerScope(svc *service.BankSoalReviewerScope) *BankSoalReviewerScope {
	return &BankSoalReviewerScope{svc: svc}
}

func (h *BankSoalReviewerScope) List(w http.ResponseWriter, r *http.Request) {
	if !bankSoalReviewerScopeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeBankSoalReviewerScopeListRow(row))
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *BankSoalReviewerScope) Upsert(w http.ResponseWriter, r *http.Request) {
	if !bankSoalReviewerScopeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body bankSoalReviewerScopeBody
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	input, err := reviewerScopeInputFromBody(body)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	input.AssignedBy = actorID
	row, err := h.svc.Upsert(r.Context(), input)
	if err != nil {
		writeClientError(w, err, "Scope reviewer Bank Soal tidak valid")
		return
	}
	api.OK(w, serializeBankSoalReviewerScopeRow(row))
}

func (h *BankSoalReviewerScope) Delete(w http.ResponseWriter, r *http.Request) {
	if !bankSoalReviewerScopeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "Scope reviewer tidak valid")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeClientError(w, err, "Scope reviewer Bank Soal tidak dapat dihapus")
		return
	}
	api.NoContent(w)
}

func bankSoalReviewerScopeAccessAllowed(r *http.Request) bool {
	return hasAnyRole(r, "admin") || hasAnyPermission(r, "bank_soal.assign_reviewer", "bank_soal.settings")
}

func reviewerScopeInputFromBody(body bankSoalReviewerScopeBody) (service.BankSoalReviewerScopeInput, error) {
	userID, err := parseUUID(body.UserID)
	if err != nil {
		return service.BankSoalReviewerScopeInput{}, err
	}
	subjectID, err := parseOptionalUUID(body.SubjectID)
	if err != nil {
		return service.BankSoalReviewerScopeInput{}, err
	}
	gradeLevel := pgtype.Int2{}
	if body.GradeLevel != nil {
		gradeLevel = pgtype.Int2{Int16: *body.GradeLevel, Valid: true}
	}
	canReview := true
	if body.CanReview != nil {
		canReview = *body.CanReview
	}
	canApprove := false
	if body.CanApprove != nil {
		canApprove = *body.CanApprove
	}
	return service.BankSoalReviewerScopeInput{
		UserID:     userID,
		SubjectID:  subjectID,
		GradeLevel: gradeLevel,
		CanReview:  canReview,
		CanApprove: canApprove,
	}, nil
}

func serializeBankSoalReviewerScopeListRow(row db.ListBankSoalReviewerScopesRow) map[string]any {
	payload := map[string]any{
		"id":                       pgUUIDString(row.ID),
		"user_id":                  pgUUIDString(row.UserID),
		"username":                 row.Username,
		"user_display_name":        row.UserDisplayName,
		"user_roles":               decodeJSONBytes(row.UserRoles),
		"subject_id":               pgUUIDString(row.SubjectID),
		"subject_name":             row.SubjectName,
		"subject_code":             row.SubjectCode,
		"grade_level":              pgInt2Value(row.GradeLevel),
		"can_review":               row.CanReview,
		"can_approve":              row.CanApprove,
		"assigned_by":              pgUUIDString(row.AssignedBy),
		"assigned_by_display_name": row.AssignedByDisplayName,
		"created_at":               row.CreatedAt,
		"updated_at":               row.UpdatedAt,
	}
	return payload
}

func serializeBankSoalReviewerScopeRow(row db.UpsertBankSoalReviewerScopeRow) map[string]any {
	return map[string]any{
		"id":          pgUUIDString(row.ID),
		"user_id":     pgUUIDString(row.UserID),
		"subject_id":  pgUUIDString(row.SubjectID),
		"grade_level": pgInt2Value(row.GradeLevel),
		"can_review":  row.CanReview,
		"can_approve": row.CanApprove,
		"assigned_by": pgUUIDString(row.AssignedBy),
		"created_at":  row.CreatedAt,
		"updated_at":  row.UpdatedAt,
	}
}

func pgInt2Value(value pgtype.Int2) any {
	if !value.Valid {
		return nil
	}
	return value.Int16
}
