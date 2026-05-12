package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type academicStore interface {
	ListAcademicYears(ctx context.Context) ([]db.AcademicYear, error)
	GetAcademicYearByID(ctx context.Context, id pgtype.UUID) (db.AcademicYear, error)
	CountAcademicYearNameConflicts(ctx context.Context, name string) (int32, error)
	DeactivateAcademicYears(ctx context.Context) error
	ActivateAcademicYear(ctx context.Context, id pgtype.UUID) (db.AcademicYear, error)
	ListSchoolClasses(ctx context.Context) ([]db.ListSchoolClassesRow, error)
	ListSubjects(ctx context.Context) ([]db.ListSubjectsRow, error)
	ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error)
	ListTimetableSlots(ctx context.Context) ([]db.ListTimetableSlotsRow, error)
	ListYearRolloverStudents(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListYearRolloverStudentsRow, error)
	ListYearRolloverHomeroomAssignments(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListYearRolloverHomeroomAssignmentsRow, error)
	ListYearRolloverHomeroomAssignmentDetails(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListYearRolloverHomeroomAssignmentDetailsRow, error)
	CountActiveHomeroomAssignmentByClass(ctx context.Context, classID pgtype.UUID) (int32, error)
	ListAcademicImportStudents(ctx context.Context) ([]db.ListAcademicImportStudentsRow, error)
	ListAcademicImportTeachers(ctx context.Context) ([]db.ListAcademicImportTeachersRow, error)
	GetActiveAcademicYear(ctx context.Context) (db.AcademicYear, error)
	ListWeeklyTimetableClasses(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListWeeklyTimetableClassesRow, error)
	ListWeeklyTimetableTeachers(ctx context.Context) ([]db.ListWeeklyTimetableTeachersRow, error)
	ListWeeklyTimetableSubjects(ctx context.Context) ([]db.ListWeeklyTimetableSubjectsRow, error)
	ListWeeklyTimetableAssignments(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListWeeklyTimetableAssignmentsRow, error)
	ListWeeklyTimetableSlots(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListWeeklyTimetableSlotsRow, error)
	ListTimetableConflicts(ctx context.Context, academicYearID pgtype.UUID) ([]db.ListTimetableConflictsRow, error)
	GetAcademicStats(ctx context.Context) (db.GetAcademicStatsRow, error)
	GetAcademicDashboardSummary(ctx context.Context) (db.GetAcademicDashboardSummaryRow, error)
	ListCurriculumProfiles(ctx context.Context) ([]db.CurriculumProfile, error)
	GetActiveCurriculumProfile(ctx context.Context) (db.CurriculumProfile, error)
	ListCurriculumSubjectAllocations(ctx context.Context, arg db.ListCurriculumSubjectAllocationsParams) ([]db.ListCurriculumSubjectAllocationsRow, error)
	GetCurriculumSummaryByLevel(ctx context.Context, curriculumProfileID pgtype.UUID) ([]db.GetCurriculumSummaryByLevelRow, error)
	ListClassCurriculumAssignments(ctx context.Context, curriculumProfileID pgtype.UUID) ([]db.ListClassCurriculumAssignmentsRow, error)
	GetSubject(ctx context.Context, id pgtype.UUID) (db.GetSubjectRow, error)
	CountSubjectCodeConflicts(ctx context.Context, arg db.CountSubjectCodeConflictsParams) (int32, error)
	CreateAcademicYear(ctx context.Context, arg db.CreateAcademicYearParams) (db.AcademicYear, error)
	CreateSchoolClass(ctx context.Context, arg db.CreateSchoolClassParams) (db.SchoolClass, error)
	CreateHomeroomAssignment(ctx context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error)
	CreateSubject(ctx context.Context, arg db.CreateSubjectParams) (db.Subject, error)
	UpdateSubject(ctx context.Context, arg db.UpdateSubjectParams) (db.Subject, error)
	CreateClassSubjectAssignment(ctx context.Context, arg db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error)
	CreateTimetableSlot(ctx context.Context, arg db.CreateTimetableSlotParams) (db.TimetableSlot, error)
	PromoteYearRolloverStudent(ctx context.Context, arg db.PromoteYearRolloverStudentParams) (int64, error)
	GetTimetableSlot(ctx context.Context, id pgtype.UUID) (db.TimetableSlot, error)
	UpdateTimetableSlot(ctx context.Context, arg db.UpdateTimetableSlotParams) (db.TimetableSlot, error)
	DeleteAcademicYear(ctx context.Context, id pgtype.UUID) error
	DeleteSchoolClass(ctx context.Context, id pgtype.UUID) error
	DeleteSubject(ctx context.Context, id pgtype.UUID) error
	DeleteClassSubjectAssignment(ctx context.Context, id pgtype.UUID) error
	DeleteTimetableSlot(ctx context.Context, id pgtype.UUID) error
	GetClassSubjectAssignment(ctx context.Context, id pgtype.UUID) (db.GetClassSubjectAssignmentRow, error)
	LockTimetableMutationScope(ctx context.Context, lockKey string) (int64, error)
	CountTimetableConflicts(ctx context.Context, arg db.CountTimetableConflictsParams) (int32, error)
	CountTimetableRoomConflicts(ctx context.Context, arg db.CountTimetableRoomConflictsParams) (int32, error)
}

