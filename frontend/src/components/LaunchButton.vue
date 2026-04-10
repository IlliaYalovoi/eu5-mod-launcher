<script setup lang="ts">
import { ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useSettingsStore } from '../stores/settings'

const settingsStore = useSettingsStore()
const { isLaunching, launchError, lastLaunchAt } = storeToRefs(settingsStore)

const isSuccessFlash = ref(false)
let successTimer: number | null = null

watch(lastLaunchAt, (value) => {
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

function onLaunch(): void {
  void settingsStore.launchGame()
}

function onDismissError(event: MouseEvent): void {
  event.stopPropagation()
  settingsStore.clearLaunchError()
}
</script>

<template>
  <button
    class="launch-button"
    :class="{
      'launch-button--loading': isLaunching,
      'launch-button--success': isSuccessFlash,
      'launch-button--error': launchError,
    }"
    type="button"
    :disabled="isLaunching"
    @click="onLaunch"
  >
    <span v-if="isLaunching" class="spinner" aria-hidden="true" />
    <span v-if="isLaunching">Launching...</span>
    <span v-else-if="isSuccessFlash">✓ Launched</span>
    <span v-else>Launch Game</span>
    <span
      v-if="launchError"
      class="dismiss"
      role="button"
      tabindex="0"
      title="Dismiss launch error"
      @click="onDismissError"
    >
      ×
    </span>
  </button>
</template>

<style scoped>
.launch-button {
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

.launch-button:hover:not(:disabled) {
  background: var(--color-bg-elevated);
}

.launch-button:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.launch-button:disabled {
  opacity: 0.75;
  cursor: wait;
}

.launch-button--success {
  border-color: var(--color-success);
  color: var(--color-success);
}

.launch-button--error {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: var(--border-width) solid currentColor;
  border-top-color: transparent;
  border-radius: var(--radius-pill);
  animation: spin var(--duration-spinner) linear infinite;
}

.dismiss {
  margin-left: var(--space-1);
  font-size: 1rem;
  line-height: 1;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

