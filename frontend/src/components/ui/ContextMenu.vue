<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

interface MenuItem {
  id: string
  label: string
  icon?: string
  danger?: boolean
  disabled?: boolean
  children?: MenuItem[]
}

const props = defineProps<{
  open: boolean
  x: number
  y: number
  items: MenuItem[]
  targetID: string
}>()

const emit = defineEmits<{
  (event: 'select', payload: { itemID: string; targetID: string }): void
  (event: 'close'): void
}>()

const menuRef = ref<HTMLElement | null>(null)
const left = ref(0)
const top = ref(0)
const submenuStyles = ref<Record<string, Record<string, string>>>({})
const submenuDirections = ref<Record<string, 'left' | 'right'>>({})

function reposition(): void {
  const menu = menuRef.value
  if (!menu) {
    return
  }

  const padding = 8
  let nextLeft = props.x
  let nextTop = props.y

  const maxLeft = window.innerWidth - menu.offsetWidth - padding
  const maxTop = window.innerHeight - menu.offsetHeight - padding

  if (nextLeft > maxLeft) {
    nextLeft = Math.max(padding, props.x - menu.offsetWidth)
  }
  if (nextTop > maxTop) {
    nextTop = Math.max(padding, props.y - menu.offsetHeight)
  }

  left.value = Math.max(padding, nextLeft)
  top.value = Math.max(padding, nextTop)
}

function requestClose(): void {
  emit('close')
}

function onOutsideClick(event: MouseEvent): void {
  if (!props.open) {
    return
  }
  const target = event.target as Node | null
  if (!target) {
    return
  }
  if (menuRef.value && !menuRef.value.contains(target)) {
    requestClose()
  }
}

function onKeydown(event: KeyboardEvent): void {
  if (props.open && event.key === 'Escape') {
    event.preventDefault()
    requestClose()
  }
}

function onScroll(): void {
  if (props.open) {
    requestClose()
  }
}

function onSelect(item: MenuItem): void {
  if (item.disabled || (item.children && item.children.length > 0)) {
    return
  }
  emit('select', { itemID: item.id, targetID: props.targetID })
  requestClose()
}

function onSubmenuEnter(itemID: string, event: MouseEvent): void {
  const node = event.currentTarget as HTMLElement | null
  if (!node) {
    return
  }

  const submenu = node.querySelector<HTMLElement>('.submenu')
  if (!submenu) {
    return
  }

  const nodeRect = node.getBoundingClientRect()
  const submenuRect = submenu.getBoundingClientRect()
  const padding = 8
  const gap = 4

  let x = nodeRect.width + gap
  let y = 0
  let direction: 'left' | 'right' = 'right'

  if (nodeRect.right + gap + submenuRect.width + padding > window.innerWidth) {
    x = -submenuRect.width - gap
    direction = 'left'
  }

  const overflowBottom = nodeRect.top + submenuRect.height + padding - window.innerHeight
  if (overflowBottom > 0) {
    y = -overflowBottom
  }

  const minTopOffset = -nodeRect.top + padding
  if (y < minTopOffset) {
    y = minTopOffset
  }

  submenuDirections.value[itemID] = direction
  submenuStyles.value[itemID] = {
    left: `${x}px`,
    top: `${y}px`,
  }
}

function onSafeZoneMousedown(event: MouseEvent): void {
  // Keep hover-bridge behavior, but left click in the bridge closes the whole context menu.
  if (event.button !== 0) {
    return
  }
  event.preventDefault()
  event.stopPropagation()
  requestClose()
}

