# Structured Logging Monorepo

Dokumen ini menjadi runbook logger standar untuk MTsN 2 Kolut setelah patch Bank Soal save error.

## Tujuan

- Satu request bisa ditelusuri dari SvelteKit web-admin ke Go core-api memakai `x-request-id`.
- Error internal tetap masuk log server, tetapi pesan user tetap aman/generik.
- Payload yang dicatat hanya metadata aman: ID referensi, tipe, status, jumlah opsi, dan panjang teks; bukan isi soal penuh, token, password, cookie, atau connection string.

## Format log standar

Semua service menulis JSON per baris dengan field utama:

- `time`: timestamp ISO atau timestamp logger runtime.
- `service`: `core-api`, `web-admin`, atau `pusaka-worker`.
- `level`: `info`, `warn`, `error`.
- `event`: nama event stabil, contoh `cbt_question_create_db_failed`.
- `request_id`: ID korelasi jika request HTTP.
- `module`: modul aplikasi, contoh `bank-soal`.
- `status`, `method`, `route`/`path`: untuk request HTTP/proxy.
- `error_kind`, `sqlstate`, `constraint`, `table`, `column`: hanya di log server untuk error PostgreSQL.

## Bank Soal save tracing

Ketika simpan soal gagal:

1. Ambil `x-request-id` dari response browser atau log web-admin.
2. Cari log web-admin:
   ```bash
   pm2 logs mtsn2kolut-web-admin --lines 300 | grep '<request_id>'
   ```
3. Cari log core-api:
   ```bash
   pm2 logs mtsn2kolut-core-api --lines 500 | grep '<request_id>'
   ```
4. Event penting:
   - `bank_soal_questions_upstream_error` di web-admin.
   - `cbt_question_create_request_received` di core-api.
   - `cbt_question_create_reference_invalid` jika `subject_id`/`event_id` stale/tidak ada.
   - `cbt_question_create_db_failed` jika gagal di DB; lihat `sqlstate` dan `constraint`.
   - `cbt_question_create_success` jika sukses.

## Redaction policy

Wajib `[REDACTED]` untuk field/key yang mengandung:

- `authorization`, `cookie`, `password`, `passwd`
- `token`, `secret`, `api_key`
- `database_url`, `connection_string`, `jwt`

Pusaka Worker meng-hash `username`/`pusaka_username` menjadi `username_hash` agar absensi tetap bisa ditelusuri tanpa membuka identitas login mentah.

## Patch bug simpan soal

Root problem class yang ditutup patch ini: UI bisa membawa metadata tersimpan/stale sehingga `subject_id` atau `event_id` lolos readiness check frontend tetapi gagal foreign-key/check di backend dan sebelumnya dimask menjadi “Pembuatan soal CBT tidak valid”.

Core-api sekarang memvalidasi referensi sebelum insert/update:

- `subject_id` harus ada di tabel `subjects`.
- Jika `event_id` dikirim, event harus ada di `cbt_exam_events`.
- Jika invalid, response user tetap aman tetapi actionable:
  - `subject_id tidak ditemukan. Pilih ulang mapel dari daftar terbaru`
  - `event_id tidak ditemukan. Matikan mode khusus kegiatan atau pilih ulang kegiatan`

## Operator checklist setelah deploy

- Health core-api OK.
- Web-admin halaman root OK.
- `npm check`, Go tests/build sudah PASS sebelum restart.
- Coba simpan satu soal Bank Soal.
- Bila gagal, gunakan `request_id` untuk cek log boundary web-admin dan core-api.
