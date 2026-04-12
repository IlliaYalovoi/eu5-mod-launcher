<script setup lang="ts">
import { computed, ref } from 'vue'
import { LaunchGame } from '../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'

const props = defineProps<{
  playsetNames: string[]
  gameActivePlaysetIndex: number
}>()

const isLaunching = ref(false)
const launchError = ref<string>('')

const gameLabel = computed(() => {
  return 'ENTER GAME'
})

const isSuccessFlash = ref(false)
let successTimer: ReturnType<typeof setTimeout> | null = null

async function onLaunch() {
  if (isLaunching.value) return
  isLaunching.value = true
  launchError.value = ''
  try {
    await LaunchGame()
    isSuccessFlash.value = true
    if (successTimer !== null) clearTimeout(successTimer)
    successTimer = setTimeout(() => { isSuccessFlash.value = false; successTimer = null }, 1500)
  } catch (err) {
    const msg = errorMessage(err)
    launchError.value = msg
    showToast({ type: 'error', message: msg })
  } finally {
    isLaunching.value = false
  }
}

function onDismissError(event: MouseEvent) {
  event.stopPropagation()
  launchError.value = ''
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
    <span v-else>{{ gameLabel }}</span>
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
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  width: 100%;
  padding: 14px;
  background: var(--accent);
  color: var(--bg-body);
  border: 1px solid var(--border-strong);
  border-radius: 3px;
  font-family: var(--font-display);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 2px;
  transition: all var(--transition-fast);
  cursor: pointer;
  box-shadow: 0 4px 0 var(--border-strong);
}

.launch-button:hover:not(:disabled) {
  background: var(--accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 5px 0 var(--border-strong);
}

.launch-button:active:not(:disabled) {
  transform: translateY(2px);
  box-shadow: 0 2px 0 var(--border-strong);
}

.launch-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.launch-button--success {
  border-color: #22c55e;
  color: #22c55e;
}

.launch-button--error {
  border-color: #ef4444;
  color: #ef4444;
}

.spinner {
  width: 1rem;
  height: 1rem;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
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

