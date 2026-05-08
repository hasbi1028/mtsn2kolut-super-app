# Release Checklist — Flutter CBT BYOD

Checklist ini untuk operator sekolah saat menyiapkan APK Android internal bagi siswa BYOD dan menjalankan Phase 5 rehearsal.

Status: sinkron Phase 30 Mobile RC Build and Release Package per 2026-05-08. Gunakan bersama `docs/exam-api.md`, `docs/cbt-proposal-integration-phase-5.md`, `docs/cbt-proposal-integration-phase-13.md`, `docs/cbt-proposal-integration-phase-23-26.md`, `docs/cbt-proposal-integration-phase-27-30.md`, `docs/cbt-smoke-checklist.md`, dan `docs/cbt-operator-runbook.md`.

## Boundary Phase 5

- [ ] Tidak deploy, tidak PM2 restart, dan tidak mengganti runtime dari checklist mobile ini.
- [ ] Tidak menjalankan `make db-migrate`, migration live, atau ad hoc SQL dari checklist mobile ini.
- [ ] Tidak membuat public SvelteKit route tree `/api/cbt/**` baru untuk runtime siswa.
- [ ] Tidak membawa PocketBase, SQLite, atau Alpine menjadi runtime CBT.
- [ ] Flutter berbicara langsung ke `services/core-api` melalui `/api/exam/*`, bukan melalui SvelteKit BFF.
- [ ] BYOD tidak setara kiosk penuh; device-owner bukan baseline untuk perangkat siswa pribadi.

## Phase 13 Mobile Release Candidate and Device Matrix

Isi bagian ini untuk APK yang benar-benar dipasang pada perangkat uji.

- [ ] RC identifier:
- [ ] APK SHA-256 hash:
- [ ] signing mode: release keystore / debug signing untuk uji teknis internal saja.
- [ ] `API_BASE_URL`:
- [ ] Flutter SDK absolute path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- [ ] minimum two Android vendors dicatat di `apps/mobile/DEVICE_TEST_MATRIX.md`.
- [ ] background/resume diuji dan melewati resume/status gate.
- [ ] heartbeat memperbarui last-contact atau menampilkan warning terkendali.
- [ ] pending answer tetap tersimpan saat koneksi terganggu.
- [ ] submit guard menahan submit saat pending sync/degraded mode belum pulih.
- [ ] device mismatch menghasilkan guidance `409` terkendali.
- [ ] screenshot protection / `FLAG_SECURE` diverifikasi sebagai deterrence/evidence, bukan kiosk guarantee.
- [ ] network disturbance menghasilkan status yang dipahami siswa/pengawas.

## Phase 24 Anti-Cheat BYOD Evidence Completion

- [ ] Manual evidence status masih `pending_manual_evidence` sampai perangkat Android nyata diuji.
- [ ] No fabricated real-device PASS; hasil lulus hanya boleh ditulis setelah operator/pengawas menguji perangkat fisik.
- [ ] Tidak ada klaim real-device PASS dari emulator, template, atau dokumentasi saja.
- [ ] Minimal dua vendor Android nyata diuji oleh operator/pengawas.
- [ ] `FLAG_SECURE` screenshot/recent-preview deterrence dicatat sebagai bukti terbatas, bukan kiosk guarantee.
- [ ] app switch event, resume gate, heartbeat loss, pending answer recovery, manual submit guard, device mismatch `409`, stale connection warning, dan final submit koneksi sehat tercatat.
- [ ] RC identifier, signing mode, `API_BASE_URL`, dan APK SHA-256 hash sama antara checklist, device matrix, dan evidence bundle.

## Phase 30 Mobile RC Build and Release Package

- [ ] RC identifier:
- [ ] commit hash:
- [ ] version name/code:
- [ ] APK path:
- [ ] APK file size:
- [ ] APK SHA-256 hash:
- [ ] signing status: release keystore / debug signing untuk uji teknis internal saja / unknown.
- [ ] `API_BASE_URL`:
- [ ] Flutter SDK path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- [ ] Flutter version:
- [ ] `flutter analyze`: pass / fail / skipped.
- [ ] `flutter test`: pass / fail / skipped.
- [ ] `flutter build apk --release`: pass / fail / skipped.
- [ ] `sha256sum build/app/outputs/flutter-apk/app-release.apk`: pass / fail / skipped.
- [ ] Android manifest `INTERNET` permission: pass / fail.
- [ ] Android `android:allowBackup="false"`: pass / fail.
- [ ] real-device PASS claimed only after physical Android operator test.
- [ ] minimum two Android vendors tetap `pending_manual_evidence` sampai perangkat nyata diuji.

## Sebelum Build

