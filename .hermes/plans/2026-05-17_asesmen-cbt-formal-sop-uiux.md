# Asesmen CBT Formal SOP UI/UX Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menaikkan modul Asesmen CBT dari “mesin ujian digital yang kuat” menjadi workflow asesmen madrasah yang natural, formal, mudah dipahami panitia/guru/proktor/siswa, dan audit-safe.

**Architecture:** Perubahan dilakukan bertahap: mulai dari vocabulary/copy UI yang aman tanpa schema change, lalu Pusat Kegiatan Asesmen/SOP wizard, approval/finalisasi, arsip/berita acara, mode proktor sederhana, mobile CBT copy polish, dan contract hardening API. SvelteKit tetap BFF/proxy; semua data/mutasi domain tetap lewat Go core-api + sqlc + migration jika perlu.

**Tech Stack:** SvelteKit 2 + Svelte 5 + Tailwind/shadcn-svelte (`apps/web-admin`), Flutter (`apps/mobile`), Go/Chi/sqlc/PostgreSQL (`services/core-api`), PM2 production deploy.

**Source Review:** `docs/reviews/asesmen-cbt-swarm-review-2026-05-16.md`

---

## Non-Negotiable Rules

- Jangan deploy/restart PM2 production tanpa approval eksplisit.
- Jangan akses PostgreSQL langsung dari SvelteKit; SvelteKit hanya BFF/proxy.
- Jangan tampilkan credential, token raw, connection string, API key, atau password dalam log/laporan.
- Jangan broad rewrite modul CBT. Patch bertahap, kompatibel, dan mudah rollback.
- Semua migration harus additive/backward-compatible sampai data existing diaudit.
- Jangan merusak compatibility contract `/api/cbt/*`, `/api/bank-soal/*`, dan `/api/asesmen/*` tanpa keputusan eksplisit.
- Bank Soal tetap route/domain standalone; Asesmen CBT memakai Bank Soal sebagai input paket/kegiatan.

---

## Sprint 0 — Safety Gate & Baseline Audit

**Objective:** Mengunci baseline sebelum perubahan UI/SOP agar tidak mengganggu production CBT yang sudah dipakai.

**Files:**
- Create: `docs/contracts/asesmen-cbt-formal-sop.md`
- Create: `docs/contracts/asesmen-cbt-route-map.md`
- Create: `docs/reviews/asesmen-cbt-baseline-audit-YYYY-MM-DD.md`
- Read-only audit: `apps/web-admin/src/routes/asesmen/**`
- Read-only audit: `apps/web-admin/src/routes/bank-soal/**`
- Read-only audit: `apps/web-admin/src/routes/api/asesmen/**`
- Read-only audit: `apps/web-admin/src/lib/server/cbt-backend-proxy/**`
- Read-only audit: `services/core-api/cmd/api/main.go`
- Read-only audit: `apps/mobile/lib/src/**`

### Task 0.1: Create formal glossary contract

**Objective:** Menetapkan istilah resmi sebelum patch copy/UI.

**Steps:**
1. Create `docs/contracts/asesmen-cbt-formal-sop.md`.
2. Add glossary:
   - Modul: `Asesmen CBT`
   - Event: `Kegiatan Asesmen`
   - Package: `Paket Soal`
   - Session: `Sesi Ujian`
   - Room: `Ruang Ujian`
   - Proctoring: `Pengawasan Ruang`
   - Code/token: `Token Ujian`, `Token Ruang`
   - Review soal: `Verifikasi Soal`
   - Published: `Terbit`
   - Pool: `Kumpulan Soal`
   - Draw: `Ambil Acak`
   - Restore: `Pulihkan Sesi`
   - Fingerprint: `Penanda Perangkat`
3. Add forbidden UI terms unless in developer docs: `event`, `endpoint`, `payload`, `pool`, `draw`, `seed`, `stale`, `fingerprint`, `restore`, `builder`, `proctoring`.
4. Add allowed technical exceptions for route names, code identifiers, API docs.

