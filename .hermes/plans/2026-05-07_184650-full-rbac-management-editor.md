# Full RBAC Management Editor Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menyempurnakan manajemen RBAC agar admin dapat mengelola role, permission, role-permission mapping, user-role assignment, dan dampak perubahan secara aman dari Web Admin.

**Architecture:** Backend Go sudah memiliki fondasi RBAC dinamis dan endpoint `PUT /api/rbac/roles/{code}/permissions`; patch utama adalah melengkapi BFF SvelteKit, client helper, UI matrix/drawer, audit/preview dampak, dan test kontrak agar UI benar-benar bisa mengatur permission per role. Role sistem tetap dilindungi untuk metadata/status, tetapi permission role sistem boleh diatur melalui guard backend yang sudah ada.

**Tech Stack:** Go/Chi/sqlc/PostgreSQL core-api, SvelteKit 2/Svelte 5 web-admin, shadcn-svelte components, Vitest, Go tests, PM2 deploy.

---

## 1. Current Context

Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`

Temuan terakhir:

- Backend RBAC cukup kuat:
  - `services/core-api/internal/handler/rbac.go`
  - `services/core-api/internal/service/rbac.go`
  - route backend di `services/core-api/cmd/api/main.go`
  - tersedia `PUT /api/rbac/roles/{code}/permissions`
- Frontend saat ini belum lengkap:
  - `/settings/users` bisa edit metadata custom role: `name`, `description`
  - role sistem tidak bisa diedit metadata/status, by design
  - belum ada UI untuk assign/revoke permission pada role
  - belum ada BFF route `/api/rbac/roles/[code]/permissions/+server.ts`
  - belum ada client helper `updateRBACRolePermissions(...)`
- Data shape perlu dirapikan:
  - backend matrix mengembalikan role-permission rows
  - frontend idealnya memakai map `Record<roleCode, string[]>` untuk UI matrix
- Sasaran release:
  - RBAC bisa diklaim lengkap/mumpuni untuk manajemen role-permission operasional
  - tidak mengorbankan guard admin terakhir, self-lockout, role sistem, dan critical permissions

---

## 2. Release Acceptance Criteria

Patch dianggap selesai jika:

1. Admin dapat membuat custom role.
2. Admin dapat edit metadata custom role.
3. Admin tidak dapat edit metadata role sistem (`admin`, `guru`, `staf`, `kesiswaan`, `siswa`, `ortu`).
4. Admin dapat mengatur permission role melalui UI matrix/drawer.
5. Permission role sistem bisa diatur secara aman, kecuali guard backend menolak perubahan berbahaya.
6. Admin mendapat preview dampak sebelum simpan:
   - permission ditambah
   - permission dicabut
   - jumlah user terdampak
   - warning session/token invalidation
   - warning permission kritikal
7. Semua perubahan role-permission memanggil backend `PUT /api/rbac/roles/{code}/permissions` melalui BFF.
8. Setelah role-permission berubah, user terdampak invalidated/revoked oleh backend existing flow.
9. UI tidak menyebut “Edit Role” secara ambigu; dipisah menjadi “Edit Info” dan “Atur Permission”.
10. Backend/BFF/frontend unauth smoke untuk RBAC endpoint return `401`, bukan `404`.
11. Quality gates pass:
    - `go test ./...`
    - targeted Vitest RBAC/users tests
    - `npm run check`
    - `npm run build`
    - `make build-backend build-web build-worker`
    - PM2 restart + smoke
12. Worktree clean dan commit dibuat.

---

## 3. Non-Goals / Batasan

- Jangan mengubah kode role sistem (`admin`, `guru`, dst.) karena `code` sebaiknya immutable.
- Jangan membuka edit metadata role sistem kecuali user secara eksplisit meminta perubahan kebijakan.
- Jangan menghapus guard backend last-admin/self-lockout/critical permissions.
- Jangan membuat permission baru sembarangan tanpa migration/seed yang jelas.
- Jangan menyimpan atau menampilkan credential, token, secret, DB URL; selalu redact.
- Jangan memindahkan PostgreSQL access ke frontend/worker; tetap via core-api.

---

## 4. Proposed UX

Tambahkan panel atau halaman RBAC khusus.

Rekomendasi final: buat halaman baru:

- `/settings/rbac`

Dan biarkan `/settings/users` fokus ke user lifecycle.

Namun untuk patch minimal-kompatibel, bisa mulai dari panel di `/settings/users`, lalu refactor ke `/settings/rbac`. Plan ini memakai pendekatan yang lebih rapi: **buat `/settings/rbac`**.

### Layout `/settings/rbac`

1. Header:
   - “Manajemen RBAC”
   - copy: “Kelola role, permission, dan hak akses secara dinamis.”

2. Summary cards:
   - total role
   - total permission
   - role sistem
   - role custom

3. Tabs/sections:
   - **Role**: daftar role, create role, edit info, status
   - **Permission**: daftar permission, create/edit/status
   - **Matrix Akses**: role-permission editor

4. Matrix editor:
   - kiri: daftar role
   - kanan: permission grouped by module
   - checkbox per permission
   - search permission
   - filter module
   - tombol:
     - “Reset Perubahan”
     - “Simpan Permission Role”

5. Confirmation dialog:
   - role: `guru`
   - permission ditambah: list
   - permission dicabut: list
   - user terdampak: angka jika tersedia, atau copy “Semua user dengan role ini akan login ulang”
   - warning critical permission jika ada

---

## 5. Task-by-Task Plan

### Task 1: Add BFF proxy for role-permission update

**Objective:** Browser dapat memanggil backend `PUT /api/rbac/roles/{code}/permissions` lewat SvelteKit BFF.

**Files:**

- Create: `apps/web-admin/src/routes/api/rbac/roles/[code]/permissions/+server.ts`
- Test: existing/new BFF proxy tests if available under `apps/web-admin/src/routes/api/**`

**Implementation sketch:**

```ts
import { json } from '@sveltejs/kit';
import type { RequestEvent } from '@sveltejs/kit';
import { apiPath, proxy, handleRouteError, readRequestJson, requiredRouteParam } from '$lib/server/api';

export const PUT = async (event: RequestEvent) => {
	try {
		const code = requiredRouteParam(event.params.code, 'code');
		const body = await readRequestJson<Record<string, unknown>>(event.request);
		const res = await proxy(event).put(apiPath`/api/rbac/roles/${code}/permissions`, body);
		return json(res);
	} catch (e) {
		return handleRouteError(e, 'rbac role permissions PUT');
	}
};
```

**Verify:**

```bash
cd apps/web-admin
npm run test:unit -- src/routes/api/rbac/**/*.test.ts
```

If no existing route test pattern exists, add a focused test following existing BFF proxy conventions.

---

### Task 2: Add client helper for role-permission update

**Objective:** UI punya helper typed untuk menyimpan permission role.

**Files:**

- Modify: `apps/web-admin/src/lib/client/rbac-users.ts`
- Test: `apps/web-admin/src/lib/client/rbac-users.test.ts` if exists, otherwise create.

**Add function:**

```ts
export async function updateRBACRolePermissions(code: string, permissions: string[]) {
	const res = await fetch(`/api/rbac/roles/${encodeURIComponent(code)}/permissions`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ permissions })
	});
	return readClientJson<{ ok: boolean }>(res);
}
```

**Important:** use existing `readClientJson`/client API helper style already in file.

**Verify:**

```bash
cd apps/web-admin
npm run test:unit -- src/lib/client/rbac-users.test.ts
```

---

### Task 3: Normalize RBAC matrix data shape for frontend

**Objective:** UI dapat memakai role-permission mapping dengan mudah dan tidak tergantung shape mentah backend.

**Files:**

- Modify: `apps/web-admin/src/lib/client/rbac-users.ts`
- Test: `apps/web-admin/src/lib/client/rbac-users.test.ts`

**Recommended types:**

```ts
export type RBACRolePermissionRow = {
	role_code: string;
	permission_code: string;
};

