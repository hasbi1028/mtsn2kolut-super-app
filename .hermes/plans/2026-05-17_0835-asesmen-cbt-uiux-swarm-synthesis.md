# Swarm Review Synthesis — Asesmen CBT & Bank Soal UI/UX

> **For Hermes:** Plan/review mode only. Do not implement until user explicitly approves.

**Goal:** Menyintesis hasil 5 swarm agents atas plan UI/UX Asesmen CBT & Bank Soal agar menjadi opsi redesign yang lebih sederhana, rapi, dan tetap best-practice untuk operasional madrasah.

**Source plan:** `.hermes/plans/2026-05-17_082409-asesmen-cbt-uiux-redesign-map-plan.md`

---

## 1. Swarm Agents

1. **A1 — UX Architect / IA**
   - Fokus: information architecture, struktur modul, tab, sidebar.
2. **A2 — Operator/Panitia CBT**
   - Fokus: alur sebelum ujian, saat ujian, setelah ujian.
3. **A3 — UI Designer Admin Dashboard**
   - Fokus: visual clutter, cards/tabs/right rail, komponen reusable.
4. **A4 — Accessibility & Mobile-first Reviewer**
   - Fokus: HP/tablet/laptop, proctoring, keyboard, aria, responsive.
5. **A5 — Product Manager Sekolah**
   - Fokus: MVP paling berdampak, fitur yang disembunyikan/dipindah.

---

## 2. Konsensus Utama Swarm

Semua agents setuju arah plan sudah benar, tetapi masih bisa dibuat **lebih sederhana**.

Konsensus:

1. **Jangan tampilkan semua kemampuan sistem sekaligus.**
   Sidebar dan halaman detail harus menyembunyikan fitur lanjutan sampai user butuh.

2. **Asesmen CBT harus berbasis alur harian sekolah:**
   - sebelum ujian,
   - saat ujian,
   - setelah ujian.

3. **Detail Kegiatan jangan terlalu banyak tab.**
   7 tab berisiko ramai. Gabungkan menjadi 4–5 area kerja.

4. **Right rail jangan permanen.**
   Pakai hanya bila membantu keputusan saat ini: blocker, checklist, quick action, audit ringkas.

5. **Pengawasan harus exception-first.**
   Default tampilkan peserta/ruang bermasalah, bukan tabel semua peserta.

6. **Mobile-first harus masuk sejak awal.**
   Jangan jadikan mobile hanya adaptasi dari desktop 70/30.

7. **Bank Soal perlu role-aware, tapi struktur jangan beda total per role.**
   Guru/admin melihat prioritas berbeda, tetapi tetap dalam IA yang sama.

---

## 3. Opsi Rekomendasi

### Opsi A — Minimal Polish

**Konsep:** Struktur lama dipertahankan, hanya dirapikan visual dan copy.

**Isi:**
- Tetap pakai sidebar/route saat ini.
- Tambah shared page header.
- Batasi cards/tabs.
- Mode Sederhana pengawasan default.
- Rename istilah teknis.

**Kelebihan:**
- Paling cepat.
- Risiko kecil.
- Tidak banyak perubahan perilaku.

**Kekurangan:**
- Rasa “ramai” masih mungkin tersisa.
- Sidebar dan workflow tetap terasa kurang natural.

**Cocok jika:**
- Ingin perbaikan cepat tanpa perubahan struktur.

---

### Opsi B — Operational Redesign Layer **(Direkomendasikan)**

**Konsep:** Tidak rewrite backend/route, tetapi UI dipaketkan ulang berdasarkan alur kerja madrasah.

**Asesmen CBT menjadi:**
1. **Hari Ini**
2. **Persiapan Ujian**
3. **Monitor Ujian**
4. **Hasil & Berita Acara**
5. **Arsip**
6. **Aplikasi Siswa**

**Bank Soal menjadi:**
1. **Dashboard Bank Soal**
2. **Kelola Soal**
3. **Review & Terbitkan**
4. **Mutu Soal**
5. **Pengaturan**

**Kelebihan:**
- Lebih sederhana bagi user non-teknis.
- Tidak perlu rewrite besar.
- Route lama tetap bisa dipakai.
- Dampak UX besar.
- Cocok untuk madrasah: panitia berpikir “hari ini, persiapan, monitor, hasil, arsip”.

**Kekurangan:**
- Perlu disiplin menyembunyikan fitur lanjutan.
- Perlu beberapa komponen layout reusable.

