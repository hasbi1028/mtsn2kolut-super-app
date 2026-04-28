# Current Work Plan

This file is a live handoff for the next agent or Claude Code session.

## Current Objective

Sprint 1 — Arsitektur CBT Opsi B: **Event + Rooms + Token + Exam API untuk Flutter**.

> **Keputusan arsitektur final (2026-04-28):**
> - Portal siswa (mengerjakan ujian) = **Flutter Android** — bukan SvelteKit.
> - Flutter belum tersedia → Sprint 1 hanya backend API + admin UI.
> - Sprint 1b (Flutter app) dikerjakan setelah Flutter tersedia.
> - Arsitektur CBT menggunakan **Opsi B**: cbt_exam_events + cbt_exam_rooms sebagai tabel terpisah.
> - `event_id` nullable di sesi → ulangan harian (oleh guru mapel) tidak perlu buat Event.
> - Ruangan punya kapasitas wajib, FK dari peserta ke ruangan → sistem bisa validasi kapasitas.

---

## What Is Already Done

### Repository and Architecture

- Monorepo: `apps/web-admin`, `services/core-api`, `services/pusaka-worker`
- Deploy contract, docs, PM2 configs tersedia

### Project Policy Files

- `AGENTS.md`, `CLAUDE.md`, `docs/architecture.md`, `docs/ui-guidelines.md`, `docs/deployment.md`
- Service policy: `apps/web-admin/AGENTS.md`, `services/core-api/AGENTS.md`, `services/pusaka-worker/AGENTS.md`

### Backend — COMPLETE ✅

- Migration 001–005 tersedia dan sudah diaplikasikan ke PostgreSQL
- sqlc generated, `go build ./...` dan `go test ./...` pass bersih

**Semua API endpoints yang sudah ada:**
- `/api/academic` (GET/POST/DELETE)
- `/api/students` (GET/POST/DELETE)
- `/api/employees` (GET/POST/PUT/DELETE + pusaka-status + update-pusaka)
- `/api/cbt/questions` (GET/POST/DELETE)
- `/api/cbt/packages` (GET/POST/DELETE)
- `/api/cbt/sessions` (GET/POST)
- `/api/cbt/sessions/{id}` (GET/DELETE)
- `/api/cbt/sessions/{id}/status` (PATCH)
- `/api/cbt/sessions/{id}/participants` (GET)
- `/api/cbt/sessions/{id}/enroll` (POST — enroll satu kelas, token masih kosong)
- `/api/cbt/sessions/{id}/score` (POST — hitung skor)
- `/api/cbt/sessions/{id}/results` (GET — skor + info per peserta)
- `/api/cbt/sessions/{id}/participants/{pid}/answer` (POST)
- `/api/cbt/sessions/{id}/participants/{pid}/answers` (GET)
- `/api/jobs` + run-all, cancel, cancel-all, stats
- `/api/attendance`, `/api/schedules`, `/api/settings`
- `/api/worker/*` (claim, heartbeat, complete, fail, config, status)
- `/api/scheduler/tick`
- `/api/auth/*` (login, refresh, change-password)
- `/health`

### Frontend — COMPLETE ✅

**UI Stack:**
- Tailwind CSS v4 (`@tailwindcss/vite`) + shadcn-svelte nova
- Tema hijau institusional (`--color-primary: oklch(0.38 0.13 145)`)
- Sidebar dengan groups, icons, mobile support
- `dialog/index.ts` diekspor dengan namespace (Root/Content/Header/Title/Description/Footer/Trigger)

**Semua halaman sudah shadcn + tema hijau:**
- `/` — Dashboard operasional
- `/employees` — Manajemen pegawai + dialog Pusaka + konfirmasi SURE
- `/jobs` — Riwayat job dengan filter + auto-refresh
- `/attendance` — Rekap absensi harian + status worker
- `/attendance/queue` — Antrian job absensi
- `/settings` — Worker settings, jadwal otomatis, ubah password
- `/academic` — Tahun ajaran, kelas, mata pelajaran
- `/students` — Daftar siswa
- `/cbt/questions` — Bank soal
- `/cbt/packages` — Paket ujian
- `/cbt/sessions` — Sesi ujian (create, schedule, activate, enroll, finish)
- `/cbt/sessions/[id]` — Hasil: skor per peserta, stat kelulusan, ekspor CSV

