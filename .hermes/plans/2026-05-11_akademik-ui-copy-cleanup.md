# Akademik UI Copy Cleanup Implementation Plan

> **For Hermes:** Use subagent-driven-development or Codex pass-through to implement this plan task-by-task.

**Goal:** Membersihkan istilah developer/teknis dari UI Akademik dan menggantinya dengan bahasa operasional madrasah yang jelas untuk admin, staf, guru, dan operator.

**Architecture:** Tidak mengubah skema database dan tidak mengubah business logic. Fokus pada copywriting UI, struktur label, helper text, empty state, dialog confirmation, dan konsistensi istilah di halaman Akademik. Backend hanya disentuh bila response/error message domain perlu diterjemahkan agar lebih user-facing.

**Tech Stack:** SvelteKit 2/Svelte 5 web-admin, Go core-api, thin BFF proxy tetap dipertahankan.

---

## Prinsip Bahasa UI

### Dilarang untuk UI user-facing

Jangan tampilkan istilah developer berikut pada UI:

- sprint
- phase / fase teknis
- apply
- dry-run
- endpoint
- API
- payload
- commit
- deploy
- migration
- seed
- debug
- implementation
- rollout/rollover mentah tanpa penjelasan
- query
- JSON
- token jika maksudnya bukan token ujian/akses yang dipahami user

### Padanan bahasa operasional

| Istilah developer | Padanan user-facing |
|---|---|
| Sprint | Tahap / Pengembangan tahap ini / Fitur baru |
| Apply rollover | Terapkan kenaikan kelas |
| Preview rollover | Pratinjau kenaikan kelas |
| Dry-run import | Cek data sebelum impor |
| Import/export | Impor/Unduh data |
| Endpoint/API | Layanan sistem |
| Payload/JSON | Data yang dikirim |
| Commit/deploy | Pembaruan sistem |
| Migration | Penyesuaian struktur data |
| Seed | Pengisian data awal |
| Debug | Pemeriksaan masalah |
| Token rollover | Kode konfirmasi / kalimat konfirmasi |
| Matrix | Tabel penugasan |
| Dirty changes | Perubahan belum disimpan |

### Tone yang diinginkan

- Formal ringan, cocok untuk madrasah.
- Singkat dan langsung menjelaskan aksi.
- Hindari bahasa startup/SaaS/developer.
- Gunakan kata kerja operasional: `Simpan`, `Batalkan`, `Periksa`, `Terapkan`, `Unduh`, `Muat ulang`, `Pindahkan`, `Aktifkan`.
- Semua aksi berisiko harus menjelaskan dampak, bukan istilah teknis.

---

## Task 1: Audit semua copy user-facing Akademik

**Objective:** Membuat daftar semua label, heading, deskripsi, toast, dialog, empty state, dan error text di modul Akademik.

**Files:**
- Inspect: `apps/web-admin/src/routes/akademik/+page.svelte`
- Inspect: `apps/web-admin/src/routes/akademik/rombel/+page.svelte`
- Inspect: `apps/web-admin/src/routes/akademik/rombel/[id]/+page.svelte`
- Inspect: `apps/web-admin/src/routes/akademik/mapel/+page.svelte`
- Inspect: `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte`
- Inspect: `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`
- Inspect: `apps/web-admin/src/routes/akademik/tahun-ajaran/+page.svelte`
- Inspect: `apps/web-admin/src/routes/students/+page.svelte`
- Inspect optional backend errors: `services/core-api/internal/handler/academic.go`
- Inspect optional backend errors: `services/core-api/internal/service/academic*.go`

**Step 1:** Search istilah developer.

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
python3 - <<'PY'
from pathlib import Path
terms = ['sprint','Sprint','apply','Apply','dry-run','Dry-run','endpoint','API','payload','commit','deploy','migration','seed','debug','implementation','rollover','Rollover','matrix','Matrix','dirty']
paths = list(Path('apps/web-admin/src/routes/akademik').rglob('*.svelte')) + [Path('apps/web-admin/src/routes/students/+page.svelte')]
for p in paths:
    text = p.read_text(errors='ignore')
    for i,line in enumerate(text.splitlines(),1):
        if any(t in line for t in terms):
            print(f'{p}:{i}: {line.strip()}')
