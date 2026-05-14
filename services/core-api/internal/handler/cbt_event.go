package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtEvent struct {
	svc cbtEventService
}

type cbtEventService interface {
	List(ctx context.Context) ([]db.ListCbtExamEventsRow, error)
	ListForUser(ctx context.Context, userID pgtype.UUID) ([]db.ListCbtExamEventsRow, error)
	Get(ctx context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error)
	Overview(ctx context.Context, id pgtype.UUID) (service.CbtEventOverview, error)
	ListPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventPackagesRow, error)
	ListSessions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSessionsReadinessRow, error)
	QuestionCompleteness(ctx context.Context, eventID pgtype.UUID) (service.CbtQuestionCompleteness, error)
	GetQuestionRequirements(ctx context.Context, eventID pgtype.UUID) (db.GetCbtEventQuestionRequirementsRow, error)
	UpsertQuestionRequirements(ctx context.Context, eventID pgtype.UUID, in service.SaveCbtEventQuestionRequirementsInput) (db.UpsertCbtEventQuestionRequirementsRow, error)
	GetResults(ctx context.Context, id pgtype.UUID) ([]db.GetEventResultsRow, error)
	GetExamCards(ctx context.Context, id pgtype.UUID) ([]db.GetEventExamCardsRow, error)
	Create(ctx context.Context, in service.CreateCbtEventInput) (db.CbtExamEvent, error)
	Update(ctx context.Context, id pgtype.UUID, in service.CreateCbtEventInput) (db.CbtExamEvent, error)
	UpdateStatus(ctx context.Context, id pgtype.UUID, status string) (db.CbtExamEvent, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	ListMembers(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventMembersRow, error)
	CreateMember(ctx context.Context, eventID pgtype.UUID, in service.SaveCbtEventMemberInput) (db.CbtEventMember, error)
	UpdateMember(ctx context.Context, eventID pgtype.UUID, id pgtype.UUID, in service.SaveCbtEventMemberInput) (db.CbtEventMember, error)
	DeleteMember(ctx context.Context, eventID pgtype.UUID, id pgtype.UUID) error
	ListQuestionTargets(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectTargetsRow, error)
	UpsertQuestionTarget(ctx context.Context, eventID pgtype.UUID, in service.SaveCbtEventSubjectTargetInput) (db.CbtEventSubjectTarget, error)
	DeleteQuestionTarget(ctx context.Context, eventID pgtype.UUID, subjectID pgtype.UUID) error
	CanRead(ctx context.Context, eventID, userID pgtype.UUID) (bool, error)
}

func NewCbtEvent(svc *service.CbtEvent) *CbtEvent { return &CbtEvent{svc: svc} }

func (h *CbtEvent) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var (
		rows []db.ListCbtExamEventsRow
		err  error
	)
	if adminAccessAllowed(r) {
		rows, err = h.svc.List(r.Context())
	} else {
		claims, ok := api.ClaimsFromContext(r.Context())
		if !ok {
			api.Unauthorized(w)
			return
		}
		userID, userErr := authUserID(claims)
		if userErr != nil {
			api.Unauthorized(w)
			return
		}
		rows, err = h.svc.ListForUser(r.Context(), userID)
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) Get(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, id) {
		return
	}
	row, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) Overview(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, id) {
		return
	}
	row, err := h.svc.Overview(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) Readiness(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var (
		events []db.ListCbtExamEventsRow
		err    error
	)
	if adminAccessAllowed(r) {
		events, err = h.svc.List(r.Context())
	} else {
		claims, ok := api.ClaimsFromContext(r.Context())
		if !ok {
			api.Unauthorized(w)
			return
		}
		userID, userErr := authUserID(claims)
		if userErr != nil {
			api.Unauthorized(w)
			return
		}
		events, err = h.svc.ListForUser(r.Context(), userID)
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	if len(events) == 0 {
		api.OK(w, map[string]any{
			"events":         []db.ListCbtExamEventsRow{},
			"selected_event": nil,
			"overview":       nil,
		})
		return
	}

	selectedID, err := parseOptionalUUID(r.URL.Query().Get("event_id"))
	if err != nil {
		api.BadRequest(w, "Kegiatan asesmen tidak valid")
		return
	}
	if !selectedID.Valid {
		selectedID = defaultReadinessEventID(events)
	}
	allowed := false
	for _, event := range events {
		if event.ID == selectedID {
			allowed = true
			break
		}
	}
	if !allowed {
		api.Forbidden(w)
		return
	}
	overview, err := h.svc.Overview(r.Context(), selectedID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"events":         events,
		"selected_event": overview.Event,
		"overview":       overview,
	})
}

