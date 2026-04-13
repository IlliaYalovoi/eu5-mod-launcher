# Task 01 — Go: Mod Scanner

## Goal

Implement the `internal/mods` package. It must scan a directory of installed mods, parse each mod's metadata file, and return a slice of structured `Mod` values.

## Context

Paradox games store mods as subdirectories. Each mod directory contains a descriptor file — for EU5 assume it is `metadata.json` in `.metadata`. Example can be found in `tasks/examples/mod_structure/.metadata`.

## Deliverables

### `internal/mods/mod.go`

Define the `Mod` struct:

```go
type Mod struct {
    ID          string   // directory name, used as stable identifier
    Name        string   // human-readable name from descriptor
    Version     string
    Tags        []string
    Description string
    ThumbnailPath string  // absolute path to thumbnail image, empty if none
    DirPath     string   // absolute path to mod directory
    Enabled     bool     // managed by loadorder package, default false
}
```

### `internal/mods/scanner.go`

```go
// ScanDir walks dirPath and returns one Mod per valid mod subdirectory.
// Errors reading individual mods are logged and skipped, not fatal.
func ScanDir(dirPath string) ([]Mod, error)
```

### `internal/mods/descriptor.go`

```go
// ParseDescriptor reads a descriptor.mod file and fills mod metadata fields.
// Unknown keys are silently ignored.
func ParseDescriptor(path string) (name, version, description string, tags []string, err error)
```

## Acceptance criteria

- `ScanDir` on a directory with 3 mock mod subdirectories returns 3 `Mod` values with correct metadata.
- A subdirectory with no `descriptor.mod` is skipped without error.
- All fields default gracefully if descriptor key is missing (empty string / empty slice).
- Write a `_test.go` file with at least one table-driven test using a temp directory.

## Notes for agent

- Do not import anything outside stdlib for this task.
- Do not add Wails imports — this package must be pure Go.
- The `Enabled` field is intentionally not set here — the load order package owns that.
