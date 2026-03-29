<script setup lang="ts">
import type { Mod } from '../types'
import BaseBadge from './ui/BaseBadge.vue'
import BaseTag from './ui/BaseTag.vue'

const props = defineProps<{
  mod: Mod
}>()

const emit = defineEmits<{
  (event: 'toggle'): void
}>()

const fallbackThumbnail =
  "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='40' height='40' viewBox='0 0 40 40'%3E%3Crect width='40' height='40' rx='6' fill='%23222a35'/%3E%3Cg fill='none' stroke='%23b9b09b' stroke-width='2'%3E%3Cpath d='M8 27l9-9 7 7 4-4 4 6'/%3E%3Crect x='8' y='8' width='24' height='24' rx='4'/%3E%3C/g%3E%3C/svg%3E"

function onImageError(event: Event): void {
  const image = event.target as HTMLImageElement
  image.src = fallbackThumbnail
}

function onToggle(): void {
  emit('toggle')
}
</script>

<template>
  <article class="mod-card">
    <img
      class="thumbnail"
      :src="props.mod.ThumbnailPath || fallbackThumbnail"
      :alt="`${props.mod.Name} thumbnail`"
      loading="lazy"
      @error="onImageError"
    />

    <div class="info">
      <h3 class="name">{{ props.mod.Name }}</h3>
      <p class="version">Version {{ props.mod.Version || 'Unknown' }}</p>

      <div v-if="props.mod.Tags.length > 0" class="tags">
        <BaseTag v-for="tag in props.mod.Tags" :key="tag" :label="tag" />
      </div>
    </div>

    <button class="toggle" type="button" :aria-label="`Toggle ${props.mod.Name}`" @click="onToggle">
      <BaseBadge :enabled="props.mod.Enabled" />
    </button>
  </article>
</template>

<style scoped>
.mod-card {
  display: grid;
  grid-template-columns: 2.5rem 1fr auto;
  gap: var(--space-3);
  align-items: start;
  padding: var(--space-3);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
}

.thumbnail {
  width: 2.5rem;
  height: 2.5rem;
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-sm);
  object-fit: cover;
  background: var(--color-bg-panel);
}

.info {
  min-width: 0;
}

.name {
  color: var(--color-text-primary);
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.3;
}

.version {
  margin-top: var(--space-1);
  color: var(--color-text-muted);
  font-family: var(--font-mono), monospace;
  font-size: 0.75rem;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.toggle {
  border: 0;
  background: transparent;
  padding: var(--space-1);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.toggle:hover {
  background: var(--color-bg-panel);
}

.toggle:focus-visible {
  outline: var(--border-width) solid var(--color-border-strong);
}
</style>

