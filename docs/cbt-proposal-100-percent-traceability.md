# CBT Proposal 100 Percent Traceability Matrix

Dokumen ini mengunci definisi “100% implementasi proposal” untuk Sistem CBT MTsN 2 Kolaka Utara. Proposal DOCX tetap menjadi sumber kebutuhan produk, tetapi implementasi resmi mengikuti arsitektur monorepo: Web Admin SvelteKit, Core API Go + PostgreSQL, Flutter Android, dan Ops evidence. Tidak ada stack runtime PocketBase / SQLite / Alpine.js dan tidak ada public `/api/cbt/**` route tree baru.

## Status vocabulary

- `Implemented runtime`: tersedia pada aplikasi/runtime atau kontrak API aktif.
- `Implemented docs/evidence`: tersedia sebagai guard, checklist, evidence generator, atau SOP.
- `Needs runtime implementation`: masih perlu patch fitur aplikasi.
- `Needs manual evidence`: perlu bukti HP Android nyata, rehearsal operator, atau sign-off.
- `Adapted/out-of-scope`: kebutuhan proposal diadaptasi karena beda arsitektur, BYOD limitation, keamanan, atau scope rilis.

## Final traceability matrix

| Requirement proposal | Owner module | Current status | Acceptance evidence | Blocking status | Target phase |
| --- | --- | --- | --- | --- | --- |
| Question types: multiple_choice | Web Admin / Core API / Flutter | Implemented runtime | Composer options, `/api/exam/*`, scoring exact label, Flutter single option | Not blocking | Phase 17 |
| Question types: multiple_answer | Web Admin / Core API / Flutter | Implemented runtime | Deterministic sorted labels `A,C`, scoring order-insensitive, Flutter multi answer contract | Not blocking after Phase 17 hardening | Phase 17 |
| Question types: essay | Web Admin / Core API | Implemented runtime | Rubric authoring, manual score, ungraded essay list | Not blocking; depends on korektor workflow evidence | Phase 17 / Phase 27 |
| Question types: short_answer | Web Admin / Core API | Implemented runtime | Alias answer key separated by `|`, whitespace/case normalization in scoring | Not blocking after Phase 17 hardening | Phase 17 |
| Question types: matching | Web Admin / Core API | Implemented runtime | Pair answer key `A=1;B=2`, pair scoring sorted by left label | Not blocking | Phase 17 |
| Question types: ordering | Web Admin / Core API / Flutter | Implemented runtime | Authoring contract, answer key and Flutter answer format comma-separated labels like `B,A,C`, scoring exact sequence | Not blocking after Phase 18 | Phase 18 |
| Question types: true_false | Web Admin / Core API / Flutter | Implemented runtime | Fixed pair `A=Benar`, `B=Salah`, Flutter fallback submits `A`/`B`, backend scoring tolerates legacy `true`/`false` cached answers | Not blocking | Phase 17 / Phase 19 |
| Question types: agree_disagree | Web Admin / Core API / Flutter | Implemented runtime | Fixed pair `A=Setuju`, `B=Tidak Setuju`, label-compatible Flutter fallback | Not blocking | Phase 17 / Phase 19 |
| Question types: hotspot | Flutter / Web Admin / Core API | Adapted/out-of-scope for BYOD CBT v1 | `docs/cbt-hotspot-design-decision.md`; `hotspot_safe_status: adapted_deferred`; Flutter unsupported guard; no fake hotspot | Feature parity blocker only if user chooses hotspot v1 after schema/UI/scoring design | Phase 20 |
| Question types: upload_answer / file_upload | Flutter / Web Admin / Core API / Ops | Policy deferred | `docs/cbt-upload-answer-policy.md`; `upload_answer_safe_status: policy_deferred`; no upload runtime until secure storage/auth/MIME/size/retention/review tests exist | Feature parity blocker only if proposal file answer is mandatory | Phase 21 |
| Audio/video prompt/response | Web Admin / Flutter / Ops | Prompt partial; recording policy deferred | Existing image/audio prompt fields; `docs/cbt-media-prompt-response-policy.md`; `recording_answer_safe_status: policy_deferred`; no recording upload outside file-answer policy | Deferred for recording; prompt media remains bounded to existing fields | Phase 22 |
| Anti-cheat BYOD layers | Flutter / Core API / Ops | Implemented docs/evidence + partial runtime | Lifecycle telemetry, screenshot protection where OS allows, heartbeat, submit guard, audit event, deterministic APK hash instructions, Phase 30 Mobile RC Build and Release Package evidence | Needs manual evidence on physical Android devices; no fabricated PASS | Phase 24 / Phase 30 |
| Proctoring dashboard | Web Admin / Core API | Implemented runtime + evidence helper | Room dashboard, participant events, suspicious flag, handover recap, token-free evidence CSV, print pack, no screen preview/remote desktop, Phase 27 Operator Rehearsal Workflow Completion checklist | Needs operator rehearsal evidence | Phase 23 / Phase 27 |
| Reports | Web Admin / Core API | Implemented docs/evidence + current runtime exports | Official template matrix; print HTML/browser PDF; CSV Excel-compatible where safe; no broad token spreadsheet export | Binary PDF/XLSX endpoints deferred unless safe pattern is approved | Phase 26 |
| Analytics | Web Admin / Core API | Implemented runtime/docs for tested metrics | Item analysis query, difficulty/discrimination/distribution, unanswered count, per-type accuracy helper; Cronbach alpha deferred | No fake psychometrics; Cronbach requires real tested implementation | Phase 25 |
| Infrastructure/ops | Ops | Implemented docs/evidence | PM2 architecture, backup scripts, release preflight/readiness scripts, backup path/latest symlink/checksum verification, `pg_restore --list` evidence, DR RTO/RPO checklist | Production go still requires explicit deploy instruction and sign-off; do not restore over live DB | Phase 28 / Phase 31 |
| Security and ISO controls | Ops / Backend / Web Admin | Implemented docs/evidence + partial runtime | RBAC controls, rate limit/trusted proxy, token/device binding, audit trail, evidence redaction, secret scan, backup/restore evidence; Security and ISO-Control Alignment Evidence is control alignment, not certification | No ISO certification claim; external certification requires formal auditor process | Phase 29 |
| Manual evidence/sign-off | Ops / Operator | Needs manual evidence | Device matrix, operator rehearsal, final go/no-go sign-off, backup/DR review, security alignment review, mobile RC package review | Blocking for production go claim | Phase 24 / Phase 27 / Phase 30 / Phase 32 |
| PocketBase / Alpine.js / SQLite proposal stack | Architecture | Adapted/out-of-scope | Official boundary: PostgreSQL only via Core API; SvelteKit and Flutter remain runtime | Not blocking; forbidden to add as runtime | Phase 16 |
| Public `/api/cbt/**` route tree | Architecture / Security | Adapted/out-of-scope | Official boundary: Flutter uses `/api/exam/*`; Web Admin uses BFF/backend proxy routes | Not blocking; forbidden to add | Phase 16 |

