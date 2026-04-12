# MOD CHRONICLE UI REFINEMENT (PUPPET UI & CONCEPT B)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refine the UI to match Concept B while stripping frontend state management in favor of a "Puppet" architecture (Backend as sole source of truth for mod lists, groups, and active states).

**Architecture:**
- **State Statelessness**: Frontend components (`LoadOrderPanel`, `ModRepository`) should not maintain their own lists. They call a backend `Refresh` which returns the full data.
- **Backend Orchestration**: Move Group creation/deletion/rules logic entirely to backend Go services.
- **Visual Puppet**: Components render raw data from props/store only.

**Tech Stack:** Go (Backend/Service), Vue 3 (Reactive Template), Pinia (Thin Proxy).

---

### Task 1: Backend Group Management API

**Files:**
- Modify: `internal/launcher/playsets.go` (or relevant service)
- Modify: `internal/launcher/app.go` (Wails exposure)

- [ ] **Step 1: Add backend methods for Group operations**
Expose methods that perform the action AND return the latest `LauncherLayout`.
```go
func (a *App) CreateGroup(name string) (domain.LauncherLayout, error)
func (a *App) DeleteGroup(groupID string) (domain.LauncherLayout, error)
func (a *App) UpdateGroupRule(groupID string, rule string) (domain.LauncherLayout, error)
```

- [ ] **Step 2: Commit**
```bash
git add internal/launcher/*.go
git commit -m "feat(backend): implement group management service methods"
```

---

### Task 2: Refactor Frontend to Puppet State

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/ModRepository.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Strip helper state from LoadOrderPanel**
Remove `blocks` and `localNumberByModID` calculations if possible. Instead, the backend should ideally return the expanded layout structure. Ensure `load()` simply overwrites the store.

- [ ] **Step 2: Stabilize Repository Pane**
Fix the layout shift bug by removing any conditional `v-if` on the repository list during load. Use a permanent layout container.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/components/*.vue
git commit -m "refactor(ui): align with puppet architecture - remove local state"
```

---

### Task 3: Concept B Sidebar & Playset Nesting

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Nest Playset Selector under active game only**
Match Concept B: `playset-dropdown` appears only inside the active game card.

- [ ] **Step 2: Styling**
Apply Concept B margins: `margin-top: 10px; margin-left: 32px;` for the nested area.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/App.vue
git commit -m "feat(ui): nest playset selector under active game card"
```

---

### Task 4: UI Components (Toggles, Rules, Context)

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/ui/ContextMenu.vue`
- Modify: `frontend/src/assets/main.css`

- [ ] **Step 1: Replace X mark with Toggle Switch**
Implement Concept B toggle styles in CSS. The toggle trigger calls backend `DisableMod`/`EnableMod` then reloads everything.

- [ ] **Step 2: Fix Context Menu readable background**
Adjust opacity and blur to ensure readability.

- [ ] **Step 3: Add "Manage Groups" button to toolbar**
Triggers a modal or inline management view.

- [ ] **Step 4: Commit**
```bash
git add frontend/src/assets/main.css frontend/src/components/*.vue
git commit -m "feat(ui): implement Concept B toggles and group management button"
```

---

### Task 5: Theme & Texture Polishing

**Files:**
- Modify: `frontend/src/assets/main.css`

- [ ] **Step 1: Apply Background Texture**
Add the linen/parchment SVG pattern to `body` from `concept_B.html`.

- [ ] **Step 2: Single-line Title Fixes**
Adjust typography to prevent "Load Order Chronicle" from wrapping awkwardly in different themes.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/assets/main.css
git commit -m "style(ui): final Concept B visual polishing and textures"
```
