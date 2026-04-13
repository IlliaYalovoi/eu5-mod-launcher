package launcher

import (
	"context"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
	"eu5-mod-launcher/internal/steam"
	"path/filepath"
	"strings"
	"sync"
)

type appServices struct {
	modsService     *service.ModsService
	loadorderSvc    *service.LoadOrderService
	settingsSvc     *service.SettingsService
	layoutSvc       *service.LayoutService[LauncherLayout]
	launchSvc       *service.LaunchService
	gameSvc         *service.GameService
	playsetSvc      *service.PlaysetService
	constraintsRepo repo.ConstraintsRepository
	playsetRepo     repo.PlaysetRepository
	settingsRepo    repo.SettingsRepository
	layoutRepo      repo.LayoutRepository
	loadOrderRepo   repo.LoadOrderRepo
	conGraph        *domain.Graph
	conService      *service.ConstraintsService
	gameDetection   *game.Detector
	steamClient     workshopMetadataFetcher
	steamMeta       *steam.MetadataCache
	steamImage      *steam.ImageCache
	steamDesc       *steam.DescriptionImageCache
}

type App struct {
	ctx context.Context

	svc  appServices
	game game.Adapter

	settings       appSettings
	playsetNames   []string
	gameActiveIdx  int
	launcherIdx    int
	modPathByID    map[string]string
	launcherLayout LauncherLayout
	loadOrder      domain.LoadOrder
	gamePaths      domain.GamePaths

	imageDataMu sync.RWMutex
	imageData   map[string]string

	openURL      func(goos, rawURL string) error
	openInAppURL func(url string) error

	constraintsPath string
	settingsPath    string
	layoutPath      string
	activeGameID    domain.GameID
}

func (a *App) Startup(ctx context.Context) {
	logging.Infof("Startup: BEGIN (activeGameID=%q)", a.activeGameID)
	a.ctx = ctx
	a.initCoreServices()
	logging.Debugf("Startup: initCoreServices done")
	a.initLoadOrder()
	logging.Debugf("Startup: initLoadOrder done")
	a.initGamePaths()
	logging.Debugf("Startup: initGamePaths done (paths: %+v)", a.gamePaths)

	dir := filepath.Dir(a.svc.loadOrderRepo.Path())
	a.constraintsPath = filepath.Join(dir, constraintsFileName)
	a.settingsPath = filepath.Join(dir, settingsFileName)
	a.layoutPath = filepath.Join(dir, launcherLayoutFile)
	logging.Debugf("Startup: paths set (settings=%q, constraints=%q, layout=%q)", a.settingsPath, a.constraintsPath, a.layoutPath)

	loads := a.loadStartupState()
	logging.Debugf("Startup: state loaded (settingsErr=%v, constraintsErr=%v, layoutErr=%v)", loads.settingsErr, loads.constraintsErr, loads.layoutErr)
	a.applyStartupSettings(loads.settings, loads.settingsErr)
	a.loadStartupPlaysetState()
	a.applyStartupConstraints(loads.constraints, loads.constraintsErr)
	a.applyStartupLayout(loads.layout, loads.layoutErr)

	logging.Infof("Startup: COMPLETED (playsets=%q, localMods=%q, workshopRoots=%d, gameExeAuto=%q, gameExeEffective=%q, gameActive=%s, gameActiveIdx=%d, launcherIdx=%d)",
		a.gamePaths.PlaysetsPath, a.effectiveModsDir(), len(a.gamePaths.WorkshopModDirs),
		a.gamePaths.GameExePath, a.effectiveGameExe(), a.activeGameID, a.gameActiveIdx, a.launcherIdx)
}

