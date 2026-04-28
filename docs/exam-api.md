# Flutter Exam API Documentation

Dokumentasi ini merinci API yang digunakan oleh portal siswa (Flutter) untuk mengerjakan ujian CBT.

## Base URL
`https://api-cbt.mtsn2kolut.sch.id` (Sesuaikan dengan environment)

## Autentikasi
Aplikasi Flutter tidak menggunakan JWT admin. Sebagai gantinya, autentikasi menggunakan **Token Ujian** (8-digit hex) yang didapat siswa dari kartu ujian atau pengawas.

Setelah login berhasil, semua request selanjutnya wajib menyertakan header:
`X-Exam-Token: [TOKEN_SISWA]`

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

---

## 2. Status Progres
Mengecek sisa waktu dan progres pengerjaan di server.

- **Endpoint:** `GET /api/exam/status`
- **Headers:** `X-Exam-Token`
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
- **Headers:** `X-Exam-Token`
- **Response:** `200 OK`

---

## 4. Record Event (Anti-Cheat)
Mencatat aktivitas mencurigakan atau perpindahan status aplikasi.

- **Endpoint:** `POST /api/exam/event`
- **Headers:** `X-Exam-Token`
- **Body:**
```json
{
  "event_type": "app_switch", 
  "data": { "reason": "user minimized app" }
}
```
- **Event Types:** `app_switch`, `screenshot_attempt`, `warning`.

---

## 5. Simpan Jawaban
Mengirim jawaban untuk satu soal. Panggil setiap kali siswa memilih/mengubah jawaban.

- **Endpoint:** `POST /api/exam/answer`
- **Headers:** `X-Exam-Token`
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

---

## 6. Submit Ujian
Finalisasi pengerjaan. Setelah ini, token tidak bisa digunakan lagi untuk menjawab.

- **Endpoint:** `POST /api/exam/submit`
- **Headers:** `X-Exam-Token`
- **Response:** `200 OK`
