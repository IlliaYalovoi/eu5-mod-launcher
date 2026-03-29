package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SettingsService struct{}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

func (s *SettingsService) EffectivePath(custom, auto string) string {
	if strings.TrimSpace(custom) != "" {
		return custom
	}
	return auto
}

func (s *SettingsService) NormalizeModsDir(path string) (string, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return "", nil
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("resolve mods dir %q: %w", clean, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("mods dir %q does not exist: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("mods dir %q is not a directory", abs)
	}
	return abs, nil
}

func (s *SettingsService) NormalizeGameExe(path string) (string, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return "", nil
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("resolve game executable %q: %w", clean, err)
	}
	if !strings.EqualFold(filepath.Ext(abs), ".exe") {
		return "", fmt.Errorf("game executable %q must be an .exe file", abs)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("game executable %q not accessible: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("game executable %q is a directory", abs)
	}
	return abs, nil
}

func (s *SettingsService) ShouldLaunchViaSteam(goos, exePath string) bool {
	if goos != "windows" {
		return false
	}
	normalized := strings.ToLower(filepath.ToSlash(exePath))
	return strings.Contains(normalized, "/steamapps/common/europa universalis v/")
}
