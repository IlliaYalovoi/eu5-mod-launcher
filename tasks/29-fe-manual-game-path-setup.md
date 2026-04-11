# Task 29 — Frontend: Manual Game Path Setup Popup

## Goal

When user selects an undetected game, offer explicit manual path setup for install/documents dirs.

## Context

Depends on: tasks 27-28.

## Deliverables

### Popup/modal

Create modal shown on undetected game click with:

- selected game name
- install directory field + folder picker
- documents directory field + folder picker
- validate + save actions

### Backend calls

Use `SetGamePaths(gameID, installDir, documentsDir)`.

### UX behavior

- block save until required paths valid/non-empty
- show inline backend validation errors
- on success:
  - close modal
  - refresh game detection list
  - activate game if now valid

## Acceptance criteria

- User can manually configure missing game paths from UI.
- Invalid/manual setup errors are clear.
- Successful setup immediately enables game selection.

## Notes for agent

- Keep confirmation explicit; no silent auto-save.
- Reuse existing picker patterns used in settings panel.

