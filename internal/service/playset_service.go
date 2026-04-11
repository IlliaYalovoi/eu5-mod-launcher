package service

import (
	"errors"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/repo"
	"fmt"
)

type PlaysetService struct {
	repo repo.PlaysetRepo
}

var errPlaysetIndexOutOfRange = errors.New("playset index is out of range")

func NewPlaysetService(repository repo.PlaysetRepo) *PlaysetService {
	if repository == nil {
		repository = repo.NewFilePlaysetRepo()
	}
	return &PlaysetService{repo: repository}
}

func (*PlaysetService) ResolveLauncherIndex(total, gameActive int, preferred *int) int {
	if total <= 0 {
		return -1
	}
	if preferred != nil && *preferred >= 0 && *preferred < total {
		return *preferred
	}
	if gameActive >= 0 && gameActive < total {
		return gameActive
	}
	return 0
}

func (*PlaysetService) ValidateIndex(index, total int) error {
	if index < 0 || index >= total {
		return fmt.Errorf("%w: %d", errPlaysetIndexOutOfRange, index)
	}
	return nil
}

func (s *PlaysetService) List(path string) ([]string, domain.PlaysetIndex, error) {
	return s.repo.ListPlaysets(path)
}

func (s *PlaysetService) Load(path string, index domain.PlaysetIndex) (domain.LoadOrder, map[string]string, error) {
	return s.repo.LoadState(path, index)
}

func (s *PlaysetService) Save(path string, index domain.PlaysetIndex, order domain.LoadOrder, modPathByID map[string]string) error {
	return s.repo.SaveState(path, index, order, modPathByID)
}
