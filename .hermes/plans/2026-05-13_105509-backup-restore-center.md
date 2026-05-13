# Backup & Restore Center Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build a safe in-app Backup & Restore Center for MTsN 2 Kolut that lets admins monitor PostgreSQL backup health, list/download backups, run manual backups, and prepare restore safely without exposing secrets or enabling accidental production data loss.

**Architecture:** SvelteKit web-admin remains a BFF/UI layer and must not access PostgreSQL or backup folders directly. Go core-api exposes admin-only backup management endpoints that read sanitized backup metadata and invoke existing server-side scripts. Restore production is intentionally gated behind simulation/generate-command first, with destructive restore deferred until explicit multi-layer approval exists.

**Tech Stack:** SvelteKit web-admin, Go core-api, PostgreSQL `pg_dump`/`pg_restore`, systemd timer/service, sqlc only if persistent audit/job tables are added, PM2 deployment.

---

## Current Baseline

Existing daily backup is already active:

- systemd timer: `mtsn2kolut-postgresql-backup.timer`
- schedule: every day `00:00 WITA`
- script: `/home/servermtsn2kolut/mtsn2kolut-super-app/deploy/backup-postgresql.sh`
- backup dir: `/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql`
- retention: `30` days
- latest symlink: `latest.dump`

Important guardrails:

- Never display `DATABASE_URL`, DB username/password, token, or connection string.
- All logs returned to UI must be sanitized.
- Restore production must not be one-click in the first version.
- Backup/restore actions must be admin-only and audited.
- SvelteKit must call Go API/BFF only; no direct shell/db access from frontend.

---

# Options

## Option A — Recommended: Safe Backup Center in 4 Sprints

**Best for production.** Build observability first, then manual backup, then restore simulation/command, then offsite backup.

Pros:
- safest path
- immediately useful
- avoids accidental restore
- aligns with current architecture
- easy to validate sprint by sprint

Cons:
- restore automation comes later

Recommended sequence:
1. Sprint 1: Read-only Backup Center
2. Sprint 2: Manual Backup from UI
3. Sprint 3: Restore Safety / Simulation / Generate Command
4. Sprint 4: Offsite Backup Monitoring

## Option B — Backup Center + Manual Backup Only

Build dashboard/list/download/manual backup, but no restore UI.

Pros:
- very safe
- fast
- enough for routine operations

Cons:
- restore still manual/SOP only

Use if restore is rare and should always be done by operator on server.

## Option C — Full Backup & Restore Automation

Build read/list/backup/download/restore production fully from UI.

Pros:
- complete feature
- fastest restore during emergency

Cons:
- high risk
- needs strict locking, double confirmation, service stop/start, pre-restore backup, audit trail, rollback SOP
- dangerous if admin account compromised

Not recommended until Option A Sprint 1–3 are stable.

## Final Recommendation

Choose **Option A**.

Start with **Sprint Backup 1: Read-only Backup Center** because it is safe and proves the backup system from inside the app. Then add manual backup. Restore should start as **simulation + command generation**, not direct production restore.

---

# Permission Model

Add permissions:

- `backup.read` — view Backup Center status and list backup files
- `backup.download` — download backup files
- `backup.create` — run manual backup
- `backup.restore_plan` — validate backup and generate restore command
- `backup.restore_execute` — reserved for future destructive restore execution, not enabled initially

Admin role may be allowed by legacy fallback, but dynamic RBAC permissions should be supported.

---

# API Contract Draft

Base path:

```text
/api/system/backups
```

## GET `/api/system/backups/status`

Returns sanitized status:

```json
{
  "timer_name": "mtsn2kolut-postgresql-backup.timer",
  "service_name": "mtsn2kolut-postgresql-backup.service",
  "timer_enabled": true,
  "timer_active": true,
  "schedule": "*-*-* 00:00:00",
  "timezone": "WITA",
  "last_run_at": "2026-05-13T00:00:01+08:00",
  "last_run_success": true,
  "next_run_at": "2026-05-14T00:00:00+08:00",
  "latest_backup": {
    "name": "pusaka_20260513_000001.dump",
    "size_bytes": 1420865,
    "created_at": "2026-05-13T00:00:01+08:00",
    "sha256_available": true
  },
  "retention_days": 30,
  "backup_count": 28,
  "backup_dir_size_bytes": 30408704,
  "health": "ok",
  "warnings": []
}
```

## GET `/api/system/backups`

Returns list of sanitized backup metadata:

```json
{
  "items": [
    {
      "id": "pusaka_20260513_000001.dump",
      "name": "pusaka_20260513_000001.dump",
      "kind": "scheduled",
      "size_bytes": 1420865,
      "created_at": "2026-05-13T00:00:01+08:00",
      "sha256": "optional-if-file-exists",
      "is_latest": true,
      "downloadable": true
    }
  ],
  "meta": {
    "total": 28
  }
}
```

## GET `/api/system/backups/{id}/download`

Downloads a `.dump` file.

