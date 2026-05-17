# Flutter Windows GitHub Actions + CBT Center Release Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menambahkan jalur build Windows untuk aplikasi CBT Mobile Flutter melalui GitHub Actions, lalu menyiapkan CBT Center agar bisa menyajikan download Windows ZIP berdampingan dengan Android APK.

**Architecture:** Build Windows tidak dilakukan di server Linux Mint karena Flutter Windows membutuhkan toolchain Windows/Visual Studio/MSBuild. Server Linux tetap menjadi release center. GitHub Actions memakai runner `windows-latest` untuk membuat ZIP Windows release, sedangkan server menyajikan artifact lewat endpoint `/releases/mobile/*` yang sudah dipakai APK Android.

**Tech Stack:** Flutter 3.41.x, GitHub Actions `windows-latest`, PowerShell packaging, SvelteKit web-admin release center, shell publish scripts, external release directory `/home/servermtsn2kolut/releases/mtsn2kolut-mobile/`.

---

## Current State

- Repo: `/home/servermtsn2kolut/mtsn2kolut-super-app`
- Flutter app: `apps/mobile`
- Server OS: Linux Mint 22.3
- Flutter installed at: `/home/servermtsn2kolut/development/flutter/bin/flutter`
- Current app platforms found:
  - `apps/mobile/android`
  - `apps/mobile/linux`
- Missing platform:
  - `apps/mobile/windows`
- Existing APK release center docs:
  - `docs/mobile-apk-release-center.md`
- Existing Android publish script:
  - `scripts/publish-mobile-apk.sh`
- Existing operator page:
  - `apps/web-admin/src/routes/asesmen/aplikasi-siswa/release/+page.svelte`
- Existing public Android URLs:
  - `/releases/mobile/latest-arm64.apk`
  - `/releases/mobile/latest.json`
  - `/releases/mobile/latest-arm64.sha256`
  - `/releases/mobile/latest-qr.svg`

## Important Safety Notes

1. Do **not** build Windows EXE directly on Linux Mint; use GitHub Actions Windows runner or an actual Windows machine.
2. Do **not** commit release binaries (`.apk`, `.exe`, `.zip`) into git.
3. Do **not** store SSH passwords, tokens, or server credentials in workflow YAML.
4. Initial workflow should only upload GitHub artifact; server publish can stay manual first.
5. Existing mobile tests are currently not fully clean because several stale UI/copy expectations fail. Do not hide this. Use `flutter analyze` and targeted smoke/build gates for the initial Windows pipeline, then fix tests in a separate sprint if needed.
6. Current working tree has uncommitted mobile changes. Before implementation, review whether those changes are intentional and either commit/stash them or include them explicitly.

---

## Target User Flow

### Operator flow

1. Admin/operator opens:
   - `/asesmen/aplikasi-siswa/release`
2. Page shows two download cards:
   - Android APK: `latest-arm64.apk`
   - Windows CBT ZIP: `latest-windows-x64.zip`
3. Operator can copy link, download checksum, and scan QR.
4. Siswa/lab Windows downloads ZIP, extracts it, and runs CBT app executable.

### Developer/release flow v1: manual publish

1. Developer pushes code to GitHub.
2. Developer opens GitHub Actions.
3. Runs workflow **Build CBT Mobile Windows**.
4. Downloads artifact ZIP from GitHub.
5. Uploads ZIP to server.
6. Runs publish script on server:
   - validates ZIP
   - writes checksum
   - updates Windows manifest
   - archives prior releases
7. CBT Center immediately shows Windows release; no PM2 restart needed if release routes already support it.

### Developer/release flow v2: automatic publish later

1. Workflow builds Windows ZIP.
2. GitHub Actions uploads ZIP to server via SSH/SCP using GitHub Secrets.
3. Server runs publish script remotely.
4. CBT Center updates automatically.

Implement v1 first. Add v2 only after v1 is stable.

---

# Sprint 0 — Preflight and GitHub Readiness

## Task 0.1: Verify repo and protect current worktree

