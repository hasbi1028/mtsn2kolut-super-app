# Plan: Hapus Data Seed Asesmen Manual APK

## Tujuan
Membersihkan data seed/uji manual asesmen APK yang dibuat untuk pengujian CBT Mobile, tanpa menyentuh data asesmen produksi.

## Backup Database
Backup database sudah dibuat sebelum deletion plan ini dijalankan.

- File backup: `/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pusaka_20260510_214125.dump`
- Symlink latest: `/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/latest.dump`
- SHA256: `31550373115fbfede6af71e4b3dad459c25e1ec97ab15c1f27c3a339968f8ed9`
- Format: PostgreSQL custom-format dump (`pg_dump --format=custom`)

Restore darurat jika diperlukan:

```bash
pg_restore --clean --if-exists --dbname "$DATABASE_URL" /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pusaka_20260510_214125.dump
```

> Catatan: command restore di atas bersifat destruktif dan hanya untuk kondisi darurat setelah dipastikan downtime/rollback scope.

## Target yang akan dihapus
Berdasarkan inspeksi database, data seed yang ditemukan:

- Student test:
  - `students.id = 9d1fc75b-821f-4cd2-bf40-5a74a7a26f46`
  - `nis = TEST-APK-001`
  - `nama = SISWA TES APK`
- Session test:
  - `cbt_exam_sessions.id = 2800c76a-33f4-4676-a0b9-8d77aadd54f4`
  - `title = TES MANUAL APK - AKTIF`
- Package test:
  - `cbt_packages.id = 144638c1-da8d-4ace-be2c-b6d7cfb9ef22`
  - `title = Paket Tes Manual APK CBT`
- Questions test:
  - `cbt_questions.id = 9a996317-16c1-464b-aa01-ee0f31ba4e4b`, `code = TES-APK-001`
  - `cbt_questions.id = cdae8cdd-db17-4ad1-8fc3-f57f5ce6875f`, `code = TES-APK-002`

## Ringkasan dependensi yang ditemukan

- `cbt_exam_rooms`: 1 row
- `cbt_exam_participants`: 1 row
- `cbt_participant_events`: 12 rows
- `cbt_student_answers`: 2 rows
- `cbt_package_questions`: 2 rows
- `parent_students`: 0 rows
- `users linked to student`: 0 rows
- `student_certificates`: 0 rows
- `student_transfers`: 0 rows
- `grade_entries`: 0 rows
- Questions referenced outside test package: 0 rows
- Student answers for test questions outside test participant: 0 rows

## Deletion order
Deletion harus dilakukan dalam transaksi dan urutan aman:

1. Lock/resolve target rows by exact identifiers:
   - `students.nis = 'TEST-APK-001'`
   - `cbt_exam_sessions.title ILIKE '%TES MANUAL APK%'`
   - package linked from target session or title `Paket Tes Manual APK CBT`
   - question codes `TES-APK-001`, `TES-APK-002`
2. Delete child rows first:
   - `cbt_student_answers` for target participant
   - `cbt_participant_events` for target participant
   - `cbt_exam_participants` for target session/student
   - `cbt_exam_rooms` for target session
   - `cbt_package_questions` for target package
3. Delete parent rows:
   - `cbt_exam_sessions`
   - `cbt_packages`
   - `cbt_questions` with codes `TES-APK-001`, `TES-APK-002`
   - `students` where `nis = 'TEST-APK-001'`
4. Verify all target counts are zero.
5. Commit only if verification is clean; otherwise rollback.

## Proposed SQL transaction

