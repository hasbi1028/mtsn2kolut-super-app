# Project Policy — MTs Negeri 2 Kolaka Utara

This monorepo powers the academic and operational systems for MTs Negeri 2 Kolaka Utara.

## Runtime Units

| Unit | Path | Stack | Role |
|------|------|-------|------|
| **Web Admin** | `apps/web-admin` | SvelteKit 2 + Svelte 5 (runes) + Tailwind v4 + shadcn-svelte | Admin/guru BFF frontend, session/cookie owner |
| **Core API** | `services/core-api` | Go 1.26 + Chi v5 + sqlc + pgx + PostgreSQL | Domain logic, database owner, scheduler, queue |
| **Pusaka Worker** | `services/pusaka-worker` | TypeScript + Playwright (Chromium) | Async job consumer for PUSAKA attendance automation |
| **Flutter App** | `apps/mobile` | Flutter | Student-facing CBT exam client |

## Non-Negotiable Architecture Rules

1. **PostgreSQL is owned only by `services/core-api`.** No other runtime unit reads or writes PostgreSQL directly.
2. **All runtime database access must go through Go backend.** SQL must stay explicit and typed through `sqlc`.
3. **`apps/web-admin` is a BFF/proxy only.** It must not read or write PostgreSQL directly, nor import SQLite or Drizzle as runtime storage.
4. **`services/pusaka-worker` is an API client only.** It must not read or write PostgreSQL directly.
5. **Legacy SQLite files** (`backup.db`, `data/pusaka.sqlite`) are one-way import artifacts only. Never active runtime storage.
6. **Monorepo source, but 3 separate VPS deployments.** One repo does not mean one server.
7. **Safe deploy order:** backend code → migrations → backend restart + health check → frontend → worker.
8. **PUSAKA is a bounded subsystem.** Canonical contracts use `/api/pusaka/*`; legacy runtime aliases such as `/api/jobs`, `/api/attendance`, `/api/schedules`, `/api/settings`, and `/api/worker` are retired.
9. **`employees` stays general.** Employee master data covers all school staff; PUSAKA only manages the eligible subset (`PNS`/`PPPK`) via `pusaka_accounts` and `/pusaka/*` screens.
10. **No employee-scoped PUSAKA aliases in the BFF.** PUSAKA employee operations must proxy only through `/api/pusaka/employees/*`, not `/api/employees/{id}/*`.

## Backend Architecture Rules

- **Handlers are thin:** parse input → call service → map HTTP response. No business logic.
- **Services own business rules** and orchestration. Use transactions for multi-step writes.
- **SQL lives in `services/core-api/db/queries`.** One SQL file per domain aggregate.
- **Generated code lives in `services/core-api/internal/repository/postgres`.** Never edit by hand — regenerate with `make db-sqlc`.
- **No ORMs.** `sqlc` is the required typed-SQL pattern. No GORM, no Ent, no Drizzle.
- **Migrations are the only schema evolution path.** No ad hoc `ALTER TABLE` in production.
- **Error mapping must be explicit.** Use sentinel errors from `internal/domain/errors.go`.
- **Test with interfaces.** Services that touch the database should accept store interfaces for testability.
- **Middleware stack:** RequestID → RequestLog → Recoverer → Content-Type → Auth/RateLimit.

## Frontend Architecture Rules

- **SvelteKit is UI/BFF only.** No backend business logic leaks into frontend code.
- **Use `shadcn-svelte` (bits-ui) for all new UI primitives.** Do not add one-off component patterns.
- **Use `sonner` toast via the shared `ui/sonner` wrapper for notifications.** Do not add per-page local toast banners for new work.
- **Loading UX baseline:** always use the combination of global route progress, section/page skeletons, and explicit button loading states for fetch and mutation UX. For new or revised screens, prefer:
  - global route progress for navigation/page transitions,
  - skeletons for initial section/page data fetching,
  - loading buttons for submit/refresh/export/mutation actions.
  Avoid falling back to plain loading text on primary screens when a skeleton is practical.
