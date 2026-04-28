package main

import (
	"context"
	"log"
	"net/http"
	"os"
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
	_ = godotenv.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	q := db.New(pool)

	authSvc := service.NewAuth(q, mustEnv("JWT_SECRET"), getEnv("ADMIN_PASSWORD", ""))
	academicSvc := service.NewAcademic(q)
	empSvc := service.NewEmployee(q)
	studentSvc := service.NewStudent(q)
	jobSvc := service.NewJob(q)
	attSvc := service.NewAttendance(q)
	questionSvc := service.NewCbtQuestion(q)
	packageSvc := service.NewCbtPackage(pool)
	sessionSvc := service.NewCbtSession(pool)
	eventSvc := service.NewCbtEvent(pool)
	examSvc := service.NewExam(pool)
	schedSvc := service.NewSchedule(q)
	settSvc := service.NewSetting(q)
	schedulerSvc := service.NewScheduler(q, jobSvc, settSvc)

	if err := authSvc.SeedAdmin(ctx); err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if err := settSvc.SeedDefaults(ctx); err != nil {
		log.Fatalf("seed settings: %v", err)
	}
	schedulerSvc.Start(ctx)

	authH := handler.NewAuth(authSvc)
	academicH := handler.NewAcademic(academicSvc)
	empH := handler.NewEmployee(empSvc)
	healthH := handler.NewHealth(pool, jobSvc, settSvc)
	jobH := handler.NewJob(jobSvc)
	attH := handler.NewAttendance(attSvc)
	studentH := handler.NewStudent(studentSvc)
	questionH := handler.NewCbtQuestion(questionSvc)
	packageH := handler.NewCbtPackage(packageSvc)
	sessionH := handler.NewCbtSession(sessionSvc)
	eventH := handler.NewCbtEvent(eventSvc)
	examH := handler.NewExam(examSvc)
	schedH := handler.NewSchedule(schedSvc)
	settH := handler.NewSetting(settSvc)
	schedulerH := handler.NewScheduler(schedulerSvc)
	workerH := handler.NewWorker(jobSvc, attSvc, settSvc)

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

	r.Group(func(r chi.Router) {
		r.Use(mw.InternalKeyOrJWT(internalKey, jwtSecret, authSvc.CurrentAuthVersion))
		r.Post("/api/auth/change-password", authH.ChangePassword)

		r.Get("/api/employees", empH.List)
		r.Post("/api/employees", empH.Create)
		r.Get("/api/employees/{id}", empH.Get)
		r.Get("/api/employees/{id}/pusaka-status", empH.GetPusakaStatus)
		r.Post("/api/employees/{id}/update-pusaka", empH.UpdatePusakaCredentials)
		r.Put("/api/employees/{id}", empH.Update)
		r.Delete("/api/employees/{id}", empH.Delete)

		r.Get("/api/academic", academicH.Overview)
		r.Post("/api/academic/{entity}", academicH.Create)
		r.Delete("/api/academic/{entity}/{id}", academicH.Delete)

		r.Get("/api/students", studentH.List)
		r.Post("/api/students", studentH.Create)
		r.Delete("/api/students/{id}", studentH.Delete)

		r.Get("/api/cbt/questions", questionH.List)
		r.Post("/api/cbt/questions", questionH.Create)
		r.Delete("/api/cbt/questions/{id}", questionH.Delete)

		r.Get("/api/cbt/packages", packageH.List)
		r.Post("/api/cbt/packages", packageH.Create)
		r.Delete("/api/cbt/packages/{id}", packageH.Delete)

		// CBT Events (kegiatan ujian)
		r.Get("/api/cbt/events", eventH.List)
		r.Post("/api/cbt/events", eventH.Create)
		r.Get("/api/cbt/events/{id}", eventH.Get)
		r.Put("/api/cbt/events/{id}", eventH.Update)
		r.Patch("/api/cbt/events/{id}/status", eventH.UpdateStatus)
		r.Delete("/api/cbt/events/{id}", eventH.Delete)

		// CBT Sessions
		r.Get("/api/cbt/sessions", sessionH.List)
		r.Post("/api/cbt/sessions", sessionH.Create)
		r.Get("/api/cbt/sessions/{id}", sessionH.Get)
		r.Patch("/api/cbt/sessions/{id}/status", sessionH.UpdateStatus)
		r.Delete("/api/cbt/sessions/{id}", sessionH.Delete)

		// Participants & Enrollment
		r.Get("/api/cbt/sessions/{id}/participants", sessionH.ListParticipants)
		r.Post("/api/cbt/sessions/{id}/enroll", sessionH.EnrollClass)
		r.Post("/api/cbt/sessions/{id}/enroll-grade", sessionH.EnrollGrade)
		r.Post("/api/cbt/sessions/{id}/enroll-school", sessionH.EnrollSchool)
		r.Post("/api/cbt/sessions/{id}/generate-tokens", sessionH.GenerateTokens)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/regenerate-token", sessionH.RegenerateToken)

		// Rooms & Shuffle
		r.Get("/api/cbt/sessions/{id}/rooms", sessionH.ListRooms)
		r.Post("/api/cbt/sessions/{id}/rooms", sessionH.CreateRoom)
		r.Delete("/api/cbt/sessions/{id}/rooms/{rid}", sessionH.DeleteRoom)
		r.Post("/api/cbt/sessions/{id}/shuffle-rooms", sessionH.ShuffleRooms)

		// Scoring & Results
		r.Post("/api/cbt/sessions/{id}/score", sessionH.ScoreSession)
		r.Get("/api/cbt/sessions/{id}/results", sessionH.GetResults)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/answer", sessionH.RecordAnswer)
		r.Get("/api/cbt/sessions/{id}/participants/{pid}/answers", sessionH.GetParticipantAnswers)

		// Proctoring
		r.Get("/api/cbt/sessions/{id}/proctoring", sessionH.GetProctoringStatus)
		r.Post("/api/cbt/sessions/{id}/participants/{pid}/flag", sessionH.FlagParticipant)

		// Essay Grading
		r.Get("/api/cbt/sessions/{id}/ungraded-essays", sessionH.ListUngradedEssays)
		r.Post("/api/cbt/sessions/{id}/answers/{aid}/grade-essay", sessionH.GradeEssay)

		r.Get("/api/jobs", jobH.List)
		r.Post("/api/jobs", jobH.Create)
		r.Get("/api/jobs/stats", jobH.Stats)
		r.Post("/api/jobs/run-all", jobH.RunAll)
		r.Post("/api/jobs/cancel", jobH.CancelEmployee)
		r.Post("/api/jobs/cancel-all", jobH.CancelAll)
		r.Post("/api/jobs/sync-attendance", jobH.SyncAttendance)

		r.Get("/api/attendance", attH.List)
		r.Get("/api/attendance/by-date/{date}", attH.ByDate)
		r.Get("/api/attendance/by-employee/{id}", attH.ByEmployee)

		r.Get("/api/schedules", schedH.List)
		r.Put("/api/schedules/{id}", schedH.Upsert)

		r.Get("/api/settings", settH.List)
		r.Put("/api/settings/{key}", settH.Upsert)
		r.Post("/api/scheduler/tick", schedulerH.Tick)
	})

	r.Group(func(r chi.Router) {
		r.Use(mw.WorkerKey(workerKey))
		r.Post("/api/worker/claim", workerH.Claim)
		r.Get("/api/worker/config", workerH.Config)
		r.Get("/api/worker/status", workerH.GetStatus)
		r.Post("/api/worker/heartbeat", workerH.Heartbeat)
		r.Post("/api/worker/jobs/{id}/complete", workerH.Complete)
		r.Post("/api/worker/jobs/{id}/fail", workerH.Fail)
		r.Post("/api/worker/attendance", workerH.UpsertAttendance)
	})

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("pusaka-api listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env: %s", key)
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
