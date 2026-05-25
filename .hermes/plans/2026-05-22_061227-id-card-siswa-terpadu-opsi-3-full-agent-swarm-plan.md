# ID Card Siswa Terpadu Opsi 3 Full Function Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task with Agent Swarm, two-stage review, and parent verification gates.

**Goal:** Membangun fitur **ID Card Siswa Terpadu MTsN 2 Kolut** full fungsi: template resmi Alternatif A, lifecycle kartu, QR verifikasi aman, portal login QR+PIN, presensi kegiatan, perpustakaan, validasi CBT, audit scan, dan SOP operasional.

**Architecture:** Core API Go/PostgreSQL/sqlc menjadi sumber data dan kebijakan keamanan. Web-admin SvelteKit hanya BFF/proxy + UI/print. QR bukan password; QR hanya token kartu random yang di-hash server-side dan memulai verifikasi/login challenge.

**Tech Stack:** Go core-api, PostgreSQL, sqlc, SvelteKit web-admin/BFF, Tailwind, print CSS CR80/A4, RBAC existing, portal siswa existing.

---

## 0. Agent Swarm yang Sudah Digunakan untuk Menyusun Plan

Plan ini disusun dari 4 subagent paralel:

1. **Backend/Data/API Agent** — migration, sqlc, service, handler, endpoint, test target.
2. **Frontend/UI/Print Agent** — route SvelteKit, komponen kartu, PDF/print, portal QR, scanner petugas.
3. **Security/Audit/SOP Agent** — lifecycle, privacy, rate limit, role scope, SOP kartu hilang/rusak.
4. **Implementation Swarm Strategy Agent** — pembagian wave implementasi, ownership file, review gate.

Dokumen pendukung yang sudah dibuat oleh subagent:

- `docs/plans/2026-05-22-id-card-siswa-terpadu-opsi-3-frontend-ui-print-pdf.md`
- `docs/plans/id-card-siswa-terpadu-security-audit-sop-plan.md`
- `.hermes/plans/2026-05-22-id-card-siswa-terpadu-agent-swarm-strategy.md`

---

## 1. Prinsip Produk dan Keamanan

Nama fitur:

```text
ID Card Siswa Terpadu MTsN 2 Kolut
```

Bukan:

```text
Kartu CBT
```

Prinsip wajib:

- QR bukan password.
- QR bukan auto-login permanen.
- QR tidak berisi password, PIN, token CBT, NIK/KK, alamat, nomor orang tua, atau biodata lengkap.
- QR berisi URL/token random, misalnya `/s/idc/<token>`.
- Token QR disimpan sebagai hash/HMAC server-side.
- Raw token hanya muncul saat generate/print, jangan disimpan dan jangan dilog.
- Kartu hilang/rusak/reissue harus revoke token lama dan generate token baru.
- Portal login tetap wajib PIN/password siswa.
- CBT tetap wajib token ruang/proktor dan validasi jadwal/session/participant.
- Semua scan dan aksi lifecycle diaudit.

---

## 2. Fitur Full Fungsi Opsi 3

### 2.1 Template Resmi Alternatif A

Style yang dipilih user:

- hijau-emas resmi madrasah;
- logo + nama MTsN 2 Kolaka Utara;
- foto siswa;
- nama, NIS, kelas, angkatan;
- ID kartu;
- sisi belakang QR verifikasi besar + instruksi;
- peringatan “QR bukan password”.

### 2.2 Lifecycle Kartu

Status kartu:

```text
draft
active
lost
damaged
revoked
replaced
expired
suspended
```

Aksi:

- generate kartu;
- activate;
- mark printed;
- report lost;
- report damaged;
- suspend;
- reactivate;
- revoke;
- reissue/cetak ulang;
- view event history;
- view scan audit.

### 2.3 Portal QR+PIN

Alur:

```text
Scan QR kartu
→ sistem validasi kartu active + siswa active
→ buat login challenge singkat
→ tampil nama/kelas masked
→ siswa input PIN
→ challenge consumed
→ session portal siswa dibuat
```

### 2.4 Public Verify Terbatas

Tanpa login, halaman hanya boleh menampilkan:

- status kartu valid/tidak;
- nama siswa terbatas atau nama lengkap jika disetujui kebijakan;
- foto kecil jika diperlukan untuk verifikasi fisik;
- kelas;
- madrasah;
- masa berlaku.

