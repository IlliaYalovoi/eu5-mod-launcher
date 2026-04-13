package eu5

import (
	"eu5-mod-launcher/internal/game"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	adapterID          = "eu5"
	steamWorkshopAppID = "3450310"
)

type Adapter struct{}

func (s *Adapter) ID() string {
	return adapterID
}

func (s *Adapter) DetectInstances() ([]game.Instance, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home: %w", err)
	}

	// Docs/User config path
	// Paradox games usually put files in Documents/Paradox Interactive/Game Name
	// Based on paths.go logic
	userConfigPath := filepath.Join(home, "Documents", "Paradox Interactive", "Europa Universalis V")

	libRoots := discoverSteamLibraryRoots()

	var installPath string
	var exePath string

	for _, root := range libRoots {
		candidate := filepath.Join(root, "steamapps", "common", "Europa Universalis V")
		candidateExe := filepath.Join(candidate, "binaries", "eu5.exe")
		if fileExists(candidateExe) {
			installPath = candidate
			exePath = candidateExe
			break
		}
	}

	instance := game.Instance{
		GameID:          adapterID,
		InstallPath:     installPath,
		UserConfigPath:  userConfigPath,
		LocalModsDir:    filepath.Join(userConfigPath, "mod"),
		WorkshopModDirs: discoverWorkshopModDirs(steamWorkshopAppID),
		GameExePath:     exePath,
	}

	return []game.Instance{instance}, nil
}

func (s *Adapter) LoadMods(inst game.Instance) ([]game.ModEntry, error) {
	return nil, nil // Task 2 focus on discovery
}

func (s *Adapter) LoadPlaysets(inst game.Instance) ([]game.Playset, error) {
	return nil, nil // Task 2 focus on discovery
}

func (s *Adapter) SavePlayset(inst game.Instance, p game.Playset) error {
	return nil // Task 2 focus on discovery
}

// Helpers moved from paths.go

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

func discoverSteamLibraryRoots() []string {
	steamRoot := ""
	if runtime.GOOS == "windows" {
		fallbacks := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Steam"),
			filepath.Join(os.Getenv("ProgramFiles"), "Steam"),
		}
		for _, candidate := range fallbacks {
			if dirExists(candidate) {
				steamRoot = filepath.Clean(candidate)
				break
			}
		}
	}

	if steamRoot == "" {
		return nil
	}

	libraryFoldersPath := filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf")
	libraryRoots := append([]string{steamRoot}, parseLibraryFoldersVDF(libraryFoldersPath)...)
	seen := make(map[string]struct{}, len(libraryRoots))
	out := make([]string, 0, len(libraryRoots))

	for _, root := range libraryRoots {
		if strings.TrimSpace(root) == "" {
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
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}

	matches := regexp.MustCompile(`"path"\s*"([^"]+)"`).FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]string, 0, len(matches))
	for _, match := range matches {
		raw := match[1]
		raw = strings.ReplaceAll(raw, `\\`, `\`)
		if strings.TrimSpace(raw) == "" {
			continue
		}
		out = append(out, filepath.Clean(raw))
	}

	return out
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
