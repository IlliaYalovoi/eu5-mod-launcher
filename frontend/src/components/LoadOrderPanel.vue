<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import draggable from 'vuedraggable'
import type { LauncherLayout, Mod } from '../types'
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
                  <span class="group-stats">
                    <span class="enabled-count">{{ getCategoryStats(block.modIds).enabledCount }}</span>
                    <span class="total-count">/ {{ block.modIds.length }} mods</span>
                  </span>
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
                <template #item="{ element: modID }">
                  <article
                    class="mod-row"
                    :class="{ 'is-disabled': !modsByID[modID]?.enabled }"
                    @click="emit('select-mod', modID)"
                    @contextmenu.stop.prevent="onItemContextMenu($event, modID)"
                  >
                  <div class="mod-handle">⠿</div>
                  <div class="mod-info">
                    <div class="mod-name-row">
                      <span class="mod-name">{{ modsByID[modID]?.name || modID }}</span>
                      <span v-if="(modsByID[modID] as any)?.constraints && (modsByID[modID] as any).constraints.length > 0" class="mod-conflict-badge" title="Has constraints/rules">!</span>
                    </div>
                    <span class="mod-id">{{ modsByID[modID]?.tags?.join(', ') || modID }}</span>
                  </div>
                  <div
                    class="toggle"
                    :class="{ on: modsByID[modID]?.enabled }"
                    @click.stop="handleDisable(modID)"
                  ></div>
                </article>
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
  border-color: var(--accent);
}

.group-block.is-collapsed {
  background: rgba(255, 255, 255, 0.02);
}

.group-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  background: rgba(255, 255, 255, 0.05);
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border);
  cursor: pointer;
  user-select: none;
}

.group-header:hover {
  background: rgba(255, 255, 255, 0.08);
}

.group-header.is-collapsed {
  border-bottom: none;
}

.group-title-zone {
  flex: 1;
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
}

.group-name {
  font-family: var(--font-display);
  font-size: 0.95rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--accent);
  font-weight: 700;
}

.group-stats {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.enabled-count {
  color: var(--success);
  font-weight: 700;
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

.toggle {
  width: 34px;
  height: 18px;
  background: #333;
  border-radius: 9px;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
  border: 1px solid var(--border);
}

.toggle.on {
  background: var(--success);
  border-color: var(--success);
}

.toggle::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  top: 1px;
  left: 1px;
  transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.toggle.on::after {
  transform: translateX(16px);
}

.panel-header-right {
  display: flex;
  gap: var(--space-3);
}

.header-btn {
  background: var(--bg-panel);
  border: 1px solid var(--accent);
  color: var(--accent);
  padding: 6px 16px;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-weight: 700;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.header-btn:hover {
  background: var(--accent);
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
  color: var(--accent);
}

.group-delete {
  color: #ef4444;
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
  padding: var(--space-2);
  gap: 2px;
  min-height: 1rem;
}

.mod-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  background: transparent;
  border-radius: var(--radius-sm);
  border-left: 3px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.mod-row:hover {
  background: rgba(255, 255, 255, 0.04);
  border-left-color: var(--accent);
}

.mod-row.is-disabled {
  opacity: 0.6;
}

.mod-handle {
  cursor: grab !important;
  color: var(--text-muted);
  opacity: 0.3;
  font-size: 0.7rem;
}

.mod-row:hover .mod-handle {
  opacity: 0.7;
}

.mod-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.mod-name-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.mod-name {
  font-weight: 600;
  font-size: 0.85rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text);
}

.mod-conflict-badge {
  background: #f59e0b;
  color: black;
  font-size: 0.65rem;
  font-weight: 900;
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  flex-shrink: 0;
}

.mod-id {
  font-size: 0.7rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
}
</style>
