# UI Redesign Fixes Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the UI layout and UX according to new requirements (remove drag&drop, add togglers, popup mod details, settings gear icon, manage groups popup).

**Architecture:** 
1. Remove `vuedraggable` from `LoadOrderPanel.vue` and `LoadOrderItem.vue`. The load order is now just a list.
2. Add a `Toggle` component (the custom styled switch from `redesign.html`) and integrate it into `ModCard.vue` and `LoadOrderItem.vue`.
3. Extract `ModDetailsPanel.vue` from `App.vue` layout into a popup/modal using `BaseModal.vue`. Clicking anywhere on a mod will show the modal.
4. Add a cog icon `⚙️` to the sidebar header to open global settings instead of a titlebar button.
5. Create a `ManageGroupsModal.vue` to handle category creation/deletion, accessible via a "MANAGE GROUPS" button in the `LoadOrderPanel.vue` header.

**Tech Stack:** Vue 3, Composition API (`<script setup lang="ts">`), CSS Variables, Pinia

---

### Task 1: Add Settings Icon to Sidebar

**Files:**
- Modify: `frontend/src/components/Sidebar.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Add the settings button to the sidebar header**
In `Sidebar.vue`:
```vue
// In <script setup>:
import { ref, computed } from 'vue'
// Add defineEmits to talk to App.vue (which holds the BaseModal for SettingsPanel)
const emit = defineEmits<{
  (event: 'open-settings'): void
}>()

// In <template>:
    <div class="sidebar-section header">
      <h2>MOD ORGANIZER</h2>
      <button class="settings-btn" type="button" @click="emit('open-settings')" title="Settings">⚙️</button>
    </div>

// In <style scoped>:
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.settings-btn {
  background: transparent;
  border: none;
  color: var(--color-text-muted);
  font-size: 18px;
  cursor: pointer;
  padding: 4px;
  transition: color 0.2s;
}

.settings-btn:hover {
  color: var(--color-text-primary);
}
```

- [ ] **Step 2: Update App.vue to remove the hidden titlebar**
In `App.vue`:
```vue
// In <template>:
    <aside class="sidebar">
      <Sidebar @open-settings="openSettings" />
    </aside>

// Remove <header class="titlebar">...</header> completely from <template>
// Remove .titlebar {} completely from <style scoped>
```

- [ ] **Step 3: Commit**
```bash
git add frontend/src/components/Sidebar.vue frontend/src/App.vue
git commit -m "feat: add settings gear icon to sidebar and remove titlebar"
```

### Task 2: Create Toggle Component

**Files:**
- Create: `frontend/src/components/ui/ModToggle.vue`

- [ ] **Step 1: Create the Toggle component**
```vue
<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <div class="toggle" :class="{ on: modelValue }" @click.stop="toggle"></div>
</template>

<style scoped>
.toggle { 
  width: 36px; 
  height: 18px; 
  background: #444; 
  border-radius: 10px; 
  position: relative; 
  cursor: pointer; 
  flex-shrink: 0;
  transition: background 0.2s;
}
.toggle.on { 
  background: var(--color-success); 
}
.toggle::after { 
  content: ''; 
  position: absolute; 
  width: 14px; 
  height: 14px; 
  background: white; 
  border-radius: 50%; 
  top: 2px; 
  left: 2px; 
  transition: 0.2s; 
}
.toggle.on::after { 
  left: 20px; 
}
</style>
```

- [ ] **Step 2: Commit**
```bash
git add frontend/src/components/ui/ModToggle.vue
git commit -m "feat: create ModToggle component"
```

### Task 3: Add Toggles and Remove Drag&Drop from LoadOrderPanel

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/LoadOrderItem.vue`

- [ ] **Step 1: Clean up LoadOrderItem.vue**
Remove `.drag-handle`, `.mod-handle` and add `ModToggle`. Note that we fetch the mod object using `modsStore.getMod(props.modID)`.
```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useModsStore } from '../stores/mods'
import ModToggle from './ui/ModToggle.vue'

const props = defineProps<{
  modID: string
}>()

const emit = defineEmits<{
  (event: 'contextmenu', payload: { modID: string; x: number; y: number }): void
  (event: 'open-constraints', modID: string): void
  (event: 'select', modID: string): void
}>()

const modsStore = useModsStore()
const mod = computed(() => modsStore.getMod(props.modID))

function onContextMenu(event: MouseEvent): void {
  event.preventDefault()
  emit('contextmenu', {
    modID: props.modID,
    x: event.clientX,
    y: event.clientY,
  })
}

function toggleEnabled(value: boolean) {
  modsStore.setEnabled(props.modID, value)
}
</script>

<template>
  <div v-if="mod" class="mod-row" @contextmenu.prevent="onContextMenu" @click="emit('select', mod.ID)">
    <ModToggle :model-value="mod.Enabled" @update:model-value="toggleEnabled" />
    <span class="name">{{ mod.Name }}</span>
    <span class="version">v{{ mod.Version || '?' }}</span>
  </div>
</template>

<style scoped>
.mod-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.02);
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}

.mod-row:hover {
  background: rgba(255,255,255,0.05);
}

.name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.version {
  color: var(--color-text-muted);
  font-size: 11px;
}
</style>
```

