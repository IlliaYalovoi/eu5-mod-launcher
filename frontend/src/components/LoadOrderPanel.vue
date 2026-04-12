<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import draggable from 'vuedraggable'
import type { LauncherLayout, Mod } from '../types'
import {
  CreateLauncherCategory,
  DeleteLauncherCategory,
  GetAllMods,
  GetGameActivePlaysetIndex,
  GetLauncherActivePlaysetIndex,
  GetLauncherLayout,
  GetPlaysetNames,
  SaveCompiledLoadOrder,
  SetLauncherLayout,
} from '../../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import AutosortButton from './AutosortButton.vue'
import CycleErrorPanel from './CycleErrorPanel.vue'
import LaunchButton from './LaunchButton.vue'
import BaseButton from './ui/BaseButton.vue'

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
  (event: 'game-changed'): void
}>()

// State owned by this panel
const launcherLayout = ref<LauncherLayout>({ ungrouped: [], categories: [], order: [], collapsed: {} })
const allMods = ref<Mod[]>([])
const playsetNames = ref<string[]>([])
const gameActivePlaysetIndex = ref(-1)

const blocks = ref<EditableBlock[]>([])
const persistError = ref<string | null>(null)
const categoryName = ref('')
const saveError = ref<string | null>(null)
const isSavingCompiled = ref(false)

const globalEditModID = ref('')
const globalEditValue = ref('')
const localEditModID = ref('')
const localEditValue = ref('')
const editingCategoryID = ref('')
const editingCategoryName = ref('')

const modsByID = computed(() => {
  const byID: Record<string, Mod> = {}
  for (const mod of allMods.value) {
    byID[mod.ID] = mod
  }
  return byID
})

const compiledOrder = computed(() => {
  const out: string[] = []
  const seen: Record<string, boolean> = {}
  for (const block of blocks.value) {
    for (const id of block.modIds) {
      if (seen[id]) continue
      seen[id] = true
      out.push(id)
    }
  }
  return out
})

const numberByModID = computed(() => {
  const out: Record<string, number> = {}
  for (let i = 0; i < compiledOrder.value.length; i += 1) {
    out[compiledOrder.value[i]] = i + 1
  }
  return out
})

const localNumberByModID = computed(() => {
  const out: Record<string, number> = {}
  for (const block of blocks.value) {
    if (block.isUngrouped) continue
    for (let i = 0; i < block.modIds.length; i += 1) {
      out[block.modIds[i]] = i + 1
    }
  }
  return out
})

const blockByModID = computed(() => {
  const out: Record<string, EditableBlock> = {}
  for (const block of blocks.value) {
    for (const id of block.modIds) {
      out[id] = block
    }
  }
  return out
})

const activeCountLabel = computed(() => `${compiledOrder.value.length} mods active`)

