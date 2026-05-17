# Opsi B — Operational Redesign Layer Implementation Plan

> **For Hermes:** Plan mode only. Do not implement this plan until user explicitly approves. Use `subagent-driven-development` when executing, task-by-task with review after every sprint.

**Goal:** Merapikan UI/UX Asesmen CBT dan Bank Soal agar lebih sederhana, rapi, operasional, mobile-friendly, tetapi tetap lengkap tanpa rewrite backend besar.

**Architecture:** Opsi B tidak menghapus route/backend lama. Redesign dilakukan sebagai lapisan UI operasional: fitur lama tetap ada, tetapi ditampilkan ulang dalam alur kerja madrasah: Hari Ini → Persiapan → Monitor → Hasil & BA → Arsip. SvelteKit tetap sebagai UI/BFF; tidak boleh akses database langsung dari frontend.

**Tech Stack:** SvelteKit/Svelte 5, existing shadcn-style components, Go core-api sebagai backend, PostgreSQL/sqlc, PM2 deploy hanya setelah approval eksplisit.

---

## 1. Prinsip Opsi B

Opsi B adalah **Operational Redesign Layer**.

Artinya:

- Tidak rewrite besar.
- Tidak hapus route lama.
- Tidak ubah backend kecuali benar-benar diperlukan.
- UI dipaketkan ulang berdasarkan alur kerja sekolah/madrasah.
- Fitur teknis tetap ada, tetapi tidak ditampilkan semua di layar utama.
- Fitur lanjutan masuk ke:
  - `Mode Lengkap`,
  - drawer/detail,
  - accordion,
  - dropdown `Fitur Lanjutan`,
  - halaman pengaturan.

Tujuan rasa UI:

- lebih tenang,
- tidak padat,
- mudah dipahami guru/panitia,
- tetap kuat untuk admin/operator,
- aman untuk hari-H ujian.

---

## 2. Target Struktur Akhir

## 2.1 Asesmen CBT

Struktur final yang disarankan:

1. **Hari Ini / Dashboard CBT**
   - Apa ujian hari ini?
   - Sesi mana sedang berjalan?
   - Peserta/ruang mana bermasalah?
   - BA/hasil apa yang belum selesai?

2. **Persiapan Ujian**
   - Buat/cek kegiatan.
   - Paket soal.
   - Peserta.
   - Sesi & ruang.
   - Token/kartu.
   - Checklist kesiapan.

3. **Monitor Ujian**
   - Sesi aktif.
   - Pengawasan ruang.
   - Peserta bermasalah.
   - Insiden.

4. **Hasil & Berita Acara**
   - Jawaban tersimpan.
   - Koreksi/nilai.
   - BA sesi/ruang.
   - Unduh rekap.

5. **Arsip**
   - Kegiatan selesai.
   - Dokumen final.
   - Checklist arsip.

6. **Aplikasi Siswa**
   - Panduan.
   - Release APK.
   - Matrix perangkat.
   - Troubleshooting.

## 2.2 Bank Soal

Struktur final yang disarankan:

1. **Dashboard Bank Soal**
2. **Kelola Soal**
3. **Review & Terbitkan**
4. **Mutu Soal**
5. **Pengaturan**

Catatan istilah:

- `Analisis Butir` dinaungi dulu oleh `Mutu Soal` jika data item-analysis belum lengkap.
- `Tambah Soal` menjadi primary action, bukan harus selalu sidebar item.
- `Impor`, `Mapel & KD`, `Scope Reviewer` masuk ke Pengaturan/aksi pendukung sesuai role.

---

## 3. Rules Anti-Ramai

Rules ini menjadi acceptance criteria setiap sprint:

1. Satu layar = satu tujuan utama.
2. Satu primary action per layar.
3. Summary card maksimal 3 sebelum konten utama.
4. Tabs maksimal 5.
5. Right rail optional, bukan selalu tampil.
6. Table untuk data banyak; card untuk ringkasan/workflow.
7. Drawer untuk detail sekunder.
8. Fitur teknis masuk `Mode Lengkap` atau `Fitur Lanjutan`.
9. Proctoring mobile harus card-first.
10. Accessibility masuk sejak sprint pertama, bukan polish akhir.

---

## 4. File/Area Existing yang Relevan

### 4.1 Asesmen CBT

