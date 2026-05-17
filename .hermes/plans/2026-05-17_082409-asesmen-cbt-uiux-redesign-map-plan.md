# Asesmen CBT UI/UX Redesign Mapping & Plan

> **For Hermes:** Plan mode only. Do not implement this plan until user explicitly approves. Use subagent-driven-development skill task-by-task after approval.

**Goal:** Merapikan ulang UI/UX modul Asesmen CBT dan Bank Soal dengan memetakan struktur existing dari sidebar sampai submodul/fitur, lalu menyusun redesign bertahap yang lebih rapi, operasional, dan mudah dipakai panitia/guru/pengawas.

**Architecture:** SvelteKit tetap sebagai BFF/proxy dan UI layer; tidak ada akses DB langsung dari frontend. Redesign diprioritaskan di navigasi, information architecture, reusable page shell, dan konsistensi istilah sebelum mengubah behavior backend. Semua perubahan nantinya harus additive, backward-compatible, dan diuji dengan `svelte-check`, route contract tests, Go tests/build bila backend disentuh.

**Tech Stack:** SvelteKit/Svelte 5, shadcn-style UI components, Go core-api, PostgreSQL/sqlc, PM2 deploy.

---

## 1. Mode Plan — Batasan Saat Ini

Permintaan user: **"Mari masuk mode plan tanpa edit dulu"**.

Yang dilakukan pada sesi ini:
- Audit read-only struktur sidebar dan route existing.
- Membuat dokumen plan ini di `.hermes/plans/`.

Yang tidak dilakukan:
- Tidak mengubah UI/page/component production.
- Tidak menjalankan deploy/restart.
- Tidak mengubah API/backend.
- Tidak commit.

Catatan: file plan ini adalah satu-satunya file yang dibuat sesuai aturan plan mode.

---

## 2. Peta Sidebar Existing

Sumber: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

### 2.1 Beranda

- `/notifications` — Notifikasi
- `/akademik/kesiapan` — Kesiapan Akademik & Rapor

Masalah IA:
- `Kesiapan Akademik & Rapor` juga bagian Akademik/Rapor, tetapi dimunculkan di Beranda.
- Beranda belum menjadi landing operasional lintas modul yang jelas.

### 2.2 Portal

- `/portal/siswa` — Portal Siswa
- `/portal/orang-tua` — Portal Orang Tua

Masalah IA:
- Aman; role-specific.

### 2.3 Akademik

- `/akademik` — Ringkasan Akademik
- `/akademik/tahun-ajaran` — Tahun Ajaran
- `/akademik/kurikulum` — Struktur Kurikulum
- `/akademik/rombel` — Rombel
- `/akademik/mapel` — Mapel
- `/akademik/guru-mapel` — Guru Mapel
- `/akademik/jam-pelajaran` — Jam Pelajaran
- `/akademik/jadwal` — Jadwal
- `/akademik/beban-guru` — Beban Guru
- `/journal` — Jurnal Kelas

Masalah IA:
- Akademik sudah padat, tetapi relatif terstruktur.
- Jurnal Kelas berada di Akademik, tetapi path-nya `/journal`, tidak konsisten dengan prefix `/akademik`.

### 2.4 Siswa & Orang Tua

- `/students` — Siswa
- `/parents` — Orang Tua
- `/kesiswaan` — Kesiswaan

Masalah IA:
- Bahasa path campur Indonesia/English (`students`, `parents`).
- Fitur siswa/orangtua berkaitan dengan Akademik dan Portal; perlu cross-link yang rapi.

### 2.5 Nilai & Rapor

- `/grades` — Input Nilai
- `/grades/rapor` — Rapor Siswa

Masalah IA:
- Path English, label Indonesia.
- Belum jelas hubungan dengan Asesmen CBT/non-tes sebagai sumber nilai.

### 2.6 Bank Soal

- `/bank-soal` — Dashboard Bank Soal
- `/bank-soal/daftar` — Daftar Soal
- `/bank-soal/tambah` — Tambah Soal
- `/bank-soal/verifikasi` — Verifikasi Soal
- `/bank-soal/penerbitan` — Penerbitan Soal
- `/bank-soal/impor` — Impor Soal
- `/bank-soal/analisis-butir` — Analisis Butir
- `/bank-soal/mapel-kd` — Mapel & KD
- `/bank-soal/pengaturan` — Pengaturan Bank Soal

Masalah IA:
- Banyak submenu flat di sidebar.
- Ada alur berbeda yang dicampur: authoring, review, publish, import, analytics, settings.
- `Analisis Butir` kurang tepat jika belum berbasis metrik item-analysis lengkap; bisa jadi `Mutu Soal`/`Kesehatan Bank Soal`.

### 2.7 Asesmen CBT

- `/asesmen` — Beranda Asesmen
- `/asesmen/paket` — Paket Soal
- `/asesmen/kegiatan` — Kegiatan Asesmen
- `/asesmen/persiapan` — Persiapan Asesmen
- `/asesmen/aplikasi-siswa/release` — Aplikasi Siswa CBT
- `/asesmen/pelaksanaan` — Pelaksanaan Ujian
- `/asesmen/hasil` — Hasil Asesmen

Masalah IA utama:
- Sidebar hanya menampilkan 7 item, tetapi route aktual Asesmen jauh lebih banyak.
- `Paket Soal`, `Kegiatan`, `Persiapan`, `Pelaksanaan`, `Hasil` adalah tahapan workflow, tetapi UI terasa seperti halaman terpisah, bukan satu alur terpandu.
- Fitur penting seperti sesi, pengawasan ruang, berita acara, arsip, kartu ujian tersembunyi di route detail, sehingga user merasa “kurang rapi”.
- `Aplikasi Siswa CBT` hanya menunjuk release, padahal ada panduan dan matrix device.

### 2.8 Tata Usaha

