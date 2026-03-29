package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"

	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/mods"
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
	loStore         *loadorder.Store
	loState         loadorder.State
	conGraph        *graph.Graph
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
	return &App{
		loState:         loadorder.State{OrderedIDs: []string{}},
		conGraph:        graph.New(),
		modPathByID:     map[string]string{},
		launcherLayout:  LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
		playsetNames:    []string{},
		gameActiveIndex: -1,
		launcherIndex:   -1,
	}
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

	settings, err := loadSettings(a.settingsPath)
	if err != nil {
		logging.Warnf("startup: load settings, using defaults: %v", err)
	}
	a.settings = settings

	if a.gamePaths.PlaysetsPath != "" {
		playsetNames, gameActiveIndex, err := loadorder.ListPlaysets(a.gamePaths.PlaysetsPath)
		if err != nil {
			logging.Warnf("startup: read playset list: %v", err)
		} else {
			a.playsetNames = playsetNames
			a.gameActiveIndex = gameActiveIndex
			a.launcherIndex = resolveLauncherPlaysetIndex(len(playsetNames), gameActiveIndex, a.settings.LauncherActivePlaysetIndex)

			playsetState, pathByID, loadErr := loadorder.LoadStateFromPlaysets(a.gamePaths.PlaysetsPath, a.launcherIndex)
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

	loadedGraph, err := graph.LoadConstraints(a.constraintsPath)
	if err != nil {
		logging.Warnf("startup: load constraints, using empty graph: %v", err)
		a.conGraph = graph.New()
	} else {
		a.conGraph = loadedGraph
	}

	storedLayout, err := loadLauncherLayout(a.layoutPath)
	if err != nil {
		logging.Warnf("startup: load launcher layout, using defaults: %v", err)
		storedLayout = defaultLauncherLayout(a.loState.OrderedIDs)
	}

	a.launcherLayout = normalizeLauncherLayout(storedLayout, a.loState.OrderedIDs)
	if err := saveLauncherLayout(a.layoutPath, a.launcherLayout); err != nil {
		logging.Warnf("startup: persist normalized launcher layout: %v", err)
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

	allMods, err := mods.ScanDirs(scanRoots)
	if err != nil {
		logging.Errorf("mods scan failed for roots %q: %v", scanRoots, err)
		return nil, fmt.Errorf("scan mods roots %q: %w", scanRoots, err)
	}

	enabled := make(map[string]struct{}, len(a.loState.OrderedIDs))
	for _, id := range a.loState.OrderedIDs {
		enabled[id] = struct{}{}
	}

	for i := range allMods {
		a.modPathByID[allMods[i].ID] = allMods[i].DirPath
		_, ok := enabled[allMods[i].ID]
		allMods[i].Enabled = ok
	}

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

	next := uniqueIDs(ids)
	newState := loadorder.State{OrderedIDs: next}
	if err := a.loStore.Save(newState); err != nil {
		return fmt.Errorf("save fallback load order: %w", err)
	}

	if a.gamePaths.PlaysetsPath != "" {
		if err := loadorder.SaveStateToPlaysets(a.gamePaths.PlaysetsPath, a.launcherIndex, newState, a.modPathByID); err != nil {
			return fmt.Errorf("save load order to playsets %q: %w", a.gamePaths.PlaysetsPath, err)
		}
	}

	a.loState = newState
	a.launcherLayout = normalizeLauncherLayout(a.launcherLayout, a.loState.OrderedIDs)
	if err := saveLauncherLayout(a.layoutPath, a.launcherLayout); err != nil {
		logging.Warnf("set load order: failed to save launcher layout: %v", err)
	}
	return nil
}

// SetModEnabled enables or disables a single mod ID.
func (a *App) SetModEnabled(id string, enabled bool) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set mod enabled for %q: %w", id, err)
	}

	next := append([]string(nil), a.loState.OrderedIDs...)
	index := -1
	for i, current := range next {
		if current == id {
			index = i
			break
		}
	}

	if enabled {
		if index < 0 {
			next = append(next, id)
		}
	} else if index >= 0 {
		next = append(next[:index], next[index+1:]...)
	}

	return a.SetLoadOrder(next)
}

// GetConstraints returns all active constraints.
func (a *App) GetConstraints() []graph.Constraint {
	if err := a.ensureReady(); err != nil {
		logging.Errorf("GetConstraints called before initialization: %v", err)
		return []graph.Constraint{}
	}
	return a.conGraph.All()
}

