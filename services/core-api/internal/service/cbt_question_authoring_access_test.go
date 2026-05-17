package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtQuestionRequireCreateQuestionAccess(t *testing.T) {
	ctx := context.Background()
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000001001")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000001002")
	otherSubjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000001003")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000001004")

	tests := []struct {
		name      string
		actor     CbtQuestionActor
		eventID   pgtype.UUID
		members   []db.CbtEventMember
		wantError error
	}{
		{
			name:    "admin can create for event without membership lookup match",
			actor:   CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
			eventID: eventID,
		},
		{
			name:    "guru can create global bank question without event membership",
			actor:   CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
			eventID: pgtype.UUID{},
		},
		{
			name:    "create permission can create global bank question",
			actor:   CbtQuestionActor{Username: "author", Permissions: []string{"bank_soal.create"}},
			eventID: pgtype.UUID{},
		},
		{
			name:      "non author role cannot create global bank question",
			actor:     CbtQuestionActor{Username: "student", Roles: []string{"siswa"}},
			eventID:   pgtype.UUID{},
			wantError: domain.ErrForbidden,
		},
		{
			name:    "event pembuat_soal member can create matching subject",
			actor:   CbtQuestionActor{UserID: actorID, Username: "writer", Roles: []string{"guru"}},
			eventID: eventID,
			members: []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRolePembuatSoal}},
		},
		{
			name:    "event pembuat_soal wildcard subject can create any subject",
			actor:   CbtQuestionActor{UserID: actorID, Username: "writer", Roles: []string{"guru"}},
			eventID: eventID,
			members: []db.CbtEventMember{{EventID: eventID, SubjectID: pgtype.UUID{}, Role: db.CbtEventMemberRolePembuatSoal}},
		},
		{
			name:      "event member wrong subject cannot create",
			actor:     CbtQuestionActor{UserID: actorID, Username: "writer", Roles: []string{"guru"}},
			eventID:   eventID,
			members:   []db.CbtEventMember{{EventID: eventID, SubjectID: otherSubjectID, Role: db.CbtEventMemberRolePembuatSoal}},
			wantError: domain.ErrForbidden,
		},
		{
			name:      "event reviewer cannot create questions",
			actor:     CbtQuestionActor{UserID: actorID, Username: "reviewer", Roles: []string{"guru"}},
			eventID:   eventID,
			members:   []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}},
			wantError: domain.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCbtQuestion(nil)
			svc.q = &fakeQuestionStore{membersByUser: tt.members}
			err := svc.requireCreateQuestion(ctx, tt.actor, tt.eventID, subjectID)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("requireCreateQuestion() error = %v, want %v", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("requireCreateQuestion() error = %v", err)
			}
		})
	}
}

func TestCbtQuestionActorHasEventQuestionRole(t *testing.T) {
	ctx := context.Background()
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000002001")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000002002")
	otherEventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000002003")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000002004")

	svc := NewCbtQuestion(nil)
	svc.q = &fakeQuestionStore{membersByUser: []db.CbtEventMember{
		{EventID: otherEventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer},
		{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRolePembuatSoal},
		{EventID: eventID, SubjectID: pgtype.UUID{}, Role: db.CbtEventMemberRolePanitia},
	}}

	ok, err := svc.actorHasEventQuestionRole(ctx, CbtQuestionActor{UserID: actorID}, eventID, subjectID, db.CbtEventMemberRoleReviewer)
	if err != nil {
		t.Fatalf("actorHasEventQuestionRole() error = %v", err)
	}
	if ok {
		t.Fatal("actorHasEventQuestionRole() matched reviewer role from another event")
	}

	ok, err = svc.actorHasEventQuestionRole(ctx, CbtQuestionActor{UserID: actorID}, eventID, subjectID, db.CbtEventMemberRolePembuatSoal)
	if err != nil || !ok {
		t.Fatalf("actorHasEventQuestionRole(pembuat_soal) = %v, %v; want true, nil", ok, err)
	}

	ok, err = svc.actorHasEventQuestionRole(ctx, CbtQuestionActor{UserID: actorID}, eventID, subjectID, db.CbtEventMemberRolePanitia)
	if err != nil || !ok {
		t.Fatalf("actorHasEventQuestionRole(panitia wildcard) = %v, %v; want true, nil", ok, err)
	}

	ok, err = svc.actorHasEventQuestionRole(ctx, CbtQuestionActor{}, eventID, subjectID, db.CbtEventMemberRolePanitia)
	if err != nil || ok {
		t.Fatalf("actorHasEventQuestionRole(missing actor id) = %v, %v; want false, nil", ok, err)
	}
}

