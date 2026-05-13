# CBT Role Smoke Scaffold

This is a staging/manual-friendly smoke scaffold for the web-admin CBT role flows. It does not require secrets in the repo and skips cleanly unless all required environment variables are present.

## Scope

The smoke script checks:

- admin login can reach `/bank-soal/tambah`
- guru login can reach `/bank-soal/tambah`
- admin can reach `/asesmen/kegiatan`
- guru is blocked from `/asesmen/kegiatan`
- retired `/cbt/questions` redirects to `/bank-soal/*`, preserving important query values while dropping retired experiment modes

## Requirements

The app does not ship a Playwright dependency by default. If no compatible Playwright package is installed in the operator environment, the script exits successfully with a skip message.

Required environment variables:

```bash
WEB_ADMIN_SMOKE_BASE_URL=https://admin.example.sch.id
WEB_ADMIN_SMOKE_ADMIN_USERNAME=admin-user
WEB_ADMIN_SMOKE_ADMIN_PASSWORD='[REDACTED_ADMIN_PASSWORD]'
WEB_ADMIN_SMOKE_GURU_USERNAME=guru-user
WEB_ADMIN_SMOKE_GURU_PASSWORD='[REDACTED_GURU_PASSWORD]'
```

Optional environment variables:

```bash
WEB_ADMIN_SMOKE_HEADLESS=false
WEB_ADMIN_SMOKE_TIMEOUT_MS=20000
WEB_ADMIN_SMOKE_IGNORE_HTTPS_ERRORS=true
```

## Run

From `apps/web-admin`:

```bash
npm run smoke:cbt:roles
```

If browser smoke execution is needed on a staging/operator machine, install Playwright there rather than committing browser dependencies solely for this scaffold:

```bash
npm install --no-save playwright
npx playwright install chromium
npm run smoke:cbt:roles
```

## Safety Notes

Use only staging or controlled operator accounts. Do not commit real credentials to `.env`, shell history snippets, or documentation. The script performs read/navigation checks only and does not create CBT data.
