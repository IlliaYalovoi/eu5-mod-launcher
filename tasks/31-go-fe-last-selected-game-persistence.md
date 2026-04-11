# Task 31 — Go + Frontend: Persist Last Selected Game

## Goal

Remember and restore last selected game across launcher restarts.

## Context

Depends on: tasks 27-28.

## Deliverables

### Backend settings

Add `LastSelectedGameID` to settings contract/repository.

Rules:

- persist on active game change
- on startup, try restore previous game
- if missing/unsupported, fallback to first detected game
- if no detected games, fallback to default supported entry

### Frontend integration

- initialize active game from backend startup payload or dedicated API
- keep store/backend in sync when user switches games

### Tests

- settings repo round-trip for last selected game
- startup selection fallback matrix
- unsupported previously selected game handling

## Acceptance criteria

- App reopens with previously selected valid game.
- Safe fallback behavior when previous game no longer available.
- No startup crash on empty or stale persisted value.

## Notes for agent

- Keep restore logic deterministic and side-effect free.
- Ensure persisted game switch triggers same refresh pipeline as manual switch.

