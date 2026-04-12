# Task 05 — Frontend: Pinia Stores & Wails Bindings Integration

## Goal

Set up the Pinia stores that will be the single source of truth for the Vue app. Wire them to the auto-generated Wails Go bindings. No UI components in this task — only data layer.

## Context

After task 04, running `wails dev` generates TypeScript bindings in `frontend/src/wailsjs/go/main/App.ts`. This task consumes those bindings inside Pinia stores. All components will read from stores and call store actions — they never import Wails bindings directly.

## Deliverables

### `frontend/src/stores/mods.ts`

```ts
// Pinia store
export const useModsStore = defineStore('mods', () => {
  const allMods = ref<Mod[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function fetchAll(): Promise<void>     // calls GetAllMods()
  async function setEnabled(id: string, enabled: boolean): Promise<void>

  const enabledMods = computed(() =>           // derived: mods that are enabled
    allMods.value.filter(m => m.Enabled)
  )

  return { allMods, enabledMods, isLoading, error, fetchAll, setEnabled }
})
```

### `frontend/src/stores/loadorder.ts`

```ts
export const useLoadOrderStore = defineStore('loadorder', () => {
  const orderedIDs = ref<string[]>([])

  async function fetch(): Promise<void>        // calls GetLoadOrder()
  async function persist(ids: string[]): Promise<void>  // calls SetLoadOrder()
  async function autosort(): Promise<void>     // calls Autosort(), updates orderedIDs
  // autosort must surface cycle errors to the UI (store an autosortError ref)

  return { orderedIDs, fetch, persist, autosort, autosortError }
})
```

### `frontend/src/stores/constraints.ts`

```ts
export const useConstraintsStore = defineStore('constraints', () => {
  const constraints = ref<Constraint[]>([])

  async function fetch(): Promise<void>
  async function add(from: string, to: string): Promise<void>
  async function remove(from: string, to: string): Promise<void>

  // helper: get all constraints involving a given mod ID
  function forMod(id: string): Constraint[]

  return { constraints, fetch, add, remove, forMod }
})
```

### `frontend/src/stores/settings.ts`

```ts
export const useSettingsStore = defineStore('settings', () => {
  const modsDir = ref('')

  async function fetch(): Promise<void>         // calls GetModsDir()
  async function setModsDir(path: string): Promise<void>  // calls SetModsDir()

  return { modsDir, fetch, setModsDir }
})
```

### `frontend/src/types.ts`

Mirror the Go types as TypeScript interfaces:

```ts
export interface Mod {
  id: string
  name: string
  version: string
  tags: string[]
  description: string
  ThumbnailPath: string
  DirPath: string
  Enabled: boolean
}

export interface Constraint {
  from: string
  to: string
}
```

### `frontend/src/main.ts`

Bootstrap: install Pinia, fetch initial data from all stores after app mount.

## Acceptance criteria

- `useModsStore().fetchAll()` called from browser console populates `allMods` with data from Go.
- Errors from Go are caught and stored in `error` ref, not thrown to the console unhandled.
- No Wails binding imports exist anywhere except inside store files.

## Notes for agent

- Wails bindings live at `frontend/src/wailsjs/go/main/App.ts` — import from there.
- The auto-generated file may use slightly different casing — check the actual generated file before writing imports.
- Use `async/await` consistently — Wails bindings return Promises.
