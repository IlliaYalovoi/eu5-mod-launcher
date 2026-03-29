package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"

	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/mods"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	constraintsFileName = "constraints.json"
	settingsFileName    = "settings.json"
	launcherLayoutFile  = "launcher_layout.json"
	eu5SteamAppID       = "3450310"
)

// App wires Wails-exposed methods to internal business packages.
type App struct {
	ctx             context.Context
	gamePaths       loadorder.GamePaths
	settings        appSettings
	playsetNames    []string
	gameActiveIndex int
	launcherIndex   int
	modPathByID     map[string]string
	launcherLayout  LauncherLayout
	modsService     *service.ModsService
	loadorderSvc    *service.LoadOrderService
	settingsSvc     *service.SettingsService
	layoutSvc       *service.LayoutService[LauncherLayout]
	launchSvc       *service.LaunchService
	playsetSvc      *service.PlaysetService
	constraintsRepo repo.ConstraintsRepository
	playsetRepo     repo.PlaysetRepository
	settingsRepo    repo.SettingsRepository
	layoutRepo      repo.LayoutRepository
	loStore         *loadorder.Store
	loState         loadorder.State
	conGraph        *graph.Graph
	conService      *service.ConstraintsService
	constraintsPath string
	settingsPath    string
	layoutPath      string
}

