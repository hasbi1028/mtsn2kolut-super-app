# MTs Negeri 2 Kolaka Utara — Super App Strategic Plan

> **Status:** Sprints 1-9 Complete | PUSAKA Isolation Phase 1-3 Complete | Lightweight Ops Hardening Complete | Sprint 11 In Progress | Sprint 15 Library Module Complete | Sprint 16 Public Website Foundation Complete | Sprint 16B Public Website Polish Complete | Rapor Print View Complete | Jurnal Kelas Complete | Sprint 17 Persuratan Complete | Sprint 18 Tata Kelola Madrasah Complete | Sprint 19 SKP Mirror Complete | Sprint 20 Bukti Mutu Complete | Sprint 21 Kesiswaan Foundation Complete | Sprint 22 Surat Keterangan Complete | Sprint 23 Kesiswaan Engagement Complete | Sprint 24 Arsip TU Complete | Sprint 25 RKT/RKJM Execution Complete | Sprint 26 TU Dashboard Complete | Sprint 27 Renstra/IKU Alignment Complete | Sprint 28 Governance Print Pack Complete | Sprint 29 School Profile Complete | Sprint 30 Print Surface Letterhead Complete | Sprint 31 Compliance Actions Complete | Sprint 32 Compliance Print/Export Complete | Sprint 33 Compliance Escalation Board Complete | Sprint 34 Compliance Quick Status Complete | Sprint 35 Compliance Evidence Capture Complete | Sprint 36 Compliance Meeting Pack Complete | Sprint 37 Compliance Deadline Calendar Complete | Sprint 38 Governance Control Center Complete | Sprint 39 Compliance Advanced Filters Complete | Sprint 40 PIC Briefing Pack Complete | Sprint 41 8 SNP Briefing Pack Complete | Sprint 42 Evidence Briefing Pack Complete | Sprint 43 Siklus Dokumen Module Complete | Sprint 44 Dokumen Integration Control Center Complete | Sprint 45 Cross-Module Document Shortcuts Complete | Sprint 46 Document Cycle Audit Timeline Complete | Sprint 47 Document Cycle Audit Test Coverage Complete | Sprint 48 BFF Staff Operations Gate Complete | Sprint 49 Backend Staff Gate Claim Hardening Complete | Sprint 50 Document Cycle Completion Readiness Complete | Sprint 51 Final Archive Requirement Complete | Sprint 52 Document Cycle Status Transition Lock Complete | Sprint 53 Document Cycle Detail Status Actions Complete | Sprint 54 Document Cycle Traceability Detail Complete | Sprint 55 External Compliance Checklist Complete | Sprint 56 External Checklist CSV Export Complete | Sprint 57 SvelteKit BFF Async Boundary Hardening Complete | Sprint 58 CBT Legacy Proctoring & Bank Soal Migration Complete | Sprint 59 CBT Soal Legacy Full Modal Complete | Sprint 60 Unified CBT Question Bank Complete | Sprint 61 CBT Essay Authoring & Manual Scoring Complete | Sprint 62A CBT Questions UI Retirement Complete | Sprint 62B CBT Draft Review Flow Complete | Sprint 62C CBT Type-Aware Composer Complete | Sprint 62D CBT Short Answer Scoring & Mobile Renderer Complete | Sprint 62E CBT Agree/Disagree & Template Retirement Complete | Sprint 62F CBT Matching Authoring & Scoring Complete | Sprint 62G CBT Matching Distractors Complete | Sprint 62H CBT Multi-Type CSV Import Complete | Sprint 62I CBT Confirm Dialog Layer Hotfix Complete | Sprint 62J CBT Multi-Type CSV Export Complete | Sprint 62K CBT CSV Template & Roundtrip Hardening Complete | Sprint 62L CBT Non-Test Assessment Foundation Complete | Sprint 62M CBT Non-Test Roster & Scoring Complete | Sprint 62N CBT Non-Test Grade/Rapor Sync Complete | Sprint 62O CBT Non-Test Grade Traceability Complete | Sprint 62P CBT Non-Test Grade Source Lock Complete | Sprint 62Q CBT Non-Test Sync Freshness Complete | Sprint 62R CBT Non-Test Sync Filter Complete | Sprint 62 CBT Juknis Question Types Complete | Sprint 63 CBT Ruangan & Pengawas Operasional Complete | Sprint 64 CBT Ruang Saya Pengawas Complete | Sprint 65 CBT Room Handover Complete | Sprint 66 CBT Session Ops Recap Complete | Sprint 67 CBT Package Blueprint Coverage Complete | Sprint 68 CBT Item Analysis Complete | Sprint 69 CBT Item Revision Loop Complete | Sprint 70 CBT Revision Queue Complete | Sprint 71 CBT Revision Trace Complete | Sprint 72 CBT Revision Source Filter Complete | Sprint 73 CBT Revision Resubmit Complete | Sprint 74 CBT Reviewer Queue Complete | Sprint 75 CBT Approved Publish Queue Complete | Sprint 76 CBT Package Published Guard Complete | Sprint 77 CBT Package Quality Summary Complete | Sprint 78 CBT Session Package Quality Gate Complete | Last Updated: 2026-05-03
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

### ✅ Sprint 44 — Dokumen Integration Control Center (COMPLETE)
- [x] Tutup temuan akses Library/Inventory dengan role gate backend handler dan BFF hook untuk `admin`/`staf`.
- [x] Tambahkan recovery job PUSAKA `running` yang stale ke flow `failed`/retry agar unique active-job guard tidak memblokir pegawai permanen setelah worker crash.
- [x] Scope file foto siswa Kesiswaan mengikuti scope data siswa; guru hanya bisa membaca foto siswa di kelas ajarnya.
- [x] Migration `052_document_cycle_integrations.sql` menambahkan bidang dokumen, tracker sistem eksternal, link RKT/RKJM/RKAM, SKP/target kinerja, compliance action, evidence, dan arsip ke siklus dokumen.
- [x] Katalog siklus dokumen dipetakan ke bidang TU, Kesiswaan, Kurikulum, Sarpras, Governance, Keuangan, dan Eksternal; tracker SKP BKN/e-Kinerja, EMIS, SIPKA, SIMAK-BMN, RKAM/BOS, Perkin, IKU, LAKIP/LKj, dan EDM tercatat sebagai checklist internal, bukan pengganti portal resmi.
- [x] `/document-cycles` diperkuat menjadi radar dokumen kepala madrasah dengan filter bidang/tracker eksternal, panel perhatian, tracker kepatuhan eksternal, quick action bukti/arsip/verifikasi, dan tab `Peta Keterhubungan`.
- [x] Peta keterhubungan menampilkan jejak dokumen ke arsip TU, evidence 8 SNP, dokumen tata kelola, RKT/RKJM/RKAM, SKP, dan compliance action.

### ✅ Sprint 45 — Cross-Module Document Shortcuts (COMPLETE)
- [x] `/document-cycles` membaca deep link query untuk `period_year`, `status`, `frequency`, `domain_area`, `external_system`, `tab`, `search`, dan `reminder_only`.
- [x] Dashboard TU menambahkan pintasan ke siklus dokumen bidang TU.
- [x] Dashboard Governance menambahkan pintasan ke siklus dokumen bidang Governance langsung ke tab `Peta Keterhubungan`.
- [x] Kesiswaan menambahkan pintasan domain Kesiswaan untuk role yang memang boleh membuka pusat siklus dokumen.
- [x] Akademik/Kurikulum menambahkan pintasan domain Kurikulum.
- [x] Inventory/Sarpras menambahkan pintasan domain Sarpras dengan filter tracker SIMAK-BMN.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` / `suggestions: []` meskipun proses MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 46 — Document Cycle Audit Timeline (COMPLETE)
- [x] Tambahkan query `ListDocumentCycleEventsByObligation` untuk membaca riwayat event siklus dokumen beserta username aktor.
- [x] Status-change event sekarang menyimpan `from_status` dan `to_status`, bukan hanya status tujuan.
- [x] Backend expose `GET /api/document-cycles/obligations/{id}/events` dengan role gate yang sama seperti modul Governance/Siklus Dokumen.
- [x] BFF menambahkan proxy `/api/document-cycles/obligations/[id]/events` tanpa akses database langsung dari SvelteKit.
- [x] Panel detail `/document-cycles` menampilkan `Riwayat Audit` dengan skeleton, retry, empty state, waktu WITA, aktor, catatan, dan transisi status.
- [x] Verification: `make db-sqlc`, `go test ./...`, Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 47 — Document Cycle Audit Test Coverage (COMPLETE)
- [x] Tambahkan unit test service untuk memastikan perubahan status siklus dokumen membaca status lama sebelum update.
- [x] Test mengunci payload event audit `status_changed`, termasuk `from_status`, `to_status`, catatan yang sudah dinormalisasi, dan `actor_user_id`.
- [x] Test memastikan status update berhenti bila lookup status lama gagal, sehingga event audit tidak dibuat dari data transisi yang tidak diketahui.
- [x] Test memastikan `ListEvents` meneruskan obligation ID ke store dan mengembalikan event timeline.
- [x] Verification: `go test ./...`, `git diff --check`, dan root `make check`.

### ✅ Sprint 48 — BFF Staff Operations Gate (COMPLETE)
- [x] BFF `hooks.server.ts` sekarang eksplisit menutup `/document-cycles` dan `/api/document-cycles` untuk role `admin`/`staf`.
- [x] Namespace Governance dan TU juga masuk gate operasi staf di BFF (`/governance`, `/api/governance`, `/tu`, `/api/tu`) agar konsisten dengan page guard dan backend role gate.
- [x] Role resolver hook memakai fallback `role` tunggal ketika `roles` array tidak ada, konsisten dengan page-level guards yang sudah ada.
- [x] Verification: `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 49 — Backend Staff Gate Claim Hardening (COMPLETE)
- [x] Helper backend `libraryAccessAllowed`, `inventoryAccessAllowed`, `governanceAccessAllowed`, dan `tuAccessAllowed` sekarang menolak request tanpa JWT claims, bukan fallback allow.
- [x] Test akses staf diperluas untuk mengunci `admin`/`staf` diterima, `guru` ditolak, dan request tanpa claims ditolak di Library, Inventory, Governance, dan TU.
- [x] Hardening ini menjaga Document Cycles, Governance, TU, Library, dan Inventory tetap eksplisit bergantung pada JWT context meskipun route saat ini sudah berada dalam JWT middleware group.
- [x] Verification: `go test ./...`, `git diff --check`, dan root `make check`.

