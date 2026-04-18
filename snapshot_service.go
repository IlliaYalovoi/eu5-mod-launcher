package main

import (
	"fmt"
	"time"
)

func (a *App) nextSnapshotRevision(gameID string) int64 {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	next := a.snapshotRevisionMap[gameID] + 1
	a.snapshotRevisionMap[gameID] = next
	return next
}

func (a *App) cacheSnapshot(snapshot GameSnapshot) {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	a.snapshotCache[snapshot.GameID] = snapshot
}

func (a *App) getCachedSnapshot(gameID string) (GameSnapshot, bool) {
	a.snapshotCacheMu.RLock()
	defer a.snapshotCacheMu.RUnlock()
	snapshot, ok := a.snapshotCache[gameID]
	return snapshot, ok
}

func (a *App) markSnapshotStale(gameID string) {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	if snapshot, ok := a.snapshotCache[gameID]; ok {
		snapshot.Meta.Stale = true
		a.snapshotCache[gameID] = snapshot
	}
}

func (a *App) invalidateSnapshot(gameID string) {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	delete(a.snapshotCache, gameID)
	a.snapshotRevisionMap[gameID] = a.snapshotRevisionMap[gameID] + 1
}

func (a *App) invalidateActiveSnapshot() {
	active := a.GetActiveGameID()
	if active == "" {
		return
	}
	a.invalidateSnapshot(active)
}

func (a *App) snapshotSettingsFromCurrent() SnapshotSettings {
	return SnapshotSettings{
		ModsDirStatus:       a.GetModsDirStatus(),
		GameExe:             a.GetGameExe(),
		AutoDetectedGameExe: a.GetAutoDetectedGameExe(),
		ConfigPath:          a.GetConfigPath(),
		GameVersion:         a.GetGameVersion(),
		GameVersionOverride: a.GetGameVersionOverride(),
		AvailableGames:      a.GetAvailableGames(),
	}
}

func (a *App) buildSnapshotFromCurrentContextLocked() (GameSnapshot, error) {
	modsList, err := a.GetAllMods()
	if err != nil {
		return GameSnapshot{}, err
	}

	gameID := a.GetActiveGameID()
	if gameID == "" {
		return GameSnapshot{}, fmt.Errorf("build snapshot: active game is empty")
	}

	revision := a.nextSnapshotRevision(gameID)
	now := time.Now().UnixMilli()
	constraints := a.GetConstraints()
	layout := a.GetLauncherLayout()
	order := a.GetLoadOrder()
	playsets := a.GetPlaysetNames()

	snapshot := GameSnapshot{
		GameID:                     gameID,
		Mods:                       modsList,
		LoadOrder:                  append([]string(nil), order...),
		LauncherLayout:             layout,
		Constraints:                constraints,
		PlaysetNames:               append([]string(nil), playsets...),
		GameActivePlaysetIndex:     a.GetGameActivePlaysetIndex(),
		LauncherActivePlaysetIndex: a.GetLauncherActivePlaysetIndex(),
		Settings:                   a.snapshotSettingsFromCurrent(),
		Meta: SnapshotMeta{
			Revision:  revision,
			FetchedAt: now,
			Stale:     false,
		},
	}

	return snapshot, nil
}

func (a *App) withGameContextLocked(gameID string, fn func() (GameSnapshot, error)) (GameSnapshot, error) {
	current := a.GetActiveGameID()
	if gameID == "" {
		gameID = current
	}
	if gameID == current {
		return fn()
	}

	if err := a.setActiveGameInternal(gameID, false); err != nil {
		return GameSnapshot{}, err
	}

	snapshot, err := fn()
	restoreErr := a.setActiveGameInternal(current, false)
	if err != nil {
		if restoreErr != nil {
			return GameSnapshot{}, fmt.Errorf("build snapshot for %q: %w (restore active game %q failed: %v)", gameID, err, current, restoreErr)
		}
		return GameSnapshot{}, err
	}
	if restoreErr != nil {
		return GameSnapshot{}, fmt.Errorf("restore active game %q after snapshot for %q: %w", current, gameID, restoreErr)
	}

	return snapshot, nil
}

func (a *App) GetGameSnapshot(gameID string) (GameSnapshot, error) {
	target := gameID
	if target == "" {
		target = a.GetActiveGameID()
	}
	if target == "" {
		return GameSnapshot{}, fmt.Errorf("get snapshot: target game is empty")
	}

	if cached, ok := a.getCachedSnapshot(target); ok && !cached.Meta.Stale {
		return cached, nil
	}

	a.snapshotBuildMu.Lock()
	defer a.snapshotBuildMu.Unlock()

	staleSnapshot, hasStaleSnapshot := a.getCachedSnapshot(target)
	snapshot, err := a.withGameContextLocked(target, func() (GameSnapshot, error) {
		return a.buildSnapshotFromCurrentContextLocked()
	})
	if err != nil {
		if hasStaleSnapshot {
			staleSnapshot.Meta.Stale = true
			a.cacheSnapshot(staleSnapshot)
			return staleSnapshot, nil
		}
		return GameSnapshot{}, err
	}

	a.cacheSnapshot(snapshot)
	return snapshot, nil
}

func (a *App) SetActiveGameAndGetSnapshot(gameID string) (GameSnapshot, error) {
	a.snapshotBuildMu.Lock()
	defer a.snapshotBuildMu.Unlock()

	previous := a.GetActiveGameID()
	target := gameID
	if target == "" {
		target = previous
	}
	if target == "" {
		return GameSnapshot{}, fmt.Errorf("set active game and get snapshot: target game is empty")
	}

	switched := target != previous
	if switched {
		if err := a.setActiveGameInternal(target, false); err != nil {
			return GameSnapshot{}, err
		}
	}

	snapshot, err := a.buildSnapshotFromCurrentContextLocked()
	if err != nil {
		if switched {
			if restoreErr := a.setActiveGameInternal(previous, false); restoreErr != nil {
				return GameSnapshot{}, fmt.Errorf("build snapshot for switched game %q: %w (restore previous game %q failed: %v)", target, err, previous, restoreErr)
			}
		}
		return GameSnapshot{}, err
	}

	a.cacheSnapshot(snapshot)
	return snapshot, nil
}

func (a *App) WarmNonActiveGameSnapshots() (map[string]GameSnapshot, error) {
	active := a.GetActiveGameID()
	games := a.GetAvailableGames()
	out := make(map[string]GameSnapshot, len(games))

	for _, gameID := range games {
		if gameID == active {
			continue
		}

		a.snapshotBuildMu.Lock()
		snapshot, err := a.withGameContextLocked(gameID, func() (GameSnapshot, error) {
			return a.buildSnapshotFromCurrentContextLocked()
		})
		a.snapshotBuildMu.Unlock()
		if err != nil {
			a.markSnapshotStale(gameID)
			continue
		}

		a.cacheSnapshot(snapshot)
		out[gameID] = snapshot
	}

	return out, nil
}
