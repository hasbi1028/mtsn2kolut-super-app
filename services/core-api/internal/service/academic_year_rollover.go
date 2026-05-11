package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const academicYearActivationChallenge = "AKTIFKAN"
const academicYearRolloverApplyPrefix = "TERAPKAN KENAIKAN KELAS"

var academicYearNamePattern = regexp.MustCompile(`^\d{4}/\d{4}$`)

type YearRolloverPreviewInput struct {
	SourceAcademicYearID pgtype.UUID
	TargetAcademicYearID pgtype.UUID
}

type YearRolloverApplyInput struct {
	SourceAcademicYearID pgtype.UUID
	TargetAcademicYearID pgtype.UUID
	Confirmation         string
	SafetyToken          string
}

type YearRolloverPreview struct {
	SourceAcademicYearID   string                       `json:"source_academic_year_id"`
	SourceAcademicYearName string                       `json:"source_academic_year_name"`
	TargetAcademicYearID   string                       `json:"target_academic_year_id"`
	TargetAcademicYearName string                       `json:"target_academic_year_name"`
	ApplyChallenge         string                       `json:"apply_challenge"`
	Counts                 YearRolloverPreviewCounts    `json:"counts"`
	ClassesToCreate        []YearRolloverClassToCreate  `json:"classes_to_create"`
	StudentsToPromote      []YearRolloverStudentMove    `json:"students_to_promote"`
	StudentsWithoutNext    []YearRolloverStudentWarning `json:"students_without_next_class"`
	Warnings               []string                     `json:"warnings"`
}

type YearRolloverApplyResult struct {
	SourceAcademicYearID   string                       `json:"source_academic_year_id"`
	SourceAcademicYearName string                       `json:"source_academic_year_name"`
	TargetAcademicYearID   string                       `json:"target_academic_year_id"`
	TargetAcademicYearName string                       `json:"target_academic_year_name"`
	Counts                 YearRolloverApplyCounts      `json:"counts"`
	Classes                []YearRolloverClassApply     `json:"classes"`
	StudentsPromoted       []YearRolloverStudentMove    `json:"students_promoted"`
	StudentsSkipped        []YearRolloverStudentWarning `json:"students_skipped"`
	Warnings               []string                     `json:"warnings"`
}

type YearRolloverPreviewCounts struct {
	ClassesToCreate          int `json:"classes_to_create"`
	StudentsToPromote        int `json:"students_to_promote"`
	StudentsWithoutNextClass int `json:"students_without_next_class"`
	HomeroomAssignmentsCopy  int `json:"homeroom_assignments_to_copy"`
	SubjectAssignmentsCopy   int `json:"subject_assignments_to_copy"`
	TimetableSlotsCopy       int `json:"timetable_slots_to_copy"`
}

type YearRolloverApplyCounts struct {
	ClassesCreated       int `json:"classes_created"`
	ClassesReused        int `json:"classes_reused"`
	StudentsPromoted     int `json:"students_promoted"`
	StudentsSkipped      int `json:"students_skipped"`
	HomeroomsCopied      int `json:"homerooms_copied"`
	AssignmentsCopied    int `json:"assignments_copied"`
	TimetableSlotsCopied int `json:"timetable_slots_copied"`
}

type YearRolloverClassToCreate struct {
	SourceClassID string `json:"source_class_id"`
	SourceCode    string `json:"source_code"`
	SourceName    string `json:"source_name"`
	SourceLevel   string `json:"source_level"`
	TargetCode    string `json:"target_code"`
	TargetName    string `json:"target_name"`
	TargetLevel   string `json:"target_level"`
}

type YearRolloverClassApply struct {
	SourceClassID string `json:"source_class_id"`
	SourceCode    string `json:"source_code"`
	SourceName    string `json:"source_name"`
	SourceLevel   string `json:"source_level"`
	TargetClassID string `json:"target_class_id"`
	TargetCode    string `json:"target_code"`
	TargetName    string `json:"target_name"`
	TargetLevel   string `json:"target_level"`
	Action        string `json:"action"`
}

type YearRolloverStudentMove struct {
	StudentID     string `json:"student_id"`
	NIS           string `json:"nis"`
	NISN          string `json:"nisn"`
	Nama          string `json:"nama"`
	FromClassID   string `json:"from_class_id"`
	FromClassCode string `json:"from_class_code"`
	ToClassID     string `json:"to_class_id"`
	ToClassCode   string `json:"to_class_code"`
	ToClassName   string `json:"to_class_name"`
}

type YearRolloverStudentWarning struct {
	StudentID     string `json:"student_id"`
	NIS           string `json:"nis"`
	NISN          string `json:"nisn"`
	Nama          string `json:"nama"`
	FromClassID   string `json:"from_class_id"`
	FromClassCode string `json:"from_class_code"`
	Reason        string `json:"reason"`
}

type AcademicImportDryRunInput struct {
	Kind string `json:"kind"`
	CSV  string `json:"csv"`
}

type AcademicImportDryRunResult struct {
	Kind        string                   `json:"kind"`
	TotalRows   int                      `json:"total_rows"`
	AddCount    int                      `json:"add_count"`
	UpdateCount int                      `json:"update_count"`
	SkipCount   int                      `json:"skip_count"`
	ErrorCount  int                      `json:"error_count"`
	Rows        []AcademicImportRowPlan  `json:"rows"`
	RowErrors   []AcademicImportRowError `json:"row_errors"`
}

type AcademicImportRowPlan struct {
	Row     int    `json:"row"`
	Action  string `json:"action"`
	Summary string `json:"summary"`
}

type AcademicImportRowError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func normalizeAcademicYearParams(p db.CreateAcademicYearParams) (db.CreateAcademicYearParams, error) {
	p.Name = strings.TrimSpace(p.Name)
	if !academicYearNamePattern.MatchString(p.Name) {
		return db.CreateAcademicYearParams{}, fmt.Errorf("%w: format tahun ajaran harus YYYY/YYYY", domain.ErrBadRequest)
	}
	parts := strings.Split(p.Name, "/")
	startYear, _ := strconv.Atoi(parts[0])
	endYear, _ := strconv.Atoi(parts[1])
	if endYear != startYear+1 {
		return db.CreateAcademicYearParams{}, fmt.Errorf("%w: rentang tahun ajaran harus berurutan", domain.ErrBadRequest)
	}
	if !p.StartDate.Valid || !p.EndDate.Valid {
		return db.CreateAcademicYearParams{}, fmt.Errorf("%w: tanggal tahun ajaran wajib diisi", domain.ErrBadRequest)
	}
	if !p.StartDate.Time.Before(p.EndDate.Time) {
		return db.CreateAcademicYearParams{}, fmt.Errorf("%w: tanggal selesai harus setelah tanggal mulai", domain.ErrBadRequest)
	}
	if p.IsActive {
		return db.CreateAcademicYearParams{}, fmt.Errorf("%w: buat tahun ajaran baru sebagai nonaktif, lalu aktifkan secara eksplisit", domain.ErrBadRequest)
	}
	return p, nil
}

