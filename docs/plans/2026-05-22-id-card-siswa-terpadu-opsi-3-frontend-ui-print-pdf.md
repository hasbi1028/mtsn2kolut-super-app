# ID Card Siswa Terpadu Opsi 3 — Frontend/UI/Print/PDF Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Membangun frontend web-admin dan portal untuk **ID Card Siswa Terpadu MTsN 2 Kolaka Utara Opsi 3**: kartu identitas fisik dengan QR sebagai shortcut identitas/login challenge + PIN, preview depan-belakang, cetak PDF massal per kelas, verifikasi QR publik terbatas, serta mode petugas presensi/perpustakaan/CBT.

**Architecture:** Core API tetap pemilik data/lifecycle kartu dan QR; SvelteKit web-admin hanya UI + BFF proxy. UI memakai desain **Alternatif A hijau-emas** yang sudah disukai user, dengan renderer kartu shared agar preview, cetak PDF, dan portal konsisten. QR bukan password; semua login/scan sensitif tetap butuh PIN, role, scope, audit, dan policy backend.

**Tech Stack:** SvelteKit 2, Svelte 5 runes, Tailwind v4, shadcn-svelte, existing `AsyncContent`, BFF `/api/*` routes, Core API Go/PostgreSQL/sqlc. PDF/print awal menggunakan browser print CSS (`@page`) dari halaman cetak; opsi server-side PDF/Playwright hanya jika nanti dibutuhkan.

---

## 0. Prinsip Produk & Keamanan yang Harus Terkunci

- Nama fitur di UI: **ID Card Siswa Terpadu** atau **Kartu Siswa**, bukan “Kartu CBT”.
- QR berisi handle/token random, bukan NISN/NIS/nama/password/PIN/token ujian.
- QR hanya memulai konteks:
  - publik: verifikasi terbatas;
  - portal: identifikasi kartu → form PIN → session;
  - petugas: scan kartu → backend cek role/scope/konteks → aksi modul.
- CBT tetap memakai token ruang/sesi terpisah; kartu hanya validasi identitas peserta.
- Semua error UI harus aman: “Kartu tidak dapat diverifikasi” tanpa membocorkan apakah token valid, kartu hilang, atau siswa nonaktif untuk publik.
- Admin/kesiswaan boleh melihat alasan lengkap sesuai permission.

---

## 1. Information Architecture & Routes

### 1.1 Sidebar / Navigasi Web Admin

