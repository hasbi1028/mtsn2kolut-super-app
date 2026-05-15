# Asesmen Kelengkapan Soal Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menambah fitur monitoring kelengkapan soal per kegiatan ujian, dengan dukungan target tingkat/rombel, guru-mapel-rombel, dan mode perakitan soal yang fleksibel.

**Architecture:** Fitur ditempatkan di detail kegiatan asesmen sebagai tab `Kelengkapan Soal`. Backend menghitung matriks kebutuhan dari target kegiatan + jadwal mengajar/rombel + bank soal. Implementasi bertahap: mulai read-only report, lalu konfigurasi target, lalu integrasi paket/randomisasi.

**Tech Stack:** SvelteKit web-admin, Go core-api, PostgreSQL migrations, sqlc queries, existing CBT/Bank Soal/Academic modules.

---

## Prinsip Desain

1. Monitoring mengikuti `target_levels`/target kelas pada kegiatan.
2. Default unit monitoring: `guru + mapel + rombel`.
3. Kelas yang tidak masuk target kegiatan tidak dihitung sebagai kurang.
4. Monitoring kelengkapan soal dipisah dari perakitan paket/randomisasi.
5. Implementasi awal aman: read-only rekap dulu, tidak mengubah alur Bank Soal/Paket yang sudah berjalan.

---

## Phase 0: Discovery dan Konfirmasi Data Existing

### Task 0.1: Audit struktur data Bank Soal, mapel, rombel, jadwal mengajar

**Objective:** Pastikan field yang bisa dipakai untuk menghitung kebutuhan soal.

**Files to inspect:**
- `services/core-api/db/migrations/*bank*`
- `services/core-api/db/queries/*bank*`
- `services/core-api/db/queries/academic*.sql`
- `services/core-api/db/queries/asesmen*.sql`
- `apps/web-admin/src/routes/bank-soal/**`
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`

**Verify:**
- Ada relasi soal ke mapel/tingkat/kelas/guru/pembuat.
- Ada data jadwal/assignment guru mengajar per rombel dan mapel.
- Ada tipe soal PG/essay yang bisa dihitung.
- Ada status soal draft/review/published.

### Task 0.2: Tentukan aturan hitung soal

**Objective:** Dokumentasikan definisi “terkumpul” dan “lengkap”.

**Recommended default:**
- Soal dihitung jika:
  - sesuai kegiatan atau sesuai filter mapel/tingkat/rombel kegiatan,
  - dibuat oleh guru yang ditugaskan/mengajar,
  - tipe soal sesuai target: PG/essay,
  - status minimal `published` untuk kesiapan final.
- Tambahkan opsi filter status:
  - `draft + review + published` untuk progres awal,
  - `published only` untuk kesiapan final.

**Verify:**
- Panitia bisa membedakan “sudah dibuat” vs “sudah siap pakai”.

---

## Phase 1: Read-only Kelengkapan Soal di Detail Kegiatan

### Task 1.1: Tambah query backend untuk matriks kebutuhan soal

**Objective:** Buat endpoint read-only yang mengembalikan progres kelengkapan soal per guru-mapel-rombel.

**Proposed endpoint:**
- `GET /api/asesmen/events/{id}/question-completeness`

**Core response shape:**
```json
{
  "event": {
    "id": "...",
    "title": "Ujian Semester Genap",
    "target_levels": ["VII", "VIII"]
  },
  "summary": {
    "total_rows": 48,
    "complete_rows": 30,
    "incomplete_rows": 18,
    "not_started_rows": 6,
    "pg_required": 960,
    "pg_available": 720,
    "essay_required": 240,
    "essay_available": 180
  },
  "rows": [
    {
      "level": "VII",
      "class_id": "...",
      "class_code": "VII A",
      "subject_id": "...",
      "subject_name": "Matematika",
      "teacher_id": "...",
      "teacher_name": "Guru 1",
      "target_pg": 20,
      "available_pg": 18,
      "target_essay": 5,
      "available_essay": 5,
      "status": "incomplete",
      "missing_pg": 2,
      "missing_essay": 0
    }
  ],
  "excluded_levels": ["IX"]
}
```

**Backend files likely modified:**
- `services/core-api/db/queries/asesmen.sql` or new query file if existing convention allows.
- Generated sqlc file after running sqlc.
- `services/core-api/internal/service/...`
- `services/core-api/internal/handler/...`
- `services/core-api/cmd/api/main.go`

**Safety:**
- No writes.
- No package changes.
- No Bank Soal mutations.

### Task 1.2: Add SvelteKit proxy route

**Objective:** Web-admin can fetch the new backend endpoint.

**Files:**
- Create/modify: `apps/web-admin/src/routes/api/asesmen/events/[id]/question-completeness/+server.ts`

**Behavior:**
- Require authenticated user.
- Proxy to core-api endpoint.
- Use existing `proxy(event).fetch(...)` and `jsonProxyResponse(...)` pattern.

### Task 1.3: Add tab “Kelengkapan Soal” to detail kegiatan

**Objective:** Surface read-only monitoring inside event detail.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`

