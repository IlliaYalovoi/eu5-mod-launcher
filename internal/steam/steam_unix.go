//go:build !windows

package steam

func findSteamInstallPath() string {
	return defaultSteamInstallPath()
}

func defaultSteamInstallPath() string {
	return ""
}