**Verification:**
```bash
test -f docs/contracts/asesmen-cbt-formal-sop.md
```
Expected: exit 0.

**Commit:**
```bash
git add docs/contracts/asesmen-cbt-formal-sop.md
git commit -m "docs(asesmen): define formal CBT glossary"
```

### Task 0.2: Map backend/BFF/UI assessment routes

**Objective:** Mengetahui mismatch route sebelum menambah SOP UI.

**Steps:**
1. Generate backend route list from `services/core-api/cmd/api/main.go` using a small script or manual grep.
2. Generate BFF route list from `apps/web-admin/src/routes/api/asesmen/**/+server.ts`.
3. Generate UI usage list from `apps/web-admin/src/routes/asesmen/**` and `apps/web-admin/src/lib/server/cbt-backend-proxy/**`.
4. Save map to `docs/contracts/asesmen-cbt-route-map.md`.
5. Mark each route as one of:
   - `public-ui-used`
   - `bff-only-custom`
   - `backend-only-internal`
   - `legacy-compat`
   - `missing-alias-to-decide`

**Verification:**
```bash
test -f docs/contracts/asesmen-cbt-route-map.md
grep -n "GET /api/asesmen/sessions" docs/contracts/asesmen-cbt-route-map.md || true
grep -n "legacy-compat" docs/contracts/asesmen-cbt-route-map.md
```
Expected: document contains route map and compatibility notes.

**Commit:**
```bash
git add docs/contracts/asesmen-cbt-route-map.md
git commit -m "docs(asesmen): map CBT route contracts"
```

### Task 0.3: Current-state audit report

**Objective:** Menyimpan snapshot kondisi production/domain tanpa mutasi.

**Steps:**
1. Inspect counts/status distributions for events, sessions, packages, participants, proctoring incidents, audit logs.
2. Use backend/API or safe psql from server env only if needed; do not print credentials.
3. Save report to `docs/reviews/asesmen-cbt-baseline-audit-YYYY-MM-DD.md`.
4. Include explicit note: no production deploy/restart done.

**Verification:**
```bash
test -f docs/reviews/asesmen-cbt-baseline-audit-YYYY-MM-DD.md
```

**Commit:**
```bash
git add docs/reviews/asesmen-cbt-baseline-audit-YYYY-MM-DD.md
git commit -m "docs(asesmen): record CBT baseline audit"
```

---

## Sprint A — Bahasa & Formalisasi UI Asesmen CBT

**Objective:** Membuat UI terasa natural dan formal tanpa mengubah behavior besar.

**Scope:** Copy-only / low-risk UI polish.

**Files likely modified:**
- `apps/web-admin/src/routes/asesmen/+page.svelte`
- `apps/web-admin/src/routes/asesmen/persiapan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/hasil/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/**/*.svelte`
- `apps/web-admin/src/routes/asesmen/paket/**/*.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/**/*.svelte`
- `apps/web-admin/src/routes/bank-soal/**/*.svelte`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

### Task A.1: Replace user-facing “event” with “kegiatan”

**Objective:** Menghilangkan istilah developer/startup dari UI operator.

**Steps:**
1. Search UI strings only:
   ```bash
   grep -RIn "Event\|event" apps/web-admin/src/routes/asesmen apps/web-admin/src/routes/bank-soal apps/web-admin/src/lib/components/sidebar | head -100
   ```
2. Replace visible labels/copy:
   - `Event` → `Kegiatan`
   - `Paket Soal Event` → `Paket Soal Kegiatan`
   - `Kembali ke Event` → `Kembali ke Kegiatan`
   - `Buat/Cek Sesi Event` → `Buat/Cek Sesi Kegiatan`
3. Do not rename API fields such as `event_id` in payload or backend types.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add apps/web-admin/src
 git commit -m "style(asesmen): formalize kegiatan wording"
