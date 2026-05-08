# CBT Proposal Integration Phase 0

Status: Phase 0 alignment document, 2026-05-08.

Dokumen ini menjadi rujukan resmi untuk menyerap proposal Sistem CBT MTsN 2 Kolaka Utara ke monorepo tanpa mengubah runtime architecture yang sudah aktif. Proposal DOCX diperlakukan sebagai sumber kebutuhan produk dan operasi sekolah, bukan sebagai instruksi untuk mengganti stack monorepo.

Phase 0 tidak mengubah runtime code, database schema, deployment topology, PM2 process, atau kontrak production. Output Phase 0 adalah penyelarasan dokumen, guard test ringan, dan acceptance criteria sebelum fase implementasi berikutnya.

## Keputusan Architecture

CBT di monorepo tetap mengikuti boundary berikut:

| Area | Runtime Resmi | Tanggung Jawab |
| --- | --- | --- |
| Web Admin | `apps/web-admin` - SvelteKit 2, Svelte 5, Tailwind v4, shadcn-svelte | Bank Soal, Asesmen, Proctor/Pengawasan, Hasil, session cookie, BFF/proxy |
| Backend | `services/core-api` - Go, Chi, sqlc, pgx, PostgreSQL | Owner state, API, timer, audit, scoring, token ujian, media authorization, scheduler/queue |
| Student App | `apps/mobile` - Flutter Android | Portal ujian siswa utama, anti-cheat utama, BYOD telemetry, local answer safety, restore UX |
| Worker | `services/pusaka-worker` - TypeScript Playwright | PUSAKA attendance automation only, tidak ikut runtime CBT |

PocketBase, SQLite, and Alpine are not runtime architecture for this monorepo.

Artinya:

- PocketBase tidak menjadi API server, auth server, database owner, admin panel, atau storage runtime CBT.
- SQLite tidak menjadi database runtime CBT, web-admin, worker, atau mobile. File SQLite lama hanya boleh menjadi artefak import satu arah.
- Alpine dari proposal tidak menjadi deployment runtime atau base image production. Deployment server tetap native PM2 sesuai `deploy/DEPLOY.md`; Flutter dibuild sebagai APK.
- Tidak ada service baru untuk CBT pada Phase 0. `services/core-api` tetap owner PostgreSQL dan exam API.
- `apps/web-admin` tetap BFF/UI dan tidak boleh membaca PostgreSQL langsung.
- Flutter berbicara langsung ke exam API backend, bukan melalui SvelteKit, untuk live exam runtime.

## Alur Sistem Resmi

```text
Admin/Guru/Staf
  -> apps/web-admin
  -> BFF /api/bank-soal/* dan /api/asesmen/*
  -> services/core-api
  -> PostgreSQL

Siswa
  -> apps/mobile Flutter
  -> services/core-api /api/exam/*
  -> PostgreSQL

Pengawas/Proctor
  -> apps/web-admin /asesmen/pengawasan atau /asesmen/pelaksanaan
  -> services/core-api
  -> audit, participant status, anti-cheat events
```

Bank Soal berdiri di route `/bank-soal/*`. Asesmen memakai hub `/asesmen`, `/asesmen/persiapan`, `/asesmen/pelaksanaan`, dan `/asesmen/hasil`. Legacy `/cbt*` dan `/api/cbt*` hanya compatibility path sesuai aturan repo.

## Roadmap Fase 0-5

### Fase 0 - Alignment and Guard

Tujuan: menyerap proposal ke dokumen resmi tanpa runtime change.

Scope:

- Tetapkan boundary Web Admin, Core API, Flutter, dan PUSAKA worker.
- Tegaskan proposal tidak membawa PocketBase, SQLite, atau Alpine ke runtime monorepo.
- Dokumentasikan roadmap, kontrak API awal, anti-cheat Flutter, RBAC/audit, validasi, dan acceptance criteria.
- Tambahkan docs guard ringan agar dokumen tidak bergeser dari keputusan architecture.

Acceptance:

- Dokumen Phase 0 ada di `docs/`.
- Dokumen anti-cheat Flutter ada bila diperlukan.
- Guard test membuktikan dokumen menyebut Flutter sebagai portal ujian siswa dan melarang PocketBase/SQLite/Alpine sebagai runtime.
- `git diff --check` bersih.
- Test docs guard relevan berjalan.

### Fase 1 - Contract Stabilization

