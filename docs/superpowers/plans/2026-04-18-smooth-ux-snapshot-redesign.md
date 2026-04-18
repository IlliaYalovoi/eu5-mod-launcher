# Smooth UX Snapshot Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace multi-call frontend refresh flow with backend snapshot APIs + snapshot-driven UI state so game switching, mod toggles, and theme transitions feel continuous (no blank/jerky states).

**Architecture:** Add backend `GameSnapshot` aggregation endpoints and per-game snapshot cache/revisioning. Frontend consumes snapshots via one orchestrator store that controls active/visible state, crossfade timing, background warming, and stale-while-revalidate behavior. Existing domain stores become thin wrappers so current component structure can be migrated without one-shot rewrite.

**Tech Stack:** Go (Wails backend), Vue 3 Composition API, Pinia, TypeScript, CSS transitions.

---

## Implementation constraints (must follow)

- Backend remains source of truth.
- For legacy sqlite games use `launcher-v2.sqlite` as primary DB path.
- Do not add new tests (project rule). Verification = build/typecheck/manual checks.
- Required verification before completion:
  - `go build ./...`
  - `go vet ./...` (interfaces change in this plan, so run it)
  - `cd frontend && npx tsc --noEmit`

---

## File structure and ownership map

### Backend (Go)

- **Create:** `snapshot_models.go`
  - Wails-exported DTOs (`GameSnapshot`, nested settings/meta structs).
- **Create:** `snapshot_service.go`
  - Snapshot builder, cache/revision helpers, context-safe snapshot fetch methods.
- **Modify:** `app.go`
  - Add snapshot fields to `App`, initialize maps, expose new Wails methods, invalidate cache from mutators.

### Frontend (TypeScript/Vue)

- **Create:** `frontend/src/stores/snapshots.ts`
  - Snapshot orchestrator (bootstrap, switch, warm, refresh, request tokens, transition flags).
- **Modify:** `frontend/src/types.ts`
  - Snapshot TS interfaces.
- **Modify:** `frontend/src/stores/settings.ts`
  - Delegate game switching/fetch flow to snapshot store.
- **Modify:** `frontend/src/stores/mods.ts`
  - Use snapshot-backed mod list; keep metadata/unsubscribe behavior.
- **Modify:** `frontend/src/stores/loadorder.ts`
  - Use snapshot-backed load order/layout/playsets.
- **Modify:** `frontend/src/stores/constraints.ts`
  - Use snapshot-backed constraints.
- **Modify:** `frontend/src/main.ts`
  - Remove old multi-store bootstrap; start snapshot bootstrap.
- **Modify:** `frontend/src/App.vue`
  - Add snapshot-driven crossfade + startup skeleton gating + theme-transition class.
- **Modify:** `frontend/src/components/Sidebar.vue`
  - Use snapshot switch state and avoid direct hard refetches.
- **Modify:** `frontend/src/components/LoadOrderPanel.vue`
  - TransitionGroup for load-order rows and optional skeleton state.
- **Modify:** `frontend/src/components/ModListPanel.vue`
  - TransitionGroup for repository cards and optional skeleton state.
- **Modify:** `frontend/src/assets/main.css`
  - Global color/text transition contract + reduced-motion fallback + crossfade classes.
- **Modify (generated):** `frontend/wailsjs/go/main/App.d.ts`
- **Modify (generated):** `frontend/wailsjs/go/main/App.js`
- **Modify (generated):** `frontend/wailsjs/go/models.ts`

---

## Task 1: Add backend snapshot DTOs and cache fields

**Files:**
- Create: `snapshot_models.go`
- Modify: `app.go`

**Test:**
- Command-only verification (`go build ./...`), no new tests.

- [ ] **Step 1: Create snapshot DTO file**

```go
// snapshot_models.go
package main

import (
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/mods"
)

type SnapshotMeta struct {
	Revision  int64 `json:"revision"`
	FetchedAt int64 `json:"fetchedAt"`
	Stale     bool  `json:"stale"`
}

type SnapshotSettings struct {
	ModsDirStatus       ModsDirStatus `json:"modsDirStatus"`
	GameExe             string        `json:"gameExe"`
	AutoDetectedGameExe string        `json:"autoDetectedGameExe"`
	ConfigPath          string        `json:"configPath"`
	GameVersion         string        `json:"gameVersion"`
	GameVersionOverride string        `json:"gameVersionOverride"`
	AvailableGames      []string      `json:"availableGames"`
}

type GameSnapshot struct {
	GameID                string            `json:"gameID"`
	Mods                  []mods.Mod        `json:"mods"`
	LoadOrder             []string          `json:"loadOrder"`
	LauncherLayout        LauncherLayout    `json:"launcherLayout"`
	Constraints           []graph.Constraint `json:"constraints"`
	PlaysetNames          []string          `json:"playsetNames"`
	GameActivePlaysetIdx  int               `json:"gameActivePlaysetIndex"`
	LauncherActivePlayset int               `json:"launcherActivePlaysetIndex"`
	Settings              SnapshotSettings  `json:"settings"`
	Meta                  SnapshotMeta      `json:"meta"`
}
```

