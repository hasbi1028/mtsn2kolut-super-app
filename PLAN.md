# MTs Negeri 2 Kolaka Utara — Super App Strategic Plan

> **Status:** Sprints 1-9 Complete | PUSAKA Isolation Phase 1-3 Complete | Lightweight Ops Hardening Complete | Sprint 11 In Progress | Sprint 15 Library Module Complete | Sprint 16 Public Website Foundation Complete | Last Updated: 2026-04-30
> This file is the master roadmap. Update after each sprint completion.

---

## Executive Summary

Three runtime units deployed across 3 VPS:
- **Web Admin** (SvelteKit) — admin & guru BFF
- **Core API** (Go + sqlc + PostgreSQL) — backend, migrations, scheduler
- **Pusaka Worker** (Playwright) — async PUSAKA attendance automation
- **Flutter App** (planned) — student CBT client

---

## ✅ Completed Sprints (1-5)

### Sprint 1 — CBT Foundation
- [x] Migrations 001-008
- [x] Backend Exam API
- [x] Flutter API documentation
- [x] UI: Room management, shuffle seats, token generation

### Sprint 2 — Master Data CRUD
- [x] Edit soal CBT
- [x] Manajemen siswa
- [x] Manajemen pegawai
- [x] Koreksi essay

### Sprint 3 — Reports & Monitoring
- [x] Attendance reports
- [x] CBT reports
- [x] CSV export
- [x] Dashboard stats

### Sprint 4 — Multi-user & Roles
- [x] Migration 009 (users + audit_logs)
- [x] Auth middleware
- [x] User CRUD
- [x] Audit trail
- [x] BFF role gate

### Sprint 5 — Guru Data Scoping & Hardening
- [x] JWT forwarding
- [x] Guru data scoping
- [x] Dashboard guru
- [x] Structured logging

---

## ✅ Sprint 6A — CBT Standardized Bank, Cohort Scope, Seat Plan, and Print Ops (COMPLETE)
- [x] Standardize question bank schema
- [x] Add question asset upload
- [x] Add adaptive session cohort model
- [x] Make `mix_policy` adaptive
- [x] Add participant `seat_no`
- [x] Add printable exam cards
- [x] Add printable session minutes
- [x] shared `sonner` toast system
- [x] Stronger seat validation
- [x] Extended package builder and session UI
- [x] Responsive PUSAKA layout
- [x] Responsive table-to-card layouts
- [x] CBT event `target_levels`
- [x] Harden CBT question bank

---

## ✅ Sprint 9 — Academic Foundation & RBAC Expansion (COMPLETE)
- [x] Identity vs Entity Refactor guard rails on account creation and profile linkage
- [x] RBAC Expansion for `admin`, `guru`, `staf`, `siswa`, and `ortu`
- [x] Student lifecycle status controls (`prospective`, `active`, `alumni`, `mutated`) wired to student activation and linked user-account activation
- [x] Parent-Student Linking from parent management plus visibility in student management
- [x] User multi-role admin UI wired for employee/student/parent profiles
- [x] Parent management UI wired through BFF routes, including link/unlink child flow
- [x] User account activation / suspension controls in admin UI
- [x] Unified auth dashboard foundation for `admin`, `guru`, `staf`, `siswa`, and `ortu`
- [x] Staff Rotation Logic via employee active/nonactive transitions with linked user-account sync
- [x] Prospective student portal / PPDB initial registration flow via public BFF route

## ✅ PUSAKA Isolation Phase 1 (COMPLETE)
- [x] Canonical backend namespace `/api/pusaka/*` for jobs, attendance, schedules, settings, and worker runtime
- [x] Worker migrated to `/api/pusaka/worker/*`
- [x] Web admin PUSAKA pages migrated to `/api/pusaka/*`
- [x] Legacy `/attendance` and `/jobs` pages redirected to canonical `/pusaka/*`
- [x] Legacy backend and BFF paths kept as compatibility bridges during transition

## 📋 Next Sprint Candidates

### ✅ PUSAKA Isolation Phase 2 (COMPLETE)
- [x] Move PUSAKA credentials out of `employees` into dedicated `pusaka_accounts`
- [x] Rename generic service/handler ownership toward explicit `Pusaka*` internal naming where helpful
- [x] Retire legacy `/api/jobs`, `/api/attendance`, `/api/schedules`, `/api/settings`, and `/api/worker` runtime aliases after clients fully migrate

### ✅ PUSAKA Isolation Phase 3 (COMPLETE)
- [x] Drop legacy `employees.pusaka_*` columns after the migration window by moving integration ownership fully to `pusaka_accounts`
- [x] Keep package layout stable for now; explicit `Pusaka*` ownership naming is sufficient until real churn justifies a deeper package split
- [x] Keep `/employees` as master pegawai umum and move PUSAKA account setup/operations to `/pusaka/employees`
- [x] Remove legacy BFF aliases for employee-scoped PUSAKA actions under `/api/employees/{id}/*`

### Sprint 7 — Flutter CBT App (MVP)
- [ ] Initialize Flutter project
- [ ] Student login screen
- [ ] Question renderer
- [ ] Anti-cheat

### ✅ Lightweight Ops Hardening (COMPLETE)
- [x] Basic CI for `go test ./...`, `npm run check`, and worker typecheck
- [x] PostgreSQL backup automation via root make target and deploy script
- [x] Uptime / health monitoring via root make targets and deploy health script
### Sprint 8 — Real-time Proctoring
- [ ] WebSocket endpoint
- [ ] Live proctoring dashboard

### Sprint 10 — Notifications & Communication
- [ ] WhatsApp notification service
- [ ] Exam reminders