Likely files:

- `apps/web-admin/src/routes/asesmen/+page.svelte`
- `apps/web-admin/src/routes/asesmen/persiapan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/hasil/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/archive/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/exam-cards/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/members/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/print-pack/+page.svelte`
- `apps/web-admin/src/routes/asesmen/aplikasi-siswa/+page.svelte`
- `apps/web-admin/src/routes/asesmen/aplikasi-siswa/release/+page.svelte`
- `apps/web-admin/src/routes/asesmen/aplikasi-siswa/matrix/+page.svelte`

### 4.2 Bank Soal

Likely files:

- `apps/web-admin/src/routes/bank-soal/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/daftar/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/tambah/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/soal/[id]/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/verifikasi/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/penerbitan/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/analisis-butir/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/impor/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/mapel-kd/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/pengaturan/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/PublishBankSoalPage.svelte`

### 4.3 Sidebar/Navigation

Likely files:

- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- `apps/web-admin/src/lib/components/Sidebar.svelte`
- route access tests if available:
  - `apps/web-admin/src/lib/server/route-access.test.ts`
  - `apps/web-admin/src/lib/server/assessment-route-contract.test.ts`

---

## 5. Layout System Minimum

Jangan membuat komponen terlalu banyak. Buat minimum reusable layer:

### 5.1 `PageHeader`

Likely create:

- `apps/web-admin/src/lib/components/ops/PageHeader.svelte`

Fungsi:

- title,
- subtitle,
- breadcrumb/context optional,
- primary action,
- secondary actions dalam dropdown jika lebih dari 2 aksi.

Rules:

- Primary action maksimal 1.
- Secondary action maksimal 1 terlihat langsung; lainnya dropdown.

### 5.2 `ContextStrip`

Likely create:

- `apps/web-admin/src/lib/components/ops/ContextStrip.svelte`

Fungsi:

- tahun ajaran,
- semester,
- kegiatan aktif,
- sesi aktif,
- status singkat.

Rules:

- Tampil hanya jika konteks membantu keputusan.
- Jangan membuat context bar panjang.

### 5.3 `WorkflowCard`

Likely create:

- `apps/web-admin/src/lib/components/ops/WorkflowCard.svelte`

Fungsi:

- card pintu masuk workflow seperti Persiapan, Monitor, Hasil, Arsip.

Isi minimal:

- judul,
- status singkat,
- next action,
- badge blocker jika ada.

### 5.4 `MetricCard`

Likely create:

- `apps/web-admin/src/lib/components/ops/MetricCard.svelte`

Rules:

- Maksimal 3 cards di area atas.
- Warna harus konsisten dan tidak ramai.

### 5.5 `EntityTabs`

Likely create:

- `apps/web-admin/src/lib/components/ops/EntityTabs.svelte`

Rules:

- Maksimal 5 tabs.
- Support badge/count.
- Keyboard/focus accessible.

### 5.6 `BlockerPanel`

Likely create:

- `apps/web-admin/src/lib/components/ops/BlockerPanel.svelte`

Rules:

- Hanya tampil kalau ada blocker.
- Jangan tampil sebagai card kosong permanen.

### 5.7 `InfoDrawer` / `StatusRail`

Likely create:

- `apps/web-admin/src/lib/components/ops/StatusRail.svelte`
- optional later: `apps/web-admin/src/lib/components/ops/InfoDrawer.svelte`

Rules:

- Desktop: optional side rail.
- Mobile: drawer/accordion.
- Tidak memuat konten panjang.

### 5.8 `ActionBar`

Likely create:

- `apps/web-admin/src/lib/components/ops/ActionBar.svelte`

Rules:

- Hanya untuk save, approval, bulk action.
- Tidak semua halaman punya sticky action.

---

## 6. Sprint Breakdown

## Sprint 0 — Read-only Visual Audit & Safety Gate

**Objective:** Pastikan implementasi tidak asal tebak, punya baseline visual dan route map sebelum edit.

**Files:**

