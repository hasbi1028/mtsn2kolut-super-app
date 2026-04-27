package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// ClaimsKey is the context key used by the JWT middleware to store parsed claims.
const ClaimsKey contextKey = "claims"

// ClaimsFromContext returns the JWT claims stored by the auth middleware, if any.
func ClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	c, ok := ctx.Value(ClaimsKey).(jwt.MapClaims)
	return c, ok
}

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type PageMeta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

type PagedResponse struct {
	Data any      `json:"data"`
	Meta PageMeta `json:"meta"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Response{Data: data})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, Response{Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Err(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, Response{Error: msg})
}

func BadRequest(w http.ResponseWriter, msg string)    { Err(w, http.StatusBadRequest, msg) }
func Conflict(w http.ResponseWriter, msg string)      { Err(w, http.StatusConflict, msg) }
func Unauthorized(w http.ResponseWriter)               { Err(w, http.StatusUnauthorized, "unauthorized") }
func Forbidden(w http.ResponseWriter)                  { Err(w, http.StatusForbidden, "forbidden") }
func NotFound(w http.ResponseWriter)                   { Err(w, http.StatusNotFound, "not found") }
func Internal(w http.ResponseWriter, err error)        { Err(w, http.StatusInternalServerError, err.Error()) }
