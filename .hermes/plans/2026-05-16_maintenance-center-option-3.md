# Maintenance Center Option 3 Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build a full Maintenance Center for MTsN 2 Kolut that supports global/module/read-only maintenance, admin bypass, scheduling, health + backup visibility, audit logs, and safe operational workflows.

**Architecture:** Implement maintenance state in Go core-api with PostgreSQL-backed configuration and audit logs. SvelteKit web-admin remains BFF/UI only and calls core-api; no direct DB access from SvelteKit. Enforcement is layered: backend middleware/guards for API mutations and SvelteKit route guard for user-facing maintenance page.

**Tech Stack:** Go core-api, PostgreSQL migrations, sqlc, SvelteKit web-admin, PM2 deployment, existing Backup Center/status services.

---

## Current Context / Assumptions

- Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
- No existing maintenance feature was found in Go/Svelte code search.
- Existing PM2 process `mtsn2kolut-maintenance` exists at ops level, but there is no app-level Maintenance Center UI/API yet.
- Existing Backup Center is already known from app memory and should be integrated as status/checklist, not duplicated.
- Production deploy/restart must not happen until user explicitly approves implementation/deploy.
- All sensitive values must remain `[REDACTED]`; never expose env, tokens, database URLs, or passwords.

## Recommended Scope: Option 3, Implemented in 4 Sprints

Option 3 is a complete Maintenance Center, but implementation should be staged to reduce risk:

1. **Sprint M0 — Discovery and safe contract**
2. **Sprint M1 — Global/module/read-only maintenance core**
3. **Sprint M2 — Scheduler and module-aware enforcement**
4. **Sprint M3 — Health, backup checklist, audit UX, notifications foundation**
5. **Sprint M4 — Polish, operational docs, and production rollout**

---

# Sprint M0 — Discovery and Data Contract

## Task M0.1: Inspect existing routing, auth, roles, system endpoints

**Objective:** Identify exact auth middleware, role model, settings routes, backup center APIs, and frontend layout hooks before adding maintenance.

**Files to inspect:**
- `services/core-api/cmd/api/main.go`
- `services/core-api/internal/handler/*`
- `services/core-api/internal/middleware/*`
- `services/core-api/internal/service/system_backup.go`
- `services/core-api/db/queries/*.sql`
- `apps/web-admin/src/hooks.server.ts`
- `apps/web-admin/src/routes/settings/**`
- `apps/web-admin/src/routes/api/**`
- sidebar/nav component under `apps/web-admin/src/lib/**`

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
```

**Expected:** Existing baseline passes before feature work.

## Task M0.2: Write maintenance contract document

**Objective:** Create a durable implementation contract for API, UI states, and enforcement rules.

**Create:**
- `docs/contracts/maintenance-center.md`

**Content must include:**
- Maintenance modes:
  - `off`
  - `global`
  - `module`
  - `read_only`
- Affected modules:
  - `global`
  - `auth`
  - `dashboard`
  - `akademik`
  - `students`
  - `bank_soal`
  - `cbt`
  - `pusaka`
  - `backup_restore`
  - `settings`
- Bypass roles:
  - `superadmin`
  - `admin`
  - explicit user IDs if supported later
- Blocked actions:
  - mutating API requests during global/module maintenance
  - selected module routes during module mode
  - write actions during read-only mode
- Allowed actions:
  - login page
  - maintenance status public endpoint
  - superadmin/admin settings maintenance page
  - health endpoint

**Commit:**

```bash
git add docs/contracts/maintenance-center.md
git commit -m "docs(maintenance): add maintenance center contract"
```

---

# Sprint M1 — Core Maintenance Mode and UI

## Task M1.1: Add database migration

**Objective:** Store maintenance windows and audit logs safely.

**Create:**
- `services/core-api/db/migrations/089_system_maintenance_center.sql`

**Schema:**

```sql
CREATE TABLE IF NOT EXISTS system_maintenance_windows (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title text NOT NULL,
  message text NOT NULL,
  mode text NOT NULL CHECK (mode IN ('global', 'module', 'read_only')),
  affected_modules text[] NOT NULL DEFAULT ARRAY['global']::text[],
  starts_at timestamptz,
  ends_at timestamptz,
  is_active boolean NOT NULL DEFAULT false,
  allow_admin_bypass boolean NOT NULL DEFAULT true,
  bypass_roles text[] NOT NULL DEFAULT ARRAY['superadmin','admin']::text[],
  severity text NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'critical')),
  created_by uuid,
  updated_by uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_system_maintenance_windows_active
  ON system_maintenance_windows (is_active, starts_at, ends_at);