PY
```

**Step 2:** Tambahkan hasil audit ke file sementara atau catatan commit.

Expected: daftar lokasi copy yang perlu diganti.

---

## Task 2: Buat kamus copy Akademik reusable

**Objective:** Menghindari label yang berulang dan menjaga bahasa konsisten.

**Files:**
- Create: `apps/web-admin/src/lib/academic/copy.ts`

**Content direction:**

```ts
export const academicCopy = {
  actions: {
    save: 'Simpan',
    cancel: 'Batalkan',
    refresh: 'Muat ulang',
    previewPromotion: 'Pratinjau kenaikan kelas',
    applyPromotion: 'Terapkan kenaikan kelas',
    checkImport: 'Cek data sebelum impor',
    downloadTemplate: 'Unduh template'
  },
  labels: {
    academicYear: 'Tahun ajaran',
    homeroomTeacher: 'Wali kelas',
    subjectTeacher: 'Guru mapel',
    subjectAssignment: 'Penugasan guru mapel',
    weeklySchedule: 'Jadwal mingguan',
    unsavedChanges: 'Perubahan belum disimpan'
  },
  helper: {
    promotionPreview: 'Periksa daftar rombel, siswa, wali kelas, guru mapel, dan jadwal yang akan disiapkan untuk tahun ajaran tujuan.',
    promotionApply: 'Tindakan ini menyiapkan data tahun ajaran tujuan berdasarkan pratinjau. Data tahun ajaran lama tidak dihapus.',
    importCheck: 'Unggah data untuk diperiksa terlebih dahulu. Sistem belum menyimpan perubahan sebelum Bapak/Ibu menekan tombol simpan/impor.'
  }
} as const;
```

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

**Implementation note:** Selesai pada Tahap 2. Kamus copy dibuat di `apps/web-admin/src/lib/academic/copy.ts` dengan kelompok `actions`, `labels`, `helper`, `empty`, `status`, dan `terms` untuk dipakai bertahap pada halaman Akademik.

---

## Task 3: Bersihkan halaman Tahun Ajaran

**Objective:** Mengganti istilah teknis di halaman paling sensitif: tahun ajaran, pratinjau, dan kenaikan kelas.

**Files:**
- Modify: `apps/web-admin/src/routes/akademik/tahun-ajaran/+page.svelte`

**Changes:**
- `Preview rollover` → `Pratinjau kenaikan kelas`
- `Apply rollover` → `Terapkan kenaikan kelas`
- `Dry-run import` → `Cek data sebelum impor`
- `Safety token` → `Kalimat konfirmasi`
- `Result counts` → `Ringkasan hasil`
- `Warnings` → `Catatan yang perlu diperiksa`
- `students_skipped` → `Siswa perlu tindak lanjut manual`
- `classes_created` → `Rombel baru dibuat`
- `classes_reused` → `Rombel yang sudah ada digunakan`
- `homerooms_copied` → `Wali kelas disalin`
- `assignments_copied` → `Guru mapel disalin`
- `timetable_slots_copied` → `Jadwal disalin`

**UX rule:** Dialog berisiko harus menjelaskan dampak:

> Data tahun ajaran lama tidak dihapus. Sistem akan menyiapkan rombel tujuan, menyalin wali kelas/guru mapel/jadwal bila belum ada, dan memindahkan siswa VII ke VIII serta VIII ke IX sesuai pratinjau.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

**Implementation note:** Selesai pada Tahap 3. Halaman `apps/web-admin/src/routes/akademik/tahun-ajaran/+page.svelte` dibersihkan dari istilah user-facing seperti `Preview`, `Apply`, `Rollover`, `Dry-run`, `Challenge`, `Slot`, dan istilah teknis backend/API; diganti menjadi `Pratinjau kenaikan kelas`, `Terapkan kenaikan kelas`, `Cek data sebelum impor`, `Kalimat konfirmasi`, dan `Jadwal`.

---

## Task 4: Bersihkan Dashboard Akademik

**Objective:** Dashboard menjadi bahasa status sekolah, bukan status teknis.

**Files:**
- Modify: `apps/web-admin/src/routes/akademik/+page.svelte`

**Changes:**
- Gunakan judul seperti `Ringkasan Akademik`.
- Ganti istilah health/check teknis menjadi `Kelengkapan data`.
- Ganti wording “belum lengkap” dengan aksi lanjut yang jelas:
  - `Rombel tanpa wali kelas` → `Rombel belum memiliki wali kelas`
  - `Guru mapel belum lengkap` → `Penugasan guru mapel belum lengkap`
  - `Jadwal bentrok` → `Jadwal perlu diperiksa`

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

**Implementation note:** Selesai pada Tahap 4. Dashboard `apps/web-admin/src/routes/akademik/+page.svelte` diganti menjadi `Ringkasan Akademik`, memakai istilah `Kelengkapan data akademik`, `Rombel belum memiliki wali kelas`, `Penugasan guru mapel belum lengkap`, `Jadwal perlu diperiksa`, dan quick action bebas istilah `matrix`, `preview rollover`, serta `import-export`.

---

## Task 5: Bersihkan Mapel dan Guru Mapel

**Objective:** Menghindari istilah matrix dan metadata yang terasa teknis.

**Files:**
- Modify: `apps/web-admin/src/routes/akademik/mapel/+page.svelte`
- Modify: `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte`

**Changes:**
- `Matrix guru mapel` → `Tabel penugasan guru mapel`
- `Metadata mapel` → `Pengaturan mapel`
- `Assessment subject` → `Dipakai untuk asesmen`
- `Report subject` → `Masuk rapor`
- `Schedule activity` → `Aktivitas jadwal`
- `Default weekly hours` → `JP per minggu`
- `Display order` → `Urutan tampil`

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

**Implementation note:** Selesai pada Tahap 5. Halaman `apps/web-admin/src/routes/akademik/mapel/+page.svelte` dan `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte` dibersihkan dari istilah user-facing seperti `matrix/matriks`, `cell`, `backend`, `default`, `flag`, `error`, dan `Kegiatan` yang rancu. Copy diganti menjadi `Pengaturan mapel`, `Dipakai untuk asesmen`, `Masuk rapor`, `Aktivitas jadwal`, `JP per minggu`, `Tabel penugasan guru mapel`, `Penugasan lengkap`, dan `Perlu dilengkapi`.

---

## Task 6: Bersihkan Rombel, Detail Rombel, Jadwal, dan Siswa

**Objective:** Menghilangkan copy teknis dan menyamakan istilah aksi.

**Files:**
- Modify: `apps/web-admin/src/routes/akademik/rombel/+page.svelte`
- Modify: `apps/web-admin/src/routes/akademik/rombel/[id]/+page.svelte`
- Modify: `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`
- Modify: `apps/web-admin/src/routes/students/+page.svelte`

**Changes:**
- `dirty`/`unsaved` → `Perubahan belum disimpan`
- `slot` jika tampil ke user → `jam pelajaran` atau `jadwal`
- `assignment` → `penugasan`
- `bulk` → `aksi massal`
- `selected` → `dipilih`
- Hindari istilah `class` di UI; gunakan `rombel` atau `kelas` sesuai konteks.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

**Implementation note:** Selesai pada Tahap 6. Halaman `apps/web-admin/src/routes/akademik/rombel/+page.svelte`, `apps/web-admin/src/routes/akademik/rombel/[id]/+page.svelte`, `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`, dan `apps/web-admin/src/routes/students/+page.svelte` dibersihkan dari istilah user-facing seperti `slot`, `backend`, `assignment`, `legacy`, `lifecycle`, `permission`, `preview`, `generate`, `CBT`, `portal`, `role`, dan `drawer`. Copy diganti menjadi `jam pelajaran`, `layanan sistem`, `penugasan`, `data wali lama`, `status siswa`, `izin pengelolaan akun siswa`, `pratinjau`, `buat akun`, `asesmen`, `akun orang tua`, `peran`, dan `panel siswa`.

---

## Task 7: Backend domain error copy audit

**Objective:** Error yang muncul ke UI tidak terasa seperti error teknis.

**Files:**
- Modify if needed: `services/core-api/internal/handler/academic.go`
- Modify if needed: `services/core-api/internal/service/academic.go`
- Modify if needed: `services/core-api/internal/service/academic_year_rollover.go`

**Changes:**
- Pastikan pesan domain seperti:
  - `tahun ajaran tujuan wajib dipilih`
  - `kalimat konfirmasi tidak sesuai`
  - `rombel tujuan belum tersedia`
  - `jadwal bentrok perlu diperiksa`
- Hindari pesan teknis seperti `invalid payload`, `sql`, `constraint`, `token invalid` untuk user.

**Verification:**

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-akademik-copy-cleanup ./cmd/api
```

