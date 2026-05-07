package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type rombelStore interface {
	ListRombels(ctx context.Context) ([]db.ListRombelsRow, error)
	GetRombelDetail(ctx context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error)
	ListStudentsByClassWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error)
	ListRombelSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error)
	GetRombelSubjectAssignment(ctx context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error)
	CreateRombelSubjectAssignment(ctx context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error)
	UpdateRombelSubjectAssignment(ctx context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error)
	CountRombelSubjectAssignmentDependents(ctx context.Context, arg db.CountRombelSubjectAssignmentDependentsParams) (db.CountRombelSubjectAssignmentDependentsRow, error)
	DeleteRombelSubjectAssignment(ctx context.Context, arg db.DeleteRombelSubjectAssignmentParams) (int64, error)
	ListRombelTimetableSlots(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error)
	GetRombelTimetableSlot(ctx context.Context, arg db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error)
	CreateRombelTimetableSlot(ctx context.Context, arg db.CreateRombelTimetableSlotParams) (db.CreateRombelTimetableSlotRow, error)
	UpdateRombelTimetableSlot(ctx context.Context, arg db.UpdateRombelTimetableSlotParams) (db.UpdateRombelTimetableSlotRow, error)
	DeleteRombelTimetableSlot(ctx context.Context, arg db.DeleteRombelTimetableSlotParams) (int64, error)
	LockTimetableMutationScope(ctx context.Context, lockKey string) (int64, error)
	CountTimetableConflicts(ctx context.Context, arg db.CountTimetableConflictsParams) (int32, error)
	CountTimetableRoomConflicts(ctx context.Context, arg db.CountTimetableRoomConflictsParams) (int32, error)
	ListHomeroomAssignmentsByClass(ctx context.Context, classID pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error)
	CreateHomeroomAssignment(ctx context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error)
	UpdateHomeroomAssignment(ctx context.Context, arg db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error)
	DeleteHomeroomAssignment(ctx context.Context, id pgtype.UUID) error
}

type Rombel struct {
	q  rombelStore
	tx classJournalTxStarter
}

func NewRombel(q *db.Queries) *Rombel { return &Rombel{q: q} }

func NewRombelWithPool(pool *pgxpool.Pool) *Rombel {
	if pool == nil {
		return &Rombel{q: db.New(nil)}
	}
	return &Rombel{q: db.New(pool), tx: pool}
}

func (s *Rombel) List(ctx context.Context) ([]db.ListRombelsRow, error) {
	return s.q.ListRombels(ctx)
}

func (s *Rombel) Get(ctx context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error) {
	return s.q.GetRombelDetail(ctx, id)
}

func (s *Rombel) ListStudentsWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error) {
	return s.q.ListStudentsByClassWithParents(ctx, classID)
}

func (s *Rombel) ListSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error) {
	return s.q.ListRombelSubjectAssignments(ctx, classID)
}

func (s *Rombel) GetSubjectAssignment(ctx context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error) {
	return s.q.GetRombelSubjectAssignment(ctx, arg)
}

func (s *Rombel) CreateSubjectAssignment(ctx context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error) {
	return s.q.CreateRombelSubjectAssignment(ctx, arg)
}

func (s *Rombel) UpdateSubjectAssignment(ctx context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error) {
	current, err := s.q.GetRombelSubjectAssignment(ctx, db.GetRombelSubjectAssignmentParams{
		ClassID: arg.ClassID,
		ID:      arg.ID,
	})
	if err != nil {
		return db.UpdateRombelSubjectAssignmentRow{}, err
	}
	if !sameRombelUUID(current.SubjectID, arg.SubjectID) {
		counts, err := s.q.CountRombelSubjectAssignmentDependents(ctx, db.CountRombelSubjectAssignmentDependentsParams{
			ClassID: arg.ClassID,
			ID:      arg.ID,
		})
		if err != nil {
			return db.UpdateRombelSubjectAssignmentRow{}, err
		}
		if counts.TotalTimetableSlots > 0 || counts.TotalJournalSessions > 0 || counts.TotalGradeComponents > 0 || counts.TotalGradeFinalizations > 0 {
			return db.UpdateRombelSubjectAssignmentRow{}, fmt.Errorf("%w: mata pelajaran tidak dapat diganti karena penugasan sudah dipakai oleh jadwal, jurnal, atau nilai", domain.ErrConflict)
		}
	}
	return s.q.UpdateRombelSubjectAssignment(ctx, arg)
}

