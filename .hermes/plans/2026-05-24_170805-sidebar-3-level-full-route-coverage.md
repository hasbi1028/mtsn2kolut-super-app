# Sidebar 3-Level Full Route Coverage Plan

> **For Hermes:** implement with subagent-driven-development where practical, but parent must verify all diffs/tests before deploy.

## Goal

Membuat sidebar web-admin MTsN 2 Kolut mendukung menu 3-level:

```text
Level 1: Area kerja madrasah
Level 2: Workflow/sub-area
Level 3: Halaman/aksi konkret
```

Target UX: semua route admin/operasional statis yang penting terjangkau dari sidebar/command palette, tanpa memasukkan route dinamis/print/detail sebagai menu statis yang membingungkan.

## Scope

- Frontend only: `apps/web-admin`.
- No DB change.
- No backend change.
- Preserve RBAC/security behavior; sidebar remains UX visibility only.
- Preserve existing pins/recent/quick access/command palette.
- Deploy: rebuild and restart `mtsn2kolut-web-admin` only after validation.

## Agent Swarm Findings

Swarm audit menemukan:

- Total route pages: 123 `+page.svelte`.
- Klasifikasi:
  - sidebar leaf/kandidat leaf: 85
  - contextual/dynamic/detail/print: 21
  - public/portal/auth/runtime: 16
  - hidden/legacy: 1 (`/academic`)
- Current sidebar model masih 2-level: `group -> items`.
- Current consumers assume flat leaf list:
  - RBAC filter
  - active state
  - pin/recent
  - quick access
  - command palette
  - RBAC preview
  - tests
- Existing config/test mismatch: config has `/bank-soal/laporan`; old test may not include it. Treat `/bank-soal/laporan` as valid route and update test expectations.

## Route Coverage Rules

### Include as sidebar leaf

Static protected/admin/operator routes such as:

- `/akademik`, `/akademik/rombel`, `/akademik/jadwal`
- `/bank-soal`, `/bank-soal/daftar`, `/bank-soal/tambah`, `/bank-soal/impor`, `/bank-soal/verifikasi`, `/bank-soal/penerbitan`, `/bank-soal/mapel-kd`, `/bank-soal/analisis-butir`, `/bank-soal/laporan`, `/bank-soal/pengaturan`
- `/asesmen`, `/asesmen/persiapan`, `/asesmen/pelaksanaan`, `/asesmen/pengawasan`, `/asesmen/hasil`, `/asesmen/non-tes`, `/asesmen/kegiatan`, `/asesmen/paket`, `/asesmen/sesi`, `/asesmen/aplikasi-siswa`, `/asesmen/aplikasi-siswa/matrix`, `/asesmen/aplikasi-siswa/release`
- `/tu/*` static pages, `/document-cycles`, `/governance/actions/*` static pages
- `/library`, `/inventory`, `/website`, `/pusaka`, `/settings/*`

### Do not include as static sidebar leaf

Dynamic/contextual routes:

- any route with `[id]`, `[rid]`, `[slug]`, `[token]`, `[participant_id]`
- detail pages reachable from parent list
- print/report detail pages reachable from contextual buttons

### Public/runtime/auth routes

Do not clutter admin sidebar with runtime/public details. Keep portal/public only where operator-friendly:

- portal leaf for `/portal/siswa`, `/portal/orang-tua`, `/jadwal`
- website admin manages public content; public `/berita`, `/pengumuman`, `/kontak`, `/profil`, `/ppdb` should remain preview/contextual links, not main admin workflow leaf unless current config already intentionally includes them.
- `/login`, `/maintenance`, `/ujian`, `/s/idc/[token]`, CBT runtime routes are not normal admin leaves.

## Data Model Plan

Add recursive compatible sidebar model in `sidebar-config.ts` or helper files:

```ts
export type SidebarLeafItem = {
  kind?: 'item';
  href: string;
  label: string;
  icon: string;
  roles?: string[];
  permissions: string[];
  roleFallbacks?: string[];
  allowAuthenticatedFallback?: boolean;
  pinnable?: boolean;
};

export type SidebarFolderItem = {
  kind: 'folder';
  id: string;
  label: string;
  icon: string;
  permissions?: string[];
  roleFallbacks?: string[];
  allowAuthenticatedFallback?: boolean;
  children: SidebarNavNode[];
  pinnable?: false;
};

export type SidebarNavNode = SidebarLeafItem | SidebarFolderItem;
export type SidebarFlatItem = SidebarLeafItem & {
  group: string;
  ancestors: string[];
  breadcrumb: string[];
};
```

Compatibility rule: `kind?: 'item'` keeps existing flat leaf entries valid.

