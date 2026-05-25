# Strategi Agent Swarm — Implementasi Aman ID Card Siswa Terpadu

> **Status:** plan only, belum implementasi.  
> **Scope repo:** `/home/servermtsn2kolut/mtsn2kolut-super-app`  
> **Prinsip utama:** ID Card Siswa Terpadu, bukan sekadar Kartu CBT. QR adalah identitas kartu/challenge starter, **bukan password**, bukan token ujian, dan bukan auto-login permanen.

## 1. Tujuan Produk

Membangun fitur **ID Card Siswa Terpadu MTsN 2 Kolut** yang mendukung secara bertahap:

- Identitas fisik siswa dengan template cetak CR80.
- QR verifikasi kartu dengan token acak yang bisa dicabut/diganti.
- Portal siswa: scan QR + PIN/password, bukan QR-only login.
- Presensi: scan oleh guru/piket/kesiswaan sesuai scope.
- Perpustakaan: scan anggota untuk kunjungan/pinjam/kembali oleh staf.
- CBT: scan kartu untuk verifikasi identitas peserta oleh proktor/pengawas, tetap wajib token ruang/sesi terpisah.
- Audit lifecycle kartu dan audit semua scan allowed/denied.

## 2. Guardrail Non-Negotiable

1. **No deploy/restart tanpa approval eksplisit.** Parent hanya boleh build/test. PM2 restart, migrasi produksi, dan deploy menunggu izin.
2. **Core API adalah pemilik data.** PostgreSQL hanya via `services/core-api`; SvelteKit hanya BFF/proxy.
3. **QR bukan credential.** QR tidak menyimpan password, PIN, token CBT, NIK/KK, alamat, nomor HP orang tua, atau raw NISN.
4. **Token QR random + hash server-side.** Token mentah hanya muncul saat cetak terkontrol; DB menyimpan HMAC/hash.
5. **Lifecycle eksplisit.** `active`, `lost`, `revoked`, `replaced`, `expired`, `suspended/damaged` harus memiliki audit.
6. **CBT tetap terpisah.** ID card tidak menggantikan token ruang/sesi CBT, jadwal, assignment, atau policy perangkat.
7. **Additive migration dulu.** Jangan rename/drop/enforce ketat sebelum audit data dan compatibility gate.
8. **Parent verifies.** Laporan subagent tidak cukup; parent wajib cek `git diff`, build, test, dan konflik file.
9. **Hindari konflik file.** Setiap agent hanya boleh menyentuh path miliknya; file pusat seperti `cmd/api/main.go` dan route registry diintegrasikan oleh parent/wave integrator.
10. **Jangan commit artifact.** Hindari `git add -A`; staging eksplisit saja jika nanti implementasi disetujui.

## 3. Kondisi Awal yang Sudah Ditemukan

- Referensi arsitektur sudah ada: `docs/architecture/id-card-siswa-security-data.md`.
- Mockup dummy sudah ada di `reports/id-card-mockup/`.
- Belum ditemukan implementasi `student_cards` / `student_card_*` di Go/SQL berdasarkan scan cepat.
- Modul terkait yang sudah ada:
  - Siswa/portal: `services/core-api/internal/service/student_portal.go`, `internal/handler/student_portal.go`, `db/queries/students.sql`.
  - Presensi: `services/core-api/internal/service/attendance.go`, `internal/handler/attendance.go`, `db/queries/attendance.sql`.
  - Perpustakaan: `services/core-api/internal/service/library.go`, `internal/handler/library.go`, `db/queries/library.sql`.
  - CBT: `services/core-api/internal/service/cbt_session.go`, `internal/handler/cbt_session.go`, `db/queries/cbt_sessions.sql`.

## 4. Model Agent Swarm

Gunakan **swarm per wave**, bukan semua agent mengedit sekaligus. Default 5 agent paralel untuk discovery/review; untuk implementasi gunakan worktree atau file ownership ketat.

### Parent Orchestrator

Tanggung jawab parent:

- Membuat branch/worktree implementasi.
- Menjalankan pre-flight `git status --short`.
- Membagi scope agent dan daftar file yang boleh diedit.
- Menggabungkan hasil per wave.
- Menjalankan verifikasi final:
  - `git diff --stat`
  - `cd services/core-api && make db-sqlc` jika query berubah
  - `cd services/core-api && go test ./internal/handler/ ./internal/service/ -count=1 -timeout 120s -short`
  - `cd services/core-api && go vet ./...`
  - `cd services/core-api && go build -o bin/api ./cmd/api`
  - `cd apps/web-admin && rm -rf build && npm run build` bila frontend berubah
