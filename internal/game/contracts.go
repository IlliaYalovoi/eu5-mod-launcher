package game

import "eu5-mod-launcher/internal/loadorder"

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

type ModListAdapter interface {
	GameID() GameID
	Descriptor() GameDescriptor
	DiscoverPaths() (loadorder.GamePaths, error)
	ListModLists(playsetsPath string) ([]string, int, error)
	ImportModList(playsetsPath string, listIndex int) (loadorder.State, map[string]string, error)
	ExportModList(playsetsPath string, listIndex int, state loadorder.State, modPathByID map[string]string) error
}

