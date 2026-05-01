package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeLetterService struct {
	*service.Letter

	classifications     []db.LetterClassification
	classificationsErr  error
	incomingSearch      string
	incomingStatus      string
	incomingRows        []db.ListIncomingLettersRow
	incomingErr         error
	createIncomingArg   service.CreateIncomingParams
	createIncomingRow   db.IncomingLetter
	createIncomingErr   error
	getIncomingID       pgtype.UUID
	getIncomingRow      db.GetIncomingLetterRow
	getIncomingErr      error
	updateIncomingArg   service.UpdateIncomingParams
	updateIncomingRow   db.IncomingLetter
	updateIncomingErr   error
	incomingStatusID    pgtype.UUID
	incomingStatusValue string
	incomingStatusRow   db.IncomingLetter
	incomingStatusErr   error
	deleteIncomingID    pgtype.UUID
	deleteIncomingErr   error
	outgoingSearch      string
	outgoingRows        []db.ListOutgoingLettersRow
	outgoingErr         error
	createOutgoingArg   service.CreateOutgoingParams
	createOutgoingRow   db.OutgoingLetter
	createOutgoingErr   error
	getOutgoingID       pgtype.UUID
	getOutgoingRow      db.GetOutgoingLetterRow
	getOutgoingErr      error
	updateOutgoingArg   service.UpdateOutgoingParams
	updateOutgoingRow   db.OutgoingLetter
	updateOutgoingErr   error
	deleteOutgoingID    pgtype.UUID
	deleteOutgoingErr   error
	previewCode         string
	previewTanggal      string
	preview             string
	previewErr          error
	dispositionLetterID string
	dispositionStatus   string
	dispositionRows     []db.ListDispositionsRow
	dispositionErr      error
	createDispArg       service.CreateDispositionParams
	createDispRow       db.LetterDisposition
	createDispErr       error
	getDispID           pgtype.UUID
	getDispRow          db.GetDispositionRow
	getDispErr          error
	updateDispArg       service.UpdateDispositionParams
	updateDispRow       db.LetterDisposition
	updateDispErr       error
	deleteDispID        pgtype.UUID
	deleteDispErr       error
}

func (f *fakeLetterService) ListClassifications(context.Context) ([]db.LetterClassification, error) {
	return f.classifications, f.classificationsErr
}

func (f *fakeLetterService) ListIncoming(_ context.Context, search, filterStatus string) ([]db.ListIncomingLettersRow, error) {
	f.incomingSearch = search
	f.incomingStatus = filterStatus
	return f.incomingRows, f.incomingErr
}

func (f *fakeLetterService) CreateIncoming(_ context.Context, p service.CreateIncomingParams) (db.IncomingLetter, error) {
	f.createIncomingArg = p
	return f.createIncomingRow, f.createIncomingErr
}

func (f *fakeLetterService) GetIncoming(_ context.Context, id pgtype.UUID) (db.GetIncomingLetterRow, error) {
	f.getIncomingID = id
	return f.getIncomingRow, f.getIncomingErr
}

func (f *fakeLetterService) UpdateIncoming(_ context.Context, p service.UpdateIncomingParams) (db.IncomingLetter, error) {
	f.updateIncomingArg = p
	return f.updateIncomingRow, f.updateIncomingErr
}

func (f *fakeLetterService) UpdateIncomingStatus(_ context.Context, id pgtype.UUID, status string) (db.IncomingLetter, error) {
	f.incomingStatusID = id
	f.incomingStatusValue = status
	return f.incomingStatusRow, f.incomingStatusErr
}

func (f *fakeLetterService) DeleteIncoming(_ context.Context, id pgtype.UUID) error {
	f.deleteIncomingID = id
	return f.deleteIncomingErr
}

func (f *fakeLetterService) ListOutgoing(_ context.Context, search string) ([]db.ListOutgoingLettersRow, error) {
	f.outgoingSearch = search
	return f.outgoingRows, f.outgoingErr
}

func (f *fakeLetterService) CreateOutgoing(_ context.Context, p service.CreateOutgoingParams) (db.OutgoingLetter, error) {
	f.createOutgoingArg = p
	return f.createOutgoingRow, f.createOutgoingErr
}

func (f *fakeLetterService) GetOutgoing(_ context.Context, id pgtype.UUID) (db.GetOutgoingLetterRow, error) {
	f.getOutgoingID = id
	return f.getOutgoingRow, f.getOutgoingErr
}

