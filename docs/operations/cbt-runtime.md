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

## DEMO vs Simulasi/Gladi/Ujian nyata

- **DEMO harus meniru UI dan alur kontrol ujian nyata**: masuk, konfirmasi identitas, menunggu pengawas membuka ujian, mengerjakan soal, telemetry perangkat, simpan bertahap, dan kumpulkan.
- **DEMO bukan Simulasi/Gladi/Ujian nyata**. Mode ini tidak membuat kegiatan, sesi, peserta, paket, jawaban, nilai, audit, atau mutasi apa pun di API/database produksi.
- **DEMO lokal boleh memakai soal contoh bergaya Informatika**, termasuk contoh yang selaras dengan paket seed Informatika sistem, agar operator bisa melihat tampilan soal yang realistis.
- Jika operator bertanya apakah DEMO bisa memakai soal Informatika seed sistem dari kemarin: untuk DEMO lokal jawabannya **tidak secara langsung dari DB/API produksi**. Salin/kurasi contoh bergaya Informatika ke fixture lokal bila perlu; jangan membaca atau mengubah paket seed produksi dari mode DEMO.
- Untuk **Simulasi, Gladi Bersih, dan Ujian nyata**, sumber soal harus berasal dari paket/kegiatan server: buat atau pilih kegiatan, pilih paket, dan gunakan Bank Soal. Jika tersedia, gunakan **Bank Soal seed sistem Informatika** sebagai pool/paket awal yang kemudian tetap dikelola melalui alur paket resmi.
- Perbedaan utama: DEMO memvalidasi pengalaman pengguna; Simulasi/Gladi memvalidasi data operasional end-to-end dengan event, paket, sesi, peserta, pengawas, jawaban, nilai, dan audit di server.

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
- Memakai data contoh lokal, idealnya soal contoh bergaya Informatika agar mirip paket seed sistem tanpa membaca DB/API produksi.
- Tidak membutuhkan token/sesi server.
- Tidak mengirim jawaban ke server produksi.
- Muncul watermark/copy **MODE DEMO — DATA CONTOH**.
- Alurnya boleh dibuat sama dengan ujian nyata, tetapi penyimpanan dan submit tetap lokal.

## Mode DEMO Browser

Akses siswa:

```text
/ujian?demo=1
```

Akses pengawas ruang:

```text
/pengawas-ujian?demo=1
```

Karakteristik:

- Memakai data contoh lokal bergaya Informatika.
- Tidak login ke API produksi.
- Jawaban/status ruang disimpan lokal sementara di halaman.
- Portal siswa meniru alur ujian nyata: masuk, identitas, soal, simpan jawaban, kumpulkan.
- Portal pengawas meniru alur pengawasan nyata: lembar ruang, status ruang, tombol Mulai Ujian, Hubungi Admin, dan Tutup Ujian.
- Cocok untuk uji cepat browser/perangkat, layout, flow menunggu pengawas, dan perilaku offline/focus.
- Tidak boleh dipakai sebagai pengganti Simulasi/Gladi. Untuk Simulasi/Gladi, siapkan kegiatan dan paket server dari Bank Soal/paket seed Informatika bila tersedia.

## Lembar Pengawas Ruang

- Pengawas tidak memakai kartu personal.
- Akses pengawasan dicetak sebagai **Lembar Pengawas Ruang**: satu QR+PIN untuk satu ruang dan sesi.
- Jika pengawas utama berhalangan, guru pengganti yang ditunjuk dapat memakai lembar ruang yang sama tanpa menerbitkan ulang akses personal.
- Route/API lama tetap memakai nama kompatibilitas `supervisor-access-cards`, tetapi copy operasional harus menyebut lembar pengawas ruang.

## Kompatibilitas Android

- Baseline APK dipatok `minSdk = 24` (Android 7.0+) karena hasil build/dependency Flutter saat ini membutuhkan SDK 24.
- Rekomendasi operasional tetap Android 8+ untuk ujian resmi.
- Android 7 bersifat best effort; Android 6 ke bawah gunakan Browser Darurat atau perangkat cadangan.
