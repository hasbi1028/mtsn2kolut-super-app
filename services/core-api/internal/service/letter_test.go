package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeLetterStore struct {
	classifications []db.LetterClassification
	classErr        error

	incomingSeqYear int32
	incomingSeq     int32
	incomingSeqErr  error
	outgoingSeqArg  db.IssueOutgoingLetterSequenceParams
	outgoingSeq     int32
	outgoingSeqErr  error

	listIncomingArg   db.ListIncomingLettersParams
	listIncomingRows  []db.ListIncomingLettersRow
	getIncomingID     pgtype.UUID
	getIncomingRow    db.GetIncomingLetterRow
	createIncomingArg db.CreateIncomingLetterParams
	createIncomingErr error
	updateIncomingArg db.UpdateIncomingLetterParams
	updateIncomingErr error
	statusArgs        []db.UpdateIncomingLetterStatusParams
	statusErr         error
	deleteIncomingID  pgtype.UUID
	deleteIncomingErr error

	listOutgoingSearch string
	listOutgoingRows   []db.ListOutgoingLettersRow
	getOutgoingID      pgtype.UUID
	getOutgoingRow     db.GetOutgoingLetterRow
	createOutgoingArg  db.CreateOutgoingLetterParams
	createOutgoingErr  error
	updateOutgoingArg  db.UpdateOutgoingLetterParams
	updateOutgoingErr  error
	deleteOutgoingID   pgtype.UUID
	deleteOutgoingErr  error

	listDispositionsArg  db.ListDispositionsParams
	listDispositionsRows []db.ListDispositionsRow
	getDispositionID     pgtype.UUID
	getDispositionRow    db.GetDispositionRow
	createDispositionArg db.CreateDispositionParams
	createDispositionErr error
	updateDispositionArg db.UpdateDispositionParams
	updateDispositionErr error
	deleteDispositionID  pgtype.UUID
	deleteDispositionErr error
	countLetterID        pgtype.UUID
	countDisposition     int64
	countDispositionErr  error
}

func (f *fakeLetterStore) ListLetterClassifications(ctx context.Context) ([]db.LetterClassification, error) {
	return f.classifications, f.classErr
}

func (f *fakeLetterStore) IssueIncomingLetterSequence(ctx context.Context, year int32) (int32, error) {
	f.incomingSeqYear = year
	return f.incomingSeq, f.incomingSeqErr
}

func (f *fakeLetterStore) IssueOutgoingLetterSequence(ctx context.Context, arg db.IssueOutgoingLetterSequenceParams) (int32, error) {
	f.outgoingSeqArg = arg
	return f.outgoingSeq, f.outgoingSeqErr
}

func (f *fakeLetterStore) ListIncomingLetters(ctx context.Context, arg db.ListIncomingLettersParams) ([]db.ListIncomingLettersRow, error) {
	f.listIncomingArg = arg
	return f.listIncomingRows, nil
}

func (f *fakeLetterStore) GetIncomingLetter(ctx context.Context, id pgtype.UUID) (db.GetIncomingLetterRow, error) {
	f.getIncomingID = id
	return f.getIncomingRow, nil
}

func (f *fakeLetterStore) CreateIncomingLetter(ctx context.Context, arg db.CreateIncomingLetterParams) (db.IncomingLetter, error) {
	f.createIncomingArg = arg
	if f.createIncomingErr != nil {
		return db.IncomingLetter{}, f.createIncomingErr
	}
	return db.IncomingLetter{
		ID:                   letterTestUUID(1),
		NomorSurat:           arg.NomorSurat,
		NomorAgenda:          arg.NomorAgenda,
		TanggalSurat:         arg.TanggalSurat,
		TanggalTerima:        arg.TanggalTerima,
		Asal:                 arg.Asal,
		Perihal:              arg.Perihal,
		Sifat:                arg.Sifat,
		Catatan:              arg.Catatan,
		ReceivedByEmployeeID: arg.ReceivedByEmployeeID,
	}, nil
}

func (f *fakeLetterStore) UpdateIncomingLetter(ctx context.Context, arg db.UpdateIncomingLetterParams) (db.IncomingLetter, error) {
	f.updateIncomingArg = arg
	if f.updateIncomingErr != nil {
		return db.IncomingLetter{}, f.updateIncomingErr
	}
	return db.IncomingLetter{
		ID:            arg.ID,
		NomorSurat:    arg.NomorSurat,
		TanggalSurat:  arg.TanggalSurat,
		TanggalTerima: arg.TanggalTerima,
		Asal:          arg.Asal,
		Perihal:       arg.Perihal,
		Sifat:         arg.Sifat,
		Catatan:       arg.Catatan,
	}, nil
}

