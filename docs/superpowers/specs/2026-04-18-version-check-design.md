# Version Check Design Spec

## 1. Overview
[Feature] Mod compatibility version check.
[Action] Add version detection per game, parse mod supported version, calculate compatibility, display warnings/badges.
[Reason] Allow users to identify outdated or incompatible mods.

## 2. Backend Data Changes
- Add `GameVersionOverride` to `repo.GameSettingsData`.
- Update `mods.Descriptor` to extract `supported_version`.
- Update `mods.Mod` to include `SupportedVersion` and `IsCompatible`.
- Add `DetectVersion() (string, error)` to `game.Adapter` interface.
- Add `GetGameVersion()` to backend API, or attach to game context.
- Update `LoadMods` to compute `IsCompatible` for each `mods.Mod` using prefix matching.

## 3. Version Detection Logic (Game Specific)
Extract final semantic version segment (e.g. `1.18.2` from `titus/release/1.18.2` or `release_1.37.5`).
Fallback logic per adapter:
- **CK3**: `GameVersionOverride` -> `titus_branch.txt` -> `clausewitz_branch.txt` -> "unknown"
- **EU4**: `GameVersionOverride` -> `eu4branch.txt` -> `clausewitz_branch.txt` -> "unknown"
- **EU5**: `GameVersionOverride` -> `caesar_branch.txt` -> `clausewitz_branch.txt` -> "unknown"
- **HOI4**: `GameVersionOverride` -> "unknown" (files contain "None")
- **Vic3**: `GameVersionOverride` -> `caligula_branch.txt` -> `clausewitz_branch.txt` -> "unknown"

## 4. Compatibility Matching
- [Action] Implement `IsVersionCompatible(gameVersion, supportedVersion)`.
- [Rule] Perform prefix matching. Example: `supported_version="1.37.*"`. Remove `*` -> `1.37.`. Check if `gameVersion` starts with `1.37.`.
- [Rule] If exact match (e.g. `1.0.0` == `1.0.0`), return `true`.
- [Rule] If `gameVersion` is "unknown", `IsCompatible` is `false` (or handled gracefully).

## 5. Frontend UI
- **ModListPanel**: Add non-intrusive soft warning banner at top if active game version is "unknown".
- **GameSettingsModal**: Add text input for "Game Version (Override)" to allow manual setting (critical for HOI4/EU5).
- **ModCard**:
  - Add icon indicator (yellow warning sign/exclamation mark) when `IsCompatible` == false. (Default ON).
  - Add card coloring logic for incompatible mods. (Default OFF, via code variable `CompatibilityCardColoringEnabled = false`).
