package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLayout struct {
	Items []string
}

var errBoom = errors.New("boom")

func TestLayoutServicePersistNormalizesAndSaves(t *testing.T) {
	saved := fakeLayout{}
	svc := NewLayoutService(func(layout fakeLayout, _ []string) fakeLayout {
		return fakeLayout{Items: append(layout.Items, "normalized")}
	}, func(layout fakeLayout) error {
		saved = layout
		return nil
	})

	next := fakeLayout{Items: []string{"a"}}
	err := svc.Persist(&next, []string{"mod1"})
	require.NoError(t, err)
	require.Len(t, next.Items, 2)
	require.Len(t, saved.Items, 2)
	assert.Equal(t, "normalized", next.Items[1])
	assert.Equal(t, "normalized", saved.Items[1])
}

func TestLayoutServicePersistSaveError(t *testing.T) {
	svc := NewLayoutService(func(layout fakeLayout, _ []string) fakeLayout {
		return layout
	}, func(_ fakeLayout) error {
		return errBoom
	})

	next := fakeLayout{}
	err := svc.Persist(&next, nil)
	assert.ErrorIs(t, err, errBoom)
}
