# Device Test Matrix — BYOD CBT

Gunakan matriks ini saat uji perangkat Android siswa. Tujuannya agar hasil uji bisa dibandingkan antar vendor, versi Android, dan jenis koneksi.

Status: sinkron per 2026-05-03. Matriks ini dipakai bersama release checklist dan prosedur BYOD; hasilnya menjadi bukti operasional, bukan jaminan kiosk penuh.

## Cara Pakai

1. Satu baris untuk satu perangkat.
2. Isi status:
   - `Lulus`
   - `Perlu perhatian`
   - `Gagal`
3. Tambahkan catatan singkat jika ada perilaku aneh.

## Kolom yang Disarankan

| Vendor | Model | Android | RAM | Koneksi | Install APK | Login Token | Simpan PG | Simpan Uraian | Restore | Audio | Gambar | Status `Waspada` | Status `Menurun` | Submit | Catatan |
|--------|-------|---------|-----|---------|-------------|-------------|-----------|---------------|---------|-------|--------|------------------|------------------|--------|---------|
| Samsung | Galaxy A14 | 14 | 4 GB | Wi-Fi | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | - |
| Xiaomi | Redmi Note 11 | 13 | 4 GB | Data |  |  |  |  |  |  |  |  |  |  |  |
| Oppo | A57 | 13 | 4 GB | Wi-Fi |  |  |  |  |  |  |  |  |  |  |  |
| Vivo | Y21 | 12 | 4 GB | Data |  |  |  |  |  |  |  |  |  |  |  |

## Fokus Pengujian

Per perangkat, minimal cek:

1. APK bisa dipasang
2. login token berhasil
3. jawaban PG tersimpan
4. jawaban uraian tersimpan
5. restore sesi setelah app ditutup/buka lagi
6. audio soal bisa diputar jika ada
7. gambar soal tampil
8. status `Waspada` muncul saat kontak server stale
9. status `Menurun` muncul saat gangguan sinkron berulang
10. submit berhasil saat koneksi sehat
11. guidance panel muncul jelas untuk kondisi `403/409` atau server tidak terjangkau
12. tidak ada token/kunci jawaban yang terlihat di layar siswa

## Catatan yang Sebaiknya Dicatat

- vendor agresif mematikan app di background
- audio gagal di format tertentu
- gambar lambat muncul
- restore butuh waktu terlalu lama
- status terlalu sering berubah ke `Waspada`
- submit sering tertahan walau jaringan tampak baik

## Kriteria Prioritas Perangkat

### Layak dipakai produksi awal

- semua fungsi inti `Lulus`
- tidak sering masuk `Menurun`
- restore dan submit stabil

### Layak dengan catatan

- fungsi inti jalan
- ada issue minor seperti audio lambat atau perlu refresh manual sesekali

### Tidak direkomendasikan

- login/restore tidak stabil
- background policy terlalu agresif
- submit sering gagal pada jaringan yang sama dengan perangkat lain yang sehat
