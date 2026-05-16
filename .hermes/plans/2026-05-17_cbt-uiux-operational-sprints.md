# CBT UI/UX Operational Sprints Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task. Do not deploy/restart PM2 unless the user explicitly approves.

**Goal:** Menyempurnakan UI/UX Asesmen CBT agar lebih sesuai praktik CBT sekolah/madrasah: cepat dipakai saat hari-H, mudah dipahami guru/pengawas/siswa, audit-safe, dan tetap backward-compatible.

**Architecture:** Perubahan dilakukan bertahap dan additive. SvelteKit tetap sebagai BFF/proxy ke Go core-api; tidak ada akses DB langsung dari frontend. Backend changes hanya melalui Go service/handler/sqlc/migration bila dibutuhkan. Prioritas awal adalah UX/copy/layout yang aman tanpa migration, lalu fitur arsip/pengesahan yang memanfaatkan API approval existing.

**Tech Stack:** SvelteKit web-admin, Svelte 5, Go core-api, PostgreSQL/sqlc, Flutter mobile CBT, PM2 production runtime.

---

## Source Review

Plan ini berasal dari swarm review UI/UX CBT:

- Report: `docs/reviews/asesmen-cbt-uiux-swarm-review-2026-05-17.md`
- Commit basis: `92775fd feat(asesmen): formalize CBT SOP workflow`
- Scope utama:
  - `apps/web-admin/src/routes/asesmen/**`
  - `apps/web-admin/src/routes/bank-soal/**`
  - `apps/mobile/lib/src/**`
  - `services/core-api/**` bila butuh endpoint/readiness/approval tambahan

## Non-Negotiable Constraints

- Jangan deploy/restart PM2 tanpa approval eksplisit.
- Jangan membaca/menampilkan secrets/credential/connection string.
- Jangan akses DB dari SvelteKit; SvelteKit hanya BFF/proxy.
- Jangan broad rewrite; patch kecil, backward-compatible.
- Jangan menghapus/rename data production atau route lama.
- Token ujian/token ruang adalah data sensitif; jangan tampilkan di dokumen arsip final secara default.
- Setiap sprint harus lulus validasi minimal:

```bash
npm --prefix apps/web-admin run check
```

Jika backend disentuh:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-uiux-sprintN ./cmd/api
```

Jika mobile disentuh dan Flutter tersedia:

```bash
cd apps/mobile
flutter analyze
flutter test
```

Jika Flutter tidak tersedia, catat eksplisit: `flutter not installed`.

---

## Sprint UX CBT 0 — Safe Baseline & Regression Guard

**Goal:** Mengunci baseline UI/UX, route, terminology, dan risiko sebelum perubahan.

**Outcome:** Implementer punya peta perubahan, tidak merusak route/API existing, dan tidak menyentuh untracked file yang tidak terkait.

### Task 0.1: Confirm branch and worktree

**Files:** none

**Steps:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git branch --show-current
git status --short
git log --oneline -5
```

**Expected:** branch kerja jelas; catat untracked existing. Jangan commit file tidak terkait seperti infographic/review lama kecuali diminta.

### Task 0.2: Re-read review and plan inputs

**Files:**
- Read: `docs/reviews/asesmen-cbt-uiux-swarm-review-2026-05-17.md`
- Read: `docs/contracts/asesmen-cbt-formal-sop.md`
- Read: `docs/contracts/asesmen-cbt-route-map.md`

**Steps:**

```bash
sed -n '1,220p' docs/reviews/asesmen-cbt-uiux-swarm-review-2026-05-17.md
sed -n '1,220p' docs/contracts/asesmen-cbt-formal-sop.md
sed -n '1,220p' docs/contracts/asesmen-cbt-route-map.md
```

**Expected:** implementer memahami prioritas P0/P1.

### Task 0.3: Add/update UI terminology checklist

**Objective:** Membuat checklist istilah yang dipakai lintas sprint.

**Files:**
- Modify: `docs/contracts/asesmen-cbt-formal-sop.md`

**Content to add/update:**

