# Smooth UX Snapshot Redesign Spec

**Date:** 2026-04-18  
**Status:** Approved (design)  
**Topic:** Multi-game smoothness rewrite with backend snapshot aggregation

---

## 1. Problem Statement

Current UI state flow is backend-authoritative but visually unstable during fetch-heavy actions (especially game switching and mod enable/disable refresh cycles). Users see brief disappear/reappear behavior because frontend composes screen state from multiple async store fetches and transient reset states.

### Symptoms observed
- Game switch triggers multiple store refetches (`settings`, `mods`, `loadorder`) and hard state resets.
- Components often render intermediate empty/loading states even when stale data would be acceptable.
- Theme changes happen abruptly with no global transition layer.
- Loading indicators are text-only and appear in places where subtle progressive loading should be used.

---

## 2. Design Goals

1. **Backend remains sole source of truth.**
2. **No blank-frame transitions** on game switch.
3. **Background warm caches** for non-active games (startup + sparse refresh).
4. **Smooth global theme transitions** when switching games.
5. **Skeleton loaders only for truly cold/long loads** (mostly startup / first uncached access).
6. **Aggressive UX polish acceptable** (active dev stage, broad refactor allowed).

### Non-goals
- Backward-compatible frontend state shape.
- Minimal-diff implementation.
- New test suite additions (existing tests must still pass).

---

## 3. High-Level Solution

Use **backend snapshot aggregation** as primary read model, then drive UI from cached snapshots with stale-while-revalidate behavior.

### Core decisions (approved)
- Approach: **Backend-first aggregator endpoint(s)**.
- Prefetch strategy: **Startup warm + sparse refresh**.
- Switch UX: **Keep previous view visible, then crossfade to target snapshot**.
- Theme transition: **Global color tween**.
- Rollout: **Phased incremental**.
- Refresh cadence: **5–10 min target window** (default implementation can start at ~7 min).

---

## 4. Backend Architecture

### 4.1 New Snapshot Read Model

Introduce backend type:

```go
type GameSnapshot struct {
    GameID            string
    Mods              []mods.Mod
    LoadOrder         []string
    LauncherLayout    LauncherLayout
    Constraints       []graph.Constraint
    PlaysetNames      []string
    GameActiveIndex   int
    LauncherActiveIdx int
    Settings          SnapshotSettings
    Metadata          SnapshotMetadata
}

type SnapshotSettings struct {
    ModsDirStatus        ModsDirStatus
    GameExe              string
    AutoDetectedGameExe  string
    ConfigPath           string
    GameVersion          string
    GameVersionOverride  string
    AvailableGames       []string
}

type SnapshotMetadata struct {
    Revision  int64
    FetchedAt int64
    Stale     bool
}
```

Notes:
- Exact field naming can align with existing Wails model style.
- `Revision` enables stale-response rejection on frontend.

### 4.2 New Backend APIs

Add aggregated calls:
- `GetGameSnapshot(gameID string) (GameSnapshot, error)`
- `SetActiveGameAndGetSnapshot(gameID string) (GameSnapshot, error)`
- `WarmNonActiveGameSnapshots() (map[string]GameSnapshot, error)` *(or equivalent warm call returning partial status)*

Behavior:
- `SetActiveGameAndGetSnapshot` performs switch + `refreshState` + snapshot build in one backend transaction-like flow.
- `GetGameSnapshot` can serve either active or non-active game context.
- Snapshot reads for non-active games must not mutate user-visible active game state; implementation can use isolated per-game loaders or temporary context swap with guaranteed restore.
- Warm endpoints must never change active game selection.

### 4.3 Snapshot Build Pipeline

Single builder function (conceptual):
1. Resolve target game context.
2. Ensure core state loaded (`refreshState` / equivalent per-game load path).
3. Gather mods/load order/layout/constraints/settings.
4. Stamp `Revision` + `FetchedAt`.
5. Store in backend snapshot cache.

### 4.4 Cache + Invalidation

Maintain per-game snapshot cache:
- `map[string]GameSnapshot`
- `map[string]int64` revisions
- guarded by mutex/RWMutex

Invalidate/rebuild affected snapshot on mutating operations:
- `EnableMod`, `DisableMod`, `SetLoadOrder`, `SetLauncherLayout`
- constraint updates
- playset selection changes
- settings changes affecting mod scan/version/path resolution

### 4.5 Background Warm Strategy

- After initial startup snapshot for active game, spawn background warm for all other games.
- Run sparse refresh on non-active cached snapshots every 5–10 min.
- Warm work must be low priority and must not block active game requests.
- On warm failure: keep previous cached snapshot + mark stale.

### 4.6 Legacy SQLite Path Invariant

