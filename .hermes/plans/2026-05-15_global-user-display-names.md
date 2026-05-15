# Global User Display Names Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Seluruh UI aplikasi MTsN 2 Kolut menampilkan nama manusiawi pengguna/orang (`display_name`, nama pegawai, nama siswa, nama orang tua) alih-alih UUID/internal ID/internal username, kecuali di layar debug/admin teknis.

**Architecture:** Backend tetap menyimpan dan menerima identifier stabil (`id`, `username`) untuk permission, audit, relasi, dan filter. API menambahkan field label manusiawi (`*_display_name`, `*_name`, `*_label`) berdampingan dengan identifier teknis. SvelteKit/web-admin wajib memakai helper display-name dengan fallback aman, dan hanya menampilkan ID di detail teknis/copy/debug.

**Tech Stack:** Go core-api, PostgreSQL/sqlc, SvelteKit web-admin, PM2 deploy.

---

## Prinsip Global

1. **Identifier teknis tetap ada di backend:** UUID/username tidak dihapus dari DB atau kontrak internal.
2. **UI utama pakai nama:** gunakan `display_name`/`nama`/`name` bila tersedia.
3. **Fallback berurutan:** nama profil terkait → `users.display_name` → `username` → ID hanya jika benar-benar tidak ada label.
4. **ID hanya di mode teknis:** audit detail, copy ID, debug admin, dan log teknis boleh menampilkan ID dengan label jelas `ID internal`.
5. **Tidak broad rewrite:** rollout bertahap per modul, dengan validasi per sprint.

---

## Definition of Done Global

- Semua daftar, kartu, filter, dropdown, badge, dan drawer utama tidak menampilkan UUID/internal ID ketika ada nama.
- Endpoint yang mengembalikan actor/owner/assignee/creator/reviewer menyertakan label manusiawi.
- UI filter tetap memakai identifier stabil sebagai `value`, tapi teks option menampilkan nama.
- `npm --prefix apps/web-admin run check` PASS.
- `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml` PASS jika query diubah.
- `go test ./internal/handler ./internal/service ./internal/repository/postgres` PASS jika backend diubah.
- `go build -o /tmp/core-api-user-display-stageN ./cmd/api` PASS jika backend diubah.
- Build/restart PM2 hanya setelah validasi dan sesuai kebutuhan.

---

## Sprint 0 — Policy, Helper, dan Audit Otomatis

**Objective:** Membuat standar teknis agar sprint berikutnya konsisten dan mudah dicek.

**Files:**
- Already created: `docs/contracts/user-display-name-policy.md`
- Create: `apps/web-admin/src/lib/utils/display-name.ts`
- Create: `scripts/audit-user-id-display.mjs`
- Modify: `.hermes/plans/2026-05-15_global-user-display-names.md`

### Task 0.1: Buat helper display name frontend

**Objective:** Satu helper standar untuk memilih nama yang ditampilkan.

**Create:** `apps/web-admin/src/lib/utils/display-name.ts`

```ts
export type DisplayNameCandidate = {
  display_name?: string | null;
  name?: string | null;
  nama?: string | null;
  full_name?: string | null;
  username?: string | null;
  id?: string | null;
};

export function displayName(value: DisplayNameCandidate | string | null | undefined, fallback = 'Tidak diketahui') {
  if (typeof value === 'string') return value.trim() || fallback;
  if (!value) return fallback;
  return (
    value.display_name?.trim() ||
    value.nama?.trim() ||
    value.name?.trim() ||
    value.full_name?.trim() ||
    value.username?.trim() ||
    value.id?.trim() ||
    fallback
  );
}

export function actorDisplayName(prefix: string, row: Record<string, unknown>, fallback = 'Tidak diketahui') {
  return displayName({
    display_name: String(row[`${prefix}_display_name`] ?? ''),
    nama: String(row[`${prefix}_nama`] ?? ''),
    name: String(row[`${prefix}_name`] ?? ''),
    username: String(row[`${prefix}_username`] ?? ''),
    id: String(row[`${prefix}_id`] ?? '')
  }, fallback);
}
```

### Task 0.2: Tambahkan audit script kasar

