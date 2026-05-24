# Opsi 0 Asesmen/CBT Playwright E2E Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menambahkan **Opsi 0** sebagai jalur uji end-to-end Playwright/smoke untuk memastikan alur Asesmen/CBT aman sebelum Opsi A/B/C simulasi siswa.

**Architecture:** Opsi 0 tidak mengubah data produksi. Jalur awal mengikuti pola repo saat ini: browser smoke script `apps/web-admin/scripts/*.mjs` yang optional-skip bila env/Playwright belum tersedia, lalu bisa dinaikkan ke Playwright Test formal setelah fixture stabil. Runtime ujian siswa diuji melalui BFF `/api/exam/*` karena UI ujian utama ada di mobile app, sementara web-admin menguji admin/proktor/panitia.

**Tech Stack:** SvelteKit web-admin, Go core-api, PostgreSQL isolated test DB, Node.js browser smoke with Playwright runtime (`playwright` or `@playwright/test`), Vitest contract tests.

---

## Executive Summary

Opsi 0 adalah **gladi teknis otomatis** sebelum Opsi A/B/C:

- **Tidak memakai data produksi**.
- **Tidak mengandalkan soal/paket/sesi Informatika yang belum siap**.
- **Tidak langsung full Playwright Test runner**, karena repo belum punya `playwright.config.*` dan pola existing adalah script smoke manual.
- **Memvalidasi hal paling kritis:** login role, route Asesmen, paket/sesi/command center reachable, exam proxy, token masking, answer-key redaction, dan response security.

### Kenapa Opsi 0 perlu ada

Review sebelumnya menunjukkan modul aplikasi cukup baik, tetapi data simulasi Informatika belum siap. Opsi 0 menjawab gap itu dengan membuat jalur test otomatis yang bisa dijalankan di DB test terisolasi sebelum siswa masuk.

---

## Findings dari Agent Swarm

### Existing test/browser infra

Repo belum memiliki:

- `playwright.config.ts/js`
- folder `e2e/`
- dependency formal `@playwright/test`

Repo sudah memiliki pola smoke browser manual:

- `apps/web-admin/scripts/cbt-role-smoke.mjs`
- `apps/web-admin/scripts/bank-soal-role-e2e.mjs`

Pola existing:

- dynamic import `playwright`, fallback `@playwright/test`
- skip bersih jika Playwright/env tidak tersedia
- env credential tidak dicommit
- role smoke via login UI `/login?from=...`

### Existing data/test seed

Script penting:

- `scripts/setup-cbt-test-db.sh`
- `services/core-api/db/scripts/seed_mobile_cbt.js`
- `services/core-api/db/scripts/seed_bank_soal_role_e2e.js`

Rekomendasi swarm: gunakan DB test terisolasi dengan `TEST_DATABASE_URL`, bukan production DB.

### Existing routes yang diuji

Admin/proktor/operator:

- `/asesmen`
- `/asesmen/persiapan`
- `/asesmen/pelaksanaan`
- `/asesmen/pengawasan`
- `/asesmen/hasil`
- `/asesmen/kegiatan`
- `/asesmen/paket`
- `/asesmen/sesi`
- `/ujian/command-center`

Exam API via BFF:

- `POST /api/exam/login`
- `GET /api/exam/status`
- `POST /api/exam/heartbeat`
- `POST /api/exam/event`
- `POST /api/exam/answer` *(optional destructive; only on resettable DB test)*
- `POST /api/exam/submit` *(optional destructive; only on resettable DB test)*

---

## Opsi 0 Scope

### Opsi 0.1 — Safe Smoke, non-destructive

Wajib sebagai tahap pertama.

Menguji:

1. Login admin/proktor/guru seed.
2. Route Asesmen utama reachable/forbidden sesuai role.
3. Protected API unauthenticated return `401`/redirect login.
4. `/api/exam/*` tanpa token/fingerprint ditolak.
5. `/api/exam/login` dengan token seed berhasil.
6. Response `/api/exam/login` tidak mengandung answer key/correct marker.
7. Response exam proxy memakai `cache-control: no-store`.
8. Student portal schedule hanya menampilkan token masked sebelum reveal.
9. Room token salah tidak membocorkan token peserta.

Tidak melakukan:

- submit jawaban final,
- scoring,
- mutasi data produksi,
- regenerate token produksi.

