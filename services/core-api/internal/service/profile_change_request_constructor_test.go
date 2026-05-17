package service

import "testing"

func TestNewProfileChangeRequestReturnsServiceWithStore(t *testing.T) {
	svc := NewProfileChangeRequest(nil)
	if svc == nil {
		t.Fatal("NewProfileChangeRequest(nil) = nil, want service")
	}
	if svc.q == nil {
		t.Fatal("NewProfileChangeRequest(nil).q = nil, want query store")
	}
	if svc.tx != nil {
		t.Fatalf("NewProfileChangeRequest(nil).tx = %v, want no transaction starter", svc.tx)
	}
}

func TestNewProfileChangeRequestWithPoolReturnsServiceWithStores(t *testing.T) {
	svc := NewProfileChangeRequestWithPool(nil)
	if svc == nil {
		t.Fatal("NewProfileChangeRequestWithPool(nil) = nil, want service")
	}
	if svc.q == nil {
		t.Fatal("NewProfileChangeRequestWithPool(nil).q = nil, want query store")
	}
	if svc.tx == nil {
		t.Fatal("NewProfileChangeRequestWithPool(nil).tx = nil, want transaction starter")
	}
}
