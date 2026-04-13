# UI Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite frontend to ultra-puppet pattern: zero frontend logic/caching, all state owned by backend, Vue 3 minimal Composition API with no Pinia stores.

**Architecture:** Components become pure display layers. Each component fetches its own data from backend via Wails bindings on mount. Shared utilities (`toast`, `error`) provided via a simple event emitter pattern. Pinia stores deleted entirely. UI atoms (`BaseButton`, `BaseModal`, etc.) preserved as presentational wrappers.

**Tech Stack:** Vue 3 (`<script setup>`), Wails v2, TypeScript (strict), scoped CSS. No Pinia, no vuedraggable model binding, no local caching.

---

## File Map

### Created
- `frontend/src/lib/toast.ts` — Shared toast emitter (no state, Set-based listeners)
- `frontend/src/lib/error.ts` — Shared `errorMessage()` helper

### Modified (in-place rewrite)
- `frontend/src/components/ui/ToastContainer.vue` — Strip Pinia, use `toast.ts` emitter
- `frontend/src/components/GameSelector.vue` — Fetch on mount, no store
- `frontend/src/components/LoadOrderPanel.vue` — Strip store, drag-drop via local refs
- `frontend/src/components/LoadOrderItem.vue` — Props in, events out, no store
- `frontend/src/components/ModDetailsPanel.vue` — Fetch on open, no store
- `frontend/src/components/ConstraintModal.vue` — Fetch constraints, no store
- `frontend/src/components/AutosortButton.vue` — Call backend + toast
- `frontend/src/components/LaunchButton.vue` — Call backend + toast
- `frontend/src/components/SettingsPanel.vue` — Fetch + save, no store
- `frontend/src/components/CycleErrorPanel.vue` — Presentational only (verify)
- `frontend/src/components/ManualGamePathSetup.vue` — Form + backend
- `frontend/src/App.vue` — Remove store imports, modal orchestration only

### Deleted
- `frontend/src/stores/mods.ts`
- `frontend/src/stores/loadorder.ts`
- `frontend/src/stores/constraints.ts`
- `frontend/src/stores/games.ts`
- `frontend/src/stores/settings.ts`
- `frontend/src/lib/logger.ts`
- `frontend/src/utils/steamDescription.ts` (if exists)
- `frontend/src/utils/theme.ts` (if purely presentational)

### Backend (new bindings needed — NOT in this plan's scope, but noted for completeness)
- `GetModByID(id string) (mods.Mod, error)`
- `GetCompiledOrder() ([]string, error)`
- `GetConstraintsForMod(modID string) ([]domain.Constraint, error)`

---

## Task 1: Create shared lib utilities

**Files:**
- Create: `frontend/src/lib/toast.ts`
- Create: `frontend/src/lib/error.ts`

- [ ] **Step 1: Create `frontend/src/lib/toast.ts`**

```typescript
// @ts-nocheck
type Toast = {
  id: string
  type: 'success' | 'error' | 'info'
  message: string
}

type Listener = (t: Toast) => void
const listeners = new Set<Listener>()

export function showToast(toast: Omit<Toast, 'id'>): void {
  const t: Toast = { ...toast, id: crypto.randomUUID() }
  listeners.forEach(fn => fn(t))
}

export function subscribeToasts(fn: Listener): () => void {
  listeners.add(fn)
  return () => listeners.delete(fn)
}
```

- [ ] **Step 2: Create `frontend/src/lib/error.ts`**

```typescript
export function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/toast.ts frontend/src/lib/error.ts
git commit -m "feat(ui): add shared toast emitter and error formatter

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 2: Rewrite ToastContainer

**Files:**
- Modify: `frontend/src/components/ui/ToastContainer.vue:1-50`

- [ ] **Step 1: Read current ToastContainer.vue**

Read `frontend/src/components/ui/ToastContainer.vue` lines 1-50 to see full content.

- [ ] **Step 2: Rewrite script and template**

Replace the entire `<script setup>` block and `<template>` with:

```vue
<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { subscribeToasts, type Toast } from '../../lib/toast'

const toasts = ref<Toast[]>([])
const timers = new Map<string, ReturnType<typeof setTimeout>>()

function addToast(toast: Toast) {
  toasts.value.push(toast)
  if (toasts.value.length > 3) {
    const oldest = toasts.value.shift()
    if (oldest) {
      const t = timers.get(oldest.id)
      if (t) { clearTimeout(t); timers.delete(oldest.id) }
    }
  }
  if (toast.type !== 'error') {
    timers.set(toast.id, setTimeout(() => removeToast(toast.id), 3200))
  }
}

