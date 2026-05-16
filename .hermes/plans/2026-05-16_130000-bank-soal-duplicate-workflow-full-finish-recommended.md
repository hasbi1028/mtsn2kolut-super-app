# Recommended Plan — Full Finish Duplikasi Draft Soal & Workflow Bank Soal

## Goal

Menyelesaikan perbaikan duplikasi draft soal dan error workflow Bank Soal sampai aman untuk production, termasuk:

- guard frontend terhadap double-submit,
- guard backend terhadap duplicate draft dan request paralel,
- workflow submit-review yang idempotent,
- validasi build/test,
- deploy aman,
- audit + cleanup duplikasi lama hanya yang benar-benar aman,
- dokumentasi/commit bersih.

## Status Saat Plan Dibuat

- Patch emergency `bda70d4 fix(bank-soal): prevent duplicate draft submissions` sudah pernah di-commit dan deploy.
- Review lanjutan menemukan patch emergency masih perlu diperkuat agar tahan race condition dan tidak false-positive duplicate.
- Patch review tambahan sudah dibuat lokal dan sudah sempat lulus validasi sebelumnya, tetapi command validasi terakhir terinterupsi (`exit 130`) saat `npm --prefix apps/web-admin run check`; jadi wajib diulang dari awal sebelum commit/deploy.
- File plan review sebelumnya sudah ada:
  - `.hermes/plans/2026-05-16_120850-bank-soal-draft-duplicate-guard.md`
- Plan recommended ini adalah rencana final eksekusi penuh.

## Prinsip Rekomendasi

1. **Jangan cleanup data dulu sebelum deploy guard baru.**
   - Kalau cleanup dilakukan sebelum guard kuat aktif, duplikasi bisa muncul lagi.
2. **Deploy backend guard dulu, baru cleanup duplikasi lama.**
3. **Cleanup harus conservative.**
   - Hapus hanya duplicate draft yang belum dipakai, belum masuk paket/sesi/jawaban, dan benar-benar identik menurut fingerprint aman.
4. **Selalu backup sebelum cleanup data.**
5. **Frontend restart hanya jika ada build web-admin baru.**
   - Patch review lanjutan fokus backend, tapi karena ada perubahan frontend dari patch emergency dan SvelteKit pernah dibuild, jika `npm run build` dijalankan maka PM2 web-admin harus langsung restart agar tidak stale chunks.

## Scope In

- Finalisasi patch review:
  - advisory lock untuk duplicate create draft,
  - fingerprint duplicate yang lebih lengkap,
  - submit-review idempotent,
  - test coverage duplicate/workflow.
- Commit patch review + plan.
- Deploy core-api production.
- Smoke test production.
- Backup database sebelum cleanup.
- Audit duplikasi lama.
- Cleanup duplicate draft lama yang aman.
- Verifikasi ulang setelah cleanup.

## Scope Out

- Tidak menghapus soal published/approved/used.
- Tidak mengubah isi paket CBT/UTS yang sudah ada.
- Tidak rewrite besar UI Bank Soal.
- Tidak push GitHub karena auth GitHub masih belum tersedia.

## Recommended Execution Plan

### Phase 0 — Preflight & Freeze

1. Pastikan working tree hanya berisi perubahan terkait:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git diff --stat
```

2. Pastikan tidak ada file seed/plan lama ikut commit tanpa sengaja.
3. Catat commit saat ini:

```bash
git log --oneline -5
```

4. Karena command validasi terakhir terinterupsi (`exit 130`), ulang validasi penuh di Phase 1.

### Phase 1 — Final Validation Lokal

Jalankan validasi penuh sebelum commit:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
gofmt -w internal/service/cbt_question.go internal/service/cbt_question_authoring.go internal/service/cbt_question_workflow.go internal/service/cbt_question_test.go
go test ./internal/service ./internal/handler ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-duplicate-workflow-review-fix ./cmd/api
cd ../..
git diff --check
```

Expected:

- `svelte-check`: 0 errors, 0 warnings.
- Go tests: PASS.
- Go build: PASS.
- `git diff --check`: no whitespace errors.

