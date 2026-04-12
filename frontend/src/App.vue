<script lang="ts" setup>
import {computed, onMounted, onUnmounted, reactive, ref, watch} from 'vue'
import { useActiveGameStore } from './stores/activeGame'
import type { LauncherLayout, Mod } from './types'
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
  detailsOpen.value = true
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
  <div class="shell">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1 class="app-title">PDX Mod Organizer</h1>
      </div>

      <nav class="sidebar-nav">
        <div class="nav-section">
          <h3 class="nav-label">Games</h3>
          <div class="games-list">
            <div
              v-for="game in supportedGames"
              :key="game.id"
              class="game-item"
            >
              <button
                class="game-card"
                :class="{
                  'game-card--active': activeGameStore.activeGameID === game.id,
                  'game-card--undetected': !game.detected
                }"
                @click="onGameClick(game)"
              >
                <span class="game-icon">⚔️</span>
                <span class="game-name">{{ game.name }}</span>
              </button>

              <div
                v-if="activeGameStore.activeGameID === game.id && playsetNames.length > 0"
                class="playset-selector"
              >
                <select
                  class="playset-select"
                  :value="launcherActivePlaysetIndex"
                  @change="onLauncherPlaysetChange"
                >
                  <option v-for="(name, index) in playsetNames" :key="`${name}-${index}`" :value="index">
                    {{ name }}
                  </option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </nav>

      <div class="sidebar-footer">
        <div class="footer-actions">
          <button class="footer-btn" @click="settingsOpen = true">Settings</button>
        </div>
        <LaunchButton
          :playset-names="playsetNames"
          :game-active-playset-index="launcherActivePlaysetIndex"
        />
      </div>
    </aside>

    <main class="content-area">
      <header class="toolbar" v-if="activeGameStore.activeGameID">
        <h1 class="target-game-title">{{ supportedGames.find(g => g.id === activeGameStore.activeGameID)?.name }}</h1>
      </header>
      <div class="main-split">
        <LoadOrderPanel
          ref="loadOrderPanel"
          @select-mod="onModSelect"
          @contextmenu="openContextMenu"
          @open-constraints="modID => { constraintModal.modID = modID; constraintModal.open = true }"
          @load-order-changed="refreshPanels"
          @manage-groups="manageGroupsModal.open = true"
          @autosort="handleAutosort"
        />
        <ModRepository
          ref="modRepository"
          @select-mod="onModSelect"
          @mod-enabled="refreshPanels"
        />
      </div>
    </main>

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

    <!-- Mod Details Modal -->
    <BaseModal :open="detailsOpen" @close="closeDetails">
      <ModDetailsPanel :mod="selectedMod" @close="closeDetails" />
    </BaseModal>

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
.shell {
  display: flex;
  height: 100%;
  width: 100%;
}

.sidebar {
  width: 280px;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: var(--space-5) var(--space-4);
}

.app-title {
  font-family: var(--font-display);
  font-size: 1.5rem;
  color: var(--accent);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  margin: 0;
}

.sidebar-nav {
  flex: 1;
  padding: 0 var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  overflow-y: auto;
}

.nav-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin-bottom: var(--space-2);
}

.games-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.game-card {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  text-align: left;
}

.game-card:hover {
  background: var(--bg-elevated);
}

.game-card--active {
  background: var(--bg-panel);
  border: 1px solid var(--accent);
}

.game-card--undetected {
  opacity: 0.5;
}

.game-name {
  font-weight: 600;
  font-size: 0.9rem;
}

.game-item {
  display: flex;
  flex-direction: column;
}

.playset-selector {
  margin-top: 10px;
  margin-left: 32px;
}

.playset-select {
  width: 100%;
  padding: var(--space-2);
  background: var(--bg-body);
  border: 1px solid var(--border);
  color: var(--accent);
  border-radius: var(--radius-sm);
  outline: none;
  font-size: 0.8rem;
}

.playset-select:focus {
  border-color: var(--accent);
}

.sidebar-footer {
  padding: var(--space-4);
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.footer-actions {
  display: flex;
  justify-content: flex-end;
}

.footer-btn {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.footer-btn:hover {
  color: var(--accent);
}

.toolbar {
  padding: 20px 40px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--border);
  background: rgba(0,0,0,0.1);
}

.target-game-title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 1.5rem;
  letter-spacing: 0.05em;
}

.toolbar-actions {
  display: flex;
  gap: 20px;
}

.toolbar-btn {
  background: transparent;
  border: 1px solid var(--accent);
  color: var(--accent);
  padding: 5px 15px;
  font-family: var(--font-body);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.toolbar-btn:hover {
  background: var(--accent);
  color: var(--bg-body);
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

.content-area {
  flex: 1;
  min-width: 0;
  background: var(--bg-body);
  display: flex;
  flex-direction: column;
}

.main-split {
  display: grid;
  grid-template-columns: 1fr 340px;
  flex: 1;
  overflow: hidden;
}
</style>
