# Dynamic RBAC Migration Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Migrate MTsN 2 Kolut Super App from scattered hardcoded role checks to a backend-owned dynamic RBAC system with roles, permissions, role-permission matrix, user-role assignment, backend authorization middleware, frontend capability contract, and admin UI for safe role management.

**Architecture:** Keep PostgreSQL owned by `services/core-api`. Add dynamic RBAC tables and services in the Go API, preserve compatibility with the current `user_account_roles`/`user_role` enum during transition, and gradually migrate backend routes plus SvelteKit guards/sidebar from role-based checks to permission-based capabilities. Security-critical invariants remain enforced in backend, not only frontend.

**Tech Stack:** Go Chi API + sqlc + PostgreSQL migrations; SvelteKit 2/Svelte 5 web admin; existing JWT/session/auth_version flow; PM2 deployment.

---

## Current Baseline

Relevant files:

- Backend routes: `services/core-api/cmd/api/main.go`
- Backend auth middleware: `services/core-api/internal/middleware/auth.go`
- Backend user handler: `services/core-api/internal/handler/user.go`
- Backend auth service: `services/core-api/internal/service/auth.go`
- Backend user lifecycle: `services/core-api/internal/service/user_lifecycle.go`
- User queries: `services/core-api/db/queries/users.sql`
- Auth sessions queries: `services/core-api/db/queries/auth_sessions.sql`
- Web route guard: `apps/web-admin/src/hooks.server.ts`
- Web route access helper: `apps/web-admin/src/lib/server/route-access.ts`
- Web auth model/helper: `apps/web-admin/src/lib/server/auth.ts`
- Sidebar config: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- User management page: `apps/web-admin/src/routes/settings/users/+page.svelte`

Existing roles:

- `admin`
- `guru`
- `staf`
- `kesiswaan`
- `siswa`
- `ortu`

Existing tables:

- `users`
- `user_account_roles`
- `auth_sessions`
- `audit_logs`
- profile entities: `employees`, `students`, `parents`

Key safety requirements:

1. Do not break existing login/session behavior.
2. Do not remove current roles during initial migration.
3. Do not rely on frontend-only protection for admin/user safety.
4. Preserve existing `auth_version` semantics for session invalidation.
5. All user/role/permission mutations must be auditable.
6. Prevent last active admin lockout.

---

## Target Model

### Role assignment

A user can have many roles. Roles can be system roles or custom roles.

### Permission assignment

A role can have many permissions. Permissions are canonical capability strings, e.g.:

- `users.read`
- `users.create`
- `users.update`
- `users.deactivate`
- `users.reset_password`
- `users.manage_roles`
- `roles.read`
- `roles.manage`
- `audit.read`
- `bank_soal.read`
- `bank_soal.create`
- `bank_soal.review`
- `asesmen.read`
- `asesmen.proctor`
- `pusaka.read`
- `pusaka.manage`

### Authorization flow

1. Login verifies user and active session.
2. Backend resolves user roles and permissions.
3. JWT/session response includes `roles` and `permissions`.
4. Backend middleware authorizes protected routes with permissions.
5. Frontend uses `permissions` only to show/hide routes/menu; backend remains source of truth.
6. Any role/permission/user mutation increments affected users' `auth_version` or revokes sessions where appropriate.

---

## Database Design

### New tables

Create migration `services/core-api/db/migrations/069_dynamic_rbac_foundation.sql`.

Recommended DDL:

```sql
CREATE TABLE IF NOT EXISTS rbac_roles (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  code        TEXT        NOT NULL UNIQUE,
  name        TEXT        NOT NULL,
  description TEXT        NOT NULL DEFAULT '',
  is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (code ~ '^[a-z][a-z0-9_]*$')
);

CREATE TABLE IF NOT EXISTS rbac_permissions (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  code        TEXT        NOT NULL UNIQUE,
  module      TEXT        NOT NULL,
  action      TEXT        NOT NULL,
  description TEXT        NOT NULL DEFAULT '',
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (code ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$')
);

CREATE TABLE IF NOT EXISTS rbac_role_permissions (
  role_id       UUID        NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
  permission_id UUID        NOT NULL REFERENCES rbac_permissions(id) ON DELETE CASCADE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS rbac_user_roles (
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id    UUID        NOT NULL REFERENCES rbac_roles(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_rbac_role_permissions_permission ON rbac_role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_rbac_user_roles_role ON rbac_user_roles(role_id);
```

