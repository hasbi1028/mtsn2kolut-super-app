# Deploy PUSAKA Worker ke DomCloud via Passenger

Worker kini berjalan seperti web app: **punya HTTP server + route health**,
sehingga bisa di-deploy ke DomCloud dengan fitur **Passenger** (tanpa Docker).

## Port & Route Health

| Route | Fungsi |
|-------|--------|
| `GET /` | Ringkasan status (HTML) |
| `GET /health` (alias `/healthz`) | Health check — `200 {"status":"ok",...}` saat sehat, `503` saat shutdown |
| `GET /metrics` | Metrics Prometheus text |

Port diambil dari env **`PORT`** (Passenger/Paas meng-inject ini) → fallback `WORKER_PORT` → `8091`.
Passenger untuk Node.js memberikan env `PORT` dan app cukup `listen` di port itu — sudah ditangani otomatis.

## Alur Deploy (Passenger)

1. **Push kode ke repo** (folder `services/pusaka-worker/` — bisa via
   `git subtree split --prefix=services/pusaka-worker` kalau repo ini monorepo).
2. Di dashboard **DomCloud**:
   - Import repo / upload folder `services/pusaka-worker`
   - Pilih engine **Node.js** (dijalankan via **Passenger**)
   - Startup file: `dist/index.js` (sudah di `main` package.json + `Passengerfile.json`)
3. **Env vars wajib** (lihat tabel di bawah)
4. Passenger otomatis menjalankan `npm install` → `postinstall` menjalankan
   `npm run build` (TypeScript) → lalu `node dist/index.js`
   (tooling build `typescript`/`tsx` sudah di `dependencies` supaya tetap
   ter-install saat `npm install --production`).
5. Health check di DomCloud: path `/health` — worker dianggap hidup jika 200.

> Catatan: kalau DomCloud menginstal browser Playwright terpisah, pastikan
> `playwright install chromium` (dengan deps OS) dijalankan sekali —
> misalnya via command di panel: `npx playwright install --with-deps chromium`.
> Worker butuh Chromium untuk scrape PUSAKA Kemenag.

## Environment Variables WAJIB

| Variabel | Contoh | Keterangan |
|----------|--------|------------|
| `BACKEND_URL` | `https://mtsn2kolut.sch.id` | URL core-api (TANPA kredensial) |
| `WORKER_API_KEY` | `<rahasia>` | API key worker (sama dengan `WORKER_API_KEY` di backend/.env) |

## Environment Opsional

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `WORKER_PORT` | `8091` | Port HTTP health (override kalau `PORT` tidak di-inject) |
| `WORKER_ID` | `worker-<hostname>-<pid>` | Nama worker (muncul di /health & status backend) |
| `WORKER_CONCURRENCY` | `5` | Maks konsumen paralel awal (di-override backend saat sync) |
| `HEADLESS` | `true` | Mode headless browser |
| `POLL_MS` | `8000` | Interval polling job |
| `WORKER_LOG_PATH` | `../logs/worker.log` | Lokasi log |
| `SCREENSHOT_DIR` | `../logs/screenshots` | Lokasi screenshot debug |

## Catatan Penting

- **Worker TIDAK menyentuh PostgreSQL** — komunikasi hanya via `BACKEND_URL`
  (canonical `/api/pusaka/worker/*`).
- **`BACKEND_URL` dari DomCloud** harus bisa diakses worker: pakai URL publik
  (`https://mtsn2kolut.sch.id`) atau IP server backend — **jangan** `localhost`
  (di DomCloud itu berarti server DomCloud sendiri).
- Worker harus tetap jalan sebagai proses panjang; jangan matikan via panel
  "stop on idle" (kalau DomCloud punya opsi itu).
- `Passengerfile.json` tersedia untuk mode standalone (`passenger start`) dan
  sebagai referensi konfigurasi Nginx Passenger di panel.
