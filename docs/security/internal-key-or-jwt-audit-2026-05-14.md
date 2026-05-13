# Audit InternalKeyOrJWT — Sprint Security Hardening 1

Tanggal: 2026-05-14

## Scope

Audit penggunaan middleware `InternalKeyOrJWT` pada Go core-api untuk memastikan internal key tidak menjadi bypass pada route user-facing/sensitif.

## Hasil pencarian

Query repo:

```bash
rg "InternalKeyOrJWT" services/core-api
```

Temuan:

- `services/core-api/internal/middleware/auth.go` — definisi middleware.
- `services/core-api/internal/middleware/more_test.go` — unit test middleware.

Tidak ada pemakaian production route di `cmd/api/main.go` atau handler route lain pada saat audit ini.

## Keputusan Sprint 1

- `InternalKeyOrJWT` tidak dipakai pada route user-facing.
- Middleware tetap dipertahankan untuk kompatibilitas test/kemungkinan kebutuhan service internal masa depan.
- Jika dipakai nanti, wajib:
  - hanya untuk route internal/non-user-facing,
  - memakai key per service/scope,
  - mencatat actor internal pada audit log,
  - tidak digabung dengan permission middleware yang mengandalkan claims user,
  - review security sebelum deploy.

## Perubahan terkait

- JWT middleware kini dipin ke `HS256`; token HMAC selain HS256 ditolak.
