package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeUserProfileCandidateStore struct {
	employeeArg db.ListEmployeeProfileCandidatesParams
	employee    []db.ListEmployeeProfileCandidatesRow
	employeeErr error

	studentArg db.ListStudentProfileCandidatesByClassParams
	student    []db.ListStudentProfileCandidatesByClassRow
	studentErr error

	parentArg db.ListParentProfileCandidatesByChildClassParams
	parent    []db.ListParentProfileCandidatesByChildClassRow
	parentErr error
}

func (f *fakeUserProfileCandidateStore) ListEmployeeProfileCandidates(ctx context.Context, arg db.ListEmployeeProfileCandidatesParams) ([]db.ListEmployeeProfileCandidatesRow, error) {
	f.employeeArg = arg
	return f.employee, f.employeeErr
}

func (f *fakeUserProfileCandidateStore) ListStudentProfileCandidatesByClass(ctx context.Context, arg db.ListStudentProfileCandidatesByClassParams) ([]db.ListStudentProfileCandidatesByClassRow, error) {
	f.studentArg = arg
	return f.student, f.studentErr
}

func (f *fakeUserProfileCandidateStore) ListParentProfileCandidatesByChildClass(ctx context.Context, arg db.ListParentProfileCandidatesByChildClassParams) ([]db.ListParentProfileCandidatesByChildClassRow, error) {
	f.parentArg = arg
	return f.parent, f.parentErr
}

func profileCandidateUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func TestUserProfileCandidatesScopeEmployeeAndStudentRoles(t *testing.T) {
	employeeID := profileCandidateUUID(1)
	linkedUserID := profileCandidateUUID(2)
	classID := profileCandidateUUID(3)
	studentID := profileCandidateUUID(4)
	store := &fakeUserProfileCandidateStore{
		employee: []db.ListEmployeeProfileCandidatesRow{{
			ID:             employeeID,
			Nama:           "Hasbi Guru",
			Identifier:     "198001",
			LinkedUserID:   linkedUserID,
			LinkedUsername: "hasbi",
		}},
		student: []db.ListStudentProfileCandidatesByClassRow{{
			ID:         studentID,
			Nama:       "Ahmad Siswa",
			Identifier: "2026001",
			ClassID:    classID,
			ClassName:  "VII.A",
		}},
	}
	svc := NewUserProfileCandidateService(store)

	employees, err := svc.List(context.Background(), UserProfileCandidateFilter{
		Role:          "guru",
		Query:         "  hasbi  ",
		IncludeLinked: true,
		Limit:         250,
	})
	if err != nil {
		t.Fatalf("List(guru) error = %v", err)
	}
	if store.employeeArg.Q != "hasbi" || !store.employeeArg.IncludeLinked || store.employeeArg.LimitCount != 100 {
		t.Fatalf("employee query args = %+v, want trimmed query/include/capped limit", store.employeeArg)
	}
	if store.studentArg.ClassID.Valid || store.parentArg.ClassID.Valid {
		t.Fatalf("employee role should not call student/parent stores: student=%+v parent=%+v", store.studentArg, store.parentArg)
	}
	if len(employees.Candidates) != 1 || employees.Candidates[0].ProfileType != "employee" || employees.Candidates[0].ID != employeeID.String() || !employees.Candidates[0].IsLinked || employees.Candidates[0].LinkedUsername != "hasbi" {
		t.Fatalf("employee candidates = %+v", employees.Candidates)
	}

	students, err := svc.List(context.Background(), UserProfileCandidateFilter{
		Role:    "siswa",
		ClassID: classID.String(),
		Query:   "ahmad",
		Limit:   25,
	})
	if err != nil {
		t.Fatalf("List(siswa) error = %v", err)
	}
	if store.studentArg.ClassID != classID || store.studentArg.Q != "ahmad" || store.studentArg.IncludeLinked || store.studentArg.LimitCount != 25 {
		t.Fatalf("student query args = %+v, want class-scoped args", store.studentArg)
	}
	if len(students.Candidates) != 1 || students.Candidates[0].ProfileType != "student" || students.Candidates[0].ClassID != classID.String() {
		t.Fatalf("student candidates = %+v", students.Candidates)
	}
}

func TestUserProfileCandidatesRequireClassForStudentAndParentRoles(t *testing.T) {
	svc := NewUserProfileCandidateService(&fakeUserProfileCandidateStore{})

	for _, role := range []string{"siswa", "ortu"} {
		_, err := svc.List(context.Background(), UserProfileCandidateFilter{Role: role})
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("List(%s without class_id) error = %v, want bad request", role, err)
		}
	}

	if _, err := svc.List(context.Background(), UserProfileCandidateFilter{Role: "superadmin"}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("List(invalid role) error = %v, want bad request", err)
	}
}

func TestUserProfileCandidatesAggregateParentsByChildClass(t *testing.T) {
	classID := profileCandidateUUID(10)
	parentID := profileCandidateUUID(11)
	childOneID := profileCandidateUUID(12)
	childTwoID := profileCandidateUUID(13)
	store := &fakeUserProfileCandidateStore{
		parent: []db.ListParentProfileCandidatesByChildClassRow{
			{ID: parentID, Nama: "Wali Ahmad", Identifier: "081234567890", ClassID: classID, ClassName: "VII.A", ChildID: childOneID, ChildNama: "Ahmad"},
			{ID: parentID, Nama: "Wali Ahmad", Identifier: "081234567890", ClassID: classID, ClassName: "VII.A", ChildID: childTwoID, ChildNama: "Aminah"},
		},
	}
	svc := NewUserProfileCandidateService(store)

	got, err := svc.List(context.Background(), UserProfileCandidateFilter{
		Role:    "ortu",
		ClassID: classID.String(),
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("List(ortu) error = %v", err)
	}
	if store.parentArg.ClassID != classID || store.parentArg.LimitCount != 10 {
		t.Fatalf("parent query args = %+v, want child-class scoped args", store.parentArg)
	}
	if len(got.Candidates) != 1 {
		t.Fatalf("parent candidates len = %d, want 1: %+v", len(got.Candidates), got.Candidates)
	}
	parent := got.Candidates[0]
	if parent.ProfileType != "parent" || parent.ID != parentID.String() || len(parent.Children) != 2 {
		t.Fatalf("parent candidate = %+v, want deduped parent with two children", parent)
	}
	if strings.Contains(parent.Identifier, "081234567890") {
		t.Fatalf("parent identifier exposes raw phone number: %q", parent.Identifier)
	}
}

func TestUserProfileCandidateSQLKeepsFilteringInPostgres(t *testing.T) {
	raw, err := os.ReadFile("../../db/queries/users.sql")
	if err != nil {
		t.Fatalf("read users.sql: %v", err)
	}
	sql := string(raw)
	for _, expected := range []string{
		"-- name: ListEmployeeProfileCandidates",
		"-- name: ListStudentProfileCandidatesByClass",
		"-- name: ListParentProfileCandidatesByChildClass",
		"s.class_id = sqlc.arg(class_id)",
		"JOIN parent_students",
		"linked.id IS NULL",
	} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("users.sql missing backend-side candidate filtering marker %q", expected)
		}
	}
}
