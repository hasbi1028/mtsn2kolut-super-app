# Plan: Pindahkan field Tingkat keluar dari Detail Advance

## Goal

Field **Tingkat** pada composer Bank Soal tidak boleh tersembunyi saat mode **Pemula**. Field ini harus tampil sebagai bagian dari **Metadata / Data wajib** atau area dasar yang selalu terlihat, bukan di dalam blok `Detail Advance` yang hanya muncul saat mode **Advance**.

## Pemahaman UI saat ini

Dari screenshot dan inspeksi kode:

- Halaman: composer Bank Soal mobile (`/bank-soal` / `/bank-soal/tambah`).
- File utama: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`.
- Saat ini section `Metadata` selalu tampil dan berisi:
  - Mata Pelajaran
  - Kesulitan
  - Bobot lokal
  - Mode Arab / RTL
- Section `Detail Advance` hanya tampil saat `isAdvanceMode`.
- Field **Tingkat** (`fGradeLevel`, input `id="f-grade-level"`) saat ini berada di dalam `Detail Advance`, sehingga tersembunyi ketika mode Pemula aktif.

## Proposed approach

Pindahkan UI input **Tingkat** dari section `Detail Advance` ke section `Metadata` yang selalu tampil.

Prioritas perilaku:

1. **Tingkat selalu terlihat di mode Pemula maupun Advance.**
2. Nilai tetap memakai state dan payload yang sama: `fGradeLevel` / `gradeLevel`.
3. Tidak mengubah backend/database karena field sudah ada dan sudah masuk payload.
4. Tidak mengubah arti mode Advance; Advance tetap untuk data lanjutan seperti Fase, Topik, Kognitif, CP, TP, KD, Indikator, HOTS, workflow.

## Step-by-step implementation

1. Edit `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`.
2. Di section `Metadata`, tambahkan field **Tingkat** setelah **Mata Pelajaran** atau sebelum **Kesulitan**.
   - Rekomendasi urutan mobile:
     1. Mata Pelajaran
     2. Tingkat
     3. Kesulitan
     4. Bobot lokal
     5. Mode Arab / RTL
3. Hapus blok **Tingkat** dari grid `Detail Advance` agar tidak duplikat.
4. Pertahankan binding yang sama:
   - `id="f-grade-level"`
   - `type="number"`
   - `min="1"`
   - `max="12"`
   - `bind:value={fGradeLevel}`
5. Sesuaikan layout grid metadata supaya tetap rapi di mobile dan desktop.
   - Saat mobile: field tersusun vertikal/nyaman dibaca.
   - Saat desktop: grid bisa bertambah kolom agar tidak terlalu sempit.
6. Bila validasi/readiness sudah menganggap `gradeLevel` penting, pastikan hint/label tidak tersembunyi.
   - Jika belum wajib secara validasi, cukup jadikan terlihat sebagai data dasar dulu.
   - Jika diminta wajib eksplisit, tambahkan penanda `*` dan readiness check khusus.

## Files likely to change

- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`

Kemungkinan tidak perlu mengubah:

- Backend Go
- Migration DB
- SQLC
- API payload
- Offline draft helper

Karena `gradeLevel` sudah ada di state, draft payload, dan save payload.

## Tests / validation

Setelah implementasi:

1. Jalankan web check:
   - `npm run check` di `apps/web-admin`
2. Jalankan test relevan bila ada:
   - unit test web terkait Bank Soal/offline composer bila tersedia
3. Build web:
   - `npm run build` di `apps/web-admin`
4. Manual smoke test mobile:
   - buka composer Bank Soal
   - pilih mode **Pemula**
   - pastikan **Tingkat** tampil di Metadata/Data wajib
   - pilih mode **Advance**
   - pastikan **Tingkat** tetap tampil dan tidak muncul ganda di Detail Advance
   - simpan draft/soal dan pastikan nilai tingkat tetap masuk payload

## Risiko / catatan

- Risiko utama hanya layout: section metadata di mobile bisa lebih panjang. Ini sesuai tujuan karena Tingkat memang data dasar.
- Jangan mengubah field `Fase` kecuali diminta. Untuk sekarang yang dipindahkan hanya **Tingkat**.
- Jangan commit file plan unrelated yang sudah ada sebelumnya: `.hermes/plans/2026-05-09_094321-bank-soal-composer-responsive.md`.
