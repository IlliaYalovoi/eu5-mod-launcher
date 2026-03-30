package service

import (
	"fmt"
)

type LayoutService[T any] struct {
	normalize func(T, []string) T
	save      func(T) error
}

func NewLayoutService[T any](normalize func(T, []string) T, save func(T) error) *LayoutService[T] {
	return &LayoutService[T]{normalize: normalize, save: save}
}

func (s *LayoutService[T]) Normalize(layout *T, enabled []string) {
	if s == nil || s.normalize == nil || layout == nil {
		return
	}
	*layout = s.normalize(*layout, enabled)
}

func (s *LayoutService[T]) Persist(layout *T, enabled []string) error {
	s.Normalize(layout, enabled)
	if s == nil || s.save == nil || layout == nil {
		return nil
	}
	if err := s.save(*layout); err != nil {
		return fmt.Errorf("save layout: %w", err)
	}
	return nil
}
