# MTs Negeri 2 Kolaka Utara — Super App Strategic Plan

> **Status:** Sprints 1-9 Complete | PUSAKA Isolation Phase 1-3 Complete | Lightweight Ops Hardening Complete | Sprint 11 In Progress | Sprint 15 Library Module Complete | Sprint 16 Public Website Foundation Complete | Sprint 16B Public Website Polish Complete | Rapor Print View Complete | Last Updated: 2026-04-30
> This file is the master roadmap. Update after each sprint completion.

---

## Executive Summary

Three runtime units deployed across 3 VPS:
- **Web Admin** (SvelteKit) — admin & guru BFF
- **Core API** (Go + sqlc + PostgreSQL) — backend, migrations, scheduler
- **Pusaka Worker** (Playwright) — async PUSAKA attendance automation
- **Flutter App** (initialized) — student CBT client

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

### ✅ Sprint 7 — Flutter CBT App (MVP) (COMPLETE)
- [x] Initialize Flutter project in `apps/mobile`
- [x] Student login screen with token + configurable API base URL
- [x] Question renderer for pilihan ganda and uraian with per-question save flow
- [x] Anti-cheat baseline via lifecycle/app-switch event logging, disabled back navigation, heartbeat, and server status sync
- [x] Local session restore baseline via persisted token/base URL/current question/answer snapshot
- [x] Android secure-screen baseline via `FLAG_SECURE`
- [x] BYOD-aware resume gate and pending-answer sync queue to maximize deterrence and answer safety on student-owned Android devices
- [x] Visible sync-state chip plus warning-event telemetry for repeat resume, local-only answer saves, and blocked submit due to pending sync
- [x] Rich-content-ready renderer baseline for `stimulus_html` and `stem_html` payloads with safe text normalization fallback

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
- [x] Session management baseline — `sub=user_id`, `last_used_at` on auth sessions, active-session list and per-session revoke in settings
- [x] Session metadata baseline — client IP, user-agent, and derived device label shown in active-session management
- [x] Auth audit baseline — structured audit events for login, refresh, logout, logout-all, and per-session revoke
- [x] Auth audit UX baseline — audit trail screen now has quick `Auth & Session` filtering for structured auth events
- [x] Session rename baseline — users can update their own `device_label` from the settings screen
- [x] Access-session validation baseline — access JWTs now require a still-active referenced auth session (`ssid`) in addition to per-user auth-version checks
- [x] Loading UX baseline — global route progress, reusable skeleton primitive, and reusable loading button applied across primary admin screens, key PUSAKA pages, parents/users management, library flows, website editorial, PPDB public form, CBT session operations, print views, question composer, and employee/PUSAKA operational components
- [x] Empty-state UX baseline — academic, students, gradebook, and question-bank screens now use more guided empty states with clearer next-step cues instead of plain “data kosong” table fallbacks
- [x] Recovery-state UX baseline — key admin, PUSAKA, and library overview screens now surface inline retry panels for failed fetches instead of relying only on toast messages
- [x] Density & filter-bar polish baseline — library catalog and circulation screens now use denser summary cards, clearer filter surfaces, and more consistent header spacing with the rest of the admin panel
- [x] Ops-admin route polish baseline — employees, parents, and user-management screens now have stronger route-level hierarchy, summary context, and clearer empty/recovery surfaces
- [x] Website editorial polish baseline — news, announcement, and public-page management now share denser summary context, clearer search surfaces, and better empty/recovery states via the shared content manager
- [x] Success-state inline baseline — major create/publish/link flows now surface contextual success panels in-page for employee creation, parent-child linking, and website editorial actions instead of relying only on transient toasts
- [x] Ops-component polish baseline — `GeneralEmployeeList`, `EmployeeList`, and `ScheduleList` now use denser control surfaces, clearer internal empty states, and inline success feedback for day-to-day operator actions
- [x] UI copy consistency baseline — the most visible mixed English/Indonesian admin copy in editorial, PUSAKA controls, and user-status badges is now normalized toward one institutional Indonesian tone
- [x] Public-site copy consistency baseline — homepage, list, and detail content screens now use a more consistent institutional Indonesian tone for publishing states, CTAs, and public-facing helper text
- [x] Form microcopy baseline — operator-facing placeholders and helper text in employee, student, user, and website editorial forms are now more directive and less generic
- [x] Dashboard copy baseline — the shared `/` dashboard now uses more consistent role titles, stat labels, and empty-state copy across guru, siswa, orang tua, admin, and staf experiences
- [x] CBT copy consistency baseline — bank soal, komposer soal, paket ujian, kegiatan ujian, serta daftar dan detail sesi now use more consistent institutional Indonesian wording for review flow, publication state, mode naming, scoring, token actions, coverage labels, and operator CTAs
- [x] CBT action-safety baseline — package deletion, session deletion, sensitive session status changes, mass token generation, room shuffle, seat auto-assignment, room deletion, and participant token reset now use stronger operator confirmation plus persistent inline operation feedback
- [x] CBT workflow-safety baseline — question-bank workflow actions such as delete, submit review, approve, publish, and archive now expose stronger confirmation and persistent inline feedback instead of relying only on toast
- [x] PUSAKA action-safety baseline — scheduler trigger, rekap massal, cancel-all, account enable/disable, account deletion, individual run actions, and stop-job actions now provide stronger operator confirmation with clearer inline success/error feedback
- [x] Status badge consistency baseline — key admin screens now prefer one visible passive-state term (`Nonaktif`) while preserving action verbs like `Nonaktifkan`, reducing mixed status wording across users, students, and academic master data
- [x] Public detail reading baseline — berita and pengumuman detail pages now use a calmer reading layout with stronger header hierarchy, cleaner content density, and a lightweight side summary instead of one long undecorated article column
- [x] Public listing rhythm baseline — berita and pengumuman listing pages now use a stronger editorial header, more stable card rhythm, and a cleaner published-content grid for public visitors
- [x] PPDB public landing baseline — `/ppdb` now has a fuller public-facing layout with clearer registration flow, preparation guidance, and post-submit expectations instead of a single bare form card
- [x] Public static-page support baseline — `/profil` and `/kontak` now enrich the shared public detail template with page-specific side guidance instead of relying only on a generic article sidebar
- [x] Public homepage polish baseline — homepage now has stronger hero support cards and clearer editorial CTA rhythm so its information density matches the refined public detail and listing pages
- [x] Public shell polish baseline — shared header and footer now provide stronger school-facing navigation, clearer service CTAs, and a more complete institutional frame for all public pages
- [x] Svelte key hygiene baseline — high-traffic admin screens now key their dynamic `#each` blocks more consistently, reducing autofixer noise and improving DOM stability in sessions, events, users, and students screens
- [x] Svelte reactivity hygiene baseline — lightweight reactive collection cleanup has started, including replacing mutable `Set` usage in CBT package selection with `SvelteSet` patterns that satisfy Svelte 5 diagnostics without widening the refactor scope
- [x] Rapor print view — printable HTML layout at `/grades/rapor` using existing `/api/grades` endpoint, color-coded scores, school header, signature area, sidebar entry

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

