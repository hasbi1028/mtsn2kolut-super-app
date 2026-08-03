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

### Chromium — Auto-Install (Watchdog)

Worker butuh Chromium untuk scrape PUSAKA Kemenag. DomCloud dikenal sering
**menghapus browser cache** (folder `~/.cache/ms-playwright`), kadang tiap
beberapa hari. Karena itu worker punya **watchdog otomatis**:

- **Saat start**: cek executable Chromium → kalau hilang, langsung
  `npx playwright install chromium` sebelum mulai memproses job.
- **Periodik** (default tiap 6 jam, atur via `BROWSER_CHECK_MS`): cek ulang —
  kalau browser dihapus platform, worker meng-install ulang sendiri tanpa
  perlu turun tangan manual.
- **`PLAYWRIGHT_BROWSERS_PATH=0`**: browser di-install ke dalam folder project
  (`node_modules/playwright-core/.local-browsers`), bukan `~/.cache` —
  lebih tahan terhadap pembersihan cache home oleh platform.
- Status saat sedang install di `/health`: `"status": "starting"` (HTTP 200),
  berubah `"running"` setelah worker siap.

> Jika DomCloud juga mereset **seluruh** folder app (termasuk `node_modules`),
> `npm install` akan menjalankan `postinstall` (build) dan watchdog akan
> meng-install Chromium kembali saat start — worker pulih sendiri.

> Deps OS browser (libnss3, dll): biasanya sudah ada di image Passenger
> DomCloud. Kalau muncul error `libnss3.so` saat scrape, jalankan sekali via
> panel: `npx playwright install-deps chromium` (butuh akses root/apt).

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
| `WORKER_CONCURRENCY` | `3` | **Hard cap per instance** — backend TIDAK bisa menaikkan di atas nilai ini. Default aman `3`; naikkan per instance sesuai RAM |
| `HEADLESS` | `true` | Mode headless browser |
| `POLL_MS` | `8000` | Interval polling job |
| `BROWSER_CHECK_MS` | `21600000` (6 jam) | Interval watchdog cek/install Chromium |
| `PLAYWRIGHT_BROWSERS_PATH` | `0` | Browser di dalam folder project (tahan pembersihan cache) |
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
- `Passengerfile.json` tersedia untuk mode standalone (`passenger start`) dan sebagai
  referensi konfigurasi Nginx Passenger di panel.

## ⚙️ Mengatur Concurrency dengan Banyak Instance

Worker mensinkronkan `max_concurrent` dari backend (setting global di UI
`/settings`) setiap ~30 detik. Kalau deploy **banyak instance**, semua akan
ikut nilai global yang sama → total concurrent = instance × nilai global
(mis. 5 × 35 = 175 — terlalu besar & berisiko diblokir PUSAKA Kemenag).

**Solusi (hybrid):** `WORKER_CONCURRENCY` = **hard cap per instance**
(default `3`). Backend tidak bisa menaikkan di atas cap, tapi tetap bisa
**menurunkan semua serentak** lewat UI (kontrol darurat).

```
concurrency efektif tiap instance = min(setting backend, WORKER_CONCURRENCY)
```

### Contoh — 5 Instance (default 3 → total 15, aman)

| Instance | `WORKER_CONCURRENCY` | Efektif (jika backend ≥ 3) |
|----------|---------------------|----------------------------|
| worker-1 | *(tidak diset)* | 3 |
| worker-2 | *(tidak diset)* | 3 |
| worker-3 | *(tidak diset)* | 3 |
| worker-4 | *(tidak diset)* | 3 |
| worker-5 | *(tidak diset)* | 3 |
| **Total** | | **15** |

Mau total 35? Set `WORKER_CONCURRENCY=7` di semua instance
(5 × 7 = 35 — batas yang sudah terbukti aman). Untuk instance RAM 1GB
tetap pakai default `3`.

### Aturan Praktis

1. **UI `/settings` (Max Concurrent)** = nilai global; set angka tertinggi
   yang diizinkan (mis. 35) sebagai ceiling.
2. **Env `WORKER_CONCURRENCY` per instance** = batas masing-masing (default 3);
   naikkan hanya untuk instance yang RAM-nya cukup.
3. Mau turunkan semua sekarang? Cukup ubah UI → semua instance ikut turun
   dalam ≤30 detik.
4. Cap default 3 berlaku otomatis walau env tidak diset — aman kalau lupa
   mengatur env di instance baru.
