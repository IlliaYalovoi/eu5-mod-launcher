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
    class="launch-btn"
    :class="{
      'loading': isLaunching,
      'launch-btn--success': isSuccessFlash,
      'launch-btn--error': launchError,
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
.launch-btn {
  background: linear-gradient(to bottom, #b9935a, #8e6d3d);
  color: #1a1814;
  border: 1px solid #5a4623;
  padding: 14px;
  border-radius: 3px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 2px;
  cursor: pointer;
  box-shadow: 0 4px 0 #5a4623;
  width: 100%;
  text-align: center;
  transition: transform 0.1s, box-shadow 0.1s;
}

.launch-btn:active:not(:disabled) {
  transform: translateY(2px);
  box-shadow: 0 2px 0 #5a4623;
}

.launch-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  box-shadow: 0 4px 0 #3a2e18;
  background: linear-gradient(to bottom, #8a734e, #6e5532);
}

.loading {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { opacity: 0.8; }
  50% { opacity: 1; }
  100% { opacity: 0.8; }
}

.launch-btn--success {
  border-color: var(--color-success, #4caf50);
  color: var(--color-success, #4caf50);
}

.launch-btn--error {
  border-color: var(--color-danger, #f44336);
  color: var(--color-danger, #f44336);
}

.spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  display: inline-block;
  margin-right: 8px;
  vertical-align: middle;
}

.dismiss {
  margin-left: 8px;
  font-size: 1rem;
  line-height: 1;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
