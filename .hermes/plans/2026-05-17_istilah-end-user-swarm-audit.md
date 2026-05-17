# Swarm Audit — Cleanup Istilah Teknis ke Bahasa End-User

> **Mode:** audit/rekomendasi. Jangan implementasi sebelum user menyetujui.
>
> **Target:** UI Asesmen CBT + Bank Soal + sebagian Portal/Mobile CBT agar istilah teknis/developer-facing tidak tampil ke guru, pengawas, siswa, dan operator madrasah.

## Ringkasan Swarm

Swarm 5 agent meninjau dari perspektif:

1. UI Asesmen CBT
2. UI Bank Soal
3. Kamus istilah lintas modul
4. Copywriting guru/pengawas non-teknis
5. QA risiko perubahan istilah

Kesimpulan bersama:

- Banyak struktur UI sudah lebih rapi setelah Opsi B.
- Masih ada istilah teknis/Inggris yang perlu dibersihkan dari copy visible.
- Perubahan harus **label/copy only**, jangan rename route/API/permission code.
- Prioritas tertinggi adalah teks yang terlihat oleh siswa, guru, pengawas, dan operator saat ujian.

## Prinsip Cleanup

1. Ganti **teks visible UI**, bukan kontrak teknis.
2. Jangan rename route/API/permission code tanpa migration/alias.
3. Pertahankan istilah yang sudah umum: CBT, token, HOTS, CP/TP/KD, PG.
4. Kalau istilah teknis tetap perlu, beri helper singkat.
5. Warning ujian harus pendek: apa yang terjadi, apakah aman, apa tindakan berikutnya.

## Kamus Final Rekomendasi

### Asesmen CBT

- `Event` → **Kegiatan Asesmen** / **Kegiatan**
- `Session` → **Sesi Ujian** / **Sesi**
- `Room` → **Ruang Ujian** / **Ruang**
- `Proctoring` → **Pengawasan Ujian** / **Pengawasan Ruang**
- `Command Center` → **Panel Pengawasan** / **Pusat Pemantauan Ujian**
- `Readiness` → **Kesiapan**
- `Blocker` → **Kendala Penghambat** / **Perlu tindakan**
- `Workflow` → **Alur Kerja** / **Alur**
- `Archive` → **Arsip** / **Arsipkan** / **Diarsipkan**
- `API` → **data sistem** / **layanan sistem**
- `Event atensi` → **kejadian perhatian**
- `Flag` → **tanda atensi**
- `App Switch` → **keluar/pindah aplikasi**
- `Screenshot` → **percobaan tangkapan layar**
- `Mode Lengkap` → **Rincian lengkap** / **Lihat rincian lanjutan**
- `data teknis sesi` → **rincian sesi** / **informasi lanjutan sesi**
- `Audit` → **Riwayat perubahan** / **Catatan tindakan**
- `Repositori` → **Bank Soal** / **Daftar Soal**
- `Draft` → **Konsep** / **Belum dikirim**
- `Refresh` → **Muat ulang**

### Token/Kode

Swarm berbeda sedikit, jadi rekomendasi final:

- Untuk akses peserta: **Token Ujian**
- Untuk ruang/pengawas: **Token Ruang** atau **Kode Ruang**
- Untuk siswa: gunakan satu istilah yang paling konsisten dengan kartu ujian/SOP. Rekomendasi QA: **Token Ujian** dan **Token Ruang**.
- Hindari `Token Rahasia`.

### Bank Soal

- `BFF Bank Soal` → **Data Bank Soal** / **Sistem Bank Soal**
- `endpoint` → **layanan sistem** / **data sistem**
- `summary` → **ringkasan**
- `sample/sampel endpoint` → **contoh data**
- `evidence/data` → **data pendukung** / **data belum tersedia**
- `Readiness Bank Soal` → **Kesiapan Bank Soal**
- `Grade` → **Nilai kesiapan**
- `Review Backlog` → **Antrean Verifikasi** / **Tumpukan Soal Menunggu Verifikasi**
- `Queue Bank Soal` → **Antrean Bank Soal** / **Antrean Verifikasi**
- `Review & Terbitkan` → **Verifikasi & Terbitkan**
- `Buka Review` → **Buka Verifikasi**
- `Reviewer` → **Pemeriksa Soal** / **Penelaah Soal**
- `Approval` → **Persetujuan** / **Setujui**
- `Approved` → **Disetujui**
- `Published` → **Terbit**
- `bulk publish` → **Terbitkan sekaligus**
- `bank_soal.publish` visible → **izin menerbitkan Bank Soal**
- `Metadata` → **Identitas Soal** / **Kelengkapan Data Soal**
- `pagination` → **halaman daftar**
- `Studio Penyusun` → **Ruang Penyusunan Soal**
- `Studio Impor` → **Ruang Impor Soal**
- `Blueprint` → **Kisi-kisi**
- `Stimulus` → **Bacaan/Gambar Pendukung**
- `Skoring deterministik` → **Kunci jawaban jelas**
- `Distraktor` → **Pilihan pengecoh**
- `IndexedDB/httpOnly` → jangan tampil; gunakan **tersimpan di perangkat dan aman melalui sesi login**

## P0 — Wajib Dibersihkan Dulu

### Asesmen CBT

- `Command Center`
- `Monitoring proctoring live`
- `Proctoring`
- `Token Rahasia`
- `API`
- `event atensi`, `flag`, `App Switch`, `Screenshot`
- `repositori`
- visible `Kode Ujian/Kode Ruang` jika SOP final memakai Token

### Bank Soal

- `BFF Bank Soal`
- `endpoint`
- `summary/sampel endpoint`
- `Review Backlog`
- `Queue Bank Soal`
- `approved/published/bulk publish`
- visible permission code seperti `bank_soal.publish`
- `IndexedDB/httpOnly` jika tampil di penyusun soal

