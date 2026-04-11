package service

import (
	"os"
	"path/filepath"
	"testing"

	"eu5-mod-launcher/internal/repo"
)

// mockSettingsRepo implements SettingsRepository for testing.
type mockSettingsRepo struct {
	data map[string]repo.AppSettingsData
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{data: make(map[string]repo.AppSettingsData)}
}

func (m *mockSettingsRepo) Load(path string) (repo.AppSettingsData, error) {
	if d, ok := m.data[path]; ok {
		return d, nil
	}
	return repo.AppSettingsData{}, nil
}

func (m *mockSettingsRepo) Save(path string, settings repo.AppSettingsData) error {
	m.data[path] = settings
	return nil
}

func TestListSupportedGames_Default(t *testing.T) {
	svc := NewGameDetectionService(newMockSettingsRepo())

	games, err := svc.ListSupportedGames("/nonexistent/path.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(games) != 2 {
		t.Errorf("expected 2 games, got %d", len(games))
	}

	// Should have eu5 and vic3
	ids := make(map[string]bool)
	for _, g := range games {
		ids[g.ID] = true
	}
	if !ids["eu5"] {
		t.Error("expected eu5 in results")
	}
	if !ids["vic3"] {
		t.Error("expected vic3 in results")
	}
}

func TestListSupportedGames_SortingDetectedFirst(t *testing.T) {
	tmpDir := t.TempDir()
	eu5Install := filepath.Join(tmpDir, "steamapps", "common", "Europa Universalis V")

	if err := os.MkdirAll(eu5Install, 0o755); err != nil {
		t.Fatalf("create test dir: %v", err)
	}

	mock := newMockSettingsRepo()
	svc := NewGameDetectionService(mock)

	// Override paths - set install dir which exists
	_ = svc.SetGamePaths("/fake/path", "eu5", eu5Install, "")
	_ = svc.SetGamePaths("/fake/path", "vic3", "", "") // not detected

	games, err := svc.ListSupportedGames("/fake/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First game should be detected (eu5 with valid install dir)
	if !games[0].Detected {
		t.Errorf("expected first game (eu5) to be detected, got detected=%v", games[0].Detected)
	}
}

func TestSetGamePaths_PersistsOverrides(t *testing.T) {
	mock := newMockSettingsRepo()
	svc := NewGameDetectionService(mock)

	err := svc.SetGamePaths("/test/settings.json", "eu5", "/custom/eu5", "/custom/docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	games, err := svc.ListSupportedGames("/test/settings.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, g := range games {
		if g.ID == "eu5" {
			if g.InstallDir != "/custom/eu5" {
				t.Errorf("expected install dir /custom/eu5, got %s", g.InstallDir)
			}
			if g.DocumentsDir != "/custom/docs" {
				t.Errorf("expected docs dir /custom/docs, got %s", g.DocumentsDir)
			}
		}
	}
}

func TestSetGamePaths_UnknownGameIgnored(t *testing.T) {
	mock := newMockSettingsRepo()
	svc := NewGameDetectionService(mock)

	// Should not error on unknown game
	err := svc.SetGamePaths("/test/settings.json", "unknown_game", "/path", "/docs")
	if err != nil {
		t.Errorf("expected no error for unknown game, got %v", err)
	}
}

func TestListSupportedGames_StableOrdering(t *testing.T) {
	mock := newMockSettingsRepo()
	svc := NewGameDetectionService(mock)

	// Run multiple times
	var firstOrder []string
	for i := 0; i < 3; i++ {
		games, err := svc.ListSupportedGames("/test/settings.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		order := make([]string, len(games))
		for j, g := range games {
			order[j] = g.ID
		}

		if i == 0 {
			firstOrder = order
		} else {
			for j, id := range order {
				if id != firstOrder[j] {
					t.Errorf("order not stable: iteration %d had %v, iteration 0 had %v", i, order, firstOrder)
					break
				}
			}
		}
	}
}