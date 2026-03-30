package service

import (
	"errors"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/repo"
	"fmt"
)

type PlaysetService struct {
	repo repo.PlaysetRepository
}

var errPlaysetIndexOutOfRange = errors.New("playset index is out of range")

func NewPlaysetService(repository repo.PlaysetRepository) *PlaysetService {
	if repository == nil {
		repository = repo.NewFilePlaysetRepository()
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

func (s *PlaysetService) List(path string) ([]string, int, error) {
	return s.repo.ListPlaysets(path)
}

func (s *PlaysetService) Load(path string, index int) (loadorder.State, map[string]string, error) {
	return s.repo.LoadState(path, index)
}

func (s *PlaysetService) Save(path string, index int, state loadorder.State, modPathByID map[string]string) error {
	return s.repo.SaveState(path, index, state, modPathByID)
}
