package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"eu5-mod-launcher/internal/domain"
)

const categoryIDPrefix = "category:"
const defaultUngroupedCategoryID = "category:ungrouped"

type LauncherCategory struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	ModIDs []string `json:"mod_ids"`
}

type LauncherLayout struct {
	Ungrouped  []string           `json:"ungrouped"`
	Categories []LauncherCategory `json:"categories"`
	Order      []string           `json:"order,omitempty"`
	Collapsed  map[string]bool    `json:"collapsed,omitempty"`
}

func defaultLauncherLayout(enabled []string) LauncherLayout {
	return LauncherLayout{
		Ungrouped:  append([]string(nil), enabled...),
		Categories: []LauncherCategory{},
		Order:      []string{defaultUngroupedCategoryID},
		Collapsed:  map[string]bool{},
	}
}

func loadLauncherLayout(path string) (LauncherLayout, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LauncherLayout{}, nil
		}
		return LauncherLayout{}, fmt.Errorf("read launcher layout %q: %w", path, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		return LauncherLayout{}, nil
	}

	var layout LauncherLayout
	if err := json.Unmarshal(content, &layout); err != nil {
		return LauncherLayout{}, fmt.Errorf("decode launcher layout %q: %w", path, err)
	}
	if layout.Ungrouped == nil {
		layout.Ungrouped = []string{}
	}
	if layout.Categories == nil {
		layout.Categories = []LauncherCategory{}
	}
	if layout.Order == nil {
		layout.Order = []string{}
	}
	if layout.Collapsed == nil {
		layout.Collapsed = map[string]bool{}
	}
	for i := range layout.Categories {
		if layout.Categories[i].ModIDs == nil {
			layout.Categories[i].ModIDs = []string{}
		}
	}

	return layout, nil
}

func saveLauncherLayout(path string, layout LauncherLayout) error {
	if layout.Ungrouped == nil {
		layout.Ungrouped = []string{}
	}
	if layout.Categories == nil {
		layout.Categories = []LauncherCategory{}
	}
	if layout.Order == nil {
		layout.Order = []string{}
	}
	if layout.Collapsed == nil {
		layout.Collapsed = map[string]bool{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create launcher layout dir for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(layout, "", "  ")
	if err != nil {
		return fmt.Errorf("encode launcher layout %q: %w", path, err)
	}
	payload = append(payload, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o644); err != nil {
		return fmt.Errorf("write launcher layout tmp %q: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace launcher layout %q: %w", path, err)
	}
	return nil
}

func normalizeLauncherLayout(layout LauncherLayout, enabled []string) LauncherLayout {
	enabledSet := make(map[string]struct{}, len(enabled))
	for _, id := range enabled {
		enabledSet[id] = struct{}{}
	}

	seen := make(map[string]struct{}, len(enabled))
	normalized := LauncherLayout{Ungrouped: []string{}, Categories: []LauncherCategory{}, Order: []string{}, Collapsed: map[string]bool{}}

	for _, cat := range layout.Categories {
		if strings.TrimSpace(cat.ID) == "" {
			continue
		}
		name := strings.TrimSpace(cat.Name)
		if name == "" {
			name = cat.ID
		}
		next := LauncherCategory{ID: cat.ID, Name: name, ModIDs: []string{}}
		for _, modID := range cat.ModIDs {
			if _, ok := enabledSet[modID]; !ok {
				continue
			}
			if _, exists := seen[modID]; exists {
				continue
			}
			seen[modID] = struct{}{}
			next.ModIDs = append(next.ModIDs, modID)
		}
		normalized.Categories = append(normalized.Categories, next)
	}

	availableOrderIDs := map[string]struct{}{defaultUngroupedCategoryID: {}}
	for _, cat := range normalized.Categories {
		availableOrderIDs[cat.ID] = struct{}{}
	}
	seenOrder := map[string]struct{}{}
	for _, id := range layout.Order {
		if _, ok := availableOrderIDs[id]; !ok {
			continue
		}
		if _, exists := seenOrder[id]; exists {
			continue
		}
		seenOrder[id] = struct{}{}
		normalized.Order = append(normalized.Order, id)
	}
	if _, exists := seenOrder[defaultUngroupedCategoryID]; !exists {
		normalized.Order = append(normalized.Order, defaultUngroupedCategoryID)
		seenOrder[defaultUngroupedCategoryID] = struct{}{}
	}
	for _, cat := range normalized.Categories {
		if _, exists := seenOrder[cat.ID]; exists {
			continue
		}
		normalized.Order = append(normalized.Order, cat.ID)
		seenOrder[cat.ID] = struct{}{}
	}

	for _, modID := range layout.Ungrouped {
		if _, ok := enabledSet[modID]; !ok {
			continue
		}
		if _, exists := seen[modID]; exists {
			continue
		}
		seen[modID] = struct{}{}
		normalized.Ungrouped = append(normalized.Ungrouped, modID)
	}

	for _, modID := range enabled {
		if _, exists := seen[modID]; exists {
			continue
		}
		seen[modID] = struct{}{}
		normalized.Ungrouped = append(normalized.Ungrouped, modID)
	}

	for id, collapsed := range layout.Collapsed {
		if _, ok := availableOrderIDs[id]; ok {
			normalized.Collapsed[id] = collapsed
		}
	}

	return normalized
}

func compileLauncherLayout(layout LauncherLayout) []string {
	out := make([]string, 0, len(layout.Ungrouped))
	seen := make(map[string]struct{})
	categoryByID := map[string]LauncherCategory{}
	for _, cat := range layout.Categories {
		categoryByID[cat.ID] = cat
	}

	for _, orderID := range layout.Order {
		modIDs := []string{}
		if orderID == defaultUngroupedCategoryID {
			modIDs = layout.Ungrouped
		} else if cat, ok := categoryByID[orderID]; ok {
			modIDs = cat.ModIDs
		}

		for _, id := range modIDs {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}

	return out
}

func generateCategoryID(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	if slug == "" {
		slug = "category"
	}
	return fmt.Sprintf("%s%s-%d", categoryIDPrefix, slug, time.Now().UnixNano())
}

func isCategoryID(id string) bool {
	return domain.IsCategoryID(id)
}

func categoryNameMap(layout LauncherLayout) map[string]string {
	out := make(map[string]string, len(layout.Categories))
	for _, cat := range layout.Categories {
		out[cat.ID] = cat.Name
	}
	return out
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
