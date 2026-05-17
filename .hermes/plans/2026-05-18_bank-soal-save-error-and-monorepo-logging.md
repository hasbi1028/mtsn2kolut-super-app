# Bank Soal Save Error + Monorepo Logging Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task. Jangan deploy/restart PM2 sebelum approval eksplisit.

**Goal:** Menemukan dan memperbaiki bug gagal simpan soal CBT yang tampil sebagai `Pembuatan soal CBT tidak valid`, lalu membangun logger opsi B/structured logger untuk seluruh monorepo agar error serupa bisa dilacak dari frontend sampai backend.

**Architecture:** Implementasi dibagi dua jalur: (1) bugfix terarah untuk flow simpan soal Bank Soal dengan evidence dan regression test; (2) standar structured logging monorepo dengan request ID propagation, redaction global, server-side SvelteKit logger, Go slog wrapper, Pusaka Worker logger alignment, dan dokumentasi operasional. SvelteKit tetap BFF/proxy; Go core-api tetap pemilik DB dan validasi utama.

**Tech Stack:** Go core-api (`slog`, `net/http`, sqlc, PostgreSQL/pgx), SvelteKit web-admin (server routes/proxy/hooks), Node Pusaka Worker (Winston), PM2 logs, Telegram/CLI operational workflow.

---

## Non-Negotiable Safety Rules

- Tidak ada credential/API key/token/password/connection string dalam log, commit, atau laporan.
- Jangan log full cookie, authorization header, JWT, session token, atau DB URL.
- Jangan log full isi soal atau kunci jawaban secara terbuka; gunakan summary non-sensitif: panjang teks, jumlah opsi, tipe soal, subject/event id, target level, difficulty.
- SvelteKit tidak boleh akses DB langsung; hanya proxy/BFF ke Go API.
- `safeClientMessage` tetap menjaga pesan aman ke client; detail teknis hanya masuk log internal.
- Jangan deploy/restart PM2 sampai user menyetujui.
- Backup wajib sebelum migration production; rencana ini untuk logging tidak wajib migration kecuali nanti memilih error dashboard DB.
- Commit terpisah per milestone.

---

## Acceptance Criteria

### Bug Simpan Soal

- Saat `POST /api/bank-soal/questions` gagal, log internal menunjukkan penyebab spesifik:
  - validation error field apa, atau
  - PostgreSQL `sqlstate`, `constraint`, `table`, `column`, dan sanitized message.
- Root cause error user ditemukan dengan evidence, bukan tebakan.
- Ada regression test untuk bug yang ditemukan.
- Jika bug karena `event_id`/`subject_id` stale atau invalid, backend memberi validasi domain yang jelas di log dan tetap pesan client aman.
- Setelah fix, simpan soal normal berhasil.

### Logger Opsi B — Structured Logger Monorepo

- Semua request web-admin → core-api membawa `X-Request-ID`.
- Core API memakai logger wrapper konsisten dengan fields standar.
- SvelteKit server/proxy punya logger terstruktur.
- Pusaka Worker Winston diselaraskan field dan redaction-nya.
- Redaction global tersedia dan dites.
- Log bisa dicari dengan satu `request_id` dari frontend proxy sampai backend.
- Validasi lulus:
  - `npm --prefix apps/web-admin run check`
  - `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml`
  - `go test ./internal/handler ./internal/service ./internal/repository/postgres`
  - `go build -o /tmp/core-api-logging ./cmd/api`

---

# Phase 0 — Baseline Audit Tanpa Perubahan Behavior

## Task 0.1: Snapshot status repo dan service

**Objective:** Pastikan titik awal bersih sebelum patch.

