# MTs Negeri 2 Kolaka Utara — Super App Strategic Plan

> **Status:** Sprints 1-9 Complete | PUSAKA Isolation Phase 1-3 Complete | Lightweight Ops Hardening Complete | Sprint 11 In Progress | Sprint 15 Library Module Complete | Sprint 16 Public Website Foundation Complete | Sprint 16B Public Website Polish Complete | Rapor Print View Complete | Jurnal Kelas Complete | Sprint 17 Persuratan Complete | Sprint 18 Tata Kelola Madrasah Complete | Sprint 19 SKP Mirror Complete | Sprint 20 Bukti Mutu Complete | Sprint 21 Kesiswaan Foundation Complete | Sprint 22 Surat Keterangan Complete | Sprint 23 Kesiswaan Engagement Complete | Sprint 24 Arsip TU Complete | Sprint 25 RKT/RKJM Execution Complete | Sprint 26 TU Dashboard Complete | Sprint 27 Renstra/IKU Alignment Complete | Sprint 28 Governance Print Pack Complete | Sprint 29 School Profile Complete | Sprint 30 Print Surface Letterhead Complete | Sprint 31 Compliance Actions Complete | Sprint 32 Compliance Print/Export Complete | Sprint 33 Compliance Escalation Board Complete | Sprint 34 Compliance Quick Status Complete | Sprint 35 Compliance Evidence Capture Complete | Sprint 36 Compliance Meeting Pack Complete | Sprint 37 Compliance Deadline Calendar Complete | Sprint 38 Governance Control Center Complete | Sprint 39 Compliance Advanced Filters Complete | Sprint 40 PIC Briefing Pack Complete | Sprint 41 8 SNP Briefing Pack Complete | Sprint 42 Evidence Briefing Pack Complete | Sprint 43 Siklus Dokumen Module Complete | Last Updated: 2026-05-01
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
- [x] Human-readable session snapshot metadata cache for clearer restore UX
- [x] Dedicated exam-complete screen after submit / auto-submit
- [x] Schedule-aware restore/completion metadata cache (`start`, `end`, `duration`)
- [x] Lightweight remote image rendering baseline for question stimulus/stem media URLs
- [x] Dedicated restore-failed screen with last known session metadata
- [x] Human-readable connection freshness indicators (`kontak server terakhir`, `gangguan terakhir`)
- [x] Lightweight in-app audio playback baseline for `stimulus_audio_url` and `stem_audio_url`
- [x] Repeated connection-failure warning panel with explicit retry action
- [x] Per-question audio playback state so audio-enabled questions can show `sudah diputar / belum diputar`
- [x] Degraded-mode guard that holds manual submit when repeated sync failures cross the BYOD safety threshold
- [x] Internal APK distribution README for local run, release build, and realistic BYOD trial guidance
- [x] Restore-card health metadata snapshot (`last contact`, `last failure`, repeated-failure count) for clearer BYOD reconnect context
- [x] Internal APK release checklist for operator verification and field trial preparation
- [x] Visual connection-health card in exam shell for clearer BYOD sync interpretation
- [x] Dedicated mobile status-guide screen explaining `Tersambung`, `Lokal`, `Gangguan`, and `Menurun`
- [x] Heartbeat-quality `Waspada` state when last server contact becomes stale even before full degraded mode
- [x] Pengawas-intervention escalation when `Waspada` persists too long without fresh server contact
- [x] Human-readable stale-duration label inside the pengawas-intervention panel
- [x] Second-threshold BYOD stale escalation that hardens the pengawas panel and emits a distinct urgent warning event
- [x] Human-readable escalation-threshold hint inside the pengawas stale-warning panel
- [x] Operator quick-start guide for pengawas during BYOD field trials
- [x] End-to-end BYOD trial procedure covering operator, pengawas, siswa, disturbance simulation, and submit readiness
- [x] Mobile payload compatibility checklist embedded in `docs/exam-api.md` for backend release discipline
- [x] Structured release-note template for exam payload changes in `docs/exam-payload-release-template.md`
- [x] Per-device BYOD test matrix for vendor/model/Android-version comparison during field trials
- [x] Web-admin `/cbt/byod` summary page for pengawas/operator with mobile status legend, submit checklist, and trial references
- [x] Web-admin `/cbt/byod/matrix` companion page for vendor/device comparison during BYOD field trials
- [x] Web-admin `/cbt/byod/release` readiness page for backend/mobile/rollout release checks
- [x] Backend exam payload/status hardening so Flutter-critical fields (`session.title`, `is_submitted`, rich content, media/audio URLs) are actually emitted and covered by backend contract tests
- [x] Handler-level proxy URL tests for exam asset/media absolutization behind `X-Forwarded-*`
- [x] Handler-level login/status JSON envelope tests so mobile contract is checked at the HTTP response layer too
- [x] Handler-level request-forwarding tests for exam login/answer/event parsing into service arguments (token, device fingerprint, forwarded IP, question ID, answer text, event type, event data)
- [x] Handler-level exam error-semantics tests for mobile-critical `404/403/409/401` cases
- [x] Handler-level `answer` / `submit` error-semantics tests for `409` submitted and `403` exam-window-closed cases
- [x] Flutter exam UX now maps backend `404/403/409` semantics into clearer login/restore/save/submit guidance for BYOD sessions
- [x] Flutter exam guidance mappings extracted into testable helpers with dedicated unit coverage
- [x] Persistent `403/409` guidance panel in Flutter exam shell with tested urgency mapping
- [x] Persistent `403/409` guidance panel extended to Flutter token-login and restore-failed flows with tested notice mapping
- [x] Widget-test coverage for rendered login and restore-failed guidance panels so entry-flow BYOD UX is locked at the UI layer too
- [x] Widget-test coverage for rendered warning/danger guidance panels inside the running Flutter exam shell
- [x] Widget-test coverage for shell-level BYOD connection panels (`Waspada`, repeated sync warning, and degraded mode`) driven from restored snapshot state
- [x] Widget-test coverage for the secured resume overlay in both pre-check and in-progress resume states
- [x] Widget-test coverage for restore-health chips and per-question audio state in Flutter exam surfaces
- [x] Widget-test coverage for cached restore card in login flow and lightweight media card in exam shell
- [x] Widget-test coverage for `stimulus_html` and `stem_html` rendering in Flutter exam shell
- [x] Widget-test coverage for sync-chip labels `Tersambung`, `Lokal`, `Waspada`, and `Menurun`
- [x] Widget-test coverage for transient sync-chip labels `Sinkron` and `Cek Ulang`
- [x] Widget-test coverage for sync-chip error label `Gangguan`
- [x] Widget-test coverage for restore-health labels across login and restore-failed entry surfaces (`Terakhir stabil`, `Pernah terganggu`, `Belum ada riwayat koneksi`, `Perlu perhatian koneksi`)
- [x] Widget-test coverage for finish-screen variants and the in-app BYOD status-guide screen
- [x] Unit-test coverage for exam formatting helpers (`formatExamSchedule`, `formatRestoreHealthLabel`, `formatRestoreClock`)
- [x] Unit-test coverage for rich exam text normalization (`normalizeExamText`) including paragraphs, bullets, entities, and whitespace cleanup
- [x] Unit-test coverage for exam model parsers (`ExamLoginPayload`, `ExamQuestion`, `ExamStatusPayload`) including rich fields, media/audio URLs, defaults, and `is_submitted`
- [x] Unit-test coverage for `ExamApiClient` envelope unwrap, token-header transport, login/heartbeat/answer/event/submit outbound payload contracts, missing-data failure, malformed/non-object JSON mapping, transport-failure mapping, and backend message extraction precedence
- [x] Unit-test coverage for transport-level guidance copy (`statusCode == null`) across login, restore, answer, and submit flows
- [x] Widget-test coverage for transport-level warning panels on login, restore-failed, and exam shell surfaces
- [x] Unit-test coverage for `ExamSessionStore` base URL persistence, snapshot roundtrip, invalid/malformed snapshot fallback, and clear-snapshot behavior
- [x] Unit-test coverage for `ExamSessionSnapshot` `toJson/fromJson`, dynamic coercion, and safe defaults
- [x] Handler-level unauthorized coverage for exam `heartbeat` and `event` telemetry routes
- [x] Handler-level success-envelope coverage for exam `heartbeat` and `event` telemetry routes
- [x] Handler-level success-envelope coverage for exam `answer` and `submit` routes
- [x] Handler-level bad-request coverage for malformed exam login/event/answer requests (`invalid json`, missing token, invalid question_id`)
- [x] Handler-level bad-request coverage for empty exam telemetry `event_type`
- [x] Handler-level `401 unauthorized` coverage for exam `answer` and `submit` mutations
- [x] Handler-level `500` coverage for exam telemetry (`heartbeat`, `event`) unexpected service failures
- [x] Handler-level `500` coverage for exam `status` unexpected service failure

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
- [x] Grade component lifecycle — draft/publish toggle for rapor visibility, safe component editing, and backend guard against lowering `max_score` below existing student scores
- [x] Grade readiness summary — gradebook now shows readiness to print rapor based on published components and filled student values, with direct navigation into the rapor print screen when the assignment is ready
- [x] Grade bulk-save flow — guru can now accumulate multiple score/note edits within one component and save all pending row changes in one action
- [x] Grade quick-fill flow — guru can draft mass score/note fills for all rows or only still-empty rows before committing changes with bulk save
- [x] Grade finalization checkpoint — persisted finalize/reopen flow per assignment, backend readiness enforcement, and mutation lock while finalized
- [x] Grade finalization overview — gradebook now exposes a per-assignment recap for `Siap Difinalkan`, `Sudah Final`, and `Perlu Dilengkapi` so guru/admin can triage readiness across multiple kelas-mapel from one screen
- [x] Grade finalization triage flow — gradebook now adds recap filters plus a quick shortcut into the next `Siap Difinalkan` assignment so operator review can move through readiness checkpoints faster
- [x] Grade finalization ops filter/export — rekap finalisasi now supports quick search by kelas/mapel/guru and CSV export of the currently filtered readiness list for operator follow-up
- [x] Grade batch finalization flow — operator can now finalize all currently filtered `Siap Difinalkan` assignments in one guarded pass while backend readiness checks still apply per assignment
- [x] Grade teacher filter flow — rekap finalisasi now supports narrowing triage, export, and batch-finalize actions by guru without adding a new backend contract
- [x] Grade class rollup report — filtered readiness triage now also rolls up per kelas so wali kelas/operator can see how many mapel are ready, final, or still blocked before scanning detailed assignment rows
- [x] Grade homeroom focus panel — operator can now select one rolled-up kelas and immediately inspect all filtered mapel statuses for that class in a wali-kelas-friendly summary panel
- [x] Grade homeroom export flow — the focused class panel can now export a wali-kelas-friendly CSV report for the currently selected class and active filters
- [x] Grade teacher dashboard rollup — filtered readiness triage now also summarizes readiness per guru so operator can spot whose kelas-mapel sets are already ready or still blocked before drilling into detail
- [x] Grade teacher dashboard export — the filtered guru rollup can now be exported as a compact CSV summary for operator follow-up and academic coordination
- [x] Grade batch reopen flow — operator can now reopen all currently filtered finalized assignments in one guarded pass for controlled post-final corrections
- [x] Backend internal-error hygiene — 500 responses now return a generic client-safe message while raw details stay in server logs
- [x] Sensitive entrypoint rate limiting — login, refresh, public registration, and exam login now use a concurrency-safe per-IP limiter with forwarded-IP awareness
- [x] Typed auth locals — web-admin auth locals/page data now use a shared explicit auth-user type instead of `as any`
- [x] Mobile metadata cleanup — Flutter app description and Android app label no longer use default scaffold metadata
- [x] Reduce internal-key blast radius — main protected/admin backend routes now require real JWT context, and CBT asset file access is limited to real JWT users or active exam participants via `exam_token`
- [x] Mobile snapshot hardening — sensitive exam snapshot fields now live in secure storage while SharedPreferences retains only lightweight restore metadata, with legacy snapshot compatibility kept for existing installs
- [x] Mobile BYOD login hardening — device fingerprint is now documented and surfaced only as a telemetry hint, while API base URL override is moved behind an operator-facing panel instead of staying as a primary student input
- [x] Worker modularization baseline — `services/pusaka-worker` no longer keeps config, API client, parsers, Playwright flows, logging, and supervision in one `index.ts`; the entrypoint is now thin and the main concerns are split into dedicated modules
- [x] Exam shell maintainability baseline — `ExamShellScreen` now delegates connection/sync derivation and presentational support widgets into dedicated files, reducing the size and regression surface of the main screen
- [x] Runtime-data hygiene baseline — root `.gitignore` now explicitly ignores `services/core-api/data/` so local backend runtime data is less likely to leak into commits
- [x] Runtime-log hygiene baseline — `.gitignore` now also explicitly ignores `services/logs/` alongside root `logs/` so local service log output is less likely to leak into commits
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
- [x] Timetable slot foundation — migration 036 `timetable_slots` with day, time range, room, and notes per class-subject assignment
- [x] Backend timetable CRUD baseline via academic domain (`ListTimetableSlots`, create, delete)
- [x] Academic overview now returns timetable slots alongside years, classes, subjects, and assignments
- [x] Web-admin academic page now has a `Jadwal` tab for listing and adding timetable slots
- [x] Timetable edit flow and basic conflict guard — operator can update slots and backend rejects overlapping class/guru schedules on the same day/time
- [x] Timetable ops filters and weekly matrix — operator can filter by kelas/guru/search and inspect one class in a weekly matrix view
- [x] Portal timetable read baseline — dashboard guru dan siswa sekarang menampilkan jadwal mengajar / jadwal pelajaran dari slot timetable yang tersusun
- [x] Timetable class export — operator dapat mengekspor CSV dari kelas fokus pada matriks mingguan
- [x] Timetable room conflict guard — backend sekarang juga menolak bentrok ruang pada hari/jam yang sama bila room label dipakai ganda
- [x] Parent timetable visibility — dashboard orang tua sekarang juga menampilkan ringkasan jadwal setiap anak yang terhubung
- [x] Portal timetable page — route `/jadwal` sekarang memberi tampilan penuh untuk guru, siswa, dan orang tua di luar card dashboard ringkas
- [x] Portal timetable export — route `/jadwal` sekarang mendukung ekspor CSV sesuai konteks guru, siswa, dan orang tua tanpa kontrak backend baru
- [x] Portal timetable day filter — route `/jadwal` sekarang bisa difokuskan ke hari tertentu dan ekspor mengikuti filter aktif
- [x] Portal timetable guru filters — route `/jadwal` sekarang punya filter kelas dan mapel ringan untuk guru agar slot mengajar lebih cepat ditriase

