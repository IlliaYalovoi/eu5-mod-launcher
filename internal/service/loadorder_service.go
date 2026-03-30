package service

import (
	"eu5-mod-launcher/internal/domain"
	"fmt"
	"strings"
)

type LoadOrderService struct{}

func NewLoadOrderService() *LoadOrderService {
	return &LoadOrderService{}
}

func (*LoadOrderService) ValidateAndNormalize(ids []string) ([]string, error) {
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

func (*LoadOrderService) Enable(current []string, modID string) ([]string, error) {
	if _, err := domain.ParseModID(modID); err != nil {
		return nil, err
	}

	next := append([]string(nil), current...)
	for _, currentID := range next {
		if currentID == modID {
			return next, nil
		}
	}

	return append(next, modID), nil
}

func (*LoadOrderService) Disable(current []string, modID string) ([]string, error) {
	if _, err := domain.ParseModID(modID); err != nil {
		return nil, err
	}

	next := append([]string(nil), current...)
	for i, currentID := range next {
		if currentID != modID {
			continue
		}
		return append(next[:i], next[i+1:]...), nil
	}

	return next, nil
}