### Opsi 0.2 — Full Resettable E2E, destructive-safe di DB test

Dilakukan hanya bila memakai DB test yang resettable.

Menguji:

1. Login exam siswa seed.
2. Ambil status/daftar soal.
3. Kirim heartbeat.
4. Catat event proctoring ringan.
5. Jawab 1 PG dan 1 Essay.
6. Submit.
7. Submit ulang harus `409`.
8. Admin/proktor membuka dashboard/hitung skor/hasil.

---

## File yang akan dibuat/diubah

### New files

- Create: `apps/web-admin/scripts/asesmen-cbt-opsi0-e2e.mjs`
- Create: `apps/web-admin/docs/asesmen-cbt-opsi0-e2e.md`
- Create: `apps/web-admin/src/lib/cbt/asesmen-cbt-opsi0-e2e-coverage.test.ts`

### Modify

- Modify: `apps/web-admin/package.json`
  - Add script: `smoke:asesmen-cbt:opsi0`
- Modify: root `package.json`
  - Add script: `smoke:web:asesmen-cbt:opsi0`

### Optional next step files

Only after Opsi 0.1 passes:

- Create: `services/core-api/db/scripts/seed_asesmen_cbt_opsi0_e2e.js`
- Create: `services/core-api/db/scripts/seed_asesmen_cbt_opsi0_e2e.test.js`
- Modify: `services/core-api/db/scripts/package.json`
- Modify: root `package.json`

---

## Environment Contract

Use these env vars. Do not commit actual values.

```bash
WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL=http://127.0.0.1:8021
WEB_ADMIN_ASESMEN_CBT_OPSI0_ADMIN_USERNAME='[REDACTED_EXAMPLE_USERNAME]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_ADMIN_PASSWORD='[REDACTED_EXAMPLE_PASSWORD]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_PROCTOR_USERNAME='[REDACTED_EXAMPLE_USERNAME]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_PROCTOR_PASSWORD='[REDACTED_EXAMPLE_PASSWORD]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_GURU_USERNAME='[REDACTED_EXAMPLE_USERNAME]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_GURU_PASSWORD='[REDACTED_EXAMPLE_PASSWORD]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_EXAM_TOKEN='[REDACTED_EXAMPLE_EXAM_TOKEN]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_ROOM_TOKEN='[REDACTED_EXAMPLE_ROOM_TOKEN]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_BAD_ROOM_TOKEN='[REDACTED_EXAMPLE_BAD_ROOM_TOKEN]'
WEB_ADMIN_ASESMEN_CBT_OPSI0_SESSION_ID=optional
WEB_ADMIN_ASESMEN_CBT_OPSI0_EVENT_ID=optional
WEB_ADMIN_ASESMEN_CBT_OPSI0_HEADLESS=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_TIMEOUT_MS=20000
WEB_ADMIN_ASESMEN_CBT_OPSI0_IGNORE_HTTPS_ERRORS=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_ALLOW_SKIP=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=false
```

Rules:

- If env is incomplete and `ALLOW_SKIP=true`, script exits `0` with SKIP message.
- If env is incomplete and `ALLOW_SKIP=false`, script exits non-zero.
- Strict mode (`ALLOW_SKIP=false`) requires complete safe-smoke env for admin, proctor, guru, `EXAM_TOKEN`, and `ROOM_TOKEN`.
- If `DESTRUCTIVE=false`, do not call `/answer` or `/submit`.
- If `DESTRUCTIVE=true`, require `WEB_ADMIN_ASESMEN_CBT_OPSI0_CONFIRM_TEST_DB=true`.

---

## Data Safety Gate

Before running Opsi 0.2:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
TEST_DATABASE_URL='[REDACTED_TEST_DATABASE_URL]' scripts/setup-cbt-test-db.sh --reset
```

Never run destructive E2E against production DB.

The script must refuse destructive mode unless:

```bash
WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_CONFIRM_TEST_DB=true
```

---

# Implementation Tasks

## Task 1: Add coverage contract test for Opsi 0 script/docs

**Objective:** Ensure the smoke script and docs remain discoverable in CI/unit tests.

**Files:**

- Create: `apps/web-admin/src/lib/cbt/asesmen-cbt-opsi0-e2e-coverage.test.ts`

**Step 1: Write failing test**

```ts
import { describe, expect, it } from 'vitest';
import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const root = resolve(process.cwd(), '..', '..');
const scriptPath = resolve(process.cwd(), 'scripts/asesmen-cbt-opsi0-e2e.mjs');
const docsPath = resolve(process.cwd(), 'docs/asesmen-cbt-opsi0-e2e.md');
const pkgPath = resolve(process.cwd(), 'package.json');

