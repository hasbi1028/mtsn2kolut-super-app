# Deploy Contract

Monorepo ini tetap dideploy sebagai tiga unit runtime terpisah:

- `apps/web-admin` ke VPS frontend
- `services/core-api` ke VPS backend
- `services/pusaka-worker` ke VPS worker

Satu repo tidak berarti satu VPS. Yang penting hanya boundary source code berada di satu tempat, sedangkan proses build, env, dan runtime tetap dipisahkan per service. `apps/mobile` tidak masuk PM2/VPS; Flutter CBT dibuild sebagai APK internal.

## Status Terkini

Sinkron per 2026-05-04:

- deploy server tetap 3 VPS: frontend, backend, worker
- backend tetap owner migration PostgreSQL
- web-admin tetap BFF/UI
- worker tetap API client ke `/api/pusaka/worker/*`
- Flutter CBT memakai exam API langsung dan harus dicek lewat `docs/exam-api.md`
- sebelum ujian besar, ikuti `docs/cbt-operator-runbook.md` dan jalankan `docs/cbt-smoke-checklist.md` setelah backend/frontend deploy
- bila staging punya akun uji dan browser tooling, jalankan scaffold `apps/web-admin/docs/cbt-role-smoke.md`

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
- `TRUSTED_PROXY_CIDRS` diisi hanya dengan CIDR reverse proxy/load balancer terpercaya jika backend berada di belakang proxy
- `WORKER_API_KEY` diset dan sama dengan nilai di worker VPS

Worker VPS:

- Node.js terpasang
- Chromium dependencies untuk Playwright tersedia
- repo checkout di server
- `.env` tersedia di `services/pusaka-worker/.env`
- `WORKER_API_KEY` wajib untuk backend non-lokal dan harus sama dengan backend
- `WORKER_API_TIMEOUT_MS` diset eksplisit bila default `10000` ms tidak sesuai kondisi jaringan
- `WORKER_LOG_PATH` dan `SCREENSHOT_DIR` mengarah ke path non-public dengan permission/retention terbatas; default PM2 Sprint 5 berada di `/home/servermtsn2kolut/logs/mtsn2kolut-super-app`

## Command per VPS

Preflight PM2 config dari root repo, tanpa restart:

```bash
node scripts/validate-pm2-configs.mjs
```

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
cd services/pusaka-worker && npm install && npm run build
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

Checklist APK release CBT sebelum dibagikan:

- pastikan permission Android `INTERNET` ada di manifest/APK sehingga aplikasi dapat menghubungi backend
- gunakan `API_BASE_URL` HTTPS untuk staging/produksi; HTTP hanya boleh untuk dev lokal terkontrol
- token uji baru harus 32 karakter hex dan cocok dengan sesi yang akan diuji
- signing APK harus sesuai prosedur internal Android release; catat versi, commit, dan tanggal build
- uji pada perangkat nyata lintas vendor dan catat hasil di `apps/mobile/DEVICE_TEST_MATRIX.md`

## Urutan deploy yang aman

1. Deploy backend code lebih dulu.
2. Apply migration PostgreSQL dari backend path.
3. Verifikasi env backend: `TRUSTED_PROXY_CIDRS` dan `WORKER_API_KEY`.
4. Restart backend dan cek health.
5. Deploy frontend.
6. Verifikasi env worker: `WORKER_API_KEY`, `WORKER_API_TIMEOUT_MS`, `WORKER_LOG_PATH`, dan `SCREENSHOT_DIR`.
7. Deploy/restart worker.
8. Untuk rilis CBT, jalankan smoke checklist admin/guru, scaffold web-admin bila tersedia, dan token login Flutter.

Urutan ini aman karena schema owner ada di backend. Frontend dan worker cukup menyesuaikan kontrak API yang sudah naik duluan.

## Preflight migration CBT

Wajib dilakukan sebelum migration CBT yang menyentuh token, peserta, seat, event member, atau scope soal.

1. Backup database terlebih dahulu:

```bash
make ops-backup
```

