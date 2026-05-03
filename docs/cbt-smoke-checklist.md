# CBT Smoke Checklist — Admin/Guru Visibility & Runtime Integrity

Gunakan checklist ini sebelum ujian besar, setelah deploy backend/frontend CBT, dan setelah migration token CBT dijalankan di staging/produksi. Fokusnya adalah memastikan batas role admin/guru, token, kunci jawaban, export, dan submit tidak regress.

## 1. Prasyarat

- [ ] Backend Go sudah deploy dan health check hijau.
- [ ] Migration CBT terbaru sudah dijalankan sesuai urutan deploy.
- [ ] Jika `060_cbt_exam_token_hardening.sql` baru dijalankan, kartu ujian/token peserta sudah dicetak ulang.
- [ ] Frontend web-admin sudah deploy dari commit yang sama atau lebih baru dari backend.
- [ ] Tersedia akun uji:
  - `admin` panitia/operator.
  - `guru_a` sebagai penulis minimal 1 soal.
  - `guru_b` sebagai guru lain pada mapel/scope berbeda atau bukan penulis soal `guru_a`.
- [ ] Tersedia sesi uji kecil dengan minimal 2 peserta, 1 ruang, dan 1 pengawas.
- [ ] Tersedia minimal 1 soal objektif dan 1 soal essay/uraian untuk menguji kunci, rubrik, grading, dan skor.

## 2. Bank Soal: Label & Export

### Admin

1. Login sebagai `admin`.
2. Buka `/cbt/soal`.
3. Pastikan tombol export tertulis `Export CSV`.
4. Jalankan export dengan filter aktif yang kecil, misalnya satu mapel.
5. Buka CSV hasil download.

Kriteria lulus:

- [ ] CSV berhasil diunduh.
- [ ] CSV berisi soal sesuai filter admin.
- [ ] Kolom kunci/jawaban terisi sesuai data soal.
- [ ] Tidak ada error toast atau response 403.

### Guru

1. Login sebagai `guru_a`.
2. Buka `/cbt/soal`.
3. Pastikan tombol export tertulis `Export Soal Saya`.
4. Jalankan export pada filter yang sama.
5. Buka CSV hasil download.

Kriteria lulus:

- [ ] CSV berhasil diunduh.
- [ ] CSV hanya berisi soal dengan penulis `guru_a`.
- [ ] Soal milik `guru_b` atau guru lain tidak ikut keluar.
- [ ] Toast sukses menyebut `soal saya`.

## 3. Bank Soal: Kunci Jawaban Detail

### Admin

1. Login sebagai `admin`.
2. Buka detail/edit soal milik `guru_a`.

Kriteria lulus:

- [ ] Kunci jawaban/rubrik terlihat sesuai data soal.

### Guru Penulis

1. Login sebagai `guru_a`.
2. Buka soal yang ditulis oleh `guru_a`.

Kriteria lulus:

- [ ] Kunci jawaban soal sendiri terlihat.
- [ ] Edit tetap mengikuti guard soal terkunci/published sesuai workflow.

### Guru Non-Penulis

1. Login sebagai `guru_b`.
2. Buka detail soal milik `guru_a` dari route yang masih bisa dibaca oleh scope guru.

Kriteria lulus:

- [ ] Kunci jawaban kosong/tidak terbuka.
- [ ] Metadata yang aman tetap terbaca bila memang guru punya akses baca.
- [ ] Tidak ada cara melihat kunci melalui export guru.

## 4. Berita Acara / Minutes: Token Peserta

### Admin

1. Login sebagai `admin`.
2. Buka detail sesi uji.
3. Buka/cetak berita acara atau endpoint minutes.

Kriteria lulus:

- [ ] Token peserta terlihat untuk kebutuhan cetak operasional.
- [ ] Peserta, ruang, dan seat tampil sesuai sesi.

### Guru/Pengawas

1. Login sebagai guru/pengawas yang hanya berhak pada sesi/ruang terkait.
2. Buka berita acara atau data peserta sesi.

Kriteria lulus:

- [ ] Token peserta tidak tampil atau bernilai kosong.
- [ ] Peserta yang tampil sesuai scope guru/pengawas, bukan semua peserta lintas sesi/mapel.

## 5. Runtime: Skor, Grading, dan Status Submit

1. Pastikan ada peserta uji yang sudah mulai ujian tetapi belum final submit.
2. Dari admin, nilai satu jawaban essay peserta tersebut.
3. Jalankan hitung skor sesi jika tersedia.
4. Cek daftar peserta dan status peserta.
5. Minta peserta melanjutkan ujian atau menyimpan jawaban lagi.

Kriteria lulus:

- [ ] Grading essay tidak mengisi `submitted_at` untuk peserta yang belum submit.
- [ ] Hitung skor tidak mengisi `submitted_at` untuk peserta yang belum submit.
- [ ] Peserta belum submit masih bisa menyimpan jawaban dan final submit sesuai jadwal.
- [ ] Duplicate final submit ditolak sebagai konflik yang terkendali, bukan 500.

## 6. Token Login Flutter Setelah Hardening

1. Pakai kartu ujian/token terbaru setelah migration token.
2. Login dari perangkat pertama.
3. Coba login token yang sama dari perangkat kedua.
4. Lakukan heartbeat, simpan jawaban, dan submit dari perangkat yang valid.

Kriteria lulus:

- [ ] Token baru dapat login pada perangkat pertama.
- [ ] Token yang sudah device-bound tidak dapat dipakai perangkat lain tanpa prosedur reset yang sah.
- [ ] Heartbeat dan simpan jawaban sukses.
- [ ] Submit sukses hanya sekali.

## 7. Evidence Wajib Dicatat

Catat hasil smoke test dalam tiket/deployment note:

- tanggal dan jam uji
- environment: staging atau produksi
- commit backend/frontend
- akun role yang dipakai
- sesi ujian uji
- hasil export admin dan guru
- screenshot label export admin/guru
- screenshot detail soal admin/guru penulis/guru non-penulis
- screenshot berita acara/minutes admin vs guru
- hasil uji grading sebelum submit
- hasil login token Flutter
- daftar temuan dan keputusan: lulus / lulus dengan catatan / tunda deploy

## 8. Keputusan

- [ ] Lulus: boleh lanjut ujian/deploy.
- [ ] Lulus dengan catatan: boleh lanjut jika temuan tidak menyentuh token, kunci jawaban, submit, atau sinkron jawaban.
- [ ] Tunda: wajib jika ada kebocoran kunci/token, peserta belum submit terkunci, atau Flutter tidak bisa login/simpan/submit.