**Files:** Tidak ada perubahan.

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git status --short
git branch --show-current
git log --oneline -5
pm2 list
curl -fsS http://127.0.0.1:8080/health
curl -fsSI http://127.0.0.1:8021/ | head
```

**Expected:** Working tree bersih atau hanya file plan ini; core-api dan web-admin online.

---

## Task 0.2: Catat flow simpan soal saat ini

**Objective:** Dokumentasikan route dan file yang terlibat agar patch tidak melebar.

**Files to inspect:**

- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `apps/web-admin/src/lib/server/cbt-backend-proxy/questions/+server.ts`
- `apps/web-admin/src/routes/api/bank-soal/questions/+server.ts`
- `services/core-api/internal/handler/cbt_question_authoring.go`
- `services/core-api/internal/handler/cbt_question.go`
- `services/core-api/internal/service/cbt_question.go`
- `services/core-api/internal/service/cbt_question_authoring.go`
- `services/core-api/internal/handler/errors.go`

**Verification:** Buat catatan singkat di Implementation Notes bawah plan ini.

---

# Phase 1 — Diagnostic Logging Khusus Bank Soal Dulu

Tujuan phase ini: bukan langsung fix. Kita tambahkan instrumentasi aman agar percobaan simpan berikutnya menunjukkan field/constraint penyebab.

## Task 1.1: Tambah request ID utility di web-admin server

**Objective:** Web-admin dapat membuat/meneruskan request ID untuk proxy request.

**Files:**

- Create: `apps/web-admin/src/lib/server/request-id.ts`

**Implementation sketch:**

```ts
import { randomUUID } from 'node:crypto';

export const REQUEST_ID_HEADER = 'x-request-id';

export function getOrCreateRequestId(headers: Headers): string {
  const existing = headers.get(REQUEST_ID_HEADER) || headers.get('x-correlation-id');
  return existing && existing.trim() ? existing.trim() : `req_${randomUUID()}`;
}
```

**Test/verification:**

```bash
npm --prefix apps/web-admin run check
```

---

## Task 1.2: Tambah server logger web-admin dengan redaction dasar

**Objective:** SvelteKit server route/proxy punya logger JSON terstruktur.

**Files:**

- Create: `apps/web-admin/src/lib/server/logger.ts`

**Fields standar:**

- `time`
- `level`
- `service: web-admin`
- `request_id`
- `route`
- `method`
- `status`
- `duration_ms`
- `event`
- `error`
- `metadata`

**Redaction keys:**

- `authorization`
- `cookie`
- `set-cookie`
- `password`
- `token`
- `secret`
- `api_key`
- `database_url`
- `connection_string`

**Implementation sketch:**

```ts
type LogLevel = 'debug' | 'info' | 'warn' | 'error';

type LogRecord = {
  level: LogLevel;
  service: 'web-admin';
  event: string;
  request_id?: string;
  route?: string;
  method?: string;
  status?: number;
  duration_ms?: number;
  error?: string;
  metadata?: Record<string, unknown>;
};

const REDACT_KEYS = /(authorization|cookie|password|token|secret|api[_-]?key|database[_-]?url|connection[_-]?string)/i;

function redact(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(redact);
  if (value && typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([key, val]) => [
        key,
        REDACT_KEYS.test(key) ? '[REDACTED]' : redact(val)
      ])
    );
  }
  return value;
}

