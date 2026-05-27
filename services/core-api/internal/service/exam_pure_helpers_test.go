package service

import (
	"encoding/json"
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

func TestExamDeterministicRuntimeRandomization(t *testing.T) {
	questions := []db.GetExamQuestionsRow{
		{ID: mustUUID(t, "10000000-0000-0000-0000-000000000011"), Code: "Q1", QuestionType: "multiple_choice", OptionA: "A1", OptionB: "B1", OptionC: "C1"},
		{ID: mustUUID(t, "10000000-0000-0000-0000-000000000012"), Code: "Q2", QuestionType: "multiple_choice", OptionA: "A2", OptionB: "B2", OptionC: "C2"},
		{ID: mustUUID(t, "10000000-0000-0000-0000-000000000013"), Code: "Q3", QuestionType: "multiple_choice", OptionA: "A3", OptionB: "B3", OptionC: "C3"},
	}

	first := orderQuestionsWithDrawSeeded(questions, nil, true, 0, 0, "session:participant")
	second := orderQuestionsWithDrawSeeded(questions, nil, true, 0, 0, "session:participant")
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("orderQuestionsWithDrawSeeded() should be stable for same participant seed")
	}

	firstOptions := ensureOptionOrderSeeded(questions, nil, true, "session:participant:options")
	secondOptions := ensureOptionOrderSeeded(questions, nil, true, "session:participant:options")
	if !reflect.DeepEqual(firstOptions, secondOptions) {
		t.Fatalf("ensureOptionOrderSeeded() should be stable for same participant seed")
	}
	if len(firstOptions) != len(questions) {
		t.Fatalf("ensureOptionOrderSeeded() generated %d orders, want %d", len(firstOptions), len(questions))
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

func TestExamApplyQuestionDrawSeparatesObjectiveAndEssay(t *testing.T) {
	objectiveOne := mustUUID(t, "31000000-0000-0000-0000-000000000001")
	essayOne := mustUUID(t, "31000000-0000-0000-0000-000000000002")
	objectiveTwo := mustUUID(t, "31000000-0000-0000-0000-000000000003")
	essayTwo := mustUUID(t, "31000000-0000-0000-0000-000000000004")
	otherID := mustUUID(t, "31000000-0000-0000-0000-000000000005")
	questions := []db.GetExamQuestionsRow{
		{ID: objectiveOne, Code: "PG-1", QuestionType: "multiple_choice"},
		{ID: essayOne, Code: "ES-1", QuestionType: "essay"},
		{ID: objectiveTwo, Code: "PG-2", QuestionType: "true_false"},
		{ID: essayTwo, Code: "ES-2", QuestionType: "essay"},
		{ID: otherID, Code: "OT-1", QuestionType: "unsupported"},
	}

	if got := applyQuestionDraw(questions, 0, 0); !reflect.DeepEqual(got, questions) {
		t.Fatalf("applyQuestionDraw(no draw) = %#v, want original", got)
	}

	got := applyQuestionDraw(questions, 1, 1)
	if len(got) != 2 {
		t.Fatalf("applyQuestionDraw(1,1) len = %d, want 2", len(got))
	}
	var objectiveCount, essayCount int
	lastPosition := -1
	positions := map[string]int{}
	for i, question := range questions {
		positions[pgUUIDString(question.ID)] = i
	}
	for _, question := range got {
		pos := positions[pgUUIDString(question.ID)]
		if pos < lastPosition {
			t.Fatalf("applyQuestionDraw() order = %#v, want original relative order", got)
		}
		lastPosition = pos
		switch question.QuestionType {
		case "essay":
			essayCount++
		case "multiple_choice", "true_false":
			objectiveCount++
		default:
			t.Fatalf("applyQuestionDraw() included unsupported question %#v", question)
		}
	}
	if objectiveCount != 1 || essayCount != 1 {
		t.Fatalf("applyQuestionDraw(1,1) counts objective/essay = %d/%d, want 1/1", objectiveCount, essayCount)
	}

	otherOnly := []db.GetExamQuestionsRow{{ID: otherID, Code: "OT-1", QuestionType: "unsupported"}}
	if got := applyQuestionDraw(otherOnly, 1, 0); !reflect.DeepEqual(got, otherOnly) {
		t.Fatalf("applyQuestionDraw(other-only) = %#v, want fallback original", got)
	}
}

func TestExamOptionOrderHelpersParseEnsureApplyAndCanonicalize(t *testing.T) {
	firstID := mustUUID(t, "32000000-0000-0000-0000-000000000001")
	secondID := mustUUID(t, "32000000-0000-0000-0000-000000000002")
	essayID := mustUUID(t, "32000000-0000-0000-0000-000000000003")
	singleID := mustUUID(t, "32000000-0000-0000-0000-000000000004")
	firstKey := pgUUIDString(firstID)
	secondKey := pgUUIDString(secondID)
	existingOrder := map[string][]string{firstKey: []string{"C", "A", "B"}}
	rows := []db.GetExamQuestionsRow{
		{
			ID:           firstID,
			QuestionType: "multiple_choice",
			Options:      marshalJSON([]QuestionOption{{Label: "A", Text: "Alpha"}, {Label: "B", Text: "Beta"}, {Label: "C", Text: "Gamma"}}),
			OptionA:      "Alpha",
			OptionB:      "Beta",
			OptionC:      "Gamma",
		},
		{ID: secondID, QuestionType: "multiple_answer", OptionA: "One", OptionB: "Two", OptionC: "Three"},
		{ID: essayID, QuestionType: "essay", OptionA: "Ignored", OptionB: "Ignored"},
		{ID: singleID, QuestionType: "multiple_choice", OptionA: "Only one"},
	}

	if got := ensureOptionOrder(rows, existingOrder, false); len(got) != 0 {
		t.Fatalf("ensureOptionOrder(randomize=false) = %#v, want empty map", got)
	}
	got := ensureOptionOrder(rows, existingOrder, true)
	if !reflect.DeepEqual(got[firstKey], []string{"C", "A", "B"}) {
		t.Fatalf("ensureOptionOrder() existing first = %#v, want preserved C,A,B", got[firstKey])
	}
	if labels := got[secondKey]; len(labels) != 3 || !sameStringSet(labels, []string{"A", "B", "C"}) {
		t.Fatalf("ensureOptionOrder() generated second = %#v, want labels A/B/C", labels)
	}
	if _, ok := got[pgUUIDString(essayID)]; ok {
		t.Fatalf("ensureOptionOrder() generated essay order = %#v", got[pgUUIDString(essayID)])
	}
	if _, ok := got[pgUUIDString(singleID)]; ok {
		t.Fatalf("ensureOptionOrder() generated single-option order = %#v", got[pgUUIDString(singleID)])
	}

	orderJSON, err := json.Marshal(map[string][]string{firstKey: []string{"C", "A", "B"}})
	if err != nil {
		t.Fatalf("marshal option order: %v", err)
	}
	if parsed := parseOptionOrder(orderJSON); !reflect.DeepEqual(parsed[firstKey], []string{"C", "A", "B"}) {
		t.Fatalf("parseOptionOrder(valid) = %#v, want C,A,B", parsed[firstKey])
	}
	for _, raw := range [][]byte{nil, []byte("not-json"), []byte("null")} {
		if parsed := parseOptionOrder(raw); len(parsed) != 0 {
			t.Fatalf("parseOptionOrder(%q) = %#v, want empty", string(raw), parsed)
		}
	}

	applied := applyExamOptionOrder(rows[0], []string{"C", "A", "B"})
	options := decodeQuestionOptions(applied.Options)
	if len(options) != 3 || options[0].Label != "A" || options[0].Text != "Gamma" || options[1].Label != "B" || options[1].Text != "Alpha" || options[2].Label != "C" || options[2].Text != "Beta" {
		t.Fatalf("applyExamOptionOrder(JSON) options = %#v, want relabeled C/A/B text order", options)
	}
	if applied.OptionA != "Gamma" || applied.OptionB != "Alpha" || applied.OptionC != "Beta" {
		t.Fatalf("applyExamOptionOrder(JSON) legacy columns = %q/%q/%q, want Gamma/Alpha/Beta", applied.OptionA, applied.OptionB, applied.OptionC)
	}
	legacyApplied := applyExamOptionOrder(rows[1], []string{"C", "A"})
	legacyOptions := decodeQuestionOptions(legacyApplied.Options)
	if len(legacyOptions) != 2 || legacyOptions[0].Label != "A" || legacyOptions[0].Text != "Three" || legacyApplied.OptionA != "Three" || legacyApplied.OptionB != "One" {
		t.Fatalf("applyExamOptionOrder(legacy partial) = options %#v columns %q/%q", legacyOptions, legacyApplied.OptionA, legacyApplied.OptionB)
	}
	unchanged := applyExamOptionOrder(rows[2], []string{"B", "A"})
	if !reflect.DeepEqual(unchanged, rows[2]) {
		t.Fatalf("applyExamOptionOrder(essay) = %#v, want unchanged", unchanged)
	}
	missing := applyExamOptionOrder(rows[0], []string{"Z"})
	if !reflect.DeepEqual(missing, rows[0]) {
		t.Fatalf("applyExamOptionOrder(missing labels) = %#v, want unchanged", missing)
	}

	if answer := canonicalizeRandomizedAnswer(orderJSON, firstID, " A, c , Z "); answer != "C,B,Z" {
		t.Fatalf("canonicalizeRandomizedAnswer(mapped) = %q, want C,B,Z", answer)
	}
	if answer := canonicalizeRandomizedAnswer(orderJSON, firstID, "Z"); answer != "Z" {
		t.Fatalf("canonicalizeRandomizedAnswer(unchanged) = %q, want Z", answer)
	}
	if answer := canonicalizeRandomizedAnswer(nil, firstID, "A"); answer != "A" {
		t.Fatalf("canonicalizeRandomizedAnswer(no order) = %q, want A", answer)
	}
}

func sameStringSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	counts := map[string]int{}
	for _, value := range got {
		counts[value]++
	}
	for _, value := range want {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	return true
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
