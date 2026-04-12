<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { GetAllMods, GetLoadOrder } from '../../wailsjs/go/launcher/App'
import type { Mod } from '../../types'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  (event: 'close'): void
  (event: 'add-mod', modID: string): void
}>()

const allMods = ref<Mod[]>([])
const orderedIDs = ref<string[]>([])
const query = ref('')
const activeIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)
const selectedModIDs = ref<Set<string>>(new Set())

async function load() {
  const [mods, order] = await Promise.all([GetAllMods(), GetLoadOrder()])
  allMods.value = mods as unknown as Mod[]
  orderedIDs.value = order
}

const orderedSet = computed(() => new Set(orderedIDs.value))

const filteredMods = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return allMods.value
  return allMods.value.filter((mod) => {
    if (mod.name.toLowerCase().includes(q)) return true
    if (mod.tags.some((t) => t.toLowerCase().includes(q))) return true
    return false
  })
})

const sortedMods = computed(() => {
  return [...filteredMods.value].sort((a, b) => {
    const aIn = orderedSet.value.has(a.id)
    const bIn = orderedSet.value.has(b.id)
    if (aIn && !bIn) return -1
    if (!aIn && bIn) return 1
    return a.name.localeCompare(b.name)
  })
})

watch(() => props.open, (isOpen) => {
  if (isOpen) { query.value = ''; activeIndex.value = 0; selectedModIDs.value = new Set(); void load(); void nextTick().then(() => inputRef.value?.focus()) }
})

watch(filteredMods, () => { if (activeIndex.value >= sortedMods.value.length) activeIndex.value = Math.max(0, sortedMods.value.length - 1) })

function close(): void { emit('close') }
function onOverlayClick(event: MouseEvent): void { if (event.target === event.currentTarget) close() }
function toggleMod(modID: string): void { const next = new Set(selectedModIDs.value); next.has(modID) ? next.delete(modID) : next.add(modID); selectedModIDs.value = next }
function selectMod(mod: Mod): void { emit('add-mod', mod.id) }
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') { close(); return }
  if (event.key === 'ArrowDown') { event.preventDefault(); activeIndex.value = Math.min(activeIndex.value + 1, sortedMods.value.length - 1); return }
  if (event.key === 'ArrowUp') { event.preventDefault(); activeIndex.value = Math.max(activeIndex.value - 1, 0); return }
  if (event.key === 'Enter') { const mod = sortedMods.value[activeIndex.value]; if (mod) selectMod(mod); return }
}
</script>

<template>
  <Transition name="overlay-fade">
    <div v-if="open" class="overlay" @click="onOverlayClick">
      <div class="search-dialog" role="dialog" aria-modal="true" aria-label="Search and add mods">
        <div class="search-row">
          <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.35-4.35" />
          </svg>
          <input
            ref="inputRef"
            v-model="query"
            class="search-input"
            type="text"
            placeholder="Search mods by name, tag..."
            spellcheck="false"
            @keydown="onKeydown"
          />
          <kbd class="esc-hint">Esc</kbd>
        </div>

        <div class="results">
          <div v-if="sortedMods.length === 0" class="empty">
            No mods match "{{ query }}"
          </div>
          <button
            v-for="(mod, index) in sortedMods"
            :key="mod.id"
            class="result-row"
            :class="{
              'result-row--active': index === activeIndex,
              'result-row--in-load-order': orderedSet.has(mod.id),
            }"
            type="button"
            @click="selectMod(mod)"
            @mouseenter="activeIndex = index"
          >
            <span class="result-name">{{ mod.name }}</span>
            <span v-if="orderedSet.has(mod.id)" class="result-badge">in load order</span>
            <span v-if="mod.tags.length > 0" class="result-tags">{{ mod.tags.slice(0, 2).join(', ') }}</span>
          </button>
        </div>

        <div class="search-footer">
          <span class="hint"><kbd>↑↓</kbd> navigate</span>
          <span class="hint"><kbd>↵</kbd> select</span>
          <span class="hint"><kbd>Esc</kbd> close</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 8vh;
  background: var(--color-overlay);
  z-index: 1200;
}

.search-dialog {
  width: min(36rem, 92vw);
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-elevated);
  overflow: hidden;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
}

.search-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  border-bottom: var(--border-width) solid var(--color-border);
}

.search-icon {
  flex-shrink: 0;
  color: var(--color-text-muted);
}

.search-input {
  flex: 1;
  border: 0;
  background: transparent;
  color: var(--color-text-primary);
  font-size: 1rem;
  font-family: var(--font-body);
  outline: none;
}

.search-input::placeholder {
  color: var(--color-text-muted);
}

.esc-hint {
  padding: 0.15rem 0.4rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  font-family: var(--font-mono), monospace;
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.results {
  flex: 1;
  overflow: auto;
  padding: var(--space-2);
}

.empty {
  padding: var(--space-6);
  text-align: center;
  color: var(--color-text-muted);
}

.result-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: var(--space-3) var(--space-3);
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-primary);
  text-align: left;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.result-row:hover,
.result-row--active {
  background: var(--color-bg-panel);
}

.result-row--in-load-order {
  opacity: 0.7;
}

.result-name {
  flex: 1;
  font-size: 0.9rem;
}

.result-badge {
  flex-shrink: 0;
  padding: 0.1rem 0.5rem;
  border: var(--border-width) solid var(--color-success);
  border-radius: var(--radius-pill);
  font-size: 0.68rem;
  color: var(--color-success);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.result-tags {
  flex-shrink: 0;
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.search-footer {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  border-top: var(--border-width) solid var(--color-border);
  background: var(--color-bg-panel);
}

.hint {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--color-text-muted);
  font-size: 0.72rem;
}

.hint kbd {
  padding: 0.1rem 0.3rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  font-family: var(--font-mono), monospace;
  font-size: 0.68rem;
}

.overlay-fade-enter-active,
.overlay-fade-leave-active {
  transition: opacity var(--transition-base);
}

.overlay-fade-enter-from,
.overlay-fade-leave-to {
  opacity: 0;
}
</style>
