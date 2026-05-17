package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

func plainRequest(method, target, body string) *http.Request {
	return httptest.NewRequest(method, target, strings.NewReader(body))
}
