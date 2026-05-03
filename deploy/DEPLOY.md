# Deploy Contract

Monorepo ini tetap dideploy sebagai tiga unit runtime terpisah:

- `apps/web-admin` ke VPS frontend
- `services/core-api` ke VPS backend
- `services/pusaka-worker` ke VPS worker

Satu repo tidak berarti satu VPS. Yang penting hanya boundary source code berada di satu tempat, sedangkan proses build, env, dan runtime tetap dipisahkan per service. `apps/mobile` tidak masuk PM2/VPS; Flutter CBT dibuild sebagai APK internal.

## Status Terkini

Sinkron per 2026-05-03:

- deploy server tetap 3 VPS: frontend, backend, worker
- backend tetap owner migration PostgreSQL
- web-admin tetap BFF/UI
- worker tetap API client ke `/api/pusaka/worker/*`
- Flutter CBT memakai exam API langsung dan harus dicek lewat `docs/exam-api.md`
- sebelum ujian besar, jalankan `docs/cbt-smoke-checklist.md` setelah backend/frontend deploy

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

APK Flutter CBT:

```bash
cd /path/to/mtsn2kolut-super-app/apps/mobile
flutter pub get
flutter analyze
flutter test
flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
```

Distribusi APK mengikuti `apps/mobile/RELEASE_CHECKLIST.md` dan `apps/mobile/BYOD_TRIAL_PROCEDURE.md`.

## Urutan deploy yang aman

1. Deploy backend code lebih dulu.
2. Apply migration PostgreSQL dari backend path.
3. Restart backend dan cek health.
4. Deploy frontend.
5. Deploy worker.
6. Untuk rilis CBT, jalankan smoke checklist admin/guru dan token login Flutter.

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

Shortcut lokal dari root repo:

```bash
make ops-health
make ops-health-backend
make ops-health-frontend
make ops-health-worker
```

## Backup PostgreSQL minimum

Dari root repo:

```bash
make ops-backup
```

Opsional override retention:

```bash
BACKUP_RETENTION_DAYS=14 make ops-backup
```

Contoh cron ringan di VPS backend:

```bash
0 2 * * * cd /path/to/mtsn2kolut-super-app && BACKUP_RETENTION_DAYS=14 ./deploy/scripts/backup.sh /backups/mtsn2kolut >> /var/log/mtsn2kolut-backup.log 2>&1
```

Backup ini menghasilkan dump PostgreSQL terkompresi `.sql.gz` dan membersihkan file yang lebih tua dari retention window.

## CI ringan

Repo ini sekarang punya workflow CI dasar untuk:

- `go test ./...` pada `services/core-api`
- `npm run check` pada `apps/web-admin`
- `tsc --noEmit` pada `services/pusaka-worker`

CI ini sengaja ringan dan hanya memeriksa kualitas dasar, bukan deploy automation.

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
- APK Flutter harus diverifikasi terhadap kontrak `docs/exam-api.md` sebelum dibagikan ke siswa
