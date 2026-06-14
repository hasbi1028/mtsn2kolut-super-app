package handler

import (
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// pageSize parses a string page-size parameter with a fallback default.
// Values are clamped to [1, 200].
func pageSize(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	if v > 200 {
		v = 200
	}
	return v
}

// pageNum parses a string page-number parameter with a fallback default.
func pageNum(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// timestamptzRFC3339 formats a pgtype.Timestamptz as an RFC3339 string.
func timestamptzRFC3339(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z07:00")
}

// boolQuery parses a string query parameter as a boolean.
// Accepts "true", "1", "yes" (case-insensitive) as true; everything else is false.
func boolQuery(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

// int32Query parses a string query parameter as int32 with zero as default.
func int32Query(raw string) int32 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0
	}
	return int32(v)
}
