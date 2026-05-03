# MTSN 2 Kolut Super App

Monorepo ini menampung tiga runtime server dan satu client mobile:

- `apps/web-admin` untuk UI SvelteKit dan cookie/session admin
- `services/core-api` untuk Go Chi API, `sqlc`, migration, scheduler, queue, dan PostgreSQL ownership
- `services/pusaka-worker` untuk worker Playwright yang pull job dari API
- `apps/mobile` untuk APK Flutter CBT siswa berbasis BYOD

Prinsip boundary yang dipakai:

- PostgreSQL hanya dimiliki `services/core-api`
- query SQL typed tetap lewat `sqlc`
- `apps/web-admin` tidak mengakses DB langsung
- `services/pusaka-worker` tidak mengakses DB langsung
- file SQLite lama, bila masih ada, hanya dipakai sebagai sumber migrasi satu arah

## Status Terkini

Sinkron per 2026-05-03:

- roadmap aktif ada di [PLAN.md](/home/hasbiopm/mtsn2kolut-super-app/PLAN.md) dan sudah mencapai Sprint 96 Documentation Sync
- server tetap dideploy ke 3 VPS: frontend, backend, dan worker
- Flutter CBT adalah client/APK siswa, bukan VPS runtime
- Bank Soal web-admin aktif ada di `/cbt/soal`; `/cbt/questions` hanya redirect legacy
- kontrak data Bank Soal tetap lewat `/api/cbt/questions/*`
- smoke operasional CBT ada di [docs/cbt-smoke-checklist.md](/home/hasbiopm/mtsn2kolut-super-app/docs/cbt-smoke-checklist.md)
- ledger review aktif ada di [findings.md](/home/hasbiopm/mtsn2kolut-super-app/findings.md)

## Struktur

```text
apps/
  web-admin/
  mobile/
services/
  core-api/
  pusaka-worker/
tools/
```

## Setup

```bash
make install
cd services/pusaka-worker && npx playwright install chromium
```

Atau manual:

```bash
cd services/core-api/db/scripts && npm install
cd /home/hasbiopm/mtsn2kolut-super-app/services/core-api && go mod download
cd /home/hasbiopm/mtsn2kolut-super-app/apps/web-admin && npm install
cd /home/hasbiopm/mtsn2kolut-super-app/services/pusaka-worker && npm install
```

## Database PostgreSQL

Jalankan database lokal:

```bash
cd services/core-api
podman compose up -d
```

Apply migration ke database existing:

```bash
cd services/core-api/db/scripts
DATABASE_URL=postgresql://pusaka:password@localhost:5432/pusaka npm run migrate:pg
```

Migrasi data lama dari SQLite bila masih diperlukan:

```bash
cd services/core-api/db/scripts
SQLITE_PATH=../../../apps/web-admin/data/pusaka.sqlite \
DATABASE_URL=postgresql://pusaka:password@localhost:5432/pusaka \
npm run migrate
```

Catatan:

- `apps/web-admin/data/pusaka.sqlite` diperlakukan sebagai artefak legacy untuk import data lama
- runtime aktif tidak boleh menulis business state ke SQLite frontend
- file [services/core-api/docker-compose.yml](/home/hasbiopm/mtsn2kolut-super-app/services/core-api/docker-compose.yml) tetap kompatibel dipakai lewat Podman Compose

## Jalankan

Backend dev:

```bash
make dev-backend
```

Catatan:

- target ini sekarang melakukan preflight koneksi PostgreSQL ke `localhost:5432`
- jika PostgreSQL belum aktif dan Podman tersedia, script akan mencoba `podman compose up -d` otomatis dari `services/core-api`
- jika Podman tidak tersedia, script akan fallback ke Docker bila daemon Docker aktif
- jika tetap tidak ada listener PostgreSQL, target akan berhenti dengan pesan aksi yang lebih jelas

Backend Go:

```bash
make dev-backend
```

Web admin:

```bash
make dev-web
```

Worker:

```bash
make dev-worker
```

## Verifikasi

```bash
make check
```

Perintah ini menjalankan:

- `svelte-check` untuk `apps/web-admin`
- `tsc --noEmit` untuk worker
- `go test ./...` untuk `services/core-api`

## Endpoint ownership

- Go API di `services/core-api` adalah owner endpoint aplikasi dan worker
- SvelteKit berperan sebagai web app/BFF tipis ke Go API
- Worker berbicara langsung ke Go API untuk claim, heartbeat, complete, dan fail job
- SvelteKit tidak lagi menjadi transit untuk protocol worker

## Deploy

Monorepo ini tetap ditujukan untuk 3 VPS terpisah: frontend, backend, dan worker. Kontrak deploy detail ada di `deploy/DEPLOY.md`, dan config PM2 per service tersedia di `deploy/pm2/`.

Flutter CBT dibuild dan didistribusikan sebagai APK internal. Panduan build dan uji BYOD ada di `apps/mobile/README.md`, `apps/mobile/RELEASE_CHECKLIST.md`, dan `apps/mobile/BYOD_TRIAL_PROCEDURE.md`.