### Compatibility strategy

During phases 1-3:

- Keep `user_account_roles` untouched.
- Mirror existing enum roles into `rbac_roles`.
- Backfill `rbac_user_roles` from `user_account_roles`.
- Auth can read RBAC tables first, then fallback to `user_account_roles` if needed.

After all routes/UI use dynamic RBAC and tests pass, plan a separate cleanup phase to retire old role enum dependency. Do not do cleanup in the foundation phase.

---

## Default Permissions

Seed the following initial permissions. Keep the first batch intentionally small and stable.

### System/User

- `dashboard.read`
- `notifications.read`
- `settings.account`
- `settings.school_profile`
- `audit.read`
- `users.read`
- `users.create`
- `users.update`
- `users.deactivate`
- `users.reset_password`
- `users.manage_roles`
- `roles.read`
- `roles.manage`

### Academic

- `academic.read`
- `academic.manage`
- `students.read`
- `students.manage`
- `parents.read`
- `parents.manage`
- `grades.read`
- `grades.manage`
- `journal.read`
- `journal.manage`

### Bank Soal

- `bank_soal.read`
- `bank_soal.create`
- `bank_soal.update`
- `bank_soal.review`
- `bank_soal.publish`
- `bank_soal.import`
- `bank_soal.delete`
- `bank_soal.analytics`
- `bank_soal.settings`

### Asesmen

- `asesmen.read`
- `asesmen.package_manage`
- `asesmen.event_manage`
- `asesmen.session_manage`
- `asesmen.participant_manage`
- `asesmen.proctor`
- `asesmen.score`
- `asesmen.result_read`
- `asesmen.result_manage`

### Kesiswaan/TU/PUSAKA

- `kesiswaan.read`
- `kesiswaan.manage`
- `letters.read`
- `letters.manage`
- `archives.read`
- `archives.manage`
- `inventory.read`
- `inventory.manage`
- `library.read`
- `library.manage`
- `governance.read`
- `governance.manage`
- `document_cycles.read`
- `document_cycles.manage`
- `pusaka.read`
- `pusaka.manage`
- `pusaka.sync`
- `pusaka.credentials_manage`
- `website.read`
- `website.manage`

---

## Default Role Matrix

### `admin`

- all permissions

### `guru`

- `dashboard.read`
- `notifications.read`
- `settings.account`
- `students.read`
- `grades.read`
- `grades.manage`
- `journal.read`
- `journal.manage`
- `bank_soal.read`
- `bank_soal.create`
- `bank_soal.update`
- `bank_soal.review`
- `bank_soal.import`
- `bank_soal.analytics`
- `bank_soal.settings`
- `asesmen.read`
- `asesmen.score`
- `asesmen.result_read`

### `staf`

- `dashboard.read`
- `notifications.read`
- `settings.account`
- `letters.read`
- `letters.manage`
- `archives.read`
- `archives.manage`
- `inventory.read`
- `inventory.manage`
- `library.read`
- `library.manage`
- `governance.read`
- `governance.manage`
- `document_cycles.read`
- `document_cycles.manage`
- `asesmen.read`
- `asesmen.proctor`

### `kesiswaan`

- `dashboard.read`
- `notifications.read`
- `settings.account`
- `students.read`
- `students.manage`
- `parents.read`
- `kesiswaan.read`
- `kesiswaan.manage`

### `siswa`

- `dashboard.read`
- `notifications.read`
- `settings.account`

### `ortu`

- `dashboard.read`
- `notifications.read`
- `settings.account`

Future recommended custom roles after foundation:

- `operator`
- `kepala_madrasah`
- `wali_kelas`
- `proktor`
- `staf_tu`
- `pustakawan`