func (s *Academic) ActivateYear(ctx context.Context, id pgtype.UUID, confirmation string) (db.AcademicYear, error) {
	if strings.ToUpper(strings.TrimSpace(confirmation)) != academicYearActivationChallenge {
		return db.AcademicYear{}, fmt.Errorf("%w: konfirmasi aktivasi tahun ajaran tidak sesuai", domain.ErrBadRequest)
	}
	var row db.AcademicYear
	err := s.withAcademicStore(ctx, func(store academicStore) error {
		current, err := store.GetAcademicYearByID(ctx, id)
		if err != nil {
			return err
		}
		if current.IsActive {
			row = current
			return nil
		}
		if err := store.DeactivateAcademicYears(ctx); err != nil {
			return err
		}
		activated, err := store.ActivateAcademicYear(ctx, id)
		if err != nil {
			return err
		}
		row = activated
		return nil
	})
	return row, err
}

func (s *Academic) PreviewYearRollover(ctx context.Context, input YearRolloverPreviewInput) (YearRolloverPreview, error) {
	if !input.TargetAcademicYearID.Valid {
		return YearRolloverPreview{}, fmt.Errorf("%w: tahun ajaran tujuan wajib dipilih", domain.ErrBadRequest)
	}
	source, target, err := resolveYearRolloverYears(ctx, s.q, input.SourceAcademicYearID, input.TargetAcademicYearID)
	if err != nil {
		return YearRolloverPreview{}, err
	}

	classes, err := s.q.ListSchoolClasses(ctx)
	if err != nil {
		return YearRolloverPreview{}, err
	}
	students, err := s.q.ListYearRolloverStudents(ctx, source.ID)
	if err != nil {
		return YearRolloverPreview{}, err
	}
	homerooms, err := s.q.ListYearRolloverHomeroomAssignments(ctx, source.ID)
	if err != nil {
		return YearRolloverPreview{}, err
	}
	assignments, err := s.q.ListClassSubjectAssignments(ctx)
	if err != nil {
		return YearRolloverPreview{}, err
	}
	slots, err := s.q.ListTimetableSlots(ctx)
	if err != nil {
		return YearRolloverPreview{}, err
	}

	sourceClasses := make(map[string]db.ListSchoolClassesRow)
	targetByCode := make(map[string]db.ListSchoolClassesRow)
	for _, class := range classes {
		switch pgUUIDString(class.AcademicYearID) {
		case pgUUIDString(source.ID):
			if class.IsActive {
				sourceClasses[pgUUIDString(class.ID)] = class
			}
		case pgUUIDString(target.ID):
			targetByCode[normalizeLookupKey(class.Code)] = class
		}
	}

	targetBySourceClass := make(map[string]db.ListSchoolClassesRow)
	classesToCreate := []YearRolloverClassToCreate{}
	inactiveTargetClasses := map[string]bool{}
	for _, class := range sortedClasses(sourceClasses) {
		nextLevel, ok := nextAcademicLevel(class.Level)
		if !ok {
			continue
		}
		targetCode := rolloverNextLabel(class.Code, class.Level, nextLevel)
		targetName := rolloverNextLabel(class.Name, class.Level, nextLevel)
		targetClass, exists := targetByCode[normalizeLookupKey(targetCode)]
		if exists {
			if !targetClass.IsActive {
				inactiveTargetClasses[pgUUIDString(class.ID)] = true
				continue
			}
			targetBySourceClass[pgUUIDString(class.ID)] = targetClass
			continue
		}
		targetBySourceClass[pgUUIDString(class.ID)] = plannedRolloverClass(target, targetCode, targetName, nextLevel)
		classesToCreate = append(classesToCreate, YearRolloverClassToCreate{
			SourceClassID: pgUUIDString(class.ID),
			SourceCode:    class.Code,
			SourceName:    class.Name,
			SourceLevel:   class.Level,
			TargetCode:    targetCode,
			TargetName:    targetName,
			TargetLevel:   nextLevel,
		})
	}

	studentsToPromote := []YearRolloverStudentMove{}
	studentsWithoutNext := []YearRolloverStudentWarning{}
	for _, student := range students {
		sourceClassID := pgUUIDString(student.ClassID)
		targetClass, exists := targetBySourceClass[sourceClassID]
		if !exists {
			reason := "Rombel tujuan belum ada"
			if _, ok := nextAcademicLevel(student.ClassLevel); !ok {
				reason = "Tingkat akhir, perlu proses kelulusan atau mutasi manual"
			} else if inactiveTargetClasses[sourceClassID] {
				reason = "Rombel tujuan ada tetapi nonaktif"
			}
			studentsWithoutNext = append(studentsWithoutNext, YearRolloverStudentWarning{
				StudentID:     pgUUIDString(student.ID),
				NIS:           student.Nis,
				NISN:          student.Nisn,
				Nama:          student.Nama,
				FromClassID:   sourceClassID,
				FromClassCode: student.ClassCode,
				Reason:        reason,
			})
			continue
		}
		studentsToPromote = append(studentsToPromote, YearRolloverStudentMove{
			StudentID:     pgUUIDString(student.ID),
			NIS:           student.Nis,
			NISN:          student.Nisn,
			Nama:          student.Nama,
			FromClassID:   sourceClassID,
			FromClassCode: student.ClassCode,
			ToClassID:     pgUUIDString(targetClass.ID),
			ToClassCode:   targetClass.Code,
			ToClassName:   targetClass.Name,
		})
	}

	assignmentsByID := map[string]db.ListClassSubjectAssignmentsRow{}
	subjectAssignmentsCopy := 0
	for _, assignment := range assignments {
		assignmentsByID[pgUUIDString(assignment.ID)] = assignment
		if _, ok := targetBySourceClass[pgUUIDString(assignment.ClassID)]; ok {
			subjectAssignmentsCopy++
		}
	}
	timetableSlotsCopy := 0
	for _, slot := range slots {
		assignment, ok := assignmentsByID[pgUUIDString(slot.AssignmentID)]
		if ok {
			if _, hasTarget := targetBySourceClass[pgUUIDString(assignment.ClassID)]; hasTarget {
				timetableSlotsCopy++
			}
		}
	}
	homeroomCopy := 0
	for _, row := range homerooms {
		if _, ok := targetBySourceClass[pgUUIDString(row.ClassID)]; ok {
			homeroomCopy += int(row.Total)
		}
	}

	warnings := []string{
		"Pratinjau ini belum menyimpan perubahan.",
		"Penerapan kenaikan kelas wajib memakai kalimat konfirmasi dan berjalan aman; data tahun lama tidak dihapus.",
	}
	if len(classesToCreate) > 0 {
		warnings = append(warnings, "Beberapa rombel tujuan belum ada dan akan dibuat saat penerapan kenaikan kelas.")
	}
	if len(studentsWithoutNext) > 0 {
		warnings = append(warnings, "Ada siswa yang belum punya rombel tujuan atau berada di tingkat akhir.")
	}
	if len(inactiveTargetClasses) > 0 {
		warnings = append(warnings, "Ada rombel tujuan yang sudah ada tetapi nonaktif; aktifkan terlebih dahulu bila ingin dipakai.")
	}

	return YearRolloverPreview{
		SourceAcademicYearID:   pgUUIDString(source.ID),
		SourceAcademicYearName: source.Name,
		TargetAcademicYearID:   pgUUIDString(target.ID),
		TargetAcademicYearName: target.Name,
		ApplyChallenge:         yearRolloverApplyChallenge(source.Name, target.Name),
		Counts: YearRolloverPreviewCounts{
			ClassesToCreate:          len(classesToCreate),
			StudentsToPromote:        len(studentsToPromote),
			StudentsWithoutNextClass: len(studentsWithoutNext),
			HomeroomAssignmentsCopy:  homeroomCopy,
			SubjectAssignmentsCopy:   subjectAssignmentsCopy,
			TimetableSlotsCopy:       timetableSlotsCopy,
		},
		ClassesToCreate:     classesToCreate,
		StudentsToPromote:   studentsToPromote,
		StudentsWithoutNext: studentsWithoutNext,
		Warnings:            warnings,
	}, nil
}

