# Operator Quick Start — Uji BYOD CBT (arsip/nonaktif)

Panduan singkat ini ditujukan untuk pengawas atau operator sekolah saat mendampingi uji coba APK CBT di perangkat Android milik siswa.

**Status operasional tahun ini:** Portal Ujian Web (`/ujian`) adalah runtime resmi siswa. Panduan APK ini dipertahankan sebagai arsip/tahap lanjutan; jangan jadikan instruksi utama pengawas atau siswa tahun ini.

Status: sinkron per 2026-05-03. Panduan ini untuk BYOD; jangan menyampaikan bahwa aplikasi memberi jaminan kiosk penuh.

## Sebelum Siswa Masuk

1. Pastikan backend ujian aktif dan `API_BASE_URL` yang dipakai APK sudah benar.
2. Pastikan token ujian tersedia dan sesuai sesi.
3. Jika token/kartu baru saja diregenerasi, gunakan kartu terbaru.
4. Minta siswa:
   - baterai cukup
   - koneksi data/Wi-Fi stabil
   - tidak mengganti perangkat di tengah ujian
5. Pastikan APK sudah terpasang sebelum sesi dimulai.

## Saat Login

1. Siswa membuka aplikasi.
2. Siswa memasukkan token ujian.
3. Jika aplikasi mendeteksi sesi lama, cek kartu restore:
   - `Terakhir stabil` aman untuk lanjut cek ulang
   - `Pernah terganggu` perlu perhatian
   - `Perlu perhatian koneksi` artinya sesi sebelumnya tidak sehat

## Arti Status di Layar Ujian

- `Tersambung`
  Artinya perangkat baru saja terhubung baik ke server dan tidak ada jawaban lokal tertahan.

- `Lokal`
  Artinya ada jawaban yang masih aman di perangkat dan menunggu sinkron ulang.

- `Waspada`
  Artinya sudah cukup lama tidak ada kontak baru ke server. Belum tentu gagal, tapi perlu refresh status.

- `Gangguan`
  Artinya simpan jawaban atau refresh status baru saja bermasalah.

- `Menurun`
  Artinya gangguan sudah berulang dan submit manual memang ditahan sampai sinkron membaik.

## Jika Koneksi Mulai Bermasalah

1. Minta siswa tetap berada di layar ujian.
2. Tekan tombol sinkron ulang.
3. Periksa panel koneksi:
   - jika kembali `Tersambung`, siswa bisa lanjut
   - jika masih `Lokal` atau `Waspada`, terus pantau
   - jika `Menurun`, jangan izinkan submit dulu

## Jika Siswa Menutup atau Berpindah App

1. Saat aplikasi kembali, akan muncul pengecekan ulang.
2. Minta siswa menunggu sampai sistem selesai memeriksa status.
3. Jika restore gagal, gunakan token aktif untuk masuk ulang sesuai arahan pengawas.

## Sebelum Submit

1. Pastikan status tidak `Menurun`.
2. Pastikan tidak ada panel yang menyebut jawaban masih menunggu sinkron.
3. Baru izinkan siswa menekan `Kirim Ujian`.
