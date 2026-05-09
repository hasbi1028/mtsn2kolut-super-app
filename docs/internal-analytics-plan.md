# Rencana Internal Analytics MTsN 2 Kolaka Utara

Status: Tahap/Fase 6, dashboard read model internal. Fase 0 menetapkan arah, batas, dan readiness. Fase 1 menambahkan schema analytics internal melalui migration draft dan sqlc query contract saja. Fase 2 menambahkan endpoint ingestion minimum di Core API. Fase 3 menambahkan BFF proxy internal. Fase 4 menambahkan scaffold public website yang fail-closed. Fase 5 menambahkan instrumentasi internal web-admin bernilai tinggi. Fase 6 menambahkan read model agregat dan halaman dashboard internal. Tidak ada deploy, live migration, restart PM2, public unauthenticated collector, raw event export, atau dependency analytics pihak ketiga.

Catatan baseline: Fase 0 tidak menambahkan runtime ingestion table dan Fase 0 tidak menambahkan migration. Batas itu sudah berubah secara staged pada Fase 1-6 melalui migration draft, endpoint internal, BFF proxy internal, helper fail-closed, instrumentasi internal, dan dashboard agregat.

## Prinsip Utama

- **internal-only, no third-party analytics**: seluruh pengukuran aktivitas tetap berada di sistem MTsN 2 Kolaka Utara. Tidak memakai Google Analytics, Meta Pixel, Plausible Cloud, PostHog Cloud, Hotjar, CDN tracker, atau beacon pihak ketiga.
- **privacy-first**: analytics dipakai untuk evaluasi layanan sekolah, keamanan operasional, dan perbaikan alur kerja. Analytics bukan alat pemantauan personal yang mengumpulkan rahasia, kredensial, atau identitas sensitif penuh.
- **PostgreSQL/Core API ownership**: semua data analytics masa depan hanya boleh ditulis dan dibaca melalui `services/core-api`. PostgreSQL tetap dimiliki Core API. `apps/web-admin` hanya BFF/UI dan tidak boleh menulis langsung ke database.
- **event allowlist**: hanya event yang tercatat di `docs/internal-analytics-event-catalog.md` yang boleh dikirim. Event baru harus masuk dokumen katalog dulu sebelum implementasi runtime.
- **separation from audit_logs**: analytics tidak menggantikan `audit_logs`. Audit tetap menjadi catatan mutasi keamanan/kepatuhan. Analytics hanya agregasi aktivitas produk dan sinyal operasional yang sudah disanitasi.
- **phased plan**: implementasi wajib bertahap dari kontrak, schema, ingestion, dashboard, retensi, sampai review keamanan. Tidak boleh langsung menambahkan tracking luas tanpa guard dan uji.

## Arsitektur Internal

Alur target:

1. **SvelteKit/BFF (`apps/web-admin`)**
   - UI admin/guru/staf dan public website hanya mengirim event allowlisted ke BFF route analytics yang akan dibuat pada fase berikutnya.
   - BFF meneruskan JWT user asli untuk event internal setelah login.
   - Public website hanya mengirim event publik minim identitas, tanpa raw IP dan tanpa raw user agent dari browser.
2. **Go Core API (`services/core-api`)**
   - Menjadi satu-satunya pemilik validasi event, sanitasi metadata, rate limit, role gate, dan penyimpanan.
   - Handler tetap tipis: parse request, panggil service analytics, map response.
   - Service analytics masa depan wajib menolak event di luar allowlist dan metadata yang mengandung forbidden sensitive keys.
3. **PostgreSQL**
   - Menyimpan event dan agregat hanya setelah ada migration resmi.
   - Tidak ada runtime database access dari Web Admin, worker, atau Flutter.
   - Retensi dan agregasi wajib dirancang sejak schema pertama, bukan ditambahkan belakangan.
4. **Web Admin Dashboard**
   - Dashboard internal akan membaca agregat melalui Core API.
   - Tampilan awal fokus pada tren publik, kesehatan operasional modul, dan sinyal keamanan yang sudah diringkas.
   - Export hanya untuk role dengan permission eksplisit.

## Batas Deployment

