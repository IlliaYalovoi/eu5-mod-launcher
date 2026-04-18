# UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remake the UI to match the "The Chronicler" / "Parchment & Ink" mockup provided in `redesign.html`.

**Architecture:** We will convert the current 4-column layout (`App.vue`) into a 2-main-column layout (`Sidebar` and `Main`). `Sidebar` will absorb the game selector, playset selector (from `ModListPanel`), and launch button (from `LoadOrderPanel`). `Main` will contain the `LoadOrderPanel` (Mod Groups) and `ModListPanel` (Repository) in a responsive grid. Styles will be updated in `main.css` using CSS variables to match the mockup.

**Tech Stack:** Vue 3, Composition API (`<script setup lang="ts">`), CSS Variables, Pinia

---

### Task 1: Update Global CSS and CSS Variables

**Files:**
- Modify: `frontend/src/assets/main.css`

- [ ] **Step 1: Replace CSS variables and base styles**

```css
@import url('https://fonts.googleapis.com/css2?family=Crimson+Pro:wght@400;600;700&family=Cinzel:wght@500;700&display=swap');

:root,
:root[data-theme='dark'] {
  /* EU5 / Caesar Theme (Parchment & Ink) */
  --color-bg-base: #1a1814;
  --color-bg-panel: #24211b; /* Sidebar */
  --color-bg-elevated: #2d2922; /* Card */
  --color-border: #3d382e;
  --color-border-strong: #b9935a;
  --color-text-primary: #d4cfc3;
  --color-text-secondary: #8c8578;
  --color-text-muted: #8c8578;
  --color-accent: #b9935a;
  --color-accent-glow: rgba(185, 147, 90, 0.2);
  --color-accent-hover: #8e6d3d;
  --color-danger: #9e3939;
  --color-success: #5c7c51;
  --color-disabled: #7f8692;
  --color-overlay: rgba(0, 0, 0, 0.7);
  --color-scrollbar-track: rgba(26, 24, 20, 0.7);
  --color-scrollbar-thumb: rgba(185, 147, 90, 0.5);
  --color-scrollbar-thumb-hover: rgba(185, 147, 90, 0.8);
  --color-scrollbar-thumb-border: rgba(36, 33, 27, 0.7);

  /* Legacy overrides */
  --color-legacy-html-bg: #1a1814;
  --color-legacy-plain-white: #d4cfc3;
  --color-legacy-text-dark: #333333;
  --color-legacy-input-bg: #1a1814;
  --color-legacy-input-bg-hover: #24211b;

  /* Typography */
  --font-display: 'Crimson Pro', 'Georgia', serif;
  --font-body: 'Crimson Pro', 'Georgia', serif;
}

body {
  margin: 0;
  color: var(--color-text-primary);
  font-family: var(--font-body);
  background-image: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><rect fill="%231a1814" width="100" height="100"/><path d="M0 0l100 100M100 0L0 100" stroke="%23221f1a" stroke-width="0.5"/></svg>');
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/assets/main.css
git commit -m "style: update global css variables to match new aesthetic"
```

### Task 2: Refactor App Layout

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Update App.vue grid layout and classes**

```vue
<style scoped>
.shell {
  display: grid;
  /* Sidebar, Main Content, Details (if open) */
  grid-template-columns: 280px 1fr auto;
  grid-template-rows: 100vh;
  grid-template-areas:
    'sidebar content details';
  height: 100%;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
  overflow: hidden;
}

.titlebar {
  display: none; /* Removed, title moved to toolbar */
}

.sidebar {
  grid-area: sidebar;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-panel);
  border-right: 2px solid var(--color-border);
  box-shadow: 4px 0 15px rgba(0,0,0,0.5);
  overflow-y: auto;
}

.content {
  grid-area: content;
  display: flex;
  flex-direction: column;
  background: transparent;
  height: 100vh;
}

.details {
  grid-area: details;
  border-left: 2px solid var(--color-border);
  background: var(--color-bg-panel);
}

.toast {
  position: fixed;
  right: var(--space-5);
  bottom: var(--space-5);
  z-index: 1300;
  max-width: 24rem;
  padding: var(--space-3) var(--space-4);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

.toast--success {
  border-color: var(--color-success);
}

.toast--error {
  border-color: var(--color-danger);
  color: var(--color-danger);
}
</style>
```

- [ ] **Step 2: Update App.vue template structure**
Replace the layout structure with the new `aside.sidebar` and `main.content`.

