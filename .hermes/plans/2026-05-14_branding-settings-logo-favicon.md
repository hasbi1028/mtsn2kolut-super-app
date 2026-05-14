# Plan — Settings Branding: Logo, Favicon, PWA Icons

## Tujuan
Membuat fitur di Settings untuk mengelola identitas visual aplikasi: logo, favicon, app/PWA icon, warna tema, dan teks branding, dengan backend Go sebagai source of truth dan SvelteKit hanya sebagai BFF/UI.

## Prinsip
- Jangan akses DB dari SvelteKit.
- Simpan config branding di `app_settings` via core-api.
- Asset branding disimpan di backend filesystem, bukan `static/` web-admin.
- Public branding harus punya fallback static agar login/page tetap render jika API gagal.
- Mutasi branding wajib RBAC `settings.branding` dan admin granted via migration.
- Upload harus server-side validated: byte limit, MIME sniffing, image decode config, dimension, square requirements untuk icon.
- SVG upload tetap diblokir kecuali ada sanitizer khusus.
- Dynamic favicon/PWA pakai cache-busting version.

## Scope Sprint Branding 1
1. Backend:
   - Migration permission `settings.branding` + admin grant.
   - Typed `BrandingSettings` di service settings.
   - Public read endpoint `GET /api/public/branding`.
   - Admin endpoints:
     - `GET /api/branding`
     - `PUT /api/branding`
     - `POST /api/branding/assets/{purpose}`
     - `DELETE /api/branding/assets/{purpose}` reset to default.
   - Public asset serve `GET /api/branding/assets/{purpose}` atau route setara.
   - Store uploaded assets in `BRANDING_ASSET_DIR` default `data/branding`.

2. Web-admin/BFF:
   - BFF routes for branding JSON and upload/reset.
   - `/settings/branding` page with cards for identity, logo/mark/favicon/PWA icon, preview, reset.
   - Sidebar menu Settings → Branding.
   - Route access policy.
   - Root layout loads public branding for dynamic head + fallback.
   - Dynamic manifest route using branding with fallback.
   - Sidebar/login/public shell consume branding if available where safe.

3. Verification:
   - `npm --prefix apps/web-admin run check`
   - `npm --prefix apps/web-admin run build`
   - `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`
   - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
   - `go build -o /tmp/core-api-branding-settings ./cmd/api`

## Out of Scope
- Deploy/restart PM2 unless requested separately.
- Advanced image derivative generation/ICO generation if not already available. MVP may require separate uploads per purpose with validation.
- Formal report/document logo replacement unless explicitly wired in a later sprint.

## Swarm Discovery Summary
- Backend: use `app_settings`, `service.Setting`, `handler/setting.go`, upload helpers in `file_security.go`, website media pattern, RBAC migration pattern.
- Frontend: extend settings navigation, use BFF proxy style, upload UI pattern from account avatar, dynamic head/manifest/cache-busting.
- Best practice: purpose allowlist, server validation, fallback defaults, cache-busting version, no SVG uploads.
