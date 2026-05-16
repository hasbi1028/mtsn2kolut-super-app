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

func TestCBTQuestionEventIDMigrationKeepsBankQuestionsReusable(t *testing.T) {
	content, err := os.ReadFile("../../../db/migrations/067_cbt_question_event_nullable.sql")
	if err != nil {
		t.Fatalf("read cbt question event nullable migration: %v", err)
	}
	sql := strings.ToLower(string(content))
	required := []string{
		"alter table cbt_questions",
		"alter column event_id drop not null",
	}
	for _, needle := range required {
		if !strings.Contains(sql, needle) {
			t.Fatalf("question event nullable migration missing %q", needle)
		}
	}
}

func TestCBTQuestionFilteredQuerySupportsReusableAndEventPoolScopes(t *testing.T) {
	querySQL, err := os.ReadFile("../../../db/queries/cbt_questions.sql")
	if err != nil {
		t.Fatalf("read cbt questions query: %v", err)
	}
	query := strings.ToLower(string(querySQL))
	generated := strings.ToLower(listCbtQuestionsFiltered)
	required := []string{
		"sqlc.arg(scope_filter)::text = 'global' and q.event_id is null",
		"sqlc.arg(scope_filter)::text = 'event_pool'",
		"q.event_id is null or (sqlc.arg(event_id)::uuid is not null and q.event_id = sqlc.arg(event_id)::uuid)",
	}
	for _, needle := range required {
		if !strings.Contains(query, needle) {
			t.Fatalf("ListCbtQuestionsFiltered source query missing scope clause %q", needle)
		}
	}

	generatedAlternatives := [][]string{
		{
			"$4::text = 'global' and q.event_id is null",
			"$4::text = 'event_pool'",
			"q.event_id is null or ($5::uuid is not null and q.event_id = $5::uuid)",
		},
		{
			"$6::text = 'global' and q.event_id is null",
			"$6::text = 'event_pool'",
			"q.event_id is null or ($7::uuid is not null and q.event_id = $7::uuid)",
		},
	}
	matched := false
	for _, requiredGroup := range generatedAlternatives {
		groupMatched := true
		for _, needle := range requiredGroup {
			if !strings.Contains(generated, needle) {
				groupMatched = false
				break
			}
		}
		if groupMatched {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("ListCbtQuestionsFiltered generated query missing reusable/event-pool scope clauses")
	}
}
