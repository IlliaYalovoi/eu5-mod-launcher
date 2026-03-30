package graph_test

import (
	"eu5-mod-launcher/internal/graph"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveLoadConstraintsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "constraints.json")

	g := graph.New()
	g.Add("A", "C")
	g.Add("B", "A")

	if err := graph.SaveConstraints(path, g); err != nil {
		t.Fatalf("SaveConstraints() error = %v", err)
	}

	loaded, err := graph.LoadConstraints(path)
	if err != nil {
		t.Fatalf("LoadConstraints() error = %v", err)
	}

	if !reflect.DeepEqual(loaded.All(), g.All()) {
		t.Fatalf("loaded constraints = %v, want %v", loaded.All(), g.All())
	}
}

func TestLoadConstraintsMissingFileReturnsEmptyGraph(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	g, err := graph.LoadConstraints(path)
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
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	g, err := graph.LoadConstraints(path)
	if err != nil {
		t.Fatalf("LoadConstraints() error = %v", err)
	}
	want := []graph.Constraint{
		{Type: graph.ConstraintTypeAfter, From: "A", To: "B"},
		{Type: graph.ConstraintTypeAfter, From: "C", To: "A"},
	}
	if got := g.All(); !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadConstraints() = %v, want %v", got, want)
	}
}
