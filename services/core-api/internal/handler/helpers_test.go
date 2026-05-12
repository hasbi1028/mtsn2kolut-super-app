package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func handlerTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func TestCbtEventTargetLevelHelpers(t *testing.T) {
	got := normalizeTargetLevels([]string{" ix ", "VII", "viii", "VII", "", " ix"})
	want := []string{"IX", "VII", "VIII"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("normalizeTargetLevels() = %v, want %v", got, want)
	}
	if got := normalizeTargetLevels(nil); len(got) != 0 {
		t.Fatalf("normalizeTargetLevels(nil) len = %d, want 0", len(got))
	}
	if err := validateTargetLevels([]string{"VII", "VIII", "IX"}); err != nil {
		t.Fatalf("validateTargetLevels(valid) error = %v", err)
	}
	if err := validateTargetLevels([]string{"X"}); err == nil || err.Error() != "target_levels hanya boleh berisi VII, VIII, atau IX" {
		t.Fatalf("validateTargetLevels(invalid) error = %v, want target-level validation", err)
	}
}

func TestWriteClientErrorMapsStatuses(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "nil", err: nil, wantStatus: http.StatusBadRequest},
		{name: "unauthorized", err: domain.ErrUnauthorized, wantStatus: http.StatusUnauthorized},
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: http.StatusForbidden},
		{name: "not found sentinel", err: domain.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "pgx no rows", err: pgx.ErrNoRows, wantStatus: http.StatusNotFound},
		{name: "conflict sentinel", err: domain.ErrConflict, wantStatus: http.StatusConflict},
		{name: "bad request sentinel", err: domain.ErrBadRequest, wantStatus: http.StatusBadRequest},
		{name: "weak password", err: domain.ErrWeakPassword, wantStatus: http.StatusBadRequest},
		{name: "access text", err: errors.New("akses ditolak untuk role ini"), wantStatus: http.StatusForbidden},
		{name: "not found text", err: errors.New("data tidak ditemukan"), wantStatus: http.StatusNotFound},
		{name: "duplicate text", err: errors.New("nomor sudah ada"), wantStatus: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeClientError(rec, tt.err, "fallback aman")
			if rec.Code != tt.wantStatus {
				t.Fatalf("writeClientError() status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestSafeClientMessageHidesDatabaseDetails(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		fallback string
		want     string
	}{
		{name: "empty", err: errors.New(" "), fallback: "fallback aman", want: "fallback aman"},
		{name: "duplicate key", err: errors.New("duplicate key value violates unique constraint"), fallback: "Data tidak valid", want: "Data tidak valid"},
		{name: "foreign key", err: errors.New("insert violates foreign key constraint"), fallback: "Data tidak valid", want: "Data tidak valid"},
		{name: "sqlstate", err: errors.New("syntax error SQLSTATE 42601"), fallback: "Data tidak valid", want: "Data tidak valid"},
		{name: "user message", err: errors.New("nama wajib diisi"), fallback: "Data tidak valid", want: "nama wajib diisi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeClientMessage(tt.err, tt.fallback); got != tt.want {
				t.Fatalf("safeClientMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDocumentCycleClaimsHelpers(t *testing.T) {
	employeeID := "00000000-0000-0000-0000-000000000001"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"guru", "waka_kurikulum"},
		"eid":   employeeID,
	}))
	if !documentCycleHasRole(req, "waka_kurikulum") {
		t.Fatal("documentCycleHasRole([]any) = false, want true")
	}
	if documentCycleHasRole(req, "admin") {
		t.Fatal("documentCycleHasRole(admin) = true, want false")
	}
	id, ok := documentCycleEmployeeID(req)
	if !ok || !id.Valid {
		t.Fatalf("documentCycleEmployeeID() = %v, %v; want valid id", id, ok)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []string{"admin"},
		"role":  "guru",
		"eid":   "bad",
	}))
	if !documentCycleHasRole(req, "admin") {
		t.Fatal("documentCycleHasRole([]string) = false, want true")
	}
	if id, ok := documentCycleEmployeeID(req); ok || id.Valid {
		t.Fatalf("documentCycleEmployeeID(invalid) = %v, %v; want invalid false", id, ok)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"admin", "guru"},
		"eid":   "00000000-0000-0000-0000-000000000008",
	}))
	if id, ok := documentCycleEmployeeID(req); ok || id.Valid {
		t.Fatalf("documentCycleEmployeeID(admin+guru) = %v, %v; want invalid admin-wide scope", id, ok)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"role": "kepala_madrasah"}))
	if !documentCycleHasRole(req, "kepala_madrasah") {
		t.Fatal("documentCycleHasRole(role) = false, want true")
	}
}