### ✅ Sprint 50 — Document Cycle Completion Readiness (COMPLETE)
- [x] Tambahkan query readiness finalisasi siklus dokumen untuk mengecek PIC, verifikator, dan tautan bukti utama.
- [x] Service `UpdateObligationStatus` sekarang menolak status `completed` bila dokumen belum punya PIC penyusun, verifikator, dan minimal satu tautan dokumen/evidence/tindak lanjut/arsip.
- [x] Test service mengunci finalisasi lengkap berhasil, finalisasi tidak lengkap ditolak tanpa update/event audit, dan status selain `completed` tidak memaksa readiness finalisasi.
- [x] UI `/document-cycles` menampilkan readiness finalisasi di panel detail dan menonaktifkan tombol `Selesai` pada baris yang belum lengkap.
- [x] Verification: `make db-sqlc`, `go test ./...`, Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 51 — Final Archive Requirement (COMPLETE)
- [x] Finalisasi siklus dokumen sekarang wajib memiliki `archive_document_id`, bukan cukup dengan evidence/dokumen tata kelola saja.
- [x] Service readiness error dan UI readiness copy disederhanakan ke syarat final utama: PIC penyusun, verifikator, dan arsip digital.
- [x] Test service menambahkan kasus evidence tanpa arsip tetap ditolak untuk status `completed`.
- [x] Verification: `go test ./...`, Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 52 — Document Cycle Status Transition Lock (COMPLETE)
- [x] Backend service mengunci alur status siklus dokumen: `Belum Mulai -> Draft -> Menunggu Verifikasi -> Selesai`, dengan koreksi dari `Menunggu Verifikasi/Selesai` kembali ke `Draft`.
- [x] Finalisasi `completed` hanya bisa dilakukan dari `waiting_verification`, sehingga loncatan `draft -> completed` ditolak sebelum readiness/arsip dicek.
- [x] Test service mengunci invalid transition tidak melakukan readiness lookup, update status, atau event audit.
- [x] UI `/document-cycles` menonaktifkan tombol status yang tidak valid untuk status saat ini.
- [x] Verification: `go test ./...`, Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 53 — Document Cycle Detail Status Actions (COMPLETE)
- [x] Panel detail `/document-cycles` sekarang menampilkan aksi cepat status sesuai alur backend yang terkunci.
- [x] Aksi `Mulai Draft`, `Koreksi Draft`, `Kembalikan Belum Mulai`, `Ajukan Verifikasi`, dan `Tandai Selesai` hanya muncul pada status yang relevan.
- [x] Aksi finalisasi di detail ikut memakai readiness PIC, verifikator, dan arsip digital sehingga tidak menawarkan finalisasi yang belum siap.
- [x] Shortcut `Tambah Bukti` dan `Tautkan Arsip` tetap tersedia sebagai jalur operasional untuk melengkapi syarat finalisasi.
- [x] Verification: Svelte autofixer membaca file dengan `issues: []` / `suggestions: []`, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 54 — Document Cycle Traceability Detail (COMPLETE)
- [x] `/document-cycles` mendukung deep link `selected_obligation` untuk membuka detail kewajiban dokumen tertentu setelah data dashboard dimuat.
- [x] Tombol `Detail` dari tabel monitoring, panel perhatian, dan tab `Peta Keterhubungan` memilih dokumen yang sama, memperbarui URL, lalu memfokuskan panel detail.
- [x] Panel detail menambahkan ringkasan jejak dokumen: status, periode, PIC, verifikator, SNP/regulasi, arsip resmi, dan catatan verifikasi.
- [x] Peta keterhubungan sekarang punya aksi langsung menuju detail sehingga kepala madrasah bisa menelusuri bukti dan arsip tanpa kembali ke tabel utama.
- [x] Verification: Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 55 — External Compliance Checklist (COMPLETE)
- [x] `/document-cycles` menambahkan tab `Kepatuhan Eksternal` sebagai checklist internal untuk SKP, EMIS, SIPKA, SIMAK-BMN, RKAM/BOS, Perkin, IKU, LAKIP/LKj, dan EDM.
- [x] Checklist menampilkan status internal `Belum input`, `Sedang proses`, `Sudah input`, `Perlu revisi`, dan `Selesai` yang diturunkan dari status siklus dokumen, keterlambatan, serta bukti tertaut.
- [x] Setiap baris checklist menampilkan portal, dokumen/periode, PIC, tenggat, evidence, arsip, tindak lanjut, dan aksi detail ke panel monitoring yang sama.
- [x] Panel tracker eksternal kini punya tombol `Checklist` untuk fokus ke portal tertentu melalui filter tracker yang sudah ada.
- [x] Verification: Svelte autofixer membaca file dengan `issues: []` / `suggestions: []` meskipun MCP exit 1 karena timeout fetch eksternal, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 56 — External Checklist CSV Export (COMPLETE)
- [x] Tab `Kepatuhan Eksternal` menambahkan ekspor CSV untuk rekap rapat/audit internal tanpa menggantikan portal resmi eksternal.
- [x] CSV memuat portal, status checklist, status siklus dokumen, kode/nama dokumen, periode, bidang, SNP, regulasi, PIC, verifikator, tenggat, evidence, arsip, tindak lanjut, dan catatan verifikasi.
- [x] Export memakai data yang sedang aktif di dashboard/filter, sehingga tombol checklist portal dan filter tracker menghasilkan rekap sesuai konteks yang dilihat kepala madrasah.
- [x] Verification: Svelte autofixer membaca file dengan `issues: []` / `suggestions: []`, `npm run check`, `git diff --check`, dan root `make check`.

