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

Current sync point: AGENTS.md Review Sync, 2026-05-05. All completed sprints (17–96) are now reflected in AGENTS.md. `/cbt/soal` is the active Bank Soal UI; `/cbt/questions` is legacy redirect only. CBT navigation uses role-based hubs: `/cbt/persiapan`, `/cbt/pelaksanaan`, `/cbt/hasil`. Kesiswaan, Tata Usaha, Governance, Document Cycles, Jurnal Kelas, Jadwal, Inventory, dan Non-test Assessments sudah aktif.