Tujuan: menstabilkan kontrak exam API yang dipakai Flutter sebelum fitur proposal diperluas.

Web Admin:

- Pastikan menu dan page release aplikasi siswa menaut ke `docs/exam-api.md`.
- Tidak menambah local storage state untuk domain ujian selain draft UI yang eksplisit.

Core API:

- Kunci response envelope `data` untuk login, status, answer, submit, heartbeat, dan event.
- Jaga HTTP semantics mobile: `404` token tidak ditemukan, `403` sesi belum aktif/waktu tertutup, `409` device mismatch/sudah submitted, `401` konteks peserta hilang.
- Tambahkan atau pertahankan handler tests untuk payload login/status/answer/submit/event.

Flutter:

- Pastikan parser model tetap menerima rich content, media/audio URL, `is_submitted`, dan schedule metadata.
- Pastikan mapping guidance untuk `403/409` tetap teruji.

Acceptance:

- `docs/exam-api.md` menjadi kontrak release mobile.
- Backend dan Flutter test suite mengunci field kritis dan error semantics.
- Tidak ada key jawaban, token peserta lain, atau URL media berisi token di payload siswa.

### Fase 2 - Web Admin Workflow Alignment

Tujuan: memetakan fitur proposal ke UI operator/guru yang sudah ada.

Web Admin:

- Bank Soal: daftar, tambah/editor, impor, verifikasi, rich content, readiness, KaTeX, RTL, dan workflow review tetap di `/bank-soal/*`.
- Asesmen Persiapan: paket, kegiatan/event, sesi, ruang, peserta, pengawas, token, kartu ujian.
- Asesmen Pelaksanaan: dashboard pengawas, status peserta, warning anti-cheat, seat/room view, day-of actions.
- Asesmen Hasil: hasil peserta, scoring, export, essay/non-test follow-up.
- Semua fetch baru memakai `/api/bank-soal/*` dan `/api/asesmen/*`.

Core API:

- Menjadi owner semua state workflow, query, mutation, audit, dan validation.
- Handler tetap thin; service memegang business rules.

Flutter:

- Tidak ikut authoring Bank Soal atau manajemen sesi.
- Hanya konsumsi payload ujian yang sudah published dan active.

Acceptance:

- Tidak ada page baru yang memakai direct database access.
- Tidak ada active UI baru yang fetch `/api/cbt/*` kecuali compatibility test/redirect.
- RBAC web-admin dan backend cocok untuk admin, guru, staf, dan pengawas sesuai role.

### Fase 3 - Backend Runtime Hardening

Tujuan: memperkuat runtime ujian yang menjadi source of truth.

Core API:

- Timer ujian, status peserta, duplicate submit, scoring, audit, dan proctor event berada di service backend.
- PostgreSQL tetap owner state via migrations dan sqlc.
- Answer write harus idempotent pada konflik yang dapat diprediksi dan tetap eksplisit untuk duplicate submit.
- Scoring tidak boleh menandai peserta belum submit sebagai submitted.
- Rate limit exam login dan auth/public endpoints memakai trusted proxy policy.
- Upload/media ujian memakai allowlist MIME/extension, `nosniff`, dan disposition aman.

Web Admin:

- Menampilkan status dan warning dari backend, bukan menghitung sendiri state kritis ujian.

Flutter:

- Mengirim heartbeat, answer, event, dan submit sesuai kontrak.
- Menahan submit manual ketika pending answer sync atau koneksi terdegradasi melewati threshold yang disepakati.

Acceptance:

- Handler tests mengunci response envelope dan error mapping.
- Service tests menutup timer, duplicate submit, scoring, dan role visibility.
- Audit event mencatat mutasi penting dan event exam dengan pesan client-safe.

### Fase 4 - Flutter BYOD Anti-Cheat and Resilience

Tujuan: menjadikan Flutter lokasi utama anti-cheat realistis untuk BYOD.

Flutter:

- `FLAG_SECURE` aktif untuk mengurangi screenshot/recent preview.
- Lifecycle observer mengirim app background/resume warning.
- Heartbeat berkala dan stale-contact detection memunculkan status `Tersambung`, `Lokal`, `Waspada`, dan `Menurun`.
- Local answer safety memakai storage aman untuk token, fingerprint, answers, dan pending answers.
- Resume gate memaksa status refresh sebelum lanjut saat app kembali dari background atau restore.
- Warning panel memberi instruksi siswa/pengawas sesuai urgency.

