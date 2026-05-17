package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type beginErrProfileChangeTx struct{ err error }

func (b beginErrProfileChangeTx) Begin(ctx context.Context) (pgx.Tx, error) { return nil, b.err }

func TestProfileChangeRequestWithStoreUsesDirectStoreAndPropagatesBeginError(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	svc := &ProfileChangeRequest{q: store}
	called := false
	if err := svc.withStore(context.Background(), func(got profileChangeRequestStore) error {
		called = true
		if got != store {
			t.Fatalf("withStore direct store = %T, want fake store", got)
		}
		return domain.ErrBadRequest
	}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("withStore direct error = %v, want callback error", err)
	}
	if !called {
		t.Fatal("withStore direct did not call callback")
	}

	beginErr := errors.New("begin failed")
	svc = &ProfileChangeRequest{q: store, tx: beginErrProfileChangeTx{err: beginErr}}
	called = false
	if err := svc.withStore(context.Background(), func(profileChangeRequestStore) error {
		called = true
		return nil
	}); !errors.Is(err, beginErr) {
		t.Fatalf("withStore begin error = %v, want %v", err, beginErr)
	}
	if called {
		t.Fatal("withStore called callback after begin error")
	}
}

func TestApplyApprovedProfileChangeRequestCoversOfficialFieldsAndValidation(t *testing.T) {
	store := newFakeProfileChangeRequestStore()
	ctx := context.Background()
	employeeID := documentCycleTestUUID(71)
	studentID := documentCycleTestUUID(72)
	parentID := documentCycleTestUUID(73)

	tests := []struct {
		name string
		row  db.ProfileChangeRequest
		want func(t *testing.T)
	}{
		{name: "employee name", row: db.ProfileChangeRequest{ProfileType: "employee", TargetEmployeeID: employeeID, FieldKey: "nama", RequestedValue: "Guru Baru"}, want: func(t *testing.T) {
			if store.employeeName != "Guru Baru" {
				t.Fatalf("employeeName = %q", store.employeeName)
			}
		}},
		{name: "employee birthdate", row: db.ProfileChangeRequest{ProfileType: "employee", TargetEmployeeID: employeeID, FieldKey: "tanggal_lahir", RequestedValue: "1990-05-17"}, want: func(t *testing.T) {
			if !store.employeeBirthdate.Valid || store.employeeBirthdate.Time.Format("2006-01-02") != "1990-05-17" {
				t.Fatalf("employeeBirthdate = %+v", store.employeeBirthdate)
			}
		}},
		{name: "student name", row: db.ProfileChangeRequest{ProfileType: "student", TargetStudentID: studentID, FieldKey: "nama", RequestedValue: "Siswa Baru"}, want: func(t *testing.T) {
			if store.studentName != "Siswa Baru" {
				t.Fatalf("studentName = %q", store.studentName)
			}
		}},
		{name: "student birthdate", row: db.ProfileChangeRequest{ProfileType: "student", TargetStudentID: studentID, FieldKey: "tanggal_lahir", RequestedValue: "2011-01-02"}, want: func(t *testing.T) {
			if !store.studentBirthdate.Valid || store.studentBirthdate.Time.Format("2006-01-02") != "2011-01-02" {
				t.Fatalf("studentBirthdate = %+v", store.studentBirthdate)
			}
		}},
		{name: "student address", row: db.ProfileChangeRequest{ProfileType: "student", TargetStudentID: studentID, FieldKey: "alamat", RequestedValue: "Alamat Baru"}, want: func(t *testing.T) {
			if store.studentAddress != "Alamat Baru" {
				t.Fatalf("studentAddress = %q", store.studentAddress)
			}
		}},
		{name: "parent name", row: db.ProfileChangeRequest{ProfileType: "parent", TargetParentID: parentID, FieldKey: "nama", RequestedValue: "Wali Baru"}, want: func(t *testing.T) {
			if store.parentName != "Wali Baru" {
				t.Fatalf("parentName = %q", store.parentName)
			}
		}},
		{name: "parent phone", row: db.ProfileChangeRequest{ProfileType: "parent", TargetParentID: parentID, FieldKey: "phone", RequestedValue: "0812999"}, want: func(t *testing.T) {
			if store.parentPhone != "0812999" {
				t.Fatalf("parentPhone = %q", store.parentPhone)
			}
		}},
		{name: "parent address", row: db.ProfileChangeRequest{ProfileType: "parent", TargetParentID: parentID, FieldKey: "address", RequestedValue: "Jalan Baru"}, want: func(t *testing.T) {
			if store.parentAddress != "Jalan Baru" {
				t.Fatalf("parentAddress = %q", store.parentAddress)
			}
		}},
		{name: "parent nik", row: db.ProfileChangeRequest{ProfileType: "parent", TargetParentID: parentID, FieldKey: "nik", RequestedValue: "1234567890123456"}, want: func(t *testing.T) {
			if store.parentNik != "1234567890123456" {
				t.Fatalf("parentNik = %q", store.parentNik)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyApprovedProfileChangeRequest(ctx, store, tt.row); err != nil {
				t.Fatalf("applyApprovedProfileChangeRequest() error = %v", err)
			}
			tt.want(t)
		})
	}

	if err := applyApprovedProfileChangeRequest(ctx, store, db.ProfileChangeRequest{ProfileType: "employee", FieldKey: "nama", RequestedValue: "No ID"}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("missing target error = %v, want bad request", err)
	}
	if err := applyApprovedProfileChangeRequest(ctx, store, db.ProfileChangeRequest{ProfileType: "student", TargetStudentID: studentID, FieldKey: "tanggal_lahir", RequestedValue: "bad-date"}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("bad value error = %v, want bad request", err)
	}
	if err := applyApprovedProfileChangeRequest(ctx, store, db.ProfileChangeRequest{ProfileType: "unknown", FieldKey: "nama", RequestedValue: "X"}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("unknown profile error = %v, want bad request", err)
	}
}

type zeroAffectedProfileChangeStore struct{ *fakeProfileChangeRequestStore }

func (z zeroAffectedProfileChangeStore) UpdateStudentOfficialPhone(ctx context.Context, arg db.UpdateStudentOfficialPhoneParams) (int64, error) {
	return 0, nil
}

func TestApplyApprovedProfileChangeRequestMapsZeroAffectedToNotFound(t *testing.T) {
	store := zeroAffectedProfileChangeStore{newFakeProfileChangeRequestStore()}
	err := applyApprovedProfileChangeRequest(context.Background(), store, db.ProfileChangeRequest{
		ProfileType:     "student",
		TargetStudentID: documentCycleTestUUID(74),
		FieldKey:        "phone",
		RequestedValue:  "0812000",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("zero affected error = %v, want not found", err)
	}
}

var _ profileChangeRequestStore = zeroAffectedProfileChangeStore{}
