package main

import (
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	"eu5-mod-launcher/internal/loadorder"
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

	cmd := buildLaunchCommand(exe, args)
	if cmd == nil {
		t.Fatalf("buildLaunchCommand() returned nil command")
	}
	if cmd.Path != exe {
		t.Fatalf("buildLaunchCommand().Path = %q, want %q", cmd.Path, exe)
	}
	if len(cmd.Args) != 3 {
		t.Fatalf("buildLaunchCommand().Args len = %d, want 3", len(cmd.Args))
	}
	if cmd.Args[0] != exe || cmd.Args[1] != "--foo" || cmd.Args[2] != "bar" {
		t.Fatalf("buildLaunchCommand().Args = %v", cmd.Args)
	}
	if cmd.Dir != filepath.Dir(exe) {
		t.Fatalf("buildLaunchCommand().Dir = %q, want %q", cmd.Dir, filepath.Dir(exe))
	}
	if cmd.SysProcAttr == nil {
		t.Fatalf("buildLaunchCommand().SysProcAttr = nil, want detached attrs")
	}
}

func TestShouldLaunchViaSteam(t *testing.T) {
	steamPath := filepath.Join("C:", "Program Files (x86)", "Steam", "steamapps", "common", "Europa Universalis V", "binaries", "eu5.exe")
	got := shouldLaunchViaSteam(steamPath)
	if goruntime.GOOS == "windows" && !got {
		t.Fatalf("shouldLaunchViaSteam(%q) = false, want true on windows", steamPath)
	}
	if goruntime.GOOS != "windows" && got {
		t.Fatalf("shouldLaunchViaSteam(%q) = true, want false on non-windows", steamPath)
	}
}

func TestBuildSteamLaunchCommand(t *testing.T) {
	cmd, err := buildSteamLaunchCommand(nil)
	if err != nil {
		t.Fatalf("buildSteamLaunchCommand() error = %v", err)
	}
	if cmd == nil {
		t.Fatalf("buildSteamLaunchCommand() returned nil command")
	}
	if cmd.SysProcAttr == nil {
		t.Fatalf("buildSteamLaunchCommand().SysProcAttr = nil, want detached attrs")
	}
}
