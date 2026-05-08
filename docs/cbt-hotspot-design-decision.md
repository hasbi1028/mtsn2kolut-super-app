# CBT Hotspot Design Decision

Status: formal Phase 20 decision, 2026-05-08.

Decision: adapted/out-of-scope for BYOD CBT v1.

```text
hotspot_safe_status: adapted_deferred
```

## Decision

Hotspot is not implemented as runtime in Phase 20. No fake hotspot runtime is allowed. A hotspot question needs more than an objective option label: it needs image geometry, answer shape, tolerance rules, review display, and secure authoring. Implementing a visual placeholder that submits a dummy label would overclaim proposal support and produce unreliable scoring.

Flutter keeps the existing Flutter unsupported guard for `hotspot`. If such a question appears in a package, the student sees a manual-supervisor panel and should call the pengawas. The app must not submit fabricated hotspot answers.

## Required Design Before Runtime

A later hotspot v1 must define and test:

- coordinate system: normalized image coordinates, orientation behavior, viewport scaling, and tap serialization.
- image asset requirement: backend-owned image asset, safe MIME, stable dimensions, and authenticated media fetch.
- scoring tolerance: exact zone shape, tolerance radius or polygon inclusion, rounding rules, and deterministic server-side scoring.
- authoring UI: zone creation, preview, validation, and reviewer visibility.
- Flutter runtime: image fit behavior, tap marker, local restore, pending answer sync, and accessibility fallback.
- result display: selected coordinate/zone, correct zone, and human-readable review copy.

## Phase 20 Guard

Phase 20 makes no schema migration in Phase 20, no Core API scoring change, no Web Admin hotspot editor, and no Flutter tap-coordinate capture. It records the formal adapted/deferred decision so release evidence can distinguish a deliberate BYOD-safe gap from an incomplete fake implementation.

## Boundary

- PostgreSQL remains owned only by `services/core-api`.
- Flutter remains on `/api/exam/*`.
- No public `/api/cbt/**` route tree is created.
- No deploy, PM2 restart, migration, live SQL write, or secret-bearing evidence is part of this decision.
