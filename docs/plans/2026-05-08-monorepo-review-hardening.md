# Monorepo Review Hardening Implementation Plan

> **For Hermes:** Use the mtsn2kolut-super-app conventions and validate each security-sensitive route/path change before commit.

**Goal:** Fix the review blockers before committing the remaining dirty monorepo changes: preserve non-test assessment RBAC, keep CBT asset file delivery working, make mobile essay limit match backend byte budget, and keep old backup script path compatible.

**Architecture:** Keep active BFF/public API names as `/api/bank-soal/*` and `/api/asesmen/*`, but verify each alias has equivalent backend guard/route behavior. Prefer additive compatibility wrappers/routes over deletion because this branch already has mixed staged changes. Tests must lock path dispatch and route contracts so this regression does not return.

**Tech Stack:** SvelteKit/TypeScript BFF, Go Chi backend, Flutter mobile, Bash ops scripts.

---

## Phase 1 — Lock web-admin CBT backend path dispatch

### Task 1: Add path-dispatch regression tests

**Objective:** Prove BFF helper routes Bank Soal, Asesmen, non-test assessment, and asset file paths to the intended backend namespaces.

**Files:**
- Create: `apps/web-admin/src/lib/server/cbt-backend-paths.test.ts`
- Modify: `apps/web-admin/src/lib/server/cbt-backend-paths.ts`

**Test cases:**
- `cbtBackendPath('/questions')` => `/api/bank-soal/questions`
- `cbtBackendPath('/questions/123')` => `/api/bank-soal/questions/123`
- `cbtBackendPath('/assets')` => `/api/bank-soal/assets`
- `cbtBackendPath('/assets/abc/file')` => `/api/bank-soal/assets/abc/file` only if backend alias is added in Phase 2
- `cbtBackendPath('/non-test-assessments')` => `/api/asesmen/non-test-assessments` only after backend RBAC is matched in Phase 2
- `cbtBackendPath('/sessions')` => `/api/asesmen/sessions`

**Validation command:**
```bash
cd apps/web-admin
npm run test:unit -- src/lib/server/cbt-backend-paths.test.ts
```

---

## Phase 2 — Fix backend alias parity and RBAC

### Task 2: Add native Bank Soal asset file alias

**Objective:** Ensure BFF asset file URLs that now resolve to `/api/bank-soal/assets/{id}/file` do not 404.

**Files:**
- Modify: `services/core-api/cmd/api/main.go`
- Test/modify: `services/core-api/cmd/api/routes_contract_test.go` or equivalent route contract test

**Implementation:**
- Add a public/protected asset-file alias using the exact same auth middleware and handler as legacy CBT file route:
  - Existing: `/api/cbt/assets/{id}/file`
  - Add: `/api/bank-soal/assets/{id}/file`
- Use `mw.ExamTokenOrJWT(...)` and `questionAssetH.File`, same as legacy route.

**Validation command:**
```bash
cd services/core-api
go test ./cmd/api ./internal/handler ./internal/service
```

### Task 3: Match native Asesmen non-test RBAC to legacy CBT RBAC

**Objective:** Prevent `/api/asesmen/non-test-assessments*` from being broader than `/api/cbt/non-test-assessments*`.

**Files:**
- Modify: `services/core-api/cmd/api/main.go`
- Test/modify: `services/core-api/cmd/api/routes_contract_test.go` or middleware/route contract test

**Implementation:**
- Keep list/get/submissions read endpoints under `requireAsesmenRead`.
- Keep create/update/delete/generate/upsert under `requireAsesmenScore`.
- Keep sync-grade under both `requireAsesmenScore` and `requireGradesManage`, matching legacy route.
- Do not use broad `requireCbt` on native non-test mutating endpoints.

**Validation command:**
```bash
cd services/core-api
go test ./cmd/api ./internal/handler ./internal/service
```

---

## Phase 3 — Align mobile answer limit with backend byte cap

### Task 4: Add mobile guard for serialized answer size or lower essay cap

**Objective:** Avoid students writing an essay that the backend rejects due to 64 KB `MaxBytesReader` JSON budget.

**Files:**
- Modify: `apps/mobile/lib/src/screens/exam_shell_screen.dart`
- Test/modify: `apps/mobile/test/exam_*_test.dart` if test harness supports the affected helper

**Implementation preference:**
- Prefer an explicit UTF-8/serialized payload size helper if submit code is easy to isolate.
- If not, lower `kEssayAnswerMaxChars` from `32000` to a conservative `12000`–`16000` and update UX copy/comment.
- Keep `kShortAnswerMaxChars = 256`.

**Validation command:**
```bash
cd apps/mobile
flutter test
```

---

## Phase 4 — Preserve old backup script path

### Task 5: Restore `deploy/scripts/backup.sh` as a compatibility wrapper

**Objective:** Prevent older cron/manual jobs calling `deploy/scripts/backup.sh` from breaking after the PostgreSQL backup script rename.

**Files:**
- Restore/create: `deploy/scripts/backup.sh`
- Verify: `Makefile`, `deploy/DEPLOY.md`, `findings.md`

**Implementation:**
```bash
#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "${SCRIPT_DIR}/../backup-postgresql.sh" "$@"
```

**Validation command:**
```bash
bash -n deploy/scripts/backup.sh
bash -n deploy/backup-postgresql.sh
```

---

## Phase 5 — Validation, staging, and commits

### Task 6: Run targeted gates

**Objective:** Verify patched risk areas before broad gates.

**Commands:**
```bash
git diff --check
cd services/core-api && go test ./cmd/api ./internal/handler ./internal/service
cd apps/web-admin && npm run test:unit -- src/lib/server/cbt-backend-paths.test.ts src/lib/server/proxy-routes.test.ts src/lib/server/route-access.test.ts
cd services/pusaka-worker && npm run check
bash -n deploy/scripts/backup.sh deploy/backup-postgresql.sh
```

### Task 7: Run broader gates if targeted gates pass

**Objective:** Catch cross-module regressions before commit.

**Commands:**
```bash
cd services/core-api && go test ./...
cd apps/web-admin && npm run check
cd apps/web-admin && npm run test:unit
cd apps/mobile && flutter test
```

### Task 8: Commit split by scope

**Objective:** Keep reviewable history and avoid mixing unrelated dirty files.

**Suggested commits:**
1. `fix(api): align asesmen and bank soal backend aliases`
2. `fix(mobile): keep essay answers under backend payload limit`
3. `fix(ops): keep legacy backup script wrapper`
4. Existing security hardening/Pusaka worker/docs may be committed separately after final validation.

**Final verification before each commit:**
```bash
git diff --cached --check
git diff --cached --stat
git status --short
```
