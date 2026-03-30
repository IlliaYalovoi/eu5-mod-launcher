package steam_test

import (
	"bytes"
	"eu5-mod-launcher/internal/steam"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageCacheDownloadAndReuse(t *testing.T) {
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

	cache, err := steam.NewImageCache(filepath.Join(t.TempDir(), "cache"), 20, &http.Client{Timeout: 5 * time.Second})
	require.NoError(t, err)

	item := steam.WorkshopItem{ItemID: "111", PreviewURL: server.URL + "/111.png"}
	firstPath, err := cache.EnsureStored(item)
	require.NoError(t, err)
	require.NotEmpty(t, firstPath)

	cachedPath := cache.CachedPath("111")
	require.Equal(t, firstPath, cachedPath)

	secondPath, err := cache.EnsureStored(item)
	require.NoError(t, err)
	assert.Equal(t, firstPath, secondPath)
	assert.Equal(t, 1, hitCount)
}

func TestImageCacheDecodeGuardAndInvalidURL(t *testing.T) {
	t.Parallel()

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, writeErr := w.Write([]byte("not-an-image"))
		require.NoError(t, writeErr)
	}))
	t.Cleanup(badServer.Close)

	cache, err := steam.NewImageCache(filepath.Join(t.TempDir(), "cache"), 20, &http.Client{Timeout: 5 * time.Second})
	require.NoError(t, err)

	_, err = cache.RefreshStored(steam.WorkshopItem{ItemID: "222", PreviewURL: badServer.URL + "/bad.png"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image decode guard failed")

	_, err = cache.RefreshStored(steam.WorkshopItem{ItemID: "333", PreviewURL: "file:///tmp/xx.png"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "preview url scheme is unsupported")
}

func buildTinyPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	var out bytes.Buffer
	err := png.Encode(&out, img)
	require.NoError(t, err)
	return out.Bytes()
}