export type RBACMatrix = {
	roles: RBACRole[];
	permissions: RBACPermission[];
	role_permissions: RBACRolePermissionRow[];
	user_roles?: unknown[];
};

export function rolePermissionMap(matrix: RBACMatrix): Record<string, string[]> {
	const map: Record<string, string[]> = {};
	for (const role of matrix.roles ?? []) map[role.code] = [];
	for (const row of matrix.role_permissions ?? []) {
		if (!map[row.role_code]) map[row.role_code] = [];
		if (!map[row.role_code].includes(row.permission_code)) {
			map[row.role_code].push(row.permission_code);
		}
	}
	for (const key of Object.keys(map)) map[key].sort();
	return map;
}
```

**Tests:**

- empty matrix returns empty object
- known roles get empty arrays
- duplicate rows are deduped
- permissions sorted

---

### Task 4: Add RBAC route access and sidebar entry

**Objective:** `/settings/rbac` terlindungi dan muncul hanya untuk user dengan akses role management.

**Files:**

- Modify: `apps/web-admin/src/lib/server/route-access.ts`
- Modify: `apps/web-admin/src/lib/server/route-access.test.ts`
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`

**Policy:**

- `/settings/rbac` requires `roles.read` for viewing.
- Mutating buttons require `roles.manage` via backend/BFF; frontend may hide/disable if user lacks `roles.manage`.
- Keep admin fallback if current pattern uses it.

