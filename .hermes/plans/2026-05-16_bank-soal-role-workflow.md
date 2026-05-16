# Bank Soal Role Workflow Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Implement Bank Soal workflow berbasis role: pembuat soal, reviewer mapel, approver/publisher, operator paket, dan admin; dengan visibilitas soal yang aman, audit jelas, dan soal published/used tidak berubah tanpa versi baru.

**Architecture:** Gunakan arsitektur existing: PostgreSQL + migrations + sqlc di Go core-api, lalu SvelteKit sebagai BFF/proxy ke Go API. Implementasi dibuat additive dan bertahap: perbaiki permission/scope dulu, lalu workflow action, UI review/approval, audit, dan integrasi paket CBT. SvelteKit tidak boleh akses DB langsung.

**Tech Stack:** PostgreSQL migrations, Go core-api, sqlc, SvelteKit web-admin, RBAC existing, PM2 deploy manual setelah diminta.

---

## Current findings

- Bank Soal sudah punya route UI:
  - `/bank-soal`
  - `/bank-soal/daftar`
  - `/bank-soal/tambah`
  - `/bank-soal/verifikasi`
  - `/bank-soal/penerbitan`
  - `/bank-soal/impor`
  - `/bank-soal/analisis-butir`
  - `/bank-soal/mapel-kd`
  - `/bank-soal/pengaturan`
  - `/bank-soal/soal/[id]`
- Sidebar sudah mengenal permission:
  - `bank_soal.read`
  - `bank_soal.create`
  - `bank_soal.review`
  - `bank_soal.publish`
  - `bank_soal.import`
  - `bank_soal.analytics`
  - `bank_soal.settings`
- Migration `015_cbt_question_bank_standardization.sql` sudah menambahkan kolom workflow di `cbt_questions`:
  - `workflow_status` dengan nilai saat ini: `draft`, `review`, `approved`, `rejected`
  - `author_username`
  - `reviewer_username`
  - `reviewed_at`
  - `approver_username`
  - `approved_at`
  - `writer_notes`
  - `review_notes`
- Migration `080_bank_soal_teacher_permission_hardening.sql` sudah menghapus permission sensitif dari role `guru`:
  - `bank_soal.review`
  - `bank_soal.import`
  - `bank_soal.settings`
  - `bank_soal.publish`
  - `bank_soal.delete`
- Ada plan terpisah untuk versioning:
  - `.hermes/plans/2026-05-15_bank-soal-question-versioning.md`
- Plan ini fokus role workflow. Jika butuh immutable version untuk published/used, jalankan plan versioning sebelum/bersamaan Sprint 5.

---

## Target role model

### Role operasional

1. **Guru Pembuat Soal**
   - Membuat soal.
   - Melihat dan mengedit soal sendiri saat `draft`, `rejected`, atau `revision_needed`.
   - Submit soal ke review.

2. **Reviewer Mapel**
   - Melihat soal `submitted/review` sesuai mapel/tingkat scope.
   - Memberi catatan revisi.
   - Menandai soal layak/tidak layak.
   - Tidak boleh approve final/publish.
   - Tidak boleh review soal sendiri kecuali admin override.

3. **Approver Akademik**
   - Melihat soal yang sudah direview/recommended.
   - Approve/publish soal.
   - Mengembalikan ke revisi jika belum layak.

4. **Operator CBT / Paket Ujian**
   - Melihat soal `approved/published`.
   - Memakai soal ke paket CBT.
   - Tidak mengubah isi soal.

5. **Admin Sistem**
   - Mengelola semua akses, import/export, audit, setting, dan emergency correction.

### Permission target

- Existing yang dipertahankan:
  - `bank_soal.read`
  - `bank_soal.create`
  - `bank_soal.review`
  - `bank_soal.publish`
  - `bank_soal.import`
  - `bank_soal.analytics`
  - `bank_soal.settings`
  - `bank_soal.delete`
- Tambahan yang disarankan:
  - `bank_soal.update_own`
  - `bank_soal.submit`
  - `bank_soal.approve`
  - `bank_soal.assign_reviewer`
  - `bank_soal.read_all`
  - `bank_soal.use_in_package`
  - `bank_soal.audit`

---

## Target workflow status

Existing `workflow_status` saat ini: `draft`, `review`, `approved`, `rejected`.

Target status yang lebih jelas:

1. `draft` — masih dibuat oleh author.
2. `submitted` — dikirim ke reviewer, belum diputuskan.
3. `revision_needed` — reviewer meminta revisi.
4. `reviewed` — reviewer menyatakan layak, menunggu approver.
5. `approved` — disetujui akhir, siap masuk paket.
6. `published` — resmi siap dipakai/terkunci.
7. `rejected` — ditolak.
8. `archived` — tidak dipakai untuk paket baru.

