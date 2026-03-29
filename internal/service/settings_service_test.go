package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsServiceNormalizeModsDir(t *testing.T) {
	svc := NewSettingsService()
	dir := t.TempDir()
	got, err := svc.NormalizeModsDir(dir)
	if err != nil {
		t.Fatalf("NormalizeModsDir() error = %v", err)
	}
	if got == "" {
		t.Fatalf("NormalizeModsDir() returned empty path")
	}
}

func TestSettingsServiceNormalizeGameExe(t *testing.T) {
	svc := NewSettingsService()
	exe := filepath.Join(t.TempDir(), "eu5.exe")
	if err := os.WriteFile(exe, []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := svc.NormalizeGameExe(exe)
	if err != nil {
		t.Fatalf("NormalizeGameExe() error = %v", err)
	}
	if got == "" {
		t.Fatalf("NormalizeGameExe() returned empty path")
	}
}

func TestSettingsServiceShouldLaunchViaSteam(t *testing.T) {
	svc := NewSettingsService()
	path := filepath.Join("C:", "Program Files (x86)", "Steam", "steamapps", "common", "Europa Universalis V", "binaries", "eu5.exe")
	if !svc.ShouldLaunchViaSteam("windows", path) {
		t.Fatalf("ShouldLaunchViaSteam() = false, want true")
	}
	if svc.ShouldLaunchViaSteam("linux", path) {
		t.Fatalf("ShouldLaunchViaSteam() = true on linux, want false")
	}
}