- `/tu`, `/tu/surat-masuk`, `/tu/surat-keluar`, `/tu/surat-keterangan`, `/tu/disposisi`, `/tu/arsip`, `/tu/compliance-pack`
- `/document-cycles`, `/document-cycles/verifikasi`
- `/governance/actions`, `/governance`

Masalah IA:
- Banyak domain berbeda digabung; bukan fokus plan ini.

### 2.9 Aset & Layanan

- `/library`, `/library/books`, `/library/loans`
- `/inventory`, `/inventory/items`

### 2.10 Website

- `/website`, `/website/posts`, `/website/announcements`, `/website/pages`

### 2.11 Pegawai & Kehadiran

- `/employees`
- `/pusaka`, `/pusaka/employees`, `/pusaka/kehadiran`, `/pusaka/summary`, `/pusaka/telegram-laporan`, `/pusaka/antrian`

### 2.12 Pengaturan

- `/settings/account`, `/settings/school-profile`, `/settings/branding`, `/settings/users`, `/settings/rbac`, `/settings/user-change-requests`, `/settings/backups`, `/settings/maintenance`, `/settings/audit-logs`, `/settings/analytics`, `/settings`

---

## 3. Peta Route Existing — Asesmen CBT

Sumber audit: `apps/web-admin/src/routes/asesmen/**/+page.svelte`

### 3.1 Landing & fase besar

- `/asesmen`
  - Landing Asesmen: tugas hari ini, alur CBT sederhana, pintasan sesuai peran.
- `/asesmen/persiapan`
  - Fase persiapan, prinsip persiapan, akses ke kegiatan/paket/aplikasi.
- `/asesmen/pelaksanaan`
  - Fase pelaksanaan, ritme operator, akses pengawasan dan aplikasi BYOD.
- `/asesmen/hasil`
  - Fase hasil, pilihan hasil CBT.

Masalah UX:
- Fase besar sudah ada, tetapi tiap halaman terasa sebagai “dashboard mini” sendiri.
- Tidak ada layout konsisten berupa stepper/kanban workflow lintas halaman.
- User harus ingat harus masuk halaman mana.

### 3.2 Kegiatan Asesmen

- `/asesmen/kegiatan`
  - List kegiatan dan sesi ujian.
  - Filter status kegiatan.
  - Buat kegiatan.
- `/asesmen/kegiatan/new`
  - Buat kegiatan asesmen.
  - Identitas kegiatan.
  - Persiapan CBT.
- `/asesmen/kegiatan/[id]`
  - Detail kegiatan.
  - Langkah berikutnya.
  - Kesiapan kegiatan.
  - Timeline SOP kegiatan.
  - Pengesahan SOP.
  - Kelengkapan data.
  - Kelengkapan soal per guru/mapel/rombel.
  - Hasil & analisis.
- `/asesmen/kegiatan/[id]/members`
  - Daftar penugasan event/kegiatan.
  - Kegiatan ujian.
  - Buka komposer Bank Soal.
- `/asesmen/kegiatan/[id]/exam-cards`
  - Kartu ujian kegiatan.
  - Cek sesi event, penugasan/peserta, sesi dan ruang.
- `/asesmen/kegiatan/[id]/archive`
  - Arsip kegiatan asesmen.
  - Checklist arsip.
  - Status pengesahan.
  - Sesi dan dokumen BA.

Masalah UX:
- Terlalu banyak fungsi penting di detail kegiatan dalam satu halaman panjang.
- Pengesahan SOP + readiness + hasil + kelengkapan soal bercampur; perlu grouping tab yang lebih rapi.
- Route `archive`, `members`, `exam-cards` belum terasa sebagai bagian natural dari detail kegiatan karena sidebar/secondary nav belum konsisten.
- Label “Event” masih muncul pada beberapa halaman seperti kartu ujian (`Kembali ke Event`, `Cek sesi event`) dan perlu dibersihkan.

### 3.3 Paket Soal

- `/asesmen/paket`
  - Daftar paket, buat paket, relasi ke kegiatan.
- `/asesmen/paket/new`
  - Builder paket baru.
  - Preview draw.
- `/asesmen/paket/[id]`
  - Metadata paket.
  - Soal dalam paket.
  - Tambah dari Bank Soal.
  - Blueprint & mutu.
  - Lock/snapshot/revisi.

Masalah UX:
- Paket sebagai objek teknis penting tetapi perlu dikaitkan jelas dengan “Kegiatan” dan “Bank Soal”.
- Perlu state label konsisten: Draft, Siap Review, Siap Digunakan, Terkunci, Revisi.

### 3.4 Sesi dan Ruang

- `/asesmen/sesi`
  - List sesi ujian.
  - Buat sesi baru.
  - Daftarkan siswa ke sesi.
  - Ubah jadwal sesi.
- `/asesmen/sesi/new`
  - Buat sesi ujian baru.
- `/asesmen/sesi/[id]`
  - Monitoring sesi.
  - Daftar nilai.
  - Akurasi per tipe soal.
  - Matriks analisis butir.
  - Ruangan & pengawas.
  - Peserta per ruangan.
  - Rekap ruang & register insiden.
- `/asesmen/sesi/[id]/minutes`
  - Berita acara sesi ujian.
  - Token peserta sudah dimasking.
- `/asesmen/sesi/[id]/proctoring`
  - Panel pengawasan lintas ruang.
  - Peserta, terkunci, terputus, warning, ringkasan ruang.
- `/asesmen/sesi/[id]/proctoring/report`
  - Rekap insiden & BA CBT lintas sesi.
- `/asesmen/sesi/[id]/rooms/[rid]/print-pack`
  - Paket pengawas ruang CBT.
- `/asesmen/sesi/[id]/rooms/[rid]/proctoring`
  - Panel pengawas ruang.
  - Mode Sederhana/Mode Lengkap.
  - Filter peserta dan tindakan pengawas.
- `/asesmen/sesi/[id]/rooms/[rid]/proctoring/report`
  - Rekap insiden ruang dan BA ruang.