2. Konfirmasi tidak ada jendela ujian aktif atau segera dimulai. Jangan migrate saat siswa bisa login token, heartbeat, simpan jawaban, atau submit.
3. Cek token peserta duplikat sebelum migration 060:

```sql
SELECT token, COUNT(*)
FROM cbt_exam_participants
WHERE token <> ''
GROUP BY token
HAVING COUNT(*) > 1;
```

4. Cek seat invalid sebelum migration 061:

```sql
SELECT id, session_id, room_id, seat_no
FROM cbt_exam_participants
WHERE seat_no IS NOT NULL
  AND seat_no <= 0;
```

5. Cek seat duplikat sebelum migration 061:

```sql
SELECT room_id, seat_no, COUNT(*)
FROM cbt_exam_participants
WHERE room_id IS NOT NULL
  AND seat_no IS NOT NULL
GROUP BY room_id, seat_no
HAVING COUNT(*) > 1;
```

6. Jalankan migration dari backend path dengan `make db-migrate` atau command DB script yang setara.
7. Setelah migration, jalankan `make ops-health` dan smoke CBT sebelum sesi ujian diaktifkan.

Catatan migration terbaru:

- `060_cbt_exam_token_hardening.sql` tidak merotasi token yang sudah ada. Migration ini hanya menolak token non-empty yang duplikat dan menambahkan unique index token. Cetak ulang kartu hanya setelah operator menjalankan regenerasi/repair token eksplisit atau data kartu seperti ruang/seat berubah.
- `061_cbt_participant_seat_invariants.sql` menolak `seat_no <= 0` dan duplikasi `(room_id, seat_no)`, lalu menambahkan constraint positif dan unique index room-seat.
- `062_cbt_event_members_question_scope.sql` menambahkan role anggota event `panitia`, `pembuat_soal`, `reviewer`, `proktor`, `pengawas`, dan `korektor`. Hanya `pembuat_soal`, `reviewer`, dan `korektor` yang boleh punya `subject_id`; role lain bersifat event/ruang tanpa scope mapel.

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

## Target Makefile operasional

Jalankan dari root repo kecuali command Flutter langsung yang disebutkan eksplisit.

| Target/command | Fungsi |
|----------------|--------|
| `make ops-backup` | Backup PostgreSQL sebelum migration/deploy. |
| `make db-migrate` | Apply migration PostgreSQL memakai `DATABASE_URL` atau `.env` backend. |
| `make db-migrate-local` | Apply migration ke database lokal yang diizinkan eksplisit. |
| `make ops-health` | Health check backend, frontend, dan worker. |
| `make ops-health-backend` | Health check backend saja. |
| `make ops-health-frontend` | Health check frontend saja. |
| `make ops-health-worker` | Health check worker saja. |
| `make check` | Check/test utama repo yang sudah terhubung di Makefile. |
| `make ci-check` | Check ringan gaya CI tanpa vet/audit tambahan. |
| `make check-web` | `npm run check` untuk SvelteKit web-admin. |
| `make check-mobile` | `flutter analyze` untuk Flutter CBT. |
| `make test-mobile` | `flutter test` untuk Flutter CBT. |
| `make mobile-release-apk API_BASE_URL=https://api.sekolah.example` | Build APK release Flutter dengan HTTPS base URL. |
| `cd apps/mobile && flutter pub get` | Install dependency Flutter sebelum analyze/test/build. |
| `cd apps/mobile && flutter build apk --release --dart-define=API_BASE_URL=https://...` | Command build APK langsung bila tidak memakai Makefile. |

## Backup PostgreSQL minimum

Dari root repo:

```bash
make ops-backup
```

`make ops-backup` memanggil `deploy/backup-postgresql.sh`, yang juga dipakai oleh systemd timer di VPS backend. Output:

- `~/backups/mtsn2kolut-super-app/postgresql/<db>_<timestamp>.dump` (custom format, bisa dipulihkan selektif via `pg_restore`).
- `<dump>.sha256` checksum.
- `latest.dump` symlink ke dump terbaru.
- Log eksekusi di `~/backups/mtsn2kolut-super-app/logs/postgresql-backup.log`.
- Retention 30 hari otomatis.

