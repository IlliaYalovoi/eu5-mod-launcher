# Version Check Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement mod compatibility version checking per game with prefix matching and manual override support.

**Architecture:** Add `SupportedVersion` to mods and `GameVersionOverride` to settings. Implement `DetectVersion` on `game.Adapter` that reads game specific branch files and extracts version. Compute `IsCompatible` on backend during mod loading and expose to UI. Update UI to show warnings.

**Tech Stack:** Go, Vue 3, TypeScript

---

### Task 1: Update Data Models

**Files:**
- Modify: `internal/repo/settings_repo.go`
- Modify: `internal/mods/descriptor.go`
- Modify: `internal/mods/mod.go`
- Modify: `internal/game/adapter.go`

- [ ] **Step 1: Add GameVersionOverride to settings**
Modify `internal/repo/settings_repo.go` to add `GameVersionOverride` to `GameSettingsData`:
```go
type GameSettingsData struct {
	ModsDir             string `json:"modsDir,omitempty"`
	GameExe             string `json:"gameExe,omitempty"`
	GameVersionOverride string `json:"gameVersionOverride,omitempty"`
}
```

- [ ] **Step 2: Add SupportedVersion to Descriptor**
Modify `internal/mods/descriptor.go` to add `SupportedVersion` to `Descriptor`:
```go
type Descriptor struct {
	Name             string
	Version          string
	SupportedVersion string
	Description      string
	Tags             []string
}
```
Update `parseJSONDescriptor` to extract `supported_version`:
```go
	supportedVersion := extractJSONString(payload, "supported_version")
	// ... add to returned Descriptor
```
Update `parseTextDescriptor` switch statement to extract `supported_version`:
```go
		case "supported_version":
			parsed.SupportedVersion = value
```

- [ ] **Step 3: Add Version fields to Mod**
Modify `internal/mods/mod.go` to add `SupportedVersion` and `IsCompatible`:
```go
type Mod struct {
	ID               string
	Name             string
	Version          string
	SupportedVersion string
	Tags             []string
	Description      string
	ThumbnailPath    string
	DirPath          string
	Enabled          bool
	IsCompatible     bool
}
```

- [ ] **Step 4: Add DetectVersion to Adapter**
Modify `internal/game/adapter.go` to add `DetectVersion` to the `Adapter` interface:
```go
type Adapter interface {
	ID() string
	DetectInstances() ([]Instance, error)
	LoadMods(inst Instance) ([]ModEntry, error)
	LoadPlaysets(inst Instance) ([]Playset, error)
	SavePlayset(inst Instance, p Playset) error
	DetectVersion(inst Instance, override string) (string, error)
}
```

- [ ] **Step 5: Run existing tests to ensure no compilation errors**
Run: `go build ./...`
Run: `go test ./...`

- [ ] **Step 6: Commit**
```bash
git add internal/repo/settings_repo.go internal/mods/descriptor.go internal/mods/mod.go internal/game/adapter.go
git commit -m "feat(backend): add version check data models"
```

### Task 2: Implement DetectVersion in Adapters

**Files:**
- Modify: `internal/adapters/eu5/adapter.go`
- Modify: `internal/adapters/legacy/sqlite.go`

- [ ] **Step 1: Implement DetectVersion for EU5**
Modify `internal/adapters/eu5/adapter.go`. Add `DetectVersion` method:
```go
import (
	"os"
	"path/filepath"
	"strings"
)

func (a *Adapter) DetectVersion(inst game.Instance, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	
	// Check caesar_branch.txt then clausewitz_branch.txt
	for _, filename := range []string{"caesar_branch.txt", "clausewitz_branch.txt"} {
		content, err := os.ReadFile(filepath.Join(inst.InstallPath, filename))
		if err == nil {
			return extractVersion(string(content)), nil
		}
	}
	return "unknown", nil
}

func extractVersion(content string) string {
	content = strings.TrimSpace(content)
	parts := strings.Split(content, "/")
	if len(parts) > 0 {
		content = parts[len(parts)-1]
	}
	parts = strings.Split(content, "_")
	if len(parts) > 0 {
		content = parts[len(parts)-1]
	}
	return content
}
```

- [ ] **Step 2: Implement DetectVersion for Legacy games**
Modify `internal/adapters/legacy/sqlite.go`. Add `DetectVersion` method (add similar imports if needed):
```go
import (
	"strings"
)

func (s *SqliteAdapter) DetectVersion(inst game.Instance, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	
	var primaryFile string
	switch s.id {
	case "ck3":
		primaryFile = "titus_branch.txt"
	case "eu4":
		primaryFile = "eu4branch.txt"
	case "victoria3":
		primaryFile = "caligula_branch.txt"
	case "hoi4":
		// HOI4 files contain "None", handled below or just override
		primaryFile = "ho4branch.txt"
	default:
		primaryFile = s.id + "_branch.txt"
	}

	for _, filename := range []string{primaryFile, "clausewitz_branch.txt"} {
		content, err := os.ReadFile(filepath.Join(inst.InstallPath, filename))
		if err == nil {
			str := strings.TrimSpace(string(content))
			if str != "None" {
				return extractVersion(str), nil
			}
		}
	}
	return "unknown", nil
}

func extractVersion(content string) string {
	content = strings.TrimSpace(content)
	parts := strings.Split(content, "/")
	if len(parts) > 0 {
		content = parts[len(parts)-1]
	}
	parts = strings.Split(content, "_")
	if len(parts) > 0 {
		content = parts[len(parts)-1]
	}
	return content
}
```
*(Note: Duplicate `extractVersion` can be placed in `internal/utils/fs.go` or kept local. If placing in `utils`, modify `internal/utils/fs.go` and use `utils.ExtractVersion` in both places)*

