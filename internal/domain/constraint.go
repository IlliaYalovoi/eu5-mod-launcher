package domain

type ConstraintType string

const (
	ConstraintAfter ConstraintType = "after"
	ConstraintFirst ConstraintType = "first"
	ConstraintLast  ConstraintType = "last"
)

type Constraint struct {
	Type  ConstraintType `json:"type,omitempty"`
	From  string         `json:"from,omitempty"`
	To    string         `json:"to,omitempty"`
	ModID string         `json:"modId,omitempty"`
}
