package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONDisallowUnknownJSONFieldsOption(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	validReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ali"}`))
	validRec := httptest.NewRecorder()
	var valid payload
	if ok := decodeJSON(validRec, validReq, &valid, 1024, disallowUnknownJSONFields); !ok {
		t.Fatalf("decodeJSON(valid strict) ok=false, status=%d body=%s", validRec.Code, validRec.Body.String())
	}
	if valid.Name != "Ali" {
		t.Fatalf("decodeJSON(valid strict) name = %q, want Ali", valid.Name)
	}

	unknownReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ali","unknown":true}`))
	unknownRec := httptest.NewRecorder()
	var unknown payload
	if ok := decodeJSON(unknownRec, unknownReq, &unknown, 1024, disallowUnknownJSONFields); ok {
		t.Fatal("decodeJSON(strict unknown field) ok=true, want false")
	}
	if unknownRec.Code != http.StatusBadRequest {
		t.Fatalf("decodeJSON(strict unknown field) status = %d, want %d", unknownRec.Code, http.StatusBadRequest)
	}

	lenientReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Ali","unknown":true}`))
	lenientRec := httptest.NewRecorder()
	var lenient payload
	if ok := decodeJSON(lenientRec, lenientReq, &lenient, 1024); !ok {
		t.Fatalf("decodeJSON(lenient unknown field) ok=false, status=%d body=%s", lenientRec.Code, lenientRec.Body.String())
	}
}

func TestMaskProfileChangeWordAndValue(t *testing.T) {
	wordCases := map[string]string{
		"":              "",
		"A":             "*",
		"Ali":           "A**",
		"Siti":          "S***",
		"LongNameValue": "L******",
		"Élan":          "É***",
	}
	for input, want := range wordCases {
		if got := maskProfileChangeWord(input); got != want {
			t.Fatalf("maskProfileChangeWord(%q) = %q, want %q", input, got, want)
		}
	}

	if got, want := maskProfileChangeExportValue("nama", " Ali  Bin "), "A** B**"; got != want {
		t.Fatalf("maskProfileChangeExportValue(name) = %q, want %q", got, want)
	}
	if got, want := maskProfileChangeExportValue("tanggal_lahir", "2010-05-17"), "2010-**-**"; got != want {
		t.Fatalf("maskProfileChangeExportValue(date) = %q, want %q", got, want)
	}
	if got := maskProfileChangeExportValue("tanggal_lahir", "20"); got != "****-**-**" {
		t.Fatalf("maskProfileChangeExportValue(short date) = %q, want masked date", got)
	}
}
