# CBT Operator Runbook

Status: sinkron per 2026-05-06. Gunakan runbook ini sebagai alur end-to-end operator untuk menyiapkan CBT, menjalankan ujian, dan menutup sesi tanpa melanggar boundary backend, BFF, dan Flutter BYOD.

Runbook ini melengkapi `docs/cbt-smoke-checklist.md`, `docs/exam-api.md`, `apps/mobile/RELEASE_CHECKLIST.md`, dan `apps/mobile/BYOD_TRIAL_PROCEDURE.md`.

## Prinsip Operasional

- Semua data CBT dibuat melalui web-admin dan Go API; tidak ada perubahan langsung ke PostgreSQL saat operasi harian.
- Backend adalah sumber kebenaran token, paket, peserta, kursi, proctoring, jawaban, dan scoring.
- Web-admin dipakai operator/guru/panitia untuk administrasi; Flutter APK dipakai siswa untuk runtime ujian.
- Token dan kunci jawaban adalah data sensitif. Admin/panitia berwenang melihat token untuk cetak operasional; guru biasa tidak boleh melihat token lintas scope.
- Migration dan deploy dilakukan di luar jendela ujian aktif. Jangan menjalankan migration saat siswa sedang login, menyimpan jawaban, heartbeat, atau submit.

## 1. Master Data dan User

1. Verifikasi tahun ajaran aktif, kelas, mapel, rombel, siswa, guru, dan penugasan guru-mapel sudah lengkap.
2. Pastikan akun web-admin tersedia untuk operator/panitia, guru pembuat soal, reviewer, proktor, pengawas, dan korektor sesuai kebutuhan event.
3. Pastikan role sistem pengguna sesuai kebijakan sekolah. Minimal `admin` dapat mengelola event; guru/staf terkait hanya diberi akses yang diperlukan.
4. Pastikan siswa yang ikut ujian sudah masuk kelas/scope yang benar sebelum event dan sesi dibuat.
5. Catat akun uji admin dan guru untuk smoke checklist sebelum ujian besar.

Kriteria siap:

- [ ] Tidak ada siswa ujian tanpa kelas/scope yang benar.
- [ ] Akun panitia/operator dapat login web-admin.
- [ ] Guru pembuat soal dapat membuka `/bank-soal/tambah` dan reviewer dapat membuka `/bank-soal/verifikasi` sesuai scope.
- [ ] Pengawas/proktor punya akun operasional bila dashboard pengawas dipakai.

## 2. Role Event CBT

Migration `062_cbt_event_members_question_scope.sql` menambahkan anggota event CBT dengan role operasional berikut.

| Role event | Subject scope | Tanggung jawab operasional |
|------------|---------------|----------------------------|
| `panitia` | Tidak memakai `subject_id` | Mengelola kegiatan ujian, jadwal, peserta, komunikasi operator, dan keputusan siap/tunda. |
| `pembuat_soal` | Wajib/bermakna per mapel saat dipakai untuk authoring | Menulis dan memperbaiki soal untuk mapel yang ditugaskan dalam event. |
| `reviewer` | Wajib/bermakna per mapel saat dipakai untuk review | Memeriksa kualitas soal, kunci, rubrik, readiness, dan memberi keputusan review/publish sesuai mapel. |
| `proktor` | Tidak memakai `subject_id` | Mengoperasikan ruang/sesi, token/kartu, reset akses sesuai prosedur, dan koordinasi teknis saat ujian. |
| `pengawas` | Tidak memakai `subject_id` | Mengawasi siswa di ruang, memantau status perangkat/koneksi, mencatat kejadian, dan eskalasi ke proktor/panitia. |
| `korektor` | Wajib/bermakna per mapel saat dipakai untuk penilaian | Menilai essay/uraian dan memastikan skor masuk sesuai rubrik untuk mapel yang ditugaskan. |

Catatan migration 062:

- Tabel `cbt_event_members` mengikat `event_id`, `user_id`, optional `employee_id`, optional `subject_id`, dan `role`.
- Constraint subject scope mengizinkan `subject_id` hanya untuk role `pembuat_soal`, `reviewer`, dan `korektor`; role `panitia`, `proktor`, dan `pengawas` harus bersifat event/ruang tanpa `subject_id`.
- Kombinasi assignment unik per event, user, role, dan subject mencegah duplikasi tugas yang sama.
- Kolom `cbt_questions.event_id` menghubungkan soal ke event sehingga authoring/review dapat discope ke kegiatan ujian.

## 3. Buat Kegiatan Ujian

1. Buat event CBT untuk periode ujian, misalnya PAT, PAS, Asesmen Madrasah, atau simulasi.
2. Isi nama, rentang tanggal, status operasional, dan catatan panitia.
3. Tambahkan anggota event sesuai role pada bagian sebelumnya.
4. Untuk pembuat soal, reviewer, dan korektor, tentukan mapel yang menjadi scope tugasnya.
5. Pastikan panitia/proktor/pengawas tidak diberi `subject_id` pada assignment event.

Kriteria siap:

- [ ] Event tampil pada daftar kegiatan CBT.
- [ ] Role event lengkap untuk minimal satu mapel/sesi uji.
- [ ] Tidak ada user yang mendapat tugas dobel untuk role dan mapel yang sama.

## 4. Event Members dan Tanggung Jawab Harian

1. Panitia menetapkan timeline upload soal, review, publish, paket, token, dan gladi bersih.
2. Pembuat soal mengisi atau mengimpor soal sesuai template dan mapel.
3. Reviewer memeriksa kunci, rubrik, bobot, tag kurikulum, readiness, dan status publish.
4. Proktor menyiapkan ruang, peserta, kartu/token, perangkat cadangan, dan kontak eskalasi.
5. Pengawas menerima daftar hadir, kartu/token sesuai prosedur, denah kursi, dan panduan status koneksi BYOD.
6. Korektor menilai essay setelah ujian selesai atau saat jadwal koreksi dibuka.

Kriteria siap:

- [ ] Setiap mapel punya pembuat soal dan reviewer yang jelas.
- [ ] Setiap ruang/sesi punya proktor atau pengawas yang tercatat secara operasional.
- [ ] Korektor ditetapkan untuk mapel yang memiliki essay/uraian.

## 5. Upload atau Import Soal

1. Buka `/bank-soal/tambah` sebagai pembuat soal atau admin.
2. Pilih mode `beginner` untuk input cepat atau `advance` untuk metadata lengkap.
3. Isi stem, stimulus, opsi, kunci, rubrik, bobot, CP/TP/KD, tingkat kesulitan, dan tag sesuai kebutuhan.
4. Untuk import, buka `/bank-soal/impor`, gunakan template CSV yang berlaku, dan cek hasil preview/import.
5. Unggah media hanya melalui fitur aset backend; jangan menyisipkan file dari storage publik di luar allowlist.
6. Simpan draft dan cek readiness score serta preview KaTeX/RTL bila ada konten matematika atau Arab.

Kriteria siap:

- [ ] Minimal soal objektif dan essay/uraian sudah tersedia untuk mapel uji.
- [ ] Kunci jawaban dan rubrik hanya terlihat oleh admin atau penulis/reviewer berwenang.
- [ ] Asset gambar/audio/PDF terbuka melalui URL backend yang sah.

## 6. Review dan Publish Soal

1. Reviewer membuka soal sesuai event/mapel scope.
2. Periksa kebenaran stem, opsi, kunci, rubrik essay, bobot, dan metadata kurikulum.
3. Minta revisi bila ada soal tidak siap, ambigu, atau bocor kunci.
4. Publish hanya soal yang siap masuk paket.
5. Jangan mengubah soal published menjelang ujian tanpa prosedur revisi/paket ulang yang disetujui panitia.

Kriteria siap:

- [ ] Semua soal dalam paket berasal dari soal published atau status workflow yang diizinkan.
- [ ] Soal essay punya rubrik koreksi.
- [ ] Tidak ada soal draft yang tidak sengaja masuk paket ujian resmi.

## 7. Buat Paket Ujian

