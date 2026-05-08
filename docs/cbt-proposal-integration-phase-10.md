# CBT Proposal Integration Phase 10 - Post-Deploy Evidence Handoff Hardening

Status: post-deploy evidence handoff hardening only, 2026-05-08.

Phase 10 follows CBT Phase 9 commit `bf9df69`. It is a docs/tests/ops evidence tooling only phase for the handoff package after successful deploy/smoke. It does not add product scope, deployment automation, schema work, or new CBT runtime routes.

## Scope

Phase 10 scope:

- define the post-deploy handoff package after successful deploy/smoke.
- add the handoff checklist to `docs/cbt-release-evidence-template.md`.
- extend `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts` as the Phase 10 guard.
- keep the evidence handoff manually reviewable by operator, rollback owner, follow-up owner, and reviewer.

Phase 10 is not:

- No product runtime.
- No DB schema/migrations.
- No deploy automation that restarts PM2.
- No live DB writes.
- No `/api/cbt` public routes.
- not a new public student runtime contract.
- not a replacement for the existing manual deploy process in `deploy/DEPLOY.md`.

## Handoff Package

After backend, frontend, and worker have already been deployed through the documented manual process, and after post-deploy smoke has passed, archive a handoff package with these fields:

- commit hash:
- branch:
- evidence date/time:
- operator:
- reviewer:
- rollback owner:
- follow-up owner:
- release decision: lanjut / tunda / rollback.

The handoff package should reference the generated evidence bundle rather than paste secrets or raw environment output.

## PM2 Status Summary Placeholders

Record PM2 status summary placeholders manually from the deployed VPS targets. Do not run PM2-changing commands as evidence collection.

| Target | Process name | Status | Restarts | Uptime | Notes |
|--------|--------------|--------|----------|--------|-------|
| Backend VPS | | | | | |
| Frontend VPS | | | | | |
| Worker VPS | | | | | |

The table is a summary for review. It is not an instruction to restart, reload, stop, delete, or mutate any PM2 process.

## Health Smoke Outputs

Attach or reference health smoke outputs after deploy:

- `make ops-health`:
- `make ops-health-backend`:
- `make ops-health-frontend`:
- `make ops-health-worker`:
- `bash deploy/scripts/health-check.sh bank-soal`:

Health smoke outputs should show pass/fail, timestamp, target host context, and any non-secret operator notes. The smoke remains read-only and must not deploy, restart PM2, run migration, write SQL, or change runtime state.

## Manifest/Checksum Evidence

Attach the release evidence bundle produced before handoff:

- `cbt-release-preflight.md`:
- `cbt-release-preflight.json`:
- `cbt-release-manifest.json`:
- algorithm: `sha256`.
- manifest/checksum evidence verified: ya / tidak.
- secret scan status: pass / fail.

The manifest/checksum evidence should be reviewed before the package is accepted. If the manifest is missing, not parseable, or contains checksum gaps, the handoff remains incomplete.

## Secret Hygiene

The handoff package must not contain:

- password.
- token ujian mentah.
- JWT/access/refresh token.
- API key or worker key.
- answer key.
- database DSN.
- live host credential.
- unredacted `Authorization: Bearer ...` values.

Use summaries and paths to evidence artifacts. Do not paste raw `printenv`, `.env`, PM2 env dumps, database dumps, or full request headers.

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

Targeted validation for Phase 10:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
cd apps/web-admin && npm run check
make -n ops-health
make ops-health
cd services/core-api && go test ./...
cd services/core-api && go build -o /dev/null ./cmd/api
cd apps/mobile && flutter analyze
cd apps/mobile && flutter test
```

`make ops-health` is allowed as read-only live health smoke. Do not run deploy, PM2 restart, live migration, ad hoc SQL, or PM2-changing commands as part of Phase 10 implementation validation.

## Acceptance Criteria

Phase 10 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 10 dan handoff section belum ada.
- `docs/cbt-proposal-integration-phase-10.md` describes the post-deploy handoff package.
- `docs/cbt-release-evidence-template.md` includes the Phase 9/10 post-deploy handoff section.
- handoff package captures commit hash, PM2 status summary placeholders, health smoke outputs, manifest/checksum evidence, rollback owner, follow-up owner, and secret hygiene.
- no product runtime, DB schema/migration, deploy automation, PM2 restart, live DB write, secret, or `/api/cbt` public route change is introduced.
