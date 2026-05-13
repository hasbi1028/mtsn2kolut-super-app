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

Credential bisa disiapkan manual seperti sebelumnya, atau dibuat dengan seed DB terkontrol di bawah.

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
WEB_ADMIN_BANK_SOAL_E2E_USE_SEEDED_DEFAULTS=true # pakai akun hasil seed di bawah
```

## Controlled seed

Seed manual tersedia untuk membuat 6 akun/persona smoke tanpa menyimpan credential nyata di repo. Script ini idempoten, memakai role RBAC khusus `bank_soal_e2e_*` untuk persona non-admin, mengaktifkan akun seed, memastikan `must_change_password=false`, dan hanya mengubah akun sesuai username seed.

Default username:

- `e2e_bank_soal_admin`
- `e2e_bank_soal_creator`
- `e2e_bank_soal_reviewer`
- `e2e_bank_soal_importer`
- `e2e_bank_soal_readonly`
- `e2e_bank_soal_no_access`

Password **wajib** diberikan lewat environment; repo tidak menyediakan password default. Gunakan password dummy terkontrol khusus E2E dan jangan pakai credential pribadi/produksi:

```bash
BANK_SOAL_E2E_PASSWORD='[REDACTED_E2E_PASSWORD]' \
DATABASE_URL='postgresql://example_user:example_password@example-host:5432/example_db' \
npm run db:seed:bank-soal-e2e-roles
```

Override username/password per persona bila diperlukan:

```bash
BANK_SOAL_E2E_CREATOR_USERNAME=e2e_bs_creator_a \
BANK_SOAL_E2E_CREATOR_PASSWORD='[REDACTED_CREATOR_PASSWORD]' \
BANK_SOAL_E2E_PASSWORD='[REDACTED_SHARED_PASSWORD]' \
DATABASE_URL='postgresql://example_user:example_password@example-host:5432/example_db' \
npm run db:seed:bank-soal-e2e-roles
```

Untuk local development eksplisit tanpa `DATABASE_URL`:

```bash
ALLOW_LOCAL_DATABASE_URL=true \
LOCAL_DATABASE_URL='postgresql://example_user:example_password@localhost:5432/example_db' \
BANK_SOAL_E2E_PASSWORD='[REDACTED_E2E_PASSWORD]' \
npm run db:seed:bank-soal-e2e-roles
```

Permission matrix seed:

- `admin`: role `admin`, akses penuh dari RBAC sistem.
- `creator`: `bank_soal.read`, `bank_soal.create`, `bank_soal.update`, `bank_soal.analytics`.
- `reviewer`: `bank_soal.read`, `bank_soal.review`, `bank_soal.analytics`.
- `importer`: `bank_soal.read`, `bank_soal.import`, `bank_soal.analytics`.
- `readonly`: `bank_soal.read`, `bank_soal.analytics`.
- `no-access`: tidak punya permission `bank_soal.*`.

Setelah seed, smoke bisa dijalankan dengan credential eksplisit atau memakai default seed:

```bash
WEB_ADMIN_BANK_SOAL_E2E_BASE_URL=http://localhost:5173 \
WEB_ADMIN_BANK_SOAL_E2E_USE_SEEDED_DEFAULTS=true \
BANK_SOAL_E2E_PASSWORD='[REDACTED_E2E_PASSWORD]' \
npm --prefix apps/web-admin run smoke:bank-soal:roles
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
