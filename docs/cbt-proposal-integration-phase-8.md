# CBT Proposal Integration Phase 8 - Release Evidence Manifest and Secret-Scan Hardening

Status: manifest/checksum/secret-scan hardening only, 2026-05-08.

Phase 8 bukan product roadmap baru, bukan deploy, bukan runtime CBT, dan bukan perubahan kontrak siswa. Fase ini hanya memperkuat verifier read-only `deploy/scripts/cbt-release-preflight.sh` supaya evidence bundle lebih mudah diaudit setelah Phase 7 commit `5f9cc67`.

## Scope

Phase 8 scope:

- menulis `cbt-release-manifest.json` di output directory preflight.
- mencatat SHA-256 untuk markdown/json/log evidence yang dihasilkan.
- mencatat daftar artifact yang diikutkan dalam bundle evidence.
- menjalankan secret scan ringan terhadap markdown/json/log evidence yang dihasilkan.
- menggagalkan preflight bila evidence masih memuat pola token, password, secret, API key, atau answer key yang belum direduksi.
- memperluas guard di `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts`.

Phase 8 bukan:

- deploy.
- PM2 restart.
- live migration.
- ad hoc SQL.
- perubahan skema database.
- route tree publik baru.
- perubahan kontrak runtime mobile.
- product/runtime/deploy work.

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

## Manifest dan Secret Scan

`cbt-release-manifest.json` adalah machine-readable evidence index, bukan artifact deploy. Isinya wajib mencakup:

- path artifact relatif terhadap output directory.
- jenis artifact (`markdown`, `json`, atau `log`).
- ukuran file dalam byte.
- checksum SHA-256.
- status secret scan dan daftar file yang discan.

Secret scan Phase 8 sengaja ringan. Tujuannya adalah guard terakhir atas output yang baru dihasilkan, bukan pengganti pemeriksaan manual server log. Pola yang wajib ditolak meliputi unredacted bearer authorization, token assignment, password assignment, secret assignment, API key assignment, dan answer key assignment.

## Validasi

Validasi targeted Phase 8:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
deploy/scripts/cbt-release-preflight.sh --output tmp/cbt-release-preflight-phase-8
```

Jalankan `npm run check` bila ada perubahan Svelte atau TypeScript surface yang perlu dicek lint/type/a11y. Phase 8 ini hanya dokumentasi, shell verifier evidence, manifest/checksum, dan secret-scan regression coverage.

## Acceptance Criteria

Phase 8 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen/ekspektasi Phase 8 belum ada.
- verifier tetap read-only dan hanya menulis markdown/json/log/manifest evidence ke output directory.
- `cbt-release-manifest.json` dibuat dan dapat diparse.
- manifest memuat SHA-256 untuk `cbt-release-preflight.md`, `cbt-release-preflight.json`, dan `logs/*.log`.
- secret scan pass pada output normal.
- synthetic output berisi secret-like text gagal dengan exit non-zero.
- tidak ada deploy, PM2 restart, `make db-migrate`, migrasi live, ad hoc SQL, atau public `/api/cbt/**` route tree baru.
