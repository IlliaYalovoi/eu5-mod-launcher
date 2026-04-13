package graph_test

import (
	"eu5-mod-launcher/internal/graph"
	"reflect"
	"testing"
)

func TestGraphAddRemoveAndQueries(t *testing.T) {
	g := graph.New()
	g.Add("A", "B")
	g.Add("A", "B") // duplicate no-op
	g.Add("C", "A")
	g.AddFirst("F")
	g.AddLast("L")

	wantAll := []graph.Constraint{
		{Type: graph.ConstraintTypeAfter, From: "A", To: "B"},
		{Type: graph.ConstraintTypeAfter, From: "C", To: "A"},
		{Type: graph.ConstraintTypeFirst, ModID: "F"},
		{Type: graph.ConstraintTypeLast, ModID: "L"},
	}
	if got := g.All(); !reflect.DeepEqual(got, wantAll) {
		t.Fatalf("All() = %v, want %v", got, wantAll)
	}

	wantForA := []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "A", To: "B"}, {Type: graph.ConstraintTypeAfter, From: "C", To: "A"}}
	if got := g.ConstraintsFor("A"); !reflect.DeepEqual(got, wantForA) {
		t.Fatalf("ConstraintsFor(A) = %v, want %v", got, wantForA)
	}
	wantForF := []graph.Constraint{{Type: graph.ConstraintTypeFirst, ModID: "F"}}
	if got := g.ConstraintsFor("F"); !reflect.DeepEqual(got, wantForF) {
		t.Fatalf("ConstraintsFor(F) = %v, want %v", got, wantForF)
	}

	g.Remove("A", "B")
	g.RemoveFirst("F")
	wantAfterRemove := []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "C", To: "A"}, {Type: graph.ConstraintTypeLast, ModID: "L"}}
	if got := g.All(); !reflect.DeepEqual(got, wantAfterRemove) {
		t.Fatalf("All() after remove = %v, want %v", got, wantAfterRemove)
	}

	g.Remove("A", "B") // no-op
	if got := g.All(); !reflect.DeepEqual(got, wantAfterRemove) {
		t.Fatalf("All() after duplicate remove = %v, want %v", got, wantAfterRemove)
	}
}