- Tidak deploy/restart tanpa approval.

## 5. Pembagian Wave dan Agent

### Wave 0 — Read-only Audit & Contract Freeze

**Tujuan:** memahami existing schema/API/UI, mengunci kontrak, dan mencegah salah desain sebelum coding.

- **A0.1 Repo Auditor**
  - Read-only.
  - Audit struktur siswa, portal, presensi, library, CBT, RBAC, route registry.
  - Output: peta file dan titik integrasi.

- **A0.2 Security Architect**
  - Read-only.
  - Review `docs/architecture/id-card-siswa-security-data.md` dan baseline security.
  - Output: invariants, threat model QR/PIN/scan, rate limit, audit requirements.

- **A0.3 Data/Migration Planner**
  - Read-only.
  - Tentukan additive schema: `student_cards`, `student_card_events`, optional `student_card_scan_sessions`, `student_portal_credentials` reuse/extension.
  - Output: DDL draft + rollback-safe notes.

- **A0.4 UI/UX + Print Planner**
  - Read-only.
  - Rancang route admin/kesiswaan, bulk issue, print CR80, scan screens.
  - Output: UI sitemap + print template constraints.

- **A0.5 Integration Planner**
  - Read-only.
  - Map presensi, library, CBT hooks tanpa mengubah policy existing.
  - Output: integration contract and test matrix.

**Parent gate:** approve contract sebelum Wave 1. Jika ada konflik desain, stop dan minta keputusan user.

### Wave 1 — Backend Foundation: Migrations, SQLC, Service, API

**Tujuan:** fondasi kartu dan lifecycle aman tanpa menyentuh UI besar.

- **A1 Backend Migration/API Agent**
  - Boleh edit:
    - `services/core-api/db/migrations/*_student_cards*.sql`
    - `services/core-api/db/queries/student_cards.sql`
    - `services/core-api/internal/service/student_card*.go`
    - `services/core-api/internal/handler/student_card*.go`
    - test terkait `student_card*`
  - Jangan edit `cmd/api/main.go` kecuali parent menunjuk sebagai integrator.
  - Deliverable:
    - issue/activate/suspend/revoke/reissue/print-token lifecycle
    - QR token generation + HMAC/hash lookup
    - audit event append-only
    - explicit errors and RBAC permission checks

- **A1b RBAC/Route Integrator** *(bisa parent atau agent tunggal setelah A1 selesai)*
  - Boleh edit:
    - migration RBAC permission grant admin
    - route registration di `services/core-api/cmd/api/main.go`
    - auth/permission mapping jika ada pattern existing
  - Tidak paralel dengan agent lain yang menyentuh route registry.

**Parent gate:** `make db-sqlc`, Go unit tests, vet/build. Tidak apply migration produksi.

### Wave 2 — Frontend Admin/Kesiswaan + BFF

**Tujuan:** UI pengelolaan kartu, lifecycle, audit, dan cetak massal.

- **A2 BFF Proxy Agent**
  - Boleh edit:
    - `apps/web-admin/src/routes/api/student-cards/**`
    - proxy helper terkait jika ada pattern shared
    - BFF tests untuk envelope/error
  - Tidak menaruh business logic atau DB access di SvelteKit.

- **A3 Frontend UI Agent**
  - Boleh edit:
    - route admin/kesiswaan ID card, misalnya `apps/web-admin/src/routes/kesiswaan/id-cards/**` atau route yang disepakati Wave 0
    - component lokal ID card
    - akses menu/sidebar hanya setelah backend route valid
  - Deliverable:
    - daftar kartu/status
    - issue/reissue/revoke/suspend/reset PIN action dengan confirm dialog
    - audit view
    - bulk select siswa untuk generate/print

**Parent gate:** `npm run build`, route access smoke, no direct DB import.

### Wave 3 — Print Template / CR80 Renderer

**Tujuan:** template cetak resmi, aman, dan tidak membocorkan data sensitif.

