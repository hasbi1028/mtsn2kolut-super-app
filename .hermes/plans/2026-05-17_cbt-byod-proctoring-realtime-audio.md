# CBT BYOD Proctoring Realtime + Audio Warning Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Membangun portal pengawas CBT untuk skema BYOD yang realtime, audit-safe, punya klasifikasi pelanggaran ringan/sedang/berat/teknis, audio warning, tindakan pengawas, dan berita acara/export.

**Architecture:** Go core-api tetap menjadi pemilik domain, PostgreSQL, migration, sqlc, policy risiko, audit, dan endpoint. SvelteKit web-admin hanya menjadi UI + BFF/proxy ke core-api; tidak boleh akses database langsung. Mobile/student-facing app hanya mengirim telemetry ter-whitelist; keputusan risiko final dihitung backend dan diverifikasi pengawas.

**Tech Stack:** PostgreSQL migrations + sqlc, Go services/handlers, SvelteKit admin frontend, SSE/polling fallback, shadcn-svelte UI, browser Web Audio API untuk audio warning.



---

# Swarm Review v2 — Required Patches Before Implementation

**Review date:** 2026-05-17  
**Review mode:** 5-agent swarm review.  
**Overall verdict:** `REQUEST_CHANGES` before implementation. The plan direction is approved, but implementation must incorporate the patches below.

## Swarm Review Agents

1. **Backend/Security reviewer** — data model, RBAC, audit, transaction boundaries, sqlc/Go architecture.
2. **Frontend/UX reviewer** — portal pengawas, audio UX, accessibility, alarm fatigue, Indonesian labels.
3. **CBT/BYOD Anti-cheat reviewer** — severity/risk policy, false positives, technical-vs-cheating separation.
4. **Ops/Production reviewer** — migration, backup, rollback, PM2 deploy order, smoke safety.
5. **QA/Test reviewer** — fixture strategy, automated/manual tests, SSE/polling, performance acceptance.

## Global Plan Changes Required

These rules supersede any weaker wording later in this document.

### 1. Reuse Existing Participant Event Model Unless Audit Proves Otherwise

Repo already has an existing participant event flow/table from earlier CBT migrations and proctoring pages. Sprint 1 must **audit and reuse/extend the existing participant event table** instead of blindly creating a duplicate event table.

Required implementation decision in Sprint 0/1:

- If existing table can support proctoring: extend it additively with severity/ack/dedup fields.
- If a new table is unavoidable: document why, define dual-read/compatibility behavior, avoid duplicate live writes, and add migration/backfill strategy.

Do not implement both ambiguous live event sources without a single canonical timeline contract.

### 2. RBAC Scope Matrix Is Mandatory

Before adding acknowledge/action/report endpoints, define and implement explicit scope checks:

- **Admin/Panitia:** may monitor all sessions/rooms and perform high-impact actions.
- **Assigned room proctor:** may monitor/action only participants in assigned room.
- **Guru mapel:** read/limited action only if existing teacher/session/class-subject authorization allows.
- **Unauthorized actor:** must receive `401/403`; do not leak whether unrelated event IDs exist.

For `POST .../events/{event_id}/acknowledge`, backend must lookup event → participant/session/room, then scope-check before mutation.

If new RBAC permissions are introduced, the same migration must grant them to `admin` via `rbac_role_permissions`, following repo policy.

### 3. High-Impact Actions Must Be Atomic and Audited

For these actions, state mutation and audit insert must happen in one DB transaction:

- `unlock_access`
- `hold_access`
- `reset_device_binding`
- `force_submit`
- `escalate_to_committee`
- `mark_technical_issue` when it affects report state

Required audit fields:

- `actor_user_id` **NOT NULL**
- `actor_username_snapshot`
- `actor_employee_id` when available
- `session_id`, `room_id`, `participant_id` as applicable
- `action_type`, `reason`, `notes`
- `request_id` if available
- `source_ip` if safely available
- `created_at`

Existing reset/unlock/force-submit routes must be refactored to call the new audited action service. No legacy bypass path.

### 4. Race-Safe Risk Updates and Dedup

Telemetry recording must use a service method like `RecordProctorTelemetry(ctx, participantID, input)` that:

1. starts a transaction;
2. locks participant row (`SELECT ... FOR UPDATE`) or equivalent;
3. validates/normalizes event type;
4. sanitizes payload;
5. applies race-safe dedup/cooldown;
6. inserts canonical event;
7. updates participant `risk_score`, `violation_count`, `risk_level`, `locked_at`, `locked_reason`;
8. commits.

One lifecycle episode must not be counted as multiple violations. Backlogged telemetry after reconnect must use original event timestamp/correlation window and dedup logic.

### 5. Technical Events Must Never Become Cheating by Themselves

Backend invariant:

- `severity = technical` ⇒ `risk_delta = 0`
- `severity = technical` ⇒ `violation_count` does not increase
- `severity = technical` ⇒ `locked_at` is not set

If another suspicious event happens during a technical issue, record it as a separate event with its own evidence. Do not “upgrade” a technical event into cheating.

### 6. UI Must Use Backend-Provided Decision Fields

Svelte may group/display data, but must not reimplement risk policy. Live-summary payload must include UI-ready fields:

Per participant minimum:

- `participant_id`, `nama`, `nis`, `room_id`, `room_name`, `seat_no`
- `connection_status`: `online | terlambat | terputus | selesai`
- `sync_status`: `sinkron | belum_sinkron | tertahan | tidak_diketahui`
- `risk_level`, `risk_score`, `locked_at`
- `needs_action: boolean`
- `action_priority: number`
- `needs_action_reason`
- `latest_event`
- `unacknowledged_count`

