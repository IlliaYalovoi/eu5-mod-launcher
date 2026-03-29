package repo

import "eu5-mod-launcher/internal/graph"

type ConstraintsRepository interface {
	Load(path string) (*graph.Graph, error)
	Save(path string, g *graph.Graph) error
}

type FileConstraintsRepository struct{}

func NewFileConstraintsRepository() *FileConstraintsRepository {
	return &FileConstraintsRepository{}
}

func (r *FileConstraintsRepository) Load(path string) (*graph.Graph, error) {
	return graph.LoadConstraints(path)
}

func (r *FileConstraintsRepository) Save(path string, g *graph.Graph) error {
	return graph.SaveConstraints(path, g)
}
