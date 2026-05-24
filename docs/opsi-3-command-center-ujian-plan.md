# Architecture Plan Opsi 3 — Command Center Ujian

MTsN 2 Kolaka Utara Super App  
Fokus: backend, data model, API, impor jadwal pengawas, pembaruan master data pegawai/guru dari daftar foto, validasi, rollback, dan audit-safe rollout.  
Tanggal rencana: 24 Mei 2026

---

## 1. Ringkasan Eksekutif

**Opsi 3 Command Center Ujian** adalah modul operasional ujian yang memusatkan:

- master kegiatan ujian,
- ruang ujian,
- sesi/hari ujian,
- penugasan pengawas per ruang,
- impor jadwal dari dokumen/foto/CSV,
- koreksi nama pegawai/guru sebagai master data,
- audit trail setiap perubahan,
- validasi sebelum publish,
- rollback aman bila impor atau update data bermasalah.

Prinsip utama:

- **PostgreSQL hanya dimiliki `services/core-api`.**
- **Web Admin SvelteKit hanya BFF/proxy**, bukan pemilik data.
- **Semua SQL eksplisit di `services/core-api/db/queries` dan digenerate sqlc.**
- **Tidak langsung menulis jadwal final dari hasil OCR/foto.** Semua masuk ke staging import, divalidasi, direview, baru di-commit.
- **Nama pengawas dari foto diperlakukan sebagai master data rujukan**, tetapi update ke `employees` tetap dibuat sebagai batch yang bisa diaudit dan di-rollback.

---

## 2. Kondisi Repo dan Baseline yang Relevan

Berdasarkan struktur repo saat ini:

- Backend: `services/core-api` Go + Chi + sqlc + pgx + PostgreSQL.
- Frontend admin: `apps/web-admin` SvelteKit BFF/proxy.
- Tabel existing yang relevan:
  - `employees`: master pegawai umum.
  - `users`: akun login, bisa terkait `employee_id`.
  - `audit_logs`: audit umum existing.
  - `academic_years`: tahun pelajaran.
  - `cbt_exam_events`: kegiatan ujian.
  - `cbt_exam_sessions`: sesi ujian.
  - `cbt_exam_rooms`: ruang ujian per sesi.
  - `school_rooms`: master ruang fisik.
  - `cbt_room_proctors`: pengawas/proktor per ruang ujian.
  - `cbt_participant_events`, `cbt_proctor_actions`: audit runtime CBT/proctoring.

Implikasi desain:

- Jangan membuat sistem ujian paralel yang lepas dari CBT/Assessment existing.
- Tambahkan lapisan **Command Center** sebagai orchestration, import, readiness, dan audit di atas tabel CBT existing.
- Gunakan tabel staging/import untuk data yang datang dari foto agar aman.

---

## 3. Data dari Foto yang Harus Dijadikan Master Pengawas

Daftar pengawas dari foto yang harus dipakai untuk pembaruan nama di aplikasi:

1. Taufik S.Ag M.Pd
2. Kardi S.Pd
3. Irmawati Nur S.Ag M.Pd
4. Supriadi S.Pd
5. Dra Ratnawati
6. Drs H Chaeruddin
7. Andi Rasnia S.Ag
8. Sitti Rafidah S.Pd.I
9. Herniati S.Pd
10. Muh Saing S.E
11. Sutra S.Ag
12. Saimang S.Ag
13. Mida S.Ag
14. Rusnawati S.Pd
15. Wati Zaelani Sidi Ellyanando S.Pd
16. Susianti S.Pd
17. Rustiani S.Pd
18. Nirmawati Tahir S.Pd
19. Nurul Fitri Usman S.Pd
20. Kaharuddin S.Pd
21. Abdillah S.Pd
22. Yurnianti S.Pd
23. Mutmainnah S.Pd
24. Andi Mirna S.Pd
25. Nur Ilmi S.Pd
26. Mirnawati S.Pd
27. Nurunnisa Alimah A S.Pd

Catatan kualitas data dari foto:

