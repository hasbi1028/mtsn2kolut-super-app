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

**Implementation Notes (2026-05-12):**

- Membersihkan copy operator pada 25 file Svelte modul Asesmen.
- Istilah teknis user-facing yang diganti mencakup token, proctoring, session, event, package, readiness, matrix, release/build/version, endpoint/API/base URL, payload/JSON, dan pesan render failed.
- Padanan utama: kode ujian/kode ruang/kode akses, pengawasan ujian, sesi ujian, kegiatan asesmen, paket soal, kesiapan ujian, tabel perangkat, rilis/versi aplikasi, alamat layanan sistem, data pengawasan, dan belum dapat ditampilkan.
- Tidak mengubah route, endpoint, tipe data, nama variabel internal, atau logic operasional ujian.
- Audit residual Tahap 2–3: `visible 0`, `code_only 443`.
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 3 — Bank Soal

Fokus pada tambah soal, impor soal, verifikasi, analisis butir, pengaturan bank soal, dan mapel/KD.

**Implementation Notes (2026-05-12):**

- Membersihkan copy operator pada komponen/halaman Bank Soal yang terkena audit: dashboard kesehatan, daftar soal, panel impor, pemeriksaan, workspace penyusun soal, dan pengaturan.
- Istilah teknis user-facing yang diganti mencakup template, dry-run, error, permission, event, bulk, workflow, dan render failed.
- Padanan utama: format isian, cek data sebelum impor, masalah, izin akses, kegiatan, aksi massal, alur verifikasi, dan belum dapat ditampilkan.
- Tidak mengubah alur impor, penyusun soal, verifikasi, atau integrasi layanan.
- Audit residual Tahap 2–3: `visible 0`, `code_only 443`.
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 4 — Pengaturan & Akun

Fokus pada pengguna, hak akses, riwayat aktivitas, permintaan perubahan akun, profil madrasah, dan ringkasan penggunaan.

**Implementation Notes (2026-05-12):**

- Membersihkan copy operator pada modul Pengaturan/Akun: users, account, analytics, audit logs, dan hak akses.
- Istilah teknis user-facing yang diganti mencakup RBAC, role, permission, matrix, slug, diff, warning, analytics, audit trail, Core API, token/session, dan metadata.
- Padanan utama: hak akses pengguna, peran, izin akses, tabel akses, alamat singkat, daftar perubahan, peringatan dampak, ringkasan penggunaan, riwayat aktivitas, layanan utama, kode akses/sesi, dan info.
- Tidak mengubah route, endpoint, tipe data, nama variabel internal, atau logic pengelolaan akun/hak akses.
- Audit residual Tahap 4–5: `visible 0`, `code_only 243`.
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 5 — TU/Surat & Dokumen

Fokus pada surat masuk/keluar, disposisi, arsip, surat keterangan, siklus dokumen, dan paket kelengkapan.

**Implementation Notes (2026-05-12):**

- Membersihkan copy operator pada modul TU/Surat dan Dokumen: dashboard TU, arsip, surat keterangan, surat masuk, dan siklus dokumen.
- Istilah teknis user-facing yang diganti mencakup template, metadata, audit, UUID, event, dan istilah teknis dokumen yang tampil ke operator.
- Padanan utama: format surat, info dokumen, pemeriksaan administrasi, ID pegawai, kejadian, dan riwayat pemeriksaan.
- Tidak mengubah struktur data arsip/surat, API, route, atau alur pembuatan/verifikasi dokumen.
- Audit residual Tahap 4–5: `visible 0`, `code_only 243`.
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 6 — Pusaka/Kehadiran

Fokus pada pegawai, kehadiran, ringkasan, antrian proses, sinkron data, dan coba ulang.

**Implementation Notes (2026-05-12):**

- Membersihkan copy operator pada modul PUSAKA/Kehadiran: ringkasan PUSAKA, antrian pekerjaan, pegawai PUSAKA, data kehadiran, dan ringkasan kehadiran.
- Istilah teknis user-facing yang diganti mencakup job, worker, scheduler, queue phrase, dashboard, eligible, dan backend pada pesan operasional PUSAKA.
- Padanan utama: pekerjaan, petugas sistem, jadwal otomatis, antrian, halaman ringkasan, pegawai yang memenuhi syarat, dan layanan sistem.
- Tidak mengubah route, endpoint, worker queue, format data PUSAKA, atau alur rekap.
- Audit residual Tahap 6–7: `visible 0`, `code_only 114` (termasuk MIME `text/csv` code-only pada ekspor nilai).
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 7 — Portal, Kesiswaan, Nilai/Rapor

