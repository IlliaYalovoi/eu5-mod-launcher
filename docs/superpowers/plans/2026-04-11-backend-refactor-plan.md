# Backend Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the EU5 Mod Launcher Go backend to reduce code by >30%, eliminate boilerplate, replace logrus with slog, add pkgerrors and mapstructure, reorganize into 3 bounded contexts, and implement manual DI via factory.

**Architecture:** Manual dependency injection via `NewLauncher(deps Dependencies) *App` factory in `wire.go`. Bounded contexts: `launcher/` (Wails App + mod scanning + loadorder + constraints + graph), `game/` (detection + adapters + launch), `steam/` (workshop + metadata + cache). Sentinel errors with `Err*` prefix, `fmt.Errorf` wrapping, `pkgerrors.Wrap` at Wails boundaries only.

**Tech Stack:** Go 1.26, Wails v2, logrus→slog, pkgerrors, mapstructure, resty, gjson, xdg

---

## File Structure (Post-Refactor)

```
internal/
  domain/          # Shared types only
    constraint.go
    errors.go
    game.go
    loadorder.go
    mod.go

  launcher/        # Wails-exposed App, mod scanning, loadorder, constraints, graph
    app_mods.go
    app_game.go
    app_constraints.go
    app_layout.go
    app_workshop.go
    app_conversion.go
    app_structs.go
    settings.go
    wire.go        # NewApp factory, Dependencies struct
    loadorder.go
    playsets.go
    graph.go

  game/            # Game detection, adapters, launch process
    detection.go
    adapter.go
    eu5.go
    launch.go

  steam/           # Workshop, metadata, image cache
    workshop.go
    client.go
    cache.go
    metadata.go
    images.go
    descriptions.go
    steam.go       # steamAppID const + helpers

  repo/            # Interfaces + file-backed implementations
    constraints_repo.go
    layout_repo.go
    settings_repo.go
    playset_repo.go
    loadorder_repo.go

  service/         # Thin orchestration layer (unchanged structurally)
    constraints_service.go
    loadorder_service.go
    mods_service.go
    game_service.go
    layout_service.go
    playset_service.go
    settings_service.go
    launch_service.go

main (root)/
  main.go          # Entry point, wires launcher via factory
```

---

## Dependency Changes

| Dependency | Action |
|---|---|
| `github.com/sirupsen/logrus` | **Remove** — replace with `log/slog` |
| `github.com/pkgerrors` | **Add** — error stack traces |
| `github.com/mitchellh/mapstructure` | **Add** — struct mapping |
| `gopkg.in/yaml.v3` | Keep |
| `resty`, `gjson`, `xdg` | Keep |

---

## Task 1: Prune Dead Code and Empty Files

**Files:**
- Delete: `internal/domain/parse.go` (empty)
- Delete: `internal/domain/types.go` (empty)
- Delete: `internal/game/contracts.go` (naked type aliases — import domain directly)
- Delete: `feature_unsubscribe_disabled.go` (build-tag file, merge logic into app_workshop.go)
- Delete: `feature_unsubscribe_enabled.go` (build-tag file, merge logic into app_workshop.go)
- Delete: `launcher_layout.go` (merged into launcher/app_layout.go)

**Steps:**

- [ ] **Step 1: Delete empty domain files**

```bash
rm /home/illia/code/eu5-mod-launcher/internal/domain/parse.go
rm /home/illia/code/eu5-mod-launcher/internal/domain/types.go
```

- [ ] **Step 2: Delete game/contracts.go**

```bash
rm /home/illia/code/eu5-mod-launcher/internal/game/contracts.go
```

Update imports in files that reference `game.GameDescriptor`, `game.GameID`, `game.GameIDEU5`, `game.GameIDVic3` — they should import `eu5-mod-launcher/internal/domain` directly. Find affected files:

```bash
grep -rl "game\.GameID\|game\.GameIDEU5\|game\.GameIDVic3\|game\.GameDescriptor" /home/illia/code/eu5-mod-launcher --include="*.go"
```

Update each to use `domain.GameID`, `domain.GameIDEU5`, `domain.GameIDVic3`, `domain.GameDescriptor` instead.

- [ ] **Step 3: Delete build-tag feature_unsubscribe files**