Masalah UX:
- Sesi detail menjadi pusat terlalu banyak hal: monitoring, nilai, analisis, ruang, peserta, insiden.
- Pengawas membutuhkan tampilan sangat simpel saat ujian, sedangkan admin/operator butuh kontrol lengkap. Mode sudah mulai ada, tetapi perlu layout lebih bersih.
- Perlu membedakan jelas “Sesi”, “Ruang”, dan “Peserta” secara visual.

### 3.5 Aplikasi Siswa CBT/BYOD

- `/asesmen/aplikasi-siswa`
  - Panduan aplikasi siswa BYOD.
  - Arti status koneksi.
  - Checklist sebelum kirim.
- `/asesmen/aplikasi-siswa/release`
  - Daftar pemeriksaan rilis APK.
  - Unduh APK terbaru.
  - Pemeriksaan layanan/APK/rilis.
- `/asesmen/aplikasi-siswa/matrix`
  - Perbandingan perangkat BYOD.
  - Contoh tabel perangkat.
  - Fokus uji minimal.

Masalah UX:
- Sidebar hanya langsung ke release, padahal ada landing dan matrix.
- Perlu satu “Aplikasi Siswa” hub dengan tab: Panduan, Release, Matrix Perangkat, Troubleshooting.

### 3.6 Non-Tes

- `/asesmen/non-tes`
  - Daftar asesmen non-tes.

Masalah UX:
- Tidak muncul di sidebar Asesmen CBT; domainnya “Asesmen” tapi bukan CBT. Perlu dipisah: CBT vs Non-Tes.

---

## 4. Peta Route Existing — Bank Soal

Sumber audit: `apps/web-admin/src/routes/bank-soal/**/+page.svelte`

### 4.1 Dashboard dan daftar

- `/bank-soal`
  - Dashboard Bank Soal; memakai komponen workspace/list/health.
- `/bank-soal/daftar`
  - Daftar Soal; kemungkinan memakai `BankSoalListPage.svelte`.
- `/bank-soal/tambah`
  - Tambah soal; kemungkinan memakai composer/workspace.
- `/bank-soal/soal/[id]`
  - Detail/editor soal.

Masalah UX:
- Dashboard/list/editor punya banyak panel dan istilah; perlu page shell konsisten.
- `Tambah Soal` sebagai item sidebar mungkin terlalu rendah level; bisa jadi primary action di dashboard/list.

### 4.2 Workflow mutu

- `/bank-soal/verifikasi`
  - Verifikasi soal.
- `/bank-soal/penerbitan`
  - Penerbitan soal.
- `/bank-soal/analisis-butir`
  - Saat audit, heading masih “Analisis Butir”.

Masalah UX:
- Verifikasi dan penerbitan adalah queue workflow, cocok jadi tab dalam “Review & Terbitkan”.
- `Analisis Butir` sebaiknya diganti bila belum benar-benar item analysis berbasis response siswa; alternatif: “Mutu Soal”, “Kesehatan Bank Soal”, atau “Pemantauan Mutu Soal”.

### 4.3 Import dan konfigurasi

- `/bank-soal/impor`
  - Impor soal.
- `/bank-soal/mapel-kd`
  - Mapel & KD Coverage.
- `/bank-soal/pengaturan`
  - Pengaturan & scope reviewer.
  - Alur standar.
  - Standar kualitas minimum.
  - Integrasi modul.

Masalah UX:
- `Mapel & KD`, `Pengaturan`, `Impor` adalah admin/config/supporting actions, sebaiknya tidak sejajar dengan workflow harian guru.
- Perlu role-based simplification: guru melihat “Tulis Soal”, “Soal Saya”, “Perlu Revisi”; admin melihat “Review”, “Terbitkan”, “Pengaturan”.

---

## 5. Diagnosis Kenapa UI Terasa Kurang Rapi

### 5.1 Sidebar terlalu datar untuk workflow yang panjang

Bank Soal dan Asesmen CBT adalah workflow multi-tahap, tetapi sidebar menampilkan daftar halaman flat. Akibatnya:
- User tidak tahu urutan kerja.
- Fitur operasional tersembunyi di route detail.
- Ada banyak “dashboard mini” yang tumpang tindih.

### 5.2 Detail page terlalu padat

Contoh:
- Detail Kegiatan memuat readiness, SOP, kelengkapan soal, hasil, analisis, tindakan berikutnya.
- Detail Sesi memuat monitoring, nilai, analisis, ruang, peserta, register insiden.

Perlu pola: **Header ringkas + tab/section nav + action rail**.

### 5.3 Istilah belum 100% konsisten

Masih ditemukan/berpotensi muncul:
- Event → seharusnya Kegiatan Asesmen.
- Analisis Butir → jangan dipakai jika belum item analysis sejati.
- Kode Ujian/Kode Ruang → Token Ujian/Token Ruang.

### 5.4 Mode peran belum cukup jelas

Operator, guru, proktor, admin, siswa/orangtua membutuhkan fokus berbeda:
- Guru: tulis soal, revisi, lihat soal saya.
- Admin/panitia: kegiatan, paket, peserta, sesi, arsip.
- Pengawas: token ruang, peserta bermasalah, tindakan cepat.
- Kepala/pimpinan: ringkasan readiness/hasil.

### 5.5 Belum ada desain sistem khusus “operasional CBT”

Komponen seperti status card, readiness step, SOP approval, proctor action, token display, empty state, dan page header perlu distandardkan agar halaman terlihat satu keluarga.

---

## 6. Rekomendasi Information Architecture Baru

### 6.1 Sidebar level atas yang disarankan

#### Bank Soal

Ubah dari banyak item flat menjadi 4–5 item utama:

1. **Dashboard Bank Soal** — `/bank-soal`
2. **Kelola Soal** — `/bank-soal/daftar`
   - Primary action: Tambah Soal
   - Filter: Soal Saya, Draft, Revisi, Siap Review, Terbit
