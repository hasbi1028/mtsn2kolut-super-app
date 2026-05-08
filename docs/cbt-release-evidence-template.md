# CBT Release Evidence Template

Gunakan template ini untuk arsip operational readiness CBT. Template ini bukan instruksi deploy.

## Identitas Rilis

- Tanggal evidence:
- Operator:
- Reviewer:
- Branch:
- Commit:
- Phase baseline: Phase 5 commit `bec16c8` + Phase 6 evidence automation + Phase 8 manifest/checksum/secret-scan hardening + Phase 9 commit `bf9df69` post-deploy smoke/runbook hardening + Phase 10 handoff package + Phase 11 archive/retention rules after Phase 10 handoff + Phase 12 evidence index/retrieval policy after Phase 11 archive/retention + Phase 13 Mobile Release Candidate and Device Matrix + Phase 14 Operator Rehearsal and Proctor Evidence + Phase 15 Final CBT Release Readiness Sign-off.
- Final baseline marker:
  - `CBT Phase 15 final baseline`
- Keputusan rilis: lanjut / tunda / rehearsal ulang.

## Boundary Wajib

- No deploy, no PM2 restart, no `make db-migrate`.
- No live migration dan no ad hoc SQL.
- Tidak menjalankan `psql` untuk `ALTER`, `UPDATE`, `DELETE`, `INSERT`, atau DDL/DML lain.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter runtime siswa memakai `/api/exam/*`.
- Tidak ada runtime siswa melalui `/api/cbt/login` atau `/api/cbt/status`.
- `apps/web-admin` tetap BFF/proxy.
- `services/core-api` tetap owner PostgreSQL.
- `apps/mobile` tetap BYOD Flutter client; tidak ada klaim kiosk penuh.

## Output Preflight

- Command:
  - `deploy/scripts/cbt-release-preflight.sh --output <path>`
- Output directory:
- Markdown evidence:
  - `cbt-release-preflight.md`
- JSON evidence:
  - `cbt-release-preflight.json`
- Manifest/checksum evidence:
  - `cbt-release-manifest.json`
- Log directory:
  - `logs/`
- Catatan secret hygiene:
  - output tidak boleh memuat password, token mentah, API key, answer key, atau secret lain.
  - secret scan preflight harus `pass`; jika `fail`, bundle menjadi blocker sampai evidence dibersihkan dan dibuat ulang.

## Final audit/readiness artifacts

Gunakan bagian ini setelah final audit proposal dan final readiness evidence dibuat. Bagian ini membedakan automated evidence completed dari manual evidence requiring physical Android devices and operator rehearsal.

- Proposal gap audit:
  - `docs/cbt-proposal-gap-audit.md`
- Final release evidence:
  - `docs/cbt-release-final-evidence.md`
- Final readiness command:
  - `deploy/scripts/cbt-final-readiness.sh --output <path>`
- Generated final readiness artifacts:
  - `cbt-proposal-gap-audit.json`
  - `cbt-final-evidence.json`
  - `cbt-final-signoff.json`
  - `cbt-final-readiness.md`
- automated evidence completed:
  - current git commit/branch:
  - docs existence:
  - `git diff --check`:
  - tests manifest:
  - safe health command status:
  - generated evidence secret scan:
- manual evidence requiring physical Android devices and operator rehearsal:
  - device matrix status:
  - operator rehearsal status:
  - final sign-off status:
  - minimum two Android vendors:
  - proctor evidence:
  - role/scope/token boundary:
  - event/audit evidence:
  - go/no-go:
- Decision hygiene:
  - default generated sign-off must be `pending_manual_signoff` when manual evidence is missing.
  - generated readiness may use `ready_for_rehearsal` only when operator flags manual evidence complete.
  - generated readiness must not claim production go from this host alone.

## Commit dan Status Git

- `git rev-parse HEAD`:
- `git status --short --branch`:
- Working tree bersih: ya / tidak.
- Jika tidak bersih, file yang relevan:

## Tool Availability

- `git`:
- `node`:
- `npm`:
- `go`:
- Flutter SDK path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- `flutter`:

## Preflight Checks

