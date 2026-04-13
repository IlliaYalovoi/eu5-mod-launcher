<script lang="ts" setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import type { LauncherLayout, Mod } from './types'
import {
  DisableMod,
  GetGameActivePlaysetIndex,
  GetLauncherActivePlaysetIndex,
  GetLauncherLayout,
  GetLoadOrder,
  GetAllMods,
  GetModsDirStatus,
  ListSupportedGames,
  SetActiveGame,
  SetLauncherActivePlaysetIndex,
  SetLoadOrder,
  SaveCompiledLoadOrder,
  UnsubscribeWorkshopMod,
  IsUnsubscribeEnabled,
} from './wailsjs/go/launcher/App'
import { showToast } from './lib/toast'
import { errorMessage } from './lib/error'
import LoadOrderPanel from './components/LoadOrderPanel.vue'
import ModDetailsPanel from './components/ModDetailsPanel.vue'
import ConstraintModal from './components/ConstraintModal.vue'
import ContextMenu from './components/ui/ContextMenu.vue'
import BaseModal from './components/ui/BaseModal.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import SearchOverlay from './components/ui/SearchOverlay.vue'
import ToastContainer from './components/ui/ToastContainer.vue'
import ManualGamePathSetup from './components/ManualGamePathSetup.vue'

type MenuItem = {
  id: string
  label: string
  icon?: string
  danger?: boolean
  disabled?: boolean
  children?: MenuItem[]
}

// Game sidebar state
const supportedGames = ref<Array<{ id: string; name: string; detected: boolean }>>([])
const activeGameID = ref('')
const requiresManualPaths = ref(false)

// Playset state
const playsetNames = ref<string[]>([])
const launcherActivePlaysetIndex = ref(-1)

// Load order state (used for context menu)
const orderedIDs = ref<string[]>([])
const launcherLayout = ref<LauncherLayout>({ ungrouped: [], categories: [], order: [], collapsed: {} })

// Mod detail state
const selectedMod = ref<Mod | null>(null)
const allMods = ref<Mod[]>([])

// Unsubscribe feature
const unsubscribeFeatureEnabled = ref(false)
const unsubscribeLoadingByModID = ref<Record<string, boolean>>({})

const ungroupedCategoryID = 'category:ungrouped'

// Slide-over panels
const detailsOpen = ref(false)
const settingsOpen = ref(false)
const searchOpen = ref(false)
const manualSetupOpen = ref(false)
const setupGameID = ref<string>('')
const setupGameName = ref<string>('')

const contextMenu = reactive({ open: false, x: 0, y: 0, targetID: '' })
const constraintModal = reactive({ open: false, modID: '' })

async function load() {
  try {
    const [games, gameIdx, layout, order, names, unsubEnabled, dirsStatus, mods] = await Promise.all([
      ListSupportedGames(),
      GetGameActivePlaysetIndex(),
      GetLauncherLayout(),
      GetLoadOrder(),
      GetPlaysetNames(),
      IsUnsubscribeEnabled(),
      GetModsDirStatus(),
      GetAllMods(),
    ])
    supportedGames.value = games.map((g: any) => ({ id: g.id, name: g.name, detected: g.detected }))
    const detected = supportedGames.value.find((g: any) => g.detected)
    activeGameID.value = detected?.id || supportedGames.value[0]?.id || ''
    launcherActivePlaysetIndex.value = gameIdx
    launcherLayout.value = layout as LauncherLayout
    orderedIDs.value = order
    playsetNames.value = names
    allMods.value = mods as Mod[]
    unsubscribeFeatureEnabled.value = unsubEnabled
    requiresManualPaths.value = !(dirsStatus as any).autoDetectedExists && !(dirsStatus as any).effectiveExists
    if (requiresManualPaths.value) settingsOpen.value = true
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

function handleGlobalKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    if (searchOpen.value) { searchOpen.value = false; return }
    if (detailsOpen.value) { detailsOpen.value = false; return }
    if (settingsOpen.value && !requiresManualPaths.value) { settingsOpen.value = false; return }
  }
  if ((event.ctrlKey || event.metaKey) && event.key === 'k') { event.preventDefault(); searchOpen.value = !searchOpen.value; return }
  if ((event.ctrlKey || event.metaKey) && event.key === 's') { event.preventDefault(); void SaveCompiledLoadOrder(); return }
  if ((event.ctrlKey || event.metaKey) && event.key === 'f') { event.preventDefault(); searchOpen.value = true; return }
}

onMounted(() => { window.addEventListener('keydown', handleGlobalKeydown); void load() })
onUnmounted(() => { window.removeEventListener('keydown', handleGlobalKeydown) })