---

## Arsitektur CBT — Opsi B (Target)

### Schema Baru (Migration 006)

```
cbt_exam_events                       ← BARU: Kegiatan ujian
  id, title, exam_type, scope,
  academic_year_id, status

cbt_exam_sessions                     ← DIUBAH: event_id nullable, class_id nullable
  id, event_id (FK nullable),
  package_id, class_id (nullable),
  title, scheduled_start, scheduled_end, status

cbt_exam_rooms                        ← BARU: Ruangan per sesi
  id, session_id (FK), room_name, capacity

cbt_exam_participants                 ← DIUBAH: room_id nullable, token di-generate saat enroll
  id, session_id, student_id,
  room_id (FK nullable → cbt_exam_rooms),
  token (auto-generate saat enroll),
  joined_at, submitted_at, score
```

### Skenario yang Didukung

| Skenario | Dibuat Oleh | Alur |
|---|---|---|
| **Ulangan Harian** | Guru mapel | Sesi mandiri (event_id NULL) → enroll 1 kelas → mulai |
| **UTS / UAS** | Admin/Wakakur | Buat Event → buat sesi per mapel → enroll per tingkat → buat ruangan → shuffle → mulai |
| **UAM** | Admin | Buat Event → sesi → enroll seluruh sekolah → ruangan → shuffle → mulai |
| **Try Out** | Admin/Guru | Fleksibel, bisa dengan atau tanpa Event |

### Alur CBT Admin Lengkap (Opsi B)

```
[Opsional] Buat Kegiatan (Event) → UTS/UAS/UAM
  ↓
Buat Sesi → pilih paket soal, waktu mulai/selesai
  ↓
Enroll Peserta → by kelas / by tingkat / by sekolah
  ↓
Generate Token → otomatis saat enroll, bisa regenerate
  ↓
[Opsional] Buat Ruangan → nama + kapasitas
  ↓
[Opsional] Shuffle Peserta → acak siswa ke ruangan
  ↓
Aktifkan Sesi → status: active
  ↓
Siswa login via Flutter (token) → kerjakan soal → submit
  ↓
Selesaikan Sesi → hitung skor → lihat hasil
```

---

## Sprint 1 — Opsi B Foundation ← SEKARANG

### Langkah 1: Migration 006

File: `services/core-api/db/migrations/006_cbt_events_rooms_anticheat.sql`

```sql
CREATE TYPE cbt_exam_type AS ENUM ('ulangan','uts','uas','uam','tryout','lainnya');

CREATE TABLE cbt_exam_events (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title            TEXT NOT NULL,
  exam_type        cbt_exam_type NOT NULL DEFAULT 'lainnya',
  scope            TEXT NOT NULL DEFAULT 'class',  -- 'class'|'grade'|'school'
  academic_year_id UUID REFERENCES academic_years(id) ON DELETE SET NULL,
  status           TEXT NOT NULL DEFAULT 'draft',  -- draft|active|finished
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cbt_exam_rooms (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id  UUID NOT NULL REFERENCES cbt_exam_sessions(id) ON DELETE CASCADE,
  room_name   TEXT NOT NULL,
  capacity    INT  NOT NULL DEFAULT 30,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Log aktivitas peserta (basis audit anti-cheat)
CREATE TABLE cbt_participant_events (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  participant_id  UUID NOT NULL REFERENCES cbt_exam_participants(id) ON DELETE CASCADE,
  event_type      TEXT NOT NULL,  -- login|heartbeat|answer|submit|app_switch|screenshot_attempt|warning
  event_data      JSONB,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE cbt_exam_sessions
  ADD COLUMN event_id UUID REFERENCES cbt_exam_events(id) ON DELETE SET NULL,
  ALTER COLUMN class_id DROP NOT NULL;

ALTER TABLE cbt_exam_participants
  ADD COLUMN room_id            UUID REFERENCES cbt_exam_rooms(id) ON DELETE SET NULL,
  ADD COLUMN device_fingerprint TEXT,
  ADD COLUMN question_order     JSONB,       -- ["uuid-q5","uuid-q1",...] — urutan soal acak per peserta
  ADD COLUMN last_heartbeat     TIMESTAMPTZ,
  ADD COLUMN app_switch_count   INT NOT NULL DEFAULT 0,
  ADD COLUMN screenshot_attempt INT NOT NULL DEFAULT 0,
  ADD COLUMN login_ip           TEXT,
  ADD COLUMN suspicious_flag    BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_cbt_events_status          ON cbt_exam_events (status);
CREATE INDEX idx_cbt_rooms_session          ON cbt_exam_rooms (session_id);
CREATE INDEX idx_cbt_participants_room      ON cbt_exam_participants (room_id);
CREATE INDEX idx_participant_events_pid     ON cbt_participant_events (participant_id, created_at DESC);
CREATE INDEX idx_participant_events_type    ON cbt_participant_events (event_type, created_at DESC);
CREATE UNIQUE INDEX idx_cbt_rooms_name      ON cbt_exam_rooms (session_id, room_name);
```

