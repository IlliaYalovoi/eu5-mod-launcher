package graph

import (
	"errors"
	"reflect"
	"testing"
)

func TestGraphSort(t *testing.T) {
	testCases := []struct {
		name    string
		input   []string
		edges   []Constraint
		want    []string
		wantErr error
	}{
		{
			name:  "no constraints keeps input order",
			input: []string{"A", "B", "C"},
			want:  []string{"A", "B", "C"},
		},
		{
			name:  "single constraint puts C before A",
			input: []string{"A", "B", "C"},
			edges: []Constraint{{Type: ConstraintTypeAfter, From: "A", To: "C"}},
			want:  []string{"B", "C", "A"},
		},
		{
			name:  "multi-constraint chain",
			input: []string{"A", "B", "C", "D"},
			edges: []Constraint{{Type: ConstraintTypeAfter, From: "B", To: "A"}, {Type: ConstraintTypeAfter, From: "C", To: "B"}, {Type: ConstraintTypeAfter, From: "D", To: "C"}},
			want:  []string{"A", "B", "C", "D"},
		},
		{
			name:    "cycle returns ErrCycle",
			input:   []string{"A", "B"},
			edges:   []Constraint{{Type: ConstraintTypeAfter, From: "A", To: "B"}, {Type: ConstraintTypeAfter, From: "B", To: "A"}},
			wantErr: ErrCycle,
		},
		{
			name:  "stable order for unrelated mods",
			input: []string{"X", "A", "Y", "B"},
			edges: []Constraint{{Type: ConstraintTypeAfter, From: "B", To: "A"}},
			want:  []string{"X", "A", "Y", "B"},
		},
		{
			name:  "constraints outside input are ignored",
			input: []string{"A", "B"},
			edges: []Constraint{{Type: ConstraintTypeAfter, From: "A", To: "Z"}, {Type: ConstraintTypeAfter, From: "Y", To: "B"}},
			want:  []string{"A", "B"},
		},
		{
			name:  "multiple first and last are alphabetic after topo",
			input: []string{"zeta", "alpha", "beta", "omega", "gamma"},
			edges: []Constraint{
				{Type: ConstraintTypeAfter, From: "gamma", To: "beta"},
				{Type: ConstraintTypeFirst, ModID: "zeta"},
				{Type: ConstraintTypeFirst, ModID: "alpha"},
				{Type: ConstraintTypeLast, ModID: "omega"},
			},
			want: []string{"alpha", "zeta", "beta", "gamma", "omega"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := New()
			for _, edge := range tc.edges {
				switch edge.Type {
				case ConstraintTypeFirst:
					g.AddFirst(edge.ModID)
				case ConstraintTypeLast:
					g.AddLast(edge.ModID)
				default:
					g.Add(edge.From, edge.To)
				}
			}

			got, err := g.Sort(tc.input)
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
		})
	}
}