**Objective:** Avoid overwriting current mobile changes before adding Windows platform/workflow files.

**Files:**
- Inspect only: repo root

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git branch --show-current
git log --oneline -5
```

**Expected:**

- Branch is correct, likely `feature/comprehensive-improvements`.
- Working tree may show current mobile changes. Decide before editing:
  - If changes are intentional and ready: commit them first.
  - If unrelated: stash them.
  - If generated/experimental: do not include in this Windows plan.

**Acceptance Criteria:**

- No unknown working tree changes are accidentally mixed into Windows build pipeline commit.

---

## Task 0.2: Confirm GitHub remote exists

**Objective:** GitHub Actions only works if repo is pushed to GitHub.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git remote -v
```

**Expected:**

- At least one GitHub remote exists, usually `origin`.

**If no GitHub remote:**

1. Create GitHub repository.
2. Add remote:

```bash
git remote add origin git@github.com:<owner>/<repo>.git
```

3. Push branch:

```bash
git push -u origin feature/comprehensive-improvements
```

**Acceptance Criteria:**

- GitHub repository can receive commits.

---

## Sprint 0 Findings — 2026-05-17

Sprint 0 preflight sudah dijalankan dengan hasil:

- Branch aktif: `feature/comprehensive-improvements`.
- Remote GitHub tersedia: `origin https://github.com/hasbi1028/mtsn2kolut-super-app.git`.
- Branch lokal sudah terhubung ke `origin/feature/comprehensive-improvements`.
- Status sync: lokal `ahead 23`, `behind 0`; artinya 23 commit lokal belum dipush ke GitHub.
- Last local head: `dcdf728 feat(cbt): add BYOD proctoring realtime actions`.
- Remote head saat audit: `ce5aa9ee326a`.
- `.github/workflows/` sudah ada dengan workflow existing: `.github/workflows/ci.yml`.
- `apps/mobile/windows/` belum ada; Sprint 1 masih perlu generate Windows runner.
- Working tree belum bersih dan tidak boleh langsung dicampur dengan Sprint Windows:
  - Modified:
    - `apps/mobile/lib/src/screens/exam_login_screen.dart`
    - `apps/mobile/lib/src/screens/exam_shell_screen.dart`
    - `apps/mobile/test/anti_cheat_guard_test.dart`
    - `apps/mobile/test/exam_api_test.dart`
    - `apps/mobile/test/exam_error_messages_test.dart`
    - `apps/mobile/test/widget_test.dart`
  - Untracked:
    - `.hermes/plans/2026-05-17_flutter-windows-github-actions-cbt-center.md`
    - `apps/mobile/test/coverage_gap_pure_dart_test.dart`
    - `apps/mobile/test/exam_login_coverage_gap_test.dart`

Audit cepat perubahan mobile menunjukkan perubahan ini tampaknya terkait coverage/test dan copy CBT Mobile, bukan Windows runner/GitHub Actions. Sebelum Sprint 1, pilih salah satu:

1. Commit perubahan mobile tersebut sebagai pekerjaan terpisah jika memang resmi.
2. Stash perubahan mobile tersebut agar Sprint Windows bersih.
3. Batalkan perubahan jika hanya eksperimen.

Rekomendasi: jangan mulai Sprint 1 sebelum working tree dirapikan, supaya commit Windows hanya berisi `apps/mobile/windows/**`, workflow GitHub Actions, script/docs release Windows, dan UI CBT Center terkait.

---

# Sprint 1 — Add Windows Platform to Flutter App

## Task 1.1: Create Flutter Windows runner files

**Objective:** Add `apps/mobile/windows/` so Flutter can build desktop Windows.

**Files:**
- Create: `apps/mobile/windows/**`
- Modify possibly: `apps/mobile/.metadata`

