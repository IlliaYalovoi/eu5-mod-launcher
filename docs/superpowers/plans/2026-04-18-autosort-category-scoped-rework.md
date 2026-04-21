# Autosort Category-Scoped Rework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make autosort produce the same ordering model the UI shows: categories sorted by category constraints, mods sorted only inside their own category, then final `orderedIds` compiled by stacking category blocks.

**Architecture:** Replace global full-list mod sort in `Autosort()` with a two-level pipeline: (1) sort category blocks using category constraints, (2) sort each category’s mod list independently using mod constraints, then compile final order from sorted blocks. Add guardrails on both sides: UI must not allow creating mod constraints across categories; backend must reject cross-category mod constraints (both during add and during autosort validation of existing data).

**Tech Stack:** Go (Wails backend), Vue 3 (`<script setup>`), Pinia, TypeScript.

---

## Implementation constraints (must follow)

- Keep editing on `main` branch (project rule).
- No new tests (project rule).
- Existing tests must pass.
- Required verification before completion:
  - `go build ./...`
  - `go vet ./...` (constructor/service contract changes in this plan)
  - `go test ./...`
  - `cd frontend && npx tsc --noEmit`

---

## File structure and ownership map

### Backend (Go)

- **Modify:** `app.go`
  - Rewrite `Autosort()` to category-scoped sorting pipeline.
  - Add helpers for category block extraction/update and mod→category indexing.
  - Add autosort-time validation for cross-category mod constraints.
  - Pass new category-resolver callback into `ConstraintsService`.
  - Remove obsolete global-sort-to-layout sync helpers no longer needed.

- **Modify:** `internal/service/constraints_service.go`
  - Extend service dependencies with `modCategory` resolver callback.
  - Reject adding `after` constraints between mods in different categories.
  - Keep category↔category behavior unchanged.

### Frontend (Vue/TS)

- **Modify:** `frontend/src/components/ConstraintModal.vue`
  - Restrict mod target picker to mods from same category as current mod.
  - Keep category target mode unchanged.
  - Add explicit helper text so user sees rule in UI.

- **Modify:** `frontend/src/components/CycleErrorPanel.vue`
  - Treat non-cycle autosort errors as generic autosort failures (so cross-category error is not mislabeled as cycle).

---

## Task 1: Add category-scope autosort helpers in backend

**Files:**
- Modify: `app.go`

**Verification:**
- `go build ./...`

- [ ] **Step 1: Add mod→category indexing helpers near autosort helpers**

```go
func buildModCategoryIndex(layout LauncherLayout) map[string]string {
	out := make(map[string]string, len(layout.Ungrouped))
	for _, modID := range layout.Ungrouped {
		if strings.TrimSpace(modID) == "" {
			continue
		}
		out[modID] = defaultUngroupedCategoryID
	}
	for i := range layout.Categories {
		category := layout.Categories[i]
		for _, modID := range category.ModIDs {
			if strings.TrimSpace(modID) == "" {
				continue
			}
			out[modID] = category.ID
		}
	}
	return out
}

func categoryModIDs(layout LauncherLayout, categoryID string) []string {
	if categoryID == defaultUngroupedCategoryID {
		return append([]string(nil), layout.Ungrouped...)
	}
	for i := range layout.Categories {
		if layout.Categories[i].ID == categoryID {
			return append([]string(nil), layout.Categories[i].ModIDs...)
		}
	}
	return []string{}
}

func setCategoryModIDs(layout *LauncherLayout, categoryID string, modIDs []string) {
	if layout == nil {
		return
	}
	if categoryID == defaultUngroupedCategoryID {
		layout.Ungrouped = append([]string(nil), modIDs...)
		return
	}
	for i := range layout.Categories {
		if layout.Categories[i].ID == categoryID {
			layout.Categories[i].ModIDs = append([]string(nil), modIDs...)
			return
		}
	}
}
```

- [ ] **Step 2: Add autosort-time validator for cross-category mod constraints**

