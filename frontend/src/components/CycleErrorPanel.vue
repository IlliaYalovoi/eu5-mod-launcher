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
  <Transition name="cycle-panel-slide">
    <section v-if="autosortError" class="cycle-panel" role="alert">
      <h3 class="title">Constraint Cycle Detected</h3>
      <p class="message">{{ autosortError }}</p>
      <pre v-if="cycleDiagram" class="diagram">{{ cycleDiagram }}</pre>
      <div class="actions">
        <button class="action" type="button" @click="onDismiss">Dismiss</button>
        <button class="action action--primary" type="button" :disabled="!firstCycleMod" @click="onOpenConstraints">
          Open constraints for {{ firstCycleMod || '...' }}
        </button>
      </div>
    </section>
  </Transition>
</template>

<style scoped>
.cycle-panel {
  margin-top: var(--space-4);
  padding: var(--space-4);
  border: var(--border-width) solid var(--color-danger);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-danger);
}

.title {
  font-size: 0.95rem;
  margin-bottom: var(--space-2);
}

.message {
  color: var(--color-text-secondary);
  margin-bottom: var(--space-2);
}

.diagram {
  margin: 0;
  margin-bottom: var(--space-3);
  padding: var(--space-2);
  border-radius: var(--radius-sm);
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.actions {
  display: flex;
  gap: var(--space-2);
}

.action {
  min-height: 2rem;
  padding: 0 var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-primary);
  cursor: pointer;
}

.action--primary {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cycle-panel-slide-enter-active,
.cycle-panel-slide-leave-active {
  transition: opacity var(--transition-base), transform var(--transition-base), max-height var(--transition-base);
  overflow: hidden;
}

.cycle-panel-slide-enter-from,
.cycle-panel-slide-leave-to {
  opacity: 0;
  transform: translateY(calc(-1 * var(--space-2)));
  max-height: 0;
}

.cycle-panel-slide-enter-to,
.cycle-panel-slide-leave-from {
  opacity: 1;
  transform: translateY(0);
  max-height: 16rem;
}
</style>

