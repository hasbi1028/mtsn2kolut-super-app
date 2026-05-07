# Student & Parent Account Portal Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Membangun akun dan portal terbatas untuk `siswa` dan `orang tua/wali` secara aman, terhubung ke master data siswa/orang tua, serta konsisten dengan RBAC dinamis yang sudah ada.

**Architecture:** Gunakan model `users` yang sudah mendukung `student_id` dan `parent_id`, role RBAC sistem yang sudah ada (`siswa`, `ortu`), serta relasi `parents` + `parent_students` untuk membatasi akses orang tua hanya ke anaknya. Semua akses data tetap lewat Go Core API + SvelteKit BFF; frontend tidak boleh akses PostgreSQL langsung. Implementasi dilakukan bertahap: foundation akun, generator admin/operator, portal read-only, lalu self-service terbatas + approval.

**Tech Stack:** Go 1.26 + chi + sqlc + PostgreSQL, SvelteKit 2/Svelte 5 + BFF routes, RBAC dynamic permissions, PM2 deployment.

---

## 1. Current Context / Assumptions

### Existing foundation yang sudah tersedia

- Tabel `users` sudah punya link profil:
  - `employee_id`
  - `student_id`
  - `parent_id`
- `UserLifecycle.UpdateProfileLink()` sudah menjaga satu user hanya boleh terkait ke satu profile type.
- Role RBAC sistem sudah tersedia di migration `069_dynamic_rbac_foundation.sql`:
  - `siswa` = Siswa
  - `ortu` = Orang Tua/Wali
- Permission dasar untuk `siswa` dan `ortu` saat ini masih minimal:
  - `dashboard.read`
  - `notifications.read`
  - `settings.account`
- Master data siswa sudah ada di `students`.
- Master data orang tua sudah ada di `parents`.
- Relasi orang tua-anak sudah ada di `parent_students` dengan metadata:
  - `relationship`
  - `is_primary_contact`
  - `notes`
- Halaman admin yang relevan sudah ada:
  - `/students`
  - `/parents`
  - `/settings/users`
  - `/settings/rbac`
- Self-service akun internal sudah ada di `/settings/account` dan sudah mendukung profile type student/parent untuk contact/avatar path di query `users.sql`.

### Key decision

- Siswa dan orang tua **dibuatkan akun**, tetapi akses awal **read-only/terbatas**.
- Data resmi tidak boleh diedit langsung oleh siswa/orang tua. Perubahan data resmi harus lewat change request + approval.
- Untuk tahap awal, pembuatan akun dilakukan oleh admin/operator dari web-admin, bukan registrasi publik mandiri.
- Password awal harus sementara dan user wajib ganti password pada login pertama.

---

## 2. Target User Experience

### 2.1 Admin/operator

Admin/operator dapat:

1. Melihat status akun pada daftar siswa:
   - belum punya akun
   - akun aktif
   - akun nonaktif
   - terakhir login
2. Membuat akun siswa per siswa.
3. Membuat akun siswa massal per rombel/angkatan.
4. Melihat status akun pada daftar orang tua/wali.
5. Membuat akun orang tua/wali per parent record.
6. Membuat akun orang tua/wali massal berdasarkan relasi `parent_students`.
7. Reset password siswa/orang tua.
8. Nonaktifkan/aktifkan akun siswa/orang tua.
9. Melihat hasil generate akun tanpa menampilkan password lama/rahasia setelah layar ditutup.

### 2.2 Siswa

Siswa dapat login dan hanya melihat data miliknya sendiri:

- Dashboard siswa.
- Profil saya.
- Jadwal saya.
- Rombel/kelas saya.
- Hasil asesmen/nilai yang sudah dirilis.
- Notifikasi.
- Pengaturan akun sendiri.
- Change request untuk data resmi tertentu.

### 2.3 Orang tua/wali

Orang tua dapat login dan hanya melihat data anak yang terhubung lewat `parent_students`:

- Dashboard orang tua.
- Daftar anak.
- Profil anak.
- Jadwal anak.
- Kehadiran/presensi anak jika data tersedia.
- Nilai/hasil asesmen yang sudah dirilis.
- Notifikasi.
- Pengaturan akun sendiri.
- Change request data anak/orang tua tertentu dengan approval.

---

## 3. Security / RBAC Principles