const contextMenuItems = computed<MenuItem[]>(() => {
  const isCategory = contextMenu.targetID.indexOf('category:') === 0
  const currentCategoryForTarget = (() => {
    for (const cat of launcherLayout.value.categories) {
      if (cat.modIds.includes(contextMenu.targetID)) return cat.id
    }
    if (launcherLayout.value.ungrouped.includes(contextMenu.targetID)) return ungroupedCategoryID
    return ''
  })()

  const moveToCategoryItems: MenuItem[] = [
    { id: `move_to_category:${ungroupedCategoryID}`, label: 'Move to category: Ungrouped', icon: '📁', disabled: currentCategoryForTarget === ungroupedCategoryID },
    ...launcherLayout.value.categories.map((cat) => ({ id: `move_to_category:${cat.id}`, label: `Move to category: ${cat.name}`, icon: '📁', disabled: currentCategoryForTarget === cat.id })),
  ]
  const moveToCategoryMenu: MenuItem = { id: 'move_to_category_menu', label: 'Move to category', icon: '📁', disabled: moveToCategoryItems.every((item) => item.disabled), children: moveToCategoryItems }

  if (isCategory) return [{ id: 'add_constraint', label: 'Add constraint...', icon: '⛓' }, { id: 'view_constraints', label: 'View constraints', icon: '📖' }]

  const index = orderedIDs.value.indexOf(contextMenu.targetID)
  const missing = index < 0
  const atTop = index === 0
  const atBottom = index === orderedIDs.value.length - 1
  const unsubLoading = unsubscribeLoadingByModID.value[contextMenu.targetID] ?? false

  const items: MenuItem[] = [
    { id: 'add_constraint', label: 'Add constraint...', icon: '⛓' },
    { id: 'view_constraints', label: 'View constraints', icon: '📖' },
    moveToCategoryMenu,
    { id: 'move_top', label: 'Move to top', icon: '⬆', disabled: missing || atTop },
    { id: 'move_bottom', label: 'Move to bottom', icon: '⬇', disabled: missing || atBottom },
    { id: 'disable_mod', label: 'Remove from load order', icon: '⛔', danger: true, disabled: missing },
  ]
  if (unsubscribeFeatureEnabled.value) {
    items.splice(5, 0, { id: 'unsubscribe_workshop', label: unsubLoading ? 'Unsubscribing...' : 'Unsubscribe from Workshop...', icon: '🧹', danger: true, disabled: missing || unsubLoading })
  }
  return items
})

async function moveModToCategory(modID: string, categoryID: string): Promise<void> {
  const next: LauncherLayout = {
    ungrouped: [...launcherLayout.value.ungrouped],
    categories: launcherLayout.value.categories.map((cat) => ({ id: cat.id, name: cat.name, modIds: [...cat.modIds] })),
    order: launcherLayout.value.order ? [...launcherLayout.value.order] : undefined,
    collapsed: launcherLayout.value.collapsed ? { ...launcherLayout.value.collapsed } : undefined,
  }
  next.ungrouped = next.ungrouped.filter((id) => id !== modID)
  for (const cat of next.categories) cat.modIds = cat.modIds.filter((id) => id !== modID)
  if (categoryID === ungroupedCategoryID) next.ungrouped.push(modID)
  else { const t = next.categories.find((cat) => cat.id === categoryID); if (t) t.modIds.push(modID) }
  try {
    launcherLayout.value = next
    orderedIDs.value = [...next.ungrouped, ...next.categories.flatMap((c) => c.modIds)]
  } catch (err) { showToast({ type: 'error', message: errorMessage(err) }) }
}

function openContextMenu(event: { modID: string; x: number; y: number }): void { contextMenu.open = true; contextMenu.x = event.x; contextMenu.y = event.y; contextMenu.targetID = event.modID }
function closeContextMenu(): void { contextMenu.open = false }
function closeConstraintModal(): void { constraintModal.open = false; constraintModal.modID = '' }
function openSettings(): void { searchOpen.value = false; detailsOpen.value = false; settingsOpen.value = true }
function closeSettings(): void { if (!requiresManualPaths.value) settingsOpen.value = false }
function openConstraintModal(modID: string): void { constraintModal.modID = modID; constraintModal.open = true }
function openSearch(): void { searchOpen.value = true }
function closeSearch(): void { searchOpen.value = false }
function closeDetails(): void { detailsOpen.value = false }

async function onLauncherPlaysetChange(event: Event): Promise<void> {
  const target = event.target as HTMLSelectElement
  const index = parseInt(target.value, 10)
  if (index.toString() !== target.value || index === launcherActivePlaysetIndex.value) return
  try {
    await SetLauncherActivePlaysetIndex(index)
    launcherActivePlaysetIndex.value = index
  } catch (err) { showToast({ type: 'error', message: errorMessage(err) }) }
}

