# KMA 1503 Tahun 2025 Kurikulum Merdeka Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menyesuaikan modul Akademik, Mapel, Guru Mapel, Jadwal, Nilai/Rapor, dan peringkat internal aplikasi MTsN 2 Kolaka Utara agar selaras dengan KMA 1503 Tahun 2025 Kurikulum Merdeka MTs.

**Architecture:** Tambahkan layer data kurikulum sebagai sumber kebenaran resmi per tahun ajaran dan tingkat, lalu jadikan Master Mapel, Penugasan Guru, Jadwal, dan Nilai/Rapor membaca dari struktur tersebut. SvelteKit tetap sebagai BFF/proxy ke Go core-api; PostgreSQL hanya diakses backend Go melalui sqlc/service/handler.

**Tech Stack:** Go core-api, PostgreSQL migrations + sqlc, SvelteKit web-admin, PM2 deploy, existing shadcn/Svelte components.

---

## 0. Referensi dan keputusan domain

### Referensi resmi

- PDF: `docs/references/KMA_1503_Tahun_2025_Pedoman_Implementasi_Kurikulum_kamimadrasah.pdf`
- Markdown OCR: `docs/references/KMA_1503_Tahun_2025_Pedoman_Implementasi_Kurikulum_kamimadrasah.md`
- Review awal: `docs/references/kma-1503-review-rekomendasi-aplikasi.md`

### Prinsip domain

1. **Kurikulum resmi harus dipisah dari Master Mapel.** Master Mapel adalah daftar entitas mapel/kegiatan, sedangkan alokasi JP berbeda per tingkat dan tahun ajaran.
2. **Kelas VII/VIII/IX memiliki alokasi yang berbeda.** Jangan hanya memakai `subjects.default_weekly_hours` global.
3. **Intrakurikuler dan kokurikuler harus dipisah.** P5RA/Kokurikuler tidak boleh diperlakukan sebagai mapel nilai biasa.
4. **BK adalah layanan**, bukan mapel peringkat/rata-rata.
5. **Muatan Lokal dan Mapel Pilihan perlu batas JP.** Operator boleh mengatur, tetapi ada guard sesuai rentang KMA.
6. **Peringkat/ranking tidak tampil default di rapor resmi.** Peringkat hanya fitur internal yang bisa diaktifkan dengan aturan jelas.
7. **Jadwal harus bergerak menuju nomor JP.** 1 JP MTs = 40 menit.

### Target struktur JP wajib MTs KMA 1503

- VII/VIII: 36 minggu, total wajib 1.512 JP/tahun = 42 JP/minggu.
- IX: 32 minggu, total wajib 1.344 JP/tahun = 42 JP/minggu.
- Baseline mingguan: 38 JP intrakurikuler + 4 JP kokurikuler.
- Tambahan madrasah maksimal 6 JP/minggu.

> Catatan verifikasi: baris Akidah Akhlak kelas IX pada scan perlu verifikasi manual kembali. Secara total, jika koku 32 JP, maka total harus 96 JP/tahun agar total wajib IX tetap 1.344 JP/tahun.

---

## 1. Scope implementasi

### In scope

- Model kurikulum KMA per tahun ajaran/tingkat.
- Seed struktur Kurikulum Merdeka MTs KMA 1503/2025.
- Master Mapel berbasis kategori dan flags operasional.
- Alokasi JP per mapel per tingkat.
- Penugasan guru mapel membawa JP dan validasi struktur KMA.
- Jadwal berbasis nomor JP/template jam.
- Validasi total JP per rombel, per mapel, dan beban guru.
- Rapor Kurikulum Merdeka fase awal dan pengaturan peringkat internal.
- Dashboard kesiapan kurikulum.

### Out of scope tahap awal

- Sinkronisasi eksternal EMIS/RDM.
- E-rapor resmi lengkap sampai format nasional final.
- AI otomatis menyusun jadwal penuh.
- Mobile siswa/orang tua untuk rapor detail.

---

## 2. Data model target

### 2.1 Tabel baru: `curriculum_profiles`

**Tujuan:** menyimpan profil kurikulum/regulasi yang berlaku.

Kolom minimal:

```sql
CREATE TABLE curriculum_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  regulation_reference TEXT NOT NULL DEFAULT '',
  education_level TEXT NOT NULL DEFAULT 'MTs',
  effective_academic_year_id UUID REFERENCES academic_years(id) ON DELETE SET NULL,
  status TEXT NOT NULL DEFAULT 'draft',
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_curriculum_profiles_status CHECK (status IN ('draft','active','archived'))
);
```

### 2.2 Tabel baru: `curriculum_subject_allocations`

**Tujuan:** menyimpan alokasi JP resmi per profil, tingkat, dan mapel.

Kolom minimal:

```sql
CREATE TABLE curriculum_subject_allocations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  curriculum_profile_id UUID NOT NULL REFERENCES curriculum_profiles(id) ON DELETE CASCADE,
  subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  level TEXT NOT NULL,
  subject_group TEXT NOT NULL DEFAULT 'wajib',
  intra_annual_hours INTEGER NOT NULL DEFAULT 0,
  koku_annual_hours INTEGER NOT NULL DEFAULT 0,
  total_annual_hours INTEGER NOT NULL DEFAULT 0,
  intra_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  koku_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  total_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  lesson_minutes INTEGER NOT NULL DEFAULT 40,
  display_order INTEGER NOT NULL DEFAULT 0,
  counts_for_schedule BOOLEAN NOT NULL DEFAULT TRUE,
  counts_for_report BOOLEAN NOT NULL DEFAULT TRUE,
  counts_for_assessment BOOLEAN NOT NULL DEFAULT TRUE,
  counts_for_ranking BOOLEAN NOT NULL DEFAULT TRUE,
  is_required BOOLEAN NOT NULL DEFAULT TRUE,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_curriculum_allocation_level CHECK (level IN ('VII','VIII','IX')),
  CONSTRAINT chk_curriculum_allocation_group CHECK (subject_group IN ('wajib','pilihan','muatan_lokal','layanan','kokurikuler','kegiatan')),
  CONSTRAINT uq_curriculum_allocation_profile_level_subject UNIQUE (curriculum_profile_id, level, subject_id)
);
```

### 2.3 Tabel baru: `class_curriculum_assignments`

**Tujuan:** menetapkan kurikulum aktif per rombel.

```sql
CREATE TABLE class_curriculum_assignments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id UUID NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
  curriculum_profile_id UUID NOT NULL REFERENCES curriculum_profiles(id) ON DELETE RESTRICT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_class_curriculum_active UNIQUE (class_id, curriculum_profile_id)
);
```

### 2.4 Extend `class_subject_assignments` atau tabel detail baru

Disarankan tabel detail baru agar tidak mengganggu assignment lama:

```sql
CREATE TABLE class_subject_allocation_overrides (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id UUID NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
  curriculum_allocation_id UUID REFERENCES curriculum_subject_allocations(id) ON DELETE SET NULL,
  intra_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  koku_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  additional_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  total_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  is_customized BOOLEAN NOT NULL DEFAULT FALSE,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_assignment_allocation_override UNIQUE (assignment_id)
);
```

### 2.5 Tabel baru: `lesson_period_templates`

**Tujuan:** jadwal berbasis nomor JP.

```sql
CREATE TABLE lesson_period_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 1 AND 6),
  period_number INTEGER NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  activity_type TEXT NOT NULL DEFAULT 'pelajaran',
  label TEXT NOT NULL DEFAULT '',
  is_counted_as_lesson BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT ck_lesson_period_time CHECK (end_time > start_time),
  CONSTRAINT chk_lesson_period_activity CHECK (activity_type IN ('pelajaran','istirahat','upacara','pembiasaan','kokurikuler','ekstrakurikuler','lainnya')),
  CONSTRAINT uq_lesson_period_template UNIQUE (academic_year_id, day_of_week, period_number)
);
```

### 2.6 Extend `timetable_slots`

Tambahkan opsional:

```sql
ALTER TABLE timetable_slots
  ADD COLUMN IF NOT EXISTS lesson_period_id UUID REFERENCES lesson_period_templates(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS slot_type TEXT NOT NULL DEFAULT 'pelajaran',
  ADD COLUMN IF NOT EXISTS lesson_hours NUMERIC(5,2) NOT NULL DEFAULT 1;
```

### 2.7 Rapor/peringkat

Tambahkan pengaturan:

```sql
CREATE TABLE report_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  curriculum_profile_id UUID REFERENCES curriculum_profiles(id) ON DELETE SET NULL,
  show_ranking_on_report BOOLEAN NOT NULL DEFAULT FALSE,
  ranking_method TEXT NOT NULL DEFAULT 'intrakurikuler_average',
  ranking_tie_policy TEXT NOT NULL DEFAULT 'same_rank',
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_report_settings_year UNIQUE (academic_year_id),
  CONSTRAINT chk_report_ranking_method CHECK (ranking_method IN ('intrakurikuler_average','weighted_by_jp','report_subject_average')),
  CONSTRAINT chk_report_tie_policy CHECK (ranking_tie_policy IN ('same_rank','dense_rank','ordinal'))
);
```

Opsional tahap lanjut:

```sql
CREATE TABLE student_report_descriptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id UUID NOT NULL REFERENCES class_subject_assignments(id) ON DELETE CASCADE,
  student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  semester TEXT NOT NULL DEFAULT 'Ganjil',
  achievement_description TEXT NOT NULL DEFAULT '',
  improvement_description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_student_report_description UNIQUE (assignment_id, student_id, semester)
);
```

---

## 3. Tahapan implementasi

## Tahap 0 — Baseline audit dan safety

**Objective:** Mengunci kondisi awal sebelum perubahan besar.

**Files:**
- Read: `docs/references/kma-1503-review-rekomendasi-aplikasi.md`
- Read: `services/core-api/db/queries/academic.sql`
- Read: `apps/web-admin/src/routes/akademik/mapel/+page.svelte`
- Read: `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte`
- Read: `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`
- Read: `apps/web-admin/src/routes/grades/+page.svelte`
- Read: `apps/web-admin/src/routes/grades/rapor/+page.svelte`

**Steps:**

1. Jalankan status repo:
   ```bash
   git status --short
   git branch --show-current
   git log --oneline -5
   ```

2. Snapshot struktur DB tanpa menampilkan credential:
   ```bash
   set -a; source .env >/dev/null 2>&1; set +a
   psql "$DATABASE_URL" -P pager=off -c "
   select table_name from information_schema.tables
   where table_schema='public'
     and table_name in ('subjects','school_classes','class_subject_assignments','timetable_slots','grade_components','grade_entries')
   order by table_name;"
   ```

3. Jalankan baseline checks:
   ```bash
   npm --prefix apps/web-admin run check
   cd services/core-api
   /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   go test ./internal/handler ./internal/service ./internal/repository/postgres
   go build -o /tmp/core-api-kma-baseline ./cmd/api
   ```

4. Commit hanya jika ada dokumentasi baseline baru:
   ```bash
   git add .hermes/plans/2026-05-12_092239-kma1503-kurikulum-merdeka-comprehensive.md
   git commit -m "docs(academic): plan KMA 1503 curriculum alignment"
   ```

**Expected:** baseline PASS sebelum coding.

---

## Tahap 1 — Migration struktur kurikulum

**Objective:** Menambahkan tabel kurikulum dan alokasi JP tanpa mengubah perilaku existing.

**Files:**
- Create: `services/core-api/db/migrations/089_curriculum_kma_foundation.sql`
- Modify: `services/core-api/db/queries/academic.sql`

**Steps:**

1. Buat migration tabel:
   - `curriculum_profiles`
   - `curriculum_subject_allocations`
   - `class_curriculum_assignments`
   - `report_settings`

2. Tambahkan query sqlc minimal:
   - `ListCurriculumProfiles`
   - `GetActiveCurriculumProfile`
   - `ListCurriculumSubjectAllocations`
   - `GetCurriculumSummaryByLevel`
   - `ListClassCurriculumAssignments`

3. Generate sqlc:
   ```bash
   cd services/core-api
   /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   ```

4. Test migration di transaction production-like DB sebelum commit:
   ```bash
   set -a; source ../../.env >/dev/null 2>&1; set +a
   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "BEGIN; \i db/migrations/089_curriculum_kma_foundation.sql; ROLLBACK;"
   ```

5. Validasi backend:
   ```bash
   go test ./internal/repository/postgres ./internal/service ./internal/handler
   go build -o /tmp/core-api-kma-stage1 ./cmd/api
   ```

