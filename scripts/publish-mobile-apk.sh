#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="${MOBILE_RELEASE_DIR:-/home/servermtsn2kolut/releases/mtsn2kolut-mobile}"
BASE_URL="${MOBILE_RELEASE_BASE_URL:-https://mtsn2kolut.sch.id}"
CHANNEL="${MOBILE_RELEASE_CHANNEL:-production}"
APP_NAME="${MOBILE_RELEASE_APP_NAME:-MTsN 2 Kolut CBT Mobile}"
ABI="${MOBILE_RELEASE_ABI:-arm64-v8a}"
NOTES="${MOBILE_RELEASE_NOTES:-}"
APK_PATH=""
COMMIT=""
VERSION_NAME=""
VERSION_CODE=""
EXPECTED_SHA256=""
DRY_RUN=0

APK_EXPLICIT=0
COMMIT_EXPLICIT=0
BASE_URL_EXPLICIT=0
CHANNEL_EXPLICIT=0
APP_NAME_EXPLICIT=0
ABI_EXPLICIT=0
NOTES_EXPLICIT=0
VERSION_NAME_EXPLICIT=0
VERSION_CODE_EXPLICIT=0
EXPECTED_SHA256_EXPLICIT=0

usage() {
  cat <<'USAGE'
Usage: scripts/publish-mobile-apk.sh [options]

Publish APK hasil build Flutter ke Release Center SvelteKit tanpa rebuild web.

Production release wajib eksplisit: --apk, --commit, --base-url, --channel,
--app-name, --abi, --version-name, --version-code, --expected-sha256, dan
--notes harus diberikan. Non-production boleh memakai default/env lokal.

Options:
  --apk PATH                  APK yang akan dipublish.
  --commit HASH              Commit build APK.
  --release-dir DIR          Folder release. Default: $MOBILE_RELEASE_DIR atau /home/servermtsn2kolut/releases/mtsn2kolut-mobile.
  --base-url URL             Domain publik.
  --channel CHANNEL          Channel rilis, misalnya production atau staging.
  --app-name NAME            Nama aplikasi di manifest.
  --abi ABI                  ABI artifact, misalnya arm64-v8a.
  --version-name VERSION     Version name yang harus cocok dengan apps/mobile/pubspec.yaml.
  --version-code CODE        Version code yang harus cocok dengan apps/mobile/pubspec.yaml.
  --expected-sha256 SHA256   SHA-256 APK yang diharapkan; wajib untuk production.
  --notes 'A|B|C'            Catatan rilis dipisahkan pipe.
  --dry-run                  Validasi input dan tampilkan rencana tanpa menulis file.
  -h, --help                 Tampilkan bantuan.

Contoh production:
  APK=apps/mobile/build/app/outputs/flutter-apk/app-arm64-v8a-release.apk
  SHA="$(sha256sum "$APK" | awk '{print $1}')"
  scripts/publish-mobile-apk.sh \
    --apk "$APK" \
    --commit "$(git rev-parse --short HEAD)" \
    --base-url https://mtsn2kolut.sch.id \
    --channel production \
    --app-name 'MTsN 2 Kolut CBT Mobile' \
    --abi arm64-v8a \
    --version-name 1.0.0 \
    --version-code 1 \
    --expected-sha256 "$SHA" \
    --notes 'Flutter analyze lulus|Flutter test lulus|Uji perangkat operator lulus' \
    --dry-run
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apk) APK_PATH="${2:-}"; APK_EXPLICIT=1; shift 2 ;;
    --commit) COMMIT="${2:-}"; COMMIT_EXPLICIT=1; shift 2 ;;
    --release-dir) RELEASE_DIR="${2:-}"; shift 2 ;;
    --base-url) BASE_URL="${2:-}"; BASE_URL_EXPLICIT=1; shift 2 ;;
    --channel) CHANNEL="${2:-}"; CHANNEL_EXPLICIT=1; shift 2 ;;
    --app-name) APP_NAME="${2:-}"; APP_NAME_EXPLICIT=1; shift 2 ;;
    --abi) ABI="${2:-}"; ABI_EXPLICIT=1; shift 2 ;;
    --version-name) VERSION_NAME="${2:-}"; VERSION_NAME_EXPLICIT=1; shift 2 ;;
    --version-code) VERSION_CODE="${2:-}"; VERSION_CODE_EXPLICIT=1; shift 2 ;;
    --expected-sha256) EXPECTED_SHA256="${2:-}"; EXPECTED_SHA256_EXPLICIT=1; shift 2 ;;
    --notes) NOTES="${2:-}"; NOTES_EXPLICIT=1; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

cd "$ROOT_DIR"

