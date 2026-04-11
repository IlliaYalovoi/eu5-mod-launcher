package main

import "eu5-mod-launcher/internal/repo"

func toRepoSettings(s appSettings) repo.AppSettingsData {
	return repo.AppSettingsData{
		ModsDir:                    s.ModsDir,
		GameExe:                    s.GameExe,
		GameArgs:                   append([]string(nil), s.GameArgs...),
		LauncherActivePlaysetIndex: s.LauncherActivePlaysetIndex,
	}
}

func fromRepoSettings(s repo.AppSettingsData) appSettings {
	return appSettings{
		ModsDir:                    s.ModsDir,
		GameExe:                    s.GameExe,
		GameArgs:                   append([]string(nil), s.GameArgs...),
		LauncherActivePlaysetIndex: s.LauncherActivePlaysetIndex,
	}
}

func toRepoLayout(layout LauncherLayout) repo.LauncherLayoutData {
	cats := make([]repo.LauncherCategoryData, 0, len(layout.Categories))
	for _, c := range layout.Categories {
		cats = append(cats, repo.LauncherCategoryData{
			ID:     c.ID,
			Name:   c.Name,
			ModIDs: append([]string(nil), c.ModIDs...),
		})
	}
	collapsed := map[string]bool{}
	for id, v := range layout.Collapsed {
		collapsed[id] = v
	}
	return repo.LauncherLayoutData{
		Ungrouped:  append([]string(nil), layout.Ungrouped...),
		Categories: cats,
		Order:      append([]string(nil), layout.Order...),
		Collapsed:  collapsed,
	}
}

func fromRepoLayout(layout repo.LauncherLayoutData) LauncherLayout {
	cats := make([]LauncherCategory, 0, len(layout.Categories))
	for _, c := range layout.Categories {
		cats = append(cats, LauncherCategory{
			ID:     c.ID,
			Name:   c.Name,
			ModIDs: append([]string(nil), c.ModIDs...),
		})
	}
	collapsed := map[string]bool{}
	for id, v := range layout.Collapsed {
		collapsed[id] = v
	}
	return LauncherLayout{
		Ungrouped:  append([]string(nil), layout.Ungrouped...),
		Categories: cats,
		Order:      append([]string(nil), layout.Order...),
		Collapsed:  collapsed,
	}
}