```md
## UI Terminology Cleanup Targets

- Event -> Kegiatan Asesmen
- Kode Ruang -> Token Ruang
- Kode Ujian -> Token Ujian
- Kirim -> Sudah kirim / Jawaban terkirim
- Pindah -> Keluar aplikasi
- Tangkapan -> Coba tangkap layar
- Published -> Terbit
- Export -> Ekspor
- Advanced Bank Soal -> Pengelolaan Lanjutan Bank Soal
- Analisis Butir -> Pemantauan Mutu Soal, unless true item-analysis metrics exist
- Heartbeat -> Status/koneksi
- Anti-switch -> Tetap di aplikasi
- Fingerprint -> Penanda perangkat
```

**Verification:** Docs-only, no build needed unless combined with code.

---

## Sprint UX CBT 1 — Simple Proctor Mode + Terminology Cleanup

**Goal:** Membuat layar pengawas lebih cepat dipakai saat hari-H dan membersihkan istilah CBT utama.

**Primary Users:** proktor/pengawas ruang, operator CBT hari-H.

**Commit message:** `feat(asesmen): add simple proctor mode`

### Task 1.1: Inspect live proctoring pages

**Files:**
- Read: `apps/web-admin/src/routes/asesmen/pengawasan/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`

**Objective:** Identify existing state variables, participant statuses, action handlers, and labels.

**Verification:** Notes added to plan implementation notes or commit body.

### Task 1.2: Add display mode state to room proctoring page

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Implementation:** Add local state:

```ts
type ProctorViewMode = 'simple' | 'complete';
let proctorViewMode: ProctorViewMode = 'simple';
```

Add toggle near header:

```svelte
<div class="inline-flex rounded-full border border-slate-200 bg-white p-1 text-xs font-semibold">
  <button
    type="button"
    class:...={proctorViewMode === 'simple'}
    on:click={() => (proctorViewMode = 'simple')}
  >
    Mode Sederhana
  </button>
  <button
    type="button"
    class:...={proctorViewMode === 'complete'}
    on:click={() => (proctorViewMode = 'complete')}
  >
    Mode Lengkap
  </button>
</div>
```

**Acceptance Criteria:**
- Default `Mode Sederhana`.
- `Mode Lengkap` tetap menampilkan semua fitur existing.
- Tidak menghapus fitur existing.

### Task 1.3: Create prioritized “Butuh Tindakan” participant list

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Implementation idea:** derive participants:

```ts
function needsAttention(participant: RoomParticipant): boolean {
  return Boolean(
    participant.is_locked ||
    participant.risk_level === 'high' ||
    participant.connection_status === 'offline' ||
    participant.connection_status === 'stale' ||
    participant.needs_attention ||
    participant.submit_status !== 'submitted'
  );
}

$: attentionParticipants = participants.filter(needsAttention);
```

Use actual field names from file; do not invent if types differ.

**UI:** In simple mode show cards:
- Nama/NIS/no meja.
- Status utama: Terkunci, Terputus, Perlu perhatian, Belum kirim.
- Last seen/time if available.
- Actions: Periksa, Peringatkan, Instruksi masuk ulang, Buka kunci if locked, Atur ulang akses.

**Acceptance Criteria:**
- Pengawas can see urgent participants without horizontal table scroll.
- Empty state: `Tidak ada peserta yang butuh tindakan saat ini.`

### Task 1.4: Add quick filters for participant list

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Filters:**
- Butuh Tindakan
- Terkunci
- Terputus
- Belum Kirim
- Sudah Kirim
- Semua

**Acceptance Criteria:**
- Default filter: Butuh Tindakan during active session.
- Filter buttons have `aria-pressed` and `aria-label`.

### Task 1.5: Make unlock action visible in room panel

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Acceptance Criteria:**
- If participant is locked, `Buka Kunci` button appears as primary action.
- Copy explains: `Buka akses hanya setelah pengawas memeriksa perangkat/siswa.`
- Uses existing backend action; do not create new API unless current action missing.

### Task 1.6: Replace high-risk `window.prompt` with structured modal

**Files:**
- Modify: room proctoring page and/or shared component if already available.

**Preferred component:** Use existing dialog/modal component in app if present. If no component exists, create small local modal markup.

**Fields:**
- Reason select:
  - Perangkat sudah diverifikasi
  - Gangguan jaringan
  - Salah keluar aplikasi
  - Ganti perangkat disetujui
  - Lainnya
- Optional note.
- Confirm button.

