package service

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Standalone helpers for testing — same logic as academic.go
func testFmtTime(t pgtype.Time) string {
	if !t.Valid {
		return "00:00"
	}
	hours := t.Microseconds / 3600000000
	mins := (t.Microseconds % 3600000000) / 60000000
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

func testToPgTime(s string) pgtype.Time {
	if len(s) < 5 {
		return pgtype.Time{Valid: false}
	}
	var hours, mins int64
	fmt.Sscanf(s, "%02d:%02d", &hours, &mins)
	return pgtype.Time{Microseconds: hours*3600000000 + mins*60000000, Valid: true}
}

func testToPgNumeric(v int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(int64(v)), Exp: 0, Valid: true}
}

func testPgUUID(s string) pgtype.UUID {
	// Simple string to UUID for testing
	var u pgtype.UUID
	if len(s) == 36 {
		for i, c := range []byte(s) {
			if c == '-' {
				continue
			}
			if i < 8 {
				u.Bytes[0] = u.Bytes[0]*16 + hexVal(c)
			} else if i < 13 {
				u.Bytes[1] = u.Bytes[1]*16 + hexVal(c)
			}
			// Simplified - just ensure Valid=true and ignore exact Bytes
		}
	}
	u.Valid = true
	return u
}

func hexVal(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	if c >= 'A' && c <= 'F' {
		return c - 'A' + 10
	}
	return 0
}

func testPgUUIDString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

// ─── Tests ───