```go
var errCrossCategoryModConstraint = errors.New("mod constraint crosses category boundary")

func validateModConstraintsWithinCategories(constraints []graph.Constraint, categoryByMod map[string]string) error {
	for i := range constraints {
		constraint := constraints[i]
		typ := constraint.Type
		if typ == "" {
			typ = graph.ConstraintTypeAfter
		}
		if typ != graph.ConstraintTypeAfter {
			continue
		}
		if isCategoryID(constraint.From) || isCategoryID(constraint.To) {
			continue
		}

		fromCategory, fromOK := categoryByMod[constraint.From]
		toCategory, toOK := categoryByMod[constraint.To]
		if !fromOK || !toOK {
			continue // ignore constraints for mods currently not enabled/present in layout
		}
		if fromCategory != toCategory {
			return fmt.Errorf(
				"%w: %q (%s) -> %q (%s)",
				errCrossCategoryModConstraint,
				constraint.From,
				fromCategory,
				constraint.To,
				toCategory,
			)
		}
	}
	return nil
}
```

- [ ] **Step 3: Build check after helper additions**

Run: `go build ./...`  
Expected: build succeeds.

- [ ] **Step 4: Commit Task 1**

```bash
git add app.go
git commit -m "refactor(autosort): add category-scope helper primitives"
```

---

## Task 2: Rewrite `Autosort()` to category-first, category-local mod sorting

**Files:**
- Modify: `app.go`

**Verification:**
- `go build ./...`

- [ ] **Step 1: Replace global `conGraph.Sort(a.loState.OrderedIDs)` flow in `Autosort()` with category-scoped pipeline**

```go
func (a *App) Autosort() ([]string, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("autosort: %w", err)
	}

	previousOrder := append([]string(nil), a.loState.OrderedIDs...)
	previousLayout := a.launcherLayout

	layout := normalizeLauncherLayout(a.launcherLayout, a.loState.OrderedIDs)
	constraints := a.conGraph.All()
	categoryByMod := buildModCategoryIndex(layout)
	if err := validateModConstraintsWithinCategories(constraints, categoryByMod); err != nil {
		return nil, fmt.Errorf("sort constraints: %w", err)
	}

	categoryByID := indexCategoriesByID(layout.Categories)
	blockIDs := completeCategoryBlockOrder(layout)
	sortGraph := buildCategorySortGraph(blockIDs, constraints)
	order, err := sortCategoryBlocks(blockIDs, sortGraph, categoryByID)
	if err != nil {
		return nil, fmt.Errorf("sort category constraints: %w", err)
	}

	layout.Order = order
	for _, blockID := range order {
		ids := categoryModIDs(layout, blockID)
		if len(ids) == 0 {
			continue
		}
		sorted, err := a.conGraph.Sort(ids)
		if err != nil {
			return nil, fmt.Errorf("sort constraints in %q: %w", blockID, err)
		}
		setCategoryModIDs(&layout, blockID, sorted)
	}

	compiled := compileLauncherLayout(layout)
	if saveErr := a.SetLoadOrder(compiled); saveErr != nil {
		return nil, fmt.Errorf("persist autosorted load order: %w", saveErr)
	}

	a.launcherLayout = layout
	if err := a.layoutRepo.Save(a.layoutPath, toRepoLayout(a.launcherLayout)); err != nil {
		if rollbackErr := a.SetLoadOrder(previousOrder); rollbackErr != nil {
			logging.Errorf("autosort rollback failed after layout save error: %v", rollbackErr)
		}
		a.launcherLayout = previousLayout
		return nil, fmt.Errorf("save launcher layout after autosort: %w", err)
	}

	a.invalidateActiveSnapshot()
	return append([]string(nil), a.loState.OrderedIDs...), nil
}
```

- [ ] **Step 2: Remove obsolete autosort-layout sync functions that relied on global position map**

Delete these now-unused functions from `app.go`:

```go
func (a *App) reorderLauncherLayoutAfterAutosort(sorted []string) (LauncherLayout, error)
func buildIDPositionMap(sorted []string) map[string]int
func sortIDsByPosition(ids []string, position map[string]int, fallback int) []string
func sortLayoutModIDs(layout *LauncherLayout, position map[string]int, sortedCount int)
func sortLayoutModIDsSequential(layout *LauncherLayout, position map[string]int, sortedCount int)
func sortLayoutModIDsConcurrent(layout *LauncherLayout, position map[string]int, sortedCount, workers int)
```

Also delete now-unused autosort constants if no remaining references:

```go
const (
	maxSortWorkers      = 8
	minLayoutForWorkers = 8
)
```

