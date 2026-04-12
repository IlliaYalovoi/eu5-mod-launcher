package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eu5-mod-launcher/internal/domain"
)

type LauncherCategoryData struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	ModIDs []string `json:"modIds"`
}

type LauncherLayoutData struct {
	Groups      []LauncherCategoryData `json:"groups"`
	Constraints []domain.Constraint    `json:"constraints"` // Group-to-Group rules
	Ungrouped   []string               `json:"ungrouped"`
	Categories  []LauncherCategoryData `json:"categories"`
	Order       []string               `json:"order,omitempty"`
	Collapsed   map[string]bool        `json:"collapsed,omitempty"`
}

type LayoutRepository interface {
	Load(path string) (LauncherLayoutData, error)
	Save(path string, layout LauncherLayoutData) error
}

type FileLayoutRepository struct{}

func NewFileLayoutRepository() *FileLayoutRepository {
	return &FileLayoutRepository{}
}

func (*FileLayoutRepository) Load(path string) (LauncherLayoutData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return LauncherLayoutData{}, nil
		}
		return LauncherLayoutData{}, fmt.Errorf("read launcher layout %q: %w", path, err)
	}
	if strings.TrimSpace(string(content)) == "" {
		return LauncherLayoutData{}, nil
	}

	var layout LauncherLayoutData
	if err := json.Unmarshal(content, &layout); err != nil {
		return LauncherLayoutData{}, fmt.Errorf("decode launcher layout %q: %w", path, err)
	}
	if layout.Groups == nil {
		layout.Groups = []LauncherCategoryData{}
	}
	if layout.Constraints == nil {
		layout.Constraints = []domain.Constraint{}
	}
	if layout.Ungrouped == nil {
		layout.Ungrouped = []string{}
	}
	if layout.Categories == nil {
		layout.Categories = []LauncherCategoryData{}
	}
	if layout.Order == nil {
		layout.Order = []string{}
	}
	if layout.Collapsed == nil {
		layout.Collapsed = map[string]bool{}
	}
	for i := range layout.Categories {
		if layout.Categories[i].ModIDs == nil {
			layout.Categories[i].ModIDs = []string{}
		}
	}
	for i := range layout.Groups {
		if layout.Groups[i].ModIDs == nil {
			layout.Groups[i].ModIDs = []string{}
		}
	}
	return layout, nil
}

func (*FileLayoutRepository) Save(path string, layout LauncherLayoutData) error {
	if layout.Groups == nil {
		layout.Groups = []LauncherCategoryData{}
	}
	if layout.Constraints == nil {
		layout.Constraints = []domain.Constraint{}
	}
	if layout.Ungrouped == nil {
		layout.Ungrouped = []string{}
	}
	if layout.Categories == nil {
		layout.Categories = []LauncherCategoryData{}
	}
	if layout.Order == nil {
		layout.Order = []string{}
	}
	if layout.Collapsed == nil {
		layout.Collapsed = map[string]bool{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create launcher layout dir for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(layout, "", "  ")
	if err != nil {
		return fmt.Errorf("encode launcher layout %q: %w", path, err)
	}
	payload = append(payload, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o600); err != nil {
		return fmt.Errorf("write launcher layout tmp %q: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		if removeErr := os.Remove(tmp); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf(
				"replace launcher layout %q: %w; cleanup temp %q: %s",
				path,
				err,
				tmp,
				removeErr.Error(),
			)
		}
		return fmt.Errorf("replace launcher layout %q: %w", path, err)
	}
	return nil
}
