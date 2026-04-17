<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import draggable from 'vuedraggable'
import { storeToRefs } from 'pinia'
import type { LauncherLayout, Mod } from '../types'
import AutosortButton from './AutosortButton.vue'
import CycleErrorPanel from './CycleErrorPanel.vue'
import { useLoadOrderStore } from '../stores/loadorder'
import { useModsStore } from '../stores/mods'
import BaseButton from './ui/BaseButton.vue'
import ModListPanel from './ModListPanel.vue'

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
}>()

const loadOrderStore = useLoadOrderStore()
const modsStore = useModsStore()

const { launcherLayout } = storeToRefs(loadOrderStore)
const { allMods } = storeToRefs(modsStore)

const blocks = ref<EditableBlock[]>([])
const persistError = ref<string | null>(null)
const categoryName = ref('')
const saveError = ref<string | null>(null)
const isSavingCompiled = ref(false)

const globalEditModID = ref('')
const globalEditValue = ref('')
const localEditModID = ref('')
const localEditValue = ref('')

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
      if (seen[id]) {
        continue
      }
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
    if (block.isUngrouped) {
      continue
    }
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

watch(
  launcherLayout,
  (value) => {
    const collapsed = value.collapsed || {}
    const categoryByID: Record<string, { id: string; name: string; modIds: string[] }> = {}
    for (const category of value.categories) {
      categoryByID[category.id] = {
        id: category.id,
        name: category.name,
        modIds: [...category.modIds],
      }
    }

    const order = value.order && value.order.length > 0 ? [...value.order] : [ungroupedID, ...value.categories.map((c) => c.id)]
    const next: EditableBlock[] = []
    const seen: Record<string, boolean> = {}

    for (const id of order) {
      if (seen[id]) {
        continue
      }
      seen[id] = true

      if (id === ungroupedID) {
        next.push({
          id: ungroupedID,
          name: 'Ungrouped',
          modIds: [...value.ungrouped],
          isUngrouped: true,
          collapsed: !!collapsed[ungroupedID],
        })
        continue
      }

      const category = categoryByID[id]
      if (!category) {
        continue
      }
      next.push({
        id: category.id,
        name: category.name,
        modIds: [...category.modIds],
        isUngrouped: false,
        collapsed: !!collapsed[category.id],
      })
    }

    if (!seen[ungroupedID]) {
      next.unshift({
        id: ungroupedID,
        name: 'Ungrouped',
        modIds: [...value.ungrouped],
        isUngrouped: true,
        collapsed: !!collapsed[ungroupedID],
      })
    }

    for (const category of value.categories) {
      if (seen[category.id]) {
        continue
      }
      next.push({
        id: category.id,
        name: category.name,
        modIds: [...category.modIds],
        isUngrouped: false,
        collapsed: !!collapsed[category.id],
      })
    }

    blocks.value = next
  },
  { immediate: true },
)

onMounted(() => {
  if (launcherLayout.value.ungrouped.length === 0 && launcherLayout.value.categories.length === 0) {
    void loadOrderStore.fetch()
  }
})

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
  await loadOrderStore.persistLauncherLayout(toLayout())
}

function persistLayout(): void {
  void persistLayoutAsync().catch((err: unknown) => {
    persistError.value = err instanceof Error ? err.message : String(err)
  })
}

function onCycleOpenConstraints(modID: string): void {
  emit('open-constraints', modID)
}

function onCreateCategory(): void {
  const name = categoryName.value.trim()
  if (!name) {
    return
  }
  persistError.value = null
  void loadOrderStore
    .createCategory(name)
    .then(() => {
      categoryName.value = ''
    })
    .catch((err: unknown) => {
      persistError.value = err instanceof Error ? err.message : String(err)
    })
}

function onDeleteCategory(categoryID: string): void {
  persistError.value = null
  void loadOrderStore.deleteCategory(categoryID).catch((err: unknown) => {
    persistError.value = err instanceof Error ? err.message : String(err)
  })
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
    .then(() => loadOrderStore.saveCompiled())
    .catch((err: unknown) => {
      saveError.value = err instanceof Error ? err.message : String(err)
    })
    .finally(() => {
      isSavingCompiled.value = false
    })
}
</script>

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
        <CycleErrorPanel @open-constraints="onCycleOpenConstraints" />

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
                    <article class="mod-row" @contextmenu.stop.prevent="onItemContextMenu($event, modID)">
                      <button class="mod-handle" type="button" aria-label="Drag mod">☰</button>

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
                        <button v-else class="number-btn" type="button" @click="beginGlobalEdit(modID)">{{ numberByModID[modID] }}</button>
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
                          <button v-else class="number-btn secondary" type="button" @click="beginLocalEdit(modID)">
                            {{ localNumberByModID[modID] }}
                          </button>
                        </template>
                      </div>

                      <span class="mod-name">{{ getMod(modID)?.Name || modID }}</span>
                    </article>
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

<style scoped>
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

.mod-name {
  color: var(--color-text-primary);
}
</style>