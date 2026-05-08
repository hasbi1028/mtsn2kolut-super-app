# CBT Proposal 100% Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Menyelesaikan integrasi proposal “Sistem CBT MTsN 2 Kolaka Utara” sampai bisa diklaim 100% secara produk, runtime, evidence operasional, dan sign-off resmi — tetap dalam arsitektur monorepo yang sudah disetujui.

**Architecture:** Proposal tidak diimplementasikan mentah-mentah sebagai PocketBase/SQLite/Alpine. Semua fitur masuk ke arsitektur resmi: Web Admin SvelteKit untuk Bank Soal/Asesmen/Pengawasan/Hasil, Core API Go + PostgreSQL sebagai source of truth, Flutter sebagai aplikasi ujian siswa/BYOD anti-cheat, dan PUSAKA worker tetap terpisah dari runtime CBT.

**Tech Stack:** SvelteKit 2/Svelte 5, Go/Chi/sqlc/PostgreSQL, Flutter, PM2, existing `/api/exam/*`, `/bank-soal/*`, `/asesmen/*`, evidence scripts under `deploy/scripts/`, docs/tests guard.

---

## Current context

Baseline yang sudah selesai:

- Phase 0–15 proposal alignment/readiness sudah committed.
- Final evidence/gap audit framework sudah committed.
- Flutter runtime follow-up sudah menambah:
  - `ordering`
  - `true_false`
  - `agree_disagree`
  - guard aman untuk `hotspot`, `upload_answer`, `file_upload`
- Live health terakhir PASS.
- Perubahan Flutter terbaru belum otomatis dipakai siswa sampai APK baru dibuild dan dirilis.

Target “100%” harus dipahami sebagai:

1. Semua fitur proposal yang relevan sudah diadaptasi ke monorepo.
2. Fitur proposal yang tidak realistis untuk BYOD diberi status “adapted/out-of-scope” dengan bukti dan alasan resmi.
3. Runtime, scoring, authoring, reports, analytics, anti-cheat, evidence, dan rehearsal punya validasi.
4. Ada device test nyata, operator rehearsal nyata, evidence bundle, dan final go/no-go sign-off.

---

## Non-negotiable constraints

- Jangan membuat stack baru PocketBase/SQLite/Alpine.
- Jangan menghidupkan public `/api/cbt/**` route tree.
- Flutter tetap memakai `/api/exam/*` langsung ke Core API.
- Web Admin tetap memakai `/bank-soal/*`, `/asesmen/*`, dan BFF `/api/bank-soal/*` / `/api/asesmen/*`.
- PostgreSQL hanya diakses oleh `services/core-api`.
- Semua schema change lewat migration idempotent.
- Jangan deploy, restart PM2, menjalankan migration live, atau write live data tanpa instruksi eksplisit.
- Semua fase code harus TDD, validasi, independent review, commit terpisah.
- Secrets/log/env harus disanitasi.

---

## Phase 16 — Proposal-to-Product Final Traceability Matrix

**Objective:** Mengunci definisi “100% proposal” agar tidak ada fitur proposal yang terlewat atau salah klaim.

**Files likely to change:**

- Modify: `docs/cbt-proposal-gap-audit.md`
- Create: `docs/cbt-proposal-100-percent-traceability.md`
- Modify: `apps/web-admin/src/lib/cbt/proposal-integration-docs.test.ts`

**Tasks:**

1. Ekstrak ulang daftar requirement proposal dari DOCX.
2. Kelompokkan requirement menjadi:
   - Implemented runtime
   - Implemented docs/evidence
   - Needs runtime implementation
   - Needs manual evidence
   - Adapted/out-of-scope by architecture/BYOD
3. Tambahkan kolom:
   - owner module: Web Admin / Core API / Flutter / Ops
   - acceptance evidence
   - blocking status
   - target phase
4. Tambahkan docs guard test agar matrix wajib menyebut:
   - question types
   - anti-cheat
   - proctoring
   - reports
   - analytics
   - infrastructure/ops
   - manual evidence/sign-off
5. Commit: `docs: add cbt proposal 100 percent traceability matrix`

**Validation:**

