# Hybrid Accordion for Bank Soal & Paket Soal Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menambahkan tampilan detail expandable/accordion pada Daftar Bank Soal dan Daftar Paket Soal agar soal LaTeX/tabel, opsi, kunci, pembahasan, metadata, dan komposisi paket bisa dicek cepat tanpa masuk halaman edit.

**Architecture:** Implementasi bersifat frontend-only pada SvelteKit web-admin. Desktop tetap memakai tabel ringkas untuk scanning cepat, lalu setiap baris punya expandable detail row. Mobile memakai card accordion. Detail menggunakan `RichContent` agar LaTeX/tabel HTML dirender aman dan konsisten.

**Tech Stack:** SvelteKit/Svelte 5 runes, Tailwind v4, shadcn-svelte Button/Badge/Card/Table, existing `RichContent.svelte`, existing `TablePagination`, PM2 adapter-node deployment.

---

## Prinsip UX

1. **Jangan ganti tabel desktop sepenuhnya.** Desktop tetap tabel supaya admin cepat membandingkan banyak soal/paket.
2. **Tambahkan detail on-demand.** Detail hanya muncul saat operator menekan `Detail`, supaya halaman tidak terlalu padat.
3. **Mobile pakai card accordion.** Di HP, tabel sempit kurang nyaman; card ringkas + detail lebih cocok.
4. **Soal LaTeX/tabel wajib render.** Jangan tampil mentah `$$...$$` atau `<table>...` di area detail.
5. **Aksi utama tetap mudah.** `Edit/Lihat` tetap terlihat; aksi sekunder tetap di `Lainnya` sesuai pola saat ini.
6. **Tidak ubah backend dan DB.** Data yang dibutuhkan sudah ada di list payload; jika detail tidak lengkap, fallback ke ringkasan yang tersedia dulu.

---

## Acceptance Criteria

### Bank Soal

- Desktop `/bank-soal/daftar` tetap menampilkan tabel utama.
- Setiap baris punya tombol `Detail` atau ikon expand/collapse.
- Saat dibuka, detail row menampilkan:
  - kode soal,
  - soal penuh dengan `RichContent`,
  - opsi A/B/C/D/E jika ada,
  - kunci jawaban jika role boleh melihat,
  - pembahasan jika ada,
  - metadata pembuat/mapel/kelas/materi/C-level/HOTS/status/pemakaian,
  - catatan penulis/reviewer jika ada,
  - aksi yang sudah ada tetap tersedia tanpa mengubah permission guard.
- Mobile card memiliki area ringkas + tombol `Detail` yang membuka isi yang sama.
- Hanya satu detail row terbuka per halaman secara default agar daftar tidak berat. Membuka baris lain menutup baris sebelumnya.
- Render soal tabel seperti `MTK-VIII-PG-0032` terlihat sebagai tabel, bukan HTML mentah.
- Render LaTeX seperti `MTK-VIII-PG-0021` terlihat sebagai rumus, bukan teks mentah.

### Paket Soal

- Desktop `/asesmen/paket` tetap menampilkan daftar/tabel paket.
- Setiap paket punya `Detail` yang membuka ringkasan paket:
  - nama/kode paket,
  - mapel/kelas,
  - jumlah soal,
  - status,
  - jadwal/kegiatan jika tersedia,
  - komposisi ringkas jika datanya tersedia di payload,
  - aksi `Kelola/Edit/Lihat` tetap terlihat.
- Mobile daftar paket memakai card accordion.
- Jika daftar isi soal paket belum tersedia di payload halaman list, jangan fetch tambahan dulu pada Sprint 1; tampilkan metadata paket yang ada. Fetch detail paket bisa menjadi Sprint 2.

---

## Files & Known Paths

### Bank Soal

- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
- Uses existing: `apps/web-admin/src/lib/components/RichContent.svelte`
- Uses existing: `apps/web-admin/src/lib/components/ui/pagination/table-pagination.svelte`
- Route wrapper: `apps/web-admin/src/routes/bank-soal/daftar/+page.svelte`

