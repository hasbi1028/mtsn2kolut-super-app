package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"mtsn2kolut-super-app/backend/internal/handler"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	ratelimit "mtsn2kolut-super-app/backend/internal/middleware/rate_limit"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	_ = godotenv.Load()

	mainCtx, mainCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer mainCancel()

	pool, err := pgxpool.New(mainCtx, mustEnv("DATABASE_URL"))
	if err != nil {
		slog.Error("connect db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := db.New(pool)

	authSvc := service.NewAuth(q, mustEnv("JWT_SECRET"), getEnv("ADMIN_PASSWORD", ""), getEnv("AVATAR_DIR", "uploads/avatars"))
	academicSvc := service.NewAcademicWithPool(pool)
	rombelSvc := service.NewRombelWithPool(pool)
	gradeSvc := service.NewGrade(q)
	empSvc := service.NewEmployee(q)
	studentSvc := service.NewStudent(q)
	parentSvc := service.NewParent(q)
	portalSvc := service.NewPortal(q)
	studentPortalSvc := service.NewStudentPortal(q)
	parentPortalSvc := service.NewParentPortal(q)
	websiteSvc := service.NewWebsite(q)
	pusakaJobSvc := service.NewPusakaJobWithPool(pool)
	pusakaAttendanceSvc := service.NewPusakaAttendance(q)
	pusakaAttendanceTelegramSvc := service.NewPusakaAttendanceTelegram(q, getEnv("TELEGRAM_BOT_TOKEN", ""))
	cbtEventSvc := service.NewCbtEvent(pool)
	cbtSessionSvc := service.NewCbtSession(pool)
	questionAssetSvc := service.NewCbtQuestionAsset(q, getEnv("CBT_ASSET_DIR", "data/cbt-assets"))
	examSvc := service.NewExam(pool)
	pusakaScheduleSvc := service.NewPusakaSchedule(q)
	empSchedSvc := service.NewEmployeeSchedule(q)
	settSvc := service.NewSetting(q)
	auditSvc := service.NewAudit(q)
	internalAnalyticsSvc := service.NewInternalAnalytics(q)
	systemBackupSvc := service.NewSystemBackup(service.SystemBackupConfig{
		BackupDir:             getEnv("POSTGRES_BACKUP_DIR", service.DefaultSystemBackupDir),
		ScriptPath:            getEnv("POSTGRES_BACKUP_SCRIPT_PATH", service.DefaultSystemBackupScriptPath),
		TimerName:             getEnv("POSTGRES_BACKUP_TIMER_NAME", service.DefaultSystemBackupTimerName),
		ServiceName:           getEnv("POSTGRES_BACKUP_SERVICE_NAME", service.DefaultSystemBackupServiceName),
		RetentionDays:         int(int32Env("POSTGRES_BACKUP_RETENTION_DAYS", int32(service.DefaultSystemBackupRetentionDays))),
		OffsiteProvider:       getEnv("POSTGRES_BACKUP_OFFSITE_PROVIDER", ""),
		OffsiteTargetLabel:    getEnv("POSTGRES_BACKUP_OFFSITE_TARGET_LABEL", ""),
		OffsiteStatusFile:     getEnv("POSTGRES_BACKUP_OFFSITE_STATUS_FILE", ""),
		OffsiteRemoteDir:      getEnv("POSTGRES_BACKUP_OFFSITE_REMOTE_DIR", ""),
		OffsiteStaleThreshold: time.Duration(int32Env("POSTGRES_BACKUP_OFFSITE_STALE_HOURS", int32(service.DefaultOffsiteBackupStaleHours))) * time.Hour,
	})
	systemMaintenanceSvc := service.NewSystemMaintenance(q, systemBackupSvc)
	pusakaSchedulerSvc := service.NewPusakaScheduler(q, pusakaJobSvc, settSvc, auditSvc)
	notificationSvc := service.NewNotification(q)
	librarySvc := service.NewLibrary(q)
	inventorySvc := service.NewInventory(q)
	journalSvc := service.NewClassJournalWithPool(pool)
	letterSvc := service.NewLetter(q)
	governanceSvc := service.NewGovernance(q)
	documentCycleSvc := service.NewDocumentCycle(q)
	kesiswaanSvc := service.NewKesiswaanWithPool(pool, getEnv("STUDENT_PHOTO_DIR", "data/student-photos"))
	studentIDCardSvc := service.NewStudentIDCard(q)
	studentCertificateSvc := service.NewStudentCertificate(pool)
	archiveSvc := service.NewArchive(q, getEnv("ARCHIVE_DIR", "data/archives"))
	rbacSvc := service.NewRBACWithPool(pool)
	profileChangeRequestSvc := service.NewProfileChangeRequestWithPool(pool)
	websiteMediaH := handler.NewWebsiteMedia(getEnv("WEBSITE_MEDIA_DIR", "data/website-media"))
	brandingH := handler.NewBranding(settSvc, getEnv("BRANDING_ASSET_DIR", "data/branding"))

	if err := authSvc.SeedAdmin(mainCtx); err != nil {
		slog.Error("seed admin", "error", err)
		os.Exit(1)
	}
	if err := settSvc.SeedDefaults(mainCtx); err != nil {
		slog.Error("seed settings", "error", err)
		os.Exit(1)
	}
	pusakaSchedulerSvc.Start(mainCtx)
	pusakaAttendanceTelegramSvc.Start(mainCtx)
	internalAnalyticsRollupCancel := internalAnalyticsSvc.StartRollupLoop(mainCtx, service.InternalAnalyticsRollupLoopConfig{
		Interval:       durationEnv("INTERNAL_ANALYTICS_ROLLUP_INTERVAL", 10*time.Minute),
		Lookback:       durationEnv("INTERNAL_ANALYTICS_ROLLUP_LOOKBACK", 48*time.Hour),
		Limit:          int32Env("INTERNAL_ANALYTICS_ROLLUP_LIMIT", 10000),
		RunImmediately: boolEnv("INTERNAL_ANALYTICS_ROLLUP_RUN_IMMEDIATELY", true),
	})

	authH := handler.NewAuth(authSvc, q)
	academicH := handler.NewAcademic(academicSvc)
	rombelH := handler.NewRombel(rombelSvc)
	gradeH := handler.NewGrade(gradeSvc, q)
	empH := handler.NewEmployee(empSvc)
	healthH := handler.NewHealth(pool, pusakaJobSvc, settSvc, internalAnalyticsSvc)
	pusakaJobH := handler.NewPusakaJob(pusakaJobSvc)
	pusakaAttendanceH := handler.NewPusakaAttendance(pusakaAttendanceSvc)
	pusakaAttendanceTelegramH := handler.NewPusakaAttendanceTelegram(pusakaAttendanceTelegramSvc)
	studentH := handler.NewStudent(studentSvc)
	parentH := handler.NewParent(parentSvc)
	portalH := handler.NewPortal(portalSvc)
	studentPortalH := handler.NewStudentPortal(studentPortalSvc)
	parentPortalH := handler.NewParentPortal(parentPortalSvc)
	websiteH := handler.NewWebsite(websiteSvc)
	cbtEventH := handler.NewCbtEvent(cbtEventSvc)
	cbtSessionH := handler.NewCbtSession(cbtSessionSvc, q)
	questionAssetH := handler.NewCbtQuestionAsset(questionAssetSvc)
	userH := handler.NewUserWithPool(pool)
	pusakaScheduleH := handler.NewPusakaSchedule(pusakaScheduleSvc)
	empSchedH := handler.NewEmployeeSchedule(empSchedSvc)
	settH := handler.NewSetting(settSvc)
	internalAnalyticsH := handler.NewInternalAnalytics(internalAnalyticsSvc)
	systemBackupH := handler.NewSystemBackup(systemBackupSvc)
	systemMaintenanceH := handler.NewSystemMaintenance(systemMaintenanceSvc)
	pusakaSchedulerH := handler.NewPusakaScheduler(pusakaSchedulerSvc)
	pusakaWorkerH := handler.NewPusakaWorker(pusakaJobSvc, pusakaAttendanceSvc, settSvc)
	notificationH := handler.NewNotification(notificationSvc)
	libraryH := handler.NewLibrary(librarySvc, q)
	inventoryH := handler.NewInventory(inventorySvc, q)
	journalH := handler.NewClassJournal(journalSvc, q)
	letterH := handler.NewLetter(letterSvc, q)
	governanceH := handler.NewGovernance(governanceSvc, q)
	documentCycleH := handler.NewDocumentCycle(documentCycleSvc, q)
	kesiswaanH := handler.NewKesiswaan(kesiswaanSvc, q)
	studentIDCardH := handler.NewStudentIDCard(studentIDCardSvc, authSvc)
	studentCertificateH := handler.NewStudentCertificate(studentCertificateSvc)
	archiveH := handler.NewArchive(archiveSvc, q)
	rbacH := handler.NewRBAC(rbacSvc)
	profileChangeRequestH := handler.NewProfileChangeRequest(profileChangeRequestSvc)

	jwtSecret := mustEnv("JWT_SECRET")
	workerKey := mustEnv("WORKER_API_KEY")
	internalAPIKey := getEnv("INTERNAL_API_KEY", "")
	trustedProxies := splitCSVEnv("TRUSTED_PROXY_CIDRS")
	authRateLimit := ratelimit.RateLimitWithTrustedProxies(5, 1, trustedProxies)
	refreshRateLimit := ratelimit.RateLimitWithTrustedProxies(10, 1, trustedProxies)
	logoutRateLimit := ratelimit.RateLimitWithTrustedProxies(10, 2, trustedProxies)
	passwordRateLimit := ratelimit.RateLimitWithTrustedProxies(3, 0.1, trustedProxies)
	publicRegisterRateLimit := ratelimit.RateLimitWithTrustedProxies(3, 0.2, trustedProxies)
	studentCardPortalRateLimit := ratelimit.RateLimitWithTrustedProxies(10, 0.5, trustedProxies)
	publicSiteRateLimit := ratelimit.RateLimitWithTrustedProxies(60, 30, trustedProxies)
	analyticsIngestionRateLimit := ratelimit.RateLimitWithTrustedProxies(30, 10, trustedProxies)
	publicAnalyticsRateLimit := ratelimit.RateLimitWithTrustedProxies(20, 5, trustedProxies)

	r := chi.NewRouter()
	r.Use(mw.RequestID)
	r.Use(mw.CORS("http://localhost:7300", "http://127.0.0.1:7300"))
	r.Use(mw.RequestLog)
	r.Use(chimw.Recoverer)
	r.Use(chimw.SetHeader("Content-Type", "application/json"))

	r.Get("/health", healthH.Get)

	r.With(authRateLimit).Post("/api/auth/login", authH.Login)
	r.With(refreshRateLimit).Post("/api/auth/refresh", authH.Refresh)
	r.With(logoutRateLimit).Post("/api/auth/logout", authH.Logout)
	r.With(publicRegisterRateLimit).Post("/api/public/register-student", studentH.PublicRegister)
	r.With(publicSiteRateLimit).Get("/api/public/site/posts", websiteH.ListPublishedPosts)
	r.With(publicSiteRateLimit).Get("/api/public/site/posts/featured", websiteH.ListFeaturedPosts)
	r.With(publicSiteRateLimit).Get("/api/public/site/posts/{slug}", websiteH.GetPublishedPost)
	r.With(publicSiteRateLimit).Get("/api/public/site/announcements", websiteH.ListPublishedAnnouncements)
	r.With(publicSiteRateLimit).Get("/api/public/site/announcements/{slug}", websiteH.GetPublishedAnnouncement)
	r.With(publicSiteRateLimit).Get("/api/public/site/pages/{slug}", websiteH.GetPublishedPage)
	r.With(publicSiteRateLimit).Get("/api/public/branding", brandingH.Public)
	r.Get("/api/website/media/{filename}", websiteMediaH.File)
	r.Get("/api/branding/file/{filename}", brandingH.Asset)
	r.With(publicAnalyticsRateLimit, mw.InternalKey(internalAPIKey)).Post("/api/internal-analytics/public-events", internalAnalyticsH.CreatePublicEvent)
	r.Get("/api/system/maintenance/status", systemMaintenanceH.Status)
	r.With(publicSiteRateLimit).Get("/api/public/student-cards/verify/{token}", studentIDCardH.PublicVerify)
	r.With(studentCardPortalRateLimit).Post("/api/public/student-cards/portal-login/start", studentIDCardH.PortalLoginStart)
	r.With(studentCardPortalRateLimit).Post("/api/public/student-cards/portal-login/complete", studentIDCardH.PortalLoginComplete)

	// CBT asset files are not public-by-obscurity. They may be accessed either by
	// authenticated requests or active exam participants using the exam token
	// attached to exam payload asset URLs.
	assetFileGuard := mw.ExamTokenOrJWT(jwtSecret, authSvc.CurrentAuthVersion, authSvc.ValidateAccessSession, examSvc.GetParticipantByToken)
	r.With(assetFileGuard).Get("/api/cbt/assets/{id}/file", questionAssetH.File)

	requireAdmin := mw.RequireAdmin()
	requireCbtRead := mw.RequireAnyPermissionOrRole([]string{"cbt.read", "cbt.manage"}, "admin", "guru", "staf")
	requireCbtManage := mw.RequireAnyPermissionOrRole([]string{"cbt.manage"}, "admin", "guru")
	requireGradesRead := mw.RequireAnyPermission("grades.read", "grades.manage")
	requireGradesManage := mw.RequirePermission("grades.manage")
	requireJournalRead := mw.RequireAnyPermissionOrRole([]string{"journal.read", "journal.manage", "journal.read_all", "journal.manage_all"}, "admin", "guru")
	requireJournalManage := mw.RequireAnyPermissionOrRole([]string{"journal.manage", "journal.manage_all"}, "admin", "guru")
	requireStaff := mw.RequireAnyPermissionOrRole([]string{"library.read", "library.manage", "inventory.read", "inventory.manage", "letters.read", "letters.manage", "archives.read", "archives.manage", "governance.read", "governance.manage", "document_cycles.read", "document_cycles.manage"}, "admin", "staf")
	requireKesiswaanManage := mw.RequireAnyPermissionOrRole([]string{"students.manage", "kesiswaan.manage"}, "admin", "kesiswaan")
	requireAcademicManage := mw.RequireAnyPermissionOrRole([]string{"academic.manage"}, "admin")
	requireStudentsManage := mw.RequireAnyPermissionOrRole([]string{"students.manage"}, "admin", "kesiswaan")
	requireIDCardsRead := mw.RequireAnyPermissionOrRole([]string{"id_cards.read", "id_cards.manage"}, "admin", "kesiswaan")
	requireIDCardsManage := mw.RequireAnyPermissionOrRole([]string{"id_cards.manage"}, "admin", "kesiswaan")
	requireIDCardsScan := mw.RequireAnyPermissionOrRole([]string{"id_cards.scan", "id_cards.manage"}, "admin", "kesiswaan", "guru", "staf")
	requireIDCardsAudit := mw.RequireAnyPermissionOrRole([]string{"id_cards.audit", "id_cards.manage"}, "admin", "kesiswaan")
	requireParentsManage := mw.RequireAnyPermissionOrRole([]string{"parents.manage"}, "admin", "kesiswaan")
	requireWebsiteManage := mw.RequireAnyPermissionOrRole([]string{"website.manage"}, "admin")
	requirePusakaRead := mw.RequireAnyPermissionOrRole([]string{"pusaka.read", "pusaka.manage", "pusaka.sync", "pusaka.credentials_manage"}, "admin")
	requirePusakaManage := mw.RequireAnyPermissionOrRole([]string{"pusaka.manage", "pusaka.sync", "pusaka.credentials_manage"}, "admin")
	requireUsersRead := mw.RequirePermission("users.read")
	requireUsersCreate := mw.RequirePermission("users.create")
	requireUsersDeactivate := mw.RequirePermission("users.deactivate")
	requireUsersResetPassword := mw.RequirePermission("users.reset_password")
	requireUsersResetPasswordOrAdmin := mw.RequireAnyPermissionOrRole([]string{"users.reset_password"}, "admin")
	requireUsersUpdate := mw.RequirePermission("users.update")
	requireUsersManageRoles := mw.RequirePermission("users.manage_roles")
	requireStudentAccountsManage := mw.RequireAnyPermissionOrRole([]string{"student_accounts.manage"}, "admin")
	requireParentAccountsManage := mw.RequireAnyPermissionOrRole([]string{"parent_accounts.manage"}, "admin")
	requireStudentPortalRead := mw.RequireAnyPermissionOrRole([]string{"student_portal.read"}, "siswa")
	requireParentPortalRead := mw.RequireAnyPermissionOrRole([]string{"parent_portal.read"}, "ortu")
	requireProfileChangesReview := mw.RequireAnyPermissionOrRole([]string{service.ProfileChangesReviewPermission}, "admin")
	requireRolesRead := mw.RequirePermission("roles.read")
	requireRolesManage := mw.RequirePermission("roles.manage")
	requireAuditRead := mw.RequirePermission("audit.read")
	requireAnalyticsRead := mw.RequireAnyPermissionOrRole([]string{"analytics.read"}, "admin")
	requireAnalyticsExport := mw.RequireAnyPermissionOrRole([]string{"analytics.export"}, "admin")
	requireBackupRead := mw.RequireAnyPermissionOrRole([]string{"backup.read"}, "admin")
	requireBackupDownload := mw.RequireAnyPermissionOrRole([]string{"backup.download"}, "admin")
	requireBackupCreate := mw.RequireAnyPermissionOrRole([]string{"backup.create"}, "admin")
	requireBackupRestorePlan := mw.RequireAnyPermissionOrRole([]string{"backup.restore_plan"}, "admin")
	requireSchoolProfileSettings := mw.RequirePermission("settings.school_profile")
	requireBrandingSettings := mw.RequireAnyPermissionOrRole([]string{"settings.branding"}, "admin")
	requireEmployeesRead := mw.RequireAnyPermissionOrRole([]string{"employees.read", "employees.manage"}, "admin")
	requireEmployeesManage := mw.RequireAnyPermissionOrRole([]string{"employees.manage"}, "admin")

	r.Group(func(r chi.Router) {
		r.Use(mw.JWT(jwtSecret, authSvc.CurrentAuthVersion, authSvc.ValidateAccessSession))
		r.Use(mw.Maintenance(systemMaintenanceSvc, 5*time.Second))
		r.Use(mw.Audit(q))
		r.Use(mw.InternalAnalytics(internalAnalyticsSvc))
		r.Get("/api/auth/account", authH.GetAccount)
		r.Get("/api/auth/account/change-history", authH.GetAccountChangeHistory)
		r.Patch("/api/auth/account/contact", authH.UpdateAccountContact)
		r.Get("/api/auth/account/change-request-fields", profileChangeRequestH.ListSelfRequestableFields)
		r.Get("/api/auth/account/change-requests", profileChangeRequestH.ListOwn)
		r.Post("/api/auth/account/change-requests", profileChangeRequestH.CreateOwn)
		r.Post("/api/auth/account/change-requests/{id}/cancel", profileChangeRequestH.CancelOwn)
		r.Post("/api/auth/account/avatar", authH.UploadAccountAvatar)
		r.Delete("/api/auth/account/avatar", authH.DeleteAccountAvatar)
		r.Get("/api/auth/account/avatar/{filename}", authH.AccountAvatarFile)
		r.With(passwordRateLimit).Post("/api/auth/change-password", authH.ChangePassword)
		r.Post("/api/auth/logout-all", authH.LogoutAll)
		r.Get("/api/auth/sessions", authH.ListSessions)
		r.Get("/api/auth/preferences/sidebar", authH.GetSidebarPreferences)
		r.Patch("/api/auth/preferences/sidebar", authH.UpdateSidebarPreferences)
		r.Patch("/api/auth/sessions/{id}", authH.UpdateSessionLabel)
		r.Delete("/api/auth/sessions/{id}", authH.RevokeSession)
		r.Get("/api/notifications", notificationH.List)
		r.Get("/api/notifications/unread-count", notificationH.CountUnread)
		r.Post("/api/notifications/read-all", notificationH.MarkAllRead)
		r.Post("/api/notifications/{id}/read", notificationH.MarkRead)
		r.With(analyticsIngestionRateLimit).Post("/api/internal-analytics/events", internalAnalyticsH.CreateEvent)
		r.With(requireAnalyticsRead).Get("/api/internal-analytics/summary", internalAnalyticsH.Summary)
		r.With(requireAnalyticsRead).Get("/api/internal-analytics/daily", internalAnalyticsH.ListDailyAggregates)
		r.With(requireAnalyticsExport).Get("/api/internal-analytics/export", internalAnalyticsH.ExportAggregates)
		r.With(requireBackupRead).Get("/api/system/backups/status", systemBackupH.Status)
		r.With(requireBackupRead).Get("/api/system/backups/offsite", systemBackupH.OffsiteStatus)
		r.With(requireBackupRead).Get("/api/system/backups", systemBackupH.List)
		r.With(requireBackupCreate).Post("/api/system/backups/run", systemBackupH.RunManual)
		r.With(requireBackupRead).Get("/api/system/backups/jobs/{job_id}", systemBackupH.Job)
		r.With(requireBackupRestorePlan).Post("/api/system/backups/{id}/validate-restore", systemBackupH.ValidateRestore)
		r.With(requireBackupRestorePlan).Post("/api/system/backups/{id}/restore-command", systemBackupH.RestoreCommand)
		r.With(requireBackupDownload).Get("/api/system/backups/{id}/download", systemBackupH.Download)
		r.With(requireAdmin).Get("/api/system/maintenance/health-summary", systemMaintenanceH.HealthSummary)
		r.With(requireAdmin).Get("/api/system/maintenance/windows", systemMaintenanceH.ListWindows)
		r.With(requireAdmin).Post("/api/system/maintenance/windows", systemMaintenanceH.CreateWindow)
		r.With(requireAdmin).Put("/api/system/maintenance/windows/{id}", systemMaintenanceH.UpdateWindow)
		r.With(requireAdmin).Post("/api/system/maintenance/windows/{id}/activate", systemMaintenanceH.ActivateWindow)
		r.With(requireAdmin).Post("/api/system/maintenance/windows/{id}/deactivate", systemMaintenanceH.DeactivateWindow)
		r.With(requireAdmin).Get("/api/system/maintenance/audit-logs", systemMaintenanceH.ListAuditLogs)
		r.Get("/api/school-profile", settH.SchoolProfile)
		r.With(requireSchoolProfileSettings).Put("/api/school-profile", settH.UpdateSchoolProfile)
		r.With(requireBrandingSettings).Get("/api/branding", brandingH.Get)
		r.With(requireBrandingSettings).Put("/api/branding", brandingH.Update)
		r.With(requireBrandingSettings).Post("/api/branding/assets/{purpose}", brandingH.UploadAsset)
		r.With(requireBrandingSettings).Delete("/api/branding/assets/{purpose}", brandingH.ResetAsset)

		// Employees use dynamic RBAC; admin keeps the default grant.
		r.Group(func(r chi.Router) {
			r.With(requireEmployeesRead).Get("/api/employees", empH.List)
			r.With(requireEmployeesManage).Post("/api/employees", empH.Create)
			r.With(requireEmployeesRead).Get("/api/employees/{id}", empH.Get)
			r.With(requireEmployeesManage).Put("/api/employees/{id}", empH.Update)
			r.With(requireEmployeesManage).Patch("/api/employees/{id}/status", empH.UpdateStatus)
			r.With(requireEmployeesManage).Delete("/api/employees/{id}", empH.Delete)
			r.With(requirePusakaRead).Get("/api/pusaka/employees", empH.ListPusakaEligibleWithStatus)
			r.With(requirePusakaManage).Patch("/api/pusaka/employees/{id}/account-status", empH.UpdatePusakaAccountStatus)
			r.With(requirePusakaManage).Delete("/api/pusaka/employees/{id}/account", empH.DeletePusakaAccount)
			r.With(requirePusakaRead).Get("/api/pusaka/employees/{id}/audit-logs", empH.ListPusakaAuditLogs)
			r.With(requirePusakaRead).Get("/api/pusaka/employees/{id}/schedules", empSchedH.List)
			r.With(requirePusakaManage).Post("/api/pusaka/employees/{id}/schedules", empSchedH.Upsert)
			r.With(requirePusakaManage).Delete("/api/pusaka/employees/{id}/schedules/{scheduleId}", empSchedH.Delete)
		})

		r.Get("/api/academic", academicH.Overview)
		r.Get("/api/academic/subjects", academicH.ListSubjectsOnly)
		r.Get("/api/academic/dashboard", academicH.GetDashboard)
		r.Get("/api/academic/readiness", academicH.GetReadiness)
		r.Get("/api/academic/stats", academicH.GetStats)
		r.Get("/api/academic/curriculum", academicH.GetCurriculumOverview)
		r.Get("/api/academic/curriculum/profiles", academicH.ListCurriculumProfiles)
		r.Get("/api/academic/curriculum/allocations", academicH.ListCurriculumAllocations)
		r.Get("/api/academic/curriculum/summary", academicH.GetCurriculumSummary)
		r.Get("/api/academic/timetable/weekly", academicH.GetWeeklyTimetable)
		r.Get("/api/academic/timetable/conflicts", academicH.GetTimetableConflicts)
		r.Get("/api/academic/lesson-periods", academicH.GetLessonPeriods)
		r.With(requireAcademicManage).Post("/api/academic/lesson-periods", academicH.CreateLessonPeriod)
		r.With(requireAcademicManage).Put("/api/academic/lesson-periods/{id}", academicH.UpdateLessonPeriod)
		r.With(requireAcademicManage).Delete("/api/academic/lesson-periods/{id}", academicH.DeleteLessonPeriod)
		r.Get("/api/academic/teacher-workload", academicH.GetTeacherWorkload)
		r.With(requireAcademicManage).Post("/api/academic/years/{id}/activate", academicH.ActivateYear)
		r.With(requireAcademicManage).Post("/api/academic/year-rollover/preview", academicH.PreviewYearRollover)
		r.With(requireAcademicManage).Post("/api/academic/year-rollover/apply", academicH.ApplyYearRollover)
		r.With(requireAcademicManage).Post("/api/academic/import-export/dry-run", academicH.DryRunAcademicImport)
		r.Get("/api/academic/subject-assignment-matrix", rombelH.GetSubjectAssignmentMatrix)
		r.With(requireAcademicManage).Put("/api/academic/subject-assignment-matrix", rombelH.UpdateSubjectAssignmentMatrixCell)
		r.Get("/api/academic/rombel", rombelH.List)
		r.Get("/api/academic/rombel/{id}", rombelH.Get)
		r.With(requireAcademicManage).Put("/api/academic/rombel/{id}", rombelH.UpdateIdentity)
		r.Get("/api/academic/rombel/{id}/students", rombelH.ListStudents)
		r.Get("/api/academic/rombel/{id}/subject-assignments", rombelH.ListSubjectAssignments)
		r.Get("/api/academic/rombel/{id}/subject-assignments/{assignmentID}", rombelH.GetSubjectAssignment)
		r.With(requireAcademicManage).Post("/api/academic/rombel/{id}/subject-assignments", rombelH.CreateSubjectAssignment)
		r.With(requireAcademicManage).Put("/api/academic/rombel/{id}/subject-assignments/{assignmentID}", rombelH.UpdateSubjectAssignment)
		r.With(requireAcademicManage).Delete("/api/academic/rombel/{id}/subject-assignments/{assignmentID}", rombelH.DeleteSubjectAssignment)
		r.Get("/api/academic/rombel/{id}/timetable-slots", rombelH.ListTimetableSlots)
		r.Get("/api/academic/rombel/{id}/timetable-slots/{slotID}", rombelH.GetTimetableSlot)
		r.With(requireAcademicManage).Post("/api/academic/rombel/{id}/timetable-slots", rombelH.CreateTimetableSlot)
		r.With(requireAcademicManage).Put("/api/academic/rombel/{id}/timetable-slots/{slotID}", rombelH.UpdateTimetableSlot)
		r.With(requireAcademicManage).Delete("/api/academic/rombel/{id}/timetable-slots/{slotID}", rombelH.DeleteTimetableSlot)
		r.With(requireJournalManage).Post("/api/academic/rombel/{id}/timetable-slots/{slotID}/journal-session", journalH.OpenSessionFromTimetableSlot)
		r.Get("/api/academic/rombel/{id}/homeroom-assignments", rombelH.ListHomeroomAssignments)
		r.With(requireAcademicManage).Post("/api/academic/rombel/{id}/homeroom-assignments", rombelH.CreateHomeroomAssignment)
		r.With(requireAcademicManage).Put("/api/academic/rombel/{id}/homeroom-assignments/{assignmentID}", rombelH.UpdateHomeroomAssignment)
		r.With(requireAcademicManage).Delete("/api/academic/rombel/{id}/homeroom-assignments/{assignmentID}", rombelH.DeleteHomeroomAssignment)
		r.With(requireAcademicManage).Post("/api/academic/{entity}", academicH.Create)
		r.With(requireAcademicManage).Put("/api/academic/{entity}/{id}", academicH.Update)
		r.With(requireAcademicManage).Delete("/api/academic/{entity}/{id}", academicH.Delete)
		r.With(requireGradesRead).Get("/api/grades", gradeH.Overview)
		r.With(requireGradesRead).Get("/api/grades/report-settings", gradeH.GetReportSettings)
		r.With(requireGradesManage).Put("/api/grades/report-settings", gradeH.UpdateReportSettings)
		r.With(requireGradesManage).Post("/api/grades/components", gradeH.CreateComponent)
		r.With(requireGradesManage).Post("/api/grades/assignments/{id}/description", gradeH.UpsertStudentSubjectDescription)
		r.With(requireGradesManage).Post("/api/grades/assignments/{id}/finalize", gradeH.FinalizeAssignment)
		r.With(requireGradesManage).Delete("/api/grades/assignments/{id}/finalize", gradeH.ReopenAssignment)
		r.With(requireGradesManage).Put("/api/grades/components/{id}", gradeH.UpdateComponent)
		r.With(requireGradesManage).Patch("/api/grades/components/{id}/publish", gradeH.SetComponentPublished)
		r.With(requireGradesManage).Delete("/api/grades/components/{id}", gradeH.DeleteComponent)
		r.With(requireGradesManage).Post("/api/grades/components/{id}/entries", gradeH.UpsertEntry)

		r.With(requireJournalRead).Get("/api/journal", journalH.Overview)
		r.With(requireJournalManage).Post("/api/journal/sessions", journalH.CreateSession)
		r.With(requireJournalRead).Get("/api/journal/sessions/{id}", journalH.GetSession)
		r.With(requireJournalManage).Put("/api/journal/sessions/{id}", journalH.UpdateSession)
		r.With(requireAdmin).Delete("/api/journal/sessions/{id}", journalH.DeleteSession)
		r.With(requireJournalManage).Post("/api/journal/sessions/{id}/attendances", journalH.BulkUpsertAttendances)

		r.Get("/api/students", studentH.GuruAwareList)
		r.With(requireStudentsManage).Post("/api/students", studentH.Create)
		r.With(requireStudentsManage).Put("/api/students/{id}", studentH.Update)
		r.With(requireStudentsManage).Patch("/api/students/{id}/lifecycle", studentH.UpdateLifecycle)
		r.With(requireStudentsManage).Delete("/api/students/{id}", studentH.Delete)
		r.With(requireIDCardsRead).Get("/api/student-id-cards", studentIDCardH.List)
		r.With(requireIDCardsManage).Post("/api/student-id-cards/generate", studentIDCardH.Generate)
		r.With(requireIDCardsRead).Get("/api/student-id-cards/{id}", studentIDCardH.Get)
		r.With(requireIDCardsManage).Post("/api/student-id-cards/{id}/print-event", studentIDCardH.MarkPrinted)
		r.With(requireIDCardsManage).Patch("/api/student-id-cards/{id}/status", studentIDCardH.UpdateStatus)
		r.With(requireIDCardsManage).Post("/api/student-id-cards/{id}/reissue", studentIDCardH.Reissue)
		r.With(requireIDCardsRead).Get("/api/student-id-cards/{id}/events", studentIDCardH.Events)
		r.With(requireIDCardsAudit).Get("/api/student-id-cards/{id}/audit-logs", studentIDCardH.AuditLogs)
		r.With(requireIDCardsScan).Post("/api/student-id-cards/scan/attendance", studentIDCardH.AttendanceScan)
		r.With(requireIDCardsScan).Post("/api/student-id-cards/scan/library", studentIDCardH.LibraryScan)

		r.Get("/api/parents", parentH.List)
		r.With(requireParentsManage).Post("/api/parents", parentH.Create)
		r.Get("/api/parents/{id}", parentH.Get)
		r.With(requireParentsManage).Put("/api/parents/{id}", parentH.Update)
		r.With(requireParentsManage).Delete("/api/parents/{id}", parentH.Delete)
		r.Get("/api/parents/{id}/children", parentH.ListChildren)
		r.With(requireParentsManage).Post("/api/parents/{id}/link", parentH.LinkStudent)
		r.With(requireParentsManage).Post("/api/parents/{id}/unlink", parentH.UnlinkStudent)
		r.With(requireStudentPortalRead).Get("/api/portal/student/me", portalH.StudentMe)
		r.With(requireStudentPortalRead).Get("/api/portal/student/profile", studentPortalH.Profile)
		r.With(requireStudentPortalRead).Get("/api/portal/student/schedule", studentPortalH.Schedule)
		r.With(requireStudentPortalRead).Get("/api/portal/student/results", studentPortalH.Results)
		r.With(requireStudentsManage).Get("/api/portal/preview/students", studentPortalH.PreviewStudents)
		r.With(requireStudentsManage).Get("/api/portal/preview/students/{studentID}/profile", studentPortalH.PreviewProfile)
		r.With(requireStudentsManage).Get("/api/portal/preview/students/{studentID}/schedule", studentPortalH.PreviewSchedule)
		r.With(requireStudentsManage).Get("/api/portal/preview/students/{studentID}/results", studentPortalH.PreviewResults)
		r.Get("/api/portal/guru/timetable", portalH.TeacherTimetable)
		r.With(requireParentPortalRead).Get("/api/portal/parent/me", portalH.ParentMe)
		r.With(requireParentPortalRead).Get("/api/portal/parent/children", parentPortalH.Children)
		r.With(requireParentPortalRead).Get("/api/portal/parent/children/{studentID}/profile", parentPortalH.ChildProfile)
		r.With(requireParentPortalRead).Get("/api/portal/parent/children/{studentID}/schedule", parentPortalH.ChildSchedule)
		r.With(requireParentPortalRead).Get("/api/portal/parent/children/{studentID}/results", parentPortalH.ChildResults)
		r.With(requireParentsManage).Get("/api/portal/preview/parents", parentPortalH.PreviewParents)
		r.With(requireParentsManage).Get("/api/portal/preview/parents/{parentID}/children", parentPortalH.PreviewChildren)
		r.With(requireParentsManage).Get("/api/portal/preview/parents/{parentID}/children/{studentID}/profile", parentPortalH.PreviewChildProfile)
		r.With(requireParentsManage).Get("/api/portal/preview/parents/{parentID}/children/{studentID}/schedule", parentPortalH.PreviewChildSchedule)
		r.With(requireParentsManage).Get("/api/portal/preview/parents/{parentID}/children/{studentID}/results", parentPortalH.PreviewChildResults)
		r.With(requireWebsiteManage).Get("/api/website/content", websiteH.List)
		r.With(requireWebsiteManage).Post("/api/website/content", websiteH.Create)
		r.With(requireWebsiteManage).Put("/api/website/content/{id}", websiteH.Update)
		r.With(requireWebsiteManage).Delete("/api/website/content/{id}", websiteH.Delete)
		r.With(requireWebsiteManage).Post("/api/website/media", websiteMediaH.Upload)

		// Legacy CBT runtime contracts retained for existing clients.
		r.With(requireCbtRead).Get("/api/cbt/events", cbtEventH.List)
		r.With(requireCbtManage).Post("/api/cbt/events", cbtEventH.Create)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}", cbtEventH.Get)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}/overview", cbtEventH.Overview)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}/readiness", cbtEventH.Readiness)
		r.With(requireCbtManage).Put("/api/cbt/events/{id}", cbtEventH.Update)
		r.With(requireCbtManage).Patch("/api/cbt/events/{id}/status", cbtEventH.UpdateStatus)
		r.With(requireCbtManage).Delete("/api/cbt/events/{id}", cbtEventH.Delete)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}/packages", cbtEventH.ListPackages)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}/sessions", cbtEventH.ListSessions)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}/question-targets", cbtEventH.ListQuestionTargets)
		r.With(requireCbtManage).Put("/api/cbt/events/{id}/question-targets", cbtEventH.UpsertQuestionTarget)
		r.With(requireCbtRead).Get("/api/cbt/events/{id}/members", cbtEventH.ListMembers)
		r.With(requireCbtManage).Post("/api/cbt/events/{id}/members", cbtEventH.CreateMember)
		r.With(requireCbtManage).Put("/api/cbt/events/{id}/members/{member_id}", cbtEventH.UpdateMember)
		r.With(requireCbtManage).Delete("/api/cbt/events/{id}/members/{member_id}", cbtEventH.DeleteMember)
		r.With(requireCbtRead).Get("/api/cbt/sessions", cbtSessionH.List)
		r.With(requireCbtRead).Get("/api/cbt/sessions/{id}", cbtSessionH.Get)
		r.With(requireCbtManage).Post("/api/cbt/sessions", cbtSessionH.Create)
		r.With(requireCbtManage).Patch("/api/cbt/sessions/{id}/status", cbtSessionH.UpdateStatus)
		r.With(requireCbtManage).Patch("/api/cbt/sessions/{id}/schedule", cbtSessionH.UpdateSchedule)
		r.With(requireCbtManage).Delete("/api/cbt/sessions/{id}", cbtSessionH.Delete)
		r.With(requireCbtManage).Post("/api/cbt/sessions/{id}/finalize-overdue", cbtSessionH.FinalizeOverdue)

		r.Group(func(r chi.Router) {
			r.Use(requireStaff)

			// Library — admin + staf
			r.Get("/api/library/stats", libraryH.Stats)
			r.Get("/api/library/books", libraryH.ListBooks)
			r.Post("/api/library/books", libraryH.CreateBook)
			r.Put("/api/library/books/{id}", libraryH.UpdateBook)
			r.Delete("/api/library/books/{id}", libraryH.DeleteBook)
			r.Get("/api/library/loans", libraryH.ListLoans)
			r.Post("/api/library/loans", libraryH.LoanBook)
			r.Post("/api/library/loans/{id}/return", libraryH.ReturnBook)
			r.Post("/api/library/loans/{id}/lunas", libraryH.MarkDendaLunas)

			// Inventory — admin + staf
			r.Get("/api/inventory/stats", inventoryH.Stats)
			r.Get("/api/inventory/rooms", inventoryH.ListSchoolRooms)
			r.Post("/api/inventory/rooms", inventoryH.CreateSchoolRoom)
			r.Get("/api/inventory/rooms/{id}", inventoryH.GetSchoolRoom)
			r.Put("/api/inventory/rooms/{id}", inventoryH.UpdateSchoolRoom)
			r.Delete("/api/inventory/rooms/{id}", inventoryH.DeleteSchoolRoom)
			r.Get("/api/inventory/items", inventoryH.ListItems)
			r.Post("/api/inventory/items", inventoryH.CreateItem)
			r.Patch("/api/inventory/items", inventoryH.BatchUpdateItems)
			r.Get("/api/inventory/items/{id}/history", inventoryH.ListItemEvents)
			r.Put("/api/inventory/items/{id}", inventoryH.UpdateItem)
			r.Delete("/api/inventory/items/{id}", inventoryH.DeleteItem)

			// Tata Usaha — admin + staf
			r.Get("/api/tu/surat/klasifikasi", letterH.ListClassifications)
			r.Get("/api/tu/surat/outgoing/preview-number", letterH.PreviewOutgoingNumber)
			r.Get("/api/tu/surat/incoming", letterH.ListIncoming)
			r.Post("/api/tu/surat/incoming", letterH.CreateIncoming)
			r.Get("/api/tu/surat/incoming/{id}", letterH.GetIncoming)
			r.Put("/api/tu/surat/incoming/{id}", letterH.UpdateIncoming)
			r.Patch("/api/tu/surat/incoming/{id}/status", letterH.UpdateIncomingStatus)
			r.Delete("/api/tu/surat/incoming/{id}", letterH.DeleteIncoming)
			r.Get("/api/tu/surat/outgoing", letterH.ListOutgoing)
			r.Post("/api/tu/surat/outgoing", letterH.CreateOutgoing)
			r.Get("/api/tu/surat/outgoing/{id}", letterH.GetOutgoing)
			r.Put("/api/tu/surat/outgoing/{id}", letterH.UpdateOutgoing)
			r.Delete("/api/tu/surat/outgoing/{id}", letterH.DeleteOutgoing)
			r.Get("/api/tu/surat/disposisi", letterH.ListDispositions)
			r.Post("/api/tu/surat/disposisi", letterH.CreateDisposition)
			r.Get("/api/tu/surat/disposisi/{id}", letterH.GetDisposition)
			r.Put("/api/tu/surat/disposisi/{id}", letterH.UpdateDisposition)
			r.Delete("/api/tu/surat/disposisi/{id}", letterH.DeleteDisposition)
			r.Get("/api/tu/surat-keterangan/templates", studentCertificateH.ListTemplates)
			r.Get("/api/tu/surat-keterangan/students", studentCertificateH.ListStudents)
			r.Get("/api/tu/surat-keterangan", studentCertificateH.List)
			r.Post("/api/tu/surat-keterangan", studentCertificateH.Create)
			r.Get("/api/tu/surat-keterangan/{id}", studentCertificateH.Get)
			r.Post("/api/tu/surat-keterangan/{id}/cancel", studentCertificateH.Cancel)
			r.Get("/api/tu/archives/stats", archiveH.Stats)
			r.Get("/api/tu/archives/categories", archiveH.ListCategories)
			r.Post("/api/tu/archives/categories", archiveH.CreateCategory)
			r.Put("/api/tu/archives/categories/{id}", archiveH.UpdateCategory)
			r.Delete("/api/tu/archives/categories/{id}", archiveH.DeleteCategory)
			r.Get("/api/tu/archives/documents", archiveH.ListDocuments)
			r.Post("/api/tu/archives/documents", archiveH.UploadDocument)
			r.Get("/api/tu/archives/documents/{id}", archiveH.GetDocument)
			r.Put("/api/tu/archives/documents/{id}", archiveH.UpdateDocument)
			r.Delete("/api/tu/archives/documents/{id}", archiveH.DeleteDocument)
			r.Get("/api/tu/archives/documents/{id}/file", archiveH.File)

			// Tata Kelola Madrasah — admin + staf
			r.Get("/api/governance/stats", governanceH.Stats)
			r.Get("/api/governance/snp-matrix", governanceH.SNPMatrix)
			r.Get("/api/governance/employee-options", governanceH.EmployeeOptions)
			r.Get("/api/governance/units", governanceH.ListUnits)
			r.Post("/api/governance/units", governanceH.CreateUnit)
			r.Put("/api/governance/units/{id}", governanceH.UpdateUnit)
			r.Delete("/api/governance/units/{id}", governanceH.DeleteUnit)
			r.Get("/api/governance/positions", governanceH.ListPositions)
			r.Post("/api/governance/positions", governanceH.CreatePosition)
			r.Put("/api/governance/positions/{id}", governanceH.UpdatePosition)
			r.Delete("/api/governance/positions/{id}", governanceH.DeletePosition)
			r.Get("/api/governance/assignments", governanceH.ListAssignments)
			r.Post("/api/governance/assignments", governanceH.CreateAssignment)
			r.Put("/api/governance/assignments/{id}", governanceH.UpdateAssignment)
			r.Delete("/api/governance/assignments/{id}", governanceH.DeleteAssignment)
			r.Get("/api/governance/documents", governanceH.ListDocuments)
			r.Post("/api/governance/documents", governanceH.CreateDocument)
			r.Put("/api/governance/documents/{id}", governanceH.UpdateDocument)
			r.Delete("/api/governance/documents/{id}", governanceH.DeleteDocument)
			r.Get("/api/governance/programs", governanceH.ListPrograms)
			r.Post("/api/governance/programs", governanceH.CreateProgram)
			r.Put("/api/governance/programs/{id}", governanceH.UpdateProgram)
			r.Delete("/api/governance/programs/{id}", governanceH.DeleteProgram)
			r.Get("/api/governance/work-plan-items", governanceH.ListWorkPlanItems)
			r.Post("/api/governance/work-plan-items", governanceH.CreateWorkPlanItem)
			r.Put("/api/governance/work-plan-items/{id}", governanceH.UpdateWorkPlanItem)
			r.Delete("/api/governance/work-plan-items/{id}", governanceH.DeleteWorkPlanItem)
			r.Get("/api/governance/performance-targets", governanceH.ListPerformanceTargets)
			r.Post("/api/governance/performance-targets", governanceH.CreatePerformanceTarget)
			r.Put("/api/governance/performance-targets/{id}", governanceH.UpdatePerformanceTarget)
			r.Delete("/api/governance/performance-targets/{id}", governanceH.DeletePerformanceTarget)
			r.Get("/api/governance/evidence-items", governanceH.ListEvidenceItems)
			r.Post("/api/governance/evidence-items", governanceH.CreateEvidenceItem)
			r.Put("/api/governance/evidence-items/{id}", governanceH.UpdateEvidenceItem)
			r.Delete("/api/governance/evidence-items/{id}", governanceH.DeleteEvidenceItem)
			r.Get("/api/governance/compliance-actions", governanceH.ListComplianceActions)
			r.Post("/api/governance/compliance-actions", governanceH.CreateComplianceAction)
			r.Put("/api/governance/compliance-actions/{id}", governanceH.UpdateComplianceAction)
			r.Delete("/api/governance/compliance-actions/{id}", governanceH.DeleteComplianceAction)

			// Siklus Dokumen — admin + staf
			r.Get("/api/document-cycles/stats", documentCycleH.Stats)
			r.Get("/api/document-cycles/catalogs", documentCycleH.ListCatalogs)
			r.Post("/api/document-cycles/catalogs", documentCycleH.CreateCatalog)
			r.Put("/api/document-cycles/catalogs/{id}", documentCycleH.UpdateCatalog)
			r.Delete("/api/document-cycles/catalogs/{id}", documentCycleH.DeleteCatalog)
			r.Get("/api/document-cycles/obligations", documentCycleH.ListObligations)
			r.Get("/api/document-cycles/verification-queue", documentCycleH.ListVerificationQueue)
			r.Post("/api/document-cycles/obligations/generate-year", documentCycleH.GenerateYear)
			r.Put("/api/document-cycles/obligations/{id}", documentCycleH.UpdateObligation)
			r.Get("/api/document-cycles/obligations/{id}/events", documentCycleH.ListEvents)
			r.Patch("/api/document-cycles/obligations/{id}/status", documentCycleH.UpdateObligationStatus)
			r.Delete("/api/document-cycles/obligations/{id}", documentCycleH.DeleteObligation)
		})

		// Kesiswaan — admin + kesiswaan manage; guru read scoped to own classes
		r.Get("/api/kesiswaan/stats", kesiswaanH.Stats)
		r.Get("/api/kesiswaan/classes", kesiswaanH.ClassOptions)
		r.Get("/api/kesiswaan/students", kesiswaanH.ListStudents)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/students/{id}/profile", kesiswaanH.UpdateStudentProfile)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/students/{id}/photo", kesiswaanH.UploadStudentPhoto)
		r.Get("/api/kesiswaan/student-photos/{filename}", kesiswaanH.StudentPhotoFile)
		r.Get("/api/kesiswaan/violation-categories", kesiswaanH.ListCategories)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/violation-categories", kesiswaanH.CreateCategory)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/violation-categories/{id}", kesiswaanH.UpdateCategory)
		r.With(requireKesiswaanManage).Delete("/api/kesiswaan/violation-categories/{id}", kesiswaanH.DeleteCategory)
		r.Get("/api/kesiswaan/violations", kesiswaanH.ListViolations)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/violations", kesiswaanH.CreateViolation)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/violations/{id}", kesiswaanH.UpdateViolation)
		r.With(requireKesiswaanManage).Delete("/api/kesiswaan/violations/{id}", kesiswaanH.DeleteViolation)
		r.Get("/api/kesiswaan/achievements", kesiswaanH.ListAchievements)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/achievements", kesiswaanH.CreateAchievement)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/achievements/{id}", kesiswaanH.UpdateAchievement)
		r.With(requireKesiswaanManage).Delete("/api/kesiswaan/achievements/{id}", kesiswaanH.DeleteAchievement)
		r.Get("/api/kesiswaan/extracurriculars", kesiswaanH.ListExtracurriculars)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/extracurriculars", kesiswaanH.CreateExtracurricular)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/extracurriculars/{id}", kesiswaanH.UpdateExtracurricular)
		r.With(requireKesiswaanManage).Delete("/api/kesiswaan/extracurriculars/{id}", kesiswaanH.DeleteExtracurricular)
		r.Get("/api/kesiswaan/extracurricular-members", kesiswaanH.ListExtracurricularMembers)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/extracurricular-members", kesiswaanH.CreateExtracurricularMember)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/extracurricular-members/{id}", kesiswaanH.UpdateExtracurricularMember)
		r.With(requireKesiswaanManage).Delete("/api/kesiswaan/extracurricular-members/{id}", kesiswaanH.DeleteExtracurricularMember)
		r.Get("/api/kesiswaan/counseling-sessions", kesiswaanH.ListCounselingSessions)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/counseling-sessions", kesiswaanH.CreateCounselingSession)
		r.With(requireKesiswaanManage).Put("/api/kesiswaan/counseling-sessions/{id}", kesiswaanH.UpdateCounselingSession)
		r.With(requireKesiswaanManage).Delete("/api/kesiswaan/counseling-sessions/{id}", kesiswaanH.DeleteCounselingSession)
		r.Get("/api/kesiswaan/student-transfers", kesiswaanH.ListStudentTransfers)
		r.With(requireKesiswaanManage).Post("/api/kesiswaan/student-transfers", kesiswaanH.CreateStudentTransfer)

		// Jobs / Attendance / Schedules / Settings — admin-only; Users/RBAC pilot use dynamic permissions.
		r.Group(func(r chi.Router) {
			r.Use(requirePusakaManage)

			r.Get("/api/pusaka/jobs", pusakaJobH.List)
			r.Post("/api/pusaka/jobs", pusakaJobH.Create)
			r.Get("/api/pusaka/jobs/stats", pusakaJobH.Stats)
			r.Post("/api/pusaka/jobs/run-all", pusakaJobH.RunAll)
			r.Post("/api/pusaka/jobs/cancel", pusakaJobH.CancelEmployee)
			r.Post("/api/pusaka/jobs/cancel-all", pusakaJobH.CancelAll)
			r.Post("/api/pusaka/jobs/sync-attendance", pusakaJobH.SyncAttendance)

			r.Get("/api/pusaka/attendance", pusakaAttendanceH.List)
			r.Get("/api/pusaka/attendance/summary", pusakaAttendanceH.GetSummary)
			r.Get("/api/pusaka/attendance/by-date/{date}", pusakaAttendanceH.ByDate)
			r.Get("/api/pusaka/attendance/by-employee/{id}", pusakaAttendanceH.ByEmployee)
			r.Get("/api/pusaka/attendance-telegram/settings", pusakaAttendanceTelegramH.GetSettings)
			r.Put("/api/pusaka/attendance-telegram/settings", pusakaAttendanceTelegramH.UpdateSettings)
			r.Get("/api/pusaka/attendance-telegram/logs", pusakaAttendanceTelegramH.ListLogs)
			r.Post("/api/pusaka/attendance-telegram/send", pusakaAttendanceTelegramH.SendNow)
			r.Post("/api/pusaka/attendance-telegram/tick", pusakaAttendanceTelegramH.Tick)

			r.Get("/api/pusaka/schedules", pusakaScheduleH.List)
			r.Post("/api/pusaka/schedules", pusakaScheduleH.Create)
			r.Put("/api/pusaka/schedules/{id}", pusakaScheduleH.Update)
			r.Delete("/api/pusaka/schedules/{id}", pusakaScheduleH.Delete)

			r.Get("/api/pusaka/settings", settH.List)
			r.Put("/api/pusaka/settings/{key}", settH.Upsert)

			r.Get("/api/pusaka/employees/{id}/pusaka-status", empH.GetPusakaStatus)
			r.Post("/api/pusaka/employees/{id}/update-pusaka", empH.UpdatePusakaCredentials)
			r.Get("/api/pusaka/employees/{id}/schedules", empSchedH.List)
			r.Post("/api/pusaka/employees/{id}/schedules", empSchedH.Upsert)
			r.Delete("/api/pusaka/employees/{id}/schedules/{scheduleId}", empSchedH.Delete)

			r.Post("/api/pusaka/scheduler/tick", pusakaSchedulerH.Tick)
			r.Get("/api/pusaka/worker/status", pusakaWorkerH.GetStatus)
		})

		r.With(requireUsersRead).Get("/api/users", userH.List)
		r.With(requireUsersCreate).Get("/api/users/profile-candidates", userH.ListProfileCandidates)
		r.With(requireProfileChangesReview).Get("/api/users/change-requests", profileChangeRequestH.ListAdmin)
		r.With(requireProfileChangesReview).Get("/api/users/change-requests/pending-count", profileChangeRequestH.CountPendingAdmin)
		r.With(requireProfileChangesReview).Get("/api/users/change-requests/export", profileChangeRequestH.ExportAdminCSV)
		r.With(requireProfileChangesReview).Patch("/api/users/change-requests/{id}", profileChangeRequestH.Review)
		r.With(requireAuditRead).Get("/api/users/audit-logs", userH.ListAuditLogs)
		r.With(requireUsersCreate).Get("/api/users/generate-from-employees/preview", userH.PreviewEmployeeAccountGeneration)
		r.With(requireUsersCreate).Post("/api/users/generate-from-employees", userH.GenerateEmployeeAccounts)
		r.With(requireStudentAccountsManage).Get("/api/users/student-accounts/preview", userH.PreviewStudentAccounts)
		r.With(requireStudentAccountsManage).Post("/api/users/student-accounts/generate", userH.GenerateStudentAccounts)
		r.With(requireParentAccountsManage).Get("/api/users/parent-accounts/preview", userH.PreviewParentAccounts)
		r.With(requireParentAccountsManage).Post("/api/users/parent-accounts/generate", userH.GenerateParentAccounts)
		r.With(requireUsersCreate).Post("/api/users", userH.Create)
		r.With(requireUsersDeactivate).Patch("/api/users/{id}/status", userH.UpdateStatus)
		r.With(requireUsersResetPassword).Post("/api/users/{id}/reset-password", userH.ResetPassword)
		r.With(requireUsersResetPasswordOrAdmin).Post("/api/users/{id}/force-password-change", userH.ForcePasswordChange)
		r.With(requireUsersUpdate).Patch("/api/users/{id}/profile-link", userH.UpdateProfileLink)
		r.With(requireUsersManageRoles).Patch("/api/users/{id}/roles", rbacH.UpdateUserRoles)
		r.With(requireUsersDeactivate).Delete("/api/users/{id}", userH.Delete)

		r.With(requireRolesRead).Get("/api/rbac/roles", rbacH.ListRoles)
		r.With(requireRolesManage).Post("/api/rbac/roles", rbacH.CreateRole)
		r.With(requireRolesManage).Put("/api/rbac/roles/{code}", rbacH.UpdateRole)
		r.With(requireRolesManage).Patch("/api/rbac/roles/{code}/status", rbacH.SetRoleStatus)
		r.With(requireRolesRead).Get("/api/rbac/permissions", rbacH.ListPermissions)
		r.With(requireRolesManage).Post("/api/rbac/permissions", rbacH.CreatePermission)
		r.With(requireRolesManage).Put("/api/rbac/permissions/{code}", rbacH.UpdatePermission)
		r.With(requireRolesManage).Patch("/api/rbac/permissions/{code}/status", rbacH.SetPermissionStatus)
		r.With(requireRolesRead).Get("/api/rbac/matrix", rbacH.ListMatrix)
		r.With(requireRolesManage).Put("/api/rbac/roles/{code}/permissions", rbacH.UpdateRolePermissions)
	})

	r.Group(func(r chi.Router) {
		r.Use(mw.WorkerKey(workerKey))
		r.Post("/api/pusaka/worker/claim", pusakaWorkerH.Claim)
		r.Get("/api/pusaka/worker/config", pusakaWorkerH.Config)
		r.Post("/api/pusaka/worker/heartbeat", pusakaWorkerH.Heartbeat)
		r.Post("/api/pusaka/worker/jobs/{id}/complete", pusakaWorkerH.Complete)
		r.Post("/api/pusaka/worker/jobs/{id}/fail", pusakaWorkerH.Fail)
		r.Post("/api/pusaka/worker/attendance", pusakaWorkerH.UpsertAttendance)
	})

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-mainCtx.Done()
	slog.Info("shutting down...")

	internalAnalyticsRollupCancel()
	pusakaAttendanceTelegramSvc.Stop()
	pusakaSchedulerSvc.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("shutdown complete")
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env", "key", key)
		os.Exit(1)
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durationEnv(key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	if parsed, err := time.ParseDuration(raw); err == nil {
		return parsed
	}
	if minutes, err := strconv.Atoi(raw); err == nil {
		return time.Duration(minutes) * time.Minute
	}
	slog.Warn("invalid duration env, using default", "key", key, "value", raw)
	return def
}

func int32Env(key string, def int32) int32 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		slog.Warn("invalid int env, using default", "key", key, "value", raw)
		return def
	}
	return int32(parsed)
}

func boolEnv(key string, def bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		slog.Warn("invalid bool env, using default", "key", key, "value", raw)
		return def
	}
	return parsed
}

func splitCSVEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}
