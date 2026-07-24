package service

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type TimetableService struct {
	q *db.Queries
}

func NewTimetableService(q *db.Queries) *TimetableService {
	return &TimetableService{q: q}
}

// ─── Lesson Period Templates ───

type LessonPeriodItem struct {
	ID                string `json:"id"`
	AcademicYearID    string `json:"academic_year_id"`
	AcademicYearName  string `json:"academic_year_name"`
	DayOfWeek         int16  `json:"day_of_week"`
	PeriodNumber      int32  `json:"period_number"`
	StartTime         string `json:"start_time"`
	EndTime           string `json:"end_time"`
	ActivityType      string `json:"activity_type"`
	Label             string `json:"label"`
	IsCountedAsLesson bool   `json:"is_counted_as_lesson"`
}

func fmtTime(t pgtype.Time) string {
	if !t.Valid {
		return "00:00"
	}
	hours := t.Microseconds / 3600000000
	mins := (t.Microseconds % 3600000000) / 60000000
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

func (s *TimetableService) ListLessonPeriods(ctx context.Context, academicYearID string) ([]LessonPeriodItem, error) {
	rows, err := s.q.ListLessonPeriodTemplates(ctx, pgUUID(academicYearID))
	if err != nil {
		return nil, fmt.Errorf("list lesson periods: %w", err)
	}
	items := make([]LessonPeriodItem, len(rows))
	for i, r := range rows {
		items[i] = LessonPeriodItem{
			ID:                pgUUIDString(r.ID),
			AcademicYearID:    pgUUIDString(r.AcademicYearID),
			AcademicYearName:  r.AcademicYearName,
			DayOfWeek:         r.DayOfWeek,
			PeriodNumber:      r.PeriodNumber,
			StartTime:         fmtTime(r.StartTime),
			EndTime:           fmtTime(r.EndTime),
			ActivityType:      r.ActivityType,
			Label:             r.Label,
			IsCountedAsLesson: r.IsCountedAsLesson,
		}
	}
	return items, nil
}

// ─── Weekly Timetable Data ───

type TTClass struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Level string `json:"level"`
}

