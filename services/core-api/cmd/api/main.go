package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	authSvc := service.NewAuth(q, mustEnv("JWT_SECRET"), getEnv("ADMIN_PASSWORD", ""))
	academicSvc := service.NewAcademic(q)
	gradeSvc := service.NewGrade(q)
	empSvc := service.NewEmployee(q)
	studentSvc := service.NewStudent(q)
	parentSvc := service.NewParent(q)
	portalSvc := service.NewPortal(q)
	websiteSvc := service.NewWebsite(q)
	pusakaJobSvc := service.NewPusakaJob(q)
	pusakaAttendanceSvc := service.NewPusakaAttendance(q)
	questionSvc := service.NewCbtQuestion(q)
	questionAssetSvc := service.NewCbtQuestionAsset(q, getEnv("CBT_ASSET_DIR", "data/cbt-assets"))
	packageSvc := service.NewCbtPackage(pool)
	sessionSvc := service.NewCbtSession(pool)
	eventSvc := service.NewCbtEvent(pool)
	examSvc := service.NewExam(pool)
	pusakaScheduleSvc := service.NewPusakaSchedule(q)
	empSchedSvc := service.NewEmployeeSchedule(q)
	settSvc := service.NewSetting(q)
	auditSvc := service.NewAudit(q)
	pusakaSchedulerSvc := service.NewPusakaScheduler(q, pusakaJobSvc, settSvc, auditSvc)
	notificationSvc := service.NewNotification(q)
	librarySvc := service.NewLibrary(q)
	inventorySvc := service.NewInventory(q)
	journalSvc := service.NewClassJournal(q)
	letterSvc := service.NewLetter(q)
	governanceSvc := service.NewGovernance(q)
	documentCycleSvc := service.NewDocumentCycle(q)
	kesiswaanSvc := service.NewKesiswaanWithPool(pool, getEnv("STUDENT_PHOTO_DIR", "data/student-photos"))
	studentCertificateSvc := service.NewStudentCertificate(pool)
	archiveSvc := service.NewArchive(q, getEnv("ARCHIVE_DIR", "data/archives"))
	websiteMediaH := handler.NewWebsiteMedia(getEnv("WEBSITE_MEDIA_DIR", "data/website-media"))

	if err := authSvc.SeedAdmin(mainCtx); err != nil {
		slog.Error("seed admin", "error", err)
		os.Exit(1)
	}
	if err := settSvc.SeedDefaults(mainCtx); err != nil {
		slog.Error("seed settings", "error", err)
		os.Exit(1)
	}
	pusakaSchedulerSvc.Start(mainCtx)

	authH := handler.NewAuth(authSvc, q)
	academicH := handler.NewAcademic(academicSvc)
	gradeH := handler.NewGrade(gradeSvc, q)
	empH := handler.NewEmployee(empSvc)
	healthH := handler.NewHealth(pool, pusakaJobSvc, settSvc)
	pusakaJobH := handler.NewPusakaJob(pusakaJobSvc)
	pusakaAttendanceH := handler.NewPusakaAttendance(pusakaAttendanceSvc)
	studentH := handler.NewStudent(studentSvc)
	parentH := handler.NewParent(parentSvc)
	portalH := handler.NewPortal(portalSvc)
	websiteH := handler.NewWebsite(websiteSvc)
	questionH := handler.NewCbtQuestion(questionSvc, q)
	questionAssetH := handler.NewCbtQuestionAsset(questionAssetSvc)
	packageH := handler.NewCbtPackage(packageSvc, q)
	sessionH := handler.NewCbtSession(sessionSvc, q)
	eventH := handler.NewCbtEvent(eventSvc)
	examH := handler.NewExam(examSvc)
	userH := handler.NewUser(q)
	pusakaScheduleH := handler.NewPusakaSchedule(pusakaScheduleSvc)
	empSchedH := handler.NewEmployeeSchedule(empSchedSvc)
	settH := handler.NewSetting(settSvc)
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
	studentCertificateH := handler.NewStudentCertificate(studentCertificateSvc)
	archiveH := handler.NewArchive(archiveSvc, q)

	jwtSecret := mustEnv("JWT_SECRET")
	workerKey := mustEnv("WORKER_API_KEY")
	examTokenMW := mw.ExamToken(examSvc.GetParticipantByToken)
	authRateLimit := ratelimit.RateLimit(5, 1)
	refreshRateLimit := ratelimit.RateLimit(10, 1)
	publicRegisterRateLimit := ratelimit.RateLimit(3, 0.2)
	examLoginRateLimit := ratelimit.RateLimit(8, 1)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(mw.RequestLog)
	r.Use(chimw.Recoverer)
	r.Use(chimw.SetHeader("Content-Type", "application/json"))

	r.Get("/health", healthH.Get)

	r.With(authRateLimit).Post("/api/auth/login", authH.Login)
	r.With(refreshRateLimit).Post("/api/auth/refresh", authH.Refresh)
	r.Post("/api/auth/logout", authH.Logout)
	r.With(publicRegisterRateLimit).Post("/api/public/register-student", studentH.PublicRegister)
	r.Get("/api/public/site/posts", websiteH.ListPublishedPosts)
	r.Get("/api/public/site/posts/featured", websiteH.ListFeaturedPosts)
	r.Get("/api/public/site/posts/{slug}", websiteH.GetPublishedPost)
	r.Get("/api/public/site/announcements", websiteH.ListPublishedAnnouncements)
	r.Get("/api/public/site/announcements/{slug}", websiteH.GetPublishedAnnouncement)
	r.Get("/api/public/site/pages/{slug}", websiteH.GetPublishedPage)
	r.Get("/api/website/media/{filename}", websiteMediaH.File)

	// CBT asset files are not public-by-obscurity. They may be accessed either by
	// authenticated admin/guru requests or by active exam participants using the
	// exam token attached to exam payload asset URLs.
	r.With(mw.ExamTokenOrJWT(jwtSecret, authSvc.CurrentAuthVersion, authSvc.ValidateAccessSession, examSvc.GetParticipantByToken)).Get("/api/cbt/assets/{id}/file", questionAssetH.File)

	// Exam endpoints — authenticated via X-Exam-Token (no JWT needed)
	r.With(examLoginRateLimit).Post("/api/exam/login", examH.Login)
	r.Group(func(r chi.Router) {
		r.Use(examTokenMW)
		r.Get("/api/exam/status", examH.Status)
		r.Post("/api/exam/heartbeat", examH.Heartbeat)
		r.Post("/api/exam/event", examH.RecordEvent)
		r.Post("/api/exam/answer", examH.SubmitAnswer)
		r.Post("/api/exam/submit", examH.Submit)
	})

	requireAdmin := mw.RequireAdmin()
	requireCbt := mw.RequireAnyRole("admin", "guru")
	requireStaff := mw.RequireAnyRole("admin", "staf")
	requireKesiswaanManage := mw.RequireAnyRole("admin", "kesiswaan")

	r.Group(func(r chi.Router) {
		r.Use(mw.JWT(jwtSecret, authSvc.CurrentAuthVersion, authSvc.ValidateAccessSession))
		r.Use(mw.Audit(q))
		r.Post("/api/auth/change-password", authH.ChangePassword)
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
		r.Get("/api/school-profile", settH.SchoolProfile)
		r.With(requireAdmin).Put("/api/school-profile", settH.UpdateSchoolProfile)

		// Employees are admin-only
		r.Group(func(r chi.Router) {
			r.Use(requireAdmin)
			r.Get("/api/employees", empH.List)
			r.Post("/api/employees", empH.Create)
			r.Get("/api/employees/{id}", empH.Get)
			r.Put("/api/employees/{id}", empH.Update)
			r.Patch("/api/employees/{id}/status", empH.UpdateStatus)
			r.Delete("/api/employees/{id}", empH.Delete)
			r.Get("/api/pusaka/employees", empH.ListPusakaEligibleWithStatus)
			r.Patch("/api/pusaka/employees/{id}/account-status", empH.UpdatePusakaAccountStatus)
			r.Delete("/api/pusaka/employees/{id}/account", empH.DeletePusakaAccount)
			r.Get("/api/pusaka/employees/{id}/audit-logs", empH.ListPusakaAuditLogs)
			r.Get("/api/pusaka/employees/{id}/schedules", empSchedH.List)
			r.Post("/api/pusaka/employees/{id}/schedules", empSchedH.Upsert)
			r.Delete("/api/pusaka/employees/{id}/schedules/{scheduleId}", empSchedH.Delete)
		})

		r.Get("/api/academic", academicH.Overview)
		r.Get("/api/academic/stats", academicH.GetStats)
		r.With(requireAdmin).Post("/api/academic/{entity}", academicH.Create)
		r.With(requireAdmin).Put("/api/academic/{entity}/{id}", academicH.Update)
		r.With(requireAdmin).Delete("/api/academic/{entity}/{id}", academicH.Delete)
		r.Get("/api/grades", gradeH.Overview)
		r.Post("/api/grades/components", gradeH.CreateComponent)
		r.Post("/api/grades/assignments/{id}/finalize", gradeH.FinalizeAssignment)
		r.Delete("/api/grades/assignments/{id}/finalize", gradeH.ReopenAssignment)
		r.Put("/api/grades/components/{id}", gradeH.UpdateComponent)
		r.Patch("/api/grades/components/{id}/publish", gradeH.SetComponentPublished)
		r.Delete("/api/grades/components/{id}", gradeH.DeleteComponent)
		r.Post("/api/grades/components/{id}/entries", gradeH.UpsertEntry)

		r.Get("/api/journal", journalH.Overview)
		r.Post("/api/journal/sessions", journalH.CreateSession)
		r.Get("/api/journal/sessions/{id}", journalH.GetSession)
		r.Put("/api/journal/sessions/{id}", journalH.UpdateSession)
		r.With(requireAdmin).Delete("/api/journal/sessions/{id}", journalH.DeleteSession)
		r.Post("/api/journal/sessions/{id}/attendances", journalH.BulkUpsertAttendances)

		r.Get("/api/students", studentH.GuruAwareList)
		r.With(requireAdmin).Post("/api/students", studentH.Create)
		r.With(requireAdmin).Put("/api/students/{id}", studentH.Update)
		r.With(requireAdmin).Patch("/api/students/{id}/lifecycle", studentH.UpdateLifecycle)
		r.With(requireAdmin).Delete("/api/students/{id}", studentH.Delete)

		r.Get("/api/parents", parentH.List)
		r.With(requireAdmin).Post("/api/parents", parentH.Create)
		r.Get("/api/parents/{id}", parentH.Get)
		r.With(requireAdmin).Put("/api/parents/{id}", parentH.Update)
		r.With(requireAdmin).Delete("/api/parents/{id}", parentH.Delete)
		r.Get("/api/parents/{id}/children", parentH.ListChildren)
		r.With(requireAdmin).Post("/api/parents/{id}/link", parentH.LinkStudent)
		r.With(requireAdmin).Post("/api/parents/{id}/unlink", parentH.UnlinkStudent)
		r.Get("/api/portal/student/me", portalH.StudentMe)
		r.Get("/api/portal/guru/timetable", portalH.TeacherTimetable)
		r.Get("/api/portal/parent/me", portalH.ParentMe)
		r.With(requireAdmin).Get("/api/website/content", websiteH.List)
		r.With(requireAdmin).Post("/api/website/content", websiteH.Create)
		r.With(requireAdmin).Put("/api/website/content/{id}", websiteH.Update)
		r.With(requireAdmin).Delete("/api/website/content/{id}", websiteH.Delete)
		r.With(requireAdmin).Post("/api/website/media", websiteMediaH.Upload)

		r.With(requireCbt).Get("/api/cbt/questions", questionH.List)
		r.With(requireCbt).Post("/api/cbt/questions", questionH.Create)
		r.With(requireCbt).Get("/api/cbt/questions/export", questionH.ExportCSV)
		r.With(requireCbt).Get("/api/cbt/questions/template", questionH.TemplateCSV)
		r.With(requireCbt).Post("/api/cbt/questions/import-legacy", questionH.ImportLegacyCSV)
		r.With(requireCbt).Get("/api/cbt/questions/{id}", questionH.Get)
		r.With(requireCbt).Put("/api/cbt/questions/{id}", questionH.Update)
		r.With(requireCbt).Post("/api/cbt/questions/{id}/duplicate", questionH.Duplicate)
		r.With(requireCbt).Patch("/api/cbt/questions/{id}/workflow", questionH.WorkflowAction)
		r.With(requireCbt).Delete("/api/cbt/questions/{id}", questionH.Delete)
		r.With(requireCbt).Get("/api/cbt/assets", questionAssetH.List)
		r.With(requireCbt).Post("/api/cbt/assets", questionAssetH.Upload)

		r.With(requireCbt).Get("/api/cbt/packages", packageH.List)
		r.With(requireAdmin).Post("/api/cbt/packages", packageH.Create)
		r.With(requireAdmin).Delete("/api/cbt/packages/{id}", packageH.Delete)

		// CBT Events (kegiatan ujian) — admin manages, guru reads
		r.With(requireCbt).Get("/api/cbt/events", eventH.List)
		r.With(requireAdmin).Post("/api/cbt/events", eventH.Create)
		r.With(requireCbt).Get("/api/cbt/events/{id}", eventH.Get)
		r.With(requireCbt).Get("/api/cbt/events/{id}/results", eventH.GetResults)
		r.With(requireCbt).Get("/api/cbt/events/{id}/exam-cards", eventH.GetExamCards)
		r.With(requireAdmin).Put("/api/cbt/events/{id}", eventH.Update)
		r.With(requireAdmin).Patch("/api/cbt/events/{id}/status", eventH.UpdateStatus)
		r.With(requireAdmin).Delete("/api/cbt/events/{id}", eventH.Delete)

		// CBT Sessions
		r.With(requireCbt).Get("/api/cbt/sessions", sessionH.GuruAwareList)
		r.With(requireAdmin).Post("/api/cbt/sessions", sessionH.Create)
		r.With(requireCbt).Get("/api/cbt/sessions/{id}", sessionH.Get)
		r.With(requireAdmin).Patch("/api/cbt/sessions/{id}/status", sessionH.UpdateStatus)
		r.With(requireAdmin).Delete("/api/cbt/sessions/{id}", sessionH.Delete)

		// Participants & Enrollment
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/participants", sessionH.GuruAwareParticipants)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/enroll", sessionH.Enroll)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/enroll-grade", sessionH.EnrollGrade)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/enroll-school", sessionH.EnrollSchool)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/generate-tokens", sessionH.GenerateTokens)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/participants/{pid}/regenerate-token", sessionH.RegenerateToken)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/participants/{pid}/reset-access", sessionH.ResetParticipantAccess)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/participants/{pid}/seat", sessionH.AssignSeat)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/seats/auto", sessionH.AutoAssignSeats)

		// Rooms & Shuffle
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/rooms", sessionH.ListRooms)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/rooms", sessionH.CreateRoom)
		r.With(requireCbt).Delete("/api/cbt/sessions/{id}/rooms/{rid}", sessionH.DeleteRoom)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/shuffle-rooms", sessionH.ShuffleRooms)
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/minutes", sessionH.GetMinutes)

		// Scoring & Results
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/score", sessionH.ScoreSession)
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/results", sessionH.GuruAwareResults)
		r.With(requireAdmin).Post("/api/cbt/sessions/{id}/participants/{pid}/answer", sessionH.RecordAnswer)
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/participants/{pid}/answers", sessionH.GetParticipantAnswers)

		// Proctoring
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/proctoring", sessionH.GetProctoringStatus)
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/proctoring/events", sessionH.ListParticipantEvents)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/participants/{pid}/flag", sessionH.FlagParticipant)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/participants/{pid}/force-submit", sessionH.ForceSubmitParticipant)

		// Essay Grading
		r.With(requireCbt).Get("/api/cbt/sessions/{id}/ungraded-essays", sessionH.ListUngradedEssays)
		r.With(requireCbt).Post("/api/cbt/sessions/{id}/answers/{aid}/grade-essay", sessionH.GradeEssay)

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

		// Jobs / Attendance / Schedules / Settings / Users — admin-only
		r.Group(func(r chi.Router) {
			r.Use(requireAdmin)

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

			r.Get("/api/users", userH.List)
			r.Get("/api/users/audit-logs", userH.ListAuditLogs)
			r.Post("/api/users", userH.Create)
			r.Patch("/api/users/{id}/status", userH.UpdateStatus)
			r.Delete("/api/users/{id}", userH.Delete)

		})
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
