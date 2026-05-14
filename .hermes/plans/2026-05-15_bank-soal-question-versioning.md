# Bank Soal Question Versioning Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build full question versioning for Bank Soal so approved/published/used questions are immutable, revisions become new linked versions, and packages/sessions keep historical question versions stable.

**Architecture:** Additive DB migration adds question lineage fields to `cbt_questions`, keeps current question rows as immutable version records, and tracks latest/current version per lineage with a stable `version_group_id`. Existing package/session answer relations continue pointing to exact `cbt_questions.id`, so historical exams remain valid. Web-admin routes all “lihat detail” traffic to a composer-based detail page with read-only/editable behavior based on workflow/usage locks.

**Tech Stack:** PostgreSQL migrations, Go core-api + sqlc, SvelteKit BFF/web-admin, PM2 deploy.

---

## Current findings

- `cbt_questions` already has `version INTEGER NOT NULL DEFAULT 1`, but it is currently incremented on every in-place update in `UpdateCbtQuestion`; it is not a true lineage/version number.
- Existing safe revision action already exists:
  - UI list action: **Buat Revisi Baru** calls `POST /api/bank-soal/questions/{id}/revision`.
  - Backend service: `DuplicateForRevision` copies the question and records audit metadata `source_question_id`.
- Existing safe return action already exists:
  - workflow action `return_revision` for approved, un-used, un-published questions.
- Existing package/answer integrity is good because `cbt_package_questions.question_id` and `cbt_student_answers.question_id` point to the exact question row.
- Missing pieces for real Opsi C:
  - no durable lineage fields (`version_group_id`, `version_number`, `source_question_id`, `supersedes_question_id`, `is_latest_version`);
  - no endpoint to list version history;
  - composer detail still needs official read-only/view mode instead of modal-only detail;
  - list UI needs latest/history cues and route to composer detail.

---

## Non-negotiable safety rules

1. No in-place edit for published/used questions.
2. Packages/sessions must keep pointing to the exact old `question_id`.
3. New revision creates a new `cbt_questions.id` linked to old lineage.
4. Migration must be additive and backwards compatible.
5. Deploy order: backup DB -> migration -> sqlc/core-api build -> web-admin build -> restart PM2 -> health/smoke.
6. Do not display or store DB credentials; use `[REDACTED]` if encountered.

---

## Target UX

### `/bank-soal/soal/[id]` composer detail page

Use the same composer UI for both viewing and editing:

- Draft/perlu revisi: editable composer.
- Review/approved/published/used: read-only composer.
- Approved but not used: read-only + **Kembalikan ke Revisi**.
- Published/used: read-only + **Buat Revisi Baru**.
- Version panel: show `v1`, `v2`, latest/current badge, source/superseded relation, and link to all versions.

### `/bank-soal/daftar`

- **Lihat** opens composer detail page, not modal.
- Latest version badge on each row.
- Optional filter: latest only / all versions.
- For historical versions: show “Versi lama — terkunci”.

---

## Data model proposal

Add to `cbt_questions`:

```sql
ALTER TABLE cbt_questions
  ADD COLUMN IF NOT EXISTS version_group_id UUID,
  ADD COLUMN IF NOT EXISTS version_number INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS source_question_id UUID REFERENCES cbt_questions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS supersedes_question_id UUID REFERENCES cbt_questions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS is_latest_version BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS version_note TEXT NOT NULL DEFAULT '';
```

Backfill:

```sql
UPDATE cbt_questions
SET version_group_id = COALESCE(version_group_id, id)
WHERE version_group_id IS NULL;
```

Indexes:

```sql
CREATE INDEX IF NOT EXISTS idx_cbt_questions_version_group ON cbt_questions(version_group_id, version_number DESC);
CREATE INDEX IF NOT EXISTS idx_cbt_questions_latest_version ON cbt_questions(is_latest_version, workflow_status, status);
CREATE INDEX IF NOT EXISTS idx_cbt_questions_source_question ON cbt_questions(source_question_id);
CREATE INDEX IF NOT EXISTS idx_cbt_questions_supersedes_question ON cbt_questions(supersedes_question_id);
```

Optional post-backfill constraint:

```sql
ALTER TABLE cbt_questions
  ADD CONSTRAINT chk_cbt_questions_version_number_positive CHECK (version_number >= 1);
```

Do **not** add a partial unique index on `(version_group_id) WHERE is_latest_version` until we have cleaned possible concurrent/legacy revision edge cases; add later after monitoring.