### Paket Soal

- Modify: `apps/web-admin/src/routes/asesmen/paket/+page.svelte`
- Detail page existing: `apps/web-admin/src/routes/asesmen/paket/[id]/+page.svelte`
- New optional reusable component:
  - Create: `apps/web-admin/src/lib/components/assessment/ExpandableDetailPanel.svelte` OR keep local helpers first.

### Shared Optional Components

- Prefer local implementation first for lower risk.
- If duplicated markup becomes large, extract later:
  - Create: `apps/web-admin/src/lib/components/bank-soal/QuestionDetailPanel.svelte`
  - Create: `apps/web-admin/src/lib/components/assessment/PackageDetailPanel.svelte`

---

## Sprint 0 — Read-Only Audit & Safety Gate

### Task 0.1: Inspect current Bank Soal list data shape

**Objective:** Confirm fields available in `Question` type before adding detail UI.

**Files:**
- Read: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Steps:**
1. Search for `type Question` or `interface Question`.
2. Record available fields: `question_text`, `stem_html`, `stem_latex`, `options`, `answer_key`, `explanation`, `explanation_html`, `writer_notes`, `review_notes`, `material_topic`, `cognitive_level`, `hots_flag`, `package_count`, `answer_count`, `author_username`, display author fields.
3. Confirm current `RichContent` import exists from the previous LaTeX list-preview patch.

**Verification:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
python3 - <<'PY'
from pathlib import Path
p=Path('apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte')
text=p.read_text()
for key in ['RichContent','answer_key','explanation','writer_notes','review_notes','options']:
    print(key, key in text)
PY
```

Expected: `RichContent` true; note which detail fields already exist.

### Task 0.2: Inspect current Paket list data shape

**Objective:** Confirm what `/asesmen/paket` already has before planning detail panel contents.

**Files:**
- Read: `apps/web-admin/src/routes/asesmen/paket/+page.svelte`

**Steps:**
1. Locate package type/interface.
2. Record available fields for status, subject, grade, count, schedule, activity, question count.
3. Identify current table/card layout and action buttons.

**Verification:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
python3 - <<'PY'
from pathlib import Path
p=Path('apps/web-admin/src/routes/asesmen/paket/+page.svelte')
text=p.read_text()
print('chars', len(text))
for key in ['Table', 'Card', 'TablePagination', 'question', 'status', 'package']:
    print(key, key in text)
PY
```

Expected: file readable and current layout identified.

---

## Sprint 1 — Bank Soal Expandable Detail

### Task 1.1: Add expanded row state

**Objective:** Track which Bank Soal row is open.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Implementation:**

Add near other `$state` declarations:

```ts
let expandedQuestionId = $state<string | null>(null);

function toggleQuestionDetail(questionId: string) {
    expandedQuestionId = expandedQuestionId === questionId ? null : questionId;
}

function isQuestionExpanded(questionId: string): boolean {
    return expandedQuestionId === questionId;
}
```

When `load()` refreshes list data, reset stale expanded ID only if the ID no longer exists:

```ts
if (expandedQuestionId && !questions.some((question) => question.id === expandedQuestionId)) {
    expandedQuestionId = null;
}
```

**Verification:**

```bash
cd apps/web-admin
npm run check
```

Expected: no Svelte/TS errors.

### Task 1.2: Create local helper functions for rendered question/detail fields

**Objective:** Centralize fallback logic so table and mobile card use identical content.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Implementation shape:**

