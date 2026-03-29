<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue?: string
    placeholder?: string
  }>(),
  {
    modelValue: '',
    placeholder: 'Search mods...',
  },
)

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

function onInput(event: Event): void {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

function clearValue(): void {
  emit('update:modelValue', '')
}
</script>

<template>
  <div class="search-input">
    <input
      :value="props.modelValue"
      :placeholder="props.placeholder"
      type="text"
      spellcheck="false"
      @input="onInput"
    />
    <button
      v-if="props.modelValue"
      class="clear-button"
      type="button"
      aria-label="Clear search"
      @click="clearValue"
    >
      ×
    </button>
  </div>
</template>

<style scoped>
.search-input {
  position: relative;
}

.search-input input {
  width: 100%;
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-7) var(--space-2) var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  font-family: var(--font-mono), monospace;
  transition: border-color var(--transition-fast), background var(--transition-fast);
}

.search-input input::placeholder {
  color: var(--color-text-muted);
}

.search-input input:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.clear-button {
  position: absolute;
  top: 50%;
  right: var(--space-2);
  transform: translateY(-50%);
  width: 1.5rem;
  height: 1.5rem;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
  transition: color var(--transition-fast), background var(--transition-fast);
}

.clear-button:hover {
  background: var(--color-bg-panel);
  color: var(--color-text-primary);
}

.clear-button:focus-visible {
  outline: var(--border-width) solid var(--color-border-strong);
}
</style>

