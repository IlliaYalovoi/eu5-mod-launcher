package service

import (
	"fmt"
	"strings"

	"eu5-mod-launcher/internal/domain"
)

type LoadOrderService struct{}

func NewLoadOrderService() *LoadOrderService {
	return &LoadOrderService{}
}

func (s *LoadOrderService) ValidateAndNormalize(ids []string) ([]string, error) {
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			continue
		}
		if _, err := domain.ParseModID(id); err != nil {
			return nil, fmt.Errorf("invalid mod id %q: %w", id, err)
		}
	}

	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func (s *LoadOrderService) ToggleEnabled(current []string, id string, enabled bool) ([]string, error) {
	if _, err := domain.ParseModID(id); err != nil {
		return nil, err
	}

	next := append([]string(nil), current...)
	index := -1
	for i, currentID := range next {
		if currentID == id {
			index = i
			break
		}
	}

	if enabled {
		if index < 0 {
			next = append(next, id)
		}
	} else if index >= 0 {
		next = append(next[:index], next[index+1:]...)
	}

	return next, nil
}
