package repo

type SettingsRepo interface {
	Load(path string) (AppSettingsData, error)
	Save(path string, settings AppSettingsData) error
}