- [ ] **Step 2: Extend `App` struct with snapshot state**

```go
// app.go (inside type App struct)
	snapshotBuildMu     sync.Mutex
	snapshotCacheMu     sync.RWMutex
	snapshotCache       map[string]GameSnapshot
	snapshotRevisionMap map[string]int64
```

- [ ] **Step 3: Initialize snapshot fields in `NewApp()`**

```go
// app.go (inside NewApp)
app := &App{
	loState:             loadorder.State{OrderedIDs: []string{}},
	conGraph:            graph.New(),
	modPathByID:         map[string]string{},
	launcherLayout:      LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}},
	imageDataURLs:       map[string]string{},
	playsetNames:        []string{},
	gameActiveIndex:     -1,
	launcherIndex:       -1,
	snapshotCache:       map[string]GameSnapshot{},
	snapshotRevisionMap: map[string]int64{},
}
```

- [ ] **Step 4: Build sanity check**

Run: `go build ./...`  
Expected: successful build.

- [ ] **Step 5: Commit Task 1**

```bash
git add snapshot_models.go app.go
git commit -m "feat(snapshot): add snapshot DTOs and app cache fields"
```

---

## Task 2: Implement snapshot builder + new backend APIs

**Files:**
- Create: `snapshot_service.go`
- Modify: `app.go`

**Test:**
- Command-only verification (`go build ./...`).

- [ ] **Step 1: Add revision/cache helpers**

```go
// snapshot_service.go
package main

import "time"

func (a *App) nextSnapshotRevision(gameID string) int64 {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	next := a.snapshotRevisionMap[gameID] + 1
	a.snapshotRevisionMap[gameID] = next
	return next
}

func (a *App) cacheSnapshot(s GameSnapshot) {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	a.snapshotCache[s.GameID] = s
}

func (a *App) markSnapshotStale(gameID string) {
	a.snapshotCacheMu.Lock()
	defer a.snapshotCacheMu.Unlock()
	if snap, ok := a.snapshotCache[gameID]; ok {
		snap.Meta.Stale = true
		a.snapshotCache[gameID] = snap
	}
}
```

- [ ] **Step 2: Add settings extraction helper**

```go
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
```

- [ ] **Step 3: Add builder for current active context**

```go
func (a *App) buildSnapshotFromCurrentContextLocked() (GameSnapshot, error) {
	modsList, err := a.GetAllMods()
	if err != nil {
		return GameSnapshot{}, err
	}
	gameID := a.GetActiveGameID()
	revision := a.nextSnapshotRevision(gameID)
	now := time.Now().UnixMilli()

	constraints := a.GetConstraints()
	layout := a.GetLauncherLayout()
	order := a.GetLoadOrder()
	playsets := a.GetPlaysetNames()

	snapshot := GameSnapshot{
		GameID:                gameID,
		Mods:                  modsList,
		LoadOrder:             append([]string(nil), order...),
		LauncherLayout:        layout,
		Constraints:           constraints,
		PlaysetNames:          append([]string(nil), playsets...),
		GameActivePlaysetIdx:  a.GetGameActivePlaysetIndex(),
		LauncherActivePlayset: a.GetLauncherActivePlaysetIndex(),
		Settings:              a.snapshotSettingsFromCurrent(),
		Meta: SnapshotMeta{
			Revision:  revision,
			FetchedAt: now,
			Stale:     false,
		},
	}
	return snapshot, nil
}
```

- [ ] **Step 4: Add active-game switch wrapper with guaranteed restore**

```go
func (a *App) withGameContextLocked(gameID string, fn func() (GameSnapshot, error)) (GameSnapshot, error) {
	current := a.GetActiveGameID()
	if gameID == "" {
		gameID = current
	}
	if gameID == current {
		return fn()
	}

	if err := a.SetActiveGame(gameID); err != nil {
		return GameSnapshot{}, err
	}
	defer func() {
		_ = a.SetActiveGame(current)
	}()

	return fn()
}
```

