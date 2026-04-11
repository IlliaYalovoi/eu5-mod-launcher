package service

import (
	"eu5-mod-launcher/internal/logging"
	"eu5-mod-launcher/internal/mods"
	"fmt"
)

type ModsService struct{}

func NewModsService() *ModsService {
	return &ModsService{}
}

func (*ModsService) Discover(
	scanRoots, enabledIDs []string,
	knownPaths map[string]string,
) ([]mods.Mod, map[string]string, error) {
	logging.Debugf("mods-service: scanning %d roots for mods", len(scanRoots))

	allMods, err := mods.ScanDirs(scanRoots)
	if err != nil {
		logging.Errorf("mods-service: scan failed: %v", err)
		return nil, nil, fmt.Errorf("scan mods roots %q: %w", scanRoots, err)
	}

	enabled := make(map[string]struct{}, len(enabledIDs))
	for _, id := range enabledIDs {
		enabled[id] = struct{}{}
	}

	nextPaths := make(map[string]string, len(knownPaths)+len(allMods))
	for id, path := range knownPaths {
		nextPaths[id] = path
	}
	for i := range allMods {
		nextPaths[allMods[i].ID] = allMods[i].DirPath
		_, ok := enabled[allMods[i].ID]
		allMods[i].Enabled = ok
	}

	logging.Infof("mods-service: discovered %d mods (%d enabled)", len(allMods), len(enabled))

	return allMods, nextPaths, nil
}
