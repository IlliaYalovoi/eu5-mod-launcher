# Task 13 — Go: Detached Game Launcher

## Goal

Add a backend launcher API that starts the game executable as a fully detached process so the game keeps running even if the launcher exits.

## Context

Depends on: task 04 (app bridge), task 12 (settings with game executable).

This task is backend-only. No frontend button in this task.

## Deliverables

### `app.go`

Add:

```go
// LaunchGame starts the configured game executable in a detached process.
// Returns nil once the child process has been spawned.
func (a *App) LaunchGame() error
```

Behavior:
- Resolves effective game executable path (custom override or autodetected fallback).
- Validates that executable exists and is executable.
- Starts process detached from launcher lifecycle:
  - Windows: independent process group / no inherited console lock.
  - Linux/macOS: detached session equivalent.
- Optional args support via settings (if present); otherwise launch with no args.
- Logs launch attempt and PID when available.

### `settings.go`

If needed, extend settings schema with optional launch arguments:

```json
{ "game_args": ["..."] }
```

(keep backwards compatibility with existing settings file)

### Tests

Add focused tests for:
- missing executable path -> error
- invalid path -> error
- launch command construction for each OS branch (unit level)

## Acceptance criteria

- Calling `LaunchGame()` returns quickly after spawning process.
- Closing launcher does not terminate game process.
- Errors are returned (not panic) for missing/invalid executable path.

## Notes for agent

- Keep implementation in backend only; frontend wiring is next task.
- Avoid shell invocation when possible; prefer direct process exec.
- Wrap errors with context (`fmt.Errorf(... %w ...)`).

