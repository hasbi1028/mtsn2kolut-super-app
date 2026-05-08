# CBT Security/RBAC Hardening Evidence — Plan B1-B4

Status: implemented evidence hardening baseline on 2026-05-08. This document is a release-control guard for Plan B security work, not an ISO certification claim.

## Plan B1 — RBAC/API Surface Inventory

Core boundaries:
- Unauthenticated Web Admin pages must return `302`.
- Unauthenticated protected BFF/API routes must return `401`.
- public `/api/cbt/**` route tree remains forbidden in SvelteKit Web Admin; the focused smoke check verifies `/api/cbt/questions` is not public `200`. CBT naming may remain only in internal backend/proxy compatibility paths where the Go API already owns legacy CBT internals.
- Flutter student runtime remains `/api/exam/*` and is protected by exam token middleware, not Web Admin session cookies.
- Web Admin BFF routes use `/api/bank-soal/*` and `/api/asesmen/*`.
- PostgreSQL remains owned only by `services/core-api`.

| Area | Web Admin route | BFF route | Core API endpoint | Required access | Unauthenticated smoke | Export/evidence rule |
| --- | --- | --- | --- | --- | --- | --- |
| Bank Soal dashboard | `/bank-soal` | `/api/bank-soal/summary` | native `/api/bank-soal/summary` | `bank_soal.read` or allowed role fallback | page `302`, API `401` | no token/export payload |
| Bank Soal composer | `/bank-soal/tambah` | `/api/bank-soal/questions` | native Bank Soal question create/update | `bank_soal.create` | page `302`, API `401` | draft/export must not expose secrets |
| Bank Soal review | `/bank-soal/verifikasi` | `/api/bank-soal/questions/*/review` | native Bank Soal review/publish handlers | `bank_soal.review` / `bank_soal.publish` | page `302`, API `401` | reviewer notes only, no token |
| Import soal | `/bank-soal/impor` | `/api/bank-soal/import` | native Bank Soal import handlers | `bank_soal.import` | page `302`, API `401` | import preview no credentials |
| Asesmen overview | `/asesmen` | `/api/asesmen/events` | native assessment event handlers | `asesmen.read` | page `302`, API `401` | no participant token exposure |
| Session setup | `/asesmen/sesi` | `/api/asesmen/sessions` | native session handlers | `asesmen.session_manage` | page `302`, API `401` | generated tokens stay protected |
| Proctor room | `/asesmen/sesi/[id]/rooms/[rid]/proctoring` | `/api/asesmen/sessions/[id]/rooms/[rid]/proctoring` | native proctoring handlers | `asesmen.proctor` | page `302`, API `401` | evidence CSV is token-free |
| Proctor actions | same room page | reset/force-submit/flag participant BFF routes | native proctor participant action handlers | `asesmen.proctor` or stricter action permission | API `401` | action audit only |
| Results | `/asesmen/hasil` and session result pages | `/api/asesmen/sessions/[id]/results`, item-analysis, operational recap | native results/report handlers | `asesmen.results_read` / management permission | page `302`, API `401` | CSV/JSON outputs are redacted and CSV-safe |
| Mobile exam status | Flutter only | none | `/api/exam/status` | exam token middleware | API `401` without token | no Web Admin cookie auth |
| Mobile exam answer | Flutter only | none | `/api/exam/answer` | exam token middleware, participant/device/session scope | API `401` without token | answer body cap enforced |
| Mobile exam submit | Flutter only | none | `/api/exam/submit` | exam token middleware, participant/device/session scope | API `401` without token | submitted or closed sessions fail closed |

## Plan B2 — Protected Route Smoke Matrix

The focused smoke matrix lives in `docs/cbt-security-smoke-matrix.md` and `bash deploy/scripts/health-check.sh cbt-security`.

## Plan B3 — Evidence Export Redaction and CSV Safety

Exam tokens, passwords, bearer tokens, API keys, and raw secrets must be redacted from evidence exports. CSV injection cells starting with `=`, `+`, `-`, `@`, tab, CR, or LF must be escaped. Proctor evidence helpers sanitize generated rows before download/export.

## Plan B4 — Backend Exam Token/Device Boundary Review

`/api/exam/*` is token-scoped, not session-cookie-scoped. Existing backend tests and current validation verify these boundaries:
- missing participant token context returns `401`;
- device mismatch returns `409`;
- device fingerprint required returns `400`;
- malformed/trailing JSON returns `400`;
- oversized login/answer/submit/event bodies are blocked before service calls;
- answer/submit on not-started, already-submitted, or window-closed sessions fails closed;
- submitted or closed sessions fail closed;
- question IDs outside participant exam scope are rejected.

Final production acceptance remains blocked until manual Android device matrix, operator rehearsal, and explicit sign-off are complete.
