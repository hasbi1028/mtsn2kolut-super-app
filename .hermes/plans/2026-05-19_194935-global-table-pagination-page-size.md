# Global Table Pagination + Page Size Implementation Plan

> **For Hermes:** Use subagent-driven-development and orchestrasi-5-agent skills to implement this plan task-by-task. Implementation must not start until the user explicitly approves this plan.

**Goal:** Terapkan pola pagination bernomor + pilihan jumlah data per halaman secara konsisten ke semua tabel/list besar di Web Admin MTsN 2 Kolut.

**Architecture:** Bangun komponen pagination reusable di Web Admin terlebih dahulu, lalu migrasikan tabel bertahap berdasarkan prioritas. Endpoint backend harus menerima kontrak pagination kompatibel tanpa mematahkan client lama. Rollout dilakukan lewat agent swarm/worktree per domain agar file yang disentuh tidak saling konflik.

**Tech Stack:** SvelteKit 2, Svelte 5 runes, shadcn-svelte UI primitives, Go Chi API, PostgreSQL/sqlc, PM2 deployment.

---

## 1. Current Context / Assumptions

- Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
- Branch aktif dari hasil swarm: `feature/comprehensive-improvements`, ahead dari remote.
- Working tree utama saat audit: bersih.
- User meminta plan dulu; jangan patch sebelum approval.
- Setelah implementasi disetujui:
  - format/tidy touched files,
  - jalankan check/build/test,
  - commit patch selesai,
  - deploy/restart PM2 hanya jika disetujui atau dibutuhkan untuk verifikasi.
- GitHub push masih blocker karena auth remote belum tersedia.
- Web Admin tidak boleh akses DB langsung; semua data melalui Go backend/BFF.
- BFF harus tetap kompatibel dengan `readClientApiData` yang sudah unwrap envelope `{ items: [...] }` / `{ data: ... }`.

---

## 2. Swarm Review Summary

5 agent review-only sudah dijalankan:

1. **Inventory Architect**
   - Menemukan sekitar 82 file Svelte dengan table/pagination indicators.
   - Sekitar 155 table instance lintas `routes` dan shared components.
   - Mengelompokkan tabel menjadi P0 foundation, P1 high-impact, P2 medium, P3 static/report.

2. **Shared Component Designer**
   - Merekomendasikan komponen reusable `TablePagination` / `DataPagination`.
   - Komponen tidak fetch data; hanya UI/state callback.
   - Parent menangani URL sync dan data load agar SSR-safe.

3. **Backend/API Contract Reviewer**
   - Menemukan dua kontrak aktif: `limit/offset` dan `page/per_page`.
   - Rekomendasi canonical internal: `limit/offset`, tetap terima alias `page/per_page`.
   - Jangan ubah endpoint array lama menjadi object secara breaking.

4. **UI/UX Rollout Reviewer**
   - Standar label Indonesia:
     - `Menampilkan 1–24 dari 371 soal`
     - `Tampilkan [24] per halaman`
     - `Awal`, `Sebelumnya`, `Berikutnya`, `Akhir`
   - Mobile stacked, desktop numbered pagination.

5. **QA/Execution Planner**
   - Eksekusi disarankan memakai worktree/branch per agent.
   - Integrasi dilakukan bertahap oleh integration lead.
   - Deploy/restart hanya setelah approval eksplisit.

---

## 3. Standard UX Rules

### 3.1 Page Size Options

Default universal:

```text
12, 24, 48, 96
```

Default per tipe halaman:

- Bank Soal / konten panjang / card-rich: default `12`
- Tabel admin umum: default `24`
- Audit/log/antrian: default `24` atau `48`
- Jangan expose `500`, `1000`, `2000` di UI interaktif.

### 3.2 Copywriting

Gunakan label seragam:

- Range:
  - `Menampilkan 1–24 dari 371 soal`
  - `Menampilkan 25–48 dari 371 pengguna`
  - `Tidak ada data yang sesuai`
- Selector:
  - `Tampilkan [24] per halaman`
- Page mobile:
  - `Halaman 2 dari 16`