CREATE TABLE IF NOT EXISTS system_maintenance_audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  maintenance_id uuid REFERENCES system_maintenance_windows(id) ON DELETE SET NULL,
  actor_user_id uuid,
  action text NOT NULL,
  reason text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_system_maintenance_audit_logs_created_at
  ON system_maintenance_audit_logs (created_at DESC);
```

**Validation:**

```bash
cd services/core-api
# apply in transaction on test/staging DB if available
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
```

## Task M1.2: Add sqlc queries

**Objective:** Provide typed DB access for maintenance state and audit log.

**Create:**
- `services/core-api/db/queries/system_maintenance.sql`

**Queries:**

```sql
-- name: GetActiveMaintenanceWindow :one
SELECT *
FROM system_maintenance_windows
WHERE is_active = true
  AND (starts_at IS NULL OR starts_at <= now())
  AND (ends_at IS NULL OR ends_at >= now())
ORDER BY
  CASE mode WHEN 'global' THEN 1 WHEN 'read_only' THEN 2 ELSE 3 END,
  created_at DESC
LIMIT 1;

-- name: ListMaintenanceWindows :many
SELECT *
FROM system_maintenance_windows
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateMaintenanceWindow :one
INSERT INTO system_maintenance_windows (
  title, message, mode, affected_modules, starts_at, ends_at,
  is_active, allow_admin_bypass, bypass_roles, severity, created_by, updated_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11
)
RETURNING *;

-- name: UpdateMaintenanceWindow :one
UPDATE system_maintenance_windows
SET title = $2,
    message = $3,
    mode = $4,
    affected_modules = $5,
    starts_at = $6,
    ends_at = $7,
    allow_admin_bypass = $8,
    bypass_roles = $9,
    severity = $10,
    updated_by = $11,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetMaintenanceWindowActive :one
UPDATE system_maintenance_windows
SET is_active = $2,
    updated_by = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateMaintenanceAuditLog :one
