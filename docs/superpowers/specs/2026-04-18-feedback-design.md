# Feedback Improvements Design Spec

## 1. Overview
[Feature] UX and UI improvements based on user feedback.
[Action] Fix scrolling, update compatibility visual indicators, change list filtering behavior, add auto-detect button, and show correct version metrics.
[Reason] To address user feedback and improve launcher usability and logic.

## 2. Backend Data Changes
- **Compatibility Logic**: 
  - Update `IsVersionCompatible` in `internal/service/mods_service.go`.
  - If `supportedVersion == ""`, treat it as `ANY` and return `true`.

## 3. Frontend UI

### 3.1 Settings
- **GameSettingsModal.vue**:
  - Next to "Game Version (Override)" input, add a small "Auto Detect" button.
  - Clicking "Auto Detect" clears the `gameVersionOverride` input/state so the backend reverts to auto-detection on next scan.

### 3.2 Scrolling Fix
- **LoadOrderPanel.vue & ModListPanel.vue**:
  - The recent layout changes broke vertical scrolling inside flex containers.
  - Add `min-height: 0` (or `min-h-0` class) to flex children (`.view-content`, `.group-container`, `.list-body`) that need to scroll, preventing them from expanding beyond the flex container's height.

### 3.3 Load Order Warning Signs
- **LoadOrderItem.vue**:
  - Add the same `⚠️` yellow icon indicator for incompatible mods (`!mod.isCompatible`) next to the version tag.

### 3.4 Mod Lists Filtering
- **Load Order (LoadOrderPanel.vue / store)**:
  - Should *only* display enabled mods.
- **Mod Repository (ModListPanel.vue / store)**:
  - Should *only* display disabled mods.
  - (Update the computed lists in `ModListPanel` / Pinia store so that enabling a mod moves it to the load order view, and disabling moves it back to the repository).

### 3.5 Version Display
- **ModCard.vue & LoadOrderItem.vue**:
  - Replace the display of `mod.version` with `mod.supportedVersion`.
  - Add fallback logic (e.g. if `supportedVersion` is empty, show "ANY").
- **ModDetailsPanel.vue**:
  - Display the actual `mod.version` inside the mod details popup (e.g., under the mod title or in a stats row).