- Kegiatan: **Evaluasi Semester Genap TP 2025/2026**.
- Ruang: **R.I sampai R.VIII**.
- Rentang tanggal final hasil koreksi: **4–6 Juni 2026 dan 8–10 Juni 2026**. Bagian yang tertulis Senin dan seterusnya mengikuti **nama hari**: Senin 08 Juni, Selasa 09 Juni, Rabu 10 Juni 2026.
- Koreksi hari/tanggal sudah diputuskan: bagian Senin dan seterusnya digeser menjadi **Senin 08 Juni 2026, Selasa 09 Juni 2026, Rabu 10 Juni 2026**. Tidak ada ujian pada Minggu 07 Juni 2026.
- Ada **duplikasi nomor pengawas 23** pada jadwal/foto. Sesuai arahan panitia, duplikasi ini **boleh tetap masuk sebagai catatan pending koreksi**, tidak memblokir import awal, dan diperbaiki manual besok sebelum final/pelaksanaan.
- Master nomor pengawas valid hanya **1–27**. Nomor di luar rentang atau kosong harus error.

---

## 4. Target Arsitektur Opsi 3

### 4.1 Lapisan Backend

Modul baru di Core API:

- `internal/service/exam_command_center.go`
- `internal/handler/exam_command_center.go`
- `db/queries/exam_command_center.sql`
- migration baru, misalnya `120_exam_command_center.sql`

Tanggung jawab service:

- Membuat/mengelola event ujian.
- Import staging jadwal pengawas.
- Validasi data staging.
- Apply batch update master pegawai.
- Apply jadwal final ke `cbt_exam_sessions`, `cbt_exam_rooms`, dan `cbt_room_proctors`.
- Membuat audit log dan rollback snapshot.
- Menyediakan overview/readiness Command Center.

### 4.2 Lapisan Web Admin

Web Admin hanya proxy/BFF:

- route UI: `/asesmen/pelaksanaan/command-center` atau `/ujian/command-center`.
- BFF API: `/api/command-center/*` proxy ke Core API.
- Tidak ada akses DB langsung.

### 4.3 Alur Data

1. Admin upload/paste jadwal dari foto/OCR/CSV.
2. Backend membuat `exam_import_batches` status `draft`.
3. Data pengawas dari foto masuk `exam_import_supervisor_master_rows`.
4. Jadwal per ruang masuk `exam_import_schedule_rows`.
5. Backend menjalankan validasi:
   - nomor pengawas valid 1–27,
   - nama sesuai master,
   - ruang R.I–R.VIII valid,
   - tanggal final 4–6 Juni dan 8–10 Juni 2026 valid,
   - hari/tanggal konsisten,
   - tidak ada konflik satu pengawas di dua ruang pada slot sama,
   - tidak ada duplicate role utama di satu ruang/slot,
   - nomor 23 duplikat ditandai sebagai `pending_koreksi_pengawas`, bukan blocker awal.
6. Admin review diff.
7. Admin klik **Apply Master Data** untuk update nama/alias pegawai.
8. Admin klik **Publish Jadwal** untuk membuat sesi/ruang/pengawas final.
9. Semua perubahan mencatat audit + snapshot rollback.

---

## 5. Data Model yang Direkomendasikan

### 5.1 Pertahankan Tabel Existing Sebagai Source of Truth Final

Final runtime tetap memakai:

- `cbt_exam_events`
- `cbt_exam_sessions`
- `cbt_exam_rooms`
- `school_rooms`
- `cbt_room_proctors`
- `employees`
- `audit_logs`

### 5.2 Tambahan Kolom Ringan pada `employees`

Tujuan: master nomor pengawas dari foto bisa stabil tanpa menyalahgunakan `nip` atau `pegawai_uid`.

```sql
ALTER TABLE employees
  ADD COLUMN IF NOT EXISTS exam_supervisor_code INTEGER,
  ADD COLUMN IF NOT EXISTS display_name_source TEXT NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS name_verified_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS name_verified_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_exam_supervisor_code
  ON employees(exam_supervisor_code)
  WHERE exam_supervisor_code IS NOT NULL;

ALTER TABLE employees
  ADD CONSTRAINT chk_employees_exam_supervisor_code
  CHECK (exam_supervisor_code IS NULL OR exam_supervisor_code BETWEEN 1 AND 999);
```

Catatan:

- Untuk kasus foto ini gunakan kode 1–27.
- Jangan pakai kode ini untuk identitas negara/resmi; ini kode operasional jadwal ujian.
- `display_name_source = 'photo_import_2026_genap'` saat nama berasal dari daftar foto dan sudah disetujui.

### 5.3 Tabel Batch Import

