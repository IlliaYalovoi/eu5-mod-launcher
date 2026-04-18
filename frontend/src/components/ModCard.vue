<script setup lang="ts">
import type { Mod } from '../types'
import ModToggle from './ui/ModToggle.vue'

const CompatibilityCardColoringEnabled = false

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
  <div class="disabled-mod" :class="{ selected: props.selected, 'incompatible-card': !props.mod.IsCompatible && CompatibilityCardColoringEnabled }" @click="onSelect">
    <div class="info">
      <span class="name">{{ props.mod.Name }}</span>
      <div class="version-info">
        <span class="version-text">{{ props.mod.SupportedVersion }}</span>
        <span v-if="!props.mod.IsCompatible" class="warning-icon" title="Incompatible game version">⚠️</span>
      </div>
    </div>
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
  transition: opacity 0.2s, background 0.2s, border-color 0.2s;
}

.disabled-mod.incompatible-card {
  background: color-mix(in oklab, var(--color-danger) 20%, var(--color-bg-elevated));
  border-color: color-mix(in oklab, var(--color-danger) 50%, var(--color-border));
}

.disabled-mod.selected {
  border-color: var(--color-border-strong);
  box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--color-border-strong) 50%, transparent);
}

.disabled-mod:hover {
  opacity: 1;
  background: rgba(255,255,255,0.05);
}

.info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  margin-right: var(--space-3);
}

.name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.version-info {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: 2px;
}

.version-text {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.warning-icon {
  font-size: 0.85rem;
  color: #eab308;
}
</style>