### Sprint 13 — Inventory & Asset Management
- [x] Inventory foundation — migration 037 `inventory_items` dengan kode, kategori, lokasi, kondisi, satuan, jumlah total/baik, batas restok, dan catatan
- [x] Backend inventory baseline — stats + CRUD item inventaris melalui `/api/inventory/*`
- [x] Web-admin inventory baseline — dashboard `/inventory` dan master data `/inventory/items` untuk admin/staf
- [x] Inventory ops reporting — dashboard inventaris sekarang punya ringkasan per lokasi dan daftar barang mendukung ekspor CSV berdasarkan filter aktif
- [x] Inventory history baseline — perubahan create/update/delete barang sekarang tercatat dan bisa dilihat dari dialog riwayat di daftar inventaris
- [x] Inventory batch mutation baseline — operator sekarang bisa memilih banyak barang lalu memindahkan lokasi atau menyamakan kondisi sekaligus dari daftar inventaris
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

### Cross-Sprint Backlog
- [x] Publish scheduling (deferred from 16B — low priority until editorial demand is proven)
- [x] Featured content homepage widget (pull from `/api/public/site/posts/featured`)
- [ ] Flutter Student App — CBT exam client (API ready via `docs/exam-api.md`)

### ✅ Jurnal Kelas (COMPLETE)
- [x] Migration 034 — `class_journal_sessions`, `class_journal_attendances` + enum `journal_attendance_status`
- [x] sqlc queries — count/create/get/list/update/delete sesi, upsert/list kehadiran, rekap per assignment, list active students per kelas, get class-subject assignment
- [x] Service `ClassJournal` — pre-populate kehadiran 'hadir' saat sesi dibuat, ownership check via JWT `eid`, normalize status, filter assignment per guru
- [x] Handler `ClassJournal` — overview/create/get/update/delete sesi + bulk upsert kehadiran (admin/guru, delete admin-only, 409 untuk tanggal duplikat)
- [x] Routes `/api/journal/*` di `cmd/api/main.go`
- [x] BFF proxy `/api/journal`, `/api/journal/sessions`, `/api/journal/sessions/[id]`, `/api/journal/sessions/[id]/attendances`
- [x] Halaman `/journal` — assignment selector, tab Daftar Pertemuan + Rekap Kehadiran, dialog tambah pertemuan
- [x] Halaman `/journal/[id]` — detail sesi + edit, grid kehadiran siswa dengan toggle H/S/I/A per baris dan catatan
- [x] Sidebar group "Akademik" — item "Jurnal Kelas" (admin/guru) + ikon `journal`

