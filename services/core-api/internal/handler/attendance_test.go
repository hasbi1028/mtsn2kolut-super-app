package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeAttendanceHandlerService struct {
	listLimit  int32
	listOffset int32
	listRows   []db.ListAttendanceRow
	listTotal  int64
	listErr    error

	rangeStart pgtype.Date
	rangeEnd   pgtype.Date
	rangeRows  []db.ListAttendanceInRangeRow
	rangeErr   error

	dateArg  pgtype.Date
	dateRows []db.ListAttendanceByDateRow
	dateErr  error

	summaryStart pgtype.Date
	summaryEnd   pgtype.Date
	summaryRows  []db.GetMonthlyAttendanceSummaryRow
	summaryErr   error

	employeeID     pgtype.UUID
	employeeLimit  int32
	employeeOffset int32
	employeeRows   []db.ListAttendanceByEmployeeRow
	employeeErr    error
}

func (f *fakeAttendanceHandlerService) List(ctx context.Context, limit, offset int32) ([]db.ListAttendanceRow, int64, error) {
	f.listLimit = limit
	f.listOffset = offset
	return f.listRows, f.listTotal, f.listErr
}

func (f *fakeAttendanceHandlerService) ListInRange(ctx context.Context, start, end pgtype.Date) ([]db.ListAttendanceInRangeRow, error) {
	f.rangeStart = start
	f.rangeEnd = end
	return f.rangeRows, f.rangeErr
}

func (f *fakeAttendanceHandlerService) ByDate(ctx context.Context, date pgtype.Date) ([]db.ListAttendanceByDateRow, error) {
	f.dateArg = date
	return f.dateRows, f.dateErr
}

func (f *fakeAttendanceHandlerService) GetSummary(ctx context.Context, start, end pgtype.Date) ([]db.GetMonthlyAttendanceSummaryRow, error) {
	f.summaryStart = start
	f.summaryEnd = end
	return f.summaryRows, f.summaryErr
}

func (f *fakeAttendanceHandlerService) ByEmployee(ctx context.Context, empID pgtype.UUID, limit, offset int32) ([]db.ListAttendanceByEmployeeRow, error) {
	f.employeeID = empID
	f.employeeLimit = limit
	f.employeeOffset = offset
	return f.employeeRows, f.employeeErr
}

func TestAttendanceHandlersForwardValidRequests(t *testing.T) {
	employeeID := handlerTestUUID(161)
	svc := &fakeAttendanceHandlerService{
		listRows:     []db.ListAttendanceRow{{EmployeeID: employeeID, EmployeeNama: "Guru"}},
		listTotal:    12,
		rangeRows:    []db.ListAttendanceInRangeRow{{EmployeeID: employeeID, EmployeeNama: "Guru"}},
		dateRows:     []db.ListAttendanceByDateRow{{EmployeeID: employeeID, EmployeeNama: "Guru"}},
		summaryRows:  []db.GetMonthlyAttendanceSummaryRow{{EmployeeID: employeeID, EmployeeNama: "Guru"}},
		employeeRows: []db.ListAttendanceByEmployeeRow{{EmployeeID: employeeID, EmployeeNama: "Guru"}},
	}
	h := &PusakaAttendance{svc: svc}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/attendance?page=3&per_page=25", ""))
	if rec.Code != http.StatusOK || svc.listLimit != 25 || svc.listOffset != 50 {
		t.Fatalf("List() status/limit/offset = %d/%d/%d, want 200/25/50", rec.Code, svc.listLimit, svc.listOffset)
	}

	rec = httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/attendance?start_date=2026-05-01&end_date=2026-05-31", ""))
	if rec.Code != http.StatusOK || !svc.rangeStart.Valid || !svc.rangeEnd.Valid || svc.rangeStart.Time.Format("2006-01-02") != "2026-05-01" || svc.rangeEnd.Time.Format("2006-01-02") != "2026-05-31" {
		t.Fatalf("List(range) status/start/end = %d/%v/%v, want parsed date range", rec.Code, svc.rangeStart, svc.rangeEnd)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodGet, "/attendance/date/2026-05-02", ""), "date", "2026-05-02")
	h.ByDate(rec, req)
	if rec.Code != http.StatusOK || !svc.dateArg.Valid || svc.dateArg.Time.Format("2006-01-02") != "2026-05-02" {
		t.Fatalf("ByDate() status/date = %d/%v, want 200/2026-05-02", rec.Code, svc.dateArg)
	}

	rec = httptest.NewRecorder()
	h.GetSummary(rec, adminRequest(http.MethodGet, "/attendance/summary?start_date=2026-05-01&end_date=2026-05-31", ""))
	if rec.Code != http.StatusOK || !svc.summaryStart.Valid || !svc.summaryEnd.Valid {
		t.Fatalf("GetSummary() status/start/end = %d/%v/%v, want parsed range", rec.Code, svc.summaryStart, svc.summaryEnd)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodGet, "/attendance/employees/"+employeeID.String()+"?page=2&per_page=10", ""), "id", employeeID.String())
	h.ByEmployee(rec, req)
	if rec.Code != http.StatusOK || svc.employeeID != employeeID || svc.employeeLimit != 10 || svc.employeeOffset != 10 {
		t.Fatalf("ByEmployee() status/id/limit/offset = %d/%v/%d/%d, want 200/%v/10/10", rec.Code, svc.employeeID, svc.employeeLimit, svc.employeeOffset, employeeID)
	}
}