**Test cases:**

- plain guru cannot access `/settings/rbac`
- user with `roles.read` can access
- sidebar item visible for `roles.read`

---

### Task 5: Create `/settings/rbac` page shell

**Objective:** Buat halaman RBAC dedicated dengan data load, recovery state, dan tabs/sections awal.

**Files:**

- Create: `apps/web-admin/src/routes/settings/rbac/+page.svelte`
- Optional create component folder: `apps/web-admin/src/routes/settings/rbac/_components/`

**Page behavior:**

- onMount fetch `fetchRBACMatrix()`
- show `AsyncContent`/`RecoveryPanel` pattern like `/settings/users`
- show summary cards
- section Role, Permission, Matrix

**Do not yet remove existing RBAC panel from `/settings/users` in this task.** Keep compatible first.

**Verify:**

```bash
cd apps/web-admin
npm run check
```

---

### Task 6: Extract shared RBAC role/permission management logic

**Objective:** Hindari duplikasi terlalu besar antara `/settings/users` dan `/settings/rbac`.

**Files:**

- Create: `apps/web-admin/src/lib/rbac/matrix.ts`
- Optional create: `apps/web-admin/src/lib/rbac/role-permissions.ts`
- Modify: `apps/web-admin/src/routes/settings/rbac/+page.svelte`
- Later modify: `apps/web-admin/src/routes/settings/users/+page.svelte`

**Helpers:**

- `rolePermissionMap(matrix)`
- `permissionsByModule(permissions)`
- `diffPermissions(before, after)`
- `isCriticalPermission(code)`

**Tests:**

- Create: `apps/web-admin/src/lib/rbac/matrix.test.ts`

Test:

- grouping by module
- diff add/remove unchanged
- critical detection for `roles.manage`, `users.manage_roles`

---

### Task 7: Implement Role Info management in `/settings/rbac`

**Objective:** Custom role metadata management tetap ada di halaman baru.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/rbac/+page.svelte`

**Behavior:**

- `+ Role` calls existing `createRBACRole`
- `Edit Info` calls existing `updateRBACRole`
- `Nonaktifkan/Aktifkan` calls existing `setRBACRoleActive`
- disable metadata/status buttons for `role.is_system`

**UX copy:**

- Replace ambiguous `Edit` with `Edit Info`
- Show badge `System` and tooltip/copy: “Role sistem tidak dapat diubah info/statusnya, tetapi permission dapat diatur dengan guard keamanan.”

---

### Task 8: Implement Permission management in `/settings/rbac`

**Objective:** Permission CRUD/status tetap tersedia di halaman baru.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/rbac/+page.svelte`

