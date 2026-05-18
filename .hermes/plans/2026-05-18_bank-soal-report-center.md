# Bank Soal Report Center Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build a permanent Bank Soal reporting center so admins/panitia can view, filter, export, and share Bank Soal input/progress/honor reports without asking AI/manual SQL.

**Architecture:** Core API owns all report queries, policy, and export generation. SvelteKit web-admin stays a thin BFF/proxy and renders an operator-friendly report UI. Export PNG/PDF/Excel uses backend-generated report data/templates so outputs are consistent and audit-safe.

**Tech Stack:** Go 1.26 + Chi + sqlc + PostgreSQL in `services/core-api`; SvelteKit 2/Svelte 5 + Tailwind/shadcn in `apps/web-admin`; headless Chrome or HTML-to-PNG helper for PNG export if approved; CSV/XLSX/PDF export in later sprint.

---

## Product Scope — Opsi 3

Create a **Bank Soal Report Center** with these report tabs:

1. **Input Soal** — by pembuat + mapel + tingkat + status, like the image report already generated manually.
2. **Progres Mapel** — target vs actual per mapel/tingkat/kegiatan; highlights belum lengkap.
3. **Perlu Revisi** — questions awaiting teacher revision, grouped by pembuat/mapel/tingkat/reviewer note status.
4. **Reviewer & Verifikasi** — workload/status for reviewers/approvers.
5. **Siap Paket** — published/approved question pool readiness for package building.
6. **Honor/Tugas** — volume per person for SK/honor basis: pembuat soal, reviewer, korektor essay, proktor/pengawas if later linked to UAS event roles.

## Non-Negotiable Constraints

- No direct PostgreSQL from SvelteKit. All report data must come from Go Core API.
- No production deploy/restart until explicit approval.
- Do not expose credentials, tokens, answer keys, or full question content in broad reports.
- Export should include metadata: generated_at WITA, filters, source system, and note for `system`/seed rows.
- Keep reports permission-scoped:
  - `admin`/panitia/kurikulum: all reports.
  - guru biasa: own input only unless assigned reviewer/panitia.
  - reviewer: scope by mapel/tingkat/event when reviewer scope is active.
- Existing Bank Soal authoring/review flow must not be changed in this plan.

## Recommended Route/Menu

Frontend route:

```text
/bank-soal/laporan
```

Sidebar:

```text
Bank Soal
- Daftar Soal
- Tambah Soal
- Import Soal
- Verifikasi
- Laporan
```

BFF routes:

```text
/apps/web-admin/src/routes/api/bank-soal/reports/+server.ts
/apps/web-admin/src/routes/api/bank-soal/reports/export/+server.ts
```

Core API routes:

```text
GET  /api/cbt/questions/reports
POST /api/cbt/questions/reports/export
```

Compatibility BFF exposes them as `/api/bank-soal/reports*` for the UI domain split.

---

## Sprint 0 — Safe Rollout Gate & Contract

### Task 0.1: Audit current Bank Soal reporting data

**Objective:** Confirm fields and data quality before writing report contracts.

**Files:**
- Read: `services/core-api/db/migrations/004_cbt_foundation.sql`
- Read: `services/core-api/db/migrations/062_cbt_event_members_question_scope.sql`
- Read: `services/core-api/db/queries/cbt_questions.sql`
- Read: `services/core-api/db/queries/bank_soal_reviewer_scopes.sql`

**Steps:**
1. Query safe counts only: total questions by month/status/type/author/mapel/tingkat.
2. Query number of rows with missing `target_level`, missing `subject_id`, `author_username='system'`.
3. Query active events and whether questions have `event_id` populated.
4. Do not print DB credentials.

**Verification:**
- Produce a short audit note in plan implementation notes.
- No code changes.

### Task 0.2: Create report API contract document

**Objective:** Document filters, response shapes, export shapes, and access policy before implementation.

**Files:**
- Create: `docs/contracts/bank-soal-report-center.md`

**Required sections:**
- Report tabs.
- Filter model.
- Summary cards.
- Table columns per tab.
- Export formats.
- Access policy.
- Data-quality flags.

**Filter contract:**

```ts
type BankSoalReportFilters = {
  report: 'input' | 'progress' | 'revision' | 'reviewer' | 'readiness' | 'honor';
  periodPreset?: 'today' | 'this_week' | 'this_month' | 'custom';
  startDate?: string; // YYYY-MM-DD WITA
  endDate?: string;   // YYYY-MM-DD WITA inclusive UI, exclusive backend +1 day
  eventId?: string;
  subjectId?: string;
  targetLevel?: 'VII' | 'VIII' | 'IX' | '';
  authorUsername?: string;
  workflowStatus?: string;
  includeSystem?: boolean;
  groupBy?: 'author_subject_level' | 'author' | 'subject' | 'level' | 'status';
};
```

