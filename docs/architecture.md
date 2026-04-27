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

## Data Ownership

- PostgreSQL belongs only to `services/core-api`.
- Legacy SQLite files, if still present, are one-way import artifacts only.
- Frontend and worker must treat backend API as the source of truth.

## Migration Strategy

- Prefer staged changes.
- Preserve deployability across frontend, backend, and worker VPS targets.
- Add new modules as bounded slices, not cross-cutting hacks.

## CBT Direction

The preferred CBT build order is:
1. master academic data
2. student and placement data
3. roles and permissions
4. question bank
5. exam packages
6. exam sessions and participants
7. monitoring and results

Do not jump straight to full student exam runtime before the authoring and scheduling model is stable.
