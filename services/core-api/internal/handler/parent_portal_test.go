package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestParentPortalPreviewParentsAndChildrenMapSafeDTOs(t *testing.T) {
	parentID := handlerTestUUID(231)
	childID := handlerTestUUID(232)
	classID := handlerTestUUID(233)
	svc := &fakeParentPortalSelfService{
		previewParentRows: []db.ListParentPortalPreviewParentsRow{{
			ID: parentID, Nama: "Wali A", Phone: "08123", LinkedStudentCount: 2,
		}},
		childrenRows: []db.ListParentChildrenRow{
			{ID: childID, Nis: "1001", Nama: "Siswa A", ClassID: classID, ClassName: pgtype.Text{String: "VII A", Valid: true}, Relationship: "ayah", IsPrimaryContact: true, Notes: "jemput sore"},
			{ID: handlerTestUUID(234), Nis: "1002", Nama: "Siswa B", Relationship: 42},
		},
	}
	h := &ParentPortal{svc: svc}

	rec := httptest.NewRecorder()
	h.PreviewParents(rec, httptest.NewRequest(http.MethodGet, "/api/portal/parent/preview/parents", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("PreviewParents() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var parentsPayload struct {
		Data struct {
			Preview bool `json:"preview"`
			Parents []struct {
				ID                 string `json:"id"`
				Nama               string `json:"nama"`
				Phone              string `json:"phone"`
				LinkedStudentCount int64  `json:"linked_student_count"`
			} `json:"parents"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parentsPayload); err != nil {
		t.Fatalf("PreviewParents() json unmarshal failed: %v", err)
	}
	if !parentsPayload.Data.Preview || len(parentsPayload.Data.Parents) != 1 || parentsPayload.Data.Parents[0].ID != parentID.String() || parentsPayload.Data.Parents[0].LinkedStudentCount != 2 {
		t.Fatalf("PreviewParents() payload = %+v, want mapped preview parent", parentsPayload.Data)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/preview/parents/"+parentID.String()+"/children", nil), "parentID", parentID.String())
	h.PreviewChildren(rec, req)
	if rec.Code != http.StatusOK || svc.childrenUserID != parentID {
		t.Fatalf("PreviewChildren() status/parentID = %d/%s, want 200/%s; body=%s", rec.Code, svc.childrenUserID.String(), parentID.String(), rec.Body.String())
	}
	var childrenPayload struct {
		Data struct {
			Preview  bool `json:"preview"`
			Children []struct {
				ID           string `json:"id"`
				ClassID      string `json:"class_id"`
				ClassName    string `json:"class_name"`
				Relationship string `json:"relationship"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &childrenPayload); err != nil {
		t.Fatalf("PreviewChildren() json unmarshal failed: %v", err)
	}
	if !childrenPayload.Data.Preview || len(childrenPayload.Data.Children) != 2 || childrenPayload.Data.Children[0].Relationship != "ayah" || childrenPayload.Data.Children[1].Relationship != "" {
		t.Fatalf("PreviewChildren() payload = %+v, want mapped children with non-string relationship blanked", childrenPayload.Data)
	}
	if childrenPayload.Data.Children[0].ClassID != classID.String() || childrenPayload.Data.Children[0].ClassName != "VII A" {
		t.Fatalf("PreviewChildren() class mapping = %+v, want class id/name", childrenPayload.Data.Children[0])
	}
}

func TestParentPortalPreviewRequestHelpersValidateRouteIDs(t *testing.T) {
	parentID := handlerTestUUID(235)
	studentID := handlerTestUUID(236)
	h := &ParentPortal{svc: &fakeParentPortalSelfService{}}
	req := withRouteParams(httptest.NewRequest(http.MethodGet, "/", nil), "parentID", parentID.String(), "studentID", studentID.String())
	rec := httptest.NewRecorder()

	gotParentID, gotStudentID, ok := h.previewChildRequest(rec, req)
	if !ok || gotParentID != parentID || gotStudentID != studentID || rec.Code != http.StatusOK {
		t.Fatalf("previewChildRequest(valid) = %s/%s/%v status %d, want route IDs/true/200", gotParentID.String(), gotStudentID.String(), ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	_, ok = h.previewParentID(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/", nil), "parentID", "not-a-uuid"))
	if ok || rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "ID orang tua tidak valid") {
		t.Fatalf("previewParentID(invalid) ok/status/body = %v/%d/%s, want false/400/parent error", ok, rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParams(httptest.NewRequest(http.MethodGet, "/", nil), "parentID", parentID.String(), "studentID", "bad-student")
	_, _, ok = h.previewChildRequest(rec, req)
	if ok || rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid student_id") {
		t.Fatalf("previewChildRequest(invalid student) ok/status/body = %v/%d/%s, want false/400/student error", ok, rec.Code, rec.Body.String())
	}
}

func TestParentPortalRemainingPreviewChildEndpointsUseParentAndStudentRouteIDs(t *testing.T) {
	parentID := handlerTestUUID(244)
	studentID := handlerTestUUID(245)
	scheduleRowID := handlerTestUUID(246)
	resultsSessionID := handlerTestUUID(247)
	svc := &fakeParentPortalSelfService{
		profileRow: db.GetParentPortalChildProfileRow{ID: studentID, Nama: "Anak Preview", Nis: "3001"},
		scheduleRows: []db.ListParentPortalChildTimetableRow{{
			ID:          scheduleRowID,
			DayOfWeek:   1,
			SubjectName: "IPA",
		}},
		resultsRows: []db.ListParentPortalChildExamSessionsRow{{
			SessionID:    resultsSessionID,
			SessionTitle: "Ujian Preview",
			PackageTitle: "Paket A",
		}},
	}
	h := &ParentPortal{svc: svc}

	tests := []struct {
		name         string
		handler      func(http.ResponseWriter, *http.Request)
		path         string
		wantFragment string
	}{
		{name: "profile", handler: h.PreviewChildProfile, path: "/api/portal/parent/preview/parents/" + parentID.String() + "/children/" + studentID.String() + "/profile", wantFragment: "Anak Preview"},
		{name: "schedule", handler: h.PreviewChildSchedule, path: "/api/portal/parent/preview/parents/" + parentID.String() + "/children/" + studentID.String() + "/schedule", wantFragment: "IPA"},
		{name: "results", handler: h.PreviewChildResults, path: "/api/portal/parent/preview/parents/" + parentID.String() + "/children/" + studentID.String() + "/results", wantFragment: "Ujian Preview"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := withRouteParams(httptest.NewRequest(http.MethodGet, tt.path, nil), "parentID", parentID.String(), "studentID", studentID.String())
			tt.handler(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"preview":true`) || !strings.Contains(rec.Body.String(), tt.wantFragment) {
				t.Fatalf("body = %s, want preview flag and %q", rec.Body.String(), tt.wantFragment)
			}
		})
	}
	if svc.profileUserID != parentID || svc.profileStudentID != studentID {
		t.Fatalf("PreviewChildProfile args = parent %s student %s, want %s/%s", svc.profileUserID.String(), svc.profileStudentID.String(), parentID.String(), studentID.String())
	}
	if svc.scheduleUserID != parentID || svc.scheduleStudentID != studentID {
		t.Fatalf("PreviewChildSchedule args = parent %s student %s, want %s/%s", svc.scheduleUserID.String(), svc.scheduleStudentID.String(), parentID.String(), studentID.String())
	}
	if svc.resultsUserID != parentID || svc.resultsStudentID != studentID {
		t.Fatalf("PreviewChildResults args = parent %s student %s, want %s/%s", svc.resultsUserID.String(), svc.resultsStudentID.String(), parentID.String(), studentID.String())
	}
}

func TestParentPortalHandlersUseAuthenticatedUserIDAndPathChildID(t *testing.T) {
	userID := handlerTestUUID(220)
	pathStudentID := handlerTestUUID(221)
	claimParentID := handlerTestUUID(222)
	queryStudentID := handlerTestUUID(223)
	svc := &fakeParentPortalSelfService{}
	h := &ParentPortal{svc: svc}
	req := withClaims(
		withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/children/"+pathStudentID.String()+"/profile?student_id="+queryStudentID.String(), nil), "studentID", pathStudentID.String()),
		jwt.MapClaims{
			"roles": []any{"ortu"},
			"sub":   userID.String(),
			"pid":   claimParentID.String(),
		},
	)
	rec := httptest.NewRecorder()

	h.ChildProfile(rec, req)

	if rec.Code != http.StatusOK || svc.profileUserID != userID || svc.profileStudentID != pathStudentID {
		t.Fatalf("ChildProfile() status/user/student = %d/%s/%s, want 200/%s/%s; body=%s", rec.Code, svc.profileUserID.String(), svc.profileStudentID.String(), userID.String(), pathStudentID.String(), rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), queryStudentID.String()) || strings.Contains(rec.Body.String(), claimParentID.String()) {
		t.Fatalf("ChildProfile() leaked query/claim scope IDs: %s", rec.Body.String())
	}
}

func TestParentPortalHandlersForbidUnlinkedChildAndRedactResultSecrets(t *testing.T) {
	userID := handlerTestUUID(224)
	studentID := handlerTestUUID(225)

	rec := httptest.NewRecorder()
	(&ParentPortal{svc: &fakeParentPortalSelfService{scheduleErr: domain.ErrForbidden}}).ChildSchedule(
		rec,
		withClaims(
			withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/children/"+studentID.String()+"/schedule", nil), "studentID", studentID.String()),
			jwt.MapClaims{"roles": []any{"ortu"}, "sub": userID.String()},
		),
	)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ChildSchedule(unlinked) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	resultsSvc := &fakeParentPortalSelfService{
		resultsRows: []db.ListParentPortalChildExamSessionsRow{{
			SessionID:    handlerTestUUID(226),
			SessionTitle: "Ujian IPA",
			PackageTitle: "Paket IPA",
		}},
	}
	rec = httptest.NewRecorder()
	(&ParentPortal{svc: resultsSvc}).ChildResults(
		rec,
		withClaims(
			withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/children/"+studentID.String()+"/results", nil), "studentID", studentID.String()),
			jwt.MapClaims{"permissions": []any{"parent_portal.read"}, "sub": userID.String()},
		),
	)
	if rec.Code != http.StatusOK || resultsSvc.resultsUserID != userID || resultsSvc.resultsStudentID != studentID {
		t.Fatalf("ChildResults() status/user/student = %d/%s/%s, want 200/%s/%s; body=%s", rec.Code, resultsSvc.resultsUserID.String(), resultsSvc.resultsStudentID.String(), userID.String(), studentID.String(), rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "token") || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("ChildResults() exposed secret result data: %s", rec.Body.String())
	}
}

func TestParentPortalHandlersRequireAuthenticatedClaims(t *testing.T) {
	rec := httptest.NewRecorder()
	(&ParentPortal{svc: &fakeParentPortalSelfService{}}).Children(rec, httptest.NewRequest(http.MethodGet, "/api/portal/parent/children", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Children(no claims) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

type fakeParentPortalSelfService struct {
	previewParentRows []db.ListParentPortalPreviewParentsRow
	previewParentErr  error

	childrenUserID pgtype.UUID
	childrenRows   []db.ListParentChildrenRow
	childrenErr    error

	profileUserID    pgtype.UUID
	profileStudentID pgtype.UUID
	profileRow       db.GetParentPortalChildProfileRow
	profileErr       error

	scheduleUserID    pgtype.UUID
	scheduleStudentID pgtype.UUID
	scheduleRows      []db.ListParentPortalChildTimetableRow
	scheduleErr       error

	resultsUserID    pgtype.UUID
	resultsStudentID pgtype.UUID
	resultsRows      []db.ListParentPortalChildExamSessionsRow
	resultsErr       error
}

func (f *fakeParentPortalSelfService) ListParentPortalPreviewParents(ctx context.Context) ([]db.ListParentPortalPreviewParentsRow, error) {
	if f.previewParentRows != nil || f.previewParentErr != nil {
		return f.previewParentRows, f.previewParentErr
	}
	return []db.ListParentPortalPreviewParentsRow{{ID: handlerTestUUID(230), Nama: "Wali A", LinkedStudentCount: 1}}, nil
}

func (f *fakeParentPortalSelfService) Children(ctx context.Context, userID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	f.childrenUserID = userID
	return f.childrenRows, f.childrenErr
}

func (f *fakeParentPortalSelfService) ChildrenByParentID(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	return f.Children(ctx, parentID)
}

func (f *fakeParentPortalSelfService) ChildProfile(ctx context.Context, userID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error) {
	f.profileUserID = userID
	f.profileStudentID = studentID
	if f.profileErr != nil {
		return db.GetParentPortalChildProfileRow{}, f.profileErr
	}
	if f.profileRow.ID.Valid {
		return f.profileRow, nil
	}
	return db.GetParentPortalChildProfileRow{ID: studentID, Nama: "Siswa A"}, nil
}

func (f *fakeParentPortalSelfService) ChildProfileByParentID(ctx context.Context, parentID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error) {
	return f.ChildProfile(ctx, parentID, studentID)
}

func (f *fakeParentPortalSelfService) ChildSchedule(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error) {
	f.scheduleUserID = userID
	f.scheduleStudentID = studentID
	return f.scheduleRows, f.scheduleErr
}

func (f *fakeParentPortalSelfService) ChildScheduleByParentID(ctx context.Context, parentID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error) {
	return f.ChildSchedule(ctx, parentID, studentID)
}

func (f *fakeParentPortalSelfService) ChildResults(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error) {
	f.resultsUserID = userID
	f.resultsStudentID = studentID
	return f.resultsRows, f.resultsErr
}

func (f *fakeParentPortalSelfService) ChildResultsByParentID(ctx context.Context, parentID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error) {
	return f.ChildResults(ctx, parentID, studentID)
}
