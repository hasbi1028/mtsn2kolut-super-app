# Project Policy

This repository is a monorepo for MTs Negeri 2 Kolaka Utara.

Current runtime units:
- `apps/web-admin` for the SvelteKit admin web app
- `services/core-api` for the Go Chi API, `sqlc`, migrations, scheduler, queue, and PostgreSQL ownership
- `services/pusaka-worker` for the Playwright worker

## Non-Negotiable Architecture Rules

- PostgreSQL is owned only by `services/core-api`.
- All runtime database access must go through Go backend code.
- SQL must remain explicit and typed through `sqlc`.
- `apps/web-admin` must not read or write PostgreSQL directly.
- `services/pusaka-worker` must not read or write PostgreSQL directly.
- Legacy SQLite files may exist only as import artifacts, never as active runtime storage.

## Backend Rules

- Keep Go code idiomatic and clean.
- Keep handlers thin: parse input, call services, map HTTP responses.
- Keep business rules in the service layer.
- Keep SQL in `services/core-api/db/queries`.
- Keep generated code in `services/core-api/internal/repository/postgres`.
- Prefer clear domain boundaries over clever abstractions.
- Do not replace `sqlc` with ORM-heavy patterns.
- Prefer incremental schema evolution through migrations.

## Frontend Rules

- Use SvelteKit as UI/BFF, not as the owner of backend business logic.
- Use `shadcn-svelte` (nova style) for all new UI primitives — it is now installed.
- Tailwind CSS v4 is now active via `@tailwindcss/vite` — use Tailwind utility classes.
- `$lib/utils.ts` provides `cn`, `WithElementRef`, `WithoutChild`, `WithoutChildren` — use these, do not add duplicates.
- New pages use the institutional light theme defined in `src/app.css` `@theme` block.
- Existing dark-theme pages (dashboard, employees, jobs, attendance, settings) use legacy CSS — do not rewrite them unless explicitly asked.
- Match the visual direction to an education context for MTs Negeri 2 Kolaka Utara.
- Avoid generic SaaS styling, generic dashboards, and purple-gradient AI aesthetics.
- Prefer layouts that feel institutional, calm, trustworthy, and clear for admin, guru, and sekolah workflows.
- Preserve mobile usability and readable spacing.
- Always run `npm run check` in `apps/web-admin` before finalizing Svelte changes.

## Worker Rules

- Worker is an API client of `services/core-api`.
- Worker runtime protocols should stay direct to backend, not tunneled through SvelteKit.
- Keep worker logic focused on job claiming, execution, retry, and reporting.

## Delivery Rules

- Prefer safe, staged migration over rewrites.
- Do not collapse service boundaries for convenience.
- Before changing architecture, preserve deployability to 3 VPS targets.
- For reviews or changes, prioritize correctness, operational safety, and maintainability.

## Deployment Model

- Monorepo source, but 3 separate deployments:
- frontend VPS runs `apps/web-admin`
- backend VPS runs `services/core-api`
- worker VPS runs `services/pusaka-worker`

Migration order:
1. backend code
2. PostgreSQL migration
3. backend restart and health check
4. frontend deploy
5. worker deploy

## Product Direction

- Build toward school operations and CBT readiness in staged slices.
- Prefer strong master data and academic foundations before advanced exam flows.
- Keep admin and guru workflows first-class.

## Current Build Status

**Completed slices (as of 2026-04-27):**
- Backend CBT foundation: migration 004 + 005, sqlc generated, all handlers compiling and tests passing
- Frontend UI stack: Tailwind v4 + shadcn-svelte nova
- Sidebar navigation: replaces topnav — grouped with icons, mobile-friendly, all modules listed
- Academic master data: `/academic` (tahun ajaran, kelas, mata pelajaran)
- Students: `/students`
- CBT question bank: `/cbt/questions`
- CBT package builder: `/cbt/packages`
- CBT exam sessions: `/cbt/sessions` — create, schedule, activate, finish, enroll students
- CBT results: `/cbt/sessions/[id]` — skor per peserta, stat kelulusan, ekspor CSV, trigger scoring
- Navigation: Sidebar.svelte — Nav.svelte is no longer used in layout

**Next planned slice:**
- Student-facing exam interface (token login + soal display + submit jawaban)
- Token generation untuk `cbt_exam_participants`

## Documentation Map

- Architecture details: `docs/architecture.md`
- UI and UX direction: `docs/ui-guidelines.md`
- Deployment contract: `docs/deployment.md`
- Service-specific rules:
- `apps/web-admin/AGENTS.md`
- `services/core-api/AGENTS.md`
- `services/pusaka-worker/AGENTS.md`