```

### Task A.2: Standardize Asesmen/Ujian/CBT page titles

**Objective:** Membuat navigasi utama konsisten.

**Target labels:**
- `/asesmen`: `Beranda Asesmen CBT`
- `/asesmen/persiapan`: `Persiapan Asesmen CBT`
- `/asesmen/pelaksanaan`: `Pelaksanaan Ujian CBT`
- `/asesmen/hasil`: `Hasil Asesmen CBT`
- `/asesmen/kegiatan`: `Kegiatan Asesmen`
- `/asesmen/sesi`: `Sesi Ujian`
- `/asesmen/pengawasan`: `Pengawasan Ruang`

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen apps/web-admin/src/lib/components/sidebar
 git commit -m "style(asesmen): standardize CBT page titles"
```

### Task A.3: Standardize token/kode, review/verifikasi, proctoring/pengawasan

**Objective:** Mengurangi kebingungan proktor, guru, dan siswa.

**Replacements:**
- `Kode Ujian` → `Token Ujian` unless first-use text says `Token Ujian/Kode Ujian`.
- `Review Soal` → `Verifikasi Soal` in user-facing Bank Soal labels.
- `Proctoring` → `Pengawasan Ruang` in UI Indonesian copy.
- `Panel Ruang` remains allowed as component/action label.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add apps/web-admin/src
 git commit -m "style(asesmen): align token and verification wording"
```

### Task A.4: Remove developer terms from operator/guru UI

**Objective:** Membuat copy tidak terasa teknis.

**Replace examples:**
- `endpoint` → `layanan sistem`
- `payload membawa event_id` → `paket otomatis ditautkan ke kegiatan ini`
- `pool soal` → `kumpulan soal`
- `draw` → `ambil acak`
- `seed` → `kode acak opsional`
- `locked` → `dikunci`
- `stale` → `kontak server terlalu lama`
- `fingerprint` → `penanda perangkat`
- `restore` → `pulihkan sesi`

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add apps/web-admin/src
 git commit -m "style(asesmen): remove technical UI wording"
```

---

## Sprint B — Pusat Kegiatan Asesmen + SOP Wizard

**Objective:** Menjadikan kegiatan sebagai pusat workflow panitia, bukan hanya detail event.

**Architecture:** Tambah frontend SOP dashboard terlebih dahulu dengan data existing + readiness existing. Jika perlu endpoint summary baru, buat additive query/service/handler.

