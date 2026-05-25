package db

import (
	"strings"
	"testing"
)

func TestCbtParticipantTokenGenerationUsesEightHexCharacters(t *testing.T) {
	queries := map[string]string{
		"EnrollClassToSession":       enrollClassToSession,
		"EnrollGradeToSession":       enrollGradeToSession,
		"EnrollSchoolToSession":      enrollSchoolToSession,
		"GenerateTokensForSession":   generateTokensForSession,
		"RegenerateParticipantToken": regenerateParticipantToken,
	}

	for name, query := range queries {
		if !strings.Contains(query, "gen_random_bytes(4)") {
			t.Fatalf("%s must generate 8-character hex tokens with gen_random_bytes(4): %s", name, query)
		}
		if strings.Contains(query, "gen_random_bytes(16)") || strings.Contains(query, "{32}") {
			t.Fatalf("%s still contains legacy 32-character token generation/validation: %s", name, query)
		}
	}
}
