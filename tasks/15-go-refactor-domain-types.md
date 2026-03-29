# Task 15 — Go Refactor: Domain Types & Strict API Contracts

## Goal

Introduce explicit domain types and stricter method contracts to reduce ad-hoc string handling and improve maintainability.

## Context

Depends on: tasks 01-04.
Can be implemented incrementally without changing user-visible behavior.

## Deliverables

### New package `internal/domain` (or equivalent)

Define typed identifiers and DTOs:

```go
type ModID string
type CategoryID string
type PlaysetIndex int
```

Add validation helpers:
- `ParseModID(string) (ModID, error)`
- `IsCategoryID(string) bool`
- etc.

### Backend signatures

Refactor internal-facing methods to use domain types where possible.
Keep Wails boundary methods JSON-friendly, with conversion at boundary.

### Error taxonomy

Introduce typed sentinel/domain errors (e.g. invalid target type, not found, conflict), then wrap with context.

### Tests

Add tests for type parsing/validation and conversion boundaries.

## Acceptance criteria

- App behavior unchanged for valid inputs.
- Invalid input handling is more explicit and consistently typed.
- New domain package has unit tests.

## Notes for agent

- Avoid large file moves in one step; keep refactor reviewable.
- Preserve Wails-exposed method compatibility in this task.

