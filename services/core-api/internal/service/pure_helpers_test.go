package service

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestPusakaAttendanceTelegramPureHelpers(t *testing.T) {
	if got := defaultAttendanceTelegramSendDays(); !reflect.DeepEqual(got, []int32{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("defaultAttendanceTelegramSendDays() = %#v", got)
	}

	days, err := normalizeSendDays([]int32{6, 1, 6, 0})
	if err != nil {
		t.Fatalf("normalizeSendDays(valid) returned error: %v", err)
	}
	if want := []int32{6, 1, 0}; !reflect.DeepEqual(days, want) {
		t.Fatalf("normalizeSendDays(valid) = %#v, want %#v", days, want)
	}
	if _, err := normalizeSendDays([]int32{7}); err == nil {
		t.Fatal("normalizeSendDays accepted out-of-range day")
	}
	if !sendDayEnabled([]int32{2, 4}, 4) || sendDayEnabled([]int32{2, 4}, 5) || sendDayEnabled([]int32{9}, 9) {
		t.Fatal("sendDayEnabled returned unexpected result")
	}

	times, err := normalizeSendTimes([]string{" 7:05 ", "07:05", "18:30", ""}, "17:00")
	if err != nil {
		t.Fatalf("normalizeSendTimes(valid) returned error: %v", err)
	}
	if want := []string{"07:05", "18:30"}; !reflect.DeepEqual(times, want) {
		t.Fatalf("normalizeSendTimes(valid) = %#v, want %#v", times, want)
	}
	fallback, err := normalizeSendTimes(nil, "")
	if err != nil || !reflect.DeepEqual(fallback, []string{"17:00"}) {
		t.Fatalf("normalizeSendTimes fallback = %#v, %v; want [17:00], nil", fallback, err)
	}
	if _, err := normalizeSendTimes([]string{"24:00"}, "17:00"); err == nil {
		t.Fatal("normalizeSendTimes accepted invalid clock time")
	}
	if _, err := normalizeSendTimes([]string{"00:00", "01:00", "02:00", "03:00", "04:00", "05:00", "06:00", "07:00", "08:00"}, "17:00"); err == nil {
		t.Fatal("normalizeSendTimes accepted more than 8 values")
	}

	micros, err := parseHHMMToMicros("01:30")
	if err != nil || micros != int64(90*60*1_000_000) {
		t.Fatalf("parseHHMMToMicros(01:30) = %d, %v", micros, err)
	}
	if got := microsToHHMM(micros); got != "01:30" {
		t.Fatalf("microsToHHMM(%d) = %q, want 01:30", micros, got)
	}
}

func TestCbtApprovalAndQuestionRequirementValidators(t *testing.T) {
	validID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	in := normalizeCbtApprovalInput(SaveCbtApprovalInput{
		EntityType:   " event ",
		EntityID:     validID,
		ApprovalType: " package_ready ",
		Notes:        " siap ",
		ActorUserID:  validID,
	})
	if in.EntityType != "event" || in.ApprovalType != "package_ready" || in.Notes != "siap" {
		t.Fatalf("normalizeCbtApprovalInput() = %#v", in)
	}
	if err := validateCbtApprovalInput(in); err != nil {
		t.Fatalf("validateCbtApprovalInput(valid) returned error: %v", err)
	}

	invalidCases := []SaveCbtApprovalInput{
		{EntityType: "bad", EntityID: validID, ApprovalType: "package_ready", ActorUserID: validID},
		{EntityType: "event", ApprovalType: "package_ready", ActorUserID: validID},
		{EntityType: "event", EntityID: validID, ApprovalType: "bad", ActorUserID: validID},
		{EntityType: "event", EntityID: validID, ApprovalType: "package_ready"},
	}
	for _, tc := range invalidCases {
		if err := validateCbtApprovalInput(tc); err == nil {
			t.Fatalf("validateCbtApprovalInput(%#v) returned nil error", tc)
		}
	}

	for _, entityType := range []string{"event", "session", "package", "result"} {
		if !validCbtApprovalEntityType(entityType) {
			t.Fatalf("validCbtApprovalEntityType(%q) = false", entityType)
		}
	}
	if validCbtApprovalEntityType("question") {
		t.Fatal("validCbtApprovalEntityType accepted invalid value")
	}
	for _, approvalType := range []string{"package_ready", "participants_rooms_ready", "tokens_cards_ready", "results_verified", "final_archive", "session_minutes", "room_handover"} {
		if !validCbtApprovalType(approvalType) {
			t.Fatalf("validCbtApprovalType(%q) = false", approvalType)
		}
	}
	if validCbtApprovalType("unexpected") {
		t.Fatal("validCbtApprovalType accepted invalid value")
	}

	for _, mode := range []string{"per_rombel", "per_level", "pool_level_subject"} {
		if !validCbtQuestionRequirementScopeMode(mode) {
			t.Fatalf("validCbtQuestionRequirementScopeMode(%q) = false", mode)
		}
	}
	if validCbtQuestionRequirementScopeMode("school") {
		t.Fatal("validCbtQuestionRequirementScopeMode accepted invalid value")
	}
	for _, filter := range []string{"published_only", "all_progress"} {
		if !validCbtQuestionRequirementStatusFilter(filter) {
			t.Fatalf("validCbtQuestionRequirementStatusFilter(%q) = false", filter)
		}
	}
	if validCbtQuestionRequirementStatusFilter("draft_only") {
		t.Fatal("validCbtQuestionRequirementStatusFilter accepted invalid value")
	}
}