```bash
cd apps/web-admin
npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
npm run check
```

---

## Phase 17 — End-to-End Question Type Contract Hardening

**Objective:** Menyamakan authoring → package/session → Flutter answer → scoring → result untuk tipe soal yang sudah ada.

**Scope:**

- `multiple_choice`
- `multiple_answer`
- `essay`
- `short_answer`
- `matching`
- `ordering`
- `true_false`
- `agree_disagree`

**Files likely to change:**

- `services/core-api/internal/service/*cbt*`
- `services/core-api/internal/handler/*cbt*`
- `services/core-api/db/queries/*cbt*`
- `services/core-api/db/migrations/NNN_cbt_question_type_contracts.sql` if needed
- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `apps/mobile/lib/src/models.dart`
- `apps/mobile/lib/src/screens/exam_shell_screen.dart`
- tests under `services/core-api/internal/**`, `apps/web-admin/**`, `apps/mobile/test/**`

**Tasks:**

1. Write failing Core API tests for normalized scoring per type.
2. Ensure answer normalization is deterministic:
   - multiple answer sorted labels
   - matching sorted pairs
   - ordering exact sequence
   - short answer alias matching
   - true_false and agree_disagree label compatibility
3. Write Web Admin tests for composer payloads per type.
4. Write Flutter tests for answer encoding per type.
5. Implement only minimal changes needed.
6. Update `docs/cbt-proposal-gap-audit.md` status for supported types.
7. Commit: `feat: harden cbt question type contracts`

**Validation:**

```bash
cd services/core-api
GOCACHE=/tmp/go-build go test ./...
GOCACHE=/tmp/go-build go build -o /dev/null ./cmd/api

cd apps/web-admin
npm run test:unit
npm run check

cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
```

---

## Phase 18 — Ordering Authoring + Scoring Completion

**Objective:** Menjadikan `ordering` bukan hanya Flutter runtime, tetapi end-to-end: authoring, validation, scoring, result display.

**Files likely to change:**

- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `services/core-api/internal/repository/postgres/cbt_sessions.sql.go` generated after query changes if needed
- `services/core-api/db/queries/cbt_sessions.sql`
- `services/core-api/internal/service/cbt_question_test.go`
- `apps/web-admin/src/routes/bank-soal/**/*test*`

**Tasks:**

1. Add TDD tests for Web Admin composer `ordering` option sequence and answer key.
2. Add Core API scoring tests: exact ordered labels must match; wrong sequence false.
3. Add result display copy for ordering answer/key.
4. Keep answer format comma-separated labels, e.g. `B,A,C`.
5. Commit: `feat: complete ordering authoring and scoring`

**Validation:** same as Phase 17 plus targeted tests.

---

## Phase 19 — True/False and Agree/Disagree Scoring + Authoring Audit

**Objective:** Memastikan `true_false` dan `agree_disagree` konsisten dari Web Admin sampai hasil.

**Files likely to change:**

- `apps/web-admin/src/routes/bank-soal/_components/SoalWorkspacePage.svelte`
- `services/core-api/db/queries/cbt_sessions.sql`
- `services/core-api/internal/service/*test.go`
- `apps/mobile/test/widget_test.dart`

**Tasks:**

1. Lock Web Admin fixed-pair answer key:
   - true_false: existing behavior atau official selected format harus didokumentasikan.
   - agree_disagree: `A=Setuju`, `B=Tidak Setuju`.
2. Ensure Flutter fallback label matches Web Admin scoring.
3. Ensure result display shows human-readable labels.
4. Add regression tests for backend-provided options.
5. Commit: `feat: align fixed pair exam scoring`

---

## Phase 20 — Hotspot Design Decision + Safe Runtime v1

**Objective:** Menentukan apakah hotspot benar-benar dibutuhkan untuk rilis 100%, lalu implementasi aman v1 jika ya.

**Decision gate:** Hotspot membutuhkan schema/API/UI/scoring khusus. Jangan implement fake hotspot.

**Option A — Adapted/out-of-scope:**

- Tetapkan hotspot sebagai out-of-scope untuk BYOD CBT v1.
- Evidence: guard Flutter sudah ada.
- Cocok jika target rilis cepat.

