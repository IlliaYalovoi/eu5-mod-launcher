package repo

import "eu5-mod-launcher/internal/domain"

type PlaysetRepo interface {
	ListPlaysets(playsetsPath string) ([]string, domain.PlaysetIndex, error)
	LoadState(playsetsPath string, idx domain.PlaysetIndex) (domain.LoadOrder, map[string]string, error)
	SaveState(playsetsPath string, idx domain.PlaysetIndex, order domain.LoadOrder, modPathByID map[string]string) error
}
