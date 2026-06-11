package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtQuestionSummaryUsesActorCapabilitiesForScopedQueries(t *testing.T) {
	userID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000201")
	store := &fakeQuestionStore{
		summaryCounts:    db.GetCbtQuestionSummaryCountsRow{Total: 9, Draft: 2, PackageReady: 4},
		summarySubjects:  []db.ListCbtQuestionSummaryBySubjectRow{{SubjectID: mustQuestionUUID(t, "00000000-0000-0000-0000-000000000202"), SubjectName: "Matematika", SubjectCode: "MTK", Total: 3}},
		summaryCognitive: []db.ListCbtQuestionSummaryByCognitiveLevelRow{{CognitiveLevel: "C3", Total: 5}},
		summaryRecent:    []db.ListCbtQuestionSummaryRecentRow{{ID: mustQuestionUUID(t, "00000000-0000-0000-0000-000000000203"), Code: "Q-1"}},
	}
	svc := NewCbtQuestion(nil)
	svc.q = store

	got, err := svc.Summary(context.Background(), CbtQuestionActor{
		UserID:      userID,
		Username:    " guru.mtk ",
		Permissions: []string{"bank_soal.use_in_package"},
	})
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if got.Counts.Total != 9 || len(got.BySubject) != 1 || len(got.ByCognitiveLevel) != 1 || len(got.Recent) != 1 {
		t.Fatalf("Summary() = %+v, want populated sections", got)
	}
	if store.summaryCountsArg.IsAdmin {
		t.Fatalf("Summary() IsAdmin = true, want false for scoped non-admin actor")
	}
	if !store.summaryCountsArg.CanUseInPackage || !store.summarySubjectArg.CanUseInPackage || !store.summaryCogArg.CanUseInPackage || !store.summaryRecentArg.CanUseInPackage {
		t.Fatalf("Summary() did not propagate package-use capability to all query args")
	}
	if store.summaryCountsArg.ActorUsername != "guru.mtk" || store.summarySubjectArg.ActorUsername != "guru.mtk" || store.summaryCogArg.ActorUsername != "guru.mtk" || store.summaryRecentArg.ActorUsername != "guru.mtk" {
		t.Fatalf("Summary() did not trim/propagate username args: counts=%q subject=%q cog=%q recent=%q", store.summaryCountsArg.ActorUsername, store.summarySubjectArg.ActorUsername, store.summaryCogArg.ActorUsername, store.summaryRecentArg.ActorUsername)
	}
	if store.summaryCountsArg.ActorUserID != userID || store.summarySubjectArg.ActorUserID != userID || store.summaryCogArg.ActorUserID != userID || store.summaryRecentArg.ActorUserID != userID {
		t.Fatalf("Summary() did not propagate actor user id")
	}
}

func TestCbtQuestionSummaryAdminReadAllAndActorHelpers(t *testing.T) {
	store := &fakeQuestionStore{}
	svc := NewCbtQuestion(nil)
	svc.q = store

	if _, err := svc.Summary(context.Background(), CbtQuestionActor{Roles: []string{" admin "}, Permissions: []string{"bank_soal.use_in_package"}}); err != nil {
		t.Fatalf("Summary(admin) error = %v", err)
	}
	if !store.summaryCountsArg.IsAdmin || !store.summarySubjectArg.IsAdmin || !store.summaryCogArg.IsAdmin || !store.summaryRecentArg.IsAdmin {
		t.Fatalf("Summary(admin) did not propagate IsAdmin to all summary queries")
	}
	if !store.summaryCountsArg.CanUseInPackage {
		t.Fatalf("Summary(admin with bank_soal.use_in_package) CanUseInPackage = false, want true")
	}

	actor := CbtQuestionActor{Roles: []string{" guru ", "admin"}, Permissions: []string{" bank_soal.publish ", "bank_soal.read_all"}}
	if !actor.HasRole("admin") || !actor.IsAdmin() {
		t.Fatalf("actor admin role helpers returned false")
	}
	if !actor.HasPermission("bank_soal.publish") || !actor.CanPublishBankSoal() || !actor.CanReadAllBankSoal() {
		t.Fatalf("actor permission helpers returned false")
	}
	if actor.HasRole("") || actor.HasPermission("") || actor.HasRole("staf") || actor.HasPermission("missing") {
		t.Fatalf("actor helpers returned true for empty/missing target")
	}
}

func TestWorkflowScopeGradeLevel(t *testing.T) {
	tests := []struct {
		name  string
		input pgtype.Text
		want  pgtype.Int2
	}{
		{name: "invalid", input: pgtype.Text{}, want: pgtype.Int2{}},
		{name: "VII", input: pgtype.Text{String: "VII", Valid: true}, want: pgtype.Int2{Int16: 7, Valid: true}},
		{name: "VIII", input: pgtype.Text{String: "VIII", Valid: true}, want: pgtype.Int2{Int16: 8, Valid: true}},
		{name: "IX", input: pgtype.Text{String: " IX ", Valid: true}, want: pgtype.Int2{Int16: 9, Valid: true}},
		{name: "unknown", input: pgtype.Text{String: "X", Valid: true}, want: pgtype.Int2{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := workflowScopeGradeLevel(tt.input); got != tt.want {
				t.Fatalf("workflowScopeGradeLevel(%+v) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
