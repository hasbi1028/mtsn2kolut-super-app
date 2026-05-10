#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="${MOBILE_RELEASE_DIR:-/home/servermtsn2kolut/releases/mtsn2kolut-mobile}"
BASE_URL="${MOBILE_RELEASE_BASE_URL:-https://mtsn2kolut.sch.id}"
CHANNEL="${MOBILE_RELEASE_CHANNEL:-production}"
APP_NAME="${MOBILE_RELEASE_APP_NAME:-MTsN 2 Kolut CBT Mobile}"
ABI="${MOBILE_RELEASE_ABI:-arm64-v8a}"
NOTES="${MOBILE_RELEASE_NOTES:-Anti-cheat Phase 2/3 aktif|Server-side violation counter dan lock aktif|Dashboard pengawas menampilkan risk peserta}"
APK_PATH=""
COMMIT=""
DRY_RUN=0

usage() {
  cat <<'USAGE'
Usage: scripts/publish-mobile-apk.sh [options]

Publish APK hasil build Flutter ke Release Center SvelteKit tanpa rebuild web.

Options:
  --apk PATH           APK yang akan dipublish. Default: auto-pick APK arm64 terbaru.
  --commit HASH        Commit build APK. Default: git rev-parse --short HEAD.
  --release-dir DIR    Folder release. Default: $MOBILE_RELEASE_DIR atau /home/servermtsn2kolut/releases/mtsn2kolut-mobile.
  --base-url URL       Domain publik. Default: $MOBILE_RELEASE_BASE_URL atau https://mtsn2kolut.sch.id.
  --notes 'A|B|C'      Catatan rilis dipisahkan pipe.
  --dry-run            Validasi input dan tampilkan rencana tanpa menulis file.
  -h, --help           Tampilkan bantuan.

Contoh:
  scripts/publish-mobile-apk.sh
  scripts/publish-mobile-apk.sh --apk apps/mobile/build/app/outputs/flutter-apk/app-arm64-v8a-release.apk
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apk) APK_PATH="${2:-}"; shift 2 ;;
    --commit) COMMIT="${2:-}"; shift 2 ;;
    --release-dir) RELEASE_DIR="${2:-}"; shift 2 ;;
    --base-url) BASE_URL="${2:-}"; shift 2 ;;
    --notes) NOTES="${2:-}"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

cd "$ROOT_DIR"

if [[ -z "$COMMIT" ]]; then
  COMMIT="$(git rev-parse --short HEAD)"
fi

