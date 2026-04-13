# Task 06 — Frontend: Design System & Global Styles

## Goal

Establish the visual identity of the launcher. Define CSS custom properties, typography, spacing scale, and base component stubs that all subsequent UI tasks will use. No logic — only look and feel.

## Context

This is a **game mod launcher** — it should feel like it belongs next to a grand strategy game. Think: dark interface, tactical/cartographic aesthetic, serious but not grim. Avoid generic SaaS dashboard looks.

Suggested direction (agent may refine): **dark parchment-meets-command-terminal**. Deep desaturated background, warm amber/sepia accents, monospaced or slightly editorial headings, crisp borders instead of soft shadows.

## Deliverables

### `frontend/src/assets/main.css` (global stylesheet)

Define all CSS custom properties on `:root`:

```css
/* Colors */
--color-bg-base        /* main window background */
--color-bg-panel       /* panel / sidebar background */
--color-bg-elevated    /* cards, dropdowns, modals */
--color-border         /* default border color */
--color-border-strong  /* focused / active border */
--color-text-primary
--color-text-secondary
--color-text-muted
--color-accent         /* primary action color */
--color-accent-hover
--color-danger         /* error / cycle warning */
--color-success        /* enabled state */

/* Typography */
--font-display   /* headings — something with character */
--font-body      /* readable body text */
--font-mono      /* mod IDs, paths, technical strings */

/* Spacing scale (4px base) */
--space-1 through --space-8

/* Borders */
--radius-sm
--radius-md
--border-width: 1px

/* Transitions */
--transition-fast: 120ms ease
--transition-base: 200ms ease
```

Apply a CSS reset and base body styles. No scrollbar on the main window (panels scroll internally).

### `frontend/src/components/ui/BaseButton.vue`

Variants: `primary`, `ghost`, `danger`. Props: `variant`, `disabled`, `loading`. Small animated spinner for loading state (CSS only).

### `frontend/src/components/ui/BaseTag.vue`

Displays a single mod tag. Small pill shape. Uses `--color-accent` tint.

### `frontend/src/components/ui/BaseBadge.vue`

Shows enabled/disabled status. Two states: green dot + "Enabled", grey dot + "Disabled".

### `frontend/src/components/ui/BaseModal.vue`

Slot-based modal with overlay. Props: `open` (bool), emits `close`. Smooth open/close transition. Traps focus. Closes on Escape key and overlay click.

### `frontend/src/App.vue`

Shell layout: fixed titlebar area, left sidebar (~280px), main content area. No content yet — just the layout skeleton using CSS grid or flexbox.

## Acceptance criteria

- All CSS variables defined and used consistently (no hardcoded hex values anywhere except the variable definitions themselves).
- `BaseModal` opens and closes cleanly with transition, Escape key works.
- `BaseButton` variants are visually distinct.
- Running `wails dev` shows the shell layout with no console errors.

## Notes for agent

- Pick **one** display font from Google Fonts that fits the aesthetic (load via `@import` in CSS or `index.html`). Suggested candidates: `Cinzel`, `Libre Baskerville`, `IM Fell English`, `Playfair Display` — or choose something better-fitting.
- All components in this task are **presentational only** — no store imports.
- Keep component files under ~100 lines each. Complexity lives in CSS, not script.
