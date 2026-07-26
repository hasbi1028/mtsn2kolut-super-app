# Web Admin Policy

This app is the SvelteKit admin frontend for MTs Negeri 2 Kolaka Utara.

## Role

- Provide admin and guru-facing UI.
- Act as a thin BFF/proxy to the Go API where needed.
- Own cookies, session handling, and frontend UX.

## Current Baseline — 2026-05-06

- `/bank-soal` is the active standalone Bank Soal list/landing route.
- `/bank-soal/tambah`, `/bank-soal/verifikasi`, and `/bank-soal/impor` own authoring, review/verification, and import UI.
- `/bank-soal/komposer`, `/bank-soal/review`, `/bank-soal/import`, `/cbt/soal*`, `/cbt/bank-soal*`, and `/cbt/questions*` are compatibility redirects to `/bank-soal/*`.
- `/asesmen` is the active assessment launcher; legacy `/cbt*` operational routes redirect to `/asesmen*` or Bank Soal according to context.
- Active frontend fetches use `/api/bank-soal/*` for Bank Soal and `/api/asesmen/*` for assessment. `/api/cbt/*` routes remain live deprecated compatibility proxies to the same Go API semantics for legacy clients, redirects, proxy tests, and staged transition safety.
- Library and Inventory paths are gated for `admin` and `staf` in hooks/route-access helpers.
- Kesiswaan and staff operation gates in SvelteKit are UX/BFF guards; backend role gates remain the source of truth.

## Must Keep

- Do not move backend business rules into SvelteKit.
- Do not introduce direct database access.
- Do not reintroduce SQLite or Drizzle as runtime ownership.
- Prefer server routes only as thin adapters to backend endpoints.
- Keep `/api/*` routes as BFF/proxy adapters only; no direct PostgreSQL, SQLite, or Drizzle runtime storage.

## UI Direction

- Use `shadcn-svelte` for new UI primitives where practical.
- Design for a school administration context, not a startup SaaS dashboard.
- Favor calm, credible, education-appropriate visuals.
- Use clear hierarchy, strong legibility, and practical workflows for admin and guru users.
- Avoid noisy novelty and avoid generic AI-looking templates.

## Design System (Kemenag Green Theme)

- **Typography**: Plus Jakarta Sans (Google Fonts). Weights: 400, 600, 700, 800, 900.
- **Primary color**: `#16a34a` (Kemenag green). Hover: `#15803d`.
- **Background**: `#f8fafc` / `white`. Surface: `#f1f5f9`.
- **Text**: `#0f172a` (primary), `#64748b` (secondary), `#94a3b8` (muted).
- **Cards**: Border radius `18px`, border `1.5px solid #f1f5f9`.
- **Buttons**: Border radius `14px`, uppercase text, `font-weight: 900`.
- **Inputs**: Border radius `14px`, green focus ring `rgba(22,163,74,0.1)`.
- **Badges/Kickers**: `0.65rem`, `font-weight: 900`, `uppercase`, `letter-spacing: 0.18em`.
- See skill `kemenag-superapp-theme` for full design tokens and patterns.

## UX Direction

- Prioritize speed of data entry and clarity of status.
- Tables, forms, schedule views, and monitoring screens should feel dependable and easy to scan.
- Important states must be obvious: active, inactive, draft, published, scheduled, running, success, failed.
- Mobile support must remain acceptable even for admin pages.
- Bank Soal authoring should remain compact and mode-based so teachers are not overwhelmed.

## Technical Direction

- Keep pages and components simple to reason about.
- Reuse shared UI primitives before adding one-off patterns.
- Validate Svelte files before finalizing edits.
- Keep unit coverage around retired route redirects and BFF stream/proxy contracts when changing Bank Soal or assessment routes.

---

## ═══ STATE PERSISTENCE PATTERN ═══

### Masalah

SvelteKit SPA navigation menghancurkan (destroy) komponen lama dan membuat baru. Semua `$state` hilang saat pindah halaman.

### Solusi Wajib — 3 Tingkatan Persistence

Setiap state yang penting (tab, filter, draft, wizard step) WAJIB dipersist dengan salah satu metode berikut:

#### 1️⃣ URL Search Params — untuk view/filter state (PRIORITAS)

```svelte
<script lang="ts">
  import { readUrlParam, writeUrlParam } from '$lib/stores/persistent';

  // Baca dari URL saat mount (survive refresh)
  let activeTab = $state(readUrlParam('tab', 'journal'));

  // Simpan ke URL saat berubah
  function setTab(tab: string) {
    activeTab = tab;
    writeUrlParam('tab', tab);
  }
</script>
```