- [ ] **Step 5: Expose Wails method `GetGameSnapshot`**

```go
func (a *App) GetGameSnapshot(gameID string) (GameSnapshot, error) {
	a.snapshotBuildMu.Lock()
	defer a.snapshotBuildMu.Unlock()

	snapshot, err := a.withGameContextLocked(gameID, func() (GameSnapshot, error) {
		return a.buildSnapshotFromCurrentContextLocked()
	})
	if err != nil {
		return GameSnapshot{}, err
	}
	a.cacheSnapshot(snapshot)
	return snapshot, nil
}
```

- [ ] **Step 6: Expose Wails method `SetActiveGameAndGetSnapshot`**

```go
func (a *App) SetActiveGameAndGetSnapshot(gameID string) (GameSnapshot, error) {
	a.snapshotBuildMu.Lock()
	defer a.snapshotBuildMu.Unlock()

	if err := a.SetActiveGame(gameID); err != nil {
		return GameSnapshot{}, err
	}
	snapshot, err := a.buildSnapshotFromCurrentContextLocked()
	if err != nil {
		return GameSnapshot{}, err
	}
	a.cacheSnapshot(snapshot)
	return snapshot, nil
}
```

- [ ] **Step 7: Expose Wails method `WarmNonActiveGameSnapshots`**

```go
func (a *App) WarmNonActiveGameSnapshots() (map[string]GameSnapshot, error) {
	a.snapshotBuildMu.Lock()
	defer a.snapshotBuildMu.Unlock()

	active := a.GetActiveGameID()
	games := a.GetAvailableGames()
	out := make(map[string]GameSnapshot, len(games))

	for _, gameID := range games {
		if gameID == active {
			continue
		}
		snapshot, err := a.withGameContextLocked(gameID, func() (GameSnapshot, error) {
			return a.buildSnapshotFromCurrentContextLocked()
		})
		if err != nil {
			a.markSnapshotStale(gameID)
			continue
		}
		a.cacheSnapshot(snapshot)
		out[gameID] = snapshot
	}
	return out, nil
}
```

- [ ] **Step 8: Build sanity check**

Run: `go build ./...`  
Expected: successful build with new exported methods.

- [ ] **Step 9: Commit Task 2**

```bash
git add snapshot_service.go app.go
git commit -m "feat(snapshot): add aggregated snapshot APIs"
```

---

## Task 3: Invalidate snapshots after mutating backend operations

**Files:**
- Modify: `snapshot_service.go`
- Modify: `app.go`

**Test:**
- Command-only verification (`go build ./...`).

- [ ] **Step 1: Add invalidation helper methods**

```go
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
```

- [ ] **Step 2: Invalidate after load-order/layout mutations**

Insert this block immediately before successful `return nil` in each method:
- `SetLoadOrder`
- `SetLauncherLayout`
- `SaveCompiledLoadOrder`
- `Autosort`
- `SetLauncherActivePlaysetIndex`

```go
a.invalidateActiveSnapshot()
```

- [ ] **Step 3: Invalidate after constraint mutations**

Insert this block immediately before successful `return nil` in each method:
- `AddConstraint`
- `RemoveConstraint`
- `AddLoadFirst`
- `AddLoadLast`
- `RemoveLoadFirst`
- `RemoveLoadLast`

```go
a.invalidateActiveSnapshot()
```

- [ ] **Step 4: Invalidate after settings/path mutations**

Insert this block immediately before successful `return nil` in each method:
- `SetGameVersionOverride`
- `SetGameExe`
- `SetModsDir`

```go
a.invalidateActiveSnapshot()
```

- [ ] **Step 5: Invalidate switched target on `SetActiveGame`**

Insert this block at the end of `SetActiveGame`, after `refreshState()` path and before returning success:

```go
a.invalidateSnapshot(gameID)
```

- [ ] **Step 6: Build sanity check**

Run: `go build ./...`  
Expected: successful build.

- [ ] **Step 7: Commit Task 3**

```bash
git add app.go snapshot_service.go
git commit -m "fix(snapshot): invalidate cache after state mutations"
```

---

## Task 4: Update Wails frontend bindings + TypeScript snapshot contracts

**Files:**
- Modify (generated): `frontend/wailsjs/go/main/App.d.ts`
- Modify (generated): `frontend/wailsjs/go/main/App.js`
- Modify (generated): `frontend/wailsjs/go/models.ts`
- Modify: `frontend/src/types.ts`

