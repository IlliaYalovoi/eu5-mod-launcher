package vic3

import (
	"database/sql"
	"errors"
	"fmt"

	"eu5-mod-launcher/internal/domain"

	_ "modernc.org/sqlite"
)

var ErrNoActivePlayset = errors.New("no active playset found")

type SQLitePlaysetRepo struct{}

func (*SQLitePlaysetRepo) ListPlaysets(dbPath string) ([]string, domain.PlaysetIndex, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, 0, fmt.Errorf("open sqlite db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, name FROM playsets
		WHERE isRemoved = 0
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, 0, fmt.Errorf("query playsets: %w", err)
	}
	defer rows.Close()

	names := make([]string, 0)
	var activeIdx domain.PlaysetIndex
	idx := domain.PlaysetIndex(0)

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, 0, fmt.Errorf("scan playset row: %w", err)
		}
		names = append(names, name)
		idx++
	}

	// Find the active playset index.
	activeRows, err := db.Query(`SELECT name FROM playsets WHERE isRemoved = 0 AND isActive = 1 ORDER BY name ASC`)
	if err != nil {
		return names, 0, fmt.Errorf("query active playset: %w", err)
	}
	defer activeRows.Close()

	activeName := ""
	if activeRows.Next() {
		if err := activeRows.Scan(&activeName); err != nil {
			return names, 0, fmt.Errorf("scan active playset: %w", err)
		}
	}

	for i, name := range names {
		if name == activeName {
			activeIdx = domain.PlaysetIndex(i)
			break
		}
	}

	return names, activeIdx, nil
}

func (*SQLitePlaysetRepo) LoadState(dbPath string, idx domain.PlaysetIndex) (domain.LoadOrder, map[string]string, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return domain.LoadOrder{}, nil, fmt.Errorf("open sqlite db: %w", err)
	}
	defer db.Close()

	// Get playset name at index.
	names, err := listPlaysetNames(db)
	if err != nil {
		return domain.LoadOrder{}, nil, err
	}
	if idx < 0 || int(idx) >= len(names) {
		return domain.LoadOrder{}, nil, fmt.Errorf("playset index out of range: %d", idx)
	}
	playsetName := names[idx]

	// Find playset ID by name.
	var playsetID string
	err = db.QueryRow(`SELECT id FROM playsets WHERE name = ? AND isRemoved = 0`, playsetName).Scan(&playsetID)
	if err != nil {
		return domain.LoadOrder{}, nil, fmt.Errorf("find playset id: %w", err)
	}

	// JOIN to get ordered mods.
	rows, err := db.Query(`
		SELECT pm.modId, m.dirPath
		FROM playsets_mods pm
		JOIN mods m ON m.id = pm.modId
		WHERE pm.playsetId = ? AND pm.enabled = 1
		ORDER BY pm.position ASC
	`, playsetID)
	if err != nil {
		return domain.LoadOrder{}, nil, fmt.Errorf("query playset mods: %w", err)
	}
	defer rows.Close()

	orderedIDs := make([]string, 0)
	modPathByID := make(map[string]string)

	for rows.Next() {
		var modID, dirPath string
		if err := rows.Scan(&modID, &dirPath); err != nil {
			return domain.LoadOrder{}, nil, fmt.Errorf("scan mod row: %w", err)
		}
		orderedIDs = append(orderedIDs, modID)
		if dirPath != "" {
			modPathByID[modID] = dirPath
		}
	}

	return domain.LoadOrder{
		GameID:       domain.GameIDVic3,
		PlaysetIdx:   idx,
		ActiveModIDs: orderedIDs,
	}, modPathByID, nil
}

func (*SQLitePlaysetRepo) SaveState(dbPath string, idx domain.PlaysetIndex, order domain.LoadOrder, modPathByID map[string]string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get playset name at index.
	names, err := listPlaysetNames(db)
	if err != nil {
		return err
	}
	if idx < 0 || int(idx) >= len(names) {
		return fmt.Errorf("playset index out of range: %d", idx)
	}
	playsetName := names[idx]

	// Find playset ID by name.
	var playsetID string
	err = tx.QueryRow(`SELECT id FROM playsets WHERE name = ? AND isRemoved = 0`, playsetName).Scan(&playsetID)
	if err != nil {
		return fmt.Errorf("find playset id: %w", err)
	}

	// Clear existing entries.
	_, err = tx.Exec(`DELETE FROM playsets_mods WHERE playsetId = ?`, playsetID)
	if err != nil {
		return fmt.Errorf("delete existing playset mods: %w", err)
	}

	// Insert new entries.
	for position, modID := range order.ActiveModIDs {
		_, err = tx.Exec(`
			INSERT INTO playsets_mods (playsetId, modId, enabled, position)
			VALUES (?, ?, 1, ?)
		`, playsetID, modID, position)
		if err != nil {
			return fmt.Errorf("insert playset mod: %w", err)
		}
	}

	// Mark playset as custom and not removed.
	_, err = tx.Exec(`
		UPDATE playsets SET isRemoved = 0, loadOrder = 'custom' WHERE id = ?
	`, playsetID)
	if err != nil {
		return fmt.Errorf("update playset: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func listPlaysetNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT name FROM playsets WHERE isRemoved = 0 ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("query playset names: %w", err)
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan name: %w", err)
		}
		names = append(names, name)
	}
	return names, nil
}
