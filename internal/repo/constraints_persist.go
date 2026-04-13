package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eu5-mod-launcher/internal/domain"
)

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
		typ := string(constraint.Type)
		if typ == "" {
			typ = string(domain.ConstraintAfter)
		}
		switch typ {
		case string(domain.ConstraintFirst):
			if constraint.ModID != "" {
				constraintGraph.AddFirst(constraint.ModID)
			}
		case string(domain.ConstraintLast):
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