3. **Review & Terbitkan** — bisa tetap route existing `/bank-soal/verifikasi` dan `/bank-soal/penerbitan`, tetapi tampil sebagai tab satu shell.
4. **Mutu Soal** — rename dari Analisis Butir jika metrik belum sejati.
5. **Pengaturan** — berisi Mapel & KD, Import, Scope Reviewer, standar kualitas.

Item `Tambah Soal`, `Impor Soal`, `Mapel & KD` sebaiknya tidak wajib jadi sidebar utama; bisa jadi action/tab di halaman terkait.

#### Asesmen CBT

Ubah dari halaman flat menjadi 5 workflow utama:

1. **Dashboard Asesmen** — `/asesmen`
2. **Persiapan** — `/asesmen/persiapan`
   - Paket Soal
   - Kegiatan Asesmen
   - Peserta/Ruang/Token/Kartu
   - Aplikasi Siswa
3. **Pelaksanaan** — `/asesmen/pelaksanaan`
   - Sesi Hari Ini
   - Pengawasan Ruang
   - Insiden
4. **Hasil & Koreksi** — `/asesmen/hasil`
   - Nilai sesi
   - Jawaban essay/manual scoring bila ada
   - Export hasil
5. **Arsip & Berita Acara** — bisa landing baru atau masuk tab Kegiatan
   - BA sesi/ruang
   - Checklist arsip
   - Pengesahan final

`Non-Tes` sebaiknya menjadi item terpisah: **Asesmen Non-Tes**, bukan dicampur di CBT.

### 6.2 Secondary navigation per domain

Gunakan secondary nav/tab di dalam domain, bukan menambah semua route ke sidebar.

Contoh Detail Kegiatan:
- Ringkasan
- Paket & Soal
- Peserta & Ruang
- Token & Kartu
- Sesi
- Pengesahan SOP
- Hasil
- Arsip

Contoh Detail Sesi:
- Ringkasan
- Ruang
- Peserta
- Pengawasan
- Nilai
- Insiden/BA

Contoh Bank Soal:
- Semua Soal
- Draft Saya
- Perlu Revisi
- Review
- Terbit
- Mutu

---

## 7. Target Desain Baru

### 7.1 Pola halaman standar

Setiap halaman utama sebaiknya memakai struktur:

1. **PageHeader**
   - judul pendek
   - deskripsi 1 baris
   - status badge
   - 1–2 primary actions saja
2. **WorkflowNav / Stepper**
   - menunjukkan posisi user di alur
3. **Content cards**
   - maksimal 2 kolom desktop
   - satu tujuan per card
4. **Right action rail** untuk halaman detail besar
   - tindakan penting
   - readiness blockers
   - quick links
5. **Empty/Error/Loading state konsisten**

### 7.2 Visual hierarchy

- Hindari terlalu banyak card setara dalam satu viewport.
- Card utama: data yang harus diputuskan hari ini.
- Card sekunder: informasi pendukung.
- Gunakan badge warna konsisten:
  - Hijau: siap/aman/selesai
  - Kuning: perlu dicek
  - Merah: blocker/perlu tindakan
  - Biru: informasi/progress
  - Abu: draft/nonaktif

### 7.3 Copywriting

- Semua istilah CBT harus operasional dan Indonesia.
- Hindari istilah teknis developer: `workflow`, `event`, `heartbeat`, `fingerprint`, `anti-switch`.
- Gunakan:
  - Kegiatan Asesmen
  - Token Ruang
  - Token Ujian
  - Status Koneksi
  - Keluar Aplikasi
  - Coba Tangkap Layar
  - Penanda Perangkat
  - Sudah Kirim

---

## 8. Rekomendasi Tata Letak Dalam Modul

Prinsip penting: jangan hanya merapikan sidebar. Setiap modul perlu punya pola halaman dalam yang konsisten supaya user tidak merasa masuk ke “halaman berbeda-beda” setiap klik.

### 8.1 Pola Layout Global untuk Asesmen/Bank Soal

Gunakan pola 5 bagian di halaman utama dan detail:

1. **Header Operasional**
   - Judul singkat, contoh: `Kegiatan Asesmen`, `Detail Kegiatan`, `Panel Pengawas Ruang`.
   - Deskripsi 1 kalimat, bukan paragraf panjang.
   - Status utama: Draft, Siap, Berjalan, Selesai, Diarsipkan.
   - Primary action maksimal 1–2 tombol.

2. **Context Bar**
   - Tahun ajaran, semester, kegiatan aktif, tanggal, paket, sesi, ruang.
   - Untuk detail: breadcrumb ringan seperti `Asesmen > Kegiatan > UTS Genap`.
   - Untuk pengawas: tampilkan `Sesi`, `Ruang`, `Token Ruang`, `Jumlah Peserta`.

3. **Workflow Tabs / Section Nav**
   - Tab sesuai alur kerja, bukan berdasarkan teknis database.
   - Tab harus stabil di semua detail: Ringkasan, Data, Kesiapan, Pelaksanaan, Hasil, Arsip.

4. **Main Content + Right Rail**
   - Desktop: main content 70%, right rail 30%.
   - Mobile: right rail turun ke bawah sebagai accordion.
   - Main content berisi pekerjaan utama.
   - Right rail berisi blocker, tindakan berikutnya, quick links, status SOP.

5. **Bottom Sticky Action hanya saat perlu**
   - Dipakai untuk perubahan belum disimpan, approval, atau operasi massal.
   - Jangan semua halaman punya sticky bar agar tidak ramai.

---

### 8.2 Modul Asesmen CBT — Landing `/asesmen`

**Tujuan halaman:** bukan sekadar dashboard, tetapi “pintu masuk kerja hari ini”.

**Layout rekomendasi:**

- Header:
  - Judul: `Asesmen CBT`
  - Deskripsi: `Pantau persiapan, pelaksanaan, hasil, dan arsip kegiatan asesmen.`
  - Tombol utama: `Buat Kegiatan`, `Buka Pelaksanaan Hari Ini`