export function log(record: LogRecord) {
  const safe = redact({
    time: new Date().toISOString(),
    ...record,
    service: 'web-admin'
  });
  const line = JSON.stringify(safe);
  if (record.level === 'error') console.error(line);
  else if (record.level === 'warn') console.warn(line);
  else console.log(line);
}
```

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

---

## Task 1.3: Instrument proxy Bank Soal questions

**Objective:** Proxy mencatat upstream 400/500 dengan request ID dan summary payload aman.

**Files:**

- Modify: `apps/web-admin/src/lib/server/cbt-backend-proxy/questions/+server.ts`

**Payload summary aman:**

- `has_subject_id`
- `has_event_id`
- `question_type`
- `target_level` / `grade_level`
- `difficulty`
- `workflow_status`
- `options_count`
- `stem_length`
- `question_text_length`
- `authoring_mode`

**Do not log:**

- full `answer_key`
- full `question_text`
- full stem HTML
- cookies/token

**Required behavior:**

- Tambahkan header `X-Request-ID` saat proxy ke core-api.
- Jika upstream gagal, log `bank_soal_questions_upstream_error`.
- Response ke client tetap aman dan kompatibel.

**Verification:**

```bash
npm --prefix apps/web-admin run check
```

---

## Task 1.4: Tambah Go logging package untuk core-api

**Objective:** Buat wrapper kecil di Go agar log fields dan redaction konsisten.

**Files:**

- Create: `services/core-api/internal/platform/logging/logging.go`
- Test: `services/core-api/internal/platform/logging/logging_test.go`

**Functions:**

- `RequestIDFromHeader(r *http.Request) string`
- `WithRequestID(ctx context.Context, id string) context.Context`
- `RequestID(ctx context.Context) string`
- `RedactAttrs(attrs ...slog.Attr) []slog.Attr`
- `ErrorAttrs(err error) []slog.Attr`

**Test cases:**

- redacts `authorization`, `password`, `token`, `secret`, `database_url`
- preserves safe keys like `subject_id`, `event_id`, `route`
- request ID roundtrip via context

**Commands:**

```bash
cd services/core-api
go test ./internal/platform/logging
```

---

## Task 1.5: Request ID middleware di core-api

**Objective:** Core API menerima/membuat request ID dan menaruhnya di context + response header.

**Files:**

- Modify: `services/core-api/cmd/api/main.go` atau lokasi middleware existing jika ada.
- Test: existing handler/middleware test jika tersedia; jika belum, minimal compile + integration check.

**Behavior:**

- Baca `X-Request-ID` dari request.
- Jika kosong, generate `req_<uuid>`.
- Set `X-Request-ID` di response.
- Simpan di `context.Context`.
- Request log existing menambahkan field `request_id`.

**Commands:**

```bash
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-request-id ./cmd/api
```

---

## Task 1.6: Tambah PostgreSQL error classifier

**Objective:** Error DB bisa dicatat detail teknisnya di log internal tanpa bocor ke client.

**Files:**

- Create: `services/core-api/internal/platform/logging/pgerror.go`
- Test: `services/core-api/internal/platform/logging/pgerror_test.go`

**Extract fields:**

- `sqlstate`
- `constraint`
- `table`
- `column`
- `severity`
- sanitized message

**Implementation note:**

Gunakan `errors.As(err, *pgconn.PgError)` jika dependency sudah tersedia melalui pgx.

**Test cases:**

- fake/constructed `pgconn.PgError` returns correct attrs.
- non-PgError returns generic error kind.

**Commands:**

```bash
cd services/core-api
go test ./internal/platform/logging
```

---

## Task 1.7: Instrument handler/service simpan soal

**Objective:** Saat create/update soal gagal, log menunjukkan penyebab spesifik.

**Files:**

- Modify: `services/core-api/internal/handler/cbt_question_authoring.go`
- Modify: `services/core-api/internal/service/cbt_question_authoring.go`
- Modify if needed: `services/core-api/internal/service/cbt_question.go`

**Log events:**

- `cbt_question_create_request_received`
- `cbt_question_create_validation_failed`
- `cbt_question_create_db_failed`
- `cbt_question_create_success`

**Fields aman:**

- `request_id`
- `user_id`/`username` jika tersedia
- `subject_id`
- `event_id_present`
- `event_id`
- `question_type`
- `target_level`
- `difficulty`
- `workflow_status`
- `options_count`
- `stem_length`
- `question_text_length`
- `authoring_mode`
- `sqlstate`
- `constraint`
- `table`
- `column`

**Do not log:**

- full stem
- full option text
- answer key value
- explanation full text

**Verification:**

```bash
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-bank-soal-diagnostic ./cmd/api
```

---

# Phase 2 — Reproduce Error dan Temukan Root Cause

## Task 2.1: Deploy diagnostic logger hanya setelah approval

**Objective:** Menjalankan logger diagnostik di production agar percobaan user berikutnya menghasilkan evidence.

**Requires approval:** Ya. User harus menyetujui deploy/restart PM2.

**Order:**

1. Build web-admin.
2. Build core-api.
3. Restart PM2 service terkait.
4. Health check.
5. User coba simpan soal sekali.
6. Ambil log dengan request ID.

**Commands, after approval only:**

```bash
npm --prefix apps/web-admin run build
cd services/core-api && go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
pm2 restart mtsn2kolut-web-admin --update-env
curl -fsS http://127.0.0.1:8080/health
curl -fsSI http://127.0.0.1:8021/ | head
```

---

## Task 2.2: Capture one failing request log

**Objective:** Ambil evidence dari satu percobaan simpan soal yang gagal.

**Commands:**

```bash
pm2 logs mtsn2kolut-web-admin --lines 150 --nostream
pm2 logs mtsn2kolut-core-api --lines 200 --nostream
```

**Expected:** Ada request ID sama di web-admin dan core-api.

**Decision tree:**

- Jika `constraint=cbt_questions_event_id_fkey`: bug terkait `event_id` stale/invalid.
- Jika `constraint=cbt_questions_subject_id_fkey`: bug terkait `subject_id` stale/invalid.
- Jika `constraint=chk_cbt_questions_target_level`: bug target level/grade mapping.
- Jika `constraint=chk_cbt_questions_workflow_status`: bug workflow_status payload.
- Jika `sqlstate=23505`: bug unique/generated code/version group.
- Jika service validation failed: field payload invalid sebelum DB.

---

# Phase 3 — Patch Bug Simpan Soal Berdasarkan Evidence

Bugfix hanya dilakukan setelah Phase 2 mengonfirmasi root cause.

## Task 3.1A: Jika root cause `event_id` stale/invalid

**Objective:** Backend menolak event invalid dengan error domain yang jelas di log, dan frontend tidak mengirim event stale.

**Files likely:**

- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Modify: `services/core-api/internal/service/cbt_question_authoring.go`
- Test: `services/core-api/internal/service/cbt_question_code_test.go` atau test baru `cbt_question_authoring_validation_test.go`

**Backend behavior:**

- Jika `event_id` dikirim, validasi event exists sebelum insert.
- Jika event tidak ada atau tidak usable, return domain validation error, log `invalid_event_id`.
- Jangan biarkan sampai FK constraint generic.

**Frontend behavior:**

- Saat metadata mode beruntun menyimpan `selectedEventId`, validasi ulang terhadap daftar event aktif/current.
- Jika event tidak ada, clear `selectedEventId` dan tampilkan pesan untuk pilih ulang kegiatan.

**Regression test:**

- Input dengan event UUID valid format tapi tidak ada di DB/store → service returns invalid event domain error.

---

## Task 3.1B: Jika root cause `subject_id` stale/invalid

**Objective:** Backend memvalidasi subject exists dan frontend membersihkan metadata subject stale.

**Files likely:**

- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Modify: `services/core-api/internal/service/cbt_question_authoring.go`
- Test: `services/core-api/internal/service/cbt_question_authoring_validation_test.go`

**Backend behavior:**

- Validasi `subject_id` exists sebelum insert.
- Jika tidak ada, log `invalid_subject_id`.
- Client tetap menerima pesan aman.

**Frontend behavior:**

- Pada load metadata beruntun, jika `subject_id` tidak ada di option list terbaru, clear metadata dan minta pilih mapel ulang.

**Regression test:**

- Subject UUID valid format tapi tidak ada → domain error, bukan raw FK.

---

## Task 3.1C: Jika root cause target level / workflow / difficulty

**Objective:** Samakan enum frontend dan backend.

**Files likely:**

- Modify: `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- Modify: `services/core-api/internal/service/cbt_question_authoring.go`
- Test: service validation tests.