```ts
function questionStem(question: Question): string {
    return compactText(question.stem_html || question.question_text || question.stem_latex, 'Soal belum memiliki teks.');
}

function questionExplanation(question: Question): string {
    return compactText(question.explanation_html || question.explanation, '');
}

function questionOptions(question: Question): Array<{ label: string; html: string; text: string }> {
    const rawOptions = Array.isArray(question.options) ? question.options : [];
    if (rawOptions.length > 0) {
        return rawOptions
            .map((option) => ({
                label: compactText(option.label, ''),
                html: compactText(option.html || option.text, ''),
                text: compactText(option.text || option.html, '')
            }))
            .filter((option) => option.label && (option.html || option.text));
    }
    return [
        ['A', question.option_a],
        ['B', question.option_b],
        ['C', question.option_c],
        ['D', question.option_d],
        ['E', question.option_e]
    ]
        .map(([label, text]) => ({ label, html: compactText(text, ''), text: compactText(text, '') }))
        .filter((option) => option.html || option.text);
}
```

Adjust exact field names after Task 0.1 audit. Do not invent fields that do not exist.

**Verification:**

```bash
cd apps/web-admin
npm run check
```

Expected: no type errors.

### Task 1.3: Add `QuestionDetailPanel` markup inside `BankSoalListPage.svelte`

**Objective:** Render full question, options, key, explanation, metadata, and notes in one reusable block.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Implementation shape:**

Use a Svelte snippet or local repeated block. Preferred if Svelte syntax in file already uses snippets:

```svelte
{#snippet questionDetailPanel(question)}
    <div class="rounded-2xl border bg-muted/20 p-4 text-sm space-y-4">
        <div class="flex flex-wrap items-center gap-2">
            <Badge variant="outline">{compactText(question.code, 'Tanpa kode')}</Badge>
            <Badge variant="secondary">{compactText(question.workflow_status, 'draft')}</Badge>
            <Badge variant="outline">{compactText(question.target_level, 'Tanpa tingkat')}</Badge>
        </div>

        <section class="space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Soal</p>
            <div class="rounded-xl border bg-background p-3">
                <RichContent content={questionStem(question)} />
            </div>
        </section>

        {#if questionOptions(question).length}
            <section class="space-y-2">
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Pilihan Jawaban</p>
                <div class="grid gap-2 md:grid-cols-2">
                    {#each questionOptions(question) as option (option.label)}
                        <div class="rounded-lg border bg-background p-3">
                            <div class="flex gap-2">
                                <span class="font-semibold">{option.label}.</span>
                                <div class="min-w-0 flex-1"><RichContent content={option.html || option.text} /></div>
                            </div>
                        </div>
                    {/each}
                </div>
            </section>
        {/if}

        {#if question.answer_key}
            <section class="rounded-xl border bg-background p-3">
                <p class="text-xs text-muted-foreground">Kunci Jawaban</p>
                <p class="font-semibold">{question.answer_key}</p>
            </section>
        {/if}

        {#if questionExplanation(question)}
            <section class="space-y-2">
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Pembahasan</p>
                <div class="rounded-xl border bg-background p-3">
                    <RichContent content={questionExplanation(question)} />
                </div>
            </section>
        {/if}

        <section class="grid gap-2 text-xs text-muted-foreground sm:grid-cols-2 lg:grid-cols-4">
            <p><span class="font-medium text-foreground">Mapel:</span> {compactText(question.subject_name, '-')}</p>
            <p><span class="font-medium text-foreground">Pembuat:</span> {compactText(question.author_display_name || question.author_username, '-')}</p>
            <p><span class="font-medium text-foreground">Materi:</span> {compactText(question.material_topic, '-')}</p>
            <p><span class="font-medium text-foreground">Level:</span> {compactText(question.cognitive_level, '-')}</p>
        </section>

        {#if question.writer_notes || question.review_notes}
            <section class="rounded-xl border border-amber-200 bg-amber-50 p-3 text-xs text-amber-900 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-100">
                {#if question.writer_notes}<p><b>Catatan penulis:</b> {question.writer_notes}</p>{/if}
                {#if question.review_notes}<p><b>Catatan reviewer:</b> {question.review_notes}</p>{/if}
            </section>
        {/if}
    </div>
{/snippet}
```

Adjust field names after audit. If `Badge`/`RichContent` already imported, reuse existing imports.

**Verification:**

```bash
cd apps/web-admin
npm run check
```

Expected: no Svelte syntax error.

