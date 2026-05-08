# CBT Proposal Gap Audit

Status: final proposal gap audit baseline, 2026-05-08.

Proposal source: `/home/servermtsn2kolut/.hermes/document_cache/doc_2954fa0c7a5c_Proposal_Sistem_CBT_MTsN2_Kolaka_Utara.docx`.

This feature-by-feature matrix compares the proposal sections against the current monorepo baseline at commit `488e105`. The audit is intentionally adapted to the actual architecture: SvelteKit Web Admin, Go Core API, PostgreSQL, Flutter BYOD exam client, and PUSAKA worker. PocketBase / Alpine.js / SQLite proposal stack is adapted to the approved monorepo boundaries and is not introduced as runtime.

Status vocabulary:

- Implemented: supported by current repository features or guarded operational evidence.
- Partial: present in the product but not every proposal detail is implemented.
- Planned manual evidence: requires physical Android devices, operator rehearsal, live health observation, or ops-controlled evidence outside this host.
- Out of scope adapted: proposal detail is intentionally replaced by the current approved architecture or BYOD constraint.

## Feature-by-feature matrix

| Proposal area | Proposal section | Status | Current repository evidence | Next action |
|---------------|------------------|--------|-----------------------------|-------------|
| Multi-mode assessment | 4.1 Multi-Mode Assessment Engine | Partial | CBT sessions/events, Bank Soal, paket, non-test assessments, results hubs, and Flutter exam runtime exist. Diagnostic adaptive CAT, game-based leaderboard, and unlimited-attempt formative behavior are not baseline. | Keep CBT and non-test assessment split; add new modes only through existing `/asesmen/*`, `/bank-soal/*`, and Go API contracts after explicit product decision. |
| Question types | 4.2 Tipe Soal yang Didukung | Partial | Pilihan ganda, essay/uraian, rich stem/stimulus, assets, KaTeX-ready authoring, import/export, and manual essay grading are present. Flutter runtime recognizes ordering questions, renders reorder controls from the existing options list, restores saved order, and sends comma-separated labels to `/api/exam/answer`. `true_false` now uses the fixed Web Admin/Core API labels `A=Benar` and `B=Salah` for fallback answers, with backend scoring compatibility for old cached `true`/`false` answers. `agree_disagree` uses `A=Setuju` and `B=Tidak Setuju`, preserves backend-provided options when present, restores saved answers, and sends the selected label through `/api/exam/answer`. Flutter guards unsupported proposal runtime types (`hotspot`, `upload_answer`, `file_upload`) with an explicit manual-supervisor panel instead of blank objective rendering or fake answer submission. Hotspot is formally `hotspot_safe_status: adapted_deferred`; upload/file answer is `upload_answer_safe_status: policy_deferred`; recording answer is `recording_answer_safe_status: policy_deferred`. Full hotspot interaction, upload-answer storage, video prompt, and audio/video answer recording remain deferred behind design/policy. | Keep unsupported runtime guard in place; implement full hotspot/upload-answer/recording only after explicit schema/API, storage, scoring, review, and Flutter rendering tests. |
| Anti-cheat layers | 4.3 Sistem Anti-Cheat Berlapis | Partial | Strong exam tokens, device fingerprint telemetry, `FLAG_SECURE`, app-switch/resume events, heartbeat, degraded-mode submit guard, local pending-answer safety, and proctor warnings are present. Browser lockdown, OS-level desktop agent, face verification, VPN detection, and hard kiosk guarantees are not baseline. | Continue BYOD-realistic deterrence; collect real-device evidence and do not claim full kiosk/device-owner control. |
| Room management | 4.4 Room Management | Implemented | Sessions, rooms, seat numbers, room proctors, readiness, print packs, handover, participant operations, and room-level proctoring routes exist. QR check-in and every proposal code-access behavior remain evidence/backlog dependent. | Verify room setup in operator rehearsal and record any missing check-in or access-code needs as backlog. |
| Proctor dashboard | 4.5 Dashboard Pengawas | Partial | Proctoring routes support room monitoring, warnings/events, force submit, reset access, participant flagging, room print packs, token-free room evidence CSV, and operational recaps. Evidence mode covers heartbeat, app background/resume, device mismatch, submit guard, stale connection, warning, force submit, reset access, and export/print evidence. Screen preview and desktop-agent controls are intentionally absent for BYOD. | Capture real operator rehearsal evidence; device mismatch remains a controlled `/api/exam/*` `409` evidence flow, not full fingerprint exposure. |
| Audit trail | 4.5.3 Audit Trail & Evidensi, 5.3 Audit Log | Partial | Backend audit middleware, CBT event telemetry, auth audit events, session/room audit routes, and evidence templates exist. Immutable hash chain and formal chain-of-custody records are not implemented as product runtime. | Use generated readiness/evidence files plus ops-controlled archive; consider hash-chain audit only after a separate security design. |
| Analytics | 8.1 Analitik Butir Soal | Partial | Results hubs, event/session results, item-analysis route, scoring, essay grading, difficulty index, discrimination index, answer distribution, unanswered count, per-type accuracy helper, and grade sync foundations exist. Cronbach alpha is explicitly `documented_deferred_until_formula_and_dataset_are_tested`. | Keep item-analysis evidence in post-exam review and do not emit fake psychometric values. |
| Reports | 8.2 Laporan yang Tersedia | Partial | CSV exports, printable exam cards, minutes/berita acara, operational recap, session/event results, token-free proctor evidence CSV, and official template matrix exist. PDF parity maps to print HTML/browser PDF; Excel parity maps to CSV where safe. | Use current printable/export artifacts; no new binary PDF/XLSX endpoints or broad token spreadsheet export in this phase. |
| ISO controls | 5.3, 6.1, 6.2, 6.3 | Partial | RBAC, JWT refresh sessions, auth audit events, rate-limit/trusted-proxy baseline, upload hygiene, generic 500 hygiene, docs guard, manifest/checksum evidence, and secret-scan evidence are present. Formal ISO certification, complete ISMS/QMS controls, penetration test, and semester disaster-recovery tests require manual/ops evidence. | Record ISO as control alignment, not certification; attach manual security/ops evidence before any external claim. |
| Infrastructure | 5.1 Stack Teknologi, 5.2 Arsitektur Sistem, 9.2 Kebutuhan Infrastruktur | Out of scope adapted | Approved runtime is 3 VPS targets with native PM2, SvelteKit Web Admin, Go Core API, PostgreSQL, PUSAKA worker, and Flutter APK. No PocketBase, runtime SQLite, Alpine-only UI, Caddy requirement, Docker runtime, or single-server collapse is introduced. | Preserve service boundaries and deploy order; gather `make ops-health` evidence as read-only status, not deploy automation. |
| Risks | 11 Manajemen Risiko | Planned manual evidence | Some mitigations exist in code/docs: server-side timer, autosave/local pending answers, heartbeat, restore, auth hardening, backups/runbooks, and audit evidence templates. UPS, LAN stability, operator behavior, real-device policy, backup restore, and security audit evidence require physical/ops rehearsal. | Complete device matrix, operator rehearsal, evidence bundle review, and final go/no-go sign-off before production acceptance. |

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak mengubah product runtime.
- Tidak mengubah schema database atau migration.
- Tidak melakukan live DB writes.
- Tidak mengubah route `/api/cbt`.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.