6. Commit:
   ```bash
   git add services/core-api/db/migrations/089_curriculum_kma_foundation.sql services/core-api/db/queries/academic.sql services/core-api/internal/repository/postgres
   git commit -m "feat(academic): add curriculum structure foundation"
   ```

**Expected:** tabel baru ada, belum ada UI berubah.

---

## Tahap 2 — Seed KMA 1503 MTs VII/VIII/IX

**Objective:** Mengisi struktur KMA 1503 sebagai profil kurikulum aktif/konsep yang bisa diverifikasi operator.

**Files:**
- Create: `services/core-api/db/migrations/090_seed_kma1503_mts_curriculum.sql`
- Create: `docs/references/kma1503-mts-allocation-table.md`

**Steps:**

1. Buat mapping subject code resmi:
   - `QH`, `AA`, `FIKIH`, `SKI`, `BARAB`, `PP`, `BIND`, `MTK`, `IPA`, `IPS`, `BING`, `PJOK`, `INF`, `SBDP` atau mapping ke `SBD` existing.

2. Tambahkan subject jika belum ada:
   - `SBDP` atau normalisasi `SBD` menjadi `Seni, Budaya, dan Prakarya`.
   - `KKA` untuk Koding dan Kecerdasan Artifisial bila madrasah ingin memakai mapel pilihan.

3. Seed `curriculum_profiles`:
   ```text
   code: KMA1503-2025-MTS
   name: Kurikulum Merdeka MTs KMA 1503 Tahun 2025
   regulation_reference: KMA 1503 Tahun 2025
   status: draft atau active sesuai keputusan operator
   ```

4. Seed alokasi VII/VIII/IX sesuai tabel review.

5. Tambahkan query validasi total:
   ```sql
   SELECT level, SUM(total_weekly_hours), SUM(intra_weekly_hours), SUM(koku_weekly_hours)
   FROM curriculum_subject_allocations
   WHERE curriculum_profile_id = (...)
   GROUP BY level;
   ```

6. Test migration di transaction:
   ```bash
   cd services/core-api
   set -a; source ../../.env >/dev/null 2>&1; set +a
   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "BEGIN; \i db/migrations/090_seed_kma1503_mts_curriculum.sql; ROLLBACK;"
   ```

7. Validasi:
   - VII total 42 JP/minggu.
   - VIII total 42 JP/minggu.
   - IX total 42 JP/minggu.

8. Commit:
   ```bash
   git add services/core-api/db/migrations/090_seed_kma1503_mts_curriculum.sql docs/references/kma1503-mts-allocation-table.md
   git commit -m "feat(academic): seed KMA 1503 MTs curriculum allocation"
   ```

**Expected:** struktur KMA tersimpan, total JP valid.

---

## Tahap 3 — Backend API Kurikulum

**Objective:** Menyediakan layanan sistem untuk membaca profil kurikulum dan alokasi JP.

**Files:**
- Modify: `services/core-api/internal/service/academic.go`
- Modify: `services/core-api/internal/handler/academic.go`
- Modify: `services/core-api/cmd/api/main.go`
- Modify: `services/core-api/internal/handler/academic_success_test.go`
- Modify: `services/core-api/internal/service/simple_wrappers_test.go`

**API target:**

- `GET /api/academic/curriculum/profiles`
- `GET /api/academic/curriculum/active`
- `GET /api/academic/curriculum/allocations?profile_id=&level=`
- `GET /api/academic/curriculum/summary?profile_id=`

**Steps:**

1. Tambahkan service DTO:
   - `CurriculumProfile`
   - `CurriculumSubjectAllocation`
   - `CurriculumLevelSummary`

2. Tambahkan handler JSON response dengan bahasa operator.

3. Tambahkan route di `cmd/api/main.go`.

4. Tambahkan handler tests:
   - success list profiles.
   - success list allocations.
   - invalid UUID menghasilkan pesan operator.

5. Validasi:
   ```bash
   cd services/core-api
   gofmt -w internal/service/academic.go internal/handler/academic.go internal/handler/academic_success_test.go internal/service/simple_wrappers_test.go cmd/api/main.go
   /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   go test ./internal/handler ./internal/service ./internal/repository/postgres
   go build -o /tmp/core-api-kma-stage3 ./cmd/api
   ```