type TTSubject struct {
	ID                 string `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	Category           string `json:"category"`
	IsScheduleActivity bool   `json:"is_schedule_activity"`
	DefaultWeeklyHours float64 `json:"default_weekly_hours"`
	DisplayOrder       int32  `json:"display_order"`
}

type TTAssignment struct {
	ID                      string  `json:"id"`
	ClassID                 string  `json:"class_id"`
	ClassCode               string  `json:"class_code"`
	ClassName               string  `json:"class_name"`
	ClassLevel              string  `json:"class_level"`
	SubjectID               string  `json:"subject_id"`
	SubjectCode             string  `json:"subject_code"`
	SubjectName             string  `json:"subject_name"`
	SubjectCategory         string  `json:"subject_category"`
	IsScheduleActivity      bool    `json:"is_schedule_activity"`
	TeacherEmployeeID       string  `json:"teacher_employee_id"`
	TeacherName             string  `json:"teacher_name"`
	ExpectedIntraWeeklyHours float64 `json:"expected_intra_weekly_hours"`
	ExpectedTotalHours      float64 `json:"expected_total_weekly_hours"`
}

type TTSlot struct {
	ID                 string `json:"id"`
	AssignmentID       string `json:"assignment_id"`
	DayOfWeek          int16  `json:"day_of_week"`
	PeriodNumber       int32  `json:"period_number"`
	LessonPeriodLabel  string `json:"lesson_period_label"`
	SlotType           string `json:"slot_type"`
	LessonHours        float64 `json:"lesson_hours"`
	StartTime          string `json:"start_time"`
	EndTime            string `json:"end_time"`
	RoomLabel          string `json:"room_label"`
	Notes              string `json:"notes"`
	ClassID            string `json:"class_id"`
	ClassName          string `json:"class_name"`
	ClassCode          string `json:"class_code"`
	ClassLevel         string `json:"class_level"`
	SubjectID          string `json:"subject_id"`
	SubjectName        string `json:"subject_name"`
	SubjectCode        string `json:"subject_code"`
	SubjectCategory    string `json:"subject_category"`
	TeacherEmployeeID  string `json:"teacher_employee_id"`
	TeacherName        string `json:"teacher_name"`
}

type TTTeacher struct {
	ID        string `json:"id"`
	NIP       string `json:"nip"`
	Nama      string `json:"nama"`
	UnitKerja string `json:"unit_kerja"`
}

type WeeklyTimetableData struct {
	Classes     []TTClass      `json:"classes"`
	Subjects    []TTSubject    `json:"subjects"`
	Teachers    []TTTeacher    `json:"teachers"`
	Assignments []TTAssignment `json:"assignments"`
	Slots       []TTSlot       `json:"slots"`
}

func (s *TimetableService) GetWeeklyData(ctx context.Context, academicYearID string) (*WeeklyTimetableData, error) {
	ayID := pgUUID(academicYearID)

	classes, err := s.q.ListWeeklyTimetableClasses(ctx, ayID)
	if err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	subjects, err := s.q.ListWeeklyTimetableSubjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	teachers, err := s.q.ListWeeklyTimetableTeachers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	assignments, err := s.q.ListWeeklyTimetableAssignments(ctx, ayID)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	slots, err := s.q.ListWeeklyTimetableSlots(ctx, ayID)
	if err != nil {
		return nil, fmt.Errorf("list slots: %w", err)
	}

	result := &WeeklyTimetableData{
		Classes:     make([]TTClass, len(classes)),
		Subjects:    make([]TTSubject, len(subjects)),
		Teachers:    make([]TTTeacher, len(teachers)),
		Assignments: make([]TTAssignment, len(assignments)),
		Slots:       make([]TTSlot, len(slots)),
	}
	for i, c := range classes {
		result.Classes[i] = TTClass{ID: pgUUIDString(c.ID), Code: c.Code, Name: c.Name, Level: c.Level}
	}
	for i, s := range subjects {
		result.Subjects[i] = TTSubject{
			ID: pgUUIDString(s.ID), Code: s.Code, Name: s.Name, Category: s.Category,
			IsScheduleActivity: s.IsScheduleActivity,
			DefaultWeeklyHours: float64(s.DefaultWeeklyHours), DisplayOrder: s.DisplayOrder,
		}
	}
	for i, t := range teachers {
		result.Teachers[i] = TTTeacher{ID: pgUUIDString(t.ID), NIP: t.Nip, Nama: t.Nama, UnitKerja: t.UnitKerja}
	}
	for i, a := range assignments {
		result.Assignments[i] = TTAssignment{
			ID: pgUUIDString(a.ID), ClassID: pgUUIDString(a.ClassID),
			ClassCode: a.ClassCode, ClassName: a.ClassName, ClassLevel: a.ClassLevel,
			SubjectID: pgUUIDString(a.SubjectID), SubjectCode: a.SubjectCode,
			SubjectName: a.SubjectName, SubjectCategory: a.SubjectCategory,
			IsScheduleActivity: a.IsScheduleActivity,
			TeacherEmployeeID:  pgUUIDString(a.TeacherEmployeeID), TeacherName: a.TeacherName,
			ExpectedIntraWeeklyHours: a.ExpectedIntraWeeklyHours,
			ExpectedTotalHours:       a.ExpectedWeeklyHours,
		}
	}
	for i, s := range slots {
		result.Slots[i] = TTSlot{
			ID: pgUUIDString(s.ID), AssignmentID: pgUUIDString(s.AssignmentID),
			DayOfWeek: s.DayOfWeek, PeriodNumber: s.PeriodNumber,
			LessonPeriodLabel: s.LessonPeriodLabel, SlotType: s.SlotType,
			LessonHours: s.LessonHours,
			StartTime:   fmtTime(s.StartTime), EndTime: fmtTime(s.EndTime),
			RoomLabel: s.RoomLabel, Notes: s.Notes,
			ClassID: pgUUIDString(s.ClassID), ClassName: s.ClassName,
			ClassCode: s.ClassCode, ClassLevel: s.ClassLevel,
			SubjectID: pgUUIDString(s.SubjectID), SubjectName: s.SubjectName,
			SubjectCode: s.SubjectCode, SubjectCategory: s.SubjectCategory,
			TeacherEmployeeID: pgUUIDString(s.TeacherEmployeeID), TeacherName: s.TeacherName,
		}
	}
	return result, nil
}

// ─── Slot CRUD ───

type CreateSlotRequest struct {
	AssignmentID string `json:"assignment_id"`
	DayOfWeek    int16  `json:"day_of_week"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	RoomLabel    string `json:"room_label"`
	SlotType     string `json:"slot_type"`
	LessonHours  int32  `json:"lesson_hours"`
}

