# UI/UX Redesign Design

## Goal
Transform PDX Mod Organizer from generic utility-style three-column launcher into premium, game-native mod command center with shared interaction model and per-game visual themes.

## Product Direction
The app should feel like a themed command console for grand strategy mod curation rather than a neutral admin tool. It must preserve core jobs already supported by the app: selecting a game, managing load order, browsing available mods, inspecting details, managing rules, and launching the game. The redesign should improve visual authority, atmosphere, and hierarchy without fragmenting the core mental model across games.

## Design Strategy
Use a hybrid model:
- one shared interaction foundation
- one shared shell architecture
- one shared component contract
- per-game theme packs that change visual identity and selected layout modules

This approach balances flavor and maintainability. Each supported game should feel distinct, but users should not need to relearn the app when switching titles.

## Core Shell Architecture
Replace the current equal-weight three-column structure with a role-based shell:

### Left Command Rail
Purpose:
- game identity
- game switching
- launch action
- playset status
- global settings and utilities
- persistent high-authority actions

Characteristics:
- stable across workspace modes
- visually strong branded anchor
- contains current game title, iconography, and launch summary

### Center Workspace
Purpose:
- primary task surface
- mode-dependent working area

Primary modes:
- Load Order
- Discover
- Rules / Conflicts
- future Collections / Playsets if needed

This region should hold the app’s dominant interaction surface. It should be visually flexible enough to support list-heavy management, filtered discovery, and future rule tooling.

### Right Inspector
Purpose:
- contextual detail surface
- single source of truth for currently selected entity

Supported inspector states:
- selected mod
- selected group/category
- selected rule/conflict
- empty selection overview

The inspector should remain persistent and context-driven rather than acting as a narrow one-off details panel.

## Interaction Model
### Shared Principles
- one primary action per area
- selection drives inspector
- hover previews, click commits focus
- drag and drop remains supported where useful
- key workflows must remain understandable without relying only on drag and drop
- dangerous actions are visually separated from primary flows

### Load Order Mode
Load Order mode becomes a curated command surface rather than a raw list.

Requirements:
- grouped stack remains central
- group headers act as control bars
- groups display count, state, and quick actions
- mods display only key metadata by default
- secondary metadata moves into badges or inspector
- rule/conflict severity appears inline and visible
- drag affordances must feel deliberate and premium

### Discover Mode
Discover mode promotes repository browsing into first-class workflow.

Requirements:
- repository becomes primary center workspace view
- search expands into richer filter and sort controls
- item selection updates inspector
- add/remove actions are explicit and visually clear
- empty states provide guidance and next steps

### Rules / Conflicts Mode
Rules need a dedicated workflow surface rather than being hidden behind isolated actions.

Requirements:
- expose constraints, ordering rules, and conflict states clearly
- make unresolved issues scannable
- use severity and category markers
- allow contextual drill-down through inspector

### Launch Flow
Launch should feel authoritative and ceremonial without becoming slow.

Requirements:
- left rail surfaces active game, playset, enabled mod count, and warning summary
- launch action sits in stable prominent zone
- unresolved warnings can be reviewed before launch

## Visual Language
Premium feel must come from a coherent system, not isolated decoration.

### Shared Rules
- establish strong depth hierarchy: world background, shell, panels, interactive cards
- rely less on repetitive borders and more on contrast, framing, material, and shadow
- use larger, more deliberate title moments
- reduce repeated small labels and generic utility text
- keep semantic states visually distinct: enabled, disabled, warning, conflict, missing dependency

### Shared Token Contract
Each theme pack must define:
- background treatment
- panel material
- border or chrome treatment
- display typography
- UI typography
- accent palette
- semantic state palette
- icon style family
- spacing density
- motion profile
- decorative asset slots

### Theme Directions
#### Europa Universalis IV / V
Imperial atlas language:
- gilded lines
- cartographic textures
- heraldic framing
- ceremonial controls
- serif-forward display typography

