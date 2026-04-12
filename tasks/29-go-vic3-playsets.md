# 32 — Vic3 SQLite Playset Repository

## Context

Victoria 3 uses Paradox Launcher v2, which stores playsets (mod lists) in a local SQLite database (`launcher-v2.sqlite`) instead of a JSON file like EU5. The launcher database lives at:
```
%USERPROFILE%\Documents\Paradox Interactive\Victoria 3\launcher-v2.sqlite
```

## Schema (from example DB)

### `playsets` table
| Column | Type | Notes |
|--------|------|-------|
| id | char(36) | UUID primary key |
| name | varchar(255) | Display name |
| isActive | boolean | Whether this playset was last used |
| loadOrder | varchar(255) | "custom" for user-created |
| isRemoved | boolean | Soft delete flag |

### `playsets_mods` table (junction)
| Column | Type | Notes |
|--------|------|-------|
| playsetId | char(36) | FK to playsets.id |
| modId | char(36) | FK to mods.id |
| enabled | boolean | Whether mod is active in this playset |
| position | integer | Sort order within playset |

### `mods` table
| Column | Type | Notes |
|--------|------|-------|
| id | char(36) | UUID primary key |
| steamId | varchar(255) | Steam workshop ID |
| name / displayName | varchar(255) | Human-readable name |
| version | varchar(255) | Mod version string |
| dirPath | varchar(255) | Local path to mod directory |
| status | varchar(255) | "ready_to_play", "unsubscribed", etc. |
| tags | json | Array of version constraints like `["1.12"]` |

## What's needed

### 1. New `vic3` package — `internal/game/vic3/`

Create `internal/game/vic3/` with:

**`sqlite_playset_repo.go`** — implements `repo.PlaysetRepo`:
```go
type SQLitePlaysetRepo struct{}

func (*SQLitePlaysetRepo) ListPlaysets(dbPath string) ([]string, domain.PlaysetIndex, error)
func (*SQLitePlaysetRepo) LoadState(dbPath string, idx domain.PlaysetIndex) (domain.LoadOrder, map[string]string, error)
func (*SQLitePlaysetRepo) SaveState(dbPath string, idx domain.PlaysetIndex, order domain.LoadOrder, modPathByID map[string]string) error
```

- `ListPlaysets`: SELECT non-removed playsets ordered by name, return names list and index of the one with `isActive=1` (or 0 if none).
- `LoadState`: JOIN playsets_mods + mods on the selected playset, return ordered IDs and `modPathByID` map (modID→dirPath).
- `SaveState`: Use a transaction — clear playsets_mods entries for the playset, then INSERT new entries with position. Mark the playset as `isRemoved=0` and `loadOrder='custom'`.

**`vic3_adapter.go`** — wraps SQLitePlaysetRepo as a game.Adapter:
```go
type Vic3Adapter struct {
    playsets *SQLitePlaysetRepo
}

func (*Vic3Adapter) GameID() domain.GameID { return domain.GameIDVic3 }
func (*Vic3Adapter) Descriptor() domain.GameDescriptor { ... }
func (*Vic3Adapter) DiscoverPaths() (domain.GamePaths, error)
func (a *Vic3Adapter) PlaysetRepo() repo.PlaysetRepo { return a.playsets }
```

- `DiscoverPaths`: Standard Paradox paths under `%USERPROFILE%\Documents\Paradox Interactive\Victoria 3\`. PlaysetsPath = `launcher-v2.sqlite` path. LocalModsDir = `./mod`. Workshop = Steam workshop content dir for app ID 529340.

### 2. Register Vic3Adapter in GameService

In `NewGameService` or wherever adapters are wired, add Vic3Adapter alongside EU5Adapter.

### 3. Detection stub for Vic3 (can reuse existing game detection)

Game detection at `internal/game/detection.go` should already handle Vic3 once it exists. Ensure it returns `Detected: true` when `launcher-v2.sqlite` is found.

## Dependencies

- `modernc.org/sqlite` — pure-Go SQLite driver (check if already in go.mod).

## Out of scope

- Launching Vic3 via Paradox Launcher (launches via game exe like EU5).
- Any Steam workshop subscription management for Vic3.
