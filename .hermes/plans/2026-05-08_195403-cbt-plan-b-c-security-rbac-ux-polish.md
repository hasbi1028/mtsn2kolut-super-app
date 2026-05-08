# CBT Plan B+C — Security/RBAC Hardening and UX Polish Implementation Plan

> **For Hermes:** Use subagent-driven-development or Codex delegation to implement this plan task-by-task with validation and independent review before commit/deploy.

**Goal:** Harden CBT security/RBAC boundaries and polish the CBT user experience before manual device testing and final production acceptance.

**Architecture:** Keep the existing monorepo runtime: Web Admin (`apps/web-admin`) for Bank Soal/Asesmen/proctor/results workflows, Core API (`services/core-api`) as source of truth for auth/RBAC/exam/session/scoring/audit, Flutter (`apps/mobile`) as the student exam portal, and PostgreSQL owned only by Core API. This plan must not introduce public `/api/cbt/**`, PocketBase, SQLite, Alpine runtime, live DB writes outside migrations, or final `production_go=true` claims.

**Tech Stack:** SvelteKit 2/Svelte 5 + Tailwind/shadcn tokens, Go/Chi/sqlc/PostgreSQL, Flutter, PM2 deploy smoke checks.

---

## Scope Summary

### Plan B — Security/RBAC Hardening

Focus areas:
- Role/access consistency for CBT pages and BFF routes.
- Backend authorization parity for Bank Soal, Asesmen, Proctoring, Hasil, report/evidence exports.
- Token/device/session boundary review for `/api/exam/*`.
- Evidence/export redaction and CSV injection prevention.
- Regression tests for unauthenticated/protected route behavior.

### Plan C — UX Polish

Focus areas:
- Flutter student exam usability: clearer states, submit guard, offline/pending answers, unsupported question messaging.
- Web Admin proctor dashboard UX: evidence summary, stale/heartbeat warnings, action affordances.
- Results/report UX: per-type accuracy, print/CSV guidance, pending/deferred labels.
- Operator-friendly copy and consistent Indonesian terminology.
- Dark-mode/token compatibility in internal Web Admin CBT surfaces.

---

## Non-Negotiable Boundaries

- Do not create public SvelteKit route tree `/api/cbt/**`.
- Flutter stays on Core API `/api/exam/*`.
- `services/core-api` remains the only PostgreSQL owner.
- No ad hoc SQL. Schema changes must be migrations only.
- No PocketBase/SQLite/Alpine runtime.
- No fake kiosk/screen preview/remote desktop claim.
- No real-device PASS claim without physical Android evidence.
- No `production_go=true` until manual device matrix, operator rehearsal, and sign-off are complete.
- Do not expose exam tokens, passwords, env values, API keys, or secrets in evidence/export/UI.

---

## Phase B1 — RBAC/API Surface Inventory

**Objective:** Build a verified map of CBT-facing pages, BFF routes, backend endpoints, and required permissions.

**Files:**
- Create: `docs/cbt-security-rbac-hardening-plan.md`
- Modify/Test: `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts`

**Steps:**
1. Inventory Web Admin pages:
   - `/bank-soal/*`
   - `/asesmen/*`
   - `/settings/rbac`
   - evidence/report export pages.
2. Inventory BFF routes under `apps/web-admin/src/routes/api/**` related to Bank Soal/Asesmen/RBAC/proctoring/results.
3. Inventory Core API native endpoints consumed by CBT flows.
4. Produce a permission matrix with columns:
   - Area
   - Frontend route
   - BFF route
   - Backend endpoint
   - Required permission/role
   - Expected unauthenticated status (`302` for pages, `401` for APIs)
   - Evidence redaction requirement.
5. Add docs guard test to lock this matrix exists and states no public `/api/cbt/**` route tree.

**Verification:**
```bash
cd apps/web-admin
npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
npm run check
```

---

## Phase B2 — Protected Route Smoke Matrix

**Objective:** Add repeatable smoke checks for unauthenticated/protected CBT routes.

**Files:**
- Modify: `deploy/scripts/health-check.sh` or add a focused helper under `deploy/scripts/`
- Create: `docs/cbt-security-smoke-matrix.md`
- Test: web-admin docs/route tests as applicable.

**Steps:**
1. Define expected unauthenticated smoke behavior:
   - Web pages: `302` redirect to login.
   - BFF APIs: `401`.
   - Native backend protected APIs: `401`.