1. Buat paket untuk event, mapel, kelas/grade/scope, dan durasi yang sesuai.
2. Tambahkan soal dari bank soal yang sudah siap.
3. Atur bobot, jumlah soal, shuffle, dan kebijakan campuran sesuai scope.
4. Validasi total bobot dan cakupan materi.
5. Freeze/publish paket sesuai workflow operasional sebelum sesi dibuat.

Kriteria siap:

- [ ] Paket berisi soal sesuai mapel dan scope peserta.
- [ ] Total bobot dan jumlah soal sesuai ketentuan panitia.
- [ ] Paket tidak mengandung soal lintas mapel/scope yang tidak diizinkan.

## 8. Buat Sesi Ujian

1. Buat sesi dari paket yang sudah siap.
2. Set jadwal mulai, selesai, durasi, scope kelas/grade/custom, dan mode assignment.
3. Pastikan jendela ujian tidak overlap dengan migration/deploy terencana.
4. Atur status sesi secara bertahap: draft, scheduled, active, completed, atau status lain yang tersedia.
5. Jangan mengaktifkan sesi sebelum peserta, ruang, kursi, token, dan smoke check selesai.

Kriteria siap:

- [ ] Jadwal sesi sesuai timezone WITA dan kalender ujian.
- [ ] Sesi belum aktif saat operator masih menata peserta/kursi/token.
- [ ] Paket yang dipakai adalah paket final yang disetujui.

## 9. Tambah Peserta

1. Tambahkan peserta dari kelas/grade/custom scope sesuai aturan event.
2. Pastikan tidak ada siswa dobel di sesi yang sama.
3. Verifikasi NIS/NISN/nama/kelas sebelum kartu dicetak.
4. Jika peserta pindah ruang atau tidak ikut, ubah sebelum token/kartu final dibagikan.

Kriteria siap:

- [ ] Jumlah peserta di web-admin sama dengan daftar panitia.
- [ ] Peserta tidak masuk sesi yang salah.
- [ ] Peserta cadangan atau susulan diberi prosedur terpisah bila diperlukan.

## 10. Ruang, Seat, Proktor, dan Pengawas

1. Buat atau pilih ruang ujian.
2. Assign peserta ke ruang dan seat.
3. Gunakan auto-assignment atau shuffle seat bila sesuai kebijakan panitia.
4. Verifikasi tidak ada `seat_no <= 0` dan tidak ada duplikasi `(room_id, seat_no)`.
5. Tetapkan proktor/pengawas per ruang secara operasional dan bagikan kontak eskalasi.
6. Cetak daftar hadir, denah/meja, dan paket pengawas bila tersedia.

Kriteria siap:

- [ ] Setiap peserta punya ruang yang benar.
- [ ] Seat unik dalam ruang.
- [ ] Pengawas tahu arti status `Tersambung`, `Lokal`, `Gangguan`, `Waspada`, dan `Menurun` dari panduan BYOD.

## 11. Token, Kartu, dan Reprint

1. Generate token hanya saat sesi, peserta, ruang, dan seat sudah final atau mendekati final.
2. Token harus berupa nilai acak kuat, 32 karakter hex untuk token baru.
3. Cetak kartu ujian dari data terbaru setelah token dibuat atau setelah operator menjalankan regenerasi/repair eksplisit.
4. Migration `060_cbt_exam_token_hardening.sql` tidak merotasi token yang sudah ada. Migration itu hanya menolak token legacy yang duplikat dan menambahkan uniqueness untuk token non-empty.
5. Reprint kartu hanya wajib setelah operator secara eksplisit menjalankan regenerasi token, repair token, atau perubahan data kartu seperti ruang/seat yang perlu tercetak ulang.
6. Jangan menyatakan semua kartu harus dicetak ulang hanya karena migration 060 dijalankan, kecuali migration gagal karena duplikasi dan operator memperbaiki token terdampak.

Kriteria siap:

- [ ] Token/kartu yang dibagikan adalah hasil cetak setelah perubahan token/ruang/seat terakhir.
- [ ] Token tidak terlihat pada akun guru/pengawas yang tidak berwenang.
- [ ] Prosedur reset/regenerasi token membutuhkan persetujuan panitia/proktor.

