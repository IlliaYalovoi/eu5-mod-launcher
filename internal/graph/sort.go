package graph

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrCycle is returned when a topological sort is impossible.
var ErrCycle = errors.New("constraint cycle detected")

// Sort takes an ordered list of enabled mod IDs and returns a new ordering
// that satisfies all constraints in g.
// Mods not present in the input list are ignored even if they appear in constraints.
// Returns ErrCycle if constraints are unsatisfiable.
func (g *Graph) Sort(modIDs []string) ([]string, error) {
	nodes := uniqueInOrder(modIDs)
	if len(nodes) == 0 {
		return []string{}, nil
	}

	present, position := buildNodeMetadata(nodes)
	adj, indegree := buildAdjacency(nodes, g.All(), present)
	sortAdjacency(nodes, adj, position)

	result, err := topologicalOrder(nodes, adj, indegree)
	if err != nil {
		return nil, err
	}

	return g.partitionByMarkers(result)
}

func buildNodeMetadata(nodes []string) (map[string]struct{}, map[string]int) {
	present := make(map[string]struct{}, len(nodes))
	position := make(map[string]int, len(nodes))
	for i := range nodes {
		id := nodes[i]
		present[id] = struct{}{}
		position[id] = i
	}

	return present, position
}

func buildAdjacency(
	nodes []string,
	constraints []Constraint,
	present map[string]struct{},
) (map[string][]string, map[string]int) {
	adj := make(map[string][]string, len(nodes))
	indegree := make(map[string]int, len(nodes))
	for i := range nodes {
		id := nodes[i]
		adj[id] = []string{}
		indegree[id] = 0
	}

	for i := range constraints {
		c := constraints[i]
		if c.Type != ConstraintTypeAfter {
			continue
		}

		_, fromInInput := present[c.From]
		_, toInInput := present[c.To]
		if !fromInInput || !toInInput {
			continue
		}

		adj[c.To] = append(adj[c.To], c.From)
		indegree[c.From]++
	}

	return adj, indegree
}

func sortAdjacency(nodes []string, adj map[string][]string, position map[string]int) {
	for i := range nodes {
		from := nodes[i]
		sort.Slice(adj[from], func(i, j int) bool {
			return position[adj[from][i]] < position[adj[from][j]]
		})
	}
}

func topologicalOrder(nodes []string, adj map[string][]string, indegree map[string]int) ([]string, error) {
	queue := make([]string, 0, len(nodes))
	for i := range nodes {
		id := nodes[i]
		if indegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]string, 0, len(nodes))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, next := range adj[current] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(result) == len(nodes) {
		return result, nil
	}

	remaining := make([]string, 0, len(nodes)-len(result))
	for i := range nodes {
		id := nodes[i]
		if indegree[id] > 0 {
			remaining = append(remaining, id)
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrCycle, strings.Join(remaining, " -> "))
}

func (g *Graph) partitionByMarkers(result []string) ([]string, error) {
	firstGroup := make([]string, 0)
	middleGroup := make([]string, 0, len(result))
	lastGroup := make([]string, 0)

	for i := range result {
		id := result[i]
		if g.HasFirst(id) && g.HasLast(id) {
			return nil, fmt.Errorf("%w: conflicting first/last markers on %s", ErrCycle, id)
		}
		if g.HasFirst(id) {
			firstGroup = append(firstGroup, id)
			continue
		}
		if g.HasLast(id) {
			lastGroup = append(lastGroup, id)
			continue
		}
		middleGroup = append(middleGroup, id)
	}

	sort.Strings(firstGroup)
	sort.Strings(lastGroup)

	out := make([]string, 0, len(result))
	out = append(out, firstGroup...)
	out = append(out, middleGroup...)
	out = append(out, lastGroup...)

	return out, nil
}

func uniqueInOrder(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