**Cocok jika:**
- Bapak ingin tampilan lebih rapi, sederhana, tetapi tetap lengkap.

---

### Opsi C — Full Workflow Rewrite

**Konsep:** Ubah semua Asesmen/Bank Soal menjadi wizard/workflow terpadu penuh.

**Kelebihan:**
- Paling ideal secara UX jika dimulai dari nol.

**Kekurangan:**
- Risiko tinggi.
- Banyak halaman berubah.
- Butuh QA besar.
- User bisa bingung karena perubahan drastis.

**Cocok jika:**
- Ada waktu panjang dan siap training ulang.

---

## 4. Rekomendasi Final Swarm

Pilih **Opsi B — Operational Redesign Layer**.

Alasan:
- Sederhana, tapi tidak dangkal.
- Tetap best-practice: progressive disclosure, role-aware, mobile-first, accessibility sejak awal.
- Tidak membongkar backend.
- Bisa diterapkan bertahap.
- Fitur lama tidak hilang; hanya dipindah ke konteks yang tepat.

---

## 5. Struktur UI Final yang Disarankan

### 5.1 Asesmen CBT

#### Menu utama

1. **Hari Ini**
   - Ujian hari ini.
   - Sesi sedang berlangsung.
   - Peserta/ruang bermasalah.
   - Tombol cepat: buka monitor, cek BA, lihat hasil.

2. **Persiapan Ujian**
   - Buat/cek kegiatan.
   - Paket soal.
   - Peserta.
   - Sesi & ruang.
   - Token/kartu.
   - Checklist kesiapan.

3. **Monitor Ujian**
   - Sesi aktif.
   - Pengawasan ruang.
   - Peserta bermasalah.
   - Insiden.

4. **Hasil & Berita Acara**
   - Jawaban tersimpan.
   - Koreksi/nilai.
   - BA sesi/ruang.
   - Unduh rekap.

5. **Arsip**
   - Kegiatan selesai.
   - Dokumen final.
   - Checklist arsip.

6. **Aplikasi Siswa**
   - Panduan.
   - Release APK.
   - Matrix perangkat.
   - Troubleshooting.

#### Detail Kegiatan — versi lebih sederhana

Kurangi dari 7 tab menjadi 5 area:

1. **Ringkasan**
   - status kegiatan,
   - readiness,
   - langkah berikutnya,
   - blocker utama.

2. **Persiapan**
   - paket soal,
   - peserta,
   - ruang,
   - sesi,
   - token/kartu,
   - checklist SOP.

3. **Pelaksanaan**
   - sesi berjalan,
   - monitor,
   - ruang,
   - peserta bermasalah,
   - insiden.

4. **Hasil & BA**
   - submit/jawaban,
   - koreksi,
   - nilai,
   - berita acara.

5. **Arsip**
   - dokumen final,
   - status pengesahan,
   - audit ringkas.

#### Detail Sesi — versi lebih sederhana

Kurangi menjadi 4 area:

1. **Monitor**
2. **Peserta**
3. **Masalah/Insiden**
4. **Hasil & BA**

#### Pengawasan Ruang

Default:
- **Mode Sederhana**.
- Tampilkan peserta bermasalah dulu.
- Card list untuk HP/tablet.
- Table lengkap hanya di Mode Lengkap.

---

### 5.2 Bank Soal

#### Menu utama

1. **Dashboard Bank Soal**
   - pekerjaan hari ini,
   - draft saya,
   - perlu revisi,
   - menunggu review,
   - tombol tambah soal.

2. **Kelola Soal**
   - daftar soal,
   - tambah/import,
   - filter mapel/tingkat/status.

3. **Review & Terbitkan**
   - menunggu review,
   - perlu revisi,
   - siap terbit,
   - terbit.

4. **Mutu Soal**
   - coverage mapel/KD,
   - kelengkapan metadata,
   - distribusi tipe soal,
   - analisis butir jika data siswa sudah lengkap.

5. **Pengaturan**
   - scope reviewer,
   - mapel/KD,
   - import,
   - standar mutu,
   - integrasi asesmen.

#### Editor Soal

Default jangan langsung 3 kolom penuh.

Rekomendasi:
- Laptop/desktop biasa: 2 kolom.
  - kiri/besar: editor soal,
  - kanan/collapsible: preview + checklist mutu.
