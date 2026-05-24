# Runbook Opsi 0 Asesmen/CBT E2E

Dokumen ini menjelaskan cara menjalankan smoke end-to-end **Opsi 0** untuk modul Asesmen/CBT web-admin. Opsi 0 adalah gladi teknis otomatis sebelum Opsi A/B/C simulasi siswa, dengan fokus utama pada keamanan jalur login role, route Asesmen, proxy ujian `/api/exam/*`, masking token, dan redaksi kunci jawaban.

## Tujuan

Opsi 0 dipakai untuk memastikan alur kritis Asesmen/CBT siap diuji lebih lanjut tanpa bergantung pada data simulasi penuh. Mode awal bersifat aman:

- tidak memakai data produksi;
- tidak melakukan submit jawaban final;
- tidak melakukan scoring;
- tidak melakukan regenerate token produksi;
- tidak menjalankan endpoint mutasi destruktif kecuali mode destructive diaktifkan secara eksplisit pada database test yang resettable.

## Prinsip keselamatan data

Selalu gunakan database test terisolasi melalui `TEST_DATABASE_URL`. Jangan menjalankan Opsi 0 destructive pada database production, staging yang berisi data nyata, atau database yang tidak bisa di-reset.

Contoh persiapan database lokal/resettable:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
TEST_DATABASE_URL='[REDACTED_TEST_DATABASE_URL]' scripts/setup-cbt-test-db.sh --reset
```

Aturan wajib:

- `TEST_DATABASE_URL` harus menunjuk ke database khusus test.
- Mode default adalah non-destructive.
- `WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=false` tidak boleh memanggil `/api/exam/answer` atau `/api/exam/submit`.
- `WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=true` hanya boleh digunakan bersama `WEB_ADMIN_ASESMEN_CBT_OPSI0_CONFIRM_TEST_DB=true`.

## Kontrak environment

Nilai rahasia tidak boleh dicommit. Simpan di shell lokal, CI secret, atau mekanisme secret manager.

```bash
WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL=http://127.0.0.1:8021
WEB_ADMIN_ASESMEN_CBT_OPSI0_ADMIN_USERNAME='[REDACTED_EXAMPLE_USERNAME]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_ADMIN_PASSWORD='[REDACTED_EXAMPLE_PASSWORD]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_PROCTOR_USERNAME='[REDACTED_EXAMPLE_USERNAME]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_PROCTOR_PASSWORD='[REDACTED_EXAMPLE_PASSWORD]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_GURU_USERNAME='[REDACTED_EXAMPLE_USERNAME]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_GURU_PASSWORD='[REDACTED_EXAMPLE_PASSWORD]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_EXAM_TOKEN='[REDACTED_EXAMPLE_EXAM_TOKEN]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_ROOM_TOKEN='[REDACTED_EXAMPLE_ROOM_TOKEN]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_BAD_ROOM_TOKEN='[REDACTED_EXAMPLE_BAD_ROOM_TOKEN]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_SESSION_ID=optional
WEB_ADMIN_ASESMEN_CBT_OPSI0_EVENT_ID=optional
WEB_ADMIN_ASESMEN_CBT_OPSI0_HEADLESS=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_TIMEOUT_MS=20000
WEB_ADMIN_ASESMEN_CBT_OPSI0_IGNORE_HTTPS_ERRORS=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_ALLOW_SKIP=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=false
```

Perilaku env:

- Jika env belum lengkap dan `WEB_ADMIN_ASESMEN_CBT_OPSI0_ALLOW_SKIP=true`, script boleh keluar `0` dengan pesan `SKIP`.
- Jika env belum lengkap dan `WEB_ADMIN_ASESMEN_CBT_OPSI0_ALLOW_SKIP=false`, script harus gagal non-zero.
- Strict mode (`ALLOW_SKIP=false`) membutuhkan env lengkap untuk safe smoke: admin, proktor, guru, `EXAM_TOKEN`, dan `ROOM_TOKEN`.
- Jika Playwright belum tersedia dan skip diizinkan, script boleh keluar `0` dengan pesan `SKIP`.

## Cara menjalankan

Dari root monorepo:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run smoke:asesmen-cbt:opsi0
```

