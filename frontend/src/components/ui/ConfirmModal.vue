<script setup lang="ts">
import BaseModal from './BaseModal.vue'
import BaseButton from './BaseButton.vue'

const props = defineProps<{
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'confirm'): void
}>()
</script>

<template>
  <BaseModal :open="open" @close="emit('close')">
    <div class="confirm-body">
      <h2 class="title">{{ title }}</h2>
      <p class="message">{{ message }}</p>
      <div class="actions">
        <BaseButton variant="ghost" @click="emit('close')">Cancel</BaseButton>
        <BaseButton :variant="danger ? 'danger' : 'primary'" @click="emit('confirm')">
          {{ confirmLabel || 'Confirm' }}
        </BaseButton>
      </div>
    </div>
  </BaseModal>
</template>

<style scoped>
.confirm-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.title {
  font-family: var(--font-display), serif;
  font-size: 1rem;
  color: var(--text);
}

.message {
  color: var(--text-muted);
  font-size: 0.9rem;
  line-height: 1.5;
}

.actions {
  display: flex;
  gap: var(--space-2);
  justify-content: flex-end;
}
</style>