**Modify:** `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

Tambahkan item di group **Siswa & Orang Tua**:

```ts
{ href: '/kesiswaan/kartu-siswa', label: 'Kartu Siswa', icon: 'id-card', roles: ['admin', 'kesiswaan'], permissions: ['student_card.read'] }
```

Catatan:
- Jika icon registry belum mengenal `id-card`, tambahkan mapping lucide `IdCard`/`IdCardLanyard` di komponen icon sidebar yang relevan.
- Jangan pindahkan `/students`; fitur kartu masuk submodule Kesiswaan.

### 1.2 Web Admin Routes

Create routes:

- `apps/web-admin/src/routes/kesiswaan/kartu-siswa/+page.svelte`
  - daftar kartu/siswa, filter kelas/status, bulk issue/activate/print, preview drawer.
- `apps/web-admin/src/routes/kesiswaan/kartu-siswa/[card_id]/+page.svelte`
  - detail kartu, lifecycle, audit, actions revoke/reissue/reset PIN.
- `apps/web-admin/src/routes/kesiswaan/kartu-siswa/cetak/+page.svelte`
  - print layout massal per kelas/filter, browser print/PDF.
- `apps/web-admin/src/routes/kesiswaan/kartu-siswa/scanner/+page.svelte`
  - mode petugas scan QR untuk presensi/perpustakaan/CBT dengan context selector.

BFF proxy routes:

- `apps/web-admin/src/routes/api/kesiswaan/student-cards/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/student-cards/[id]/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/student-cards/[id]/print-token/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/student-cards/[id]/activate/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/student-cards/[id]/revoke/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/students/[student_id]/cards/reissue/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/students/[student_id]/portal-pin/reset/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/student-cards/[id]/audit/+server.ts`
- `apps/web-admin/src/routes/api/kesiswaan/card-scans/resolve/+server.ts`

### 1.3 Public / Portal Routes

Create routes in existing web-admin SvelteKit app (current portal lives under `apps/web-admin/src/routes/portal/...`):

- `apps/web-admin/src/routes/s/idc/[token]/+page.svelte`
  - halaman publik verifikasi QR terbatas + tombol “Masuk Portal dengan PIN”.
- `apps/web-admin/src/routes/portal/siswa/qr-login/+page.svelte`
  - flow scan QR + PIN bila siswa membuka dari QR atau kamera.
- Extend `apps/web-admin/src/routes/portal/siswa/+page.svelte`
  - card status widget: status kartu, masa berlaku, tombol lapor kartu hilang, ganti PIN.

BFF proxy routes:

- `apps/web-admin/src/routes/api/public/student-card/verify/+server.ts`
- `apps/web-admin/src/routes/api/portal/siswa/card-login/start/+server.ts`
- `apps/web-admin/src/routes/api/portal/siswa/card-login/verify-pin/+server.ts`
- `apps/web-admin/src/routes/api/portal/siswa/card-status/+server.ts`
- `apps/web-admin/src/routes/api/portal/siswa/card-lost-report/+server.ts`

---

## 2. Shared Frontend Types & Client Modules

### Task 2.1: Create student-card client API module

**Files:**
- Create: `apps/web-admin/src/lib/client/student-cards.ts`
- Test: `apps/web-admin/src/lib/client/student-cards.test.ts`

Types to define:

```ts
export type StudentCardStatus = 'draft' | 'active' | 'suspended' | 'lost' | 'revoked' | 'replaced' | 'expired' | 'damaged';

export type StudentCardListItem = {
  id: string;
  student_id: string;
  card_no: string;
  student_name: string;
  nis: string;
  nisn_masked?: string;
  class_id: string;
  class_name: string;
  class_code: string;
  photo_url?: string | null;
  status: StudentCardStatus;
  issued_at?: string | null;
  printed_at?: string | null;
  expires_at?: string | null;
  qr_public_url?: string | null; // only returned in print-token scope
  qr_svg?: string | null;        // optional backend-rendered QR
  last_seen_at?: string | null;
};

