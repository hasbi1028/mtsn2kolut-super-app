# CBT Opsi A — Login Siswa Tanpa Token Manual Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Siswa membuka portal CBT khusus di link/port terpisah, memasukkan **NISN** sebagai identitas dan **NISN** sebagai password/kode masuk simulasi, lalu sistem menampilkan ujian yang ditugaskan tanpa token peserta atau token ruang manual.

**Architecture:** Go core-api tetap menjadi owner business state dan akses CBT. Portal CBT khusus dibuat sebagai surface terpisah dari web-admin agar siswa tidak masuk ke admin shell. SvelteKit portal CBT hanya BFF/proxy dan UI minimal. Runtime `/api/exam/*` tetap dipertahankan untuk kompatibilitas, tetapi ditambah jalur launch baru berbasis NISN + assignment + policy gate server-side untuk sesi simulasi.

**Tech Stack:** Go + Chi + sqlc + PostgreSQL, SvelteKit/Svelte 5, Vitest, Playwright smoke script pattern existing.

---

## Scope Opsi A

### Yang dibangun

- Portal CBT khusus di link/port terpisah dari web-admin, misalnya `https://cbt.mtsn2kolut.sch.id` atau port internal `8031`.
- Login simulasi sederhana: siswa memasukkan **NISN** di kolom identitas dan **NISN** di kolom password/kode.
- Mode akses sesi CBT baru untuk simulasi/web: setelah validasi NISN, siswa melihat ujian aktif miliknya.
- Endpoint backend start berbasis portal CBT: `POST /api/cbt-portal/login` lalu `POST /api/cbt-portal/participants/{participantID}/start`.
- Halaman siswa **Ujian Saya** di portal CBT khusus, bukan admin dashboard.
- Tombol **Mulai Ujian** yang tidak meminta token manual.
- Policy sesi eksplisit agar fitur ini tidak aktif untuk semua ujian secara otomatis.
- Audit event untuk start attempt/success/blocked dan room-token bypass.
- Test backend/frontend + smoke safe.

### Yang tidak dibangun di Opsi A awal

- Menghapus sistem token lama.
- Menghapus aplikasi Android.
- Menghapus room token dari database.
- Hardening hash-only room token total.
- Proctor room gate official exam lengkap. Itu masuk Opsi B/ujian resmi.
- Menggunakan NISN=NISN untuk ujian resmi bernilai tinggi. Pola ini hanya untuk simulasi/gladi, kecuali nanti ditambah gate pengawas/OTP/QR.

### Default keamanan

Semua sesi existing tetap aman:

- `access_mode = 'secure_exam'`
- `student_portal_direct_login_enabled = false`
- `require_room_token_for_web = true`

Opsi A hanya aktif pada sesi yang eksplisit diaktifkan admin/panitia.

### Penyesuaian Tambahan: Portal CBT Khusus NISN-only

Requirement tambahan dari madrasah:

- Portal CBT tidak memakai link web-admin utama.
- Portal CBT berjalan sebagai surface khusus di link/port berbeda.
- Siswa login cukup dengan NISN.
- Kolom password/kode juga diisi NISN agar siswa tidak perlu mengingat password lain.
- Setelah login NISN, siswa langsung melihat ujian aktif miliknya.

Rekomendasi teknis:

1. **Pisahkan portal siswa dari web-admin.** Jangan arahkan siswa ke admin shell. Buat app/surface khusus seperti `apps/cbt-portal` atau instance SvelteKit khusus dengan route terbatas.
2. **Anggap NISN=NISN sebagai kode simulasi, bukan password keamanan umum.** Jangan menyimpan password plaintext. Backend cukup memvalidasi NISN terhadap student + participant assignment + sesi yang mengizinkan `nisn_direct_login_enabled`.
3. **Batasi hanya sesi simulasi.** Default semua sesi tetap `secure_exam`. Untuk ujian resmi, NISN-only harus ditambah pengawas/room gate/OTP/QR sebelum dipakai.
4. **Tidak membuat session web-admin umum.** Login NISN portal CBT hanya menghasilkan session/credential scoped ke CBT runtime dan participant, bukan akun admin/siswa penuh.
5. **Rate limit dan audit wajib.** Karena NISN mudah ditebak, endpoint login harus punya rate limit per IP/NISN, blocked attempts, generic error, dan audit hash.