type Academic struct {
	q  academicStore
	tx classJournalTxStarter
}

type WeeklyTimetable struct {
	ActiveAcademicYearID   pgtype.UUID                            `json:"active_academic_year_id"`
	ActiveAcademicYearName string                                 `json:"active_academic_year_name"`
	Classes                []db.ListWeeklyTimetableClassesRow     `json:"classes"`
	Teachers               []db.ListWeeklyTimetableTeachersRow    `json:"teachers"`
	Subjects               []db.ListWeeklyTimetableSubjectsRow    `json:"subjects"`
	Assignments            []db.ListWeeklyTimetableAssignmentsRow `json:"assignments"`
	Slots                  []WeeklyTimetableSlot                  `json:"slots"`
	Conflicts              []db.ListTimetableConflictsRow         `json:"conflicts"`
}

type WeeklyTimetableSlot struct {
	db.ListWeeklyTimetableSlotsRow
	ConflictStatus string `json:"conflict_status"`
	ConflictLabel  string `json:"conflict_label"`
	ConflictCount  int    `json:"conflict_count"`
}

type CurriculumOverview struct {
	Profiles         []db.CurriculumProfile                 `json:"profiles"`
	ActiveProfile    *db.CurriculumProfile                  `json:"active_profile"`
	Allocations      []CurriculumAllocation                 `json:"allocations"`
	SummaryByLevel   []CurriculumLevelSummary               `json:"summary_by_level"`
	ClassAssignments []db.ListClassCurriculumAssignmentsRow `json:"class_assignments"`
}

type CurriculumAllocation struct {
	db.ListCurriculumSubjectAllocationsRow
	IntraWeeklyHours float64 `json:"intra_weekly_hours"`
	KokuWeeklyHours  float64 `json:"koku_weekly_hours"`
	TotalWeeklyHours float64 `json:"total_weekly_hours"`
}

type CurriculumLevelSummary struct {
	db.GetCurriculumSummaryByLevelRow
	IntraWeeklyHours float64 `json:"intra_weekly_hours"`
	KokuWeeklyHours  float64 `json:"koku_weekly_hours"`
	TotalWeeklyHours float64 `json:"total_weekly_hours"`
	StatusLabel      string  `json:"status_label"`
}

func NewAcademic(q *db.Queries) *Academic { return &Academic{q: q} }

func NewAcademicWithPool(pool *pgxpool.Pool) *Academic {
	if pool == nil {
		return &Academic{q: db.New(nil)}
	}
	return &Academic{q: db.New(pool), tx: pool}
}

func (s *Academic) ListYears(ctx context.Context) ([]db.AcademicYear, error) {
	return s.q.ListAcademicYears(ctx)
}

func (s *Academic) ListClasses(ctx context.Context) ([]db.ListSchoolClassesRow, error) {
	return s.q.ListSchoolClasses(ctx)
}

func (s *Academic) ListSubjects(ctx context.Context) ([]db.ListSubjectsRow, error) {
	return s.q.ListSubjects(ctx)
}

func (s *Academic) ListAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	return s.q.ListClassSubjectAssignments(ctx)
}

