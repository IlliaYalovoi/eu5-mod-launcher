# UI Rewrite Design — 2026-04-12

## Overview

Complete rewrite of frontend to achieve: (1) ultra-puppet UI — backend owns all state, frontend is pure display, (2) token-optimized — small files, maximum reuse, minimal boilerplate, (3) Vue 3 minimal Composition API — strip Pinia, no local caching.

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Puppet level | A — Ultra-puppet | Explicit requirement: no frontend logic, no caching |
| Framework | C — Vue minimal | Stay on Vue, strip stores, use bare Composition API |
| Component arch | B — Keep panels, strip logic | Existing panels become thin containers |
| Data fetching | A — Fetch-per-component | Each component calls backend on mount |
| Error handling | B — Toast + inline loading | One toast container, loading local per component |

---

## Architecture

### Data Flow

```
Component mounts → onMounted fetch → Backend → Data → Template render
                                    ↓
                              Error → Toast
```

**Rules:**
- NO Pinia stores
- NO `ref()` for data that comes from backend (use `onMounted` fetch)
- NO `computed()` for derived state from backend data
- YES: local UI state (panel open/closed, input values, hover)
- YES: `onMounted` calls backend, sets local `ref()` with result

### Component Hierarchy

```
App.vue                    → Layout shell, modals, toast container
├── GameSelector           → Fetch + display detected games, switch active
├── LoadOrderPanel         → Container, drag-drop, fetch-per-mod cards
│   ├── LoadOrderItem      → Single mod row, fetch own state, emit events
│   └── LaunchButton       → Fetch + call launch
├── ModDetailsPanel        → Fetch mod details on open
├── ConstraintModal        → Fetch constraints, add/remove via backend
├── SettingsPanel          → Fetch settings, save via backend
├── AutosortButton         → Call backend autosort, refresh
├── CycleErrorPanel        → Display only
├── ManualGamePathSetup    → Form + backend save
└── ToastContainer         → Singleton, no props
```

### File Structure

```
frontend/src/
├── main.ts
├── App.vue                 # Layout shell, modal orchestration
├── types.ts                # Shared types (from backend)
├── wailsjs/                # GENERATED — DO NOT EDIT
│   └── go/launcher/App.ts  # Wails bindings
├── components/
│   ├── GameSelector.vue     # Game list + active switch
│   ├── LoadOrderPanel.vue   # Drag-drop container
│   ├── LoadOrderItem.vue    # Single mod row
│   ├── ModDetailsPanel.vue  # Mod info + workshop data
│   ├── ConstraintModal.vue  # Constraint CRUD
│   ├── SettingsPanel.vue   # Settings form
│   ├── AutosortButton.vue  # Trigger + error display
│   ├── LaunchButton.vue     # Launch game action
│   ├── CycleErrorPanel.vue # Cycle display
│   ├── ManualGamePathSetup.vue
│   └── ui/
│       ├── ToastContainer.vue  # Error/success toasts
│       ├── BaseButton.vue      # Reused button style
│       ├── BaseBadge.vue       # Reused tag style
│       ├── BaseModal.vue        # Reused modal shell
│       ├── SearchInput.vue     # Reused search
│       └── ModPicker.vue       # Mod selection dropdown
└── lib/
    ├── toast.ts            # Shared toast emitter (no state)
    └── error.ts            # Shared error formatter
```

**Deleted:**
- `stores/` (all Pinia stores — gone)
- `utils/theme.ts` (if purely presentational)
- `utils/steamDescription.ts` (logic moves to backend)
- `lib/logger.ts` (backend logging only)

---

## Component Spec

### App.vue

**Role:** Layout shell, modal orchestration, toast host.

**State (local only):**
- `detailsOpen: boolean`
- `settingsOpen: boolean`
- `constraintModalOpen: boolean`
- `constraintTargetModID: string`

**Behavior:**
- No fetch on mount — delegates to child components
- Renders panels as `<Teleport to="body">`
- Catches toast events, renders ToastContainer

### LoadOrderPanel.vue

**Role:** Drag-drop container, fetches layout + order, passes to LoadOrderItem.

**Fetch on mount:**
```ts
const layout = await GetLauncherLayout()
const order = await GetLoadOrder()
```

**Renders:** `draggable` list of `LoadOrderItem` components with `v-model`.

**On drag end:** calls `SetLoadOrder(orderedIDs)` then `SaveCompiledLoadOrder()`.

**Emits:** `select-mod`, `context-menu`, `open-constraints`.

### LoadOrderItem.vue

**Role:** Single mod row, fetches own mod data if needed.

**Props:** `modID: string`, `index: number`, `layout: LauncherLayout`

**Fetch on mount:** none — data comes from parent via props.

**State (local):**
- `hovered: boolean`
- `editingName: boolean` (for category rename inline)

**Events emitted:**
- `select`
- `contextmenu(x, y, modID)`
- `open-constraints(modID)`
- `toggle-enabled(modID)` → calls backend `EnableMod`/`DisableMod`, then emits refresh

### ToastContainer.vue

**Role:** Singleton — receives toast events, renders stack.

**Implementation:**
```ts
// lib/toast.ts — shared emitter, NO state
type Toast = { id: string; type: 'success'|'error'|'info'; message: string }
const listeners = new Set<(t: Toast) => void>()
export function showToast(t: Toast) { listeners.forEach(fn => fn(t)) }
```

**Pattern:** ToastContainer subscribes to emitter on mount, unsubscribes on unmount. Components call `showToast(...)` directly — no props, no context.

### UI Atoms (`ui/`)

**All are presentational wrappers:**
- `BaseButton.vue` — button with variants (primary/danger/ghost)
- `BaseBadge.vue` — tag/label display
- `BaseModal.vue` — teleport + backdrop + close
- `SearchInput.vue` — input with icon
- `ModPicker.vue` — dropdown for picking a mod

