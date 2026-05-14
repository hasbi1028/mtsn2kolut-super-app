# Operasional: PM2 config production

Sprint 5 menetapkan `deploy/pm2/*.config.cjs` sebagai single source of truth untuk production. Root `ecosystem.config.cjs` hanya wrapper kompatibilitas yang menggabungkan tiga config deploy agar command lama tidak drift.

## File sumber

- Backend VPS: `deploy/pm2/backend.config.cjs`
- Frontend VPS: `deploy/pm2/web.config.cjs`
- Worker VPS: `deploy/pm2/worker.config.cjs`
- Shared default: `deploy/pm2/common.cjs`
- Validator drift: `scripts/validate-pm2-configs.mjs`

## Preflight non-mutating

Jalankan dari root repo sebelum PM2 config diterapkan:

```bash
node scripts/validate-pm2-configs.mjs
```

Validator memeriksa:

- root `ecosystem.config.cjs` masih mirror dari `deploy/pm2/*.config.cjs`,
- path `cwd` dan `env_file` absolut dan sesuai checkout repo,
- path log PM2 dan worker artifact mengarah ke luar repo,
- `kill_timeout`, restart policy, `watch: false`, `instances`, dan `exec_mode` konsisten.

## Path operasional

Default log production:

```text
/home/servermtsn2kolut/logs/mtsn2kolut-super-app/
```

Override aman bila topology server berbeda:

```bash
export MTSN2KOLUT_LOG_DIR=/home/servermtsn2kolut/logs/mtsn2kolut-super-app
```

Untuk worker, `WORKER_LOG_PATH` dan `SCREENSHOT_DIR` ikut default PM2 di atas kecuali dioverride dari `.env` atau environment operator.

## Command per VPS

Gunakan config unit, bukan copy manual app block:

```bash
pm2 start deploy/pm2/backend.config.cjs
pm2 start deploy/pm2/web.config.cjs
pm2 start deploy/pm2/worker.config.cjs
```

Perubahan config baru efektif setelah operator menjalankan start/restart PM2 secara manual sesuai deploy window. Sprint ini tidak melakukan restart.