**Command:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter create --platforms=windows .
```

**Expected:**

- New folder `apps/mobile/windows/` appears.
- Existing Dart files should not be rewritten materially.

**Verification:**

```bash
git status --short apps/mobile/windows apps/mobile/.metadata
```

**Acceptance Criteria:**

- `apps/mobile/windows/runner/main.cpp` exists.
- `apps/mobile/windows/CMakeLists.txt` exists.

---

## Task 1.2: Set Windows app display metadata

**Objective:** Make Windows app name school-friendly instead of generic `mobile` if needed.

**Files:**
- Modify: `apps/mobile/windows/runner/Runner.rc`
- Modify: `apps/mobile/windows/runner/main.cpp` or generated title constants if applicable

**Implementation Notes:**

Use display name:

```text
MTsN 2 Kolut CBT
```

Do not change package identity unless Flutter runner requires it.

**Verification:**

```bash
git diff -- apps/mobile/windows
```

**Acceptance Criteria:**

- Windows title/file metadata references MTsN 2 Kolut CBT or CBT Mobile, not a generic demo name.

---

## Task 1.3: Add Windows platform docs note

**Objective:** Document that Windows builds happen on Windows runner, not Linux server.

**Files:**
- Modify: `apps/mobile/README.md` if exists, otherwise create `docs/mobile-windows-release.md`

**Content Required:**

Include:

```markdown
# CBT Mobile Windows Build

Windows release builds are produced with GitHub Actions on `windows-latest` because Flutter Windows requires Visual Studio/MSBuild and Windows SDK. The Linux Mint server remains the release center and does not build `.exe` directly.
```

**Acceptance Criteria:**

- A maintainer can understand why Windows build is external.

---

# Sprint 2 — GitHub Actions Windows Build Artifact

## Task 2.1: Create workflow directory if missing

**Objective:** Ensure `.github/workflows` exists.

**Files:**
- Create: `.github/workflows/`

**Command:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
mkdir -p .github/workflows
```

**Acceptance Criteria:**

- Directory exists.

---

## Task 2.2: Add Windows build workflow

**Objective:** Build Windows Flutter app on GitHub-hosted Windows runner and upload ZIP artifact.

**Files:**
- Create: `.github/workflows/build-mobile-windows.yml`

**Workflow content:**

```yaml
name: Build CBT Mobile Windows

on:
  workflow_dispatch:
  push:
    paths:
      - 'apps/mobile/**'
      - '.github/workflows/build-mobile-windows.yml'

permissions:
  contents: read

jobs:
  build-windows:
    name: Build Windows x64 ZIP
    runs-on: windows-latest
    timeout-minutes: 45

    defaults:
      run:
        working-directory: apps/mobile

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Setup Flutter
        uses: subosito/flutter-action@v2
        with:
          channel: stable
          cache: true

      - name: Show Flutter doctor
        run: flutter doctor -v

      - name: Enable Windows desktop
        run: flutter config --enable-windows-desktop

      - name: Install dependencies
        run: flutter pub get

      - name: Analyze
        run: flutter analyze

      # Full `flutter test` can be enabled after stale UI/copy tests are updated.
      # For initial Windows packaging, analyze + build are the release gates.
      - name: Build Windows release
        run: flutter build windows --release

      - name: Package Windows release
        shell: pwsh
        working-directory: .
        run: |
          $commit = "${{ github.sha }}".Substring(0, 7)
          $versionLine = Select-String -Path apps/mobile/pubspec.yaml -Pattern '^version:' | Select-Object -First 1
          $versionRaw = ($versionLine.Line -replace '^version:\s*', '').Trim()
          $versionSafe = $versionRaw -replace '[^A-Za-z0-9._+-]', '-'
          $package = "mtsn2kolut-cbt-windows-$commit-v$versionSafe.zip"
          $releaseDir = "apps/mobile/build/windows/x64/runner/Release"
          if (!(Test-Path $releaseDir)) {
            throw "Release directory not found: $releaseDir"
          }
          Compress-Archive -Path "$releaseDir/*" -DestinationPath $package -Force
          $sha = (Get-FileHash -Algorithm SHA256 $package).Hash.ToLower()
          "$sha  $package" | Out-File -FilePath "$package.sha256" -Encoding ascii
          "WINDOWS_PACKAGE=$package" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append
          "WINDOWS_SHA=$sha" | Out-File -FilePath $env:GITHUB_ENV -Encoding utf8 -Append

      - name: Upload Windows artifact
        uses: actions/upload-artifact@v4
        with:
          name: mtsn2kolut-cbt-windows
          path: |
            ${{ env.WINDOWS_PACKAGE }}
            ${{ env.WINDOWS_PACKAGE }}.sha256
          if-no-files-found: error
          retention-days: 30
```

