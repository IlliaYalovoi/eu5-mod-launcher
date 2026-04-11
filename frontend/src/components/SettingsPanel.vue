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
const copiedField = ref<string | null>(null)

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

async function copyToClipboard(text: string, field: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(text)
    copiedField.value = field
    setTimeout(() => {
      if (copiedField.value === field) {
        copiedField.value = null
      }
    }, 1500)
  } catch {
    // clipboard not available
  }
}

function onResetModsDir(): void {
  void withBusy(async () => {
    await settingsStore.autoDetectModsDir()
    await modsStore.fetchAll()
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
      <div class="field-label-row">
        <label class="label">Mods Directory</label>
        <span v-if="!modsDirStatus.usingCustomDir && modsDirStatus.autoDetectedExists" class="badge badge--auto">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" />
          </svg>
          Auto-detected
        </span>
        <span v-else-if="modsDirStatus.usingCustomDir" class="badge badge--custom">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" aria-hidden="true">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
          </svg>
          Custom
        </span>
      </div>
      <div class="value-row">
        <input class="value" type="text" :value="modsDirStatus.effectiveDir || 'Not configured'" readonly />
        <button
          v-if="modsDirStatus.effectiveDir"
          class="copy-btn"
          :class="{ 'copy-btn--copied': copiedField === 'modsDir' }"
          type="button"
          :aria-label="copiedField === 'modsDir' ? 'Copied' : 'Copy path'"
          @click="copyToClipboard(modsDirStatus.effectiveDir || '', 'modsDir')"
        >
          <svg v-if="copiedField !== 'modsDir'" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
          </svg>
          <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </button>
      </div>
      <p v-if="modsDirStatus.usingCustomDir" class="hint">
        Using custom path override.
        <button class="inline-link" type="button" @click="onResetModsDir">Reset to auto-detect</button>
      </p>
      <div class="actions">
        <BaseButton variant="ghost" :disabled="busy" @click="onBrowseModsDir">Browse...</BaseButton>
        <BaseButton variant="ghost" :disabled="busy" @click="onAutoDetectModsDir">Auto detect</BaseButton>
      </div>
    </div>

    <div class="field">
      <div class="field-label-row">
        <label class="label">Game Executable</label>
        <span v-if="gameExe && autoDetectedGameExe && gameExe === autoDetectedGameExe" class="badge badge--auto">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" />
          </svg>
          Auto-detected
        </span>
        <span v-else-if="gameExe" class="badge badge--custom">
          <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" aria-hidden="true">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
          </svg>
          Custom
        </span>
      </div>
      <div class="value-row">
        <input class="value" type="text" :value="gameExe || 'Not configured'" readonly />
        <button
          v-if="gameExe"
          class="copy-btn"
          :class="{ 'copy-btn--copied': copiedField === 'gameExe' }"
          type="button"
          :aria-label="copiedField === 'gameExe' ? 'Copied' : 'Copy path'"
          @click="copyToClipboard(gameExe || '', 'gameExe')"
        >
          <svg v-if="copiedField !== 'gameExe'" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
          </svg>
          <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </button>
      </div>
      <div class="actions">
        <BaseButton variant="ghost" :disabled="busy" @click="onBrowseGameExe">Browse...</BaseButton>
        <BaseButton variant="ghost" :disabled="busy" @click="onAutoDetectGameExe">Auto detect</BaseButton>
      </div>
    </div>

    <div class="field">
      <div class="field-label-row">
        <label class="label">Config Path</label>
      </div>
      <div class="value-row">
        <input class="value" type="text" :value="configPath" readonly />
        <button
          v-if="configPath"
          class="copy-btn"
          :class="{ 'copy-btn--copied': copiedField === 'configPath' }"
          type="button"
          :aria-label="copiedField === 'configPath' ? 'Copied' : 'Copy path'"
          @click="copyToClipboard(configPath, 'configPath')"
        >
          <svg v-if="copiedField !== 'configPath'" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
          </svg>
          <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </button>
      </div>
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
  gap: var(--space-5);
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

.close:hover {
  background: var(--color-bg-panel);
}

.required-note {
  padding: var(--space-3);
  border: var(--border-width) solid var(--color-danger);
  border-radius: var(--radius-sm);
  color: var(--color-danger);
  font-size: 0.85rem;
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

.badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: 0.1rem 0.5rem;
  border-radius: var(--radius-pill);
  font-size: 0.68rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.badge--auto {
  border: var(--border-width) solid var(--color-success);
  color: var(--color-success);
}

.badge--custom {
  border: var(--border-width) solid var(--color-accent);
  color: var(--color-accent);
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

.copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: color var(--transition-fast), border-color var(--transition-fast);
}

.copy-btn:hover {
  color: var(--color-text-primary);
  border-color: var(--color-accent);
}

.copy-btn--copied {
  color: var(--color-success);
  border-color: var(--color-success);
}

.hint {
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.inline-link {
  border: 0;
  background: transparent;
  color: var(--color-accent);
  cursor: pointer;
  font-size: inherit;
  padding: 0;
  text-decoration: underline;
}

.inline-link:hover {
  color: var(--color-accent-hover);
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
