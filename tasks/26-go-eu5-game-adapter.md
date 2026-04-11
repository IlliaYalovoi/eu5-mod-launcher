# Task 26 — Go: EU5 Adapter Implementation

## Goal

Implement EU5 adapter as first concrete realization of new game adapter interfaces.

## Context

Depends on: task 25.
Existing EU5 playset/load-order behavior must be preserved.

## Deliverables

### Package

Create EU5 adapter package (suggested: `internal/game/eu5`).

### Adapter implementation

Implement required interface methods:

- detect/install/document roots for EU5
- import current mod list/order from EU5 format
- export compiled list back to EU5 format
- normalize/validate EU5-specific paths

### Integration

Register EU5 adapter in adapter registry/factory.

### Tests

- fixture-based import/export compatibility tests
- path normalization tests
- error mapping tests for malformed/absent files

## Acceptance criteria

- EU5 runs fully via adapter layer.
- Import/export output remains compatible with current game behavior.
- No regression in load-order persistence flow for EU5.

## Notes for agent

- Reuse existing `internal/loadorder` + playset logic where possible.
- Keep adapter API strict; avoid leaking EU5-specific types into generic package.