---

## Task 1: Add migration for version lineage

**Objective:** Add version lineage columns with safe backfill.

**Files:**
- Create: `services/core-api/db/migrations/102_cbt_question_versioning.sql`

**Steps:**
1. Create migration with additive columns, backfill, indexes, positive check.
2. Validate migration in transaction against production-like DB:
   ```bash
   cd services/core-api
   set -a; source .env >/dev/null 2>&1; set +a
   { echo 'BEGIN;'; cat db/migrations/102_cbt_question_versioning.sql; echo 'ROLLBACK;'; } | psql "$DATABASE_URL" -v ON_ERROR_STOP=1
   ```
3. Expected: exit 0, no credential output.

---

## Task 2: Extend sqlc question queries

**Objective:** Expose version lineage fields and version history queries.

**Files:**
- Modify: `services/core-api/db/queries/cbt_questions.sql`
- Generated: `services/core-api/internal/repository/postgres/cbt_questions.sql.go`

**Add/modify:**
1. Include fields in list/detail SELECT projections:
   - `version_group_id`
   - `version_number`
   - `source_question_id`
   - `supersedes_question_id`
   - `is_latest_version`
   - `version_note`
2. Add create params for lineage fields.
3. Add update params for `is_latest_version` and `version_note` only if needed.
4. Add query:
   ```sql
   -- name: ListCbtQuestionVersions :many
   SELECT q.id, q.code, q.workflow_status, q.status, q.version_number,
          q.is_latest_version, q.source_question_id, q.supersedes_question_id,
          q.version_note, q.created_at, q.updated_at,
          q.author_username, q.reviewer_username, q.approver_username
   FROM cbt_questions q
   WHERE q.version_group_id = (
     SELECT version_group_id FROM cbt_questions WHERE id = $1
   )
   ORDER BY q.version_number DESC, q.created_at DESC;
   ```
5. Add query:
   ```sql
   -- name: MarkCbtQuestionVersionNotLatest :exec
   UPDATE cbt_questions SET is_latest_version = FALSE, updated_at = NOW()
   WHERE id = $1;
   ```
6. Run:
   ```bash
   cd services/core-api
   /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   ```

---

## Task 3: Extend service model and creation/update logic

**Objective:** Ensure new rows get correct lineage and old rows become non-latest when making revision.

**Files:**
- Modify: `services/core-api/internal/service/cbt_question.go`
- Modify: `services/core-api/internal/service/cbt_question_authoring.go`
- Modify: `services/core-api/internal/service/cbt_question_workflow.go`
- Modify tests under `services/core-api/internal/service/cbt_question_test.go`

**Rules:**
- New normal question:
  - `version_group_id = id` is hard to set during insert with generated UUID unless query uses CTE; acceptable pattern: create row, then if `version_group_id` missing, update it to own id. Better: update create SQL to generate id in CTE and insert id into both `id` and `version_group_id`.
- Duplicate/revision:
  - `version_group_id = current.version_group_id` if set, otherwise `current.id`.
  - `version_number = max(version_number in group) + 1`.
  - `source_question_id = current.id`.
  - `supersedes_question_id = current.id`.
  - `is_latest_version = TRUE` for new row.
  - Mark old/current row `is_latest_version = FALSE`.
- Draft duplicate not intended as revision can either start a new group or be a version copy depending on action:
  - `DuplicateAsDraft`: new group, version 1.
  - `DuplicateForRevision`: same group, next version.

**Tests:**
- `DuplicateForRevision` marks source not latest.
- `DuplicateForRevision` creates version `n+1` in same group.
- `DuplicateAsDraft` creates independent group/version 1.
- Used/published source remains unchanged and package references remain untouched.

---

## Task 4: Add version history endpoint

**Objective:** Let UI show version chain.

**Files:**
- Modify: `services/core-api/internal/handler/cbt_question.go`
- Modify/add handler methods/tests.
- Modify: `services/core-api/cmd/api/main.go`

**Endpoint:**

```text
GET /api/bank-soal/questions/{id}/versions
```

Response shape:

```json
{
  "items": [
    {
      "id": "uuid",
      "code": "MTK-001-REV-...",
      "version_number": 2,
      "is_latest_version": true,
      "workflow_status": "rejected",
      "status": "draft",
      "source_question_id": "uuid",
      "supersedes_question_id": "uuid",
      "version_note": "...",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

**Validation:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
```

---

