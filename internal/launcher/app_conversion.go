package launcher

import (
	"eu5-mod-launcher/internal/repo"

	"github.com/mitchellh/mapstructure"
)

func toRepoSettings(s appSettings) repo.AppSettingsData {
	var result repo.AppSettingsData
	if err := mapstructure.Decode(s, &result); err != nil {
		return repo.AppSettingsData{}
	}
	return result
}

func fromRepoSettings(s repo.AppSettingsData) appSettings {
	var result appSettings
	if err := mapstructure.Decode(s, &result); err != nil {
		return appSettings{}
	}
	return result
}

func toRepoLayout(layout LauncherLayout) repo.LauncherLayoutData {
	var result repo.LauncherLayoutData
	if err := mapstructure.Decode(layout, &result); err != nil {
		return repo.LauncherLayoutData{}
	}
	return result
}

func fromRepoLayout(layout repo.LauncherLayoutData) LauncherLayout {
	var result LauncherLayout
	if err := mapstructure.Decode(layout, &result); err != nil {
		return LauncherLayout{}
	}
	return result
}