describe('Asesmen CBT Opsi 0 E2E smoke coverage contract', () => {
  it('keeps the Opsi 0 smoke script, docs, and npm script discoverable', () => {
    expect(existsSync(scriptPath)).toBe(true);
    expect(existsSync(docsPath)).toBe(true);

    const script = readFileSync(scriptPath, 'utf8');
    expect(script).toContain('WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL');
    expect(script).toContain('/api/exam/login');
    expect(script).toContain('assertNoSensitiveExamKeys');
    expect(script).toContain('WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE');

    const docs = readFileSync(docsPath, 'utf8');
    expect(docs).toContain('Opsi 0');
    expect(docs).toContain('TEST_DATABASE_URL');
    expect(docs).toContain('tidak memakai data produksi');

    const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'));
    expect(pkg.scripts['smoke:asesmen-cbt:opsi0']).toBe('node scripts/asesmen-cbt-opsi0-e2e.mjs');
  });
});
```

**Step 2: Run RED**

```bash
cd apps/web-admin
npm run test:unit -- --run src/lib/cbt/asesmen-cbt-opsi0-e2e-coverage.test.ts
```

Expected: FAIL because files/scripts do not exist yet.

---

## Task 2: Create the Opsi 0 smoke script skeleton

**Objective:** Add a safe Playwright smoke runner with env validation, skip behavior, login helper, and security deep-scan helpers.

**Files:**

- Create: `apps/web-admin/scripts/asesmen-cbt-opsi0-e2e.mjs`

**Implementation outline:**

- Dynamic import `playwright`, fallback `@playwright/test`.
- `env()` helper.
- `requireEnv()` helper honoring `ALLOW_SKIP`.
- `loginAs(page, username, password, fromPath)` helper copied/adapted from existing smoke scripts.
- `expectReachable(page, path, expectedTexts)` helper.
- `expectForbidden(page, path)` helper.
- `apiRequest(context, method, path, body, headers)` helper.
- `assertNoSensitiveExamKeys(value)` deep scanner.
- `assertNoRawTokenInText(text, token)` helper.
- `main()` with two phases:
  - role route smoke,
  - exam proxy safe smoke.

**Sensitive keys scanner must reject keys matching:**

```js
/answer[-_ ]?key|correct[-_ ]?answers?|kunci[-_ ]?jawaban|is_correct|correct/i
```

**Run GREEN:**

```bash
cd apps/web-admin
node scripts/asesmen-cbt-opsi0-e2e.mjs
```

Expected with no env: SKIP and exit 0 if `ALLOW_SKIP=true` default.

---

## Task 3: Add docs for running Opsi 0

**Objective:** Provide operator/developer runbook.

**Files:**

- Create: `apps/web-admin/docs/asesmen-cbt-opsi0-e2e.md`

**Docs must include:**

- Purpose: Opsi 0 before Opsi A/B/C.
- Data safety: use `TEST_DATABASE_URL`, never production DB.
- Env variables.
- Commands:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run smoke:asesmen-cbt:opsi0
```

- Local DB test flow:

```bash
TEST_DATABASE_URL='[REDACTED_TEST_DATABASE_URL]' scripts/setup-cbt-test-db.sh --reset
```

- What is tested in safe mode vs destructive mode.
- Troubleshooting Playwright dependency missing.
- Expected output examples.

---

## Task 4: Wire npm scripts

**Objective:** Make Opsi 0 runnable consistently.

**Files:**

- Modify: `apps/web-admin/package.json`
- Modify: root `package.json`

**Add scripts:**

In `apps/web-admin/package.json`:

```json
"smoke:asesmen-cbt:opsi0": "node scripts/asesmen-cbt-opsi0-e2e.mjs"
```

In root `package.json`:

```json
"smoke:web:asesmen-cbt:opsi0": "npm --prefix apps/web-admin run smoke:asesmen-cbt:opsi0"
```

**Run:**

