<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import type { Mod } from '../types'
import ModCard from './ModCard.vue'
import SearchInput from './ui/SearchInput.vue'
import { useModsStore } from '../stores/mods'
import { useLoadOrderStore } from '../stores/loadorder'

const modsStore = useModsStore()
const loadOrderStore = useLoadOrderStore()
const { allMods, isLoading, error } = storeToRefs(modsStore)
const { playsetNames, gameActivePlaysetIndex, launcherActivePlaysetIndex } = storeToRefs(loadOrderStore)
const searchText = ref('')
const isSwitchingPlayset = ref(false)
const playsetError = ref<string | null>(null)

const filteredMods = computed(() => {
  const query = searchText.value.trim().toLowerCase()
  if (!query) {
    return allMods.value
  }
  return allMods.value.filter((mod) => mod.Name.toLowerCase().includes(query))
})

const emptyMessage = computed(() => {
  if (allMods.value.length === 0) {
    return 'No mods were discovered in local or workshop directories.'
  }
  return 'No mods match your search query.'
})

const hasPlaysetChoices = computed(() => playsetNames.value.length > 0)

const gameActiveLabel = computed(() => {
  const index = gameActivePlaysetIndex.value
  if (index < 0 || index >= playsetNames.value.length) {
    return 'Unknown'
  }
  return playsetNames.value[index]
})

onMounted(() => {
  if (allMods.value.length === 0 && !isLoading.value) {
    void modsStore.fetchAll()
  }
  if (playsetNames.value.length === 0) {
    void loadOrderStore.fetch()
  }
})

function toggleMod(mod: Mod): void {
  void modsStore.setEnabled(mod.ID, !mod.Enabled)
}

function onLauncherPlaysetChange(event: Event): void {
  const target = event.target as HTMLSelectElement
  const index = parseInt(target.value, 10)
  if (index.toString() !== target.value || index === launcherActivePlaysetIndex.value) {
    return
  }

  isSwitchingPlayset.value = true
  playsetError.value = null
  void loadOrderStore
    .setLauncherPlayset(index)
    .then(() => modsStore.fetchAll())
    .catch((err: unknown) => {
      playsetError.value = err instanceof Error ? err.message : String(err)
    })
    .finally(() => {
      isSwitchingPlayset.value = false
    })
}
</script>

<template>
  <section class="mod-list-panel" aria-label="All mods panel">
    <div class="playset-block">
      <label class="playset-label" for="launcher-playset">Launcher active playset</label>
      <select
        id="launcher-playset"
        class="playset-select"
        :disabled="!hasPlaysetChoices || isSwitchingPlayset"
        :value="launcherActivePlaysetIndex"
        @change="onLauncherPlaysetChange"
      >
        <option v-for="(name, index) in playsetNames" :key="`${name}-${index}`" :value="index">
          {{ name }}
        </option>
      </select>
      <p class="playset-hint">Game active: {{ gameActiveLabel }}</p>
      <p v-if="playsetError" class="playset-error">{{ playsetError }}</p>
    </div>

    <SearchInput v-model="searchText" placeholder="Search by mod name..." />

    <div class="list-body">
      <div v-if="isLoading" class="state loading">
        <span class="spinner" aria-hidden="true" />
        Loading mods...
      </div>

      <p v-else-if="error" class="state error">{{ error }}</p>

      <p v-else-if="filteredMods.length === 0" class="state empty">{{ emptyMessage }}</p>

      <div v-else class="cards">
        <ModCard v-for="mod in filteredMods" :key="mod.ID" :mod="mod" @toggle="toggleMod(mod)" />
      </div>
    </div>
  </section>
</template>

<style scoped>
.mod-list-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  height: 100%;
  min-height: 0;
}

.playset-block {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.playset-label {
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.playset-select {
  min-height: 2.25rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  padding: var(--space-2) var(--space-3);
}

.playset-select:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.playset-hint {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.playset-error {
  color: var(--color-danger);
  font-size: 0.75rem;
}

.list-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.cards {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding-right: var(--space-1);
}

.state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 7rem;
  padding: var(--space-4);
  border: var(--border-width) dashed var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  text-align: center;
}

.error {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.spinner {
  width: 1rem;
  height: 1rem;
  border: var(--border-width) solid var(--color-text-secondary);
  border-top-color: transparent;
  border-radius: 999px;
  animation: spin 700ms linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