#### Hearts of Iron IV
Command table language:
- military dossier panels
- map-grid overlays
- clipped geometry
- stamped metadata treatments
- utilitarian but premium hierarchy

#### Victoria 3
Industrial ledger language:
- engraved separators
- brass and machinery cues
- ledger/table motifs
- catalog rhythm
- early industrial visual authority

#### Stellaris
Sci-fi command bridge language:
- holographic layers
- glow-edged panels
- scan-line or signal motifs
- soft neon semantic states
- controlled motion feedback

### Readability Guardrail
Themes must not reduce usability. Decorative texture, noise, glow, or thematic ornament should be reduced or removed wherever it weakens legibility or scanning.

## Concrete Redesign Targets In Current App
### App Shell
Current shell in `frontend/src/App.vue` should shift from fixed utility layout to themed command shell with stronger zoning and contextual titles.

### Load Order Panel
Current load order surface in `frontend/src/components/LoadOrderPanel.vue` already has strong functional primitives, but lacks premium hierarchy and visible system state. It should evolve into a task-focused command surface with stronger group bars, clearer selection states, inline rule markers, and improved drag affordances.

### Mod Repository
Current repository in `frontend/src/components/ModRepository.vue` behaves like a side bin. It should become a discovery workspace with richer cards, filters, sorting, and stronger ties to selection and inspector behavior.

### Mod Details Panel
Current details panel in `frontend/src/components/ModDetailsPanel.vue` is structurally promising. It should be generalized into persistent inspector supporting mod, group, rule, and overview states with stronger hierarchy, artwork, metadata blocks, and contextual actions.

### Controls
Current generic toggles, dashed containers, and raw search field treatments should be replaced by themed control primitives that still share interaction semantics across games.

## Information Hierarchy
The new hierarchy should be:
1. current game and active mode
2. active workspace content
3. selected entity details in inspector
4. secondary metadata and system status
5. global utilities and settings

This reduces current equal-weight competition between columns and gives the app a clear focal rhythm.

## Premium Overview Moments
The redesign should include at least one contextual overview state that surfaces:
- active game identity
- active playset summary
- enabled mod count
- unresolved conflicts or rule issues
- repository scale or source summary
- recent system state where useful

This gives the app a stronger sense of place and progress even before a mod is selected.

## Component Direction
Introduce or restyle shared components around consistent semantic roles:
- command rail sections
- workspace mode switcher
- inspector panels
- themed action buttons
- state badges
- filter chips
- summary cards
- selectable mod rows/cards
- empty and loading states

Theme packs may change chrome and expression, but not the core behavioral contract of these components.

## Motion Direction
Motion should be restrained and theme-aware.

Guidelines:
- use subtle state transitions for hover, selection, expansion, and loading
- avoid decorative animation that distracts from list scanning
- let theme packs tune easing and visual response
- keep drag feedback crisp and readable

## Accessibility And UX Guardrails
- preserve strong contrast for text and state indicators
- maintain clear selected, hover, focus, and disabled states
- do not rely on color alone for conflicts or warnings
- keep dense information scannable at desktop sizes
- ensure themes remain readable under long mod names and heavy metadata

## Recommended Delivery Shape
### Phase 1
- define theme token system
- define shell regions and base layout
- refactor current UI into command rail, workspace, and inspector without changing core backend behavior

### Phase 2
- introduce workspace modes: Load Order, Discover, Rules / Conflicts
- upgrade load order and repository interactions
- add overview and empty states

### Phase 3
- add per-game theme packs for EU, HOI4, Victoria 3, and Stellaris
- refine iconography, textures, and motion
- strengthen launch flow presentation

### Phase 4
- optionally add deeper game-specific modules or layout variations once core system proves stable

## Recommendation
Proceed with hybrid themed shell system rather than fully separate game layouts. Shared behavior with selective themed layout modules provides strongest balance of identity, scalability, and UX consistency.

## Out Of Scope For This Design
- backend/domain model redesign
- new gameplay-aware mod logic
- mobile or responsive redesign for small screens
- broad feature expansion beyond current launcher responsibilities