func (s *Academic) ApplyYearRollover(ctx context.Context, input YearRolloverApplyInput) (YearRolloverApplyResult, error) {
	if !input.TargetAcademicYearID.Valid {
		return YearRolloverApplyResult{}, fmt.Errorf("%w: tahun ajaran tujuan wajib dipilih", domain.ErrBadRequest)
	}
	var result YearRolloverApplyResult
	err := s.withAcademicStore(ctx, func(store academicStore) error {
		source, target, err := resolveYearRolloverYears(ctx, store, input.SourceAcademicYearID, input.TargetAcademicYearID)
		if err != nil {
			return err
		}
		expectedChallenge := yearRolloverApplyChallenge(source.Name, target.Name)
		if !yearRolloverApplyAuthorized(input, expectedChallenge) {
			return fmt.Errorf("%w: kalimat konfirmasi penerapan kenaikan kelas tidak sesuai", domain.ErrBadRequest)
		}
		applyResult, err := applyYearRolloverInStore(ctx, store, source, target)
		if err != nil {
			return err
		}
		result = applyResult
		return nil
	})
	return result, err
}

type rolloverAssignmentRef struct {
	ID                pgtype.UUID
	ClassID           pgtype.UUID
	SubjectID         pgtype.UUID
	TeacherEmployeeID pgtype.UUID
}

func resolveYearRolloverYears(ctx context.Context, store academicStore, sourceID, targetID pgtype.UUID) (db.AcademicYear, db.AcademicYear, error) {
	target, err := store.GetAcademicYearByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.AcademicYear{}, db.AcademicYear{}, fmt.Errorf("%w: tahun ajaran tujuan tidak ditemukan", domain.ErrBadRequest)
		}
		return db.AcademicYear{}, db.AcademicYear{}, err
	}
	var source db.AcademicYear
	if sourceID.Valid {
		source, err = store.GetAcademicYearByID(ctx, sourceID)
	} else {
		source, err = store.GetActiveAcademicYear(ctx)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.AcademicYear{}, db.AcademicYear{}, fmt.Errorf("%w: tahun ajaran sumber tidak ditemukan", domain.ErrBadRequest)
		}
		return db.AcademicYear{}, db.AcademicYear{}, err
	}
	if pgUUIDString(source.ID) == pgUUIDString(target.ID) {
		return db.AcademicYear{}, db.AcademicYear{}, fmt.Errorf("%w: tahun ajaran sumber dan tujuan harus berbeda", domain.ErrBadRequest)
	}
	return source, target, nil
}

func applyYearRolloverInStore(ctx context.Context, store academicStore, source, target db.AcademicYear) (YearRolloverApplyResult, error) {
	result := YearRolloverApplyResult{
		SourceAcademicYearID:   pgUUIDString(source.ID),
		SourceAcademicYearName: source.Name,
		TargetAcademicYearID:   pgUUIDString(target.ID),
		TargetAcademicYearName: target.Name,
		Warnings: []string{
			"Penerapan kenaikan kelas berjalan aman; data tahun lama tidak dihapus.",
		},
	}

	classes, err := store.ListSchoolClasses(ctx)
	if err != nil {
		return YearRolloverApplyResult{}, err
	}
	sourceClasses, targetByCode := splitRolloverClasses(classes, source.ID, target.ID)
	targetBySourceClass := map[string]db.ListSchoolClassesRow{}
	inactiveTargetClasses := map[string]bool{}
	for _, class := range sortedClasses(sourceClasses) {
		nextLevel, ok := nextAcademicLevel(class.Level)
		if !ok {
			continue
		}
		sourceClassKey := pgUUIDString(class.ID)
		targetCode := rolloverNextLabel(class.Code, class.Level, nextLevel)
		targetName := rolloverNextLabel(class.Name, class.Level, nextLevel)
		if targetClass, exists := targetByCode[normalizeLookupKey(targetCode)]; exists {
			if !targetClass.IsActive {
				inactiveTargetClasses[sourceClassKey] = true
				result.Classes = append(result.Classes, rolloverClassApplyRow(class, targetClass, "skipped_inactive_target"))
				result.Warnings = append(result.Warnings, fmt.Sprintf("Rombel tujuan %s sudah ada tetapi nonaktif; data dari %s dilewati.", targetClass.Code, class.Code))
				continue
			}
			targetBySourceClass[sourceClassKey] = targetClass
			result.Counts.ClassesReused++
			result.Classes = append(result.Classes, rolloverClassApplyRow(class, targetClass, "reused"))
			continue
		}
		created, err := store.CreateSchoolClass(ctx, db.CreateSchoolClassParams{
			AcademicYearID: target.ID,
			Code:           targetCode,
			Name:           targetName,
			Level:          nextLevel,
			IsActive:       true,
		})
		if err != nil {
			return YearRolloverApplyResult{}, err
		}
		targetClass := schoolClassToListRow(created, target.Name)
		targetByCode[normalizeLookupKey(targetCode)] = targetClass
		targetBySourceClass[sourceClassKey] = targetClass
		result.Counts.ClassesCreated++
		result.Classes = append(result.Classes, rolloverClassApplyRow(class, targetClass, "created"))
	}

	if result.Counts.ClassesCreated > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%d rombel tujuan baru dibuat.", result.Counts.ClassesCreated))
	}
	if result.Counts.ClassesReused > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%d rombel tujuan yang sudah ada dipakai ulang.", result.Counts.ClassesReused))
	}

	if err := copyYearRolloverHomerooms(ctx, store, source, target, targetBySourceClass, &result); err != nil {
		return YearRolloverApplyResult{}, err
	}
	targetAssignmentsBySource, err := copyYearRolloverAssignments(ctx, store, targetBySourceClass, &result)
	if err != nil {
		return YearRolloverApplyResult{}, err
	}
	if err := copyYearRolloverTimetableSlots(ctx, store, targetAssignmentsBySource, &result); err != nil {
		return YearRolloverApplyResult{}, err
	}
	if err := promoteYearRolloverStudents(ctx, store, source, targetBySourceClass, inactiveTargetClasses, &result); err != nil {
		return YearRolloverApplyResult{}, err
	}
	result.Counts.StudentsSkipped = len(result.StudentsSkipped)
	if result.Counts.StudentsSkipped > 0 {
		result.Warnings = append(result.Warnings, "Sebagian siswa dilewati dan perlu tindak lanjut manual.")
	}
	return result, nil
}