| Check | Command | Status | Log |
|-------|---------|--------|-----|
| Diff whitespace | `git diff --check` | pass / fail | |
| Docs guard | `cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts` | pass / fail / skipped | |
| Web check | `cd apps/web-admin && npm run check` | pass / fail / skipped | |
| Go test | `cd services/core-api && go test ./...` | pass / fail / skipped | |
| Go build | `cd services/core-api && go build -o /dev/null ./cmd/api` | pass / fail / skipped | |
| Flutter doctor | `cd apps/mobile && flutter doctor` | pass / fail / skipped | |
| Flutter analyze | `cd apps/mobile && flutter analyze` | pass / fail / skipped | |
| Flutter test | `cd apps/mobile && flutter test` | pass / fail / skipped | |
| Secret scan | generated markdown/json/log evidence | pass / fail | |

## Manifest dan Checksums

- `cbt-release-manifest.json` tersedia: ya / tidak.
- Manifest parseable sebagai JSON: ya / tidak.
- Algorithm: `sha256`.
- Artifact wajib tercatat:
  - `cbt-release-preflight.md`
  - `cbt-release-preflight.json`
  - `logs/git-head.log`
  - `logs/git-status.log`
  - `logs/git-diff-check.log`
- Semua checksum berbentuk SHA-256 64 hex: ya / tidak.
- Secret scan status di manifest: pass / fail.

## Mobile Evidence

- APK/release candidate identifier:
- RC identifier:
- APK SHA-256 hash:
- signing mode:
- `API_BASE_URL`:
- Flutter SDK tersedia di host evidence: ya / tidak.
- Flutter SDK absolute path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- `flutter analyze`: pass / fail / skipped.
- `flutter test`: pass / fail / skipped.
- Device matrix updated:
  - `apps/mobile/DEVICE_TEST_MATRIX.md`
- Minimal dua perangkat nyata lintas vendor: pass / fail / belum diuji.
- minimum two Android vendors: pass / fail / belum diuji.
- background/resume:
- heartbeat:
- pending answer:
- submit guard:
- device mismatch:
- screenshot protection / `FLAG_SECURE`:
- network disturbance:
- BYOD risks observed:

## Phase 13 Mobile Release Candidate and Device Matrix

Isi bagian ini untuk APK final candidate sebelum operator rehearsal.

- RC identifier:
- APK SHA-256 hash:
- signing mode: release keystore / debug signing untuk uji teknis internal saja.
- `API_BASE_URL`:
- Flutter SDK absolute path:
  - `/home/servermtsn2kolut/development/flutter/bin`
- Device matrix file:
  - `apps/mobile/DEVICE_TEST_MATRIX.md`
- minimum two Android vendors: pass / fail / belum diuji.
- background/resume: pass / fail / belum diuji.
- heartbeat: pass / fail / belum diuji.
- pending answer: pass / fail / belum diuji.
- submit guard: pass / fail / belum diuji.
- device mismatch: pass / fail / belum diuji.
- screenshot protection / `FLAG_SECURE`: pass / fail / belum diuji.
- network disturbance: pass / fail / belum diuji.
- RC accepted for rehearsal: ya / tidak.
- Notes:

## Backend dan Web Evidence

- `/api/exam/*` contract checked:
  - `POST /api/exam/login`
  - `GET /api/exam/status`
  - `POST /api/exam/heartbeat`
  - `POST /api/exam/event`
  - `POST /api/exam/answer`
  - `POST /api/exam/submit`
- Public `/api/cbt/**` route tree baru: tidak ada / ada.
- Web-admin docs guard: pass / fail.
- Web-admin `npm run check`: pass / fail / skipped.
- Core API `go test ./...`: pass / fail / skipped.
- Core API build check: pass / fail / skipped.

## Data dan Operasional

- Backup PostgreSQL terbaru tercatat: ya / tidak.
- Migration state dicek read-only: ya / tidak.
- Tidak ada migrasi live dari Phase 6: ya / tidak.
- Audit log tersedia untuk role berwenang: ya / tidak.
- Server log dicek untuk panic berulang dan secret leakage: ya / tidak.
- Temuan:

