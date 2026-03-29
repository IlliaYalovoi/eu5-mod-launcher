# Task 25 — Integration Smoke Test v2

## Goal

Validate the post-MVP feature set (launching, refactor stability, concurrency changes, Steam metadata, unsubscribe workflow) end-to-end.

## Context

Depends on: tasks 13-24.
Manual QA checklist task.

## Checklist

## A. Launch flow
- [ ] Effective game executable resolves (auto or custom override).
- [ ] "Launch Game" starts game.
- [ ] Launcher can close while game keeps running.

## B. Load order + constraints
- [ ] Mod constraints (mod↔mod) still work.
- [ ] Category constraints (category↔category) still work.
- [ ] Mixed constraints are rejected with clear error.
- [ ] Autosort updates both game order and launcher category order.

## C. Settings fallback/override
- [ ] Auto-detect works for mods dir and game executable.
- [ ] Custom overrides work and persist.
- [ ] Reset to auto works.
- [ ] Required-path modal appears only when needed.

## D. Performance/concurrency
- [ ] Large mod corpus scan remains stable.
- [ ] No race issues in targeted tests.
- [ ] UI remains responsive during scans.

## E. Steam metadata
- [ ] Workshop mod details load (title/description/thumbnail).
- [ ] Cached metadata/thumbnails are reused across restarts.

## F. Unsubscribe
- [ ] Unsubscribe available for workshop mods only.
- [ ] Confirmation shown.
- [ ] Success/error paths handled correctly.

## Deliverables

- Short report in `tasks/examples/smoke-v2-report.md` with:
  - date
  - OS/build used
  - pass/fail per section
  - known follow-ups

## Notes for agent

- Keep this task manual-first.
- Attach screenshots/gifs only if useful; not required.