**Objective:** Menemukan kandidat UI yang masih menampilkan ID/username mentah.

**Create:** `scripts/audit-user-id-display.mjs`

Script mencari pola di `.svelte`:
- `created_by`
- `updated_by`
- `user_id`
- `_id}`
- `author_username`
- `reviewer_username`
- `approver_username`
- `actor_username`

Expected output: daftar file + line number untuk triage manual.

### Task 0.3: Validasi Sprint 0

Run:

```bash
npm --prefix apps/web-admin run check
node scripts/audit-user-id-display.mjs
```

Expected:
- `svelte-check found 0 errors and 0 warnings`
- audit menghasilkan daftar kandidat, bukan necessarily kosong.

### Task 0.4: Commit Sprint 0

```bash
git add docs/contracts/user-display-name-policy.md apps/web-admin/src/lib/utils/display-name.ts scripts/audit-user-id-display.mjs .hermes/plans/2026-05-15_global-user-display-names.md AGENTS.md
git commit -m "chore(ui): add user display name policy and audit helper"
```

**Sprint 0 implementation result — 2026-05-15:**
- Added shared frontend helper `apps/web-admin/src/lib/utils/display-name.ts` with `displayName()`, `actorDisplayName()`, and `optionLabel()`.
- Added audit script `scripts/audit-user-id-display.mjs` to list candidate raw ID/username UI usage for staged cleanup.
- Added global AGENTS.md rule: user-facing UI must use human-readable labels and reserve internal IDs for explicit technical/debug/admin detail surfaces.
- Validation commands:
  - `npm --prefix apps/web-admin run check`
  - `node scripts/audit-user-id-display.mjs`

---

## Sprint 1 — CBT/Bank Soal/Asesmen Actor Names

**Objective:** Modul paling sering terlihat oleh guru/panitia tidak lagi menampilkan username/ID teknis untuk pembuat, reviewer, approver, panitia, pengawas.

**Files likely touched:**
- Existing: `services/core-api/db/queries/cbt_questions.sql`
- Existing: `services/core-api/internal/handler/cbt_question_serialize.go`
- Existing: `apps/web-admin/src/routes/asesmen/paket/[id]/+page.svelte`
- Audit and patch candidates under:
  - `apps/web-admin/src/routes/bank-soal/`
  - `apps/web-admin/src/routes/asesmen/`
  - `services/core-api/db/queries/*cbt*.sql`

### Task 1.1: Complete question actor display names

**Objective:** Pastikan semua response soal punya `author_display_name`, `reviewer_display_name`, `approver_display_name` bila query berbasis row list/detail/version.

**Steps:**
1. Audit `cbt_questions.sql` semua SELECT yang expose actor username.
2. Tambahkan join `users` dan `employees` bila belum ada.
3. Expose field display name di serializer.
4. Jalankan sqlc.

Run:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-user-display-stage1 ./cmd/api
```

### Task 1.2: Patch UI Bank Soal daftar/verifikasi/detail

**Objective:** Semua label pembuat/reviewer/approver pakai display name.

**Candidates:**
- `apps/web-admin/src/routes/bank-soal/daftar/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/verifikasi/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `apps/web-admin/src/routes/asesmen/paket/[id]/+page.svelte`

**Pattern:**
- Filter value: tetap `author_username`.
- Filter option text: `author_display_name`.
- Badge/card text: `displayName(...)`.

### Task 1.3: Patch Paket/Kegiatan/Sesi CBT actor names

**Objective:** Panitia/pengawas/session actor tampil sebagai nama.

**Candidates:**
- `apps/web-admin/src/routes/asesmen/kegiatan/`
- `apps/web-admin/src/routes/asesmen/sesi/`
- `apps/web-admin/src/routes/asesmen/pengawasan/`

Patch endpoint Go bila response hanya punya ID/username.

### Task 1.4: Validasi Sprint 1

