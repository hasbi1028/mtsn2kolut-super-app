# PUSAKA Attendance Telegram Report Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build an admin-configurable feature that renders a clean image report for the PUSAKA daftar hadir ringkas and sends it to the user's Telegram on a manual or scheduled basis.

**Architecture:** Use Option B: render a dedicated report image from attendance data, not a browser screenshot. SvelteKit remains a thin BFF/proxy. Go core-api owns data retrieval, report generation, Telegram delivery, schedule config, job execution, and audit logs.

**Tech Stack:** SvelteKit web-admin, Go core-api, PostgreSQL/sqlc migrations, Telegram Bot API, existing scheduler/worker patterns, image rendering via server-side HTML/SVG-to-PNG or Go image template.

---

## Scope

### Included
- Admin settings for Telegram attendance report schedule.
- Manual "Kirim Telegram" button on daftar hadir ringkas.
- Rendered PNG/JPG report image with caption statistics.
- Scheduled automatic send based on admin-defined time.
- Delivery audit log.
- Role-gated access.

### Deferred
- Multi-recipient/group/channel management beyond initial configured targets.
- WhatsApp delivery.
- Excel/PDF export unless added in a later sprint.
- Sending raw frontend screenshots.

---

## Recommended Delivery Shape

Telegram message:

```text
📋 Daftar Hadir Ringkas
MTsN 2 Kolaka Utara
Tanggal: 16 Mei 2026

Total Pegawai: 32
Sudah Masuk: 24
Belum Masuk: 8
Terlambat: 3
Sudah Pulang: 20

Dikirim otomatis dari Sistem MTsN 2 Kolaka Utara.
```

Attachment: generated image containing a clean report table:
- Header madrasah
- Date/time
- Summary cards
- Table: No, Nama Pegawai, Masuk, Pulang, Status
- Footer: generated timestamp

---

## Data Model

### Task 1: Add Telegram report settings migration

**Objective:** Store admin-configurable schedule and target settings.

**Files:**
- Create: `services/core-api/db/migrations/0XX_pusaka_attendance_telegram_reports.sql`
- Modify: `services/core-api/db/queries/*.sql` as needed

**Tables:**

```sql
CREATE TABLE IF NOT EXISTS pusaka_attendance_telegram_settings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  is_enabled boolean NOT NULL DEFAULT false,
  send_time time NOT NULL DEFAULT '17:00',
  timezone text NOT NULL DEFAULT 'Asia/Makassar',
  target_chat_id text NOT NULL,
  include_caption boolean NOT NULL DEFAULT true,
  include_image boolean NOT NULL DEFAULT true,
  report_mode text NOT NULL DEFAULT 'ringkas',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS pusaka_attendance_telegram_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_date date NOT NULL,
  target_chat_id text NOT NULL,
  send_mode text NOT NULL, -- manual | scheduled
  status text NOT NULL, -- success | failed
  telegram_message_id text,
  error_message text,
  requested_by uuid,
  sent_at timestamptz NOT NULL DEFAULT now()
);
```

**Notes:**
- `target_chat_id` should not be exposed broadly in UI; mask when displayed.
- If RBAC uses permission seeds, add an admin permission in the same migration.