6. Commit:
   ```bash
   git add services/core-api
   git commit -m "feat(academic): expose curriculum allocation API"
   ```

**Expected:** backend API kurikulum berjalan dan teruji.

---

## Tahap 4 — UI Akademik > Kurikulum

**Objective:** Operator bisa melihat struktur KMA, JP per tingkat, dan status total JP.

**Files:**
- Create: `apps/web-admin/src/routes/api/academic/curriculum/+server.ts`
- Create: `apps/web-admin/src/routes/akademik/kurikulum/+page.svelte`
- Modify: sidebar/navigation file existing untuk menu Akademik.

**UI target:**

- Judul: `Struktur Kurikulum`
- Filter profil kurikulum.
- Tab tingkat: VII, VIII, IX.
- Tabel mapel:
  - urutan;
  - mapel;
  - kelompok;
  - JP intrakurikuler/tahun;
  - JP kokurikuler/tahun;
  - total JP/tahun;
  - total JP/minggu;
  - masuk jadwal/rapor/peringkat.
- Ringkasan:
  - total wajib;
  - total intra;
  - total koku;
  - status sesuai 42 JP/minggu.

**Steps:**

1. Buat BFF proxy.
2. Buat page dengan AsyncContent/RecoveryPanel pattern existing.
3. Tambahkan badge status `Sesuai KMA` / `Perlu ditinjau`.
4. Tambahkan link ke file referensi Markdown/PDF jika UI mendukung.
5. Validasi frontend:
   ```bash
   npm --prefix apps/web-admin run check
   npm --prefix apps/web-admin run build
   ```

6. Commit:
   ```bash
   git add apps/web-admin/src/routes/api/academic/curriculum apps/web-admin/src/routes/akademik/kurikulum apps/web-admin/src
   git commit -m "feat(academic): add curriculum structure page"
   ```

**Expected:** operator dapat melihat struktur KMA tanpa mengedit data dulu.

---

## Tahap 5 — Master Mapel KMA-ready

**Objective:** Menyesuaikan Master Mapel agar kategori/flags mengikuti kebutuhan KMA.

**Files:**
- Modify: `services/core-api/db/migrations/091_subject_kma_metadata.sql`
- Modify: `apps/web-admin/src/routes/akademik/mapel/+page.svelte`
- Modify: backend subject service/handler jika perlu.

**Perubahan data:**

- Tambah kolom bila belum ada:
  - `subject_group` atau gunakan `category` dengan nilai lebih kaya.
  - `counts_for_ranking BOOLEAN DEFAULT TRUE`.
  - `is_local_content BOOLEAN DEFAULT FALSE`.
  - `is_choice_subject BOOLEAN DEFAULT FALSE`.

**Rekomendasi status:**

- BK:
  - kategori: `layanan`
  - masuk jadwal: boleh
  - masuk asesmen: false
  - masuk rapor: false atau catatan saja
  - masuk peringkat: false

- P5RA/Kokurikuler:
  - kategori: `kokurikuler`
  - masuk jadwal: true
  - masuk asesmen: false atau khusus projek
  - masuk rapor: khusus deskripsi
  - masuk peringkat: false

- Muatan Lokal:
  - kategori: `muatan_lokal`
  - masuk rapor: configurable
  - masuk peringkat: configurable sesuai kebijakan madrasah

**Steps:**

1. Migration metadata tambahan.
2. Update query `ListSubjects`, `CreateSubject`, `UpdateSubject`.
3. Update DTO/service/handler.
4. UI Mapel: tampilkan kategori KMA, masuk peringkat, muatan lokal/pilihan.
5. Validasi:
   ```bash
   npm --prefix apps/web-admin run check
   cd services/core-api
   /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   go test ./internal/handler ./internal/service ./internal/repository/postgres
   go build -o /tmp/core-api-kma-stage5 ./cmd/api
   ```

6. Commit:
   ```bash
   git add services/core-api apps/web-admin/src/routes/akademik/mapel
   git commit -m "feat(academic): align subjects with KMA categories"
   ```

**Expected:** mapel bisa dibedakan antara wajib, pilihan, mulok, layanan, dan kokurikuler.

---

## Tahap 6 — Penugasan Guru Mapel + JP

