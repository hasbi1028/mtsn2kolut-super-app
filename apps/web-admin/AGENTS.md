# Web Admin Policy

This app is the SvelteKit admin frontend for MTs Negeri 2 Kolaka Utara.

## Role

- Provide admin and guru-facing UI.
- Act as a thin BFF/proxy to the Go API where needed.
- Own cookies, session handling, and frontend UX.

## Current Baseline — 2026-05-03

- `/cbt/soal` is the only active Bank Soal UI.
- `/cbt/questions` is retired and must stay a redirect to `/cbt/soal`.
- BFF `/api/cbt/questions/*` routes remain thin proxies/stream proxies to Go API.
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

## UX Direction

- Prioritize speed of data entry and clarity of status.
- Tables, forms, schedule views, and monitoring screens should feel dependable and easy to scan.
- Important states must be obvious: active, inactive, draft, published, scheduled, running, success, failed.
- Mobile support must remain acceptable even for admin pages.
- CBT authoring should remain compact and mode-based so teachers are not overwhelmed.

## Technical Direction

- Keep pages and components simple to reason about.
- Reuse shared UI primitives before adding one-off patterns.
- Validate Svelte files before finalizing edits.
- Keep unit coverage around retired route redirects and BFF stream/proxy contracts when changing CBT routes.
