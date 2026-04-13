# Task 22 — Frontend: Steam Mod Details Panel

## Goal

Show enriched Steam workshop details for selected mods in the UI (title, description, thumbnail, workshop link).

## Context

Depends on: tasks 20-21.

## Deliverables

### Store updates

Extend mods/details store with async metadata loading:
- loading/error state per selected mod
- cache-aware retrieval via backend APIs

### New component

`frontend/src/components/ModDetailsPanel.vue`:
- shows selected mod basic info + steam-enriched fields
- thumbnail image with fallback
- workshop link button when available

### Integration

Wire panel into existing layout so selecting a mod updates details panel.

## Acceptance criteria

- Clicking workshop mod shows steam title/description/thumbnail.
- Non-workshop mods show local/basic info without error.
- Loading/error states are clear and non-blocking.

## Notes for agent

- Keep panel reactive and lightweight.
- Avoid refetch storm on rapid selection changes.

