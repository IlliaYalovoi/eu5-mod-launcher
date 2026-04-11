package domain

type GameID string

const (
	GameIDEU5  GameID = "eu5"
	GameIDVic3 GameID = "vic3"
)

type GameDescriptor struct {
	ID           GameID `json:"id"`
	DisplayName  string `json:"displayName"`
	Detected     bool   `json:"detected"`
	InstallDir   string `json:"installDir"`
	DocumentsDir string `json:"documentsDir"`
}

type GamePaths struct {
	PlaysetsPath    string   `json:"playsetsPath"`
	LocalModsDir    string   `json:"localModsDir"`
	WorkshopModDirs []string `json:"workshopModDirs"`
	GameExePath     string   `json:"gameExePath"`
}
