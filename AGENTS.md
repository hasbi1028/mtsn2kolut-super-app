# Project Policy — MTs Negeri 2 Kolaka Utara

This monorepo powers the academic and operational systems for MTs Negeri 2 Kolaka Utara.

## Runtime Units

| Unit | Path | Stack | Role |
|------|------|-------|------|
| **Web Admin** | `apps/web-admin` | SvelteKit 2 + Svelte 5 (runes) + Tailwind v4 + shadcn-svelte | Admin/guru BFF frontend, session/cookie owner |
| **Core API** | `services/core-api` | Go 1.26 + Chi v5 + sqlc + pgx + PostgreSQL | Domain logic, database owner, scheduler, queue |
| **Pusaka Worker** | `services/pusaka-worker` | TypeScript + Playwright (Chromium) | Async job consumer for PUSAKA attendance automation |
| **Flutter App** *(planned)* | `apps/mobile` *(not yet created)* | Flutter + Serverpod *(not yet created)* | Student-facing CBT exam client |

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
- **Tailwind CSS v4** via `@tailwindcss/vite`. Use utility classes, not custom CSS files.
- **Type safety:** Always use `<script lang="ts">`. Explicit interfaces for props and state. No `any`.
- **Accessibility:** All form labels use `for` + `id`. Run `npm run check` before finalizing.
- **Theme:** Institutional green — `oklch(0.38 0.13 145)`. No dark theme, no purple, no generic SaaS.
- **Dialog imports** use `import * as Dialog` (namespace form) from `$lib/components/ui/dialog` — this avoids naming collisions with bits-ui components and maintains consistency across pages.
- **API proxy** renames fields where documented (e.g., `employee_nama` → `nama`, `employee_nip` → `nip` in Jobs API).
- **Auth:** JWT access + refresh tokens stored as httpOnly cookies. Session handled via SvelteKit hooks.
- **Always run `npm run check` (a11y + types) before finalizing Svelte changes.**
- **CBT UI direction:** educational, institutional, and operator-friendly for MTsN 2 Kolaka Utara. Avoid generic SaaS dashboards for exam operations and printable artifacts.
- **Question bank experimentation:** `/cbt/questions` is the experiment hub for multiple frontend authoring routes. Variants may differ in UX, but they must keep the same backend contract, validation rules, beginner/advance semantics, workflow semantics, and storage model.
- **Question bank authoring uses two UX modes:** `beginner` for quick teacher input with minimal required fields, and `advance` for full blueprint/workflow authoring. Both modes must write to the same backend model and API contract.

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

### 📋 Planned Future Phases
1. **Academic Foundation & RBAC Expansion** — Unified `users` table with many-to-many roles (`admin`, `teacher`, `student`, `staff`, `parent`). Student lifecycle (`active`, `alumni`, `prospective`) and Parent-child linking.
2. **Flutter Student App** — CBT exam client for students on tablet/phone (API sudah ready: `docs/exam-api.md`).
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
