package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
)

const defaultJSONBodyLimit = 1 << 20 // 1 MiB for ordinary JSON mutation endpoints.

type decodeJSONOption func(*json.Decoder)

func disallowUnknownJSONFields(dec *json.Decoder) {
	dec.DisallowUnknownFields()
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, limit int64, opts ...decodeJSONOption) bool {
	if limit <= 0 {
		limit = defaultJSONBodyLimit
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	for _, opt := range opts {
		if opt != nil {
			opt(dec)
		}
	}
	if err := dec.Decode(dst); err != nil {
		api.BadRequest(w, decodeJSONErrorMessage(err))
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		api.BadRequest(w, "Payload JSON hanya boleh berisi satu objek")
		return false
	}
	return true
}

func decodeJSONErrorMessage(err error) string {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntaxErr):
		return "Payload JSON tidak valid"
	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return fmt.Sprintf("Field %s tidak sesuai tipe yang diminta", typeErr.Field)
		}
		return "Payload JSON berisi tipe data yang tidak sesuai"
	case errors.Is(err, io.ErrUnexpectedEOF):
		return "Payload JSON tidak lengkap"
	case errors.Is(err, io.EOF):
		return "Payload JSON wajib diisi"
	case errors.Is(err, http.ErrBodyReadAfterClose):
		return "Payload JSON tidak dapat dibaca"
	default:
		if err.Error() == "http: request body too large" {
			return "Payload JSON terlalu besar"
		}
		return "Payload JSON tidak valid"
	}
}
