# CBT Proposal Integration Phase 5

Status: rehearsal, rollout, and post-exam review guard, 2026-05-08.

Dokumen ini menutup integrasi proposal Sistem CBT MTsN 2 Kolaka Utara pada Phase 5 - Rehearsal, Rollout, and Post-Exam Review. Scope fase ini adalah kesiapan operasional yang dapat diaudit, bukan schema change, deploy automation, atau penggantian runtime.

Alur rehearsal yang wajib diuji adalah Bank Soal -> Asesmen Persiapan -> Pelaksanaan/Pengawasan -> Flutter APK -> Hasil/Post-exam review.

## Dokumen Terkait

- `docs/cbt-smoke-checklist.md`: checklist rehearsal end-to-end dan acceptance evidence utama.
- `docs/exam-api.md`: kontrak runtime mobile untuk `POST /api/exam/login`, `GET /api/exam/status`, `POST /api/exam/heartbeat`, `POST /api/exam/event`, `POST /api/exam/answer`, dan `POST /api/exam/submit`.
- `apps/mobile/RELEASE_CHECKLIST.md`: quality gate dan distribusi internal APK Flutter BYOD.
- `apps/mobile/DEVICE_TEST_MATRIX.md`: catatan perangkat nyata saat rehearsal.

## Boundary Phase 5

- Tidak deploy, tidak PM2 restart, dan tidak mengubah topology 3 VPS dari pekerjaan Phase 5 ini.
- Tidak menjalankan `make db-migrate`, migrasi live, atau ad hoc `ALTER TABLE`; migration state hanya dicek sebagai preflight.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru. Runtime siswa tetap menggunakan `/api/exam/*`; `/api/cbt/*` hanya compatibility/deprecated path yang sudah ada untuk transisi web-admin lama.
- Tidak mengubah runtime architecture ke PocketBase, SQLite, Alpine, Docker runtime, atau service CBT baru.
- Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`, bukan melalui SvelteKit BFF.
- `apps/web-admin` tetap UI/BFF untuk Bank Soal, Asesmen, Pengawasan, dan Hasil; tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, audit log, server log, scoring, submit, dan event exam.
- BYOD tidak setara kiosk penuh; device-owner bukan baseline untuk perangkat siswa pribadi.

## Preflight Rehearsal

Operator teknis dan panitia harus mencatat hasil berikut sebelum rehearsal dibuka:

- health check backend dan database pool sehat.
- audit log dapat dibaca oleh role berwenang dan menerima mutasi rehearsal.
- server log backend/frontend tersedia untuk investigasi dan tidak memuat panic berulang, secret, password, token mentah, atau answer key.
- backup PostgreSQL terbaru sudah ada, timestamp dan lokasi restore tercatat.
- migration state sesuai release yang diuji; bila pending/tidak jelas, rehearsal ditunda.
- commit backend, web-admin, dan APK Flutter dicatat.

Preflight ini bersifat verifikasi. Phase 5 tidak menjadi instruksi untuk menjalankan deploy, PM2 restart, atau migrasi live.

## Rehearsal End-to-End

1. Bank Soal: soal dibuat/diimpor melalui `/bank-soal/tambah` atau `/bank-soal/impor`, diverifikasi di `/bank-soal/verifikasi`, lalu dicek agar kunci jawaban tidak bocor lintas role.
2. Asesmen Persiapan: paket, kegiatan, sesi, peserta, ruang, seat, pengawas, token, dan kartu ujian disiapkan melalui route canonical `/asesmen/*`.
3. Pelaksanaan/Pengawasan: dashboard `/asesmen/pelaksanaan` dan `/asesmen/pengawasan` memantau login, heartbeat, warning BYOD, status submit, dan ruang/seat.
4. Flutter APK: perangkat nyata menguji login token, heartbeat, answer save, restore, network disturbance, warning, dan submit sesuai `apps/mobile/RELEASE_CHECKLIST.md`; asset soal serta endpoint `answer`/`submit`/`heartbeat`/`event` tetap harus menolak request tanpa konteks peserta sah.
5. Hasil/Post-exam review: `/asesmen/hasil` dipakai untuk scoring, export, review event anti-cheat, audit log, dan temuan pengawas.

Checklist smoke juga mempertahankan guard operasional lama: event member role dan subject scope, export/detail Bank Soal admin/guru/guru non-penulis, berita acara/minutes token visibility, duplicate `(room_id, seat_no)`, serta endpoint/security boundary checks.

Flutter SDK bisa tidak tersedia di host yang menjalankan rehearsal checklist. Bila begitu, quality gate mobile (`flutter analyze`, `flutter test`, dan build APK) harus dicatat sebagai blocker validasi mobile di host tersebut dan dijalankan di mesin lain/CI sebelum APK dipakai untuk ujian resmi.

## Rollout Decision

Keputusan rollout harus memakai evidence dari `docs/cbt-smoke-checklist.md`.

Lulus bila:

- Minimal satu rehearsal end-to-end lulus dari Bank Soal sampai Hasil/Post-exam review.
- Health check, audit log, server log, backup, dan migration state sudah dicatat.
- Minimal dua perangkat nyata lintas vendor lulus login token, heartbeat, answer save, restore, network disturbance, warning, dan submit.
- Flutter payload/event tidak memuat token mentah, password, atau answer key.
- Tidak ada role yang melihat token/kunci jawaban lintas scope.

Tunda bila:

- Migration state tidak jelas.
- Token, answer key, jawaban, submit, scoring, atau role scope tidak dapat dipercaya.
- Flutter tidak bisa login, menyimpan jawaban, restore aman, atau submit pada koneksi sehat.
- Pengawas tidak dapat melihat status/warning yang diperlukan untuk operasi hari ujian.

## Rollback Plan

Rollback plan wajib tersedia sebelum rehearsal dianggap selesai:

- Jika runtime ujian bermasalah, hentikan sesi baru dan jangan buka gelombang berikutnya.
- Langkah utama rollback adalah pertahankan data backend: jawaban, token, audit, event, dan server log tidak boleh dihapus untuk "membersihkan" rehearsal.
- Jika APK release bermasalah, distribusikan APK sebelumnya yang sudah lulus rehearsal dan ulangi uji perangkat nyata.
- Jika web-admin bermasalah tetapi backend/Flutter tetap menjaga answer save dan submit, tahan mutasi operasional baru dan eskalasi panitia/teknis.
- Jika backend bermasalah, ambil evidence health check, audit log, server log, backup, dan migration state sebelum perubahan perbaikan.
- Jangan rollback database dengan ad hoc SQL; gunakan prosedur release/restore resmi yang disetujui penanggung jawab backend.

## Guard dan Validasi

Validasi minimal untuk perubahan dokumen Phase 5:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
```

Jalankan `npm run check` hanya bila ada perubahan Svelte. Jalankan `flutter analyze` dan `flutter test` bila ada perubahan runtime Flutter atau saat membuat APK release; Flutter SDK bisa tidak tersedia di host dokumentasi, sehingga hasilnya boleh dicatat sebagai blocker validasi mobile sampai dijalankan di environment yang memiliki SDK.