---

## Phase 0 — Preflight and Safety Snapshot

### Task 0.1: Verify repo and runtime state

**Objective:** Ensure no uncommitted work and current app health is known before RBAC migration.

**Files:** none.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git log --oneline -5
pm2 list
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/health
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8021/
```

**Expected:** working tree clean; core-api `200`; web-admin `200`.

### Task 0.2: Snapshot current RBAC data

**Objective:** Capture current users/roles before schema changes.

**Files:** optional output file under `docs/rbac/current-rbac-snapshot-YYYY-MM-DD.md`.

**Command:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
set -a; source services/core-api/.env; set +a
psql "$DATABASE_URL" -Atc "
SELECT 'users='||count(*) FROM users;
SELECT 'active='||count(*) FROM users WHERE is_active;
SELECT role||'='||count(*) FROM user_account_roles GROUP BY role ORDER BY role;
SELECT 'multi_role='||count(*) FROM (SELECT user_id FROM user_account_roles GROUP BY user_id HAVING count(*) > 1) x;
"
```

**Expected:** returns current user and role counts.

---

## Phase 1 — Database Foundation

### Task 1.1: Add RBAC migration

**Objective:** Create dynamic RBAC tables and seed default roles/permissions/matrix.

**Files:**

- Create: `services/core-api/db/migrations/069_dynamic_rbac_foundation.sql`

**Implementation notes:**

1. Create tables from the Database Design section.
2. Seed default roles using `INSERT ... ON CONFLICT (code) DO UPDATE`.
3. Seed default permissions using `INSERT ... ON CONFLICT (code) DO UPDATE`.
4. Seed role-permission matrix by joining `rbac_roles.code` and `rbac_permissions.code`.
5. Backfill user-role assignments:

```sql
INSERT INTO rbac_user_roles (user_id, role_id)
SELECT uar.user_id, rr.id
FROM user_account_roles uar
JOIN rbac_roles rr ON rr.code = uar.role::text
ON CONFLICT (user_id, role_id) DO NOTHING;
```

**Verification:**

```bash
set -a; source services/core-api/.env; set +a
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f services/core-api/db/migrations/069_dynamic_rbac_foundation.sql
psql "$DATABASE_URL" -Atc "SELECT count(*) FROM rbac_roles; SELECT count(*) FROM rbac_permissions; SELECT count(*) FROM rbac_role_permissions; SELECT count(*) FROM rbac_user_roles;"
```

**Expected:** roles >= 6, permissions >= 40, role_permissions > 0, user_roles matches current assigned users.

### Task 1.2: Add sqlc queries

**Objective:** Add typed queries for RBAC service.

**Files:**

- Create: `services/core-api/db/queries/rbac.sql`

**Queries to add:**

- `ListRbacRoles`
- `ListRbacPermissions`
- `ListRbacRolePermissions`
- `ListRbacUserRoles`
- `GetUserPermissionCodes`
- `GetUserRoleCodesFromRbac`
- `ReplaceRolePermissions` cannot be one sqlc query; use transaction with `DeleteRolePermissions` + `AddRolePermissionByCode`.
- `DeleteRolePermissions`
- `AddRolePermissionByCode`
- `DeleteUserRbacRoles`
- `AddUserRbacRoleByCode`
- `CountActiveAdminsByRbac`
- `UserHasRbacRole`
- `ListUserIDsByRoleCode`

**Example query:**

```sql
-- name: GetUserPermissionCodes :many
SELECT DISTINCT p.code
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
JOIN rbac_role_permissions rp ON rp.role_id = r.id
JOIN rbac_permissions p ON p.id = rp.permission_id
WHERE ur.user_id = $1
  AND r.is_active = TRUE
  AND p.is_active = TRUE
ORDER BY p.code;
```

**Verification:**

```bash
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go generate ./...
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./...
```

---

## Phase 2 — Backend RBAC Service

### Task 2.1: Create RBAC service

**Objective:** Centralize role/permission read and mutation logic.

**Files:**