**Test:**
- Command-only verification (`cd frontend && npx tsc --noEmit`).

- [ ] **Step 1: Regenerate Wails module bindings**

Run: `wails generate module`  
Expected: command exits successfully.

- [ ] **Step 2: Verify exported functions exist in `App.d.ts`**

Expected new signatures:

```ts
export function GetGameSnapshot(arg1:string):Promise<main.GameSnapshot>;
export function SetActiveGameAndGetSnapshot(arg1:string):Promise<main.GameSnapshot>;
export function WarmNonActiveGameSnapshots():Promise<Record<string, main.GameSnapshot>>;
```

- [ ] **Step 3: Verify wrapper functions exist in `App.js`**

```js
export function GetGameSnapshot(arg1) {
  return window['go']['main']['App']['GetGameSnapshot'](arg1);
}

export function SetActiveGameAndGetSnapshot(arg1) {
  return window['go']['main']['App']['SetActiveGameAndGetSnapshot'](arg1);
}

export function WarmNonActiveGameSnapshots() {
  return window['go']['main']['App']['WarmNonActiveGameSnapshots']();
}
```

- [ ] **Step 4: Add explicit app-local TS interfaces in `frontend/src/types.ts`**

```ts
export interface SnapshotMeta {
  revision: number
  fetchedAt: number
  stale: boolean
}

export interface SnapshotSettings {
  modsDirStatus: {
    effectiveDir: string
    autoDetectedDir: string
    customDir: string
    usingCustomDir: boolean
    autoDetectedExists: boolean
    effectiveExists: boolean
  }
  gameExe: string
  autoDetectedGameExe: string
  configPath: string
  gameVersion: string
  gameVersionOverride: string
  availableGames: string[]
}

export interface GameSnapshot {
  gameID: string
  mods: Mod[]
  loadOrder: string[]
  launcherLayout: LauncherLayout
  constraints: Constraint[]
  playsetNames: string[]
  gameActivePlaysetIndex: number
  launcherActivePlaysetIndex: number
  settings: SnapshotSettings
  meta: SnapshotMeta
}
```

- [ ] **Step 5: Typecheck frontend**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

- [ ] **Step 6: Commit Task 4**

```bash
git add frontend/wailsjs/go/main/App.d.ts frontend/wailsjs/go/main/App.js frontend/wailsjs/go/models.ts frontend/src/types.ts
git commit -m "chore(bindings): expose snapshot APIs to frontend"
```

---

## Task 5: Add snapshot orchestrator store

**Files:**
- Create: `frontend/src/stores/snapshots.ts`

**Test:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Create store skeleton and state model**

```ts
// frontend/src/stores/snapshots.ts
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { GameSnapshot } from '../types'
import {
  GetActiveGameID,
  GetGameSnapshot,
  SetActiveGameAndGetSnapshot,
  WarmNonActiveGameSnapshots,
} from '../../wailsjs/go/main/App'

type StartupState = 'cold' | 'warming' | 'ready'
type SwitchState = 'idle' | 'loading' | 'committing'

export const useSnapshotsStore = defineStore('snapshots', () => {
  const snapshotsByGameID = ref<Record<string, GameSnapshot>>({})
  const activeGameID = ref('')
  const visibleGameID = ref('')
  const startupState = ref<StartupState>('cold')
  const switchState = ref<SwitchState>('idle')
  const latestRequestID = ref(0)
  const transitionNonce = ref(0)
  const warmTimer = ref<number | null>(null)
```

- [ ] **Step 2: Add safe apply helper with revision guard**

```ts
  function applySnapshot(next: GameSnapshot): void {
    const current = snapshotsByGameID.value[next.gameID]
    if (current && next.meta.revision < current.meta.revision) {
      return
    }
    snapshotsByGameID.value = {
      ...snapshotsByGameID.value,
      [next.gameID]: next,
    }
  }
```

- [ ] **Step 3: Add bootstrap action (active snapshot first)**

```ts
  async function bootstrap(): Promise<void> {
    startupState.value = 'cold'
    const active = await GetActiveGameID()
    const snapshot = (await GetGameSnapshot(active || '')) as GameSnapshot
    applySnapshot(snapshot)
    activeGameID.value = snapshot.gameID
    visibleGameID.value = snapshot.gameID
    startupState.value = 'warming'
    void warmNonActive()
    startWarmLoop()
    startupState.value = 'ready'
  }
```

