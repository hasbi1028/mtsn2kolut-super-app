# CBT Upload/File Answer Policy

Status: Phase 21 design policy, runtime deferred, 2026-05-08.

```text
upload_answer_safe_status: policy_deferred
```

No upload-answer runtime is enabled by this policy alone. `upload_answer` and `file_upload` stay guarded/deferred until the full storage, auth, MIME, size, retention, and review path can be implemented and tested safely.

## Policy Requirements

Before any runtime upload/file answer work, the implementation must define:

- max size: conservative per-file and per-exam-participant limits, enforced before storage.
- allowed MIME: explicit allowlist, extension match, content sniffing, and rejection of scriptable content.
- storage location: backend-owned non-public path, no direct SvelteKit or Flutter filesystem writes, no PostgreSQL direct access outside Core API.
- retention: retention period, cleanup owner, evidence archive handling, and deletion log.
- access control: participant may upload only for their own live exam context; reviewers may read only through authorized Web Admin/Core API paths.
- virus/abuse review: quarantine or operator review procedure for suspicious files, failed MIME checks, oversized files, and repeated abuse.
- manual review: scoring and rubric workflow for file answers, with audit trail and redacted evidence handling.

## Runtime Gate

Runtime may only proceed after tests cover:

- token/fingerprint participant scope under `/api/exam/*`.
- max size and MIME rejection.
- safe filename/storage path generation.
- `X-Content-Type-Options: nosniff` and safe disposition on reviewer downloads.
- manual review authorization and audit events.
- Flutter pending upload, retry, failure, and restored-session behavior.

Phase 21 does not add upload endpoints, migrations, storage directories, Flutter file picker UI, or reviewer download UI.

## Boundary

- No public `/api/cbt/**` upload route.
- No PocketBase, SQLite, or Alpine runtime.
- No schema migration, live DB write, deploy, or PM2 restart.
- No recording upload unless it satisfies this policy with stricter audio/video MIME, size, retention, and review controls.
