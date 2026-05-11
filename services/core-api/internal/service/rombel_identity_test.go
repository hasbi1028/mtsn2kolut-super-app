package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type noopRombelStore struct{}

func (noopRombelStore) ListRombels(context.Context) ([]db.ListRombelsRow, error) {
	return nil, nil
}
func (noopRombelStore) GetRombelDetail(context.Context, pgtype.UUID) (db.GetRombelDetailRow, error) {
	return db.GetRombelDetailRow{}, nil
}
func (noopRombelStore) CountRombelCodeConflicts(context.Context, db.CountRombelCodeConflictsParams) (int32, error) {
	return 0, nil
}
func (noopRombelStore) UpdateRombelIdentity(context.Context, db.UpdateRombelIdentityParams) (db.UpdateRombelIdentityRow, error) {
	return db.UpdateRombelIdentityRow{}, nil
}
func (noopRombelStore) ListStudentsByClassWithParents(context.Context, pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error) {
	return nil, nil
}
func (noopRombelStore) ListRombelSubjectAssignments(context.Context, pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error) {
	return nil, nil
}
func (noopRombelStore) GetRombelSubjectAssignment(context.Context, db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error) {
	return db.GetRombelSubjectAssignmentRow{}, nil
}
func (noopRombelStore) CreateRombelSubjectAssignment(context.Context, db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error) {
	return db.CreateRombelSubjectAssignmentRow{}, nil
}
func (noopRombelStore) UpdateRombelSubjectAssignment(context.Context, db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error) {
	return db.UpdateRombelSubjectAssignmentRow{}, nil
}
func (noopRombelStore) CountRombelSubjectAssignmentDependents(context.Context, db.CountRombelSubjectAssignmentDependentsParams) (db.CountRombelSubjectAssignmentDependentsRow, error) {
	return db.CountRombelSubjectAssignmentDependentsRow{}, nil
}
func (noopRombelStore) DeleteRombelSubjectAssignment(context.Context, db.DeleteRombelSubjectAssignmentParams) (int64, error) {
	return 0, nil
}
func (noopRombelStore) ListRombelTimetableSlots(context.Context, pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error) {
	return nil, nil
}
func (noopRombelStore) GetRombelTimetableSlot(context.Context, db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error) {
	return db.GetRombelTimetableSlotRow{}, nil
}
func (noopRombelStore) CreateRombelTimetableSlot(context.Context, db.CreateRombelTimetableSlotParams) (db.CreateRombelTimetableSlotRow, error) {
	return db.CreateRombelTimetableSlotRow{}, nil
}
func (noopRombelStore) UpdateRombelTimetableSlot(context.Context, db.UpdateRombelTimetableSlotParams) (db.UpdateRombelTimetableSlotRow, error) {
	return db.UpdateRombelTimetableSlotRow{}, nil
}
func (noopRombelStore) DeleteRombelTimetableSlot(context.Context, db.DeleteRombelTimetableSlotParams) (int64, error) {
	return 0, nil
}
func (noopRombelStore) LockTimetableMutationScope(context.Context, string) (int64, error) {
	return 0, nil
}
func (noopRombelStore) CountTimetableConflicts(context.Context, db.CountTimetableConflictsParams) (int32, error) {
	return 0, nil
}
func (noopRombelStore) CountTimetableRoomConflicts(context.Context, db.CountTimetableRoomConflictsParams) (int32, error) {
	return 0, nil
}
func (noopRombelStore) ListHomeroomAssignmentsByClass(context.Context, pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error) {
	return nil, nil
}
func (noopRombelStore) CreateHomeroomAssignment(context.Context, db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error) {
	return db.CreateHomeroomAssignmentRow{}, nil
}
func (noopRombelStore) UpdateHomeroomAssignment(context.Context, db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error) {
	return db.UpdateHomeroomAssignmentRow{}, nil
}
func (noopRombelStore) DeleteHomeroomAssignment(context.Context, pgtype.UUID) error {
	return nil
}

type fakeRombelIdentityStore struct {
	noopRombelStore
	detail       db.GetRombelDetailRow
	detailErr    error
	conflictArg  db.CountRombelCodeConflictsParams
	conflicts    int32
	conflictErr  error
	updateArg    db.UpdateRombelIdentityParams
	updateCalled bool
	updateErr    error
}