- Row 1 — Today Command Center:
  - Card `Kegiatan berjalan/hari ini`
  - Card `Sesi perlu pengawasan`
  - Card `Blocker persiapan`
  - Card `Hasil menunggu koreksi/arsip`

- Row 2 — Workflow besar:
  - Step card 1: `Persiapan`
  - Step card 2: `Pelaksanaan`
  - Step card 3: `Hasil & Koreksi`
  - Step card 4: `Arsip & BA`

- Row 3 — Quick action berdasarkan role:
  - Admin: buat kegiatan, cek paket, cek peserta, cetak kartu.
  - Guru: susun soal, lihat paket terkait, koreksi.
  - Pengawas: masuk panel ruang.

**Yang harus dihindari:**
- Terlalu banyak daftar link kecil.
- Semua fitur ditaruh sebagai card sama besar.

---

### 8.3 Modul Persiapan Asesmen `/asesmen/persiapan`

**Tujuan halaman:** checklist pra-ujian sampai siap jalan.

**Layout rekomendasi:**

- Header:
  - Judul: `Persiapan Asesmen`
  - Status: jumlah kegiatan belum siap.

- Main layout:
  - Kiri: **Preparation Checklist** sebagai stepper vertikal:
    1. Buat Kegiatan
    2. Pilih Paket Soal
    3. Lengkapi Peserta
    4. Buat Sesi & Ruang
    5. Cetak/Bagikan Kartu & Token
    6. Sahkan SOP Persiapan
  - Kanan: **Blocker Panel**:
    - paket kurang soal
    - peserta belum punya ruang
    - sesi belum punya pengawas
    - token belum siap

- Bagian bawah:
  - List kegiatan draft/akan datang dengan status readiness.

**Komponen ideal:**
- `ReadinessChecklist.svelte`
- `BlockerList.svelte`
- `OperationalStepCard.svelte`

---

### 8.4 Modul Kegiatan Asesmen List `/asesmen/kegiatan`

**Tujuan halaman:** memilih dan mengelola kegiatan, bukan menampilkan semua detail.

**Layout rekomendasi:**

- Header + primary action `Buat Kegiatan`.
- Filter bar ringkas:
  - Tahun ajaran
  - Semester
  - Status
  - Tingkat
  - Search
- Table/card hybrid:
  - Nama kegiatan
  - Status
  - Tanggal/sesi
  - Readiness score
  - Paket lengkap?
  - Peserta lengkap?
  - SOP status
  - Action: `Buka`, `Kartu`, `Arsip`

**Mode tampilan:**
- Desktop: table dengan sticky action column.
- Mobile: card list per kegiatan.

**Yang harus dipindah keluar:**
- Jangan tampilkan detail sesi panjang di list. Cukup ringkasan dan status.

---

### 8.5 Detail Kegiatan `/asesmen/kegiatan/[id]`

**Tujuan halaman:** menjadi hub resmi satu kegiatan asesmen.

**Layout rekomendasi paling penting:**

- Header:
  - Nama kegiatan
  - Status kegiatan
  - Tahun ajaran/semester
  - Tanggal mulai-selesai
  - Primary action sesuai status: `Lanjutkan Persiapan`, `Buka Pelaksanaan`, `Arsipkan`

- Context summary cards kecil:
  - Paket
  - Peserta
  - Sesi/Ruang
  - Token/Kartu
  - SOP
  - Hasil

- Tab utama:
  1. **Ringkasan**
     - readiness score
     - langkah berikutnya
     - blocker utama
  2. **Paket & Soal**
     - paket per mapel/rombel
     - kelengkapan soal
     - link ke Bank Soal/Paket
  3. **Peserta & Ruang**
     - peserta per rombel
     - ruang virtual/fisik
     - status kartu/token
  4. **Sesi**
     - daftar sesi
     - status jadwal
     - pengawas/ruang
  5. **Pengesahan SOP**
     - milestone approval ringkas
     - catatan audit
     - tombol sahkan/cabut
  6. **Hasil**
     - ringkasan submission
     - nilai/koreksi
     - export
  7. **Arsip**
     - BA, kartu, paket bukti, checklist final

- Right rail:
  - `Tindakan Berikutnya`
  - `Blocker`
  - `Dokumen Cepat`: Kartu Ujian, BA Sesi, Arsip
  - `Audit ringkas`

**Masalah yang diselesaikan:**
- Detail yang sekarang panjang dipecah menjadi tab operasional.
- Pengesahan SOP tidak lagi terasa “menumpuk” di tengah halaman.

---

### 8.6 Paket Soal `/asesmen/paket` dan `/asesmen/paket/[id]`

**Tujuan halaman:** memastikan paket siap dipakai kegiatan.

**List paket layout:**
- Header: `Paket Soal CBT`
- Filter: kegiatan, mapel, tingkat, status, guru.
- Table:
  - Paket
  - Kegiatan terkait
  - Mapel/rombel
  - Jumlah soal
  - Requirement
  - Status lock/snapshot
  - Action: `Detail`, `Isi Soal`, `Preview`

**Detail paket layout:**
- Header paket + status lock.
- Summary row:
  - total soal
  - PG/essay
  - target requirement
  - randomisasi
  - versi/snapshot
- Tab:
  1. Metadata
  2. Soal dalam paket
  3. Blueprint & Mutu
  4. Randomisasi
  5. Riwayat Revisi
- Right rail:
  - checklist siap digunakan
  - kegiatan yang memakai paket
  - risiko: soal kurang, tipe soal tidak sesuai

---

### 8.7 Sesi Ujian `/asesmen/sesi` dan `/asesmen/sesi/[id]`

**Tujuan halaman:** operator memahami sesi, ruang, peserta, status submit.

**List sesi layout:**
- Header: `Sesi Ujian`
- Filter: kegiatan, tanggal, status, ruang, pengawas.
- Group by tanggal/kegiatan.
- Row/table:
  - sesi
  - waktu
  - kegiatan
  - paket
  - peserta
  - ruang
  - status submit
  - action: `Pantau`, `BA`, `Detail`

