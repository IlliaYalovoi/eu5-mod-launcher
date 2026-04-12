//go:build windows

package launcher

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

var steamInstallPathFinder = windowsSteamInstallPath

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
		if closeErr := key.Close(); closeErr != nil {
			continue
		}
		if err != nil {
			continue
		}
		if dirExists(installPath) {
			return filepath.Clean(installPath)
		}
	}

	return defaultSteamInstallPath()
}

func defaultSteamInstallPath() string {
	fallbacks := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Steam"),
		filepath.Join(os.Getenv("ProgramFiles"), "Steam"),
	}

	for _, candidate := range fallbacks {
		if dirExists(candidate) {
			return filepath.Clean(candidate)
		}
	}

	return ""
}