func TestTeacherClaimHelpersRespectRolesArrayAndRoleFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   "00000000-0000-0000-0000-000000000002",
		"usr":   "guru.mapel",
		"sub":   "user-2",
	}))

	if got := cbtSessionTeacherID(req); !got.Valid {
		t.Fatal("cbtSessionTeacherID(roles array) invalid, want valid")
	}
	if got := gradeTeacherEmployeeID(req); !got.Valid {
		t.Fatal("gradeTeacherEmployeeID(roles array) invalid, want valid")
	}
	if got := journalEmployeeID(req); !got.Valid {
		t.Fatal("journalEmployeeID(roles array) invalid, want valid")
	}
	if !journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(roles array) = false, want true")
	}
	if got := currentUsername(req); got != "guru.mapel" {
		t.Fatalf("currentUsername() = %q, want %q", got, "guru.mapel")
	}
	if got := websiteActorUsername(req); got != "guru.mapel" {
		t.Fatalf("websiteActorUsername() = %q, want %q", got, "guru.mapel")
	}
	subOnlyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	subOnlyReq = subOnlyReq.WithContext(context.WithValue(subOnlyReq.Context(), api.ClaimsKey, jwt.MapClaims{"sub": "fallback-sub"}))
	if got := currentUsername(subOnlyReq); got != "fallback-sub" {
		t.Fatalf("currentUsername(sub fallback) = %q, want fallback-sub", got)
	}
	if got := currentUsername(httptest.NewRequest(http.MethodGet, "/", nil)); got != "" {
		t.Fatalf("currentUsername(no claims) = %q, want empty", got)
	}
	if hasAnyRole(httptest.NewRequest(http.MethodGet, "/", nil), "admin") {
		t.Fatal("hasAnyRole(no claims) = true, want false")
	}
	if !journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(guru) = false, want true")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"admin", "guru"},
		"eid":   "00000000-0000-0000-0000-000000000003",
	}))
	if got := cbtSessionTeacherID(req); got.Valid {
		t.Fatal("cbtSessionTeacherID(admin+guru) valid, want invalid admin-wide scope")
	}
	if got := gradeTeacherEmployeeID(req); got.Valid {
		t.Fatal("gradeTeacherEmployeeID(admin+guru) valid, want invalid admin-wide scope")
	}
	if got := journalEmployeeID(req); got.Valid {
		t.Fatal("journalEmployeeID(admin+guru) valid, want invalid admin-wide scope")
	}
	if !journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(admin+guru) = false, want true")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role": "guru",
		"eid":  "00000000-0000-0000-0000-000000000004",
	}))
	if got := gradeTeacherEmployeeID(req); !got.Valid {
		t.Fatal("gradeTeacherEmployeeID(role fallback) invalid, want valid")
	}
	if got := journalEmployeeID(req); !got.Valid {
		t.Fatal("journalEmployeeID(role fallback) invalid, want valid")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"permissions": []any{"journal.read"},
	}))
	if journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(journal.read without eid) = true, want false")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"permissions": []any{"journal.read"},
		"eid":         "00000000-0000-0000-0000-000000000006",
	}))
	if !journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(journal.read with eid) = false, want true")
	}
	if got := journalEmployeeID(req); !got.Valid {
		t.Fatal("journalEmployeeID(journal.read with eid) invalid, want scoped employee")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"permissions": []any{"journal.read_all"},
	}))
	if !journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(journal.read_all without eid) = false, want true")
	}
	if got := journalEmployeeID(req); got.Valid {
		t.Fatal("journalEmployeeID(journal.read_all) valid, want all-scope")
	}
}

func TestSensitiveAccessHelpersDenyWithoutClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if gradeAccessAllowed(req) {
		t.Fatal("gradeAccessAllowed(no claims) = true, want false")
	}
	if journalAccessAllowed(req) {
		t.Fatal("journalAccessAllowed(no claims) = true, want false")
	}
	access := getKesiswaanAccess(req)
	if access.canRead || access.canManage || access.teacherEmployeeID.Valid {
		t.Fatalf("getKesiswaanAccess(no claims) = %+v, want zero access", access)
	}
}

