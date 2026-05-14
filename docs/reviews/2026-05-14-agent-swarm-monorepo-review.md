# Review Monorepo MTsN 2 Kolut Super App — Agent Swarm

Tanggal: 2026-05-14 00:46 WITA
Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
Branch saat review: `feature/comprehensive-improvements`
Mode: read-only review, tidak mengubah source aplikasi.

## Swarm spesialis

1. Frontend/SvelteKit & UI/UX maintainability.
2. Backend Go core-api architecture & test coverage.
3. Database/migrations/sqlc/data lifecycle.
4. Security/privacy/auth/PII/CBT token handling.
5. DevOps/operations/performance/release readiness.

## Baseline cepat

- Web-admin: SvelteKit/Svelte 5, 489 route files, 316 BFF `+server.ts` routes.
- Backend: Go core-api, chi/pgx/sqlc, 262 Go files, 108 Go test files, 97 migrations.
- DB query generation: `sqlc vet` lulus menurut agent DB; generated code sinkron.
- Backend tests: `go test ./...` lulus menurut agent backend.
- Web check: `npm run check` lulus menurut agent frontend.
- Web unit tests: `npm run test:unit` gagal 6 test dari 464 menurut agent frontend.
- Working tree punya untracked plan/seed scripts yang tidak disentuh.

## Executive verdict

Monorepo ini sudah berkembang menjadi sistem madrasah operasional yang cukup matang: domain luas, boundary SvelteKit BFF vs Go API relatif terjaga, auth/RBAC serius, backup center sudah defensif, dan test backend cukup kuat. Namun ukuran sistem sudah melewati titik di mana refactor ad-hoc aman. Risiko terbesar sekarang bukan satu bug tunggal, tetapi kombinasi: test frontend merah, file UI sangat besar, beberapa pola keamanan/operasional belum distandardisasi, log/artifact runtime berada di repo checkout, dan seed production-like masih untracked.

Prioritas saya: stabilisasi quality gates dan operational hygiene dulu, lalu pecah modul besar secara aman per sprint.

## P0 — harus ditangani dulu

### 1. Web unit tests merah

- Command yang dilaporkan agent frontend: `npm run test:unit`.
- Hasil: 6 test file gagal, 6 test gagal dari 464.
- Area gagal:
  - `apps/web-admin/src/routes/public-load.test.ts`
  - `apps/web-admin/src/lib/cbt/proctor-evidence.test.ts`
  - `apps/web-admin/src/lib/client/active-api-aliases.test.ts`
  - `apps/web-admin/src/routes/students/student-account-actions.test.ts`
  - `apps/web-admin/src/routes/settings/users/users-rbac-scope.test.ts`
  - `apps/web-admin/src/routes/settings/analytics/page-source.test.ts`

Rekomendasi:
- Jadikan Sprint Stabilitas 1: hijaukan `npm run test:unit` tanpa mengubah behavior produksi kecuali memang bug.
- Kurangi test berbasis copy literal, pindahkan ke assertion kontrak behavior/API path/permission.
- Tegaskan keputusan kompatibilitas `/api/cbt/*`: masih live deprecated compatibility atau benar-benar dihapus.

### 2. Secret nyata berada di file `.env` lokal core-api

- File: `services/core-api/.env`.
- Agent security menemukan `JWT_SECRET` nyata; nilai tidak dicetak.

Rekomendasi:
- Pastikan `.env` tidak tracked dan permission file ketat.
- Jika pernah masuk backup/artifact/log, rotasi `JWT_SECRET`.
- Tambahkan secret scanning CI/pre-commit: gitleaks/trufflehog.

### 3. Seed UTS production-like masih untracked dan tanpa guard lingkungan

- File:
  - `scripts/seed-uts-genap-2025-2026.sql`
  - `scripts/seed-uts-genap-2025-2026-sessions.sql`

Risiko:
- Salah eksekusi dapat membuat/mengubah event, paket, sesi, ruang, peserta, dan token CBT resmi.