func defaultReadinessEventID(events []db.ListCbtExamEventsRow) pgtype.UUID {
	for _, event := range events {
		if event.Status == "draft" || event.Status == "active" {
			return event.ID
		}
	}
	return events[0].ID
}

func (h *CbtEvent) ListPackages(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, id) {
		return
	}
	rows, err := h.svc.ListPackages(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) ListSessions(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, id) {
		return
	}
	rows, err := h.svc.ListSessions(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) QuestionCompleteness(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, id) {
		return
	}
	rows, err := h.svc.QuestionCompleteness(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) GetQuestionRequirements(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, id) {
		return
	}
	row, err := h.svc.GetQuestionRequirements(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) UpsertQuestionRequirements(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var input service.SaveCbtEventQuestionRequirementsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpsertQuestionRequirements(r.Context(), id, input)
	if err != nil {
		writeClientError(w, err, "Pengaturan target kelengkapan soal tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) GetResults(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	rows, err := h.svc.GetResults(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) GetExamCards(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	rows, err := h.svc.GetExamCards(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Title          string   `json:"title"`
		ExamType       string   `json:"exam_type"`
		Scope          string   `json:"scope"`
		TargetLevels   []string `json:"target_levels"`
		AcademicYearID string   `json:"academic_year_id"`
		Status         string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if body.Title == "" {
		api.BadRequest(w, "title required")
		return
	}
	examType := db.CbtExamType(body.ExamType)
	if examType == "" {
		examType = db.CbtExamTypeLainnya
	}
	scope := body.Scope
	if scope == "" {
		scope = "class"
	}
	targetLevels := normalizeTargetLevels(body.TargetLevels)
	if err := validateTargetLevels(targetLevels); err != nil {
		writeClientError(w, err, "Data kegiatan asesmen tidak valid")
		return
	}

	var ayID pgtype.UUID
	if body.AcademicYearID != "" {
		var err error
		ayID, err = parseUUID(body.AcademicYearID)
		if err != nil {
			api.BadRequest(w, "Tahun ajaran tidak valid")
			return
		}
	}

	row, err := h.svc.Create(r.Context(), service.CreateCbtEventInput{
		Title:          body.Title,
		ExamType:       examType,
		Scope:          scope,
		TargetLevels:   targetLevels,
		AcademicYearID: ayID,
		Status:         body.Status,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Data kegiatan asesmen tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *CbtEvent) Update(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Title          string   `json:"title"`
		ExamType       string   `json:"exam_type"`
		Scope          string   `json:"scope"`
		TargetLevels   []string `json:"target_levels"`
		AcademicYearID string   `json:"academic_year_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	targetLevels := normalizeTargetLevels(body.TargetLevels)
	if err := validateTargetLevels(targetLevels); err != nil {
		writeClientError(w, err, "Data kegiatan asesmen tidak valid")
		return
	}
	var ayID pgtype.UUID
	if body.AcademicYearID != "" {
		ayID, err = parseUUID(body.AcademicYearID)
		if err != nil {
			api.BadRequest(w, "Tahun ajaran tidak valid")
			return
		}
	}
	row, err := h.svc.Update(r.Context(), id, service.CreateCbtEventInput{
		Title:          body.Title,
		ExamType:       db.CbtExamType(body.ExamType),
		Scope:          body.Scope,
		TargetLevels:   targetLevels,
		AcademicYearID: ayID,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpdateStatus(r.Context(), id, body.Status)
	if err != nil {
		writeDomainOrInternal(w, err, "Status kegiatan asesmen tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "Hapus kegiatan asesmen tidak valid")
		return
	}
	api.NoContent(w)
}

func (h *CbtEvent) ListMembers(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID kegiatan asesmen tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, eventID) {
		return
	}
	rows, err := h.svc.ListMembers(r.Context(), eventID)
	if err != nil {
		writeClientError(w, err, "Data anggota kegiatan asesmen tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) CreateMember(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID kegiatan asesmen tidak valid")
		return
	}
	input, err := cbtEventMemberInputFromRequest(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.CreateMember(r.Context(), eventID, input)
	if err != nil {
		writeClientError(w, err, "Data anggota kegiatan asesmen tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *CbtEvent) UpdateMember(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, memberID, ok := cbtEventMemberRouteIDs(w, r)
	if !ok {
		return
	}
	input, err := cbtEventMemberInputFromRequest(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.UpdateMember(r.Context(), eventID, memberID, input)
	if err != nil {
		writeClientError(w, err, "Data anggota kegiatan asesmen tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) DeleteMember(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, memberID, ok := cbtEventMemberRouteIDs(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteMember(r.Context(), eventID, memberID); err != nil {
		writeClientError(w, err, "Data anggota kegiatan asesmen tidak valid")
		return
	}
	api.NoContent(w)
}

func (h *CbtEvent) ListQuestionTargets(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID kegiatan asesmen tidak valid")
		return
	}
	if !h.requireEventReadAccess(w, r, eventID) {
		return
	}
	rows, err := h.svc.ListQuestionTargets(r.Context(), eventID)
	if err != nil {
		writeClientError(w, err, "Target soal kegiatan asesmen tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) UpsertQuestionTarget(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID kegiatan asesmen tidak valid")
		return
	}
	var body struct {
		SubjectID       string `json:"subject_id"`
		TargetQuestions int32  `json:"target_questions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		api.BadRequest(w, "Mata pelajaran tidak valid")
		return
	}
	row, err := h.svc.UpsertQuestionTarget(r.Context(), eventID, service.SaveCbtEventSubjectTargetInput{SubjectID: subjectID, TargetQuestions: body.TargetQuestions})
	if err != nil {
		writeClientError(w, err, "Target soal kegiatan asesmen tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) DeleteQuestionTarget(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID kegiatan asesmen tidak valid")
		return
	}
	subjectID, err := parseUUID(chi.URLParam(r, "subject_id"))
	if err != nil {
		api.BadRequest(w, "ID mata pelajaran tidak valid")
		return
	}
	if err := h.svc.DeleteQuestionTarget(r.Context(), eventID, subjectID); err != nil {
		writeClientError(w, err, "Target soal kegiatan asesmen tidak valid")
		return
	}
	api.NoContent(w)
}

func cbtEventMemberInputFromRequest(r *http.Request) (service.SaveCbtEventMemberInput, error) {
	var body struct {
		UserID     string `json:"user_id"`
		EmployeeID string `json:"employee_id"`
		SubjectID  string `json:"subject_id"`
		Role       string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return service.SaveCbtEventMemberInput{}, errors.New("Data yang dikirim tidak valid")
	}
	userID, err := parseUUID(body.UserID)
	if err != nil {
		return service.SaveCbtEventMemberInput{}, errors.New("Pengguna tidak valid")
	}
	var employeeID pgtype.UUID
	if strings.TrimSpace(body.EmployeeID) != "" {
		employeeID, err = parseUUID(body.EmployeeID)
		if err != nil {
			return service.SaveCbtEventMemberInput{}, errors.New("Pegawai tidak valid")
		}
	}
	var subjectID pgtype.UUID
	if strings.TrimSpace(body.SubjectID) != "" {
		subjectID, err = parseUUID(body.SubjectID)
		if err != nil {
			return service.SaveCbtEventMemberInput{}, errors.New("Mata pelajaran tidak valid")
		}
	}
	role := db.CbtEventMemberRole(strings.TrimSpace(body.Role))
	if !validCbtEventMemberRole(role) {
		return service.SaveCbtEventMemberInput{}, errors.New("Peran pengguna tidak valid")
	}
	return service.SaveCbtEventMemberInput{UserID: userID, EmployeeID: employeeID, SubjectID: subjectID, Role: role}, nil
}

func validCbtEventMemberRole(role db.CbtEventMemberRole) bool {
	switch role {
	case db.CbtEventMemberRolePanitia, db.CbtEventMemberRolePembuatSoal, db.CbtEventMemberRoleReviewer, db.CbtEventMemberRoleProktor, db.CbtEventMemberRolePengawas, db.CbtEventMemberRoleKorektor:
		return true
	default:
		return false
	}
}

func (h *CbtEvent) requireEventReadAccess(w http.ResponseWriter, r *http.Request, eventID pgtype.UUID) bool {
	if adminAccessAllowed(r) {
		return true
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return false
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Forbidden(w)
		return false
	}
	allowed, err := h.svc.CanRead(r.Context(), eventID, userID)
	if err != nil {
		api.Internal(w, err)
		return false
	}
	if !allowed {
		api.Forbidden(w)
		return false
	}
	return true
}

func cbtEventMemberRouteIDs(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	eventID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID kegiatan asesmen tidak valid")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	memberID, err := parseUUID(chi.URLParam(r, "member_id"))
	if err != nil {
		api.BadRequest(w, "ID anggota kegiatan tidak valid")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return eventID, memberID, true
}

func normalizeTargetLevels(levels []string) []string {
	if len(levels) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(levels))
	seen := map[string]struct{}{}
	for _, level := range levels {
		normalized := strings.ToUpper(strings.TrimSpace(level))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	slices.Sort(out)
	return out
}

func validateTargetLevels(levels []string) error {
	allowed := map[string]struct{}{
		"VII":  {},
		"VIII": {},
		"IX":   {},
	}
	for _, level := range levels {
		if _, ok := allowed[level]; !ok {
			return errors.New("target_levels hanya boleh berisi VII, VIII, atau IX")
		}
	}
	return nil
}