**Verification:**
- Contract can be read by both frontend/backend implementers.
- User-facing labels are Indonesian.

**Commit:**

```bash
git add docs/contracts/bank-soal-report-center.md .hermes/plans/2026-05-18_bank-soal-report-center.md
git commit -m "docs(bank-soal): plan report center"
```

---

## Sprint 1 — Core API Input Report

### Task 1.1: Add sqlc query for input recap

**Objective:** Add typed SQL for the first tab: Input Soal grouped by pembuat + mapel + tingkat.

**Files:**
- Modify: `services/core-api/db/queries/cbt_questions.sql`
- Generated later: `services/core-api/internal/repository/postgres/cbt_questions.sql.go`

**Add query concept:**

```sql
-- name: GetCbtQuestionInputReport :many
WITH filtered AS (
  SELECT
    q.id,
    q.author_username,
    q.subject_id,
    q.target_level,
    q.question_type::text AS question_type,
    q.workflow_status::text AS workflow_status,
    (q.created_at AT TIME ZONE 'Asia/Makassar') AS created_wita,
    COALESCE(NULLIF(e.nama,''), NULLIF(u.display_name,''), NULLIF(q.author_username,''), '[Tidak tercatat]') AS author_name,
    COALESCE(NULLIF(s.name,''), '[Mapel tidak tercatat]') AS subject_name
  FROM cbt_questions q
  LEFT JOIN subjects s ON s.id = q.subject_id
  LEFT JOIN users u ON u.username = q.author_username
  LEFT JOIN employees e ON e.id = u.employee_id
  WHERE q.created_at >= sqlc.arg(start_at)::timestamptz
    AND q.created_at < sqlc.arg(end_at)::timestamptz
    AND (sqlc.arg(include_system)::boolean OR COALESCE(q.author_username,'') <> 'system')
    AND (sqlc.narg(subject_id)::uuid IS NULL OR q.subject_id = sqlc.narg(subject_id)::uuid)
    AND (COALESCE(sqlc.narg(target_level)::text, '') = '' OR q.target_level = sqlc.narg(target_level)::text)
    AND (COALESCE(sqlc.narg(author_username)::text, '') = '' OR q.author_username = sqlc.narg(author_username)::text)
    AND (COALESCE(sqlc.narg(workflow_status)::text, '') = '' OR q.workflow_status::text = sqlc.narg(workflow_status)::text)
), status_by_group AS (
  SELECT author_name, author_username, subject_name, target_level, workflow_status, COUNT(*) AS cnt
  FROM filtered
  GROUP BY author_name, author_username, subject_name, target_level, workflow_status
)
SELECT
  f.author_name,
  COALESCE(NULLIF(f.author_username,''), '-') AS username,
  f.subject_name,
  COALESCE(NULLIF(f.target_level,''), 'Belum tercatat') AS level_name,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE f.question_type = 'multiple_choice') AS pg,
  COUNT(*) FILTER (WHERE f.question_type = 'essay') AS essay,
  COUNT(*) FILTER (WHERE f.question_type NOT IN ('multiple_choice','essay')) AS other,
  MIN(f.created_wita) AS first_input,
  MAX(f.created_wita) AS last_input,
  COALESCE(
    string_agg(DISTINCT s.workflow_status || ':' || s.cnt::text, ', ' ORDER BY s.workflow_status || ':' || s.cnt::text),
    ''
  ) AS statuses
FROM filtered f
LEFT JOIN status_by_group s
  ON s.author_name = f.author_name
 AND COALESCE(s.author_username,'') = COALESCE(f.author_username,'')
 AND s.subject_name = f.subject_name
 AND COALESCE(s.target_level,'') = COALESCE(f.target_level,'')
GROUP BY f.author_name, f.author_username, f.subject_name, f.target_level
ORDER BY CASE WHEN COALESCE(f.author_username,'') = 'system' THEN 1 ELSE 0 END, f.author_name, f.subject_name, f.target_level;
```

