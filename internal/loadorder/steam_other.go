//go:build !windows

package loadorder

import (
	"os"
	"path/filepath"
)

func init() {
	steamInstallPathFinder = defaultSteamInstallPath
}

func defaultSteamInstallPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	// Standard Linux Steam paths
	candidates := []string{
		filepath.Join(home, ".local/share/Steam"),
		filepath.Join(home, ".steam/steam"),
	}
	for _, c := range candidates {
		if dirExists(c) {
			return c
		}
	}
	return ""
}
