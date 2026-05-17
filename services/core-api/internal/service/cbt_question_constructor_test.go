package service

import "testing"

func TestNewCbtQuestionWithPoolReturnsServiceWithStores(t *testing.T) {
	svc := NewCbtQuestionWithPool(nil)
	if svc == nil {
		t.Fatal("NewCbtQuestionWithPool(nil) = nil, want service")
	}
	if svc.q == nil {
		t.Fatal("NewCbtQuestionWithPool(nil).q = nil, want query store")
	}
	if svc.tx == nil {
		t.Fatal("NewCbtQuestionWithPool(nil).tx = nil, want transaction starter")
	}
}
