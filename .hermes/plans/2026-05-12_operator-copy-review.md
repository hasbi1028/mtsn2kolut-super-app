# Review Bahasa Operator Semua Modul

> **For Hermes:** Kerjakan bertahap per modul. Jangan deploy/restart PM2 sebelum diminta. Gunakan audit dan kamus copy global sebagai acuan.

**Goal:** Membersihkan istilah teknis yang tampil ke operator/guru/siswa/orang tua di seluruh modul non-Akademik.

**Architecture:** UI SvelteKit tetap memakai BFF/proxy existing. Tahap awal hanya audit dan kamus copy global, lalu refactor copy per modul tanpa mengubah alur data. Pesan dari layanan sistem dibersihkan pada tahap backend terpisah setelah UI module cleanup stabil.

**Tech Stack:** SvelteKit, TypeScript, Go core-api, PM2, PostgreSQL/sqlc.

---

## Prinsip bahasa

- Operator madrasah melihat bahasa operasional, bukan istilah developer.
- Istilah teknis boleh tetap ada di nama variabel, route, tipe data, dan komentar teknis, tetapi tidak sebagai teks tombol, judul, helper, badge, toast, atau pesan masalah.
- Istilah yang sudah familiar boleh dipertahankan jika diberi konteks, contoh CBT → Ujian Berbasis Komputer.
- Setiap tahap harus menjalankan `npm --prefix apps/web-admin run check`; tahap backend juga menjalankan Go test/build.

## Kamus awal

- endpoint/API → layanan sistem
- payload/JSON → data yang dikirim
- token → kode akses/kode ujian/kode ruang sesuai konteks
- matrix → tabel pengaturan/tabel penugasan
- debug → pemeriksaan masalah
- release/build/version → rilis/versi aplikasi
- bulk → aksi massal
- selected → dipilih
- role/RBAC/permission → peran/hak akses/izin akses
- session → sesi ujian/sesi kegiatan
- proctoring → pengawasan ujian
- participant → peserta
- randomization → pengacakan
- readiness → kesiapan
- package → paket soal/paket dokumen sesuai konteks
- template → format isian/format dokumen
- validation → pemeriksaan data
- duplicate → data ganda
- parser → pembaca file
- workflow/cycle → alur kerja/siklus dokumen
- compliance/evidence → kelengkapan/bukti pendukung
- sync/queue/retry/failed → sinkronkan/antrian proses/coba ulang/belum berhasil
- slug/SEO/publish/draft → alamat halaman/pengaturan pencarian/terbitkan/konsep

## Tahap 0 — Audit lintas modul

**Objective:** Peta awal istilah teknis di seluruh halaman non-Akademik.

**Files:**
- Create: `.hermes/plans/2026-05-12_operator-copy-audit.md`
- Create/Update: `.hermes/plans/2026-05-12_operator-copy-review.md`

**Hasil audit awal:**

- Total kandidat user-facing: **268**
- Modul terdampak: **21**
- Detail lengkap: `.hermes/plans/2026-05-12_operator-copy-audit.md`

**Prioritas berdasarkan audit:**
- Asesmen: 121 kandidat
- Pengaturan & Akun: 51 kandidat
- Tata Kelola: 29 kandidat
- Dokumen: 9 kandidat
- Nilai/Rapor: 9 kandidat
- Portal: 9 kandidat
- TU/Surat: 9 kandidat
- Inventaris: 5 kandidat
- Pusaka/Kehadiran: 5 kandidat
- Dashboard Umum: 3 kandidat
- Bank Soal: 3 kandidat
- Perpustakaan: 3 kandidat

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

## Tahap 1 — Kamus copy global

**Objective:** Membuat kamus bahasa operator reusable untuk modul non-Akademik.

**Files:**
- Create: `apps/web-admin/src/lib/copy/operator.ts`

**Isi minimum:**
- `commonActions`
- `commonTerms`
- `commonStatus`
- `commonMessages`
- `moduleTerms` untuk Asesmen, Bank Soal, Pengaturan, Pusaka, TU, Portal, Website, Inventaris, Perpustakaan

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

**Commit target:**

```bash
git add .hermes/plans/2026-05-12_operator-copy-review.md .hermes/plans/2026-05-12_operator-copy-audit.md apps/web-admin/src/lib/copy/operator.ts
git commit -m "docs(ui): audit operator-facing copy across modules"
```

## Tahap 2 — Asesmen

Fokus pada kegiatan asesmen, sesi ujian, pengawasan ujian, paket soal, aplikasi siswa, kartu ujian, berita acara, dan hasil.

## Tahap 3 — Bank Soal

Fokus pada tambah soal, impor soal, verifikasi, analisis butir, pengaturan bank soal, dan mapel/KD.

## Tahap 4 — Pengaturan & Akun

Fokus pada pengguna, hak akses, riwayat aktivitas, permintaan perubahan akun, profil madrasah, dan ringkasan penggunaan.

## Tahap 5 — TU/Surat & Dokumen

Fokus pada surat masuk/keluar, disposisi, arsip, surat keterangan, siklus dokumen, dan paket kelengkapan.

## Tahap 6 — Pusaka/Kehadiran

Fokus pada pegawai, kehadiran, ringkasan, antrian proses, sinkron data, dan coba ulang.

## Tahap 7 — Portal, Kesiswaan, Nilai/Rapor

Fokus pada portal siswa/orang tua, akun, nilai, rapor, dan status siswa.

## Tahap 8 — Website/Publikasi

Fokus pada halaman, berita, pengumuman, publikasi, konsep, terbitkan, alamat halaman, dan lampiran.

## Tahap 9 — Operasional lain

Fokus pada inventaris, perpustakaan, PPDB, tata kelola, notifikasi, dan dashboard umum.

## Tahap 10 — Pesan layanan sistem

Bersihkan pesan validasi dan masalah di `services/core-api/internal/handler` dan `services/core-api/internal/service`.

## Tahap 11 — Guard global

Buat `scripts/check-operator-ui-copy.sh` agar istilah teknis tidak muncul kembali pada UI.

## Tahap 12 — Final validation dan deploy

Jalankan guard global, frontend check/build, backend sqlc/test/build, lalu deploy dan restart PM2 hanya setelah diminta.
