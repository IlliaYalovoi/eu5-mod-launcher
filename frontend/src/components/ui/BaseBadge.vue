<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    enabled?: boolean
  }>(),
  {
    enabled: false,
  },
)
</script>

<template>
  <span class="base-badge" :class="{ 'base-badge--enabled': props.enabled }">
    <span class="dot" aria-hidden="true" />
    {{ props.enabled ? 'Enabled' : 'Disabled' }}
  </span>
</template>

<style scoped>
.base-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-disabled);
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: var(--radius-pill);
  background: var(--color-disabled);
}

.base-badge--enabled {
  color: var(--color-success);
}

.base-badge--enabled .dot {
  background: var(--color-success);
}
</style>