**Acceptance Criteria:**
- Unlock/reset/incident actions no longer rely on raw prompt where feasible.
- Existing behavior preserved if modal submission fails.

### Task 1.7: Terminology cleanup for proctoring and phase pages

**Files likely:**
- Modify: `apps/web-admin/src/routes/asesmen/pengawasan/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`

**Replace visible copy:**
- `Kode Ruang` -> `Token Ruang`
- `Kirim` -> `Sudah kirim` or `Jawaban terkirim`
- `Pindah` -> `Keluar aplikasi`
- `Tangkapan` -> `Coba tangkap layar`
- `Atur Ulang` -> `Atur ulang akses`
- `Masuk Ulang` -> `Instruksi masuk ulang`
- `Ruang dikunci` -> clarify as `Data ruang dikunci operator` if it means room metadata locked.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

### Task 1.8: Sprint 1 tests and commit

**Commands:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/cbt-backend-paths.test.ts src/lib/server/assessment-route-contract.test.ts
```

**Commit:**

```bash
git add apps/web-admin/src/routes/asesmen docs/contracts/asesmen-cbt-formal-sop.md
git commit -m "feat(asesmen): add simple proctor mode"
```

---

## Sprint UX CBT 2 — Archive Kegiatan + Visible SOP Approval

**Goal:** Membuat finalisasi, pengesahan, dan arsip kegiatan terlihat nyata di UI.

**Primary Users:** operator CBT, kepala/panitia, admin madrasah.

**Commit message:** `feat(asesmen): add activity archive and approval panel`

### Task 2.1: Inspect approval endpoints and BFF

**Files:**
- Read: `services/core-api/internal/handler/cbt_approval.go`
- Read: `services/core-api/internal/service/cbt_approval.go`
- Read: `apps/web-admin/src/routes/api/asesmen/approvals/+server.ts`
- Read: `apps/web-admin/src/routes/api/asesmen/approvals/[id]/revoke/+server.ts`
- Read: `apps/web-admin/src/lib/server/cbt-backend-proxy/approvals/+server.ts`

**Acceptance Criteria:** know request/response body for list/create/revoke.

### Task 2.2: Add frontend approval client helper

**Files:**
- Create: `apps/web-admin/src/lib/asesmen/approval-client.ts` or local functions in detail page if project avoids helper.

**Functions:**
- `listApprovals(entityType, entityId)`
- `createApproval(payload)`
- `revokeApproval(id, reason)`

**Acceptance Criteria:** typed and uses `/api/asesmen/approvals` BFF only.

### Task 2.3: Add “Pengesahan SOP” panel to detail kegiatan

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`

**Panel milestones:**
- Sahkan Paket Soal (`package_ready`)
- Sahkan Peserta & Ruang (`participants_rooms_ready`)
- Sahkan Token & Kartu (`tokens_cards_ready`)
- Sahkan Hasil (`result_verification`)
- Finalkan & Arsipkan (`final_archive`)

**UI per milestone:**
- Status: Belum disahkan / Disahkan / Dicabut.
- Actor and timestamp if approved.
- Note.
- Primary action if allowed: `Sahkan`.
- Secondary action if approved and allowed: `Cabut pengesahan`.

**Acceptance Criteria:**
- No hard lock of old workflow yet.
- Clear copy: `Pengesahan ini mencatat audit formal, belum memblokir alur lama.`

### Task 2.4: Add archive route

**Files:**
- Create: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/archive/+page.svelte`

**Content:**
- Header: `Arsip Kegiatan Asesmen`
- Checklist documents:
  - Kartu ujian/token distribusi
  - Daftar hadir
  - Berita acara sesi/ruang
  - Rekap hasil
  - Rekap insiden
  - Audit pengesahan
  - Catatan final arsip
- Action links:
  - Back to detail kegiatan
  - Cetak kartu ujian
  - Buka sesi/BA
  - Buka hasil kegiatan
  - Buka pengawasan/report if route exists

**Acceptance Criteria:**
- Existing link `/asesmen/kegiatan/{id}/archive` no longer broken.
- If data unavailable, show actionable placeholder, not blank.

### Task 2.5: Mask token in archive context

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- Possibly create helper: `apps/web-admin/src/lib/asesmen/token-display.ts`

**Behavior:**
- Default BA/arsip mode masks token:
  - `ABCD1234` -> `AB••••34`
- Add warning and explicit option for pre-exam distribution if raw token display is necessary.

**Copy:**

```text
Dokumen arsip menyamarkan token ujian. Tampilkan token lengkap hanya untuk distribusi sebelum ujian dan simpan terbatas.
```

**Acceptance Criteria:**
- Token raw is not default in archive/final print context.
- Kartu ujian distribution can still show token where required.

### Task 2.6: Sprint 2 verification and commit

**Commands:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/cbt-backend-paths.test.ts src/lib/server/assessment-route-contract.test.ts
```