func (f *fakeLetterStore) UpdateIncomingLetterStatus(ctx context.Context, arg db.UpdateIncomingLetterStatusParams) (db.IncomingLetter, error) {
	f.statusArgs = append(f.statusArgs, arg)
	if f.statusErr != nil {
		return db.IncomingLetter{}, f.statusErr
	}
	return db.IncomingLetter{ID: arg.ID, Status: arg.Status}, nil
}

func (f *fakeLetterStore) DeleteIncomingLetter(ctx context.Context, id pgtype.UUID) error {
	f.deleteIncomingID = id
	return f.deleteIncomingErr
}

func (f *fakeLetterStore) ListOutgoingLetters(ctx context.Context, search string) ([]db.ListOutgoingLettersRow, error) {
	f.listOutgoingSearch = search
	return f.listOutgoingRows, nil
}

func (f *fakeLetterStore) GetOutgoingLetter(ctx context.Context, id pgtype.UUID) (db.GetOutgoingLetterRow, error) {
	f.getOutgoingID = id
	return f.getOutgoingRow, nil
}

func (f *fakeLetterStore) CreateOutgoingLetter(ctx context.Context, arg db.CreateOutgoingLetterParams) (db.OutgoingLetter, error) {
	f.createOutgoingArg = arg
	if f.createOutgoingErr != nil {
		return db.OutgoingLetter{}, f.createOutgoingErr
	}
	return db.OutgoingLetter{
		ID:                 letterTestUUID(2),
		NomorSurat:         arg.NomorSurat,
		ClassificationCode: arg.ClassificationCode,
		TanggalSurat:       arg.TanggalSurat,
		Tujuan:             arg.Tujuan,
		Perihal:            arg.Perihal,
		Sifat:              arg.Sifat,
		Catatan:            arg.Catatan,
		IssuedByEmployeeID: arg.IssuedByEmployeeID,
	}, nil
}

func (f *fakeLetterStore) UpdateOutgoingLetter(ctx context.Context, arg db.UpdateOutgoingLetterParams) (db.OutgoingLetter, error) {
	f.updateOutgoingArg = arg
	if f.updateOutgoingErr != nil {
		return db.OutgoingLetter{}, f.updateOutgoingErr
	}
	return db.OutgoingLetter{ID: arg.ID, TanggalSurat: arg.TanggalSurat, Tujuan: arg.Tujuan, Perihal: arg.Perihal, Sifat: arg.Sifat, Catatan: arg.Catatan}, nil
}

func (f *fakeLetterStore) DeleteOutgoingLetter(ctx context.Context, id pgtype.UUID) error {
	f.deleteOutgoingID = id
	return f.deleteOutgoingErr
}

func (f *fakeLetterStore) ListDispositions(ctx context.Context, arg db.ListDispositionsParams) ([]db.ListDispositionsRow, error) {
	f.listDispositionsArg = arg
	return f.listDispositionsRows, nil
}

func (f *fakeLetterStore) GetDisposition(ctx context.Context, id pgtype.UUID) (db.GetDispositionRow, error) {
	f.getDispositionID = id
	return f.getDispositionRow, nil
}

func (f *fakeLetterStore) CreateDisposition(ctx context.Context, arg db.CreateDispositionParams) (db.LetterDisposition, error) {
	f.createDispositionArg = arg
	if f.createDispositionErr != nil {
		return db.LetterDisposition{}, f.createDispositionErr
	}
	return db.LetterDisposition{
		ID:                   letterTestUUID(3),
		IncomingLetterID:     arg.IncomingLetterID,
		AssigneeEmployeeID:   arg.AssigneeEmployeeID,
		Instruksi:            arg.Instruksi,
		DisposedByEmployeeID: arg.DisposedByEmployeeID,
		Status:               db.DispositionStatusTerkirim,
	}, nil
}

func (f *fakeLetterStore) UpdateDisposition(ctx context.Context, arg db.UpdateDispositionParams) (db.LetterDisposition, error) {
	f.updateDispositionArg = arg
	if f.updateDispositionErr != nil {
		return db.LetterDisposition{}, f.updateDispositionErr
	}
	return db.LetterDisposition{ID: arg.ID, Instruksi: arg.Instruksi, CatatanTindakLanjut: arg.CatatanTindakLanjut, Status: arg.Status}, nil
}

func (f *fakeLetterStore) DeleteDisposition(ctx context.Context, id pgtype.UUID) error {
	f.deleteDispositionID = id
	return f.deleteDispositionErr
}

func (f *fakeLetterStore) CountDispositionsForLetter(ctx context.Context, incomingLetterID pgtype.UUID) (int64, error) {
	f.countLetterID = incomingLetterID
	return f.countDisposition, f.countDispositionErr
}

func letterTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func mustLetterUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("Scan(%q) error = %v", value, err)
	}
	return id
}