### Sprint 11 — Rapor / Grade Management
- [x] Grade schema foundation (`grade_components`, `grade_entries`)
- [x] Backend gradebook API foundation (`/api/grades`)
- [x] Web admin gradebook page foundation (`/grades`)
- [x] Auth hardening baseline — explicit suspended/weak-password errors, backend password policy, non-destructive `SeedAdmin`
- [x] BFF auth forwarding baseline — authenticated proxy helpers now require bearer JWT instead of silently falling back to internal key
- [x] Refresh session baseline — session-backed refresh tokens with backend revoke on logout and rotation on refresh
- [x] Logout-all baseline — per-user `auth_version` plus revoke-all refresh sessions without affecting other users
- [ ] Rapor PDF generation

### CBT Question Authoring UX
- [x] Beginner mode for quick teacher authoring with minimal required fields
- [x] Advance mode for full CP/TP/KD, workflow, rich asset, and LaTeX authoring
- [x] Beginner review flow (`Lengkapi di Advanced`, `Ajukan Review`)
- [x] Participant-style preview for question authoring
- [x] Stronger workflow role gate (`guru/admin` review, `admin` approve/publish/archive)
- [x] Six-route frontend experiment for `/cbt/questions` (`studio`, `wizard`, `grid`, `document`, `review`, `package-fit`) with one shared backend contract
- [x] Local browser-based teacher evaluation notes per variant for manual UX comparison

### ✅ Sprint 6B — Komposer Soal `/cbt/soal` (COMPLETE)
- [x] New route `/cbt/soal` — dedicated question composer with richer authoring UX
- [x] 9 pre-built question templates by mata pelajaran (Standar, Stimulus, Matematika, Cerita Hitung, Grafik, Arab Mufradat, Arab Qiraah, Dalil, Sains Analisis)
- [x] Readiness score — real-time 0–100% progress bar (8 checks: subject, stem, options A-D, answerKey, weight)
- [x] Quality signals — 4 UX heuristics (stem length, distraktor variety, option balance, media support)
- [x] Split preview Dialog — large modal with form left + live KaTeX-rendered preview right
- [x] RTL toggle — Arabic/Quran question preview mode (`dir="rtl"`)
- [x] Draft autosave — localStorage, 700ms debounce, auto-restore on open
- [x] KaTeX two-pass renderer utility (`src/lib/utils/render-rich-math.ts`)
- [x] All data via existing Go API BFF (no new backend routes, no SQLite/Drizzle)
- [x] Sidebar "Komposer Soal" entry under CBT group with pen-tool icon

### Sprint 12 — Schedule & Timetable
### Sprint 13 — Inventory & Asset Management
### Sprint 14 — Fee & Payment Management
### ✅ Sprint 15 — Library System (COMPLETE)
- [x] Migration 027 — `library_books`, `library_loans` tables with FK to existing `students` and `employees`
- [x] sqlc queries — ListBooks, GetBook, CreateBook, UpdateBook, DeleteBook, Decrement/IncrementTersedia, ListLoans, GetLoan, CreateLoan, UpdateLoanReturn, MarkLoanDendaLunas, CountActiveLoansForMember, GetLibraryStats
- [x] Go service — `LoanBook` (availability + max 3 loan check), `ReturnBook` (denda calculation), `MarkDendaLunas`, `GetStats`
- [x] Go handler — Stats, ListBooks, CreateBook, UpdateBook, DeleteBook, ListLoans, LoanBook, ReturnBook, MarkDendaLunas (admin + staf RBAC)
- [x] Router registration in `cmd/api/main.go` (`/api/library/*` routes)
- [x] BFF proxy routes — stats, books, books/[id], loans, loans/[id]/return, loans/[id]/lunas
- [x] Dashboard `/library` — 6 stat cards, active loans table, overdue loans table
- [x] Katalog Buku `/library/books` — search, kategori filter, table, beginner/advance form dialog, delete confirm
- [x] Peminjaman `/library/loans` — tab filter, autocomplete member/book search, beginner/advance loan dialog, return confirm with denda estimate, lunas confirm
- [x] Beginner/advance mode on all forms (consistent with CBT question bank UX)
- [x] Sidebar "Perpustakaan" group (Dashboard, Katalog Buku, Peminjaman) — admin + staf only
- [x] `npm run check` — 0 errors, 0 warnings

### ✅ Sprint 16 — Public Website Foundation (COMPLETE)
- [x] Backend website content domain via migration 028 (`website_contents` for `page`, `post`, `announcement`)
- [x] sqlc queries and Go service/handler for admin CRUD and public published reads
- [x] Public routes `/`, `/profil`, `/berita`, `/berita/[slug]`, `/pengumuman`, `/pengumuman/[slug]`, `/ppdb`, `/kontak`
- [x] Dedicated public shell for unauthenticated visitors while keeping admin dashboard behavior on `/` for logged-in users
- [x] Editorial admin screens under `/website`, `/website/posts`, `/website/announcements`, `/website/pages`
- [x] Public homepage aggregation for profile, PPDB information, latest posts, and recent announcements
- [x] Global SvelteKit `+error.svelte` for shared 403/404/500-style error handling across public and admin pages
- [x] `go test ./...` and `npm run check` green after integration

### Next Recommendation — Sprint 16B Public Website Polish
- [ ] Add cover-image upload/reuse instead of URL-only cover input for website content
- [ ] Add publish scheduling and featured content flags for homepage curation
- [ ] Add richer public SEO metadata per page/post/announcement
- [ ] Add dedicated custom pages for `500` maintenance-style incidents if operational need appears

---

## Architecture Decisions Log
(See previous version for details)

---

## Risk Register
(See previous version for details)
