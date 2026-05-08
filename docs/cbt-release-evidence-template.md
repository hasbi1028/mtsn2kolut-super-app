# CBT Release Evidence Template

Gunakan template ini untuk arsip operational readiness CBT. Template ini bukan instruksi deploy.

## Identitas Rilis

- Tanggal evidence:
- Operator:
- Reviewer:
- Branch:
- Commit:
- Phase baseline: Phase 5 commit `bec16c8` + Phase 6 evidence automation + Phase 8 manifest/checksum/secret-scan hardening + Phase 9 commit `bf9df69` post-deploy smoke/runbook hardening + Phase 10 handoff package + Phase 11 archive/retention rules after Phase 10 handoff.
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
- Flutter SDK tersedia di host evidence: ya / tidak.
- `flutter analyze`: pass / fail / skipped.
- `flutter test`: pass / fail / skipped.
- Device matrix updated:
  - `apps/mobile/DEVICE_TEST_MATRIX.md`
- Minimal dua perangkat nyata lintas vendor: pass / fail / belum diuji.
- BYOD risks observed:

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
