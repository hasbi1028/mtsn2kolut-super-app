package api

import (
	"encoding/json"
	"net/http"
)

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
func Unauthorized(w http.ResponseWriter)               { Err(w, http.StatusUnauthorized, "unauthorized") }
func Forbidden(w http.ResponseWriter)                  { Err(w, http.StatusForbidden, "forbidden") }
func NotFound(w http.ResponseWriter)                   { Err(w, http.StatusNotFound, "not found") }
func Internal(w http.ResponseWriter, err error)        { Err(w, http.StatusInternalServerError, err.Error()) }
