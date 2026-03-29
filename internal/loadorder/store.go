package loadorder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store persists and retrieves load order state from a JSON file.
type Store struct {
	configPath string
}

// State is the persisted load order format.
type State struct {
	OrderedIDs []string `json:"ordered_ids"` // enabled mods in load order
}

// New opens (or creates) the store at the given config file path.
func New(configPath string) (*Store, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, errors.New("config path must not be empty")
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute config path %q: %w", configPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, fmt.Errorf("create loadorder directory for %q: %w", absPath, err)
	}

	return &Store{configPath: absPath}, nil
}

// Load reads current state from disk.
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

// Save writes state to disk atomically (write to temp file, rename).
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
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return fmt.Errorf("write temporary loadorder file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, s.configPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace loadorder file %q: %w", s.configPath, err)
	}

	return nil
}

// ConfigPath returns the resolved absolute path used by this store.
func (s *Store) ConfigPath() string {
	return s.configPath
}
