# MTs Negeri 2 Kolaka Utara — Super App Strategic Plan

> **Status:** Sprints 1-5 Complete | Sprint 6 Next | Last Updated: 2026-04-28
> This file is the master roadmap. Update after each sprint completion.

---

## Executive Summary

Three runtime units deployed across 3 VPS:
- **Web Admin** (SvelteKit) — admin & guru BFF
- **Core API** (Go + sqlc + PostgreSQL) — backend, migrations, scheduler
- **Pusaka Worker** (Playwright) — async PUSAKA attendance automation
- **Flutter App** (planned) — student CBT client

---

## ✅ Completed Sprints (1-3)

### Sprint 1 — CBT Foundation
- [x] Migrations 001-008 (initial schema → CBT events, rooms, anti-cheat)
- [x] Backend Exam API (login, status, heartbeat, event, answer, submit)
- [x] Flutter API documentation (`docs/exam-api.md`)
- [x] UI: Room management, shuffle seats, token generation

### Sprint 2 — Master Data CRUD
- [x] Edit soal CBT (inline dialog)
- [x] Manajemen siswa (edit + assign class)
- [x] Manajemen pegawai (edit + PUSAKA credentials)
- [x] Koreksi essay (manual grading UI)

### Sprint 3 — Reports & Monitoring
- [x] Attendance reports (date range filter, monthly summary)
- [x] CBT reports (event-wide results, session results)
- [x] CSV export on all major tables (WITA localized)
- [x] Dashboard stats (student/class/subject counts)

---

## ✅ Sprint 4 — Multi-user & Roles (COMPLETE)

**Goal:** Admin/guru role separation with audit trail.

| # | Task | Files Touched | Status |
|---|------|---------------|--------|
| 4.1 | Migration `009_multi_user_roles.sql` (users + audit_logs) | `db/migrations/` | ✅ Done |
| 4.2 | Auth middleware role checking (`RequireAdmin`) | `internal/middleware/auth.go` | ✅ Done |
| 4.3 | Login returns role info via JWT claims (`role`, `uid`, `eid`) | `internal/service/auth.go` | ✅ Done |
| 4.4 | User CRUD endpoints | `internal/handler/user.go`, `db/queries/users.sql` | ✅ Done |
| 4.5 | Audit trail middleware (auto-log all 2xx mutations) | `internal/middleware/audit.go` | ✅ Done |
| 4.6 | BFF role gate in `hooks.server.ts` (admin-only paths) | `apps/web-admin/src/hooks.server.ts` | ✅ Done |
| 4.7 | Role-filtered sidebar | `apps/web-admin/src/lib/components/Sidebar.svelte` | ✅ Done |
| 4.8 | User management page + Audit Trail page | `apps/web-admin/src/routes/settings/users/`, `audit-logs/` | ✅ Done |
| 4.9 | SeedAdmin via service.Auth (no SQL fake hash seed) | `internal/service/auth.go` | ✅ Done |
| 4.10 | Backend admin gate on routes (employees, jobs, attendance, settings, users, master CUD, cbt events CUD) | `cmd/api/main.go` | ✅ Done |

**Bug Fixes (in Sprint 4 batch):**
- ✅ `handler/user.go` `pgUUIDString` returned empty string (broken stub).
- ✅ `service/utils.go` `pageSize` was placeholder returning 0 — removed.
- ✅ Migration 009 had fake bcrypt hash that would lock out admin — removed.
- ✅ `main.go` registered `POST /api/settings` (handler expects `PUT /api/settings/{key}`).
- ✅ `auth_test.go` was broken after auth refactor to users table — rewritten.

**Acceptance verified:**
- ✅ `go test ./...` passes (4 service test files green).
- ✅ `go build ./...` clean.
- ✅ `go vet ./...` clean.
- ✅ `npm run check` 0 errors (17 pre-existing a11y warnings, not Sprint 4 introduced).
- ⚠️ **Not yet:** guru data scoping (sees same data as admin in CBT pages — only routes are gated, not row-level filter). Pushed to Sprint 5.

---

## 📋 Future Sprints (Roadmap)

## ✅ Sprint 5 — Guru Data Scoping & Hardening (COMPLETE)

**Priority:** High
**Estimated:** 1-2 weeks ✅ Done
**Dependencies:** Sprint 4 completion