Core API:

- Menyimpan event anti-cheat dan connection warnings sebagai audit/proctor evidence.
- Menyediakan endpoint status yang cukup untuk Flutter memulihkan session tanpa mengambil answer key.

Web Admin:

- Proctor dashboard menampilkan event dan connection risk dari backend secara ringkas.

Acceptance:

- Flutter unit/widget tests menutup guidance, status chip, restore, media/audio, dan event payload.
- BYOD limitation tertulis jelas di docs dan UI guidance.
- Tidak ada klaim kiosk/device-owner untuk perangkat siswa pribadi.

### Fase 5 - Rehearsal, Rollout, and Post-Exam Review

Tujuan: membawa sistem ke operasi sekolah dengan rehearsal yang dapat diaudit.

Web Admin:

- Operator menjalankan checklist Bank Soal, Asesmen Persiapan, Pelaksanaan, dan Hasil.
- Pengawas memakai dashboard proctoring saat simulasi.

Core API:

- Health check, audit, logs, backup, dan migration state dicek sebelum ujian besar.
- Data hasil dan event anti-cheat dapat diekspor/direview setelah ujian.

Flutter:

- APK release diverifikasi dengan `apps/mobile/RELEASE_CHECKLIST.md`.
- Uji beberapa perangkat nyata mencakup login token, heartbeat, answer save, restore, network disturbance, warning, dan submit.

Acceptance:

- `docs/cbt-smoke-checklist.md` selesai dengan catatan hasil rehearsal.
- Minimal satu rehearsal end-to-end dari Bank Soal sampai Hasil lulus.
- Rollback plan jelas: hentikan sesi baru, pertahankan data backend, distribusikan APK sebelumnya bila client release bermasalah.

## Pembagian Tanggung Jawab

### Web Admin

`apps/web-admin` bertanggung jawab atas pengalaman operator, guru, staf, dan pengawas:

- Bank Soal authoring, import, verification, readiness signal.
- Asesmen preparation, session scheduling, participant setup, card/token print.
- Proctor monitoring, participant status, warning review.
- Hasil, scoring review, export, dan laporan.
- BFF proxy dengan JWT forwarding dari session cookie.

Web Admin tidak bertanggung jawab atas:

- Timer ujian yang menentukan sah/tidaknya jawaban.
- Scoring final.
- Penyimpanan jawaban siswa.
- Direct PostgreSQL access.
- Runtime anti-cheat siswa.

### Core API

`services/core-api` bertanggung jawab atas state dan aturan domain:

- PostgreSQL ownership, migrations, sqlc queries.
- Exam token generation and validation.
- Exam login, status, heartbeat, answer, event, submit.
- Timer, duplicate-submit semantics, scoring, audit logs.
- RBAC and role visibility.
- Media/asset authorization.
- Client-safe errors and server-side logs.

Core API tidak boleh menyerahkan state kritis ujian ke SvelteKit, Flutter, PocketBase, SQLite, atau worker.

### Flutter

`apps/mobile` bertanggung jawab sebagai portal ujian siswa utama:

- Token login.
- Rendering soal, opsi, rich text, image, audio.
- Local answer safety and pending sync.
- Heartbeat and event telemetry.
- BYOD anti-cheat UX: secure screen, lifecycle warnings, restore gates, degraded connection guidance.
- Explicit final submit flow.

Flutter tidak menjadi source of truth untuk waktu, scoring, answer key, participant status, atau role admin/guru.

## Kontrak API Awal Mobile Exam

Kontrak lengkap tetap berada di `docs/exam-api.md`. Ringkasan awal untuk integrasi proposal:

| Flow | Method and Path | Auth | Fungsi |
| --- | --- | --- | --- |
| Login | `POST /api/exam/login` | body token + `device_fingerprint` | Bind token, ambil peserta, sesi, room, soal, progres, sisa waktu |
| Status | `GET /api/exam/status` | `X-Exam-Token`, `X-Device-Fingerprint` | Ambil progres, sisa waktu, `is_submitted` |
| Heartbeat | `POST /api/exam/heartbeat` | `X-Exam-Token`, `X-Device-Fingerprint` | Tandai client masih aktif |
| Event | `POST /api/exam/event` | `X-Exam-Token`, `X-Device-Fingerprint` | Catat anti-cheat/warning/telemetry |
| Answer | `POST /api/exam/answer` | `X-Exam-Token`, `X-Device-Fingerprint` | Simpan jawaban satu soal |
| Submit | `POST /api/exam/submit` | `X-Exam-Token`, `X-Device-Fingerprint` | Finalisasi ujian |

