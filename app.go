package main

import (
	"errors"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/steam"
	"fmt"
)

type startupLoads struct {
	settings       repo.AppSettingsData
	settingsErr    error
	constraints    *graph.Graph
	constraintsErr error
	layout         repo.LauncherLayoutData
	layoutErr      error
}

const (
	constraintsFileName = "constraints.json"
	settingsFileName    = "settings.json"
	launcherLayoutFile = "launcher_layout.json"
	eu5SteamAppID      = "3450310"
)

var (
	errLauncherCategoryNameEmpty = errors.New("launcher category name must not be empty")
	errAppStorageNotInitialized  = errors.New("app storage is not initialized")
	errExternalLinkInvalid       = errors.New("external link is invalid")
	errUnsubscribeDisabled       = errors.New("unsubscribe feature is disabled")
	errWorkshopItemIDInvalid     = errors.New("workshop item id is invalid")
)

type workshopMetadataFetcher interface {
	FetchWorkshopMetadata(ids []string) (map[string]steam.WorkshopItem, error)
}

func NewApp() *App {
	return &App{
		loadOrder:     domain.LoadOrder{OrderedIDs: []string{}},
		modPathByID:   map[string]string{},
		launcherLayout: LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
		imageData:     map[string]string{},
		playsetNames:  []string{},
		gameActiveIdx: -1,
		launcherIdx:   -1,
		activeGameID:  domain.GameIDEU5,
	}
}

func (*App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
