package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var romanMonths = [...]string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}

type letterStore interface {
	ListLetterClassifications(ctx context.Context) ([]db.LetterClassification, error)
	IssueIncomingLetterSequence(ctx context.Context, year int32) (int32, error)
	IssueOutgoingLetterSequence(ctx context.Context, arg db.IssueOutgoingLetterSequenceParams) (int32, error)
	ListIncomingLetters(ctx context.Context, arg db.ListIncomingLettersParams) ([]db.ListIncomingLettersRow, error)
	GetIncomingLetter(ctx context.Context, id pgtype.UUID) (db.GetIncomingLetterRow, error)
	CreateIncomingLetter(ctx context.Context, arg db.CreateIncomingLetterParams) (db.IncomingLetter, error)
	UpdateIncomingLetter(ctx context.Context, arg db.UpdateIncomingLetterParams) (db.IncomingLetter, error)
	UpdateIncomingLetterStatus(ctx context.Context, arg db.UpdateIncomingLetterStatusParams) (db.IncomingLetter, error)
	DeleteIncomingLetter(ctx context.Context, id pgtype.UUID) error
	ListOutgoingLetters(ctx context.Context, search string) ([]db.ListOutgoingLettersRow, error)
	GetOutgoingLetter(ctx context.Context, id pgtype.UUID) (db.GetOutgoingLetterRow, error)
	CreateOutgoingLetter(ctx context.Context, arg db.CreateOutgoingLetterParams) (db.OutgoingLetter, error)
	UpdateOutgoingLetter(ctx context.Context, arg db.UpdateOutgoingLetterParams) (db.OutgoingLetter, error)
	DeleteOutgoingLetter(ctx context.Context, id pgtype.UUID) error
	ListDispositions(ctx context.Context, arg db.ListDispositionsParams) ([]db.ListDispositionsRow, error)
	GetDisposition(ctx context.Context, id pgtype.UUID) (db.GetDispositionRow, error)
	CreateDisposition(ctx context.Context, arg db.CreateDispositionParams) (db.LetterDisposition, error)
	UpdateDisposition(ctx context.Context, arg db.UpdateDispositionParams) (db.LetterDisposition, error)
	DeleteDisposition(ctx context.Context, id pgtype.UUID) error
	CountDispositionsForLetter(ctx context.Context, incomingLetterID pgtype.UUID) (int64, error)
}

type Letter struct{ q letterStore }

func NewLetter(q *db.Queries) *Letter { return &Letter{q: q} }

func (s *Letter) ListClassifications(ctx context.Context) ([]db.LetterClassification, error) {
	return s.q.ListLetterClassifications(ctx)
}

// =====================
// Incoming Letters
// =====================

func (s *Letter) ListIncoming(ctx context.Context, search, filterStatus string) ([]db.ListIncomingLettersRow, error) {
	return s.q.ListIncomingLetters(ctx, db.ListIncomingLettersParams{
		Search:       strings.TrimSpace(search),
		FilterStatus: strings.TrimSpace(filterStatus),
	})
}

func (s *Letter) GetIncoming(ctx context.Context, id pgtype.UUID) (db.GetIncomingLetterRow, error) {
	return s.q.GetIncomingLetter(ctx, id)
}

type CreateIncomingParams struct {
	NomorSurat           string
	TanggalSurat         string
	TanggalTerima        string
	Asal                 string
	Perihal              string
	Sifat                string
	Catatan              string
	ReceivedByEmployeeID string
}