Per alarm event minimum:

- `event_id`, `severity`, `category`
- `participant_id`, `participant_name`, `room_name`
- `label_id`, `message_id`
- `created_at`, `acknowledged_at`
- `requires_note`
- `audio_key`
- `is_mass_technical_issue`

### 7. Audio Alarm Must Be Rate-Limited and Visual-First

Audio is opt-in and never the only indicator. Required limits:

- warning: max 1 sound / participant / event type / 2 minutes
- medium: max 1 sound / participant / 60 seconds
- critical: max 1 sound / participant / 30 seconds
- technical mass: max 1 sound / room / 2 minutes
- global room cap: max 5 sounds / minute; excess becomes visual-only

Required controls:

- “Aktifkan Suara”
- “Tes suara”
- “Senyapkan 5 menit”
- “Senyapkan 10 menit”
- “Senyapkan sampai sesi selesai”
- browser/audio failure state: “Suara tidak bisa diputar — gunakan peringatan visual”

### 8. Accessibility and Tablet UX Are Acceptance Criteria

Room proctor page must be tablet/HP-friendly:

- card-first layout on small screens; full table only in “Rincian lengkap”;
- sticky header with room, live status, and token control;
- “Butuh Tindakan” pinned near top;
- participant detail as drawer/bottom sheet on small screens;
- high-impact actions only inside detail/dialog, not as direct table buttons.

Accessibility requirements:

- status not conveyed by color alone;
- every badge has text;
- alarm updates use non-spammy `aria-live`;
- dialog/drawer traps focus and restores focus;
- filter tabs/chips expose selected state;
- tap targets are comfortable for tablet use.

### 9. Room Token Visibility Policy

Room token must be operationally visible to assigned proctors, but controlled:

- default masked, e.g. `••••1234`;
- button “Tampilkan token”;
- optional auto-hide timeout;
- “Salin token” action;
- do not expose room token outside authorized session/room scope.

### 10. Final Indonesian UI Glossary

Use these labels consistently:

- `heartbeat` → **Status koneksi**
- `stale` → **Kontak terlambat**
- `offline` → **Terputus**
- `pending sync` → **Jawaban belum terkirim**
- `locked` → **Akses ditahan**
- `device mismatch` → **Perangkat tidak sesuai**
- `acknowledge` → **Sudah dicek**
- `alarm tray` → **Daftar peringatan**
- `live mode` → **Pemantauan langsung**
- `critical` → **Kritis**
- `technical` → **Gangguan teknis**
- `force submit` → **Kirim paksa jawaban**
- `reset device binding` → **Reset perangkat peserta**
- `command center` → **Pusat pengawasan panitia**

---

## Anti-cheat Policy Patches

### Screenshot Policy

`FLAG_SECURE` is prevention, not proof. Standard Android/BYOD screenshot “detection” is not fully reliable.

Rules:

- One screenshot event must not auto-lock.
- Ambiguous screenshot event is warning/low-medium only.
- Screenshot becomes medium only with corroborating evidence: repeated app switch, valid platform callback, tamper signal, or proctor verification.
- Reports should phrase this as “indikasi percobaan tangkapan layar”, not definitive cheating unless verified.

### Device Binding & Token Sharing Policy

Rules:

- Device fingerprint must be a non-sensitive derived identifier; do not store raw hardware secrets.
- Weak mismatch → warning/medium and requires proctor check.
- Strong mismatch/collision/token reuse → critical pending verification; lock eligible only under defined conditions.
- Reset device binding requires assigned proctor/admin actor, reason, and notes.
- Device change due to battery/network/hardware can be marked technical, not cheating.

### Event Type Risk Matrix

Sprint 1 must define a deterministic event type → severity → risk_delta → lock eligibility table. Initial recommended mapping:

- `focus_lost_short`: warning, `+5`, no lock
- `app_switch_once`: warning, `+10`, no lock
- `app_switch_repeated`: medium, `+25`, lock only after repeated threshold
- `background_over_threshold`: medium, `+30`, lock only after repeated threshold
- `split_screen_detected`: medium, `+30`, lock only after repeated threshold
- `pip_detected`: medium, `+25`, lock only after repeated threshold
- `overlay_suspicious_confirmed`: medium, `+20/+30`, no single-event lock
- `screenshot_attempt_ambiguous`: warning, `+5/+10`, no lock
- `screenshot_attempt_valid`: medium, `+20/+25`, no single-event lock
- `device_mismatch_weak`: medium, `+25`, no auto-lock
- `device_mismatch_strong`: critical, `+60`, may hold access pending verification
- `token_reuse_confirmed`: critical, `+80`, lock eligible
- `root_emulator_weak`: medium, `+25`, no single-event lock
- `root_emulator_strong`: critical, `+60/+80`, lock eligible with corroboration
- `offline_short`: technical or warning, `+0/+5 max`, no lock
- `offline_mass`: technical, `+0`, no lock
- `pending_sync`: technical/warning, `+0`, no lock
- `submit_held_pending_sync`: technical, `+0`, no lock

### Auto-lock Thresholds

Allowed auto-lock:

- `risk_score >= 80` after dedup and not from technical events;
- `token_reuse_confirmed`;
- `device_mismatch_strong` plus corroborating signal;
- 3 valid medium events in 10 minutes after prior warning.

Forbidden auto-lock:

- pending sync;
- offline / heartbeat late;
- focus lost short;
- one app switch;
- one screenshot ambiguous;
- mass technical issue.

### Autosave / Pending Sync Contract

Sprint 1/2 must define these fields in telemetry/dashboard if not already available:

- `last_local_save_at`
- `last_synced_at`
- `pending_answer_count`
- `sync_state`: `synced | pending | failed | unknown`

Submit guard:

- normal submit final only when server knows answers are synced;
- pending sync appears to proctor as “Jawaban belum terkirim”;
- force submit during pending sync needs audited proctor/admin action and must state risk of incomplete answers.

---

## QA/Test Patches

### Test Fixture Strategy

Add fixture builders for:

- sessions, rooms, participants;
- participant states: normal, warning, high, locked, offline, pending sync, submitted;
- event/action timeline;
- mass technical issue;
- actors: admin, assigned room proctor, teacher, unauthorized user;
- fake clock/time injection for dedup/cooldown;
- frontend fixture payloads for dashboard, alarm tray, report/export.

### Required Test Coverage

Risk policy tests must include:

- boundary scores `0, 19, 20, 49, 50, 79, 80`;
- cooldown inside/outside window;
- technical event never increases `violation_count`;
- locked participant remains locked until audited unlock;
- client-supplied severity/risk ignored;
- unknown event rejected/ignored safely.

Telemetry handler tests must include:

- valid telemetry creates normalized event;
- forged severity/risk ignored;
- oversized JSON rejected via `http.MaxBytesReader`;
- malformed JSON rejected;
- dedup prevents double risk;
- unauthorized request rejected.

Acknowledge/action tests must include:

- critical event without note rejected;
- acknowledge is idempotent;
- actor stored;
- unauthorized actor forbidden;
- scope mismatch forbidden/not leaked;
- unlock/hold/reset/force_submit mutate state and create audit row atomically;
- force submit blocks or explicitly audits pending sync.

SSE/BFF tests must include:

- `Content-Type: text/event-stream`;
- scoped payload by session/room;
- auth/permission enforced;
- client disconnect cleanup;
- frontend fallback to polling when SSE fails;
- BFF forwards method/body/cookies and does not access DB.

Frontend tests must include:

- audio opt-in;
- mute/snooze;
- throttle;
- browser AudioContext unavailable;
- visual fallback;
- CSV escaping and JSON evidence redaction for reports.

### Full Validation Must Include Unit Tests

