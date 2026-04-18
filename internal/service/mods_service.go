package service

import (
	"eu5-mod-launcher/internal/mods"
	"fmt"
	"maps"
	"strings"
)

type ModsService struct{}

func NewModsService() *ModsService {
	return &ModsService{}
}

func IsVersionCompatible(gameVersion, supportedVersion string) bool {
	if supportedVersion == "" {
		return true // Treat empty supported version as ANY
	}
	if gameVersion == "unknown" {
		return false
	}
	if gameVersion == supportedVersion {
		return true
	}
	prefix := strings.ReplaceAll(supportedVersion, "*", "")
	return strings.HasPrefix(gameVersion, prefix)
}

func (*ModsService) Discover(
	scanRoots, enabledIDs []string,
	knownPaths map[string]string,
	gameVersion string,
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
		allMods[i].IsCompatible = IsVersionCompatible(gameVersion, allMods[i].SupportedVersion)
	}

	return allMods, nextPaths, nil
}