### ✅ Sprint 17 — Tata Usaha: Persuratan Core (COMPLETE)
- [x] Migration 040 — `letter_classifications`, `incoming_letters`, `outgoing_letters`, `letter_dispositions`, `incoming_letter_sequences`, `outgoing_letter_sequences` (3 ENUMs + 6 tables)
- [x] Migration 041 — seed 31 kode klasifikasi Kemenag (PP/KP/KU/KS/HM/HK/OT/TI)
- [x] `db/queries/letters.sql` + `internal/repository/postgres/letters.sql.go` (manual generated-style)
- [x] Service `IssueOutgoingLetterNumber` — auto-number per (tahun, klasifikasi) dengan UNIQUE safety net; `IssueIncomingLetterSequence` untuk nomor agenda `AGD/{year}/{seq}`
- [x] `data/letters/.gitkeep` — storage directory for letter attachments
- [x] Lifecycle surat masuk: `baru → didisposisi → selesai → arsip`; disposisi auto-transitions to `didisposisi`
- [x] Backend: `internal/service/letter.go` + `internal/handler/letter.go` + 19 routes `/api/tu/surat/*` (admin/staf)
- [x] BFF proxy routes: `/api/tu/surat/incoming/`, `/api/tu/surat/outgoing/`, `/api/tu/surat/disposisi/`, `/api/tu/surat/klasifikasi`
- [x] Frontend pages: `/tu/surat-masuk`, `/tu/surat-keluar` (auto-number preview), `/tu/disposisi`
- [x] Sidebar group "Tata Usaha" + ikon `mail`, `inbox`, `mail-forward`