**Option B — Implement Hotspot v1:**

**Files likely to change:**

- new migration for hotspot answer shape if existing `options` insufficient
- Core API question DTO/scoring
- Web Admin composer hotspot editor
- Flutter hotspot image display/tap coordinate capture
- tests across Web/Go/Flutter

**Tasks for Option B:**

1. Write design doc: coordinate system, image asset requirement, scoring tolerance.
2. Migration for hotspot metadata if needed.
3. Web Admin authoring: upload/select image, define target zones.
4. Core API scoring: coordinate-in-zone.
5. Flutter UI: image with tap marker; answer as safe JSON or normalized coordinate string.
6. Result display: selected coordinate/zone.
7. Commit series:
   - `docs: design cbt hotspot runtime`
   - `feat: add hotspot question contract`
   - `feat: add flutter hotspot exam runtime`

**Recommendation:** Do this only after Phase 17–19.

---

## Phase 21 — Upload/File Answer Runtime v1

**Objective:** Mendukung jawaban upload bila proposal wajib memuat file answer.

**Decision gate:** Wajib ada storage policy, file size, MIME whitelist, virus/abuse policy, scoring/manual review.

**Files likely to change:**

- Core API upload endpoint under existing exam/auth contract, not public `/api/cbt`
- storage path/config
- migration for answer attachment metadata
- Web Admin manual review UI
- Flutter file picker/upload UI

**Tasks:**

1. Design file policy:
   - max size
   - allowed MIME
   - storage location
   - retention
   - access control
2. Backend tests for upload auth, size, MIME, participant scope.
3. Flutter tests for upload pending/failure/retry state.
4. Web Admin tests for reviewer download/access.
5. Implement v1 manual-scored upload answer.
6. Commit series:
   - `docs: design cbt upload answer policy`
   - `feat: add cbt upload answer backend`
   - `feat: add flutter upload answer runtime`

---

## Phase 22 — Audio/Video Prompt and Response Policy

**Objective:** Menutup gap proposal audio/video secara realistis.

**Scope split:**

- Audio/video prompt/media: likely already partially supported by asset fields.
- Audio/video answer/recording: separate high-risk feature.

**Tasks:**

1. Audit current media fields and Flutter prompt playback.
2. Add tests for audio prompt one-time/required play if needed.
3. Decide whether recording answers are needed.
4. If recording answers required, reuse upload-answer policy with stricter MIME/size.
5. Update evidence/docs.
6. Commit: `feat: harden cbt media prompt runtime` or separate upload recording commit.

---

## Phase 23 — Proctor Dashboard Full Evidence Mode

**Objective:** Menyempurnakan dashboard pengawas agar proposal proctoring punya bukti lengkap.

**Files likely to change:**

- `apps/web-admin/src/routes/asesmen/pengawasan/**`
- Core API proctor/session handlers
- tests under web-admin and Go

**Tasks:**

1. Lock room/session participant status display.
2. Add event timeline evidence view:
   - heartbeat
   - app backgrounded/resumed
   - device mismatch
   - submit guard
   - stale connection
3. Add proctor action evidence:
   - warning sent
   - force submit
   - reset access
4. Add export/print evidence per session/room.
5. Commit: `feat: add cbt proctor evidence dashboard`

**Explicit non-goal:** screen preview/remote desktop unless school-managed device policy exists.

---

## Phase 24 — Anti-Cheat BYOD Evidence Completion

**Objective:** Mengubah anti-cheat dari “ada telemetry” menjadi “terbukti di perangkat nyata”.

**Files likely to change:**

- `apps/mobile/DEVICE_TEST_MATRIX.md`
- `apps/mobile/RELEASE_CHECKLIST.md`
- `docs/cbt-release-final-evidence.md`
- potentially Flutter tests/docs

**Tasks:**

1. Build RC APK.
2. Test minimal 2 vendor Android nyata.
3. Fill matrix for:
   - FLAG_SECURE screenshot prevention
   - app switch event
   - resume gate
   - heartbeat loss
   - pending answer recovery
   - manual submit guard