If backend changes are added:

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-uiux-sprint2 ./cmd/api
```

**Commit:**

```bash
git add apps/web-admin/src/routes/asesmen apps/web-admin/src/lib/asesmen services/core-api docs/contracts
git commit -m "feat(asesmen): add activity archive and approval panel"
```

---

## Sprint UX CBT 3 — Bank Soal Terminology + Readiness Refinement

**Goal:** Menyesuaikan UI Bank Soal dengan praktik guru/reviewer CBT dan menghindari label yang menyesatkan.

**Primary Users:** guru penyusun soal, reviewer/verifikator, operator paket soal.

**Commit message:** `feat(bank-soal): refine CBT question bank UX`

### Task 3.1: Rename “Analisis Butir” when true psychometrics unavailable

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/analisis-butir/+page.svelte`
- Modify sidebar/menu if label appears there: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`

**Replacement:**
- Page title: `Pemantauan Mutu Soal`
- Subtitle: explain this is metadata/workflow readiness, not empirical item analysis.

**Acceptance Criteria:**
- No user-facing claim of item difficulty/discrimination unless metrics exist.

### Task 3.2: Clean Bank Soal technical/English labels

**Files likely:**
- `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/CatalogShortcutPanel.svelte`
- `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`
- `apps/web-admin/src/routes/bank-soal/mapel-kd/+page.svelte`
- `apps/web-admin/src/routes/bank-soal/pengaturan/+page.svelte`

**Replace:**
- `Advanced Bank Soal` -> `Pengelolaan Lanjutan Bank Soal`
- `Mapel & KD Coverage` -> `Cakupan Mapel & KD`
- `Export` -> `Ekspor`
- `Penugasan Event` -> `Penugasan Kegiatan`
- `Published` -> `Terbit`
- Remove visible `Sprint`, `additive`, `hard enforcement`, `foundation` from user UI.

### Task 3.3: Fix “Telah Diverifikasi” count/filter mismatch

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalListPage.svelte`

**Options:**
- Preferred: split cards into `Disetujui` and `Terbit`.
- Alternative: click `Telah Diverifikasi` filters both approved + published.

**Acceptance Criteria:**
- Count shown equals list after clicking.

### Task 3.4: Add warning/guard before approving incomplete questions

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/ReviewBankSoalPage.svelte`

**Rule:**
- Objective/multiple choice must have answer key.
- Essay must have rubric or at least show strong warning if rubric missing.

**UI:**
- Disable `Setujui` or show confirmation modal with red warning.
- Copy:

```text
Soal belum memiliki kunci/rubrik yang lengkap. Setujui hanya jika verifikator sudah memastikan kelengkapan secara manual.
```

**Acceptance Criteria:** reviewer cannot accidentally approve incomplete question without seeing warning.

### Task 3.5: Clarify Bank Soal vs Asesmen domain

**Files:** same Bank Soal components.

**Copy:**

```text
Bank Soal menyimpan butir soal. Paket Soal mengambil soal terbit dari Bank Soal. Sesi Ujian dikelola di Asesmen CBT.
```

**Acceptance Criteria:** no confusion that Bank Soal itself schedules sessions.

### Task 3.6: Sprint 3 verification and commit

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/cbt-backend-paths.test.ts src/lib/server/assessment-route-contract.test.ts
```

**Commit:**

```bash
git add apps/web-admin/src/routes/bank-soal apps/web-admin/src/lib/components/sidebar docs/contracts
git commit -m "feat(bank-soal): refine CBT question bank UX"
```

---

## Sprint UX CBT 4 — Mobile CBT Copy Alignment + A11y/Responsive Polish

