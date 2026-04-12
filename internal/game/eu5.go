package game

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/steam"
)

type EU5Adapter struct {
	playsets repo.PlaysetRepo
}

func NewEU5Adapter(playsets repo.PlaysetRepo) *EU5Adapter {
	if playsets == nil {
		playsets = repo.NewFilePlaysetRepo()
	}
	return &EU5Adapter{playsets: playsets}
}

func (*EU5Adapter) GameID() domain.GameID { return domain.GameIDEU5 }

func (*EU5Adapter) Descriptor() domain.GameDescriptor {
	return domain.GameDescriptor{ID: domain.GameIDEU5, DisplayName: "Europa Universalis V"}
}

func (*EU5Adapter) DiscoverPaths() (domain.GamePaths, error) {
	paths, err := steam.DiscoverGamePaths()
	if err != nil {
		return domain.GamePaths{}, err
	}
	return domain.GamePaths{
		LocalModsDir:    paths.LocalModsDir,
		PlaysetsPath:    paths.PlaysetsPath,
		WorkshopModDirs: paths.WorkshopModDirs,
		GameExePath:     paths.GameExePath,
	}, nil
}

func (a *EU5Adapter) PlaysetRepo() repo.PlaysetRepo { return a.playsets }
