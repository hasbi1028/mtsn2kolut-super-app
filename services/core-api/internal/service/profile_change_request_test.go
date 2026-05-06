package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeProfileChangeRequestStore struct {
	employeeProfile db.GetOwnedEmployeeOfficialProfileRow
	studentProfile  db.GetOwnedStudentOfficialProfileRow
	parentProfile   db.GetOwnedParentOfficialProfileRow
	employeeErr     error
	studentErr      error
	parentErr       error

	requests map[pgtype.UUID]db.ProfileChangeRequest
	audits   []db.CreateAuditLogParams
	listArg  db.ListProfileChangeRequestsParams
	countArg db.CountProfileChangeRequestsParams

	employeeName      string
	employeeBirthdate pgtype.Date
	studentName       string
	studentBirthdate  pgtype.Date
	studentParentName string
	parentName        string
}

func newFakeProfileChangeRequestStore() *fakeProfileChangeRequestStore {
	return &fakeProfileChangeRequestStore{
		employeeErr: pgx.ErrNoRows,
		studentErr:  pgx.ErrNoRows,
		parentErr:   pgx.ErrNoRows,
		requests:    map[pgtype.UUID]db.ProfileChangeRequest{},
	}
}

func (f *fakeProfileChangeRequestStore) GetOwnedEmployeeOfficialProfile(ctx context.Context, id pgtype.UUID) (db.GetOwnedEmployeeOfficialProfileRow, error) {
	return f.employeeProfile, f.employeeErr
}

func (f *fakeProfileChangeRequestStore) GetOwnedStudentOfficialProfile(ctx context.Context, id pgtype.UUID) (db.GetOwnedStudentOfficialProfileRow, error) {
	return f.studentProfile, f.studentErr
}

func (f *fakeProfileChangeRequestStore) GetOwnedParentOfficialProfile(ctx context.Context, id pgtype.UUID) (db.GetOwnedParentOfficialProfileRow, error) {
	return f.parentProfile, f.parentErr
}