- **Tailwind CSS v4** via `@tailwindcss/vite`. Use utility classes, not custom CSS files.
- **Type safety:** Always use `<script lang="ts">`. Explicit interfaces for props and state. No `any`.
- **Accessibility:** All form labels use `for` + `id`. Run `npm run check` before finalizing.
- **Theme:** Institutional green — `oklch(0.38 0.13 145)`. No dark theme, no purple, no generic SaaS.
- **Dialog imports** use `import * as Dialog` (namespace form) from `$lib/components/ui/dialog` — this avoids naming collisions with bits-ui components and maintains consistency across pages.
- **API proxy** renames fields where documented (e.g., `employee_nama` → `nama`, `employee_nip` → `nip` in Jobs API).
- **Auth:** JWT access + refresh tokens stored as httpOnly cookies. Session handled via SvelteKit hooks.
- **Auth hardening baseline:** backend is the source of truth for password policy. Minimum password length is 8, weak passwords are rejected server-side, suspended accounts must return controlled auth errors, and `SeedAdmin` must not overwrite an existing admin password during normal startup.
- **BFF auth forwarding baseline:** authenticated SvelteKit proxy routes must forward the real user JWT to the Go API by default. `X-Internal-Key` is reserved for explicit internal/public helper use, not as a silent fallback for authenticated BFF traffic.
- **Refresh session baseline:** refresh tokens are session-backed and revocable. Login creates an auth session, refresh rotates it, and logout must revoke it in the backend before clearing browser cookies.
- **Per-user token invalidation:** user access/refresh tokens must carry a per-user auth version, not a global shared version. “Logout all sessions” should revoke all refresh sessions for that user and bump only that user's auth version.
- **Session management UX:** authenticated users may list their own active sessions and revoke individual sessions from the settings screen. JWT `sub` is the user ID; human-readable username travels in a separate claim.
- **Session metadata:** auth sessions should preserve human-useful client metadata (`device_label`, `ip_address`, `user_agent`) so the settings screen can show recognizable device entries instead of opaque token IDs.
- **Session labeling:** users may rename their own auth session `device_label` from the settings screen; session labels remain user-owned metadata, not a separate device registry.
- **Auth audit events:** successful `login`, `refresh`, `logout`, `logout-all`, and per-session revoke should emit structured audit events with explicit auth-focused action names, not rely only on generic method/path middleware logs.
- **Access token session validation:** authenticated JWT requests must validate both per-user auth version and the referenced live auth session (`ssid`) unless the route is using the explicit internal-key bypass.
- **Always run `npm run check` (a11y + types) before finalizing Svelte changes.**
- **Flutter exam anti-cheat is BYOD-constrained.** Students use their own Android phones, so `apps/mobile` must maximize deterrence and telemetry with realistic controls such as heartbeat, app-switch/resume events, secure screen, visible sync state, resume re-check gates, local answer safety, and server-facing warning events when sync or resume becomes risky. Do not imply school-managed kiosk guarantees that BYOD cannot enforce.
- **CBT UI direction:** educational, institutional, and operator-friendly for MTsN 2 Kolaka Utara. Avoid generic SaaS dashboards for exam operations and printable artifacts.
- **Public site direction:** educational, institutional, and trustworthy for MTsN 2 Kolaka Utara. Public routes must feel like a real school website, not a reused admin dashboard shell.
- **Question bank experimentation:** `/cbt/questions` is the experiment hub for multiple frontend authoring routes. Variants may differ in UX, but they must keep the same backend contract, validation rules, beginner/advance semantics, workflow semantics, and storage model.
- **Question bank authoring uses two UX modes:** `beginner` for quick teacher input with minimal required fields, and `advance` for full blueprint/workflow authoring. Both modes must write to the same backend model and API contract.
- **`/cbt/soal` is the Komposer Soal route** — a dedicated question composer with template quick-start, real-time readiness scoring, quality signals, split preview with KaTeX rendering, RTL toggle for Arabic questions, and localStorage draft autosave. It proxies all data through the existing Go API BFF; no direct DB access, no new backend routes.
- **`/library/*` is the library module** — accessible to `admin` and `staf` roles only. Member data reuses existing `students` and `employees` tables; no separate member table. All forms use beginner/advance mode toggles consistent with the CBT question bank UX pattern.
- **Custom local Dialog component** — `Dialog.Root` accepts only `open: $bindable(bool)` and `children`. `Dialog.Content` and `Dialog.Description` do not accept a `class` prop. Use `bind:open={boolState}` with separate bool state variables; no `onOpenChange` callback.
- **Public shell boundary:** unauthenticated public pages (`/`, `/profil`, `/berita`, `/pengumuman`, `/ppdb`, `/kontak`) must render in the public website shell, while authenticated admin/guru pages continue to use the admin shell.
- **Website editorial admin:** public content management lives under `/website/*` and proxies only to Go API content routes. Do not introduce a separate CMS runtime or client-side persistence.
- **Error handling:** keep a global SvelteKit `+error.svelte` experience for public and admin routes. New pages should rely on centralized error UX before adding page-local fallback banners.