```sql
BEGIN;

WITH
  test_student AS (
    SELECT id FROM students WHERE nis = 'TEST-APK-001'
  ),
  test_sessions AS (
    SELECT id, package_id FROM cbt_exam_sessions
    WHERE title ILIKE '%TES MANUAL APK%'
       OR id IN (SELECT session_id FROM cbt_exam_participants WHERE student_id IN (SELECT id FROM test_student))
  ),
  test_participants AS (
    SELECT id FROM cbt_exam_participants
    WHERE student_id IN (SELECT id FROM test_student)
       OR session_id IN (SELECT id FROM test_sessions)
  ),
  test_packages AS (
    SELECT id FROM cbt_packages
    WHERE id IN (SELECT package_id FROM test_sessions)
       OR title = 'Paket Tes Manual APK CBT'
  ),
  test_questions AS (
    SELECT id FROM cbt_questions
    WHERE code IN ('TES-APK-001', 'TES-APK-002')
       OR id IN (SELECT question_id FROM cbt_package_questions WHERE package_id IN (SELECT id FROM test_packages))
  ),
  del_answers AS (
    DELETE FROM cbt_student_answers WHERE participant_id IN (SELECT id FROM test_participants) RETURNING 1
  ),
  del_events AS (
    DELETE FROM cbt_participant_events WHERE participant_id IN (SELECT id FROM test_participants) RETURNING 1
  ),
  del_participants AS (
    DELETE FROM cbt_exam_participants WHERE id IN (SELECT id FROM test_participants) RETURNING 1
  ),
  del_rooms AS (
    DELETE FROM cbt_exam_rooms WHERE session_id IN (SELECT id FROM test_sessions) RETURNING 1
  ),
  del_pkg_questions AS (
    DELETE FROM cbt_package_questions WHERE package_id IN (SELECT id FROM test_packages) RETURNING 1
  ),
  del_sessions AS (
    DELETE FROM cbt_exam_sessions WHERE id IN (SELECT id FROM test_sessions) RETURNING 1
  ),
  del_packages AS (
    DELETE FROM cbt_packages WHERE id IN (SELECT id FROM test_packages) RETURNING 1
  ),
  del_questions AS (
    DELETE FROM cbt_questions WHERE id IN (SELECT id FROM test_questions) RETURNING 1
  ),
  del_students AS (
    DELETE FROM students WHERE id IN (SELECT id FROM test_student) RETURNING 1
  )
SELECT
  (SELECT count(*) FROM del_answers) AS deleted_answers,
  (SELECT count(*) FROM del_events) AS deleted_events,
  (SELECT count(*) FROM del_participants) AS deleted_participants,
  (SELECT count(*) FROM del_rooms) AS deleted_rooms,
  (SELECT count(*) FROM del_pkg_questions) AS deleted_package_questions,
  (SELECT count(*) FROM del_sessions) AS deleted_sessions,
  (SELECT count(*) FROM del_packages) AS deleted_packages,
  (SELECT count(*) FROM del_questions) AS deleted_questions,
  (SELECT count(*) FROM del_students) AS deleted_students;

-- Verification query must return zero counts before COMMIT.
-- COMMIT;
-- ROLLBACK; -- if any count unexpected
```

## Validation setelah deletion

Run verification:

```sql
SELECT count(*) FROM students WHERE nis='TEST-APK-001' OR nama ILIKE '%SISWA TES APK%';
SELECT count(*) FROM cbt_exam_sessions WHERE title ILIKE '%TES MANUAL APK%';
SELECT count(*) FROM cbt_packages WHERE title = 'Paket Tes Manual APK CBT';
SELECT count(*) FROM cbt_questions WHERE code IN ('TES-APK-001','TES-APK-002');
```

Expected: semua `0`.

## Risiko / batasan

- Deletion bersifat permanen kecuali restore dari backup.
- Jika ternyata ada data uji lain yang tidak memakai pola `TEST-APK-001`, `TES MANUAL APK`, atau `TES-APK-*`, data itu tidak akan ikut terhapus.
- File APK Release Center dan patch aplikasi tidak dihapus; plan ini hanya data seed asesmen di PostgreSQL.

## Status
Belum menjalankan deletion. Menunggu konfirmasi eksplisit dari pemilik sistem.
