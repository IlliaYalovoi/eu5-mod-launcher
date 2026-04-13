package steam_test

import (
	"eu5-mod-launcher/internal/steam"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataCacheHitMissAndExpiry(t *testing.T) {
	t.Parallel()

	cacheRoot := filepath.Join(t.TempDir(), "cache")
	cache, err := steam.NewMetadataCache(cacheRoot, 20*time.Millisecond, 100)
	require.NoError(t, err)

	lookup, err := cache.Get("123")
	require.NoError(t, err)
	assert.False(t, lookup.Hit)
	assert.False(t, lookup.Stale)

	err = cache.SetMany(map[string]steam.WorkshopItem{
		"123": {ItemID: "123", Title: "First"},
	})
	require.NoError(t, err)

	lookup, err = cache.Get("123")
	require.NoError(t, err)
	require.True(t, lookup.Hit)
	assert.False(t, lookup.Stale)
	assert.Equal(t, "First", lookup.Item.Title)

	time.Sleep(30 * time.Millisecond)
	lookup, err = cache.Get("123")
	require.NoError(t, err)
	require.True(t, lookup.Hit)
	assert.True(t, lookup.Stale)
	assert.Equal(t, "123", lookup.Item.ItemID)
}

func TestMetadataCacheResolveManyAndPersistence(t *testing.T) {
	t.Parallel()

	cacheRoot := filepath.Join(t.TempDir(), "cache")
	cache, err := steam.NewMetadataCache(cacheRoot, 1*time.Hour, 100)
	require.NoError(t, err)

	err = cache.SetMany(map[string]steam.WorkshopItem{
		"111": {ItemID: "111", Title: "One"},
		"222": {ItemID: "222", Title: "Two"},
	})
	require.NoError(t, err)

	resolved, err := cache.ResolveMany([]string{"111", "333", "222"})
	require.NoError(t, err)
	require.Len(t, resolved.Stale, 0)
	require.Equal(t, []string{"333"}, resolved.Missing)
	require.Len(t, resolved.Fresh, 2)
	assert.Equal(t, "One", resolved.Fresh["111"].Title)

	second, err := steam.NewMetadataCache(cacheRoot, 1*time.Hour, 100)
	require.NoError(t, err)
	lookup, err := second.Get("222")
	require.NoError(t, err)
	require.True(t, lookup.Hit)
	assert.False(t, lookup.Stale)
	assert.Equal(t, "Two", lookup.Item.Title)
}