## Worker Architecture Rules

- **Worker is an API client of `services/core-api`.** No direct backend access.
- **Keep communication direct to Go API.** No tunneling through SvelteKit.
- **Use canonical PUSAKA worker routes.** Worker runtime should call `/api/pusaka/worker/*`.
- **Never own business state.** All state lives in PostgreSQL via the Go API.
- **Keep concurrency configurable.** Safe default: 5 consumers.
- **Explicit retry, claim, and failure reporting.** No silent swallowing of errors.
- **Graceful shutdown** on SIGINT/SIGTERM with drain timeout.
- **Heartbeat every 30s** to report consumer count and health status.

## Delivery & Operational Rules

- **Prefer safe, staged migration over rewrites.** One sprint at a time.
- **Do not collapse service boundaries for convenience.** Each unit has a clear role.
- **Isolate PUSAKA by contract first.** Only move storage or package ownership further when the boundary is already stable and pain is proven.
- **Preserve deployability to 3 VPS targets.** Every change should be safe for independent rollout.
- **Prioritize correctness, operational safety, and maintainability.**
- **No CI/CD yet** — deploy via `git pull` + `make build` + `pm2 restart` per VPS (documented in `deploy/DEPLOY.md`).
- **No Dockerfiles** — services run natively with PM2 process manager.

## Product Roadmap

### ✅ Completed (Sprints 1–3)
- **CBT Foundation:** Full backend (migrations, handlers, services, sqlc), question bank (PG + Essay), package builder, exam events, exam sessions, proctoring, room management, token generation, essay grading.
- **Academic Master Data:** Academic years, school classes, subjects, class-subject assignments, students (with class assignment).
- **Attendance Automation:** PUSAKA integration via Playwright worker, checkin/checkout, scrape, queue management, job history, monthly summaries.
- **Reports & Monitoring:** Date range filters, CSV export with WITA localization, event-wide results, academic dashboard stats.
- **UI Foundation:** Tailwind v4 + shadcn-svelte, institutional green theme, sidebar navigation, mobile-friendly.

### ✅ Sprint 4 — Multi-user & Roles (Done)
- **Users table** (`users`) dengan role enum `admin/guru` dan `employee_id` link.
- **Audit log table** (`audit_logs`) — semua mutasi tercatat lewat middleware.
- **SeedAdmin** lewat `service.Auth.SeedAdmin()` (single source of truth, bukan SQL seed).
- **Backend role gate:** `mw.RequireAdmin(internalKey)` — JWT role check, internal-key bypass untuk BFF.
- **Audit middleware** (`mw.Audit`) — auto-log semua POST/PUT/PATCH/DELETE dengan `user_id`, `path`, `method`, `status`.
- **BFF role gate:** SvelteKit `hooks.server.ts` block guru dari `/employees`, `/academic`, `/attendance`, `/settings`, `/api/users`, dst.
- **Frontend:** halaman `/settings/users` (CRUD pengguna), `/settings/audit-logs` (view audit trail).
- **Sidebar:** menu otomatis ter-filter berdasarkan `user.role`.
- **JWT claims** include `role`, `uid`, `eid` — frontend `getUserFromToken()` decode.