### ✅ Sprint 18 — Tata Kelola Madrasah (COMPLETE)
- [x] Governance schema foundation for organizational units, positions, active assignments, governance documents, and program indicators.
- [x] Backend `/api/governance/*` contract owned by `services/core-api` with sqlc queries, thin handlers, service validation, and `admin`/`staf` access rules.
- [x] Web-admin BFF proxies under `/api/governance/*` that forward the real user JWT and do not access PostgreSQL directly.
- [x] `/governance` operator screen for struktur organisasi, dokumen tata kelola, program/indikator, and 8 SNP evidence matrix.
- [x] Sidebar entry and route-level role guard for `admin` and `staf`.
- [x] Verification: sqlc generation, Go tests/build, Svelte autofixer, and `npm run check`.

### ✅ Sprint 19 — SKP Mirror & Cascading Kinerja (COMPLETE)
- [x] Internal performance target table linked to employees, governance positions, and governance programs.
- [x] Backend `/api/governance/performance-targets/*` contract for target list/create/update/delete with service validation.
- [x] Web-admin BFF proxy for performance targets that forwards JWT to Go API.
- [x] `/governance` SKP/Kinerja tab for target kerja pegawai, indikator, progres, evidence URL, and review notes.
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte autofixer, and `npm run check`.

### ✅ Sprint 20 — Register Bukti Mutu & 8 SNP (COMPLETE)
- [x] Dedicated evidence-item table linked to documents, programs, performance targets, units, and SNP standards.
- [x] Backend `/api/governance/evidence-items/*` contract for evidence list/create/update/delete with validation.
- [x] 8 SNP matrix and governance stats include dedicated evidence readiness counts.
- [x] Web-admin BFF proxy for evidence items.
- [x] `/governance` Bukti Mutu tab for evidence status, owner unit, linked program/target, file URL, and review notes.
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte check, and focused autofixer where available.

### ✅ Sprint 21 — Kesiswaan: Foundation (COMPLETE)
- [x] Migration extend `students` (NIK, tanggal/tempat lahir, alamat, agama, anak-ke, foto, no telp)
- [x] Tambah enum `kesiswaan` ke `user_role` + sidebar role guard
- [x] Migration `violation_categories`, `student_violations`, `student_achievements`
- [x] Auto-calc total poin pelanggaran per siswa
- [x] Form prestasi dengan tingkat sekolah/kab/prov/nasional/internasional
- [x] Upload foto siswa ke `data/student-photos/`
- [x] Sidebar group "Kesiswaan" + protected `/kesiswaan` entry
- [x] Guru read-only untuk siswa di kelasnya saja (data scoping seperti pattern grades)
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte autofixer, and `npm run check`.

