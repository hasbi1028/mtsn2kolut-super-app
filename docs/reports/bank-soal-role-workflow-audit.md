# Bank Soal Role Workflow — Sprint 0 Production Audit

> Sprint 0 audit for `.hermes/plans/2026-05-16_bank-soal-role-workflow.md`.
>
> Scope: read-only production/semi-production data discovery before any RBAC/workflow behavior change. No credentials are stored in this report.

## Audit metadata

- Date/time: 2026-05-16 08:48:45 WITA
- Branch: `feature/comprehensive-improvements`
- Plan: `.hermes/plans/2026-05-16_bank-soal-role-workflow.md`
- Audit SQL: `scripts/audit-bank-soal-role-workflow.sql`
- Raw audit output: `/tmp/bank-soal-role-workflow-audit.out` on the local machine only
- Production data changed: **No**
- PM2 deploy/restart: **No**
- Credential exposure: **No**

## Executive summary

Sprint 0 is **GO for Sprint 1 additive-only implementation** with safety conditions.

Key findings:

- Total Bank Soal questions currently audited: **49**.
- Workflow statuses found: `approved`, `draft`, `review`.
- Unknown/incompatible workflow statuses: **0**.
- Existing package-question links: **81 links** across **4 packages**, all using approved/published-compatible questions.
- Package links using non-approved/non-published questions: **0**.
- Dynamic RBAC currently grants `guru` privileged Bank Soal permissions (`bank_soal.review`, `bank_soal.update`, `bank_soal.analytics`) to **76 guru users**.
- `class_subject_assignments` currently returns **0 reviewer candidates**, so reviewer scopes cannot be auto-derived from guru-mapel yet.
- Existing DB workflow constraint only allows old statuses: `draft`, `review`, `approved`, `rejected`.
- Existing Bank Soal/versioning related tables are only `cbt_questions`, `cbt_question_assets`, and `cbt_question_audit_logs`; no `bank_soal_*` workflow tables exist yet.

Sprint 1 may safely proceed if it remains additive:

- Add new permissions without revoking existing grants.
- Add reviewer scope table without enforcing it yet.
- Do not lock down guru visibility yet because reviewer scopes are not populated.
- Keep legacy `review` status accepted.

## 1. Question workflow/status distribution

- `approved` + `published`: **33** questions
  - missing author username: 0
  - missing reviewer username: 0
  - missing approver username: 0
  - missing reviewed_at: 0
  - missing approved_at: 0
- `draft` + `draft`: **1** question
  - missing author username: 0
  - missing reviewer username: 1
  - missing approver username: 1
  - missing reviewed_at: 1
  - missing approved_at: 1
- `review` + `draft`: **15** questions
  - missing author username: 0
  - missing reviewer username: 0
  - missing approver username: 15
  - missing reviewed_at: 0
  - missing approved_at: 15

Interpretation:

- Existing operational workflow is still using legacy `review` as “submitted/menunggu review”.
- `review` must remain supported during Sprint 1–3.
- The 15 `review/draft` questions must not be hidden from existing admin/reviewer flows during rollout.

## 2. Status compatibility risk

Unknown statuses outside target compatibility set: **0**.

Target-compatible statuses checked:

- `draft`
- `review` legacy alias
- `submitted`
- `revision_needed`
- `reviewed`
- `approved`
- `published`
- `rejected`
- `archived`

Current constraint:

```text
chk_cbt_questions_workflow_status = CHECK workflow_status IN ('draft', 'review', 'approved', 'rejected')
```

Risk:

- Before backend can write `submitted`, `revision_needed`, `reviewed`, `published`, or `archived`, Sprint 2 needs a backward-compatible constraint migration.
- Do not remove `review` from allowed values until all old data/code is migrated.

## 3. Existing package/event impact

Package-question links by status:

- `approved` + `published`: **81** package-question links
  - affected packages: **4**
  - unique questions: **31**

Non-approved/non-published-compatible package-question links: **0**.

Decision:

- Sprint 4 hardening is lower risk because existing package links already use approved/published-compatible questions.
- Still implement Sprint 4 in report/compatibility mode first, then hard block only for new package assignments.

## 4. Existing Bank Soal RBAC permissions

Current role permissions include:

- `admin` has:
  - `bank_soal.analytics`
  - `bank_soal.create`
  - `bank_soal.delete`
  - `bank_soal.import`
  - `bank_soal.publish`
  - `bank_soal.read`
  - `bank_soal.review`
  - `bank_soal.settings`
  - `bank_soal.update`