Script ini memakai `flock` agar tidak overlap dengan jalankan ganda (systemd timer + manual `make ops-backup`).

### Restore drill (uji rutin di staging)

Lakukan restore drill secara periodik (mis. tiap rilis besar) agar backup terbukti dapat dipakai:

```bash
# 1. Identifikasi dump terbaru
ls -lh ~/backups/mtsn2kolut-super-app/postgresql/latest.dump
sha256sum -c ~/backups/mtsn2kolut-super-app/postgresql/latest.dump.sha256

# 2. Buat database staging kosong (JANGAN ke database produksi)
createdb -h 127.0.0.1 -U pusaka pusaka_restore_drill

# 3. Restore dump
PGPASSWORD=*** pg_restore \
  --host=127.0.0.1 --username=pusaka --dbname=pusaka_restore_drill \
  --no-owner --no-privileges --clean --if-exists \
  ~/backups/mtsn2kolut-super-app/postgresql/latest.dump

# 4. Smoke check — jumlah row tabel kunci
psql -h 127.0.0.1 -U pusaka pusaka_restore_drill -c "SELECT count(*) FROM users;"
psql -h 127.0.0.1 -U pusaka pusaka_restore_drill -c "SELECT count(*) FROM cbt_questions;"

# 5. Bersihkan database drill
dropdb -h 127.0.0.1 -U pusaka pusaka_restore_drill
```

Catat waktu restore aktual sebagai RTO baseline. Bila restore gagal atau checksum mismatch, perbaiki source backup pipeline sebelum next deploy.

## CI ringan

Repo ini sekarang punya workflow CI dasar untuk:

- `go test ./...` pada `services/core-api`
- `npm run check` pada `apps/web-admin`
- `tsc --noEmit` pada `services/pusaka-worker`

CI ini sengaja ringan dan hanya memeriksa kualitas dasar, bukan deploy automation.

## Rate limit dan proxy

Rate limit sensitif saat ini bersifat in-memory per proses backend. Model ini cocok untuk baseline satu backend VPS/proses, tetapi belum cukup bila backend dijalankan dalam beberapa replica, PM2 cluster, atau load balancer yang membagi trafik ke banyak proses.

Sebelum scaling backend horizontal:

- pindahkan counter rate limit ke shared store seperti Redis atau mekanisme counter/advisory PostgreSQL, atau terapkan limit setara di edge proxy terpercaya
- pertahankan limiter aplikasi sebagai lapisan tambahan setelah shared/edge limiter aktif
- uji ulang batas login, refresh, registrasi publik, dan login token ujian di staging
- pastikan format log/proxy membuat IP client asli dapat dikorelasikan dengan log backend

`TRUSTED_PROXY_CIDRS` hanya boleh berisi CIDR reverse proxy/load balancer yang benar-benar menjadi hop terpercaya di depan backend. Jangan isi dengan CIDR publik luas, range CDN yang tidak menjadi direct trusted hop, atau `0.0.0.0/0`, karena header `X-Forwarded-For` palsu dapat mengacaukan rate limit dan konteks audit.

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
2. jalankan `npm install && npm run build` bila source, dependency, atau lockfile worker berubah
3. restart PM2 worker

## Rules yang tidak boleh dilanggar

- migration PostgreSQL hanya dijalankan dari `services/core-api/db/scripts`
- frontend tidak boleh memiliki schema DB runtime sendiri
- worker tidak boleh mengakses PostgreSQL langsung
- perubahan `sqlc` harus selalu digenerate ulang dari `services/core-api/db`
- APK Flutter harus diverifikasi terhadap kontrak `docs/exam-api.md` sebelum dibagikan ke siswa
- jangan isi `TRUSTED_PROXY_CIDRS` dengan CIDR publik luas; hanya proxy yang dipercaya boleh menentukan IP client forwarded
- jangan commit/log `WORKER_API_KEY`, log PUSAKA, atau screenshot worker; simpan di path non-public dan gunakan rotasi/retention
- jangan bypass upload allowlist atau header `nosniff` backend saat menambah endpoint file baru
