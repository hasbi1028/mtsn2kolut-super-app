# CBT Proposal Integration Phase 19-22 - Fixed Pair and High-Risk Question Decisions

Status: Phase 19-22 scoped implementation record, 2026-05-08.

This phase continues from the Phase 16-18 baseline at commit `cebd789`. The scope is limited to fixed-pair question alignment, formal hotspot decision, upload/file-answer policy, and media prompt/response policy. It does not deploy, restart PM2, run migrations, write live data, create a public CBT route tree, or introduce PocketBase, SQLite, Alpine, or a separate CBT runtime.

## Phase 19 - Fixed Pair Authoring and Scoring

The official fixed-pair answer labels are:

- `true_false`: `A=Benar`, `B=Salah`.
- `agree_disagree`: `A=Setuju`, `B=Tidak Setuju`.

Web Admin authoring, Core API validation, Flutter fallback options, scoring, and result review should use those labels. Backend scoring may tolerate old Flutter `true` / `false` cached answers for `true_false`, but new Flutter fallback answers must submit `A` or `B`.

Result displays should be human-readable: Benar/Salah and Setuju/Tidak Setuju, not opaque boolean strings.

## Phase 20 - Hotspot Decision

Decision status:

```text
hotspot_safe_status: adapted_deferred
```

Hotspot is adapted/out-of-scope for BYOD CBT v1. There is no fake hotspot runtime. Flutter keeps the unsupported-question guard and asks the student to call the pengawas for manual handling if a package contains `hotspot`.

Full hotspot runtime requires a separate approved design for coordinate system, image asset requirement, scoring tolerance, authoring UI, answer serialization, result display, and migration/API safety. Phase 20 does not add schema, scoring, authoring, or Flutter tap-coordinate capture.

## Phase 21 - Upload/File Answer Policy

Decision status:

```text
upload_answer_safe_status: policy_deferred
```

Upload/file answers are policy-designed but runtime-deferred. No upload-answer endpoint, schema, Flutter file picker, or reviewer download path is enabled by this phase.

Runtime may be revisited only after secure storage, participant-scoped auth, max size, allowed MIME, extension checks, safe disposition, `nosniff`, retention, access control, virus/abuse review, and manual review workflow can be tested safely.

## Phase 22 - Media Prompt and Recording Policy

Decision status:

```text
recording_answer_safe_status: policy_deferred
```

Image and audio prompt support remains bounded to existing exam payload fields and Flutter rendering. Audio/video answer recording is not enabled. Recording upload must reuse the Phase 21 upload/file answer policy and pass stricter MIME, size, retention, access-control, and review requirements before any runtime work.

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

Phase 19-22 diterima bila:

- `true_false` and `agree_disagree` fixed-pair labels are documented and tested.
- Flutter fallback labels match Web Admin/Core API scoring.
- hotspot has a formal adapted/deferred decision with no fake hotspot.
- upload/file answer has a policy but remains runtime-deferred.
- media prompt and recording policy is explicit, with recording deferred behind upload/file answer policy.
- no public `/api/cbt/**` runtime route, schema migration, deployment, PM2 restart, live SQL write, or secret is introduced.