func (s *Academic) ListTimetableSlots(ctx context.Context) ([]db.ListTimetableSlotsRow, error) {
	return s.q.ListTimetableSlots(ctx)
}

func (s *Academic) GetWeeklyTimetable(ctx context.Context) (WeeklyTimetable, error) {
	year, err := s.q.GetActiveAcademicYear(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return emptyWeeklyTimetable(), nil
		}
		return WeeklyTimetable{}, err
	}
	classes, err := s.q.ListWeeklyTimetableClasses(ctx, year.ID)
	if err != nil {
		return WeeklyTimetable{}, err
	}
	teachers, err := s.q.ListWeeklyTimetableTeachers(ctx)
	if err != nil {
		return WeeklyTimetable{}, err
	}
	subjects, err := s.q.ListWeeklyTimetableSubjects(ctx)
	if err != nil {
		return WeeklyTimetable{}, err
	}
	assignments, err := s.q.ListWeeklyTimetableAssignments(ctx, year.ID)
	if err != nil {
		return WeeklyTimetable{}, err
	}
	slots, err := s.q.ListWeeklyTimetableSlots(ctx, year.ID)
	if err != nil {
		return WeeklyTimetable{}, err
	}
	conflicts, err := s.q.ListTimetableConflicts(ctx, year.ID)
	if err != nil {
		return WeeklyTimetable{}, err
	}
	return WeeklyTimetable{
		ActiveAcademicYearID:   year.ID,
		ActiveAcademicYearName: year.Name,
		Classes:                classes,
		Teachers:               teachers,
		Subjects:               subjects,
		Assignments:            assignments,
		Slots:                  annotateWeeklyTimetableSlots(slots, conflicts),
		Conflicts:              conflicts,
	}, nil
}

func (s *Academic) GetTimetableConflicts(ctx context.Context) ([]db.ListTimetableConflictsRow, error) {
	year, err := s.q.GetActiveAcademicYear(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []db.ListTimetableConflictsRow{}, nil
		}
		return nil, err
	}
	return s.q.ListTimetableConflicts(ctx, year.ID)
}

func (s *Academic) GetStats(ctx context.Context) (db.GetAcademicStatsRow, error) {
	return s.q.GetAcademicStats(ctx)
}

func (s *Academic) GetDashboardSummary(ctx context.Context) (db.GetAcademicDashboardSummaryRow, error) {
	return s.q.GetAcademicDashboardSummary(ctx)
}

func (s *Academic) ListCurriculumProfiles(ctx context.Context) ([]db.CurriculumProfile, error) {
	return s.q.ListCurriculumProfiles(ctx)
}

func (s *Academic) GetActiveCurriculumProfile(ctx context.Context) (db.CurriculumProfile, error) {
	return s.q.GetActiveCurriculumProfile(ctx)
}

func (s *Academic) ListCurriculumAllocations(ctx context.Context, profileID pgtype.UUID, level string) ([]CurriculumAllocation, error) {
	levelArg := pgtype.Text{}
	if strings.TrimSpace(level) != "" {
		levelArg = pgtype.Text{String: strings.ToUpper(strings.TrimSpace(level)), Valid: true}
	}
	rows, err := s.q.ListCurriculumSubjectAllocations(ctx, db.ListCurriculumSubjectAllocationsParams{CurriculumProfileID: profileID, Level: levelArg})
	if err != nil {
		return nil, err
	}
	out := make([]CurriculumAllocation, 0, len(rows))
	for _, row := range rows {
		out = append(out, CurriculumAllocation{
			ListCurriculumSubjectAllocationsRow: row,
			IntraWeeklyHours:                    numericToFloat64(row.IntraWeeklyHours),
			KokuWeeklyHours:                     numericToFloat64(row.KokuWeeklyHours),
			TotalWeeklyHours:                    numericToFloat64(row.TotalWeeklyHours),
		})
	}
	return out, nil
}