**Behavior:**

- `+ Permission` calls `createRBACPermission`
- `Edit` calls `updateRBACPermission`
- status toggle calls `setRBACPermissionActive`
- critical permission warning on status toggle

**Important:** Backend remains source of truth for critical permission rejection.

---

### Task 9: Implement role-permission matrix UI

**Objective:** Admin bisa memilih role lalu centang permission per module.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/rbac/+page.svelte`
- Better create: `apps/web-admin/src/routes/settings/rbac/_components/RolePermissionMatrix.svelte`

**Component props:**

```ts
export type RolePermissionMatrixProps = {
	roles: RBACRole[];
	permissions: RBACPermission[];
	rolePermissions: Record<string, string[]>;
	onSave: (roleCode: string, permissions: string[]) => Promise<void>;
};
```

**UI state:**

- selectedRoleCode
- draftPermissions Set<string>
- search text
- module filter
- dirty boolean from diff

**Behavior:**

- selecting role loads current permissions into draft
- checking/unchecking updates draft only
- `Reset Perubahan` restores current
- `Simpan Permission Role` opens confirmation dialog before API call

---

### Task 10: Add confirmation/diff preview before saving role permissions

**Objective:** Admin memahami dampak sebelum simpan.

**Files:**

- Modify/create: `RolePermissionMatrix.svelte`
- Use existing `confirmAction` if sufficient, or add a richer local dialog component.

**Preview content:**

- Role target
- Permission ditambah
- Permission dicabut
- Critical permission warning if add/remove includes:
  - `roles.manage`
  - `users.manage_roles`
- Session invalidation warning:
  - “User dengan role ini akan diminta login ulang karena token/session dicabut oleh backend.”

**If user count is not available yet:** show generic warning. Do not invent counts.

---

### Task 11: Save role permissions and refresh matrix

**Objective:** Matrix benar-benar menyimpan perubahan via backend and reload state.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/rbac/+page.svelte`
- Modify: `RolePermissionMatrix.svelte`

**Flow:**

1. call `updateRBACRolePermissions(roleCode, Array.from(draft).sort())`
2. toast success
3. refresh `fetchRBACMatrix()`
4. update role-permission map
5. clear dirty state

**Error handling:**

- show backend message via existing `overviewErrorMessage` style
- for rejected critical admin changes, show backend rejection message

---

### Task 12: Clean `/settings/users` RBAC section duplication

**Objective:** User management page tidak terlalu padat.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/users/+page.svelte`

**Recommended change:**

- Keep only user-account operations and user-role toggles.
- Replace large “Manajemen Role & Permission Dinamis” panel with callout/card:
  - title: “Manajemen RBAC”
  - description: “Kelola role, permission, dan matrix akses di halaman khusus.”
  - button: `/settings/rbac`

**Do not remove user role assignment from user table.** Assigning role to user belongs in `/settings/users`.

---

### Task 13: Add frontend tests for matrix helpers

**Objective:** Data transform/diff/critical helpers locked by tests.

**Files:**

- Create/Modify: `apps/web-admin/src/lib/rbac/matrix.test.ts`

**Run:**

```bash
cd apps/web-admin
npm run test:unit -- src/lib/rbac/matrix.test.ts
```

---

### Task 14: Add frontend tests for route/sidebar RBAC access

**Objective:** `/settings/rbac` access cannot regress.

**Files:**

- Modify: `apps/web-admin/src/lib/server/route-access.test.ts`
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`

**Cases:**

- no auth -> protected route behavior existing hooks should handle
- plain guru denied
- user with `roles.read` allowed
- sidebar item visible with `roles.read`
- mutating UI remains backend-protected by `roles.manage`

---

### Task 15: Add/extend BFF route tests

**Objective:** New role-permission route returns protected path behavior and proxies correct backend URL.

**Files:**

- Existing route proxy test location, or create near route if project convention allows.

**Cases:**

