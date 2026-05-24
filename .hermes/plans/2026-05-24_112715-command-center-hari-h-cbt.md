# Command Center Hari-H CBT Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Membangun Command Center Hari-H CBT untuk memantau dan mengendalikan 8 ruang UAS Genap dari satu layar yang sederhana, audit-safe, mobile-friendly, dan cepat dipakai operator madrasah.

**Architecture:** Tambahkan lapisan UI operasional di detail sesi CBT yang menggabungkan ringkasan global, grid 8 ruang, panel masalah aktif, dan aksi cepat yang mengarah ke route proctoring/print/handover existing. Implementasi awal memakai data existing dari detail sesi/proctoring; bila multi-fetch atau data gap terasa berat, tambahkan endpoint agregat backend `GET /api/cbt/sessions/{id}/command-center` melalui BFF `/api/asesmen/sessions/{id}/command-center`.

**Tech Stack:** SvelteKit/Svelte 5 web-admin, shadcn-svelte-style components existing, Svelte/Vitest tests, Go Chi backend + sqlc PostgreSQL jika endpoint baru dibutuhkan, PM2 deployment.

---

## Current Context

- Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
- Frontend utama:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/session-detail.model.ts`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/print-pack/+page.svelte`
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`
- Backend terkait:
  - `services/core-api/internal/handler/cbt_session.go`
  - `services/core-api/internal/service/cbt_session.go`
  - `services/core-api/db/queries/*cbt*`
- Kondisi data UAS saat ini:
  - 105 sesi UAS Genap.
  - Tiap sesi punya 8 ruang: Ruang Ujian 01 s.d. Ruang Ujian 08.
  - Total baris ruang UAS: 840.
- Prinsip user:
  - audit-safe;
  - UI sederhana untuk madrasah;
  - opsi/plan sebelum edit;
  - backup sebelum DB write;
  - test/build sebelum deploy;
  - jangan bocorkan token/password/cookie/credential.

---

## UX Target

Command Center harus menjawab dalam <= 5 detik:

1. Sesi ini sedang apa?
2. 8 ruang mana yang aman dan mana yang perlu atensi?
3. Berapa peserta login, online, submit, offline, atau bermasalah?
4. Ruang mana yang belum final/handover?
5. Operator harus klik apa dulu?

### Desktop layout

1. Sticky header sesi.
2. Summary cards global.
3. Active Issue Panel.
4. Grid 8 kartu ruang, ideal 4 kolom x 2 baris.
5. Recent event feed / rekap penutupan.

### Mobile layout

1. Sticky compact header.
2. Summary strip horizontal.
3. Panel masalah aktif.
4. List ruang, ruang bermasalah di atas.
5. Bottom quick nav: `Masalah`, `Ruang`, `Log`, `Rekap`.

---

## Data Model / Derived State

Minimal Command Center room card:

```ts
type CommandCenterRoomCard = {
  room_id: string;
  room_name: string;
  room_status: 'setup' | 'ready' | 'running' | 'attention' | 'critical' | 'finishing' | 'final';
  participant_count: number;
  joined_count: number;
  online_count?: number;
  submitted_count: number;
  no_show_count: number;
  suspicious_count: number;
  incident_event_count: number;
  missing_seat_count: number;
  force_submit_count: number;
  reset_access_count: number;
  proctor_count?: number;
  primary_proctor_name?: string | null;
  allow_web_fallback?: boolean;
  handover_locked?: boolean;
  handover_updated_at?: string | null;
  latest_event_at?: string | null;
};
```

Minimal global summary:

```ts
type CommandCenterSummary = {
  room_count: number;
  rooms_ready_count: number;
  participant_count: number;
  joined_count: number;
  submitted_count: number;
  online_count?: number;
  no_show_count: number;
  suspicious_count: number;
  incident_event_count: number;
  force_submit_count: number;
  reset_access_count: number;
  handover_locked_count: number;
  handover_missing_count: number;
  active_issue_count: number;
  last_updated_at: string;
  live_mode: 'live' | 'polling' | 'stale';
};
```

Active issue:

```ts
type CommandCenterIssue = {
  id: string;
  severity: 'critical' | 'warning' | 'info';
  scope: 'session' | 'room' | 'participant';
  room_id?: string;
  participant_id?: string;
  title: string;
  description: string;
  primary_action_label: string;
  primary_action_href?: string;
  requires_note?: boolean;
};
```

---

## Phase 0 — Safe Discovery / Contract Freeze

### Task 0.1: Confirm existing data source and routes

**Objective:** Pastikan apakah Command Center bisa memakai data existing tanpa endpoint baru.

**Files:**
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/session-detail.model.ts`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/+server.ts`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- Read: `services/core-api/internal/handler/cbt_session.go`
- Read: `services/core-api/internal/service/cbt_session.go`

**Steps:**
1. Identify props/data already loaded in session detail page.
2. List all room-level counters already available.
3. Identify missing counters: online, heartbeat stale, fallback active, proctor names, handover status.
4. Decide:
   - If >=80% data is already present, implement frontend derived state first.
   - If major counters require many endpoint calls, implement aggregate endpoint first.

**Verification:** Produce a short note in implementation PR describing chosen data path.

---

### Task 0.2: Security rule for token masking

**Objective:** Pastikan Command Center tidak membocorkan raw room token.

**Files:**
- Read/modify later: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- Potential create: `apps/web-admin/src/lib/utils/token-mask.ts`
- Potential test: `apps/web-admin/src/lib/utils/token-mask.test.ts`

**Rules:**
- Raw `room_token` must not render in Command Center.
- Default view uses masked token or no token.
- Full token only allowed in print-pack/reveal flow with explicit permission.
- No token in toast/error/URL/console.

---

## Phase 1 — TDD for Frontend Command Center

### Task 1.1: Add static/source regression test for Command Center shell

**Objective:** Create failing test that requires Command Center labels/components to exist.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/command-center-ux.test.ts`

**Test assertions:**

```ts
import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const page = readFileSync('src/routes/asesmen/sesi/[id]/+page.svelte', 'utf8');

describe('CBT Command Center UX', () => {
  it('exposes a Hari-H command center shell', () => {
    expect(page).toContain('Command Center Hari-H');
    expect(page).toContain('Ruang siap');
    expect(page).toContain('Masalah Aktif');
    expect(page).toContain('Grid 8 Ruang');
  });

  it('does not render raw room token in operational recap labels', () => {
    expect(page).not.toContain('Token rahasia {room.room_token');
    expect(page).not.toContain('{room.room_token ||');
  });
});
```

**Run:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npx vitest run src/routes/asesmen/command-center-ux.test.ts
```

**Expected:** FAIL until Command Center shell/masking is implemented.

---

### Task 1.2: Add utility tests for status derivation

**Objective:** Define deterministic room status rules.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/command-center.model.ts`
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/command-center.model.test.ts`

**Status rules:**
- `critical`: locked participant, offline not submitted near end, over capacity, no proctor while active.
- `attention`: suspicious_count > 0, incident_event_count > 0, missing_seat_count > 0, fallback active.
- `finishing`: time remaining <= 10 minutes and submitted_count < participant_count.
- `final`: handover_locked.
- `running`: active with no major issue.
- `ready`: before start and setup complete.
- `setup`: before start but missing rooms/proctors/seats.

**Test examples:**

```ts
expect(deriveRoomStatus({ participant_count: 24, submitted_count: 24, handover_locked: true })).toBe('final');
expect(deriveRoomStatus({ participant_count: 24, submitted_count: 12, suspicious_count: 2 })).toBe('attention');
expect(deriveRoomStatus({ participant_count: 24, missing_seat_count: 1 })).toBe('attention');
```

**Run:**

```bash
npx vitest run src/routes/asesmen/sesi/[id]/command-center.model.test.ts
```

---

### Task 1.3: Add issue prioritization tests

**Objective:** Make Active Issue Panel predictable.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/command-center.model.ts`
- Modify test: `apps/web-admin/src/routes/asesmen/sesi/[id]/command-center.model.test.ts`

**Rules:**
1. locked/cannot continue
2. offline not submitted
3. near end not submitted
4. suspicious/anti-cheat attention
5. handover missing after finish
6. setup blockers before start

**Verification:** Unit test sorted issue array.

---

## Phase 2 — UI Components

### Task 2.1: Create `CommandCenterHeader` component

**Objective:** Sticky header with session name, status, schedule, live/polling/stale label, and refresh.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/CommandCenterHeader.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`

**UI copy:**
- `Command Center Hari-H`
- `Terakhir diperbarui ...`
- `Live`, `Polling`, or `Data mungkin stale`
- `Refresh`
- `Buka Panel Pengawasan`

**Verification:** `npm run check`.

---

### Task 2.2: Create `SessionLiveSummaryCards`

**Objective:** Render top 5 global cards.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/SessionLiveSummaryCards.svelte`
- Modify: `+page.svelte`

**Cards:**
1. `Login Peserta`
2. `Jawaban Terkirim`
3. `Terputus/Lambat`
4. `Perlu Atensi`
5. `Serah Terima Ruang`

**Acceptance:** Summary visible without opening tabs.

---

### Task 2.3: Create `ActiveIssuePanel`

**Objective:** Show prioritized issue list with clear next action.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/ActiveIssuePanel.svelte`
- Modify: `+page.svelte`

**Empty state:**

```text
Tidak ada masalah aktif. Pantau login, koneksi, dan progres submit secara berkala.
```

**Issue item format:**
- severity badge;
- room/participant scope;
- short detail;
- action button `Buka Ruang`, `Cek Belum Submit`, `Lengkapi Handover`.

**Acceptance:** Problems appear above the room grid.

---

### Task 2.4: Create `RoomStatusGrid` and `RoomStatusCard`

**Objective:** Render 8 cards in desktop grid and mobile list.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/RoomStatusGrid.svelte`
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/RoomStatusCard.svelte`
- Modify: `+page.svelte`

**Room card minimum fields:**
- room name;
- status badge;
- peserta count;
- login count;
- submit count;
- atensi count;
- pengawas;
- handover status;
- buttons: `Buka Ruang`, `Log`, `Cetak`, `Handover`.

**Desktop:** `grid-cols-1 md:grid-cols-2 xl:grid-cols-4`.

**Mobile:** cards sorted with `critical/attention` first.

---

### Task 2.5: Add `RecentEventFeed` / audit strip

**Objective:** Show latest operational events without overwhelming operator.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/RecentEventFeed.svelte`
- Modify: `+page.svelte`

**Content:**
- Latest app switch / screenshot attempt / reset access / force submit / handover updated.
- No raw tokens.

**Acceptance:** Operator sees recent activity trail.

---

## Phase 3 — Integrate into Detail Sesi Page

### Task 3.1: Add `commandCenter` derived state in `+page.svelte`

**Objective:** Map existing `session`, `rooms`, `operationalRecap`, and `participants` into command-center model.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- Modify/create: `apps/web-admin/src/routes/asesmen/sesi/[id]/command-center.model.ts`

**Steps:**
1. Import model helpers.
2. Build `commandCenterSummary`.
3. Build `commandCenterRooms`.
4. Build `commandCenterIssues`.
5. Mask tokens before passing data into components.

**Acceptance:** No backend change required if data complete enough.

---

### Task 3.2: Insert Command Center section in page

**Objective:** Add Command Center as a top-level tab/section, preferably default when session is active or close to start.

**Files:**
- Modify: `+page.svelte`

**Recommendation:**
- Add tab label: `Command Center`.
- If status is `active`, default selected tab should be Command Center.
- Keep existing tabs for `Ruangan`, `Peserta`, `Operasional`, `Log`, etc.

**Acceptance:** Existing workflows remain accessible; Command Center is an overlay/aggregation, not a destructive rewrite.

---

### Task 3.3: Hide dangerous setup actions from Command Center

**Objective:** Ensure Command Center focuses on monitoring and action, not setup mutation.

**Files:**
- Modify: `+page.svelte` if actions are shared.

**Rules:**
- No `Hapus Ruang` in Command Center.
- No `Acak Peserta` in active Command Center.
- No `Ubah kapasitas` during running session.
- Sensitive actions link to existing dialogs/routes with note requirement.

---

## Phase 4 — Optional Backend Aggregate Endpoint

Only do this if Phase 0 finds current data requires many client fetches or lacks counters.

### Task 4.1: Define backend response contract test

**Objective:** Add failing test for `GET /api/cbt/sessions/{id}/command-center`.

**Files:**
- Create: `services/core-api/internal/handler/cbt_command_center_test.go`
- Create: `services/core-api/internal/service/cbt_command_center_test.go`

**Must assert:**
- returns summary;
- returns 8 rooms;
- no raw token;
- permission check;
- proctor sees only assigned rooms if scoped.

**Run:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
go test ./internal/service ./internal/handler -run 'CommandCenter|CbtCommandCenter' -count=1
```

---

### Task 4.2: Add service/query implementation

**Objective:** Aggregate room and participant counters in backend.

**Files:**
- Modify/create query: `services/core-api/db/queries/cbt_command_center.sql`
- Run sqlc: `cd services/core-api && sqlc generate -f db/sqlc.yaml`
- Create/modify service: `services/core-api/internal/service/cbt_command_center.go`
- Modify handler: `services/core-api/internal/handler/cbt_session.go` or new handler file.
- Modify router: `services/core-api/cmd/api/main.go`

**Security:** Do not select raw room token unless reveal endpoint.

---

### Task 4.3: Add SvelteKit BFF proxy

**Objective:** Expose backend command-center endpoint to frontend.

**Files:**
- Create: `apps/web-admin/src/routes/api/asesmen/sessions/[id]/command-center/+server.ts`
- Or follow existing proxy pattern under `apps/web-admin/src/lib/server/cbt-backend-proxy/...`

**Acceptance:** Auth/session forwarded like existing asesmen routes; no sensitive field added.

---

## Phase 5 — Token Masking and Sensitive Actions

### Task 5.1: Add token masking utility

**Objective:** Centralize masking logic.

**Files:**
- Create: `apps/web-admin/src/lib/utils/token-mask.ts`
- Create: `apps/web-admin/src/lib/utils/token-mask.test.ts`

**Example behavior:**

```ts
maskToken('abcdef1234567890') // 'abcd••••7890'
maskToken('short') // '••••'
```

**Rule:** Never use CSS-only hiding for secrets.

---

### Task 5.2: Replace visible raw room token in operational recap

**Objective:** Fix known token display risk.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`

**Acceptance:** Search must not find direct render of `room.room_token` in page body except passed to print/reveal-safe route if needed.

---

### Task 5.3: Standardize sensitive action dialogs

**Objective:** Ensure actions like reset/force submit/reveal token require note and clear consequence.

**Files:**
- Inspect/modify existing proctoring action code:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- Potential create shared component:
  - `apps/web-admin/src/routes/asesmen/sesi/[id]/SensitiveActionDialog.svelte`

**Acceptance:** No one-click destructive action in Command Center.

---

## Phase 6 — Print / Handover / End-of-Session Flow

### Task 6.1: Add `Cetak Semua Paket Ruang` entry from Command Center

**Objective:** Improve 8-room workflow.

**Files:**
- Modify Command Center UI components.
- Potential route future: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/print-pack/+page.svelte`

**MVP:** Dialog listing 8 print-pack links.

**Future:** Generate combined PDF/page.

---

### Task 6.2: Add handover progress panel

**Objective:** Make after-session finalization clear.

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/sesi/[id]/HandoverProgressPanel.svelte`
- Modify: `+page.svelte`

**UI:**
- `2/8 ruang final`
- list rooms not final;
- link `Lengkapi Handover`.

---

## Phase 7 — Tests and Verification

### Frontend targeted tests

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npx vitest run src/routes/asesmen/command-center-ux.test.ts
npx vitest run src/routes/asesmen/sesi/[id]/command-center.model.test.ts
npm run check
npm run build
```

### Backend tests if endpoint added

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
go test ./internal/service ./internal/handler -run 'CommandCenter|CbtCommandCenter' -count=1
go test ./internal/service ./internal/handler -count=1
```

### Static secret scan before merge

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
grep -RIn --exclude-dir=node_modules --exclude-dir=.svelte-kit --exclude-dir=build \
  -E 'console\.log|console\.debug|room_token|rawToken|token.*audit|audit.*token|answer_key|correct_answer|kunci_jawaban' \
  apps/web-admin/src services/core-api/internal
```

Review manually; not every match is a bug, but no raw token should render in Command Center.

### Browser smoke

Existing:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run smoke:cbt:roles
```

Future script recommendation:

```bash
npm run smoke:cbt:command-center
```

Smoke acceptance:
- admin sees Command Center;
- 8 room cards visible;
- room issue links work;
- raw token not visible in body text before reveal;
- guru/proctor sees only allowed scope;
- protected routes redirect/403 correctly.

---

## Phase 8 — Deployment / Rollout

### Pre-deploy

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
go test ./internal/service ./internal/handler -count=1 # from services/core-api if backend touched
```

If DB migration or data write is needed:

```bash
bash deploy/backup-postgresql.sh
```

### Deploy order

1. Build frontend.
2. If backend touched: build/test backend and restart `mtsn2kolut-core-api`.
3. Restart web-admin immediately after frontend build to avoid stale chunks.
4. Health check.
5. Browser smoke.

Commands used in this environment:

```bash
pm2 restart mtsn2kolut-core-api   # only if backend touched
pm2 restart mtsn2kolut-web-admin
curl -fsS http://127.0.0.1:8080/health
```

---

## Risks and Tradeoffs

1. **Token leak risk**
   - Mitigation: mask at serializer/derived-state boundary, tests, static scan.
2. **Command Center becomes too crowded**
   - Mitigation: only show summary + issues + 8 cards; move full details to room route.
3. **Polling load**
   - Mitigation: start with manual refresh / 15s polling; endpoint aggregate if needed.
4. **Role scope bug**
   - Mitigation: backend permission tests if endpoint new; smoke admin vs guru/proctor.
5. **Over-refactor risk**
   - Mitigation: keep existing routes and tabs; Command Center is additive.
6. **Data gap for online/heartbeat**
   - Mitigation: show available counters first; mark unavailable as `Belum tersedia`, add endpoint later.

---

## Acceptance Criteria

### Information

- [ ] Operator sees `Command Center Hari-H` in detail sesi.
- [ ] Operator sees total 8 ruang in one desktop screen.
- [ ] Summary cards show login, submit, atensi, fallback/offline, handover.
- [ ] Each room card shows room name, peserta, login, submit, atensi, pengawas, handover, and quick actions.
- [ ] Status has text label, not color-only.

### Workflow

- [ ] Before start: blockers setup appear first.
- [ ] During running: active issues appear above grid.
- [ ] Near end: not-submitted/offline issues are prioritized.
- [ ] After finish: rooms without handover/finalization appear first.
- [ ] One click/tap opens room proctoring.

### Security / Audit

- [ ] Raw room token not rendered in Command Center.
- [ ] Sensitive actions use confirmation and note.
- [ ] No raw token in toast/error/URL/console/audit metadata.
- [ ] Role scope respected.

### Mobile

- [ ] Mobile layout uses card list, not horizontal table.
- [ ] Problem rooms sorted first.
- [ ] Sticky header shows status and refresh.
- [ ] Sensitive dialogs usable on mobile.

### Verification

- [ ] Targeted Vitest passes.
- [ ] `npm run check` passes.
- [ ] `npm run build` passes.
- [ ] Backend tests pass if backend touched.
- [ ] Health check OK after restart.
- [ ] Browser smoke passes or manual smoke documented.

---

## Suggested Implementation Order

1. Model helpers + tests.
2. Token masking utility + tests.
3. Command Center shell + static UX test.
4. Summary cards.
5. Active Issue Panel.
6. Room grid/cards.
7. Integrate into detail sesi default tab behavior.
8. Handover and print quick links.
9. Optional aggregate backend endpoint if needed.
10. Smoke tests and deploy.

---

## Open Questions

1. Apakah Command Center harus menjadi tab baru di detail sesi, atau route khusus seperti `/asesmen/sesi/[id]/command-center`?
   - Rekomendasi awal: tab/section di detail sesi agar tidak memecah workflow.
2. Apakah token reveal diperlukan di Command Center?
   - Rekomendasi awal: tidak. Link ke print-pack atau panel room yang sudah memiliki guard.
3. Apakah guru/pengawas akan memakai Command Center yang sama?
   - Rekomendasi: admin melihat semua 8 ruang; pengawas hanya ruang assigned.
4. Apakah live update cukup polling 15 detik?
   - Rekomendasi awal: polling 15 detik + manual refresh; realtime bisa fase lanjutan.
