package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStudentCertificateListTemplatesForbiddenWithoutTUAccess(t *testing.T) {
	h := NewStudentCertificate(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/tu/surat-keterangan/templates", nil)
	rec := httptest.NewRecorder()

	h.ListTemplates(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestInventoryStatsForbiddenWithoutStaffAccess(t *testing.T) {
	h := NewInventory(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/inventory/stats", nil)
	rec := httptest.NewRecorder()

	h.Stats(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