### ✅ Sprint 5 — Guru Data Scoping & Hardening (Done)
- **JWT forwarding** — 33 BFF proxy routes now forward `Authorization: Bearer` from cookies, audit log has real `user_id`
- **Guru data scoping** — 4 new SQL queries filter CBT sessions/results/participants/students by `class_subject_assignments.teacher_employee_id`
- **Dashboard guru** — role-aware dashboard with teacher stats (active sessions, ungraded essays, students, subjects)
- **Graceful shutdown** — `signal.NotifyContext` + `srv.Shutdown` with 15s timeout
- **Structured logging** — `log/slog` JSON handler throughout backend
- **Token refresh** — `handleFetch` hook auto-refreshes on 401 mid-session
- **Legacy components migrated** — Nav (no-op), AttendanceTable + JobTable → shadcn, dark-theme CSS removed
- **Audit log retention** — `DeleteOldAuditLogs` (>90 days) via scheduler cleanup
- **Health endpoint** — DB pool stats, `runtime.Version()`, scheduler/worker heartbeat detail
- **`make check`** — includes `go vet` + `npm audit --audit-level=high`

### ✅ Sprint 6A — CBT Standardized Bank, Cohort Scope, and Print Ops (Done)
- **Question bank standardization** — rich content (`stem_html`, `stimulus_html`, LaTeX), curriculum metadata (`CP/TP/KD`), workflow review, asset storage, versioning fields.
- **Question asset management** — backend-owned asset upload for image/audio/PDF through `cbt_question_assets`.
- **Adaptive session scope** — `scope_type`, `scope_ref`, `mix_policy`, `assignment_mode`, `allow_cross_grade`, `is_special_event` live on `cbt_exam_sessions`.
- **Adaptive mix policy rule** — `class -> same_class`, `grade -> same_grade`, `school/custom -> mixed_scope`.
- **Seat plan support** — participants can store `seat_no`; sessions support auto/manual seat assignment by room.
- **Print operations** — event exam cards and session minutes/berita acara are printable HTML routes, not PDF generators.