Tidak boleh tampil:

- NIK/KK;
- alamat;
- nomor orang tua;
- riwayat layanan;
- NISN lengkap jika tidak perlu;
- token/password/PIN;
- informasi CBT.

### 2.5 Presensi Kegiatan

Alur aman:

```text
Guru/petugas login
→ pilih event/kegiatan terbuka
→ scan QR kartu
→ sistem tampilkan foto/nama/kelas
→ catat hadir idempotent
```

Catatan:

- Default jangan self-scan siswa, karena QR bisa difoto.
- Duplicate scan harus idempotent atau tampil “sudah hadir”.
- Event harus `open`.

### 2.6 Perpustakaan

Alur:

```text
Petugas perpustakaan scan kartu
→ tampil anggota/siswa + pinjaman aktif
→ petugas konfirmasi pinjam/kembali
```

QR hanya lookup anggota, transaksi tetap dikonfirmasi staf.

### 2.7 CBT Validation

Alur:

```text
Proktor/pengawas scan kartu
→ validasi kartu active
→ cek siswa adalah participant event/session/room
→ catat event student_card_validated
→ token ruang/proktor tetap wajib untuk CBT
```

Kartu tidak boleh membuka ujian sendiri.

---

## 3. Data Model / Migration Plan

### Task 1: Audit Schema Existing

**Objective:** Pastikan nama tabel/kolom existing untuk students, users, portal, attendance, library, CBT.

**Files:** read-only audit.

**Commands:**

```bash
cd services/core-api
set -a; . ./.env; set +a
PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\dt"
PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\d students"
PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "\d users"
```

**Acceptance:** daftar tabel/kolom final diketahui sebelum migration ditulis.

### Task 2: Migration Foundation Card Lifecycle

**Objective:** Tambah tabel lifecycle kartu dan template official Alternatif A.

**Files:**

- Create: `services/core-api/db/migrations/<next>_student_id_cards_foundation.sql`

**Conceptual SQL:**

```sql
CREATE TYPE student_card_status_enum AS ENUM (
  'draft', 'active', 'lost', 'damaged', 'revoked', 'replaced', 'expired', 'suspended'
);

CREATE TYPE student_card_event_type_enum AS ENUM (
  'generated', 'printed', 'activated', 'reported_lost', 'reported_damaged',
  'suspended', 'reactivated', 'revoked', 'reissued', 'replaced', 'expired',
  'pin_reset_requested', 'pin_reset_completed'
);

CREATE TYPE student_card_scan_module_enum AS ENUM (
  'public_verify', 'portal_login', 'event_attendance', 'library', 'cbt_validation', 'admin_services'
);
```

Create tables:

- `student_card_templates`
- `student_cards`
- `student_card_events`
- `student_card_scan_audit`

Key constraints:

```sql
CREATE UNIQUE INDEX idx_student_cards_one_active_per_student
ON student_cards(student_id)
WHERE status = 'active';
```

**Acceptance:** migration additive, no drop/rename, admin data unaffected.

### Task 3: Migration Portal Credential and Challenge

**Objective:** Tambah tabel PIN/challenge terpisah dari token QR.

**Files:**

- Create: `services/core-api/db/migrations/<next>_student_id_card_portal_login.sql`

Tables:

- `student_portal_credentials`
- `student_card_login_challenges`

Rules:

- PIN hash only.
- Challenge expires and consumed once.
- Failed PIN increments and can lock account.

### Task 4: Migration Presensi Kegiatan and Integration Hooks

**Objective:** Siapkan presensi kegiatan via kartu dan optional hooks library/CBT.

**Files:**

- Create: `services/core-api/db/migrations/<next>_student_card_service_integrations.sql`

Tables/changes:

- `student_activity_events`
- `student_activity_attendance`
- optional `library_loans.card_id`, `library_loans.card_scan_audit_id` if matching table exists.
- CBT preferably use participant event/audit, not bypass token.

### Task 5: Migration RBAC Permissions

**Objective:** Tambah permission granular dan grant admin.

Permissions:

```text
student_cards.read
student_cards.manage
student_cards.print
student_cards.audit
student_activity.read
student_activity.manage
student_activity.scan
library.card_scan
cbt.card_validate
student_portal.card_login
```

