package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	configPath string
}

var errConfigPathEmpty = errors.New("config path must not be empty")

type State struct {
	OrderedIDs []string `json:"orderedIds"`
}

// New opens (or creates) the store at the given config file path.
func New(configPath string) (*Store, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, errConfigPathEmpty
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute config path %q: %w", configPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o750); err != nil {
		return nil, fmt.Errorf("create loadorder directory for %q: %w", absPath, err)
	}

	return &Store{configPath: absPath}, nil
}

func (s *Store) Load() (State, error) {
	content, err := os.ReadFile(s.configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{OrderedIDs: []string{}}, nil
		}
		return State{}, fmt.Errorf("read loadorder file %q: %w", s.configPath, err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return State{OrderedIDs: []string{}}, nil
	}

	var state State
	if err := json.Unmarshal(content, &state); err != nil {
		return State{}, fmt.Errorf("decode loadorder file %q: %w", s.configPath, err)
	}
	if state.OrderedIDs == nil {
		state.OrderedIDs = []string{}
	}

	return state, nil
}

func (s *Store) Save(state State) error {
	if state.OrderedIDs == nil {
		state.OrderedIDs = []string{}
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode loadorder state for %q: %w", s.configPath, err)
	}
	payload = append(payload, '\n')

	tmpPath := s.configPath + ".tmp"
	if err = os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write temporary loadorder file %q: %w", tmpPath, err)
	}

	if err = os.Rename(tmpPath, s.configPath); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf(
				"replace loadorder file %q: %w; cleanup temp %q: %s",
				s.configPath,
				err,
				tmpPath,
				removeErr.Error(),
			)
		}
		return fmt.Errorf("replace loadorder file %q: %w", s.configPath, err)
	}

	return nil
}

func (s *Store) ConfigPath() string {
	return s.configPath
}
