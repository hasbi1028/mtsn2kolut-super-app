package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type documentCycleStore interface {
	GetDocumentCycleStats(ctx context.Context, periodYear int32) (db.GetDocumentCycleStatsRow, error)
	ListDocumentCycleCatalogs(ctx context.Context, arg db.ListDocumentCycleCatalogsParams) ([]db.ListDocumentCycleCatalogsRow, error)
	CreateDocumentCycleCatalog(ctx context.Context, arg db.CreateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error)
	UpdateDocumentCycleCatalog(ctx context.Context, arg db.UpdateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error)
	DeleteDocumentCycleCatalog(ctx context.Context, id pgtype.UUID) error
	ListDocumentCycleObligations(ctx context.Context, arg db.ListDocumentCycleObligationsParams) ([]db.ListDocumentCycleObligationsRow, error)
	GenerateDocumentCycleYearObligations(ctx context.Context, arg db.GenerateDocumentCycleYearObligationsParams) (db.GenerateDocumentCycleYearObligationsRow, error)
	UpdateDocumentCycleObligation(ctx context.Context, arg db.UpdateDocumentCycleObligationParams) (db.DocumentCycleObligation, error)
	UpdateDocumentCycleObligationStatus(ctx context.Context, arg db.UpdateDocumentCycleObligationStatusParams) (db.DocumentCycleObligation, error)
	DeleteDocumentCycleObligation(ctx context.Context, id pgtype.UUID) error
	CreateDocumentCycleEvent(ctx context.Context, arg db.CreateDocumentCycleEventParams) (db.DocumentCycleEvent, error)
}

type DocumentCycle struct{ q documentCycleStore }

func NewDocumentCycle(q *db.Queries) *DocumentCycle { return &DocumentCycle{q: q} }

type DocumentCycleGenerateResult struct {
	PeriodYear int32 `json:"period_year"`
	Generated  int   `json:"generated"`
	Inserted   int   `json:"inserted"`
}

func (s *DocumentCycle) Stats(ctx context.Context, periodYear int32) (db.GetDocumentCycleStatsRow, error) {
	return s.q.GetDocumentCycleStats(ctx, normalizeDocumentCycleYear(periodYear))
}

func (s *DocumentCycle) ListCatalogs(ctx context.Context, search, frequency string, activeOnly bool) ([]db.ListDocumentCycleCatalogsRow, error) {
	return s.q.ListDocumentCycleCatalogs(ctx, db.ListDocumentCycleCatalogsParams{
		Search:     strings.TrimSpace(search),
		Frequency:  normalizeDocumentCycleFrequencyFilter(frequency),
		ActiveOnly: activeOnly,
	})
}

func (s *DocumentCycle) CreateCatalog(ctx context.Context, arg db.CreateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	arg = normalizeDocumentCycleCatalogCreate(arg)
	if err := validateDocumentCycleCatalog(arg.Code, arg.Title, arg.Frequency, arg.SnpStandard, arg.DeadlineDaysAfterPeriod, arg.ReminderDaysBeforeDue); err != nil {
		return db.DocumentCycleCatalog{}, err
	}
	return s.q.CreateDocumentCycleCatalog(ctx, arg)
}

func (s *DocumentCycle) UpdateCatalog(ctx context.Context, arg db.UpdateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	arg = normalizeDocumentCycleCatalogUpdate(arg)
	if err := validateDocumentCycleCatalog(arg.Code, arg.Title, arg.Frequency, arg.SnpStandard, arg.DeadlineDaysAfterPeriod, arg.ReminderDaysBeforeDue); err != nil {
		return db.DocumentCycleCatalog{}, err
	}
	return s.q.UpdateDocumentCycleCatalog(ctx, arg)
}

func (s *DocumentCycle) DeleteCatalog(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteDocumentCycleCatalog(ctx, id)
}

func (s *DocumentCycle) ListObligations(ctx context.Context, arg db.ListDocumentCycleObligationsParams) ([]db.ListDocumentCycleObligationsRow, error) {
	arg.Search = strings.TrimSpace(arg.Search)
	arg.Status = normalizeDocumentCycleStatusFilter(arg.Status)
	arg.Frequency = normalizeDocumentCycleFrequencyFilter(arg.Frequency)
	arg.PeriodYear = normalizeDocumentCycleYear(arg.PeriodYear)
	return s.q.ListDocumentCycleObligations(ctx, arg)
}

func (s *DocumentCycle) UpdateObligation(ctx context.Context, actorUserID pgtype.UUID, arg db.UpdateDocumentCycleObligationParams) (db.DocumentCycleObligation, error) {
	arg.Notes = strings.TrimSpace(arg.Notes)
	arg.VerificationNotes = strings.TrimSpace(arg.VerificationNotes)
	if !arg.DueDate.Valid || !arg.ReminderDate.Valid {
		return db.DocumentCycleObligation{}, fmt.Errorf("tanggal pengingat dan jatuh tempo wajib diisi")
	}
	if arg.ReminderDate.Time.After(arg.DueDate.Time) {
		return db.DocumentCycleObligation{}, fmt.Errorf("tanggal pengingat tidak boleh setelah jatuh tempo")
	}
	row, err := s.q.UpdateDocumentCycleObligation(ctx, arg)
	if err != nil {
		return db.DocumentCycleObligation{}, err
	}
	_, _ = s.q.CreateDocumentCycleEvent(ctx, db.CreateDocumentCycleEventParams{
		ObligationID: row.ID,
		EventType:    "updated",
		Notes:        "Kewajiban dokumen diperbarui.",
		ActorUserID:  actorUserID,
	})
	return row, nil
}