- `guru` has:
  - `bank_soal.analytics`
  - `bank_soal.create`
  - `bank_soal.read`
  - `bank_soal.review`
  - `bank_soal.update`
- E2E/test roles also exist:
  - `bank_soal_e2e_creator`
  - `bank_soal_e2e_importer`
  - `bank_soal_e2e_readonly`
  - `bank_soal_e2e_reviewer`

User counts by dynamic RBAC role and Bank Soal permission:

- `admin`: **4 users** with each admin Bank Soal permission.
- `guru`: **76 users** with `bank_soal.analytics`, `bank_soal.create`, `bank_soal.read`, `bank_soal.review`, and `bank_soal.update`.
- E2E/test roles: **1 user** each.

Account role distribution:

- `admin`: **2 users**
  - linked to employee: 1
  - not linked to employee: 1
- `guru`: **76 users**
  - linked to employee: 76
  - not linked to employee: 0
- `siswa`: **164 users**
  - linked to employee: 0
  - not linked to employee: 164

Risk:

- `guru` currently has `bank_soal.review`; removing it immediately would change behavior for 76 users.
- Sprint 1 must **not revoke** existing guru permissions without explicit approval.
- Hardening should be separate after reviewer scope is configured and smoke-tested.

## 5. Reviewer/approver candidates from guru-mapel

Reviewer candidates from `class_subject_assignments`: **0 rows**.

User-linked teacher assignment candidates: **0 rows**.

Interpretation:

- Current production DB cannot auto-generate reviewer scope from guru-mapel yet.
- Sprint 1 should create reviewer scope infrastructure, but leave scopes empty or allow manual admin setup.
- Do not enforce reviewer-scope visibility until scopes are populated.

Recommended action before hard enforcement:

1. Populate `class_subject_assignments` through the Academic/Guru Mapel Matrix flow, or
2. Let admin manually assign reviewer scope per mapel/tingkat in `/bank-soal/pengaturan`.

## 6. Code/UI compatibility scan

Compatibility scan command:

```bash
git grep -n "workflow_status\|bank_soal.review\|bank_soal.publish\|review_notes\|reviewer_username" -- services/core-api apps/web-admin
```

Result:

- Total matches: **489**
- Files touched by matches: **84**

Largest match groups:

- `services/core-api/internal/repository/postgres/cbt_questions.sql.go`: 95
- `services/core-api/db/queries/cbt_questions.sql`: 65
- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`: 32
- `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`: 21
- `services/core-api/internal/handler/cbt_question_workflow.go`: 19
- `apps/web-admin/src/routes/bank-soal/_components/soal-workspace.model.ts`: 14
- `services/core-api/internal/handler/cbt_question_serialize.go`: 13
- `apps/web-admin/src/routes/bank-soal/_components/PublishBankSoalPage.svelte`: 9

Interpretation:

- Workflow logic is already spread through backend queries, generated sqlc, service/handler, and Bank Soal UI components.
- Sprint 2 must update compatibility carefully and regenerate sqlc.
- Sprint 3 UI must keep legacy `review` readable as “Menunggu Review”.

## 7. Go/no-go for Sprint 1

Decision: **GO, with restrictions**.

Allowed in Sprint 1:

- Add new permission rows idempotently.
- Add reviewer scope table idempotently.
- Add read/list/upsert/delete APIs for reviewer scope.
- Add admin UI for reviewer scope management.
- Add tests and documentation.

Not allowed in Sprint 1 without explicit approval:

- Revoke `bank_soal.review`/`bank_soal.update`/`bank_soal.analytics` from `guru`.
- Enforce scope-based filtering that hides existing questions.
- Rename `review` data to `submitted`.
- Restrict package builder behavior.
- Deploy/restart PM2.

## 8. Recommended Sprint 1 safety gates

- Migration must be additive and idempotent.
- New permissions must use `INSERT ... ON CONFLICT`.
- Reviewer scope table must start empty unless admin explicitly fills it.
- Admin/read_all escape hatch must be designed before any filtering changes.
- No permission revoke in Sprint 1.
- Existing `review` status remains accepted.
- No production deploy/restart unless explicitly requested.

## 9. Validation log

Completed:

```bash
npm --prefix apps/web-admin run check
# PASS: svelte-check found 0 errors and 0 warnings

cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
# PASS

go test ./internal/handler ./internal/service ./internal/repository/postgres
# PASS: handler/service/repository packages ok

go build -o /tmp/core-api-bank-soal-workflow-sprint0 ./cmd/api
# PASS
```
