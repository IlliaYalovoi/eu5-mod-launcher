package legacy

import (
	"eu5-mod-launcher/internal/game"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func prepareTestDB(t *testing.T) (string, *sqlx.DB) {
	tmpDir, err := os.MkdirTemp("", "sqlite_test_")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "launcher-v2.sqlite")
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE playsets (
		id TEXT PRIMARY KEY,
		name TEXT,
		isRemoved INTEGER DEFAULT 0
	);
	CREATE TABLE playsets_mods (
		playsetId TEXT,
		modId TEXT,
		enabled INTEGER,
		position INTEGER
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return tmpDir, db
}

func TestSavePlayset(t *testing.T) {
	tmpDir, db := prepareTestDB(t)
	defer os.RemoveAll(tmpDir)
	defer db.Close()

	adapter := NewSqliteAdapter("test", "Test Game", "123")
	inst := game.Instance{UserConfigPath: tmpDir}

	playset := game.Playset{
		ID:   "p1",
		Name: "Test Playset",
		Entries: []game.ModEntry{
			{ID: "m1", Enabled: true, Position: 0},
			{ID: "m2", Enabled: false, Position: 1},
		},
	}

	// First save
	err := adapter.SavePlayset(inst, playset)
	if err != nil {
		t.Errorf("SavePlayset failed: %v", err)
	}

	// Verify
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM playsets_mods WHERE playsetId = ?", "p1")
	if err != nil {
		t.Fatalf("failed to count mods: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 mods, got %d", count)
	}

	// Update (re-save with changes)
	playset.Entries = []game.ModEntry{
		{ID: "m1", Enabled: false, Position: 0},
	}
	err = adapter.SavePlayset(inst, playset)
	if err != nil {
		t.Errorf("SavePlayset update failed: %v", err)
	}

	err = db.Get(&count, "SELECT COUNT(*) FROM playsets_mods WHERE playsetId = ?", "p1")
	if err != nil {
		t.Fatalf("failed to count mods after update: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 mod after update, got %d", count)
	}
}