func TestKesiswaanAccessRespectsAdminWideOverGuruScope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   "00000000-0000-0000-0000-000000000005",
	}))
	access := getKesiswaanAccess(req)
	if !access.canRead || access.canManage {
		t.Fatalf("getKesiswaanAccess(guru) = %+v, want read-only guru scope", access)
	}
	if !access.teacherEmployeeID.Valid {
		t.Fatalf("getKesiswaanAccess(guru) teacherEmployeeID invalid, want valid scope")
	}
	if got := kesiswaanEmployeeID(req); !got.Valid {
		t.Fatalf("kesiswaanEmployeeID(guru) invalid, want valid guru scope")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"admin", "guru"},
		"eid":   "00000000-0000-0000-0000-000000000006",
	}))
	access = getKesiswaanAccess(req)
	if !access.canRead || !access.canManage {
		t.Fatalf("getKesiswaanAccess(admin+guru) = %+v, want admin-wide manage access", access)
	}
	if access.teacherEmployeeID.Valid {
		t.Fatalf("getKesiswaanAccess(admin+guru) teacherEmployeeID = %v, want invalid admin-wide scope", access.teacherEmployeeID)
	}
	if got := kesiswaanEmployeeID(req); got.Valid {
		t.Fatalf("kesiswaanEmployeeID(admin+guru) = %v, want invalid admin-wide scope", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role": "guru",
		"eid":  "00000000-0000-0000-0000-000000000007",
	}))
	access = getKesiswaanAccess(req)
	if !access.canRead || access.canManage || !access.teacherEmployeeID.Valid {
		t.Fatalf("getKesiswaanAccess(role fallback guru) = %+v, want guru-scoped read access", access)
	}
	if got := kesiswaanEmployeeID(req); !got.Valid {
		t.Fatalf("kesiswaanEmployeeID(role fallback guru) invalid, want valid guru scope")
	}
}

func TestPortalRoleHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"siswa", "guru"},
	}))
	if !portalHasAnyRole(req, "siswa") {
		t.Fatal("portalHasAnyRole(siswa) = false, want true")
	}
	if portalHasAnyRole(req, "ortu") {
		t.Fatal("portalHasAnyRole(ortu) = true, want false")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"role": "ortu",
	}))
	if !portalHasAnyRole(req, "ortu") {
		t.Fatal("portalHasAnyRole(role fallback) = false, want true")
	}
}

func TestEmployeeHelpers(t *testing.T) {
	id := handlerTestUUID(1)
	row := db.GetEmployeeRow{
		ID:              id,
		Nip:             "19800101",
		Nama:            "Guru PNS",
		UnitKerja:       "Kurikulum",
		EmploymentType:  "pns",
		PusakaUsername:  "guru.pns",
		PusakaPassword:  "secret",
		PusakaIsEnabled: true,
		IsActive:        true,
	}
	sanitized := sanitizeEmployee(row)
	if sanitized.ID != id || !sanitized.PusakaEligible || !sanitized.HasPusakaAccount || !sanitized.PusakaIsEnabled {
		t.Fatalf("sanitizeEmployee() = %+v, want eligible PUSAKA account", sanitized)
	}

	employees := sanitizeEmployees([]db.ListEmployeesRow{
		{ID: handlerTestUUID(2), Nama: "Pegawai PPPK", EmploymentType: "pppk", PusakaUsername: "pppk"},
		{ID: handlerTestUUID(3), Nama: "Honorer", EmploymentType: "honorer"},
	})
	if len(employees) != 2 {
		t.Fatalf("sanitizeEmployees() len = %d, want 2", len(employees))
	}
	if !employees[0].PusakaEligible || !employees[0].HasPusakaAccount {
		t.Fatalf("sanitizeEmployees()[0] = %+v, want eligible account", employees[0])
	}
	if employees[1].PusakaEligible || employees[1].HasPusakaAccount {
		t.Fatalf("sanitizeEmployees()[1] = %+v, want ineligible without account", employees[1])
	}

	parsed, err := parseUUID("00000000-0000-0000-0000-000000000004")
	if err != nil || !parsed.Valid {
		t.Fatalf("parseUUID() = %v, %v; want valid id", parsed, err)
	}
	if got := pgUUIDString(parsed); got != "00000000-0000-0000-0000-000000000004" {
		t.Fatalf("pgUUIDString() = %q, want parsed uuid", got)
	}
	if got := pgUUIDString(pgtype.UUID{}); got != "" {
		t.Fatalf("pgUUIDString(invalid) = %q, want empty", got)
	}
}

func TestEmployeeClientMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "nil", err: nil, want: "fallback"},
		{name: "pns only", err: errors.New("only pns or pppk employees can have pusaka accounts"), want: "akun PUSAKA hanya untuk pegawai PNS atau PPPK"},
		{name: "invalid employment", err: errors.New("jenis kepegawaian tidak valid"), want: "jenis kepegawaian tidak valid"},
		{name: "disable before change", err: errors.New("disable or remove the pusaka account before changing employee type"), want: "nonaktifkan atau hapus akun PUSAKA sebelum mengubah jenis kepegawaian"},
		{name: "not configured", err: errors.New("pusaka account is not configured"), want: "akun PUSAKA belum dikonfigurasi"},
		{name: "fallback safe", err: errors.New("nama wajib diisi"), want: "nama wajib diisi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := employeeClientMessage(tt.err, "fallback"); got != tt.want {
				t.Fatalf("employeeClientMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