## Task 5: Add SvelteKit BFF routes

**Objective:** Keep SvelteKit as proxy/BFF only.

**Files:**
- Create: `apps/web-admin/src/routes/api/bank-soal/questions/[id]/versions/+server.ts`
- Create/modify: `apps/web-admin/src/lib/server/cbt-backend-proxy/questions/[id]/versions/+server.ts`

**Pattern:** follow existing files under:
- `apps/web-admin/src/lib/server/cbt-backend-proxy/questions/[id]/revision/+server.ts`
- `apps/web-admin/src/routes/api/bank-soal/questions/[id]/revision/+server.ts`

---

## Task 6: Composer detail page read-only/edit mode

**Objective:** Replace modal detail with composer-based detail.

**Files:**
- Create: `apps/web-admin/src/routes/bank-soal/soal/[id]/+page.svelte`
- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Modify model helpers: `apps/web-admin/src/routes/bank-soal/_components/soal-workspace.model.ts`

**Implement:**
- Add props/mode support to workspace:
  - `questionId?: string`
  - `mode?: 'create' | 'detail'`
  - computed `readOnly` based on workflow/status/package_count/answer_count/latest flag.
- Disable all inputs and save buttons when `readOnly`.
- Banner copy:
  - approved not used: “Soal sudah disetujui. Kembalikan ke revisi untuk mengedit.”
  - published/used: “Soal sudah dipakai/terbit. Buat revisi baru agar riwayat ujian lama tetap valid.”
  - old version: “Ini versi lama. Paket lama tetap memakai versi ini.”
- Actions in composer header:
  - **Kembalikan ke Revisi** if approved and not used.
  - **Buat Revisi Baru** if published/used/old version.
  - **Ajukan Review** when editable revision.

---

## Task 7: Version panel UI

**Objective:** Show all versions and navigation.

**Files:**
- Create: `apps/web-admin/src/routes/bank-soal/_components/QuestionVersionPanel.svelte`
- Modify: `SoalWorkspacePage.svelte` to include panel.

**UI:**
- List version chips: v1, v2, v3.
- Latest badge.
- Current row highlighted.
- Clicking a version opens `/bank-soal/soal/{id}`.

---

## Task 8: Change list/review navigation away from modal detail

**Objective:** Make list “Lihat” open composer detail page.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
- Modify: `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`

**Rules:**
- `Lihat` opens `/bank-soal/soal/{id}`.
- Existing modal detail can remain temporarily but should not be primary path.
- Add latest/version badges in list rows.

---

## Task 9: Verification and deploy

**Objective:** Validate, commit, deploy safely.

**Commands:**

```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-versioning ./cmd/api
cd ../..
npm --prefix apps/web-admin run build
```

Before production migration/deploy:

```bash
set -a; source services/core-api/.env >/dev/null 2>&1; set +a
stamp=$(date +%Y%m%d-%H%M%S)
dir=/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql
mkdir -p "$dir"
out="$dir/pre-bank-soal-versioning-$stamp.dump"
pg_dump "$DATABASE_URL" -Fc -f "$out"
sha256sum "$out"
```

Deploy order:

```bash
set -a; source services/core-api/.env >/dev/null 2>&1; set +a
npm run db:migrate
cd services/core-api
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
cd ../..
npm --prefix apps/web-admin run build
pm2 restart mtsn2kolut-web-admin --update-env
pm2 save
curl -fsS http://127.0.0.1:8080/health
curl -s -o /tmp/bank-soal-versioning-smoke.html -w '%{http_code}' http://127.0.0.1:8021/bank-soal/daftar
```

Expected:
- health `ok`
- web status `302` if not logged in, or `200` if session exists.

---

## Suggested implementation commits

1. `feat(bank-soal): add question version lineage schema`
2. `feat(bank-soal): track safe question revision versions`
3. `feat(bank-soal): add question version history API`
4. `feat(bank-soal): add composer read-only detail route`
5. `feat(bank-soal): surface question version history in UI`

---

## Recommendation

Execute in **two deployable phases**:

### Phase 1 — Backend lineage foundation
- Migration + sqlc + service + version history endpoint.
- No big UI refactor yet.
- Safe to deploy first.

### Phase 2 — Composer detail UX
- Route `/bank-soal/soal/[id]`.
- Read-only composer mode.
- Version panel.
- List/review navigation changes.

This reduces risk because DB lineage can be validated before changing the main teacher/reviewer workflow.