- [ ] **Step 2: Clean up LoadOrderPanel.vue**
Replace `vuedraggable` with standard `v-for` loops. We remove `draggable` imports and usage entirely.

```vue
<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import AutosortButton from './AutosortButton.vue'
import CycleErrorPanel from './CycleErrorPanel.vue'
import LoadOrderItem from './LoadOrderItem.vue'
import ModListPanel from './ModListPanel.vue'
import BaseButton from './ui/BaseButton.vue'
import { useLoadOrderStore } from '../stores/loadorder'
import { useModsStore } from '../stores/mods'

// ... keep existing emits and refs, EXCEPT for the draggable ones. ...
// NOTE: Remove `import draggable from 'vuedraggable'`

const loadOrderStore = useLoadOrderStore()
const modsStore = useModsStore()
// ... existing state ...

function selectMod(modID: string) {
  modsStore.setSelectedModID(modID)
  // We'll hook this up to open the modal in the next task
}
</script>

<template>
  <div class="main-wrapper">
    <header class="toolbar">
      <h1 class="playset-title">Load Order</h1>
      <div class="toolbar-actions">
        <BaseButton variant="ghost" class="action-btn" @click="emit('manage-groups')">MANAGE GROUPS</BaseButton>
        <AutosortButton />
      </div>
    </header>

    <div class="view-content">
      <div class="group-container">
        <CycleErrorPanel @open-constraints="emit('open-constraints', $event)" />

        <p v-if="persistError" class="alert">{{ persistError }}</p>
        <p v-else-if="saveError" class="alert">{{ saveError }}</p>

        <div class="list-wrap">
          <section v-for="block in blocks" :key="block.id" class="mod-group" @contextmenu="onItemContextMenu($event, block.id)">
            <div class="group-header">
              <div class="header-left">
                <strong>{{ block.name }}</strong>
                <span v-if="block.isUngrouped" class="group-rule">Priority: Default</span>
                <span v-else class="group-rule">Priority: {{ block.id }}</span>
              </div>
              <div class="header-actions">
                <button class="fold" type="button" @click="onToggleCollapse(block.id)">{{ block.collapsed ? '+' : '-' }}</button>
              </div>
            </div>

            <div v-if="!block.collapsed" class="mod-list">
              <LoadOrderItem
                v-for="modID in block.modIds"
                :key="modID"
                :mod-i-d="modID"
                @contextmenu="onItemContextMenu($event, modID)"
                @open-constraints="emit('open-constraints', modID)"
                @select="selectMod"
              />
            </div>
          </section>
        </div>
      </div>

      <ModListPanel />
    </div>
  </div>
</template>

<style scoped>
/* Keep existing styles, add: */
.group-rule {
  font-size: 11px;
  color: var(--color-accent);
  text-transform: uppercase;
  margin-left: 10px;
}
.toolbar-actions {
  display: flex;
  gap: 20px;
}
.action-btn {
  background: transparent;
  border: 1px solid var(--color-accent);
  color: var(--color-accent);
  padding: 5px 15px;
  cursor: pointer;
  text-transform: uppercase;
  font-family: var(--font-body);
  font-size: 13px;
}
</style>
```

- [ ] **Step 3: Add Toggles to ModListPanel.vue (Repository)**
Update `ModCard.vue` to remove the old checkbox and use `ModToggle.vue` aligned to the right.

```vue
<!-- ModCard.vue -->
<script setup lang="ts">
import type { Mod } from '../types'
import ModToggle from './ui/ModToggle.vue'

// ... existing props and emits ...

function onToggle(value: boolean): void {
  emit('toggle', value)
}
</script>

<template>
  <div class="disabled-mod" :class="{ selected: props.selected }" @click="onSelect">
    <span class="name">{{ props.mod.Name }}</span>
    <ModToggle :model-value="props.mod.Enabled" @update:model-value="onToggle" />
  </div>
</template>

<style scoped>
.disabled-mod {
  padding: 8px 12px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: 2px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  opacity: 0.7;
  cursor: pointer;
  transition: opacity 0.2s, background 0.2s;
}
.disabled-mod:hover {
  opacity: 1;
  background: rgba(255,255,255,0.05);
}
.name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
```
In `ModListPanel.vue`, update `@toggle="toggleMod(mod)"` to `@toggle="(val) => toggleMod(mod, val)"` and change `toggleMod` to take `val`.

- [ ] **Step 4: Commit**
```bash
git add frontend/src/components/LoadOrderPanel.vue frontend/src/components/LoadOrderItem.vue frontend/src/components/ModListPanel.vue frontend/src/components/ModCard.vue
git commit -m "feat: remove drag and drop, add toggles to load order and repo"
```

