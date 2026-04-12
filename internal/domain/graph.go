package domain

import "sort"

type Graph struct {
	edges map[string]map[string]struct{}
	first map[string]struct{}
	last  map[string]struct{}
}

func NewGraph() *Graph {
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

func (g *Graph) HasOutgoingAfter(modID string) bool {
	neighbors, ok := g.edges[modID]
	return ok && len(neighbors) > 0
}

func (g *Graph) HasIncomingAfter(modID string) bool {
	for _, tos := range g.edges {
		if _, ok := tos[modID]; ok {
			return true
		}
	}
	return false
}

func (g *Graph) All() []Constraint {
	out := make([]Constraint, 0)
	for from, tos := range g.edges {
		for to := range tos {
			out = append(out, Constraint{
				Type:     ConstraintAfter,
				FromID:   from,
				FromType: TargetMod,
				ToID:     to,
				ToType:   TargetMod,
			})
		}
	}
	for id := range g.first {
		out = append(out, Constraint{
			Type:     ConstraintFirst,
			FromID:   id,
			FromType: TargetMod,
		})
	}
	for id := range g.last {
		out = append(out, Constraint{
			Type:     ConstraintLast,
			FromID:   id,
			FromType: TargetMod,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		if out[i].FromID == out[j].FromID {
			return out[i].ToID < out[j].ToID
		}
		return out[i].FromID < out[j].FromID
	})

	return out
}