| # | Task | Why | Scope | Status |
|---|------|-----|-------|--------|
| 5.1 | **JWT forwarding di BFF** — 33 proxy routes updated, `handleFetch` middleware | Audit log now has real user_id, backend sees user identity | BFF + backend mw | ✅ |
| 5.2 | **Guru data scoping** — 4 new SQL queries (sessions, results, participants, students by teacher) | Row-level filtering via `class_subject_assignments` | backend | ✅ |
| 5.3 | **Dashboard guru** — role-aware dashboard (sesi aktif, essay pending, siswa) | UX guru | frontend | ✅ |
| 5.4 | Graceful shutdown — `signal.NotifyContext`, scheduler Stop(), `srv.Shutdown` | Prevents connection drops on restart | backend | ✅ |
| 5.5 | Structured logging — `log/slog` JSON handler, all `log.Printf` → `slog.Info/Error` | Debugging & monitoring | backend | ✅ |
| 5.6 | Token refresh silent on 401 — `handleFetch` hook intercepts 401, refreshes, retries | Auth reliability | frontend | ✅ |
| 5.7 | Migrate legacy components — Nav (no-op), AttendanceTable + JobTable → shadcn Table, legacy CSS removed | Dark-theme CSS gone | frontend | ✅ |
| 5.8 | Audit log retention — `DeleteOldAuditLogs` (>90 days), cleanup in scheduler | Tabel tidak grow unbounded | backend | ✅ |
| 5.9 | Health endpoint detail — DB pool stats, `runtime.Version()`, scheduler ticks | Operational visibility | backend | ✅ |
| 5.10 | `make check` now includes `go vet` + `npm audit --audit-level=high` | Security baseline | all | ✅ |

**Verification:**
- ✅ `go build ./...` clean
- ✅ `go test ./...` passes (4 service test files)
- ✅ `go vet ./...` clean
- ✅ `npm run check` 0 errors (17 pre-existing a11y warnings)
- ✅ 33 BFF proxy routes now forward `Authorization: Bearer` from cookies
- ✅ Guru data scoped via `class_subject_assignments.teacher_employee_id = claims.eid`
- ✅ Dashboard renders guru-specific stats (sessions, essays, students, subjects)

### Sprint 6 — Infrastructure & CI/CD (NEXT)

**Priority:** High
**Estimated:** 2-3 weeks
**Dependencies:** Sprint 5 completion

| # | Task | Why | Scope |
|---|------|-----|-------|
| 6.1 | Set up GitHub Actions CI (lint, test, build) | Catch errors before deploy | infra |
| 6.2 | Add Dockerfiles for all 3 services | Reproducible builds | infra |
| 6.3 | Set up staging environment | Safe testing before prod | infra |
| 6.4 | Add PostgreSQL backup automation | Data safety | deploy |
| 6.5 | Add structured monitoring (prometheus metrics) | Operational insight | backend |
| 6.6 | Set up uptime monitoring (healthchecks.io or similar) | Alert on downtime | infra |
| 6.7 | Replace in-memory rate limiter with redis or DB-backed | Multi-instance scaling | backend |

### Sprint 7 — Flutter CBT App (MVP)

**Priority:** High
**Estimated:** 4-6 weeks
**Dependencies:** Sprints 1-6, Exam API endpoint stability
**API Contract:** `docs/exam-api.md` (already written)

| # | Task | Scope |
|---|------|-------|
| 7.1 | Initialize Flutter project (`apps/mobile`) with Serverpod | mobile |
| 7.2 | Student login screen (exam token) | mobile |
| 7.3 | Exam status & instructions screen | mobile |
| 7.4 | Question renderer (PG A-E, multiple choice) | mobile |
| 7.5 | Question navigation (numbered grid, prev/next) | mobile |
| 7.6 | Answer submission (save per-question, auto-save timer) | mobile |
| 7.7 | Anti-cheat: app switch detection, screenshot attempt, heartbeat | mobile |
| 7.8 | Exam timer + auto-submit on timeout | mobile |
| 7.9 | Submit & results screen | mobile |
| 7.10 | Test with real exam sessions | mobile + backend |

### Sprint 8 — Real-time Proctoring

**Priority:** Medium
**Estimated:** 2-3 weeks
**Dependencies:** Sprint 7 (Flutter app active)

| # | Task | Scope |
|---|------|-------|
| 8.1 | WebSocket endpoint for real-time proctoring | backend |
| 8.2 | Live proctoring dashboard (auto-refresh via WS) | frontend |
| 8.3 | Proctor alerts (offline, suspicious, app switch) | frontend |
| 8.4 | Proctor actions (warn, terminate session) | backend + frontend |
| 8.5 | Session recording playback | backend + frontend |

### Sprint 9 — Notifications & Communication

**Priority:** Medium
**Estimated:** 2-3 weeks
**Dependencies:** Sprint 4 (user system), Sprint 7 (students)

| # | Task | Scope |
|---|------|-------|
| 9.1 | WhatsApp notification service (via API gateway) | backend |
| 9.2 | Exam reminders (scheduled notification) | backend + scheduler |
| 9.3 | Attendance reminders for employees | backend + worker |
| 9.4 | Announcement broadcast (admin → all users) | frontend + backend |

### Sprint 10 — Rapor / Grade Management