### Langkah 2: SQL Queries Baru

File: `services/core-api/db/queries/cbt_events.sql`
- `ListCbtExamEvents`, `GetCbtExamEvent`, `CreateCbtExamEvent`, `UpdateCbtExamEventStatus`, `DeleteCbtExamEvent`

File: `services/core-api/db/queries/cbt_rooms.sql`
- `ListCbtExamRooms` (dengan jumlah peserta), `CreateCbtExamRoom`, `DeleteCbtExamRoom`

Update `cbt_sessions.sql`:
- `EnrollClassToSession` → generate token: `encode(gen_random_bytes(4), 'hex')`
- `EnrollGradeToSession` (BARU) — enroll semua siswa aktif satu tingkat (grade)
- `EnrollSchoolToSession` (BARU) — enroll semua siswa aktif
- `RegenerateParticipantToken` (BARU)
- `AssignParticipantRoom` (BARU)
- `ListParticipantsByRoom`
- `GetParticipantByToken` (BARU) — untuk ExamToken middleware
- `UpdateParticipantQuestionOrder` (BARU) — simpan urutan soal acak
- `UpdateParticipantHeartbeat` (BARU)
- `IncrementParticipantAppSwitch` (BARU)
- `IncrementParticipantScreenshot` (BARU)
- `SetParticipantSuspiciousFlag` (BARU)
- `InsertParticipantEvent` (BARU) — log aktivitas
- `ListParticipantEvents` (BARU)
- `GetSessionProctoringStatus` (BARU) — semua peserta dengan status live

### Langkah 3: sqlc generate

Setelah semua query ditambah, jalankan `sqlc generate` di `services/core-api`.

### Langkah 4: Service + Handler Baru

**Events:**
- `internal/service/cbt_event.go` — CRUD + status update
- `internal/handler/cbt_event.go` — thin HTTP handler

**Rooms:**
- Method baru di `internal/service/cbt_session.go` — CreateRoom, DeleteRoom, ShuffleRooms
- Method baru di `internal/handler/cbt_session.go` — ListRooms, CreateRoom, DeleteRoom, ShuffleRooms

**Token & Enrollment:**
- Update `EnrollClass` di service → query baru yang generate token
- Tambah `EnrollGrade`, `EnrollSchool`, `RegenerateToken` di service + handler

### Langkah 5: Exam API untuk Flutter

File baru: `internal/handler/exam.go`
File baru: `internal/service/exam.go`
File baru: `internal/middleware/exam_token.go`

**Endpoints:**
```
POST /api/exam/login      — body: {token, device_fingerprint} → return session+soal+question_order
GET  /api/exam/status     — header: X-Exam-Token → progres + sisa waktu server-side
POST /api/exam/heartbeat  — header: X-Exam-Token → update last_heartbeat
POST /api/exam/event      — header: X-Exam-Token, body: {event_type, data} → log app_switch/screenshot
POST /api/exam/answer     — header: X-Exam-Token, body: {question_id, answer} → validasi window waktu
POST /api/exam/submit     — header: X-Exam-Token → finalisasi + set submitted_at
```

**Response `POST /api/exam/login`:**
```json
{
  "participant_id": "uuid",
  "student": { "nis": "...", "nama": "..." },
  "session": {
    "id": "uuid", "title": "...",
    "scheduled_start": "...", "scheduled_end": "...",
    "duration_minutes": 90
  },
  "room": { "room_name": "Ruang 1" },
  "questions": [
    { "id": "uuid", "code": "M001", "question_text": "...",
      "option_a": "...", "option_b": "...", "option_c": "...",
      "option_d": "...", "option_e": "..." }
  ],
  "answered_count": 0,
  "total_questions": 30,
  "time_remaining_seconds": 5400
}
```

