package service

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestExamDrawQuestionGroup(t *testing.T) {
	questions := []db.GetExamQuestionsRow{
		{ID: mustUUID(t, "10000000-0000-0000-0000-000000000001"), Code: "Q1"},
		{ID: mustUUID(t, "10000000-0000-0000-0000-000000000002"), Code: "Q2"},
		{ID: mustUUID(t, "10000000-0000-0000-0000-000000000003"), Code: "Q3"},
	}

	for _, drawCount := range []int32{0, -1, 3, 4} {
		t.Run("returns original when count out of draw range", func(t *testing.T) {
			got := drawQuestionGroup(questions, drawCount)
			if !reflect.DeepEqual(got, questions) {
				t.Fatalf("drawQuestionGroup(..., %d) = %#v, want original %#v", drawCount, got, questions)
			}
		})
	}

	got := drawQuestionGroup(questions, 2)
	if len(got) != 2 {
		t.Fatalf("drawQuestionGroup(..., 2) len = %d, want 2", len(got))
	}
	seen := map[string]bool{}
	allowed := map[string]bool{}
	for _, question := range questions {
		allowed[pgUUIDString(question.ID)] = true
	}
	for _, question := range got {
		id := pgUUIDString(question.ID)
		if !allowed[id] {
			t.Fatalf("drawQuestionGroup returned unknown question ID %q", id)
		}
		if seen[id] {
			t.Fatalf("drawQuestionGroup returned duplicate question ID %q", id)
		}
		seen[id] = true
	}
}

func TestExamOptionRandomizationHelpers(t *testing.T) {
	for _, tt := range []struct {
		questionType string
		want         bool
	}{
		{questionType: "multiple_choice", want: true},
		{questionType: "multiple_answer", want: true},
		{questionType: "essay", want: false},
		{questionType: "true_false", want: false},
		{questionType: "", want: false},
	} {
		t.Run(tt.questionType, func(t *testing.T) {
			if got := supportsOptionRandomization(tt.questionType); got != tt.want {
				t.Fatalf("supportsOptionRandomization(%q) = %v, want %v", tt.questionType, got, tt.want)
			}
		})
	}

	row := db.GetExamQuestionsRow{
		QuestionType: "multiple_choice",
		Options:      marshalJSON([]QuestionOption{{Label: " a ", Text: "Alpha"}, {Label: "b", Text: "Beta"}, {Label: " ", Text: "Blank"}}),
		OptionA:      "Legacy A",
		OptionB:      "Legacy B",
	}
	if got, want := optionLabelsForQuestion(row), []string{"A", "B"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("optionLabelsForQuestion(JSON) = %#v, want %#v", got, want)
	}

	legacyRow := db.GetExamQuestionsRow{QuestionType: "multiple_choice", OptionA: "Alpha", OptionB: "Beta", OptionD: "Delta"}
	if got, want := optionLabelsForQuestion(legacyRow), []string{"A", "B", "D"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("optionLabelsForQuestion(legacy) = %#v, want %#v", got, want)
	}

	for _, tt := range []struct {
		idx  int
		want string
	}{
		{idx: -1, want: ""},
		{idx: 0, want: "A"},
		{idx: 25, want: "Z"},
		{idx: 26, want: ""},
	} {
		if got := labelFromIndex(tt.idx); got != tt.want {
			t.Fatalf("labelFromIndex(%d) = %q, want %q", tt.idx, got, tt.want)
		}
	}
}

func TestExamParticipantCommandFromEvent(t *testing.T) {
	createdAt := time.Date(2026, 5, 17, 12, 30, 0, 0, time.UTC)
	row := db.ListPendingParticipantCommandsRow{
		ID:        mustUUID(t, "20000000-0000-0000-0000-000000000001"),
		EventType: "proctor_command",
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
		EventData: marshalJSON(map[string]any{
			"command_type": " reconnect ",
			"message":      " Please reconnect ",
			"severity":     " warning ",
			"actor":        " proctor-1 ",
		}),
	}

	got := participantCommandFromEvent(row)
	if got.ID != "20000000-0000-0000-0000-000000000001" || got.Type != "reconnect" || got.Message != "Please reconnect" {
		t.Fatalf("participantCommandFromEvent() core fields = %#v", got)
	}
	if got.Severity != "warning" || got.Actor != "proctor-1" || got.IssuedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("participantCommandFromEvent() metadata = %#v", got)
	}
	if got.RawEvent != "proctor_command" || got.LegacyLabel != "Login ulang" || got.Payload == nil {
		t.Fatalf("participantCommandFromEvent() event fields = %#v", got)
	}

	fallback := participantCommandFromEvent(db.ListPendingParticipantCommandsRow{EventData: []byte(`{"type":"unlock_notice"}`)})
	if fallback.Type != "unlock_notice" || fallback.LegacyLabel != "Akses dibuka" {
		t.Fatalf("participantCommandFromEvent(type fallback) = %#v", fallback)
	}

	invalid := participantCommandFromEvent(db.ListPendingParticipantCommandsRow{EventData: []byte(`not-json`)})
	if invalid.Type != "warning_message" || invalid.LegacyLabel != "Peringatan" || invalid.Payload != nil {
		t.Fatalf("participantCommandFromEvent(invalid JSON) = %#v", invalid)
	}
}