- [ ] **Step 4: Add switch action with request token and no blank frame**

```ts
  async function switchGame(gameID: string): Promise<void> {
    if (!gameID || switchState.value !== 'idle') {
      return
    }

    const requestID = latestRequestID.value + 1
    latestRequestID.value = requestID
    switchState.value = 'loading'

    const snapshot = (await SetActiveGameAndGetSnapshot(gameID)) as GameSnapshot
    if (requestID !== latestRequestID.value) {
      return
    }

    applySnapshot(snapshot)
    activeGameID.value = gameID
    switchState.value = 'committing'
    transitionNonce.value += 1

    window.setTimeout(() => {
      if (requestID !== latestRequestID.value) {
        return
      }
      visibleGameID.value = gameID
      switchState.value = 'idle'
    }, 220)
  }
```

- [ ] **Step 5: Add refresh + warm loop actions**

```ts
  async function refreshActive(): Promise<void> {
    const gameID = activeGameID.value || visibleGameID.value
    if (!gameID) {
      return
    }
    const snapshot = (await GetGameSnapshot(gameID)) as GameSnapshot
    applySnapshot(snapshot)
  }

  async function warmNonActive(): Promise<void> {
    const result = (await WarmNonActiveGameSnapshots()) as Record<string, GameSnapshot>
    for (const key of Object.keys(result)) {
      applySnapshot(result[key])
    }
  }

  function startWarmLoop(): void {
    if (warmTimer.value !== null) {
      window.clearInterval(warmTimer.value)
    }
    warmTimer.value = window.setInterval(() => {
      void warmNonActive()
    }, 7 * 60 * 1000)
  }
```

- [ ] **Step 6: Export computed selectors**

```ts
  const visibleSnapshot = computed(() => snapshotsByGameID.value[visibleGameID.value] || null)
  const activeSnapshot = computed(() => snapshotsByGameID.value[activeGameID.value] || null)
  const hasColdStart = computed(() => startupState.value === 'cold' || !visibleSnapshot.value)

  return {
    snapshotsByGameID,
    activeGameID,
    visibleGameID,
    startupState,
    switchState,
    transitionNonce,
    visibleSnapshot,
    activeSnapshot,
    hasColdStart,
    bootstrap,
    switchGame,
    refreshActive,
    warmNonActive,
    startWarmLoop,
  }
})
```

- [ ] **Step 7: Typecheck store**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

- [ ] **Step 8: Commit Task 5**

```bash
git add frontend/src/stores/snapshots.ts
git commit -m "feat(frontend): add snapshot orchestrator store"
```

---

## Task 6: Refactor existing stores to use snapshot state (no hard resets)

**Files:**
- Modify: `frontend/src/stores/settings.ts`
- Modify: `frontend/src/stores/mods.ts`
- Modify: `frontend/src/stores/loadorder.ts`
- Modify: `frontend/src/stores/constraints.ts`

**Test:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Settings store delegates switch/fetch to snapshot store**

```ts
// settings.ts (inside setup)
import { useSnapshotsStore } from './snapshots'
const snapshotsStore = useSnapshotsStore()

const activeSnapshot = computed(() => snapshotsStore.activeSnapshot)

const activeGameID = computed(() => activeSnapshot.value?.gameID || 'eu5')
const availableGames = computed(() => activeSnapshot.value?.settings.availableGames || [])
const gameVersion = computed(() => activeSnapshot.value?.settings.gameVersion || 'unknown')

async function fetch(): Promise<void> {
  if (!snapshotsStore.activeSnapshot) {
    await snapshotsStore.bootstrap()
    return
  }
  await snapshotsStore.refreshActive()
}

async function setGame(id: string): Promise<void> {
  await snapshotsStore.switchGame(id)
}
```

- [ ] **Step 2: Remove game switch hard refetch fan-out**

```ts
// settings.ts - remove old pattern:
// await Promise.all([fetch(), modsStore.fetchAll(), loadOrderStore.fetch()])

// replace with:
async function setGame(id: string): Promise<void> {
  await snapshotsStore.switchGame(id)
}
```

- [ ] **Step 3: Mods store reads list from snapshot and keeps selection stable**

