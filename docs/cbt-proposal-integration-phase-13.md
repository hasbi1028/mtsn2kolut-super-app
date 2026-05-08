# CBT Proposal Integration Phase 13 - Mobile Release Candidate and Device Matrix

Status: Mobile Release Candidate and Device Matrix, 2026-05-08.

Phase 13 follows CBT Phase 12 commit `a3e99bd`. It is docs/tests/ops readiness only for the final Flutter APK release candidate record and real-device BYOD matrix. It does not add product runtime, deployment automation, schema work, live DB writes, or new CBT runtime routes.

## Scope

Phase 13 scope:

- add this Phase 13 release-candidate note.
- harden `apps/mobile/DEVICE_TEST_MATRIX.md` for RC evidence.
- harden `apps/mobile/RELEASE_CHECKLIST.md` for RC identifier, APK hash/signing mode, `API_BASE_URL`, and Flutter SDK path.
- harden `docs/cbt-release-evidence-template.md` so the mobile evidence bundle can be reviewed without rerunning deploy steps.
- keep the work limited to docs/tests/ops readiness only.

Phase 13 is not:

- No product runtime.
- No DB schema/migrations.
- No deploy automation that restarts PM2.
- No live DB writes.
- No `/api/cbt` public routes.
- not a new APK distribution channel.
- not a device-owner or kiosk guarantee.
- not a backend release, migration, or live exam state change.

## Release Candidate Evidence

Record one mobile release candidate before APK field testing:

- RC identifier:
- APK SHA-256 hash:
- APK path or archive reference:
- signing mode: release keystore / debug signing for internal technical test only.
- `API_BASE_URL`:
- Flutter SDK absolute path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- Flutter validation commands:
  - `/home/servermtsn2kolut/development/flutter/bin/flutter analyze`
  - `/home/servermtsn2kolut/development/flutter/bin/flutter test`
- backend/web commit used by the RC:
- operator:
- reviewer:

The APK hash should be computed from the exact APK installed on test devices. Do not paste keystore passwords, token ujian mentah, API keys, JWTs, or other secrets into evidence.

## Device Matrix Minimum

The RC device matrix must cover minimum two Android vendors on real devices. Record the result in `apps/mobile/DEVICE_TEST_MATRIX.md` before the APK is accepted for rehearsal.

Required matrix fields:

- RC identifier.
- APK SHA-256 hash.
- signing mode.
- `API_BASE_URL`.
- vendor and model.
- Android version.
- connection type.
- tester/operator.
- result.
- notes and blocker status.

The matrix is operational evidence, not a claim that every BYOD device behaves like a managed kiosk.

## BYOD Anti-Cheat and Resilience Scenarios

Each release candidate must record these scenarios on real devices:

- background/resume.
- heartbeat.
- pending answer.
- submit guard.
- device mismatch.
- screenshot protection / `FLAG_SECURE`.
- network disturbance.

Expected evidence:

- background/resume reaches a controlled status refresh or restore gate before the student continues.
- heartbeat updates last-contact state or produces a readable warning when degraded.
- pending answer survives temporary network loss and is not silently discarded.
- submit guard blocks final submit when pending answer sync or degraded mode makes the session unsafe.
- device mismatch returns controlled `409` guidance and preserves local pending state where applicable.
- screenshot protection / `FLAG_SECURE` is verified as deterrence/evidence, not kiosk proof.
- network disturbance produces `Lokal`, `Gangguan`, `Waspada`, or `Menurun` state that pengawas can interpret.

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak menjalankan `psql` untuk `ALTER`, `UPDATE`, `DELETE`, `INSERT`, atau DDL/DML lain.
- Tidak mengubah product runtime.
- Tidak mengubah schema database atau migration.
- Tidak melakukan live DB writes.
- Tidak mengubah route `/api/cbt`.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.

## Validation

Targeted validation for Phase 13:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
cd apps/web-admin && npm run check
cd apps/web-admin && npm run test:unit
make ops-health
cd services/core-api && go test ./...
cd services/core-api && go build -o /dev/null ./cmd/api
cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter analyze
cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter test
```

`make ops-health` is allowed as read-only live health smoke. Do not run deploy, PM2 restart, live migration, ad hoc SQL, PM2-changing commands, or public `/api/cbt` route work as part of Phase 13 implementation validation.

## Acceptance Criteria

Phase 13 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 13 belum ada.
- `docs/cbt-proposal-integration-phase-13.md` documents Mobile Release Candidate and Device Matrix readiness.
- `apps/mobile/DEVICE_TEST_MATRIX.md`, `apps/mobile/RELEASE_CHECKLIST.md`, and `docs/cbt-release-evidence-template.md` include RC identifier, APK SHA-256 hash, signing mode, `API_BASE_URL`, minimum two Android vendors, BYOD scenarios, and `/home/servermtsn2kolut/development/flutter/bin`.
- BYOD scenarios include background/resume, heartbeat, pending answer, submit guard, device mismatch, screenshot protection / `FLAG_SECURE`, and network disturbance.
- no product runtime, DB schema/migration, deploy automation, PM2 restart, live DB write, secret, or `/api/cbt` public route change is introduced.
