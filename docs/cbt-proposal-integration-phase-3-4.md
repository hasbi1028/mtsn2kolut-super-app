# CBT Proposal Integration Phase 3-4

Status: runtime hardening and BYOD resilience guard, 2026-05-08.

Dokumen ini mencatat implementasi Fase 3 dan Fase 4 integrasi proposal Sistem CBT MTsN 2 Kolaka Utara. Scope fase ini tetap memperkuat runtime yang sudah ada, bukan mengganti architecture atau membuat service CBT baru.

## Boundary Tetap

| Runtime | Peran |
| --- | --- |
| `services/core-api` | Source of truth untuk token, timer, autosave jawaban, submit final, audit/event, scoring, status peserta, dan media authorization. |
| `apps/mobile` | Portal ujian siswa utama dan anti-cheat BYOD utama: secure screen, lifecycle telemetry, restore gate, pending-answer safety, connection guidance, dan final submit UX. |
| `apps/web-admin` | Workflow operator/pengawas; menampilkan status dan warning dari backend, bukan runtime ujian siswa. |

Non-goal yang tetap berlaku:

- Tidak menjalankan migration live, deploy, PM2 restart, atau `make db-migrate`.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru untuk runtime siswa.
- Tidak mengubah runtime architecture ke PocketBase, SQLite, Alpine, Docker runtime, atau service CBT baru.
- Tidak mengklaim BYOD sebagai kiosk penuh. Device-owner/kiosk hanya opsi masa depan bila sekolah memakai perangkat terkelola.

## Phase 3 - Backend Runtime Hardening

Perubahan Core API:

- Body JSON exam runtime sekarang dibaca dengan helper yang mengunci satu JSON object saja; trailing JSON ditolak sebagai `400 invalid json`.
- Body limit mobile-facing:
  - `POST /api/exam/login`: 4 KiB.
  - `POST /api/exam/event`: 16 KiB.
  - `POST /api/exam/answer`: 64 KiB serialized JSON.
  - `POST /api/exam/heartbeat` dan `POST /api/exam/submit`: empty body atau `{}` saja, maksimum 1 KiB.
- Oversized exam request dikembalikan sebagai `413 request body too large`, tetap dalam envelope error client-safe.
- Login token dan `event_type` di-trim sebelum validasi dan service call.
- `ExamToken` dan `ExamTokenOrJWT` sekarang membedakan:
  - `401 unauthorized` untuk token/fingerprint yang belum membentuk participant context sah, termasuk missing `X-Device-Fingerprint` atau participant belum device-bound.
  - `409 token already bound to another device` hanya untuk mismatch terhadap fingerprint yang sudah terikat.
- Response sukses tetap wrapped: `data.status = ok|recorded|submitted`.

Guard yang relevan:

- `services/core-api/internal/handler/exam_test.go` mengunci body cap, trailing JSON rejection, trimmed fields, participant context, success envelope, dan error mapping.
- `services/core-api/internal/middleware/exam_token_test.go` dan `services/core-api/internal/middleware/more_test.go` mengunci `401` vs `409` device-context semantics.
- `services/core-api/cmd/api/routes_contract_test.go` tetap mengunci runtime siswa di `/api/exam/*` dan login rate limit dengan trusted proxy policy.

## Phase 4 - Flutter BYOD Anti-Cheat and Resilience

Perubahan Flutter:

- `apps/mobile/lib/src/exam_events.dart` menambahkan helper taxonomy event BYOD yang testable untuk `app_switch`, `warning`, `resume_exam`, `repeat_resume_attempt`, `answer_saved_local_only`, `submit_blocked_pending_sync`, `auto_submit_blocked_pending_sync`, `submit_blocked_degraded_mode`, `degraded_mode_entered`, `stale_connection_attention`, `stale_connection_escalated`, `back_button_attempt`, dan `manual_submit`.
- Helper event menyaring field sensitif seperti token, password, dan answer key dari nested telemetry payload sebelum dikirim.
- Exam shell memakai helper taxonomy tersebut agar payload event tidak tersebar sebagai raw string yang mudah drift.
- Resume gate tidak lagi dibuka bila status refresh gagal; siswa tetap berada di mode aman sampai status berhasil dicek ulang.
- Restore dari snapshot masuk ke exam shell dengan `initialResumeCheckRequired: true`, sehingga sesi lama wajib melewati status check sebelum lanjut.
- Pending-answer flush yang mendapat `409 exam already submitted` diperlakukan sebagai terminal server state: shell sync status, masuk layar selesai bila server sudah final, dan snapshot lokal dibersihkan. `409 token already bound to another device` tetap menjaga jawaban pending lokal agar tidak hilang.
- Batas jawaban uraian Flutter dikunci konservatif pada 15.000 karakter, lalu `ExamApiClient` menghitung exact serialized UTF-8 JSON body untuk `/api/exam/answer` sebelum request. Jawaban yang melewati 64 KiB ditahan client-side dengan `413` agar tidak menjadi pending permanen.
- Guidance helper menambahkan mapping `statusFailureMessage/statusFailureNotice` serta `401` untuk answer/submit agar missing participant context berbeda dari device mismatch.

Guard yang relevan:

- `apps/mobile/test/exam_events_test.dart` mengunci payload taxonomy dan sanitasi telemetry.
- `apps/mobile/test/exam_api_test.dart` mengunci `sendExamEvent` tetap mengirim envelope event yang sama dengan backend.
- `apps/mobile/test/widget_test.dart` mengunci resume gate tetap aktif saat status refresh gagal dan pending flush `409` menjadi terminal.
- `apps/mobile/test/exam_error_messages_test.dart` mengunci guidance status/answer/submit untuk transport, `401`, `403`, dan `409`.

## Validasi Minimal Fase 3-4

```bash
git diff --check
cd services/core-api && GOCACHE=/tmp/go-build go test ./internal/handler ./internal/middleware ./cmd/api -run 'TestExam|Test.*Routes'
cd apps/mobile && flutter test test/exam_events_test.dart test/exam_api_test.dart test/exam_error_messages_test.dart test/widget_test.dart
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
```

`flutter test` membutuhkan Flutter SDK tersedia di environment lokal/CI. Bila SDK tidak tersedia di host, catat sebagai blocker validasi mobile dan jalankan sebelum merge.