**Goal:** Menyamakan bahasa siswa/pengawas, mengurangi istilah teknis, dan meningkatkan accessibility/responsive UI.

**Primary Users:** siswa, pengawas, admin yang memakai tablet/HP.

**Commit message:** `feat(cbt): polish student copy and accessibility`

### Task 4.1: Align mobile status terms

**Files:**
- Modify: `apps/mobile/lib/src/exam_error_messages.dart`
- Modify: `apps/mobile/lib/src/screens/exam_shell_screen.dart`
- Modify: `apps/mobile/lib/src/screens/exam_shell_widgets.dart`
- Modify web guide if needed: `apps/web-admin/src/routes/asesmen/aplikasi-siswa/+page.svelte`

**Canonical student terms:**
- `Aman`
- `Perlu Sinkron`
- `Perlu Pengawas`

**Acceptance Criteria:** web student guide maps technical statuses to these student terms.

### Task 4.2: Improve 423 locked copy tone

**Files:**
- Modify: `apps/mobile/lib/src/exam_error_messages.dart`
- Search related mobile screens for locked copy.

**Preferred copy:**

```text
Akses ujian Anda sedang ditahan sementara oleh sistem/pengawas. Jawaban tetap aman di perangkat ini. Tetap di tempat, jangan login berulang, dan panggil pengawas untuk membuka akses jika sudah boleh lanjut.
```

**Acceptance Criteria:** copy is calm, not punitive.

### Task 4.3: Reduce technical mobile terms

**Files:**
- Modify: `apps/mobile/lib/src/screens/exam_login_screen.dart`
- Modify: other screen files if terms appear.

**Replace:**
- `Heartbeat aktif` -> `Status dipantau`
- `Anti-switch dasar` -> `Tetap di aplikasi`
- `Simpan bertahap` -> `Jawaban tersimpan`
- `fingerprint` -> `penanda perangkat`
- `32 karakter heksadesimal` -> `token dari kartu ujian/pengawas`

### Task 4.4: Make restore failure CTA safer

**Files:**
- Modify: `apps/mobile/lib/src/screens/exam_restore_failed_screen.dart`

**Behavior:**
- For 409/423/connection errors: primary CTA should be `Panggil Pengawas` or guidance card, not direct login retry.
- Keep `Kembali ke Login` as secondary if necessary.

**Acceptance Criteria:** student is not encouraged to repeatedly login when locked/device conflict/pending sync exists.

### Task 4.5: Hide raw URL/type technical details from student main copy

**Files:**
- Modify media/question widgets in `apps/mobile/lib/src/screens/exam_shell_widgets.dart` or relevant file.

**Replace:**
- Raw media URL in error -> friendly message; technical details hidden or labeled `Info untuk pengawas`.
- `question_type` -> `Kode tipe soal untuk pengawas`.

### Task 4.6: Add ARIA semantics to web tabs/filters

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/sesi/+page.svelte` if filter cards need labels.

**Acceptance Criteria:**
- Tab container uses `role="tablist"`.
- Buttons use `role="tab"`, `aria-selected`, `aria-controls`.
- Panels use `role="tabpanel"`.
- Filter cards have `aria-label="Filter sesi: ..."`.

### Task 4.7: Make SOP timeline mobile-friendly

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`

**Behavior:**
- Timeline collapsed by default on mobile or hidden behind `Lihat SOP formal` accordion.
- Default view shows `Langkah berikutnya` and concise readiness.

### Task 4.8: Improve empty states with CTA

**Files likely:**
- `apps/web-admin/src/routes/asesmen/kegiatan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`

**Examples:**
- `Belum ada kegiatan ujian` -> button `Buat Kegiatan`.
- `Belum ada sesi ujian` -> button `Buat Sesi` / `Kelola Paket`.
- `Tidak ada sesi pada filter ini` -> `Reset Filter`.

