<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import type { Mod } from '../types'
import ModCard from './ModCard.vue'
import SearchInput from './ui/SearchInput.vue'
import { useModsStore } from '../stores/mods'
import { useSettingsStore } from '../stores/settings'

const modsStore = useModsStore()
const settingsStore = useSettingsStore()
const { allMods, isLoading, error, selectedModID } = storeToRefs(modsStore)
const { gameVersion } = storeToRefs(settingsStore)
const searchText = ref('')

const filteredMods = computed(() => {
  const query = searchText.value.trim().toLowerCase()
  const disabled = allMods.value.filter(mod => !mod.Enabled)
  if (!query) {
    return disabled
  }
  return disabled.filter((mod) => mod.Name.toLowerCase().includes(query))
})

const emptyMessage = computed(() => {
  if (allMods.value.filter(m => !m.Enabled).length === 0) {
    return 'No disabled mods found.'
  }
  return 'No mods match your search query.'
})

function toggleMod(mod: Mod, value: boolean) {
  modsStore.setEnabled(mod.ID, value)
}

function selectMod(mod: Mod) {
  modsStore.selectMod(mod.ID)
}
</script>

<template>
  <aside class="repository">
    <div v-if="gameVersion === 'unknown'" class="warning-banner">
      Unknown game version - please set it manually in settings for correct mod compatibility check.
    </div>

    <div class="repo-title">Mod Repository (Disabled)</div>
    <SearchInput v-model="searchText" placeholder="Search unmanaged mods..." class="search-box" />

    <div class="list-body">
      <div v-if="isLoading" class="state loading">Loading mods...</div>
      <p v-else-if="error" class="state error">{{ error }}</p>
      <p v-else-if="filteredMods.length === 0" class="state empty">{{ emptyMessage }}</p>
      <div v-else class="cards">
        <ModCard
          v-for="mod in filteredMods"
          :key="mod.ID"
          :mod="mod"
          :selected="mod.ID === selectedModID"
          @toggle="(val: boolean) => toggleMod(mod, val)"
          @select="selectMod(mod)"
        />
      </div>
    </div>
    <div class="repo-footer">Items here are NOT in the current playset.</div>
  </aside>
</template>

<style scoped>
.repository {
  background: rgba(0,0,0,0.15);
  border: 1px dashed var(--color-border);
  padding: 20px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  height: 100%;
  min-height: 0;
}

.repo-title {
  font-size: 13px;
  text-transform: uppercase;
  color: var(--color-text-muted);
  letter-spacing: 1px;
}

.warning-banner {
  background: rgba(133, 77, 14, 0.3);
  color: #facc15;
  padding: 8px;
  text-align: center;
  font-size: 14px;
  border-bottom: 1px solid rgba(161, 98, 7, 0.5);
  border-radius: 4px;
}

.search-box {
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  padding: 10px;
  color: var(--color-text-primary);
  border-radius: 4px;
  width: 100%;
  box-sizing: border-box;
}

.list-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.cards {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.state {
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
  padding: 20px;
}

.error {
  color: var(--color-danger);
}

.repo-footer {
  text-align: center;
  color: var(--color-text-muted);
  font-size: 11px;
  margin-top: 10px;
}
</style>