# Flutter Exam API Documentation

Dokumentasi ini merinci API yang digunakan oleh portal siswa (Flutter) untuk mengerjakan ujian CBT.

## Status Kontrak Saat Ini

Sinkron per 2026-05-08:

- Flutter memakai token ujian, bukan JWT admin.
- Token ujian backend memakai hex acak kuat dan tidak boleh disisipkan ke URL media.
- `X-Device-Fingerprint` adalah telemetry/resume hint BYOD, bukan bukti identitas perangkat yang kuat.
- Response exam tetap dibungkus dalam envelope `data`.
- Error semantics yang penting untuk mobile harus stabil: `404` token tidak ditemukan, `403` sesi tidak aktif/waktu tertutup, `409` device mismatch atau sudah submitted, dan `401` jika konteks peserta hilang.
- Endpoint `answer` dan `submit` harus menjaga response sukses `data.status = recorded|submitted`.
- Endpoint `heartbeat` dan `event` harus menjaga response sukses `data.status = ok|recorded`.
- Request JSON runtime harus berupa satu JSON object tanpa trailing payload. Body terlalu besar dikembalikan sebagai `413 request body too large`.
- Perubahan payload wajib direview dengan `docs/exam-payload-release-template.md`.
- Fase 1 integrasi proposal mengunci endpoint runtime Flutter tetap di `/api/exam/*`; Web Admin tidak membuat public route tree `/api/cbt/**` baru untuk runtime siswa.

## Batas Body Request

| Endpoint | Batas |
| --- | --- |
| `POST /api/exam/login` | 4 KiB serialized JSON |
| `POST /api/exam/event` | 16 KiB serialized JSON |
| `POST /api/exam/answer` | 64 KiB serialized JSON |
| `POST /api/exam/heartbeat` | empty body atau `{}` saja, maksimum 1 KiB |
| `POST /api/exam/submit` | empty body atau `{}` saja, maksimum 1 KiB |

Body yang melewati batas di atas mengembalikan `413` dengan error generic `request body too large`. Flutter harus menjaga limit jawaban uraian tetap aman di bawah budget serialized JSON endpoint `answer`.

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
  "device_fingerprint": "android:student-phone:install-..."
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
  - `403 Forbidden`: Sesi ujian belum aktif, belum mulai, atau window sudah tertutup.
  - `409 Conflict`: Token sudah terikat dengan perangkat lain atau ujian sudah submitted.
  - `413 Payload Too Large`: Body login melewati 4 KiB.
  - `500 Internal Server Error`: Error internal dengan pesan generic `internal server error`.

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
- **Error Responses:**
  - `401 Unauthorized`: Header token/fingerprint belum membentuk konteks peserta yang sah, termasuk fingerprint kosong atau belum terikat setelah login valid.
  - `409 Conflict`: Token sudah terikat ke fingerprint perangkat lain.
  - `500 Internal Server Error`: Error internal dengan pesan generic `internal server error`.

---

## 3. Heartbeat
Wajib dipanggil secara berkala (misal setiap 30-60 detik) untuk menandakan siswa masih aktif di aplikasi.

- **Endpoint:** `POST /api/exam/heartbeat`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Response:** `200 OK`
- **Body:** kosong atau `{}`.
- **Success Envelope:**
```json
{
  "data": { "status": "ok" }
}
```
- **Error Responses:**
  - `400 Bad Request`: Body bukan empty object.
  - `401 Unauthorized`: Header token/fingerprint belum membentuk konteks peserta yang sah.
  - `409 Conflict`: Token sudah terikat ke fingerprint perangkat lain.
  - `429 Too Many Requests`: Heartbeat terlalu sering untuk peserta/action yang sama.
  - `413 Payload Too Large`: Body heartbeat melewati 1 KiB.
  - `500 Internal Server Error`: Error internal dengan pesan generic `internal server error`.

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
- **Error Responses:**
  - `400 Bad Request`: JSON tidak valid atau `event_type` kosong.
  - `401 Unauthorized`: Header token/fingerprint belum membentuk konteks peserta yang sah.
  - `409 Conflict`: Token sudah terikat ke fingerprint perangkat lain.
  - `429 Too Many Requests`: Event terlalu sering untuk peserta/action yang sama.
  - `413 Payload Too Large`: Body event melewati 16 KiB.
  - `500 Internal Server Error`: Error internal dengan pesan generic `internal server error`.

### Taxonomy Event BYOD Flutter

Backend menerima event client sebagai evidence proctor/audit dan tetap client-safe. Event data tidak boleh berisi password, token ujian mentah di nested payload, answer key, credential admin, atau data perangkat yang lebih sensitif dari fingerprint telemetry yang sudah disepakati.

Payload event umum:

```json
{
  "event_type": "warning",
  "data": {
    "reason": "stale_connection_attention",
    "seconds_since_last_contact": 95,
    "failure_count": 3
  }
}
```

Taxonomy yang saat ini dipakai atau disiapkan untuk Flutter BYOD:

| `event_type` | `data.reason` / data utama | Tujuan operasional |
|--------------|----------------------------|--------------------|
| `app_switch` | `state`, `device_fingerprint` | App meninggalkan/berubah lifecycle dari permukaan ujian. |
| `warning` | `resume_exam` | Resume gate berjalan setelah app kembali foreground. |
| `warning` | `repeat_resume_attempt` | Resume terjadi lebih dari sekali dan perlu korelasi pengawas. |
| `warning` | `answer_saved_local_only` | Jawaban tersimpan lokal karena sync gagal. |
| `warning` | `submit_blocked_pending_sync` | Submit manual ditahan sampai pending answer sync aman. |
| `warning` | `auto_submit_blocked_pending_sync` | Auto-submit waktu habis tertahan karena pending answer belum aman. |
| `warning` | `submit_blocked_degraded_mode` | Submit manual ditahan karena koneksi menurun. |
| `warning` | `degraded_mode_entered` | Perangkat memasuki mode koneksi menurun. |
| `warning` | `stale_connection_attention` | Kontak server stale dan pengawas perlu memperhatikan. |
| `warning` | `stale_connection_escalated` | Stale sudah urgent dan butuh intervensi pengawas/proktor. |
| `warning` | `back_button_attempt` | Siswa mencoba tombol kembali saat ujian berjalan. |
| `warning` | `manual_submit` | Submit manual berhasil dikirim. |
| `screenshot_attempt` | platform signal bila tersedia | Sinyal deterrence; bukan bukti lengkap pada BYOD. |

Event eksplisit seperti `heartbeat_failed`, `heartbeat_recovered`, `restore_attempted`, dan `restore_failed` boleh ditambahkan secara backward-compatible setelah Flutter mengirim payload yang stabil.

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
  - `400 Bad Request`: JSON tidak valid, `question_id` malformed, atau soal bukan bagian dari ujian peserta.
  - `401 Unauthorized`: Header token/fingerprint belum membentuk konteks peserta yang sah.
  - `403 Forbidden`: Sesi belum mulai atau waktu ujian sudah habis.
  - `409 Conflict`: Ujian sudah disubmit sebelumnya.
  - `429 Too Many Requests`: Save answer terlalu sering untuk peserta/action yang sama.
  - `413 Payload Too Large`: Body answer melewati 64 KiB serialized JSON.
  - `500 Internal Server Error`: Error internal dengan pesan generic `internal server error`.
- **Success Envelope:**
```json
{
  "data": { "status": "recorded" }
}
```

### Format Jawaban Soal

Flutter harus mengirim format jawaban yang selaras dengan Web Admin/Core API:

| `question_type` | Format `answer` |
| --- | --- |
| `multiple_choice` | satu label opsi, contoh `B` |
| `multiple_answer` | label dipisah koma dan disortir deterministik, contoh `A,C` |
| `ordering` | label dipisah koma dalam urutan final siswa, contoh `B,A,C` |
| `matching` | pasangan kiri=kanan dipisah titik koma, contoh `A=1;B=2` |
| `short_answer` | teks pendek siswa |
| `essay` | teks uraian siswa, tetap di bawah budget 64 KiB serialized JSON |
| `true_false` | fixed pair `A=Benar`, `B=Salah`; backend scoring masih toleran terhadap cache lama `true`/`false` |
| `agree_disagree` | fixed pair `A=Setuju`, `B=Tidak Setuju` |

`hotspot`, `upload_answer`, dan `file_upload` tidak boleh dikirim sebagai jawaban palsu. Flutter menampilkan guard pengawas/manual sampai ada desain schema/API/storage/scoring yang disetujui.

---

## 6. Submit Ujian
Finalisasi pengerjaan. Setelah ini, token tidak bisa digunakan lagi untuk menjawab.

- **Endpoint:** `POST /api/exam/submit`
- **Headers:** `X-Exam-Token`, `X-Device-Fingerprint`
- **Body:** kosong atau `{}`.
- **Response:** `200 OK`
- **Success Envelope:**
```json
{
  "data": { "status": "submitted" }
}
```
- **Error Responses:**
  - `400 Bad Request`: Body bukan empty object.
  - `401 Unauthorized`: Header token/fingerprint belum membentuk konteks peserta yang sah.
  - `403 Forbidden`: Sesi belum mulai atau waktu ujian sudah tertutup.
  - `409 Conflict`: Ujian sudah disubmit sebelumnya.
  - `429 Too Many Requests`: Submit terlalu sering untuk peserta/action yang sama.
  - `413 Payload Too Large`: Body submit melewati 1 KiB.
  - `500 Internal Server Error`: Error internal dengan pesan generic `internal server error`.

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
- [ ] request body caps tetap selaras dengan Flutter, terutama batas serialized JSON `answer` 64 KiB, cap uraian mobile konservatif 15.000 karakter, dan guard exact serialized UTF-8 body di `ExamApiClient`
- [ ] fixed-pair `true_false` tetap `A=Benar` / `B=Salah` dan `agree_disagree` tetap `A=Setuju` / `B=Tidak Setuju`
- [ ] hotspot, upload/file answer, video prompt, dan recording answer tidak diklaim sebagai runtime sebelum desain/policy serta tes aman tersedia
- [ ] token/kunci jawaban tidak bocor melalui payload siswa atau URL media

Untuk rilis yang lebih formal, gunakan template:

- [docs/exam-payload-release-template.md](./exam-payload-release-template.md)
