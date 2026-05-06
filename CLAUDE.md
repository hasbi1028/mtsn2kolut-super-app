# Claude Code Project Memory

Use the following project files as the source of truth for this repository:

@AGENTS.md
@docs/architecture.md
@docs/ui-guidelines.md
@docs/deployment.md
@docs/cbt-smoke-checklist.md
@docs/exam-api.md
@findings.md
@apps/web-admin/AGENTS.md
@services/core-api/AGENTS.md
@services/pusaka-worker/AGENTS.md

If two instructions conflict, prefer the more specific service-level file for the files being edited, while preserving the global architecture rules in `AGENTS.md`.

Current sync point: AGENTS.md Route Naming Sync, 2026-05-06. All completed sprints (17–96) are reflected in AGENTS.md. `/bank-soal`, `/bank-soal/tambah`, `/bank-soal/impor`, and `/bank-soal/verifikasi` are the active Bank Soal UI routes. Because the app had not been released yet, old user-facing Bank Soal/CBT compatibility routes (`/bank-soal/komposer`, `/bank-soal/import`, `/bank-soal/review`, `/cbt*`) and external BFF `/api/cbt/*` routes were removed. Assessment navigation uses role-based hubs under `/asesmen`, including `/asesmen/persiapan`, `/asesmen/pelaksanaan`, and `/asesmen/hasil`. Kesiswaan, Tata Usaha, Governance, Document Cycles, Jurnal Kelas, Jadwal, Inventory, dan Non-test Assessments sudah aktif.