func (s *Letter) CreateIncoming(ctx context.Context, p CreateIncomingParams) (db.IncomingLetter, error) {
	p.NomorSurat = strings.TrimSpace(p.NomorSurat)
	p.Asal = strings.TrimSpace(p.Asal)
	p.Perihal = strings.TrimSpace(p.Perihal)
	if p.NomorSurat == "" {
		return db.IncomingLetter{}, fmt.Errorf("nomor surat wajib diisi")
	}
	if p.Asal == "" {
		return db.IncomingLetter{}, fmt.Errorf("asal surat wajib diisi")
	}
	if p.Perihal == "" {
		return db.IncomingLetter{}, fmt.Errorf("perihal wajib diisi")
	}
	var tglSurat pgtype.Date
	if err := tglSurat.Scan(p.TanggalSurat); err != nil {
		return db.IncomingLetter{}, fmt.Errorf("format tanggal surat tidak valid")
	}
	var tglTerima pgtype.Date
	if err := tglTerima.Scan(p.TanggalTerima); err != nil {
		return db.IncomingLetter{}, fmt.Errorf("format tanggal terima tidak valid")
	}
	sifat := db.LetterSifat(p.Sifat)
	if sifat == "" {
		sifat = db.LetterSifatBiasa
	}

	year := int32(tglTerima.Time.Year())
	seq, err := s.q.IssueIncomingLetterSequence(ctx, year)
	if err != nil {
		return db.IncomingLetter{}, fmt.Errorf("gagal generate nomor agenda: %w", err)
	}
	nomorAgenda := fmt.Sprintf("AGD/%d/%04d", year, seq)

	var receivedByID pgtype.UUID
	if strings.TrimSpace(p.ReceivedByEmployeeID) != "" {
		if err := receivedByID.Scan(p.ReceivedByEmployeeID); err != nil {
			return db.IncomingLetter{}, fmt.Errorf("received_by_employee_id tidak valid")
		}
	}

	return s.q.CreateIncomingLetter(ctx, db.CreateIncomingLetterParams{
		NomorSurat:           p.NomorSurat,
		NomorAgenda:          nomorAgenda,
		TanggalSurat:         tglSurat,
		TanggalTerima:        tglTerima,
		Asal:                 p.Asal,
		Perihal:              p.Perihal,
		Sifat:                sifat,
		FilePath:             "",
		Catatan:              p.Catatan,
		ReceivedByEmployeeID: receivedByID,
	})
}

type UpdateIncomingParams struct {
	ID            string
	NomorSurat    string
	TanggalSurat  string
	TanggalTerima string
	Asal          string
	Perihal       string
	Sifat         string
	Catatan       string
}

func (s *Letter) UpdateIncoming(ctx context.Context, p UpdateIncomingParams) (db.IncomingLetter, error) {
	id, err := parseLetterUUID(p.ID)
	if err != nil {
		return db.IncomingLetter{}, fmt.Errorf("id tidak valid")
	}
	p.NomorSurat = strings.TrimSpace(p.NomorSurat)
	p.Asal = strings.TrimSpace(p.Asal)
	p.Perihal = strings.TrimSpace(p.Perihal)
	if p.NomorSurat == "" {
		return db.IncomingLetter{}, fmt.Errorf("nomor surat wajib diisi")
	}
	if p.Asal == "" {
		return db.IncomingLetter{}, fmt.Errorf("asal surat wajib diisi")
	}
	if p.Perihal == "" {
		return db.IncomingLetter{}, fmt.Errorf("perihal wajib diisi")
	}
	var tglSurat pgtype.Date
	if err := tglSurat.Scan(p.TanggalSurat); err != nil {
		return db.IncomingLetter{}, fmt.Errorf("format tanggal surat tidak valid")
	}
	var tglTerima pgtype.Date
	if err := tglTerima.Scan(p.TanggalTerima); err != nil {
		return db.IncomingLetter{}, fmt.Errorf("format tanggal terima tidak valid")
	}
	sifat := db.LetterSifat(p.Sifat)
	if sifat == "" {
		sifat = db.LetterSifatBiasa
	}
	return s.q.UpdateIncomingLetter(ctx, db.UpdateIncomingLetterParams{
		ID: id, NomorSurat: p.NomorSurat, TanggalSurat: tglSurat,
		TanggalTerima: tglTerima, Asal: p.Asal, Perihal: p.Perihal,
		Sifat: sifat, Catatan: p.Catatan,
	})
}

func (s *Letter) UpdateIncomingStatus(ctx context.Context, id pgtype.UUID, status string) (db.IncomingLetter, error) {
	st := db.LetterStatus(status)
	switch st {
	case db.LetterStatusBaru, db.LetterStatusDidisposisi, db.LetterStatusSelesai, db.LetterStatusArsip:
	default:
		return db.IncomingLetter{}, fmt.Errorf("status tidak valid")
	}
	return s.q.UpdateIncomingLetterStatus(ctx, db.UpdateIncomingLetterStatusParams{ID: id, Status: st})
}

func (s *Letter) DeleteIncoming(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteIncomingLetter(ctx, id)
}

// =====================
// Outgoing Letters
// =====================

func (s *Letter) ListOutgoing(ctx context.Context, search string) ([]db.ListOutgoingLettersRow, error) {
	return s.q.ListOutgoingLetters(ctx, strings.TrimSpace(search))
}

func (s *Letter) GetOutgoing(ctx context.Context, id pgtype.UUID) (db.GetOutgoingLetterRow, error) {
	return s.q.GetOutgoingLetter(ctx, id)
}

type CreateOutgoingParams struct {
	ClassificationCode   string
	TanggalSurat         string
	Tujuan               string
	Perihal              string
	Sifat                string
	Catatan              string
	IssuedByEmployeeID   string
	ManualNomor          string
}