func (s *Academic) GetCurriculumSummary(ctx context.Context, profileID pgtype.UUID) ([]CurriculumLevelSummary, error) {
	rows, err := s.q.GetCurriculumSummaryByLevel(ctx, profileID)
	if err != nil {
		return nil, err
	}
	out := make([]CurriculumLevelSummary, 0, len(rows))
	for _, row := range rows {
		label := "Perlu ditinjau"
		if row.ComplianceStatus == "sesuai" {
			label = "Sesuai KMA"
		}
		out = append(out, CurriculumLevelSummary{
			GetCurriculumSummaryByLevelRow: row,
			IntraWeeklyHours:               numericToFloat64(row.IntraWeeklyHours),
			KokuWeeklyHours:                numericToFloat64(row.KokuWeeklyHours),
			TotalWeeklyHours:               numericToFloat64(row.TotalWeeklyHours),
			StatusLabel:                    label,
		})
	}
	return out, nil
}

func (s *Academic) ListClassCurriculumAssignments(ctx context.Context, profileID pgtype.UUID) ([]db.ListClassCurriculumAssignmentsRow, error) {
	return s.q.ListClassCurriculumAssignments(ctx, profileID)
}

func (s *Academic) GetCurriculumOverview(ctx context.Context, profileID pgtype.UUID, level string) (CurriculumOverview, error) {
	profiles, err := s.q.ListCurriculumProfiles(ctx)
	if err != nil {
		return CurriculumOverview{}, err
	}
	selectedID := profileID
	var active *db.CurriculumProfile
	if !selectedID.Valid {
		profile, err := s.q.GetActiveCurriculumProfile(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return CurriculumOverview{Profiles: profiles, Allocations: []CurriculumAllocation{}, SummaryByLevel: []CurriculumLevelSummary{}, ClassAssignments: []db.ListClassCurriculumAssignmentsRow{}}, nil
			}
			return CurriculumOverview{}, err
		}
		selectedID = profile.ID
		active = &profile
	} else {
		for _, profile := range profiles {
			if profile.ID == selectedID {
				copyProfile := profile
				active = &copyProfile
				break
			}
		}
	}
	allocations, err := s.ListCurriculumAllocations(ctx, selectedID, level)
	if err != nil {
		return CurriculumOverview{}, err
	}
	summary, err := s.GetCurriculumSummary(ctx, selectedID)
	if err != nil {
		return CurriculumOverview{}, err
	}
	classes, err := s.q.ListClassCurriculumAssignments(ctx, selectedID)
	if err != nil {
		return CurriculumOverview{}, err
	}
	return CurriculumOverview{Profiles: profiles, ActiveProfile: active, Allocations: allocations, SummaryByLevel: summary, ClassAssignments: classes}, nil
}

func numericToFloat64(value pgtype.Numeric) float64 {
	if !value.Valid || value.Int == nil {
		return 0
	}
	rat := new(big.Rat).SetInt(value.Int)
	if value.Exp > 0 {
		rat.Mul(rat, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(value.Exp)), nil)))
	} else if value.Exp < 0 {
		rat.Quo(rat, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-value.Exp)), nil)))
	}
	floatValue, _ := rat.Float64()
	return math.Round(floatValue*100) / 100
}

func emptyWeeklyTimetable() WeeklyTimetable {
	return WeeklyTimetable{
		Classes:     []db.ListWeeklyTimetableClassesRow{},
		Teachers:    []db.ListWeeklyTimetableTeachersRow{},
		Subjects:    []db.ListWeeklyTimetableSubjectsRow{},
		Assignments: []db.ListWeeklyTimetableAssignmentsRow{},
		Slots:       []WeeklyTimetableSlot{},
		Conflicts:   []db.ListTimetableConflictsRow{},
	}
}

func annotateWeeklyTimetableSlots(slots []db.ListWeeklyTimetableSlotsRow, conflicts []db.ListTimetableConflictsRow) []WeeklyTimetableSlot {
	countBySlot := make(map[string]int, len(slots))
	invalidBySlot := make(map[string]bool, len(slots))
	for _, conflict := range conflicts {
		slotKey := conflict.SlotID.String()
		if slotKey != "" {
			countBySlot[slotKey]++
			if conflict.ConflictType == "invalid_time_range" {
				invalidBySlot[slotKey] = true
			}
		}
		if conflict.RelatedSlotID.Valid {
			relatedKey := conflict.RelatedSlotID.String()
			countBySlot[relatedKey]++
		}
	}
	out := make([]WeeklyTimetableSlot, 0, len(slots))
	for _, slot := range slots {
		key := slot.ID.String()
		status := "ok"
		label := "Aman"
		if invalidBySlot[key] {
			status = "invalid_time_range"
			label = "Rentang waktu tidak valid"
		} else if countBySlot[key] > 0 {
			status = "conflict"
			label = "Bentrok"
		}
		out = append(out, WeeklyTimetableSlot{
			ListWeeklyTimetableSlotsRow: slot,
			ConflictStatus:              status,
			ConflictLabel:               label,
			ConflictCount:               countBySlot[key],
		})
	}
	return out
}