```sql
CREATE TABLE exam_import_batches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id UUID REFERENCES cbt_exam_events(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  source_type TEXT NOT NULL DEFAULT 'photo_ocr',
  source_label TEXT NOT NULL DEFAULT '',
  source_checksum TEXT NOT NULL DEFAULT '',
  academic_year_id UUID REFERENCES academic_years(id) ON DELETE SET NULL,
  date_start DATE NOT NULL,
  date_end DATE NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft',
  validation_summary JSONB NOT NULL DEFAULT '{}'::jsonb,
  applied_master_at TIMESTAMPTZ,
  applied_schedule_at TIMESTAMPTZ,
  applied_by UUID REFERENCES users(id) ON DELETE SET NULL,
  rollback_of_batch_id UUID REFERENCES exam_import_batches(id) ON DELETE SET NULL,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_exam_import_batches_status CHECK (
    status IN ('draft', 'validated', 'needs_review', 'applied_master', 'published', 'rolled_back', 'rejected')
  )
);
```

### 5.4 Tabel Master Row Pengawas dari Foto

```sql
CREATE TABLE exam_import_supervisor_master_rows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id UUID NOT NULL REFERENCES exam_import_batches(id) ON DELETE CASCADE,
  supervisor_code INTEGER NOT NULL,
  raw_name TEXT NOT NULL,
  normalized_name TEXT NOT NULL,
  matched_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
  match_strategy TEXT NOT NULL DEFAULT 'none',
  match_confidence NUMERIC(5,2) NOT NULL DEFAULT 0,
  action TEXT NOT NULL DEFAULT 'review',
  validation_errors JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_exam_import_supervisor_code_per_batch UNIQUE(batch_id, supervisor_code),
  CONSTRAINT chk_exam_import_supervisor_code CHECK (supervisor_code BETWEEN 1 AND 999),
  CONSTRAINT chk_exam_import_supervisor_action CHECK (
    action IN ('create_employee', 'update_employee_name', 'link_existing', 'skip', 'review')
  )
);
```

Untuk data foto ini importer membuat 27 row. Bila ada duplikasi nomor pada daftar master, constraint akan gagal di staging atau ditandai sebelum insert, bukan mengubah final.

### 5.5 Tabel Row Jadwal Pengawas dari Foto

Karena jadwal perlu import per ruang sesuai nomor kode pengawas:

```sql
CREATE TABLE exam_import_schedule_rows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id UUID NOT NULL REFERENCES exam_import_batches(id) ON DELETE CASCADE,
  row_no INTEGER NOT NULL,
  exam_date DATE NOT NULL,
  day_name_raw TEXT NOT NULL DEFAULT '',
  session_label TEXT NOT NULL DEFAULT '',
  starts_at TIME,
  ends_at TIME,
  room_code TEXT NOT NULL,
  room_name_raw TEXT NOT NULL DEFAULT '',
  supervisor_code INTEGER NOT NULL,
  supervisor_role TEXT NOT NULL DEFAULT 'utama',
  raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  matched_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
  matched_school_room_id UUID REFERENCES school_rooms(id) ON DELETE SET NULL,
  target_session_id UUID,
  target_exam_room_id UUID,
  validation_errors JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_exam_import_schedule_supervisor_code CHECK (supervisor_code BETWEEN 1 AND 999),
  CONSTRAINT chk_exam_import_schedule_role CHECK (supervisor_role IN ('utama', 'pendamping', 'cadangan'))
);

CREATE INDEX idx_exam_import_schedule_batch_date_room
  ON exam_import_schedule_rows(batch_id, exam_date, room_code);

CREATE INDEX idx_exam_import_schedule_batch_supervisor_slot
  ON exam_import_schedule_rows(batch_id, exam_date, session_label, supervisor_code);
```

Catatan:

- Jika foto hanya punya dua pengawas per ruang, row dibuat dua kali: role `utama` dan `pendamping`.
- Jika waktu ujian tidak terbaca, `session_label` wajib ada, misalnya `pagi`, `sesi_1`, atau `hari_1`. Nanti di-normalisasi saat publish.

### 5.6 Tabel Snapshot Rollback

```sql
CREATE TABLE exam_import_apply_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id UUID NOT NULL REFERENCES exam_import_batches(id) ON DELETE CASCADE,
  snapshot_type TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id UUID,
  before_data JSONB,
  after_data JSONB,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_exam_import_snapshot_type CHECK (
    snapshot_type IN ('master_employee', 'event', 'session', 'room', 'proctor')
  )
);

CREATE INDEX idx_exam_import_snapshots_batch
  ON exam_import_apply_snapshots(batch_id, created_at DESC);
```

