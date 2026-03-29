package service

import (
	"errors"
	"path/filepath"
	"testing"

	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/repo"
)

type failingConstraintsRepo struct{}

func (r *failingConstraintsRepo) Load(path string) (*graph.Graph, error) {
	return graph.New(), nil
}

func (r *failingConstraintsRepo) Save(path string, g *graph.Graph) error {
	return errors.New("write failed")
}

func newConstraintsServiceForTest(t *testing.T) *ConstraintsService {
	t.Helper()
	path := filepath.Join(t.TempDir(), "constraints.json")
	return NewConstraintsService(graph.New(), path, nil, func(target string) []string {
		if target == "bundle:a" {
			return []string{"mod1", "mod2"}
		}
		if target == "bundle:b" {
			return []string{"mod3"}
		}
		if target == "empty" {
			return nil
		}
		return []string{target}
	}, domain.IsCategoryID)
}

func TestConstraintsServiceAddConstraintRejectsMixedTypes(t *testing.T) {
	svc := newConstraintsServiceForTest(t)
	err := svc.AddConstraint("category:graphics", "modA")
	if err == nil {
		t.Fatalf("AddConstraint() error = nil, want error")
	}
	if !errors.Is(err, domain.ErrTypeMismatch) {
		t.Fatalf("AddConstraint() error = %v, want ErrTypeMismatch", err)
	}
}

func TestConstraintsServiceAddLoadLastCategoryAllowsEmptyMembers(t *testing.T) {
	svc := newConstraintsServiceForTest(t)
	if err := svc.AddLoadLast("category:graphics"); err != nil {
		t.Fatalf("AddLoadLast() error = %v", err)
	}
	all := svc.All()
	if len(all) != 1 {
		t.Fatalf("All() len = %d, want 1", len(all))
	}
	if all[0].Type != graph.ConstraintTypeLast || all[0].ModID != "category:graphics" {
		t.Fatalf("All()[0] = %#v, want last on category:graphics", all[0])
	}
}

func TestConstraintsServiceAddConstraintExpandsBundleTargets(t *testing.T) {
	svc := newConstraintsServiceForTest(t)
	if err := svc.AddConstraint("bundle:a", "bundle:b"); err != nil {
		t.Fatalf("AddConstraint() error = %v", err)
	}
	all := svc.All()
	if len(all) != 2 {
		t.Fatalf("All() len = %d, want 2", len(all))
	}
}

func TestConstraintsServiceSaveFailureRollsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "constraints.json")
	svc := NewConstraintsService(graph.New(), path, &failingConstraintsRepo{}, func(target string) []string {
		if target == "bundle:a" {
			return []string{"mod1", "mod2"}
		}
		return []string{target}
	}, domain.IsCategoryID)

	if err := svc.AddConstraint("bundle:a", "mod3"); err == nil {
		t.Fatalf("AddConstraint() error = nil, want save failure")
	}
	if got := svc.All(); len(got) != 0 {
		t.Fatalf("All() after failed save = %v, want empty (rolled back)", got)
	}
}

var _ repo.ConstraintsRepository = (*failingConstraintsRepo)(nil)
