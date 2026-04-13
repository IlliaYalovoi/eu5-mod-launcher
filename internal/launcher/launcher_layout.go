package launcher

import (
	"eu5-mod-launcher/internal/domain"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	categoryIDPrefix           = "category:"
	defaultUngroupedCategoryID = "category:ungrouped"
)

type LauncherCategory struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	ModIDs []string `json:"modIds"`
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

func normalizeLauncherLayout(layout LauncherLayout, enabled []string) LauncherLayout {
	enabledSet := buildEnabledSet(enabled)
	seen := make(map[string]struct{}, len(enabled))

	normalized := LauncherLayout{
		Ungrouped:  []string{},
		Categories: normalizeCategories(layout.Categories, enabledSet, seen),
		Order:      []string{},
		Collapsed:  map[string]bool{},
	}

	availableOrderIDs := buildAvailableOrderIDs(normalized.Categories)
	normalized.Order = normalizeOrder(layout.Order, normalized.Categories, availableOrderIDs)
	normalized.Ungrouped = normalizeUngrouped(layout.Ungrouped, enabled, enabledSet, seen)
	normalized.Collapsed = normalizeCollapsed(layout.Collapsed, availableOrderIDs)

	return normalized
}

func buildEnabledSet(enabled []string) map[string]struct{} {
	enabledSet := make(map[string]struct{}, len(enabled))
	for _, id := range enabled {
		enabledSet[id] = struct{}{}
	}
	return enabledSet
}

func normalizeCategories(
	categories []LauncherCategory,
	enabledSet map[string]struct{},
	seen map[string]struct{},
) []LauncherCategory {
	normalized := make([]LauncherCategory, 0, len(categories))
	for i := range categories {
		cat := categories[i]
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
		normalized = append(normalized, next)
	}

	return normalized
}

func buildAvailableOrderIDs(categories []LauncherCategory) map[string]struct{} {
	available := map[string]struct{}{defaultUngroupedCategoryID: {}}
	for i := range categories {
		available[categories[i].ID] = struct{}{}
	}
	return available
}

func normalizeOrder(
	order []string,
	categories []LauncherCategory,
	availableOrderIDs map[string]struct{},
) []string {
	out := make([]string, 0, len(order)+len(categories)+1)
	seenOrder := map[string]struct{}{}
	for _, id := range order {
		if _, ok := availableOrderIDs[id]; !ok {
			continue
		}
		if _, exists := seenOrder[id]; exists {
			continue
		}
		seenOrder[id] = struct{}{}
		out = append(out, id)
	}
	if _, exists := seenOrder[defaultUngroupedCategoryID]; !exists {
		out = append(out, defaultUngroupedCategoryID)
		seenOrder[defaultUngroupedCategoryID] = struct{}{}
	}
	for i := range categories {
		catID := categories[i].ID
		if _, exists := seenOrder[catID]; exists {
			continue
		}
		out = append(out, catID)
		seenOrder[catID] = struct{}{}
	}
	return out
}

func normalizeUngrouped(
	ungrouped []string,
	enabled []string,
	enabledSet map[string]struct{},
	seen map[string]struct{},
) []string {
	out := make([]string, 0, len(enabled))
	for _, modID := range ungrouped {
		if _, ok := enabledSet[modID]; !ok {
			continue
		}
		if _, exists := seen[modID]; exists {
			continue
		}
		seen[modID] = struct{}{}
		out = append(out, modID)
	}

	for _, modID := range enabled {
		if _, exists := seen[modID]; exists {
			continue
		}
		seen[modID] = struct{}{}
		out = append(out, modID)
	}

	return out
}

func normalizeCollapsed(collapsed map[string]bool, availableOrderIDs map[string]struct{}) map[string]bool {
	out := map[string]bool{}
	for id, isCollapsed := range collapsed {
		if _, ok := availableOrderIDs[id]; ok {
			out[id] = isCollapsed
		}
	}
	return out
}

func compileLauncherLayout(layout LauncherLayout) []string {
	out := make([]string, 0, len(layout.Ungrouped))
	seen := make(map[string]struct{})
	categoryByID := make(map[string]LauncherCategory, len(layout.Categories))
	for i := range layout.Categories {
		cat := layout.Categories[i]
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

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