### 5.7 Audit Domain Khusus Command Center

Bisa memakai `audit_logs` existing. Jika ingin query lebih cepat dan metadata lebih kaya, tambahkan domain audit:

```sql
CREATE TABLE exam_command_center_audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id UUID REFERENCES exam_import_batches(id) ON DELETE SET NULL,
  event_id UUID REFERENCES cbt_exam_events(id) ON DELETE SET NULL,
  action TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id UUID,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  actor_username_snapshot TEXT NOT NULL DEFAULT '',
  request_id TEXT NOT NULL DEFAULT '',
  source_ip TEXT NOT NULL DEFAULT '',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exam_cc_audit_batch ON exam_command_center_audit_logs(batch_id, created_at DESC);
CREATE INDEX idx_exam_cc_audit_event ON exam_command_center_audit_logs(event_id, created_at DESC);
```

Rekomendasi: tulis ke `audit_logs` umum dan tabel domain khusus untuk investigasi detail.

---

## 6. API Contract Backend

Prefix disarankan: `/api/exam-command-center` atau `/api/asesmen/command-center`.

### 6.1 Event dan Overview

- `GET /api/exam-command-center/events`
  - list kegiatan ujian.
- `POST /api/exam-command-center/events`
  - membuat event `Evaluasi Semester Genap TP 2025/2026`.
- `GET /api/exam-command-center/events/{event_id}/overview`
  - ringkasan status:
    - jumlah sesi,
    - jumlah ruang,
    - ruang tanpa pengawas,
    - konflik jadwal,
    - status publish,
    - audit terbaru.

### 6.2 Import Batch

- `POST /api/exam-command-center/import-batches`
  - body:
    ```json
    {
      "title": "Import Jadwal Pengawas Evaluasi Semester Genap TP 2025/2026",
      "source_type": "photo_ocr",
      "source_label": "foto_jadwal_pengawas_juni_2026",
      "date_start": "2026-06-04",
      "date_end": "2026-06-10"
    }
    ```

- `POST /api/exam-command-center/import-batches/{batch_id}/supervisors`
  - upsert staging 27 pengawas dari foto.

- `POST /api/exam-command-center/import-batches/{batch_id}/schedule-rows`
  - upload CSV/JSON hasil transkripsi jadwal per ruang.
  - format JSON row minimal:
    ```json
    {
      "row_no": 1,
      "exam_date": "2026-06-04",
      "day_name_raw": "Kamis",
      "session_label": "pagi",
      "room_code": "R.I",
      "supervisor_code": 1,
      "supervisor_role": "utama"
    }
    ```

- `POST /api/exam-command-center/import-batches/{batch_id}/validate`
  - jalankan validasi tanpa apply.

- `GET /api/exam-command-center/import-batches/{batch_id}/validation-report`
  - menampilkan error/warning/diff.

### 6.3 Apply dan Rollback

- `POST /api/exam-command-center/import-batches/{batch_id}/apply-master-data`
  - update/link/create `employees` berdasarkan daftar foto.
  - wajib `dry_run=false` dan `confirm_token` atau `expected_validation_hash`.

- `POST /api/exam-command-center/import-batches/{batch_id}/publish-schedule`
  - membuat/memperbarui event/session/room/proctor final.
  - hanya boleh jika batch `validated` atau `applied_master` tanpa blocking errors.

- `POST /api/exam-command-center/import-batches/{batch_id}/rollback`
  - rollback perubahan yang dibuat batch:
    - hapus proctor assignment dari batch,
    - hapus room/session/event yang dibuat batch bila belum dipakai peserta,
    - revert nama/kode pengawas employee dari snapshot,
    - status batch menjadi `rolled_back`.

- `GET /api/exam-command-center/import-batches/{batch_id}/audit`
  - audit lengkap batch.

---

## 7. Format Import Jadwal Pengawas

### 7.1 Format CSV yang Direkomendasikan

Kolom wajib:

- `tanggal` — `YYYY-MM-DD`
- `hari` — nama hari dari foto, untuk validasi
- `sesi` — misalnya `pagi`, `sesi_1`, atau jam ujian
- `ruang` — `R.I` sampai `R.VIII`
- `kode_pengawas` — angka 1–27
- `peran` — `utama` / `pendamping` / `cadangan`

Contoh:

```csv
tanggal,hari,sesi,ruang,kode_pengawas,peran
2026-06-04,Kamis,pagi,R.I,1,utama
2026-06-04,Kamis,pagi,R.I,2,pendamping
2026-06-04,Kamis,pagi,R.II,3,utama
```

### 7.2 Normalisasi Ruang

Mapping ruang foto:

- `R.I` -> `R.I`
- `R.II` -> `R.II`
- `R.III` -> `R.III`
- `R.IV` -> `R.IV`
- `R.V` -> `R.V`
- `R.VI` -> `R.VI`
- `R.VII` -> `R.VII`
- `R.VIII` -> `R.VIII`

Backend harus menerima variasi aman:

- `RI`, `R I`, `R. I` dinormalisasi ke `R.I`.
- `Ruang I` dinormalisasi ke `R.I`.

Tetapi variasi ambigu, misalnya `R.IIII`, harus error.

### 7.3 Normalisasi Hari/Tanggal

Untuk Indonesia/WITA, validasi kalender Gregorian:

- 2026-06-04 = Kamis
- 2026-06-05 = Jumat
- 2026-06-06 = Sabtu
- 2026-06-07 = Minggu, tidak dipakai untuk ujian
- 2026-06-08 = Senin, dipakai untuk baris yang tertulis Senin
- 2026-06-09 = Selasa, dipakai untuk baris yang tertulis Selasa
- 2026-06-10 = Rabu, dipakai untuk baris yang tertulis Rabu

Karena foto awal bermasalah pada **07–09 Juni** dan sudah ada keputusan panitia untuk mengikuti nama hari, importer harus:

- tetap menyimpan `day_name_raw`,
- menghitung `computed_day_name`,
- menghasilkan warning/error jika berbeda,
- tidak boleh publish bila mismatch belum di-acknowledge admin.

Kebijakan rekomendasi:

- menerapkan koreksi final: Senin=2026-06-08, Selasa=2026-06-09, Rabu=2026-06-10.
- menandai 2026-06-07 sebagai libur/tidak ada ujian.
- menyimpan catatan audit bahwa koreksi tanggal mengikuti instruksi panitia.

---

## 8. Validasi Wajib

### 8.1 Validasi Master Pengawas

Blocking error:

- `supervisor_code` kosong.
- kode di luar 1–27 untuk batch ini.
- duplikasi kode dalam master row.
- nama kosong.
- nama master berbeda drastis dari existing employee yang sudah punya kode sama.
- satu employee existing dipetakan ke dua kode berbeda tanpa approval.

Warning:

- nama hanya beda gelar/titik/spasi.
- match confidence rendah.
- employee belum punya NIP/tanggal lahir.
- perlu create employee baru karena tidak ditemukan.

### 8.2 Validasi Jadwal

Blocking error:

- tanggal di luar 2026-06-04 sampai 2026-06-10, kecuali Minggu 2026-06-07 tidak dipakai.
- ruang bukan R.I–R.VIII.
- kode pengawas tidak ada di master batch.
- employee untuk kode belum resolved.
- satu ruang/sesi tidak punya pengawas utama.
- satu ruang/sesi punya lebih dari satu pengawas utama jika kebijakan hanya satu.
- satu pengawas ditempatkan di dua ruang pada tanggal+sesi yang sama.
- duplikasi nomor pengawas 23 pada slot Al-Qur'an Hadis ditandai pending koreksi, bukan blocker awal sesuai arahan panitia.

Warning:

- hari dari foto tidak sama dengan tanggal kalender.
- jadwal pada Minggu/libur selain 2026-06-07 yang memang dikosongkan.
- pengawas yang sama terlalu sering berturut-turut.
- ruang belum ada di `school_rooms`.
- jumlah pengawas per ruang tidak sesuai kebijakan, misalnya kurang dari dua.

### 8.3 Validasi Publish

Sebelum `publish-schedule`:

- event harus ada atau dibuat otomatis dalam transaksi.
- academic year `2025/2026` resolved.
- semua ruang R.I–R.VIII resolved ke `school_rooms` atau dibuat sebagai placeholder dengan `is_exam_eligible=true`.
- tidak ada blocking errors.
- admin menyetujui koreksi tanggal Senin 08, Selasa 09, Rabu 10 Juni 2026; duplikasi nomor 23 dicatat pending koreksi.
- `expected_validation_hash` cocok dengan report terakhir agar tidak apply data yang berubah setelah review.

