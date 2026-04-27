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

	"github.com/pusaka/backend/internal/handler"
	mw "github.com/pusaka/backend/internal/middleware"
	db "github.com/pusaka/backend/internal/repository/postgres"
	"github.com/pusaka/backend/internal/service"
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

	authSvc := service.NewAuth(q, mustEnv("JWT_SECRET"), getEnv("ADMIN_PASSWORD", "admin"))
	empSvc := service.NewEmployee(q)
	jobSvc := service.NewJob(q)
	attSvc := service.NewAttendance(q)
	schedSvc := service.NewSchedule(q)
	settSvc := service.NewSetting(q)

	if err := authSvc.SeedAdmin(ctx); err != nil {
		log.Printf("warn: seed admin: %v", err)
	}

	authH := handler.NewAuth(authSvc)
	empH := handler.NewEmployee(empSvc)
	jobH := handler.NewJob(jobSvc)
	attH := handler.NewAttendance(attSvc)
	schedH := handler.NewSchedule(schedSvc)
	settH := handler.NewSetting(settSvc)
	workerH := handler.NewWorker(jobSvc, attSvc)

	jwtSecret   := mustEnv("JWT_SECRET")
	workerKey   := mustEnv("WORKER_API_KEY")
	internalKey := getEnv("INTERNAL_API_KEY", "")

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.SetHeader("Content-Type", "application/json"))

	r.Post("/api/auth/login", authH.Login)

	r.Group(func(r chi.Router) {
		r.Use(mw.InternalKeyOrJWT(internalKey, jwtSecret))
		r.Post("/api/auth/change-password", authH.ChangePassword)

		r.Get("/api/employees", empH.List)
		r.Post("/api/employees", empH.Create)
		r.Get("/api/employees/{id}", empH.Get)
		r.Put("/api/employees/{id}", empH.Update)
		r.Delete("/api/employees/{id}", empH.Delete)

		r.Get("/api/jobs", jobH.List)
		r.Post("/api/jobs", jobH.Create)
		r.Get("/api/jobs/stats", jobH.Stats)
		r.Post("/api/jobs/run-all", jobH.RunAll)
		r.Post("/api/jobs/cancel", jobH.CancelEmployee)
		r.Post("/api/jobs/cancel-all", jobH.CancelAll)

		r.Get("/api/attendance", attH.List)
		r.Get("/api/attendance/by-date/{date}", attH.ByDate)
		r.Get("/api/attendance/by-employee/{id}", attH.ByEmployee)

		r.Get("/api/schedules", schedH.List)
		r.Put("/api/schedules/{id}", schedH.Upsert)

		r.Get("/api/settings", settH.List)
		r.Put("/api/settings/{key}", settH.Upsert)
	})

	r.Group(func(r chi.Router) {
		r.Use(mw.WorkerKey(workerKey))
		r.Post("/api/worker/claim", workerH.Claim)
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