Catatan kompatibilitas:
- Status lama `review` dipetakan sebagai `submitted` di UI, atau tetap diterima sebagai alias selama transisi.
- Jangan langsung rename massal tanpa migration backfill yang aman.

---

## Target visibility rules

### Guru biasa

Boleh lihat:
- Soal milik sendiri di semua status.
- Soal `approved/published` sesuai mapel/tingkat yang dia ampu, jika permission `bank_soal.read` diberikan.

Tidak boleh lihat:
- Draft guru lain.
- Submitted/revision/rejected guru lain kecuali dia reviewer scope tersebut.
- Soal lintas mapel/tingkat tanpa assignment.
- Kunci jawaban soal guru lain sebelum `approved/published`.

### Reviewer

Boleh lihat:
- Soal submitted/revision/reviewed sesuai mapel/tingkat scope reviewer.
- Soal milik sendiri mengikuti aturan guru biasa.

Tidak boleh:
- Review soal sendiri, kecuali admin override tercatat audit.
- Approve/publish final.

### Approver/Publisher

Boleh lihat:
- Semua soal yang sudah reviewed/approved/published dalam scope akademik.
- Metadata dan riwayat review.

Boleh aksi:
- approve
- publish
- return to revision
- archive

### Operator CBT

Boleh lihat:
- Soal approved/published yang boleh dipakai paket.

Tidak boleh:
- Ubah isi soal.
- Ambil soal draft/submitted/revision/rejected ke paket resmi.

---

## Production-safe rollout principles

Plan ini dirancang agar aman walaupun database sudah berisi soal, paket, event CBT, role, dan permission aktif.

Non-negotiable safety rules:

1. **Audit dulu, baru ubah behavior.** Jangan deploy filter/lock baru sebelum tahu jumlah data terdampak.
2. **Migration harus additive dan backwards-compatible.** Tabel/kolom/permission baru boleh ditambah; jangan drop data atau rename status lama tanpa fase transisi.
3. **Status lama `review` tetap diterima.** UI boleh menampilkan sebagai “Menunggu Review”, tetapi backend tetap menerima nilai lama sampai semua data/code siap.
4. **Soft enforcement sebelum hard enforcement.** Awalnya tampilkan warning dan filter opt-in; hard block baru aktif setelah reviewer scope dan permission lengkap.
5. **Admin/read_all tetap punya escape hatch.** Jangan sampai data terlihat “hilang” dari admin saat rollout.
6. **Paket/event existing tidak boleh rusak.** Pembatasan soal approved/published diterapkan dulu untuk paket baru; paket lama hanya dilaporkan sampai data siap.
7. **Published/used question tidak diedit in-place.** Jika sudah dipakai, gunakan revision/versioning flow.
8. **Tidak ada deploy/restart PM2 tanpa approval eksplisit.** Setiap sprint selesai dengan validate + commit; production deploy terpisah.
9. **Backup sebelum migration production.** Simpan path backup dan checksum di laporan deploy, bukan di memory permanen.
10. **Credential tidak boleh ditampilkan.** Semua `.env`, `DATABASE_URL`, token, password harus disembunyikan sebagai `[REDACTED]` jika muncul.

---

## Sprint 0 — Production audit & safe rollout gate

**Objective:** Mengukur kondisi data production yang sudah ada agar Sprint 1–5 tidak memutus akses guru, reviewer, admin, atau paket CBT existing.

**Files:**
- Create: `docs/reports/bank-soal-role-workflow-audit.md`
- Optional script: `scripts/audit-bank-soal-role-workflow.sql`

### Task 0.1: Audit status soal existing

**Objective:** Mengetahui distribusi `workflow_status` dan risiko status lama.

**SQL draft:**
```sql
SELECT workflow_status, status, COUNT(*) AS total
FROM cbt_questions
GROUP BY workflow_status, status
ORDER BY workflow_status, status;
```

**Check:**
- Berapa soal `draft`, `review`, `approved`, `rejected`.
- Apakah ada status di luar constraint target.
- Apakah `author_username`, `reviewer_username`, `approver_username` banyak yang kosong.

**Output:** Masukkan ringkasan ke `docs/reports/bank-soal-role-workflow-audit.md` tanpa credential.

### Task 0.2: Audit soal yang sudah masuk paket/event CBT

**Objective:** Mencegah hard enforcement `approved/published only` merusak paket/event existing.