**Middleware ExamToken:**
- Baca header `X-Exam-Token`
- Query participant by token, validasi sesi `active`
- Inject `participant_id` + `session_id` ke context
- Return 401 jika token tidak valid, 403 jika sesi belum/sudah selesai

### Langkah 6: Update Router

Tambah ke `cmd/api/main.go`:

```
// Admin (JWT-protected)
GET/POST              /api/cbt/events
GET/PATCH/DELETE      /api/cbt/events/{id}
GET/POST              /api/cbt/sessions/{id}/rooms
DELETE                /api/cbt/rooms/{id}
POST                  /api/cbt/sessions/{id}/enroll-grade
POST                  /api/cbt/sessions/{id}/enroll-school
POST                  /api/cbt/sessions/{id}/shuffle-rooms
POST                  /api/cbt/sessions/{id}/generate-tokens
POST                  /api/cbt/sessions/{id}/participants/{pid}/regenerate-token
GET                   /api/cbt/sessions/{id}/proctoring
POST                  /api/cbt/sessions/{id}/participants/{pid}/flag

// Exam (ExamToken middleware — no JWT)
POST  /api/exam/login
GET   /api/exam/status
POST  /api/exam/heartbeat
POST  /api/exam/event
POST  /api/exam/answer
POST  /api/exam/submit
```

### Langkah 7: Frontend Admin (SvelteKit)

**Halaman baru:**
- `/cbt/events` — list kegiatan ujian (UTS/UAS/UAM/ulangan/tryout)
- `/cbt/events/[id]` — detail kegiatan: daftar sesi yang tergabung

**Update halaman existing:**
- `/cbt/sessions/[id]` — tambah 2 tab baru:
  - Tab "Ruangan" — daftar ruangan + kapasitas + jumlah peserta, tombol tambah/hapus ruangan, tombol Shuffle
  - Tab "Peserta" — tabel peserta dengan kolom token + ruangan, tombol Salin Token, tombol Regenerate Token
- `/cbt/sessions` (form buat sesi) — tambah pilihan event_id (optional)

**SvelteKit API proxies baru:**
- `/api/cbt/events` (GET, POST)
- `/api/cbt/events/[id]` (GET, PATCH, DELETE)
- `/api/cbt/sessions/[id]/rooms` (GET, POST)
- `/api/cbt/rooms/[id]` (DELETE)
- `/api/cbt/sessions/[id]/enroll-grade` (POST)
- `/api/cbt/sessions/[id]/enroll-school` (POST)
- `/api/cbt/sessions/[id]/shuffle-rooms` (POST)
- `/api/cbt/sessions/[id]/generate-tokens` (POST)
- `/api/cbt/sessions/[id]/participants/[pid]/regenerate-token` (POST)

---

## Sprint 1b — Flutter Portal Siswa (setelah Flutter tersedia)

1. Flutter app: layar input token → tampil soal ujian → jawab → submit
2. Konsumsi `POST /api/exam/login` → render soal
3. `POST /api/exam/answer` per jawaban (atau batch saat submit)
4. `POST /api/exam/submit` → layar konfirmasi selesai
5. `GET /api/exam/status` untuk sisa waktu dan progres

---

## Sprint 2 — Edit Master Data

1. Form edit inline / modal untuk Soal CBT
2. Form edit Siswa + dropdown assign kelas
3. Form edit Pegawai (nama, NIP, unit kerja)
4. Edit judul/deskripsi Paket Ujian + Event

---

## Sprint 3 — Laporan & Monitoring

1. Filter rentang tanggal + export CSV absensi
2. Live monitoring CBT — siapa sudah submit, siapa belum, sisa waktu, per ruangan
3. Rekap kehadiran per bulan per pegawai
4. Laporan hasil ujian per kegiatan (bukan hanya per sesi)

---

## Sprint 4 — Multi-user (Jangka Panjang)

1. Akun guru mapel dengan akses terbatas (kelas/mata pelajaran sendiri)
2. Roles: admin, guru
3. Migration baru untuk users/roles table