## Phase 9/10 Post-Deploy Handoff

Isi bagian ini setelah deploy manual selesai, post-deploy smoke lulus, dan evidence bundle siap diarsipkan. Bagian ini adalah Phase 10 handoff package, bukan instruksi deploy.

- Phase 9 commit `bf9df69` smoke/runbook baseline diverifikasi: ya / tidak.
- Commit hash deployed:
- Branch:
- Evidence timestamp:
- Operator:
- Reviewer:
- Rollback owner:
- Follow-up owner:
- Keputusan handoff: lanjut / tunda / rollback.

PM2 status summary placeholders:

| Target | Process name | Status | Restarts | Uptime | Notes |
|--------|--------------|--------|----------|--------|-------|
| Backend VPS | | | | | |
| Frontend VPS | | | | | |
| Worker VPS | | | | | |

Health smoke outputs:

- `make ops-health`:
- `make ops-health-backend`:
- `make ops-health-frontend`:
- `make ops-health-worker`:
- `bash deploy/scripts/health-check.sh bank-soal`:

Manifest/checksum evidence:

- `cbt-release-preflight.md` reviewed: ya / tidak.
- `cbt-release-preflight.json` parseable: ya / tidak.
- `cbt-release-manifest.json` reviewed: ya / tidak.
- SHA-256 checksum evidence complete: ya / tidak.
- Secret scan status: pass / fail.

Secret hygiene:

- Tidak ada password/token/API key/worker key/answer key/JWT mentah di handoff: ya / tidak.
- Tidak ada DB DSN, `.env`, `printenv`, PM2 env dump, atau full request header di handoff: ya / tidak.
- Semua host credential dan bearer value tetap redacted: ya / tidak.

## Phase 11 Evidence Archive and Retention

Isi bagian ini setelah Phase 10 handoff package lengkap dan evidence bundle siap dipindahkan dari staging ke arsip. Phase 11 archive/retention rules after Phase 10 handoff adalah dokumentasi ops/read-only, bukan deploy automation.

- Ops-controlled archive path:
- Not committed tmp: ya / tidak.
- Archive path convention:
  - `<ops-controlled-storage>/cbt-evidence/<yyyy>/<yyyy-mm-dd>-<release-id>-<short-commit>/`
- Archive created by:
- Archive reviewed by:
- Archive timestamp:
- Release identifier:
- Short commit:

Required artifacts:

- `cbt-release-preflight.md`: ada / tidak.
- `cbt-release-preflight.json`: ada / tidak.
- `cbt-release-manifest.json`: ada / tidak.
- `logs/git-head.log`: ada / tidak.
- `logs/git-status.log`: ada / tidak.
- `logs/git-diff-check.log`: ada / tidak.
- selected optional validation logs actually run: ada / tidak / skipped.
- completed release evidence template copy: ada / tidak.
- Phase 10 handoff package: ada / tidak.

SHA-256 manifest verification:

- Manifest parseable as JSON: ya / tidak.
- Algorithm is `sha256`: ya / tidak.
- Required artifacts listed: ya / tidak.
- SHA-256 values are 64 lowercase hex characters: ya / tidak.
- Recomputed checksums match archived files: ya / tidak.
- Secret scan status is `pass`: ya / tidak.
- Verification notes:

Retention owner:

- Owner:
- Reviewer:
- Retention period placeholder:
- Review date:
- Expiry date placeholder:
- Hold/exception reason:

Redaction/no secrets:

- No password/token/API key/worker key/answer key/JWT mentah: ya / tidak.
- No DB DSN, `.env`, `printenv`, PM2 env dump, full request header, or database dump: ya / tidak.
- Bearer values and host credentials redacted: ya / tidak.
- Contaminated artifact rejected or rebuilt if found: ya / tidak / n/a.

Access control:

- Owner group:
- Allowed readers:
- Allowed writers:
- Public web exposure checked: ya / tidak.
- Repository commit exposure checked: ya / tidak.
- Archive is outside public web roots and Git-tracked tmp paths: ya / tidak.

Restore/read-back check:

- Archive opened from a fresh shell/session: ya / tidak.
- `cbt-release-manifest.json` readable: ya / tidak.
- SHA-256 recomputed for markdown, JSON, and at least one log: ya / tidak.
- Completed evidence template readable without live service access: ya / tidak.
- Phase 10 handoff package matches release identifier and commit: ya / tidak.
- Read-back result: pass / fail.
- Notes:

Deletion/expiry log placeholder:

| Date | Archive path | Action | Authorized by | Performed by | Reason | SHA-256 manifest checked | Notes |
|------|--------------|--------|---------------|--------------|--------|--------------------------|-------|
| | | review / extend / delete | | | | ya / tidak | |

## Phase 12 Evidence Index and Retrieval

Isi bagian ini setelah Phase 11 archive/retention record tersedia dan evidence perlu ditemukan kembali melalui redacted index. Phase 12 evidence index/retrieval policy after Phase 11 archive/retention adalah dokumentasi ops/read-only, bukan search service atau deploy automation.

- Index storage location:
- Outside public web roots: ya / tidak.
- Outside Git-tracked tmp/generated evidence paths: ya / tidak.
- Not committed tmp data except template docs: ya / tidak.
- Index owner:
- Reviewer:

Index fields:

| Release id | Commit | Date | Archive path | Manifest SHA-256 | Owner | Retention status | Access classification | Notes |
|------------|--------|------|--------------|------------------|-------|------------------|-----------------------|-------|
| | | | | | | active / hold / under review / expired / stale / missing | restricted / confidential / other | |

Redaction/search boundary:

- Search uses redacted metadata only: release id, commit, date, owner, retention status, access classification, manifest SHA-256.
- No raw secrets: ya / tidak.
- No full logs: ya / tidak.
- No env, `.env`, `printenv`, PM2 env dump, database dump, or full request headers: ya / tidak.
- No password/token/API key/worker key/answer key/JWT mentah: ya / tidak.
- Archive path is redacted or ops-local, not a public URL: ya / tidak.

Lookup/read-back workflow:

- Index lookup by release id / commit / date:
- Owner and retention status checked: ya / tidak.
- Access classification checked: ya / tidak.
- Archive opened read-only from ops-controlled storage: ya / tidak.
- Minimum evidence needed identified: ya / tidak.
- Read-back result:
- Notes:

Checksum verification before retrieval:

- Recomputed `cbt-release-manifest.json` SHA-256:
- Matches index Manifest SHA-256: ya / tidak.
- Manifest parseable as JSON: ya / tidak.
- Manifest algorithm is `sha256`: ya / tidak.
- Required artifact checksums recomputed before retrieval: ya / tidak.
- Secret scan status is `pass`: ya / tidak.
- Retrieval allowed: ya / tidak.

Access request/audit placeholders:

| Date | Release id | Commit | Requester | Purpose | Owner approval | Reviewer | Result | Notes |
|------|------------|--------|-----------|---------|----------------|----------|--------|-------|
| | | | | | approved / rejected / pending | | retrieved / denied / stale / missing | |

Stale/missing archive handling:

- Index row status: active / stale / missing / hold / expired.
- Archive path exists: ya / tidak.
- Manifest SHA-256 matches index: ya / tidak.
- Retrieval restricted pending owner review: ya / tidak / n/a.
- Owner follow-up:
- Resolution notes:

## Phase 14 Operator Rehearsal and Proctor Evidence

Isi bagian ini setelah rehearsal operator final selesai. Phase 14 Operator Rehearsal and Proctor Evidence adalah bukti manual/read-only, bukan instruksi deploy.

- Operator flow:
  - Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review.
- proctor evidence:
- role/scope/token boundary:
- event/audit evidence:
- go/no-go rehearsal:
- operator:
- pengawas:
- reviewer:
- rollback owner:
- blockers:
- accepted operational notes:
- follow-up owner:

Operator flow evidence:

- Bank Soal authoring/import/review captured: ya / tidak.
- Asesmen Persiapan paket/kegiatan/sesi setup captured: ya / tidak.
- Pelaksanaan/Pengawasan heartbeat/status/warning captured: ya / tidak.
- Flutter APK login, answer save, restore, warning, and submit captured: ya / tidak.
- Hasil/Post-exam review captured: ya / tidak.

