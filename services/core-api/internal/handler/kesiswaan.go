package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type kesiswaanService interface {
	Stats(ctx context.Context, teacherEmployeeID pgtype.UUID) (service.KesiswaanStats, error)
	ClassOptions(ctx context.Context) ([]db.ListKesiswaanClassOptionsRow, error)
	ListStudents(ctx context.Context, search, status, classID string, teacherEmployeeID pgtype.UUID) ([]db.ListKesiswaanStudentsRow, error)
	UpdateStudentProfile(ctx context.Context, arg db.UpdateKesiswaanStudentProfileParams) (db.Student, error)
	SaveStudentPhoto(ctx context.Context, input service.UploadStudentPhotoInput) (db.Student, error)
	StudentPhotoPath(filename string) (string, bool)
	CanReadStudentPhoto(ctx context.Context, filename string, teacherEmployeeID pgtype.UUID) (bool, error)
	ListCategories(ctx context.Context, search string) ([]db.ViolationCategory, error)
	CreateCategory(ctx context.Context, arg db.CreateViolationCategoryParams) (db.ViolationCategory, error)
	UpdateCategory(ctx context.Context, arg db.UpdateViolationCategoryParams) (db.ViolationCategory, error)
	DeleteCategory(ctx context.Context, id pgtype.UUID) error
	ListViolations(ctx context.Context, search, status, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentViolationsRow, error)
	CreateViolation(ctx context.Context, arg db.CreateStudentViolationParams) (db.StudentViolation, error)
	UpdateViolation(ctx context.Context, arg db.UpdateStudentViolationParams) (db.StudentViolation, error)
	DeleteViolation(ctx context.Context, id pgtype.UUID) error
	ListAchievements(ctx context.Context, search, level, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentAchievementsRow, error)
	CreateAchievement(ctx context.Context, arg db.CreateStudentAchievementParams) (db.StudentAchievement, error)
	UpdateAchievement(ctx context.Context, arg db.UpdateStudentAchievementParams) (db.StudentAchievement, error)
	DeleteAchievement(ctx context.Context, id pgtype.UUID) error
	ListExtracurriculars(ctx context.Context, search string, activeOnly bool) ([]db.ListExtracurricularsRow, error)
	CreateExtracurricular(ctx context.Context, arg db.CreateExtracurricularParams) (db.Extracurricular, error)
	UpdateExtracurricular(ctx context.Context, arg db.UpdateExtracurricularParams) (db.Extracurricular, error)
	DeleteExtracurricular(ctx context.Context, id pgtype.UUID) error
	ListExtracurricularMembers(ctx context.Context, search, status, extracurricularID, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListExtracurricularMembersRow, error)
	CreateExtracurricularMember(ctx context.Context, arg db.CreateExtracurricularMemberParams) (db.ExtracurricularMember, error)
	UpdateExtracurricularMember(ctx context.Context, arg db.UpdateExtracurricularMemberParams) (db.ExtracurricularMember, error)
	DeleteExtracurricularMember(ctx context.Context, id pgtype.UUID) error
	ListCounselingSessions(ctx context.Context, search, status, studentID string, canReadConfidential bool, teacherEmployeeID pgtype.UUID) ([]db.ListCounselingSessionsRow, error)
	CreateCounselingSession(ctx context.Context, arg db.CreateCounselingSessionParams) (db.CounselingSession, error)
	UpdateCounselingSession(ctx context.Context, arg db.UpdateCounselingSessionParams) (db.CounselingSession, error)
	DeleteCounselingSession(ctx context.Context, id pgtype.UUID) error
	ListStudentTransfers(ctx context.Context, search, transferType, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentTransfersRow, error)
	CreateStudentTransfer(ctx context.Context, arg db.CreateStudentTransferParams) (db.StudentTransfer, error)
}

type Kesiswaan struct {
	svc   kesiswaanService
	audit cbtAuthoringAuditWriter
}

func NewKesiswaan(svc *service.Kesiswaan, audit ...cbtAuthoringAuditWriter) *Kesiswaan {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &Kesiswaan{svc: svc, audit: writer}
}