4. Attach APK hash/version/API base URL evidence.
5. Commit docs/evidence update: `docs: record cbt android device evidence`

**Manual evidence required:** yes.

---

## Phase 25 — Analytics and Item Analysis Completion

**Objective:** Menutup gap analitik proposal.

**Files likely to change:**

- Core API results/item-analysis queries/services
- `apps/web-admin/src/routes/bank-soal/analisis-butir/+page.svelte`
- `apps/web-admin/src/routes/asesmen/hasil/**`
- tests

**Tasks:**

1. Audit current item-analysis fields.
2. Add tested metrics only:
   - difficulty index
   - discrimination index if sample allows
   - distractor selection counts
   - unanswered count
   - per-type accuracy
3. Decide whether Cronbach alpha is required.
4. If yes, add formula doc and tests against fixed dataset.
5. Add CSV export for item analysis.
6. Commit: `feat: expand cbt item analysis metrics`

**Risk:** Avoid fake psychometrics if data/sample insufficient.

---

## Phase 26 — Reports PDF/Excel Parity

**Objective:** Menutup gap proposal laporan.

**Likely report set:**

- Daftar peserta
- Kartu peserta/token
- Berita acara
- Absensi ruang
- Rekap hasil sesi
- Analisis butir
- Proctor/audit event recap
- Final evidence bundle index

**Files likely to change:**

- Web Admin report/export routes
- Core API export endpoints if needed
- docs/evidence templates

**Tasks:**

1. Inventory current printable/CSV reports.
2. Define official report templates.
3. Add missing CSV first.
4. Add printable HTML/PDF path if project already has a safe print pattern.
5. Add tests for permissions and redaction.
6. Commit: `feat: complete cbt report exports`

---

## Phase 27 — Operator Rehearsal Workflow Completion

**Objective:** Membuktikan operator/pengawas bisa menjalankan CBT tanpa developer.

**Files likely to change:**

- `docs/cbt-smoke-checklist.md`
- `docs/cbt-release-evidence-template.md`
- `docs/cbt-release-final-evidence.md`
- possibly Web Admin UX copy/help panels

**Tasks:**

1. Run rehearsal script:
   - create question
   - review/publish
   - create package
   - create session/room
   - assign participants
   - login Flutter
   - answer/save/submit
   - proctor warning
   - view result
   - export report
2. Record blockers.
3. Patch UX/copy/small bugs found.
4. Update evidence template with real outputs.
5. Commit: `docs: record cbt operator rehearsal evidence`

**Manual evidence required:** yes.

---

## Phase 28 — Infrastructure, Backup, Restore, and DR Evidence

**Objective:** Menutup gap proposal terkait operasional server dan risiko.

**Files likely to change:**

- `deploy/scripts/*`
- `docs/cbt-release-final-evidence.md`
- `docs/cbt-smoke-checklist.md`
- backup/restore docs

**Tasks:**

1. Verify backup script runs and produces PostgreSQL backup.
2. Verify restore rehearsal on non-production/temp DB if possible.
3. Verify PM2 startup/restart procedures.
4. Verify health scripts.
5. Add evidence for:
   - backup path
   - latest symlink
   - checksum
   - restore rehearsal status
6. Commit: `docs: add cbt backup restore evidence`

**Do not restore over live DB.**

---

## Phase 29 — Security and ISO-Control Alignment Evidence

**Objective:** Menutup gap proposal ISO/security as control alignment, bukan klaim sertifikasi.

**Tasks:**

1. Review RBAC routes used by CBT.
2. Review secret scan evidence.
3. Review upload/media hygiene if Phase 21/22 implemented.
4. Review audit logs for key actions.
5. Add evidence page:
   - RBAC controls
   - rate limit
   - token/device binding
   - audit trail
   - evidence redaction
   - backup/restore
6. Commit: `docs: add cbt security control evidence`

**Non-goal:** claim ISO certification without external audit.

---

## Phase 30 — Mobile RC Build and Release Package

**Objective:** Menjadikan semua Flutter runtime changes tersedia untuk siswa.

**Files likely to change:**