- [ ] **Step 3: Build check after pipeline rewrite**

Run: `go build ./...`  
Expected: build succeeds and no unused symbols remain.

- [ ] **Step 4: Commit Task 2**

```bash
git add app.go
git commit -m "fix(autosort): sort by category blocks then category-local mod constraints"
```

---

## Task 3: Enforce cross-category mod-constraint rule in backend service

**Files:**
- Modify: `internal/service/constraints_service.go`
- Modify: `app.go`

**Verification:**
- `go build ./...`
- `go vet ./...`

- [ ] **Step 1: Extend `ConstraintsService` with mod-category resolver dependency**

```go
type ConstraintsService struct {
	graph       *graph.Graph
	repo        repo.ConstraintsRepository
	path        string
	expand      func(string) []string
	isCategory  func(string) bool
	modCategory func(string) string
}

func NewConstraintsService(
	constraintGraph *graph.Graph,
	constraintsPath string,
	repository repo.ConstraintsRepository,
	expand func(string) []string,
	isCategory func(string) bool,
	modCategory func(string) string,
) *ConstraintsService {
	if expand == nil {
		expand = func(string) []string { return nil }
	}
	if isCategory == nil {
		isCategory = domain.IsCategoryID
	}
	if repository == nil {
		repository = repo.NewFileConstraintsRepository()
	}
	if modCategory == nil {
		modCategory = func(string) string { return "" }
	}
	return &ConstraintsService{
		graph:       constraintGraph,
		path:        constraintsPath,
		repo:        repository,
		expand:      expand,
		isCategory:  isCategory,
		modCategory: modCategory,
	}
}
```

- [ ] **Step 2: Add explicit same-category validation for mod constraints in `AddConstraint`**

```go
var errCrossCategoryModConstraint = errors.New("mod constraints must stay in same category")

func (s *ConstraintsService) validateSameCategory(fromIDs, toIDs []string) error {
	for _, fromID := range fromIDs {
		fromCategory := strings.TrimSpace(s.modCategory(fromID))
		if fromCategory == "" {
			continue
		}
		for _, toID := range toIDs {
			toCategory := strings.TrimSpace(s.modCategory(toID))
			if toCategory == "" {
				continue
			}
			if fromCategory != toCategory {
				return fmt.Errorf("%w: %q (%s) -> %q (%s)", errCrossCategoryModConstraint, fromID, fromCategory, toID, toCategory)
			}
		}
	}
	return nil
}
```

Insert call in `AddConstraint` right after `fromIDs`/`toIDs` resolution and before `applyWithRollback`:

```go
if err := s.validateSameCategory(fromIDs, toIDs); err != nil {
	return fmt.Errorf("add constraint %q -> %q: %w", from, target, err)
}
```

- [ ] **Step 3: Wire `App` category resolver into service initialization**

```go
func (a *App) initConstraintsService() {
	if a.conGraph == nil {
		a.conGraph = graph.New()
	}
	a.conService = service.NewConstraintsService(
		a.conGraph,
		a.constraintsPath,
		a.constraintsRepo,
		a.expandConstraintTarget,
		isCategoryID,
		a.categoryForConstraintMod,
	)
}

func (a *App) categoryForConstraintMod(modID string) string {
	layout := normalizeLauncherLayout(a.launcherLayout, a.loState.OrderedIDs)
	for _, id := range layout.Ungrouped {
		if id == modID {
			return defaultUngroupedCategoryID
		}
	}
	for i := range layout.Categories {
		cat := layout.Categories[i]
		for _, id := range cat.ModIDs {
			if id == modID {
				return cat.ID
			}
		}
	}
	return ""
}
```

- [ ] **Step 4: Run backend verification**

Run:
- `go build ./...`
- `go vet ./...`

Expected: both commands succeed.

- [ ] **Step 5: Commit Task 3**

```bash
git add internal/service/constraints_service.go app.go
git commit -m "fix(constraints): reject cross-category mod constraints"
```

---

## Task 4: Restrict UI constraint picker to same-category mods and clarify autosort error label

**Files:**
- Modify: `frontend/src/components/ConstraintModal.vue`
- Modify: `frontend/src/components/CycleErrorPanel.vue`

**Verification:**
- `cd frontend && npx tsc --noEmit`

