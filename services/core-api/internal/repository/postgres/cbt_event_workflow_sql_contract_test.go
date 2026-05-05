package db

import (
	"strings"
	"testing"
)

func TestCbtEventWorkflowQueriesStayEventScoped(t *testing.T) {
	queries := map[string]string{
		"event sessions":    listCbtEventSessionsReadiness,
		"event packages":    listCbtEventPackages,
		"package list":      listCbtPackages,
		"package questions": listCbtPackageQuestions,
	}
	for name, sql := range queries {
		t.Run(name, func(t *testing.T) {
			lower := strings.ToLower(sql)
			if !strings.Contains(lower, "event_id") {
				t.Fatalf("%s query must include event_id scoping", name)
			}
		})
	}
}

func TestCbtPackageCreatePersistsEventID(t *testing.T) {
	lower := strings.ToLower(createCbtPackage)
	if !strings.Contains(lower, "event_id") {
		t.Fatal("CreateCbtPackage must persist nullable event_id")
	}
}

func TestCbtEventReadinessQuestionCountsIncludeReusableBankQuestions(t *testing.T) {
	queries := map[string]string{
		"overview summary": getCbtEventOverviewSummary,
		"subject matrix":   listCbtEventSubjectMatrix,
		"subject targets":  listCbtEventSubjectTargets,
	}

	for name, sql := range queries {
		t.Run(name, func(t *testing.T) {
			lower := strings.ToLower(sql)
			if !strings.Contains(lower, "q.event_id is null and q.status = 'published'") {
				t.Fatalf("%s query must include reusable/global published questions", name)
			}
			if !strings.Contains(lower, "q.event_id =") {
				t.Fatalf("%s query must include same-event questions", name)
			}
		})
	}
}

func TestCbtEventReadinessQuestionCountsStayTargetSubjectScoped(t *testing.T) {
	queries := map[string]string{
		"overview summary": getCbtEventOverviewSummary,
		"subject matrix":   listCbtEventSubjectMatrix,
		"subject targets":  listCbtEventSubjectTargets,
	}

	for name, sql := range queries {
		t.Run(name, func(t *testing.T) {
			lower := strings.ToLower(sql)
			if !strings.Contains(lower, "q.subject_id") {
				t.Fatalf("%s query must scope reusable question counts by subject", name)
			}
		})
	}
}