**Objective:** Guru Mapel menyimpan dan menampilkan alokasi JP per rombel-mapel.

**Files:**
- Create: migration `092_class_subject_allocation_overrides.sql`
- Modify: `services/core-api/db/queries/academic.sql`
- Modify: `services/core-api/internal/service/academic.go`
- Modify: `services/core-api/internal/handler/academic.go`
- Modify: `apps/web-admin/src/routes/akademik/guru-mapel/+page.svelte`

**Backend target:**

- Matrix assignment response menyertakan:
  - `curriculum_allocation_id`
  - `intra_weekly_hours`
  - `koku_weekly_hours`
  - `additional_weekly_hours`
  - `total_weekly_hours`
  - `is_customized`
  - `compliance_status`

**UI target:**

- Cell Guru Mapel menampilkan:
  - guru;
  - total JP;
  - status sesuai/kurang/lebih;
  - tombol edit detail JP.

**Validation:**

- default JP mengikuti `curriculum_subject_allocations` sesuai tingkat rombel.
- tambahan JP maksimal 6 per minggu total rombel.
- pilihan maksimal 2 JP/minggu.
- muatan lokal dalam rentang KMA.

**Steps:**

1. Tambah migration override.
2. Update query matrix read/write.
3. Update service validation.
4. Update UI matrix.
5. Tambah tests service untuk:
   - default allocation applied.
   - over-allocation rejected.
   - missing curriculum returns warning, not crash.
6. Validasi full.
7. Commit:
   ```bash
   git commit -m "feat(academic): add weekly hours to teacher assignments"
   ```

**Expected:** Guru Mapel menjadi sumber awal beban JP per rombel dan per guru.

---

## Tahap 7 — Dashboard kepatuhan kurikulum

**Objective:** Operator melihat ringkasan kesesuaian KMA dari dashboard akademik.

**Files:**
- Modify: `services/core-api/db/queries/academic.sql`
- Modify: `services/core-api/internal/service/academic.go`
- Modify: `services/core-api/internal/handler/academic.go`
- Modify: `apps/web-admin/src/routes/akademik/+page.svelte`

**Metrics:**

- Rombel tanpa profil kurikulum.
- Rombel total JP kurang dari 42.
- Rombel total JP lebih dari 48.
- Mapel wajib belum punya guru.
- Guru dengan beban kurang dari 24 JTM/minggu.
- Jadwal yang belum memenuhi JP assignment.
- Mapel non-peringkat yang masih masuk rapor/peringkat.

**Steps:**

1. Tambah query summary.
2. Tambah handler/service fields.
3. Update dashboard `Ringkasan Akademik` dengan card `Kesesuaian Kurikulum`.
4. Validasi query ke DB.
5. Commit:
   ```bash
   git commit -m "feat(academic): add curriculum compliance dashboard"
   ```

**Expected:** kepala/operator bisa melihat masalah kurikulum sebelum menyusun rapor.

---

## Tahap 8 — Template jam pelajaran

**Objective:** Menyediakan master nomor JP agar jadwal tidak hanya jam bebas.

**Files:**
- Create: migration `093_lesson_period_templates.sql`
- Modify: `services/core-api/db/queries/timetable.sql` atau `academic.sql`
- Modify: service/handler timetable.
- Create: `apps/web-admin/src/routes/akademik/jam-pelajaran/+page.svelte`

**UI target:**

- Template per hari.
- Nomor JP.
- Jam mulai/selesai.
- Tipe kegiatan.
- Durasi otomatis.
- Duplikasi template Senin ke hari lain.

**Steps:**

1. Tambah table.
2. Seed template default MTsN 2 Kolut jika disetujui operator.
3. Buat CRUD API.
4. Buat UI editable.
5. Validasi frontend/backend.
6. Commit:
   ```bash
   git commit -m "feat(academic): add lesson period templates"
   ```

**Expected:** jadwal punya dasar nomor JP yang konsisten.

---

## Tahap 9 — Jadwal berbasis nomor JP dan validasi JP

**Objective:** Jadwal mingguan menghitung JP otomatis dan membandingkannya dengan alokasi guru mapel.

**Files:**
- Modify: migration/queries timetable.
- Modify: `apps/web-admin/src/routes/akademik/jadwal/+page.svelte`
- Modify: backend timetable service/handler.