- **A4 Print Template Agent**
  - Boleh edit:
    - `apps/web-admin/src/lib/id-card/**`
    - route print khusus, misalnya `apps/web-admin/src/routes/kesiswaan/id-cards/print/**`
    - test/render helper bila ada
  - Output:
    - layout CR80 85.60 × 53.98 mm depan-belakang
    - QR besar di belakang
    - card ID, nama, kelas, foto, masa berlaku
    - warning “QR bukan password”
    - mode dummy preview dan mode real via API
  - Larangan: cetak password/PIN/token CBT/NIK/KK/alamat/phone orang tua/raw token text.

**Parent gate:** browser/print visual QA bila memungkinkan, `npm run build`.

### Wave 4 — Portal QR + PIN Login

**Tujuan:** flow login aman: scan QR sebagai shortcut identitas + PIN/password.

- **A5 Portal QR/Login Agent**
  - Boleh edit:
    - `services/core-api/internal/service/student_portal*.go`
    - `services/core-api/internal/handler/student_portal*.go`
    - endpoint `card-login/start` dan `card-login/verify-pin`
    - `apps/web-admin/src/lib/client/student-portal.ts` atau route portal terkait
  - Deliverable:
    - short-lived challenge/session 2-5 menit
    - PIN verify + rate limit/lockout
    - audit success/fail/denied
  - Larangan: QR-only session; menampilkan biodata sensitif sebelum PIN valid.

**Parent gate:** tests untuk expired challenge, wrong PIN, revoked card, inactive student, rate limit.

### Wave 5 — Integrasi Presensi, Perpustakaan, CBT

Pecah menjadi 3 agent agar tidak konflik.

- **A6 Attendance Integration Agent**
  - Boleh edit:
    - `services/core-api/internal/service/attendance*.go`
    - `services/core-api/internal/handler/attendance*.go`
    - `services/core-api/db/queries/attendance.sql`
    - UI scanner presensi bila disepakati
  - Deliverable: idempotent scan per window, scope guru/piket, audit `attendance_used`.

- **A7 Library Integration Agent**
  - Boleh edit:
    - `services/core-api/internal/service/library*.go`
    - `services/core-api/internal/handler/library*.go`
    - `services/core-api/db/queries/library.sql`
    - UI scanner perpustakaan
  - Deliverable: resolve member by card, visit/loan flow with staff confirmation, audit `library_used`.

- **A8 CBT Identity Agent**
  - Boleh edit:
    - `services/core-api/internal/service/cbt_session*.go`
    - `services/core-api/internal/handler/cbt_session*.go`
    - `services/core-api/db/queries/cbt_sessions.sql`
    - proctor UI scanner bila disepakati
  - Deliverable: proctor identity scan for event/session/room participant verification.
  - Larangan: mengganti token ruang/sesi; jangan melemahkan existing CBT token policy.

**Parent gate:** focused Go tests per module + regression CBT session tests.

### Wave 6 — Tests, Security Review, QA, Fixes

- **A9 Test Agent**
  - Tambah unit/integration tests untuk lifecycle, portal QR+PIN, revoked/lost card, scan denied, idempotency.

- **A10 Security Reviewer**
  - Read-only review: token leakage, QR bearer risk, rate limit, role-scope bypass, audit gaps, sensitive data in print/UI.

- **A11 Code Reviewer**
  - Review maintainability, sqlc correctness, error mapping, route guards, Svelte state, build risk.

- **A12 QA Agent**
  - Jalankan test/build yang relevan dan laporkan log penting.

**Parent gate:** semua critical/high harus ditutup sebelum user diminta approval deploy.

## 6. Strategi Menghindari Konflik File

1. **Wave serial untuk file pusat:** `cmd/api/main.go`, route registry, sidebar/navigation, shared access helpers hanya disentuh oleh satu integrator setelah implementation agents selesai.
2. **File ownership per domain:** migration/query/service/handler dibagi per domain (`student_cards`, `attendance`, `library`, `cbt_session`).
3. **Gunakan git worktree untuk implementasi paralel berat:** satu worktree per agent, lalu parent cherry-pick terarah.
4. **Kontrak API ditulis dulu:** endpoint, request/response, error semantics, permission; UI agent tidak menebak shape backend.
5. **Tidak ada agent yang menjalankan formatter global atau `git add -A`.**
6. **Generated files:** hanya agent backend foundation atau parent yang menjalankan `make db-sqlc`; jangan beberapa agent generate bersamaan.

