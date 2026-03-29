# Task 24 — Frontend: Unsubscribe Workflow (Context Menu + Details)

## Goal

Expose workshop unsubscribe in UI from both context menu and mod details panel.

## Context

Depends on: tasks 22-23.

## Deliverables

### Context menu integration

For workshop mods add action:
- `Unsubscribe from Workshop...`

With confirmation prompt before execution.

### Mod details integration

In details panel add action button:
- `Unsubscribe`

### State behavior

- pending state while request is in-flight
- success toast/info + refresh mod list
- error surfaced inline/toast

## Acceptance criteria

- User can unsubscribe from both entry points.
- After unsubscribe success, mod list refreshes and reflects new state after next scan.
- Errors are clear and do not break UI.

## Notes for agent

- Keep confirmation explicit (avoid accidental unsubscribe).
- Deduplicate unsubscribe requests per mod while pending.

