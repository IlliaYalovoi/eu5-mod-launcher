<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import draggable from 'vuedraggable'
import type { LauncherLayout, Mod } from '../types'
import LoadOrderItem from './LoadOrderItem.vue'
import {
  CreateLauncherCategory,
  DeleteLauncherCategory,
  GetAllMods,
  GetLauncherLayout,
  SaveCompiledLoadOrder,
  SetLauncherLayout,
  DisableMod,
} from '../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'

type EditableBlock = {
  id: string
  name: string
  modIds: string[]
  isUngrouped: boolean
  collapsed: boolean
}

const ungroupedID = 'category:ungrouped'

const emit = defineEmits<{
  (event: 'contextmenu', payload: { modID: string; x: number; y: number }): void
  (event: 'open-constraints', modID: string): void
  (event: 'select-mod', modID: string): void
  (event: 'load-order-changed'): void
  (event: 'manage-groups'): void
  (event: 'autosort'): void
}>()

// State owned by this panel
const launcherLayout = ref<LauncherLayout>({ ungrouped: [], categories: [], order: [], collapsed: {} })
const allMods = ref<Mod[]>([])
const persistError = ref<string | null>(null)
const categoryName = ref('')

const modsByID = computed(() => {
  const byID: Record<string, Mod> = {}
  for (const mod of allMods.value) byID[mod.id] = mod
  return byID
})

const getCategoryStats = (modIds: string[]) => {
  const mods = modIds.map(id => modsByID.value[id]).filter(Boolean)
  const enabledCount = mods.filter(m => m.enabled).length
  return { enabledCount, total: modIds.length }
}

const toggleCollapse = async (blockId: string) => {
  const layout = { ...launcherLayout.value }
  if (!layout.collapsed) layout.collapsed = {}
  layout.collapsed[blockId] = !layout.collapsed[blockId]
  launcherLayout.value = layout
  await SetLauncherLayout(layout as any)
}

const blocks = computed(() => {
  const value = launcherLayout.value
  const collapsed = value.collapsed || {}
  const categoryByID: Record<string, { id: string; name: string; modIds: string[] }> = {}
  for (const category of value.categories) {
    categoryByID[category.id] = { id: category.id, name: category.name, modIds: [...category.modIds] }
  }
  const order = value.order && value.order.length > 0 ? [...value.order] : [ungroupedID, ...value.categories.map((c) => c.id)]
  const next: EditableBlock[] = []
  const seen: Record<string, boolean> = {}

  for (const id of order) {
    if (seen[id]) continue
    seen[id] = true
    if (id === ungroupedID) {
      next.push({ id: ungroupedID, name: 'Ungrouped', modIds: [...value.ungrouped], isUngrouped: true, collapsed: !!collapsed[ungroupedID] })
      continue
    }
    const category = categoryByID[id]
    if (!category) continue
    next.push({ id: category.id, name: category.name, modIds: [...category.modIds], isUngrouped: false, collapsed: !!collapsed[category.id] })
  }
  if (!seen[ungroupedID]) {
    next.unshift({ id: ungroupedID, name: 'Ungrouped', modIds: [...value.ungrouped], isUngrouped: true, collapsed: !!collapsed[ungroupedID] })
  }
  for (const category of value.categories) {
    if (seen[category.id]) continue
    next.push({ id: category.id, name: category.name, modIds: [...category.modIds], isUngrouped: false, collapsed: !!collapsed[category.id] })
  }
  return next
})

const blocksModel = computed({
  get: () => blocks.value,
  set: (val) => persistLayoutAsync(val)
})

