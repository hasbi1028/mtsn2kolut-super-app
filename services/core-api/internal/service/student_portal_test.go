package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestStudentPortalDerivesStudentIDFromUserAccount(t *testing.T) {
	userID := testGenerationUUID(31)
	studentID := testGenerationUUID(32)
	store := &fakeStudentPortalStore{studentID: studentID}
	svc := &StudentPortal{q: store}

	profile, err := svc.Profile(context.Background(), userID)
	if err != nil {
		t.Fatalf("Profile() error = %v", err)
	}
	if store.studentIDUserID != userID {
		t.Fatalf("GetPortalStudentIDByUserID userID = %s, want %s", store.studentIDUserID.String(), userID.String())
	}
	if store.profileStudentID != studentID || profile.ID != studentID {
		t.Fatalf("Profile() used studentID = %s/%s, want %s", store.profileStudentID.String(), profile.ID.String(), studentID.String())
	}

	if _, err := svc.Schedule(context.Background(), userID); err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}
	if store.scheduleStudentID != studentID {
		t.Fatalf("Schedule() used studentID = %s, want %s", store.scheduleStudentID.String(), studentID.String())
	}

	if _, err := svc.Results(context.Background(), userID); err != nil {
		t.Fatalf("Results() error = %v", err)
	}
	if store.resultsStudentID != studentID {
		t.Fatalf("Results() used studentID = %s, want %s", store.resultsStudentID.String(), studentID.String())
	}
}

func TestStudentPortalRejectsUsersWithoutLinkedStudent(t *testing.T) {
	svc := &StudentPortal{q: &fakeStudentPortalStore{studentIDErr: pgx.ErrNoRows}}

	if _, err := svc.Profile(context.Background(), testGenerationUUID(33)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Profile() error = %v, want ErrForbidden", err)
	}
}

type fakeStudentPortalStore struct {
	studentID       pgtype.UUID
	studentIDErr    error
	studentIDUserID pgtype.UUID

	profileStudentID  pgtype.UUID
	scheduleStudentID pgtype.UUID
	resultsStudentID  pgtype.UUID
}

func (f *fakeStudentPortalStore) GetPortalStudentIDByUserID(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	f.studentIDUserID = userID
	if f.studentIDErr != nil {
		return pgtype.UUID{}, f.studentIDErr
	}
	return f.studentID, nil
}

func (f *fakeStudentPortalStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error) {
	f.profileStudentID = id
	return db.GetStudentByIDRow{ID: id, Nama: "Siswa A"}, nil
}

func (f *fakeStudentPortalStore) ListStudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	f.scheduleStudentID = studentID
	return []db.ListStudentTimetableRow{{ID: testGenerationUUID(34), SubjectName: "IPA"}}, nil
}

func (f *fakeStudentPortalStore) ListStudentExamSessions(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	f.resultsStudentID = studentID
	return []db.ListStudentExamSessionsRow{{SessionID: testGenerationUUID(35), SessionTitle: "Ujian IPA", Token: "secret-token"}}, nil
}