func TestLetterIncomingListCreateUpdateAndStatus(t *testing.T) {
	receivedByID := "00000000-0000-0000-0000-000000000009"
	parsedReceiver := mustLetterUUID(t, receivedByID)
	store := &fakeLetterStore{incomingSeq: 12}
	svc := &Letter{q: store}

	if _, err := svc.ListIncoming(context.Background(), "  undangan  ", "  baru  "); err != nil {
		t.Fatalf("ListIncoming() error = %v", err)
	}
	if store.listIncomingArg.Search != "undangan" || store.listIncomingArg.FilterStatus != "baru" {
		t.Fatalf("ListIncoming() arg = %+v, want trimmed filters", store.listIncomingArg)
	}

	incoming, err := svc.CreateIncoming(context.Background(), CreateIncomingParams{
		NomorSurat:           "  001/KM/V/2026  ",
		TanggalSurat:         "2026-05-01",
		TanggalTerima:        "2026-05-02",
		Asal:                 "  Kemenag  ",
		Perihal:              "  Undangan Rapat  ",
		Catatan:              "catatan",
		ReceivedByEmployeeID: receivedByID,
	})
	if err != nil {
		t.Fatalf("CreateIncoming() error = %v", err)
	}
	if incoming.NomorAgenda != "AGD/2026/0012" || store.incomingSeqYear != 2026 {
		t.Fatalf("CreateIncoming() agenda/year = %q/%d, want generated agenda for 2026", incoming.NomorAgenda, store.incomingSeqYear)
	}
	arg := store.createIncomingArg
	if arg.NomorSurat != "001/KM/V/2026" || arg.Asal != "Kemenag" || arg.Perihal != "Undangan Rapat" {
		t.Fatalf("CreateIncomingLetter() arg = %+v, want trimmed fields", arg)
	}
	if arg.Sifat != db.LetterSifatBiasa || arg.ReceivedByEmployeeID != parsedReceiver {
		t.Fatalf("CreateIncomingLetter() arg = %+v, want default sifat and parsed receiver", arg)
	}

	_, err = svc.UpdateIncoming(context.Background(), UpdateIncomingParams{
		ID:            "00000000-0000-0000-0000-000000000001",
		NomorSurat:    "  002/KM/V/2026  ",
		TanggalSurat:  "2026-05-03",
		TanggalTerima: "2026-05-04",
		Asal:          "  Kanwil  ",
		Perihal:       "  Balasan  ",
		Sifat:         "penting",
		Catatan:       "cek",
	})
	if err != nil {
		t.Fatalf("UpdateIncoming() error = %v", err)
	}
	if store.updateIncomingArg.NomorSurat != "002/KM/V/2026" || store.updateIncomingArg.Sifat != db.LetterSifatPenting {
		t.Fatalf("UpdateIncomingLetter() arg = %+v, want trimmed number and explicit sifat", store.updateIncomingArg)
	}

	if _, err := svc.UpdateIncomingStatus(context.Background(), letterTestUUID(4), "selesai"); err != nil {
		t.Fatalf("UpdateIncomingStatus() error = %v", err)
	}
	if len(store.statusArgs) != 1 || store.statusArgs[0].Status != db.LetterStatusSelesai {
		t.Fatalf("UpdateIncomingLetterStatus() args = %+v, want selesai", store.statusArgs)
	}
	if _, err := svc.UpdateIncomingStatus(context.Background(), letterTestUUID(4), "invalid"); err == nil || err.Error() != "status tidak valid" {
		t.Fatalf("UpdateIncomingStatus(invalid) error = %v, want invalid status", err)
	}
}

func TestLetterIncomingValidationStopsBeforeSequence(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateIncomingParams
		wantErr string
	}{
		{name: "empty nomor", params: CreateIncomingParams{Asal: "A", Perihal: "P", TanggalSurat: "2026-05-01", TanggalTerima: "2026-05-02"}, wantErr: "nomor surat wajib diisi"},
		{name: "empty asal", params: CreateIncomingParams{NomorSurat: "N", Perihal: "P", TanggalSurat: "2026-05-01", TanggalTerima: "2026-05-02"}, wantErr: "asal surat wajib diisi"},
		{name: "empty perihal", params: CreateIncomingParams{NomorSurat: "N", Asal: "A", TanggalSurat: "2026-05-01", TanggalTerima: "2026-05-02"}, wantErr: "perihal wajib diisi"},
		{name: "bad tanggal surat", params: CreateIncomingParams{NomorSurat: "N", Asal: "A", Perihal: "P", TanggalSurat: "bad", TanggalTerima: "2026-05-02"}, wantErr: "format tanggal surat tidak valid"},
		{name: "bad tanggal terima", params: CreateIncomingParams{NomorSurat: "N", Asal: "A", Perihal: "P", TanggalSurat: "2026-05-01", TanggalTerima: "bad"}, wantErr: "format tanggal terima tidak valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeLetterStore{incomingSeq: 1}
			svc := &Letter{q: store}

			_, err := svc.CreateIncoming(context.Background(), tt.params)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("CreateIncoming() error = %v, want %q", err, tt.wantErr)
			}
			if store.incomingSeqYear != 0 {
				t.Fatalf("IssueIncomingLetterSequence() year = %d, want not called", store.incomingSeqYear)
			}
		})
	}
}