type kesiswaanAccess struct {
	canRead           bool
	canManage         bool
	teacherEmployeeID pgtype.UUID
}

func getKesiswaanAccess(r *http.Request) kesiswaanAccess {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		hasAdmin := kesiswaanHasRole(claims, "admin")
		hasKesiswaan := kesiswaanHasRole(claims, "kesiswaan")
		hasGuru := kesiswaanHasRole(claims, "guru")
		access := kesiswaanAccess{
			canRead:   hasAdmin || hasKesiswaan || hasGuru,
			canManage: hasAdmin || hasKesiswaan,
		}
		if hasGuru && !access.canManage {
			if eid, _ := claims["eid"].(string); strings.TrimSpace(eid) != "" {
				_ = access.teacherEmployeeID.Scan(strings.TrimSpace(eid))
			}
			if !access.teacherEmployeeID.Valid {
				access.canRead = false
			}
		}
		return access
	}
	return kesiswaanAccess{}
}

func kesiswaanHasRole(claims map[string]any, expected string) bool {
	return mw.HasAnyRole(jwt.MapClaims(claims), expected)
}

func kesiswaanEmployeeID(r *http.Request) pgtype.UUID {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if kesiswaanHasRole(claims, "admin") || !kesiswaanHasRole(claims, "guru") {
			return pgtype.UUID{}
		}
		if eid, _ := claims["eid"].(string); strings.TrimSpace(eid) != "" {
			var id pgtype.UUID
			if err := id.Scan(strings.TrimSpace(eid)); err == nil {
				return id
			}
		}
	}
	return pgtype.UUID{}
}

func (h *Kesiswaan) Stats(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.Stats(r.Context(), access.teacherEmployeeID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Kesiswaan) ClassOptions(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ClassOptions(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Kesiswaan) ListStudents(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListStudents(r.Context(), q.Get("search"), q.Get("status"), q.Get("class_id"), access.teacherEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, data)
}

func (h *Kesiswaan) UpdateStudentProfile(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanStudentProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	tanggalLahir, err := service.ParseKesiswaanOptionalDate(body.TanggalLahir)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	row, err := h.svc.UpdateStudentProfile(r.Context(), db.UpdateKesiswaanStudentProfileParams{
		ID:           id,
		Nik:          body.Nik,
		TempatLahir:  body.TempatLahir,
		TanggalLahir: tanggalLahir,
		Alamat:       body.Alamat,
		Agama:        body.Agama,
		AnakKe:       service.KesiswaanOptionalInt4(body.AnakKe),
		Phone:        body.Phone,
		ParentName:   body.ParentName,
		ParentPhone:  body.ParentPhone,
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_STUDENT_PROFILE_UPDATE", "student", pgUUIDString(row.ID), map[string]any{
		"student_id":   pgUUIDString(row.ID),
		"nik":          row.Nik,
		"parent_name":  row.ParentName,
		"parent_phone": row.ParentPhone,
	})
	api.OK(w, row)
}

func (h *Kesiswaan) UploadStudentPhoto(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 3<<20)
	if err := r.ParseMultipartForm(3 << 20); err != nil {
		api.BadRequest(w, "multipart form tidak valid (maks 3 MB)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file wajib diisi")
		return
	}
	defer file.Close()
	validated, err := validateUploadedFile(header.Filename, file, 3<<20, true)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.SaveStudentPhoto(r.Context(), service.UploadStudentPhotoInput{
		StudentID:    id,
		OriginalName: header.Filename,
		MimeType:     validated.MimeType,
		FileSize:     int64(len(validated.Data)),
		File:         bytes.NewReader(validated.Data),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_STUDENT_PHOTO_UPLOAD", "student", pgUUIDString(row.ID), map[string]any{
		"photo_url": row.PhotoUrl,
	})
	api.OK(w, row)
}

func (h *Kesiswaan) StudentPhotoFile(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	filename := chi.URLParam(r, "filename")
	canRead, err := h.svc.CanReadStudentPhoto(r.Context(), filename, access.teacherEmployeeID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	if !canRead {
		api.NotFound(w)
		return
	}
	path, ok := h.svc.StudentPhotoPath(filename)
	if !ok {
		api.NotFound(w)
		return
	}
	secureFileResponseHeaders(w, mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))), filename)
	http.ServeFile(w, r, path)
}

