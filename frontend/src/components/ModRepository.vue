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
const sortBy = ref<'name-asc' | 'name-desc' | 'id'>('name-asc')
const loading = ref(false)

const filteredMods = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  const mods = allMods.value
    .filter(m => !m.enabled) // Repository shows disabled mods
    .filter(m => !query || m.name.toLowerCase().includes(query) || m.id.toLowerCase().includes(query))

  return [...mods].sort((a, b) => {
    switch (sortBy.value) {
      case 'name-asc':
        return a.name.localeCompare(b.name)
      case 'name-desc':
        return b.name.localeCompare(a.name)
      case 'id':
        return a.id.localeCompare(b.id)
      default:
        return 0
    }
  })
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
      <div class="repo-header-top">
        <h2 class="repo-title">Discover</h2>
        <div class="repo-count">{{ filteredMods.length }} mods available</div>
      </div>
      <div class="filter-bar">
        <div class="search-box">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search mods..."
            class="repo-search"
          />
        </div>
        <div class="sort-box">
          <select v-model="sortBy" class="repo-sort">
            <option value="name-asc">Name A-Z</option>
            <option value="name-desc">Name Z-A</option>
            <option value="id">ID</option>
          </select>
        </div>
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
  height: 100%;
  border-radius: var(--radius-lg);
  background: var(--bg-surface);
  border: 1px solid var(--border);
  min-height: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.repo-header {
  padding: var(--space-5) var(--space-6);
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.repo-header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.repo-title {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text);
  margin: 0;
}

.repo-count {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.filter-bar {
  display: flex;
  gap: var(--space-3);
  align-items: center;
}

.search-box {
  flex: 1;
}

.repo-search {
  width: 100%;
  padding: var(--space-2) var(--space-4);
  background: var(--bg-elevated);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  color: var(--text);
  font-size: 0.875rem;
  transition: all var(--transition-fast);
}

.repo-search:focus {
  border-color: var(--accent-primary);
  outline: none;
  box-shadow: 0 0 0 2px rgba(var(--accent-primary-rgb), 0.2);
}

.sort-box {
  width: 140px;
}

.repo-sort {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-elevated);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  color: var(--text);
  font-size: 0.8125rem;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='white' viewBox='0 0 16 16'%3E%3Cpath d='M7.247 11.14 2.451 5.658C1.885 5.013 2.345 4 3.204 4h9.592a1 1 0 0 1 .753 1.659l-4.796 5.48a1 1 0 0 1-1.506 0z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
  padding-right: 2rem;
}

.repo-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-4);
  background: rgba(0, 0, 0, 0.05);
}

.repo-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-3);
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  cursor: pointer;
  transition: all var(--transition-medium);
}

.repo-item:hover {
  transform: translateY(-2px);
  border-color: var(--accent-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  background: var(--bg-elevated);
}

.repo-item-main {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.repo-item-name {
  font-weight: 600;
  font-size: 1rem;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.repo-item-id {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.enable-btn {
  background: transparent;
  padding: 0;
  border: none;
}

.toggle {
  width: 40px;
  height: 20px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 10px;
  position: relative;
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.toggle::after {
  content: '';
  position: absolute;
  width: 14px;
  height: 14px;
  background: var(--text-muted);
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: all var(--transition-medium);
}

.repo-item:hover .toggle {
  border-color: var(--accent-primary);
}

.enable-btn:hover .toggle::after {
  background: var(--accent-primary);
}

.repo-empty, .repo-loading {
  padding: var(--space-5);
  text-align: center;
  color: var(--text-muted);
  font-size: 0.85rem;
}
</style>
