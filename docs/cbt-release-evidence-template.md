# CBT Release Evidence Template

Gunakan template ini untuk arsip operational readiness CBT. Template ini bukan instruksi deploy.

## Identitas Rilis

- Tanggal evidence:
- Operator:
- Reviewer:
- Branch:
- Commit:
- Phase baseline: Phase 5 commit `bec16c8` + Phase 6 evidence automation + Phase 8 manifest/checksum/secret-scan hardening + Phase 9 commit `bf9df69` post-deploy smoke/runbook hardening + Phase 10 handoff package.
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
