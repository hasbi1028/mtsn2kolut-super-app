package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func mustTargetedHelperUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("scan uuid %q: %v", raw, err)
	}
	return id
}

func requestWithURLParam(rawURL, name, value string, body io.Reader) *http.Request {
	req := httptest.NewRequest(http.MethodGet, rawURL, body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(name, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestSchoolRoomPayloadDecodeAndParams(t *testing.T) {
	body := bytes.NewBufferString(`{
		"code":" R-01 ","name":" Lab CBT ","building":" A ","floor":" 2 ",
		"room_type":" lab ","location_note":" east wing ","default_capacity":36,
		"exam_capacity":30,"condition":" good ","is_exam_eligible":false,
		"network_ready":true,"power_ready":true,"notes":" siap ujian "}`)
	req := httptest.NewRequest(http.MethodPost, "/rooms", body)
	rec := httptest.NewRecorder()

	payload, ok := decodeSchoolRoomPayload(rec, req)
	if !ok {
		t.Fatal("decodeSchoolRoomPayload returned ok=false for valid JSON")
	}
	if payload.examEligible() {
		t.Fatal("examEligible() = true, want false when explicit false is supplied")
	}
	create := payload.toCreateParams()
	if create.Code != "R-01" || create.Name != "Lab CBT" || create.Building != "A" || create.RoomType != "lab" || create.LocationNote != "east wing" || create.Condition != "good" || create.Notes != "siap ujian" {
		t.Fatalf("toCreateParams did not trim string fields: %#v", create)
	}
	if create.DefaultCapacity != 36 || create.ExamCapacity != 30 || create.IsExamEligible || !create.NetworkReady || !create.PowerReady {
		t.Fatalf("toCreateParams produced unexpected non-string fields: %#v", create)
	}
	update := payload.toUpdateParams()
	if update.Code != create.Code || update.Name != create.Name || update.IsExamEligible != create.IsExamEligible || update.Notes != create.Notes {
		t.Fatalf("toUpdateParams diverged from create normalization: create=%#v update=%#v", create, update)
	}
	if !((schoolRoomPayload{}).examEligible()) {
		t.Fatal("examEligible() = false, want true when is_exam_eligible is omitted")
	}

	badReq := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewBufferString(`{"code"`))
	badRec := httptest.NewRecorder()
	if _, ok := decodeSchoolRoomPayload(badRec, badReq); ok {
		t.Fatal("decodeSchoolRoomPayload accepted malformed JSON")
	}
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("malformed JSON status = %d, want %d", badRec.Code, http.StatusBadRequest)
	}
}

func TestRombelOptionalParsersHomeroomAndRelationship(t *testing.T) {
	rec := httptest.NewRecorder()
	if id, ok := parseOptionalUUIDParam(rec, "   ", "academic_year_id"); !ok || id.Valid {
		t.Fatalf("blank optional UUID = (%#v,%v), want invalid/true", id, ok)
	}
	validUUID := "11111111-1111-1111-1111-111111111111"
	if id, ok := parseOptionalUUIDParam(httptest.NewRecorder(), validUUID, "academic_year_id"); !ok || !id.Valid {
		t.Fatalf("valid optional UUID = (%#v,%v), want valid/true", id, ok)
	}
	badUUIDRec := httptest.NewRecorder()
	if _, ok := parseOptionalUUIDParam(badUUIDRec, "not-a-uuid", "academic_year_id"); ok || badUUIDRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid optional UUID ok/status = %v/%d, want false/%d", ok, badUUIDRec.Code, http.StatusBadRequest)
	}

	if date, ok := parseOptionalDateParam(httptest.NewRecorder(), "2026-05-17", "start_date"); !ok || !date.Valid || optionalDateArg(date) == nil {
		t.Fatalf("valid optional date = (%#v,%v), want valid/true and non-nil arg", date, ok)
	}
	if date, ok := parseOptionalDateParam(httptest.NewRecorder(), "", "end_date"); !ok || date.Valid || optionalDateArg(date) != nil {
		t.Fatalf("blank optional date = (%#v,%v), want invalid/true and nil arg", date, ok)
	}
	badDateRec := httptest.NewRecorder()
	if _, ok := parseOptionalDateParam(badDateRec, "17-05-2026", "start_date"); ok || badDateRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid optional date ok/status = %v/%d, want false/%d", ok, badDateRec.Code, http.StatusBadRequest)
	}

	classID := mustTargetedHelperUUID(t, "22222222-2222-2222-2222-222222222222")
	params, ok := parseCreateHomeroomAssignment(httptest.NewRecorder(), classID, homeroomAssignmentRequest{
		EmployeeID:     "33333333-3333-3333-3333-333333333333",
		AcademicYearID: validUUID,
		StartDate:      "2026-05-17",
		EndDate:        "2026-05-31",
		Notes:          " wali utama ",
	})
	if !ok {
		t.Fatal("parseCreateHomeroomAssignment returned ok=false for valid input")
	}
	if params.ClassID != classID || !params.HomeroomIsActive || params.Notes != "wali utama" || params.HomeroomAcademicYearID == nil || params.HomeroomStartDate == nil || !params.HomeroomEndDate.Valid {
		t.Fatalf("parseCreateHomeroomAssignment produced unexpected params: %#v", params)
	}
	inactive := false
	update, ok := parseUpdateHomeroomAssignment(httptest.NewRecorder(), mustTargetedHelperUUID(t, "44444444-4444-4444-4444-444444444444"), homeroomAssignmentRequest{
		EmployeeID: "33333333-3333-3333-3333-333333333333",
		IsActive:   &inactive,
		Notes:      " nonaktif ",
	})
	if !ok || update.HomeroomIsActive || update.Notes != "nonaktif" {
		t.Fatalf("parseUpdateHomeroomAssignment = (%#v,%v), want inactive trimmed", update, ok)
	}

	if got := relationshipString([]byte("mother")); got != "mother" {
		t.Fatalf("relationshipString([]byte) = %q, want mother", got)
	}
	if got := relationshipString("father"); got != "father" {
		t.Fatalf("relationshipString(string) = %q, want father", got)
	}
	if got := relationshipString(123); got != "123" {
		t.Fatalf("relationshipString(int) = %q, want 123", got)
	}
}