func TestAttendanceHandlersRejectInvalidRequests(t *testing.T) {
	h := &PusakaAttendance{svc: &fakeAttendanceHandlerService{}}
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		method     string
		target     string
		routeKey   string
		routeValue string
		wantStatus int
	}{
		{name: "list invalid start", fn: h.List, method: http.MethodGet, target: "/attendance?start_date=bad&end_date=2026-05-31", wantStatus: http.StatusBadRequest},
		{name: "list invalid end", fn: h.List, method: http.MethodGet, target: "/attendance?start_date=2026-05-01&end_date=bad", wantStatus: http.StatusBadRequest},
		{name: "by date invalid", fn: h.ByDate, method: http.MethodGet, target: "/attendance/date/bad", routeKey: "date", routeValue: "bad", wantStatus: http.StatusBadRequest},
		{name: "summary missing dates", fn: h.GetSummary, method: http.MethodGet, target: "/attendance/summary", wantStatus: http.StatusBadRequest},
		{name: "summary invalid start", fn: h.GetSummary, method: http.MethodGet, target: "/attendance/summary?start_date=bad&end_date=2026-05-31", wantStatus: http.StatusBadRequest},
		{name: "summary invalid end", fn: h.GetSummary, method: http.MethodGet, target: "/attendance/summary?start_date=2026-05-01&end_date=bad", wantStatus: http.StatusBadRequest},
		{name: "by employee invalid id", fn: h.ByEmployee, method: http.MethodGet, target: "/attendance/employees/bad", routeKey: "id", routeValue: "bad", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(tt.method, tt.target, "")
			if tt.routeKey != "" {
				req = withRouteParam(req, tt.routeKey, tt.routeValue)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestAttendanceHandlersMapServiceErrorsAndForbidden(t *testing.T) {
	employeeID := handlerTestUUID(162)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*PusakaAttendance, http.ResponseWriter, *http.Request)
		svc        *fakeAttendanceHandlerService
		req        *http.Request
		wantStatus int
	}{
		{name: "list forbidden", handler: (*PusakaAttendance).List, req: httptest.NewRequest(http.MethodGet, "/attendance", strings.NewReader("")), wantStatus: http.StatusForbidden},
		{name: "list service error", handler: (*PusakaAttendance).List, svc: &fakeAttendanceHandlerService{listErr: errDB}, req: adminRequest(http.MethodGet, "/attendance", ""), wantStatus: http.StatusInternalServerError},
		{name: "range service error", handler: (*PusakaAttendance).List, svc: &fakeAttendanceHandlerService{rangeErr: errDB}, req: adminRequest(http.MethodGet, "/attendance?start_date=2026-05-01&end_date=2026-05-31", ""), wantStatus: http.StatusInternalServerError},
		{name: "date forbidden", handler: (*PusakaAttendance).ByDate, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/attendance/date/2026-05-02", nil), "date", "2026-05-02"), wantStatus: http.StatusForbidden},
		{name: "date service error", handler: (*PusakaAttendance).ByDate, svc: &fakeAttendanceHandlerService{dateErr: errDB}, req: withRouteParam(adminRequest(http.MethodGet, "/attendance/date/2026-05-02", ""), "date", "2026-05-02"), wantStatus: http.StatusInternalServerError},
		{name: "summary forbidden", handler: (*PusakaAttendance).GetSummary, req: httptest.NewRequest(http.MethodGet, "/attendance/summary", nil), wantStatus: http.StatusForbidden},
		{name: "summary service error", handler: (*PusakaAttendance).GetSummary, svc: &fakeAttendanceHandlerService{summaryErr: errDB}, req: adminRequest(http.MethodGet, "/attendance/summary?start_date=2026-05-01&end_date=2026-05-31", ""), wantStatus: http.StatusInternalServerError},
		{name: "employee forbidden", handler: (*PusakaAttendance).ByEmployee, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/attendance/employees/"+employeeID.String(), nil), "id", employeeID.String()), wantStatus: http.StatusForbidden},
		{name: "employee service error", handler: (*PusakaAttendance).ByEmployee, svc: &fakeAttendanceHandlerService{employeeErr: errDB}, req: withRouteParam(adminRequest(http.MethodGet, "/attendance/employees/"+employeeID.String(), ""), "id", employeeID.String()), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeAttendanceHandlerService{}
			}
			rec := httptest.NewRecorder()
			tt.handler(&PusakaAttendance{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
