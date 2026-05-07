# RBAC Management Readiness — Phase 8

Dokumen ini mengunci kontrak rilis internal untuk **Full RBAC Management Editor** MTsN 2 Kolaka Utara. Dokumen ini harus tetap sinkron dengan backend Go, SvelteKit BFF, dan UI Web Admin.

## Scope final

### `/settings/rbac`
Halaman pusat manajemen RBAC. Fitur yang tersedia:

- Ringkasan role, permission, mapping role-permission, dan user-role.
- Role metadata controls untuk custom role:
  - create custom role,
  - edit name/description custom role,
  - toggle active custom role.
- Permission Catalog controls:
  - create permission,
  - edit module/action/description,
  - search/filter catalog,
  - toggle status permission non-kritikal.
- Role-permission matrix editor:
  - pilih role,
  - checkbox permission per module,
  - search/filter,
  - preview diff added/removed,
  - warning critical permissions,
  - confirmation overlay,
  - save via `PUT /api/rbac/roles/{code}/permissions`.

### `/settings/users`
Halaman user management tetap fokus pada akun dan assignment role user:

- lifecycle akun,
- status user,
- reset password,
- profile link,
- role assignment user.

CRUD role metadata, Permission Catalog, dan role-permission matrix tidak boleh digandakan di `/settings/users`; arahkan admin ke `/settings/rbac`.

## Endpoint contract

### Read endpoints

- `GET /api/rbac/matrix`
- `GET /api/rbac/roles`
- `GET /api/rbac/permissions`

Akses baca membutuhkan `roles.read` atau `roles.manage` melalui claim permission, dengan fallback role admin.

### Mutating endpoints

- `POST /api/rbac/roles`
- `PATCH /api/rbac/roles/{code}`
- `PATCH /api/rbac/roles/{code}/status`
- `POST /api/rbac/permissions`
- `PATCH /api/rbac/permissions/{code}`
- `PATCH /api/rbac/permissions/{code}/status`
- `PUT /api/rbac/roles/{code}/permissions`
- `PATCH /api/users/{id}/roles`

Mutating endpoint RBAC membutuhkan `roles.manage` atau `users.manage_roles` sesuai domain aksi. Semua endpoint browser-consumed harus punya SvelteKit BFF proxy dan unauthenticated smoke harus mengembalikan `401`, bukan `404`.

## Critical permissions

Daftar critical permission wajib selaras antara backend dan frontend:

- `roles.manage`
- `users.manage_roles`
- `users.manage`
- `permissions.manage`
- `rbac.manage`

Aturan final:

- UI menampilkan warning dan guard untuk critical permission.
- Backend tetap sumber kebenaran.
- Critical permission aktif tidak boleh dicabut dari role `admin` bila hanya ada satu admin aktif.
- Critical permission tidak boleh dinonaktifkan bila hanya ada satu admin aktif.
- Last active admin tidak boleh kehilangan role `admin`.
- Last active admin dan self-demotion harus fail-closed di backend, bukan hanya frontend.

## Session invalidation

**Session invalidation** wajib terjadi setelah perubahan yang berdampak ke privilege:

- `PATCH /api/users/{id}/roles`:
  - sinkronisasi dynamic RBAC ke legacy compatibility roles,
  - increment `auth_version`,
  - revoke active sessions user target,
  - audit `USER_ROLES_UPDATED`.
- `PUT /api/rbac/roles/{code}/permissions`:
  - user yang memegang role terdampak harus di-invalidate,
  - audit `RBAC_ROLE_PERMISSIONS_UPDATED`.
- Role/permission status change:
  - affected active users harus di-invalidate,
  - audit status mutation.

Tujuannya: token lama gagal setelah privilege dicabut; login/token baru membaca role dan permission terkini dari `rbac_user_roles` + `rbac_role_permissions`.

## Guard/test baseline

Regression tests yang menjadi release gate:

- Backend service RBAC:
  - list matrix aggregation,
  - user role replacement,
  - last active admin protection,
  - critical admin permission protection,
  - permission status protection,
  - session invalidation and audit.
- Backend handler RBAC:
  - read endpoint permission guard,
  - mutate endpoint permission guard,
  - invalid payload handling.
- Frontend:
  - `rbac-users.test.ts`,
  - `matrix.test.ts`,
  - `roles.test.ts`,
  - `permissions.test.ts`,
  - `route-access.test.ts`,
  - `sidebar-config.test.ts`,
  - `users-rbac-scope.test.ts`,
  - `rbac-readiness.test.ts`.

## Smoke checklist

Expected unauthenticated smoke:

- Core API `GET /health` → `200`.
- Core API `GET /api/rbac/matrix` → `401`.
- Core API `GET /api/rbac/roles` → `401`.
- Core API `POST /api/rbac/roles` → `401`.
- Core API `PUT /api/rbac/roles/admin/permissions` → `401`.
- Web BFF `GET /api/rbac/matrix` → `401`.
- Web BFF `POST /api/rbac/roles` → `401`.
- Web BFF `PUT /api/rbac/roles/admin/permissions` → `401`.
- Web `/settings/rbac` unauthenticated → `302`.
- Web `/settings/users` unauthenticated → `302`.

Jika PM2 baru direstart dan web-admin sempat `000`, retry singkat sebelum menyatakan gagal.

## Deployment checklist

1. `go test ./...` dari `services/core-api`.
2. Frontend targeted RBAC unit tests.
3. `npm run check` dari `apps/web-admin`.
4. `npm run build` dari `apps/web-admin`.
5. `git diff --check` dan static scan added lines.
6. Build runtime sesuai perubahan:
   - backend: `make build-backend`, restart `deploy/pm2/backend.config.cjs`,
   - frontend: `make build-web`, restart `deploy/pm2/web.config.cjs` bila UI/BFF berubah.
7. PM2 status online.
8. Smoke checklist di atas PASS.

## Phase 8 verdict

Phase 8 dinyatakan PASS bila:

- readiness doc ini ada dan dijaga test,
- semua guard/test baseline lulus,
- smoke checklist lulus,
- tidak ada High/Medium security finding pada diff,
- perubahan final sudah committed,
- dirty unrelated tidak ikut commit.
