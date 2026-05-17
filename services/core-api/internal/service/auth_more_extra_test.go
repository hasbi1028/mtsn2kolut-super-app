package service

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func authMoreTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed}, Valid: true}
}

func TestValidateAccountAvatarRejectsEdgeInputs(t *testing.T) {
	validID := authMoreTestUUID(201)
	base := UploadAccountAvatarInput{UserID: validID, Ext: ".png", MimeType: "image/png", FileSize: 1, File: strings.NewReader("x")}

	cases := []struct {
		name string
		in   UploadAccountAvatarInput
		want error
	}{
		{name: "invalid user", in: func() UploadAccountAvatarInput { in := base; in.UserID = pgtype.UUID{}; return in }(), want: domain.ErrUnauthorized},
		{name: "nil file", in: func() UploadAccountAvatarInput { in := base; in.File = nil; return in }()},
		{name: "zero size", in: func() UploadAccountAvatarInput { in := base; in.FileSize = 0; return in }()},
		{name: "too large", in: func() UploadAccountAvatarInput { in := base; in.FileSize = maxAccountAvatarBytes + 1; return in }()},
		{name: "unknown extension", in: func() UploadAccountAvatarInput { in := base; in.Ext = ".gif"; in.MimeType = "image/gif"; return in }()},
		{name: "mime mismatch", in: func() UploadAccountAvatarInput { in := base; in.Ext = ".jpg"; in.MimeType = "image/png"; return in }()},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAccountAvatar(tt.in)
			if err == nil {
				t.Fatal("validateAccountAvatar() error = nil, want error")
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("validateAccountAvatar() error = %v, want %v", err, tt.want)
			}
		})
	}

	if err := validateAccountAvatar(func() UploadAccountAvatarInput { in := base; in.Ext = " .WEBP "; in.MimeType = "image/webp"; return in }()); err != nil {
		t.Fatalf("validateAccountAvatar(valid trimmed upper extension) error = %v", err)
	}
}

func TestAuthUpdateAccountContactEdgeValidation(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	employeeUserID := authMoreTestUUID(202)
	employeeID := authMoreTestUUID(203)
	store.users["employee"] = db.User{ID: employeeUserID, Username: "employee", EmployeeID: employeeID, IsActive: true}
	store.employeeContact[employeeID] = fakeAccountContact{phone: " 0812 ", email: "old@example.test", address: " Old address "}
	svc := &Auth{q: store, jwtSecret: []byte("secret")}

	phone := " 0813 "
	row, err := svc.UpdateAccountContact(ctx, employeeUserID, AccountContactPatch{Phone: &phone})
	if err != nil {
		t.Fatalf("UpdateAccountContact(employee phone only) error = %v", err)
	}
	if row.ContactPhone != "0813" || row.ContactEmail != "old@example.test" || row.ContactAddress != "Old address" {
		t.Fatalf("updated contact = phone %q email %q address %q, want phone patched and nil fields preserved/trimmed", row.ContactPhone, row.ContactEmail, row.ContactAddress)
	}

	badEmail := "not an email"
	if _, err := svc.UpdateAccountContact(ctx, employeeUserID, AccountContactPatch{Email: &badEmail}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("UpdateAccountContact(invalid employee email) = %v, want bad request", err)
	}

	studentUserID := authMoreTestUUID(204)
	studentID := authMoreTestUUID(205)
	store.users["student"] = db.User{ID: studentUserID, Username: "student", StudentID: studentID, IsActive: true}
	store.studentContact[studentID] = fakeAccountContact{phone: "1", address: "home"}
	studentEmail := "student@example.test"
	if _, err := svc.UpdateAccountContact(ctx, studentUserID, AccountContactPatch{Email: &studentEmail}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("UpdateAccountContact(student email patch) = %v, want bad request", err)
	}
	if got := store.studentContact[studentID]; got.phone != "1" || got.address != "home" {
		t.Fatalf("student contact mutated after rejected email: %+v", got)
	}
}

