# Architecture

## System Shape

This repository is a monorepo with three deployable server runtime units and one mobile client:

- `apps/web-admin`
- `services/core-api`
- `services/pusaka-worker`
- `apps/mobile`

The monorepo is for source organization, not for collapsing runtime topology. The three server units still deploy to separate VPS targets; `apps/mobile` is built and distributed as an internal Android APK for student CBT.

## Current Baseline

As of 2026-05-06:

- `services/core-api` remains the only PostgreSQL owner for the implemented system.
- `apps/web-admin` remains a BFF/UI layer and must not read or write database state directly.
- `services/pusaka-worker` talks directly to `/api/pusaka/worker/*` and never owns business state.
- `apps/mobile` talks to the exam API using exam tokens and BYOD-oriented safeguards.
- `/bank-soal`, `/bank-soal/tambah`, `/bank-soal/impor`, and `/bank-soal/verifikasi` are the active question-bank UI routes. `/bank-soal/komposer`, `/bank-soal/import`, `/bank-soal/review`, `/cbt/soal*`, `/cbt/bank-soal*`, and `/cbt/questions*` are legacy compatibility redirects only.
- `/asesmen` and its subroutes are the active web-admin assessment UI. Legacy operational `/cbt*` routes redirect to `/asesmen*` or Bank Soal according to context.
- Active web-admin Bank Soal UI fetches use `/api/bank-soal/*`; active assessment UI fetches use `/api/asesmen/*`. `/api/cbt/*` remains live only as a deprecated BFF compatibility/proxy namespace to the same Go API semantics for legacy clients, route access tests, mobile asset compatibility, and staged transition safety.
- Manual CBT smoke rehearsal is documented in `docs/cbt-smoke-checklist.md`.

## API Compatibility Status

As of 2026-05-06, new web-admin client code must use `/api/bank-soal/*` for Bank Soal and `/api/asesmen/*` for assessment. The SvelteKit BFF still keeps 66 legacy `/api/cbt/*` route handlers so old bookmarks, tests, and staged clients continue to work, but responses served through the `/api/cbt/*` namespace carry compatibility deprecation headers. The new alias namespaces must not carry those headers.

Audit summary for remaining `/api/cbt` references in `apps/web-admin/src`:

- Legacy BFF routes: `src/routes/api/cbt/**` continues to proxy to the unchanged Go API.
- New BFF aliases: `src/routes/api/bank-soal/**` and `src/routes/api/asesmen/**` re-export the same handler semantics while presenting the new client-facing namespaces.
- Route access and proxy tests: `src/lib/server/route-access*.ts`, `src/lib/server/proxy-routes.test.ts`, and retired `/cbt*` redirect tests keep coverage for old and new paths.
- Active UI/client compatibility guard: `src/lib/client/active-api-aliases.test.ts` prevents active Svelte pages/components from fetching `/api/cbt`.
- Mobile/client compatibility: the Go exam API may still return `/api/cbt/assets/*/file` media URLs for Flutter exam payloads; this web-admin BFF deprecation does not change that backend contract.

## Ownership Boundaries

### `apps/web-admin`

- SvelteKit admin web application
- session and cookie ownership
- admin and guru-facing workflows
- thin BFF or proxy behavior only

### `services/core-api`

- Go Chi API
- domain logic
- `sqlc`
- PostgreSQL
- migrations
- scheduler
- queue state

### `services/pusaka-worker`

- background execution
- claims jobs from backend
- reports result state back to backend

### `apps/mobile`

- Flutter Android client for student CBT
- token-based exam login
- local answer safety and restore metadata for BYOD conditions
- heartbeat, warning telemetry, and visible connection-health guidance

## PUSAKA Subsystem Boundary

- PUSAKA attendance automation is a bounded subsystem inside this monorepo.
- Its canonical backend namespace is `/api/pusaka/*`.
- Legacy runtime aliases for `/api/jobs`, `/api/attendance`, `/api/schedules`, and `/api/worker` have been retired.
- Only legacy UI entry paths such as `/attendance` and `/jobs` remain, and they redirect to canonical `/pusaka/*` pages.
- PUSAKA operational UI should live under `/pusaka/*`; legacy `/attendance` and `/jobs` surfaces should redirect there.
- `employees` remains the master data table for all school staff.
- Only eligible employment types (`pns`, `pppk`) should appear in PUSAKA account setup and operations.
- `/employees` is the master employee screen; `/pusaka/employees` is the PUSAKA operator screen.
- BFF proxies for PUSAKA employee operations live only under `/api/pusaka/employees/*`.
- PUSAKA credentials now live in `pusaka_accounts` as the integration-owned source of truth.
- Legacy `employees.pusaka_*` columns have been removed; employee identity and PUSAKA integration credentials are now explicitly separated.
- Internal code uses explicit `Pusaka*` ownership naming. A deeper package split is intentionally deferred until it would reduce, not increase, maintenance churn.

## Data Ownership

- PostgreSQL belongs only to `services/core-api`.
- Legacy SQLite files, if still present, are one-way import artifacts only.
- Frontend and worker must treat backend API as the source of truth.
- Flutter must treat the exam API as the source of truth and must not talk to PostgreSQL or SvelteKit for live exam runtime.

## Migration Strategy

- Prefer staged changes.
- Preserve deployability across frontend, backend, and worker VPS targets.
- Add new modules as bounded slices, not cross-cutting hacks.

## CBT & Academic Direction

The preferred build order for the academic and CBT ecosystem is:
1. **Academic Foundation**: Core master data (years, classes, subjects, teacher assignments).
2. **Identity & RBAC**: Unified user system with 5 roles (Admin, Teacher, Student, Staff, Parent) and entity decoupling.
3. **Student Lifecycle**: Management of active students, alumni, and prospective applicants.
4. **Question Bank**: Rich-text and asset-backed bank with versioning.
5. **Exam Packages & Sessions**: Adaptive cohort models and seat planning.
6. **Student Exam Runtime**: Mobile-first Flutter client (Sprint 7+).
7. **Proctoring & Analytics**: Real-time monitoring and score processing.
8. **Academic Reporting**: Report cards (Rapor) and performance tracking.

Do not jump straight to student exam runtime before the authoring, identity, and scheduling model is stable.

## Planned CBT Engine Boundary

`services/cbt-engine` is documented in `PLAN.md` as a proposed future extraction for live exam runtime only. It is not implemented yet. Until that sprint starts, do not change the current rule that `services/core-api` owns PostgreSQL. Any future exception must be narrow: Core API keeps the main school database, while CBT Engine may own only a runtime exam database/schema after explicit implementation.
