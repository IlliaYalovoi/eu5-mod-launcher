# Legacy SQLite Adapter Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix legacy SQLite playsets not mapping mod names correctly and fix enable/disable state not persisting properly.

**Architecture:** Wire up `SqlitePlaysetRepository.SaveState` to the adapter's `SavePlayset`. Use `LoadMods` to map UUIDs from the database to standard directory-based identifiers (`steam_mod_id` or folder name) to align SQLite with the rest of the application. Map `OrderedIDs` back to UUIDs on save, and delete disabled mods from the playset configuration.

**Tech Stack:** Go, standard library.

---

### Task 1: Update LegacyAdapter Interface

**Files:**
- Modify: `internal/repo/playset_repo.go:37-43`

- [ ] **Step 1: Add new methods to LegacyAdapter**

Update `LegacyAdapter` interface in `internal/repo/playset_repo.go` to match `SqliteAdapter` methods needed for tracking mods and saving playsets.

```go
type LegacyAdapter interface {
	LoadPlaysets(inst game.Instance) ([]game.Playset, error)
	GetModNames(inst game.Instance) (map[string]string, error)
	LoadMods(inst game.Instance) ([]game.ModEntry, error)
	SavePlayset(inst game.Instance, p game.Playset) error
	ID() string
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/repo/playset_repo.go
git commit -m "fix(legacy): add LoadMods and SavePlayset to adapter interface"
```

### Task 2: Fix LoadState for SQLite Repository

**Files:**
- Modify: `internal/repo/playset_repo.go:65-82`

- [ ] **Step 1: Update LoadState to map UUIDs to stable identifiers and filter by enabled status**

Rewrite `SqlitePlaysetRepository.LoadState` to fetch mod directory paths, extract the standard identifier (`steam_mod_id` or dir name) from them using `filepath.Base`, and only return enabled entries. This aligns SQLite's internal UUIDs with the common ID format expected by `App.GetAllMods()`.

```go
func (r *SqlitePlaysetRepository) LoadState(path string, index int) (loadorder.State, map[string]string, error) {
	playsets, err := r.adapter.LoadPlaysets(r.inst)
	if err != nil {
		return loadorder.State{}, nil, err
	}
	if index < 0 || index >= len(playsets) {
		return loadorder.State{}, nil, nil
	}

	mods, err := r.adapter.LoadMods(r.inst)
	if err != nil {
		return loadorder.State{}, nil, err
	}

	uuidToSteamID := make(map[string]string)
	steamIDToPath := make(map[string]string)

	for _, m := range mods {
		norm := filepath.Clean(filepath.ToSlash(m.Path))
		steamID := filepath.Base(norm)
		if steamID == "" || steamID == "." {
			continue
		}
		uuidToSteamID[m.ID] = steamID
		steamIDToPath[steamID] = m.Path
	}

	p := playsets[index]
	ids := make([]string, 0, len(p.Entries))
	for _, e := range p.Entries {
		if !e.Enabled {
			continue
		}
		steamID, ok := uuidToSteamID[e.ID]
		if !ok {
			steamID = e.ID
		}
		ids = append(ids, steamID)
	}
	
	return loadorder.State{OrderedIDs: ids}, steamIDToPath, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/repo/playset_repo.go
git commit -m "fix(legacy): properly map UUIDs to mod IDs in load state"
```

### Task 3: Implement SaveState for SQLite Repository

**Files:**
- Modify: `internal/repo/playset_repo.go:83-85`

- [ ] **Step 1: Wire up SaveState to write load orders back to SQLite using UUIDs**

Implement `SqlitePlaysetRepository.SaveState` so it converts the application's standard IDs back to the database's internal UUIDs using the `modPathByID` provided by the application and the `mods` mapping. As requested, this will drop all disabled mods from the playset payload because `state.OrderedIDs` only contains enabled mods.

```go
func (r *SqlitePlaysetRepository) SaveState(path string, index int, state loadorder.State, modPathByID map[string]string) error {
	playsets, err := r.adapter.LoadPlaysets(r.inst)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(playsets) {
		return nil
	}

	mods, err := r.adapter.LoadMods(r.inst)
	if err != nil {
		return err
	}

	pathToUUID := make(map[string]string)
	for _, m := range mods {
		norm := filepath.Clean(filepath.ToSlash(m.Path))
		pathToUUID[norm] = m.ID
	}

	p := playsets[index]
	newEntries := make([]game.ModEntry, 0, len(state.OrderedIDs))

	for i, steamID := range state.OrderedIDs {
		pathValue := modPathByID[steamID]
		norm := filepath.Clean(filepath.ToSlash(pathValue))

		uuid, ok := pathToUUID[norm]
		if !ok {
			for _, m := range mods {
				mNorm := filepath.Clean(filepath.ToSlash(m.Path))
				if filepath.Base(mNorm) == steamID {
					uuid = m.ID
					break
				}
			}
			if uuid == "" {
				continue
			}
		}

		newEntries = append(newEntries, game.ModEntry{
			ID:       uuid,
			Enabled:  true,
			Position: i,
		})
	}

	p.Entries = newEntries
	return r.adapter.SavePlayset(r.inst, p)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/repo/playset_repo.go
git commit -m "fix(legacy): implement sqlite playset persistence"
```
