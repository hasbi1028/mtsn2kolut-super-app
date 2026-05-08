# Device Test Matrix — BYOD CBT

Gunakan matriks ini saat uji perangkat Android siswa. Tujuannya agar hasil uji bisa dibandingkan antar vendor, versi Android, dan jenis koneksi.

Status: sinkron Phase 24 Anti-Cheat BYOD Evidence Completion per 2026-05-08. Matriks ini dipakai bersama release checklist dan prosedur BYOD; hasilnya menjadi bukti operasional, bukan jaminan kiosk penuh.

Manual evidence status: `pending_manual_evidence` until real Android devices are tested by operator/pengawas. Current status fields must be updated in the evidence bundle before final go/no-go.

Phase 24 rule: No fabricated real-device PASS. Real-device PASS cannot be claimed from repository docs, emulator-only checks, or generated templates. Keep rows pending/manual until an operator tests physical Android devices and signs the evidence bundle.

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

## Phase 24 deterministic build and hash instructions

Jalankan di mesin operator/CI yang memiliki Flutter SDK, lalu salin RC identifier dan hash ke matriks ini.

```bash
cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter pub get
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
/home/servermtsn2kolut/development/flutter/bin/flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
sha256sum build/app/outputs/flutter-apk/app-release.apk
```

Evidence wajib tetap manual untuk minimal dua vendor Android nyata:

- `FLAG_SECURE` screenshot/recent-preview deterrence.
- app switch event.
- resume gate.
- heartbeat loss.
- pending answer recovery.
- manual submit guard.
- device mismatch `409`.
- stale connection warning.
- final submit saat koneksi sehat.

## Two-vendor manual placeholders

These rows are placeholders for manual evidence. Replace Vendor A and Vendor B with real devices from at least two Android vendors before sign-off.

| Evidence row | Vendor placeholder | Current status | Requires physical Android device | Operator/reviewer | Notes |
|--------------|--------------------|----------------|----------------------------------|-------------------|-------|
| Device 1 | Vendor A | pending_manual_evidence | ya | | isi setelah uji perangkat nyata |
| Device 2 | Vendor B | pending_manual_evidence | ya | | isi setelah uji perangkat nyata |

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
