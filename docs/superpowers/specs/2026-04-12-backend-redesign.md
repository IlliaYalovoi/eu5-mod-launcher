# BACKEND DESIGN: MOD CHRONICLE CORE
Date: 2026-04-12
Status: DRAFT

## 1. DATA MODELS (REVISED)

### A. The Playset (Unified)
A Playset is now strictly a **subset of enabled mods** with an associated **Constraint Graph**.
Legacy playsets (Vic3) will be "filtered" on load:
- Any mod in the legacy playset with `enabled: true` is included.
- Any mod with `enabled: false` is moved to the **Repository** (Global unassigned pool).

```go
// internal/domain/loadorder.go
type LoadOrder struct {
    GameID      GameID       `json:"gameId"`
    PlaysetIdx  PlaysetIndex `json:"playsetIdx"`
    ActiveModIDs []string    `json:"activeModIds"` // ONLY enabled mods
}
```

### B. Hierarchical Constraints
We introduce `GroupID` to the constraint system.

```go
// internal/domain/constraint.go
type TargetType string
const (
    TargetMod   TargetType = "mod"
    TargetGroup TargetType = "group"
)

type Constraint struct {
    Type       ConstraintType `json:"type"`
    FromID     string         `json:"fromId"`     // ModID or GroupID
    FromType   TargetType     `json:"fromType"`
    ToID       string         `json:"toId"`       // ModID or GroupID
    ToType     TargetType     `json:"toType"`
}
```

## 2. REPOSITORY LOGIC
The "Repository" is a virtual collection calculated on the fly:
`Repository = ListAllMods(game) - CurrentPlayset.ActiveModIDs`

### API Endpoints (Wails)
- `GetRepositoryMods() []Mod`: Returns mods not in the current playset.
- `EnableMod(modID, groupID)`: Adds mod to playset and optionally assigns to a group.
- `DisableMod(modID)`: Removes mod from playset (it implicitly returns to Repository).

## 3. THE GRAPH SORTER (DAG)
The sorter must now handle a two-stage topological sort:
1. **Sort Groups**: Based on Group-to-Group constraints.
2. **Sort Mods**: Within each group, based on Mod-to-Mod constraints.
3. **Global Constraints**: Mods can still have cross-group constraints, which may trigger a validation error if they conflict with group-level ordering.

## 4. PERSISTENCE (`layout.json`)
```json
{
  "groups": [
    { "id": "grp_1", "name": "Overhauls", "modIds": ["mod_a", "mod_b"] }
  ],
  "constraints": [
    { "type": "after", "fromId": "grp_1", "fromType": "group", "toId": "grp_2", "toType": "group" }
  ]
}
```
