# Task 17 — Go Refactor: Repositories, Interfaces, and Boundary Tests

## Goal

Introduce repository interfaces for persistence and external integrations, then add boundary-focused tests to harden maintainability.

## Context

Depends on: task 16.

## Deliverables

### Repository interfaces

Define interfaces for:
- settings persistence
- launcher layout persistence
- constraints persistence
- playset persistence

Provide concrete file-backed implementations in `internal/repo/...`.

### Service wiring

Services from task 16 depend on interfaces, not direct file IO.

### Boundary tests

Add tests with mocks/fakes to verify:
- service behavior under repository errors
- transaction-like sequences (e.g. autosort + layout save)
- no partial state leaks on failures

## Acceptance criteria

- File IO paths are isolated from business flow.
- Services are mock-testable.
- Existing behavior preserved.

## Notes for agent

- Keep concrete implementations thin.
- Prefer table-driven tests for service+repo contracts.