func TestLetterIncomingPropagatesStoreAndReceiverErrors(t *testing.T) {
	expectedErr := errors.New("store failed")

	store := &fakeLetterStore{incomingSeqErr: expectedErr}
	svc := &Letter{q: store}
	_, err := svc.CreateIncoming(context.Background(), CreateIncomingParams{
		NomorSurat:    "001/KM/V/2026",
		TanggalSurat:  "2026-05-01",
		TanggalTerima: "2026-05-02",
		Asal:          "Kemenag",
		Perihal:       "Undangan",
	})
	if err == nil || !strings.Contains(err.Error(), "gagal generate nomor agenda") {
		t.Fatalf("CreateIncoming(sequence error) error = %v, want wrapped agenda error", err)
	}

	store = &fakeLetterStore{incomingSeq: 1}
	svc = &Letter{q: store}
	_, err = svc.CreateIncoming(context.Background(), CreateIncomingParams{
		NomorSurat:           "001/KM/V/2026",
		TanggalSurat:         "2026-05-01",
		TanggalTerima:        "2026-05-02",
		Asal:                 "Kemenag",
		Perihal:              "Undangan",
		ReceivedByEmployeeID: "bad",
	})
	if err == nil || err.Error() != "received_by_employee_id tidak valid" {
		t.Fatalf("CreateIncoming(bad receiver) error = %v, want invalid receiver", err)
	}

	store = &fakeLetterStore{incomingSeq: 1, createIncomingErr: expectedErr}
	svc = &Letter{q: store}
	_, err = svc.CreateIncoming(context.Background(), CreateIncomingParams{
		NomorSurat:    "001/KM/V/2026",
		TanggalSurat:  "2026-05-01",
		TanggalTerima: "2026-05-02",
		Asal:          "Kemenag",
		Perihal:       "Undangan",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("CreateIncoming(create error) error = %v, want %v", err, expectedErr)
	}
}

func TestLetterUpdateIncomingValidationAndStoreError(t *testing.T) {
	validID := "00000000-0000-0000-0000-000000000021"
	tests := []struct {
		name    string
		params  UpdateIncomingParams
		wantErr string
	}{
		{name: "bad id", params: UpdateIncomingParams{ID: "bad"}, wantErr: "id tidak valid"},
		{name: "empty nomor", params: UpdateIncomingParams{ID: validID, Asal: "A", Perihal: "P", TanggalSurat: "2026-05-01", TanggalTerima: "2026-05-02"}, wantErr: "nomor surat wajib diisi"},
		{name: "empty asal", params: UpdateIncomingParams{ID: validID, NomorSurat: "N", Perihal: "P", TanggalSurat: "2026-05-01", TanggalTerima: "2026-05-02"}, wantErr: "asal surat wajib diisi"},
		{name: "empty perihal", params: UpdateIncomingParams{ID: validID, NomorSurat: "N", Asal: "A", TanggalSurat: "2026-05-01", TanggalTerima: "2026-05-02"}, wantErr: "perihal wajib diisi"},
		{name: "bad tanggal surat", params: UpdateIncomingParams{ID: validID, NomorSurat: "N", Asal: "A", Perihal: "P", TanggalSurat: "bad", TanggalTerima: "2026-05-02"}, wantErr: "format tanggal surat tidak valid"},
		{name: "bad tanggal terima", params: UpdateIncomingParams{ID: validID, NomorSurat: "N", Asal: "A", Perihal: "P", TanggalSurat: "2026-05-01", TanggalTerima: "bad"}, wantErr: "format tanggal terima tidak valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Letter{q: &fakeLetterStore{}}
			_, err := svc.UpdateIncoming(context.Background(), tt.params)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("UpdateIncoming() error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	expectedErr := errors.New("update failed")
	svc := &Letter{q: &fakeLetterStore{updateIncomingErr: expectedErr}}
	_, err := svc.UpdateIncoming(context.Background(), UpdateIncomingParams{
		ID:            validID,
		NomorSurat:    "002/KM/V/2026",
		TanggalSurat:  "2026-05-03",
		TanggalTerima: "2026-05-04",
		Asal:          "Kanwil",
		Perihal:       "Balasan",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateIncoming(store error) error = %v, want %v", err, expectedErr)
	}
}

func TestLetterOutgoingNumbersAndMutations(t *testing.T) {
	issuedByID := "00000000-0000-0000-0000-000000000008"
	parsedIssuer := mustLetterUUID(t, issuedByID)
	store := &fakeLetterStore{outgoingSeq: 7}
	svc := &Letter{q: store}

	number, err := svc.IssueOutgoingLetterNumber(context.Background(), " 421 ", "2026-05-01")
	if err != nil {
		t.Fatalf("IssueOutgoingLetterNumber() error = %v", err)
	}
	if number != "007/421/MTs.20.05/V/2026" {
		t.Fatalf("IssueOutgoingLetterNumber() = %q, want generated number", number)
	}
	if store.outgoingSeqArg.Year != 2026 || store.outgoingSeqArg.ClassificationCode != "421" {
		t.Fatalf("IssueOutgoingLetterSequence() arg = %+v, want year/classification", store.outgoingSeqArg)
	}

	if _, err := svc.ListOutgoing(context.Background(), "  rapat  "); err != nil {
		t.Fatalf("ListOutgoing() error = %v", err)
	}
	if store.listOutgoingSearch != "rapat" {
		t.Fatalf("ListOutgoing() search = %q, want trimmed", store.listOutgoingSearch)
	}

	outgoing, err := svc.CreateOutgoing(context.Background(), CreateOutgoingParams{
		ClassificationCode: " 421 ",
		TanggalSurat:       "2026-05-01",
		Tujuan:             "  Komite  ",
		Perihal:            "  Undangan  ",
		IssuedByEmployeeID: issuedByID,
	})
	if err != nil {
		t.Fatalf("CreateOutgoing() error = %v", err)
	}
	if outgoing.NomorSurat != "007/421/MTs.20.05/V/2026" {
		t.Fatalf("CreateOutgoing() nomor = %q, want generated number", outgoing.NomorSurat)
	}
	arg := store.createOutgoingArg
	if arg.ClassificationCode != "421" || arg.Tujuan != "Komite" || arg.Perihal != "Undangan" {
		t.Fatalf("CreateOutgoingLetter() arg = %+v, want trimmed fields", arg)
	}
	if arg.Sifat != db.LetterSifatBiasa || arg.IssuedByEmployeeID != parsedIssuer {
		t.Fatalf("CreateOutgoingLetter() arg = %+v, want default sifat and parsed issuer", arg)
	}

	manualStore := &fakeLetterStore{outgoingSeq: 99}
	manualSvc := &Letter{q: manualStore}
	if _, err := manualSvc.CreateOutgoing(context.Background(), CreateOutgoingParams{
		ClassificationCode: "800",
		TanggalSurat:       "2026-05-01",
		Tujuan:             "TU",
		Perihal:            "Surat Balasan",
		ManualNomor:        "  MANUAL/1  ",
	}); err != nil {
		t.Fatalf("CreateOutgoing(manual) error = %v", err)
	}
	if manualStore.createOutgoingArg.NomorSurat != "MANUAL/1" || manualStore.outgoingSeqArg.Year != 0 {
		t.Fatalf("CreateOutgoing(manual) arg = %+v, seq = %+v; want manual number without sequence", manualStore.createOutgoingArg, manualStore.outgoingSeqArg)
	}

	_, err = svc.UpdateOutgoing(context.Background(), UpdateOutgoingParams{
		ID:           "00000000-0000-0000-0000-000000000002",
		TanggalSurat: "2026-05-02",
		Tujuan:       "  Kemenag  ",
		Perihal:      "  Laporan  ",
		Sifat:        "rahasia",
		Catatan:      "arsip",
	})
	if err != nil {
		t.Fatalf("UpdateOutgoing() error = %v", err)
	}
	if store.updateOutgoingArg.Tujuan != "Kemenag" || store.updateOutgoingArg.Perihal != "Laporan" || store.updateOutgoingArg.Sifat != db.LetterSifatRahasia {
		t.Fatalf("UpdateOutgoingLetter() arg = %+v, want normalized outgoing update", store.updateOutgoingArg)
	}
}

func TestLetterOutgoingValidationAndPreview(t *testing.T) {
	store := &fakeLetterStore{outgoingSeq: 1}
	svc := &Letter{q: store}

	if _, err := svc.CreateOutgoing(context.Background(), CreateOutgoingParams{TanggalSurat: "2026-05-01", Tujuan: "A", Perihal: "B"}); err == nil || err.Error() != "kode klasifikasi wajib diisi" {
		t.Fatalf("CreateOutgoing(no classification) error = %v, want classification required", err)
	}
	if _, err := svc.CreateOutgoing(context.Background(), CreateOutgoingParams{ClassificationCode: "421", TanggalSurat: "bad", Tujuan: "A", Perihal: "B"}); err == nil || err.Error() != "format tanggal surat tidak valid" {
		t.Fatalf("CreateOutgoing(bad date) error = %v, want date validation", err)
	}
	if _, err := svc.UpdateOutgoing(context.Background(), UpdateOutgoingParams{ID: "bad"}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("UpdateOutgoing(bad id) error = %v, want id validation", err)
	}
	if _, err := svc.UpdateOutgoing(context.Background(), UpdateOutgoingParams{ID: "00000000-0000-0000-0000-000000000022", TanggalSurat: "2026-05-01", Perihal: "B"}); err == nil || err.Error() != "tujuan surat wajib diisi" {
		t.Fatalf("UpdateOutgoing(no tujuan) error = %v, want tujuan required", err)
	}
	if _, err := svc.UpdateOutgoing(context.Background(), UpdateOutgoingParams{ID: "00000000-0000-0000-0000-000000000022", TanggalSurat: "2026-05-01", Tujuan: "A"}); err == nil || err.Error() != "perihal wajib diisi" {
		t.Fatalf("UpdateOutgoing(no perihal) error = %v, want perihal required", err)
	}
	if _, err := svc.UpdateOutgoing(context.Background(), UpdateOutgoingParams{ID: "00000000-0000-0000-0000-000000000022", TanggalSurat: "bad", Tujuan: "A", Perihal: "B"}); err == nil || err.Error() != "format tanggal surat tidak valid" {
		t.Fatalf("UpdateOutgoing(bad date) error = %v, want date validation", err)
	}

	preview, err := svc.PreviewOutgoingNumber("421", "2026-05-01")
	if err != nil {
		t.Fatalf("PreviewOutgoingNumber() error = %v", err)
	}
	if preview != "XXX/421/MTs.20.05/V/2026" {
		t.Fatalf("PreviewOutgoingNumber() = %q, want preview number", preview)
	}
	if _, err := svc.PreviewOutgoingNumber("421", "bad"); err == nil || err.Error() != "tanggal tidak valid" {
		t.Fatalf("PreviewOutgoingNumber(bad date) error = %v, want invalid date", err)
	}
	if _, err := svc.IssueOutgoingLetterNumber(context.Background(), "421", "bad"); err == nil || err.Error() != "tanggal tidak valid" {
		t.Fatalf("IssueOutgoingLetterNumber(bad date) error = %v, want invalid date", err)
	}
	if _, err := svc.IssueOutgoingLetterNumber(context.Background(), " ", "2026-05-01"); err == nil || err.Error() != "kode klasifikasi wajib diisi" {
		t.Fatalf("IssueOutgoingLetterNumber(blank code) error = %v, want classification required", err)
	}

	seqErr := errors.New("sequence unavailable")
	store.outgoingSeqErr = seqErr
	_, err = svc.IssueOutgoingLetterNumber(context.Background(), "421", "2026-05-01")
	if err == nil || !strings.Contains(err.Error(), "gagal generate nomor surat") {
		t.Fatalf("IssueOutgoingLetterNumber(seq error) error = %v, want wrapped sequence error", err)
	}
}

func TestLetterOutgoingCreateAndUpdateErrorBranches(t *testing.T) {
	expectedErr := errors.New("store failed")

	svc := &Letter{q: &fakeLetterStore{outgoingSeq: 1}}
	if _, err := svc.CreateOutgoing(context.Background(), CreateOutgoingParams{ClassificationCode: "421", TanggalSurat: "2026-05-01", Perihal: "B"}); err == nil || err.Error() != "tujuan surat wajib diisi" {
		t.Fatalf("CreateOutgoing(no tujuan) error = %v, want tujuan required", err)
	}
	if _, err := svc.CreateOutgoing(context.Background(), CreateOutgoingParams{ClassificationCode: "421", TanggalSurat: "2026-05-01", Tujuan: "A"}); err == nil || err.Error() != "perihal wajib diisi" {
		t.Fatalf("CreateOutgoing(no perihal) error = %v, want perihal required", err)
	}

	svc = &Letter{q: &fakeLetterStore{outgoingSeqErr: expectedErr}}
	_, err := svc.CreateOutgoing(context.Background(), CreateOutgoingParams{
		ClassificationCode: "421",
		TanggalSurat:       "2026-05-01",
		Tujuan:             "Komite",
		Perihal:            "Undangan",
	})
	if err == nil || !strings.Contains(err.Error(), "gagal generate nomor surat") {
		t.Fatalf("CreateOutgoing(sequence error) error = %v, want wrapped sequence error", err)
	}

	svc = &Letter{q: &fakeLetterStore{outgoingSeq: 1}}
	_, err = svc.CreateOutgoing(context.Background(), CreateOutgoingParams{
		ClassificationCode: "421",
		TanggalSurat:       "2026-05-01",
		Tujuan:             "Komite",
		Perihal:            "Undangan",
		IssuedByEmployeeID: "bad",
	})
	if err == nil || err.Error() != "issued_by_employee_id tidak valid" {
		t.Fatalf("CreateOutgoing(bad issuer) error = %v, want invalid issuer", err)
	}

	svc = &Letter{q: &fakeLetterStore{outgoingSeq: 1, createOutgoingErr: expectedErr}}
	_, err = svc.CreateOutgoing(context.Background(), CreateOutgoingParams{
		ClassificationCode: "421",
		TanggalSurat:       "2026-05-01",
		Tujuan:             "Komite",
		Perihal:            "Undangan",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("CreateOutgoing(create error) error = %v, want %v", err, expectedErr)
	}

	svc = &Letter{q: &fakeLetterStore{updateOutgoingErr: expectedErr}}
	_, err = svc.UpdateOutgoing(context.Background(), UpdateOutgoingParams{
		ID:           "00000000-0000-0000-0000-000000000023",
		TanggalSurat: "2026-05-02",
		Tujuan:       "Kemenag",
		Perihal:      "Laporan",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateOutgoing(store error) error = %v, want %v", err, expectedErr)
	}
}

func TestLetterDispositionsCreateUpdateAndList(t *testing.T) {
	incomingID := "00000000-0000-0000-0000-000000000011"
	assigneeID := "00000000-0000-0000-0000-000000000012"
	disposedByID := "00000000-0000-0000-0000-000000000013"
	parsedIncoming := mustLetterUUID(t, incomingID)
	parsedAssignee := mustLetterUUID(t, assigneeID)
	parsedDisposedBy := mustLetterUUID(t, disposedByID)
	store := &fakeLetterStore{}
	svc := &Letter{q: store}

	if _, err := svc.ListDispositions(context.Background(), incomingID, "  terkirim  "); err != nil {
		t.Fatalf("ListDispositions() error = %v", err)
	}
	if store.listDispositionsArg.IncomingLetterID != parsedIncoming || store.listDispositionsArg.FilterStatus != "terkirim" {
		t.Fatalf("ListDispositions() arg = %+v, want parsed id and trimmed status", store.listDispositionsArg)
	}
	if _, err := svc.ListDispositions(context.Background(), "bad", ""); err == nil || err.Error() != "incoming_letter_id tidak valid" {
		t.Fatalf("ListDispositions(bad id) error = %v, want invalid id", err)
	}

	disp, err := svc.CreateDisposition(context.Background(), CreateDispositionParams{
		IncomingLetterID:     incomingID,
		AssigneeEmployeeID:   assigneeID,
		Instruksi:            "  Mohon tindak lanjut  ",
		DisposedByEmployeeID: disposedByID,
	})
	if err != nil {
		t.Fatalf("CreateDisposition() error = %v", err)
	}
	if disp.IncomingLetterID != parsedIncoming || disp.AssigneeEmployeeID != parsedAssignee || disp.DisposedByEmployeeID != parsedDisposedBy {
		t.Fatalf("CreateDisposition() = %+v, want parsed ids", disp)
	}
	if store.createDispositionArg.Instruksi != "Mohon tindak lanjut" {
		t.Fatalf("CreateDisposition() instruksi = %q, want trimmed", store.createDispositionArg.Instruksi)
	}
	if len(store.statusArgs) != 1 || store.statusArgs[0].ID != parsedIncoming || store.statusArgs[0].Status != db.LetterStatusDidisposisi {
		t.Fatalf("CreateDisposition() status args = %+v, want incoming status didisposisi", store.statusArgs)
	}

	_, err = svc.UpdateDisposition(context.Background(), UpdateDispositionParams{
		ID:                  "00000000-0000-0000-0000-000000000014",
		Instruksi:           "  Cek dokumen  ",
		CatatanTindakLanjut: "  Sudah dibalas  ",
		Status:              "selesai",
	})
	if err != nil {
		t.Fatalf("UpdateDisposition() error = %v", err)
	}
	if store.updateDispositionArg.Instruksi != "Cek dokumen" || store.updateDispositionArg.CatatanTindakLanjut != "Sudah dibalas" || store.updateDispositionArg.Status != db.DispositionStatusSelesai {
		t.Fatalf("UpdateDisposition() arg = %+v, want normalized status/update", store.updateDispositionArg)
	}
	if _, err := svc.UpdateDisposition(context.Background(), UpdateDispositionParams{ID: "00000000-0000-0000-0000-000000000014", Status: "invalid"}); err == nil || err.Error() != "status disposisi tidak valid" {
		t.Fatalf("UpdateDisposition(invalid status) error = %v, want invalid status", err)
	}
	if _, err := svc.CreateDisposition(context.Background(), CreateDispositionParams{IncomingLetterID: "bad"}); err == nil || err.Error() != "incoming_letter_id tidak valid" {
		t.Fatalf("CreateDisposition(bad incoming id) error = %v, want invalid incoming id", err)
	}
}

func TestLetterDispositionErrorBranches(t *testing.T) {
	incomingID := "00000000-0000-0000-0000-000000000031"
	assigneeID := "00000000-0000-0000-0000-000000000032"
	expectedErr := errors.New("store failed")
	svc := &Letter{q: &fakeLetterStore{}}

	if _, err := svc.CreateDisposition(context.Background(), CreateDispositionParams{IncomingLetterID: incomingID, AssigneeEmployeeID: "bad"}); err == nil || err.Error() != "assignee_employee_id tidak valid" {
		t.Fatalf("CreateDisposition(bad assignee) error = %v, want invalid assignee", err)
	}
	if _, err := svc.CreateDisposition(context.Background(), CreateDispositionParams{IncomingLetterID: incomingID, AssigneeEmployeeID: assigneeID, DisposedByEmployeeID: "bad"}); err == nil || err.Error() != "disposed_by_employee_id tidak valid" {
		t.Fatalf("CreateDisposition(bad disposed by) error = %v, want invalid disposer", err)
	}

	svc = &Letter{q: &fakeLetterStore{createDispositionErr: expectedErr}}
	_, err := svc.CreateDisposition(context.Background(), CreateDispositionParams{
		IncomingLetterID:   incomingID,
		AssigneeEmployeeID: assigneeID,
		Instruksi:          "Cek dokumen",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("CreateDisposition(create error) error = %v, want %v", err, expectedErr)
	}

	if _, err := svc.UpdateDisposition(context.Background(), UpdateDispositionParams{ID: "bad"}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("UpdateDisposition(bad id) error = %v, want invalid id", err)
	}

	svc = &Letter{q: &fakeLetterStore{updateDispositionErr: expectedErr}}
	_, err = svc.UpdateDisposition(context.Background(), UpdateDispositionParams{
		ID:     "00000000-0000-0000-0000-000000000033",
		Status: "dibaca",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateDisposition(store error) error = %v, want %v", err, expectedErr)
	}
}

func TestLetterReadDeleteDelegations(t *testing.T) {
	id := letterTestUUID(20)
	store := &fakeLetterStore{
		classifications:      []db.LetterClassification{{Code: "421", Name: "Kurikulum", IsActive: true}},
		getIncomingRow:       db.GetIncomingLetterRow{ID: id, NomorSurat: "001"},
		getOutgoingRow:       db.GetOutgoingLetterRow{ID: id, NomorSurat: "002"},
		getDispositionRow:    db.GetDispositionRow{ID: id, Instruksi: "cek"},
		listIncomingRows:     []db.ListIncomingLettersRow{{ID: id}},
		listOutgoingRows:     []db.ListOutgoingLettersRow{{ID: id}},
		listDispositionsRows: []db.ListDispositionsRow{{ID: id}},
	}
	svc := &Letter{q: store}

	classifications, err := svc.ListClassifications(context.Background())
	if err != nil {
		t.Fatalf("ListClassifications() error = %v", err)
	}
	if len(classifications) != 1 || classifications[0].Code != "421" {
		t.Fatalf("ListClassifications() = %+v, want store rows", classifications)
	}

	if _, err := svc.GetIncoming(context.Background(), id); err != nil {
		t.Fatalf("GetIncoming() error = %v", err)
	}
	if _, err := svc.GetOutgoing(context.Background(), id); err != nil {
		t.Fatalf("GetOutgoing() error = %v", err)
	}
	if _, err := svc.GetDisposition(context.Background(), id); err != nil {
		t.Fatalf("GetDisposition() error = %v", err)
	}
	if store.getIncomingID != id || store.getOutgoingID != id || store.getDispositionID != id {
		t.Fatalf("read ids = %v/%v/%v, want %v", store.getIncomingID, store.getOutgoingID, store.getDispositionID, id)
	}

	if err := svc.DeleteIncoming(context.Background(), id); err != nil {
		t.Fatalf("DeleteIncoming() error = %v", err)
	}
	if err := svc.DeleteOutgoing(context.Background(), id); err != nil {
		t.Fatalf("DeleteOutgoing() error = %v", err)
	}
	if err := svc.DeleteDisposition(context.Background(), id); err != nil {
		t.Fatalf("DeleteDisposition() error = %v", err)
	}
	if store.deleteIncomingID != id || store.deleteOutgoingID != id || store.deleteDispositionID != id {
		t.Fatalf("delete ids = %v/%v/%v, want %v", store.deleteIncomingID, store.deleteOutgoingID, store.deleteDispositionID, id)
	}
}
