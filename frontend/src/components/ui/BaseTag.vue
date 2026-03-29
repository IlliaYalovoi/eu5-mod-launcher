<script setup lang="ts">
defineProps<{
  label: string
}>()
</script>

<template>
  <span class="base-tag">{{ label }}</span>
</template>

<style scoped>
.base-tag {
  display: inline-flex;
  align-items: center;
  padding: var(--space-1) var(--space-3);
  border: var(--border-width) solid var(--color-accent);
  border-radius: 999px;
  color: var(--color-accent);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}
</style>