**Fix examples:**

- Normalize target level ke `VII|VIII|IX`.
- Normalize difficulty ke `easy|medium|hard`.
- Normalize workflow status ke enum DB.

**Regression test:**

- Payload dari UI saat ini harus lolos normalization dan create params.

---

## Task 3.2: Manual smoke test simpan soal

**Objective:** Pastikan bug benar-benar selesai.

**Manual flow:**

1. Buka Bank Soal → Buat Soal.
2. Pilih mapel, tingkat, kesulitan, tipe PG.
3. Isi stem dan opsi A-D.
4. Simpan.
5. Pastikan sukses.
6. Coba mode Buat Beruntun.
7. Coba dengan dan tanpa `Khusus kegiatan ini`.

**Log expectation:**

- Success event tercatat dengan request ID.
- Tidak ada secret/full soal/kunci jawaban di log.

---

# Phase 4 — Logger Opsi B untuk Seluruh Monorepo

Ini implementasi standardisasi setelah bug flow utama aman.

## Task 4.1: Standar schema log monorepo

**Objective:** Dokumen schema agar frontend/backend/worker konsisten.

**Files:**

- Create: `docs/contracts/monorepo-logging.md`

**Schema fields:**

```json
{
  "time": "2026-05-18T06:39:12+08:00",
  "level": "info|warn|error|debug",
  "service": "core-api|web-admin|pusaka-worker|mobile-api",
  "env": "production|development|test",
  "request_id": "req_xxx",
  "trace_id": "optional",
  "user_id": "optional",
  "username": "optional",
  "role": "optional",
  "route": "optional",
  "method": "optional",
  "status": 200,
  "duration_ms": 12,
  "event": "domain_event_name",
  "module": "bank-soal|cbt|academic|pusaka|auth",
  "error_kind": "validation|postgres_constraint|upstream|internal",
  "error": "sanitized message",
  "metadata": {}
}
```

