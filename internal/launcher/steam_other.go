//go:build !windows

package launcher

var steamInstallPathFinder = discoverSteamInstallPath

func discoverSteamInstallPath() string {
	return ""
}