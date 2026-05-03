# UI Guidelines

## Product Tone

This product serves MTs Negeri 2 Kolaka Utara. The interface should feel:

- educational
- orderly
- trustworthy
- institutional without feeling cold
- practical for admin and guru workflows

## Current UI Baseline

As of 2026-05-03:

- CBT Bank Soal uses `/cbt/soal` as the active UI; `/cbt/questions` is a retired redirect only.
- CBT authoring should stay compact, teacher-friendly, and mode-based (`Pemula` / `Advance`).
- CBT operational screens should prioritize readiness, role visibility, token safety, proctor action clarity, and print-friendly artifacts.
- BYOD helper surfaces should use direct Indonesian guidance for siswa/pengawas instead of security claims that cannot be guaranteed on student-owned devices.

## Visual Direction

- Prefer calm, grounded colors over flashy gradients.
- Avoid generic startup SaaS aesthetics.
- Avoid purple-heavy AI-style palettes.
- Favor readable typography and clear section hierarchy.
- Use visual emphasis to support tasks, not to decorate empty space.

## Component Direction

- Prefer `shadcn-svelte` for new UI primitives.
- Keep tables and forms highly scannable.
- Use cards, tabs, badges, dialogs, and alerts consistently.
- Prefer denser operational panels over oversized cards for CBT authoring and monitoring screens.
- Important actions should be visually obvious but not aggressive.

## UX Direction

- Optimize for operational clarity.
- Design for admin and guru users who need confidence and speed.
- Make statuses explicit and easy to compare.
- Keep destructive actions deliberate and confirmable.
- Preserve acceptable mobile behavior, especially for monitoring and approval screens.
- Keep autosave/draft indicators compact; status should be visible without stealing vertical space from authoring fields.

## Avoid

- generic KPI dashboards with no school context
- novelty-first motion
- cramped tables
- unclear status colors
- oversized decorative hero sections on internal admin screens