```ts
// mods.ts
const snapshotsStore = useSnapshotsStore()
const selectedModID = ref('')

const allMods = computed(() => snapshotsStore.activeSnapshot?.mods || [])
const enabledMods = computed(() => allMods.value.filter((m) => m.Enabled))

async function fetchAll(): Promise<void> {
  await snapshotsStore.refreshActive()
}

async function setEnabled(id: string, enabled: boolean): Promise<void> {
  if (enabled) {
    await EnableMod(id)
  } else {
    await DisableMod(id)
  }
  await snapshotsStore.refreshActive()
}
```

- [ ] **Step 4: Load-order store reads from snapshot**

```ts
// loadorder.ts
const snapshotsStore = useSnapshotsStore()

const orderedIDs = computed(() => snapshotsStore.activeSnapshot?.loadOrder || [])
const launcherLayout = computed(() => snapshotsStore.activeSnapshot?.launcherLayout || emptyLauncherLayout)
const playsetNames = computed(() => snapshotsStore.activeSnapshot?.playsetNames || [])

async function fetch(): Promise<void> {
  await snapshotsStore.refreshActive()
}

async function persist(ids: string[]): Promise<void> {
  await SetLoadOrder(ids)
  await snapshotsStore.refreshActive()
}
```

- [ ] **Step 5: Constraints store reads from snapshot**

```ts
// constraints.ts
const snapshotsStore = useSnapshotsStore()

const constraints = computed(() => snapshotsStore.activeSnapshot?.constraints || [])

async function fetch(): Promise<void> {
  await snapshotsStore.refreshActive()
}

async function add(from: string, to: string): Promise<void> {
  await AddConstraint(from, to)
  await snapshotsStore.refreshActive()
}
```

- [ ] **Step 6: Keep thumbnail polling but refresh snapshot (not hard fetch reset)**

```ts
// mods.ts startPolling
function startPolling(): void {
  window.setInterval(async () => {
    try {
      const available = await HasNewThumbnails()
      if (available) {
        await snapshotsStore.refreshActive()
      }
    } catch {
      // ignore polling errors
    }
  }, 1000)
}
```

- [ ] **Step 7: Typecheck after store migration**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

- [ ] **Step 8: Commit Task 6**

```bash
git add frontend/src/stores/settings.ts frontend/src/stores/mods.ts frontend/src/stores/loadorder.ts frontend/src/stores/constraints.ts
git commit -m "refactor(stores): drive domain stores from snapshot state"
```

---

## Task 7: Replace startup bootstrap path with snapshot bootstrap

**Files:**
- Modify: `frontend/src/main.ts`

**Test:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Remove old multi-store bootstrap function**

```ts
// remove old bootstrapData Promise.allSettled of 4 stores
```

- [ ] **Step 2: Bootstrap snapshots store after mount**

```ts
import { useSnapshotsStore } from './stores/snapshots'

app.use(pinia)
app.mount('#app')

async function bootstrapData(): Promise<void> {
  const snapshotsStore = useSnapshotsStore(pinia)
  await snapshotsStore.bootstrap()
}

void bootstrapData()
```

- [ ] **Step 3: Typecheck bootstrap update**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

- [ ] **Step 4: Commit Task 7**

```bash
git add frontend/src/main.ts
git commit -m "refactor(startup): bootstrap UI from snapshot store"
```

---

## Task 8: Implement App-level crossfade and startup skeleton gating

**Files:**
- Modify: `frontend/src/App.vue`

**Test:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Wire App to snapshots store state**

```ts
// App.vue <script setup>
import { useSnapshotsStore } from './stores/snapshots'

const snapshotsStore = useSnapshotsStore()
const appThemeClass = computed(() => `theme-${snapshotsStore.visibleGameID || 'eu5'}`)
const isSwitching = computed(() => snapshotsStore.switchState !== 'idle')
const isCold = computed(() => snapshotsStore.hasColdStart)
```

- [ ] **Step 2: Add crossfade layer wrapper around main content**

```vue
<main class="content" aria-label="Main content area">
  <div v-if="isCold" class="startup-skeleton">
    <div class="skeleton-line"></div>
    <div class="skeleton-line"></div>
    <div class="skeleton-line"></div>
  </div>
  <div v-else class="content-layer" :class="{ 'content-layer--switching': isSwitching }" :key="snapshotsStore.transitionNonce">
    <LoadOrderPanel @contextmenu="openContextMenu" @open-constraints="openConstraintModal" @manage-groups="manageGroupsOpen = true" />
  </div>
</main>
```

- [ ] **Step 3: Keep old layer visible during switching window**

