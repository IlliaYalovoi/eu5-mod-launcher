# Multi-Game Redesign Phase 1: Core Domain & EU5 Adapter

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the `GameAdapter` interface and migrate existing EU5 logic into a specialized `JsonAdapter`.

**Architecture:** Introduction of a registry-based adapter system. `App` will begin delegating to the `GameService` instead of hardcoded paths.

**Tech Stack:** Go, `github.com/adrg/xdg v0.5.3`.

---

### Task 1: Define Core Domain Entities

**Files:**
- Create: `internal/game/domain.go`
- Create: `internal/game/adapter.go`

- [ ] **Step 1: Create domain entities**

```go
package game

type Definition struct {
    ID          string
    DisplayName string
    SteamAppID  string
}

type Instance struct {
    GameID          string
    InstallPath     string
    UserConfigPath  string
    LocalModsDir    string
    WorkshopModDirs []string
    GameExePath     string
}

type ModEntry struct {
    ID       string
    Path     string
    Enabled  bool
    Position int
}

type Playset struct {
    ID      string
    Name    string
    Entries []ModEntry
}
```

- [ ] **Step 2: Define GameAdapter interface**

```go
package game

type Adapter interface {
    ID() string
    DetectInstances() ([]Instance, error)
    LoadMods(inst Instance) ([]ModEntry, error)
    LoadPlaysets(inst Instance) ([]Playset, error)
    SavePlayset(inst Instance, p Playset) error
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/game/*.go
git commit -m "feat: define core game domain and adapter interface"
```

---

### Task 2: Implement EU5 JSON Adapter

**Files:**
- Create: `internal/adapters/eu5/adapter.go`
- Modify: `internal/loadorder/paths.go` (extract logic to adapter)

- [ ] **Step 1: Create EU5 Adapter implementation**

```go
package eu5

import "eu5-mod-launcher/internal/game"

type Adapter struct {}

func (a *Adapter) ID() string { return "eu5" }

// Move logic from internal/loadorder/paths.go into these methods
func (a *Adapter) DetectInstances() ([]game.Instance, error) { 
    // ... implementation using xdg and steam discovery
    return nil, nil 
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/adapters/eu5/adapter.go
git commit -m "feat: implement EU5 JsonAdapter"
```

---

### Task 3: Initialize Game Service

**Files:**
- Create: `internal/service/game_service.go`

- [ ] **Step 1: Create GameService to manage active context**

```go
package service

import "eu5-mod-launcher/internal/game"

type GameService struct {
    adapters       map[string]game.Adapter
    activeGame     string
    activeInstance *game.Instance
}

func NewGameService() *GameService {
    return &GameService{adapters: make(map[string]game.Adapter)}
}

func (s *GameService) Register(a game.Adapter) {
    s.adapters[a.ID()] = a
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/service/game_service.go
git commit -m "feat: add GameService for context management"
```
