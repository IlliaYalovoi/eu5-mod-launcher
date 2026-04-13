package service

import (
	"eu5-mod-launcher/internal/game"
	"fmt"
	"sync"
)

type GameService struct {
	mu             sync.RWMutex
	adapters       map[string]game.Adapter
	activeGameID   string
	activeInstance *game.Instance
}

func NewGameService() *GameService {
	return &GameService{
		adapters: make(map[string]game.Adapter),
	}
}

func (s *GameService) Register(a game.Adapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adapters[a.ID()] = a
}

func (s *GameService) SetActiveGame(gameID string, instanceIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	adapter, ok := s.adapters[gameID]
	if !ok {
		return fmt.Errorf("no adapter registered for game: %s", gameID)
	}

	instances, err := adapter.DetectInstances()
	if err != nil {
		return fmt.Errorf("detect instances for %s: %w", gameID, err)
	}

	if instanceIndex < 0 || instanceIndex >= len(instances) {
		return fmt.Errorf("invalid instance index: %d", instanceIndex)
	}

	s.activeGameID = gameID
	s.activeInstance = &instances[instanceIndex]
	return nil
}

func (s *GameService) GetActiveInstance() (*game.Instance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.activeInstance == nil {
		return nil, fmt.Errorf("no active game instance set")
	}
	return s.activeInstance, nil
}

func (s *GameService) GetAdapter(id string) game.Adapter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.adapters[id]
}

func (s *GameService) GetAdapters() []game.Adapter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adapters := make([]game.Adapter, 0, len(s.adapters))
	for _, a := range s.adapters {
		adapters = append(adapters, a)
	}
	return adapters
}
