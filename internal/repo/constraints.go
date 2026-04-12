package repo

import "eu5-mod-launcher/internal/domain"

type ConstraintGraph interface {
	Add(from, to string)
	AddFirst(modID string)
	AddLast(modID string)
	Remove(from, to string)
	RemoveFirst(modID string)
	RemoveLast(modID string)
	HasFirst(modID string) bool
	HasLast(modID string) bool
	HasOutgoingAfter(modID string) bool
	HasIncomingAfter(modID string) bool
	ConstraintsFor(modID string) []domain.Constraint
	All() []domain.Constraint
}

type ConstraintsRepository interface {
	Load(path string) (*domain.Graph, error)
	Save(path string, g *domain.Graph) error
}

type FileConstraintsRepository struct{}

func NewFileConstraintsRepository() *FileConstraintsRepository {
	return &FileConstraintsRepository{}
}

func (*FileConstraintsRepository) Load(path string) (*domain.Graph, error) {
	return LoadConstraints(path)
}

func (*FileConstraintsRepository) Save(path string, g *domain.Graph) error {
	return SaveConstraints(path, g)
}
