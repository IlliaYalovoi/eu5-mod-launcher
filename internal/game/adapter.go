package game

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/repo"
)

type Adapter interface {
	GameID() domain.GameID
	Descriptor() domain.GameDescriptor
	DiscoverPaths() (domain.GamePaths, error)
	PlaysetRepo() repo.PlaysetRepo
}