Proctor evidence:

- heartbeat and connection status captured: ya / tidak.
- background/resume warning/event captured: ya / tidak.
- pending answer or network disturbance evidence captured: ya / tidak.
- submit guard and device mismatch evidence captured: ya / tidak.
- screenshot protection / `FLAG_SECURE` checked as deterrence evidence: ya / tidak.
- evidence redacted for token/password/API key/JWT/answer key: ya / tidak.

Role/scope/token boundary:

- admin/panitia token authority checked: ya / tidak.
- guru subject/class scope checked: ya / tidak.
- pengawas/proktor session/room scope checked: ya / tidak.
- token visibility restricted to authorized operational contexts: ya / tidak.
- answer key/rubric visibility restricted: ya / tidak.
- Flutter student runtime remains `/api/exam/*`: ya / tidak.

Event/audit evidence:

- Bank Soal audit/event:
- Asesmen Persiapan audit/event:
- Pelaksanaan/Pengawasan audit/event:
- Flutter APK exam event:
- Hasil/Post-exam audit/event:

## Phase 15 Final CBT Release Readiness Sign-off

Isi bagian ini sebagai keputusan akhir setelah Phase 13 mobile RC evidence dan Phase 14 rehearsal evidence lengkap.

- Final sign-off summary Phase 0-15:
- go/no-go:
- decision date/time WITA:
- operator:
- reviewer:
- rollback owner:
- follow-up owner:
- evidence bundle:
- validation commands:
- final baseline marker:
  - `CBT Phase 15 final baseline`

Phase 0-15 summary:

| Phase | Evidence status | Notes |
|-------|-----------------|-------|
| Phase 0 architecture alignment and guard | pass / fail / n/a | |
| Phase 1 exam API contract stabilization | pass / fail / n/a | |
| Phase 2 Web Admin workflow alignment | pass / fail / n/a | |
| Phase 3 backend runtime hardening | pass / fail / n/a | |
| Phase 4 Flutter BYOD anti-cheat and resilience | pass / fail / n/a | |
| Phase 5 rehearsal, rollout, and post-exam review | pass / fail / n/a | |
| Phase 6 production readiness evidence automation | pass / fail / n/a | |
| Phase 7 release preflight verifier/test hardening | pass / fail / n/a | |
| Phase 8 manifest/checksum/secret-scan hardening | pass / fail / n/a | |
| Phase 9 live deploy smoke/runbook hardening | pass / fail / n/a | |
| Phase 10 post-deploy evidence handoff | pass / fail / n/a | |
| Phase 11 evidence archive and retention | pass / fail / n/a | |
| Phase 12 evidence index and retrieval | pass / fail / n/a | |
| Phase 13 mobile release candidate and device matrix | pass / fail / n/a | |
| Phase 14 operator rehearsal and proctor evidence | pass / fail / n/a | |
| Phase 15 final CBT release readiness sign-off | pass / fail / n/a | |

Evidence bundle:

- completed evidence template: ada / tidak.
- `cbt-release-preflight.md`: ada / tidak.
- `cbt-release-preflight.json`: ada / tidak.
- `cbt-release-manifest.json`: ada / tidak.
- optional validation logs actually run: ada / tidak / skipped.
- Phase 13 device matrix: ada / tidak.
- Phase 14 proctor evidence: ada / tidak.
- archive path:
- manifest SHA-256:
- redacted index lookup reference:

Validation commands:

```bash
git diff --check
cd apps/web-admin && npm run test:unit -- src/lib/cbt/proposal-integration-docs.test.ts
cd apps/web-admin && npm run check
cd apps/web-admin && npm run test:unit
make ops-health
cd services/core-api && go test ./...
cd services/core-api && go build -o /dev/null ./cmd/api
cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter analyze
cd apps/mobile && /home/servermtsn2kolut/development/flutter/bin/flutter test
```

Final sign-off:

- go/no-go:
- rollback owner:
- accepted risks:
- blocker list:
- final baseline marker present: ya / tidak.

## Phase 19-22 Fixed Pair and High-Risk Question Decisions

