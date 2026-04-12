package launcher

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/game"
	gamevic3 "eu5-mod-launcher/internal/game/vic3"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
	"eu5-mod-launcher/internal/steam"
)

type Dependencies struct {
	SettingsRepo    repo.SettingsRepository
	ConstraintsRepo repo.ConstraintsRepository
	LayoutRepo      repo.LayoutRepository
	PlaysetRepo     repo.PlaysetRepository
	LoadOrderRepo   repo.LoadOrderRepo
	SteamClient     *steam.Client
}

func NewLauncher(deps Dependencies) *App {
	a := &App{
		activeGameID: domain.GameIDEU5,
		loadOrder:    domain.LoadOrder{ActiveModIDs: []string{}},
		modPathByID:  map[string]string{},
		launcherLayout: LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
		imageData:    map[string]string{},
		playsetNames: []string{},
		gameActiveIdx: -1,
		launcherIdx:  -1,
	}
	a.svc.settingsRepo = deps.SettingsRepo
	a.svc.layoutRepo = deps.LayoutRepo
	a.svc.constraintsRepo = deps.ConstraintsRepo
	a.svc.playsetRepo = deps.PlaysetRepo
	a.svc.loadOrderRepo = deps.LoadOrderRepo
	a.svc.steamClient = deps.SteamClient
	a.svc.gameDetection = game.NewDetector(deps.SettingsRepo)
	a.svc.modsService = service.NewModsService()
	a.svc.loadorderSvc = service.NewLoadOrderService()
	a.svc.settingsSvc = service.NewSettingsService()
	a.svc.launchSvc = service.NewLaunchService()
	a.svc.gameSvc = service.NewGameService(
		game.NewEU5Adapter(nil),
		gamevic3.NewVic3Adapter(),
	)
	a.svc.playsetSvc = service.NewPlaysetService(deps.PlaysetRepo)
	a.svc.conGraph = domain.NewGraph()
	return a
}