### ✅ Sprint 7 — Flutter CBT App MVP (Done)
- **Initialized mobile client** — `apps/mobile` now exists as the Flutter student exam client baseline.
- **Exam token login** — students can log in with token + API base URL against `docs/exam-api.md`.
- **Question shell** — supports pilihan ganda and uraian flows with per-question save, countdown, heartbeat, status refresh, and final submit.
- **Session restore baseline** — active exam snapshot (token, base URL, current question index, local answers) is persisted locally and can be restored on app reopen.
- **Anti-cheat baseline** — BYOD-aware deterrence via Android secure screen (`FLAG_SECURE`), back-navigation blocking, lifecycle-driven `app_switch` event logging, resume re-check gating, visible sync-state chips, local pending-answer sync protection, and explicit warning events for repeat resume / blocked submit conditions.
- **Question rendering baseline** — mobile renderer is ready for richer payloads (`stimulus_html`, `stem_html`) with safe text normalization instead of assuming only plain `question_text`.
- **Session recovery UX baseline** — local exam snapshots should preserve human-readable metadata (student, session title, room) so restore flows remain understandable during reconnects, and completed exams should transition to a dedicated finish screen rather than leaving students on the active question shell.
- **Session schedule UX baseline** — local exam snapshots should also preserve schedule metadata (start, end, duration) so restore and completion screens remain informative even while reconnecting.
- **Lightweight media support baseline** — Flutter exam screens may render simple remote image media from the exam payload (`stimulus_media_url`, `stem_media_url`) without introducing a heavier media stack first.
- **Lightweight audio support baseline** — Flutter exam screens may also play simple remote audio prompts from payload URLs (`stimulus_audio_url`, `stem_audio_url`) using an in-app player before considering richer media workflows.
- **Restore failure UX baseline** — when a cached exam session can no longer be restored, mobile should show a dedicated explanation screen with the last known session metadata instead of dropping students into a generic login error.
- **Connection freshness UX baseline** — exam shells should expose human-readable last-contact / last-failure timestamps so BYOD connectivity issues are easier for students and pengawas to interpret.
- **Repeated-connection-failure UX baseline** — if heartbeat or sync fails repeatedly, mobile should escalate from passive status text to a visible warning panel with an explicit retry action.
- **Degraded-mode UX baseline** — if repeated sync failures cross a BYOD risk threshold, manual submit should be held until status refresh or pending-answer sync recovers enough to trust the session again.
- **Per-question audio state baseline** — when question audio URLs are present, the exam shell should show whether audio for that question has already been played so students and pengawas have clearer progress cues.
- **Internal APK distribution baseline** — `apps/mobile/README.md` should document local run, analyze/test, release APK build, and realistic BYOD distribution guidance for internal school trials.
- **Restore health baseline** — persisted mobile exam snapshots should preserve lightweight connection-health metadata so restore cards and restore-failed screens can show whether the previous session was recently stable or repeatedly degraded.
- **Internal release checklist baseline** — `apps/mobile/RELEASE_CHECKLIST.md` should exist for operator-facing APK trial preparation, verification, and BYOD field guidance.
- **Connection-legend baseline** — the mobile exam shell should expose a more visual connection-health summary, and the app should provide a simple status-guide screen so students and pengawas can interpret `Tersambung`, `Lokal`, `Gangguan`, and `Menurun` consistently.
- **Heartbeat-quality baseline** — mobile connection status should not rely only on the latest error flag; it should also consider stale last-contact timing so the shell can surface a softer `Waspada` state before a session is fully degraded.
- **Supervisor-intervention baseline** — if `Waspada` persists too long without fresh server contact, the mobile shell should escalate to a visible pengawas-intervention panel, show how long the session has stayed stale, and emit a one-time server warning event for that stale episode.
- **Escalated-stale baseline** — if that stale `Waspada` condition continues beyond a second threshold, the panel should harden into a stronger “intervene now” state and emit a distinct escalation warning event so backend audit can distinguish soft attention from urgent BYOD intervention.
- **Visible-threshold baseline** — stale/intervention panels should expose the relevant escalation threshold in human terms so pengawas can tell whether the device is still in early attention mode or has crossed into urgent intervention territory.
- **Operator quick-start baseline** — `apps/mobile/OPERATOR_QUICKSTART.md` should exist as a short non-technical guide for pengawas/operator during BYOD field trials.
- **BYOD trial-procedure baseline** — `apps/mobile/BYOD_TRIAL_PROCEDURE.md` should exist as the end-to-end school trial procedure covering operator, pengawas, siswa, disturbance simulation, and submit readiness.
- **Backend payload-check baseline** — `docs/exam-api.md` should include a mobile compatibility checklist so exam-payload changes in the Go backend are reviewed against the live Flutter contract before release.
- **Exam payload contract baseline** — backend exam login/status handlers must preserve the mobile-critical fields that Flutter actively parses (`session.title`, `is_submitted`, rich stem/stimulus fields, and resolved media/audio URLs), and this contract should be guarded by backend tests rather than docs alone.
- **Proxy URL baseline** — exam asset/media URLs returned to Flutter must stay correct when the API sits behind reverse proxies; handler-level tests should cover `X-Forwarded-Proto` / `X-Forwarded-Host` absolutization instead of assuming direct localhost access.
- **Exam response-wrapper baseline** — exam handler tests should cover the final JSON envelope (`api.OK`) for login/status, not only service return values, so mobile payload regressions are caught at the HTTP layer too.
- **Exam error-semantics baseline** — handler tests should also lock the mobile-relevant error contract for exam login/status (`404` token missing, `403` session inactive, `409` device mismatch, `401` missing participant context) so restore/login flows do not silently drift.
- **Answer/submit error baseline** — exam handler tests should also lock `answer` and `submit` conflict/forbidden semantics (`409` already submitted, `403` time window closed) because Flutter relies on those distinctions for retry, local-save, and final-submit guidance.
- **Flutter exam-status UX baseline** — the mobile client should translate backend exam status codes into role-appropriate student/pengawas guidance, especially for login/restore (`404/403/409`) and answer/submit flows (`403/409`), instead of showing one generic server message for every failure.
- **Flutter message-mapping baseline** — these exam status-code mappings should live in testable helper logic, not only inside widget state methods, so BYOD guidance copy can be verified with fast Flutter tests.
- **Persistent exam-guidance baseline** — for the most important `403/409` exam states, mobile should elevate from plain text into a persistent guidance panel with an explicit urgency level so siswa and pengawas can distinguish “wait”, “verify”, and “stop retrying”.
- **Login/restore guidance baseline** — that persistent guidance pattern should also appear on the early mobile entry flow (token login and restore-failed states), not only inside the running exam shell, so `403/409` conditions are explained before students retry blindly.
- **Entry-flow widget-test baseline** — mobile guidance panels for login and restore-failed flows should be covered by widget tests, not only helper-unit tests, so the actual rendered UX for `403/409` states stays stable.
- **Exam-shell widget-test baseline** — the persistent guidance panel inside the running exam shell should also have widget coverage for warning/danger states so shell-level `403/409` guidance remains visible after UI refactors.
- **Connection-panel widget-test baseline** — the mobile shell’s BYOD risk panels (`Koneksi perlu diperhatikan`, `Perlu intervensi pengawas`, `Mode koneksi menurun aktif`) should also have widget coverage driven from restored snapshot state, so operational connection UX stays stable as the shell evolves.
- **Resume-overlay widget-test baseline** — the “Mode ujian diamankan” overlay should have widget coverage for both the pre-check state and the in-progress resume-check state, because BYOD recovery UX is a critical safety surface for students and pengawas.
- **Restore/audio widget-test baseline** — Flutter should also keep widget coverage for restore-health chips and per-question audio state (`belum diputar` / `sudah diputar`) so media-assisted exam UX and restore context remain stable during UI iteration.
- **Login-restore/media widget-test baseline** — the token-entry screen should have widget coverage for its cached restore card, and the exam shell should have widget coverage for lightweight remote media surfaces, so the two most visible BYOD helper surfaces stay stable.
- **Rich-content widget-test baseline** — the Flutter exam shell should also have widget coverage for rendered `stimulus_html` and `stem_html` content so backend rich-content payload improvements do not silently regress the readable student surface.
- **Sync-chip widget-test baseline** — the Flutter exam shell should keep widget coverage for the main sync-chip labels (`Tersambung`, `Lokal`, `Waspada`, `Menurun`) because those app-bar signals are the fastest operational read for BYOD health during exams.
- **Transient sync-chip baseline** — widget coverage should also include the transient sync-chip states `Sinkron` and `Cek Ulang`, because they communicate active recovery and resume-gate behavior that is easy to break during shell refactors.
- **Error sync-chip baseline** — widget coverage should also include the `Gangguan` app-bar sync-chip state, because it is distinct from `Menurun` and should stay visible when the shell hits a non-degraded server error.
- **Restore-health label baseline** — login and restore-failed Flutter entry surfaces should keep widget coverage for all restore-health labels (`Terakhir stabil`, `Pernah terganggu`, `Belum ada riwayat koneksi`, `Perlu perhatian koneksi`) so operator-facing reconnect context stays stable during BYOD UX iteration.
- **Finish/status-guide widget baseline** — the Flutter finish screen and the in-app BYOD status-guide screen should keep widget coverage, because both are operator-facing explanatory surfaces that are easy to regress during visual cleanup even though they are operationally important.
- **Exam-format helper baseline** — Flutter formatting helpers for schedule strings, restore-health labels, and restore clocks should keep direct unit coverage so small BYOD wording/format regressions are caught without depending only on widget trees.
- **Rich-text helper baseline** — the Flutter rich-question text normalizer should keep direct unit coverage for paragraph breaks, list bullets, entity decoding, and whitespace cleanup, because backend exam payloads now rely on richer HTML-like content even in the BYOD MVP.
- **Exam telemetry auth baseline** — `heartbeat` and `event` endpoints must keep explicit unauthorized behavior at the handler level; BYOD telemetry routes should not silently accept missing participant context.
- **Exam telemetry success baseline** — `heartbeat` and `event` handlers should also keep their wrapped success payloads stable (`data.status = ok|recorded`) because the mobile client treats these as lightweight control acknowledgements during BYOD operation.
- **Answer/submit success-envelope baseline** — exam handler tests should also lock the wrapped success payloads for `answer` and `submit` (`data.status = recorded|submitted`) so the Flutter client never depends on undocumented envelope drift for its core save/finalize flow.
- **Exam bad-request baseline** — handler tests should also lock explicit `400` contracts for malformed exam requests (`invalid json`, missing login token, invalid `question_id`) because mobile and future clients need deterministic guidance for local validation vs. server rejection.
- **Exam mutation-auth baseline** — exam mutating handlers such as `answer` and `submit` should have explicit `401 unauthorized` coverage too, so participant-context regressions are caught before they silently affect the most important BYOD write paths.
- **Telemetry error baseline** — exam telemetry handlers should also lock their unexpected-error `500` behavior, not only success and unauthorized paths, so heartbeat/event failures stay explicit to operators and mobile clients.
- **Status error baseline** — `GET /api/exam/status` should also have explicit unexpected-error `500` coverage, because mobile resume and BYOD health logic depend heavily on that one read path staying deterministic.
- **Exam-payload release-template baseline** — `docs/exam-payload-release-template.md` should exist as the structured release-note template for backend payload changes that may affect the Flutter exam client.
- **Device-matrix baseline** — `apps/mobile/DEVICE_TEST_MATRIX.md` should exist as the structured per-vendor/per-device scorecard for BYOD field trials so hardware issues are tracked systematically, not only via ad-hoc notes.
- **Admin BYOD summary baseline** — `/cbt/byod` should exist in the web admin as a compact operator-facing summary of mobile status meanings, submit readiness, and trial references without requiring a new backend service.
- **Admin BYOD matrix baseline** — `/cbt/byod/matrix` should exist in the web admin as a readable companion view for the per-device BYOD test matrix, again without requiring a backend feature or new persistence.
- **Admin BYOD release baseline** — `/cbt/byod/release` should exist in the web admin as the operator-facing readiness summary for backend payload checks, mobile verification, and rollout preparation.