---

## 9. Strategi Update Nama Pegawai/Guru dari Foto

### 9.1 Prinsip

- `employees.nama` adalah master display untuk banyak modul, jadi update harus hati-hati.
- Foto menjadi sumber operasional, bukan bukti identitas formal.
- Perubahan nama harus reversible.

### 9.2 Matching Strategy

Urutan pencocokan:

1. Jika `exam_supervisor_code` sudah ada, gunakan kode.
2. Exact normalized name:
   - lowercase,
   - hapus titik gelar,
   - collapse whitespace,
   - normalisasi `Muh.`/`Muh`, `S.Pd.I`/`S Pd I`.
3. Trigram/fuzzy search terhadap `employees.nama`.
4. Manual review.

### 9.3 Apply Master Data

Dalam satu transaksi:

- lock batch row `FOR UPDATE`.
- ambil semua master rows.
- untuk setiap row:
  - create snapshot `before_data` employee.
  - update `employees.exam_supervisor_code`.
  - update `employees.nama` hanya jika action `update_employee_name` disetujui.
  - set `display_name_source`, `name_verified_at`, `name_verified_by`.
  - create audit log.
- update status batch ke `applied_master`.

Untuk 27 nama foto, rekomendasi awal:

- Jika employee existing sudah ada tapi beda format gelar kecil, lakukan `update_employee_name`.
- Jika tidak ada, buat employee minimal:
  - `nama`,
  - `unit_kerja = 'MTsN 2 Kolaka Utara'`,
  - `employment_type = 'honorer'` atau `unknown` sesuai enum/validasi existing,
  - `is_active = true`,
  - `exam_supervisor_code`.
- Jangan mengisi NIP palsu.

---

## 10. Strategi Publish Jadwal Final

### 10.1 Event

Buat atau gunakan event:

- `title`: `Evaluasi Semester Genap TP 2025/2026`
- `exam_type`: karena enum existing tidak punya `semester`, gunakan `uas` atau `lainnya` sesuai kebijakan sekolah. Rekomendasi teknis aman: `uas` bila evaluasi semester setara PAS/SAS, atau `lainnya` jika hanya command center non-CBT.
- `scope`: `school`
- `academic_year_id`: TP 2025/2026
- `status`: `draft` saat import, `active` setelah publish.

### 10.2 Sessions

Untuk setiap kombinasi tanggal+sesi:

- buat `cbt_exam_sessions` jika belum ada.
- nama sesi: `Evaluasi Semester Genap 2025/2026 - 2026-06-04 pagi`.
- simpan relasi event via `event_id`.

Jika `cbt_exam_sessions` membutuhkan `package_id`/field tertentu dari schema existing, buat mode Command Center yang hanya membuat sesi saat paket tersedia, atau tambahkan kolom nullable secara migration terpisah setelah review schema lengkap.

### 10.3 Rooms

Untuk setiap room `R.I`–`R.VIII` per session:

- resolve/create `school_rooms`.
- buat `cbt_exam_rooms`:
  - `session_id`,
  - `school_room_id`,
  - `room_name = room_code`,
  - `room_name_snapshot`,
  - `capacity` dari `school_rooms.exam_capacity`,
  - `status = 'ready'` setelah semua pengawas valid.

### 10.4 Proctors

Untuk setiap row jadwal:

- insert `cbt_room_proctors`:
  - `exam_room_id`,
  - `employee_id`,
  - `role`: `utama`/`pendamping`/`cadangan`,
  - `assigned_by`: admin.
- constraint existing mencegah employee sama double di room yang sama.
- service harus menambah validasi cross-room sebelum insert.

### 10.5 Idempotency

`publish-schedule` harus idempotent:

- gunakan `batch_id` dan snapshot untuk tahu entity yang dibuat.
- sebelum insert, cari existing berdasarkan event+session+room+employee.
- jika sama, skip dan audit `noop`.
- jika beda, require explicit `replace_existing=true`.

---

## 11. RBAC dan Keamanan

Permission baru yang disarankan:

- `exam_command_center.read`
- `exam_command_center.import`
- `exam_command_center.validate`
- `exam_command_center.apply_master`
- `exam_command_center.publish`
- `exam_command_center.rollback`
- `exam_command_center.audit`

Migration permission wajib grant ke `admin` dalam migration yang sama.

Role rekomendasi:

- `admin`: semua permission.
- `staf` atau panitia ujian: read, import, validate, audit terbatas.
- `guru`: read assignment miliknya saja, bukan apply/publish.

Keamanan API:

- Semua mutasi harus authenticated.
- Semua body JSON dibatasi ukuran dengan `http.MaxBytesReader`.
- Upload CSV dibatasi ukuran dan MIME/extension allowlist.
- Request mutasi mencatat `request_id`, `source_ip`, actor username snapshot.

---

## 12. Audit-Safe Rollout

### 12.1 Phase 0 — Design Lock

- Finalisasi format CSV impor jadwal.
- Terapkan kebijakan hari/tanggal final: Minggu 07 Juni 2026 libur, Senin 08, Selasa 09, Rabu 10 Juni 2026.
- Duplikasi nomor 23 di foto dicatat sebagai pending koreksi panitia besok, belum diganti otomatis.

### 12.2 Phase 1 — Migration Additive

Deploy migration additive saja:

- tambah kolom nullable ke `employees`,
- tambah tabel import/staging/snapshot/audit,
- tambah RBAC permission.

Tidak mengubah flow existing.

Rollback phase ini:

- drop tabel baru,
- drop kolom baru jika belum dipakai,
- hapus permission baru.

### 12.3 Phase 2 — Backend Dry-Run

Deploy endpoint import dan validate:

- belum ada apply final.
- admin bisa upload/paste data.
- hasil hanya validation report.

Rollback:

- matikan route dengan feature flag.
- data staging bisa tetap disimpan atau dihapus batch.

### 12.4 Phase 3 — Apply Master Data Terbatas

Aktifkan `apply-master-data`:

- hanya admin.
- wajib validation hash.
- create snapshot before/after.
- audit every employee touched.

Rollback:

- endpoint rollback employee dari `exam_import_apply_snapshots`.
- revert `nama`, `exam_supervisor_code`, `display_name_source`, `name_verified_at`, `name_verified_by`.

### 12.5 Phase 4 — Publish Jadwal Terbatas

Aktifkan `publish-schedule`:

- hanya untuk satu event.
- publish ke draft/ready rooms.
- tidak mengganggu peserta existing.

Rollback:

- hapus `cbt_room_proctors` yang dibuat batch.
- hapus `cbt_exam_rooms` yang dibuat batch jika belum ada peserta.
- hapus `cbt_exam_sessions` yang dibuat batch jika belum ada peserta.
- event dikembalikan `draft` atau dihapus jika hanya dibuat batch dan belum punya relasi lain.

### 12.6 Phase 5 — UI Command Center

- tampilkan readiness dashboard.
- import wizard.
- validation report.
- approval warning.
- audit timeline.
- rollback button hanya untuk admin.

---

## 13. Rollback Detail

### 13.1 Rollback Master Data

Syarat:

- batch status `applied_master` atau `published`.
- tidak ada batch lain setelahnya yang mengubah employee sama, atau perlu conflict review.

Algoritma:

1. Lock batch.
2. Ambil snapshot `snapshot_type='master_employee'` urut newest-first.
3. Untuk tiap employee:
   - bandingkan current data dengan `after_data`.
   - jika current != after_data, tandai conflict dan jangan rollback otomatis.
   - jika sama, restore `before_data`.
4. Audit `rollback_master_employee`.
5. Set batch `rolled_back` bila schedule juga sudah rollback.

### 13.2 Rollback Jadwal

Syarat:

- belum ada peserta submit/aktivitas runtime penting pada session terkait.
- jika sudah ada peserta/hasil, rollback tidak boleh hard delete; gunakan soft cancel/revision event.

Algoritma:

1. Lock batch.
2. Ambil snapshot `proctor`, hapus proctor assignment yang dibuat batch.
3. Ambil snapshot `room`, hapus room yang dibuat batch jika tidak punya peserta.
4. Ambil snapshot `session`, hapus session yang dibuat batch jika aman.
5. Event kembali `draft` atau `finished/cancelled` sesuai status existing.
6. Audit `rollback_schedule`.

### 13.3 Rollback dengan Batch Koreksi

Jika data sudah dipakai ujian, jangan rollback fisik. Buat **correction batch**:

- batch baru dengan `rollback_of_batch_id`.
- perubahan bersifat append-only.
- audit menghubungkan batch lama dan koreksi.

---

## 14. Rencana Implementasi Teknis Detail

### Backend

