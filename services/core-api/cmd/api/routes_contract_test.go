package main

import (
	"os"
	"path/filepath"
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
		`r.Get("/api/exam/commands", examH.Commands)`,
		`r.Post("/api/exam/commands/{cid}/ack", examH.AcknowledgeCommand)`,
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

func TestInternalAnalyticsIngestionRouteStaysJWTProtected(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	jwtStart := strings.Index(source, "r.Use(mw.JWT(jwtSecret")
	workerStart := strings.Index(source, "r.Use(mw.WorkerKey(workerKey))")
	if jwtStart < 0 || workerStart <= jwtStart {
		t.Fatalf("authenticated route block not found")
	}
	authenticatedBlock := source[jwtStart:workerStart]
	const analyticsRoute = `r.With(analyticsIngestionRateLimit).Post("/api/internal-analytics/events", internalAnalyticsH.CreateEvent)`
	if !strings.Contains(authenticatedBlock, analyticsRoute) {
		t.Fatalf("internal analytics ingestion route must be registered only inside the JWT-authenticated API block")
	}
	if strings.Contains(source[:jwtStart], "/api/internal-analytics/events") {
		t.Fatalf("internal analytics ingestion must not be exposed before JWT middleware")
	}
	if strings.Contains(source, `"/api/internal-analytics/events/export"`) || strings.Contains(source, `"/api/internal-analytics/raw"`) {
		t.Fatalf("internal analytics must not expose raw/export event ingestion routes")
	}
}

func TestPublicAnalyticsCollectorUsesInternalKeyBeforeJWT(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	jwtStart := strings.Index(source, "r.Use(mw.JWT(jwtSecret")
	if jwtStart < 0 {
		t.Fatalf("authenticated route block not found")
	}
	publicRoute := `r.With(publicAnalyticsRateLimit, mw.InternalKey(internalAPIKey)).Post("/api/internal-analytics/public-events", internalAnalyticsH.CreatePublicEvent)`
	if !strings.Contains(source[:jwtStart], publicRoute) {
		t.Fatalf("public analytics collector must be before JWT and guarded by internal key route %q", publicRoute)
	}
	for _, want := range []string{
		`internalAPIKey := getEnv("INTERNAL_API_KEY", "")`,
		`publicAnalyticsRateLimit := ratelimit.RateLimitWithTrustedProxies`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("public analytics collector setup missing %q", want)
		}
	}
}

func TestInternalAnalyticsReadRoutesUseAnalyticsReadPermission(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	jwtStart := strings.Index(source, "r.Use(mw.JWT(jwtSecret")
	workerStart := strings.Index(source, "r.Use(mw.WorkerKey(workerKey))")
	if jwtStart < 0 || workerStart <= jwtStart {
		t.Fatalf("authenticated route block not found")
	}
	authenticatedBlock := source[jwtStart:workerStart]
	if !strings.Contains(source, `requireAnalyticsRead := mw.RequireAnyPermissionOrRole([]string{"analytics.read"}, "admin")`) {
		t.Fatalf("internal analytics read routes must define an analytics.read permission-first guard with admin fallback")
	}
	for _, want := range []string{
		`r.With(requireAnalyticsRead).Get("/api/internal-analytics/summary", internalAnalyticsH.Summary)`,
		`r.With(requireAnalyticsRead).Get("/api/internal-analytics/daily", internalAnalyticsH.ListDailyAggregates)`,
	} {
		if !strings.Contains(authenticatedBlock, want) {
			t.Fatalf("internal analytics read route contract missing %q inside JWT block", want)
		}
	}
	if strings.Contains(source[:jwtStart], "/api/internal-analytics/summary") || strings.Contains(source[:jwtStart], "/api/internal-analytics/daily") {
		t.Fatalf("internal analytics read routes must not be exposed before JWT middleware")
	}
}

