# Task 03 — Go: Constraint Graph & Autosort

## Goal

Implement `internal/graph` package. It stores directed "loads-after" constraints between mod IDs and can produce a topologically sorted load order, detecting cycles.

## Context

A constraint is a directed edge: `A → B` means "A loads after B" (B comes first). The autosort takes the user's current enabled mod list and reorders it to satisfy all constraints. If the constraints form a cycle, autosort returns an error describing the cycle.

## Deliverables

### `internal/graph/graph.go`

```go
type Graph struct {
    // unexported
}

type Constraint struct {
    From string `json:"from"` // this mod ID...
    To   string `json:"to"`  // ...loads after this mod ID
}

// New creates an empty graph.
func New() *Graph

// Add adds a loads-after constraint: `from` will be placed after `to`.
// Adding a duplicate is a no-op.
func (g *Graph) Add(from, to string)

// Remove removes a specific constraint. No-op if not present.
func (g *Graph) Remove(from, to string)

// ConstraintsFor returns all constraints where from == modID or to == modID.
func (g *Graph) ConstraintsFor(modID string) []Constraint

// All returns all constraints.
func (g *Graph) All() []Constraint
```

### `internal/graph/sort.go`

```go
// Sort takes an ordered list of enabled mod IDs and returns a new ordering
// that satisfies all constraints in g.
// Mods not present in the input list are ignored even if they appear in constraints.
// Returns ErrCycle if constraints are unsatisfiable.
func (g *Graph) Sort(modIDs []string) ([]string, error)

// ErrCycle is returned when a topological sort is impossible.
// The error message should name the mods involved.
var ErrCycle = ...
```

### `internal/graph/persist.go`

```go
// MarshalJSON / UnmarshalJSON or explicit Save/Load for the constraint set,
// so constraints survive app restarts alongside the load order.
// Store as a JSON array of Constraint structs.
func SaveConstraints(path string, g *Graph) error
func LoadConstraints(path string) (*Graph, error)
```

## Acceptance criteria

- `Sort(["A","B","C"])` with constraint `A loads after C` returns `["C", ..., "A"]` (C before A).
- Cycle `A→B, B→A` returns `ErrCycle`.
- Mods with no constraints preserve their relative input order (stable sort behavior).
- Round-trip save/load preserves all constraints.
- Table-driven tests covering: no constraints, single constraint, multi-constraint chain, cycle.

## Notes for agent

- Use Kahn's algorithm (BFS-based topological sort) — it's straightforward to implement and easy to extract cycle info from.
- stdlib only.
- No Wails imports.