- Create: `services/core-api/internal/service/rbac.go`
- Create: `services/core-api/internal/service/rbac_test.go`

**Service methods:**

```go
type RBAC struct { /* q + tx */ }

func (s *RBAC) ListMatrix(ctx context.Context) (Matrix, error)
func (s *RBAC) GetUserPermissions(ctx context.Context, userID pgtype.UUID) ([]string, error)
func (s *RBAC) GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]string, error)
func (s *RBAC) ReplaceRolePermissions(ctx context.Context, roleCode string, permissionCodes []string, actorID pgtype.UUID) error
func (s *RBAC) ReplaceUserRoles(ctx context.Context, userID pgtype.UUID, roleCodes []string, actorID pgtype.UUID) error
func (s *RBAC) EnsureCanChangeUserRoles(ctx context.Context, targetUserID pgtype.UUID, nextRoleCodes []string, actorID pgtype.UUID) error
```

**Safety rules:**

- Cannot remove the last active admin.
- Cannot remove own admin/role-management capability if actor would lock themselves out.
- Cannot assign inactive roles.
- Cannot assign unknown roles.
- `admin` system role should always retain all permissions, or at least all critical permissions.

**Tests:**

- replacing a normal user's roles succeeds
- assigning unknown role fails
- removing last active admin fails
- admin role cannot lose `roles.manage`/`users.manage_roles` if no other active admin exists

### Task 2.2: Add RBAC handler

**Objective:** Expose admin endpoints for matrix read/update.

**Files:**

- Create: `services/core-api/internal/handler/rbac.go`
- Create: `services/core-api/internal/handler/rbac_test.go`
- Modify: `services/core-api/cmd/api/main.go`

**Endpoints:**

```text
GET /api/rbac/roles
GET /api/rbac/permissions
GET /api/rbac/matrix
PUT /api/rbac/roles/{code}/permissions
PATCH /api/users/{id}/roles
```

**Authorization for initial implementation:**

- Use existing `requireAdmin` first.
- Switch to `RequirePermission("roles.manage")` in Phase 3 after middleware exists.

**Audit actions:**

- `RBAC_ROLE_PERMISSIONS_UPDATED`
- `USER_ROLES_UPDATED`

**Tests:**

- forbidden without admin context
- list matrix returns roles/permissions/mappings
- update role permissions validates body
- update user roles validates last-admin rule

---

## Phase 3 — Permission Middleware and Auth Contract

### Task 3.1: Extend auth user model with permissions

**Objective:** Include `permissions` in backend current-user/login responses and frontend auth type.

**Files:**

- Modify: `services/core-api/internal/service/auth.go`
- Modify: `services/core-api/internal/handler/auth.go`
- Modify: `apps/web-admin/src/lib/server/auth.ts`
- Modify related tests in `services/core-api/internal/service/*auth*_test.go` and `apps/web-admin/src/lib/server/*auth*.test.ts` if present.

**Behavior:**

- Resolve RBAC permissions for authenticated user.
- Include `permissions: []string` in current user response.
- Preserve `roles` for compatibility.
- If RBAC tables are missing in a dev database, fail clearly; do not silently allow elevated access.

**Verification:**

```bash
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./...
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run test:unit -- src/lib/server/auth.test.ts
```

### Task 3.2: Add backend permission middleware

**Objective:** Support permission-based route gates.

**Files:**

- Modify: `services/core-api/internal/middleware/auth.go`
- Add tests: `services/core-api/internal/middleware/auth_test.go`

**Middleware API:**

```go
func RequirePermission(permission string) func(http.Handler) http.Handler
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler
```

**Context requirement:**

- JWT middleware must place `permissions` in request context, similar to roles.

**Tests:**

- allowed when permission exists
- forbidden when missing
- forbidden when unauthenticated
- `RequireAnyPermission` allows any one permission

### Task 3.3: Route pilot migration

**Objective:** Convert the most sensitive admin routes to permission checks first.

**Files:**

- Modify: `services/core-api/cmd/api/main.go`

**Routes to migrate first:**

