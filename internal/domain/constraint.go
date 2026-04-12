package domain

type ConstraintType string

const (
	ConstraintAfter ConstraintType = "after"
	ConstraintFirst ConstraintType = "first"
	ConstraintLast  ConstraintType = "last"
)

type TargetType string
const (
	TargetMod   TargetType = "mod"
	TargetGroup TargetType = "group"
)

type Constraint struct {
	Type     ConstraintType `json:"type"`
	FromID   string         `json:"fromId"`
	FromType TargetType     `json:"fromType"`
	ToID     string         `json:"toId"`
	ToType   TargetType     `json:"toType"`
}