1. **Fail closed:** jika `student_id`/`parent_id` tidak ada di JWT/account summary, portal siswa/orang tua harus menolak akses.
2. **No broad student listing for siswa/ortu:** siswa tidak boleh hit endpoint daftar semua siswa; orang tua tidak boleh daftar semua siswa.
3. **Parent child scope:** semua endpoint orang tua wajib filter lewat `parent_students.parent_id = current user's parent_id`.
4. **Student self scope:** semua endpoint siswa wajib filter `students.id = current user's student_id`.
5. **No direct official data mutation:** siswa/orang tua hanya boleh request perubahan, bukan update langsung kolom resmi.
6. **Session invalidation:** reset password, role change, deactivate account harus increment `auth_version` dan revoke sessions seperti pola user lifecycle existing.
7. **Audit log:** create/generate/reset/deactivate akun siswa/orang tua wajib mencatat audit.
8. **No plaintext password persistence:** password sementara hanya dikembalikan pada response generate, tidak disimpan plaintext.
9. **Compatibility-safe RBAC:** seed permission baru secara additive; jangan menghapus role/permission existing.

---

## 4. Proposed Permission Model

### 4.1 New permissions to seed

Tambahkan migration baru, misalnya `081_student_parent_account_portal.sql` atau nomor berikutnya sesuai repo state saat implementasi.

Permission siswa:

- `student_portal.read` — akses dashboard/portal siswa.
- `student_portal.profile_read` — melihat profil sendiri.
- `student_portal.schedule_read` — melihat jadwal sendiri.
- `student_portal.grades_read` — melihat nilai/hasil sendiri yang dirilis.
- `student_portal.assessment_take` — mengikuti asesmen/ujian sebagai siswa.
- `student_portal.profile_change_request` — mengajukan perubahan data resmi sendiri.

Permission orang tua:

- `parent_portal.read` — akses dashboard/portal orang tua.
- `parent_portal.children_read` — melihat daftar anak terhubung.
- `parent_portal.child_profile_read` — melihat profil anak.
- `parent_portal.child_schedule_read` — melihat jadwal anak.
- `parent_portal.child_attendance_read` — melihat kehadiran anak.
- `parent_portal.child_grades_read` — melihat nilai/hasil anak yang dirilis.
- `parent_portal.profile_change_request` — mengajukan perubahan data orang tua/anak.

Permission admin/operator untuk akun eksternal:

- `student_accounts.manage` — generate/reset/deactivate akun siswa.
- `parent_accounts.manage` — generate/reset/deactivate akun orang tua.

### 4.2 Role defaults

Grant default:

- Role `siswa`:
  - `dashboard.read`
  - `notifications.read`
  - `settings.account`
  - `student_portal.read`
  - `student_portal.profile_read`
  - `student_portal.schedule_read`
  - `student_portal.grades_read`
  - `student_portal.assessment_take`
  - `student_portal.profile_change_request`

- Role `ortu`:
  - `dashboard.read`
  - `notifications.read`
  - `settings.account`
  - `parent_portal.read`
  - `parent_portal.children_read`
  - `parent_portal.child_profile_read`
  - `parent_portal.child_schedule_read`
  - `parent_portal.child_attendance_read`
  - `parent_portal.child_grades_read`
  - `parent_portal.profile_change_request`

- Role `admin`:
  - semua permission baru.

- Role `kesiswaan` atau operator custom:
  - `students.read`
  - `parents.read`
  - `student_accounts.manage`
  - `parent_accounts.manage`

Catatan: jika belum ada role `operator`, jangan tambah role baru dulu kecuali user minta. Gunakan RBAC Manager untuk assign permission ke custom role nanti.

---

## 5. Username & Password Policy

### 5.1 Siswa

Rekomendasi username default:

```text
<nisn>
```

Fallback jika NISN kosong/duplikat:

```text
s<NISM/NIS Lokal>
```

Jika masih collision:

```text
s<NISM/NIS Lokal>-<2 digit sequence>
```

### 5.2 Orang tua/wali

Rekomendasi username default:

```text
ortu<NISN anak utama>
```

Jika satu parent memiliki banyak anak, pakai anak pertama/primary contact sebagai basis.

Fallback jika NISN anak kosong:

```text
ortu<last8digit-phone>
```

Jika collision:

```text
ortu<last8digit-phone>-<2 digit sequence>
```

### 5.3 Password awal

Pilihan aman bertahap:

- Tahap awal: password sementara random, misalnya 10-12 karakter, hanya ditampilkan sekali pada hasil generate.
- Tambahkan flag wajib ganti password pertama kali.

Jika schema belum punya flag wajib ganti password, tambahkan kolom:

```sql
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;
```

Kemudian:

- Generate akun: `must_change_password = TRUE`.
- Reset password admin: `must_change_password = TRUE`.
- User sukses ganti password sendiri: `must_change_password = FALSE`, `password_changed_at = NOW()`.