Every sprint validation must include:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-sprintN ./cmd/api
```

If `test:unit` has pre-existing unrelated failures, document exact failure and run targeted tests for changed files.

### Performance / Realtime Acceptance

Sprint 2 must add performance validation:

- representative dataset target: at least 30 rooms × 30 participants if feasible in test/staging;
- no unbounded per-participant DB query loop;
- inspect relevant query plan with `EXPLAIN`;
- define p95 response target for live-summary before production rollout;
- SSE/polling interval must not overload DB during mass exam.

---

## Production/Ops Patches

### Safer Deploy Order

Production order must be:

1. explicit approval;
2. pre-deploy validation and build core-api artifact;
3. backup + checksum + restore-list verification;
4. transactional/idempotent migration;
5. restart core-api;
6. health + PM2/log check;
7. build web-admin;
8. restart web-admin immediately after successful build;
9. web-admin PM2/log check;
10. smoke tests;
11. observe logs 5–15 minutes.

Do not apply migration before confirming the new core-api builds successfully.

### Backup Command Must Use One Dump Variable

Use this pattern:

```bash
BACKUP_DIR=/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql
TS=$(date +%Y%m%d-%H%M%S)
DUMP="$BACKUP_DIR/pre-cbt-proctoring-$TS.dump"
mkdir -p "$BACKUP_DIR"
set -a; source .env >/dev/null 2>&1; set +a
pg_dump "$DATABASE_URL" -Fc -f "$DUMP"
sha256sum "$DUMP" > "$DUMP.sha256"
sha256sum -c "$DUMP.sha256"
pg_restore --list "$DUMP" >/tmp/pre-cbt-proctoring-$TS.restore-list.txt
```

Never use `set -x` while loading `.env`. Never print `DATABASE_URL`.

### Transactional Migration

If migration contains only transactional statements, production command should use:

```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -1 -f db/migrations/114_cbt_proctoring_policy_actions.sql
```

Migration must be idempotent if using manual `psql -f` deployment pattern.

### Rollback SOP Required

Add a rollback task before deploy:

- if backup fails → abort;
- if migration dry-run fails → abort;
- if migration fails → stop, inspect DB state, do not restart apps;
- if migration succeeds but core-api fails → rollback core-api binary/commit; keep DB if additive/backward-compatible;
- if web-admin fails → rollback/rebuild web-admin; core-api can stay if healthy;
- destructive DB restore requires separate explicit approval.

Restore command pattern must be documented but not run without approval:

```bash
pg_restore --clean --if-exists --no-owner --dbname "$DATABASE_URL" "$DUMP"
```

### Smoke Test Safety

Production smoke writes are allowed only on staging or a dedicated `SMOKE/TEST` session with no real participant and excluded from reports.

For official production sessions, smoke tests must be read-only:

- open pages;
- verify data loads;
- no fixture event;
- no acknowledge;
- no proctor action;
- no force submit/reset/unlock/hold.

### Post-deploy Monitoring

Required checks:

```bash
pm2 status
pm2 logs mtsn2kolut-core-api --lines 100 --nostream
pm2 logs mtsn2kolut-web-admin --lines 100 --nostream
curl -fsS http://127.0.0.1:8080/health
```

Also verify:

- proctoring summary endpoint returns 200 for safe session;
- report endpoint returns 200 for safe session;
- web-admin page loads without 500;
- no SvelteKit missing chunk/manifest error;
- no repeated DB query/migration errors.

---

## Updated Implementation Gate

Do not start Sprint 1 implementation until Sprint 0 has produced:

1. canonical event table decision;
2. RBAC scope matrix;
3. deterministic event risk matrix;
4. live-summary UI contract;
5. test fixture strategy;
6. production rollback SOP.

After those are documented, the sprint sequence remains valid.


---

## Prinsip Operasional BYOD

1. **Jangan menghukum otomatis dari satu sinyal ambigu.** BYOD rawan false positive: notifikasi sistem, koneksi putus, tombol Home tidak sengaja, OEM overlay, baterai, keyboard, dan jaringan.
2. **Backend yang menghitung risk policy.** Client hanya kirim event; server whitelist event type, dedup/cooldown, severity, risk score, dan status.
3. **Pengawas sebagai verifikator utama.** Pelanggaran sedang/berat harus masuk portal pengawas dan membutuhkan acknowledge/tindakan manusia.
4. **Gangguan teknis dipisah dari kecurangan.** Offline massal, pending sync, dan server unreachable tidak boleh dilabeli curang.
5. **Semua tindakan berdampak tinggi wajib audit.** Unlock, reset device, force submit, hold access, clear incident, dan escalation wajib actor + reason + notes.
6. **No production deploy/restart tanpa approval eksplisit.** Build/test boleh; migration/deploy/restart production hanya setelah persetujuan.

---

## Current Baseline yang Harus Dipertahankan

Existing yang sudah ada dan harus direuse:

- Admin routes:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/report/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- Existing anti-cheat columns from `085_cbt_anti_cheat_enforcement.sql`:
  - `cbt_exam_participants.violation_count`
  - `risk_score`
  - `risk_level`
  - `locked_at`
  - `locked_reason`
- Existing query/service area:
  - `services/core-api/db/queries/cbt_sessions.sql`
  - `services/core-api/internal/service/cbt_session.go`
  - `services/core-api/internal/handler/cbt_session.go`
- Existing runtime participant data:
  - heartbeat
  - app switch count
  - screenshot attempt
  - suspicious flag
  - locked/risk fields
  - proctoring dashboard and event list

Implementation must be additive/backward-compatible and avoid broad rewrite.

---

## Severity Model Final

Use these internal severity levels.

### Level `info`

Purpose: normal timeline only.

Examples:
- participant login
- heartbeat normal
- answer saved
- participant submitted
- proctor acknowledged event

Risk impact: `0`.

Audio: none.

### Level `warning`

Purpose: low-risk warning; monitor if repeated.

Examples:
- one app switch
- focus lost short
- heartbeat late
- answer pending sync
- offline short

Risk impact: `5–15`, cooldown enforced.

Audio: one soft ding, throttled.

### Level `medium`

Purpose: needs proctor attention.

Examples:
- repeated app switch
- background over threshold
- split screen detected
- PiP detected
- overlay suspicious
- screenshot attempt sufficiently valid

Risk impact: `20–35`.

Audio: two-tone warning, requires acknowledge when shown in alarm tray.

### Level `critical`

Purpose: high risk; may hold/lock access depending policy.

Examples:
- device mismatch
- room token/device used elsewhere
- root/emulator/tamper strong signal
- repeated medium violations
- access locked
- suspected manipulated client

Risk impact: `50–80` and may set `locked_at` for strong/repeated signals.

Audio: short alarm, requires acknowledge + notes/action.

### Level `technical`

Purpose: technical issue, not cheating.

Examples:
- offline massal
- many heartbeat late in one room
- server unreachable
- pending sync massal
- submit held because answers not fully synced

Risk impact: `0` unless paired with other suspicious events.

Audio: different technical tone.

---

## Risk Policy Final

### Risk score thresholds

- `0–19`: `normal`
- `20–49`: `warning`
- `50–79`: `high`
- `80+`: `locked`

### Auto-lock rules

Auto-lock is allowed only for:

- device mismatch with strong evidence
- token/device reuse across participants
- repeated medium violations above threshold
- critical tamper/root/emulator signal
- repeated app background/split-screen/PiP after warning

Auto-lock is not allowed for:

- single focus lost
- one app switch
- short offline
- pending sync
- mass technical issue

### Dedup/cooldown

Implement event dedup so one lifecycle transition does not become multiple penalties.

Recommended cooldown keys:

- `{participant_id}:{event_type}` for 60–120 seconds
- `{participant_id}:focus_lost` for 30 seconds
- `{participant_id}:app_backgrounded` for 60 seconds
- `{room_id}:technical_mass_offline` for 120 seconds

---

# Sprint 0 — Safe Audit & Contract

**Goal:** Audit current CBT proctoring implementation, define backend contract, and avoid breaking existing exam flows.

## Task 0.1 — Audit current event schema and routes

**Objective:** Identify current tables, endpoints, event types, and pages before modifying behavior.

**Files:**
- Read: `services/core-api/db/migrations/085_cbt_anti_cheat_enforcement.sql`
- Read: `services/core-api/db/queries/cbt_sessions.sql`
- Read: `services/core-api/internal/service/cbt_session.go`
- Read: `services/core-api/internal/handler/cbt_session.go`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Steps:**
1. Search existing event tables and proctor handlers.
2. List all current event types accepted by mobile/student-facing endpoints.
3. List high-impact proctor actions already available.
4. Confirm existing BFF paths under `apps/web-admin/src/routes/api/asesmen` if present.

**Verification:**
```bash
git status --short
```
Expected: no code changes yet except later docs.

## Task 0.2 — Create proctoring contract doc

**Objective:** Document API payloads, event taxonomy, severity, and UI requirements before implementation.

**Files:**
- Create: `docs/contracts/cbt-byod-proctoring-realtime-audio.md`

**Content must include:**
- event type whitelist
- severity mapping
- risk thresholds
- proctor actions
- report/export fields
- UI labels in Indonesian
- audio behavior and throttling
- non-goals: no direct DB from SvelteKit, no automatic punishment for one ambiguous event

**Verification:**
```bash
test -f docs/contracts/cbt-byod-proctoring-realtime-audio.md
```
Expected: exit `0`.

**Commit:**
```bash
git add docs/contracts/cbt-byod-proctoring-realtime-audio.md .hermes/plans/2026-05-17_cbt-byod-proctoring-realtime-audio.md
git commit -m "docs(cbt): plan byod proctoring portal"
```

---

# Sprint 1 — Event Severity & Risk Policy BYOD

**Goal:** Backend can normalize telemetry into severity/risk without double-counting and without punishing technical issues.

## Task 1.1 — Add additive proctoring policy migration

**Objective:** Add fields/tables needed for event severity, acknowledge state, and audit-safe actions.

**Files:**
- Create: `services/core-api/db/migrations/114_cbt_proctoring_policy_actions.sql`

**Migration design:**

Add or create, using `IF NOT EXISTS` where possible:

- `cbt_proctor_events`
  - `id uuid primary key default gen_random_uuid()`
  - `session_id uuid not null`
  - `room_id uuid null`
  - `participant_id uuid null`
  - `event_type text not null`
  - `severity text not null check in ('info','warning','medium','critical','technical')`
  - `risk_delta integer not null default 0`
  - `dedup_key text not null default ''`
  - `event_data jsonb not null default '{}'::jsonb`
  - `created_at timestamptz not null default now()`
  - `acknowledged_at timestamptz null`
  - `acknowledged_by uuid null`
  - `acknowledge_note text not null default ''`
- `cbt_proctor_actions`
  - `id uuid primary key default gen_random_uuid()`
  - `session_id uuid not null`
  - `room_id uuid null`
  - `participant_id uuid null`
  - `action_type text not null`
  - `reason text not null`
  - `notes text not null default ''`
  - `actor_user_id uuid null`
  - `created_at timestamptz not null default now()`
- Indexes:
  - `(session_id, created_at desc)`
  - `(room_id, created_at desc)`
  - `(participant_id, created_at desc)`
  - `(session_id, severity, acknowledged_at)`
  - unique partial-ish dedup if safe, or plain index on `(participant_id, dedup_key, created_at desc)`

**Important:** If existing event table already exists, extend it instead of creating duplicate. Decide after audit.

**Verification:**
```bash
cd services/core-api
set -a; source ../../.env >/dev/null 2>&1; set +a
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "BEGIN; \i db/migrations/114_cbt_proctoring_policy_actions.sql; ROLLBACK;"
```
Expected: migration applies in transaction and rolls back. Do not print credentials.

## Task 1.2 — Add sqlc queries for proctor events/actions

**Objective:** Typed DB access for event list, create event, acknowledge, and actions.

**Files:**
- Create or modify: `services/core-api/db/queries/cbt_proctoring.sql`
- Generated: `services/core-api/internal/repository/postgres/cbt_proctoring.sql.go`

**Queries:**
- `CreateCbtProctorEvent`
- `ListCbtProctorEventsBySession`
- `ListCbtProctorEventsByRoom`
- `GetRecentCbtProctorEventByDedupKey`
- `AcknowledgeCbtProctorEvent`
- `CreateCbtProctorAction`
- `ListCbtProctorActionsBySession`
- `ListCbtProctorActionsByRoom`

**Verification:**
```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
```
Expected: exit `0`.

## Task 1.3 — Implement backend risk policy service

**Objective:** Centralize severity mapping, risk scoring, dedup, cooldown, and lock rules.

**Files:**
- Create: `services/core-api/internal/service/cbt_proctoring_policy.go`
- Test: `services/core-api/internal/service/cbt_proctoring_policy_test.go`

**Required functions:**
- `NormalizeProctorEventType(raw string) (eventType string, ok bool)`
- `ClassifyProctorSeverity(eventType string, data map[string]any) SeverityDecision`
- `ShouldDedup(eventType string) bool`
- `DedupWindow(eventType string) time.Duration`
- `RiskLevelFromScore(score int, violationCount int, lockedAt sql.NullTime) string`
- `ShouldAutoLock(decision SeverityDecision, participantState ParticipantRiskState) bool`

**Test cases:**
- focus lost short → warning or info, no lock
- one app switch → warning, no lock
- repeated app switch → medium/high after threshold
- split screen → medium
- device mismatch → critical
- offline mass → technical, no risk
- pending sync → technical/warning, no lock
- unknown event type → rejected

**Verification:**
```bash
cd services/core-api
go test ./internal/service -run 'TestCbtProctoringPolicy' -count=1
```
Expected: PASS.

## Task 1.4 — Integrate policy into mobile/student telemetry endpoint

**Objective:** When app reports anti-cheat telemetry, backend records normalized event and updates participant risk safely.

**Files:**
- Modify: `services/core-api/internal/service/cbt_session.go`
- Modify: `services/core-api/internal/handler/cbt_session.go`
- Modify tests: `services/core-api/internal/handler/cbt_session_success_test.go` or matching existing test file

**Rules:**
- Whitelist event types.
- Reject oversized JSON body using `http.MaxBytesReader` if not already enforced.
- Never trust client-supplied severity/risk score.
- Dedup repeated lifecycle events.
- Technical events do not increase cheating risk.
- Store event_data after schema-safe normalization.
- Update `cbt_exam_participants.risk_score`, `risk_level`, `violation_count`, `locked_at`, `locked_reason` in service/transaction.

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service -run 'Test.*Cbt.*Proctor|Test.*AntiCheat|Test.*Telemetry' -count=1
```
Expected: PASS. If test names differ, run full package tests.