### ✅ Sprint 15 — Library System (Done)
- **Library schema** — `library_books` and `library_loans` tables (migration 027). No separate member table; loans reference existing `students` and `employees` via FK.
- **Loan rules** — max 3 active loans per member enforced at service layer; `tersedia` decremented/incremented atomically around loan/return.
- **Denda calculation** — computed at return time in Go service (`ceil(overdue_hours/24) × denda_per_hari`), stored in `denda_total`. `denda_lunas` tracks cash settlement separately.
- **RBAC** — `/library/*` and `/api/library/*` accessible to `admin` and `staf` roles only.
- **Beginner/advance mode** — both book catalog form and loan form have mode toggles consistent with CBT question bank UX pattern.
- **Frontend** — Dashboard `/library`, Katalog `/library/books`, Peminjaman `/library/loans`. Sidebar "Perpustakaan" group visible to `admin` + `staf`.

### ✅ Sprint 16 — Public Website Foundation (Done)
- **Public website shell** — unauthenticated routes for `/`, `/profil`, `/berita`, `/pengumuman`, `/ppdb`, and `/kontak` use a dedicated public shell while authenticated users keep the admin dashboard experience on `/`.
- **Website content domain** — `website_contents` table with `page`, `post`, and `announcement` content kinds, plus published/draft lifecycle, slugging, and backend HTML sanitization.
- **Public endpoints** — canonical read-only routes under `/api/public/site/*` for posts, announcements, and pages.
- **Editorial admin** — admin-only management screens under `/website`, `/website/posts`, `/website/announcements`, and `/website/pages`.
- **Homepage aggregation** — public homepage highlights PPDB, school profile, latest posts, and recent announcements.
- **Global error UX** — centralized SvelteKit `+error.svelte` for 403/404/500 style failures across public and admin shells.

