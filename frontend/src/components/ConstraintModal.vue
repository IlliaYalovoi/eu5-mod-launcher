<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import BaseModal from './ui/BaseModal.vue'
import BaseButton from './ui/BaseButton.vue'
import ModPicker from './ui/ModPicker.vue'
import { useConstraintsStore } from '../stores/constraints'
import { useLoadOrderStore } from '../stores/loadorder'
import { useModsStore } from '../stores/mods'
import type { Mod } from '../types'

type ExistingRow = {
  key: string
  text: string
  type: string
  from: string
  to: string
  modID: string
}

const props = defineProps<{
  open: boolean
  modID: string
}>()

const emit = defineEmits<{ (event: 'close'): void }>()

const constraintsStore = useConstraintsStore()
const loadOrderStore = useLoadOrderStore()
const modsStore = useModsStore()
const { allMods } = storeToRefs(modsStore)
const { launcherLayout } = storeToRefs(loadOrderStore)

const direction = ref<'after' | 'before' | 'first' | 'last'>('after')
const pickedModID = ref<string | null>(null)
const addError = ref<string | null>(null)
const isCategoryTarget = computed(() => props.modID.indexOf('category:') === 0)

const modsByID = computed(() => {
  const byID: Record<string, Mod> = {}
  for (const mod of allMods.value) {
    byID[mod.ID] = mod
  }
  return byID
})

const categoryNames = computed(() => {
  const byID: Record<string, string> = {
    'category:ungrouped': 'Ungrouped',
  }
  for (const category of launcherLayout.value.categories) {
    byID[category.id] = category.name
  }
  return byID
})

const modCategoryByID = computed<Record<string, string>>(() => {
  const byID: Record<string, string> = {}
  for (const modID of launcherLayout.value.ungrouped) {
    byID[modID] = 'category:ungrouped'
  }
  for (const category of launcherLayout.value.categories) {
    for (const modID of category.modIds) {
      byID[modID] = category.id
    }
  }
  return byID
})

const currentCategoryID = computed(() => modCategoryByID.value[props.modID] || '')

const currentTargetName = computed(() => {
  return modsByID.value[props.modID]?.Name || categoryNames.value[props.modID] || props.modID
})

const targetNoun = computed(() => (props.modID.indexOf('category:') === 0 ? 'category' : 'mod'))
const subjectLabel = computed(() => `This ${targetNoun.value}`)

const currentConstraints = computed(() => constraintsStore.forMod(props.modID))

const existingRows = computed(() => {
  return currentConstraints.value.map((constraint): ExistingRow => {
    const type = constraint.type ?? 'after'

    if (type === 'first') {
      return {
        key: `first:${constraint.modId}`,
        text: `${subjectLabel.value} is marked to load first`,
        type,
        from: '',
        to: '',
        modID: constraint.modId || '',
      }
    }

    if (type === 'last') {
      return {
        key: `last:${constraint.modId}`,
        text: `${subjectLabel.value} is marked to load last`,
        type,
        from: '',
        to: '',
        modID: constraint.modId || '',
      }
    }

    if (constraint.from === props.modID) {
      const targetID = constraint.to || ''
      return {
        key: `${constraint.from || ''}->${targetID}`,
        text: `${subjectLabel.value} always loads after ${modsByID.value[targetID]?.Name || categoryNames.value[targetID] || targetID}`,
        type,
        from: constraint.from || '',
        to: targetID,
        modID: '',
      }
    }

    const sourceID = constraint.from || ''
    const targetID = constraint.to || ''

    return {
      key: `${sourceID}->${targetID}`,
      text: `${modsByID.value[sourceID]?.Name || categoryNames.value[sourceID] || sourceID} always loads after ${subjectLabel.value.toLowerCase()}`,
      type,
      from: sourceID,
      to: targetID,
      modID: '',
    }
  })
})