// AddConstraint adds and persists a loads-after relationship.
func (a *App) AddConstraint(from, to string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, to, err)
	}
	fromCategory := isCategoryID(from)
	toCategory := isCategoryID(to)
	if fromCategory != toCategory {
		return fmt.Errorf("add constraint %q -> %q: categories can only constrain categories, and mods can only constrain mods", from, to)
	}

	if fromCategory {
		if err := a.addAfterConstraintSingle(from, to); err != nil {
			return fmt.Errorf("add category constraint %q -> %q: %w", from, to, err)
		}
		if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
			return fmt.Errorf("save constraints after add %q -> %q: %w", from, to, err)
		}
		return nil
	}

	fromIDs := a.expandConstraintTarget(from)
	toIDs := a.expandConstraintTarget(to)
	if len(fromIDs) == 0 || len(toIDs) == 0 {
		return fmt.Errorf("add constraint %q -> %q: no mods resolved from target", from, to)
	}

	for _, fromID := range fromIDs {
		for _, toID := range toIDs {
			if fromID == toID {
				continue
			}
			if err := a.addAfterConstraintSingle(fromID, toID); err != nil {
				return fmt.Errorf("add constraint %q -> %q expanded as %q -> %q: %w", from, to, fromID, toID, err)
			}
		}
	}

	if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
		return fmt.Errorf("save constraints after add %q -> %q: %w", from, to, err)
	}

	return nil
}

// AddLoadFirst marks a mod as load-first.
func (a *App) AddLoadFirst(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add load-first %q: %w", modID, err)
	}
	if modID == "" {
		return fmt.Errorf("add load-first: mod id must not be empty")
	}
	if isCategoryID(modID) {
		if a.conGraph.HasLast(modID) {
			return fmt.Errorf("add load-first %q: conflict: target is already marked load last", modID)
		}
		if a.conGraph.HasOutgoingAfter(modID) {
			return fmt.Errorf("add load-first %q: conflict: target has 'loads after' dependencies", modID)
		}
		a.conGraph.AddFirst(modID)
		if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
			return fmt.Errorf("save constraints after add load-first %q: %w", modID, err)
		}
		return nil
	}
	targets := a.expandConstraintTarget(modID)
	if len(targets) == 0 {
		return fmt.Errorf("add load-first %q: no mods resolved from target", modID)
	}
	for _, target := range targets {
		if a.conGraph.HasLast(target) {
			return fmt.Errorf("add load-first %q: conflict: mod %q is already marked load last", modID, target)
		}
		if a.conGraph.HasOutgoingAfter(target) {
			return fmt.Errorf("add load-first %q: conflict: mod %q has 'loads after' dependencies", modID, target)
		}
		a.conGraph.AddFirst(target)
	}

	if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
		return fmt.Errorf("save constraints after add load-first %q: %w", modID, err)
	}

	return nil
}

// AddLoadLast marks a mod as load-last.
func (a *App) AddLoadLast(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add load-last %q: %w", modID, err)
	}
	if modID == "" {
		return fmt.Errorf("add load-last: mod id must not be empty")
	}
	if isCategoryID(modID) {
		if a.conGraph.HasFirst(modID) {
			return fmt.Errorf("add load-last %q: conflict: target is already marked load first", modID)
		}
		if a.conGraph.HasIncomingAfter(modID) {
			return fmt.Errorf("add load-last %q: conflict: target has incoming constraints", modID)
		}
		a.conGraph.AddLast(modID)
		if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
			return fmt.Errorf("save constraints after add load-last %q: %w", modID, err)
		}
		return nil
	}
	targets := a.expandConstraintTarget(modID)
	if len(targets) == 0 {
		return fmt.Errorf("add load-last %q: no mods resolved from target", modID)
	}
	for _, target := range targets {
		if a.conGraph.HasFirst(target) {
			return fmt.Errorf("add load-last %q: conflict: mod %q is already marked load first", modID, target)
		}
		if a.conGraph.HasIncomingAfter(target) {
			return fmt.Errorf("add load-last %q: conflict: mod %q has incoming constraints", modID, target)
		}
		a.conGraph.AddLast(target)
	}

	if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
		return fmt.Errorf("save constraints after add load-last %q: %w", modID, err)
	}

	return nil
}

// RemoveConstraint removes and persists a loads-after relationship.
func (a *App) RemoveConstraint(from, to string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove constraint %q -> %q: %w", from, to, err)
	}
	fromCategory := isCategoryID(from)
	toCategory := isCategoryID(to)
	if fromCategory != toCategory {
		return fmt.Errorf("remove constraint %q -> %q: categories can only constrain categories, and mods can only constrain mods", from, to)
	}
	if fromCategory {
		a.conGraph.Remove(from, to)
		if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
			return fmt.Errorf("save constraints after remove %q -> %q: %w", from, to, err)
		}
		return nil
	}
	fromIDs := a.expandConstraintTarget(from)
	toIDs := a.expandConstraintTarget(to)
	for _, fromID := range fromIDs {
		for _, toID := range toIDs {
			a.conGraph.Remove(fromID, toID)
		}
	}
	if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
		return fmt.Errorf("save constraints after remove %q -> %q: %w", from, to, err)
	}

	return nil
}