---

## 6. Phased Implementation Roadmap

## Phase 0 — Preflight Audit & Final Schema Decision

**Objective:** Memastikan kondisi repo/live schema sebelum implementasi.

**Files to inspect:**

- `services/core-api/db/migrations/069_dynamic_rbac_foundation.sql`
- `services/core-api/db/migrations/071_user_account_metadata_soft_delete.sql`
- `services/core-api/db/migrations/078_rombels_parent_relationships.sql`
- `services/core-api/db/queries/users.sql`
- `services/core-api/db/queries/students.sql`
- `services/core-api/db/queries/parents.sql`
- `services/core-api/internal/service/user_lifecycle.go`
- `services/core-api/internal/service/user_generation.go`
- `services/core-api/internal/handler/user.go`
- `apps/web-admin/src/routes/students/+page.svelte`
- `apps/web-admin/src/routes/parents/+page.svelte`
- `apps/web-admin/src/routes/settings/users/+page.svelte`
- `apps/web-admin/src/lib/server/route-access.ts`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

**Commands:**

```bash
git status --short | cat
git log -1 --oneline
cd services/core-api && go test ./internal/service -run 'TestUser|TestRBAC' -count=1
cd apps/web-admin && npm run test:unit -- src/lib/components/sidebar/sidebar-config.test.ts src/lib/server/route-access.test.ts
```

**Acceptance criteria:**

- Worktree scoped before editing.
- Existing user lifecycle tests still pass.
- Current RBAC seed state known.
- Decide next migration number from actual `services/core-api/db/migrations` list.

---

## Phase 1 — RBAC & Account Metadata Foundation

**Objective:** Add permissions and mandatory password-change metadata.

**Files:**

- Create: `services/core-api/db/migrations/NNN_student_parent_account_portal.sql`
- Modify: `services/core-api/db/queries/users.sql`
- Modify generated sqlc output after `sqlc generate`.
- Modify: `services/core-api/internal/service/auth.go` if login/change-password needs `must_change_password` claim/response.
- Modify: `services/core-api/internal/handler/auth.go` for response field if needed.

**Migration content outline:**

```sql
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('student_portal.read', 'student_portal', 'read', 'Mengakses portal siswa.'),
    ('student_portal.profile_read', 'student_portal', 'profile_read', 'Melihat profil siswa sendiri.'),
    ('student_portal.schedule_read', 'student_portal', 'schedule_read', 'Melihat jadwal siswa sendiri.'),
    ('student_portal.grades_read', 'student_portal', 'grades_read', 'Melihat hasil/nilai siswa sendiri yang dirilis.'),
    ('student_portal.assessment_take', 'student_portal', 'assessment_take', 'Mengikuti asesmen sebagai siswa.'),
    ('student_portal.profile_change_request', 'student_portal', 'profile_change_request', 'Mengajukan perubahan data siswa sendiri.'),
    ('parent_portal.read', 'parent_portal', 'read', 'Mengakses portal orang tua/wali.'),
    ('parent_portal.children_read', 'parent_portal', 'children_read', 'Melihat daftar anak terhubung.'),
    ('parent_portal.child_profile_read', 'parent_portal', 'child_profile_read', 'Melihat profil anak.'),
    ('parent_portal.child_schedule_read', 'parent_portal', 'child_schedule_read', 'Melihat jadwal anak.'),
    ('parent_portal.child_attendance_read', 'parent_portal', 'child_attendance_read', 'Melihat kehadiran anak.'),
    ('parent_portal.child_grades_read', 'parent_portal', 'child_grades_read', 'Melihat hasil/nilai anak yang dirilis.'),
    ('parent_portal.profile_change_request', 'parent_portal', 'profile_change_request', 'Mengajukan perubahan data orang tua/anak.'),
    ('student_accounts.manage', 'student_accounts', 'manage', 'Membuat/reset/nonaktifkan akun siswa.'),
    ('parent_accounts.manage', 'parent_accounts', 'manage', 'Membuat/reset/nonaktifkan akun orang tua/wali.')
)
INSERT INTO rbac_permissions (code, module, action, description)
SELECT code, module, action, description FROM permission_seed
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();
```

**Tests:**

- Add service/handler auth tests proving `must_change_password` is returned/handled if implemented in auth contract.
- Add RBAC readiness/permission catalog test for new permissions if existing tests lock catalog behavior.

**Validation:**

```bash
cd services/core-api
sqlc generate -f db/sqlc.yaml
go test ./internal/service -run 'TestAuth|TestUser|TestRBAC' -count=1
go test ./...
```

**Commit:**

