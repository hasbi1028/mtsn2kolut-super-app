# CBT Standardization Contract

Implemented scope for Sprint 0-5 standardization.

## Sprint 0 — Safety Baseline
- Additive-only migration strategy.
- Deploy order: backup DB -> apply migrations -> build/restart core-api -> health check -> build/restart web-admin -> smoke test.

## Sprint 1 — Operational Readiness Board
- `GET /api/asesmen/readiness` returns selectable events plus selected overview.
- `GET /api/asesmen/events/{id}/readiness` aliases event readiness/overview.
- Readiness summarizes kegiatan, target soal, package, session, peserta, ruang/kursi, pengawas, token/kartu, and blockers.

## Sprint 2 — Package Lock & Snapshot
- Migration `101_cbt_standardization_runtime.sql` adds lock metadata to `cbt_packages` and immutable `cbt_package_question_snapshots`.
- Session status transition to `scheduled` or `active` locks/snapshots package transactionally.
- Runtime and scoring prefer snapshot rows when available, fallback to live package questions for legacy sessions.

## Sprint 3 — Token Hardening
- Additive token hash/version/reveal/revoke metadata is present for participant and room tokens.
- New generation/regeneration writes hash metadata while keeping legacy plaintext columns for compatibility during migration.
- Runtime login supports hashed-token verification with legacy fallback.

## Sprint 4 — Runtime Randomization + Auto Submit
- Participant option order is persisted in `cbt_exam_participants.option_order` when package randomizes options.
- Student-visible option labels can be randomized while submitted answers are canonicalized before storage/scoring.
- `POST /api/asesmen/sessions/{id}/finalize-overdue` auto-submits overdue participants and scores the session idempotently.

## Sprint 5 — Result Sync + Remediasi
- `GET /api/asesmen/sessions/{id}/grade-sync-preflight` reports safe gradebook sync preflight.
- `GET /api/asesmen/sessions/{id}/remedial-candidates?threshold=75` lists candidates and KD/indicator/material gaps.
- Broad grade overwrite is intentionally not performed without a later explicit operator-confirmed workflow.
