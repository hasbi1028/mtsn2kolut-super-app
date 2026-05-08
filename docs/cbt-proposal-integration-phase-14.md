# CBT Proposal Integration Phase 14 - Operator Rehearsal and Proctor Evidence

Status: operator rehearsal and proctor evidence hardening only, 2026-05-08.

Phase 14 follows Phase 13 mobile release-candidate hardening. It is docs/tests/ops evidence readiness for the final operator rehearsal path and proctor evidence capture. It does not add product runtime, deployment automation, schema work, live DB writes, or new CBT runtime routes.

## Scope

Phase 14 scope:

- add this Phase 14 rehearsal and proctor evidence note.
- harden `docs/cbt-smoke-checklist.md` with the final operator flow.
- harden `docs/cbt-release-evidence-template.md` with proctor evidence and go/no-go rehearsal fields.
- keep the work as operator rehearsal and proctor evidence hardening only.

Phase 14 is not:

- No product runtime.
- No DB schema/migrations.
- No deploy automation that restarts PM2.
- No live DB writes.
- No `/api/cbt` public routes.
- not a deploy runbook that restarts services.
- not a new proctoring service or storage service.
- not a live exam mutation.

## Operator Flow

The final rehearsal flow is:

Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review.

Evidence should show the operator can move through:

- Bank Soal authoring/import/review using canonical `/bank-soal/*` routes.
- Asesmen Persiapan setup using `/asesmen/persiapan` and canonical paket/kegiatan/sesi routes.
- Pelaksanaan/Pengawasan monitoring using `/asesmen/pelaksanaan` and `/asesmen/pengawasan`.
- Flutter APK login, heartbeat, answer save, warning, restore, and submit using `/api/exam/*`.
- Hasil/Post-exam review using `/asesmen/hasil`, audit/event review, and follow-up classification.

## Proctor Evidence

Proctor evidence should be concrete enough for go/no-go rehearsal:

- session, room, participant, and device identifiers that are safe to record.
- heartbeat and connection status seen by pengawas.
- BYOD warning/event evidence for background/resume, network disturbance, pending answer, submit guard, and device mismatch.
- screenshot protection / `FLAG_SECURE` check recorded as deterrence evidence.
- timestamp, operator, pengawas, and reviewer.
- issue severity: blocker ujian / catatan operasional / backlog produk.

Do not paste raw exam tokens, full JWTs, answer keys, passwords, API keys, or full request headers into proctor evidence.

## Role/Scope/Token Boundary

The rehearsal must explicitly check role/scope/token boundary:

- admin/panitia may perform approved setup and operational token print where allowed.
- guru sees only authorized subject, class, package, scoring, and result scope.
- pengawas/proktor sees only assigned session/room/peserta operational state.
- token visibility is restricted to authorized operational contexts.
- question keys and rubrics do not leak to unauthorized roles.
- Flutter uses token-scoped `/api/exam/*`; no student runtime through `/api/cbt/login` or `/api/cbt/status`.

## Event/Audit Evidence

Capture event/audit evidence for:

- Bank Soal changes that feed the tested package.
- Asesmen Persiapan session, room, participant, token, and seat changes.
- Pelaksanaan/Pengawasan warning and heartbeat observations.
- Flutter APK login, heartbeat, answer, event, and submit behavior.
- Hasil/Post-exam scoring/export/review activity.

Event/audit evidence must be redacted when copied into release notes. Keep raw logs in approved evidence storage only.

## Go/No-Go Rehearsal

The go/no-go rehearsal decision should be explicit:

- go/no-go:
- operator:
- pengawas:
- reviewer:
- rollback owner:
- blockers:
- accepted operational notes:
- follow-up owner:

No-go if role/scope/token boundary fails, answer save/submit cannot be trusted, heartbeat/proctor evidence is missing, APK device matrix fails minimum two Android vendors, or any evidence contains secrets.

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

Targeted validation for Phase 14:

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

`make ops-health` is allowed as read-only live health smoke. Do not run deploy, PM2 restart, live migration, ad hoc SQL, PM2-changing commands, or public `/api/cbt` route work as part of Phase 14 implementation validation.

## Acceptance Criteria

Phase 14 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 14 belum ada.
- `docs/cbt-proposal-integration-phase-14.md` documents the operator rehearsal and proctor evidence hardening only scope.
- `docs/cbt-smoke-checklist.md` and `docs/cbt-release-evidence-template.md` include Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review.
- proctor evidence, role/scope/token boundary, event/audit evidence, and go/no-go rehearsal are documented.
- no product runtime, DB schema/migration, deploy automation, PM2 restart, live DB write, secret, or `/api/cbt` public route change is introduced.
