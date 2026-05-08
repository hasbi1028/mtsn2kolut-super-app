# CBT Final Release Evidence

Status: final release evidence baseline, 2026-05-08.

Current commit baseline: `c823183` (`docs: finalize cbt release readiness phases`).

This document records the current final evidence split after CBT Phase 15. It does not perform deploy, PM2 restart, migration, live database writes, live exam mutation, or public CBT route work.

## Automated evidence completed on this host

These items can be generated or checked from the repository host without physical Android devices or live exam mutation:

- Proposal gap audit document: `docs/cbt-proposal-gap-audit.md`.
- Final release evidence document: `docs/cbt-release-final-evidence.md`.
- Final readiness script: `deploy/scripts/cbt-final-readiness.sh`.
- Generated ignored output directory: `tmp/cbt-final-readiness` by default, or a caller-provided output directory.
- Generated JSON artifacts:
  - `cbt-proposal-gap-audit.json`.
  - `cbt-final-evidence.json`.
  - `cbt-final-signoff.json`.
- Generated markdown artifact:
  - `cbt-final-readiness.md`.
- Automated checks represented in the final readiness bundle:
  - current git commit and branch.
  - `git diff --check`.
  - docs existence.
  - tests manifest.
  - safe health command status, skipped by default unless explicitly requested with `--run-ops-health`.
  - generated evidence secret scan.

## Manual evidence requiring physical Android devices and operator rehearsal

These items remain `pending_manual_evidence` or `pending_manual_signoff` unless an operator explicitly marks them complete in an evidence bundle:

- Device matrix: `apps/mobile/DEVICE_TEST_MATRIX.md`.
  - Status: `pending_manual_evidence`.
  - Requires physical Android devices.
  - Minimum two Android vendors.
  - Must include login token, heartbeat, background/resume, pending answer, submit guard, device mismatch, screenshot protection / `FLAG_SECURE`, and network disturbance evidence.
- Operator rehearsal:
  - Status: `pending_manual_evidence`.
  - Must cover Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review.
  - Must include proctor evidence, role/scope/token boundary, event/audit evidence, and go/no-go rehearsal notes.
- Final go/no-go sign-off:
  - Status: `pending_manual_signoff`.
  - Must name operator, reviewer, rollback owner, follow-up owner, blockers, accepted risks, and evidence bundle location.

Tidak ada klaim production go tanpa evidence perangkat nyata dan rehearsal operator. The readiness script may report `ready_for_rehearsal` when manual completion flags are provided, but it still records `production_go: false` because production acceptance is an operator decision outside this host.

## Final readiness command

```bash
deploy/scripts/cbt-final-readiness.sh --output tmp/cbt-final-readiness
```

Optional manual flags for an operator-created bundle:

```bash
deploy/scripts/cbt-final-readiness.sh \
  --manual-device-matrix-complete \
  --manual-operator-rehearsal-complete \
  --manual-final-signoff-complete
```

Optional read-only health status:

```bash
deploy/scripts/cbt-final-readiness.sh --run-ops-health
```

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak mengubah product runtime.
- Tidak mengubah schema database atau migration.
- Tidak melakukan live DB writes.
- Tidak mengubah route `/api/cbt`.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.