func copyYearRolloverHomerooms(ctx context.Context, store academicStore, source, target db.AcademicYear, targetBySourceClass map[string]db.ListSchoolClassesRow, result *YearRolloverApplyResult) error {
	homerooms, err := store.ListYearRolloverHomeroomAssignmentDetails(ctx, source.ID)
	if err != nil {
		return err
	}
	for _, homeroom := range homerooms {
		targetClass, ok := targetBySourceClass[pgUUIDString(homeroom.ClassID)]
		if !ok {
			continue
		}
		existing, err := store.CountActiveHomeroomAssignmentByClass(ctx, targetClass.ID)
		if err != nil {
			return err
		}
		if existing > 0 {
			continue
		}
		_, err = store.CreateHomeroomAssignment(ctx, db.CreateHomeroomAssignmentParams{
			ClassID:                targetClass.ID,
			HomeroomIsActive:       true,
			EmployeeID:             homeroom.EmployeeID,
			HomeroomAcademicYearID: target.ID,
			HomeroomStartDate:      target.StartDate,
			HomeroomEndDate:        pgtype.Date{},
			Notes:                  rolloverCopiedNotes(homeroom.Notes, source.Name),
		})
		if err != nil {
			return err
		}
		result.Counts.HomeroomsCopied++
	}
	return nil
}

func copyYearRolloverAssignments(ctx context.Context, store academicStore, targetBySourceClass map[string]db.ListSchoolClassesRow, result *YearRolloverApplyResult) (map[string]rolloverAssignmentRef, error) {
	assignments, err := store.ListClassSubjectAssignments(ctx)
	if err != nil {
		return nil, err
	}
	targetAssignmentsBySource := map[string]rolloverAssignmentRef{}
	targetByClassSubject := map[string]rolloverAssignmentRef{}
	for _, assignment := range assignments {
		targetByClassSubject[classSubjectKey(assignment.ClassID, assignment.SubjectID)] = assignmentRefFromList(assignment)
	}
	for _, assignment := range assignments {
		targetClass, ok := targetBySourceClass[pgUUIDString(assignment.ClassID)]
		if !ok {
			continue
		}
		targetKey := classSubjectKey(targetClass.ID, assignment.SubjectID)
		if existing, exists := targetByClassSubject[targetKey]; exists {
			targetAssignmentsBySource[pgUUIDString(assignment.ID)] = existing
			continue
		}
		created, err := store.CreateClassSubjectAssignment(ctx, db.CreateClassSubjectAssignmentParams{
			ClassID:           targetClass.ID,
			SubjectID:         assignment.SubjectID,
			TeacherEmployeeID: assignment.TeacherEmployeeID,
		})
		if err != nil {
			return nil, err
		}
		ref := assignmentRefFromModel(created)
		targetByClassSubject[targetKey] = ref
		targetAssignmentsBySource[pgUUIDString(assignment.ID)] = ref
		result.Counts.AssignmentsCopied++
	}
	return targetAssignmentsBySource, nil
}

func copyYearRolloverTimetableSlots(ctx context.Context, store academicStore, targetAssignmentsBySource map[string]rolloverAssignmentRef, result *YearRolloverApplyResult) error {
	slots, err := store.ListTimetableSlots(ctx)
	if err != nil {
		return err
	}
	targetAssignmentIDs := map[string]bool{}
	for _, assignment := range targetAssignmentsBySource {
		targetAssignmentIDs[pgUUIDString(assignment.ID)] = true
	}
	existingTargetSlots := map[string]bool{}
	sourceSlots := []db.ListTimetableSlotsRow{}
	for _, slot := range slots {
		if targetAssignmentIDs[pgUUIDString(slot.AssignmentID)] {
			existingTargetSlots[timetableSlotKey(slot.AssignmentID, slot.DayOfWeek, slot.StartTime, slot.EndTime, slot.RoomLabel)] = true
		}
		if _, ok := targetAssignmentsBySource[pgUUIDString(slot.AssignmentID)]; ok {
			sourceSlots = append(sourceSlots, slot)
		}
	}
	for _, slot := range sourceSlots {
		targetAssignment := targetAssignmentsBySource[pgUUIDString(slot.AssignmentID)]
		key := timetableSlotKey(targetAssignment.ID, slot.DayOfWeek, slot.StartTime, slot.EndTime, slot.RoomLabel)
		if existingTargetSlots[key] {
			continue
		}
		_, err := store.CreateTimetableSlot(ctx, db.CreateTimetableSlotParams{
			AssignmentID: targetAssignment.ID,
			DayOfWeek:    slot.DayOfWeek,
			StartTime:    slot.StartTime,
			EndTime:      slot.EndTime,
			RoomLabel:    slot.RoomLabel,
			Notes:        slot.Notes,
		})
		if err != nil {
			return err
		}
		existingTargetSlots[key] = true
		result.Counts.TimetableSlotsCopied++
	}
	return nil
}