**Detail sesi layout:**
- Header sesi: kegiatan, mapel/paket, waktu, status.
- Summary cards:
  - peserta
  - hadir/login
  - sedang ujian
  - terkunci/terputus
  - sudah kirim
- Tab:
  1. Ringkasan
  2. Ruang
  3. Peserta
  4. Pengawasan
  5. Nilai/Koreksi
  6. Insiden & BA
- Right rail:
  - status waktu
  - tindakan cepat
  - link BA/paket pengawas

---

### 8.8 Pengawasan Ruang `/asesmen/sesi/[id]/rooms/[rid]/proctoring`

**Tujuan halaman:** pengawas tidak bingung dan tidak melihat terlalu banyak kontrol.

**Default layout: Mode Sederhana**

- Header sticky:
  - Nama sesi
  - Ruang
  - Token Ruang
  - Jam
  - Status koneksi
- Summary besar:
  - Total peserta
  - Terhubung
  - Perlu perhatian
  - Sudah kirim
- Main list default: hanya peserta bermasalah/perlu perhatian.
- Toggle: `Mode Sederhana` / `Mode Lengkap`.

**Mode Sederhana actions:**
- `Periksa`
- `Kirim Peringatan`
- `Instruksi Masuk Ulang`
- `Buka Kunci`

**Mode Lengkap tambahan:**
- semua peserta
- event log
- action force submit jika diizinkan
- export bukti

**Right rail:**
- SOP pengawas 4 langkah:
  1. Cek identitas/perangkat
  2. Pantau peserta bermasalah
  3. Catat tindakan
  4. Kunci BA setelah selesai

**Yang harus dihindari:**
- Tabel besar sebagai tampilan default pengawas.
- Tombol destruktif tampil sejajar dengan tindakan biasa.

---

### 8.9 Berita Acara, Arsip, dan Print Pack

**Tujuan halaman:** dokumen resmi mudah dicetak/diarsip.

**Layout BA Sesi/Ruang:**
- Header dokumen resmi.
- Metadata: madrasah, kegiatan, sesi, ruang, pengawas.
- Rekap angka.
- Register insiden.
- Tanda tangan.
- Tombol: `Cetak`, `Unduh CSV`, `Kembali`.

**Layout Arsip Kegiatan:**
- Checklist final di atas.
- Status pengesahan SOP.
- Dokumen per sesi/ruang dalam accordion.
- Token peserta tetap masked.
- Action final: `Tandai Arsip Lengkap` jika backend tersedia nanti.

---

### 8.10 Aplikasi Siswa CBT

**Tujuan halaman:** panitia/guru tahu APK mana yang dipakai dan cara troubleshooting.

**Hub `/asesmen/aplikasi-siswa` layout:**
- Header: `Aplikasi Siswa CBT`
- Cards:
  - Release APK aktif
  - Panduan siswa
  - Checklist perangkat
  - Matrix uji BYOD
  - Troubleshooting
- Tab:
  1. Panduan
  2. Rilis APK
  3. Matrix Perangkat
  4. Troubleshooting

**Release page layout:**
- APK aktif paling atas.
- Versi, tanggal, checksum bila tersedia.
- Checklist rilis.
- Link unduh.

---

### 8.11 Bank Soal Dashboard `/bank-soal`

**Tujuan halaman:** guru/admin langsung tahu pekerjaan soal hari ini.

**Layout role-aware:**

- Untuk guru:
  - Draft saya
  - Perlu revisi
  - Siap review
  - Terbit
  - Tombol utama: `Tambah Soal`

- Untuk admin/reviewer:
  - Menunggu review
  - Perlu penerbitan
  - Soal bermasalah
  - Coverage mapel/KD
  - Tombol utama: `Review Soal`

**Content:**
- Row 1: KPI workflow.
- Row 2: antrean pekerjaan.
- Row 3: mutu/coverage.

---

### 8.12 Kelola Soal `/bank-soal/daftar`, `/bank-soal/tambah`, `/bank-soal/soal/[id]`

**Daftar soal layout:**
- Filter left/top:
  - mapel
  - tingkat
  - status
  - pembuat
  - tipe soal
  - keyword
- Table/list:
  - kode/preview singkat soal
  - mapel/KD
  - tingkat
  - tipe
  - status
  - pemilik
  - aksi
- Bulk action untuk admin/reviewer jika aman.

**Editor soal layout:**
- Header: status soal + actions.
- 3-column desktop:
  - kiri: metadata/context
  - tengah: editor soal dan jawaban
  - kanan: quality checklist, version, preview
- Mobile: stepper:
  1. Metadata
  2. Soal
  3. Jawaban
  4. Pembahasan
  5. Review

---

### 8.13 Review & Penerbitan Bank Soal

**Tujuan halaman:** reviewer/admin bekerja dari queue, bukan mencari manual.

**Layout:**
- Tabs:
  - Menunggu Review
  - Perlu Revisi
  - Siap Terbit
  - Terbit
- Queue table:
  - soal
  - mapel/tingkat
  - guru
  - umur antrean
  - isu mutu
  - action
- Detail drawer:
  - preview soal
  - checklist review
  - catatan reviewer
  - approve/reject/request revision

---

### 8.14 Mutu Soal / Analisis Butir

**Tujuan halaman:** memantau kesehatan bank soal, bukan sekadar grafik.

**Jika belum ada data hasil siswa lengkap:**
- Rename label UI menjadi `Mutu Soal`.
- Isi dengan:
  - kelengkapan metadata
  - distribusi tipe soal
  - coverage mapel/KD
  - soal belum direview
  - soal terlalu pendek/panjang

**Jika nanti item analysis lengkap tersedia:**
- Tab tambahan `Analisis Butir`:
  - tingkat kesukaran
  - daya pembeda
  - distraktor
  - reliabilitas

---

### 8.15 Pengaturan Bank Soal

**Tujuan halaman:** admin mengatur standar, bukan guru harian.