**Verification:**

```bash
git diff -- .github/workflows/build-mobile-windows.yml
```

**Acceptance Criteria:**

- Workflow has `workflow_dispatch` for manual run.
- Workflow uses `windows-latest`.
- Workflow uploads ZIP and SHA256.
- Workflow does not contain secrets.

---

## Task 2.3: Add local workflow documentation for operator/developer

**Objective:** Make the workflow usable by someone unfamiliar with GitHub Actions.

**Files:**
- Create: `docs/mobile-windows-github-actions.md`

**Required sections:**

```markdown
# Build Windows CBT App with GitHub Actions

## What this does
GitHub runs a temporary Windows machine, builds Flutter Windows, and gives us a ZIP artifact.

## How to run
1. Open GitHub repository.
2. Click Actions.
3. Click Build CBT Mobile Windows.
4. Click Run workflow.
5. Wait until green.
6. Download artifact `mtsn2kolut-cbt-windows`.

## What to upload to CBT Center
Upload the ZIP, not only `.exe`, because Flutter Windows needs DLL and `data/` assets.

## Safety
Do not paste server passwords into workflow YAML. Use GitHub Secrets only in the automatic-publish sprint.
```

**Acceptance Criteria:**

- Non-technical operator can follow the basic run/download steps.

---

# Sprint 3 — Manual Windows Release Publish Script on Server

## Task 3.1: Add Windows release publish script

**Objective:** Publish Windows ZIP to the same external release directory pattern as Android APK.

**Files:**
- Create: `scripts/publish-mobile-windows.sh`

**Design:**

Input required for production:

- `--zip PATH`
- `--commit HASH`
- `--base-url https://mtsn2kolut.sch.id`
- `--channel production`
- `--app-name 'MTsN 2 Kolut CBT Windows'`
- `--arch x64`
- `--version-name VERSION`
- `--version-code CODE`
- `--expected-sha256 SHA`
- `--notes 'A|B|C'`

Output files:

```text
/home/servermtsn2kolut/releases/mtsn2kolut-mobile/
├── latest-windows-x64.zip
├── latest-windows-x64.sha256
├── latest-windows.json
├── latest-windows-qr.svg
└── history/
    ├── mtsn2kolut-cbt-windows-<commit>-v<version>-<code>-x64.zip
    └── mtsn2kolut-cbt-windows-<commit>-v<version>-<code>-x64.sha256
```

**Implementation requirements:**

- Validate file exists, non-empty, ends with `.zip`.
- Validate ZIP magic `PK`.
- Validate expected SHA256 matches actual.
- Validate commit exists locally if publishing from repo server.
- Read/validate version from `apps/mobile/pubspec.yaml`.
- Generate JSON manifest.
- Generate QR SVG using Python `qrcode` if installed.
- Do not restart PM2.
- Do not rebuild web-admin.

**Verification command:**

```bash
scripts/publish-mobile-windows.sh --help
```

**Acceptance Criteria:**

- Script dry-run works.
- Script refuses missing `--expected-sha256` in production.

---

## Task 3.2: Test publish script with a dummy ZIP in staging mode

**Objective:** Verify script behavior without touching production latest Windows release.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
TMP_ZIP=/tmp/mtsn2kolut-cbt-windows-test.zip
python3 - <<'PY'
from pathlib import Path
import zipfile
p = Path('/tmp/mtsn2kolut-cbt-windows-test.zip')
with zipfile.ZipFile(p, 'w') as z:
    z.writestr('README.txt', 'test only')
