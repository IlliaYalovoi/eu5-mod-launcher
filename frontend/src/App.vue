<script lang="ts" setup>
import {computed, onMounted, onUnmounted, reactive, ref, watch} from 'vue'
import { useActiveGameStore } from './stores/activeGame'
import type { LauncherLayout, Mod, WorkspaceMode } from './types'
import {
  GetGameActivePlaysetIndex,
  GetAllMods,
  GetModsDirStatus,
  GetPlaysetNames,
  ListSupportedGames,
  SetActiveGame,
  SetLauncherActivePlaysetIndex,
  SaveCompiledLoadOrder,
} from './wailsjs/go/launcher/App'
import { showToast } from './lib/toast'
import { errorMessage } from './lib/error'
import LoadOrderPanel from './components/LoadOrderPanel.vue'
import ModRepository from './components/ModRepository.vue'
import ModDetailsPanel from './components/ModDetailsPanel.vue'
import LaunchButton from './components/LaunchButton.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import ToastContainer from './components/ui/ToastContainer.vue'
import ManualGamePathSetup from './components/ManualGamePathSetup.vue'
import ContextMenu from './components/ui/ContextMenu.vue'
import ConstraintModal from './components/ConstraintModal.vue'
import BaseModal from './components/ui/BaseModal.vue'

const activeGameStore = useActiveGameStore()

const currentMode = ref<WorkspaceMode>('load-order')

type MenuItem = {
  id: string
  label: string
  icon?: string
  danger?: boolean
  disabled?: boolean
  children?: MenuItem[]
}

// Component refs
const loadOrderPanel = ref<InstanceType<typeof LoadOrderPanel> | null>(null)
const modRepository = ref<InstanceType<typeof ModRepository> | null>(null)

// Sidebar state
const supportedGames = ref<Array<{ id: string; name: string; detected: boolean }>>([])
const requiresManualPaths = ref(false)

// Playset state
const playsetNames = ref<string[]>([])
const launcherActivePlaysetIndex = ref(-1)

// Mod state
const selectedMod = ref<Mod | null>(null)
const allMods = ref<Mod[]>([])
const detailsOpen = ref(false)

const contextMenu = reactive({ open: false, x: 0, y: 0, targetID: '' })
const constraintModal = reactive({ open: false, modID: '' })
const manageGroupsModal = reactive({ open: false })
const newGroupName = ref('')

// Panes
const settingsOpen = ref(false)
const manualSetupOpen = ref(false)
const setupGameID = ref<string>('')
const setupGameName = ref<string>('')

