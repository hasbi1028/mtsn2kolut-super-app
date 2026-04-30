# Findings — `mtsn2kolut-super-app`

Dokumen ini merangkum backlog review yang **masih aktif** per 2026-05-01. Temuan yang sudah ditutup tidak lagi dicampur ke daftar utama agar prioritas patch tetap akurat.

---

## Ringkasan prioritas aktif

### Medium priority
1. Pecah `services/pusaka-worker/src/index.ts` menjadi modul yang lebih kecil
2. Refactor `apps/mobile/lib/src/screens/exam_shell_screen.dart`
3. Hardening penyimpanan snapshot exam di mobile
4. Jangan perlakukan fingerprint device mobile saat ini sebagai identitas kuat

### Low priority
5. Evaluasi kebutuhan field API base URL yang bisa diubah siswa di mobile
6. Bersihkan working tree dan pastikan file data sensitif tidak ikut commit

---

## 1) MEDIUM — Worker PUSAKA masih terlalu banyak tanggung jawab dalam satu file

**Area:** `services/pusaka-worker`

**File:**
- `services/pusaka-worker/src/index.ts`

**Masalah:**
Satu file besar masih memegang config, backend client, supervisor loop, Playwright flow, parsing, logging, dan shutdown.

**Arah patch:**
- Pecah minimal menjadi:
- `config.ts`
- `api-client.ts`
- `worker-supervisor.ts`
- `pusaka-runner.ts`
- `parsers.ts`
- `logger.ts`

**Acceptance check:**
- `index.ts` menjadi entrypoint tipis
- concern worker lebih modular dan mudah dites

---

## 2) MEDIUM — `ExamShellScreen` masih menjadi hotspot regresi mobile

**Area:** `apps/mobile`

**File:**
- `apps/mobile/lib/src/screens/exam_shell_screen.dart`

**Masalah:**
Screen ini masih memegang lifecycle, timer, sync/degraded logic, persistence, submit rules, telemetry, dan UI tree sekaligus.

**Arah patch:**
- extract sync/degraded state logic
- extract persistence/session orchestration
- extract presentational widgets
- pertimbangkan controller/service yang lebih tipis dan testable

**Acceptance check:**
- file screen utama berkurang signifikan
- logic penting pindah ke unit yang bisa diuji terpisah

---

## 3) MEDIUM — Snapshot exam mobile masih disimpan plaintext di SharedPreferences

**Area:** `apps/mobile`

**File:**
- `apps/mobile/lib/src/exam_session_store.dart`

**Masalah:**
Snapshot masih menyimpan token, jawaban, dan metadata penting dalam JSON plaintext di `SharedPreferences`.

**Kenapa ini bermasalah:**
Pada skenario BYOD, data lokal lebih berisiko diakses, terutama pada device shared, rooted, atau backup tertentu.

**Arah patch:**
- minimalkan data yang benar-benar perlu disimpan
- pisahkan metadata non-sensitif dan data sensitif
- pertimbangkan encrypt/secure storage untuk token/jawaban/snapshot penting

**Acceptance check:**
- restore UX tetap cukup
- data sensitif lokal lebih terlindungi dibanding plaintext saat ini

---

## 4) MEDIUM — Fingerprint device mobile saat ini masih lemah

**Area:** `apps/mobile`

**File:**
- `apps/mobile/lib/src/screens/exam_login_screen.dart`

**Masalah:**
Fingerprint masih dibuat dari `Platform.operatingSystem` dan `Platform.localHostname`.

**Kenapa ini bermasalah:**
Ini cukup sebagai telemetry hint, tapi lemah jika diperlakukan sebagai identity binding yang kuat.

**Arah patch:**
- dokumentasikan jelas bahwa fingerprint ini hanya hint
- pastikan backend/client tidak membuat keputusan security penting yang bergantung hanya pada fingerprint ini
- jika butuh device identity yang lebih baik, definisikan strategi BYOD yang realistis

**Acceptance check:**
- role fingerprint di arsitektur jelas
- tidak ada keputusan security penting yang bergantung hanya pada fingerprint saat ini

---

## 5) LOW — Field API base URL masih editable di mobile student app

**Area:** `apps/mobile`

**File:**
- `apps/mobile/lib/src/screens/exam_login_screen.dart`

**Masalah:**
Siswa masih dapat mengubah `Alamat server API` langsung dari UI.

**Arah patch:**
- jika ini hanya untuk trial internal, pertimbangkan mode operator/debug saja
- untuk rilis sekolah, pertimbangkan base URL fixed atau tersembunyi di mode admin/operator

---

## 6) LOW — Working tree dan runtime data perlu tetap dijaga dari commit

**Area:** repo root / operational hygiene

**Masalah:**
Repo masih punya runtime/local data paths yang mudah ikut terseret ke commit jika tidak disiplin, terutama:
- `services/core-api/data/`
- migrasi/artefak lokal yang belum siap rilis

**Arah patch / housekeeping:**
- review `.gitignore` untuk runtime/local asset folders
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

---

## Urutan patch yang disarankan sekarang

### Batch 1 — mobile hardening
- [ ] Review local snapshot strategy di mobile
- [ ] Tegaskan role fingerprint sebagai telemetry hint
- [ ] Evaluasi API base URL editable

### Batch 2 — maintainability
- [ ] Pecah worker `index.ts`
- [ ] Refactor `ExamShellScreen`

### Batch 3 — hygiene
- [ ] Review `.gitignore` dan local runtime data hygiene