```bash
rm /home/illia/code/eu5-mod-launcher/feature_unsubscribe_disabled.go
rm /home/illia/code/eu5-mod-launcher/feature_unsubscribe_enabled.go
```

- [ ] **Step 4: Delete launcher_layout.go**

```bash
rm /home/illia/code/eu5-mod-launcher/launcher_layout.go
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor: prune dead code and empty files"
```

---

## Task 2: Add New Dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Steps:**

- [ ] **Step 1: Add pkgerrors and mapstructure to go.mod**

```bash
cd /home/illia/code/eu5-mod-launcher
go get github.com/pkg/errors@v0.9.1
go get github.com/mitchellh/mapstructure@v1.5.2
```

- [ ] **Step 2: Remove logrus from go.mod**

```bash
go mod edit -droprequire github.com/sirupsen/logrus
go mod tidy
```

- [ ] **Step 3: Verify build still passes**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "refactor: remove logrus, add pkgerrors and mapstructure"
```

---

## Task 3: Create New Package Structure — Move Files

**Files:**
- Create: `internal/launcher/` (directory)
- Create: `internal/game/` (directory)
- Create: `internal/steam/` (directory)

**Steps:**

- [ ] **Step 1: Create launcher/ directory and move app_*.go files**

```bash
mkdir -p /home/illia/code/eu5-mod-launcher/internal/launcher
# Move app_*.go files
mv /home/illia/code/eu5-mod-launcher/app_mods.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/app_game.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/app_constraints.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/app_layout.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/app_workshop.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/app_conversion.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/app_structs.go /home/illia/code/eu5-mod-launcher/internal/launcher/
mv /home/illia/code/eu5-mod-launcher/settings.go /home/illia/code/eu5-mod-launcher/internal/launcher/
```

- [ ] **Step 2: Move loadorder/ files to launcher/**

```bash
mv /home/illia/code/eu5-mod-launcher/internal/loadorder/*.go /home/illia/code/eu5-mod-launcher/internal/launcher/ 2>/dev/null || true
# Move playsets.go and graph if separate
mv /home/illia/code/eu5-mod-launcher/internal/loadorder/playsets.go /home/illia/code/eu5-mod-launcher/internal/launcher/ 2>/dev/null || true
```

- [ ] **Step 3: Move graph/ files to launcher/**

```bash
mv /home/illia/code/eu5-mod-launcher/internal/graph/*.go /home/illia/code/eu5-mod-launcher/internal/launcher/
```

- [ ] **Step 4: Move steam/ files to internal/steam/ (rename existing)**

```bash
# The existing steam package is at internal/steam already, but rename internal/steam to internal/steam to match design
# Check if steam files already in internal/steam
ls /home/illia/code/eu5-mod-launcher/internal/steam/
```

If files exist at `internal/steam/`, they are already in the right place. The design wants `steam.go` with `steamAppID` const added.

- [ ] **Step 5: Update package declarations in all moved files**

For each moved file, change `package main` to `package launcher`.

```bash
# Update package declarations
sed -i 's/^package main$/package launcher/' /home/illia/code/eu5-mod-launcher/internal/launcher/*.go
```

- [ ] **Step 6: Move game detection service to game/**

```bash
mkdir -p /home/illia/code/eu5-mod-launcher/internal/game
mv /home/illia/code/eu5-mod-launcher/internal/service/game_detection_service.go /home/illia/code/eu5-mod-launcher/internal/game/detection.go
# Rename the service type to Detector
sed -i 's/type GameDetectionService struct/type Detector struct/' /home/illia/code/eu5-mod-launcher/internal/game/detection.go
sed -i 's/NewGameDetectionService/NewDetector/' /home/illia/code/eu5-mod-launcher/internal/game/detection.go
sed -i 's/func NewGameDetectionService/func NewDetector/' /home/illia/code/eu5-mod-launcher/internal/game/detection.go
sed -i 's/GameDetectionService/Detector/g' /home/illia/code/eu5-mod-launcher/internal/game/detection.go
```

- [ ] **Step 7: Move game adapter files**

```bash
mv /home/illia/code/eu5-mod-launcher/internal/game/adapter.go /home/illia/code/eu5-mod-launcher/internal/game/
mv /home/illia/code/eu5-mod-launcher/internal/game/eu5/adapter.go /home/illia/code/eu5-mod-launcher/internal/game/eu5.go
rmdir /home/illia/code/eu5-mod-launcher/internal/game/eu5 2>/dev/null || true
```

- [ ] **Step 8: Create steam/steam.go with steamAppID**

```bash
# Create steam.go with the single const
cat > /home/illia/code/eu5-mod-launcher/internal/steam/steam.go << 'EOF'
package steam

const steamAppID = "3450310"
EOF
```

- [ ] **Step 9: Verify go build**

```bash
cd /home/illia/code/eu5-mod-launcher && go build ./...
```

Fix any import or package errors. Expect failures in main.go and app_wiring.go (which are handled in later tasks).

- [ ] **Step 10: Commit**

```bash
git add -A && git commit -m "refactor: create new package structure (launcher/, game/, steam/)"
```

---

## Task 4: Replace Logrus with slog

**Files:**
- Replace: `internal/logging/logger.go`

**Steps:**

- [ ] **Step 1: Rewrite logger.go with slog**

```go
// internal/logging/logger.go
package logging

import (
	"log/slog"
	"os"
	"strings"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: parseLevel(os.Getenv("EU5_LOG_LEVEL")),
}))

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Debugf(format string, args ...any) {
	logger.Debug(format, args...)
}

func Infof(format string, args ...any) {
	logger.Info(format, args...)
}

func Warnf(format string, args ...any) {
	logger.Warn(format, args...)
}

func Errorf(format string, args ...any) {
	logger.Error(format, args...)
}
```

- [ ] **Step 2: Run go build to verify**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "refactor: replace logrus with slog stdlib"
```

---

## Task 5: Implement wire.go Factory

**Files:**
- Create: `internal/launcher/wire.go`
- Modify: `main.go`
- Modify: `internal/launcher/app.go` (rename from root app.go)

**Steps:**

- [ ] **Step 1: Rename root app.go to internal/launcher/app.go.old**

```bash
mv /home/illia/code/eu5-mod-launcher/app.go /home/illia/code/eu5-mod-launcher/internal/launcher/app.go.old
```

- [ ] **Step 2: Read the full app_wiring.go to understand App struct and initCoreServices**

Read `/home/illia/code/eu5-mod-launcher/app_wiring.go` fully (it's ~300 lines). Key things to extract:
- App struct fields
- initCoreServices() construction logic
- startup() lifecycle
- appServices struct

- [ ] **Step 3: Create internal/launcher/wire.go**

```go
// internal/launcher/wire.go
package launcher

import (
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
	"eu5-mod-launcher/internal/steam"
	"eu5-mod-launcher/internal/game"
)

type Dependencies struct {
	SettingsRepo    repo.SettingsRepository
	ConstraintsRepo repo.ConstraintsRepository
	LayoutRepo     repo.LayoutRepository
	PlaysetRepo    repo.PlaysetRepository
	LoadOrderRepo  repo.LoadOrderRepo
	SteamClient    *steam.Client
	MetadataCache  *steam.MetadataCache
	ImageCache     *steam.ImageCache
	GameDetector   *game.Detector
}

func NewLauncher(deps Dependencies) *App {
	a := &App{
		svc: appServices{
			settingsRepo:    deps.SettingsRepo,
			constraintsRepo: deps.ConstraintsRepo,
			layoutRepo:     deps.LayoutRepo,
			playsetRepo:    deps.PlaysetRepo,
			loadOrderRepo:  deps.LoadOrderRepo,
			steamClient:    deps.SteamClient,
			steamMeta:      deps.MetadataCache,
			steamImage:     deps.ImageCache,
			gameDetection:  deps.GameDetector,
		},
	}
	a.initCoreServices()
	return a
}
```

Note: `appServices` and `initCoreServices` will be moved from app_wiring.go into the launcher package. The `mustBeReady()` guard will be added as a new method.

- [ ] **Step 4: Update main.go to use NewLauncher**

```go
package main

import (
	"context"
	"embed"
	"log/slog"
	"os"

	"eu5-mod-launcher/internal/launcher"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/steam"
	"eu5-mod-launcher/internal/game"

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
		LoadOrderRepo:  repo.NewFileLoadOrderRepo(),
		SteamClient:    steam.NewClient(),
		MetadataCache:  steam.NewMetadataCache(),
		ImageCache:     steam.NewImageCache(),
		GameDetector:   game.NewDetector(),
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
```

- [ ] **Step 5: Build and fix errors iteratively**

```bash
go build ./...
```

Fix type mismatches, missing fields, and interface implementations one by one. This step is intentionally iterative — read error messages and fix.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor: implement wire.go factory with Dependencies struct"
```

---

## Task 6: Add mustBeReady() Guard to All Wails Methods

**Files:**
- Modify: `internal/launcher/app_*.go` (all Wails-exposed files)

**Steps:**

- [ ] **Step 1: Add ErrNotInitialized to domain/errors.go**

```go
// Add to internal/domain/errors.go
var ErrNotInitialized = errors.New("app not initialized")
```

- [ ] **Step 2: Add mustBeReady() to App struct in launcher package**

Add to `internal/launcher/app.go` (or a new file if App struct is split):

```go
func (a *App) mustBeReady() error {
	if !a.initialized {
		return domain.ErrNotInitialized
	}
	return nil
}
```

Note: The `initialized` field may not exist yet — it may be called `appStorageInitialized` or similar. Find the field that tracks startup completion in the current app_wiring.go or app.go startup() method.

- [ ] **Step 3: Add mustBeReady guard to every Wails method in app_*.go files**

For each public method (capitalized) in `internal/launcher/`, add the guard at the start:

```go
func (a *App) GetAllMods() ([]*mods.Mod, error) {
	if err := a.mustBeReady(); err != nil {
		return nil, err
	}
	// existing method body
}
```

Use grep to find all exported methods in the launcher package:

```bash
grep -n "^func (a \*App) [A-Z]" /home/illia/code/eu5-mod-launcher/internal/launcher/*.go
```

Apply to each method.

- [ ] **Step 4: Build and verify**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor: add mustBeReady() guard to all Wails methods"
```

---

## Task 7: Deduplicate launch_process Files

**Files:**
- Delete: `launch_process_unix.go` (root)
- Delete: `launch_process_windows.go` (root)
- Delete: `internal/service/launch_process_unix.go`
- Delete: `internal/service/launch_process_windows.go`
- Create: `internal/game/launch.go`

**Steps:**

- [ ] **Step 1: Read both launch_process files to understand content**

Read `launch_process_unix.go` and `launch_process_windows.go` at root and in `internal/service/`. These are duplicated.

- [ ] **Step 2: Create internal/game/launch.go with merged content**

Combine the platform-specific logic into a single file with build tags:

```go
// internal/game/launch.go
package game

//go:build linux || darwin
// +build linux darwin

func applyDetachedProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
```

And for Windows:

```go
// internal/game/launch_windows.go
package game

//go:build windows
// +build windows

func applyDetachedProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
```

- [ ] **Step 3: Delete the duplicate files**

```bash
rm /home/illia/code/eu5-mod-launcher/launch_process_unix.go
rm /home/illia/code/eu5-mod-launcher/launch_process_windows.go
rm /home/illia/code/eu5-mod-launcher/internal/service/launch_process_unix.go
rm /home/illia/code/eu5-mod-launcher/internal/service/launch_process_windows.go
```

- [ ] **Step 4: Update imports in launch_service.go**

```go
// internal/service/launch_service.go should import "eu5-mod-launcher/internal/game"
// and call game.ApplyDetachedProcessAttributes(cmd) instead of the local function
```

- [ ] **Step 5: Build and verify**

```bash
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor: deduplicate launch_process files into internal/game/launch.go"
```

---

## Task 8: Single steamAppID Const

**Files:**
- Modify: Files that define `eu5SteamAppID = "3450310"`
- Ensure: `internal/steam/steam.go` has the single definition

**Steps:**

- [ ] **Step 1: Find all occurrences of steam app ID literals**

```bash
grep -rn "3450310\|steamAppID\|eu5SteamAppID" /home/illia/code/eu5-mod-launcher --include="*.go"
```

Common locations: `app_wiring.go`, `loadorder/paths.go`, `game_detection_service.go`.

- [ ] **Step 2: Update each to use internal/steam.SteamAppID**

Import `eu5-mod-launcher/internal/steam` and use `steam.SteamAppID` in each file.

- [ ] **Step 3: Ensure internal/steam/steam.go exports it**

```go
// internal/steam/steam.go
package steam

const SteamAppID = "3450310" // exported for use by other packages
```

- [ ] **Step 4: Build and verify**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor: single steamAppID const in internal/steam/steam.go"
```

---

## Task 9: Replace Manual Struct Mapping with mapstructure

**Files:**
- Modify: `internal/repo/settings_repo.go`
- Modify: `internal/repo/layout_repo.go`
- Modify: Files using `toRepoSettings`, `fromRepoSettings`, `toRepoLayout`, `fromRepoLayout`

**Steps:**

- [ ] **Step 1: Find all manual mapping functions**

```bash
grep -rn "toRepoSettings\|fromRepoSettings\|toRepoLayout\|fromRepoLayout" /home/illia/code/eu5-mod-launcher --include="*.go"
```

- [ ] **Step 2: Replace each with mapstructure.Decode**

For example, replace:

```go
func toRepoSettings(from AppSettings) repo.AppSettingsData {
	return repo.AppSettingsData{
		ModsDir:    from.ModsDir,
		GameExe:    from.GameExe,
		GameArgs:   from.GameArgs,
		// ...
	}
}
```

With:

```go
var repoData repo.AppSettingsData
if err := mapstructure.Decode(from, &repoData); err != nil {
	return repo.AppSettingsData{}, fmt.Errorf("map settings to repo: %w", err)
}
```

- [ ] **Step 3: Add mapstructure import**

```go
import "github.com/mitchellh/mapstructure"
```

- [ ] **Step 4: Build and verify**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor: replace manual struct mapping with mapstructure"
```

---

## Task 10: Remove Old Root Files and Verify Build

**Files:**
- Delete: `app.go` (moved to launcher/)
- Delete: `app_aux.go` (consolidated)
- Delete: `app_wiring.go` (consolidated into wire.go)
- Delete: `main.go` (replaced with new main.go)
- Delete: `app_structs.go` (merged into launcher/)
- Delete: `settings.go` (merged into launcher/)
- Delete: `launch_process_*.go` (handled in Task 7)
- Delete: `feature_unsubscribe_*.go` (handled in Task 1)
- Delete: `launcher_layout.go` (handled in Task 1)

**Steps:**

- [ ] **Step 1: Delete old root files**

```bash
cd /home/illia/code/eu5-mod-launcher
# These should mostly already be deleted or moved, but verify:
ls -la *.go 2>/dev/null | grep -v main.go
# Delete app_*.go files at root (they've been moved to launcher/):
rm -f app.go app_aux.go app_wiring.go app_mods.go app_game.go app_constraints.go app_layout.go app_workshop.go app_conversion.go app_structs.go settings.go
```

- [ ] **Step 2: Remove old internal/ directories if empty**

```bash
rmdir internal/logging 2>/dev/null || true  # logging replaced
rmdir internal/graph 2>/dev/null || true   # moved to launcher
rmdir internal/loadorder 2>/dev/null || true  # moved to launcher
```

- [ ] **Step 3: Full build verification**

```bash
go build ./...
```

Expected: clean build with zero errors.

- [ ] **Step 4: Run existing tests**

```bash
go test ./...
```

- [ ] **Step 5: Final commit**

```bash
git add -A && git commit -m "refactor: remove old root files, final build verification"
```

---

## Post-Implementation Verification Checklist

- [ ] `go build ./...` passes with zero errors
- [ ] `go test ./...` passes (existing tests)
- [ ] No `github.com/sirupsen/logrus` imports remain
- [ ] `steamAppID` defined in exactly one place: `internal/steam/steam.go`
- [ ] `launch_process` files exist in exactly one place: `internal/game/launch.go` (+ `launch_windows.go`)
- [ ] `domain/parse.go` and `domain/types.go` deleted
- [ ] `feature_unsubscribe_*.go` build-tag files deleted
- [ ] `game/contracts.go` deleted (domain imported directly)
- [ ] `launcher_layout.go` deleted
- [ ] All Wails methods have `mustBeReady()` guard
- [ ] `pkgerrors` and `mapstructure` imported in files that need them
- [ ] `slog` used instead of logrus throughout