## Helper Plan

Create `sidebar-tree.ts`:

- `isSidebarFolder(node)`
- `isSidebarLeaf(node)`
- `flattenSidebarNavGroups(groups): SidebarFlatItem[]`
- `filterSidebarTreeByAccess(groups, roles, permissions): SidebarNavGroup[]`
- `findSidebarActiveTrail(pathname, groups): { group, ancestors, href } | null`
- optional `leafBreadcrumb(item)` helper

Use flattened leaves for:

- active href
- command palette
- quick access
- pin/recent normalization
- duplicate href tests
- RBAC preview

## Component Plan

### `Sidebar.svelte`

- Replace `nav.flatMap(section.items)` with `flattenSidebarNavGroups(nav)`.
- Keep `dashboardNavItem` special root.
- Pin/recent stay leaf-only.
- Active group computed from flattened leaf `.group`.
- Command items include `ancestors`/breadcrumb.

### `SidebarNavSection.svelte`

- Render recursive nodes.
- Folder = `<button type="button" aria-expanded aria-controls>`.
- Leaf = existing `<a href>` markup with pin button.
- Parent folders are not pinnable and do not call `rememberRecent`.
- Active child opens ancestor folder.
- Mobile leaf click closes drawer.
- Collapsed desktop remains safe: no leaked long labels; use title/breadcrumb.

### Command Palette

- Only leaf items are command results.
- Search matches label, href, group, ancestors.
- Display breadcrumb: `Bank Soal › Mutu & Laporan › Analisis Butir`.

### Quick Access

- Leaf-only.
- Existing storage format stays `string[]` hrefs.

## Proposed IA Groups

### Beranda

- Ringkasan
  - Dashboard `/`
  - Notifikasi `/notifications`
  - Kesiapan Akademik & Rapor `/akademik/kesiapan`

### Portal

- Portal Pengguna
  - Portal Siswa `/portal/siswa`
  - Portal Orang Tua `/portal/orang-tua`
  - Jadwal Saya `/jadwal`

### Akademik

- Data Tahun & Kurikulum
  - Ringkasan Akademik `/akademik`
  - Tahun Ajaran `/akademik/tahun-ajaran`
  - Struktur Kurikulum `/akademik/kurikulum`
  - Mata Pelajaran `/akademik/mapel`
- Rombel & Pengajaran
  - Rombel `/akademik/rombel`
  - Guru Mapel `/akademik/guru-mapel`
  - Beban Guru `/akademik/beban-guru`
- Jadwal & Jurnal
  - Jam Pelajaran `/akademik/jam-pelajaran`
  - Jadwal Pelajaran `/akademik/jadwal`
  - Jurnal Kelas `/journal`

### Siswa & Orang Tua

- Data Siswa
  - Siswa `/students`
  - Orang Tua `/parents`
  - Kesiswaan `/kesiswaan`
- Kartu Siswa
  - Cetak Kartu Siswa `/kesiswaan/kartu-siswa`
  - Scanner Kartu `/kesiswaan/kartu-siswa/scan`

### Nilai & Rapor

- Penilaian
  - Input Nilai `/grades`
  - Penilaian Non-Tes `/asesmen/non-tes`
- Rapor
  - Rapor Siswa `/grades/rapor`
  - Kesiapan Akademik & Rapor `/akademik/kesiapan` only if duplicate accepted; otherwise keep canonical in Beranda/Akademik.

Avoid duplicate hrefs if current tests require uniqueness; prefer one canonical leaf and contextual links elsewhere.

### Bank Soal

- Kelola Soal
  - Dashboard Bank Soal `/bank-soal`
  - Daftar Soal `/bank-soal/daftar`
  - Tambah Soal `/bank-soal/tambah`
  - Impor Soal `/bank-soal/impor`
  - Cetak Soal `/bank-soal/cetak`
- Review & Penerbitan
  - Verifikasi Soal `/bank-soal/verifikasi`
  - Penerbitan Soal `/bank-soal/penerbitan`
- Referensi & Analisis
  - Mapel & KD `/bank-soal/mapel-kd`
  - Analisis Butir `/bank-soal/analisis-butir`
  - Laporan Bank Soal `/bank-soal/laporan`
  - Pengaturan Bank Soal `/bank-soal/pengaturan`

### Asesmen Ujian

- Dashboard & Persiapan
  - Dashboard Ujian `/asesmen`
  - Persiapan Ujian `/asesmen/persiapan`
  - Kegiatan Ujian `/asesmen/kegiatan`
  - Paket Soal Ujian `/asesmen/paket`
  - Sesi Ujian `/asesmen/sesi`