require_value() {
  local label="$1"
  local value="$2"
  if [[ -z "$value" ]]; then
    printf 'ERROR: %s wajib diisi.\n' "$label" >&2
    exit 1
  fi
}

require_explicit_production() {
  local flag="$1"
  local explicit="$2"
  if [[ "$CHANNEL" == "production" && "$explicit" != "1" ]]; then
    printf 'ERROR: production release wajib memberi %s secara eksplisit.\n' "$flag" >&2
    exit 1
  fi
}

read_pubspec_version() {
  python3 - <<'PY'
from pathlib import Path
import re
p = Path('apps/mobile/pubspec.yaml')
text = p.read_text() if p.exists() else ''
m = re.search(r'^version:\s*([^\s#]+)', text, re.M)
raw = m.group(1) if m else '1.0.0+1'
name, sep, code = raw.partition('+')
print(name)
print(code if sep else '1')
PY
}

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

require_explicit_production "--apk" "$APK_EXPLICIT"
require_explicit_production "--commit" "$COMMIT_EXPLICIT"
require_explicit_production "--base-url" "$BASE_URL_EXPLICIT"
require_explicit_production "--channel" "$CHANNEL_EXPLICIT"
require_explicit_production "--app-name" "$APP_NAME_EXPLICIT"
require_explicit_production "--abi" "$ABI_EXPLICIT"
require_explicit_production "--version-name" "$VERSION_NAME_EXPLICIT"
require_explicit_production "--version-code" "$VERSION_CODE_EXPLICIT"
require_explicit_production "--expected-sha256" "$EXPECTED_SHA256_EXPLICIT"
require_explicit_production "--notes" "$NOTES_EXPLICIT"

require_value "--apk" "$APK_PATH"
require_value "--commit" "$COMMIT"
require_value "--base-url" "$BASE_URL"
require_value "--channel" "$CHANNEL"
require_value "--app-name" "$APP_NAME"
require_value "--abi" "$ABI"

