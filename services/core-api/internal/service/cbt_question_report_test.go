package service

import (
	"testing"
	"time"
)

func TestBankSoalReportRangeAllCoversHistoricalRows(t *testing.T) {
	start, end, label, err := bankSoalReportRange(BankSoalReportFilters{PeriodPreset: "all"})
	if err != nil {
		t.Fatalf("bankSoalReportRange returned error: %v", err)
	}
	if label != "Semua periode" {
		t.Fatalf("label=%q want %q", label, "Semua periode")
	}

	loc := bankSoalReportLocation()
	oldRowTime := time.Date(2023, 1, 1, 0, 0, 0, 0, loc)
	if start.After(oldRowTime) {
		t.Fatalf("start=%s should include old row time %s", start, oldRowTime)
	}
	if !end.After(time.Now().In(loc)) {
		t.Fatalf("end=%s should be after current time", end)
	}
}