**SQL draft:**
```sql
SELECT q.workflow_status, q.status, COUNT(*) AS total
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
GROUP BY q.workflow_status, q.status
ORDER BY q.workflow_status, q.status;
```

**Check:**
- Apakah paket existing memakai soal non-approved.
- Jika iya, Sprint 4 harus memakai **report-only mode** dulu untuk paket lama.

### Task 0.3: Audit role dan permission Bank Soal existing

**Objective:** Memastikan perubahan permission tidak mengunci guru/admin yang sedang aktif.

**SQL draft:**
```sql
SELECT r.code AS role_code, p.code AS permission_code, COUNT(*) AS grants
FROM rbac_role_permissions rp
JOIN rbac_roles r ON r.id = rp.role_id
JOIN rbac_permissions p ON p.id = rp.permission_id
WHERE p.code LIKE 'bank_soal.%'
GROUP BY r.code, p.code
ORDER BY r.code, p.code;
```

**Check:**
- Role `guru` hanya punya permission authoring aman.
- Role admin masih punya permission lengkap.
- Permission sensitif seperti review/publish/import/settings tidak tersebar ke semua guru.

### Task 0.4: Audit calon reviewer/approver dari data guru-mapel

**Objective:** Menentukan default reviewer scope tanpa input manual berlebihan.

**Data source:**
- `class_subject_assignments`
- `employees`/`users` sesuai mapping existing.
- `subjects`

**Output:**
- Daftar guru per mapel/tingkat.
- Rekomendasi reviewer mapel awal, tetapi **jangan auto-grant** sebelum approval admin.

### Task 0.5: Compatibility scan code/UI

**Objective:** Cari code yang masih hardcode status lama atau permission lama.

**Commands:**
```bash
git grep -n "workflow_status\|bank_soal.review\|bank_soal.publish\|review_notes\|reviewer_username" -- services/core-api apps/web-admin | tee /tmp/bank-soal-workflow-compat.txt
```

**Check:**
- UI yang hanya mengenal `review` harus tetap kompatibel.
- Backend mapper harus menerima status lama dan baru.
- Sidebar/route permission tidak boleh mendadak menyembunyikan halaman penting dari admin.

### Task 0.6: Write rollout decision report

**Files:**
- Create/Update: `docs/reports/bank-soal-role-workflow-audit.md`

**Report sections:**
- Data counts per status.
- Paket existing yang memakai status non-approved.
- Role/permission existing.
- Calon reviewer/approver.
- Compatibility risks.
- Go/no-go decision for Sprint 1.

**Verification:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-workflow-sprint0 ./cmd/api
```
Expected: PASS or clearly documented pre-existing failure.

**Commit:**
```bash
git add docs/reports scripts .hermes/plans/2026-05-16_bank-soal-role-workflow.md
 git commit -m "docs(bank-soal): add production-safe workflow rollout audit"
```

---

## Sprint 1 — RBAC & scope foundation

**Objective:** Menyiapkan permission, role mapping, dan scope reviewer tanpa mengubah workflow besar.

### Task 1.1: Tambahkan permission Bank Soal baru

**Files:**
- Create: `services/core-api/db/migrations/103_bank_soal_role_workflow_permissions.sql`
- Check existing permission seed/migration sebelum finalisasi nama file.

**Steps:**
1. Tambahkan permission jika belum ada:
   - `bank_soal.update_own`
   - `bank_soal.submit`
   - `bank_soal.approve`
   - `bank_soal.assign_reviewer`
   - `bank_soal.read_all`
   - `bank_soal.use_in_package`
   - `bank_soal.audit`
2. Grant default aman hanya jika Sprint 0 menunjukkan role tersebut memang ada dan aktif:
   - role `guru`: `bank_soal.read`, `bank_soal.create`, `bank_soal.update_own`, `bank_soal.submit`
   - role admin/superadmin: permission lengkap, termasuk `read_all`, `assign_reviewer`, `audit`
3. **Jangan revoke permission existing pada migration Sprint 1** kecuali sudah ada laporan Sprint 0 + approval eksplisit. Jika ada permission sensitif terlanjur diberikan ke guru, tulis sebagai temuan dan siapkan migration hardening terpisah.
4. Pastikan migration idempotent:
   - `INSERT ... ON CONFLICT DO NOTHING`
   - `CREATE TABLE IF NOT EXISTS`
   - `CREATE INDEX IF NOT EXISTS`
5. Jalankan migration dry-run transaction:
   ```bash
   cd services/core-api
   set -a; source ../../.env >/dev/null 2>&1; set +a
   { echo 'BEGIN;'; cat db/migrations/103_bank_soal_role_workflow_permissions.sql; echo 'ROLLBACK;'; } | psql "$DATABASE_URL" -v ON_ERROR_STOP=1
   ```
   Expected: exit 0, tidak menampilkan credential.

### Task 1.2: Tambahkan tabel reviewer scope

**Files:**
- Create/extend migration: `services/core-api/db/migrations/103_bank_soal_role_workflow_permissions.sql`

**Schema proposal:**
```sql
CREATE TABLE IF NOT EXISTS bank_soal_reviewer_scopes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  subject_id UUID REFERENCES subjects(id) ON DELETE CASCADE,
  grade_level SMALLINT,
  can_review BOOLEAN NOT NULL DEFAULT TRUE,
  can_approve BOOLEAN NOT NULL DEFAULT FALSE,
  assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, subject_id, grade_level)
);

