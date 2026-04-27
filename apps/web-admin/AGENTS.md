# Web Admin Policy

This app is the SvelteKit admin frontend for MTs Negeri 2 Kolaka Utara.

## Role

- Provide admin and guru-facing UI.
- Act as a thin BFF/proxy to the Go API where needed.
- Own cookies, session handling, and frontend UX.

## Must Keep

- Do not move backend business rules into SvelteKit.
- Do not introduce direct database access.
- Do not reintroduce SQLite or Drizzle as runtime ownership.
- Prefer server routes only as thin adapters to backend endpoints.

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

## Technical Direction

- Keep pages and components simple to reason about.
- Reuse shared UI primitives before adding one-off patterns.
- Validate Svelte files before finalizing edits.
