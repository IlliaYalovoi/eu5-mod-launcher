<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { GetConfigPath, OpenConfigFolder } from '../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import BaseButton from './ui/BaseButton.vue'

const props = defineProps<{ required?: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const configPath = ref('')
const error = ref<string | null>(null)
const busy = ref(false)
const copiedField = ref<string | null>(null)

async function load() {
  try {
    configPath.value = await GetConfigPath()
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

async function withBusy(action: () => Promise<void>): Promise<void> {
  error.value = null
  busy.value = true
  try { await action() }
  catch (err) { error.value = errorMessage(err) }
  finally { busy.value = false }
}

function onOpenConfigFolder(): void {
  void withBusy(async () => { await OpenConfigFolder() })
}

async function copyToClipboard(text: string, field: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(text)
    copiedField.value = field
    setTimeout(() => { if (copiedField.value === field) copiedField.value = null }, 1500)
  } catch { /* clipboard not available */ }
}

onMounted(load)
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
  font-size: 1.25rem;
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.close {
  width: 2rem;
  height: 2rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
}

.close:hover {
  background: var(--bg-elevated);
  color: var(--text);
}

.required-note {
  padding: var(--space-3);
  border: 1px solid #ef4444;
  border-radius: var(--radius-sm);
  color: #ef4444;
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
  color: var(--text-muted);
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
  border-radius: 9999px;
  font-size: 0.68rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.badge--auto {
  border: 1px solid var(--success);
  color: var(--success);
}

.badge--custom {
  border: 1px solid var(--accent);
  color: var(--accent);
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
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-body);
  color: var(--text);
  font-family: var(--font-mono), monospace;
  font-size: 0.8rem;
}

.copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  flex-shrink: 0;
  transition: color var(--transition-fast), border-color var(--transition-fast);
}

.copy-btn:hover {
  color: var(--text);
  border-color: var(--accent);
}

.copy-btn--copied {
  color: var(--success);
  border-color: var(--success);
}

.hint {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.inline-link {
  border: 0;
  background: transparent;
  color: var(--accent);
  cursor: pointer;
  font-size: inherit;
  padding: 0;
  text-decoration: underline;
}

.inline-link:hover {
  color: var(--accent-hover);
}

.actions {
  display: flex;
  gap: var(--space-2);
}

.error {
  color: #ef4444;
  font-size: 0.85rem;
}
</style>