- Frontend, backend, dan worker tetap deployment terpisah sesuai 3 VPS target.
- Analytics public/internal tidak boleh melewati PUSAKA worker.
- `services/pusaka-worker` tetap API client untuk PUSAKA saja dan tidak menjadi collector analytics.
- Tidak ada SDK analytics pihak ketiga di frontend, backend, worker, atau mobile.
- Deploy order masa depan tetap: backend code, migration, backend restart dan health check, frontend, worker bila perlu.
- Fase 0 ini tidak melakukan deploy, tidak menjalankan migration, tidak mengubah PM2, dan tidak menulis data produksi.

## Catatan Implementasi Fase 1

Fase 1 menambahkan schema analytics internal sebagai kontrak data-layer awal, bukan runtime tracking. Perubahan dibatasi pada:

- migration draft `services/core-api/db/migrations/083_internal_analytics_schema.sql` untuk tabel event mentah dan agregat harian internal analytics.
- sqlc query source `services/core-api/db/queries/internal_analytics.sql` untuk create analytics event, get event by id, list events for rollup, delete expired events, upsert daily aggregate, dan list daily aggregates.
- guard test murah yang memastikan schema privacy-first, retensi, indeks, agregat, pemisahan dari `audit_logs`, dan kontrak query tetap ada.

Batas Fase 1:

- tidak menambahkan ingestion API.
- tidak menambahkan BFF route.
- tidak menambahkan frontend tracking.
- tidak menambahkan worker/runtime handler.
- tidak menjalankan migration live.
- tidak deploy.
- tidak restart PM2.

## Catatan Implementasi Fase 2

Fase 2 - Core API ingestion minimum menambahkan jalur tulis awal di `services/core-api` saja. Endpoint ingestion bersifat JWT protected untuk seluruh event pada fase ini sehingga tidak ada public unauthenticated collector. Handler memakai body cap kecil, strict JSON object decoding, dan response receipt aman yang hanya berisi id event, nama event, group, `occurred_at`, dan `retention_expires_at`.

Validasi Fase 2 berada di service Core API:

- event allowlist hardcoded dari `docs/internal-analytics-event-catalog.md`.
- `event_group` harus cocok dengan group katalog event.
- `source_surface` wajib salah satu dari `public_website`, `web_admin`, `core_api`, `mobile_app`, atau `system`.
- metadata wajib object/map, punya batas ukuran, dan ditolak bila mengandung forbidden sensitive keys secara case-insensitive termasuk variasi token, cookie, authorization, secret/API key, NIK/NIP/NISN, device fingerprint, kredensial PUSAKA, raw IP, raw user agent, SQL, stack trace, request body, response body, atau raw payload.
- `retention_expires_at` default memakai server time: public 90 hari, internal 180 hari, dan security 180 hari.

Batas Fase 2:

- tidak menambahkan dashboard/read API.
- tidak menambahkan export API.
- tidak menambahkan retention job atau rollup runtime.
- tidak menambahkan BFF route.
- tidak menambahkan frontend tracking.
- tidak menjalankan migration live.
- tidak deploy.
- tidak restart PM2.

## Catatan Implementasi Fase 3-6

### Fase 3 - BFF proxy contract

Fase 3 menambahkan route BFF `POST /api/internal-analytics/events` di `apps/web-admin` sebagai proxy tipis ke Core API `POST /api/internal-analytics/events`.

Kontrak runtime:

- route BFF membutuhkan session melalui `event.locals.user`; unauthenticated request ditolak `401` sebelum proxy.
- BFF membaca JSON dengan body cap kecil dan meneruskan payload ke Core API tanpa direct DB access.
- BFF menggunakan pola proxy existing sehingga internal event melakukan forward JWT user asli, bukan `X-Internal-Key`.
- BFF tidak mencatat raw metadata, tidak menambahkan public unauthenticated route, dan tidak memakai third-party script.

### Fase 4 - Public website instrumentation

Core API ingestion masih JWT protected. Karena itu public runtime deferred dan fail-closed sampai kebijakan public collector disetujui.

Kontrak Fase 4:

- helper public hanya membuat preview payload tersanitasi untuk event allowlisted.
- public helper tidak mengirim request jaringan.
- tidak membuka unauthenticated public collector.
- tidak ada third-party analytics package, script, pixel, atau browser beacon.
- public payload tidak membawa raw URL query, raw user agent, token, cookie, NIK/NIP/NISN, device fingerprint, atau isi form.