async function onGameClick(game: { id: string; name: string; detected: boolean }): Promise<void> {
  if (!game.detected) { setupGameID.value = game.id; setupGameName.value = game.name; manualSetupOpen.value = true; return }
  try {
    await SetActiveGame(game.id)
    activeGameID.value = game.id
  } catch (err) { showToast({ type: 'error', message: errorMessage(err) }) }
}

function onSearchAddMod(modID: string): void {
  searchOpen.value = false
  selectedMod.value = allMods.value.find(m => m.ID === modID) || null
  detailsOpen.value = true
}
function onModSelect(modID: string): void {
  if (modID === selectedMod.value?.ID && detailsOpen.value) { detailsOpen.value = false; return }
  searchOpen.value = false; settingsOpen.value = false
  selectedMod.value = allMods.value.find(m => m.ID === modID) || null
  detailsOpen.value = true
}

async function handleMenuAction(event: { itemID: string; targetID: string }): Promise<void> {
  const { itemID, targetID } = event
  const isCategory = targetID.indexOf('category:') === 0
  if (isCategory && itemID === 'disable_mod') return
  const current = [...orderedIDs.value]
  const index = current.indexOf(targetID)

  if (itemID === 'disable_mod') {
    try { await DisableMod(targetID); orderedIDs.value = current.filter((id) => id !== targetID) }
    catch (err) { showToast({ type: 'error', message: errorMessage(err) }) }
    return
  }
  if (itemID === 'unsubscribe_workshop') {
    unsubscribeLoadingByModID.value[targetID] = true
    try { await UnsubscribeWorkshopMod(targetID) }
    catch (err) { showToast({ type: 'error', message: errorMessage(err) }) }
    finally { unsubscribeLoadingByModID.value[targetID] = false }
    return
  }
  if (itemID.indexOf('move_to_category:') === 0) { await moveModToCategory(targetID, itemID.slice('move_to_category:'.length)); return }
  if (itemID === 'move_top' && index > 0) { current.splice(index, 1); current.unshift(targetID); try { await SetLoadOrder(current); orderedIDs.value = current } catch (err) { showToast({ type: 'error', message: errorMessage(err) }) } return }
  if (itemID === 'move_bottom' && index >= 0 && index < current.length - 1) { current.splice(index, 1); current.push(targetID); try { await SetLoadOrder(current); orderedIDs.value = current } catch (err) { showToast({ type: 'error', message: errorMessage(err) }) } return }
  if (itemID === 'add_constraint' || itemID === 'view_constraints') { openConstraintModal(targetID); return }
}
</script>

<template>
  <div class="shell">
    <header class="commandbar">
      <div class="commandbar-left">
        <span class="app-title">EU5 Mod Launcher</span>
        <div class="game-switcher" v-if="playsetNames.length > 0">
          <select
            class="game-select"
            :value="launcherActivePlaysetIndex"
            @change="onLauncherPlaysetChange"
          >
            <option v-for="(name, index) in playsetNames" :key="`${name}-${index}`" :value="index">
              {{ name }}
            </option>
          </select>
        </div>
      </div>
      <div class="commandbar-center">
        <button class="search-trigger" type="button" @click="openSearch">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.35-4.35" />
          </svg>
          <span>Search mods...</span>
          <kbd>Ctrl+K</kbd>
        </button>
      </div>
      <div class="commandbar-right">
        <button class="icon-btn" type="button" aria-label="Settings" @click="openSettings">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <circle cx="12" cy="12" r="3" />
            <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
          </svg>
        </button>
      </div>
    </header>

    <aside class="game-sidebar">
      <div class="sidebar-header">
        <h3 class="sidebar-title">Games</h3>
      </div>
      <nav class="sidebar-nav">
        <button
          v-for="game in supportedGames"
          :key="game.id"
          class="game-item"
          :class="{
            'game-item--detected': game.detected,
            'game-item--undetected': !game.detected,
            'game-item--active': activeGameID === game.id
          }"
          @click="onGameClick(game)"
        >
          <div class="game-item-icon">🎮</div>
          <div class="game-item-info">
            <div class="game-item-name">{{ game.name }}</div>
            <div v-if="!game.detected" class="game-item-status">Not detected</div>
          </div>
        </button>
      </nav>
    </aside>

    <main class="content" aria-label="Main content area">
      <LoadOrderPanel
        @contextmenu="openContextMenu"
        @open-constraints="openConstraintModal"
        @select-mod="onModSelect"
      />
    </main>

    <!-- Mod details slide-over -->
    <Transition name="slide-over">
      <aside v-if="detailsOpen" class="details-slide" aria-label="Mod details">
        <button class="details-close" type="button" aria-label="Close details" @click="closeDetails">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <path d="M18 6 6 18M6 6l12 12" />
          </svg>
        </button>
        <ModDetailsPanel :open="detailsOpen" :mod="selectedMod" />
      </aside>
    </Transition>

    <!-- Settings slide-over -->
    <BaseModal :open="settingsOpen" :slide-over="true" @close="closeSettings">
      <SettingsPanel :required="requiresManualPaths" @close="closeSettings" />
    </BaseModal>

    <!-- Search overlay -->
    <SearchOverlay :open="searchOpen" @close="closeSearch" @add-mod="onSearchAddMod" />

    <!-- Context menu -->
    <ContextMenu
      :open="contextMenu.open"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :items="contextMenuItems"
      :target-i-d="contextMenu.targetID"
      @close="closeContextMenu"
      @select="handleMenuAction"
    />

    <!-- Constraint modal -->
    <ConstraintModal :open="constraintModal.open" :mod-i-d="constraintModal.modID" @close="closeConstraintModal" />

    <!-- Toast notifications -->
    <ToastContainer />

    <!-- Manual Game Path Setup -->
    <ManualGamePathSetup
      v-if="manualSetupOpen"
      :gameID="setupGameID"
      :gameName="setupGameName"
      :open="manualSetupOpen"
      @close="manualSetupOpen = false"
    />
  </div>
