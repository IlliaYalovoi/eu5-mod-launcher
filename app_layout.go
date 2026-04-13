package main

import (
	"fmt"
	"strings"
)

func (a *App) GetLauncherLayout() LauncherLayout {
	return a.launcherLayout
}

func (a *App) SetLauncherLayout(layout LauncherLayout) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("set launcher layout: %w", err)
	}
	next := layout
	if err := a.svc.layoutSvc.Persist(&next, a.loadOrder.OrderedIDs); err != nil {
		return fmt.Errorf("save launcher layout: %w", err)
	}
	a.launcherLayout = next
	return nil
}

func (a *App) CreateLauncherCategory(name string) (LauncherCategory, error) {
	if err := a.ensureReady(); err != nil {
		return LauncherCategory{}, fmt.Errorf("create launcher category: %w", err)
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return LauncherCategory{}, fmt.Errorf("create launcher category: %w", errLauncherCategoryNameEmpty)
	}
	created := LauncherCategory{ID: generateCategoryID(trimmed), Name: trimmed, ModIDs: []string{}}
	a.launcherLayout.Categories = append(a.launcherLayout.Categories, created)
	next := a.launcherLayout
	if err := a.svc.layoutSvc.Persist(&next, a.loadOrder.OrderedIDs); err != nil {
		return LauncherCategory{}, fmt.Errorf("save launcher layout after category create: %w", err)
	}
	a.launcherLayout = next
	return created, nil
}

func (a *App) DeleteLauncherCategory(categoryID string) error {
	if err := a.ensureReady(); err != nil {
		return fmt.Errorf("delete launcher category: %w", err)
	}
	trimmedID := strings.TrimSpace(categoryID)
	if trimmedID == "" {
		return nil
	}
	found := -1
	for i, cat := range a.launcherLayout.Categories {
		if cat.ID == trimmedID {
			found = i
			break
		}
	}
	if found < 0 {
		return nil
	}
	// Move category mods to ungrouped before removing
	a.launcherLayout.Categories[found].ModIDs = nil
	newCategories := append(a.launcherLayout.Categories[:found], a.launcherLayout.Categories[found+1:]...)
	a.launcherLayout.Categories = newCategories
	next := a.launcherLayout
	if err := a.svc.layoutSvc.Persist(&next, a.loadOrder.OrderedIDs); err != nil {
		return fmt.Errorf("save launcher layout after category delete: %w", err)
	}
	a.launcherLayout = next
	return nil
}

func (a *App) SaveCompiledLoadOrder() ([]string, error) {
	if err := a.ensureReady(); err != nil {
		return nil, fmt.Errorf("save compiled load order: %w", err)
	}
	next := a.launcherLayout
	a.svc.layoutSvc.Normalize(&next, a.loadOrder.OrderedIDs)
	a.launcherLayout = next
	compiled := compileLauncherLayout(a.launcherLayout)
	if err := a.SetLoadOrder(compiled); err != nil {
		return nil, fmt.Errorf("persist compiled load order: %w", err)
	}
	return append([]string(nil), a.loadOrder.OrderedIDs...), nil
}