## 7. Acceptance Criteria per Modul

### Kartu/Lifecycle

- Admin/kesiswaan bisa issue, print, activate, suspend, revoke, reissue.
- Satu kartu active per siswa.
- Reissue revoke/replace kartu lama dan token lama langsung ditolak.
- Audit semua perubahan lifecycle.

### QR/Token

- Token random minimal 128 bit, lebih baik 192/256 bit.
- DB menyimpan hash/HMAC, bukan raw token.
- Scan revoked/lost/expired/suspended ditolak dan diaudit.
- Tidak ada sensitive data di QR atau print.

### Portal

- Login manual tetap ada.
- QR login membutuhkan PIN/password.
- Challenge punya TTL dan one-time/used semantics.
- Wrong PIN/rate limit/lockout tercakup test.

### Presensi

- Pemindai wajib login dan punya permission/scope.
- Scan idempotent dalam window yang sama.
- Foto siswa/status tampil untuk verifikasi visual petugas.

### Perpustakaan

- Staf/admin resolve siswa by card.
- Pinjam/kembali tetap konfirmasi petugas.
- Kunjungan idempotent per hari/window.

### CBT

- Proktor/pengawas scan kartu hanya untuk verifikasi identitas peserta.
- Token ruang/sesi CBT tetap wajib.
- Audit menyimpan event/session/room/proctor/result.

## 8. Verification Matrix Parent

Sebelum final implementasi dinyatakan selesai:

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git diff --stat

cd services/core-api
make db-sqlc
go test ./internal/handler/ ./internal/service/ -count=1 -timeout 120s -short
go vet ./...
go build -o bin/api ./cmd/api

cd ../../apps/web-admin
rm -rf build && npm run build
```

Tambahan smoke non-deploy setelah server lokal tersedia:

- Resolve revoked card => 403/denied + audit.
- Resolve active card for portal => challenge created, no session until PIN.
- Verify PIN wrong => fail + rate limit counter.
- Verify PIN correct => session created.
- Attendance scan double submit => idempotent, no duplicate.
- Library scan => member resolved but transaction needs confirmation.
- CBT identity scan => participant matched, but exam still requires room token.

## 9. Approval Gates

- **Gate A:** Setelah Wave 0, minta approval kontrak data/API/UI jika ada keputusan besar.
- **Gate B:** Setelah Wave 1 backend foundation lulus test, minta approval lanjut UI/integrasi bila scope melebar.
- **Gate C:** Setelah semua build/test pass, minta approval deploy/migration/restart.
- **Gate D:** Setelah deploy disetujui, deploy order: migration verified → backend restart → health check → frontend build/restart → smoke test.

## 10. Template Prompt Subagent Implementasi

Setiap subagent harus menerima format ini:

```text
TASK: <nama wave/agent>
REPO: /home/servermtsn2kolut/mtsn2kolut-super-app
BAHASA LAPORAN: Indonesia
MODE: implementasi terbatas / read-only sesuai wave

BOLEH EDIT:
- <path spesifik>

DILARANG EDIT:
- file milik agent lain
- deploy/restart/pm2
- migration produksi
- secrets/.env
- generated files kecuali ditugaskan

INVARIANTS:
- QR bukan password
- token QR random, hash server-side
- CBT token ruang/sesi tetap terpisah
- Core API owns DB; SvelteKit BFF only
- audit allowed/denied wajib

OUTPUT:
- ringkasan perubahan
- file berubah
- test/build yang dijalankan
- issue/risiko tersisa
```

## 11. Rekomendasi Urutan Eksekusi Nanti

1. Jalankan Wave 0 read-only swarm dulu.
2. Parent finalisasi kontrak dan file ownership.
3. Jalankan Wave 1 backend foundation di worktree/branch terpisah.
4. Parent verifikasi backend.
5. Jalankan Wave 2 + Wave 3 frontend/print setelah kontrak API stabil.
6. Jalankan Wave 4 portal login.
7. Jalankan Wave 5 integrasi presensi/library/CBT satu per satu atau worktree paralel dengan parent integrator.
8. Jalankan Wave 6 review/QA.
9. Minta approval deploy/restart bila semua gate hijau.
