<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import BaseButton from './ui/BaseButton.vue'
import BaseModal from './ui/BaseModal.vue'
import { useSettingsStore } from '../stores/settings'

const props = defineProps<{
  open: boolean,
  gameID: string
}>()

const emit = defineEmits<{
  (event: 'close'): void
}>()

const settingsStore = useSettingsStore()
const { modsDirStatus, gameExe, autoDetectedGameExe } = storeToRefs(settingsStore)

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
  })
}

function onAutoDetectModsDir(): void {
  void withBusy(async () => {
    await settingsStore.autoDetectModsDir()
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
</script>

<template>
  <BaseModal :open="props.open" @close="emit('close')">
    <section class="game-settings" aria-label="Game specific settings">
      <header class="head">
        <h2 class="title">{{ gameID.toUpperCase() }} Settings</h2>
        <button class="close" type="button" @click="emit('close')">×</button>
      </header>

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
        <label class="label">Game Executable</label>
        <input class="value" type="text" :value="gameExe || 'Not configured'" readonly />
        <p class="hint">
          Source: {{ gameExe && autoDetectedGameExe && gameExe === autoDetectedGameExe ? 'Auto-detected' : gameExe ? 'Custom override' : 'Not configured' }}
        </p>
        <div class="actions">
          <BaseButton variant="ghost" :disabled="busy" @click="onBrowseGameExe">Browse...</BaseButton>
          <BaseButton variant="ghost" :disabled="busy" @click="onAutoDetectGameExe">Auto detect</BaseButton>
        </div>
      </div>

      <p v-if="error" class="error">{{ error }}</p>
    </section>
  </BaseModal>
</template>

<style scoped>
.game-settings {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  width: 32rem;
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
