package service

import (
	"context"
	"testing"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtEventListHelpersReturnEmptySlicesForNilStoreRows(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(31)
	svc := &CbtEvent{q: &fakeCbtEventStore7A{}}

	packages, err := svc.ListPackages(ctx, eventID)
	if err != nil || packages == nil || len(packages) != 0 {
		t.Fatalf("ListPackages(nil rows) = %#v, %v; want empty slice, nil", packages, err)
	}
	sessions, err := svc.ListSessions(ctx, eventID)
	if err != nil || sessions == nil || len(sessions) != 0 {
		t.Fatalf("ListSessions(nil rows) = %#v, %v; want empty slice, nil", sessions, err)
	}
	members, err := svc.ListMembers(ctx, eventID)
	if err != nil || members == nil || len(members) != 0 {
		t.Fatalf("ListMembers(nil rows) = %#v, %v; want empty slice, nil", members, err)
	}
	targets, err := svc.ListQuestionTargets(ctx, eventID)
	if err != nil || targets == nil || len(targets) != 0 {
		t.Fatalf("ListQuestionTargets(nil rows) = %#v, %v; want empty slice, nil", targets, err)
	}
}

func TestCbtEventOverviewBuildsReadyStateWithEmptyChildSlices(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(32)
	store := &fakeCbtEventStore7A{overviewRow: db.GetCbtEventOverviewSummaryRow{
		ID:                  eventID,
		MemberCount:         1,
		TargetQuestionCount: 10,
		PublishedQuestions:  10,
		ActivePackageCount:  1,
		SessionCount:        1,
		RoomCount:           1,
		ParticipantCount:    3,
		TokenReadyCount:     3,
		SubmittedCount:      3,
		ScoredCount:         3,
		JoinedCount:         2,
	}}
	svc := &CbtEvent{q: store}

	got, err := svc.Overview(ctx, eventID)
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if got.Event.ID != eventID {
		t.Fatalf("Overview event ID = %v, want %v", got.Event.ID, eventID)
	}
	if len(got.Members) != 0 || got.Members == nil || len(got.QuestionTargets) != 0 || got.QuestionTargets == nil || len(got.SubjectMatrix) != 0 || got.SubjectMatrix == nil || len(got.Packages) != 0 || got.Packages == nil || len(got.Sessions) != 0 || got.Sessions == nil {
		t.Fatalf("Overview child slices = members %#v targets %#v matrix %#v packages %#v sessions %#v, want non-nil empty slices", got.Members, got.QuestionTargets, got.SubjectMatrix, got.Packages, got.Sessions)
	}
	if got.BlockingReasons == nil || len(got.BlockingReasons) != 0 {
		t.Fatalf("BlockingReasons = %#v, want empty non-nil slice", got.BlockingReasons)
	}
	if !got.Readiness.AuthoringReady || !got.Readiness.PackageReady || !got.Readiness.SessionReady || !got.Readiness.RoomReady || !got.Readiness.TokenReady || !got.Readiness.CardReady || !got.Readiness.RuntimeStarted || !got.Readiness.ResultsReady {
		t.Fatalf("Readiness = %#v, want all ready", got.Readiness)
	}
}

func TestCbtEventListForUserInvalidAndFilteredMembership(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(33)
	otherID := cbtEventTestUUID(34)
	userID := cbtEventTestUUID(35)
	svc := &CbtEvent{q: &fakeCbtEventStore7A{
		listEventsRows:    []db.ListCbtExamEventsRow{{ID: eventID, Title: "Allowed"}, {ID: otherID, Title: "Hidden"}},
		membersByUserRows: []db.CbtEventMember{{EventID: eventID}},
	}}

	rows, err := svc.ListForUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListForUser() error = %v", err)
	}
	if len(rows) != 1 || rows[0].ID != eventID {
		t.Fatalf("ListForUser() rows = %#v, want only member event", rows)
	}
	rows, err = svc.ListForUser(ctx, cbtEventTestUUID(0))
	if err != nil || len(rows) != 1 {
		t.Fatalf("ListForUser(valid zero UUID) rows = %#v, err = %v; want membership-filtered row", rows, err)
	}
	rows, err = svc.ListForUser(ctx, db.CbtEventMember{}.UserID)
	if err != nil || rows == nil || len(rows) != 0 {
		t.Fatalf("ListForUser(invalid user) rows = %#v, err = %v; want empty slice, nil", rows, err)
	}
}
