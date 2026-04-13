package loadorder

import (
	"errors"
	"eu5-mod-launcher/internal/adapters/eu5"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var errAppDataNotSet = errors.New("APPDATA is not set")

// GamePaths groups auto-discovered EU5 locations used by the launcher.
type GamePaths struct {
	PlaysetsPath    string
	LocalModsDir    string
	WorkshopModDirs []string
	GameExePath     string
}

// DefaultConfigPath returns the platform-appropriate path for the config file.
// Windows: %APPDATA%\EU5ModLauncher\loadorder.json
// Linux:   $XDG_CONFIG_HOME/eu5-mod-launcher/loadorder.json
//
//	(falls back to $HOME/.config/... if XDG not set)
func DefaultConfigPath() (string, error) {
	return defaultConfigPathForOS(runtime.GOOS, os.Getenv)
}

func defaultConfigPathForOS(goos string, getenv func(string) string) (string, error) {
	switch goos {
	case "windows":
		appData := getenv("APPDATA")
		if appData == "" {
			return "", errAppDataNotSet
		}
		return filepath.Join(appData, "EU5ModLauncher", "loadorder.json"), nil
	case "linux":
		base := getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve user home for linux config path: %w", err)
			}
			base = filepath.Join(home, ".config")
			return filepath.Join(base, "eu5-mod-launcher", "loadorder.json"), nil
		}
		return path.Join(base, "eu5-mod-launcher", "loadorder.json"), nil
	default:
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolve user config dir for %q: %w", goos, err)
		}
		return filepath.Join(base, "eu5-mod-launcher", "loadorder.json"), nil
	}
}

// DiscoverGamePaths resolves standard EU5 locations and Steam workshop roots.
func DiscoverGamePaths() (GamePaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return GamePaths{}, fmt.Errorf("resolve user home for game paths: %w", err)
	}

	docsRoot := filepath.Join(home, "Documents", "Paradox Interactive", "Europa Universalis V")

	return GamePaths{
		PlaysetsPath:    filepath.Join(docsRoot, "playsets.json"),
		LocalModsDir:    filepath.Join(docsRoot, "mod"),
		WorkshopModDirs: discoverWorkshopModDirs(eu5.SteamWorkshopAppID),
		GameExePath:     discoverGameExePath(),
	}, nil
}

var steamInstallPathFinder func() string

func findSteamInstallPath() string {
	if steamInstallPathFinder == nil {
		// Ensure it's initialized for functional discovery in this file
		return ""
	}
	return steamInstallPathFinder()
}

func discoverWorkshopModDirs(appID string) []string {
	libraryRoots := discoverSteamLibraryRoots()
	if len(libraryRoots) == 0 {
		return nil
	}

	out := make([]string, 0, len(libraryRoots))
	for _, cleanRoot := range libraryRoots {
		workshopDir := filepath.Join(cleanRoot, "steamapps", "workshop", "content", appID)
		if dirExists(workshopDir) {
			out = append(out, workshopDir)
		}
	}

	return out
}

func discoverGameExePath() string {
	libraryRoots := discoverSteamLibraryRoots()
	for _, root := range libraryRoots {
		candidate := filepath.Join(root, "steamapps", "common", "Europa Universalis V", "binaries", "eu5.exe")
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func discoverSteamLibraryRoots() []string {
	steamRoot := findSteamInstallPath()
	if steamRoot == "" {
		return nil
	}

	libraryFoldersPath := filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf")
	libraryRoots := append([]string{steamRoot}, parseLibraryFoldersVDF(libraryFoldersPath)...)
	seen := make(map[string]struct{}, len(libraryRoots))
	out := make([]string, 0, len(libraryRoots))

	for _, root := range libraryRoots {
		if root == "" {
			continue
		}

		cleanRoot := filepath.Clean(root)
		if _, ok := seen[cleanRoot]; ok {
			continue
		}
		seen[cleanRoot] = struct{}{}
		out = append(out, cleanRoot)
	}

	return out
}

func parseLibraryFoldersVDF(vdfPath string) []string {
	// Re-use logic since we are keeping it functional for now
	// but using common helpers involves too much refactor of internal/adapters
	// so keep simple copy here until DiscoverGamePaths is gone
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}

	// Simple extraction logic from original code
	const pathKey = "\"path\""
	lines := strings.Split(string(content), "\n")
	var paths []string
	for _, line := range lines {
		if strings.Contains(line, pathKey) {
			parts := strings.Split(line, "\"")
			if len(parts) >= 4 {
				p := parts[3]
				p = strings.ReplaceAll(p, "\\\\", "\\")
				paths = append(paths, filepath.Clean(p))
			}
		}
	}
	return paths
}

func dirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	return err == nil && info.IsDir()
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}
