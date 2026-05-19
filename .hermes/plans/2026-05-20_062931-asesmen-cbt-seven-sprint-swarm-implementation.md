# Asesmen/CBT Seven-Sprint Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task with spec review and code-quality review after each task.

**Goal:** Implement all seven recommended Asesmen/CBT improvement sprints from the swarm review: formal UI language, high-impact dialogs, SOP gates, final archive, security hardening, mobile/proctoring event contracts, and proctor day-mode/reporting.

**Architecture:** Keep Go core-api as the only PostgreSQL owner; SvelteKit remains BFF/proxy/UI only. Roll out additively and audit-safe: migrations first, compatibility/read-only/report-only mode where needed, then enforcement gates after tests and explicit approval.

**Tech Stack:** SvelteKit 2/Svelte 5 runes, shadcn-svelte, sonner, Go/Chi/sqlc/pgx/PostgreSQL, Flutter mobile CBT.

---

## Sprint 0 — Safety Gate and Baseline

**Objective:** Freeze scope, capture baseline, and avoid breaking live CBT data.

**Files:**
- Read: `docs/reviews/2026-05-20_061102-asesmen-cbt-swarm-review.md`
- Read: `docs/contracts/asesmen-cbt-formal-sop.md`
- Read: `services/core-api/db/migrations/*cbt*`
- Read: `apps/web-admin/src/routes/asesmen/**`

**Tasks:**
1. Record `git status --short`, branch, and latest commit.
2. Run baseline checks:
   - `cd apps/web-admin && npm run check`
   - `cd services/core-api && bash -lc 'source ~/.profile 2>/dev/null || true; go test ./internal/handler ./internal/service -run "Cbt|Exam|Proctor" -count=1 -timeout 120s'`
3. Export counts/status distribution for CBT events, sessions, approvals, rooms, participants. Do not print credentials.
4. Confirm no deployment/restart until implementation patches pass and user approves.

**Acceptance:** Baseline report saved under `docs/reviews/` or included in sprint notes.

---

## Sprint 1 — Formal UI Language Cleanup

**Objective:** Remove technical/internal terms from Asesmen/CBT UI and standardize formal madrasah wording.

