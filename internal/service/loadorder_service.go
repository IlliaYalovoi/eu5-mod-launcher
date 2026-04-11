package service

import (
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/logging"
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
			logging.Warnf("loadorder-service: invalid mod id %q: %v", id, err)
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

	logging.Debugf("loadorder-service: validated %d ids -> %d normalized", len(ids), len(out))
	return out, nil
}

func (*LoadOrderService) Enable(current []string, modID string) ([]string, error) {
	if _, err := domain.ParseModID(modID); err != nil {
		logging.Warnf("loadorder-service: enable invalid mod id %q: %v", modID, err)
		return nil, err
	}

	next := append([]string(nil), current...)
	for _, currentID := range next {
		if currentID == modID {
			logging.Debugf("loadorder-service: mod %q already enabled", modID)
			return next, nil
		}
	}

	logging.Debugf("loadorder-service: enabled mod %q (total: %d)", modID, len(next)+1)
	return append(next, modID), nil
}

func (*LoadOrderService) Disable(current []string, modID string) ([]string, error) {
	if _, err := domain.ParseModID(modID); err != nil {
		logging.Warnf("loadorder-service: disable invalid mod id %q: %v", modID, err)
		return nil, err
	}

	next := append([]string(nil), current...)
	for i, currentID := range next {
		if currentID != modID {
			continue
		}
		logging.Debugf("loadorder-service: disabled mod %q (remaining: %d)", modID, len(next)-1)
		return append(next[:i], next[i+1:]...), nil
	}

	return next, nil
}