INSERT INTO system_maintenance_audit_logs (maintenance_id, actor_user_id, action, reason, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListMaintenanceAuditLogs :many
SELECT *
FROM system_maintenance_audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
```

**Validation:**

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
```

## Task M1.3: Add Go service

**Objective:** Centralize maintenance validation, status, activation, deactivation, and audit logging.

**Create:**
- `services/core-api/internal/service/system_maintenance.go`
- `services/core-api/internal/service/system_maintenance_test.go`

**Service responsibilities:**
- Normalize mode/module names.
- Validate title/message required.
- Validate `ends_at > starts_at` when both present.
- Prevent activating a maintenance window that already ended.
- Write audit log for create/update/activate/deactivate.
- Return public-safe status object without secrets.

**Test cases:**
- create rejects empty title/message.
- create rejects invalid mode.
- activate rejects expired window.
- status returns active window.
- deactivate writes audit event.

**Validation:**

```bash
cd services/core-api
go test ./internal/service -run Maintenance -v
```

## Task M1.4: Add Go handler and routes

**Objective:** Expose authenticated admin APIs and public status endpoint.

**Create/Modify:**
- `services/core-api/internal/handler/system_maintenance.go`
- `services/core-api/internal/handler/system_maintenance_test.go`
- `services/core-api/cmd/api/main.go`

**Endpoints:**

```text
GET    /api/system/maintenance/status          public-safe
GET    /api/system/maintenance/windows         admin only
POST   /api/system/maintenance/windows         admin only
PUT    /api/system/maintenance/windows/{id}    admin only
POST   /api/system/maintenance/windows/{id}/activate    admin only
POST   /api/system/maintenance/windows/{id}/deactivate  admin only
GET    /api/system/maintenance/audit-logs      admin only
```

**Test cases:**
- public status works without auth if existing framework allows public API.
- create requires admin auth.
- activate returns audit log side effect.
- invalid mode returns 400.

**Validation:**

```bash
cd services/core-api
go test ./internal/handler -run Maintenance -v
```

## Task M1.5: Add backend maintenance guard middleware

**Objective:** Enforce maintenance on API mutations safely.

**Modify/Create:**
- `services/core-api/internal/middleware/maintenance.go`
- `services/core-api/cmd/api/main.go`

**Rules:**
- Always allow:
  - `/health`
  - `/api/system/maintenance/status`
  - admin maintenance endpoints
  - auth/login endpoints required to let admins enter
- If `global` active:
  - block non-bypass users for most API requests.
- If `module` active:
  - block API routes mapped to affected modules.
- If `read_only` active:
  - allow safe methods `GET`, `HEAD`, `OPTIONS`; block `POST`, `PUT`, `PATCH`, `DELETE` unless bypass.

**Important:** Cache active status for a short TTL, e.g. 5 seconds, to avoid a DB query on every request.

**Test cases:**
- global blocks normal user mutation.
- read-only allows GET but blocks POST.
- admin bypass works when enabled.
- public health/status still works.

## Task M1.6: Add SvelteKit BFF routes

**Objective:** Keep SvelteKit as BFF/proxy only.

**Create:**
- `apps/web-admin/src/routes/api/system/maintenance/status/+server.ts`
- `apps/web-admin/src/routes/api/system/maintenance/windows/+server.ts`
- `apps/web-admin/src/routes/api/system/maintenance/windows/[id]/+server.ts`
- `apps/web-admin/src/routes/api/system/maintenance/windows/[id]/activate/+server.ts`
- `apps/web-admin/src/routes/api/system/maintenance/windows/[id]/deactivate/+server.ts`
- `apps/web-admin/src/routes/api/system/maintenance/audit-logs/+server.ts`

**Pattern:** Follow existing BFF proxy helpers and auth cookie forwarding.

**Validation:**

```bash
npm --prefix apps/web-admin run check
```

## Task M1.7: Add Maintenance Center UI

**Objective:** Create admin-facing page under Settings.

**Create:**
- `apps/web-admin/src/routes/settings/maintenance/+page.svelte`

**UI sections:**
- Status summary card.
- Create maintenance form.
- Active maintenance card with deactivate button.
- Windows table.
- Audit log preview.

**Form fields:**
- title
- message
- mode
- affected modules
- starts_at
- ends_at
- severity
- allow admin bypass
- bypass roles

**UX safety:**
- confirm dialog for activate/deactivate.
- warning if global maintenance will affect students/guru.
- clear label that admin bypass remains enabled by default.

## Task M1.8: Add public maintenance page and layout guard

**Objective:** Show a clean maintenance page to affected users.

**Create:**
- `apps/web-admin/src/routes/maintenance/+page.svelte`

**Modify:**
- `apps/web-admin/src/hooks.server.ts` or existing layout server guard.

**Rules:**
- If active global maintenance and user is not bypass role:
  - redirect normal app route to `/maintenance`.
- `/maintenance`, `/login`, assets, API status are allowed.
- If maintenance off:
  - `/maintenance` can show “Sistem aktif normal” or redirect to `/login`.

**Validation:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

## Sprint M1 full verification

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/service ./internal/handler ./internal/repository/postgres
go build -o /tmp/core-api-maintenance-m1 ./cmd/api
```

**Commit:**

```bash
git add services/core-api apps/web-admin docs/contracts/maintenance-center.md
git commit -m "feat(system): add maintenance center core"
```

---

# Sprint M2 — Scheduler and Module-level Enforcement

## Task M2.1: Add scheduler evaluation

**Objective:** Support scheduled maintenance without manual activation at exact time.

**Approach:** Avoid a separate cron initially. Treat a window as active if:

```text
is_active = true AND starts_at <= now() AND ends_at >= now()
```

Add UI labels:
- scheduled
- active now
- ended
- inactive

**Modify:**
- Go service status mapping.
- Svelte status display.

## Task M2.2: Add module route map

**Objective:** Define which routes are affected by each module.

**Create:**
- `services/core-api/internal/service/maintenance_modules.go`
- `apps/web-admin/src/lib/maintenance/modules.ts`

**Module route examples:**

```text
bank_soal: /api/cbt/questions, /api/cbt/packages, /bank-soal
cbt: /api/cbt/events, /api/cbt/sessions, /cbt, /asesmen
akademik: /api/academic, /akademik
students: /api/students, /students
pusaka: /api/pusaka, /pusaka
backup_restore: /api/system/backups, /settings/backups
settings: /api/settings, /settings
```

## Task M2.3: Module-level UI warning banners

**Objective:** Show contextual banners only on affected modules.

**Modify/Create:**
- app layout banner component under `apps/web-admin/src/lib/components/maintenance/MaintenanceBanner.svelte`
- route/layout integration.

**Behavior:**
- global: full page redirect for non-bypass.
- module: show banner and block module routes/actions.
- read-only: show banner “Mode baca saja aktif”.

## Task M2.4: Tests

**Backend tests:**
- module `bank_soal` blocks bank soal mutation only.
- module `akademik` does not block bank soal route.
- read-only blocks all mutations except bypass.

**Frontend checks:**

```bash
npm --prefix apps/web-admin run check
```

**Commit:**

```bash
git commit -m "feat(system): add scheduled module maintenance guard"
```

---

# Sprint M3 — Health, Backup Checklist, Audit UX, Notifications Foundation

## Task M3.1: Add health summary endpoint

**Objective:** Make Maintenance Center useful before activating maintenance.

**Endpoint:**

```text
GET /api/system/maintenance/health-summary
```

**Data:**
- core-api status from internal health.
- database connected.
- server time/db time.
- backup last status from existing Backup Center service.
- disk usage if existing system status service supports it.
- active CBT/session count if cheap and already available.

**Do not expose:**
- env values.
- connection strings.
- filesystem secrets.

## Task M3.2: Add pre-maintenance checklist UI

**Modify:**
- `apps/web-admin/src/routes/settings/maintenance/+page.svelte`

**Checklist items:**
- Backend health OK.
- DB connected.
- Last backup exists.
- Disk space OK.
- No active CBT session warning.
- Admin bypass enabled.

## Task M3.3: Audit log full page/table

**Objective:** Make audit usable for accountability.

**UI:**
- filter by action.
- filter by date.
- show actor display name if backend supports it.
- show metadata pretty preview without secrets.

## Task M3.4: Notification foundation

**Objective:** Prepare notification hooks without requiring full WhatsApp/Telegram config in first pass.

**Backend service:**
- on activate/deactivate, emit internal notification event/log.
- optional env-configured webhook later.

**Do not hardcode credentials.**

**Commit:**

```bash
git commit -m "feat(system): add maintenance health checklist and audit UX"
```

---

# Sprint M4 — Production Rollout and Operational SOP

## Task M4.1: Write SOP docs

**Create:**
- `docs/ops/maintenance-center-sop.md`

**Include:**
- when to use global vs module vs read-only.
- deployment maintenance procedure.
- CBT emergency procedure.
- backup before maintenance procedure.
- rollback rules.

## Task M4.2: Production migration and deployment plan

**Required order:**

1. Backup DB.
2. Apply migration `089_system_maintenance_center.sql`.
3. Build core-api.
4. Restart core-api.
5. Health check.
6. Build web-admin.
7. Restart web-admin immediately after build.
8. Smoke test.
9. Create a test maintenance window as admin.
10. Activate/deactivate test window.
11. Verify audit log.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/service ./internal/handler ./internal/repository/postgres
go build -o /tmp/core-api-maintenance-center ./cmd/api
```

**Deploy only after explicit approval.**

## Task M4.3: Browser smoke test

Manual/browser test:

1. Login as superadmin.
2. Open `/settings/maintenance`.
3. Create module maintenance for `bank_soal`.
4. Activate.
5. Open Bank Soal as normal teacher/user.
6. Confirm blocked/banner works.
7. Confirm admin bypass works.
8. Deactivate.
9. Confirm normal access restored.
10. Check audit log.

**Commit:**

```bash
git commit -m "docs(system): add maintenance center operational SOP"
```

---

## Acceptance Criteria

### Functional

- Admin can create, update, activate, deactivate maintenance windows.
- Public status endpoint returns active maintenance safely.
- Global maintenance blocks non-bypass users.
- Module maintenance affects only selected modules.
- Read-only mode blocks mutations but allows reads.
- Admin bypass works when enabled.
- Audit logs are created for every important action.
- Maintenance page is shown to affected users.
- Settings page shows health/checklist/audit.

### Safety

- No secrets exposed in API/UI/logs.
- Admin cannot accidentally lock all admins out if bypass enabled.
- `/health`, login, and maintenance status remain accessible.
- Backup status is read-only unless explicitly implementing backup action.
- No deploy/restart until explicit approval.

### Verification

- `npm --prefix apps/web-admin run check` PASS.
- `npm --prefix apps/web-admin run build` PASS.
- `sqlc generate` PASS.
- `go test ./internal/service ./internal/handler ./internal/repository/postgres` PASS.
- `go build` PASS.
- Browser smoke test PASS.

---

## Risks and Mitigations

### Risk: Locking out all users/admins

**Mitigation:** Admin bypass enabled by default, login/status routes always allowed, emergency DB rollback SOP documented.

### Risk: Maintenance middleware blocks health or auth

**Mitigation:** Explicit allowlist for `/health`, auth/login, static assets, and maintenance status.

### Risk: Too many DB reads per request

**Mitigation:** In-memory active-status cache with short TTL, e.g. 5 seconds.

### Risk: Module mapping misses some routes

**Mitigation:** Start with broad prefixes per module and add tests for critical modules Bank Soal, CBT, Akademik, Students.

### Risk: Web-admin stale chunks after build

**Mitigation:** If `npm run build` is run for production, immediately restart `mtsn2kolut-web-admin`.

---

## Recommended First Implementation Cut

If starting implementation immediately, do **Sprint M1 only first**, because it delivers usable value with controlled risk:

---

## Completion Notes — 2026-05-16

Implemented M0–M4 in one pass after user requested full implementation:

- Added contract: `docs/contracts/maintenance-center.md`.
- Added migration: `services/core-api/db/migrations/111_system_maintenance_center.sql`.
- Added sqlc queries/generated repository for maintenance windows and audit logs.
- Added Go service/handler for status, windows, activation/deactivation, audit logs, and health summary.
- Added backend maintenance middleware with short cached status lookup, global/module/read-only enforcement, and admin bypass.
- Added module mapping for `global`, `auth`, `dashboard`, `akademik`, `students`, `bank_soal`, `cbt`, `pusaka`, `backup_restore`, and `settings`.
- Added SvelteKit BFF routes under `/api/system/maintenance/*`.
- Added UI routes `/settings/maintenance` and `/maintenance`.
- Added root-layout maintenance status load, affected-route redirect to `/maintenance`, and admin bypass banner.
- Added sidebar link under Settings.

Verification completed:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/service ./internal/handler ./internal/repository/postgres
go build -o /tmp/core-api-maintenance-center ./cmd/api
```

All commands passed. Migration was also syntax-validated against PostgreSQL inside `BEGIN; ... ROLLBACK;` with no persisted schema change.

Deployment note: `npm run build` was executed for verification. Per MTsN 2 Kolut stale-chunk runbook, production deploy should apply the migration, build/restart core-api, then immediately restart/rebuild web-admin if promoting this build. No production migration/restart was performed as part of this implementation unless explicitly requested separately.

- DB tables.
- Status/create/update/activate/deactivate APIs.
- Backend read-only/global guard.
- Basic `/settings/maintenance` UI.
- Public `/maintenance` page.
- Audit log.

Then proceed to M2/M3 after M1 is stable.
