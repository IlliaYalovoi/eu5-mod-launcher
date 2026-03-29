package service

import (
	"errors"
	"testing"
)

type fakeLayout struct {
	Items []string
}

func TestLayoutServicePersistNormalizesAndSaves(t *testing.T) {
	saved := fakeLayout{}
	svc := NewLayoutService(func(layout fakeLayout, _ []string) fakeLayout {
		return fakeLayout{Items: append(layout.Items, "normalized")}
	}, func(layout fakeLayout) error {
		saved = layout
		return nil
	})

	next, err := svc.Persist(fakeLayout{Items: []string{"a"}}, []string{"mod1"})
	if err != nil {
		t.Fatalf("Persist() error = %v", err)
	}
	if len(next.Items) != 2 || next.Items[1] != "normalized" {
		t.Fatalf("Persist() = %#v", next)
	}
	if len(saved.Items) != 2 || saved.Items[1] != "normalized" {
		t.Fatalf("save() received %#v", saved)
	}
}

func TestLayoutServicePersistSaveError(t *testing.T) {
	svc := NewLayoutService(func(layout fakeLayout, _ []string) fakeLayout {
		return layout
	}, func(layout fakeLayout) error {
		return errors.New("boom")
	})

	_, err := svc.Persist(fakeLayout{}, nil)
	if err == nil {
		t.Fatalf("Persist() error = nil, want error")
	}
}
