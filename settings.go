package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type appSettings struct {
	ModsDir                    string   `json:"mods_dir,omitempty"`
	GameExe                    string   `json:"game_exe,omitempty"`
	GameArgs                   []string `json:"game_args,omitempty"`
	LauncherActivePlaysetIndex *int     `json:"launcher_active_playset_index,omitempty"`
}

func loadSettings(path string) (appSettings, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return appSettings{}, nil
		}
		return appSettings{}, fmt.Errorf("read settings %q: %w", path, err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return appSettings{}, nil
	}

	var settings appSettings
	if err := json.Unmarshal(content, &settings); err != nil {
		return appSettings{}, fmt.Errorf("decode settings %q: %w", path, err)
	}

	return settings, nil
}

func saveSettings(path string, settings appSettings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create settings dir for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings %q: %w", path, err)
	}
	payload = append(payload, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return fmt.Errorf("write temporary settings file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace settings file %q: %w", path, err)
	}

	return nil
}