func promoteYearRolloverStudents(ctx context.Context, store academicStore, source db.AcademicYear, targetBySourceClass map[string]db.ListSchoolClassesRow, inactiveTargetClasses map[string]bool, result *YearRolloverApplyResult) error {
	students, err := store.ListYearRolloverStudents(ctx, source.ID)
	if err != nil {
		return err
	}
	for _, student := range students {
		sourceClassID := pgUUIDString(student.ClassID)
		targetClass, exists := targetBySourceClass[sourceClassID]
		if !exists {
			reason := "Rombel tujuan belum tersedia"
			if _, ok := nextAcademicLevel(student.ClassLevel); !ok {
				reason = "Tingkat akhir, perlu proses kelulusan atau mutasi manual"
			} else if inactiveTargetClasses[sourceClassID] {
				reason = "Rombel tujuan ada tetapi nonaktif"
			}
			result.StudentsSkipped = append(result.StudentsSkipped, rolloverStudentWarning(student, reason))
			continue
		}
		affected, err := store.PromoteYearRolloverStudent(ctx, db.PromoteYearRolloverStudentParams{
			TargetClassID: targetClass.ID,
			StudentID:     student.ID,
			SourceClassID: student.ClassID,
		})
		if err != nil {
			return err
		}
		if affected == 0 {
			result.StudentsSkipped = append(result.StudentsSkipped, rolloverStudentWarning(student, "Data siswa sudah berubah sebelum penerapan kenaikan kelas selesai"))
			continue
		}
		result.Counts.StudentsPromoted++
		result.StudentsPromoted = append(result.StudentsPromoted, YearRolloverStudentMove{
			StudentID:     pgUUIDString(student.ID),
			NIS:           student.Nis,
			NISN:          student.Nisn,
			Nama:          student.Nama,
			FromClassID:   sourceClassID,
			FromClassCode: student.ClassCode,
			ToClassID:     pgUUIDString(targetClass.ID),
			ToClassCode:   targetClass.Code,
			ToClassName:   targetClass.Name,
		})
	}
	return nil
}

func (s *Academic) DryRunAcademicImport(ctx context.Context, input AcademicImportDryRunInput) (AcademicImportDryRunResult, error) {
	kind := normalizeImportKind(input.Kind)
	if kind == "" {
		return AcademicImportDryRunResult{}, fmt.Errorf("%w: jenis data akademik yang akan diperiksa tidak valid", domain.ErrBadRequest)
	}
	records, err := readAcademicCSV(input.CSV)
	if err != nil {
		return AcademicImportDryRunResult{}, err
	}
	result := AcademicImportDryRunResult{Kind: kind}
	if len(records) == 0 {
		return result, nil
	}
	header, err := academicImportHeader(kind, records[0])
	if err != nil {
		return AcademicImportDryRunResult{}, err
	}
	refs, err := s.importDryRunReferences(ctx)
	if err != nil {
		return AcademicImportDryRunResult{}, err
	}
	seen := map[string]int{}
	for idx, record := range records[1:] {
		rowNumber := idx + 2
		if csvRowBlank(record) {
			result.SkipCount++
			result.Rows = append(result.Rows, AcademicImportRowPlan{Row: rowNumber, Action: "skip", Summary: "Baris kosong"})
			continue
		}
		result.TotalRows++
		plan, rowErrors := validateAcademicImportRow(kind, rowNumber, record, header, refs, seen)
		if len(rowErrors) > 0 {
			result.ErrorCount++
			result.RowErrors = append(result.RowErrors, rowErrors...)
			result.Rows = append(result.Rows, AcademicImportRowPlan{Row: rowNumber, Action: "needs_review", Summary: rowErrors[0].Message})
			continue
		}
		switch plan.Action {
		case "add":
			result.AddCount++
		case "update":
			result.UpdateCount++
		default:
			result.SkipCount++
		}
		result.Rows = append(result.Rows, plan)
	}
	return result, nil
}

type importDryRunReferences struct {
	activeYearID                     string
	yearsByName                      map[string]db.AcademicYear
	classesByYearCode                map[string]db.ListSchoolClassesRow
	activeClassesByCode              map[string]db.ListSchoolClassesRow
	subjectsByCode                   map[string]db.ListSubjectsRow
	studentsByNIS                    map[string]db.ListAcademicImportStudentsRow
	teachersByNIP                    map[string]db.ListAcademicImportTeachersRow
	teachersByName                   map[string]db.ListAcademicImportTeachersRow
	assignmentsByClassSubject        map[string]db.ListClassSubjectAssignmentsRow
	assignmentsByClassSubjectTeacher map[string]db.ListClassSubjectAssignmentsRow
	slotsByAssignmentDayTime         map[string]db.ListTimetableSlotsRow
}

