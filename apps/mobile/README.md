# Flutter CBT Client

Flutter Android client untuk peserta CBT MTs Negeri 2 Kolaka Utara.

Status: sinkron per 2026-05-03. Aplikasi ini adalah client/APK BYOD, bukan service VPS.

## Fokus MVP

- login token peserta ke backend Go
- render soal pilihan ganda dan uraian
- heartbeat + refresh status
- simpan jawaban lokal dan sinkron ulang saat koneksi membaik
- restore sesi aktif bila aplikasi dibuka kembali
- baseline deterrence BYOD: `FLAG_SECURE`, blok tombol kembali, event `app_switch`, resume re-check
- guidance panel untuk kondisi backend penting seperti `403`, `409`, dan transport failure
- secure snapshot split: token/fingerprint/jawaban sensitif tidak disimpan bersama metadata restore ringan

## Constraint BYOD

Perangkat adalah milik siswa sendiri, jadi aplikasi ini tidak bisa menjanjikan mode kiosk penuh seperti device sekolah yang dikelola. Targetnya adalah:

- menjaga jawaban tetap aman di perangkat saat jaringan buruk
- memperjelas status sinkron dan gangguan koneksi
- mengirim telemetry peringatan ke server saat perilaku resume/switch berisiko
- menahan submit manual saat mode koneksi menurun berulang

Yang tidak boleh diklaim:

- bukan kiosk penuh untuk semua vendor Android
- fingerprint perangkat bukan bukti identitas kuat
- `FLAG_SECURE` bukan jaminan anti-cheat total

## Menjalankan Lokal

```bash
cd apps/mobile
flutter pub get
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080
```

Untuk device fisik Android di jaringan yang sama, ganti `API_BASE_URL` ke IP backend lokal, misalnya:

```bash
flutter run --dart-define=API_BASE_URL=http://192.168.1.20:8080
```

Untuk preview lokal di desktop Linux, gunakan:

```bash
flutter run -d linux --dart-define=API_BASE_URL=http://127.0.0.1:8080
```

Catatan:

- target utama produksi tetap Android BYOD
- Linux desktop hanya untuk preview lokal, demo operator, dan debugging cepat
- desktop Linux membutuhkan dependency sistem GStreamer untuk plugin audio

## Quality Checks

```bash
cd apps/mobile
dart format lib test
flutter analyze
flutter test
```

## Build APK Internal

Build ini ditujukan untuk uji coba internal BYOD, bukan distribusi Play Store.

```bash
cd apps/mobile
flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
```

Hasil build:

```text
build/app/outputs/flutter-apk/app-release.apk
```

## Distribusi Internal ke Siswa/Pengawas

Langkah yang disarankan:

1. simpan APK rilis pada penyimpanan internal sekolah atau bagikan lewat tautan resmi sekolah
2. gunakan satu `API_BASE_URL` yang sama untuk satu gelombang uji coba
3. minta siswa memasang APK sebelum hari ujian
4. uji login token dan heartbeat minimal dengan 2-3 perangkat berbeda
5. siapkan fallback jaringan dan pengawas yang memantau status sinkron

Catatan:

- karena BYOD, beberapa vendor Android bisa membatasi audio, jaringan latar, atau agresif membunuh aplikasi
- jika perangkat sering gagal sinkron, peserta harus diminta tetap di layar ujian dan pengawas memeriksa status koneksi sebelum submit

## Paket Payload yang Sudah Didukung

- `question_text`
- `stem_html`
- `stimulus_html`
- `stem_media_url`
- `stimulus_media_url`
- `stem_audio_url`
- `stimulus_audio_url`

Media saat ini masih ringan:

- gambar remote sederhana
- audio remote sederhana

Belum ada:

- cache media offline penuh
- video
- mode kiosk Android terkelola

## Dokumen Operasional Terkait

- [RELEASE_CHECKLIST.md](./RELEASE_CHECKLIST.md)
- [OPERATOR_QUICKSTART.md](./OPERATOR_QUICKSTART.md)
- [BYOD_TRIAL_PROCEDURE.md](./BYOD_TRIAL_PROCEDURE.md)
- [DEVICE_TEST_MATRIX.md](./DEVICE_TEST_MATRIX.md)
- [../../docs/exam-api.md](../../docs/exam-api.md)
- [../../docs/cbt-smoke-checklist.md](../../docs/cbt-smoke-checklist.md)
