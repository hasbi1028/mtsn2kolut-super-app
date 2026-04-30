# Findings — `mtsn2kolut-super-app`

Dokumen ini merangkum backlog review yang **masih aktif** per 2026-05-01. Temuan yang sudah ditutup tidak lagi dicampur ke daftar utama agar prioritas patch tetap akurat.

---

## Ringkasan prioritas aktif

### High priority
1. Review dan harden public serving untuk file asset CBT
2. Kurangi blast radius `INTERNAL_API_KEY`

### Medium priority
3. Pecah `services/pusaka-worker/src/index.ts` menjadi modul yang lebih kecil
4. Refactor `apps/mobile/lib/src/screens/exam_shell_screen.dart`
5. Hardening penyimpanan snapshot exam di mobile
6. Jangan perlakukan fingerprint device mobile saat ini sebagai identitas kuat

### Low priority
7. Evaluasi kebutuhan field API base URL yang bisa diubah siswa di mobile
8. Bersihkan working tree dan pastikan file data sensitif tidak ikut commit

---

## 1) HIGH — File asset CBT masih bisa diakses lewat route file berbasis ID

**Area:** `services/core-api`

**Files:**
- `services/core-api/cmd/api/main.go`
- `services/core-api/internal/handler/cbt_question_asset.go`

**Masalah:**
Route `/api/cbt/assets/{id}/file` sekarang memang sudah tidak public polos, tetapi tetap perlu ditinjau sebagai boundary security khusus karena ia dipakai lintas konteks admin/guru dan peserta ujian.

**Kenapa ini penting:**
Asset soal adalah bagian dari boundary CBT. Akses file berbasis ID saja tetap perlu dipastikan hanya terbuka untuk konteks yang benar:
- admin/guru terautentikasi
- peserta ujian aktif dengan exam token valid

**Arah patch:**
- review kembali kontrak `ExamTokenOrJWT(...)` pada route file asset
- pastikan tidak ada jalur bypass yang terlalu longgar
- pertimbangkan signed URL atau tokenized file access kalau kebutuhan media makin kaya

**Acceptance check:**
- asset tidak bisa diakses di luar konteks admin/guru atau peserta ujian aktif
- render media di UI authoring dan mobile exam tetap berjalan

---

## 2) HIGH — `INTERNAL_API_KEY` masih punya blast radius besar

**Area:** `services/core-api`

**Files:**
- `services/core-api/internal/middleware/auth.go`
- `services/core-api/cmd/api/main.go`

**Masalah:**
`X-Internal-Key` masih dipakai sebagai bypass umum pada `InternalKeyOrJWT(...)` dan `RequireAdmin(...)`.

**Kenapa ini bermasalah:**
Satu shared secret masih membuka surface yang luas. Jika bocor, banyak route authenticated/admin ikut terbuka.

**Arah patch:**
- batasi internal key hanya pada route yang benar-benar perlu bypass internal
- audit route yang sebenarnya sudah aman memakai JWT user biasa
- pertimbangkan pemisahan scope/key bila bypass tetap diperlukan

**Acceptance check:**
- route user-facing biasa selalu mengandalkan JWT user
- bypass internal hanya berlaku pada route yang eksplisit diizinkan
- admin route tidak otomatis terbuka hanya karena shared internal key

---

## 3) MEDIUM — Worker PUSAKA masih terlalu banyak tanggung jawab dalam satu file

**Area:** `services/pusaka-worker`

**File:**
- `services/pusaka-worker/src/index.ts`

**Masalah:**
Satu file besar masih memegang config, backend client, supervisor loop, Playwright flow, parsing, logging, dan shutdown.

**Arah patch:**
Pecah minimal menjadi:
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

## 4) MEDIUM — `ExamShellScreen` masih menjadi hotspot regresi mobile

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

## 5) MEDIUM — Snapshot exam mobile masih disimpan plaintext di SharedPreferences

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

## 6) MEDIUM — Fingerprint device mobile saat ini masih lemah

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

## 7) LOW — Field API base URL masih editable di mobile student app

**Area:** `apps/mobile`

**File:**
- `apps/mobile/lib/src/screens/exam_login_screen.dart`

**Masalah:**
Siswa masih dapat mengubah `Alamat server API` langsung dari UI.

**Arah patch:**
- jika ini hanya untuk trial internal, pertimbangkan mode operator/debug saja
- untuk rilis sekolah, pertimbangkan base URL fixed atau tersembunyi di mode admin/operator

---

## 8) LOW — Working tree dan runtime data perlu tetap dijaga dari commit

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

---

## Urutan patch yang disarankan sekarang

### Batch 1 — security boundary
- [ ] Review file-serving asset CBT
- [ ] Kurangi blast radius `INTERNAL_API_KEY`

### Batch 2 — mobile hardening
- [ ] Review local snapshot strategy di mobile
- [ ] Tegaskan role fingerprint sebagai telemetry hint
- [ ] Evaluasi API base URL editable

### Batch 3 — maintainability
- [ ] Pecah worker `index.ts`
- [ ] Refactor `ExamShellScreen`

### Batch 4 — hygiene
- [ ] Review `.gitignore` dan local runtime data hygiene
