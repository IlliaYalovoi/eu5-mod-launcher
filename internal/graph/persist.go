package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveConstraints stores constraints as a JSON array of Constraint structs.
func SaveConstraints(path string, g *Graph) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create constraints directory for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(g.All(), "", "  ")
	if err != nil {
		return fmt.Errorf("encode constraints for %q: %w", path, err)
	}
	payload = append(payload, '\n')

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write constraints file %q: %w", path, err)
	}

	return nil
}

// LoadConstraints loads constraints from a JSON array and returns a graph.
func LoadConstraints(path string) (*Graph, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, fmt.Errorf("read constraints file %q: %w", path, err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return New(), nil
	}

	var constraints []Constraint
	if err := json.Unmarshal(content, &constraints); err != nil {
		var legacy []struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if legacyErr := json.Unmarshal(content, &legacy); legacyErr != nil {
			return nil, fmt.Errorf("decode constraints file %q: %w", path, err)
		}

		constraints = make([]Constraint, 0, len(legacy))
		for _, item := range legacy {
			constraints = append(constraints, Constraint{Type: ConstraintTypeAfter, From: item.From, To: item.To})
		}
	}

	g := New()
	for _, c := range constraints {
		typ := c.Type
		if typ == "" {
			typ = ConstraintTypeAfter
		}
		switch typ {
		case ConstraintTypeFirst:
			if c.ModID != "" {
				g.AddFirst(c.ModID)
			}
		case ConstraintTypeLast:
			if c.ModID != "" {
				g.AddLast(c.ModID)
			}
		default:
			if c.From != "" && c.To != "" {
				g.Add(c.From, c.To)
			}
		}
	}

	return g, nil
}
