# Akademik Editable UI/UX Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Membuat modul Akademik MTsN 2 Kolaka Utara lebih mudah diedit, cepat dipakai operator, dan aman untuk data tahun ajaran, rombel, siswa, wali kelas, guru mapel, mapel, dan jadwal.

**Architecture:** Tetap mengikuti arsitektur monorepo: SvelteKit `apps/web-admin` sebagai UI/BFF proxy, Go `services/core-api` sebagai pemilik PostgreSQL, SQL melalui `sqlc`, perubahan schema melalui migration. Implementasi dibuat bertahap: foundation UX reusable, dashboard akademik, editable rombel/siswa, matrix guru mapel, lalu jadwal mingguan dan wizard tahun ajaran.

**Tech Stack:** SvelteKit 2 + Svelte 5 runes + Tailwind v4 + shadcn-svelte + sonner, Go 1.26 + Chi + sqlc + PostgreSQL.

---

## Product Decisions

### Target UX Level

Implement **Level 2 — Editable Cepat** lebih dulu:

- Inline edit pada field aman.
- Drawer/detail panel untuk edit kompleks.
- Bulk action untuk pekerjaan operator.
- Validasi sebelum simpan.
- Audit/preview perubahan untuk operasi besar.

Naikkan ke **Level 3** setelah fondasi stabil:

- Matrix editor penuh.
- Drag & drop jadwal.
- Import/export Excel dengan dry-run.
- Wizard tahun ajaran baru.

### Scope Menu Akademik Akhir

```text
Akademik
├─ Dashboard Akademik
├─ Tahun Ajaran & Semester
├─ Rombel
├─ Siswa
├─ Wali Kelas
├─ Mapel
├─ Guru Mapel
├─ Jadwal Pelajaran
├─ Kenaikan & Mutasi
└─ Pengaturan Akademik
```

### Guardrails

- Jangan akses PostgreSQL dari SvelteKit secara langsung.
- Semua mutasi lewat Go API.
- Data akademik operasional harus ada validasi backend.
- Bulk update wajib punya preview/dry-run jika berdampak banyak data.
- Perubahan jadwal/guru mapel/rombel harus bisa ditelusuri minimal dari audit log yang sudah ada atau event history baru.

---

# Phase 0 — Discovery & Baseline

## Task 0.1: Inventory Route dan API Akademik Saat Ini

**Objective:** Mendokumentasikan kemampuan akademik saat ini agar implementasi tidak duplikatif.

**Files:**
- Read: `apps/web-admin/src/routes/akademik/rombel/+page.svelte`
- Read: `apps/web-admin/src/routes/akademik/rombel/[id]/+page.svelte`
- Read: `apps/web-admin/src/routes/students/+page.svelte`
- Read: `apps/web-admin/src/routes/api/academic/**/+server.ts`
- Read: `services/core-api/db/queries/academic.sql`
- Read: `services/core-api/db/queries/rombels.sql`
- Read: `services/core-api/db/queries/students.sql`
- Read: `services/core-api/internal/handler/*academic*` and `*rombel*` if present

**Steps:**
1. List all existing routes and API endpoints.
2. Mark each endpoint as read-only or mutable.
3. Identify missing endpoints for dashboard, inline edit, bulk edit, mapel flags, and matrix guru mapel.
4. Save short findings to this plan under `Implementation Notes` before coding.

**Verification:**
- There is a concrete endpoint gap list.
- No existing endpoint is accidentally rebuilt.

---

## Task 0.2: Define Academic UX Data Contracts

**Objective:** Menetapkan payload frontend/backend untuk dashboard dan editable screens.

**Files:**
- Create: `docs/contracts/academic-editable-ui.md` or append to existing docs if available.

**Contract Draft:**

```ts
type AcademicDashboardSummary = {
  active_academic_year: string;
  active_semester: string;
  total_classes: number;
  total_active_students: number;
  students_without_class: number;
  classes_without_homeroom: number;
  subject_assignments_missing_teacher: number;
  timetable_conflicts: number;
  student_accounts_missing: number;
  parent_accounts_missing: number;
};
```

```ts
type EditableRombelRow = {
  id: string;
  code: string;
  name: string;
  level: string;
  is_active: boolean;
  homeroom_teacher_id: string | null;
  homeroom_teacher_name: string;
  total_students: number;
  total_subject_assignments: number;
  total_timetable_slots: number;
};
```

