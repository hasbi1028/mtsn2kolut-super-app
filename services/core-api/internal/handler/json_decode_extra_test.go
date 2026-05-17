package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONErrorMessageEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "syntax", err: &json.SyntaxError{}, want: "Payload JSON tidak valid"},
		{name: "type with field", err: &json.UnmarshalTypeError{Field: "age"}, want: "Field age tidak sesuai tipe yang diminta"},
		{name: "type without field", err: &json.UnmarshalTypeError{}, want: "Payload JSON berisi tipe data yang tidak sesuai"},
		{name: "unexpected eof", err: io.ErrUnexpectedEOF, want: "Payload JSON tidak lengkap"},
		{name: "empty body", err: io.EOF, want: "Payload JSON wajib diisi"},
		{name: "read after close", err: http.ErrBodyReadAfterClose, want: "Payload JSON tidak dapat dibaca"},
		{name: "body too large", err: errors.New("http: request body too large"), want: "Payload JSON terlalu besar"},
		{name: "unknown field", err: errors.New(`json: unknown field "extra"`), want: "Payload JSON tidak valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeJSONErrorMessage(tt.err); got != tt.want {
				t.Fatalf("decodeJSONErrorMessage(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestDecodeJSONErrorMessageBodyTooLargeThroughDecoder(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat(" ", 8)+`{"name":"Ali"}`))
	rec := httptest.NewRecorder()
	var body payload
	if ok := decodeJSON(rec, req, &body, 4); ok {
		t.Fatalf("decodeJSON(too large) ok=true, want false")
	}
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Payload JSON terlalu besar") {
		t.Fatalf("decodeJSON(too large) status/body = %d/%s, want 400 too large message", rec.Code, rec.Body.String())
	}
}