## Phase 16 conclusion

Traceability is complete enough to prevent overclaiming. “100% proposal” may only be claimed after each row is either implemented runtime, implemented evidence, completed manual evidence, or formally adapted/out-of-scope with rationale.

## Phase 17 question type contract

Canonical supported types for end-to-end hardening are: `multiple_choice`, `multiple_answer`, `essay`, `short_answer`, `matching`, `ordering`, `true_false`, and `agree_disagree`.

- `multiple_choice`: one label, exact label scoring.
- `multiple_answer`: comma-separated labels sorted deterministically, e.g. `A,C`.
- `essay`: manual scoring only; automated correctness remains `NULL`.
- `short_answer`: aliases separated by `|`; case, repeated whitespace, and non-breaking spaces are normalized.
- `matching`: semicolon-separated pairs, e.g. `A=1;B=2`; comparison is deterministic by sorted pairs.
- `ordering`: comma-separated labels in exact sequence, e.g. `B,A,C`; comparison preserves order.
- `true_false`: fixed labels `A=Benar`, `B=Salah`.
- `agree_disagree`: fixed labels `A=Setuju`, `B=Tidak Setuju`.

## Phase 18 ordering completion boundary

Ordering is completed as a monorepo contract without schema migration: Web Admin authoring exposes `ordering`, Core API validates all option labels exactly once, scoring checks exact comma-separated sequence, and Flutter already submits/restores the same format. Result display can use the existing participant answer/key fields until Phase 26 report export polish.

