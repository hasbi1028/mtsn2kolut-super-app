# Mobile APK Release Center — Arsip Nonaktif

Release Center ini menyajikan artifact APK CBT Mobile lewat SvelteKit web-admin yang berjalan di port `8021` dan diteruskan oleh cloudflared. Tidak ada konfigurasi Nginx yang dibutuhkan.

**Status operasional tahun ini:** Portal Ujian Web (`/ujian`) adalah runtime resmi siswa. Source dan artifact Flutter APK tetap disimpan sebagai arsip/nonaktif untuk audit, verifikasi checksum, demo teknis, dan tahap lanjutan. Jangan mengarahkan siswa mengunduh APK dari halaman ini sebagai jalur utama ujian tahun ini kecuali ada keputusan operasional baru.

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

## Workflow rilis APK baru / arsip tahap lanjutan

Workflow ini hanya untuk publish artifact APK arsip, uji teknis internal, atau tahap lanjutan. Untuk operasi ujian tahun ini, arahkan siswa ke Portal Ujian Web (`/ujian`) dan portal pengawasan web.

Dari root repo:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
/home/servermtsn2kolut/development/flutter/bin/flutter build apk --release --split-per-abi

cd /home/servermtsn2kolut/mtsn2kolut-super-app
APK=apps/mobile/build/app/outputs/flutter-apk/app-arm64-v8a-release.apk
SHA="$(sha256sum "$APK" | awk '{print $1}')"
VERSION_RAW="$(python3 - <<'PY'
from pathlib import Path
import re
text = Path('apps/mobile/pubspec.yaml').read_text()
print(re.search(r'^version:\s*([^\s#]+)', text, re.M).group(1))
PY
)"
VERSION_NAME="${VERSION_RAW%%+*}"
VERSION_CODE="${VERSION_RAW#*+}"
scripts/publish-mobile-apk.sh \
  --apk "$APK" \
  --commit "$(git rev-parse --short HEAD)" \
  --base-url https://mtsn2kolut.sch.id \
  --channel production \
  --app-name 'MTsN 2 Kolut CBT Mobile' \
  --abi arm64-v8a \
  --version-name "$VERSION_NAME" \
  --version-code "$VERSION_CODE" \
  --expected-sha256 "$SHA" \
  --notes 'Flutter analyze lulus|Flutter test lulus|Uji perangkat operator lulus' \
  --dry-run
scripts/publish-mobile-apk.sh \
  --apk "$APK" \
  --commit "$(git rev-parse --short HEAD)" \
  --base-url https://mtsn2kolut.sch.id \
  --channel production \
  --app-name 'MTsN 2 Kolut CBT Mobile' \
  --abi arm64-v8a \
  --version-name "$VERSION_NAME" \
  --version-code "$VERSION_CODE" \
  --expected-sha256 "$SHA" \
  --notes 'Flutter analyze lulus|Flutter test lulus|Uji perangkat operator lulus'
```

Untuk production, script menolak auto-pick artifact. `--apk`, `--commit`, `--expected-sha256`, dan metadata rilis harus eksplisit. `--version-name` dan `--version-code` harus cocok dengan `apps/mobile/pubspec.yaml`.

Jika file APK sudah diberi nama khusus, arahkan langsung dengan metadata yang sama eksplisit:

```bash
APK=apps/mobile/build/app/outputs/flutter-apk/mtsn2kolut-mobile-b6c762d-anti-cheat-phase23-arm64-v8a-release.apk
SHA="$(sha256sum "$APK" | awk '{print $1}')"
VERSION_RAW="$(python3 - <<'PY'
from pathlib import Path
import re
text = Path('apps/mobile/pubspec.yaml').read_text()
print(re.search(r'^version:\s*([^\s#]+)', text, re.M).group(1))
PY
)"
VERSION_NAME="${VERSION_RAW%%+*}"
VERSION_CODE="${VERSION_RAW#*+}"
scripts/publish-mobile-apk.sh \
  --apk "$APK" \
  --commit b6c762d \
  --base-url https://mtsn2kolut.sch.id \
  --channel production \
  --app-name 'MTsN 2 Kolut CBT Mobile' \
  --abi arm64-v8a \
  --version-name "$VERSION_NAME" \
  --version-code "$VERSION_CODE" \
  --expected-sha256 "$SHA" \
  --notes 'Flutter analyze lulus|Flutter test lulus|Uji perangkat operator lulus'
```

Setelah script selesai:

- `latest-arm64.apk` langsung berubah sebagai artifact arsip/nonaktif.
- `latest.json` langsung berubah.
- QR code langsung berubah jika URL berubah.
- Dashboard release otomatis membaca metadata baru dengan label arsip/nonaktif.
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
3. Update `latest.json`, `latest-arm64.sha256`, dan `latest-qr.svg` dengan `scripts/publish-mobile-apk.sh --apk <file-history> --commit <commit-lama> --expected-sha256 <sha> ...` jika ingin metadata ikut konsisten.
4. Verifikasi URL publik.

## Batasan keamanan dan copy distribusi

- Dashboard release tetap berada di area login web-admin.
- Endpoint APK/manifest/QR dibuat public by URL untuk kebutuhan arsip/verifikasi dan tahap lanjutan, bukan instruksi utama siswa tahun ini.
- Artifact APK yang tersimpan hanya boleh berasal dari domain `mtsn2kolut.sch.id`.
- Jangan distribusikan APK dari sumber lain.
- Copy operasional untuk siswa/pengawas harus menyebut **Portal Ujian Web** sebagai jalur resmi tahun ini.

## Boundary deploy

- Script publish artifact tidak melakukan migration.
- Script publish artifact tidak restart PM2.
- Script publish artifact tidak rebuild web-admin.
- Endpoint SvelteKit harus sudah dideploy satu kali; setelah itu update APK cukup lewat script.