**Priority:** Medium
**Estimated:** 3-4 weeks
**Dependencies:** Sprint 1-4 (academic data + users)

| # | Task | Scope |
|---|------|-------|
| 10.1 | Grade schema & migrations (score types, weighting) | backend |
| 10.2 | Grade entry UI (per subject per student) | frontend |
| 10.3 | Grade calculation & final score processing | backend |
| 10.4 | Rapor PDF generation | backend |
| 10.5 | Rapor view/download for guru & admin | frontend |

### Sprint 11 — Schedule & Timetable

**Priority:** Medium
**Estimated:** 3-4 weeks
**Dependencies:** Sprint 10 (academic foundation)

| # | Task | Scope |
|---|------|-------|
| 11.1 | Weekly schedule schema (time slots, teacher-subject-class) | backend |
| 11.2 | Timetable builder UI (drag or select-based) | frontend |
| 11.3 | Teacher schedule view | frontend |
| 11.4 | Conflict detection (teacher double-booked, room overlap) | backend |

### Sprint 12 — Inventory & Asset Management

**Priority:** Low
**Estimated:** 2-3 weeks
**Dependencies:** None (standalone module)

| # | Task | Scope |
|---|------|-------|
| 12.1 | Asset schema (category, location, condition) | backend |
| 12.2 | Asset CRUD with barcode/RFID support | frontend + backend |
| 12.3 | Asset tracking (check-in/check-out for borrowing) | backend |
| 12.4 | Inventory reports & maintenance scheduling | frontend + backend |

### Sprint 13 — Fee & Payment Management

**Priority:** Low
**Estimated:** 3-4 weeks
**Dependencies:** Sprint 10 (student data finalized)

| # | Task | Scope |
|---|------|-------|
| 13.1 | Fee schema (types, amounts, due dates) | backend |
| 13.2 | Student billing & payment tracking | frontend + backend |
| 13.3 | Payment gateway integration (optional) | backend |
| 13.4 | Arrears reports & reminders | frontend + backend |

### Sprint 14 — Library System

**Priority:** Low
**Estimated:** 3-4 weeks
**Dependencies:** None (standalone)

| # | Task | Scope |
|---|------|-------|
| 14.1 | Book catalog schema & CRUD | backend |
| 14.2 | Borrow/return flow with due dates | backend + frontend |
| 14.3 | Library dashboard (popular books, overdue items) | frontend |
| 14.4 | Student library card / history view | frontend |

---

## Technical Debt Backlog

Items that don't fit a dedicated sprint but should be addressed incrementally:

| Item | Effort | Impact | Notes |
|------|--------|--------|-------|
| Increase test coverage (target: >40%) | Medium | High | Start with integration tests for critical paths (auth, exam, CBT) |
| Add e2e tests with Playwright | High | Medium | Worker already uses Playwright — reuse pattern |
| Refactor worker `src/index.ts` (617 lines) | Medium | Medium | Extract modules: browser, pusaka-client, consumer-pool |
| Add OpenAPI spec for all endpoints | Medium | High | Currently no API docs except exam-api.md |
| Remove unused SQLite backup artifacts | Low | Low | `backup.db`, `data/pusaka.sqlite` should be cleaned up |
| Standardize error responses in frontend | Low | Medium | Toast notifications for failed API calls |
| Add DB connection pooling tuning | Low | Medium | Current defaults may not be optimal for concurrent load |
| Document CBT exam flow end-to-end | Low | Medium | For operator training manuals |

---

## Architecture Decisions Log

| Date | Decision | Rationale |
|------|----------|-----------|
| Sprint 1 | PostgreSQL owned by core-api only | Clear data boundary, prevents sync issues |
| Sprint 1 | sqlc over GORM/Ent | Typed SQL, no magic, full control over queries |
| Sprint 1 | Chi over Gin/Echo | Lightweight, stdlib-compatible, middleware pattern |
| Sprint 2 | shadcn-svelte + Tailwind v4 | Consistent UI, institutional theme, good DX |
| Sprint 3 | CSV export via frontend | No backend streaming complexity for now |
| Sprint 4 | Users table separate from employees | Future SSO/OIDC support, cleaner separation |
| Sprint 4 | Audit log in PostgreSQL (not file) | Queryable, integratable with reports |

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| PUSAKA website changes break worker | Medium | High | Screenshot-based debugging, retry logic, alert on failures |
| No backups tested | High | High | Add automated backup with restore testing to Sprint 6 |
| Single dev bottleneck | High | Medium | Document architecture decisions, keep code review process |
| Flutter app complexity underestimated | Medium | High | Start MVP in Sprint 7, iterate based on real usage |
| PostgreSQL single-instance failure | Low | High | Add read replica or backup VPS (future) |
| No rate limiting across instances | Medium | Medium | Move to Redis-based rate limiting in Sprint 6 |