**Verification:**

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
```

Expected: exit 0.

### Task 1.2: Add service DTO and date-filter normalization

**Objective:** Convert query params into safe WITA start/end timestamps and report DTO.

**Files:**
- Modify/Create: `services/core-api/internal/service/cbt_question_report.go`
- Test: `services/core-api/internal/service/cbt_question_report_test.go`

**Rules:**
- `this_month` = WITA month start to next month start.
- `today` = WITA day start to next day start.
- `custom` requires start/end and converts end inclusive UI to exclusive backend.
- Max range initially 1 year to avoid heavy accidental exports.
- `includeSystem=false` by default.

**Verification:**

```bash
cd services/core-api
go test ./internal/service -run 'TestCbtQuestionReport'
```

Expected: pass.

### Task 1.3: Add handler endpoint for input report

**Objective:** Expose `GET /api/cbt/questions/reports?report=input&periodPreset=this_month`.

**Files:**
- Modify/Create: `services/core-api/internal/handler/cbt_question_report.go`
- Modify: `services/core-api/cmd/api/main.go`
- Test: `services/core-api/internal/handler/cbt_question_report_test.go`

**Response shape:**

```json
{
  "filters": { "report": "input", "period_preset": "this_month" },
  "generated_at": "2026-05-18T23:00:00+08:00",
  "summary": {
    "total": 234,
    "pg": 187,
    "essay": 46,
    "other": 1,
    "authors": 10,
    "subjects": 8,
    "missing_level": 1,
    "system_rows": 25
  },
  "statuses": [{ "status": "submitted", "total": 69 }],
  "rows": []
}
```

**Access policy:**
- Require authenticated user.
- Admin or `bank_soal.analytics` can see all.
- Guru without analytics sees own rows only.

**Verification:**

```bash
cd services/core-api
go test ./internal/handler -run 'TestCbtQuestionReport'
go test ./internal/handler ./internal/service ./internal/repository/postgres
```

Expected: pass.

### Task 1.4: Add BFF proxy route

**Objective:** Keep web-admin as BFF/proxy only.

**Files:**
- Create: `apps/web-admin/src/routes/api/bank-soal/reports/+server.ts`

**Behavior:**
- Forward query string to Core API `/api/cbt/questions/reports`.
- Preserve auth/session behavior matching existing Bank Soal BFF routes.
- Do not access DB.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: 0 errors/warnings.

---

## Sprint 2 — Report Center UI Shell + Input Tab

### Task 2.1: Add report page route

**Objective:** Create the Bank Soal report center page.

**Files:**
- Create: `apps/web-admin/src/routes/bank-soal/laporan/+page.svelte`
- Optional create: `apps/web-admin/src/routes/bank-soal/_components/BankSoalReportFilters.svelte`
- Optional create: `apps/web-admin/src/routes/bank-soal/_components/BankSoalReportSummaryCards.svelte`
- Optional create: `apps/web-admin/src/routes/bank-soal/_components/BankSoalInputReportTable.svelte`

**UI layout:**
- Header: `Laporan Bank Soal`.
- Tabs: Input Soal, Progres Mapel, Perlu Revisi, Reviewer, Siap Paket, Honor/Tugas.
- Disabled/coming-soon state for tabs not implemented in Sprint 2.
- Filter bar with period, mapel, tingkat, pembuat, status, include/exclude system.
- Summary cards.
- Table.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

Expected: 0 errors/warnings.

### Task 2.2: Add operator-friendly table states

**Objective:** Make report useful during real operations.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/_components/BankSoalInputReportTable.svelte`

**States:**
- Loading skeleton.
- Empty: `Belum ada soal pada periode ini`.
- Error with retry.
- Blue highlight for `system` rows.
- Orange highlight for missing `Belum tercatat` level/mapel.

**Verification:**
- Manual browser check `/bank-soal/laporan` authenticated.
- Confirm filters refresh data.

### Task 2.3: Add sidebar/menu entry

**Objective:** Make feature discoverable.

**Files:**
- Modify: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Possibly modify: `apps/web-admin/src/lib/server/route-access.ts`

**Rule:**
- Show to admin/panitia/kurikulum/reviewer users with report access.
- Hide for siswa/ortu.

**Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

Expected: pass.

**Commit:**

```bash
git add apps/web-admin/src/routes/bank-soal/laporan apps/web-admin/src/routes/api/bank-soal/reports apps/web-admin/src/lib/components/sidebar/sidebar-config.ts
git commit -m "feat(bank-soal): add report center input tab"
```

---

## Sprint 3 — Export Gambar PNG + PDF/Excel Foundation

### Task 3.1: Decide export renderer

**Objective:** Pick safe export approach before code.

**Preferred option:** Core API creates HTML report and renders PNG using an installed renderer if allowed in deployment. If headless Chrome dependency is considered too heavy for Core API, add an internal web-admin server export endpoint that calls Core API for data and uses browser rendering, but still never queries DB directly.

