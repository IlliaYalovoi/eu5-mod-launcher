<script setup lang="ts">
import type { Mod } from '../types'

const props = defineProps<{
  mod: Mod
  index: number
  isDisabled?: boolean
}>()

const emit = defineEmits<{
  (event: 'contextmenu', payload: { modID: string; x: number; y: number }): void
  (event: 'toggle', modID: string): void
  (event: 'click', modID: string): void
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
  <article
    class="load-order-item"
    :class="{ 'is-disabled': isDisabled }"
    @click="emit('click', mod.id)"
    @contextmenu.stop.prevent="onContextMenu"
  >
    <div class="mod-handle">⠿</div>
    <div class="mod-index">{{ index + 1 }}</div>
    <div class="mod-info">
      <div class="mod-name-row">
        <span class="mod-name" :title="mod.name">{{ mod.name }}</span>
        <span
          v-if="mod.hasConflict"
          class="mod-severity-badge"
          title="Has conflicts"
        >
          !
        </span>
      </div>
      <span v-if="mod.tags && mod.tags.length > 0" class="mod-tags">{{ mod.tags.join(', ') }}</span>
    </div>
    <div
      class="toggle-switch"
      :class="{ 'is-on': mod.enabled }"
      @click.stop="emit('toggle', mod.id)"
    ></div>
  </article>
</template>

<style scoped>
.load-order-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  background: var(--card-bg, rgba(255, 255, 255, 0.02));
  border: var(--border-width, 1px) solid var(--card-border, var(--border));
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  min-height: 3.5rem;
}

.load-order-item:hover {
  background: var(--card-bg-hover, rgba(255, 255, 255, 0.05));
  border-color: var(--accent-primary, var(--accent));
}

.mod-index {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-muted);
  width: 1.5rem;
  text-align: right;
  opacity: 0.5;
}

.load-order-item.is-disabled {
  opacity: 0.6;
}

.mod-handle {
  cursor: grab !important;
  color: var(--text-muted);
  opacity: 0.3;
  font-size: 0.8rem;
  flex-shrink: 0;
}

.load-order-item:hover .mod-handle {
  opacity: 0.7;
}

.mod-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: 2px;
}

.mod-name-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.mod-name {
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text);
}

.mod-severity-badge {
  background: var(--warning, #f59e0b);
  color: black;
  font-size: 0.65rem;
  font-weight: 900;
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  flex-shrink: 0;
}

.mod-tags {
  font-size: 0.7rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toggle-switch {
  width: 34px;
  height: 18px;
  background: var(--bg-panel, #333);
  border-radius: 9px;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
  border: 1px solid var(--border);
  transition: all 0.2s;
}

.toggle-switch.is-on {
  background: var(--accent-primary, var(--success));
  border-color: var(--accent-primary, var(--success));
}

.toggle-switch::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  top: 1px;
  left: 1px;
  transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.toggle-switch.is-on::after {
  transform: translateX(16px);
}
</style>

