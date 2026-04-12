# DESIGN SPEC: THE CHRONICLER (MOD CHRONICLE REDESIGN)
Date: 2026-04-12
Status: APPROVED (Conceptual)

## 1. VISION
A stylized, "Grand Strategy" themed mod manager that eliminates manual load order fatigue through a Hierarchy of Rules (Groups > Mods). The UI adapts its theme dynamically based on the active game.

## 2. UI ARCHITECTURE (UX FLOW)
### A. The Sidebar (Command Center)
- **Game Selector**: Vertically stacked buttons. 
- **Playset Chooser**: Nested dropdown appeared only for the active game.
- **Global Settings/Launch Path**: Context menu or icon on Game button to trigger "Manual Path Setup".
- **Footer (Fixed)**: 
  - Dynamic "ENTER [GAME]" button.
  - Active Mod Count & System Status (Validation/Cycles).
  - Main Settings Cog (Global preferences).

### B. The Management View (Main)
- **The Chronicler's List**: 1-Column list of "Active" groups and mods.
- **Rules System**: 
  - **Groups**: User-defined clusters (e.g., "Overhauls"). Rule: `Group A loads after Group B`.
  - **Mods**: Rule: `Mod A loads after Mod B` (within or across groups).
  - **Auto-Sort**: Backend resolves the directed acyclic graph (DAG) automatically.
- **Enable/Disable Logic**:
  - Toggling a mod "OFF" in the list moves it to the **Repository**.
  - Toggling a "disabled" mod in the Repository moves it to the **Load Order** (end of list or assigned group).
- **The Repository (Right Pane)**: Compact, searchable list of mods present on disk but not in the current playset.

### C. Mod Interactions
- **Left Click**: Full-screen centered Modal/Popup.
  - Features: Thumbnail gallery, Steam description (HTML), "View in Workshop", and "Unsubscribe".
  - Behavior: Closes on ESC or clicking the darkened backdrop.
- **Right Click**: Context Menu.
  - Actions: "Manage Rules", "View in Workshop", "Unsubscribe" (Workshop only), "View Local Files".

## 3. TECHNICAL ARCHITECTURE (BACKEND CONTRACTS)
### A. Domain Models
- `Playset`: Now represents the *entire* valid load order.
- `Constraint`: Unified graph edge representing `Before/After` for both Mods and Groups.
- `Repository`: Virtual view of `AllMods MINUS ActivePlaysetMods`.

### B. Dynamic Theming
- Frontend receives `ThemeConfig` from backend based on the `ActiveGame`.
- CSS Variables (`--accent`, `--bg-image`, `--font-primary`) are swapped globally.

## 4. CONVERSION PLAN
### Phase 1: Data Migration
1. Rewrite `LauncherLayout` to support Group-level constraints.
2. Adapt Vic3/Legacy playsets to the "EU5 Approach" (Playset = Enabled Mods).

### Phase 2: Component Overhaul
1. Delete `frontend/src/assets/main.css` (Reset styles).
2. Create `MainShell.vue` (Sidebar + Footer + Toolbar).
3. Create `RuleManager.vue` (Group/Mod constraint editor).
4. Create `ModGalleryModal.vue` (Centered popup for details).

### Phase 3: Integration
1. Wire Pinia store to be the exclusive source of truth for `activeGame` and `activePlayset`.
2. Connect `slog` (backend) to a "Developer Console" or "Status Logs" in launcher settings.