**Decision criteria:**
- Existing server has `google-chrome` installed.
- Output must be stable for Telegram/WhatsApp.
- No secrets in generated HTML.
- Temp files cleaned.

**Files:**
- Update: `docs/contracts/bank-soal-report-center.md`

### Task 3.2: Add export request/response contract

**Objective:** Standardize export workflow.

**Endpoint:**

```text
POST /api/cbt/questions/reports/export
```

**Body:**

```json
{
  "report": "input",
  "format": "png",
  "filters": { "periodPreset": "this_month", "includeSystem": true }
}
```

**Response:**

```json
{
  "file_name": "laporan-bank-soal-input-2026-05.png",
  "content_type": "image/png",
  "download_url": "/api/cbt/questions/reports/export/files/..."
}
```

or direct binary response if simpler.

### Task 3.3: Implement PNG template for Input Soal

**Objective:** Generate a polished PNG like the manual image report.

**Files:**
- Create: `services/core-api/internal/service/report_templates/bank_soal_input_report.html.tmpl`
- Create/Modify: `services/core-api/internal/service/cbt_question_report_export.go`
- Modify: `services/core-api/internal/handler/cbt_question_report.go`

**Template must include:**
- Title.
- Period WITA.
- Summary cards.
- Workflow badges.
- Table.
- Notes for `system` and missing data.

**Verification:**
- Unit test template does not include raw secrets.
- Manual export produces PNG and opens as valid image.

### Task 3.4: Add frontend export buttons

**Objective:** Allow operators to download the generated image.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/laporan/+page.svelte`
- Optional create: `apps/web-admin/src/routes/api/bank-soal/reports/export/+server.ts`

**Buttons:**

```text
Download Gambar
Download Excel
Download PDF
```

Initially:
- PNG active.
- Excel/PDF can be disabled with label `Tahap berikutnya` unless implemented in same sprint.

**Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api && go test ./internal/handler ./internal/service ./internal/repository/postgres && go build -o /tmp/core-api-banksoal-report ./cmd/api
```

Expected: pass.

**Commit:**

```bash
git add services/core-api apps/web-admin docs/contracts/bank-soal-report-center.md
git commit -m "feat(bank-soal): export input report as image"
```

---

## Sprint 4 — Additional Report Tabs

### Task 4.1: Progres Mapel report

**Objective:** Show completeness per mapel/tingkat/event.

**Data:**
- Target PG/Essay from event requirements when available.
- Actual counts from `cbt_questions` filtered by event/mapel/tingkat/status.
- Missing target flags.

**Columns:**

```text
Mapel | Tingkat | Target PG | PG Masuk | Target Essay | Essay Masuk | Approved/Published | Kurang | Status
```

**Verification:**
- UAS/UTS panitia can identify mapel belum lengkap.

### Task 4.2: Perlu Revisi report

**Objective:** List revision-needed questions without exposing full content.

**Columns:**

```text
Pembuat | Mapel | Tingkat | Jumlah Perlu Revisi | Catatan Terakhir | Revisi Terakhir | Umur Revisi
```

**Guard:**
- Do not show full question stem/answer key in broad report.

### Task 4.3: Reviewer & Verifikasi report

**Objective:** Monitor review workload and status.

**Columns:**

```text
Reviewer | Scope Mapel/Tingkat | Menunggu Review | Direview | Minta Revisi | Disetujui | Rata-rata Waktu Review
```

**Data source:**
- `bank_soal_reviewer_scopes`.
- Question audit logs/workflow timestamps if available.

### Task 4.4: Siap Paket report

**Objective:** Show question pool readiness for package builder.

**Columns:**

```text
Mapel | Tingkat | Published | Approved | PG | Essay | Belum Dipakai | Sudah Dipakai | Siap Paket?
```

### Task 4.5: Implement tabs in UI

**Objective:** Wire all additional reports into the report center.

**Files:**
- Modify: `apps/web-admin/src/routes/bank-soal/laporan/+page.svelte`
- Add components under `apps/web-admin/src/routes/bank-soal/_components/BankSoal*ReportTable.svelte`
- Add Core API queries/service methods/handler switch cases.

**Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml && go test ./internal/handler ./internal/service ./internal/repository/postgres
```

Expected: pass.

---

## Sprint 5 — Honor/Tugas Report for SK & Payment Basis

### Task 5.1: Define honor report dimensions

**Objective:** Make report usable as SK/honor basis without deciding nominal rates.

**Dimensions:**

```text
Nama
Tugas
Mapel
Tingkat/Rombel
Kegiatan
Volume
Satuan
Bukti Data
Keterangan
```

**Supported tasks:**
- Pembuat Soal: volume by soal/paket/mapel/tingkat.
- Reviewer: volume by reviewed/approved/requested revision counts.
- Korektor Essay: volume by essay answers scored once scoring data is linked.
- Proktor/Pengawas: volume by session/room assignment once event member/session-room scope exists.

### Task 5.2: Add Honor/Tugas report endpoint

**Objective:** Aggregate work volume per person.

**Files:**
- Modify: `services/core-api/db/queries/cbt_questions.sql` or create `services/core-api/db/queries/bank_soal_reports.sql`
- Modify/Create: `services/core-api/internal/service/cbt_question_report.go`

**Rule:**
- Report volume only; do not calculate nominal honor unless a separate approved rate table exists.

### Task 5.3: Add Honor/Tugas UI tab and export

**Objective:** Produce SK/honor-friendly table.

**Columns:**

```text
No | Nama | Tugas | Mapel | Tingkat | Kegiatan | Volume | Satuan | Keterangan
```

**Exports:**
- PNG for sharing.
- Excel/PDF for administration.

**Verification:**
- Sample report can be used as Lampiran SK/rekap tugas.

---

## Sprint 6 — PDF/Excel Export + Telegram Share

### Task 6.1: Add Excel export

**Objective:** Give operators editable spreadsheet output.

**Option:** Generate CSV first if XLSX library is not yet approved.

**Output sheets:**
- Summary.
- Detail rows.
- Filters/metadata.

### Task 6.2: Add PDF export

**Objective:** Provide archive-ready report.

**Output:**
- A4 landscape when possible.
- Header with madrasah name, title, period, generated time.
- Page numbers.

### Task 6.3: Add Telegram share option

**Objective:** Send generated image/PDF to configured Telegram home/channel if approved.

**Guard:**
- Only admin/panitia.
- Confirmation modal before sending.
- Audit log: actor, report, filters, destination.

---

## Sprint 7 — Audit, Permissions, and Effective Access

### Task 7.1: Add report access audit

**Objective:** Track sensitive report access/export.

**Data to log:**
- Actor user.
- Report type.
- Filters.
- Export format.
- Timestamp.
- Destination if shared.

### Task 7.2: Add effective access copy

**Objective:** Explain why a user sees full vs own-only reports.

**UI text examples:**

```text
Anda melihat semua data karena memiliki izin bank_soal.analytics.
```

```text
Anda melihat data milik sendiri. Hubungi admin jika ditugaskan sebagai panitia/reviewer.
```

### Task 7.3: Tighten access after impact preview

**Objective:** Avoid overbroad report access while not breaking current users.

**Steps:**
1. Produce impact report of users with current access.
2. Review with user/admin.
3. Only then adjust role permissions if needed.

---

## Deployment Plan

Do not deploy until user explicitly approves.

When approved:

```bash
# Backend validation
cd /home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-banksoal-report-center ./cmd/api

# Frontend validation
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
git diff --check
```

Deploy order if backend routes changed:

```bash
cd services/core-api
cp -p bin/api bin/api.backup-banksoal-report-center-$(date +%Y%m%d-%H%M%S)
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
curl -fsS http://127.0.0.1:8080/health

cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run build
pm2 restart mtsn2kolut-web-admin --update-env
curl -I http://127.0.0.1:8021/bank-soal/laporan
```

## Acceptance Criteria

- Admin can open `/bank-soal/laporan` and see Input Soal report for `Bulan Ini` by default.
- Filters update data without requiring AI/manual SQL.
- PNG export matches the existing Telegram-ready table style.
- Data-quality flags highlight `system` seed rows and missing level/mapel rows.
- Guru without full report permission only sees own data.
- Additional tabs show useful operational reports: progress, revision, reviewer, readiness, honor/tugas.
- All relevant checks pass: Svelte check/build, sqlc generate, Go tests/build.
- No credentials or answer keys appear in report payloads or exports.

## Suggested Commit Sequence

1. `docs(bank-soal): plan report center`
2. `feat(bank-soal): add input report api`
3. `feat(bank-soal): add report center input tab`
4. `feat(bank-soal): export input report image`
5. `feat(bank-soal): add progress and revision reports`
6. `feat(bank-soal): add honor task report`
7. `feat(bank-soal): add report pdf excel exports`
8. `chore(bank-soal): audit report access`
