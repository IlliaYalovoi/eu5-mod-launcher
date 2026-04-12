<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { subscribeToasts, type Toast } from '../../lib/toast'

const toasts = ref<Toast[]>([])
const timers = new Map<string, ReturnType<typeof setTimeout>>()

function addToast(toast: Toast) {
  toasts.value.push(toast)
  if (toasts.value.length > 3) {
    const oldest = toasts.value.shift()
    if (oldest) {
      const t = timers.get(oldest.id)
      if (t) { clearTimeout(t); timers.delete(oldest.id) }
    }
  }
  if (toast.type !== 'error') {
    timers.set(toast.id, setTimeout(() => removeToast(toast.id), 3200))
  }
}

function removeToast(id: string) {
  const t = timers.get(id)
  if (t) { clearTimeout(t); timers.delete(id) }
  const idx = toasts.value.findIndex((toast) => toast.id === id)
  if (idx >= 0) toasts.value.splice(idx, 1)
}

const overflowCount = computed(() => Math.max(0, toasts.value.length - 3))
const visibleToasts = computed(() => toasts.value.slice(0, 3))

let unsubscribe: (() => void) | null = null

onMounted(() => { unsubscribe = subscribeToasts(addToast) })
onUnmounted(() => { if (unsubscribe) unsubscribe(); timers.forEach(t => clearTimeout(t)) })
</script>

<template>
  <div class="toast-container" aria-live="polite" aria-atomic="false">
    <TransitionGroup name="toast" tag="div" class="toast-list">
      <div
        v-for="toast in visibleToasts"
        :key="toast.id"
        class="toast"
        :class="`toast--${toast.type}`"
        role="status"
      >
        <span class="toast-icon" aria-hidden="true">
          <template v-if="toast.type === 'success'">✓</template>
          <template v-else-if="toast.type === 'error'">✕</template>
          <template v-else-if="toast.type === 'warning'">⚠</template>
          <template v-else>ℹ</template>
        </span>
        <span class="toast-message">{{ toast.message }}</span>
        <button class="toast-close" type="button" aria-label="Dismiss" @click="removeToast(toast.id)">×</button>
      </div>
    </TransitionGroup>
    <p v-if="overflowCount > 0" class="overflow-badge">+{{ overflowCount }} more</p>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  bottom: var(--space-5);
  right: var(--space-5);
  z-index: 1400;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: var(--space-2);
  pointer-events: none;
}

.toast-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.toast {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 18rem;
  max-width: 26rem;
  padding: var(--space-3) var(--space-4);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  pointer-events: all;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
}

.toast--success {
  border-color: var(--color-success);
}

.toast--success .toast-icon {
  color: var(--color-success);
}

.toast--error {
  border-color: var(--color-danger);
}

.toast--error .toast-icon {
  color: var(--color-danger);
}

.toast--warning {
  border-color: var(--color-warning);
}

.toast--warning .toast-icon {
  color: var(--color-warning);
}

.toast--info {
  border-color: var(--color-info);
}

.toast--info .toast-icon {
  color: var(--color-info);
}

.toast-icon {
  flex-shrink: 0;
  font-size: 0.85rem;
  font-weight: 700;
}

.toast-message {
  flex: 1;
  color: var(--color-text-primary);
  font-size: 0.85rem;
  line-height: 1.4;
}

.toast-close {
  flex-shrink: 0;
  border: 0;
  background: transparent;
  color: var(--color-text-muted);
  font-size: 1rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  transition: color var(--transition-fast);
}

.toast-close:hover {
  color: var(--color-text-primary);
}

.overflow-badge {
  color: var(--color-text-muted);
  font-size: 0.75rem;
  pointer-events: none;
}

.toast-enter-active {
  transition: all 200ms ease-out;
}

.toast-leave-active {
  transition: all 160ms ease-in;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(1rem);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(1rem);
}

.toast-move {
  transition: transform 200ms ease;
}
</style>