func (s *DocumentCycle) UpdateObligationStatus(ctx context.Context, actorUserID pgtype.UUID, id pgtype.UUID, status, notes string) (db.DocumentCycleObligation, error) {
	status = normalizeDocumentCycleStatus(status)
	notes = strings.TrimSpace(notes)
	if !validDocumentCycleStatuses[status] {
		return db.DocumentCycleObligation{}, fmt.Errorf("status dokumen tidak valid")
	}
	row, err := s.q.UpdateDocumentCycleObligationStatus(ctx, db.UpdateDocumentCycleObligationStatusParams{
		ID:     id,
		Status: status,
		Notes:  notes,
	})
	if err != nil {
		return db.DocumentCycleObligation{}, err
	}
	_, _ = s.q.CreateDocumentCycleEvent(ctx, db.CreateDocumentCycleEventParams{
		ObligationID: row.ID,
		EventType:    "status_changed",
		ToStatus:     status,
		Notes:        notes,
		ActorUserID:  actorUserID,
	})
	return row, nil
}

func (s *DocumentCycle) DeleteObligation(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteDocumentCycleObligation(ctx, id)
}

func (s *DocumentCycle) GenerateYear(ctx context.Context, actorUserID pgtype.UUID, periodYear int32) (DocumentCycleGenerateResult, error) {
	periodYear = normalizeDocumentCycleYear(periodYear)
	if periodYear < 2000 {
		return DocumentCycleGenerateResult{}, fmt.Errorf("tahun dokumen tidak valid")
	}
	result, err := s.q.GenerateDocumentCycleYearObligations(ctx, db.GenerateDocumentCycleYearObligationsParams{
		PeriodYear:      periodYear,
		CreatedByUserID: actorUserID,
	})
	if err != nil {
		return DocumentCycleGenerateResult{}, err
	}
	return DocumentCycleGenerateResult{
		PeriodYear: periodYear,
		Generated:  int(result.Generated),
		Inserted:   int(result.Inserted),
	}, nil
}

func normalizeDocumentCycleYear(year int32) int32 {
	if year == 0 {
		return int32(time.Now().Year())
	}
	return year
}

func normalizeDocumentCycleCatalogCreate(arg db.CreateDocumentCycleCatalogParams) db.CreateDocumentCycleCatalogParams {
	arg.Code = strings.ToUpper(strings.TrimSpace(arg.Code))
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Frequency = normalizeDocumentCycleFrequency(arg.Frequency)
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.RegulationRef = strings.TrimSpace(arg.RegulationRef)
	arg.Description = strings.TrimSpace(arg.Description)
	return arg
}

func normalizeDocumentCycleCatalogUpdate(arg db.UpdateDocumentCycleCatalogParams) db.UpdateDocumentCycleCatalogParams {
	arg.Code = strings.ToUpper(strings.TrimSpace(arg.Code))
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Frequency = normalizeDocumentCycleFrequency(arg.Frequency)
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.RegulationRef = strings.TrimSpace(arg.RegulationRef)
	arg.Description = strings.TrimSpace(arg.Description)
	return arg
}

func normalizeDocumentCycleFrequency(value string) string {
	return normalizeGovernanceText(strings.ToLower(strings.TrimSpace(value)), "monthly")
}

func normalizeDocumentCycleFrequencyFilter(value string) string {
	frequency := strings.ToLower(strings.TrimSpace(value))
	if frequency == "" {
		return ""
	}
	return frequency
}

func normalizeDocumentCycleStatus(value string) string {
	return normalizeGovernanceText(strings.ToLower(strings.TrimSpace(value)), "not_started")
}

func normalizeDocumentCycleStatusFilter(value string) string {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == "" {
		return ""
	}
	return status
}

func validateDocumentCycleCatalog(code, title, frequency, snpStandard string, deadlineDays, reminderDays int32) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode dokumen wajib diisi")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("nama dokumen wajib diisi")
	}
	if !validDocumentCycleFrequencies[frequency] {
		return fmt.Errorf("frekuensi dokumen tidak valid")
	}
	if !validGovernanceSNPStandards[snpStandard] {
		return fmt.Errorf("standar SNP tidak valid")
	}
	if deadlineDays < 0 || deadlineDays > 365 {
		return fmt.Errorf("batas jatuh tempo harus 0 sampai 365 hari")
	}
	if reminderDays < 0 || reminderDays > 60 {
		return fmt.Errorf("batas pengingat harus 0 sampai 60 hari")
	}
	return nil
}

var validDocumentCycleFrequencies = map[string]bool{
	"daily":     true,
	"weekly":    true,
	"monthly":   true,
	"quarterly": true,
	"semester":  true,
	"annual":    true,
	"four_year": true,
	"five_year": true,
}

var validDocumentCycleStatuses = map[string]bool{
	"not_started":          true,
	"draft":                true,
	"waiting_verification": true,
	"completed":            true,
}