func (f *fakeLetterService) UpdateOutgoing(_ context.Context, p service.UpdateOutgoingParams) (db.OutgoingLetter, error) {
	f.updateOutgoingArg = p
	return f.updateOutgoingRow, f.updateOutgoingErr
}

func (f *fakeLetterService) DeleteOutgoing(_ context.Context, id pgtype.UUID) error {
	f.deleteOutgoingID = id
	return f.deleteOutgoingErr
}

func (f *fakeLetterService) PreviewOutgoingNumber(classificationCode, tanggalSurat string) (string, error) {
	f.previewCode = classificationCode
	f.previewTanggal = tanggalSurat
	return f.preview, f.previewErr
}

func (f *fakeLetterService) ListDispositions(_ context.Context, incomingLetterID, filterStatus string) ([]db.ListDispositionsRow, error) {
	f.dispositionLetterID = incomingLetterID
	f.dispositionStatus = filterStatus
	return f.dispositionRows, f.dispositionErr
}

func (f *fakeLetterService) CreateDisposition(_ context.Context, p service.CreateDispositionParams) (db.LetterDisposition, error) {
	f.createDispArg = p
	return f.createDispRow, f.createDispErr
}

func (f *fakeLetterService) GetDisposition(_ context.Context, id pgtype.UUID) (db.GetDispositionRow, error) {
	f.getDispID = id
	return f.getDispRow, f.getDispErr
}

func (f *fakeLetterService) UpdateDisposition(_ context.Context, p service.UpdateDispositionParams) (db.LetterDisposition, error) {
	f.updateDispArg = p
	return f.updateDispRow, f.updateDispErr
}

func (f *fakeLetterService) DeleteDisposition(_ context.Context, id pgtype.UUID) error {
	f.deleteDispID = id
	return f.deleteDispErr
}

func tuStaffRequest(method, target, body string, employeeID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"staf"},
		"eid":   employeeID.String(),
	})
}

