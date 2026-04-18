package main

import (
	"context"
	"errors"
	"eu5-mod-launcher/internal/adapters/eu5"
	"eu5-mod-launcher/internal/adapters/legacy"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/mods"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
	"eu5-mod-launcher/internal/steam"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	constraintsFileName     = "constraints.json"
	settingsFileName        = "settings.json"
	launcherLayoutFile      = "launcher_layout.json"
	eu5SteamAppID           = "3450310"
	sortPriorityFirst       = 0
	sortPriorityMiddle      = 1
	sortPriorityLast        = 2
	maxSortWorkers          = 8
	minLayoutForWorkers     = 8
	steamMetadataTTL        = 24 * time.Hour
	steamMetadataMaxEntries = 5000
	steamImageMaxEntries    = 1000
	steamDescImageMaxEntry  = 3000
)

var (
	errLauncherCategoryNameEmpty = errors.New("launcher category name must not be empty")
	errAppStorageNotInitialized  = errors.New("app storage is not initialized")
	errSteamCacheRootEmpty       = errors.New("steam cache root is empty")
	errWorkshopItemIDInvalid     = errors.New("workshop item id is invalid")
	errWorkshopOpenInAppFallback = errors.New("in-app workshop fallback unavailable")
	errExternalLinkInvalid       = errors.New("external link is invalid")
	errUnsubscribeDisabled       = errors.New("unsubscribe feature is disabled")
)

// workshopMetadataFetcher is an interface for fetching workshop metadata.
type workshopMetadataFetcher interface {
	FetchWorkshopMetadata(ids []string) (map[string]steam.WorkshopItem, error)
}

type appSettings struct {
	GameArgs                   []string                         `json:"gameArgs,omitempty"`
	LauncherActivePlaysetIndex *int                             `json:"launcherActivePlaysetIndex,omitempty"`
	Games                      map[string]repo.GameSettingsData `json:"games,omitempty"`
}

