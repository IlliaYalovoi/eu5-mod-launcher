<script setup lang="ts">
import type { Mod } from '../types'

const props = defineProps<{
  mod: Mod
  index: number
}>()

const emit = defineEmits<{
  (event: 'contextmenu', payload: { modID: string; x: number; y: number }): void
}>()

function onContextMenu(event: MouseEvent): void {
  event.preventDefault()
  emit('contextmenu', {
    modID: props.mod.id,
    x: event.clientX,
    y: event.clientY,
  })
}
</script>

<template>
  <article class="load-order-item" @contextmenu.prevent="onContextMenu">
    <button class="drag-handle" type="button" aria-label="Drag to reorder" title="Drag to reorder">
      ☰
    </button>
    <span class="index">{{ index + 1 }}</span>
    <span class="name">{{ mod.name }}</span>
  </article>
</template>

<style scoped>
.load-order-item {
  display: grid;
  grid-template-columns: 2rem 2.5rem 1fr;
  align-items: center;
  gap: var(--space-2);
  min-height: 2.5rem;
  padding: var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
}

.drag-handle {
  border: 0;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: grab;
  font-size: 1rem;
  line-height: 1;
}

.drag-handle:active {
  cursor: grabbing;
}

.drag-handle:focus-visible {
  outline: var(--border-width) solid var(--color-border-strong);
  border-radius: var(--radius-sm);
}

.index {
  font-family: var(--font-mono), monospace;
  color: var(--color-text-muted);
  text-align: right;
}

.name {
  color: var(--color-text-primary);
  font-weight: 700;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>