### Fase 5 - Internal app instrumentation

Fase 5 menambahkan helper `source_surface web_admin` dan instrumentasi minimal pada halaman bernilai tinggi:

- `dashboard.view` pada dashboard utama.
- `bank_soal.list_view` pada root Bank Soal.
- `asesmen.hub_view` pada hub Asesmen.
- `pusaka.dashboard_view` pada dashboard PUSAKA.
- `users.list_view` pada Manajemen User.
- `rbac.roles_view` pada Manajemen RBAC.

Kontrak helper:

- hanya event allowlisted yang dikirim.
- analytics failure fail-silent dan tidak merusak UI.
- tidak mengirim raw URL query, raw user agent, token, cookie, NIP/NISN/NIK, device fingerprint, PUSAKA credential, request body, response body, SQL, atau stack trace.
- tidak mengirim raw user agent dari browser.
- instrumentasi berjalan dari `onMount`, bukan SSR side effect.

### Fase 6 - Dashboard read model

Fase 6 menambahkan Core API read endpoint agregat:

- `GET /api/internal-analytics/summary`
- `GET /api/internal-analytics/daily`

Keduanya membaca agregat dari `internal_analytics_daily_aggregates`, bukan raw event rows. Web Admin menambahkan BFF GET route dengan path yang sama dan halaman `/settings/analytics` untuk menampilkan summary, top event, dan tren harian.

Guard dan batas:

- read endpoint memakai `analytics.read` dengan fallback admin transisi.
- `analytics.security_read` disiapkan sebagai permission terpisah untuk panel security bila nanti diekspos.
- dashboard tidak expose raw event metadata, actor user id, raw event body, atau raw event export.
- Fase 6 tidak menambahkan raw event export, public collector, deploy, live migration, atau restart PM2.

## RBAC Permissions Yang Direncanakan

Permission berikut adalah kontrak rencana. Fase 0 tidak menambah seed permission atau route guard runtime.

| Permission | Rencana akses | Catatan |
|------------|---------------|---------|
| `analytics.read` | Membaca dashboard analytics internal dan ringkasan publik. | Cocok untuk admin dan staf/operator yang diberi mandat. |
| `analytics.export` | Mengekspor laporan analytics yang sudah diagregasi. | Harus dibatasi, tercatat di audit, dan tidak berisi field sensitif. |
| `analytics.manage` | Mengelola konfigurasi analytics internal seperti retensi, allowlist aktif, dan sampling. | Admin terbatas. Perubahan harus masuk audit. |
| `analytics.security_read` | Membaca panel sinyal keamanan analytics. | Untuk admin atau staf keamanan yang ditunjuk. Tidak membuka rahasia mentah. |

Aturan RBAC:

- Route analytics masa depan wajib memakai permission eksplisit, bukan hanya role nama besar.
- Fallback role lama boleh ada hanya sebagai masa transisi dan harus terdokumentasi.
- Export wajib membutuhkan `analytics.export`, meskipun user sudah punya `analytics.read`.
- Panel security membutuhkan `analytics.security_read`, bukan otomatis terbuka untuk semua pembaca analytics.

## Privacy-First Rules

- Jangan simpan password, token, cookie, authorization header, secret, API key, kredensial PUSAKA, raw IP, raw user agent, atau device fingerprint.
- Jangan simpan NIK, NIP, atau NISN penuh di event analytics. Jika perlu korelasi operasional, gunakan ID internal yang sudah ada dan role-gated, atau hash harian yang tidak dapat dibalik setelah disetujui fase keamanan.
- Metadata harus berbentuk allowlist per event group, bukan dump payload request.
- Public website analytics harus tetap agregat dan minim identitas.
- Internal app analytics boleh membawa `module`, `route_group`, `role`, `permission_code`, atau status hasil aksi selama tidak membawa isi jawaban, rahasia, atau identitas sensitif penuh.
- Error analytics hanya menyimpan kode/status dan kategori yang aman, bukan stack trace, SQL, body request, atau pesan error internal mentah.
- Export harus berisi agregat, bukan event mentah per individu, kecuali ada kebutuhan investigasi yang disetujui dan dijaga permission khusus.

## Pemisahan Dari audit_logs

`audit_logs` tetap menjadi sumber catatan resmi untuk mutasi dan aksi keamanan:

- login, logout, refresh, revoke session, perubahan role, perubahan permission, CRUD penting, dan aksi administratif tetap dicatat sebagai audit.
- analytics dapat menghitung tren dari event operasional, tetapi tidak menjadi bukti kepatuhan utama.
- analytics tidak boleh menghapus atau melemahkan kewajiban audit middleware.
- analytics security panel boleh merujuk agregasi sinyal dari audit di masa depan, tetapi penyimpanan dan kontrak maknanya tetap terpisah.

Perbedaan kontrak:

| Area | audit_logs | internal analytics |
|------|------------|--------------------|
| Tujuan | Bukti mutasi, keamanan, kepatuhan | Tren penggunaan, kesehatan alur, sinyal operasional |
| Detail | Aksi spesifik yang wajib diaudit | Event allowlisted yang disanitasi |
| Akses | Admin/security sesuai kebijakan audit | Permission analytics terpisah |
| Retensi | Mengikuti kebijakan audit | Agregasi dan expiry lebih agresif |
| Export | Bukti audit terbatas | Laporan agregat terbatas |

## Prinsip Retensi

- Retensi event mentah harus singkat dan punya expiry otomatis sejak fase schema pertama.
- Agregat harian/bulanan boleh disimpan lebih lama karena sudah mengurangi risiko identitas.
- Public analytics sebaiknya lebih cepat diagregasi daripada internal app analytics.
- Security analytics boleh punya retensi berbeda, tetapi tetap tanpa rahasia mentah.
- Retensi harus dapat dijelaskan di dashboard admin dan runbook operator.
- Delete/rollup job masa depan harus idempotent, terjadwal, dan terukur di health/readiness.

Rencana default yang harus divalidasi pada fase schema:

- event mentah public: 30-90 hari.
- event mentah internal: 90-180 hari bila benar-benar diperlukan.
- agregat harian: 12-24 bulan.
- agregat bulanan: sesuai kebutuhan laporan sekolah.

## Batas Data Awal

Data yang boleh direncanakan:

- route group, event name, module, role, permission code, result status, response class, duration bucket, date bucket, content category, file type, CTA key, export type, and sanitized error category.

Data yang tidak boleh direncanakan tanpa review keamanan lanjutan:

- raw request body, raw query string, full URL dengan parameter sensitif, full IP, raw UA, exact GPS, isi jawaban siswa, isi soal penuh, isi dokumen, kredensial PUSAKA, secret, atau identifier nasional penuh.

## Staged Phases 1-10

Fase 0 ini hanya membuat kontrak. Tahap berikutnya wajib tetap kecil dan dapat direview:

1. **Fase 1 - Schema design dan migration draft**
   - Rancang tabel event/aggregate internal dengan retensi sejak awal.
   - Tambahkan migration hanya setelah katalog event stabil.
   - Tambahkan sqlc query eksplisit, tanpa ORM.
2. **Fase 2 - Core API ingestion minimum**
   - Tambahkan endpoint ingestion internal untuk event allowlisted.
   - Validasi metadata per event group dan tolak forbidden sensitive keys.
   - Endpoint awal bersifat JWT protected, memakai body cap, dan belum membuka public unauthenticated ingestion.
3. **Fase 3 - BFF proxy contract**
   - Tambahkan route BFF yang meneruskan event ke Core API.
   - Forward JWT user asli untuk internal app event.
   - Public event tetap minim identitas.
4. **Fase 4 - Public website instrumentation**
   - Instrumentasi page view, CTA, dan download publik yang sudah allowlisted.
   - Tidak ada third-party script.
   - Pastikan consent/copy publik sesuai kebutuhan sekolah.
5. **Fase 5 - Internal app instrumentation**
   - Instrumentasi dashboard, bank soal, asesmen, PUSAKA, users/RBAC, dan security event sesuai katalog.
   - Mulai dari event agregat bernilai tinggi, bukan semua klik.
6. **Fase 6 - Dashboard read model**
   - Tambahkan API read aggregate dan dashboard Web Admin.
   - Terapkan `analytics.read` dan `analytics.security_read`.
   - Tampilkan skeleton/loading sesuai baseline UI.
7. **Phase 7 - Export dan reporting**
   - Tambahkan export agregat dengan `analytics.export`.
   - Catat export sebagai audit event.
   - Pastikan export tidak memuat event mentah sensitif.