**Layout:**
- Tabs:
  1. Scope Reviewer
  2. Mapel & KD
  3. Import
  4. Standar Mutu
  5. Integrasi Asesmen
- Jangan semua ditampilkan sebagai card panjang dalam satu halaman.

---

### 8.16 Rekomendasi Prioritas Desain Dalam Modul

Urutan terbaik menurut saya:

1. **Detail Kegiatan Asesmen** — paling terasa ramai dan pusat workflow.
2. **Pengawasan Ruang** — paling kritikal saat ujian berlangsung.
3. **Detail Sesi** — menghubungkan ruang, peserta, BA, hasil.
4. **Bank Soal Dashboard + Daftar Soal** — pekerjaan harian guru/admin.
5. **Review & Penerbitan** — workflow admin/reviewer.
6. **Arsip/BA** — dokumen resmi dan audit.

---

## 9. Swarm Review Decision — Opsi Final

Hasil 5-agent swarm review disimpan lengkap di:

`/home/servermtsn2kolut/mtsn2kolut-super-app/.hermes/plans/2026-05-17_0835-asesmen-cbt-uiux-swarm-synthesis.md`

### Konsensus Swarm

Plan awal benar arahnya, tetapi masih bisa lebih sederhana. Semua reviewer sepakat:

- Asesmen CBT harus mengikuti alur harian madrasah: **sebelum ujian → saat ujian → setelah ujian**.
- Detail Kegiatan jangan 7 tab; sederhanakan menjadi 5 area kerja.
- Right rail jangan permanen; tampilkan hanya jika membantu keputusan saat ini.
- Pengawasan ruang harus exception-first: tampilkan peserta/ruang bermasalah dulu.
- Mobile-first dan accessibility harus masuk sejak sprint pertama, bukan polish akhir.
- Bank Soal perlu role-aware, tetapi struktur navigasi jangan berubah total per role.

### Opsi yang Dipilih

**Opsi B — Operational Redesign Layer** direkomendasikan.

Artinya:

- Tidak rewrite backend.
- Tidak hapus route lama.
- UI dipaketkan ulang agar user melihat alur sederhana.
- Fitur teknis dipindah ke `Mode Lengkap`, detail page, drawer, atau `Fitur Lanjutan`.

### Struktur Asesmen CBT Final

1. **Hari Ini / Dashboard CBT**
2. **Persiapan Ujian**
3. **Monitor Ujian**
4. **Hasil & Berita Acara**
5. **Arsip**
6. **Aplikasi Siswa**

### Detail Kegiatan Final — 5 Area

1. **Ringkasan**
2. **Persiapan**
3. **Pelaksanaan**
4. **Hasil & BA**
5. **Arsip**

### Detail Sesi Final — 4 Area

1. **Monitor**
2. **Peserta**
3. **Masalah/Insiden**
4. **Hasil & BA**

### Bank Soal Final

1. **Dashboard Bank Soal**
2. **Kelola Soal**
3. **Review & Terbitkan**
4. **Mutu Soal**
5. **Pengaturan**

### Rule Anti-Ramai

- Satu layar = satu tujuan utama.
- Satu primary action per layar.
- Summary card maksimal 3 sebelum konten utama.
- Tabs maksimal 5.
- Right rail optional.
- Table untuk data banyak; card untuk ringkasan/workflow.
- Drawer untuk detail sekunder.
- Fitur teknis masuk `Mode Lengkap` atau `Fitur Lanjutan`.
- Mobile proctoring harus card-first.

---

## 10. Rencana Implementasi Bertahap — Operational Redesign Layer

### Sprint Redesign 0 — Audit Visual dan Sitemap Final

**Objective:** Membuat audit final route/sidebar dan sitemap baru sebelum edit UI.

**Files likely read only:**
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- `apps/web-admin/src/routes/asesmen/**/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/**/+page.svelte`

**Deliverables:**
- Dokumen sitemap final.
- Daftar label yang harus diganti.
- Screenshot/current-state evidence jika perlu via browser.

**Verification:**
- Tidak ada file UI diubah.

### Sprint Redesign 1 — Inner Layout Blueprint + Shared Shell

**Objective:** Rapikan dulu tata letak dalam halaman agar modul terasa rapi walaupun sidebar belum berubah besar.

**Files likely created:**
- `apps/web-admin/src/lib/components/asesmen/OperationalPageHeader.svelte`
- `apps/web-admin/src/lib/components/asesmen/ContextBar.svelte`
- `apps/web-admin/src/lib/components/asesmen/WorkflowNav.svelte`
- `apps/web-admin/src/lib/components/asesmen/ActionRail.svelte`
- `apps/web-admin/src/lib/components/asesmen/ReadinessChecklist.svelte`
- `apps/web-admin/src/lib/components/asesmen/BlockerList.svelte`
- `apps/web-admin/src/lib/components/asesmen/EmptyState.svelte`

**Apply first to shell/wrapper only:**
- `/asesmen`
- `/asesmen/persiapan`
- `/asesmen/pelaksanaan`
- `/asesmen/hasil`
- `/bank-soal`

**Changes:**
- Standarkan header, deskripsi, primary actions, context bar.
- Standarkan tab/section nav untuk halaman besar.
- Standarkan right rail untuk blocker/tindakan berikutnya.
- Jangan ubah backend.
- Jangan menghapus route lama.

**Validation:**
```bash
npm --prefix apps/web-admin run check
```

### Sprint Redesign 2 — Detail Kegiatan Redesign

**Objective:** Merapikan `/asesmen/kegiatan/[id]` sebagai hub kegiatan.

