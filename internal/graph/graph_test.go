package graph

import (
	"reflect"
	"testing"
)

func TestGraphAddRemoveAndQueries(t *testing.T) {
	g := New()
	g.Add("A", "B")
	g.Add("A", "B") // duplicate no-op
	g.Add("C", "A")
	g.AddFirst("F")
	g.AddLast("L")

	wantAll := []Constraint{
		{Type: ConstraintTypeAfter, From: "A", To: "B"},
		{Type: ConstraintTypeAfter, From: "C", To: "A"},
		{Type: ConstraintTypeFirst, ModID: "F"},
		{Type: ConstraintTypeLast, ModID: "L"},
	}
	if got := g.All(); !reflect.DeepEqual(got, wantAll) {
		t.Fatalf("All() = %v, want %v", got, wantAll)
	}

	wantForA := []Constraint{{Type: ConstraintTypeAfter, From: "A", To: "B"}, {Type: ConstraintTypeAfter, From: "C", To: "A"}}
	if got := g.ConstraintsFor("A"); !reflect.DeepEqual(got, wantForA) {
		t.Fatalf("ConstraintsFor(A) = %v, want %v", got, wantForA)
	}
	wantForF := []Constraint{{Type: ConstraintTypeFirst, ModID: "F"}}
	if got := g.ConstraintsFor("F"); !reflect.DeepEqual(got, wantForF) {
		t.Fatalf("ConstraintsFor(F) = %v, want %v", got, wantForF)
	}

	g.Remove("A", "B")
	g.RemoveFirst("F")
	wantAfterRemove := []Constraint{{Type: ConstraintTypeAfter, From: "C", To: "A"}, {Type: ConstraintTypeLast, ModID: "L"}}
	if got := g.All(); !reflect.DeepEqual(got, wantAfterRemove) {
		t.Fatalf("All() after remove = %v, want %v", got, wantAfterRemove)
	}

	g.Remove("A", "B") // no-op
	if got := g.All(); !reflect.DeepEqual(got, wantAfterRemove) {
		t.Fatalf("All() after duplicate remove = %v, want %v", got, wantAfterRemove)
	}
}
