package service

import (
	"eu5-mod-launcher/internal/mods"
	"fmt"
	"maps"
)

type ModsService struct{}

func NewModsService() *ModsService {
	return &ModsService{}
}

func (*ModsService) Discover(
	scanRoots, enabledIDs []string,
	knownPaths map[string]string,
) ([]mods.Mod, map[string]string, error) {
	allMods, err := mods.ScanDirs(scanRoots)
	if err != nil {
		return nil, nil, fmt.Errorf("scan mods roots %q: %w", scanRoots, err)
	}

	enabled := make(map[string]struct{}, len(enabledIDs))
	for _, id := range enabledIDs {
		enabled[id] = struct{}{}
	}

	nextPaths := make(map[string]string, len(knownPaths)+len(allMods))
	maps.Copy(nextPaths, knownPaths)
	for i := range allMods {
		nextPaths[allMods[i].ID] = allMods[i].DirPath
		_, ok := enabled[allMods[i].ID]
		allMods[i].Enabled = ok
	}

	return allMods, nextPaths, nil
}
