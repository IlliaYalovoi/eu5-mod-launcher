# Task 25 — Go Refactor: Game Adapter Interfaces

## Goal

Hide game-specific mod list import/export logic behind explicit interfaces so backend can support multiple Paradox games.

## Context

Depends on: tasks 16-17.
Current implementation is EU5-centric in `app.go` + internal services.

## Deliverables

### Domain contracts

Create backend game adapter contracts in a new package (suggested: `internal/game`):

- `GameID` type (`eu5`, `vic3`, ...)
- `GameDescriptor` (id, display name, install/document paths, detected status)
- `ModListAdapter` interface for:
  - read current enabled/ordered mods
  - write compiled load order
  - discover game-specific roots/paths
- optional companion interfaces if needed:
  - `GameDetector`
  - `GamePathValidator`

### Service boundary

Introduce orchestrator service (suggested: `internal/service/game_service.go`) that:

- resolves adapter by `GameID`
- delegates import/export through interface only
- returns typed errors (`unsupported game`, `paths missing`, `invalid config`)

### App bridge

Refactor `app.go` to call orchestrator, not game-specific storage directly.

## Acceptance criteria

- No direct EU5 playset/mod-list format assumptions remain in generic flows.
- All mod list import/export calls pass through adapter interfaces.
- Existing EU5 behavior still works through adapter path.

## Notes for agent

- Keep API serialization-friendly for Wails.
- Prefer incremental migration: add interface + adapter registry, then switch call sites.