---

## Task 8: Tambah guard pencarian istilah developer

**Objective:** Mencegah istilah developer muncul lagi di halaman Akademik.

**Files:**
- Create: `scripts/check-academic-ui-copy.sh`

**Script idea:**

```bash
#!/usr/bin/env bash
set -euo pipefail
patterns='sprint|endpoint|payload|commit|deploy|migration|debug|implementation|dry-run|Dry-run|Apply rollover|Preview rollover'
if grep -RInE "$patterns" apps/web-admin/src/routes/akademik apps/web-admin/src/routes/students/+page.svelte; then
  echo "Developer-facing terms found in academic UI. Replace with user-facing copy."
  exit 1
fi
```

**Verification:**

```bash
bash scripts/check-academic-ui-copy.sh
```

---

## Final validation

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-akademik-copy-cleanup ./cmd/api
```

## Commit suggestion

```bash
git add apps/web-admin/src/lib/academic/copy.ts \
  apps/web-admin/src/routes/akademik \
  apps/web-admin/src/routes/students/+page.svelte \
  services/core-api/internal/handler/academic.go \
  services/core-api/internal/service/academic.go \
  services/core-api/internal/service/academic_year_rollover.go \
  scripts/check-academic-ui-copy.sh \
  .hermes/plans/2026-05-11_akademik-ui-copy-cleanup.md

git commit -m "refactor(academic): use operator-friendly UI copy"
```

## Deploy note

Jika validasi PASS, deploy normal mengikuti kebiasaan repo:

1. build backend `bin/api`
2. restart `mtsn2kolut-core-api`
3. health check
4. build web-admin
5. restart `mtsn2kolut-web-admin`
6. `pm2 save`
7. smoke check halaman akademik
