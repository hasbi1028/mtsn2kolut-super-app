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