```bash
git add services/core-api/db/migrations/NNN_student_parent_account_portal.sql services/core-api/db/queries/users.sql services/core-api/internal/repository/postgres services/core-api/internal/service services/core-api/internal/handler
git commit -m "feat: add student parent account portal rbac foundation"
```

---

## Phase 2 — Backend Account Generation Services

**Objective:** Add reusable account generators for siswa and orang tua/wali.

**Files:**

- Create: `services/core-api/internal/service/student_account_generation.go`
- Create: `services/core-api/internal/service/student_account_generation_test.go`
- Create: `services/core-api/internal/service/parent_account_generation.go`
- Create: `services/core-api/internal/service/parent_account_generation_test.go`
- Modify: `services/core-api/db/queries/users.sql`
- Modify generated sqlc output.

**Queries to add:**

- `ListStudentAccountGenerationCandidates`
  - candidate if `students.is_active = TRUE`
  - include existing linked user if any.
  - include username collision info.
- `ListParentAccountGenerationCandidates`
  - candidate if parent has at least one `parent_students` link.
  - include primary/first child NISN for username basis.
  - include existing linked user if any.
- `CreateUserWithMustChangePassword` or extend `CreateUser` to include `must_change_password`.

**Service behavior:**

- Preview returns:
  - total
  - ready
  - skipped
  - candidates[]
  - generated username
  - reason if skipped
- Generate:
  - transaction per batch or one transaction with item-level failure handling.
  - create `users` row linked to `student_id` or `parent_id`.
  - add compatibility role in `user_account_roles` if enum supports role; if enum does not include `ortu`, verify current enum values before using.
  - add dynamic role in `rbac_user_roles` by role code `siswa`/`ortu`.
  - hash random password.
  - set `must_change_password = TRUE`.
  - audit action:
    - `STUDENT_ACCOUNT_GENERATED`
    - `PARENT_ACCOUNT_GENERATED`

**Important design note:**

If legacy `user_role` enum/table `user_account_roles` does not support `ortu`, do not force legacy compatibility. Dynamic RBAC should be source of truth; test login/token permission claims from `rbac_user_roles`.

**Tests:**

- Candidate with NISN gets username NISN.
- Student with existing user is skipped.
- Username collision gets deterministic suffix.
- Parent with multiple children uses primary contact/first child deterministic basis.
- Parent without child link is skipped.
- Generate assigns correct dynamic RBAC role.
- Password is returned only in result item and not stored plaintext.
- Audit is written.

**Validation:**

```bash
cd services/core-api
sqlc generate -f db/sqlc.yaml
gofmt -w internal/service/student_account_generation.go internal/service/student_account_generation_test.go internal/service/parent_account_generation.go internal/service/parent_account_generation_test.go
go test ./internal/service -run 'TestStudentAccountGeneration|TestParentAccountGeneration' -count=1
go test ./...
```

**Commit:**

```bash
git add services/core-api/db/queries/users.sql services/core-api/internal/repository/postgres services/core-api/internal/service
git commit -m "feat: add student and parent account generators"
```

---

## Phase 3 — Backend Admin API Endpoints

**Objective:** Expose protected endpoints for preview/generate/reset/status management.

**Files:**

- Modify: `services/core-api/internal/handler/user.go` or create `services/core-api/internal/handler/account_generation.go`
- Modify: `services/core-api/cmd/api/main.go`
- Add tests: `services/core-api/internal/handler/user_success_test.go` or new handler test file.

**Routes:**

```text
GET  /api/users/student-accounts/preview
POST /api/users/student-accounts/generate
GET  /api/users/parent-accounts/preview
POST /api/users/parent-accounts/generate
POST /api/users/{id}/force-password-change
```

Alternative scoped routes if preferred:

```text
POST /api/students/{id}/account
POST /api/parents/{id}/account
```

Recommended initial endpoints:

- Batch preview/generate first, because existing employee generator follows this pattern.
- Per-record account buttons can call generate with filters later.

**Authorization:**

- Student account endpoints: require `student_accounts.manage` or admin fallback.
- Parent account endpoints: require `parent_accounts.manage` or admin fallback.
- Do not rely only on `adminAccessAllowed`; use dynamic permission middleware where available.

**Tests:**

- Unauthenticated returns `401`.
- Authenticated without permission returns `403`.
- Admin/permissioned actor can preview/generate.
- Response does not expose existing password hashes.

**Validation:**

```bash
cd services/core-api
go test ./internal/handler -run 'Test.*Account.*Generation|TestUser' -count=1
go test ./...
```

**Commit:**