func (s *Rombel) DeleteSubjectAssignment(ctx context.Context, arg db.DeleteRombelSubjectAssignmentParams) error {
	counts, err := s.q.CountRombelSubjectAssignmentDependents(ctx, db.CountRombelSubjectAssignmentDependentsParams{
		ClassID: arg.ClassID,
		ID:      arg.ID,
	})
	if err != nil {
		return err
	}
	if counts.TotalTimetableSlots > 0 || counts.TotalJournalSessions > 0 || counts.TotalGradeComponents > 0 || counts.TotalGradeFinalizations > 0 {
		return fmt.Errorf("%w: penugasan guru mapel masih dipakai oleh jadwal, jurnal, atau nilai", domain.ErrConflict)
	}
	rows, err := s.q.DeleteRombelSubjectAssignment(ctx, arg)
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func sameRombelUUID(a, b pgtype.UUID) bool {
	if a.Valid != b.Valid {
		return false
	}
	return a.Bytes == b.Bytes
}

func (s *Rombel) ListTimetableSlots(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error) {
	return s.q.ListRombelTimetableSlots(ctx, classID)
}

func (s *Rombel) GetTimetableSlot(ctx context.Context, arg db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error) {
	return s.q.GetRombelTimetableSlot(ctx, arg)
}

func (s *Rombel) CreateTimetableSlot(ctx context.Context, arg db.CreateRombelTimetableSlotParams) (db.CreateRombelTimetableSlotRow, error) {
	var row db.CreateRombelTimetableSlotRow
	err := s.withRombelStore(ctx, func(store rombelStore) error {
		if err := s.ensureTimetableSlotAvailable(ctx, store, arg.ClassID, arg.AssignmentID, arg.DayOfWeek, arg.StartTime, arg.EndTime, arg.RoomLabel, pgtype.UUID{}); err != nil {
			return err
		}
		var err error
		row, err = store.CreateRombelTimetableSlot(ctx, arg)
		return err
	})
	return row, err
}

func (s *Rombel) UpdateTimetableSlot(ctx context.Context, arg db.UpdateRombelTimetableSlotParams) (db.UpdateRombelTimetableSlotRow, error) {
	var row db.UpdateRombelTimetableSlotRow
	err := s.withRombelStore(ctx, func(store rombelStore) error {
		if _, err := store.GetRombelTimetableSlot(ctx, db.GetRombelTimetableSlotParams{
			ClassID: arg.ClassID,
			ID:      arg.ID,
		}); err != nil {
			return err
		}
		if err := s.ensureTimetableSlotAvailable(ctx, store, arg.ClassID, arg.AssignmentID, arg.DayOfWeek, arg.StartTime, arg.EndTime, arg.RoomLabel, arg.ID); err != nil {
			return err
		}
		var err error
		row, err = store.UpdateRombelTimetableSlot(ctx, arg)
		return err
	})
	return row, err
}

func (s *Rombel) DeleteTimetableSlot(ctx context.Context, arg db.DeleteRombelTimetableSlotParams) error {
	rows, err := s.q.DeleteRombelTimetableSlot(ctx, arg)
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Rombel) ensureTimetableSlotAvailable(ctx context.Context, store rombelStore, classID, assignmentID pgtype.UUID, dayOfWeek int16, startTime, endTime pgtype.Time, roomLabel string, excludeSlotID pgtype.UUID) error {
	assignment, err := store.GetRombelSubjectAssignment(ctx, db.GetRombelSubjectAssignmentParams{
		ClassID: classID,
		ID:      assignmentID,
	})
	if err != nil {
		return err
	}
	if err := lockTimetableMutationScopes(ctx, store, dayOfWeek, assignment.ClassID, assignment.TeacherEmployeeID, roomLabel); err != nil {
		return err
	}
	conflicts, err := store.CountTimetableConflicts(ctx, db.CountTimetableConflictsParams{
		DayOfWeek:         dayOfWeek,
		StartTime:         startTime,
		EndTime:           endTime,
		ClassID:           assignment.ClassID,
		TeacherEmployeeID: assignment.TeacherEmployeeID,
		ExcludeSlotID:     excludeSlotID,
	})
	if err != nil {
		return err
	}
	if conflicts > 0 {
		return fmt.Errorf("%w: slot bentrok dengan jadwal kelas atau guru pada waktu yang sama", domain.ErrConflict)
	}
	roomConflicts, err := store.CountTimetableRoomConflicts(ctx, db.CountTimetableRoomConflictsParams{
		DayOfWeek:     dayOfWeek,
		StartTime:     startTime,
		EndTime:       endTime,
		RoomLabel:     strings.TrimSpace(roomLabel),
		ExcludeSlotID: excludeSlotID,
	})
	if err != nil {
		return err
	}
	if roomConflicts > 0 {
		return fmt.Errorf("%w: slot bentrok dengan penggunaan ruang pada waktu yang sama", domain.ErrConflict)
	}
	return nil
}

func (s *Rombel) withRombelStore(ctx context.Context, fn func(rombelStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type timetableLockStore interface {
	LockTimetableMutationScope(ctx context.Context, lockKey string) (int64, error)
}

func lockTimetableMutationScopes(ctx context.Context, store timetableLockStore, dayOfWeek int16, classID, teacherID pgtype.UUID, roomLabel string) error {
	keys := []string{
		fmt.Sprintf("timetable:day:%d:class:%s", dayOfWeek, classID.String()),
		fmt.Sprintf("timetable:day:%d:teacher:%s", dayOfWeek, teacherID.String()),
	}
	if room := strings.ToLower(strings.TrimSpace(roomLabel)); room != "" {
		keys = append(keys, fmt.Sprintf("timetable:day:%d:room:%s", dayOfWeek, room))
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := store.LockTimetableMutationScope(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Rombel) ListHomeroomAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error) {
	return s.q.ListHomeroomAssignmentsByClass(ctx, classID)
}

func (s *Rombel) CreateHomeroomAssignment(ctx context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error) {
	return s.q.CreateHomeroomAssignment(ctx, arg)
}

func (s *Rombel) UpdateHomeroomAssignment(ctx context.Context, arg db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error) {
	return s.q.UpdateHomeroomAssignment(ctx, arg)
}

func (s *Rombel) DeleteHomeroomAssignment(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteHomeroomAssignment(ctx, id)
}
