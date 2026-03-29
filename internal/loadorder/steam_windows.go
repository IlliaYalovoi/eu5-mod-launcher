//go:build windows

package loadorder

import (
	"golang.org/x/sys/windows/registry"
	"path/filepath"
)

func init() {
	steamInstallPathFinder = windowsSteamInstallPath
}

func windowsSteamInstallPath() string {
	registryLocations := []struct {
		root registry.Key
		path string
	}{
		{root: registry.CURRENT_USER, path: `Software\Valve\Steam`},
		{root: registry.LOCAL_MACHINE, path: `SOFTWARE\WOW6432Node\Valve\Steam`},
		{root: registry.LOCAL_MACHINE, path: `SOFTWARE\Valve\Steam`},
	}

	for _, location := range registryLocations {
		key, err := registry.OpenKey(location.root, location.path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		installPath, _, err := key.GetStringValue("InstallPath")
		_ = key.Close()
		if err != nil {
			continue
		}
		if dirExists(installPath) {
			return filepath.Clean(installPath)
		}
	}

	return defaultSteamInstallPath()
}