func (f *fakeRombelIdentityStore) GetRombelDetail(context.Context, pgtype.UUID) (db.GetRombelDetailRow, error) {
	if f.detailErr != nil {
		return db.GetRombelDetailRow{}, f.detailErr
	}
	return f.detail, nil
}

func (f *fakeRombelIdentityStore) CountRombelCodeConflicts(_ context.Context, arg db.CountRombelCodeConflictsParams) (int32, error) {
	f.conflictArg = arg
	return f.conflicts, f.conflictErr
}

func (f *fakeRombelIdentityStore) UpdateRombelIdentity(_ context.Context, arg db.UpdateRombelIdentityParams) (db.UpdateRombelIdentityRow, error) {
	f.updateCalled = true
	f.updateArg = arg
	if f.updateErr != nil {
		return db.UpdateRombelIdentityRow{}, f.updateErr
	}
	return db.UpdateRombelIdentityRow{
		ID:       arg.ID,
		Code:     arg.Code,
		Name:     arg.Name,
		Level:    arg.Level,
		IsActive: arg.IsActive,
	}, nil
}

func TestRombelUpdateIdentityValidatesAndUpdates(t *testing.T) {
	classID := documentCycleTestUUID(201)
	store := &fakeRombelIdentityStore{
		detail: db.GetRombelDetailRow{ID: classID, IsActive: true, TotalStudents: 0},
	}
	svc := &Rombel{q: store}

	row, err := svc.UpdateIdentity(context.Background(), db.UpdateRombelIdentityParams{
		ID:       classID,
		Code:     " VII.A ",
		Name:     " VII A ",
		Level:    "vii",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("UpdateIdentity() error = %v, want nil", err)
	}
	if row.Code != "VII.A" || row.Name != "VII A" || row.Level != "VII" || !row.IsActive {
		t.Fatalf("UpdateIdentity() row = %+v, want trimmed and normalized identity", row)
	}
	if store.conflictArg.ID != classID || store.conflictArg.Code != "VII.A" {
		t.Fatalf("conflict arg = %+v, want class/code", store.conflictArg)
	}
	if !store.updateCalled || store.updateArg.Code != "VII.A" || store.updateArg.Name != "VII A" || store.updateArg.Level != "VII" {
		t.Fatalf("update arg = %+v called=%v, want normalized update", store.updateArg, store.updateCalled)
	}
}

func TestRombelUpdateIdentityRejectsInvalidAndUnsafeChanges(t *testing.T) {
	classID := documentCycleTestUUID(202)
	tests := []struct {
		name    string
		store   *fakeRombelIdentityStore
		arg     db.UpdateRombelIdentityParams
		wantErr error
	}{
		{
			name:    "empty code",
			store:   &fakeRombelIdentityStore{},
			arg:     db.UpdateRombelIdentityParams{ID: classID, Code: " ", Name: "VII A", Level: "VII", IsActive: true},
			wantErr: domain.ErrBadRequest,
		},
		{
			name:    "invalid level",
			store:   &fakeRombelIdentityStore{},
			arg:     db.UpdateRombelIdentityParams{ID: classID, Code: "VII.A", Name: "VII A", Level: "X", IsActive: true},
			wantErr: domain.ErrBadRequest,
		},
		{
			name:    "deactivate with active students",
			store:   &fakeRombelIdentityStore{detail: db.GetRombelDetailRow{ID: classID, IsActive: true, TotalStudents: 3}},
			arg:     db.UpdateRombelIdentityParams{ID: classID, Code: "VII.A", Name: "VII A", Level: "VII", IsActive: false},
			wantErr: domain.ErrConflict,
		},
		{
			name:    "duplicate code",
			store:   &fakeRombelIdentityStore{detail: db.GetRombelDetailRow{ID: classID, IsActive: true}, conflicts: 1},
			arg:     db.UpdateRombelIdentityParams{ID: classID, Code: "VII.A", Name: "VII A", Level: "VII", IsActive: true},
			wantErr: domain.ErrConflict,
		},
	}
	for _, tt := range tests {
		svc := &Rombel{q: tt.store}
		_, err := svc.UpdateIdentity(context.Background(), tt.arg)
		if !errors.Is(err, tt.wantErr) {
			t.Fatalf("%s: error = %v, want %v", tt.name, err, tt.wantErr)
		}
		if tt.name != "duplicate code" && tt.store.updateCalled {
			t.Fatalf("%s: update called for rejected change", tt.name)
		}
	}
}
