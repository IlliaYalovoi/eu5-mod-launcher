# Redesign Specification: Multi-Game Paradox Mod Launcher

**Date:** 2026-04-13
**Status:** Draft / Pending Review
**Topic:** Multi-game support and backend refactor

## 1. Current State Analysis

The current codebase is a monolithic Europa Universalis V (EU5) launcher. While functional, it possesses structural characteristics that prevent easy extension to other Paradox Interactive titles.

### Identified Issues
- **Hardcoded Logic:** `internal/loadorder/paths.go` contains EU5-specific hardcoded paths and Steam AppIDs.
- **Domain Leakage:** `App.go` (the Wails interface) contains complex sorting logic, category management, and Steam metadata caching that should belong in dedicated services.
- **Incompatible Models:** Current playset handling assumes the EU5 JSON format. Other games (HOI4, CK3, Stellaris) use a SQLite database (`launcher-v2.db`) with a different schema.
- **Previous Failure:** Reverted commits indicate that adding multi-game support previously resulted in a "broken" state, likely due to a lack of clear abstraction between the UI and game-specific disk formats.

---

## 2. Domain Model Redesign

We will move from raw file manipulation to a **Normalized Domain Model**.

### Core Entities
- **GameDefinition:** Metadata (ID, Name, SteamID) and the adapter type.
- **GameInstance:** A specific installation on the user's system (InstallPath, UserConfigPath, LocalModsDir, WorkshopDirs).
- **Playset:** A normalized collection of mod entries.
- **ModEntry:** `ID`, `Path`, `Enabled`, `Position` (load order).

### GameAdapter Interface
Each game or group of games with shared formats will implement this interface:

```go
type GameAdapter interface {
    // Discovery
    DetectInstances() ([]GameInstance, error)
    
    // Data Loading
    LoadMods(inst GameInstance) ([]Mod, error)
    LoadPlaysets(inst GameInstance) ([]Playset, error)
    
    // Persistence
    SavePlayset(inst GameInstance, p Playset) error
}
```

---

## 3. Adaptation Layer Design

The system must present a consistent **EU5-style View Model** to the frontend, regardless of how the data is stored on disk.

### Transformation Logic
1. **Source (SQLite/JSON):** The Adapter loads the raw data.
2. **Domain Model:** The Adapter maps raw data to a `Playset` entity (which lists ALL known mods for that playset, marked as `enabled` or `disabled`).
3. **View Model (Frontend):** 
   - `EnabledList`: Mods where `Enabled == true`, ordered by `Position`.
   - `DisabledList`: Mods where `Enabled == false`, ordered alphabetically.
4. **Action (User moves/enables a mod):** The backend updates the Domain Model and triggers the Adapter to persist the change back to the source (e.g., updating a SQLite row or rewriting a JSON file).

---

## 4. Backend Architecture

### Layered Structure
1. **Transport (Wails Bindings):** `App.go` - Translates frontend calls to Service calls.
2. **Game Service:** Manages the "Active Game" context. Resolves which `GameAdapter` and `GameInstance` to use.
3. **Adapters:**
   - `JsonAdapter`: Handles EU5-style JSON playsets.
   - `SqliteAdapter`: Handles legacy Paradox Launcher SQLite databases.
4. **Domain:** Pure business logic (sorting, constraints, graph resolution).
5. **Infrastructure:** `repo/`, `steam/`, `xdg/`.

### Package Restructure
- `internal/game/` - Registry, Definitions, Interfaces.
- `internal/adapters/` - Concrete implementations (EU5, HOI4, etc.).
- `internal/service/` - High-level orchestration.

---

## 5. Frontend Simplification

The Vue/TS layer will be reduced to a **Thin UI**.

### Constraints
- **No JS Sorting:** Sorting and filtering occur in Go.
- **No Logic in Pinia:** Stores act as simple reactive wrappers for backend data.
- **Theming:** Based on the active game ID.
  - Sidebar: Shows all detected games.
  - Transition: Switching a game calls `App.SetActiveGame(id)`, clears local state, and re-fetches all data.

---

## 6. Migration Plan

1. **Phase 1: Interfaces & EU5 Adapter:** Refactor existing EU5 logic into a `JsonAdapter` without breaking current functionality.
2. **Phase 2: Context Management:** Implement `GameService` to support multiple instances/games.
3. **Phase 3: SQLite Support:** Implement the `SqliteAdapter` using `modernc.org/sqlite`. Test with Victoria 3 or HOI4.
4. **Phase 4: UI Update:** Implement the side-rail game selector and the dynamic theming system.

---

## 7. Risk Analysis

- **Data Corruption (SQLite):** Directly editing the launcher database can be destructive.
  - **Mitigation:** Always backup `launcher-v2.db` before write; use transactions.
- **Path Resolution:** Paradox games use various paths across Linux/Windows.
  - **Mitigation:** Use `github.com/adrg/xdg` for standard paths; implement robust Steam library discovery.
- **Performance:** SQL queries vs. JSON reads.
  - **Mitigation:** Ensure indices are used; cache metadata.

---

## 8. Requirements Summary
- Use `adrg/xdg` for config.
- Use `go-resty` for Steam API.
- Use `modernc.org/sqlite` (CGO-free).
- Keep Frontend logic-free.
