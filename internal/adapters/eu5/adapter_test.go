package eu5

import (
	"encoding/json"
	"eu5-mod-launcher/internal/game"
	"os"
	"path/filepath"
	"testing"
)

func TestEU5Adapter_SavePlayset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "eu5_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	adapter := &Adapter{}
	inst := game.Instance{
		UserConfigPath: tmpDir,
	}

	playset := game.Playset{
		ID:   "test-playset",
		Name: "Test Playset",
		Entries: []game.ModEntry{
			{ID: "mod-1", Enabled: true, Position: 0},
			{ID: "mod-2", Enabled: false, Position: 1},
		},
	}

	err = adapter.SavePlayset(inst, playset)
	if err != nil {
		t.Errorf("SavePlayset failed: %v", err)
	}

	playsetPath := filepath.Join(tmpDir, "playsets", "test-playset.json")
	if _, err := os.Stat(playsetPath); os.IsNotExist(err) {
		t.Errorf("expected playset file to exist at %s", playsetPath)
	}

	content, err := os.ReadFile(playsetPath)
	if err != nil {
		t.Fatalf("failed to read playset file: %v", err)
	}

	var saved game.Playset
	if err := json.Unmarshal(content, &saved); err != nil {
		t.Fatalf("failed to unmarshal saved playset: %v", err)
	}

	if saved.ID != playset.ID || saved.Name != playset.Name {
		t.Errorf("saved playset mismatch: got %+v, want %+v", saved, playset)
	}

	if len(saved.Entries) != len(playset.Entries) {
		t.Errorf("saved playset entries length mismatch: got %d, want %d", len(saved.Entries), len(playset.Entries))
	}
}
