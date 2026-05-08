# CBT Proposal Integration Phase 9 - Live Deploy Smoke and Runbook Hardening

Status: post-deploy smoke/runbook hardening only, 2026-05-08.

Phase 9 follows CBT Phase 8 commit `231d908` and the live deploy. This phase does not add product scope. It records the observed live deploy issue and locks the operational fix: `make ops-health` failed because `deploy/scripts/health-check.sh` was not executable, while `bash deploy/scripts/health-check.sh all` passed.

## Scope

Phase 9 scope is docs/tests/ops script/Makefile only:

- document post-deploy smoke expectations for CBT release operations.
- document the Makefile health fix.
- update `ops-health` Makefile targets so they invoke `deploy/scripts/health-check.sh` through `bash`.
- ensure the health check does not require executable bit on `deploy/scripts/health-check.sh`.
- extend `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts` as the Phase 9 guard.

Phase 9 is not:

- product runtime work.
- database schema work.
- migration work.
- `/api/cbt` route work.
- deployment automation.
- PM2 restart work.
- a change to `deploy/scripts/cbt-release-preflight.sh`.

## Observed Deploy Issue

Observed issue after live deploy:

- `make ops-health` failed because `deploy/scripts/health-check.sh` was not executable.
- `bash deploy/scripts/health-check.sh all` passed against the deployed services.
- The failure was a Makefile invocation assumption, not a backend, frontend, worker, database, or CBT route regression.

Makefile health fix:

- `ops-health` invokes `bash deploy/scripts/health-check.sh all`.
- `ops-health-backend` invokes `bash deploy/scripts/health-check.sh backend`.
- `ops-health-frontend` invokes `bash deploy/scripts/health-check.sh frontend`.
- `ops-health-worker` invokes `bash deploy/scripts/health-check.sh worker`.
- ops-health targets invoke `deploy/scripts/health-check.sh` through `bash`.
- The Makefile path does not require executable bit on `deploy/scripts/health-check.sh`.

## Post-Deploy Smoke

Post-deploy smoke is an operator check after backend, frontend, and worker have already been deployed through the documented manual process. The smoke does not deploy, restart, migrate, or mutate database state.

Recommended post-deploy smoke:

```bash
make ops-health
make ops-health-backend
make ops-health-frontend
make ops-health-worker
bash deploy/scripts/health-check.sh bank-soal
```

If `make` is unavailable during incident handling, the equivalent direct shell checks are:

```bash
bash deploy/scripts/health-check.sh all
bash deploy/scripts/health-check.sh backend
bash deploy/scripts/health-check.sh frontend
bash deploy/scripts/health-check.sh worker
```

CBT release smoke should then follow `docs/cbt-smoke-checklist.md` for operator-facing Bank Soal, Asesmen, Flutter BYOD, and hasil checks. The smoke must keep Flutter connected directly to `services/core-api` through `/api/exam/*`; it must not introduce runtime siswa traffic through `/api/cbt/login` or `/api/cbt/status`.

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak menjalankan `psql` untuk `ALTER`, `UPDATE`, `DELETE`, `INSERT`, atau DDL/DML lain.
- Tidak mengubah product runtime.
- Tidak mengubah schema database atau migration.
- Tidak mengubah route `/api/cbt`.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.
- `deploy/scripts/cbt-release-preflight.sh` boundary remains intact; Phase 9 only references it as the existing read-only release-evidence verifier.
- No secrets, tokens, passwords, API keys, answer keys, or live host credentials are recorded in this document or tests.

## Validation

Targeted validation for Phase 9:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
make -n ops-health
make -n ops-health-backend
make -n ops-health-frontend
make -n ops-health-worker
```

Do not run deploy, PM2 restart, live migration, ad hoc SQL, or PM2-changing commands as part of Phase 9 implementation validation.

## Acceptance Criteria

Phase 9 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 9 belum ada dan Makefile masih memakai direct script execution.
- Makefile `ops-health*` recipes call `bash deploy/scripts/health-check.sh ...`.
- `ops-health*` recipes no longer depend on executable bit for `deploy/scripts/health-check.sh`.
- post-deploy smoke commands are documented.
- `deploy/scripts/cbt-release-preflight.sh` remains unchanged and read-only in boundary.
- no product runtime, DB schema, migration, deploy automation, PM2 restart, secret, or `/api/cbt` route change is introduced.
