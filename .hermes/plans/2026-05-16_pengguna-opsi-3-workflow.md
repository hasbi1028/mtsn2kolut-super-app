# Manajemen Pengguna Opsi 3 Workflow Implementation Plan

> **For Hermes:** Implement directly or delegate to Codex. Do not deploy/restart production without explicit approval.

**Goal:** Convert Manajemen Pengguna into a workflow-based account center for Pegawai, Siswa, Orang Tua, and Admin/RBAC with safer UX, previews, drawer-style focus, and bulk/generate workflows using existing API boundaries.

**Architecture:** Keep SvelteKit as BFF/UI only. Use existing Go core-api endpoints for users, RBAC, profile candidates, employee generation, student generation, parent generation, reset password, role assignment, profile link, and status changes. Avoid new DB migrations unless strictly necessary.

**Tech Stack:** SvelteKit/Svelte 5 web-admin, Go core-api, PostgreSQL/sqlc existing services.

---

## Scope

Implement the full Opsi 3 UX as UI workflow completion using existing backend capabilities:

1. Replace `/settings/users` layout with tabbed workflow center:
   - Ringkasan
   - Pegawai
   - Siswa
   - Orang Tua
   - Admin & Hak Akses
   - Generate Akun
   - Audit/Permintaan
2. Add account health/status signals:
   - belum tertaut profil
   - belum pernah login
   - nonaktif
   - multi-peran
   - admin/sensitive role
3. Add type-based filtering and search per tab.
4. Add selection toolbar for visible users.
5. Move creation to workflow cards/drawer-like panel:
   - choose account type
   - choose profile source
   - choose roles
   - set username/password
6. Add dedicated Generate Akun workflow:
   - Pegawai preview/generate/export
   - Siswa preview/generate/export
   - Orang Tua preview/generate/export
   - status breakdown ready/created/skipped/failed
7. Improve role/RBAC navigation:
   - show role summary
   - link to `/settings/rbac`
   - warn about admin/sensitive role edits
8. Add audit/request quick links.
9. Keep existing actions functional:
   - reset password
   - force password change if available
   - active/inactive
   - delete as deactivate
   - update profile link
   - update roles

## Files

- Modify: `apps/web-admin/src/routes/settings/users/+page.svelte`
- Modify: `apps/web-admin/src/lib/client/rbac-users.ts` if student/parent generation client helpers are missing.
- Optionally add small helper types/functions inside the page; avoid broad shared refactor.

## Verification

Run:

```bash
npm --prefix apps/web-admin run check
cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-users-workflow ./cmd/api
```

Expected: all pass. If backend is unchanged, sqlc/tests/build should remain pass.

## Commit

```bash
git add apps/web-admin/src/routes/settings/users/+page.svelte apps/web-admin/src/lib/client/rbac-users.ts .hermes/plans/2026-05-16_pengguna-opsi-3-workflow.md
git commit -m "feat(users): add workflow-based account management"
```
