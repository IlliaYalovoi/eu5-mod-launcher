package repo

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/loadorder"
)

type FileLoadOrderRepo struct {
	store *loadorder.Store
}

func NewFileLoadOrderRepo(store *loadorder.Store) *FileLoadOrderRepo {
	return &FileLoadOrderRepo{store: store}
}

func (r *FileLoadOrderRepo) Path() string { return r.store.ConfigPath() }

func (r *FileLoadOrderRepo) Load() (domain.LoadOrder, error) {
	state, err := r.store.Load()
	if err != nil {
		return domain.LoadOrder{}, err
	}
	return domain.LoadOrder{
		GameID:     domain.GameIDEU5,
		PlaysetIdx: 0,
		OrderedIDs: state.OrderedIDs,
	}, nil
}

func (r *FileLoadOrderRepo) Save(order domain.LoadOrder) error {
	return r.store.Save(loadorder.State{OrderedIDs: order.OrderedIDs})
}

