package handler

import (
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
)

func adminAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin")
}

func cbtAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin", "guru") || mw.HasAnyPermission(claims,
		"cbt.read", "cbt.manage", "cbt.proctor", "cbt.score", "cbt.result_read", "cbt.result_manage",
	)
}

func cbtOpsAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin", "guru", "staf") || mw.HasAnyPermission(claims, "cbt.proctor")
}

func academicReadAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin", "guru", "staf", "kesiswaan") ||
		mw.HasAnyPermission(claims, "academic.read", "academic.manage")
}

func academicSubjectListAccessAllowed(r *http.Request) bool {
	return academicReadAccessAllowed(r) || cbtAccessAllowed(r)
}