func TestCbtQuestionRequireDuplicateSourceAccess(t *testing.T) {
	ctx := context.Background()
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000003001")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000003002")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000003003")

	tests := []struct {
		name      string
		actor     CbtQuestionActor
		current   db.GetCbtQuestionRow
		members   []db.CbtEventMember
		wantError error
	}{
		{
			name:    "admin can duplicate any source",
			actor:   CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
			current: db.GetCbtQuestionRow{AuthorUsername: "other", EventID: eventID, SubjectID: subjectID},
		},
		{
			name:    "global author can duplicate own source without event role",
			actor:   CbtQuestionActor{Username: "author"},
			current: db.GetCbtQuestionRow{AuthorUsername: "author", SubjectID: subjectID},
		},
		{
			name:    "event author must still have pembuat_soal role",
			actor:   CbtQuestionActor{UserID: actorID, Username: "author"},
			current: db.GetCbtQuestionRow{AuthorUsername: "author", EventID: eventID, SubjectID: subjectID},
			members: []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRolePembuatSoal}},
		},
		{
			name:      "event author without pembuat_soal role cannot duplicate own source",
			actor:     CbtQuestionActor{UserID: actorID, Username: "author"},
			current:   db.GetCbtQuestionRow{AuthorUsername: "author", EventID: eventID, SubjectID: subjectID},
			members:   []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}},
			wantError: domain.ErrForbidden,
		},
		{
			name:    "reviewer member can duplicate another author event source",
			actor:   CbtQuestionActor{UserID: actorID, Username: "reviewer"},
			current: db.GetCbtQuestionRow{AuthorUsername: "author", EventID: eventID, SubjectID: subjectID},
			members: []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}},
		},
		{
			name:    "panitia wildcard member can duplicate another author event source",
			actor:   CbtQuestionActor{UserID: actorID, Username: "panitia"},
			current: db.GetCbtQuestionRow{AuthorUsername: "author", EventID: eventID, SubjectID: subjectID},
			members: []db.CbtEventMember{{EventID: eventID, SubjectID: pgtype.UUID{}, Role: db.CbtEventMemberRolePanitia}},
		},
		{
			name:      "unrelated actor cannot duplicate source",
			actor:     CbtQuestionActor{UserID: actorID, Username: "other"},
			current:   db.GetCbtQuestionRow{AuthorUsername: "author", EventID: eventID, SubjectID: subjectID},
			wantError: domain.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCbtQuestion(nil)
			svc.q = &fakeQuestionStore{membersByUser: tt.members}
			err := svc.requireDuplicateSourceAccess(ctx, tt.actor, tt.current)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("requireDuplicateSourceAccess() error = %v, want %v", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("requireDuplicateSourceAccess() error = %v", err)
			}
		})
	}
}

func TestCbtQuestionCreateAccessAndWorkflowGuards(t *testing.T) {
	ctx := context.Background()
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000004001")
	createdID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000004002")
	store := &fakeQuestionStore{createRow: db.CbtQuestion{ID: createdID, SubjectID: subjectID, WorkflowStatus: "draft", AuthorUsername: "guru"}}
	svc := NewCbtQuestion(nil)
	svc.q = store

	input := SaveCbtQuestionInput{
		SubjectID:      subjectID,
		QuestionText:   "Apa ibu kota Indonesia?",
		QuestionType:   "multiple_choice",
		OptionA:        "Jakarta",
		OptionB:        "Bandung",
		AnswerKey:      "A",
		WorkflowStatus: "draft",
		AuthorUsername: "guru",
		Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
	}
	row, err := svc.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if row.ID != createdID || store.createCalls != 1 || store.auditCalls != 1 {
		t.Fatalf("Create() row/calls = %+v createCalls=%d auditCalls=%d, want created row and audit", row, store.createCalls, store.auditCalls)
	}
	if store.createParams.AuthorUsername != "guru" || store.createParams.WorkflowStatus != "draft" || store.createParams.Status != db.CbtQuestionStatusEnumDraft {
		t.Fatalf("Create() params = %+v, want draft authored by guru", store.createParams)
	}

	store.createCalls = 0
	input.Status = db.CbtQuestionStatusEnumPublished
	input.WorkflowStatus = "published"
	_, err = svc.Create(ctx, input)
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Create(published workflow bypass) error = %v, want ErrBadRequest", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("Create(published workflow bypass) createCalls = %d, want 0", store.createCalls)
	}
}

func TestCbtQuestionCreateRejectsMissingSubjectOrEventBeforeInsert(t *testing.T) {
	ctx := context.Background()
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000004101")
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000004102")
	baseInput := SaveCbtQuestionInput{
		SubjectID:      subjectID,
		EventID:        eventID,
		QuestionText:   "Apa ibu kota Indonesia?",
		QuestionType:   "multiple_choice",
		OptionA:        "Jakarta",
		OptionB:        "Bandung",
		AnswerKey:      "A",
		WorkflowStatus: "draft",
		AuthorUsername: "guru",
		Actor:          CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
	}

	t.Run("missing subject", func(t *testing.T) {
		store := &fakeQuestionStore{subjectMissing: true, createRow: db.CbtQuestion{ID: mustQuestionUUID(t, "00000000-0000-0000-0000-000000004103"), SubjectID: subjectID}}
		svc := NewCbtQuestion(nil)
		svc.q = store
		_, err := svc.Create(ctx, baseInput)
		if !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "subject_id tidak ditemukan") {
			t.Fatalf("Create(missing subject) error = %v, want subject bad request", err)
		}
		if store.createCalls != 0 {
			t.Fatalf("Create(missing subject) createCalls = %d, want 0", store.createCalls)
		}
	})

	t.Run("missing event", func(t *testing.T) {
		store := &fakeQuestionStore{eventMissing: true, createRow: db.CbtQuestion{ID: mustQuestionUUID(t, "00000000-0000-0000-0000-000000004104"), SubjectID: subjectID}}
		svc := NewCbtQuestion(nil)
		svc.q = store
		_, err := svc.Create(ctx, baseInput)
		if !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "event_id tidak ditemukan") {
			t.Fatalf("Create(missing event) error = %v, want event bad request", err)
		}
		if store.createCalls != 0 {
			t.Fatalf("Create(missing event) createCalls = %d, want 0", store.createCalls)
		}
	})
}

