package service

import (
	"errors"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/repo"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingConstraintsRepo struct{}

var errWriteFailed = errors.New("write failed")

func (*failingConstraintsRepo) Load(_ string) (*graph.Graph, error) {
	return graph.New(), nil
}

func (*failingConstraintsRepo) Save(_ string, _ *graph.Graph) error {
	return errWriteFailed
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
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTypeMismatch)
}

func TestConstraintsServiceAddLoadLastCategoryAllowsEmptyMembers(t *testing.T) {
	svc := newConstraintsServiceForTest(t)
	require.NoError(t, svc.AddLoadLast("category:graphics"))
	all := svc.All()
	require.Len(t, all, 1)
	assert.Equal(t, graph.ConstraintTypeLast, all[0].Type)
	assert.Equal(t, "category:graphics", all[0].ModID)
}

func TestConstraintsServiceAddConstraintExpandsBundleTargets(t *testing.T) {
	svc := newConstraintsServiceForTest(t)
	require.NoError(t, svc.AddConstraint("bundle:a", "bundle:b"))
	all := svc.All()
	assert.Len(t, all, 2)
}

func TestConstraintsServiceSaveFailureRollsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "constraints.json")
	svc := NewConstraintsService(graph.New(), path, &failingConstraintsRepo{}, func(target string) []string {
		if target == "bundle:a" {
			return []string{"mod1", "mod2"}
		}
		return []string{target}
	}, domain.IsCategoryID)

	assert.Error(t, svc.AddConstraint("bundle:a", "mod3"))
	assert.Empty(t, svc.All())
}

var _ repo.ConstraintsRepository = (*failingConstraintsRepo)(nil)
