package steam_test

import (
	"eu5-mod-launcher/internal/steam"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescriptionImageCacheDownloadAndReuse(t *testing.T) {
	t.Parallel()

	pngPayload := buildTinyPNG(t)
	var hitCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hitCount++
		w.Header().Set("Content-Type", "image/png")
		_, writeErr := w.Write(pngPayload)
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	cache, err := steam.NewDescriptionImageCache(
		filepath.Join(t.TempDir(), "cache"),
		50,
		&http.Client{Timeout: 5 * time.Second},
	)
	require.NoError(t, err)

	path1, err := cache.EnsureStored("123", server.URL+"/a.png")
	require.NoError(t, err)
	require.NotEmpty(t, path1)

	path2, err := cache.EnsureStored("123", server.URL+"/a.png")
	require.NoError(t, err)
	assert.Equal(t, path1, path2)
	assert.Equal(t, 1, hitCount)
}

func TestDescriptionImageCacheRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	cache, err := steam.NewDescriptionImageCache(
		filepath.Join(t.TempDir(), "cache"),
		50,
		&http.Client{Timeout: 5 * time.Second},
	)
	require.NoError(t, err)

	_, err = cache.EnsureStored("not-id", "https://example.com/a.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workshop item id")

	_, err = cache.EnsureStored("123", "file:///tmp/a.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description image url scheme is unsupported")
}

func TestDescriptionImageCacheInvalidExistingFileIsRedownloaded(t *testing.T) {
	t.Parallel()

	pngPayload := buildTinyPNG(t)
	var hitCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hitCount++
		w.Header().Set("Content-Type", "image/png")
		_, writeErr := w.Write(pngPayload)
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	cache, err := steam.NewDescriptionImageCache(
		filepath.Join(t.TempDir(), "cache"),
		50,
		&http.Client{Timeout: 5 * time.Second},
	)
	require.NoError(t, err)

	path1, err := cache.EnsureStored("123", server.URL+"/broken.png")
	require.NoError(t, err)
	require.NotEmpty(t, path1)
	require.Equal(t, 1, hitCount)

	err = os.WriteFile(path1, []byte("not-an-image"), 0o600)
	require.NoError(t, err)

	path2, err := cache.EnsureStored("123", server.URL+"/broken.png")
	require.NoError(t, err)
	assert.Equal(t, path1, path2)
	assert.Equal(t, 2, hitCount)
}

func TestDescriptionImageCacheRepairsLegacyIMGExtension(t *testing.T) {
	t.Parallel()

	pngPayload := buildTinyPNG(t)
	var hitCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hitCount++
		w.Header().Set("Content-Type", "image/png")
		_, writeErr := w.Write(pngPayload)
		require.NoError(t, writeErr)
	}))
	t.Cleanup(server.Close)

	cacheRoot := filepath.Join(t.TempDir(), "cache")
	cache, err := steam.NewDescriptionImageCache(
		cacheRoot,
		50,
		&http.Client{Timeout: 5 * time.Second},
	)
	require.NoError(t, err)

	itemID := "123"
	imageURL := server.URL + "/noext"

	firstPath, err := cache.EnsureStored(itemID, imageURL)
	require.NoError(t, err)
	require.Equal(t, 1, hitCount)

	legacyName := strings.TrimSuffix(firstPath, filepath.Ext(firstPath)) + ".img"
	err = os.Rename(firstPath, legacyName)
	require.NoError(t, err)

	resolved, err := cache.EnsureStored(itemID, imageURL)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSuffix(legacyName, ".img")+".png", resolved)
	assert.Equal(t, 1, hitCount)
}