```text
/api/users -> users.read/users.create/users.deactivate/users.manage_roles
/api/users/audit-logs -> audit.read
/api/rbac/* -> roles.read/roles.manage
/api/school-profile PUT -> settings.school_profile
```

**Verification:**

```bash
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./...
```

---

## Phase 4 — User Management Hardening

### Task 4.1: Add user mutation endpoints

**Objective:** Complete admin user lifecycle before exposing dynamic role UI.

**Files:**

- Modify: `services/core-api/internal/handler/user.go`
- Modify: `services/core-api/internal/service/user_lifecycle.go`
- Modify: `services/core-api/db/queries/users.sql`
- Modify tests: `services/core-api/internal/handler/user_success_test.go`, `services/core-api/internal/service/user_lifecycle_test.go`

**Endpoints:**

```text
PATCH /api/users/{id}/roles
POST /api/users/{id}/reset-password
PATCH /api/users/{id}/profile-link
```

**Safety:**

- reset-password must validate password policy
- changing profile link must preserve one-profile-type invariant
- role change must preserve last active admin
- deactivating admin must preserve last active admin
- self-demotion must not remove the actor's ability to manage users/roles unless another admin exists

### Task 4.2: Session invalidation on role/permission changes

**Objective:** Ensure permission changes take effect quickly and safely.

**Files:**

- Modify: `services/core-api/internal/service/rbac.go`
- Modify: `services/core-api/db/queries/users.sql`
- Modify: `services/core-api/db/queries/auth_sessions.sql` if needed

**Behavior:**

- On user role replacement: increment target user's `auth_version` and revoke all active sessions except optionally actor's current session if actor is target and still authorized.
- On role permission replacement: increment `auth_version` for all users assigned to that role; optionally revoke active sessions.

**Verification:** tests assert `IncrementUserAuthVersion`/session revocation is called.

---

## Phase 5 — Frontend Capability Contract

### Task 5.1: Add frontend permission helpers

**Objective:** Make frontend route/menu logic capability-based.

**Files:**

- Modify: `apps/web-admin/src/lib/server/auth.ts`
- Create: `apps/web-admin/src/lib/auth/permissions.ts`
- Create: `apps/web-admin/src/lib/auth/permissions.test.ts`

**Helpers:**

```ts
export function userPermissions(user: AuthUser | undefined): string[];
export function hasPermission(user: AuthUser | undefined, permission: string): boolean;
export function hasAnyPermission(user: AuthUser | undefined, permissions: readonly string[]): boolean;
```

**Compatibility:**

- During migration, if `permissions` is empty, keep fallback role mapping only in frontend tests/dev if necessary. Production should depend on backend permissions.

### Task 5.2: Migrate web route guard

**Objective:** Replace role-specific guards with permission-specific guards.

**Files:**

- Modify: `apps/web-admin/src/lib/server/route-access.ts`
- Modify: `apps/web-admin/src/hooks.server.ts`
- Modify tests: `apps/web-admin/src/hooks.server.test.ts`, `apps/web-admin/src/lib/server/route-access.test.ts`

**Mapping examples:**

- `/settings/users` -> `users.read`
- `/settings/audit-logs` -> `audit.read`
- `/bank-soal` -> `bank_soal.read`
- `/bank-soal/tambah` -> `bank_soal.create`
- `/bank-soal/verifikasi` -> `bank_soal.review`
- `/bank-soal/analisis-butir` -> `bank_soal.analytics`
- `/bank-soal/pengaturan` -> `bank_soal.settings`
- `/asesmen/paket` -> `asesmen.package_manage` for mutations; `asesmen.read` for safe reads
- `/pusaka` -> `pusaka.read`
- `/pusaka/employees` credentials mutation -> `pusaka.credentials_manage`

### Task 5.3: Migrate sidebar config

**Objective:** Sidebar visibility follows permissions, not hardcoded roles.

**Files:**

- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Modify: `apps/web-admin/src/lib/components/Sidebar.svelte`
- Modify tests: `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`, `apps/web-admin/src/lib/components/sidebar/sidebar-active.test.ts`