**Features:**

- Pilih nomor JP dari template.
- Tipe kegiatan: pelajaran, istirahat, upacara, pembiasaan, kokurikuler.
- Hitung JP mapel per rombel.
- Indikator:
  - sesuai alokasi;
  - kurang JP;
  - lebih JP;
  - bentrok guru;
  - bentrok ruang;
  - bentrok rombel.

**Steps:**

1. Extend `timetable_slots`.
2. Update query list/create/update slot.
3. Update UI jadwal weekly grid.
4. Tambah validation service.
5. Tests:
   - overlapping teacher rejected/warned.
   - total JP calculation correct.
6. Commit:
   ```bash
   git commit -m "feat(academic): validate timetable against weekly hours"
   ```

**Expected:** jadwal membantu operator memenuhi struktur KMA, bukan hanya mencatat jam.

---

## Tahap 10 — Beban tatap muka guru

**Objective:** Menghitung JTM guru dari penugasan dan jadwal.

**Files:**
- Modify: `services/core-api/db/queries/academic.sql`
- Modify: academic service/handler.
- Create: `apps/web-admin/src/routes/akademik/beban-guru/+page.svelte`

**Metrics:**

- Total JTM dari intrakurikuler.
- JTM kokurikuler/fasilitator.
- Koordinator kokurikuler setara 2 JTM per rombel/tahun sesuai catatan KMA.
- Status < 24, 24–40, > 40.

**Steps:**

1. Query agregasi per guru.
2. Handler `GET /api/academic/teacher-workload`.
3. UI list guru.
4. Filter status kurang/cukup/lebih.
5. Export CSV/unduh data bila diperlukan.
6. Commit:
   ```bash
   git commit -m "feat(academic): add teacher workload overview"
   ```

**Expected:** operator/kepala bisa memantau beban tatap muka guru.

---

## Tahap 11 — Rapor Kurikulum Merdeka fase 1

**Objective:** Rapor mulai mendukung deskripsi capaian dan pemisahan mapel yang masuk/tidak masuk peringkat.

**Files:**
- Create migration `094_report_settings_and_descriptions.sql`
- Modify: `apps/web-admin/src/routes/grades/+page.svelte`
- Modify: `apps/web-admin/src/routes/grades/rapor/+page.svelte`
- Modify: backend grades service/handler/queries.

**Features:**

- Pengaturan rapor per tahun ajaran:
  - tampilkan peringkat di rapor: default false;
  - metode peringkat;
  - kebijakan nilai sama.
- Deskripsi capaian per siswa-mapel.
- Rapor print hanya menampilkan mapel `counts_for_report=true`.
- Peringkat hanya jika `show_ranking_on_report=true`.

**Steps:**

1. Migration report settings.
2. Query settings.
3. Service peringkat internal.
4. UI pengaturan rapor.
5. Update halaman cetak rapor.
6. Tests ranking:
   - exclude BK/P5RA.
   - same average uses same rank when configured.
7. Commit:
   ```bash
   git commit -m "feat(grades): add curriculum report settings"
   ```

**Expected:** rapor lebih aman untuk Kurikulum Merdeka dan ranking tidak otomatis muncul.

---

## Tahap 12 — Validasi, deploy, dan dokumentasi operator

**Objective:** Deploy perubahan setelah semua validasi PASS dan dokumentasi operator tersedia.

**Files:**
- Create/Modify: `docs/operator/kurikulum-merdeka-kma1503.md`
- Modify: `.hermes/plans/2026-05-12_092239-kma1503-kurikulum-merdeka-comprehensive.md` implementation notes.

**Validation full:**

```bash
bash scripts/check-operator-ui-copy.sh
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-kma-final ./cmd/api
```

**Pre-deploy DB backup:**

```bash
set -a; source .env >/dev/null 2>&1; set +a
TS=$(TZ=Asia/Makassar date +%Y%m%d-%H%M%S)
mkdir -p /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql
pg_dump "$DATABASE_URL" -Fc -f "/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-kma1503-curriculum-$TS.dump"
sha256sum "/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-kma1503-curriculum-$TS.dump"
```

**Migration production:**

Gunakan mekanisme migration existing repo. Jika belum ada runner otomatis, jalankan migration SQL satu per satu dengan catatan eksplisit dan backup sudah tersedia.