## P1 — Ganti Bertahap

- `Dashboard` → **Beranda/Ringkasan** bila ingin lebih awam
- `Review/Reviewer` → **Verifikasi/Pemeriksa Soal**
- `Draft` → **Konsep**
- `Metadata` → **Identitas Soal/Kelengkapan Data Soal**
- `Audit` → **Riwayat Perubahan**
- `Mode Lengkap` → **Rincian lengkap**
- `BYOD` → **Perangkat Siswa** dengan helper “perangkat siswa sendiri”
- `CBT` tetap sebagai nama modul, helper: “ujian berbasis komputer”

## P2 — Cukup Tooltip/Helper

- `HOTS` — tetap boleh, helper “soal berpikir tingkat tinggi”.
- `CP/TP/KD` — tetap boleh untuk guru.
- `PG/PG Kompleks` — tetap boleh.
- `Token` — tetap boleh, tapi harus jelas Token Ujian vs Token Ruang.
- `Reviewer` — boleh di role teknis, tetapi proses UI tetap “Verifikasi”.

## Top Rewrite yang Paling Berdampak

1. `Panduan BYOD` → **Panduan Perangkat Siswa**
2. `SCS · Bank Soal` → **Bank Soal Madrasah**
3. `Kelola filter, status, dan pagination soal` → **Cari soal, lihat status, dan pindah halaman daftar soal.**
4. `Cek coverage metadata dan materi` → **Cek kelengkapan mapel, KD, dan materi soal.**
5. `Pengaturan & SOP — alur kerja, standar kualitas, dan integrasi` → **Aturan Bank Soal — atur pemeriksaan, standar soal, dan aturan penggunaan.**
6. `Menunggu Review` → **Menunggu Diperiksa** / **Menunggu Verifikasi**
7. `Mode advance` → **Mode Lengkap**
8. `Metadata` → **Identitas Soal**
9. `Blueprint` → **Kisi-kisi**
10. `Stimulus` → **Bacaan/Gambar Pendukung**
11. `Skoring deterministik` → **Kunci jawaban jelas**
12. `Distraktor kanan` → **Pilihan pengecoh**
13. `CSV` helper → **file yang sudah disiapkan**
14. `IndexedDB/httpOnly` → **tersimpan di perangkat dan dikirim lewat sesi login**
15. `Perubahan lokal menunggu sinkronisasi` → **Perubahan sudah tersimpan di perangkat ini. Hubungkan internet, lalu klik Sinkronkan.**
16. `Repositori bersama` → **Bank Soal umum**
17. `Submit final` → **Kirim jawaban akhir**
18. `restore` → **Pulihkan sesi**
19. `degraded mode` → **koneksi menurun**
20. `intervensi keras` → **pengawas wajib memeriksa perangkat**

## Batasan Aman Implementasi

Jangan ubah:

- route path seperti `/asesmen/...`, `/bank-soal/...`, `/proctoring`
- API/BFF path
- permission code seperti `asesmen.proctor`, `bank_soal.publish`
- DB enum/field
- test selector internal kecuali test label visible memang mengikuti copy baru

Boleh ubah:

- judul halaman
- deskripsi
- tombol
- badge/status visible
- empty/error/loading state
- helper text
- sidebar labels
- dokumen glossary/SOP UI

## Verifikasi Setelah Cleanup

Minimal:

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
npm --prefix apps/web-admin run test:unit -- \
  src/lib/components/sidebar/sidebar-config.test.ts \
  src/routes/bank-soal/final-route-map.test.ts \
  src/routes/bank-soal/health-dashboard.test.ts
```

Jika menyentuh route-access/permission:

```bash
npm --prefix apps/web-admin run test:unit -- \
  src/lib/server/assessment-route-contract.test.ts \
  src/lib/server/route-access.test.ts
```

Jika mobile app ikut dibersihkan:

```bash
# dari folder mobile sesuai struktur repo
flutter analyze
flutter test
```

## Rekomendasi Sprint Cleanup

### Sprint Copy 1 — P0 Web Admin

- Bersihkan Asesmen CBT visible copy P0.
- Bersihkan Bank Soal visible copy P0.
- Jangan ubah route/API/permission.

### Sprint Copy 2 — Penyusun Soal & Bank Soal Detail

- Metadata → Identitas Soal.
- Blueprint → Kisi-kisi.
- Stimulus → Bacaan/Gambar Pendukung.
- Review → Verifikasi.
- Draft → Konsep.

### Sprint Copy 3 — Portal/Mobile CBT

- Token/Kode diseragamkan.
- Pesan koneksi/sinkronisasi dipendekkan.
- Warning dibuat lebih tenang dan instruktif.

### Sprint Copy 4 — SOP/Glosarium

- Buat glosarium istilah baru.
- Sertakan “nama lama → nama baru” agar operator tidak kaget.

## Keputusan yang Perlu User Pilih

1. Proses Bank Soal pakai **Verifikasi** atau tetap **Review**?
   - Rekomendasi swarm: **Verifikasi**.
2. Akses peserta pakai **Token Ujian** atau **Kode Ujian**?
   - Rekomendasi QA: **Token Ujian** agar selaras dengan SOP formal.
3. `Dashboard` diganti menjadi **Beranda/Ringkasan** atau tetap Dashboard?
   - Rekomendasi: gunakan **Beranda** untuk menu, **Ringkasan** untuk section.
4. `Mode Lengkap` tetap atau diganti **Rincian lengkap**?
   - Rekomendasi: **Rincian lengkap** untuk accordion/detail; **Mode Lengkap** hanya jika benar-benar toggle mode.
