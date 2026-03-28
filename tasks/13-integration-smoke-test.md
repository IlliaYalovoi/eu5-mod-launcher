# Task 13 — Integration Smoke Test

## Goal

Manually verify the complete application works end-to-end after all previous tasks are done. This is a **human-executed checklist**, not an automated test.

## Setup

1. Create a fake mods directory with at least 4 mock mod subdirectories:
   ```
   ~/fake-mods/
     mod_alpha/descriptor.mod
     mod_beta/descriptor.mod
     mod_gamma/descriptor.mod
     mod_delta/descriptor.mod
   ```

2. Each `descriptor.mod` should have at minimum:
   ```
   name="Mod Alpha"
   version="1.0"
   tags={"economy","military"}
   ```

3. Run `wails dev` and point Settings → Mods Directory to `~/fake-mods/`.

---

## Checklist

### Mod Discovery
- [ ] All 4 mods appear in the mod list panel
- [ ] Mod names and tags display correctly
- [ ] Search filters the list correctly
- [ ] A mod with no `descriptor.mod` does NOT appear (create a 5th empty directory to test)

### Enable / Disable
- [ ] Toggling a mod enabled moves it into the load order panel
- [ ] Toggling a mod disabled removes it from the load order panel
- [ ] Enabled state persists after app restart (close and reopen `wails dev`)

### Load Order
- [ ] Drag a mod row up and down — order changes visually
- [ ] After dragging, new order persists after app restart
- [ ] Load index numbers are correct and contiguous (1, 2, 3...)

### Context Menu
- [ ] Right-clicking a load order item shows the context menu
- [ ] Menu appears near cursor, not off-screen
- [ ] Pressing Escape closes the menu
- [ ] Clicking outside closes the menu
- [ ] "Move to top" works
- [ ] "Move to bottom" works
- [ ] "Disable mod" disables and removes from list

### Constraints
- [ ] Right-click → "Add constraint..." opens the constraint modal
- [ ] Add constraint: "Mod Alpha loads after Mod Beta" → appears in list
- [ ] Add constraint: "Mod Alpha loads before Mod Gamma" → appears in list (reversed direction)
- [ ] Delete a constraint — it disappears from the list
- [ ] Constraints persist after app restart

### Autosort
- [ ] With constraints set (Alpha after Beta), click Autosort → Alpha moves after Beta
- [ ] Create a cycle: Alpha → Beta, Beta → Alpha
- [ ] Click Autosort → Cycle error panel appears naming both mods
- [ ] "Open constraints" button in error panel opens constraint modal for the right mod
- [ ] Dismiss clears the error

### Settings
- [ ] "Browse..." opens a native folder picker
- [ ] Changing mods dir rescans and updates the mod list
- [ ] Config path displayed is a valid, real path

---

## Known limitations to document (not bugs)

- Mods inside `.zip` archives are not scanned (by design, deferred).
- No "Launch Game" button (deferred).
- No profile/preset system (deferred).

---

## If something fails

For each failed checklist item, create a bug note in this format and file it as `tasks/bugs/BUG-NNN.md`:

```markdown
# BUG-NNN — Short description

**Failed step**: [checklist item text]
**Observed**: what actually happened
**Expected**: what should have happened
**Likely cause**: your guess
**Task to fix**: which task file is responsible
```