Fokus pada portal siswa/orang tua, akun, nilai, rapor, dan status siswa.

**Implementation Notes (2026-05-12):**

- Membersihkan copy operator/siswa pada Portal Siswa dan halaman ujian siswa, serta meninjau Kesiswaan, Nilai/Rapor, dan Orang Tua.
- Istilah teknis user-facing yang diganti mencakup token ujian/ruang dan dashboard guru.
- Padanan utama: kode ujian, kode ruang, kode akses, dan ringkasan guru.
- Tidak mengubah route portal, endpoint ujian, validasi kode ruang, format ekspor nilai, atau logika rapor.
- Audit residual Tahap 6–7: `visible 0`, `code_only 114` (termasuk MIME `text/csv` code-only pada ekspor nilai).
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 8 — Website/Publikasi

Fokus pada halaman, berita, pengumuman, publikasi, konsep, terbitkan, alamat halaman, dan lampiran.

**Implementation Notes (2026-05-12):**

- Meninjau modul Website/Publikasi: halaman website, berita, pengumuman, halaman publik berita/pengumuman, profil, dan kontak.
- Membersihkan copy pengumuman dari istilah `workflow draft ke publish` menjadi bahasa operator: konsep sampai terbit.
- Tidak mengubah route publik, jenis konten, komponen pengelola konten, slug internal, atau alur publikasi.
- Audit residual Tahap 8–9: `visible 0`, `code_only 167`.
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 9 — Operasional lain

Fokus pada inventaris, perpustakaan, PPDB, tata kelola, notifikasi, dan dashboard umum.

**Implementation Notes (2026-05-12):**

- Meninjau modul operasional lain: Inventaris, Perpustakaan, PPDB, Tata Kelola, Notifikasi, Jurnal, Jadwal umum, dan halaman Academic lama.
- Membersihkan istilah user-facing `matriks mingguan` pada halaman Academic lama menjadi `tabel mingguan`.
- Sisa kandidat seperti MIME `text/csv` pada ekspor inventaris/tata kelola/jadwal terklasifikasi code-only karena bukan teks yang tampil ke operator.
- Tidak mengubah route, ekspor CSV, logika inventaris/perpustakaan/PPDB/tata kelola, atau komponen publik.
- Audit residual Tahap 8–9: `visible 0`, `code_only 167`.
- Verifikasi: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 10 — Pesan layanan sistem

Bersihkan pesan validasi dan masalah di `services/core-api/internal/handler` dan `services/core-api/internal/service`.

**Implementation Notes (2026-05-12):**

- Membersihkan pesan layanan sistem yang bisa diteruskan ke UI dari istilah teknis seperti `invalid json`, `invalid id`, `event CBT`, `bulk workflow`, `token`, `session id`, `participant id`, dan `backend/database` style message.
- Padanan utama: `Data yang dikirim tidak valid`, `ID data tidak valid`, `kegiatan asesmen`, `aksi massal alur verifikasi soal`, `kode ujian/kode ruang`, `sesi ujian`, `peserta ujian`, dan `layanan sistem`.
- Scope file: handler/service Go pada `services/core-api/internal/handler` dan `services/core-api/internal/service`; test expectation diperbarui mengikuti copy operator baru.
- Tidak mengubah route API, struktur response, query sqlc, kontrak data, atau flow bisnis.
- Verifikasi backend: `sqlc generate`, `go test ./internal/handler ./internal/service ./internal/repository/postgres`, dan `go build -o /tmp/core-api-operator-copy ./cmd/api` PASS.

## Tahap 11 — Guard global

Buat `scripts/check-operator-ui-copy.sh` agar istilah teknis tidak muncul kembali pada UI.

**Implementation Notes (2026-05-12):**

- Membuat guard global executable `scripts/check-operator-ui-copy.sh` untuk memindai `apps/web-admin/src/routes/**/*.svelte`.
- Guard mendeteksi istilah teknis user-facing lintas modul dan mengabaikan code-only seperti import path, route, MIME type, atribut HTML, endpoint internal, dan template literal dinamis.
- Guard memicu cleanup tambahan pada copy UI yang masih lolos dari tahap sebelumnya: shortcut permission, backend, metadata audit, legacy editor, export CSV, dan workflow draft/publish.
- Hasil guard: `PASS: tidak ada istilah teknis user-facing pada UI operator. Code-only diabaikan: 1144.`
- Verifikasi frontend: `npm --prefix apps/web-admin run check` PASS (`0 errors and 0 warnings`).

## Tahap 12 — Final validation dan deploy

Jalankan guard global, frontend check/build, backend sqlc/test/build, lalu deploy dan restart PM2 hanya setelah diminta.