export type StudentCardListPayload = {
  items: StudentCardListItem[];
  total: number;
  page: number;
  page_size: number;
  classes: Array<{ id: string; code: string; name: string }>;
  status_counts: Record<StudentCardStatus | 'none', number>;
};
```

Client functions:

- `fetchStudentCards(filters, fetcher)`
- `issueStudentCards(studentIds, fetcher)`
- `fetchStudentCardDetail(cardId, fetcher)`
- `fetchStudentCardPrintBatch(filters, fetcher)`
- `activateStudentCard(cardId, fetcher)`
- `revokeStudentCard(cardId, reason, fetcher)`
- `reissueStudentCard(studentId, reason, fetcher)`
- `resetStudentPortalPin(studentId, fetcher)`
- `resolveCardScan(payload, fetcher)`

Use existing `clientApiPath` + `readClientApiData` from `$lib/client/api`; remember `readClientApiData` unwraps `{items}`/`{data}` envelopes.

### Task 2.2: Portal card client module

**Files:**
- Modify: `apps/web-admin/src/lib/client/student-portal.ts`
- Test: `apps/web-admin/src/lib/client/student-portal.test.ts`

Add types/functions:

- `StudentPortalCardStatus`
- `startStudentCardLogin(tokenOrUrl, fetcher)`
- `verifyStudentCardPin(challengeId, pin, fetcher)`
- `fetchStudentPortalCardStatus(fetcher)`
- `reportStudentCardLost(reason, fetcher)`

---

## 3. Shared UI Components

Create component folder:

`apps/web-admin/src/lib/components/student-card/`

### Task 3.1: `StudentCardPreview.svelte`

**Purpose:** Renderer visual kartu CR80 depan-belakang untuk preview admin, portal, dan print.

Props:

```ts
type Props = {
  card: StudentCardRenderModel;
  side?: 'front' | 'back' | 'both';
  size?: 'screen' | 'print';
  showPrintGuides?: boolean;
  qrState?: 'ready' | 'loading' | 'failed' | 'missing';
};
```

Design Alternatif A hijau-emas:

- CR80 ratio: `85.60mm × 53.98mm` for print; screen uses `aspect-[856/540]`.
- Front:
  - header hijau tua, aksen emas tipis;
  - logo madrasah kiri;
  - title “KARTU IDENTITAS SISWA”;
  - photo frame 25×32mm;
  - nama besar, NIS internal/card no, kelas, tahun/masa berlaku;
  - footer domain/contact.
- Back:
  - QR besar kanan/tengah;
  - card no dan instruksi pengembalian;
  - warning: “QR bukan password. Login tetap memakai PIN.”;
  - optional Code 128 placeholder fallback jika nanti backend menyediakan barcode.

Fallback states:

- Tanpa foto: tampilkan silhouette/inisial di frame, label kecil “Foto belum tersedia”. Jangan pecah layout.
- QR gagal/missing: tampilkan kotak dashed + teks “QR belum tersedia — cetak ulang token”, dan disable checkbox cetak untuk kartu tersebut.
- Nama panjang: `line-clamp-2`, ukuran font menurun via class `text-[...]`, jangan overflow.

### Task 3.2: `StudentCardStatusBadge.svelte`

Map status:

- `active`: hijau, “Aktif”
- `draft`: abu/kuning, “Draft”
- `suspended`: kuning, “Ditahan”
- `lost`: merah, “Hilang”
- `revoked`: merah gelap, “Dicabut”
- `replaced`: biru/abu, “Diganti”
- `expired`: abu, “Kedaluwarsa”
- `damaged`: oranye, “Rusak”
- no card: outline, “Belum ada kartu”

### Task 3.3: `StudentCardFilterBar.svelte`

Use structured operator filter pattern:

- Row 1: search nama/NIS/card no; quick actions kanan.
- Row 2: kelas/rombel, status kartu, status siswa, angkatan/tahun.
- Row 3: printed state, photo state, QR state, sort.
- Active chips: kelas, status, tanpa foto, belum dicetak, aktif.

Tailwind safety:

- every input/select: `min-w-0 w-full`;
- filters split at `sm/md/lg/xl`, no `xs:` unless custom breakpoint exists;
- table wrapper `overflow-x-auto`, table `min-w-[900px]`.

### Task 3.4: `QrScannerPanel.svelte`

**Purpose:** reusable scanner UI for petugas and QR-login page.

Capabilities:

- Primary: paste/manual input QR token/url.
- Progressive enhancement: camera scanner via browser BarcodeDetector/html5-qrcode only after permission; feature-detect and show fallback if unavailable.
- Large mobile buttons: “Mulai Kamera”, “Tempel Token”, “Input Manual”.
- Emits `scan` event with raw token/url; parent calls backend.
- Error copy safe and actionable.

---

## 4. Web Admin Main Page: `/kesiswaan/kartu-siswa`

### Layout

**Header:**

- Title: “Kartu Siswa”
- Subtitle: “Terbitkan, cetak, dan kelola ID Card Siswa Terpadu.”
- Buttons:
  - “Terbitkan Kartu” (bulk issue untuk siswa terpilih/tanpa kartu)
  - “Cetak PDF” → navigasi ke `/kesiswaan/kartu-siswa/cetak` dengan query filter/selection
  - “Mode Scan Petugas” → `/kesiswaan/kartu-siswa/scanner`

**Metric cards:**

- Aktif
- Belum punya kartu
- Perlu foto
- Kartu bermasalah (lost/revoked/suspended)

**Content desktop:**

- Filter panel.
- Table columns:
  - checkbox;
  - Siswa (foto mini, nama, NIS, kelas);
  - No Kartu;
  - Status;
  - Masa berlaku;
  - Cetak terakhir;
  - Last scan;
  - Aksi.
- Row actions:
  - Preview;
  - Detail;
  - Aktifkan;
  - Reset PIN;
  - Cetak Ulang/Reissue;
  - Tandai Hilang/Cabut.

**Content mobile:**

- Card list one-column.
- Sticky bottom bulk bar when selected: `N dipilih · Terbitkan · Cetak`.
- Preview opens full-screen sheet with front/back tabs.

### UX Operator

- Bulk issue must show preflight modal:
  - jumlah siswa dipilih;
  - berapa sudah punya kartu aktif;
  - berapa tanpa foto;
  - berapa tidak aktif;
  - action default: buat hanya untuk siswa belum punya kartu.
- Bulk print preflight:
  - kartu yang siap cetak;
  - kartu tanpa QR/token tidak siap;
  - kartu tanpa foto boleh dicetak dengan placeholder only if operator confirms.
- Reissue modal must require reason: hilang, rusak, bocor/difoto, data salah, lainnya.
- Revoke/lost modal must warn: “QR lama langsung ditolak untuk login/scan.”

---

## 5. Detail Page: `/kesiswaan/kartu-siswa/[card_id]`

### Sections

1. **Identity header**: photo, name, class, status, card no.
2. **Card preview** front/back.
3. **Lifecycle actions**:
   - Aktifkan;
   - Suspend/aktifkan kembali;
   - Tandai hilang;
   - Cabut;
   - Reissue kartu;
   - Reset PIN portal siswa.
4. **Audit timeline**:
   - created/printed/activated/scanned/denied/reissued/reset pin;
   - module chips: portal, attendance, library, cbt, admin.
5. **Safe diagnostics**:
   - last scan time;
   - last failure reason for admin only;
   - QR prefix only, never raw token.

### Mobile

- Preview collapses to tabs.
- Actions in “Tindakan” sheet, not a wide button row.
- Audit timeline cards with compact timestamps.

---

## 6. Print/PDF Page: `/kesiswaan/kartu-siswa/cetak`

### Objective

Mencetak massal kartu depan-belakang per kelas/filter dengan hasil PDF stabil dari browser print.

### Screen Layout

**Non-print UI (`print:hidden`):**

- Header “Cetak Kartu Siswa”.
- Class/filter summary.
- Controls:
  - Kelas/rombel;
  - status kartu: active/draft/printed/unprinted;
  - sisi cetak: depan saja, belakang saja, depan-belakang;
  - layout: CR80 grid A4;
  - opsi: tampilkan garis potong, cetak placeholder foto, sertakan kartu draft.
- Preflight panel:
  - Total kartu dimuat;
  - Siap cetak;
  - Tanpa foto;
  - QR gagal/missing;
  - Kartu nonaktif/revoked (excluded by default).
- Buttons:
  - “Muat Data”;
  - “Preview Semua”;
  - “Cetak / Simpan PDF” (`window.print()`);
  - “Kembali”.

**Print UI (`print:block`):**

- A4 portrait recommended.
- CSS:

```css
@media print {
  @page { size: A4 portrait; margin: 8mm; }
  .screen-only { display: none !important; }
  .print-sheet { display: grid; grid-template-columns: repeat(2, 85.6mm); gap: 5mm 6mm; }
  .id-card-print { width: 85.6mm; height: 53.98mm; break-inside: avoid; page-break-inside: avoid; }
  .page-break { break-after: page; }
}
```

### Duplex Strategy

Implement initially as two print sections:

- **Mode depan-belakang manual aman:**
  1. Print all fronts page 1..n.
  2. Operator flips paper per printer SOP.
  3. Print all backs in same order, mirrored alignment if necessary.
- UI text must include printer note:
  - “Untuk duplex manual, cetak bagian depan dulu, masukkan kembali kertas sesuai arah printer, lalu cetak bagian belakang.”

Avoid generating a misleading “perfect duplex” if printer behavior unknown.

### PDF Performance

- Do not auto-load all students on page load.
- Require explicit “Muat Data”.
- Default limit per print batch: one class/rombel; if “semua kelas”, show confirmation.
- Show “Ditampilkan N dari M kartu”; if partial due to limit, disable print until operator chooses full class or acknowledges partial.

### Print Fallbacks

- Foto hilang: placeholder silhouette/inisial; show preflight warning.
- QR gagal: exclude by default; allow “Cetak tanpa QR” only for identity-only draft proof, watermark “DRAFT — QR BELUM AKTIF”.
- Kartu revoked/lost/replaced: excluded; if admin forces archive print, watermark status besar.

---

## 7. Public Verification Page: `/s/idc/[token]`

### UX

Show limited public data only:

- logo + “Verifikasi Kartu Siswa MTsN 2 Kolaka Utara”;
- status high-level:
  - valid/aktif: “Kartu terdaftar dan aktif”;
  - invalid/expired/revoked: generic “Kartu tidak dapat diverifikasi” for public;
- display allowed fields only:
  - initials or first name + masked name if policy allows;
  - class code optionally;
  - card no;
  - photo thumbnail optional only if policy approves; safer default: no full photo public.
- CTA:
  - “Masuk Portal Siswa dengan PIN” → `/portal/siswa/qr-login?challenge=...` or start flow.
  - “Kembalikan kartu ini ke MTsN 2 Kolaka Utara” contact.

### Failure States

- Invalid token: generic not found.
- Card lost/revoked: generic not verifiable + return instruction.
- Backend down: “Verifikasi belum dapat dilakukan. Coba lagi nanti.”
- Never expose token, exact status reason, NISN, parent phone, DOB, address.

---

## 8. Portal QR + PIN Login

### Route: `/portal/siswa/qr-login`

Flow:

1. If URL contains token/challenge from `/s/idc/[token]`, call `startStudentCardLogin`.
2. Show identity hint only after backend allows:
   - “Kartu dikenali: A*** · VII A” or “Siswa MTsN 2 Kolaka Utara”.
3. PIN form:
   - 6-digit input, numeric keypad on mobile.
   - submit `verifyStudentCardPin(challengeId, pin)`.
4. On success redirect to `/portal/siswa`.
5. On `pin_reset_required`, redirect/show “Buat PIN baru” flow if backend supports.

### Mobile Layout

- Full-width card, large scan/PIN controls.
- Camera scan fallback manual input.
- Clear privacy copy: “QR bukan password. Tetap masukkan PIN pribadi.”

### Security UX

- After 3–5 failed PIN attempts show lockout copy from backend.
- Do not say “PIN salah untuk siswa X” before authenticated.
- Never cache raw QR token in localStorage; challenge ID only in memory or short-lived query/server session.

---

## 9. Portal Siswa Card Widget

Modify `apps/web-admin/src/routes/portal/siswa/+page.svelte`:

Add a compact section after profile summary:

- “Kartu Siswa Saya”
- Status badge: aktif/belum aktif/hilang/dicabut.
- Card no masked.
- Masa berlaku.
- Last printed/last seen.
- Actions:
  - “Lapor Kartu Hilang” opens confirm reason form.
  - “Ganti PIN” links existing account/password flow or new PIN flow.

Do not show QR token or printable QR to student unless product explicitly allows digital card later. For Opsi 3 physical card, portal is status/reporting only.

---

## 10. Mode Petugas Scanner: `/kesiswaan/kartu-siswa/scanner`

### Context Selector

At top, choose mode:

- **Presensi**
  - context: kelas/kegiatan/gerbang/sesi;
  - backend endpoint eventually records attendance idempotently.
- **Perpustakaan**
  - context: kunjungan, pinjam, kembali;
  - after scan, show member card and next field barcode buku.
- **CBT**
  - context: kegiatan/sesi/ruang;
  - validate participant identity only; show reminder token ruang tetap terpisah.

### Scanner Result Panel

On successful resolve:

- large photo/placeholder;
- nama, kelas, status kartu;
- module-specific status:
  - presensi: hadir tercatat / sudah tercatat;
  - library: anggota ditemukan / limit pinjam;
  - CBT: peserta sesuai ruang / bukan peserta ruang ini.

On denied:

- red/yellow panel with actionable reason for authorized staff:
  - kartu hilang/revoked;
  - siswa nonaktif;
  - di luar scope kelas/ruang;
  - token invalid;
  - rate limited.

### Mobile Scanner UX

- Camera viewport top, result bottom.
- Sticky “Scan berikutnya”.
- Haptic/sound optional; do not add dependency until needed.
- Manual input always available.

---

## 11. BFF Proxy Pattern

All SvelteKit `/api/kesiswaan/*`, `/api/portal/siswa/*`, and `/api/public/*` routes must be thin adapters:

- read session/cookies;
- forward method, query, JSON body to Core API;
- preserve response status;
- never direct PostgreSQL;
- never generate QR token in frontend;
- no business rules beyond UX-friendly input normalization.

Use existing proxy helpers if available in project; otherwise mirror nearby BFF routes such as `apps/web-admin/src/routes/api/bank-soal/soal-support/authors/+server.ts` and assessment proxy routes.

---

## 12. Styling & Layout Standards

- Follow Alternatif A: green/gold, official madrasah, calm credible layout.
- Use `Card`, `Button`, `Input`, `Badge`, dialog/sheet components from shadcn-svelte where practical.
- Use `AsyncContent` or `{#await}` promise lifecycle for data loading.
- Initial page: skeletons, not plain “loading”.
- Use `role="alert"` for mutation errors and print preflight warnings.
- Table accessibility: caption, `scope="col"`, row action `aria-label`.
- Avoid `xs:` Tailwind unless theme defines it; use default breakpoints.
- Use `overflow-x-auto` wrappers and `min-w-0` on grid items.

---

## 13. Tests & Verification Plan

### Unit tests

- `apps/web-admin/src/lib/client/student-cards.test.ts`
  - query serialization for filters;
  - repeated status params if multi-select;
  - unwrap response correctly;
  - mutation body shape.
- `apps/web-admin/src/lib/client/student-portal.test.ts`
  - card login start/verify endpoints;
  - no raw token persistence.
- Component tests if existing harness supports:
  - `StudentCardStatusBadge` status labels;
  - `StudentCardPreview` fallback no photo/QR failed.

### Manual UI verification

- Desktop `/kesiswaan/kartu-siswa` no horizontal page overflow at 1366px and 1024px.
- Mobile viewport 390px:
  - filter collapses;
  - cards fit;
  - sticky bulk bar not hiding content.
- Print page:
  - A4 preview cards are CR80 ratio;
  - fronts/backs aligned;
  - QR failed cards excluded/watermarked correctly.
- Public `/s/idc/[token]`:
  - valid token shows limited data;
  - invalid/revoked shows safe generic state.
- Portal QR+PIN:
  - scan → PIN → portal success;
  - wrong PIN lockout state;
  - no direct login from QR.
- Scanner modes:
  - presensi, library, CBT context selectors;
  - denied states legible.

### Build command

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
rm -rf build && npm run build
```

Expected: build passes without Svelte/TS errors.

---

## 14. Implementation Task Breakdown

### Task 1: Add route/nav skeletons only

**Objective:** Expose Kartu Siswa IA without business logic.

**Files:**
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Create: route placeholders listed in sections 1.2 and 1.3.

**Verify:** `npm run build` passes; sidebar item appears for admin/kesiswaan once permissions exist.

### Task 2: Add client modules and tests

**Objective:** Centralize API contracts.

**Files:**
- Create: `src/lib/client/student-cards.ts`
- Modify: `src/lib/client/student-portal.ts`
- Create/modify tests.

**Verify:** targeted Vitest if configured; final `npm run build`.

### Task 3: Build shared card preview components

**Objective:** One renderer for screen and print.

**Files:**
- Create: `src/lib/components/student-card/StudentCardPreview.svelte`
- Create: `StudentCardStatusBadge.svelte`
- Create: `StudentCardEmptyPhoto.svelte` optional.

**Verify:** render with mock data in route skeleton; inspect no overflow.

### Task 4: Build list/filter page

**Objective:** Operator can list/filter/select and preview cards.

**Files:**
- Implement: `/kesiswaan/kartu-siswa/+page.svelte`
- Create: `StudentCardFilterBar.svelte`

**Verify:** mocked/real API states: empty, loaded, error, selected bulk bar.

### Task 5: Build detail/lifecycle page

**Objective:** One card management page with audit and actions.

**Files:**
- Implement: `/kesiswaan/kartu-siswa/[card_id]/+page.svelte`

**Verify:** status-specific buttons visible/disabled correctly; dangerous actions require reason.

### Task 6: Build print/PDF page

**Objective:** Produce print-safe CR80 front/back PDF from browser.

**Files:**
- Implement: `/kesiswaan/kartu-siswa/cetak/+page.svelte`
- Extend card renderer print classes.

**Verify:** browser print preview A4; no screen UI printed; fallback states correct.

### Task 7: Build public verification page

**Objective:** QR opens safe limited verification and portal login CTA.

**Files:**
- Implement: `/s/idc/[token]/+page.svelte`
- Add BFF public verify route.

**Verify:** invalid/revoked states do not leak sensitive details.

### Task 8: Build QR+PIN portal login flow

**Objective:** QR shortcut starts challenge, PIN creates portal session.

**Files:**
- Implement: `/portal/siswa/qr-login/+page.svelte`
- Extend `student-portal.ts`.

**Verify:** PIN form works, no QR-only login path.

### Task 9: Add portal card widget

**Objective:** Student can see card status and report lost card.

**Files:**
- Modify: `/portal/siswa/+page.svelte`

**Verify:** mobile and desktop layout remain clean; preview admin mode safe.

### Task 10: Build scanner mode page

**Objective:** Staff can use one scanner page for presensi/library/CBT contexts.

**Files:**
- Implement: `/kesiswaan/kartu-siswa/scanner/+page.svelte`
- Create: `QrScannerPanel.svelte`

**Verify:** manual input works even without camera; contexts produce correct payload.

### Task 11: Add BFF proxies

**Objective:** Wire UI to Core API without direct DB/business logic.

**Files:**
- Create BFF routes listed in section 1.2/1.3.

**Verify:** proxy tests or smoke with mocked backend responses; status codes preserved.

### Task 12: Final QA and responsive/print audit

**Objective:** Ensure operator-ready quality.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
rm -rf build && npm run build
```

Manual browser checks:

- `/kesiswaan/kartu-siswa`
- `/kesiswaan/kartu-siswa/cetak`
- `/s/idc/demo-invalid`
- `/portal/siswa/qr-login`
- `/kesiswaan/kartu-siswa/scanner`

---

## 15. Backend/API Dependencies to Confirm Before Frontend Implementation

Frontend can be built with mock/demo states first, but production wiring needs Core API to expose:

- list cards with class/status filters and pagination;
- issue/reissue/activate/revoke/reset PIN endpoints;
- print-token batch endpoint returning QR render payload only for authorized print operation;
- public verify endpoint with limited response shape;
- portal card-login challenge + verify PIN endpoints;
- scanner resolve endpoint with context and module-specific policy;
- audit endpoint.

If backend is not ready, implement UI with disabled actions + “Menunggu API kartu siswa” notices, but keep route/component contracts aligned with this plan.

---

## 16. Acceptance Criteria

- Admin/kesiswaan can open **Kesiswaan → Kartu Siswa**.
- Operator can filter by kelas/status and preview front/back card.
- Print page can generate browser PDF for one class, CR80 front/back, with safe preflight.
- Cards without photos do not break layout; cards without QR are blocked or watermarked.
- Public QR verification reveals only limited safe data.
- Portal QR login requires PIN and never logs in from QR alone.
- Student portal shows card status and lost-card reporting.
- Staff scanner supports presensi/perpustakaan/CBT contexts with clear allowed/denied states.
- No SvelteKit route reads PostgreSQL directly; all API calls go through BFF/Core API.
- `rm -rf build && npm run build` passes.