func (a *App) initCoreServices() {
	a.svc.settingsRepo = repo.NewFileSettingsRepository()
	a.svc.layoutRepo = repo.NewFileLayoutRepository()
	a.svc.modsService = service.NewModsService()
	a.svc.loadorderSvc = service.NewLoadOrderService()
	a.svc.settingsSvc = service.NewSettingsService()
	a.svc.launchSvc = service.NewLaunchService()
	a.svc.gameSvc = service.NewGameService()
	a.svc.constraintsRepo = repo.NewFileConstraintsRepository()
	a.svc.gameDetection = game.NewDetector(a.svc.settingsRepo)
	a.svc.playsetRepo = repo.NewFilePlaysetRepo()
	a.svc.playsetSvc = service.NewPlaysetService(a.svc.playsetRepo)
	a.svc.steamClient = steam.NewClient()
	a.svc.conGraph = domain.NewGraph()
	a.openURL = a.svc.launchSvc.OpenURL
	a.openInAppURL = a.openURLInApp

	if adapter, err := a.svc.gameSvc.ResolveAdapter(a.activeGameID); err == nil {
		a.game = adapter
	}

	a.svc.layoutSvc = service.NewLayoutService(normalizeLauncherLayout, func(layout LauncherLayout) error {
		if strings.TrimSpace(a.layoutPath) == "" {
			return nil
		}
		return a.svc.layoutRepo.Save(a.layoutPath, toRepoLayout(layout))
	})
}

func (a *App) initLoadOrder() {
	path, err := loadorder.DefaultConfigPath()
	if err != nil {
		logging.Errorf("startup: resolve default loadorder path: %v", err)
		return
	}
	store, err := loadorder.New(path)
	if err != nil {
		logging.Errorf("startup: initialize loadorder store: %v", err)
		return
	}
	a.svc.loadOrderRepo = repo.NewFileLoadOrderRepo(store)

	state, err := a.svc.loadOrderRepo.Load()
	if err != nil {
		logging.Warnf("startup: load fallback loadorder state, using empty: %v", err)
		a.loadOrder = domain.LoadOrder{OrderedIDs: []string{}}
	} else {
		a.loadOrder = state
	}
}

func (a *App) initGamePaths() {
	var err error
	a.gamePaths, err = a.svc.gameSvc.DiscoverPaths(a.activeGameID)
	if err != nil {
		logging.Errorf("startup: auto-discover game paths: %v", err)
	}
	// Apply manual overrides (but settingsPath may not be set yet, so use loadOrderRepo path)
	if a.svc.loadOrderRepo != nil {
		settingsPath := filepath.Join(filepath.Dir(a.svc.loadOrderRepo.Path()), settingsFileName)
		if settings, err := a.svc.settingsRepo.Load(settingsPath); err == nil && settings.GamePaths != nil {
			if override, ok := settings.GamePaths[string(a.activeGameID)]; ok {
				if override.DocumentsDir != "" {
					a.gamePaths.LocalModsDir = override.DocumentsDir + "/mod"
					a.gamePaths.PlaysetsPath = override.DocumentsDir + "/playsets.json"
				}
				if override.InstallDir != "" {
					a.gamePaths.GameExePath = override.InstallDir
				}
			}
		}
	}
}

func (a *App) initConstraintsService() {
	a.svc.conService = service.NewConstraintsService(
		a.svc.conGraph, a.constraintsPath, a.svc.constraintsRepo,
		a.expandConstraintTarget, domain.IsCategoryID,
	)
}

func (a *App) ensureReady() error {
	if a.svc.loadOrderRepo == nil {
		return errAppStorageNotInitialized
	}
	if a.svc.conGraph == nil {
		a.svc.conGraph = domain.NewGraph()
	}
	if a.loadOrder.OrderedIDs == nil {
		a.loadOrder.OrderedIDs = []string{}
	}
	if a.modPathByID == nil {
		a.modPathByID = map[string]string{}
	}
	if a.playsetNames == nil {
		a.playsetNames = []string{}
	}
	if a.imageData == nil {
		a.imageData = map[string]string{}
	}
	if a.settingsPath == "" {
		a.settingsPath = filepath.Join(filepath.Dir(a.svc.loadOrderRepo.Path()), settingsFileName)
	}
	if a.layoutPath == "" {
		a.layoutPath = filepath.Join(filepath.Dir(a.svc.loadOrderRepo.Path()), launcherLayoutFile)
	}
	if a.coreServicesMissing() {
		a.initCoreServices()
	}
	if err := a.ensureSteamCaches(); err != nil {
		return err
	}
	if a.svc.conService == nil {
		a.initConstraintsService()
	}
	return nil
}

func (a *App) coreServicesMissing() bool { return a.svc.modsService == nil }

func (a *App) loadStartupState() startupLoads {
	var wg sync.WaitGroup
	wg.Add(3)
	var settings repo.AppSettingsData
	var settingsErr error
	var constraints *domain.Graph
	var constraintsErr error
	var layout repo.LauncherLayoutData
	var layoutErr error

	go func() {
		defer wg.Done()
		settings, settingsErr = a.svc.settingsRepo.Load(a.settingsPath)
	}()
	go func() {
		defer wg.Done()
		constraints, constraintsErr = a.svc.constraintsRepo.Load(a.constraintsPath)
	}()
	go func() {
		defer wg.Done()
		layout, layoutErr = a.svc.layoutRepo.Load(a.layoutPath)
	}()
	wg.Wait()
	return startupLoads{settings, settingsErr, constraints, constraintsErr, layout, layoutErr}
}

func (a *App) applyStartupSettings(settings repo.AppSettingsData, settingsErr error) {
	if settingsErr != nil {
		logging.Warnf("startup: load settings, using defaults: %v", settingsErr)
	}
	a.settings = fromRepoSettings(settings)
}

func (a *App) loadStartupPlaysetState() {
	logging.Debugf("loadStartupPlaysetState: BEGIN (PlaysetsPath=%q, activeGameID=%q)", a.gamePaths.PlaysetsPath, a.activeGameID)
	if a.gamePaths.PlaysetsPath == "" {
		logging.Debugf("loadStartupPlaysetState: PlaysetsPath is empty, skipping")
		return
	}
	names, idx, err := a.svc.gameSvc.ListModLists(a.activeGameID, a.gamePaths.PlaysetsPath)
	if err != nil {
		logging.Warnf("loadStartupPlaysetState: read playset list: %v", err)
		return
	}
	logging.Debugf("loadStartupPlaysetState: found %d playsets, gameActiveIdx=%d", len(names), idx)
	a.playsetNames = names
	a.gameActiveIdx = idx
	a.launcherIdx = a.svc.playsetSvc.ResolveLauncherIndex(len(names), idx, a.settings.LauncherActivePlaysetIndex)
	logging.Debugf("loadStartupPlaysetState: resolved launcherIdx=%d", a.launcherIdx)

	state, pathByID, loadErr := a.svc.gameSvc.ImportModList(a.activeGameID, a.gamePaths.PlaysetsPath, a.launcherIdx)
	if loadErr != nil {
		logging.Warnf("loadStartupPlaysetState: load selected playset state, using fallback: %v", loadErr)
		return
	}
	logging.Debugf("loadStartupPlaysetState: loaded %d ordered IDs, %d path mappings", len(state.OrderedIDs), len(pathByID))
	a.loadOrder = state
	for id, path := range pathByID {
		a.modPathByID[id] = path
	}
}

func (a *App) applyStartupConstraints(g *domain.Graph, err error) {
	if err != nil {
		logging.Warnf("startup: load constraints, using empty: %v", err)
		a.svc.conGraph = domain.NewGraph()
		a.initConstraintsService()
		return
	}
	if g == nil {
		g = domain.NewGraph()
	}
	a.svc.conGraph = g
	a.initConstraintsService()
}

func (a *App) applyStartupLayout(layout repo.LauncherLayoutData, err error) {
	if err != nil {
		logging.Warnf("startup: load launcher layout, using defaults: %v", err)
		layout = toRepoLayout(defaultLauncherLayout(a.loadOrder.OrderedIDs))
	}
	next := fromRepoLayout(layout)
	layoutErr := a.svc.layoutSvc.Persist(&next, a.loadOrder.OrderedIDs)
	a.launcherLayout = next
	if layoutErr != nil {
		logging.Warnf("startup: persist normalized launcher layout: %v", layoutErr)
	}
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

func (a *App) expandConstraintTarget(target string) []string {
	categoryMembers := make(map[string][]string)
	for _, cat := range a.launcherLayout.Categories {
		if cat.ID == target {
			return append([]string(nil), cat.ModIDs...)
		}
		categoryMembers[cat.ID] = cat.ModIDs
	}
	return nil
}
