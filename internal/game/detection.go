package game

import (
	"os"
	"path/filepath"
	"sort"

	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/steam"
)

type DetectedGame struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	IconKey          string `json:"iconKey"`
	Detected         bool   `json:"detected"`
	InstallDir       string `json:"installDir"`
	DocumentsDir     string `json:"documentsDir"`
	NeedsManualSetup bool   `json:"needsManualSetup"`
}

type Detector struct {
	settingsRepo   repo.SettingsRepository
	supportedGames []gameInfo
}

type gameInfo struct {
	id          domain.GameID
	name        string
	iconKey     string
	appID       string
	documents   func() string
	installDirs func() []string
}

var (
	vic3SteamAppID = "529340"
	vic3DocsRoot   = func() string {
		home, _ := os.UserHomeDir()
		if home == "" {
			return ""
		}
		return filepath.Join(home, "Documents", "Paradox Interactive", "Victoria 3")
	}
	vic3InstallDirs = func() []string {
		home, _ := os.UserHomeDir()
		if home == "" {
			return nil
		}
		libraryRoots := steam.DiscoverSteamLibraryRoots()
		dirs := []string{}
		if len(libraryRoots) > 0 {
			dirs = append(dirs, filepath.Join(libraryRoots[0], "steamapps", "common", "Victoria 3"))
		}
		common := filepath.Join(home, "Games", "Paradox Interactive", "Victoria 3")
		if _, err := os.Stat(common); err == nil {
			dirs = append(dirs, common)
		}
		return dirs
	}
)

func NewDetector(settingsRepo repo.SettingsRepository) *Detector {
	if settingsRepo == nil {
		settingsRepo = repo.NewFileSettingsRepository()
	}

	return &Detector{
		settingsRepo: settingsRepo,
		supportedGames: []gameInfo{
			{
				id:        domain.GameIDEU5,
				name:      "Europa Universalis V",
				iconKey:   "eu5",
				appID:     "3450310",
				documents: eu5DocumentsDir,
				installDirs: func() []string {
					return discoverEU5InstallDirs()
				},
			},
			{
				id:        domain.GameIDVic3,
				name:      "Victoria 3",
				iconKey:   "vic3",
				appID:     vic3SteamAppID,
				documents: vic3DocsRoot,
				installDirs: vic3InstallDirs,
			},
		},
	}
}

func eu5DocumentsDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "Documents", "Paradox Interactive", "Europa Universalis V")
}

func discoverEU5InstallDirs() []string {
	libraryRoots := steam.DiscoverSteamLibraryRoots()
	dirs := make([]string, 0, len(libraryRoots))
	for _, root := range libraryRoots {
		candidate := filepath.Join(root, "steamapps", "common", "Europa Universalis V")
		if _, err := os.Stat(candidate); err == nil {
			dirs = append(dirs, candidate)
		}
	}
	return dirs
}

func (s *Detector) ListSupportedGames(settingsPath string) ([]DetectedGame, error) {
	overrides, err := s.loadOverrides(settingsPath)
	if err != nil {
		logging.Warnf("game-detection: load overrides: %v", err)
	}

	result := make([]DetectedGame, 0, len(s.supportedGames))

	for _, gi := range s.supportedGames {
		override := overrides[gi.id]

		docsDir := gi.documents()
		installDir := s.detectInstallDir(gi.installDirs())

		if override.InstallDir != "" {
			installDir = override.InstallDir
			logging.Debugf("game-detection: %s using override install dir: %s", gi.name, installDir)
		}
		if override.DocumentsDir != "" {
			docsDir = override.DocumentsDir
			logging.Debugf("game-detection: %s using override documents dir: %s", gi.name, docsDir)
		}

		detected := installDir != "" && dirExists(installDir)

		result = append(result, DetectedGame{
			ID:               string(gi.id),
			Name:             gi.name,
			IconKey:          gi.iconKey,
			Detected:         detected,
			InstallDir:       installDir,
			DocumentsDir:     docsDir,
			NeedsManualSetup: !detected && (installDir != "" || docsDir != ""),
		})

		if detected {
			logging.Infof("game-detection: %s detected at %s", gi.name, installDir)
		} else {
			logging.Debugf("game-detection: %s not detected (install=%q, docs=%q)", gi.name, installDir, docsDir)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Detected != result[j].Detected {
			return result[i].Detected
		}
		return result[i].ID < result[j].ID
	})

	logging.Infof("game-detection: found %d games (%d detected)", len(result), func() int {
		n := 0
		for _, g := range result {
			if g.Detected {
				n++
			}
		}
		return n
	}())

	return result, nil
}

func (s *Detector) SetGamePaths(settingsPath, gameID, installDir, documentsDir string) error {
	gi := s.findGameInfo(domain.GameID(gameID))
	if gi == nil {
		logging.Warnf("game-detection: set paths for unknown game: %s", gameID)
		return nil
	}

	overrides, err := s.loadOverrides(settingsPath)
	if err != nil {
		return err
	}

	overrides[gi.id] = pathOverride{
		InstallDir:   installDir,
		DocumentsDir: documentsDir,
	}

	logging.Infof("game-detection: saved manual paths for %s (install=%q, docs=%q)", gi.name, installDir, documentsDir)

	return s.saveOverrides(settingsPath, overrides)
}

type pathOverride struct {
	InstallDir   string
	DocumentsDir string
}

func (s *Detector) loadOverrides(path string) (map[domain.GameID]pathOverride, error) {
	settings, err := s.settingsRepo.Load(path)
	if err != nil {
		return nil, err
	}

	overrides := make(map[domain.GameID]pathOverride)
	if settings.GamePaths == nil {
		return overrides, nil
	}

	for gid, paths := range settings.GamePaths {
		overrides[domain.GameID(gid)] = pathOverride{
			InstallDir:   paths.InstallDir,
			DocumentsDir: paths.DocumentsDir,
		}
	}

	return overrides, nil
}

func (s *Detector) saveOverrides(path string, overrides map[domain.GameID]pathOverride) error {
	settings, _ := s.settingsRepo.Load(path)

	if settings.GamePaths == nil {
		settings.GamePaths = make(map[string]repo.GamePathOverride)
	}

	for gid, ov := range overrides {
		settings.GamePaths[string(gid)] = repo.GamePathOverride{
			InstallDir:   ov.InstallDir,
			DocumentsDir: ov.DocumentsDir,
		}
	}

	return s.settingsRepo.Save(path, settings)
}

func (s *Detector) findGameInfo(id domain.GameID) *gameInfo {
	for i := range s.supportedGames {
		if s.supportedGames[i].id == id {
			return &s.supportedGames[i]
		}
	}
	return nil
}

func (s *Detector) detectInstallDir(dirs []string) string {
	for _, dir := range dirs {
		if dirExists(dir) {
			return dir
		}
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
