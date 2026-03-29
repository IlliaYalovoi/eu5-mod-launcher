package main

import (
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	"eu5-mod-launcher/internal/loadorder"
	"eu5-mod-launcher/internal/service"
)

func newReadyAppForLaunchTest(t *testing.T) *App {
	t.Helper()

	storePath := filepath.Join(t.TempDir(), "loadorder.json")
	store, err := loadorder.New(storePath)
	if err != nil {
		t.Fatalf("loadorder.New() error = %v", err)
	}

	return &App{
		loStore:         store,
		conGraph:        nil,
		loState:         loadorder.State{OrderedIDs: []string{}},
		modPathByID:     map[string]string{},
		playsetNames:    []string{},
		settingsPath:    filepath.Join(filepath.Dir(storePath), settingsFileName),
		layoutPath:      filepath.Join(filepath.Dir(storePath), launcherLayoutFile),
		launcherLayout:  LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
		gameActiveIndex: -1,
		launcherIndex:   -1,
	}
}

func TestLaunchGame_MissingExecutable(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.settings.GameExe = ""
	app.gamePaths.GameExePath = ""

	err := app.LaunchGame()
	if err == nil {
		t.Fatalf("LaunchGame() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "executable path") {
		t.Fatalf("LaunchGame() error = %v, expected executable path context", err)
	}
}

func TestLaunchGame_InvalidExecutablePath(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.settings.GameExe = filepath.Join(t.TempDir(), "missing", "eu5.exe")

	err := app.LaunchGame()
	if err == nil {
		t.Fatalf("LaunchGame() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "stat executable") {
		t.Fatalf("LaunchGame() error = %v, expected stat executable context", err)
	}
}

func TestBuildLaunchCommand(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "eu5.exe")
	args := []string{"--foo", "bar"}
	svc := service.NewLaunchService()

	cmd := svc.BuildLaunchCommand(exe, args)
	if cmd == nil {
		t.Fatalf("BuildLaunchCommand() returned nil command")
	}
	if cmd.Path != exe {
		t.Fatalf("BuildLaunchCommand().Path = %q, want %q", cmd.Path, exe)
	}
	if len(cmd.Args) != 3 {
		t.Fatalf("BuildLaunchCommand().Args len = %d, want 3", len(cmd.Args))
	}
	if cmd.Args[0] != exe || cmd.Args[1] != "--foo" || cmd.Args[2] != "bar" {
		t.Fatalf("BuildLaunchCommand().Args = %v", cmd.Args)
	}
	if cmd.Dir != filepath.Dir(exe) {
		t.Fatalf("BuildLaunchCommand().Dir = %q, want %q", cmd.Dir, filepath.Dir(exe))
	}
	if cmd.SysProcAttr == nil {
		t.Fatalf("BuildLaunchCommand().SysProcAttr = nil, want detached attrs")
	}
}

func TestShouldLaunchViaSteam(t *testing.T) {
	steamPath := filepath.Join("C:", "Program Files (x86)", "Steam", "steamapps", "common", "Europa Universalis V", "binaries", "eu5.exe")
	svc := service.NewSettingsService()
	got := svc.ShouldLaunchViaSteam(goruntime.GOOS, steamPath)
	if goruntime.GOOS == "windows" && !got {
		t.Fatalf("shouldLaunchViaSteam(%q) = false, want true on windows", steamPath)
	}
	if goruntime.GOOS != "windows" && got {
		t.Fatalf("shouldLaunchViaSteam(%q) = true, want false on non-windows", steamPath)
	}
}

func TestBuildSteamLaunchCommand(t *testing.T) {
	svc := service.NewLaunchService()
	cmd, err := svc.BuildSteamLaunchCommand(goruntime.GOOS, eu5SteamAppID)
	if err != nil {
		t.Fatalf("BuildSteamLaunchCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatalf("BuildSteamLaunchCommand() returned nil command")
	}
	if cmd.SysProcAttr == nil {
		t.Fatalf("BuildSteamLaunchCommand().SysProcAttr = nil, want detached attrs")
	}
}