func TestExamStringFromMap(t *testing.T) {
	values := map[string]any{"message": "  hello  ", "count": 3, "nil": nil}
	if got := stringFromMap(values, "message"); got != "hello" {
		t.Fatalf("stringFromMap(message) = %q, want hello", got)
	}
	for _, key := range []string{"count", "nil", "missing"} {
		if got := stringFromMap(values, key); got != "" {
			t.Fatalf("stringFromMap(%q) = %q, want empty", key, got)
		}
	}
	if got := stringFromMap(nil, "message"); got != "" {
		t.Fatalf("stringFromMap(nil) = %q, want empty", got)
	}
}

func TestCbtSessionParticipantCommandHelpers(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  string
	}{
		{input: " reviewed ", want: "reviewed"},
		{input: "CLEARED", want: "cleared"},
		{input: "warning_given", want: "warning_given"},
		{input: "locked", want: "locked"},
		{input: "submitted", want: "submitted"},
		{input: "escalated", want: "escalated"},
		{input: "ignored", want: ""},
	} {
		if got := normalizeIncidentAction(tt.input); got != tt.want {
			t.Fatalf("normalizeIncidentAction(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	for _, tt := range []struct {
		input    string
		wantType string
		wantMsg  string
		wantSev  string
	}{
		{input: " WARNING_MESSAGE ", wantType: ParticipantCommandWarningMessage, wantMsg: "Tetap di aplikasi ujian dan ikuti arahan pengawas.", wantSev: "warning"},
		{input: "reconnect", wantType: ParticipantCommandReconnect, wantMsg: "Silakan hubungi pengawas untuk login ulang setelah akses diverifikasi.", wantSev: "warning"},
		{input: "unlock_notice", wantType: ParticipantCommandUnlockNotice, wantMsg: "Akses ujian sudah dibuka. Lanjutkan hanya setelah pengawas memberi arahan.", wantSev: "info"},
		{input: "unknown", wantType: "", wantMsg: "Tetap di aplikasi ujian dan ikuti arahan pengawas.", wantSev: "warning"},
	} {
		gotType := normalizeParticipantCommand(tt.input)
		if gotType != tt.wantType {
			t.Fatalf("normalizeParticipantCommand(%q) = %q, want %q", tt.input, gotType, tt.wantType)
		}
		if gotMsg := defaultParticipantCommandMessage(gotType); gotMsg != tt.wantMsg {
			t.Fatalf("defaultParticipantCommandMessage(%q) = %q, want %q", gotType, gotMsg, tt.wantMsg)
		}
		if gotSev := participantCommandSeverity(gotType); gotSev != tt.wantSev {
			t.Fatalf("participantCommandSeverity(%q) = %q, want %q", gotType, gotSev, tt.wantSev)
		}
	}
}

func TestCbtSessionMapCbtRoomSetupError(t *testing.T) {
	if got := mapCbtRoomSetupError(nil); got != nil {
		t.Fatalf("mapCbtRoomSetupError(nil) = %v, want nil", got)
	}

	plainErr := errors.New("plain")
	if got := mapCbtRoomSetupError(plainErr); !errors.Is(got, plainErr) {
		t.Fatalf("mapCbtRoomSetupError(plain) = %v, want original", got)
	}

	for _, tt := range []struct {
		code       string
		wantIs     error
		wantSubstr string
	}{
		{code: "23505", wantIs: domain.ErrConflict, wantSubstr: "bertabrakan"},
		{code: "23503", wantIs: domain.ErrBadRequest, wantSubstr: "referensi"},
		{code: "23514", wantIs: domain.ErrBadRequest, wantSubstr: "nilai"},
	} {
		err := mapCbtRoomSetupError(&pgconn.PgError{Code: tt.code})
		if !errors.Is(err, tt.wantIs) {
			t.Fatalf("mapCbtRoomSetupError(%s) = %v, want errors.Is %v", tt.code, err, tt.wantIs)
		}
		if !strings.Contains(err.Error(), tt.wantSubstr) {
			t.Fatalf("mapCbtRoomSetupError(%s) = %q, want contains %q", tt.code, err.Error(), tt.wantSubstr)
		}
	}

	unknown := &pgconn.PgError{Code: "99999"}
	var target *pgconn.PgError
	if got := mapCbtRoomSetupError(unknown); !errors.As(got, &target) || target != unknown {
		t.Fatalf("mapCbtRoomSetupError(unknown) = %v, want original pg error", got)
	}
}
