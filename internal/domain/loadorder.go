package domain

type LoadOrder struct {
	GameID       GameID       `json:"gameId"`
	PlaysetIdx   PlaysetIndex `json:"playsetIdx"`
	ActiveModIDs []string     `json:"activeModIds"` // Filtered list
}
