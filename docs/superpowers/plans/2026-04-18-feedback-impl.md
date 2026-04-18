# Feedback Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement UX improvements, fix scrolling, fix Mod list filtering, and update compatibility visual indicators.

**Architecture:** Update `IsVersionCompatible` on backend to handle empty supported version. Fix flexbox properties for `ModListPanel.vue` and `LoadOrderPanel.vue`. Filter store usage to segregate enabled/disabled mods. Expose and use `SupportedVersion`. Update `GameSettingsModal.vue` with an Auto Detect button.

**Tech Stack:** Go, Vue 3, TypeScript, Tailwind/CSS Variables

---

### Task 1: Backend Compatibility Logic Update

**Files:**
- Modify: `internal/service/mods_service.go`

- [ ] **Step 1: Treat empty supported version as ANY**
Modify `internal/service/mods_service.go` function `IsVersionCompatible`. If `supportedVersion == ""` return `true` instead of `false`.

```go
func IsVersionCompatible(gameVersion, supportedVersion string) bool {
	if supportedVersion == "" {
		return true // Treat empty supported version as ANY
	}
	if gameVersion == "unknown" {
		return false
	}
	if gameVersion == supportedVersion {
		return true
	}
	prefix := strings.ReplaceAll(supportedVersion, "*", "")
	return strings.HasPrefix(gameVersion, prefix)
}
```

- [ ] **Step 2: Test the backend logic**
Run: `go build ./...`
Run: `go test ./...`

- [ ] **Step 3: Commit**
```bash
git add internal/service/mods_service.go
git commit -m "fix(backend): treat empty supported version as compatible"
```

### Task 2: Auto Detect Button

**Files:**
- Modify: `frontend/src/components/GameSettingsModal.vue`
- Modify: `frontend/src/stores/settings.ts`

- [ ] **Step 1: Add Auto Detect button next to input**
Modify `frontend/src/components/GameSettingsModal.vue` to add a button. Use flexbox to align input and button horizontally.
```vue
<div class="mt-4">
  <label class="block text-sm font-medium text-surface-400 mb-1">Game Version (Override)</label>
  <div class="flex gap-2">
    <input 
      type="text" 
      v-model="settingsStore.getGameSettings(activeGame).gameVersionOverride"
      class="w-full bg-surface-800 border border-surface-700 rounded-md px-3 py-2 text-surface-200 focus:outline-none focus:border-primary-500"
      placeholder="e.g. 1.37.5"
    />
    <button 
      class="px-3 py-2 bg-surface-700 hover:bg-surface-600 border border-surface-600 rounded-md text-sm whitespace-nowrap"
      @click="() => settingsStore.setGameVersionOverride('')"
    >
      Auto Detect
    </button>
  </div>
</div>
```

- [ ] **Step 2: Ensure `setGameVersionOverride` exists or add it**
Check `frontend/src/stores/settings.ts`. It has `setGameVersionOverride(version: string)`. But wait, `GameSettingsModal.vue` accesses `getGameSettings(activeGame).gameVersionOverride` directly via `v-model`. 
To ensure the override clears properly, simply update the reactive state in the click handler:
```vue
    <button 
      class="px-3 py-2 bg-surface-700 hover:bg-surface-600 border border-surface-600 rounded-md text-sm whitespace-nowrap text-surface-200"
      @click="() => { settingsStore.getGameSettings(activeGame).gameVersionOverride = '' }"
    >
      Auto Detect
    </button>
```

- [ ] **Step 3: Test build**
Run: `cd frontend && npm run build` (This ensures typescript types are strictly okay)

- [ ] **Step 4: Commit**
```bash
git add frontend/src/components/GameSettingsModal.vue
git commit -m "feat(ui): add auto detect button to clear version override"
```

### Task 3: Fix Scrolling and List Filtering

**Files:**
- Modify: `frontend/src/components/ModListPanel.vue`
- Modify: `frontend/src/components/LoadOrderPanel.vue`

- [ ] **Step 1: Filter disabled mods in ModListPanel**
Modify `frontend/src/components/ModListPanel.vue` `filteredMods` computed property to only return disabled mods.
```typescript
const filteredMods = computed(() => {
  const query = searchText.value.trim().toLowerCase()
  const disabled = allMods.value.filter(mod => !mod.Enabled)
  if (!query) {
    return disabled
  }
  return disabled.filter((mod) => mod.Name.toLowerCase().includes(query))
})
```

