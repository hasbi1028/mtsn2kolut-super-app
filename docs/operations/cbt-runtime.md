# CBT Runtime

## Jalur utama

- Client utama siswa untuk ujian resmi adalah Flutter APK di `apps/mobile`.
- APK Flutter mendukung kontrol perangkat yang lebih kuat daripada browser: fokus layar, deteksi app switch/split-screen best effort, device fingerprint, heartbeat, penyimpanan jawaban, dan telemetry anti-cheat.
- Build rilis yang disiapkan:
  - Split ABI: `app-arm64-v8a-release.apk`, `app-armeabi-v7a-release.apk`, `app-x86_64-release.apk` jika dibutuhkan emulator.
  - Universal: `app-release.apk` untuk distribusi sederhana.

## CBT Portal dedicated

- Source legacy `apps/cbt-portal` tetap disimpan.
- PM2 `mtsn2kolut-cbt-portal` tidak dijalankan default.
- Jika perlu rollback sementara:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/cbt-portal
pm2 start start.sh --name mtsn2kolut-cbt-portal --update-env
pm2 save
```

## Browser Darurat

- Browser fallback tersedia melalui route web-admin `/ujian`.
- Ini bukan pengganti keamanan Flutter APK.
- Gunakan hanya saat APK tidak dapat dipakai pada perangkat tertentu dan pengawas mengizinkan.
- Label operasional: **Mode Darurat / Browser — Pengawasan Wajib**.
- Login real memakai Token Ujian + Token Ruang dan mengirim `client_type=web_fallback` ke backend.

## Mode DEMO Flutter

Mode DEMO Flutter digunakan untuk tes cepat perangkat/tampilan/alur tanpa membuat simulasi atau gladi di database.

```bash
export PATH="/home/servermtsn2kolut/development/flutter/bin:$PATH"
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/mobile
flutter build apk --debug --dart-define=CBT_DEMO_MODE=true
flutter build apk --release --dart-define=CBT_DEMO_MODE=true
```

Karakteristik:

- Tombol **Masuk Mode DEMO** muncul hanya jika `CBT_DEMO_MODE=true`.
- Memakai data contoh lokal.
- Tidak membutuhkan token/sesi server.
- Tidak mengirim jawaban ke server produksi.
- Muncul watermark/copy **MODE DEMO — DATA CONTOH**.

## Mode DEMO Browser

Akses:

```text
/ujian?demo=1
```

Karakteristik:

- Memakai data contoh lokal.
- Tidak login ke API produksi.
- Jawaban disimpan lokal sementara di halaman.
- Cocok untuk uji cepat Android lama/browser, layout, dan perilaku offline/focus.

## Kompatibilitas Android

- Baseline APK dipatok `minSdk = 24` (Android 7.0+) karena hasil build/dependency Flutter saat ini membutuhkan SDK 24.
- Rekomendasi operasional tetap Android 8+ untuk ujian resmi.
- Android 7 bersifat best effort; Android 6 ke bawah gunakan Browser Darurat atau perangkat cadangan.
