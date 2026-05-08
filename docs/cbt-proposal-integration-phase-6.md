# CBT Proposal Integration Phase 6 - Production Readiness Evidence Automation

Status: operational release-evidence automation/guard, 2026-05-08.

Phase 6 bukan roadmap produk baru. roadmap resmi tetap Phase 0 sampai Phase 5; fase ini hanya menambahkan bukti kesiapan produksi yang repeatable untuk CBT setelah commit Phase 5 `bec16c8`. Scope-nya adalah dokumentasi, template evidence, dan preflight read-only. Tidak ada deploy, tidak ada restart, tidak ada migration execution, dan tidak ada perubahan arsitektur runtime.

## Release Evidence Bundle

Evidence bundle adalah folder hasil preflight yang bisa dilampirkan ke keputusan rehearsal/rollout CBT. Bundle minimal berisi:

- `cbt-release-preflight.md`: ringkasan operator-friendly untuk commit, status repo, tool availability, check yang dijalankan, check yang dilewati, dan lokasi log.
- `cbt-release-preflight.json`: versi structured evidence untuk arsip atau audit internal.
- `logs/*.log`: output command yang sudah direduksi agar tidak sengaja membawa secret/token/password mentah.
- catatan manual berbasis `docs/cbt-release-evidence-template.md`.

Default output script berada di `tmp/cbt-release-preflight-<timestamp>/`, karena `tmp/` tidak ikut commit. Operator boleh memakai path eksplisit dengan `--output <dir>` selama path tersebut hanya dipakai untuk evidence, bukan sebagai lokasi deploy.

## Preflight Commands

Command dasar yang aman untuk Phase 6:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
deploy/scripts/cbt-release-preflight.sh
```

Command opsional untuk bundle lebih lengkap:

```bash
deploy/scripts/cbt-release-preflight.sh --run-web-docs-guard --output tmp/cbt-release-preflight-phase-6
deploy/scripts/cbt-release-preflight.sh --run-web-docs-guard --run-web-check
deploy/scripts/cbt-release-preflight.sh --run-go-test --run-go-build
deploy/scripts/cbt-release-preflight.sh --run-flutter-doctor --run-flutter-analyze --run-flutter-test
```

Flutter SDK path yang dipakai sebagai default evidence automation adalah:

```text
/home/servermtsn2kolut/development/flutter/bin
```

Bila Flutter SDK tidak tersedia di host preflight, hasil mobile dicatat sebagai skipped/blocker sesuai template. Jangan mengganti kondisi itu dengan klaim lulus manual tanpa menjalankan `flutter analyze` dan `flutter test` di host yang memiliki SDK.

## No-Deploy / No-Migration Boundary

Phase 6 wajib tetap berada di boundary berikut:

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak menjalankan `psql` untuk `ALTER`, `UPDATE`, `DELETE`, `INSERT`, atau DDL/DML lain.
- Tidak menjalankan script backup/health/deploy yang mengubah state operasional.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`, bukan melalui SvelteKit BFF.
- `/api/cbt/*` tetap hanya compatibility/deprecated route yang sudah ada untuk web-admin lama dan transisi backend; bukan runtime siswa baru.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.
- `apps/web-admin` tetap UI/BFF; tidak membaca PostgreSQL langsung.
- `apps/mobile` tetap Flutter BYOD client dengan batasan BYOD realistis, bukan kiosk/device-owner guarantee.

## Script Safety Rules

`deploy/scripts/cbt-release-preflight.sh` harus:

- mengumpulkan commit/status git dan ketersediaan tool tanpa membaca environment secret.
- menjalankan check berat hanya lewat flag eksplisit.
- memakai Flutter dari `/home/servermtsn2kolut/development/flutter/bin` bila tersedia, lalu fallback ke `PATH` hanya untuk evidence availability.
- menulis markdown/json/log hanya ke `tmp/` default atau path `--output`.
- meredaksi output yang menyerupai token/password/secret/API key.
- gagal non-zero bila command yang diminta operator gagal atau tool wajib untuk command yang diminta tidak tersedia.
- memperlakukan skipped check sebagai evidence, bukan sebagai bukti lulus.

## Acceptance Criteria

Phase 6 diterima bila:

- `docs/cbt-proposal-integration-phase-6.md` menjelaskan release evidence bundle, preflight commands, no-deploy/no-migration boundary, Flutter SDK path, dan acceptance criteria.
- `docs/cbt-release-evidence-template.md` tersedia sebagai template evidence ops.
- `deploy/scripts/cbt-release-preflight.sh` tersedia, read-only terhadap runtime, dan menulis `cbt-release-preflight.md` serta `cbt-release-preflight.json`.
- Guard test mengunci bahwa Phase 6 tidak berubah menjadi arsitektur baru, deploy automation, PM2 restart, live migration, atau public `/api/cbt/**` route tree baru.
- Runtime siswa tetap dikonfirmasi sebagai `/api/exam/*`; tidak ada kontrak login/status mobile di namespace `/api/cbt`.
- Validasi targeted lulus: `git diff --check` dan `cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts`.
- `npm run check` hanya wajib bila ada perubahan Svelte; Phase 6 ini tidak memerlukan perubahan Svelte.
- `flutter analyze` dan `flutter test` dijalankan hanya bila membuat evidence mobile final atau bila ada perubahan Flutter; jika SDK tidak tersedia, kekurangan itu dicatat sebagai blocker evidence mobile, bukan diabaikan.

## Operator Handoff

Gunakan Phase 6 sebagai bukti rilis, bukan tombol rilis. Setelah bundle dibuat, isi `docs/cbt-release-evidence-template.md` di luar commit atau salin sebagai artefak ops sesuai kebutuhan panitia. Keputusan lanjut/tunda tetap mengikuti Phase 5 rehearsal evidence di `docs/cbt-smoke-checklist.md` dan hasil perangkat nyata pada `apps/mobile/DEVICE_TEST_MATRIX.md`.