- Buttons:
  - `Awal`
  - `Sebelumnya`
  - `Berikutnya`
  - `Akhir`

### 3.3 Behavior Rules

- Limit/page size adalah preferensi tampilan, bukan filter.
- Saat limit berubah:
  - reset `page = 1`
  - reload data dengan `offset = 0`
  - update URL bila halaman mendukung URL state.
- Saat filter/search/sort berubah:
  - reset `page = 1`
  - pertahankan `pageSize`
- Saat `Bersihkan filter`:
  - filter/search hilang,
  - `page = 1`,
  - `pageSize` tetap.
- URL invalid harus fallback aman:
  - `limit=abc`, `limit=0`, `limit=-1`, `limit=9999` -> default.
  - `page=-3`, `page=abc` -> `1`.
- Jika `page` melebihi total halaman setelah fetch:
  - clamp ke halaman terakhir valid atau reload page `1` sesuai konteks.
- Setelah delete/archive item terakhir di halaman terakhir:
  - jangan tinggalkan halaman kosong palsu.

### 3.4 Desktop Layout

```text
Menampilkan 1–24 dari 371 soal
Tampilkan [24 ▼] per halaman

[Awal] [Sebelumnya] [1] [2] [3] […] [16] [Berikutnya] [Akhir]
```

Recommended classes conceptually:

- footer root: `rounded-xl border border-border bg-card p-3 text-sm shadow-sm`
- layout: `flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between`
- numbered buttons visible on `sm`/`md` and up.

### 3.5 Mobile Layout

```text
Menampilkan 1–12 dari 371 soal
Halaman 1 dari 31

Tampilkan [12 ▼] per halaman

[Sebelumnya]        [Berikutnya]
```

Rules:

- Hide full numbered pagination on small screens.
- Use stacked layout.
- Buttons minimum touch height: `min-h-10` or `min-h-11`.
- Avoid horizontal overflow.

### 3.6 Accessibility

- Footer wrapper: `nav aria-label="Navigasi halaman daftar ..."`
- Select: visible label or `aria-label="Jumlah data per halaman"`
- Current page: `aria-current="page"`
- Buttons:
  - `aria-label="Buka halaman sebelumnya"`
  - `aria-label="Buka halaman berikutnya"`
  - `aria-label="Buka halaman pertama"`
  - `aria-label="Buka halaman terakhir"`
- Summary text should use `aria-live="polite"`.

---

## 4. Backend/API Contract Plan

### 4.1 Canonical Contract

For new/migrated collection endpoints, use:

```text
limit=<number>&offset=<number>
```

Add compatibility aliases:

```text
page=<number>&per_page=<number>
```

Resolution rule:

1. If `limit/offset` present, use it.
2. Else if `page/per_page` present, convert:
   - `limit = per_page`
   - `offset = (page - 1) * per_page`
3. If both present, `limit/offset` wins.

### 4.2 Response Meta

Target shape for migrated BFF responses:

```json
{
  "items": [],
  "meta": {
    "total": 123,
    "limit": 24,
    "offset": 48,
    "page": 3,
    "per_page": 24,
    "has_next": true,
    "has_prev": true
  }
}
```

Backend may still return existing `api.PagedResponse` or inline `{ items, total, limit, offset }`. BFF/frontend adapters should normalize without breaking old clients.

### 4.3 Non-breaking Rule

Do **not** abruptly change endpoints that currently return arrays into paginated objects unless all callers are migrated.

Safer approach:

- First: endpoints accept pagination params but keep old full behavior if params absent.
- Then: UI sends params and expects meta.
- Later: enforce default limit only after all callers are compatible.

### 4.4 Max Limits

Recommended:

- Normal table: max `96` in UI, backend max `200`.
- Search/autocomplete: max `50`.
- Realtime/log panel: max `200` or `500` only if needed.
- Export endpoints: separate path, may use `2000+` but not exposed as table page size.

---

## 5. Shared Frontend Component Plan

### Task 1: Create shared pagination component

**Objective:** Add one reusable component for all tables.

**Files:**

- Create: `apps/web-admin/src/lib/components/ui/pagination/table-pagination.svelte`
- Create: `apps/web-admin/src/lib/components/ui/pagination/index.ts`

**Component responsibilities:**

- Render summary range.
- Render page-size selector.
- Render desktop numbered pagination with ellipsis.
- Render mobile compact pagination.
- Emit changes via callback props.
- No data fetch inside component.
- No top-level `window` access.

**Suggested props:**

```ts
page: number;
limit: number;
total: number;
pageSizeOptions?: number[];
defaultLimit?: number;
itemLabel?: string;
loading?: boolean;
disabled?: boolean;
showFirstLast?: boolean;
siblingCount?: number;
ariaLabel?: string;
onchange?: (detail: {
  page: number;
  limit: number;
  offset: number;
  reason: 'page' | 'limit' | 'first' | 'previous' | 'next' | 'last';
}) => void;
```

**Implementation notes:**

- Use Svelte 5 `$props`, `$bindable`, `$derived` if appropriate.
- Clamp page internally for display.
- Do not mutate URL in component by default.
- Parent owns `load(page, limit)` and `syncUrl()`.

### Task 2: Create pagination utility helpers

**Objective:** Avoid copy-pasting parse/clamp/range logic.

**Files:**

- Create: `apps/web-admin/src/lib/utils/pagination.ts`

**Helpers:**

- `normalizePage(value, defaultPage = 1): number`
- `normalizePageSize(value, allowed, defaultLimit): number`
- `calculatePaginationRange(total, page, limit)`
- `buildPageItems(page, pageCount, siblingCount)`
- `offsetForPage(page, limit)`

### Task 3: Add route-level URL helpers if needed

**Objective:** Standardize URL sync for list pages.

**Files:**

- Create or extend: `apps/web-admin/src/lib/utils/url-state.ts`

**Rules:**

- Set `page` only when `page > 1`.
- Set `limit` only when non-default, unless page explicitly requires always storing it.
- Preserve existing filters/search params.
- Use `history.replaceState`, not push spam.
- Guard `typeof window === 'undefined'`.

---

## 6. Rollout Priority

### P0 — Foundation + pilot

1. Add shared `TablePagination` component.
2. Add pagination helpers.
3. Pilot on a simple existing paginated page:
   - `apps/web-admin/src/routes/settings/audit-logs/+page.svelte`
   - Reason: already has `page/per_page`, low business complexity.
4. Pilot on Bank Soal list:
   - `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
   - Reason: user’s original pain point and already has `limit/offset`.

### P1 — High-impact tables

#### Master people/account

- `apps/web-admin/src/routes/students/+page.svelte`
- `apps/web-admin/src/routes/parents/+page.svelte`
- `apps/web-admin/src/routes/employees/+page.svelte` via shared employee components
- `apps/web-admin/src/routes/settings/users/+page.svelte`
- `apps/web-admin/src/routes/settings/user-change-requests/+page.svelte`

#### Bank Soal suite

- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/PublishBankSoalPage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`
- `apps/web-admin/src/routes/bank-soal/pengaturan/+page.svelte` if table grows.

#### Assessment/CBT