Rekomendasi:
- Pindahkan ke folder operasional terkontrol dengan SOP, atau commit sebagai artifact resmi yang direview.
- Tambahkan guard approval variable, target DB check, dry-run/apply mode, preview count, dan audit output.

### 4. Log runtime besar dan tidak terlihat ada rotasi/capping PM2

- File/path:
  - `services/logs/backend-out-0.log` sekitar 513 MB.
  - `logs/backend-out-1.log` sekitar 56 MB.
  - `ecosystem.config.cjs`, `deploy/pm2/*.config.cjs`.

Rekomendasi:
- Pasang `pm2-logrotate` atau system `logrotate`.
- Pindahkan log ke path operasional di luar repo.
- Tambah disk/log growth check sebelum deploy.

## P1 — prioritas tinggi

### Frontend/SvelteKit

1. File UI terlalu besar:
   - `SoalWorkspacePage.svelte` sekitar 4551 baris.
   - `asesmen/sesi/[id]/+page.svelte` sekitar 2720 baris.
   - `governance/+page.svelte` sekitar 2370 baris.
   - `document-cycles/+page.svelte` sekitar 2052 baris.
   - `grades/+page.svelte` sekitar 2000 baris.

   Rekomendasi:
   - Ekstrak `*.client.ts`, `*.model.ts`, dan subkomponen per panel/table/dialog.
   - Mulai dari Bank Soal dan Asesmen karena domain high-risk.

2. BFF routes terlalu banyak dan berulang:
   - 316 `+server.ts` routes.

   Rekomendasi:
   - Buat helper/factory proxy per domain.
   - Jika pakai catch-all, wajib allowlist path + method, jangan generic proxy bebas.

3. Public route policy dobel:
   - `apps/web-admin/src/routes/+layout.svelte`
   - `apps/web-admin/src/lib/server/route-access.ts`

   Rekomendasi:
   - Satukan ke `src/lib/routes/public-policy.ts`.

4. Proctoring stream polling 1.5 detik berisiko beban ujian:
   - `apps/web-admin/src/lib/server/cbt-backend-proxy/proctoring-stream.ts`

   Rekomendasi:
   - Backend-native SSE/WebSocket, atau minimal backoff/rate budget/cursor.

### Backend Go

1. Sanitasi HTML Bank Soal masih regex-based:
   - `services/core-api/internal/service/cbt_question.go`
   - `services/core-api/internal/service/cbt_question_authoring.go`

   Rekomendasi:
   - Pindahkan ke `bluemonday` allowlist policy.
   - Tambah test bypass `javascript:`, SVG/onload, encoded payload, malformed tags.

2. JSON handler banyak belum memakai `MaxBytesReader`/decode policy konsisten:
   - contoh `student.go`, `academic.go`, `cbt_session.go`, `kesiswaan.go`, `rbac.go`.

   Rekomendasi:
   - Buat helper decode JSON standar dengan limit, optional `DisallowUnknownFields`, dan no trailing JSON.

3. `cbt_session.go` terlalu besar dan authz orchestration kompleks:
   - `services/core-api/internal/handler/cbt_session.go` sekitar 2507 baris.

   Rekomendasi:
   - Pecah per bounded context: participants, rooms, proctoring, results.
   - Pindahkan rule domain authorization ke service policy.

4. Kesiswaan `ClassOptions` tidak scope ke kelas guru:
   - `services/core-api/internal/handler/kesiswaan.go`
   - `services/core-api/db/queries/kesiswaan.sql`

   Rekomendasi:
   - Tambah query/service scoped by teacher dan test akses guru.

5. Error mapping masih campuran sentinel + string matching:
   - `internal/handler/errors.go`
   - `internal/domain/errors.go`

   Rekomendasi:
   - Perluas typed/sentinel domain errors dan kurangi string matching.

### Database/data lifecycle

1. Backup file permission belum dipaksa privat:
   - `deploy/backup-postgresql.sh`

   Rekomendasi:
   - Tambahkan `umask 077`, `chmod 600`, directory mode 700.