func (s *Academic) CreateYear(ctx context.Context, p db.CreateAcademicYearParams) (db.AcademicYear, error) {
	normalized, err := normalizeAcademicYearParams(p)
	if err != nil {
		return db.AcademicYear{}, err
	}
	p = normalized
	conflicts, err := s.q.CountAcademicYearNameConflicts(ctx, p.Name)
	if err != nil {
		return db.AcademicYear{}, err
	}
	if conflicts > 0 {
		return db.AcademicYear{}, fmt.Errorf("%w: tahun ajaran sudah ada", domain.ErrConflict)
	}
	return s.q.CreateAcademicYear(ctx, p)
}

func (s *Academic) CreateClass(ctx context.Context, p db.CreateSchoolClassParams) (db.SchoolClass, error) {
	return s.q.CreateSchoolClass(ctx, p)
}

func (s *Academic) CreateSubject(ctx context.Context, p db.CreateSubjectParams) (db.Subject, error) {
	normalized, err := normalizeCreateSubjectParams(p)
	if err != nil {
		return db.Subject{}, err
	}
	p = normalized
	return s.q.CreateSubject(ctx, p)
}

func (s *Academic) UpdateSubject(ctx context.Context, p db.UpdateSubjectParams) (db.Subject, error) {
	normalized, err := normalizeUpdateSubjectParams(p)
	if err != nil {
		return db.Subject{}, err
	}
	p = normalized
	if _, err := s.q.GetSubject(ctx, p.ID); err != nil {
		return db.Subject{}, err
	}
	conflicts, err := s.q.CountSubjectCodeConflicts(ctx, db.CountSubjectCodeConflictsParams{
		ID:   p.ID,
		Code: p.Code,
	})
	if err != nil {
		return db.Subject{}, err
	}
	if conflicts > 0 {
		return db.Subject{}, fmt.Errorf("%w: kode mapel sudah dipakai", domain.ErrConflict)
	}
	return s.q.UpdateSubject(ctx, p)
}

func (s *Academic) CreateAssignment(ctx context.Context, p db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error) {
	return s.q.CreateClassSubjectAssignment(ctx, p)
}

func (s *Academic) CreateTimetableSlot(ctx context.Context, p db.CreateTimetableSlotParams) (db.TimetableSlot, error) {
	var row db.TimetableSlot
	err := s.withAcademicStore(ctx, func(store academicStore) error {
		if err := s.ensureTimetableSlotAvailable(ctx, store, p.AssignmentID, p.DayOfWeek, p.StartTime, p.EndTime, p.RoomLabel, pgtype.UUID{}); err != nil {
			return err
		}
		var err error
		row, err = store.CreateTimetableSlot(ctx, p)
		return err
	})
	return row, err
}

func (s *Academic) UpdateTimetableSlot(ctx context.Context, p db.UpdateTimetableSlotParams) (db.TimetableSlot, error) {
	var row db.TimetableSlot
	err := s.withAcademicStore(ctx, func(store academicStore) error {
		if _, err := store.GetTimetableSlot(ctx, p.ID); err != nil {
			return err
		}
		if err := s.ensureTimetableSlotAvailable(ctx, store, p.AssignmentID, p.DayOfWeek, p.StartTime, p.EndTime, p.RoomLabel, p.ID); err != nil {
			return err
		}
		var err error
		row, err = store.UpdateTimetableSlot(ctx, p)
		return err
	})
	return row, err
}

func (s *Academic) DeleteYear(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteAcademicYear(ctx, id)
}

func (s *Academic) DeleteClass(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteSchoolClass(ctx, id)
}

func (s *Academic) DeleteSubject(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteSubject(ctx, id)
}

