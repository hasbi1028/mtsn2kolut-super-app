# CBT Runtime

## Jalur resmi tahun ini

- Runtime operasional resmi untuk ujian siswa tahun ini adalah **Portal Ujian Web** di SvelteKit (`/ujian`).
- Portal Ujian Web dipakai bersama portal pengawasan web, kartu/QR/PIN, dan panel ruang agar peserta/pengawas cukup membuka web sekolah tanpa instalasi APK.
- Istilah lama `Browser Darurat` masih dapat muncul pada kontrak/internal rollout lama, tetapi secara operasional copy siswa/pengawas harus menyebut **Portal Ujian Web**.
- Kontrol anti-cheat web tetap bersifat best effort: fokus layar, fullscreen, visibility/focus telemetry, heartbeat, penyimpanan jawaban lokal sementara, dan pengawasan ruang wajib.

## Flutter APK Android

- Source Flutter APK di `apps/mobile` dan artifact rilis tetap disimpan untuk arsip, audit, demo teknis, dan tahap lanjutan.
- APK Flutter **tidak aktif sebagai jalur operasional resmi tahun ini** kecuali ada keputusan tertulis baru dari admin/operator.
- Download center APK tetap tersedia sebagai arsip/nonaktif agar artifact lama bisa diverifikasi, tetapi tidak boleh dipromosikan sebagai jalur masuk ujian siswa tahun ini.
- Build yang tersimpan:
  - Split ABI: `app-arm64-v8a-release.apk`, `app-armeabi-v7a-release.apk`, `app-x86_64-release.apk` jika dibutuhkan emulator/uji teknis.
  - Universal: `app-release.apk` untuk arsip atau distribusi uji terbatas tahap lanjutan.

## CBT Portal dedicated

- Source legacy `apps/cbt-portal` tetap disimpan.
- PM2 `mtsn2kolut-cbt-portal` tidak dijalankan default.
- Jika perlu rollback sementara:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/cbt-portal
pm2 start start.sh --name mtsn2kolut-cbt-portal --update-env
pm2 save
```

## Portal Ujian Web / istilah lama Browser Darurat

- Portal Ujian Web tersedia melalui route web-admin `/ujian` dan menjadi runtime siswa resmi tahun ini.
- Label operasional: **Portal Ujian Web — Pengawasan Wajib**.
- Jika backend/telemetry masih memakai `client_type=web_fallback`, perlakukan itu sebagai nama kompatibilitas internal, bukan copy UI untuk siswa/pengawas.
- Login dan credential mengikuti rollout yang berlaku (QR+PIN/kartu ujian saat siap; token lama hanya sebagai kompatibilitas transisi).

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