async function load() {
  try {
    const [layout, mods] = await Promise.all([
      GetLauncherLayout(),
      GetAllMods(),
    ])
    // The layout might be cached or lagging if we rely on a computed 'blocksModel'
    // setter to handle changes elsewhere. Explicitly reset.
    launcherLayout.value = layout as any as LauncherLayout
    allMods.value = mods as any as Mod[]
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

onMounted(load)

async function persistLayoutAsync(newBlocks: EditableBlock[]): Promise<void> {
  persistError.value = null
  const layout: LauncherLayout = {
    ungrouped: [],
    categories: [],
    order: [],
    collapsed: {}
  }

  for (const block of newBlocks) {
    if (layout.order) layout.order.push(block.id)
    if (block.collapsed) {
      if (!layout.collapsed) layout.collapsed = {}
      layout.collapsed[block.id] = true
    }
    if (block.isUngrouped) layout.ungrouped = [...block.modIds]
    else layout.categories.push({ id: block.id, name: block.name, modIds: [...block.modIds] })
  }

  try {
    await SetLauncherLayout(layout as any)
    emit('load-order-changed')
  } catch (err) {
    persistError.value = errorMessage(err)
    throw err
  }
}

async function handleDisable(modID: string) {
  try {
    await DisableMod(modID)
    await load()
    emit('load-order-changed')
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

function handleDragEnd() {
  void persistLayoutAsync(blocks.value)
}

function onCategoryCreate() {
  // Logic moved to Parent or Modal via manage-groups event
}

function onCategoryDelete(categoryID: string) {
  void DeleteLauncherCategory(categoryID)
    .then(() => load())
    .catch((err: unknown) => {
      showToast({ type: 'error', message: errorMessage(err) })
    })
}

function onItemContextMenu(event: MouseEvent, targetID: string): void {
  event.preventDefault()
  emit('contextmenu', { modID: targetID, x: event.clientX, y: event.clientY })
}

defineExpose({ load, launcherLayout })
</script>

<template>
  <div class="load-order-panel">
    <header class="panel-header">
      <div class="panel-header-left">
        <h2 class="panel-title">Load Order</h2>
      </div>
      <div class="panel-header-right">
        <button class="header-btn" @click="emit('manage-groups')">MANAGE GROUPS</button>
        <button class="header-btn" @click="emit('autosort')">AUTOSORT</button>
      </div>
    </header>

    <div class="load-order-list">
      <draggable v-model="blocksModel" item-key="id" handle=".group-handle" :animation="150">
        <template #item="{ element: block }">
          <section class="group-block" :class="{ 'is-collapsed': block.collapsed }" v-if="block.modIds.length > 0 || block.isUngrouped">
              <header class="group-header" :class="{ 'is-collapsed': block.collapsed }" @contextmenu="onItemContextMenu($event, block.id)" @click="toggleCollapse(block.id)">
                <span class="group-handle" @click.stop>⠿</span>
                <div class="group-title-zone">
                  <h3 class="group-name">{{ block.name }}</h3>
                  <div class="group-stats-pill">
                    <span class="enabled-count">{{ getCategoryStats(block.modIds).enabledCount }}</span>
                    <span class="total-count">/ {{ block.modIds.length }}</span>
                  </div>
                </div>
                <div class="group-actions">
                  <button v-if="!block.isUngrouped" class="group-delete" @click.stop="onCategoryDelete(block.id)">×</button>
                  <span class="collapse-icon">{{ block.collapsed ? '▼' : '▲' }}</span>
                </div>
              </header>

              <draggable
                v-show="!block.collapsed"
                :model-value="block.modIds"
                @update:model-value="val => { persistLayoutAsync(blocks.map(b => b.id === block.id ? { ...b, modIds: val } : b)) }"
                item-key="id"
                group="mods"
                handle=".mod-handle"
                :animation="150"
                class="mods-list"
              >
                <template #item="{ element: modID, index }">
                  <LoadOrderItem
                    v-if="modsByID[modID]"
                    :mod="modsByID[modID]"
                    :index="index"
                    :is-disabled="!modsByID[modID].enabled"
                    @click="emit('select-mod', modID)"
                    @toggle="handleDisable"
                    @contextmenu="payload => emit('contextmenu', payload)"
                  />
                  <div v-else class="mod-missing">Missing Mod: {{ modID }}</div>
                </template>
              </draggable>
          </section>
        </template>
      </draggable>
    </div>
  </div>
</template>

<style scoped>
.load-order-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: var(--space-5);
  gap: var(--space-5);
  min-height: 0;
  overflow: hidden;
  background: var(--bg-body);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: var(--space-3);
  border-bottom: 2px solid var(--border);
}

.panel-title {
  font-family: var(--font-display);
  font-size: 1.25rem;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.15em;
  font-weight: 800;
}

.load-order-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding-right: var(--space-2);
}

.group-block {
  background: var(--bg-sidebar);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.group-block:hover {
  border-color: var(--accent-primary, var(--accent));
}

.group-block.is-collapsed {
  background: rgba(255, 255, 255, 0.02);
}

.group-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  background: var(--card-bg, rgba(255, 255, 255, 0.05));
  padding: var(--space-2) var(--space-4);
  border-bottom: 1px solid var(--border);
  cursor: pointer;
  user-select: none;
  min-height: 3rem;
}

.group-header:hover {
  background: var(--card-bg-hover, rgba(255, 255, 255, 0.08));
}

.group-header.is-collapsed {
  border-bottom: none;
}

.group-title-zone {
  flex: 1;
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.group-name {
  font-family: var(--font-display);
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--accent-primary, var(--accent));
  font-weight: 800;
}

.group-stats-pill {
  background: rgba(0, 0, 0, 0.2);
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 700;
  display: flex;
  gap: 4px;
  border: 1px solid var(--border);
}

.enabled-count {
  color: var(--success);
}

.total-count {
  color: var(--text-muted);
}

.group-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.collapse-icon {
  font-size: 0.7rem;
  color: var(--text-muted);
  width: 1rem;
  text-align: center;
}

.panel-header-right {
  display: flex;
  gap: var(--space-3);
}

.header-btn {
  background: var(--bg-panel);
  border: 1px solid var(--accent-primary, var(--accent));
  color: var(--accent-primary, var(--accent));
  padding: 6px 16px;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-weight: 700;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.header-btn:hover {
  background: var(--accent-primary, var(--accent));
  color: var(--bg-body);
}

.group-handle {
  cursor: grab !important;
  color: var(--text-muted);
  font-size: 0.8rem;
  opacity: 0.5;
}

.group-handle:hover {
  opacity: 1;
  color: var(--accent-primary, var(--accent));
}

.group-delete {
  color: var(--error, #ef4444);
  font-size: 1.1rem;
  line-height: 1;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.group-delete:hover {
  opacity: 1;
}

.mods-list {
  display: flex;
  flex-direction: column;
  padding: var(--space-3);
  gap: var(--space-2);
  min-height: 1rem;
}

.mod-missing {
  padding: var(--space-2);
  color: var(--text-muted);
  font-size: 0.8rem;
  font-style: italic;
  border: 1px dashed var(--border);
  border-radius: var(--radius-sm);
}
</style>
