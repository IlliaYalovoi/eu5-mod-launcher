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

	gv := strings.TrimSpace(strings.ToLower(gameVersion))
	gv = strings.TrimPrefix(gv, "v.")
	gv = strings.TrimPrefix(gv, "v")

	sv := strings.TrimSpace(strings.ToLower(supportedVersion))
	sv = strings.TrimPrefix(sv, "v.")
	sv = strings.TrimPrefix(sv, "v")

	if gv == sv {
		return true
	}

	gParts := strings.Split(gv, ".")
	sParts := strings.Split(sv, ".")

	for i, sPart := range sParts {
		sPart = strings.TrimSpace(sPart)
		if sPart == "*" {
			continue
		}

		if i >= len(gParts) || gParts[i] != sPart {
			return false
		}
	}

	return true
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