- `apps/web-admin/src/routes/asesmen/paket/+page.svelte`
- `apps/web-admin/src/routes/asesmen/paket/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

#### Operational modules

- `apps/web-admin/src/routes/kesiswaan/+page.svelte`
- `apps/web-admin/src/routes/tu/surat-masuk/+page.svelte`
- `apps/web-admin/src/routes/tu/surat-keluar/+page.svelte`
- `apps/web-admin/src/routes/tu/disposisi/+page.svelte`
- `apps/web-admin/src/routes/tu/arsip/+page.svelte`
- `apps/web-admin/src/routes/tu/surat-keterangan/+page.svelte`
- `apps/web-admin/src/routes/library/books/+page.svelte`
- `apps/web-admin/src/routes/library/loans/+page.svelte`
- `apps/web-admin/src/routes/inventory/items/+page.svelte`
- `apps/web-admin/src/routes/governance/actions/+page.svelte`
- `apps/web-admin/src/routes/document-cycles/verifikasi/+page.svelte`
- `apps/web-admin/src/routes/pusaka/antrian/+page.svelte`
- `apps/web-admin/src/routes/pusaka/kehadiran/+page.svelte`

### P2 — Medium priority / client-side pagination first

Use shared component with local array slicing first if server-side backend is not ready.

- Akademik master pages:
  - `akademik/kurikulum`, `mapel`, `rombel`, `guru-mapel`, `beban-guru`, `tahun-ajaran`
- Grades and journal pages:
  - `grades/+page.svelte`
  - `journal/+page.svelte`
- PUSAKA summary and telegram logs.
- Inventory overview tables.

### P3 — Keep static/top-N/print/report initially

Do not force pagination unless data volume proves large:

- Print packs
- Minutes/report pages
- Dashboard top-N cards
- Compliance packs
- Portal previews
- Backup/maintenance static health tables

---

## 7. Agent Swarm Execution Plan

### 7.1 Approval Gate

Before implementation:

1. Present this plan to user.
2. Wait for explicit approval.
3. Check repo status again:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git rev-parse --abbrev-ref HEAD
git rev-parse HEAD
```

4. Do not deploy/restart PM2.

### 7.2 Worktree Setup

Use new worktrees; do not reuse old security sprint worktrees.

Base repo:

```text
/home/servermtsn2kolut/mtsn2kolut-super-app
```

Worktree root:

```text
/home/servermtsn2kolut/mtsn2kolut-swarm-worktrees
```

Branches:

- `swarm/pagination-foundation`
- `swarm/pagination-banksoal-settings`
- `swarm/pagination-master-data`
- `swarm/pagination-asesmen`
- `swarm/pagination-integration`

### 7.3 Batch 1 — Foundation + pilots

Run 5 agents in parallel, but avoid editing the same file.

#### Agent A1 — Foundation shared component

**Allowed paths:**

- `apps/web-admin/src/lib/components/ui/pagination/**`
- `apps/web-admin/src/lib/utils/pagination.ts`
- `apps/web-admin/src/lib/utils/url-state.ts` if needed.

**Tasks:**

- Create `TablePagination`.
- Create pagination helpers.
- Add minimal unit tests if frontend test pattern exists.

#### Agent A2 — Pilot settings audit logs

**Allowed paths:**

- `apps/web-admin/src/routes/settings/audit-logs/+page.svelte`
- BFF route only if needed.

**Tasks:**

- Replace local pagination footer with shared component.
- Preserve `page/per_page` backend contract.
- Verify URL/state behavior.

#### Agent A3 — Pilot Bank Soal list

**Allowed paths:**