- [ ] **Step 2: Fix scrolling in ModListPanel**
Ensure `.list-body` has `min-height: 0`.
Modify `<style>` in `frontend/src/components/ModListPanel.vue`:
```css
.list-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
```

- [ ] **Step 3: Filter enabled mods in LoadOrderPanel and fix scrolling**
Modify `frontend/src/components/LoadOrderPanel.vue`. When mapping `blocks`, filter `category.modIds` and `value.ungrouped` to only include enabled mods.
```typescript
      if (id === ungroupedID) {
        next.push({
          id: ungroupedID,
          name: 'Ungrouped',
          modIds: value.ungrouped.filter(modID => {
            const mod = allMods.value.find(m => m.ID === modID)
            return mod && mod.Enabled
          }),
          isUngrouped: true,
          collapsed: !!collapsed[ungroupedID],
        })
        continue
      }

      const category = categoryByID[id]
      if (!category) continue
      next.push({
        id: category.id,
        name: category.name,
        modIds: category.modIds.filter(modID => {
          const mod = allMods.value.find(m => m.ID === modID)
          return mod && mod.Enabled
        }),
        isUngrouped: false,
        collapsed: !!collapsed[category.id],
      })
```
Fix scrolling in `<style>`:
```css
.view-content {
  flex: 1;
  min-height: 0;
  padding: 20px 40px;
  overflow: hidden;
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 30px;
}
```
*(Also ensure `.group-container` already has `overflow-y: auto`, which it does based on previous checks)*

- [ ] **Step 4: Test build**
Run: `cd frontend && npm run build`

- [ ] **Step 5: Commit**
```bash
git add frontend/src/components/ModListPanel.vue frontend/src/components/LoadOrderPanel.vue
git commit -m "fix(ui): scroll areas and separate enabled/disabled mods correctly"
```

### Task 4: Mod Version / Compatibility Indicators

**Files:**
- Modify: `frontend/src/components/LoadOrderItem.vue`
- Modify: `frontend/src/components/ModCard.vue`
- Modify: `frontend/src/components/ModDetailsPanel.vue`

- [ ] **Step 1: Update LoadOrderItem.vue**
Show `mod.SupportedVersion` (with 'ANY' fallback) and add compatibility warning icon.
```vue
    <span class="name">{{ mod.Name }}</span>
    <span class="version" style="display: flex; gap: 4px; align-items: center;">
      v{{ mod.SupportedVersion || 'ANY' }}
      <span v-if="!mod.IsCompatible" class="text-yellow-500" title="Incompatible game version">⚠️</span>
    </span>
```

- [ ] **Step 2: Update ModCard.vue**
Change `mod.Version` display to `mod.SupportedVersion` (with 'ANY' fallback).
```vue
    <div class="flex items-center gap-2">
      <span class="text-xs text-surface-400">v{{ props.mod.SupportedVersion || 'ANY' }}</span>
      <span v-if="!props.mod.IsCompatible" class="text-yellow-500" title="Incompatible game version">
        ⚠️
      </span>
    </div>
```
*(Wait, `mod.IsCompatible` and `mod.SupportedVersion` use `props.mod.xxx` in the template depending on the component's internal property name. Adjust accurately based on existing file structure)*

- [ ] **Step 3: Show Mod Version in Details**
Modify `frontend/src/components/ModDetailsPanel.vue`.
Locate the block displaying Mod details and ensure both the supported game version and actual mod version are visible.
Find the description paragraph or tags list and add a line/span:
```vue
        <p class="text-sm text-surface-400 mb-4">
          Mod Version: v{{ selectedMod.Version || 'Unknown' }}
        </p>
```

- [ ] **Step 4: Test build**
Run: `cd frontend && npm run build`

- [ ] **Step 5: Commit**
```bash
git add frontend/src/components/LoadOrderItem.vue frontend/src/components/ModCard.vue frontend/src/components/ModDetailsPanel.vue
git commit -m "feat(ui): display supported version on cards and add compatibility indicator to load order"
```
