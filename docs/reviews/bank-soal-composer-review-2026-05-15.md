# Review Halaman Komposer Soal Bank Soal — 2026-05-15

## Scope

Review UX/flow halaman komposer Bank Soal setelah workflow versioning dan route detail `/bank-soal/soal/[id]`.

Sumber review:
- Screenshot mobile dari pengguna.
- Inspect kode `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`.
- Inspect model helper `apps/web-admin/src/routes/bank-soal/_components/soal-workspace.model.ts`.

## Ringkasan

Komposer sudah kuat secara fitur: mendukung banyak bentuk soal, mode pemula/advance, autosave lokal, preview siswa, validasi kesiapan, quality signals, mode fokus, upload gambar, workflow draft/review, read-only detail, revision/version panel.

Gap utama saat ini bukan di kapabilitas, tetapi di ergonomi dan kompleksitas operasional:

1. Halaman terlalu padat untuk mobile dan guru non-teknis.
2. Mode detail read-only masih memakai struktur editor, sehingga terasa seperti halaman edit yang dinonaktifkan.
3. Autosave/draft lokal kuat tetapi belum cukup terlihat dan belum punya recovery UX yang sederhana.
4. Validasi kualitas sudah ada, tetapi belum berubah menjadi rekomendasi konkret per field.
5. Versioning sudah aman, tetapi navigasi antar versi masih minimal.
6. Review/publish context belum cukup jelas untuk aktor berbeda: admin, pembuat soal, reviewer.

## Temuan dan Rekomendasi

### P0 — Reliability/bug guard

- Tambahkan regression test atau smoke test untuk route `/bank-soal/soal/[id]` agar ID route path tidak lagi tertimpa query param.
- Tambahkan state loading/error yang eksplisit ketika detail soal gagal dimuat. Saat ini user bisa mengira halaman kosong.
- Pisahkan `openEdit` dan `openReadOnlyDetail` secara konseptual agar read-only tidak ikut risiko path edit/draft.

### P1 — Mobile-first composer

- Jadikan composer mobile sebagai wizard step-by-step:
  1. Metadata
  2. Pertanyaan
  3. Opsi/Kunci/Rubrik
  4. Preview & Validasi
  5. Simpan/Kirim Review
- Sticky footer saat ini bagus, tetapi di mobile perlu ringkas: hanya tombol utama + progress; tombol sekunder masuk menu `Lainnya`.
- Tambahkan progress step kecil di atas, bukan hanya kartu alur yang memakan ruang.

### P1 — Detail/read-only mode

- Untuk `/bank-soal/soal/[id]`, defaultkan ke tampilan `Detail Soal` yang bersih:
  - metadata ringkas,
  - soal + opsi,
  - kunci/rubrik,
  - riwayat versi,
  - tombol aksi sesuai status.
- Berikan tombol `Edit/Revisi` untuk masuk ke editor jika memang boleh. Jangan langsung tampil seperti form editor yang disable.

### P1 — Navigation and context

- Breadcrumb perlu jelas: `Bank Soal > Detail Soal > vN` atau `Bank Soal > Edit Soal`.
- Tombol `Tutup` dari detail route sebaiknya kembali ke daftar dengan filter terakhir, bukan selalu mengubah mode internal.
- Simpan filter daftar terakhir di URL query agar guru tidak kehilangan konteks setelah melihat detail.

### P1 — Version panel

- Panel versi perlu menampilkan:
  - versi aktif/current,
  - versi terbaru/latest,
  - status setiap versi,
  - tanggal dibuat/diperbarui,
  - catatan versi.
- Tambahkan compare ringan v lama vs v terbaru: minimal metadata dan stem preview.

### P2 — Draft and autosave UX

- Tampilkan status autosave dengan timestamp yang mudah dibaca: `Draft lokal tersimpan 07:55`.
- Tambahkan drawer `Pulihkan Draft` bila ada draft lokal yang berbeda dari server.
- Saat membuka detail read-only, pastikan draft lokal tidak aktif dan tampilkan info `Mode lihat — draft lokal tidak digunakan`.

### P2 — Quality assistant

- Quality signals sudah ada tetapi masih berupa indikator. Ubah jadi checklist aksi:
  - `Stem terlalu pendek — tambahkan konteks/perintah yang lebih jelas`.
  - `Opsi mirip/duplikat — periksa opsi B dan C`.
  - `Panjang opsi tidak seimbang — ringkas opsi D`.
- Untuk Bank Soal madrasah, tambahkan template CP/TP/indikator sesuai mapel/tingkat bila data tersedia.

### P2 — Reviewer workflow

- Reviewer membutuhkan panel khusus:
  - quick approve/reject,
  - catatan reviewer inline,
  - checklist telaah: materi benar, opsi tidak ambigu, kunci benar, level kognitif sesuai, bahasa jelas.
- Tampilkan diff dari versi sebelumnya untuk revisi agar reviewer tidak membaca ulang semua.

### P2 — Accessibility and keyboard

- Pastikan mode fokus bisa ditutup dengan Escape.
- Pastikan tombol dan editor memiliki label yang jelas untuk screen reader.
- Hindari teks 10px terlalu banyak pada mobile; minimal informasi penting 12–14px.

### P3 — Performance/maintainability

- `SoalWorkspacePage.svelte` sudah sangat besar. Pisahkan menjadi:
  - `ComposerHeader.svelte`
  - `ComposerMetadataPanel.svelte`
  - `ComposerQuestionEditor.svelte`
  - `ComposerOptionsEditor.svelte`
  - `ComposerPreviewPanel.svelte`
  - `ComposerVersionPanel.svelte`
  - `ComposerFooterActions.svelte`
- Pindahkan load/detail/version logic ke helper/store agar bug shadowing props lebih mudah dicegah.

## Sprint Rekomendasi

### Sprint Composer 1 — Stabilitas Detail dan Mobile Read-only

- Tambah loading/error state detail route.
- Buat tampilan detail read-only khusus, bukan editor disabled.
- Perbaiki tombol kembali/tutup ke daftar dengan filter terakhir.
- Tambah smoke test route detail.

### Sprint Composer 2 — Mobile Wizard Editing

- Wizard step metadata → soal → jawaban → preview.
- Sticky footer ringkas.
- Step progress mobile.
- Validasi per step.

### Sprint Composer 3 — Reviewer UX

- Review checklist.
- Catatan reviewer inline.
- Queue review lebih cepat.
- Diff versi sederhana.

### Sprint Composer 4 — Refactor Maintainability

- Pecah komponen besar.
- Buat store/helper typed untuk state editor.
- Tambah unit/regression tests untuk route detail, read-only, version panel, autosave.

## Prioritas Eksekusi

Rekomendasi mulai dari Sprint Composer 1 karena langsung menutup gap yang tadi muncul: detail route kosong, read-only membingungkan, dan navigasi kembali dari detail.
