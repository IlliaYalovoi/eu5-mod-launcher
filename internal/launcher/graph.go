package launcher

import (
	"encoding/json"
	"eu5-mod-launcher/internal/domain"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GraphSorter struct {
	g *domain.Graph
}

func NewGraphSorter(g *domain.Graph) *GraphSorter {
	return &GraphSorter{g: g}
}

func (s *GraphSorter) Sort(modIDs []string) ([]string, error) {
	nodes := uniqueInOrder(modIDs)
	if len(nodes) == 0 {
		return []string{}, nil
	}

	present, position := buildNodeMetadata(nodes)
	adj, indegree := buildAdjacency(nodes, s.g.All(), present)
	sortAdjacency(nodes, adj, position)

	result, err := topologicalOrder(nodes, adj, indegree)
	if err != nil {
		return nil, err
	}

	return partitionByMarkers(s.g, result)
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
	constraints []domain.Constraint,
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
		if c.Type != domain.ConstraintAfter {
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

	return nil, fmt.Errorf("%w: %s", domain.ErrCycle, strings.Join(remaining, " -> "))
}

func partitionByMarkers(g *domain.Graph, result []string) ([]string, error) {
	firstGroup := make([]string, 0)
	middleGroup := make([]string, 0, len(result))
	lastGroup := make([]string, 0)

	for i := range result {
		id := result[i]
		if g.HasFirst(id) && g.HasLast(id) {
			return nil, fmt.Errorf("%w: conflicting first/last markers on %s", domain.ErrCycle, id)
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

func SaveConstraints(path string, g *domain.Graph) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create constraints directory for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(g.All(), "", "  ")
	if err != nil {
		return fmt.Errorf("encode constraints for %q: %w", path, err)
	}
	payload = append(payload, '\n')

	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return fmt.Errorf("write constraints file %q: %w", path, err)
	}

	return nil
}

func LoadConstraints(path string) (*domain.Graph, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.NewGraph(), nil
		}
		return nil, fmt.Errorf("read constraints file %q: %w", path, err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return domain.NewGraph(), nil
	}

	constraints, err := decodeConstraints(content, path)
	if err != nil {
		return nil, err
	}

	constraintGraph := domain.NewGraph()
	applyConstraints(constraintGraph, constraints)

	return constraintGraph, nil
}

func decodeConstraints(content []byte, path string) ([]domain.Constraint, error) {
	var constraints []domain.Constraint
	if err := json.Unmarshal(content, &constraints); err != nil {
		var legacy []struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if legacyErr := json.Unmarshal(content, &legacy); legacyErr != nil {
			return nil, fmt.Errorf("decode constraints file %q: %w", path, err)
		}

		constraints = make([]domain.Constraint, 0, len(legacy))
		for i := range legacy {
			item := legacy[i]
			constraints = append(constraints, domain.Constraint{Type: domain.ConstraintAfter, From: item.From, To: item.To})
		}
	}

	return constraints, nil
}

func applyConstraints(constraintGraph *domain.Graph, constraints []domain.Constraint) {
	for i := range constraints {
		constraint := constraints[i]
		typ := constraint.Type
		if typ == "" {
			typ = domain.ConstraintAfter
		}
		switch typ {
		case domain.ConstraintFirst:
			if constraint.ModID != "" {
				constraintGraph.AddFirst(constraint.ModID)
			}
		case domain.ConstraintLast:
			if constraint.ModID != "" {
				constraintGraph.AddLast(constraint.ModID)
			}
		default:
			if constraint.From != "" && constraint.To != "" {
				constraintGraph.Add(constraint.From, constraint.To)
			}
		}
	}
}
