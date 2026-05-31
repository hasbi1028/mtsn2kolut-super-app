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
		"bank_soal.read", "bank_soal.create", "bank_soal.update", "bank_soal.review", "bank_soal.publish", "bank_soal.import", "bank_soal.delete", "bank_soal.analytics", "bank_soal.settings",
		"asesmen.read", "asesmen.package_manage", "asesmen.event_manage", "asesmen.session_manage", "asesmen.participant_manage", "asesmen.proctor", "asesmen.score", "asesmen.result_read", "asesmen.result_manage",
	)
}

func cbtOpsAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin", "guru", "staf") || mw.HasAnyPermission(claims, "asesmen.proctor")
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
