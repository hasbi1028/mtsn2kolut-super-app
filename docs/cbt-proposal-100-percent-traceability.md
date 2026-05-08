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
| Question types: true_false | Web Admin / Core API / Flutter | Implemented runtime | Fixed pair `A=Benar`, `B=Salah`, label-compatible Flutter fallback | Not blocking | Phase 17 / Phase 19 |
| Question types: agree_disagree | Web Admin / Core API / Flutter | Implemented runtime | Fixed pair `A=Setuju`, `B=Tidak Setuju`, label-compatible Flutter fallback | Not blocking | Phase 17 / Phase 19 |
| Question types: hotspot | Flutter / Web Admin / Core API | Adapted/out-of-scope for v1 unless explicitly approved | Flutter unsupported guard; decision record required before schema/UI/scoring | Feature parity blocker only if user chooses hotspot v1 | Phase 20 |
| Question types: upload_answer / file_upload | Flutter / Web Admin / Core API / Ops | Adapted/out-of-scope for v1 unless storage policy approved | Flutter unsupported guard; storage/MIME/retention/manual review policy required | Feature parity blocker only if proposal file answer is mandatory | Phase 21 |
| Audio/video prompt/response | Web Admin / Flutter / Ops | Needs runtime implementation or adapted policy | Media prompt policy, storage and offline APK constraints | Deferred until policy | Phase 22 |
| Anti-cheat BYOD layers | Flutter / Core API / Ops | Implemented docs/evidence + partial runtime | Lifecycle telemetry, screenshot protection where OS allows, heartbeat, submit guard, audit event | Needs manual evidence on physical Android devices | Phase 24 |
| Proctoring dashboard | Web Admin / Core API | Implemented runtime + needs evidence mode | Room dashboard, participant events, suspicious flag, handover recap | Needs operator rehearsal evidence | Phase 23 / Phase 27 |
| Reports | Web Admin / Core API | Partial | Session results and item analysis exist; PDF/Excel parity still needs final export evidence | Blocking for full proposal reporting parity | Phase 26 |
| Analytics | Web Admin / Core API | Partial | Item analysis query, difficulty/discrimination/distribution | Needs UI/report completeness evidence | Phase 25 |
| Infrastructure/ops | Ops | Implemented docs/evidence | PM2 architecture, backup scripts, release preflight/readiness scripts | Production go still requires explicit deploy instruction and sign-off | Phase 28 / Phase 31 |
| Manual evidence/sign-off | Ops / Operator | Needs manual evidence | Device matrix, operator rehearsal, final go/no-go sign-off | Blocking for production go claim | Phase 24 / Phase 27 / Phase 32 |
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
