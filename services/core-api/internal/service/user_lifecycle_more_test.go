package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestUserLifecycleWithStoreWithoutTxUsesPrimaryStoreAndPropagatesError(t *testing.T) {
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}
	expected := errors.New("callback failed")
	var gotStore userLifecycleStore

	err := svc.withStore(context.Background(), func(s userLifecycleStore) error {
		gotStore = s
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("withStore() error = %v, want %v", err, expected)
	}
	if gotStore != store {
		t.Fatalf("withStore() used store %T, want primary fake store", gotStore)
	}
}

func TestUserLifecycleDeleteAsDeactivateStopsWhenAdminGuardFails(t *testing.T) {
	store := &fakeUserLifecycleStore{userHasAdmin: true, adminCount: 1}
	svc := &UserLifecycle{q: store}

	err := svc.DeleteAsDeactivate(context.Background(), userLifecycleTestUUID(151), userLifecycleTestUUID(152))
	if err == nil {
		t.Fatal("DeleteAsDeactivate(last admin) error = nil, want guard error")
	}
	if len(store.softDeleteCalls) != 0 || len(store.revokeCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("calls = softDelete %v revoke %v audit %v, want no mutations after guard failure", store.softDeleteCalls, store.revokeCalls, store.auditCalls)
	}
}

func TestUserLifecycleDeleteAsDeactivateStopsBeforeAuditWhenRevokeFails(t *testing.T) {
	userID := userLifecycleTestUUID(153)
	expected := errors.New("revoke failed")
	store := &fakeUserLifecycleStore{revokeErr: expected}
	svc := &UserLifecycle{q: store}

	err := svc.DeleteAsDeactivate(context.Background(), userID, userLifecycleTestUUID(154))
	if !errors.Is(err, expected) {
		t.Fatalf("DeleteAsDeactivate(revoke failure) = %v, want %v", err, expected)
	}
	if len(store.softDeleteCalls) != 1 || store.softDeleteCalls[0] != userID {
		t.Fatalf("softDeleteCalls = %+v, want delete before revoke failure", store.softDeleteCalls)
	}
	if len(store.auditCalls) != 0 {
		t.Fatalf("auditCalls = %+v, want no audit after revoke failure", store.auditCalls)
	}
}

func TestUserLifecycleResetPasswordStopsAfterPasswordUpdateFailure(t *testing.T) {
	userID := userLifecycleTestUUID(155)
	expected := errors.New("password update failed")
	store := &fakeUserLifecycleStore{
		userByID:    db.GetUserByIDRow{ID: userID, Username: "operator"},
		passwordErr: expected,
	}
	svc := &UserLifecycle{q: store}

	err := svc.ResetPassword(context.Background(), userID, "newSecret123", userLifecycleTestUUID(156))
	if !errors.Is(err, expected) {
		t.Fatalf("ResetPassword(password update failure) = %v, want %v", err, expected)
	}
	if len(store.passwordCalls) != 1 || store.passwordCalls[0].ID != userID {
		t.Fatalf("passwordCalls = %+v, want attempted update", store.passwordCalls)
	}
	if len(store.mustChangePasswordIDs) != 0 || len(store.versionCalls) != 0 || len(store.revokeCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("calls after password failure = mustChange %v version %v revoke %v audit %v, want none", store.mustChangePasswordIDs, store.versionCalls, store.revokeCalls, store.auditCalls)
	}
}

func TestUserLifecycleResetPasswordStopsBeforeRevokeWhenVersionIncrementFails(t *testing.T) {
	userID := userLifecycleTestUUID(157)
	expected := errors.New("version failed")
	store := &fakeUserLifecycleStore{
		userByID:   db.GetUserByIDRow{ID: userID, Username: "operator"},
		versionErr: expected,
	}
	svc := &UserLifecycle{q: store}

	err := svc.ResetPassword(context.Background(), userID, "newSecret123", userLifecycleTestUUID(158))
	if !errors.Is(err, expected) {
		t.Fatalf("ResetPassword(version failure) = %v, want %v", err, expected)
	}
	if len(store.passwordCalls) != 1 || len(store.mustChangePasswordIDs) != 1 || len(store.versionCalls) != 1 {
		t.Fatalf("calls before version failure = password %v mustChange %v version %v", store.passwordCalls, store.mustChangePasswordIDs, store.versionCalls)
	}
	if len(store.revokeCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("calls after version failure = revoke %v audit %v, want none", store.revokeCalls, store.auditCalls)
	}
}

func TestUserLifecycleUpdateProfileLinkAllowsClearingAllLinksAndAuditsEmptyMetadata(t *testing.T) {
	userID := userLifecycleTestUUID(159)
	actorID := userLifecycleTestUUID(160)
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateProfileLink(context.Background(), userID, ProfileLink{}, actorID); err != nil {
		t.Fatalf("UpdateProfileLink(clear all) error = %v", err)
	}
	if len(store.profileCalls) != 1 || store.profileCalls[0].ID != userID || store.profileCalls[0].EmployeeID.Valid || store.profileCalls[0].StudentID.Valid || store.profileCalls[0].ParentID.Valid {
		t.Fatalf("profileCalls = %+v, want one clear-all call", store.profileCalls)
	}
	if len(store.versionCalls) != 1 || len(store.revokeCalls) != 1 || len(store.auditCalls) != 1 {
		t.Fatalf("version/revoke/audit calls = %v/%v/%v, want invalidation and audit", store.versionCalls, store.revokeCalls, store.auditCalls)
	}
	var metadata map[string]any
	if err := json.Unmarshal(store.auditCalls[0].Metadata, &metadata); err != nil {
		t.Fatalf("audit metadata unmarshal: %v", err)
	}
	if metadata["employee_id"] != "" || metadata["student_id"] != "" || metadata["parent_id"] != "" {
		t.Fatalf("audit metadata = %+v, want empty profile ids for clear-all", metadata)
	}
}

func TestUserLifecycleUpdateProfileLinkStopsBeforeRevokeWhenVersionFails(t *testing.T) {
	userID := userLifecycleTestUUID(161)
	employeeID := userLifecycleTestUUID(162)
	expected := errors.New("version failed")
	store := &fakeUserLifecycleStore{versionErr: expected}
	svc := &UserLifecycle{q: store}

	err := svc.UpdateProfileLink(context.Background(), userID, ProfileLink{EmployeeID: employeeID}, userLifecycleTestUUID(163))
	if !errors.Is(err, expected) {
		t.Fatalf("UpdateProfileLink(version failure) = %v, want %v", err, expected)
	}
	if len(store.profileCalls) != 1 || len(store.versionCalls) != 1 {
		t.Fatalf("profile/version calls = %+v/%+v, want both attempted", store.profileCalls, store.versionCalls)
	}
	if len(store.revokeCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("revoke/audit calls = %+v/%+v, want stopped before revoke/audit", store.revokeCalls, store.auditCalls)
	}
}
