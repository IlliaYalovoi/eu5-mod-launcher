package service

import (
	"errors"
	"eu5-mod-launcher/internal/game"
	gameeu5 "eu5-mod-launcher/internal/game/eu5"
	"eu5-mod-launcher/internal/loadorder"
	"fmt"
	"strings"
)

var (
	ErrUnsupportedGame = errors.New("unsupported game")
	ErrGamePathsMissing = errors.New("game paths are missing")
)

type GameService struct {
	adapters map[game.GameID]game.ModListAdapter
}

func NewGameService(adapters ...game.ModListAdapter) *GameService {
	registered := make(map[game.GameID]game.ModListAdapter)
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

func (s *GameService) ResolveAdapter(id game.GameID) (game.ModListAdapter, error) {
	adapter, ok := s.adapters[id]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedGame, id)
	}
	return adapter, nil
}

func (s *GameService) DiscoverPaths(id game.GameID) (loadorder.GamePaths, error) {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return loadorder.GamePaths{}, err
	}
	paths, err := adapter.DiscoverPaths()
	if err != nil {
		return loadorder.GamePaths{}, err
	}
	if strings.TrimSpace(paths.PlaysetsPath) == "" {
		return loadorder.GamePaths{}, fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return paths, nil
}

func (s *GameService) ListModLists(id game.GameID, playsetsPath string) ([]string, int, error) {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return nil, -1, err
	}
	if strings.TrimSpace(playsetsPath) == "" {
		return nil, -1, fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return adapter.ListModLists(playsetsPath)
}

func (s *GameService) ImportModList(
	id game.GameID,
	playsetsPath string,
	listIndex int,
) (loadorder.State, map[string]string, error) {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return loadorder.State{}, nil, err
	}
	if strings.TrimSpace(playsetsPath) == "" {
		return loadorder.State{}, nil, fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return adapter.ImportModList(playsetsPath, listIndex)
}

func (s *GameService) ExportModList(
	id game.GameID,
	playsetsPath string,
	listIndex int,
	state loadorder.State,
	modPathByID map[string]string,
) error {
	adapter, err := s.ResolveAdapter(id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(playsetsPath) == "" {
		return fmt.Errorf("%w: %s playsets path", ErrGamePathsMissing, id)
	}
	return adapter.ExportModList(playsetsPath, listIndex, state, modPathByID)
}

