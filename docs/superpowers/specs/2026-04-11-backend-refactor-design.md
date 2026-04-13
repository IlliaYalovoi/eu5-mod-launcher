# Backend Refactor Design — EU5 Mod Launcher

## Status
Approved by user 2026-04-11.

---

## 1. Goals

- Reduce total Go code by >30%
- Eliminate boilerplate, empty files, duplicate types, dead code
- Replace logrus with slog (stdlib)
- Add mapstructure to eliminate manual struct mapping
- Add pkgerrors for error stack traces
- Collapse repo layer to essential interfaces only
- Reorganize into 3 bounded contexts
- Preserve all existing functionality
- Frontend compatibility not required during refactor

---

## 2. Package Structure

```
internal/
  domain/          # Shared types only (no empty files)
    constraint.go
    errors.go
    game.go
    loadorder.go
    mod.go

  launcher/        # Wails-exposed App, mod scanning, loadorder, constraints, graph
    app_mods.go
    app_game.go
    app_constraints.go
    app_layout.go
    app_workshop.go
    app_conversion.go
    app_structs.go
    settings.go
    wire.go        # NewApp factory, dependency construction
    loadorder.go
    playsets.go
    graph.go

  game/            # Game detection, adapters, launch process
    detection.go
    adapter.go
    eu5.go
    launch.go

  steam/           # Workshop, metadata, image cache
    workshop.go
    client.go
    cache.go
    metadata.go
    images.go
    descriptions.go
    steam.go       # steamAppID const + helpers

  repo/            # Interfaces + file-backed implementations
    constraints_repo.go
    layout_repo.go
    settings_repo.go
    playset_repo.go
    loadorder_repo.go

  service/         # Thin orchestration layer (kept as-is structurally)
    constraints_service.go
    loadorder_service.go
    mods_service.go
    game_service.go
    layout_service.go
    playset_service.go
    settings_service.go
    launch_service.go

main (root)/
  main.go          # Entry point, wires launcher via factory
```

**Removed from root:** `app.go`, `app_aux.go`, `app_wiring.go`, `launch_process_*.go` (duplicate), `feature_unsubscribe_*.go`, `launcher_layout.go`

**Removed from domain:** `parse.go`, `types.go` (empty), `game/contracts.go` (naked aliases collapsed)

---

## 3. Dependency Changes

| Dependency | Action |
|---|---|
| `github.com/sirupsen/logrus` | **Remove** — replace with `log/slog` |
| `github.com/pkgerrors` | **Add** — error stack traces |
| `github.com/mitchellh/mapstructure` | **Add** — struct mapping |
| `gopkg.in/yaml.v4` | Keep (for future config) |
| `resty`, `gjson`, `xdg` | Keep |

---

## 4. Key Design Decisions

### 4.1 DI: Manual Factory

```go
// internal/launcher/wire.go
type Dependencies struct {
  SettingsRepo   repo.SettingsRepo
  ConstraintsRepo repo.ConstraintsRepo
  LayoutRepo     repo.LayoutRepo
  PlaysetRepo    repo.PlaysetRepo
  LoadOrderRepo  repo.LoadOrderRepo
  SteamClient    *steam.Client
  MetadataCache  *steam.MetadataCache
  ImageCache     *steam.ImageCache
  GameDetector   *game.Detector
}

func NewLauncher(deps Dependencies) *App
```

No DI framework. Construct `Dependencies` in `main.go`, pass to `NewLauncher`.

### 4.2 Repo Layer

Keep interfaces for extension points:
- `ConstraintsRepo` — multi-game constraint storage
- `PlaysetRepo` — multi-game playset storage
- `LayoutRepo` — multi-game layout storage
- `SettingsRepo` — multi-game settings

Single-impl repos (`LoadOrderRepo`, `FileLayoutRepo`) become concrete types.

### 4.3 Error Handling

- Sentinel errors: `var ErrSomething = errors.New("something")` in each package
- Wrapping: `fmt.Errorf("context: %w", err)` at call sites
- Stacks: `pkgerrors.Wrap` at top-level boundary (Wails method entry)
- Logging: `slog.Error` at Wails method boundary only — never log + return
- Naming: `Err` prefix for all sentinel errors

### 4.4 `mustBeReady()` Guard

Replace `ensureReady()` + inconsistent `logging.Errorf` with:
```go
func (a *App) mustBeReady() error {
  if !a.initialized {
    return ErrNotInitialized
  }
  return nil
}
```
All Wails methods: `if err := a.mustBeReady(); err != nil { return nil, err }`

### 4.5 Steam App ID

Single const in `internal/steam/steam.go`:
```go
const steamAppID = "3450310"
```
Removed from: `app_wiring.go`, `loadorder/paths.go`, `game_detection_service.go`

### 4.6 Deduplication

`launch_process_unix.go` and `launch_process_windows.go` exist at root AND in `internal/service/`. Consolidate to `internal/game/launch.go` only.

### 4.7 Struct Mapping

Replace manual `toRepoSettings`, `fromRepoSettings`, `toRepoLayout`, `fromRepoLayout` with `mapstructure.Decode`.

### 4.8 `ensureSteamCaches` on-demand creation

Current: caches re-created on first access if nil.
Change: caches created in `wire.go` factory, injected as dependencies. No lazy creation.

### 4.9 Dead Code Removal

- `app_aux.go` stubs (`ensureSteamCaches`, `resolveImageSource` — move前者 to `wire.go`, 后者 remains stub)
- `feature_unsubscribe_*.go` build-tag files — remove feature flag, just keep the code
- `domain/parse.go`, `domain/types.go` — delete (empty)
- `game/contracts.go` — remove naked type aliases, import `domain` directly
- `steamItemIDForMod` — keep stub

---

## 5. Data Flow (unchanged)

```
main.go
  main() → NewLauncher(deps) → App struct
Wails runtime → App.WailsMethod()
  → mustBeReady() check
  → service call (via appServices fields)
  → repo call
  → file/steam/registry
```

---

## 6. Implementation Order

1. Prune dead code and empty files (lowest risk)
2. Add new deps (`slog`, `pkgerrors`, `mapstructure`) — update all error + logging calls
3. Create new package structure (`launcher/`, `game/`, `steam/`) — move files
4. Implement `wire.go` factory with `Dependencies` struct
5. Collapse repo interfaces to essential set
6. Add `mustBeReady()` guard to all Wails methods
7. Deduplicate `launch_process` files → `internal/game/launch.go`
8. Single `steamAppID` const
9. Replace manual struct mapping with `mapstructure`
10. Remove old root files, verify build passes

---

## 7. Out of Scope

- Frontend changes
- `steamItemIDForMod` implementation
- New config format (YAML/TOML)
- Any new features
- Test changes (existing tests should pass; do not add new tests)