2. Include key routes:
   - `/bank-soal`, `/bank-soal/tambah`, `/bank-soal/verifikasi`, `/bank-soal/impor`, `/bank-soal/analisis-butir`
   - `/asesmen`, `/asesmen/sesi`, `/asesmen/hasil`, `/asesmen/pengawasan`
   - `/api/bank-soal/summary`
   - `/api/asesmen/*` where safe to smoke without IDs.
   - `/api/exam/status` expected `401` without token.
3. Ensure smoke output redacts tokens and headers.
4. Document expected results and rollback if a route returns `404`/`500`.

**Verification:**
```bash
bash deploy/scripts/health-check.sh all
# plus new focused smoke command if added
```

---

## Phase B3 — Evidence Export Redaction and CSV Safety

**Objective:** Harden evidence/report exports against token leakage and CSV injection.

**Files:**
- Inspect/Modify: proctor evidence helpers under `apps/web-admin/src/lib/cbt/`
- Inspect/Modify: session/proctor/result pages under `apps/web-admin/src/routes/asesmen/**`
- Tests: existing or new `*.test.ts` near export helpers.

**Steps:**
1. Locate evidence CSV/export helpers.
2. Add/verify redaction for:
   - exam token
   - bearer/session token
   - password/secret/api key
   - raw device fingerprint if too detailed.
3. Add CSV injection guard for cells starting with `=`, `+`, `-`, `@`, tab, CR/LF.
4. Add tests with malicious values:
   - `=HYPERLINK(...)`
   - `+SUM(1,1)`
   - raw token-like strings.
5. Keep exports role-bound and token-free.

**Verification:**
```bash
cd apps/web-admin
npm run test:unit -- src/lib/cbt/proctor-evidence.test.ts
npm run test:unit
npm run check
```

---

## Phase B4 — Backend Exam Token/Device Boundary Review

**Objective:** Verify `/api/exam/*` token/device/session handling remains fail-closed.

**Files:**
- Inspect/Modify tests in `services/core-api/internal/handler` and `services/core-api/internal/service`.
- Optional docs update: `docs/exam-api.md`.

**Steps:**
1. Confirm unauthenticated `/api/exam/status` returns `401`.
2. Confirm token/device mismatch returns controlled `409` where expected.
3. Confirm stale/finished/submitted sessions fail closed for answer/submit.
4. Confirm answer size/body validation remains enforced.
5. Add missing negative tests only if a gap is found.
6. Do not change live data.

**Verification:**
```bash
cd services/core-api
GOCACHE=/tmp/go-build /home/servermtsn2kolut/.local/go/bin/go test ./...
GOCACHE=/tmp/go-build /home/servermtsn2kolut/.local/go/bin/go build -o /tmp/mtsn2kolut-core-api-plan-b ./cmd/api
```

---

## Phase C1 — Flutter Exam UX Polish

**Objective:** Make student-facing exam states clearer without changing backend contracts.

**Files:**
- Modify: `apps/mobile/lib/src/screens/exam_shell_screen.dart`
- Modify: `apps/mobile/lib/src/screens/exam_login_screen.dart` if login copy needs improvement.
- Tests: `apps/mobile/test/widget_test.dart` and related tests.

**Steps:**
1. Improve Indonesian copy for:
   - pending local answer
   - reconnect/resume gate
   - stale heartbeat
   - submit guard
   - unsupported hotspot/upload/file types.
2. Add clear color/icon/label hierarchy for:
   - Aman
   - Perlu sinkron
   - Perlu pengawas
   - Tidak didukung aplikasi siswa.
3. Keep answer serialization unchanged.
4. Add widget tests for changed labels/states.

**Verification:**
```bash
cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
```

---

## Phase C2 — Proctor Dashboard UX Polish

**Objective:** Make proctoring evidence easier to interpret during rehearsal.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- Modify helpers/tests in `apps/web-admin/src/lib/cbt/`.

**Steps:**
1. Add a compact status summary for:
   - online/heartbeat OK
   - stale connection
   - app background/resume events
   - device mismatch
   - submitted/force submitted.
2. Make evidence export button copy explicit: `Export Evidence CSV (tanpa token)`.
3. Add operator guidance panel:
   - when to warn student
   - when to reset access
   - when to force submit.
4. Preserve existing proctor actions and backend calls.
5. Add tests for evidence helper labels.

**Verification:**
```bash
cd apps/web-admin
npm run test:unit -- src/lib/cbt/proctor-evidence.test.ts
npm run check
```

---

## Phase C3 — Results and Report UX Polish

