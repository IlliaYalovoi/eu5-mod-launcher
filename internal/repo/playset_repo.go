package repo

import (
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/loadorder"
)

type PlaysetRepository interface {
	ListPlaysets(path string) ([]string, int, error)
	LoadState(path string, index int) (loadorder.State, map[string]string, error)
	SaveState(path string, index int, state loadorder.State, modPathByID map[string]string) error
}

type FilePlaysetRepository struct{}

func NewFilePlaysetRepository() *FilePlaysetRepository {
	return &FilePlaysetRepository{}
}

func (*FilePlaysetRepository) ListPlaysets(path string) ([]string, int, error) {
	return loadorder.ListPlaysets(path)
}

func (*FilePlaysetRepository) LoadState(path string, index int) (loadorder.State, map[string]string, error) {
	return loadorder.LoadStateFromPlaysets(path, index)
}

func (r *FilePlaysetRepository) SaveState(
	path string,
	index int,
	state loadorder.State,
	modPathByID map[string]string,
) error {
	return loadorder.SaveStateToPlaysets(path, index, state, modPathByID)
}

type LegacyAdapter interface {
	LoadPlaysets(inst game.Instance) ([]game.Playset, error)
	ID() string
}

type SqlitePlaysetRepository struct {
	adapter LegacyAdapter
	inst    game.Instance
}

func NewSqlitePlaysetRepository(adapter LegacyAdapter, inst game.Instance) *SqlitePlaysetRepository {
	return &SqlitePlaysetRepository{adapter: adapter, inst: inst}
}

func (r *SqlitePlaysetRepository) ListPlaysets(path string) ([]string, int, error) {
	playsets, err := r.adapter.LoadPlaysets(r.inst)
	if err != nil {
		return nil, -1, err
	}
	names := make([]string, len(playsets))
	for i, p := range playsets {
		names[i] = p.Name
	}
	// TODO: Get active from DB
	return names, 0, nil
}

func (r *SqlitePlaysetRepository) LoadState(path string, index int) (loadorder.State, map[string]string, error) {
	playsets, err := r.adapter.LoadPlaysets(r.inst)
	if err != nil {
		return loadorder.State{}, nil, err
	}
	if index < 0 || index >= len(playsets) {
		return loadorder.State{}, nil, nil
	}
	p := playsets[index]
	ids := make([]string, len(p.Entries))
	for i, e := range p.Entries {
		ids[i] = e.ID
	}
	// Fix: Load entries returns enabled mods
	return loadorder.State{OrderedIDs: ids}, map[string]string{}, nil
}

func (r *SqlitePlaysetRepository) SaveState(path string, index int, state loadorder.State, modPathByID map[string]string) error {
	return nil // SQLite saving TBD
}
