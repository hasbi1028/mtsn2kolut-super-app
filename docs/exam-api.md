# Flutter Exam API Documentation

Dokumentasi ini merinci API yang digunakan oleh portal siswa (Flutter) untuk mengerjakan ujian CBT.

## Status Kontrak Saat Ini

Sinkron per 2026-05-03:

- Flutter memakai token ujian, bukan JWT admin.
- Token ujian backend memakai hex acak kuat dan tidak boleh disisipkan ke URL media.
- `X-Device-Fingerprint` adalah telemetry/resume hint BYOD, bukan bukti identitas perangkat yang kuat.
- Response exam tetap dibungkus dalam envelope `data`.
- Error semantics yang penting untuk mobile harus stabil: `404` token tidak ditemukan, `403` sesi tidak aktif/waktu tertutup, `409` device mismatch atau sudah submitted, dan `401` jika konteks peserta hilang.
- Endpoint `answer` dan `submit` harus menjaga response sukses `data.status = recorded|submitted`.
- Endpoint `heartbeat` dan `event` harus menjaga response sukses `data.status = ok|recorded`.
- Perubahan payload wajib direview dengan `docs/exam-payload-release-template.md`.

## Base URL
`https://api-cbt.mtsn2kolut.sch.id` (Sesuaikan dengan environment)

## Autentikasi
Aplikasi Flutter tidak menggunakan JWT admin. Sebagai gantinya, autentikasi menggunakan **Token Ujian** 128-bit hex yang didapat siswa dari kartu ujian atau pengawas.

Setelah login berhasil, semua request selanjutnya wajib menyertakan header:
`X-Exam-Token: [TOKEN_SISWA]`
`X-Device-Fingerprint: [FINGERPRINT_PERANGKAT]`

---

## 1. Login Ujian
Mendaftarkan perangkat dan mendapatkan data soal.

- **Endpoint:** `POST /api/exam/login`
- **Body:**
```json
{
  "token": "a1b2c3d4",
  "device_fingerprint": "unique-device-id-or-imei"
}
```
- **Success Response (200 OK):**
```json
{
  "data": {
    "participant_id": "uuid",
    "student": { "nis": "12345", "nama": "Ahmad Siswa" },
    "session": {
      "id": "uuid",
      "title": "Ujian Matematika Kelas VII",
      "scheduled_start": "2026-04-28T08:00:00Z",
      "scheduled_end": "2026-04-28T10:00:00Z",
      "duration_minutes": 90
    },
    "room": { "room_name": "Ruang 01" },
    "questions": [
      {
        "id": "uuid-q1",
        "question_text": "1 + 1 = ...",
        "options": [
          {"label": "A", "text": "1"},
          {"label": "B", "text": "2"},
          {"label": "C", "text": "3"},
          {"label": "D", "text": "4"}
        ]
      }
    ],
    "answered_count": 0,
    "total_questions": 30,
    "time_remaining_seconds": 5400
  }
}
```
- **Error Responses:**
  - `404 Not Found`: Token tidak ditemukan.
  - `403 Forbidden`: Sesi ujian belum aktif atau sudah berakhir.
  - `409 Conflict`: Token sudah terikat dengan perangkat lain.

### Payload Kompatibilitas Mobile

Backend sebaiknya menjaga kontrak payload login ini tetap stabil untuk aplikasi Flutter.

Field yang saat ini dipakai mobile:

- `student.nis`
- `student.nama`
- `session.id`
- `session.title`
- `session.scheduled_start`
- `session.scheduled_end`
- `session.duration_minutes`
- `room.room_name`
- `answered_count`
- `total_questions`
- `time_remaining_seconds`

Field soal yang saat ini aman dipakai mobile:

- `id`
- `question_text`
- `stem_html`
- `stimulus_html`
- `stem_media_url`
- `stimulus_media_url`
- `stem_audio_url`
- `stimulus_audio_url`
- `options[].label`
- `options[].text`

Checklist sebelum backend mengubah payload:

1. Jangan hapus field yang sudah dipakai mobile tanpa migration contract yang jelas.
2. Untuk field rich content baru, tetap sediakan fallback plain text bila memungkinkan.
3. URL media harus absolut atau konsisten dapat di-resolve oleh app, tanpa menyisipkan token ujian di query string.
4. Untuk soal tanpa media/audio, kirim string kosong atau omit dengan bentuk yang tetap aman diparse.
5. Jangan ubah arti `answered_count`, `total_questions`, dan `time_remaining_seconds` karena dipakai untuk restore, progress, dan submit guard.
6. Jika menambah jenis media baru, dokumentasikan dulu sebelum dianggap wajib didukung mobile.
7. Fetch media/audio CBT oleh peserta wajib memakai header `X-Exam-Token` dan `X-Device-Fingerprint`; jangan mengembalikan URL berisi `exam_token`.

---

## 2. Status Progres
Mengecek sisa waktu dan progres pengerjaan di server.

- **Endpoint:** `GET /api/exam/status`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Success Response (200 OK):**
```json
{
  "data": {
    "answered_count": 5,
    "total_questions": 30,
    "time_remaining_seconds": 4200,
    "is_submitted": false
  }
}
```

---

## 3. Heartbeat
Wajib dipanggil secara berkala (misal setiap 30-60 detik) untuk menandakan siswa masih aktif di aplikasi.

- **Endpoint:** `POST /api/exam/heartbeat`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Response:** `200 OK`
- **Success Envelope:**
```json
{
  "data": { "status": "ok" }
}
```

---

## 4. Record Event (Anti-Cheat)
Mencatat aktivitas mencurigakan atau perpindahan status aplikasi.

- **Endpoint:** `POST /api/exam/event`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Body:**
```json
{
  "event_type": "app_switch", 
  "data": { "reason": "user minimized app" }
}
```
- **Event Types:** `app_switch`, `screenshot_attempt`, `warning`.
- **Success Envelope:**
```json
{
  "data": { "status": "recorded" }
}
```

---

## 5. Simpan Jawaban
Mengirim jawaban untuk satu soal. Panggil setiap kali siswa memilih/mengubah jawaban.

- **Endpoint:** `POST /api/exam/answer`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Body:**
```json
{
  "question_id": "uuid-q1",
  "answer": "B"
}
```
- **Error Responses:**
  - `403 Forbidden`: Waktu ujian sudah habis.
  - `409 Conflict`: Ujian sudah disubmit sebelumnya.
- **Success Envelope:**
```json
{
  "data": { "status": "recorded" }
}
```

---

## 6. Submit Ujian
Finalisasi pengerjaan. Setelah ini, token tidak bisa digunakan lagi untuk menjawab.

- **Endpoint:** `POST /api/exam/submit`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Response:** `200 OK`
- **Success Envelope:**
```json
{
  "data": { "status": "submitted" }
}
```
- **Error Responses:**
  - `403 Forbidden`: Waktu ujian sudah tertutup.
  - `409 Conflict`: Ujian sudah disubmit sebelumnya.

---

## Checklist Backend Sebelum Rilis ke Mobile

Gunakan daftar ini saat mengubah endpoint exam agar app Flutter tidak diam-diam rusak:

- [ ] response `POST /api/exam/login` masih memuat field dasar siswa, sesi, ruang, dan progres
- [ ] bentuk `questions[]` tetap kompatibel dengan renderer PG/uraian
- [ ] media/image/audio baru tidak membuat app wajib mengunduh format yang belum didukung
- [ ] nilai `time_remaining_seconds` tetap akurat untuk countdown dan auto-submit
- [ ] perubahan event type/warning semantics tetap backward-compatible
- [ ] perubahan error code login/status/submit sudah ditinjau dampaknya ke restore flow
- [ ] `answer`, `submit`, `heartbeat`, dan `event` masih mengembalikan success envelope yang sama
- [ ] token/kunci jawaban tidak bocor melalui payload siswa atau URL media

Untuk rilis yang lebih formal, gunakan template:

- [docs/exam-payload-release-template.md](./exam-payload-release-template.md)
