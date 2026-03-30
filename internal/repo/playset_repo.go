package repo

import (
	"eu5-mod-launcher/internal/loadorder"
)

type PlaysetRepository interface {
	ListPlaysets(path string) ([]string, int, error)
	LoadState(path string, index int) (loadorder.State, map[string]string, error)
	SaveState(path string, index int, state loadorder.State, modPathByID map[string]string) error
}

type FilePlaysetRepository struct{}

func NewFilePlaysetRepository() *FilePlaysetRepository {
	return &FilePlaysetRepository{}
}

func (*FilePlaysetRepository) ListPlaysets(path string) ([]string, int, error) {
	return loadorder.ListPlaysets(path)
}

func (*FilePlaysetRepository) LoadState(path string, index int) (loadorder.State, map[string]string, error) {
	return loadorder.LoadStateFromPlaysets(path, index)
}

func (*FilePlaysetRepository) SaveState(
	path string,
	index int,
	state loadorder.State,
	modPathByID map[string]string,
) error {
	return loadorder.SaveStateToPlaysets(path, index, state, modPathByID)
}
