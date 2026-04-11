# Task 28 — Frontend: Multi-Game Sidebar and Selection

## Goal

Add left sidebar for game switching with detected-state UX.

## Context

Depends on: task 27.
Launcher currently assumes single game context.

## Deliverables

### Sidebar UI

Replace/extend left area with game list:

- icon placeholder + name per game
- detected game: colored + clickable
- undetected game: grey + non-active style
- detected games shown first

### Interaction

- click detected game => switch active game context
- click undetected game => open setup prompt flow (task 29)

### Store changes

Create/extend game store (suggested: `frontend/src/stores/games.ts`):

- `supportedGames`
- `activeGameID`
- `fetchSupportedGames()`
- `setActiveGame(gameID)`

### Wiring

On game switch, trigger data refresh for mod list/load order/settings for selected game.

## Acceptance criteria

- Sidebar shows EU5 + Vic3 with proper enabled/disabled visuals.
- Detected games always sorted to top.
- Switching detected game updates visible launcher state.

## Notes for agent

- Use placeholder icons now; keep icon contract extensible.
- Keep keyboard focus/aria semantics for accessibility.