2. Restore plan aman karena non-eksekusi, tetapi command perlu lebih defensif:
   - `services/core-api/internal/service/system_backup.go`

   Rekomendasi:
   - Tambah `--single-transaction --exit-on-error`, checksum verification, restore-to-staging-first SOP.

3. Banyak migration `CREATE INDEX` non-concurrent:
   - Runner membungkus migration dalam transaksi: `services/core-api/db/scripts/apply_migrations.js`.

   Rekomendasi:
   - Tambah kategori non-transactional/online index migration dengan `CREATE INDEX CONCURRENTLY`.

4. Default `BASELINE_ON_EXISTING_SCHEMA=true` berisiko salah target DB:
   - `services/core-api/db/scripts/apply_migrations.js`.

   Rekomendasi:
   - Production normal harus eksplisit `false`; legacy adoption saja yang `true` dengan approval.

5. Search banyak pakai `ILIKE '%term%'`:
   - contoh `kesiswaan.sql`, `governance.sql`, `cbt_questions.sql`, `users.sql`.

   Rekomendasi:
   - Tambah `pg_trgm` + GIN/GiST untuk domain besar, atau full-text search.

### Security/privacy

1. `InternalKeyOrJWT` berpotensi bypass bila dipakai pada route sensitif:
   - `services/core-api/internal/middleware/auth.go`.

   Rekomendasi:
   - Audit semua usage, pisahkan key per service/scope, audit actor internal, dan jangan pakai untuk user-facing route.

2. JWT method menerima semua HMAC, bukan pin HS256:
   - `services/core-api/internal/middleware/auth.go`.

   Rekomendasi:
   - Pin `jwt.SigningMethodHS256` atau allowed methods parser.

3. Cookie session SameSite=Lax tanpa CSRF token global untuk mutasi:
   - `apps/web-admin/src/hooks.server.ts`.

   Rekomendasi:
   - Tambah CSRF token/header untuk POST/PUT/PATCH/DELETE BFF.
   - Validasi Origin/Referer untuk mutasi.

4. Token peserta CBT plaintext dan muncul pada listing:
   - `services/core-api/db/queries/cbt_sessions.sql`.

   Rekomendasi:
   - Simpan hash token, listing masked, reveal token endpoint khusus dengan audit/rate limit.

5. BFF logging upstream error dapat bocorkan PII/secret:
   - `apps/web-admin/src/lib/server/api.ts`.

   Rekomendasi:
   - Redact error message sebelum log; structured logging field aman.

### DevOps/operations

1. PM2 config root vs deploy-specific tidak konsisten:
   - `ecosystem.config.cjs`
   - `deploy/pm2/*.config.cjs`

   Rekomendasi:
   - Jadikan deploy PM2 config sebagai single source of truth production.
   - Samakan path log absolut, env_file, kill_timeout, restart policy.

2. Runtime DB/data dan binary backup besar berada di checkout repo:
   - `services/core-api/data/pgdata/`
   - `services/core-api/bin/api.backup-*`

   Rekomendasi:
   - Pindahkan runtime data/log/binary backup ke luar repo; beri retensi.

3. CI belum cukup kuat:
   - `.github/workflows/ci.yml`.

   Rekomendasi:
   - Tambah web `npm run test:unit` dan `npm run build`.
   - Tambah backend build, worker test/build, optional mobile analyze/test.

4. Mobile APK release default auto-pick artifact terbaru:
   - `scripts/publish-mobile-apk.sh`.

   Rekomendasi:
   - Production wajib `--apk`, `--commit`, metadata eksplisit, sha256, version check, dry-run.

## P2 — backlog kualitas

- Ganti `SELECT *` di query sqlc dengan kolom eksplisit.
- Migrasi pagination domain besar dari OFFSET ke keyset pagination.
- Tambah contract tests untuk SQL scoping: parent portal, student portal, Kesiswaan, Bank Soal redaction.
- Standardisasi constructor service Go agar interface-friendly.
- Tambah observability: route latency/status, DB pool, worker heartbeat, backup age, disk/log growth.
- Keluarkan artifact tracked seperti `apps/web-admin/frontend.zip` bila bukan source canonical.

