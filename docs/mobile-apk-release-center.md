# Mobile APK Release Center

Release Center ini menyajikan APK CBT Mobile terbaru lewat SvelteKit web-admin yang berjalan di port `8021` dan diteruskan oleh cloudflared. Tidak ada konfigurasi Nginx yang dibutuhkan.

## URL produksi

- Dashboard operator: `/asesmen/aplikasi-siswa/release`
- APK terbaru: `https://mtsn2kolut.sch.id/releases/mobile/latest-arm64.apk`
- Manifest terbaru: `https://mtsn2kolut.sch.id/releases/mobile/latest.json`
- Checksum terbaru: `https://mtsn2kolut.sch.id/releases/mobile/latest-arm64.sha256`
- QR download: `https://mtsn2kolut.sch.id/releases/mobile/latest-qr.svg`
- Riwayat rilis: `https://mtsn2kolut.sch.id/releases/mobile/history.json`

## Lokasi artifact server

```text
/home/servermtsn2kolut/releases/mtsn2kolut-mobile/
├── latest-arm64.apk
├── latest-arm64.sha256
├── latest.json
├── latest-qr.svg
└── history/
    ├── history.json
    ├── mtsn2kolut-cbt-<commit>-arm64.apk
    └── mtsn2kolut-cbt-<commit>-arm64.sha256
```

Folder ini berada di luar source web-admin agar APK baru bisa dipublish tanpa rebuild web.

## Workflow rilis APK baru

Dari root repo:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
/home/servermtsn2kolut/development/flutter/bin/flutter build apk --release --split-per-abi

cd /home/servermtsn2kolut/mtsn2kolut-super-app
scripts/publish-mobile-apk.sh --apk apps/mobile/build/app/outputs/flutter-apk/app-arm64-v8a-release.apk
```

Jika file APK sudah diberi nama khusus, arahkan langsung:

```bash
scripts/publish-mobile-apk.sh --apk apps/mobile/build/app/outputs/flutter-apk/mtsn2kolut-mobile-b6c762d-anti-cheat-phase23-arm64-v8a-release.apk --commit b6c762d
```

Setelah script selesai:

- `latest-arm64.apk` langsung berubah.
- `latest.json` langsung berubah.
- QR code langsung berubah jika URL berubah.
- Dashboard release otomatis membaca metadata baru.
- Tidak perlu `npm run build`.
- Tidak perlu restart PM2 web-admin.

## Perintah validasi

```bash
curl -I https://mtsn2kolut.sch.id/releases/mobile/latest-arm64.apk
curl -fsS https://mtsn2kolut.sch.id/releases/mobile/latest.json | python3 -m json.tool
curl -fsS https://mtsn2kolut.sch.id/releases/mobile/latest-arm64.sha256
sha256sum /home/servermtsn2kolut/releases/mtsn2kolut-mobile/latest-arm64.apk
```

Header APK yang diharapkan:

```text
content-type: application/vnd.android.package-archive
cache-control: no-store, max-age=0
content-disposition: attachment; filename="mtsn2kolut-cbt-latest-arm64.apk"
```

## Rollback manual

1. Pilih APK dari `history/`.
2. Copy ke `latest-arm64.apk`.
3. Update `latest.json`, `latest-arm64.sha256`, dan `latest-qr.svg` dengan `scripts/publish-mobile-apk.sh --apk <file-history> --commit <commit-lama>` jika ingin metadata ikut konsisten.
4. Verifikasi URL publik.

## Batasan keamanan

- Dashboard release tetap berada di area login web-admin.
- Endpoint APK/manifest/QR dibuat public by URL agar siswa/pengawas bisa scan dan download cepat.
- APK resmi hanya dari domain `mtsn2kolut.sch.id`.
- Jangan distribusikan APK dari sumber lain.

## Boundary deploy

- Script publish artifact tidak melakukan migration.
- Script publish artifact tidak restart PM2.
- Script publish artifact tidak rebuild web-admin.
- Endpoint SvelteKit harus sudah dideploy satu kali; setelah itu update APK cukup lewat script.