if [[ "$CHANNEL" == "production" && "$BASE_URL" != https://* ]]; then
  echo "ERROR: production base URL wajib memakai https://." >&2
  exit 1
fi

if [[ ! -f "$APK_PATH" ]]; then
  echo "ERROR: APK tidak ada: $APK_PATH" >&2
  exit 1
fi
if [[ ! -s "$APK_PATH" ]]; then
  echo "ERROR: APK kosong: $APK_PATH" >&2
  exit 1
fi
if [[ "$(basename "$APK_PATH")" != *.apk ]]; then
  echo "ERROR: artifact harus berekstensi .apk." >&2
  exit 1
fi

python3 - "$APK_PATH" <<'PY'
from pathlib import Path
import sys
apk = Path(sys.argv[1])
with apk.open('rb') as fh:
    magic = fh.read(4)
if magic[:2] != b'PK':
    raise SystemExit('ERROR: APK tidak terlihat sebagai ZIP/APK valid.')
PY

if ! git rev-parse --verify --quiet "${COMMIT}^{commit}" >/dev/null; then
  printf 'ERROR: --commit tidak ditemukan sebagai commit lokal: %s\n' "$COMMIT" >&2
  exit 1
fi
COMMIT_FULL="$(git rev-parse "${COMMIT}^{commit}")"
COMMIT_SHORT="$(git rev-parse --short "${COMMIT}^{commit}")"

mapfile -t PUBSPEC_VERSION < <(read_pubspec_version)
PUBSPEC_VERSION_NAME="${PUBSPEC_VERSION[0]}"
PUBSPEC_VERSION_CODE="${PUBSPEC_VERSION[1]}"
if [[ -z "$VERSION_NAME" ]]; then VERSION_NAME="$PUBSPEC_VERSION_NAME"; fi
if [[ -z "$VERSION_CODE" ]]; then VERSION_CODE="$PUBSPEC_VERSION_CODE"; fi

if [[ "$VERSION_NAME" != "$PUBSPEC_VERSION_NAME" || "$VERSION_CODE" != "$PUBSPEC_VERSION_CODE" ]]; then
  printf 'ERROR: versi metadata (%s+%s) tidak cocok dengan apps/mobile/pubspec.yaml (%s+%s).\n' \
    "$VERSION_NAME" "$VERSION_CODE" "$PUBSPEC_VERSION_NAME" "$PUBSPEC_VERSION_CODE" >&2
  exit 1
fi
if [[ ! "$VERSION_CODE" =~ ^[0-9]+$ ]] || (( 10#$VERSION_CODE <= 0 )); then
  echo "ERROR: --version-code harus integer positif." >&2
  exit 1
fi

if [[ -n "$EXPECTED_SHA256" ]]; then
  EXPECTED_SHA256="$(printf '%s' "$EXPECTED_SHA256" | tr '[:upper:]' '[:lower:]')"
  if [[ ! "$EXPECTED_SHA256" =~ ^[a-f0-9]{64}$ ]]; then
    echo "ERROR: --expected-sha256 harus 64 karakter hex." >&2
    exit 1
  fi
fi

ACTUAL_SHA256="$(sha256sum "$APK_PATH" | awk '{print $1}')"
if [[ -n "$EXPECTED_SHA256" && "$ACTUAL_SHA256" != "$EXPECTED_SHA256" ]]; then
  printf 'ERROR: sha256 APK tidak cocok. actual=%s expected=%s\n' "$ACTUAL_SHA256" "$EXPECTED_SHA256" >&2
  exit 1
fi

if [[ "$CHANNEL" == "production" && -z "${NOTES// }" ]]; then
  echo "ERROR: production release wajib punya --notes yang tidak kosong." >&2
  exit 1
fi

BASE_URL="${BASE_URL%/}"
DOWNLOAD_URL="/releases/mobile/latest-arm64.apk"
CHECKSUM_URL="/releases/mobile/latest-arm64.sha256"
QR_URL="/releases/mobile/latest-qr.svg"
SAFE_VERSION="$(printf '%s' "$VERSION_NAME" | sed 's/[^A-Za-z0-9._-]/-/g')"
SAFE_ABI="$(printf '%s' "$ABI" | sed 's/[^A-Za-z0-9._-]/-/g')"
ARCHIVE_NAME="mtsn2kolut-cbt-${COMMIT_SHORT}-v${SAFE_VERSION}-${VERSION_CODE}-${SAFE_ABI}.apk"
ARCHIVE_URL="/releases/mobile/history/${ARCHIVE_NAME}"

printf 'Mobile APK publish plan\n'
printf '  root        : %s\n' "$ROOT_DIR"
printf '  apk         : %s\n' "$APK_PATH"
printf '  release dir : %s\n' "$RELEASE_DIR"
printf '  channel     : %s\n' "$CHANNEL"
printf '  app name    : %s\n' "$APP_NAME"
printf '  abi         : %s\n' "$ABI"
printf '  commit      : %s (%s)\n' "$COMMIT_SHORT" "$COMMIT_FULL"
printf '  version     : %s (%s)\n' "$VERSION_NAME" "$VERSION_CODE"
printf '  sha256      : %s\n' "$ACTUAL_SHA256"
printf '  public URL  : %s%s\n' "$BASE_URL" "$DOWNLOAD_URL"
printf '  archive     : %s\n' "$ARCHIVE_URL"

if [[ "$DRY_RUN" == "1" ]]; then
  echo "Dry-run only; no files written."
  exit 0
fi

mkdir -p "$RELEASE_DIR/history"

python3 - "$APK_PATH" "$RELEASE_DIR" "$BASE_URL" "$COMMIT_SHORT" "$COMMIT_FULL" "$APP_NAME" "$CHANNEL" "$ABI" "$VERSION_NAME" "$VERSION_CODE" "$NOTES" "$ARCHIVE_NAME" "$ACTUAL_SHA256" <<'PY'
from pathlib import Path
import datetime as dt
import json
import shutil
import sys

apk_path = Path(sys.argv[1]).resolve()
release_dir = Path(sys.argv[2]).resolve()
base_url, commit_short, commit_full, app_name, channel, abi, version_name, version_code, notes_raw, archive_name, expected_sha = sys.argv[3:14]
history_dir = release_dir / 'history'
release_dir.mkdir(parents=True, exist_ok=True)
history_dir.mkdir(parents=True, exist_ok=True)

archive = history_dir / archive_name
latest = release_dir / 'latest-arm64.apk'
shutil.copy2(apk_path, archive)
shutil.copy2(apk_path, latest)

import hashlib
sha = hashlib.sha256(latest.read_bytes()).hexdigest()
if sha != expected_sha:
    raise SystemExit(f'ERROR: checksum berubah setelah copy. latest={sha} expected={expected_sha}')

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
    'version_code': int(version_code),
    'commit': commit_short,
    'commit_full': commit_full,
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
entry = {key: manifest[key] for key in ['app_name', 'channel', 'platform', 'abi', 'version_name', 'version_code', 'commit', 'commit_full', 'build_time', 'published_at', 'size_bytes', 'sha256', 'server_url', 'notes']}
entry.update({
    'file_name': archive_name,
    'download_url': archive_url,
    'checksum_url': f'/releases/mobile/history/{archive_name.replace(".apk", ".sha256")}',
})
history = [item for item in history if item.get('commit_full') != commit_full or item.get('file_name') != archive_name]
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
