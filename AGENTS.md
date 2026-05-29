# Project Policy — MTs Negeri 2 Kolaka Utara

This monorepo powers the academic and operational systems for MTs Negeri 2 Kolaka Utara.

## Runtime Units

| Unit | Path | Stack | Role |
|------|------|-------|------|
| **Web Admin** | `apps/web-admin` | SvelteKit 2 + Svelte 5 (runes) + Tailwind v4 + shadcn-svelte | Admin/guru BFF frontend, session/cookie owner |
| **Core API** | `services/core-api` | Go 1.26 + Chi v5 + sqlc + pgx + PostgreSQL | Domain logic, database owner, scheduler, queue |
| **Pusaka Worker** | `services/pusaka-worker` | TypeScript + Playwright (Chromium) | Async job consumer for PUSAKA attendance automation |
| **Flutter App** | `apps/mobile` | Flutter | Student-facing CBT exam client |

## Current Baseline — 2026-05-06

- Latest completed roadmap checkpoint: Sprint 96 Documentation Sync. Significant additional modules have shipped since the last AGENTS.md sync — see Completed Sprints for the full ledger.
- Server deployment topology is still 3 VPS targets: frontend, backend, and worker. `apps/mobile` is a student BYOD APK/client, not a VPS runtime.
- `/bank-soal` is the active standalone Bank Soal list route. `/bank-soal/tambah`, `/bank-soal/impor`, and `/bank-soal/verifikasi` own authoring, import, and review/verification UI.
- `/bank-soal/komposer`, `/bank-soal/import`, `/bank-soal/review`, `/cbt/soal*`, `/cbt/bank-soal*`, and `/cbt/questions*` are compatibility redirects to `/bank-soal/*`.
- BFF aliases `/api/bank-soal/*` and `/api/asesmen/*` are available for the UI domain split and are required for new web-admin client code. During the transition, `/api/cbt/*` and `/api/cbt/questions/*` remain live deprecated compatibility contracts to the same Go API semantics and are not removed yet.
- Assessment web-admin navigation uses role-based hub pages: `/asesmen` (launcher), `/asesmen/persiapan` (preparation workflow), `/asesmen/pelaksanaan` (day-of execution), and `/asesmen/hasil` (results). Assessment focuses on paket, kegiatan/event, sesi, pengawasan, pelaksanaan, hasil, aplikasi siswa, and non-test assessment; Bank Soal stands outside the assessment route tree.
- CBT runtime hardening is active: exam tokens are strong random hex values, answer keys/tokens are role-redacted, duplicate submit is explicit, and scoring must not mark unsubmitted participants as submitted.
- PUSAKA worker jobs have stale-running recovery in the backend claim/scheduler path; the worker still must report complete/fail explicitly.
- Library and Inventory are admin/staf scoped in both backend route grouping and SvelteKit navigation/proxy gate.
- Kesiswaan student photo reads are scoped consistently with student visibility for admin/kesiswaan and teacher-owned class access.
- Tata Usaha (persuratan, arsip) and Tata Kelola (governance) modules are live and admin/staf scoped in both backend and BFF.
- Document Cycles module is live: cataloged, scheduled, monitorable document obligations for admin/staf.
- Non-test assessments (penilaian non-tes) live as a separate workflow from CBT; grade sync into rapor is active.
- `findings.md` is now a current review ledger: no open High findings from the 2026-05-01 review remain active.

## Non-Negotiable Architecture Rules

