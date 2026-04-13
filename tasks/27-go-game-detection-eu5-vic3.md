# Task 27 — Go: Multi-Game Detection (EU5 + Vic3)

## Goal

Add backend game detection for supported titles, starting with Europa Universalis V and Victoria 3.

## Context

Depends on: task 25.
Task 26 should be available for EU5 adapter; Vic3 adapter can be detection-only in this phase.

## Deliverables

### Detection service

Create detection service (suggested: `internal/service/game_detection_service.go`) that returns sorted game inventory:

```go
type DetectedGame struct {
    ID               string
    Name             string
    IconKey          string
    Detected         bool
    InstallDir       string
    DocumentsDir     string
    NeedsManualSetup bool
}
```

Rules:

- support IDs: `eu5`, `vic3`
- detected games first
- stable ordering among detected/undetected buckets

### App API

Expose Wails methods:

- `ListSupportedGames() ([]DetectedGame, error)`
- `SetGamePaths(gameID, installDir, documentsDir string) error`

### Persistence

Persist manual overrides in settings repository keyed by `GameID`.

### Tests

- detector tests for both games
- ordering tests (detected first)
- override merge tests (auto-detect + manual)

## Acceptance criteria

- Backend reports both EU5 and Vic3 with detection state.
- Manual paths can be set and are reflected in subsequent list calls.
- Detection results deterministic and UI-ready.

## Notes for agent

- Keep detection implementation platform-aware but isolated.
- Detection should not fail entire call when one game probe errors.

