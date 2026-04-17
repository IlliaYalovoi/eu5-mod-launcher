<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import type { Mod } from '../types'
import ModCard from './ModCard.vue'
import SearchInput from './ui/SearchInput.vue'
import { useModsStore } from '../stores/mods'

const modsStore = useModsStore()
const { allMods, isLoading, error, selectedModID } = storeToRefs(modsStore)
const searchText = ref('')

const filteredMods = computed(() => {
  const query = searchText.value.trim().toLowerCase()
  if (!query) {
    return allMods.value
  }
  return allMods.value.filter((mod) => mod.Name.toLowerCase().includes(query))
})

const emptyMessage = computed(() => {
  if (allMods.value.length === 0) {
    return 'No mods were discovered.'
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
}

.repo-title {
  font-size: 13px;
  text-transform: uppercase;
  color: var(--color-text-muted);
  letter-spacing: 1px;
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