# Task 04 — Go: Wails App Bridge (`app.go`)

## Goal

Wire the three internal packages (`mods`, `loadorder`, `graph`) into the Wails `App` struct. All public methods on `App` become callable from the Vue frontend as TypeScript functions.

## Context

This task assumes tasks 01, 02, and 03 are complete. Do not reimplement logic here — only delegate to the internal packages. Keep method signatures flat and JSON-serializable.

## Deliverables

### `app.go` — full replacement of the Wails template stub

```go
type App struct {
    ctx        context.Context
    modsDir    string
    loStore    *loadorder.Store
    loState    loadorder.State
    conGraph   *graph.Graph
    // paths for constraint persistence
    constraintsPath string
}

func NewApp() *App

// startup is called by Wails when the app window is ready.
func (a *App) startup(ctx context.Context)
```

#### Methods to expose (these become TS bindings):

```go
// -- Mod discovery --
func (a *App) GetAllMods() ([]mods.Mod, error)
// Returns all scanned mods, with Enabled field set from current load order state.

// -- Load order --
func (a *App) GetLoadOrder() []string          // ordered list of enabled mod IDs
func (a *App) SetLoadOrder(ids []string) error  // persist new order
func (a *App) SetModEnabled(id string, enabled bool) error

// -- Constraints --
func (a *App) GetConstraints() []graph.Constraint
func (a *App) AddConstraint(from, to string) error
func (a *App) RemoveConstraint(from, to string) error

// -- Autosort --
func (a *App) Autosort() ([]string, error)
// Runs graph.Sort on the current enabled list, saves result, returns new order.
// Returns error string (not panic) on cycle.

// -- Settings --
func (a *App) GetModsDir() string
func (a *App) SetModsDir(path string) error
// Persists modsDir to settings file and triggers a rescan.
```

### `settings.go` (new file)

Simple JSON settings file (separate from load order) storing:
```json
{ "mods_dir": "..." }
```

Path: same config dir as load order, filename `settings.json`.

## Acceptance criteria

- `wails dev` starts without compile errors after this task.
- Calling `GetAllMods()` from browser devtools console returns a JSON array.
- `SetLoadOrder` followed by app restart preserves order.
- An invalid `modsDir` (nonexistent path) returns an error from `GetAllMods`, not a crash.

## Notes for agent

- Use `wails.LogInfo` / `wails.LogError` for structured logging inside app methods.
- Method receivers must be on `*App`, not `App`, or Wails won't bind them.
- Return types must be serializable — no channels, no unexported types.
- Wails will generate `frontend/src/wailsjs/go/` bindings automatically on `wails dev` — do not create those files manually.
