package service

import "testing"

func TestNewExamReturnsServiceWithStoreAndPool(t *testing.T) {
	svc := NewExam(nil)
	if svc == nil {
		t.Fatal("NewExam(nil) = nil, want service")
	}
	if svc.q == nil {
		t.Fatal("NewExam(nil).q = nil, want query store")
	}
	if svc.pool != nil {
		t.Fatalf("NewExam(nil).pool = %v, want nil pool passthrough", svc.pool)
	}
}