Jika memakai shell lokal dengan env lengkap:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
export TEST_DATABASE_URL='[REDACTED_TEST_DATABASE_URL]'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL='http://127.0.0.1:8021'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_ADMIN_USERNAME='...'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_ADMIN_PASSWORD='...'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_PROCTOR_USERNAME='...'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_PROCTOR_PASSWORD='...'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_EXAM_TOKEN='...'
export WEB_ADMIN_ASESMEN_CBT_OPSI0_ROOM_TOKEN='...'
npm --prefix apps/web-admin run smoke:asesmen-cbt:opsi0
```

## Yang diuji pada safe mode

Safe mode adalah mode default ketika `WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=false`.

Pemeriksaan yang diharapkan:

- Login admin/proktor/guru seed melalui UI web-admin.
- Route Asesmen utama reachable atau forbidden sesuai role:
  - `/asesmen`
  - `/asesmen/persiapan`
  - `/asesmen/pelaksanaan`
  - `/asesmen/pengawasan`
  - `/asesmen/hasil`
  - `/ujian/command-center`
- Endpoint protected tanpa autentikasi ditolak.
- `/api/exam/status` tanpa token/fingerprint ditolak.
- `POST /api/exam/login` dengan token seed berhasil.
- Response login ujian tidak membocorkan kunci jawaban atau marker benar/salah. Implementasi script harus memakai deep scanner seperti `assertNoSensitiveExamKeys`.
- Header response exam proxy memakai `cache-control: no-store`.
- Token peserta tidak muncul mentah pada teks yang seharusnya masked.
- Room token salah tidak membocorkan token peserta.

## Yang diuji pada destructive mode

Destructive mode hanya untuk database test yang resettable. Aktifkan hanya jika semua gate terpenuhi:

```bash
export WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=true
export WEB_ADMIN_ASESMEN_CBT_OPSI0_CONFIRM_TEST_DB=true
export TEST_DATABASE_URL='[REDACTED_TEST_DATABASE_URL]'
```

Pemeriksaan tambahan yang dapat dilakukan pada Opsi 0.2 lanjutan (saat ini script Opsi 0.1 hanya memasang guard dan belum mengeksekusi `/api/exam/answer` atau `/api/exam/submit`):

- Login ujian siswa seed.
- Ambil status/daftar soal.
- Kirim heartbeat.
- Catat event proctoring ringan.
- Jawab satu soal pilihan ganda dan satu essay melalui `/api/exam/answer`.
- Submit melalui `/api/exam/submit`.
- Submit ulang harus ditolak, misalnya `409`.
- Admin/proktor membuka dashboard sesi atau hasil.

Jangan gunakan mode ini pada data produksi.

## Contoh output

Env atau Playwright belum tersedia, skip diizinkan:

```text
SKIP Asesmen CBT Opsi 0 E2E: Playwright/env belum tersedia.
```

Safe mode berhasil:

```text
PASS role route smoke
PASS exam proxy safe smoke
PASS no sensitive exam keys found
```

Destructive mode ditolak karena belum konfirmasi DB test:

```text
FAIL destructive mode requires WEB_ADMIN_ASESMEN_CBT_OPSI0_CONFIRM_TEST_DB=true
```

## Troubleshooting

### Script tidak ditemukan

Pastikan script Opsi 0 sudah tersedia dan npm script `smoke:asesmen-cbt:opsi0` mengarah ke:

```bash
node scripts/asesmen-cbt-opsi0-e2e.mjs
```

### Dependency Playwright belum tersedia

Script mengikuti pola smoke web-admin yang melakukan dynamic import `playwright` lalu fallback ke `@playwright/test`. Jika dependency belum ada dan skip diizinkan, hasil yang benar adalah `SKIP`, bukan crash. Jika ingin menjalankan browser sungguhan, pasang dependency sesuai kebijakan repo dan environment CI/lokal.

### Login role gagal

Periksa:

- base URL `WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL` sudah menunjuk ke web-admin yang aktif;
- kredensial seed admin/proktor/guru benar;
- backend core-api terhubung ke database test yang sama;
- cookie/session domain tidak tertukar antar environment.

### `/api/exam/login` gagal

Periksa:

- `WEB_ADMIN_ASESMEN_CBT_OPSI0_EXAM_TOKEN` dan `WEB_ADMIN_ASESMEN_CBT_OPSI0_ROOM_TOKEN` berasal dari seed test;
- sesi ujian seed aktif pada waktu eksekusi;
- fingerprint/header yang dikirim script sesuai kontrak BFF;
- backend menggunakan database test, bukan production.

## Catatan untuk CI

Contract test Vitest menjaga agar script, docs, dan npm script tetap discoverable. Test tersebut sengaja memeriksa string penting seperti `WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL`, `/api/exam/login`, `assertNoSensitiveExamKeys`, `WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE`, `TEST_DATABASE_URL`, dan frasa `tidak memakai data produksi`.