**UI changes:**
- Add tab after `Persiapan`:
  - `Ringkasan`
  - `Persiapan`
  - `Kelengkapan Soal`
  - `Operasional`
  - `Hasil`
- Fetch `/api/asesmen/events/${eventId}/question-completeness`.
- Show banner:
  - `Target kegiatan ini: VII, VIII. Kelas IX tidak dihitung.`
- Show cards:
  - Total lengkap
  - Belum lengkap
  - Belum mulai
  - Progres PG/Essay
- Show list/matrix rows.

### Task 1.4: Add filters and CSV export client-side

**Objective:** Panitia can quickly find incomplete rows.

**Filters:**
- Tingkat
- Rombel
- Mapel
- Guru
- Status: Semua, Lengkap, Belum lengkap, Belum mulai

**CSV columns:**
- Tingkat
- Rombel
- Mapel
- Guru
- PG
- Essay
- Status
- Kekurangan

**Safety:**
- Client-side only.
- No DB changes.

### Task 1.5: Tests for Phase 1

**Objective:** Prevent regression in route and UI rendering.

**Tests:**
- Backend service test for:
  - target VII/VIII excludes IX.
  - different teacher same subject different rombel becomes separate rows.
  - status complete/incomplete/not_started computed correctly.
- Web-admin page-source test for:
  - tab label `Kelengkapan Soal` exists.
  - target banner exists.
  - filter labels exist.

**Commands:**
- `go test ./...` in `services/core-api`
- existing web-admin test/check command per repo convention.

---

## Phase 2: Konfigurasi Target Soal per Kegiatan

### Task 2.1: Add DB table for event question requirements

**Objective:** Target tidak hard-coded 20 PG + 5 essay.

**Migration idea:**
- Create table: `cbt_event_question_requirements`
- Columns:
  - `id`
  - `event_id`
  - `scope_mode`: `per_rombel`, `per_level`, `pool_level_subject`
  - `level`
  - `class_id` nullable
  - `subject_id` nullable
  - `target_pg` default 20
  - `target_essay` default 5
  - timestamps

**Initial scope:**
- Add default row at event level only:
  - `scope_mode = per_rombel`
  - `target_pg = 20`
  - `target_essay = 5`

**Safety:**
- Migration additive only.
- Existing features unaffected.

### Task 2.2: Add settings panel in Kelengkapan Soal

**Objective:** Admin can configure target for the event.

**UI location:**
- `Asesmen → Kegiatan → Detail → Kelengkapan Soal → Pengaturan Target`

**Fields:**
- Mode monitoring:
  - Per Rombel (default)
  - Per Tingkat
  - Pool Bersama Tingkat-Mapel
- Target PG
- Target Essay
- Status soal yang dihitung:
  - Semua progres
  - Published only

**Safety:**
- Keep default if no config exists.
- Confirm before changing mode if rows already exist.

### Task 2.3: Backend CRUD for target config

**Endpoints:**
- `GET /api/asesmen/events/{id}/question-requirements`
- `PUT /api/asesmen/events/{id}/question-requirements`

**Permissions:**
- Read: `asesmen.read`
- Write: `asesmen.event_manage` or existing equivalent.

---

## Phase 3: Deep Linking to Bank Soal

### Task 3.1: Add filtered “Buka Bank Soal” action per row

**Objective:** Dari baris yang kurang, admin/guru bisa langsung membuka Bank Soal dengan filter sesuai.

**Example link:**
- `/bank-soal?tambah?event_id=...&subject_id=...&class_id=...&teacher_id=...`
- Or existing composer link if route supports it:
  - `/bank-soal/tambah?event_id=...&subject_id=...&level=VII&class_id=...`

**Safety:**
- Link only; no automatic creation.

### Task 3.2: Add event/requirement context in Bank Soal composer