func (s *Academic) importDryRunReferences(ctx context.Context) (importDryRunReferences, error) {
	years, err := s.q.ListAcademicYears(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	classes, err := s.q.ListSchoolClasses(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	subjects, err := s.q.ListSubjects(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	assignments, err := s.q.ListClassSubjectAssignments(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	slots, err := s.q.ListTimetableSlots(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	students, err := s.q.ListAcademicImportStudents(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	teachers, err := s.q.ListAcademicImportTeachers(ctx)
	if err != nil {
		return importDryRunReferences{}, err
	}
	var activeYearID string
	for _, year := range years {
		if year.IsActive {
			activeYearID = pgUUIDString(year.ID)
			break
		}
	}
	refs := importDryRunReferences{
		activeYearID:                     activeYearID,
		yearsByName:                      map[string]db.AcademicYear{},
		classesByYearCode:                map[string]db.ListSchoolClassesRow{},
		activeClassesByCode:              map[string]db.ListSchoolClassesRow{},
		subjectsByCode:                   map[string]db.ListSubjectsRow{},
		studentsByNIS:                    map[string]db.ListAcademicImportStudentsRow{},
		teachersByNIP:                    map[string]db.ListAcademicImportTeachersRow{},
		teachersByName:                   map[string]db.ListAcademicImportTeachersRow{},
		assignmentsByClassSubject:        map[string]db.ListClassSubjectAssignmentsRow{},
		assignmentsByClassSubjectTeacher: map[string]db.ListClassSubjectAssignmentsRow{},
		slotsByAssignmentDayTime:         map[string]db.ListTimetableSlotsRow{},
	}
	for _, year := range years {
		refs.yearsByName[normalizeLookupKey(year.Name)] = year
	}
	for _, class := range classes {
		yearClassKey := importPairKey(pgUUIDString(class.AcademicYearID), class.Code)
		refs.classesByYearCode[yearClassKey] = class
		if activeYearID != "" && pgUUIDString(class.AcademicYearID) == activeYearID && class.IsActive {
			refs.activeClassesByCode[normalizeLookupKey(class.Code)] = class
		}
	}
	for _, subject := range subjects {
		if subject.IsActive {
			refs.subjectsByCode[normalizeLookupKey(subject.Code)] = subject
		}
	}
	for _, student := range students {
		refs.studentsByNIS[normalizeLookupKey(student.Nis)] = student
	}
	for _, teacher := range teachers {
		if teacher.Nip != "" {
			refs.teachersByNIP[normalizeLookupKey(teacher.Nip)] = teacher
		}
		refs.teachersByName[normalizeLookupKey(teacher.Nama)] = teacher
	}
	for _, assignment := range assignments {
		classSubjectKey := importPairKey(pgUUIDString(assignment.ClassID), pgUUIDString(assignment.SubjectID))
		refs.assignmentsByClassSubject[classSubjectKey] = assignment
		refs.assignmentsByClassSubjectTeacher[importTripleKey(pgUUIDString(assignment.ClassID), pgUUIDString(assignment.SubjectID), pgUUIDString(assignment.TeacherEmployeeID))] = assignment
	}
	for _, slot := range slots {
		refs.slotsByAssignmentDayTime[importTripleKey(pgUUIDString(slot.AssignmentID), strconv.Itoa(int(slot.DayOfWeek)), formatAcademicTime(slot.StartTime))+"|"+formatAcademicTime(slot.EndTime)] = slot
	}
	return refs, nil
}

func validateAcademicImportRow(kind string, rowNumber int, row []string, header map[string]int, refs importDryRunReferences, seen map[string]int) (AcademicImportRowPlan, []AcademicImportRowError) {
	switch kind {
	case "siswa":
		return validateStudentImportRow(rowNumber, row, header, refs, seen)
	case "rombel":
		return validateClassImportRow(rowNumber, row, header, refs, seen)
	case "guru_mapel":
		return validateSubjectTeacherImportRow(rowNumber, row, header, refs, seen)
	case "jadwal":
		return validateTimetableImportRow(rowNumber, row, header, refs, seen)
	default:
		return AcademicImportRowPlan{}, []AcademicImportRowError{{Row: rowNumber, Message: "Jenis data yang akan diimpor tidak valid"}}
	}
}

func validateStudentImportRow(rowNumber int, row []string, header map[string]int, refs importDryRunReferences, seen map[string]int) (AcademicImportRowPlan, []AcademicImportRowError) {
	nis := cell(row, header, "nis")
	nama := cell(row, header, "nama")
	gender := cell(row, header, "jenis kelamin")
	classCode := cell(row, header, "kode rombel")
	errs := []AcademicImportRowError{}
	requireCell(&errs, rowNumber, "NIS", nis)
	requireCell(&errs, rowNumber, "Nama", nama)
	if normalizeGender(gender) == "" {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Jenis Kelamin", Message: "Jenis kelamin harus L/P atau Laki-laki/Perempuan"})
	}
	if classCode != "" {
		if _, ok := refs.activeClassesByCode[normalizeLookupKey(classCode)]; !ok {
			errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Kode Rombel", Message: "Rombel aktif tidak ditemukan"})
		}
	}
	key := "siswa|" + normalizeLookupKey(nis)
	if seenRow, duplicate := seen[key]; duplicate {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "NIS", Message: fmt.Sprintf("NIS duplikat dengan baris %d", seenRow)})
	}
	seen[key] = rowNumber
	if len(errs) > 0 {
		return AcademicImportRowPlan{}, errs
	}
	if _, exists := refs.studentsByNIS[normalizeLookupKey(nis)]; exists {
		return AcademicImportRowPlan{Row: rowNumber, Action: "update", Summary: "Perbarui siswa " + nama + " (" + nis + ")"}, nil
	}
	return AcademicImportRowPlan{Row: rowNumber, Action: "add", Summary: "Tambah siswa " + nama + " (" + nis + ")"}, nil
}

func validateClassImportRow(rowNumber int, row []string, header map[string]int, refs importDryRunReferences, seen map[string]int) (AcademicImportRowPlan, []AcademicImportRowError) {
	code := cell(row, header, "kode rombel")
	name := cell(row, header, "nama rombel")
	level := cell(row, header, "tingkat")
	yearName := cell(row, header, "tahun ajaran")
	errs := []AcademicImportRowError{}
	requireCell(&errs, rowNumber, "Kode Rombel", code)
	requireCell(&errs, rowNumber, "Nama Rombel", name)
	if _, ok := normalizeAcademicLevel(level); !ok {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Tingkat", Message: "Tingkat harus VII, VIII, atau IX"})
	}
	year, ok := refs.yearsByName[normalizeLookupKey(yearName)]
	if !ok {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Tahun Ajaran", Message: "Tahun ajaran tidak ditemukan"})
	}
	key := "rombel|" + normalizeLookupKey(yearName) + "|" + normalizeLookupKey(code)
	if seenRow, duplicate := seen[key]; duplicate {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Kode Rombel", Message: fmt.Sprintf("Rombel duplikat dengan baris %d", seenRow)})
	}
	seen[key] = rowNumber
	if len(errs) > 0 {
		return AcademicImportRowPlan{}, errs
	}
	if _, exists := refs.classesByYearCode[importPairKey(pgUUIDString(year.ID), code)]; exists {
		return AcademicImportRowPlan{Row: rowNumber, Action: "update", Summary: "Perbarui rombel " + code + " tahun " + yearName}, nil
	}
	return AcademicImportRowPlan{Row: rowNumber, Action: "add", Summary: "Tambah rombel " + code + " tahun " + yearName}, nil
}

func validateSubjectTeacherImportRow(rowNumber int, row []string, header map[string]int, refs importDryRunReferences, seen map[string]int) (AcademicImportRowPlan, []AcademicImportRowError) {
	class, subject, teacher, errs := resolveAcademicAssignmentRefs(rowNumber, row, header, refs)
	key := "guru_mapel|" + normalizeLookupKey(cell(row, header, "kode rombel")) + "|" + normalizeLookupKey(cell(row, header, "kode mapel"))
	if seenRow, duplicate := seen[key]; duplicate {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Kode Mapel", Message: fmt.Sprintf("Penugasan guru mapel duplikat dengan baris %d", seenRow)})
	}
	seen[key] = rowNumber
	if len(errs) > 0 {
		return AcademicImportRowPlan{}, errs
	}
	assignment, exists := refs.assignmentsByClassSubject[importPairKey(pgUUIDString(class.ID), pgUUIDString(subject.ID))]
	if !exists {
		return AcademicImportRowPlan{Row: rowNumber, Action: "add", Summary: "Tambah guru mapel " + subject.Code + " untuk " + class.Code}, nil
	}
	if pgUUIDString(assignment.TeacherEmployeeID) == pgUUIDString(teacher.ID) {
		return AcademicImportRowPlan{Row: rowNumber, Action: "skip", Summary: "Guru mapel " + subject.Code + " untuk " + class.Code + " sudah sama"}, nil
	}
	return AcademicImportRowPlan{Row: rowNumber, Action: "update", Summary: "Perbarui guru mapel " + subject.Code + " untuk " + class.Code}, nil
}

func validateTimetableImportRow(rowNumber int, row []string, header map[string]int, refs importDryRunReferences, seen map[string]int) (AcademicImportRowPlan, []AcademicImportRowError) {
	class, subject, teacher, errs := resolveAcademicAssignmentRefs(rowNumber, row, header, refs)
	dayValue := cell(row, header, "hari")
	startValue := cell(row, header, "jam mulai")
	endValue := cell(row, header, "jam selesai")
	day, ok := parseAcademicDay(dayValue)
	if !ok {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Hari", Message: "Hari harus Senin-Sabtu atau angka 1-6"})
	}
	start, startErr := ParseAcademicTimeInput(startValue)
	if startErr != nil {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Jam Mulai", Message: startErr.Error()})
	}
	end, endErr := ParseAcademicTimeInput(endValue)
	if endErr != nil {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Jam Selesai", Message: endErr.Error()})
	}
	if startErr == nil && endErr == nil && start.Microseconds >= end.Microseconds {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Jam Selesai", Message: "Jam selesai harus setelah jam mulai"})
	}
	assignment, assignmentExists := refs.assignmentsByClassSubjectTeacher[importTripleKey(pgUUIDString(class.ID), pgUUIDString(subject.ID), pgUUIDString(teacher.ID))]
	if len(errs) == 0 && !assignmentExists {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Guru Mapel", Message: "Penugasan guru mapel belum ada"})
	}
	key := "jadwal|" + normalizeLookupKey(cell(row, header, "kode rombel")) + "|" + normalizeLookupKey(cell(row, header, "kode mapel")) + "|" + strconv.Itoa(day) + "|" + normalizeImportTime(startValue) + "|" + normalizeImportTime(endValue)
	if seenRow, duplicate := seen[key]; duplicate {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Jam Mulai", Message: fmt.Sprintf("Jam pelajaran duplikat dengan baris %d", seenRow)})
	}
	seen[key] = rowNumber
	if len(errs) > 0 {
		return AcademicImportRowPlan{}, errs
	}
	slotKey := importTripleKey(pgUUIDString(assignment.ID), strconv.Itoa(day), normalizeImportTime(startValue)) + "|" + normalizeImportTime(endValue)
	if _, exists := refs.slotsByAssignmentDayTime[slotKey]; exists {
		return AcademicImportRowPlan{Row: rowNumber, Action: "update", Summary: "Perbarui jam pelajaran " + class.Code + " " + subject.Code + " " + dayValue}, nil
	}
	return AcademicImportRowPlan{Row: rowNumber, Action: "add", Summary: "Tambah jam pelajaran " + class.Code + " " + subject.Code + " " + dayValue}, nil
}

func resolveAcademicAssignmentRefs(rowNumber int, row []string, header map[string]int, refs importDryRunReferences) (db.ListSchoolClassesRow, db.ListSubjectsRow, db.ListAcademicImportTeachersRow, []AcademicImportRowError) {
	classCode := cell(row, header, "kode rombel")
	subjectCode := cell(row, header, "kode mapel")
	teacherNIP := cell(row, header, "nip guru")
	teacherName := cell(row, header, "nama guru")
	errs := []AcademicImportRowError{}
	class, classOK := refs.activeClassesByCode[normalizeLookupKey(classCode)]
	if !classOK {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Kode Rombel", Message: "Rombel aktif tidak ditemukan"})
	}
	subject, subjectOK := refs.subjectsByCode[normalizeLookupKey(subjectCode)]
	if !subjectOK {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "Kode Mapel", Message: "Mapel aktif tidak ditemukan"})
	}
	var teacher db.ListAcademicImportTeachersRow
	teacherOK := false
	if teacherNIP != "" {
		teacher, teacherOK = refs.teachersByNIP[normalizeLookupKey(teacherNIP)]
	} else if teacherName != "" {
		teacher, teacherOK = refs.teachersByName[normalizeLookupKey(teacherName)]
	}
	if !teacherOK {
		errs = append(errs, AcademicImportRowError{Row: rowNumber, Field: "NIP Guru", Message: "Guru aktif tidak ditemukan"})
	}
	return class, subject, teacher, errs
}

func readAcademicCSV(raw string) ([][]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("%w: file impor kosong", domain.ErrBadRequest)
	}
	reader := csv.NewReader(bytes.NewReader([]byte(strings.TrimPrefix(raw, "\ufeff"))))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	records := [][]string{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: data impor tidak valid pada baris %d", domain.ErrBadRequest, len(records)+1)
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%w: file impor kosong", domain.ErrBadRequest)
	}
	return records, nil
}

func academicImportHeader(kind string, row []string) (map[string]int, error) {
	required := map[string][]string{
		"siswa":      {"nis", "nisn", "nama", "jenis kelamin", "kode rombel", "status"},
		"rombel":     {"kode rombel", "nama rombel", "tingkat", "tahun ajaran", "aktif"},
		"guru_mapel": {"kode rombel", "kode mapel", "nip guru", "nama guru"},
		"jadwal":     {"kode rombel", "kode mapel", "nip guru", "nama guru", "hari", "jam mulai", "jam selesai", "ruang", "catatan"},
	}
	index := map[string]int{}
	for i, cell := range row {
		index[normalizeLookupKey(cell)] = i
	}
	missing := []string{}
	for _, field := range required[kind] {
		if _, ok := index[field]; !ok {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: kolom data impor kurang: %s", domain.ErrBadRequest, strings.Join(missing, ", "))
	}
	return index, nil
}

func normalizeImportKind(value string) string {
	switch normalizeLookupKey(value) {
	case "siswa", "student", "students":
		return "siswa"
	case "rombel", "kelas", "class", "classes":
		return "rombel"
	case "guru_mapel", "guru mapel", "assignment", "assignments":
		return "guru_mapel"
	case "jadwal", "jadwal pelajaran", "timetable":
		return "jadwal"
	default:
		return ""
	}
}

func splitRolloverClasses(classes []db.ListSchoolClassesRow, sourceYearID, targetYearID pgtype.UUID) (map[string]db.ListSchoolClassesRow, map[string]db.ListSchoolClassesRow) {
	sourceClasses := make(map[string]db.ListSchoolClassesRow)
	targetByCode := make(map[string]db.ListSchoolClassesRow)
	for _, class := range classes {
		switch pgUUIDString(class.AcademicYearID) {
		case pgUUIDString(sourceYearID):
			if class.IsActive {
				sourceClasses[pgUUIDString(class.ID)] = class
			}
		case pgUUIDString(targetYearID):
			targetByCode[normalizeLookupKey(class.Code)] = class
		}
	}
	return sourceClasses, targetByCode
}

func sortedClasses(classes map[string]db.ListSchoolClassesRow) []db.ListSchoolClassesRow {
	out := make([]db.ListSchoolClassesRow, 0, len(classes))
	for _, class := range classes {
		out = append(out, class)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Level != out[j].Level {
			return out[i].Level < out[j].Level
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func plannedRolloverClass(target db.AcademicYear, code, name, level string) db.ListSchoolClassesRow {
	return db.ListSchoolClassesRow{
		Code:             code,
		Name:             name,
		Level:            level,
		IsActive:         true,
		AcademicYearID:   target.ID,
		AcademicYearName: target.Name,
	}
}

func schoolClassToListRow(class db.SchoolClass, academicYearName string) db.ListSchoolClassesRow {
	return db.ListSchoolClassesRow{
		ID:               class.ID,
		Code:             class.Code,
		Name:             class.Name,
		Level:            class.Level,
		IsActive:         class.IsActive,
		CreatedAt:        class.CreatedAt,
		UpdatedAt:        class.UpdatedAt,
		AcademicYearID:   class.AcademicYearID,
		AcademicYearName: academicYearName,
	}
}

func rolloverClassApplyRow(source, target db.ListSchoolClassesRow, action string) YearRolloverClassApply {
	return YearRolloverClassApply{
		SourceClassID: pgUUIDString(source.ID),
		SourceCode:    source.Code,
		SourceName:    source.Name,
		SourceLevel:   source.Level,
		TargetClassID: pgUUIDString(target.ID),
		TargetCode:    target.Code,
		TargetName:    target.Name,
		TargetLevel:   target.Level,
		Action:        action,
	}
}

func rolloverStudentWarning(student db.ListYearRolloverStudentsRow, reason string) YearRolloverStudentWarning {
	return YearRolloverStudentWarning{
		StudentID:     pgUUIDString(student.ID),
		NIS:           student.Nis,
		NISN:          student.Nisn,
		Nama:          student.Nama,
		FromClassID:   pgUUIDString(student.ClassID),
		FromClassCode: student.ClassCode,
		Reason:        reason,
	}
}

func yearRolloverApplyChallenge(sourceName, targetName string) string {
	return fmt.Sprintf("%s %s KE %s", academicYearRolloverApplyPrefix, strings.TrimSpace(sourceName), strings.TrimSpace(targetName))
}

func yearRolloverApplyAuthorized(input YearRolloverApplyInput, expectedChallenge string) bool {
	confirmation := strings.TrimSpace(input.Confirmation)
	safetyToken := strings.TrimSpace(input.SafetyToken)
	return confirmation == expectedChallenge || safetyToken == expectedChallenge
}

func classSubjectKey(classID, subjectID pgtype.UUID) string {
	return importPairKey(pgUUIDString(classID), pgUUIDString(subjectID))
}

func assignmentRefFromList(row db.ListClassSubjectAssignmentsRow) rolloverAssignmentRef {
	return rolloverAssignmentRef{
		ID:                row.ID,
		ClassID:           row.ClassID,
		SubjectID:         row.SubjectID,
		TeacherEmployeeID: row.TeacherEmployeeID,
	}
}

func assignmentRefFromModel(row db.ClassSubjectAssignment) rolloverAssignmentRef {
	return rolloverAssignmentRef{
		ID:                row.ID,
		ClassID:           row.ClassID,
		SubjectID:         row.SubjectID,
		TeacherEmployeeID: row.TeacherEmployeeID,
	}
}

func timetableSlotKey(assignmentID pgtype.UUID, dayOfWeek int16, startTime, endTime pgtype.Time, roomLabel string) string {
	return strings.Join([]string{
		pgUUIDString(assignmentID),
		strconv.Itoa(int(dayOfWeek)),
		strconv.FormatInt(startTime.Microseconds, 10),
		strconv.FormatInt(endTime.Microseconds, 10),
		normalizeLookupKey(roomLabel),
	}, "|")
}

func rolloverCopiedNotes(notes, sourceYearName string) string {
	notes = strings.TrimSpace(notes)
	prefix := "Disalin dari tahun ajaran " + strings.TrimSpace(sourceYearName)
	if notes == "" {
		return prefix
	}
	return prefix + " - " + notes
}

func nextAcademicLevel(level string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "VII", "7":
		return "VIII", true
	case "VIII", "8":
		return "IX", true
	default:
		return "", false
	}
}

func normalizeAcademicLevel(level string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "VII", "7":
		return "VII", true
	case "VIII", "8":
		return "VIII", true
	case "IX", "9":
		return "IX", true
	default:
		return "", false
	}
}

func rolloverNextLabel(value, currentLevel, nextLevel string) string {
	value = strings.TrimSpace(value)
	currentLevel, ok := normalizeAcademicLevel(currentLevel)
	if !ok {
		return nextLevel + " " + value
	}
	upper := strings.ToUpper(value)
	if strings.HasPrefix(upper, currentLevel) {
		return nextLevel + value[len(currentLevel):]
	}
	return strings.TrimSpace(nextLevel + " " + value)
}

func normalizeLookupKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func importPairKey(a, b string) string {
	return normalizeLookupKey(a) + "|" + normalizeLookupKey(b)
}

func importTripleKey(a, b, c string) string {
	return normalizeLookupKey(a) + "|" + normalizeLookupKey(b) + "|" + normalizeLookupKey(c)
}

func cell(row []string, header map[string]int, field string) string {
	idx, ok := header[normalizeLookupKey(field)]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func requireCell(errs *[]AcademicImportRowError, row int, field, value string) {
	if strings.TrimSpace(value) == "" {
		*errs = append(*errs, AcademicImportRowError{Row: row, Field: field, Message: field + " wajib diisi"})
	}
}

func csvRowBlank(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func normalizeGender(value string) string {
	switch normalizeLookupKey(value) {
	case "l", "laki-laki", "laki laki", "male":
		return "male"
	case "p", "perempuan", "female":
		return "female"
	default:
		return ""
	}
}

func parseAcademicDay(value string) (int, bool) {
	switch normalizeLookupKey(value) {
	case "1", "senin":
		return 1, true
	case "2", "selasa":
		return 2, true
	case "3", "rabu":
		return 3, true
	case "4", "kamis":
		return 4, true
	case "5", "jumat", "jum'at":
		return 5, true
	case "6", "sabtu":
		return 6, true
	default:
		return 0, false
	}
}

func normalizeImportTime(value string) string {
	parsed, err := ParseAcademicTimeInput(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return formatAcademicTime(parsed)
}

func formatAcademicTime(value pgtype.Time) string {
	if !value.Valid {
		return ""
	}
	parsed := value
	totalSeconds := parsed.Microseconds / 1_000_000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}