Run:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-user-display-stage1 ./cmd/api
```

### Task 1.5: Commit/deploy Sprint 1

Commit:

```bash
git commit -m "feat(cbt): show user display names across assessment flows"
```

Deploy if requested/needed:
1. build backend binary and backup old binary.
2. restart `mtsn2kolut-core-api`.
3. build web-admin.
4. restart `mtsn2kolut-web-admin`.
5. smoke test `/health`, web root, protected route expected redirect/401.

**Sprint 1 implementation result — 2026-05-15:**
- Completed CBT/Bank Soal/Asesmen actor-name cleanup for the high-visibility flows.
- `cbt_questions.sql` now exposes `actor_display_name` for question timeline rows in addition to existing author/reviewer/approver display-name fields.
- Updated Go service/handler interfaces and tests for `ListCbtQuestionTimelineRow` after sqlc generation.
- Patched Bank Soal workspace/detail/review components to render `displayName(...)` for author, reviewer, approver, and timeline actors.
- Patched Asesmen event members page so member names and user dropdowns prefer employee/profile names; usernames remain secondary labels, internal IDs are hidden from primary UI.
- Validated:
  - `/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`
  - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
  - `go build -o /tmp/core-api-user-display-sprint1 ./cmd/api`
  - `npm --prefix apps/web-admin run check`
  - `npm --prefix apps/web-admin run build`
  - targeted audit for raw actor username/ID rendering in `bank-soal` and `asesmen` routes.
- Deployment intentionally deferred; deploy/restart only when requested.

---

## Sprint 2 — Settings, Users, RBAC, Audit Logs

**Objective:** Halaman admin pengguna tetap teknis bila perlu, tapi daftar utama dan audit event memakai nama manusiawi.

**Files likely touched:**
- `apps/web-admin/src/routes/settings/users/+page.svelte`
- `apps/web-admin/src/routes/settings/rbac/+page.svelte`
- `apps/web-admin/src/routes/settings/audit-logs/+page.svelte`
- `apps/web-admin/src/routes/settings/user-change-requests/+page.svelte`
- Go queries/handlers untuk audit/change requests bila hanya expose user ID.

### Task 2.1: Settings Users table

**Objective:** Tabel user menampilkan nama utama + username kecil/secondary, bukan ID.

UI pattern:

```svelte
<div class="font-medium">{displayName(user)}</div>
<div class="text-xs text-muted-foreground">@{user.username}</div>
```

UUID hanya di drawer detail teknis/copy.

### Task 2.2: RBAC assignments

**Objective:** Dropdown/daftar assignment role menampilkan nama pengguna.

Backend: endpoint role assignment harus include `user_display_name` bila belum.

### Task 2.3: Audit logs

**Objective:** Actor dan target tampil nama, detail ID tetap collapsible.

Pattern:
- Main row: `actor_display_name`.
- Secondary: action + module.
- Technical detail: `actor_username`, `actor_id`, `target_id` di accordion/copy.

### Task 2.4: Validasi dan commit Sprint 2

Run check/build + Go tests bila backend diubah.

Commit:

```bash
git commit -m "feat(settings): prefer display names for users and audit actors"
```

---

## Sprint 3 — Akademik, Siswa, Orang Tua, Pegawai

**Objective:** Modul operasional akademik tidak menampilkan ID rombel/siswa/pegawai/orang tua di UI utama.

**Files likely touched:**
- `apps/web-admin/src/routes/students/+page.svelte`
- `apps/web-admin/src/routes/parents/+page.svelte`
- `apps/web-admin/src/routes/employees/+page.svelte`
- `apps/web-admin/src/routes/akademik/rombel/+page.svelte`
- `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte`
- `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`

### Task 3.1: Student/parent selectors

**Objective:** Dropdown dan drawer siswa/orang tua memakai nama + kelas/NISN sebagai secondary text.

Pattern:
- Primary: `student.nama`
- Secondary: `NISN/NISM`, kelas, status.
- Jangan tampilkan UUID.

### Task 3.2: Employee/teacher selectors

**Objective:** Wali kelas, guru mapel, jadwal, beban guru pakai nama pegawai.

Pattern:
- Primary: `employee.nama`
- Secondary: `NIP` atau jabatan bila perlu.
- ID internal tidak ditampilkan.

### Task 3.3: Rombel/class labels

**Objective:** Rombel tampil `VII.A / Kelas VII A`, bukan class ID.

### Task 3.4: Validasi dan commit Sprint 3

Run:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

Jika backend berubah, tambah sqlc/test/build Go.

Commit:

```bash
git commit -m "feat(academic): replace internal ids with display labels"
```

---

## Sprint 4 — TU, Governance, Document Cycles, Arsip

**Objective:** Workflow surat/dokumen menampilkan nama pegawai/pemohon/penanggung jawab, bukan ID/username.

**Files likely touched:**
- `apps/web-admin/src/routes/tu/`
- `apps/web-admin/src/routes/governance/`
- `apps/web-admin/src/routes/document-cycles/`
- Go queries untuk created_by/updated_by/assigned_to bila diperlukan.

### Task 4.1: Surat masuk/keluar/arsip

Patch creator/assignee/recipient labels.

### Task 4.2: Governance action owners

Patch owner, reviewer, evidence uploader labels.

### Task 4.3: Document cycle reviewers/approvers

Patch reviewer/approver/submitter labels.

### Task 4.4: Validasi dan commit Sprint 4

Commit:

```bash
git commit -m "feat(tu): show display names in document workflows"
```

---

## Sprint 5 — Inventory, Notifications, Analytics, Misc Cleanup

**Objective:** Modul sisa dibersihkan agar konsisten seluruh app.

**Files likely touched:**
- `apps/web-admin/src/routes/inventory/`
- `apps/web-admin/src/routes/notifications/`
- `apps/web-admin/src/routes/settings/analytics/`
- `apps/web-admin/src/routes/library/`
- `apps/web-admin/src/routes/journal/`

### Task 5.1: Inventory owner/handler

Tampilkan nama pemegang/penanggung jawab.

### Task 5.2: Notifications actor/recipient

Tampilkan nama pengirim/penerima.

### Task 5.3: Analytics actor labels

ID tetap untuk event raw, dashboard ringkas pakai nama/label.

### Task 5.4: Final audit

Run:

```bash
node scripts/audit-user-id-display.mjs
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

