# Multi-Game Redesign Phase 3: Frontend Adaptation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update the Vue frontend to support game switching and dynamic theming.

**Architecture:** Add a Sidebar for game selection and update Pinia stores to handle game-specific state refreshing.

**Tech Stack:** Vue 3, TypeScript, Tailwind/CSS.

---

### Task 1: Game Selection Sidebar

**Files:**
- Create: `frontend/src/components/Sidebar.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Create Game Sidebar**

```vue
<template>
  <aside class="w-16 bg-gray-900 flex flex-col items-center py-4 space-y-4">
    <div v-for="game in games" :key="game.id" @click="selectGame(game.id)">
      <img :src="game.icon" :class="{ 'ring-2 ring-blue-500': activeGame === game.id }" />
    </div>
  </aside>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/Sidebar.vue
git commit -m "feat: add game selection sidebar"
```

---

### Task 2: Dynamic Theming

**Files:**
- Modify: `frontend/src/style.css`
- Modify: `frontend/src/stores/settings.ts`

- [ ] **Step 1: Define CSS variables for themes**

```css
:root { --accent: #ffcc00; }
.theme-hoi4 { --accent: #990000; }
.theme-ck3 { --accent: #4a90e2; }
```

- [ ] **Step 2: Apply theme class to App root**

```typescript
const themeClass = computed(() => `theme-${settingsStore.activeGameID}`)
```

- [ ] **Step 3: Commit**

```bash
git commit -m "feat: implement game-specific dynamic theming"
```
