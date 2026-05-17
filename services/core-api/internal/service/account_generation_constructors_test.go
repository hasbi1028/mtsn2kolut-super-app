package service

import "testing"

func TestAccountGenerationConstructorsSetDefaultRolesAndStores(t *testing.T) {
	student := NewStudentAccountGenerator(nil)
	if student == nil {
		t.Fatalf("NewStudentAccountGenerator(nil) = nil")
	}
	if student.q == nil {
		t.Fatalf("NewStudentAccountGenerator(nil).q = nil, want queries store")
	}
	if student.tx != nil {
		t.Fatalf("NewStudentAccountGenerator(nil).tx = %T, want nil transaction starter", student.tx)
	}
	if student.accountRole() != StudentAccountRole {
		t.Fatalf("student account role = %q, want %q", student.accountRole(), StudentAccountRole)
	}

	studentWithPool := NewStudentAccountGeneratorWithPool(nil)
	if studentWithPool == nil {
		t.Fatalf("NewStudentAccountGeneratorWithPool(nil) = nil")
	}
	if studentWithPool.q == nil {
		t.Fatalf("NewStudentAccountGeneratorWithPool(nil).q = nil, want db queries store")
	}
	if studentWithPool.tx == nil {
		t.Fatalf("NewStudentAccountGeneratorWithPool(nil).tx = nil, want pool transaction starter")
	}
	if studentWithPool.accountRole() != StudentAccountRole {
		t.Fatalf("student with pool account role = %q, want %q", studentWithPool.accountRole(), StudentAccountRole)
	}

	parent := NewParentAccountGenerator(nil)
	if parent == nil {
		t.Fatalf("NewParentAccountGenerator(nil) = nil")
	}
	if parent.q == nil {
		t.Fatalf("NewParentAccountGenerator(nil).q = nil, want queries store")
	}
	if parent.tx != nil {
		t.Fatalf("NewParentAccountGenerator(nil).tx = %T, want nil transaction starter", parent.tx)
	}
	if parent.accountRole() != ParentAccountRole {
		t.Fatalf("parent account role = %q, want %q", parent.accountRole(), ParentAccountRole)
	}

	parentWithPool := NewParentAccountGeneratorWithPool(nil)
	if parentWithPool == nil {
		t.Fatalf("NewParentAccountGeneratorWithPool(nil) = nil")
	}
	if parentWithPool.q == nil {
		t.Fatalf("NewParentAccountGeneratorWithPool(nil).q = nil, want db queries store")
	}
	if parentWithPool.tx == nil {
		t.Fatalf("NewParentAccountGeneratorWithPool(nil).tx = nil, want pool transaction starter")
	}
	if parentWithPool.accountRole() != ParentAccountRole {
		t.Fatalf("parent with pool account role = %q, want %q", parentWithPool.accountRole(), ParentAccountRole)
	}
}

func TestAccountGenerationAccountRoleFallsBackAfterBlankOverride(t *testing.T) {
	student := &StudentAccountGenerator{role: " \t\n "}
	if got := student.accountRole(); got != StudentAccountRole {
		t.Fatalf("blank student role fallback = %q, want %q", got, StudentAccountRole)
	}

	parent := &ParentAccountGenerator{role: " \t\n "}
	if got := parent.accountRole(); got != ParentAccountRole {
		t.Fatalf("blank parent role fallback = %q, want %q", got, ParentAccountRole)
	}
}