**Files likely modified/created:**
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- Create: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/sop/+page.svelte` or integrate as tab in existing detail
- Create: `apps/web-admin/src/routes/api/asesmen/events/[id]/sop-readiness/+server.ts`
- Modify: `services/core-api/db/queries/cbt_events.sql`
- Modify: `services/core-api/internal/service/cbt_event.go`
- Modify: `services/core-api/internal/handler/cbt_event.go`
- Modify: `services/core-api/cmd/api/main.go`
- Tests: `services/core-api/internal/handler/*cbt_event*test.go`, service tests if available

### Task B.1: Define SOP stage model in docs and TypeScript

**Objective:** Menetapkan 10 tahap resmi tanpa schema change dulu.

**Stages:**
1. `draft` — Draft
2. `question_authoring` — Pengisian Soal
3. `question_verification` — Telaah/Verifikasi Soal
4. `package_ready` — Paket Siap
5. `participants_rooms_ready` — Peserta & Ruang Siap
6. `tokens_cards_ready` — Token & Kartu Siap
7. `execution` — Pelaksanaan
8. `grading` — Koreksi
9. `result_verification` — Verifikasi Hasil
10. `final_archive` — Final & Arsip

**Steps:**
1. Add SOP stage definitions to `docs/contracts/asesmen-cbt-formal-sop.md`.
2. Create small TS constant if needed:
   - `apps/web-admin/src/lib/asesmen/sop-stages.ts`
3. Use display labels from glossary.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

**Commit:**
```bash
git add docs/contracts/asesmen-cbt-formal-sop.md apps/web-admin/src/lib/asesmen/sop-stages.ts
 git commit -m "feat(asesmen): define SOP stage model"
```

### Task B.2: Add SOP timeline UI to kegiatan detail

**Objective:** Operator melihat posisi kegiatan dan next action.

**UI requirements:**
- Timeline 10 tahap.
- Each stage shows: status `Siap`, `Perlu Perhatian`, `Belum Siap`, `Sedang Berjalan`.
- Each stage shows 1–3 next actions linking to existing pages.
- No mutating action yet; read-only guidance.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```
Expected: PASS.

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen/kegiatan apps/web-admin/src/lib/asesmen
 git commit -m "feat(asesmen): add kegiatan SOP timeline"
```

### Task B.3: Add event SOP readiness endpoint if UI data is insufficient

**Objective:** Backend returns normalized readiness status per SOP stage.

**Endpoint:**
- `GET /api/asesmen/events/{id}/sop-readiness`

**Response shape:**
```json
{
  "event_id": "uuid",
  "stages": [
    {
      "key": "package_ready",
      "label": "Paket Siap",
      "status": "warning",
      "blocking_count": 0,
      "warning_count": 2,
      "next_actions": [
        { "label": "Lengkapi paket soal", "href": "/asesmen/kegiatan/..." }
      ]
    }
  ]
}
```

**Backend steps:**
1. Add SQL summary query in `services/core-api/db/queries/cbt_events.sql`.
2. Run sqlc:
   ```bash
   cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
   ```
3. Add service method.
4. Add handler method.
5. Register route in `cmd/api/main.go`.
6. Add BFF proxy route.

**Verification:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-asesmen-sop ./cmd/api
```
Expected: all PASS.

**Commit:**
```bash
git add services/core-api apps/web-admin/src/routes/api/asesmen/events
 git commit -m "feat(asesmen): expose kegiatan SOP readiness"
```

---

## Sprint C — Approval, Lock, dan Finalisasi Formal

**Objective:** Menambah pengesahan formal untuk paket, peserta/ruang/token, hasil, dan kegiatan final secara additive.

**Requires migration:** yes, additive only.

**Potential DB design:**
- Create `cbt_approval_records`
  - `id uuid`
  - `entity_type text` (`event`, `session`, `package`, `result`)
  - `entity_id uuid`
  - `approval_type text`
  - `status text` (`draft`, `approved`, `revoked`)
  - `approved_by uuid null`
  - `approved_at timestamptz null`
  - `notes text null`
  - `created_at`, `updated_at`
- Add indexes; do not enforce hard blocks yet.

### Task C.1: Add approval records migration

**Objective:** Menyimpan approval tanpa mengubah data existing.

**Files:**
- Create: `services/core-api/db/migrations/NNN_cbt_approval_records.sql`

**Verification:**
```bash
cd services/core-api
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/migrations/NNN_cbt_approval_records.sql --single-transaction
```
Use only in safe test/transaction first; do not expose DB URL.

**Commit:**
```bash
git add services/core-api/db/migrations/NNN_cbt_approval_records.sql
 git commit -m "feat(asesmen): add CBT approval records"
```

### Task C.2: Add approval API and audit domain events

**Objective:** Approve/revoke formal milestones with audit metadata.

**Endpoints:**
- `GET /api/asesmen/approvals?entity_type=&entity_id=`
- `POST /api/asesmen/approvals`
- `POST /api/asesmen/approvals/{id}/revoke`

**High-impact audit actions:**
- `CBT_PACKAGE_APPROVED`
- `CBT_PARTICIPANTS_ROOMS_APPROVED`
- `CBT_TOKENS_CARDS_APPROVED`
- `CBT_RESULTS_APPROVED`
- `CBT_EVENT_FINALIZED`

**Verification:**
```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-asesmen-approval ./cmd/api
npm --prefix ../../apps/web-admin run check
```

**Commit:**
```bash
git add services/core-api apps/web-admin/src/routes/api/asesmen/approvals
 git commit -m "feat(asesmen): add formal approval workflow"
```

### Task C.3: Integrate approval cards into SOP timeline

**Objective:** Operator/panitia bisa melihat dan melakukan pengesahan dari Pusat Kegiatan.

**UI requirements:**
- Approval cards for package, participants/rooms, tokens/cards, results, final archive.
- Show approver name, role, timestamp, notes.
- Typed confirmation for final/fatal actions.
- Never show token raw.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen/kegiatan apps/web-admin/src/routes/api/asesmen/approvals
 git commit -m "feat(asesmen): surface approval cards in SOP timeline"
```

---

## Sprint D — Berita Acara & Arsip Digital

**Objective:** Membuat output formal yang siap dipakai panitia/madrasah.

**Files likely modified/created:**
- `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`
- `apps/web-admin/src/routes/api/asesmen/events/[id]/archive/+server.ts`
- `apps/web-admin/src/routes/api/asesmen/sessions/[id]/minutes/+server.ts`
- `services/core-api/internal/service/cbt_session*.go`
- `services/core-api/internal/handler/cbt_session*.go`

### Task D.1: Expand minutes/BA data contract

**Objective:** BA memuat data SOP resmi.

**Add fields:**
- Hadir/tidak hadir/susulan.
- Peserta force submit/reset/unlock.
- Insiden ruang.
- Tindakan proktor.
- Handover status.
- Tanda tangan: proktor, operator, ketua panitia, kepala madrasah optional.

**Verification:**
```bash
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
```

**Commit:**
```bash
git add services/core-api apps/web-admin/src/routes/asesmen/sesi
 git commit -m "feat(asesmen): enrich CBT minutes contract"
```

### Task D.2: Add event archive page/action

**Objective:** Panitia bisa mengunduh paket arsip kegiatan.

**Initial scope:**
- UI checklist archive; export per document first.
- ZIP export may be phase 2 if backend not ready.

**Documents:**
- Kartu ujian.
- Daftar hadir.
- BA sesi/ruang.
- Rekap hasil.
- Rekap insiden.
- Audit ringkas.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen/kegiatan apps/web-admin/src/routes/api/asesmen
 git commit -m "feat(asesmen): add kegiatan archive checklist"
```

---

## Sprint E — Mode Proktor/Pengawas Sederhana

**Objective:** Membuat pengawasan lebih natural bagi pengawas non-IT.

**Files likely modified:**
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- related proctoring components if any

### Task E.1: Add simplified status buckets

**Objective:** Mengelompokkan peserta berdasarkan kasus operasional.

**Buckets:**
- Belum masuk
- Sedang mengerjakan
- Perlu bantuan
- Jaringan bermasalah
- Terkunci anti-cheat
- Sudah selesai

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen/sesi
 git commit -m "feat(asesmen): add simplified proctor status buckets"
```

### Task E.2: Add SOP action panel per bucket

**Objective:** Pengawas tahu tindakan berikutnya.

**Action mapping examples:**
- Jaringan bermasalah → sinkron ulang, catat insiden.
- Terkunci anti-cheat → unlock, catat insiden.
- Belum masuk → cek token/kartu, reset akses.
- Sudah selesai → verifikasi submit.

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

**Commit:**
```bash
git add apps/web-admin/src/routes/asesmen/sesi
 git commit -m "feat(asesmen): add proctor SOP action panel"
```

---

## Sprint F — Mobile CBT Student Copy Polish

**Objective:** Pesan di aplikasi siswa lebih sederhana, formal, dan tidak menampilkan error mentah.

**Files likely modified:**
- `apps/mobile/lib/src/exam_error_messages.dart`
- `apps/mobile/lib/src/screens/exam_login_screen.dart`
- `apps/mobile/lib/src/screens/exam_shell_screen.dart`
- `apps/mobile/lib/src/screens/exam_shell_connection.dart`
- `apps/mobile/lib/src/screens/exam_completed_screen.dart`
- `apps/mobile/lib/src/anti_cheat_guard.dart`

### Task F.1: Add 423 Locked error mapping

**Objective:** Status anti-cheat locked tampil ramah siswa.

**Copy:**
> Ujian dikunci sementara oleh sistem pengawasan. Tetap di tempat dan minta pengawas memeriksa akses Anda.

**Verification:**
```bash
cd apps/mobile
flutter test
```
If no tests available, run:
```bash
cd apps/mobile
flutter analyze
```

**Commit:**
```bash
git add apps/mobile/lib/src/exam_error_messages.dart
 git commit -m "fix(mobile): map locked exam errors clearly"
```

### Task F.2: Replace technical labels on student UI

**Objective:** Mengganti istilah teknis dengan bahasa siswa.

**Replace:**
- `Heartbeat aktif` → `Koneksi dipantau`
- `Anti-switch dasar` → `Keluar aplikasi tercatat`
- `Submit` → `Kirim jawaban`
- `Restore` → `Pulihkan sesi`
- `Sinkron` → `Terkirim ke server` / `Belum terkirim`

**Verification:**
```bash
cd apps/mobile
flutter analyze
```

**Commit:**
```bash
git add apps/mobile/lib/src
 git commit -m "style(mobile): simplify CBT student wording"
```

### Task F.3: Block raw backend errors from student-facing fallback

**Objective:** Tidak ada pesan Inggris/teknis mentah untuk siswa.

**Fallback copy:**
- Login: `Login belum berhasil. Periksa token dan minta bantuan pengawas.`
- Submit: `Jawaban belum bisa dikirim. Tetap di layar ini dan minta pengawas memeriksa status ujian.`

**Verification:**
```bash
cd apps/mobile
flutter analyze
```

**Commit:**
```bash
git add apps/mobile/lib/src/exam_error_messages.dart
 git commit -m "fix(mobile): avoid raw exam backend errors"
```

---

## Sprint G — Route/API Contract Hardening

**Objective:** Mengurangi risiko drift backend/BFF/UI untuk Asesmen CBT.

**Files likely modified/created:**
- Create: `apps/web-admin/src/routes/api/asesmen/sessions/[id]/+server.ts` if decided public.
- Create/update BFF aliases found missing.
- Create: `apps/web-admin/src/lib/server/assessment-route-contract.test.ts` or existing test location.
- Create: `services/core-api/internal/handler/assessment_route_contract_test.go` if useful.
- Docs: `docs/contracts/asesmen-cbt-route-map.md`

### Task G.1: Decide and document missing route aliases

**Objective:** Tidak semua mismatch otomatis bug; dokumentasikan keputusan.

**Decision categories:**
- Add BFF alias now.
- Backend-only internal.
- Legacy only.
- Deprecated and no UI usage.

**Verification:**
```bash
grep -n "missing-alias-to-decide" docs/contracts/asesmen-cbt-route-map.md && exit 1 || true
```
Expected: no undecided route remains.

**Commit:**
```bash
git add docs/contracts/asesmen-cbt-route-map.md
 git commit -m "docs(asesmen): resolve route alias decisions"
```

### Task G.2: Add BFF aliases for public contract gaps

**Objective:** Public UI contract tidak 404 di BFF.

**Candidate routes from review:**
- `GET /api/asesmen/sessions/{id}`
- `DELETE /api/asesmen/sessions/{id}` if used/allowed
- `GET /api/asesmen/sessions/{id}/participants/{pid}/answers` if public to admin/guru

**Verification:**
```bash
npm --prefix apps/web-admin run check
```

**Commit:**
```bash
git add apps/web-admin/src/routes/api/asesmen
 git commit -m "fix(asesmen): align BFF assessment route aliases"
```

### Task G.3: Add route contract tests

**Objective:** Future sprint tidak membuat drift baru.

**Steps:**
1. Add frontend route alias test if existing test infra supports it.
2. Add backend route registration smoke test if feasible.
3. At minimum add static script/test checking expected aliases exist.

**Verification:**
```bash
npm --prefix apps/web-admin run check
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
```

**Commit:**
```bash
git add apps/web-admin services/core-api docs/contracts/asesmen-cbt-route-map.md
 git commit -m "test(asesmen): add CBT route contract checks"
```

---

## Full Verification Matrix

Run after each sprint:

```bash
npm --prefix apps/web-admin run check
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-asesmen-formal-sop ./cmd/api
```

For mobile sprint:

```bash
cd apps/mobile
flutter analyze
flutter test
```

If `flutter test` has no suite or pre-existing failure, record clearly in sprint report.

---

## Deployment Plan — Only After Explicit Approval

**Do not run these until user approves deploy.**

1. Confirm clean git status except intentional files.
2. Backup production DB before migrations:
   ```bash
   # Use PM2/env safely; do not print DATABASE_URL.
   pg_dump --format=custom --file /home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql/pre-asesmen-formal-sop-YYYYMMDD-HHMMSS.dump "$DATABASE_URL"
   sha256sum /path/to/backup.dump
   ```
3. Apply additive migrations in order.
4. Build core-api:
   ```bash
   cd services/core-api
   cp -p bin/api bin/api.backup-asesmen-formal-sop-YYYYMMDD-HHMMSS
   go build -o bin/api ./cmd/api
   pm2 restart mtsn2kolut-core-api --update-env
   curl -fsS http://127.0.0.1:8080/health
   ```
5. Build/restart web-admin if frontend changed:
   ```bash
   npm --prefix apps/web-admin run build
   pm2 restart mtsn2kolut-web-admin --update-env
   ```
   Important: after any `npm --prefix apps/web-admin run build`, restart PM2 immediately.
6. Smoke test:
   - `/asesmen`
   - `/asesmen/kegiatan`
   - one kegiatan detail/SOP page
   - `/asesmen/paket`
   - `/asesmen/sesi`
   - `/asesmen/pengawasan`
   - `/asesmen/hasil`
7. `pm2 save` after successful health/smoke.

---

## Recommended Execution Order

1. Sprint 0 — safety/doc baseline.
2. Sprint A — language/formality polish. Low risk, immediate UX improvement.
3. Sprint G.1 — route decision docs, before building deeper SOP dependencies.
4. Sprint B — Pusat Kegiatan Asesmen/SOP wizard.
5. Sprint E — proctor simple mode.
6. Sprint F — mobile CBT copy polish.
7. Sprint C — approval/finalisasi with migration.
8. Sprint D — BA/arsip digital.
9. Sprint G.2–G.3 — route aliases/tests as contract hardening, or earlier if blockers appear.

## Acceptance Criteria

- UI no longer exposes “event” as primary user-facing term in Asesmen CBT.
- Main assessment pages use consistent labels: Asesmen CBT, Kegiatan Asesmen, Sesi Ujian, Pengawasan Ruang, Hasil Asesmen.
- Kegiatan detail has visible SOP timeline/checklist with next actions.
- Approval/finalisasi is additive and audit-safe.
- BA/arsip captures attendance, incidents, interventions, signatures, and result summary.
- Proctor view offers simple case-based statuses and actions.
- Mobile app maps locked/error states to Bahasa Indonesia student-friendly messages.
- Route contract doc has no unresolved public aliases.
- Verification commands pass or pre-existing failures are documented.
- No production deploy/restart is done without explicit approval.
