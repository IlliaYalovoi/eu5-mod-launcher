package repo

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/loadorder"
)

type FilePlaysetRepo struct{}

func NewFilePlaysetRepo() *FilePlaysetRepo { return &FilePlaysetRepo{} }

func (*FilePlaysetRepo) ListPlaysets(path string) ([]string, domain.PlaysetIndex, error) {
	names, idx, err := loadorder.ListPlaysets(path)
	return names, domain.PlaysetIndex(idx), err
}

func (*FilePlaysetRepo) LoadState(path string, idx domain.PlaysetIndex) (domain.LoadOrder, map[string]string, error) {
	state, paths, err := loadorder.LoadStateFromPlaysets(path, int(idx))
	if err != nil {
		return domain.LoadOrder{}, nil, err
	}
	return domain.LoadOrder{GameID: domain.GameIDEU5, PlaysetIdx: idx, OrderedIDs: state.OrderedIDs}, paths, nil
}

func (*FilePlaysetRepo) SaveState(path string, idx domain.PlaysetIndex, order domain.LoadOrder, modPathByID map[string]string) error {
	return loadorder.SaveStateToPlaysets(path, int(idx), loadorder.State{OrderedIDs: order.OrderedIDs}, modPathByID)
}

// Backward compat aliases
type PlaysetRepository = PlaysetRepo
var NewFilePlaysetRepository = NewFilePlaysetRepo
