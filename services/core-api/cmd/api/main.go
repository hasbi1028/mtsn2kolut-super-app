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
	tempSvc := service.NewEmployee(q)
	pusakaJobSvc := service.NewPusakaJobWithPool(pool)
	pusakaAttendanceSvc := service.NewPusakaAttendance(q)
	pusakaAttendanceTelegramSvc := service.NewPusakaAttendanceTelegram(q, getEnv("TELEGRAM_BOT_TOKEN", ""))
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
	rbacSvc := service.NewRBACWithPool(pool)
	profileChangeRequestSvc := service.NewProfileChangeRequestWithPool(pool)
	academicSvc := service.NewSemesterService(q, pool)
	curriculumSvc := service.NewCurriculumService(q)
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
	tempH := handler.NewEmployee(tempSvc)
	healthH := handler.NewHealth(pool, pusakaJobSvc, settSvc, internalAnalyticsSvc)
	pusakaJobH := handler.NewPusakaJob(pusakaJobSvc)
	pusakaAttendanceH := handler.NewPusakaAttendance(pusakaAttendanceSvc)
	pusakaAttendanceTelegramH := handler.NewPusakaAttendanceTelegram(pusakaAttendanceTelegramSvc)
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
	rbacH := handler.NewRBAC(rbacSvc)
	profileChangeRequestH := handler.NewProfileChangeRequest(profileChangeRequestSvc)
	academicH := handler.NewAcademicHandler(academicSvc)
	curriculumH := handler.NewCurriculumHandler(curriculumSvc, academicSvc)
	assignSvc := service.NewSubjectAssignmentService(q)
	assignH := handler.NewSubjectAssignmentHandler(assignSvc)

	kesiswaanSvc := service.NewKesiswaanService(q)
	kesiswaanH := handler.NewKesiswaanHandler(kesiswaanSvc)

	timetableSvc := service.NewTimetableService(q)
	timetableH := handler.NewTimetableHandler(timetableSvc)

	journalSvc := service.NewClassJournalService(q)
	journalH := handler.NewClassJournal(journalSvc)

	jwtSecret := mustEnv("JWT_SECRET")
	workerKey := mustEnv("WORKER_API_KEY")
	internalAPIKey := getEnv("INTERNAL_API_KEY", "")
	trustedProxies := splitCSVEnv("TRUSTED_PROXY_CIDRS")
	authRateLimit := ratelimit.RateLimitWithTrustedProxies(5, 1, trustedProxies)
	refreshRateLimit := ratelimit.RateLimitWithTrustedProxies(10, 1, trustedProxies)
	logoutRateLimit := ratelimit.RateLimitWithTrustedProxies(10, 2, trustedProxies)
	passwordRateLimit := ratelimit.RateLimitWithTrustedProxies(3, 0.1, trustedProxies)
	publicSiteRateLimit := ratelimit.RateLimitWithTrustedProxies(60, 30, trustedProxies)
	publicAnalyticsRateLimit := ratelimit.RateLimitWithTrustedProxies(20, 5, trustedProxies)
	analyticsIngestionRateLimit := ratelimit.RateLimitWithTrustedProxies(30, 10, trustedProxies)

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
	r.With(publicSiteRateLimit).Get("/api/public/branding", brandingH.Public)
	r.Get("/api/branding/file/{filename}", brandingH.Asset)
	r.With(publicAnalyticsRateLimit, mw.InternalKey(internalAPIKey)).Post("/api/internal-analytics/public-events", internalAnalyticsH.CreatePublicEvent)
	r.Get("/api/system/maintenance/status", systemMaintenanceH.Status)

	requireAdmin := mw.RequireAdmin()
	requirePusakaRead := mw.RequireAnyPermissionOrRole([]string{"pusaka.read", "pusaka.manage", "pusaka.sync", "pusaka.credentials_manage"}, "admin")
	requirePusakaManage := mw.RequireAnyPermissionOrRole([]string{"pusaka.manage", "pusaka.sync", "pusaka.credentials_manage"}, "admin")
	requireUsersRead := mw.RequirePermission("users.read")
	requireUsersCreate := mw.RequirePermission("users.create")
	requireUsersDeactivate := mw.RequirePermission("users.deactivate")
	requireUsersResetPassword := mw.RequirePermission("users.reset_password")
	requireUsersResetPasswordOrAdmin := mw.RequireAnyPermissionOrRole([]string{"users.reset_password"}, "admin")
	requireUsersUpdate := mw.RequirePermission("users.update")
	requireUsersManageRoles := mw.RequirePermission("users.manage_roles")
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
			r.With(requireEmployeesRead).Get("/api/employees", tempH.List)
			r.With(requireEmployeesManage).Post("/api/employees", tempH.Create)
			r.With(requireEmployeesRead).Get("/api/employees/{id}", tempH.Get)
			r.With(requireEmployeesManage).Put("/api/employees/{id}", tempH.Update)
			r.With(requireEmployeesManage).Patch("/api/employees/{id}/status", tempH.UpdateStatus)
			r.With(requireEmployeesManage).Delete("/api/employees/{id}", tempH.Delete)
			r.With(requirePusakaRead).Get("/api/pusaka/employees", tempH.ListPusakaEligibleWithStatus)
			r.With(requirePusakaManage).Patch("/api/pusaka/employees/{id}/account-status", tempH.UpdatePusakaAccountStatus)
			r.With(requirePusakaManage).Delete("/api/pusaka/employees/{id}/account", tempH.DeletePusakaAccount)
			r.With(requirePusakaRead).Get("/api/pusaka/employees/{id}/audit-logs", tempH.ListPusakaAuditLogs)
			r.With(requirePusakaRead).Get("/api/pusaka/employees/{id}/schedules", empSchedH.List)
			r.With(requirePusakaManage).Post("/api/pusaka/employees/{id}/schedules", empSchedH.Upsert)
			r.With(requirePusakaManage).Delete("/api/pusaka/employees/{id}/schedules/{scheduleId}", empSchedH.Delete)
		})

		// Academic — master data semester
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireAnyPermissionOrRole([]string{"academic.read", "academic.manage"}, "admin"))
			r.Get("/api/academic/semesters", academicH.ListSemesters)
			r.Get("/api/academic/semesters/active", academicH.GetActiveSemester)
			r.Get("/api/academic/semesters/{id}", academicH.GetSemester)
			r.Post("/api/academic/semesters", academicH.CreateSemester)
			r.Post("/api/academic/semesters/{id}/activate", academicH.ActivateSemester)
			r.Delete("/api/academic/semesters/{id}", academicH.DeleteSemester)
			r.Get("/api/academic/rombels", academicH.ListSchoolClasses)
			r.Get("/api/academic/rombels/unassigned-students", academicH.ListUnassignedStudents)
			r.Post("/api/academic/rombels", academicH.CreateSchoolClass)
			r.Get("/api/academic/rombels/{id}", academicH.GetSchoolClass)
			r.Get("/api/academic/rombels/{id}/students", academicH.ListRombelStudents)
			r.Post("/api/academic/rombels/{id}/students", academicH.AssignStudent)
			r.Delete("/api/academic/rombels/{id}/students/{studentId}", academicH.RemoveStudentFromClass)
			r.Delete("/api/academic/rombels/{id}", academicH.DeleteSchoolClass)
			r.Get("/api/academic/rombels/{id}/homeroom", academicH.GetHomeroom)
			r.Post("/api/academic/rombels/{id}/homeroom", academicH.SetHomeroom)
			r.Get("/api/academic/curriculum/profiles", curriculumH.ListProfiles)
			r.Get("/api/academic/curriculum/profiles/active", curriculumH.GetActiveProfile)
			r.Post("/api/academic/curriculum/profiles", curriculumH.CreateProfile)
			r.Post("/api/academic/curriculum/profiles/{id}/activate", curriculumH.ActivateProfile)
			r.Delete("/api/academic/curriculum/profiles/{id}", curriculumH.DeleteProfile)
			r.Put("/api/academic/curriculum/profiles/{id}", curriculumH.UpdateProfile)
			r.Get("/api/academic/curriculum/profiles/{id}/allocations", curriculumH.ListAllocations)
			r.Post("/api/academic/curriculum/profiles/{id}/allocations", curriculumH.CreateAllocation)
			r.Delete("/api/academic/curriculum/allocations/{id}", curriculumH.DeleteAllocation)
			r.Put("/api/academic/curriculum/allocations/{id}", curriculumH.UpdateAllocation)
			r.Get("/api/academic/curriculum/assignments", curriculumH.ListAssignments)
			r.Post("/api/academic/curriculum/assignments", curriculumH.CreateAssignment)
			r.Delete("/api/academic/curriculum/assignments/{id}", curriculumH.DeleteAssignment)
		})

		// Subject Assignments — Assign Guru ke Mapel per Rombel
		r.Get("/api/academic/subject-assignments", assignH.GetMatrix)
		r.Post("/api/academic/subject-assignments", assignH.UpsertCell)
		r.Delete("/api/academic/subject-assignments/{id}", assignH.DeleteCell)

		// Kesiswaan — data murid
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireAnyPermissionOrRole([]string{"kesiswaan.read", "kesiswaan.manage"}, "admin"))
			r.Get("/api/kesiswaan/murid", kesiswaanH.ListMurid)
			r.Put("/api/kesiswaan/murid/{id}/profile", kesiswaanH.UpdateMuridProfile)
			r.Post("/api/kesiswaan/murid", kesiswaanH.CreateMurid)
			r.Delete("/api/kesiswaan/murid/{id}", kesiswaanH.DeleteMurid)
		})

		// Timetable — jadwal pelajaran
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireAnyPermissionOrRole([]string{"academic.read", "academic.manage"}, "admin"))
			r.Get("/api/academic/timetable/weekly", timetableH.GetWeeklyData)
			r.Post("/api/academic/timetable/slots", timetableH.CreateSlot)
			r.Put("/api/academic/timetable/slots/{id}", timetableH.UpdateSlot)
			r.Delete("/api/academic/timetable/slots/{id}", timetableH.DeleteSlot)
		})

		// Class Journal — jurnal belajar harian
		r.Group(func(r chi.Router) {
			r.Use(mw.RequireAnyPermissionOrRole([]string{"journal.read", "journal.manage"}, "admin"))
			r.Get("/api/class-journal", journalH.Overview)
			r.Post("/api/class-journal/sessions", journalH.CreateSession)
			r.Get("/api/class-journal/sessions/{id}/attendances", journalH.ListAttendances)
			r.Put("/api/class-journal/sessions/{id}/attendances", journalH.BulkUpsertAttendances)
			r.Get("/api/class-journal/sessions/{id}", journalH.GetSession)
			r.Delete("/api/class-journal/sessions/{id}", journalH.DeleteSession)
			r.Get("/api/class-journal/summary", journalH.AttendanceSummary)
			r.Post("/api/academic/rombel/{id}/timetable-slots/{slotID}/journal-session", journalH.OpenSessionFromTimetableSlot)
		})

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

			r.Get("/api/pusaka/employees/{id}/pusaka-status", tempH.GetPusakaStatus)
			r.Post("/api/pusaka/employees/{id}/update-pusaka", tempH.UpdatePusakaCredentials)
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