**Type change:**

```ts
export type SidebarNavItem = {
  href: string;
  label: string;
  icon: string;
  permissions?: string[];
  anyPermissions?: string[];
  roles?: string[]; // temporary compatibility only
  pinnable?: boolean;
};
```

**Completion criterion:** New/sidebar tests verify `guru` equivalent permission set can see Bank Soal and Asesmen pages without relying on role strings.

---

## Phase 6 — RBAC Admin UI

### Task 6.1: Add `/settings/roles` page

**Objective:** Let admin manage dynamic role-permission matrix.

**Files:**

- Create: `apps/web-admin/src/routes/settings/roles/+page.svelte`
- Add sidebar entry in `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Add route guard in `apps/web-admin/src/lib/server/route-access.ts`

**Features:**

- list roles
- list permissions grouped by module
- show permission matrix as grouped checkboxes
- edit role display name/description later; foundation can focus on permission matrix
- disable destructive changes for `admin` if it would remove critical permissions
- show warning when a role change will revoke sessions

### Task 6.2: Enhance `/settings/users`

**Objective:** Use dynamic role list and expose safe role editing.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/users/+page.svelte`

**Changes:**

- load roles from `/api/rbac/roles`
- remove hardcoded role chip list for create/edit
- add edit roles action for existing users
- add reset password action
- add profile link edit action
- show whether user has dynamic permission count
- show guard badge for last active admin

**Performance note:** Do not load all employees/students/parents forever. If time permits in this phase, introduce async searchable pickers; otherwise track it as follow-up.

---

## Phase 7 — Full Route Migration

### Task 7.1: Convert backend route groups by module

**Objective:** Replace remaining role middleware with permission middleware module-by-module.

**Files:**

- Modify: `services/core-api/cmd/api/main.go`
- Add/modify backend route access tests if available.

**Recommended order:**

1. Users/RBAC/Audit
2. PUSAKA
3. Website/School profile
4. Bank Soal
5. Asesmen
6. Students/Parents/Kesiswaan
7. TU/Library/Inventory/Governance/Document Cycles
8. Grades/Journal/Academic

**Important:** Preserve guru data-scoping service logic. Permission grants route entry; service-level scope still limits data to assigned classes/events where applicable.

### Task 7.2: Convert frontend routes by module

**Objective:** Remove role-based gates from web admin except compatibility-only fallbacks.

**Files:**

- Modify: `apps/web-admin/src/lib/server/route-access.ts`
- Modify: `apps/web-admin/src/hooks.server.ts`
- Modify route/sidebar tests.

**Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
```

---

## Phase 8 — CRUD Role & Permission Management

Implemented in commit `d841b1c`: backend CRUD role/permission endpoints, service/handler safety guards, audit actions, frontend RBAC client helpers, and transitional management controls in `/settings/users`.

### Task 8.1: Dynamic role and permission management

**Objective:** allow authorized operators to create/update/activate/deactivate custom roles and permissions without code changes while preserving backend safety invariants.

**Implemented files:**

- `services/core-api/db/queries/rbac.sql`
- `services/core-api/internal/service/rbac.go`
- `services/core-api/internal/handler/rbac.go`
- `apps/web-admin/src/lib/client/rbac-users.ts`
- `apps/web-admin/src/routes/settings/users/+page.svelte`

---

## Phase 9 — Final Stabilization & Cleanup Readiness

Phase 9 does **not** remove legacy role fallback yet. It freezes the permission catalog, documents the remaining compatibility policy, and adds regression tests so route/sidebar permission strings cannot drift from DB seed permissions.

### Task 9.1: Permission catalog and guard tests

**Objective:** make permission codes operator-visible and testable across DB seed, frontend route guards, and sidebar metadata.

**Files:**

- Add: `docs/rbac-permission-catalog.md`
- Add: `apps/web-admin/src/lib/rbac/permission-catalog.ts`
- Add: `apps/web-admin/src/lib/rbac/permission-catalog.test.ts`

### Task 9.2: Decide retirement of old `user_account_roles`

**Objective:** Remove old enum role dependency only after dynamic RBAC is proven.

**Options:**

- Option A: Keep `user_account_roles` as read-only compatibility for one release cycle.
- Option B: Replace it with a view over `rbac_user_roles` for compatibility.
- Option C: Fully remove and update sqlc/generated code.

Recommended: Option A first, then Option B/C in a separate cleanup sprint.

### Task 9.3: Future removal of frontend role fallbacks

**Objective:** Ensure all UI uses `permissions` only.

**Files:**

- `apps/web-admin/src/lib/server/route-access.ts`
- `apps/web-admin/src/hooks.server.ts`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

**Guard test:** add tests that fail if new sidebar items use `roles` instead of `permissions`.

---

## Validation Commands

Run after every implementation phase:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
cd services/core-api
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go test ./...
PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH go build -o bin/api ./cmd/api
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git diff --check && git diff --cached --check
```

