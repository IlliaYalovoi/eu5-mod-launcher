package service

import (
	"errors"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/game"
	gameeu5 "eu5-mod-launcher/internal/game/eu5"
	"fmt"
	"strings"
)

var (
	ErrUnsupportedGame = errors.New("unsupported game")
	ErrGamePathsMissing = errors.New("game paths are missing")
)

type GameService struct {
	adapters map[domain.GameID]game.Adapter
}

func NewGameService(adapters ...game.Adapter) *GameService {
	registered := make(map[domain.GameID]game.Adapter)
	if len(adapters) == 0 {
		defaultAdapter := gameeu5.NewAdapter(nil)
		registered[defaultAdapter.GameID()] = defaultAdapter
	} else {
		for _, adapter := range adapters {
			if adapter == nil {
				continue
			}
			registered[adapter.GameID()] = adapter
		}
	}
	return &GameService{adapters: registered}
}

func (s *GameService) ResolveAdapter(id domain.GameID) (game.Adapter, error) {
	adapter, ok := s.adapters[id]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedGame, id)
	}
	return adapter, nil
}

func (s *GameService) DiscoverPaths(id domain.GameID) (domain.GamePaths, error) {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return domain.GamePaths{}, err
	}
	paths, err := adapter.DiscoverPaths()
	if err != nil {
		return domain.GamePaths{}, err
	}
	if strings.TrimSpace(paths.PlaysetsPath) == "" {
		return domain.GamePaths{}, fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return paths, nil
}

func (s *GameService) ListModLists(id domain.GameID, playsetsPath string) ([]string, int, error) {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return nil, -1, err
	}
	if strings.TrimSpace(playsetsPath) == "" {
		return nil, -1, fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	names, idx, err := adapter.PlaysetRepo().ListPlaysets(playsetsPath)
	return names, int(idx), err
}

func (s *GameService) ImportModList(
	id domain.GameID,
	playsetsPath string,
	listIndex int,
) (domain.LoadOrder, map[string]string, error) {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return domain.LoadOrder{}, nil, err
	}
	if strings.TrimSpace(playsetsPath) == "" {
		return domain.LoadOrder{}, nil, fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return adapter.PlaysetRepo().LoadState(playsetsPath, domain.PlaysetIndex(listIndex))
}

func (s *GameService) ExportModList(
	id domain.GameID,
	playsetsPath string,
	listIndex int,
	order domain.LoadOrder,
	modPathByID map[string]string,
) error {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(playsetsPath) == "" {
		return fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return adapter.PlaysetRepo().SaveState(playsetsPath, domain.PlaysetIndex(listIndex), order, modPathByID)
}
