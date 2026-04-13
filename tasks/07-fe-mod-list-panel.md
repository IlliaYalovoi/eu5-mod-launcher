# Task 07 — Frontend: Mod List Panel (All Mods)

## Goal

Build the left/main panel that shows all discovered mods. The user can search, see mod info, and toggle mods enabled/disabled from here.

## Context

Depends on: task 05 (stores), task 06 (design system). Uses `useModsStore`. This panel is **read-only about load order** — it doesn't care about position, only enabled/disabled state.

## Deliverables

### `frontend/src/components/ModListPanel.vue`

Top-level panel component. Renders:
- A search/filter input at the top
- Scrollable list of `ModCard` components
- Empty state when no mods found or search yields nothing

Props: none (reads directly from store).

Behavior:
- On mount: calls `modsStore.fetchAll()` if `allMods` is empty.
- Shows a loading spinner while `isLoading` is true.
- Shows error message if `error` is set.
- Search filters by mod name (case-insensitive, substring match).

### `frontend/src/components/ModCard.vue`

Single mod row/card. Displays:
- Mod name (primary text)
- Version and tags (secondary, using `BaseTag`)
- Enabled/disabled toggle (`BaseBadge` or a toggle switch)
- Thumbnail if `ThumbnailPath` is set (small, fixed 40×40px, fallback icon)

Props: `mod: Mod`
Emits: `toggle` (called when enable/disable is clicked)

The parent `ModListPanel` handles the actual store call on `toggle`.

### `frontend/src/components/ui/SearchInput.vue`

Reusable search input with clear button (×). Emits `update:modelValue`. Uses `--font-mono` for input text to reinforce the tactical aesthetic.

## Acceptance criteria

- Panel renders all mods returned by `GetAllMods()` with correct metadata.
- Search input filters the visible list reactively.
- Toggling a mod calls `modsStore.setEnabled()` and the badge updates.
- Panel is independently scrollable (does not scroll the whole window).
- Empty state message is shown when mod list is genuinely empty (not just loading).

## Notes for agent

- Do **not** implement drag-and-drop here — that is the load order panel's job (task 08).
- `ModCard` should emit events up rather than calling the store directly, keeping it reusable.
- Thumbnail `img` tag must have `loading="lazy"` and a fallback `onerror` handler that swaps to an inline SVG placeholder.
