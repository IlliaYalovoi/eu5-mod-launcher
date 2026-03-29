package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveLoadConstraintsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "constraints.json")

	g := New()
	g.Add("A", "C")
	g.Add("B", "A")

	if err := SaveConstraints(path, g); err != nil {
		t.Fatalf("SaveConstraints() error = %v", err)
	}

	loaded, err := LoadConstraints(path)
	if err != nil {
		t.Fatalf("LoadConstraints() error = %v", err)
	}

	if !reflect.DeepEqual(loaded.All(), g.All()) {
		t.Fatalf("loaded constraints = %v, want %v", loaded.All(), g.All())
	}
}

func TestLoadConstraintsMissingFileReturnsEmptyGraph(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	g, err := LoadConstraints(path)
	if err != nil {
		t.Fatalf("LoadConstraints() error = %v", err)
	}
	if got := g.All(); len(got) != 0 {
		t.Fatalf("LoadConstraints() returned constraints %v, want empty", got)
	}
}

func TestLoadConstraintsLegacyFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "constraints.json")
	legacy := `[
  {"from":"A","to":"B"},
  {"from":"C","to":"A"}
]`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	g, err := LoadConstraints(path)
	if err != nil {
		t.Fatalf("LoadConstraints() error = %v", err)
	}
	want := []Constraint{
		{Type: ConstraintTypeAfter, From: "A", To: "B"},
		{Type: ConstraintTypeAfter, From: "C", To: "A"},
	}
	if got := g.All(); !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadConstraints() = %v, want %v", got, want)
	}
}
