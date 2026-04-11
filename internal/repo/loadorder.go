package repo

import "eu5-mod-launcher/internal/domain"

type LoadOrderRepo interface {
	Path() string
	Load() (domain.LoadOrder, error)
	Save(order domain.LoadOrder) error
}
