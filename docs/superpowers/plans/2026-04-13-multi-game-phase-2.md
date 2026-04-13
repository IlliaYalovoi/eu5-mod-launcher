# Multi-Game Redesign Phase 2: SQLite & Legacy Adapter

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the SQLite adapter to support legacy Paradox games (HOI4, CK3, Stellaris, V3).

**Architecture:** A generic `SqliteAdapter` that maps `launcher-v2.db` tables (mods, playsets, playsets_modules) to our normalized Domain Model.

**Tech Stack:** Go, `modernc.org/sqlite v1.48.2`.

---

### Task 1: Basic SQLite Connectivity

**Files:**
- Create: `internal/adapters/legacy/sqlite.go`

- [ ] **Step 1: Setup SQLite adapter boilerplate**

```go
package legacy

import (
    "database/sql"
    _ "modernc.org/sqlite"
)

type SqliteAdapter struct {
    GameID     string
    SteamAppID string
}

func (a *SqliteAdapter) ID() string { return a.GameID }
```

- [ ] **Step 2: Implement playset loading**

```go
func (a *SqliteAdapter) LoadPlaysets(inst game.Instance) ([]game.Playset, error) {
    db, err := sql.Open("sqlite", inst.UserConfigPath + "/launcher-v2.db")
    if err != nil { return nil, err }
    defer db.Close()
    // Query playsets and playsets_modules
    return nil, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/adapters/legacy/sqlite.go
git commit -m "feat: implement generic SQLite adapter for legacy games"
```

---

### Task 2: Register Legacy Games

**Files:**
- Modify: `main.go` or `app.go`

- [ ] **Step 1: Register HOI4, CK3, Stellaris in GameService**

```go
gameSvc.Register(&legacy.SqliteAdapter{GameID: "hoi4", SteamAppID: "394360"})
gameSvc.Register(&legacy.SqliteAdapter{GameID: "ck3", SteamAppID: "1158310"})
```

- [ ] **Step 2: Commit**

```bash
git commit -m "feat: register legacy Paradox games"
```