func TestCbtQuestionDuplicateForRevisionSetsVersionSourceAndAccess(t *testing.T) {
	ctx := context.Background()
	questionID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000005001")
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000005002")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000005003")
	versionGroupID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000005004")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000005005")
	newID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000005006")

	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{
			ID:             questionID,
			EventID:        eventID,
			SubjectID:      subjectID,
			Code:           "Q-OLD",
			QuestionText:   "Soal lama",
			QuestionType:   "multiple_choice",
			OptionA:        "A",
			OptionB:        "B",
			AnswerKey:      "A",
			Status:         db.CbtQuestionStatusEnumPublished,
			WorkflowStatus: "published",
			VersionGroupID: versionGroupID,
			VersionNumber:  3,
			AuthorUsername: "author",
		},
		membersByUser:     []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRolePembuatSoal}},
		createRow:         db.CbtQuestion{ID: newID, EventID: eventID, SubjectID: subjectID, WorkflowStatus: "rejected", Status: db.CbtQuestionStatusEnumDraft},
		nextVersionNumber: 5,
	}
	svc := NewCbtQuestion(nil)
	svc.q = store

	row, err := svc.DuplicateForRevision(ctx, questionID, CbtQuestionActor{UserID: actorID, Username: "author"}, "analisis butir")
	if err != nil {
		t.Fatalf("DuplicateForRevision() error = %v", err)
	}
	if row.ID != newID || store.createCalls != 1 || store.auditCalls != 1 || store.markGroupLatestCalls != 1 {
		t.Fatalf("DuplicateForRevision() row/calls = %+v create=%d audit=%d markGroup=%d", row, store.createCalls, store.auditCalls, store.markGroupLatestCalls)
	}
	if store.markGroupNotLatestID != versionGroupID {
		t.Fatalf("MarkCbtQuestionVersionGroupNotLatest id = %v, want %v", store.markGroupNotLatestID, versionGroupID)
	}
	if store.createParams.SourceQuestionID != questionID || store.createParams.SupersedesQuestionID != questionID {
		t.Fatalf("DuplicateForRevision() source/supersedes = %v/%v, want original id", store.createParams.SourceQuestionID, store.createParams.SupersedesQuestionID)
	}
	if store.createParams.VersionGroupID != versionGroupID || store.createParams.VersionNumber != 5 || !store.createParams.IsLatestVersion {
		t.Fatalf("DuplicateForRevision() version params = group %v number %d latest %v, want group v5 latest", store.createParams.VersionGroupID, store.createParams.VersionNumber, store.createParams.IsLatestVersion)
	}
	if store.createParams.WorkflowStatus != "rejected" || store.createParams.Status != db.CbtQuestionStatusEnumDraft {
		t.Fatalf("DuplicateForRevision() workflow/status = %q/%q, want rejected/draft", store.createParams.WorkflowStatus, store.createParams.Status)
	}
}

func TestCbtQuestionVersionsRequiresSourceAccess(t *testing.T) {
	ctx := context.Background()
	questionID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000006001")
	eventID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000006002")
	subjectID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000006003")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000006004")
	store := &fakeQuestionStore{
		detail:      db.GetCbtQuestionDetailRow{ID: questionID, EventID: eventID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "submitted", AuthorUsername: "author", AnswerKey: "A"},
		versionRows: []db.ListCbtQuestionVersionsRow{{ID: questionID, Code: "Q-1", VersionNumber: 1, IsLatestVersion: true}},
	}
	svc := NewCbtQuestion(nil)
	svc.q = store

	_, err := svc.Versions(ctx, questionID, CbtQuestionActor{UserID: actorID, Username: "other"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Versions(unrelated actor) error = %v, want ErrForbidden", err)
	}

	store.membersByUser = []db.CbtEventMember{{EventID: eventID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}}
	rows, err := svc.Versions(ctx, questionID, CbtQuestionActor{UserID: actorID, Username: "reviewer"})
	if err != nil {
		t.Fatalf("Versions(reviewer member) error = %v", err)
	}
	if len(rows) != 1 || rows[0].ID != questionID || rows[0].VersionNumber != 1 {
		t.Fatalf("Versions(reviewer member) = %+v, want version row", rows)
	}
}