async function load() {
  try {
    const [layout, mods, names, gameIdx] = await Promise.all([
      GetLauncherLayout(),
      GetAllMods(),
      GetPlaysetNames(),
      GetGameActivePlaysetIndex(),
    ])
    launcherLayout.value = layout as LauncherLayout
    allMods.value = mods as Mod[]
    playsetNames.value = names
    gameActivePlaysetIndex.value = gameIdx
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

function syncBlocksFromLayout() {
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
  blocks.value = next
}

watch(launcherLayout, syncBlocksFromLayout, { immediate: true })

onMounted(load)

function getMod(id: string): Mod | null {
  return modsByID.value[id] || null
}

function modItemKey(value: string): string {
  return value
}

function onItemContextMenu(event: MouseEvent, targetID: string): void {
  event.preventDefault()
  emit('contextmenu', { modID: targetID, x: event.clientX, y: event.clientY })
}

function toLayout(): LauncherLayout {
  const collapsed: Record<string, boolean> = {}
  const order: string[] = []
  const categories: { id: string; name: string; modIds: string[] }[] = []
  let ungrouped: string[] = []

  for (const block of blocks.value) {
    order.push(block.id)
    if (block.collapsed) {
      collapsed[block.id] = true
    }
    if (block.isUngrouped) {
      ungrouped = [...block.modIds]
      continue
    }
    categories.push({ id: block.id, name: block.name, modIds: [...block.modIds] })
  }

  return { ungrouped, categories, order, collapsed }
}

async function persistLayoutAsync(): Promise<void> {
  persistError.value = null
  try {
    await SetLauncherLayout(toLayout() as any)
  } catch (err) {
    persistError.value = errorMessage(err)
    throw err
  }
}

function persistLayout(): void {
  void persistLayoutAsync().catch((err: unknown) => {
    persistError.value = err instanceof Error ? err.message : String(err)
  })
}

function onCycleOpenConstraints(modID: string): void {
  emit('open-constraints', modID)
}

function onModClick(modID: string): void {
  emit('select-mod', modID)
}

function onCreateCategory(): void {
  const name = categoryName.value.trim()
  if (!name) return
  persistError.value = null
  void CreateLauncherCategory(name)
    .then(() => {
      categoryName.value = ''
      return GetLauncherLayout()
    })
    .then((layout) => {
      launcherLayout.value = layout as LauncherLayout
    })
    .catch((err: unknown) => {
      persistError.value = err instanceof Error ? err.message : String(err)
    })
}

function onDeleteCategory(categoryID: string): void {
  persistError.value = null
  void DeleteLauncherCategory(categoryID)
    .then(() => GetLauncherLayout())
    .then((layout) => { launcherLayout.value = layout as LauncherLayout })
    .catch((err: unknown) => { persistError.value = err instanceof Error ? err.message : String(err) })
}

function onToggleCollapse(blockID: string): void {
  const block = blocks.value.find((item) => item.id === blockID)
  if (!block) {
    return
  }
  block.collapsed = !block.collapsed
  persistLayout()
}

function beginGlobalEdit(modID: string): void {
  globalEditModID.value = modID
  globalEditValue.value = String(numberByModID.value[modID] || 1)
  void nextTick().then(() => {
    const input = document.querySelector<HTMLInputElement>(`[data-global-edit="${modID}"]`)
    input?.focus()
    input?.select()
  })
}

function beginLocalEdit(modID: string): void {
  localEditModID.value = modID
  localEditValue.value = String(localNumberByModID.value[modID] || 1)
  void nextTick().then(() => {
    const input = document.querySelector<HTMLInputElement>(`[data-local-edit="${modID}"]`)
    input?.focus()
    input?.select()
  })
}

function cancelGlobalEdit(): void {
  globalEditModID.value = ''
}

function cancelLocalEdit(): void {
  localEditModID.value = ''
}

function beginCategoryEdit(block: EditableBlock): void {
  if (block.isUngrouped) {
    return
  }
  editingCategoryID.value = block.id
  editingCategoryName.value = block.name
  void nextTick().then(() => {
    const input = document.querySelector<HTMLInputElement>(`[data-category-edit="${block.id}"]`)
    input?.focus()
    input?.select()
  })
}

function cancelCategoryEdit(): void {
  editingCategoryID.value = ''
  editingCategoryName.value = ''
}

function confirmCategoryEdit(): void {
  if (!editingCategoryID.value) return
  const trimmed = editingCategoryName.value.trim()
  if (!trimmed) { editingCategoryID.value = ''; return }
  const next = {
    ungrouped: [...launcherLayout.value.ungrouped],
    categories: launcherLayout.value.categories.map((cat) =>
      cat.id === editingCategoryID.value ? { ...cat, name: trimmed } : { ...cat },
    ),
    order: launcherLayout.value.order ? [...launcherLayout.value.order] : undefined,
    collapsed: launcherLayout.value.collapsed ? { ...launcherLayout.value.collapsed } : undefined,
  }
  void SetLauncherLayout(next as any)
    .then(() => GetLauncherLayout())
    .then((layout) => { launcherLayout.value = layout as LauncherLayout; editingCategoryID.value = '' })
    .catch((err: unknown) => { persistError.value = err instanceof Error ? err.message : String(err); editingCategoryID.value = '' })
}

function moveModByGlobalIndex(modID: string, desiredOneBased: number): void {
  if (blocks.value.length === 0) {
    return
  }

  for (const block of blocks.value) {
    const idx = block.modIds.indexOf(modID)
    if (idx >= 0) {
      block.modIds.splice(idx, 1)
    }
  }

  const positions: Array<{ blockIndex: number; localIndex: number; modID: string }> = []
  for (let b = 0; b < blocks.value.length; b += 1) {
    const mods = blocks.value[b].modIds
    for (let i = 0; i < mods.length; i += 1) {
      positions.push({ blockIndex: b, localIndex: i, modID: mods[i] })
    }
  }

  const target = Math.max(1, Math.min(desiredOneBased, positions.length + 1))
  if (target === positions.length + 1) {
    blocks.value[blocks.value.length - 1].modIds.push(modID)
    return
  }

  const anchor = positions[target - 1]
  blocks.value[anchor.blockIndex].modIds.splice(anchor.localIndex, 0, modID)
}

function confirmGlobalEdit(modID: string): void {
  const parsed = Number.parseInt(globalEditValue.value, 10)
  if (Number.isNaN(parsed)) {
    globalEditModID.value = ''
    return
  }

  moveModByGlobalIndex(modID, parsed)
  globalEditModID.value = ''
  persistLayout()
}

function confirmLocalEdit(modID: string): void {
  const parsed = Number.parseInt(localEditValue.value, 10)
  if (Number.isNaN(parsed)) {
    localEditModID.value = ''
    return
  }

  const block = blockByModID.value[modID]
  if (!block) {
    localEditModID.value = ''
    return
  }

  const currentIndex = block.modIds.indexOf(modID)
  if (currentIndex < 0) {
    localEditModID.value = ''
    return
  }

  block.modIds.splice(currentIndex, 1)
  const target = Math.max(1, Math.min(parsed, block.modIds.length + 1))
  block.modIds.splice(target - 1, 0, modID)

  localEditModID.value = ''
  persistLayout()
}

function onSaveCompiled(): void {
  saveError.value = null
  isSavingCompiled.value = true
  void persistLayoutAsync()
    .then(() => SaveCompiledLoadOrder())
    .catch((err: unknown) => { saveError.value = err instanceof Error ? err.message : String(err) })
    .finally(() => { isSavingCompiled.value = false })
}
</script>

<template>
  <section class="load-order-panel" aria-label="Load order panel">
    <header class="head">
      <div>
        <h2 class="title">Load Order</h2>
        <p class="count">{{ activeCountLabel }}</p>
      </div>
      <div class="head-actions">
        <LaunchButton :playset-names="playsetNames" :game-active-playset-index="gameActivePlaysetIndex" />
        <AutosortButton @sorted="load" />
        <BaseButton :loading="isSavingCompiled" @click="onSaveCompiled">Save to Game</BaseButton>
      </div>
    </header>

    <div class="category-create">
      <input
        v-model="categoryName"
        class="category-input"
        type="text"
        placeholder="New category name..."
        @keydown.enter.prevent="onCreateCategory"
      />
      <BaseButton variant="ghost" @click="onCreateCategory">+ Category</BaseButton>
    </div>

    <p v-if="persistError" class="alert">{{ persistError }}</p>
    <p v-else-if="saveError" class="alert">{{ saveError }}</p>

    <div class="list-wrap">
      <draggable v-model="blocks" item-key="id" handle=".category-handle" :animation="150" @end="persistLayout">
        <template #item="{ element: block }">
          <section
            v-if="block.modIds.length > 0 || block.isUngrouped"
            class="bucket category-block"
            @contextmenu="onItemContextMenu($event, block.id)"
          >
            <div class="category-head">
              <button class="category-handle" type="button" aria-label="Drag category">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <circle cx="9" cy="6" r="1.5" /><circle cx="15" cy="6" r="1.5" />
                  <circle cx="9" cy="12" r="1.5" /><circle cx="15" cy="12" r="1.5" />
                  <circle cx="9" cy="18" r="1.5" /><circle cx="15" cy="18" r="1.5" />
                </svg>
              </button>
              <span class="category-dot" aria-hidden="true" />

              <template v-if="editingCategoryID === block.id">
                <input
                  v-model="editingCategoryName"
                  :data-category-edit="block.id"
                  class="category-name-input"
                  type="text"
                  @keydown.enter.prevent="confirmCategoryEdit"
                  @keydown.esc.prevent="cancelCategoryEdit"
                />
                <button class="confirm" type="button" @click="confirmCategoryEdit">✓</button>
              </template>
              <template v-else>
                <button class="category-name-btn" type="button" @click="beginCategoryEdit(block)">
                  <h3 class="bucket-title">{{ block.name }}</h3>
                </button>
              </template>

              <span class="category-count">{{ block.modIds.length }}</span>
              <button class="fold" type="button" @click="onToggleCollapse(block.id)">{{ block.collapsed ? '+' : '-' }}</button>
              <button v-if="!block.isUngrouped" class="delete-category" type="button" @click="onDeleteCategory(block.id)" aria-label="Delete category">×</button>
            </div>

            <draggable
              v-if="!block.collapsed"
              v-model="block.modIds"
              :item-key="modItemKey"
              :group="{ name: 'mods' }"
              handle=".mod-handle"
              :animation="150"
              @end="persistLayout"
            >
              <template #item="{ element: modID }">
                <article class="mod-row" @contextmenu.stop.prevent="onItemContextMenu($event, modID)" @click="onModClick(modID)">
                  <button class="mod-handle" type="button" aria-label="Drag mod">
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                      <circle cx="9" cy="6" r="1.5" /><circle cx="15" cy="6" r="1.5" />
                      <circle cx="9" cy="12" r="1.5" /><circle cx="15" cy="12" r="1.5" />
                      <circle cx="9" cy="18" r="1.5" /><circle cx="15" cy="18" r="1.5" />
                    </svg>
                  </button>

                  <div class="mod-number-cell">
                    <template v-if="globalEditModID === modID">
                      <input
                        v-model="globalEditValue"
                        :data-global-edit="modID"
                        class="number-input"
                        type="number"
                        min="1"
                        @keydown.enter.prevent="confirmGlobalEdit(modID)"
                        @keydown.esc.prevent="cancelGlobalEdit"
                      />
                      <button class="confirm" type="button" @click="confirmGlobalEdit(modID)">✓</button>
                    </template>
                    <button v-else class="number-btn" type="button" @click.stop="beginGlobalEdit(modID)">{{ numberByModID[modID] }}</button>
                  </div>

                  <div class="mod-local-number-cell">
                    <template v-if="!block.isUngrouped">
                      <template v-if="localEditModID === modID">
                        <input
                          v-model="localEditValue"
                          :data-local-edit="modID"
                          class="number-input secondary"
                          type="number"
                          min="1"
                          @keydown.enter.prevent="confirmLocalEdit(modID)"
                          @keydown.esc.prevent="cancelLocalEdit"
                        />
                        <button class="confirm" type="button" @click="confirmLocalEdit(modID)">✓</button>
                      </template>
                      <button v-else class="number-btn secondary" type="button" @click.stop="beginLocalEdit(modID)">
                        {{ localNumberByModID[modID] }}
                      </button>
                    </template>
                  </div>

                  <span class="mod-name">{{ getMod(modID)?.Name || modID }}</span>
                </article>
              </template>
            </draggable>

            <div v-if="!block.collapsed && block.modIds.length === 0" class="empty-hint">
              Drop mods here
            </div>
          </section>
        </template>
      </draggable>
    </div>

    <CycleErrorPanel @open-constraints="onCycleOpenConstraints" />
  </section>
</template>

<style scoped>
.load-order-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  height: 100%;
  min-height: 0;
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.head-actions {
  display: flex;
  gap: var(--space-2);
}

.title {
  font-family: var(--font-display), serif;
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.count {
  margin-top: var(--space-1);
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.category-create {
  display: flex;
  gap: var(--space-2);
}

.category-input {
  flex: 1;
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
}

.alert {
  padding: var(--space-3);
  border: var(--border-width) solid var(--color-danger);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-danger);
  font-size: 0.85rem;
}

.list-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding-right: var(--space-1);
}

.bucket {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
}

.bucket-title {
  color: var(--color-text-secondary);
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.category-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--cat-1);
  flex-shrink: 0;
}

.category-count {
  margin-left: auto;
  margin-right: var(--space-2);
  padding: 0.1rem 0.4rem;
  border-radius: var(--radius-pill);
  background: var(--color-bg-base);
  color: var(--color-text-muted);
  font-size: 0.7rem;
  font-family: var(--font-mono), monospace;
}

.category-head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.fold {
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  min-width: 1.5rem;
  min-height: 1.5rem;
  cursor: pointer;
}

.delete-category {
  margin-left: auto;
  border: 0;
  background: transparent;
  color: var(--color-danger);
  cursor: pointer;
}

.category-handle,
.mod-handle {
  border: 0;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: grab;
}

.mod-row {
  display: grid;
  grid-template-columns: 1.5rem 3.2rem 3.2rem 1fr;
  gap: var(--space-2);
  align-items: center;
  min-height: 2rem;
  padding: var(--space-2);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast);
  border-left: 3px solid transparent;
}

