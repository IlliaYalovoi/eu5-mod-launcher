package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ListPlaysets(path string) ([]string, int, error) {
	root, err := readPlaysetsRoot(path)
	if err != nil {
		return nil, -1, fmt.Errorf("list playsets from %q: %w", path, err)
	}

	playsets, ok := root["playsets"].([]any)
	if !ok || len(playsets) == 0 {
		return []string{}, -1, nil
	}

	names := make([]string, 0, len(playsets))
	gameActiveIndex := -1

	for i, item := range playsets {
		playset, ok := item.(map[string]any)
		if !ok {
			names = append(names, fmt.Sprintf("Playset %d", i+1))
			continue
		}

		name := ""
		if rawName, ok := playset["name"].(string); ok {
			name = rawName
		}
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("Playset %d", i+1)
		}
		names = append(names, name)

		if gameActiveIndex < 0 {
			if isActive, ok := playset["isActive"].(bool); ok && isActive {
				gameActiveIndex = i
			}
		}
	}

	if gameActiveIndex < 0 && len(playsets) > 0 {
		gameActiveIndex = 0
	}

	return names, gameActiveIndex, nil
}

func LoadStateFromPlaysets(path string, playsetIndex int) (State, map[string]string, error) {
	root, err := readPlaysetsRoot(path)
	if err != nil {
		return State{}, nil, fmt.Errorf("load playset state from %q: %w", path, err)
	}

	playsets, ok := root["playsets"].([]any)
	if !ok || len(playsets) == 0 {
		return State{ActiveModIDs: []string{}}, map[string]string{}, nil
	}

	resolvedIndex := resolvePlaysetIndex(playsets, playsetIndex, gameActiveIndex(playsets))
	entries := playsetEntries(playsets[resolvedIndex])

	ids := make([]string, 0, len(entries))
	pathsByID := make(map[string]string, len(entries))
	seen := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		if !entry.enabled {
			continue
		}

		id := modIDFromPath(entry.path)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
		pathsByID[id] = normalizeModPath(entry.path)
	}

	return State{ActiveModIDs: ids}, pathsByID, nil
}

func SaveStateToPlaysets(path string, playsetIndex int, state State, idToPath map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create playsets dir for %q: %w", path, err)
	}

	root, err := loadPlaysetsRoot(path)
	if err != nil {
		return err
	}

	playsets := ensurePlaysets(root)
	selectedPlayset := ensureSelectedPlayset(playsets, playsetIndex)
	orderedEntries := buildOrderedEntries(state.ActiveModIDs, idToPath)

	selectedPlayset["orderedListMods"] = orderedEntries
	root["playsets"] = playsets

	payload, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("encode playsets payload for %q: %w", path, err)
	}
	payload = append(payload, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write temporary playsets file %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf(
				"replace playsets file %q: %w; cleanup temp %q: %s",
				path,
				err,
				tmpPath,
				removeErr.Error(),
			)
		}
		return fmt.Errorf("replace playsets file %q: %w", path, err)
	}

	return nil
}

func loadPlaysetsRoot(path string) (map[string]any, error) {
	root := map[string]any{}
	content, err := os.ReadFile(path)
	if err == nil && strings.TrimSpace(string(content)) != "" {
		if decodeErr := json.Unmarshal(content, &root); decodeErr != nil {
			return nil, fmt.Errorf("decode existing playsets file %q: %w", path, decodeErr)
		}
	}

	if _, ok := root["file_version"]; !ok {
		root["file_version"] = "1.0.0"
	}

	return root, nil
}

func ensureSelectedPlayset(playsets []any, preferredIndex int) map[string]any {
	resolvedIndex := resolvePlaysetIndex(playsets, preferredIndex, gameActiveIndex(playsets))
	selectedPlayset, ok := playsets[resolvedIndex].(map[string]any)
	if ok && selectedPlayset != nil {
		return selectedPlayset
	}

	selectedPlayset = map[string]any{}
	playsets[resolvedIndex] = selectedPlayset

	return selectedPlayset
}

func buildOrderedEntries(orderedIDs []string, idToPath map[string]string) []any {
	orderedEntries := make([]any, 0, len(orderedIDs))
	seen := make(map[string]struct{}, len(orderedIDs))
	for _, id := range orderedIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		pathValue := normalizeModPath(idToPath[id])
		if pathValue == "" {
			continue
		}

		orderedEntries = append(orderedEntries, map[string]any{
			"path":      toGamePath(pathValue),
			"isEnabled": true,
		})
	}

	return orderedEntries
}

type playsetEntry struct {
	path    string
	enabled bool
}

func playsetEntries(playsetRaw any) []playsetEntry {
	playset, ok := playsetRaw.(map[string]any)
	if !ok {
		return nil
	}

	mods, ok := playset["orderedListMods"].([]any)
	if !ok {
		return nil
	}

	out := make([]playsetEntry, 0, len(mods))
	for _, modItem := range mods {
		modEntry, ok := modItem.(map[string]any)
		if !ok {
			continue
		}

		pathValue := ""
		if rawPath, ok := modEntry["path"].(string); ok {
			pathValue = rawPath
		}
		enabled := true
		if rawEnabled, ok := modEntry["isEnabled"].(bool); ok {
			enabled = rawEnabled
		}

		out = append(out, playsetEntry{path: pathValue, enabled: enabled})
	}

	return out
}

func readPlaysetsRoot(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read playsets file %q: %w", path, err)
	}

	if strings.TrimSpace(string(content)) == "" {
		return map[string]any{}, nil
	}

	root := map[string]any{}
	if err := json.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("decode playsets file %q: %w", path, err)
	}

	return root, nil
}

func gameActiveIndex(playsets []any) int {
	for i, item := range playsets {
		playset, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if isActive, ok := playset["isActive"].(bool); ok && isActive {
			return i
		}
	}

	if len(playsets) == 0 {
		return -1
	}
	return 0
}

func resolvePlaysetIndex(playsets []any, requested, gameActive int) int {
	if requested >= 0 && requested < len(playsets) {
		return requested
	}
	if gameActive >= 0 && gameActive < len(playsets) {
		return gameActive
	}
	if len(playsets) == 0 {
		return -1
	}
	return 0
}

func ensurePlaysets(root map[string]any) []any {
	playsets, ok := root["playsets"].([]any)
	if ok && len(playsets) > 0 {
		return playsets
	}

	return []any{
		map[string]any{
			"name":                  "Default",
			"isAutomaticallySorted": false,
			"orderedListMods":       []any{},
		},
	}
}

func modIDFromPath(raw string) string {
	normalized := normalizeModPath(raw)
	if normalized == "" {
		return ""
	}
	return filepath.Base(normalized)
}

func normalizeModPath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimRight(trimmed, "\\/")
	if trimmed == "" {
		return ""
	}
	withSlashes := strings.ReplaceAll(trimmed, "\\", "/")
	return filepath.Clean(filepath.FromSlash(withSlashes))
}

func toGamePath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(path))
	if !strings.HasSuffix(normalized, "/") {
		normalized += "/"
	}
	return normalized
}