### ✅ Sprint 16B — Public Website Polish (COMPLETE)
- [x] Migration 033 — `is_featured`, `meta_title`, `meta_description` columns on `website_contents`
- [x] `ListFeaturedWebsiteContents` query and handler (GET `/api/public/site/posts/featured`)
- [x] `ListPublishedWebsiteContents` orders featured content first
- [x] Cover image file upload via `WebsiteMedia` handler (POST `/api/website/media`, GET `/api/website/media/{filename}`) — md5-hashed filenames, MIME validation, path traversal protection
- [x] BFF proxy at `/api/website/media` forwards JWT to Go backend
- [x] `WebsiteContentManager`: featured badge, cover upload button, image preview, SEO section with character counter
- [x] SEO meta tags on `/berita/[slug]` and `/pengumuman/[slug]` — `og:title`, `og:description`, `og:image`, `description`
- [x] `go build ./...` and `npm run check` — 0 errors, 0 warnings

### Next Recommendation — Sprint 17
- [ ] Publish scheduling (deferred from 16B — low priority until editorial demand is proven)
- [ ] Featured content homepage widget (pull from `/api/public/site/posts/featured`)
- [ ] Flutter Student App — CBT exam client (API ready via `docs/exam-api.md`)

---

## Architecture Decisions Log
(See previous version for details)

---

## Risk Register
(See previous version for details)
