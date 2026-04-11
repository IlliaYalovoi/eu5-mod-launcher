package service

import (
	"errors"
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/loadorder"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeModListAdapter struct {
	id               game.GameID
	paths            loadorder.GamePaths
	discoverErr      error
	listNames        []string
	listActive       int
	listErr          error
	importState      loadorder.State
	importPathByID   map[string]string
	importErr        error
	exportedPath     string
	exportedIndex    int
	exportedState    loadorder.State
	exportedPathByID map[string]string
	exportErr        error
}

func (f *fakeModListAdapter) GameID() game.GameID { return f.id }
func (f *fakeModListAdapter) Descriptor() game.GameDescriptor {
	return game.GameDescriptor{ID: f.id, DisplayName: string(f.id)}
}
func (f *fakeModListAdapter) DiscoverPaths() (loadorder.GamePaths, error) { return f.paths, f.discoverErr }
func (f *fakeModListAdapter) ListModLists(_ string) ([]string, int, error) { return f.listNames, f.listActive, f.listErr }
func (f *fakeModListAdapter) ImportModList(_ string, _ int) (loadorder.State, map[string]string, error) {
	return f.importState, f.importPathByID, f.importErr
}
func (f *fakeModListAdapter) ExportModList(path string, index int, state loadorder.State, modPathByID map[string]string) error {
	f.exportedPath = path
	f.exportedIndex = index
	f.exportedState = state
	f.exportedPathByID = modPathByID
	return f.exportErr
}

func TestGameServiceResolveAdapter_Unsupported(t *testing.T) {
	svc := NewGameService()
	_, err := svc.ResolveAdapter(game.GameIDVic3)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedGame)
}

func TestGameServiceImportExport_DelegatesToAdapter(t *testing.T) {
	adapter := &fakeModListAdapter{
		id:             game.GameID("test"),
		paths:          loadorder.GamePaths{PlaysetsPath: "x"},
		importState:    loadorder.State{OrderedIDs: []string{"a", "b"}},
		importPathByID: map[string]string{"a": "/mods/a"},
	}
	svc := NewGameService(adapter)

	state, pathByID, err := svc.ImportModList(adapter.id, "playsets.json", 2)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, state.OrderedIDs)
	assert.Equal(t, "/mods/a", pathByID["a"])

	err = svc.ExportModList(adapter.id, "playsets.json", 3, state, pathByID)
	require.NoError(t, err)
	assert.Equal(t, "playsets.json", adapter.exportedPath)
	assert.Equal(t, 3, adapter.exportedIndex)
	assert.Equal(t, []string{"a", "b"}, adapter.exportedState.OrderedIDs)
}

func TestGameServiceDiscoverPaths_ValidatesPlaysetsPath(t *testing.T) {
	svc := NewGameService(&fakeModListAdapter{id: game.GameID("test")})
	_, err := svc.DiscoverPaths(game.GameID("test"))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrGamePathsMissing)
}

func TestGameServiceDiscoverPaths_PropagatesAdapterError(t *testing.T) {
	adapterErr := errors.New("probe failed")
	svc := NewGameService(&fakeModListAdapter{id: game.GameID("test"), discoverErr: adapterErr})
	_, err := svc.DiscoverPaths(game.GameID("test"))
	require.Error(t, err)
	assert.ErrorIs(t, err, adapterErr)
}

