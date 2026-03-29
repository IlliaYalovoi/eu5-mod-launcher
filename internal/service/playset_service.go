package service

import (
	"fmt"

	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/repo"
)

type PlaysetService struct {
	repo repo.PlaysetRepository
}

func NewPlaysetService(repository repo.PlaysetRepository) *PlaysetService {
	if repository == nil {
		repository = repo.NewFilePlaysetRepository()
	}
	return &PlaysetService{repo: repository}
}

func (s *PlaysetService) ResolveLauncherIndex(total, gameActive int, preferred *int) int {
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

func (s *PlaysetService) ValidateIndex(index, total int) error {
	if index < 0 || index >= total {
		return fmt.Errorf("playset index %d is out of range", index)
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
