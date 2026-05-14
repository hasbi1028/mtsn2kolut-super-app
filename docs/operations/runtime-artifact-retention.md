# Operasional: runtime artifact dan retensi

Runtime artifact production tidak menjadi source canonical. Log PM2, screenshot worker, data runtime, dan binary backup harus berada di path operasional di luar checkout repo.

## Path rekomendasi

```text
/home/servermtsn2kolut/logs/mtsn2kolut-super-app/
/home/servermtsn2kolut/logs/mtsn2kolut-super-app/screenshots/
/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/
/home/servermtsn2kolut/backups/mtsn2kolut-super-app/binaries/
/home/servermtsn2kolut/releases/mtsn2kolut-mobile/
```

Jangan taruh artifact runtime di:

```text
logs/
services/logs/
services/core-api/data/
services/core-api/bin/api.backup-*
apps/web-admin/build/
services/pusaka-worker/dist/
```

## Retensi aman

Script retensi tersedia dan default-nya hanya dry-run:

```bash
deploy/scripts/runtime-artifact-retention.sh
```

Apply harus eksplisit setelah daftar file diperiksa:

```bash
deploy/scripts/runtime-artifact-retention.sh --apply
```

Script menolak target di dalam checkout repo, sehingga tidak bisa dipakai untuk membersihkan source tree secara tidak sengaja.

Override path atau hari retensi:

```bash
deploy/scripts/runtime-artifact-retention.sh \
  --log-dir /home/servermtsn2kolut/logs/mtsn2kolut-super-app \
  --screenshot-dir /home/servermtsn2kolut/logs/mtsn2kolut-super-app/screenshots \
  --binary-backup-dir /home/servermtsn2kolut/backups/mtsn2kolut-super-app/binaries \
  --log-days 14 \
  --screenshot-days 14 \
  --binary-days 30
```

Script ini tidak memindahkan artifact production. Pemindahan dari checkout repo ke path operasional harus dilakukan manual saat maintenance window, setelah service dihentikan atau log rotation dipastikan aman.