**Redaction policy:** Document mandatory redaction list.

---

## Task 4.2: Core API logger rollout

**Objective:** Replace ad-hoc `slog` calls with helper gradually.

**Files likely:**

- Modify: `services/core-api/cmd/api/main.go`
- Create/modify: `services/core-api/internal/platform/logging/*`
- Gradual modify: critical handlers first:
  - auth
  - bank soal
  - CBT event/session/package
  - academic
  - backup center

**Scope limit for first commit:**

- Middleware + Bank Soal + common error handler only.
- Jangan mass rewrite semua file sekaligus.

**Verification:**

```bash
cd services/core-api
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-logging ./cmd/api
```

---

## Task 4.3: SvelteKit logger rollout

**Objective:** Semua server routes/proxy penting memakai logger dan request ID.

**Files likely:**

- Create: `apps/web-admin/src/lib/server/logger.ts`
- Create: `apps/web-admin/src/lib/server/request-id.ts`
- Modify: `apps/web-admin/src/hooks.server.ts` jika ada/layak.
- Modify priority routes:
  - `apps/web-admin/src/lib/server/cbt-backend-proxy/questions/+server.ts`
  - academic BFF routes
  - CBT package/event/session BFF routes
  - auth/session routes

**Verification:**

```bash
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

---

## Task 4.4: Pusaka Worker Winston alignment

**Objective:** Worker log memakai schema sama dan redaction sama.

**Files likely:**

- Search first: `services/pusaka-worker/**`
- Modify logger config file if exists.
- If no central logger exists, create one:
  - `services/pusaka-worker/src/logger.ts` or existing path equivalent.

**Fields to add:**

- `service: pusaka-worker`
- `job_id`
- `request_id` if job triggered by API
- `username` only if safe and relevant
- `event`
- `duration_ms`

**Redaction:**

- PUSAKA username/password/token/cookie must be `[REDACTED]`.

**Verification:**

Use package script available in `services/pusaka-worker/package.json`; if none, at minimum TypeScript/build check.

---

## Task 4.5: Redaction tests across layers

**Objective:** Prove secrets cannot leak in logs.

**Backend Go tests:**

- `authorization=Bearer abc` → `[REDACTED]`
- `password=abc` → `[REDACTED]`
- `database_url=postgres://...` → `[REDACTED]`

**Web-admin tests/check:**

- If test framework exists, add unit test.
- If no unit test infra, include simple script/test function or rely on TS check plus manual node snippet.

**Pusaka Worker tests:**

- Same redaction cases where test infra exists.

---

## Task 4.6: Operational log commands doc

**Objective:** Admin/operator bisa mencari error dengan request ID.

**Files:**

- Create: `docs/ops/logging-runbook.md`

**Include commands:**

```bash
pm2 logs mtsn2kolut-web-admin --lines 200 --nostream | grep 'req_'
pm2 logs mtsn2kolut-core-api --lines 300 --nostream | grep 'request_id'
pm2 logs mtsn2kolut-core-api --lines 500 --nostream | grep 'cbt_question_create_db_failed'
```

**Include examples:**

- Foreign key error event.
- Validation error event.
- Upstream proxy error event.

---

# Phase 5 — Validation, Commit, and Rollout

## Task 5.1: Full validation before commit

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-logging ./cmd/api
```

**Expected:** All PASS.

---

## Task 5.2: Review diff for secret leaks and unrelated files

**Commands:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
git diff --stat
git diff --name-only
git diff --check
git diff | grep -Ei 'password|token|secret|database_url|postgres://' || true
```

**Expected:** No secret values. Only intentional files changed.

---

## Task 5.3: Commit milestones

**Recommended commits:**

```bash
git add .hermes/plans/2026-05-18_bank-soal-save-error-and-monorepo-logging.md
git commit -m "docs: plan bank soal save fix and monorepo logging"

# after diagnostic logging
git add apps/web-admin/src/lib/server services/core-api/internal/platform/logging ...
git commit -m "feat(logging): add request scoped diagnostics for bank soal"

# after bugfix
git add relevant/files
git commit -m "fix(bank-soal): resolve question save validation failure"

# after monorepo rollout
git add docs/contracts/monorepo-logging.md docs/ops/logging-runbook.md relevant/files
git commit -m "feat(logging): standardize structured logs across monorepo"
```

---

## Task 5.4: Production rollout after approval

**Requires explicit approval:** Ya.

**Order:**

1. Confirm no DB migration needed for logger phase.
2. Build backend and frontend.
3. Backup current binaries if replacing directly.
4. Restart core-api.
5. Restart web-admin.
6. Health check.
7. Smoke test Bank Soal save.
8. Capture request ID log sample.

**Commands after approval only:**

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app
npm --prefix apps/web-admin run build
cd services/core-api
cp -p bin/api bin/api.backup-logging-$(date +%Y%m%d-%H%M%S)
go build -o bin/api ./cmd/api
pm2 restart mtsn2kolut-core-api --update-env
pm2 restart mtsn2kolut-web-admin --update-env
curl -fsS http://127.0.0.1:8080/health
curl -fsSI http://127.0.0.1:8021/ | head
```

---

# Recommended Implementation Order

Saya rekomendasikan urutan ini:

1. Commit plan ini.
2. Implement Phase 1 diagnostic logger Bank Soal.
3. Jalankan check/build/test lokal.
4. Minta approval deploy diagnostic.
5. User coba simpan soal sekali.
6. Ambil log request ID.
7. Patch bug berdasarkan evidence.
8. Test + deploy bugfix setelah approval.
9. Lanjutkan logger Opsi B secara monorepo bertahap.

---

# Implementation Notes

Update bagian ini saat eksekusi:

- Current git branch:
- Initial git status:
- Diagnostic deploy approved by:
- Failing request ID:
- Confirmed root cause:
- Bugfix commit:
- Logging rollout commit:
- Production deploy timestamp:
- Smoke test result:

