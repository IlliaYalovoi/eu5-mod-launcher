<script setup lang="ts">
import { computed } from 'vue'

type ButtonVariant = 'primary' | 'ghost' | 'danger'

const props = withDefaults(
  defineProps<{
    variant?: ButtonVariant
    disabled?: boolean
    loading?: boolean
  }>(),
  {
    variant: 'primary',
    disabled: false,
    loading: false,
  },
)

const classes = computed(() => ['base-button', `base-button--${props.variant}`])
</script>

<template>
  <button :class="classes" :disabled="disabled || loading" type="button">
    <span v-if="loading" class="spinner" aria-hidden="true" />
    <span class="label"><slot /></span>
  </button>
</template>

<style scoped>
.base-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-4);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  font-family: var(--font-body);
  font-weight: 700;
  letter-spacing: 0.02em;
  cursor: pointer;
  transition: background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
}

.base-button:focus-visible {
  outline: none;
  border-color: var(--color-border-strong);
}

.base-button--primary {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--bg-body);
}

.base-button--primary:hover:not(:disabled) {
  background: var(--accent-hover);
  border-color: var(--accent-hover);
}

.base-button--ghost {
  background: transparent;
  color: var(--text);
  border-color: var(--border);
}

.base-button--ghost:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.base-button--danger {
  background: transparent;
  border-color: #ef4444;
  color: #ef4444;
}

.base-button--danger:hover:not(:disabled) {
  background: #ef4444;
  color: white;
}

.base-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: var(--border-width) solid var(--color-bg-base);
  border-top-color: transparent;
  border-radius: var(--radius-pill);
  animation: spin var(--duration-spinner) linear infinite;
}

.base-button--ghost .spinner,
.base-button--danger .spinner {
  border-color: currentColor;
  border-top-color: transparent;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