watch(
  () => [props.open, props.x, props.y],
  ([isOpen]) => {
    if (!isOpen) {
      return
    }
    void nextTick().then(() => {
      reposition()
    })
  },
)

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      window.addEventListener('mousedown', onOutsideClick)
      window.addEventListener('keydown', onKeydown)
      window.addEventListener('scroll', onScroll, true)
      return
    }
    window.removeEventListener('mousedown', onOutsideClick)
    window.removeEventListener('keydown', onKeydown)
    window.removeEventListener('scroll', onScroll, true)
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  window.removeEventListener('mousedown', onOutsideClick)
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('scroll', onScroll, true)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="context-menu-fade">
      <div
        v-if="open"
        ref="menuRef"
        class="context-menu"
        role="menu"
        :style="{ left: `${left}px`, top: `${top}px` }"
      >
        <div
          v-for="item in items"
          :key="item.id"
          class="menu-node"
          :class="{
            'menu-node--disabled': item.disabled,
            'menu-node--submenu-left': submenuDirections[item.id] === 'left',
          }"
          @mouseenter="onSubmenuEnter(item.id, $event)"
        >
          <button
            class="menu-item"
            :class="{ 'menu-item--danger': item.danger, 'menu-item--disabled': item.disabled, 'menu-item--has-children': item.children }"
            type="button"
            :disabled="item.disabled && !item.children"
            @click="onSelect(item)"
          >
            <span v-if="item.icon" class="icon" aria-hidden="true">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
            <span v-if="item.children && item.children.length > 0" class="submenu-arrow" aria-hidden="true">▸</span>
          </button>

          <div
            v-if="item.children && item.children.length > 0 && submenuDirections[item.id] === 'left'"
            class="submenu-safe-zone submenu-safe-zone--left"
            @mousedown="onSafeZoneMousedown"
          />

          <div
            v-if="item.children && item.children.length > 0 && submenuDirections[item.id] !== 'left'"
            class="submenu-safe-zone submenu-safe-zone--right"
            @mousedown="onSafeZoneMousedown"
          />

          <div v-if="item.children && item.children.length > 0" class="submenu" role="menu" :style="submenuStyles[item.id]">
            <button
              v-for="child in item.children"
              :key="child.id"
              class="menu-item"
              :class="{ 'menu-item--danger': child.danger, 'menu-item--disabled': child.disabled }"
              type="button"
              :disabled="child.disabled"
              @click="onSelect(child)"
            >
              <span v-if="child.icon" class="icon" aria-hidden="true">{{ child.icon }}</span>
              <span>{{ child.label }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.context-menu {
  position: fixed;
  z-index: 1200;
  display: flex;
  flex-direction: column;
  min-width: 13rem;
  padding: var(--space-2);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-panel);
  opacity: 0.95;
  backdrop-filter: blur(8px);
  box-shadow: 0 10px 25px rgba(0,0,0,0.5);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 2rem;
  padding: var(--space-2) var(--space-3);
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text);
  text-align: left;
  cursor: pointer;
}

.menu-item--has-children {
  justify-content: space-between;
}

.menu-node {
  position: relative;
}

.menu-node--disabled .submenu {
  display: none;
}

.submenu-safe-zone {
  position: absolute;
  top: 0;
  width: 1.25rem;
  height: 100%;
  z-index: 1202;
  background: transparent;
}

.submenu-safe-zone--left {
  right: 100%;
}

.submenu-safe-zone--right {
  left: 100%;
}

.submenu {
  position: absolute;
  top: 0;
  left: calc(100% + var(--space-1));
  z-index: 1201;
  display: none;
  flex-direction: column;
  min-width: 14rem;
  max-height: min(22rem, calc(100vh - 1rem));
  overflow: auto;
  padding: var(--space-2);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-panel);
  opacity: 0.95;
  backdrop-filter: blur(8px);
}

.menu-node:hover .submenu,
.menu-node:focus-within .submenu {
  display: flex;
}

.submenu-arrow {
  margin-left: var(--space-3);
  color: var(--text-muted);
}

.menu-item:hover:not(:disabled) {
  background: var(--bg-elevated);
}

.menu-item:focus-visible {
  outline: 1px solid var(--accent);
}

.menu-item--danger {
  color: #ef4444;
}

.menu-item--disabled {
  color: var(--text-muted);
}

.menu-item:disabled {
  cursor: not-allowed;
}

.icon {
  width: 1rem;
  text-align: center;
}

.context-menu-fade-enter-active,
.context-menu-fade-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}

.context-menu-fade-enter-from,
.context-menu-fade-leave-to {
  opacity: 0;
  transform: translateY(var(--space-1));
}
</style>