**Gunakan untuk:** tab aktif, pagination, search query, date range filter, sort order.
**Keuntungan:** bisa di-bookmark, bisa refresh, state aman.

#### 2️⃣ SessionStorage Store — untuk wizard/flow state

Gunakan `persistState()` atau store khusus dari `$lib/stores/`:

```svelte
<script lang="ts">
  import { persistState } from '$lib/stores/persistent';
  import { journalFlow } from '$lib/stores/journal-flow.svelte';

  // Simple: generic store
  const filterStore = persistState('search_filter', '');
  $filterStore = 'cari sesuatu';  // auto-save ke sessionStorage

  // Complex: dedicated store dengan business logic
  $journalFlow = { ...$journalFlow, selectedClassId: kelas.id };
</script>
```

**Gunakan untuk:** multi-step wizard, selected items, data sementara antar halaman.
**Keuntungan:** bertahan selama tab browser (session), simple API.

#### 3️⃣ Draft Auto-Save — untuk form input

```svelte
<script lang="ts">
  import { journalDraft } from '$lib/stores/journal-draft.svelte';
  // atau
  import { createDraft } from '$lib/stores/persistent';

  // Init dari saved draft
  let form = $state({ ...journalDraft });

  // Clear setelah submit sukses
  journalDraft.clear();
</script>
```

**Gunakan untuk:** form create/edit, multi-step form, input panjang.
**Keuntungan:** tidak kehilangan input saat tidak sengaja navigasi.

### Aturan

1. **URL params adalah PRIORITAS UTAMA** — untuk semua view state (tab, filter, pagination).
2. **SessionStorage untuk Wizard** — gunakan `persistState()` atau store dedicated.
3. **Hindari `$state` tanpa persistence** — kecuali state yang benar-benar lokal (popup open/close, hover).
4. **Hindari `localStorage` kecuali** — sidebar collapse, theme preference, data yang perlu survive tab close.
5. **Semua store ada di `$lib/stores/`** — jangan bikin store di komponen. Satu file per domain (journal-flow, employee, dsb).
6. **Gunakan `writeUrlParams()` untuk batch update** — lebih efisien dari individual `writeUrlParam()`.
7. **File utility:** `$lib/stores/persistent.ts` — berisi `persistState()`, `readUrlParam()`, `writeUrlParam()`, `writeUrlParams()`, `createDraft()`.
8. **Untuk tab state:** selalu kombinasikan URL param + `<svelte:window>` atau `onMount` restore agar state survive refresh.

## ═══ CSS & TAILWIND v4 ARCHITECTURE ═══

### CSS Cascade Layer — DaisyUI + Tailwind v4 Import Order (KRITIS)

`app.css` import order menentukan layer priority. **WAJIB**:

```css
/* ✅ BENAR: DaisyUI layer < Tailwind utilities layer */
@import "daisyui/daisyui.css" layer(daisyui);
@import 'tailwindcss';
@plugin 'tailwindcss-animate';
```

**JANGAN** import tailwindcss dulu, baru daisyui. Karena DaisyUI punya universal reset `* { padding: 0 }` di @layer base. Jika daisyui layer > utilities layer, semua padding/margin/width/height Tailwind jadi 0px.

```
Layer order (benar): daisyui → theme → base → components → utilities (MENANG)
Layer order (salah):  theme → base → components → utilities → daisyui (MENANG → reset semua!)
```

### @theme — Hindari Circular var() Reference

**JANGAN** lakukan ini:
```css
@theme { --color-primary: var(--primary); }
:root   { --primary: var(--color-primary); } /* ← CIRCULAR! */
```
Akibat: `--color-primary` = empty, `bg-primary` = transparan, `text-primary` = hitam.

**WAJIB** pakai value eksplisit untuk yang punya circular potensial:
```css
@theme {
    --color-primary: oklch(48% 0.17 145);          /* #16a34a */
    --color-primary-foreground: oklch(98% 0 0);    /* white */
}
```

Untuk non-circular (yang referensi DaisyUI variables via unlayered alias), tetap aman:
```css
@theme {
    --color-background: var(--background); /* :root { --background: var(--color-base-200) } — aman */
}
```

### Semantic Colors — Pastikan Berbeda dari Background

Background: `--color-base-200` = `oklch(97.5% 0 0)` ≈ `#f8fafc`