**Primary files:**
- `apps/web-admin/src/routes/asesmen/+page.svelte`
- `apps/web-admin/src/routes/asesmen/persiapan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- `apps/web-admin/src/routes/asesmen/hasil/+page.svelte`
- `apps/web-admin/src/routes/asesmen/paket/+page.svelte`
- `apps/web-admin/src/routes/asesmen/paket/new/+page.svelte`
- `apps/web-admin/src/routes/asesmen/paket/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/kegiatan/**/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/new/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/**/+page.svelte`
- `apps/web-admin/src/routes/api/exam/[...path]/+server.ts`
- `apps/mobile/lib/src/**`

**Copy map:**
- `Quality Gate Paket` → `Pemeriksaan Kesiapan Paket`
- `Mix policy` → `Pengaturan Pembagian Peserta`
- `Nilai ini dikirim ke backend` → `Pilihan ini digunakan sistem saat membagi peserta ke ruang ujian.`
- `backend` → `sistem` / `layanan ujian` / omit
- `Flutter` → `Aplikasi siswa`
- `Payload ujian terlalu besar` → `Data jawaban terlalu besar untuk dikirim. Ringkas jawaban atau minta bantuan pengawas.`
- `Hasil & BA` → `Hasil & Berita Acara`
- `Paket Builder` → `Penyusunan Paket`
- `proctoring` → `Pengawasan Ruang`
- `acknowledge` → `Tandai Sudah Diperiksa`
- `pending sync` → `Jawaban Belum Tersinkron`

**Tasks:**
1. Add a small glossary comment/reference if useful, but avoid overengineering.
2. Patch page copy in small batches: hub/results, package builder, session creation, proctoring/report, mobile strings.
3. Run a static scan for banned terms in operator-facing Svelte/mobile copy.
4. Run `npm run check`; if mobile changed and Flutter exists, run Flutter tests/analysis, otherwise document toolchain limitation.

**Acceptance:** No user-facing `Quality Gate`, `Mix policy`, raw `backend`, `Flutter` in operator UI except developer docs/comments.

---

## Sprint 2 — Formal Dialogs for High-Impact Actions

**Objective:** Replace native `window.confirm`/`window.prompt` with audit-safe dialogs and required reasons/notes.

**Primary files:**
- `apps/web-admin/src/routes/asesmen/paket/[id]/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- Optional shared component: `apps/web-admin/src/lib/components/asesmen/ProctorActionDialog.svelte`
- Optional shared component: `apps/web-admin/src/lib/components/asesmen/ConfirmActionDialog.svelte`

**Actions covered:**
- Clone/revisi paket.
- Lock/kunci paket.
- Unlock/buka kunci peserta.
- Acknowledge/tandai kejadian diperiksa.
- Browser Darurat / web fallback.
- Force submit peserta.

**Tasks:**
1. Create/reuse a Svelte dialog pattern with title, impact summary, reason textarea/input, cancel, and loading state.
2. Replace package clone prompt with structured dialog.
3. Replace package lock confirm with structured dialog and reason text sent to API where supported.
4. Replace session-level proctoring unlock/acknowledge prompts.
5. Replace room-level browser darurat prompt.
6. Upgrade force-submit UI to require explicit reason/notes, especially when pending sync exists.
7. Run `npm run check`.

**Acceptance:** Static scan under `apps/web-admin/src/routes/asesmen` returns zero `window.confirm` and zero `window.prompt` for Asesmen/CBT pages.

---

## Sprint 3 — SOP State Machine and Approval Gates

**Objective:** Convert SOP readiness from display-only into an enforceable, auditable workflow.

**Backend files:**
- Migration: `services/core-api/db/migrations/NNN_cbt_event_sop_state_machine.sql`
- Queries: `services/core-api/db/queries/cbt_events.sql` or new `cbt_event_sop.sql`
- Generated: `services/core-api/internal/repository/postgres/*` via sqlc
- Service: `services/core-api/internal/service/cbt_event.go`
- Service: `services/core-api/internal/service/cbt_approval.go`
- Handler: `services/core-api/internal/handler/cbt_event.go`
- Router: `services/core-api/cmd/api/main.go`
- Tests: `services/core-api/internal/service/*cbt_event*test.go`, `services/core-api/internal/handler/*cbt_event*test.go`

**Frontend files:**
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/+page.svelte`
- BFF proxy routes under `apps/web-admin/src/routes/api/asesmen/**`
- Route map tests under `apps/web-admin/src/lib/server/*test.ts`

**Model:**
- `draft`
- `question_authoring`
- `question_verification`
- `package_ready`
- `participants_rooms_ready`
- `tokens_cards_ready`
- `execution`
- `grading`
- `result_verification`
- `final_archive`
- optional `cancelled`

**Tasks:**
1. Add additive migration columns to `cbt_exam_events`: `sop_state`, `sop_state_updated_at`, `sop_state_updated_by`, `sop_state_note`.
2. Add audit table `cbt_event_sop_transitions`.
3. Add sqlc queries for get/update/list transition/gate snapshot.
4. Add service transition validator with allowed transition matrix and gate snapshot preconditions.
5. Initially implement report-only/soft enforcement if existing data is incomplete.
6. Add handler endpoints: get SOP state, transition SOP state, list transition history.
7. Add BFF aliases and route-contract tests.
8. Add UI SOP panel with current state, blockers, next action, and audit trail.
9. Run sqlc generate, Go tests, frontend tests/check.

**Acceptance:** Admin/panitia can advance Kegiatan Asesmen through formal SOP stages only when backend preconditions pass; every transition is audited.

---

## Sprint 4 — Final Archive and Formal Document Completeness

**Objective:** Make final archive safe, verifiable, and tied to formal evidence/documents.

**Backend files:**
- Migration: `services/core-api/db/migrations/NNN_cbt_event_archive_documents.sql`
- Queries: `services/core-api/db/queries/cbt_approvals.sql`, `cbt_events.sql`, optional `cbt_archive.sql`
- Service: `services/core-api/internal/service/cbt_approval.go`, `cbt_event.go`
- Handler: `services/core-api/internal/handler/cbt_approval.go`, `cbt_event.go`

**Frontend files:**
- `apps/web-admin/src/routes/asesmen/kegiatan/[id]/archive/+page.svelte`
- `apps/web-admin/src/routes/asesmen/hasil/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/report/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`

**Final archive preconditions:**
- All sessions finished/cancelled with valid reason.
- All participants have final status: submitted, absent, excused, no-show, force-submitted, etc.
- Manual/essay grading completed.
- Results verified.
- Proctor incidents acknowledged/reviewed.
- Room/session BA present or explicitly marked not required with reason.
- Final approval actor/note recorded.

**Tasks:**
1. Add archive document checklist model/queries if not already covered.
2. Add final archive precondition service that returns blockers/warnings.
3. Wire `final_archive` approval to hard gate after soft mode is validated.
4. Upgrade archive page to show document completeness and blocker list.
5. Upgrade `/asesmen/hasil` into post-exam work center: ungraded, unsubmitted, pending incidents, unverified results.
6. Add print/export links for formal archive pack.
7. Run Go and frontend checks.

**Acceptance:** Final archive cannot be approved until blockers are resolved or formally waived where allowed; page clearly shows why.

---

## Sprint 5 — Security Hardening

**Objective:** Harden exam proxy, body limits, token/answer-key redaction, and admin/proctor JSON body handling.

**Files:**
- `apps/web-admin/src/routes/api/exam/[...path]/+server.ts`
- `apps/web-admin/src/lib/server/api.ts`
- `services/core-api/internal/handler/exam.go`
- `services/core-api/internal/handler/request_origin.go`
- `services/core-api/internal/handler/cbt_question_serialize.go`
- `services/core-api/internal/handler/cbt_question_authoring.go`
- `services/core-api/internal/handler/cbt_session.go`
- Tests under `apps/web-admin/src/lib/server/*.test.ts`, `services/core-api/internal/handler/*test.go`

**Tasks:**
1. Stop forwarding client-supplied `x-forwarded-for` and `x-real-ip` through `/api/exam` BFF.
2. Replace `text.length` body limit with byte-count/stream limit; align route-specific limits: login 4KiB, event 16KiB, answer 64KiB, heartbeat/submit 1KiB.
3. Add BFF unit tests for multibyte body limit and stripped spoofed IP headers.
4. Add Go handler tests for 413 over-limit on all `/api/exam/*` JSON endpoints.
5. Add defense-in-depth redaction tests for answer keys in question list/detail and student exam payload.
6. Consider shared limited JSON decoder for high-impact admin/proctor handlers reachable directly on Go API.
7. Run `go test ./internal/handler ./internal/service` and `npm run check`.

**Acceptance:** Student and non-author/non-admin paths cannot receive answer keys; `/api/exam` BFF cannot be used to spoof client IP headers; body limits are byte-accurate.

---

## Sprint 6 — Mobile/Backend Event Contract and Audio Policy

**Objective:** Ensure Flutter anti-cheat/proctoring events map correctly to backend policy and dashboard alerts.

**Files:**
- `apps/mobile/lib/src/exam_events.dart`
- `apps/mobile/test/exam_events_test.dart`
- `services/core-api/internal/service/cbt_proctoring_policy.go`
- `services/core-api/internal/service/cbt_proctoring_policy_test.go`
- `services/core-api/internal/service/exam.go`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

**Tasks:**
1. Add backend contract tests for every Flutter event type/reason: `anti_cheat_violation`, `app_switch`, `screenshot_attempt`, focus/background/split/PiP, pending sync, technical events.
2. Ensure technical events have zero risk delta and do not auto-lock.
3. Add mapping for `anti_cheat_violation` based on reason if currently missing.
4. Add/verify telemetry sanitization tests for nested sensitive fields.
5. Update frontend audio alert logic to use backend-provided `audio_key`, `severity`, and `category`, not hardcoded raw event names.
6. Add cooldown per participant/category in UI if needed.
7. Run Go tests and mobile tests if Flutter toolchain is available.

**Acceptance:** All mobile event payloads accepted/classified as intended; technical events are not treated as cheating; audio follows backend policy.

---

## Sprint 7 — Proctor Day Mode and Final Reports/BA

**Objective:** Make CBT day-of operation simple for proctors and strengthen formal reports/BA.

**Files:**
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/minutes/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/proctoring/report/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/report/+page.svelte`
- Optional components:
  - `apps/web-admin/src/lib/components/asesmen/ProctorDayModePanel.svelte`
  - `apps/web-admin/src/lib/components/asesmen/ProctorActionDialog.svelte`
  - `apps/web-admin/src/lib/components/asesmen/IncidentSummaryCard.svelte`

**Day-mode priorities:**
- Butuh tindakan sekarang.
- Gangguan koneksi.
- Peserta terkunci.
- Belum submit mendekati akhir.
- Insiden belum diperiksa.
- Pending sinkronisasi.

**Report/BA fields:**
- daftar hadir/status peserta,
- status submit,
- technical incidents separated from suspected misconduct,
- proctor actions with actor/reason/note/time,
- reset/unlock/force-submit summary,
- room handover/signature section,
- results verification/final archive link.

**Tasks:**
1. Add day-mode toggle/default view for proctor/room pages.
2. Add summary priority cards and one-click filters.
3. Reuse Sprint 2 dialogs for all actions.
4. Upgrade report pages with formal sections and language.
5. Add print-friendly layout for BA/report.
6. Run `npm run check`, browser smoke if login path available.

**Acceptance:** Proctor can identify urgent participants in <=10 seconds; reports are formal enough for arsip/panitia review.

---

## Room Mixing / Randomization Sprintlet

**Objective:** Make room randomization rules explicit and enforce `mix_policy`, `assignment_mode`, and `allow_cross_grade`.

**Current finding:** Today `shuffleRooms()` randomizes all participants in one pool and fills rooms by capacity. It does not yet enforce `same_class`, `same_grade`, `mixed_scope`, `allow_cross_grade`, gender, or accommodation rules.

**Files:**
- `services/core-api/internal/service/cbt_session.go`
- `services/core-api/db/queries/cbt_sessions.sql`
- `services/core-api/internal/service/cbt_session_test.go`
- `services/core-api/internal/handler/cbt_session*.go`
- `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`
- `apps/web-admin/src/routes/asesmen/sesi/new/+page.svelte`
- room setup/proctoring pages under `apps/web-admin/src/routes/asesmen/sesi/[id]/**`

**Policy:**
- `same_class`: do not mix rombel in one room unless room capacity forces warning/manual override.
- `same_grade`: may mix 7A/7B/etc, but must not mix grade VII and VIII.
- `mixed_scope`: may mix across grade/rombel only for school/custom/special event; if cross-grade true require `is_special_event` and explicit UI warning.
- `random_balanced`: balance rooms by count.
- `random_by_gender`: balance by gender where student metadata exists.
- `random_by_accommodation`: place accommodations with constraints/warnings where metadata exists.

**Tasks:**
1. Extend `ListParticipantsByRoom` or add `ListParticipantsForRoomAssignment` to include class code, class level, gender, accommodation flags.
2. Add service tests for each policy: class-only, same-grade, mixed-scope, cross-grade blocked, cross-grade allowed special event.
3. Implement policy-aware assignment grouping.
4. Add readiness warning that surfaces mixed-grade rooms and policy violations before activation.
5. Update UI labels from `Mix policy` to `Pengaturan Pembagian Peserta` and explain consequences.
6. Run Go tests + frontend checks.

**Acceptance:** The system either intentionally mixes rooms according to selected policy, or blocks/warns when policy would be violated.

---

## Agent Swarm Execution Strategy

1. Parent controller keeps one authoritative working tree and runs final verification.
2. Use subagents per sprint or per non-overlapping file slice.
3. Do not let two agents edit the same Svelte page in parallel.
4. For backend migrations/sqlc, only one backend agent owns the migration/query/service chain per sprint.
5. Each implementation task gets:
   - implementer agent,
   - spec reviewer agent,
   - code-quality/security reviewer agent.
6. Parent runs final gates after every sprint:
   - `cd services/core-api && bash -lc 'source ~/.profile 2>/dev/null || true; go test ./internal/handler ./internal/service -count=1 -timeout 120s'`
   - `cd services/core-api && bash -lc 'source ~/.profile 2>/dev/null || true; go build -o bin/api ./cmd/api'`
   - `cd apps/web-admin && npm run check`
   - `cd apps/web-admin && rm -rf build && npm run build`
7. Commit each stable sprint separately.
8. Deploy order only after approval: backend code → migrations → backend restart/health → frontend build/restart → smoke.

---

## Verification Bundle

```bash
cd /home/servermtsn2kolut/mtsn2kolut-super-app

git status --short

cd services/core-api
bash -lc 'source ~/.profile 2>/dev/null || true; go test ./internal/handler ./internal/service -count=1 -timeout 120s'
bash -lc 'source ~/.profile 2>/dev/null || true; go build -o bin/api ./cmd/api'

cd ../../apps/web-admin
npm run check
rm -rf build && npm run build
```

## Deployment Reminder

Do not deploy/restart or run migrations during active exam windows. Get explicit user approval before production deploy/restart.
