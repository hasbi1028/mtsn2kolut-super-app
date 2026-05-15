# Bank Soal Composer Responsive UX Plan

> **For Hermes:** Planning only. Do not implement until user selects an option. If executed later, use Codex CLI relay or subagent-driven-development, then build/test and commit.

**Goal:** Merapikan tampilan halaman `/bank-soal/tambah` / komposer Bank Soal di PC tanpa merusak pengalaman HP/tablet, autosave offline, dan workflow review.

**Current context:**
- Entry page: `apps/web-admin/src/routes/bank-soal/tambah/+page.svelte`
- Main component: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Composer currently rendered as one large inline section at lines around `3761-4454`.
- Current desktop layout uses `xl:grid-cols-[minmax(0,1.55fr)_minmax(18rem,0.5fr)]` with preview sidebar sticky at `xl` only.
- Header, readiness rail, metadata, editor blocks, options, preview, and sticky footer all live in one component. This makes desktop feel cramped/uneven and hard to tune.

---

## Recommended Options

### Option A — Low-risk polish: keep current structure, adjust responsive classes

**What changes:**
- Keep `SoalWorkspacePage.svelte` as-is.
- Tune width, spacing, grid columns, sticky preview, and bottom action bar.
- Make PC layout more balanced with a wider editor and a cleaner right preview.

**Expected desktop result:**
- Editor area no longer feels squeezed.
- Preview has consistent width and better sticky top offset.
- Metadata row wraps more naturally.
- Bottom buttons align better.

**Pros:**
- Fastest.
- Lowest risk.
- Minimal code movement.
- Good if target is just “kurang rapi di PC”.

**Cons:**
- Component remains very large.
- Future UX improvements still harder.
- Does not fundamentally improve workflow clarity.

**Estimated implementation:** 1 focused patch.

---

### Option B — Recommended: desktop 3-zone composer shell

**What changes:**
- Preserve existing data/state logic in `SoalWorkspacePage.svelte`.
- Refactor composer UI into clearer zones:
  1. **Top sticky composer bar**: title, mode, scope, readiness, draft status.
  2. **Main authoring canvas**: metadata + question + options/rubric with comfortable max width.
  3. **Right inspector**: live preview, validation issues, quick jump/stage rail.
- Desktop breakpoint:
  - `lg`: single main column + collapsible preview behavior.
  - `xl/2xl`: two-column layout, e.g. `minmax(0, 1fr) 22rem/24rem`.
- Move noisy status chips out of the body into a compact top bar or right inspector.

**Expected desktop result:**
- PC view feels like a real “studio”: editor left, review/preview right.
- Less vertical clutter before users reach the question editor.
- More predictable alignment for metadata and options.
- HP remains tabbed Editor/Preview as it already is.

**Pros:**
- Best balance between improvement and safety.
- No backend changes.
- Can be implemented incrementally.
- Makes future UI easier.

**Cons:**
- Slightly more refactor than Option A.
- Needs careful visual QA across question types.

**Estimated implementation:** 2-3 focused patches.

---

### Option C — Full workflow redesign: wizard/stepper composer

**What changes:**
- Convert composer into steps:
  1. Konteks & metadata
  2. Isi soal / stimulus
  3. Opsi & kunci / rubrik
  4. Preview & ajukan review
- Right side shows summary + readiness.
- Desktop can still show multi-pane view, but primary interaction becomes step-by-step.

**Expected desktop result:**
- Very clean for teachers who create one question at a time.
- Reduces cognitive load significantly.

**Pros:**
- Most user-friendly for non-technical guru.
- Clear validation per step.
- Good long-term UX.

**Cons:**
- Highest risk.
- More state/navigation logic.
- Needs more testing so autosave, focus editor, and multiple question types do not regress.
- Slower to deliver.

**Estimated implementation:** 1 larger feature branch, multiple commits.

---

## Recommendation

Choose **Option B**.

Reason: current code already has the required features; the main problem is layout density on PC. Option B cleans the desktop experience without changing save APIs, permissions, offline queue, or database. It also avoids the risk of a full wizard rewrite.

---

## Option B Implementation Plan

### Task 1: Baseline inspection and screenshot QA

**Objective:** Capture current desktop/mobile issues before editing.

**Files:**
- Inspect: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Optional browser QA: `/bank-soal/tambah`

**Steps:**
1. Run app or use existing PM2/dev environment if available.
2. Login with test/admin/guru account.
3. Capture screenshots at:
   - Desktop 1440px or wider.
   - Laptop 1280px.
   - Tablet width.
   - Mobile width.
4. Note visible issues:
   - header too tall,
   - metadata cramped,
   - preview too narrow/wide,
   - action bar alignment,
   - stage/readiness rail crowding,
   - option cards wrapping awkwardly.

**Verification:** Baseline screenshots saved for comparison.

---

### Task 2: Introduce composer layout constants/classes