1. **PostgreSQL is owned only by `services/core-api`.** No other runtime unit reads or writes PostgreSQL directly.
2. **All runtime database access must go through Go backend.** SQL must stay explicit and typed through `sqlc`.
3. **`apps/web-admin` is a BFF/proxy only.** It must not read or write PostgreSQL directly, nor import SQLite or Drizzle as runtime storage.
4. **`services/pusaka-worker` is an API client only.** It must not read or write PostgreSQL directly.
5. **Legacy SQLite files** (`backup.db`, `data/pusaka.sqlite`) are one-way import artifacts only. Never active runtime storage.
6. **Monorepo source, but 3 separate server VPS deployments.** One repo does not mean one server; Flutter is distributed as an APK/client.
7. **Safe deploy order:** backend code → migrations → backend restart + health check → frontend → worker.
8. **PUSAKA is a bounded subsystem.** Canonical contracts use `/api/pusaka/*`; legacy runtime aliases such as `/api/jobs`, `/api/attendance`, `/api/schedules`, `/api/settings`, and `/api/worker` are retired.
9. **`employees` stays general.** Employee master data covers all school staff; PUSAKA only manages the eligible subset (`PNS`/`PPPK`) via `pusaka_accounts` and `/pusaka/*` screens.
10. **No employee-scoped PUSAKA aliases in the BFF.** PUSAKA employee operations must proxy only through `/api/pusaka/employees/*`, not `/api/employees/{id}/*`.
11. **Kesiswaan & Tata Usaha modules are role-scoped.** `/api/kesiswaan/*` is owned by `admin` and `kesiswaan`; `guru` may only read kesiswaan data for siswa di kelasnya, including student photo files. `/api/tu/*` is owned by `admin` and `staf`. Cross-role access requires explicit handler-level allow rules, not implicit fallthrough.
12. **Letter numbering is centrally issued.** Outgoing letter numbers must be allocated through the backend `IssueOutgoingLetterNumber` service to preserve `(year, classification)` sequence integrity. Manual override is allowed for legacy import / surat balasan, but must still respect `UNIQUE (nomor_surat)`.
13. **Confidential BK records are role-gated.** Counseling rows flagged `is_confidential = true` must only be readable by `admin` and `kesiswaan`. Guru biasa (tanpa role kesiswaan) tidak boleh melihat catatan rahasia bahkan untuk siswa di kelasnya.
14. **Library and Inventory are staff-scoped operations.** `/library/*`, `/inventory/*`, `/api/library/*`, and `/api/inventory/*` are limited to `admin` and `staf`; sidebar visibility is UX only, not the security boundary.

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
- **Async data-fetching baseline:** new or revised Svelte screens that fetch client-side data should model read operations as `Promise<T>` and render them with the shared `AsyncContent` pattern or direct `{#await promise}` blocks. Avoid scattered `loading`/`error` booleans when pending/fulfilled/rejected UI can be represented by the promise lifecycle.
- **Svelte boundary baseline:** async sections that render remote data should use `<svelte:boundary onerror={handler}>` through `AsyncContent` or an equivalent local boundary so render-time failures get a controlled fallback, retry path, and server/client log context instead of breaking the whole page.
- **Background refresh UX:** auto-refresh screens should avoid blanking already-rendered data on every poll. Keep a small refresh/loading indicator for background fetches, use skeletons for the first load or filter changes, and show retryable error fallback only when the active data request fails.
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
- **Human-readable display names:** user-facing UI must not show UUID/internal IDs or raw technical usernames when a human label is available. Keep IDs/usernames in backend contracts for relations, permission, audit, and filter values, but render `*_display_name`, `nama`, `name`, or the shared `displayName()` helper in tables, cards, filters, badges, dropdowns, and drawers. Internal IDs may appear only in explicit technical/debug/admin detail surfaces labeled as `ID internal`.
- **Flutter exam anti-cheat is BYOD-constrained.** Students use their own Android phones, so `apps/mobile` must maximize deterrence and telemetry with realistic controls such as heartbeat, app-switch/resume events, secure screen, visible sync state, resume re-check gates, local answer safety, and server-facing warning events when sync or resume becomes risky. Do not imply school-managed kiosk guarantees that BYOD cannot enforce.
- **Mobile device fingerprint baseline:** the current Flutter device fingerprint is only a lightweight telemetry/resume hint for BYOD flows. It must not be treated as a strong device identity proof for critical security decisions.
- **Mobile operator-settings baseline:** the exam login screen may keep API base URL override support for trials, but it should stay in an operator/debug surface rather than as a primary student-facing field.
- **CBT UI direction:** educational, institutional, and operator-friendly for MTsN 2 Kolaka Utara. Avoid generic SaaS dashboards for exam operations and printable artifacts.
- **Public site direction:** educational, institutional, and trustworthy for MTsN 2 Kolaka Utara. Public routes must feel like a real school website, not a reused admin dashboard shell.
- **Question bank UI canonical routes:** `/bank-soal` is Daftar Soal, `/bank-soal/tambah` is Tambah/Editor Soal, `/bank-soal/impor` is Impor Soal, and `/bank-soal/verifikasi` is Review/Verifikasi Soal. Old Bank Soal paths (`/bank-soal/komposer`, `/bank-soal/import`, `/bank-soal/review`, `/cbt/soal*`, `/cbt/bank-soal*`, `/cbt/questions*`) are thin compatibility redirects only. New web-admin clients must use `/api/bank-soal/questions/*`; BFF `/api/cbt/questions/*` stays as deprecated compatibility to the unchanged backend semantics.
- **Question bank retired-route guard:** keep test coverage around `/cbt/questions` and `/cbt/soal*` redirect semantics so legacy bookmarks preserve `question_id` and retained modes while retired experiment modes do not re-enter the product.
- **Question bank authoring uses two UX modes:** `beginner` for quick teacher input with minimal required fields, and `advance` for full blueprint/workflow authoring. Both modes must write to the same backend model and API contract.
- **`/bank-soal/tambah` is the Tambah/Editor Soal route** — a dedicated question composer with template quick-start, real-time readiness scoring, quality signals, split preview with KaTeX rendering, RTL toggle for Arabic questions, and localStorage draft autosave. It proxies all data through the existing Go API BFF; no direct DB access, no new backend routes.
- **Assessment navigation uses role-based hub pages.** `/asesmen` is the role-aware launcher (admin/guru/staf). `/asesmen/persiapan` is the step-by-step preparation hub (admin/guru). `/asesmen/pelaksanaan` is the day-of execution hub (admin/guru/staf). `/asesmen/hasil` is the results hub (admin/guru). `/asesmen/paket`, `/asesmen/kegiatan`, `/asesmen/sesi`, `/asesmen/pengawasan`, `/asesmen/aplikasi-siswa`, and `/asesmen/non-tes` are canonical user-facing assessment routes; legacy `/cbt*` routes redirect to the matching `/asesmen*` or Bank Soal target.
- **Non-test assessment authoring is separate from live exam sessions.** `/asesmen/non-tes` and `/api/asesmen/non-test-assessments/*` own the new web-admin penilaian non-tes lifecycle and grade sync; deprecated `/api/cbt/non-test-assessments/*` BFF compatibility remains live for old clients. Do not mix non-test routes with live exam session logic.
- **`/library/*` is the library module** — accessible to `admin` and `staf` roles only. Member data reuses existing `students` and `employees` tables; no separate member table. All forms use beginner/advance mode toggles consistent with the CBT question bank UX pattern.
- **Custom local Dialog component** — `Dialog.Root` accepts only `open: $bindable(bool)` and `children`. `Dialog.Content` and `Dialog.Description` do not accept a `class` prop. Use `bind:open={boolState}` with separate bool state variables; no `onOpenChange` callback.
- **Public shell boundary:** unauthenticated public pages (`/`, `/profil`, `/berita`, `/pengumuman`, `/ppdb`, `/kontak`) must render in the public website shell, while authenticated admin/guru pages continue to use the admin shell.
- **Website editorial admin:** public content management lives under `/website/*` and proxies only to Go API content routes. Do not introduce a separate CMS runtime or client-side persistence.
- **Error handling:** keep a global SvelteKit `+error.svelte` experience for public and admin routes. New pages should rely on centralized error UX before adding page-local fallback banners.
- **Internal error hygiene:** backend 500 responses must return a generic client-safe message; raw internal error details belong in server logs, not API responses.
- **Rate limiting baseline:** sensitive public/auth entrypoints such as login, refresh, public registration, and exam token login should be explicitly rate-limited with a concurrency-safe limiter implementation that respects forwarded client IPs.
- **Trusted proxy baseline:** forwarded client IP headers may influence rate limiting only when the source proxy is listed in `TRUSTED_PROXY_CIDRS`; never trust broad public CIDRs.
- **Upload/file hygiene:** backend-owned uploads must use explicit extension/MIME allowlists, reject scriptable content, and serve files with `X-Content-Type-Options: nosniff` plus safe disposition.