func toPgTime(s string) pgtype.Time {
	if len(s) < 5 {
		return pgtype.Time{Valid: false}
	}
	var hours, mins int64
	fmt.Sscanf(s, "%02d:%02d", &hours, &mins)
	return pgtype.Time{Microseconds: hours*3600000000 + mins*60000000, Valid: true}
}

func toPgNumeric(v int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(int64(v)), Exp: 0, Valid: true}
}

func (s *TimetableService) CreateSlot(ctx context.Context, req CreateSlotRequest) error {
	if req.SlotType == "" {
		req.SlotType = "pelajaran"
	}
	if req.LessonHours <= 0 {
		req.LessonHours = 1
	}
	_, err := s.q.CreateTimetableSlot(ctx, db.CreateTimetableSlotParams{
		AssignmentID: pgUUID(req.AssignmentID),
		DayOfWeek:    req.DayOfWeek,
		StartTime:    toPgTime(req.StartTime),
		EndTime:      toPgTime(req.EndTime),
		RoomLabel:    req.RoomLabel,
		Notes:        "",
		SlotType:     req.SlotType,
		LessonHours:  toPgNumeric(req.LessonHours),
	})
	if err != nil {
		return fmt.Errorf("create slot: %w", err)
	}
	return nil
}

type UpdateSlotRequest struct {
	ID           string `json:"id"`
	AssignmentID string `json:"assignment_id"`
	DayOfWeek    int16  `json:"day_of_week"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	RoomLabel    string `json:"room_label"`
	SlotType     string `json:"slot_type"`
	LessonHours  int32  `json:"lesson_hours"`
}

func (s *TimetableService) UpdateSlot(ctx context.Context, req UpdateSlotRequest) error {
	if req.SlotType == "" {
		req.SlotType = "pelajaran"
	}
	if req.LessonHours <= 0 {
		req.LessonHours = 1
	}
	_, err := s.q.UpdateTimetableSlot(ctx, db.UpdateTimetableSlotParams{
		ID:           pgUUID(req.ID),
		AssignmentID: pgUUID(req.AssignmentID),
		DayOfWeek:    req.DayOfWeek,
		StartTime:    toPgTime(req.StartTime),
		EndTime:      toPgTime(req.EndTime),
		RoomLabel:    req.RoomLabel,
		Notes:        "",
		SlotType:     req.SlotType,
		LessonHours:  toPgNumeric(req.LessonHours),
	})
	if err != nil {
		return fmt.Errorf("update slot: %w", err)
	}
	return nil
}

func (s *TimetableService) DeleteSlot(ctx context.Context, id string) error {
	return s.q.DeleteTimetableSlot(ctx, pgUUID(id))
}

func (s *TimetableService) GetActiveAcademicYear(ctx context.Context) (string, error) {
	row, err := s.q.GetActiveAcademicYear(ctx)
	if err != nil {
		return "", fmt.Errorf("get active academic year: %w", err)
	}
	return pgUUIDString(row.ID), nil
}

var _ = math.Max(0, 0) // keep import