func TestGroupRombelStudentsPreservesOrderAndParents(t *testing.T) {
	studentA := mustTargetedHelperUUID(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	studentB := mustTargetedHelperUUID(t, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	parentA1 := mustTargetedHelperUUID(t, "11111111-1111-1111-1111-111111111111")
	parentA2 := mustTargetedHelperUUID(t, "22222222-2222-2222-2222-222222222222")
	rows := []db.ListStudentsByClassWithParentsRow{
		{StudentID: studentA, Nis: "001", Nisn: "NISN1", StudentName: "Amin", Gender: db.GenderEnumL, ParentName: "Legacy A", ParentPhone: "0801", StudentPhone: "0811", StudentAddress: "Addr A", IsActive: true, Status: db.StudentStatusEnumActive, ParentID: parentA1, ParentNama: "Ibu Amin", ParentPhoneLinked: "0821", Relationship: []byte("mother"), IsPrimaryContact: true, RelationshipNotes: "utama"},
		{StudentID: studentA, Nis: "001", Nisn: "NISN1", StudentName: "Amin", Gender: db.GenderEnumL, IsActive: true, Status: db.StudentStatusEnumActive, ParentID: parentA2, ParentNama: "Ayah Amin", ParentPhoneLinked: "0822", Relationship: "father"},
		{StudentID: studentB, Nis: "002", StudentName: "Budi", Gender: db.GenderEnumL, IsActive: false, Status: db.StudentStatusEnumMutated},
	}

	got := groupRombelStudents(rows)
	if len(got) != 2 {
		t.Fatalf("groupRombelStudents length = %d, want 2: %#v", len(got), got)
	}
	if got[0].ID != studentA || got[0].Nama != "Amin" || got[1].ID != studentB || got[1].Nama != "Budi" {
		t.Fatalf("groupRombelStudents did not preserve first-seen student order/details: %#v", got)
	}
	if len(got[0].Parents) != 2 || got[0].Parents[0].Relationship != "mother" || !got[0].Parents[0].IsPrimaryContact || got[0].Parents[1].Relationship != "father" {
		t.Fatalf("groupRombelStudents parent aggregation mismatch: %#v", got[0].Parents)
	}
	if got[1].Parents == nil || len(got[1].Parents) != 0 {
		t.Fatalf("student without valid parent should have an empty non-nil parents slice: %#v", got[1].Parents)
	}
}

func TestSystemMaintenanceParsersAndDecoders(t *testing.T) {
	if got := int32Param("25", 50); got != 25 {
		t.Fatalf("int32Param valid = %d, want 25", got)
	}
	for _, raw := range []string{"", "bad", "999999999999999999999"} {
		if got := int32Param(raw, 50); got != 50 {
			t.Fatalf("int32Param(%q) = %d, want fallback 50", raw, got)
		}
	}
	limit, offset := listLimitOffset(httptest.NewRequest(http.MethodGet, "/maintenance?limit=10&offset=5", nil), 50)
	if limit != 10 || offset != 5 {
		t.Fatalf("listLimitOffset = (%d,%d), want (10,5)", limit, offset)
	}
	limit, offset = listLimitOffset(httptest.NewRequest(http.MethodGet, "/maintenance?limit=nope", nil), 50)
	if limit != 50 || offset != 0 {
		t.Fatalf("listLimitOffset fallback = (%d,%d), want (50,0)", limit, offset)
	}

	from := optionalTimeQuery("2026-05-17T10:00:00Z")
	if from == nil || from.UTC().Format(time.RFC3339) != "2026-05-17T10:00:00Z" {
		t.Fatalf("optionalTimeQuery valid = %#v, want parsed RFC3339", from)
	}
	if optionalTimeQuery("not-time") != nil || optionalTimeQuery(" ") != nil {
		t.Fatal("optionalTimeQuery should return nil for blank/invalid values")
	}

	allowBypass := false
	starts := "2026-05-17T10:00:00Z"
	ends := "2026-05-17T11:00:00Z"
	payload := `{"title":"Upgrade","message":"Maintenance","mode":"scheduled","affected_modules":["cbt","inventory"],"starts_at":"` + starts + `","ends_at":"` + ends + `","is_active":true,"allow_admin_bypass":false,"bypass_roles":["admin"],"severity":"warning","reason":"deploy"}`
	req := httptest.NewRequest(http.MethodPost, "/maintenance", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()
	decoded, ok := decodeMaintenanceWindowRequest(rec, req)
	if !ok {
		t.Fatal("decodeMaintenanceWindowRequest returned ok=false for valid payload")
	}
	if decoded.Title != "Upgrade" || decoded.StartsAt == nil || decoded.EndsAt == nil || decoded.AllowAdminBypass == nil || *decoded.AllowAdminBypass != allowBypass {
		t.Fatalf("decoded maintenance request mismatch: %#v", decoded)
	}
	actor := mustTargetedHelperUUID(t, "55555555-5555-5555-5555-555555555555")
	input := decoded.toServiceInput(actor)
	if input.Title != decoded.Title || input.ActorUserID != actor || input.Reason != "deploy" || !reflect.DeepEqual(input.AffectedModules, []string{"cbt", "inventory"}) || input.AllowAdminBypass == nil || *input.AllowAdminBypass {
		t.Fatalf("toServiceInput mismatch: %#v", input)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/maintenance", bytes.NewBufferString(`{"title"`))
	badRec := httptest.NewRecorder()
	if _, ok := decodeMaintenanceWindowRequest(badRec, badReq); ok || badRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid maintenance payload ok/status = %v/%d, want false/%d", ok, badRec.Code, http.StatusBadRequest)
	}

	if got := decodeMaintenanceReason(httptest.NewRequest(http.MethodPost, "/activate", bytes.NewBufferString(`{"reason":" urgent deploy "}`))); got != "urgent deploy" {
		t.Fatalf("decodeMaintenanceReason = %q, want trimmed reason", got)
	}
	if got := decodeMaintenanceReason(httptest.NewRequest(http.MethodPost, "/activate", bytes.NewBufferString(`not-json`))); got != "" {
		t.Fatalf("decodeMaintenanceReason invalid JSON = %q, want empty", got)
	}
	if got := decodeMaintenanceReason(httptest.NewRequest(http.MethodPost, "/activate", nil)); got != "" {
		t.Fatalf("decodeMaintenanceReason nil body = %q, want empty", got)
	}
}

func TestSystemMaintenanceURLUUIDAndActorUserID(t *testing.T) {
	validID := "66666666-6666-6666-6666-666666666666"
	req := requestWithURLParam("/maintenance/"+validID, "id", validID, nil)
	id, ok := urlUUID(httptest.NewRecorder(), req, "id")
	if !ok || !id.Valid || pgUUIDString(id) != validID {
		t.Fatalf("urlUUID valid = (%s,%v), want %s/true", pgUUIDString(id), ok, validID)
	}

	badReq := requestWithURLParam("/maintenance/not-a-uuid", "id", "not-a-uuid", nil)
	badRec := httptest.NewRecorder()
	if _, ok := urlUUID(badRec, badReq, "id"); ok || badRec.Code != http.StatusBadRequest {
		t.Fatalf("urlUUID invalid ok/status = %v/%d, want false/%d", ok, badRec.Code, http.StatusBadRequest)
	}

	req = httptest.NewRequest(http.MethodGet, "/maintenance", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"uid": validID}))
	if got := actorUserID(req); !got.Valid || pgUUIDString(got) != validID {
		t.Fatalf("actorUserID uid = %s valid=%v, want %s", pgUUIDString(got), got.Valid, validID)
	}
	subID := "77777777-7777-7777-7777-777777777777"
	req = httptest.NewRequest(http.MethodGet, "/maintenance", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"sub": subID}))
	if got := actorUserID(req); !got.Valid || pgUUIDString(got) != subID {
		t.Fatalf("actorUserID sub = %s valid=%v, want %s", pgUUIDString(got), got.Valid, subID)
	}
}
