# MOD LAUNCHER — SPEC

## PURPOSE

Desktop app for managing Paradox-style mods across multiple game titles.

- Platform: Wails v2 (Go backend + Vue 3 + TypeScript)
- Output: single native binary
- Scope:
    - detect installed games (EU5, Vic3)
    - read mod metadata
    - manage load order per game
    - enforce ordering constraints
    - write final config
    - optionally launch game

---

## CORE FLOWS

1. GAME DETECTION
    - probe known registry locations
    - support manual path override
    - report detected vs undetected games

2. DISCOVERY
    - scan game-specific mods directories
    - extract metadata (name, version, tags, thumbnail)
    - Workshop support via Steam

3. LOAD ORDER
    - per-game mod enable/disable
    - reorder via drag-and-drop

4. CONSTRAINTS
    - define relations: AFTER(X), BEFORE(Y)
    - stored as directed graph per game

5. AUTOSORT
    - topological sort
    - detect + report cycles

6. LAUNCH
    - write resolved order to config
    - start game process

---

## ARCHITECTURE

```
/ (root)
  main.go           → entry point, wires launcher via factory

internal/
  domain/            → shared types (no business logic)
    constraint.go
    errors.go
    game.go
    loadorder.go
    mod.go

  launcher/          → Wails-exposed App, mod scanning, loadorder, constraints
    app_mods.go
    app_game.go
    app_constraints.go
    app_layout.go
    app_workshop.go
    app_conversion.go
    app_structs.go
    settings.go
    wire.go          → NewApp factory, dependency construction
    loadorder.go
    playsets.go
    graph.go

  game/              → game detection, adapters, launch process
    detection.go
    adapter.go
    eu5.go
    launch.go

  steam/             → Workshop, metadata, image cache
    workshop.go
    client.go
    cache.go
    metadata.go
    images.go
    descriptions.go
    steam.go         → steamAppID const + helpers

  repo/              → interfaces + file-backed implementations
    constraints_repo.go
    layout_repo.go
    settings_repo.go
    playset_repo.go
    loadorder_repo.go

  service/           → thin orchestration layer
    constraints_service.go
    loadorder_service.go
    mods_service.go
    game_service.go
    layout_service.go
    playset_service.go
    settings_service.go
    launch_service.go

frontend/
  src/
    components/
    stores/          → Pinia (SOURCE OF TRUTH)
    views/
    wailsjs/         → GENERATED (DO NOT EDIT)
```

---

## TECH (FIXED)

- GUI: Wails v2
- FE: Vue 3 + TypeScript + Vite
- State: Pinia
- DnD: vuedraggable (SortableJS)
- Styling: scoped CSS + variables
- Persistence: JSON (user config dir)
- Graph: pure Go (NO frontend logic)
- Logging: slog (stdlib)
- Error stacks: pkgerrors
- Struct mapping: mapstructure

---

## BACKEND CONTRACTS (STRICT)

- Wails exposure = methods on `App`
- Return ONLY plain structs (NO interfaces)
- Function signature: `(Result, error)`
- ALL business logic → backend
- frontend = view ONLY
- File paths: use `filepath`, NEVER string concat

---

## FRONTEND RULES (STRICT)

- Pinia = SINGLE SOURCE OF TRUTH
- Components:
    - read from stores ONLY
    - NO persistent local state
- State updates:
    - call backend
    - replace store state with response
- Errors: backend → rejected promise, MUST be handled

---

## DATA FLOW

```
Frontend → Go method → Result → Store update → UI
```

NO DIRECT STATE MUTATION

---

## GRAPH RULES

- constraints = directed edges
- sorting = topological sort
- cycles: must be detected, reported, NOT crash

---

## NON-GOALS

- NO business logic in Vue
- NO frontend-side sorting logic
- NO manual state syncing
- NO editing `wailsjs/`