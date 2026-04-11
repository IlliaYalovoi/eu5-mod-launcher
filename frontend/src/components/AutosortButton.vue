<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useLoadOrderStore } from '../stores/loadorder'

const loadOrderStore = useLoadOrderStore()
const { isSorting, autosortError, lastSortedAt } = storeToRefs(loadOrderStore)

const isSuccessFlash = ref(false)
let successTimer: number | null = null

watch(lastSortedAt, (value) => {
  if (!value) {
    return
  }
  isSuccessFlash.value = true
  if (successTimer !== null) {
    window.clearTimeout(successTimer)
  }
  successTimer = window.setTimeout(() => {
    isSuccessFlash.value = false
    successTimer = null
  }, 1500)
})

const cycleNodes = computed(() => {
  const err = autosortError.value || ''
  const marker = err.toLowerCase().lastIndexOf('cycle detected:')
  if (marker < 0) return []
  return err
    .slice(marker + 'cycle detected:'.length)
    .split('->')
    .map((s) => s.trim())
    .filter(Boolean)
})

const cycleDiagram = computed(() => {
  if (cycleNodes.value.length === 0) return ''
  return cycleNodes.value.join(' → ')
})

function onClick(): void {
  if (autosortError.value) {
    return
  }
  void loadOrderStore.autosort()
}
</script>

<template>
  <div class="autosort-wrapper">
    <button
      class="autosort-button"
      :class="{
        'autosort-button--sorting': isSorting,
        'autosort-button--success': isSuccessFlash,
        'autosort-button--error': !!autosortError,
      }"
      type="button"
      :disabled="isSorting || !!autosortError"
      @click="onClick"
    >
      <span v-if="isSorting" class="spinner" aria-hidden="true" />
      <span v-if="isSorting">Sorting...</span>
      <span v-else-if="isSuccessFlash">✓ Sorted</span>
      <span v-else-if="autosortError">⚠ Cycle</span>
      <span v-else>Auto-sort</span>
    </button>

    <div v-if="autosortError" class="cycle-breadcrumb">
      <span class="cycle-text">{{ cycleDiagram }}</span>
      <span class="cycle-arrow" aria-hidden="true">↻</span>
    </div>
  </div>
</template>

<style scoped>
.autosort-wrapper {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.autosort-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-4);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-primary);
  font-weight: 700;
  cursor: pointer;
  transition: border-color var(--transition-fast), background var(--transition-fast), color var(--transition-fast);
}

.autosort-button:hover:not(:disabled) {
  background: var(--color-bg-elevated);
}

.autosort-button:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.autosort-button:disabled {
  opacity: 0.75;
  cursor: wait;
}

.autosort-button--success {
  border-color: var(--color-success);
  color: var(--color-success);
}

.autosort-button--error {
  border-color: var(--color-danger);
  color: var(--color-danger);
  cursor: not-allowed;
}

.spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: var(--border-width) solid currentColor;
  border-top-color: transparent;
  border-radius: var(--radius-pill);
  animation: spin var(--duration-spinner) linear infinite;
}

.cycle-breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  padding: 0.2rem 0.5rem;
  border: var(--border-width) solid var(--color-danger);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-danger);
  font-family: var(--font-mono), monospace;
  font-size: 0.7rem;
  max-width: 16rem;
  overflow: hidden;
}

.cycle-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cycle-arrow {
  flex-shrink: 0;
  font-size: 0.8rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

