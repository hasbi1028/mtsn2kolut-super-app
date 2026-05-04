package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
)

func writeClientError(w http.ResponseWriter, err error, fallback string) {
	message := ""
	if err != nil {
		message = strings.ToLower(strings.TrimSpace(err.Error()))
	}

	switch {
	case err == nil:
		api.BadRequest(w, fallback)
	case errors.Is(err, domain.ErrUnauthorized):
		api.Unauthorized(w)
	case errors.Is(err, domain.ErrForbidden):
		api.Forbidden(w)
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		api.NotFound(w)
	case errors.Is(err, domain.ErrConflict):
		api.Conflict(w, safeClientMessage(err, fallback))
	case errors.Is(err, domain.ErrBadRequest), errors.Is(err, domain.ErrWeakPassword):
		api.BadRequest(w, safeClientMessage(err, fallback))
	case strings.Contains(message, "akses ditolak"), strings.Contains(message, "forbidden"):
		api.Forbidden(w)
	case strings.Contains(message, "tidak ditemukan"), strings.Contains(message, "not found"):
		api.NotFound(w)
	case strings.Contains(message, "sudah ada"), strings.Contains(message, "duplicate"), strings.Contains(message, "unique"):
		api.Conflict(w, safeClientMessage(err, fallback))
	default:
		api.BadRequest(w, safeClientMessage(err, fallback))
	}
}

func writeDomainOrInternal(w http.ResponseWriter, err error, fallback string) {
	if err == nil {
		api.BadRequest(w, fallback)
		return
	}
	if errors.Is(err, domain.ErrUnauthorized) || errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrBadRequest) || errors.Is(err, pgx.ErrNoRows) {
		writeClientError(w, err, fallback)
		return
	}
	api.Internal(w, err)
}

func safeClientMessage(err error, fallback string) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return fallback
	}

	lower := strings.ToLower(message)
	if looksLikeConstraintError(lower) {
		if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
			return fallback
		}
		if strings.Contains(lower, "foreign key") {
			return fallback
		}
		if strings.Contains(lower, "syntax error") || strings.Contains(lower, "sqlstate") {
			return fallback
		}
	}

	return message
}

func looksLikeConstraintError(message string) bool {
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "violates") ||
		strings.Contains(message, "constraint") ||
		strings.Contains(message, "sqlstate") ||
		strings.Contains(message, "uq_") ||
		strings.Contains(message, "fk_") ||
		strings.Contains(message, "unique")
}