```vue
<template>
  <div class="shell" :class="appThemeClass">
    <aside class="sidebar">
      <Sidebar />
    </aside>
    <main class="content" aria-label="Main content area">
      <LoadOrderPanel @contextmenu="openContextMenu" @open-constraints="openConstraintModal" />
    </main>
    <aside class="details">
      <ModDetailsPanel />
    </aside>
    <div v-if="unsubscribeNotice" class="toast" :class="`toast--${unsubscribeNotice.type}`" role="status" aria-live="polite">
      {{ unsubscribeNotice.message }}
    </div>
    <ContextMenu
      :open="contextMenu.open"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :items="contextMenuItems"
      :target-i-d="contextMenu.targetID"
      @close="closeContextMenu"
      @select="handleMenuAction"
    />
    <ConstraintModal :open="constraintModal.open" :mod-i-d="constraintModal.modID" @close="closeConstraintModal" />
    <BaseModal :open="settingsOpen" @close="closeSettings">
      <SettingsPanel :required="requiresManualPaths" @close="closeSettings" />
    </BaseModal>
  </div>
</template>
```
Remove any reference to `ModListPanel` in `App.vue` (it moves to `LoadOrderPanel`).

- [ ] **Step 3: Run tsc to verify types**
```bash
npx tsc --noEmit --project frontend/tsconfig.node.json && npx tsc --noEmit --project frontend/tsconfig.app.json
```

- [ ] **Step 4: Commit**
```bash
git add frontend/src/App.vue
git commit -m "refactor: convert app layout to match new mockup"
```

### Task 3: Refactor Sidebar Component

**Files:**
- Modify: `frontend/src/components/Sidebar.vue`

- [ ] **Step 1: Write the updated script and template**

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useSettingsStore } from '../stores/settings'
import { useLoadOrderStore } from '../stores/loadorder'
import GameSettingsModal from './GameSettingsModal.vue'
import LaunchButton from './LaunchButton.vue'

const settingsStore = useSettingsStore()
const loadOrderStore = useLoadOrderStore()

const { playsetNames, launcherActivePlaysetIndex, activeCountLabel } = storeToRefs(loadOrderStore)

const gameIcons: Record<string, string> = {
  eu5: '⚜️',
  hoi4: '🎖️',
  ck3: '👑',
  stellaris: '🚀',
  vic3: '🎩',
}

const gameNames: Record<string, string> = {
  eu5: 'Project Caesar',
  hoi4: 'Hearts of Iron IV',
  ck3: 'Crusader Kings III',
  stellaris: 'Stellaris',
  vic3: 'Victoria 3',
}

const gameSettingsModal = ref({
  open: false,
  gameID: '',
})

const isSwitchingPlayset = ref(false)

function selectGame(id: string) {
  settingsStore.setGame(id)
}

async function openGameSettings(id: string) {
  await settingsStore.setGame(id)
  gameSettingsModal.value = {
    open: true,
    gameID: id,
  }
}

function closeGameSettings() {
  gameSettingsModal.value.open = false
}

async function onLauncherPlaysetChange(event: Event) {
  const target = event.target as HTMLSelectElement
  const index = parseInt(target.value, 10)
  if (Number.isNaN(index)) return

  isSwitchingPlayset.value = true
  try {
    await loadOrderStore.setActivePlayset(index)
  } finally {
    isSwitchingPlayset.value = false
  }
}

const hasPlaysetChoices = computed(() => playsetNames.value.length > 0)
</script>

<template>
  <div class="sidebar-wrapper">
    <div class="sidebar-section header">
      <h2>MOD ORGANIZER</h2>
    </div>

    <nav class="game-nav sidebar-section">
      <div
        v-for="gameID in settingsStore.availableGames"
        :key="gameID"
        class="game-item"
      >
        <button
          class="game-btn"
          :class="{ active: settingsStore.activeGameID === gameID }"
          :title="(gameNames[gameID] || gameID.toUpperCase()) + ' (Right click for settings)'"
          @click="selectGame(gameID)"
          @contextmenu.prevent="openGameSettings(gameID)"
        >
          <span class="icon">{{ gameIcons[gameID] || '🎮' }}</span>
          <span>{{ gameNames[gameID] || gameID.toUpperCase() }}</span>
        </button>

        <div v-if="settingsStore.activeGameID === gameID" class="playset-selector">
          <select
            class="playset-dropdown"
            :disabled="!hasPlaysetChoices || isSwitchingPlayset"
            :value="launcherActivePlaysetIndex"
            @change="onLauncherPlaysetChange"
          >
            <option v-for="(name, index) in playsetNames" :key="`${name}-${index}`" :value="index">
              {{ name }}
            </option>
          </select>
        </div>
      </div>

      <GameSettingsModal
        v-if="gameSettingsModal.open"
        :open="gameSettingsModal.open"
        :game-i-d="gameSettingsModal.gameID"
        @close="closeGameSettings"
      />
    </nav>

    <div class="sidebar-footer">
      <div class="play-area">
        <div class="stats-mini">
          <span>{{ activeCountLabel }} Mods Active</span>
        </div>
        <LaunchButton />
        <div class="status-indicator">
          ● LOAD ORDER READY
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sidebar-section {
  padding: 20px;
  border-bottom: 1px solid var(--color-border);
}

