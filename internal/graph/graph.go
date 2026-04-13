package graph

import (
	"sort"
)

const (
	ConstraintTypeAfter = "after"
	ConstraintTypeFirst = "first"
	ConstraintTypeLast  = "last"
)

// Constraint represents one ordering rule.
// - type=after: from loads after to
// - type=first: mod_id should be pushed to the front group
// - type=last:  mod_id should be pushed to the back group.
type Constraint struct {
	Type  string `json:"type,omitempty"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
	ModID string `json:"modId,omitempty"`
}

type Graph struct {
	edges map[string]map[string]struct{}
	first map[string]struct{}
	last  map[string]struct{}
}

func New() *Graph {
	return &Graph{
		edges: make(map[string]map[string]struct{}),
		first: make(map[string]struct{}),
		last:  make(map[string]struct{}),
	}
}

func (g *Graph) Add(from, to string) {
	if g.edges[from] == nil {
		g.edges[from] = make(map[string]struct{})
	}
	g.edges[from][to] = struct{}{}
}

func (g *Graph) AddFirst(modID string) {
	g.first[modID] = struct{}{}
}

func (g *Graph) AddLast(modID string) {
	g.last[modID] = struct{}{}
}

func (g *Graph) Remove(from, to string) {
	neighbors, ok := g.edges[from]
	if !ok {
		return
	}

	delete(neighbors, to)
	if len(neighbors) == 0 {
		delete(g.edges, from)
	}
}

func (g *Graph) RemoveFirst(modID string) {
	delete(g.first, modID)
}

func (g *Graph) RemoveLast(modID string) {
	delete(g.last, modID)
}

func (g *Graph) HasFirst(modID string) bool {
	_, ok := g.first[modID]
	return ok
}

func (g *Graph) HasLast(modID string) bool {
	_, ok := g.last[modID]
	return ok
}

// HasOutgoingAfter reports whether modID has any "loads-after" dependencies.
func (g *Graph) HasOutgoingAfter(modID string) bool {
	neighbors, ok := g.edges[modID]
	return ok && len(neighbors) > 0
}

// HasIncomingAfter reports whether any mod depends on modID.
func (g *Graph) HasIncomingAfter(modID string) bool {
	for _, tos := range g.edges {
		if _, ok := tos[modID]; ok {
			return true
		}
	}
	return false
}

// ConstraintsFor returns all constraints where modID participates.
func (g *Graph) ConstraintsFor(modID string) []Constraint {
	all := g.All()
	out := make([]Constraint, 0)
	for i := range all {
		c := all[i]
		switch c.Type {
		case ConstraintTypeFirst, ConstraintTypeLast:
			if c.ModID == modID {
				out = append(out, c)
			}
		default:
			if c.From == modID || c.To == modID {
				out = append(out, c)
			}
		}
	}
	return out
}

// All returns all constraints.
func (g *Graph) All() []Constraint {
	out := make([]Constraint, 0)
	for from, tos := range g.edges {
		for to := range tos {
			out = append(out, Constraint{Type: ConstraintTypeAfter, From: from, To: to})
		}
	}
	for id := range g.first {
		out = append(out, Constraint{Type: ConstraintTypeFirst, ModID: id})
	}
	for id := range g.last {
		out = append(out, Constraint{Type: ConstraintTypeLast, ModID: id})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		if out[i].Type == ConstraintTypeAfter {
			if out[i].From == out[j].From {
				return out[i].To < out[j].To
			}
			return out[i].From < out[j].From
		}
		return out[i].ModID < out[j].ModID
	})

	return out
}