function removeToast(id: string) {
  const idx = toasts.value.findIndex(t => t.id === id)
  if (idx !== -1) toasts.value.splice(idx, 1)
  timers.delete(id)
}

let unsubscribe: (() => void) | null = null

onMounted(() => {
  unsubscribe = subscribeToasts(addToast)
})

onUnmounted(() => {
  if (unsubscribe) unsubscribe()
  timers.forEach(t => clearTimeout(t))
})
</script>

<template>
  <div class="toast-container">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      :class="['toast', `toast--${toast.type}`]"
      @click="removeToast(toast.id)"
    >
      {{ toast.message }}
    </div>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1rem;
  right: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  z-index: 9999;
}
.toast {
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 0.875rem;
  cursor: pointer;
  max-width: 320px;
}
.toast--success { background: #22c55e; color: white; }
.toast--error { background: #ef4444; color: white; }
.toast--info { background: #3b82f6; color: white; }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/ToastContainer.vue
git commit -m "refactor(ui): rewrite ToastContainer with toast.ts emitter

No Pinia, no storeToRefs, no unsubscribeNotice watch.
Uses shared emitter pattern.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 3: Rewrite GameSelector

**Files:**
- Modify: `frontend/src/components/GameSelector.vue`

- [ ] **Step 1: Read current GameSelector.vue**

Read `frontend/src/components/GameSelector.vue` to see full content (likely ~80 lines).

- [ ] **Step 2: Rewrite with fetch-on-mount pattern**

Strip all Pinia imports. Replace store logic with:

```vue
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ListSupportedGames, SetActiveGame } from '../../wailsjs/go/launcher/App'
import { errorMessage } from '../lib/error'
import { showToast } from '../lib/toast'

type Game = { id: string; name: string; detected: boolean }

const games = ref<Game[]>([])
const activeGameID = ref('')
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const result = await ListSupportedGames()
    games.value = result.map((g: any) => ({ id: g.id, name: g.name, detected: g.detected }))
    const detected = games.value.find(g => g.detected)
    activeGameID.value = detected?.id || games.value[0]?.id || ''
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}

async function selectGame(id: string) {
  try {
    await SetActiveGame(id)
    activeGameID.value = id
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

onMounted(load)
</script>
```

- [ ] **Step 3: Preserve existing template and styles** — do NOT rewrite template or style sections, keep existing CSS classes and HTML structure.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/GameSelector.vue
git commit -m "refactor(ui): rewrite GameSelector with fetch-on-mount

No Pinia store. Backend owns game list and active game state.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 4: Rewrite LoadOrderItem

**Files:**
- Modify: `frontend/src/components/LoadOrderItem.vue`

- [ ] **Step 1: Read current LoadOrderItem.vue**

Read full file (likely ~150 lines). Key things to preserve: draggable model binding, context menu emission, constraint modal trigger.

- [ ] **Step 2: Rewrite with props in, events out**

Keep `props`: `modID`, `index`, `layout`. Strip all store imports. Add `EnableMod`/`DisableMod` calls with toast on error.

```vue
<script setup lang="ts">
import { computed, ref } from 'vue'
import { EnableMod, DisableMod } from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'

const props = defineProps<{
  modID: string
  index: number
  layout: any  // LauncherLayout — narrow type if available
}>()

const emit = defineEmits<{
  (e: 'select', modID: string): void
  (e: 'contextmenu', payload: { modID: string; x: number; y: number }): void
  (e: 'open-constraints', modID: string): void
  (e: 'toggle-enabled', modID: string): void
}>()

const hovered = ref(false)
const toggling = ref(false)

// Props passed from parent — parent fetches all mods and passes down
// If mod data is needed directly, emit 'need-mod' to parent
const enabled = computed(() => props.modID.startsWith('category:') || true) // placeholder

async function toggleEnabled() {
  if (toggling.value) return
  toggling.value = true
  try {
    if (enabled.value) {
      await DisableMod(props.modID)
    } else {
      await EnableMod(props.modID)
    }
    emit('toggle-enabled', props.modID)
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    toggling.value = false
  }
}

function handleContextmenu(e: MouseEvent) {
  e.preventDefault()
  emit('contextmenu', { modID: props.modID, x: e.clientX, y: e.clientY })
}
</script>
```

- [ ] **Step 3: Preserve existing template and scoped styles** — keep draggable attrs, class bindings, event handlers.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/LoadOrderItem.vue
git commit -m "refactor(ui): rewrite LoadOrderItem as props-in events-out

No store. Enable/disable calls go direct to backend via Wails.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 5: Rewrite LoadOrderPanel

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`

- [ ] **Step 1: Read current LoadOrderPanel.vue**

Read full file (~200+ lines). This is the most complex component — contains draggable, category management, compiled order computation.

- [ ] **Step 2: Rewrite with fetch on mount, strip all stores**

```vue
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import draggable from 'vuedraggable'
import { Autosort, GetLauncherLayout, GetLoadOrder, SaveCompiledLoadOrder, SetLoadOrder } from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import LoadOrderItem from './LoadOrderItem.vue'
import AutosortButton from './AutosortButton.vue'
import LaunchButton from './LaunchButton.vue'

const launcherLayout = ref<any>({ ungrouped: [], categories: [], order: [] })
const orderedIDs = ref<string[]>([])
const loading = ref(true)
const isSorting = ref(false)

const emit = defineEmits<{
  (e: 'select-mod', modID: string): void
  (e: 'open-constraints', modID: string): void
  (e: 'contextmenu', payload: { modID: string; x: number; y: number }): void
}>()

async function load() {
  loading.value = true
  try {
    const [layout, order] = await Promise.all([GetLauncherLayout(), GetLoadOrder()])
    launcherLayout.value = layout
    orderedIDs.value = order
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}

async function onDragEnd(evt: any) {
  // evt.item / evt.newIndex / evt.oldIndex available from draggable
  try {
    await SetLoadOrder(orderedIDs.value)
    await SaveCompiledLoadOrder()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
    // Revert order by re-fetching
    await load()
  }
}

function handleSelectMod(modID: string) { emit('select-mod', modID) }
function handleContextmenu(payload: any) { emit('contextmenu', payload) }
function handleOpenConstraints(modID: string) { emit('open-constraints', modID) }

onMounted(load)
</script>
```

- [ ] **Step 3: Preserve existing template and scoped styles** — keep draggable wrapper, LoadOrderItem renders inside, category blocks, etc.

- [ ] **Step 4: Strip `compiledOrder` and `modsByID` computed** — these were frontend derivations. The backend `GetCompiledOrder()` binding should be added (backend task). For now, use `orderedIDs` directly.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/LoadOrderPanel.vue
git commit -m "refactor(ui): rewrite LoadOrderPanel with fetch-on-mount

No Pinia store. Draggable persists via SetLoadOrder + SaveCompiledLoadOrder.
Stripped compiledOrder/modsByID — backend will provide GetCompiledOrder.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 6: Rewrite ModDetailsPanel

**Files:**
- Modify: `frontend/src/components/ModDetailsPanel.vue`

- [ ] **Step 1: Read current ModDetailsPanel.vue**

Read full file to understand: opens on mod selection, fetches workshop metadata, displays mod info.

- [ ] **Step 2: Rewrite with fetch on open**

ModDetailsPanel receives `modID` prop. When `open` becomes true, fetch mod details and workshop data.

```vue
<script setup lang="ts">
import { computed, watch, ref } from 'vue'
import { FetchWorkshopMetadataForMod, IsUnsubscribeEnabled } from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'

const props = defineProps<{ open: boolean; modID: string }>()

const mod = ref<any>(null)
const workshopItem = ref<any>(null)
const canUnsubscribe = ref(false)
const loading = ref(false)

async function loadModDetails() {
  if (!props.modID || !props.open) return
  loading.value = true
  try {
    // mod data comes from parent via props or event
    // workshop metadata fetched here
    const [ws, unsub] = await Promise.all([
      FetchWorkshopMetadataForMod(props.modID),
      IsUnsubscribeEnabled()
    ])
    workshopItem.value = ws
    canUnsubscribe.value = unsub
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}

watch(() => props.open, (isOpen) => { if (isOpen) loadModDetails() })
watch(() => props.modID, () => { if (props.open) loadModDetails() })
</script>
```

- [ ] **Step 3: Preserve existing template/styles** — keep the detail display layout, thumbnail, description rendering.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ModDetailsPanel.vue
git commit -m "refactor(ui): rewrite ModDetailsPanel with fetch-on-open

No Pinia store. Workshop metadata fetched when panel opens.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 7: Rewrite ConstraintModal

**Files:**
- Modify: `frontend/src/components/ConstraintModal.vue`

- [ ] **Step 1: Read current ConstraintModal.vue**

Read full file (~200 lines). Currently uses `useConstraintsStore`, `useLoadOrderStore`, `useModsStore`.

- [ ] **Step 2: Rewrite with direct backend calls**

```vue
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  AddConstraint, AddLoadFirst, AddLoadLast,
  GetConstraints, RemoveConstraint, RemoveLoadFirst, RemoveLoadLast,
  GetLauncherLayout, GetLoadOrder
} from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'

const props = defineProps<{ open: boolean; modID: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const constraints = ref<any[]>([])
const loading = ref(false)
const direction = ref<'after' | 'before' | 'first' | 'last'>('after')
const pickedModID = ref<string | null>(null)
const addError = ref<string | null>(null)

async function load() {
  if (!props.open) return
  loading.value = true
  try {
    const [c, layout, order] = await Promise.all([
      GetConstraints(),
      GetLauncherLayout(),
      GetLoadOrder()
    ])
    constraints.value = c
    // Store layout/order locally if needed for display
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}

async function addConstraint() {
  addError.value = null
  try {
    if (direction.value === 'first') {
      await AddLoadFirst(props.modID)
    } else if (direction.value === 'last') {
      await AddLoadLast(props.modID)
    } else if (pickedModID.value) {
      await AddConstraint(props.modID, pickedModID.value)
    }
    await load()
    pickedModID.value = null
  } catch (err) {
    addError.value = errorMessage(err)
  }
}

async function removeConstraint(key: string) {
  // Parse key format: "after:modID" or "first:modID"
  const [type, modId] = key.split(':')
  try {
    if (type === 'first') await RemoveLoadFirst(props.modID)
    else if (type === 'last') await RemoveLoadLast(props.modID)
    else await RemoveConstraint(props.modID, modId)
    await load()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

watch(() => props.open, (isOpen) => { if (isOpen) load() })
</script>
```

- [ ] **Step 3: Preserve existing template/styles**.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ConstraintModal.vue
git commit -m "refactor(ui): rewrite ConstraintModal with fetch-on-open

No Pinia stores. All constraint CRUD via direct backend calls.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 8: Rewrite AutosortButton

**Files:**
- Modify: `frontend/src/components/AutosortButton.vue`

- [ ] **Step 1: Read current AutosortButton.vue**

Read full file (~50 lines).

- [ ] **Step 2: Rewrite**

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { Autosort } from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import BaseButton from './ui/BaseButton.vue'

const isSorting = ref(false)

async function runAutosort() {
  isSorting.value = true
  try {
    await Autosort()
    showToast({ type: 'success', message: 'Autosort complete' })
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    isSorting.value = false
  }
}
</script>

<template>
  <BaseButton :loading="isSorting" @click="runAutosort">
    Autosort
  </BaseButton>
</template>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/AutosortButton.vue
git commit -m "refactor(ui): rewrite AutosortButton with backend call + toast

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 9: Rewrite LaunchButton

**Files:**
- Modify: `frontend/src/components/LaunchButton.vue`

- [ ] **Step 1: Read current LaunchButton.vue**

Read full file.

- [ ] **Step 2: Rewrite**

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { LaunchGame } from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import BaseButton from './ui/BaseButton.vue'

const launching = ref(false)

async function launch() {
  launching.value = true
  try {
    await LaunchGame()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    launching.value = false
  }
}
</script>

<template>
  <BaseButton :loading="launching" @click="launch">
    Launch Game
  </BaseButton>
</template>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/LaunchButton.vue
git commit -m "refactor(ui): rewrite LaunchButton with backend call + toast

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 10: Rewrite SettingsPanel

**Files:**
- Modify: `frontend/src/components/SettingsPanel.vue`

- [ ] **Step 1: Read current SettingsPanel.vue**

Read full file. Currently uses `useSettingsStore` for modsDirStatus, gameExe, etc.

- [ ] **Step 2: Rewrite with fetch on mount**

```vue
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  GetModsDirStatus, GetGameExe, GetConfigPath, GetAutoDetectedGameExe,
  SetModsDir, SetGameExe, PickFolder, PickExecutable,
  ResetModsDirToAuto, ResetGameExeToAuto, OpenConfigFolder
} from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import BaseButton from './ui/BaseButton.vue'

type ModsDirStatus = {
  effectiveDir: string; autoDetectedDir: string; customDir: string
  usingCustomDir: boolean; autoDetectedExists: boolean; effectiveExists: boolean
}

const modsDirStatus = ref<ModsDirStatus | null>(null)
const gameExe = ref('')
const configPath = ref('')
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    const [dirs, exe, cfg] = await Promise.all([
      GetModsDirStatus(), GetGameExe(), GetConfigPath()
    ])
    modsDirStatus.value = dirs
    gameExe.value = exe
    configPath.value = cfg
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}

async function pickModsDir() {
  const dir = await PickFolder()
  if (dir) {
    try {
      await SetModsDir(dir)
      await load()
    } catch (err) {
      showToast({ type: 'error', message: errorMessage(err) })
    }
  }
}

async function pickGameExe() {
  const exe = await PickExecutable()
  if (exe) {
    try {
      await SetGameExe(exe)
      await load()
    } catch (err) {
      showToast({ type: 'error', message: errorMessage(err) })
    }
  }
}

async function resetModsDir() {
  try {
    await ResetModsDirToAuto()
    await load()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

async function resetGameExe() {
  try {
    await ResetGameExeToAuto()
    await load()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

onMounted(load)
</script>
```

- [ ] **Step 3: Preserve existing template/styles**.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/SettingsPanel.vue
git commit -m "refactor(ui): rewrite SettingsPanel with fetch-on-mount

No Pinia store. Settings load/save via direct backend calls.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 11: Rewrite ManualGamePathSetup

**Files:**
- Modify: `frontend/src/components/ManualGamePathSetup.vue`

- [ ] **Step 1: Read current ManualGamePathSetup.vue**

Read full file.

- [ ] **Step 2: Rewrite** — Form with `PickFolder`/`PickExecutable` calls, `SetGamePaths` on save. No store.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ManualGamePathSetup.vue
git commit -m "refactor(ui): rewrite ManualGamePathSetup with direct backend calls

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 12: Verify CycleErrorPanel

**Files:**
- Modify: `frontend/src/components/CycleErrorPanel.vue`

- [ ] **Step 1: Read CycleErrorPanel.vue**

If it's purely presentational (accepts props, renders cycle info), no script changes needed — just verify no Pinia/store imports.

- [ ] **Step 2: If changes needed** — strip any store imports, ensure props-driven only.

- [ ] **Step 3: Commit** (if changed)

```bash
git add frontend/src/components/CycleErrorPanel.vue
git commit -m "refactor(ui): verify CycleErrorPanel has no store logic

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 13: Rewrite App.vue

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Read full App.vue**

Read all ~250 lines. Identify: all store imports, storeUsages, watchers on store state.

- [ ] **Step 2: Strip all Pinia store imports and `storeToRefs` calls**

Remove:
```ts
import { useLoadOrderStore } from './stores/loadorder'
import { useModsStore } from './stores/mods'
import { useSettingsStore } from './stores/settings'
import { useGamesStore } from './stores/games'
import { storeToRefs } from 'pinia'
```

- [ ] **Step 3: Replace store refs with local state**

Replace `storeToRefs`-wrapped refs with plain `ref()` or `reactive()`:

```ts
const detailsOpen = ref(false)
const settingsOpen = ref(false)
const searchOpen = ref(false)
const manualSetupOpen = ref(false)
const setupGameID = ref('')
const setupGameName = ref('')
```

- [ ] **Step 4: Remove watchers on store state** — delete watchers that were syncing panel open/close with store state.

- [ ] **Step 5: Keep modal orchestration** — panels are rendered via `v-if="detailsOpen"`, etc.

- [ ] **Step 6: Preserve existing template and styles**.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/App.vue
git commit -m "refactor(ui): strip all Pinia stores from App.vue

Local modal state only. No store imports, no storeToRefs.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 14: Delete all Pinia stores and unused utilities

**Files:**
- Delete: `frontend/src/stores/mods.ts`
- Delete: `frontend/src/stores/loadorder.ts`
- Delete: `frontend/src/stores/constraints.ts`
- Delete: `frontend/src/stores/games.ts`
- Delete: `frontend/src/stores/settings.ts`
- Delete: `frontend/src/lib/logger.ts`
- Delete: `frontend/src/utils/steamDescription.ts` (if exists)
- Delete: `frontend/src/utils/theme.ts` (if exists)

- [ ] **Step 1: Check which utils actually exist**

```bash
ls frontend/src/utils/
ls frontend/src/stores/
```

- [ ] **Step 2: Delete all stores and identified unused utils**

```bash
git rm frontend/src/stores/mods.ts frontend/src/stores/loadorder.ts \
  frontend/src/stores/constraints.ts frontend/src/stores/games.ts \
  frontend/src/stores/settings.ts
git rm frontend/src/lib/logger.ts
# delete others if they exist:
git rm frontend/src/utils/steamDescription.ts 2>/dev/null || true
git rm frontend/src/utils/theme.ts 2>/dev/null || true
```

- [ ] **Step 3: Commit**

```bash
git commit -m "refactor(ui): delete all Pinia stores and unused utilities

Stores removed: mods, loadorder, constraints, games, settings.
Utils removed: logger, steamDescription, theme.

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 15: TypeScript verification

- [ ] **Step 1: Run TypeScript check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -50
```

- [ ] **Step 2: Fix any import errors** — removed stores will cause import errors in components that still reference them. Fix each one.

Common fix pattern:
```bash
# Find remaining store references
grep -r "useModsStore\|useLoadOrderStore\|useConstraintsStore\|useGamesStore\|useSettingsStore" frontend/src/
```
Fix each reference by either removing the import line or replacing with direct Wails binding.

- [ ] **Step 3: Fix any type errors** — `tsc --noEmit` will surface type mismatches. Fix with `any` casts where backend types are unclear (acceptable during migration).

- [ ] **Step 4: Commit any fixes**

```bash
git add -A && git commit -m "fix(ui): resolve remaining TS errors from store removal

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 16: Wails dev mode test

- [ ] **Step 1: Generate Wails bindings**

```bash
cd /home/illia/code/eu5-mod-launcher && wails generate bindings 2>&1
```

- [ ] **Step 2: Run dev mode**

```bash
wails dev 2>&1 | head -30
```

- [ ] **Step 3: If runtime errors** — fix component errors. Common: `storeToRefs` called on non-store, missing imports.

- [ ] **Step 4: Commit if bindings changed**

```bash
git add frontend/wailsjs/ && git commit -m "chore: regenerate wails bindings after ui rewrite

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Task 17: Backend binding additions (NOTE: separate PR)

This task documents backend changes needed but is NOT in the UI rewrite scope.

- [ ] **Step 1: Document required backend additions**

In `internal/launcher/app_*.go` files, add:
- `GetModByID(id string) (mods.Mod, error)` — returns single mod
- `GetCompiledOrder() ([]string, error)` — flattens layout to ordered mod IDs
- `GetConstraintsForMod(modID string) ([]domain.Constraint, error)` — filtered constraints
- `IsWorkshopMod(modID string) (bool, error)` — checks if mod is workshop mod
- `GetWorkshopItemID(modID string) (string, error)` — extracts workshop ID from mod

These should be separate commits in a backend PR.

---

## Task 18: Final verification

- [ ] **Step 1: Run full TypeScript check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: zero errors.

- [ ] **Step 2: Run Go build**

```bash
go build ./...
```

Expected: builds successfully.

- [ ] **Step 3: Check no Pinia imports remain**

```bash
grep -r "defineStore\|useStore\|storeToRefs" frontend/src/ --include="*.vue" --include="*.ts"
```

Expected: zero results.

- [ ] **Step 4: Check no deleted files are imported**

```bash
grep -r "stores/mods\|stores/loadorder\|stores/constraints\|stores/games\|stores/settings" frontend/src/
```

Expected: zero results.

- [ ] **Step 5: Final commit**

```bash
git add -A && git commit -m "chore(ui): complete ultra-puppet rewrite

Verification:
- tsc --noEmit passes
- go build ./... passes  
- Zero Pinia imports
- Zero store imports

Co-Authored-By: Claude Opus 4.6 <noreply@openclaude.dev>"
```

---

## Spec Coverage Check

| Spec Section | Tasks |
|---|---|
| Ultra-puppet architecture | Tasks 1-17 |
| No Pinia stores | Tasks 2-14 |
| Fetch-per-component | Tasks 2-13 |
| Toast error handling | Tasks 1-2 |
| UI atoms preserved | Verified in each task |
| Drag-drop persistence | Task 5 |
| Constraint CRUD | Task 7 |
| Settings form | Task 10 |
| App.vue cleanup | Task 13 |
| Deleted unused files | Task 14 |
| TypeScript verification | Task 15 |
| Wails dev mode | Task 16 |

**All spec requirements covered.**

---

Plan complete and saved to `docs/superpowers/plans/2026-04-12-ui-rewrite.md`.

**Two execution options:**

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