Security:
- path traversal blocked
- only files under configured backup directory
- only `.dump` files
- requires `backup.download`
- audit log entry created

## POST `/api/system/backups/run`

Runs manual backup.

Request:

```json
{
  "reason": "before migration akademik sprint"
}
```

Response:

```json
{
  "job_id": "backup-20260513-110000",
  "status": "running"
}
```

Sprint 2 may implement simple synchronous execution first if backup stays fast, but a lock is required to prevent double-run.

## GET `/api/system/backups/jobs/{job_id}`

Returns job status and sanitized output.

## POST `/api/system/backups/{id}/validate-restore`

Validates backup file without touching production.

Possible checks:
- file exists
- file extension `.dump`
- `pg_restore --list` succeeds
- optional checksum matches

Response:

```json
{
  "valid": true,
  "format": "pg_dump custom",
  "object_count": 210,
  "warnings": []
}
```

## POST `/api/system/backups/{id}/restore-command`

Generates safe manual restore command/SOP, not executing it.

Response must redact secrets and include placeholders:

```json
{
  "requires_manual_server_access": true,
  "steps": [
    "Create pre-restore backup",
    "Stop app services",
    "Restore selected dump using pg_restore",
    "Run health check",
    "Restart services"
  ],
  "command_preview": "# Generated restore SOP with [REDACTED] credentials"
}
```

---

# UI Plan

Route:

```text
/settings/backups
```

Menu label:

```text
Backup & Restore
```

Sections:

1. Status cards
   - Backup health
   - Latest backup
   - Next schedule
   - Retention
   - Total backup size

2. Backup actions
   - Refresh status
   - Create backup now
   - Download latest

3. Backup table
   - name
   - created at
   - size
   - kind
   - checksum status
   - actions: download, validate, generate restore SOP

4. Restore safety panel
   - warning that production restore is destructive
   - validate selected backup
   - generate command/SOP
   - no direct restore button in Sprint 1–3

5. Offsite warning panel
   - local-only backup warning
   - recommendation to configure external sync

---

# Sprint Backup 1 — Read-only Backup Center

**Objective:** Give admins a safe dashboard to verify backup health and download backups.

**Scope:** No write actions except audit/download logs if existing audit mechanism supports it.

## Task 1: Add backup configuration constants in core-api

**Files:**
- Create/modify: `services/core-api/internal/service/system_backup.go`

Add constants/env-driven config:

- backup dir: `/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql`
- timer name: `mtsn2kolut-postgresql-backup.timer`
- service name: `mtsn2kolut-postgresql-backup.service`
- retention days: default `30`

Verification:

```bash
cd services/core-api
go test ./internal/service
```

## Task 2: Implement safe backup file scanner

**Objective:** List `.dump` files without exposing arbitrary paths.

Rules:
- only read configured directory
- ignore non `.dump`
- compute size and modified time
- mark `latest.dump` target
- classify kind:
  - `scheduled` if filename starts `pusaka_`
  - `manual-pre-change` if filename starts `pre-`
  - `manual` otherwise

Tests:
- ignores path traversal
- sorts newest first
- handles missing directory gracefully

## Task 3: Implement systemd status parser

**Objective:** Read timer status safely.

Implementation approach:
- execute `systemctl show` with explicit unit names only
- use fields like:
  - `UnitFileState`
  - `ActiveState`
  - `NextElapseUSecRealtime`
  - `LastTriggerUSec`
- no shell interpolation with user input

Commands should use `exec.CommandContext`, not `sh -c` with user values.

## Task 4: Add Go handler endpoints

**Files:**
- Modify: `services/core-api/internal/handler/system_backup.go`
- Modify: `services/core-api/cmd/api/main.go`

Endpoints:

```text
GET /api/system/backups/status
GET /api/system/backups
GET /api/system/backups/{id}/download
```

Permission requirement:
- `backup.read` for status/list
- `backup.download` for download

## Task 5: Add web-admin BFF routes

**Files:**
- Create: `apps/web-admin/src/routes/api/system/backups/status/+server.ts`
- Create: `apps/web-admin/src/routes/api/system/backups/+server.ts`
- Create: `apps/web-admin/src/routes/api/system/backups/[id]/download/+server.ts`

These should proxy to core-api and preserve auth.

## Task 6: Add route access permissions

**Files:**
- Modify: `apps/web-admin/src/lib/server/route-access.ts`
- Modify: `apps/web-admin/src/lib/server/route-access.test.ts`

Rules:
- `/settings/backups` requires `backup.read`
- `GET /api/system/backups/status` requires `backup.read`
- `GET /api/system/backups` requires `backup.read`
- `GET /api/system/backups/{id}/download` requires `backup.download`
- admin role fallback remains allowed

## Task 7: Build Backup Center UI

**Files:**
- Create: `apps/web-admin/src/routes/settings/backups/+page.svelte`
- Modify sidebar/nav file used by settings menu

UI states:
- loading
- healthy
- warning if latest backup older than 24 hours
- error if timer inactive
- empty backup list

