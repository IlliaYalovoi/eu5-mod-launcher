package service

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOpenURLCommand(t *testing.T) {
	t.Parallel()

	svc := NewLaunchService()
	cmd, err := svc.BuildOpenURLCommand("windows", "https://steamcommunity.com/sharedfiles/filedetails/?id=123")
	require.NoError(t, err)
	require.NotNil(t, cmd)
	assert.Equal(t, "rundll32.exe", filepath.Base(cmd.Path))
}

func TestBuildOpenURLCommandRejectsUnsupportedScheme(t *testing.T) {
	t.Parallel()

	svc := NewLaunchService()
	_, err := svc.BuildOpenURLCommand("windows", "file:///tmp/test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported url scheme")
}