### ✅ Sprint 22 — Tata Usaha: Surat Keterangan Siswa (COMPLETE)
- [x] Migration `certificate_templates`, `student_certificates` (dengan `snapshot_data` JSONB)
- [x] Seed template: aktif, lulus, pindah, kehilangan dokumen, mengikuti kegiatan
- [x] Reuse `IssueOutgoingLetterNumber` helper path via outgoing-letter sequence (klasifikasi default `PP.00.4` siswa)
- [x] Backend `/api/tu/surat-keterangan/*` contract with templates, student options, issue, detail, cancel
- [x] BFF proxy `/api/tu/surat-keterangan/*` that forwards JWT to Go API
- [x] Render HTML print route `/tu/surat-keterangan/[id]/print`
- [x] Halaman `/tu/surat-keterangan` dengan generator + history
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte autofixer, and `npm run check`.

### ✅ Sprint 23 — Kesiswaan: Engagement (COMPLETE)
- [x] Migration `extracurriculars`, `extracurricular_members`
- [x] Migration `counseling_sessions` (dengan `is_confidential BOOLEAN`)
- [x] Migration `student_transfers` (mutasi keluar/masuk dengan transaksi update `students.status`)
- [x] BK confidentiality guard: hanya `admin` & `kesiswaan` lihat catatan rahasia; guru hanya menerima catatan non-rahasia yang scoped ke kelasnya
- [x] Backend `/api/kesiswaan/*` contract for ekskul, anggota ekskul, BK, and mutasi siswa
- [x] BFF proxy routes that forward JWT to Go API without direct DB access
- [x] `/kesiswaan` tabs for Ekskul, BK, and Mutasi with operator forms and history tables
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte autofixer, and `npm run check`.

### ✅ Sprint 24 — Tata Usaha: Aset & Arsip (COMPLETE)
- [x] Migration `048_tu_archives.sql` untuk `archive_categories` dan `archive_documents` dengan kategori awal, status arsip, retensi, checksum, dan metadata file.
- [x] Backend `/api/tu/archives/*` contract untuk statistik, kategori, register dokumen, unggah file, edit metadata, hapus, dan stream file arsip.
- [x] File arsip tersimpan di `data/archives/` melalui backend Go; web-admin tetap BFF/proxy tanpa direct DB atau client persistence.
- [x] BFF proxy `/api/tu/archives/*` meneruskan JWT pengguna asli, termasuk multipart upload dan file streaming.
- [x] Halaman `/tu/arsip` dengan statistik, filter register, form unggah/edit metadata, kategori arsip, dan akses file.
- [x] Sidebar Tata Usaha menambahkan entry `Arsip` untuk `admin` dan `staf`.
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte autofixer, and `npm run check`.

### ✅ Sprint 25 — RKT/RKJM Program Execution (COMPLETE)
- [x] Migration `049_governance_work_plan_items.sql` untuk item pelaksanaan RKT/RKJM tahunan yang link ke program, dokumen sumber, unit, pegawai penanggung jawab, dan bukti mutu.
- [x] Backend `/api/governance/work-plan-items/*` contract untuk list/create/update/delete dengan validasi program, jadwal, status, progres, anggaran, dan realisasi.
- [x] Governance stats menambahkan total item RKT/RKJM, item selesai/terkendala, total anggaran, dan realisasi.
- [x] BFF proxy `/api/governance/work-plan-items/*` meneruskan JWT pengguna asli ke Go API.
- [x] Halaman `/governance` menambahkan tab `RKT/RKJM` dengan ringkasan anggaran, tabel kegiatan, export CSV, print, dan form item pelaksanaan.
- [x] Jalur dokumen → program → RKT/RKJM → SKP/bukti mutu tetap satu peta di modul tata kelola.
- [x] Verification: sqlc generation, Go tests/vet/build, Svelte autofixer, `npm run check`, and `git diff --check`.

### ✅ Sprint 26 — TU Dashboard & Compliance Print Pack (COMPLETE)
- [x] Tambahkan ringkasan dashboard TU lintas surat masuk/keluar, surat keterangan, disposisi, arsip, inventaris, dan RKT/RKJM.
- [x] Tambahkan printable compliance pack untuk kepala/staf: register surat, status disposisi, arsip/retensi, inventaris perhatian, dan bukti RKT/RKJM terpilih.
- [x] Siapkan export CSV/print yang konsisten untuk kebutuhan administrasi audit madrasah.
- [x] Sidebar Tata Usaha menambahkan entry `Dashboard TU` dan `Paket Kepatuhan` untuk `admin` dan `staf`.
- [x] Verification: Svelte autofixer reported `issues: []` for both TU pages, `npm run check`, and `git diff --check`.

### ✅ Sprint 27 — Renstra/IKU Alignment Map (COMPLETE)
- [x] Tambahkan peta Visi/Misi, Renstra/RKJM, Perkin/IKU, RKT/RKJM, SKP, dan bukti mutu pada modul tata kelola.
- [x] Tampilkan gap pemetaan: program tanpa IKU, tanpa dokumen sumber, tanpa RKT, tanpa SKP, atau tanpa bukti.
- [x] Sediakan export CSV dan print untuk bahan rapat kepala/staf serta review administrasi Kemenag/SIPKA internal.
- [x] Verification: Svelte autofixer reported `issues: []`, `npm run check`, and `git diff --check`.