// App wires Wails-exposed methods to internal business packages.
type App struct {
	ctx             context.Context
	gamePaths       loadorder.GamePaths
	settings        appSettings
	playsetNames    []string
	gameActiveIndex int
	launcherIndex   int
	gameService     *service.GameService
	modPathByID     map[string]string
	launcherLayout  LauncherLayout
	modsService     *service.ModsService
	loadorderSvc    *service.LoadOrderService
	settingsSvc     *service.SettingsService
	layoutSvc       *service.LayoutService[LauncherLayout]
	launchSvc       *service.LaunchService
	playsetSvc      *service.PlaysetService
	steamClient     workshopMetadataFetcher
	steamMetaCache  *steam.MetadataCache
	steamImageCache *steam.ImageCache
	steamDescCache  *steam.DescriptionImageCache
	imageDataURLMu  sync.RWMutex
	imageDataURLs   map[string]string
	openURL         func(goos, rawURL string) error
	openInAppURL    func(url string) error
	constraintsRepo repo.ConstraintsRepository
	playsetRepo     repo.PlaysetRepository
	settingsRepo    repo.SettingsRepository
	layoutRepo      repo.LayoutRepository
	thumbSync       *steam.ThumbnailSync
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

type startupLoads struct {
	settings       repo.AppSettingsData
	settingsErr    error
	constraints    *graph.Graph
	constraintsErr error
	layout         repo.LauncherLayoutData
	layoutErr      error
}

// NewApp creates a new App application struct.
func NewApp() *App {
	app := &App{
		loState:         loadorder.State{OrderedIDs: []string{}},
		conGraph:        graph.New(),
		modPathByID:     map[string]string{},
		launcherLayout:  LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
		imageDataURLs:   map[string]string{},
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
	a.gameService = service.NewGameService()
	a.gameService.Register(&eu5.Adapter{})
	a.gameService.Register(game.Adapter(legacy.NewSqliteAdapter("hoi4", "Hearts of Iron IV", "394360")))
	a.gameService.Register(game.Adapter(legacy.NewSqliteAdapter("ck3", "Crusader Kings III", "1158310")))
	a.gameService.Register(game.Adapter(legacy.NewSqliteAdapter("stellaris", "Stellaris", "281990")))
	a.gameService.Register(game.Adapter(legacy.NewSqliteAdapter("vic3", "Victoria 3", "529340")))

	a.setActiveGameOnStartup()

	a.modsService = service.NewModsService()
	a.loadorderSvc = service.NewLoadOrderService()
	a.settingsSvc = service.NewSettingsService()
	a.launchSvc = service.NewLaunchService()
	a.playsetSvc = service.NewPlaysetService(a.playsetRepo)
	a.steamClient = steam.NewClient()
	a.openURL = a.launchSvc.OpenURL
	a.openInAppURL = a.openURLInApp
	a.layoutSvc = service.NewLayoutService(normalizeLauncherLayout, func(layout LauncherLayout) error {
		if strings.TrimSpace(a.layoutPath) == "" {
			return nil
		}
		return a.layoutRepo.Save(a.layoutPath, toRepoLayout(layout))
	})
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.refreshState()

	if err := a.ensureSteamCaches(); err == nil {
		a.thumbSync.StartPeriodicCleanup(a.ctx, 15*time.Minute, 30*24*time.Hour)
		if mods, err := a.GetAllMods(); err == nil {
			modIDs := make([]string, 0, len(mods))
			for _, m := range mods {
				if itemID := a.workshopItemIDForMod(m.ID); itemID != "" {
					modIDs = append(modIDs, itemID)
				}
			}
			go a.thumbSync.SyncAll(a.ctx, modIDs)
		}
	}
}

func (a *App) refreshState() {
	loadorderPath, err := loadorder.DefaultConfigPath()
	if err != nil {
		logging.Errorf("refreshState: resolve default loadorder path: %v", err)
		return
	}

	gameID := a.GetActiveGameID()
	if gameID != "" {
		dir := filepath.Dir(loadorderPath)
		loadorderPath = filepath.Join(dir, gameID+"_loadorder.json")
	}

	store, err := loadorder.New(loadorderPath)
	if err != nil {
		logging.Errorf("refreshState: initialize loadorder store: %v", err)
		return
	}
	a.loStore = store

	state, err := a.loStore.Load()
	if err != nil {
		logging.Warnf("refreshState: load fallback loadorder state, using empty: %v", err)
		a.loState = loadorder.State{OrderedIDs: []string{}}
	} else {
		a.loState = state
	}

	if a.gameService == nil {
		return
	}
	inst, err := a.gameService.GetActiveInstance()
	if err != nil {
		logging.Warnf("refreshState: no active game instance, skipping paths refresh")
	} else {
		a.gamePaths = loadorder.GamePaths{
			PlaysetsPath:    filepath.Join(inst.UserConfigPath, "playsets.json"),
			LocalModsDir:    inst.LocalModsDir,
			WorkshopModDirs: inst.WorkshopModDirs,
			GameExePath:     inst.GameExePath,
		}
		// Special case for legacy SQLite games
		adapter := a.gameService.GetAdapter(inst.GameID)
		if _, ok := adapter.(repo.LegacyAdapter); ok && inst.GameID != "eu5" {
			a.gamePaths.PlaysetsPath = filepath.Join(inst.UserConfigPath, "launcher-v2.sqlite")
			// Already corrected in adapter to check .db too
		}
	}

	configDir := filepath.Dir(a.loStore.ConfigPath())
	gameID = a.GetActiveGameID()
	a.constraintsPath = filepath.Join(configDir, gameID+"_constraints.json")
	a.settingsPath = filepath.Join(configDir, settingsFileName)
	a.layoutPath = filepath.Join(configDir, gameID+"_launcher_layout.json")

	loads := a.loadStartupState()
	a.applyStartupSettings(loads.settings, loads.settingsErr)
	a.loadStartupPlaysetState()
	a.applyStartupConstraints(loads.constraints, loads.constraintsErr)
	a.applyStartupLayout(loads.layout, loads.layoutErr)

	logging.Infof(
		"app state refreshed (playsets=%q, localMods=%q, workshopRoots=%d, gameExeAuto=%q, "+
			"gameExeEffective=%q, gameActive=%d, launcherActive=%d)",
		a.gamePaths.PlaysetsPath,
		a.effectiveModsDir(),
		len(a.gamePaths.WorkshopModDirs),
		a.gamePaths.GameExePath,
		a.effectiveGameExe(),
		a.gameActiveIndex,
		a.launcherIndex,
	)
}

func (a *App) loadStartupState() startupLoads {
	out := startupLoads{}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		out.settings, out.settingsErr = a.settingsRepo.Load(a.settingsPath)
	}()
	go func() {
		defer wg.Done()
		out.constraints, out.constraintsErr = a.constraintsRepo.Load(a.constraintsPath)
	}()
	go func() {
		defer wg.Done()
		out.layout, out.layoutErr = a.layoutRepo.Load(a.layoutPath)
	}()
	wg.Wait()

	return out
}

func (a *App) applyStartupSettings(settings repo.AppSettingsData, settingsErr error) {
	if settingsErr != nil {
		logging.Warnf("startup: load settings, using defaults: %v", settingsErr)
	}
	a.settings = fromRepoSettings(settings)
}

func (a *App) loadStartupPlaysetState() {
	if a.gamePaths.PlaysetsPath == "" {
		a.playsetNames = []string{}
		a.gameActiveIndex = -1
		a.launcherIndex = -1
		a.loState = loadorder.State{OrderedIDs: []string{}}
		return
	}

	playsetNames, gameActiveIndex, err := a.playsetSvc.List(a.gamePaths.PlaysetsPath)
	if err != nil {
		logging.Warnf("startup: read playset list: %v", err)
		a.playsetNames = []string{}
		a.gameActiveIndex = -1
		a.launcherIndex = -1
		a.loState = loadorder.State{OrderedIDs: []string{}}
		return
	}

	a.playsetNames = playsetNames
	a.gameActiveIndex = gameActiveIndex
	a.launcherIndex = a.playsetSvc.ResolveLauncherIndex(
		len(playsetNames),
		gameActiveIndex,
		a.settings.LauncherActivePlaysetIndex,
	)

	playsetState, pathByID, loadErr := a.playsetSvc.Load(a.gamePaths.PlaysetsPath, a.launcherIndex)
	if loadErr != nil {
		logging.Warnf("startup: load selected playset state, using fallback state: %v", loadErr)
		return
	}

	a.loState = playsetState
	maps.Copy(a.modPathByID, pathByID)
}

func (a *App) applyStartupConstraints(loadedGraph *graph.Graph, constraintsErr error) {
	if constraintsErr != nil {
		logging.Warnf("startup: load constraints, using empty graph: %v", constraintsErr)
		a.conGraph = graph.New()
		a.initConstraintsService()
		return
	}

	if loadedGraph == nil {
		loadedGraph = graph.New()
	}
	a.conGraph = loadedGraph
	a.initConstraintsService()
}

func (a *App) applyStartupLayout(repoLayout repo.LauncherLayoutData, layoutLoadErr error) {
	if layoutLoadErr != nil {
		logging.Warnf("startup: load launcher layout, using defaults: %v", layoutLoadErr)
		repoLayout = toRepoLayout(defaultLauncherLayout(a.loState.OrderedIDs))
	}

	nextLayout := fromRepoLayout(repoLayout)
	layoutErr := a.layoutSvc.Persist(&nextLayout, a.loState.OrderedIDs)
	a.launcherLayout = nextLayout
	if layoutErr != nil {
		logging.Warnf("startup: persist normalized launcher layout: %v", layoutErr)
	}
}

// GetAllMods returns all discovered mods and marks Enabled from load order state.
func (a *App) GetAllMods() ([]mods.Mod, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("get all mods: %w", err)
	}

	gameVersion := "unknown"
	gameID := a.GetActiveGameID()
	if inst, err := a.gameService.GetActiveInstance(); err == nil {
		override := ""
		if config, ok := a.settings.Games[gameID]; ok {
			override = config.GameVersionOverride
		}
		adapter := a.gameService.GetAdapter(gameID)
		if v, err := adapter.DetectVersion(*inst, override); err == nil {
			gameVersion = v
		}
	}

	scanRoots := make([]string, 0, 1+len(a.gamePaths.WorkshopModDirs))
	scanRoots = append(scanRoots, a.effectiveModsDir())
	scanRoots = append(scanRoots, a.gamePaths.WorkshopModDirs...)

	allMods, nextPaths, err := a.modsService.Discover(scanRoots, a.loState.OrderedIDs, a.modPathByID, gameVersion)
	if err != nil {
		logging.Errorf("mods scan failed for roots %q: %v", scanRoots, err)
		return nil, fmt.Errorf("get all mods: %w", err)
	}
	a.modPathByID = nextPaths

	for i := range allMods {
		itemID := a.workshopItemIDForMod(allMods[i].ID)
		if itemID == "" || a.steamImageCache == nil {
			continue
		}
		if cachedPath := a.steamImageCache.CachedPath(itemID); cachedPath != "" {
			if src := a.resolveImageSource(cachedPath); src != "" {
				allMods[i].ThumbnailPath = src
			} else {
				allMods[i].ThumbnailPath = cachedPath
			}
		}
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

	next, err := a.loadorderSvc.ValidateAndNormalize(ids)
	if err != nil {
		return fmt.Errorf("set load order: %w", err)
	}
	newState := loadorder.State{OrderedIDs: next}
	if saveErr := a.loStore.Save(newState); saveErr != nil {
		return fmt.Errorf("save fallback load order: %w", saveErr)
	}

	if a.gamePaths.PlaysetsPath != "" {
		saveErr := a.playsetSvc.Save(
			a.gamePaths.PlaysetsPath,
			a.launcherIndex,
			newState,
			a.modPathByID,
		)
		if saveErr != nil {
			return fmt.Errorf("save load order to playsets %q: %w", a.gamePaths.PlaysetsPath, saveErr)
		}
	}

	a.loState = newState
	nextLayout := a.launcherLayout
	err = a.layoutSvc.Persist(&nextLayout, a.loState.OrderedIDs)
	if err != nil {
		logging.Warnf("set load order: failed to save launcher layout: %v", err)
	} else {
		a.launcherLayout = nextLayout
	}
	return nil
}

// EnableMod enables a single mod ID.
func (a *App) EnableMod(id string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("enable mod %q: %w", id, err)
	}

	next, err := a.loadorderSvc.Enable(a.loState.OrderedIDs, id)
	if err != nil {
		return fmt.Errorf("enable mod %q: %w", id, err)
	}

	return a.SetLoadOrder(next)
}

