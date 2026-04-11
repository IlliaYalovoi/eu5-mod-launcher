package eu5

import (
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/repo"
)

type Adapter struct {
	playsets repo.PlaysetRepository
}

func NewAdapter(playsets repo.PlaysetRepository) *Adapter {
	if playsets == nil {
		playsets = repo.NewFilePlaysetRepository()
	}
	return &Adapter{playsets: playsets}
}

func (*Adapter) GameID() game.GameID {
	return game.GameIDEU5
}

func (*Adapter) Descriptor() game.GameDescriptor {
	return game.GameDescriptor{ID: game.GameIDEU5, DisplayName: "Europa Universalis V"}
}

func (*Adapter) DiscoverPaths() (loadorder.GamePaths, error) {
	return loadorder.DiscoverGamePaths()
}

func (a *Adapter) ListModLists(playsetsPath string) ([]string, int, error) {
	return a.playsets.ListPlaysets(playsetsPath)
}

func (a *Adapter) ImportModList(playsetsPath string, listIndex int) (loadorder.State, map[string]string, error) {
	return a.playsets.LoadState(playsetsPath, listIndex)
}

func (a *Adapter) ExportModList(
	playsetsPath string,
	listIndex int,
	state loadorder.State,
	modPathByID map[string]string,
) error {
	return a.playsets.SaveState(playsetsPath, listIndex, state, modPathByID)
}