### Task 4.9: Sprint 4 verification and commit

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/cbt-backend-paths.test.ts src/lib/server/assessment-route-contract.test.ts
cd apps/mobile
if command -v flutter >/dev/null 2>&1; then flutter analyze && flutter test; else echo 'flutter not installed'; fi
```

**Commit:**

```bash
git add apps/mobile/lib/src apps/web-admin/src/routes/asesmen docs/contracts
git commit -m "feat(cbt): polish student copy and accessibility"
```

---

## Sprint UX CBT 5 — Final Integration Review & Production Readiness

**Goal:** Memastikan semua sprint konsisten, tidak ada UX route patah, dan siap diajukan deploy dengan urutan aman.

**Commit message:** `docs(asesmen): document CBT UIUX rollout readiness`

### Task 5.1: End-to-end route audit

**Commands:**

```bash
git status --short
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/cbt-backend-paths.test.ts src/lib/server/assessment-route-contract.test.ts
```

**Check manually:**
- `/asesmen`
- `/asesmen/persiapan`
- `/asesmen/kegiatan`
- `/asesmen/kegiatan/[id]`
- `/asesmen/kegiatan/[id]/archive`
- `/asesmen/pelaksanaan`
- `/asesmen/pengawasan`
- room proctoring page
- `/asesmen/hasil`
- Bank Soal key pages
- mobile copy compile/analyze if Flutter available

### Task 5.2: Full verification

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts src/lib/server/cbt-backend-paths.test.ts src/lib/server/assessment-route-contract.test.ts
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-cbt-uiux-final ./cmd/api
cd ../../apps/mobile
if command -v flutter >/dev/null 2>&1; then flutter analyze && flutter test; else echo 'flutter not installed'; fi
```

### Task 5.3: Static secret scan

```bash
git diff --check
git diff HEAD | grep '^+' | grep -iE '(api_key|secret|password|token|passwd)\s*=\s*["'"'][^"'"']{6,}["'"']' || true
```

Expected: no hardcoded secrets. Note: textual `token` labels in UI are expected; only assignments to literal secrets are blockers.

### Task 5.4: Update final rollout notes

**Files:**
- Create: `docs/reviews/asesmen-cbt-uiux-implementation-readiness-2026-05-17.md`

**Include:**
- Completed sprint list.
- Verification results.
- Routes added/changed.
- Known limitations.
- Deployment prerequisites.
- Explicit note: no production deploy/restart performed.

### Task 5.5: Final commit

```bash
git add docs/reviews/asesmen-cbt-uiux-implementation-readiness-2026-05-17.md
git commit -m "docs(asesmen): document CBT UIUX rollout readiness"
```

---

## Deployment Plan — Only After Explicit Approval

Do not run this section unless user explicitly says to deploy/restart.

### Pre-deploy

```bash
git status --short
git log --oneline -5
```

If migrations changed:
- Backup DB first.
- Run migration in production using existing repo workflow without printing secrets.
- Confirm migration success.

### Backend deploy order

1. Apply migrations if any.
2. Build core-api.
3. Restart PM2 `mtsn2kolut-core-api`.
4. Health check `http://127.0.0.1:8080/health`.

### Web-admin deploy order

1. `npm --prefix apps/web-admin run build`
2. Restart PM2 `mtsn2kolut-web-admin` immediately after build to avoid stale manifest/chunk errors.
3. Smoke test key pages.

### Smoke test pages

- `/asesmen`
- `/asesmen/kegiatan`
- `/asesmen/kegiatan/{existing_id}`
- `/asesmen/kegiatan/{existing_id}/archive`
- `/asesmen/pengawasan`
- room proctoring page if existing session available
- `/bank-soal`

---

## Suggested Execution Order

1. Sprint 0 — baseline and terminology checklist.
2. Sprint 1 — Simple Proctor Mode + terminology cleanup.
3. Sprint 2 — Archive Kegiatan + visible approval panel.
4. Sprint 3 — Bank Soal UX refinement.
5. Sprint 4 — Mobile copy + a11y/responsive polish.
6. Sprint 5 — integration review and readiness doc.

Each sprint should be implemented and committed independently to keep rollback easy.

## Acceptance Criteria for “All Done”

- Pengawas has a simple mode for live room operations.
- Token Ruang/Token Ujian terminology is consistent in operator UI.
- Detail Kegiatan has visible pengesahan SOP panel.
- Archive route exists and is not broken.
- Token raw is not default in final archive/BA context.
- Bank Soal labels no longer misrepresent “Analisis Butir” if psychometrics are absent.
- Mobile CBT copy is calm, consistent, and avoids technical terms where possible.
- Web tabs/filters have improved accessibility labels.
- All required checks pass.
- No deploy/restart was done without explicit approval.