Header wajib setelah login:

```http
X-Exam-Token: <token ujian siswa>
X-Device-Fingerprint: <fingerprint telemetry/resume hint>
```

`X-Device-Fingerprint` bukan identitas perangkat kuat. Ia hanya membantu resume, deteksi mismatch, dan telemetry BYOD.

Payload login minimal yang harus stabil untuk Flutter:

```json
{
  "data": {
    "participant_id": "uuid",
    "student": {
      "nis": "12345",
      "nama": "Ahmad Siswa"
    },
    "session": {
      "id": "uuid",
      "title": "Ujian Matematika Kelas VII",
      "scheduled_start": "2026-05-08T00:00:00Z",
      "scheduled_end": "2026-05-08T02:00:00Z",
      "duration_minutes": 90
    },
    "room": {
      "room_name": "Ruang 01"
    },
    "questions": [
      {
        "id": "uuid-q1",
        "question_type": "multiple_choice",
        "question_text": "1 + 1 = ...",
        "stem_html": "<p>1 + 1 = ...</p>",
        "stimulus_html": "",
        "stem_media_url": "",
        "stimulus_media_url": "",
        "stem_audio_url": "",
        "stimulus_audio_url": "",
        "options": [
          { "label": "A", "text": "1" },
          { "label": "B", "text": "2" }
        ]
      }
    ],
    "answered_count": 0,
    "total_questions": 30,
    "time_remaining_seconds": 5400,
    "is_submitted": false
  }
}
```

Response sukses tetap memakai envelope `data`. Response error harus client-safe dan mempertahankan semantics:

- `404` untuk token tidak ditemukan.
- `403` untuk sesi belum aktif, sesi sudah tertutup, atau waktu submit/jawab tidak sah.
- `409` untuk device mismatch atau ujian sudah submitted.
- `401` untuk request tanpa konteks peserta/token yang sah.

## Event Anti-Cheat Flutter

Phase 0 menetapkan event taxonomy awal. Implementasi backend dapat tetap menerima event generik selama transisi, tetapi Flutter dan docs harus bergerak ke event yang lebih eksplisit.

Envelope event:

```json
{
  "event_type": "app_backgrounded",
  "data": {
    "occurred_at": "2026-05-08T01:15:00Z",
    "sequence": 12,
    "reason": "lifecycle_paused",
    "connection_state": "waspada",
    "stale_seconds": 95,
    "pending_answer_count": 2
  }
}
```

Event awal yang disarankan:

| Event | Sumber Flutter | Tujuan Backend/Web Admin |
| --- | --- | --- |
| `app_backgrounded` | App masuk background/minimized | Warning proctor bahwa siswa keluar dari permukaan ujian |
| `app_resumed` | App kembali foreground | Korelasi durasi keluar app dan resume gate |
| `resume_gate_shown` | Flutter meminta status refresh sebelum lanjut | Bukti UX pengaman berjalan |
| `heartbeat_failed` | Heartbeat gagal | Indikasi gangguan koneksi |
| `heartbeat_recovered` | Heartbeat kembali sukses | Tutup episode gangguan |
| `stale_connection_attention` | Kontak server stale melewati threshold awal | Panel perhatian pengawas |
| `stale_connection_escalated` | Stale melewati threshold urgent | Panel intervensi segera |
| `answer_saved_local` | Jawaban tersimpan lokal karena sync gagal | Data pending lokal perlu dipantau |
| `answer_sync_recovered` | Pending answer berhasil disinkronkan | Risiko lokal menurun |
| `manual_submit_blocked` | Submit ditahan karena sync/status belum aman | Mencegah final submit tidak terpercaya |
| `screenshot_attempt` | Best-effort platform signal bila tersedia | Sinyal deterrence, bukan bukti lengkap |
| `restore_attempted` | Siswa mencoba restore session | Audit resume BYOD |
| `restore_failed` | Cache/session tidak dapat dipulihkan | Bantuan operator dan analisis pasca kejadian |

Event data tidak boleh menyertakan password, token mentah di nested payload, answer key, atau informasi perangkat yang lebih sensitif dari fingerprint telemetry yang sudah disepakati.

## BYOD vs Kiosk/Device-Owner

