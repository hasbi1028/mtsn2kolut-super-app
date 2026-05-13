# Operasional: log rotation dan runtime artifacts

Sprint 0 safety gate menambahkan runbook ini karena log runtime PM2 pernah ditemukan besar di checkout repo. Tujuannya mencegah disk penuh dan memisahkan source code dari data operasional.

## Prinsip

- Log runtime, database local data, dan backup binary tidak menjadi bagian source canonical.
- Log production ditulis ke path operasional di luar checkout repo bila memungkinkan.
- Rotation wajib punya batas ukuran, jumlah arsip, dan retensi.
- Perubahan OS/PM2 dilakukan manual oleh operator; repo hanya menyediakan konfigurasi/runbook.

## Opsi A — pm2-logrotate

Jalankan sebagai user PM2 yang sama dengan proses aplikasi:

```bash
pm2 install pm2-logrotate
pm2 set pm2-logrotate:max_size 20M
pm2 set pm2-logrotate:retain 14
pm2 set pm2-logrotate:compress true
pm2 set pm2-logrotate:dateFormat YYYY-MM-DD_HH-mm-ss
pm2 set pm2-logrotate:rotateInterval '0 0 * * *'
pm2 save
```

Verifikasi:

```bash
pm2 conf pm2-logrotate
pm2 status
```

## Opsi B — system logrotate

Contoh konfigurasi tersedia di:

```text
deploy/logrotate/mtsn2kolut-super-app.conf
```

Install manual sebagai root/operator OS:

```bash
sudo cp deploy/logrotate/mtsn2kolut-super-app.conf /etc/logrotate.d/mtsn2kolut-super-app
sudo logrotate -d /etc/logrotate.d/mtsn2kolut-super-app
```

Jika path log production berbeda, edit path pada file tersebut sebelum dipasang.

## Preflight disk sebelum deploy

```bash
df -h .
du -sh logs services/logs apps/logs 2>/dev/null || true
find logs services/logs apps/logs -type f -size +100M -print 2>/dev/null || true
```

Deploy besar sebaiknya ditunda bila disk hampir penuh atau ada log >100 MB tanpa rotation.

## Target jangka pendek

- Production PM2 memakai `deploy/pm2/*.config.cjs` sebagai single source of truth.
- Log diarahkan ke path luar repo, misalnya `/home/servermtsn2kolut/logs/mtsn2kolut-super-app/`.
- Binary backup lama dipindahkan ke `/home/servermtsn2kolut/backups/...` dengan retensi.
