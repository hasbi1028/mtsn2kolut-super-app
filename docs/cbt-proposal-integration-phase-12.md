# CBT Proposal Integration Phase 12 - Evidence Index and Retrieval Hardening

Status: redacted evidence index/search and retrieval policy after Phase 11 archive/retention, 2026-05-08.

Phase 12 follows CBT Phase 11 commit `0d9391f`. It is docs/tests/ops evidence policy only. It defines how authorized operators find, verify, request, and read back archived CBT release evidence without exposing secrets, full logs, environment dumps, or public runtime routes. It does not add product scope, deployment automation, schema work, live DB writes, or new CBT runtime routes.

## Scope

Phase 12 scope:

- define a redacted evidence index/search and retrieval policy after Phase 11 archive/retention.
- add a Phase 12 evidence index/retrieval section to `docs/cbt-release-evidence-template.md`.
- extend `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts` as the Phase 12 guard.
- keep lookup, retrieval, and read-back manual, auditable, and read-only.

Phase 12 is not:

- No product runtime.
- No DB schema/migrations.
- No deploy automation that restarts PM2.
- No live DB writes.
- No `/api/cbt` public routes.
- not a search service.
- not a document database.
- not a public evidence portal.
- not an automation that restores archives into live systems.

## Evidence Index Policy

The evidence index is a redacted operator lookup ledger. It may be a controlled ops spreadsheet, private document, or access-controlled text/CSV/JSON file managed outside runtime services. It is not product data and must not become committed tmp data except template docs.

Required index fields:

| Field | Policy |
|-------|--------|
| release id | Human release identifier used in the Phase 10/11 handoff and archive path. |
| commit | Full or short Git commit for the archived evidence. |
| date | Evidence date or archive date in `yyyy-mm-dd` form. |
| archive path | Redacted or ops-local path to the archive, never a public URL. |
| manifest sha256 | SHA-256 of the archived `cbt-release-manifest.json`. |
| owner | Retention or archive owner accountable for access decisions. |
| retention status | Active, under review, hold, expired, missing, or stale. |
| access classification | Restricted, confidential, or other school-approved classification. |

Optional index fields may include reviewer, expiry date placeholder, access request ticket, read-back status, and notes. Do not add raw log content, command output, environment values, database records, tokens, passwords, answer keys, or credential material to the index.

## Storage and Search Boundary

The index must be stored outside public web roots and outside SvelteKit static/public assets. It must also stay out of Git-tracked tmp or generated evidence directories. Template docs in this repository may describe the index shape, but real index rows are ops-controlled records, not committed tmp data except template docs.

Search must use redacted metadata only:

- release id.
- commit.
- date.
- owner.
- retention status.
- access classification.
- manifest sha256.

Search must not use or expose:

- no raw secrets.
- no full logs.
- no env.
- no raw `.env`, `printenv`, PM2 env dump, database dump, full request headers, bearer values, exam tokens, JWTs, worker keys, API keys, passwords, or answer keys.

## Lookup/Read-Back Workflow

Use this lookup/read-back workflow before retrieving archive contents:

1. Search the redacted index by release id, commit, or date.
2. Confirm owner, retention status, access classification, and archive path.
3. Confirm the requester is allowed by the access classification and owner policy.
4. Open the archive read-only from ops-controlled storage.
5. Recompute SHA-256 for `cbt-release-manifest.json` and compare it with the index manifest sha256.
6. Parse `cbt-release-manifest.json` and verify its algorithm is `sha256`.
7. Recompute required artifact checksums before retrieval or sharing.
8. Read back only the minimum evidence needed for the request.
9. Record read-back result, requester, reviewer, timestamp, and any follow-up.

Retrieval should stop if the checksum verification before retrieval fails, the archive is missing, retention status blocks access, or the requester is not authorized.

## Access Request/Audit Placeholders

Record access request/audit placeholders alongside the index or in the school-approved ops tracker:

| Date | Release id | Commit | Requester | Purpose | Owner approval | Reviewer | Result | Notes |
|------|------------|--------|-----------|---------|----------------|----------|--------|-------|
| | | | | | approved / rejected / pending | | retrieved / denied / stale / missing | |

Access audit entries must not include raw secrets, full logs, or env values. Use release id, commit, manifest sha256, and redacted archive path references instead.

## Stale/Missing Archive Handling

If an index row points to a stale or missing archive:

- mark retention status as `stale` or `missing`.
- do not delete the index row during triage.
- restrict retrieval until the owner reviews the mismatch.
- compare available handoff metadata with the release id, commit, archive path, and manifest sha256.
- record whether the archive was moved, expired, quarantined, or never completed.
- create an owner follow-up for archive repair, retention decision, or formal exception.

If an archive exists but its manifest checksum does not match the index manifest sha256, treat it as stale until reviewed. Do not rebuild, overwrite, or delete it in Phase 12.

## Checksum Verification Before Retrieval

Before any evidence is retrieved or shared:

- recompute SHA-256 for `cbt-release-manifest.json`.
- compare the recomputed value with the index manifest sha256.
- parse the manifest and confirm `algorithm` is `sha256`.
- recompute checksums for required artifacts needed by the request.
- verify secret scan status is `pass`.
- record pass/fail in the access request/audit placeholders.

Failed checksum verification blocks retrieval. Keep the archive restricted, notify the owner, and record the failure without exposing raw evidence.

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

Targeted validation for Phase 12:

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

`make ops-health` is allowed as read-only live health smoke. Do not run deploy, PM2 restart, live migration, ad hoc SQL, PM2-changing commands, or public `/api/cbt` route work as part of Phase 12 implementation validation.

## Acceptance Criteria

Phase 12 diterima bila:

- guard test lulus setelah sebelumnya RED karena dokumen Phase 12 dan evidence index/retrieval section belum ada.
- `docs/cbt-proposal-integration-phase-12.md` describes the redacted evidence index/search and retrieval policy after Phase 11 archive/retention.
- index fields include release id, commit, date, archive path, manifest sha256, owner, retention status, and access classification.
- index storage is outside public web roots and not committed as tmp data except template docs.
- policy blocks raw secrets, full logs, env dumps, and secret-bearing evidence.
- lookup/read-back workflow requires checksum verification before retrieval.
- access request/audit placeholders and stale/missing archive handling are documented.
- no product runtime, DB schema/migration, deploy automation, PM2 restart, live DB write, secret, or `/api/cbt` public route change is introduced.