```vue
<div class="content-transition-wrap">
  <Transition name="game-crossfade" mode="out-in">
    <div :key="`${snapshotsStore.visibleGameID}-${snapshotsStore.transitionNonce}`" class="content-layer">
      <LoadOrderPanel
        @contextmenu="openContextMenu"
        @open-constraints="openConstraintModal"
        @manage-groups="manageGroupsOpen = true"
      />
    </div>
  </Transition>
</div>
```

- [ ] **Step 4: Add switching indicator (subtle, non-blocking)**

```vue
<div v-if="isSwitching" class="sync-indicator" role="status" aria-live="polite">
  Syncing game snapshot...
</div>
```

- [ ] **Step 5: Typecheck App updates**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

- [ ] **Step 6: Commit Task 8**

```bash
git add frontend/src/App.vue
git commit -m "feat(ui): add snapshot-based crossfade and startup skeleton"
```

---

## Task 9: Update Sidebar and list panels for smooth motion

**Files:**
- Modify: `frontend/src/components/Sidebar.vue`
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/ModListPanel.vue`

**Test:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Sidebar uses snapshot switching state and action**

```ts
// Sidebar.vue
import { useSnapshotsStore } from '../stores/snapshots'
const snapshotsStore = useSnapshotsStore()

function selectGame(id: string) {
  void snapshotsStore.switchGame(id)
}

const availableGames = computed(() => snapshotsStore.visibleSnapshot?.settings.availableGames || [])
const activeGameID = computed(() => snapshotsStore.activeGameID)
```

- [ ] **Step 2: Disable game buttons only during switch commit**

```vue
<button
  class="game-btn"
  :disabled="snapshotsStore.switchState !== 'idle'"
  :class="{ active: activeGameID === gameID }"
  @click="selectGame(gameID)"
>
```

- [ ] **Step 3: Add TransitionGroup for load-order rows**

```vue
<TransitionGroup v-if="!block.collapsed" name="mod-row" tag="div" class="mod-list">
  <LoadOrderItem
    v-for="modID in block.modIds"
    :key="modID"
    :mod-i-d="modID"
    @contextmenu="onModContextMenu"
    @open-constraints="emit('open-constraints', modID)"
    @select="selectMod"
  />
</TransitionGroup>
```

- [ ] **Step 4: Add TransitionGroup for repository cards**

```vue
<TransitionGroup v-else class="cards" name="repo-card" tag="div">
  <ModCard
    v-for="mod in filteredMods"
    :key="mod.ID"
    :mod="mod"
    :selected="mod.ID === selectedModID"
    @toggle="(val: boolean) => toggleMod(mod, val)"
    @select="selectMod(mod)"
  />
</TransitionGroup>
```

- [ ] **Step 5: Add move/fade CSS classes for both groups**

```css
.mod-row-move,
.repo-card-move {
  transition: transform 180ms ease, opacity 180ms ease;
}

.mod-row-enter-active,
.mod-row-leave-active,
.repo-card-enter-active,
.repo-card-leave-active {
  transition: opacity 180ms ease, transform 180ms ease;
}

.mod-row-enter-from,
.mod-row-leave-to,
.repo-card-enter-from,
.repo-card-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
```

- [ ] **Step 6: Typecheck motion updates**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

- [ ] **Step 7: Commit Task 9**

```bash
git add frontend/src/components/Sidebar.vue frontend/src/components/LoadOrderPanel.vue frontend/src/components/ModListPanel.vue
git commit -m "feat(ui): animate game/list transitions with snapshot states"
```

---

## Task 10: Add global theme and text-color transition contract

**Files:**
- Modify: `frontend/src/assets/main.css`
- Modify: `frontend/src/App.vue`

**Test:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Add dedicated theme transition token**

```css
:root,
:root[data-theme='dark'] {
  --transition-theme: 220ms cubic-bezier(0.22, 1, 0.36, 1);
}
```

- [ ] **Step 2: Apply color/text/border transitions to app shell + text nodes**

```css
.shell,
.shell .sidebar,
.shell .content,
.shell .mod-group,
.shell .disabled-mod,
.shell .mod-row,
.shell .game-btn,
.shell .repo-title,
.shell .name,
.shell .subtitle,
.shell .version,
.shell .state {
  transition:
    background-color var(--transition-theme),
    color var(--transition-theme),
    border-color var(--transition-theme),
    box-shadow var(--transition-theme);
}
```

- [ ] **Step 3: Add App-level class to enable/disable animated window**

```vue
<div class="shell" :class="[appThemeClass, { 'shell--theme-transition': isSwitching }]">
```

```css
.shell--theme-transition {
  will-change: background-color, color, border-color;
}
```

- [ ] **Step 4: Add reduced-motion guard**

```css
@media (prefers-reduced-motion: reduce) {
  .shell,
  .shell * {
    transition-duration: 1ms !important;
    animation-duration: 1ms !important;
  }
}
```

- [ ] **Step 5: Add App crossfade classes**

```css
.game-crossfade-enter-active,
.game-crossfade-leave-active {
  transition: opacity 200ms ease, transform 200ms ease;
}