- For legacy sqlite titles, snapshot loading must use `launcher-v2.sqlite` as primary path (not `launcher-v2.db`).
- If fallback compatibility checks exist, they must not override `.sqlite` first-path behavior.

---

## 5. Frontend Architecture

### 5.1 Snapshot-Orchestrator Store

Create single orchestrator store (new):
- `snapshotsByGameID`
- `activeGameID`
- `visibleGameID`
- `startupState` (`cold | warming | ready`)
- `switchState` (`idle | loading | committing`)
- `latestRequestID`

Existing domain stores can be:
1. folded into orchestrator, or
2. retained as thin read-only selectors over active snapshot.

Preferred direction: fold into orchestrator for consistency and to reduce multi-store race conditions.

### 5.2 Switch Flow (No Flicker)

1. User selects target game.
2. Keep current `visibleGameID` rendered.
3. Start `SetActiveGameAndGetSnapshot(target)` request with request token.
4. On success (latest token only):
   - update `snapshotsByGameID[target]`
   - set `activeGameID=target`
   - trigger crossfade transition from `visible` to `target`
   - after transition, set `visibleGameID=target`
5. On failure:
   - retain current visible snapshot
   - show non-blocking error toast/retry affordance

### 5.3 Loading/Skeleton Rules

- **Cold startup / first uncached game:** full skeletons.
- **Cached snapshot exists:** render stale snapshot immediately + subtle syncing indicator.
- Avoid replacing full panels with loading text when stale snapshot can be shown.

---

## 6. Motion and Visual System

### 6.1 Global Theme Tween

- Keep per-game theme classes (`theme-eu5`, `theme-hoi4`, etc.).
- Add global transition contract for major color vars and visual surfaces (~180–220ms).
- Apply transition on shell + major panels + cards + borders + accent states.
- Respect `prefers-reduced-motion` (reduce/disable non-essential transitions).

### 6.2 Content Crossfade

- Dual-layer render container for game content.
- Old layer remains interactive-disabled but visible during short transition.
- New layer enters with opacity/transform animation.
- Swap complete, old layer unmounts.

### 6.3 List Motion

- Use `TransitionGroup` for mod lists and load-order rows.
- Animate only transform/opacity for reorder/toggle moves.
- No expensive layout-heavy animation properties.

---

## 7. Error Handling and Consistency

- Each switch/warm request gets request ID.
- Discard late responses that are not latest in flight.
- Frontend applies snapshot only if `revision` is >= current cached revision for that game.
- Warm failures mark stale and retry next cycle; never block main flow.
- Switch failures preserve previous visible data; no blank fallback.

---

## 8. File/Module Impact (Expected)

### Backend
- `app.go` (new snapshot endpoints + orchestration)
- potential new files under `internal/service/` for snapshot assembly and cache
- Wails model/binding generation updates for `GameSnapshot`

### Frontend
- new orchestrator snapshot store under `frontend/src/stores/`
- `Sidebar.vue` switch logic to orchestrator actions
- `App.vue` shell to render transition container
- `LoadOrderPanel.vue`, `ModListPanel.vue`, details components to consume snapshot-derived state
- `frontend/src/assets/main.css` transition tokens and reduced-motion handling

---

## 9. Rollout Plan (Phased)

1. **Phase 1:** Backend snapshot model + APIs + revisioning.
2. **Phase 2:** Frontend orchestrator store + atomic switch path.
3. **Phase 3:** Theme tween + content crossfade + list animations.
4. **Phase 4:** Background warm scheduler + stale badge/sync polish.
5. **Phase 5:** Remove old multi-call fetch/switch path and dead transitional logic.

---

## 10. Verification Plan

Required checks after implementation phases:
- `go build ./...`
- `go vet ./...` *(if interfaces changed)*
- `tsc --noEmit`

Manual UX checks:
- startup skeleton only on cold load
- no blank frame on game switch
- smooth global theme transition on each game switch
- enable/disable/reorder transitions feel continuous (no disappear/reappear flash)
- backend remains authoritative after all mutations

---

## 11. Risks and Mitigations

### Risk: Snapshot build overhead on slower systems
- Mitigation: warm non-active games sparsely, not aggressively.

### Risk: Concurrency races between warm and active switch
- Mitigation: request tokens + per-game revision checks + lock discipline backend side.

### Risk: Regression from broad refactor
- Mitigation: phased rollout with checkpoint validation after each phase.

---

## 12. Visual Companion Artifact

Companion mock/diagram file for manual opening:
- `docs/superpowers/specs/2026-04-18-smooth-ux-visual-companion.html`

It illustrates:
- snapshot switch timeline
- stale-while-revalidate behavior
- global theme tween intent
- skeleton-vs-stale rendering rules
