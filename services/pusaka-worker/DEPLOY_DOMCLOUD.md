# Deploy PUSAKA Worker ke DomCloud

Worker kini berjalan seperti web app: **punya port HTTP + route health**,
sehingga bisa di-deploy ke DomCloud (atau PaaS lain) dan di-monitor.

## Port & Route

| Route | Fungsi |
|-------|--------|
| `GET /` | Ringkasan status (HTML) |
| `GET /health` (alias `/healthz`) | Health check — `200 {"status":"ok",...}` saat sehat, `503` saat shutdown |
| `GET /metrics` | Metrics Prometheus text |

Port dipilih dari env dengan prioritas: **`WORKER_PORT` → `PORT` (konvensi PaaS) → 8091**.

## Cara Deploy (2 opsi)

### Opsi A — Repo terpisah (disarankan)

DomCloud mendeploy satu repo per aplikasi. Pisahkan folder worker menjadi
repo sendiri (subtree split), lalu deploy dengan Dockerfile:

```bash
# Dari root monorepo, sekali saja:
git subtree split --prefix=services/pusaka-worker -b deploy-worker
git push <url-repo-domcloud> deploy-worker:master
```

Di dashboard DomCloud:
1. Import repo tersebut
2. Pilih **Dockerfile** sebagai metode build
3. Set env vars (di bawah)
4. Health check path: `/health`

### Opsi B — Deploy langsung dari monorepo

Kalau DomCloud bisa build subfolder, cukup arahkan ke `services/pusaka-worker/`
dan pastikan Dockerfile terbaca.

## Environment Variables WAJIB

| Variabel | Contoh | Keterangan |
|----------|--------|------------|
| `BACKEND_URL` | `https://mtsn2kolut.sch.id` atau `http://<ip>:8080` | URL core-api (TANPA kredensial) |
| `WORKER_API_KEY` | `<rahasia>` | API key worker (sama dengan `WORKER_API_KEY` di .env core-api/server) |
| `PORT` | `8091` | Port HTTP (DomCloud sering inject sendiri) |

## Environment Opsional

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `WORKER_PORT` | `8091` | Port HTTP health (override `PORT`) |
| `WORKER_ID` | `worker-<hostname>-<pid>` | Nama worker (muncul di /health & status backend) |
| `WORKER_CONCURRENCY` | `5` | Maksimum konsumen paralel awal (di-override backend saat sync) |
| `HEADLESS` | `true` | Mode headless browser |
| `POLL_MS` | `8000` | Interval polling job |
| `WORKER_LOG_PATH` | `../logs/worker.log` | Lokasi log |
| `SCREENSHOT_DIR` | `../logs/screenshots` | Lokasi screenshot debug |

## Catatan

- **Worker TIDAK menyentuh PostgreSQL** — komunikasi hanya via `BACKEND_URL`
  (canonical `/api/pusaka/worker/*`).
- **BACKEND_URL dari DomCloud**: pakai URL publik (`https://mtsn2kolut.sch.id`)
  atau IP server backend yang bisa diakses; jangan `localhost`.
- Health server tetap berjalan saat worker shutdown (status `shutting_down`,
  HTTP 503) sampai proses keluar — aman untuk platform yang menunggu graceful stop.
- Gambar Docker menginstal Chromium (`playwright install --with-deps chromium`)
  sehingga scrape PUSAKA berfungsi di container.
