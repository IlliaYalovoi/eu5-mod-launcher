# Task 02 — Go: Load Order Store

## Goal

Implement `internal/loadorder` package. This package owns the persistent state of which mods are enabled and in what order. It reads and writes a JSON file stored in the OS user config directory.

## Context

This is pure data persistence — no UI, no Wails. The load order is an ordered list of mod IDs (strings). Enabled = present in the list. Position in the list = load index.

## Deliverables

### `internal/loadorder/store.go`

```go
type Store struct {
    // unexported fields
}

type State struct {
    OrderedIDs []string `json:"ordered_ids"` // enabled mods in load order
}

// New opens (or creates) the store at the given config file path.
func New(configPath string) (*Store, error)

// Load reads current state from disk.
func (s *Store) Load() (State, error)

// Save writes state to disk atomically (write to temp file, rename).
func (s *Store) Save(state State) error

// ConfigPath returns the resolved absolute path used by this store.
func (s *Store) ConfigPath() string
```

### `internal/loadorder/paths.go`

```go
// DefaultConfigPath returns the platform-appropriate path for the config file.
// Windows: %APPDATA%\EU5ModLauncher\loadorder.json
// Linux:   $XDG_CONFIG_HOME/eu5-mod-launcher/loadorder.json
//          (falls back to $HOME/.config/... if XDG not set)
func DefaultConfigPath() (string, error)
```

## Acceptance criteria

- Round-trip test: `Save` then `Load` returns identical state.
- Atomic write: if the process crashes mid-write, the previous file is not corrupted (use `os.Rename` after writing to a `.tmp`).
- `New` with a path whose parent directory doesn't exist creates the directory.
- `DefaultConfigPath` returns a non-empty path on both Windows and Linux (test with build tags or just verify logic by inspection).

## Notes for agent

- stdlib only: `os`, `encoding/json`, `path/filepath`.
- No Wails imports.
- The file format is internal — it does not need to match any game format.