### ✅ Sprint 28 — Governance Print Pack (COMPLETE)
- [x] Tambahkan paket cetak tata kelola formal: struktur organisasi, jalur komando, pejabat aktif, tupoksi, dokumen strategis, peta IKU/RKT/SKP, dan 8 SNP.
- [x] Sediakan export CSV ringkas untuk lampiran rapat/review administrasi kepala dan staf.
- [x] Hubungkan paket cetak dari halaman `/governance` tanpa menambah akses database di frontend.
- [x] Verification: Svelte autofixer reported `issues: []` and `suggestions: []` for the print-pack page, `npm run check`, and `git diff --check`.

### ✅ Sprint 29 — School Profile & Letterhead Foundation (COMPLETE)
- [x] Tambahkan API profil madrasah resmi berbasis `app_settings` untuk kop surat, identitas satuan kerja, dan tanda tangan kepala madrasah.
- [x] Tambahkan halaman admin `/settings/school-profile` untuk mengelola nama madrasah, NSM/NPSN, alamat, kontak, dan kepala madrasah.
- [x] Pakai profil resmi pada permukaan cetak utama agar tidak lagi bergantung pada placeholder kop surat.
- [x] Sidebar Sistem menambahkan entry `Profil Madrasah` untuk admin.
- [x] Verification: Go tests/vet/build, Svelte autofixer on touched components, `npm run check`, and `git diff --check`.

### ✅ Sprint 30 — Print Surface Letterhead Standardization (COMPLETE)
- [x] Tambahkan helper frontend bersama `$lib/school-profile` untuk tipe profil madrasah, default resmi, fetch via BFF `/api/school-profile`, dan format alamat kop.
- [x] Reuse helper profil pada Governance Print Pack dan TU Compliance Pack agar paket cetak memakai identitas madrasah yang sama.
- [x] Kartu ujian CBT dan berita acara sesi CBT sekarang memakai kop resmi dari profil madrasah, termasuk blok tanda tangan kepala madrasah pada berita acara.
- [x] Cetak rapor dan surat keterangan memakai nama/kop/tanda tangan kepala madrasah dari profil resmi, bukan hardcoded placeholder.
- [x] Verification: Svelte autofixer membaca komponen tersentuh dan melaporkan `issues: []` / `suggestions: []` meskipun proses MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan `git diff --no-index --check` untuk file baru/untracked.