.header h2 {
  font-size: 18px;
  color: var(--color-accent);
  margin: 0;
  font-family: var(--font-display);
}

.game-nav {
  flex: 1;
  overflow-y: auto;
}

.game-item {
  margin-bottom: 8px;
}

.game-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  border-radius: 4px;
  text-align: left;
  transition: all 0.2s;
  font-family: var(--font-body);
}

.game-btn.active {
  color: var(--color-text-primary);
  background: var(--color-bg-elevated);
  border-color: var(--color-accent);
  box-shadow: 0 0 10px var(--color-accent-glow);
}

.game-btn .icon {
  font-size: 1.2rem;
}

.playset-selector {
  margin-top: 10px;
  margin-left: 32px;
}

.playset-dropdown {
  width: 100%;
  background: var(--color-bg-base);
  color: var(--color-accent);
  border: 1px solid var(--color-border);
  padding: 8px;
  border-radius: 2px;
  font-size: 13px;
  outline: none;
  font-family: var(--font-body);
}

.sidebar-footer {
  margin-top: auto;
  padding: 20px;
  background: rgba(0,0,0,0.2);
  border-top: 2px solid var(--color-border);
}

.play-area {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stats-mini {
  font-size: 12px;
  color: var(--color-text-muted);
  display: flex;
  justify-content: space-between;
}

.status-indicator {
  font-size: 11px;
  text-align: center;
  color: var(--color-success);
  margin-top: 5px;
}
</style>
```

- [ ] **Step 2: Commit**
```bash
git add frontend/src/components/Sidebar.vue
git commit -m "feat: combine game and playset navigation in Sidebar"
```

### Task 4: Refactor LoadOrderPanel and integrate ModListPanel

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/ModListPanel.vue`

- [ ] **Step 1: Clean up ModListPanel.vue**
Remove the playset logic completely, just make it the Repository list.

```vue
<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import type { Mod } from '../types'
import ModCard from './ModCard.vue'
import SearchInput from './ui/SearchInput.vue'
import { useModsStore } from '../stores/mods'

const modsStore = useModsStore()
const { allMods, isLoading, error, selectedModID } = storeToRefs(modsStore)
const searchText = ref('')

const filteredMods = computed(() => {
  const query = searchText.value.trim().toLowerCase()
  if (!query) {
    return allMods.value
  }
  return allMods.value.filter((mod) => mod.Name.toLowerCase().includes(query))
})

const emptyMessage = computed(() => {
  if (allMods.value.length === 0) {
    return 'No mods were discovered.'
  }
  return 'No mods match your search query.'
})

function toggleMod(mod: Mod) {
  modsStore.setEnabled(mod.ID, !mod.Enabled)
}

function selectMod(mod: Mod) {
  modsStore.setSelectedModID(mod.ID)
}
</script>

<template>
  <aside class="repository">
    <div class="repo-title">Mod Repository (Disabled)</div>
    <SearchInput v-model="searchText" placeholder="Search unmanaged mods..." class="search-box" />

    <div class="list-body">
      <div v-if="isLoading" class="state loading">Loading mods...</div>
      <p v-else-if="error" class="state error">{{ error }}</p>
      <p v-else-if="filteredMods.length === 0" class="state empty">{{ emptyMessage }}</p>
      <div v-else class="cards">
        <ModCard
          v-for="mod in filteredMods"
          :key="mod.ID"
          :mod="mod"
          :selected="mod.ID === selectedModID"
          @toggle="toggleMod(mod)"
          @select="selectMod(mod)"
        />
      </div>
    </div>
    <div class="repo-footer">Items here are NOT in the current playset.</div>
  </aside>
</template>

<style scoped>
.repository {
  background: rgba(0,0,0,0.15);
  border: 1px dashed var(--color-border);
  padding: 20px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  height: 100%;
}

.repo-title {
  font-size: 13px;
  text-transform: uppercase;
  color: var(--color-text-muted);
  letter-spacing: 1px;
}

.search-box {
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  padding: 10px;
  color: var(--color-text-primary);
  border-radius: 4px;
  width: 100%;
  box-sizing: border-box;
}

.list-body {
  flex: 1;
  overflow-y: auto;
}

.cards {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.state {
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
  padding: 20px;
}

.error {
  color: var(--color-danger);
}

.repo-footer {
  text-align: center;
  color: var(--color-text-muted);
  font-size: 11px;
  margin-top: 10px;
}
</style>
```

- [ ] **Step 2: Update LoadOrderPanel.vue layout**
Include ModListPanel in LoadOrderPanel for the grid layout. Note the `import ModListPanel from './ModListPanel.vue'` addition to the top of `<script setup>`.

```vue
<template>
  <div class="main-wrapper">
    <header class="toolbar">
      <h1 class="playset-title">Load Order</h1>
      <div class="toolbar-actions">
        <AutosortButton />
        <BaseButton :loading="isSavingCompiled" @click="onSaveCompiled" class="action-btn">Save to Game</BaseButton>
      </div>
    </header>

    <div class="view-content">
      <div class="group-container">
        <!-- keep category list items as is but styled -->
        <div class="category-create">
          <input v-model="categoryName" class="category-input" type="text" placeholder="New category name..." />
          <BaseButton variant="ghost" @click="onCreateCategory">Create Category</BaseButton>
        </div>

        <p v-if="persistError" class="alert">{{ persistError }}</p>
        <p v-else-if="saveError" class="alert">{{ saveError }}</p>

        <div class="list-wrap">
          <draggable v-model="blocks" item-key="id" handle=".category-handle" :animation="150" @end="persistLayout">
            <template #item="{ element: block }">
              <section class="bucket category-block mod-group" @contextmenu="onItemContextMenu($event, block.id)">
                <div class="category-head group-header">
                  <div class="header-left">
                    <button class="category-handle" type="button" aria-label="Drag category">☰</button>
                    <strong>{{ block.name }}</strong>
                  </div>
                  <div class="header-actions">
                    <button class="fold" type="button" @click="onToggleCollapse(block.id)">{{ block.collapsed ? '+' : '-' }}</button>
                    <button v-if="!block.isUngrouped" class="delete-category" type="button" @click="onDeleteCategory(block.id)">×</button>
                  </div>
                </div>

                <draggable
                  v-if="!block.collapsed"
                  v-model="block.modIds"
                  :item-key="modItemKey"
                  :group="{ name: 'mods' }"
                  handle=".mod-handle"
                  :animation="150"
                  @end="persistLayout"
                  class="items mod-list"
                >
                  <template #item="{ element: modID }">
                    <LoadOrderItem
                      :mod-i-d="modID"
                      @contextmenu="onItemContextMenu($event, modID)"
                      @open-constraints="emit('open-constraints', modID)"
                    />
                  </template>
                </draggable>
              </section>
            </template>
          </draggable>
        </div>
      </div>

      <ModListPanel />
    </div>
  </div>
</template>
```

Add these styles to `LoadOrderPanel.vue`:
```css
.main-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.toolbar {
  padding: 20px 40px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--color-border);
}

.playset-title {
  margin: 0;
  font-size: 24px;
  font-family: var(--font-display);
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
}

.view-content {
  flex: 1;
  padding: 20px 40px;
  overflow-y: auto;
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 30px;
}

.group-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.mod-group {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  margin-bottom: 20px;
}

.group-header {
  padding: 12px 16px;
  background: rgba(255,255,255,0.03);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--color-border);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.list-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
```

- [ ] **Step 3: Commit**
```bash
git add frontend/src/components/LoadOrderPanel.vue frontend/src/components/ModListPanel.vue
git commit -m "feat: redesign main content view with dual-pane layout"
```

### Task 5: Refactor Launch Button

**Files:**
- Modify: `frontend/src/components/LaunchButton.vue`

- [ ] **Step 1: Update styles**
```vue
<style scoped>
.launch-btn {
  background: linear-gradient(to bottom, #b9935a, #8e6d3d);
  color: #1a1814;
  border: 1px solid #5a4623;
  padding: 14px;
  border-radius: 3px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 2px;
  cursor: pointer;
  box-shadow: 0 4px 0 #5a4623;
  width: 100%;
  text-align: center;
  transition: transform 0.1s, box-shadow 0.1s;
}

.launch-btn:active:not(:disabled) {
  transform: translateY(2px);
  box-shadow: 0 2px 0 #5a4623;
}

.launch-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  box-shadow: 0 4px 0 #3a2e18;
  background: linear-gradient(to bottom, #8a734e, #6e5532);
}

.loading {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { opacity: 0.8; }
  50% { opacity: 1; }
  100% { opacity: 0.8; }
}
</style>
```

- [ ] **Step 2: Commit**
```bash
git add frontend/src/components/LaunchButton.vue
git commit -m "style: update launch button to match chronicler theme"
```

### Task 6: Final Verification

- [ ] **Step 1: Check types and build**
```bash
cd frontend && npm run typecheck || npx tsc --noEmit
go build ./...
```
Expected: PASS

- [ ] **Step 2: Commit any lingering fixes**
```bash
git commit -am "fix: resolve any lingering ui bugs"
```