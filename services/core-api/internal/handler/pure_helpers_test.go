package handler

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestPureHelpersBoolExtAndRoles(t *testing.T) {
	boolCases := map[string]bool{
		"1": true, " true ": true, "YES": true, "y": true, "on": true,
		"0": false, "false": false, "": false, "maybe": false,
	}
	for input, want := range boolCases {
		if got := parseBoolFormValue(input); got != want {
			t.Fatalf("parseBoolFormValue(%q) = %v, want %v", input, got, want)
		}
	}

	extCases := map[string]string{
		".JPEG": ".jpg",
		".jpg":  ".jpg",
		".PNG":  ".png",
		"":      "",
	}
	for input, want := range extCases {
		if got := normalizedBrandingExt(input); got != want {
			t.Fatalf("normalizedBrandingExt(%q) = %q, want %q", input, got, want)
		}
	}

	validRoles := []db.CbtEventMemberRole{
		db.CbtEventMemberRolePanitia,
		db.CbtEventMemberRolePembuatSoal,
		db.CbtEventMemberRoleReviewer,
		db.CbtEventMemberRoleProktor,
		db.CbtEventMemberRolePengawas,
		db.CbtEventMemberRoleKorektor,
	}
	for _, role := range validRoles {
		if !validCbtEventMemberRole(role) {
			t.Fatalf("validCbtEventMemberRole(%q) = false, want true", role)
		}
	}
	if validCbtEventMemberRole(db.CbtEventMemberRole("observer")) {
		t.Fatal("validCbtEventMemberRole accepted an unknown role")
	}
}

func TestTargetLevelHelpersNormalizeAndValidate(t *testing.T) {
	got := normalizeTargetLevels([]string{" ix ", "VII", "", "vii", "VIII"})
	want := []string{"IX", "VII", "VIII"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeTargetLevels() = %#v, want %#v", got, want)
	}
	if got := normalizeTargetLevels(nil); len(got) != 0 {
		t.Fatalf("normalizeTargetLevels(nil) length = %d, want 0", len(got))
	}
	if err := validateTargetLevels([]string{"VII", "VIII", "IX"}); err != nil {
		t.Fatalf("validateTargetLevels(valid) returned error: %v", err)
	}
	if err := validateTargetLevels([]string{"X"}); err == nil {
		t.Fatal("validateTargetLevels accepted an invalid level")
	}
}

func TestLessonPeriodParsers(t *testing.T) {
	counted := false
	body := lessonPeriodRequest{
		AcademicYearID:    "11111111-1111-1111-1111-111111111111",
		DayOfWeek:         2,
		PeriodNumber:      3,
		StartTime:         "07:15",
		EndTime:           "08:00",
		ActivityType:      " lesson ",
		Label:             " PAI ",
		IsCountedAsLesson: &counted,
	}

	createParams, ok := parseCreateLessonPeriod(httptest.NewRecorder(), body)
	if !ok {
		t.Fatal("parseCreateLessonPeriod returned ok=false for valid input")
	}
	if createParams.ActivityType != "lesson" || createParams.Label != "PAI" || createParams.IsCountedAsLesson {
		t.Fatalf("parseCreateLessonPeriod did not trim/default fields: %#v", createParams)
	}
	if createParams.StartTime.Microseconds <= 0 || createParams.EndTime.Microseconds <= createParams.StartTime.Microseconds {
		t.Fatalf("parseCreateLessonPeriod produced unexpected times: start=%d end=%d", createParams.StartTime.Microseconds, createParams.EndTime.Microseconds)
	}

	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	updateParams, ok := parseUpdateLessonPeriod(httptest.NewRecorder(), id, body)
	if !ok {
		t.Fatal("parseUpdateLessonPeriod returned ok=false for valid input")
	}
	if updateParams.ID != id || updateParams.ActivityType != "lesson" || updateParams.Label != "PAI" {
		t.Fatalf("parseUpdateLessonPeriod produced unexpected params: %#v", updateParams)
	}

	if _, ok := parseCreateLessonPeriod(httptest.NewRecorder(), lessonPeriodRequest{AcademicYearID: "not-a-uuid"}); ok {
		t.Fatal("parseCreateLessonPeriod accepted an invalid academic year id")
	}
	invalidTime := body
	invalidTime.StartTime = "not-a-time"
	if _, _, ok := parseLessonPeriodTimes(httptest.NewRecorder(), invalidTime); ok {
		t.Fatal("parseLessonPeriodTimes accepted an invalid start time")
	}
}