## Worker Architecture Rules

- **Worker is an API client of `services/core-api`.** No direct backend access.
- **Keep communication direct to Go API.** No tunneling through SvelteKit.
- **Use canonical PUSAKA worker routes.** Worker runtime should call `/api/pusaka/worker/*`.
- **Recover stale PUSAKA jobs in backend.** Worker failure reporting remains required, but backend claim/scheduler flows must continue recovering `running` jobs that exceed the stale threshold.
- **Never own business state.** All state lives in PostgreSQL via the Go API.
- **Keep concurrency configurable.** Safe default: 5 consumers.
- **Explicit retry, claim, and failure reporting.** No silent swallowing of errors.
- **Graceful shutdown** on SIGINT/SIGTERM with drain timeout.
- **Heartbeat every 30s** to report consumer count and health status.
- **Worker secret/log hygiene:** `WORKER_API_KEY` must be provided outside local/test environments, logs must redact secrets/user credentials, and worker screenshots/logs must stay in non-public ignored paths with retention.

## Delivery & Operational Rules

- **Prefer safe, staged migration over rewrites.** One sprint at a time.
- **Use multi-agent parallel work for complex development/review.** For broad codebase exploration, review, diagnosis, or multi-area implementation, split work across available OpenCode subagents in parallel, then consolidate findings before editing. Keep direct single-agent work for small, obvious changes.
- **Do not collapse service boundaries for convenience.** Each unit has a clear role.
- **Isolate PUSAKA by contract first.** Only move storage or package ownership further when the boundary is already stable and pain is proven.
- **Preserve deployability to 3 VPS targets.** Every change should be safe for independent rollout.
- **Prioritize correctness, operational safety, and maintainability.**
- **No deployment automation/CD yet** — lightweight CI checks may exist, but deploy remains `git pull` + `make build` + `pm2 restart` per VPS (documented in `deploy/DEPLOY.md`).
- **No Dockerfiles for runtime deployment** — services run natively with PM2 process manager; local database containers are dev-only helpers, not deployment topology.
- **Bun is development-only acceleration.** Bun may be used for local install, dev server, check, and test loops through explicit `*-bun` commands. Do not switch production runtime, PM2 commands, deployment docs, or official lockfile ownership to Bun without a dedicated decision and staging soak test. `package-lock.json` remains the canonical npm lockfile.

