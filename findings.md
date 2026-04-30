# Findings — `mtsn2kolut-super-app`

Dokumen ini merangkum backlog review yang **masih aktif** per 2026-05-01. Temuan yang sudah ditutup tidak lagi dicampur ke daftar utama agar prioritas patch tetap akurat.

---

## Ringkasan prioritas aktif

### Low priority
1. Jaga hygiene working tree dan commit tetap fokus per-slice

---

## 1) LOW — Working tree dan runtime data perlu tetap dijaga dari commit

**Area:** repo root / operational hygiene

**Masalah:**
Walau folder runtime sensitif utama sekarang sudah di-ignore, repo masih bisa terlihat “dirty” karena slice kerja paralel atau artefak lokal yang memang belum siap rilis.

**Arah patch / housekeeping:**
- pertahankan `.gitignore` untuk runtime/local asset folders
- pastikan file data lokal/sensitif tidak ikut commit
- jaga commit tetap fokus per-slice

---

## Temuan yang sudah ditutup

- Multipart upload `cbt/assets` proxy tidak lagi memaksa `Content-Type: application/json`; `FormData` sekarang memakai auth header tanpa content-type paksa.
- Route `pusaka/worker/status` sudah memakai proxy canonical ke `/api/pusaka/worker/status`, bukan lagi fetch ke `${event.url.origin}/health`.
- Raw internal error backend tidak lagi dibocorkan mentah ke client; respons 500 sekarang generik, detail tetap di log server.
- Rate limiting sekarang sudah thread-safe, memakai cleanup TTL, membaca forwarded IP, dan dipasang pada login/refresh/public registration/exam login.
- Escape hatch `as any` pada auth locals SvelteKit sudah dihapus; auth user sekarang typed lewat shared `AuthUser`.
- Metadata mobile dasar sudah dirapikan dari scaffold default (`pubspec` description dan Android app label).
- Surface user-facing protected/admin route tidak lagi menerima bypass internal key; route CBT asset file sekarang menerima hanya JWT user nyata atau `exam_token` peserta aktif.
- Snapshot exam mobile tidak lagi menyimpan payload sensitif bersama metadata restore di `SharedPreferences`; token, fingerprint, jawaban, dan pending answers sekarang dipisah ke secure storage, dengan fallback baca snapshot legacy untuk migrasi mulus.
- Fingerprint mobile sekarang diposisikan eksplisit sebagai telemetry hint BYOD, bukan identitas kuat perangkat.
- Field `Alamat server API` tidak lagi menjadi input utama siswa; sekarang tersembunyi di panel `Pengaturan Operator` yang dibuka hanya saat diperlukan.
- Worker PUSAKA tidak lagi menumpuk semua concern pada satu `index.ts`; config, logging, parser, HTTP client, Playwright runner, dan supervisor loop sekarang dipisah ke modul yang lebih kecil.
- `ExamShellScreen` tidak lagi memegang semua widget support dan perhitungan state koneksi di satu file; connection view-model dan widget presentational sekarang dipisah ke unit terpisah, dan file screen utama turun signifikan.
- `.gitignore` sekarang mencakup `services/core-api/data/` agar runtime data lokal tidak mudah ikut terseret ke commit.

---

## Urutan patch yang disarankan sekarang

### Batch 1 — hygiene
- [ ] Jaga commit tetap fokus dan jangan ikut membawa perubahan parallel worktree yang belum siap rilis