## Task 8: Validate Sprint 1

Commands:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-backup-center-sprint1 ./cmd/api
```

Expected:
- all pass

Commit:

```bash
git add services/core-api apps/web-admin
 git commit -m "feat(system): add backup center dashboard"
```

---

# Sprint Backup 2 — Manual Backup from UI

**Objective:** Allow admin to create a manual backup safely.

## Task 1: Add backup run endpoint

Endpoint:

```text
POST /api/system/backups/run
```

Guardrails:
- requires `backup.create`
- server-side lock file or process lock
- reject if backup already running
- sanitize logs
- reason required or optional max 200 chars

## Task 2: Execute existing script safely

Script:

```text
deploy/backup-postgresql.sh
```

Implementation:
- `exec.CommandContext` with fixed script path
- timeout, e.g. 10 minutes
- no user input in command args except sanitized reason in audit metadata

## Task 3: Add job status

Minimum viable:
- synchronous response if backup completes quickly

Better:
- in-memory job state or DB-backed `system_jobs`
- statuses: `running`, `success`, `failed`

## Task 4: Add UI button

UI:
- button `Buat Backup Sekarang`
- confirmation dialog
- show progress
- refresh list after success

## Task 5: Validate and commit

Commands same as Sprint 1.

Commit:

```bash
git commit -m "feat(system): allow manual database backups"
```

---

# Sprint Backup 3 — Restore Safety, Validation, and SOP

**Objective:** Make restore safer by validating dumps and generating an operator SOP without direct production restore.

## Task 1: Add validate restore endpoint

Endpoint:

```text
POST /api/system/backups/{id}/validate-restore
```

Implementation:

```bash
pg_restore --list selected.dump
```

Rules:
- fixed file path under backup dir
- timeout
- output summarized, not dumped raw if too large

## Task 2: Add restore SOP generator

Endpoint:

```text
POST /api/system/backups/{id}/restore-command
```

Output:
- manual steps
- generated command with placeholders and `[REDACTED]`
- reminder to create pre-restore backup first
- reminder to stop PM2 services before destructive restore

## Task 3: Add UI restore safety panel

UI:
- select backup
- validate restore
- generate SOP
- copy SOP
- strong warning: no direct production restore yet

## Task 4: Validate and commit

Commit:

```bash
git commit -m "feat(system): add backup restore validation workflow"
```

---

# Sprint Backup 4 — Offsite Backup Monitoring

**Objective:** Reduce risk of losing backups if the server disk fails.

Options:

1. `rclone` to Google Drive / S3-compatible storage
2. `rsync` to another server/NAS
3. External disk mount with periodic sync

Recommended MVP:
- add offsite sync script outside app first
- UI only reads status/logs
- do not store cloud secrets in app DB

UI:
- offsite configured: yes/no
- last sync time
- last sync status
- number of remote backups
- warning if offsite sync older than 24 hours

Commit:

```bash
git commit -m "feat(system): monitor offsite backup sync"
```

---

# Future Sprint — Direct Production Restore

Only after Sprints 1–3 are stable.

Required safeguards:

1. Permission `backup.restore_execute`
2. Re-authentication/password confirmation
3. Type exact phrase:

```text
SAYA PAHAM RESTORE AKAN MENIMPA DATABASE
```

4. Auto pre-restore backup
5. Stop `mtsn2kolut-core-api` and `mtsn2kolut-web-admin`
6. Run restore
7. Run migrations/checks
8. Restart services
9. Health check
10. Audit log
11. Emergency rollback SOP

This should remain disabled by default.

---

# Acceptance Criteria

## Sprint 1

- Admin can open `/settings/backups`
- Dashboard shows active timer, last backup, next backup, count, retention
- List shows `.dump` backups newest first
- Download works for authorized admin
- Path traversal is blocked
- Non-admin/non-permission user is forbidden
- No credentials appear in UI/API/logs

## Sprint 2

- Admin can run manual backup
- Double-click/double-run is blocked
- Backup appears in list after success
- Failure is shown safely without secrets

## Sprint 3

- Admin can validate dump file with `pg_restore --list`
- Admin can generate restore SOP
- UI does not execute production restore
- SOP contains redacted placeholders only

## Sprint 4

- Offsite backup status visible
- UI warns if offsite sync stale/failing

---

# Verification Checklist Before Deploy

Run:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-backup-center ./cmd/api
```

Manual smoke:

```bash
curl -fsS http://127.0.0.1:8080/health
curl -I http://127.0.0.1:8021/settings/backups
```

Deploy order when implementation is approved:

1. Build core-api if backend changed
2. Restart `mtsn2kolut-core-api`
3. Build web-admin
4. Restart `mtsn2kolut-web-admin`
5. Health check
6. UI smoke test

---

# Recommendation Summary

Implement Option A.

Start immediately with:

```text
Sprint Backup 1 — Read-only Backup Center
```

Do not implement direct production restore yet. Add restore validation and SOP generation first.