## Product Roadmap

### Completed Summary

- **Foundation & Platform (Sprints 1–6A):** CBT backend, academic master data, PUSAKA attendance automation, reporting, Tailwind/shadcn UI foundation, multi-user roles, audit logs, guru scoping, JWT forwarding, health checks, graceful shutdown, structured logging, standardized question bank, adaptive CBT scope, question assets, seat plans, and printable CBT operations.
- **Flutter CBT BYOD (Sprint 7+):** Flutter student exam client with token login, question shell, restore, heartbeat, anti-cheat telemetry, rich content/media/audio handling, connection-health guidance, BYOD operator docs, payload compatibility docs, handler/client tests, and admin views under `/asesmen/aplikasi-siswa`.
- **Gradebook & Rapor (Sprint 11+):** printable report cards, published grade components, readiness summaries, bulk save, quick-fill, assignment finalization, and finalization overview.
- **Library & Public Website (Sprints 15–16B):** library catalog/loans with staff RBAC and no separate member table; public website shell, public content API, editorial admin, featured content, SEO metadata, and cover image upload.
- **Kesiswaan & Tata Usaha (Sprints 17–24):** student affairs extensions, photos, confidential BK records, extracurriculars, mutations, surat masuk/keluar, disposisi, surat keterangan siswa, arsip TU, and role-scoped frontend/backend access.
- **Governance & Document Cycles (Sprints 18–56):** 8 SNP governance, RKT/RKJM execution items, compliance actions, control center, briefing packs, document obligations, audit timeline, readiness/status locks, traceability, external checklist, and CSV export.
- **Academic Ops & Assets (Sprints 34–37):** class journal, timetable, and inventory modules with proper sidebar visibility and RBAC.
- **Assessment & CBT Ops (Sprints 57, 62L–96):** Svelte async boundary baseline, non-test assessments with grade sync, room/proctoring operations, session recap, blueprint coverage, item analysis, revision/reviewer/publish workflows, quality/readiness boards, hardened tokens, seat invariants, event member scope, package-event linkage, assessment hub navigation, CBT smoke checklist, and retired-route redirect guards.