type ModsDirStatus struct {
	EffectiveDir       string `json:"effectiveDir"`
	AutoDetectedDir    string `json:"autoDetectedDir"`
	CustomDir          string `json:"customDir"`
	UsingCustomDir     bool   `json:"usingCustomDir"`
	AutoDetectedExists bool   `json:"autoDetectedExists"`
	EffectiveExists    bool   `json:"effectiveExists"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{
		loState:         loadorder.State{OrderedIDs: []string{}},
		conGraph:        graph.New(),
		modPathByID:     map[string]string{},
		launcherLayout:  LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
		playsetNames:    []string{},
		gameActiveIndex: -1,
		launcherIndex:   -1,
	}
	app.initCoreServices()
	return app
}

func (a *App) initCoreServices() {
	a.constraintsRepo = repo.NewFileConstraintsRepository()
	a.playsetRepo = repo.NewFilePlaysetRepository()
	a.settingsRepo = repo.NewFileSettingsRepository()
	a.layoutRepo = repo.NewFileLayoutRepository()

	a.modsService = service.NewModsService()
	a.loadorderSvc = service.NewLoadOrderService()
	a.settingsSvc = service.NewSettingsService()
	a.launchSvc = service.NewLaunchService()
	a.playsetSvc = service.NewPlaysetService(a.playsetRepo)
	a.layoutSvc = service.NewLayoutService(normalizeLauncherLayout, func(layout LauncherLayout) error {
		if strings.TrimSpace(a.layoutPath) == "" {
			return nil
		}
		return a.layoutRepo.Save(a.layoutPath, toRepoLayout(layout))
	})
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	loadorderPath, err := loadorder.DefaultConfigPath()
	if err != nil {
		logging.Errorf("startup: resolve default loadorder path: %v", err)
		return
	}

	store, err := loadorder.New(loadorderPath)
	if err != nil {
		logging.Errorf("startup: initialize loadorder store: %v", err)
		return
	}
	a.loStore = store

	state, err := a.loStore.Load()
	if err != nil {
		logging.Warnf("startup: load fallback loadorder state, using empty: %v", err)
		a.loState = loadorder.State{OrderedIDs: []string{}}
	} else {
		a.loState = state
	}

	a.gamePaths, err = loadorder.DiscoverGamePaths()
	if err != nil {
		logging.Errorf("startup: auto-discover game paths: %v", err)
	}

	configDir := filepath.Dir(a.loStore.ConfigPath())
	a.constraintsPath = filepath.Join(configDir, constraintsFileName)
	a.settingsPath = filepath.Join(configDir, settingsFileName)
	a.layoutPath = filepath.Join(configDir, launcherLayoutFile)

	repoSettings, err := a.settingsRepo.Load(a.settingsPath)
	if err != nil {
		logging.Warnf("startup: load settings, using defaults: %v", err)
	}
	a.settings = fromRepoSettings(repoSettings)

	if a.gamePaths.PlaysetsPath != "" {
		playsetNames, gameActiveIndex, err := a.playsetSvc.List(a.gamePaths.PlaysetsPath)
		if err != nil {
			logging.Warnf("startup: read playset list: %v", err)
		} else {
			a.playsetNames = playsetNames
			a.gameActiveIndex = gameActiveIndex
			a.launcherIndex = a.playsetSvc.ResolveLauncherIndex(len(playsetNames), gameActiveIndex, a.settings.LauncherActivePlaysetIndex)

			playsetState, pathByID, loadErr := a.playsetSvc.Load(a.gamePaths.PlaysetsPath, a.launcherIndex)
			if loadErr != nil {
				logging.Warnf("startup: load selected playset state, using fallback state: %v", loadErr)
			} else {
				a.loState = playsetState
				for id, path := range pathByID {
					a.modPathByID[id] = path
				}
			}
		}
	}

	loadedGraph, err := a.constraintsRepo.Load(a.constraintsPath)
	if err != nil {
		logging.Warnf("startup: load constraints, using empty graph: %v", err)
		a.conGraph = graph.New()
	} else {
		a.conGraph = loadedGraph
	}
	a.initConstraintsService()

	repoLayout, err := a.layoutRepo.Load(a.layoutPath)
	if err != nil {
		logging.Warnf("startup: load launcher layout, using defaults: %v", err)
		repoLayout = toRepoLayout(defaultLauncherLayout(a.loState.OrderedIDs))
	}

	nextLayout, layoutErr := a.layoutSvc.Persist(fromRepoLayout(repoLayout), a.loState.OrderedIDs)
	a.launcherLayout = nextLayout
	if layoutErr != nil {
		logging.Warnf("startup: persist normalized launcher layout: %v", layoutErr)
	}

	logging.Infof(
		"app startup completed (playsets=%q, localMods=%q, workshopRoots=%d, gameExeAuto=%q, gameExeEffective=%q, gameActive=%d, launcherActive=%d)",
		a.gamePaths.PlaysetsPath,
		a.effectiveModsDir(),
		len(a.gamePaths.WorkshopModDirs),
		a.gamePaths.GameExePath,
		a.effectiveGameExe(),
		a.gameActiveIndex,
		a.launcherIndex,
	)
}

// GetAllMods returns all discovered mods and marks Enabled from load order state.
func (a *App) GetAllMods() ([]mods.Mod, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("get all mods: %w", err)
	}

	scanRoots := make([]string, 0, 1+len(a.gamePaths.WorkshopModDirs))
	scanRoots = append(scanRoots, a.effectiveModsDir())
	scanRoots = append(scanRoots, a.gamePaths.WorkshopModDirs...)

	allMods, nextPaths, err := a.modsService.Discover(scanRoots, a.loState.OrderedIDs, a.modPathByID)
	if err != nil {
		logging.Errorf("mods scan failed for roots %q: %v", scanRoots, err)
		return nil, fmt.Errorf("get all mods: %w", err)
	}
	a.modPathByID = nextPaths

	return allMods, nil
}

// GetLoadOrder returns ordered enabled mod IDs.
func (a *App) GetLoadOrder() []string {
	if err := a.ensureReady(); err != nil {
		logging.Errorf("GetLoadOrder called before initialization: %v", err)
		return []string{}
	}

	return append([]string(nil), a.loState.OrderedIDs...)
}

// SetLoadOrder replaces and persists the current enabled ordered mod IDs.
func (a *App) SetLoadOrder(ids []string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set load order: %w", err)
	}

	next, err := a.loadorderSvc.ValidateAndNormalize(ids)
	if err != nil {
		return fmt.Errorf("set load order: %w", err)
	}
	newState := loadorder.State{OrderedIDs: next}
	if err := a.loStore.Save(newState); err != nil {
		return fmt.Errorf("save fallback load order: %w", err)
	}

	if a.gamePaths.PlaysetsPath != "" {
		if err := a.playsetSvc.Save(a.gamePaths.PlaysetsPath, a.launcherIndex, newState, a.modPathByID); err != nil {
			return fmt.Errorf("save load order to playsets %q: %w", a.gamePaths.PlaysetsPath, err)
		}
	}

	a.loState = newState
	nextLayout, err := a.layoutSvc.Persist(a.launcherLayout, a.loState.OrderedIDs)
	if err != nil {
		logging.Warnf("set load order: failed to save launcher layout: %v", err)
	} else {
		a.launcherLayout = nextLayout
	}
	return nil
}

// SetModEnabled enables or disables a single mod ID.
func (a *App) SetModEnabled(id string, enabled bool) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set mod enabled for %q: %w", id, err)
	}

	next, err := a.loadorderSvc.ToggleEnabled(a.loState.OrderedIDs, id, enabled)
	if err != nil {
		return fmt.Errorf("set mod enabled for %q: %w", id, err)
	}

	return a.SetLoadOrder(next)
}

// GetConstraints returns all active constraints.
func (a *App) GetConstraints() []graph.Constraint {
	if err := a.ensureReady(); err != nil {
		logging.Errorf("GetConstraints called before initialization: %v", err)
		return []graph.Constraint{}
	}
	return a.conService.All()
}

// AddConstraint adds and persists a loads-after relationship.
func (a *App) AddConstraint(from, to string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, to, err)
	}
	if err := a.conService.AddConstraint(from, to); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, to, err)
	}
	return nil
}

// AddLoadFirst marks a mod as load-first.
func (a *App) AddLoadFirst(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add load-first %q: %w", modID, err)
	}
	if err := a.conService.AddLoadFirst(modID); err != nil {
		return fmt.Errorf("add load-first %q: %w", modID, err)
	}
	return nil
}

// AddLoadLast marks a mod as load-last.
func (a *App) AddLoadLast(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add load-last %q: %w", modID, err)
	}
	if err := a.conService.AddLoadLast(modID); err != nil {
		return fmt.Errorf("add load-last %q: %w", modID, err)
	}
	return nil
}

// RemoveConstraint removes and persists a loads-after relationship.
func (a *App) RemoveConstraint(from, to string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove constraint %q -> %q: %w", from, to, err)
	}
	if err := a.conService.RemoveConstraint(from, to); err != nil {
		return fmt.Errorf("remove constraint %q -> %q: %w", from, to, err)
	}
	return nil
}

// RemoveLoadFirst removes the load-first marker from a mod.
func (a *App) RemoveLoadFirst(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove load-first %q: %w", modID, err)
	}
	if err := a.conService.RemoveLoadFirst(modID); err != nil {
		return fmt.Errorf("remove load-first %q: %w", modID, err)
	}
	return nil
}

// RemoveLoadLast removes the load-last marker from a mod.
func (a *App) RemoveLoadLast(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove load-last %q: %w", modID, err)
	}
	if err := a.conService.RemoveLoadLast(modID); err != nil {
		return fmt.Errorf("remove load-last %q: %w", modID, err)
	}

	return nil
}

// Autosort reorders enabled mods by constraints, persists, and returns new order.
func (a *App) Autosort() ([]string, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("autosort: %w", err)
	}
	previousOrder := append([]string(nil), a.loState.OrderedIDs...)
	previousLayout := a.launcherLayout

	sorted, err := a.conGraph.Sort(a.loState.OrderedIDs)
	if err != nil {
		return nil, fmt.Errorf("sort constraints: %w", err)
	}

	if err := a.SetLoadOrder(sorted); err != nil {
		return nil, fmt.Errorf("persist autosorted load order: %w", err)
	}

	nextLayout, err := a.reorderLauncherLayoutAfterAutosort(sorted)
	if err != nil {
		if rollbackErr := a.SetLoadOrder(previousOrder); rollbackErr != nil {
			logging.Errorf("autosort rollback failed after category-sort error: %v", rollbackErr)
		}
		a.launcherLayout = previousLayout
		return nil, fmt.Errorf("sort category constraints: %w", err)
	}
	a.launcherLayout = nextLayout
	if err := a.layoutRepo.Save(a.layoutPath, toRepoLayout(a.launcherLayout)); err != nil {
		if rollbackErr := a.SetLoadOrder(previousOrder); rollbackErr != nil {
			logging.Errorf("autosort rollback failed after layout save error: %v", rollbackErr)
		}
		a.launcherLayout = previousLayout
		return nil, fmt.Errorf("save launcher layout after autosort: %w", err)
	}

	return append([]string(nil), a.loState.OrderedIDs...), nil
}

// GetLauncherLayout returns the launcher-only categorized ordering model.
func (a *App) GetLauncherLayout() LauncherLayout {
	a.launcherLayout = a.layoutSvc.Normalize(a.launcherLayout, a.loState.OrderedIDs)
	return a.launcherLayout
}

// SetLauncherLayout replaces launcher-only categorized ordering model.
func (a *App) SetLauncherLayout(layout LauncherLayout) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set launcher layout: %w", err)
	}

	next, err := a.layoutSvc.Persist(layout, a.loState.OrderedIDs)
	if err != nil {
		return fmt.Errorf("save launcher layout: %w", err)
	}
	a.launcherLayout = next

	return nil
}

// CreateLauncherCategory creates an empty category container in launcher layout.
func (a *App) CreateLauncherCategory(name string) (LauncherCategory, error) {
	if err := a.ensureReady(); err != nil {
		return LauncherCategory{}, fmt.Errorf("create launcher category: %w", err)
	}

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return LauncherCategory{}, fmt.Errorf("create launcher category: name must not be empty")
	}

	created := LauncherCategory{ID: generateCategoryID(trimmed), Name: trimmed, ModIDs: []string{}}
	a.launcherLayout.Categories = append(a.launcherLayout.Categories, created)
	next, err := a.layoutSvc.Persist(a.launcherLayout, a.loState.OrderedIDs)
	if err != nil {
		return LauncherCategory{}, fmt.Errorf("save launcher layout after category create: %w", err)
	}
	a.launcherLayout = next

	return created, nil
}

// DeleteLauncherCategory removes a category and returns its mods to ungrouped section.
func (a *App) DeleteLauncherCategory(categoryID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("delete launcher category %q: %w", categoryID, err)
	}
	if _, err := domain.ParseCategoryID(categoryID); err != nil {
		return fmt.Errorf("delete launcher category %q: %w", categoryID, err)
	}

	next := LauncherLayout{Ungrouped: append([]string(nil), a.launcherLayout.Ungrouped...), Categories: []LauncherCategory{}}

	for _, cat := range a.launcherLayout.Categories {
		if cat.ID == categoryID {
			next.Ungrouped = append(next.Ungrouped, cat.ModIDs...)
			continue
		}
		next.Categories = append(next.Categories, cat)
	}

	normalized, err := a.layoutSvc.Persist(next, a.loState.OrderedIDs)
	if err != nil {
		return fmt.Errorf("save launcher layout after category delete %q: %w", categoryID, err)
	}
	a.launcherLayout = normalized

	return nil
}

// SaveCompiledLoadOrder compiles launcher layout into game order and persists to playsets.
func (a *App) SaveCompiledLoadOrder() ([]string, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("save compiled load order: %w", err)
	}

	a.launcherLayout = a.layoutSvc.Normalize(a.launcherLayout, a.loState.OrderedIDs)
	compiled := compileLauncherLayout(a.launcherLayout)
	if err := a.SetLoadOrder(compiled); err != nil {
		return nil, fmt.Errorf("persist compiled load order: %w", err)
	}

	return append([]string(nil), a.loState.OrderedIDs...), nil
}

// GetModsDir returns the effective mods directory (custom override or autodetected fallback).
func (a *App) GetModsDir() string {
	return a.effectiveModsDir()
}

// GetAutoDetectedModsDir returns the autodetected local mods directory.
func (a *App) GetAutoDetectedModsDir() string {
	return a.gamePaths.LocalModsDir
}

// GetModsDirStatus returns mods directory source and availability details.
func (a *App) GetModsDirStatus() ModsDirStatus {
	autoDir := a.gamePaths.LocalModsDir
	effectiveDir := a.effectiveModsDir()
	return ModsDirStatus{
		EffectiveDir:       effectiveDir,
		AutoDetectedDir:    autoDir,
		CustomDir:          strings.TrimSpace(a.settings.ModsDir),
		UsingCustomDir:     strings.TrimSpace(a.settings.ModsDir) != "",
		AutoDetectedExists: dirExists(autoDir),
		EffectiveExists:    dirExists(effectiveDir),
	}
}

// GetGameExe returns effective game executable path (custom override or autodetected fallback).
func (a *App) GetGameExe() string {
	return a.effectiveGameExe()
}

// LaunchGame starts the configured game executable in a detached process.
func (a *App) LaunchGame() error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("launch game: %w", err)
	}

	absExe, err := a.launchSvc.ValidateExecutable(strings.TrimSpace(a.effectiveGameExe()))
	if err != nil {
		return fmt.Errorf("launch game: %w", err)
	}

	if a.settingsSvc.ShouldLaunchViaSteam(goruntime.GOOS, absExe) {
		steamCmd, err := a.launchSvc.BuildSteamLaunchCommand(goruntime.GOOS, eu5SteamAppID)
		if err != nil {
			logging.Warnf("launch game: steam launch unavailable, falling back to direct executable: %v", err)
		} else if err := steamCmd.Start(); err == nil {
			steamPID := 0
			if steamCmd.Process != nil {
				steamPID = steamCmd.Process.Pid
			}
			logging.Infof("launch game: started via steam appid=%s pid=%d", eu5SteamAppID, steamPID)
			return nil
		} else {
			logging.Warnf("launch game: steam launch failed, falling back to direct executable: %v", err)
		}
	}

	cmd := a.launchSvc.BuildLaunchCommand(absExe, a.settings.GameArgs)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch game: start detached process %q: %w", absExe, err)
	}

	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	logging.Infof("launch game: started detached process %q pid=%d", absExe, pid)

	return nil
}

// GetAutoDetectedGameExe returns autodetected EU5 executable path.
func (a *App) GetAutoDetectedGameExe() string {
	return a.gamePaths.GameExePath
}

// SetGameExe persists game executable path.
func (a *App) SetGameExe(path string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set game executable: %w", err)
	}

	clean, err := a.settingsSvc.NormalizeGameExe(path)
	if err != nil {
		return fmt.Errorf("set game executable: %w", err)
	}

	a.settings.GameExe = clean
	if err := a.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("save settings with game executable: %w", err)
	}

	return nil
}

// ResetGameExeToAuto clears custom executable override and returns autodetected fallback.
func (a *App) ResetGameExeToAuto() (string, error) {
	if err := a.SetGameExe(""); err != nil {
		return "", err
	}
	return a.gamePaths.GameExePath, nil
}

// GetConfigPath returns settings file path.
func (a *App) GetConfigPath() string {
	return a.settingsPath
}

// OpenConfigFolder asks OS to open settings directory.
func (a *App) OpenConfigFolder() error {
	dir := filepath.Dir(a.settingsPath)
	if err := a.launchSvc.OpenDirectory(goruntime.GOOS, dir); err != nil {
		return fmt.Errorf("open config folder %q: %w", dir, err)
	}
	return nil
}

// PickFolder opens a native folder picker.
func (a *App) PickFolder() (string, error) {
	if err := a.ensureReady(); err != nil {
		return "", fmt.Errorf("pick folder: %w", err)
	}

	path, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Mods Directory"})
	if err != nil {
		return "", fmt.Errorf("open directory dialog: %w", err)
	}

	return path, nil
}

// PickExecutable opens a native executable picker.
func (a *App) PickExecutable() (string, error) {
	if err := a.ensureReady(); err != nil {
		return "", fmt.Errorf("pick executable: %w", err)
	}

	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Select Game Executable",
		Filters: []wruntime.FileFilter{{
			DisplayName: "Executable (*.exe)",
			Pattern:     "*.exe",
		}},
	})
	if err != nil {
		return "", fmt.Errorf("open file dialog: %w", err)
	}

	return path, nil
}

// GetPlaysetNames returns all available playset display names.
func (a *App) GetPlaysetNames() []string {
	if err := a.ensureReady(); err != nil {
		logging.Errorf("GetPlaysetNames called before initialization: %v", err)
		return []string{}
	}
	return append([]string(nil), a.playsetNames...)
}

// GetGameActivePlaysetIndex returns the game-owned active playset index.
func (a *App) GetGameActivePlaysetIndex() int {
	return a.gameActiveIndex
}

// GetLauncherActivePlaysetIndex returns the launcher-owned editing playset index.
func (a *App) GetLauncherActivePlaysetIndex() int {
	return a.launcherIndex
}

// SetLauncherActivePlaysetIndex switches the launcher editing target.
func (a *App) SetLauncherActivePlaysetIndex(index int) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set launcher active playset index %d: %w", index, err)
	}
	if _, err := domain.ParsePlaysetIndex(index); err != nil {
		return fmt.Errorf("set launcher active playset index %d: %w", index, err)
	}
	if err := a.playsetSvc.ValidateIndex(index, len(a.playsetNames)); err != nil {
		return fmt.Errorf("set launcher active playset index %d: %w", index, err)
	}

	playsetState, pathByID, err := a.playsetSvc.Load(a.gamePaths.PlaysetsPath, index)
	if err != nil {
		return fmt.Errorf("load playset at index %d: %w", index, err)
	}

	a.launcherIndex = index
	a.loState = playsetState
	for id, path := range pathByID {
		a.modPathByID[id] = path
	}

	if err := a.loStore.Save(playsetState); err != nil {
		return fmt.Errorf("save fallback loadorder for selected playset %d: %w", index, err)
	}

	selectedIndex := index
	a.settings.LauncherActivePlaysetIndex = &selectedIndex
	if err := a.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("persist launcher active playset %d: %w", index, err)
	}

	return nil
}

// SetModsDir persists custom mods directory override.
func (a *App) SetModsDir(path string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set mods dir: %w", err)
	}

	clean, err := a.settingsSvc.NormalizeModsDir(path)
	if err != nil {
		return fmt.Errorf("set mods dir: %w", err)
	}
	a.settings.ModsDir = clean

	if err := a.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("save settings with mods dir: %w", err)
	}

	return nil
}

// ResetModsDirToAuto clears custom override and returns autodetected fallback.
func (a *App) ResetModsDirToAuto() (string, error) {
	if err := a.SetModsDir(""); err != nil {
		return "", err
	}
	return a.gamePaths.LocalModsDir, nil
}

func (a *App) ensureReady() error {
	if a.loStore == nil {
		return errors.New("app storage is not initialized")
	}
	if a.conGraph == nil {
		a.conGraph = graph.New()
	}
	if a.loState.OrderedIDs == nil {
		a.loState.OrderedIDs = []string{}
	}
	if a.modPathByID == nil {
		a.modPathByID = map[string]string{}
	}
	if a.playsetNames == nil {
		a.playsetNames = []string{}
	}
	if a.settingsPath == "" && a.loStore != nil {
		a.settingsPath = filepath.Join(filepath.Dir(a.loStore.ConfigPath()), settingsFileName)
	}
	if a.layoutPath == "" && a.loStore != nil {
		a.layoutPath = filepath.Join(filepath.Dir(a.loStore.ConfigPath()), launcherLayoutFile)
	}
	if a.modsService == nil || a.loadorderSvc == nil || a.settingsSvc == nil || a.layoutSvc == nil {
		a.initCoreServices()
	}
	if a.conService == nil {
		a.initConstraintsService()
	}
	return nil
}

func (a *App) initConstraintsService() {
	if a.conGraph == nil {
		a.conGraph = graph.New()
	}
	a.conService = service.NewConstraintsService(a.conGraph, a.constraintsPath, a.constraintsRepo, a.expandConstraintTarget, isCategoryID)
}

func (a *App) effectiveModsDir() string {
	return a.settingsSvc.EffectivePath(a.settings.ModsDir, a.gamePaths.LocalModsDir)
}

func (a *App) effectiveGameExe() string {
	return a.settingsSvc.EffectivePath(a.settings.GameExe, a.gamePaths.GameExePath)
}

func (a *App) expandConstraintTarget(target string) []string {
	if !isCategoryID(target) {
		if strings.TrimSpace(target) == "" {
			return nil
		}
		return []string{target}
	}

	ids := make(map[string]struct{})
	for _, category := range a.launcherLayout.Categories {
		if category.ID != target {
			continue
		}
		for _, modID := range category.ModIDs {
			if strings.TrimSpace(modID) == "" {
				continue
			}
			ids[modID] = struct{}{}
		}
	}

	return sortedKeys(ids)
}

func (a *App) reorderLauncherLayoutAfterAutosort(sorted []string) (LauncherLayout, error) {
	layout := normalizeLauncherLayout(a.launcherLayout, sorted)

	position := make(map[string]int, len(sorted))
	for i, id := range sorted {
		position[id] = i
	}

	sortModIDs := func(ids []string) []string {
		out := append([]string(nil), ids...)
		sort.Slice(out, func(i, j int) bool {
			pi, okI := position[out[i]]
			pj, okJ := position[out[j]]
			if !okI {
				pi = len(sorted) + 1_000_000
			}
			if !okJ {
				pj = len(sorted) + 1_000_000
			}
			if pi == pj {
				return out[i] < out[j]
			}
			return pi < pj
		})
		return out
	}

	layout.Ungrouped = sortModIDs(layout.Ungrouped)
	for i := range layout.Categories {
		layout.Categories[i].ModIDs = sortModIDs(layout.Categories[i].ModIDs)
	}

	categoryByID := map[string]LauncherCategory{}
	for _, cat := range layout.Categories {
		categoryByID[cat.ID] = cat
	}

	blockIDs := append([]string(nil), layout.Order...)
	if len(blockIDs) == 0 {
		blockIDs = append(blockIDs, defaultUngroupedCategoryID)
		for _, cat := range layout.Categories {
			blockIDs = append(blockIDs, cat.ID)
		}
	}

	present := make(map[string]struct{}, len(blockIDs))
	for _, id := range blockIDs {
		present[id] = struct{}{}
	}
	if _, ok := present[defaultUngroupedCategoryID]; !ok {
		blockIDs = append(blockIDs, defaultUngroupedCategoryID)
		present[defaultUngroupedCategoryID] = struct{}{}
	}
	for _, cat := range layout.Categories {
		if _, ok := present[cat.ID]; !ok {
			blockIDs = append(blockIDs, cat.ID)
			present[cat.ID] = struct{}{}
		}
	}

	orderIndex := map[string]int{}
	for i, id := range blockIDs {
		orderIndex[id] = i
	}

	adj := make(map[string][]string, len(blockIDs))
	indegree := make(map[string]int, len(blockIDs))
	for _, id := range blockIDs {
		adj[id] = []string{}
		indegree[id] = 0
	}

	firstSet := map[string]struct{}{}
	lastSet := map[string]struct{}{}
	for _, constraint := range a.conGraph.All() {
		switch constraint.Type {
		case graph.ConstraintTypeFirst:
			if isCategoryID(constraint.ModID) {
				if _, ok := indegree[constraint.ModID]; ok {
					firstSet[constraint.ModID] = struct{}{}
				}
			}
		case graph.ConstraintTypeLast:
			if isCategoryID(constraint.ModID) {
				if _, ok := indegree[constraint.ModID]; ok {
					lastSet[constraint.ModID] = struct{}{}
				}
			}
		default:
			if !isCategoryID(constraint.From) || !isCategoryID(constraint.To) {
				continue
			}
			_, fromOk := indegree[constraint.From]
			_, toOk := indegree[constraint.To]
			if !fromOk || !toOk {
				continue
			}
			adj[constraint.To] = append(adj[constraint.To], constraint.From)
			indegree[constraint.From]++
		}
	}

	priorityRank := func(id string) int {
		_, isFirst := firstSet[id]
		_, isLast := lastSet[id]
		if isFirst {
			return 0
		}
		if isLast {
			return 2
		}
		return 1
	}

	blockName := func(id string) string {
		if id == defaultUngroupedCategoryID {
			return "Ungrouped"
		}
		if cat, ok := categoryByID[id]; ok {
			return cat.Name
		}
		return id
	}

	queue := make([]string, 0, len(blockIDs))
	for _, id := range blockIDs {
		if indegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]string, 0, len(blockIDs))
	for len(queue) > 0 {
		sort.Slice(queue, func(i, j int) bool {
			ri := priorityRank(queue[i])
			rj := priorityRank(queue[j])
			if ri != rj {
				return ri < rj
			}
			if ri == 1 {
				oi := orderIndex[queue[i]]
				oj := orderIndex[queue[j]]
				if oi != oj {
					return oi < oj
				}
			}
			return strings.ToLower(blockName(queue[i])) < strings.ToLower(blockName(queue[j]))
		})

		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, next := range adj[current] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(result) != len(blockIDs) {
		remaining := make([]string, 0)
		for _, id := range blockIDs {
			if indegree[id] > 0 {
				remaining = append(remaining, id)
			}
		}
		return layout, fmt.Errorf("%w: category cycle %s", graph.ErrCycle, strings.Join(remaining, " -> "))
	}

	layout.Order = result
	return layout, nil
}

func toRepoSettings(settings appSettings) repo.AppSettingsData {
	return repo.AppSettingsData{
		ModsDir:                    settings.ModsDir,
		GameExe:                    settings.GameExe,
		GameArgs:                   append([]string(nil), settings.GameArgs...),
		LauncherActivePlaysetIndex: settings.LauncherActivePlaysetIndex,
	}
}

func fromRepoSettings(settings repo.AppSettingsData) appSettings {
	return appSettings{
		ModsDir:                    settings.ModsDir,
		GameExe:                    settings.GameExe,
		GameArgs:                   append([]string(nil), settings.GameArgs...),
		LauncherActivePlaysetIndex: settings.LauncherActivePlaysetIndex,
	}
}

func toRepoLayout(layout LauncherLayout) repo.LauncherLayoutData {
	categories := make([]repo.LauncherCategoryData, 0, len(layout.Categories))
	for _, category := range layout.Categories {
		categories = append(categories, repo.LauncherCategoryData{ID: category.ID, Name: category.Name, ModIDs: append([]string(nil), category.ModIDs...)})
	}
	collapsed := map[string]bool{}
	for id, value := range layout.Collapsed {
		collapsed[id] = value
	}
	return repo.LauncherLayoutData{
		Ungrouped:  append([]string(nil), layout.Ungrouped...),
		Categories: categories,
		Order:      append([]string(nil), layout.Order...),
		Collapsed:  collapsed,
	}
}

func fromRepoLayout(layout repo.LauncherLayoutData) LauncherLayout {
	categories := make([]LauncherCategory, 0, len(layout.Categories))
	for _, category := range layout.Categories {
		categories = append(categories, LauncherCategory{ID: category.ID, Name: category.Name, ModIDs: append([]string(nil), category.ModIDs...)})
	}
	collapsed := map[string]bool{}
	for id, value := range layout.Collapsed {
		collapsed[id] = value
	}
	return LauncherLayout{
		Ungrouped:  append([]string(nil), layout.Ungrouped...),
		Categories: categories,
		Order:      append([]string(nil), layout.Order...),
		Collapsed:  collapsed,
	}
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

// Greet keeps the template method available for quick binding checks.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
