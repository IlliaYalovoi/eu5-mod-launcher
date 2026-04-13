# EU5 Mod Launcher — Project Overview

## What this is

A standalone desktop application for managing mods for Europa Universalis 5 (and similar Paradox-style games). Built with **Wails v2** (Go backend + Vue 3 + TypeScript frontend), compiled to a single native binary.

The launcher sits outside the game itself. It reads mod metadata from a well-known directory, lets the user curate a load order, define ordering constraints between mods, and writes the final load order back to a format the game understands.

---

## Core user flows

1. **Discovery** — App scans the mods directory and shows all installed mods with metadata (name, version, tags, thumbnail if present).
2. **Load order curation** — User enables/disables mods and arranges them via drag-and-drop in an ordered list.
3. **Constraint authoring** — User right-clicks a mod to define relations: "this mod always loads after X" or "always before Y". These are stored as a directed graph.
4. **Autosort** — User triggers a sort that resolves the constraint graph (topological sort) and reorders the active list accordingly, warning on cycles.
5. **Launch** — App writes the resolved load order to the game's config and optionally launches the game executable.

---

## Architecture

```
/                        ← Wails project root
├── main.go              ← Wails bootstrap
├── app.go               ← Go backend: all business logic exposed to frontend
├── internal/
│   ├── mods/            ← Mod scanning, metadata parsing
│   ├── loadorder/       ← Load order state, persistence
│   └── graph/           ← Constraint graph, topological sort
├── frontend/
│   ├── src/
│   │   ├── components/  ← Vue components
│   │   ├── stores/      ← Pinia stores (mod list, load order, settings)
│   │   ├── views/       ← Top-level page views
│   │   └── wailsjs/     ← Auto-generated Go bindings (DO NOT edit)
│   └── ...
└── tasks/               ← You are here
```

---

## Tech choices

| Concern            | Choice                                          |
|--------------------|-------------------------------------------------|
| GUI framework      | Wails v2                                        |
| Frontend           | Vue 3 + TypeScript + Vite                       |
| State management   | Pinia                                           |
| Drag and drop      | vuedraggable (wraps SortableJS)                 |
| Styling            | CSS custom properties + scoped component styles |
| Config persistence | JSON file in OS user config dir                 |
| Graph / sort       | Pure Go, no exposure to frontend                |

---

## Task index

Tasks are designed to be **maximally independent**. Each one has a clear input, clear output, and minimal assumptions about other tasks being done first. Do them roughly in order, but most can be handed to an AI agent as a standalone context.

| # | File | Scope |
|---|---|---|
| 01 | `01-go-mod-scanner.md` | Go: scan mods directory, parse metadata |
| 02 | `02-go-loadorder-store.md` | Go: persist & load the load order JSON |
| 03 | `03-go-constraint-graph.md` | Go: constraint graph data structure + topological sort |
| 04 | `04-go-app-bridge.md` | Go: wire internal packages into Wails `app.go` methods |
| 05 | `05-fe-project-setup.md` | Frontend: Pinia stores skeleton, Wails bindings integration |
| 06 | `06-fe-design-system.md` | Frontend: global CSS design tokens, typography, base component stubs |
| 07 | `07-fe-mod-list-panel.md` | Frontend: "All mods" panel, search/filter, enable toggle |
| 08 | `08-fe-load-order-panel.md` | Frontend: ordered list of active mods, drag-and-drop reorder |
| 09 | `09-fe-context-menu.md` | Frontend: right-click context menu component (reusable) |
| 10 | `10-fe-constraint-modal.md` | Frontend: modal for adding/viewing constraints on a mod |
| 11 | `11-fe-autosort.md` | Frontend: autosort button, cycle error display |
| 12 | `12-fe-settings.md` | Frontend: settings panel (mods path, game executable path) |
| 13 | `13-go-detached-game-launcher.md` | Go: launch game as detached process |
| 14 | `14-fe-launch-controls.md` | Frontend: launch button and launch-state UX |
| 15 | `15-go-refactor-domain-types.md` | Go refactor: domain types and strict contracts |
| 16 | `16-go-refactor-service-layer.md` | Go refactor: extract service layer from app glue |
| 17 | `17-go-refactor-repositories-and-boundaries.md` | Go refactor: repository interfaces and boundary tests |
| 18 | `18-go-concurrent-mod-scan.md` | Go performance: concurrent scanner pipeline |
| 19 | `19-go-concurrency-audit.md` | Go performance: profile-driven concurrency improvements |
| 20 | `20-go-steam-workshop-metadata.md` | Go: Steam workshop metadata client |
| 21 | `21-go-steam-metadata-cache.md` | Go: metadata + thumbnail cache layer |
| 22 | `22-fe-steam-mod-details.md` | Frontend: Steam-enriched mod details panel |
| 22.5 | `22.5-fe-steam-description-rendering-and-open-priority.md` | Frontend+Go: Steam BBCode rendering, description image cache, workshop open fallback priority |
| 23 | `23-go-steam-unsubscribe.md` | Go: unsubscribe workshop item action |
| 24 | `24-fe-unsubscribe-workflow.md` | Frontend: unsubscribe UX from context/details |
| 25 | `25-go-game-adapter-interfaces.md` | Go refactor: game adapter interfaces for mod list import/export |
| 26 | `26-go-eu5-game-adapter.md` | Go: EU5 concrete adapter over new game interface |
| 27 | `27-go-game-detection-eu5-vic3.md` | Go: detect supported games (EU5, Vic3) + manual path overrides |
| 28 | `28-fe-multi-game-sidebar.md` | Frontend: left game sidebar, detected-state ordering and switching |
| 28.5 | `28.5-fe-ui-reform.md` | Frontend: bold UI reform — 2-row layout, slide-over panels, keyboard shortcuts |
| 28.6 | `28.6-fe-ui-reform-pt2.md` | Frontend: 7 UI bug fixes from 28.5 review |
| 29 | `29-fe-manual-game-path-setup.md` | Frontend: popup workflow for manual install/documents paths |
| 30 | `30-fe-game-themes.md` | Frontend: per-game theme tokens and runtime theme switch |
| 31 | `31-go-fe-last-selected-game-persistence.md` | Go+Frontend: persist/restore last selected game |
| 32 | `32-go-vic3-playsets.md` | Go: Vic3 SQLite playset repository |
| 33 | `33-go-mod-game-version-check.md` | Go+Frontend: mod/game version compatibility check |
---

## Conventions to keep consistent across tasks

- All Go public methods on `App` struct are what Wails exposes — keep them flat and serialization-friendly (return plain structs, not interfaces).
- Frontend never mutates backend state directly — always calls a Go method, then refreshes from the returned value.
- Pinia stores are the single source of truth on the frontend. Components read from stores, never from local component state for anything that needs to persist.
- Error handling: Go methods return `(Result, error)`. Wails surfaces errors as rejected JS promises. Frontend must handle them.
- File paths use `filepath` package on Go side — never string concatenation.
- Everything that can be done on backend should be done on backend — frontend is just a view layer. No business logic in Vue components.