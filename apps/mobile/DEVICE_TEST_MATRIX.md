# Device Test Matrix — BYOD CBT

Gunakan matriks ini saat uji perangkat Android siswa. Tujuannya agar hasil uji bisa dibandingkan antar vendor, versi Android, dan jenis koneksi.

Status: sinkron Phase 13 Mobile Release Candidate and Device Matrix per 2026-05-08. Matriks ini dipakai bersama release checklist dan prosedur BYOD; hasilnya menjadi bukti operasional, bukan jaminan kiosk penuh.

## Phase 13 Mobile Release Candidate and Device Matrix

Sebelum rehearsal operator, catat release candidate yang diuji:

- RC identifier:
- APK SHA-256 hash:
- signing mode: release keystore / debug signing untuk uji teknis internal saja.
- `API_BASE_URL`:
- Flutter SDK absolute path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- minimum two Android vendors: ya / tidak.

Skenario wajib per RC:

- background/resume.
- heartbeat.
- pending answer.
- submit guard.
- device mismatch.
- screenshot protection / `FLAG_SECURE`.
- network disturbance.

## Cara Pakai

1. Satu baris untuk satu perangkat.
2. Isi status:
   - `Lulus`
   - `Perlu perhatian`
   - `Gagal`
3. Tambahkan catatan singkat jika ada perilaku aneh.

## Kolom yang Disarankan

Baris yang sudah terisi di bawah adalah contoh format, bukan hasil sertifikasi perangkat. Ganti dengan hasil uji nyata sekolah sebelum dipakai untuk keputusan operasional.

| RC identifier | APK SHA-256 hash | signing mode | `API_BASE_URL` | Vendor | Model | Android | RAM | Koneksi | Install APK | Login Token | background/resume | heartbeat | pending answer | submit guard | device mismatch | screenshot protection / `FLAG_SECURE` | network disturbance | Submit | Catatan |
|---------------|-----------------|--------------|----------------|--------|-------|---------|-----|---------|-------------|-------------|-------------------|-----------|----------------|--------------|-----------------|----------------------------------------|---------------------|--------|---------|
| Contoh: RC-2026-05-08-01 | contoh hash 64 hex | release keystore | https://api.sekolah.example | Samsung | Galaxy A14 | 14 | 4 GB | Wi-Fi | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | Lulus | contoh baris, ganti dengan hasil nyata |
| Contoh: RC-2026-05-08-01 | contoh hash 64 hex | release keystore | https://api.sekolah.example | Xiaomi | Redmi Note 11 | 13 | 4 GB | Data |  |  |  |  |  |  |  |  |  |  | contoh baris |
| Contoh: RC-2026-05-08-01 | contoh hash 64 hex | release keystore | https://api.sekolah.example | Oppo | A57 | 13 | 4 GB | Wi-Fi |  |  |  |  |  |  |  |  |  |  | contoh baris |
| Contoh: RC-2026-05-08-01 | contoh hash 64 hex | release keystore | https://api.sekolah.example | Vivo | Y21 | 12 | 4 GB | Data |  |  |  |  |  |  |  |  |  |  | contoh baris |

## Fokus Pengujian

Per perangkat, minimal cek:

1. APK bisa dipasang
2. login token 32 karakter dari kartu ujian berhasil
3. jawaban PG tersimpan
4. jawaban uraian tersimpan
5. restore sesi setelah app ditutup/buka lagi
6. background/resume melewati resume/status gate sebelum siswa lanjut
7. heartbeat memperbarui last-contact atau memunculkan warning terkendali
8. pending answer bertahan saat koneksi putus sementara
9. submit guard menahan submit saat pending sync/degraded mode belum pulih
10. device mismatch menghasilkan guidance `409` terkendali
11. screenshot protection / `FLAG_SECURE` mengurangi screenshot/recent preview sesuai kemampuan BYOD
12. network disturbance memunculkan status yang dapat dipahami
13. audio soal bisa diputar jika ada
14. gambar soal tampil
15. status `Waspada` muncul saat kontak server stale
16. status `Menurun` muncul saat gangguan sinkron berulang
17. submit berhasil saat koneksi sehat
18. guidance panel muncul jelas untuk kondisi `403/409` atau server tidak terjangkau
19. tidak ada token/kunci jawaban yang terlihat di layar siswa

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