### ✅ Sprint 16B — Public Website Polish (Done)
- **Featured flags** — `is_featured` on `website_contents` (migration 033); `ListFeaturedWebsiteContents` query + handler (GET `/api/public/site/posts/featured`); published list orders featured first.
- **SEO metadata** — `meta_title` and `meta_description` columns; `og:title`, `og:description`, `og:image`, and `description` meta tags on `/berita/[slug]` and `/pengumuman/[slug]`.
- **Cover image upload** — `WebsiteMedia` handler uploads images to `data/website-media/` (md5 filename, MIME validation, path-traversal guard); BFF proxy forwards JWT; `WebsiteContentManager` shows upload button, image preview, and file picker.
- **Editorial UX** — featured badge in content list, SEO section with character counters.

### ✅ Sprint 11C — Rapor Print View (Done)
- **Printable report card** — `/grades/rapor` HTML-first print layout using existing `/api/grades` endpoint; assignment selector, school header, grade table with color-coded scores (green ≥80, amber ≥65, red <65), class average footer, signature area shown only on print.
- **Sidebar entry** — "Cetak Rapor" under Akademik group, visible to `admin` and `guru`.

### 📋 Planned Future Phases
1. **Academic Foundation & RBAC Expansion** — Unified `users` table with many-to-many roles (`admin`, `teacher`, `student`, `staff`, `parent`). Student lifecycle (`active`, `alumni`, `prospective`) and Parent-child linking.
2. **Flutter Student App Enhancements** — Build on top of the initialized CBT exam client in `apps/mobile` with stronger offline resilience, richer BYOD-aware anti-cheat telemetry, richer rich-content rendering, and safer internal distribution / packaging.
3. **Real-time Proctoring** — WebSocket-based live monitoring.
4. **Notifications & Reminders** — WhatsApp/Telegram for exam schedules, attendance.
5. **Raport / Grade Management** — Academic grading, report cards integrated with CBT scores.
6. **Schedule & Timetable** — Class schedules, teacher assignments UI.
7. **PUSAKA Isolation** — completed through Phase 3 for current scope: canonical `/api/pusaka/*`, `pusaka_accounts` as integration owner, and legacy `employees.pusaka_*` columns removed. Deeper package extraction is deferred until code churn justifies it.

