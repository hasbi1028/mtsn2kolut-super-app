package db

import (
	"os"
	"strings"
	"testing"
)

func TestCBTQuestionFilteredQueryRedactsAnswerKeys(t *testing.T) {
	querySQL, err := os.ReadFile("../../../db/queries/cbt_questions.sql")
	if err != nil {
		t.Fatalf("read cbt questions query: %v", err)
	}

	query := strings.ToLower(string(querySQL))
	generated := strings.ToLower(listCbtQuestionsFiltered)
	required := []string{
		"case",
		"when sqlc.arg(is_admin)::bool",
		"q.author_username = sqlc.arg(actor_username)::text",
		"m.user_id = sqlc.arg(actor_user_id)::uuid",
		"m.role in ('reviewer', 'panitia')",
		"then q.answer_key else '' end as answer_key",
	}
	for _, needle := range required {
		if !strings.Contains(query, needle) {
			t.Fatalf("ListCbtQuestionsFiltered source query missing answer-key redaction clause %q", needle)
		}
	}

	generatedRequired := []string{
		"case",
		"when $1::bool",
		"q.author_username = $2::text",
		"m.user_id = $3::uuid",
		"m.role in ('reviewer', 'panitia')",
		"then q.answer_key else '' end as answer_key",
	}
	for _, needle := range generatedRequired {
		if !strings.Contains(generated, needle) {
			t.Fatalf("ListCbtQuestionsFiltered generated query missing answer-key redaction clause %q", needle)
		}
	}
}
