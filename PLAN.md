# Current Work Plan

This file is a live handoff for the next agent or Claude Code session.

## Current Objective

CBT exam loop selesai di sisi admin. Langkah selanjutnya adalah student portal dan/atau polish UI.

---

## What Is Already Done

### Repository and Architecture

- Monorepo: `apps/web-admin`, `services/core-api`, `services/pusaka-worker`
- Deploy contract, docs, PM2 configs tersedia

### Project Policy Files

- `AGENTS.md`, `CLAUDE.md`, `docs/architecture.md`, `docs/ui-guidelines.md`, `docs/deployment.md`
- Service policy: `apps/web-admin/AGENTS.md`, `services/core-api/AGENTS.md`, `services/pusaka-worker/AGENTS.md`

### Backend — COMPLETE ✅

- Migration 004: academic foundation (years, classes, subjects, students, cbt_questions, cbt_packages)
- Migration 005: fix gender enum L/P, add 'archived' status, drop question code unique, add exam sessions tables
- sqlc generate — semua types generated
- `go build ./...` dan `go test ./...` pass bersih

**Semua API endpoints:**
- `/api/academic` (GET/POST/DELETE)
- `/api/students` (GET/POST/DELETE)
- `/api/cbt/questions` (GET/POST/DELETE)
- `/api/cbt/packages` (GET/POST/DELETE)
- `/api/cbt/sessions` (GET/POST)
- `/api/cbt/sessions/{id}` (GET)
- `/api/cbt/sessions/{id}/status` (PATCH)
- `/api/cbt/sessions/{id}` (DELETE — draft only)
- `/api/cbt/sessions/{id}/participants` (GET)
- `/api/cbt/sessions/{id}/enroll` (POST — enroll satu kelas)
- `/api/cbt/sessions/{id}/score` (POST — hitung skor)
- `/api/cbt/sessions/{id}/results` (GET — skor + info per peserta)
- `/api/cbt/sessions/{id}/participants/{pid}/answer` (POST — record jawaban)
- `/api/cbt/sessions/{id}/participants/{pid}/answers` (GET)

### Frontend — COMPLETE ✅

**UI Stack:**
- Tailwind CSS v4 + shadcn-svelte nova
- Sidebar navigation dengan groups, icons, mobile support
- Institutional light theme

**Halaman:**
- `/` — Dashboard (legacy dark)
- `/academic` — Tahun ajaran, kelas, mata pelajaran
- `/students` — Daftar siswa
- `/cbt/questions` — Bank soal
- `/cbt/packages` — Paket ujian
- `/cbt/sessions` — Sesi ujian (create, schedule, activate, enroll, finish)
- `/cbt/sessions/[id]` — Halaman hasil: skor per peserta, stat kelulusan, ekspor CSV, trigger scoring

---

## Alur CBT Admin Yang Sudah Bisa Digunakan

```
Buat mata pelajaran → buat kelas + tahun ajaran → input soal →
buat paket ujian → buat sesi → daftarkan siswa → mulai sesi →
selesaikan → hitung skor → lihat hasil → ekspor CSV
```

---

## Recommended Next Steps

### Step 1: Smoke Test End-to-End

1. Apply DB migrations: `004` lalu `005` (urutan penting)
2. Start backend: `cd services/core-api && go run ./cmd/api`
3. Start frontend: `cd apps/web-admin && npm run dev`
4. Test setiap halaman dari academic → sessions → results
5. Verifikasi status transitions dan ekspor CSV

### Step 2: Student-Facing Exam Interface (Next Big Slice)

Saat ini jawaban hanya bisa direkam via API langsung. Untuk CBT real:
1. Buat route `/exam/[token]` atau app terpisah untuk siswa
2. Endpoint login siswa dengan token dari `cbt_exam_participants.token`
3. Tampilkan soal dari paket yang terkait
4. Kirim jawaban ke `POST /api/cbt/sessions/{id}/participants/{pid}/answer`
5. Auto-submit saat waktu habis

### Step 3: Token Generation untuk Peserta

Saat ini field `token` di `cbt_exam_participants` masih kosong.
Tambahkan logika generate token unik saat `EnrollClass` dipanggil.

### Step 4: UI Polish (Prioritas Rendah)

- Skeleton loaders menggantikan teks "Memuat data..."
- Pagination di halaman siswa dan soal
- Bulk import siswa via CSV

### Step 5: Migrasi Halaman Legacy (Opsional)

Halaman: dashboard, employees, jobs, attendance, settings masih pakai dark CSS.
Migrasi ke Tailwind/shadcn bisa dilakukan bertahap — bukan blocker.

---

## Important Context

- Deploy order: migration 004 → 005 → restart backend → deploy frontend
- Go clean architecture dan sqlc wajib dipertahankan
- Sidebar.svelte = nav utama; Nav.svelte tidak lagi dipakai di layout
- New pages = light institutional theme; legacy pages = dark CSS (preserved)
