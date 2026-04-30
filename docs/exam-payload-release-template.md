# Exam Payload Release Template

Gunakan template ini setiap kali backend Go mengubah kontrak payload endpoint exam yang dipakai aplikasi Flutter.

## Ringkasan Perubahan

- tanggal rilis:
- branch / commit:
- endpoint terdampak:
  - `POST /api/exam/login`
  - `GET /api/exam/status`
  - `POST /api/exam/heartbeat`
  - `POST /api/exam/event`
  - `POST /api/exam/answer`
  - `POST /api/exam/submit`

## Tujuan Perubahan

Jelaskan singkat:

- kenapa payload perlu diubah
- apakah ini menambah field baru, mengubah field lama, atau menghapus field lama
- apakah ini wajib untuk rilis sekarang atau masih bisa ditunda

## Perubahan Field

### Field baru

- nama field:
- endpoint:
- tipe data:
- contoh nilai:
- dipakai mobile untuk apa:
- fallback jika field belum ada:

### Field berubah

- nama field lama:
- nama field baru:
- endpoint:
- perubahan tipe / arti:
- dampak ke mobile:
- apakah ada compatibility bridge:

### Field dihapus

- nama field:
- endpoint:
- alasan penghapusan:
- bukti bahwa mobile tidak lagi bergantung pada field ini:

## Dampak ke Flutter

Checklist:

- [ ] login screen aman
- [ ] restore snapshot aman
- [ ] render soal aman
- [ ] media image/audio aman
- [ ] progress dan countdown aman
- [ ] submit guard aman
- [ ] event warning semantics tetap masuk akal

## Strategi Kompatibilitas

- [ ] perubahan additive saja
- [ ] field lama dipertahankan sementara
- [ ] fallback plain text tetap ada
- [ ] string kosong / omit tetap aman diparse
- [ ] URL media tetap valid untuk mobile

Jelaskan jika ada catatan khusus:

- …

## Verifikasi Sebelum Rilis

- [ ] review terhadap [docs/exam-api.md](./exam-api.md)
- [ ] review terhadap `apps/mobile/lib/src/models.dart`
- [ ] review terhadap `apps/mobile/lib/src/exam_api.dart`
- [ ] `flutter analyze`
- [ ] `flutter test`
- [ ] uji manual login token
- [ ] uji manual render soal
- [ ] uji manual restore
- [ ] uji manual submit

## Catatan Rollout

- apakah perlu update APK lebih dulu:
- apakah backend bisa dirilis tanpa memaksa update mobile:
- apakah perlu pemberitahuan ke operator/pengawas:

## Keputusan Akhir

- [ ] aman dirilis tanpa update mobile
- [ ] aman dirilis, tetapi perlu uji lapangan terbatas
- [ ] tunda rilis sampai app mobile disesuaikan