## Task 1.5 — Full Sprint 1 validation

**Commands:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-sprint1 ./cmd/api
```

**Commit:**
```bash
git add services/core-api/db/migrations/114_cbt_proctoring_policy_actions.sql services/core-api/db/queries/cbt_proctoring.sql services/core-api/internal/service services/core-api/internal/handler services/core-api/internal/repository/postgres docs/contracts/cbt-byod-proctoring-realtime-audio.md
git commit -m "feat(cbt): add byod proctoring risk policy"
```

---

# Sprint 2 — Portal Pengawas Realtime

**Goal:** Pengawas ruang dan panitia melihat status peserta realtime, kelompok risiko, dan daftar butuh tindakan.

## Task 2.1 — Backend dashboard payload per session and room

**Objective:** Provide dashboard payload grouped by status, severity, room, and participant.

**Files:**
- Modify: `services/core-api/db/queries/cbt_proctoring.sql`
- Modify: `services/core-api/internal/service/cbt_session.go` or create `cbt_proctoring.go`
- Modify: `services/core-api/internal/handler/cbt_session.go`

**Endpoints:**
- `GET /api/asesmen/sessions/{id}/proctoring/live-summary`
- `GET /api/asesmen/sessions/{id}/rooms/{rid}/proctoring/live-summary`

**Payload:**
- session info
- room summaries
- counts:
  - normal
  - warning
  - medium/high
  - critical/locked
  - technical
  - offline
  - submitted
  - pending sync
- participants sorted by action priority
- latest events
- unacknowledged alarm count

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service -run 'Test.*Proctoring.*Summary' -count=1
```

## Task 2.2 — SSE stream with polling fallback

**Objective:** Add realtime stream without breaking current polling.

**Files:**
- Modify: `services/core-api/internal/handler/cbt_session.go`
- Modify or create BFF route in web-admin:
  - `apps/web-admin/src/routes/api/asesmen/sessions/[id]/proctoring/live-summary/+server.ts`
  - `apps/web-admin/src/routes/api/asesmen/sessions/[id]/rooms/[rid]/proctoring/live-summary/+server.ts`

**Endpoint:**
- `GET /api/asesmen/sessions/{id}/proctoring/stream`
- `GET /api/asesmen/sessions/{id}/rooms/{rid}/proctoring/stream`

**Rules:**
- SSE sends compact event update every 2–5 seconds or when new event arrives.
- UI fallback uses polling every 5 seconds if SSE fails.
- Do not expose secrets/tokens beyond existing authorized proctor view.

**Verification:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api && go test ./internal/handler -run 'Test.*SSE|Test.*Stream|Test.*Proctoring' -count=1
```

## Task 2.3 — Refactor session proctoring page into operational dashboard

**Objective:** Make `/asesmen/sesi/[id]/proctoring` a command center for all rooms.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- Create reusable components if useful:
  - `apps/web-admin/src/lib/components/cbt/proctoring/ProctorStatusCards.svelte`
  - `apps/web-admin/src/lib/components/cbt/proctoring/ProctorParticipantTable.svelte`
  - `apps/web-admin/src/lib/components/cbt/proctoring/ProctorEventTimeline.svelte`

**UI sections:**
- header: session title, status, time, live mode badge
- room status cards
- “Butuh Tindakan” panel
- participant table with filters:
  - Semua
  - Butuh Tindakan
  - Waspada
  - Pelanggaran
  - Kritis
  - Teknis
  - Selesai
- latest event timeline
- clear Indonesian labels:
  - `heartbeat` → Status koneksi
  - `device mismatch` → Perangkat tidak sesuai
  - `locked` → Akses ditahan
  - `pending sync` → Jawaban belum terkirim

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

## Task 2.4 — Refactor room proctoring page for pengawas ruang

**Objective:** Make `/asesmen/sesi/[id]/rooms/[rid]/proctoring` focused for one room.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**UI sections:**
- token ruang prominent
- status room/session
- cards:
  - belum login
  - sedang ujian
  - butuh tindakan
  - akses ditahan
  - jawaban belum sinkron
  - sudah submit
- list “Butuh Tindakan” always top
- participant detail drawer/panel:
  - timeline event
  - risk reason
  - connection status
  - answer sync status
  - action buttons placeholder for Sprint 4

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

## Task 2.5 — Full Sprint 2 validation and commit

**Commands:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-sprint2 ./cmd/api
```

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen/sesi apps/web-admin/src/routes/api/asesmen apps/web-admin/src/lib/components/cbt services/core-api
git commit -m "feat(cbt): add realtime proctoring dashboard"
```

---

# Sprint 3 — Audio Warning & Alarm Acknowledge

**Goal:** Portal pengawas berbunyi hanya untuk event penting, bisa dimute, throttle, dan wajib acknowledge untuk alarm penting.

## Task 3.1 — Audio alert utility

**Objective:** Create safe browser audio utility with user opt-in.

**Files:**
- Create: `apps/web-admin/src/lib/cbt/proctor-audio.ts`
- Test if project has TS tests pattern: `apps/web-admin/src/lib/cbt/proctor-audio.test.ts`

**Functions:**
- `canPlayProctorAudio()`
- `playProctorTone(level: 'warning' | 'medium' | 'critical' | 'technical')`
- `shouldThrottleAudio(key: string, level: string, now: number)`
- `audioLabel(level)`

**Rules:**
- Audio disabled by default until user clicks “Aktifkan Suara”.
- Warning: soft single ding.
- Medium: two-tone warning.
- Critical: short alarm.
- Technical: distinct non-cheating technical tone.
- Throttle per participant/event.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

## Task 3.2 — Alarm tray component

**Objective:** Show unacknowledged important events with clear priority.

**Files:**
- Create: `apps/web-admin/src/lib/components/cbt/proctoring/ProctorAlarmTray.svelte`
- Modify:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**UI behavior:**
- Shows medium/critical/technical events not acknowledged.
- Critical pinned at top.
- Button: “Acknowledge / Sudah dicek”.
- Requires note for critical and technical mass issue.
- Audio button states:
  - Suara Nonaktif
  - Suara Aktif
  - Senyapkan 10 Menit

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

## Task 3.3 — Backend acknowledge endpoint

**Objective:** Persist acknowledge state.

**Files:**
- Modify: `services/core-api/db/queries/cbt_proctoring.sql`
- Modify: `services/core-api/internal/service/cbt_proctoring.go`
- Modify: `services/core-api/internal/handler/cbt_session.go`
- Add/modify tests.

**Endpoint:**
- `POST /api/asesmen/proctoring/events/{event_id}/acknowledge`

**Body:**
```json
{
  "note": "Sudah diperiksa pengawas"
}
```

**Rules:**
- Must require authenticated proctor/admin actor.
- Critical event requires non-empty note.
- Store actor user ID if available.
- Idempotent: already acknowledged returns OK with existing state.

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service -run 'Test.*Acknowledge.*Proctor' -count=1
```

