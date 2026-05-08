# CBT Proposal Integration Phase 15 - Final CBT Release Readiness Sign-off

Status: Final CBT Release Readiness Sign-off, 2026-05-08.

Phase 15 closes the CBT proposal integration docs/tests/ops readiness sequence. It records the final Phase 0-15 sign-off gate, evidence bundle requirements, validation commands, rollback owner, and final baseline marker. It does not add product runtime, deployment automation, schema work, live DB writes, or new CBT runtime routes.

## Scope

Phase 15 scope:

- add this Final CBT Release Readiness Sign-off note.
- harden `docs/cbt-release-evidence-template.md` with final sign-off summary Phase 0-15.
- require an explicit go/no-go decision, rollback owner, evidence bundle reference, validation commands, and final baseline marker.
- keep the work limited to docs/tests/ops readiness only.

Phase 15 is not:

- No product runtime.
- No DB schema/migrations.
- No deploy automation that restarts PM2.
- No live DB writes.
- No `/api/cbt` public routes.
- not a release deploy command.
- not a database backup/restore action.
- not a public student API change.

## Final Sign-Off Summary Phase 0-15

The final evidence template must record a compact Phase 0-15 summary:

- Phase 0 architecture alignment and guard.
- Phase 1 exam API contract stabilization.
- Phase 2 Web Admin workflow alignment.
- Phase 3 backend runtime hardening.
- Phase 4 Flutter BYOD anti-cheat and resilience.
- Phase 5 rehearsal, rollout, and post-exam review.
- Phase 6 production readiness evidence automation.
- Phase 7 release preflight verifier/test hardening.
- Phase 8 manifest/checksum/secret-scan hardening.
- Phase 9 live deploy smoke/runbook hardening.
- Phase 10 post-deploy evidence handoff.
- Phase 11 evidence archive and retention.
- Phase 12 evidence index and retrieval.
- Phase 13 mobile release candidate and device matrix.
- Phase 14 operator rehearsal and proctor evidence.
- Phase 15 final CBT release readiness sign-off.

## Go/No-Go Decision

Record the final go/no-go decision before release acceptance:

- go/no-go:
- decision date/time WITA:
- operator:
- reviewer:
- rollback owner:
- follow-up owner:
- blockers:
- accepted risks:
- next checkpoint:

No-go if evidence bundle is incomplete, validation commands fail without accepted owner sign-off, role/scope/token boundary is broken, mobile device matrix is incomplete, or secrets appear in evidence.

## Evidence Bundle

The final evidence bundle should reference:

- completed `docs/cbt-release-evidence-template.md`.
- `cbt-release-preflight.md`.
- `cbt-release-preflight.json`.
- `cbt-release-manifest.json`.
- optional validation logs actually run.
- Phase 13 device matrix evidence.
- Phase 14 proctor evidence.
- archive path and manifest SHA-256.
- redacted index row or lookup reference.

The bundle must stay outside public web roots and should not include raw secrets, full logs, `.env`, PM2 env dump, DB dumps, raw exam tokens, answer keys, JWTs, or API keys.

## Validation Commands

Run or record status for these validation commands:

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

These validation commands are verification only. `make ops-health` is read-only health smoke. Do not use this phase to deploy, restart PM2, run migrations, write SQL, or create public `/api/cbt` routes.

## Final Baseline Marker

Use this explicit final baseline marker in the completed evidence:

```text
CBT Phase 15 final baseline
```

The final baseline marker means the docs/tests/ops readiness sequence Phase 0-15 has a reviewed evidence bundle and go/no-go decision. It does not mean services were deployed automatically.

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

## Acceptance Criteria

Phase 15 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 15 belum ada.
- `docs/cbt-proposal-integration-phase-15.md` documents Final CBT Release Readiness Sign-off.
- `docs/cbt-release-evidence-template.md` includes Phase 15 Final CBT Release Readiness Sign-off with final sign-off summary Phase 0-15, go/no-go, rollback owner, evidence bundle, validation commands, and final baseline marker.
- the final baseline marker is exactly `CBT Phase 15 final baseline`.
- no product runtime, DB schema/migration, deploy automation, PM2 restart, live DB write, secret, or `/api/cbt` public route change is introduced.
