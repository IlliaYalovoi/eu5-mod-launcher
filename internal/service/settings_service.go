package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SettingsService struct{}

const windowsOS = "windows"

var (
	errModsDirNotDirectory  = errors.New("mods dir is not a directory")
	errGameExecutableNotEXE = errors.New("game executable must be an .exe file")
	errGameExecutableIsDir  = errors.New("game executable is a directory")
)

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

func (*SettingsService) EffectivePath(custom, auto string) string {
	if strings.TrimSpace(custom) != "" {
		return custom
	}
	return auto
}

func (*SettingsService) NormalizeModsDir(path string) (string, error) {
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
		return "", fmt.Errorf("%w: %q", errModsDirNotDirectory, abs)
	}
	return abs, nil
}

func (*SettingsService) NormalizeGameExe(path string) (string, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return "", nil
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("resolve game executable %q: %w", clean, err)
	}
	if !strings.EqualFold(filepath.Ext(abs), ".exe") {
		return "", fmt.Errorf("%w: %q", errGameExecutableNotEXE, abs)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("game executable %q not accessible: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%w: %q", errGameExecutableIsDir, abs)
	}
	return abs, nil
}

func (*SettingsService) ShouldLaunchViaSteam(goos, exePath string) bool {
	if goos != windowsOS {
		return false
	}
	normalized := strings.ToLower(filepath.ToSlash(exePath))
	return strings.Contains(normalized, "/steamapps/common/europa universalis v/")
}
