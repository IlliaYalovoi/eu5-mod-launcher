# Task 12 — Frontend: Settings Panel

## Goal

Build a settings panel where the user can configure the mods directory path and (optionally) the game executable path.

## Context

Depends on: task 05 (stores), task 06 (design system).
Uses `useSettingsStore`. This panel is low-complexity — the main UX challenge is the file/folder picker, which in Wails requires a Go-side dialog call.

## Deliverables

### Go side — add to `app.go` (small addition to task 04)

```go
// Opens a native OS folder picker dialog. Returns selected path or "" if cancelled.
func (a *App) PickFolder() (string, error)

// Opens a native OS file picker for executables.
func (a *App) PickExecutable() (string, error)
```

Use `github.com/wailsapp/wails/v2/pkg/runtime` dialog functions:
- `runtime.OpenDirectoryDialog(ctx, options)`
- `runtime.OpenFileDialog(ctx, options)`

### `frontend/src/components/SettingsPanel.vue`

A settings form with:

#### Mods Directory
- Label: "Mods Directory"
- Text input showing current path (read-only display, not directly editable)
- "Browse..." button → calls `PickFolder()` binding → updates via `settingsStore.setModsDir()`
- After setting, triggers `modsStore.fetchAll()` to rescan

#### Game Executable (optional, can be left blank)
- Label: "Game Executable (optional)"
- Same pattern: display path + Browse button using `PickExecutable()`
- Stored in settings.json as `"game_exe": "..."`

#### Info section
- Show current config file path (from `GetConfigPath()` — add this trivial getter to `app.go`)
- "Open config folder" link — calls `runtime.BrowserOpenURL` with the folder path as a `file://` URL so the OS opens it in Explorer/Nautilus

### `frontend/src/stores/settings.ts` update

Add `gameExe` ref alongside `modsDir`. Both fetched and persisted symmetrically.

## Acceptance criteria

- "Browse..." opens a native OS folder picker on Windows.
- Selecting a new mods directory immediately triggers a mod rescan.
- Settings survive app restart (persisted to `settings.json`).
- Empty game exe path is valid and displays a helpful placeholder ("Not configured").
- Config path display shows the actual resolved path.

## Notes for agent

- The settings panel can be a slide-in side panel or a separate view — whatever fits the layout from task 06.
- Do not implement "Launch Game" button here — that is out of scope for this task (and the whole project as currently scoped). Just store the path.
- On Linux, `PickFolder` may open a GTK dialog — this is fine and expected.
