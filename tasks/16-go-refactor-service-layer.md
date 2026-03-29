# Task 16 — Go Refactor: Service Layer Extraction from `app.go`

## Goal

Split `app.go` orchestration into cohesive services so business logic is testable outside Wails glue code.

## Context

Depends on: task 15.
Current backend works but `app.go` has become a large multi-responsibility file.

## Deliverables

### New services (suggested)

- `internal/service/mods_service.go`
- `internal/service/loadorder_service.go`
- `internal/service/constraints_service.go`
- `internal/service/layout_service.go`
- `internal/service/settings_service.go`

Each service should encapsulate one bounded area and expose small methods.

### `app.go`

`App` methods become thin wrappers:
- validate readiness
- call service
- map/return results

### Tests

Service-level tests for key scenarios (happy path + conflicts + IO errors).

## Acceptance criteria

- `app.go` reduced substantially and reads as API adapter.
- Core behaviors covered by service tests.
- No regression in existing public methods.

## Notes for agent

- Do not change frontend API in this task.
- Keep migration incremental (extract one service at a time).

