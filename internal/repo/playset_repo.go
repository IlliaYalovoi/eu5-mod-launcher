package repo

import (
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/loadorder"
	"path/filepath"
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
	GetModNames(inst game.Instance) (map[string]string, error)
	LoadMods(inst game.Instance) ([]game.ModEntry, error)
	SavePlayset(inst game.Instance, p game.Playset) error
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

	mods, err := r.adapter.LoadMods(r.inst)
	if err != nil {
		return loadorder.State{}, nil, err
	}

	uuidToSteamID := make(map[string]string)
	steamIDToPath := make(map[string]string)

	for _, m := range mods {
		norm := filepath.Clean(filepath.ToSlash(m.Path))
		steamID := filepath.Base(norm)
		if steamID == "" || steamID == "." {
			continue
		}
		uuidToSteamID[m.ID] = steamID
		steamIDToPath[steamID] = m.Path
	}

	p := playsets[index]
	ids := make([]string, 0, len(p.Entries))
	for _, e := range p.Entries {
		if !e.Enabled {
			continue
		}
		steamID, ok := uuidToSteamID[e.ID]
		if !ok {
			steamID = e.ID
		}
		ids = append(ids, steamID)
	}

	return loadorder.State{OrderedIDs: ids}, steamIDToPath, nil
}

func (r *SqlitePlaysetRepository) SaveState(path string, index int, state loadorder.State, modPathByID map[string]string) error {
	playsets, err := r.adapter.LoadPlaysets(r.inst)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(playsets) {
		return nil
	}

	mods, err := r.adapter.LoadMods(r.inst)
	if err != nil {
		return err
	}

	pathToUUID := make(map[string]string)
	for _, m := range mods {
		norm := filepath.Clean(filepath.ToSlash(m.Path))
		pathToUUID[norm] = m.ID
	}

	p := playsets[index]
	newEntries := make([]game.ModEntry, 0, len(state.OrderedIDs))

	for i, steamID := range state.OrderedIDs {
		pathValue := modPathByID[steamID]
		norm := filepath.Clean(filepath.ToSlash(pathValue))

		uuid, ok := pathToUUID[norm]
		if !ok {
			for _, m := range mods {
				mNorm := filepath.Clean(filepath.ToSlash(m.Path))
				if filepath.Base(mNorm) == steamID {
					uuid = m.ID
					break
				}
			}
			if uuid == "" {
				for _, m := range mods {
					if m.ID == steamID {
						uuid = m.ID
						break
					}
				}
				if uuid == "" {
					continue
				}
			}
		}

		newEntries = append(newEntries, game.ModEntry{
			ID:       uuid,
			Enabled:  true,
			Position: i,
		})
	}

	p.Entries = newEntries
	return r.adapter.SavePlayset(r.inst, p)
}