- [ ] **Step 3: Run existing tests**
Run: `go build ./...`
Run: `go test ./...`

- [ ] **Step 4: Commit**
```bash
git add internal/adapters/eu5/adapter.go internal/adapters/legacy/sqlite.go
git commit -m "feat(backend): implement game version detection per adapter"
```

### Task 3: Compute Compatibility during Mod Loading

**Files:**
- Modify: `internal/service/mods_service.go`
- Modify: `frontend/src/types.ts`
- Modify: `frontend/src/stores/settings.ts`

- [ ] **Step 1: Compute IsCompatible in Mods Service**
Identify where `mods.ParseDescriptor` is called (likely `scanner.go` or `mods_service.go`). Add a helper to match versions in `internal/mods/scanner.go` or `mods_service.go`:
```go
func IsVersionCompatible(gameVersion, supportedVersion string) bool {
	if gameVersion == "unknown" || supportedVersion == "" {
		return false
	}
	if gameVersion == supportedVersion {
		return true
	}
	prefix := strings.ReplaceAll(supportedVersion, "*", "")
	return strings.HasPrefix(gameVersion, prefix)
}
```
In `internal/service/mods_service.go` (or wherever `Mod` structs are built), call `gameAdapter.DetectVersion` to get current game version, and set `mod.SupportedVersion = desc.SupportedVersion` and `mod.IsCompatible = IsVersionCompatible(gameVersion, desc.SupportedVersion)`. *Ensure you expose GameVersion to frontend via settings or a dedicated endpoint.*

- [ ] **Step 2: Update Typescript Types**
Update `frontend/src/types.ts`:
```typescript
export interface Mod {
	id: string;
	name: string;
	version: string;
	supportedVersion: string;
	tags: string[];
	description: string;
	thumbnailPath: string;
	dirPath: string;
	enabled: boolean;
	isCompatible: boolean;
}

export interface GameSettingsData {
	modsDir?: string;
	gameExe?: string;
	gameVersionOverride?: string;
}
```
Update `frontend/src/stores/settings.ts` to include `gameVersionOverride`.

- [ ] **Step 3: Compile and Test**
Run: `go build ./...`
Run: `tsc --noEmit`

- [ ] **Step 4: Commit**
```bash
git add internal/service/mods_service.go frontend/src/types.ts frontend/src/stores/settings.ts
git commit -m "feat: compute mod compatibility on backend"
```

### Task 4: Frontend UI - Settings & Soft Warning

**Files:**
- Modify: `frontend/src/components/GameSettingsModal.vue`
- Modify: `frontend/src/components/ModListPanel.vue`

- [ ] **Step 1: Add Game Version Override Input**
Modify `frontend/src/components/GameSettingsModal.vue`. Add an input field for `gameVersionOverride`:
```vue
<div class="mt-4">
  <label class="block text-sm font-medium text-surface-400 mb-1">Game Version (Override)</label>
  <input 
    type="text" 
    v-model="settingsStore.getGameSettings(activeGame).gameVersionOverride"
    class="w-full bg-surface-800 border border-surface-700 rounded-md px-3 py-2 text-surface-200 focus:outline-none focus:border-primary-500"
    placeholder="e.g. 1.37.5"
  />
</div>
```

- [ ] **Step 2: Add Soft Warning for Unknown Version**
Modify `frontend/src/components/ModListPanel.vue` to show a warning banner if the active game's detected version is "unknown" (or if no override and detection fails):
```vue
<template>
  <div class="flex flex-col h-full">
    <!-- Add Banner Here -->
    <div v-if="detectedGameVersion === 'unknown'" class="bg-yellow-900/30 text-yellow-400 p-2 text-center text-sm border-b border-yellow-700/50">
      Unknown game version - please set it manually in settings for correct mod compatibility check.
    </div>
    <!-- Existing content -->
```
*(Ensure `detectedGameVersion` is fetched/computed from the backend).*

- [ ] **Step 3: Type check**
Run: `tsc --noEmit`

- [ ] **Step 4: Commit**
```bash
git add frontend/src/components/GameSettingsModal.vue frontend/src/components/ModListPanel.vue
git commit -m "feat(ui): add version override setting and soft warning"
```

### Task 5: Frontend UI - ModCard Indicators

**Files:**
- Modify: `frontend/src/components/ModCard.vue`

- [ ] **Step 1: Add Warning Icon and Card Coloring**
Modify `frontend/src/components/ModCard.vue`:
```vue
<script setup lang="ts">
const CompatibilityCardColoringEnabled = false; // Code toggle
// ... existing props/imports
</script>

<template>
  <div 
    class="relative rounded-lg border transition-colors overflow-hidden"
    :class="[
      !mod.isCompatible && CompatibilityCardColoringEnabled ? 'bg-red-900/20 border-red-700/50' : 'bg-surface-800 border-surface-700 hover:border-surface-600',
      // ... existing classes
    ]"
  >
    <!-- Somewhere inside the card header or near version -->
    <div class="flex items-center gap-2">
      <span class="text-xs text-surface-400">v{{ mod.version }}</span>
      <span v-if="!mod.isCompatible" class="text-yellow-500" title="Incompatible game version">
        ⚠️
      </span>
    </div>
    <!-- ... -->
  </div>
</template>
```

- [ ] **Step 2: Type check**
Run: `tsc --noEmit`

- [ ] **Step 3: Commit**
```bash
git add frontend/src/components/ModCard.vue
git commit -m "feat(ui): add mod compatibility indicators on cards"
```
