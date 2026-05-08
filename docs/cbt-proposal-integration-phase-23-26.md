# CBT Proposal Integration Phase 23-26 - Evidence, Analytics, and Reports

Status: Phase 23-26 scoped implementation record, 2026-05-08.

This phase continues from HEAD `488e105` Phase 19-22. Scope is limited to proctor evidence mode, BYOD anti-cheat evidence completion templates, item-analysis guardrails, and report template parity. It does not deploy, restart PM2, run migrations, write live data, create a public CBT route tree, or introduce PocketBase, SQLite, Alpine, or a separate CBT runtime.

## Phase 23 - Proctor Dashboard Full Evidence Mode

Dashboard pengawas ruang now has a small evidence helper and token-free CSV export for room evidence. Existing runtime support is used only where already present:

- heartbeat: Core API records `heartbeat` events and `last_heartbeat`; dashboard shows Online/Waspada/Offline.
- app background/resume: Flutter sends `app_switch` and `warning` reasons such as `resume_exam` and `repeat_resume_attempt`.
- device mismatch: official evidence remains the controlled `/api/exam/*` `409 token already bound to another device` flow and Flutter local-answer preservation; dashboard evidence may record related warning/reset notes without exposing full device fingerprint.
- submit guard: warning reasons `submit_blocked_pending_sync`, `auto_submit_blocked_pending_sync`, and `submit_blocked_degraded_mode`.
- stale connection: warning reasons `stale_connection_attention`, `stale_connection_escalated`, degraded-mode warnings, and stale `last_heartbeat`.
- warning: generic `warning` telemetry remains visible in the room event timeline.
- force submit: `proctor_force_submit` event from approved pengawas/operator action.
- reset access: `proctor_reset_access` event from approved pengawas/operator action.
- export/print evidence: room evidence CSV from `/asesmen/sesi/[id]/rooms/[rid]/proctoring` and print pack from `/asesmen/sesi/[id]/rooms/[rid]/print-pack`.

Explicit non-goal: no screen preview, no remote desktop, no browser lockdown claim, and no BYOD kiosk guarantee. Evidence mode is telemetry and operational audit, not device-owner surveillance.

Implementation guard:

- `apps/web-admin/src/lib/cbt/proctor-evidence.ts`
- `apps/web-admin/src/lib/cbt/proctor-evidence.test.ts`
- `apps/web-admin/src/routes/asesmen/sesi/[id]/rooms/[rid]/proctoring/+page.svelte`

## Phase 24 - Anti-Cheat BYOD Evidence Completion

Manual real-device evidence is required and cannot be fabricated. Current status remains:

```text
anti_cheat_byod_manual_status: pending_manual_evidence
real_device_pass_claim: forbidden_until_operator_tested
minimum_android_vendors: 2
```

The repository can provide templates and deterministic build/hash instructions, but it cannot mark real devices PASS from this host.

Deterministic operator commands for an RC build:

```bash
cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter pub get
/home/servermtsn2kolut/development/flutter/bin/flutter analyze
/home/servermtsn2kolut/development/flutter/bin/flutter test
/home/servermtsn2kolut/development/flutter/bin/flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
sha256sum build/app/outputs/flutter-apk/app-release.apk
```

Manual matrix must cover `FLAG_SECURE` screenshot/recent-preview deterrence, app switch event, resume gate, heartbeat loss, pending answer recovery, manual submit guard, device mismatch `409`, stale connection warning, and final submit on a healthy connection. Results stay `pending_manual_evidence` until real Android devices are tested by operator/pengawas.

## Phase 25 - Analytics and Item Analysis Completion

Existing runtime item-analysis metrics are accepted only where formulas already exist:

- difficulty index.
- discrimination index with sample-size interpretation guard.
- distractor / answer selection distribution.
- unanswered / blank count.
- per-type accuracy summary using existing item-analysis rows.

Cronbach alpha is not emitted in this phase.

```text
Cronbach alpha: documented_deferred_until_formula_and_dataset_are_tested
```

No fake psychometrics are allowed. Cronbach alpha requires a separately implemented formula, fixed validation dataset, service/UI contract, and tests before any value can appear in product reports.

Implementation guard:

- `apps/web-admin/src/lib/cbt/item-analysis-evidence.ts`
- `apps/web-admin/src/lib/cbt/item-analysis-evidence.test.ts`
- `/asesmen/sesi/[id]` item-analysis tab summarizes accuracy per question type from existing rows.

## Phase 26 - Reports PDF/Excel Parity

Official report parity follows safe existing patterns:

- PDF parity means print-ready HTML with browser "Save as PDF" where print routes already exist.
- Excel parity means CSV that opens in spreadsheet tools where CSV export already exists.
- No new binary PDF/XLSX endpoint is introduced in this phase.
- Sensitive token reports remain print-only or role-bound; no broad token spreadsheet export is added.

Official template matrix:

| Report | Canonical route/artifact | PDF path | Excel path | Redaction / guard |
|--------|--------------------------|----------|------------|-------------------|
| Daftar peserta | `/asesmen/sesi/[id]/rooms/[rid]/print-pack` | print HTML / browser PDF | not applicable for sensitive tokens | token print only for authorized operations |
| Kartu peserta/token | `/asesmen/kegiatan/[id]/exam-cards` | print HTML / browser PDF | not applicable for sensitive tokens | token output stays operational print material |
| Berita acara | `/asesmen/sesi/[id]/minutes` | print HTML / browser PDF | not applicable | role-bound token visibility |
| Absensi ruang | `/asesmen/sesi/[id]/rooms/[rid]/print-pack` | print HTML / browser PDF | not applicable for sensitive tokens | attendance/token stays room print pack |
| Rekap hasil sesi | `/asesmen/sesi/[id]` | print HTML / browser PDF | CSV Excel-compatible | no answer key |
| Analisis butir | `/asesmen/sesi/[id]` | print HTML / browser PDF | CSV Excel-compatible | answer key role redaction, no fake psychometrics |
| Proctor/audit event recap | `/asesmen/sesi/[id]/rooms/[rid]/proctoring` | print pack / browser PDF evidence | token-free CSV evidence | no raw token, password, answer key, or full fingerprint |
| Final evidence bundle index | `docs/cbt-release-evidence-template.md` + readiness JSON | markdown/archive | JSON manifest | secret scan and ops-controlled archive |

Implementation guard:

- `apps/web-admin/src/lib/cbt/report-template-matrix.ts`
- `apps/web-admin/src/lib/cbt/report-template-matrix.test.ts`

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak mengubah schema database atau migration.
- Tidak melakukan live DB writes.
- Tidak mengubah route `/api/cbt`.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.

## Acceptance Criteria

Phase 23-26 diterima bila:

- proctor evidence mode covers heartbeat, app background/resume, device mismatch, submit guard, stale connection, warning, force submit, reset access, and export/print evidence without screen preview or remote desktop.
- BYOD anti-cheat real-device matrix remains pending/manual unless operator evidence is actually supplied.
- item-analysis documents implemented metrics and keeps Cronbach alpha deferred until real tested implementation exists.
- report template matrix maps PDF/Excel parity to existing print HTML and CSV-safe patterns.
- no public `/api/cbt/**` runtime route, schema migration, deployment, PM2 restart, live SQL write, or secret is introduced.
