package legacy

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"eu5-mod-launcher/internal/game"

	_ "modernc.org/sqlite"
)

type SqliteAdapter struct{}

func (s *SqliteAdapter) ID() string {
	return "sqlite"
}

func (s *SqliteAdapter) LoadPlaysets(instance game.Instance) ([]game.Playset, error) {
	dbPath := filepath.Join(instance.UserConfigPath, "launcher-v2.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name FROM playsets WHERE isRemoved = 0")
	if err != nil {
		return nil, fmt.Errorf("failed to query playsets: %w", err)
	}
	defer rows.Close()

	var playsets []game.Playset
	for rows.Next() {
		var p game.Playset
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("failed to scan playset: %w", err)
		}

		entries, err := s.loadModEntries(db, p.ID)
		if err != nil {
			return nil, err
		}
		p.Entries = entries
		playsets = append(playsets, p)
	}

	return playsets, nil
}

func (s *SqliteAdapter) loadModEntries(db *sql.DB, playsetID string) ([]game.ModEntry, error) {
	query := `
		SELECT modId, enabled, position
		FROM playsets_mods
		WHERE playsetId = ?
		ORDER BY position ASC`

	rows, err := db.Query(query, playsetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query mod entries: %w", err)
	}
	defer rows.Close()

	var entries []game.ModEntry
	for rows.Next() {
		var e game.ModEntry
		var enabled int
		if err := rows.Scan(&e.ID, &enabled, &e.Position); err != nil {
			return nil, fmt.Errorf("failed to scan mod entry: %w", err)
		}
		e.Enabled = enabled != 0
		entries = append(entries, e)
	}

	return entries, nil
}