Isi bagian ini saat review proposal question-type parity setelah baseline `cebd789`.

- Phase 19 fixed-pair evidence:
  - `true_false` labels `A=Benar`, `B=Salah`: pass / fail / n/a.
  - `agree_disagree` labels `A=Setuju`, `B=Tidak Setuju`: pass / fail / n/a.
  - Flutter fallback submits Web Admin/Core API labels: pass / fail / n/a.
  - backend scoring compatibility for old `true`/`false` true_false cache: pass / fail / n/a.
- Phase 20 hotspot decision:
  - `hotspot_safe_status: adapted_deferred`.
  - no fake hotspot runtime: pass / fail / n/a.
  - design doc: `docs/cbt-hotspot-design-decision.md`.
- Phase 21 upload/file answer policy:
  - `upload_answer_safe_status: policy_deferred`.
  - policy doc: `docs/cbt-upload-answer-policy.md`.
  - no upload runtime enabled without secure storage/auth/MIME/size/retention/review tests: pass / fail / n/a.
- Phase 22 media prompt/response policy:
  - existing image/audio prompt fields reviewed: pass / fail / n/a.
  - `recording_answer_safe_status: policy_deferred`.
  - policy doc: `docs/cbt-media-prompt-response-policy.md`.
  - no recording upload enabled outside upload/file answer policy: pass / fail / n/a.

## Phase 23-26 Evidence, Analytics, and Reports

Isi bagian ini setelah evidence mode, BYOD device evidence, analytics, dan report matrix direview.

Phase 23 proctor evidence mode:

- heartbeat captured in dashboard/evidence CSV: pass / fail / pending.
- app background/resume captured: pass / fail / pending.
- device mismatch `409` evidence captured without exposing full fingerprint: pass / fail / pending.
- submit guard warning captured: pass / fail / pending.
- stale connection warning captured: pass / fail / pending.
- warning telemetry captured: pass / fail / pending.
- force submit action evidence captured: pass / fail / pending.
- reset access action evidence captured: pass / fail / pending.
- export/print evidence captured: pass / fail / pending.
- no screen preview / no remote desktop claim: ya / tidak.

Phase 24 anti-cheat BYOD evidence:

- `anti_cheat_byod_manual_status`: pending_manual_evidence / complete_by_operator.
- real-device PASS claimed only after physical Android operator test: ya / tidak.
- minimum two Android vendors tested: ya / tidak.
- APK SHA-256 hash:
- `API_BASE_URL`:

Phase 25 analytics/item analysis:

- difficulty index: pass / fail / n/a.
- discrimination index with sample guard: pass / fail / n/a.
- distractor / answer distribution: pass / fail / n/a.
- unanswered count: pass / fail / n/a.
- per-type accuracy: pass / fail / n/a.
- Cronbach alpha remains documented/deferred unless real tested implementation exists: ya / tidak.

Phase 26 report template matrix:

- official template matrix reviewed: pass / fail / pending.
- print HTML / browser PDF path available where safe: pass / fail / pending.
- CSV Excel-compatible path available where safe: pass / fail / pending.
- no broad token spreadsheet export: ya / tidak.
- no new binary PDF/XLSX endpoint introduced without approved safe pattern: ya / tidak.

## Acceptance Decision

Lanjut bila:

- evidence bundle markdown/json/manifest tersedia.
- manifest memuat checksum SHA-256 untuk markdown/json/log evidence.
- secret scan generated evidence lulus.
- command yang dipilih operator lulus atau skipped dengan alasan tertulis.
- Flutter mobile final tidak dianggap lulus jika SDK/check mobile belum tersedia.
- tidak ada pelanggaran no deploy/no restart/no migration boundary.
- runtime siswa tetap `/api/exam/*`.

Tunda bila:

- command wajib gagal.
- evidence memuat secret/token/password/answer key.
- migration state tidak jelas.
- public route tree `/api/cbt/**` baru muncul untuk runtime siswa.
- Flutter tidak dapat membuktikan login, heartbeat, answer save, restore, warning, dan submit pada perangkat nyata.

## Tindak Lanjut

- Owner:
- Due date:
- Catatan panitia/operator:
