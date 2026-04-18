package legacy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"eu5-mod-launcher/internal/game"
	"eu5-mod-launcher/internal/utils"

	"eu5-mod-launcher/internal/logging"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type SqliteAdapter struct {
	id          string
	displayName string
	steamAppID  string

	db     *sqlx.DB
	dbPath string
	mu     sync.Mutex
}

func NewSqliteAdapter(id, displayName, steamAppID string) *SqliteAdapter {
	return &SqliteAdapter{
		id:          id,
		displayName: displayName,
		steamAppID:  steamAppID,
	}
}

func (s *SqliteAdapter) ID() string {
	return s.id
}

func (s *SqliteAdapter) DetectInstances() ([]game.Instance, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	userConfigPath := filepath.Join(home, "Documents", "Paradox Interactive", s.displayName)
	libRoots := utils.DiscoverSteamLibraryRoots()

	var installPath string
	var exePath string

	for _, root := range libRoots {
		candidate := filepath.Join(root, "steamapps", "common", s.displayName)
		candidateExe := filepath.Join(candidate, "binaries", s.id+".exe")
		if !utils.FileExists(candidateExe) {
			candidateExe = filepath.Join(candidate, s.id+".exe")
		}
		if utils.FileExists(candidateExe) {
			installPath = candidate
			exePath = candidateExe
			break
		}
	}

	return []game.Instance{
		{
			GameID:          s.id,
			InstallPath:     installPath,
			UserConfigPath:  userConfigPath,
			LocalModsDir:    filepath.Join(userConfigPath, "mod"),
			WorkshopModDirs: utils.DiscoverWorkshopModDirs(s.steamAppID),
			GameExePath:     exePath,
		},
	}, nil
}

func (s *SqliteAdapter) DetectVersion(inst game.Instance, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	var primaryFile string
	switch s.id {
	case "ck3":
		primaryFile = "titus_branch.txt"
	case "eu4":
		primaryFile = "eu4branch.txt"
	case "victoria3":
		primaryFile = "caligula_branch.txt"
	case "hoi4":
		// HOI4 files contain "None", handled below or just override
		primaryFile = "ho4branch.txt"
	default:
		primaryFile = s.id + "_branch.txt"
	}

	for _, filename := range []string{primaryFile, "clausewitz_branch.txt"} {
		content, err := os.ReadFile(filepath.Join(inst.InstallPath, filename))
		if err == nil {
			str := strings.TrimSpace(string(content))
			if str != "None" {
				return utils.ExtractVersion(str), nil
			}
		}
	}
	return "unknown", nil
}

func (s *SqliteAdapter) getDB(instance game.Instance) (*sqlx.DB, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dbPath := filepath.Join(instance.UserConfigPath, "launcher-v2.sqlite")
	if !utils.FileExists(dbPath) {
		return nil, fmt.Errorf("database not found at %s", dbPath)
	}

	if s.db != nil && s.dbPath == dbPath {
		return s.db, nil
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logging.Warnf("failed to close database at %s, err: %e", dbPath, err)
		}

	}

	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	s.db = db
	s.dbPath = dbPath
	return s.db, nil
}

type dbPlayset struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

type dbModEntry struct {
	ModID    string `db:"modId"`
	Enabled  bool   `db:"enabled"`
	Position int    `db:"position"`
}

type dbMod struct {
	ID          string `db:"id"`
	DisplayName string `db:"displayName"`
	DirPath     string `db:"dirPath"`
}

func (s *SqliteAdapter) LoadPlaysets(instance game.Instance) ([]game.Playset, error) {
	db, err := s.getDB(instance)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var dbPlaysets []dbPlayset
	err = db.Select(&dbPlaysets, "SELECT id, name FROM playsets WHERE isRemoved = 0")
	if err != nil {
		err = db.Select(&dbPlaysets, "SELECT id, name FROM playsets")
		if err != nil {
			return nil, fmt.Errorf("failed to query playsets: %w", err)
		}
	}

	playsets := make([]game.Playset, 0, len(dbPlaysets))
	for _, dp := range dbPlaysets {
		var entries []dbModEntry
		err = db.Select(&entries, `
			SELECT modId, enabled, position
			FROM playsets_mods
			WHERE playsetId = ?
			ORDER BY position ASC`, dp.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query playset mods: %w", err)
		}

		gameEntries := make([]game.ModEntry, len(entries))
		for i, e := range entries {
			gameEntries[i] = game.ModEntry{
				ID:       e.ModID,
				Enabled:  e.Enabled,
				Position: e.Position,
			}
		}

		playsets = append(playsets, game.Playset{
			ID:      dp.ID,
			Name:    dp.Name,
			Entries: gameEntries,
		})
	}

	return playsets, nil
}

func (s *SqliteAdapter) LoadMods(instance game.Instance) ([]game.ModEntry, error) {
	db, err := s.getDB(instance)
	if err != nil {
		return nil, nil
	}

	var dbMods []dbMod
	err = db.Select(&dbMods, "SELECT id, displayName, dirPath FROM mods WHERE isRemoved = 0")
	if err != nil {
		err = db.Select(&dbMods, "SELECT id, displayName, dirPath FROM mods")
		if err != nil {
			return nil, fmt.Errorf("failed to query mods: %w", err)
		}
	}

	entries := make([]game.ModEntry, len(dbMods))
	for i, m := range dbMods {
		entries[i] = game.ModEntry{
			ID:   m.ID,
			Path: m.DirPath,
		}
	}
	return entries, nil
}

// GetModNames returns a map of mod ID to display name.
func (s *SqliteAdapter) GetModNames(instance game.Instance) (map[string]string, error) {
	db, err := s.getDB(instance)
	if err != nil {
		return nil, err
	}
	var dbMods []dbMod
	err = db.Select(&dbMods, "SELECT id, displayName FROM mods WHERE isRemoved = 0")
	if err != nil {
		err = db.Select(&dbMods, "SELECT id, displayName FROM mods")
		if err != nil {
			return nil, err
		}
	}
	m := make(map[string]string)
	for _, dm := range dbMods {
		m[dm.ID] = dm.DisplayName
	}
	return m, nil
}

func (s *SqliteAdapter) SavePlayset(inst game.Instance, p game.Playset) error {
	db, err := s.getDB(inst)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM playsets_mods WHERE playsetId = ?", p.ID)
	if err != nil {
		return fmt.Errorf("failed to clear playset mods: %w", err)
	}

	for _, entry := range p.Entries {
		_, err = tx.Exec(`
			INSERT INTO playsets_mods (playsetId, modId, enabled, position)
			VALUES (?, ?, ?, ?)`,
			p.ID, entry.ID, entry.Enabled, entry.Position)
		if err != nil {
			return fmt.Errorf("failed to insert mod entry: %w", err)
		}
	}

	return tx.Commit()
}
