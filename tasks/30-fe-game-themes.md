# Task 30 — Frontend: Per-Game Theme System

## Goal

Apply game-specific visual themes (colors/background/tokens) based on selected game.

## Context

Depends on: task 28.

## Deliverables

### Theme model

Define theme mapping by game ID (suggested: `frontend/src/utils/gameTheme.ts`):

- `eu5` theme tokens
- `vic3` theme tokens
- fallback default theme

### Runtime apply

On active game change, set theme on root container (class or CSS variables).

### Styles

Adjust existing components to use theme tokens where needed:

- shell backgrounds
- panel surfaces
- primary accents
- state highlights (selected/hover)

## Acceptance criteria

- Switching game updates launcher theme immediately.
- Themes are isolated by game and do not require page reload.
- Existing component readability/contrast remains acceptable.

## Notes for agent

- Do not hardcode colors deep in components; use central tokens.
- Keep placeholder assets/theme values easy to replace later.