### ✅ Sprint 31 — Governance Compliance Action Tracker (COMPLETE)
- [x] Migration `050_governance_compliance_actions.sql` menambahkan register tindak lanjut kepatuhan untuk gap Renstra/IKU/RKT/RKJM/SKP/8 SNP dengan PIC, prioritas, tenggat, status, dan bukti.
- [x] Backend `/api/governance/compliance-actions/*` contract untuk list/create/update/delete dengan validasi service, sqlc typed queries, dan statistik dashboard tata kelola.
- [x] BFF proxy `/api/governance/compliance-actions/*` meneruskan JWT pengguna asli ke Go API tanpa akses database langsung dari SvelteKit.
- [x] Halaman `/governance/actions` menyediakan ringkasan risiko, filter register, CRUD tindak lanjut, serta saran otomatis dari gap pemetaan dokumen/program/RKT/SKP/bukti mutu.
- [x] Navigasi tata kelola menambahkan akses `Tindak Lanjut` dari halaman `/governance` dan sidebar untuk `admin`/`staf`.
- [x] Verification: `make db-sqlc`, Go tests/vet/build, Svelte autofixer dengan `issues: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 32 — Governance Compliance Print & Export (COMPLETE)
- [x] Helper paket cetak tata kelola sekarang ikut memuat register tindak lanjut kepatuhan dari `/api/governance/compliance-actions`.
- [x] Paket cetak `/governance/print-pack` menambahkan metrik tindak lanjut, aksi kritis, serta tabel formal PIC/status/tenggat/bukti untuk kebutuhan rapat, audit, dan review Kemenag.
- [x] Export CSV paket tata kelola sekarang mencakup tindak lanjut kepatuhan beserta kaitan program/dokumen/SKP/bukti.
- [x] Halaman `/governance/actions` menambahkan export CSV sesuai filter aktif dan pintasan ke paket cetak formal.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 33 — Governance Compliance Escalation Board (COMPLETE)
- [x] Halaman `/governance/actions` menambahkan filter cepat untuk `Lewat Tenggat`, `Prioritas Tinggi`, `Tanpa PIC`, dan `Menunggu Bukti`.
- [x] Register tindak lanjut sekarang bisa difokuskan dari panel eskalasi tanpa mengubah kontrak backend atau akses database frontend.
- [x] Panel `Eskalasi PIC/Unit` mengurutkan beban tindak lanjut aktif berdasarkan lewat tenggat, prioritas, jumlah terbuka, dan tenggat terdekat.
- [x] Panel `Sebaran 8 SNP` menunjukkan standar yang masih punya tindak lanjut aktif beserta jumlah prioritas tinggi, lewat tenggat, dan menunggu bukti.
- [x] Export CSV tindak lanjut tetap mengikuti filter aktif sehingga hasil rapat bisa langsung diekspor sesuai fokus eskalasi.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 34 — Governance Compliance Quick Status Workflow (COMPLETE)
- [x] Register `/governance/actions` menambahkan aksi cepat status langsung dari tabel: `Mulai`, `Bukti`, `Selesai`, `Buka Ulang`, dan `Aktifkan`.
- [x] Aksi cepat memakai kontrak PUT `/api/governance/compliance-actions/{id}` melalui BFF, tidak menambah akses database frontend atau route backend baru.
- [x] Quick status memakai button loading state per baris agar rapat tindak lanjut tidak memicu submit ganda saat jaringan lambat.
- [x] Penyelesaian tanpa bukti meminta konfirmasi eksplisit supaya administrasi 8 SNP/Renstra/SKP tetap sadar bukti.
- [x] Helper form update direuse untuk modal edit dan quick status agar payload tetap konsisten.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 35 — Governance Compliance Evidence Capture (COMPLETE)
- [x] Register `/governance/actions` menambahkan aksi `Catat Bukti` untuk membuka dialog ringkas bukti/catatan tindak lanjut.
- [x] Dialog bukti dapat mengaitkan item bukti mutu, menyimpan URL/lokasi bukti, dan menulis catatan tindak lanjut tanpa membuka form lengkap.
- [x] Operator dapat memilih `Simpan Bukti` atau `Simpan & Selesai`; penyelesaian tetap mewajibkan minimal URL/lokasi bukti atau item bukti.
- [x] Mutasi bukti tetap memakai kontrak PUT `/api/governance/compliance-actions/{id}` melalui BFF dan helper payload yang sama dengan edit/quick status.
- [x] Loading state dialog mencegah submit ganda saat pencatatan bukti dilakukan di rapat atau review administrasi.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []`, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 36 — Governance Compliance Meeting Pack (COMPLETE)
- [x] Tambahkan route `/governance/actions/meeting-pack` untuk paket rapat tindak lanjut kepatuhan yang bisa dicetak dan diekspor CSV.
- [x] Paket rapat menyajikan ringkasan total/terbuka/prioritas tinggi/lewat tenggat/menunggu bukti/tanpa PIC dengan kop profil madrasah resmi.
- [x] Agenda keputusan otomatis memecah isu rapat: lewat tenggat, prioritas tinggi, validasi bukti, dan penetapan PIC kosong.
- [x] Tabel prioritas pembahasan, beban PIC/unit, sebaran 8 SNP, ruang keputusan rapat, dan tanda tangan disediakan untuk rapat kepala/staf.
- [x] Halaman `/governance/actions` menambahkan pintasan `Paket Rapat` tanpa menambah backend route atau akses database frontend.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 37 — Governance Compliance Deadline Calendar (COMPLETE)
- [x] Tambahkan route `/governance/actions/calendar` untuk kalender/timeline tenggat tindak lanjut kepatuhan.
- [x] Kalender mengelompokkan action aktif ke bucket `Lewat Tenggat`, `Hari Ini`, `7 Hari ke Depan`, `30 Hari ke Depan`, `Setelah 30 Hari`, dan `Tanpa Tenggat`.
- [x] Halaman kalender menyediakan ringkasan timeline, kartu bucket, tabel detail, print view, dan export CSV.
- [x] Halaman `/governance/actions` menambahkan pintasan `Kalender` tanpa menambah backend route atau akses database frontend.
- [x] Kalender memakai kop profil madrasah resmi dan data dari BFF governance yang sudah ada.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun salah satu proses MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 38 — Governance Control Center (COMPLETE)
- [x] Halaman utama `/governance` menampilkan statistik tindak lanjut kepatuhan: total, terbuka, kritis, dan selesai.
- [x] Control strip `Kendali Tindak Lanjut Kepatuhan` menampilkan ringkasan terbuka/kritis/selesai di dashboard tata kelola utama.
- [x] Quick links dari `/governance` menuju register tindak lanjut, kalender tenggat, paket rapat, dan paket cetak.
- [x] Statistik memakai field `GetGovernanceStats` yang sudah ada, tanpa menambah backend route atau akses database frontend.
- [x] Link internal memakai `resolve()` agar aman terhadap base path SvelteKit.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check.

### ✅ Sprint 39 — Governance Compliance Advanced Filters (COMPLETE)
- [x] Register `/governance/actions` menambahkan filter lanjutan untuk tahun, sumber tindak lanjut, standar 8 SNP, dan PIC/unit.
- [x] Filter lanjutan bekerja bersama pencarian, status, prioritas, dan filter cepat eskalasi tanpa menambah backend route atau akses database frontend.
- [x] Opsi tahun dan PIC/unit dibangun dari data register yang sudah dimuat sehingga tetap konsisten dengan data BFF governance.
- [x] Export CSV tindak lanjut tetap memakai `filteredActions`, sehingga hasil ekspor mengikuti semua filter aktif.
- [x] Tombol `Reset Filter` mengosongkan pencarian, filter lanjutan, dan fokus eskalasi dalam satu aksi.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file untracked/tersentuh.