- `PUT /api/rbac/roles/guru/permissions` proxies to `/api/rbac/roles/guru/permissions`
- payload `{ permissions: [...] }` preserved
- route unauth smoke expected `401`, not `404`

---

### Task 16: Backend review — validate no service patch needed

**Objective:** Confirm backend already covers required invariants; only patch if gap found.

**Files to inspect:**

- `services/core-api/internal/service/rbac.go`
- `services/core-api/internal/handler/rbac.go`
- `services/core-api/internal/service/rbac_test.go`
- `services/core-api/cmd/api/routes_contract_test.go`

**Expected existing guard:**

- `ReplaceRolePermissions` validates role active
- validates permission active
- prevents last admin losing critical admin permissions
- invalidates affected users
- audits `RBAC_ROLE_PERMISSIONS_UPDATED`

**Patch only if needed:**

- Add test proving non-admin custom role permission update invalidates users.
- Add handler test for `roles.manage` required if current coverage missing.
- Consider zero-permission role policy. Current backend rejects empty permissions. Keep as-is unless user wants “role kosong”.

---

### Task 17: Add operational copy and docs

**Objective:** Admin memahami RBAC behavior.

**Files:**

- Create: `docs/rbac-management.md` or update existing docs if present.
- Optional: add inline help panel in `/settings/rbac`.

**Docs content:**

- difference between role metadata, role-permission mapping, user-role assignment
- why role system info/status locked
- session invalidation after RBAC changes
- critical permissions
- recommended role patterns:
  - `operator_bank_soal`
  - `reviewer_bank_soal`
  - `operator_asesmen`

---

### Task 18: Full validation

**Objective:** Semua quality gates lulus sebelum deploy.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app

git diff --check

cd services/core-api
GOCACHE=/tmp/go-cache go test ./...

cd ../../apps/web-admin
npm run test:unit -- \
  src/lib/rbac/matrix.test.ts \
  src/lib/client/rbac-users.test.ts \
  src/lib/server/route-access.test.ts \
  src/lib/components/sidebar/sidebar-config.test.ts
npm run check
npm run build

cd ../..
make build-backend build-web build-worker
```

Expected:

- all pass
- no warnings/errors in `npm run check`

---

### Task 19: Runtime deploy and smoke

**Objective:** Pastikan route baru bekerja di runtime.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
make db-migrate
pm2 restart deploy/pm2/backend.config.cjs
pm2 restart deploy/pm2/web.config.cjs
pm2 restart deploy/pm2/worker.config.cjs
pm2 status
```

**Smoke unauth expected:**

```bash
curl -sS -o /tmp/rbac-matrix.out -w '%{http_code}\n' http://127.0.0.1:8021/api/rbac/matrix
# expected 401

curl -sS -o /tmp/rbac-role-perms.out -w '%{http_code}\n' \
  -X PUT -H 'Content-Type: application/json' \
  --data '{"permissions":["roles.read"]}' \
  http://127.0.0.1:8021/api/rbac/roles/guru/permissions
# expected 401, not 404

curl -sS -o /tmp/web.out -w '%{http_code}\n' http://127.0.0.1:8021/
# expected 200

curl -sS -o /tmp/health.out -w '%{http_code}\n' http://127.0.0.1:8080/health
# expected 200
```

**Authenticated smoke if session tooling available:**

- login as admin
- open `/settings/rbac`
- select role `guru`
- verify permissions display
- try save no-op or safe test custom role if available
- verify toast success and matrix refresh

---

### Task 20: Final review and commit

**Objective:** Clean release-ready commit.

**Steps:**

```bash
git status --short
git diff --stat
git diff --check
```

Run static scan on diff for:

- secrets
- connection string
- token/password leaks
- unsafe eval
- suspicious shell injection

Commit:

```bash
git add apps/web-admin services/core-api docs .hermes/plans
# Include only intended files.
git commit -m "feat: add full rbac role permission editor"
git status --short
```

Expected final:

- worktree clean
- final report includes commit hash

---

## 6. Files Likely to Change

