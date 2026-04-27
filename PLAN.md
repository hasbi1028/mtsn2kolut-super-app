# Current Work Plan

This file is a live handoff for the next agent or Claude Code session.

## Current Objective

UI seluruh halaman sudah dimigrasikan ke shadcn-svelte + tema hijau institusional.
Langkah selanjutnya: **Sprint 1 — CBT bisa dipakai siswa** (token generation + halaman ujian).

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

**Semua API endpoints:**
- `/api/academic` (GET/POST/DELETE)
- `/api/students` (GET/POST/DELETE)
- `/api/employees` (GET/POST/PUT/DELETE + pusaka-status + update-pusaka)
- `/api/cbt/questions` (GET/POST/DELETE)
- `/api/cbt/packages` (GET/POST/DELETE)
- `/api/cbt/sessions` (GET/POST)
- `/api/cbt/sessions/{id}` (GET/DELETE)
- `/api/cbt/sessions/{id}/status` (PATCH)
- `/api/cbt/sessions/{id}/participants` (GET)
- `/api/cbt/sessions/{id}/enroll` (POST — enroll satu kelas)
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
- `/` — Dashboard operasional (queue stats, job terbaru, kontrol scheduler)
- `/employees` — Manajemen pegawai + dialog kredensial Pusaka
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

**SvelteKit API proxies tersedia:**
- `/api/employees`, `/api/employees/[id]`, `/api/employees/[id]/test-pusaka`
- `/api/jobs`, `/api/jobs/run-now`, `/api/jobs/run-all`, `/api/jobs/cancel`, `/api/jobs/cancel-all`
- `/api/attendance`, `/api/queue/stats`
- `/api/worker/status`, `/api/worker/restart`
- `/api/schedules`, `/api/settings`
- `/api/academic`, `/api/students`
- `/api/cbt/questions`, `/api/cbt/packages`, `/api/cbt/sessions`
- `/api/cbt/sessions/[id]/status`, `/api/cbt/sessions/[id]/enroll`
- `/api/cbt/sessions/[id]/score`, `/api/cbt/sessions/[id]/results`
- `/api/scheduler/tick`
- `/api/auth/change-password`, `/api/auth/logout`

---

## Alur CBT Admin Yang Sudah Bisa Digunakan

```
Buat mata pelajaran → buat kelas + tahun ajaran → input soal →
buat paket ujian → buat sesi → daftarkan siswa → mulai sesi →
selesaikan → hitung skor → lihat hasil → ekspor CSV
```

---

## Known Gaps & Bugs

### Belum Ada (Fungsional)
- **Token generation** — field `token` di `cbt_exam_participants` masih kosong (`''`)
- **Interface siswa** — tidak ada halaman untuk siswa login + mengerjakan ujian
- **Edit master data** — hampir semua halaman hanya Create + Delete, tidak ada Update untuk:
  - Soal CBT, Paket ujian, Data siswa, Data akademik
- **Assign kelas ke siswa** — UI belum ada, meski `class_id` sudah ada di schema
- **Filter rentang tanggal absensi** — hanya filter satu hari, belum rentang
- **Export absensi** — belum ada CSV/Excel untuk absensi

### Risiko Teknis
- `cbt_exam_sessions.class_id` FK wajib → satu sesi hanya untuk satu kelas
- Token ujian kosong → siswa bisa akses jawaban tanpa autentikasi jika endpoint tidak dijaga
- Single admin account → belum ada multi-user/roles guru

---

## Next Steps — Prioritas

### Sprint 1 — CBT Bisa Dipakai Siswa ← SELANJUTNYA

1. **Backend:** Generate token saat `EnrollClass` dipanggil — update `cbt_exam_participants.token` dengan UUID/random string unik per peserta
2. **Frontend Session Detail:** Tampilkan token per peserta di tabel, tombol "Generate Ulang Token"
3. **Halaman `/exam/[token]`:** Login siswa dengan token, tampil soal, submit jawaban, konfirmasi selesai
4. **Backend middleware:** Validasi token sebelum endpoint submit jawaban bisa diakses

### Sprint 2 — Edit Master Data

1. Form edit inline / modal untuk Soal CBT
2. Form edit Siswa + dropdown assign kelas
3. Form edit Pegawai (nama, NIP, unit kerja)
4. Edit judul/deskripsi Paket Ujian

### Sprint 3 — Laporan & Monitoring

1. Filter rentang tanggal + export CSV absensi
2. Live monitoring CBT — siapa sudah submit, siapa belum, sisa waktu
3. Rekap kehadiran per bulan per pegawai

### Sprint 4 — Multi-user (Jangka Panjang)

1. Akun guru dengan akses terbatas (kelas/mata pelajaran sendiri)
2. Roles: admin, guru, siswa
3. Migration baru untuk users/roles table

---

## Important Context

- Deploy order: backend code → migration → restart backend → deploy frontend → deploy worker
- Go clean architecture dan sqlc wajib dipertahankan
- Sidebar.svelte = nav utama; Nav.svelte tidak lagi dipakai di layout
- Semua halaman baru = light institutional green theme
- Legacy CSS di app.css dipertahankan tapi sudah tidak dipakai halaman manapun
- Jobs proxy (`/api/jobs/+server.ts`) rename `employee_nama` → `nama`, `employee_nip` → `nip`
- `vite.config.js` — `ssr: { noExternal: ['lucide-svelte', 'bits-ui', 'tailwind-variants'] }`
