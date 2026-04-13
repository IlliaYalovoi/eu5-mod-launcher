package domain

import (
	"fmt"
	"strings"
)

type (
	ModID        string
	CategoryID   string
	PlaysetIndex int
)

const CategoryPrefix = "category:"

type Mod struct {
	ID            ModID
	Name          string
	Version       string
	Tags          []string
	Description   string
	ThumbnailPath string
	DirPath       string
	Enabled       bool
}

func IsCategoryID(raw string) bool {
	return len(raw) > len(CategoryPrefix) && raw[:len(CategoryPrefix)] == CategoryPrefix
}

func ParseModID(raw string) (ModID, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("mod id: %w", ErrEmptyValue)
	}
	if strings.ContainsAny(id, " \t\r\n") {
		return "", fmt.Errorf("mod id %q: %w", raw, ErrInvalidID)
	}
	if IsCategoryID(id) {
		return "", fmt.Errorf("mod id %q: %w", raw, ErrTypeMismatch)
	}
	return ModID(id), nil
}

func ParseCategoryID(raw string) (CategoryID, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("category id: %w", ErrEmptyValue)
	}
	if !IsCategoryID(id) {
		return "", fmt.Errorf("category id %q: %w", raw, ErrInvalidID)
	}
	if len(id) == len(CategoryPrefix) {
		return "", fmt.Errorf("category id %q: %w", raw, ErrInvalidID)
	}
	return CategoryID(id), nil
}

func ParsePlaysetIndex(value int) (PlaysetIndex, error) {
	if value < 0 {
		return 0, fmt.Errorf("playset index %d: %w", value, ErrOutOfRange)
	}
	return PlaysetIndex(value), nil
}
