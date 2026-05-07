# Findings — `mtsn2kolut-super-app`

Dokumen ini adalah ledger review aktif per 2026-05-04 setelah final verification pass lintas backend, web, worker, dan mobile.

---

## Ringkasan Saat Ini

Tidak ada temuan **High** yang masih terbuka dari review 2026-05-01. Tiga temuan aktif sebelumnya sudah tertutup di kode:

- Library/Inventory sekarang digate untuk `admin` dan `staf` di backend route group dan BFF route-access.
- Job PUSAKA `running` yang stale sekarang dipulihkan backend sebelum claim/scheduler berjalan.
- File foto siswa Kesiswaan sekarang mengecek scope siswa melalui `CanReadKesiswaanStudentPhoto`.

Prioritas yang tersisa bersifat operasional dan peningkatan mutu, bukan bug akses kritis yang diketahui. Checkpoint 5 juga menutup gap compliance berupa helper Podman DB baru yang sempat masuk working tree; target/script itu dibatalkan agar tidak menambah tooling container atau mengubah topology deploy. Checkpoint 6 menyelaraskan ledger dengan security hardening lintas backend/web/mobile/worker: trusted proxy CIDR untuk rate-limit IP forwarding, worker API key/timeout/log/screenshot hygiene, upload allowlist, dan `nosniff` response headers.

**Catatan sqlc final verification:** `make db-sqlc` sekarang memprovision `sqlc` repo-local ke `.tools/bin/sqlc` bila binary belum tersedia, lalu menjalankan generate terhadap `services/core-api/db/sqlc.yaml`. Direktori `.tools/` di-ignore agar binary tool tidak masuk source control.

---

## Rekomendasi Aktif

### Medium — Jalankan CBT smoke rehearsal di staging

**Area:** CBT Bank Soal, sesi ujian, Flutter token login.

**Status:** Belum bisa dibuktikan penuh tanpa akun/data staging nyata.

**Tindakan:**
Gunakan [docs/cbt-smoke-checklist.md](./docs/cbt-smoke-checklist.md) setelah backend/frontend deploy untuk memastikan:

- admin/guru export sesuai role
- kunci jawaban tidak bocor ke guru non-penulis
- token minutes hanya terlihat sesuai scope
- scoring tidak mengunci peserta yang belum submit
- Flutter bisa login, heartbeat, simpan jawaban, dan submit

### Medium — Pertimbangkan browser smoke test role-based

**Area:** `apps/web-admin`.

**Status:** Unit test sudah mengunci helper export, stream CSV, dan redirect `/cbt/questions`; login browser admin/guru belum otomatis.

**Tindakan:**
Jika checklist manual mulai sering dipakai, ubah bagian stabil menjadi Playwright smoke test role-based. Jangan menambah suite e2e besar sebelum data seed dan akun uji staging jelas.

### Medium — Validasi env security hardening saat deploy

**Area:** backend reverse proxy, PUSAKA worker, upload/file serving.

**Status:** Code hardening sudah masuk working tree, tetapi nilai environment produksi harus diverifikasi di VPS masing-masing.

**Tindakan:**
Sebelum restart produksi, cek `TRUSTED_PROXY_CIDRS` di backend, samakan `WORKER_API_KEY` backend/worker, set `WORKER_API_TIMEOUT_MS` sesuai kondisi jaringan, dan pastikan `WORKER_LOG_PATH`/`SCREENSHOT_DIR` berada di direktori non-public dengan permission/retention terbatas.

### Low — Perkuat seat plan constraint

**Area:** CBT session rooms / participants.

**Status:** Backend sudah punya readiness dan assignment guard, tetapi uniqueness kursi per ruang masih terutama dijaga di service/UI.

**Tindakan:**
Pertimbangkan constraint database unik untuk `(room_id, seat_no)` ketika format seat number sudah stabil di operasional.

### Low — Drill restore PostgreSQL secara periodik

**Area:** backup operations.

**Status:** `make ops-backup` sekarang memanggil `deploy/backup-postgresql.sh` (custom-format dump dengan SHA-256 checksum, retention 30 hari, `flock` overlap guard). Restore drill prosedur tertulis di `deploy/DEPLOY.md`.

**Tindakan:**
Jalankan restore drill ke database staging tiap rilis besar atau minimal kuartalan. Catat RTO aktual sebagai baseline.

### Low — Tunda CBT Engine sampai rehearsal/load signal cukup

**Area:** arsitektur CBT runtime.

**Status:** Rencana extraction `services/cbt-engine` sudah terdokumentasi di `PLAN.md`, tetapi belum diimplementasikan.

**Tindakan:**
Jangan mengubah aturan PostgreSQL ownership sekarang. Mulai extraction hanya jika rehearsal/load test membuktikan live exam runtime perlu dipisah dari Core API.

---

## Temuan 2026-05-01 Yang Sudah Ditutup

### Library/Inventory RBAC

**Sebelumnya:** route backend dan BFF belum konsisten membatasi Library/Inventory.

**Kondisi sekarang:**

- `services/core-api/cmd/api/main.go` mengelompokkan `/api/library/*` dan `/api/inventory/*` di bawah `requireStaff`.
- `apps/web-admin/src/hooks.server.ts` memakai `isStaffOperationPath` untuk `/library`, `/inventory`, `/api/library`, dan `/api/inventory`.
- `apps/web-admin/src/lib/server/route-access.test.ts` mengunci helper path staff operation.

### PUSAKA stale running job

**Sebelumnya:** job yang mati setelah claim bisa tertahan di status `running`.

**Kondisi sekarang:**

- `services/core-api/db/queries/jobs.sql` memiliki `RecoverStaleRunningJobs`.
- `services/core-api/internal/service/job.go` menjalankan recovery sebelum `Claim`.
- `services/core-api/internal/service/scheduler.go` juga menjalankan recovery sebelum schedule claim.

### Scope foto siswa Kesiswaan

**Sebelumnya:** file foto siswa belum mengikuti scope list siswa.

**Kondisi sekarang:**

- `services/core-api/db/queries/kesiswaan.sql` memiliki `CanReadKesiswaanStudentPhoto`.
- `services/core-api/internal/service/kesiswaan.go` mengecek photo URL dan `teacher_employee_id`.
- `services/core-api/internal/handler/kesiswaan.go` menolak file foto yang tidak berada dalam scope.

---

## Catatan Operasional Git

- Runtime/local data harus tetap di-ignore: `services/core-api/data/`, `services/logs/`, root `logs/`, `*.db`, `*.sqlite`, `*.db-shm`, `*.db-wal`, `*.sqlite-shm`, dan `*.sqlite-wal`.
- Commit tetap dibuat per-slice dan tidak mencampur runtime artifact.
- `PLAN.md` harus diperbarui pada setiap sprint/checkpoint.
