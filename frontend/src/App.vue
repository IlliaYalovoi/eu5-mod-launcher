<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import LoadOrderPanel from './components/LoadOrderPanel.vue'
import ModDetailsPanel from './components/ModDetailsPanel.vue'
import ConstraintModal from './components/ConstraintModal.vue'
import ContextMenu from './components/ui/ContextMenu.vue'
import BaseModal from './components/ui/BaseModal.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import SearchOverlay from './components/ui/SearchOverlay.vue'
import ToastContainer from './components/ui/ToastContainer.vue'
import { useLoadOrderStore } from './stores/loadorder'
import { useModsStore } from './stores/mods'
import { useSettingsStore } from './stores/settings'

type MenuItem = {
  id: string
  label: string
  icon?: string
  danger?: boolean
  disabled?: boolean
  children?: MenuItem[]
}

const loadOrderStore = useLoadOrderStore()
const modsStore = useModsStore()
const settingsStore = useSettingsStore()
const { orderedIDs, launcherLayout, playsetNames, launcherActivePlaysetIndex } = storeToRefs(loadOrderStore)
const { unsubscribeFeatureEnabled, unsubscribeNotice } = storeToRefs(modsStore)
const { requiresManualPaths } = storeToRefs(settingsStore)

const ungroupedCategoryID = 'category:ungrouped'

// Slide-over panels
const detailsOpen = ref(false)
const settingsOpen = ref(false)
const searchOpen = ref(false)

const contextMenu = reactive({
  open: false,
  x: 0,
  y: 0,
  targetID: '',
})

const constraintModal = reactive({
  open: false,
  modID: '',
})

let unsubscribeNoticeTimeout: ReturnType<typeof setTimeout> | null = null

watch(
  requiresManualPaths,
  (required) => {
    if (required) {
      settingsOpen.value = true
    }
  },
  { immediate: true },
)

watch(unsubscribeNotice, (notice) => {
  if (unsubscribeNoticeTimeout) {
    clearTimeout(unsubscribeNoticeTimeout)
    unsubscribeNoticeTimeout = null
  }
  if (!notice) {
    return
  }
  unsubscribeNoticeTimeout = setTimeout(() => {
    modsStore.clearUnsubscribeNotice()
  }, 3200)
})

onBeforeUnmount(() => {
  if (unsubscribeNoticeTimeout) {
    clearTimeout(unsubscribeNoticeTimeout)
    unsubscribeNoticeTimeout = null
  }
})

function handleGlobalKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    if (searchOpen.value) {
      searchOpen.value = false
      return
    }
    if (detailsOpen.value) {
      detailsOpen.value = false
      return
    }
    if (settingsOpen.value && !requiresManualPaths.value) {
      settingsOpen.value = false
      return
    }
  }

  if ((event.ctrlKey || event.metaKey) && event.key === 'k') {
    event.preventDefault()
    searchOpen.value = !searchOpen.value
    return
  }

  if ((event.ctrlKey || event.metaKey) && event.key === 's') {
    event.preventDefault()
    void loadOrderStore.saveCompiled()
    return
  }

  if ((event.ctrlKey || event.metaKey) && event.key === 'f') {
    event.preventDefault()
    searchOpen.value = true
    return
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
})

const contextMenuItems = computed<MenuItem[]>(() => {
  const isCategory = contextMenu.targetID.indexOf('category:') === 0

  const currentCategoryForTarget = (() => {
    for (const category of launcherLayout.value.categories) {
      if (category.modIds.includes(contextMenu.targetID)) {
        return category.id
      }
    }
    if (launcherLayout.value.ungrouped.includes(contextMenu.targetID)) {
      return ungroupedCategoryID
    }
    return ''
  })()

  const moveToCategoryItems: MenuItem[] = [
    {
      id: `move_to_category:${ungroupedCategoryID}`,
      label: 'Move to category: Ungrouped',
      icon: '📁',
      disabled: currentCategoryForTarget === ungroupedCategoryID,
    },
    ...launcherLayout.value.categories.map((category) => ({
      id: `move_to_category:${category.id}`,
      label: `Move to category: ${category.name}`,
      icon: '📁',
      disabled: currentCategoryForTarget === category.id,
    })),
  ]

  const moveToCategoryMenu: MenuItem = {
    id: 'move_to_category_menu',
    label: 'Move to category',
    icon: '📁',
    disabled: moveToCategoryItems.every((item) => item.disabled),
    children: moveToCategoryItems,
  }

  if (isCategory) {
    return [
      { id: 'add_constraint', label: 'Add constraint...', icon: '⛓' },
      { id: 'view_constraints', label: 'View constraints', icon: '📖' },
    ]
  }

  const index = orderedIDs.value.indexOf(contextMenu.targetID)
  const missing = index < 0
  const atTop = index === 0
  const atBottom = index === orderedIDs.value.length - 1
  const isWorkshopMod = modsStore.isWorkshopMod(contextMenu.targetID)
  const unsubscribeLoading = modsStore.isUnsubscribeLoading(contextMenu.targetID)
  const canShowUnsubscribe = unsubscribeFeatureEnabled.value

  const items: MenuItem[] = [
    { id: 'add_constraint', label: 'Add constraint...', icon: '⛓' },
    { id: 'view_constraints', label: 'View constraints', icon: '📖' },
    moveToCategoryMenu,
    { id: 'move_top', label: 'Move to top', icon: '⬆', disabled: missing || atTop },
    { id: 'move_bottom', label: 'Move to bottom', icon: '⬇', disabled: missing || atBottom },
    { id: 'disable_mod', label: 'Remove from load order', icon: '⛔', danger: true, disabled: missing },
  ]

  if (canShowUnsubscribe) {
    items.splice(5, 0, {
      id: 'unsubscribe_workshop',
      label: unsubscribeLoading ? 'Unsubscribing...' : 'Unsubscribe from Workshop...',
      icon: '🧹',
      danger: true,
      disabled: missing || !isWorkshopMod || unsubscribeLoading,
    })
  }

  return items
})