**Objective:** Improve hasil/laporan interpretation without fake analytics.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte` or relevant hasil pages.
- Modify: `apps/web-admin/src/lib/cbt/item-analysis-evidence.ts`
- Modify: `apps/web-admin/src/lib/cbt/report-template-matrix.ts`
- Tests: `item-analysis-evidence.test.ts`, `report-template-matrix.test.ts`.

**Steps:**
1. Make per-type accuracy labels clearer.
2. Clearly label deferred metrics such as Cronbach alpha.
3. Keep PDF parity as print/browser PDF and Excel parity as safe CSV/JSON.
4. Add empty-state copy for sessions without enough responses.
5. Add/adjust tests for labels and deferred statuses.

**Verification:**
```bash
cd apps/web-admin
npm run test:unit -- src/lib/cbt/item-analysis-evidence.test.ts src/lib/cbt/report-template-matrix.test.ts
npm run check
```

---

## Phase C4 — Internal Dark-Mode Token Sweep for CBT Surfaces

**Objective:** Ensure CBT internal surfaces follow theme tokens and remain readable in dark mode.

**Files:**
- Inspect/Modify CBT-related Svelte pages/components under:
  - `apps/web-admin/src/routes/bank-soal/**`
  - `apps/web-admin/src/routes/asesmen/**`
  - `apps/web-admin/src/lib/cbt/**`

**Steps:**
1. Search for light-only classes in CBT/internal pages:
   - `bg-white`
   - `text-slate-*`
   - `border-slate-*`
   - `bg-slate-*`
   - raw hex colors.
2. Replace with tokens where appropriate:
   - `bg-background`, `bg-card`, `text-foreground`, `text-muted-foreground`, `border-border`, `bg-muted`, `text-primary`, `text-warning`.
3. Preserve print-only and intentionally colored status badges where they already have dark variants.
4. Run check/build.

**Verification:**
```bash
cd apps/web-admin
npm run check
npm run test:unit
npm run build
```

---

## Final Validation Gate

Run before commit:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git diff --check

cd apps/web-admin
npm run test:unit
npm run check
npm run build

cd ../../services/core-api
GOCACHE=/tmp/go-build /home/servermtsn2kolut/.local/go/bin/go test ./...
GOCACHE=/tmp/go-build /home/servermtsn2kolut/.local/go/bin/go build -o /tmp/mtsn2kolut-core-api-plan-b-c ./cmd/api

cd ../../apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test

cd ../..
make ops-health
```

Security hygiene before commit:

```bash
python3 - <<'PY'
import re, subprocess, sys
patterns=[
  re.compile(r"(api_key|secret|password|token|passwd)\\s*=\\s*['\"][^'\"]{6,}['\"]", re.I),
  re.compile(r"POST /api/cbt/login|GET /api/cbt/status"),
]
diff=subprocess.check_output(['git','diff'], text=True, errors='replace')
issues=[]
for line in diff.splitlines():
  if line.startswith('+') and not line.startswith('+++'):
    if any(p.search(line) for p in patterns):
      issues.append(line)
if issues:
  print('\n'.join(issues))
  sys.exit(1)
print('security-diff-scan=PASS')
PY
```

Independent review:
- Use `requesting-code-review` skill with a fresh subagent.
- Block commit on security/logic blockers.

---

## Deploy Decision

This plan should be committed first. Deploy/restart PM2 only after explicit approval unless the user asks for end-to-end deploy.

If deploying after implementation:
1. `make ops-backup`
2. `make build-backend build-web build-worker`
3. `make db-migrate` if migrations exist/reconciled
4. PM2 restart backend/web/worker with `--update-env`
5. `pm2 save`
6. Smoke:
   - `bash deploy/scripts/health-check.sh all`
   - backend `/health` = `200`
   - web root = `200`
   - protected pages = `302`
   - protected APIs = `401`

---

## Acceptance Criteria

Plan B is complete when:
- Security/RBAC matrix is documented and guard-tested.
- Protected route smoke expectations are documented/repeatable.
- Evidence/report exports are token-free and CSV-safe.
- `/api/exam/*` token/device/session boundaries are covered by tests or documented existing tests.

Plan C is complete when:
- Flutter exam UX states are clearer and tested.
- Proctor dashboard evidence UX is clearer and tested.
- Results/report UX labels are clearer without fake analytics.
- CBT internal surfaces pass check/test and avoid obvious light-only dark-mode regressions.

Final state remains:
- `production_go=false` until manual Android device evidence, operator rehearsal, and explicit sign-off are complete.
