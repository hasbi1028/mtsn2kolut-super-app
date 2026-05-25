# Bank Soal Table-Safe Composer Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task. Use fresh subagents with two-stage review: spec compliance first, code quality second. Do **not** deploy/restart production until Hasbi explicitly approves implementation/deploy.

**Goal:** Mencegah tabel soal Bank Soal rusak/terhambur saat soal dibuka untuk revisi dan mencegah HTML tabel yang sudah rusak tersimpan kembali ke database.

**Architecture:** Implementasi dilakukan sebagai guard bertahap, bukan migrasi editor besar. Server/DB tetap menjadi sumber kebenaran untuk soal existing. Frontend menambah deteksi struktur tabel, guard sebelum simpan/sync offline, dan cleanup draft lokal. Backend hanya ditambah test/regression jika diperlukan untuk memastikan sanitizer tetap mempertahankan struktur tabel.

**Tech Stack:** SvelteKit/Svelte 5, TypeScript, Quill `LegacyRichTextEditor`, `RichContent`, IndexedDB offline queue, Go Core API, PostgreSQL, sqlc, bluemonday sanitizer.

---

## Current Context

Patch yang sudah ada:

- Tahap 1 CSS table hotfix:
  - `apps/web-admin/src/lib/components/LegacyRichTextEditor.svelte`
  - `apps/web-admin/src/lib/components/RichContent.svelte`
- Tahap 2 prompt local draft:
  - `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
  - commit terakhir: `7705828 fix(bank-soal): prompt before restoring local edit drafts`

Temuan agent swarm:

1. Backend/API/storage kemungkinan bukan penyebab utama. `stem_html`/`stimulus_html` disanitasi tetapi tabel dasar diizinkan oleh `AllowTables()`.
2. Titik risiko utama adalah frontend editor Quill (`LegacyRichTextEditor`) yang memakai `dangerouslyPasteHTML` dan hydrate via `root.innerHTML`.
3. Patch local-draft prompt sudah mencegah restore otomatis, tetapi masih ada risiko:
   - user memilih `Pakai Konsep Lokal`;
   - offline queue lama otomatis sync;
   - `Muat Data Server Saja` tidak selalu menghapus draft lokal lama;
   - belum ada guard sebelum save untuk mencegah tabel rusak tersimpan.
4. Migrasi penuh Tiptap layak tetapi bukan quick fix; sebaiknya setelah guard stabil.

## Scope This Plan

### Included

- Audit DB/fixtures untuk soal yang punya tabel.
- Utility frontend untuk menganalisis struktur tabel HTML.
- Guard sebelum `Simpan Konsep` / `Kirim Verifikasi` / PUT update jika HTML tabel tampak rusak.
- Guard offline queue agar payload lama berisi tabel tidak otomatis menimpa server tanpa keputusan user.
- Cleanup draft lokal ketika user memilih data server.
- Test/unit regression untuk utility dan minimal backend sanitizer check.
- Agent swarm workflow untuk implementasi, review, dan validasi.

### Not Included

- Migrasi penuh Quill ke Tiptap.
- Full server-only composer.
- Perubahan skema database.
- Mass repair DB untuk soal yang sudah rusak, kecuali hanya audit/report.
- Deploy/restart tanpa persetujuan eksplisit.

---

## Swarm Workflow Best Practice

### Roles

Use max 4 implementer/reviewer lanes at a time; avoid overlapping writes.

1. **Coordinator / Parent Agent**
   - Holds plan.
   - Creates todo list.
   - Dispatches subagents.
   - Integrates diffs.
   - Runs final verification.
   - Commits only after verification.

2. **Agent A — Data & Backend Verification**
   - Owns SQL audit/report and backend sanitizer tests.
   - Files likely touched only under `services/core-api/internal/service/*_test.go` and optional `reports/` if approved.

3. **Agent B — Frontend Table Analyzer Utility**
   - Owns pure TypeScript utility and tests.
   - Files likely touched:
     - `apps/web-admin/src/lib/bank-soal/table-html-guard.ts`
     - corresponding test file if test framework exists.

4. **Agent C — Composer Save/Local Draft Guard**
   - Owns `SoalWorkspacePage.svelte` integration.
   - Must not touch backend.

5. **Agent D — Offline Queue Guard**
   - Owns `bank-soal-offline.ts` queue inspection helpers and `SoalWorkspacePage.svelte` sync behavior.
   - Coordinate with Agent C because both may touch `SoalWorkspacePage.svelte`; run sequentially if needed.

6. **Reviewer Agents**
   - Spec reviewer: checks exact task requirements.
   - Quality reviewer: checks correctness, maintainability, UX copy, edge cases, no credentials.

### Hard Rules

- Do not let two implementer agents edit `SoalWorkspacePage.svelte` simultaneously.
- Subagents must return changed file list and verification commands they ran.
- Parent must distrust self-reported success until it runs final checks.
- Parent must inspect `git status --short` before staging.
- Use explicit `git add <file>` only; never `git add -A`.
- No production PM2 restart until user approves.
- No credential output in logs/reports.

### Review Gates

Each coding task must pass:

1. **Spec Compliance Review**
   - Does it satisfy the exact task?
   - Did it avoid extra scope?
   - Did it touch only allowed files?

2. **Code Quality Review**
   - Is the logic safe?
   - Are tests meaningful?
   - Are user-facing warnings clear in Indonesian?
   - Does it avoid accidental data loss?

3. **Integration Gate**
   - `npm --prefix apps/web-admin run check`
   - `npm --prefix apps/web-admin run build`
   - `git diff --check`
   - If Go changed:
     - `cd services/core-api && go test ./internal/service -run 'HTML|Sanitize|Question' -count=1`
     - `cd services/core-api && go test ./...`
     - `cd services/core-api && go build -o /tmp/core-api-banksoal-table-check ./cmd/api`

---

## Implementation Tasks

### Task 0: Pre-flight Audit and Baseline

**Objective:** Capture the current DB/table state and confirm whether corruption is currently in DB or only editor/display.

**Files:**
- Create optional report: `reports/bank-soal-table-audit/YYYYMMDD-HHMMSS-table-html-audit.md`
- No source edits.

**Steps:**

1. Run read-only SQL to count HTML tables:

```sql
SELECT count(*) AS questions_with_stem_table
FROM cbt_questions
WHERE lower(stem_html) LIKE '%<table%';
```

2. Sample recent table questions:

```sql
SELECT id, code, updated_at,
       lower(stem_html) LIKE '%<table%' AS has_table,
       lower(stem_html) LIKE '%<tr%' AS has_tr,
       lower(stem_html) LIKE '%<td%' AS has_td,
       lower(stem_html) LIKE '%<th%' AS has_th,
       left(stem_html, 800) AS sample
FROM cbt_questions
WHERE lower(stem_html) LIKE '%<table%'
ORDER BY updated_at DESC
LIMIT 20;
```

3. Check options with tables:

```sql
SELECT id, code, options
FROM cbt_questions
WHERE options::text ILIKE '%<table%'
ORDER BY updated_at DESC
LIMIT 20;
```

4. If user gives a specific broken question ID/code, compare:

```sql
SELECT id, code, stem_html, question_text, options
FROM cbt_questions
WHERE id = '<QUESTION_ID_OR_CODE_LOOKUP>' OR code = '<CODE>';
```

**Expected Result:** Report states whether DB HTML still has normal table structure. No mutation.

**Commit:** None unless report is intentionally added to repo; prefer uncommitted local report unless user wants saved artifact.

---

### Task 1: Add Pure Table HTML Analyzer Utility

**Objective:** Create a pure frontend utility that can detect table structure health without depending on Svelte component state.

**Files:**
- Create: `apps/web-admin/src/lib/bank-soal/table-html-guard.ts`
- Create/modify test if existing frontend test setup supports it, e.g. `apps/web-admin/src/lib/bank-soal/table-html-guard.test.ts`

**Design:**

Export types/functions:

```ts
export type TableHtmlIssueSeverity = 'info' | 'warning' | 'blocker';

export type TableHtmlIssue = {
  severity: TableHtmlIssueSeverity;
  code:
    | 'table-without-rows'
    | 'table-without-cells'
    | 'cell-count-dropped'
    | 'row-count-dropped'
    | 'contains-complex-table-attrs'
    | 'suspicious-flat-table-text';
  message: string;
};

export type TableHtmlSummary = {
  tableCount: number;
  rowCount: number;
  cellCount: number;
  hasTable: boolean;
  hasRows: boolean;
  hasCells: boolean;
  hasComplexAttrs: boolean;
};

export function summarizeTableHtml(html: string): TableHtmlSummary;

export function compareTableHtmlBeforeSave(args: {
  serverHtml: string;
  draftHtml: string;
  fieldLabel: string;
}): TableHtmlIssue[];

export function hasBlockingTableHtmlIssues(issues: TableHtmlIssue[]): boolean;
```

**Rules:**

- If no table in either server/draft, no blocker.
- If draft has `<table` but no `<tr`, blocker.
- If draft has `<table` but no `<td`/`<th`, blocker.
- If server had tables and draft table count becomes zero, blocker.
- If server cell count > 0 and draft cell count drops by >= 40%, blocker.
- If complex attrs exist (`colspan`, `rowspan`, `width=`, inline `style=` on table/cells), warning only, not blocker.
- Avoid browser-only APIs if tests run in Node. Prefer regex/DOMParser fallback pattern; if DOMParser is used, guard for browser.

**Verification:**

- Unit tests cover:
  - valid simple table;
  - table without rows;
  - table without cells;
  - server table lost in draft;
  - large cell count drop;
  - complex attrs warning.

**Commands:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

**Commit Message:**

```bash
git add apps/web-admin/src/lib/bank-soal/table-html-guard.ts <test-file-if-created>
git commit -m "feat(bank-soal): add table html guard utility"
```

---

### Task 2: Store Server Baseline HTML During Edit/Revisi

**Objective:** Keep the latest server HTML baseline so save guard can compare current editor output against the clean loaded source.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`

**Implementation Notes:**

Add state similar to:

```ts
let serverBaselineHtml = $state<{
  stem: string;
  stimulus: string;
  explanation: string;
  rubric: string;
  options: Record<string, string>;
} | null>(null);
```

When opening edit from detail server:

- after loading `d.stem_html`, `d.stimulus_html`, etc., store baseline.
- for fallback from list item, store whatever server/list HTML exists, but mark it as weaker baseline if needed.

When creating a new question:

- baseline is null or empty.

When user chooses `Pakai Konsep Lokal`:

- keep server baseline unchanged.
- add warning that save may be blocked if local draft table structure differs from server.

**Verification:**

- No behavior change yet except state tracking.
- `npm --prefix apps/web-admin run check`
- `npm --prefix apps/web-admin run build`

**Commit Message:**

```bash
git add apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte
git commit -m "feat(bank-soal): track server table html baseline"
```

---

### Task 3: Add Save Guard for Table HTML

**Objective:** Block accidental save if current editor HTML appears to have lost table structure compared to server baseline.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Import: `apps/web-admin/src/lib/bank-soal/table-html-guard.ts`

**Implementation Notes:**

Add helper in component:

```ts
function collectTableGuardIssues(): TableHtmlIssue[] {
  // compare serverBaselineHtml.stem vs fStem, stimulus vs fStimulus, explanation/rubric if relevant
  // compare option baselines by key/index vs current option html
}
```

Before final save/update request:

- In the save handler used by `Simpan Konsep` and `Kirim Verifikasi`, call guard before `fetch`.
- If blocker exists:
  - prevent save;
  - show `toast.error` in Indonesian;
  - open a small dialog/list of affected fields if existing dialog pattern is available.

Suggested copy:

> “Simpan dibatalkan karena struktur tabel tampak berubah/rusak pada bagian {field}. Muat ulang data server atau pilih Data Server & Hapus Lokal sebelum menyimpan.”

Important behavior:

- New questions with newly inserted tables should not be blocked just because there is no baseline.
- Existing questions with tables should be blocked only if table structure disappears/drops suspiciously.
- Do not block normal text edits inside table cells if row/cell structure remains stable.
- Allow an explicit override only if user is admin/reviewer? Prefer no override for first patch to avoid accidental data loss. If override needed, make it a separate future task.

**Verification:**

Manual test cases with dev data or mocked component state:

1. Existing server table unchanged → save allowed.
2. Existing server table text edited but same cells → save allowed.
3. Existing server table lost all `<td>` → save blocked.
4. Existing server table replaced by plain text → save blocked.
5. New question with table → save allowed.

Commands:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
git diff --check
```

**Commit Message:**

```bash
git add apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte apps/web-admin/src/lib/bank-soal/table-html-guard.ts
git commit -m "fix(bank-soal): block saves with damaged table html"
```

---

### Task 4: Strengthen Local Draft Choice Cleanup

**Objective:** Ensure choosing server data prevents stale local table drafts from reappearing later.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`

**Current Risk:**

- `Muat Data Server Saja` does not delete local draft.
- If user closes before autosave rewrites draft with clean server data, old draft can appear again.

**Implementation Options:**

Preferred:

- Rename/adjust option behavior:
  - `Muat Data Server Saja` keeps current session server-only but also marks old local draft ignored for this question/session.
  - `Data Server & Hapus Lokal` deletes local draft permanently.
- Add an internal `ignoredLocalDraftKeys` set/session state so the same key is not prompted again during same page lifecycle.
- Consider making `Muat Data Server Saja` also overwrite local draft with server snapshot immediately after the choice, not after 700ms autosave delay.

Safer minimal behavior:

- When choosing `Muat Data Server Saja`, immediately save current server-loaded payload into local draft key as clean snapshot after `suppressDraftAutosave` is released.
- When choosing `Data Server & Hapus Lokal`, delete the draft and do not recreate it until user edits.

**Verification:**

1. Create local draft, open edit, choose server only, close/reopen quickly.
2. Confirm old local draft does not immediately reappear as default.
3. Choose server & delete local; confirm prompt no longer appears for same key.

**Commands:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

**Commit Message:**

```bash
git add apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte
git commit -m "fix(bank-soal): prevent stale table drafts from reappearing"
```

---

### Task 5: Guard Offline Queue Sync for Table Payloads

**Objective:** Prevent old offline queue payloads containing table HTML from silently overwriting clean server data.

**Files:**
- Modify: `apps/web-admin/src/lib/client/bank-soal-offline.ts`
- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`

**Design:**

Add helper in offline module:

```ts
export function syncItemContainsTableHtml(item: BankSoalQuestionSyncItem): boolean;
```

Possible behavior in `syncBankSoalOfflineQueue()`:

- If item payload contains `<table`, do not auto-sync immediately.
- Mark as `needs_review` or return it as blocked/pending requiring user action.
- Show toast/banner in composer:
  - “Ada sinkronisasi offline lama berisi tabel. Buka dan periksa sebelum dikirim agar tidak menimpa data server.”

Minimal safer variant:

- For existing question update (`method === 'PUT'` or item has `id`) and payload has `<table`, skip auto-sync and show warning.
- New offline draft creation may still sync if no server baseline exists.

Need inspect actual `BankSoalQuestionSyncItem` shape before implementation.

**Acceptance Criteria:**

- Offline queue item with table and existing question is not silently PUT synced.
- User sees warning/actionable message.
- Non-table offline items still sync as before.
- No data loss: skipped queue remains available unless user explicitly deletes.

**Verification:**

- Unit/manual test by creating mock IndexedDB queue item if feasible.
- Browser console or component smoke to confirm no uncaught errors.

Commands:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
git diff --check
```

**Commit Message:**

```bash
git add apps/web-admin/src/lib/client/bank-soal-offline.ts apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte
git commit -m "fix(bank-soal): require review before syncing table offline edits"
```

---

### Task 6: Backend Sanitizer Regression Tests for Tables

**Objective:** Prove backend sanitizer preserves basic table structure and document allowed/removed attrs.

**Files:**
- Modify/create test: `services/core-api/internal/service/cbt_question_authoring_html_extra_test.go` or existing sanitizer test file.

**Tests:**

Add cases for:

1. Simple `table > tbody > tr > td` survives sanitization.
2. `<th>` behavior is documented: either survives or transformed/removed according to current policy.
3. `colspan`/`rowspan` current behavior is documented.
4. Dangerous payload inside table is removed:
   - `<script>`
   - `style="background:url(javascript:...)"`
   - event handler attrs.

**Verification:**

```bash
cd services/core-api
go test ./internal/service -run 'HTML|Sanitize|Question' -count=1
go test ./internal/service -count=1
```

If only tests changed and pass, no backend deploy required.

**Commit Message:**

```bash
git add services/core-api/internal/service/<test-file>.go
git commit -m "test(bank-soal): cover table html sanitizer behavior"
```

---

### Task 7: End-to-End Manual Smoke Plan

**Objective:** Verify the complete user workflow before deployment.

**Files:**
- Optional local report: `reports/bank-soal-table-audit/YYYYMMDD-HHMMSS-e2e-smoke.md`

**Manual Scenarios:**

1. Create a new draft question with a 3x3 table.
2. Save draft.
3. Reopen edit/revisi.
4. Confirm table displays with horizontal scroll, not squeezed.
5. Edit text inside a cell.
6. Save.
7. Query DB to confirm `<table>`, `<tr>`, `<td>` still present.
8. Simulate broken HTML by dev-only manipulation if safe; confirm save guard blocks.
9. Create local draft conflict; confirm dialog behavior.
10. Simulate/inspect offline queue behavior; confirm table queue does not auto-overwrite.

**Protected Route Smoke:**

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8021/bank-soal/daftar
# Expected unauthenticated: 302
```

**Full Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
git diff --check

cd services/core-api
go test ./...
go build -o /tmp/core-api-banksoal-table-check ./cmd/api
```

---

### Task 8: Integration Review and Final Commit/Deploy Decision

**Objective:** Validate all tasks together, then ask user before deployment.

**Steps:**

1. Parent agent runs:

```bash
git status --short
git diff --stat
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
git diff --check
```

2. If Go changed:

```bash
cd services/core-api
go test ./...
go build -o /tmp/core-api-banksoal-table-check ./cmd/api
```

3. Dispatch final integration reviewer:
   - Check if table guard blocks only dangerous cases.
   - Check no accidental full server-only migration.
   - Check no credentials/log leaks.
   - Check UX copy is clear for guru/operator.

4. If all passes, create final focused commit if earlier tasks were not committed separately.

5. Ask user before deploy:

> “Patch sudah siap dan tervalidasi. Apakah deploy/restart web-admin sekarang?”

6. If approved and only frontend changed:

```bash
pm2 restart mtsn2kolut-web-admin --update-env
curl -s -o /dev/null -w "%{http_code}" http://localhost:8021/bank-soal/daftar
```

Expected protected route: `302`.

7. If backend changed and deploy is approved:
   - Build service binary according to project runbook.
   - Restart `mtsn2kolut-core-api` first.
   - Health check `http://127.0.0.1:8080/health`.
   - Restart web-admin if frontend changed.

---

## Open Questions Before Implementation

1. Should table save guard have an admin override? Recommendation: **No override in first patch**.
2. Should `Muat Data Server Saja` delete local draft by default? Recommendation: keep separate wording but immediately overwrite local snapshot with server data to avoid stale draft returning.
3. For offline queue table payloads, should we skip sync or prompt user? Recommendation: **skip auto-sync and show warning** for existing-question updates with tables.
4. Should reports under `reports/` be committed? Recommendation: no, unless user explicitly wants artifacts saved in repo.

---

## Risk Assessment

### Low Risk

- Pure table analyzer utility.
- Backend sanitizer tests.
- DB read-only audit.

### Medium Risk

- Save guard could false-positive block legitimate table edits.
- Offline queue skip could delay legitimate offline work.
- Local draft cleanup behavior could surprise users if copy is unclear.

### High Risk / Deferred

- Full Quill → Tiptap migration.
- Full server-only composer.
- Schema changes for structured table blocks.

---

## Recommended Agent Execution Order

1. Parent: Pre-flight status and DB audit.
2. Agent A: backend sanitizer tests and DB audit report.
3. Agent B: table analyzer utility and tests.
4. Spec review B.
5. Quality review B.
6. Agent C: server baseline + save guard in `SoalWorkspacePage.svelte`.
7. Spec review C.
8. Quality review C.
9. Agent D: local draft cleanup and offline queue guard.
10. Spec review D.
11. Quality review D.
12. Parent: full validation gate.
13. Final integration reviewer.
14. Commit/report; deploy only after user approval.

---

## Success Criteria

- Existing clean server table is not overwritten by broken local/editor state.
- Revisi/edit with valid table can still save normal text edits.
- Broken/flattened table HTML is blocked before save.
- Old offline queue table payload cannot silently overwrite server data.
- Guru receives clear Indonesian guidance when save is blocked.
- Frontend build/check passes.
- Backend tests pass if backend tests are added.
- Production restart only after explicit approval.