Jika gagal:

- Jangan deploy.
- Patch hanya error yang relevan.
- Ulang Phase 1.

### Phase 2 — Commit Patch Review

Jika Phase 1 PASS:

```bash
git add \
  services/core-api/db/queries/cbt_questions.sql \
  services/core-api/internal/repository/postgres/cbt_questions.sql.go \
  services/core-api/internal/service/cbt_question.go \
  services/core-api/internal/service/cbt_question_authoring.go \
  services/core-api/internal/service/cbt_question_test.go \
  services/core-api/internal/service/cbt_question_workflow.go \
  .hermes/plans/2026-05-16_120850-bank-soal-draft-duplicate-guard.md \
  .hermes/plans/2026-05-16_130000-bank-soal-duplicate-workflow-full-finish-recommended.md

git commit -m "fix(bank-soal): harden duplicate draft workflow guard"
```

Commit ini melengkapi `bda70d4`.

### Phase 3 — Deploy Backend Guard

Karena patch review lanjutan fokus backend dan sqlc generated code, deploy core-api dulu:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
stamp=$(date +%Y%m%d-%H%M%S)
cp -p bin/api "bin/api.backup-bank-soal-duplicate-workflow-${stamp}"
cp -p /tmp/core-api-bank-soal-duplicate-workflow-review-fix bin/api
pm2 restart mtsn2kolut-core-api --update-env
```

Health check:

```bash
for i in {1..20}; do
  curl -fsS http://127.0.0.1:8080/health && break
  sleep 0.5
