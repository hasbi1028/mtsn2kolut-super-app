# Worker Policy

This service is the worker runtime for asynchronous job execution.

## Role

- Claim jobs from the Go API
- Execute worker tasks safely
- Report completion, failure, and retry state back to the Go API

## Must Keep

- Never access PostgreSQL directly.
- Never become a second owner of business state.
- Keep communication direct to `services/core-api`.
- Keep concurrency configurable, but prefer safe defaults.

## Implementation Direction

- Favor operational robustness over feature sprawl.
- Keep retries, claim loops, and failure reporting explicit.
- Avoid coupling worker internals to frontend code or UI concerns.
