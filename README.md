# pusaka-sveltekit-worker (NEW REPO)

Repo baru untuk manajemen scraping absensi Pusaka berbasis:
- **SvelteKit fullstack** (UI + API + scheduler)
- **Playwright worker** (proses queue job)
- **SQLite**

## Struktur

- `frontend/` → app SvelteKit
- `worker/` → worker polling queue jobs
- `data/pusaka.sqlite` → database SQLite (dibuat otomatis)

## Setup

```bash
cd frontend && npm install
cd ../worker && npm install
cd ../worker && npx playwright install chromium
```

## Jalankan

### 1) Frontend/API + Scheduler

```bash
cd frontend
npm run dev
```

Scheduler internal aktif otomatis saat server jalan, dan akan enqueue sesuai tabel `schedules` (default: pagi 07:00, sore 16:00 WITA).

### 2) Worker

```bash
cd worker
npm run start
```

## Endpoint

- `GET /api/employees`
- `POST /api/employees`
- `GET /api/jobs`
- `POST /api/jobs/run-now`
- `GET /api/schedules`
- `PUT /api/schedules`
- `POST /api/scheduler/tick` (manual trigger scheduler)
- `GET /api/attendance`

## Catatan implementasi worker

Worker melakukan:
1. claim job `queued` dari SQLite
2. login ke `https://pusaka-v3.kemenag.go.id/`
3. buka menu `Absensi` -> tombol `Riwayat Presensi`
4. parse blok **hari ini** (timezone Asia/Makassar)
5. simpan `jam_masuk` dan `jam_pulang` ke `attendance_records`

Jika data hari ini tidak ada / login gagal, job ditandai `failed` dengan `error_message`.
