<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  AddConstraint, AddLoadFirst, AddLoadLast,
  GetAllMods, GetConstraints, GetLauncherLayout,
  RemoveConstraint, RemoveLoadFirst, RemoveLoadLast,
} from '../wailsjs/go/launcher/App'
import type { Constraint, LauncherLayout, Mod } from '../types'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import BaseModal from './ui/BaseModal.vue'
import BaseButton from './ui/BaseButton.vue'
import ModPicker from './ui/ModPicker.vue'

type ExistingRow = {
  key: string
  text: string
  type: string
  from: string
  to: string
  modID: string
}

const props = defineProps<{ open: boolean; modID: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const allMods = ref<Mod[]>([])
const constraints = ref<Constraint[]>([])
const launcherLayout = ref<LauncherLayout>({ ungrouped: [], categories: [], order: [], collapsed: {} })
const direction = ref<'after' | 'before' | 'first' | 'last'>('after')
const pickedModID = ref<string | null>(null)
const addError = ref<string | null>(null)

const isCategoryTarget = computed(() => props.modID.indexOf('category:') === 0)

const modsByID = computed(() => {
  const byID: Record<string, Mod> = {}
  for (const mod of allMods.value) byID[mod.ID] = mod
  return byID
})

const categoryNames = computed(() => {
  const byID: Record<string, string> = { 'category:ungrouped': 'Ungrouped' }
  for (const cat of launcherLayout.value.categories) byID[cat.id] = cat.name
  return byID
})

const currentTargetName = computed(() => modsByID.value[props.modID]?.Name || categoryNames.value[props.modID] || props.modID)
const targetNoun = computed(() => props.modID.indexOf('category:') === 0 ? 'category' : 'mod')
const subjectLabel = computed(() => `This ${targetNoun.value}`)

const currentConstraints = computed(() =>
  constraints.value.filter(c => c.from === props.modID || c.to === props.modID || c.modId === props.modID)
)

const existingRows = computed(() => {
  return currentConstraints.value.map((c): ExistingRow => {
    const type = c.type ?? 'after'
    if (type === 'first') return { key: `first:${c.modId}`, text: `${subjectLabel.value} is marked to load first`, type, from: '', to: '', modID: c.modId || '' }
    if (type === 'last') return { key: `last:${c.modId}`, text: `${subjectLabel.value} is marked to load last`, type, from: '', to: '', modID: c.modId || '' }
    if (c.from === props.modID) {
      const t = c.to || ''
      return { key: `${c.from}->${t}`, text: `${subjectLabel.value} always loads after ${modsByID.value[t]?.Name || categoryNames.value[t] || t}`, type, from: c.from || '', to: t, modID: '' }
    }
    const s = c.from || '', t = c.to || ''
    return { key: `${s}->${t}`, text: `${modsByID.value[s]?.Name || categoryNames.value[s] || s} always loads after ${subjectLabel.value.toLowerCase()}`, type, from: s, to: t, modID: '' }
  })
})

const availableMods = computed(() => {
  const blocked: Record<string, boolean> = {}
  for (const c of currentConstraints.value) {
    const type = c.type ?? 'after'
    if (type !== 'after') continue
    if (direction.value === 'after' && c.from === props.modID && c.to) blocked[c.to] = true
    if (direction.value === 'before' && c.to === props.modID && c.from) blocked[c.from] = true
  }
  const result: Mod[] = []
  if (isCategoryTarget.value) {
    if (props.modID !== 'category:ungrouped' && !blocked['category:ungrouped']) result.push({ ID: 'category:ungrouped', Name: '[Category] Ungrouped', Version: '', Tags: [], Description: '', ThumbnailPath: '', DirPath: '', Enabled: true })
    for (const cat of launcherLayout.value.categories) {
      if (cat.id === props.modID || blocked[cat.id]) continue
      result.push({ ID: cat.id, Name: `[Category] ${cat.name}`, Version: '', Tags: [], Description: '', ThumbnailPath: '', DirPath: '', Enabled: true })
    }
    return result
  }
  for (const mod of allMods.value) {
    if (mod.ID === props.modID || blocked[mod.ID]) continue
    result.push(mod)
  }
  return result
})

async function load() {
  const [c, mods, layout] = await Promise.all([GetConstraints(), GetAllMods(), GetLauncherLayout()])
  constraints.value = c as Constraint[]
  allMods.value = mods as Mod[]
  launcherLayout.value = layout as LauncherLayout
}

watch(() => props.open, (isOpen) => {
  if (!isOpen) return
  direction.value = 'after'
  pickedModID.value = null
  addError.value = null
  void load()
})

async function onDelete(row: ExistingRow) {
  try {
    if (row.type === 'first') await RemoveLoadFirst(row.modID)
    else if (row.type === 'last') await RemoveLoadLast(row.modID)
    else await RemoveConstraint(row.from, row.to)
    await load()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

async function onAddConstraint() {
  addError.value = null
  if (!pickedModID.value && (direction.value === 'after' || direction.value === 'before')) {
    addError.value = 'Select a mod first.'
    return
  }
  try {
    if (direction.value === 'first') await AddLoadFirst(props.modID)
    else if (direction.value === 'last') await AddLoadLast(props.modID)
    else if (direction.value === 'after') await AddConstraint(props.modID, pickedModID.value as string)
    else await AddConstraint(pickedModID.value as string, props.modID)
    pickedModID.value = null
    await load()
  } catch (err) {
    addError.value = errorMessage(err)
  }
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

.error {
  color: var(--color-danger);
  font-size: 0.85rem;
}
</style>


