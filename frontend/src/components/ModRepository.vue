<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import type { Mod } from '../types'
import { GetAllMods, EnableMod } from '../wailsjs/go/launcher/App'
import { showToast } from '../lib/toast'
import { errorMessage } from '../lib/error'

const emit = defineEmits<{
  (event: 'select-mod', modID: string): void
  (event: 'mod-enabled'): void
}>()

const allMods = ref<Mod[]>([])
const searchQuery = ref('')
const loading = ref(false)

const filteredMods = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  return allMods.value
    .filter(m => !m.enabled) // Repository shows disabled mods
    .filter(m => !query || m.name.toLowerCase().includes(query) || m.id.toLowerCase().includes(query))
    .sort((a, b) => a.name.localeCompare(b.name))
})

async function load() {
  loading.value = true
  try {
    const mods = await GetAllMods()
    allMods.value = mods as any as Mod[]
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  } finally {
    loading.value = false
  }
}

async function handleEnable(modID: string) {
  try {
    await EnableMod(modID)
    showToast({ type: 'success', message: 'Mod added to load order' })
    await load()
    emit('mod-enabled')
  } catch (err) {
    showToast({ type: 'error', message: errorMessage(err) })
  }
}

onMounted(load)

defineExpose({ load })
</script>

<template>
  <div class="mod-repository">
    <header class="repo-header">
      <h2 class="repo-title">Mod Repository</h2>
      <div class="search-box">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search repository..."
          class="repo-search"
        />
      </div>
    </header>

    <div class="repo-list">
      <div
        v-for="mod in filteredMods"
        :key="mod.id"
        class="repo-item"
        @click="emit('select-mod', mod.id)"
      >
        <div class="repo-item-main">
          <span class="repo-item-name">{{ mod.name }}</span>
          <span class="repo-item-id">{{ mod.tags?.join(', ') || mod.id }}</span>
        </div>
        <button class="enable-btn" @click.stop="handleEnable(mod.id)">
          <div class="toggle"></div>
        </button>
      </div>
      <div v-if="filteredMods.length === 0 && !loading" class="repo-empty">
        No mods found in repository.
      </div>
      <div v-if="loading" class="repo-loading">Loading repository...</div>
    </div>
  </div>
</template>

<style scoped>
.mod-repository {
  display: flex;
  flex-direction: column;
  height: calc(100% - var(--space-8));
  margin: var(--space-4);
  border: 1px dashed var(--border);
  border-radius: var(--radius-lg);
  background: rgba(0, 0, 0, 0.15);
  min-height: 0;
}

.repo-header {
  padding: var(--space-4);
  background: transparent;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.repo-title {
  font-family: var(--font-display);
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-muted);
}

.repo-search {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
}

.repo-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-2);
}

.repo-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-1);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.repo-item:hover {
  background: var(--bg-elevated);
}

.repo-item-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.repo-item-name {
  font-weight: 600;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.repo-item-id {
  font-size: 0.7rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.enable-btn {
  background: transparent;
  padding: 0;
}

.toggle {
  width: 36px;
  height: 18px;
  background: #444;
  border-radius: 10px;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
}

.toggle::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: 0.2s;
}

.enable-btn:hover .toggle {
  background: #555;
}

.repo-empty, .repo-loading {
  padding: var(--space-5);
  text-align: center;
  color: var(--text-muted);
  font-size: 0.85rem;
}
</style>