**Acceptance:** admin mendapat semua permission; role lain sesuai kebijakan awal.

---

## 4. Backend API Plan

### Task 6: sqlc Queries

**Files:**

- Create: `services/core-api/db/queries/student_cards.sql`
- Create: `services/core-api/db/queries/student_portal_credentials.sql`
- Create: `services/core-api/db/queries/student_activity_attendance.sql`

Queries required:

- list/get/create/update cards;
- get by token hash;
- active card by student;
- insert events/audit;
- create/get/consume login challenge;
- credential PIN update/failure/lockout;
- activity event and attendance scan.

Run:

```bash
cd services/core-api
sqlc generate
```

### Task 7: StudentCard Service

**Files:**

- Create: `services/core-api/internal/service/student_card.go`
- Create: `services/core-api/internal/service/student_card_test.go`

Responsibilities:

- generate secure random token;
- hash/HMAC token;
- generate card number;
- create/activate/print/lost/damaged/suspend/reactivate/reissue;
- public verify redacted;
- start/complete QR+PIN challenge;
- scan for attendance/library/CBT;
- insert audit for success/failure.

Tests:

- QR token not stored raw;
- only one active card per student;
- reissue invalidates old token;
- revoked/lost card cannot login/scan;
- wrong PIN lockout;
- expired challenge rejected;
- audit written for success/fail.

### Task 8: StudentCard Handler

**Files:**

- Create: `services/core-api/internal/handler/student_card.go`
- Create: `services/core-api/internal/handler/student_card_test.go`
- Modify: `services/core-api/cmd/api/main.go`

Endpoint groups:

```http
GET    /api/student-cards
GET    /api/student-cards/{id}
POST   /api/student-cards/generate
POST   /api/student-cards/{id}/activate
POST   /api/student-cards/{id}/mark-printed
POST   /api/student-cards/{id}/report-lost
POST   /api/student-cards/{id}/report-damaged
POST   /api/student-cards/{id}/suspend
POST   /api/student-cards/{id}/reactivate
POST   /api/student-cards/{id}/reissue
GET    /api/student-cards/{id}/events
GET    /api/student-cards/{id}/scan-audit
GET    /api/public/student-card/verify?t=...
POST   /api/portal/student/card-login/start
POST   /api/portal/student/card-login/complete
POST   /api/portal/student/pin/set
POST   /api/portal/student/pin/change
POST   /api/student-cards/{id}/reset-pin
POST   /api/student-activity/events
POST   /api/student-activity/events/{id}/open
POST   /api/student-activity/events/{id}/close
POST   /api/student-activity/events/{id}/attendance/scan-card
POST   /api/library/cards/scan
POST   /api/asesmen/sessions/{sessionID}/participants/validate-card
```

Handler rules:

- `http.MaxBytesReader`;
- reject unknown JSON fields;
- enforce RBAC in backend;
- redacted public responses;
- no raw QR token in logs/errors.

---

## 5. Frontend/UI/Print Plan

### Task 9: BFF Proxy Routes

**Files:** create under:

```text
apps/web-admin/src/routes/api/student-cards/**/+server.ts
apps/web-admin/src/routes/api/student-activity/**/+server.ts
apps/web-admin/src/routes/api/library/cards/scan/+server.ts
apps/web-admin/src/routes/api/asesmen/sessions/[sessionID]/participants/validate-card/+server.ts
```

Rules:

- BFF only forwards.
- Use existing backend proxy helper.
- Do not store/hash QR in frontend.

### Task 10: Admin Route — Kartu Siswa List

**Route:**

```text
/kesiswaan/kartu-siswa
```

**Files:**

- Create: `apps/web-admin/src/routes/kesiswaan/kartu-siswa/+page.server.ts`
- Create: `apps/web-admin/src/routes/kesiswaan/kartu-siswa/+page.svelte`
- Modify sidebar/navigation config.

UI:

- header summary cards;
- filter kelas/status/pencarian;
- table desktop + mobile cards;
- actions: generate, preview, cetak, reissue, report lost/damaged, reset PIN;
- status badge.

### Task 11: Student Card Components

**Files:**

