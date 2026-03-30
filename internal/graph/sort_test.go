package graph_test

import (
	"errors"
	"eu5-mod-launcher/internal/graph"
	"reflect"
	"testing"
)

type sortCase struct {
	name    string
	input   []string
	edges   []graph.Constraint
	want    []string
	wantErr error
}

func TestGraphSort(t *testing.T) {
	testCases := []sortCase{
		{
			name:  "no constraints keeps input order",
			input: []string{"A", "B", "C"},
			want:  []string{"A", "B", "C"},
		},
		{
			name:  "single constraint puts C before A",
			input: []string{"A", "B", "C"},
			edges: []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "A", To: "C"}},
			want:  []string{"B", "C", "A"},
		},
		{
			name:  "multi-constraint chain",
			input: []string{"A", "B", "C", "D"},
			edges: []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "B", To: "A"}, {Type: graph.ConstraintTypeAfter, From: "C", To: "B"}, {Type: graph.ConstraintTypeAfter, From: "D", To: "C"}},
			want:  []string{"A", "B", "C", "D"},
		},
		{
			name:    "cycle returns ErrCycle",
			input:   []string{"A", "B"},
			edges:   []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "A", To: "B"}, {Type: graph.ConstraintTypeAfter, From: "B", To: "A"}},
			wantErr: graph.ErrCycle,
		},
		{
			name:  "stable order for unrelated mods",
			input: []string{"X", "A", "Y", "B"},
			edges: []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "B", To: "A"}},
			want:  []string{"X", "A", "Y", "B"},
		},
		{
			name:  "constraints outside input are ignored",
			input: []string{"A", "B"},
			edges: []graph.Constraint{{Type: graph.ConstraintTypeAfter, From: "A", To: "Z"}, {Type: graph.ConstraintTypeAfter, From: "Y", To: "B"}},
			want:  []string{"A", "B"},
		},
		{
			name:  "multiple first and last are alphabetic after topo",
			input: []string{"zeta", "alpha", "beta", "omega", "gamma"},
			edges: []graph.Constraint{
				{Type: graph.ConstraintTypeAfter, From: "gamma", To: "beta"},
				{Type: graph.ConstraintTypeFirst, ModID: "zeta"},
				{Type: graph.ConstraintTypeFirst, ModID: "alpha"},
				{Type: graph.ConstraintTypeLast, ModID: "omega"},
			},
			want: []string{"alpha", "zeta", "beta", "gamma", "omega"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runSortCase(t, tc)
		})
	}
}

func runSortCase(t *testing.T, tc sortCase) {
	t.Helper()

	constraintGraph := graph.New()
	applyEdges(constraintGraph, tc.edges)

	got, err := constraintGraph.Sort(tc.input)
	if tc.wantErr != nil {
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("Sort() error = %v, want %v", err, tc.wantErr)
		}
		return
	}

	if err != nil {
		t.Fatalf("Sort() unexpected error = %v", err)
	}
	if !reflect.DeepEqual(got, tc.want) {
		t.Fatalf("Sort() = %v, want %v", got, tc.want)
	}
}

func applyEdges(constraintGraph *graph.Graph, edges []graph.Constraint) {
	for i := range edges {
		edge := edges[i]
		switch edge.Type {
		case graph.ConstraintTypeFirst:
			constraintGraph.AddFirst(edge.ModID)
		case graph.ConstraintTypeLast:
			constraintGraph.AddLast(edge.ModID)
		default:
			constraintGraph.Add(edge.From, edge.To)
		}
	}
}
