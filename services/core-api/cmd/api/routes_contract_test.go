package main

import (
	"os"
	"strings"
	"testing"
)

func TestNativeBankSoalRoutesUseGranularQuestionPermissions(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "// Native Bank Soal API aliases")
	end := strings.Index(source, "// Native Asesmen API aliases")
	if start < 0 || end <= start {
		t.Fatalf("native Bank Soal route block not found")
	}
	block := source[start:end]
	if strings.Contains(block, "requireCbt)") || strings.Contains(block, "With(requireCbt)") {
		t.Fatalf("native Bank Soal API block must not use requireCbt fallback guards:\n%s", block)
	}
	required := []string{
		`r.With(requireBankSoalRead).Get("/api/bank-soal/questions/summary"`,
		`r.With(requireBankSoalRead).Get("/api/bank-soal/questions"`,
		`r.With(requireBankSoalCreate).Post("/api/bank-soal/questions"`,
		`r.With(requireBankSoalReviewWorkflow).Patch("/api/bank-soal/questions/bulk-workflow"`,
		`r.With(requireBankSoalImport).Post("/api/bank-soal/questions/import-legacy"`,
		`r.With(requireBankSoalUpdate).Put("/api/bank-soal/questions/{id}"`,
		`r.With(requireBankSoalDelete).Delete("/api/bank-soal/questions/{id}"`,
		`r.With(requireBankSoalAssetUpload).Post("/api/bank-soal/assets"`,
	}
	for _, want := range required {
		if !strings.Contains(block, want) {
			t.Fatalf("native Bank Soal API block missing granular guard %q:\n%s", want, block)
		}
	}
}

func TestQuestionServiceUsesPoolBackedTransactionsInAPI(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if !strings.Contains(string(raw), "questionSvc := service.NewCbtQuestionWithPool(pool)") {
		t.Fatalf("API must construct CbtQuestion with pool-backed transaction support")
	}
}
