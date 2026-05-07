package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type portalService interface {
	StudentOverview(ctx context.Context, studentID pgtype.UUID) (db.GetStudentByIDRow, []db.ListStudentParentsRow, []db.ListStudentExamSessionsRow, error)
	StudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error)
	TeacherTimetable(ctx context.Context, employeeID pgtype.UUID) ([]db.ListTeacherTimetableRow, error)
	ParentChildrenTimetable(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenTimetableRow, error)
	ParentOverview(ctx context.Context, parentID pgtype.UUID) (db.Parent, []db.ListParentChildrenRow, error)
}

type Portal struct {
	svc portalService
}

func NewPortal(svc *service.Portal) *Portal { return &Portal{svc: svc} }

func (h *Portal) StudentMe(w http.ResponseWriter, r *http.Request) {
	if !portalHasAnyRoleOrPermission(r, []string{"siswa"}, []string{"student_portal.read"}) {
		api.Forbidden(w)
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	rawID, _ := claims["sid"].(string)
	if rawID == "" {
		api.Forbidden(w)
		return
	}
	var studentID pgtype.UUID
	if err := studentID.Scan(rawID); err != nil {
		api.Forbidden(w)
		return
	}
	student, parents, sessions, err := h.svc.StudentOverview(r.Context(), studentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	timetable, err := h.svc.StudentTimetable(r.Context(), studentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"student":   student,
		"parents":   parents,
		"sessions":  studentPortalResults(sessions),
		"timetable": timetable,
	})
}

func (h *Portal) TeacherTimetable(w http.ResponseWriter, r *http.Request) {
	if !portalHasAnyRole(r, "guru") {
		api.Forbidden(w)
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	rawID, _ := claims["eid"].(string)
	if rawID == "" {
		api.Forbidden(w)
		return
	}
	var employeeID pgtype.UUID
	if err := employeeID.Scan(rawID); err != nil {
		api.Forbidden(w)
		return
	}
	timetable, err := h.svc.TeacherTimetable(r.Context(), employeeID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"timetable": timetable,
	})
}

func (h *Portal) ParentMe(w http.ResponseWriter, r *http.Request) {
	if !portalHasAnyRoleOrPermission(r, []string{"ortu"}, []string{"parent_portal.read"}) {
		api.Forbidden(w)
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	rawID, _ := claims["pid"].(string)
	if rawID == "" {
		api.Forbidden(w)
		return
	}
	var parentID pgtype.UUID
	if err := parentID.Scan(rawID); err != nil {
		api.Forbidden(w)
		return
	}
	parent, children, err := h.svc.ParentOverview(r.Context(), parentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	timetable, err := h.svc.ParentChildrenTimetable(r.Context(), parentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"parent":    parent,
		"children":  children,
		"timetable": timetable,
	})
}

func portalHasAnyRole(r *http.Request, roles ...string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, roles...)
}

func portalHasAnyRoleOrPermission(r *http.Request, roles []string, permissions []string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, roles...) || mw.HasAnyPermission(claims, permissions...)
}