### Task 1.4: Add desktop table expand button and detail row

**Objective:** Add a `Detail` button and a new row below expanded question.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Implementation pattern:**

Inside desktop table row actions, add:

```svelte
<Button
    variant="outline"
    size="sm"
    type="button"
    aria-expanded={isQuestionExpanded(question.id)}
    aria-controls={`question-detail-${question.id}`}
    on:click={() => toggleQuestionDetail(question.id)}
>
    {isQuestionExpanded(question.id) ? 'Tutup' : 'Detail'}
</Button>
```

Immediately after the main `<Table.Row>` for each question:

```svelte
{#if isQuestionExpanded(question.id)}
    <Table.Row id={`question-detail-${question.id}`} class="bg-muted/10">
        <Table.Cell colspan={/* number of visible desktop columns */} class="p-4">
            {@render questionDetailPanel(question)}
        </Table.Cell>
    </Table.Row>
{/if}
```

Use exact `colspan` matching current table headers. Count columns manually from the file.

**Verification:**

```bash
cd apps/web-admin
npm run check
npm run build
```

Expected: build passes.

### Task 1.5: Add mobile card accordion detail

**Objective:** Mobile card gets the same detail panel without entering edit page.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Implementation pattern:**

In the mobile card action row:

```svelte
<Button
    variant="outline"
    size="sm"
    type="button"
    aria-expanded={isQuestionExpanded(question.id)}
    aria-controls={`question-mobile-detail-${question.id}`}
    on:click={() => toggleQuestionDetail(question.id)}
>
    {isQuestionExpanded(question.id) ? 'Tutup detail' : 'Detail'}
</Button>
```

Below summary area:

```svelte
{#if isQuestionExpanded(question.id)}
    <div id={`question-mobile-detail-${question.id}`} class="pt-3">
        {@render questionDetailPanel(question)}
    </div>
{/if}
```

**Verification:**

```bash
cd apps/web-admin
npm run check
npm run build
```

Expected: build passes.

### Task 1.6: Browser smoke Bank Soal detail

**Objective:** Confirm LaTeX and table question render correctly after PM2 restart.

**Files:**
- No source changes.

**Steps:**
1. Build clean.
2. Restart frontend if user approved deployment.
3. Open `/bank-soal/daftar?search=MTK-VIII-PG-0032`.
4. Login or use existing authenticated browser session if available.
5. Click `Detail`.
6. Verify table appears as rendered table.
7. Repeat search `MTK-VIII-PG-0021`; verify LaTeX renders.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
rm -rf build && npm run build
pm2 restart mtsn2kolut-web-admin --update-env
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8021/bank-soal/daftar
```

Expected: protected route returns `302` to login.

---

## Sprint 2 — Paket Soal Expandable Detail

### Task 2.1: Add expanded package state

**Objective:** Track which package row/card is expanded.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/paket/+page.svelte`

**Implementation:**

```ts
let expandedPackageId = $state<string | null>(null);

function togglePackageDetail(packageId: string) {
    expandedPackageId = expandedPackageId === packageId ? null : packageId;
}

function isPackageExpanded(packageId: string): boolean {
    return expandedPackageId === packageId;
}
```

Use actual package ID field name from Task 0.2.

**Verification:**

```bash
cd apps/web-admin
npm run check
```

Expected: no errors.

### Task 2.2: Add package detail panel markup