func (s *Letter) IssueOutgoingLetterNumber(ctx context.Context, classificationCode, tanggalSurat string) (string, error) {
	var tgl pgtype.Date
	if err := tgl.Scan(tanggalSurat); err != nil {
		return "", fmt.Errorf("tanggal tidak valid")
	}
	t := tgl.Time
	year := int32(t.Year())
	month := t.Month()

	seq, err := s.q.IssueOutgoingLetterSequence(ctx, db.IssueOutgoingLetterSequenceParams{
		Year:               year,
		ClassificationCode: classificationCode,
	})
	if err != nil {
		return "", fmt.Errorf("gagal generate nomor surat: %w", err)
	}
	roman := romanMonths[month]
	return fmt.Sprintf("%03d/%s/MTs.20.05/%s/%d", seq, classificationCode, roman, year), nil
}

func (s *Letter) CreateOutgoing(ctx context.Context, p CreateOutgoingParams) (db.OutgoingLetter, error) {
	p.Tujuan = strings.TrimSpace(p.Tujuan)
	p.Perihal = strings.TrimSpace(p.Perihal)
	p.ClassificationCode = strings.TrimSpace(p.ClassificationCode)
	if p.ClassificationCode == "" {
		return db.OutgoingLetter{}, fmt.Errorf("kode klasifikasi wajib diisi")
	}
	if p.Tujuan == "" {
		return db.OutgoingLetter{}, fmt.Errorf("tujuan surat wajib diisi")
	}
	if p.Perihal == "" {
		return db.OutgoingLetter{}, fmt.Errorf("perihal wajib diisi")
	}
	var tglSurat pgtype.Date
	if err := tglSurat.Scan(p.TanggalSurat); err != nil {
		return db.OutgoingLetter{}, fmt.Errorf("format tanggal surat tidak valid")
	}

	var nomor string
	if strings.TrimSpace(p.ManualNomor) != "" {
		nomor = strings.TrimSpace(p.ManualNomor)
	} else {
		t := tglSurat.Time
		year := int32(t.Year())
		month := t.Month()
		seq, err := s.q.IssueOutgoingLetterSequence(ctx, db.IssueOutgoingLetterSequenceParams{
			Year:               year,
			ClassificationCode: p.ClassificationCode,
		})
		if err != nil {
			return db.OutgoingLetter{}, fmt.Errorf("gagal generate nomor surat: %w", err)
		}
		nomor = fmt.Sprintf("%03d/%s/MTs.20.05/%s/%d", seq, p.ClassificationCode, romanMonths[month], year)
	}

	sifat := db.LetterSifat(p.Sifat)
	if sifat == "" {
		sifat = db.LetterSifatBiasa
	}

	var issuedByID pgtype.UUID
	if strings.TrimSpace(p.IssuedByEmployeeID) != "" {
		if err := issuedByID.Scan(p.IssuedByEmployeeID); err != nil {
			return db.OutgoingLetter{}, fmt.Errorf("issued_by_employee_id tidak valid")
		}
	}

	return s.q.CreateOutgoingLetter(ctx, db.CreateOutgoingLetterParams{
		NomorSurat:         nomor,
		ClassificationCode: p.ClassificationCode,
		TanggalSurat:       tglSurat,
		Tujuan:             p.Tujuan,
		Perihal:            p.Perihal,
		Sifat:              sifat,
		FilePath:           "",
		Catatan:            p.Catatan,
		IssuedByEmployeeID: issuedByID,
	})
}

type UpdateOutgoingParams struct {
	ID           string
	TanggalSurat string
	Tujuan       string
	Perihal      string
	Sifat        string
	Catatan      string
}

func (s *Letter) UpdateOutgoing(ctx context.Context, p UpdateOutgoingParams) (db.OutgoingLetter, error) {
	id, err := parseLetterUUID(p.ID)
	if err != nil {
		return db.OutgoingLetter{}, fmt.Errorf("id tidak valid")
	}
	p.Tujuan = strings.TrimSpace(p.Tujuan)
	p.Perihal = strings.TrimSpace(p.Perihal)
	if p.Tujuan == "" {
		return db.OutgoingLetter{}, fmt.Errorf("tujuan surat wajib diisi")
	}
	if p.Perihal == "" {
		return db.OutgoingLetter{}, fmt.Errorf("perihal wajib diisi")
	}
	var tglSurat pgtype.Date
	if err := tglSurat.Scan(p.TanggalSurat); err != nil {
		return db.OutgoingLetter{}, fmt.Errorf("format tanggal surat tidak valid")
	}
	sifat := db.LetterSifat(p.Sifat)
	if sifat == "" {
		sifat = db.LetterSifatBiasa
	}
	return s.q.UpdateOutgoingLetter(ctx, db.UpdateOutgoingLetterParams{
		ID: id, TanggalSurat: tglSurat, Tujuan: p.Tujuan,
		Perihal: p.Perihal, Sifat: sifat, Catatan: p.Catatan,
	})
}