### ✅ Sprint 40 — Governance PIC Briefing Pack (COMPLETE)
- [x] Tambahkan route `/governance/actions/owner-briefing` untuk lembar briefing tindak lanjut per PIC/unit.
- [x] Briefing mengelompokkan action aktif per PIC/unit, menampilkan prioritas, status, tenggat, bukti, kaitan program/dokumen/SKP, serta ruang arahan/paraf.
- [x] Halaman briefing menyediakan filter PIC/unit, print view, dan export CSV sesuai pilihan filter.
- [x] Ringkasan briefing menampilkan jumlah PIC/unit, tugas terbuka, prioritas tinggi, lewat tenggat, menunggu bukti, dan tanpa bukti.
- [x] Halaman `/governance/actions` menambahkan pintasan `Briefing PIC` tanpa menambah backend route atau akses database frontend.
- [x] Data briefing memakai helper `fetchGovernancePrintPackData()` dan kontrak BFF governance yang sudah ada.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []`, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 41 — Governance 8 SNP Briefing Pack (COMPLETE)
- [x] Tambahkan route `/governance/actions/snp-briefing` untuk lembar briefing tindak lanjut per standar 8 SNP.
- [x] Briefing mengelompokkan action aktif berdasarkan SNP, menampilkan PIC/unit, prioritas, status, tenggat, bukti, kaitan program/dokumen/SKP, dan ruang keputusan review.
- [x] Halaman briefing menyediakan filter standar SNP, print view, dan export CSV sesuai pilihan filter.
- [x] Ringkasan briefing menampilkan jumlah standar aktif, tugas terbuka, PIC/unit terlibat, prioritas tinggi, lewat tenggat, menunggu bukti, dan tanpa bukti.
- [x] Halaman `/governance/actions` menambahkan pintasan `Briefing SNP` tanpa menambah backend route atau akses database frontend.
- [x] Data briefing memakai helper `fetchGovernancePrintPackData()` dan kontrak BFF governance yang sudah ada.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 42 — Governance Evidence Briefing Pack (COMPLETE)
- [x] Tambahkan route `/governance/actions/evidence-briefing` untuk lembar validasi bukti tindak lanjut.
- [x] Briefing mengelompokkan action aktif ke fokus `Menunggu Bukti`, `Lewat Tenggat Tanpa Bukti`, `Prioritas Tinggi Tanpa Bukti`, `Tanpa Bukti`, dan `Bukti Tercatat`.
- [x] Halaman briefing menyediakan filter fokus bukti, print view, export CSV, kartu ringkasan, dan tabel validasi bukti.
- [x] Tabel validasi menampilkan PIC/unit, status, prioritas, tenggat, kaitan program/dokumen/SKP, bukti, dan ruang validasi.
- [x] Halaman `/governance/actions` menambahkan pintasan `Briefing Bukti` tanpa menambah backend route atau akses database frontend.
- [x] Data briefing memakai helper `fetchGovernancePrintPackData()` dan kontrak BFF governance yang sudah ada.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun salah satu proses MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, `git diff --check`, dan whitespace check untuk file baru/untracked.

### ✅ Sprint 43 — Siklus Dokumen Module (COMPLETE)
- [x] Tambahkan modul mandiri `Siklus Dokumen` dengan namespace backend `/api/document-cycles/*`, bukan sebagai submenu kecil di Governance.
- [x] Migration `051_document_cycles.sql` menambahkan `document_cycle_catalogs`, `document_cycle_obligations`, dan `document_cycle_events` untuk katalog, jadwal, pengingat, status, dan audit ringan progres dokumen.
- [x] Seed katalog awal dari siklus dokumen MTsN: harian, mingguan, bulanan, triwulan, semester, tahunan, RKJM 4 tahunan, dan Renstra 5 tahunan, termasuk SKP BKN, Perkin, IKU, RKT/RKAM, LAKIP, EDM, dan pemetaan 8 SNP.
- [x] Backend service menambahkan generator kewajiban tahunan idempotent, status workflow `not_started -> draft -> waiting_verification -> completed`, tanggal pengingat, jatuh tempo, PIC, verifikator, dan tautan ke dokumen tata kelola/evidence/arsip.
- [x] Generator kewajiban tahunan dioptimalkan menjadi bulk SQL satu round-trip agar tombol Generate Tahun tidak melewati write-timeout API saat membuat ratusan jadwal periodik.
- [x] BFF proxy `/api/document-cycles/*` meneruskan JWT pengguna asli ke Go API tanpa akses database langsung dari SvelteKit.
- [x] Halaman standalone `/document-cycles` menyediakan dashboard kepala madrasah, panel perhatian, filter pengingat/status/frekuensi, quick status, detail monitoring, dan manajemen katalog.
- [x] Sidebar menambahkan grup `Siklus Dokumen` dengan entry `Monitoring Dokumen`, serta default pin untuk role `staf`.
- [x] Migration `051_document_cycles.sql` sudah dijalankan; tabel siklus dokumen aktif dengan 30 katalog awal, termasuk RKJM dan Renstra dengan tenggat panjang.
- [x] Jadwal siklus dokumen tahun 2026 sudah digenerate: 374 kewajiban aktif, 374 status `Belum Mulai`, 101 lewat tempo, dan 10 masuk pengingat awal.
- [x] Verification: `make db-sqlc`, `make db-migrate`, `go test ./...`, `go vet ./...`, API health/generator/stats check, Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch dokumentasi eksternal, `npm run check`, dan `git diff --check`.

### ✅ Repository Hygiene Checkpoint — Source Commit Discipline (COMPLETE)
- [x] Pisahkan artefak runtime SQLite sidecar (`*.db-shm`, `*.db-wal`, `*.sqlite-shm`, `*.sqlite-wal`) dari source control sesuai aturan legacy SQLite hanya sebagai artefak impor.
- [x] Exclude lokal dokumen review `Siklus_Dokumen_MTsN_Lengkap*.docx` agar tidak ikut commit fitur.
- [x] Commit dibuat bertahap: hygiene repository, dokumentasi temuan, lalu fitur aplikasi.
- [x] Verification: staged diff dicek dengan `git diff --cached --check` dan status git dirapikan setelah commit.

---

## Architecture Decisions Log
(See previous version for details)

---

## Risk Register
(See previous version for details)