- Read only:
  - `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
  - `apps/web-admin/src/routes/asesmen/**/+page.svelte`
  - `apps/web-admin/src/routes/bank-soal/**/+page.svelte`

**Steps:**

1. Ambil peta route Asesmen dan Bank Soal.
2. Catat halaman yang paling ramai:
   - detail kegiatan,
   - detail sesi,
   - pengawasan ruang,
   - bank soal dashboard/list/editor.
3. Catat istilah teknis yang harus diganti:
   - Event → Kegiatan Asesmen.
   - Readiness → Kesiapan.
   - Proctoring → Pengawasan Ujian.
   - Analisis Butir → Mutu Soal jika belum item-analysis penuh.
4. Buat screenshot/current-state evidence jika perlu memakai browser tool.
5. Tidak mengedit code.

**Verification:**

- Tidak ada diff UI/code selain dokumen plan/audit jika dibuat.
- `git status --short` hanya menampilkan plan/audit yang disengaja.

**Commit:**

- Tidak perlu commit jika hanya plan mode.

---

## Sprint 1 — Shared Operational Shell + Hari Ini + Detail Kegiatan Simple

**Objective:** Rasa rapi langsung terasa dengan shared shell dan penyederhanaan halaman Asesmen paling penting.

### Task 1.1 — Create minimal operational components

**Files:**

- Create:
  - `apps/web-admin/src/lib/components/ops/PageHeader.svelte`
  - `apps/web-admin/src/lib/components/ops/ContextStrip.svelte`
  - `apps/web-admin/src/lib/components/ops/WorkflowCard.svelte`
  - `apps/web-admin/src/lib/components/ops/MetricCard.svelte`
  - `apps/web-admin/src/lib/components/ops/EntityTabs.svelte`
  - `apps/web-admin/src/lib/components/ops/BlockerPanel.svelte`

**Acceptance Criteria:**

- Komponen visual minimal, tidak dekoratif berlebihan.
- Bisa dipakai tanpa backend baru.
- Props sederhana.
- Keyboard focus terlihat.
- Semua button non-submit memakai `type="button"`.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 1.2 — Redesign `/asesmen` as Hari Ini / Dashboard CBT

**Files:**

- Modify:
  - `apps/web-admin/src/routes/asesmen/+page.svelte`

**Layout Target:**

1. Header:
   - Title: `Hari Ini / Dashboard CBT` atau `Asesmen CBT`.
   - Subtitle: status persiapan, pelaksanaan, hasil, arsip.
   - Primary action: `Buat Kegiatan` atau `Buka Monitor Hari Ini`.

2. Top metrics maksimal 3:
   - Ujian hari ini.
   - Sesi sedang/akan berjalan.
   - Perlu tindakan.

3. Workflow cards:
   - Persiapan Ujian.
   - Monitor Ujian.
   - Hasil & BA.
   - Arsip.

4. Quick actions kecil:
   - Buat Kegiatan.
   - Cek Kesiapan.
   - Buka Monitor.
   - Lihat Hasil.

**Hide/Move:**

- Jangan tampilkan daftar route kecil terlalu banyak.
- Jangan jadikan semua fitur card sama besar.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 1.3 — Simplify Detail Kegiatan to 5 areas

**Files:**

- Modify:
  - `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`

**Layout Target:**

1. Header compact:
   - nama kegiatan,
   - status,
   - tahun ajaran/semester,
   - tanggal,
   - primary action sesuai status.

2. Metrics maksimal 3:
   - Kesiapan.
   - Peserta/Sesi.
   - Hasil/BA atau blocker.

3. Tabs maksimal 5:
   - Ringkasan.
   - Persiapan.
   - Pelaksanaan.
   - Hasil & BA.
   - Arsip.

4. Ringkasan tab:
   - next action,
   - blocker utama,
   - progress per tahap.

5. Persiapan tab:
   - paket soal,
   - peserta,
   - ruang/sesi,
   - token/kartu,
   - checklist SOP.

6. Pelaksanaan tab:
   - sesi berjalan,
   - link monitor,
   - peserta bermasalah,
   - insiden.

7. Hasil & BA tab:
   - submit/jawaban,
   - koreksi/nilai,
   - BA sesi/ruang.

8. Arsip tab:
   - dokumen final,
   - status pengesahan,
   - audit ringkas.

**Hide/Move:**

- `Pengesahan SOP` tidak menjadi tab utama sendiri; masuk Persiapan atau Arsip sesuai konteks.
- Right rail hanya muncul jika ada blocker/quick action penting.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 1.4 — Sprint 1 QA

**Commands:**

```bash
npm --prefix apps/web-admin run check
```

Optional if route tests exist:

```bash
npm --prefix apps/web-admin run test:unit -- src/lib/server/assessment-route-contract.test.ts
```

**Manual Smoke:**

- `/asesmen`
- `/asesmen/kegiatan`
- one existing `/asesmen/kegiatan/[id]` if data/route available.

**Commit after Sprint 1:**

```bash
git add apps/web-admin/src/lib/components/ops apps/web-admin/src/routes/asesmen