- Desktop besar: 3 kolom opsional.
- Mobile: stepper.

---

## 6. Layout System Minimum

Swarm menyarankan jangan membuat terlalu banyak komponen. Minimum saja:

1. **PageHeader**
   - title,
   - subtitle,
   - breadcrumb/context,
   - primary action.

2. **ContextStrip**
   - tahun ajaran,
   - kegiatan/sesi aktif,
   - status singkat.

3. **WorkflowCard**
   - dipakai di landing/hub.

4. **MetricCard**
   - maksimal 3 kartu utama per halaman.

5. **EntityTabs**
   - maksimal 5 tabs.

6. **DataTableShell**
   - filter,
   - search,
   - empty state,
   - row action,
   - drawer detail.

7. **BlockerPanel**
   - hanya tampil kalau ada blocker.

8. **StatusRail / InfoDrawer**
   - desktop: optional right rail,
   - mobile: drawer/accordion.

9. **ActionBar**
   - hanya untuk save/approval/bulk action.

---

## 7. Rules Sederhana agar UI Tidak Ramai

1. Satu layar = satu tujuan utama.
2. Satu primary action per layar.
3. Summary card maksimal 3 sebelum konten utama.
4. Tabs maksimal 5.
5. Right rail optional, bukan selalu tampil.
6. Tabel untuk data banyak; card untuk ringkasan/workflow.
7. Detail kecil pakai drawer, bukan halaman/tab baru.
8. Fitur teknis masuk `Mode Lengkap` atau `Fitur Lanjutan`.
9. Mobile harus card-first untuk proctoring.
10. Aksesibilitas masuk dari sprint pertama, bukan polish akhir.

---

## 8. MVP 3 Sprint yang Direkomendasikan

### Sprint 1 — Shell + Hari Ini + Detail Kegiatan Simple

**Tujuan:** rasa rapi langsung terasa.

Deliverables:
- PageHeader/ContextStrip/WorkflowCard/MetricCard basic.
- `/asesmen` menjadi `Hari Ini / Dashboard CBT`.
- Detail Kegiatan menjadi 5 area: Ringkasan, Persiapan, Pelaksanaan, Hasil & BA, Arsip.
- BlockerPanel hanya muncul saat ada blocker.

### Sprint 2 — Monitor Ujian Mobile-first

**Tujuan:** hari-H ujian sederhana.

Deliverables:
- Pengawasan ruang default Mode Sederhana.
- HP/tablet card-first.
- Table lengkap hanya Mode Lengkap.
- Aksi cepat aman: periksa, peringatkan, instruksi masuk ulang, buka kunci.
- Accessibility wajib: focus, aria, touch target, dialog confirmation.

### Sprint 3 — Bank Soal Role Cleanup

**Tujuan:** guru/admin tidak bingung.

Deliverables:
- Dashboard Bank Soal role-aware tapi struktur sama.
- Kelola Soal sederhana.
- Review & Terbitkan sebagai queue.
- Mutu Soal menggantikan/menaungi Analisis Butir jika belum item-analysis penuh.
- Editor default 2 kolom + side panel collapsible; 3 kolom opsional.

### Sprint 4 — Sidebar Cleanup + Polish

**Tujuan:** navigasi akhir rapi setelah layout dalam stabil.

Deliverables:
- Sidebar diringkas.
- Fitur lanjutan dipindah ke detail/context.
- Empty/loading/error state konsisten.
- Route contract tests.
- Build + smoke test.

---

## 9. Keputusan yang Disarankan untuk User

Rekomendasi final:

> **Pakai Opsi B: Operational Redesign Layer. Mulai dari 3 sprint inti: Hari Ini/Detail Kegiatan, Monitor Ujian, Bank Soal. Sidebar cleanup dilakukan setelah layout dalam stabil.**

Ini menjaga prinsip user:
- sederhana,
- rapi,
- tetap lengkap,
- best-practice,
- tidak rewrite besar,
- cocok untuk operasional madrasah.

---

## 10. Open Decisions

1. Nama menu utama Asesmen: pilih `Hari Ini` atau `Dashboard CBT`?
2. Tab Detail Kegiatan: setuju 5 area atau tetap 7 tab?
3. `Analisis Butir` diganti ke `Mutu Soal` dulu?
4. Pengawasan HP/tablet default card-list, table hanya mode lengkap?
5. Sidebar cleanup dilakukan setelah 3 sprint inti?
