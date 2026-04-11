package eu5

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/repo"
)

type Adapter struct {
	playsets repo.PlaysetRepo
}

func NewAdapter(playsets repo.PlaysetRepo) *Adapter {
	if playsets == nil {
		playsets = repo.NewFilePlaysetRepo()
	}
	return &Adapter{playsets: playsets}
}

func (*Adapter) GameID() domain.GameID { return domain.GameIDEU5 }

func (*Adapter) Descriptor() domain.GameDescriptor {
	return domain.GameDescriptor{ID: domain.GameIDEU5, DisplayName: "Europa Universalis V"}
}

func (*Adapter) DiscoverPaths() (domain.GamePaths, error) {
	paths, err := loadorder.DiscoverGamePaths()
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

func (a *Adapter) PlaysetRepo() repo.PlaysetRepo { return a.playsets }