## Task 3.4 — Full Sprint 3 validation and commit

**Commands:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-sprint3 ./cmd/api
```

**Commit:**
```bash
git add apps/web-admin/src/lib/cbt/proctor-audio.ts apps/web-admin/src/lib/components/cbt/proctoring apps/web-admin/src/routes/asesmen/sesi services/core-api
git commit -m "feat(cbt): add proctor audio alerts"
```

---

# Sprint 4 — Tindakan Pengawas & Audit

**Goal:** Pengawas bisa memberi warning, tahan/buka akses, reset perangkat, force submit, catat insiden, dan eskalasi dengan audit lengkap.

## Task 4.1 — Define action types and validation

**Objective:** Backend action policy is explicit and safe.

**Files:**
- Create/modify: `services/core-api/internal/service/cbt_proctoring_actions.go`
- Test: `services/core-api/internal/service/cbt_proctoring_actions_test.go`

**Action types:**
- `warn_student`
- `hold_access`
- `unlock_access`
- `reset_device_binding`
- `force_submit`
- `mark_technical_issue`
- `mark_incident`
- `escalate_to_committee`
- `clear_after_check`

**Validation:**
- All actions require actor.
- High-impact actions require reason and notes:
  - unlock_access
  - reset_device_binding
  - force_submit
  - hold_access
  - escalate_to_committee
- `force_submit` must check answer sync/pending state.
- `reset_device_binding` must clear device fingerprint only with reason.
- `mark_technical_issue` must not increase risk score.

**Verification:**
```bash
cd services/core-api
go test ./internal/service -run 'Test.*Proctor.*Action' -count=1
```

## Task 4.2 — Backend action endpoint

**Objective:** Add one safe endpoint for proctor actions.

**Files:**
- Modify: `services/core-api/internal/handler/cbt_session.go`
- Modify: `services/core-api/internal/service/cbt_proctoring_actions.go`
- Modify: `services/core-api/db/queries/cbt_proctoring.sql`

**Endpoint:**
- `POST /api/asesmen/sessions/{session_id}/participants/{participant_id}/proctor-actions`

**Body:**
```json
{
  "action_type": "unlock_access",
  "reason": "Gangguan teknis perangkat",
  "notes": "Peserta tidak sengaja keluar aplikasi saat jaringan putus. Sudah diperiksa."
}
```

**Response:**
- participant current risk/access state
- created action record
- latest timeline event/action

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service -run 'Test.*Proctor.*Action' -count=1
```

## Task 4.3 — Action dialog UI in participant detail

**Objective:** Make actions accessible but controlled from room/session proctoring pages.

**Files:**
- Create: `apps/web-admin/src/lib/components/cbt/proctoring/ProctorActionDialog.svelte`
- Modify:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**UI rules:**
- Group actions:
  - Ringan: beri peringatan, tandai dicek
  - Akses: tahan akses, buka akses
  - Perangkat: reset perangkat/token binding
  - Teknis: tandai gangguan teknis
  - Berat: eskalasi, force submit
- Show risk context before destructive action.
- Require typed notes for critical actions.
- After action, refresh dashboard and timeline.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

## Task 4.4 — Proctor action timeline and audit labels

**Objective:** Merge events and actions into a readable timeline.

**Files:**
- Modify/create: `apps/web-admin/src/lib/components/cbt/proctoring/ProctorEventTimeline.svelte`
- Modify backend summary endpoint to include recent actions.

**Timeline labels:**
- “Pengawas memberi peringatan”
- “Akses ditahan”
- “Akses dibuka oleh pengawas”
- “Perangkat direset”
- “Ditandai gangguan teknis”
- “Dieskalasi ke panitia”
- “Kirim paksa dengan catatan”

**Verification:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api && go test ./internal/handler ./internal/service ./internal/repository/postgres
```

## Task 4.5 — Full Sprint 4 validation and commit

**Commands:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-sprint4 ./cmd/api
```

**Commit:**
```bash
git add apps/web-admin/src/lib/components/cbt/proctoring apps/web-admin/src/routes/asesmen/sesi services/core-api
git commit -m "feat(cbt): add audited proctor actions"
```

---

# Sprint 5 — Berita Acara, Export, dan Command Center Panitia

