package db

import (
	"strings"
	"testing"
)

func TestCbtRoomTokenGenerationUsesShortSupervisorFriendlyFormat(t *testing.T) {
	if !strings.Contains(createCbtExamRoom, "'R' || LPAD") {
		t.Fatalf("CreateCbtExamRoom must prefix room tokens with room ordinal RNN: %s", createCbtExamRoom)
	}
	if !strings.Contains(createCbtExamRoom, "'-' || UPPER") || !strings.Contains(createCbtExamRoom, "gen_random_bytes(3)") {
		t.Fatalf("CreateCbtExamRoom must generate short uppercase RNN-XXXX room tokens: %s", createCbtExamRoom)
	}
	if strings.Contains(createCbtExamRoom, "gen_random_bytes(16)") || strings.Contains(createCbtExamRoom, "md5") {
		t.Fatalf("CreateCbtExamRoom still contains long/opaque room token generation: %s", createCbtExamRoom)
	}
}