## Release conclusion

The proposal is broadly aligned with the current CBT direction, but release acceptance must separate implemented automated evidence from manual evidence. The current host can generate audit/readiness artifacts and run static/test/build checks. It cannot prove physical Android behavior, pengawas practice, LAN/power resilience, or final operational go/no-go without real devices and operator rehearsal.

## Phase 16-18 follow-up status (2026-05-08)

- Phase 16 traceability matrix added: `docs/cbt-proposal-100-percent-traceability.md` maps proposal requirements to owner module, acceptance evidence, blocking status, and target phase.
- Phase 17 question type contract hardened for `multiple_choice`, `multiple_answer`, `essay`, `short_answer`, `matching`, `ordering`, `true_false`, and `agree_disagree`.
- Phase 18 ordering completion: ordering now has an explicit Web Admin/Core API/Flutter contract with comma-separated exact sequence labels such as `B,A,C`; scoring must preserve sequence, unlike `multiple_answer` which is sorted label-set scoring.
- Remaining proposal parity items stay tracked: `hotspot`, `upload_answer` / `file_upload`, audio/video policy, analytics/report parity, anti-cheat physical Android evidence, operator rehearsal, backup/restore evidence, and final sign-off.

## Phase 19-22 follow-up status (2026-05-08)

- Phase 19 fixed pair audit: `true_false` and `agree_disagree` are aligned to Web Admin/Core API labels and tested through Flutter fallback plus backend scoring guard.
- Phase 20 hotspot decision: `docs/cbt-hotspot-design-decision.md` records `hotspot_safe_status: adapted_deferred`; no fake hotspot runtime, schema migration, tap capture, or coordinate scoring was added.
- Phase 21 upload/file answer policy: `docs/cbt-upload-answer-policy.md` records `upload_answer_safe_status: policy_deferred`; runtime remains guarded until secure storage, participant auth, MIME/size, retention, access control, virus/abuse review, and manual review tests exist.
- Phase 22 media policy: `docs/cbt-media-prompt-response-policy.md` records current image/audio prompt support, deferred video prompt, and `recording_answer_safe_status: policy_deferred`; recording upload must not be added outside the Phase 21 policy.

## Phase 23-26 follow-up status (2026-05-08)

- Phase 23 proctor evidence mode: `docs/cbt-proposal-integration-phase-23-26.md` records dashboard evidence coverage for heartbeat, app background/resume, device mismatch, submit guard, stale connection, warning, force submit, reset access, and export/print evidence. Screen preview and remote desktop remain explicit non-goals.
- Phase 24 BYOD anti-cheat evidence: `apps/mobile/DEVICE_TEST_MATRIX.md` and `apps/mobile/RELEASE_CHECKLIST.md` keep real-device evidence as `pending_manual_evidence` until tested by operator/pengawas on at least two Android vendors. Deterministic build/hash instructions are documented without claiming PASS.
- Phase 25 item analysis: current metrics are documented and tested through helper coverage; Cronbach alpha remains deferred until a real formula and fixed dataset are implemented and tested.
- Phase 26 reports: official template matrix maps PDF to print HTML/browser PDF and Excel to CSV-safe exports; sensitive token artifacts stay print-only/role-bound.