**Objective:** Make responsive layout easier to tune without scattered class edits.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`

**Steps:**
1. Identify these regions:
   - composer shell: line around `3762`
   - header: `3763-3811`
   - scroll body: `3813`
   - main grid: `3867`
   - right aside: `4307`
   - footer actions: `4315`
2. Replace overly specific desktop grid class:
   - from: `xl:grid-cols-[minmax(0,1.55fr)_minmax(18rem,0.5fr)]`
   - to a clearer editor/inspector ratio, for example:
     - `xl:grid-cols-[minmax(0,1fr)_22rem] 2xl:grid-cols-[minmax(0,1fr)_24rem]`
3. Add max-width to main authoring canvas if needed:
   - editor column stays fluid,
   - cards avoid stretching too wide.

**Verification:** `npm run check` or app build should still pass.

---

### Task 3: Compact desktop header and status chips

**Objective:** Reduce top visual clutter on PC.

**Files:**
- Modify: `SoalWorkspacePage.svelte` header/status region around `3763-3865`

**Steps:**
1. Keep title and main action buttons visible.
2. Move some chips/status indicators into a smaller single-line strip on desktop.
3. On mobile, keep the existing stacked layout.
4. Ensure text does not overflow when event title is long.

**Suggested behavior:**
- Desktop: compact header + status strip.
- Mobile: stacked title/buttons as current.

**Verification:** Header height reduced; no horizontal overflow.

---

### Task 4: Make metadata section responsive and readable

**Objective:** Fix PC metadata row that can look cramped because it uses fixed column widths.

**Files:**
- Modify: `SoalWorkspacePage.svelte` metadata region around `3869-3955`

**Steps:**
1. Split metadata into two clear rows:
   - row 1: question type + authoring mode,
   - row 2: subject, difficulty, weight, RTL toggle.
2. Replace the fixed grid `lg:grid-cols-[8rem_minmax(0,1fr)_8rem_6.5rem_10.5rem]` with a more natural responsive grid:
   - mobile: 1 column,
   - tablet: 2 columns,
   - desktop: subject wider, other controls smaller.
3. Make `Mata Pelajaran` select take the largest width.

**Verification:** Mapel dropdown readable on PC; buttons do not wrap awkwardly.

---

### Task 5: Improve editor cards and option cards

**Objective:** Make question body and answer options easier to scan on PC.

**Files:**
- Modify: `SoalWorkspacePage.svelte` editor regions around `4028-4239`

**Steps:**
1. Ensure question editor has comfortable vertical size.
2. For option cards:
   - keep 2-column on `2xl`,
   - use 1-column or balanced 2-column on `xl` depending width.
3. Make key-answer controls visually consistent and aligned.
4. Keep rich text editor image upload behavior unchanged.

**Verification:** PG, jawaban ganda, benar/salah, menjodohkan, isian, essay all render cleanly.

---

### Task 6: Improve right inspector / preview panel

**Objective:** Make preview useful on PC instead of feeling like squeezed side content.

**Files:**
- Modify: `SoalWorkspacePage.svelte` aside around `4307-4311`
- Possibly inspect `composerPreview()` snippet if layout needs internal cleanup.

**Steps:**
1. Set right panel width to around 22-24rem on large screens.
2. Use sticky offset that accounts for page top/header if needed:
   - `xl:sticky xl:top-4` or equivalent.
3. Put validation/readiness and preview in the side panel if it reduces clutter.
4. Ensure mobile still uses Editor/Preview toggle.

**Verification:** Preview visible and stable while scrolling on desktop; hidden/toggled correctly on HP.

---

### Task 7: Clean sticky footer action bar

**Objective:** Make save/review actions clean on PC and touch-friendly on HP.

**Files:**
- Modify: `SoalWorkspacePage.svelte` footer region around `4315-4353`

**Steps:**
1. On desktop: status left, actions right, no crowded wrapping.
2. On mobile: actions can stack full width or wrap cleanly.
3. Make primary action `Ajukan Review` visually dominant.
4. Keep Ctrl+S and draft status text.

**Verification:** Buttons remain accessible and no overflow at mobile width.

---

### Task 8: Validate no functional regressions

**Objective:** Confirm UX refactor did not break composer behavior.

**Commands:**
```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run check
npm run test -- --run
npm run build
```

**Manual QA:**
- Open `/bank-soal/tambah`.
- Select Mapel.
- Test these forms:
  - Pilihan Ganda,
  - Jawaban Ganda,
  - Benar/Salah,
  - Menjodohkan,
  - Isian Singkat,
  - Essay.
- Confirm:
  - draft autosave still shows status,
  - Fokus editor opens/closes,
  - Preview updates,
  - Simpan Draft works,
  - Ajukan Review stays disabled/enabled correctly,
  - offline notice still visible.

---

### Task 9: Commit only after validation

**Objective:** Preserve safe deployment boundary.

**Suggested commit:**
```bash
git add apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte

git commit -m "style(web-admin): improve bank soal composer desktop layout"
```

**Do not deploy/restart PM2 unless user explicitly asks.**

---

## Risks and Guardrails

- **Risk:** Refactor breaks Svelte 5 snippet syntax.
  - Guardrail: keep logic untouched; change markup/classes only first.
- **Risk:** Mobile layout regresses.
  - Guardrail: maintain existing `composerMobilePanel` behavior.
- **Risk:** Rich text editor image upload or binding breaks.
  - Guardrail: do not extract editor components unless necessary.
- **Risk:** Very large component makes changes hard to review.
  - Guardrail: commit in small patches; optionally extract subcomponents only after visual polish passes.

---

## Open decision for user

Pick one:
- **A:** Cepat dan aman — polish class/layout saja.
- **B:** Rekomendasi — desktop studio 3-zone, rapi tapi tetap aman.
- **C:** Redesign penuh — wizard step-by-step, paling bagus jangka panjang tapi lebih lama.