**Goal:** Semua event, alarm, dan tindakan pengawas bisa direkap sebagai berita acara ruang/sesi/kegiatan.

## Task 5.1 — Report/export backend payload

**Objective:** Produce audit-ready report data.

**Files:**
- Modify: `services/core-api/db/queries/cbt_proctoring.sql`
- Modify/create: `services/core-api/internal/service/cbt_proctoring_report.go`
- Modify: `services/core-api/internal/handler/cbt_session.go`

**Endpoints:**
- `GET /api/asesmen/sessions/{id}/proctoring/report-data`
- `GET /api/asesmen/sessions/{id}/rooms/{rid}/proctoring/report-data`

**Payload sections:**
- session/event identity
- room identity
- proctors
- participant summary
- severity summary
- technical incident summary
- locked/unlocked history
- proctor actions
- unacknowledged events
- submitted/not submitted
- sync/pending anomalies

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service -run 'Test.*Proctor.*Report' -count=1
```

## Task 5.2 — Update report pages

**Objective:** Make report pages print/export friendly.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/report/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`

**UI sections:**
- cover summary
- peserta bermasalah
- pelanggaran by severity
- gangguan teknis
- tindakan pengawas
- catatan panitia
- signature area:
  - Pengawas 1
  - Pengawas 2
  - Ketua Panitia/Operator

**Actions:**
- print
- export CSV
- export JSON evidence

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

## Task 5.3 — Command center overview by event/kegiatan

**Objective:** Panitia can monitor all active rooms/sessions.

**Files:**
- Modify/create: `apps/web-admin/src/routes/asesmen/pengawasan/+page.svelte`
- Add BFF if needed under `apps/web-admin/src/routes/api/asesmen/proctoring/overview/+server.ts`
- Backend endpoint if missing:
  - `GET /api/asesmen/proctoring/overview?event_id=...`

**UI:**
- Kegiatan active selector
- Room cards across sessions
- Critical participants across rooms
- Technical mass issue panel
- Latest alarm feed
- Link to room detail

**Verification:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api && go test ./internal/handler ./internal/service ./internal/repository/postgres
```

## Task 5.4 — Full Sprint 5 validation and commit

**Commands:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-sprint5 ./cmd/api
```

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen apps/web-admin/src/routes/api/asesmen services/core-api docs/contracts/cbt-byod-proctoring-realtime-audio.md
git commit -m "feat(cbt): add proctoring reports and command center"
```

---

# Sprint 6 — Production Rollout & SOP

**Goal:** Safely deploy proctoring improvements only after explicit approval.

## Task 6.1 — Pre-deploy audit

**Objective:** Capture current state before migration/deploy.

**Commands:**
```bash
git status --short
git log --oneline -5
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-proctoring-final ./cmd/api
```

**DB audit queries:**
- count active CBT sessions
- count participants by risk_level
- count locked participants
- count recent events/actions
- count rooms without proctor

Do not display secrets.

## Task 6.2 — Backup before migration

**Objective:** Ensure rollback path.

**Command pattern:**
```bash
mkdir -p /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql
set -a; source .env >/dev/null 2>&1; set +a
pg_dump "$DATABASE_URL" -Fc -f /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-cbt-proctoring-$(date +%Y%m%d-%H%M%S).dump
sha256sum /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-cbt-proctoring-*.dump | tail -1
```

Only run with explicit production approval.

## Task 6.3 — Migration order

**Order:**
1. Apply migration `114_cbt_proctoring_policy_actions.sql`.
2. Build core-api binary.
3. Restart core-api.
4. Health check.
5. Build web-admin.
6. Restart web-admin immediately after build to avoid stale manifest/missing chunk errors.
7. Smoke test admin pages.

**Commands pattern:**
```bash
# Core API
cd services/core-api
set -a; source ../../.env >/dev/null 2>&1; set +a
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/migrations/114_cbt_proctoring_policy_actions.sql
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
curl -fsS http://127.0.0.1:8080/health

# Web Admin
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run build
pm2 restart mtsn2kolut-web-admin --update-env
```

**Important:** Production deploy/restart only after user says explicit approval phrase, for example: `SETUJU DEPLOY PORTAL PENGAWAS`.

## Task 6.4 — Smoke tests

**Manual smoke:**
- Open `/asesmen/pengawasan`.
- Open one session proctoring page.
- Open one room proctoring page.
- Enable audio and trigger test/fixture event in safe non-production/test session only.
- Acknowledge alarm.
- Record one proctor action with note.
- Print/export report.

**Expected:**
- no JS console errors
- no 500 backend
- no credential leak
- audio only after user opt-in
- action audit visible in timeline/report

---

# Acceptance Criteria Final

Plan is complete when all are true:

- Backend whitelists and classifies proctor events.
- Risk score has dedup/cooldown.
- Technical issue is not treated as cheating.
- Portal shows realtime session and room monitoring.
- Audio warning exists per level and can be muted.
- Medium/critical/technical alarms can be acknowledged.
- High-impact proctor actions require reason and are audited.
- Reports include events, actions, technical issues, and signature area.
- Full validation passes:
  - `npm --prefix apps/web-admin run check`
  - `sqlc generate`
  - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
  - `go build`
- No production deploy/restart without explicit approval.

---

# Recommended Execution Order

1. Sprint 0: contract and audit.
2. Sprint 1: backend policy and event normalization.
3. Sprint 2: realtime portal.
4. Sprint 3: audio and acknowledge.
5. Sprint 4: proctor actions and audit.
6. Sprint 5: report/export and command center.
7. Sprint 6: approved production rollout.

Do not combine Sprint 1–5 into one massive commit. Each sprint should have its own verification and commit.
