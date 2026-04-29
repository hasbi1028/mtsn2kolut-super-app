# Architecture

## System Shape

This repository is a monorepo with three deployable runtime units:

- `apps/web-admin`
- `services/core-api`
- `services/pusaka-worker`

The monorepo is for source organization, not for collapsing runtime topology.

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

## PUSAKA Subsystem Boundary

- PUSAKA attendance automation is a bounded subsystem inside this monorepo.
- Its canonical backend namespace is `/api/pusaka/*`.
- Legacy runtime aliases for `/api/jobs`, `/api/attendance`, `/api/schedules`, and `/api/worker` have been retired.
- Only legacy UI entry paths such as `/attendance` and `/jobs` remain, and they redirect to canonical `/pusaka/*` pages.
- PUSAKA operational UI should live under `/pusaka/*`; legacy `/attendance` and `/jobs` surfaces should redirect there.
- PUSAKA credentials now live in `pusaka_accounts` as the integration-owned source of truth.
- Legacy `employees.pusaka_*` columns have been removed; employee identity and PUSAKA integration credentials are now explicitly separated.
- Internal code uses explicit `Pusaka*` ownership naming. A deeper package split is intentionally deferred until it would reduce, not increase, maintenance churn.

## Data Ownership

- PostgreSQL belongs only to `services/core-api`.
- Legacy SQLite files, if still present, are one-way import artifacts only.
- Frontend and worker must treat backend API as the source of truth.

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
