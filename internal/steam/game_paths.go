package steam

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type GamePaths struct {
	PlaysetsPath    string
	LocalModsDir    string
	WorkshopModDirs []string
	GameExePath     string
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
		WorkshopModDirs: discoverWorkshopModDirs(SteamAppID),
		GameExePath:     discoverGameExePath(),
	}, nil
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

func DefaultConfigPath() (string, error) {
	configHome, err := xdg.ConfigFile("eu5-mod-launcher/loadorder.json")
	if err != nil {
		return "", fmt.Errorf("resolve config path via xdg: %w", err)
	}
	return configHome, nil
}
