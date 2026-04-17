<script lang="ts" setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import Sidebar from './components/Sidebar.vue'
import LoadOrderPanel from './components/LoadOrderPanel.vue'
import ModDetailsPanel from './components/ModDetailsPanel.vue'
import ConstraintModal from './components/ConstraintModal.vue'
import ContextMenu from './components/ui/ContextMenu.vue'
import BaseModal from './components/ui/BaseModal.vue'
import SettingsPanel from './components/SettingsPanel.vue'
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
modsStore.startPolling()
const { orderedIDs, launcherLayout } = storeToRefs(loadOrderStore)
const { unsubscribeFeatureEnabled, unsubscribeNotice } = storeToRefs(modsStore)
const { requiresManualPaths } = storeToRefs(settingsStore)

const ungroupedCategoryID = 'category:ungrouped'

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

const settingsOpen = ref(false)

const appThemeClass = computed(() => `theme-${settingsStore.activeGameID}`)

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

const contextMenuItems = computed<MenuItem[]>(() => {
  const isCategory = contextMenu.targetID?.indexOf('category:') === 0

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
    { id: 'disable_mod', label: 'Disable mod', icon: '⛔', danger: true, disabled: missing },
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
  settingsOpen.value = true
}

function closeSettings(): void {
  settingsOpen.value = false
}

function openConstraintModal(modID: string): void {
  constraintModal.modID = modID
  constraintModal.open = true
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
    const confirmed = window.confirm('Unsubscribe this mod from Steam Workshop? Steam may take a moment to sync.')
    if (!confirmed) {
      return
    }
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

  // Task 10 opens modal flows for these actions.
  if (event.itemID === 'add_constraint' || event.itemID === 'view_constraints') {
    openConstraintModal(targetID)
    return
  }
}
</script>

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