func normalizeCreateSubjectParams(p db.CreateSubjectParams) (db.CreateSubjectParams, error) {
	code, name, category, err := normalizeSubjectIdentity(p.Code, p.Name, p.Category)
	if err != nil {
		return db.CreateSubjectParams{}, err
	}
	if err := validateSubjectNumbers(p.DefaultWeeklyHours, p.DisplayOrder); err != nil {
		return db.CreateSubjectParams{}, err
	}
	p.Code = code
	p.Name = name
	p.Category = category
	return p, nil
}

func normalizeUpdateSubjectParams(p db.UpdateSubjectParams) (db.UpdateSubjectParams, error) {
	code, name, category, err := normalizeSubjectIdentity(p.Code, p.Name, p.Category)
	if err != nil {
		return db.UpdateSubjectParams{}, err
	}
	if err := validateSubjectNumbers(p.DefaultWeeklyHours, p.DisplayOrder); err != nil {
		return db.UpdateSubjectParams{}, err
	}
	p.Code = code
	p.Name = name
	p.Category = category
	return p, nil
}

func normalizeSubjectIdentity(code, name, category string) (string, string, string, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	category = strings.TrimSpace(category)
	if category == "" {
		category = "intrakurikuler"
	}
	if code == "" {
		return "", "", "", fmt.Errorf("%w: kode mapel wajib diisi", domain.ErrBadRequest)
	}
	if name == "" {
		return "", "", "", fmt.Errorf("%w: nama mapel wajib diisi", domain.ErrBadRequest)
	}
	return code, name, category, nil
}

func validateSubjectNumbers(defaultWeeklyHours, displayOrder int32) error {
	if defaultWeeklyHours < 0 || defaultWeeklyHours > 60 {
		return fmt.Errorf("%w: jam pelajaran per pekan harus 0-60", domain.ErrBadRequest)
	}
	if displayOrder < 0 {
		return fmt.Errorf("%w: urutan tampil tidak boleh negatif", domain.ErrBadRequest)
	}
	return nil
}

func (s *Academic) DeleteAssignment(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteClassSubjectAssignment(ctx, id)
}

func (s *Academic) DeleteTimetableSlot(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteTimetableSlot(ctx, id)
}

func (s *Academic) ensureTimetableSlotAvailable(ctx context.Context, store academicStore, assignmentID pgtype.UUID, dayOfWeek int16, startTime, endTime pgtype.Time, roomLabel string, excludeSlotID pgtype.UUID) error {
	assignment, err := store.GetClassSubjectAssignment(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("penugasan guru mapel tidak ditemukan")
	}
	if err := lockTimetableMutationScopes(ctx, store, dayOfWeek, assignment.ClassID, assignment.TeacherEmployeeID, roomLabel); err != nil {
		return err
	}
	conflicts, err := store.CountTimetableConflicts(ctx, db.CountTimetableConflictsParams{
		DayOfWeek:         dayOfWeek,
		StartTime:         startTime,
		EndTime:           endTime,
		ClassID:           assignment.ClassID,
		TeacherEmployeeID: assignment.TeacherEmployeeID,
		ExcludeSlotID:     excludeSlotID,
	})
	if err != nil {
		return err
	}
	if conflicts > 0 {
		return fmt.Errorf("jam pelajaran bentrok dengan jadwal rombel atau guru pada waktu yang sama")
	}
	roomConflicts, err := store.CountTimetableRoomConflicts(ctx, db.CountTimetableRoomConflictsParams{
		DayOfWeek:     dayOfWeek,
		StartTime:     startTime,
		EndTime:       endTime,
		RoomLabel:     strings.TrimSpace(roomLabel),
		ExcludeSlotID: excludeSlotID,
	})
	if err != nil {
		return err
	}
	if roomConflicts > 0 {
		return fmt.Errorf("jam pelajaran bentrok dengan penggunaan ruang pada waktu yang sama")
	}
	return nil
}

func (s *Academic) withAcademicStore(ctx context.Context, fn func(academicStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func ParseAcademicTimeInput(value string) (pgtype.Time, error) {
	var out pgtype.Time
	raw := strings.TrimSpace(value)
	if raw == "" {
		return out, fmt.Errorf("waktu wajib diisi")
	}
	if len(raw) == 5 {
		raw += ":00"
	}
	if err := out.Scan(raw); err != nil {
		return out, fmt.Errorf("format waktu tidak valid")
	}
	return out, nil
}
