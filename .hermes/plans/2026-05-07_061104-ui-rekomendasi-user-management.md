# Rekomendasi UI & Plan — User Management MTsN 2 Kolaka Utara

Tanggal: 2026-05-07

## Goal

Menyusun arah UI lanjutan untuk modul User Management setelah Tahap 1–Final selesai: Akun Saya, avatar, request perubahan data resmi, approval admin, indikator pending, dan export CSV.

## Rekomendasi Mode UI

### Mode yang disarankan: **Mode Hybrid Dashboard + Workflow**

Gunakan kombinasi:

1. **Dashboard ringkas** untuk admin/kepala/TU melihat status cepat.
2. **Workflow task-based** untuk admin menyelesaikan permintaan perubahan data.
3. **Self-service wizard sederhana** untuk guru/siswa/orang tua mengajukan perubahan data resmi.

Alasannya:

- Modul ini bukan sekadar CRUD user, tetapi alur kerja layanan internal madrasah.
- Admin perlu melihat prioritas: pending, aging request, approved/rejected hari ini.
- User biasa butuh alur yang tidak membingungkan: lihat profil → ajukan perubahan → pantau status.
- Data resmi harus tetap terkendali oleh approval, sesuai Opsi B.

## Rekomendasi Visual & UX

### 1. Akun Saya: jadikan **Profile Center**

Halaman `/settings/account` sebaiknya dibagi menjadi tab/section:

- **Ringkasan Profil**
  - foto/avatar
  - nama tampil
  - username
  - tipe akun
  - status akun
- **Data Resmi**
  - NIP/NISN/NIK atau identitas terkait
  - nama resmi
  - tanggal lahir
  - field terkunci dengan label `Perlu persetujuan admin`
- **Kontak Saya**
  - nomor HP
  - email jika tersedia
  - alamat
- **Permintaan Perubahan**
  - form ajukan perubahan
  - daftar status pending/approved/rejected/cancelled
- **Keamanan**
  - ganti password
  - sesi aktif
- **Riwayat Aktivitas**
  - perubahan kontak
  - avatar
  - request perubahan data

### 2. Admin Review: jadikan **Inbox Persetujuan**

Halaman `/settings/user-change-requests` sebaiknya terasa seperti inbox kerja:

- Header KPI:
  - Pending
  - Disetujui bulan ini
  - Ditolak bulan ini
  - Rata-rata waktu review
- Filter chips cepat:
  - Semua
  - Pending
  - Pegawai
  - Siswa
  - Orang Tua
  - Nama
  - Tanggal lahir
- List kiri / detail kanan untuk desktop.
- Mobile tetap card list + drawer/detail page.
- Tombol utama:
  - `Setujui`
  - `Tolak dengan catatan`
  - `Salin ringkasan`
  - `Export CSV`

### 3. Settings Page: tambahkan **Card Manajemen Data Diri**

Di `/settings`, tampilkan card:

- Akun Saya
- Manajemen Pengguna
- Permintaan Perubahan Data
- Role & Permission
- Audit Logs

Untuk reviewer/admin, card `Permintaan Perubahan Data` menampilkan badge pending.

### 4. Sidebar Badge

Tetap gunakan badge kecil, jangan terlalu ramai.

- Tampilkan hanya jika pending > 0.
- Warna: amber/orange untuk pending.
- Jangan tampilkan untuk user tanpa permission `profile_changes.review`.

### 5. Bahasa UI

Gunakan istilah madrasah yang mudah dipahami:

- `Permintaan Perubahan Data`
- `Data Resmi`
- `Menunggu Persetujuan`
- `Disetujui`
- `Ditolak`
- `Dibatalkan`
- `Catatan Admin`
- `Riwayat Perubahan`

Hindari istilah teknis seperti `profile change request`, `field key`, `reviewer permission` pada UI user.

## Plan Implementasi Lanjutan

### Sprint UI 1 — Profile Center Polish

Tujuan: membuat `/settings/account` lebih rapi dan mudah dipakai.

File kemungkinan berubah:

- `apps/web-admin/src/routes/settings/account/+page.svelte`
- `apps/web-admin/src/lib/client/account.ts`
- Komponen baru opsional:
  - `apps/web-admin/src/routes/settings/account/_components/ProfileSummaryCard.svelte`
  - `apps/web-admin/src/routes/settings/account/_components/OfficialDataCard.svelte`
  - `apps/web-admin/src/routes/settings/account/_components/ContactEditorCard.svelte`
  - `apps/web-admin/src/routes/settings/account/_components/ProfileChangeRequestPanel.svelte`
  - `apps/web-admin/src/routes/settings/account/_components/AccountSecurityPanel.svelte`
  - `apps/web-admin/src/routes/settings/account/_components/ProfileHistoryTimeline.svelte`

Langkah:

1. Pecah halaman akun menjadi komponen kecil.
2. Tambahkan tab/section responsif.
3. Tambahkan state kosong untuk user tanpa profile link.
4. Perjelas copywriting field resmi terkunci.
5. Tambahkan skeleton/loading state yang konsisten.
6. Tambahkan test bila helper/type berubah.

Validasi:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
```

### Sprint UI 2 — Admin Review Inbox

Tujuan: membuat halaman review admin lebih cepat untuk kerja harian.

File kemungkinan berubah:

- `apps/web-admin/src/routes/settings/user-change-requests/+page.svelte`
- `apps/web-admin/src/lib/client/account.ts`
- Komponen baru opsional:
  - `apps/web-admin/src/routes/settings/user-change-requests/_components/RequestKpiCards.svelte`
  - `apps/web-admin/src/routes/settings/user-change-requests/_components/RequestFilterBar.svelte`
  - `apps/web-admin/src/routes/settings/user-change-requests/_components/RequestReviewCard.svelte`
  - `apps/web-admin/src/routes/settings/user-change-requests/_components/RequestDetailPanel.svelte`
  - `apps/web-admin/src/routes/settings/user-change-requests/_components/ReviewActionDialog.svelte`

Langkah:

1. Tambahkan KPI ringkas di bagian atas.
2. Ubah list menjadi inbox-style cards.
3. Tambahkan detail panel untuk desktop.
4. Tambahkan dialog approve/reject dengan konfirmasi jelas.
5. Tambahkan filter chips cepat.
6. Pastikan export CSV memakai filter aktif.

Validasi:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
```

Smoke manual:

- Login admin.
- Buka `/settings/user-change-requests`.
- Filter pending.
- Search nama.
- Approve/reject request dummy.
- Export CSV.

### Sprint UI 3 — Settings IA & Notification Polish

Tujuan: settings lebih mudah dinavigasi dan badge pending lebih informatif.

File kemungkinan berubah:

- `apps/web-admin/src/routes/settings/+page.svelte`
- `apps/web-admin/src/lib/components/Sidebar.svelte`
- `apps/web-admin/src/lib/components/sidebar/sidebar-attention.ts`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Test sidebar terkait.

Langkah:

1. Susun settings card berdasarkan kategori:
   - Akun & Profil
   - User & Permission
   - Sistem & Audit
2. Badge pending hanya tampil pada reviewer.
3. Tambahkan tooltip/copy singkat: `Ada X permintaan data menunggu review`.
4. Pastikan akses user biasa tetap bersih.

Validasi:

```bash
npm --prefix apps/web-admin run test:unit -- sidebar
npm --prefix apps/web-admin run check
```

### Sprint UI 4 — Design System Stabilization

Tujuan: konsistensi visual seluruh modul admin.

Langkah:

1. Standardisasi card header/subtitle/action.
2. Standardisasi badge status:
   - pending: amber
   - approved: green
   - rejected: red
   - cancelled: slate/gray
3. Standardisasi empty state.
4. Standardisasi confirm dialog untuk aksi sensitif.
5. Gunakan loading button untuk aksi mutasi.
6. Gunakan AsyncContent/boundary pattern untuk fetch besar.

File kemungkinan berubah:

- Shared UI components bila sudah ada.
- Halaman account dan user-change-requests.

Validasi:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit
npm --prefix apps/web-admin run build
```

## Rekomendasi Prioritas

Urutan terbaik:

1. **Sprint UI 1 — Profile Center Polish**
2. **Sprint UI 2 — Admin Review Inbox**
3. **Sprint UI 3 — Settings IA & Notification Polish**
4. **Sprint UI 4 — Design System Stabilization**

Jika ingin cepat terlihat hasilnya di pengguna, mulai dari **Sprint UI 1**.
Jika ingin cepat membantu admin, mulai dari **Sprint UI 2**.

## Risiko & Catatan

- Jangan ubah flow approval backend kecuali perlu.
- Jangan tampilkan data sensitif penuh di export atau badge.
- Jangan tambah frontend-only permission guard tanpa backend guard.
- Jangan membuat user biasa melihat menu review admin.
- Hindari redesign besar satu file tanpa memecah komponen; halaman akan sulit dirawat.

## Open Questions

1. Apakah halaman Akun Saya ingin dibuat model **tab** atau **single-scroll cards**?
2. Untuk admin review, apakah lebih cocok **inbox split view** atau **table + detail dialog**?
3. Apakah export CSV perlu tombol khusus untuk `Pending saja`, `Bulan ini`, dan `Semua filter aktif`?

## Saran Final

Saya sarankan gunakan:

- **Akun Saya:** single-scroll cards dulu, lalu tab jika konten makin panjang.
- **Admin Review:** inbox split view untuk desktop + card/detail untuk mobile.
- **Settings:** card grid dengan badge pending.
- **Status:** badge warna konsisten.
- **Action:** dialog konfirmasi untuk approve/reject/cancel.

Ini paling cocok untuk madrasah karena mudah dipahami operator, aman untuk data resmi, dan tetap nyaman di HP.