func (h *Kesiswaan) ListCategories(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.ListCategories(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateCategory(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.CreateCategory(r.Context(), db.CreateViolationCategoryParams{
		Code:        body.Code,
		Name:        body.Name,
		Point:       body.Point,
		Severity:    body.Severity,
		Description: body.Description,
		IsActive:    boolDefault(body.IsActive, true),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Kesiswaan) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	row, err := h.svc.UpdateCategory(r.Context(), db.UpdateViolationCategoryParams{
		ID:          id,
		Code:        body.Code,
		Name:        body.Name,
		Point:       body.Point,
		Severity:    body.Severity,
		Description: body.Description,
		IsActive:    boolDefault(body.IsActive, true),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Kesiswaan) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	h.deleteManagedByID(w, r, h.svc.DeleteCategory)
}

func (h *Kesiswaan) ListViolations(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	rows, err := h.svc.ListViolations(r.Context(), q.Get("search"), q.Get("status"), q.Get("student_id"), access.teacherEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateViolation(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanViolationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	arg, ok := parseViolationRequest(w, body)
	if !ok {
		return
	}
	arg.ReportedByEmployeeID = kesiswaanEmployeeID(r)
	arg.RecordedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateViolation(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_VIOLATION_CREATE", "student_violation", pgUUIDString(row.ID), map[string]any{
		"student_id":  pgUUIDString(row.StudentID),
		"status":      row.Status,
		"recorded_by": currentUsername(r),
	})
	api.Created(w, row)
}

func (h *Kesiswaan) UpdateViolation(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanViolationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	createArg, ok := parseViolationRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateViolation(r.Context(), db.UpdateStudentViolationParams{
		ID:                   id,
		StudentID:            createArg.StudentID,
		CategoryID:           createArg.CategoryID,
		IncidentDate:         createArg.IncidentDate,
		Points:               createArg.Points,
		Description:          createArg.Description,
		ActionTaken:          createArg.ActionTaken,
		Status:               createArg.Status,
		ReportedByEmployeeID: kesiswaanEmployeeID(r),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_VIOLATION_UPDATE", "student_violation", pgUUIDString(row.ID), map[string]any{
		"student_id":  pgUUIDString(row.StudentID),
		"status":      row.Status,
		"recorded_by": currentUsername(r),
	})
	api.OK(w, row)
}

func (h *Kesiswaan) DeleteViolation(w http.ResponseWriter, r *http.Request) {
	id, ok := h.deleteManagedByID(w, r, h.svc.DeleteViolation)
	if !ok {
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_VIOLATION_DELETE", "student_violation", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
}

func (h *Kesiswaan) ListAchievements(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	rows, err := h.svc.ListAchievements(r.Context(), q.Get("search"), q.Get("level"), q.Get("student_id"), access.teacherEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateAchievement(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	arg, ok := parseAchievementRequest(w, body)
	if !ok {
		return
	}
	arg.RecordedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateAchievement(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Kesiswaan) UpdateAchievement(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	createArg, ok := parseAchievementRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateAchievement(r.Context(), db.UpdateStudentAchievementParams{
		ID:              id,
		StudentID:       createArg.StudentID,
		AchievementDate: createArg.AchievementDate,
		Title:           createArg.Title,
		Level:           createArg.Level,
		Category:        createArg.Category,
		Organizer:       createArg.Organizer,
		Description:     createArg.Description,
		DocumentUrl:     createArg.DocumentUrl,
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Kesiswaan) DeleteAchievement(w http.ResponseWriter, r *http.Request) {
	h.deleteManagedByID(w, r, h.svc.DeleteAchievement)
}

func (h *Kesiswaan) ListExtracurriculars(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	activeOnly := r.URL.Query().Get("active_only") == "true" || !access.canManage
	rows, err := h.svc.ListExtracurriculars(r.Context(), r.URL.Query().Get("search"), activeOnly)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateExtracurricular(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanExtracurricularRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	supervisorID, err := service.ParseKesiswaanOptionalUUID(body.SupervisorEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	row, err := h.svc.CreateExtracurricular(r.Context(), db.CreateExtracurricularParams{
		Code:                 body.Code,
		Name:                 body.Name,
		Category:             body.Category,
		Description:          body.Description,
		SupervisorEmployeeID: supervisorID,
		ScheduleText:         body.ScheduleText,
		IsActive:             boolDefault(body.IsActive, true),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Kesiswaan) UpdateExtracurricular(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanExtracurricularRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	supervisorID, err := service.ParseKesiswaanOptionalUUID(body.SupervisorEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	row, err := h.svc.UpdateExtracurricular(r.Context(), db.UpdateExtracurricularParams{
		ID:                   id,
		Code:                 body.Code,
		Name:                 body.Name,
		Category:             body.Category,
		Description:          body.Description,
		SupervisorEmployeeID: supervisorID,
		ScheduleText:         body.ScheduleText,
		IsActive:             boolDefault(body.IsActive, true),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Kesiswaan) DeleteExtracurricular(w http.ResponseWriter, r *http.Request) {
	h.deleteManagedByID(w, r, h.svc.DeleteExtracurricular)
}

func (h *Kesiswaan) ListExtracurricularMembers(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	rows, err := h.svc.ListExtracurricularMembers(
		r.Context(),
		q.Get("search"),
		q.Get("status"),
		q.Get("extracurricular_id"),
		q.Get("student_id"),
		access.teacherEmployeeID,
	)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateExtracurricularMember(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanExtracurricularMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	arg, ok := parseExtracurricularMemberRequest(w, body)
	if !ok {
		return
	}
	arg.RecordedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateExtracurricularMember(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Kesiswaan) UpdateExtracurricularMember(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanExtracurricularMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	createArg, ok := parseExtracurricularMemberRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateExtracurricularMember(r.Context(), db.UpdateExtracurricularMemberParams{
		ID:                id,
		ExtracurricularID: createArg.ExtracurricularID,
		StudentID:         createArg.StudentID,
		JoinedAt:          createArg.JoinedAt,
		Role:              createArg.Role,
		Status:            createArg.Status,
		Notes:             createArg.Notes,
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Kesiswaan) DeleteExtracurricularMember(w http.ResponseWriter, r *http.Request) {
	h.deleteManagedByID(w, r, h.svc.DeleteExtracurricularMember)
}

func (h *Kesiswaan) ListCounselingSessions(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	rows, err := h.svc.ListCounselingSessions(
		r.Context(),
		q.Get("search"),
		q.Get("status"),
		q.Get("student_id"),
		access.canManage,
		access.teacherEmployeeID,
	)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateCounselingSession(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanCounselingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	arg, ok := parseCounselingRequest(w, body)
	if !ok {
		return
	}
	arg.CounselorEmployeeID = kesiswaanEmployeeID(r)
	arg.RecordedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateCounselingSession(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_COUNSELING_CREATE", "counseling_session", pgUUIDString(row.ID), map[string]any{
		"student_id":      pgUUIDString(row.StudentID),
		"status":          row.Status,
		"is_confidential": row.IsConfidential,
	})
	api.Created(w, row)
}

func (h *Kesiswaan) UpdateCounselingSession(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body kesiswaanCounselingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	createArg, ok := parseCounselingRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateCounselingSession(r.Context(), db.UpdateCounselingSessionParams{
		ID:                  id,
		StudentID:           createArg.StudentID,
		SessionDate:         createArg.SessionDate,
		Topic:               createArg.Topic,
		Summary:             createArg.Summary,
		FollowUp:            createArg.FollowUp,
		Status:              createArg.Status,
		IsConfidential:      createArg.IsConfidential,
		CounselorEmployeeID: kesiswaanEmployeeID(r),
	})
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_COUNSELING_UPDATE", "counseling_session", pgUUIDString(row.ID), map[string]any{
		"student_id":      pgUUIDString(row.StudentID),
		"status":          row.Status,
		"is_confidential": row.IsConfidential,
	})
	api.OK(w, row)
}

func (h *Kesiswaan) DeleteCounselingSession(w http.ResponseWriter, r *http.Request) {
	id, ok := h.deleteManagedByID(w, r, h.svc.DeleteCounselingSession)
	if !ok {
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_COUNSELING_DELETE", "counseling_session", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
}

func (h *Kesiswaan) ListStudentTransfers(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canRead {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	rows, err := h.svc.ListStudentTransfers(r.Context(), q.Get("search"), q.Get("transfer_type"), q.Get("student_id"), access.teacherEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	api.OK(w, rows)
}

func (h *Kesiswaan) CreateStudentTransfer(w http.ResponseWriter, r *http.Request) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return
	}
	var body kesiswaanTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	arg, ok := parseStudentTransferRequest(w, body)
	if !ok {
		return
	}
	arg.RecordedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateStudentTransfer(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "KESISWAAN_STUDENT_TRANSFER_CREATE", "student_transfer", pgUUIDString(row.ID), map[string]any{
		"student_id":    pgUUIDString(row.StudentID),
		"transfer_type": row.TransferType,
		"destination":   row.DestinationSchool,
		"recorded_by":   currentUsername(r),
	})
	api.Created(w, row)
}

func (h *Kesiswaan) deleteManagedByID(w http.ResponseWriter, r *http.Request, deleteFn func(context.Context, pgtype.UUID) error) (pgtype.UUID, bool) {
	access := getKesiswaanAccess(r)
	if !access.canManage {
		api.Forbidden(w)
		return pgtype.UUID{}, false
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return pgtype.UUID{}, false
	}
	if err := deleteFn(r.Context(), id); err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return pgtype.UUID{}, false
	}
	api.NoContent(w)
	return id, true
}

type kesiswaanStudentProfileRequest struct {
	Nik          string `json:"nik"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	Alamat       string `json:"alamat"`
	Agama        string `json:"agama"`
	AnakKe       *int32 `json:"anak_ke"`
	Phone        string `json:"phone"`
	ParentName   string `json:"parent_name"`
	ParentPhone  string `json:"parent_phone"`
}

type kesiswaanCategoryRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Point       int32  `json:"point"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type kesiswaanViolationRequest struct {
	StudentID    string `json:"student_id"`
	CategoryID   string `json:"category_id"`
	IncidentDate string `json:"incident_date"`
	Points       int32  `json:"points"`
	Description  string `json:"description"`
	ActionTaken  string `json:"action_taken"`
	Status       string `json:"status"`
}

type kesiswaanAchievementRequest struct {
	StudentID       string `json:"student_id"`
	AchievementDate string `json:"achievement_date"`
	Title           string `json:"title"`
	Level           string `json:"level"`
	Category        string `json:"category"`
	Organizer       string `json:"organizer"`
	Description     string `json:"description"`
	DocumentUrl     string `json:"document_url"`
}

type kesiswaanExtracurricularRequest struct {
	Code                 string `json:"code"`
	Name                 string `json:"name"`
	Category             string `json:"category"`
	Description          string `json:"description"`
	SupervisorEmployeeID string `json:"supervisor_employee_id"`
	ScheduleText         string `json:"schedule_text"`
	IsActive             *bool  `json:"is_active"`
}

type kesiswaanExtracurricularMemberRequest struct {
	ExtracurricularID string `json:"extracurricular_id"`
	StudentID         string `json:"student_id"`
	JoinedAt          string `json:"joined_at"`
	Role              string `json:"role"`
	Status            string `json:"status"`
	Notes             string `json:"notes"`
}

type kesiswaanCounselingRequest struct {
	StudentID      string `json:"student_id"`
	SessionDate    string `json:"session_date"`
	Topic          string `json:"topic"`
	Summary        string `json:"summary"`
	FollowUp       string `json:"follow_up"`
	Status         string `json:"status"`
	IsConfidential bool   `json:"is_confidential"`
}

type kesiswaanTransferRequest struct {
	StudentID         string `json:"student_id"`
	TransferDate      string `json:"transfer_date"`
	TransferType      string `json:"transfer_type"`
	PreviousSchool    string `json:"previous_school"`
	DestinationSchool string `json:"destination_school"`
	Reason            string `json:"reason"`
	DocumentRef       string `json:"document_ref"`
	Notes             string `json:"notes"`
}

func parseViolationRequest(w http.ResponseWriter, body kesiswaanViolationRequest) (db.CreateStudentViolationParams, bool) {
	studentID, err := service.ParseKesiswaanOptionalUUID(body.StudentID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentViolationParams{}, false
	}
	categoryID, err := service.ParseKesiswaanOptionalUUID(body.CategoryID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentViolationParams{}, false
	}
	incidentDate, err := service.ParseKesiswaanDate(body.IncidentDate)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentViolationParams{}, false
	}
	return db.CreateStudentViolationParams{
		StudentID:    studentID,
		CategoryID:   categoryID,
		IncidentDate: incidentDate,
		Points:       body.Points,
		Description:  body.Description,
		ActionTaken:  body.ActionTaken,
		Status:       body.Status,
	}, true
}

func parseAchievementRequest(w http.ResponseWriter, body kesiswaanAchievementRequest) (db.CreateStudentAchievementParams, bool) {
	studentID, err := service.ParseKesiswaanOptionalUUID(body.StudentID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentAchievementParams{}, false
	}
	achievementDate, err := service.ParseKesiswaanDate(body.AchievementDate)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentAchievementParams{}, false
	}
	return db.CreateStudentAchievementParams{
		StudentID:       studentID,
		AchievementDate: achievementDate,
		Title:           body.Title,
		Level:           body.Level,
		Category:        body.Category,
		Organizer:       body.Organizer,
		Description:     body.Description,
		DocumentUrl:     body.DocumentUrl,
	}, true
}

func parseExtracurricularMemberRequest(w http.ResponseWriter, body kesiswaanExtracurricularMemberRequest) (db.CreateExtracurricularMemberParams, bool) {
	extracurricularID, err := service.ParseKesiswaanOptionalUUID(body.ExtracurricularID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateExtracurricularMemberParams{}, false
	}
	studentID, err := service.ParseKesiswaanOptionalUUID(body.StudentID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateExtracurricularMemberParams{}, false
	}
	joinedAt, err := service.ParseKesiswaanOptionalDate(body.JoinedAt)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateExtracurricularMemberParams{}, false
	}
	return db.CreateExtracurricularMemberParams{
		ExtracurricularID: extracurricularID,
		StudentID:         studentID,
		JoinedAt:          joinedAt,
		Role:              body.Role,
		Status:            body.Status,
		Notes:             body.Notes,
	}, true
}

func parseCounselingRequest(w http.ResponseWriter, body kesiswaanCounselingRequest) (db.CreateCounselingSessionParams, bool) {
	studentID, err := service.ParseKesiswaanOptionalUUID(body.StudentID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateCounselingSessionParams{}, false
	}
	sessionDate, err := service.ParseKesiswaanOptionalDate(body.SessionDate)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateCounselingSessionParams{}, false
	}
	return db.CreateCounselingSessionParams{
		StudentID:      studentID,
		SessionDate:    sessionDate,
		Topic:          body.Topic,
		Summary:        body.Summary,
		FollowUp:       body.FollowUp,
		Status:         body.Status,
		IsConfidential: body.IsConfidential,
	}, true
}

func parseStudentTransferRequest(w http.ResponseWriter, body kesiswaanTransferRequest) (db.CreateStudentTransferParams, bool) {
	studentID, err := service.ParseKesiswaanOptionalUUID(body.StudentID)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentTransferParams{}, false
	}
	transferDate, err := service.ParseKesiswaanOptionalDate(body.TransferDate)
	if err != nil {
		writeClientError(w, err, "Data kesiswaan tidak valid")
		return db.CreateStudentTransferParams{}, false
	}
	return db.CreateStudentTransferParams{
		StudentID:         studentID,
		TransferDate:      transferDate,
		TransferType:      body.TransferType,
		PreviousSchool:    body.PreviousSchool,
		DestinationSchool: body.DestinationSchool,
		Reason:            body.Reason,
		DocumentRef:       body.DocumentRef,
		Notes:             body.Notes,
	}, true
}