### ✅ Sprint 57 — SvelteKit BFF Async Boundary Hardening (COMPLETE)
- [x] Helper BFF `proxy(event)` sekarang memakai `event.fetch`, sehingga request ke Go API ikut jalur `handleFetch` dan bisa retry setelah refresh token, bukan bypass lewat `global fetch`.
- [x] Authenticated proxy memaksa JWT pengguna asli dan membuang header kontrol internal dari `proxy.fetch`, menjaga `X-Internal-Key` tidak menjadi fallback diam-diam untuk trafik BFF terautentikasi.
- [x] Route BFF yang perlu streaming/form-data atau normalisasi khusus (`PUSAKA jobs/attendance`, `CBT assets`, dan `Website media`) dipindahkan ke `proxy(event).fetch` agar tetap melewati auth forwarding dan refresh boundary.
- [x] Retry refresh di `handleFetch` sekarang memakai clone request pra-fetch untuk menjaga body POST/FormData tetap bisa dikirim ulang, sekaligus memperbarui `event.locals.accessToken` dan `event.locals.user` setelah token rotate.
- [x] Retry refresh dibatasi hanya untuk request upstream Core API yang membawa `Authorization: Bearer ...`; request publik, non-auth, non-bearer, route lokal, atau fetch eksternal tidak lagi memicu refresh token maupun clone body yang tidak perlu.
- [x] Route logout BFF memakai helper public post berbasis `event.fetch`, sehingga revoke refresh session tetap berada di boundary request SvelteKit dan tidak bypass hook fetch server.
- [x] Public page load dan login action memakai helper API berbasis `event.fetch`, sehingga request SSR public site dan auth login tidak lagi bergantung pada global fetch di luar boundary SvelteKit.
- [x] Refresh token di `hooks.server.ts` memakai fetcher eksplisit dari `handle`/`handleFetch`, sehingga rotasi token tidak lagi bergantung pada `globalThis.fetch` dan tetap testable di boundary request.
- [x] Refresh token di `handleFetch` memakai single-flight per `event.locals`, sehingga request paralel dari `Promise.all` tidak berebut merotasi refresh token yang sama.
- [x] Auth gate SvelteKit tidak lagi membentuk `locals.user` dari access token yang sudah expired saat refresh token tidak ada/gagal.
- [x] Helper proxy membaca access token terbaru dari `event.locals` pada setiap call, termasuk `proxy.fetch`, sehingga request lanjutan setelah token rotate tidak memakai JWT lama.
- [x] Helper server legacy yang memakai `global fetch` (`apiGet/apiPost/apiPublicGet/apiLogin/apiRefresh`) dihapus dari export aktif; route dan page server harus memakai `event.fetch`, `api*WithFetch`, atau `proxy(event)` agar tetap berada di boundary SvelteKit.
- [x] Helper `internalHeaders` yang tidak lagi dipakai dihapus dari BFF supaya `X-Internal-Key` tidak tersisa sebagai pola akses alternatif untuk proxy terautentikasi.
- [x] Parsing envelope API dibuat aman untuk response kosong/non-JSON, status upstream tetap dipertahankan, dan detail 5xx backend dimasking dengan pesan generik di response browser.
- [x] Error parsing `request.json()` di route BFF sekarang dipetakan sebagai `400 Payload JSON tidak valid`, bukan jatuh ke fallback `503 Backend tidak dapat dihubungi`.
- [x] Helper typed `readRequestJson<T>()` ditambahkan untuk route BFF yang membaca body JSON; semua route server di `apps/web-admin/src/routes` yang memproses JSON body sekarang melewati helper bersama agar parsing request konsisten dan testable.
- [x] Helper `readOptionalRequestJson<T>()` ditambahkan untuk route mutasi yang menerima body kosong secara sah, sehingga body kosong tetap valid tetapi JSON rusak tidak lagi diabaikan.
- [x] Route upload/file BFF memakai helper `readProxyJson` bersama untuk error parsing dan masking, termasuk upload foto siswa dan arsip TU.
- [x] Route JSON passthrough/form-data BFF sekarang memakai helper `jsonProxyResponse` bersama untuk response kosong/204, status sukses, mapping `data`, dan error masking yang konsisten.
- [x] File/image proxy memakai helper `streamProxyResponse` bersama agar header file aman diteruskan, fallback content-type/cache-control konsisten, dan error upstream tetap melewati masking `handleRouteError`.
- [x] Test route-level mengunci file proxy CBT agar tetap memakai `streamProxyResponse`, bukan kembali ke pass-through response kosong.
- [x] Alias legacy PUSAKA `/api/queue/stats` di web-admin dihapus; dashboard dan sidebar attention memakai route canonical `/api/pusaka/jobs/stats` dengan test pengunci.
- [x] Sidebar shell menangani kegagalan sinkronisasi preferensi sebagai best-effort dan logout tetap mengarahkan pengguna ke login walaupun revoke request gagal.
- [x] Sidebar pin/recent cache dipindahkan ke event-driven persistence; `pinnedLoaded` dibuat reactive agar cache localStorage benar-benar aktif setelah bootstrap.
- [x] Halaman settings memuat sesi auth dan konfigurasi PUSAKA secara paralel untuk admin, dan self-revoke session tetap redirect ke login walaupun logout best-effort gagal.
- [x] Dashboard Kesiswaan memproses fetch dan parsing data paralel dengan fallback non-JSON yang terkendali, sehingga error proxy tidak berubah menjadi SyntaxError mentah di UI.
- [x] Helper client `readClientJson` ditambahkan dengan unit test untuk response sukses, error/message backend, 5xx non-JSON, dan body kosong/malformed agar halaman baru tidak membuat parser lokal yang sulit dikunci.
- [x] Helper client `readClientJson` mendukung `204 No Content` sebagai payload `null`, lalu dashboard `/document-cycles` dipindahkan dari parser lokal ke helper bersama agar aksi DELETE dan mutasi tetap konsisten.
- [x] Helper client `readClientApiData` ditambahkan untuk membuka envelope `{ data }`, menolak envelope sukses yang membawa `error`, dan memberi fallback jelas ketika payload wajib tidak tersedia.
- [x] Helper client `readClientApiData` tetap membuka envelope `{ items }` sederhana, tetapi mempertahankan payload paginated `{ items, meta }` agar bank soal dan komposer soal tidak kehilangan metadata pagination.
- [x] Halaman Notifikasi, Antrian Verifikasi Dokumen, dan badge perhatian Sidebar memakai parsing response aman untuk error/non-JSON, sehingga fallback `AsyncContent` maupun badge shell tidak jatuh karena `SyntaxError` mentah.
- [x] Mutasi halaman Settings untuk password, sesi auth, dan konfigurasi PUSAKA memakai `readClientJson`, sehingga response error/non-JSON memakai jalur error yang sama dengan halaman BFF baru.
- [x] Load awal halaman Settings untuk sesi auth, konfigurasi worker, dan jadwal PUSAKA memakai `readClientApiData`, sehingga parser lokal envelope tidak lagi tersebar di permukaan settings.
- [x] Refresh halaman Notifikasi dan Antrian Verifikasi Dokumen memakai request id guard agar request lama tidak mematikan indikator refresh milik request terbaru saat pengguna menekan refresh/filter berulang.
- [x] Antrian PUSAKA memakai `readClientJson` dan request id guard untuk load/filter/polling, sehingga response filter lama tidak bisa menimpa daftar job filter terbaru.
- [x] Dashboard PUSAKA utama memakai `readClientApiData` untuk overview/action dan request id guard untuk polling, sehingga response lama tidak menimpa ringkasan worker/antrian terbaru.
- [x] Halaman Pegawai PUSAKA memakai `readClientApiData` untuk daftar/action run-now dan request id guard untuk polling 4 detik, sehingga data pegawai dari request lama tidak menimpa state terbaru.
- [x] Laporan Ringkasan Kehadiran dan Data Kehadiran PUSAKA memakai helper client bersama serta request id guard, sehingga perubahan rentang tanggal cepat tidak bisa ditimpa response lama.
- [x] Komponen `EmployeeList` PUSAKA memakai helper client bersama untuk test kredensial, audit, jadwal, stop job, dan mutasi akun; aksi tanpa payload tetap menerima response kosong/null secara sah.
- [x] Komponen `ScheduleList` PUSAKA Settings memakai `readClientJson` untuk tambah/hapus/simpan jadwal rekap, sehingga error/non-JSON response mengikuti jalur helper shared.
- [x] Komponen master pegawai umum (`EmployeeForm` dan `GeneralEmployeeList`) memakai `readClientJson` untuk tambah/edit/status/hapus pegawai, sehingga mutasi admin tidak lagi punya parser error lokal yang berbeda.
- [x] Komponen `WebsiteContentManager` memakai helper client bersama untuk daftar konten, upload media, dan mutasi editorial, plus request id guard agar refresh lama tidak menimpa daftar website terbaru.
- [x] Halaman Library dan Inventory memakai helper client bersama untuk overview, daftar, mutasi, dan riwayat, plus request id guard agar refresh lama tidak menimpa katalog, sirkulasi, atau inventaris terbaru.
- [x] Dashboard Kesiswaan memakai helper client bersama untuk statistik/profil/aktivitas, upload foto, simpan data, dan hapus data, plus request id guard agar load lama tidak menimpa snapshot siswa terbaru.
- [x] Halaman TU/Persuratan dan Arsip memakai helper client bersama untuk surat masuk/keluar, disposisi, surat keterangan, arsip digital, upload form-data, dan mutasi hapus/batal/status, plus request id guard untuk filter/refresh dan dialog disposisi.
- [x] Halaman Governance dan Compliance Actions memakai helper client bersama untuk register tata kelola, bukti mutu, tindak lanjut, quick status, dan evidence capture, plus request id guard agar refresh lama tidak menimpa overview aktif.
- [x] Halaman dashboard utama, PPDB awal, master pegawai, orang tua, jadwal portal, audit logs, users, dan profil madrasah memakai helper client bersama serta request id guard pada daftar/refresh yang bisa dipicu berulang.
- [x] Halaman Students, Academic, dan Journal memakai helper client bersama untuk load, simpan, hapus, lifecycle, jadwal, sesi jurnal, dan kehadiran, plus request id guard untuk overview/detail yang bisa berpindah cepat.
- [x] Halaman Grades dan Cetak Rapor memakai helper client bersama untuk overview, rapor terbit, komponen nilai, entri nilai, finalisasi, dan request id guard agar pilihan kelas-mapel cepat tidak ditimpa response lama.
- [x] Halaman CBT (`events`, `packages`, `sessions`, detail sesi, print kartu/berita acara, bank soal, dan komposer soal) memakai helper client bersama untuk envelope, mutasi, upload asset, dan request id guard pada list/detail/polling.
- [x] Sisa parser lokal di Compliance Actions, Data Kehadiran PUSAKA, dan sidebar attention dipindahkan ke helper client bersama agar pencarian global hanya menyisakan parser JSON di `src/lib/client/api.ts`.
- [x] Fallback editor lokal berbasis `contenteditable`/`execCommand` dihapus; `EditorWrapper` sekarang menstandarkan rich-text ke TipTap agar boundary Svelte, aksesibilitas tombol toolbar, dan upload media melewati satu jalur komponen.
- [x] Ekstraksi teks polos komposer soal dipindahkan ke utility `htmlToPlainText` yang unit-tested, dan renderer rich math tidak lagi melakukan assignment `innerHTML` mentah saat memasukkan hasil KaTeX inline.
- [x] Autosave komposer soal memakai signature/kunci draft turunan sebelum `$effect` menjadwalkan penyimpanan, sehingga effect tidak lagi membangun payload dan menulis status state langsung di badan reactive effect.
- [x] Helper BFF `apiPath` dan `requiredRouteParam` ditambahkan untuk validasi/encoding parameter route sebagai segmen path upstream; namespace proxy CBT yang memakai param dinamis sudah dipindahkan ke helper ini dengan unit test pengunci.
- [x] Route proxy Kesiswaan dan TU yang memakai param dinamis (`[id]`, `[filename]`) dipindahkan ke `apiPath`/`requiredRouteParam`, termasuk BK, ekskul, prestasi, pelanggaran, profil siswa, arsip, surat masuk/keluar, disposisi, dan surat keterangan.
- [x] Route proxy strategis untuk Notifikasi, Siklus Dokumen, dan Governance juga memakai `apiPath`/`requiredRouteParam`, sehingga event timeline, status dokumen, katalog, evidence, program, unit, assignment, dan compliance action tidak menyisipkan param mentah ke URL upstream.
- [x] Sisa route proxy administratif (`auth sessions`, `users`, `parents`, `website content`, `library`, dan `inventory`) ikut dipindahkan ke `apiPath`/`requiredRouteParam`, sehingga pencarian global tidak lagi menemukan interpolasi `event.params` mentah di `apps/web-admin/src/routes/api`.
- [x] Helper BFF `apiPathWithQuery` ditambahkan untuk forwarding query string yang konsisten; route list di Siklus Dokumen, Governance, Kesiswaan, TU, Library, Inventory, CBT, Grades, Journal, dan PUSAKA tidak lagi merakit `?${params}` manual atau mengirim marker query kosong.
- [x] Route proxy lama yang mengambil `id`/`entity` dari query (`academic`, `students`, dan delete CBT legacy) ikut memakai `requiredRouteParam` + `apiPath`, sehingga dynamic path upstream tidak lagi dibentuk dari query mentah.
- [x] Sisa proxy pegawai, bank soal CBT, nilai, jurnal, dan PUSAKA yang masih memakai template URL mentah dipindahkan ke `apiPath`; audit users, website content, asset download, dan jobs PUSAKA memakai `apiPathWithQuery` agar query kosong tidak bocor menjadi `?`.
- [x] Helper client `clientApiPath` dan `clientApiPathWithQuery` ditambahkan agar halaman Svelte tidak lagi merakit path/query BFF secara ad hoc; Academic, Students, CBT, Grades, dan Journal mulai memakai helper ini untuk ID/query dinamis.
- [x] Route BFF `/api/students` menambahkan handler `PUT?id=...` yang meneruskan update siswa ke backend dengan ID ter-encode, menutup mismatch lama antara halaman Students dan proxy route yang hanya punya `GET/POST/DELETE/PATCH`.
- [x] Helper server BFF sekarang menolak response upstream sukses yang kosong, bukan JSON, atau envelope tanpa `data`, sehingga page/server load tidak menerima `undefined` seolah sukses; route proxy JSON ikut memetakan payload sukses non-JSON ke error terkontrol.
- [x] Shell `Sidebar` memakai helper client bersama untuk preferensi remote dan guard versi lokal, sehingga response preferensi lama tidak menimpa pin/recent yang baru diubah pengguna.
- [x] `AsyncContent` punya default `RecoveryPanel` ketika halaman lupa menyediakan snippet gagal, sehingga promise reject dan render-time boundary error tetap punya fallback terkendali.
- [x] `/document-cycles` memisahkan data kritikal dan referensi opsional; kegagalan opsi governance/TU tidak lagi menjatuhkan seluruh dashboard radar dokumen.
- [x] Refresh dashboard `/document-cycles` memakai request id agar response lama tidak meng-overwrite state atau mematikan indikator refresh request terbaru.
- [x] Helper fetch halaman `/document-cycles` sekarang aman untuk response kosong/204 dan tidak membaca body JSON dua kali, sehingga action DELETE tidak menghasilkan error UI palsu.
- [x] Background refresh antrian PUSAKA tidak lagi mengosongkan data yang sudah tampil saat polling gagal; error penuh hanya muncul untuk load/filter aktif yang memang gagal.
- [x] Route restart worker PUSAKA diperkeras dengan guard admin handler-level, eksekusi `execFile` tanpa shell bebas, timeout, log audit ringan, dan pesan error generik.
- [x] Unit test BFF ditambahkan untuk auth forwarding, refresh retry `handleFetch`, JSON body helper wajib/opsional, non-JSON error, 5xx masking, fallback `AsyncContent`, dan hardening restart worker.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []` pada blocker yang diperiksa, `npm run test:unit`, `npm run check`, dan `git diff --check`.

### ✅ Sprint 58 — CBT Legacy Proctoring & Bank Soal Migration (COMPLETE)
- [x] Ambil perilaku proctoring terbaik dari aplikasi CBT lama tanpa membawa stack SQLite/Drizzle ke runtime monorepo.
- [x] Perkuat `/cbt/sessions/[id]` sebagai pusat proctoring: event log peserta, reset akses perangkat, force submit, flag pengawas, heartbeat, dan ringkasan jawaban.
- [x] Tambahkan backend proctor-control endpoints untuk log event sesi, reset akses peserta, dan force-submit dengan audit event dan guard guru/admin sesi.
- [x] Perkuat `/cbt/soal` sebagai komposer utama: import CSV legacy, template, readiness, quality signals, RTL, rich preview, duplicate-as-draft, dan workflow tetap memakai kontrak Go API.
- [x] Tambahkan import CSV legacy di Go API dengan parsing delimiter comma/semicolon, mapping gambar legacy, deteksi duplikat dalam file, ringkasan error, dan test service pengunci.
- [x] Semua perubahan backend tetap lewat `services/core-api`, SQL eksplisit di `db/queries`, regenerate `sqlc`, dan SvelteKit hanya BFF/proxy.
- [x] Verification: `make db-sqlc`, `go test ./...`, `go vet ./...`, Svelte autofixer untuk file tersentuh, `npm run check:web`, dan `git diff --check`.

### ✅ Sprint 59 — CBT Soal Legacy Full Modal (COMPLETE)
- [x] Tambahkan komponen `LegacyRichTextEditor` berbasis Quill dari aplikasi CBT lama dengan toolbar heading, align, bold/italic/underline, formula KaTeX, image upload, list, clean, dan kontrol ukuran gambar.
- [x] Route `/cbt/soal` memakai satu modal penuh ala legacy untuk create/edit: header, template cepat, metadata, rich text stem, kartu opsi A-D, kunci jawaban, preview, readiness, dan footer aksi dalam satu kotak besar.
- [x] Stem dan seluruh opsi A-D memakai rich text legacy, bukan textarea/TipTap untuk komposer utama.
- [x] Upload gambar editor tetap melalui callback asset CBT monorepo (`/api/cbt/assets`) dan tidak memakai endpoint legacy `/api/upload`.
- [x] Payload simpan soal menjaga kontrak backend: plain text tetap dikirim untuk pencarian/kolom legacy, HTML rich text masuk ke `stem_html` dan `options[].html`.
- [x] Verification: Svelte autofixer membaca file tersentuh dengan `issues: []`, dan `npm run check:web`.

### ✅ Sprint 60 — Unified CBT Question Bank (COMPLETE)
- [x] `/cbt/soal` dipatenkan sebagai pintu utama Bank Soal CBT dengan mode Katalog, Komposer Cepat, Editor Lanjutan, Review, dan Import Legacy dalam satu modul.
- [x] `/cbt/questions` sempat dipertahankan sebagai bridge editor lanjutan/eksperimen, lalu dipensiunkan pada Sprint 62A setelah arah produk diputuskan menjadi satu UI Bank Soal.
- [x] Sidebar CBT disederhanakan menjadi satu menu "Bank Soal" menuju `/cbt/soal`; pin default guru ikut dipindahkan dari `/cbt/questions` ke `/cbt/soal`.
- [x] Modal komposer `/cbt/soal` pernah merapikan Template Cepat menjadi segmented toolbar satu baris; fitur ini kemudian dipensiunkan pada Sprint 62E agar alur komposer lebih langsung dan ringkas.
- [x] Status draft otomatis di modal komposer dipindahkan hanya ke footer sticky, sehingga tidak ada banner status besar yang mengambil ruang konten.
- [x] Modal komposer `/cbt/soal` dipadatkan lagi: header satu baris, metadata dense row, card radius/padding lebih kecil, dan footer sticky lebih pendek.
- [x] Rich text legacy bisa di-resize vertikal; editor opsi A-D memakai mode compact dengan default tinggi lebih kecil.
- [x] Editor rich text utama `Isi Pertanyaan` juga memakai mode compact dengan default tinggi sekitar setengah dari sebelumnya.
- [x] Uji rekomendasi layout opsi dua kolom di desktop untuk mengurangi scroll panjang pada modal komposer.
- [x] Uji paket UI/UX komposer lengkap: status strip kesiapan, inspector kanan collapsible, section navigator, inline validation, mode fokus editor, autosave timestamp/footer, shortcut `Ctrl+S`, dan guard tutup modal saat draft lokal aktif.
- [x] Metadata komposer dipadatkan menjadi dense bar: judul kecil, field tinggi 32px, error mapel inline, dan helper bobot tidak lagi mengambil baris besar.
- [x] Opsi jawaban komposer dibuat fleksibel: default 4 opsi, dapat ditambah sampai 6 dan dikurangi kembali tanpa menggeser label yang sudah ada; backend ikut dilonggarkan dari constraint A-E lama agar kunci F dan format jawaban dinamis tetap bisa disimpan.
- [x] Label kecil "Komposer Soal Legacy" pada header dialog komposer dihapus agar modal terasa lebih bersih dan tidak menonjolkan istilah legacy ke pengguna.
- [x] Backend menambahkan metadata pemakaian soal (`package_count`, `answer_count`, `is_locked`, `usage`) pada list/detail agar UI bisa membedakan soal draft yang masih bisa diedit cepat dan soal yang harus direvisi lewat duplikasi.
- [x] Edit/hapus langsung diblokir untuk soal yang sudah masuk paket ujian atau memiliki jawaban siswa; approve/publish langsung via edit juga diblokir agar workflow tetap melalui aksi resmi.
- [x] Import CSV legacy sekarang mengecek duplikat terhadap isi bank soal yang sudah ada pada mapel yang sama, bukan hanya duplikat dalam file import.
- [x] Test backend ditambahkan untuk deduplikasi import, lock edit/hapus soal terpakai, dan guardrail workflow update langsung.
- [x] Verification: `make db-sqlc`, `go test ./...`, `go vet ./...`, Svelte autofixer untuk file tersentuh, `npm run check:web`, dan `git diff --check`.

### ✅ Sprint 61 — CBT Essay Authoring & Manual Scoring (COMPLETE)
- [x] Komposer `/cbt/soal` sekarang mendukung bentuk soal Pilihan Ganda dan Essay dalam mode Pemula/Advance pada satu modal utama.
- [x] Mode Pemula menjaga input ringkas untuk PG dan Essay; mode Advance membuka stimulus, blueprint kurikulum, level kognitif, HOTS, pembahasan, rubrik, dan ajukan review.
- [x] Essay disimpan lewat kontrak CBT yang sama (`question_type=essay`, `options=[]`, `answer_key=''`, `rubric_html`) tanpa route backend baru.
- [x] Scoring backend untuk essay dibenahi agar `manual_score / 100 * points`; nilai `0` tetap dianggap sudah dikoreksi, sedangkan `NULL` berarti belum dikoreksi.
- [x] Workbench Koreksi Uraian pada detail sesi menampilkan stimulus, soal, jawaban siswa, rubrik, bobot, input nilai 0-100, dan refresh skor setelah simpan.
- [x] Query `ListUngradedEssays` diperkaya dengan stem/stimulus/rubrik/bobot dan `sqlc` diregenerasi.
- [x] Verification: `make db-sqlc`, Svelte autofixer untuk komposer dan detail sesi, `npm run check:web`, `go test ./internal/service ./internal/handler`, `go test ./...`, dan `git diff --check`.

### ✅ Sprint 62A — CBT Questions UI Retirement (COMPLETE)
- [x] UI lama `/cbt/questions` dipensiunkan sebagai route authoring langsung dan diganti redirect GET ke `/cbt/soal` agar tidak ada dua pengalaman bank soal yang saling bersaing.
- [x] `/cbt/soal` tidak lagi menampilkan mode "Editor Lanjutan" yang membuka route lama; mode modul disederhanakan menjadi Katalog, Komposer Soal, Review, dan Import Legacy.
- [x] Fallback klik soal di `/cbt/soal` tidak lagi memindahkan guru ke `/cbt/questions`; soal yang aman tetap dibuka di komposer utama, sedangkan soal terkunci/review/tipe belum termigrasi diberi arahan untuk duplikasi atau menunggu migrasi Sprint 62.
- [x] Tautan lama `/cbt/questions?question_id=...` tetap diterima dan diarahkan ke `/cbt/soal?question_id=...` supaya bookmark lama tidak menjadi 404.
- [x] Policy `AGENTS.md` diperbarui: `/cbt/soal` adalah satu-satunya UI Bank Soal aktif; `/api/cbt/questions/*` tetap kontrak data canonical dan tidak dihapus.

### ✅ Sprint 62B — CBT Draft Review Flow (COMPLETE)
- [x] Komposer `/cbt/soal` memisahkan aksi `Simpan Draft` dan `Ajukan Review`; `Ctrl+S` sekarang menyimpan draft, bukan menuntut kesiapan 100%.
- [x] Draft server boleh disimpan saat mapel, stem/pertanyaan, dan bobot minimal sudah ada; opsi/kunci/rubrik lengkap baru wajib ketika soal diajukan review atau dipublish.
- [x] Backend `CbtQuestion` melonggarkan validasi hanya untuk `workflow_status=draft` dan `status=draft`, sementara review/approved/published tetap memakai validasi lengkap.
- [x] Field advance tersembunyi seperti stimulus, pembahasan, CP/TP/KD, indikator, topik, kognitif, dan HOTS tetap dikirim saat menyimpan agar tidak hilang ketika guru berpindah ke mode Pemula.
- [x] UI advance tidak lagi punya dropdown alur review ganda; alur draft/review dikendalikan dari footer komposer.

### ✅ Sprint 62C — CBT Type-Aware Composer Foundation (COMPLETE)
- [x] Komposer `/cbt/soal` sekarang memakai konfigurasi tipe soal untuk Pilihan Ganda, Pilihan Ganda Kompleks/Jawaban Ganda, Benar/Salah, Isian Singkat, dan Essay dalam satu modal.
- [x] Pemilih bentuk soal dibuat ringkas dalam metadata komposer; mode Pemula/Advance tetap menulis kontrak backend yang sama.
- [x] Readiness score, quality signals, preview siswa, validasi kunci, opsi, rubrik, dan payload save sekarang berubah sesuai tipe soal.
- [x] Jawaban Ganda mendukung opsi A-F fleksibel, checkbox multi-kunci, validasi minimal dua kunci, dan payload `answer_key` comma-separated.
- [x] Benar/Salah memakai dua opsi tetap `Benar/Salah` dengan kunci tunggal tanpa editor opsi yang tidak perlu.
- [x] Isian Singkat memakai field kunci teks dan preview satu kolom jawaban sebagai fondasi awal.
- [x] Backend `CbtQuestion` mengizinkan mode Pemula untuk tipe juknis yang sudah dikonfigurasi, menjaga draft tetap longgar, dan tetap memperketat review/publish.

### ✅ Sprint 62D — CBT Short Answer Scoring & Mobile Renderer (COMPLETE)
- [x] Isian Singkat sekarang mendukung beberapa jawaban diterima dalam satu `answer_key` dengan separator `|`, tanpa migration baru.
- [x] Backend menormalisasi alias jawaban singkat: trim, deduplikasi case-insensitive, spasi ganda/NBSP dirapikan, lalu menyimpan format deterministik.
- [x] SQL scoring `UpdateAnswerCorrectness` dan `UpdateParticipantAnswerCorrectness` sekarang menilai `short_answer` dengan alias dan normalisasi huruf besar/kecil serta spasi; toleransi typo tetap tidak aktif.
- [x] Scoring `multiple_answer` ikut dirapikan agar label yang tersimpan dengan spasi tetap dibandingkan setelah trim dan sort.
- [x] Komposer `/cbt/soal` menampilkan microcopy alias jawaban singkat, jumlah jawaban diterima, preview alias, dan mengirim payload alias yang sudah dinormalisasi.
- [x] Flutter exam model membaca `question_type`; renderer mobile membedakan essay, isian singkat, pilihan tunggal, dan jawaban ganda.
- [x] Flutter PG Kompleks/Jawaban Ganda memakai checkbox dan menyimpan jawaban sebagai label comma-separated yang cocok dengan kontrak backend all-or-nothing.
- [x] Flutter Isian Singkat memakai input pendek dengan pesan validasi khusus, bukan field essay besar.

### ✅ Sprint 62E — CBT Agree/Disagree & Template Retirement (COMPLETE)
- [x] Fitur `Template Cepat` dihapus dari modal komposer `/cbt/soal` agar pembuatan soal dimulai langsung dari metadata, bentuk soal, stem, opsi/kunci, dan readiness tanpa blok tambahan yang mengambil ruang.
- [x] Komposer `/cbt/soal` menambahkan bentuk `Setuju/Tidak Setuju` sebagai tipe semantik terpisah dari `Benar/Salah`, memakai dua opsi tetap `Setuju/Tidak Setuju`, kunci tunggal, readiness khusus, dan preview siswa yang tetap satu kontrak.
- [x] Backend `CbtQuestion` mendukung `question_type=agree_disagree` untuk mode Pemula/Advance, opsi otomatis A/B, validasi answer key objektif, dan draft/review flow yang sama dengan tipe soal lain.
- [x] Test backend ditambahkan untuk normalisasi `agree_disagree` dan pengecualian fixed-pair dari aturan minimal 4 opsi di mode Pemula.

### ✅ Sprint 62F — CBT Matching Authoring & Scoring (COMPLETE)
- [x] Komposer `/cbt/soal` menambahkan bentuk `Menjodohkan` dengan editor pasangan kiri-kanan, default 4 pasangan, fleksibel 2-6 pasangan, draft autosave, readiness, quality signals, dan preview siswa.
- [x] Payload `Menjodohkan` memakai `question_type=matching`, `options[]` berisi kolom kiri plus `match_label/match_text/match_html`, dan `answer_key` deterministik `A=1;B=2;...` tanpa migration baru.
- [x] Backend `CbtQuestion` mendukung validasi domain `matching`: minimal 2 pasangan, kolom kiri/kanan wajib lengkap, label tidak duplikat, dan answer key harus memetakan semua pasangan.
- [x] SQL scoring `UpdateAnswerCorrectness` dan `UpdateParticipantAnswerCorrectness` menilai `matching` secara all-or-nothing dengan pasangan yang disortir sebelum dibandingkan.
- [x] Flutter exam renderer menampilkan bank pasangan kanan dan dropdown per item kiri; jawaban disimpan sebagai `A=1;B=2` agar cocok dengan scoring backend.
- [x] Test backend dan Flutter ditambahkan untuk parsing, normalisasi, validasi, dan render tipe `Menjodohkan`.

### ✅ Sprint 62G — CBT Matching Distractors (COMPLETE)
- [x] Komposer `/cbt/soal` menambahkan distraktor kanan opsional untuk `Menjodohkan`, maksimal 4 pilihan ekstra, dengan editor rich text compact dan preview siswa.
- [x] Payload distraktor memakai `options[].is_distractor=true` plus `match_label/match_text/match_html`, sehingga tetap satu kontrak `options[]` tanpa migration baru.
- [x] Backend menerima dan memvalidasi distraktor: label kanan wajib unik, teks kanan wajib ada, dan distraktor tidak boleh menjadi kunci benar.
- [x] Flutter exam renderer menampilkan distraktor dalam bank pilihan kanan, tetapi tidak membuat dropdown/baris kiri tambahan; answered count baru naik saat semua item kiri terjawab.
- [x] Test backend dan Flutter diperbarui untuk parsing, validasi, dan rendering distraktor kanan.

### ✅ Sprint 62H — CBT Multi-Type CSV Import (COMPLETE)
- [x] Backend import CSV tetap kompatibel dengan format PG legacy tanpa kolom tipe, tetapi sekarang menerima kolom `tipe`/`question_type` untuk `pg_kompleks`, `benar_salah`, `setuju_tidak_setuju`, `isian`, `essay`, dan `menjodohkan`.
- [x] Parser import menormalisasi kunci jawaban per tipe: multi-kunci PG Kompleks, label tetap Benar/Salah dan Setuju/Tidak Setuju, alias Isian, default pasangan Menjodohkan, dan rubrik Essay sebagai draft.
- [x] Menjodohkan dari CSV mendukung pasangan kiri-kanan serta distraktor kanan sederhana tanpa migration baru, memakai kontrak `options[]` dan `answer_key` yang sama dengan komposer.
- [x] Layar `/cbt/soal` mengganti copy import dari legacy-only menjadi import CSV multi-tipe agar guru paham satu pintu migrasi bank soal.
- [x] Test backend ditambahkan untuk memastikan import PG legacy lama tetap berjalan dan import tipe baru memetakan `question_type`, `answer_key`, `options`, distraktor, dan rubrik dengan benar.

### ✅ Sprint 62I — CBT Confirm Dialog Layer Hotfix (COMPLETE)
- [x] Dialog konfirmasi global dinaikkan ke layer di atas dialog halaman, sehingga tombol `Tutup` pada modal penyusunan soal menampilkan konfirmasi di depan komposer, bukan di belakangnya.
- [x] Patch dibuat di komponen global confirm agar perbaikan berlaku untuk konfirmasi lain yang dipanggil dari dalam modal besar tanpa mengubah kontrak `Dialog.Root`.

### ✅ Sprint 62J — CBT Multi-Type CSV Export (COMPLETE)
- [x] Backend `CbtQuestion` menambahkan export CSV bank soal memakai filter katalog aktif (`subject_id`, workflow, tipe soal, HOTS, dan pencarian) dengan batas aman 2.000 baris per file.
- [x] Format export memakai header yang kompatibel dengan import multi-tipe: PG, PG Kompleks, Benar/Salah, Setuju/Tidak Setuju, Isian, Essay, Menjodohkan, pasangan kiri-kanan, distraktor, rubrik, pembahasan, dan metadata blueprint.
- [x] Web Admin menambahkan route BFF stream `/api/cbt/questions/export` dan tombol `Export CSV` di `/cbt/soal`, sehingga export tetap melalui Go API dan tidak membaca database langsung.
- [x] Test backend ditambahkan untuk mapping CSV multi-tipe dan handler export stream, termasuk filename, content-type, dan forwarding filter.

### ✅ Sprint 62K — CBT CSV Template & Roundtrip Hardening (COMPLETE)
- [x] Backend menambahkan template CSV resmi `template-bank-soal-cbt.csv` berisi contoh PG, PG Kompleks, Benar/Salah, Isian, Menjodohkan, dan Essay dengan header yang sama seperti export/import.
- [x] Import CSV diperluas agar membaca metadata hasil export: stimulus, pembahasan, kesulitan, kelas, CP/TP/KD, indikator, topik materi, level kognitif, dan HOTS; status/workflow tetap dipaksa draft untuk keamanan.
- [x] Web Admin menambahkan tombol `Template CSV` pada panel import `/cbt/soal` melalui BFF stream `/api/cbt/questions/template`.
- [x] Test service mengunci roundtrip export -> import dan template -> import agar tipe soal, kunci, pasangan menjodohkan, distraktor, dan metadata penting tidak drift.

### ✅ Sprint 62L — CBT Non-Test Assessment Foundation (COMPLETE)
- [x] Backend menambahkan modul `non_test_assessments` dan `non_test_assessment_submissions` lewat migration 056 untuk Praktik, Portofolio, Proyek, Penugasan, Observasi, dan Lainnya tanpa mencampur data ke `cbt_questions`.
- [x] Go API menambahkan kontrak `/api/cbt/non-test-assessments` untuk list/detail/create/update/delete serta fondasi submission/manual scoring; semua query lewat `sqlc` dan service domain tetap di `services/core-api`.
- [x] Web Admin menambahkan BFF proxy dan halaman `/cbt/non-test` dengan mode Pemula/Advance, filter, ringkasan status, form instruksi/rubrik/bukti, checklist observasi, dan CRUD awal.
- [x] Sidebar CBT menampilkan menu `Asesmen Non-Tes` agar modul mudah ditemukan terpisah dari Bank Soal, Paket Ujian, dan Sesi Ujian.
- [x] Test service mengunci normalisasi default, validasi checklist, dan filter list; verification: `make db-sqlc`, Svelte autofixer, `go test ./...`, dan `npm run check:web`.

### ✅ Sprint 62M — CBT Non-Test Roster & Scoring (COMPLETE)
- [x] Backend menambahkan query `GenerateNonTestSubmissionsForClass` untuk menyiapkan daftar siswa aktif dari kelas asesmen secara idempotent, tanpa menduplikasi submission yang sudah ada.
- [x] Go API dan BFF menambahkan endpoint `/api/cbt/non-test-assessments/{id}/submissions/generate` serta tetap memakai endpoint submission yang sama untuk koreksi manual.
- [x] Service non-tes memvalidasi nilai manual agar tidak negatif, wajib ada ketika status `reviewed`, dan tidak melebihi `max_score` asesmen.
- [x] UI `/cbt/non-test` menambahkan panel koreksi: buka dari baris asesmen, siapkan siswa, lihat status siswa, isi nilai, catatan bukti, dan feedback per siswa.
- [x] Test service diperluas untuk generate roster, validasi nilai maksimum, dan kewajiban nilai saat status `reviewed`.

### ✅ Sprint 62N — CBT Non-Test Grade/Rapor Sync (COMPLETE)
- [x] Migration 057 menambahkan link resmi `grade_component_id`, `grade_synced_at`, dan `grade_synced_by` pada `non_test_assessments` agar sumber nilai non-tes bisa ditelusuri sampai komponen nilai.
- [x] Backend menambahkan sinkronisasi transaksional dari asesmen non-tes ke Grade/Rapor: cari assignment kelas-mapel, blokir assignment final, jaga scope guru, buat/reuse komponen nilai, lalu upsert nilai siswa yang sudah `reviewed`.
- [x] Go API dan BFF menambahkan endpoint `/api/cbt/non-test-assessments/{id}/sync-grade` tanpa akses database langsung dari SvelteKit.
- [x] UI `/cbt/non-test` menampilkan status sinkron nilai dan tombol `Kirim Nilai` dari daftar maupun panel koreksi, dengan guard kelas wajib dan minimal satu submission reviewed.
- [x] Test service mengunci pembuatan komponen nilai, mapping kategori non-tes ke kategori grade, pengiriman entry nilai, marker sinkron, dan penolakan akses guru lintas assignment.

### ✅ Sprint 62O — CBT Non-Test Grade Traceability (COMPLETE)
- [x] Query `ListGradeComponents` menampilkan metadata sumber non-tes (`assessment_id`, judul, bentuk, waktu sinkron, dan aktor sinkron) melalui join read-only ke `non_test_assessments`.
- [x] UI `/grades` menandai komponen nilai yang berasal dari non-tes dengan badge `Non-Tes`, bentuk asesmen, judul asal, dan shortcut `Buka Asesmen Asal`.
- [x] Shortcut dari `/grades` menuju `/cbt/non-test?assessment_id=...` membuka panel koreksi asesmen asal bila item masih ada dalam daftar aktif.
- [x] Tidak ada migration baru; traceability memakai link `grade_component_id` dari Sprint 62N dan tetap menjaga SvelteKit sebagai UI/BFF.

### ✅ Sprint 62P — CBT Non-Test Grade Source Lock (COMPLETE)
- [x] Backend menolak edit komponen, hapus komponen, dan input nilai langsung untuk komponen grade yang berasal dari asesmen non-tes.
- [x] Publish/unpublish tetap menjadi workflow grade, tetapi isi nilai dan struktur sumber harus dikoreksi dari `/cbt/non-test` lalu disinkron ulang.
- [x] UI `/grades` membuat komponen non-tes read-only untuk edit/hapus/input nilai, serta mempertahankan shortcut `Buka Asesmen Asal`.
- [x] Test service mengunci guard non-tes agar mutasi grade tidak diteruskan ke repository ketika komponen punya sumber non-tes.

### ✅ Sprint 62Q — CBT Non-Test Sync Freshness (COMPLETE)
- [x] Query non-tes menampilkan `last_reviewed_at` dan jumlah `unsynced_reviewed_submissions` tanpa migration baru, dihitung dari submission reviewed dan waktu sinkron terakhir.
- [x] Handler list/detail meneruskan metadata freshness agar UI dapat membedakan `Siap kirim nilai`, `Tersinkron nilai`, dan `Perlu sinkron ulang`.
- [x] UI `/cbt/non-test` menambahkan ringkasan `Perlu Sinkron`, deskripsi sinkron per asesmen, dan panel koreksi berisi sinkron terakhir, review terakhir, serta jumlah nilai reviewed yang belum terkirim.
- [x] Setelah koreksi nilai pada asesmen yang sudah pernah disinkron, UI memberi peringatan agar operator mengirim ulang nilai ke gradebook.

### ✅ Sprint 62R — CBT Non-Test Sync Filter (COMPLETE)
- [x] Backend list/count asesmen non-tes menerima filter `sync_filter=needs_sync` untuk menampilkan asesmen yang punya nilai reviewed dan belum/perlu dikirim ulang ke gradebook.
- [x] Service menormalisasi filter sinkron agar nilai selain `needs_sync` diabaikan, dan test memastikan filter valid diteruskan ke query list/count.
- [x] BFF tetap hanya meneruskan query ke Go API; tidak ada akses database langsung dari SvelteKit.
- [x] UI `/cbt/non-test` menambahkan tombol cepat dan kartu ringkasan `Perlu Sinkron` yang dapat mengaktifkan filter server-side serta badge filter aktif pada daftar.

### ✅ Sprint 62 — CBT Bentuk Soal Juknis/Asesmen Madrasah (COMPLETE)
- [x] Tujuan utama: membuat bank soal CBT lengkap untuk bentuk soal tertulis yang lazim dipakai pada Asesmen Madrasah/Juknis madrasah, tanpa membuat guru kewalahan saat membuat soal harian.
- [x] Basis riset produk: bentuk tertulis yang perlu dipetakan adalah Pilihan Ganda, Pilihan Ganda Kompleks, Benar/Salah, Setuju/Tidak Setuju, Menjodohkan, Isian/Jawaban Singkat, dan Uraian/Essay; bentuk non-tes seperti Praktik, Portofolio, Penugasan, dan bentuk lain madrasah diperlakukan sebagai asesmen non-CBT dengan rubrik dan bukti.
- [x] Rujukan implementasi awal: artikel resmi Kemenag daerah tentang Asesmen Madrasah dan SOP publik Asesmen Madrasah 2025 dipakai sebagai baseline produk, tetapi implementasi harus tetap mudah disesuaikan jika juknis pusat terbaru berubah.

#### UX Mode Pemula vs Advance
- [x] Semua bentuk soal di `/cbt/soal` wajib punya mode `Pemula` dan `Advance` dengan kontrak backend yang sama agar guru bisa mulai sederhana lalu memperkaya soal tanpa migrasi data.
- [x] Mode `Pemula`: hanya tampilkan field wajib per bentuk soal, seperti mapel, tingkat kesulitan, bobot, stem, opsi/jawaban, dan kunci; hide stimulus, blueprint, pembahasan, rubrik detail, dan workflow review sampai dibutuhkan.
- [x] Mode `Advance`: buka stimulus, CP/TP/KD, level kognitif, HOTS, media, RTL Arab, pembahasan, rubrik, catatan reviewer, workflow ajukan review, dan metadata paket.
- [x] Pemilih bentuk soal memakai type picker ringkas berbasis konfigurasi tipe sehingga guru memilih bentuk soal langsung tanpa route kedua atau daftar authoring panjang.
- [x] Readiness score dan quality signals berubah sesuai tipe soal: PG fokus opsi/kunci, essay fokus rubrik, isian fokus jawaban diterima, matching fokus pasangan lengkap.

#### Phase 1 — Ekstensi Aman Di Atas Kontrak Yang Sudah Ada
- [x] Tambahkan `Pilihan Ganda Kompleks` ke komposer utama dengan opsi A-F fleksibel, multi-kunci, preview siswa, validasi minimal 2 jawaban benar, dan penilaian all-or-nothing sebagai default awal.
- [x] Tambahkan `Benar/Salah` sebagai bentuk cepat dengan dua opsi otomatis, kunci tunggal, copy khusus madrasah, dan tetap memakai rich text legacy untuk stem/stimulus.
- [x] Tambahkan `Isian/Jawaban Singkat` dengan input jawaban kunci dan preview siswa satu field.
- [x] Tambahkan alias jawaban diterima, normalisasi huruf besar/kecil/spasi, dan kebijakan exact-match yang jelas untuk Isian/Jawaban Singkat.
- [x] Pastikan create/edit/duplicate/import CSV legacy tetap tidak merusak soal PG dan Essay yang sudah berjalan.
- [x] Perluas test backend dan frontend untuk validasi payload, answer key, readiness, dan edit cepat pada bentuk soal baru.

#### Phase 2 — Bentuk Juknis Yang Butuh Model Data Lebih Kaya
- [x] Tambahkan `Setuju/Tidak Setuju` sebagai varian semantik terpisah dari Benar/Salah agar laporan dan bank soal bisa membedakan sikap/pendapat dari fakta benar-salah.
- [x] Tambahkan `Menjodohkan` dengan editor pasangan kiri-kanan, preview siswa, dan format jawaban deterministik agar scoring bisa diuji.
- [x] Tambahkan distraktor kanan opsional untuk `Menjodohkan` setelah format dasar pasangan, scoring, dan renderer stabil.
- [x] Definisikan scoring `Menjodohkan`: all-or-nothing sebagai default Pemula; partial credit dapat dibuka di Advance setelah format jawaban dan laporan stabil.
- [x] Perluas Flutter exam renderer untuk Menjodohkan sebelum tipe tersebut dipakai pada sesi ujian produksi; PG Kompleks, Benar/Salah, Setuju/Tidak Setuju, dan Isian sudah punya renderer awal.
- [x] Tambahkan import CSV bertahap untuk tipe baru: PG Kompleks, Benar/Salah, Setuju/Tidak Setuju, Isian, Essay, dan Menjodohkan.
- [x] Tambahkan export CSV bertahap untuk tipe baru dari katalog aktif, dimulai dari PG Kompleks dan Benar/Salah, lalu Isian dan Menjodohkan setelah format final.

#### Phase 3 — Asesmen Non-Tes Di Luar CBT Murni
- [x] Buat konsep `Asesmen Non-Tes` terpisah dari bank soal CBT live untuk Praktik, Portofolio, Penugasan, Proyek, atau bentuk lain yang ditetapkan madrasah.
- [x] Non-tes harus memakai instruksi tugas, rubrik, bukti/link/lampiran, status pengumpulan, dan penilaian manual; jangan dipaksa menjadi soal CBT satu-jawaban.
- [x] Integrasikan nilai non-tes ke Grade/Rapor melalui komponen nilai yang jelas, bukan ke runtime ujian token.
- [x] Mode Pemula untuk non-tes cukup berisi instruksi, bobot, tenggat, dan rubrik sederhana; mode Advance membuka kriteria rubrik, bukti wajib, penilai, dan checklist observasi.

#### Aturan Penilaian Awal
- [x] Pilihan Ganda: jawaban benar tunggal bernilai penuh, salah nol.
- [x] Pilihan Ganda Kompleks: semua kunci harus tepat untuk nilai penuh pada fase awal; partial credit disiapkan sebagai opsi Advance setelah laporan siap.
- [x] Benar/Salah dan Setuju/Tidak Setuju: exact match terhadap kunci.
- [x] Isian/Jawaban Singkat: exact match setelah normalisasi dan alias jawaban diterima; toleransi typo tidak aktif secara default.
- [x] Menjodohkan: exact match semua pasangan pada fase awal; partial credit per pasangan hanya di Advance.
- [x] Essay/Uraian: manual scoring 0-100 berbasis rubrik, dikonversi proporsional ke bobot soal seperti Sprint 61.

#### Guardrail Implementasi
- [x] Tetap patuhi boundary: SvelteKit hanya UI/BFF, semua validasi domain dan scoring di Go API, SQL eksplisit lewat `sqlc`, migration hanya di `services/core-api/db/migrations`.
- [x] Tidak menghidupkan kembali UI `/cbt/questions`; tambahkan tipe secara bertahap langsung di `/cbt/soal` sambil menjaga kontrak backend `/api/cbt/questions/*`.
- [x] Tipe baru dipakai setelah Flutter renderer, backend scoring, preview siswa, dan laporan hasil siap untuk bentuk tertulis Sprint 62.
- [x] Setiap fase harus update `PLAN.md`, menjalankan Svelte autofixer untuk file `.svelte` tersentuh, `npm run check:web`, test Go terkait, dan `git diff --check`.

### ✅ Sprint 63 — CBT Ruangan, Pengawas, dan Readiness Operasional (COMPLETE)
- [x] Tujuan utama: memisahkan jelas `ruangan fisik/aset` dari `ruang/kelompok ujian CBT`, lalu menghubungkannya tanpa menghapus data sesi lama.
- [x] Current phase: tetap di `services/core-api` sebagai backend utama; belum perlu memisahkan service baru karena pain saat ini adalah boundary data dan workflow operasional, bukan kebutuhan deploy/runtime terpisah.
- [x] Boundary keputusan: master ruangan fisik berada di domain inventory/sarpras, sedangkan ruang ujian, token ruang, peserta, meja, pengawas, dan proctoring berada di domain CBT.
- [x] Compatibility bridge: `cbt_exam_rooms.room_name` dan `capacity` tetap dipertahankan untuk sesi lama; kolom baru bersifat nullable/snapshot agar migrasi aman dan rollback tetap jelas.

#### Phase 1 — Master Ruangan Fisik dan Link Ruang CBT
- [x] Tambahkan master ruangan fisik `school_rooms` dengan kode, nama, gedung/lantai/lokasi, tipe ruang, kapasitas normal, kapasitas ujian, kondisi, status layak ujian, kesiapan jaringan/listrik, dan catatan sarpras.
- [x] Tambahkan link nullable `school_room_id` pada `cbt_exam_rooms` plus snapshot nama/kapasitas/status agar perubahan nama aset tidak merusak arsip sesi lama.
- [x] UI tab `Ruangan` pada detail sesi CBT dapat memilih ruangan dari master aset, tetapi tetap menyediakan input manual sebagai fallback transisi.
- [x] SvelteKit hanya menjadi BFF/proxy; semua CRUD ruangan fisik, validasi kapasitas, dan link ruang sesi tetap lewat Go API.

#### Phase 2 — Penugasan Pengawas Per Ruang
- [x] Tambahkan tabel `cbt_room_proctors` untuk pengawas utama, pendamping, dan cadangan per ruang ujian.
- [x] Detail sesi CBT menampilkan pengawas utama per ruang dan menyediakan aksi penugasan awal dari daftar pegawai aktif.
- [x] Otorisasi proctoring room-scoped mengikuti model legacy: dashboard ruang hanya bisa dibuka admin atau pegawai yang ditugaskan sebagai pengawas ruang tersebut.
- [x] Audit event mencatat perubahan pengawas agar penugasan ruang ujian dapat ditelusuri.

#### Phase 3 — Readiness Check dan Operasional Pengawas
- [x] Tambahkan readiness ringkas: jumlah peserta, peserta belum punya ruang, kapasitas total, ruang tanpa pengawas, dan nomor meja belum lengkap.
- [x] Blokir aktivasi sesi CBT jika peserta, ruangan, kapasitas, nomor meja, atau pengawas ruang belum siap; UI daftar sesi menampilkan alasan backend agar operator tahu tindakan koreksi berikutnya.
- [x] Buat dashboard pengawas per ruang mirip legacy proctoring: ringkasan ruang, token ruang, daftar pengawas, peserta, nomor meja, heartbeat, pelanggaran, reset akses, flag atensi, force submit, dan log ruang dibatasi ke ruang tersebut.
- [x] Tambahkan cetak paket pengawas per ruang: daftar hadir, token ruang, denah/meja, kontak operator, dan checklist kesiapan.

#### Guardrail Sprint 63
- [x] Jangan memasukkan ruangan fisik sebagai `inventory_items` biasa karena semantik stok (`jumlah_total`, `jumlah_baik`, `min_stock`) tidak cocok untuk ruang.
- [x] Jangan menghapus `cbt_exam_rooms` lama; evolusikan dengan kolom tambahan dan query baru.
- [x] Jangan memakai SQLite/Drizzle runtime dari legacy; legacy hanya referensi perilaku produk.
- [x] Verification wajib: `make db-sqlc`, Go tests terkait, Svelte autofixer untuk file `.svelte` tersentuh, `npm run check:web`, dan `git diff --check`.

### ✅ Sprint 64 — CBT Ruang Saya Pengawas (COMPLETE)
- [x] Tujuan utama: menyediakan workspace ringkas untuk pengawas agar ruang ujian yang ditugaskan mudah ditemukan tanpa menelusuri detail sesi satu per satu.
- [x] Backend menambahkan query dan endpoint `GET /api/cbt/proctoring/my-rooms`; guru/staf melihat ruang berdasarkan `employee_id` pada JWT, admin melihat ringkasan semua ruang untuk kebutuhan operator.
- [x] SvelteKit tetap BFF/proxy melalui `/api/cbt/proctoring/my-rooms`; tidak ada akses database langsung dari frontend.
- [x] Halaman `/cbt/proctoring/rooms` menampilkan kartu ruang, statistik aktif/terjadwal/peserta/atensi, filter status, pencarian, tabel ringkas, dan pintasan ke dashboard ruang, paket cetak, serta detail sesi.
- [x] Sidebar CBT menambahkan menu `Ruang Saya` untuk `admin`, `guru`, dan `staf`; quick pin guru diarahkan ke ruang pengawas agar alur monitoring lebih cepat.
- [x] Guardrail: tidak membuat service baru, tidak mengubah skema, tidak memakai SQLite/Drizzle runtime, dan tetap memakai kontrak Go API + sqlc.

### ✅ Sprint 65 — CBT Room Handover & Deep Review Follow-up (COMPLETE)
- [x] Tujuan utama: menutup gap operasional setelah ujian ruang berjalan, yaitu bukti digital serah-terima akhir ruang oleh pengawas sebelum arsip panitia.
- [x] Review mendalam modul CBT: authoring soal multi-tipe, non-tes, sesi, ruangan, pengawas, proctoring live, print pack, dan workspace pengawas sudah kuat; gap utama ada pada penutupan ruang, register insiden, rekap event-level, dan arsip final.
- [x] Tambahkan tabel `cbt_room_handovers` untuk checklist daftar hadir, submit akhir, gangguan perangkat/jaringan, kerapian ruang, pengembalian token/berkas, pengembalian aset cadangan, catatan kejadian, catatan operator, catatan serah-terima, serta lock audit.
- [x] Backend menambahkan query sqlc dan endpoint `GET/PUT /api/cbt/sessions/{id}/rooms/{rid}/handover` serta `POST /api/cbt/sessions/{id}/rooms/{rid}/handover/lock`.
- [x] Otorisasi mengikuti dashboard pengawas ruang: admin atau pegawai yang ditugaskan sebagai pengawas ruang; SvelteKit tetap BFF/proxy tanpa akses DB langsung.
- [x] Dashboard pengawas ruang sekarang memiliki panel `Serah Terima Akhir Ruang` berisi checklist compact, statistik submit/atensi/cek ulang, catatan kejadian, catatan operator, catatan serah-terima, tombol simpan draft, dan tombol kunci handover.
- [x] Penguncian handover memblokir edit lanjutan lewat backend dan frontend; konflik dikembalikan sebagai respons terkontrol, bukan silent failure.
- [x] Audit event mencatat `CBT_SESSION_ROOM_HANDOVER_SAVE` dan `CBT_SESSION_ROOM_HANDOVER_LOCK` untuk jejak operator/pengawas.
- [x] Guardrail: tidak mengubah kontrak live exam Flutter, tidak memindahkan domain CBT ke service baru, tidak memakai storage frontend, dan tetap evolutif di Go API + PostgreSQL + sqlc.

#### Rekomendasi Gap CBT Berikutnya
- [x] Buat rekap handover tingkat sesi: daftar ruang sudah/belum dikunci, ruang bermasalah, insiden, dan submit akhir untuk kepala madrasah/panitia.
- [x] Bentuk register insiden CBT awal yang menggabungkan telemetry proctoring, flag pengawas, force submit, reset akses, dan catatan manual handover.
- [x] Tambahkan export CSV rekap operasional ruang agar bukti digital bisa dibawa ke arsip/berita acara sementara.
- [x] Tambahkan rekonsiliasi pasca-sesi awal: login, submit, force submit, reset akses, no-show, atensi, app switch, screenshot, dan status handover per ruang.
- [x] Perkuat bank soal tahap awal setelah operasional stabil dengan blueprint coverage matrix CP/TP/KD saat menyusun paket ujian.
- [x] Lanjutkan analisis butir pasca-ujian berbasis hasil siswa, tingkat kesukaran, daya pembeda sederhana, kualitas opsi, dan rekomendasi revisi soal.
- [x] Hubungkan analisis butir ke bank soal dengan aksi draft revisi yang aman agar guru dapat memperbaiki soal bermasalah tanpa mengubah riwayat soal ujian.
- [x] Permudah guru menemukan draft revisi dengan antrian revisi di `/cbt/soal`, termasuk shortcut filter dan edit cepat.
- [x] Tampilkan jejak alasan revisi dari catatan reviewer/analisis butir langsung pada antrian revisi dan katalog.
- [x] Tambahkan filter sumber revisi (`Analisis Butir`, `Reviewer`, `Workflow`) agar antrian revisi tetap mudah dipilah saat jumlahnya banyak.
- [x] Tutup loop revisi dengan aksi `Ajukan Review Ulang` langsung dari antrian revisi dan katalog.
- [x] Bentuk meja kerja reviewer untuk soal `Ditinjau`, lengkap dengan aksi setuju dan minta revisi agar soal yang sudah diajukan tidak tenggelam di katalog besar.
- [x] Tutup tahap `Disetujui -> Terbit` dengan antrean soal siap terbit dan aksi publish langsung dari Bank Soal.
- [x] Perkuat pembuat paket agar hanya memakai soal `Terbit`, menampilkan sinyal mutu pool soal, dan backend menolak question ID yang belum published.
- [x] Tambahkan ringkasan mutu per paket di daftar paket: distribusi bentuk soal, HOTS, metadata gap, dan warning soal legacy yang belum terbit.
- [x] Bawa quality gate paket ke pembuatan sesi CBT agar operator tidak membuat sesi dari paket kosong, nonaktif, atau legacy yang masih berisi soal belum terbit.

### ✅ Sprint 66 — CBT Session Operational Recap (COMPLETE)
- [x] Tujuan utama: patch cepat dari review CBT setelah handover ruang, yaitu menyediakan satu layar operator untuk melihat semua ruang yang belum menutup handover, punya insiden, atau belum seimbang submit akhir.
- [x] Backend menambahkan query sqlc `GetCbtSessionOperationalRecap` dan `ListCbtSessionRoomOperationalRecap` tanpa skema baru; data diambil dari sesi, ruang, peserta, event proctoring, dan `cbt_room_handovers`.
- [x] Endpoint baru `GET /api/cbt/sessions/{id}/operational-recap` menjaga role gate sesi yang sama dengan detail CBT; SvelteKit tetap BFF/proxy.
- [x] Detail sesi CBT menambahkan tab `Rekap Ops` berisi ringkasan handover terkunci, submit akhir, ruang berinsiden, force submit/reset akses, app switch/screenshot, serta tabel register ruang.
- [x] Tabel register ruang mengurutkan ruang yang belum punya handover atau belum terkunci di atas, lalu ruang dengan event atensi tertinggi.
- [x] Tambahkan export CSV `rekap_operasional_*` dari UI untuk arsip sementara sebelum dibuat print pack/berita acara final.
- [x] Temuan paling cepat dipatch: operator sebelumnya harus membuka dashboard ruang satu per satu untuk tahu ruang mana belum dikunci atau bermasalah; sekarang detail sesi menjadi pusat kontrol pasca-ujian.

### ✅ Sprint 67 — CBT Package Blueprint Coverage (COMPLETE)
- [x] Tujuan utama: menutup patch rekomendasi CBT ke-5 tahap awal, yaitu memberi kontrol mutu paket soal sebelum sesi ujian dibuat.
- [x] Halaman `/cbt/packages` sekarang membaca bank soal paginated melalui BFF sampai beberapa halaman aman, bukan hanya halaman pertama/default list.
- [x] Builder paket menampilkan badge bentuk soal, tingkat kesulitan, level kognitif, dan HOTS pada daftar soal terbit agar operator tidak memilih secara buta.
- [x] Saat soal dipilih, muncul `Blueprint Paket Sementara` berisi jumlah soal, kombinasi CP/TP/KD, jumlah HOTS, distribusi bentuk soal, distribusi level kognitif, dan matriks CP/TP/KD/materi/level.
- [x] Soal yang belum punya CP, TP/KD, atau level kognitif diberi sinyal `perlu metadata`, sehingga bisa diperbaiki di bank soal sebelum dipakai pada paket resmi.
- [x] Pergantian mata pelajaran membersihkan checklist soal dan payload pembuatan paket hanya mengirim soal yang masih valid untuk mapel aktif.
- [x] Guardrail: patch ini UI/BFF saja, tanpa akses database langsung, tanpa migrasi, dan tidak mengubah kontrak Flutter live exam.

### ✅ Sprint 68 — CBT Item Analysis After Exam (COMPLETE)
- [x] Tujuan utama: patch temuan review CBT paling cepat setelah blueprint paket, yaitu mengubah hasil ujian menjadi umpan balik kualitas soal untuk bank soal.
- [x] Backend menambahkan query sqlc `GetSessionItemAnalysis` tanpa skema baru; metrik dihitung dari paket, peserta submit, jawaban siswa, skor, metadata soal, dan distribusi jawaban.
- [x] Endpoint `GET /api/cbt/sessions/{id}/item-analysis` memakai gate sesi yang sama dengan detail CBT; guru hanya melihat sesi mapel yang menjadi scope-nya, admin melihat semua.
- [x] Web-admin menambahkan BFF proxy `/api/cbt/sessions/[id]/item-analysis` agar frontend tetap tidak membaca database langsung.
- [x] Detail sesi CBT menambahkan tab `Analisis Butir` berisi ringkasan total butir, butir perlu review, rata-rata kesukaran, daya pembeda rendah, uraian belum dinilai, dan tabel metrik per butir.
- [x] Rekomendasi per butir mencakup `Baik`, `Belum ada submit`, `Koreksi uraian belum lengkap`, `Belum dijawab`, `Cek kunci/rubrik`, `Terlalu sulit`, `Terlalu mudah`, `Daya pembeda rendah`, dan `Banyak jawaban kosong`.
- [x] Tambahkan export CSV analisis butir untuk arsip MGMP/guru sebelum revisi bank soal.
- [x] Guardrail: tidak mengubah kontrak Flutter live exam, tidak menambah service baru, tidak menambah tabel, dan tetap memakai Go API + sqlc + BFF.

### ✅ Sprint 69 — CBT Item Revision Feedback Loop (COMPLETE)
- [x] Tujuan utama: menutup loop mutu soal pasca-ujian, dari rekomendasi analisis butir ke draft revisi yang bisa diedit guru di bank soal.
- [x] Backend `CbtQuestion` menambahkan `DuplicateForRevision`, membuat salinan aman berstatus `draft` dengan `workflow_status=rejected`, catatan review dari analisis butir, dan kode revisi bertimestamp.
- [x] Endpoint baru `POST /api/cbt/questions/{id}/revision` membuat draft revisi tanpa memutasi soal asli yang sudah masuk paket atau memiliki riwayat jawaban siswa.
- [x] Web-admin menambahkan BFF proxy `/api/cbt/questions/[id]/revision` agar UI tetap melewati Go API dan tidak mengakses database langsung.
- [x] Tab `Analisis Butir` menampilkan tombol `Buat Draft Revisi` hanya untuk rekomendasi warning/danger; aksi memakai konfirmasi, loading state per baris, toast, dan inline operation panel.
- [x] Komposer `/cbt/soal` dapat membuka draft revisi berstatus `rejected/Revisi`; saat disimpan ulang, revisi kembali menjadi draft aman untuk siklus review berikutnya.
- [x] Audit event `CBT_QUESTION_MARK_REVISION` mencatat sumber soal, workflow hasil, author, dan catatan rekomendasi.
- [x] Guardrail: tidak ada migrasi, tidak mengubah kontrak Flutter live exam, tidak mengubah soal asli, dan tetap mengikuti model workflow bank soal yang sudah ada.

### ✅ Sprint 70 — CBT Revision Queue In Question Bank (COMPLETE)
- [x] Tujuan utama: membuat draft revisi tidak tersembunyi di katalog bank soal setelah dibuat dari analisis butir atau workflow reviewer.
- [x] `/cbt/soal` sekarang memuat antrian `workflow_status=rejected` secara paralel dengan daftar utama melalui BFF/API yang sudah ada; tidak ada endpoint atau tabel baru.
- [x] Panel `Antrian Revisi Soal` muncul di atas katalog saat ada revisi, menampilkan total, shortcut `Lihat Semua Revisi`, dan kartu ringkas berisi tipe soal, mapel, preview stem, kode, penulis, dan tingkat kesulitan.
- [x] Klik kartu revisi langsung membuka komposer untuk revisi yang masih aman diedit; filter `Lihat Semua Revisi` mengubah daftar utama ke status `Revisi`.
- [x] Tab review menambahkan metrik `Perlu Revisi` agar panitia/guru dapat melihat beban revisi bersama draft, review, terbit, dan terkunci.
- [x] Guardrail: perubahan hanya UI/BFF client fetch, tetap lewat Go API, tanpa direct DB access, tanpa migrasi, dan tidak mengubah kontrak Flutter live exam.

### ✅ Sprint 71 — CBT Revision Trace Visibility (COMPLETE)
- [x] Tujuan utama: membuat guru tahu alasan revisi tanpa harus membuka detail soal satu per satu.
- [x] Frontend `/cbt/soal` sekarang membaca `review_notes`, `reviewer_username`, dan `reviewed_at` yang sudah tersedia dari list bank soal Go API.
- [x] Kartu `Antrian Revisi Soal` menampilkan sumber revisi (`Analisis Butir`, reviewer, atau workflow review), tanggal review bila ada, dan ringkasan catatan revisi.
- [x] Baris katalog berstatus `Revisi` juga menampilkan alasan revisi ringkas agar konteks tetap terlihat setelah filter `Lihat Semua Revisi`.
- [x] Guardrail: tidak ada query baru, tidak ada migrasi, tidak ada endpoint baru, dan tetap memakai kontrak BFF/Go API yang sudah ada.

### ✅ Sprint 72 — CBT Revision Source Filter (COMPLETE)
- [x] Tujuan utama: membuat antrian revisi tetap mudah dipilah ketika revisi dari analisis butir, reviewer, dan workflow umum mulai menumpuk.
- [x] Backend list bank soal menambahkan filter opsional `revision_source` pada query sqlc existing: `item_analysis`, `reviewer`, dan `workflow`.
- [x] Filter memakai data yang sudah ada (`workflow_status=rejected`, `review_notes`, dan `reviewer_username`), sehingga tidak memerlukan migrasi atau tabel baru.
- [x] Handler meneruskan query param `revision_source`; BFF `/api/cbt/questions` tetap passthrough ke Go API.
- [x] `/cbt/soal` menambahkan tombol cepat sumber revisi di panel `Antrian Revisi Soal`; filter juga mengubah daftar utama ke status `Revisi` agar guru langsung bekerja pada subset yang sama.
- [x] Guardrail: tidak mengubah kontrak Flutter live exam, tidak menambah route baru, dan tetap memakai Go API + sqlc + BFF.

### ✅ Sprint 73 — CBT Revision Resubmit Action (COMPLETE)
- [x] Tujuan utama: menutup alur kerja revisi agar guru dapat mengajukan review ulang setelah memperbaiki soal tanpa mencari aksi lain.
- [x] `/cbt/soal` menambahkan aksi `Ajukan Review Ulang` pada kartu `Antrian Revisi Soal` dan pada baris katalog berstatus `Revisi`.
- [x] Aksi memakai endpoint workflow existing `PATCH /api/cbt/questions/{id}/workflow` dengan `submit_review`; tidak menambah route atau tabel baru.
- [x] Catatan revisi lama tetap dipertahankan karena pengajuan ulang mengirim notes kosong, sehingga alasan dari analisis butir/reviewer tidak hilang.
- [x] UI memakai konfirmasi, loading state per soal, disable guard untuk soal revisi yang tidak aman, toast sukses/gagal, dan refresh daftar setelah berhasil.
- [x] Guardrail: tidak mengubah kontrak Flutter live exam, tidak menambah migrasi, dan tetap memakai BFF + Go API.

### ✅ Sprint 74 — CBT Reviewer Queue & Reject Action (COMPLETE)
- [x] Tujuan utama: membuat pekerjaan reviewer/admin terlihat jelas setelah guru mengajukan soal ke status `Ditinjau`.
- [x] `/cbt/soal` menambahkan `Antrian Review Soal` dengan total pending, kartu ringkas, filter `Lihat Semua Ditinjau`, dan aksi `Setujui` / `Minta Revisi`.
- [x] Modal keputusan review menampilkan stimulus, stem, opsi/pasangan, kunci/rubrik, catatan sebelumnya, serta textarea catatan reviewer.
- [x] Backend workflow existing menambah action `reject` pada `PATCH /api/cbt/questions/{id}/workflow`, role-gated untuk admin, dengan audit event `CBT_QUESTION_REJECT`.
- [x] `reject` mengembalikan soal ke `workflow_status=rejected`, menyimpan reviewer/catatan, dan tetap memakai model/tabel CBT yang sama tanpa migrasi.
- [x] Guardrail: tidak mengubah kontrak Flutter live exam, tidak menambah route baru di luar workflow existing, dan tetap melalui BFF + Go API.

### ✅ Sprint 75 — CBT Approved Publish Queue (COMPLETE)
- [x] Tujuan utama: menutup alur publish agar soal yang sudah `Disetujui` terlihat sebagai pekerjaan admin sebelum dapat dipilih ke paket ujian.
- [x] `/cbt/soal` menambahkan antrean `Siap Terbit ke Paket` untuk soal `workflow_status=approved` dan `status=draft`.
- [x] Aksi `Terbitkan` memakai endpoint workflow existing `PATCH /api/cbt/questions/{id}/workflow` dengan action `publish`, konfirmasi operator, loading per soal, toast hasil, dan refresh daftar.
- [x] Katalog menampilkan aksi `Terbitkan` pada baris soal yang sudah disetujui agar admin tidak harus masuk ke kartu antrean.
- [x] Backend list CBT menambah filter query `status` agar antrean siap terbit tidak tercampur dengan soal yang sudah published; sqlc sudah diregenerasi.
- [x] Guardrail: tidak menambah migrasi, tidak mengubah kontrak Flutter live exam, dan tetap memakai BFF + Go API + sqlc.

### ✅ Sprint 76 — CBT Package Published Guard & Selection Signals (COMPLETE)
- [x] Tujuan utama: memastikan paket ujian hanya tersusun dari soal yang sudah `Terbit`, bukan soal draft/review yang kebetulan terkirim dari client.
- [x] Backend `CbtPackage` menolak pembuatan paket jika ada question ID yang belum `status=published`; guard ini melindungi API meskipun UI/API client lain salah kirim.
- [x] `/cbt/packages` sekarang meminta bank soal dengan query `status=published`, sehingga pool paket lebih ringan dan konsisten dengan aturan paket ujian.
- [x] Builder paket menampilkan ringkasan pool soal terbit: jumlah terbit, metadata kurang, HOTS, dan distribusi bentuk soal sebelum operator memilih.
- [x] Setiap baris soal menampilkan badge `Terbit`, tipe soal, tingkat kesulitan, level kognitif, HOTS, dan peringatan metadata kurang agar operator tidak memilih secara buta.
- [x] Tombol `Buat Paket` baru aktif jika mapel, judul, durasi, dan minimal satu soal terbit sudah dipilih; panel readiness menampilkan alasan jika belum siap.
- [x] Guardrail: tidak menambah migrasi, tetap melalui BFF + Go API, tidak mengubah kontrak Flutter live exam, dan menambah unit test service untuk published guard.

### ✅ Sprint 77 — CBT Package Quality Summary (COMPLETE)
- [x] Tujuan utama: membuat daftar paket ujian menjadi layar review mutu awal, bukan hanya daftar administratif.
- [x] Query `ListCbtPackageQuestions` diperluas dengan metadata soal yang sudah ada: bentuk soal, kesulitan, status, workflow, CP/TP/KD, materi, level kognitif, dan HOTS; tidak ada migrasi baru.
- [x] `/cbt/packages` sekarang menyimpan payload `questions` dari API paket sebagai sumber ringkasan per paket, tetap melalui BFF/Go API tanpa akses database langsung.
- [x] Tabel desktop dan kartu mobile daftar paket menampilkan distribusi bentuk soal, jumlah HOTS, metadata gap, dan warning `belum terbit` untuk paket legacy.
- [x] Ringkasan ini membantu operator mendeteksi paket yang terlalu homogen, kurang blueprint, atau menyimpan soal lama yang belum sesuai siklus publish.
- [x] Guardrail: tidak mengubah kontrak Flutter live exam, tidak menambah endpoint baru, tidak menambah storage frontend, dan tetap memakai query sqlc existing.

### ✅ Sprint 78 — CBT Session Package Quality Gate (COMPLETE)
- [x] Tujuan utama: memindahkan sinyal mutu paket ke titik keputusan pembuatan sesi CBT, sehingga operator tidak baru sadar saat sesi sudah dibuat atau saat Flutter menyajikan soal.
- [x] Query sqlc baru `GetCbtPackageQuestionQuality` menghitung status aktif paket, total soal, soal terbit, soal belum terbit, dan metadata gap tanpa menambah tabel atau migrasi.
- [x] Backend `CbtSession.Create` sekarang menolak paket nonaktif, paket kosong, paket tanpa soal terbit, dan paket legacy yang masih menyimpan soal belum terbit; konflik dikembalikan sebagai respons terkontrol.
- [x] Form `/cbt/sessions` menampilkan panel `Quality Gate Paket` setelah paket dipilih, berisi jumlah soal, distribusi tipe, HOTS, metadata kurang, dan warning belum terbit.
- [x] Tombol `Buat Sesi` mengikuti readiness issues yang sama dengan panel, termasuk jadwal, scope peserta, lintas tingkat, dan kualitas paket.
- [x] Metadata gap tetap warning, bukan blokir, agar asesmen tetap bisa berjalan sambil guru/panitia memperbaiki blueprint bank soal.
- [x] Guardrail: tidak mengubah kontrak Flutter live exam, tidak mengubah model paket, tidak menambah storage frontend, dan tetap memakai Go API + sqlc + BFF.

### ✅ Repository Hygiene Checkpoint — Source Commit Discipline (COMPLETE)
- [x] Pisahkan artefak runtime SQLite sidecar (`*.db-shm`, `*.db-wal`, `*.sqlite-shm`, `*.sqlite-wal`) dari source control sesuai aturan legacy SQLite hanya sebagai artefak impor.
- [x] Exclude lokal dokumen review `Siklus_Dokumen_MTsN_Lengkap*.docx` agar tidak ikut commit fitur.
- [x] Commit dibuat bertahap: hygiene repository, dokumentasi temuan, lalu fitur aplikasi.
- [x] Verification: staged diff dicek dengan `git diff --cached --check` dan status git dirapikan setelah commit.

---

## 📋 Proposed Architecture Plan — Dedicated CBT Engine For Exam Runtime

### Recommendation
- [ ] Split only the live exam runtime into `services/cbt-engine`; do not microservice the whole application.
- [ ] Use a separate VPS for `services/cbt-engine` because the current VPS capacity is considered weak for around 200 concurrent students.
- [ ] Keep `services/core-api` as the main school-domain backend and owner of the main PostgreSQL database.
- [ ] Add a controlled architecture exception: `services/cbt-engine` may own a narrow CBT runtime PostgreSQL database/schema only for live exam state.
- [ ] Treat the live exam path as the critical core: token login, payload delivery, answer save, final submit, heartbeat, telemetry, status, resume, and recovery.
- [ ] Add cache only as a read-path accelerator for immutable or short-lived runtime data; do not make cache the source of truth for answers, submit state, or audit-critical telemetry.
- [ ] Preserve Flutter-facing response envelopes, HTTP status semantics, and BYOD guidance contracts during the extraction.

### Target Topology
```text
Web Admin       -> Core API      -> Main PostgreSQL
PUSAKA Worker   -> Core API      -> Main PostgreSQL

Flutter CBT App -> CBT Engine    -> CBT Runtime PostgreSQL
                       ^
                       |
                 published exam snapshot / result sync
                       |
                    Core API
```

- [ ] Count this as four application runtimes: `web-admin`, `core-api`, `pusaka-worker`, and `cbt-engine`.
- [ ] Count databases separately: main PostgreSQL remains for school/admin domains; CBT runtime PostgreSQL is narrow and exam-only.
- [ ] If hardware is limited, CBT runtime PostgreSQL may start on the same exam VPS as `cbt-engine`, but it must stay logically separate from the main database.

### Current Phase and Target
- **Current phase:** Backend separation already exists: SvelteKit is BFF/UI, Go API owns domain logic and PostgreSQL, Flutter is the CBT client.
- **Target phase now:** Dedicated critical-core service for live CBT runtime, not broad microservices.
- **Reason for moving now:** VPS capacity concern for around 200 concurrent students is a real operational pain signal.
- **Non-target:** Do not split bank soal, academic master data, PUSAKA, website, library, TU, kesiswaan, governance, or reporting into separate services.

### Observed / Expected Pain Signals To Validate
- [ ] Current VPS is not expected to comfortably handle admin/backend load plus around 200 concurrent exam clients.
- [ ] Live exam endpoints receive bursty concurrent traffic from many student devices.
- [ ] Answer-save, submit, heartbeat, and resume flows carry higher correctness risk than normal admin CRUD.
- [ ] BYOD connectivity creates repeated retry, stale-session, and local-sync edge cases.
- [ ] Operator/admin pages should not be degraded by live exam load during exam windows.
- [ ] The exam service still needs load testing to size CPU/RAM, DB pool, cache, and polling intervals before exam season.

### Boundary Assessment
- **Clean boundary:** Flutter already talks through the documented exam API contract; SvelteKit remains BFF-only; PostgreSQL access is backend-owned.
- **Boundary to split:** Admin CBT authoring/scheduling/reporting stays in Core API; live exam runtime moves to CBT Engine.
- **Boundary to avoid leaking:** Do not move general school data ownership into CBT Engine. CBT Engine receives published snapshots and writes runtime exam records only.
- **Policy update needed:** Existing rule "PostgreSQL is owned only by `services/core-api`" must be revised narrowly to "Core API owns the main school database; CBT Engine owns only the CBT runtime database."

### Decision Outcome
- [ ] Extract `services/cbt-engine` as the only new service boundary.
- [ ] Deploy `services/cbt-engine` to a separate VPS before using it in real exams.
- [ ] Keep all non-runtime CBT features in `services/core-api`.
- [ ] Keep PUSAKA worker unchanged as an API client of Core API.

### Incremental Migration Steps
- [ ] Define the `services/cbt-engine` API contract for:
  - exam token login
  - status/payload retrieval
  - answer recording
  - final submit
  - heartbeat
  - telemetry event recording
  - resume/session validation
- [ ] Add a publish/snapshot flow from Core API to CBT Engine:
  - published session metadata
  - participant/token runtime records
  - frozen question payload and options
  - media/audio URL references
  - runtime rules and schedule window
- [ ] Add idempotent result sync from CBT Engine back to Core API:
  - final submit records
  - answers or scored answer references
  - heartbeat/telemetry summary
  - sync cursor and retry status
- [ ] Keep handlers thin in both services: parse request, call service, map explicit HTTP response.
- [ ] Keep SQL explicit through `sqlc`; CBT Engine gets its own narrow query set for runtime tables only.
- [ ] Add focused tests for race-prone and mobile-critical runtime semantics:
  - duplicate submit
  - already-submitted answer/save
  - time-window closed
  - missing participant context
  - device/session mismatch
  - stale heartbeat and resume recovery
  - malformed telemetry rejection
- [ ] Add lightweight runtime metrics/log fields for:
  - answer-save latency
  - submit latency
  - heartbeat failure rate
  - duplicate/conflict response counts
  - status/login error distribution by HTTP code
  - DB query duration on runtime paths
- [ ] Point Flutter exam base URL to CBT Engine after contract parity is verified.
- [ ] Keep Core API exam endpoints as a temporary compatibility bridge or controlled redirect during migration.
- [ ] Run realistic load tests before exam season using expected device count, polling interval, answer frequency, media payload size, and final-submit burst.
- [ ] Run a rehearsal exam with non-production participants before switching real exams to CBT Engine.

### Cache Strategy For CBT Runtime
- [ ] Start with bounded in-process cache inside `services/cbt-engine`, scoped only to CBT runtime read paths.
- [ ] Cache safe read models:
  - published exam payload snapshots
  - session metadata (`title`, start/end time, duration, runtime rules)
  - resolved question list/options/media/audio URLs for a published package/session
  - participant token/session lookup with short TTL
  - rarely changing runtime configuration
- [ ] Keep PostgreSQL as source of truth for critical write paths:
  - answer saves
  - final submit state
  - already-submitted checks
  - device/session mismatch decisions
  - participant authorization
  - durable telemetry/audit events
- [ ] Use payload versioning for cache keys, for example `session_id + payload_version`, so admin republish or session changes naturally invalidate stale payloads.
- [ ] Use short TTL for token/session lookup cache, for example 30-60 seconds, and always revalidate critical state on `answer` and `submit`.
- [ ] Use singleflight-style request coalescing so concurrent logins/status requests for the same session do not stampede PostgreSQL.
- [ ] Prefer cache invalidation by publish/version change over manual broad cache clearing.
- [ ] Do not use cache to hide slow or unsafe SQL; optimize runtime queries and indexes first.
- [ ] Redis is optional for the first 200-student target; evaluate it only if CBT Engine runs multiple instances or load tests show memory-cache misses/stampedes are still painful.

### CBT Runtime Data Ownership
- [ ] Core API remains source of truth for:
  - question bank authoring
  - package building
  - session scheduling and publish decisions
  - academic master data
  - admin reports and long-term records
- [ ] CBT Engine owns runtime-only data for:
  - published exam snapshot copy
  - participant runtime token/session state
  - answers during the live exam
  - final submit status
  - heartbeat records
  - anti-cheat telemetry events
  - sync cursor/status back to Core API
- [ ] Snapshot data in CBT Engine must be immutable per `payload_version`; changes require republish from Core API.
- [ ] Runtime result sync must be idempotent so retries do not duplicate answers, submissions, or telemetry summaries.

### Service Constraints
- [ ] Prefer Go for `services/cbt-engine`; do not add a new language/runtime just for separation.
- [ ] Preserve Flutter-facing response envelopes and error semantics.
- [ ] Keep admin authoring, scheduling, and reporting in Core API unless a separate pain signal appears.
- [ ] Do not introduce cross-service transactions for normal answer/submit flows.
- [ ] Do not let CBT Engine query or mutate the main school database directly.
- [ ] Do not route live Flutter exam traffic through SvelteKit.
- [ ] Use direct Flutter-to-CBT-Engine communication for exam runtime.
- [ ] Use service-to-service auth for Core API publish/sync calls.

### Slice Map
| Slice | Owner | Allowed Dependencies | Notes |
|---|---|---|---|
| CBT Authoring | Core API | Academic master data, employees/users, assets | Bank soal, package builder, review workflow, publication |
| CBT Scheduling/Ops | Core API | Academic master data, rooms, participants | Events, sessions, tokens, seat plans, print ops |
| CBT Runtime Critical Core | CBT Engine | Published snapshots, runtime DB, in-memory cache | Live exam path used by Flutter |
| CBT Result Sync | Core API + CBT Engine | Idempotent service API | Moves completed runtime results back to admin/reporting domain |
| CBT Reporting | Core API | Synced runtime results, academic master data | Reports and exports stay outside live write path |
| Flutter Exam Client | Mobile app | CBT Engine exam API contract only | No direct DB or Core API runtime dependency during exams |

### Cross-Slice Dependency Matrix
| From | To | Allowed? | Rule |
|---|---|---:|---|
| Flutter Exam Client | CBT Runtime Critical Core | Yes | HTTP exam API to CBT Engine only |
| Flutter Exam Client | PostgreSQL | No | Never direct database access |
| Web Admin BFF | CBT Authoring/Scheduling/Reporting | Yes | BFF proxy to Go API only |
| Web Admin BFF | CBT Runtime Critical Core | Limited | Operator/proctor screens should go through Core API or explicit read-only bridge; no runtime business logic |
| Core API | CBT Runtime Critical Core | Yes | Publish immutable snapshots and receive idempotent result sync |
| CBT Runtime Critical Core | Core API | Limited | Sync results/status only; no direct main DB access |
| CBT Runtime Critical Core | Main PostgreSQL | No | CBT Engine must not connect to the main school database |
| CBT Runtime Critical Core | CBT Runtime PostgreSQL | Yes | Runtime-only ownership |
| CBT Authoring/Scheduling | CBT Runtime Critical Core | Yes | Publish/republish runtime snapshot through service API |
| PUSAKA Worker | CBT Runtime Critical Core | No | Separate bounded subsystem |

### Risks and Rollback
- [ ] Risk: service extraction increases deploy/auth/observability burden. Mitigation: keep only one new service, use Go, and keep the API surface narrow.
- [ ] Risk: two databases create sync complexity. Mitigation: immutable publish snapshots, idempotent result sync, sync cursors, and retryable reconciliation jobs.
- [ ] Risk: Flutter contract drift during refactor. Mitigation: keep HTTP handler contract tests around login/status/answer/submit/heartbeat/event.
- [ ] Risk: exam VPS still undersized. Mitigation: load test 200+ clients, reduce polling interval if needed, cache payloads, tune DB pool, and keep media payloads lightweight.
- [ ] Rollback path: keep Core API exam endpoints temporarily available until CBT Engine passes rehearsal; switch Flutter base URL back to Core API if the exam VPS fails before production rollout.

### Validation Metrics
- [ ] P95 answer-save latency during load test.
- [ ] P95 submit latency during final-submit burst.
- [ ] Runtime endpoint 5xx rate.
- [ ] Conflict/error response counts by expected category (`401`, `403`, `404`, `409`).
- [ ] DB pool saturation and slow query count on runtime queries.
- [ ] Heartbeat stale/degraded event rate during BYOD simulation.
- [ ] Admin/API health during live exam load.
- [ ] Cache hit ratio for published payload/status read paths.
- [ ] PostgreSQL query reduction on login/status/payload endpoints during burst load.
- [ ] Cache stampede prevention effectiveness during simultaneous exam start.
- [ ] Result sync lag from CBT Engine back to Core API.
- [ ] Number of unsynced final submissions after rehearsal and after real exam windows.
- [ ] Exam VPS CPU, memory, disk I/O, and network headroom at 200 simulated students.

---

## Architecture Decisions Log
(See previous version for details)

---

## Risk Register
(See previous version for details)