async function load() {
  try {
    const [games, gameIdx, names, dirsStatus, mods] = await Promise.all([
      ListSupportedGames(),
      GetGameActivePlaysetIndex(),
      GetPlaysetNames(),
      GetModsDirStatus(),
      GetAllMods(),
    ])
    supportedGames.value = games.map((g: any) => ({ id: g.id, name: g.name, detected: g.detected }))

    // Set initial active game from backend if detected
    const current = supportedGames.value.find(g => g.id === activeGameStore.activeGameID)
    if (!current || !current.detected) {
      const firstDetected = supportedGames.value.find(g => g.detected)
      if (firstDetected) {
        activeGameStore.activeGameID = firstDetected.id
        await SetActiveGame(firstDetected.id)
      }
    }

    launcherActivePlaysetIndex.value = gameIdx
    playsetNames.value = names
    allMods.value = mods as unknown as Mod[]
    requiresManualPaths.value = !(dirsStatus as any).autoDetectedExists && !(dirsStatus as any).effectiveExists
    if (requiresManualPaths.value) settingsOpen.value = true
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

async function createGroup() {
  const name = newGroupName.value.trim()
  if (!name) return
  try {
    await (window as any).go.launcher.App.CreateLauncherCategory(name)
    newGroupName.value = ''
    await load()
    refreshPanels()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

async function deleteGroup(id: string) {
  try {
    await (window as any).go.launcher.App.DeleteLauncherCategory(id)
    await load()
    refreshPanels()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

async function handleAutosort() {
  try {
    await SaveCompiledLoadOrder()
    showToast({ type: 'success', message: 'Load order updated and sorted' })
    await load()
    refreshPanels()
  } catch (err) {
     showToast({ type: 'error', message: errorMessage(err) })
  }
}

function handleGlobalKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    if (detailsOpen.value) { detailsOpen.value = false; return }
    if (settingsOpen.value && !requiresManualPaths.value) { settingsOpen.value = false; return }
  }
  if ((event.ctrlKey || event.metaKey) && event.key === 's') { event.preventDefault(); void SaveCompiledLoadOrder(); return }
}

watch(() => activeGameStore.activeGameID, (id) => {
  document.body.className = id ? `theme-${id.toLowerCase()}` : ''
}, { immediate: true })

onMounted(() => { window.addEventListener('keydown', handleGlobalKeydown); void load() })
onUnmounted(() => { window.removeEventListener('keydown', handleGlobalKeydown) })

const contextMenuItems = computed<MenuItem[]>(() => {
  return [
    { id: 'add_constraint', label: 'Add rule...', icon: '⛓' },
    { id: 'view_constraints', label: 'View rules', icon: '📖' },
  ]
})

function openContextMenu(event: { modID: string; x: number; y: number }): void {
  contextMenu.open = true
  contextMenu.x = event.x
  contextMenu.y = event.y
  contextMenu.targetID = event.modID
}

function handleMenuAction(event: { itemID: string; targetID: string }): void {
  if (event.itemID === 'add_constraint' || event.itemID === 'view_constraints') {
    constraintModal.modID = event.targetID
    constraintModal.open = true
  }
}

async function onLauncherPlaysetChange(event: Event): Promise<void> {
  const target = event.target as HTMLSelectElement
  const index = parseInt(target.value, 10)
  if (isNaN(index) || index === launcherActivePlaysetIndex.value) return
  try {
    launcherActivePlaysetIndex.value = index
    await SetLauncherActivePlaysetIndex(index)
    // We need to wait for the backend to settle before refreshing
    // or just call load() to get the fresh state for the new playset
    await load()
    refreshPanels()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
    await load() // Revert UI to match backend
  }
}

async function onGameClick(game: { id: string; name: string; detected: boolean }): Promise<void> {
  if (!game.detected) { setupGameID.value = game.id; setupGameName.value = game.name; manualSetupOpen.value = true; return }
  try {
    await SetActiveGame(game.id)
    activeGameStore.activeGameID = game.id
    await load()
    loadOrderPanel.value?.load()
    modRepository.value?.load()
  } catch (err) { showToast({ type: 'error', message: errorMessage(err) }) }
}

function onModSelect(modID: string): void {
  selectedMod.value = allMods.value.find(m => m.id === modID) || null
}

function refreshPanels() {
  loadOrderPanel.value?.load()
  modRepository.value?.load()
}

function closeDetails() {
  detailsOpen.value = false
}

function closeSettings() {
  if (!requiresManualPaths.value) settingsOpen.value = false
}
</script>

<template>
  <div class="app-shell">
    <nav class="command-rail">
      <div class="rail-header">
        <h1 class="app-title-mini">PDX</h1>
      </div>

      <div class="rail-nav">
        <button
          v-for="game in supportedGames"
          :key="game.id"
          class="rail-game-btn"
          :class="{
            'rail-game-btn--active': activeGameStore.activeGameID === game.id,
            'rail-game-btn--undetected': !game.detected
          }"
          @click="onGameClick(game)"
          :title="game.name"
        >
          <span class="game-icon">⚔️</span>
        </button>
      </div>

      <div class="rail-footer">
        <button class="rail-action-btn" @click="settingsOpen = true" title="Settings">⚙️</button>
        <LaunchButton
          v-if="activeGameStore.activeGameID"
          :playset-names="playsetNames"
          :game-active-playset-index="launcherActivePlaysetIndex"
          compact
        />
      </div>
    </nav>

    <main class="workspace-center">
      <header class="workspace-header" v-if="activeGameStore.activeGameID">
        <div class="workspace-meta">
          <h1 class="target-game-name">{{ supportedGames.find(g => g.id === activeGameStore.activeGameID)?.name }}</h1>
          <div v-if="playsetNames.length > 0" class="playset-nav">
             <select
                class="playset-select-minimal"
                :value="launcherActivePlaysetIndex"
                @change="onLauncherPlaysetChange"
              >
                <option v-for="(name, index) in playsetNames" :key="`${name}-${index}`" :value="index">
                  {{ name }}
                </option>
              </select>
          </div>
        </div>

        <nav class="mode-tabs">
          <button
            class="mode-tab"
            :class="{ 'mode-tab--active': currentMode === 'load-order' }"
            @click="currentMode = 'load-order'"
          >
            Load Order
          </button>
          <button
            class="mode-tab"
            :class="{ 'mode-tab--active': currentMode === 'discover' }"
            @click="currentMode = 'discover'"
          >
            Mod Repository
          </button>
        </nav>
      </header>

      <div class="workspace-content">
        <LoadOrderPanel
          v-if="currentMode === 'load-order'"
          ref="loadOrderPanel"
          @select-mod="onModSelect"
          @contextmenu="openContextMenu"
          @open-constraints="modID => { constraintModal.modID = modID; constraintModal.open = true }"
          @load-order-changed="refreshPanels"
          @manage-groups="manageGroupsModal.open = true"
          @autosort="handleAutosort"
        />
        <ModRepository
          v-else-if="currentMode === 'discover'"
          ref="modRepository"
          @select-mod="onModSelect"
          @mod-enabled="refreshPanels"
        />
      </div>
    </main>

    <aside class="inspector-right" :class="{ 'inspector-right--open': !!selectedMod }">
      <ModDetailsPanel :mod="selectedMod" />
    </aside>

    <!-- Manage Groups Modal -->
    <BaseModal :open="manageGroupsModal.open" @close="manageGroupsModal.open = false">
       <div class="manage-groups-view">
          <header class="modal-header">
            <h3>Manage Groups</h3>
          </header>
          <div class="group-creation">
            <input v-model="newGroupName" placeholder="New Group Name..." @keydown.enter="createGroup" class="modal-input" />
            <button @click="createGroup" class="modal-btn">+</button>
          </div>
          <div class="groups-list-managed">
            <div v-for="cat in (loadOrderPanel?.launcherLayout?.categories || [])" :key="cat.id" class="managed-group-item">
               <span>{{ cat.name }}</span>
               <button @click="deleteGroup(cat.id)" class="delete-btn">×</button>
            </div>
          </div>
       </div>
    </BaseModal>

    <!-- Mod Details Modal removed, now in inspector -->

    <!-- Context Menu -->
    <ContextMenu
      :open="contextMenu.open"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :items="contextMenuItems"
      :target-i-d="contextMenu.targetID"
      @close="contextMenu.open = false"
      @select="handleMenuAction"
    />

    <!-- Constraint Modal -->
    <ConstraintModal
      :open="constraintModal.open"
      :mod-i-d="constraintModal.modID"
      @close="constraintModal.open = false"
    />

    <!-- Settings Modal -->
    <BaseModal :open="settingsOpen" @close="closeSettings">
      <SettingsPanel :required="requiresManualPaths" @close="closeSettings" />
    </BaseModal>

    <ToastContainer />

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
.app-shell {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  background: var(--bg-body);
}

/* Command Rail */
.command-rail {
  width: 72px;
  background: var(--rail-bg, var(--bg-sidebar));
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  padding: var(--space-4) 0;
}

.rail-header {
  margin-bottom: var(--space-6);
}

.app-title-mini {
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 800;
  color: var(--accent);
}

.rail-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.rail-game-btn {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-elevated);
  transition: all var(--transition-fast);
  border: 2px solid transparent;
}

.rail-game-btn:hover {
  border-color: var(--border);
  transform: translateY(-1px);
}

.rail-game-btn--active {
  background: var(--accent);
  color: var(--bg-body);
  border-color: var(--accent);
  box-shadow: 0 0 15px var(--accent-alpha, rgba(var(--accent-rgb), 0.3));
}

.rail-game-btn--undetected {
  opacity: 0.3;
  filter: grayscale(1);
}

.rail-footer {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  align-items: center;
}

.rail-action-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  transition: transform 0.2s;
}

.rail-action-btn:hover {
  transform: rotate(30deg);
}

/* Workspace Center */
.workspace-center {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: var(--workspace-bg, var(--bg-body));
}

.workspace-header {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.workspace-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.target-game-name {
  font-family: var(--font-display);
  font-size: 1.25rem;
  margin: 0;
  color: var(--text-base);
}

.playset-select-minimal {
  padding: var(--space-1) var(--space-3);
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--accent);
  font-size: 0.85rem;
  outline: none;
}

.mode-tabs {
  display: flex;
  gap: var(--space-6);
}

.mode-tab {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-muted);
  padding-bottom: var(--space-2);
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.mode-tab:hover {
  color: var(--text-base);
}

.mode-tab--active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.workspace-content {
  flex: 1;
  overflow: hidden;
}

/* Inspector Right */
.inspector-right {
  width: 320px;
  background: var(--inspector-bg, var(--bg-panel));
  border-left: 1px solid var(--border);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  flex-shrink: 0;
}

.inspector-right--open {
  width: 420px;
}

.inspector-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-style: italic;
}

.manage-groups-view {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.modal-header h3 {
  font-family: var(--font-display);
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.group-creation {
  display: flex;
  gap: var(--space-2);
}

.modal-input {
  flex: 1;
  background: var(--bg-body);
  border: 1px solid var(--border);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
}

.modal-btn {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  color: var(--accent);
  padding: 0 var(--space-4);
  font-weight: 700;
  border-radius: var(--radius-sm);
}

.groups-list-managed {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  max-height: 400px;
  overflow-y: auto;
}

.managed-group-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.managed-group-item .delete-btn {
  color: #ef4444;
  font-size: 1.2rem;
  font-weight: 700;
}
</style>
