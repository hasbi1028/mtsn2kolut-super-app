# Release Checklist — Flutter CBT BYOD

Checklist ini untuk operator sekolah saat menyiapkan APK Android internal bagi siswa BYOD.

Status: sinkron per 2026-05-03. Gunakan bersama `docs/exam-api.md` dan `docs/cbt-smoke-checklist.md`.

## Sebelum Build

- [ ] backend Go sudah aktif dan endpoint CBT login peserta siap
- [ ] `API_BASE_URL` final untuk gelombang uji coba sudah dipastikan
- [ ] token ujian dan sesi uji tersedia untuk minimal 2-3 siswa percobaan
- [ ] pengawas paham bahwa aplikasi BYOD tidak setara kiosk penuh
- [ ] backend/frontend yang akan dipakai sudah tercatat commit-nya
- [ ] jika token hardening/migrasi CBT baru dijalankan, kartu ujian sudah dicetak ulang

## Quality Gate

Jalankan dari `apps/mobile`:

```bash
flutter pub get
dart format lib test
flutter analyze
flutter test
```

Semua harus hijau sebelum build APK.

## Build APK

```bash
flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
```

Output utama:

```text
build/app/outputs/flutter-apk/app-release.apk
```

## Verifikasi Lapangan Minimal

- [ ] install APK di minimal 2 vendor Android berbeda
- [ ] login token berhasil
- [ ] jawaban pilihan ganda tersimpan
- [ ] jawaban uraian tersimpan
- [ ] heartbeat tidak gagal terus-menerus
- [ ] restore sesi bekerja setelah app ditutup/buka lagi
- [ ] submit berhasil saat koneksi stabil
- [ ] panel warning muncul saat jaringan dimatikan sementara
- [ ] hasil setiap perangkat dicatat di `DEVICE_TEST_MATRIX.md`
- [ ] smoke admin/guru di `docs/cbt-smoke-checklist.md` tidak menemukan kebocoran token/kunci jawaban

## Distribusi Internal

- [ ] bagikan APK hanya lewat kanal resmi sekolah
- [ ] satu gelombang uji memakai satu `API_BASE_URL` yang sama
- [ ] siswa diminta memasang APK sebelum hari ujian
- [ ] siswa diberi instruksi untuk tidak mengganti perangkat di tengah sesi
- [ ] pengawas tahu arti status: `Tersambung`, `Lokal`, `Gangguan`, `Menurun`

## Catatan Operasional BYOD

- `Menurun` berarti sinkron gagal berulang; submit manual memang ditahan
- jika ada jawaban lokal menunggu sinkron, peserta tetap harus berada di layar ujian
- beberapa vendor Android agresif mematikan koneksi latar; pengawas perlu memeriksa kasus per perangkat
- `FLAG_SECURE` hanya mengurangi screenshot/recent preview, bukan jaminan anti-cheat penuh
- fingerprint perangkat hanya telemetry/resume hint, bukan identitas kuat