- Create: `apps/web-admin/src/lib/components/student-card/StudentIdCardTemplate.svelte`
- Create: `apps/web-admin/src/lib/components/student-card/StudentCardPreview.svelte`
- Create: `apps/web-admin/src/lib/components/student-card/StudentCardStatusBadge.svelte`
- Create: `apps/web-admin/src/lib/components/student-card/StudentCardFilterBar.svelte`
- Create: `apps/web-admin/src/lib/components/student-card/QrScannerPanel.svelte`

Template modes:

```text
front
back
both
print
preview
```

Style: Alternatif A hijau-emas.

Fallback:

- no photo → silhouette/avatar;
- no active QR → warning;
- revoked/lost → watermark “Tidak Aktif”.

### Task 12: Print/PDF Page

**Route:**

```text
/kesiswaan/kartu-siswa/cetak?class_id=...
```

**Files:**

- Create: `apps/web-admin/src/routes/kesiswaan/kartu-siswa/cetak/+page.server.ts`
- Create: `apps/web-admin/src/routes/kesiswaan/kartu-siswa/cetak/+page.svelte`
- Create/modify print CSS if needed.

Requirements:

- CR80 size `85.60mm × 53.98mm`;
- A4 grid;
- print depan only, belakang only, or both;
- margin/cut guide;
- mark printed endpoint after operator confirms print;
- no sensitive data.

### Task 13: Detail Page

**Route:**

```text
/kesiswaan/kartu-siswa/[card_id]
```

Shows:

- card preview;
- lifecycle events;
- scan audit;
- linked student;
- status actions.

### Task 14: Public Verify Page

**Route:**

```text
/s/idc/[token]
```

States:

- valid active;
- expired;
- lost/revoked;
- invalid token;
- rate-limited.

Data minimal only.

### Task 15: Portal QR Login Page

**Route:**

```text
/portal/siswa/qr-login
```

Flow:

- scan QR or open from QR link;
- start challenge;
- input PIN;
- login session;
- redirect `/portal/siswa`.

### Task 16: Scanner Mode for Staff

**Route:**

```text
/kesiswaan/kartu-siswa/scanner
```

Modes:

- identifikasi;
- presensi kegiatan;
- perpustakaan;
- validasi CBT.

Use shared `QrScannerPanel` with manual token input fallback.

---

## 6. Security, Audit, SOP Plan

### SOP Kartu Hilang

1. Operator cari siswa/kartu.
2. Klik **Laporkan Hilang**.
3. Sistem set `lost`/`revoked`.
4. Token lama invalid.
5. Jika perlu cetak ulang, klik **Reissue**.
6. Sistem buat kartu baru + token baru.
7. Audit event tersimpan.

### SOP Kartu Rusak

- Jika QR masih aman tapi fisik rusak: `damaged`, lalu reissue.
- Jika QR sempat difoto/bocor: langsung `revoked`, lalu reissue.

### SOP Reset PIN

- Reset PIN tidak otomatis saat reissue kartu.
- Admin/kesiswaan wajib isi alasan.
- Siswa wajib set PIN baru saat login berikutnya.

### Audit wajib

- generate/activate/print;
- lost/damaged/revoke/reissue;
- PIN reset;
- public verify success/fail;
- portal login start/complete success/fail;
- attendance/library/CBT scan;
- role/scope denial.

### Abuse controls

- public verify rate limit;
- login start rate limit;
- PIN lockout;
- scan staff rate limit per actor;
- no raw token/PIN in logs.

---

## 7. Agent Swarm Implementation Strategy

### Rule umum

- Jangan implementasi tanpa approval user.
- Jangan deploy/restart tanpa approval eksplisit.
- Jangan `git add -A`.
- Subagent tidak commit kecuali diminta; parent integrasi dan commit.
- Parent wajib verifikasi ulang semua klaim subagent.
- Setiap wave ada implementer + spec reviewer + quality/security reviewer.

### Wave 0 — Repo Audit & Final Contract

Agents:

1. Backend schema audit.
2. Frontend route/component audit.
3. Security/RBAC audit.
4. Test/CI audit.

Output:

- final file ownership;
- migration number next;
- confirmed existing table names;
- risk list.

### Wave 1 — Backend Foundation

Agents:

1. Migration + sqlc queries.
2. StudentCard service + tests.
3. Handler/routes + tests.
4. RBAC permission migration.

Parent:

- run `sqlc generate`;
- run Go tests;
- inspect migration;
- no DB production migration until approval.

### Wave 2 — Admin UI + BFF

Agents:

1. BFF proxy routes.
2. Kartu Siswa list/detail UI.
3. Status/action dialogs.
4. Sidebar/nav/access.

Parent:

- run frontend build;
- browser smoke if allowed.

### Wave 3 — Template & Print

Agents:

1. `StudentIdCardTemplate.svelte` Alternatif A.
2. Print page CR80/A4.
3. Preview/export UX.
4. visual QA browser screenshot.

Parent:

- verify print layout and screenshot.

### Wave 4 — Public Verify + Portal QR+PIN

Agents:

1. Public verify UI.
2. Portal QR login UI.
3. PIN management UI.
4. security review for redaction/replay/rate limit.

Parent:

- test invalid/revoked/expired card;
- test no sensitive data exposure.

### Wave 5 — Integrations

Agents:

1. Presensi kegiatan scan.
2. Perpustakaan card scan.
3. CBT card validation.
4. audit/reporting UI.

Parent:

- verify CBT token ruang/proktor separation.
- verify no exam token leak.

### Wave 6 — Final Review & Rollout

Agents:

1. Spec compliance review.
2. Security/code review.
3. UI/UX/mobile/print QA.
4. migration/rollback review.

Parent:

- full Go tests/build;
- frontend clean build;
- git diff/stat;
- focused commit;
- deploy only if approved.

---

## 8. Verification Commands

### Backend

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
sqlc generate
go test ./internal/service/ -run 'StudentCard|Portal|Library|Cbt' -count=1 -timeout 120s
go test ./internal/handler/ -run 'StudentCard|Portal|Library|Cbt' -count=1 -timeout 120s
go test ./internal/handler/ ./internal/service/ -count=1 -timeout 180s
go vet ./...
go build -o bin/api ./cmd/api
```

### Frontend

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app/apps/web-admin
rm -rf build
npm run build
```

### Deployment smoke, only after approval

```bash
pm2 restart mtsn2kolut-core-api --update-env
sleep 2
pm2 restart mtsn2kolut-web-admin --update-env
sleep 3
curl -fsS http://127.0.0.1:8080/health
curl -s -o /dev/null -w "%{http_code}" http://localhost:8021/kesiswaan/kartu-siswa
```

Expected protected route may return `302` to login.

---

## 9. Rollback / Safety

Before production migration:

- create DB backup;
- verify backup checksum;
- apply additive migrations only;
- no drop/rename;
- feature can stay hidden behind sidebar/permission if incomplete;
- if error after deploy, remove sidebar permission and keep tables dormant;
- revert app build if needed;
- do not delete card/audit history.

---

## 10. Open Questions for User Approval

Sebelum implementasi, perlu keputusan:

1. Masa berlaku kartu:
   - selama siswa aktif, atau per tahun pelajaran?
2. Data depan kartu:
   - pakai NIS internal saja, atau tampilkan NISN juga?
3. Public verify:
   - tampil nama lengkap + kelas, atau nama masked demi privasi?
4. Presensi:
   - mulai dari presensi kegiatan dulu, atau langsung presensi harian?
5. Perpustakaan:
   - modul library existing mau langsung diintegrasikan atau tahap berikutnya?
6. CBT:
   - validasi kartu hanya untuk proktor, atau siswa juga scan mandiri sebelum masuk?
7. Cetak:
   - PDF A4 8 kartu/halaman atau 10 kartu/halaman?
8. Domain QR final:
   - `/s/idc/<token>` di domain web-admin sekarang, atau domain publik khusus?

---

## 11. Recommended Implementation Choice

Saya rekomendasikan implementasi tetap full Opsi 3, tapi dibuka bertahap:

1. **Release 1:** lifecycle kartu + template Alternatif A + cetak PDF + public verify.
2. **Release 2:** portal QR+PIN + reset PIN + audit login.
3. **Release 3:** presensi kegiatan + scanner petugas.
4. **Release 4:** perpustakaan + CBT validation.
5. **Release 5:** audit dashboard + SOP final + polish mobile/print.

Dengan cara ini kartu bisa segera dicetak, tetapi fondasi data dan keamanan tetap benar untuk fitur penuh.