**Rule:** ZERO business logic. Props in, DOM out. No `await`, no `fetch`.

---

## Error Handling

All backend calls wrapped in try/catch:

```ts
async function loadMods() {
  loading.value = true
  try {
    data.value = await GetAllMods()
  } catch (err) {
    showToast({ id: crypto.randomUUID(), type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}
```

`errorMessage()` helper in `lib/error.ts`:
```ts
export function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
```

---

## Removed Business Logic

The following currently live in frontend stores/components and must move to backend:

| Current Location | Logic | New Location |
|---|---|---|
| `mods.ts` `modByID()` | Find mod by ID | Backend: `GetModByID(id)` |
| `mods.ts` `workshopItemIDFromDirPath()` | Parse Steam path | Backend: `GetWorkshopItemID(modID)` |
| `mods.ts` `isWorkshopMod()` | Check if workshop mod | Backend: `IsWorkshopMod(modID)` |
| `mods.ts` `allMods` filtering | Filter/disable mods | Backend: `GetAllMods(gameID)` returns filtered |
| `loadorder.ts` `compiledOrder` | Flatten layout to array | Backend: `GetCompiledOrder()` |
| `loadorder.ts` `modsByID` | Build lookup map | Backend: `GetModsByID(ids[])` |
| `constraints.ts` `forMod()` | Get constraints for mod | Backend: `GetConstraintsForMod(modID)` |
| `constraints.ts` `byType()` | Filter by type | Backend: `GetConstraints(type)` |
| `games.ts` `initialize()` | Set active game logic | Backend: `Initialize()` |

**Backend must expose:**
- `GetAllMods(gameID string)` → `[]Mod` (already exists)
- `GetModByID(id string)` → `*Mod`
- `GetWorkshopItemID(modID string)` → `string`
- `IsWorkshopMod(modID string)` → `bool`
- `GetLauncherLayout()` → `LauncherLayout` (already exists)
- `GetCompiledOrder()` → `[]string` (mod IDs flattened)
- `GetConstraintsForMod(modID string)` → `[]Constraint`
- `GetConstraints(gameID string, type string)` → `[]Constraint`
- `ListSupportedGames()` → `[]Game` (already exists)
- `SetActiveGame(id string)` (already exists)

---

## Drag-and-Drop

Current: `vuedraggable` + Pinia store reactive binding.

Rewritten: `vuedraggable` + local refs. On `end` event:
1. Read new order from `v-model`
2. Call `SetLoadOrder(newOrder)`
3. Call `SaveCompiledLoadOrder()`
4. On error: show toast, revert to previous order

No reactive store sync — just imperative fetch → update local ref → persist.

---

## Wails Bindings

Frontend imports from `../../wailsjs/go/launcher/App`.

All function signatures already return `(Result, error)`. No interface change.

**New bindings to add to backend:**
- `GetModByID(id string) (Mod, error)`
- `GetWorkshopItemID(modID string) (string, error)`
- `IsWorkshopMod(modID string) (bool, error)`
- `GetCompiledOrder() ([]string, error)`
- `GetConstraintsForMod(modID string) ([]Constraint, error)`
- `GetConstraints(gameID string, constraintType string) ([]Constraint, error)`

Existing bindings reused as-is.

---

## Migration Order

1. **Create `lib/toast.ts` and `lib/error.ts`** — shared utilities, no dependencies
2. **Create `ToastContainer.vue`** — verify toast system works
3. **Migrate `GameSelector.vue`** — fetch-per-component, no store
4. **Migrate `LoadOrderPanel.vue`** — strip store, keep drag-drop
5. **Migrate `LoadOrderItem.vue`** — strip store, props in
6. **Migrate `ModDetailsPanel.vue`** — fetch on open
7. **Migrate `ConstraintModal.vue`** — fetch constraints
8. **Migrate `AutosortButton.vue`** — call + toast on error
9. **Migrate `LaunchButton.vue`** — call + toast
10. **Migrate `SettingsPanel.vue`** — fetch + save
11. **Migrate `CycleErrorPanel.vue`** — presentational only
12. **Migrate `ManualGamePathSetup.vue`** — form + backend
13. **Strip `stores/` directory**
14. **Strip `utils/steamDescription.ts`** (if exists)
15. **Strip `lib/logger.ts` from frontend**
16. **App.vue cleanup** — remove store imports, add modal orchestration
17. **Add missing backend bindings**
18. **Verify all `tsc --noEmit` passes**
19. **Manual browser test**

---

## Reuse Opportunities

| Item | Current | Reused |
|---|---|---|
| `BaseButton.vue` | In use | Yes — all buttons |
| `BaseBadge.vue` | In use | Yes — tags, constraints |
| `BaseModal.vue` | In use | Yes — all modals |
| `SearchInput.vue` | In use | Yes — search bars |
| `ModPicker.vue` | In use | Yes — constraint modal |
| Toast system | No shared emitter | New `lib/toast.ts` |
| Error formatter | In each store | New `lib/error.ts` |

---

## File Count Change

| Category | Before | After |
|---|---|---|
| Vue components | 22 | ~18 |
| Pinia stores | 5 | 0 |
| Utils/lib | 3 | 2 |
| TS types | 1 | 1 |

Net: ~15 files removed, 1 new lib directory with 2 files.

---

## Success Criteria

- Zero `ref()` wrapping backend data in components
- Zero `computed()` deriving state from other refs (local UI computed OK)
- Zero Pinia imports in any component
- Every component error goes through toast system
- All drag-drop state persisted via backend calls only
- `tsc --noEmit` passes
- Wails dev mode works end-to-end