- [ ] **Step 1: Build computed mod→category index in `ConstraintModal.vue`**

```ts
const modCategoryByID = computed<Record<string, string>>(() => {
  const out: Record<string, string> = {}
  for (const modID of launcherLayout.value.ungrouped) {
    out[modID] = 'category:ungrouped'
  }
  for (const category of launcherLayout.value.categories) {
    for (const modID of category.modIds) {
      out[modID] = category.id
    }
  }
  return out
})
```

- [ ] **Step 2: Restrict non-category target candidates to same category in `availableMods`**

```ts
const currentCategoryID = computed(() => modCategoryByID.value[props.modID] || '')

// inside non-category branch of availableMods:
for (const mod of allMods.value) {
  if (mod.ID === props.modID) continue
  if (blocked[mod.ID]) continue
  if (!currentCategoryID.value) continue
  if (modCategoryByID.value[mod.ID] !== currentCategoryID.value) continue
  result.push(mod)
}
```

This ensures cross-category mod constraints cannot be created from UI.

- [ ] **Step 3: Add explicit UI hint text in modal**

```vue
<p v-if="!isCategoryTarget" class="hint">Only mods in same category can be constrained.</p>
```

Add style token near existing `.error` style:

```css
.hint {
  color: var(--color-text-muted);
  font-size: 0.85rem;
}
```

- [ ] **Step 4: Make autosort error panel title conditional on real cycle errors**

```ts
const isCycleError = computed(() => {
  const source = (autosortError.value || '').toLowerCase()
  return source.includes('cycle detected') || source.includes('constraint cycle detected')
})
```

```vue
<h3 class="title">{{ isCycleError ? 'Constraint Cycle Detected' : 'Autosort Failed' }}</h3>
```

- [ ] **Step 5: Run frontend verification**

Run: `cd frontend && npx tsc --noEmit`  
Expected: typecheck succeeds.

- [ ] **Step 6: Commit Task 4**

```bash
git add frontend/src/components/ConstraintModal.vue frontend/src/components/CycleErrorPanel.vue
git commit -m "fix(ui): enforce same-category mod constraints and autosort error labeling"
```

---

## Task 5: End-to-end verification and regression check

**Files:**
- No additional code changes expected (verification only).

- [ ] **Step 1: Run required project checks**

Run:
- `go build ./...`
- `go vet ./...`
- `go test ./...`
- `cd frontend && npx tsc --noEmit`

Expected: all commands pass.

- [ ] **Step 2: Manual scenario check (bug reproduction target)**

1. In UI, create/keep category layout with explicit category order.
2. Add mod-level constraints inside one category.
3. Run Autosort.
4. Confirm visual order in each category matches constraints.
5. Confirm compiled load order (`orderedIds`) equals stacked categories in UI order (category 1 mods, then category 2 mods, etc).

- [ ] **Step 3: Manual guardrail check**

1. Open constraints for mod in category A.
2. Verify picker does not show mods from category B/C.
3. If legacy cross-category mod constraint exists in constraints file, run Autosort and confirm backend returns explicit cross-category error.

- [ ] **Step 4: Final commit only if verification revealed follow-up fixes**

If verification steps required extra edits, run:

```bash
git add app.go internal/service/constraints_service.go frontend/src/components/ConstraintModal.vue frontend/src/components/CycleErrorPanel.vue
git commit -m "chore: finalize autosort category-scope verification fixes"
```

If verification needed no extra edits, skip this step.

---

## Spec-to-plan coverage check (self-review)

- Category constraints sorted with own constraints (`after`, `first`, `last`) — covered by Task 2 category-block sort.
- Mod constraints sorted only inside each category — covered by Task 2 per-block `conGraph.Sort(ids)`.
- Final `orderedIds` composed by category stacking — covered by Task 2 `compileLauncherLayout(layout)` after sorted blocks and sorted `layout.Order`.
- UI must prohibit cross-category mod constraints — covered by Task 4 same-category picker filter.
- Backend must error on cross-category mod constraints — covered by Task 3 add-time validation and Task 1/2 autosort-time validation for existing invalid constraints.

No placeholders, no deferred TODOs.

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-18-autosort-category-scoped-rework.md`.

Two execution options:

1. **Subagent-Driven (recommended)** - dispatch fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** - execute tasks in this session with checkpoints.

Which approach?