const availableMods = computed(() => {
  const blocked: Record<string, boolean> = {}

  for (const constraint of currentConstraints.value) {
    const type = constraint.type ?? 'after'
    if (type !== 'after') {
      continue
    }
    if (direction.value === 'after' && constraint.from === props.modID && constraint.to) {
      blocked[constraint.to] = true
    }
    if (direction.value === 'before' && constraint.to === props.modID && constraint.from) {
      blocked[constraint.from] = true
    }
  }

  const result: Mod[] = []
  if (isCategoryTarget.value) {
    if (props.modID !== 'category:ungrouped' && !blocked['category:ungrouped']) {
      result.push({
        ID: 'category:ungrouped',
        Name: '[Category] Ungrouped',
        Version: '',
        SupportedVersion: '',
        IsCompatible: true,
        Tags: [],
        Description: '',
        ThumbnailPath: '',
        DirPath: '',
        Enabled: true,
      })
    }

    for (const category of launcherLayout.value.categories) {
      if (category.id === props.modID) {
        continue
      }
      if (blocked[category.id]) {
        continue
      }
      result.push({
        ID: category.id,
        Name: `[Category] ${category.name}`,
        Version: '',
        SupportedVersion: '',
        IsCompatible: true,
        Tags: [],
        Description: '',
        ThumbnailPath: '',
        DirPath: '',
        Enabled: true,
      })
    }
    return result
  }

  for (const mod of allMods.value) {
    if (mod.ID === props.modID) {
      continue
    }
    if (blocked[mod.ID]) {
      continue
    }
    if (!currentCategoryID.value) {
      continue
    }
    if (modCategoryByID.value[mod.ID] !== currentCategoryID.value) {
      continue
    }
    result.push(mod)
  }

  return result
})

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) {
      return
    }
    direction.value = 'after'
    pickedModID.value = null
    addError.value = null
    void constraintsStore.fetch()
  },
)

function onDelete(row: ExistingRow): void {
  if (row.type === 'first') {
    void constraintsStore.removeLoadFirst(row.modID)
    return
  }
  if (row.type === 'last') {
    void constraintsStore.removeLoadLast(row.modID)
    return
  }
  void constraintsStore.remove(row.from, row.to)
}

function onAddConstraint(): void {
  addError.value = null
  if (!pickedModID.value && (direction.value === 'after' || direction.value === 'before')) {
    addError.value = 'Select a mod first.'
    return
  }

  const pending =
    direction.value === 'after'
      ? constraintsStore.add(props.modID, pickedModID.value as string)
      : direction.value === 'before'
        ? constraintsStore.add(pickedModID.value as string, props.modID)
        : direction.value === 'first'
          ? constraintsStore.addLoadFirst(props.modID)
          : constraintsStore.addLoadLast(props.modID)

  void pending
    .then(() => {
      pickedModID.value = null
    })
    .catch((err: unknown) => {
      addError.value = err instanceof Error ? err.message : String(err)
    })
}
</script>

<template>
  <BaseModal :open="open" @close="emit('close')">
    <div class="modal-body">
      <header class="modal-head">
        <h2 class="title">Constraints for {{ currentTargetName }}</h2>
        <button class="close-button" type="button" aria-label="Close" @click="emit('close')">×</button>
      </header>

      <section class="section">
        <h3 class="section-title">Existing constraints</h3>
        <p v-if="existingRows.length === 0" class="empty">No constraints for this {{ targetNoun }} yet.</p>
        <div v-else class="list">
          <div v-for="row in existingRows" :key="row.key" class="row">
            <span class="row-text">{{ row.text }}</span>
            <button class="delete" type="button" aria-label="Delete constraint" @click="onDelete(row)">
              ×
            </button>
          </div>
        </div>
      </section>

      <section class="section">
        <h3 class="section-title">Add constraint</h3>
        <div class="form-row">
          <span class="label">This {{ targetNoun }} loads</span>
          <select v-model="direction" class="direction">
            <option value="after">after</option>
            <option value="before">before</option>
            <option value="first">first</option>
            <option value="last">last</option>
          </select>
          <ModPicker
            v-if="direction === 'after' || direction === 'before'"
            v-model="pickedModID"
            class="picker"
            :mods="availableMods"
          />
          <span v-else class="fixed-target">(this {{ targetNoun }})</span>
        </div>
        <BaseButton variant="primary" :disabled="(direction === 'after' || direction === 'before') && !pickedModID" @click="onAddConstraint">Add</BaseButton>
        <p v-if="!isCategoryTarget" class="hint">Only mods in same category can be constrained.</p>
        <p v-if="addError" class="error">{{ addError }}</p>
      </section>
    </div>
  </BaseModal>
</template>

<style scoped>
.modal-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.title {
  font-family: var(--font-display), serif;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.close-button {
  width: 2rem;
  height: 2rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.close-button:hover {
  background: var(--color-bg-panel);
}

.section {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.section-title {
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.empty {
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

.list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
}

.row-text {
  color: var(--color-text-primary);
  font-size: 0.9rem;
}

.delete {
  border: 0;
  background: transparent;
  color: var(--color-danger);
  font-size: 1rem;
  cursor: pointer;
}

.form-row {
  display: grid;
  grid-template-columns: auto auto 1fr;
  gap: var(--space-2);
  align-items: center;
}

.label {
  color: var(--color-text-secondary);
}

.direction {
  min-height: 2.25rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
  padding: var(--space-2);
}

.picker {
  min-width: 0;
}

.fixed-target {
  color: var(--color-text-secondary);
  padding: 0 var(--space-2);
}

.hint {
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.error {
  color: var(--color-danger);
  font-size: 0.85rem;
}
</style>