func TestLetterSuccessHandlersForwardPayloads(t *testing.T) {
	incomingID := handlerTestUUID(140)
	outgoingID := handlerTestUUID(141)
	dispositionID := handlerTestUUID(142)
	employeeID := handlerTestUUID(143)
	assigneeID := handlerTestUUID(144)
	fake := &fakeLetterService{
		Letter:            &service.Letter{},
		classifications:   []db.LetterClassification{{Code: "420", Name: "Pendidikan", IsActive: true}},
		incomingRows:      []db.ListIncomingLettersRow{{ID: incomingID, NomorSurat: "01/IN", Perihal: "Undangan"}},
		createIncomingRow: db.IncomingLetter{ID: incomingID, NomorSurat: "01/IN", Perihal: "Undangan"},
		getIncomingRow:    db.GetIncomingLetterRow{ID: incomingID, NomorSurat: "01/IN", Perihal: "Undangan"},
		updateIncomingRow: db.IncomingLetter{ID: incomingID, NomorSurat: "02/IN", Perihal: "Undangan revisi"},
		incomingStatusRow: db.IncomingLetter{ID: incomingID, Status: db.LetterStatusSelesai},
		outgoingRows:      []db.ListOutgoingLettersRow{{ID: outgoingID, NomorSurat: "001/420", Perihal: "Balasan"}},
		createOutgoingRow: db.OutgoingLetter{ID: outgoingID, NomorSurat: "001/420", Perihal: "Balasan"},
		getOutgoingRow:    db.GetOutgoingLetterRow{ID: outgoingID, NomorSurat: "001/420", Perihal: "Balasan"},
		updateOutgoingRow: db.OutgoingLetter{ID: outgoingID, Perihal: "Balasan revisi"},
		preview:           "XXX/420/MTs.20.05/V/2026",
		dispositionRows:   []db.ListDispositionsRow{{ID: dispositionID, IncomingLetterID: incomingID, Instruksi: "Tindak lanjuti"}},
		createDispRow:     db.LetterDisposition{ID: dispositionID, IncomingLetterID: incomingID, AssigneeEmployeeID: assigneeID},
		getDispRow:        db.GetDispositionRow{ID: dispositionID, IncomingLetterID: incomingID, Instruksi: "Tindak lanjuti"},
		updateDispRow:     db.LetterDisposition{ID: dispositionID, Status: db.DispositionStatusSelesai},
	}
	h := &Letter{svc: fake}
	req := func(method, target, body string) *http.Request {
		return tuStaffRequest(method, target, body, employeeID)
	}

	rec := httptest.NewRecorder()
	h.ListClassifications(rec, req(http.MethodGet, "/api/tu/surat/klasifikasi", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Pendidikan") {
		t.Fatalf("ListClassifications status/body = %d/%s, want classification", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListIncoming(rec, req(http.MethodGet, "/api/tu/surat/incoming?search=undangan&status=baru", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListIncoming status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.incomingSearch != "undangan" || fake.incomingStatus != "baru" {
		t.Fatalf("ListIncoming filters = (%q, %q), want query filters", fake.incomingSearch, fake.incomingStatus)
	}

	rec = httptest.NewRecorder()
	h.CreateIncoming(rec, req(http.MethodPost, "/api/tu/surat/incoming", `{"nomor_surat":"01/IN","tanggal_surat":"2026-05-01","tanggal_terima":"2026-05-02","asal":"Kemenag","perihal":"Undangan","sifat":"biasa","catatan":"Catatan","received_by_employee_id":"`+employeeID.String()+`"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateIncoming status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createIncomingArg.NomorSurat != "01/IN" || fake.createIncomingArg.Asal != "Kemenag" || fake.createIncomingArg.ReceivedByEmployeeID != employeeID.String() {
		t.Fatalf("CreateIncoming arg = %+v, want decoded payload", fake.createIncomingArg)
	}

	rec = httptest.NewRecorder()
	h.GetIncoming(rec, withRouteParam(req(http.MethodGet, "/api/tu/surat/incoming/"+incomingID.String(), ""), "id", incomingID.String()))
	if rec.Code != http.StatusOK || fake.getIncomingID != incomingID {
		t.Fatalf("GetIncoming status/id = %d/%v, want 200/%v", rec.Code, fake.getIncomingID, incomingID)
	}

	rec = httptest.NewRecorder()
	h.UpdateIncoming(rec, withRouteParam(req(http.MethodPut, "/api/tu/surat/incoming/"+incomingID.String(), `{"nomor_surat":"02/IN","tanggal_surat":"2026-05-03","tanggal_terima":"2026-05-04","asal":"Kemenag","perihal":"Undangan revisi","sifat":"penting","catatan":"Revisi"}`), "id", incomingID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateIncoming status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateIncomingArg.ID != incomingID.String() || fake.updateIncomingArg.NomorSurat != "02/IN" {
		t.Fatalf("UpdateIncoming arg = %+v, want route id and payload", fake.updateIncomingArg)
	}

	rec = httptest.NewRecorder()
	h.UpdateIncomingStatus(rec, withRouteParam(req(http.MethodPatch, "/api/tu/surat/incoming/"+incomingID.String()+"/status", `{"status":"selesai"}`), "id", incomingID.String()))
	if rec.Code != http.StatusOK || fake.incomingStatusID != incomingID || fake.incomingStatusValue != "selesai" {
		t.Fatalf("UpdateIncomingStatus status/id/value = %d/%v/%q", rec.Code, fake.incomingStatusID, fake.incomingStatusValue)
	}

	rec = httptest.NewRecorder()
	h.DeleteIncoming(rec, withRouteParam(req(http.MethodDelete, "/api/tu/surat/incoming/"+incomingID.String(), ""), "id", incomingID.String()))
	if rec.Code != http.StatusNoContent || fake.deleteIncomingID != incomingID {
		t.Fatalf("DeleteIncoming status/id = %d/%v, want 204/%v", rec.Code, fake.deleteIncomingID, incomingID)
	}

	rec = httptest.NewRecorder()
	h.ListOutgoing(rec, req(http.MethodGet, "/api/tu/surat/outgoing?search=balasan", ""))
	if rec.Code != http.StatusOK || fake.outgoingSearch != "balasan" {
		t.Fatalf("ListOutgoing status/search = %d/%q, want 200/balasan", rec.Code, fake.outgoingSearch)
	}

	rec = httptest.NewRecorder()
	h.CreateOutgoing(rec, req(http.MethodPost, "/api/tu/surat/outgoing", `{"classification_code":"420","tanggal_surat":"2026-05-05","tujuan":"Kemenag","perihal":"Balasan","sifat":"biasa","catatan":"Catatan","manual_nomor":"MANUAL-1"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateOutgoing status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createOutgoingArg.IssuedByEmployeeID != employeeID.String() || fake.createOutgoingArg.ClassificationCode != "420" || fake.createOutgoingArg.ManualNomor != "MANUAL-1" {
		t.Fatalf("CreateOutgoing arg = %+v, want employee and payload", fake.createOutgoingArg)
	}

	rec = httptest.NewRecorder()
	h.GetOutgoing(rec, withRouteParam(req(http.MethodGet, "/api/tu/surat/outgoing/"+outgoingID.String(), ""), "id", outgoingID.String()))
	if rec.Code != http.StatusOK || fake.getOutgoingID != outgoingID {
		t.Fatalf("GetOutgoing status/id = %d/%v, want 200/%v", rec.Code, fake.getOutgoingID, outgoingID)
	}

	rec = httptest.NewRecorder()
	h.UpdateOutgoing(rec, withRouteParam(req(http.MethodPut, "/api/tu/surat/outgoing/"+outgoingID.String(), `{"tanggal_surat":"2026-05-06","tujuan":"Kemenag","perihal":"Balasan revisi","sifat":"penting","catatan":"Ok"}`), "id", outgoingID.String()))
	if rec.Code != http.StatusOK || fake.updateOutgoingArg.ID != outgoingID.String() || fake.updateOutgoingArg.Perihal != "Balasan revisi" {
		t.Fatalf("UpdateOutgoing status/arg = %d/%+v, want route id and payload", rec.Code, fake.updateOutgoingArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteOutgoing(rec, withRouteParam(req(http.MethodDelete, "/api/tu/surat/outgoing/"+outgoingID.String(), ""), "id", outgoingID.String()))
	if rec.Code != http.StatusNoContent || fake.deleteOutgoingID != outgoingID {
		t.Fatalf("DeleteOutgoing status/id = %d/%v, want 204/%v", rec.Code, fake.deleteOutgoingID, outgoingID)
	}

	rec = httptest.NewRecorder()
	h.PreviewOutgoingNumber(rec, req(http.MethodGet, "/api/tu/surat/outgoing/preview-number?classification_code=420&tanggal=2026-05-07", ""))
	if rec.Code != http.StatusOK || fake.previewCode != "420" || fake.previewTanggal != "2026-05-07" || !strings.Contains(rec.Body.String(), fake.preview) {
		t.Fatalf("PreviewOutgoingNumber status/code/date/body = %d/%q/%q/%s", rec.Code, fake.previewCode, fake.previewTanggal, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListDispositions(rec, req(http.MethodGet, "/api/tu/surat/disposisi?incoming_letter_id="+incomingID.String()+"&status=terkirim", ""))
	if rec.Code != http.StatusOK || fake.dispositionLetterID != incomingID.String() || fake.dispositionStatus != "terkirim" {
		t.Fatalf("ListDispositions status/filters = %d/%q/%q", rec.Code, fake.dispositionLetterID, fake.dispositionStatus)
	}

	rec = httptest.NewRecorder()
	h.CreateDisposition(rec, req(http.MethodPost, "/api/tu/surat/disposisi", `{"incoming_letter_id":"`+incomingID.String()+`","assignee_employee_id":"`+assigneeID.String()+`","instruksi":"Tindak lanjuti"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateDisposition status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createDispArg.IncomingLetterID != incomingID.String() || fake.createDispArg.AssigneeEmployeeID != assigneeID.String() || fake.createDispArg.DisposedByEmployeeID != employeeID.String() {
		t.Fatalf("CreateDisposition arg = %+v, want ids and disposer", fake.createDispArg)
	}

	rec = httptest.NewRecorder()
	h.GetDisposition(rec, withRouteParam(req(http.MethodGet, "/api/tu/surat/disposisi/"+dispositionID.String(), ""), "id", dispositionID.String()))
	if rec.Code != http.StatusOK || fake.getDispID != dispositionID {
		t.Fatalf("GetDisposition status/id = %d/%v, want 200/%v", rec.Code, fake.getDispID, dispositionID)
	}

	rec = httptest.NewRecorder()
	h.UpdateDisposition(rec, withRouteParam(req(http.MethodPut, "/api/tu/surat/disposisi/"+dispositionID.String(), `{"instruksi":"Selesai","catatan_tindak_lanjut":"Sudah","status":"selesai"}`), "id", dispositionID.String()))
	if rec.Code != http.StatusOK || fake.updateDispArg.ID != dispositionID.String() || fake.updateDispArg.Status != "selesai" {
		t.Fatalf("UpdateDisposition status/arg = %d/%+v, want route id and payload", rec.Code, fake.updateDispArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteDisposition(rec, withRouteParam(req(http.MethodDelete, "/api/tu/surat/disposisi/"+dispositionID.String(), ""), "id", dispositionID.String()))
	if rec.Code != http.StatusNoContent || fake.deleteDispID != dispositionID {
		t.Fatalf("DeleteDisposition status/id = %d/%v, want 204/%v", rec.Code, fake.deleteDispID, dispositionID)
	}
}

func TestLetterMutationHandlersWriteAuditEvents(t *testing.T) {
	incomingID := handlerTestUUID(145)
	outgoingID := handlerTestUUID(146)
	dispositionID := handlerTestUUID(147)
	employeeID := handlerTestUUID(148)
	assigneeID := handlerTestUUID(149)
	fake := &fakeLetterService{
		Letter:            &service.Letter{},
		createIncomingRow: db.IncomingLetter{ID: incomingID, NomorSurat: "01/IN", Perihal: "Undangan", Asal: "Kemenag"},
		updateIncomingRow: db.IncomingLetter{ID: incomingID, NomorSurat: "02/IN", Perihal: "Undangan revisi", Asal: "Kemenag"},
		incomingStatusRow: db.IncomingLetter{ID: incomingID, Status: db.LetterStatusSelesai},
		createOutgoingRow: db.OutgoingLetter{ID: outgoingID, NomorSurat: "001/420", Perihal: "Balasan", Tujuan: "Kemenag"},
		updateOutgoingRow: db.OutgoingLetter{ID: outgoingID, NomorSurat: "001/420", Perihal: "Balasan revisi", Tujuan: "Kemenag"},
		createDispRow:     db.LetterDisposition{ID: dispositionID, IncomingLetterID: incomingID, AssigneeEmployeeID: assigneeID},
		updateDispRow:     db.LetterDisposition{ID: dispositionID, Status: db.DispositionStatusSelesai},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &Letter{svc: fake, audit: audit}
	req := func(method, target, body string) *http.Request {
		r := tuStaffRequest(method, target, body, employeeID)
		return withClaims(r, jwt.MapClaims{
			"roles": []any{"staf"},
			"eid":   employeeID.String(),
			"uid":   "01000000-0000-0000-0000-000000000000",
			"sub":   "01000000-0000-0000-0000-000000000000",
			"usr":   "staf.tu",
			"ssid":  "sess-letter-1",
		})
	}

	rec := httptest.NewRecorder()
	h.CreateIncoming(rec, req(http.MethodPost, "/api/tu/surat/incoming", `{"nomor_surat":"01/IN","tanggal_surat":"2026-05-01","tanggal_terima":"2026-05-02","asal":"Kemenag","perihal":"Undangan","sifat":"biasa","catatan":"Catatan","received_by_employee_id":"`+employeeID.String()+`"}`))
	rec = httptest.NewRecorder()
	h.UpdateIncoming(rec, withRouteParam(req(http.MethodPut, "/api/tu/surat/incoming/"+incomingID.String(), `{"nomor_surat":"02/IN","tanggal_surat":"2026-05-03","tanggal_terima":"2026-05-04","asal":"Kemenag","perihal":"Undangan revisi","sifat":"penting","catatan":"Revisi"}`), "id", incomingID.String()))
	rec = httptest.NewRecorder()
	h.UpdateIncomingStatus(rec, withRouteParam(req(http.MethodPatch, "/api/tu/surat/incoming/"+incomingID.String()+"/status", `{"status":"selesai"}`), "id", incomingID.String()))
	rec = httptest.NewRecorder()
	h.DeleteIncoming(rec, withRouteParam(req(http.MethodDelete, "/api/tu/surat/incoming/"+incomingID.String(), ""), "id", incomingID.String()))

	rec = httptest.NewRecorder()
	h.CreateOutgoing(rec, req(http.MethodPost, "/api/tu/surat/outgoing", `{"classification_code":"420","tanggal_surat":"2026-05-01","tujuan":"Kemenag","perihal":"Balasan","sifat":"biasa","catatan":"Catatan"}`))
	rec = httptest.NewRecorder()
	h.UpdateOutgoing(rec, withRouteParam(req(http.MethodPut, "/api/tu/surat/outgoing/"+outgoingID.String(), `{"tanggal_surat":"2026-05-03","tujuan":"Kemenag","perihal":"Balasan revisi","sifat":"penting","catatan":"Revisi"}`), "id", outgoingID.String()))
	rec = httptest.NewRecorder()
	h.DeleteOutgoing(rec, withRouteParam(req(http.MethodDelete, "/api/tu/surat/outgoing/"+outgoingID.String(), ""), "id", outgoingID.String()))

	rec = httptest.NewRecorder()
	h.CreateDisposition(rec, req(http.MethodPost, "/api/tu/surat/disposisi", `{"incoming_letter_id":"`+incomingID.String()+`","assignee_employee_id":"`+assigneeID.String()+`","instruksi":"Tindak lanjuti"}`))
	rec = httptest.NewRecorder()
	h.UpdateDisposition(rec, withRouteParam(req(http.MethodPut, "/api/tu/surat/disposisi/"+dispositionID.String(), `{"instruksi":"Update","catatan_tindak_lanjut":"Selesai","status":"selesai"}`), "id", dispositionID.String()))
	rec = httptest.NewRecorder()
	h.DeleteDisposition(rec, withRouteParam(req(http.MethodDelete, "/api/tu/surat/disposisi/"+dispositionID.String(), ""), "id", dispositionID.String()))

	if len(audit.entries) != 10 {
		t.Fatalf("audit entries = %d, want 10", len(audit.entries))
	}
	if audit.entries[0].Action != "LETTER_INCOMING_CREATE" || audit.entries[4].Action != "LETTER_OUTGOING_CREATE" || audit.entries[7].Action != "LETTER_DISPOSITION_CREATE" {
		t.Fatalf("unexpected audit actions = %+v", audit.entries)
	}
	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["perihal"] != "Undangan" || meta["username"] != "staf.tu" {
		t.Fatalf("incoming create metadata = %+v, want perihal/username", meta)
	}
	meta = mustAuditMetadataMap(t, audit.entries[3].Metadata)
	if meta["deleted_by"] != "staf.tu" {
		t.Fatalf("incoming delete metadata = %+v, want deleted_by", meta)
	}
	meta = mustAuditMetadataMap(t, audit.entries[8].Metadata)
	if meta["status"] != string(db.DispositionStatusSelesai) {
		t.Fatalf("disposition update metadata = %+v, want status", meta)
	}
}

func TestLetterHandlersMapServiceErrors(t *testing.T) {
	id := handlerTestUUID(145)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*Letter, http.ResponseWriter, *http.Request)
		svc        *fakeLetterService
		req        *http.Request
		wantStatus int
	}{
		{name: "classifications", handler: (*Letter).ListClassifications, svc: &fakeLetterService{Letter: &service.Letter{}, classificationsErr: errDB}, req: adminRequest(http.MethodGet, "/api/tu/surat/klasifikasi", ""), wantStatus: http.StatusInternalServerError},
		{name: "incoming list", handler: (*Letter).ListIncoming, svc: &fakeLetterService{Letter: &service.Letter{}, incomingErr: errDB}, req: adminRequest(http.MethodGet, "/api/tu/surat/incoming", ""), wantStatus: http.StatusInternalServerError},
		{name: "incoming create validation", handler: (*Letter).CreateIncoming, svc: &fakeLetterService{Letter: &service.Letter{}, createIncomingErr: errors.New("nomor surat wajib diisi")}, req: adminRequest(http.MethodPost, "/api/tu/surat/incoming", `{}`), wantStatus: http.StatusBadRequest},
		{name: "incoming get", handler: (*Letter).GetIncoming, svc: &fakeLetterService{Letter: &service.Letter{}, getIncomingErr: errDB}, req: withRouteParam(adminRequest(http.MethodGet, "/api/tu/surat/incoming/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
		{name: "incoming update validation", handler: (*Letter).UpdateIncoming, svc: &fakeLetterService{Letter: &service.Letter{}, updateIncomingErr: errors.New("perihal wajib diisi")}, req: withRouteParam(adminRequest(http.MethodPut, "/api/tu/surat/incoming/"+id.String(), `{}`), "id", id.String()), wantStatus: http.StatusBadRequest},
		{name: "incoming status validation", handler: (*Letter).UpdateIncomingStatus, svc: &fakeLetterService{Letter: &service.Letter{}, incomingStatusErr: errors.New("status tidak valid")}, req: withRouteParam(adminRequest(http.MethodPatch, "/api/tu/surat/incoming/"+id.String()+"/status", `{"status":"x"}`), "id", id.String()), wantStatus: http.StatusBadRequest},
		{name: "incoming delete", handler: (*Letter).DeleteIncoming, svc: &fakeLetterService{Letter: &service.Letter{}, deleteIncomingErr: errDB}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/tu/surat/incoming/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
		{name: "outgoing list", handler: (*Letter).ListOutgoing, svc: &fakeLetterService{Letter: &service.Letter{}, outgoingErr: errDB}, req: adminRequest(http.MethodGet, "/api/tu/surat/outgoing", ""), wantStatus: http.StatusInternalServerError},
		{name: "outgoing create validation", handler: (*Letter).CreateOutgoing, svc: &fakeLetterService{Letter: &service.Letter{}, createOutgoingErr: errors.New("kode klasifikasi wajib diisi")}, req: adminRequest(http.MethodPost, "/api/tu/surat/outgoing", `{}`), wantStatus: http.StatusBadRequest},
		{name: "outgoing get", handler: (*Letter).GetOutgoing, svc: &fakeLetterService{Letter: &service.Letter{}, getOutgoingErr: errDB}, req: withRouteParam(adminRequest(http.MethodGet, "/api/tu/surat/outgoing/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
		{name: "outgoing update validation", handler: (*Letter).UpdateOutgoing, svc: &fakeLetterService{Letter: &service.Letter{}, updateOutgoingErr: errors.New("tujuan surat wajib diisi")}, req: withRouteParam(adminRequest(http.MethodPut, "/api/tu/surat/outgoing/"+id.String(), `{}`), "id", id.String()), wantStatus: http.StatusBadRequest},
		{name: "outgoing delete", handler: (*Letter).DeleteOutgoing, svc: &fakeLetterService{Letter: &service.Letter{}, deleteOutgoingErr: errDB}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/tu/surat/outgoing/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
		{name: "preview validation", handler: (*Letter).PreviewOutgoingNumber, svc: &fakeLetterService{Letter: &service.Letter{}, previewErr: errors.New("tanggal tidak valid")}, req: adminRequest(http.MethodGet, "/api/tu/surat/outgoing/preview-number", ""), wantStatus: http.StatusBadRequest},
		{name: "dispositions list validation", handler: (*Letter).ListDispositions, svc: &fakeLetterService{Letter: &service.Letter{}, dispositionErr: errors.New("incoming_letter_id tidak valid")}, req: adminRequest(http.MethodGet, "/api/tu/surat/disposisi?incoming_letter_id=bad", ""), wantStatus: http.StatusBadRequest},
		{name: "disposition create validation", handler: (*Letter).CreateDisposition, svc: &fakeLetterService{Letter: &service.Letter{}, createDispErr: errors.New("assignee_employee_id tidak valid")}, req: adminRequest(http.MethodPost, "/api/tu/surat/disposisi", `{}`), wantStatus: http.StatusBadRequest},
		{name: "disposition get", handler: (*Letter).GetDisposition, svc: &fakeLetterService{Letter: &service.Letter{}, getDispErr: errDB}, req: withRouteParam(adminRequest(http.MethodGet, "/api/tu/surat/disposisi/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
		{name: "disposition update validation", handler: (*Letter).UpdateDisposition, svc: &fakeLetterService{Letter: &service.Letter{}, updateDispErr: errors.New("status disposisi tidak valid")}, req: withRouteParam(adminRequest(http.MethodPut, "/api/tu/surat/disposisi/"+id.String(), `{}`), "id", id.String()), wantStatus: http.StatusBadRequest},
		{name: "disposition delete", handler: (*Letter).DeleteDisposition, svc: &fakeLetterService{Letter: &service.Letter{}, deleteDispErr: errDB}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/tu/surat/disposisi/"+id.String(), ""), "id", id.String()), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Letter{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