git commit -m "feat(ui): add operational assessment shell"
```

---

## Sprint 2 — Monitor Ujian Mobile-first

**Objective:** Hari-H ujian menjadi sederhana dan aman untuk pengawas/operator.

### Task 2.1 — Simplify Detail Sesi to 4 areas

**Files:**

- Modify:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`

**Layout Target:**

Tabs/sections maksimal 4:

1. Monitor.
2. Peserta.
3. Masalah/Insiden.
4. Hasil & BA.

**Top Summary maksimal 3:**

- Peserta hadir/total.
- Progress submit.
- Masalah aktif.

**Hide/Move:**

- Data teknis panjang masuk drawer/Mode Lengkap.
- Jangan semua tabel muncul di atas.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 2.2 — Make room proctoring exception-first

**Files:**

- Modify:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Layout Target:**

Default Mode Sederhana:

1. Sticky compact header:
   - sesi,
   - ruang,
   - waktu,
   - status koneksi,
   - token ruang only if safe context.

2. Status banner:
   - Butuh tindakan,
   - Terkunci,
   - Terputus,
   - Belum kirim.

3. Filter chips:
   - Butuh Tindakan.
   - Terkunci.
   - Terputus.
   - Belum Kirim.
   - Semua.

4. Main list:
   - peserta bermasalah dulu.
   - HP/tablet: card list.
   - desktop: table/list compact.

5. Actions:
   - Periksa.
   - Kirim Peringatan.
   - Instruksi Masuk Ulang.
   - Buka Kunci.

Mode Lengkap:

- semua peserta,
- event log,
- device detail,
- technical state,
- export evidence.

**Safety:**

- Destructive/rare action wajib confirmation dialog.
- Token tetap masked dalam BA/arsip.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 2.3 — Accessibility and mobile gate for monitor

**Files:**

- Modify the same proctoring/sesi files as needed.

**Acceptance Criteria:**

- Button target minimal layak sentuh.
- Button non-submit memakai `type="button"`.
- Filter/tabs punya focus visible.
- Status tidak hanya warna; ada teks.
- Critical status tidak spam screen reader.
- Dialog destructive punya title/deskripsi jelas.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Manual viewport check:

- mobile width,
- tablet width,
- desktop width.

### Task 2.4 — Sprint 2 QA

**Commands:**

```bash
npm --prefix apps/web-admin run check
```

**Manual Smoke:**

- `/asesmen/pelaksanaan`
- `/asesmen/sesi/[id]`
- `/asesmen/sesi/[id]/proctoring`
- `/asesmen/sesi/[id]/rooms/[rid]/proctoring`

**Commit after Sprint 2:**

```bash
git add apps/web-admin/src/routes/asesmen/sesi

git commit -m "feat(ui): simplify exam monitoring experience"
```

---

## Sprint 3 — Bank Soal Role Cleanup

**Objective:** Bank Soal terasa sederhana bagi guru, tetapi tetap kuat untuk admin/reviewer.

### Task 3.1 — Redesign Bank Soal Dashboard

**Files:**

- Modify:
  - `apps/web-admin/src/routes/bank-soal/+page.svelte`

**Layout Target:**

Header:

- Title: `Bank Soal`.
- Primary action: `Tambah Soal`.

Top metrics maksimal 3:

- Draft saya / Soal saya.
- Perlu revisi / Menunggu review.
- Soal terbit / Mutu perlu dicek.

Workflow cards:

- Kelola Soal.
- Review & Terbitkan.
- Mutu Soal.
- Pengaturan.

Role-aware, but same structure:

- Guru melihat prioritas draft, revisi, tambah soal.
- Admin/reviewer melihat prioritas review, penerbitan, mutu.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 3.2 — Simplify Kelola Soal / list

**Files:**

