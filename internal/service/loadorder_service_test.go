package service

import (
	"eu5-mod-launcher/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrderServiceValidateAndNormalize(t *testing.T) {
	svc := NewLoadOrderService()
	got, err := svc.ValidateAndNormalize([]string{"mod1", "", "mod1", "mod2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"mod1", "mod2"}, got)
}

func TestLoadOrderServiceValidateAndNormalizeRejectsCategoryID(t *testing.T) {
	svc := NewLoadOrderService()
	_, err := svc.ValidateAndNormalize([]string{"category:graphics"})
	assert.ErrorIs(t, err, domain.ErrTypeMismatch)
}

func TestLoadOrderServiceEnableDisable(t *testing.T) {
	svc := NewLoadOrderService()
	state, err := svc.Enable([]string{"mod1"}, "mod2")
	require.NoError(t, err)
	assert.Equal(t, []string{"mod1", "mod2"}, state)

	state, err = svc.Disable(state, "mod1")
	require.NoError(t, err)
	assert.Equal(t, []string{"mod2"}, state)
}