if [[ -z "$APK_PATH" ]]; then
  mapfile -t candidates < <(
    find apps/mobile/build/app/outputs/flutter-apk -maxdepth 1 -type f \( \
      -name '*arm64-v8a-release.apk' -o \
      -name 'app-arm64-v8a-release.apk' -o \
      -name 'app-release.apk' \
    \) -printf '%T@ %p\n' 2>/dev/null | sort -nr | awk '{ $1=""; sub(/^ /, ""); print }'
  )
  if [[ ${#candidates[@]} -eq 0 ]]; then
    echo "ERROR: APK tidak ditemukan. Jalankan flutter build lebih dulu atau pakai --apk PATH." >&2
    exit 1
  fi
  APK_PATH="${candidates[0]}"
fi

if [[ ! -f "$APK_PATH" ]]; then
  echo "ERROR: APK tidak ada: $APK_PATH" >&2
  exit 1
fi

VERSION_RAW="$(python3 - <<'PY'
from pathlib import Path
import re
p=Path('apps/mobile/pubspec.yaml')
text=p.read_text() if p.exists() else ''
m=re.search(r'^version:\s*([^\s#]+)', text, re.M)
print(m.group(1) if m else '1.0.0+1')
PY
)"
VERSION_NAME="${VERSION_RAW%%+*}"
VERSION_CODE="${VERSION_RAW#*+}"
if [[ "$VERSION_CODE" == "$VERSION_RAW" ]]; then VERSION_CODE="1"; fi

BASE_URL="${BASE_URL%/}"
DOWNLOAD_URL="/releases/mobile/latest-arm64.apk"
ARCHIVE_NAME="mtsn2kolut-cbt-${COMMIT}-arm64.apk"
if [[ "$(basename "$APK_PATH")" == *anti-cheat* ]]; then
  ARCHIVE_NAME="mtsn2kolut-cbt-${COMMIT}-anti-cheat-arm64.apk"
fi
ARCHIVE_URL="/releases/mobile/history/${ARCHIVE_NAME}"
CHECKSUM_URL="/releases/mobile/latest-arm64.sha256"
QR_URL="/releases/mobile/latest-qr.svg"

printf 'Mobile APK publish plan\n'
printf '  root        : %s\n' "$ROOT_DIR"
printf '  apk         : %s\n' "$APK_PATH"
printf '  release dir : %s\n' "$RELEASE_DIR"
printf '  commit      : %s\n' "$COMMIT"
printf '  version     : %s (%s)\n' "$VERSION_NAME" "$VERSION_CODE"
printf '  public URL  : %s%s\n' "$BASE_URL" "$DOWNLOAD_URL"

if [[ "$DRY_RUN" == "1" ]]; then
  echo "Dry-run only; no files written."
  exit 0
fi

mkdir -p "$RELEASE_DIR/history"

python3 - "$APK_PATH" "$RELEASE_DIR" "$BASE_URL" "$COMMIT" "$APP_NAME" "$CHANNEL" "$ABI" "$VERSION_NAME" "$VERSION_CODE" "$NOTES" "$ARCHIVE_NAME" <<'PY'
from pathlib import Path
import datetime as dt
import hashlib
import json
import shutil
import sys

apk_path = Path(sys.argv[1]).resolve()
release_dir = Path(sys.argv[2]).resolve()
base_url, commit, app_name, channel, abi, version_name, version_code, notes_raw, archive_name = sys.argv[3:12]
history_dir = release_dir / 'history'
release_dir.mkdir(parents=True, exist_ok=True)
history_dir.mkdir(parents=True, exist_ok=True)

archive = history_dir / archive_name
latest = release_dir / 'latest-arm64.apk'
shutil.copy2(apk_path, archive)
shutil.copy2(apk_path, latest)
sha = hashlib.sha256(latest.read_bytes()).hexdigest()
size = latest.stat().st_size
build_time = dt.datetime.fromtimestamp(apk_path.stat().st_mtime, dt.timezone(dt.timedelta(hours=8))).isoformat()
published_at = dt.datetime.now(dt.timezone(dt.timedelta(hours=8))).isoformat()
notes = [item.strip() for item in notes_raw.split('|') if item.strip()]

download_url = '/releases/mobile/latest-arm64.apk'
archive_url = f'/releases/mobile/history/{archive_name}'
checksum_url = '/releases/mobile/latest-arm64.sha256'
qr_url = '/releases/mobile/latest-qr.svg'
manifest = {
    'app_name': app_name,
    'channel': channel,
    'platform': 'android',
    'abi': abi,
    'version_name': version_name,
    'version_code': int(version_code) if str(version_code).isdigit() else version_code,
    'commit': commit,
    'build_time': build_time,
    'published_at': published_at,
    'file_name': 'latest-arm64.apk',
    'download_url': download_url,
    'absolute_download_url': base_url + download_url,
    'archive_url': archive_url,
    'checksum_url': checksum_url,
    'qr_url': qr_url,
    'size_bytes': size,
    'sha256': sha,
    'server_url': base_url,
    'notes': notes,
}
(release_dir / 'latest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
(release_dir / 'latest-arm64.sha256').write_text(f'{sha}  latest-arm64.apk\n')
(history_dir / archive_name.replace('.apk', '.sha256')).write_text(f'{sha}  {archive_name}\n')

history_path = history_dir / 'history.json'
try:
    history = json.loads(history_path.read_text()) if history_path.exists() else []
    if not isinstance(history, list):
        history = []
except Exception:
    history = []
entry = {key: manifest[key] for key in ['app_name', 'channel', 'platform', 'abi', 'version_name', 'version_code', 'commit', 'build_time', 'published_at', 'size_bytes', 'sha256', 'server_url', 'notes']}
entry.update({
    'file_name': archive_name,
    'download_url': archive_url,
    'checksum_url': f'/releases/mobile/history/{archive_name.replace(".apk", ".sha256")}',
})
history = [item for item in history if item.get('commit') != commit or item.get('file_name') != archive_name]
history.insert(0, entry)
history_path.write_text(json.dumps(history[:20], ensure_ascii=False, indent=2) + '\n')

try:
    import qrcode
    import qrcode.image.svg
    image = qrcode.make(base_url + download_url, image_factory=qrcode.image.svg.SvgPathImage, box_size=10, border=4)
    with open(release_dir / 'latest-qr.svg', 'wb') as fh:
        image.save(fh)
except Exception as exc:
    raise SystemExit(f'ERROR: gagal membuat QR SVG. Pastikan python package qrcode tersedia. Detail: {exc}')

print(json.dumps({
    'latest': str(latest),
    'archive': str(archive),
    'download_url': base_url + download_url,
    'sha256': sha,
    'size_bytes': size,
    'manifest': str(release_dir / 'latest.json'),
    'qr': str(release_dir / 'latest-qr.svg'),
}, ensure_ascii=False, indent=2))
PY

echo "Verifikasi lokal via web-admin port 8021:"
for path in /releases/mobile/latest.json /releases/mobile/latest-arm64.apk /releases/mobile/latest-arm64.sha256 /releases/mobile/latest-qr.svg /releases/mobile/history.json; do
  code="$(curl -sS -o /tmp/mobile-release-probe -w '%{http_code}' "http://127.0.0.1:8021${path}" || true)"
  bytes="$(wc -c < /tmp/mobile-release-probe 2>/dev/null || echo 0)"
  printf '  %s %s bytes=%s\n' "$code" "$path" "$bytes"
done