</template>

<style scoped>
.shell {
  display: grid;
  grid-template-rows: 3.25rem 1fr;
  grid-template-columns: 12rem 1fr;
  height: 100%;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
}

.game-sidebar {
  display: flex;
  flex-direction: column;
  border-right: var(--border-width) solid var(--color-border);
  background: var(--color-bg-panel);
  padding: var(--space-4);
  overflow: hidden;
}

.sidebar-header {
  margin-bottom: var(--space-4);
}

.sidebar-title {
  font-family: var(--font-display), serif;
  font-size: 0.9rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  flex: 1;
}

.game-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: var(--border-width) solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  cursor: pointer;
  font-size: 0.85rem;
  transition: background var(--transition-fast), border-color var(--transition-fast);
}

.game-item:hover {
  background: var(--color-bg-elevated);
  border-color: var(--color-border-strong);
}

.game-item--active {
  background: var(--color-bg-elevated);
  border-color: var(--color-accent);
}

.game-item--detected {
  opacity: 1;
}

.game-item--undetected {
  opacity: 0.6;
}

.game-item-icon {
  flex-shrink: 0;
  font-size: 1.1rem;
}

.game-item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.game-item-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.game-item-status {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.commandbar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  grid-column: 1 / -1;
  align-items: center;
  padding: 0 var(--space-4);
  border-bottom: var(--border-width) solid var(--color-border);
  background: var(--color-bg-panel);
}

.commandbar-left,
.commandbar-right {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.commandbar-right {
  justify-content: flex-end;
}

.game-switcher {
  display: flex;
  align-items: center;
}

.game-select {
  min-height: 1.75rem;
  padding: 0 var(--space-2);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  font-size: 0.78rem;
  cursor: pointer;
}

.game-select:focus-visible {
  outline: none;
  border-color: var(--color-accent);
}

.commandbar-center {
  display: flex;
  justify-content: center;
}

.app-title {
  font-family: var(--font-display), serif;
  font-size: 0.9rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.search-trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 2rem;
  padding: 0 var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  color: var(--color-text-muted);
  font-size: 0.82rem;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
  min-width: 20rem;
}

.search-trigger:hover {
  border-color: var(--color-accent);
  background: var(--color-bg-base);
  color: var(--color-text-secondary);
}

.search-trigger kbd {
  margin-left: auto;
  padding: 0.1rem 0.3rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  font-family: var(--font-mono), monospace;
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.icon-btn:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

.icon-btn:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.content {
  display: flex;
  min-height: 0;
  overflow: hidden;
  padding: var(--space-5);
  gap: var(--space-4);
}

.details-slide {
  position: fixed;
  top: 3.25rem;
  right: 0;
  bottom: 0;
  width: 38%;
  min-width: 20rem;
  max-width: 32rem;
  border-left: var(--border-width) solid var(--color-border);
  background: var(--color-bg-panel);
  overflow: auto;
  padding: var(--space-5);
  z-index: 200;
}

.details-close {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  z-index: 1;
}

.details-close:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

.slide-over-enter-active,
.slide-over-leave-active {
  transition: transform var(--transition-panel), opacity var(--transition-panel);
}

.slide-over-enter-from,
.slide-over-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