Sistem ini didesain untuk siswa memakai Android pribadi (BYOD). Anti-cheat Flutter adalah deterrence, telemetry, dan operational guidance, bukan jaminan kiosk penuh.

Yang realistis pada BYOD:

- `FLAG_SECURE` untuk mengurangi screenshot dan recent-app preview.
- Lifecycle detection untuk background/resume/app switch.
- Heartbeat dan stale-contact detection.
- Local answer safety saat koneksi terganggu.
- Resume gate dan guidance panel.
- Event telemetry untuk pengawas dan audit.
- Visible sync state agar siswa/pengawas paham risiko koneksi.

Yang tidak boleh diklaim pada BYOD:

- Mengunci seluruh perangkat seperti kiosk sekolah.
- Mencegah penggunaan perangkat kedua.
- Menjamin tidak ada kamera eksternal.
- Menjamin screen recording pada semua vendor/OS.
- Menjadikan fingerprint sebagai bukti identitas perangkat kuat.
- Mencegah siswa mematikan jaringan atau force close aplikasi.

Kiosk/device-owner hanya bisa dipertimbangkan bila sekolah menyediakan perangkat terkelola. Itu fase terpisah, bukan baseline `apps/mobile` saat ini.

## Keamanan, RBAC, dan Audit

Prinsip keamanan:

- Admin/guru/staf memakai JWT session via web-admin; siswa ujian memakai exam token.
- BFF authenticated routes harus forward JWT user asli ke Core API.
- `X-Internal-Key` bukan fallback diam-diam untuk traffic authenticated web-admin.
- Backend adalah sumber kebenaran RBAC dan permission enforcement.
- Exam token harus kuat, random, dan tidak muncul di URL media.
- Answer key, token peserta lain, dan data admin tidak boleh masuk payload Flutter.
- Semua mutasi penting dan exam warning/event masuk audit/proctor evidence backend.
- Error 500 ke client harus generik; detail internal hanya di server log.
- Rate limiting wajib untuk login, refresh, public registration, dan exam token login.
- Forwarded client IP hanya dipercaya dari proxy yang masuk `TRUSTED_PROXY_CIDRS`.

RBAC awal:

| Area | Admin | Guru | Staf | Siswa |
| --- | --- | --- | --- | --- |
| Bank Soal | Full sesuai permission | Author/review sesuai scope | Tidak default | Tidak |
| Asesmen Persiapan | Full | Paket/sesi sesuai scope | Terbatas bila ditugaskan | Tidak |
| Pelaksanaan/Proctor | Full | Pengawasan sesuai tugas | Day-of support sesuai permission | Tidak |
| Hasil | Full | Hasil kelas/mapel sesuai scope | Tidak default | Tidak via web-admin |
| Exam Runtime | Tidak memakai exam token siswa | Tidak memakai exam token siswa | Tidak memakai exam token siswa | Flutter token-only |

## Validasi

Minimal validasi Phase 0:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
```

Validasi fase berikutnya menyesuaikan area perubahan:

- Web Admin: `npm run check`, `npm run test:unit`.
- Core API: `go test ./...` atau target backend terkait.
- Flutter: `flutter analyze`, `flutter test`.
- Operasional CBT: `docs/cbt-smoke-checklist.md` dan `apps/mobile/RELEASE_CHECKLIST.md`.

Phase 0 tidak menjalankan deploy, restart PM2, migration, atau APK release.

## Acceptance Criteria Phase 0

Phase 0 dianggap selesai bila:

- Dokumen ini menjelaskan alignment proposal ke monorepo.
- Dokumen menyebut Web Admin SvelteKit untuk Bank Soal/Asesmen/Proctor/Hasil.
- Dokumen menyebut Core API Go+PostgreSQL sebagai owner state/API/timer/audit/scoring.
- Dokumen menyebut Flutter `apps/mobile` sebagai portal ujian siswa utama dan lokasi anti-cheat utama.
- Dokumen menolak PocketBase, SQLite, dan Alpine sebagai runtime.
- Dokumen memuat roadmap fase 0-5.
- Dokumen memuat pembagian tanggung jawab web/backend/flutter.
- Dokumen memuat kontrak API awal mobile exam.
- Dokumen memuat event anti-cheat Flutter.
- Dokumen memuat batasan BYOD vs kiosk/device-owner.
- Dokumen memuat keamanan, RBAC, audit, validasi, dan acceptance criteria.
- Guard test docs berjalan tanpa menyentuh runtime code.
