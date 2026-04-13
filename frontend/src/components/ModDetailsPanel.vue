<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { OpenExternalLink, OpenWorkshopItem, FetchWorkshopMetadataForMod, IsUnsubscribeEnabled, UnsubscribeWorkshopMod } from '../wailsjs/go/launcher/App'
import { renderRichDescriptionHtml, renderSteamDescriptionHtml, toDisplayImageSrc } from '../utils/steamDescription'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'
import type { Mod, WorkshopItem } from '../types'
import ConfirmModal from './ui/ConfirmModal.vue'

const props = defineProps<{ mod: Mod | null }>()

const emit = defineEmits<{
  (event: 'close'): void
}>()

const mod = ref<Mod | null>(null)
const steamMetadata = ref<WorkshopItem | null>(null)
const steamLoading = ref(false)
const steamError = ref<string>('')
const unsubscribeLoading = ref(false)
const unsubscribeError = ref<string>('')
const workshopOpenError = ref<string>('')
const unsubscribeConfirmOpen = ref(false)
const unsubscribeEnabled = ref(false)
const unsubscribeFeatureLoaded = ref(false)

const fallbackThumbnail = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='67' viewBox='0 0 120 67'%3E%3Crect width='120' height='67' rx='8' fill='%23222a35'/%3E%3Cg fill='none' stroke='%23b9b09b' stroke-width='2'%3E%3Cpath d='M20 46l20-20 17 17 8-8 12 17'/%3E%3Crect x='20' y='14' width='80' height='40' rx='5'/%3E%3C/g%3E%3C/svg%3E"

async function loadModDetails() {
  if (!props.mod?.id) return
  steamLoading.value = true
  steamError.value = ''
  try {
    if (!unsubscribeFeatureLoaded.value) {
      unsubscribeEnabled.value = await IsUnsubscribeEnabled()
      unsubscribeFeatureLoaded.value = true
    }
    const ws = await FetchWorkshopMetadataForMod(props.mod.id)
    steamMetadata.value = ws as WorkshopItem
  } catch (err) {
    steamError.value = errorMessage(err)
  } finally {
    steamLoading.value = false
  }
}

watch(() => props.mod, (m) => {
  mod.value = m
  if (m) void loadModDetails()
}, { immediate: true })

const workshopURL = computed(() => {
  const itemID = steamMetadata.value?.itemId || ''
  return itemID ? `https://steamcommunity.com/sharedfiles/filedetails/?id=${itemID}` : ''
})
const canUnsubscribe = computed(() => !!props.mod?.id && unsubscribeEnabled.value)

const steamThumbnail = computed(() =>
  toDisplayImageSrc(mod.value?.thumbnailPath || '') ||
  toDisplayImageSrc(steamMetadata.value?.previewUrl || '') ||
  fallbackThumbnail)
const localDescriptionHtml = computed(() => renderRichDescriptionHtml(mod.value?.description || ''))
const steamDescriptionHtml = computed(() => renderSteamDescriptionHtml(steamMetadata.value?.description || ''))

function retry(): void { void loadModDetails() }

async function openWorkshop(): Promise<void> {
  const itemID = steamMetadata.value?.itemId || ''
  if (!itemID) return
  workshopOpenError.value = ''
  try { await OpenWorkshopItem(itemID) }
  catch { workshopOpenError.value = 'Failed to open workshop item.' }
}

async function onSteamContentClick(event: MouseEvent): Promise<void> {
  const target = event.target as HTMLElement | null
  const anchor = target?.closest('a') as HTMLAnchorElement | null
  if (!anchor) return
  const href = anchor.getAttribute('href')?.trim() || ''
  if (!href) return
  event.preventDefault()
  workshopOpenError.value = ''
  try { await OpenExternalLink(href) }
  catch { workshopOpenError.value = 'Failed to open the selected link.' }
}

function openUnsubscribeConfirm(): void { unsubscribeConfirmOpen.value = true }
function closeUnsubscribeConfirm(): void { unsubscribeConfirmOpen.value = false }
async function confirmUnsubscribe(): Promise<void> {
  unsubscribeConfirmOpen.value = false
  if (!props.mod?.id) return
  unsubscribeLoading.value = true
  try {
    await UnsubscribeWorkshopMod(props.mod.id)
    showToast({ type: 'success', message: 'Unsubscribed successfully' })
  } catch (err) {
    unsubscribeError.value = errorMessage(err)
  } finally {
    unsubscribeLoading.value = false
  }
}
</script>