func TestAuthSaveAccountAvatarSuccessRemovesOldStoredAvatar(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	userID := authMoreTestUUID(206)
	employeeID := authMoreTestUUID(207)
	store.users["employee"] = db.User{ID: userID, Username: "employee", EmployeeID: employeeID, IsActive: true}
	store.employeeContact[employeeID] = fakeAccountContact{photoURL: accountAvatarURL("old.png")}
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/old.png", []byte("old"), 0o644); err != nil {
		t.Fatalf("write old avatar: %v", err)
	}
	svc := &Auth{q: store, jwtSecret: []byte("secret"), avatarDir: dir}

	row, err := svc.SaveAccountAvatar(ctx, UploadAccountAvatarInput{UserID: userID, Ext: ".PNG", MimeType: "image/png", FileSize: 3, File: bytes.NewReader([]byte("new"))})
	if err != nil {
		t.Fatalf("SaveAccountAvatar() error = %v", err)
	}
	if row.PhotoUrl == "" || row.PhotoUrl == accountAvatarURL("old.png") {
		t.Fatalf("PhotoUrl = %q, want new account avatar URL", row.PhotoUrl)
	}
	if _, err := os.Stat(dir + "/old.png"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old avatar stat err = %v, want removed", err)
	}
	filename, ok := accountAvatarFilename(row.PhotoUrl)
	if !ok {
		t.Fatalf("new PhotoUrl %q did not parse as account avatar", row.PhotoUrl)
	}
	content, err := os.ReadFile(dir + "/" + filename)
	if err != nil || string(content) != "new" {
		t.Fatalf("new avatar content = %q err=%v, want stored upload", content, err)
	}
}

func TestAuthSaveAccountAvatarCleansNewFileWhenAvatarUpdateFails(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	userID := authMoreTestUUID(208)
	studentID := authMoreTestUUID(209)
	store.users["student"] = db.User{ID: userID, Username: "student", StudentID: studentID, IsActive: true}
	expected := errors.New("avatar update failed")
	store.updateContactErr = expected
	dir := t.TempDir()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), avatarDir: dir}

	_, err := svc.SaveAccountAvatar(ctx, UploadAccountAvatarInput{UserID: userID, Ext: ".webp", MimeType: "image/webp", FileSize: 3, File: bytes.NewReader([]byte("new"))})
	if !errors.Is(err, expected) {
		t.Fatalf("SaveAccountAvatar(update failure) = %v, want %v", err, expected)
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("ReadDir(%s): %v", dir, readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("avatar dir entries = %d, want new file cleanup on update error", len(entries))
	}
}

func TestAuthDeleteAccountAvatarClearsProfileAndRemovesStoredFile(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	userID := authMoreTestUUID(210)
	parentID := authMoreTestUUID(211)
	store.users["parent"] = db.User{ID: userID, Username: "parent", ParentID: parentID, IsActive: true}
	store.parentContact[parentID] = fakeAccountContact{photoURL: accountAvatarURL("parent.jpg")}
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/parent.jpg", []byte("old"), 0o644); err != nil {
		t.Fatalf("write old avatar: %v", err)
	}
	svc := &Auth{q: store, jwtSecret: []byte("secret"), avatarDir: dir}

	row, err := svc.DeleteAccountAvatar(ctx, userID)
	if err != nil {
		t.Fatalf("DeleteAccountAvatar() error = %v", err)
	}
	if row.PhotoUrl != "" {
		t.Fatalf("PhotoUrl = %q, want cleared", row.PhotoUrl)
	}
	if _, err := os.Stat(dir + "/parent.jpg"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted avatar stat err = %v, want removed", err)
	}
}

func TestAuthAvatarHelpersRejectForbiddenProfilesAndMissingRows(t *testing.T) {
	store := newFakeStore()
	unlinkedUserID := authMoreTestUUID(212)
	store.users["unlinked"] = db.User{ID: unlinkedUserID, Username: "unlinked", IsActive: true}
	svc := &Auth{q: store, jwtSecret: []byte("secret"), avatarDir: t.TempDir()}

	if _, err := svc.DeleteAccountAvatar(context.Background(), unlinkedUserID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("DeleteAccountAvatar(unlinked) = %v, want forbidden", err)
	}
	if err := svc.updateOwnedAccountAvatarURL(context.Background(), authMoreTestUUID(213), "employee", accountAvatarURL("x.png")); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("updateOwnedAccountAvatarURL(no matching row) = %v, want forbidden", err)
	}
	store.updateContactErr = pgx.ErrNoRows
	if err := svc.updateOwnedAccountAvatarURL(context.Background(), unlinkedUserID, "student", accountAvatarURL("x.png")); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("updateOwnedAccountAvatarURL(pgx no rows) = %v, want forbidden", err)
	}
}