| Token | Value | Pantauan |
|-------|-------|----------|
| `--color-secondary` | `oklch(96% 0.03 145)` | Hijau lembut — **jangan** `var(--color-base-200)` (sama dgn bg!) |
| `--color-accent` | `oklch(93% 0.08 145)` | Hijau accent — **jangan** `var(--color-base-200)` |
| `--color-muted-foreground` | 60% opacity base-content | Cukup untuk teks sekunder |

### Color Contrast Minimum

| Elemen | Warna | Kontras thd white | Standar |
|--------|-------|--------------------|---------|
| Primary text | `#0f172a` (slate-900) | ~15:1 | ✅ AA/AAA |
| Secondary text | `#64748b` (slate-500) | ~5:1 | ✅ AA |
| Muted text | `#94a3b8` (slate-400) | ~3.2:1 | ❌ Hanya untuk dekoratif |
| Labels/stat | `#64748b` | ~5:1 | ✅ WAJIB untuk label terbaca |

**Rule:** Jangan pakai `#94a3b8` untuk teks yang perlu dibaca. Gunakan `#64748b` minimal.

---

## ═══ COMPONENT PATTERNS ═══

### Server-First (MPA) Pattern — PRIORITAS UNTUK DATA PAGES

Untuk halaman yang menampilkan data (tabel, list, detail), gunakan server-first:

```svelte
<!-- +page.server.ts — load data via Go API di server -->
export const load: PageServerLoad = async ({ fetch }) => {
    const res = await fetch('/api/employees');
    const employees = await res.json();
    return { employees };
};

<!-- +page.svelte — data langsung render di server -->
<script>
    let { data } = $props();
</script>

{#each data.employees as emp}
    <div>{emp.nama}</div>
{/each}
```

**Keuntungan:**
- LLM/AI tools lihat snapshot = data penuh ✅
- Tidak ada loading skeleton untuk initial load
- SEO friendly
- Form actions untuk mutation: `action="?/hapus"` + `<form method="POST">`

### Client-Fetch Pattern — untuk real-time / polling

Gunakan `AsyncContent` + `{#await promise}` hanya jika:
- Data perlu auto-refresh (PUSAKA antrian)
- Data besar yang tidak perlu di-SSR
- Halaman yang jarang diakses

### Native Forms — untuk semua mutation

```svelte
<form method="POST" action="?/hapus">
    <input type="hidden" name="id" value={item.id} />
    <button type="submit">Hapus</button>
</form>
```

**JANGAN** gunakan `fetch()` + `onMount` untuk submit form sederhana. Form actions:
- Bekerja tanpa JavaScript
- Bisa dibaca LLM tools
- Lebih sederhana

---

## ═══ MOBILE RESPONSIVE GUIDE ═══

### Tabel Responsive — Hidden Columns

Untuk tabel dengan banyak kolom di mobile, gunakan `hidden md:table-cell`:

```svelte
<!-- Header -->
<Table.Head>Pegawai</Table.Head>
<Table.Head class="hidden md:table-cell">Identitas</Table.Head>
<Table.Head class="hidden md:table-cell">PUSAKA</Table.Head>
<Table.Head class="text-right">Aksi</Table.Head>

<!-- Data cells — konsisten dengan header -->
<Table.Cell>...</Table.Cell>
<Table.Cell class="hidden md:table-cell">...</Table.Cell>
```

### Action Buttons — Responsive Padding

```svelte
<button class="px-2 md:px-3 py-1 md:py-1.5 text-xs md:text-sm font-medium">
    Edit
</button>
```

Gunakan `gap-1 md:gap-2` pada container flex agar rapat di mobile.

### Info Penting di Mobile — Pindahkan ke Cell Utama

Data yang biasanya di kolom terpisah (misal unit_kerja) bisa dipindahkan ke dalam cell Pegawai:
```svelte
<Table.Cell>
    <div class="font-medium">{nama}</div>
    <div class="hidden md:block text-xs text-muted-foreground">{unit_kerja}</div>
</Table.Cell>
```

---

## ═══ UI AUDIT CHECKLIST ═══

Sebelum deploy, periksa:

- [ ] **Spacing**: `px-8` = 32px? Cek dengan test element di browser console
- [ ] **Primary color**: `bg-primary` = hijau (`oklch(48% .17 145)`)?
- [ ] **Secondary/accent**: tidak sama dengan background?
- [ ] **Contrast**: teks `text-muted-foreground`, label `#94a3b8`? Ganti ke `#64748b` jika perlu dibaca
- [ ] **Mobile viewport**: apakah tabel overflow? Action button terlalu besar?
- [ ] **Snapshot LLM**: apakah browser snapshot menampilkan data atau hanya skeleton?
- [ ] **CSS layers**: daisyui sebelum tailwindcss di import?