### 📋 Planned Future Phases
1. **Flutter Student App Enhancements** — Stronger offline resilience, richer BYOD-aware anti-cheat telemetry, rich-content rendering, dan safer internal APK distribution di atas `apps/mobile` yang sudah ada.
2. **Real-time Proctoring** — WebSocket-based live monitoring pengawas ruang.
3. **Notifications & Reminders** — WhatsApp/Telegram untuk jadwal ujian, absensi, dan disposisi surat masuk.
4. **PUSAKA Isolation** — Selesai sampai Phase 3 (canonical `/api/pusaka/*`, `pusaka_accounts`, legacy columns dihapus). Deeper package extraction ditangguhkan sampai ada alasan nyata.
5. **CBT Engine Extraction** — `services/cbt-engine` sebagai runtime-only exception, hanya jika rehearsal/load test membuktikan perlu. Jangan ubah aturan PostgreSQL ownership sebelum sprint ini dimulai.

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
| `docs/cbt-smoke-checklist.md` | Manual CBT smoke checklist for admin/guru visibility and runtime integrity |
| `docs/exam-payload-release-template.md` | Release-note template for exam payload changes affecting Flutter |
| `findings.md` | Current review ledger and remaining operational recommendations |
| `deploy/DEPLOY.md` | Step-by-step deployment contract for 3 VPS |
| `apps/mobile/README.md` | Flutter CBT client build/run guide |
| `apps/mobile/RELEASE_CHECKLIST.md` | Internal APK release checklist |
| `apps/mobile/OPERATOR_QUICKSTART.md` | Pengawas/operator quick-start for BYOD trials |
| `apps/mobile/BYOD_TRIAL_PROCEDURE.md` | End-to-end BYOD trial procedure |
| `apps/mobile/DEVICE_TEST_MATRIX.md` | Per-device BYOD compatibility matrix |
| `apps/web-admin/AGENTS.md` | Web admin-specific policy |
| `services/core-api/AGENTS.md` | Core API-specific policy |
| `services/pusaka-worker/AGENTS.md` | Worker-specific policy |

## Known Technical Debt

1. **No automated deployment/CD.** Lightweight CI checks may run, but all deploys are manual `git pull` + build + PM2 restart.
2. **No browser e2e suite yet.** Backend, Svelte unit tests, and Flutter tests exist, but admin/guru CBT visibility still needs staging smoke rehearsal before large exams.
3. **No Dockerfiles for runtime deployment.** Services run bare-metal with PM2; compose/podman helpers are limited to local development database use.
4. **Rate limiting is in-memory per-IP** — does not scale across backend instances.
5. **Audit log entity_id is path, not real entity ID.** Middleware logs URL path (`/api/students/uuid`) into `entity_id`. Sufficient for forensics but not perfect. Future: per-handler structured audit emit.
6. **`INTERNAL_API_KEY` still exists as integration debt surface.** The main user-facing protected/admin routes and CBT asset file route rely on real JWT or exam-token context, but the shared internal key still exists as a helper primitive in middleware and should stay tightly scoped.
7. **CBT print artifacts are HTML-first.** Event cards and berita acara are printable browser views; no PDF rendering service yet.
8. **Seat plan validation is now DB-enforced.** Migration 061 adds a positive-seat constraint and unique `(room_id, seat_no)` index. Preflight SQL checks required before running migration on environments with existing data.
9. **CBT Engine extraction is only planned.** Do not change PostgreSQL ownership rules or deploy topology until `services/cbt-engine` is implemented as a controlled runtime-only exception.
