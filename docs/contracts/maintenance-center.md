# Maintenance Center Contract

Status: implemented draft for Option 3 (Sprint M0–M4)

## Purpose

Maintenance Center is the operational control surface for planned and emergency system maintenance in MTsN 2 Kolut. It supports:

- Global maintenance mode.
- Module-level maintenance mode.
- Read-only mode for state-changing API protection.
- Admin bypass roles.
- Scheduled windows.
- Public maintenance page.
- Health/checklist summary.
- Audit log for every create/update/activate/deactivate action.

## Backend API

### Public

`GET /api/system/maintenance/status`

Response data:

```json
{
  "active": true,
  "mode": "global",
  "server_time": "2026-05-16T14:00:00Z",
  "window": {
    "id": "uuid",
    "title": "Maintenance Sistem",
    "message": "Sistem sedang maintenance.",
    "mode": "global",
    "affected_modules": ["global"],
    "starts_at": "2026-05-16T14:00:00Z",
    "ends_at": "2026-05-16T15:00:00Z",
    "is_active": true,
    "allow_admin_bypass": true,
    "bypass_roles": ["admin"],
    "severity": "warning",
    "status": "active_now",
    "created_at": "2026-05-16T14:00:00Z",
    "updated_at": "2026-05-16T14:00:00Z"
  }
}
```

### Admin only

- `GET /api/system/maintenance/health-summary`
- `GET /api/system/maintenance/windows?limit=50&offset=0`
- `POST /api/system/maintenance/windows`
- `PUT /api/system/maintenance/windows/{id}`
- `POST /api/system/maintenance/windows/{id}/activate`
- `POST /api/system/maintenance/windows/{id}/deactivate`
- `GET /api/system/maintenance/audit-logs?limit=50&offset=0&action=&from=&to=`

Create/update payload:

```json
{
  "title": "Maintenance Sistem",
  "message": "Sistem sedang maintenance.",
  "mode": "global",
  "affected_modules": ["global"],
  "starts_at": "2026-05-16T14:00:00Z",
  "ends_at": "2026-05-16T15:00:00Z",
  "is_active": false,
  "allow_admin_bypass": true,
  "bypass_roles": ["admin"],
  "severity": "warning",
  "reason": "Deploy terjadwal"
}
```

Activate/deactivate payload:

```json
{
  "reason": "Maintenance selesai"
}
```

## Modes

- `global`: blocks non-bypass authenticated API access and redirects affected pages to `/maintenance`.
- `module`: blocks state-changing API requests for affected module paths.
- `read_only`: blocks all state-changing authenticated API requests except maintenance/auth safe paths.
- `off`: no active maintenance.

## Modules

Supported module codes:

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

## Frontend Routes

- `/settings/maintenance`: admin Maintenance Center.
- `/maintenance`: public maintenance information page.
- Root layout loads maintenance status and redirects non-bypass users for affected page routes.
- Root layout renders `MaintenanceBanner` for bypass users/admins so they know maintenance is active.

## Security Rules

- Maintenance management endpoints require admin access.
- Public status endpoint exposes only operational state, no secrets.
- Health summary must not expose credentials/env values.
- Backend guard fails open on status lookup error to avoid accidental lockout.
- `/api/system/maintenance/*`, `/health`, and auth login/refresh/logout stay reachable during maintenance.
- Admin bypass is role-based and auditable.

## Deployment Order

1. Backup production database.
2. Apply migration `111_system_maintenance_center.sql`.
3. Run `sqlc generate`.
4. Build and restart `core-api`.
5. Build and restart `web-admin` immediately after SvelteKit build.
6. Smoke test:
   - `/health`
   - `/api/system/maintenance/status`
   - `/maintenance`
   - `/settings/maintenance` protected route redirect/auth boundary.

## Rollback Notes

- If migration has been applied and app rollback is needed, keep migration in place because new tables are additive.
- To disable maintenance directly in SQL during emergency, set `is_active=false` on `system_maintenance_windows` after taking an audit note.
