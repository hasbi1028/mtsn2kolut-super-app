# Windows CBT Desktop Build Runbook

This runbook documents the Opsi A build path for the MTsN 2 Kolaka Utara CBT student desktop app.

## Build source

- Flutter project: `apps/mobile`
- Windows runner: `apps/mobile/windows`
- GitHub Actions workflow: `.github/workflows/flutter-windows.yml`

## How to build

1. Open the repository on GitHub.
2. Go to **Actions**.
3. Select **Flutter Windows CBT Build**.
4. Click **Run workflow**.
5. Optional: fill `build_name` and `build_number`.
6. Wait for the workflow to finish.
7. Download artifacts:
   - `mtsn2kolut-cbt-windows-x64`
   - `mtsn2kolut-cbt-windows-x86`

## Artifact behavior

- x64 artifact contains `mtsn2kolut-cbt-windows-x64.zip` with the Windows release folder.
- x86 artifact attempts the Flutter x86 target if the runner/toolchain exposes it.
- If x86 is unsupported by Flutter on the GitHub runner, the x86 artifact contains:
  - `mtsn2kolut-cbt-windows-x86-unsupported.txt`
  - `x86-build.log`

Flutter Windows is primarily x64 in modern toolchains. The x86 artifact is kept as an explicit compatibility check so release operators can verify the 32-bit request was attempted without breaking the x64 release.

## Anti-cheat behavior on Windows

The Windows runner implements the same Flutter channels used by Android:

- MethodChannel: `id.sch.mtsn2kolutara.mobile/anti_cheat`
- EventChannel: `id.sch.mtsn2kolutara.mobile/anti_cheat_events`

Windows desktop anti-cheat is best-effort and reports:

- platform: `windows`
- fullscreen state
- minimized state
- focus state
- capture-protection status via `SetWindowDisplayAffinity`

The Dart guard maps these to CBT warning events and local blocking states:

- `windows_not_fullscreen`
- `app_minimized`
- `window_focus_lost`
- existing local lock after repeated violations

## Smoke test checklist

On a Windows test machine:

- Extract the x64 zip.
- Run `mtsn2kolut_cbt.exe`.
- Confirm the app opens in fullscreen/topmost mode.
- Confirm leaving fullscreen/minimizing/focus loss is blocked or reported.
- Confirm login/session flow still reaches the CBT shell.
- Confirm anti-cheat events are visible in existing CBT/proctor telemetry when connected to the official server.

## Security notes

- No signing certificate is configured yet.
- Windows SmartScreen warnings are expected for internal unsigned builds.
- Do not distribute build artifacts outside the madrasah/testing operators.
