# Task 14 — Frontend: Launch Controls

## Goal

Add a dedicated "Launch Game" action in UI with clear status feedback and robust error handling.

## Context

Depends on: task 13 (detached backend launch API), task 06 (design system).

## Deliverables

### `frontend/src/stores/settings.ts`

Add store action:

```ts
async function launchGame(): Promise<void>
```

Behavior:
- Calls `LaunchGame()` binding.
- Exposes refs:
  - `isLaunching: boolean`
  - `launchError: string | null`
  - `lastLaunchAt: number | null`

### `frontend/src/components/LaunchButton.vue`

New component that reads launch state from store and renders:
- idle: `Launch Game`
- launching: spinner + `Launching...`
- success flash (short)
- error state (danger styling)

### Integration

Render launch action in top bar or load-order header (project-consistent placement).

## Acceptance criteria

- Button triggers backend launch.
- UI does not freeze while launching.
- Launch errors are visible and dismissible.
- Launch success gives short visual feedback.

## Notes for agent

- Keep component store-driven (no direct Wails call in component).
- Reuse existing button and state color tokens.