**Files likely changed:**
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/archive/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/exam-cards/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/members/+page.svelte`

**Design target:**
- Header kegiatan: nama, status, tanggal, readiness score.
- Tab nav: Ringkasan, Soal, Peserta & Ruang, Token & Kartu, Sesi, SOP, Hasil, Arsip.
- Action rail: tindakan berikutnya, blocker, quick links.
- SOP approval jadi panel ringkas, bukan card panjang yang memakan halaman.

**Validation:**
```bash
npm --prefix apps/web-admin run check
```

### Sprint Redesign 3 — Sesi & Proctoring Redesign

**Objective:** Membuat pengalaman pengawas lebih rapi dan tenang saat ujian.

**Files likely changed:**
- `apps/web-admin/src/routes/asesmen/sesi/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/print-pack/+page.svelte`

**Design target:**
- Sesi detail: ringkasan → ruang → peserta → nilai → insiden/BA.
- Proctoring ruang: default Mode Sederhana, action minim, fokus peserta bermasalah.
- Mode Lengkap tetap tersedia untuk admin/operator.
- Token tetap dimasking di konteks BA/arsip.

**Validation:**
```bash
npm --prefix apps/web-admin run check
```

### Sprint Redesign 4 — Bank Soal Redesign

**Objective:** Merapikan Bank Soal menjadi alur authoring-review-publish.

**Files likely changed:**
- `apps/web-admin/src/routes/bank-soal/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/PublishBankSoalPage.svelte`
- `apps/web-admin/src/routes/bank-soal/analisis-butir/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/pengaturan/+page.svelte`

**Design target:**
- Guru landing: Draft Saya, Perlu Revisi, Tambah Soal, Mutu Soal Saya.
- Admin landing: Review Queue, Publish Queue, Kelengkapan mapel/tingkat, scope reviewer.
- Rename/positioning `Analisis Butir` menjadi `Mutu Soal` jika metrik item analysis belum lengkap.
- `Tambah Soal` menjadi primary action, bukan wajib sidebar item.

**Validation:**
```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/cbt-backend-paths.test.ts
```

### Sprint Redesign 5 — Sidebar dan IA Cleanup

**Objective:** Setelah layout dalam rapi, ringkas sidebar agar mengikuti workflow baru tanpa menghapus route lama.

**Files likely changed:**
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- mungkin tests route access:
  - `apps/web-admin/src/lib/server/route-access.test.ts`
  - `apps/web-admin/src/lib/server/assessment-route-contract.test.ts`

**Changes:**
- Bank Soal: kurangi item flat, pindahkan beberapa menjadi in-page actions/tab.
- Asesmen CBT: jadikan workflow utama: Dashboard, Persiapan, Pelaksanaan, Hasil, Arsip/BA, Aplikasi Siswa.
- Tambahkan label yang lebih jelas dan konsisten.
- Pastikan role/permission tidak bocor.

**Validation:**
```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/assessment-route-contract.test.ts
```

### Sprint Redesign 6 — Responsive, Accessibility, and Visual Polish

**Objective:** Final polish agar desktop/tablet/mobile rapi.

**Checklist:**
- Keyboard focus visible.
- Buttons have `type="button"` where needed.
- Tab groups use role/aria-selected where applicable.
- Long tables get horizontal scroll with sticky key columns.
- Empty states have next action.
- Loading states not visually noisy.
- No token exposed in BA/archive by default.

**Validation:**
```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

### Sprint Redesign 7 — Production Rollout

**Objective:** Deploy only after user approval.

**Steps:**
1. Final `git diff --stat` review.
2. `npm --prefix apps/web-admin run check`.
3. Route tests.
4. Build web-admin.
5. No backend deploy unless backend changed.
6. Restart PM2 web-admin only if frontend-only.
7. Smoke test:
   - `/asesmen`
   - `/asesmen/persiapan`
   - `/asesmen/pelaksanaan`
   - `/bank-soal`
   - satu detail kegiatan existing bila tersedia.

---

## 10. Risks & Guardrails

### Risks

- Terlalu banyak perubahan UI sekaligus bisa membuat user bingung.
- Mengubah sidebar bisa memengaruhi role/permission visibility.
- Halaman Asesmen dan Bank Soal besar; refactor agresif dapat memicu Svelte type errors.
- Route lama harus tetap ada agar link existing tidak broken.

### Guardrails

- Jangan hapus route lama; redirect atau tetap accessible.
- Jangan ubah backend kecuali benar-benar perlu.
- Jangan expose token/credential.
- Setiap sprint kecil, validasi `svelte-check`.
- Jika deploy, restart web-admin segera setelah build agar manifest/chunk tidak mismatch.

---

## 11. Open Questions untuk User

1. Bapak ingin sidebar Asesmen CBT tetap 7 item seperti sekarang, atau diringkas menjadi workflow 5 item?
2. Untuk Bank Soal, apakah `Tambah Soal` tetap harus muncul di sidebar, atau cukup sebagai tombol utama di Dashboard/Daftar Soal?
3. Istilah `Analisis Butir` ingin diganti menjadi `Mutu Soal` dulu sampai metrik analisis butir benar-benar lengkap?
4. Untuk pengawas ruang, default halaman mau **Mode Sederhana** saja, dengan Mode Lengkap tersembunyi di toggle?
5. Apakah redesign fokus dulu pada Asesmen CBT, lalu Bank Soal; atau keduanya sekaligus?

---

## 12. Rekomendasi Urutan Eksekusi

Rekomendasi saya:

1. **Sprint Redesign 1:** Inner layout blueprint + shared shell.
2. **Sprint Redesign 2:** Detail Kegiatan.
3. **Sprint Redesign 3:** Sesi & Proctoring.
4. **Sprint Redesign 4:** Bank Soal.
5. **Sprint Redesign 5:** Sidebar/IA cleanup setelah layout dalam stabil.
6. **Sprint Redesign 6:** Polish.

Alasannya: Bapak benar, kalau yang dirapikan hanya sidebar tetapi isi halaman tetap ramai, hasilnya masih terasa kurang rapi. Karena itu plan disesuaikan: **mulai dari tata letak dalam modul dulu**, terutama Detail Kegiatan, Pengawasan Ruang, Detail Sesi, lalu Bank Soal. Sidebar baru diringkas setelah pola halaman dalamnya jelas.
