# Deploy Contract

Monorepo ini tetap dideploy sebagai tiga unit runtime terpisah:

- `apps/web-admin` ke VPS frontend
- `services/core-api` ke VPS backend
- `services/pusaka-worker` ke VPS worker

Satu repo tidak berarti satu VPS. Yang penting hanya boundary source code berada di satu tempat, sedangkan proses build, env, dan runtime tetap dipisahkan per service.

## Model yang direkomendasikan

Setiap VPS melakukan checkout repo yang sama, lalu hanya menjalankan service yang relevan.

Keuntungan model ini:

- path file tetap sama dengan repo sumber
- PM2 config bisa dipakai langsung
- rollback cukup checkout commit sebelumnya
- migration PostgreSQL tetap terkunci di jalur backend

## Prasyarat per VPS

Frontend VPS:

- Node.js terpasang
- repo checkout di server
- `.env` tersedia di `apps/web-admin/.env`

Backend VPS:

- Go terpasang
- PostgreSQL tersedia atau dapat diakses
- repo checkout di server
- `.env` tersedia di `services/core-api/.env`

Worker VPS:

- Node.js terpasang
- Chromium dependencies untuk Playwright tersedia
- repo checkout di server
- `.env` tersedia di `services/pusaka-worker/.env`

## Command per VPS

VPS backend:

```bash
cd /path/to/mtsn2kolut-super-app
git pull
cd services/core-api && go build -o bin/api ./cmd/api
cd ../core-api/db/scripts && npm install
cd /path/to/mtsn2kolut-super-app/services/core-api/db/scripts && npm run migrate:pg
cd /path/to/mtsn2kolut-super-app
pm2 start deploy/pm2/backend.config.cjs || pm2 restart deploy/pm2/backend.config.cjs
```

VPS frontend:

```bash
cd /path/to/mtsn2kolut-super-app
git pull
cd apps/web-admin && npm install && npm run build
cd /path/to/mtsn2kolut-super-app
pm2 start deploy/pm2/web.config.cjs || pm2 restart deploy/pm2/web.config.cjs
```

VPS worker:

```bash
cd /path/to/mtsn2kolut-super-app
git pull
cd services/pusaka-worker && npm install
cd /path/to/mtsn2kolut-super-app
pm2 start deploy/pm2/worker.config.cjs || pm2 restart deploy/pm2/worker.config.cjs
```

Jika worker VPS baru:

```bash
cd /path/to/mtsn2kolut-super-app/services/pusaka-worker
npx playwright install chromium
```

## Urutan deploy yang aman

1. Deploy backend code lebih dulu.
2. Apply migration PostgreSQL dari backend path.
3. Restart backend dan cek health.
4. Deploy frontend.
5. Deploy worker.

Urutan ini aman karena schema owner ada di backend. Frontend dan worker cukup menyesuaikan kontrak API yang sudah naik duluan.

## Health checks minimum

Backend:

```bash
curl -fsS http://127.0.0.1:8080/health
```

Frontend:

```bash
curl -I http://127.0.0.1:8021
```

Worker:

- cek `pm2 status`
- cek log worker
- pastikan heartbeat worker muncul di backend settings/status

## Rollback

Jika deploy backend gagal:

1. checkout commit sebelumnya
2. build ulang binary backend
3. restart backend
4. jika migration baru tidak backward compatible, restore database dari backup sebelum deploy

Jika deploy frontend gagal:

1. checkout commit sebelumnya
2. build ulang `apps/web-admin`
3. restart PM2 web

Jika deploy worker gagal:

1. checkout commit sebelumnya
2. jalankan `npm install` bila lockfile berubah
3. restart PM2 worker

## Rules yang tidak boleh dilanggar

- migration PostgreSQL hanya dijalankan dari `services/core-api/db/scripts`
- frontend tidak boleh memiliki schema DB runtime sendiri
- worker tidak boleh mengakses PostgreSQL langsung
- perubahan `sqlc` harus selalu digenerate ulang dari `services/core-api/db`
