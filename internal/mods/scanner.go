package mods

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"eu5-mod-launcher/internal/logging"
)

// ScanDir walks dirPath and returns one Mod per valid mod subdirectory.
// Errors reading individual mods are logged and skipped, not fatal.
func ScanDir(dirPath string) ([]Mod, error) {
	return ScanDirs([]string{dirPath})
}

// ScanDirs walks multiple roots and returns one Mod per valid mod subdirectory.
// Missing roots are skipped to support optional local/workshop layouts.
func ScanDirs(dirPaths []string) ([]Mod, error) {
	modsByID := make(map[string]Mod)
	for _, root := range dirPaths {
		if strings.TrimSpace(root) == "" {
			continue
		}

		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("resolve absolute mod root %q: %w", root, err)
		}

		entries, err := os.ReadDir(absRoot)
		if err != nil {
			if os.IsNotExist(err) {
				logging.Debugf("mods: skipping missing root %q", absRoot)
				continue
			}
			return nil, fmt.Errorf("read mod root %q: %w", absRoot, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if _, exists := modsByID[entry.Name()]; exists {
				continue
			}

			modDir := filepath.Join(absRoot, entry.Name())
			descriptorPath := filepath.Join(modDir, "descriptor.mod")
			if _, err := os.Stat(descriptorPath); err != nil {
				// EU5 example uses JSON metadata under .metadata; keep this as fallback.
				jsonFallback := filepath.Join(modDir, ".metadata", "metadata.json")
				if _, fallbackErr := os.Stat(jsonFallback); fallbackErr != nil {
					continue
				}
				descriptorPath = jsonFallback
			}

			name, version, description, tags, err := ParseDescriptor(descriptorPath)
			if err != nil {
				logging.Warnf("mods: skipping %q due to descriptor error: %v", modDir, err)
				continue
			}

			modsByID[entry.Name()] = Mod{
				ID:          entry.Name(),
				Name:        name,
				Version:     version,
				Tags:        tags,
				Description: description,
				DirPath:     modDir,
			}
		}
	}

	mods := make([]Mod, 0, len(modsByID))
	for _, mod := range modsByID {
		mods = append(mods, mod)
	}

	sort.Slice(mods, func(i, j int) bool {
		return mods[i].ID < mods[j].ID
	})

	return mods, nil
}
