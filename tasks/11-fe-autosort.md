# Task 11 — Frontend: Autosort UX & Cycle Error Display

## Goal

Polish the autosort flow: loading state during sort, success feedback, and a clear, actionable error UI when a cycle is detected.

## Context

Depends on: task 08 (autosort button already calls `loadOrderStore.autosort()`), task 05 (store has `autosortError` ref). This task upgrades the error display and success feedback from "it works" to "it feels good."

## Deliverables

### `frontend/src/components/AutosortButton.vue`

Extract the autosort button from `LoadOrderPanel` into its own component.

States:
- **Idle**: "Auto-sort" label, normal style
- **Loading**: spinner + "Sorting..." label, disabled
- **Success**: brief green checkmark flash (1.5s), then back to idle
- **Error**: button turns red outline, stays that way until user acknowledges

Emits nothing — interacts with store directly.

### `frontend/src/components/CycleErrorPanel.vue`

Shown below the load order list when `loadOrderStore.autosortError` is set.

Displays:
- Error heading: "Constraint Cycle Detected"
- The error message from the backend (names the mods involved)
- A diagram hint: plain text representation of the cycle, e.g.:
  ```
  Mod A  →  Mod B  →  Mod C  →  Mod A
  ```
  (parse the error string from Go and format it — the Go `ErrCycle` message should contain mod IDs in order)
- Two actions:
  - "Dismiss" — clears `autosortError`, does not change order
  - "Open constraints for [first mod in cycle]" — opens the constraint modal for that mod

Animated: slides down when it appears, slides back up on dismiss.

### `frontend/src/stores/loadorder.ts` update

Ensure `autosort()` action:
1. Sets a `isSorting` boolean ref to true before the call.
2. On success: updates `orderedIDs`, clears `autosortError`, sets `lastSortedAt` timestamp.
3. On error: sets `autosortError` string, does NOT update `orderedIDs`.
4. Always sets `isSorting` to false in a finally block.

## Acceptance criteria

- Clicking autosort shows spinner during the async call.
- Successful sort shows a brief green flash before returning to normal.
- Cycle error shows the `CycleErrorPanel` with the cycle chain formatted correctly.
- "Open constraints" button in the error panel opens the constraint modal pre-targeted to the first mod in the cycle.
- Dismissing the error clears it without side effects.

## Notes for agent

- The cycle error message format from Go (task 03) should be something like `"cycle detected: mod_a -> mod_b -> mod_a"` — the frontend can split on ` -> ` to build the visual chain.
- If the Go error format is different, adjust the parser here accordingly and note it.
- Keep animations CSS-only (no animation library needed for this).
