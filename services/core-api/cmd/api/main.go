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
	librarySvc := service.NewLibrary(q)

	if err := authSvc.SeedAdmin(mainCtx); err != nil {
		slog.Error("seed admin", "error", err)
		os.Exit(1)
	}
	if err := settSvc.SeedDefaults(mainCtx); err != nil {
		slog.Error("seed settings", "error", err)
		os.Exit(1)
	}
	pusakaSchedulerSvc.Start(mainCtx)

	authH := handler.NewAuth(authSvc)
	academicH := handler.NewAcademic(academicSvc)
	gradeH := handler.NewGrade(gradeSvc)
	empH := handler.NewEmployee(empSvc)
	healthH := handler.NewHealth(pool, pusakaJobSvc, settSvc)
	pusakaJobH := handler.NewPusakaJob(pusakaJobSvc)
	pusakaAttendanceH := handler.NewPusakaAttendance(pusakaAttendanceSvc)
	studentH := handler.NewStudent(studentSvc)
	parentH := handler.NewParent(parentSvc)
	portalH := handler.NewPortal(portalSvc)
	questionH := handler.NewCbtQuestion(questionSvc)
	questionAssetH := handler.NewCbtQuestionAsset(questionAssetSvc)
	packageH := handler.NewCbtPackage(packageSvc)
	sessionH := handler.NewCbtSession(sessionSvc)
	eventH := handler.NewCbtEvent(eventSvc)
	examH := handler.NewExam(examSvc)
	userH := handler.NewUser(q)
	pusakaScheduleH := handler.NewPusakaSchedule(pusakaScheduleSvc)
	empSchedH := handler.NewEmployeeSchedule(empSchedSvc)
	settH := handler.NewSetting(settSvc)
	pusakaSchedulerH := handler.NewPusakaScheduler(pusakaSchedulerSvc)
	pusakaWorkerH := handler.NewPusakaWorker(pusakaJobSvc, pusakaAttendanceSvc, settSvc)
	libraryH := handler.NewLibrary(librarySvc)

	jwtSecret := mustEnv("JWT_SECRET")
	workerKey := mustEnv("WORKER_API_KEY")
	internalKey := getEnv("INTERNAL_API_KEY", "")

	examTokenMW := mw.ExamToken(examSvc.GetParticipantByToken)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(mw.RequestLog)
	r.Use(chimw.Recoverer)
	r.Use(chimw.SetHeader("Content-Type", "application/json"))

	r.Get("/health", healthH.Get)

	r.Post("/api/auth/login", authH.Login)
	r.Post("/api/auth/refresh", authH.Refresh)
	r.Post("/api/public/register-student", studentH.PublicRegister)

	// Asset file serving is intentionally public — UUID provides sufficient obscurity,
	// and content (exam question images/PDFs) will be visible to students during exams anyway.
	// Removing the auth requirement allows <img src="..."> tags to load directly in browsers.
	r.Get("/api/cbt/assets/{id}/file", questionAssetH.File)

	// Exam endpoints — authenticated via X-Exam-Token (no JWT needed)
	r.Post("/api/exam/login", examH.Login)
	r.Group(func(r chi.Router) {
		r.Use(examTokenMW)
		r.Get("/api/exam/status", examH.Status)
		r.Post("/api/exam/heartbeat", examH.Heartbeat)
		r.Post("/api/exam/event", examH.RecordEvent)
		r.Post("/api/exam/answer", examH.SubmitAnswer)
		r.Post("/api/exam/submit", examH.Submit)
	})

	requireAdmin := mw.RequireAdmin(internalKey)

	r.Group(func(r chi.Router) {
		r.Use(mw.InternalKeyOrJWT(internalKey, jwtSecret, authSvc.CurrentAuthVersion))
		r.Use(mw.Audit(q))
		r.Post("/api/auth/change-password", authH.ChangePassword)

		// Employees are admin-only
		r.Group(func(r chi.Router) {
			r.Use(requireAdmin)
			r.Get("/api/employees", empH.List)
			r.Post("/api/employees", empH.Create)
			r.Get("/api/employees/{id}", empH.Get)
			r.Get("/api/employees/{id}/pusaka-status", empH.GetPusakaStatus)
			r.Post("/api/employees/{id}/update-pusaka", empH.UpdatePusakaCredentials)
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
		r.With(requireAdmin).Delete("/api/academic/{entity}/{id}", academicH.Delete)
		r.Get("/api/grades", gradeH.Overview)
		r.Post("/api/grades/components", gradeH.CreateComponent)
		r.Delete("/api/grades/components/{id}", gradeH.DeleteComponent)
		r.Post("/api/grades/components/{id}/entries", gradeH.UpsertEntry)

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
		r.Get("/api/portal/parent/me", portalH.ParentMe)

		r.Get("/api/cbt/questions", questionH.List)
		r.Post("/api/cbt/questions", questionH.Create)
		r.Get("/api/cbt/questions/{id}", questionH.Get)
		r.Put("/api/cbt/questions/{id}", questionH.Update)
		r.Post("/api/cbt/questions/{id}/duplicate", questionH.Duplicate)
		r.Patch("/api/cbt/questions/{id}/workflow", questionH.WorkflowAction)
		r.Delete("/api/cbt/questions/{id}", questionH.Delete)
		r.Get("/api/cbt/assets", questionAssetH.List)
		r.Post("/api/cbt/assets", questionAssetH.Upload)

		r.Get("/api/cbt/packages", packageH.List)
		r.Post("/api/cbt/packages", packageH.Create)
		r.Delete("/api/cbt/packages/{id}", packageH.Delete)

		// CBT Events (kegiatan ujian) — admin manages, guru reads
		r.Get("/api/cbt/events", eventH.List)
		r.With(requireAdmin).Post("/api/cbt/events", eventH.Create)
		r.Get("/api/cbt/events/{id}", eventH.Get)
		r.Get("/api/cbt/events/{id}/results", eventH.GetResults)
		r.Get("/api/cbt/events/{id}/exam-cards", eventH.GetExamCards)
		r.With(requireAdmin).Put("/api/cbt/events/{id}", eventH.Update)
		r.With(requireAdmin).Patch("/api/cbt/events/{id}/status", eventH.UpdateStatus)
		r.With(requireAdmin).Delete("/api/cbt/events/{id}", eventH.Delete)

		// CBT Sessions
		r.Get("/api/cbt/sessions", sessionH.GuruAwareList)
		r.Post("/api/cbt/sessions", sessionH.Create)
		r.Get("/api/cbt/sessions/{id}", sessionH.Get)
		r.With(requireAdmin).Patch("/api/cbt/sessions/{id}/status", sessionH.UpdateStatus)
		r.With(requireAdmin).Delete("/api/cbt/sessions/{id}", sessionH.Delete)

		// Participants & Enrollment
		r.Get("/api/cbt/sessions/{id}/participants", sessionH.GuruAwareParticipants)
		r.Post("/api/cbt/sessions/{id}/enroll", sessionH.Enroll)
		r.Post("/api/cbt/sessions/{id}/enroll-grade", sessionH.EnrollGrade)
		r.Post("/api/cbt/sessions/{id}/enroll-school", sessionH.EnrollSchool)
		r.Post("/api/cbt/sessions/{id}/generate-tokens", sessionH.GenerateTokens)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/regenerate-token", sessionH.RegenerateToken)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/seat", sessionH.AssignSeat)
		r.Post("/api/cbt/sessions/{id}/seats/auto", sessionH.AutoAssignSeats)

		// Rooms & Shuffle
		r.Get("/api/cbt/sessions/{id}/rooms", sessionH.ListRooms)
		r.Post("/api/cbt/sessions/{id}/rooms", sessionH.CreateRoom)
		r.Delete("/api/cbt/sessions/{id}/rooms/{rid}", sessionH.DeleteRoom)
		r.Post("/api/cbt/sessions/{id}/shuffle-rooms", sessionH.ShuffleRooms)
		r.Get("/api/cbt/sessions/{id}/minutes", sessionH.GetMinutes)

		// Scoring & Results
		r.Post("/api/cbt/sessions/{id}/score", sessionH.ScoreSession)
		r.Get("/api/cbt/sessions/{id}/results", sessionH.GuruAwareResults)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/answer", sessionH.RecordAnswer)
		r.Get("/api/cbt/sessions/{id}/participants/{pid}/answers", sessionH.GetParticipantAnswers)

		// Proctoring
		r.Get("/api/cbt/sessions/{id}/proctoring", sessionH.GetProctoringStatus)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/flag", sessionH.FlagParticipant)

		// Essay Grading
		r.Get("/api/cbt/sessions/{id}/ungraded-essays", sessionH.ListUngradedEssays)
		r.Post("/api/cbt/sessions/{id}/answers/{aid}/grade-essay", sessionH.GradeEssay)

		// Library — admin + staf
		r.Get("/api/library/stats",              libraryH.Stats)
		r.Get("/api/library/books",              libraryH.ListBooks)
		r.Post("/api/library/books",             libraryH.CreateBook)
		r.Put("/api/library/books/{id}",         libraryH.UpdateBook)
		r.Delete("/api/library/books/{id}",      libraryH.DeleteBook)
		r.Get("/api/library/loans",              libraryH.ListLoans)
		r.Post("/api/library/loans",             libraryH.LoanBook)
		r.Post("/api/library/loans/{id}/return", libraryH.ReturnBook)
		r.Post("/api/library/loans/{id}/lunas",  libraryH.MarkDendaLunas)

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
		r.Get("/api/pusaka/worker/status", pusakaWorkerH.GetStatus)
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
