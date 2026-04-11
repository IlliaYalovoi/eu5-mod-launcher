package main

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetGameExe() string { return a.effectiveGameExe() }

func (a *App) GetAutoDetectedGameExe() string { return a.gamePaths.GameExePath }

func (a *App) SetGameExe(path string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set game exe: %w", err)
	}
	clean := strings.TrimSpace(path)
	a.settings.GameExe = clean
	if err := a.svc.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("save settings with game exe: %w", err)
	}
	return nil
}

func (a *App) ResetGameExeToAuto() (string, error) {
	if err := a.SetGameExe(""); err != nil {
		return "", err
	}
	return a.gamePaths.GameExePath, nil
}

func (a *App) GetModsDir() string { return a.effectiveModsDir() }

func (a *App) GetAutoDetectedModsDir() string { return a.gamePaths.LocalModsDir }

func (a *App) GetModsDirStatus() ModsDirStatus {
	autoDir := a.gamePaths.LocalModsDir
	effectiveDir := a.effectiveModsDir()
	return ModsDirStatus{
		EffectiveDir:       effectiveDir,
		AutoDetectedDir:    autoDir,
		CustomDir:          a.settings.ModsDir,
		UsingCustomDir:     strings.TrimSpace(a.settings.ModsDir) != "",
		AutoDetectedExists: dirExists(autoDir),
		EffectiveExists:    dirExists(effectiveDir),
	}
}

func (a *App) SetModsDir(path string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set mods dir: %w", err)
	}
	clean, err := a.svc.settingsSvc.NormalizeModsDir(path)
	if err != nil {
		return fmt.Errorf("set mods dir: %w", err)
	}
	a.settings.ModsDir = clean
	if err := a.svc.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("save settings with mods dir: %w", err)
	}
	return nil
}

func (a *App) ResetModsDirToAuto() (string, error) {
	if err := a.SetModsDir(""); err != nil {
		return "", err
	}
	return a.gamePaths.LocalModsDir, nil
}

func (a *App) GetConfigPath() string {
	if a.svc.loadOrderRepo != nil {
		return a.svc.loadOrderRepo.Path()
	}
	return ""
}

func (a *App) OpenConfigFolder() error {
	cfgPath := a.GetConfigPath()
	if cfgPath == "" {
		return fmt.Errorf("open config folder: %w", errAppStorageNotInitialized)
	}
	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("open config folder: %w", err)
	}
	return a.openURL(goruntime.GOOS, dir)
}

func (a *App) PickFolder() (string, error) {
	dir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Select Mods Directory",
	})
	if err != nil {
		return "", fmt.Errorf("pick folder: %w", err)
	}
	return dir, nil
}

func (a *App) PickExecutable() (string, error) {
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Select Game Executable",
		Filters: []wruntime.FileFilter{{
			DisplayName: "Executable (*.exe)",
			Pattern:     "*.exe",
		}},
	})
	if err != nil {
		return "", fmt.Errorf("pick executable: %w", err)
	}
	return path, nil
}

func (a *App) GetPlaysetNames() []string { return a.playsetNames }

func (a *App) GetGameActivePlaysetIndex() int { return a.gameActiveIdx }

func (a *App) GetLauncherActivePlaysetIndex() int { return a.launcherIdx }

func (a *App) SetLauncherActivePlaysetIndex(idx int) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set launcher active playset index %d: %w", idx, err)
	}
	if _, err := domain.ParsePlaysetIndex(idx); err != nil {
		return fmt.Errorf("set launcher active playset index %d: %w", idx, err)
	}
	if err := a.svc.playsetSvc.ValidateIndex(idx, len(a.playsetNames)); err != nil {
		return fmt.Errorf("set launcher active playset index %d: %w", idx, err)
	}
	state, pathByID, err := a.svc.gameSvc.ImportModList(a.activeGameID, a.gamePaths.PlaysetsPath, idx)
	if err != nil {
		return fmt.Errorf("load playset at index %d: %w", idx, err)
	}
	a.launcherIdx = idx
	a.loadOrder = state
	for id, path := range pathByID {
		a.modPathByID[id] = path
	}
	if err := a.svc.loadOrderRepo.Save(state); err != nil {
		return fmt.Errorf("save fallback loadorder for selected playset %d: %w", idx, err)
	}
	selectedIdx := idx
	a.settings.LauncherActivePlaysetIndex = &selectedIdx
	if err := a.svc.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("persist launcher active playset %d: %w", idx, err)
	}
	return nil
}

func (a *App) ListSupportedGames() ([]service.DetectedGame, error) {
	return a.svc.gameDetection.ListSupportedGames(a.settingsPath)
}

func (a *App) SetGamePaths(gameID, installDir, documentsDir string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set game paths: %w", err)
	}
	settings, err := a.svc.settingsRepo.Load(a.settingsPath)
	if err != nil {
		return fmt.Errorf("set game paths: %w", err)
	}
	if settings.GamePaths == nil {
		settings.GamePaths = make(map[string]repo.GamePathOverride)
	}
	override := repo.GamePathOverride{
		InstallDir:   installDir,
		DocumentsDir: documentsDir,
	}
	settings.GamePaths[gameID] = override
	if err := a.svc.settingsRepo.Save(a.settingsPath, settings); err != nil {
		return fmt.Errorf("set game paths: %w", err)
	}
	if gameID == string(a.activeGameID) {
		return a.RefreshGamePaths()
	}
	return nil
}

func (a *App) SetActiveGame(id string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set active game: %w", err)
	}
	parsedID := domain.GameID(id)
	adapter, err := a.svc.gameSvc.ResolveAdapter(parsedID)
	if err != nil {
		return fmt.Errorf("set active game %q: %w", id, err)
	}
	a.activeGameID = parsedID
	a.game = adapter
	var gamePaths domain.GamePaths
	gamePaths, err = a.svc.gameSvc.DiscoverPaths(a.activeGameID)
	if err != nil {
		return err
	}
	a.gamePaths = gamePaths
	a.modPathByID = make(map[string]string)
	return nil
}

func (a *App) RefreshGamePaths() error {
	var err error
	a.gamePaths, err = a.svc.gameSvc.DiscoverPaths(a.activeGameID)
	if err != nil {
		return err
	}
	a.modPathByID = make(map[string]string)
	return nil
}

func (a *App) LaunchGame() error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("launch game: %w", err)
	}
	absExe, err := a.svc.launchSvc.ValidateExecutable(strings.TrimSpace(a.effectiveGameExe()))
	if err != nil {
		return fmt.Errorf("launch game: %w", err)
	}
	if a.svc.settingsSvc.ShouldLaunchViaSteam(goruntime.GOOS, absExe) {
		if launched := a.tryLaunchViaSteam(); launched {
			return nil
		}
	}
	cmd := a.svc.launchSvc.BuildLaunchCommand(absExe, a.settings.GameArgs)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch game: start detached process %q: %w", absExe, err)
	}
	return nil
}

func (a *App) tryLaunchViaSteam() bool {
	cmd, err := a.svc.launchSvc.BuildSteamLaunchCommand(goruntime.GOOS, eu5SteamAppID)
	if err != nil {
		return false
	}
	if err := cmd.Start(); err != nil {
		return false
	}
	return true
}

func dirExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