1. Buat migration `120_exam_command_center.sql`.
2. Tambahkan query sqlc:
   - create/list/get import batch,
   - insert/update supervisor master rows,
   - insert schedule rows,
   - list validation rows,
   - create snapshots,
   - create audit domain log,
   - update employee by exam supervisor code,
   - resolve school rooms.
3. Regenerate sqlc dengan `make db-sqlc`.
4. Buat service:
   - `CreateImportBatch`
   - `ImportSupervisorMasterRows`
   - `ImportScheduleRows`
   - `ValidateImportBatch`
   - `ApplyMasterData`
   - `PublishSchedule`
   - `RollbackBatch`
5. Buat handler thin:
   - parse body,
   - call service,
   - map error ke HTTP.
6. Tambahkan route dan RBAC guard.
7. Tambahkan tests:
   - validasi tanggal/hari,
   - duplikasi kode 23,
   - kode pengawas tidak ada,
   - double booking pengawas,
   - rollback employee snapshot,
   - idempotent publish.

### Frontend/BFF

1. Tambah proxy endpoint `/api/command-center/*` ke Core API.
2. Tambah halaman wizard:
   - create/select event,
   - paste/upload master pengawas,
   - upload jadwal,
   - validate,
   - review diff,
   - apply master,
   - publish,
   - audit/rollback.
3. Gunakan skeleton/AsyncContent dan toast Sonner sesuai baseline repo.
4. Jangan ada DB direct access.

### Operasional

1. Siapkan CSV hasil transkripsi jadwal foto.
2. Jalankan dry-run validasi.
3. Terapkan koreksi tanggal yang sudah disetujui dan tampilkan duplikasi nomor 23 sebagai catatan pending koreksi.
4. Apply master data.
5. Publish jadwal.
6. Export report audit sebagai arsip.

---

## 15. Risiko dan Mitigasi

- Risiko: OCR/foto salah membaca nomor pengawas.
  - Mitigasi: staging + validation + manual approval.

- Risiko: update nama pegawai merusak modul lain.
  - Mitigasi: snapshot before/after, reversible, dan hanya field nama/kode yang diubah.

- Risiko: duplikasi nomor 23 menyebabkan pengawas double booking; mitigasi sementara: tampil sebagai warning pending koreksi, bukan final locked.
  - Mitigasi: blocking validation pada tanggal+sesi sama.

- Risiko: tanggal dari foto salah; mitigasi: gunakan koreksi final Senin=08, Selasa=09, Rabu=10 Juni 2026.
  - Mitigasi: computed calendar validation dan approval eksplisit.

- Risiko: publish ganda membuat row dobel.
  - Mitigasi: idempotency key batch + uniqueness application-level + constraints existing.

- Risiko: rollback setelah ujian berjalan menghapus evidence.
  - Mitigasi: hard rollback hanya sebelum runtime; setelah runtime gunakan correction batch.

---

## 16. Definition of Done

Backend selesai jika:

- migration additive berhasil,
- sqlc generated,
- semua endpoint ada dan protected RBAC,
- dry-run validation mengeluarkan blocking errors/warnings jelas,
- apply master data membuat snapshot dan audit,
- publish schedule idempotent,
- rollback teruji,
- tests pass.

Data siap publish jika:

- 27 master pengawas resolved ke employees,
- ruang R.I–R.VIII resolved,
- semua jadwal tanggal 4–6 dan 8–10 Juni 2026 tervalidasi,
- koreksi tanggal Senin dan seterusnya sudah diterapkan,
- duplikasi nomor 23 diberi status pending koreksi panitia dan tidak menghalangi import awal,
- tidak ada double booking pengawas.

---

## 17. Rekomendasi Keputusan Sebelum Implementasi

Sebelum coding, kepala/admin perlu memutuskan:

1. Duplikasi nomor pengawas 23 akan diganti ke nomor berapa saat panitia mengoreksi besok?
2. Apakah setelah koreksi nomor 23 jadwal langsung dikunci final?
3. Apakah event ini harus memakai `exam_type='uas'` atau `exam_type='lainnya'`?
4. Apakah pengawas per ruang wajib dua orang atau boleh satu?
5. Apakah nama dari foto boleh langsung mengganti `employees.nama`, atau hanya menjadi alias/display khusus ujian?

Rekomendasi paling aman: **pakai staging + alias/kode pengawas lebih dulu, lalu update `employees.nama` hanya setelah admin menyetujui diff per pegawai.**
