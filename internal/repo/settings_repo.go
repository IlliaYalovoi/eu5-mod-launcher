package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GameSettingsData struct {
	ModsDir string `json:"modsDir,omitempty"`
	GameExe string `json:"gameExe,omitempty"`
}

type AppSettingsData struct {
	GameArgs                   []string                    `json:"gameArgs,omitempty"`
	LauncherActivePlaysetIndex *int                        `json:"launcherActivePlaysetIndex,omitempty"`
	Games                      map[string]GameSettingsData `json:"games,omitempty"`
}

type SettingsRepository interface {
	Load(path string) (AppSettingsData, error)
	Save(path string, settings AppSettingsData) error
}

type FileSettingsRepository struct{}

func NewFileSettingsRepository() *FileSettingsRepository {
	return &FileSettingsRepository{}
}

func (*FileSettingsRepository) Load(path string) (AppSettingsData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AppSettingsData{}, nil
		}
		return AppSettingsData{}, fmt.Errorf("read settings %q: %w", path, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		return AppSettingsData{}, nil
	}
	var settings AppSettingsData
	if err := json.Unmarshal(content, &settings); err != nil {
		return AppSettingsData{}, fmt.Errorf("decode settings %q: %w", path, err)
	}
	return settings, nil
}

func (*FileSettingsRepository) Save(path string, settings AppSettingsData) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create settings dir for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings %q: %w", path, err)
	}
	payload = append(payload, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write temporary settings file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf(
				"replace settings file %q: %w; cleanup temp %q: %s",
				path,
				err,
				tmpPath,
				removeErr.Error(),
			)
		}
		return fmt.Errorf("replace settings file %q: %w", path, err)
	}
	return nil
}