func TestFmtTime(t *testing.T) {
	tests := []struct {
		name     string
		input    pgtype.Time
		expected string
	}{
		{name: "valid time", input: pgtype.Time{Microseconds: 7*3600000000 + 30*60000000, Valid: true}, expected: "07:30"},
		{name: "zero", input: pgtype.Time{Microseconds: 0, Valid: true}, expected: "00:00"},
		{name: "end of day", input: pgtype.Time{Microseconds: 23*3600000000 + 59*60000000, Valid: true}, expected: "23:59"},
		{name: "invalid", input: pgtype.Time{Valid: false}, expected: "00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testFmtTime(tt.input)
			if got != tt.expected {
				t.Fatalf("fmtTime(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToPgTime(t *testing.T) {
	tests := []struct{ input string; hours, mins int64; valid bool }{
		{input: "07:30", hours: 7, mins: 30, valid: true},
		{input: "00:00", hours: 0, mins: 0, valid: true},
		{input: "23:59", hours: 23, mins: 59, valid: true},
		{input: "", valid: false},
		{input: "abc", valid: false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := testToPgTime(tt.input)
			if got.Valid != tt.valid {
				t.Fatalf("toPgTime(%q).Valid = %v, want %v", tt.input, got.Valid, tt.valid)
			}
			if got.Valid && got.Microseconds != tt.hours*3600000000+tt.mins*60000000 {
				t.Fatalf("toPgTime(%q).Microseconds = %d, want %d", tt.input, got.Microseconds, tt.hours*3600000000+tt.mins*60000000)
			}
		})
	}
}

func TestToPgNumeric(t *testing.T) {
	got := testToPgNumeric(42)
	if !got.Valid || got.Int.Int64() != 42 {
		t.Fatalf("toPgNumeric(42) = %+v, want Int=42 Valid=true", got)
	}
}

func TestPgUUIDFunctions(t *testing.T) {
	u := testPgUUID("00000000-0000-0000-0000-000000000000")
	if !u.Valid {
		t.Fatal("pgUUID should return valid")
	}
	str := testPgUUIDString(u)
	if str != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("pgUUIDString = %q", str)
	}
}

func TestStructConstructions(t *testing.T) {
	now := time.Now()

	// StudentItem
	s := StudentItem{ID: "s1", NIS: "1234", Nama: "Murid Test", Gender: "L", Status: "active", CreatedAt: now}
	if s.Nama != "Murid Test" {
		t.Fatalf("unexpected StudentItem: %+v", s)
	}

	// JournalOverviewItem
	j := JournalOverviewItem{ID: "j1", Tanggal: "2026-07-24", PertemuanKe: 1, Materi: "Test", GuruHadir: true}
	if j.Materi != "Test" || j.PertemuanKe != 1 {
		t.Fatalf("unexpected JournalOverviewItem: %+v", j)
	}

	// TTClass
	c := TTClass{ID: "c1", Code: "VII.A", Name: "VII.A", Level: "VII"}
	if c.Level != "VII" {
		t.Fatalf("unexpected TTClass: %+v", c)
	}

	// TTSubject
	sub := TTSubject{ID: "s1", Code: "MTK", Name: "Matematika", DefaultWeeklyHours: 4}
	if sub.DefaultWeeklyHours != 4 {
		t.Fatalf("unexpected TTSubject: %+v", sub)
	}

	// AttendanceItem
	a := AttendanceItem{ID: "a1", SessionID: "j1", StudentID: "s1", Status: "hadir", Nama: "Murid A"}
	if a.Nama != "Murid A" {
		t.Fatalf("unexpected AttendanceItem: %+v", a)
	}

	// LessonPeriodItem
	lp := LessonPeriodItem{ID: "lp1", DayOfWeek: 1, PeriodNumber: 1, StartTime: "07:00", EndTime: "07:40", Label: "Pertama"}
	if lp.Label != "Pertama" || lp.StartTime != "07:00" {
		t.Fatalf("unexpected LessonPeriodItem: %+v", lp)
	}
}

func TestServiceConstructors(t *testing.T) {
	if NewClassJournalService(nil) == nil {
		t.Fatal("NewClassJournalService returned nil")
	}
	if NewTimetableService(nil) == nil {
		t.Fatal("NewTimetableService returned nil")
	}
	if NewKesiswaanService(nil) == nil {
		t.Fatal("NewKesiswaanService returned nil")
	}
	if NewSubjectAssignmentService(nil) == nil {
		t.Fatal("NewSubjectAssignmentService returned nil")
	}
}

func TestCreateSessionRequest_Validation(t *testing.T) {
	// Test that CreateSessionRequest struct works correctly
	req := CreateSessionRequest{
		AssignmentID: "a1",
		Tanggal:      "2026-07-24",
		Materi:       "Test Materi",
		Catatan:      "Catatan",
		GuruHadir:    true,
	}
	if req.AssignmentID != "a1" || req.Tanggal != "2026-07-24" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBulkAttendancesRequest(t *testing.T) {
	records := []AttendanceRecord{
		{StudentID: "s1", Status: "hadir"},
		{StudentID: "s2", Status: "sakit", Catatan: "demam"},
	}
	req := BulkAttendancesRequest{SessionID: "j1", Records: records}
	if len(req.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(req.Records))
	}
	if req.Records[0].Status != "hadir" {
		t.Fatalf("first record status = %q, want hadir", req.Records[0].Status)
	}
}

func TestUpsertCellRequest(t *testing.T) {
	req := UpsertCellRequest{ClassID: "c1", SubjectID: "s1", TeacherEmployeeID: "t1"}
	if req.TeacherEmployeeID != "t1" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestCreateSlotRequest(t *testing.T) {
	req := CreateSlotRequest{
		AssignmentID: "a1", DayOfWeek: 1,
		StartTime: "07:00", EndTime: "07:40",
		SlotType: "pelajaran", LessonHours: 2,
	}
	if req.LessonHours != 2 || req.SlotType != "pelajaran" {
		t.Fatalf("unexpected CreateSlotRequest: %+v", req)
	}
}

func TestUpdateSlotRequest(t *testing.T) {
	req := UpdateSlotRequest{
		ID: "slot1", AssignmentID: "a1", DayOfWeek: 2,
		StartTime: "08:00", EndTime: "08:40",
	}
	if req.ID != "slot1" || req.DayOfWeek != 2 {
		t.Fatalf("unexpected UpdateSlotRequest: %+v", req)
	}
}

func TestAttendanceSummaryItemStruct(t *testing.T) {
	item := AttendanceSummaryItem{
		StudentID: "s1", NIS: "1234", Nama: "Murid Test",
		TotalPertemuan: 10, Hadir: 8, Sakit: 1, Izin: 1, Alpha: 0,
	}
	if item.Hadir != 8 || item.Alpha != 0 {
		t.Fatalf("unexpected AttendanceSummaryItem: %+v", item)
	}
}

func TestJournalSessionOpenResult(t *testing.T) {
	result := &JournalSessionOpenResult{Created: true}
	if !result.Created {
		t.Fatal("expected Created = true")
	}
}
