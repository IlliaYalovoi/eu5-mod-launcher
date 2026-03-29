package service

import "fmt"

type LayoutService[T any] struct {
	normalize func(T, []string) T
	save      func(T) error
}

func NewLayoutService[T any](normalize func(T, []string) T, save func(T) error) *LayoutService[T] {
	return &LayoutService[T]{normalize: normalize, save: save}
}

func (s *LayoutService[T]) Normalize(layout T, enabled []string) T {
	if s == nil || s.normalize == nil {
		return layout
	}
	return s.normalize(layout, enabled)
}

func (s *LayoutService[T]) Persist(layout T, enabled []string) (T, error) {
	normalized := s.Normalize(layout, enabled)
	if s != nil && s.save != nil {
		if err := s.save(normalized); err != nil {
			return normalized, fmt.Errorf("save layout: %w", err)
		}
	}
	return normalized, nil
}