- Modify:
  - `apps/web-admin/src/routes/bank-soal/daftar/+page.svelte`
  - `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Layout Target:**

- Search/filter compact.
- Table/list for data many.
- Row action concise.
- Detail preview via drawer if existing component allows.
- `Tambah Soal` and `Impor` as actions, not noisy sections.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 3.3 — Simplify editor/composer default layout

**Files:**

- Modify:
  - `apps/web-admin/src/routes/bank-soal/tambah/+page.svelte`
  - `apps/web-admin/src/routes/bank-soal/soal/[id]/+page.svelte`
  - `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`

**Layout Target:**

Desktop/laptop default:

- 2-column:
  - main editor,
  - collapsible side panel for preview/checklist.

Large desktop optional:

- 3-column only when space allows.

Mobile:

- stepper:
  1. Metadata.
  2. Soal.
  3. Jawaban.
  4. Pembahasan/Rubrik.
  5. Preview & Kirim Review.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 3.4 — Review & Terbitkan as queue

**Files:**

- Modify:
  - `apps/web-admin/src/routes/bank-soal/verifikasi/+page.svelte`
  - `apps/web-admin/src/routes/bank-soal/penerbitan/+page.svelte`
  - `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`
  - `apps/web-admin/src/routes/bank-soal/_components/PublishBankSoalPage.svelte`

**Layout Target:**

Tabs/queues:

- Menunggu Review.
- Perlu Revisi.
- Siap Terbit.
- Terbit.

Detail via drawer:

- preview soal,
- checklist review,
- catatan reviewer,
- approve/reject/request revision.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 3.5 — Rename/position Mutu Soal

**Files:**

- Modify:
  - `apps/web-admin/src/routes/bank-soal/analisis-butir/+page.svelte`
  - potentially `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts` later in Sprint 4.

**Layout Target:**

If item-analysis data is not fully available:

- UI label: `Mutu Soal`.
- Content:
  - coverage mapel/KD,
  - kelengkapan metadata,
  - distribusi tipe soal,
  - soal belum review,
  - indikator kualitas dasar.

If item-analysis exists later:

- Add tab `Analisis Butir` inside `Mutu Soal`.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: PASS.

### Task 3.6 — Sprint 3 QA

**Commands:**

```bash
npm --prefix apps/web-admin run check
```

Optional route tests:

```bash
npm --prefix apps/web-admin run test:unit -- src/lib/server/cbt-backend-paths.test.ts
```

**Manual Smoke:**

- `/bank-soal`
- `/bank-soal/daftar`
- `/bank-soal/tambah`
- `/bank-soal/verifikasi`
- `/bank-soal/penerbitan`
- `/bank-soal/analisis-butir`

**Commit after Sprint 3:**

```bash
git add apps/web-admin/src/routes/bank-soal

git commit -m "feat(ui): simplify bank soal workflows"
```

---

## Sprint 4 — Sidebar Cleanup + Polish

**Objective:** Setelah layout dalam stabil, sidebar baru diringkas agar tidak menampilkan semua fitur teknis.

### Task 4.1 — Cleanup Asesmen sidebar

**Files:**

- Modify:
  - `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

**Target Asesmen Sidebar:**

- Hari Ini / Dashboard CBT.
- Persiapan Ujian.
- Monitor Ujian.
- Hasil & BA.
- Arsip.
- Aplikasi Siswa.

**Move out from sidebar main:**

- Paket Soal as standalone for non-admin if too technical.
- Route members/archive/minutes/print-pack from main sidebar.
- Proctoring technical routes.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Optional:

```bash
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/assessment-route-contract.test.ts
```

### Task 4.2 — Cleanup Bank Soal sidebar

**Files:**

- Modify:
  - `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

**Target Bank Soal Sidebar:**

- Dashboard Bank Soal.
- Kelola Soal.
- Review & Terbitkan.
- Mutu Soal.
- Pengaturan.

**Move out from sidebar main:**

- Tambah Soal → primary action.
- Impor Soal → action in Kelola/Pengaturan.
- Mapel & KD → Pengaturan.
- Penerbitan maybe inside Review & Terbitkan unless admin-specific menu required.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

### Task 4.3 — Empty/loading/error state polish

**Files:**

- Modify pages touched in Sprints 1–4 as needed.

**Acceptance Criteria:**

- Empty state has next action.
- Loading not visually noisy.
- Error message human-readable.
- Token/secrets not shown.
- Dangerous actions confirmed.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

### Task 4.4 — Final build/smoke gate

**Commands:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

Expected: PASS.

If backend untouched:

- No Go build required.

If backend touched unexpectedly:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-ops-redesign ./cmd/api
```