.game-crossfade-enter-from,
.game-crossfade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
```

- [ ] **Step 6: Typecheck and visual smoke check**

Run: `cd frontend && npx tsc --noEmit`  
Expected: no TypeScript errors.

Manual check:
- switch EU5 ↔ HOI4 rapidly; text color should transition, not snap.

- [ ] **Step 7: Commit Task 10**

```bash
git add frontend/src/assets/main.css frontend/src/App.vue
git commit -m "style(theme): animate text and surface colors on game switch"
```

---

## Task 11: Final cleanup + verification pass

**Files:**
- Modify: `app.go`
- Modify: `snapshot_models.go`
- Modify: `snapshot_service.go`
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/stores/snapshots.ts`
- Modify: `frontend/src/stores/settings.ts`
- Modify: `frontend/src/stores/mods.ts`
- Modify: `frontend/src/stores/loadorder.ts`
- Modify: `frontend/src/stores/constraints.ts`
- Modify: `frontend/src/components/Sidebar.vue`
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/ModListPanel.vue`
- Modify: `frontend/src/assets/main.css`
- Modify: `frontend/wailsjs/go/main/App.d.ts`
- Modify: `frontend/wailsjs/go/main/App.js`
- Modify: `frontend/wailsjs/go/models.ts`

**Test:**
- Required full verification.

- [ ] **Step 1: Remove obsolete fetch chaining paths**

Code to remove where present:

```ts
await Promise.all([fetch(), modsStore.fetchAll(), loadOrderStore.fetch()])
selectedModID.value = ''
allMods.value = []
```

Replace with snapshot refresh/switch calls already added.

- [ ] **Step 2: Ensure no direct game switch refetch fan-out remains**

Search and confirm no leftover direct fan-out calls.

Run: `rg "Promise\.all\(\[fetch\(\), modsStore\.fetchAll\(\), loadOrderStore\.fetch\(\)\]\)" frontend/src`
Expected: no matches.

- [ ] **Step 3: Run backend build verification**

Run: `go build ./...`  
Expected: success.

- [ ] **Step 4: Run backend vet verification**

Run: `go vet ./...`  
Expected: success.

- [ ] **Step 5: Run frontend typecheck verification**

Run: `cd frontend && npx tsc --noEmit`  
Expected: success.

- [ ] **Step 6: Manual UX checklist (required)**

- Startup: skeletons appear only before first snapshot.
- Game switch: no blank screen.
- Game switch: text + surface colors tween smoothly.
- Mod enable/disable: list movement animated, no flash-disappear.
- Theme switch repeated quickly: no stuck transition state.
- Warm refresh cycle: no visible layout reset.

- [ ] **Step 7: Commit final integration changes**

```bash
git add app.go snapshot_models.go snapshot_service.go frontend/src frontend/wailsjs/go/main frontend/wailsjs/go/models.ts
git commit -m "feat(snapshot-ui): deliver smooth snapshot-driven multi-game UX"
```

---

## Post-implementation notes for next session

- If Wails generation omits new methods in `frontend/wailsjs`, rerun `wails generate module` after ensuring new Go methods are exported on `App`.
- Keep `WarmNonActiveGameSnapshots` non-blocking for active flow; failures should not bubble as fatal UI errors.
- Do not reintroduce aggressive local resets in stores (`allMods=[]`, `selectedModID=''`) on routine refresh.

---

## Coverage check against approved design

- Backend aggregated snapshots: **covered (Tasks 1-3)**
- Atomic switch API + no-blank flow: **covered (Tasks 2, 5, 8)**
- Startup warm + sparse refresh: **covered (Tasks 2, 5)**
- Skeleton on cold and subtle sync on warm: **covered (Tasks 5, 8)**
- Theme transition smoothing (including text color): **covered (Task 10)**
- Animated list movement instead of hard state jumps: **covered (Task 9)**
- Backend source of truth preserved: **covered (Tasks 2-3, 6)**
- Required project verification commands: **covered (Task 11)**
