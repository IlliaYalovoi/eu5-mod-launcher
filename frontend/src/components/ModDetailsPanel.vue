<script setup lang="ts">
import { computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useModsStore } from '../stores/mods'
import { OpenExternalLink, OpenWorkshopItem } from '../../wailsjs/go/main/App'
import { renderRichDescriptionHtml, renderSteamDescriptionHtml, toDisplayImageSrc } from '../utils/steamDescription'

const modsStore = useModsStore()
const { selectedMod, selectedSteamMetadata, selectedSteamLoading, selectedSteamError, workshopOpenError } = storeToRefs(modsStore)

const fallbackThumbnail =
  "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='67' viewBox='0 0 120 67'%3E%3Crect width='120' height='67' rx='8' fill='%23222a35'/%3E%3Cg fill='none' stroke='%23b9b09b' stroke-width='2'%3E%3Cpath d='M20 46l20-20 17 17 8-8 12 17'/%3E%3Crect x='20' y='14' width='80' height='40' rx='5'/%3E%3C/g%3E%3C/svg%3E"

watch(
  () => selectedMod.value?.ID,
  (id) => {
    if (id) {
      void modsStore.fetchSteamMetadata(id)
    }
  },
  { immediate: true },
)

const workshopURL = computed(() => {
  const itemID = selectedSteamMetadata.value?.itemId || ''
  return itemID ? `https://steamcommunity.com/sharedfiles/filedetails/?id=${itemID}` : ''
})

const steamThumbnail = computed(
  () =>
    toDisplayImageSrc(selectedMod.value?.ThumbnailPath || '') ||
    toDisplayImageSrc(selectedSteamMetadata.value?.previewUrl || '') ||
    fallbackThumbnail,
)
const localDescriptionHtml = computed(() => renderRichDescriptionHtml(selectedMod.value?.Description || ''))
const steamDescriptionHtml = computed(() => renderSteamDescriptionHtml(selectedSteamMetadata.value?.description || ''))

function retry(): void {
  if (selectedMod.value?.ID) {
    void modsStore.fetchSteamMetadata(selectedMod.value.ID)
  }
}

async function openWorkshop(): Promise<void> {
  const itemID = selectedSteamMetadata.value?.itemId || ''
  if (!itemID) {
    return
  }

  modsStore.clearWorkshopOpenError()
  try {
    await OpenWorkshopItem(itemID)
  } catch {
    modsStore.setWorkshopOpenError('Failed to open workshop item in Steam, browser, and fallback window.')
  }
}

async function onSteamContentClick(event: MouseEvent): Promise<void> {
  const target = event.target as HTMLElement | null
  const anchor = target?.closest('a') as HTMLAnchorElement | null
  if (!anchor) {
    return
  }

  const href = anchor.getAttribute('href')?.trim() || ''
  if (!href) {
    return
  }

  event.preventDefault()
  modsStore.clearWorkshopOpenError()
  try {
    await OpenExternalLink(href)
  } catch {
    modsStore.setWorkshopOpenError('Failed to open the selected link.')
  }
}
</script>