done
pm2 status mtsn2kolut-core-api
```

Expected:

- core-api online.
- health `status: ok`, `db: connected`.

Rollback jika gagal:

```bash
cp -p "bin/api.backup-bank-soal-duplicate-workflow-${stamp}" bin/api
pm2 restart mtsn2kolut-core-api --update-env
```

### Phase 4 — Optional Web-admin Restart Only If Built

Jika pada Phase 1 hanya `npm run check`, tidak perlu restart web-admin.

Jika nanti menjalankan:

```bash
npm --prefix apps/web-admin run build
```

maka wajib langsung restart:

```bash
pm2 restart mtsn2kolut-web-admin --update-env
```

Alasan: SvelteKit adapter-node bisa stale chunk kalau build tanpa restart.

### Phase 5 — Smoke Test Production

Minimal smoke non-auth:

```bash
curl -fsS http://127.0.0.1:8080/health
curl -o /dev/null -s -w '%{http_code}\n' http://127.0.0.1:8021/login
curl -o /dev/null -s -w '%{http_code}\n' http://127.0.0.1:8021/bank-soal/daftar
```

Interpretasi:

- `/login` → 200 expected.
- `/bank-soal/daftar` tanpa login → 302 expected karena protected route.

Manual smoke sebagai admin/guru:

1. Buka Bank Soal → Daftar Soal.
2. Buat draft soal baru.
3. Klik simpan cepat/double click.
4. Pastikan hanya 1 draft muncul.
5. Submit review.
6. Submit review ulang pada soal yang sudah `submitted/review`.
7. Pastikan tidak error dan status tidak turun.

### Phase 6 — Backup Database Sebelum Cleanup

Sebelum cleanup data lama:

```bash
mkdir -p /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql
stamp=$(date +%Y%m%d-%H%M%S)
pg_dump --format=custom --file "/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-bank-soal-duplicate-cleanup-${stamp}.dump" "$DATABASE_URL"
sha256sum "/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-bank-soal-duplicate-cleanup-${stamp}.dump"
```

Catatan:

- Jangan tampilkan `DATABASE_URL`.
- Simpan path backup dan SHA256 di laporan akhir.

### Phase 7 — Audit Duplikasi Lama

Buat audit query untuk melihat kandidat duplicate draft.

Kriteria kandidat:

- `status = draft` atau `workflow_status = draft/rejected/revision_needed` sesuai schema aktual.
- belum published/approved,
- belum masuk package,
- belum punya answer/submission,
- identik fingerprint aman,
- author sama,
- subject sama,
- event sama/null sama,
- target level sama,
- question type sama,
- payload utama dan metadata sama.

Output audit minimal:

- fingerprint/group key,
- jumlah duplicate,
- daftar ID,
- oldest ID yang akan dipertahankan,
- candidate IDs yang bisa dihapus,
- status pemakaian.

Audit harus dry-run dulu:

```sql
BEGIN;
-- SELECT kandidat duplicate
ROLLBACK;
```

Tidak boleh delete pada langkah audit.

### Phase 8 — Cleanup Duplikasi Lama Secara Aman

Jika audit jelas dan user setuju, cleanup dengan transaction.

Aturan cleanup:

- keep 1 row paling awal/oldest atau row yang paling lengkap metadata-nya,
- delete hanya duplicate draft unused,
- jangan delete jika ada relasi ke package/session/answers/audit penting,
- catat jumlah rows deleted,
- commit hanya jika hasil sesuai expectation.

Pattern aman:

```sql
BEGIN;
-- temp table candidate duplicate IDs
-- SELECT count before
-- DELETE only safe duplicate draft IDs
-- SELECT count after
-- jika sesuai: COMMIT;
-- jika tidak: ROLLBACK;
```

Rekomendasi: sebelum `COMMIT`, tampilkan hasil dry-run ke user jika jumlah kandidat besar/ambigu.

### Phase 9 — Post-cleanup Verification

Setelah cleanup:

1. Query ulang duplicate groups.
2. Pastikan kandidat duplicate yang sama turun menjadi 0 atau hanya tersisa yang intentionally different.
3. Buka Bank Soal di UI.
4. Pastikan list tidak lagi menampilkan spam duplicate untuk soal contoh.
5. Health check backend.

Commands:

```bash
curl -fsS http://127.0.0.1:8080/health
pm2 status mtsn2kolut-core-api mtsn2kolut-web-admin
```

### Phase 10 — Final Report

Laporan ke user harus berisi:

- Commit hash patch review.
- Deploy status core-api.
- Health check result.
- Backup path + SHA256.
- Jumlah duplicate groups sebelum cleanup.
- Jumlah row duplicate yang dihapus.
- Jumlah duplicate groups setelah cleanup.
- Catatan route protected 302 normal.
- Rekomendasi monitoring 1–2 hari.

## Files Likely Changed

Already expected changed files:

```text
services/core-api/db/queries/cbt_questions.sql
services/core-api/internal/repository/postgres/cbt_questions.sql.go
services/core-api/internal/service/cbt_question.go
services/core-api/internal/service/cbt_question_authoring.go
services/core-api/internal/service/cbt_question_test.go
services/core-api/internal/service/cbt_question_workflow.go
.hermes/plans/2026-05-16_120850-bank-soal-draft-duplicate-guard.md
.hermes/plans/2026-05-16_130000-bank-soal-duplicate-workflow-full-finish-recommended.md
```

No migration expected.
No frontend code expected beyond already-deployed emergency frontend guard.

## Risks & Mitigations

### Risk: false-positive duplicate cleanup

Mitigation:

- fingerprint lengkap,
- dry-run,
- backup,
- delete only unused draft,
- no cleanup without clear candidate list.

### Risk: deleting a soal that is already used

Mitigation:

- join/check all usage tables before delete,
- delete only if package/answer/session usage count = 0.

### Risk: backend deploy fails

Mitigation:

- binary backup before replace,
- health loop,
- immediate rollback command ready.

### Risk: SvelteKit stale chunks

Mitigation:

- do not run web-admin build unless needed,
- if build is run, immediately restart PM2 web-admin.

## Recommended Decision

Proceed with this order:

1. Validate full.
2. Commit patch review.
3. Deploy core-api guard.
4. Smoke test.
5. Backup DB.
6. Audit duplicate old data.
7. Cleanup only safe unused duplicate draft rows.
8. Verify and report.

This is safer than directly deleting duplicates first, because it closes the source of duplication before touching existing data.
