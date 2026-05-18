# Flutter Windows Desktop CBT App Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Membuat aplikasi desktop Windows untuk CBT siswa dari `apps/mobile`, mempertahankan kontrol anti-cheat yang sudah ada, dan menyediakan GitHub Actions manual untuk menghasilkan artifact Windows 64-bit dan jalur kompatibilitas 32-bit bila Flutter/toolchain mendukung.

**Architecture:** Flutter tetap menjadi client CBT siswa. Target Windows ditambahkan ke `apps/mobile/windows` dengan native runner Windows untuk fullscreen/kiosk-like behavior, focus-loss detection, screenshot/privacy prevention best-effort, dan event telemetry ke backend existing. GitHub Actions berjalan di `windows-latest`, build release Windows, zip folder Release, dan upload artifact tanpa installer/signing dulu.

**Tech Stack:** Flutter/Dart, Windows runner C++/Win32, GitHub Actions, PowerShell, Core API CBT existing.

---

## Keputusan Utama

1. **Opsi A dijalankan dulu:** GitHub Actions manual menghasilkan `.zip` artifact, bukan installer.
2. **Anti-cheat dipertahankan:** logic Dart `AntiCheatGuard` tetap dipakai, ditambah implementasi native Windows untuk channel yang sama:
   - `id.sch.mtsn2kolutara.mobile/anti_cheat`
   - `id.sch.mtsn2kolutara.mobile/anti_cheat_events`
3. **64-bit adalah target utama.** Flutter Windows desktop modern secara umum memakai Windows x64 runner. Build 64-bit wajib berhasil.
4. **32-bit dibuat sebagai compatibility track.** Karena dukungan Flutter Windows 32-bit bisa tergantung versi Flutter/toolchain, plan ini membuat gate eksplisit:
   - coba build Win32/x86 jika toolchain mendukung;
   - jika tidak didukung, GitHub Actions tetap memberi artifact/log “32-bit unsupported by current Flutter toolchain”, bukan menggagalkan build 64-bit.
5. **Tidak ada database/migration.** Ini hanya mobile/desktop client + CI.
6. **Tidak ada deploy server/PM2.** Output hanya artifact GitHub Actions.
7. **Tidak ada signing dulu.** Windows SmartScreen warning bisa muncul; signing masuk tahap berikutnya.

---

## Current-State Audit

### Repo facts yang sudah dicek

- Flutter app ada di:
  - `apps/mobile/pubspec.yaml`
- Belum ada folder:
  - `apps/mobile/windows/`
- Anti-cheat Dart sudah ada:
  - `apps/mobile/lib/src/anti_cheat_guard.dart`
- Test anti-cheat sudah ada:
  - `apps/mobile/test/anti_cheat_guard_test.dart`
- CI existing ada di:
  - `.github/workflows/ci.yml`

### Risiko Teknis

- Runner Linux server ini belum punya command `flutter`, jadi build lokal Windows tidak dilakukan di server. Verifikasi build Windows dilakukan di GitHub Actions `windows-latest`.
- Windows Flutter app tidak bisa hanya dibagikan `.exe`; perlu zip folder Release lengkap:
  - `.exe`
  - `flutter_windows.dll`
  - `data/`
  - DLL pendukung lain jika ada.
- Native anti-cheat Windows tidak setara dengan Android Device Owner/kiosk. Ini level **deteksi/deterrence**, bukan prevention penuh. Untuk lab Windows, tetap perlu SOP pengawas dan pembatasan OS/lab bila ujian resmi.

---

## Acceptance Criteria

### Build/Artifact

- GitHub Actions manual workflow tersedia:
  - `.github/workflows/flutter-windows.yml`
- Workflow bisa dijalankan manual via **Run workflow**.
- Artifact 64-bit wajib tersedia:
  - `mtsn2kolut-cbt-windows-x64.zip`
- Artifact 32-bit:
  - jika didukung toolchain: `mtsn2kolut-cbt-windows-x86.zip`
  - jika tidak didukung: upload `mtsn2kolut-cbt-windows-x86-unsupported.txt` sebagai bukti pengecekan.

### App Windows

- Folder `apps/mobile/windows/` ada dan ter-commit.
- App bisa launch sebagai Windows desktop app.
- App name/icon minimal disiapkan untuk CBT MTsN 2 Kolaka Utara.
- Release zip memuat folder runnable, bukan `.exe` tunggal.

### Anti-cheat Windows minimal

