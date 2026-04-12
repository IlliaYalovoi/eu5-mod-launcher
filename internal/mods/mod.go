package mods

type Mod struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	Tags          []string `json:"tags"`
	Description   string   `json:"description"`
	ThumbnailPath string   `json:"thumbnailPath"`
	DirPath       string   `json:"dirPath"`
	Enabled       bool     `json:"enabled"`
}
