# Bank Soal Role E2E Smoke

Browser E2E scaffold untuk memverifikasi Bank Soal dengan akun role nyata di staging/operator environment. Script ini sengaja **tidak menyimpan credential** dan default-nya skip bersih bila environment belum lengkap.

## Scope full-cover

Script `scripts/bank-soal-role-e2e.mjs` mencakup persona:

- `admin`
- `creator` — punya `bank_soal.create`
- `reviewer` — punya `bank_soal.review`/review flow
- `importer` — punya `bank_soal.import`
- `readonly` — hanya baca Bank Soal
- `no-access` — tidak punya akses Bank Soal

Route yang dicakup:

- `/bank-soal`
- `/bank-soal/daftar`
- `/bank-soal/tambah`
- `/bank-soal/verifikasi`
- `/bank-soal/impor`
- `/bank-soal/analisis-butir`
- `/bank-soal/mapel-kd`
- `/bank-soal/pengaturan`

Boundary API unauthenticated:

- `/api/bank-soal/summary` harus `401`
- `/api/bank-soal/questions` harus `401`
- `/api/bank-soal/assets` harus `401`
- `/api/cbt/questions` tidak boleh public `200`

UI permission assertions:

- user `admin`/`creator` harus melihat shortcut/link create soal di halaman yang relevan.
- user `reviewer`/`importer`/`readonly` tidak boleh melihat shortcut/link `/bank-soal/tambah`.
- route privileged harus forbidden/redirect sesuai role.

## Required env

```bash
WEB_ADMIN_BANK_SOAL_E2E_BASE_URL=https://admin.example.sch.id
WEB_ADMIN_BANK_SOAL_E2E_ADMIN_USERNAME=...
WEB_ADMIN_BANK_SOAL_E2E_ADMIN_PASSWORD=...
WEB_ADMIN_BANK_SOAL_E2E_CREATOR_USERNAME=...
WEB_ADMIN_BANK_SOAL_E2E_CREATOR_PASSWORD=...
WEB_ADMIN_BANK_SOAL_E2E_REVIEWER_USERNAME=...
WEB_ADMIN_BANK_SOAL_E2E_REVIEWER_PASSWORD=...
WEB_ADMIN_BANK_SOAL_E2E_IMPORTER_USERNAME=...
WEB_ADMIN_BANK_SOAL_E2E_IMPORTER_PASSWORD=...
WEB_ADMIN_BANK_SOAL_E2E_READONLY_USERNAME=...
WEB_ADMIN_BANK_SOAL_E2E_READONLY_PASSWORD=...
WEB_ADMIN_BANK_SOAL_E2E_NO_ACCESS_USERNAME=...
WEB_ADMIN_BANK_SOAL_E2E_NO_ACCESS_PASSWORD=...
```

Optional:

```bash
WEB_ADMIN_BANK_SOAL_E2E_HEADLESS=false
WEB_ADMIN_BANK_SOAL_E2E_TIMEOUT_MS=20000
WEB_ADMIN_BANK_SOAL_E2E_IGNORE_HTTPS_ERRORS=true
WEB_ADMIN_BANK_SOAL_E2E_ALLOW_SKIP=false # staging/CI wajib fail bila env belum lengkap
```

## Run

Dari `apps/web-admin`:

```bash
npm run smoke:bank-soal:roles
```

Jika Playwright belum tersedia di operator machine:

```bash
npm install --no-save playwright
npx playwright install chromium
npm run smoke:bank-soal:roles
```

## Safety

- Gunakan akun staging/operator terkontrol.
- Jangan commit username/password ke repo.
- Script hanya navigasi/read checks; tidak membuat soal/import data.
- Untuk CI/staging wajib, set `WEB_ADMIN_BANK_SOAL_E2E_ALLOW_SKIP=false` supaya credential/env yang hilang menjadi failure.