---

## Swarm Execution Plan

Gunakan swarm bertahap agar file ownership tidak saling tabrak.

### Wave 1 — Foundation, 4 agents paralel

- **Agent A1 Backend Migration/SQL:** migration + query/sqlc.
- **Agent A2 Backend Service/Handler:** service start endpoint + tests.
- **Agent A3 CBT Portal UI/BFF:** portal khusus siswa, proxy login/start, UI tests.
- **Agent A4 Security/QA:** NISN-only risk controls, contract tests, audit assertions, docs/runbook review.

Parent integrasi setelah Wave 1:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
gofmt -w internal/handler internal/service internal/repository/postgres
MIGRATIONS_DRY_RUN=true go test ./internal/handler ./internal/service ./internal/repository/postgres

go build -o bin/api ./cmd/api

cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

### Wave 2 — Integration, 3 agents paralel

- **Agent B1 CBT runtime page:** runtime launch state, blocked reasons, exam handoff in portal khusus.
- **Agent B2 Admin/session policy UI:** admin toggle for simulation/NISN-direct-login policy.
- **Agent B3 Smoke/ops docs:** Playwright smoke script, PM2/port runbook, operator guide.

Parent final gate:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git diff --check
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
(cd services/core-api && go test ./internal/handler ./internal/service ./internal/repository/postgres && go build -o bin/api ./cmd/api)
npm run security:scan-secrets:staged
```

---

## Data Model

### Migration: add explicit session access policy

**Create:** `services/core-api/db/migrations/XXX_cbt_student_direct_login_policy.sql`

```sql
-- +goose Up
ALTER TABLE cbt_exam_sessions
  ADD COLUMN IF NOT EXISTS access_mode TEXT NOT NULL DEFAULT 'secure_exam',
  ADD COLUMN IF NOT EXISTS student_portal_direct_login_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS require_room_token_for_web BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS nisn_direct_login_enabled BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE cbt_exam_sessions
  ADD CONSTRAINT chk_cbt_exam_sessions_access_mode
  CHECK (access_mode IN ('secure_exam', 'web_fallback', 'simulation')) NOT VALID;

ALTER TABLE cbt_exam_sessions
  VALIDATE CONSTRAINT chk_cbt_exam_sessions_access_mode;

CREATE INDEX IF NOT EXISTS idx_cbt_exam_sessions_access_policy
  ON cbt_exam_sessions (status, access_mode, student_portal_direct_login_enabled, require_room_token_for_web, nisn_direct_login_enabled);

-- +goose Down
DROP INDEX IF EXISTS idx_cbt_exam_sessions_access_policy;

ALTER TABLE cbt_exam_sessions
  DROP CONSTRAINT IF EXISTS chk_cbt_exam_sessions_access_mode;

ALTER TABLE cbt_exam_sessions
  DROP COLUMN IF EXISTS nisn_direct_login_enabled,
  DROP COLUMN IF EXISTS require_room_token_for_web,
  DROP COLUMN IF EXISTS student_portal_direct_login_enabled,
  DROP COLUMN IF EXISTS access_mode;
```

**Acceptance:** migration additive, no existing session behavior changes.

---

## Backend Design

### New/updated SQL queries

**Modify:** `services/core-api/db/queries/cbt_sessions.sql`

Add/select fields in relevant student portal CBT queries:

- `access_mode`
- `student_portal_direct_login_enabled`
- `require_room_token_for_web`
- `allow_web_fallback`
- assigned room/seat
- session status/window

Add query like:

```sql
-- name: GetStudentPortalParticipantForStart :one
SELECT
  ep.id AS participant_id,
  ep.student_id,
  ep.session_id,
  ep.room_id,
  ep.submitted_at,
  ep.locked_at,
  ep.device_fingerprint,
  s.status AS session_status,
  s.scheduled_start,
  s.scheduled_end,
  s.access_mode,
  s.student_portal_direct_login_enabled,
  s.require_room_token_for_web,
  r.allow_web_fallback,
  r.status AS room_status,
  p.id AS package_id,
  p.duration_minutes
