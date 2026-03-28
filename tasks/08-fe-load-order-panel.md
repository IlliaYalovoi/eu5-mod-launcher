# Task 08 — Frontend: Load Order Panel (Drag & Drop)

## Goal

Build the panel that shows only **enabled** mods in their load order, with drag-and-drop reordering.

## Context

Depends on: task 05 (stores), task 06 (design system), task 07 (ModCard for reference).
Uses `useLoadOrderStore` and `useModsStore`.

Install dependency: `vuedraggable@next` (Vue 3 compatible, wraps SortableJS).

## Deliverables

### `frontend/src/components/LoadOrderPanel.vue`

Main panel component.

Displays:
- Panel title: "Load Order"
- Count of active mods: "12 mods active"
- Autosort button (calls `loadOrderStore.autosort()`)
- Cycle error alert if `loadOrderStore.autosortError` is set
- Draggable list of `LoadOrderItem` components

Behavior:
- Derives the displayed list by joining `loadOrderStore.orderedIDs` with `modsStore.allMods` to get full mod objects (IDs alone aren't enough to display names).
- On drag end: calls `loadOrderStore.persist(newIDs)` with the updated order.
- Persists immediately on drop, not on a separate "save" button.

### `frontend/src/components/LoadOrderItem.vue`

Single row in the load order list.

Displays:
- Drag handle icon (left side, cursor: grab)
- Load index number (1-based, fixed width, monospace)
- Mod name
- Right-click opens context menu (emit `contextmenu` event with mod ID and mouse position up to parent — actual context menu is task 09)

Props: `mod: Mod`, `index: number`
Emits: `contextmenu` with `{ modID: string, x: number, y: number }`

### Drag-and-drop setup

Use `vuedraggable`:

```vue
<draggable
  v-model="orderedMods"
  item-key="ID"
  handle=".drag-handle"
  animation="150"
  @end="onDragEnd"
>
```

The `v-model` array is a local computed/ref derived from the store. On `@end`, sync back to store.

## Acceptance criteria

- Only enabled mods appear in this panel.
- Dragging a row and releasing updates the visible order immediately.
- Order is persisted to Go backend after each drag.
- Load index numbers update live as order changes.
- Right-click on a row emits the `contextmenu` event (context menu itself is wired in task 09).
- Autosort button calls the store action; if a cycle error is returned, a red alert appears with the error text.

## Notes for agent

- Install vuedraggable: `cd frontend && npm install vuedraggable@next`
- Do not implement the context menu popup here — just emit the event. The wiring happens in a parent view or `App.vue`.
- The local `orderedMods` array must be kept in sync with the store after every drag — do not let it diverge.
