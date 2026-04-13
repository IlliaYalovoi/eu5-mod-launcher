# Task 23 — Go: Steam Workshop Unsubscribe Action

## Goal

Allow backend to unsubscribe a workshop mod directly from launcher.

## Context

Depends on: task 20.
May require Steam protocol/deeplink or Steamworks-compatible command path.

## Deliverables

### Backend action

Expose method:

```go
func (a *App) UnsubscribeWorkshopMod(itemID string) error
```

Behavior:
- validates workshop ID
- triggers unsubscribe flow via supported mechanism
- returns actionable errors

### Safety

- no-op guard for non-workshop mods
- confirmation signal support for frontend (if needed)

### Tests

- ID validation
- command/protocol invocation formatting
- error mapping

## Acceptance criteria

- Backend method can be called for a workshop item ID and returns success/error correctly.
- Invalid IDs are rejected.

## Notes for agent

- Keep mechanism platform-aware.
- Do not block UI waiting for Steam sync completion.