FROM cbt_exam_participants ep
JOIN cbt_exam_sessions s ON s.id = ep.session_id
JOIN cbt_packages p ON p.id = s.package_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
WHERE ep.id = $1
  AND ep.student_id = $2;
```

Add query/mutation for direct launch if needed:

```sql
-- name: UpdateParticipantDirectPortalLogin :one
UPDATE cbt_exam_participants
SET device_fingerprint = COALESCE(NULLIF(device_fingerprint, ''), $2),
    login_ip = $3,
    client_type = $4,
    browser_fingerprint_hash = $5,
    client_user_agent_hash = $6,
    started_at = COALESCE(started_at, NOW()),
    last_seen_at = NOW()
WHERE id = $1
  AND (device_fingerprint = '' OR device_fingerprint = $2)
RETURNING id;
```

If actual columns differ, use existing equivalent from `UpdateParticipantLogin`.

### Service method

**Modify/Create:** `services/core-api/internal/service/student_portal.go`

Add:

```go
type StartStudentPortalExamInput struct {
    ParticipantID       pgtype.UUID
    UserID              pgtype.UUID
    DeviceFingerprint   string
    ClientType          string
    BrowserFingerprint  string
    UserAgent           string
    LoginIP             string
}

type StartStudentPortalExamResult struct {
    ParticipantID        string
    SessionID            string
    ExamRuntimeToken     string // transitional; do not show in list/cetak
    TimeRemainingSeconds int64
    RedirectPath         string
}
```

Rules:

1. Resolve authenticated `user_id -> student_id` server-side.
2. Query participant by `{participantID, studentID}` only.
3. Reject if not found with generic 404/403.
4. Require:
   - session status `active`
   - within `scheduled_start` and `scheduled_end`
   - `student_portal_direct_login_enabled=true`
   - access mode is `simulation` or `web_fallback`
   - participant has room assignment
   - not submitted
   - not locked
5. If `require_room_token_for_web=false`, bypass room token only for `client_type=web_fallback` or `simulation` policy.
6. If `require_room_token_for_web=true`, return clear blocked reason: `TOKEN_ROOM_REQUIRED` and do not expose token.
7. Enforce device binding using existing `UpdateParticipantLogin` semantics.
8. Insert audit events:
   - `student_exam_start_attempt`
   - `student_exam_start_success`
   - `student_exam_start_blocked`
   - `room_token_bypassed` when relevant
   - `web_fallback_used` when relevant
9. Do not log raw tokens, passwords, user agent, or raw fingerprint.

### Handler endpoint

**Modify:** `services/core-api/internal/handler/student_portal.go`

Add route handlers:

`POST /api/cbt-portal/login`

`POST /api/cbt-portal/participants/{participantID}/start`

Login request body:

```json
{
  "nisn": "1234567890",
  "password": "1234567890"
}
```

Login validation rules:

- `nisn` and `password` must match exactly for simulation login.
- The NISN must belong to one active student.
- There must be at least one active/current participant assignment in a session with `nisn_direct_login_enabled=true`.
- Error message must be generic: `NISN atau kode masuk tidak sesuai` to avoid enumeration.
- Do not create a normal web-admin auth session; issue only a CBT portal scoped session/cookie.

Start request body:

```json
{
  "device_fingerprint": "browser-or-device-id",
  "client_type": "web_fallback",
  "browser_fingerprint": "optional"
}
```

Response success:

```json
{
  "participant_id": "...",
  "session_id": "...",
  "runtime_token": "opaque/internal-compatible-token",
  "redirect_path": "/portal/siswa/cbt/{participantID}/ujian",
  "time_remaining_seconds": 12345
}
```

Response blocked:

```json
{
  "error": "belum_dibuka",
  "message": "Ujian belum dapat dimulai. Hubungi pengawas."
}
```

### Router registration

**Modify:** `services/core-api/cmd/api/main.go` or existing route registration file.

Register the start endpoint under authenticated student portal middleware, not public exam token middleware.

---

## Frontend Design

### Portal CBT khusus

Rekomendasi implementasi:

- **Preferred:** buat app kecil `apps/cbt-portal` agar benar-benar terpisah dari web-admin dan bisa dijalankan di port lain, misalnya `8031`.
- **Fallback cepat:** route khusus di `apps/web-admin/src/routes/cbt` tetapi dijalankan/di-expose dengan link berbeda; tetap harus menyembunyikan admin shell.

Untuk plan swarm, gunakan preferred path: `apps/cbt-portal`.

**Create:** `apps/cbt-portal/src/routes/+page.svelte`
**Create:** `apps/cbt-portal/src/routes/ujian/[participant_id]/+page.svelte`
**Create:** `apps/cbt-portal/src/routes/api/login/+server.ts`
**Create:** `apps/cbt-portal/src/routes/api/participants/[participant_id]/start/+server.ts`

### BFF/proxy route

**Create:** `apps/cbt-portal/src/routes/api/participants/[participant_id]/start/+server.ts`

Behavior:

- Require current session/cookie auth using existing BFF helpers.
- Proxy POST to core-api `/api/portal/student/cbt/{participantID}/start`.
- Never log runtime token.
- Return backend status/message exactly enough for UI.

### CBT Portal login and exam list

**Create:** `apps/cbt-portal/src/routes/+page.svelte`
**Create:** `apps/cbt-portal/src/routes/ujian/[participant_id]/+page.svelte`

UI requirement:

- Login card with fields **NISN** and **Kode Masuk**; helper text: `Untuk simulasi, kode masuk sama dengan NISN`.
- Mobile-first card: **Ujian Saya** after NISN login.
- Show only metadata:
  - title
  - mapel
  - kelas/ruang/kursi
  - start/end WITA
  - status label
  - blocked reason
- Primary CTA: **Mulai Ujian**.
- No field for token peserta.
- No field for token ruang in Opsi A simulation.
- No web-admin sidebar/navbar/admin modules.
- If blocked by policy, show human message:
  - belum waktunya
  - sesi belum aktif
  - belum dibuka pengawas
  - perangkat berbeda
  - sudah dikumpulkan
  - hubungi proktor

### Exam runtime page

**Create or modify:** `apps/web-admin/src/routes/portal/siswa/cbt/[participant_id]/ujian/+page.svelte`

Transitional approach:

- Store runtime token only in memory/session storage as required by current `/api/exam/*` BFF runtime.
- Do not display the token.
- Include heartbeat and submit flow using existing exam proxy if available.
- If runtime token missing after refresh, call start endpoint again; backend should allow same device resume.

### Admin session policy UI

**Modify:** `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte` or session settings component.

Add a safe panel:

- Label: **Mode Login Siswa**
- Options:
  - `Aman/resmi: token tetap wajib` (default)
  - `Simulasi: siswa login akun, tanpa token manual`
- Checkbox/detail:
  - `Aktifkan Ujian Saya untuk siswa`
  - `Tidak wajibkan token ruang untuk web simulasi`
- Warning text:
  - gunakan hanya untuk simulasi/gladi atau atas keputusan panitia.

Backend update endpoint may reuse existing session update route or add a small policy endpoint.

---

## Test Plan

### Backend unit tests

**Modify/Create:**

- `services/core-api/internal/service/student_portal_test.go`
- `services/core-api/internal/handler/student_portal_test.go`
- `services/core-api/internal/service/exam_test.go` if shared logic touched.

Cases:

1. Student can start own assigned active simulation without token.
2. Student cannot start another student's participant.
3. Student cannot start when direct login disabled.
4. Student cannot start secure exam without room token policy.
5. Student cannot start before scheduled_start.
6. Student cannot start after scheduled_end.
7. Student cannot start submitted participant.
8. Student cannot start locked participant.
9. Device first bind succeeds.
10. Device mismatch returns conflict.
11. Web fallback room not allowed blocks unless simulation policy explicitly allows.
12. Audit event created for success and blocked attempts.
13. Response does not include raw room token or answer key.

Run:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
go test ./internal/service ./internal/handler -run 'StudentPortal.*Cbt|StartStudentPortalExam|Exam' -count=1
```

### Frontend tests

**Create/Modify:**

- `apps/web-admin/src/routes/portal/siswa/cbt/student-direct-start.test.ts`
- or colocated Vitest test matching existing test style.

Cases:

1. Page has **Ujian Saya** and **Mulai Ujian**.
2. Page does not show token input in simulation direct-login mode.
3. Blocked reason messages are operator-friendly.
4. Start API errors are shown without raw backend stack/secret.
5. Runtime token is not rendered in DOM text.

Run:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run test:unit -- --run src/routes/portal/siswa/cbt/student-direct-start.test.ts
npm run check
npm run build
```

### Smoke test

**Create:** `apps/web-admin/scripts/cbt-student-direct-start-smoke.mjs`

Skip-by-default unless env complete:

- `CBT_PORTAL_DIRECT_START_BASE_URL`
- `CBT_PORTAL_DIRECT_START_NISN`
- `CBT_PORTAL_DIRECT_START_CODE`
- `CBT_PORTAL_DIRECT_START_PARTICIPANT_ID`

Smoke steps:

1. Login as siswa.
2. Open `/portal/siswa` or `/portal/siswa/cbt/{participantID}`.
3. Verify **Mulai Ujian** visible.
4. Click start.
5. Verify exam shell appears.
6. Do not answer/submit by default.

Add root/package script if approved:

```json
"smoke:web:cbt:direct-start": "npm --prefix apps/web-admin run smoke:cbt:direct-start"
```

---

## Rollout Plan

### Sprint 0 — safe rollout gate

1. Backup DB before migration.
2. Run migration.
3. Verify defaults preserve secure mode:

```sql
SELECT
  COUNT(*) FILTER (WHERE access_mode <> 'secure_exam') AS non_secure_mode_count,
  COUNT(*) FILTER (WHERE student_portal_direct_login_enabled IS TRUE) AS direct_login_enabled_count,
  COUNT(*) FILTER (WHERE require_room_token_for_web IS FALSE) AS room_token_bypass_count,
  COUNT(*) FILTER (WHERE nisn_direct_login_enabled IS TRUE) AS nisn_direct_login_enabled_count
FROM cbt_exam_sessions;
```

Expected after migration: all `0`.

4. No production policy enabled yet.
5. Deploy backend/frontend after tests/build.

### Enable only this simulation session

For current/next simulation, set only chosen session:

```sql
UPDATE cbt_exam_sessions
SET access_mode='simulation',
    student_portal_direct_login_enabled=true,
    nisn_direct_login_enabled=true,
    require_room_token_for_web=false,
    updated_at=now()
WHERE id = '<SESSION_ID>';
```

Then verify:

```sql
SELECT id, title, status, access_mode, student_portal_direct_login_enabled, nisn_direct_login_enabled, require_room_token_for_web
FROM cbt_exam_sessions
WHERE id = '<SESSION_ID>';
```

---

## Detailed Tasks

### Task 1: Migration and SQL policy fields

**Objective:** Add additive session policy fields with secure defaults.

**Files:**

- Create: `services/core-api/db/migrations/XXX_cbt_student_direct_login_policy.sql`
- Modify: `services/core-api/db/queries/cbt_sessions.sql`
- Generated: `services/core-api/internal/repository/postgres/*.sql.go`

**Steps:**

1. Write migration.
2. Add policy fields to student portal CBT list/detail queries.
3. Add `GetStudentPortalParticipantForStart` query.
4. Run sqlc generate.
5. Run repository/service compile tests.
6. Commit: `feat(cbt): add student direct login policy fields`.

### Task 2: Backend service direct start logic

**Objective:** Implement server-side account-authenticated start gate.

**Files:**

- Modify: `services/core-api/internal/service/student_portal.go`
- Test: `services/core-api/internal/service/student_portal_test.go`

**Steps:**

1. Write failing tests for own participant, cross-student denial, disabled policy, window closed, device mismatch.
2. Implement `StartCbtExamForStudent` or equivalent.
3. Add audit event insert helper.
4. Reuse existing hashing/safeHash helpers or centralize if needed.
5. Run service tests.
6. Commit: `feat(cbt): start assigned exam from student portal`.

### Task 3: Backend handler and route

**Objective:** Expose authenticated student portal start endpoint.

**Files:**

- Modify: `services/core-api/internal/handler/student_portal.go`
- Modify: `services/core-api/cmd/api/main.go` or route registration file
- Test: `services/core-api/internal/handler/student_portal_test.go`

**Steps:**

1. Write handler tests for success and blocked messages.
2. Parse `participantID` route param.
3. Read authenticated user from request context.
4. Return safe JSON only.
5. Map domain conflicts to 409/403/400 with Indonesian messages.
6. Run handler tests.
7. Commit: `feat(cbt): add student exam start endpoint`.

### Task 4: CBT portal app and BFF proxy

**Objective:** Create a dedicated CBT portal surface on a separate app/link/port and allow it to call backend login/start endpoints securely.

**Files:**

- Create: `apps/cbt-portal/package.json`
- Create: `apps/cbt-portal/src/routes/+page.svelte`
- Create: `apps/cbt-portal/src/routes/api/login/+server.ts`
- Create: `apps/cbt-portal/src/routes/api/participants/[participant_id]/start/+server.ts`
- Test: `apps/cbt-portal/src/routes/api/cbt-portal-proxy.test.ts`

**Steps:**

1. Write failing proxy test.
2. Implement proxy using existing backend helpers.
3. Ensure no raw token is logged.
4. Run unit test.
5. Commit: `feat(web): proxy student CBT direct start`.

### Task 5: NISN login + Ujian Saya UI

**Objective:** Show NISN login, assigned exams, and `Mulai Ujian` without token inputs.

**Files:**

- Modify/Create: `apps/cbt-portal/src/routes/+page.svelte`
- Modify/Create: `apps/cbt-portal/src/routes/ujian/[participant_id]/+page.svelte`
- Test: `apps/cbt-portal/src/routes/cbt-portal-nisn-login.test.ts`

**Steps:**

1. Add test asserting NISN + Kode Masuk fields exist and no token input exists in direct-login mode.
2. Add mobile-first exam card.
3. Add `Mulai Ujian` button calling BFF start endpoint.
4. Store runtime token only for runtime handoff, not display.
5. Add clear blocked reason UI.
6. Run unit/check/build.
7. Commit: `feat(web): add Ujian Saya direct start flow`.

### Task 6: Exam runtime handoff/resume in CBT portal

**Objective:** Let student continue into exam shell after NISN direct start.

**Files:**

- Create/Modify: `apps/cbt-portal/src/routes/ujian/[participant_id]/+page.svelte`
- Modify existing exam proxy usage if needed.
- Test: `apps/web-admin/src/routes/portal/siswa/cbt/student-exam-runtime.test.ts`

**Steps:**

1. Write test that runtime token is not rendered.
2. Implement runtime shell handoff.
3. If page refresh happens, call start endpoint again and rely on same-device resume.
4. Display answer/submit UI through existing exam API proxy.
5. Run tests/build.
6. Commit: `feat(web): launch student exam runtime from portal`.

### Task 7: Admin policy UI

**Objective:** Let admin/panitia enable Opsi A per session safely.

**Files:**

- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- Possibly create BFF proxy route if backend endpoint needed.
- Backend handler/service if no generic session update exists.
- Tests: session policy UI + backend update tests.

**Steps:**

1. Add backend policy update endpoint only if needed.
2. Add UI panel **Mode Login Siswa**.
3. Default display: secure/token required.
4. Add warning for simulation/direct-login.
5. Run tests.
6. Commit: `feat(asesmen): configure student direct login per session`.

### Task 8: Audit/reporting and proctor visibility

**Objective:** Make direct-login events visible enough for operations.

**Files:**

- Modify proctoring/session detail data mapping as needed.
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte` if relevant.
- Tests around event labels.

**Steps:**

1. Add labels for `student_exam_start_success`, `room_token_bypassed`, `web_fallback_used`.
2. Ensure no raw token in event payload UI.
3. Add count badges if practical.
4. Run tests/build.
5. Commit: `feat(asesmen): surface student direct-login audit events`.

### Task 9: Smoke script and docs

**Objective:** Provide repeatable validation and operator guide.

**Files:**

- Create: `apps/web-admin/scripts/cbt-student-direct-start-smoke.mjs`
- Create: `apps/web-admin/docs/cbt-opsi-a-login-siswa-tanpa-token.md`
- Modify: `apps/web-admin/package.json`
- Modify: root `package.json` if adding root script.

**Steps:**

1. Write skip-by-default smoke script.
2. Add docs for panitia/proktor/siswa.
3. Add package scripts.
4. Run smoke with no env: expected SKIP exit 0.
5. Run unit/check/build.
6. Commit: `test(cbt): add direct student start smoke runner`.

### Task 10: Final integration review and deploy gate

**Objective:** Verify all implementation is safe before deployment.

**Steps:**

1. Dispatch spec compliance reviewer for entire plan.
2. Dispatch security/quality reviewer.
3. Run:

```bash
git diff --check
(cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml && go test ./internal/handler ./internal/service ./internal/repository/postgres && go build -o bin/api ./cmd/api)
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
npm run security:scan-secrets:staged
```

4. Only after explicit approval: apply migration, restart PM2, smoke health routes.
5. Commit any final docs/fixes.

---

## Acceptance Criteria

### Student

- Siswa membuka portal CBT khusus, bukan web-admin utama.
- Siswa login dengan NISN dan kode masuk yang sama dengan NISN untuk sesi simulasi.
- Siswa melihat **Ujian Saya**.
- Siswa klik **Mulai Ujian**.
- Tidak ada input token peserta atau token ruang untuk session simulation/direct-login.
- Jika gagal, pesan jelas dan tidak teknis.

### Security

- Siswa tidak bisa melihat/start peserta siswa lain.
- Direct login dan NISN-only login disabled by default.
- Secure official sessions tetap butuh token/room token kecuali policy eksplisit diubah.
- Tidak ada raw token di UI list, DOM, logs, atau audit payload.
- Device mismatch tetap diblokir.
- Semua start attempt diaudit.

### Operations

- Admin bisa mengaktifkan Opsi A per sesi.
- Panitia dapat membedakan sesi secure vs simulasi.
- Proktor dapat melihat event direct login/fallback.
- Existing Android/token flow tidak rusak.

### Verification

- Backend tests pass.
- Frontend tests/check/build pass.
- Smoke script skip-by-default works.
- Secret scan staged pass.
- Protected web routes return expected login redirect when unauthenticated.

---

## Rollback Strategy

1. Disable policy on affected session:

```sql
UPDATE cbt_exam_sessions
SET access_mode='secure_exam',
    student_portal_direct_login_enabled=false,
    nisn_direct_login_enabled=false,
    require_room_token_for_web=true,
    updated_at=now()
WHERE id='<SESSION_ID>';
```

2. If code rollback needed, revert deployment to previous commit.
3. If migration rollback required, use Down migration only after confirming code no longer reads those columns.
4. Existing token-based flow remains usable because no token data is removed.

---

## Notes from Swarm Review

- Current exam login requires participant token, room token, active session, assigned room, device fingerprint, and time window.
- Student portal already lists CBT schedule and has token reveal flow; Opsi A should extend this, not rewrite from scratch.
- `allow_web_fallback` exists per room but should not mean “room token optional” by itself.
- Room token hash fields exist, but active validation still depends on plaintext `room_token`; do not clear plaintext yet.
- For official exams, add room/proctor gate in later phase before removing token requirement broadly.
- NISN=NISN is operationally simple but weak as authentication; restrict to simulation/direct-link CBT portal, rate-limit heavily, audit all attempts, and do not grant normal web-admin/student account session.
