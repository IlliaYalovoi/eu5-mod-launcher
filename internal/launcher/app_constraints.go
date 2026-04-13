package launcher

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/logging"
	"fmt"
)

func (a *App) GetConstraints() []domain.Constraint {
	if err := a.mustBeReady(); err != nil {
		return []domain.Constraint{}
	}
	return a.svc.conService.All()
}

func (a *App) AddConstraint(from, target string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if err := a.svc.conService.AddConstraint(from, target); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, target, err)
	}
	return nil
}

func (a *App) AddLoadFirst(modID string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if err := a.svc.conService.AddLoadFirst(modID); err != nil {
		return fmt.Errorf("add load-first %q: %w", modID, err)
	}
	return nil
}

func (a *App) AddLoadLast(modID string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if err := a.svc.conService.AddLoadLast(modID); err != nil {
		return fmt.Errorf("add load-last %q: %w", modID, err)
	}
	return nil
}

func (a *App) RemoveConstraint(from, target string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if err := a.svc.conService.RemoveConstraint(from, target); err != nil {
		return fmt.Errorf("remove constraint %q -> %q: %w", from, target, err)
	}
	return nil
}

func (a *App) RemoveLoadFirst(modID string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if err := a.svc.conService.RemoveLoadFirst(modID); err != nil {
		return fmt.Errorf("remove load-first %q: %w", modID, err)
	}
	return nil
}

func (a *App) RemoveLoadLast(modID string) error {
	if err := a.mustBeReady(); err != nil {
		return err
	}
	if err := a.svc.conService.RemoveLoadLast(modID); err != nil {
		return fmt.Errorf("remove load-last %q: %w", modID, err)
	}
	return nil
}

func (a *App) Autosort() ([]string, error) {
	if err := a.mustBeReady(); err != nil {
		return nil, err
	}
	prevOrder := append([]string(nil), a.loadOrder.OrderedIDs...)
	prevLayout := a.launcherLayout

	sorter := NewGraphSorter(a.svc.conGraph)
	sorted, err := sorter.Sort(a.loadOrder.OrderedIDs)
	if err != nil {
		return nil, fmt.Errorf("sort constraints: %w", err)
	}
	if err := a.SetLoadOrder(sorted); err != nil {
		return nil, fmt.Errorf("persist autosorted load order: %w", err)
	}
	nextLayout, err := a.reorderLauncherLayoutAfterAutosort(sorted)
	if err != nil {
		if rbErr := a.SetLoadOrder(prevOrder); rbErr != nil {
			logging.Errorf("autosort rollback failed after category-sort error: %v", rbErr)
		}
		a.launcherLayout = prevLayout
		return nil, fmt.Errorf("sort category constraints: %w", err)
	}
	a.launcherLayout = nextLayout
	if err := a.svc.layoutRepo.Save(a.layoutPath, toRepoLayout(a.launcherLayout)); err != nil {
		if rbErr := a.SetLoadOrder(prevOrder); rbErr != nil {
			logging.Errorf("autosort rollback failed after layout save error: %v", rbErr)
		}
		a.launcherLayout = prevLayout
		return nil, fmt.Errorf("save launcher layout: %w", err)
	}
	return append([]string(nil), a.loadOrder.OrderedIDs...), nil
}

func (a *App) reorderLauncherLayoutAfterAutosort(sortedIDs []string) (LauncherLayout, error) {
	enabledSet := make(map[string]struct{}, len(a.loadOrder.OrderedIDs))
	for _, id := range a.loadOrder.OrderedIDs {
		enabledSet[id] = struct{}{}
	}
	layout := a.launcherLayout

	// Rebuild ungrouped: mods that were ungrouped and are still enabled
	newUngrouped := make([]string, 0, len(layout.Ungrouped))
	seen := make(map[string]struct{})
	for _, id := range layout.Ungrouped {
		if _, ok := enabledSet[id]; ok {
			newUngrouped = append(newUngrouped, id)
			seen[id] = struct{}{}
		}
	}
	// Add mods that became ungrouped because their category is empty
	for _, id := range sortedIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		inAnyCategory := false
		for _, cat := range layout.Categories {
			for _, catModID := range cat.ModIDs {
				if catModID == id {
					inAnyCategory = true
					break
				}
			}
			if inAnyCategory {
				break
			}
		}
		if !inAnyCategory {
			newUngrouped = append(newUngrouped, id)
			seen[id] = struct{}{}
		}
	}
	layout.Ungrouped = newUngrouped

	// Reorder categories to match sorted order
	for i := range layout.Categories {
		layout.Categories[i].ModIDs = reorderCategoryMods(layout.Categories[i].ModIDs, seen, enabledSet)
	}

	return layout, nil
}

func reorderCategoryMods(catModIDs []string, seen map[string]struct{}, enabledSet map[string]struct{}) []string {
	result := make([]string, 0, len(catModIDs))
	for _, id := range catModIDs {
		if _, ok := enabledSet[id]; !ok {
			continue
		}
		result = append(result, id)
		seen[id] = struct{}{}
	}
	return result
}
