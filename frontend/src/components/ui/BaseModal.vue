<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const panelRef = ref<HTMLElement | null>(null)
let previousActiveElement: HTMLElement | null = null

function getFocusableElements(): HTMLElement[] {
  if (!panelRef.value) {
    return []
  }
  const selector = [
    'a[href]',
    'button:not([disabled])',
    'textarea:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(',')
  const nodes = panelRef.value.querySelectorAll<HTMLElement>(selector)
  const focusable: HTMLElement[] = []
  nodes.forEach((node) => focusable.push(node))
  return focusable
}

function focusFirstElement(): void {
  const focusable = getFocusableElements()
  if (focusable.length > 0) {
    focusable[0].focus()
  } else {
    panelRef.value?.focus()
  }
}

function onKeydown(event: KeyboardEvent): void {
  if (!props.open) {
    return
  }
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key !== 'Tab') {
    return
  }
  const focusable = getFocusableElements()
  if (focusable.length === 0) {
    event.preventDefault()
    panelRef.value?.focus()
    return
  }
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement
  if (event.shiftKey && active === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && active === last) {
    event.preventDefault()
    first.focus()
  }
}

function onOverlayClick(event: MouseEvent): void {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      previousActiveElement = document.activeElement as HTMLElement | null
      window.addEventListener('keydown', onKeydown)
      void nextTick().then(() => focusFirstElement())
      return
    }
    window.removeEventListener('keydown', onKeydown)
    previousActiveElement?.focus()
    previousActiveElement = null
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <Transition name="modal-fade">
    <div v-if="open" class="modal-overlay" @click="onOverlayClick">
      <section ref="panelRef" class="modal-panel" role="dialog" aria-modal="true" tabindex="-1">
        <slot />
      </section>
    </div>
  </Transition>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 1000;
}

.modal-panel {
  width: min(36rem, 100%);
  max-height: 100%;
  overflow: auto;
  padding: var(--space-5);
  border: var(--border-width) solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity var(--transition-base), transform var(--transition-base);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: translateY(var(--space-2));
}
</style>