Expected: PASS.

**Commit after Sprint 4:**

```bash
git add apps/web-admin/src/lib/components/sidebar apps/web-admin/src/routes

git commit -m "feat(ui): streamline assessment and question bank navigation"
```

---

## 7. Deployment Plan — Only After Approval

Do not deploy/restart until user explicitly says deploy.

If frontend-only:

1. Review diff:

```bash
git status --short
git diff --stat
```

2. Run verification:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

3. Restart PM2 web-admin only after approval.

Important project note:

- After any `npm --prefix apps/web-admin run build`, restart PM2 web-admin immediately if deploying to live, otherwise old manifest/chunk mismatch can cause 500 errors.

If backend/migration becomes necessary:

- Stop and ask approval again.
- Prepare backup if production data affected.
- Run backend validation separately.

---

## 8. Risk & Mitigation

### Risk 1 — User bingung karena menu berubah

Mitigation:

- Route lama tetap ada.
- Label lama bisa muncul sebagai secondary text sementara.
- Breadcrumb/context strip membantu orientasi.

### Risk 2 — Halaman tetap ramai karena terlalu banyak components

Mitigation:

- Enforce anti-ramai rules.
- Metrics maksimal 3.
- Tabs maksimal 5.
- Right rail optional.

### Risk 3 — Mobile proctoring tidak nyaman

Mitigation:

- Card-first on HP/tablet.
- Table lengkap hanya Mode Lengkap.
- Touch target cukup.

### Risk 4 — Sidebar cleanup mengganggu role/permission

Mitigation:

- Cleanup dilakukan terakhir.
- Route access tests bila tersedia.
- Tidak menghapus route, hanya mengubah exposure/navigasi.

### Risk 5 — Svelte type/check errors karena halaman besar

Mitigation:

- Sprint kecil.
- `npm --prefix apps/web-admin run check` setelah tiap task besar.
- Commit per sprint.

---

## 9. Acceptance Criteria Final

Plan dianggap selesai jika:

- `/asesmen` terasa sebagai pintu masuk operasional, bukan kumpulan link.
- Detail Kegiatan punya maksimal 5 area kerja.
- Detail Sesi punya maksimal 4 area kerja.
- Proctoring default Mode Sederhana dan exception-first.
- Bank Soal punya workflow jelas: Dashboard, Kelola, Review & Terbitkan, Mutu, Pengaturan.
- Sidebar lebih pendek setelah layout dalam stabil.
- `npm --prefix apps/web-admin run check` PASS.
- `npm --prefix apps/web-admin run build` PASS sebelum deploy.
- Tidak ada backend/DB change kecuali disetujui.
- Tidak ada deploy/restart tanpa approval.

---

## 10. Open Decisions untuk Bapak

Sebelum implementasi, Bapak bisa pelajari dan pilih:

1. Nama halaman utama Asesmen:
   - `Hari Ini`
   - atau `Dashboard CBT`
   - atau `Asesmen CBT`

2. Detail Kegiatan:
   - setuju 5 area: Ringkasan, Persiapan, Pelaksanaan, Hasil & BA, Arsip?

3. Bank Soal:
   - setuju `Analisis Butir` dinaungi sebagai `Mutu Soal` dulu?

4. Sidebar:
   - cleanup dilakukan setelah 3 sprint inti, bukan di awal?

5. Proctoring:
   - setuju HP/tablet default card-list, table lengkap hanya Mode Lengkap?

---

## 11. Rekomendasi Eksekusi

Saya rekomendasikan urutan:

1. **Sprint 0:** audit visual read-only.
2. **Sprint 1:** shared shell + `/asesmen` + detail kegiatan 5 area.
3. **Sprint 2:** monitor/proctoring mobile-first.
4. **Sprint 3:** Bank Soal role cleanup.
5. **Sprint 4:** sidebar cleanup + polish + build.

Jangan mulai dari sidebar dulu. Mulai dari isi halaman yang paling terasa ramai, lalu sidebar disesuaikan setelah pola dalam stabil.