### Frontend/BFF

- `apps/web-admin/src/routes/api/rbac/roles/[code]/permissions/+server.ts` — new BFF proxy
- `apps/web-admin/src/lib/client/rbac-users.ts` — helper/type normalization
- `apps/web-admin/src/lib/client/rbac-users.test.ts` — client/helper tests
- `apps/web-admin/src/lib/rbac/matrix.ts` — new helper module
- `apps/web-admin/src/lib/rbac/matrix.test.ts` — helper tests
- `apps/web-admin/src/routes/settings/rbac/+page.svelte` — new RBAC management page
- `apps/web-admin/src/routes/settings/rbac/_components/RolePermissionMatrix.svelte` — new component
- `apps/web-admin/src/routes/settings/users/+page.svelte` — remove/replace bulky RBAC panel with link/callout
- `apps/web-admin/src/lib/server/route-access.ts` — protect `/settings/rbac`
- `apps/web-admin/src/lib/server/route-access.test.ts` — route guard regression
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts` — add RBAC nav entry
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts` — sidebar regression

### Backend

Likely no production backend changes needed, but inspect/test:

- `services/core-api/internal/service/rbac.go`
- `services/core-api/internal/handler/rbac.go`
- `services/core-api/internal/service/rbac_test.go`
- `services/core-api/cmd/api/main.go`
- `services/core-api/cmd/api/routes_contract_test.go`

### Docs

- `.hermes/plans/2026-05-07_184650-full-rbac-management-editor.md`
- Optional: `docs/rbac-management.md`

---

## 7. Security Considerations

- Keep backend as source of truth; frontend disables are UX only.
- All RBAC mutation endpoints must require `roles.manage` or admin fallback as currently established.
- Viewing matrix can use `roles.read`.
- Do not allow self-lockout by frontend workaround; rely on backend guard.
- Critical permissions require warning and backend guard:
  - `roles.manage`
  - `users.manage_roles`
- Role permission changes should invalidate affected sessions; backend currently does this in `ReplaceRolePermissions`.
- Avoid logging request bodies containing credentials. RBAC payloads are permission codes only, but keep general hygiene.

---

## 8. Open Questions

1. Apakah role custom boleh memiliki **zero permissions**?
   - Current backend menolak empty list.
   - Rekomendasi: tetap tolak untuk mencegah role “kosong” membingungkan.

2. Apakah role sistem boleh diubah nama/deskripsi?
   - Current backend menolak.
   - Rekomendasi: tetap kunci metadata role sistem.

3. Apakah `/settings/rbac` perlu masuk sidebar sebagai item baru atau cukup tombol dari `/settings/users`?
   - Rekomendasi: masuk sidebar di grup Pengaturan untuk admin/RBAC manager.

4. Apakah perlu preview jumlah user terdampak secara akurat?
   - MVP: generic warning.
   - Better: backend endpoint tambahan atau matrix `user_roles` dihitung di frontend.

---

## 9. Suggested Commit Strategy

- Commit 1: BFF + client helper + matrix helper tests
- Commit 2: route/sidebar guard for `/settings/rbac`
- Commit 3: RBAC page shell + role/permission management
- Commit 4: Role permission matrix + confirmation dialog
- Commit 5: `/settings/users` cleanup + docs
- Commit 6: validation/deploy fixes if any

If user wants one final commit only, squash into:

```bash
git commit -m "feat: add full rbac role permission editor"
```

---

## 10. Final Definition of Done

- [ ] `/settings/rbac` exists and protected.
- [ ] Sidebar exposes RBAC only to authorized users.
- [ ] Role info editing still works for custom roles.
- [ ] Role system metadata/status remains protected.
- [ ] Permission CRUD/status still works.
- [ ] Role-permission matrix works via BFF to backend.
- [ ] Confirmation diff shown before saving.
- [ ] Session invalidation warning shown.
- [ ] Tests pass.
- [ ] Build passes.
- [ ] Runtime smoke passes.
- [ ] Commit created.
- [ ] Worktree clean.