func (f *fakeProfileChangeRequestStore) CreateProfileChangeRequest(ctx context.Context, arg db.CreateProfileChangeRequestParams) (db.ProfileChangeRequest, error) {
	id := documentCycleTestUUID(byte(200 + len(f.requests)))
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	row := db.ProfileChangeRequest{
		ID:               id,
		RequesterUserID:  arg.RequesterUserID,
		ProfileType:      arg.ProfileType,
		TargetEmployeeID: arg.TargetEmployeeID,
		TargetStudentID:  arg.TargetStudentID,
		TargetParentID:   arg.TargetParentID,
		FieldKey:         arg.FieldKey,
		CurrentValue:     arg.CurrentValue,
		RequestedValue:   arg.RequestedValue,
		Reason:           arg.Reason,
		Status:           db.ProfileChangeRequestStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	f.requests[id] = row
	return row, nil
}

func (f *fakeProfileChangeRequestStore) ListOwnProfileChangeRequests(ctx context.Context, requesterUserID pgtype.UUID) ([]db.ListOwnProfileChangeRequestsRow, error) {
	return nil, nil
}

func (f *fakeProfileChangeRequestStore) ListProfileChangeRequests(ctx context.Context, arg db.ListProfileChangeRequestsParams) ([]db.ListProfileChangeRequestsRow, error) {
	f.listArg = arg
	return nil, nil
}

func (f *fakeProfileChangeRequestStore) CountProfileChangeRequests(ctx context.Context, arg db.CountProfileChangeRequestsParams) (int32, error) {
	f.countArg = arg
	return 3, nil
}

func (f *fakeProfileChangeRequestStore) GetProfileChangeRequestForUpdate(ctx context.Context, id pgtype.UUID) (db.ProfileChangeRequest, error) {
	row, ok := f.requests[id]
	if !ok {
		return db.ProfileChangeRequest{}, pgx.ErrNoRows
	}
	return row, nil
}

func (f *fakeProfileChangeRequestStore) CancelOwnProfileChangeRequest(ctx context.Context, arg db.CancelOwnProfileChangeRequestParams) (db.ProfileChangeRequest, error) {
	row, ok := f.requests[arg.ID]
	if !ok || row.RequesterUserID != arg.RequesterUserID || row.Status != db.ProfileChangeRequestStatusPending {
		return db.ProfileChangeRequest{}, pgx.ErrNoRows
	}
	row.Status = db.ProfileChangeRequestStatusCancelled
	f.requests[arg.ID] = row
	return row, nil
}

func (f *fakeProfileChangeRequestStore) ReviewProfileChangeRequest(ctx context.Context, arg db.ReviewProfileChangeRequestParams) (db.ProfileChangeRequest, error) {
	row, ok := f.requests[arg.ID]
	if !ok || row.Status != db.ProfileChangeRequestStatusPending {
		return db.ProfileChangeRequest{}, pgx.ErrNoRows
	}
	row.Status = arg.Status
	row.ReviewerUserID = arg.ReviewerUserID
	row.ReviewNote = arg.ReviewNote
	row.ReviewedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	f.requests[arg.ID] = row
	return row, nil
}

func (f *fakeProfileChangeRequestStore) UpdateEmployeeOfficialName(ctx context.Context, arg db.UpdateEmployeeOfficialNameParams) (int64, error) {
	f.employeeName = arg.Nama
	return 1, nil
}

func (f *fakeProfileChangeRequestStore) UpdateEmployeeOfficialBirthdate(ctx context.Context, arg db.UpdateEmployeeOfficialBirthdateParams) (int64, error) {
	f.employeeBirthdate = arg.TanggalLahir
	return 1, nil
}

func (f *fakeProfileChangeRequestStore) UpdateStudentOfficialName(ctx context.Context, arg db.UpdateStudentOfficialNameParams) (int64, error) {
	f.studentName = arg.Nama
	return 1, nil
}

func (f *fakeProfileChangeRequestStore) UpdateStudentOfficialBirthdate(ctx context.Context, arg db.UpdateStudentOfficialBirthdateParams) (int64, error) {
	f.studentBirthdate = arg.TanggalLahir
	return 1, nil
}

func (f *fakeProfileChangeRequestStore) UpdateStudentOfficialParentName(ctx context.Context, arg db.UpdateStudentOfficialParentNameParams) (int64, error) {
	f.studentParentName = arg.ParentName
	return 1, nil
}

func (f *fakeProfileChangeRequestStore) UpdateParentOfficialName(ctx context.Context, arg db.UpdateParentOfficialNameParams) (int64, error) {
	f.parentName = arg.Nama
	return 1, nil
}

func (f *fakeProfileChangeRequestStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.audits = append(f.audits, arg)
	return db.AuditLog{}, nil
}

func TestProfileChangeRequestCreateUsesOwnedProfileAndConservativeAllowlist(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	userID := documentCycleTestUUID(11)
	employeeID := documentCycleTestUUID(12)
	store.employeeErr = nil
	store.employeeProfile = db.GetOwnedEmployeeOfficialProfileRow{
		ID:           employeeID,
		Nama:         "Nama Lama",
		TanggalLahir: documentCycleTestDate(1990, 5, 6),
	}
	svc := &ProfileChangeRequest{q: store}

	row, err := svc.Create(context.Background(), userID, CreateProfileChangeRequestInput{
		FieldKey:       " name ",
		RequestedValue: " Nama Baru ",
		Reason:         "Sesuai dokumen resmi",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if row.ProfileType != "employee" || row.TargetEmployeeID != employeeID || row.FieldKey != "nama" {
		t.Fatalf("created row = %+v, want employee nama request", row)
	}
	if row.CurrentValue != "Nama Lama" || row.RequestedValue != "Nama Baru" {
		t.Fatalf("values = %q/%q, want trimmed current/requested", row.CurrentValue, row.RequestedValue)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "ACCOUNT_CHANGE_REQUEST_CREATED" {
		t.Fatalf("audits = %#v, want create audit", store.audits)
	}

	for _, fieldKey := range []string{"nip", "nisn", "nik", "class_id", "username", "role", "status"} {
		_, err = svc.Create(context.Background(), userID, CreateProfileChangeRequestInput{
			FieldKey:       fieldKey,
			RequestedValue: "123",
			Reason:         "Tidak boleh",
		})
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("Create(%s) error = %v, want bad request", fieldKey, err)
		}
	}

	_, err = svc.Create(context.Background(), userID, CreateProfileChangeRequestInput{
		FieldKey:       "tanggal_lahir",
		RequestedValue: "bad-date",
		Reason:         "Koreksi tanggal",
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Create(bad birthdate) error = %v, want bad request", err)
	}
}

func TestProfileChangeRequestListSelfRequestableFieldsUsesCatalog(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	userID := documentCycleTestUUID(21)
	store.studentErr = nil
	store.studentProfile = db.GetOwnedStudentOfficialProfileRow{
		ID:           documentCycleTestUUID(22),
		Nama:         "Nama Siswa",
		TanggalLahir: documentCycleTestDate(2010, 7, 8),
		ParentName:   "Nama Wali",
	}
	svc := &ProfileChangeRequest{q: store}

	fields, err := svc.ListSelfRequestableFields(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListSelfRequestableFields() error = %v", err)
	}
	got := make([]string, 0, len(fields))
	for _, field := range fields {
		got = append(got, field.ProfileType+":"+field.FieldKey+":"+field.ValueType+":"+field.ReviewerPermission)
		if !field.SelfRequestable || !field.IsActive {
			t.Fatalf("field = %+v, want self requestable active field", field)
		}
	}
	want := []string{
		"student:nama:text:profile_changes.review",
		"student:tanggal_lahir:date:profile_changes.review",
		"student:parent_name:text:profile_changes.review",
	}
	if len(got) != len(want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("fields = %v, want %v", got, want)
		}
	}

	store = newFakeProfileChangeRequestStore()
	svc = &ProfileChangeRequest{q: store}
	fields, err = svc.ListSelfRequestableFields(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListSelfRequestableFields(unlinked) error = %v", err)
	}
	if len(fields) != 0 {
		t.Fatalf("fields = %+v, want none for unlinked account", fields)
	}
}

func TestProfileChangeRequestAdminListFiltersAreNormalized(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	svc := &ProfileChangeRequest{q: store}

	_, err := svc.ListAdmin(context.Background(), ProfileChangeRequestListFilter{
		Status:      " Pending ",
		ProfileType: " Student ",
		FieldKey:    "birth-date",
		Search:      "  Guru   IPA  ",
	}, 25, 50)
	if err != nil {
		t.Fatalf("ListAdmin() error = %v", err)
	}
	if store.listArg.StatusFilter != "pending" || store.listArg.ProfileTypeFilter != "student" || store.listArg.FieldKeyFilter != "tanggal_lahir" || store.listArg.Search != "Guru IPA" || store.listArg.LimitCount != 25 || store.listArg.OffsetCount != 50 {
		t.Fatalf("list arg = %+v, want normalized filters", store.listArg)
	}

	count, err := svc.CountAdmin(context.Background(), ProfileChangeRequestListFilter{Status: "all", ProfileType: "employee", FieldKey: "name"})
	if err != nil {
		t.Fatalf("CountAdmin() error = %v", err)
	}
	if count != 3 || store.countArg.StatusFilter != "" || store.countArg.ProfileTypeFilter != "employee" || store.countArg.FieldKeyFilter != "nama" {
		t.Fatalf("count=%d arg=%+v, want normalized count filters", count, store.countArg)
	}

	if _, err := svc.ListAdmin(context.Background(), ProfileChangeRequestListFilter{Status: "unknown"}, 10, 0); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("ListAdmin(invalid status) error = %v, want bad request", err)
	}
	if _, err := svc.CountAdmin(context.Background(), ProfileChangeRequestListFilter{ProfileType: "teacher"}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("CountAdmin(invalid profile) error = %v, want bad request", err)
	}
}

func TestProfileChangeRequestReviewApprovesAllowedFieldAndAudits(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	svc := &ProfileChangeRequest{q: store}
	requestID := documentCycleTestUUID(31)
	requesterID := documentCycleTestUUID(32)
	reviewerID := documentCycleTestUUID(33)
	studentID := documentCycleTestUUID(34)
	store.requests[requestID] = db.ProfileChangeRequest{
		ID:              requestID,
		RequesterUserID: requesterID,
		ProfileType:     "student",
		TargetStudentID: studentID,
		FieldKey:        "parent_name",
		CurrentValue:    "Wali Lama",
		RequestedValue:  "Wali Baru",
		Reason:          "Akte keluarga",
		Status:          db.ProfileChangeRequestStatusPending,
	}

	row, err := svc.Review(context.Background(), reviewerID, requestID, ReviewProfileChangeRequestInput{
		Status:              "approved",
		ReviewNote:          "Dokumen cocok",
		ReviewerPermissions: []string{ProfileChangesReviewPermission},
	})
	if err != nil {
		t.Fatalf("Review(approved) error = %v", err)
	}
	if row.Status != db.ProfileChangeRequestStatusApproved || store.studentParentName != "Wali Baru" {
		t.Fatalf("approved row/status = %+v, studentParentName=%q", row, store.studentParentName)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "ACCOUNT_CHANGE_REQUEST_APPROVED" {
		t.Fatalf("audits = %#v, want approve audit", store.audits)
	}
	var meta map[string]any
	if err := json.Unmarshal(store.audits[0].Metadata, &meta); err != nil {
		t.Fatalf("audit metadata json = %v", err)
	}
	if meta["field_key"] != "parent_name" || meta["status"] != "approved" {
		t.Fatalf("audit metadata = %+v, want field/status", meta)
	}
}

func TestProfileChangeRequestReviewRejectsUnsupportedFieldAndMissingReviewerPermission(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	svc := &ProfileChangeRequest{q: store}
	requestID := documentCycleTestUUID(35)
	reviewerID := documentCycleTestUUID(36)
	store.requests[requestID] = db.ProfileChangeRequest{
		ID:               requestID,
		RequesterUserID:  documentCycleTestUUID(37),
		ProfileType:      "employee",
		TargetEmployeeID: documentCycleTestUUID(38),
		FieldKey:         "nip",
		CurrentValue:     "",
		RequestedValue:   "123",
		Status:           db.ProfileChangeRequestStatusPending,
	}
	if _, err := svc.Review(context.Background(), reviewerID, requestID, ReviewProfileChangeRequestInput{
		Status:              "approved",
		ReviewerPermissions: []string{ProfileChangesReviewPermission},
	}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Review(unsupported field) error = %v, want bad request", err)
	}

	store.requests[requestID] = db.ProfileChangeRequest{
		ID:               requestID,
		RequesterUserID:  documentCycleTestUUID(37),
		ProfileType:      "employee",
		TargetEmployeeID: documentCycleTestUUID(38),
		FieldKey:         "nama",
		CurrentValue:     "Lama",
		RequestedValue:   "Baru",
		Status:           db.ProfileChangeRequestStatusPending,
	}
	if _, err := svc.Review(context.Background(), reviewerID, requestID, ReviewProfileChangeRequestInput{Status: "approved"}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Review(missing reviewer permission) error = %v, want forbidden", err)
	}
}

func TestProfileChangeRequestReviewRejectRequiresNoteAndPendingStatus(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	svc := &ProfileChangeRequest{q: store}
	requestID := documentCycleTestUUID(41)
	reviewerID := documentCycleTestUUID(42)
	store.requests[requestID] = db.ProfileChangeRequest{
		ID:              requestID,
		RequesterUserID: documentCycleTestUUID(43),
		ProfileType:     "parent",
		TargetParentID:  documentCycleTestUUID(44),
		FieldKey:        "nama",
		CurrentValue:    "Nama Lama",
		RequestedValue:  "Nama Baru",
		Status:          db.ProfileChangeRequestStatusPending,
	}

	if _, err := svc.Review(context.Background(), reviewerID, requestID, ReviewProfileChangeRequestInput{Status: "rejected", ReviewerRoles: []string{"admin"}}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Review(reject without note) error = %v, want bad request", err)
	}

	row, err := svc.Review(context.Background(), reviewerID, requestID, ReviewProfileChangeRequestInput{Status: "rejected", ReviewNote: "Dokumen tidak cukup", ReviewerRoles: []string{"admin"}})
	if err != nil {
		t.Fatalf("Review(rejected) error = %v", err)
	}
	if row.Status != db.ProfileChangeRequestStatusRejected || store.parentName != "" {
		t.Fatalf("rejected row=%+v parentName=%q, want no profile mutation", row, store.parentName)
	}
	if _, err := svc.Review(context.Background(), reviewerID, requestID, ReviewProfileChangeRequestInput{Status: "approved", ReviewerRoles: []string{"admin"}}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Review(non-pending) error = %v, want conflict", err)
	}
}

func TestProfileChangeRequestCancelOwnPendingOnly(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	svc := &ProfileChangeRequest{q: store}
	requestID := documentCycleTestUUID(51)
	requesterID := documentCycleTestUUID(52)
	store.requests[requestID] = db.ProfileChangeRequest{
		ID:               requestID,
		RequesterUserID:  requesterID,
		ProfileType:      "employee",
		TargetEmployeeID: documentCycleTestUUID(53),
		FieldKey:         "nama",
		Status:           db.ProfileChangeRequestStatusPending,
	}

	row, err := svc.CancelOwn(context.Background(), requesterID, requestID)
	if err != nil {
		t.Fatalf("CancelOwn() error = %v", err)
	}
	if row.Status != db.ProfileChangeRequestStatusCancelled {
		t.Fatalf("status = %s, want cancelled", row.Status)
	}
	if len(store.audits) != 1 || store.audits[0].Action != "ACCOUNT_CHANGE_REQUEST_CANCELLED" {
		t.Fatalf("audits = %#v, want cancel audit", store.audits)
	}
	if _, err := svc.CancelOwn(context.Background(), requesterID, requestID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("CancelOwn(non-pending) error = %v, want not found", err)
	}
}