- [ ] backend Go sudah aktif dan endpoint CBT login peserta siap
- [ ] `API_BASE_URL` final untuk gelombang uji coba sudah dipastikan dan memakai HTTPS untuk staging/produksi
- [ ] token ujian dan sesi uji tersedia untuk minimal 2-3 siswa percobaan
- [ ] token ujian yang dipakai adalah token backend terbaru, umumnya 32 karakter hex pada kartu ujian
- [ ] pengawas paham bahwa aplikasi BYOD tidak setara kiosk penuh
- [ ] backend/frontend yang akan dipakai sudah tercatat commit-nya
- [ ] jika token diregenerasi/direpair eksplisit atau room/seat berubah, kartu ujian terdampak sudah dicetak ulang; migration 060 saja tidak merotasi token
- [ ] permission Android `INTERNET` terverifikasi pada manifest/APK
- [ ] release signing disiapkan melalui `android/key.properties` lokal atau environment variable, tanpa commit file keystore/password
- [ ] Android backup tetap nonaktif (`android:allowBackup="false"`) karena snapshot ujian memuat data sensitif

## Quality Gate

Jalankan dari `apps/mobile`:

```bash
flutter pub get
dart format lib test
flutter analyze
flutter test
```

Semua harus hijau sebelum build APK.

Flutter SDK bisa tidak tersedia di host yang menjalankan checklist dokumentasi atau rehearsal. Catat sebagai blocker validasi mobile pada host tersebut, lalu jalankan quality gate di mesin operator/CI yang memiliki Flutter SDK sebelum APK dipakai untuk ujian resmi.

## Build APK

```bash
flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
```

Hash APK yang benar-benar diuji:

```bash
sha256sum build/app/outputs/flutter-apk/app-release.apk
```

Shortcut dari root repo:

```bash
make check-mobile
make test-mobile
make mobile-release-apk API_BASE_URL=https://api.sekolah.example
```

Catatan signing:

- `android/key.properties` tidak boleh masuk Git dan sudah diabaikan oleh `.gitignore`
- alternatif CI/operator: `ANDROID_KEYSTORE_PATH`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`
- jika signing belum dikonfigurasi, build lokal tetap berjalan dengan debug signing, tetapi APK itu hanya untuk uji teknis internal dan bukan rilis ujian resmi

Output utama:

```text
build/app/outputs/flutter-apk/app-release.apk
```

## Verifikasi Lapangan Minimal

- [ ] install APK di minimal 2 vendor Android berbeda sebagai perangkat nyata Phase 5 rehearsal
- [ ] minimum two Android vendors terisi lengkap dengan RC identifier, APK SHA-256 hash, signing mode, dan `API_BASE_URL`
- [ ] hasil perangkat nyata dicatat di `DEVICE_TEST_MATRIX.md`, termasuk vendor, OS, koneksi, restore, dan submit
- [ ] login token berhasil
- [ ] skenario ringkas Phase 5 rehearsal tercatat: login token, heartbeat, answer save, restore, network disturbance, warning, submit
- [ ] skenario Phase 13 tercatat: background/resume, heartbeat, pending answer, submit guard, device mismatch, screenshot protection / `FLAG_SECURE`, network disturbance
- [ ] login dengan token 32 karakter dari kartu ujian berhasil
- [ ] jawaban pilihan ganda tersimpan
- [ ] jawaban uraian tersimpan
- [ ] heartbeat tidak gagal terus-menerus
- [ ] restore sesi bekerja setelah app ditutup/buka lagi
- [ ] restore sesi lama melewati resume gate dan status check sebelum siswa lanjut
- [ ] submit berhasil saat koneksi stabil
- [ ] panel warning muncul saat jaringan dimatikan sementara
- [ ] event BYOD yang dikirim app tidak berisi token ujian mentah, password, atau answer key
- [ ] batas jawaban uraian tetap aman di bawah budget serialized JSON 64 KiB endpoint `answer`
- [ ] smoke admin/guru di `docs/cbt-smoke-checklist.md` tidak menemukan kebocoran token/kunci jawaban
- [ ] Flutter tetap memakai `/api/exam/*`; tidak ada runtime siswa melalui route login/status di namespace `/api/cbt`

## Distribusi Internal

- [ ] bagikan APK hanya lewat kanal resmi sekolah
- [ ] satu gelombang uji memakai satu `API_BASE_URL` yang sama
- [ ] siswa diminta memasang APK sebelum hari ujian
- [ ] siswa diberi instruksi untuk tidak mengganti perangkat di tengah sesi
- [ ] pengawas tahu arti status: `Tersambung`, `Sinkron`, `Cek Ulang`, `Lokal`, `Gangguan`, `Waspada`, dan `Menurun`

## Catatan Operasional BYOD

- `Menurun` berarti sinkron gagal berulang; submit manual memang ditahan
- jika ada jawaban lokal menunggu sinkron, peserta tetap harus berada di layar ujian
- beberapa vendor Android agresif mematikan koneksi latar; pengawas perlu memeriksa kasus per perangkat
- `FLAG_SECURE` hanya mengurangi screenshot/recent preview, bukan jaminan anti-cheat penuh
- fingerprint perangkat hanya telemetry/resume hint, bukan identitas kuat