- Pelaksanaan & Pengawasan
  - Pelaksanaan Ujian `/asesmen/pelaksanaan`
  - Pengawasan Ujian `/asesmen/pengawasan`
- Hasil & Aplikasi
  - Hasil Ujian `/asesmen/hasil`
  - Penilaian Non-Tes `/asesmen/non-tes`
  - Aplikasi Siswa `/asesmen/aplikasi-siswa`
  - Matrix Aplikasi `/asesmen/aplikasi-siswa/matrix`
  - Release Aplikasi `/asesmen/aplikasi-siswa/release`

### Tata Usaha

- Persuratan
  - Dashboard TU `/tu`
  - Surat Masuk `/tu/surat-masuk`
  - Surat Keluar `/tu/surat-keluar`
  - Surat Keterangan `/tu/surat-keterangan`
  - Disposisi `/tu/disposisi`
- Arsip & Kepatuhan
  - Arsip `/tu/arsip`
  - Paket Kepatuhan `/tu/compliance-pack`
  - Monitoring Dokumen `/document-cycles`
  - Verifikasi Dokumen `/document-cycles/verifikasi`
- Tata Kelola
  - Dashboard Tata Kelola `/governance`
  - Tindak Lanjut `/governance/actions`
  - Kalender `/governance/actions/calendar`
  - Paket Rapat `/governance/actions/meeting-pack`
  - Briefing Owner `/governance/actions/owner-briefing`
  - Briefing SNP `/governance/actions/snp-briefing`
  - Briefing Bukti `/governance/actions/evidence-briefing`

### Aset & Layanan

- Perpustakaan
  - Dashboard Perpustakaan `/library`
  - Katalog Buku `/library/books`
  - Peminjaman Buku `/library/loans`
- Inventaris
  - Dashboard Inventaris `/inventory`
  - Daftar Barang `/inventory/items`

### Website

- Kelola Konten
  - Dashboard Website `/website`
  - Berita `/website/posts`
  - Pengumuman `/website/announcements`
  - Halaman Publik `/website/pages`

### Pegawai & Kehadiran

- Master Pegawai
  - Master Pegawai `/employees`
- PUSAKA
  - Monitor PUSAKA `/pusaka`
  - Pegawai PUSAKA `/pusaka/employees`
  - Data Kehadiran `/pusaka/kehadiran`
  - Ringkasan Kehadiran `/pusaka/summary`
  - Laporan Telegram `/pusaka/telegram-laporan`
  - Antrian Sinkronisasi `/pusaka/antrian`

### Pengaturan

- Akun & Profil Madrasah
  - Akun Saya `/settings/account`
  - Profil Madrasah `/settings/school-profile`
  - Logo & Branding `/settings/branding`
- Pengguna & Hak Akses
  - Pengguna `/settings/users`
  - Peran & Izin Akses `/settings/rbac`
  - Perubahan Data Pengguna `/settings/user-change-requests`
- Sistem & Audit
  - Pengaturan Sistem `/settings`
  - Backup & Restore `/settings/backups`
  - Maintenance Center `/settings/maintenance`
  - Audit Aktivitas `/settings/audit-logs`
  - Statistik Penggunaan `/settings/analytics`

## TDD Tasks

1. Baseline status and read current sidebar tests.
2. Add tests for recursive flattening and route coverage.
3. Add recursive model/helper implementation.
4. Add recursive RBAC filtering tests and implementation.
5. Update active/command/quick-access tests for leaf-only + breadcrumb.
6. Update `Sidebar.svelte` derived data to use flattened leaves.
7. Update `SidebarNavSection.svelte` nested rendering.
8. Convert config to 3-level IA and include missing static route leaves.
9. Update RBAC preview if it uses flat config.
10. Validate targeted tests, full unit if practical, `npm run check`, clean `npm run build`.
11. Commit focused diff.
12. Restart `mtsn2kolut-web-admin` and smoke protected routes (302 to login expected unauthenticated).

## Verification Commands

```bash
cd apps/web-admin
npm run test:unit -- src/lib/components/sidebar
npm run check
rm -rf build && npm run build
```

If full test suite is too slow, at least run all sidebar tests plus check/build. No backend tests required unless backend files change.

## Deploy

```bash
pm2 restart mtsn2kolut-web-admin
pm2 status --no-color | grep mtsn2kolut-web-admin
curl -I -fsS http://127.0.0.1:8021/ | head -5
curl -I -fsS http://127.0.0.1:8021/bank-soal | head -5
curl -I -fsS http://127.0.0.1:8021/asesmen | head -5
```

Expected protected routes may return `302` to login.
