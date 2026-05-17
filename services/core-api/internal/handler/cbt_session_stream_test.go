package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type cbtSessionStreamRecorder struct {
	*httptest.ResponseRecorder
	flushes int
}

func (r *cbtSessionStreamRecorder) Flush() {
	r.flushes++
	r.ResponseRecorder.Flush()
}

type cbtSessionNoFlushWriter struct {
	header http.Header
	code   int
	body   strings.Builder
}

func (w *cbtSessionNoFlushWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *cbtSessionNoFlushWriter) WriteHeader(status int) {
	w.code = status
}

func (w *cbtSessionNoFlushWriter) Write(p []byte) (int, error) {
	return w.body.Write(p)
}

func cbtStreamTestRequest(method, target string, claims jwt.MapClaims, params ...string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	if claims != nil {
		req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
	}
	if len(params) > 0 {
		rctx := chi.NewRouteContext()
		for i := 0; i+1 < len(params); i += 2 {
			rctx.URLParams.Add(params[i], params[i+1])
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	return req
}

func cbtCanceledStreamRequest(method, target string, claims jwt.MapClaims, params ...string) *http.Request {
	req := cbtStreamTestRequest(method, target, claims, params...)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	return req.WithContext(ctx)
}

func TestCbtSessionStreamProctoringLiveSummaryWritesReadyAndSummaryThenExitsOnCanceledContext(t *testing.T) {
	sessionID := handlerTestUUID(10)
	fake := &fakeCbtSessionService{
		liveSummary: service.CbtProctoringLiveSummary{
			SessionID: sessionID.String(),
			Counts:    map[string]int{"participants": 2, "needs_action": 1},
		},
	}
	h := &CbtSession{svc: fake}
	req := cbtCanceledStreamRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/live/stream", jwt.MapClaims{"roles": []any{"admin"}}, "id", sessionID.String())
	rec := &cbtSessionStreamRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.StreamProctoringLiveSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("StreamProctoringLiveSummary status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want SSE", got)
	}
	if rec.Header().Get("Cache-Control") != "no-cache, no-transform" || rec.Header().Get("X-Accel-Buffering") != "no" {
		t.Fatalf("stream headers = %#v, want no-cache/no-buffer", rec.Header())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: ready\n") || !strings.Contains(body, `"mode":"live-summary"`) {
		t.Fatalf("stream body missing ready event: %s", body)
	}
	if !strings.Contains(body, "event: live_summary\n") || !strings.Contains(body, `"participants":2`) || !strings.Contains(body, `"needs_action":1`) {
		t.Fatalf("stream body missing live summary: %s", body)
	}
	if fake.liveSummarySessionID != sessionID || fake.liveSummaryRoomID.Valid {
		t.Fatalf("live summary args = %v/%v, want session and empty room", fake.liveSummarySessionID, fake.liveSummaryRoomID)
	}
	if rec.flushes < 2 {
		t.Fatalf("flushes = %d, want ready and summary flushed", rec.flushes)
	}
}

func TestCbtSessionStreamRoomProctoringLiveSummaryScopesRoomAndAssignedProctor(t *testing.T) {
	sessionID := handlerTestUUID(11)
	roomID := handlerTestUUID(12)
	employeeID := handlerTestUUID(13)
	fake := &fakeCbtSessionService{
		roomProctorAllowed: true,
		liveSummary: service.CbtProctoringLiveSummary{
			SessionID: sessionID.String(),
			RoomID:    roomID.String(),
			Counts:    map[string]int{"online": 3},
		},
	}
	h := &CbtSession{svc: fake}
	req := cbtCanceledStreamRequest(
		http.MethodGet,
		"/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctoring/live/stream",
		jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()},
		"id", sessionID.String(), "rid", roomID.String(),
	)
	rec := &cbtSessionStreamRecorder{ResponseRecorder: httptest.NewRecorder()}

	h.StreamRoomProctoringLiveSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("StreamRoomProctoringLiveSummary status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"room_id":"`+roomID.String()+`"`) || !strings.Contains(body, `"online":3`) {
		t.Fatalf("room stream body = %s, want ready room_id and room summary", body)
	}
	if fake.hasRoomSessionID != sessionID || fake.hasRoomID != roomID || fake.roomProctorSessionID != sessionID || fake.roomProctorRoomID != roomID || fake.roomProctorEmployeeID != employeeID {
		t.Fatalf("room stream scope args = hasRoom:%v/%v proctor:%v/%v/%v, want session/room/employee", fake.hasRoomSessionID, fake.hasRoomID, fake.roomProctorSessionID, fake.roomProctorRoomID, fake.roomProctorEmployeeID)
	}
	if fake.liveSummarySessionID != sessionID || fake.liveSummaryRoomID != roomID {
		t.Fatalf("room live summary args = %v/%v, want session/room", fake.liveSummarySessionID, fake.liveSummaryRoomID)
	}
}

func TestCbtSessionStreamProctoringLiveSummaryErrorBranches(t *testing.T) {
	sessionID := handlerTestUUID(14)
	errDB := errors.New("db down")

	t.Run("invalid session id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		(&CbtSession{svc: &fakeCbtSessionService{}}).StreamProctoringLiveSummary(rec, cbtStreamTestRequest(http.MethodGet, "/api/cbt/sessions/bad/proctoring/live/stream", jwt.MapClaims{"roles": []any{"admin"}}, "id", "bad"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("teacher denied before stream", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := cbtStreamTestRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/live/stream", jwt.MapClaims{"roles": []any{"guru"}, "eid": handlerTestUUID(15).String()}, "id", sessionID.String())
		(&CbtSession{svc: &fakeCbtSessionService{checkAllowed: false}}).StreamProctoringLiveSummary(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("writer without flusher", func(t *testing.T) {
		writer := &cbtSessionNoFlushWriter{}
		req := cbtStreamTestRequest(http.MethodGet, "/", jwt.MapClaims{"roles": []any{"admin"}})
		(&CbtSession{svc: &fakeCbtSessionService{}}).streamProctoringLiveSummary(writer, req, sessionID, pgtype.UUID{})
		if writer.code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500; body=%s", writer.code, writer.body.String())
		}
	})

	t.Run("summary error emits stream_error", func(t *testing.T) {
		rec := &cbtSessionStreamRecorder{ResponseRecorder: httptest.NewRecorder()}
		req := cbtCanceledStreamRequest(http.MethodGet, "/", jwt.MapClaims{"roles": []any{"admin"}})
		(&CbtSession{svc: &fakeCbtSessionService{liveSummaryErr: errDB}}).streamProctoringLiveSummary(rec, req, sessionID, pgtype.UUID{})
		body := rec.Body.String()
		if rec.Code != http.StatusOK || !strings.Contains(body, "event: stream_error\n") || !strings.Contains(body, "db down") {
			t.Fatalf("summary error status/body = %d/%s, want stream_error event", rec.Code, body)
		}
	})
}