func TestInternalAnalyticsExportRouteUsesExportPermissionAndRateLimitedIngestion(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	jwtStart := strings.Index(source, "r.Use(mw.JWT(jwtSecret")
	workerStart := strings.Index(source, "r.Use(mw.WorkerKey(workerKey))")
	if jwtStart < 0 || workerStart <= jwtStart {
		t.Fatalf("authenticated route block not found")
	}
	authenticatedBlock := source[jwtStart:workerStart]
	if !strings.Contains(source, `analyticsIngestionRateLimit := ratelimit.RateLimitWithTrustedProxies`) {
		t.Fatalf("internal analytics ingestion must use existing trusted-proxy-aware rate limiter")
	}
	if !strings.Contains(source, `requireAnalyticsExport := mw.RequireAnyPermissionOrRole([]string{"analytics.export"}, "admin")`) {
		t.Fatalf("internal analytics export route must define analytics.export permission-first guard with admin fallback")
	}
	required := []string{
		`r.With(analyticsIngestionRateLimit).Post("/api/internal-analytics/events", internalAnalyticsH.CreateEvent)`,
		`r.With(requireAnalyticsExport).Get("/api/internal-analytics/export", internalAnalyticsH.ExportAggregates)`,
	}
	for _, want := range required {
		if !strings.Contains(authenticatedBlock, want) {
			t.Fatalf("internal analytics route contract missing %q inside JWT block", want)
		}
	}
	for _, forbidden := range []string{
		`"/api/internal-analytics/events/export"`,
		`"/api/internal-analytics/raw"`,
		`"/api/internal-analytics/metadata"`,
		`GetInternalAnalyticsEvent`,
	} {
		if strings.Contains(authenticatedBlock, forbidden) {
			t.Fatalf("internal analytics must not expose raw event route/helper %q", forbidden)
		}
	}
}

func TestInternalAnalyticsAdminActivityMiddlewareAndRollupLoopAreStarted(t *testing.T) {
	raw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(raw)
	for _, want := range []string{
		`internalAnalyticsRollupCancel := internalAnalyticsSvc.StartRollupLoop(mainCtx, service.InternalAnalyticsRollupLoopConfig{`,
		`Interval:       durationEnv("INTERNAL_ANALYTICS_ROLLUP_INTERVAL", 10*time.Minute),`,
		`internalAnalyticsRollupCancel()`,
		`r.Use(mw.InternalAnalytics(internalAnalyticsSvc))`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("internal analytics startup contract missing %q", want)
		}
	}
}

func TestInternalAnalyticsPhase3BFFIsOnlyInternalWebAdminCollector(t *testing.T) {
	routePath := filepath.Clean("../../../../apps/web-admin/src/routes/api/internal-analytics/events/+server.ts")
	raw, err := os.ReadFile(routePath)
	if err != nil {
		t.Fatalf("phase 3 must add Web Admin BFF proxy route: %v", err)
	}
	source := string(raw)
	required := []string{
		"proxy(event).post('/api/internal-analytics/events'",
		"readRequestJson",
		"event.locals.user",
	}
	for _, want := range required {
		if !strings.Contains(source, want) {
			t.Fatalf("internal analytics BFF route missing %q", want)
		}
	}
	forbidden := []string{"X-Internal-Key", "metadata)", "console.log", "navigator.sendBeacon", "sendBeacon("}
	for _, needle := range forbidden {
		if strings.Contains(source, needle) {
			t.Fatalf("internal analytics BFF route must not contain %q", needle)
		}
	}
}

func TestInternalAnalyticsFrontendKeepsCollectorFirstPartyAndNoThirdPartyTracking(t *testing.T) {
	root := filepath.Clean("../../../../apps/web-admin/src")
	forbiddenNeedles := []string{
		"/api/public/internal-analytics",
		"navigator.sendBeacon",
		"sendBeacon(",
		"posthog",
		"plausible",
		"gtag(",
		"GoogleAnalytics",
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".spec.ts") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(filepath.ToSlash(path), "/routes/api/internal-analytics/events/+server.ts") {
			return nil
		}
		source := string(raw)
		for _, needle := range forbiddenNeedles {
			if strings.Contains(source, needle) {
					t.Fatalf("internal analytics must not add third-party tracking; found %q in %s", needle, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan web-admin source: %v", err)
	}
}

func TestInternalAnalyticsNoThirdPartyAnalyticsDependenciesAcrossRuntimeUnits(t *testing.T) {
	root := filepath.Clean("../../../..")
	forbiddenNeedles := []string{
		"posthog",
		"plausible",
		"google-analytics",
		"@vercel/analytics",
		"hotjar",
		"gtag(",
		"GoogleAnalytics",
		"navigator.sendBeacon",
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		slash := filepath.ToSlash(path)
		if d.IsDir() {
			if strings.Contains(slash, "/node_modules") || strings.Contains(slash, "/.git") || strings.Contains(slash, "/.svelte-kit") || strings.Contains(slash, "/build") {
				return filepath.SkipDir
			}
			return nil
		}
		if !(strings.HasSuffix(path, "package.json") || strings.HasSuffix(path, ".svelte") || strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".dart") || strings.HasSuffix(path, ".html")) {
			return nil
		}
		if strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".spec.ts") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(raw)
		for _, needle := range forbiddenNeedles {
			if strings.Contains(source, needle) {
				t.Fatalf("internal analytics must stay first-party only; found %q in %s", needle, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan runtime units: %v", err)
	}
}