**Objective:** Show package metadata and available composition without extra backend fetch.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/paket/+page.svelte`

**Implementation shape:**

```svelte
{#snippet packageDetailPanel(pkg)}
    <div class="rounded-2xl border bg-muted/20 p-4 text-sm space-y-4">
        <div class="flex flex-wrap items-center gap-2">
            <Badge variant="outline">{pkg.code || pkg.name || 'Paket'}</Badge>
            <Badge variant="secondary">{pkg.status || 'draft'}</Badge>
        </div>

        <section class="grid gap-2 text-xs text-muted-foreground sm:grid-cols-2 lg:grid-cols-4">
            <p><span class="font-medium text-foreground">Mapel:</span> {pkg.subject_name || '-'}</p>
            <p><span class="font-medium text-foreground">Kelas:</span> {pkg.target_level || pkg.grade || '-'}</p>
            <p><span class="font-medium text-foreground">Jumlah soal:</span> {pkg.question_count ?? '-'}</p>
            <p><span class="font-medium text-foreground">Dibuat:</span> {formatDate(pkg.created_at)}</p>
        </section>

        <section class="rounded-xl border bg-background p-3">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Komposisi</p>
            <p class="mt-1 text-sm text-muted-foreground">
                Komposisi detail mengikuti data yang tersedia di daftar. Detail isi soal lengkap dibuka melalui tombol Kelola/Lihat.
            </p>
        </section>
    </div>
{/snippet}
```

Use exact field names and existing date formatter from the file.

**Verification:**

```bash
cd apps/web-admin
npm run check
```

Expected: no type errors.

### Task 2.3: Add desktop table detail row for Paket Soal

**Objective:** Add `Detail` button and expanded row under selected package.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/paket/+page.svelte`

**Implementation:**

Same pattern as Bank Soal:

```svelte
<Button
    variant="outline"
    size="sm"
    type="button"
    aria-expanded={isPackageExpanded(pkg.id)}
    aria-controls={`package-detail-${pkg.id}`}
    on:click={() => togglePackageDetail(pkg.id)}
>
    {isPackageExpanded(pkg.id) ? 'Tutup' : 'Detail'}
</Button>
```

Then row:

```svelte
{#if isPackageExpanded(pkg.id)}
    <Table.Row id={`package-detail-${pkg.id}`} class="bg-muted/10">
        <Table.Cell colspan={/* package table column count */} class="p-4">
            {@render packageDetailPanel(pkg)}
        </Table.Cell>
    </Table.Row>
{/if}
```

**Verification:**

```bash
cd apps/web-admin
npm run check
npm run build
```

Expected: build passes.

### Task 2.4: Add mobile card detail for Paket Soal

**Objective:** Package list on mobile uses the same accordion behavior.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/paket/+page.svelte`

**Implementation:**

Add `Detail` button in card action row and render panel below card summary:

```svelte
{#if isPackageExpanded(pkg.id)}
    <div id={`package-mobile-detail-${pkg.id}`} class="pt-3">
        {@render packageDetailPanel(pkg)}
    </div>
{/if}
```

**Verification:**

```bash
cd apps/web-admin
npm run check
npm run build
```

Expected: build passes.

---

## Sprint 3 — Polish, Accessibility, and Performance

### Task 3.1: Add visual states and accessibility labels

**Objective:** Ensure buttons are understandable and keyboard accessible.

**Files:**
- Modify: Bank Soal and Paket Soal files.

**Requirements:**
- Buttons include `aria-expanded`.
- Buttons include `aria-controls`.
- Detail panel has matching `id`.
- Use clear text: `Detail`, `Tutup`, `Tutup detail`.
- Do not rely on icon-only button unless it has `aria-label`.

**Verification:**

```bash
cd apps/web-admin
npm run check
npm run build
```

### Task 3.2: Prevent layout overflow

**Objective:** Detail panels must not cause horizontal overflow on mobile or desktop.

**Files:**
- Modify: Bank Soal and Paket Soal files.

**Rules:**
- Wrap rich table content with `overflow-x-auto` if needed.
- Ensure detail panel has `min-w-0` where inside flex/grid.
- Keep desktop table wrapper `overflow-x-auto` only at table container.

**Verification browser console:**

```js
document.querySelectorAll('*').forEach(el => {
    const r = el.getBoundingClientRect();
    if (r.width > document.documentElement.clientWidth) {
        console.log('overflow', el.tagName, r.width, el.className);
    }
});
```

Expected: no unexpected wide elements.

### Task 3.3: Add small empty/fallback states

**Objective:** Detail panels do not look broken when data is missing.

**Rules:**
- If options absent: show `Pilihan jawaban belum tersedia.`
- If explanation absent: hide section or show muted `Pembahasan belum tersedia.`
- If package composition absent: show explanatory copy, not blank card.

**Verification:**

Search for a package/question with missing fields and expand it.

---

## Sprint 4 — Final Verification and Delivery

### Task 4.1: Run frontend validation

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
npm run check
rm -rf build && npm run build
```

Expected: both pass.

### Task 4.2: Source hygiene checks

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git diff --check
python3 - <<'PY'
from pathlib import Path
for path in [
 'apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte',
 'apps/web-admin/src/routes/asesmen/paket/+page.svelte',
]:
    text=Path(path).read_text()
    print(path, 'literal \\t:', '\\t' in text, 'literal \\n:', '\\n' in text)
PY
```

Expected: `git diff --check` clean; no accidental literal tab/newline escape artifacts.

### Task 4.3: Deploy/restart after explicit approval

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
rm -rf build && npm run build
pm2 restart mtsn2kolut-web-admin --update-env
sleep 2
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8021/bank-soal/daftar
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8021/asesmen/paket
```

Expected protected routes: `302` to login.

### Task 4.4: Browser smoke checklist

Manual/browser verification:

- `/bank-soal/daftar?search=MTK-VIII-PG-0032`
  - Click `Detail`.
  - Verify table rendered.
  - Verify key `B` and notes visible only if intended by role.
- `/bank-soal/daftar?search=MTK-VIII-PG-0021`
  - Click `Detail`.
  - Verify LaTeX fraction rendered.
- `/asesmen/paket`
  - Click `Detail` on one package.
  - Verify package metadata panel visible.
- Mobile viewport smoke:
  - Bank Soal cards open/close detail.
  - Paket cards open/close detail.
  - No horizontal overflow.

### Task 4.5: Commit focused changes

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git add apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte \
        apps/web-admin/src/routes/asesmen/paket/+page.svelte
git commit -m "feat(ui): add expandable details to question and package lists"
```

If only Bank Soal is implemented in the first pass:

```bash
git add apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte
git commit -m "feat(bank-soal): add expandable question details"
```

---

## Rollout Recommendation

### Recommended sequence

1. **Implement Bank Soal only first.** This is highest value because LaTeX/table questions already exist.
2. Validate on mobile and desktop.
3. Deploy Bank Soal improvement.
4. After stable, implement Paket Soal detail panel.

### Why not implement everything at once?

- Bank Soal detail is content-heavy and needs RichContent rendering.
- Paket Soal detail has different data shape and may need a later backend fetch for full composition.
- Splitting reduces risk and makes visual QA easier.

---

## Risks & Mitigations

1. **Risk:** Detail panel exposes answer key to roles that should not see it.
   - **Mitigation:** Reuse existing permission guard/role logic. If current list already shows answer key, follow current behavior; otherwise hide key unless reviewer/admin.

2. **Risk:** HTML table in `RichContent` overflows mobile.
   - **Mitigation:** Wrap detail content in `overflow-x-auto`; verify with `MTK-VIII-PG-0032`.

3. **Risk:** Too many panels open and page becomes heavy.
   - **Mitigation:** Single-open behavior with `expandedQuestionId`/`expandedPackageId`.

4. **Risk:** Svelte syntax errors from deeply nested table rows.
   - **Mitigation:** Use small snippets/helpers and run `npm run check` after each task.

5. **Risk:** PM2 stale chunks after build.
   - **Mitigation:** Always `rm -rf build && npm run build`, then `pm2 restart mtsn2kolut-web-admin --update-env`.

---

## Definition of Done

- Plan implemented first for Bank Soal, then optionally Paket Soal.
- `npm run check` passes.
- Clean build passes.
- Protected route smoke returns expected `302`.
- Browser smoke confirms LaTeX and table render in expanded detail.
- Mobile viewport has no horizontal overflow.
- Commit created with focused file changes.
