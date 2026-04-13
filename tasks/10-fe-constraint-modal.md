# Task 10 — Frontend: Constraint Modal

## Goal

Build the modal dialog for viewing and editing load-order constraints on a specific mod.

## Context

Depends on: task 05 (stores), task 06 (design system — `BaseModal`), task 09 (opened via context menu).
Uses `useConstraintsStore` and `useModsStore`.

## Deliverables

### `frontend/src/components/ConstraintModal.vue`

Props:
```ts
props: {
  open: boolean
  modID: string      // the mod being configured
}
emits: ['close']
```

The modal has two sections:

#### Section 1 — Existing constraints

List all constraints where this mod is involved:
- "**ModA** always loads after **this mod**"
- "**This mod** always loads after **ModB**"

Each row has a delete (×) button that calls `constraintsStore.remove(from, to)`.

Mod names are resolved from `modsStore.allMods` by ID.

#### Section 2 — Add new constraint

A two-part form:

```
[This mod] loads [after ▾] [____________ mod picker]
                   before
```

- Direction dropdown: "after" / "before"
- Mod picker: a searchable dropdown/combobox of all other known mods (from `modsStore.allMods`, excluding the current mod and mods already in an existing constraint with this mod for this direction)
- "Add" button → calls `constraintsStore.add(from, to)` with correct direction mapping

Constraint direction mapping:
- "This mod loads **after** X" → `add(thisModID, X)`  (thisModID depends on X)
- "This mod loads **before** X" → `add(X, thisModID)` (X depends on thisModID)

#### Error handling

If `constraintsStore.add()` returns an error (which it will if the backend detects an immediate cycle), show an inline error message below the form.

### `frontend/src/components/ui/ModPicker.vue`

Reusable searchable dropdown for selecting a mod from a list.

Props: `mods: Mod[]`, `modelValue: string | null`
Emits: `update:modelValue`

Behavior: type to filter, click to select, shows mod name in field after selection. Keyboard navigable (arrow keys + Enter).

## Acceptance criteria

- Modal opens with the correct mod's constraints already listed.
- Deleting a constraint removes it from the list immediately (optimistic UI is fine — refetch on error).
- Adding a valid constraint adds it to the list.
- The mod picker filters reactively as user types.
- Adding a constraint that would create a cycle shows an error message inline (does not crash or silently fail).
- Modal closes on Escape and × button.

## Notes for agent

- The mod picker does not need to be a native `<select>` — a custom div-based dropdown is preferred for styling consistency.
- Keep the "after/before" language in the UI even though the internal representation is always a directed "loads-after" edge. The conversion logic is a few lines.
- Wrap the entire modal body in a `<form>` element is NOT required — use button `@click` handlers.
