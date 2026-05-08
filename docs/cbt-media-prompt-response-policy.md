# CBT Media Prompt and Response Policy

Status: Phase 22 media policy, 2026-05-08.

## Audio/video prompt

Image and audio prompts are partially supported through the existing exam payload and Flutter renderer. The current mobile-critical fields are:

- `stem_media_url`
- `stimulus_media_url`
- `stem_audio_url`
- `stimulus_audio_url`

Audio prompt playback has a one-time play indicator in the exam shell so siswa and pengawas can see whether audio for a question has already been played. Audio fetch must remain authenticated with exam token and device fingerprint headers; token values must not be embedded in media URLs.

Video prompt is adapted/deferred for BYOD CBT v1. It should not become mandatory for exam payloads until format, size, offline/network risk, accessibility, and test coverage are approved.

## Response Recording

```text
recording_answer_safe_status: policy_deferred
```

Do not add recording upload unless covered by the upload/file answer policy. Audio/video answer recording is a high-risk upload-answer variant and must satisfy stricter MIME allowlists, max size, retention, access control, manual review, and abuse handling before runtime work.

## Test and Evidence Expectations

Phase 22 evidence should cover:

- prompt media fields remain stable in `docs/exam-api.md`.
- Flutter can render image/audio prompts without blanking the question shell.
- audio one-time play state is visible and restorable.
- video prompt and recording answers are not claimed as completed runtime.
- no public `/api/cbt/**` runtime route, schema migration, deployment, PM2 restart, or live SQL write is introduced.