// RemoveLoadFirst removes the load-first marker from a mod.
func (a *App) RemoveLoadFirst(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove load-first %q: %w", modID, err)
	}
	if isCategoryID(modID) {
		a.conGraph.RemoveFirst(modID)
		if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
			return fmt.Errorf("save constraints after remove load-first %q: %w", modID, err)
		}
		return nil
	}

	for _, target := range a.expandConstraintTarget(modID) {
		a.conGraph.RemoveFirst(target)
	}
	if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
		return fmt.Errorf("save constraints after remove load-first %q: %w", modID, err)
	}

	return nil
}

// RemoveLoadLast removes the load-last marker from a mod.
func (a *App) RemoveLoadLast(modID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove load-last %q: %w", modID, err)
	}
	if isCategoryID(modID) {
		a.conGraph.RemoveLast(modID)
		if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
			return fmt.Errorf("save constraints after remove load-last %q: %w", modID, err)
		}
		return nil
	}

	for _, target := range a.expandConstraintTarget(modID) {
		a.conGraph.RemoveLast(target)
	}
	if err := graph.SaveConstraints(a.constraintsPath, a.conGraph); err != nil {
		return fmt.Errorf("save constraints after remove load-last %q: %w", modID, err)
	}

	return nil
}

// Autosort reorders enabled mods by constraints, persists, and returns new order.
func (a *App) Autosort() ([]string, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("autosort: %w", err)
	}

	sorted, err := a.conGraph.Sort(a.loState.OrderedIDs)
	if err != nil {
		return nil, fmt.Errorf("sort constraints: %w", err)
	}

	if err := a.SetLoadOrder(sorted); err != nil {
		return nil, fmt.Errorf("persist autosorted load order: %w", err)
	}

	nextLayout, err := a.reorderLauncherLayoutAfterAutosort(sorted)
	if err != nil {
		return nil, fmt.Errorf("sort category constraints: %w", err)
	}
	a.launcherLayout = nextLayout
	if err := saveLauncherLayout(a.layoutPath, a.launcherLayout); err != nil {
		return nil, fmt.Errorf("save launcher layout after autosort: %w", err)
	}

	return append([]string(nil), a.loState.OrderedIDs...), nil
}

// GetLauncherLayout returns the launcher-only categorized ordering model.
func (a *App) GetLauncherLayout() LauncherLayout {
	a.launcherLayout = normalizeLauncherLayout(a.launcherLayout, a.loState.OrderedIDs)
	return a.launcherLayout
}

// SetLauncherLayout replaces launcher-only categorized ordering model.
func (a *App) SetLauncherLayout(layout LauncherLayout) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set launcher layout: %w", err)
	}

	a.launcherLayout = normalizeLauncherLayout(layout, a.loState.OrderedIDs)
	if err := saveLauncherLayout(a.layoutPath, a.launcherLayout); err != nil {
		return fmt.Errorf("save launcher layout: %w", err)
	}

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
	a.launcherLayout = normalizeLauncherLayout(a.launcherLayout, a.loState.OrderedIDs)

	if err := saveLauncherLayout(a.layoutPath, a.launcherLayout); err != nil {
		return LauncherCategory{}, fmt.Errorf("save launcher layout after category create: %w", err)
	}

	return created, nil
}

