<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import BaseButton from './ui/BaseButton.vue'
import { useSettingsStore } from '../stores/settings'
import { useModsStore } from '../stores/mods'
import { useLoadOrderStore } from '../stores/loadorder'
import { useGamesStore } from '../stores/games'
import { PickFolder, SetGamePaths } from '../../wailsjs/go/launcher/App'

const props = defineProps<{
  gameID: string
  gameName: string
  open: boolean
}>()
const emit = defineEmits<{ (event: 'close'): void }>()

const settingsStore = useSettingsStore()
const modsStore = useModsStore()
const loadOrderStore = useLoadOrderStore()
const gamesStore = useGamesStore()

const installDir = ref<string>('')
const documentsDir = ref<string>('')
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

async function onBrowseInstallDir(): Promise<void> {
  await withBusy(async () => {
    const picked = await PickFolder()
    if (picked) {
      installDir.value = picked
    }
  })
}

async function onBrowseDocumentsDir(): Promise<void> {
  await withBusy(async () => {
    const picked = await PickFolder()
    if (picked) {
      documentsDir.value = picked
    }
  })
}

async function onSave(): Promise<void> {
  await withBusy(async () => {
    if (!installDir.value.trim() || !documentsDir.value.trim()) {
      error.value = 'Both install and documents directories are required.'
      return
    }

    await SetGamePaths(props.gameID, installDir.value.trim(), documentsDir.value.trim())

    // Set this game as active since it's now configured
    gamesStore.setActiveGame(props.gameID)

    // Refresh game detection and mod list
    await Promise.all([
      settingsStore.fetch(),
      modsStore.fetchAll(),
      loadOrderStore.fetch()
    ])

    emit('close')
  })
}
</script>

<template>
  <div class="setup-backdrop" @click.self="emit('close')">
    <div class="setup-modal">
      <header class="modal-head">
        <h2 class="title">Setup {{ gameName }}</h2>
        <button class="close-button" type="button" aria-label="Close" @click="emit('close')">×</button>
      </header>

      <div class="field">
        <div class="field-label-row">
          <label class="label">Install Directory</label>
        </div>
        <div class="value-row">
          <input class="value" type="text" :value="installDir" placeholder="Select game install directory" readonly />
        </div>
        <div class="actions">
          <BaseButton variant="ghost" :disabled="busy" @click="onBrowseInstallDir">Browse...</BaseButton>
        </div>
      </div>

      <div class="field">
        <div class="field-label-row">
          <label class="label">Documents Directory</label>
        </div>
        <div class="value-row">
          <input class="value" type="text" :value="documentsDir" placeholder="Select game documents directory" readonly />
        </div>
        <div class="actions">
          <BaseButton variant="ghost" :disabled="busy" @click="onBrowseDocumentsDir">Browse...</BaseButton>
        </div>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <div class="modal-actions">
        <BaseButton variant="ghost" :disabled="busy" @click="emit('close')">Cancel</BaseButton>
        <BaseButton variant="primary" :disabled="busy || !installDir.trim() || !documentsDir.trim()" @click="onSave">
          Setup Game
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.setup-backdrop {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-overlay);
  z-index: 300;
}

.setup-modal {
  width: min(32rem, 90vw);
  padding: var(--space-5);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.title {
  font-family: var(--font-display), serif;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.close-button {
  width: 2rem;
  height: 2rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.close-button:hover {
  background: var(--color-bg-panel);
}

.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.field-label-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.label {
  color: var(--color-text-secondary);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 700;
}

.value-row {
  display: flex;
  gap: var(--space-2);
  align-items: center;
}

.value {
  flex: 1;
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
  font-family: var(--font-mono), monospace;
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

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  margin-top: var(--space-4);
}
</style>