**Verification:**

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/repository/postgres
```

---

## Backend API

### Task 2: Add report settings queries and service

**Objective:** Read/update schedule settings safely.

**Files:**
- Modify/Create: `services/core-api/db/queries/pusaka_attendance_telegram.sql`
- Modify: `services/core-api/internal/repository/postgres/*.sql.go` via sqlc
- Create/Modify: `services/core-api/internal/service/pusaka_attendance_telegram.go`

**Service methods:**
- `GetAttendanceTelegramSettings(ctx)`
- `UpdateAttendanceTelegramSettings(ctx, input)`
- `SendAttendanceTelegramReport(ctx, input)`
- `RunDueAttendanceTelegramReports(ctx, now)`

**Validation:**
- `send_time` required and valid.
- `timezone` allowlist: default `Asia/Makassar`.
- `target_chat_id` required when enabled.
- Only authorized roles can update/send.

---

### Task 3: Add report data query

**Objective:** Build one backend use case that returns the exact rows/statistics for the report.

**Files:**
- Modify/Create: attendance/PUSAKA query file under `services/core-api/db/queries/`
- Modify/Create: service report DTOs

**Output DTO:**

```go
type AttendanceTelegramReport struct {
  Date string
  GeneratedAt time.Time
  TotalEmployees int
  CheckedIn int
  NotCheckedIn int
  Late int
  CheckedOut int
  Rows []AttendanceTelegramReportRow
}

type AttendanceTelegramReportRow struct {
  No int
  EmployeeName string
  CheckIn string
  CheckOut string
  Status string
}
```

**Important:** Reuse the same data source as halaman daftar hadir ringkas so UI and Telegram match.

---

### Task 4: Implement image renderer

**Objective:** Generate a stable PNG/JPG report image from report DTO.

**Recommended approach:** Generate an HTML/SVG template and convert to PNG server-side.

**Options:**
1. Go-only SVG renderer: simpler, stable, no browser dependency.
2. HTML + Playwright/Chromium screenshot: prettier, but heavier.
3. Go image canvas: stable but more manual layout work.

**Recommendation:** Start with SVG-to-PNG or Go image canvas for reliability.

**Files:**
- Create: `services/core-api/internal/reporting/attendance_telegram_renderer.go`
- Create: `services/core-api/internal/reporting/attendance_telegram_renderer_test.go`

**Acceptance:**
- Produces PNG under Telegram size limits.
- Handles 32+ employees.
- Splits into multiple images if rows exceed safe height, if necessary.
- Masks internal IDs; shows human names only.

---

### Task 5: Implement Telegram sender

**Objective:** Send image + caption through Telegram Bot API from backend only.

**Files:**
- Create: `services/core-api/internal/integration/telegram/client.go`
- Modify: config/env loader to support Telegram token

**Environment:**
- `TELEGRAM_BOT_TOKEN=[REDACTED]`
- Optional default target: `PUSAKA_ATTENDANCE_TELEGRAM_CHAT_ID=[REDACTED]`

**Security:**
- Never expose bot token to frontend.
- Do not log bot token.
- Log only masked target/chat id.

**API call:**
- `sendPhoto` with `caption`.

---

### Task 6: Add core-api HTTP endpoints

**Objective:** Expose admin-controlled endpoints for BFF.

**Files:**
- Modify: `services/core-api/internal/handler/...`
- Modify: `services/core-api/cmd/api/main.go`
- Add tests under `services/core-api/internal/handler/`

**Endpoints:**

```text
GET  /api/pusaka/attendance-telegram/settings
PUT  /api/pusaka/attendance-telegram/settings
POST /api/pusaka/attendance-telegram/send
GET  /api/pusaka/attendance-telegram/logs
```

**Manual send body:**

```json
{
  "date": "2026-05-16",
  "target_chat_id": "optional override",
  "include_caption": true,
  "include_image": true
}
```

**Verification:**

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-pusaka-telegram-report ./cmd/api
```

---

## Scheduler

### Task 7: Add scheduled job runner

**Objective:** Automatically send reports at admin-defined time.

**Files:**
- Modify existing backend scheduler files, or create `services/core-api/internal/scheduler/pusaka_attendance_telegram.go`

**Behavior:**
- Every minute or every 5 minutes, check whether current WITA time matches configured `send_time`.
- Send only once per `report_date` + `target_chat_id` + `scheduled`.
- Write success/failure log.
- If disabled, do nothing.

**Idempotency rule:**
- A scheduled report for the same date/target should not be sent twice unless admin manually sends.

---

## Web Admin UI

### Task 8: Add BFF proxy endpoints

**Objective:** Keep SvelteKit as BFF only.

**Files:**
- Create: `apps/web-admin/src/routes/api/pusaka/attendance-telegram/settings/+server.ts`
- Create: `apps/web-admin/src/routes/api/pusaka/attendance-telegram/send/+server.ts`
- Create: `apps/web-admin/src/routes/api/pusaka/attendance-telegram/logs/+server.ts`

**Rules:**
- Proxy to Go API only.
- No direct DB access.
- Preserve auth/session handling.

---

### Task 9: Add manual send button to Daftar Hadir Ringkas

**Objective:** Let admin send the selected date report immediately.

**Files:**
- Modify the actual daftar hadir/PUSAKA attendance route once located.

**UI:**
- Button: `Kirim Telegram`
- Confirmation modal:
  - Tanggal
  - Target Telegram masked
  - Include image/caption toggles if needed
- Toast on success/failure.

**Acceptance:**
- Admin can send today's or selected date's report.
- Button has loading state.
- Failed send displays useful error.

---

### Task 10: Add schedule settings page/section

**Objective:** Allow admin to configure when Telegram reports are sent.

**Recommended location:**
- PUSAKA settings page if existing, or a new section under PUSAKA: `Pengaturan Laporan Telegram`.

**Fields:**
- Aktifkan kirim otomatis: on/off
- Jam kirim: time input
- Zona waktu: WITA / `Asia/Makassar`
- Target Telegram chat id: masked input
- Format: caption + gambar
- Test kirim: button
- Last sent / delivery log summary

---

## Verification & QA

### Task 11: Test backend

```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-pusaka-telegram-report ./cmd/api
```

### Task 12: Test frontend

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

### Task 13: Manual smoke test

- Open daftar hadir ringkas.
- Select date.
- Click `Kirim Telegram`.
- Confirm Telegram receives caption + image.
- Check delivery log.
- Configure schedule time +2 minutes.
- Wait for scheduled send.
- Verify duplicate scheduled send does not happen.

---

## Deployment Notes

Deployment order:
1. Run DB migration.
2. Configure Telegram env vars on backend server.
3. Build/restart core-api.
4. Build/restart web-admin immediately after build.
5. Health check backend.
6. Manual Telegram test send.

Do not deploy until explicitly approved.

---

## Suggested Sprint Breakdown

### Sprint 1 — Manual Send MVP
- Migration settings/logs.
- Backend report DTO.
- Image renderer.
- Telegram sender.
- Manual send endpoint.
- BFF proxy.
- Button on daftar hadir ringkas.

### Sprint 2 — Admin Schedule
- Settings UI.
- Scheduler runner.
- Idempotency per date/target.
- Delivery logs.
- Test send from settings.

### Sprint 3 — Polish
- Multi-target support.
- PDF option.
- Failed-send retry.
- Report templates.
- More filters: belum hadir, terlambat, pulang cepat.

---

## Acceptance Criteria

- Admin can configure report send time.
- Admin can enable/disable scheduled sending.
- Admin can manually send selected date report.
- Telegram receives a clean image report and caption.
- Delivery is logged.
- Scheduled report is sent only once per date/target.
- Token is backend-only and never exposed to web-admin.
- `npm --prefix apps/web-admin run check` passes.
- `sqlc generate`, Go tests, and Go build pass.
