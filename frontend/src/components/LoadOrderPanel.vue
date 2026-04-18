<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import type { LauncherLayout, Mod } from '../types'
import AutosortButton from './AutosortButton.vue'
import CycleErrorPanel from './CycleErrorPanel.vue'
import LoadOrderItem from './LoadOrderItem.vue'
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
  (event: 'manage-groups'): void
}>()

const loadOrderStore = useLoadOrderStore()
const modsStore = useModsStore()

const { launcherLayout } = storeToRefs(loadOrderStore)
const { allMods } = storeToRefs(modsStore)

const blocks = ref<EditableBlock[]>([])
const persistError = ref<string | null>(null)
const saveError = ref<string | null>(null)

watch(
  () => [launcherLayout.value, allMods.value] as const,
  ([value, modsValue]) => {
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
          modIds: value.ungrouped.filter(modID => {
            const mod = modsValue.find(m => m.ID === modID)
            return mod && mod.Enabled
          }),
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
        modIds: category.modIds.filter(modID => {
          const mod = modsValue.find(m => m.ID === modID)
          return mod && mod.Enabled
        }),
        isUngrouped: false,
        collapsed: !!collapsed[category.id],
      })
    }

    if (!seen[ungroupedID]) {
      next.unshift({
        id: ungroupedID,
        name: 'Ungrouped',
        modIds: value.ungrouped.filter(modID => {
          const mod = modsValue.find(m => m.ID === modID)
          return mod && mod.Enabled
        }),
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
        modIds: category.modIds.filter(modID => {
          const mod = modsValue.find(m => m.ID === modID)
          return mod && mod.Enabled
        }),
        isUngrouped: false,
        collapsed: !!collapsed[category.id],
      })
    }

    blocks.value = next
  },
  { immediate: true, deep: true },
)

onMounted(() => {
  if (launcherLayout.value.ungrouped.length === 0 && launcherLayout.value.categories.length === 0) {
    void loadOrderStore.fetch()
  }
})

function onItemContextMenu(event: MouseEvent, targetID: string): void {
  event.preventDefault()
  emit('contextmenu', { modID: targetID, x: event.clientX, y: event.clientY })
}

function onModContextMenu(payload: { modID: string; x: number; y: number }): void {
  emit('contextmenu', payload)
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

function onToggleCollapse(blockID: string): void {
  const block = blocks.value.find((item) => item.id === blockID)
  if (!block) {
    return
  }
  block.collapsed = !block.collapsed
  persistLayout()
}

function selectMod(modID: string) {
  modsStore.selectMod(modID)
}
</script>

<template>
  <div class="main-wrapper">
    <header class="toolbar">
      <h1 class="playset-title">Load Order</h1>
      <div class="toolbar-actions">
        <BaseButton variant="ghost" class="action-btn" @click="emit('manage-groups')">MANAGE GROUPS</BaseButton>
        <AutosortButton />
      </div>
    </header>

    <div class="view-content">
      <div class="group-container">
        <CycleErrorPanel @open-constraints="emit('open-constraints', $event)" />

        <p v-if="persistError" class="alert">{{ persistError }}</p>
        <p v-else-if="saveError" class="alert">{{ saveError }}</p>

        <div class="list-wrap">
          <section v-for="block in blocks" :key="block.id" class="mod-group" @contextmenu="onItemContextMenu($event, block.id)">
            <div class="group-header">
              <div class="header-left">
                <strong>{{ block.name }}</strong>
                <span v-if="block.isUngrouped" class="group-rule">Priority: Default</span>
                <span v-else class="group-rule">Priority: {{ block.id }}</span>
              </div>
              <div class="header-actions">
                <button class="fold" type="button" @click="onToggleCollapse(block.id)">{{ block.collapsed ? '+' : '-' }}</button>
              </div>
            </div>

            <div v-if="!block.collapsed" class="mod-list">
              <LoadOrderItem
                v-for="modID in block.modIds"
                :key="modID"
                :mod-i-d="modID"
                @contextmenu="onModContextMenu"
                @open-constraints="emit('open-constraints', modID)"
                @select="selectMod"
              />
            </div>
          </section>
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
  font-size: 1.5rem;
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
  text-transform: uppercase;
  font-family: var(--font-body);
  font-size: 0.9rem;
}

.view-content {
  flex: 1;
  min-height: 0;
  padding: 20px 40px;
  overflow: hidden;
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 30px;
}

.group-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  padding-right: 10px;
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

.group-rule {
  font-size: 0.75rem;
  color: var(--color-accent);
  text-transform: uppercase;
  margin-left: 10px;
}

.list-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.alert {
  padding: var(--space-3);
  border: var(--border-width) solid var(--color-danger);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-danger);
  font-size: 0.95rem;
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
</style>
