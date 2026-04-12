# MOD CHRONICLE UI REFINEMENT (CONCEPT B ALIGNMENT)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the production UI with Concept B design choices, fixing regressions in repository visibility, theme alignment, and group management.

**Architecture:**
- **Store-driven UI**: Use `activeGameStore` to control visibility and layout.
- **Component isolation**: Move group logic into a `GroupManagerModal`.
- **CSS Variable alignment**: Refine `main.css` to match Concept B exact colors and spacing.

**Tech Stack:** Vue 3, Pinia, CSS Variables, Wails.

---

### Task 1: Fix Repository Visibility & Layout

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/ModRepository.vue`

- [ ] **Step 1: Fix the layout grid in App.vue to prevent repository collapse**
```vue
/* Change main-split to use flex or fixed columns to ensure stability */
<style scoped>
.main-split {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.content-area-main {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
}
.repository-pane {
  width: 320px;
  border-left: 1px solid var(--border);
  background: rgba(0,0,0,0.15);
}
</style>
```

- [ ] **Step 2: Ensure ModRepository doesn't hide itself on load**
Verify `ModRepository.vue` doesn't have internal logic that sets its own visibility to false incorrectly.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/App.vue frontend/src/components/ModRepository.vue
git commit -m "fix(ui): restore stable side-by-side repository layout"
```

---

### Task 2: Sidebar & Playset Placement

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Move playset selector inside the active game card**
```vue
<!-- Inside games-list loop -->
<button ...>
  <span class="game-icon">⚔️</span>
  <span class="game-name">{{ game.name }}</span>
</button>
<!-- Move playset-select here IF active -->
<div v-if="activeGameStore.activeGameID === game.id" class="playset-selector">
  <select class="playset-dropdown" :value="launcherActivePlaysetIndex" @change="...">...</select>
</div>
```

- [ ] **Step 2: Styles for nested selector**
Match concept B: `margin-top: 10px; margin-left: 32px;`.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/App.vue
git commit -m "feat(ui): nest playset selector under active game"
```

---

### Task 3: Toggles & Context Menu Clarity

**Files:**
- Modify: `frontend/src/assets/main.css`
- Modify: `frontend/src/components/LoadOrderPanel.vue`
- Modify: `frontend/src/components/ui/ContextMenu.vue`

- [ ] **Step 1: Implement Toggle Switch in LoadOrderPanel**
Replace the `×` button with a CSS-based toggle `on/off`.

- [ ] **Step 2: Fix Context Menu background opacity**
Update `frontend/src/components/ui/ContextMenu.vue` to ensure background is solid `var(--bg-panel)` or slightly translucent but readable.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/assets/main.css frontend/src/components/LoadOrderPanel.vue
git commit -m "feat(ui): replace remove buttons with toggles and fix context menu"
```

---

### Task 4: Separate Group Management

**Files:**
- Create: `frontend/src/components/modals/GroupManagerModal.vue`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/components/LoadOrderPanel.vue`

- [ ] **Step 1: Create GroupManagerModal**
Move the `category-creator` and category deletion logic into this modal.

- [ ] **Step 2: Add "Manage Groups" button to toolbar**
Add button next to "Autosort" in `App.vue` or `LoadOrderPanel.vue` header.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/components/modals/GroupManagerModal.vue frontend/src/App.vue
git commit -m "feat(ui): implement dedicated group management modal"
```

---

### Task 5: Theme & Typography Polishing

**Files:**
- Modify: `frontend/src/assets/main.css`

- [ ] **Step 1: Fix heading wrapping issues**
Adjust `font-size` and `letter-spacing` for "Load Order Chronicle" title to ensure single-line consistency across themes.

- [ ] **Step 2: Align with Concept B background texture**
Add the subtle SVG pattern from `redesign_concept_B.html` to `body`.

- [ ] **Step 3: Commit**
```bash
git add frontend/src/assets/main.css
git commit -m "style(ui): final typography and texture alignment with concept B"
```
