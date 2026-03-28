# Task 09 — Frontend: Right-Click Context Menu

## Goal

Build a reusable, portal-based context menu component that appears on right-click over a mod in the load order panel.

## Context

Depends on: task 06 (design system).
This component is **fully presentational and generic** — it takes a list of menu items as props and emits action events. It has no store imports.

The context menu is mounted at the `App.vue` level (not inside `LoadOrderPanel`) to avoid z-index and overflow clipping issues.

## Deliverables

### `frontend/src/components/ui/ContextMenu.vue`

```ts
interface MenuItem {
  id: string
  label: string
  icon?: string      // optional: a single emoji or SVG string
  danger?: boolean   // renders red
  disabled?: boolean
}

// Props
props: {
  open: boolean
  x: number          // screen x position
  y: number          // screen y position
  items: MenuItem[]
  targetID: string   // which mod ID was right-clicked (passed through for context)
}

// Emits
emit('select', { itemID: string, targetID: string })
emit('close')
```

Behavior:
- Positioned absolutely at `(x, y)` with `position: fixed`.
- Auto-adjusts if near screen edge (flip left or flip up if it would overflow viewport).
- Closes on: outside click, Escape key, menu item selected, scroll.
- Animated: quick fade + slight translate-y on open.
- Only one instance exists — controlled by `App.vue`.

### `frontend/src/App.vue` update

Add context menu state:
```ts
const contextMenu = reactive({
  open: false,
  x: 0,
  y: 0,
  targetID: ''
})

function openContextMenu(e: { modID: string, x: number, y: number }) { ... }
function closeContextMenu() { ... }
function handleMenuAction(e: { itemID: string, targetID: string }) { ... }
```

Define the items array for the mod context menu:
- "Add constraint..." → opens constraint modal (task 10)
- "View constraints" → opens constraint modal in view mode
- "Move to top"
- "Move to bottom"
- "Disable mod"

Wire `LoadOrderItem`'s `contextmenu` event → `openContextMenu`.

## Acceptance criteria

- Right-clicking a load order item opens the menu at the correct cursor position.
- Menu does not overflow the window edges.
- Pressing Escape or clicking outside closes the menu.
- Selecting "Disable mod" calls `modsStore.setEnabled(targetID, false)` and closes menu.
- "Move to top" / "Move to bottom" reorders and persists via `loadOrderStore.persist()`.
- Menu items marked `danger: true` render in `--color-danger`.

## Notes for agent

- Prevent the browser's native context menu with `@contextmenu.prevent` on `LoadOrderItem`.
- Use Vue's `<Teleport to="body">` to render the menu outside the component tree.
- Do not use any context menu library — implement from scratch (~80 lines).