PY
SHA="$(sha256sum "$TMP_ZIP" | awk '{print $1}')"
scripts/publish-mobile-windows.sh \
  --zip "$TMP_ZIP" \
  --commit "$(git rev-parse --short HEAD)" \
  --base-url https://mtsn2kolut.sch.id \
  --channel staging \
  --app-name 'MTsN 2 Kolut CBT Windows' \
  --arch x64 \
  --version-name 1.0.0 \
  --version-code 1 \
  --expected-sha256 "$SHA" \
  --notes 'Dry run staging test' \
  --dry-run
```

**Acceptance Criteria:**

- Dry-run prints intended paths and SHA.
- No production latest files are overwritten.

---

# Sprint 4 — SvelteKit Release Routes for Windows Files

## Task 4.1: Locate current release file server route

**Objective:** Find existing route that serves `/releases/mobile/*` and extend it for Windows files if needed.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
find apps/web-admin/src/routes -path '*releases*' -type f -print
rg "latest-arm64|releases/mobile|application/vnd.android" apps/web-admin/src
```

**Expected:**

- There is an existing dynamic SvelteKit route or hook serving release files.

**Acceptance Criteria:**

- Exact file path is identified before editing.

---

## Task 4.2: Extend release route content types

**Objective:** Serve Windows ZIP/manifest/checksum/QR from release directory.

**Files:**
- Modify: existing release route file found in Task 4.1

**Rules:**

- `.apk` content type:
  - `application/vnd.android.package-archive`
- `.zip` content type:
  - `application/zip`
- `.json` content type:
  - `application/json; charset=utf-8`
- `.sha256` content type:
  - `text/plain; charset=utf-8`
- `.svg` content type:
  - `image/svg+xml; charset=utf-8`

**Required new paths:**

```text
/release/mobile/latest-windows-x64.zip
/release/mobile/latest-windows-x64.sha256
/release/mobile/latest-windows.json
/release/mobile/latest-windows-qr.svg
```

Actual existing URL prefix is `/releases/mobile`, so required public paths are:

```text
/releases/mobile/latest-windows-x64.zip
/releases/mobile/latest-windows-x64.sha256
/releases/mobile/latest-windows.json
/releases/mobile/latest-windows-qr.svg
```

**Security constraints:**

- Prevent path traversal (`..`).
- Serve only allowlisted filenames/extensions under `/home/servermtsn2kolut/releases/mtsn2kolut-mobile`.
- `Cache-Control: no-store, max-age=0` for latest files.
- `Content-Disposition: attachment` for `.zip`.

**Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

**Acceptance Criteria:**

- Existing Android URLs still work.
- New Windows URLs return correct headers once files exist.

---

# Sprint 5 — CBT Center UI: Android + Windows Download Cards

## Task 5.1: Extend release page manifest types

**Objective:** Allow release page to read both Android and Windows manifests.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/aplikasi-siswa/release/+page.svelte`

**Current behavior:**

- Fetches `/releases/mobile/latest.json` for Android APK.

**New behavior:**

- Fetch Android manifest:
  - `/releases/mobile/latest.json`
- Fetch Windows manifest:
  - `/releases/mobile/latest-windows.json`
- Android card remains unchanged if Windows manifest missing.
- Windows card shows fallback empty state if not published yet.

**Acceptance Criteria:**

- Page does not fail if Windows release is missing.
- Android release remains visible.

---

## Task 5.2: Add Windows download card

**Objective:** Add clear operator UI for Windows ZIP.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/aplikasi-siswa/release/+page.svelte`

**UI copy:**

- Title: `Unduh CBT Windows Terbaru`
- Badge: `Windows x64`
- Button: `Unduh ZIP Windows`
- Warning text:

```text
Untuk Windows, ekstrak ZIP terlebih dahulu lalu jalankan aplikasi di dalam folder hasil ekstrak. Jangan hanya memindahkan file .exe tanpa folder data dan DLL.
```

**Acceptance Criteria:**

- Android and Windows cards are visually distinct.
- Windows instructions clearly say extract ZIP first.
- No backend database access is introduced.

---

## Task 5.3: Update public home shortcut if needed

**Objective:** If homepage currently has only Android download shortcut, decide whether to keep Android only or link to Release Center.

**Files:**
- Inspect/modify: `apps/web-admin/src/routes/+page.svelte`

**Recommended change:**

Change shortcut from direct APK to Release Center:

```text
Download Aplikasi CBT → /asesmen/aplikasi-siswa/release
```

Reason: user can choose Android or Windows there.

**Acceptance Criteria:**

- Public/home shortcut does not imply Android-only if Windows is now available.

---

# Sprint 6 — Documentation and Manual Publish SOP

## Task 6.1: Update release center docs

**Objective:** Extend Android-only docs to Android + Windows.

**Files:**
- Modify: `docs/mobile-apk-release-center.md`
- Or rename/create companion: `docs/mobile-release-center.md`

**Required additions:**

- Windows URL list:

```text
https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.zip
https://mtsn2kolut.sch.id/releases/mobile/latest-windows.json
https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.sha256
https://mtsn2kolut.sch.id/releases/mobile/latest-windows-qr.svg
```

- GitHub Actions run instructions.
- Manual publish steps.
- Rollback steps for Windows ZIP.
- Warning: ZIP must include full Flutter Windows release folder, not only `.exe`.

**Acceptance Criteria:**

- A future operator can build, download, upload, publish, verify, and rollback Windows release.

---

## Task 6.2: Add checksums and validation section

**Objective:** Ensure every Windows ZIP can be verified before distribution.

**Docs content:**

```bash
curl -I https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.zip
curl -fsS https://mtsn2kolut.sch.id/releases/mobile/latest-windows.json | python3 -m json.tool
curl -fsS https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.sha256
sha256sum /home/servermtsn2kolut/releases/mtsn2kolut-mobile/latest-windows-x64.zip
```

Expected headers:

```text
content-type: application/zip
cache-control: no-store, max-age=0
content-disposition: attachment; filename="mtsn2kolut-cbt-latest-windows-x64.zip"
```

**Acceptance Criteria:**

- Validation commands are copy-pasteable.

---

# Sprint 7 — First Windows Build and Manual CBT Center Publish

## Task 7.1: Push workflow branch to GitHub

**Objective:** Make GitHub Actions workflow available.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git add apps/mobile/windows apps/mobile/.metadata .github/workflows/build-mobile-windows.yml docs/mobile-windows-github-actions.md docs/mobile-apk-release-center.md scripts/publish-mobile-windows.sh apps/web-admin/src/routes/asesmen/aplikasi-siswa/release/+page.svelte apps/web-admin/src/routes/+page.svelte
git commit -m "feat(mobile): add Windows release workflow and CBT Center link"
git push origin feature/comprehensive-improvements
```

**Acceptance Criteria:**

- Commit pushed to GitHub.
- Workflow visible under GitHub → Actions.

---

## Task 7.2: Run GitHub Actions manually

**Objective:** Produce first Windows ZIP.

**Manual steps:**

1. Open GitHub repo.
2. Click **Actions**.
3. Select **Build CBT Mobile Windows**.
4. Click **Run workflow**.
5. Select branch.
6. Wait until green.
7. Download artifact `mtsn2kolut-cbt-windows`.

**Acceptance Criteria:**

- ZIP artifact is available.
- SHA256 file is available.

---

## Task 7.3: Upload Windows artifact to server manually

**Objective:** Move ZIP from GitHub artifact to Linux server release workflow.

**Example local command from operator laptop:**

```bash
scp mtsn2kolut-cbt-windows-<commit>-v1.0.0+1.zip servermtsn2kolut@<server-ip>:/tmp/
scp mtsn2kolut-cbt-windows-<commit>-v1.0.0+1.zip.sha256 servermtsn2kolut@<server-ip>:/tmp/
```

If using Telegram/upload manually, save the file under `/tmp/` before publishing.

**Acceptance Criteria:**

- ZIP exists on server.

---

## Task 7.4: Publish Windows ZIP to CBT Center

**Objective:** Make Windows ZIP available at public URL.

**Commands on server:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
ZIP=/tmp/mtsn2kolut-cbt-windows-<commit>-v1.0.0+1.zip
SHA="$(sha256sum "$ZIP" | awk '{print $1}')"
VERSION_RAW="$(python3 - <<'PY'
from pathlib import Path
import re
text = Path('apps/mobile/pubspec.yaml').read_text()
print(re.search(r'^version:\s*([^\s#]+)', text, re.M).group(1))
PY
)"
VERSION_NAME="${VERSION_RAW%%+*}"
VERSION_CODE="${VERSION_RAW#*+}"
scripts/publish-mobile-windows.sh \
  --zip "$ZIP" \
  --commit "$(git rev-parse --short HEAD)" \
  --base-url https://mtsn2kolut.sch.id \
  --channel production \
  --app-name 'MTsN 2 Kolut CBT Windows' \
  --arch x64 \
  --version-name "$VERSION_NAME" \
  --version-code "$VERSION_CODE" \
  --expected-sha256 "$SHA" \
  --notes 'GitHub Actions Windows build lulus|Dipublish ke CBT Center|Operator wajib ekstrak ZIP sebelum menjalankan aplikasi'
```

**Acceptance Criteria:**

- `/releases/mobile/latest-windows-x64.zip` returns 200.
- `/releases/mobile/latest-windows.json` returns 200.
- CBT Center page shows Windows release.

---

# Sprint 8 — Verification and Smoke Tests

## Task 8.1: Verify SvelteKit build and route behavior

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

**Expected:**

- Both pass.

**If build is deployed:**

Because SvelteKit build changes require restart:

```bash
pm2 restart mtsn2kolut-web-admin --update-env
```

Only run restart when deployment approval is given.

---

## Task 8.2: Verify local release URLs

**Commands:**

```bash
curl -fsSI http://127.0.0.1:8021/releases/mobile/latest-windows-x64.zip
curl -fsS http://127.0.0.1:8021/releases/mobile/latest-windows.json | python3 -m json.tool
curl -fsS http://127.0.0.1:8021/releases/mobile/latest-windows-x64.sha256
curl -sS -o /tmp/release-page.html -w '%{http_code} %{redirect_url}\n' http://127.0.0.1:8021/asesmen/aplikasi-siswa/release
```

**Expected:**

- ZIP: HTTP 200, `application/zip`.
- JSON: valid manifest.
- SHA256: matches file.
- Release page: 302 to login if unauthenticated, which is correct auth boundary.

---

## Task 8.3: Verify public release URLs

**Commands:**

```bash
curl -L -fsSI https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.zip
curl -L -fsS https://mtsn2kolut.sch.id/releases/mobile/latest-windows.json | python3 -m json.tool
curl -L -fsS https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.sha256
```

**Expected headers:**

```text
content-type: application/zip
content-disposition: attachment; filename="mtsn2kolut-cbt-latest-windows-x64.zip"
cache-control: no-store, max-age=0
```

---

## Task 8.4: Manual Windows smoke test

**Objective:** Ensure ZIP actually runs on a Windows machine.

**Manual steps on Windows:**

1. Download ZIP from CBT Center.
2. Right-click → Extract All.
3. Open extracted folder.
4. Run `.exe`.
5. Confirm login screen opens.
6. Enter server URL if app asks for it:
   - `https://mtsn2kolut.sch.id`
7. Do not use real exam token unless testing session is prepared.

**Acceptance Criteria:**

- App opens without missing DLL error.
- Login screen appears.
- No SmartScreen/blocking issue beyond normal unsigned-app warning.

---

# Sprint 9 — Optional Automatic Publish from GitHub Actions

Do this later after manual process is proven.

## Task 9.1: Create GitHub Secrets

**Objective:** Allow workflow to upload ZIP to server securely.

**GitHub repository settings:**

Create secrets:

```text
CBT_RELEASE_HOST
CBT_RELEASE_USER
CBT_RELEASE_SSH_KEY
CBT_RELEASE_PORT
```

Do **not** put these values in code or chat logs.

---

## Task 9.2: Add optional deploy job gated by manual input

**Objective:** Let GitHub Actions publish to server only when manually requested.

**Workflow design:**

Use `workflow_dispatch` input:

```yaml
inputs:
  publish_to_server:
    description: 'Publish artifact to CBT Center server'
    required: true
    default: 'false'
    type: choice
    options:
      - 'false'
      - 'true'
```

Deploy job runs only if:

```yaml
if: github.event.inputs.publish_to_server == 'true'
```

**Acceptance Criteria:**

- Normal pushes build artifact only.
- Manual run can publish when explicitly selected.

---

# Rollback Plan

## Android rollback already exists

Use existing history APK and `publish-mobile-apk.sh`.

## Windows rollback

1. Pick a previous ZIP from:

```text
/home/servermtsn2kolut/releases/mtsn2kolut-mobile/history/
```

2. Republish it:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
ZIP=/home/servermtsn2kolut/releases/mtsn2kolut-mobile/history/<old-windows-release>.zip
SHA="$(sha256sum "$ZIP" | awk '{print $1}')"
scripts/publish-mobile-windows.sh \
  --zip "$ZIP" \
  --commit <old-commit> \
  --base-url https://mtsn2kolut.sch.id \
  --channel production \
  --app-name 'MTsN 2 Kolut CBT Windows' \
  --arch x64 \
  --version-name 1.0.0 \
  --version-code 1 \
  --expected-sha256 "$SHA" \
  --notes 'Rollback Windows CBT release'
```

3. Verify URL and checksum.

---

# Commit Strategy

Recommended commits:

1. `feat(mobile): add Flutter Windows runner`
   - `apps/mobile/windows/**`
   - `apps/mobile/.metadata`

2. `ci(mobile): build Windows CBT app with GitHub Actions`
   - `.github/workflows/build-mobile-windows.yml`
   - `docs/mobile-windows-github-actions.md`

3. `feat(web): expose Windows CBT download in release center`
   - `scripts/publish-mobile-windows.sh`
   - release route updates
   - release page updates
   - docs update

4. Optional later:
   - `ci(mobile): add gated server publish for Windows release`

---

# Final Acceptance Criteria

The work is complete when all are true:

- `apps/mobile/windows/` exists and is committed.
- GitHub Actions workflow builds Windows release on `windows-latest`.
- Workflow uploads ZIP + SHA256 artifact.
- Windows ZIP includes full Flutter release folder, not only `.exe`.
- Server has `scripts/publish-mobile-windows.sh`.
- CBT Center page shows both Android APK and Windows ZIP when manifests exist.
- Public URLs work:
  - `https://mtsn2kolut.sch.id/releases/mobile/latest-arm64.apk`
  - `https://mtsn2kolut.sch.id/releases/mobile/latest-windows-x64.zip`
- SHA256 verification passes for both Android and Windows releases.
- PM2 restart is only done if SvelteKit code is changed and deploy is explicitly approved.

---

# Recommended Implementation Order

1. Clean/confirm worktree.
2. Add Windows platform files.
3. Add GitHub Actions build artifact workflow.
4. Push and run first GitHub Actions build.
5. Create server-side Windows publish script.
6. Extend release serving route if needed.
7. Extend CBT Center UI to show Windows card.
8. Build/check web-admin.
9. Deploy/restart web-admin only after approval.
10. Manually publish first Windows ZIP.
11. Smoke test local/public URLs.
12. Test extracted ZIP on actual Windows PC.

---

# Notes for Future Automation

Automatic upload from GitHub Actions to server is intentionally deferred. It needs GitHub Secrets and careful SSH handling. Start with manual artifact download/upload first so release risk stays low.