After deployment:

```bash
pm2 restart mtsn2kolut-core-api --update-env
pm2 restart mtsn2kolut-web-admin --update-env
curl -s -o /dev/null -w "core-api %{http_code}\n" http://127.0.0.1:8080/health
curl -s -o /dev/null -w "web-admin %{http_code}\n" http://127.0.0.1:8021/
```

Smoke pages as admin:

- `/settings/users`
- `/settings/roles`
- `/settings/audit-logs`
- `/bank-soal`
- `/asesmen`
- `/pusaka`

Smoke permission denial:

- a non-admin/non-role-manager must not access `/settings/roles`
- a user without `bank_soal.read` must not access `/bank-soal`
- a user without `users.manage_roles` must not update roles via API

---

## Rollback Strategy

### If migration fails before deploy

- Do not commit.
- Drop new RBAC tables only if no production dependency exists:

```sql
DROP TABLE IF EXISTS rbac_user_roles;
DROP TABLE IF EXISTS rbac_role_permissions;
DROP TABLE IF EXISTS rbac_permissions;
DROP TABLE IF EXISTS rbac_roles;
```

### If deployed backend fails

1. Revert commit.
2. Rebuild backend.
3. Restart PM2 core-api.
4. Keep RBAC tables; they are additive and should not break old code.

### If permission matrix locks admin out

Use database emergency restore:

```sql
INSERT INTO rbac_roles (code, name, is_system, is_active)
VALUES ('admin', 'Administrator', TRUE, TRUE)
ON CONFLICT (code) DO UPDATE SET is_active = TRUE;

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
CROSS JOIN rbac_permissions p
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO rbac_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN rbac_roles r ON r.code = 'admin'
WHERE u.username = 'admin'
ON CONFLICT DO NOTHING;
```

Then restart core-api and login again.

---

## Commit Plan

Use small commits:

1. `docs: add dynamic rbac migration plan`
2. `feat(db): add dynamic rbac foundation tables`
3. `feat(api): add rbac queries and service`
4. `feat(api): expose rbac management endpoints`
5. `feat(api): add permission middleware`
6. `feat(auth): include permissions in auth contract`
7. `feat(web-admin): use permissions for route guards`
8. `feat(web-admin): add dynamic roles management UI`
9. `refactor(authz): migrate module routes to permissions`
10. `test(rbac): add permission matrix regression coverage`

---

## Acceptance Criteria

Dynamic RBAC migration is complete when:

- RBAC roles/permissions/matrix exist in database and are seeded idempotently.
- Existing users are backfilled into dynamic role assignments.
- Login/current user responses include `permissions`.
- Backend can protect routes with `RequirePermission`/`RequireAnyPermission`.
- `/settings/users` supports safe dynamic role assignment plus transitional role/permission management controls.
- Sidebar and frontend route guards use permissions with documented legacy role fallback.
- Last active admin cannot be deactivated, deleted, or stripped of role-management capability.
- Role/permission changes are audit logged.
- Role/permission changes invalidate affected user sessions or auth versions.
- Full validation commands pass.
- PM2 restart and health checks pass.
