# Internal Analytics Runbook

## Owner

Owner operasional: admin/operator MTsN 2 Kolaka Utara yang diberi permission analytics. Jalur data resmi tetap 100% internal: Web Admin -> BFF -> Go Core API -> PostgreSQL.

## Permissions

- `analytics.read`: membaca dashboard agregat internal.
- `analytics.export`: export laporan agregat.
- `analytics.security_read`: membaca sinyal keamanan yang sudah diagregasi.

Tidak ada public collector, no public collector runtime, no raw event export, dan no third-party analytics.

## Export policy

Export hanya agregat dan memakai header tetap:

`aggregate_date,event_group,event_name,source_surface,role,result,count`

Export tidak boleh memuat metadata mentah, user/session ID, raw IP, raw user agent, token, cookie, NIP/NISN/NIK, device fingerprint, query string, atau payload mentah. CSV harus aman dari formula injection.

## Retention and cleanup

Retention and cleanup memakai `retention_expires_at` dari event internal analytics. Cleanup hanya manual admin/ops invocation only melalui service/ops command yang dilindungi; jangan menjalankan cleanup terhadap live DB tanpa backup dan persetujuan operasional. Readiness memantau `expired_event_backlog_count` dan `oldest_expired_event_at`.

## Smoke checklist

- Web Admin `/settings/analytics` hanya terbuka untuk pengguna berizin.
- BFF `/api/internal-analytics/summary`, `/daily`, dan `/export` mengembalikan 401 saat unauthenticated.
- Core API `/api/internal-analytics/events` tetap JWT protected dengan body cap 16 KiB.
- Export CSV hanya berisi kolom agregat.
- Tidak ada public unauthenticated collector.
- Tidak ada script/dependency analytics pihak ketiga.

## Rollback

Rollback source dapat dilakukan dengan revert commit fitur analytics terakhir, lalu rebuild/restart sesuai runbook deploy umum bila memang sudah dideploy. Jika migration permission belum pernah dijalankan, tidak ada rollback DB. Jika sudah dijalankan, permission `analytics.export` dapat dinonaktifkan via RBAC admin daripada menghapus histori migration.

## Recovery

Jika cleanup gagal, hentikan invocation berikutnya, simpan log error yang sudah disanitasi, cek health/readiness backlog, verifikasi backup PostgreSQL terbaru, lalu ulangi cleanup secara batch kecil setelah penyebab diperbaiki. Jangan restore live DB kecuali lewat prosedur backup/restore resmi.

## 100% internal evidence

- Arsitektur hanya Web Admin -> BFF -> Go Core API -> PostgreSQL.
- Tidak ada third-party analytics.
- Tidak ada public collector.
- Tidak ada raw event export.
- Dashboard dan export hanya agregat.
- PostgreSQL tetap dimiliki Core API; Web Admin tidak direct DB.