**Objective:** Saat guru membuat soal dari link, field kegiatan/mapel/kelas terisi otomatis.

**Behavior:**
- Pre-fill event, subject, level, class/rombel.
- Do not lock fields unless explicit.
- Show info: `Soal ini dihitung untuk Kelengkapan Soal kegiatan ...`.

---

## Phase 4: Pool Bersama dan Mode Per Tingkat

### Task 4.1: Implement aggregation mode `per_level`

**Objective:** Satu target mapel per tingkat, bukan per rombel.

**Example:**
- Matematika VII = 20 PG + 5 Essay untuk semua VII A-D.

**Use case:**
- Satu paket bersama untuk semua rombel tingkat yang sama.

### Task 4.2: Implement aggregation mode `pool_level_subject`

**Objective:** Beberapa guru mapel tingkat yang sama menyetor ke pool bersama.

**Example:**
- Guru 1 dan Guru 2 sama-sama Matematika VII.
- Target pool bisa 80 PG + 20 Essay, lalu paket per rombel dirakit acak.

**UI:**
- Show contribution per guru.
- Show pool total per mapel/tingkat.

**Safety:**
- Keep `per_rombel` default.
- Pool mode opt-in only.

---

## Phase 5: Integrasi Paket dan Randomisasi

### Task 5.1: Add package readiness from completeness rows

**Objective:** Paket tidak dibuat dari sumber yang belum cukup.

**Behavior:**
- Show warning if package source has insufficient questions.
- Do not block hard initially; use warning.

### Task 5.2: Add package source mode

**Options:**
- Dari guru pengampu rombel.
- Dari semua guru mapel pada tingkat tersebut.
- Dari pool kegiatan.

### Task 5.3: Add randomization controls

**Controls:**
- Randomize question order.
- Randomize answer option order.
- Draw N PG + M Essay from eligible pool.
- Optional seed per session/rombel for reproducibility.

**Safety:**
- Log generated package composition.
- Allow preview before save.
- Do not mutate source questions.

---

## Phase 6: Notifications and Operational Extras

### Task 6.1: Export incomplete list

**Objective:** Panitia can share list manually.

**Output:**
- CSV/Excel-like CSV.
- Grouped by guru.

### Task 6.2: Reminder draft

**Objective:** Generate reminder text per guru.

**No auto-send initially.**
- Copy text.
- Later integrate WhatsApp after approval.

### Task 6.3: Dashboard summary on event list

**Objective:** Page `Asesmen → Kegiatan` shows simple readiness badge.

**Examples:**
- `Soal 80% lengkap`
- `12 belum lengkap`

---

## Recommended Execution Order

1. Phase 0 discovery.
2. Phase 1 read-only report.
3. Review with real data from one kegiatan.
4. Phase 2 configurable targets.
5. Phase 3 deep link to Bank Soal.
6. Phase 4 pool/per-level modes only if needed after trial.
7. Phase 5 randomization/package automation after monitoring is trusted.
8. Phase 6 notification/export.

---

## Acceptance Criteria MVP

MVP is done when:

1. Admin opens a kegiatan and sees tab `Kelengkapan Soal`.
2. Kegiatan Semester Genap target VII/VIII excludes IX from missing count.
3. Same mapel with different teachers across VII A-D appears as separate guru-mapel-rombel rows.
4. Default target 20 PG + 5 Essay is applied.
5. Rows show complete/incomplete/not_started correctly.
6. Admin can filter incomplete rows and export CSV.
7. No existing Bank Soal, Paket, Sesi, or Hasil flow is broken.

---

## Commit Boundaries

- `docs: add asesmen question completeness plan`
- `feat(api): add question completeness read model`
- `feat(web): add question completeness tab`
- `test: cover question completeness target filtering`
- `feat(api): add event question requirement settings`
- `feat(web): add question requirement settings panel`
- `feat(web): link completeness rows to bank soal filters`
- Later: package/randomization commits separately.

---

## Rollback Strategy

- Phase 1 can be disabled by hiding the tab; no DB write.
- Phase 2 migration is additive; existing code ignores table if feature disabled.
- Package/randomization changes must be feature-flagged or isolated until verified.

---

## Recommended MVP Scope

Build only:
- read-only completeness report,
- default target 20 PG + 5 Essay,
- target_levels exclusion,
- per guru-mapel-rombel rows,
- filters,
- CSV export.

Delay:
- pooled random package generation,
- WhatsApp reminders,
- automatic package creation.