### Task 4: Extract Mod Details into a Modal

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/ModDetailsPanel.vue`

- [ ] **Step 1: Wrap ModDetailsPanel inside a BaseModal in App.vue**
```vue
// In App.vue <script setup>:
import { computed } from 'vue'
// Add ref to track if details modal is open
const detailsOpen = computed(() => !!modsStore.selectedModID)

function closeDetails() {
  modsStore.setSelectedModID(null)
}
```

```vue
// In App.vue <template>:
    <!-- Remove <aside class="details"> -->
    <BaseModal :open="detailsOpen" @close="closeDetails">
      <div class="modal-content-wrapper">
        <ModDetailsPanel />
      </div>
    </BaseModal>

// In App.vue <style scoped>:
/* Update grid to be just 2 columns: 280px and 1fr */
.shell {
  display: grid;
  grid-template-columns: 280px 1fr;
  grid-template-rows: 100vh;
  grid-template-areas:
    'sidebar content';
  height: 100%;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
  overflow: hidden;
}

.modal-content-wrapper {
  background: var(--color-bg-panel);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  width: 600px;
  max-width: 90vw;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 30px rgba(0,0,0,0.5);
}
```

- [ ] **Step 2: Update BaseModal.vue to blur background**
In `BaseModal.vue` styles, ensure the backdrop has a blur effect:
```css
.backdrop {
  /* ... existing styles ... */
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}
```

- [ ] **Step 3: Commit**
```bash
git add frontend/src/App.vue frontend/src/components/ui/BaseModal.vue
git commit -m "feat: convert mod details into a modal popup"
```

### Task 5: Create Manage Groups Modal

**Files:**
- Create: `frontend/src/components/ManageGroupsModal.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Create ManageGroupsModal.vue**
```vue
<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useLoadOrderStore } from '../stores/loadorder'
import BaseModal from './ui/BaseModal.vue'
import BaseButton from './ui/BaseButton.vue'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (event: 'close'): void
}>()

const loadOrderStore = useLoadOrderStore()
const { launcherLayout } = storeToRefs(loadOrderStore)

const newCategoryName = ref('')
const error = ref('')

async function createCategory() {
  if (!newCategoryName.value.trim()) return
  try {
    await loadOrderStore.createCategory(newCategoryName.value.trim())
    newCategoryName.value = ''
  } catch (e: any) {
    error.value = e.message
  }
}

async function deleteCategory(id: string) {
  try {
    await loadOrderStore.deleteCategory(id)
  } catch (e: any) {
    error.value = e.message
  }
}
</script>

<template>
  <BaseModal :open="open" @close="emit('close')">
    <div class="manage-groups-modal">
      <h2>Manage Mod Groups</h2>
      
      <div class="create-form">
        <input v-model="newCategoryName" type="text" placeholder="New group name..." class="input" @keyup.enter="createCategory" />
        <BaseButton @click="createCategory">Add Group</BaseButton>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <ul class="group-list">
        <li v-for="category in launcherLayout.categories" :key="category.id" class="group-item">
          <span>{{ category.name }}</span>
          <button class="delete-btn" type="button" @click="deleteCategory(category.id)">Delete</button>
        </li>
      </ul>
      
      <div class="actions">
        <BaseButton variant="ghost" @click="emit('close')">Done</BaseButton>
      </div>
    </div>
  </BaseModal>
</template>

<style scoped>
.manage-groups-modal {
  background: var(--color-bg-panel);
  padding: 30px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  width: 400px;
}

h2 {
  margin-top: 0;
  color: var(--color-accent);
  font-family: var(--font-display);
}

.create-form {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.input {
  flex: 1;
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  padding: 8px;
  border-radius: 4px;
}

.group-list {
  list-style: none;
  padding: 0;
  margin: 0 0 20px 0;
  max-height: 300px;
  overflow-y: auto;
}

.group-item {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
}

.delete-btn {
  background: transparent;
  border: none;
  color: var(--color-danger);
  cursor: pointer;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.error {
  color: var(--color-danger);
  font-size: 13px;
  margin-bottom: 10px;
}
</style>
```

- [ ] **Step 2: Add ManageGroupsModal to App.vue**
```vue
// In App.vue <script setup>:
import ManageGroupsModal from './components/ManageGroupsModal.vue'

const manageGroupsOpen = ref(false)

// In App.vue <template>, hook it up to LoadOrderPanel
      <LoadOrderPanel @contextmenu="openContextMenu" @open-constraints="openConstraintModal" @manage-groups="manageGroupsOpen = true" />
      
// Add to bottom of template:
    <ManageGroupsModal :open="manageGroupsOpen" @close="manageGroupsOpen = false" />
```
Make sure `LoadOrderPanel.vue` emits `manage-groups` from the button we added in Task 3.

- [ ] **Step 3: Check types and build**
```bash
npx tsc --noEmit --project frontend/tsconfig.node.json && npx tsc --noEmit --project frontend/tsconfig.app.json
go build ./...
```
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add frontend/src/components/ManageGroupsModal.vue frontend/src/App.vue
git commit -m "feat: move category management to a dedicated modal popup"
```