- `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
- No backend change unless absolutely required.

**Tasks:**

- Replace hard-coded `PAGE_SIZE=12` with dynamic `pageSize`.
- Add `limit` URL support.
- Use shared pagination footer.
- Clamp page after filter/archive/delete.

#### Agent A4 — QA/a11y review

**Allowed paths:**

- Test files only, or no edits.

**Tasks:**

- Review component API.
- Check accessibility labels.
- Check mobile overflow patterns.
- Prepare regression checklist.

#### Agent A5 — Integration lead

**Allowed paths:**

- Integration branch only.

**Tasks:**

- Cherry-pick A1/A2/A3 after each passes targeted checks.
- Resolve conflicts.
- Run parent validation.

### 7.4 Batch 2 — High-impact domain rollout

After Batch 1 passes and user approves continuing:

#### Agent B1 — Master data/account tables

Paths:

- `students/+page.svelte`
- `parents/+page.svelte`
- shared employee list components
- `settings/users/+page.svelte`

#### Agent B2 — Bank Soal suite

Paths:

- `SoalWorkspacePage.svelte`
- `PublishBankSoalPage.svelte`
- `ReviewBankSoalPage.svelte`

#### Agent B3 — Asesmen/CBT tables

Paths:

- `asesmen/paket/+page.svelte`
- `asesmen/paket/[id]/+page.svelte`
- `asesmen/sesi/+page.svelte`
- `asesmen/sesi/[id]/+page.svelte`
- proctoring pages if safe.

#### Agent B4 — Operational admin tables

Paths:

- TU
- Library
- Inventory
- Document cycles
- Governance actions
- PUSAKA antrian/kehadiran

#### Agent B5 — QA/integration

Tasks:

- Run targeted route smoke tests.
- Check mobile/desktop layouts with browser if login/session available.
- Run `npm run check`/build.

### 7.5 Batch 3 — Backend pagination support

Only after UI pilots prove the component stable.

Agents split by backend domain:

1. People/accounts:
   - students, employees, users, parents.
2. Operational modules:
   - TU, library, inventory, kesiswaan.
3. Assessment:
   - sessions, participants, events, packages.
4. Governance/document cycles.
5. QA/API contract tests.

Rule:

- Backend changes must be additive.
- Existing full-array callers must not break.
- Do not enforce default server limit on legacy endpoints until their UI callers are migrated.

---

## 8. Bite-sized Implementation Tasks

### Task 1: Add pagination utilities

**Objective:** Provide tested helper functions for page/limit normalization and range calculation.

**Files:**

- Create: `apps/web-admin/src/lib/utils/pagination.ts`

**Steps:**

1. Add allowed size normalization helper.
2. Add page normalization helper.
3. Add offset calculation helper.
4. Add page item/ellipsis generator.
5. Add unit tests if frontend test framework has existing utils tests.

**Verify:**

```bash
cd apps/web-admin
npm run check
```

### Task 2: Add shared `TablePagination` component

**Objective:** One reusable table footer for all modules.

**Files:**

- Create: `apps/web-admin/src/lib/components/ui/pagination/table-pagination.svelte`
- Create: `apps/web-admin/src/lib/components/ui/pagination/index.ts`

**Steps:**

1. Build summary text.
2. Build page size selector.
3. Build desktop numbered pagination.
4. Build mobile previous/next layout.
5. Add accessibility labels.
6. Ensure no `window` usage.

**Verify:**

```bash
cd apps/web-admin
npm run check
npm run build
```

### Task 3: Pilot `settings/audit-logs`

**Objective:** Prove the component works on an existing paginated endpoint.

**Files:**

- Modify: `apps/web-admin/src/routes/settings/audit-logs/+page.svelte`

**Steps:**

1. Identify current page/per_page state.
2. Replace one-off controls with `TablePagination`.
3. Preserve existing API query shape.
4. Add URL state if currently missing.
5. Test first/prev/number/next/last/page size.

**Verify:**

```bash
cd apps/web-admin
npm run check
npm run build
```

### Task 4: Pilot `BankSoalListPage`

**Objective:** Implement the original recommended option for Bank Soal using the shared component.

**Files:**

- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Steps:**

1. Replace `PAGE_SIZE` with `DEFAULT_PAGE_SIZE` and `pageSize` state.
2. Update `pageCount`, `resultStart`, `resultEnd`.
3. Update `buildListParams(page, limit)` to send `limit` and `offset`.
4. Update `fetchOverview` and current overview fallback.
5. Read `limit` from URL.
6. Sync `limit` to URL.
7. Replace footer with `TablePagination`.
8. Clamp page after fetch and after archive/delete.
9. Verify no literal `\t` or `\n` escaped artifacts.

**Verify:**

```bash
cd apps/web-admin
npm run check
npm run build
cd ../..
cd services/core-api
go test ./internal/handler/ ./internal/service/
```

### Task 5: Parent integration review

**Objective:** Ensure pilots are safe before scaling.

**Files:**

- No direct source edits unless resolving conflicts.

**Steps:**

1. Inspect `git diff --stat`.
2. Inspect actual diff for shared component and pilots.
3. Run frontend check/build.
4. Run relevant backend tests if Bank Soal endpoint touched.
5. Browser-smoke desktop and mobile widths.
6. Commit local patch.

---

## 9. Validation Commands

From repo root:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git diff --check
```

Frontend:

```bash
cd apps/web-admin
npm run check
rm -rf build && npm run build
```

Backend if Go/BFF/API contract changed:

```bash
cd services/core-api
go test ./internal/handler/ ./internal/service/
go build ./cmd/api
```