## 12. Flutter Mobile Trial

1. Build APK release sesuai `apps/mobile/RELEASE_CHECKLIST.md`.
2. Verifikasi Android manifest APK memiliki permission `INTERNET`.
3. Build memakai HTTPS base URL produksi/staging yang benar; hindari HTTP untuk trial nyata.
4. Pastikan token login uji 32 karakter dan sesuai sesi aktif/scheduled yang akan diuji.
5. Pastikan APK ditandatangani untuk distribusi internal dan versi build dicatat.
6. Jalankan `flutter analyze`, `flutter test`, dan build release sebelum distribusi.
7. Uji di perangkat nyata lintas vendor menggunakan `apps/mobile/DEVICE_TEST_MATRIX.md`.
8. Jalankan minimal: login token, status payload, render soal rich content/media/audio bila ada, simpan jawaban, heartbeat, app-switch/resume, final submit, dan restore.

Kriteria siap:

- [ ] APK bisa akses backend melalui HTTPS dari jaringan ujian.
- [ ] Login token, answer save, heartbeat, telemetry, status, dan submit sesuai `docs/exam-api.md`.
- [ ] Device matrix mencatat vendor, OS, hasil koneksi, secure screen, restore, dan submit.

## 13. Smoke Checks Sebelum Ujian

Jalankan `docs/cbt-smoke-checklist.md` setelah backend/frontend deploy dan sebelum sesi besar aktif. Minimal smoke harus mencakup:

- admin dapat melihat/export data yang memang berwenang
- guru pembuat soal hanya melihat scope soal sendiri
- guru non-penulis tidak melihat kunci soal orang lain
- token peserta tidak bocor ke role yang tidak berwenang
- scoring tidak menandai peserta belum submit sebagai submitted
- duplicate submit menjadi konflik terkendali
- Flutter dapat login, heartbeat, simpan jawaban, dan submit

Kriteria siap:

- [ ] Hasil smoke dicatat dalam deployment note atau tiket operasional.
- [ ] Tidak ada temuan kebocoran token/kunci, gagal submit, atau gagal simpan jawaban.
- [ ] Jika ada temuan kritis, sesi tidak diaktifkan sampai diperbaiki.

## 14. Pelaksanaan Ujian

1. Aktifkan sesi hanya setelah panitia menyatakan siap.
2. Pengawas membagikan kartu/token sesuai daftar ruang.
3. Siswa login di Flutter dengan token dan base URL yang sudah ditentukan operator.
4. Pantau dashboard sesi/ruang, heartbeat, status submit, dan kejadian BYOD.
5. Jika koneksi `Waspada` atau `Menurun`, ikuti panel guidance Flutter dan eskalasi ke pengawas/proktor.
6. Untuk reset akses atau token, catat alasan, waktu, peserta, petugas, dan persetujuan.
7. Jangan deploy/migrate/restart backend saat sesi aktif kecuali insiden darurat dan panitia menyetujui risiko.

Kriteria berjalan aman:

- [ ] Mayoritas perangkat stabil dan heartbeat masuk.
- [ ] Jawaban tersimpan sebelum submit final.
- [ ] Insiden dicatat dengan peserta, ruang, device, dan waktu.

## 15. Setelah Ujian

1. Pastikan peserta final submit atau ditangani sesuai prosedur susulan/force submit yang sah.
2. Tutup sesi ketika jendela ujian selesai dan panitia menyetujui.
3. Korektor menilai essay sesuai scope mapel.
4. Jalankan scoring/finalisasi sesuai alur backend, bukan edit manual database.
5. Export hasil sesuai kebutuhan panitia dan validasi sampel skor.
6. Arsipkan evidence: deployment note, smoke result, device matrix, daftar hadir, berita acara, dan catatan insiden.

Kriteria selesai:

- [ ] Tidak ada peserta yang statusnya menggantung tanpa keputusan panitia.
- [ ] Essay sudah dikoreksi atau masuk daftar koreksi lanjutan.
- [ ] Hasil export sesuai sesi, mapel, dan scope role.
