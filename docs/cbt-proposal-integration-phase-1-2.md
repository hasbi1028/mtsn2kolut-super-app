# CBT Proposal Integration Phase 1-2

Status: contract and workflow guard, 2026-05-08.

Dokumen ini mencatat implementasi Fase 1 dan Fase 2 integrasi proposal Sistem CBT MTsN 2 Kolaka Utara ke monorepo. Fokus fase ini adalah stabilisasi kontrak dan alignment workflow yang sudah ada, bukan rewrite stack, schema, atau deployment.

## Boundary Tetap

| Runtime | Peran |
| --- | --- |
| `services/core-api` | Source of truth token ujian, timer, autosave jawaban, submit final, audit/event, scoring, media authorization, dan error semantics client-safe. |
| `apps/mobile` | Portal ujian siswa utama dan lokasi anti-cheat BYOD utama: token login, render soal, heartbeat, telemetry, local answer safety, restore gate, dan final submit. |
| `apps/web-admin` | Workflow operator/guru/staf/pengawas: Bank Soal, Asesmen, paket, kegiatan, sesi, pengawasan, hasil, dan panduan aplikasi siswa. |

Non-goal fase ini:

- Tidak menjalankan migration live, deploy, PM2 restart, atau APK release.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Tidak mengubah runtime ke PocketBase, SQLite, Alpine, Docker runtime, atau service CBT baru.
- Tidak menambah persistence baru di web-admin atau mobile untuk state domain yang sudah dimiliki Core API.

## Fase 1 - Contract Stabilization API Ujian Flutter

Kontrak mobile-critical tetap berada di `docs/exam-api.md` dan dijaga oleh handler tests Core API serta test Flutter yang sudah ada.

Endpoint yang dikunci:

| Endpoint | Owner | Kontrak stabil |
| --- | --- | --- |
| `POST /api/exam/login` | Core API | Body `token` + `device_fingerprint`, forwarded client IP, response envelope `data`, payload siswa/sesi/room/soal/progres, media/audio URL tanpa token. |
| `GET /api/exam/status` | Core API | Token participant context, response envelope `data`, `answered_count`, `total_questions`, `time_remaining_seconds`, `is_submitted`, optional `submitted_at`. |
| `POST /api/exam/heartbeat` | Core API | Token participant context, response `data.status = "ok"`. |
| `POST /api/exam/event` | Core API | Token participant context, `event_type`, `data`, response `data.status = "recorded"`. |
| `POST /api/exam/answer` | Core API | Token participant context, `question_id`, `answer`, response `data.status = "recorded"`, explicit conflict/window/scope errors. |
| `POST /api/exam/submit` | Core API | Token participant context, response `data.status = "submitted"`, duplicate submit as `409`. |

HTTP semantics yang harus stabil untuk mobile:

- `400`: JSON tidak valid atau field wajib malformed/missing.
- `401`: request token-scoped tanpa participant context yang sah.
- `403`: sesi belum aktif, sesi belum mulai, window sudah tertutup, atau action runtime tidak sah karena waktu.
- `404`: token ujian tidak ditemukan.
- `409`: token sudah terikat perangkat lain atau ujian sudah submitted.
- `500`: error internal generic, detail hanya di server log.

Guard backend yang relevan:

- `services/core-api/internal/handler/exam_test.go` mengunci envelope, payload field mobile, forwarded token/device/IP login, participant-context forwarding, malformed request, dan error mapping.
- `services/core-api/cmd/api/routes_contract_test.go` mengunci route `/api/exam/*` agar login tetap rate-limited dan endpoint runtime tetap berada di `examTokenMW`.

## Taxonomy Event BYOD Flutter

Core API menerima event client sebagai evidence proctor/audit. Flutter saat ini mengirim `event_type` berikut:

| `event_type` | `data.reason` atau data utama | Arti operasional |
| --- | --- | --- |
| `app_switch` | `state`, `device_fingerprint` | App meninggalkan/berubah lifecycle dari permukaan ujian. |
| `warning` | `resume_exam` | Resume gate berjalan setelah app kembali foreground. |
| `warning` | `repeat_resume_attempt` | Siswa kembali dari background lebih dari sekali. |
| `warning` | `answer_saved_local_only` | Jawaban tersimpan lokal karena sync gagal. |
| `warning` | `submit_blocked_pending_sync` atau `auto_submit_blocked_pending_sync` | Submit ditahan sampai jawaban lokal aman tersinkron. |
| `warning` | `submit_blocked_degraded_mode` | Submit manual ditahan karena koneksi menurun. |
| `warning` | `degraded_mode_entered` | Perangkat masuk mode koneksi menurun. |
| `warning` | `stale_connection_attention` | Pengawas perlu memperhatikan koneksi stale. |
| `warning` | `stale_connection_escalated` | Intervensi pengawas/proktor sudah urgent. |
| `warning` | `back_button_attempt` | Siswa mencoba tombol kembali. |
| `warning` | `manual_submit` | Submit manual berhasil dikirim dari Flutter. |
| `screenshot_attempt` | platform signal bila tersedia | Deterrence signal; bukan bukti lengkap pada BYOD. |

Event data tidak boleh menyertakan password, token ujian mentah di nested payload, answer key, credential admin, atau data perangkat yang lebih sensitif dari fingerprint telemetry yang sudah disepakati.

## Fase 2 - Web Admin Workflow Alignment

Workflow operator tetap memakai route canonical berikut:

```text
Bank Soal
  /bank-soal
  /bank-soal/daftar
  /bank-soal/tambah
  /bank-soal/impor
  /bank-soal/verifikasi
  /bank-soal/analisis-butir
  /bank-soal/mapel-kd
  /bank-soal/pengaturan

Asesmen
  /asesmen
  /asesmen/persiapan
  /asesmen/paket
  /asesmen/kegiatan
  /asesmen/sesi
  /asesmen/pengawasan
  /asesmen/pelaksanaan
  /asesmen/hasil
  /asesmen/aplikasi-siswa
  /asesmen/non-tes
```

Alur operasional yang dikunci:

1. Bank Soal: daftar soal, tambah/editor, impor, verifikasi/review, readiness, rich content, KaTeX/RTL, workflow publish.
2. Asesmen Persiapan: paket, kegiatan/event, sesi, peserta, ruang, seat, pengawas/proktor, token, kartu ujian.
3. Pelaksanaan/Pengawasan: dashboard sesi/ruang, heartbeat, status submit, warning BYOD, reset akses/token sesuai prosedur.
4. Hasil: scoring, essay/non-test follow-up, item analysis, export, rekap event/sesi.
5. Aplikasi Siswa/BYOD readiness: operator dan pengawas memakai panduan status Flutter, bukan menganggap BYOD sebagai kiosk penuh.

BFF namespace aktif:

- Bank Soal client baru memakai `/api/bank-soal/*`.
- Asesmen client baru memakai `/api/asesmen/*`.
- `/api/cbt/*` hanya compatibility/deprecated route yang sudah ada untuk legacy clients dan staged transition; jangan membuat route tree publik baru di namespace itu.

Guard Web Admin yang relevan:

- `apps/web-admin/src/lib/server/cbt-backend-paths.test.ts` mengunci dispatch Bank Soal ke `/api/bank-soal/*` dan Asesmen ke `/api/asesmen/*`.
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts` mengunci taxonomy navigasi Bank Soal dan Asesmen.
- `apps/web-admin/src/lib/server/route-access.test.ts` mengunci permission/role guard untuk Bank Soal, Asesmen, pengawasan, dan hasil.
- `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts` mengunci dokumen ownership, API Flutter, route canonical, non-goal runtime, dan BYOD limitation.

## Validasi Minimal Fase 1-2

```bash
git diff --check
cd services/core-api && GOCACHE=/tmp/go-build go test ./internal/handler ./cmd/api -run 'TestExam|Test.*Routes'
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts src/lib/server/cbt-backend-paths.test.ts
```

Jalankan `npm run check` hanya bila ada perubahan Svelte. Jalankan `flutter test` bila ada perubahan `apps/mobile`.
