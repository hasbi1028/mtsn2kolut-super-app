# CBT Proposal Integration Phase 31-32 — Production Candidate and Final Sign-Off

Status: completed as a deployment candidate/evidence update on 2026-05-08.

Current deployment candidate commit: `e005134 docs: complete cbt proposal phases 27 30`.

## Phase 31 — Final Production Deployment Candidate

Phase 31 moved the latest CBT proposal integration candidate to the live PM2 runtime after the user explicitly approved continuing.

### Actions performed

- Clean worktree preflight: PASS.
- Build deploy artifacts:
  - Core API: PASS.
  - Web Admin: PASS.
  - PUSAKA Worker: PASS.
- PostgreSQL backup before deploy: PASS.
- Migration reconcile/apply via `make db-migrate`: PASS; migrations `001` through `081` reported skipped/applied and completed.
- PM2 restart with `--update-env`:
  - `mtsn2kolut-core-api`: online.
  - `mtsn2kolut-web-admin`: online.
  - `mtsn2kolut-pusaka-worker`: online.
- `pm2 save`: PASS.

### Backup evidence

- Latest backup artifact: `/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pusaka_20260508_192856.dump`.
- Size: `927484 bytes`.
- SHA-256: `3f4bfaae7fc51e3d357a9a002dd5096bbb7a1e4c69bbc0433c46b72f38d77afb`.

The backup artifact may be verified read-only with `deploy/scripts/cbt-final-readiness.sh --backup-artifact <dump>`. Do not restore over the live database.

### Smoke evidence

- `bash deploy/scripts/health-check.sh all`: PASS.
- Backend `/health`: `200`.
- Web root `http://127.0.0.1:8021/`: `200`.
- Protected `/settings/rbac`: `302` unauthenticated redirect, expected.
- Protected BFF `/api/bank-soal/summary`: `401` unauthenticated, expected.
- Core API `/api/exam/status` without token: `401`, expected.
- `/bank-soal`: `302` unauthenticated redirect, expected.
- `/asesmen`: `302` unauthenticated redirect, expected.

### Boundary

- This phase did not add a public SvelteKit `/api/cbt/**` route tree.
- Flutter student runtime remains Core API `/api/exam/*`.
- PocketBase, SQLite, and Alpine.js from the proposal are not runtime architecture.
- No ad hoc SQL or live DB restore was performed.

## Phase 32 — Final Go/No-Go Sign-Off

Phase 32 finalizes the evidence framing for the 100% proposal claim, but keeps the production acceptance gate honest.

### Safe final claim

Proposal CBT MTsN 2 Kolaka Utara is implemented into the monorepo as one of these statuses:

- Implemented runtime.
- Implemented with automated/deployment evidence.
- Adapted by the approved Web Admin + Core API + PostgreSQL + Flutter architecture.
- Deferred/out-of-scope with formal rationale where proposal items are unsafe or unrealistic for BYOD CBT v1.

### Sign-off status

Automated deployment and readiness evidence is complete for the current candidate, but final operational go/no-go remains gated by manual evidence:

- Physical Android device matrix: `pending_manual_evidence` until tested on at least two real Android vendors.
- Operator/proctor rehearsal: `pending_manual_evidence` until the live workflow is rehearsed and recorded.
- Final operator sign-off: `pending_manual_signoff` until Bapak/operator explicitly approves final production acceptance.
- `production_go`: must remain `false` until those manual gates are satisfied.

### Required final sign-off wording

Use this wording only after manual evidence is complete:

> Proposal CBT MTsN 2 Kolaka Utara telah diimplementasikan 100% ke monorepo sebagai fitur runtime, evidence operasional, atau adaptasi resmi sesuai arsitektur Web Admin + Core API + PostgreSQL + Flutter. Item yang tidak diterapkan mentah-mentah seperti PocketBase/SQLite/Alpine, kiosk penuh BYOD, screen preview, desktop agent, dan ISO certification dicatat sebagai adapted/out-of-scope dengan rationale dan evidence.

Until manual evidence is complete, use:

> Source code, deployment candidate, and automated readiness evidence are complete. Final production acceptance is pending physical Android device matrix, operator rehearsal, and explicit go/no-go sign-off.

## Validation recipe

```bash
git diff --check
bash -n deploy/scripts/cbt-final-readiness.sh
bash deploy/scripts/cbt-final-readiness.sh --output tmp/cbt-final-readiness-phase31-32 --run-ops-health --backup-artifact /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/latest.dump
python3 -m json.tool tmp/cbt-final-readiness-phase31-32/cbt-final-evidence.json >/dev/null
python3 -m json.tool tmp/cbt-final-readiness-phase31-32/cbt-final-signoff.json >/dev/null
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts && npm run check
```

Full validation should also include Core API Go tests/build, full Web unit tests, Flutter analyze/test, and `make ops-health`.
