<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Mod } from '../../types'

const props = defineProps<{
  mods: Mod[]
  modelValue: string | null
}>()

const emit = defineEmits<{
  (event: 'update:modelValue', value: string | null): void
}>()

const query = ref('')
const isOpen = ref(false)
const activeIndex = ref(0)

const selectedMod = computed(() => {
  for (const mod of props.mods) {
    if (mod.id === props.modelValue) {
      return mod
    }
  }
  return null
})

const filteredMods = computed(() => {
  const term = query.value.trim().toLowerCase()
  if (!term) {
    return props.mods
  }
  return props.mods.filter((mod) => mod.name.toLowerCase().indexOf(term) >= 0)
})

watch(
  () => props.modelValue,
  () => {
    if (!isOpen.value) {
      query.value = selectedMod.value ? selectedMod.value.name : ''
    }
  },
  { immediate: true },
)

watch(filteredMods, (mods) => {
  if (mods.length === 0) {
    activeIndex.value = 0
    return
  }
  if (activeIndex.value >= mods.length) {
    activeIndex.value = mods.length - 1
  }
})

function openList(): void {
  isOpen.value = true
  activeIndex.value = 0
}

function closeList(): void {
  isOpen.value = false
  if (!query.value.trim() && selectedMod.value) {
    query.value = selectedMod.value.name
  }
}

function selectMod(mod: Mod): void {
  emit('update:modelValue', mod.id)
  query.value = mod.name
  isOpen.value = false
}

function onInput(event: Event): void {
  const target = event.target as HTMLInputElement
  query.value = target.value
  isOpen.value = true
}

function onKeydown(event: KeyboardEvent): void {
  if (!isOpen.value && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
    openList()
    return
  }

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (filteredMods.value.length > 0) {
      activeIndex.value = Math.min(activeIndex.value + 1, filteredMods.value.length - 1)
    }
    return
  }

  if (event.key === 'ArrowUp') {
    event.preventDefault()
    if (filteredMods.value.length > 0) {
      activeIndex.value = Math.max(activeIndex.value - 1, 0)
    }
    return
  }

  if (event.key === 'Enter' && isOpen.value && filteredMods.value.length > 0) {
    event.preventDefault()
    selectMod(filteredMods.value[activeIndex.value])
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    closeList()
  }
}
</script>

<template>
  <div class="mod-picker">
    <input
      :value="query"
      class="picker-input"
      type="text"
      placeholder="Search mods..."
      spellcheck="false"
      @focus="openList"
      @blur="closeList"
      @input="onInput"
      @keydown="onKeydown"
    />

    <div v-if="isOpen" class="dropdown" role="listbox">
      <button
        v-for="(mod, index) in filteredMods"
        :key="mod.id"
        class="option"
        :class="{ 'option--active': index === activeIndex }"
        type="button"
        @mousedown.prevent="selectMod(mod)"
      >
          {{ mod.name }}
      </button>
      <p v-if="filteredMods.length === 0" class="empty">No matching mods</p>
    </div>
  </div>
</template>

<style scoped>
.mod-picker {
  position: relative;
}

.picker-input {
  width: 100%;
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
}

.picker-input:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.dropdown {
  position: absolute;
  z-index: 20;
  top: calc(100% + var(--space-1));
  left: 0;
  right: 0;
  max-height: 12rem;
  overflow: auto;
  padding: var(--space-1);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
}

.option {
  display: block;
  width: 100%;
  padding: var(--space-2) var(--space-3);
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-primary);
  text-align: left;
  cursor: pointer;
}

.option:hover,
.option--active {
  background: var(--color-bg-panel);
}

.empty {
  padding: var(--space-2) var(--space-3);
  color: var(--color-text-muted);
  font-size: 0.85rem;
}
</style>


