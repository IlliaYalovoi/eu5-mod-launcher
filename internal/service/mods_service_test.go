package service

import (
	"testing"
)

func TestModsServiceDiscoverMissingRootsReturnsEmpty(t *testing.T) {
	svc := NewModsService()
	mods, paths, err := svc.Discover([]string{"C:/__definitely_missing_root__"}, []string{"mod1"}, map[string]string{})
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(mods) != 0 {
		t.Fatalf("Discover() mods len = %d, want 0", len(mods))
	}
	if len(paths) != 0 {
		t.Fatalf("Discover() paths len = %d, want 0", len(paths))
	}
}
