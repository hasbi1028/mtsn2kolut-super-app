# CBT Proposal Integration Phase 7 - Release Preflight Verifier and Test Hardening

Status: verifier/test hardening, 2026-05-08.

Phase 7 bukan product roadmap baru, bukan deploy, dan bukan perubahan runtime CBT. Fase ini hanya memperkuat automated regression coverage untuk verifier read-only `deploy/scripts/cbt-release-preflight.sh` yang dibuat pada Phase 6.

## Scope

Phase 7 scope:

- menguji `deploy/scripts/cbt-release-preflight.sh` dengan eksekusi nyata dari Vitest.
- menulis output test ke temporary ignored output directory under `tmp/`.
- memastikan markdown evidence, JSON evidence, dan log evidence dibuat.
- memastikan JSON evidence remains parseable.
- memastikan optional checks stay skipped by default.
- memastikan verifier refuses non-empty output directories.
- memastikan verifier refuses unknown flags.
- memperluas guard di `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts`.

Phase 7 bukan:

- deploy.
- PM2 restart.
- live migration.
- ad hoc SQL.
- perubahan skema database.
- route tree publik baru.
- perubahan kontrak runtime mobile.
- roadmap produk baru setelah Phase 0 sampai Phase 5.

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak menjalankan `psql` untuk `ALTER`, `UPDATE`, `DELETE`, `INSERT`, atau DDL/DML lain.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `/api/cbt/*` tetap hanya compatibility/deprecated route yang sudah ada untuk web-admin lama dan transisi backend; bukan runtime siswa baru.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.
- `apps/mobile` tetap Flutter BYOD client dengan batasan BYOD realistis.

## Regression Coverage

Automated guard Phase 7 berada di `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts`.

Coverage wajib:

- menjalankan verifier ke output directory sementara yang berada di `tmp/`.
- membaca `cbt-release-preflight.md`.
- membaca dan parse `cbt-release-preflight.json`.
- mengecek `logs/git-head.log`, `logs/git-status.log`, dan `logs/git-diff-check.log`.
- mengecek check opsional default sebagai `skipped`, termasuk web docs guard, web check, Go test/build, dan Flutter checks.
- mengecek output directory non-empty ditolak sebelum markdown/json baru dibuat.
- mengecek unknown flag ditolak dengan usage error.
- menjaga no-deploy/no-migration/no-public-`/api/cbt/**` boundary tetap tertulis.

## Validasi

Validasi targeted Phase 7:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
```

Jalankan `npm run check` bila ada perubahan Svelte atau TypeScript surface yang perlu dicek lint/type/a11y. Phase 7 ini hanya dokumentasi, shell verifier wording, dan Vitest regression coverage.

## Acceptance Criteria

Phase 7 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen/ekspektasi Phase 7 belum ada.
- verifier tetap read-only dan hanya menulis markdown/json/log evidence ke output directory.
- optional checks tetap eksplisit dan tidak berjalan tanpa flag.
- output JSON dapat diparse oleh test.
- non-empty output directory dan unknown flag gagal dengan exit code usage.
- tidak ada deploy, PM2 restart, `make db-migrate`, migrasi live, ad hoc SQL, atau public `/api/cbt/**` route tree baru.
