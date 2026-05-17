# Glosarium Istilah End-User

Tanggal: 2026-05-17

Dokumen ini menetapkan istilah yang boleh tampil kepada pengguna akhir di Web Admin, Portal/Mobile CBT, dan materi SOP. Tujuannya menjaga copy tetap mudah dipahami oleh guru, pengawas, siswa, operator madrasah, dan orang tua tanpa mengubah kontrak teknis internal.

## Batas Perubahan

- **Boleh diubah:** label menu, judul halaman, deskripsi, tombol, badge/status visible, empty/error/loading state, helper text, dan teks SOP.
- **Tidak boleh diubah hanya karena cleanup copy:** route path, API/BFF path, permission code, enum database, field payload, test selector internal, dan nama file/komponen teknis.
- Istilah teknis tetap boleh ada di kode/test kontrak selama tidak menjadi visible copy untuk pengguna akhir.
- Jika label visible berubah, test UI/sidebar boleh disesuaikan ekspektasinya, tetapi test harus tetap menjaga href/route contract yang sama.

## Prinsip Penamaan

1. Utamakan bahasa Indonesia yang biasa dipakai di madrasah.
2. Hindari istilah developer-facing seperti `endpoint`, `BFF`, `API`, `payload`, `IndexedDB`, dan `httpOnly` pada layar pengguna.
3. Jika singkatan sudah umum, boleh dipertahankan dengan helper singkat: CBT, HOTS, CP, TP, KD, PG.
4. Untuk pesan saat ujian, gunakan kalimat pendek: apa yang terjadi, apakah jawaban aman, dan tindakan berikutnya.
5. Satu konsep harus punya satu istilah utama di semua layar agar guru/siswa tidak bingung.

## Istilah Utama Asesmen CBT

- `Event`: **Kegiatan Asesmen**; pendek: **Kegiatan**.
- `Session`: **Sesi Ujian**; pendek: **Sesi**.
- `Room`: **Ruang Ujian**; pendek: **Ruang**.
- `Proctoring`: **Pengawasan Ujian** atau **Pengawasan Ruang**.
- `Command Center`: **Panel Pengawasan**; alternatif konteks besar: **Pusat Pemantauan Ujian**.
- `Readiness`: **Kesiapan**.
- `Blocker`: **Kendala Penghambat** atau **Perlu tindakan**.
- `Workflow`: **Alur Kerja** atau **Alur**.
- `Archive`: **Arsip**, **Arsipkan**, **Diarsipkan**.
- `Audit`: **Riwayat Perubahan** atau **Catatan Tindakan**.
- `Flag`: **Tanda Atensi**.
- `event atensi`: **Kejadian Perhatian**.
- `App Switch`: **Keluar/Pindah Aplikasi**.
- `Screenshot`: **Percobaan Tangkapan Layar**.
- `Mode Lengkap`: **Rincian Lengkap** untuk panel/accordion; gunakan **Mode Lengkap** hanya bila benar-benar toggle mode.
- `data teknis sesi`: **Rincian Sesi** atau **Informasi Lanjutan Sesi**.
- `Repositori`: **Bank Soal** atau **Daftar Soal** sesuai konteks.
- `Draft`: **Konsep** atau **Belum Dikirim**.
- `Refresh`: **Muat Ulang**.

## Token dan Kode Ujian

- Akses peserta: **Token Ujian**.
- Akses ruang/pengawas: **Token Ruang** atau **Kode Ruang** bila SOP ruang memakai istilah kode.
- Hindari **Token Rahasia** sebagai visible copy.
- Untuk siswa, gunakan **Token Ujian** secara konsisten di kartu ujian, halaman masuk, dan pesan bantuan.

## Istilah Utama Bank Soal

- `BFF Bank Soal`: **Data Bank Soal** atau **Sistem Bank Soal**.
- `endpoint`: **Layanan Sistem** atau **Data Sistem**.
- `summary`: **Ringkasan**.
- `sample endpoint`: **Contoh Data**.
- `evidence/data`: **Data Pendukung** atau **Data Belum Tersedia**.
- `Readiness Bank Soal`: **Kesiapan Bank Soal**.
- `Grade`: **Nilai Kesiapan**.
- `Review Backlog`: **Antrean Verifikasi** atau **Soal Menunggu Verifikasi**.
- `Queue Bank Soal`: **Antrean Verifikasi**.
- `Review & Terbitkan`: **Verifikasi & Terbitkan**.
- `Buka Review`: **Buka Verifikasi**.
- `Reviewer`: **Pemeriksa Soal** atau **Penelaah Soal**.
- `Approval`: **Persetujuan** atau **Setujui**.
- `Approved`: **Disetujui**.
- `Published`: **Terbit**.
- `bulk publish`: **Terbitkan Sekaligus**.
- Visible permission `bank_soal.publish`: **izin menerbitkan Bank Soal**.
- `Metadata`: **Identitas Soal** atau **Kelengkapan Data Soal**.
- `pagination`: **Halaman Daftar**.
- `Studio Penyusun`: **Ruang Penyusunan Soal**.
- `Studio Impor`: **Ruang Impor Soal**.
- `Blueprint`: **Kisi-kisi**.
- `Stimulus`: **Bacaan/Gambar Pendukung**.
- `Skoring deterministik`: **Kunci Jawaban Jelas**.
- `Distraktor`: **Pilihan Pengecoh**.
- `IndexedDB/httpOnly`: jangan tampil; ganti dengan **tersimpan di perangkat dan aman melalui sesi login**.

