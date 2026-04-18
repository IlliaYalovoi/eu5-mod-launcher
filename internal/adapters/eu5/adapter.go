package eu5

import (
	"encoding/json"
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	adapterID          = "eu5"
	SteamWorkshopAppID = "3450310"
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
		GameID:          s.ID(),
		InstallPath:     installPath,
		UserConfigPath:  userConfigPath,
		LocalModsDir:    filepath.Join(userConfigPath, "mod"),
		WorkshopModDirs: discoverWorkshopModDirs(SteamWorkshopAppID),
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
	playsetsDir := filepath.Join(inst.UserConfigPath, "playsets")
	if err := os.MkdirAll(playsetsDir, 0755); err != nil {
		return fmt.Errorf("create playsets directory: %w", err)
	}

	playsetPath := filepath.Join(playsetsDir, fmt.Sprintf("%s.json", p.ID))
	content, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal playset: %w", err)
	}

	if err := os.WriteFile(playsetPath, content, 0644); err != nil {
		return fmt.Errorf("write playset file: %w", err)
	}

	return nil
}

func (s *Adapter) DetectVersion(inst game.Instance, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	for _, filename := range []string{"caesar_branch.txt", "clausewitz_branch.txt"} {
		content, err := os.ReadFile(filepath.Join(inst.InstallPath, filename))
		if err == nil {
			return utils.ExtractVersion(string(content)), nil
		}
	}
	return "unknown", nil
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