Root scripts if appropriate:

```bash
npm run check:web
npm run test:web
npm run test:backend
npm run security:scan-secrets
npm run ops:validate-pm2
```

Post-build protected route smoke if deployed/restarted later:

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8021/bank-soal/daftar
# Expected: 302 when unauthenticated
```

---

## 10. Manual QA Checklist

### Desktop

- Range text correct: `Menampilkan x–y dari total ...`
- Numbered pagination appears for multi-page data.
- Current page visually active and has `aria-current="page"`.
- `Awal/Sebelumnya` disabled on first page.
- `Berikutnya/Akhir` disabled on last page.
- Page size changes reload data and reset page to 1.
- Search/filter changes reset page to 1 but preserve page size.
- Clear filters preserves page size.
- No table overflow beyond viewport.

### Mobile

- Footer stacks vertically.
- Numbered buttons hidden or compact.
- Previous/next buttons are touch-friendly.
- Select is readable and not clipped.
- Table/card list does not overflow horizontally except intentional table wrapper scroll.

### URL

- `?page=2&limit=24` opens correct state.
- Invalid `limit` falls back safely.
- Invalid `page` falls back safely.
- Large `page=999` clamps to valid page.

### Data mutation edge cases

- Archive/delete last item on last page does not leave empty fake page.
- Loading state disables repeated navigation.
- Old/stale request does not overwrite newer limit/page state.

---

## 11. Risks and Mitigations

### Risk: Breaking existing array responses

Mitigation:

- Use additive backend changes.
- UI migration opt-in via query params.
- Avoid changing response shape until callers are migrated.

### Risk: Too many files changed at once

Mitigation:

- Batch rollout.
- Pilot first.
- Domain-specific agents in separate worktrees.

### Risk: Backend count queries become expensive

Mitigation:

- Add count only on high-impact endpoints first.
- Ensure filters/indexes are reviewed.
- Keep dashboard/report/top-N pages out of P1.

### Risk: Mobile overflow regressions

Mitigation:

- Shared component mobile stacked by default.
- Use `overflow-x-auto` on table wrappers.
- Browser-smoke at 360px/390px widths.

### Risk: PM2 stale build chunks

Mitigation:

- Always `rm -rf build && npm run build` before restart.
- Restart web-admin only after approval.

### Risk: Agent branch conflicts

Mitigation:

- Worktree per agent.
- Strict allowed file scopes.
- Integration lead cherry-picks sequentially.

---

## 12. Rollback Plan

### Before deploy

- Revert/cherry-pick only within swarm branches.
- Do not reset main branch without explicit approval.

### Frontend-only rollback after deploy

1. Checkout previous-good commit.
2. Clean build:

```bash
cd apps/web-admin
rm -rf build && npm run build
```

3. Restart PM2 web-admin only after approval:

```bash
pm2 restart mtsn2kolut-web-admin
```

4. Smoke test protected route returns expected 302 unauthenticated.

### Backend rollback after deploy

1. Revert backend commit or checkout previous-good.
2. Restart `mtsn2kolut-core-api` after approval.
3. Health check/log check.
4. If frontend depended on new contract, rollback frontend too.

### Migration rollback

- Avoid destructive migrations for this pagination rollout.
- If migrations are ever added, require backup and separate approval.

---

## 13. Definition of Done

For Batch 1 foundation/pilot:

- Shared pagination component exists and builds.
- Settings audit logs uses shared pagination.
- Bank Soal list uses dynamic page size `12/24/48/96`.
- URL `page` and `limit` work.
- Mobile/desktop layouts are responsive.
- `npm run check` passes.
- clean frontend build passes.
- relevant Go tests pass if backend touched.
- Diff reviewed; no secret/build artifacts.
- Local commit created.
- No deploy/restart unless user approves.

For global rollout:

- All P1 tables migrated or explicitly marked not applicable.
- P2 tables have either client-side pagination or backend support plan.
- P3 print/report/static pages are documented as intentionally excluded.
- Consistent UX labels across modules.
- No known critical/important a11y or overflow issue remains.