## Istilah Portal/Mobile CBT

- `BYOD`: **Perangkat Siswa**; helper: “perangkat siswa sendiri”.
- `Submit final`: **Kirim Jawaban Akhir**.
- `restore`: **Pulihkan Sesi**.
- `degraded mode`: **Koneksi Menurun**.
- `offline`: **Tidak Terhubung Internet**.
- `sinkronisasi`: boleh dipakai; untuk siswa gunakan instruksi lengkap seperti **Hubungkan internet, lalu klik Sinkronkan**.
- `intervensi keras`: **Pengawas wajib memeriksa perangkat**.

## Istilah yang Boleh Dipertahankan dengan Helper

- **CBT**: helper “ujian berbasis komputer”.
- **HOTS**: helper “soal berpikir tingkat tinggi”.
- **CP/TP/KD**: lazim untuk guru; hindari di layar siswa kecuali diperlukan.
- **PG/PG Kompleks**: lazim untuk guru.
- **Token**: boleh, tetapi bedakan **Token Ujian** dan **Token Ruang**.
- **Dashboard**: boleh pada root cepat/admin; rekomendasi copy end-user adalah **Beranda** untuk menu dan **Ringkasan** untuk section.

## Prioritas Cleanup Copy

### P0

- Command Center, Monitoring proctoring live, Proctoring.
- Token Rahasia.
- API, endpoint, BFF, summary/sampel endpoint.
- event atensi, flag, App Switch, Screenshot.
- Repositori bila maksudnya Bank Soal.
- Review Backlog, Queue Bank Soal, approved/published/bulk publish.
- Visible permission code seperti `bank_soal.publish`.
- IndexedDB/httpOnly pada layar penyusun soal atau portal.

### P1

- Dashboard → Beranda/Ringkasan sesuai konteks.
- Review/Reviewer → Verifikasi/Pemeriksa Soal.
- Draft → Konsep.
- Metadata → Identitas Soal/Kelengkapan Data Soal.
- Audit → Riwayat Perubahan.
- Mode Lengkap → Rincian Lengkap bila bukan toggle.
- BYOD → Perangkat Siswa.

## Contoh Rewrite Disarankan

- **Panduan BYOD** → **Panduan Perangkat Siswa**.
- **SCS · Bank Soal** → **Bank Soal Madrasah**.
- **Kelola filter, status, dan pagination soal** → **Cari soal, lihat status, dan pindah halaman daftar soal.**
- **Cek coverage metadata dan materi** → **Cek kelengkapan mapel, KD, dan materi soal.**
- **Pengaturan & SOP — alur kerja, standar kualitas, dan integrasi** → **Aturan Bank Soal — atur pemeriksaan, standar soal, dan aturan penggunaan.**
- **Menunggu Review** → **Menunggu Verifikasi**.
- **Mode advance** → **Rincian Lengkap**.
- **Metadata** → **Identitas Soal**.
- **Blueprint** → **Kisi-kisi**.
- **Stimulus** → **Bacaan/Gambar Pendukung**.
- **Skoring deterministik** → **Kunci jawaban jelas**.
- **Distraktor kanan** → **Pilihan pengecoh**.
- **Perubahan lokal menunggu sinkronisasi** → **Perubahan sudah tersimpan di perangkat ini. Hubungkan internet, lalu klik Sinkronkan.**

## Catatan untuk Test Label

Jika cleanup copy mengubah visible label, sesuaikan hanya ekspektasi label pada test yang membaca UI/menu. Jangan mengubah route contract atau daftar href.

Target test yang paling mungkin perlu penyesuaian label:

- `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`
  - Label sidebar Bank Soal: misalnya **Dashboard Bank Soal** → **Ringkasan Bank Soal**, **Review & Terbitkan** → **Verifikasi & Terbitkan**.
  - Label sidebar Asesmen: misalnya **Hari Ini / Dashboard CBT** → **Hari Ini / Ringkasan CBT**, **Monitor Ujian** → **Pengawasan Ujian**.
  - Href dan permission tetap sama.
- `apps/web-admin/src/routes/bank-soal/health-dashboard.test.ts`
  - Text visible: **Readiness Bank Soal**, **perlu evidence/data**, **Buka Review**, atau link yang mengandung **Review**.
  - Fetch path `/api/bank-soal/*` dan larangan `/api/cbt/*` tetap sama.
- `apps/web-admin/src/routes/bank-soal/final-route-map.test.ts`
  - Umumnya tidak perlu berubah karena fokus route/href, bukan visible copy.

Verifikasi minimal setelah cleanup copy:

```bash
npm --prefix apps/web-admin run test:unit -- \
  src/lib/components/sidebar/sidebar-config.test.ts \
  src/routes/bank-soal/final-route-map.test.ts \
  src/routes/bank-soal/health-dashboard.test.ts
```