## Phase 19-22 decision status

- Phase 19 fixed-pair alignment: `true_false` uses `A=Benar`, `B=Salah`; `agree_disagree` uses `A=Setuju`, `B=Tidak Setuju`. Flutter fallback now matches Web Admin/Core API labels, while backend scoring remains tolerant of old `true`/`false` cached true/false answers.
- Phase 20 hotspot: `hotspot_safe_status: adapted_deferred`; no fake hotspot runtime, no schema migration, and no tap-coordinate scoring until a separate approved design exists.
- Phase 21 upload/file answer: `upload_answer_safe_status: policy_deferred`; no runtime upload endpoint, Flutter picker, storage path, or reviewer download UI until secure storage/auth/MIME/size/retention/review tests exist.
- Phase 22 media prompt/response: existing image/audio prompt fields remain the bounded runtime surface; video prompt is adapted/deferred and recording answers use `recording_answer_safe_status: policy_deferred`.

## Phase 23-26 evidence, analytics, and reports

- Phase 23 proctor evidence mode covers heartbeat, app background/resume, device mismatch, submit guard, stale connection, warning, force submit, reset access, and export/print evidence. Screen preview and remote desktop are explicit non-goals.
- Phase 24 anti-cheat BYOD evidence remains `pending_manual_evidence` until operator/pengawas test physical Android devices from at least two vendors. Build/hash instructions do not imply a real-device PASS.
- Phase 25 analytics accepts difficulty index, discrimination index with sample guard, answer distribution, unanswered count, and per-type accuracy. `Cronbach alpha: documented_deferred_until_formula_and_dataset_are_tested`.
- Phase 26 reports use an official template matrix: PDF parity is print HTML/browser PDF, Excel parity is CSV where safe, and sensitive token reports remain print-only or role-bound.

## Phase 27-30 follow-up status - rehearsal, DR, security, and mobile RC evidence

- Phase 27 Operator Rehearsal Workflow Completion requires Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review, proctor evidence, role/scope/token boundary, event/audit evidence, and go/no-go rehearsal.
- Phase 28 Infrastructure, Backup, Restore, and DR Evidence records backup path, latest symlink, checksum, `sha256sum -c`, `pg_restore --list`, RTO/RPO notes, and DR owner. Do not restore over live DB.
- Phase 29 Security and ISO-Control Alignment Evidence is control alignment, not certification. No ISO certification claim is made by repository evidence.
- Phase 30 Mobile RC Build and Release Package records RC identifier, commit hash, version name/code, signing status, `API_BASE_URL`, APK SHA-256 hash, and build status. Real-device PASS claimed only after physical Android operator test on minimum two Android vendors.

## Phase 31-32 follow-up status - deployed candidate and go/no-go boundary

- Phase 31 Final Production Deployment Candidate deployed commit `e005134` to the live PM2 runtime after explicit user approval, with clean worktree preflight, backend/web/worker builds, PostgreSQL backup, `make db-migrate`, PM2 restart/save, and smoke checks.
- Phase 31 smoke evidence: backend `/health` `200`, web root `200`, protected pages `302`, protected BFF/API endpoints `401`, and `bash deploy/scripts/health-check.sh all` PASS.
- Phase 32 Final Go/No-Go Sign-Off keeps the final claim auditable: automated/deployment evidence is complete, but `production_go` remains `false` until physical Android device matrix, operator rehearsal, and explicit final sign-off are complete.
- Final safe claim must describe proposal items as implemented runtime, implemented with evidence, adapted by approved architecture, or out-of-scope with formal rationale; no ISO certification or full BYOD kiosk claim is made.