**Deploy:**

```bash
cd services/core-api
TS=$(TZ=Asia/Makassar date +%Y%m%d-%H%M%S)
cp -p bin/api "bin/api.backup-kma1503-$TS"
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env

cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run build
pm2 restart mtsn2kolut-web-admin --update-env
pm2 save
```

**Smoke:**

```bash
curl -fsS http://127.0.0.1:8080/health
curl -I http://127.0.0.1:8021/
curl -I http://127.0.0.1:8021/login
```

**Commit docs:**

```bash
git add docs/operator/kurikulum-merdeka-kma1503.md .hermes/plans/2026-05-12_092239-kma1503-kurikulum-merdeka-comprehensive.md
git commit -m "docs(academic): document KMA 1503 curriculum workflow"
```

**Expected:** production online, health OK, route akademik login redirect/200 normal, PM2 saved.

---

## 4. Acceptance criteria final

### Data

- Ada satu profil kurikulum KMA 1503 MTs.
- Alokasi VII/VIII/IX tersimpan.
- Total wajib tiap tingkat = 42 JP/minggu.
- Mapel memiliki kategori KMA benar.
- BK dan P5RA tidak masuk peringkat default.

### UI

- Operator dapat melihat struktur kurikulum per tingkat.
- Operator dapat melihat JP mapel per tingkat.
- Guru Mapel menampilkan guru + JP.
- Jadwal menampilkan nomor JP.
- Dashboard memberi peringatan jika struktur belum lengkap.
- Rapor tidak menampilkan peringkat secara default.

### Backend

- Query sqlc generate bersih.
- Handler/service tests PASS.
- Validasi over-allocation berjalan.
- Pesan error memakai bahasa operator.

### Production

- Backup DB tersedia sebelum migration.
- PM2 core-api dan web-admin online.
- `/health` OK dan DB connected.
- Smoke web-admin OK.

---

## 5. Risiko dan mitigasi

### Risiko: angka KMA OCR salah

Mitigasi:

- Verifikasi tabel dari PDF asli sebelum seed final.
- Simpan catatan sumber halaman.
- Untuk baris ambigu kelas IX, jangan seed sebelum diverifikasi.

### Risiko: perubahan struktur mengganggu data existing

Mitigasi:

- Tambah tabel baru, jangan rewrite tabel lama.
- Migration additive.
- Rollback plan dengan backup DB.

### Risiko: operator bingung dengan istilah kurikulum

Mitigasi:

- Pakai istilah madrasah/operator.
- Hindari istilah developer.
- Jalankan `scripts/check-operator-ui-copy.sh`.

### Risiko: ranking dianggap rapor resmi

Mitigasi:

- Default `show_ranking_on_report=false`.
- Label UI: `Peringkat internal`.
- Harus aktif manual oleh operator.

---

## 6. Commit sequence rekomendasi

1. `docs(academic): plan KMA 1503 curriculum alignment`
2. `feat(academic): add curriculum structure foundation`
3. `feat(academic): seed KMA 1503 MTs curriculum allocation`
4. `feat(academic): expose curriculum allocation API`
5. `feat(academic): add curriculum structure page`
6. `feat(academic): align subjects with KMA categories`
7. `feat(academic): add weekly hours to teacher assignments`
8. `feat(academic): add curriculum compliance dashboard`
9. `feat(academic): add lesson period templates`
10. `feat(academic): validate timetable against weekly hours`
11. `feat(academic): add teacher workload overview`
12. `feat(grades): add curriculum report settings`
13. `docs(academic): document KMA 1503 curriculum workflow`

---

## 7. Recommended execution approach

Implement bertahap dan jangan deploy sampai satu milestone besar PASS.

Milestone aman:

- **Milestone 1:** Tahap 1–4, hanya struktur kurikulum + UI baca.
- **Milestone 2:** Tahap 5–7, mapel + guru mapel + dashboard compliance.
- **Milestone 3:** Tahap 8–10, template JP + jadwal + beban guru.
- **Milestone 4:** Tahap 11–12, rapor/peringkat + deploy final.

Setiap milestone harus:

```bash
bash scripts/check-operator-ui-copy.sh
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-kma-milestoneN ./cmd/api
```

