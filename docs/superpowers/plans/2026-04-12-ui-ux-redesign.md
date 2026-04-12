# UI/UX Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the current three-column utility layout into a premium, game-themed command center with role-based zoning (Command Rail, Workspace, Inspector).

**Architecture:** Use a persistent App Shell with defined regions. Introduce a `WorkspaceMode` state (Load Order, Discover, Rules) to swap the center area. Implement a theme token system that maps CSS variables to game-specific visual identities.

**Tech Stack:** Vue 3 (Composition API), Pinia (State Management), CSS Variables (Theming), Wails (Backend bridge).

---

### Task 1: Define Theme Token System and Component Exports

**Files:**
- Create: `frontend/src/styles/tokens.css`
- Modify: `frontend/src/assets/main.css`
- Modify: `frontend/src/types.ts`

- [ ] **Step 1: Create central theme token structure**

```css
/* frontend/src/styles/tokens.css */
:root {
  --rail-bg: var(--bg-sidebar);
  --workspace-bg: var(--bg-body);
  --inspector-bg: var(--bg-panel);
  --card-bg: var(--bg-elevated);
  --card-border: var(--border);
  --accent-primary: var(--accent);
}
```

- [ ] **Step 2: Update types for Workspace Modes**

```typescript
// frontend/src/types.ts
export type WorkspaceMode = 'load-order' | 'discover' | 'rules';
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/styles/tokens.css frontend/src/assets/main.css frontend/src/types.ts
git commit -m "style: define theme token system and workspace modes"
```

---

### Task 2: Refactor App Shell to Role-Based Regions

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Update template structure to match spec (Rail, Workspace, Inspector)**

```html
<template>
  <div class="app-shell" :class="`theme-${activeGameStore.activeGameID}`">
    <nav class="command-rail">
      <!-- Rail content: identity, launch, switching -->
    </nav>
    <main class="workspace-center">
      <!-- Workspace content: load order, discover, rules -->
    </main>
    <aside class="inspector-right">
      <!-- Inspector content -->
    </aside>
  </div>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/App.vue
git commit -m "refactor: implement role-based shell regions"
```

---

### Task 3: Evolve Load Order Panel into Command Surface

**Files:**
- Modify: `frontend/src/components/LoadOrderPanel.vue`

- [ ] **Step 1: Update group headers to control bars**

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/LoadOrderPanel.vue
git commit -m "ui: upgrade load order panel to command surface"
```

---

### Task 4: Promote Mod Repository to Discover Mode

**Files:**
- Modify: `frontend/src/components/ModRepository.vue`

- [ ] **Step 1: Update selection behavior to drive inspector**

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ModRepository.vue
git commit -m "feat: promote mod repository to discover workspace mode"
```

---

### Task 5: Generalize Mod Details into Persistent Inspector

**Files:**
- Modify: `frontend/src/components/ModDetailsPanel.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Remove modal wrap and anchor inspector in shell**

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ModDetailsPanel.vue frontend/src/App.vue
git commit -m "refactor: turn mod details into persistent contextual inspector"
```

---

### Task 6: Final Thematic Polish and Verification

**Files:**
- Modify: `frontend/src/assets/main.css`

- [ ] **Step 1: Apply game-specific visual polish (glows, textures)**

- [ ] **Step 2: Commit**

```bash
git commit -m "style: final per-game thematic polish"
```
