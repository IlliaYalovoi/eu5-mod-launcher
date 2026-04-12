# BACKEND REFACTOR: IMPLEMENTATION PLAN

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor backend domain models and graph logic to support hierarchical mod groups and rule-based sorting.

**Architecture:**
- **Domain**: Shift `LoadOrder` to an "Active-Only" model (EU5 style).
- **Constraints**: Extend `Constraint` to target both `ModID` and `GroupID`.
- **Logic**: Two-stage topological sort (Groups -> Mods within Groups).

**Tech Stack:** Go (Standard Library + internal domain/repo).

---

### Task 1: Domain Refactor (Types & Structs)

**Files:**
- Modify: `internal/domain/constraint.go`
- Modify: `internal/domain/loadorder.go`
- Modify: `internal/domain/mod.go`

- [ ] **Step 1: Update `Constraint` struct**
```go
// internal/domain/constraint.go
type TargetType string
const (
	TargetMod   TargetType = "mod"
	TargetGroup TargetType = "group"
)

type Constraint struct {
	Type     ConstraintType `json:"type"`
	FromID   string         `json:"fromId"`
	FromType TargetType     `json:"fromType"`
	ToID     string         `json:"toId"`
	ToType   TargetType     `json:"toType"`
}
```

- [ ] **Step 2: Update `LoadOrder`**
```go
// internal/domain/loadorder.go
type LoadOrder struct {
	GameID       GameID       `json:"gameId"`
	PlaysetIdx   PlaysetIndex `json:"playsetIdx"`
	ActiveModIDs []string     `json:"activeModIds"` // Filtered list
}
```

- [ ] **Step 3: Update `Mod` helper**
```go
// internal/domain/mod.go
// Ensure CategoryID is used for Group logic
type GroupID string
```

- [ ] **Step 4: Verify with `go build ./internal/domain/...`**

- [ ] **Step 5: Commit**
```bash
git add internal/domain/*.go
git commit -m "refactor(backend): update domain models for rule-based hierarchy"
```

---

### Task 2: Repository & Layout Persistence

**Files:**
- Modify: `internal/repo/layout_repo.go`

- [ ] **Step 1: Update `LauncherCategoryData`**
```go
// internal/repo/layout_repo.go
type LauncherCategoryData struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	ModIDs []string `json:"modIds"`
}

// Add Global Group Constraints to LauncherLayoutData
type LauncherLayoutData struct {
	Groups      []LauncherCategoryData `json:"groups"`
	Constraints []domain.Constraint    `json:"constraints"` // Group-to-Group rules
}
```

- [ ] **Step 2: Update `Load` and `Save` to handle new fields**

- [ ] **Step 3: Commit**
```bash
git add internal/repo/layout_repo.go
git commit -m "feat(backend): implement hierarchical persistence in layout.json"
```

---

### Task 4: Graph Sorter Rewrite (The Core)

**Files:**
- Modify: `internal/launcher/graph.go`

- [ ] **Step 1: Implement Group sorting**
Create a subset graph containing only `GroupID` nodes and their constraints. Run topological sort to get the order of groups.

- [ ] **Step 2: Implement Intra-Group sorting**
For each group, collect its `ModIDs`. Apply mod-specific constraints that reference these IDs.

- [ ] **Step 3: Handle Inter-Group Mod constraints**
Merge sorted group results. Check for valid ordering.

- [ ] **Step 4: Verify with `go build ./...`**

- [ ] **Step 5: Commit**
```bash
git add internal/launcher/graph.go
git commit -m "feat(backend): implement two-stage topological sort for groups and mods"
```