<template>
  <section class="mod-details-panel" aria-label="Mod details panel">
    <div v-if="!selectedMod" class="state empty">Select a mod to view details.</div>

    <template v-else>
      <header class="header">
        <h2 class="name">{{ selectedMod.Name }}</h2>
        <p class="subtitle">Version {{ selectedMod.Version || 'Unknown' }} · {{ selectedMod.Enabled ? 'Enabled' : 'Disabled' }}</p>
      </header>

      <img class="preview" :src="steamThumbnail" :alt="`${selectedMod.Name} preview`" loading="lazy" />

      <div class="section">
        <h3 class="section-title">Local details</h3>
        <div v-if="localDescriptionHtml" class="body steam-html" @click="onSteamContentClick" v-html="localDescriptionHtml" />
        <p v-else class="body">No local description provided.</p>
      </div>

      <div class="section">
        <h3 class="section-title">Steam details</h3>

        <p v-if="selectedSteamLoading" class="state loading">Loading workshop details...</p>
        <p v-else-if="selectedSteamError" class="state error">
          {{ selectedSteamError }}
          <button class="retry" type="button" @click="retry">Retry</button>
        </p>
        <div v-else-if="selectedSteamMetadata && selectedSteamMetadata.itemId" class="steam-content">
          <p class="steam-title">{{ selectedSteamMetadata.title || selectedMod.Name }}</p>
          <div v-if="steamDescriptionHtml" class="body steam-html" @click="onSteamContentClick" v-html="steamDescriptionHtml" />
          <p v-else class="body">No workshop description provided.</p>
          <button v-if="workshopURL" class="workshop-link" type="button" @click="openWorkshop">
            Open in Steam Workshop
          </button>
          <p v-if="workshopOpenError" class="state error">{{ workshopOpenError }}</p>
        </div>
        <p v-else class="state muted">No workshop metadata available for this mod.</p>
      </div>
    </template>
  </section>
</template>

<style scoped>
.mod-details-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  height: 100%;
  min-height: 0;
  overflow: auto;
}

.header {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.name {
  color: var(--color-text-primary);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.3;
}

.subtitle {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.preview {
  width: 100%;
  aspect-ratio: 16 / 9;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  object-fit: cover;
  background: var(--color-bg-elevated);
}

.section {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.section-title {
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.body {
  color: var(--color-text-secondary);
  font-size: 0.82rem;
  line-height: 1.45;
  white-space: pre-wrap;
}

.steam-title {
  color: var(--color-text-primary);
  font-weight: 700;
}

.workshop-link {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  min-height: 2rem;
  padding: 0 var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  cursor: pointer;
}

.workshop-link:hover {
  border-color: var(--color-border-strong);
}

.state {
  color: var(--color-text-secondary);
}

.state.error {
  color: var(--color-danger);
}

.state.muted {
  color: var(--color-text-muted);
}

.retry {
  margin-left: var(--space-2);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0 var(--space-2);
  min-height: 1.8rem;
  background: transparent;
  color: inherit;
  cursor: pointer;
}

:deep(.steam-html .steam-desc-image) {
  display: block;
  max-width: 100%;
  margin: var(--space-2) 0;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
}

:deep(.steam-html .steam-table) {
  width: 100%;
  border-collapse: collapse;
  margin: var(--space-2) 0;
}

:deep(.steam-html .steam-table th),
:deep(.steam-html .steam-table td) {
  border: var(--border-width) solid var(--color-border);
  padding: var(--space-1) var(--space-2);
  text-align: left;
}

:deep(.steam-html a) {
  color: var(--color-accent);
}

:deep(.steam-html) {
  white-space: normal;
}

:deep(.steam-html .steam-list) {
  margin: var(--space-2) 0;
  padding-left: 1.2rem;
}

:deep(.steam-html .steam-list li) {
  margin: var(--space-1) 0;
}

:deep(.steam-html blockquote) {
  margin: var(--space-2) 0;
  padding: var(--space-2);
  border-left: var(--border-width-strong) solid var(--color-border-strong);
  background: var(--color-bg-panel);
}

:deep(.steam-html pre) {
  margin: var(--space-2) 0;
  padding: var(--space-2);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  overflow-x: auto;
  white-space: pre-wrap;
}

:deep(.steam-html h1),
:deep(.steam-html h2),
:deep(.steam-html h3),
:deep(.steam-html h4),
:deep(.steam-html h5),
:deep(.steam-html h6) {
  margin: var(--space-2) 0 var(--space-1);
  color: var(--color-text-primary);
}

:deep(.steam-html hr) {
  border: 0;
  border-top: var(--border-width) solid var(--color-border);
  margin: var(--space-3) 0;
}
</style>