Triage hasil audit:
- legitimate debug/admin teknis → whitelist/comment.
- UI utama → patch.

Commit:

```bash
git commit -m "chore(ui): finish global display-name cleanup"
```

---

## Deployment Runbook Per Sprint

Jika hanya frontend berubah:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
pm2 restart mtsn2kolut-web-admin --update-env
curl -fsS -o /tmp/web-root.html -w 'web_root=%{http_code}\n' http://127.0.0.1:8021/
```

Jika backend berubah:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-user-display-stageN ./cmd/api
cp -p bin/api bin/api.backup-user-display-stageN-$(date +%Y%m%d-%H%M%S)
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
curl -fsS http://127.0.0.1:8080/health
```

Then restart web-admin if built.

---

## Suggested Timeline

- **Sprint 0:** 1 sesi — helper + audit + standard.
- **Sprint 1:** 1–2 sesi — CBT/Bank Soal/Asesmen, paling prioritas karena user sedang memakai Paket Builder.
- **Sprint 2:** 1 sesi — Settings/RBAC/Audit.
- **Sprint 3:** 1–2 sesi — Akademik/Siswa/Pegawai/Orang tua.
- **Sprint 4:** 1–2 sesi — TU/Governance/Document Cycles.
- **Sprint 5:** 1 sesi — cleanup seluruh app + final audit.

---

## Acceptance Checklist

- [ ] Paket Builder: pembuat soal tampil nama.
- [ ] Bank Soal: pembuat/reviewer/approver tampil nama.
- [ ] Asesmen: panitia/pengawas/pemilik tampil nama.
- [ ] Settings: daftar user menampilkan nama + username secondary.
- [ ] Audit logs: actor/target tampil nama, ID hanya detail teknis.
- [ ] Akademik: siswa/pegawai/orang tua/rombel tampil label manusiawi.
- [ ] TU/Governance: assignee/reviewer/owner tampil nama.
- [ ] Inventory/Notification/Analytics: user-facing label tidak pakai UUID.
- [ ] Audit script tidak menemukan ID mentah pada UI utama, atau semua temuan sudah diklasifikasi debug/admin teknis.