```bash
git add services/core-api/internal/handler services/core-api/cmd/api/main.go
git commit -m "feat: expose student parent account generation endpoints"
```

---

## Phase 4 — SvelteKit BFF Proxies & Client Helpers

**Objective:** Add browser-consumed BFF routes and typed client helpers.

**Files:**

- Create: `apps/web-admin/src/routes/api/users/student-accounts/preview/+server.ts`
- Create: `apps/web-admin/src/routes/api/users/student-accounts/generate/+server.ts`
- Create: `apps/web-admin/src/routes/api/users/parent-accounts/preview/+server.ts`
- Create: `apps/web-admin/src/routes/api/users/parent-accounts/generate/+server.ts`
- Create/modify: `apps/web-admin/src/lib/client/account-generation.ts`
- Modify tests: `apps/web-admin/src/lib/server/proxy-routes.test.ts`

**Client functions:**

```ts
export async function previewStudentAccounts(fetcher = fetch) { ... }
export async function generateStudentAccounts(fetcher = fetch) { ... }
export async function previewParentAccounts(fetcher = fetch) { ... }
export async function generateParentAccounts(fetcher = fetch) { ... }
```

**Tests:**

- Proxy route maps to correct backend path.
- Client helpers reject non-OK response with useful message.
- Client helpers parse result shape.

**Validation:**

```bash
cd apps/web-admin
npm run test:unit -- src/lib/server/proxy-routes.test.ts src/lib/client/account-generation.test.ts
npm run check
```

**Commit:**

```bash
git add apps/web-admin/src/routes/api/users apps/web-admin/src/lib/client apps/web-admin/src/lib/server/proxy-routes.test.ts
git commit -m "feat: add student parent account bff routes"
```

---

## Phase 5 — Admin UI: Account Generation in Students/Parents

**Objective:** Add admin/operator UX to create accounts from `/students` and `/parents`.

**Files:**

- Modify: `apps/web-admin/src/routes/students/+page.svelte`
- Modify: `apps/web-admin/src/routes/parents/+page.svelte`
- Possibly modify: `apps/web-admin/src/lib/client/students.ts`
- Possibly modify: `apps/web-admin/src/lib/client/parents.ts`
- Add tests where patterns exist or use raw Svelte source tests:
  - `apps/web-admin/src/routes/students/student-account-actions.test.ts`
  - `apps/web-admin/src/routes/parents/parent-account-actions.test.ts`

**UI elements:**

Students page:

- Card: “Akun Siswa”
- Buttons:
  - “Preview akun siswa”
  - “Generate akun siswa”
- Table columns/indicators:
  - Status akun
  - Username
  - Role

Parents page:

- Card: “Akun Orang Tua/Wali”
- Buttons:
  - “Preview akun orang tua”
  - “Generate akun orang tua”
- Table columns/indicators:
  - Status akun
  - Username
  - Anak terhubung

**Safety copy:**

- “Password hanya tampil sekali. Simpan/unduh hasil generate sekarang.”
- “Akun wajib mengganti password saat login pertama.”
- “Data resmi tetap dikunci dan perubahan melalui approval.”

**Tests:**

- Students page has account-generation CTA and safety copy.
- Parents page has account-generation CTA and safety copy.
- UI hides/disabled actions without permissions if permission-aware store is available.

**Validation:**

```bash
cd apps/web-admin
npm run test:unit -- src/routes/students/student-account-actions.test.ts src/routes/parents/parent-account-actions.test.ts
npm run check
npm run build
```

**Commit:**

```bash
git add apps/web-admin/src/routes/students apps/web-admin/src/routes/parents apps/web-admin/src/lib/client
git commit -m "feat: add student parent account generation ui"
```

---

## Phase 6 — Portal Routing & Access Guards

**Objective:** Add dedicated portal routes with correct role/permission guards.

**Files:**

- Modify: `apps/web-admin/src/lib/server/route-access.ts`
- Modify: `apps/web-admin/src/lib/server/route-access.test.ts`
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`
- Create: `apps/web-admin/src/routes/portal/siswa/+page.svelte`
- Create: `apps/web-admin/src/routes/portal/orang-tua/+page.svelte`

**Routes:**

```text
/portal/siswa
/portal/orang-tua
```

**Access:**

- `/portal/siswa` requires role `siswa` and/or permission `student_portal.read`.
- `/portal/orang-tua` requires role `ortu` and/or permission `parent_portal.read`.

**Sidebar:**

- For `siswa`: “Portal Siswa” with icon `user-check` or `book-open`.
- For `ortu`: “Portal Orang Tua” with icon `user-group`.
- Keep admin/guru sidebar unaffected.

**Tests:**

- `/portal/siswa` unauth redirects/401 per existing route guard pattern.
- siswa role can access `/portal/siswa`.
- ortu role cannot access `/portal/siswa` unless admin.
- ortu role can access `/portal/orang-tua`.
- sidebar entries are visible only to matching roles/permissions.

**Validation:**

```bash
cd apps/web-admin
npm run test:unit -- src/lib/server/route-access.test.ts src/lib/components/sidebar/sidebar-config.test.ts
npm run check
npm run build
```

**Commit:**

```bash
git add apps/web-admin/src/lib/server/route-access.ts apps/web-admin/src/lib/server/route-access.test.ts apps/web-admin/src/lib/components/sidebar apps/web-admin/src/routes/portal
git commit -m "feat: add student and parent portal routes"
```

---

## Phase 7 — Student Self API & Portal Data

**Objective:** Add safe self-scoped API for siswa.

**Files:**

Backend:

- Create: `services/core-api/internal/handler/student_portal.go`
- Create: `services/core-api/internal/service/student_portal.go`
- Create tests:
  - `services/core-api/internal/service/student_portal_test.go`
  - `services/core-api/internal/handler/student_portal_test.go`
- Modify: `services/core-api/db/queries/students.sql`
- Modify: `services/core-api/cmd/api/main.go`

Frontend/BFF:

- Create: `apps/web-admin/src/routes/api/portal/siswa/profile/+server.ts`
- Create: `apps/web-admin/src/routes/api/portal/siswa/schedule/+server.ts`
- Create: `apps/web-admin/src/routes/api/portal/siswa/results/+server.ts`
- Create: `apps/web-admin/src/lib/client/student-portal.ts`
- Modify: `apps/web-admin/src/routes/portal/siswa/+page.svelte`

**Backend routes:**

```text
GET /api/portal/student/profile
GET /api/portal/student/schedule
GET /api/portal/student/results
```

**Scope rule:**

- Derive `student_id` from authenticated user/JWT/account row.
- Never accept `student_id` from query/body for self routes.

**Tests:**

- User without `student_id` gets `403`.
- siswa gets only own profile.
- attempting to pass another student id is ignored/rejected.
- no endpoint returns all students.

**Validation:**

```bash
cd services/core-api
go test ./internal/service -run TestStudentPortal -count=1
go test ./internal/handler -run TestStudentPortal -count=1
go test ./...
cd apps/web-admin
npm run test:unit -- src/lib/client/student-portal.test.ts
npm run check
npm run build
```

**Commit:**

```bash
git add services/core-api apps/web-admin/src/routes/api/portal/siswa apps/web-admin/src/lib/client/student-portal.ts apps/web-admin/src/routes/portal/siswa/+page.svelte
git commit -m "feat: add student self portal data"
```

---

## Phase 8 — Parent Scoped API & Portal Data

**Objective:** Add safe parent-scoped API over `parent_students`.

**Files:**

Backend:

- Create: `services/core-api/internal/handler/parent_portal.go`
- Create: `services/core-api/internal/service/parent_portal.go`
- Create tests:
  - `services/core-api/internal/service/parent_portal_test.go`
  - `services/core-api/internal/handler/parent_portal_test.go`
- Modify: `services/core-api/db/queries/parents.sql`
- Modify: `services/core-api/cmd/api/main.go`

Frontend/BFF:

- Create: `apps/web-admin/src/routes/api/portal/orang-tua/children/+server.ts`
- Create: `apps/web-admin/src/routes/api/portal/orang-tua/children/[id]/profile/+server.ts`
- Create: `apps/web-admin/src/routes/api/portal/orang-tua/children/[id]/schedule/+server.ts`
- Create: `apps/web-admin/src/routes/api/portal/orang-tua/children/[id]/results/+server.ts`
- Create: `apps/web-admin/src/lib/client/parent-portal.ts`
- Modify: `apps/web-admin/src/routes/portal/orang-tua/+page.svelte`

**Backend routes:**

```text
GET /api/portal/parent/children
GET /api/portal/parent/children/{student_id}/profile
GET /api/portal/parent/children/{student_id}/schedule
GET /api/portal/parent/children/{student_id}/results
```

**Scope rule:**

Every child-specific query must join/exists:

```sql
EXISTS (
  SELECT 1
  FROM parent_students ps
  WHERE ps.parent_id = current_user_parent_id
    AND ps.student_id = requested_student_id
)
```

**Tests:**

- Parent sees only linked children.
- Parent cannot access unlinked child by guessing UUID.
- Parent without `parent_id` gets `403`.
- Child schedule/results endpoints apply parent scope before returning data.

**Validation:**

```bash
cd services/core-api
go test ./internal/service -run TestParentPortal -count=1
go test ./internal/handler -run TestParentPortal -count=1
go test ./...
cd apps/web-admin
npm run test:unit -- src/lib/client/parent-portal.test.ts
npm run check
npm run build
```

**Commit:**

```bash
git add services/core-api apps/web-admin/src/routes/api/portal/orang-tua apps/web-admin/src/lib/client/parent-portal.ts apps/web-admin/src/routes/portal/orang-tua/+page.svelte
git commit -m "feat: add parent scoped portal data"
```

---

## Phase 9 — First Login Password Change Enforcement

**Objective:** Enforce `must_change_password` after generated/reset passwords.

**Files:**

- Modify: `services/core-api/internal/service/auth.go`
- Modify: `services/core-api/internal/handler/auth.go`
- Modify: `apps/web-admin/src/routes/+layout.svelte` or auth/session store if redirect logic is centralized.
- Create/modify route: `apps/web-admin/src/routes/settings/account/+page.svelte` or dedicated `/change-password` page.
- Add tests for auth and frontend route behavior.

**Behavior:**

- Login response/session includes `must_change_password`.
- User with `must_change_password = TRUE` can access only:
  - password change page
  - logout endpoint
- After successful password change:
  - set `must_change_password = FALSE`
  - set `password_changed_at = NOW()`
  - increment auth version or refresh session according to current auth design.

**Tests:**

- Generated student account has `must_change_password = TRUE`.
- Login with temp password returns the flag.
- Protected app routes redirect to password-change page until password changed.
- After change password, normal portal route is accessible.

**Validation:**

```bash
cd services/core-api
go test ./internal/service -run 'TestAuth|TestStudentAccountGeneration|TestParentAccountGeneration' -count=1
go test ./...
cd apps/web-admin
npm run check
npm run build
```

**Commit:**

```bash
git add services/core-api apps/web-admin/src
git commit -m "feat: enforce password change for generated accounts"
```

---

## Phase 10 — Official Data Change Requests for Student/Parent Scope

**Objective:** Extend existing approval workflow so siswa/orang tua can request official data changes.

**Files to inspect first:**

- `services/core-api/db/migrations/075_profile_change_requests.sql`
- `services/core-api/db/queries/profile_change_requests.sql`
- `services/core-api/internal/handler/auth.go`
- `services/core-api/internal/service/account.go` or related account self-service service files.
- `apps/web-admin/src/routes/settings/account/+page.svelte`
- `apps/web-admin/src/routes/settings/user-change-requests/+page.svelte`

**Policy:**

Siswa may request changes for:

- phone
- address/alamat
- photo/avatar
- optional: emergency contact fields if available later

Orang tua may request changes for:

- own phone/address/occupation/NIK if policy allows
- child data only for safe fields and with explicit reviewer visibility

Do not allow direct mutation of:

- NISN
- NISM/NIS lokal
- nama resmi
- tanggal lahir
- kelas/rombel
- parent-child relationship

unless routed through approval.

**Tests:**

- Siswa request is tied to own `student_id`.
- Orang tua child request requires `parent_students` relation.
- Reviewer sees context and can approve/reject.
- Approval writes audit log.

**Validation:**

```bash
cd services/core-api && go test ./...
cd apps/web-admin && npm run check && npm run build
```

**Commit:**

```bash
git add services/core-api apps/web-admin/src/routes/settings/account apps/web-admin/src/routes/settings/user-change-requests
git commit -m "feat: support student parent official data change requests"
```

---

## Phase 11 — Deployment, Smoke, and Readiness Gate

**Objective:** Deploy safely and add readiness documentation/tests.

**Files:**

- Create: `docs/student-parent-account-portal-readiness.md`
- Create: `apps/web-admin/src/lib/portal/student-parent-readiness.test.ts`

**Readiness doc must cover:**

- Role/permission contract.
- Account generation policy.
- Username/password policy.
- Student self-scope rules.
- Parent child-scope rules.
- Change request policy.
- Smoke checklist.
- Rollback plan.

**Deployment order:**

```bash
# Backend validation
cd services/core-api
go test ./...

# Frontend validation
cd ../../apps/web-admin
npm run test:unit
npm run check
npm run build

# Build/deploy from repo root
cd /home/servermtsn2kolut/mtsn2kolut-super-app
make build-backend build-web
make db-migrate
pm2 restart deploy/pm2/backend.config.cjs
pm2 restart deploy/pm2/web.config.cjs
pm2 status
```

**Smoke checks:**

Use production PM2 web port `8021`, not Vite preview `4173`.

```bash
curl -sS -o /dev/null -w 'core /health => %{http_code}\n' http://127.0.0.1:8080/health
curl -sS -o /dev/null -w 'web / => %{http_code}\n' http://127.0.0.1:8021/
curl -sS -o /dev/null -w 'web /portal/siswa unauth => %{http_code}\n' http://127.0.0.1:8021/portal/siswa
curl -sS -o /dev/null -w 'web /portal/orang-tua unauth => %{http_code}\n' http://127.0.0.1:8021/portal/orang-tua
curl -sS -o /dev/null -w 'bff student preview unauth => %{http_code}\n' http://127.0.0.1:8021/api/users/student-accounts/preview
curl -sS -o /dev/null -w 'bff parent preview unauth => %{http_code}\n' http://127.0.0.1:8021/api/users/parent-accounts/preview
```

Expected:

- `/health` = 200
- `/` = 200
- protected pages unauth = 302
- protected BFF/API unauth = 401

**Final commit:**

```bash
git add docs/student-parent-account-portal-readiness.md apps/web-admin/src/lib/portal/student-parent-readiness.test.ts
git commit -m "docs: add student parent portal readiness gate"
```

---

## 7. Suggested Commit Sequence

1. `feat: add student parent account portal rbac foundation`
2. `feat: add student and parent account generators`
3. `feat: expose student parent account generation endpoints`
4. `feat: add student parent account bff routes`
5. `feat: add student parent account generation ui`
6. `feat: add student and parent portal routes`
7. `feat: add student self portal data`
8. `feat: add parent scoped portal data`
9. `feat: enforce password change for generated accounts`
10. `feat: support student parent official data change requests`
11. `docs: add student parent portal readiness gate`

---

## 8. Risks & Mitigations

### Risk 1 — Parent can access wrong child

**Mitigation:** backend service tests must prove parent cannot access unlinked `student_id`. Never trust frontend filters.

### Risk 2 — Legacy role compatibility for `ortu`

**Mitigation:** inspect enum/table before using `user_account_roles`. If legacy enum lacks `ortu`, use dynamic `rbac_user_roles` as source of truth and test JWT roles/permissions.

### Risk 3 — Password temporary exposed too broadly

**Mitigation:** show password only in generate response; do not store plaintext; add UI warning and optional CSV export with deliberate confirmation.

### Risk 4 — Official data edited without approval

**Mitigation:** self routes only allow contact/avatar where already intended; official fields go through profile-change request workflow.

### Risk 5 — Portal UX grows too large

**Mitigation:** MVP portal is read-only and minimal: profile, schedule, released results. Attendance/advanced reports can be Phase 2 after foundation is stable.

### Risk 6 — Migration drift

**Mitigation:** before `make db-migrate`, inspect `schema_migrations` and reconcile gaps. Do not blindly insert migration metadata.

---

## 9. Open Questions for User Decision

These do not block Phase 0–3, but should be decided before broad rollout:

1. Username siswa final: pakai `NISN` atau `NPSN + tahun + nomor` seperti pegawai?
2. Password awal: random aman atau sama dengan username untuk kemudahan awal? Rekomendasi: random + wajib ganti.
3. Orang tua yang punya dua anak: cukup satu akun untuk semua anak? Rekomendasi: ya, satu akun parent untuk semua relasi `parent_students`.
4. Apakah siswa boleh melihat nilai langsung, atau hanya nilai yang sudah “dirilis”? Rekomendasi: hanya yang dirilis.
5. Apakah portal orang tua boleh melihat catatan pelanggaran/BK? Rekomendasi: nanti setelah policy internal jelas.
6. Apakah registrasi mandiri orang tua dibutuhkan sekarang? Rekomendasi: tunda sampai akun buatan admin stabil.

---

## 10. MVP Cut Recommendation

Jika ingin cepat rilis internal, ambil MVP berikut dulu:

1. Phase 1 — RBAC + must-change-password foundation.
2. Phase 2 — account generator service.
3. Phase 3 — admin API endpoints.
4. Phase 4 — BFF + client helpers.
5. Phase 5 — tombol generate di `/students` dan `/parents`.
6. Phase 9 — enforce password change.

Portal read-only siswa/orang tua bisa menyusul setelah akun berhasil dibuat dan login stabil.

**MVP outcome:** admin sudah bisa membuat akun siswa/orang tua dengan aman, akun punya role benar, password wajib diganti, dan belum ada risiko kebocoran data portal karena portal read-only belum dibuka luas.
