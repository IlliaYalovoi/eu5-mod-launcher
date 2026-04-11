package domain

type LoadOrder struct {
	GameID     GameID       `json:"gameId"`
	PlaysetIdx PlaysetIndex `json:"playsetIdx"`
	OrderedIDs []string     `json:"orderedIds"`
}