8. **Phase 8 - Retention, rollup, dan cleanup**
   - Jalankan job rollup/cleanup backend.
   - Tambahkan health/readiness untuk backlog cleanup.
   - Dokumentasikan recovery bila cleanup gagal.
9. **Phase 9 - Security review dan abuse hardening**
   - Review forbidden sensitive keys, rate limit, trusted proxy behavior, dan payload rejection.
   - Tambahkan test regresi untuk bocor metadata.
   - Pastikan analytics tidak menjadi jalur eksfiltrasi.
10. **Phase 10 - Operational readiness dan handoff**
   - Lengkapi runbook, smoke checklist, dashboard owner, retensi final, dan rollback.
   - Tetapkan siapa yang boleh membaca, export, manage, dan membaca security analytics.
   - Siapkan bukti bahwa implementasi tetap 100% internal.

## Readiness Gate Fase 0

Fase 0 dianggap selesai bila:

- `docs/internal-analytics-plan.md` menjelaskan arsitektur internal-only dan no third-party.
- `docs/internal-analytics-event-catalog.md` menetapkan event allowlist dan forbidden sensitive keys.
- Docs guard Web Admin memastikan kontrak privacy-first, Core API/PostgreSQL ownership, RBAC analytics permissions, audit_logs separation, dan phased plan tetap ada.
- Tidak ada migration, API runtime, BFF route, dependency, tracking frontend, deploy, atau PM2 restart yang ikut berubah.


## Status Implementasi Fase 7-10

Status: Fase 10 selesai. Implementasi akhir tetap internal-only, no third-party analytics, privacy-first, dan berjalan melalui SvelteKit/BFF -> Go Core API -> PostgreSQL tanpa akses DB langsung dari Web Admin.

### Fase 7 - Export dan reporting

- Export hanya CSV aggregate summary/daily counts only melalui Core API dan BFF `/api/internal-analytics/export`.
- Permission `analytics.export` ditambahkan di migration `084_internal_analytics_permissions.sql` dan grant awal admin.
- Export tidak memuat metadata, `actor_user_id`, `session_id`, raw IP, raw user agent, token, cookie, NIP/NISN/NIK, atau raw payload.
- CSV injection safe: sel berawalan `=`, `+`, `-`, `@`, tab, CR, atau LF diberi prefix aman.
- Permintaan export dicatat sebagai event audit analytics `security.export_requested` dengan metadata agregat aman saja.

### Fase 8 - Retention, rollup, dan cleanup

- Rollup dan cleanup backend memakai `retention_expires_at`, `ListInternalAnalyticsEventsForRollup`, `UpsertInternalAnalyticsDailyAggregate`, dan `DeleteExpiredInternalAnalyticsEvents`.
- Cleanup bersifat manual admin/ops invocation only; tidak ada scheduler otomatis baru di fase ini.
- Health/readiness menampilkan `expired_event_backlog_count` dan `oldest_expired_event_at` sebagai agregat operasional tanpa metadata mentah.
- Recovery bila cleanup gagal didokumentasikan di `docs/internal-analytics-runbook.md`.

### Fase 9 - Security review dan abuse hardening

- Forbidden sensitive keys diperluas untuk token, authorization, bearer, cookie, NIP, NISN, NIK, device fingerprint variants, raw user agent, `query_string`, `rawQuery`, dan `full_url`.
- Ingestion tetap JWT protected dan body cap tetap 16 KiB di Core API/BFF.
- Rate limit mengikuti middleware trusted-proxy-aware rate limit pada route ingestion; tidak melemahkan auth.
- Guard memastikan tidak ada public unauthenticated collector, tidak ada raw event export, dan tidak ada third-party analytics.

### Fase 10 - Operational readiness dan handoff

- Runbook operasional ada di `docs/internal-analytics-runbook.md`.
- Owner, permission `analytics.read`, `analytics.export`, `analytics.security_read`, export policy, retensi, smoke checklist, rollback, recovery, dan bukti 100% internal terdokumentasi.
- Boundary sesi ini: tidak deploy, tidak restart PM2, tidak menjalankan live migration, tidak menjalankan cleanup terhadap live DB, tidak membuka public unauthenticated collector, tidak menambahkan raw event export, dan tidak menambahkan third-party analytics.