// DisableMod disables a single mod ID.
func (a *App) DisableMod(id string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("disable mod %q: %w", id, err)
	}

	next, err := a.loadorderSvc.Disable(a.loState.OrderedIDs, id)
	if err != nil {
		return fmt.Errorf("disable mod %q: %w", id, err)
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
func (a *App) AddConstraint(from, target string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, target, err)
	}
	if err := a.conService.AddConstraint(from, target); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, target, err)
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
func (a *App) RemoveConstraint(from, target string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("remove constraint %q -> %q: %w", from, target, err)
	}
	if err := a.conService.RemoveConstraint(from, target); err != nil {
		return fmt.Errorf("remove constraint %q -> %q: %w", from, target, err)
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

	if saveErr := a.SetLoadOrder(sorted); saveErr != nil {
		return nil, fmt.Errorf("persist autosorted load order: %w", saveErr)
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
	next := a.launcherLayout
	a.layoutSvc.Normalize(&next, a.loState.OrderedIDs)
	a.launcherLayout = next
	return a.launcherLayout
}

// SetLauncherLayout replaces launcher-only categorized ordering model.
func (a *App) SetLauncherLayout(layout LauncherLayout) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set launcher layout: %w", err)
	}

	next := layout
	if err := a.layoutSvc.Persist(&next, a.loState.OrderedIDs); err != nil {
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
		return LauncherCategory{}, fmt.Errorf("create launcher category: %w", errLauncherCategoryNameEmpty)
	}

	created := LauncherCategory{ID: generateCategoryID(trimmed), Name: trimmed, ModIDs: []string{}}
	a.launcherLayout.Categories = append(a.launcherLayout.Categories, created)
	next := a.launcherLayout
	if err := a.layoutSvc.Persist(&next, a.loState.OrderedIDs); err != nil {
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

	next := LauncherLayout{
		Ungrouped:  append([]string(nil), a.launcherLayout.Ungrouped...),
		Categories: []LauncherCategory{},
	}

	for i := range a.launcherLayout.Categories {
		cat := a.launcherLayout.Categories[i]
		if cat.ID == categoryID {
			next.Ungrouped = append(next.Ungrouped, cat.ModIDs...)
			continue
		}
		next.Categories = append(next.Categories, cat)
	}

	normalized := next
	if err := a.layoutSvc.Persist(&normalized, a.loState.OrderedIDs); err != nil {
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

	next := a.launcherLayout
	a.layoutSvc.Normalize(&next, a.loState.OrderedIDs)
	a.launcherLayout = next
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
	gameID := a.GetActiveGameID()
	autoDir := a.gamePaths.LocalModsDir
	effectiveDir := a.effectiveModsDir()
	customDir := ""
	if config, ok := a.settings.Games[gameID]; ok {
		customDir = config.ModsDir
	}
	autoDetectedExists := dirExists(autoDir)
	if !autoDetectedExists {
		// Fallback for legacy games: check if the parent Documents folder exists
		if inst, err := a.gameService.GetActiveInstance(); err == nil && dirExists(inst.UserConfigPath) {
			autoDetectedExists = true
		}
	}

	return ModsDirStatus{
		EffectiveDir:       effectiveDir,
		AutoDetectedDir:    autoDir,
		CustomDir:          strings.TrimSpace(customDir),
		UsingCustomDir:     strings.TrimSpace(customDir) != "",
		AutoDetectedExists: autoDetectedExists,
		EffectiveExists:    dirExists(effectiveDir) || autoDetectedExists,
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
		if launched := a.tryLaunchViaSteam(); launched {
			return nil
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

func (a *App) tryLaunchViaSteam() bool {
	steamCmd, err := a.launchSvc.BuildSteamLaunchCommand(goruntime.GOOS, eu5SteamAppID)
	if err != nil {
		logging.Warnf("launch game: steam launch unavailable, falling back to direct executable: %v", err)
		return false
	}

	if err := steamCmd.Start(); err != nil {
		logging.Warnf("launch game: steam launch failed, falling back to direct executable: %v", err)
		return false
	}

	steamPID := 0
	if steamCmd.Process != nil {
		steamPID = steamCmd.Process.Pid
	}
	logging.Infof("launch game: started via steam appid=%s pid=%d", eu5SteamAppID, steamPID)

	return true
}

// GetAutoDetectedGameExe returns autodetected EU5 executable path.
func (a *App) GetAutoDetectedGameExe() string {
	return a.gamePaths.GameExePath
}

// GetGameVersion returns the detected game version (or override if set).
func (a *App) GetGameVersion() string {
	gameID := a.GetActiveGameID()
	if inst, err := a.gameService.GetActiveInstance(); err == nil {
		override := ""
		if config, ok := a.settings.Games[gameID]; ok {
			override = config.GameVersionOverride
		}
		adapter := a.gameService.GetAdapter(gameID)
		if v, err := adapter.DetectVersion(*inst, override); err == nil {
			return v
		}
	}
	return "unknown"
}

// GetGameVersionOverride returns the custom game version override.
func (a *App) GetGameVersionOverride() string {
	gameID := a.GetActiveGameID()
	if config, ok := a.settings.Games[gameID]; ok {
		return config.GameVersionOverride
	}
	return ""
}

// SetGameVersionOverride sets a custom game version string and persists it.
func (a *App) SetGameVersionOverride(version string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set game version override: %w", err)
	}

	gameID := a.GetActiveGameID()
	config := a.settings.Games[gameID]
	config.GameVersionOverride = strings.TrimSpace(version)
	a.settings.Games[gameID] = config

	if err := a.settingsRepo.Save(a.settingsPath, toRepoSettings(a.settings)); err != nil {
		return fmt.Errorf("save settings with game version override: %w", err)
	}

	return nil
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

	gameID := a.GetActiveGameID()
	config := a.settings.Games[gameID]
	config.GameExe = clean
	a.settings.Games[gameID] = config

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
	maps.Copy(a.modPathByID, pathByID)

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

	gameID := a.GetActiveGameID()
	config := a.settings.Games[gameID]
	config.ModsDir = clean
	a.settings.Games[gameID] = config

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

// SetActiveGame switches the current game and re-initializes paths.
func (a *App) SetActiveGame(gameID string) error {
	if err := a.gameService.SetActiveGame(gameID, 0); err != nil {
		return err
	}

	inst, err := a.gameService.GetActiveInstance()
	if err != nil {
		return err
	}

	a.gamePaths = loadorder.GamePaths{
		PlaysetsPath:    filepath.Join(inst.UserConfigPath, "playsets.json"),
		LocalModsDir:    inst.LocalModsDir,
		WorkshopModDirs: inst.WorkshopModDirs,
		GameExePath:     inst.GameExePath,
	}

	// Update Playset Service Repository
	adapter := a.gameService.GetAdapter(gameID)
	if _, ok := adapter.(repo.LegacyAdapter); ok && gameID != "eu5" {
		sqlite := adapter.(repo.LegacyAdapter)
		a.gamePaths.PlaysetsPath = filepath.Join(inst.UserConfigPath, "launcher-v2.db")
		a.playsetRepo = repo.NewSqlitePlaysetRepository(sqlite, *inst)
	} else {
		a.gamePaths.PlaysetsPath = filepath.Join(inst.UserConfigPath, "playsets.json")
		a.playsetRepo = repo.NewFilePlaysetRepository()
	}
	a.playsetSvc = service.NewPlaysetService(a.playsetRepo)

	if a.ctx != nil {
		a.refreshState()
	}
	return nil
}

func (a *App) setActiveGameOnStartup() {
	// Default to EU5 on startup
	if err := a.SetActiveGame("eu5"); err != nil {
		logging.Errorf("failed to set active game eu5 on startup: %v", err)
	}
}

// GetAvailableGames returns a list of game IDs supported by the launcher.
func (a *App) GetAvailableGames() []string {
	adapters := a.gameService.GetAdapters()
	ids := make([]string, 0, len(adapters))
	for _, adapter := range adapters {
		ids = append(ids, adapter.ID())
	}
	sort.Strings(ids)
	return ids
}

// GetActiveGameID returns the ID of the currently active game.
func (a *App) GetActiveGameID() string {
	inst, err := a.gameService.GetActiveInstance()
	if err != nil {
		return ""
	}
	return inst.GameID
}

func (a *App) ensureReady() error {
	if a.loStore == nil {
		return errAppStorageNotInitialized
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
	if a.imageDataURLs == nil {
		a.imageDataURLs = map[string]string{}
	}
	if a.settingsPath == "" {
		a.settingsPath = filepath.Join(filepath.Dir(a.loStore.ConfigPath()), settingsFileName)
	}
	if a.layoutPath == "" {
		a.layoutPath = filepath.Join(filepath.Dir(a.loStore.ConfigPath()), launcherLayoutFile)
	}
	if a.coreServicesMissing() {
		a.initCoreServices()
	}
	if err := a.ensureSteamCaches(); err != nil {
		return err
	}
	if a.conService == nil {
		a.initConstraintsService()
	}
	return nil
}

func (a *App) ensureSteamCaches() error {
	if a.steamMetaCache != nil && a.steamImageCache != nil && a.steamDescCache != nil {
		return nil
	}

	cacheRoot := filepath.Dir(a.settingsPath)
	if strings.TrimSpace(cacheRoot) == "" && a.loStore != nil {
		cacheRoot = filepath.Dir(a.loStore.ConfigPath())
	}
	if strings.TrimSpace(cacheRoot) == "" {
		return fmt.Errorf("initialize steam caches: %w", errSteamCacheRootEmpty)
	}

	metaCache, err := steam.NewMetadataCache(cacheRoot, steamMetadataTTL, steamMetadataMaxEntries)
	if err != nil {
		return fmt.Errorf("initialize metadata cache: %w", err)
	}
	imageCache, err := steam.NewImageCache(cacheRoot, steamImageMaxEntries, nil)
	if err != nil {
		return fmt.Errorf("initialize image cache: %w", err)
	}
	descCache, err := steam.NewDescriptionImageCache(cacheRoot, steamDescImageMaxEntry, nil)
	if err != nil {
		return fmt.Errorf("initialize description image cache: %w", err)
	}

	a.steamMetaCache = metaCache
	a.steamImageCache = imageCache
	a.steamDescCache = descCache
	a.thumbSync = steam.NewThumbnailSync(a.steamClient.(*steam.Client), a.steamMetaCache, a.steamImageCache, 10)
	return nil
}

func (a *App) coreServicesMissing() bool {
	return a.modsService == nil ||
		a.loadorderSvc == nil ||
		a.settingsSvc == nil ||
		a.layoutSvc == nil ||
		a.steamClient == nil
}

func (a *App) initConstraintsService() {
	if a.conGraph == nil {
		a.conGraph = graph.New()
	}
	a.conService = service.NewConstraintsService(
		a.conGraph,
		a.constraintsPath,
		a.constraintsRepo,
		a.expandConstraintTarget,
		isCategoryID,
	)
}

func (a *App) effectiveModsDir() string {
	gameID := a.GetActiveGameID()
	if config, ok := a.settings.Games[gameID]; ok && strings.TrimSpace(config.ModsDir) != "" {
		return config.ModsDir
	}
	return a.gamePaths.LocalModsDir
}

func (a *App) effectiveGameExe() string {
	gameID := a.GetActiveGameID()
	if config, ok := a.settings.Games[gameID]; ok && strings.TrimSpace(config.GameExe) != "" {
		return config.GameExe
	}
	return a.gamePaths.GameExePath
}

func (a *App) expandConstraintTarget(target string) []string {
	if !isCategoryID(target) {
		if strings.TrimSpace(target) == "" {
			return nil
		}
		return []string{target}
	}

	ids := make(map[string]struct{})
	for i := range a.launcherLayout.Categories {
		category := a.launcherLayout.Categories[i]
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

	position := buildIDPositionMap(sorted)
	sortLayoutModIDs(&layout, position, len(sorted))

	categoryByID := indexCategoriesByID(layout.Categories)
	blockIDs := completeCategoryBlockOrder(layout)
	sortGraph := buildCategorySortGraph(blockIDs, a.conGraph.All())

	order, err := sortCategoryBlocks(blockIDs, sortGraph, categoryByID)
	if err != nil {
		return layout, err
	}

	layout.Order = order
	return layout, nil
}

func buildIDPositionMap(sorted []string) map[string]int {
	position := make(map[string]int, len(sorted))
	for i, id := range sorted {
		position[id] = i
	}
	return position
}

func indexCategoriesByID(categories []LauncherCategory) map[string]LauncherCategory {
	out := make(map[string]LauncherCategory, len(categories))
	for i := range categories {
		cat := categories[i]
		out[cat.ID] = cat
	}
	return out
}

func completeCategoryBlockOrder(layout LauncherLayout) []string {
	blockIDs := append([]string(nil), layout.Order...)
	if len(blockIDs) == 0 {
		blockIDs = append(blockIDs, defaultUngroupedCategoryID)
		for i := range layout.Categories {
			blockIDs = append(blockIDs, layout.Categories[i].ID)
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
	for i := range layout.Categories {
		catID := layout.Categories[i].ID
		if _, ok := present[catID]; ok {
			continue
		}
		blockIDs = append(blockIDs, catID)
		present[catID] = struct{}{}
	}
	return blockIDs
}

type categorySortGraph struct {
	adj      map[string][]string
	indegree map[string]int
	order    map[string]int
	firstSet map[string]struct{}
	lastSet  map[string]struct{}
}

func buildCategorySortGraph(blockIDs []string, constraints []graph.Constraint) categorySortGraph {
	adj := make(map[string][]string, len(blockIDs))
	indegree := make(map[string]int, len(blockIDs))
	order := make(map[string]int, len(blockIDs))
	for i, id := range blockIDs {
		adj[id] = []string{}
		indegree[id] = 0
		order[id] = i
	}

	firstSet := map[string]struct{}{}
	lastSet := map[string]struct{}{}
	for i := range constraints {
		constraint := constraints[i]
		switch constraint.Type {
		case graph.ConstraintTypeFirst:
			if isCategoryConstraintNode(constraint.ModID, indegree) {
				firstSet[constraint.ModID] = struct{}{}
			}
		case graph.ConstraintTypeLast:
			if isCategoryConstraintNode(constraint.ModID, indegree) {
				lastSet[constraint.ModID] = struct{}{}
			}
		default:
			if !isValidCategoryEdge(constraint.From, constraint.To, indegree) {
				continue
			}
			adj[constraint.To] = append(adj[constraint.To], constraint.From)
			indegree[constraint.From]++
		}
	}

	return categorySortGraph{
		adj:      adj,
		indegree: indegree,
		order:    order,
		firstSet: firstSet,
		lastSet:  lastSet,
	}
}

func isCategoryConstraintNode(id string, indegree map[string]int) bool {
	if !isCategoryID(id) {
		return false
	}
	_, ok := indegree[id]
	return ok
}

func isValidCategoryEdge(from, to string, indegree map[string]int) bool {
	if !isCategoryID(from) || !isCategoryID(to) {
		return false
	}
	_, fromOk := indegree[from]
	_, toOk := indegree[to]
	return fromOk && toOk
}

func sortCategoryBlocks(
	blockIDs []string,
	sortGraph categorySortGraph,
	categoryByID map[string]LauncherCategory,
) ([]string, error) {
	queue := make([]string, 0, len(blockIDs))
	for _, id := range blockIDs {
		if sortGraph.indegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]string, 0, len(blockIDs))
	for len(queue) > 0 {
		sortQueueByPriority(queue, sortGraph, categoryByID)

		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, next := range sortGraph.adj[current] {
			sortGraph.indegree[next]--
			if sortGraph.indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(result) != len(blockIDs) {
		remaining := collectRemainingCategoryCycle(blockIDs, sortGraph.indegree)
		return nil, fmt.Errorf("%w: category cycle %s", graph.ErrCycle, strings.Join(remaining, " -> "))
	}

	return result, nil
}

func sortQueueByPriority(
	queue []string,
	sortGraph categorySortGraph,
	categoryByID map[string]LauncherCategory,
) {
	sort.Slice(queue, func(i, j int) bool {
		leftRank := categoryPriorityRank(queue[i], sortGraph.firstSet, sortGraph.lastSet)
		rightRank := categoryPriorityRank(queue[j], sortGraph.firstSet, sortGraph.lastSet)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if leftRank == sortPriorityMiddle {
			leftOrder := sortGraph.order[queue[i]]
			rightOrder := sortGraph.order[queue[j]]
			if leftOrder != rightOrder {
				return leftOrder < rightOrder
			}
		}

		leftName := strings.ToLower(categoryDisplayName(queue[i], categoryByID))
		rightName := strings.ToLower(categoryDisplayName(queue[j], categoryByID))
		return leftName < rightName
	})
}

func categoryPriorityRank(id string, firstSet, lastSet map[string]struct{}) int {
	if _, isFirst := firstSet[id]; isFirst {
		return sortPriorityFirst
	}
	if _, isLast := lastSet[id]; isLast {
		return sortPriorityLast
	}
	return sortPriorityMiddle
}

func categoryDisplayName(id string, categoryByID map[string]LauncherCategory) string {
	if id == defaultUngroupedCategoryID {
		return "Ungrouped"
	}
	if cat, ok := categoryByID[id]; ok {
		return cat.Name
	}
	return id
}

func collectRemainingCategoryCycle(blockIDs []string, indegree map[string]int) []string {
	remaining := make([]string, 0)
	for _, id := range blockIDs {
		if indegree[id] > 0 {
			remaining = append(remaining, id)
		}
	}
	return remaining
}

func (a *App) workshopItemIDForMod(modID string) string {
	modPath := strings.TrimSpace(a.modPathByID[modID])
	if modPath == "" {
		return ""
	}
	return workshopItemIDFromPath(modPath, a.gamePaths.WorkshopModDirs)
}

func workshopItemIDFromPath(modPath string, workshopRoots []string) string {
	cleanModPath := filepath.Clean(strings.TrimSpace(modPath))
	if cleanModPath == "" {
		return ""
	}

	for _, root := range workshopRoots {
		cleanRoot := filepath.Clean(strings.TrimSpace(root))
		if cleanRoot == "" {
			continue
		}

		rel, err := filepath.Rel(cleanRoot, cleanModPath)
		if err != nil {
			continue
		}
		rel = filepath.Clean(rel)
		if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}

		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) == 0 {
			continue
		}
		candidate := strings.TrimSpace(parts[0])
		if isWorkshopNumericID(candidate) {
			return candidate
		}
	}

	pathParts := strings.Split(filepath.ToSlash(cleanModPath), "/")
	for i := 0; i+3 < len(pathParts); i++ {
		isWorkshopContentPrefix := strings.EqualFold(pathParts[i], "workshop") &&
			strings.EqualFold(pathParts[i+1], "content") &&
			pathParts[i+2] == eu5SteamAppID
		if !isWorkshopContentPrefix {
			continue
		}
		candidate := strings.TrimSpace(pathParts[i+3])
		if isWorkshopNumericID(candidate) {
			return candidate
		}
	}

	return ""
}

func isWorkshopNumericID(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func sortIDsByPosition(ids []string, position map[string]int, fallback int) []string {
	out := append([]string(nil), ids...)
	sort.Slice(out, func(i, j int) bool {
		pi, okI := position[out[i]]
		pj, okJ := position[out[j]]
		if !okI {
			pi = fallback
		}
		if !okJ {
			pj = fallback
		}
		if pi == pj {
			return out[i] < out[j]
		}
		return pi < pj
	})
	return out
}

func sortLayoutModIDs(layout *LauncherLayout, position map[string]int, sortedCount int) {
	workers := max(goruntime.NumCPU(), 1)
	if workers > maxSortWorkers {
		workers = maxSortWorkers
	}
	if len(layout.Categories) < minLayoutForWorkers || workers == 1 {
		sortLayoutModIDsSequential(layout, position, sortedCount)
		return
	}
	sortLayoutModIDsConcurrent(layout, position, sortedCount, workers)
}

func sortLayoutModIDsSequential(layout *LauncherLayout, position map[string]int, sortedCount int) {
	fallback := sortedCount + 1_000_000
	layout.Ungrouped = sortIDsByPosition(layout.Ungrouped, position, fallback)
	for i := range layout.Categories {
		layout.Categories[i].ModIDs = sortIDsByPosition(layout.Categories[i].ModIDs, position, fallback)
	}
}

func sortLayoutModIDsConcurrent(layout *LauncherLayout, position map[string]int, sortedCount, workers int) {
	fallback := sortedCount + 1_000_000
	layout.Ungrouped = sortIDsByPosition(layout.Ungrouped, position, fallback)

	jobs := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				layout.Categories[idx].ModIDs = sortIDsByPosition(layout.Categories[idx].ModIDs, position, fallback)
			}
		}()
	}
	for i := range layout.Categories {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}

func toRepoSettings(settings appSettings) repo.AppSettingsData {
	return repo.AppSettingsData{
		GameArgs:                   append([]string(nil), settings.GameArgs...),
		LauncherActivePlaysetIndex: settings.LauncherActivePlaysetIndex,
		Games:                      maps.Clone(settings.Games),
	}
}

func fromRepoSettings(settings repo.AppSettingsData) appSettings {
	games := maps.Clone(settings.Games)
	if games == nil {
		games = make(map[string]repo.GameSettingsData)
	}
	return appSettings{
		GameArgs:                   append([]string(nil), settings.GameArgs...),
		LauncherActivePlaysetIndex: settings.LauncherActivePlaysetIndex,
		Games:                      games,
	}
}

func toRepoLayout(layout LauncherLayout) repo.LauncherLayoutData {
	categories := make([]repo.LauncherCategoryData, 0, len(layout.Categories))
	for i := range layout.Categories {
		category := layout.Categories[i]
		categories = append(categories, repo.LauncherCategoryData{
			ID:     category.ID,
			Name:   category.Name,
			ModIDs: append([]string(nil), category.ModIDs...),
		})
	}
	collapsed := map[string]bool{}
	maps.Copy(collapsed, layout.Collapsed)
	return repo.LauncherLayoutData{
		Ungrouped:  append([]string(nil), layout.Ungrouped...),
		Categories: categories,
		Order:      append([]string(nil), layout.Order...),
		Collapsed:  collapsed,
	}
}

func fromRepoLayout(layout repo.LauncherLayoutData) LauncherLayout {
	categories := make([]LauncherCategory, 0, len(layout.Categories))
	for i := range layout.Categories {
		category := layout.Categories[i]
		categories = append(categories, LauncherCategory{
			ID:     category.ID,
			Name:   category.Name,
			ModIDs: append([]string(nil), category.ModIDs...),
		})
	}
	collapsed := map[string]bool{}
	maps.Copy(collapsed, layout.Collapsed)
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
func (*App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// OpenWorkshopItem opens workshop page using unified external-link priority rules.
func (a *App) OpenWorkshopItem(itemID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("open workshop item %q: %w", itemID, err)
	}

	normalizedID, err := normalizeWorkshopItemID(itemID)
	if err != nil {
		return fmt.Errorf("open workshop item %q: %w", itemID, err)
	}

	httpsURL := "https://steamcommunity.com/sharedfiles/filedetails/?id=" + normalizedID
	if err := a.OpenExternalLink(httpsURL); err != nil {
		return fmt.Errorf("open workshop item %q: %w", normalizedID, err)
	}

	return nil
}

// UnsubscribeWorkshopMod opens the Steam unsubscribe flow for a workshop item.
func (a *App) UnsubscribeWorkshopMod(itemID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", itemID, err)
	}
	if !a.IsUnsubscribeEnabled() {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", itemID, errUnsubscribeDisabled)
	}

	trimmedID := strings.TrimSpace(itemID)
	if trimmedID == "" {
		return nil
	}

	unsubscribeURL, err := a.launchSvc.BuildWorkshopUnsubscribeURL(trimmedID)
	if err != nil {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", itemID, err)
	}

	if err := a.OpenExternalLink(unsubscribeURL); err != nil {
		return fmt.Errorf("unsubscribe workshop mod %q: %w", trimmedID, err)
	}

	return nil
}

// IsUnsubscribeEnabled reports whether workshop unsubscribe is enabled.
func (*App) IsUnsubscribeEnabled() bool {
	return compileEnableUnsubscribe
}

// HasNewThumbnails reports whether new thumbnails were recently synced.
func (a *App) HasNewThumbnails() bool {
	if a.thumbSync == nil {
		return false
	}
	return a.thumbSync.HasNewThumbnails()
}

// OpenExternalLink opens any external URL with priority rules:
// steam links: steam client -> default browser -> in-app window
// non-steam links: default browser -> in-app window
func (a *App) OpenExternalLink(rawURL string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("open external link %q: %w", rawURL, err)
	}

	normalizedURL, linkErr := normalizeExternalLink(rawURL)
	if linkErr != nil {
		return fmt.Errorf("open external link %q: %w", rawURL, linkErr)
	}

	parsedURL, parseErr := url.Parse(normalizedURL)
	if parseErr != nil {
		return fmt.Errorf("open external link %q: parse normalized url: %w", normalizedURL, parseErr)
	}

	attempts := make([]error, 0, 3)
	if isSteamLikeLink(parsedURL) {
		steamURL := toSteamClientURL(parsedURL)
		if err := a.openURL(goruntime.GOOS, steamURL); err == nil {
			return nil
		} else {
			attempts = append(attempts, fmt.Errorf("open in steam client: %w", err))
		}
	}

	if err := a.openURL(goruntime.GOOS, normalizedURL); err == nil {
		return nil
	} else {
		attempts = append(attempts, fmt.Errorf("open in default browser: %w", err))
	}

	if err := a.openInAppURL(normalizedURL); err == nil {
		return nil
	} else {
		attempts = append(attempts, fmt.Errorf("open in wails window fallback: %w", err))
	}

	return fmt.Errorf("open external link %q: %w", normalizedURL, errors.Join(attempts...))
}

func normalizeExternalLink(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", errExternalLinkInvalid
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("%w: parse %q: %s", errExternalLinkInvalid, rawURL, err.Error())
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" && scheme != "steam" {
		return "", fmt.Errorf("%w: unsupported scheme %q", errExternalLinkInvalid, scheme)
	}
	if scheme != "steam" && strings.TrimSpace(parsed.Host) == "" {
		return "", fmt.Errorf("%w: missing host", errExternalLinkInvalid)
	}

	return parsed.String(), nil
}

func isSteamLikeLink(parsedURL *url.URL) bool {
	if parsedURL == nil {
		return false
	}
	if strings.EqualFold(parsedURL.Scheme, "steam") {
		return true
	}
	host := strings.ToLower(parsedURL.Hostname())
	if host == "steamcommunity.com" || strings.HasSuffix(host, ".steamcommunity.com") {
		return true
	}
	if host == "store.steampowered.com" || strings.HasSuffix(host, ".steampowered.com") {
		return true
	}
	return false
}

func toSteamClientURL(parsedURL *url.URL) string {
	if parsedURL == nil {
		return ""
	}
	if strings.EqualFold(parsedURL.Scheme, "steam") {
		return parsedURL.String()
	}

	if itemID := workshopItemIDFromCommunityURL(parsedURL); itemID != "" {
		return "steam://url/CommunityFilePage/" + itemID
	}

	return "steam://openurl/" + parsedURL.String()
}

func workshopItemIDFromCommunityURL(parsedURL *url.URL) string {
	if parsedURL == nil {
		return ""
	}
	host := strings.ToLower(parsedURL.Hostname())
	if host != "steamcommunity.com" && !strings.HasSuffix(host, ".steamcommunity.com") {
		return ""
	}

	queryID := strings.TrimSpace(parsedURL.Query().Get("id"))
	if queryID == "" || !isWorkshopNumericID(queryID) {
		return ""
	}

	path := strings.ToLower(strings.TrimSpace(parsedURL.Path))
	if strings.Contains(path, "/sharedfiles/filedetails") || strings.Contains(path, "/workshop/filedetails") {
		return queryID
	}

	return ""
}

func normalizeWorkshopItemID(itemID string) (string, error) {
	normalizedID := strings.TrimSpace(itemID)
	if normalizedID == "" {
		return "", errWorkshopItemIDInvalid
	}
	parsed, err := strconv.ParseUint(normalizedID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %q", errWorkshopItemIDInvalid, itemID)
	}

	return strconv.FormatUint(parsed, 10), nil
}

func (a *App) openURLInApp(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errWorkshopOpenInAppFallback
	}
	if a.ctx == nil {
		return errWorkshopOpenInAppFallback
	}

	quoted := strconv.Quote(rawURL)
	js := "window.open(" + quoted + ", '_blank', 'noopener,noreferrer')"
	wruntime.WindowExecJS(a.ctx, js)

	return nil
}

// FetchWorkshopMetadataForMod returns Steam workshop metadata for a single mod.
func (a *App) FetchWorkshopMetadataForMod(modID string) (steam.WorkshopItem, error) {
	if err := a.ensureReady(); err != nil {
		return steam.WorkshopItem{}, fmt.Errorf("fetch workshop metadata for mod %q: %w", modID, err)
	}

	itemID := a.workshopItemIDForMod(modID)
	if itemID == "" {
		return steam.WorkshopItem{}, nil
	}

	lookup, cacheErr := a.steamMetaCache.Get(itemID)
	if cacheErr != nil {
		return steam.WorkshopItem{}, fmt.Errorf("fetch workshop metadata for mod %q: %w", modID, cacheErr)
	}
	if lookup.Hit {
		prepared := a.prepareWorkshopItem(lookup.Item)
		if lookup.Stale {
			go a.revalidateWorkshopMetadata([]string{itemID})
		}
		return prepared, nil
	}

	items, err := a.fetchAndCacheWorkshopMetadata([]string{itemID})
	if err != nil {
		return steam.WorkshopItem{}, fmt.Errorf("fetch workshop metadata for mod %q: %w", modID, err)
	}
	if item, ok := items[itemID]; ok {
		return a.prepareWorkshopItem(item), nil
	}

	return steam.WorkshopItem{ItemID: itemID}, nil
}

// RefreshWorkshopMetadataForMod forces metadata refresh from Steam for one mod.
func (a *App) RefreshWorkshopMetadataForMod(modID string) (steam.WorkshopItem, error) {
	if err := a.ensureReady(); err != nil {
		return steam.WorkshopItem{}, fmt.Errorf("refresh workshop metadata for mod %q: %w", modID, err)
	}

	itemID := a.workshopItemIDForMod(modID)
	if itemID == "" {
		return steam.WorkshopItem{}, nil
	}

	items, err := a.refreshAndCacheWorkshopMetadata([]string{itemID})
	if err != nil {
		return steam.WorkshopItem{}, fmt.Errorf("refresh workshop metadata for mod %q: %w", modID, err)
	}
	if item, ok := items[itemID]; ok {
		return a.prepareRefreshedWorkshopItem(item), nil
	}

	return steam.WorkshopItem{ItemID: itemID}, nil
}

// FetchWorkshopMetadataBatch returns Steam workshop metadata for a list of mod IDs.
// Result keys are mod IDs for direct UI mapping.
func (a *App) FetchWorkshopMetadataBatch(modIDs []string) (map[string]steam.WorkshopItem, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("fetch workshop metadata batch: %w", err)
	}

	workshopToModIDs, itemIDs := a.workshopIDsByMod(modIDs)
	if len(workshopToModIDs) == 0 {
		return map[string]steam.WorkshopItem{}, nil
	}

	resolved, cacheErr := a.steamMetaCache.ResolveMany(itemIDs)
	if cacheErr != nil {
		return nil, fmt.Errorf("fetch workshop metadata batch: %w", cacheErr)
	}

	byModID := make(map[string]steam.WorkshopItem)
	a.mapWorkshopItemsToMods(byModID, workshopToModIDs, resolved.Fresh)
	a.mapWorkshopItemsToMods(byModID, workshopToModIDs, resolved.Stale)

	if len(resolved.Stale) > 0 {
		go a.revalidateWorkshopMetadata(sortedWorkshopItemIDs(resolved.Stale))
	}
	if len(resolved.Missing) == 0 {
		return byModID, nil
	}

	fetched, fetchErr := a.fetchAndCacheWorkshopMetadata(resolved.Missing)
	if fetchErr != nil {
		if len(byModID) > 0 {
			logging.Warnf("workshop metadata batch fetch fallback to cache after fetch error: %v", fetchErr)
			return byModID, nil
		}
		return nil, fmt.Errorf("fetch workshop metadata batch: %w", fetchErr)
	}
	a.mapWorkshopItemsToMods(byModID, workshopToModIDs, fetched)
	return byModID, nil
}

// RefreshWorkshopMetadataBatch forces metadata refresh for all resolvable workshop-backed mods.
func (a *App) RefreshWorkshopMetadataBatch(modIDs []string) (map[string]steam.WorkshopItem, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("refresh workshop metadata batch: %w", err)
	}

	workshopToModIDs, itemIDs := a.workshopIDsByMod(modIDs)
	if len(workshopToModIDs) == 0 {
		return map[string]steam.WorkshopItem{}, nil
	}

	fetched, err := a.refreshAndCacheWorkshopMetadata(itemIDs)
	if err != nil {
		return nil, fmt.Errorf("refresh workshop metadata batch: %w", err)
	}

	byModID := make(map[string]steam.WorkshopItem)
	a.mapWorkshopItemsToMods(byModID, workshopToModIDs, fetched)
	return byModID, nil
}

func (a *App) workshopIDsByMod(modIDs []string) (map[string][]string, []string) {
	workshopToModIDs := map[string][]string{}
	for _, modID := range modIDs {
		itemID := a.workshopItemIDForMod(modID)
		if itemID == "" {
			continue
		}
		workshopToModIDs[itemID] = append(workshopToModIDs[itemID], modID)
	}
	itemIDs := make([]string, 0, len(workshopToModIDs))
	for itemID := range workshopToModIDs {
		itemIDs = append(itemIDs, itemID)
	}
	sort.Strings(itemIDs)
	return workshopToModIDs, itemIDs
}

func (a *App) mapWorkshopItemsToMods(
	byModID map[string]steam.WorkshopItem,
	workshopToModIDs map[string][]string,
	items map[string]steam.WorkshopItem,
) {
	for itemID := range items {
		item := a.prepareWorkshopItem(items[itemID])
		for _, modID := range workshopToModIDs[itemID] {
			byModID[modID] = item
		}
	}
}

func sortedWorkshopItemIDs(items map[string]steam.WorkshopItem) []string {
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (a *App) fetchAndCacheWorkshopMetadata(itemIDs []string) (map[string]steam.WorkshopItem, error) {
	items, err := a.steamClient.FetchWorkshopMetadata(itemIDs)
	if err != nil {
		return nil, err
	}
	if setErr := a.steamMetaCache.SetMany(items); setErr != nil {
		logging.Warnf("steam metadata cache write failed: %v", setErr)
	}
	return items, nil
}

func (a *App) refreshAndCacheWorkshopMetadata(itemIDs []string) (map[string]steam.WorkshopItem, error) {
	items, err := a.steamClient.FetchWorkshopMetadata(itemIDs)
	if err != nil {
		return nil, err
	}
	if setErr := a.steamMetaCache.SetMany(items); setErr != nil {
		logging.Warnf("steam metadata cache write failed: %v", setErr)
	}
	return items, nil
}

func (a *App) ensurePreviewCached(item steam.WorkshopItem) {
	if a.steamImageCache == nil {
		return
	}
	if _, err := a.steamImageCache.EnsureStored(item); err != nil {
		logging.Debugf("steam preview cache for %q skipped: %v", item.ItemID, err)
	}
}

func (a *App) refreshPreviewCached(item steam.WorkshopItem) {
	if a.steamImageCache == nil {
		return
	}
	if _, err := a.steamImageCache.RefreshStored(item); err != nil {
		logging.Debugf("steam preview cache refresh for %q skipped: %v", item.ItemID, err)
	}
}

func (a *App) prepareWorkshopItem(item steam.WorkshopItem) steam.WorkshopItem {
	a.ensurePreviewCached(item)
	return a.rewriteDescriptionImages(item)
}

func (a *App) prepareRefreshedWorkshopItem(item steam.WorkshopItem) steam.WorkshopItem {
	a.refreshPreviewCached(item)
	return a.rewriteDescriptionImages(item)
}

func (a *App) rewriteDescriptionImages(item steam.WorkshopItem) steam.WorkshopItem {
	if a.steamDescCache == nil || strings.TrimSpace(item.ItemID) == "" {
		return item
	}

	imageURLs := steam.ExtractDescriptionImageURLs(item.Description)
	if len(imageURLs) == 0 {
		return item
	}

	replacements := make(map[string]string, len(imageURLs))
	for _, imageURL := range imageURLs {
		cachedPath, err := a.steamDescCache.EnsureStored(item.ItemID, imageURL)
		if err != nil {
			logging.Debugf("steam description image cache for %q skipped: %v", item.ItemID, err)
			continue
		}
		if src := a.resolveImageSource(cachedPath); src != "" {
			replacements[imageURL] = src
			continue
		}
		replacements[imageURL] = cachedPath
	}

	item.Description = steam.ReplaceDescriptionImageURLs(item.Description, replacements)
	return item
}

func (a *App) resolveImageSource(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	lowered := strings.ToLower(trimmed)
	if strings.HasPrefix(lowered, "http://") || strings.HasPrefix(lowered, "https://") || strings.HasPrefix(lowered, "data:image/") {
		return trimmed
	}

	cleanPath := filepath.Clean(trimmed)
	a.imageDataURLMu.RLock()
	if cached, ok := a.imageDataURLs[cleanPath]; ok && strings.TrimSpace(cached) != "" {
		a.imageDataURLMu.RUnlock()
		return cached
	}
	a.imageDataURLMu.RUnlock()

	dataURL, err := steam.EncodeImageFileAsDataURL(cleanPath)
	if err != nil {
		logging.Debugf("resolve image source %q skipped: %v", cleanPath, err)
		return ""
	}

	a.imageDataURLMu.Lock()
	a.imageDataURLs[cleanPath] = dataURL
	a.imageDataURLMu.Unlock()
	return dataURL
}

func (a *App) revalidateWorkshopMetadata(itemIDs []string) {
	if _, err := a.fetchAndCacheWorkshopMetadata(itemIDs); err != nil {
		logging.Debugf("steam metadata background revalidate skipped: %v", err)
	}
}
