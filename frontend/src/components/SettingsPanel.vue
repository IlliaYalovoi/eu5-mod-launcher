<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import BaseButton from './ui/BaseButton.vue'
import { useSettingsStore } from '../stores/settings'
import { useModsStore } from '../stores/mods'

const props = defineProps<{ required?: boolean }>()
const emit = defineEmits<{ (event: 'close'): void }>()

const settingsStore = useSettingsStore()
const modsStore = useModsStore()
const { modsDirStatus, gameExe, autoDetectedGameExe, configPath } = storeToRefs(settingsStore)

const error = ref<string | null>(null)
const busy = ref(false)

async function withBusy(action: () => Promise<void>): Promise<void> {
  error.value = null
  busy.value = true
  try {
    await action()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    busy.value = false
  }
}

function onBrowseModsDir(): void {
  void withBusy(async () => {
    await settingsStore.browseModsDir()
    await modsStore.fetchAll()
  })
}

function onAutoDetectModsDir(): void {
  void withBusy(async () => {
    await settingsStore.autoDetectModsDir()
    await modsStore.fetchAll()
  })
}

function onBrowseGameExe(): void {
  void withBusy(async () => {
    await settingsStore.browseGameExecutable()
  })
}

function onAutoDetectGameExe(): void {
  void withBusy(async () => {
    await settingsStore.autoDetectGameExecutable()
  })
}

function onOpenConfigFolder(): void {
  void withBusy(async () => {
    await settingsStore.openConfigFolder()
  })
}
</script>

<template>
  <section class="settings-panel" aria-label="Settings panel">
    <header class="head">
      <h2 class="title">Settings</h2>
      <button v-if="!props.required" class="close" type="button" aria-label="Close" @click="emit('close')">×</button>
    </header>

    <p v-if="props.required" class="required-note">
      Auto-detection could not find a valid mods directory. Please pick one to continue.
    </p>

    <div class="field">
      <label class="label">Mods Directory</label>
      <input class="value" type="text" :value="modsDirStatus.effectiveDir || 'Not configured'" readonly />
      <p class="hint">
        Source: {{ modsDirStatus.usingCustomDir ? 'Custom override' : 'Auto-detected' }}
      </p>
      <div class="actions">
        <BaseButton variant="ghost" :disabled="busy" @click="onBrowseModsDir">Browse...</BaseButton>
        <BaseButton variant="ghost" :disabled="busy" @click="onAutoDetectModsDir">Auto detect</BaseButton>
      </div>
    </div>

    <div class="field">
      <label class="label">Game Executable (optional)</label>
      <input class="value" type="text" :value="gameExe || 'Not configured'" readonly />
      <p class="hint">
        Source: {{ gameExe && autoDetectedGameExe && gameExe === autoDetectedGameExe ? 'Auto-detected' : gameExe ? 'Custom override' : 'Not configured' }}
      </p>
      <div class="actions">
        <BaseButton variant="ghost" :disabled="busy" @click="onBrowseGameExe">Browse...</BaseButton>
        <BaseButton variant="ghost" :disabled="busy" @click="onAutoDetectGameExe">Auto detect</BaseButton>
      </div>
    </div>

    <div class="field">
      <label class="label">Config Path</label>
      <input class="value" type="text" :value="configPath" readonly />
      <div class="actions">
        <BaseButton variant="ghost" :disabled="busy" @click="onOpenConfigFolder">Open config folder</BaseButton>
      </div>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
  </section>
</template>

<style scoped>
.settings-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title {
  font-family: var(--font-display), serif;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.close {
  width: 2rem;
  height: 2rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.required-note {
  padding: var(--space-3);
  border: var(--border-width) solid var(--color-danger);
  border-radius: var(--radius-sm);
  color: var(--color-danger);
}

.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.label {
  color: var(--color-text-secondary);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.value {
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
}

.hint {
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.actions {
  display: flex;
  gap: var(--space-2);
}

.error {
  color: var(--color-danger);
  font-size: 0.85rem;
}
</style>