- `apps/mobile/RELEASE_CHECKLIST.md`
- `apps/mobile/DEVICE_TEST_MATRIX.md`
- release evidence docs

**Tasks:**

1. Build APK/AAB RC with correct API base URL.
2. Record:
   - version name/code
   - commit hash
   - APK hash
   - signing status
   - target backend URL
3. Install on test devices.
4. Run smoke exam end-to-end.
5. Commit release evidence docs.

**Validation commands:**

```bash
cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
/home/servermtsn2kolut/development/flutter/bin/flutter build apk --release
sha256sum build/app/outputs/flutter-apk/app-release.apk
```

---

## Phase 31 — Final Production Deployment Candidate

**Objective:** Setelah semua code/docs/evidence siap, deploy candidate ke server live dengan backup dan smoke.

**Prerequisite:** User gives explicit deploy approval.

**Tasks:**

1. Confirm clean worktree.
2. Build backend/web/worker.
3. Backup PostgreSQL.
4. Reconcile migration drift.
5. Apply migrations if needed.
6. Restart PM2 backend/web/worker.
7. Save PM2.
8. Smoke:
   - backend health
   - frontend health
   - worker health
   - protected routes 302/401
   - CBT exam API unauth 401
   - Bank Soal/Asesmen route smoke
9. Sanitize logs.
10. Commit deployment evidence update.

**Commit:** `docs: record cbt production candidate evidence`

---

## Phase 32 — Final Go/No-Go Sign-Off

**Objective:** Membuat klaim final “proposal 100% implemented/adapted” secara aman dan dapat diaudit.

**Files likely to change:**

- `docs/cbt-release-final-evidence.md`
- `docs/cbt-proposal-100-percent-traceability.md`
- `deploy/scripts/cbt-final-readiness.sh`

**Tasks:**

1. Generate final readiness bundle.
2. Fill all manual evidence fields.
3. Ensure every proposal item status is one of:
   - implemented
   - implemented with evidence
   - adapted by approved architecture
   - out-of-scope with formal rationale
4. Set `go_no_go` only after approval.
5. Set `production_go=true` only after real sign-off.
6. Commit: `docs: finalize cbt proposal 100 percent signoff`

---

## Suggested execution order

Best order for actual work:

1. Phase 16 — traceability matrix
2. Phase 17 — question type contract hardening
3. Phase 18 — ordering end-to-end
4. Phase 19 — true_false/agree_disagree end-to-end
5. Phase 25 — analytics
6. Phase 26 — reports
7. Phase 23 — proctor evidence dashboard
8. Phase 24 — real Android evidence
9. Phase 28 — backup/restore evidence
10. Phase 29 — security evidence
11. Phase 30 — APK RC build
12. Phase 27 — operator rehearsal
13. Phase 31 — deployment candidate, only with explicit approval
14. Phase 32 — final sign-off
15. Phase 20/21/22 — hotspot/upload/audio-video only if Bapak wants proposal feature parity beyond practical CBT v1

---

## Practical definition of 100%

Because some proposal items are not realistic for BYOD or not part of the approved architecture, the final claim should be worded like this:

> “Proposal CBT MTsN 2 Kolaka Utara telah diimplementasikan 100% ke monorepo sebagai fitur runtime, evidence operasional, atau adaptasi resmi sesuai arsitektur Web Admin + Core API + PostgreSQL + Flutter. Item yang tidak diterapkan mentah-mentah seperti PocketBase/SQLite/Alpine, kiosk penuh BYOD, screen preview, desktop agent, dan ISO certification dicatat sebagai adapted/out-of-scope dengan rationale dan evidence.”

---

## Global validation gate for every code phase

Run as applicable:

```bash
git diff --check

cd services/core-api
GOCACHE=/tmp/go-build go test ./...
GOCACHE=/tmp/go-build go build -o /dev/null ./cmd/api

cd apps/web-admin
npm run test:unit
npm run check

cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test

cd /home/servermtsn2kolut/mtsn2kolut-super-app
make ops-health
```

For every phase:

- independent review before commit
- stage only intended files
- no generated binaries in git
- no secrets in output
- concise PASS report with commit hash