```ts
type SubjectMatrixCell = {
  class_id: string;
  subject_id: string;
  assignment_id: string | null;
  teacher_employee_id: string | null;
  teacher_name: string;
  status: 'complete' | 'missing_teacher' | 'missing_assignment';
};
```

**Verification:**
- Contracts explain field meaning, editability, and validation rules.

---

# Phase 1 — Shared Editable UX Foundation

## Task 1.1: Create Shared Inline Edit Components

**Objective:** Membuat komponen reusable untuk edit cepat di tabel.

**Files:**
- Create: `apps/web-admin/src/lib/components/editable/EditableTextCell.svelte`
- Create: `apps/web-admin/src/lib/components/editable/EditableSelectCell.svelte`
- Create: `apps/web-admin/src/lib/components/editable/DirtyChangeBar.svelte`
- Create: `apps/web-admin/src/lib/components/editable/types.ts`

**Behavior:**
- `EditableTextCell` supports display mode, edit mode, invalid state, dirty marker.
- `EditableSelectCell` supports options, loading state, empty option.
- `DirtyChangeBar` shows `N perubahan belum disimpan`, buttons `Batalkan` and `Simpan`.

**Verification:**
- Components compile with `npm --prefix apps/web-admin run check`.
- Components are not tied to akademik only.

---

## Task 1.2: Create Shared Detail Drawer Pattern

**Objective:** Membuat pola drawer agar detail siswa/rombel/mapel tidak pakai modal kecil.

**Files:**
- Create: `apps/web-admin/src/lib/components/EntityDrawer.svelte`
- Optional: use existing shadcn sheet/dialog if already present.

**Behavior:**
- Supports title, subtitle, actions slot, body slot.
- Mobile: full-screen sheet.
- Desktop: right-side panel width `xl`.
- Escape/click outside respects unsaved changes guard.

**Verification:**
- Rendered from a test/demo page or first consumer later.
- No direct business logic inside drawer.

---

## Task 1.3: Add Unsaved Changes Guard Utility

**Objective:** Mencegah operator kehilangan perubahan saat pindah halaman.

**Files:**
- Create: `apps/web-admin/src/lib/client/unsaved-changes.ts`

**Behavior:**
- Function `confirmDiscardChanges(hasChanges: boolean): Promise<boolean>`.
- Uses existing `confirmAction` pattern.
- Reusable by rombel, siswa, jadwal.

**Verification:**
- Manual check on first integrated page.

---

# Phase 2 — Dashboard Akademik

## Task 2.1: Backend Query Dashboard Summary

**Objective:** Menyediakan ringkasan masalah akademik dalam satu endpoint.

**Files:**
- Modify: `services/core-api/db/queries/academic.sql`
- Modify generated: `services/core-api/internal/repository/postgres/academic.sql.go`
- Modify service/handler academic files as appropriate.

**Query Needs:**
- active academic year
- total active classes
- total active students
- students without class
- classes without homeroom
- subject assignments missing teacher
- timetable conflicts placeholder if detection not ready
- missing student/parent accounts if available

**Command:**

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/service ./internal/handler ./internal/repository/postgres
```

**Verification:**
- Endpoint returns summary for current production DB.
- Does not break existing `/api/academic` consumers.

---

## Task 2.2: BFF Proxy for Dashboard Summary

**Objective:** Add SvelteKit API adapter for dashboard summary.

**Files:**
- Create: `apps/web-admin/src/routes/api/academic/dashboard/+server.ts`

**Behavior:**
- Thin proxy to Go API.
- No DB access.
- Reuse existing auth/proxy helper patterns.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

---

## Task 2.3: Create Akademik Dashboard Page

**Objective:** Membuat landing `/akademik` yang memberi status kesehatan akademik.

**Files:**
- Create: `apps/web-admin/src/routes/akademik/+page.svelte`

**UI Sections:**
1. Header: tahun ajaran aktif + semester aktif.
2. Status cards:
   - Rombel aktif
   - Siswa aktif
   - Siswa tanpa rombel
   - Rombel tanpa wali kelas
   - Guru mapel belum lengkap
   - Jadwal bentrok
3. Quick actions:
   - Kelola Rombel
   - Kelola Siswa
   - Guru Mapel
   - Jadwal Pelajaran
4. Checklist kesiapan akademik.

**Verification:**
- Page loads with skeleton, error recovery, and status cards.
- Mobile layout readable.

---

# Phase 3 — Rombel Editable

## Task 3.1: Add Update Rombel Backend Endpoint

**Objective:** Mendukung edit kode, nama, tingkat, status aktif rombel.

**Files:**
- Modify: `services/core-api/db/queries/rombels.sql`
- Modify: `services/core-api/internal/service/rombel*.go`
- Modify: `services/core-api/internal/handler/rombel*.go`
- Regenerate sqlc.

**Validation:**
- `code` required, unique per academic year.
- `name` required.
- `level` allowed: `VII`, `VIII`, `IX` unless existing system allows others.
- Cannot deactivate class if active students exist unless explicit force endpoint exists later.

**Verification:**
- Unit/handler tests for invalid duplicate code and successful update.

---

## Task 3.2: Add Inline Edit to Rombel List

**Objective:** Operator bisa edit rombel langsung dari tabel.

**Files:**
- Modify: `apps/web-admin/src/routes/akademik/rombel/+page.svelte`
- Use: `EditableTextCell`, `EditableSelectCell`, `DirtyChangeBar`

**UI Behavior:**
- Add toggle `Edit Cepat`.
- Editable columns:
  - code
  - name
  - level
  - is_active
- Dirty rows highlighted.
- Save all changed rows in sequence with progress.

**Verification:**
- Edit one rombel, save, reload, value persists.
- Cancel restores original data.
- Failed save shows toast and keeps dirty row.

---

## Task 3.3: Improve Rombel Detail Header

**Objective:** Detail rombel lebih actionable.

**Files:**
- Modify: `apps/web-admin/src/routes/akademik/rombel/[id]/+page.svelte`

**UI Additions:**
- Header cards:
  - siswa
  - wali kelas
  - guru mapel
  - jadwal
- Button group:
  - Edit Identitas Rombel
  - Atur Wali Kelas
  - Tambah Guru Mapel
  - Tambah Jadwal
- Sticky tab nav.

**Verification:**
- Existing wali kelas/guru mapel/jadwal flows still work.

---

# Phase 4 — Siswa Editable + Drawer

## Task 4.1: Add Student Patch Endpoint

**Objective:** Mendukung update sebagian data siswa tanpa form besar.

**Files:**
- Modify: `services/core-api/db/queries/students.sql`
- Modify: student service/handler.

**Fields:**
- `nis`
- `nisn`
- `nama`
- `gender`
- `class_id`
- `parent_name`
- `parent_phone`
- `is_active`
- `status`

**Validation:**
- NIS unique.
- NISN optional but if provided should be numeric-ish and unique if policy requires.
- class_id must exist in active academic year unless moving archived student.

**Verification:**
- Tests cover NIS duplicate and class change.

---

## Task 4.2: Convert Student Form to Drawer

**Objective:** Mengganti form siswa yang terasa panjang menjadi detail drawer dengan tab.

**Files:**
- Modify: `apps/web-admin/src/routes/students/+page.svelte`
- Use: `EntityDrawer.svelte`

**Drawer Tabs:**
- Profil
- Akademik
- Orang Tua
- Akun
- Riwayat

**Verification:**
- Add student works.
- Edit student works.
- Account generation actions remain accessible.

---

## Task 4.3: Add Bulk Student Actions

**Objective:** Mempercepat pengelolaan siswa massal.

**Files:**
- Modify: `apps/web-admin/src/routes/students/+page.svelte`
- Backend endpoint if not available: `PATCH /students/bulk` or equivalent.

**Bulk Actions:**
- pindah rombel
- aktif/nonaktif
- ubah status
- generate akun siswa selected/all filtered

**UX:**
- Checkbox row selection.
- Toolbar appears when selected count > 0.
- Confirmation dialog shows count and affected class/status.

**Verification:**
- Select 2 students, move class, reload.
- Unauthorized user cannot bulk mutate.

---

# Phase 5 — Mapel Editable

## Task 5.1: Add Subject Metadata Migration

**Objective:** Mapel bisa dibedakan antara mapel asesmen, rapor, dan kegiatan jadwal.

**Files:**
- Create migration: `services/core-api/db/migrations/0XX_academic_subject_metadata.sql`

**Columns to add to `subjects`:**

```sql
ALTER TABLE subjects
  ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'intrakurikuler',
  ADD COLUMN IF NOT EXISTS is_assessment_subject BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS is_report_subject BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS is_schedule_activity BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS default_weekly_hours INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;
```

**Seed behavior:**
- Upacara, Religi, Istirahat, Senam, Pembiasaan Diri should be schedule activity if inserted later, not assessment subject.
- BK/P5RA can be configured based on school policy.

**Verification:**
- Migration applies cleanly.
- Existing mapel keep active behavior.

---

## Task 5.2: Create Mapel Management Page

**Objective:** Admin bisa mengedit mapel tanpa SQL/manual seed.

**Files:**
- Create: `apps/web-admin/src/routes/akademik/mapel/+page.svelte`
- Create/proxy: `apps/web-admin/src/routes/api/academic/subjects/+server.ts`
- Backend CRUD if missing.

**Editable Columns:**
- code
- name
- category
- assessment subject
- report subject
- schedule activity
- weekly hours
- active
- display order

**Verification:**
- Add Muatan Lokal Keagamaan/Mulok Kewirausahaan from UI if not present.
- Toggle assessment flag and reload.

---

# Phase 6 — Guru Mapel Matrix

## Task 6.1: Add Matrix Read Endpoint

**Objective:** Backend returns mapel × rombel matrix for active academic year.

**Files:**
- Modify: `services/core-api/db/queries/rombels.sql` or `academic.sql`
- Add service/handler endpoint: `GET /api/academic/subject-assignment-matrix`
- Add BFF: `apps/web-admin/src/routes/api/academic/subject-assignment-matrix/+server.ts`

**Response:**

```ts
type SubjectAssignmentMatrix = {
  academic_year_id: string;
  classes: Array<{ id: string; code: string; name: string; level: string }>;
  subjects: Array<{ id: string; code: string; name: string }>;
  cells: SubjectMatrixCell[];
};
```

**Verification:**
- Returns 7 rombel × active assessment/schedule subjects for current DB.

---

## Task 6.2: Create Guru Mapel Matrix Page

**Objective:** Admin bisa assign guru per mapel per rombel dari satu layar.

**Files:**
- Create: `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte`

**UI:**
- Rows: mapel.
- Columns: rombel.
- Cell: guru selected or `Belum ada`.
- Filter: tingkat, mapel, guru.
- Cell click opens teacher select popover/drawer.
- Save changed cells.

**Verification:**
- Assign one guru to one cell.
- Copy one subject row to another class if included in this phase.
- Matrix reload reflects changes.

---

## Task 6.3: Add Copy Tools for Guru Mapel

**Objective:** Mempercepat setup mapel saat awal tahun.

**Features:**
- Copy guru mapel from one rombel to selected rombel.
- Copy from previous academic year if data exists.
- Clear selected cells with confirmation.

**Verification:**
- Copy VII.A assignments to VII.B.
- Conflict/overwrite preview appears before save.

---

# Phase 7 — Jadwal Pelajaran Basic Weekly Editor

## Task 7.1: Add Timetable Conflict Query

**Objective:** Backend bisa mendeteksi bentrok guru/jam.

**Files:**
- Modify: `services/core-api/db/queries/timetable.sql`
- Service/handler endpoint: `GET /api/academic/timetable/conflicts`

**Conflict Types:**
- same teacher at overlapping time.
- same room label at overlapping time if room_label provided.
- invalid time range.

**Verification:**
- Unit/query tests or manual fixture query.
- Dashboard uses conflict count later.

---

## Task 7.2: Create Jadwal Pelajaran Page

**Objective:** Menyediakan tampilan jadwal mingguan awal.

**Files:**
- Create: `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`

**Modes:**
- Per Rombel
- Per Guru
- Mingguan

**Initial Behavior:**
- No drag & drop yet.
- Click empty slot: add slot.
- Click existing block: edit slot.
- Show conflict badges.
- Kegiatan khusus can be displayed as blocks if represented in subjects or special activities.

**Verification:**
- Add/edit/delete slot via existing rombel timetable endpoints or new endpoints.
- Conflict badge updates after reload.

---

# Phase 8 — Tahun Ajaran Wizard

## Task 8.1: Plan Clone/Promotion Backend Safely

**Objective:** Menyiapkan workflow awal tahun ajaran tanpa langsung menerapkan destructive changes.

**Endpoint Draft:**
- `POST /api/academic/year-rollover/preview`
- `POST /api/academic/year-rollover/apply`

**Preview Should Show:**
- classes to create
- students to promote
- students without next class
- homeroom assignments to copy/clear
- subject assignments to copy
- timetable slots to copy/clear

**Guardrails:**
- Apply only after preview token/confirmation.
- Never delete previous year data.

---

## Task 8.2: Create Tahun Ajaran & Semester Page

**Objective:** Admin bisa kelola tahun ajaran dan semester dari UI.

**Files:**
- Create: `apps/web-admin/src/routes/akademik/tahun-ajaran/+page.svelte`

**Features:**
- list years
- activate year
- create new year
- start wizard rollover

**Verification:**
- Create non-active academic year.
- Active-year change uses explicit confirmation.

---

# Phase 9 — Import/Export Excel

## Task 9.1: Export Templates

**Objective:** Operator bisa unduh template Excel/CSV.

**Templates:**
- siswa
- rombel
- guru mapel
- jadwal pelajaran

**Files:**
- Backend export endpoints or SvelteKit generated CSV if data is from API only.

**Verification:**
- Download opens in spreadsheet app.
- Headers are Indonesian and match import fields.

---

## Task 9.2: Import Dry-Run

**Objective:** Import massal aman dengan preview.

**Behavior:**
- Upload file.
- Validate rows.
- Show: tambah, ubah, skip, error.
- Apply only after explicit confirmation.

**Verification:**
- Invalid row shows exact row number and message.
- Dry-run does not mutate DB.

---

# Phase 10 — Audit, Polish, and Navigation

## Task 10.1: Navigation Update

**Objective:** Menu Akademik tampil lengkap dan konsisten.

**Files:**
- Modify navigation definitions in `apps/web-admin/src/lib` or layout files where sidebar is defined.

**Menu:**
- Dashboard
- Rombel
- Siswa
- Mapel
- Guru Mapel
- Jadwal
- Tahun Ajaran

**Verification:**
- Admin sees menu.
- Unauthorized roles do not get admin-only mutators.

---

## Task 10.2: Audit Log Integration

**Objective:** Mutasi penting akademik terekam.

**Events:**
- rombel updated
- student moved class
- homeroom changed
- subject teacher changed
- timetable slot changed
- academic year activated

**Verification:**
- Audit logs page shows changes with actor and timestamp.

---

## Task 10.3: End-to-End QA Checklist

**Objective:** Validate full operator workflow.

**Scenario:**
1. Open Akademik dashboard.
2. Fix missing homeroom.
3. Edit rombel name/code.
4. Move a student to another rombel.
5. Assign teacher in guru mapel matrix.
6. Add timetable slot.
7. Confirm dashboard warning count decreases.

**Commands:**

```bash
npm --prefix apps/web-admin run check
cd services/core-api && go test ./...
```

**Manual QA:**
- Desktop 1366px.
- Tablet width.
- Mobile width.
- Slow API/loading states.
- Backend error states.

---

# Suggested Execution Order

## Sprint Akademik 1 — Foundation + Dashboard

1. Phase 0
2. Phase 1
3. Phase 2

**Deliverable:** `/akademik` dashboard + reusable editable UX components.

## Sprint Akademik 2 — Rombel & Siswa Editable

1. Phase 3
2. Phase 4

**Deliverable:** Rombel and siswa management feel editable and fast.

## Sprint Akademik 3 — Mapel & Guru Mapel Matrix

1. Phase 5
2. Phase 6

**Deliverable:** Mapel can be maintained from UI; teacher assignments can be handled per rombel matrix.

## Sprint Akademik 4 — Jadwal Basic

1. Phase 7

**Deliverable:** Jadwal pelajaran has a dedicated weekly editor and conflict checks.

## Sprint Akademik 5 — Year Rollover + Import/Export

1. Phase 8
2. Phase 9
3. Phase 10

**Deliverable:** Operator can prepare new academic year safely and import/export core academic data.

---

# Acceptance Criteria

- Admin can edit common academic data without SQL/manual backend intervention.
- Rombel and siswa can be edited from list/detail UI with validation.
- Guru mapel supports per-mapel-per-rombel assignments, matching school reality.
- Jadwal supports special activities and normal lessons.
- Dashboard surfaces incomplete academic setup clearly.
- Bulk operations have confirmation and preview where needed.
- Tests and checks pass before deployment.
- No direct DB access from web-admin.

---

# Implementation Notes

## Sprint Akademik 1 Phase 0 Discovery — 2026-05-11

- Existing web routes:
  - `/academic`: legacy/general Data Akademik page for academic years, classes, subjects, assignments, and timetable CRUD through `/api/academic`.
  - `/akademik/rombel`: rombel list with summary cards and detail links.
  - `/akademik/rombel/[id]`: rombel detail for students, wali kelas, guru mapel, timetable slots, and journal-session open action.
  - `/students`: student list/form, currently fetches `/api/students` and `/api/academic` for class options.
- Existing SvelteKit BFF academic proxies:
  - `GET /api/academic` → `GET /api/academic` Go overview, read-only.
  - `POST /api/academic?entity=...` → `POST /api/academic/{entity}`, mutable.
  - `PUT /api/academic?entity=...&id=...` → `PUT /api/academic/{entity}/{id}`, mutable, currently only timetable updates are handled by Go.
  - `DELETE /api/academic?entity=...&id=...` → `DELETE /api/academic/{entity}/{id}`, mutable.
  - `GET /api/academic/stats` → basic counts only.
  - `GET /api/academic/rombel` and `GET /api/academic/rombel/[id]`, read-only.
  - Rombel nested BFF routes for `students`, `subject-assignments`, `timetable-slots`, `homeroom-assignments`, and `journal-session`; subject/timetable/homeroom collection/item routes include create/update/delete mutators where Go already supports them.
- Existing Go academic API routes:
  - `GET /api/academic`, `GET /api/academic/stats`.
  - `GET /api/academic/rombel`, `GET /api/academic/rombel/{id}`, and nested `students`, `subject-assignments`, `timetable-slots`, `homeroom-assignments`.
  - Mutable academic routes exist for generic create/delete and timetable update, plus rombel-scoped guru mapel, jadwal, and wali kelas mutators guarded by `academic.manage`.
- Current backend query coverage:
  - `academic.sql` covers years/classes/subjects/class-subject assignments/basic stats.
  - `rombels.sql` covers rombel summaries/detail, class students with linked parents, guru mapel, timetable slots, and wali kelas assignment history.
  - `students.sql` covers list/get/create/full update/delete/status lifecycle and teacher-scoped listing.
- Contract baseline added in `docs/contracts/academic-editable-ui.md` for dashboard summary, editable rombel rows, guru mapel matrix, and shared editable component types.
- Endpoint gaps for later phases:
  - Dashboard summary endpoint was missing before Sprint 1; Sprint 1 adds `GET /api/academic/dashboard` and BFF `/api/academic/dashboard`.
  - Inline rombel identity update endpoint is missing; generic `PUT /api/academic/classes/{id}` is not implemented in Go yet.
  - Partial student patch and bulk student update endpoints are missing.
  - Subject metadata flags and subject-specific CRUD route are missing beyond generic legacy subject create/delete.
  - Guru mapel matrix endpoint/page is missing.
  - Timetable conflict counting exists for mutation validation, but no dedicated dashboard/global conflict summary endpoint exists before Sprint 1.

## Sprint Akademik 2 Phase 3 Rombel — 2026-05-11

- Added `PUT /api/academic/rombel/{id}` in Go core-api and SvelteKit BFF for editing rombel identity: `code`, `name`, `level`, and `is_active`.
- Backend validation trims code/name, normalizes level to `VII`/`VIII`/`IX`, rejects duplicate codes within the same academic year, and rejects deactivating a rombel that still has active students.
- `/akademik/rombel` now has `Edit Cepat` mode using `EditableTextCell`, `EditableSelectCell`, `DirtyChangeBar`, unsaved-change guard, sequential row saves, and row-level error messages.
- `/akademik/rombel/[id]` header was improved with stronger quick stats/actions while preserving existing wali kelas, guru mapel, siswa, and jadwal flows.

## Sprint Akademik 2 Phase 4 Siswa — 2026-05-11

- `/students` add/edit form now uses shared `EntityDrawer` from Sprint 1 while preserving existing add, edit, delete, lifecycle update, preview account, and generate account flows.
- Student selection is implemented with row/card checkboxes, a filtered-table select-all checkbox, selected count, and clear-selection toolbar.
- Bulk mutation is intentionally deferred because there is no safe selected-student bulk endpoint in the current contract; the toolbar only reports selection count and clears selection.
