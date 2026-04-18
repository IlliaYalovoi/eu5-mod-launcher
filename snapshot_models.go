package main

import (
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/mods"
)

type SnapshotMeta struct {
	Revision  int64 `json:"revision"`
	FetchedAt int64 `json:"fetchedAt"`
	Stale     bool  `json:"stale"`
}

type SnapshotSettings struct {
	ModsDirStatus       ModsDirStatus `json:"modsDirStatus"`
	GameExe             string        `json:"gameExe"`
	AutoDetectedGameExe string        `json:"autoDetectedGameExe"`
	ConfigPath          string        `json:"configPath"`
	GameVersion         string        `json:"gameVersion"`
	GameVersionOverride string        `json:"gameVersionOverride"`
	AvailableGames      []string      `json:"availableGames"`
}

type GameSnapshot struct {
	GameID                     string             `json:"gameID"`
	Mods                       []mods.Mod         `json:"mods"`
	LoadOrder                  []string           `json:"loadOrder"`
	LauncherLayout             LauncherLayout     `json:"launcherLayout"`
	Constraints                []graph.Constraint `json:"constraints"`
	PlaysetNames               []string           `json:"playsetNames"`
	GameActivePlaysetIndex     int                `json:"gameActivePlaysetIndex"`
	LauncherActivePlaysetIndex int                `json:"launcherActivePlaysetIndex"`
	Settings                   SnapshotSettings   `json:"settings"`
	Meta                       SnapshotMeta       `json:"meta"`
}
