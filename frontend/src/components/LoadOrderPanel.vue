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
          <section class="group-block" v-if="block.modIds.length > 0 || block.isUngrouped">
              <header class="group-header" @contextmenu="onItemContextMenu($event, block.id)">
                <span class="group-handle">⠿</span>
                <h3 class="group-name">{{ block.name }}</h3>
                <span class="group-count">{{ block.modIds.length }}</span>
                <button v-if="!block.isUngrouped" class="group-delete" @click="onCategoryDelete(block.id)">×</button>
              </header>

              <draggable
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
                    @click="emit('select-mod', modID)"
                    @contextmenu.stop.prevent="onItemContextMenu($event, modID)"
                  >
                  <div class="mod-info">
                    <span class="mod-name">{{ modsByID[modID]?.name || modID }}</span>
                    <span class="mod-id">{{ modsByID[modID]?.tags?.join(', ') || modID }}</span>
                  </div>
                  <div
                    class="toggle"
                    :class="{ on: true }"
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
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-family: var(--font-display);
  font-size: 1.2rem;
  color: var(--text);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.category-creator {
  display: flex;
  gap: var(--space-2);
}

.category-input {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: var(--space-1) var(--space-3);
  font-size: 0.85rem;
}

.add-btn {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  width: 2rem;
  color: var(--accent);
  font-weight: 700;
}

.load-order-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.group-block {
  background: var(--bg-sidebar);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
}

.toggle {
  width: 36px;
  height: 18px;
  background: #444;
  border-radius: 10px;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
}

.toggle.on {
  background: var(--success, #5c7c51);
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

.panel-header-right {
  display: flex;
  gap: var(--space-3);
}

.header-btn {
  background: transparent;
  border: 1px solid var(--accent);
  color: var(--accent);
  padding: 5px 15px;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-family: var(--font-body);
}

.header-btn:hover {
  background: var(--accent);
  color: var(--bg-body);
}

.group-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
  background: rgba(255, 255, 255, 0.03);
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border);
}

.group-handle {
  cursor: grab;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.group-name {
  font-family: var(--font-display);
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--accent);
  flex: 1;
}

.group-count {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.group-delete {
  color: #ef4444;
  font-size: 1.2rem;
  line-height: 1;
}

.mods-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-height: 2rem;
}

.mod-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-panel);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.mod-row:hover {
  border-color: var(--accent);
  background: var(--bg-elevated);
}

.mod-handle {
  cursor: grab;
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 0.8rem;
}

.mod-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.mod-name {
  font-weight: 600;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mod-id {
  font-size: 0.65rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.disable-btn {
  color: var(--text-muted);
  font-size: 1.2rem;
  padding: 0 var(--space-1);
}

.disable-btn:hover {
  color: #ef4444;
}
</style>
