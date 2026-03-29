<script setup lang="ts">
import { ref, watch } from 'vue'
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

function onClick(): void {
  void loadOrderStore.autosort()
}
</script>

<template>
  <button
    class="autosort-button"
    :class="{
      'autosort-button--sorting': isSorting,
      'autosort-button--success': isSuccessFlash,
      'autosort-button--error': autosortError,
    }"
    type="button"
    :disabled="isSorting"
    @click="onClick"
  >
    <span v-if="isSorting" class="spinner" aria-hidden="true" />
    <span v-if="isSorting">Sorting...</span>
    <span v-else-if="isSuccessFlash">✓ Sorted</span>
    <span v-else>Auto-sort</span>
  </button>
</template>

<style scoped>
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
}

.spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: var(--border-width) solid currentColor;
  border-top-color: transparent;
  border-radius: 999px;
  animation: spin 700ms linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