---

## Migration 007 — Tipe Soal Fleksibel (Bagian Sprint 1)

### Keputusan arsitektur soal (2026-04-28)

Soal CBT mendukung 5 tipe dengan skema yang fleksibel:

| Tipe | `question_type` | Opsi | Penilaian |
|---|---|---|---|
| Pilihan Ganda | `multiple_choice` | `options` JSONB (≥2, dinamis) | Otomatis |
| Benar/Salah | `true_false` | `options` JSONB (2 opsi tetap) | Otomatis |
| Pilih Banyak | `multiple_answer` | `options` JSONB (≥2) | Otomatis (exact set match) |
| Isian Singkat | `short_answer` | `options: []` kosong | Otomatis (exact) atau manual |
| Esai | `essay` | `options: []` kosong | Manual oleh guru |

**Format `options` JSONB:**
```json
[{"label": "A", "text": "Teks pilihan pertama"}, {"label": "B", "text": "..."}]
```

**Format `answer_key`:**
- `multiple_choice`, `true_false`, `short_answer`: string tunggal — `"A"` atau `"Benar"` atau `"42"`
- `multiple_answer`: comma-separated — `"A,C"` atau `"A,B,D"`
- `essay`: kosong `""` — penilaian via `manual_score`

**Kolom lama `option_a..e` dipertahankan** (backward compat) tapi tidak dipakai untuk soal baru.

### Schema tambahan (Migration 007)

```sql
ALTER TABLE cbt_questions
  ADD COLUMN question_type TEXT NOT NULL DEFAULT 'multiple_choice',
  ADD COLUMN options        JSONB NOT NULL DEFAULT '[]';

ALTER TABLE cbt_student_answers
  ADD COLUMN manual_score  NUMERIC(5,2),
  ADD COLUMN graded_by     TEXT,
  ADD COLUMN graded_at     TIMESTAMPTZ;
```

### Endpoint baru untuk essay grading

```
GET  /api/cbt/sessions/{id}/ungraded-essays   — list jawaban essay yang belum dinilai
POST /api/cbt/sessions/{id}/participants/{pid}/grade-essay  — guru input manual_score per soal
```

### Scoring logic update

- `ScoreSession` otomatis: skip soal `essay`, kalkulasi berdasarkan soal yang bisa di-auto-grade
- Score akhir = (auto_correct + manual_correct_equivalent) / total_questions × 100
- Essay masuk kalkulasi hanya setelah semua manual_score terisi
- `ListUngradedEssays` untuk monitoring guru

## Known Gaps & Bugs

### Belum Ada (Fungsional)
- **Soal tipe baru** — question_type + options JSONB ← Migration 007 (Sprint 1 extended)
- **Essay grading UI** — admin/guru input nilai manual ← Sprint 1 extended
- **Edit master data** — hampir semua halaman hanya Create + Delete ← Sprint 2
- **Assign kelas ke siswa** — UI belum ada, meski `class_id` sudah ada di schema ← Sprint 2
- **Filter rentang tanggal absensi** ← Sprint 3
- **Export absensi** ← Sprint 3

### Risiko Teknis
- `option_a..e` lama masih ada di schema — tidak dipakai untuk soal baru, tapi belum dibersihkan
- Essay dalam sesi yang sama dengan PG akan delay finalisasi skor sampai semua essay dinilai
- Single admin account → belum ada multi-user/roles guru → Sprint 4

---

## Important Context

- Deploy order: backend code → migration → restart backend → deploy frontend → deploy worker
- Go clean architecture dan sqlc wajib dipertahankan
- Sidebar.svelte = nav utama; Nav.svelte tidak lagi dipakai di layout
- Semua halaman baru = light institutional green theme (`oklch(0.38 0.13 145)`)
- Jobs proxy (`/api/jobs/+server.ts`) rename `employee_nama` → `nama`, `employee_nip` → `nip`
- `vite.config.js` — `ssr: { noExternal: ['lucide-svelte', 'bits-ui', 'tailwind-variants'] }`
- `dialog/index.ts` — export namespace (Root/Content/…) DAN named (Dialog/DialogContent/…)
- Setelah sqlc generate, selalu `go build ./...` dan `go test ./...` sebelum commit
- Setelah Svelte edit, selalu `npm run check` di `apps/web-admin`