<template>
  <section class="mod-details-panel" aria-label="Mod details panel">
    <div v-if="!mod" class="state empty">Select a mod to view details.</div>

    <template v-else>
      <header class="header">
        <h2 class="name">{{ mod.name }}</h2>
        <p class="subtitle">Version {{ mod.version || 'Unknown' }} · {{ mod.enabled ? 'Enabled' : 'Disabled' }}</p>
      </header>

      <img class="preview" :src="steamThumbnail" :alt="`${mod.name} preview`" loading="lazy" />

      <div class="section">
        <h3 class="section-title">Local details</h3>
        <div v-if="localDescriptionHtml" class="body steam-html" @click="onSteamContentClick" v-html="localDescriptionHtml" />
        <p v-else class="body">No local description provided.</p>
      </div>

      <div class="section">
        <h3 class="section-title">Steam details</h3>

        <p v-if="steamLoading" class="state loading">Loading workshop details...</p>
        <p v-else-if="steamError" class="state error">
          {{ steamError }}
          <button class="retry" type="button" @click="retry">Retry</button>
        </p>
        <div v-else-if="steamMetadata && steamMetadata.itemId" class="steam-content">
          <p class="steam-title">{{ steamMetadata.title || mod.name }}</p>
          <div v-if="steamDescriptionHtml" class="body steam-html" @click="onSteamContentClick" v-html="steamDescriptionHtml" />
          <p v-else class="body">No workshop description provided.</p>
          <button v-if="workshopURL" class="workshop-link" type="button" @click="openWorkshop">
            Open in Steam Workshop
          </button>
          <p v-if="workshopOpenError" class="state error">{{ workshopOpenError }}</p>
        </div>
        <p v-else class="state muted">No workshop metadata available for this mod.</p>
      </div>

      <div v-if="canUnsubscribe" class="unsubscribe-area">
        <button
          class="unsubscribe-btn"
          type="button"
          :disabled="unsubscribeLoading"
          @click="openUnsubscribeConfirm"
        >
          {{ unsubscribeLoading ? 'Unsubscribing...' : 'Unsubscribe from Workshop' }}
        </button>
        <p v-if="unsubscribeError" class="state error">{{ unsubscribeError }}</p>
      </div>
    </template>
  </section>

  <ConfirmModal
    :open="unsubscribeConfirmOpen"
    title="Unsubscribe from Workshop?"
    message="This will remove the mod from your Steam subscription. Steam may take a moment to sync."
    confirm-label="Unsubscribe"
    :danger="true"
    @close="closeUnsubscribeConfirm"
    @confirm="confirmUnsubscribe"
  />
</template>

<style scoped>
.mod-details-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  height: 100%;
  min-height: 0;
}

.header {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.name {
  color: var(--text);
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 700;
  line-height: 1.3;
}

.subtitle {
  color: var(--text-muted);
  font-size: 0.85rem;
}

.preview {
  width: 100%;
  aspect-ratio: 16 / 9;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  object-fit: cover;
  background: var(--bg-elevated);
}

.section {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.section-title {
  color: var(--accent);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  border-bottom: 1px solid var(--border);
  padding-bottom: var(--space-2);
}

.body {
  color: var(--text);
  font-size: 0.9rem;
  line-height: 1.6;
  white-space: pre-wrap;
}

.steam-title {
  color: var(--text);
  font-weight: 700;
  margin-bottom: var(--space-2);
}

.workshop-link {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  min-height: 2.25rem;
  padding: 0 var(--space-4);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-size: 0.85rem;
  transition: all var(--transition-fast);
}

.workshop-link:hover {
  border-color: var(--accent);
  background: var(--bg-elevated);
}

.unsubscribe-area {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding-top: var(--space-4);
  margin-top: auto;
  border-top: 1px solid var(--border);
}

.unsubscribe-btn {
  display: inline-flex;
  align-items: center;
  width: 100%;
  justify-content: center;
  min-height: 2.5rem;
  padding: 0 var(--space-4);
  border: 1px solid #ef4444;
  border-radius: var(--radius-sm);
  background: transparent;
  color: #ef4444;
  cursor: pointer;
  font-weight: 600;
  transition: all var(--transition-fast);
}

.unsubscribe-btn:hover:not(:disabled) {
  background: #ef4444;
  color: #fff;
}

.state.error {
  color: #ef4444;
}

.retry {
  margin-left: var(--space-2);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 0 var(--space-3);
  min-height: 1.8rem;
  background: var(--bg-body);
  color: var(--text);
  cursor: pointer;
}

:deep(.steam-html .steam-desc-image) {
  display: block;
  max-width: 100%;
  margin: var(--space-3) 0;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}

:deep(.steam-html a) {
  color: var(--accent);
  text-decoration: underline;
}

:deep(.steam-html blockquote) {
  margin: var(--space-3) 0;
  padding: var(--space-3);
  border-left: 4px solid var(--accent);
  background: var(--bg-panel);
}
</style>
