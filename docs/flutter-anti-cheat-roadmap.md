# Flutter Anti-Cheat Roadmap

Status: companion roadmap untuk `docs/cbt-proposal-integration-phase-0.md`, 2026-05-08.

Dokumen ini merangkum arah anti-cheat untuk portal ujian siswa di `apps/mobile`. Fokusnya adalah Android BYOD: siswa memakai perangkat pribadi, sehingga kontrol yang realistis adalah deterrence, telemetry, answer safety, resume gate, dan guidance pengawas. Kiosk/device-owner bukan baseline kecuali sekolah kelak menyediakan perangkat terkelola.

## Scope

Anti-cheat ownership lives in `apps/mobile`.

Backend tetap menjadi source of truth untuk token, timer, answer state, audit, scoring, dan proctor evidence. Web Admin menampilkan sinyal dari backend untuk pengawas. Flutter mengumpulkan sinyal dari perangkat siswa dan menjaga pengalaman ujian tetap aman saat koneksi atau lifecycle berubah.

## Prinsip

- Jangan klaim anti-cheat BYOD sebagai kiosk penuh.
- Jangan membuat fingerprint perangkat sebagai bukti identitas kuat.
- Jangan menyimpan token, fingerprint, answers, atau pending answers di plaintext preference yang sama dengan metadata restore umum.
- Jangan kirim answer key, password, atau token mentah di event telemetry.
- Jangan blokir final submit hanya karena sinyal lemah; blokir hanya ketika status/pending sync membuat submit tidak terpercaya.
- Semua keputusan final tentang waktu, submit, scoring, dan audit tetap di Core API.

## Roadmap

### Phase 0 - Documentation Guard

Output:

- CBT proposal integration document.
- Anti-cheat roadmap.
- Guard test yang mengunci Flutter sebagai portal ujian siswa dan lokasi anti-cheat utama.

Acceptance:

- Docs menyebut BYOD limitation.
- Docs menolak PocketBase/SQLite/Alpine sebagai runtime.
- Tidak ada runtime code, migration, deploy, atau PM2 restart.

### Phase 1 - Event Taxonomy and Contract Tests

Flutter:

- Tetapkan helper payload event yang testable.
- Pertahankan compatibility untuk event lama seperti `app_switch` dan `warning` bila backend masih memerlukannya.
- Kirim event lifecycle, heartbeat failure/recovery, stale connection, restore, dan manual-submit-blocked dengan schema data konsisten.

Core API:

- Validasi `event_type` tidak kosong.
- Simpan event dengan participant/session context.
- Response event tetap `data.status = recorded`.

Acceptance:

- Flutter unit tests menutup event builder.
- Backend handler tests menutup request forwarding dan validation.
- Event taxonomy terdokumentasi di `docs/cbt-proposal-integration-phase-0.md`.

### Phase 2 - BYOD Resilience and Guidance

Flutter:

- Status chip membedakan `Tersambung`, `Lokal`, `Waspada`, dan `Menurun`.
- Repeated sync failure memunculkan warning panel.
- Stale connection melewati threshold awal memunculkan pengawas-attention panel.
- Stale connection melewati threshold urgent memunculkan intervensi segera.
- Pending answer sync mempengaruhi submit readiness.

Core API:

- Status endpoint mempertahankan `is_submitted`, progress, dan `time_remaining_seconds`.
- Event stale/recovery dapat dikorelasikan di audit/proctor view.

Acceptance:

- Widget tests menutup status chip dan guidance panel.
- Manual submit diblokir saat pending answer/status belum aman.
- Restore flow menjelaskan status terakhir kepada siswa/pengawas.

### Phase 3 - Proctor Visibility

Web Admin:

- Dashboard pengawas menampilkan event anti-cheat terbaru per peserta.
- Sinyal ditampilkan sebagai ringkasan operasional, bukan noise mentah.
- Pengawas dapat membedakan warning koneksi, app switch, restore, dan submit risk.

Core API:

- Query proctoring efisien dan role-scoped.
- Audit event tetap client-safe.

Acceptance:

- Guru/pengawas hanya melihat sesi yang menjadi scope tugasnya.
- Admin melihat event lintas sesi sesuai permission.
- Export/laporan pasca ujian dapat memuat ringkasan event bila diperlukan.

### Phase 4 - App Hardening

Flutter:

- `FLAG_SECURE` tetap aktif di shell ujian.
- Back navigation dan route escape ditahan saat exam active.
- App resume selalu melakukan status refresh sebelum lanjut bila risk threshold tercapai.
- Operator/debug API base URL tetap berada di surface non-primer siswa.
- Optional platform integrity checks boleh dievaluasi, tetapi hanya sebagai risk signal.

Acceptance:

- Tidak ada wording yang menjanjikan perangkat terkunci penuh pada BYOD.
- Root/debug/integrity detection, bila ditambahkan, tidak menjadi satu-satunya dasar blokir siswa.
- Guidance tetap menjelaskan apa yang harus dilakukan siswa dan pengawas.

### Phase 5 - Managed Device Option

Fase ini hanya relevan bila sekolah menyediakan perangkat milik sekolah.

Kemungkinan:

- Android Enterprise atau device-owner lock task.
- App allowlist.
- Konfigurasi jaringan dan update APK terpusat.
- Kebijakan factory reset dan enrollment.

Non-goal untuk baseline BYOD:

- Memaksa perangkat pribadi menjadi device-owner.
- Mengklaim kiosk tanpa MDM/perangkat terkelola.
- Mengganti Flutter BYOD safeguards yang sudah ada.

Acceptance:

- Ada keputusan operasional sekolah tentang perangkat terkelola.
- Ada dokumen terpisah untuk enrollment, support, privacy, dan rollback.
- Baseline BYOD tetap berjalan untuk perangkat siswa pribadi.

## Event Reference

Event prioritas:

- `app_backgrounded`
- `app_resumed`
- `resume_gate_shown`
- `heartbeat_failed`
- `heartbeat_recovered`
- `stale_connection_attention`
- `stale_connection_escalated`
- `answer_saved_local`
- `answer_sync_recovered`
- `manual_submit_blocked`
- `screenshot_attempt`
- `restore_attempted`
- `restore_failed`

Minimal data:

```json
{
  "event_type": "stale_connection_attention",
  "data": {
    "occurred_at": "2026-05-08T01:20:00Z",
    "sequence": 18,
    "connection_state": "waspada",
    "stale_seconds": 120,
    "pending_answer_count": 1
  }
}
```

Data yang dilarang:

- Password.
- Answer key.
- Token peserta lain.
- Exam token di nested payload event.
- Raw screenshot atau media perangkat siswa.
- PII perangkat yang tidak diperlukan untuk operasi ujian.

## BYOD Acceptance Criteria

Sebelum anti-cheat Flutter dianggap siap untuk rehearsal besar:

- Token login, heartbeat, answer save, restore, and submit lulus di beberapa perangkat Android nyata.
- App switch/background memunculkan warning dan event.
- Koneksi terputus memunculkan status yang dapat dipahami siswa/pengawas.
- Pending answer tidak hilang saat app restart.
- Submit tidak diizinkan ketika ada pending sync yang membuat hasil tidak terpercaya.
- Pengawas dapat melihat sinyal penting dari backend/web-admin.
- Docs dan UI tidak menyebut jaminan kiosk untuk BYOD.
