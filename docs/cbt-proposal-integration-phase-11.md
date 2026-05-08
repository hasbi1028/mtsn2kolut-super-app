# CBT Proposal Integration Phase 11 - Evidence Archive and Retention Hardening

Status: evidence archive/retention rules after Phase 10 handoff, 2026-05-08.

Phase 11 follows CBT Phase 10 commit `64da343`. It is docs/tests/ops evidence tooling only. It records how CBT release evidence is archived, verified, retained, read back, and eventually expired after the Phase 10 handoff package is complete. It does not add product scope, deployment automation, schema work, live DB writes, or new CBT runtime routes.

## Scope

Phase 11 scope:

- define archive/retention rules after Phase 10 handoff.
- add a Phase 11 archive/retention section to `docs/cbt-release-evidence-template.md`.
- extend `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts` as the Phase 11 guard.
- keep archive review manual, auditable, and read-only.

Phase 11 is not:

- No product runtime.
- No DB schema/migrations.
- No deploy automation that restarts PM2.
- No live DB writes.
- No `/api/cbt` public routes.
- not a deploy runner.
- not a backup/restore automation job.
- not a new storage service or runtime retention daemon.

## Archive Path Convention

Store the final evidence archive under ops-controlled storage, not committed tmp. The operator chooses the exact mounted path, but it should follow this convention:

```text
<ops-controlled-storage>/cbt-evidence/<yyyy>/<yyyy-mm-dd>-<release-id>-<short-commit>/
```

Example placeholder:

```text
/srv/mtsn2kolut/evidence/cbt/2026/2026-05-08-phase-11-64da343/
```

Rules:

- `tmp/` output from `deploy/scripts/cbt-release-preflight.sh` is staging only.
- evidence archive paths must stay outside the Git working tree unless explicitly represented by a small redacted index note.
- the archive path should be readable by the authorized ops/reviewer group and not publicly served by web/admin/static routes.
- archive path convention must preserve release date, release identifier, and short commit for later lookup.

## Required Artifacts

The archive should contain or reference these required artifacts:

- `cbt-release-preflight.md`.
- `cbt-release-preflight.json`.
- `cbt-release-manifest.json`.
- `logs/git-head.log`.
- `logs/git-status.log`.
- `logs/git-diff-check.log`.
- selected optional validation logs that were actually run, such as web check, Go test/build, Flutter analyze/test, and `make ops-health`.
- completed `docs/cbt-release-evidence-template.md` copy or exported handoff note.
- Phase 10 handoff package with operator, reviewer, rollback owner, and follow-up owner.
- this Phase 11 archive/retention review note when completed.

Do not add raw `.env`, database dumps, PM2 env dumps, access tokens, exam tokens, answer keys, full request headers, or credential material to the archive.

## SHA-256 Manifest Verification

Before accepting the archive, verify `cbt-release-manifest.json`:

- manifest exists and parses as JSON.
- algorithm is `sha256`.
- every required artifact has a `sha256` value with 64 lowercase hex characters.
- every listed artifact path exists in the archive or is explicitly marked as externally referenced.
- recomputed SHA-256 for archived files matches the manifest.
- secret scan status in the manifest is `pass`.
- verification result is recorded in the Phase 11 archive section of `docs/cbt-release-evidence-template.md`.

If checksum verification fails, do not delete the evidence immediately. Mark the archive as incomplete, keep it restricted, and create a follow-up owner action.

## Retention Owner and Period

Record:

- retention owner:
- reviewer:
- retention period placeholder:
- review date:
- expiry date placeholder:
- exception/hold reason:

The retention period placeholder stays explicit until the school defines a formal CBT evidence retention policy. Do not invent a permanent retention period in the archive note. A legal, audit, or dispute hold should be recorded before any expiry/deletion action.

## Redaction/No Secrets

The archive must preserve redaction/no secrets discipline:

- no password.
- no token ujian mentah.
- no JWT/access/refresh token.
- no API key or worker key.
- no answer key.
- no database DSN.
- no live host credential.
- no unredacted `Authorization: Bearer ...` values.
- no raw `printenv`, `.env`, PM2 env dump, full request header, or database dump.

Use summaries, hashes, short commit IDs, and redacted paths. If secret-like content is found, restrict access, record the finding, rebuild a clean evidence bundle, and mark the contaminated artifact as rejected.

## Access Control

Archive access control must be documented:

- owner group:
- allowed readers:
- allowed writers:
- reviewer:
- storage location:
- public web exposure checked: ya / tidak.
- repository commit exposure checked: ya / tidak.

Only ops/reviewer roles that need CBT release evidence should read the archive. The archive must not be copied into public web roots, SvelteKit static assets, Git-tracked tmp paths, or shared chat/file locations without redaction review.

## Restore/Read-Back Check

A restore/read-back check is required before Phase 11 is accepted:

- open the archive from a fresh shell/session.
- read `cbt-release-manifest.json`.
- recompute SHA-256 for at least the markdown evidence, JSON evidence, and one log file.
- confirm the completed evidence template can be opened and interpreted without live service access.
- confirm the Phase 10 handoff package references the same commit and release identifier.
- record pass/fail and operator notes.

This is a read-only check. It does not restore databases, run migrations, restart PM2, or mutate live services.

## Deletion/Expiry Log Placeholder

Do not delete evidence during Phase 11 implementation. Add a deletion/expiry log placeholder for future authorized retention work:

| Date | Archive path | Action | Authorized by | Performed by | Reason | SHA-256 manifest checked | Notes |
|------|--------------|--------|---------------|--------------|--------|--------------------------|-------|
| | | review / extend / delete | | | | ya / tidak | |

Deletion is allowed only after the retention owner confirms the retention period, no hold exists, and the archive has been reviewed for required records.

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

Targeted validation for Phase 11:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
cd apps/web-admin && npm run check
make ops-health
cd services/core-api && go test ./...
cd services/core-api && go build -o /dev/null ./cmd/api
cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter analyze
cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter test
```

`make ops-health` is allowed as read-only live health smoke. Do not run deploy, PM2 restart, live migration, ad hoc SQL, or PM2-changing commands as part of Phase 11 implementation validation.

## Acceptance Criteria

Phase 11 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 11 dan archive/retention section belum ada.
- `docs/cbt-proposal-integration-phase-11.md` describes archive path convention under ops-controlled storage and not committed tmp.
- `docs/cbt-release-evidence-template.md` includes Phase 11 Evidence Archive and Retention.
- archive requirements include required artifacts, SHA-256 manifest verification, retention owner, retention period placeholder, redaction/no secrets, access control, restore/read-back check, and deletion/expiry log placeholder.
- no product runtime, DB schema/migration, deploy automation, PM2 restart, live DB write, secret, or `/api/cbt` public route change is introduced.
