<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useLoadOrderStore } from '../stores/loadorder'

const emit = defineEmits<{
  (event: 'open-constraints', modID: string): void
}>()

const loadOrderStore = useLoadOrderStore()
const { autosortError } = storeToRefs(loadOrderStore)

const cycleText = computed(() => {
  const source = autosortError.value || ''
  const lowered = source.toLowerCase()
  const marker = lowered.lastIndexOf('cycle detected:')
  if (marker >= 0) {
    return source.slice(marker + 'cycle detected:'.length).trim()
  }
  return source
})

const cycleNodes = computed(() => {
  const text = cycleText.value
  if (!text) {
    return [] as string[]
  }
  return text
    .split('->')
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
})

const cycleDiagram = computed(() => {
  if (cycleNodes.value.length === 0) {
    return ''
  }
  return cycleNodes.value.join('  ->  ')
})

const firstCycleMod = computed(() => (cycleNodes.value.length > 0 ? cycleNodes.value[0] : ''))

function onDismiss(): void {
  loadOrderStore.clearAutosortError()
}

function onOpenConstraints(): void {
  if (firstCycleMod.value) {
    emit('open-constraints', firstCycleMod.value)
  }
}
</script>

<template>
  <!-- Hidden: cycle error is now shown inline in AutosortButton -->
  <!-- Kept for backward compatibility in case store is still referenced -->
  <div v-if="false" />
</template>