## Sprint rekomendasi

### Sprint 0 — Stabilitas & safety gate

Status implementasi: dikerjakan pada 2026-05-14.

- Hijaukan `npm run test:unit`.
  - Implementasi: test drift web-admin diselaraskan dengan kontrak UI/API compatibility saat ini.
- Tambah secret scanning.
  - Implementasi: `scripts/scan-secrets.py`, root npm script `security:scan-secrets`, dan CI job `security`.
- Bereskan untracked seed UTS: commit sebagai SOP guarded atau pindah/hapus dari workspace.
  - Implementasi: `scripts/run-uts-genap-seed-guarded.sh` sebagai wrapper dry-run/apply dengan confirmation, expected DB name, konfirmasi backup, dan preview count.
- Pasang logrotate dan cek disk.
  - Implementasi: runbook `docs/operations/log-rotation.md` dan contoh config `deploy/logrotate/mtsn2kolut-super-app.conf`. Pemasangan OS/PM2 tetap manual, tidak dieksekusi dari sprint ini.

### Sprint 1 — Security hardening

- Pin JWT HS256.
- CSRF guard mutasi BFF.
- Redacted BFF logging.
- Decode JSON helper + MaxBytesReader untuk endpoint mutasi prioritas.
- Audit `InternalKeyOrJWT`.

### Sprint 2 — CBT/Bank Soal hardening

- Bluemonday sanitizer Bank Soal.
- Mask/hash token peserta CBT.
- Rate limit `/api/exam/*`.
- Proctoring stream backoff atau SSE backend.

### Sprint 3 — Frontend maintainability

- Split `SoalWorkspacePage.svelte`.
- Split `asesmen/sesi/[id]/+page.svelte`.
- Shared public route policy.
- BFF proxy helper/factory domain.

### Sprint 4 — DB & backup lifecycle

Status implementasi: dikerjakan pada 2026-05-14 di worktree `sprint-4`; tidak deploy, tidak restart PM2, tidak menjalankan migration production, dan tidak commit.

- Backup permission hardening.
  - Implementasi: `deploy/backup-postgresql.sh` memakai `umask 077`, directory mode `700`, file/log/checksum/lock mode `600`, dan tetap idempotent dengan target backup yang sama.
- Restore plan command/SOP defensive.
  - Implementasi: `services/core-api/internal/service/system_backup.go`, `deploy/DEPLOY.md`, dan `docs/deployment.md` menambahkan checksum verification, `pg_restore --single-transaction --exit-on-error`, serta restore-to-staging-first sebelum production restore dipertimbangkan. UI tetap hanya menghasilkan validasi/command plan, bukan menjalankan restore otomatis.
- Non-transactional migration support untuk online index.
  - Implementasi: `services/core-api/db/scripts/apply_migrations.js` default tetap transactional; mode non-transactional hanya aktif dengan marker eksplisit `-- mtsn2kolut:migration non-transactional` dan dibatasi ke `CREATE/DROP INDEX CONCURRENTLY`.
- `pg_trgm` untuk search domain besar.
  - Implementasi: migration `097_pg_trgm_extension.sql`, `098_pg_trgm_core_search_indexes.sql`, dan `099_pg_trgm_operations_search_indexes.sql` menyiapkan extension dan online trigram index batch untuk Bank Soal, siswa/pegawai, perpustakaan, arsip, dan tata kelola. Migration dibuat sebagai artifact rollout normal dan belum diaplikasikan ke production dari sprint ini.

### Sprint 5 — DevOps/release readiness

- PM2 config single source of truth.
- Runtime/log/binary backups outside repo.
- CI expanded to build/test gates.
- Mobile release explicit artifact enforcement.

## Catatan penting

- Review ini read-only; tidak ada source yang diubah.
- Jangan menampilkan/menyimpan secret. Temuan `.env` sengaja tidak mencetak nilai.
- Backend/core-api dan web-admin production tidak disentuh oleh review ini.