.mod-row:hover {
  border-color: var(--color-border);
  border-left-color: var(--color-accent);
  background: var(--color-bg-base);
}

.mod-number {
  font-family: var(--font-mono), monospace;
  color: var(--color-text-muted);
  text-align: right;
}

.mod-number-cell,
.mod-local-number-cell {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-1);
}

.number-btn {
  border: 0;
  background: transparent;
  color: var(--color-text-muted);
  font-family: var(--font-mono), monospace;
  cursor: pointer;
}

.number-btn.secondary {
  opacity: 0.65;
}

.number-input {
  width: 2.6rem;
  min-height: 1.6rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
  text-align: center;
  padding: 0 var(--space-1);
}

.number-input.secondary {
  opacity: 0.8;
}

.confirm {
  border: 0;
  background: transparent;
  color: var(--color-success);
  cursor: pointer;
}

.category-name-btn {
  border: 0;
  background: transparent;
  padding: 0;
  cursor: pointer;
  text-align: left;
  flex: 1;
  min-width: 0;
}

.category-name-input {
  flex: 1;
  min-width: 0;
  min-height: 1.5rem;
  padding: 0.1rem var(--space-2);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.empty-hint {
  padding: var(--space-3);
  text-align: center;
  color: var(--color-text-muted);
  font-size: 0.8rem;
  border: var(--border-width) dashed var(--color-border);
  border-radius: var(--radius-sm);
}

.mod-name {
  color: var(--color-text-primary);
}
</style>