## RBAC & User Lifecycle Policy

1. **Role Separation**: 
   - **Admin**: System-wide configuration and full access.
   - **Teacher**: Academic management, grading, and subject-specific CBT.
   - **Student**: Exam taking, schedule viewing, and report card access.
   - **Staff**: Operational administration (finance, inventory, scheduling).
   - **Parent**: Monitoring children's progress (attendance, grades, billing).
2. **Identity vs Entity**: Auth data (`users`) must be decoupled from profile data. A single identity can hold multiple roles (e.g., a teacher who is also a parent).
3. **Lifecycle States**:
   - **Prospective**: Limited access, used for PPDB (new student admission).
   - **Active**: Full access based on role.
   - **Alumni/Mutated**: Read-only access to historical records (grades, certificates), blocked from active sessions.
4. **Archiving**: Never delete users with historical academic data (grades/attendance). Use `is_active = false` or `status = 'archived'`.

## Documentation Map

| File | Purpose |
|------|---------|
| `AGENTS.md` (this file) | Global project policy, architecture, roadmap |
| `PLAN.md` | Current sprint plan and active tasks |
| `docs/architecture.md` | System architecture, ownership, CBT direction |
| `docs/deployment.md` | Deployment topology and safe order |
| `docs/ui-guidelines.md` | UI/UX direction and visual guidelines |
| `docs/exam-api.md` | Flutter exam API documentation |
| `deploy/DEPLOY.md` | Step-by-step deployment contract for 3 VPS |
| `apps/web-admin/AGENTS.md` | Web admin-specific policy |
| `services/core-api/AGENTS.md` | Core API-specific policy |
| `services/pusaka-worker/AGENTS.md` | Worker-specific policy |

## Known Technical Debt

1. **No automated CI/CD.** All deploys are manual `git pull` + build + PM2 restart.
2. **Low test coverage.** Only 4 service test files (auth, job, scheduler, setting). No integration tests, no e2e tests.
3. **No Dockerfiles.** Services run bare-metal with PM2. No containerized deployment option.
4. **Rate limiting is in-memory per-IP** — does not scale across backend instances.
5. **Audit log entity_id is path, not real entity ID.** Middleware logs URL path (`/api/students/uuid`) into `entity_id`. Sufficient for forensics but not perfect. Future: per-handler structured audit emit.
6. **BFF uses `X-Internal-Key` for all admin calls** — backend can't see real user. JWT claims propagated only when frontend forwards JWT (currently bypassed). For higher-fidelity audit, BFF should forward user JWT in `Authorization` header instead of internal key for non-public endpoints.
7. **CBT print artifacts are HTML-first.** Event cards and berita acara are printable browser views; no PDF rendering service yet.
8. **Seat plan validation is still light.** Current backend stores `seat_no` and room assignment, but does not yet enforce uniqueness per `(room_id, seat_no)` at the database level.
