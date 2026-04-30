# Prosedur Uji Coba BYOD CBT

Dokumen ini untuk uji lapangan APK CBT pada perangkat Android milik siswa.

## 1. Tujuan Uji Coba

Memastikan bahwa:

- login token berjalan
- jawaban tersimpan dan tersinkron
- restore sesi berjalan
- status koneksi dapat dibaca oleh siswa dan pengawas
- submit hanya dilakukan saat sesi benar-benar aman

## 2. Peran

### Operator teknis

- menyiapkan backend dan APK
- memastikan `API_BASE_URL` benar
- memeriksa hasil build dan distribusi APK

### Pengawas

- mendampingi siswa saat login dan selama ujian
- memantau status koneksi di layar siswa
- memutuskan kapan siswa boleh submit

### Siswa

- memakai perangkat yang sama dari awal sampai akhir
- tetap berada di layar ujian
- melapor jika status berubah ke `Gangguan`, `Waspada`, atau `Menurun`

## 3. Sebelum Hari Uji

1. Build APK release internal.
2. Pasang APK ke minimal 2-3 perangkat dari vendor berbeda.
3. Siapkan satu sesi ujian uji dan beberapa token aktif.
4. Pastikan backend Go siap menerima:
   - login
   - status
   - heartbeat
   - event warning
   - save answer
   - submit

## 4. Saat Siswa Masuk

1. Siswa buka aplikasi.
2. Masukkan token ujian.
3. Jika sesi lama terdeteksi:
   - baca kartu restore
   - jika `Perlu perhatian koneksi`, pengawas harus lebih waspada
4. Pastikan status awal tidak langsung `Menurun`.

## 5. Selama Ujian

Pengawas memantau:

- badge status di app bar
- panel kesehatan koneksi
- panel warning sinkron
- apakah ada jawaban lokal yang masih menunggu sinkron

Interpretasi cepat:

- `Tersambung`: aman
- `Lokal`: masih aman, tunggu sinkron
- `Waspada`: kontak server mulai lama, perlu refresh
- `Gangguan`: ada masalah sinkron terbaru
- `Menurun`: jangan izinkan submit

## 6. Simulasi Gangguan yang Wajib Diuji

1. Matikan jaringan sebentar saat siswa menjawab.
2. Nyalakan lagi dan pastikan jawaban lokal tersinkron.
3. Tutup aplikasi lalu buka lagi.
4. Pastikan resume gate muncul dan sesi bisa dicek ulang.
5. Coba kondisi submit saat status `Menurun` dan pastikan memang tertahan.

## 7. Sebelum Submit

Pengawas harus memastikan:

1. status bukan `Menurun`
2. tidak ada jawaban lokal tertahan
3. refresh status berhasil
4. siswa masih berada di perangkat yang sama

## 8. Setelah Uji Selesai

Catat:

- vendor/perangkat yang stabil
- vendor/perangkat yang sering gagal sinkron
- apakah audio/gambar soal tampil baik
- apakah restore bekerja baik
- apakah ada false positive yang terlalu sering membuat siswa tertahan

## 9. Kriteria Lulus Uji

- login berhasil pada perangkat target utama
- jawaban PG dan uraian dapat disimpan
- restore dapat dipulihkan
- submit berhasil saat koneksi sehat
- status koneksi dapat dipahami pengawas tanpa penjelasan teknis panjang