async function moveModToCategory(modID: string, categoryID: string): Promise<void> {
  const next = {
    ungrouped: [...launcherLayout.value.ungrouped],
    categories: launcherLayout.value.categories.map((category) => ({
      id: category.id,
      name: category.name,
      modIds: [...category.modIds],
    })),
    order: launcherLayout.value.order ? [...launcherLayout.value.order] : undefined,
    collapsed: launcherLayout.value.collapsed ? { ...launcherLayout.value.collapsed } : undefined,
  }

  next.ungrouped = next.ungrouped.filter((id) => id !== modID)
  for (const category of next.categories) {
    category.modIds = category.modIds.filter((id) => id !== modID)
  }

  if (categoryID === ungroupedCategoryID) {
    next.ungrouped.push(modID)
  } else {
    const target = next.categories.find((category) => category.id === categoryID)
    if (!target) {
      return
    }
    target.modIds.push(modID)
  }

  await loadOrderStore.persistLauncherLayout(next)
}

function openContextMenu(event: { modID: string; x: number; y: number }): void {
  contextMenu.open = true
  contextMenu.x = event.x
  contextMenu.y = event.y
  contextMenu.targetID = event.modID
}

function closeContextMenu(): void {
  contextMenu.open = false
}

function closeConstraintModal(): void {
  constraintModal.open = false
  constraintModal.modID = ''
}

function openSettings(): void {
  searchOpen.value = false
  detailsOpen.value = false
  settingsOpen.value = true
}

function closeSettings(): void {
  if (requiresManualPaths.value) {
    return
  }
  settingsOpen.value = false
}

function openConstraintModal(modID: string): void {
  constraintModal.modID = modID
  constraintModal.open = true
}

function onLauncherPlaysetChange(event: Event): void {
  const target = event.target as HTMLSelectElement
  const index = parseInt(target.value, 10)
  if (index.toString() !== target.value || index === launcherActivePlaysetIndex.value) {
    return
  }
  void loadOrderStore.setLauncherPlayset(index).then(() => modsStore.fetchAll())
}

function openSearch(): void {
  searchOpen.value = true
}

function closeSearch(): void {
  searchOpen.value = false
}

function onSearchAddMod(modID: string): void {
  searchOpen.value = false
  modsStore.selectMod(modID)
  detailsOpen.value = true
}

function onModSelect(modID: string): void {
  if (modID === modsStore.selectedModID && detailsOpen.value) {
    detailsOpen.value = false
    return
  }
  searchOpen.value = false
  settingsOpen.value = false
  modsStore.selectMod(modID)
  detailsOpen.value = true
}

function closeDetails(): void {
  detailsOpen.value = false
}

async function handleMenuAction(event: { itemID: string; targetID: string }): Promise<void> {
  const targetID = event.targetID
  const isCategory = targetID.indexOf('category:') === 0

  if (isCategory && event.itemID === 'disable_mod') {
    return
  }

  const current = [...orderedIDs.value]
  const index = current.indexOf(targetID)

  if (event.itemID === 'disable_mod') {
    await modsStore.setEnabled(targetID, false)
    return
  }

  if (event.itemID === 'unsubscribe_workshop') {
    await modsStore.unsubscribeWorkshop(targetID)
    return
  }

  if (event.itemID.indexOf('move_to_category:') === 0) {
    const categoryID = event.itemID.slice('move_to_category:'.length)
    await moveModToCategory(targetID, categoryID)
    return
  }

  if (event.itemID === 'move_top' && index > 0) {
    current.splice(index, 1)
    current.unshift(targetID)
    await loadOrderStore.persist(current)
    return
  }

  if (event.itemID === 'move_bottom' && index >= 0 && index < current.length - 1) {
    current.splice(index, 1)
    current.push(targetID)
    await loadOrderStore.persist(current)
    return
  }

  if (event.itemID === 'add_constraint' || event.itemID === 'view_constraints') {
    openConstraintModal(targetID)
    return
  }
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
        <ModDetailsPanel />
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
  </div>
</template>

<style scoped>
.shell {
  display: grid;
  grid-template-rows: 3.25rem 1fr;
  grid-template-columns: 1fr;
  height: 100%;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
}

.commandbar {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
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
