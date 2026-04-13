package main

import (
	"embed"
	"log/slog"
	"os"

	"eu5-mod-launcher/internal/launcher"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/steam"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	deps := launcher.Dependencies{
		SettingsRepo:    repo.NewFileSettingsRepository(),
		ConstraintsRepo: repo.NewFileConstraintsRepository(),
		LayoutRepo:     repo.NewFileLayoutRepository(),
		PlaysetRepo:    repo.NewFilePlaysetRepo(),
		LoadOrderRepo:  repo.NewFileLoadOrderRepo(nil),
		SteamClient:    steam.NewClient(),
	}
	app := launcher.NewLauncher(deps)

	err := wails.Run(&options.App{
		Title:  "eu5-mod-launcher",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		slog.Error("wails run failed", "err", err)
	}
}