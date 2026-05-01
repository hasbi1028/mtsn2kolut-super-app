# Findings — `mtsn2kolut-super-app`

Dokumen ini merangkum backlog review yang **masih aktif** per 2026-05-01. Temuan yang sudah ditutup tidak lagi dicampur ke daftar utama agar prioritas patch tetap akurat.

---

## Ringkasan prioritas aktif

Ada 3 temuan aktif dari review monorepo 2026-05-01:

- **High:** RBAC modul Library/Inventory bocor karena route backend dan BFF belum role-gated.
- **High:** Job PUSAKA bisa terkunci permanen di status `running` bila worker mati setelah claim.
- **Medium:** File foto siswa Kesiswaan belum mengikuti scope wali/guru yang sama dengan list siswa.

---

## High — Library/Inventory belum role-gated

**Area:** `services/core-api`, `apps/web-admin`

**Dampak:**
Modul `/library/*` seharusnya hanya untuk `admin` dan `staf`, tetapi route backend saat ini hanya berada di grup JWT tanpa pengecekan role. Hal yang sama terlihat pada Inventory. Sidebar memang menyembunyikan menu untuk role lain, tetapi akses langsung ke URL/API masih mungkin bila user punya token valid.

**Bukti review:**
- Backend mendaftarkan route Library di `services/core-api/cmd/api/main.go:299` dan Inventory di `services/core-api/cmd/api/main.go:310`.
- Handler Library seperti `ListBooks`, `CreateBook`, `UpdateBook`, dan `CreateLoan` belum memanggil allow-rule role. Helper `libraryAccessAllowed` ada di `services/core-api/internal/handler/library.go:191`, tetapi belum dipakai di handler.
- Handler Inventory seperti `ListItems`, `CreateItem`, `UpdateItem`, dan `DeleteItem` belum punya gate role eksplisit.
- BFF `apps/web-admin/src/hooks.server.ts:31` belum memasukkan `/library`, `/inventory`, `/api/library`, dan `/api/inventory` ke daftar prefix admin/staf.

**Rekomendasi patch:**
Pasang role gate eksplisit di backend untuk semua handler Library/Inventory, lalu tambahkan guard BFF agar direct navigation/proxy ikut konsisten. Sidebar tetap dianggap hanya UX, bukan security boundary.

---

## High — Job PUSAKA `running` bisa macet permanen

**Area:** `services/core-api`, `services/pusaka-worker`

**Dampak:**
Jika worker sudah berhasil claim job lalu proses mati sebelum `complete` atau `fail`, row job akan tetap `running`. Query claim hanya mengambil `queued`/`failed` yang sudah due, sedangkan pembuatan job baru ditahan oleh unique partial index untuk status `queued`/`running`. Akibatnya pekerjaan untuk employee dan tanggal yang sama bisa berhenti tanpa retry otomatis.

**Bukti review:**
- Unique partial index di `services/core-api/db/queries/jobs.sql:28` menahan duplikasi untuk status `queued` dan `running`.
- `ClaimJob` di `services/core-api/db/queries/jobs.sql:36` tidak mengambil job `running` yang stale.
- Worker shutdown timeout di `services/pusaka-worker/src/worker-supervisor.ts:218` bisa meninggalkan job yang sudah diclaim.
- `failJob` di `services/pusaka-worker/src/api-client.ts:118` hanya melempar error/log bila gagal melapor, sehingga backend tetap perlu mekanisme pemulihan stale claim.

**Rekomendasi patch:**
Tambahkan lease/heartbeat atau stale-timeout requeue di backend. Minimal: job `running` dengan `updated_at` terlalu lama dapat dikembalikan ke `queued`/`failed` secara terkontrol sebelum claim berikutnya.

---

## Medium — Foto siswa Kesiswaan belum mengikuti scope data siswa

**Area:** `services/core-api/internal/handler/kesiswaan.go`, `services/core-api/internal/service/kesiswaan.go`

**Dampak:**
List siswa Kesiswaan sudah scoped untuk guru berdasarkan kelas yang dia ampu, tetapi endpoint file foto siswa hanya memeriksa akses baca Kesiswaan secara umum. Guru yang punya akses Kesiswaan berpotensi mengambil foto siswa lain bila mengetahui ID/path yang valid.

**Bukti review:**
- Query list siswa scoped ada di `services/core-api/db/queries/kesiswaan.sql:121`.
- Handler foto siswa ada di `services/core-api/internal/handler/kesiswaan.go:193`.
- Service file foto ada di `services/core-api/internal/service/kesiswaan.go:182`.

**Rekomendasi patch:**
Tambahkan verifikasi scope siswa sebelum melayani file foto. Untuk `admin`/`kesiswaan`, izinkan penuh; untuk guru biasa, cocokkan siswa terhadap kelas/mata pelajaran yang dia ampu.

---

## Catatan operasional — Working tree dan runtime data perlu tetap dijaga dari commit

**Area:** repo root / operational hygiene

**Status saat ini:**
- runtime/local data utama sekarang sudah di-ignore:
  - `services/core-api/data/`
  - `services/logs/`
  - root `logs/`
  - file `*.db` dan `*.sqlite`
- worktree masih bisa terlihat “dirty” bila ada slice kerja paralel yang memang belum siap dirilis bersama

**Prinsip housekeeping:**
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
- `.gitignore` sekarang juga mencakup `services/logs/` agar output runtime lokal service tidak ikut terseret ke commit.
- SQLite WAL/SHM runtime sudah dipisahkan dari source control: `.gitignore` mencakup `*.db-shm`, `*.db-wal`, `*.sqlite-shm`, dan `*.sqlite-wal`, serta sidecar yang sudah terlanjur tracked dibersihkan lewat commit hygiene.

---

## Urutan patch yang disarankan sekarang

1. Tutup RBAC Library/Inventory lebih dulu karena ini surface akses langsung.
2. Tambahkan recovery stale job PUSAKA sebelum mengandalkan retry operasional.
3. Samakan scope file foto siswa dengan scope list siswa.