```bash
npm --prefix apps/web-admin run smoke:asesmen-cbt:opsi0
npm run smoke:web:asesmen-cbt:opsi0
```

Expected with missing env: SKIP exit 0.

---

## Task 5: Implement role route smoke checks

**Objective:** Validate operator/admin/proctor/guru route access from real browser session.

**Files:**

- Modify: `apps/web-admin/scripts/asesmen-cbt-opsi0-e2e.mjs`

**Checks:**

Admin/proctor expected reachable:

- `/asesmen`
- `/asesmen/persiapan`
- `/asesmen/pelaksanaan`
- `/asesmen/pengawasan`
- `/ujian/command-center`

Result role expected reachable if credential provided:

- `/asesmen/hasil`

Guru/plain expected denied if no proctor/result permissions:

- `/asesmen/pelaksanaan`
- `/asesmen/pengawasan`
- `/asesmen/hasil`

**Verification:**

Run with seeded credentials. Expected: all route checks PASS.

---

## Task 6: Implement exam proxy safe smoke

**Objective:** Validate exam runtime API security without submitting answers.

**Files:**

- Modify: `apps/web-admin/scripts/asesmen-cbt-opsi0-e2e.mjs`

**Checks:**

1. `GET /api/exam/status` without token returns not-200.
2. `POST /api/exam/login` with seed exam token + room token succeeds.
3. Login response deep-scan has no answer key/correct marker.
4. Login response has `cache-control: no-store`.
5. `GET /api/exam/status` with returned `X-Exam-Token` and fingerprint succeeds.
6. `POST /api/exam/heartbeat` succeeds.
7. Bad fingerprint/status returns not-200.
8. Oversized body to `/api/exam/event` returns `413` or non-success without upstream success.

**Non-destructive only:** do not call `/answer` or `/submit` unless destructive mode is explicitly enabled.

---

## Task 7: Implement optional destructive resettable mode

**Objective:** Add full end-to-end answer/submit path guarded by explicit test DB confirmation.

**Files:**

- Modify: `apps/web-admin/scripts/asesmen-cbt-opsi0-e2e.mjs`
- Optional Create/Modify seed files if existing `seed_mobile_cbt.js` is insufficient.

**Gate:**

Script refuses unless:

```bash
WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE=true
WEB_ADMIN_ASESMEN_CBT_OPSI0_CONFIRM_TEST_DB=true
```

**Checks:**

1. Login exam.
2. Pick first MC question from status/login payload.
3. `POST /api/exam/answer` returns recorded.
4. `POST /api/exam/submit` returns submitted.
5. Second submit returns `409`.
6. Admin route `/asesmen/sesi/$SESSION_ID` shows submitted count or results section reachable.

---

## Task 8: Final verification gate

**Objective:** Verify no regressions and script discoverability.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run test:unit -- --run src/lib/cbt/asesmen-cbt-opsi0-e2e-coverage.test.ts src/routes/api/exam/exam-proxy.test.ts src/lib/server/route-access.test.ts
npm run check
npm run build
```

If Go/seed changes were made:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
npm --prefix db/scripts test -- seed_asesmen_cbt_opsi0_e2e.test.js
cd /home/servermtsn2kolut/mtsn2kolut-super-app
```

Smoke script no-env behavior:

```bash
npm --prefix apps/web-admin run smoke:asesmen-cbt:opsi0
```

Expected: SKIP exit 0 if env missing.

---

## Acceptance Criteria

Opsi 0 is accepted when:

- `smoke:asesmen-cbt:opsi0` exists in web-admin package.
- Root alias exists.
- Docs explain DB safety and env contract.
- Contract unit test passes.
- Script skips safely without env.
- With E2E env on test DB, safe mode validates route access, exam proxy login/status/heartbeat, token redaction, and answer-key redaction.
- Destructive answer/submit mode is impossible without explicit two-env confirmation.
- `npm run check` and targeted Vitest pass.

---

## Recommended Execution Approach

Use subagent-driven-development with file ownership:

1. Agent A: contract test + docs.
2. Agent B: script skeleton + env/skip/login helpers.
3. Agent C: route access smoke checks.
4. Agent D: exam proxy safe checks + sensitive scanner.
5. Agent E: review-only security/spec compliance.

Parent integrates, runs full verification, then asks before deploy/restart. No production data mutation.
