# MOD CHRONICLE REDESIGN: IMPLEMENTATION PLAN

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Massive UI/UX redesign featuring a "Chronicler" (Grand Strategy) theme, rule-based mod groups, and separate Load Order / Repository sections.

**Architecture:**
- **Backend**: Update domain models (`Constraint`, `LoadOrder`) and `GraphSorter` to handle group-level and cross-group rules.
- **Frontend**: Zero-out existing styles and replace with a Pinia-driven, dynamic CSS theme system.
- **Workflow**: Mod enable/disable is a transfer of membership between a playset and the repository.

**Tech Stack:** Wails v2, Go, Vue 3 (Composition API), Pinia (Source of Truth).

---

### Task 1: Backend Domain Update (Constraints & Playsets)

**Files:**
- Modify: `internal/domain/constraint.go`
- Modify: `internal/domain/loadorder.go`
- Modify: `internal/domain/mod.go`

- [ ] **Step 1: Update `Constraint` struct to support `TargetType` (Mod or Group)**
```go
// internal/domain/constraint.go
type TargetType string
const (
	TargetMod   TargetType = "mod"
	TargetGroup TargetType = "group"
)

type Constraint struct {
	Type       ConstraintType `json:"type"`
	FromID     string         `json:"fromId"`
	FromType   TargetType     `json:"fromType"`
	ToID       string         `json:"toId"`
	ToType     TargetType     `json:"toType"`
}

// Ensure interface compatibility for graph building (Update BuildAdjacency later)
```

- [ ] **Step 2: Update `LoadOrder` to focus on "ActiveModIDs"**
```go
// internal/domain/loadorder.go
type LoadOrder struct {
	GameID       GameID       `json:"gameId"`
	PlaysetIdx   PlaysetIndex `json:"playsetIdx"`
	ActiveModIDs []string     `json:"activeModIds"` // ONLY enabled mods
}
```

- [ ] **Step 3: Run `go build ./...` to verify no domain breakage**
Run: `go build ./internal/domain/...`
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add internal/domain/*.go
git commit -m "feat(backend): update domain models for rule-based sorting"
```

---

### Task 2: Backend Graph Sorter Refactor

**Files:**
- Modify: `internal/launcher/graph.go`
- Modify: `internal/repo/layout_repo.go`

- [ ] **Step 1: Update `LauncherLayoutData` to store group-level constraints**
```go
// internal/repo/layout_repo.go
type LauncherCategoryData struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	ModIDs      []string            `json:"modIds"`
	Rule        string              `json:"rule,omitempty"` // E.g., "LOAD AFTER grp_basic"
}
```

- [ ] **Step 2: Update `GraphSorter` to handle multi-level topological sort**
Modify `internal/launcher/graph.go` to first build a graph of categories, then sort within each, resolving Mod-to-Mod constraints across categories.

- [ ] **Step 3: Commit**
```bash
git add internal/launcher/graph.go internal/repo/layout_repo.go
git commit -m "feat(backend): refactor graph sorter for hierarchical constraints"
```

---

### Task 3: Frontend Theme Reset & Store Overhaul

**Files:**
- Modify: `frontend/src/assets/main.css` (Complete rewrite)
- Modify: `frontend/src/App.vue`
- Create: `frontend/src/stores/activeGame.ts`

- [ ] **Step 1: Replace `main.css` with a CSS Variable-driven theme system**
```css
/* frontend/src/assets/main.css */
:root {
  --bg-body: #1a1814;
  --bg-sidebar: #24211b;
  --accent: #b9935a;
  --text: #d4cfc3;
  --border: #3d382e;
}

body.theme-caesar {
  --bg-body: #1a1814; /* Game-specific parchment */
  --accent: #b9935a; 
}

body.theme-victoria {
  --bg-body: #181d24; /* Steampunk blue/grey */
  --accent: #d8b48b;
}
```

- [ ] **Step 2: Create Pinia store for `ActiveGame` and `Theme`**
```typescript
// frontend/src/stores/activeGame.ts
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useActiveGameStore = defineStore('activeGame', () => {
  const activeGameID = ref('eu5')
  const theme = ref('caesar')

  watch(activeGameID, (id) => {
    theme.value = id === 'eu5' ? 'caesar' : 'victoria'
    document.body.className = `theme-${theme.value}`
  }, { immediate: true })

  return { activeGameID, theme }
})
```

- [ ] **Step 3: Commit**
```bash
git add frontend/src/assets/main.css frontend/src/stores/activeGame.ts
git commit -m "feat(ui): implement dynamic game-driven theme system"
```

---

### Task 4: Shell Rebuild (Sidebar & Footer)

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/LaunchButton.vue`

- [ ] **Step 1: Rewrite `App.vue` layout to match concept B (Sidebar + Main + Footer)**
Replace grid layout with:
- `aside.sidebar` (Fixed 280px)
- `main.content-area` (Flex: 1)
- `footer.sidebar-footer` (Fixed at bottom of aside)

- [ ] **Step 2: Move `LaunchButton` and `Stats` into fixed Sidebar Footer**

- [ ] **Step 3: Run `tsc --noEmit`**
Run: `npm run type-check`
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add frontend/src/App.vue frontend/src/components/LaunchButton.vue
git commit -m "feat(ui): rebuild shell layout with fixed sidebar footer"
```

---

### Task 5: Mod Chronicle List & Repository View

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Create: `frontend/src/components/ModRepository.vue`

- [ ] **Step 1: Update `LoadOrderPanel` to prioritize Group-based list**
- [ ] **Step 2: Implement `ModRepository.vue` as a right-pane search list**
- [ ] **Step 3: Implement Enable/Disable as a move operation between Store lists**

- [ ] **Step 4: Commit**
```bash
git add frontend/src/components/LoadOrderPanel.vue frontend/src/components/ModRepository.vue
git commit -m "feat(ui): separate load order from mod repository"
```

---

### Task 6: Interaction Modals (Detail & Context)

**Files:**
- Modify: `frontend/src/components/ModDetailsPanel.vue`
- Modify: `frontend/src/components/ui/BaseModal.vue`

- [ ] **Step 1: Redesign `ModDetailsPanel` as a centered popup**
- [ ] **Step 2: Add context menu actions (Add Rule, View in Workshop)**
- [ ] **Step 3: Commit**
```bash
git add frontend/src/components/ModDetailsPanel.vue
git commit -m "feat(ui): convert mod details to centered popup"
```
