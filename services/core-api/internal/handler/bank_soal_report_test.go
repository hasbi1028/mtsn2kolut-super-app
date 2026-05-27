package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestBankSoalReportExportRejectsInvalidPostJSON(t *testing.T) {
	h := &BankSoalReport{}

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{`},
		{name: "multiple objects", body: `{"format":"csv"}{"format":"png"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/questions/reports/export", strings.NewReader(tc.body)), jwt.MapClaims{
				"roles": []any{"guru"},
				"usr":   "guru.ipa",
			})
			rec := httptest.NewRecorder()

			h.Export(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), "error") {
				t.Fatalf("body=%s", rec.Body.String())
			}
		})
	}
}
