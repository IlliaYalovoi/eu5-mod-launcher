<script setup lang="ts">
import type { Mod } from '../types'
import ModToggle from './ui/ModToggle.vue'

const props = defineProps<{
  mod: Mod
  selected?: boolean
}>()

const emit = defineEmits<{
  (event: 'toggle', value: boolean): void
  (event: 'select'): void
}>()

function onToggle(value: boolean): void {
  emit('toggle', value)
}

function onSelect(): void {
  emit('select')
}
</script>

<template>
  <div class="disabled-mod" :class="{ selected: props.selected }" @click="onSelect">
    <span class="name">{{ props.mod.Name }}</span>
    <ModToggle :model-value="props.mod.Enabled" @update:model-value="onToggle" />
  </div>
</template>

<style scoped>
.disabled-mod {
  padding: var(--space-2) var(--space-3);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1rem;
  opacity: 0.7;
  cursor: pointer;
  transition: opacity 0.2s, background 0.2s;
}

.disabled-mod.selected {
  border-color: var(--color-border-strong);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-border-strong) 50%, transparent);
}

.disabled-mod:hover {
  opacity: 1;
  background: rgba(255,255,255,0.05);
}

.name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