- Fullscreen saat mulai sesi ujian / app shell aktif.
- Deteksi focus lost/window deactivate.
- Deteksi minimize/background.
- Deteksi app tidak fullscreen.
- Best-effort screenshot protection via Windows display affinity jika memungkinkan.
- Event anti-cheat tetap memakai taxonomy existing:
  - `window_focus_lost`
  - `app_backgrounded`
  - `anti_cheat_local_lock`
  - optional Windows-specific details dalam `data`.
- Jangan klaim bisa memblokir semua cheat. 32/64-bit Windows desktop tetap butuh pengawasan fisik.

### Tests

- `flutter test` pass di GitHub Actions.
- Existing `anti_cheat_guard_test.dart` tetap pass.
- Tambahkan unit test Dart untuk parsing state Windows bila ada field baru.

---

## Sprint 0 — Discovery & Safe Gate

### Task 0.1: Cek branch dan working tree

**Objective:** Pastikan perubahan dimulai dari repo state yang jelas.

**Files:** none

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git branch --show-current
```

**Expected:** working tree bersih atau hanya file rencana ini.

---

### Task 0.2: Audit Flutter target support

**Objective:** Pastikan target Windows belum ada dan tentukan perintah generate.

**Files:**
- Read: `apps/mobile/pubspec.yaml`
- Inspect: `apps/mobile/windows/`

**Commands:**

```bash
test -f apps/mobile/pubspec.yaml
find apps/mobile -maxdepth 2 -type d -name windows -print
```

**Expected:** `pubspec.yaml` ada; `windows/` mungkin belum ada.

---

### Task 0.3: Catat batasan 32-bit

**Objective:** Jangan menjanjikan 32-bit sebelum toolchain membuktikan bisa.

**Rule:**
- 64-bit Windows build = mandatory.
- 32-bit Windows build = best-effort compatibility artifact.
- Jika `flutter build windows --target-platform windows-x86` atau konfigurasi serupa tidak didukung oleh Flutter version di GitHub runner, workflow tidak boleh menggagalkan x64; harus upload file unsupported note.

---

## Sprint 1 — Enable Flutter Windows Target

### Task 1.1: Generate Windows platform folder

**Objective:** Tambahkan target Windows Flutter ke `apps/mobile`.

**Files:**
- Create: `apps/mobile/windows/**`
- Modify: `apps/mobile/pubspec.lock` jika dependency resolution berubah

**Command lokal bila Flutter tersedia:**

```bash
cd apps/mobile
flutter config --enable-windows-desktop
flutter create --platforms=windows .
```

**Jika Flutter tidak tersedia di server:**
- Jalankan lewat GitHub Actions bootstrap sementara, atau gunakan environment developer yang punya Flutter.
- Commit hasil `apps/mobile/windows/` ke repo.

**Verification:**

```bash
test -f apps/mobile/windows/runner/main.cpp
test -f apps/mobile/windows/CMakeLists.txt
```

---

### Task 1.2: Set nama aplikasi Windows

**Objective:** App tampil sebagai CBT MTsN 2 Kolaka Utara, bukan nama generic `mobile`.

**Files:**
- Modify: `apps/mobile/windows/runner/main.cpp`
- Modify: `apps/mobile/windows/runner/Runner.rc`
- Modify: `apps/mobile/windows/CMakeLists.txt` bila perlu

**Expected labels:**

```text
MTsN 2 Kolaka Utara CBT
```

**Verification:** Build Windows menampilkan executable bernama jelas, misalnya:

```text
mtsn2kolut_cbt.exe
```

Jika Flutter default tetap `mobile.exe`, rename bisa dilakukan saat packaging zip, tanpa memaksa rename target CMake dulu.

---

## Sprint 2 — Windows Anti-cheat Native Bridge

### Task 2.1: Tambahkan Windows anti-cheat method channel

**Objective:** Implementasikan native side untuk channel existing.

**Files:**
- Modify: `apps/mobile/windows/runner/flutter_window.cpp`
- Modify: `apps/mobile/windows/runner/flutter_window.h`
- Optional Create: `apps/mobile/windows/runner/anti_cheat_bridge.cpp`
- Optional Create: `apps/mobile/windows/runner/anti_cheat_bridge.h`

**Channel names harus sama dengan Dart:**

```cpp
id.sch.mtsn2kolutara.mobile/anti_cheat
id.sch.mtsn2kolutara.mobile/anti_cheat_events
```

**Method minimal:**

```text
enableSecureMode
readWindowState
enterFullscreen
exitFullscreenForDebugOnly(optional)
```

**Expected `readWindowState` map:**

```json
{
  "isMultiWindow": false,
  "isPictureInPicture": false,
  "hasWindowFocus": true,
  "secureFlagEnabled": true,
  "isFullscreen": true,
  "isMinimized": false,
  "platform": "windows"
}
```

Catatan:
- Windows tidak punya split-screen/PiP konsep yang sama dengan Android; isi `false` dan tambahkan `isFullscreen`/`isMinimized` sebagai detail Windows.
- Dart existing akan tetap aman karena field tambahan diabaikan bila belum dipakai.

---

### Task 2.2: Implement fullscreen/kiosk-like mode

**Objective:** Saat app berjalan sebagai CBT, window masuk fullscreen borderless/topmost best-effort.

**Files:**
- Modify: `apps/mobile/windows/runner/flutter_window.cpp`
- Optional bridge files dari Task 2.1

**Behavior:**
- Simpan ukuran/style window awal.
- Set fullscreen borderless memakai Win32 API.
- Jangan pakai teknik agresif yang merusak OS user.
- Jangan disable keyboard system-level tanpa kebijakan admin/MDM.

**Verification manual di Windows runner/local:**
- App terbuka fullscreen.
- Alt+Tab/focus loss menghasilkan event warning.

---

### Task 2.3: Implement secure display best-effort

**Objective:** Mencegah capture screen di Windows jika API tersedia.

**Files:**
- Modify: native bridge Windows

**Implementation note:**
- Gunakan Win32 `SetWindowDisplayAffinity` best-effort.
- Prefer flag modern `WDA_EXCLUDEFROMCAPTURE` jika tersedia, fallback `WDA_MONITOR`.
- Jika gagal, jangan crash; laporkan `secureFlagEnabled=false` supaya proctor telemetry tahu perlindungan tidak aktif.

**Important:** Ini bukan deteksi screenshot sempurna. Ini hanya prevention best-effort.

---

### Task 2.4: Emit event focus/minimize/fullscreen violation

**Objective:** Event native Windows masuk ke Dart stream melalui EventChannel.

**Files:**
- Modify: `apps/mobile/windows/runner/flutter_window.cpp`
- Optional: bridge cpp/h

**Events:**
- window deactivated → `hasWindowFocus=false`
- window activated → `hasWindowFocus=true`
- minimize → `isMinimized=true`, maps to background/focus lost
- not fullscreen while exam active → additional data `isFullscreen=false`

**Verification:** Add debug logging only if safe and non-secret; remove noisy logs before final.

---

## Sprint 3 — Dart Adaptation for Desktop State

### Task 3.1: Extend `AntiCheatWindowState` without breaking Android

**Objective:** Dart state can hold Windows-specific fields while keeping Android tests passing.

**Files:**
- Modify: `apps/mobile/lib/src/anti_cheat_guard.dart`
- Modify: `apps/mobile/test/anti_cheat_guard_test.dart`

**Add fields:**

```dart
final bool isFullscreen;
final bool isMinimized;
final String platform;
```

**Defaults:**

```dart
isFullscreen = true
isMinimized = false
platform = 'unknown'
```

**Behavior:**
- `shouldBlockInteraction` true if `!isFullscreen` or `isMinimized` while in exam shell.
- `primaryReason`:
  - minimized → `app_backgrounded`
  - focus lost → `window_focus_lost`
  - not fullscreen → `window_focus_lost` or `desktop_not_fullscreen` only if backend allows it.

**Preferred safe approach:** keep event type/reason existing (`window_focus_lost`) and put `is_fullscreen=false` in event data to avoid backend schema surprises.

---

### Task 3.2: Add tests for Windows state parsing

**Objective:** Prove Windows fields parse and block safely.

**Files:**
- Modify: `apps/mobile/test/anti_cheat_guard_test.dart`

**Test cases:**

```dart
test('parses native Windows window state map', () { ... });
test('blocks minimized Windows app as backgrounded', () { ... });
test('blocks desktop app when fullscreen is lost', () { ... });
```

**Command:**

```bash
cd apps/mobile
flutter test test/anti_cheat_guard_test.dart
```

Expected: pass.

---

## Sprint 4 — GitHub Actions Build Artifact

### Task 4.1: Add manual workflow for Windows x64

**Objective:** Build Windows release artifact manually from GitHub Actions.

**Files:**
- Create: `.github/workflows/flutter-windows.yml`

**Workflow skeleton:**

```yaml
name: Flutter Windows CBT Build

on:
  workflow_dispatch:
    inputs:
      build_name:
        description: 'Build name/version label'
        required: false
        default: 'manual'

permissions:
  contents: read

jobs:
  windows-x64:
    name: Windows x64 release
    runs-on: windows-latest
    defaults:
      run:
        working-directory: apps/mobile
    steps:
      - uses: actions/checkout@v4
      - uses: subosito/flutter-action@v2
        with:
          channel: stable
          cache: true
      - name: Flutter doctor
        run: flutter doctor -v
      - name: Enable Windows desktop
        run: flutter config --enable-windows-desktop
      - name: Pub get
        run: flutter pub get
      - name: Analyze
        run: flutter analyze
      - name: Tests
        run: flutter test
      - name: Build Windows x64
        run: flutter build windows --release
      - name: Package x64
        shell: pwsh
        run: |
          $release = "build/windows/x64/runner/Release"
          if (!(Test-Path $release)) { throw "Release folder not found: $release" }
          Compress-Archive -Path "$release/*" -DestinationPath "../../mtsn2kolut-cbt-windows-x64.zip" -Force
      - name: Upload x64 artifact
        uses: actions/upload-artifact@v4
        with:
          name: mtsn2kolut-cbt-windows-x64
          path: mtsn2kolut-cbt-windows-x64.zip
```

**Verification:** workflow completes and artifact downloadable.

---

### Task 4.2: Add x86/32-bit compatibility job

**Objective:** Memenuhi permintaan 32-bit dengan proof-based build, tanpa merusak x64.

**Files:**
- Modify: `.github/workflows/flutter-windows.yml`

**Policy:**
- Job `windows-x86` boleh `continue-on-error: true` atau internal fallback.
- Jika Flutter/toolchain mendukung x86, upload zip x86.
- Jika tidak mendukung, upload `.txt` yang menjelaskan unsupported.

**Workflow concept:**

```yaml
  windows-x86:
    name: Windows x86 compatibility check
    runs-on: windows-latest
    continue-on-error: true
    defaults:
      run:
        working-directory: apps/mobile
    steps:
      - uses: actions/checkout@v4
      - uses: subosito/flutter-action@v2
        with:
          channel: stable
          cache: true
      - name: Enable Windows desktop
        run: flutter config --enable-windows-desktop
      - name: Pub get
        run: flutter pub get
      - name: Try build Windows x86
        id: build_x86
        shell: pwsh
        run: |
          $ErrorActionPreference = "Continue"
          flutter build windows --release --target-platform windows-x86
          if ($LASTEXITCODE -ne 0) {
            "Flutter/toolchain pada runner ini belum mendukung Windows 32-bit/x86 build untuk app ini. x64 artifact tetap valid." | Out-File ../../mtsn2kolut-cbt-windows-x86-unsupported.txt -Encoding utf8
            exit 0
          }
      - name: Package x86 if available
        shell: pwsh
        run: |
          $candidates = @(
            "build/windows/x86/runner/Release",
            "build/windows/runner/Release"
          )
          $release = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
          if ($release) {
            Compress-Archive -Path "$release/*" -DestinationPath "../../mtsn2kolut-cbt-windows-x86.zip" -Force
          }
      - name: Upload x86 artifact or unsupported note
        uses: actions/upload-artifact@v4
        with:
          name: mtsn2kolut-cbt-windows-x86
          path: |
            mtsn2kolut-cbt-windows-x86.zip
            mtsn2kolut-cbt-windows-x86-unsupported.txt
          if-no-files-found: error
```

**Important:** Implementer harus menyesuaikan command x86 berdasarkan Flutter version aktual di runner. Jangan hardcode klaim berhasil sebelum Actions membuktikan.

---

### Task 4.3: Add artifact README

**Objective:** Guru/admin tahu cara menjalankan hasil zip.

**Files:**
- Create: `apps/mobile/windows-artifact/README-WINDOWS-CBT.txt` or generate in workflow

**Content minimal:**

```text
MTsN 2 Kolaka Utara CBT Windows

Cara pakai:
1. Extract zip ke folder lokal.
2. Jalankan file .exe di folder hasil extract.
3. Jangan pindahkan .exe keluar dari folder karena butuh data/ dan DLL pendukung.
4. Jika Windows menampilkan Unknown Publisher, pilih More info > Run anyway untuk testing internal.
5. Gunakan hanya saat diarahkan pengawas.
```

Workflow harus memasukkan README ke zip.

---

## Sprint 5 — Validation & Review

### Task 5.1: Local static validation

**Objective:** Validasi file YAML dan perubahan Dart.

**Commands:**

```bash
git diff --check
python3 - <<'PY'
import yaml
from pathlib import Path
for p in Path('.github/workflows').glob('*.yml'):
    yaml.safe_load(p.read_text())
    print('ok', p)
PY
```

If Python `yaml` unavailable, use Node/package existing or rely on GitHub Actions parse.

---

### Task 5.2: GitHub Actions dry run via push/PR

**Objective:** Bukti build Windows dari runner resmi.

**Steps:**
1. Commit perubahan.
2. Push branch ke GitHub jika auth tersedia.
3. Jalankan workflow manual.
4. Download artifact x64.
5. Catat apakah x86 berhasil atau unsupported.

**Expected:**
- `mtsn2kolut-cbt-windows-x64.zip` tersedia.
- `mtsn2kolut-cbt-windows-x86.zip` atau `mtsn2kolut-cbt-windows-x86-unsupported.txt` tersedia.

---

### Task 5.3: Manual Windows smoke test

**Objective:** Pastikan artifact bisa jalan di Windows nyata.

**Checklist:**
- Extract zip.
- Launch `.exe`.
- Login screen tampil.
- Masuk sesi test/dummy bila tersedia.
- Fullscreen aktif.
- Alt+Tab/focus lost memunculkan warning/telemetry.
- Minimize atau keluar fokus ditahan/dicatat.
- Submit answer tetap sinkron.
- Close/reopen tidak menghilangkan jawaban lokal yang belum sync.

---

## Rollback Plan

Karena ini hanya client build + CI:

1. Revert commit workflow/windows target jika build bermasalah:
   ```bash
   git revert <commit>
   ```
2. Tidak ada PM2/server restart.
3. Tidak ada DB rollback.
4. Jika artifact sudah didistribusikan dan bermasalah, tarik link/download dan distribusikan versi sebelumnya.

---

## Future Sprint Setelah Opsi A

### Opsi B — GitHub Release

- Trigger tag:
  - `mobile-windows-v1.0.0`
- Upload artifact ke GitHub Release.

### Opsi C — Installer + signing

- Buat installer `.msix`/`.msi`/setup `.exe`.
- Tambah icon resmi.
- Tambah code signing certificate via GitHub Secrets.

### Managed Lab Hardening

Untuk ujian resmi di lab Windows, pertimbangkan:
- akun Windows khusus ujian;
- block akses browser/file explorer via policy lab;
- allowlist app CBT;
- jaringan/SSID/VLAN CBT;
- proctor dashboard dan SOP berita acara.

---

## Done Definition

Plan ini dianggap selesai diimplementasikan bila:

- `apps/mobile/windows/` ada dan ter-commit.
- `.github/workflows/flutter-windows.yml` ada dan valid.
- `flutter test` pass di workflow.
- `windows-x64` build sukses dan artifact zip bisa didownload.
- `windows-x86` menghasilkan artifact zip atau unsupported note yang jelas.
- Anti-cheat Android existing tidak rusak.
- Anti-cheat Windows minimal berjalan: fullscreen, focus/minimize detection, secure display best-effort.
- Tidak ada secrets/credentials di commit.

---

## Implementation Notes — 2026-05-18

Sprint 0–5 implemented in this branch:

- Confirmed `apps/mobile/windows/` exists and is committed candidate for Windows desktop runner.
- Added Windows native anti-cheat bridge using the existing Flutter channels:
  - `id.sch.mtsn2kolutara.mobile/anti_cheat`
  - `id.sch.mtsn2kolutara.mobile/anti_cheat_events`
- Added best-effort Windows policies:
  - fullscreen/topmost window mode;
  - focus/minimize/fullscreen telemetry;
  - `SetWindowDisplayAffinity` capture protection with fallback;
  - Dart warning reasons for `windows_not_fullscreen` and `app_minimized`.
- Added GitHub Actions manual workflow `.github/workflows/flutter-windows.yml`:
  - x64 release zip artifact;
  - x86 compatibility attempt with unsupported-note artifact if Flutter does not expose x86.
- Added operator runbook `docs/runbooks/windows-cbt-desktop-build.md`.

Validation available on this Linux server:

- `git diff --check` — PASS.
- Static assertions for workflow/channel/file coverage — PASS.

Validation deferred to GitHub Actions/Windows runner because Flutter SDK is not installed on this production server:

- `flutter analyze`.
- `flutter test`.
- `flutter build windows --release`.
