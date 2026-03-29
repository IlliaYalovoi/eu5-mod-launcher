package mods

// Mod describes one discovered mod on disk.
type Mod struct {
	ID            string // directory name, used as stable identifier
	Name          string // human-readable name from descriptor
	Version       string
	Tags          []string
	Description   string
	ThumbnailPath string // absolute path to thumbnail image, empty if none
	DirPath       string // absolute path to mod directory
	Enabled       bool   // managed by loadorder package, default false
}
