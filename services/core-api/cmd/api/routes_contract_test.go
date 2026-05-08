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
	assetFileGuard := `assetFileGuard := mw.ExamTokenOrJWT(jwtSecret, authSvc.CurrentAuthVersion, authSvc.ValidateAccessSession, examSvc.GetParticipantByToken)`
	assetFileAlias := `r.With(assetFileGuard).Get("/api/bank-soal/assets/{id}/file"`
	if !strings.Contains(source, assetFileGuard) || !strings.Contains(source, assetFileAlias) {
		t.Fatalf("native Bank Soal asset file alias must use the same ExamTokenOrJWT guard as the legacy CBT file route")
	}
}

func TestNativeAsesmenNonTestRoutesMatchLegacyGranularGuards(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "// Native Asesmen API aliases")
	end := strings.Index(source, `r.With(requireCbt).Get("/api/asesmen/packages"`)
	if start < 0 || end <= start {
		t.Fatalf("native Asesmen non-test route block not found")
	}
	block := source[start:end]
	if strings.Contains(block, "With(requireCbt)") {
		t.Fatalf("native Asesmen non-test routes must not use broad requireCbt guards:\n%s", block)
	}
	required := []string{
		`r.With(requireAsesmenRead).Get("/api/asesmen/non-test-assessments"`,
		`r.With(requireAsesmenScore).Post("/api/asesmen/non-test-assessments"`,
		`r.With(requireAsesmenRead).Get("/api/asesmen/non-test-assessments/{id}"`,
		`r.With(requireAsesmenScore).Put("/api/asesmen/non-test-assessments/{id}"`,
		`r.With(requireAsesmenScore).Delete("/api/asesmen/non-test-assessments/{id}"`,
		`r.With(requireAsesmenScore, requireGradesManage).Post("/api/asesmen/non-test-assessments/{id}/sync-grade"`,
		`r.With(requireAsesmenRead).Get("/api/asesmen/non-test-assessments/{id}/submissions"`,
		`r.With(requireAsesmenScore).Post("/api/asesmen/non-test-assessments/{id}/submissions/generate"`,
		`r.With(requireAsesmenScore).Post("/api/asesmen/non-test-assessments/{id}/submissions"`,
	}
	for _, want := range required {
		if !strings.Contains(block, want) {
			t.Fatalf("native Asesmen non-test API block missing guard %q:\n%s", want, block)
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

func TestExamMobileRoutesRemainTokenScoped(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "// Exam endpoints")
	end := strings.Index(source, "requireAdmin :=")
	if start < 0 || end <= start {
		t.Fatalf("exam route block not found")
	}
	block := source[start:end]
	required := []string{
		`r.With(examLoginRateLimit).Post("/api/exam/login", examH.Login)`,
		`r.Use(examTokenMW)`,
		`r.Get("/api/exam/status", examH.Status)`,
		`r.Post("/api/exam/heartbeat", examH.Heartbeat)`,
		`r.Post("/api/exam/event", examH.RecordEvent)`,
		`r.Post("/api/exam/answer", examH.SubmitAnswer)`,
		`r.Post("/api/exam/submit", examH.Submit)`,
	}
	for _, want := range required {
		if !strings.Contains(block, want) {
			t.Fatalf("exam route block missing %q:\n%s", want, block)
		}
	}
	if strings.Contains(block, `"/api/cbt/exam`) || strings.Contains(block, `"/api/cbt/student`) {
		t.Fatalf("mobile exam runtime must stay on /api/exam/*, not new /api/cbt/* routes:\n%s", block)
	}
}
