# CBT Proposal Integration Phase 27-30 - Rehearsal, DR, Security, and Mobile RC Evidence

Status: Phase 27-30 scoped implementation record, 2026-05-08.

This phase continues from Phase 23-26 and stays in docs, evidence, checklist, and safe guard scope. It does not deploy, restart PM2, run live migrations, perform live DB writes, restore over the live database, create public `/api/cbt/**` runtime routes, or claim production go from repository checks alone.

## Phase 27 - Operator Rehearsal Workflow Completion

Phase 27 completes the operator rehearsal evidence checklist. The required flow is:

```text
Bank Soal to Asesmen Persiapan to Pelaksanaan/Pengawasan to Flutter APK to Hasil/Post-exam review
```

manual evidence required:

- operator rehearsal evidence bundle location.
- operator, pengawas, reviewer, rollback owner, and follow-up owner.
- Bank Soal authoring/import/review evidence.
- Asesmen Persiapan paket/kegiatan/sesi setup evidence.
- Pelaksanaan/Pengawasan proctor evidence.
- Flutter APK login, heartbeat, answer save, restore, warning, and submit evidence.
- Hasil/Post-exam review evidence.
- role/scope/token boundary evidence for admin/panitia, guru, pengawas/proktor, and siswa.
- event/audit evidence for each major workflow step.
- go/no-go rehearsal decision, blockers, accepted risks, and follow-up actions.

Evidence remains manual until operator/pengawas attach the actual rehearsal record. Repository docs or generated templates may mark the checklist as ready to fill, but may not mark a live rehearsal PASS by themselves.

## Phase 28 - Infrastructure, Backup, Restore, and DR Evidence

Phase 28 records infrastructure backup/restore/DR evidence without mutating live services.

Required evidence:

- backup path.
- latest symlink path.
- backup timestamp and file size.
- checksum file and computed SHA-256.
- `sha256sum -c` result when a checksum sidecar exists.
- `pg_restore --list` result for PostgreSQL custom-format dumps when `pg_restore` is available.
- restore rehearsal owner, target, RTO/RPO notes, and last successful drill date.
- DR contact list and decision owner.
- `make ops-health` output if explicitly run as read-only evidence.

Allowed verification:

```bash
sha256sum -c /path/to/latest.dump.sha256
pg_restore --list /path/to/latest.dump
```

Do not restore over live DB. Any full restore rehearsal must use an isolated scratch database or offline restore target with an explicit name that cannot be mistaken for production. This phase does not run `createdb`, `dropdb`, `psql`, `pg_restore --dbname`, migrations, PM2 commands, or live SQL writes.

## Phase 29 - Security and ISO-Control Alignment Evidence

Phase 29 records security and ISO-control alignment evidence. It is control alignment, not certification.

No ISO certification claim is made by this repository. External ISO 27001, ISO 9001, or similar certification would require a formal scope, management approval, auditor process, evidence sampling, and certificate issued by an accredited body.

Control-alignment evidence to attach:

- RBAC controls and role/scope tests.
- JWT session, refresh, revoke, and per-user auth-version behavior.
- rate limit and trusted proxy configuration evidence.
- token/device binding behavior for exam login/status.
- audit trail and auth audit event evidence.
- generic 500 and upload/file hygiene evidence.
- evidence redaction and generated evidence secret scan.
- backup/restore and DR verification evidence.
- mobile BYOD limitation statement and no kiosk guarantee.
- incident follow-up owner and residual-risk notes.

Security evidence must avoid raw passwords, JWTs, exam tokens, API keys, worker keys, answer keys, `.env`, PM2 env dumps, DB DSNs, full request headers, or unredacted student-sensitive dumps.

## Phase 30 - Mobile RC Build and Release Package Evidence

Phase 30 records the Flutter mobile release candidate package.

Required RC package fields:

- RC identifier.
- commit hash.
- version name/code.
- APK SHA-256 hash.
- APK path.
- APK file size.
- signing status.
- `API_BASE_URL`.
- Flutter SDK path and Flutter version when available.
- Android manifest checks for `INTERNET` and `android:allowBackup="false"`.
- quality gate status for `flutter analyze` and `flutter test`.
- build command and log path.
- device matrix status.

Build command:

```bash
cd apps/mobile
/home/servermtsn2kolut/development/flutter/bin/flutter build apk --release --dart-define=API_BASE_URL=https://api.sekolah.example
sha256sum build/app/outputs/flutter-apk/app-release.apk
```

`real-device PASS claimed only after physical Android operator test`. A successful APK build and SHA-256 hash do not prove device test PASS. Device evidence remains `pending_manual_evidence` until the APK is installed and tested on at least `minimum two Android vendors` by operator/pengawas.

## Boundary Wajib

- Tidak deploy.
- Tidak PM2 restart.
- Tidak menjalankan `make db-migrate`.
- Tidak menjalankan migrasi live.
- Tidak menjalankan ad hoc SQL.
- Tidak mengubah product runtime.
- Tidak mengubah schema database atau migration.
- Tidak melakukan live DB writes.
- Tidak restore over live DB.
- Tidak membuat public SvelteKit route tree `/api/cbt/**` baru.
- Flutter tetap berbicara langsung ke `services/core-api` melalui `/api/exam/*`.
- `apps/web-admin` tetap UI/BFF dan tidak membaca PostgreSQL langsung.
- `services/core-api` tetap owner PostgreSQL, token, timer, submit, scoring, audit, dan event exam.

## Acceptance Criteria

Phase 27-30 diterima bila:

- operator rehearsal checklist has explicit manual evidence slots and cannot be marked PASS by generated docs alone.
- backup verification records checksum and `pg_restore --list` status when a dump is provided, without restoring over live DB.
- DR evidence records RTO/RPO, owner, and last drill status without claiming a drill that did not happen.
- security evidence is framed as ISO-control alignment, not ISO certification.
- mobile RC evidence records build status, APK SHA-256 hash if available, signing status, `API_BASE_URL`, and device matrix status.
- physical Android device PASS remains forbidden until operator/pengawas test at least two Android vendors.
- no deploy, PM2 restart, migration, live SQL write, live DB restore, public `/api/cbt/**` route, or secret-bearing evidence is introduced.
