# 33 — Mod/Game Version Compatibility Check

## Context

Mods may declare `supported_game_version` in their descriptor (EU5) or `tags` array containing version constraints (Vic3). The launcher must warn when a mod is incompatible with the current game version.

### EU5 version file

`caesar_branch.txt` at game install root contains:
```
release/1.1.0
```
Format: `release/X.Y.Z` — major.minor.patch.

### EU5 mod descriptor

`descriptor.mod` line:
```
supported_version="1.1.0"
```
Can also be glob patterns like `1.*` or `1.12.*`.

### Vic3 mod tags

From example DB, `tags` is a JSON array like `["1.12"]`. Version format TBD — use stub parsing that treats any non-empty tag as a constraint (e.g., `"1.12"` means compatible with game version 1.12.x).

## What's needed

### 1. Add fields to `domain.Mod`

```go
type Mod struct {
    // existing fields...
    SupportedVersions []string  // e.g., ["1.1.0", "1.*"]
    GameVersionMismatch bool    // set by caller after comparison
}
```

### 2. Version parsing in `internal/domain/`

**`version.go`**:
```go
// ParseGameVersion parses "release/1.2.3" or "1.2.3" into (1, 2, 3).
func ParseGameVersion(s string) (major, minor, patch int, err error)

// MatchesConstraint checks if game version (major, minor, patch) satisfies
// a constraint like "1.*", "1.2.*", or "1.2.3".
func MatchesConstraint(major, minor, patch int, constraint string) bool
```

Constraint rules:
- `"1.*"` → matches any 1.x.y
- `"1.2.*"` → matches any 1.2.y
- `"1.2.3"` → matches only 1.2.3
- `"*" ` → matches anything (wildcard)

### 3. Game version provider interface

In `internal/game/adapter.go`, extend `Adapter`:
```go
type Adapter interface {
    // existing methods...
    GameVersion(installDir string) (string, error)  // e.g., "1.1.0"
}
```

- EU5: Read `caesar_branch.txt` from install dir, parse version.
- Vic3: Stub — return `"1.0.0"` for now (no known version file location).

### 4. Compatibility check in `App.GetAllMods`

After scanning mods, before returning:

```go
gameVersion, _ := a.game.GameVersion(a.gamePaths.GameExePath)
major, minor, patch, _ := domain.ParseGameVersion(gameVersion)

for i := range allMods {
    allMods[i].GameVersionMismatch = !domain.IsModCompatible(allMods[i].SupportedVersions, major, minor, patch)
}
```

`IsModCompatible` returns `true` if the constraint list is empty (no constraints = always compatible) or if any constraint matches.

### 5. Frontend warning indicator

In the mod list component (where mod name is displayed), add a visual warning badge/icon when `GameVersionMismatch` is `true`.

Implementation depends on existing component structure — likely a small warning icon (⚠ or SVG) next to the mod name, styled in amber/yellow. The badge should have a tooltip: "Incompatible with current game version".

## Out of scope

- Enforcing version constraints at launch time (warn only).
- Vic3 version file discovery (stub only).
