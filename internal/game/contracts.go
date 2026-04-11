package game

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/loadorder"
)

type GameID = domain.GameID

const (
	GameIDEU5  GameID = domain.GameIDEU5
	GameIDVic3 GameID = domain.GameIDVic3
)

type GameDescriptor = domain.GameDescriptor

type ModListAdapter interface {
	Adapter
	ListModLists(playsetsPath string) ([]string, int, error)
	ImportModList(playsetsPath string, listIndex int) (loadorder.State, map[string]string, error)
	ExportModList(playsetsPath string, listIndex int, state loadorder.State, modPathByID map[string]string) error
}