func (s *Letter) DeleteOutgoing(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteOutgoingLetter(ctx, id)
}

// =====================
// Dispositions
// =====================

func (s *Letter) ListDispositions(ctx context.Context, incomingLetterID, filterStatus string) ([]db.ListDispositionsRow, error) {
	var letterID pgtype.UUID
	if strings.TrimSpace(incomingLetterID) != "" {
		if err := letterID.Scan(incomingLetterID); err != nil {
			return nil, fmt.Errorf("incoming_letter_id tidak valid")
		}
	}
	return s.q.ListDispositions(ctx, db.ListDispositionsParams{
		IncomingLetterID: letterID,
		FilterStatus:     strings.TrimSpace(filterStatus),
	})
}

func (s *Letter) GetDisposition(ctx context.Context, id pgtype.UUID) (db.GetDispositionRow, error) {
	return s.q.GetDisposition(ctx, id)
}

type CreateDispositionParams struct {
	IncomingLetterID     string
	AssigneeEmployeeID   string
	Instruksi            string
	DisposedByEmployeeID string
}

func (s *Letter) CreateDisposition(ctx context.Context, p CreateDispositionParams) (db.LetterDisposition, error) {
	letterID, err := parseLetterUUID(p.IncomingLetterID)
	if err != nil {
		return db.LetterDisposition{}, fmt.Errorf("incoming_letter_id tidak valid")
	}
	assigneeID, err := parseLetterUUID(p.AssigneeEmployeeID)
	if err != nil {
		return db.LetterDisposition{}, fmt.Errorf("assignee_employee_id tidak valid")
	}
	var disposedByID pgtype.UUID
	if strings.TrimSpace(p.DisposedByEmployeeID) != "" {
		if err := disposedByID.Scan(p.DisposedByEmployeeID); err != nil {
			return db.LetterDisposition{}, fmt.Errorf("disposed_by_employee_id tidak valid")
		}
	}

	disp, err := s.q.CreateDisposition(ctx, db.CreateDispositionParams{
		IncomingLetterID:     letterID,
		AssigneeEmployeeID:   assigneeID,
		Instruksi:            strings.TrimSpace(p.Instruksi),
		DisposedByEmployeeID: disposedByID,
	})
	if err != nil {
		return db.LetterDisposition{}, err
	}

	// Auto-update incoming letter status to didisposisi
	_, _ = s.q.UpdateIncomingLetterStatus(ctx, db.UpdateIncomingLetterStatusParams{
		ID:     letterID,
		Status: db.LetterStatusDidisposisi,
	})

	return disp, nil
}

type UpdateDispositionParams struct {
	ID                  string
	Instruksi           string
	CatatanTindakLanjut string
	Status              string
}

func (s *Letter) UpdateDisposition(ctx context.Context, p UpdateDispositionParams) (db.LetterDisposition, error) {
	id, err := parseLetterUUID(p.ID)
	if err != nil {
		return db.LetterDisposition{}, fmt.Errorf("id tidak valid")
	}
	st := db.DispositionStatus(p.Status)
	switch st {
	case db.DispositionStatusTerkirim, db.DispositionStatusDibaca,
		db.DispositionStatusDitindaklanjuti, db.DispositionStatusSelesai:
	default:
		return db.LetterDisposition{}, fmt.Errorf("status disposisi tidak valid")
	}
	return s.q.UpdateDisposition(ctx, db.UpdateDispositionParams{
		ID:                  id,
		Instruksi:           strings.TrimSpace(p.Instruksi),
		CatatanTindakLanjut: strings.TrimSpace(p.CatatanTindakLanjut),
		Status:              st,
	})
}

func (s *Letter) DeleteDisposition(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteDisposition(ctx, id)
}

// PreviewOutgoingNumber returns what the next auto-generated outgoing letter
// number would look like for the given classification and date.
// NOTE: This does NOT consume a sequence number.
func (s *Letter) PreviewOutgoingNumber(classificationCode, tanggalSurat string) (string, error) {
	var tgl pgtype.Date
	if err := tgl.Scan(tanggalSurat); err != nil {
		return "", fmt.Errorf("tanggal tidak valid")
	}
	t := tgl.Time
	year := t.Year()
	month := t.Month()
	roman := romanMonths[month]
	return fmt.Sprintf("XXX/%s/MTs.20.05/%s/%d", classificationCode, roman, year), nil
}

func parseLetterUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}