CREATE INDEX IF NOT EXISTS idx_bank_soal_reviewer_scopes_user
  ON bank_soal_reviewer_scopes(user_id, can_review, can_approve);

CREATE INDEX IF NOT EXISTS idx_bank_soal_reviewer_scopes_subject_grade
  ON bank_soal_reviewer_scopes(subject_id, grade_level);
```

**Rules:**
- `subject_id IS NULL` berarti scope semua mapel jika role/permission memang mengizinkan.
- `grade_level IS NULL` berarti semua tingkat dalam subject tersebut.
- Hanya admin/approver dengan `bank_soal.assign_reviewer` boleh mengelola scope.

### Task 1.3: Tambahkan query sqlc reviewer scope

**Files:**
- Modify: `services/core-api/db/queries/...` pilih file Bank Soal/query yang sesuai existing.
- Generated: `services/core-api/internal/repository/postgres/*.sql.go`

**Required queries:**
- List reviewer scopes.
- Upsert reviewer scope.
- Delete reviewer scope.
- Check if user can review subject/grade.
- Check if user can approve subject/grade.

**Verification:**
```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/repository/postgres
```
Expected: PASS.

### Task 1.4: Web-admin settings UI untuk reviewer scope

**Files:**
- Create/modify: `apps/web-admin/src/routes/bank-soal/pengaturan/+page.svelte`
- Create BFF if needed: `apps/web-admin/src/routes/api/bank-soal/reviewer-scopes/+server.ts`

**UI:**
- Daftar reviewer.
- Pilih user/guru.
- Pilih mapel.
- Pilih tingkat.
- Toggle can_review/can_approve.
- Save/delete scope.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add services/core-api apps/web-admin
 git commit -m "feat(bank-soal): add role workflow permissions and reviewer scopes"
```

---

## Sprint 2 — Backend workflow actions & visibility filter

**Objective:** Semua aksi workflow Bank Soal divalidasi backend, bukan hanya UI.

### Task 2.1: Perluas workflow status secara backward-compatible

**Files:**
- Create migration: `services/core-api/db/migrations/104_bank_soal_workflow_statuses.sql`

**Migration approach:**
1. Drop constraint lama `chk_cbt_questions_workflow_status`.
2. Add constraint baru yang menerima:
   - `draft`
   - `review`
   - `submitted`
   - `revision_needed`
   - `reviewed`
   - `approved`
   - `published`
   - `rejected`
   - `archived`
3. Jangan backfill status lama dulu kecuali semua code sudah siap.

**Verification:** dry-run transaction via psql seperti Sprint 1.

### Task 2.2: Tambahkan audit trail workflow

**Files:**
- Create migration: `services/core-api/db/migrations/104_bank_soal_workflow_statuses.sql`

**Schema proposal:**
```sql
CREATE TABLE IF NOT EXISTS bank_soal_question_workflow_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id UUID NOT NULL REFERENCES cbt_questions(id) ON DELETE CASCADE,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  actor_username TEXT NOT NULL DEFAULT '',
  from_status TEXT NOT NULL DEFAULT '',
  to_status TEXT NOT NULL,
  action TEXT NOT NULL,
  note TEXT NOT NULL DEFAULT '',
  metadata JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_soal_workflow_events_question
  ON bank_soal_question_workflow_events(question_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_bank_soal_workflow_events_actor
  ON bank_soal_question_workflow_events(actor_username, created_at DESC);
```

### Task 2.3: Implement backend workflow actions

**Files:**
- Modify: `services/core-api/internal/service/...` Bank Soal question service.
- Modify: `services/core-api/internal/handler/...` Bank Soal question handler.
- Modify: `services/core-api/db/queries/...` related cbt_questions queries.
- Test: handler/service tests existing Bank Soal.

**Actions:**
- `submit_for_review`
  - allowed: author/admin
  - from: `draft`, `revision_needed`, `rejected`
  - to: `submitted`
- `request_revision`
  - allowed: reviewer/admin sesuai scope
  - from: `submitted`, `review`
  - to: `revision_needed`
- `mark_reviewed`
  - allowed: reviewer/admin sesuai scope
  - from: `submitted`, `review`
  - to: `reviewed`
- `reject`
  - allowed: reviewer/admin sesuai scope
  - from: `submitted`, `review`, `reviewed`
  - to: `rejected`
- `approve`
  - allowed: approver/admin sesuai scope
  - from: `reviewed`
  - to: `approved`
- `publish`
  - allowed: publisher/admin
  - from: `approved`
  - to: `published`
- `archive`
  - allowed: approver/admin
  - from: `approved`, `published`, `rejected`
  - to: `archived`

**Critical guardrails:**
- Reviewer tidak boleh review soal sendiri, kecuali actor punya admin override.
- Author tidak boleh approve/publish soal sendiri kecuali admin override.
- Soal yang sudah dipakai paket/sesi tidak boleh diedit in-place.
- Semua action menulis audit event.

### Task 2.4: Implement visibility filter di list/detail question

**Files:**
- Modify queries/list service Bank Soal.

**Filter logic:**
- Admin/read_all: lihat semua.
- Author: lihat own questions.
- Reviewer: lihat submitted/revision/reviewed sesuai reviewer scope.
- Approver: lihat reviewed/approved/published sesuai approver scope.
- Operator package: lihat approved/published.
- Guru biasa: lihat own + approved/published sesuai mapel/tingkat yang dia ampu.

**Tests:**
- Plain guru tidak melihat draft guru lain.
- Reviewer mapel A tidak melihat submitted mapel B.
- Approver bisa melihat reviewed.
- Operator tidak melihat draft/submitted.

**Verification:**
```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-workflow-sprint2 ./cmd/api
```
Expected: PASS.

**Commit:**
```bash
git add services/core-api
 git commit -m "feat(bank-soal): enforce workflow actions and visibility rules"
```

---

## Sprint 3 — Review & approval UI

**Objective:** Guru, reviewer, dan approver punya layar kerja yang jelas sesuai peran.

### Task 3.1: Update composer/tambah soal untuk submit workflow

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/tambah/+page.svelte`
- Modify if used: `apps/web-admin/src/routes/bank-soal/soal/[id]/+page.svelte`
- BFF action route if needed: `apps/web-admin/src/routes/api/bank-soal/questions/[id]/workflow/+server.ts`

**UI:**
- Tombol `Simpan Draft`.
- Tombol `Kirim Review`.
- Validasi metadata wajib sebelum submit.
- Tampilkan badge status.
- Jika status terkunci, composer read-only dan arahkan ke revision flow.

### Task 3.2: Update halaman verifikasi reviewer

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/verifikasi/+page.svelte`

**UI:**
- Filter: mapel, tingkat, status, author.
- Queue “Menunggu Review”.
- Detail soal read-only + panel catatan.
- Tombol:
  - `Minta Revisi`
  - `Tandai Layak`
  - `Tolak`
- Warning jika soal milik sendiri: action disabled kecuali admin.

### Task 3.3: Update halaman penerbitan/approval

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/penerbitan/+page.svelte`

**UI:**
- Queue `reviewed`.
- Tombol:
  - `Approve`
  - `Publish`
  - `Kembalikan Revisi`
  - `Arsipkan`
- Tampilkan reviewer, reviewed_at, review_notes.
- Tampilkan risiko: published/used akan terkunci.

### Task 3.4: Update daftar soal visibility cue

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/daftar/+page.svelte`

**UI:**
- Badge status workflow.
- Badge “Soal saya”, “Butuh review”, “Published”.
- Filter “Milik saya”, “Perlu review”, “Approved/Published”, “Semua yang bisa saya akses”.
- Jangan tampilkan kunci jawaban untuk item yang belum boleh dibuka.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add apps/web-admin
 git commit -m "feat(bank-soal): add review and approval workspace"
```

---

## Sprint 4 — Penggunaan soal ke paket CBT dan hardening kunci jawaban

**Objective:** Paket CBT hanya bisa mengambil soal approved/published dan kunci jawaban lebih terlindungi.

### Task 4.1: Harden package question selection backend

**Files:**
- Modify: Go service/handler package question assignment.
- Modify: queries related `cbt_package_questions`.

**Rules:**
- Paket resmi hanya boleh mengambil soal `approved`/`published`.
- Draft/submitted/revision/rejected tidak boleh masuk paket resmi.
- Admin override hanya untuk paket draft/test dan harus audit.

### Task 4.2: Hide answer key by permission/status

**Files:**
- Modify: question response mapper in Go service/handler.
- Modify: web-admin detail/list UI if needed.

**Rules:**
- Author boleh lihat kunci soal sendiri.
- Reviewer/approver boleh lihat kunci saat review/approval.
- Operator hanya melihat kunci jika permission khusus atau untuk audit tertentu.
- Guru lain tidak melihat kunci soal orang lain sebelum `approved/published`.

### Task 4.3: Add tests for package and key visibility

**Tests:**
- Package assignment rejects draft.
- Package assignment accepts approved/published.
- Plain guru cannot see another author answer key for submitted question.
- Reviewer can see answer key within scope.

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-workflow-sprint4 ./cmd/api
npm --prefix ../../apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add services/core-api apps/web-admin
 git commit -m "feat(bank-soal): restrict package usage and answer key visibility"
```

---

## Sprint 5 — Audit, reporting, and version-safe edit policy

**Objective:** Workflow siap dipakai resmi dengan audit trail, dashboard, dan kebijakan revisi aman.

### Task 5.1: Workflow timeline di detail soal

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/soal/[id]/+page.svelte`
- Add BFF if needed: `apps/web-admin/src/routes/api/bank-soal/questions/[id]/workflow-events/+server.ts`
- Backend endpoint if needed.

**UI:**
- Timeline status.
- Actor, waktu, note.
- Perubahan reviewer/approver.

### Task 5.2: Dashboard Bank Soal role-based

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/+page.svelte`
- Modify backend summary endpoint if existing.

**Cards:**
- Draft saya.
- Menunggu review saya.
- Perlu revisi.
- Menunggu approval.
- Approved/published siap paket.
- Soal kurang metadata.

### Task 5.3: Integrate with question versioning plan

**Files:**
- Use plan: `.hermes/plans/2026-05-15_bank-soal-question-versioning.md`

**Rules:**
- Published/used question is read-only.
- Edit published/used creates new version.
- Package/session remains pointing to exact old question ID.
- Detail page shows version history.

### Task 5.4: Documentation & SOP

**Files:**
- Create: `docs/contracts/bank-soal-role-workflow.md`

**Content:**
- Role matrix.
- Workflow status transition matrix.
- Visibility matrix.
- Package usage rule.
- Emergency admin override SOP.

**Verification full:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-workflow-final ./cmd/api
```
Expected: PASS.

**Commit:**
```bash
git add docs services/core-api apps/web-admin .hermes/plans/2026-05-16_bank-soal-role-workflow.md
 git commit -m "docs(bank-soal): document role workflow implementation plan"
```

---

## Implementation notes

### Sprint 1 — RBAC & reviewer scope foundation

Status: completed locally, not deployed/restarted.

Implemented:
- Additive migration `services/core-api/db/migrations/109_bank_soal_role_workflow_permissions.sql` for new Bank Soal permissions and `bank_soal_reviewer_scopes`.
- sqlc queries and generated repository for listing/upserting/deleting reviewer scopes plus review/approve scope checks.
- Go service/handler/routes for `/api/bank-soal/reviewer-scopes` guarded by `bank_soal.assign_reviewer`, `bank_soal.settings`, or admin.
- SvelteKit BFF proxy `apps/web-admin/src/routes/api/bank-soal/reviewer-scopes/+server.ts`.
- `/bank-soal/pengaturan` upgraded into governance/settings page with operational summary, SOP workflow, quality rules, integration links, and reviewer scope form/table.
- Frontend RBAC catalog/access helpers updated with additive permissions only.

Verified:
- Migration dry-run in transaction: PASS.
- `npm --prefix apps/web-admin run check`: PASS.
- `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`: PASS.
- `go test ./internal/handler ./internal/service ./internal/repository/postgres`: PASS.
- `go build -o /tmp/core-api-bank-soal-sprint1 ./cmd/api`: PASS.

Notes:
- No permission revoke was added.
- Reviewer scope is a foundation for Sprint 2+; visibility/action hard enforcement is intentionally not enabled yet.
- Production deploy requires explicit approval plus DB backup before running migration.

### Sprint 2 — Backend workflow actions & visibility filter

Status: completed locally, not deployed/restarted.

Implemented:
- Additive/backward-compatible migration `services/core-api/db/migrations/110_bank_soal_workflow_actions.sql` expands `cbt_questions.workflow_status` to accept `submitted`, `revision_needed`, `reviewed`, `published`, and `archived` while preserving legacy `review`.
- New `bank_soal_question_workflow_events` audit table plus indexes for question/actor/action timelines.
- Backend workflow actions now support `submit_for_review`, `request_revision`, `mark_reviewed`, `reject`, `approve`, `publish`, and `archive`; legacy action aliases `submit_review` and `return_revision` remain supported.
- Service guardrails added: author/admin submit only; reviewer/approver scope checks; reviewer cannot review own question unless admin; author cannot approve/publish own question unless admin; used/published in-place edit guards remain intact.
- List/summary visibility and answer-key redaction are compatibility-aware for legacy `review` vs new `submitted`, reviewer/approver scopes, and package operators seeing approved/published items.
- Existing handler/service tests updated for new workflow transition semantics and audit event writes.

Verified:
- Migration dry-run in transaction: PASS.
- `npm --prefix apps/web-admin run check`: PASS.
- `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`: PASS.
- `go test ./internal/handler ./internal/service ./internal/repository/postgres`: PASS.
- `go build -o /tmp/core-api-bank-soal-workflow-sprint2 ./cmd/api`: PASS.

Notes:
- No deploy/restart PM2 was performed.
- Production deployment must run migration `110_bank_soal_workflow_actions.sql` before starting a binary that references `bank_soal_question_workflow_events`.

### Sprint 3 — Review & approval UI

Status: completed locally, not deployed/restarted.

Implemented:
- Composer/list/detail UI now uses the new workflow action names: `submit_for_review`, `request_revision`, `mark_reviewed`, `reject`, `approve`, `publish`, and `archive` while keeping legacy `review` display compatibility.
- Review queues and catalog quick filters show clearer status groups for pending review, revision needed, reviewed, approved/published, and archived items.
- Read-only detail/reviewer panels send `mark_reviewed` for “Tandai Layak” and `request_revision` for “Minta Revisi”, matching Sprint 2 backend semantics.
- Bulk workflow toolbar supports reviewer/approver actions and skips ineligible selected rows safely.
- Workflow status labels/classes were expanded for `submitted`, `revision_needed`, `reviewed`, `published`, and `archived`.

Verified:
- `npm --prefix apps/web-admin run check`: PASS.
- `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`: PASS.
- `go test ./internal/handler ./internal/service ./internal/repository/postgres`: PASS.
- `go build -o /tmp/core-api-bank-soal-workflow-sprint3 ./cmd/api`: PASS.

Notes:
- No deploy/restart PM2 was performed.
- Sprint 3 is UI/BFF-compatible with Sprint 2 backend; production still needs migrations 109 and 110 before deploying the new binaries.

---

## Deployment runbook setelah implementasi selesai

Do not deploy until explicitly requested.

Jika diminta deploy ke production/semi-production:

0. Pastikan Sprint 0 audit sudah ada dan go/no-go = **GO**.
1. Backup DB dulu dan simpan checksum:
   ```bash
   mkdir -p /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql
   # Gunakan env existing; jangan tampilkan DATABASE_URL ke log/chat.
   pg_dump "$DATABASE_URL" -Fc -f /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-bank-soal-role-workflow-$(date +%Y%m%d-%H%M%S).dump
   sha256sum /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-bank-soal-role-workflow-*.dump | tail -1
   ```
2. Jalankan migration secara urut dan berhenti jika ada error:
   - permission/scope migration additive
   - workflow status/audit migration backwards-compatible
   - versioning migration jika Sprint 5 versioning ikut selesai
3. Setelah migration, jalankan smoke SQL non-destructive:
   - count questions by workflow_status
   - count role permissions
   - count reviewer scopes
   - count package questions by question workflow_status
4. Generate sqlc dan build core-api:
   ```bash
   cd services/core-api
   /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   go test ./internal/handler ./internal/service ./internal/repository/postgres
   go build -o bin/api ./cmd/api
   ```
5. Build web-admin:
   ```bash
   npm --prefix apps/web-admin run build
   ```
6. Restart PM2 setelah build web-admin dan backend siap. Karena web-admin SvelteKit bisa menyimpan manifest lama, restart web-admin wajib langsung setelah build:
   ```bash
   pm2 restart mtsn2kolut-core-api --update-env
   pm2 restart mtsn2kolut-web-admin --update-env
   ```
7. Health check:
   ```bash
   curl -fsS http://127.0.0.1:8080/health
   ```
8. Smoke test UI pakai akun admin dulu:
   - `/bank-soal`
   - `/bank-soal/tambah`
   - `/bank-soal/verifikasi`
   - `/bank-soal/penerbitan`
   - `/bank-soal/pengaturan`
   - `/bank-soal/daftar`
   - pilih soal detail `/bank-soal/soal/[id]`
9. Smoke test role terbatas:
   - guru biasa masih bisa lihat/buat soal sendiri.
   - reviewer hanya melihat scope-nya.
   - operator paket hanya melihat soal approved/published.
10. Jika smoke test gagal karena akses terlalu ketat:
   - rollback behavior via feature flag/config jika tersedia, atau
   - restore binary sebelumnya dan jangan rollback DB additive kecuali benar-benar perlu.
   - laporkan path backup dan error tanpa credential.

---

## Implementation notes

### Sprint 4 completed

- Backend package assignment hardened for new package question inserts/replacements:
  - `AddCbtPackageQuestion` now refuses locked packages, subject mismatch, archived questions, unsafe event scope, and questions that are not `approved`/`published` by workflow or already `published` by legacy status.
  - Service validation now rejects `draft`, `submitted`, `revision_needed`, `rejected`, and `archived` workflow questions before package rows are replaced, preserving existing package rows unless a user explicitly edits a package.
  - Existing package rows are not deleted or migrated in this sprint; this keeps rollout report-first/backward-compatible for historical packages.
- Answer key/rubric visibility hardened:
  - Detail endpoint redacts both `answer_key` and `rubric_html` unless actor is admin/read-all, author, event reviewer/panitia, or scoped reviewer/approver with matching Bank Soal permission and scope.
  - List/filter queries also redact `answer_key` and `rubric_html` using actor permission/scope flags.
- Tests added/updated for:
  - rejecting unsafe workflow questions in official package creation,
  - accepting approved/published workflow questions,
  - redacting answer/rubric for plain guru on other author questions,
  - allowing scoped reviewer/approver to view answer/rubric in review/approval statuses.
- Validation PASS:
  - `npm --prefix apps/web-admin run check`
  - `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`
  - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
  - `go build -o /tmp/core-api-bank-soal-workflow-sprint4 ./cmd/api`
- No deploy/restart PM2 was performed.

### Sprint 5 completed

- Backend summary endpoint expanded with role-workflow counters:
  - `my_draft`
  - `my_review_waiting`
  - `revision_needed`
  - `approval_waiting`
  - `package_ready`
  - `missing_metadata`
- Workflow event timeline endpoint added for detail soal:
  - backend `GET /api/cbt/questions/{id}/workflow-events`
  - SvelteKit BFF `/api/bank-soal/questions/[id]/workflow-events`
  - detail/preview timeline now shows action, status transition, actor, note, reviewer/approver metadata, and timestamp.
- Dashboard health model exposes role-workflow cards for Bank Soal operational reporting.
- Version-safe edit policy integrated with the existing detail flow: published/used/old-version questions remain read-only and create revision flow is preserved.
- SOP/contract documented in `docs/contracts/bank-soal-role-workflow.md`.
- Validation PASS:
  - `npm --prefix apps/web-admin run check`
  - `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`
  - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
  - `go build -o /tmp/core-api-bank-soal-sprint5 ./cmd/api`
- No deploy/restart PM2 was performed.

---

## Acceptance criteria

- Guru tidak bisa melihat draft/submitted guru lain kecuali dia reviewer/approver scope terkait.
- Guru bisa membuat, menyimpan draft, dan submit soal.
- Reviewer mapel bisa memberi keputusan review sesuai scope.
- Reviewer tidak bisa review soal sendiri tanpa admin override.
- Approver bisa approve/publish setelah review.
- Operator paket hanya bisa memakai soal approved/published.
- Kunci jawaban tidak bocor ke guru lain sebelum status aman.
- Semua perubahan workflow tercatat audit event.
- Published/used question tidak diedit in-place; gunakan versioning/revision flow.
- `npm --prefix apps/web-admin run check` PASS.
- `sqlc generate`, Go tests, dan Go build PASS.
- Tidak ada credential/API key/password ditampilkan atau tersimpan.

---

## Recommended execution order

0. Sprint 0: Production audit & safe rollout gate.
1. Sprint 1: RBAC & reviewer scope additive only, no revoke without approval.
2. Sprint 2: Backend workflow actions & visibility filter in compatibility/soft-enforcement mode.
3. Sprint 3: Review/approval UI, still compatible with existing status/data.
4. Sprint 4: Package CBT hardening & answer-key visibility, report-only for existing packages first.
5. Sprint 5: Audit/reporting + integrate question versioning.

Plan ini sengaja tidak langsung deploy. Setelah tiap sprint selesai: validate, review diff, commit. Deploy production hanya setelah Bapak minta eksplisit. Untuk production yang sudah berisi data, Sprint 0 wajib selesai sebelum migration/behavior change apa pun.
