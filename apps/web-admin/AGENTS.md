# Web Admin Policy

This app is the SvelteKit admin frontend for MTs Negeri 2 Kolaka Utara.

## Role

- Provide admin and guru-facing UI.
- Act as a thin BFF/proxy to the Go API where needed.
- Own cookies, session handling, and frontend UX.

## Current Baseline — 2026-05-06

- `/bank-soal` is the active standalone Bank Soal list/landing route.
- `/bank-soal/tambah`, `/bank-soal/verifikasi`, and `/bank-soal/impor` own authoring, review/verification, and import UI.
- `/bank-soal/komposer`, `/bank-soal/review`, `/bank-soal/import`, `/cbt/soal*`, `/cbt/bank-soal*`, and `/cbt/questions*` are compatibility redirects to `/bank-soal/*`.
- `/asesmen` is the active assessment launcher; legacy `/cbt*` operational routes redirect to `/asesmen*` or Bank Soal according to context.
- Active frontend fetches use `/api/bank-soal/*` for Bank Soal and `/api/asesmen/*` for assessment. `/api/cbt/*` routes remain live deprecated compatibility proxies to the same Go API semantics for legacy clients, redirects, proxy tests, and staged transition safety.
- Library and Inventory paths are gated for `admin` and `staf` in hooks/route-access helpers.
- Kesiswaan and staff operation gates in SvelteKit are UX/BFF guards; backend role gates remain the source of truth.

## Must Keep

- Do not move backend business rules into SvelteKit.
- Do not introduce direct database access.
- Do not reintroduce SQLite or Drizzle as runtime ownership.
- Prefer server routes only as thin adapters to backend endpoints.
- Keep `/api/*` routes as BFF/proxy adapters only; no direct PostgreSQL, SQLite, or Drizzle runtime storage.

## UI Direction

- Use `shadcn-svelte` for new UI primitives where practical.
- Design for a school administration context, not a startup SaaS dashboard.
- Favor calm, credible, education-appropriate visuals.
- Use clear hierarchy, strong legibility, and practical workflows for admin and guru users.
- Avoid noisy novelty and avoid generic AI-looking templates.

## Design System (Kemenag Green Theme)

- **Typography**: Plus Jakarta Sans (Google Fonts). Weights: 400, 600, 700, 800, 900.
- **Primary color**: `#16a34a` (Kemenag green). Hover: `#15803d`.
- **Background**: `#f8fafc` / `white`. Surface: `#f1f5f9`.
- **Text**: `#0f172a` (primary), `#64748b` (secondary), `#94a3b8` (muted).
- **Cards**: Border radius `18px`, border `1.5px solid #f1f5f9`.
- **Buttons**: Border radius `14px`, uppercase text, `font-weight: 900`.
- **Inputs**: Border radius `14px`, green focus ring `rgba(22,163,74,0.1)`.
- **Badges/Kickers**: `0.65rem`, `font-weight: 900`, `uppercase`, `letter-spacing: 0.18em`.
- See skill `kemenag-superapp-theme` for full design tokens and patterns.

## UX Direction

- Prioritize speed of data entry and clarity of status.
- Tables, forms, schedule views, and monitoring screens should feel dependable and easy to scan.
- Important states must be obvious: active, inactive, draft, published, scheduled, running, success, failed.
- Mobile support must remain acceptable even for admin pages.
- Bank Soal authoring should remain compact and mode-based so teachers are not overwhelmed.

## Technical Direction

- Keep pages and components simple to reason about.
- Reuse shared UI primitives before adding one-off patterns.
- Validate Svelte files before finalizing edits.
- Keep unit coverage around retired route redirects and BFF stream/proxy contracts when changing Bank Soal or assessment routes.