// DeleteLauncherCategory removes a category and returns its mods to ungrouped section.
func (a *App) DeleteLauncherCategory(categoryID string) error {
	if err := a.ensureReady(); err != nil {
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

	a.launcherLayout = normalizeLauncherLayout(next, a.loState.OrderedIDs)
	if err := saveLauncherLayout(a.layoutPath, a.launcherLayout); err != nil {
		return fmt.Errorf("save launcher layout after category delete %q: %w", categoryID, err)
	}

	return nil
}

// SaveCompiledLoadOrder compiles launcher layout into game order and persists to playsets.
func (a *App) SaveCompiledLoadOrder() ([]string, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("save compiled load order: %w", err)
	}

	a.launcherLayout = normalizeLauncherLayout(a.launcherLayout, a.loState.OrderedIDs)
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

	exePath := strings.TrimSpace(a.effectiveGameExe())
	if exePath == "" {
		return fmt.Errorf("launch game: executable path is not configured")
	}

	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("launch game: resolve executable path %q: %w", exePath, err)
	}

	info, err := os.Stat(absExe)
	if err != nil {
		return fmt.Errorf("launch game: stat executable %q: %w", absExe, err)
	}
	if info.IsDir() {
		return fmt.Errorf("launch game: executable path %q is a directory", absExe)
	}

	if shouldLaunchViaSteam(absExe) {
		steamCmd, err := buildSteamLaunchCommand(a.settings.GameArgs)
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

	cmd := buildLaunchCommand(absExe, a.settings.GameArgs)
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

	clean := strings.TrimSpace(path)
	if clean != "" {
		abs, err := filepath.Abs(clean)
		if err != nil {
			return fmt.Errorf("resolve game executable %q: %w", clean, err)
		}
		if !strings.EqualFold(filepath.Ext(abs), ".exe") {
			return fmt.Errorf("game executable %q must be an .exe file", abs)
		}
		if _, err := os.Stat(abs); err != nil {
			return fmt.Errorf("game executable %q not accessible: %w", abs, err)
		}
		clean = abs
	}

	a.settings.GameExe = clean
	if err := saveSettings(a.settingsPath, a.settings); err != nil {
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
	if err := openDirectoryInOS(dir); err != nil {
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
	if index < 0 || index >= len(a.playsetNames) {
		return fmt.Errorf("playset index %d is out of range", index)
	}

	playsetState, pathByID, err := loadorder.LoadStateFromPlaysets(a.gamePaths.PlaysetsPath, index)
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
	if err := saveSettings(a.settingsPath, a.settings); err != nil {
		return fmt.Errorf("persist launcher active playset %d: %w", index, err)
	}

	return nil
}

// SetModsDir persists custom mods directory override.
func (a *App) SetModsDir(path string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set mods dir: %w", err)
	}

	clean := strings.TrimSpace(path)
	if clean == "" {
		a.settings.ModsDir = ""
	} else {
		abs, err := filepath.Abs(clean)
		if err != nil {
			return fmt.Errorf("resolve mods dir %q: %w", clean, err)
		}
		if !dirExists(abs) {
			return fmt.Errorf("mods dir %q does not exist", abs)
		}
		a.settings.ModsDir = abs
	}

	if err := saveSettings(a.settingsPath, a.settings); err != nil {
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
	return nil
}

func (a *App) effectiveModsDir() string {
	if strings.TrimSpace(a.settings.ModsDir) != "" {
		return a.settings.ModsDir
	}
	return a.gamePaths.LocalModsDir
}

func (a *App) effectiveGameExe() string {
	if strings.TrimSpace(a.settings.GameExe) != "" {
		return a.settings.GameExe
	}
	return a.gamePaths.GameExePath
}

func (a *App) addAfterConstraintSingle(from, to string) error {
	if from == to {
		return fmt.Errorf("source and target must differ")
	}
	if a.conGraph.HasFirst(from) {
		return fmt.Errorf("conflict: %q is marked load first", from)
	}
	if a.conGraph.HasLast(to) {
		return fmt.Errorf("conflict: %q is marked load last", to)
	}
	a.conGraph.Add(from, to)
	return nil
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

func openDirectoryInOS(path string) error {
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func buildLaunchCommand(exePath string, args []string) *exec.Cmd {
	cmd := exec.Command(exePath, args...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	applyDetachedProcessAttributes(cmd)
	return cmd
}

func shouldLaunchViaSteam(exePath string) bool {
	if goruntime.GOOS != "windows" {
		return false
	}
	normalized := strings.ToLower(filepath.ToSlash(exePath))
	return strings.Contains(normalized, "/steamapps/common/europa universalis v/")
}

func buildSteamLaunchCommand(_ []string) (*exec.Cmd, error) {
	switch goruntime.GOOS {
	case "windows":
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", "steam://rungameid/"+eu5SteamAppID)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case "darwin":
		cmd := exec.Command("open", "steam://rungameid/"+eu5SteamAppID)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	default:
		cmd := exec.Command("xdg-open", "steam://rungameid/"+eu5SteamAppID)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	}
}

func uniqueIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func resolveLauncherPlaysetIndex(total int, gameActiveIndex int, preferred *int) int {
	if total <= 0 {
		return -1
	}
	if preferred != nil && *preferred >= 0 && *preferred < total {
		return *preferred
	}
	if gameActiveIndex >= 0 && gameActiveIndex < total {
		return gameActiveIndex
	}
	return 0
}

// Greet keeps the template method available for quick binding checks.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
