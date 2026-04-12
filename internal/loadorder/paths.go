package loadorder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"

	"eu5-mod-launcher/internal/steam"
)

const steamWorkshopAppID = steam.SteamAppID

type GamePaths struct {
	PlaysetsPath    string
	LocalModsDir    string
	WorkshopModDirs []string
	GameExePath     string
}

func DefaultConfigPath() (string, error) {
	configHome, err := xdg.ConfigFile("eu5-mod-launcher/loadorder.json")
	if err != nil {
		return "", fmt.Errorf("resolve config path via xdg: %w", err)
	}
	return configHome, nil
}

func DiscoverGamePaths() (GamePaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return GamePaths{}, fmt.Errorf("resolve user home for game paths: %w", err)
	}

	docsRoot := filepath.Join(home, "Documents", "Paradox Interactive", "Europa Universalis V")

	return GamePaths{
		PlaysetsPath:    filepath.Join(docsRoot, "playsets.json"),
		LocalModsDir:    filepath.Join(docsRoot, "mod"),
		WorkshopModDirs: discoverWorkshopModDirs(steamWorkshopAppID),
		GameExePath:     discoverGameExePath(),
	}, nil
}

func findSteamInstallPath() string {
	return steamInstallPathFinder()
}

func defaultSteamInstallPath() string {
	fallbacks := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Steam"),
		filepath.Join(os.Getenv("ProgramFiles"), "Steam"),
	}

	for _, candidate := range fallbacks {
		if dirExists(candidate) {
			return filepath.Clean(candidate)
		}
	}

	return ""
}

func discoverWorkshopModDirs(appID string) []string {
	libraryRoots := DiscoverSteamLibraryRoots()
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
	libraryRoots := DiscoverSteamLibraryRoots()
	for _, root := range libraryRoots {
		candidate := filepath.Join(root, "steamapps", "common", "Europa Universalis V", "binaries", "eu5.exe")
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func DiscoverSteamLibraryRoots() []string {
	steamRoot := findSteamInstallPath()
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

func dirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
