<script setup lang="ts">
import { computed } from 'vue'
import { useModsStore } from '../stores/mods'
import ModToggle from './ui/ModToggle.vue'

const props = defineProps<{
  modID: string
}>()

const emit = defineEmits<{
  (event: 'contextmenu', payload: { modID: string; x: number; y: number }): void
  (event: 'open-constraints', modID: string): void
  (event: 'select', modID: string): void
}>()

const modsStore = useModsStore()
const mod = computed(() => modsStore.getMod(props.modID))

function onContextMenu(event: MouseEvent): void {
  event.preventDefault()
  emit('contextmenu', {
    modID: props.modID,
    x: event.clientX,
    y: event.clientY,
  })
}

function toggleEnabled(value: boolean) {
  modsStore.setEnabled(props.modID, value)
}
</script>

<template>
  <div v-if="mod" class="mod-row" @contextmenu.prevent="onContextMenu" @click="emit('select', mod.ID)">
    <ModToggle :model-value="mod.Enabled" @update:model-value="toggleEnabled" />
    <span class="name">{{ mod.Name }}</span>
    <span class="version">v{{ mod.Version || '?' }}</span>
  </div>
</template>

<style scoped>
.mod-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.02);
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}

.mod-row:hover {
  background: rgba(255,255,255,0.05);
}

.name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.version {
  color: var(--color-text-muted);
  font-size: 11px;
}
</style>
